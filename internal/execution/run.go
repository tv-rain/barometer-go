package execution

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/registry"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/validation"
)

// MaxAttempts is the reviewed per-run attempt bound: the registered maximum of
// attempts per run, including the initial attempt (REQ-LIMIT-0011). A run uses
// one worker: attempts execute sequentially in the caller's goroutine and the
// engine starts no goroutines.
const MaxAttempts = 2

var (
	ErrEngine      = errors.New("execution: invalid engine")
	ErrPlan        = errors.New("execution: invalid run plan")
	ErrNoRunner    = errors.New("execution: no registered runner")
	ErrNotEligible = errors.New("execution: run not eligible")
	ErrCancelled   = errors.New("execution: run cancelled")
	ErrResult      = errors.New("execution: result construction failed")
	ErrOutput      = errors.New("execution: pending output failed")
)

// Clock is the engine's time source. Now must carry a monotonic reading for
// result timing. Wait blocks for d on this clock or until ctx is done, and then
// returns a non-nil error.
type Clock interface {
	Now() time.Time
	Wait(ctx context.Context, d time.Duration) error
}

// Jitter returns a locally generated delay in [0, max] (REQ-SCHEDULE-0029).
type Jitter interface {
	Jitter(max time.Duration) time.Duration
}

// Output is the narrow pending-output port this engine consumes. Put accepts
// one validated attempt result. An error means the result may or may not be
// durable; the engine stops the run and never retries a write. The production
// store is owned by the State, retention and clearing work; no in-memory
// store is a production fallback.
type Output interface {
	Put(ctx context.Context, v result.Value) error
}

// ContextSource supplies the attempt's Network Context for the run's path. It
// is an interim seam: acquisition, caching and path consistency belong to the
// Network paths and context work.
type ContextSource interface {
	Capture(ctx context.Context) (observation.NetworkContext, error)
}

// DialerSource supplies a fresh attempt-scoped dialer for each attempt number.
type DialerSource interface {
	Dialer(attempt int) (measurement.Dialer, error)
}

// SystemClock uses the process monotonic clock and an operating-system timer.
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now() }
func (SystemClock) Wait(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// LocalJitter draws uniformly from crypto/rand so jitter is not predictable
// from task data. A read failure yields no jitter, never a larger delay.
type LocalJitter struct{}

func (LocalJitter) Jitter(max time.Duration) time.Duration {
	if max <= 0 {
		return 0
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)+1))
	if err != nil {
		return 0
	}
	return time.Duration(n.Int64())
}

// Engine runs one run of an admitted task. It is immutable and safe for
// concurrent Run calls on independent plans.
type Engine struct {
	registry *registry.Registry
	clock    Clock
	jitter   Jitter
	output   Output
}

func NewEngine(r *registry.Registry, c Clock, j Jitter, o Output) (*Engine, error) {
	if r == nil || c == nil || j == nil || o == nil {
		return nil, ErrEngine
	}
	return &Engine{registry: r, clock: c, jitter: j, output: o}, nil
}

// Plan is one run of an admitted task. Every field is a trusted local input
// prepared after admission; none is task-controlled authority. Conditions is
// the host-directed eligibility snapshot; the engine sets its Now. Result must
// be bound to the same task ID, test and implementation.
type Plan struct {
	TaskID         string
	Test           task.TestVersion
	Parameters     task.Parameters
	Schedule       task.Schedule
	Conditions     validation.StartContext
	Implementation model.Implementation
	Result         *result.Decoder
	Context        ContextSource
	Dialers        DialerSource
}

// AttemptReport is the execution record of one terminal attempt. Usage is the
// runner's report for budget owners; it is not a result field.
type AttemptReport struct {
	Number  int
	Attempt Attempt
	Usage   measurement.Usage
}

// Report lists the run's terminal attempts in order. Attempts already written
// to the output port remain listed when Run returns an error.
type Report struct {
	RunID    string
	Attempts []AttemptReport
}

// Run executes attempts sequentially. Before any work it fails closed on an
// invalid plan, a missing runner or an ineligible run. Each terminal attempt is
// built through result construction and written to the output port before the
// next decision. A retry happens only when the attempt's reason is a retry
// candidate, the schedule permits it (validation.RetryEligible) both before and
// after the backoff wait, and the realized delay ends before not_after.
// requirement-ref: REQ-SCHEDULE-0024
// requirement-ref: REQ-SCHEDULE-0029
func (e *Engine) Run(ctx context.Context, p Plan) (Report, error) {
	if e == nil {
		return Report{}, ErrEngine
	}
	if p.TaskID == "" || p.Parameters == nil || p.Result == nil || p.Context == nil || p.Dialers == nil ||
		p.Schedule.Retry.MaxAttempts < 1 || p.Schedule.Retry.MaxAttempts > MaxAttempts {
		return Report{}, ErrPlan
	}
	runner, ok := e.registry.Runner(p.Test)
	if !ok {
		return Report{}, ErrNoRunner
	}
	if !validation.RunEligible(p.Schedule, e.conditions(p)) {
		return Report{}, ErrNotEligible
	}
	report := Report{RunID: result.NewRunID()}
	for n := 1; n <= p.Schedule.Retry.MaxAttempts; n++ {
		if ctx.Err() != nil {
			return report, ErrCancelled
		}
		a, usage, v, err := e.attempt(ctx, p, runner, report.RunID, n)
		if err != nil {
			return report, err
		}
		if e.output.Put(ctx, v) != nil {
			return report, ErrOutput
		}
		report.Attempts = append(report.Attempts, AttemptReport{Number: n, Attempt: a, Usage: usage})
		reason, retry := a.RetryReason()
		if !retry || !e.retry(ctx, p, n, reason) {
			break
		}
	}
	return report, nil
}

