package execution

import (
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/task"
)

// CleanupBound is the registered time within which cleanup after
// cancellation must stop, measured from the attempt deadline or the
// cancellation (REQ-LIMIT-0008). Waiters use it to report cleanup that has not
// finished; the engine never extends network activity to wait for it.
const CleanupBound = 2 * time.Second

// MaxAttemptDuration is the registered maximum of max_duration_ms
// (REQ-LIMIT-0006). A plan with a longer attempt deadline is invalid.
const MaxAttemptDuration = 60 * time.Second

// Registered maxima of max_bytes_sent, max_bytes_received and max_connections
// (REQ-LIMIT-0006). A plan with a larger effective limit is invalid.
const (
	MaxBytesSent     = 65536
	MaxBytesReceived = 262144
	MaxConnections   = 12
)

var (
	errScopeClosed = errors.New("execution: attempt scope closed")
	// errBudget is the error a refused metered operation returns and the
	// cancellation cause of an attempt whose budget it exhausted.
	errBudget = errors.New("execution: attempt budget exhausted")
)

// scope is the attempt-scoped dialer the engine hands to a runner in place of
// the path's dialer. It records the test start (REQ-RESULT-0051) and owns every
// connection opened through it: when the attempt context ends it closes them,
// which interrupts blocked reads and writes, and refuses later dials. close
// releases whatever the runner left open when it returns.
//
// The scope is also the attempt's meter. It counts where the attempt exchanges
// bytes with the path's stream or, after Session, with a session's plaintext
// side (REQ-LIMIT-0034), so a runner's own read caps and usage reports never
// decide what is counted. Every dial counts toward
// max_connections before it is issued, whether or not it connects. A write
// larger than the remaining sent budget is not issued, and a read requests no
// more than the remaining received budget. An operation the remaining budget
// cannot cover ends the attempt with errBudget as its cause and so releases
// every connection.
//
// A dial that the path fails with a path failure (proxy_unavailable,
// proxy_connect_failed) likewise ends the attempt with that failure as its
// cause. When the attempt had already ended, the failure only records that the
// path's dial was interrupted, which the engine reads with interrupted.
type scope struct {
	ctx      context.Context
	inner    measurement.Dialer
	targets  TargetCheck
	limits   task.AttemptLimits
	end      func(cause error)
	started  atomic.Bool
	terminal atomic.Bool
	stop     func() bool

	mu     sync.Mutex
	closed bool
	conns  map[*scopedConn]struct{}
	usage  measurement.Usage // includes bytes reserved by operations in flight
	late   task.ExecutionReason
}

func newScope(ctx context.Context, d measurement.Dialer, limits task.AttemptLimits, end func(cause error)) *scope {
	s := &scope{ctx: ctx, inner: d, limits: limits, end: end, conns: make(map[*scopedConn]struct{})}
	s.stop = context.AfterFunc(ctx, s.release)
	return s
}

// Start marks the test as started. It is idempotent.
func (s *scope) Start() { s.started.Store(true) }

func (s *scope) Started() bool { return s.started.Load() }

// RecordOutcome distinguishes an observed protocol result from interruption
// during immediate close. A condition already observed cannot be overwritten.
func (s *scope) RecordOutcome() bool {
	if s.ctx.Err() != nil {
		return false
	}
	s.terminal.Store(true)
	return true
}

// Usage reports the metered bytes and connections; after close it is final.
func (s *scope) Usage() measurement.Usage {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.usage
}

// DialContext starts the test, counts the connection and dials through the
// path's dialer. The dial ends with the attempt even when the runner passes
// another context.
// requirement-ref: REQ-LIMIT-0006
func (s *scope) DialContext(ctx context.Context, network, address string) (measurement.Conn, error) {
	if s.isClosed() || s.ctx.Err() != nil {
		return nil, errScopeClosed
	}
	return s.dial(ctx, network, address, s.targets)
}

func (s *scope) dial(ctx context.Context, network, address string, check TargetCheck) (measurement.Conn, error) {
	return s.dialVia(ctx, network, address, check, func(c context.Context, a string) (measurement.Conn, error) { return s.inner.DialContext(c, network, a) })
}

