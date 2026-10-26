package execution_test

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/execution"
	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

// runnerFunc adapts a test function to measurement.Runner.
type runnerFunc func(ctx context.Context, in measurement.Input) (measurement.Output, error)

func (f runnerFunc) Run(ctx context.Context, in measurement.Input) (measurement.Output, error) {
	return f(ctx, in)
}

func exchangeServer(t *testing.T, request, reply int) *loopback.Server {
	t.Helper()
	srv, err := loopback.ListenExchange(request, reply)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = srv.Close() })
	return srv
}

// serverBytes waits for the server to release every socket and then reports
// the bytes it read and wrote, so the counts are final.
func serverBytes(t *testing.T, srv *loopback.Server) (in, out int) {
	t.Helper()
	waitOpen(t, srv, 0)
	return srv.Bytes()
}

// budgetCase runs Exchange on a fresh exchange server under limits and checks
// the terminal state, the metered usage and the server's own counts.
type budgetCase struct {
	name     string
	limits   task.AttemptLimits
	params   loopback.ExchangeParameters
	reply    int // server reply size; the runner reads params.Reply
	want     want
	usage    measurement.Usage
	serverIn int
	accepted int
}

func (c budgetCase) run(t *testing.T) (*fixture, execution.Report) {
	t.Helper()
	srv := exchangeServer(t, c.params.Request, c.reply)
	f := newFixture(t, schedule(1))
	f.plan.Limits = c.limits
	c.params.Address = srv.Address()
	f.plan.Parameters = c.params
	o := <-f.start(t, f.run, loopback.Exchange{})
	f.expect(t, o.rep, o.err, nil, []want{c.want}, nil)
	if got := o.rep.Attempts[0].Usage; got != c.usage {
		t.Fatalf("metered usage = %+v, want %+v", got, c.usage)
	}
	waitAccepted(t, srv, c.accepted)
	in, _ := serverBytes(t, srv)
	if in != c.serverIn || srv.Accepted() != c.accepted {
		t.Fatalf("server read %d bytes on %d connections, want %d on %d", in, srv.Accepted(), c.serverIn, c.accepted)
	}
	if f.clock.Timers() != 0 {
		t.Fatalf("armed timers after the run = %d", f.clock.Timers())
	}
	return f, o.rep
}

func limits(sent, received, connections int) task.AttemptLimits {
	return task.AttemptLimits{DurationMS: 30000, BytesSent: sent, BytesReceived: received, Connections: connections}
}

var (
	completedConnected = want{"completed", "", "connected"}
	stoppedByBudget    = want{"incomplete", "resource_limit", ""}
)

// Each attempt limit at limit-1, limit and limit+1 on real loopback sockets.
// A write that would exceed max_bytes_sent is not issued, so the server reads
// nothing at limit+1; the metered counts equal the bytes the server read.
// requirement-ref: REQ-LIMIT-0006
// requirement-ref: REQ-LIMIT-0027
// requirement-ref: REQ-LIMIT-0034
func TestSentBudgetBoundaries(t *testing.T) {
	const limit = 1000
	for _, c := range []budgetCase{
		{name: "below", limits: limits(limit, 10, 1), params: loopback.ExchangeParameters{Connections: 1, Exchanges: 1, Request: limit - 1, Reply: 1}, reply: 1,
			want: completedConnected, usage: measurement.Usage{BytesSent: limit - 1, BytesReceived: 1, Connections: 1}, serverIn: limit - 1, accepted: 1},
		{name: "at", limits: limits(limit, 10, 1), params: loopback.ExchangeParameters{Connections: 1, Exchanges: 1, Request: limit, Reply: 1}, reply: 1,
			want: completedConnected, usage: measurement.Usage{BytesSent: limit, BytesReceived: 1, Connections: 1}, serverIn: limit, accepted: 1},
		{name: "above", limits: limits(limit, 10, 1), params: loopback.ExchangeParameters{Connections: 1, Exchanges: 1, Request: limit + 1, Reply: 1}, reply: 1,
			want: stoppedByBudget, usage: measurement.Usage{Connections: 1}, serverIn: 0, accepted: 1},
	} {
		t.Run(c.name, func(t *testing.T) { c.run(t) })
	}
}

