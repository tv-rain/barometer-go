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
	"github.com/tv-rain/barometer-go/internal/networkcontext"
	"github.com/tv-rain/barometer-go/internal/registry"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/validation"
)

// MaxAttempts is the reviewed per-run attempt bound: the registered maximum of
// attempts per run, including the initial attempt (REQ-LIMIT-0011). A run uses
// one worker: attempts execute sequentially in the caller's goroutine. The
// engine starts no long-lived goroutine; the attempt deadline timer and the
// attempt scope's cancellation hook each run one short callback when they
// fire.
const MaxAttempts = 2

var (
	ErrEngine      = errors.New("execution: invalid engine")
	ErrPlan        = errors.New("execution: invalid run plan")
	ErrNoRunner    = errors.New("execution: no registered runner")
	ErrNotEligible = errors.New("execution: run not eligible")
	ErrResult      = errors.New("execution: result construction failed")
	ErrOutput      = errors.New("execution: pending output failed")
)

// errDeadline is the cancellation cause of an expired attempt deadline. Any
// other cause, including a host or lifecycle cancellation of the run context,
// is recorded as cancelled_by_host.
var errDeadline = errors.New("execution: attempt deadline exceeded")

// errNetworkChanged is the cancellation cause of an access-network generation
// change during an attempt.
var errNetworkChanged = errors.New("execution: access network changed")

// errSkipped reports a retry that did not start because its network
// conditions were unsatisfied; it produces no result.
var errSkipped = errors.New("execution: retry conditions unsatisfied")

// Clock is the engine's time source. Now must carry a monotonic reading for
// result timing. Wait blocks for d on this clock or until ctx is done, and then
// returns a non-nil error. AfterFunc runs f once d has elapsed on this clock,
// unless stop prevented it first.
type Clock interface {
	Now() time.Time
	Wait(ctx context.Context, d time.Duration) error
	AfterFunc(d time.Duration, f func()) (stop func() bool)
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
	// Capacity reports how many more results the store can accept now. The
	// caller checks it before reserving a run, so execution stops before
	// pending-output capacity is exhausted; an error means no capacity.
	Capacity(ctx context.Context) (int, error)
}

