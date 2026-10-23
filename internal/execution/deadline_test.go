package execution_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/execution"
	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/networkcontext"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

// gated is a synthetic runner that signals entry, optionally starts the test,
// and then waits for its context or for release. It performs no I/O.
type gated struct {
	begin   bool
	entered chan context.Context
	release chan struct{}
	out     measurement.Output
}

func newGated(t *testing.T, begin bool) *gated {
	return &gated{begin: begin, entered: make(chan context.Context, 1), release: make(chan struct{}), out: connected(t).out}
}

func (g *gated) Run(ctx context.Context, in measurement.Input) (measurement.Output, error) {
	if g.begin {
		in.Start()
	}
	g.entered <- ctx
	select {
	case <-ctx.Done():
		return measurement.Output{}, measurement.ErrRunner
	case <-g.release:
		return g.out, nil
	}
}

// blockingCapture waits in context capture until its context ends.
type blockingCapture struct{ entered chan struct{} }

func (b blockingCapture) Capture(ctx context.Context) (networkcontext.Snapshot, error) {
	b.entered <- struct{}{}
	<-ctx.Done()
	return networkcontext.Snapshot{}, ctx.Err()
}

// blockingDialer waits in its dial until the dial context ends.
type blockingDialer struct{ entered chan struct{} }

func (d blockingDialer) DialContext(ctx context.Context, _, _ string) (measurement.Conn, error) {
	d.entered <- struct{}{}
	<-ctx.Done()
	return nil, ctx.Err()
}

type blockingDialers struct{ blockingDialer }

func (d blockingDialers) Dialer(int) (measurement.Dialer, error) { return d.blockingDialer, nil }

type outcome struct {
	rep execution.Report
	err error
}

// start runs the engine on its own goroutine with runner r.
func (f *fixture) start(t *testing.T, ctx context.Context, r measurement.Runner) <-chan outcome {
	t.Helper()
	engine, err := execution.NewEngine(newRegistry(t, r), f.clock, fixedJitter(5*time.Second), f.store)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan outcome, 1)
	go func() {
		rep, err := engine.Run(ctx, f.plan)
		done <- outcome{rep, err}
	}()
	return done
}

// The attempt deadline is max_duration_ms on the engine clock: the attempt
// context is live one millisecond before it and ends at and after it, with
// attempt_deadline_exceeded; an attempt finishing before it is completed and
// leaves no timer armed.
// requirement-ref: REQ-LIMIT-0006
// requirement-ref: REQ-RESULT-0051
func TestAttemptDeadlineBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name    string
		advance time.Duration
		expired bool
	}{
		{"before", attemptDeadline - time.Millisecond, false},
		{"at", attemptDeadline, true},
		{"after", attemptDeadline + time.Millisecond, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t, schedule(1))
			g := newGated(t, true)
			done := f.start(t, f.run, g)
			ctx := <-g.entered
			_ = f.clock.Advance(tc.advance)
			if expired := ctx.Err() != nil; expired != tc.expired {
				t.Fatalf("attempt context ended = %v after %v", expired, tc.advance)
			}
			if !tc.expired {
				close(g.release)
			}
			o := <-done
			if tc.expired {
				f.expect(t, o.rep, o.err, nil, []want{{"incomplete", "attempt_deadline_exceeded", ""}}, nil)
			} else {
				f.expect(t, o.rep, o.err, nil, []want{{"completed", "", "connected"}}, nil)
			}
			if f.clock.Timers() != 0 {
				t.Fatalf("armed timers after the run = %d", f.clock.Timers())
			}
		})
	}
}

// The deadline covers the whole attempt: expiry before the test starts, in
// context capture or in the runner before its first network operation, is
// not_run without Network Context; after the start it is incomplete.
// requirement-ref: REQ-LIMIT-0006
// requirement-ref: REQ-RESULT-0051
// requirement-ref: REQ-RESULT-0053
func TestAttemptDeadlineBeforeTheTestStarts(t *testing.T) {
	f := newFixture(t, schedule(1))
	capture := blockingCapture{entered: make(chan struct{}, 1)}
	f.plan.Context = capture
	g := newGated(t, true)
	done := f.start(t, f.run, g)
	<-capture.entered
	_ = f.clock.Advance(attemptDeadline)
	o := <-done
	f.expect(t, o.rep, o.err, nil, []want{{"not_run", "attempt_deadline_exceeded", ""}}, nil)

	f = newFixture(t, schedule(1))
	g = newGated(t, false)
	done = f.start(t, f.run, g)
	<-g.entered
	_ = f.clock.Advance(attemptDeadline)
	o = <-done
	f.expect(t, o.rep, o.err, nil, []want{{"not_run", "attempt_deadline_exceeded", ""}}, nil)
}