// Every read requests no more than the remaining max_bytes_received, so the
// runner can consume the limit exactly; a read with no budget left stops the
// attempt even though the server sent more. The meter counts what the stream
// returned, never more than the server wrote.
// requirement-ref: REQ-LIMIT-0006
// requirement-ref: REQ-LIMIT-0027
// requirement-ref: REQ-LIMIT-0034
func TestReceivedBudgetBoundaries(t *testing.T) {
	const limit = 1000
	for _, c := range []budgetCase{
		{name: "below", limits: limits(10, limit, 1), params: loopback.ExchangeParameters{Connections: 1, Exchanges: 1, Request: 1, Reply: limit - 1}, reply: limit - 1,
			want: completedConnected, usage: measurement.Usage{BytesSent: 1, BytesReceived: limit - 1, Connections: 1}, serverIn: 1, accepted: 1},
		{name: "at", limits: limits(10, limit, 1), params: loopback.ExchangeParameters{Connections: 1, Exchanges: 1, Request: 1, Reply: limit}, reply: limit,
			want: completedConnected, usage: measurement.Usage{BytesSent: 1, BytesReceived: limit, Connections: 1}, serverIn: 1, accepted: 1},
		{name: "above", limits: limits(10, limit, 1), params: loopback.ExchangeParameters{Connections: 1, Exchanges: 1, Request: 1, Reply: limit + 1}, reply: limit + 1,
			want: stoppedByBudget, usage: measurement.Usage{BytesSent: 1, BytesReceived: limit, Connections: 1}, serverIn: 1, accepted: 1},
	} {
		t.Run(c.name, func(t *testing.T) {
			srvOut := c.reply
			_, rep := c.run(t)
			if rep.Attempts[0].Usage.BytesReceived > srvOut {
				t.Fatalf("metered %d received bytes, server wrote at most %d", rep.Attempts[0].Usage.BytesReceived, srvOut)
			}
		})
	}
}

// Every dial counts toward max_connections before it is issued; a dial beyond
// the limit is not issued and the server accepts only the permitted ones.
// requirement-ref: REQ-LIMIT-0006
// requirement-ref: REQ-LIMIT-0027
func TestConnectionBudgetBoundaries(t *testing.T) {
	const limit = 3
	for _, c := range []budgetCase{
		{name: "below", limits: limits(100, 100, limit), params: loopback.ExchangeParameters{Connections: limit - 1, Exchanges: 1, Request: 1, Reply: 1}, reply: 1,
			want: completedConnected, usage: measurement.Usage{BytesSent: limit - 1, BytesReceived: limit - 1, Connections: limit - 1}, serverIn: limit - 1, accepted: limit - 1},
		{name: "at", limits: limits(100, 100, limit), params: loopback.ExchangeParameters{Connections: limit, Exchanges: 1, Request: 1, Reply: 1}, reply: 1,
			want: completedConnected, usage: measurement.Usage{BytesSent: limit, BytesReceived: limit, Connections: limit}, serverIn: limit, accepted: limit},
		{name: "above", limits: limits(100, 100, limit), params: loopback.ExchangeParameters{Connections: limit + 1, Exchanges: 1, Request: 1, Reply: 1}, reply: 1,
			want: stoppedByBudget, usage: measurement.Usage{BytesSent: limit, BytesReceived: limit, Connections: limit}, serverIn: limit, accepted: limit},
	} {
		t.Run(c.name, func(t *testing.T) {
			f, _ := c.run(t)
			if d := f.dialers.Issued()[0].Dials(); d != c.accepted {
				t.Fatalf("path dialer dials = %d, want %d", d, c.accepted)
			}
		})
	}
}

