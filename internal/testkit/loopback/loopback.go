// Package loopback is a test-only runner that performs real TCP I/O against a
// local echo server on 127.0.0.1, so execution tests can exercise real sockets
// for cancellation and budgets. It reports a tcp_connect observation in the
// TCP-01 output shape only so result construction accepts it; it is not a
// TCP-01 implementation and not conformance evidence. Production code must not
// import it, and it is registered only by tests.
package loopback

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/observation"
)

// Payload is the fixed synthetic message the runner sends and expects echoed.
var Payload = []byte("barometer-loopback")

// Server accepts loopback connections and echoes at most len(Payload) bytes on
// each. A holding server reads the payload and then keeps the connection open
// without echoing until the client closes it, so a runner blocks in its read
// and only the client side can release the socket. An exchange server answers
// every request of a fixed size with a reply of a fixed size until the client
// closes. Every server counts the bytes it reads from and writes to its
// sockets, so a test compares a client-side meter with the bytes that crossed
// the socket rather than with the client's own account.
type Server struct {
	ln             net.Listener
	hold           bool
	request, reply int
	wg             sync.WaitGroup
	mu             sync.Mutex
	conn           int
	open           int
	in, out        int
	changed        chan struct{} // closed and replaced whenever open changes
}

// MaxExchangeBytes bounds an exchange server's request and reply sizes.
const MaxExchangeBytes = 1 << 20

// Listen starts a server on an ephemeral 127.0.0.1 port.
func Listen() (*Server, error) { return listen(false, 0, 0) }

// ListenHold starts a holding server on an ephemeral 127.0.0.1 port.
func ListenHold() (*Server, error) { return listen(true, 0, 0) }

// ListenExchange starts an exchange server on an ephemeral 127.0.0.1 port.
func ListenExchange(request, reply int) (*Server, error) {
	if request < 1 || request > MaxExchangeBytes || reply < 0 || reply > MaxExchangeBytes {
		return nil, errors.New("loopback: invalid exchange sizes")
	}
	return listen(false, request, reply)
}

func listen(hold bool, request, reply int) (*Server, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	s := &Server{ln: ln, hold: hold, request: request, reply: reply, changed: make(chan struct{})}
	s.wg.Add(1)
	go s.serve()
	return s, nil
}

func (s *Server) serve() {
	defer s.wg.Done()
	for {
		c, err := s.ln.Accept()
		if err != nil {
			return
		}
		s.mu.Lock()
		s.conn++
		s.mu.Unlock()
		s.track(1) // also signals the accept to WaitAccepted
		s.wg.Add(1)
		go func(raw net.Conn) {
			defer s.wg.Done()
			defer s.track(-1)
			defer raw.Close()
			c := counted{raw, s}
			if s.request > 0 {
				s.exchange(c)
				return
			}
			buf := make([]byte, len(Payload))
			n, _ := io.ReadFull(c, buf)
			if s.hold {
				// Returns only when the client closes its side.
				_, _ = io.Copy(io.Discard, c)
				return
			}
			_, _ = c.Write(buf[:n])
		}(c)
	}
}

// exchange answers each complete request with one reply until the client
// closes or a request is cut short.
func (s *Server) exchange(c io.ReadWriter) {
	req, rep := make([]byte, s.request), make([]byte, s.reply)
	for {
		if _, err := io.ReadFull(c, req); err != nil {
			return
		}
		if _, err := c.Write(rep); err != nil {
			return
		}
	}
}

// counted records every byte the server reads or writes on a socket.
type counted struct {
	net.Conn
	s *Server
}

func (c counted) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	c.s.mu.Lock()
	c.s.in += n
	c.s.mu.Unlock()
	return n, err
}

func (c counted) Write(p []byte) (int, error) {
	n, err := c.Conn.Write(p)
	c.s.mu.Lock()
	c.s.out += n
	c.s.mu.Unlock()
	return n, err
}

// Bytes reports how many bytes the server read from and wrote to its sockets.
func (s *Server) Bytes() (in, out int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.in, s.out
}