// The first termination is kept: a deadline that expired before a host
// cancellation stays the reason, and the reverse.
// requirement-ref: REQ-RESULT-0021
func TestFirstTerminationIsKept(t *testing.T) {
	f := newFixture(t, schedule(1))
	run, cancel := context.WithCancel(context.Background())
	g := newGated(t, true)
	done := f.start(t, run, g)
	<-g.entered
	_ = f.clock.Advance(attemptDeadline)
	cancel()
	o := <-done
	f.expect(t, o.rep, o.err, nil, []want{{"incomplete", "attempt_deadline_exceeded", ""}}, nil)

	f = newFixture(t, schedule(1))
	run, cancel = context.WithCancel(context.Background())
	g = newGated(t, true)
	done = f.start(t, run, g)
	<-g.entered
	cancel()
	_ = f.clock.Advance(attemptDeadline)
	o = <-done
	f.expect(t, o.rep, o.err, nil, []want{{"incomplete", "cancelled_by_host", ""}}, nil)
}

// Host cancellation at each stage of an active attempt. Stages before the test
// starts are not_run; stages after it are incomplete.
// requirement-ref: REQ-RESULT-0051
func TestCancellationAtEveryStage(t *testing.T) {
	t.Run("context capture", func(t *testing.T) {
		f := newFixture(t, schedule(1))
		capture := blockingCapture{entered: make(chan struct{}, 1)}
		f.plan.Context = capture
		run, cancel := context.WithCancel(context.Background())
		done := f.start(t, run, newGated(t, true))
		<-capture.entered
		cancel()
		o := <-done
		f.expect(t, o.rep, o.err, nil, []want{{"not_run", "cancelled_by_host", ""}}, nil)
	})
	t.Run("before the first network operation", func(t *testing.T) {
		f := newFixture(t, schedule(1))
		run, cancel := context.WithCancel(context.Background())
		g := newGated(t, false)
		done := f.start(t, run, g)
		<-g.entered
		cancel()
		o := <-done
		f.expect(t, o.rep, o.err, nil, []want{{"not_run", "cancelled_by_host", ""}}, nil)
	})
	t.Run("during the dial", func(t *testing.T) {
		f := newFixture(t, schedule(1))
		d := blockingDialers{blockingDialer{entered: make(chan struct{}, 1)}}
		f.plan.Dialers = d
		f.plan.Parameters = loopback.Parameters{Address: "127.0.0.1:9"}
		run, cancel := context.WithCancel(context.Background())
		done := f.start(t, run, loopback.Runner{})
		<-d.entered
		cancel()
		o := <-done
		f.expect(t, o.rep, o.err, nil, []want{{"incomplete", "cancelled_by_host", ""}}, nil)
	})
	t.Run("during I/O on a real socket", func(t *testing.T) {
		srv := holdServer(t)
		f := newFixture(t, schedule(1))
		f.plan.Parameters = loopback.Parameters{Address: srv.Address()}
		run, cancel := context.WithCancel(context.Background())
		done := f.start(t, run, loopback.Runner{})
		waitOpen(t, srv, 1)
		cancel()
		o := <-done
		f.expect(t, o.rep, o.err, nil, []want{{"incomplete", "cancelled_by_host", ""}}, nil)
		waitOpen(t, srv, 0)
	})
	t.Run("at the deadline during I/O on a real socket", func(t *testing.T) {
		srv := holdServer(t)
		f := newFixture(t, schedule(1))
		f.plan.Parameters = loopback.Parameters{Address: srv.Address()}
		done := f.start(t, f.run, loopback.Runner{})
		waitOpen(t, srv, 1)
		_ = f.clock.Advance(attemptDeadline)
		o := <-done
		f.expect(t, o.rep, o.err, nil, []want{{"incomplete", "attempt_deadline_exceeded", ""}}, nil)
		waitOpen(t, srv, 0)
	})
}

// leaky dials, writes and returns an outcome without closing its connection;
// optionally it first cancels the run and then tries to dial.
type leaky struct {
	t      *testing.T
	cancel context.CancelFunc
}

func (l leaky) Run(ctx context.Context, in measurement.Input) (measurement.Output, error) {
	if l.cancel != nil {
		l.cancel()
	}
	conn, err := in.Dialer.DialContext(ctx, "tcp", in.Parameters.(loopback.Parameters).Address)
	if err != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	if _, err := conn.Write(loopback.Payload); err != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	return connected(l.t).out, nil
}