// Exchanges on a reused connection accumulate toward the byte limits while
// the connection counts once; the exchange that would exceed the sent budget
// is not issued.
// requirement-ref: REQ-LIMIT-0034
func TestBudgetAcrossConnectionReuse(t *testing.T) {
	reuse := loopback.ExchangeParameters{Connections: 1, Exchanges: 4, Request: 100, Reply: 50}
	for _, c := range []budgetCase{
		{name: "within", limits: limits(400, 200, 1), params: reuse, reply: 50,
			want: completedConnected, usage: measurement.Usage{BytesSent: 400, BytesReceived: 200, Connections: 1}, serverIn: 400, accepted: 1},
		{name: "fourth request over", limits: limits(399, 200, 1), params: reuse, reply: 50,
			want: stoppedByBudget, usage: measurement.Usage{BytesSent: 300, BytesReceived: 150, Connections: 1}, serverIn: 300, accepted: 1},
	} {
		t.Run(c.name, func(t *testing.T) { c.run(t) })
	}
}

// A failed connect still counts as a connection: with max_connections 1 a
// refused dial uses the budget, and a second dial is not issued.
// requirement-ref: REQ-LIMIT-0006
func TestFailedConnectCountsTowardConnections(t *testing.T) {
	closed := exchangeServer(t, 1, 1)
	address := closed.Address()
	_ = closed.Close()

	f := newFixture(t, schedule(1))
	f.plan.Limits = limits(10, 10, 1)
	f.plan.Parameters = loopback.ExchangeParameters{Address: address, Connections: 1, Exchanges: 1, Request: 1, Reply: 1}
	o := <-f.start(t, f.run, loopback.Exchange{})
	f.expect(t, o.rep, o.err, nil, []want{{"completed", "", "socket_error"}}, nil)
	if got := o.rep.Attempts[0].Usage; got != (measurement.Usage{Connections: 1}) {
		t.Fatalf("metered usage = %+v", got)
	}

	f = newFixture(t, schedule(1))
	f.plan.Limits = limits(10, 10, 1)
	redial := runnerFunc(func(ctx context.Context, in measurement.Input) (measurement.Output, error) {
		if _, err := in.Dialer.DialContext(ctx, "tcp", address); err == nil {
			return measurement.Output{}, measurement.ErrRunner
		}
		if _, err := in.Dialer.DialContext(ctx, "tcp", address); err == nil {
			return measurement.Output{}, measurement.ErrRunner
		}
		return refused(t).out, nil
	})
	o = <-f.start(t, f.run, redial)
	f.expect(t, o.rep, o.err, nil, []want{stoppedByBudget}, nil)
	if got := o.rep.Attempts[0].Usage; got != (measurement.Usage{Connections: 1}) || f.dialers.Issued()[0].Dials() != 1 {
		t.Fatalf("metered usage = %+v, path dials = %d", got, f.dialers.Issued()[0].Dials())
	}
}

// With zero byte and connection limits (the NET-01 case) a dial is refused
// before the path dialer is used.
// requirement-ref: REQ-LIMIT-0007
func TestZeroLimitsPermitNoTraffic(t *testing.T) {
	srv := exchangeServer(t, 1, 1)
	f := newFixture(t, schedule(1))
	f.plan.Limits = limits(0, 0, 0)
	f.plan.Parameters = loopback.ExchangeParameters{Address: srv.Address(), Connections: 1, Exchanges: 1, Request: 1, Reply: 1}
	o := <-f.start(t, f.run, loopback.Exchange{})
	f.expect(t, o.rep, o.err, nil, []want{{"not_run", "resource_limit", ""}}, nil)
	if o.rep.Attempts[0].Usage != (measurement.Usage{}) || f.dialers.Issued()[0].Dials() != 0 || srv.Accepted() != 0 {
		t.Fatalf("usage %+v, dials %d, accepted %d", o.rep.Attempts[0].Usage, f.dialers.Issued()[0].Dials(), srv.Accepted())
	}
}