func (e *Engine) conditions(p Plan) validation.StartContext {
	c := p.Conditions
	c.Now = e.clock.Now()
	return c
}

// retry applies the schedule's retry object; it never invents an attempt.
func (e *Engine) retry(ctx context.Context, p Plan, completed int, reason task.ExecutionReason) bool {
	if !validation.RetryEligible(p.Schedule, e.conditions(p), completed, reason) || p.Schedule.Retry.Backoff == nil {
		return false
	}
	b := p.Schedule.Retry.Backoff
	// With at most two attempts, fixed and exponential backoff both produce
	// the initial delay; jitter is local and the realized delay is capped.
	delay := time.Duration(b.InitialDelayMS)*time.Millisecond + e.jitter.Jitter(time.Duration(b.JitterMS)*time.Millisecond)
	if limit := time.Duration(b.MaxDelayMS) * time.Millisecond; delay > limit {
		delay = limit
	}
	if p.Schedule.NotAfter != nil && !p.Schedule.NotAfter.AfterTime(e.clock.Now().Add(delay)) {
		return false
	}
	if e.clock.Wait(ctx, delay) != nil || ctx.Err() != nil {
		return false
	}
	return validation.RetryEligible(p.Schedule, e.conditions(p), completed, reason)
}

// attempt runs one attempt and returns its terminal state and validated result.
// The test is treated as started when the runner is invoked; under the current
// rule an outer condition before that is not_run and after it is incomplete.
// A not_run result never carries Network Context.
func (e *Engine) attempt(ctx context.Context, p Plan, runner measurement.Runner, runID string, n int) (Attempt, measurement.Usage, result.Value, error) {
	pending := NewAttempt()
	started := e.clock.Now()
	nc, err := p.Context.Capture(ctx)
	if err != nil {
		a, v, err := e.stopped(p, pending, task.NetworkContextUnavailable, nil, runID, n, started)
		return a, measurement.Usage{}, v, err
	}
	dialer, err := p.Dialers.Dialer(n)
	if err != nil || dialer == nil {
		a, v, err := e.stopped(p, pending, task.InternalError, nil, runID, n, started)
		return a, measurement.Usage{}, v, err
	}
	running, _ := pending.Start()
	out, err := runner.Run(ctx, measurement.Input{Parameters: p.Parameters, Dialer: dialer})
	if err != nil {
		// Host cancellation is the only outer condition this engine observes;
		// deadline and budget coordination belong to later runtime work.
		reason := task.InternalError
		if ctx.Err() != nil {
			reason = task.CancelledByHost
		}
		a, v, err := e.stopped(p, running, reason, &nc, runID, n, started)
		return a, out.Usage, v, err
	}
	completed, _ := running.Complete()
	r, err := e.envelope(p, completed, runID, n, started)
	if err != nil {
		return Attempt{}, measurement.Usage{}, result.Value{}, err
	}
	r.NetworkContext = &nc
	r.Measurement = &model.Measurement{Outcome: out.Outcome, Observations: out.Observations}
	if v, err := p.Result.Build(r); err == nil {
		return completed, out.Usage, v, nil
	}
	// Output the test contract rejects is not retained as a measurement.
	a, v, err := e.stopped(p, running, task.InternalError, &nc, runID, n, started)
	return a, out.Usage, v, err
}

// stopped records an outer condition. Network Context is retained only for an
// incomplete attempt, never for not_run.
func (e *Engine) stopped(p Plan, a Attempt, reason task.ExecutionReason, nc *observation.NetworkContext, runID string, n int, started time.Time) (Attempt, result.Value, error) {
	stopped, err := a.Stop(reason)
	if err != nil {
		return Attempt{}, result.Value{}, ErrResult
	}
	r, err := e.envelope(p, stopped, runID, n, started)
	if err != nil {
		return Attempt{}, result.Value{}, err
	}
	if stopped.Phase() == Incomplete {
		r.NetworkContext = nc
	}
	v, err := p.Result.Build(r)
	if err != nil {
		return Attempt{}, result.Value{}, ErrResult
	}
	return stopped, v, nil
}

func (e *Engine) envelope(p Plan, a Attempt, runID string, n int, started time.Time) (model.Result, error) {
	timing, err := result.Timing(started, e.clock.Now())
	if err != nil {
		return model.Result{}, ErrResult
	}
	state := map[Phase]string{Completed: "completed", Incomplete: "incomplete", NotRun: "not_run"}[a.Phase()]
	return model.Result{
		SchemaVersion: result.Contract, TaskID: p.TaskID, RunID: runID, Attempt: n,
		Test:           model.Test{ID: p.Test.ID, Version: p.Test.Version, ResultVersion: result.Contract},
		Implementation: p.Implementation, Time: timing,
		Execution: model.Execution{State: state, Reason: a.Reason()},
	}, nil
}