// The attempt scope releases a socket the runner left open when the runner
// returns, and refuses a dial once the attempt has ended.
// requirement-ref: REQ-LIMIT-0008
func TestAttemptScopeReleasesSockets(t *testing.T) {
	srv := holdServer(t)
	f := newFixture(t, schedule(1))
	f.plan.Parameters = loopback.Parameters{Address: srv.Address()}
	o := <-f.start(t, f.run, leaky{t: t})
	f.expect(t, o.rep, o.err, nil, []want{{"completed", "", "connected"}}, nil)
	waitAccepted(t, srv, 1)
	waitOpen(t, srv, 0)
	if srv.Accepted() != 1 {
		t.Fatalf("accepted = %d", srv.Accepted())
	}

	f = newFixture(t, schedule(1))
	f.plan.Parameters = loopback.Parameters{Address: srv.Address()}
	run, cancel := context.WithCancel(context.Background())
	o = <-f.start(t, run, leaky{t: t, cancel: cancel})
	f.expect(t, o.rep, o.err, nil, []want{{"not_run", "cancelled_by_host", ""}}, nil)
	if srv.Accepted() != 1 || len(f.dialers.Issued()) != 1 || f.dialers.Issued()[0].Dials() != 0 {
		t.Fatal("dial issued after the attempt ended")
	}
}

// A cancellation racing with completion keeps whichever terminal transition
// committed first; exactly one result is written and it is never reopened.
// requirement-ref: REQ-RESULT-0052
func TestCancellationRacesWithCompletion(t *testing.T) {
	seen := map[want]int{}
	for i := range 200 {
		f := newFixture(t, schedule(1))
		run, cancel := context.WithCancel(context.Background())
		g := newGated(t, true)
		close(g.release)
		done := f.start(t, run, g)
		if i%2 == 1 {
			// Cancel while the runner chooses between its outcome and the
			// cancellation, rather than before the run.
			<-g.entered
		}
		cancel()
		o := <-done
		recs := f.records(t)
		if o.err != nil || len(recs) != 1 || len(o.rep.Attempts) != 1 {
			t.Fatalf("Run() = %+v, %v, records %d", o.rep, o.err, len(recs))
		}
		r := recs[0]
		got := want{r.Execution.State, r.Execution.Reason, ""}
		if r.Measurement != nil {
			got.outcome = r.Measurement.Outcome
		}
		switch got {
		case want{"completed", "", "connected"}, want{"incomplete", "cancelled_by_host", ""}, want{"not_run", "cancelled_by_host", ""}:
		default:
			t.Fatalf("raced result = %+v", got)
		}
		phase := map[string]execution.Phase{"completed": execution.Completed, "incomplete": execution.Incomplete, "not_run": execution.NotRun}[got.state]
		if o.rep.Attempts[0].Attempt.Phase() != phase {
			t.Fatalf("report phase %v, record %+v", o.rep.Attempts[0].Attempt.Phase(), got)
		}
		seen[got]++
	}
	t.Logf("raced outcomes: %v", seen)
}

func TestPlanDeadlineIsBounded(t *testing.T) {
	for _, d := range []time.Duration{0, -time.Millisecond, execution.MaxAttemptDuration + time.Millisecond} {
		f := newFixture(t, schedule(1), connected(t))
		f.plan.Deadline = d
		if _, err := f.engine.Run(f.run, f.plan); !errors.Is(err, execution.ErrPlan) || f.runner.Calls() != 0 {
			t.Fatalf("deadline %v: %v", d, err)
		}
	}
	f := newFixture(t, schedule(1), connected(t))
	f.plan.Deadline = execution.MaxAttemptDuration
	rep, err := f.engine.Run(f.run, f.plan)
	f.expect(t, rep, err, nil, []want{{"completed", "", "connected"}}, nil)
}

func holdServer(t *testing.T) *loopback.Server {
	t.Helper()
	srv, err := loopback.ListenHold()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = srv.Close() })
	return srv
}

// waitOpen waits for the server to observe n open connections. The timeout
// only bounds a failing test; no assertion depends on elapsed time.
// waitAccepted waits until the server has accepted n connections, so later
// accept counts and server-side byte counts are final rather than racing the
// server's accept loop.
func waitAccepted(t *testing.T, srv *loopback.Server, n int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.WaitAccepted(ctx, n); err != nil {
		t.Fatalf("server accepted %d connections, want %d", srv.Accepted(), n)
	}
}

func waitOpen(t *testing.T, srv *loopback.Server, n int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.WaitOpen(ctx, n); err != nil {
		t.Fatalf("server open connections = %d, want %d", srv.Open(), n)
	}
}