// A runner reading ahead through a large buffer is capped by the received
// budget, not by its buffer, and the bytes it read ahead count although the
// runner consumed fewer. The runner's own account is not the meter.
// requirement-ref: REQ-LIMIT-0034
func TestReadAheadIsCountedAndCapped(t *testing.T) {
	const limit = 300
	srv := exchangeServer(t, 1, 500)
	f := newFixture(t, schedule(1))
	f.plan.Limits = limits(10, limit, 1)
	f.plan.Parameters = loopback.ExchangeParameters{Address: srv.Address(), Connections: 1, Exchanges: 1, Request: 1, Reply: 10, ReadAhead: 4096}
	o := <-f.start(t, f.run, loopback.Exchange{})
	f.expect(t, o.rep, o.err, nil, []want{completedConnected}, nil)
	got := o.rep.Attempts[0].Usage.BytesReceived
	if _, out := serverBytes(t, srv); got < 10 || got > limit || got > out {
		t.Fatalf("metered %d received bytes; want 10..%d and at most the %d the server wrote", got, limit, out)
	}
}

// Each attempt has its own budget: a retried attempt after an internal error
// gets the full limits again, while an attempt stopped by resource_limit,
// which is not a retry candidate, is not retried.
// requirement-ref: REQ-LIMIT-0006
func TestEachAttemptHasItsOwnBudget(t *testing.T) {
	srv := exchangeServer(t, 600, 1)
	var mu sync.Mutex
	calls := 0
	r := runnerFunc(func(ctx context.Context, in measurement.Input) (measurement.Output, error) {
		mu.Lock()
		calls++
		n := calls
		mu.Unlock()
		out, err := loopback.Exchange{}.Run(ctx, in)
		if n == 1 && err == nil {
			return measurement.Output{}, measurement.ErrRunner
		}
		return out, err
	})
	f := newFixture(t, schedule(2, task.InternalError))
	f.plan.Limits = limits(1000, 10, 1)
	f.plan.Parameters = loopback.ExchangeParameters{Address: srv.Address(), Connections: 1, Exchanges: 1, Request: 600, Reply: 1}
	o := <-f.start(t, f.run, r)
	f.expect(t, o.rep, o.err, nil, []want{{"incomplete", "internal_error", ""}, completedConnected}, []time.Duration{35 * time.Second})
	for i, a := range o.rep.Attempts {
		if a.Usage != (measurement.Usage{BytesSent: 600, BytesReceived: 1, Connections: 1}) {
			t.Fatalf("attempt %d usage = %+v", i+1, a.Usage)
		}
	}
	if in, _ := serverBytes(t, srv); in != 1200 {
		t.Fatalf("server read %d bytes", in)
	}

	srv = exchangeServer(t, 1001, 1)
	f = newFixture(t, schedule(2, task.InternalError))
	f.plan.Limits = limits(1000, 10, 1)
	f.plan.Parameters = loopback.ExchangeParameters{Address: srv.Address(), Connections: 1, Exchanges: 1, Request: 1001, Reply: 1}
	o = <-f.start(t, f.run, loopback.Exchange{})
	f.expect(t, o.rep, o.err, nil, []want{stoppedByBudget}, nil)
	if len(f.dialers.Issued()) != 1 {
		t.Fatalf("dialers issued = %d", len(f.dialers.Issued()))
	}
}

