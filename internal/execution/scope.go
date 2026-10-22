package execution

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tv-rain/barometer-go/internal/measurement"
)

// CleanupBound is the registered time within which cleanup after
// cancellation must stop, measured from the attempt deadline or the
// cancellation (REQ-LIMIT-0008). Waiters use it to report cleanup that has not
// finished; the engine never extends network activity to wait for it.
const CleanupBound = 2 * time.Second

// MaxAttemptDuration is the registered maximum of max_duration_ms
// (REQ-LIMIT-0006). A plan with a longer attempt deadline is invalid.
const MaxAttemptDuration = 60 * time.Second

var errScopeClosed = errors.New("execution: attempt scope closed")

// scope is the attempt-scoped dialer the engine hands to a runner in place of
// the path's dialer. It records the test start (REQ-RESULT-0051) and owns every
// connection opened through it: when the attempt context ends it closes them,
// which interrupts blocked reads and writes, and refuses later dials. close
// releases whatever the runner left open when it returns.
type scope struct {
	ctx     context.Context
	inner   measurement.Dialer
	started atomic.Bool
	stop    func() bool

	mu     sync.Mutex
	closed bool
	conns  map[*scopedConn]struct{}
}

func newScope(ctx context.Context, d measurement.Dialer) *scope {
	s := &scope{ctx: ctx, inner: d, conns: make(map[*scopedConn]struct{})}
	s.stop = context.AfterFunc(ctx, s.release)
	return s
}

// Start marks the test as started. It is idempotent.
func (s *scope) Start() { s.started.Store(true) }

func (s *scope) Started() bool { return s.started.Load() }

// DialContext starts the test and dials through the path's dialer. The dial
// ends with the attempt even when the runner passes another context.
func (s *scope) DialContext(ctx context.Context, network, address string) (measurement.Conn, error) {
	if s.isClosed() || s.ctx.Err() != nil {
		return nil, errScopeClosed
	}
	s.Start()
	dctx, cancel := context.WithCancel(ctx)
	defer cancel()
	unlink := context.AfterFunc(s.ctx, cancel)
	defer unlink()
	conn, err := s.inner.DialContext(dctx, network, address)
	if err != nil {
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

// scopedConn removes itself from its scope when closed. Close is idempotent.
type scopedConn struct {
	measurement.Conn
	scope *scope
	once  sync.Once
	err   error
}

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