// ContextSource captures the attempt's access snapshot for the run's path:
// the access values that network conditions are evaluated against, the
// Network Context attached to the result and the watch for an
// access-network generation change. networkcontext.Source is the production
// source of either path; the engine acquires the snapshot's egress context
// through the path's dialer with Snapshot.Acquire.
type ContextSource interface {
	Capture(ctx context.Context) (networkcontext.Snapshot, error)
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
func (SystemClock) AfterFunc(d time.Duration, f func()) func() bool {
	return time.AfterFunc(d, f).Stop
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
// be bound to the same task ID, test and implementation. Deadline is the
// effective max_duration_ms of each attempt. Limits are the effective attempt
// limits after admission; the engine meters their byte and connection fields
// and ignores DurationMS, which Deadline carries. Network holds the task's
// network conditions, evaluated against each attempt's access snapshot.
type Plan struct {
	TaskID         string
	Test           task.TestVersion
	Parameters     task.Parameters
	Schedule       task.Schedule
	Network        task.NetworkRequirements
	Deadline       time.Duration
	Limits         task.AttemptLimits
	Conditions     validation.StartContext
	Implementation model.Implementation
	Result         *result.Decoder
	Context        ContextSource
	Dialers        DialerSource
	Targets        TargetCheck   // nil grants no network destination
	Path           PathCheck     // the proxy path's pre-attempt check; nil on the direct path
	Preconditions  Preconditions // optional attempt-start check
}

// PathCheck is the pre-attempt check of a proxy path (REQ-EXECUTION-0008): a
// valid host proxy is configured, local policy permits the task to use it and
// the test with its concrete input and options is compatible with HTTP
// CONNECT. Check returns the empty reason when the attempt may use the path,
// otherwise proxy_not_configured, proxy_not_permitted or
// proxy_path_unsupported, in that order. It makes no contact.
type PathCheck interface {
	Check() task.ExecutionReason
}

// AttemptReport is the execution record of one terminal attempt. Usage is the
// attempt scope's metered count, not the runner's report; it is not a result
// field.
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
// invalid plan or a missing runner. A run whose context has already ended
// records its first attempt as not_run with the cancellation reason and does
// nothing else. Otherwise it fails closed on an ineligible run or unsatisfied
// preconditions, which are evaluated again before a retry. Network conditions
// are evaluated in each attempt against its own access snapshot: unsatisfied
// in the first attempt, it is not_run with constraints_unsatisfied; unsatisfied
// in a retry, the run ends with no further attempt or result. Each terminal
// attempt is built through result construction and written to the output port
// before the next decision. A retry happens only when the attempt's reason is
// a retry candidate, the schedule permits it (validation.RetryEligible) both
// before and after the backoff wait, the realized delay ends before not_after
// and the run has not been cancelled: cancellation prevents every further
// attempt.
//
// Results are written with a context that the run's cancellation does not
// end, so an attempt that committed its terminal state, including the
// cancelled one, is not lost to a racing cancellation. The output port bounds
// its own writes.
// requirement-ref: REQ-SCHEDULE-0024
// requirement-ref: REQ-SCHEDULE-0029
func (e *Engine) Run(ctx context.Context, p Plan) (Report, error) {
	if e == nil {
		return Report{}, ErrEngine
	}
	if p.TaskID == "" || p.Parameters == nil || p.Result == nil || p.Context == nil || p.Dialers == nil ||
		p.Schedule.Retry.MaxAttempts < 1 || p.Schedule.Retry.MaxAttempts > MaxAttempts ||
		p.Deadline <= 0 || p.Deadline > MaxAttemptDuration || !metered(p.Limits) {
		return Report{}, ErrPlan
	}
	runner, ok := e.registry.Runner(p.Test)
	if !ok {
		return Report{}, ErrNoRunner
	}
	report := Report{RunID: result.NewRunID()}
	write := context.WithoutCancel(ctx)
	if ctx.Err() != nil {
		a, v, err := e.stopped(p, NewAttempt(), stopReason(ctx), nil, report.RunID, 1, e.clock.Now())
		if err != nil {
			return report, err
		}
		if e.output.Put(write, v) != nil {
			return report, ErrOutput
		}
		report.Attempts = append(report.Attempts, AttemptReport{Number: 1, Attempt: a})
		return report, nil
	}
	if !validation.RunEligible(p.Schedule, e.conditions(p)) || !p.satisfied(ctx) {
		return Report{}, ErrNotEligible
	}
	for n := 1; n <= p.Schedule.Retry.MaxAttempts; n++ {
		if n > 1 && (ctx.Err() != nil || !p.satisfied(ctx)) {
			break
		}
		a, usage, v, err := e.attempt(ctx, p, runner, report.RunID, n)
		if errors.Is(err, errSkipped) {
			break
		}
		if err != nil {
			return report, err
		}
		if e.output.Put(write, v) != nil {
			return report, ErrOutput
		}
		report.Attempts = append(report.Attempts, AttemptReport{Number: n, Attempt: a, Usage: usage})
		reason, retry := a.RetryReason()
		if !retry || ctx.Err() != nil || !e.retry(ctx, p, n, reason) {
			break
		}
	}
	return report, nil
}

// stopReason maps the cause of an ended context to its execution reason. The
// first cancellation is the cause, so a deadline that expired before a host
// cancellation stays attempt_deadline_exceeded, and the reverse.
func stopReason(ctx context.Context) task.ExecutionReason {
	switch cause := context.Cause(ctx); {
	case errors.Is(cause, errDeadline):
		return task.AttemptDeadlineExceeded
	case errors.Is(cause, errBudget):
		return task.ResourceLimit
	case errors.Is(cause, errNetworkChanged):
		return task.NetworkChanged
	}
	if reason, ok := pathFailure(context.Cause(ctx)); ok {
		return reason
	}
	return task.CancelledByHost
}

// cutShort reports a termination that stopped the attempt's active operation
// before the runner produced its outcome: an exhausted budget
// (REQ-LIMIT-0027) or a path failure, such as a CONNECT the proxy refused,
// which is an execution observation and never the test's (REQ-PROXY-0029).
// The outcome the runner built from the stopped operation is not retained.
func cutShort(ctx context.Context) (task.ExecutionReason, bool) {
	cause := context.Cause(ctx)
	if errors.Is(cause, errBudget) {
		return task.ResourceLimit, true
	}
	return pathFailure(cause)
}

// satisfied re-evaluates the plan's preconditions at attempt start.
func (p Plan) satisfied(ctx context.Context) bool {
	return p.Preconditions == nil || p.Preconditions.Satisfied(ctx)
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

// attempt runs one attempt under its own deadline and returns its terminal
// state and validated result. The deadline is the plan's max_duration_ms on
// the engine clock and covers context capture, the runner and cleanup
// (REQ-LIMIT-0006). The runner dials through an attempt scope that closes
// every connection when the attempt ends, by deadline or cancellation, and
// after the runner returns.
//
// The state follows the first termination. Until the test starts
// (measurement.Input) an outer condition yields not_run, and after it
// incomplete (REQ-RESULT-0051). A returned outcome yields completed, even when
// a cancellation or the deadline arrived while connections closed
// (REQ-RESULT-0052). A not_run result never carries Network Context
// (REQ-RESULT-0053).
// requirement-ref: REQ-RESULT-0051
// requirement-ref: REQ-RESULT-0052
// requirement-ref: REQ-LIMIT-0006
func (e *Engine) attempt(ctx context.Context, p Plan, runner measurement.Runner, runID string, n int) (Attempt, measurement.Usage, result.Value, error) {
	pending := NewAttempt()
	started := e.clock.Now()
	actx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	stopTimer := e.clock.AfterFunc(p.Deadline, func() { cancel(errDeadline) })
	defer stopTimer()

	// A proxy attempt that cannot use the path records why and makes no
	// contact, not even the access-context capture; it never falls back to
	// the direct path (REQ-EXECUTION-0008, REQ-EXECUTION-0010).
	if p.Path != nil {
		if reason := p.Path.Check(); reason != "" {
			a, v, err := e.stopped(p, pending, reason, nil, runID, n, started)
			return a, measurement.Usage{}, v, err
		}
	}
	snap, err := p.Context.Capture(actx)
	if actx.Err() != nil {
		a, v, err := e.stopped(p, pending, stopReason(actx), nil, runID, n, started)
		return a, measurement.Usage{}, v, err
	}
	// An access condition that cannot be determined or is not met is
	// constraints_unsatisfied; a retry it prevents is not attempted
	// (REQ-EXECUTION-0031). Without such a condition, a failed capture
	// leaves the attempt without the context its result requires.
	if err == nil && !NetworkConditions(p.Network, snap.Access()) || err != nil && AccessRequired(p.Network) {
		if n > 1 {
			return Attempt{}, measurement.Usage{}, result.Value{}, errSkipped
		}
		a, v, err := e.stopped(p, pending, task.ConstraintsUnsatisfied, nil, runID, n, started)
		return a, measurement.Usage{}, v, err
	}
	if err != nil {
		a, v, err := e.stopped(p, pending, task.InternalError, nil, runID, n, started)
		return a, measurement.Usage{}, v, err
	}
	// A generation change from the capture on ends the attempt
	// (REQ-CONTEXT-0025); its context then records network_changed.
	stopWatch := snap.Watch(func() { cancel(errNetworkChanged) })
	defer stopWatch()
	// Egress context is acquired through a dialer of the attempt's own path
	// inside the attempt's deadline, cancellation and network-change watch
	// (REQ-CONTEXT-0025). The lookup dialer is not the metered attempt
	// scope, so its bytes do not count toward the attempt budgets and it
	// does not start the test.
	if actx.Err() == nil && snap.Acquires() {
		if d, derr := p.Dialers.Dialer(n); derr == nil && d != nil {
			var policy AddressPolicy = PublicAddresses{}
			if targets, ok := p.Targets.(*Targets); ok {
				policy = targets.policy
			}
			snap, err = snap.Acquire(actx, contextFetcher(d, policy))
		}
	}
	if actx.Err() != nil {
		a, v, err := e.stopped(p, pending, stopReason(actx), nil, runID, n, started)
		return a, measurement.Usage{}, v, err
	}
	var stable, changed observation.NetworkContext
	if err == nil {
		stable, err = snap.Context(false)
	}
	if err == nil {
		changed, err = snap.Context(true)
	}
	if err != nil {
		a, v, err := e.stopped(p, pending, task.InternalError, nil, runID, n, started)
		return a, measurement.Usage{}, v, err
	}
	// A required egress field still missing after the bounded lookup is
	// network_context_unavailable; an allowlist mismatch is
	// constraints_unsatisfied, which, like an access condition, prevents a
	// retry rather than recording it (REQ-CONTEXT-0026, REQ-CONTEXT-0028).
	if reason := ContextConditions(p.Network.Context, snap.Egress()); reason != "" {
		if reason == task.ConstraintsUnsatisfied && n > 1 {
			return Attempt{}, measurement.Usage{}, result.Value{}, errSkipped
		}
		a, v, err := e.stopped(p, pending, reason, nil, runID, n, started)
		return a, measurement.Usage{}, v, err
	}
	contextFor := func(reason task.ExecutionReason) *observation.NetworkContext {
		if reason == task.NetworkChanged {
			return &changed
		}
		return &stable
	}
	dialer, err := p.Dialers.Dialer(n)
	if err != nil || dialer == nil {
		a, v, err := e.stopped(p, pending, task.InternalError, nil, runID, n, started)
		return a, measurement.Usage{}, v, err
	}
	s := newScope(actx, dialer, p.Limits, cancel)
	s.targets = p.Targets
	checkNetwork := func(family string) error {
		s.Start()
		if actx.Err() != nil {
			return measurement.ErrRunner
		}
		a := snap.Access()
		if family == "" || family == "ipv4" && a.IPv4Available == "yes" || family == "ipv6" && a.IPv6Available == "yes" {
			return nil
		}
		cancel(pathError{task.ConstraintsUnsatisfied})
		return measurement.ErrRunner
	}
	out, err := runner.Run(actx, measurement.Input{Parameters: p.Parameters, Clock: e.clock, CheckNetwork: checkNetwork, RecordOutcome: s.RecordOutcome, Dialer: s, Start: s.Start, Redirect: s.Redirect, Resolve: s.resolve})
	s.close()
	usage := s.Usage()
	running, _ := pending.Start()
	if err != nil {
		reason := task.InternalError
		if actx.Err() != nil {
			reason = stopReason(actx)
		}
		// A CONNECT exchange the attempt deadline cut short did not complete
		// within the deadline (REQ-PROXY-0028). A host cancellation or a
		// network change during the exchange keeps its own reason.
		if reason == task.AttemptDeadlineExceeded && s.interrupted() == task.ProxyConnectFailed {
			reason = task.ProxyConnectFailed
		}
		at := pending
		if s.Started() {
			at = running
		}
		a, v, err := e.stopped(p, at, reason, contextFor(reason), runID, n, started)
		return a, usage, v, err
	}
	// An exhausted budget or a path failure stopped the active operation
	// before the runner produced this outcome. State follows the actual test start.
	if reason, ok := cutShort(actx); ok && !s.terminal.Load() {
		at := pending
		if s.Started() {
			at = running
		}
		a, v, err := e.stopped(p, at, reason, &stable, runID, n, started)
		return a, usage, v, err
	}
	completed, _ := running.Complete()
	r, err := e.envelope(p, completed, runID, n, started)
	if err != nil {
		return Attempt{}, measurement.Usage{}, result.Value{}, err
	}
	r.NetworkContext = &stable
	r.Measurement = &model.Measurement{Outcome: out.Outcome, Observations: out.Observations}
	if v, err := p.Result.Build(r); err == nil {
		return completed, usage, v, nil
	}
	// Output the test contract rejects is not retained as a measurement.
	a, v, err := e.stopped(p, running, task.InternalError, &stable, runID, n, started)
	return a, usage, v, err
}

// metered reports whether the byte and connection limits are within their
// registered ranges. The engine rejects a plan beyond them rather than clamp
// it. Zero is valid and permits no traffic (REQ-LIMIT-0007).
// requirement-ref: REQ-LIMIT-0006
func metered(l task.AttemptLimits) bool {
	return l.BytesSent >= 0 && l.BytesSent <= MaxBytesSent &&
		l.BytesReceived >= 0 && l.BytesReceived <= MaxBytesReceived &&
		l.Connections >= 0 && l.Connections <= MaxConnections
}

// stopped records an outer condition. Network Context is retained only for an
// incomplete attempt, never for not_run.
// requirement-ref: REQ-RESULT-0053
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