// Once the budget is exhausted no further I/O is dispatched: later writes on
// another open connection and later dials are refused without reaching the
// path, no other dialer is requested, every socket is released, and an
// outcome the runner returns afterwards does not make the attempt completed.
// requirement-ref: REQ-LIMIT-0027
// requirement-ref: REQ-LIMIT-0028
func TestExhaustedBudgetDispatchesNoFurtherIO(t *testing.T) {
	srv := exchangeServer(t, 1, 1)
	f := newFixture(t, schedule(1))
	f.plan.Limits = limits(10, 10, 3)
	var errs []error
	r := runnerFunc(func(ctx context.Context, in measurement.Input) (measurement.Output, error) {
		a, err := in.Dialer.DialContext(ctx, "tcp", srv.Address())
		if err != nil {
			return measurement.Output{}, measurement.ErrRunner
		}
		b, err := in.Dialer.DialContext(ctx, "tcp", srv.Address())
		if err != nil {
			return measurement.Output{}, measurement.ErrRunner
		}
		waitOpen(t, srv, 2)
		_, err1 := a.Write(make([]byte, 11))
		_, err2 := b.Write(make([]byte, 1))
		_, err3 := b.Read(make([]byte, 1))
		_, err4 := in.Dialer.DialContext(ctx, "tcp", srv.Address())
		errs = []error{err1, err2, err3, err4}
		return connected(t).out, nil
	})
	o := <-f.start(t, f.run, r)
	f.expect(t, o.rep, o.err, nil, []want{stoppedByBudget}, nil)
	for i, err := range errs {
		if err == nil {
			t.Fatalf("operation %d after exhaustion succeeded", i+1)
		}
	}
	waitAccepted(t, srv, 2)
	if in, out := serverBytes(t, srv); in != 0 || out != 0 || srv.Accepted() != 2 {
		t.Fatalf("server read %d, wrote %d, accepted %d", in, out, srv.Accepted())
	}
	if len(f.dialers.Issued()) != 1 || f.dialers.Issued()[0].Dials() != 2 {
		t.Fatal("budget handling used another dialer or dialed again")
	}
	if u := o.rep.Attempts[0].Usage; u != (measurement.Usage{Connections: 2}) {
		t.Fatalf("metered usage = %+v", u)
	}
}

// Exhaustion interrupts a read blocked on another connection and releases
// every socket within the attempt, as host cancellation and the deadline do.
// requirement-ref: REQ-LIMIT-0008
// requirement-ref: REQ-LIMIT-0028
func TestExhaustionReleasesBlockedSockets(t *testing.T) {
	srv := holdServer(t)
	f := newFixture(t, schedule(1))
	f.plan.Limits = limits(2*len(loopback.Payload), 10, 2)
	r := runnerFunc(func(ctx context.Context, in measurement.Input) (measurement.Output, error) {
		a, err := in.Dialer.DialContext(ctx, "tcp", srv.Address())
		if err != nil {
			return measurement.Output{}, measurement.ErrRunner
		}
		if _, err := a.Write(loopback.Payload); err != nil {
			return measurement.Output{}, measurement.ErrRunner
		}
		blocked := make(chan error, 1)
		go func() {
			_, err := io.ReadFull(a, make([]byte, 1))
			blocked <- err
		}()
		b, err := in.Dialer.DialContext(ctx, "tcp", srv.Address())
		if err != nil {
			return measurement.Output{}, measurement.ErrRunner
		}
		waitOpen(t, srv, 2)
		_, _ = b.Write(make([]byte, len(loopback.Payload)+1))
		if err := <-blocked; err == nil {
			t.Error("blocked read returned no error")
		}
		return measurement.Output{}, measurement.ErrRunner
	})
	o := <-f.start(t, f.run, r)
	f.expect(t, o.rep, o.err, nil, []want{stoppedByBudget}, nil)
	waitOpen(t, srv, 0)
	if f.clock.Timers() != 0 {
		t.Fatalf("armed timers after the run = %d", f.clock.Timers())
	}
}

// partialDialer returns a scripted connection so partial transfers are exact.
// It complements the loopback cases and is not runtime evidence on its own.
type partialDialer struct{ conn *partialConn }