func (s *scope) dialVia(ctx context.Context, network, address string, check TargetCheck, connect func(context.Context, string) (measurement.Conn, error)) (measurement.Conn, error) {
	if s.isClosed() || s.ctx.Err() != nil {
		return nil, errScopeClosed
	}
	dctx, cancel := context.WithCancel(ctx)
	defer cancel()
	unlink := context.AfterFunc(s.ctx, cancel)
	defer unlink()
	if check == nil {
		s.end(pathError{task.LocalPolicyRejected})
		return nil, ErrTargetRejected
	}
	// Reserve a connection slot before resolution can send any traffic. A local
	// refusal releases it; concurrent dials cannot all resolve beyond the bound.
	s.mu.Lock()
	if s.usage.Connections >= s.limits.Connections {
		s.mu.Unlock()
		s.end(errBudget)
		return nil, errBudget
	}
	s.usage.Connections++
	s.mu.Unlock()
	issued := false
	defer func() {
		if !issued {
			s.mu.Lock()
			s.usage.Connections--
			s.mu.Unlock()
		}
	}()
	checked, err := check.Prepare(dctx, network, address, s.Start)
	if err != nil {
		if reason, ok := pathFailure(err); ok {
			s.pathFailed(reason, err)
		}
		return nil, err
	}
	if dctx.Err() != nil {
		return nil, errScopeClosed
	}
	s.Start()
	issued = true
	conn, err := connect(dctx, checked)
	if err != nil {
		if reason, ok := pathFailure(err); ok {
			s.pathFailed(reason, err)
		}
		return nil, err
	}
	c := &scopedConn{Conn: conn, scope: s}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		_ = conn.Close()
		return nil, errScopeClosed
	}
	s.conns[c] = struct{}{}
	s.mu.Unlock()
	return c, nil
}

// reserve takes budget for an operation of n bytes before it is issued: all n
// bytes for a write, up to n for a read. It returns the reserved count, or
// false after ending the attempt when the budget cannot cover a non-empty
// operation, or when the attempt has already ended.
// requirement-ref: REQ-LIMIT-0034
// requirement-ref: REQ-LIMIT-0027
func (s *scope) reserve(n int, sent bool) (int, bool) {
	s.mu.Lock()
	if s.closed || s.ctx.Err() != nil {
		s.mu.Unlock()
		return 0, false
	}
	used, limit := &s.usage.BytesReceived, s.limits.BytesReceived
	if sent {
		used, limit = &s.usage.BytesSent, s.limits.BytesSent
	}
	k := min(n, limit-*used)
	if n > 0 && (k <= 0 || sent && k < n) {
		s.mu.Unlock()
		s.end(errBudget)
		return 0, false
	}
	*used += k
	s.mu.Unlock()
	return k, true
}

// pathFailed ends a running attempt with the path failure as its cause, or
// records the failure of a dial that the attempt's end interrupted.
func (s *scope) pathFailed(reason task.ExecutionReason, err error) {
	if s.ctx.Err() == nil {
		s.end(err)
		return
	}
	s.mu.Lock()
	if s.late == "" {
		s.late = reason
	}
	s.mu.Unlock()
}

// interrupted reports the path failure of a dial that was in progress when the
// attempt ended, such as a CONNECT exchange the attempt deadline cut short.
func (s *scope) interrupted() task.ExecutionReason {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.late
}

// settle returns the unused part of a reservation once the operation reports
// the bytes it transferred.
func (s *scope) settle(reserved, n int, sent bool) {
	n = max(n, 0)
	if n >= reserved {
		return
	}
	s.mu.Lock()
	if sent {
		s.usage.BytesSent -= reserved - n
	} else {
		s.usage.BytesReceived -= reserved - n
	}
	s.mu.Unlock()
}

// err is the error a refused operation reports: errBudget once the budget
// ended the attempt, otherwise errScopeClosed.
func (s *scope) err() error {
	if errors.Is(context.Cause(s.ctx), errBudget) {
		return errBudget
	}
	return errScopeClosed
}

func (s *scope) isClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}

// release closes every open connection and refuses later dials.
func (s *scope) release() {
	s.mu.Lock()
	s.closed = true
	open := make([]*scopedConn, 0, len(s.conns))
	for c := range s.conns {
		open = append(open, c)
	}
	clear(s.conns)
	s.mu.Unlock()
	for _, c := range open {
		_ = c.close()
	}
}

// close ends the scope after the runner has returned.
func (s *scope) close() {
	s.stop()
	s.release()
}

// scopedConn meters its reads and writes through its scope and removes itself
// from the scope when closed. Close is idempotent.
type scopedConn struct {
	measurement.Conn
	scope   *scope
	session atomic.Bool
	once    sync.Once
	err     error
}

var errSession = errors.New("execution: connection carries a session")

// Write issues p only when the remaining sent budget covers all of it. A
// partial write counts only the bytes the stream accepted.
func (c *scopedConn) Write(p []byte) (int, error) {
	if c.session.Load() {
		return 0, errSession
	}
	return c.scope.write(c.Conn, p)
}

// Read requests no more than the remaining received budget, so bytes a caller
// reads ahead into a buffer count when they are read.
func (c *scopedConn) Read(p []byte) (int, error) {
	if c.session.Load() {
		return 0, errSession
	}
	return c.scope.read(c.Conn, p)
}

// Session moves metering from the stream to a session's plaintext side, so
// TLS handshake and record bytes are not counted (REQ-LIMIT-0034). The raw
// view still closes through this connection, and release still closes it.
// requirement-ref: REQ-LIMIT-0034
func (c *scopedConn) Session() (measurement.Conn, func(io.ReadWriter) io.ReadWriter, error) {
	if !c.session.CompareAndSwap(false, true) {
		return nil, nil, errSession
	}
	return rawConn{c}, func(rw io.ReadWriter) io.ReadWriter { return plaintext{rw, c.scope} }, nil
}

