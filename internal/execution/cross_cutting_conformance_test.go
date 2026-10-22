package execution_test

import (
	"context"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

// Runtime conformance family cases on the engine (DEV-640). They use the
// manual engine clock, the output-port fake and, for budgets, real TCP sockets
// on 127.0.0.1 through the test-only loopback dialer and exchange servers. They
// are not registered measurement, runtime, platform or field evidence.

// exchangeOn runs Exchange against another server within the same attempt.
func exchangeOn(ctx context.Context, in measurement.Input, p loopback.ExchangeParameters) (measurement.Output, error) {
	return loopback.Exchange{}.Run(ctx, measurement.Input{Parameters: p, Dialer: in.Dialer, Start: in.Start})
}

// CC-EXECUTION-FIELDS-LIMITS: attempt limits apply to each attempt. Attempt 1
// consumes the registered default budgets exactly and ends with a retryable
// reason; attempt 2 of the same run starts with full budgets, one connection
// or byte beyond either limit stops it with resource_limit, and its deadline
// runs from its own start.
// requirement-ref: REQ-EXECUTION-0011
func TestCcExecutionFieldsLimits(t *testing.T) {
	full := loopback.ExchangeParameters{Connections: 8, Exchanges: 1, Request: 4096, Reply: 16384}
	fullUsage := measurement.Usage{BytesSent: 32768, BytesReceived: 131072, Connections: 8}
	srv := exchangeServer(t, 4096, 16384)
	full.Address = srv.Address()
	// short is a server whose reply is one byte larger than its request share
	// allows, for the received-bytes boundary.
	short := exchangeServer(t, 4096, 16385)
	cases := []struct {
		name   string
		second func(ctx context.Context, in measurement.Input) (measurement.Output, error)
		want   want
		usage  measurement.Usage
	}{
		{"full budgets again", func(ctx context.Context, in measurement.Input) (measurement.Output, error) {
			return exchangeOn(ctx, in, full)
		}, completedConnected, fullUsage},
		{"a ninth connection", func(ctx context.Context, in measurement.Input) (measurement.Output, error) {
			if _, err := exchangeOn(ctx, in, full); err != nil {
				return measurement.Output{}, err
			}
			if _, err := in.Dialer.DialContext(ctx, "tcp", srv.Address()); err != nil {
				return measurement.Output{}, measurement.ErrRunner
			}
			return connected(t).out, nil
		}, stoppedByBudget, fullUsage},
		{"one byte beyond max_bytes_sent", func(ctx context.Context, in measurement.Input) (measurement.Output, error) {
			seven := full
			seven.Connections = 7
			if _, err := exchangeOn(ctx, in, seven); err != nil {
				return measurement.Output{}, err
			}
			c, err := in.Dialer.DialContext(ctx, "tcp", srv.Address())
			if err != nil {
				return measurement.Output{}, measurement.ErrRunner
			}
			if _, err := c.Write(make([]byte, 4097)); err != nil {
				return measurement.Output{}, measurement.ErrRunner
			}
			return connected(t).out, nil
		}, stoppedByBudget, measurement.Usage{BytesSent: 28672, BytesReceived: 114688, Connections: 8}},
		{"one byte beyond max_bytes_received", func(ctx context.Context, in measurement.Input) (measurement.Output, error) {
			seven := full
			seven.Connections = 7
			if _, err := exchangeOn(ctx, in, seven); err != nil {
				return measurement.Output{}, err
			}
			last := loopback.ExchangeParameters{Address: short.Address(), Connections: 1, Exchanges: 1, Request: 4096, Reply: 16385}
			return exchangeOn(ctx, in, last)
		}, stoppedByBudget, measurement.Usage{BytesSent: 32768, BytesReceived: 131072, Connections: 8}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t, schedule(2, task.InternalError))
			calls := 0
			r := runnerFunc(func(ctx context.Context, in measurement.Input) (measurement.Output, error) {
				calls++
				if calls == 1 {
					if _, err := exchangeOn(ctx, in, full); err != nil {
						return measurement.Output{}, err
					}
					_ = f.clock.Advance(20 * time.Second)
					return measurement.Output{}, measurement.ErrRunner
				}
				return tc.second(ctx, in)
			})
			o := <-f.start(t, f.run, r)
			f.expect(t, o.rep, o.err, nil, []want{{"incomplete", "internal_error", ""}, tc.want}, []time.Duration{35 * time.Second})
			if o.rep.Attempts[0].Usage != fullUsage || o.rep.Attempts[1].Usage != tc.usage {
				t.Fatalf("usage = %+v, %+v", o.rep.Attempts[0].Usage, o.rep.Attempts[1].Usage)
			}
		})
	}
	t.Run("the deadline runs from each attempt's own start", func(t *testing.T) {
		for _, tc := range []struct {
			name  string
			spend time.Duration
			want  want
		}{
			{"attempt 2 one millisecond within its deadline", attemptDeadline - time.Millisecond, completedConnected},
			{"attempt 2 at its deadline", attemptDeadline, want{"incomplete", "attempt_deadline_exceeded", ""}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				f := newFixture(t, schedule(2, task.InternalError))
				calls := 0
				r := runnerFunc(func(ctx context.Context, in measurement.Input) (measurement.Output, error) {
					calls++
					in.Start()
					if calls == 1 {
						// Attempt 1 uses 20 s; a deadline counted from it
						// would already have passed when attempt 2 starts.
						_ = f.clock.Advance(20 * time.Second)
						return measurement.Output{}, measurement.ErrRunner
					}
					_ = f.clock.Advance(tc.spend)
					if ctx.Err() != nil {
						return measurement.Output{}, ctx.Err()
					}
					return connected(t).out, nil
				})
				o := <-f.start(t, f.run, r)
				f.expect(t, o.rep, o.err, nil, []want{{"incomplete", "internal_error", ""}, tc.want}, []time.Duration{35 * time.Second})
			})
		}
	})
}