func (d partialDialer) Dialer(int) (measurement.Dialer, error) { return d, nil }
func (d partialDialer) DialContext(context.Context, string, string) (measurement.Conn, error) {
	return d.conn, nil
}

type partialConn struct {
	written   int   // bytes a Write accepts before failing
	available int   // bytes a Read returns at most
	requested []int // len(p) of every Read
}

var errShort = errors.New("synthetic short transfer")

func (c *partialConn) Write(p []byte) (int, error) {
	if len(p) > c.written {
		return c.written, errShort
	}
	return len(p), nil
}
func (c *partialConn) Read(p []byte) (int, error) {
	c.requested = append(c.requested, len(p))
	return min(len(p), c.available), nil
}
func (*partialConn) Close() error                { return nil }
func (*partialConn) SetDeadline(time.Time) error { return nil }

// A partial write counts only the accepted bytes and returns the rest of its
// reservation; a read asks the stream for no more than the remaining budget
// and counts what the stream returned.
// requirement-ref: REQ-LIMIT-0034
func TestPartialTransfersAreCountedExactly(t *testing.T) {
	conn := &partialConn{written: 3, available: 4}
	f := newFixture(t, schedule(1))
	f.plan.Dialers = partialDialer{conn}
	f.plan.Limits = limits(10, 10, 1)
	r := runnerFunc(func(ctx context.Context, in measurement.Input) (measurement.Output, error) {
		c, err := in.Dialer.DialContext(ctx, "tcp", "127.0.0.1:9")
		if err != nil {
			return measurement.Output{}, measurement.ErrRunner
		}
		if n, err := c.Write(make([]byte, 8)); n != 3 || !errors.Is(err, errShort) {
			return measurement.Output{}, measurement.ErrRunner
		}
		if n, err := c.Write(make([]byte, 7)); n != 3 || err == nil { // 7 fits in the remaining 7
			return measurement.Output{}, measurement.ErrRunner
		}
		for range 2 {
			if _, err := c.Read(make([]byte, 100)); err != nil {
				return measurement.Output{}, measurement.ErrRunner
			}
		}
		return connected(t).out, nil
	})
	o := <-f.start(t, f.run, r)
	f.expect(t, o.rep, o.err, nil, []want{completedConnected}, nil)
	if u := o.rep.Attempts[0].Usage; u != (measurement.Usage{BytesSent: 6, BytesReceived: 8, Connections: 1}) {
		t.Fatalf("metered usage = %+v", u)
	}
	if len(conn.requested) != 2 || conn.requested[0] != 10 || conn.requested[1] != 6 {
		t.Fatalf("read sizes requested from the stream = %v, want [10 6]", conn.requested)
	}
}

// Effective byte and connection limits beyond the registered maxima, or
// negative, make the plan invalid; the engine never clamps them.
// requirement-ref: REQ-LIMIT-0006
func TestPlanLimitsAreBounded(t *testing.T) {
	for _, l := range []task.AttemptLimits{
		limits(execution.MaxBytesSent+1, 10, 1), limits(10, execution.MaxBytesReceived+1, 1), limits(10, 10, execution.MaxConnections+1),
		limits(-1, 10, 1), limits(10, -1, 1), limits(10, 10, -1),
	} {
		f := newFixture(t, schedule(1), connected(t))
		f.plan.Limits = l
		if _, err := f.engine.Run(f.run, f.plan); !errors.Is(err, execution.ErrPlan) || f.runner.Calls() != 0 {
			t.Fatalf("limits %+v: %v", l, err)
		}
	}
	f := newFixture(t, schedule(1), connected(t))
	f.plan.Limits = limits(execution.MaxBytesSent, execution.MaxBytesReceived, execution.MaxConnections)
	rep, err := f.engine.Run(f.run, f.plan)
	f.expect(t, rep, err, nil, []want{completedConnected}, nil)
}