func (s *scope) write(w io.Writer, p []byte) (int, error) {
	k, ok := s.reserve(len(p), true)
	if !ok {
		return 0, s.err()
	}
	n, err := w.Write(p)
	s.settle(k, n, true)
	return n, err
}

func (s *scope) read(r io.Reader, p []byte) (int, error) {
	k, ok := s.reserve(len(p), false)
	if !ok {
		return 0, s.err()
	}
	n, err := r.Read(p[:k])
	s.settle(k, n, false)
	return n, err
}

// rawConn is the unmetered stream under a session. It refuses I/O once the
// attempt has ended, as the metered connection does.
type rawConn struct{ c *scopedConn }

func (r rawConn) Read(p []byte) (int, error) {
	if r.c.scope.isClosed() || r.c.scope.ctx.Err() != nil {
		return 0, r.c.scope.err()
	}
	return r.c.Conn.Read(p)
}
func (r rawConn) Write(p []byte) (int, error) {
	if r.c.scope.isClosed() || r.c.scope.ctx.Err() != nil {
		return 0, r.c.scope.err()
	}
	return r.c.Conn.Write(p)
}
func (r rawConn) SetDeadline(t time.Time) error { return r.c.Conn.SetDeadline(t) }
func (r rawConn) Close() error                  { return r.c.Close() }

// plaintext meters a session's application bytes through the attempt scope.
type plaintext struct {
	rw    io.ReadWriter
	scope *scope
}

func (p plaintext) Write(b []byte) (int, error) { return p.scope.write(p.rw, b) }
func (p plaintext) Read(b []byte) (int, error)  { return p.scope.read(p.rw, b) }

func (c *scopedConn) Close() error {
	c.scope.mu.Lock()
	delete(c.scope.conns, c)
	c.scope.mu.Unlock()
	return c.close()
}

func (c *scopedConn) close() error {
	c.once.Do(func() { c.err = c.Conn.Close() })
	return c.err
}

// Redirect returns no raw path dialer: all hops retain shared metering,
// cancellation and per-contact authorization.
// requirement-ref: REQ-HTTP-01-0051
func (s *scope) Redirect(tuple task.TargetTuple) (measurement.TargetDialer, error) {
	if s.targets == nil || s.isClosed() || s.ctx.Err() != nil {
		return nil, ErrTargetRejected
	}
	check, err := s.targets.Redirect(tuple)
	if err != nil {
		return nil, ErrTargetRejected
	}
	return targetDialer{scope: s, check: check}, nil
}

type targetDialer struct {
	scope *scope
	check TargetCheck
}

func (d targetDialer) DialContext(ctx context.Context, network, address string) (measurement.Conn, error) {
	return d.scope.dial(ctx, network, address, d.check)
}

func (d targetDialer) Redirect(tuple task.TargetTuple) (measurement.TargetDialer, error) {
	if d.scope.isClosed() || d.scope.ctx.Err() != nil {
		return nil, ErrTargetRejected
	}
	check, err := d.check.Redirect(tuple)
	if err != nil {
		return nil, ErrTargetRejected
	}
	return targetDialer{scope: d.scope, check: check}, nil
}

// ReadDatagram meters all returned DNS bytes, preserving truncation detection.
func (c *scopedConn) ReadDatagram(p []byte) (int, bool, bool, error) {
	k, ok := c.scope.reserve(len(p), false)
	if !ok {
		return 0, false, false, c.scope.err()
	}
	var n int
	var truncated bool
	matchingSource := true
	var err error
	if d, ok := c.Conn.(measurement.DatagramConn); ok {
		n, truncated, matchingSource, err = d.ReadDatagram(p[:k])
	} else {
		n, err = c.Conn.Read(p[:k])
		truncated = n == k && k < len(p)
	}
	c.scope.settle(k, n, false)
	if truncated && k < len(p) {
		c.scope.end(errBudget)
		return n, true, matchingSource, c.scope.err()
	}
	return n, truncated, matchingSource, err
}

// resolve retains execution ownership of the target gate and hidden-wire API.
func (s *scope) resolve(ctx context.Context, name, query string) (bool, string, error) {
	t, ok := s.targets.(interface {
		Lookup(context.Context, string, string, func()) (bool, string, error)
	})
	if !ok || s.isClosed() || s.ctx.Err() != nil {
		s.end(pathError{task.LocalPolicyRejected})
		return false, "", measurement.ErrTargetRejected
	}
	dctx, cancel := context.WithCancel(ctx)
	defer cancel()
	unlink := context.AfterFunc(s.ctx, cancel)
	defer unlink()
	found, failure, err := t.Lookup(dctx, name, query, s.Start)
	if reason, ok := pathFailure(err); ok {
		s.pathFailed(reason, err)
	}
	return found, failure, err
}