// CC-RESULT-EXECUTION-OBJECT: execution states and reasons as the engine
// writes them. An interruption before the test starts is not_run and after it
// incomplete; a recorded outcome stays completed; the first termination wins;
// measurement presence follows the state; and a plan the engine refuses
// produces no Measurement Result.
// requirement-ref: REQ-RESULT-0019
// requirement-ref: REQ-RESULT-0020
// requirement-ref: REQ-RESULT-0021
// requirement-ref: REQ-RESULT-0022
// requirement-ref: REQ-RESULT-0051
// requirement-ref: REQ-RESULT-0052
func TestCcResultExecutionObject(t *testing.T) {
	cases := []struct {
		name  string
		steps []step
		want  want
	}{
		{"completed with a positive outcome", []step{connected(t)}, completedConnected},
		{"completed with a negative outcome", []step{refused(t)}, want{"completed", "", "connection_refused"}},
		{"internal error after the test started", []step{failure()}, want{"incomplete", "internal_error", ""}},
		{"host cancellation before the test started", []step{{err: measurement.ErrRunner, cancel: true, unbegun: true}}, want{"not_run", "cancelled_by_host", ""}},
		{"host cancellation after the test started", []step{{err: measurement.ErrRunner, cancel: true}}, want{"incomplete", "cancelled_by_host", ""}},
		{"deadline before the test started", []step{{err: measurement.ErrRunner, advance: attemptDeadline, unbegun: true}}, want{"not_run", "attempt_deadline_exceeded", ""}},
		{"deadline after the test started", []step{{err: measurement.ErrRunner, advance: attemptDeadline}}, want{"incomplete", "attempt_deadline_exceeded", ""}},
		{"cancellation after the outcome was recorded", []step{{out: connected(t).out, cancel: true}}, completedConnected},
		{"deadline after the outcome was recorded", []step{{out: connected(t).out, advance: attemptDeadline}}, completedConnected},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t, schedule(1), tc.steps...)
			rep, err := f.engine.Run(f.run, f.plan)
			// expect also checks that completed results carry a measurement
			// and incomplete and not_run results here carry none, and that
			// not_run results carry no Network Context.
			f.expect(t, rep, err, nil, []want{tc.want}, nil)
			for _, r := range f.records(t) {
				if (r.Execution.State == "completed") != (r.Measurement != nil) || (r.Execution.State == "completed") != (r.Execution.Reason == "") {
					t.Fatalf("state %q with reason %q and measurement %v", r.Execution.State, r.Execution.Reason, r.Measurement != nil)
				}
			}
		})
	}
	t.Run("the first termination condition is the reason", func(t *testing.T) {
		// The deadline fires first, then the host cancels.
		f := newFixture(t, schedule(1), step{err: measurement.ErrRunner, advance: attemptDeadline, cancel: true})
		rep, err := f.engine.Run(f.run, f.plan)
		f.expect(t, rep, err, nil, []want{{"incomplete", "attempt_deadline_exceeded", ""}}, nil)
		// The host cancels first, then the deadline passes.
		f = newFixture(t, schedule(1))
		r := runnerFunc(func(ctx context.Context, in measurement.Input) (measurement.Output, error) {
			in.Start()
			f.runner.cancel()
			_ = f.clock.Advance(attemptDeadline)
			return measurement.Output{}, ctx.Err()
		})
		o := <-f.start(t, f.run, r)
		f.expect(t, o.rep, o.err, nil, []want{{"incomplete", "cancelled_by_host", ""}}, nil)
	})
	t.Run("a refused plan writes no result", func(t *testing.T) {
		// A malformed plan is a validation failure of the caller, reported as
		// an error, never as a Measurement Result.
		f := newFixture(t, schedule(1), connected(t))
		f.plan.Limits.Connections = 13 // above the registered maximum
		if _, err := f.engine.Run(f.run, f.plan); err == nil {
			t.Fatal("a plan beyond the registered limits ran")
		}
		if len(f.store.Records()) != 0 || f.runner.Calls() != 0 {
			t.Fatal("a refused plan produced a result or an attempt")
		}
	})
}