func (s *Server) track(delta int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.open += delta
	close(s.changed)
	s.changed = make(chan struct{})
}

// Open reports how many accepted connections the server still holds.
func (s *Server) Open() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.open
}

// WaitOpen blocks until exactly n accepted connections are open or ctx ends.
// It observes the server side, so a return of nil for n == 0 shows that the
// client released every socket.
func (s *Server) WaitOpen(ctx context.Context, n int) error {
	for {
		s.mu.Lock()
		open, changed := s.open, s.changed
		s.mu.Unlock()
		if open == n {
			return nil
		}
		select {
		case <-changed:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// WaitAccepted blocks until the server has accepted at least n connections or
// ctx ends. A client's dial can complete before the server accepts, so a test
// that asserts accept counts or server-side bytes waits for the accepts first.
func (s *Server) WaitAccepted(ctx context.Context, n int) error {
	for {
		s.mu.Lock()
		accepted, changed := s.conn, s.changed
		s.mu.Unlock()
		if accepted >= n {
			return nil
		}
		select {
		case <-changed:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// Address is the server's host:port.
func (s *Server) Address() string { return s.ln.Addr().String() }

// Accepted reports how many connections the server accepted.
func (s *Server) Accepted() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conn
}

// Close stops accepting and waits for open connections to finish.
func (s *Server) Close() error {
	err := s.ln.Close()
	s.wg.Wait()
	return err
}

// Parameters is the runner's synthetic normalized input.
type Parameters struct{ Address string }

func (Parameters) TaskParameters() {}

// Dialer adapts net.Dialer to the measurement seam and counts dials.
type Dialer struct {
	mu    sync.Mutex
	dials int
}

func (d *Dialer) DialContext(ctx context.Context, network, address string) (measurement.Conn, error) {
	d.mu.Lock()
	d.dials++
	d.mu.Unlock()
	var nd net.Dialer
	return nd.DialContext(ctx, network, address)
}

// Dials reports how many connections were attempted through this dialer.
func (d *Dialer) Dials() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.dials
}

// Dialers supplies a fresh Dialer per attempt and keeps them for inspection.
type Dialers struct {
	mu      sync.Mutex
	issued  []*Dialer
	FailFor map[int]bool
}

func (s *Dialers) Dialer(attempt int) (measurement.Dialer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.FailFor[attempt] {
		return nil, errors.New("loopback: injected dialer failure")
	}
	d := &Dialer{}
	s.issued = append(s.issued, d)
	return d, nil
}

// Issued returns the dialers handed out, in attempt order.
func (s *Dialers) Issued() []*Dialer {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]*Dialer(nil), s.issued...)
}

// Runner connects, sends Payload, reads the echo and closes. A refused or
// failed connect is a terminal observation; a failure after connecting is an
// execution failure (measurement.ErrRunner).
type Runner struct{}

func (Runner) Run(ctx context.Context, in measurement.Input) (measurement.Output, error) {
	p, ok := in.Parameters.(Parameters)
	if !ok || in.Dialer == nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	start := time.Now()
	conn, err := in.Dialer.DialContext(ctx, "tcp", p.Address)
	usage := measurement.Usage{Connections: 1}
	if err != nil {
		if ctx.Err() != nil {
			return measurement.Output{Usage: usage}, measurement.ErrRunner
		}
		o, err := tcpConnect(time.Since(start), "socket_error")
		if err != nil {
			return measurement.Output{Usage: usage}, measurement.ErrRunner
		}
		return measurement.Output{Outcome: "socket_error", Observations: []observation.Observation{o}, Usage: usage}, nil
	}
	defer conn.Close()
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}
	n, err := conn.Write(Payload)
	usage.BytesSent = n
	if err != nil {
		return measurement.Output{Usage: usage}, measurement.ErrRunner
	}
	echo := make([]byte, len(Payload))
	m, err := io.ReadFull(conn, echo)
	usage.BytesReceived = m
	if err != nil || string(echo) != string(Payload) {
		return measurement.Output{Usage: usage}, measurement.ErrRunner
	}
	o, err := tcpConnect(time.Since(start), "")
	if err != nil {
		return measurement.Output{Usage: usage}, measurement.ErrRunner
	}
	return measurement.Output{Outcome: "connected", Observations: []observation.Observation{o}, Usage: usage}, nil
}

// tcpConnect builds the observation through the strict observation decoder.
func tcpConnect(elapsed time.Duration, failure string) (observation.Observation, error) {
	ms := int((elapsed + time.Millisecond - 1) / time.Millisecond)
	ms = (ms + 249) / 250 * 250
	status, f := "succeeded", ""
	if failure != "" {
		status, f = "failed", `,"failure":"`+failure+`"`
	}
	return observation.DecodeProtocol(observation.Contract, []byte(`{"type":"tcp_connect","status":"`+status+`","duration_ms":`+strconv.Itoa(ms)+f+`,"details":{"address_family":"ipv4"}}`), observation.ProtocolLimits{})
}

// ExchangeParameters configure Exchange. Connections are opened one after
// another, and each carries Exchanges request/reply exchanges, so more than one
// exchange reuses the connection. Each request is one Write of Request bytes;
// each reply is read with io.ReadFull for Reply bytes, through a bufio.Reader
// of ReadAhead bytes when ReadAhead is positive.
type ExchangeParameters struct {
	Address                                           string
	Connections, Exchanges, Request, Reply, ReadAhead int
}

func (ExchangeParameters) TaskParameters() {}

// Exchange is a test-only runner for an exchange server. A failed connect is a
// terminal socket_error observation; any later failure, including a refused
// metered operation, is an execution failure. Its Usage is its own account of
// the bytes it consumed, which a meter must not rely on. Its observation
// reports a zero duration so that it fits within the attempt timing of a
// manual engine clock.
type Exchange struct{}

func (Exchange) Run(ctx context.Context, in measurement.Input) (measurement.Output, error) {
	p, ok := in.Parameters.(ExchangeParameters)
	if !ok || in.Dialer == nil || p.Connections < 1 || p.Exchanges < 1 || p.Request < 1 || p.Reply < 0 ||
		p.Request > MaxExchangeBytes || p.Reply > MaxExchangeBytes || p.ReadAhead < 0 {
		return measurement.Output{}, measurement.ErrRunner
	}
	var usage measurement.Usage
	for range p.Connections {
		conn, err := in.Dialer.DialContext(ctx, "tcp", p.Address)
		usage.Connections++
		if err != nil {
			if ctx.Err() != nil {
				return measurement.Output{Usage: usage}, measurement.ErrRunner
			}
			o, err := tcpConnect(0, "socket_error")
			if err != nil {
				return measurement.Output{Usage: usage}, measurement.ErrRunner
			}
			return measurement.Output{Outcome: "socket_error", Observations: []observation.Observation{o}, Usage: usage}, nil
		}
		if err := exchange(ctx, conn, p, &usage); err != nil {
			return measurement.Output{Usage: usage}, measurement.ErrRunner
		}
	}
	o, err := tcpConnect(0, "")
	if err != nil {
		return measurement.Output{Usage: usage}, measurement.ErrRunner
	}
	return measurement.Output{Outcome: "connected", Observations: []observation.Observation{o}, Usage: usage}, nil
}

func exchange(ctx context.Context, conn measurement.Conn, p ExchangeParameters, usage *measurement.Usage) error {
	defer conn.Close()
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}
	var r io.Reader = conn
	if p.ReadAhead > 0 {
		r = bufio.NewReaderSize(conn, p.ReadAhead)
	}
	req, rep := make([]byte, p.Request), make([]byte, p.Reply)
	for range p.Exchanges {
		n, err := conn.Write(req)
		usage.BytesSent += n
		if err != nil {
			return err
		}
		m, err := io.ReadFull(r, rep)
		usage.BytesReceived += m
		if err != nil {
			return err
		}
	}
	return nil
}
