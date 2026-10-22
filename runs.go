package barometer

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tv-rain/barometer-go/internal/admission"
	"github.com/tv-rain/barometer-go/internal/execution"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/persistence"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/validation"
)

// Reviewed implementation input (DEV-636): runs active at the same time across
// all tasks of one client. Each active run uses one goroutine, so this also
// bounds run workers. The registered limits set no ceiling.
const maxActiveRuns = 4

// RunState polling values. They are API-local and are not Measurement Result
// execution states or censorship classifications.
const (
	runActive   = "active"
	runTerminal = "terminal"
)

// pathSource is the interim seam for the direct path of one run, defined from
// this consumer. It supplies the run's Network Context source, attempt-scoped
// dialers and the result decoder bound to the task. Path selection, proxy
// handling and Network Context acquisition belong to Network paths and context
// (DEV-578); a task that needs any of them never reaches this seam.
type pathSource interface {
	direct(ctx context.Context, r pathRequest) (pathPlan, error)
}

type pathRequest struct {
	TaskID                     string
	Test                       task.TestVersion
	MaxAttempts, MaxDurationMS int
}

type pathPlan struct {
	Implementation model.Implementation
	Result         *result.Decoder
	Context        execution.ContextSource
	Dialers        execution.DialerSource
}

// runStatus is shared by the run table and the run's goroutine. A run is
// active until the engine has returned; finished is closed at that moment.
type runStatus struct {
	done     atomic.Bool
	finished chan struct{}
}

func newRunStatus() *runStatus { return &runStatus{finished: make(chan struct{})} }

func (s *runStatus) finish() {
	s.done.Store(true)
	close(s.finished)
}

// errRunCancelled is the cancellation cause CancelRun records. The engine maps
// it, like a lifecycle cancellation, to cancelled_by_host.
var errRunCancelled = errors.New("barometer: run cancelled")

// latestRun is the most recent reservation of one task in this generation.
type latestRun struct {
	occurrence int
	handle     string
}

// runSet holds the reservation state of one generation. mu serializes every
// reservation with every other one, so the due check, the ledger write and
// the run handle form one atomic step; it is held only for calls bounded by
// the generation context and the lifecycle deadline.
type runSet struct {
	mu        sync.Mutex
	active    int  // runs whose engine has not returned
	exclusive bool // an exclusive run is active
	results   int  // results the active runs may still write, at most
	latest    map[persistence.ReplayKey]latestRun
}

func newRunSet() *runSet { return &runSet{latest: make(map[persistence.ReplayKey]latestRun)} }

func (s *runSet) String() string   { return "barometer.runSet" }
func (s *runSet) GoString() string { return "barometer.runSet" }

// StartRun reserves one due run of the task the handle names and starts it on
// the orchestration engine, then returns its run handle without waiting for
// the run. A repeated start for the occurrence that already has a run returns
// that run's handle while its record is kept. A call that cannot start a run
// now returns run_not_due and reserves nothing; the core never starts a later
// run on its own.
func (c *Client) StartRun(taskHandle string) (string, error) { return c.startRun(taskHandle) }

// RunState returns "active" while the run's work remains and "terminal" after
// it has finished, whatever its attempts' outcomes.
func (c *Client) RunState(runHandle string) (string, error) { return c.runState(runHandle) }

// CancelRun prevents further attempts of the run, cancels its active work and
// waits for cleanup, for at most the registered cleanup bound. A run that is
// already terminal is left unchanged. If the run's work has not finished
// within the bound it returns cleanup_incomplete; the run then stays active,
// holds its reservation and starts no attempt, and a later call waits again.
func (c *Client) CancelRun(runHandle string) error { return c.cancelRun(runHandle) }

func (c *Client) startRun(taskHandle string) (handle string, err error) {
	if c == nil {
		return "", fail(notInitialized)
	}
	rt, err := c.reserve()
	if err != nil {
		return "", err
	}
	launched := false
	defer func() {
		if !launched {
			c.release()
		}
	}()
	defer func() {
		if recover() != nil {
			handle, err = "", fail(runNotDue)
		}
	}()
	admitted, err := rt.tasks.lookup(taskHandle)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(rt.ctx, rt.deadline)
	defer cancel()
	handle, plan, reserved, err := rt.reserveRun(ctx, taskHandle, admitted)
	if err != nil || reserved == nil {
		return handle, err
	}
	launched = true
	go rt.execute(plan, reserved, c.release)
	return handle, nil
}

// reserved is what a launched run releases when its engine returns. ctx is the
// run's context, a child of the generation context that CancelRun cancels.
type reserved struct {
	status    *runStatus
	ctx       context.Context
	cancel    context.CancelCauseFunc
	exclusive bool
	results   int
}

// reserveRun applies every start check in order and, only when all pass,
// records the run in the ledger before the run exists. It returns a nil
// reservation with the earlier handle for a repeated start.
// requirement-ref: REQ-SCHEDULE-0034
// requirement-ref: REQ-EXECUTION-0027
func (rt *runtime) reserveRun(ctx context.Context, taskHandle string, a admission.Admitted) (string, execution.Plan, *reserved, error) {
	notDue := fail(runNotDue)
	c, key, run := a.Candidate(), a.Key(), rt.run
	s := c.Schedule
	rt.runSet.mu.Lock()
	defer rt.runSet.mu.Unlock()

	now := run.clock.Now()
	rec, known, err := run.ledger.Lookup(ctx, key, now)
	if err != nil || known && !rec.Valid() {
		return "", execution.Plan{}, nil, notDue
	}
	if !known {
		rec = execution.NewAccounting(s, a.EvaluatedAt())
	}
	occurrence := execution.Occurrence(s, rec, now)
	if l, ok := rt.runSet.latest[key]; ok && l.occurrence == occurrence {
		if _, err := rt.runs.lookup(l.handle); err == nil {
			return l.handle, execution.Plan{}, nil, nil
		}
	}
	// A skipped occurrence is recorded before anything else, so its jitter is
	// drawn once and it never starts late.
	if advanced := execution.Advance(s, rec, now, run.jitter); !known || advanced != rec {
		if run.ledger.Put(ctx, key, advanced, now) != nil {
			return "", execution.Plan{}, nil, notDue
		}
		rec = advanced
	}
	_, supported := run.registry.Runner(a.Test())
	start := validation.StartContext{Now: now, DueAt: rec.NextDue, CompletedRuns: rec.Runs,
		Supported: supported, ConstraintsSatisfied: true, SafetySatisfied: true}
	if s.Window != nil && run.zone != nil {
		if loc, err := run.zone.Location(ctx); err == nil {
			start.LocalZone = loc
		}
	}
	gates := execution.Gates{Execution: c.Execution, ConsentRequired: a.ConsentRequired(), Consent: run.consent, Device: run.device}
	if !execution.Due(s, rec, now) || !validation.RunEligible(s, start) || !gates.Satisfied(ctx) {
		return "", execution.Plan{}, nil, notDue
	}
	exclusive := c.Execution.ConcurrencyMode == "exclusive"
	if rt.runSet.exclusive || exclusive && rt.runSet.active > 0 {
		return "", execution.Plan{}, nil, notDue
	}
	if rt.runSet.active >= maxActiveRuns {
		return "", execution.Plan{}, nil, fail(capacityExceeded)
	}
	attempts := s.Retry.MaxAttempts
	if free, err := rt.bundles.Capacity(ctx); err != nil || free < rt.runSet.results+attempts {
		return "", execution.Plan{}, nil, notDue
	}
	path, err := run.paths.direct(ctx, pathRequest{TaskID: c.ID, Test: a.Test(), MaxAttempts: attempts, MaxDurationMS: a.AttemptLimits().DurationMS})
	if err != nil || path.Result == nil || path.Context == nil || path.Dialers == nil {
		return "", execution.Plan{}, nil, notDue
	}

	runCtx, cancelRun := context.WithCancelCause(rt.ctx)
	r := &reserved{status: newRunStatus(), ctx: runCtx, cancel: cancelRun, exclusive: exclusive, results: attempts}
	handle, err := rt.addRun(runRecord{taskHandle: taskHandle, status: r.status, cancel: cancelRun})
	if err != nil {
		cancelRun(nil)
		return "", execution.Plan{}, nil, err
	}
	// Until the run is launched, every exit, including a panic, withdraws
	// the handle so no record stays active without a run.
	launched := false
	defer func() {
		if !launched {
			rt.runs.remove(handle)
			cancelRun(nil)
		}
	}()
	// The run is counted before it can do anything; an uncertain write
	// loses this occurrence rather than risking an extra run.
	if run.ledger.Put(ctx, key, execution.Reserve(s, rec, run.jitter), now) != nil {
		return "", execution.Plan{}, nil, notDue
	}
	if rt.ctx.Err() != nil {
		return "", execution.Plan{}, nil, fail(transitionInProgress)
	}
	launched = true
	rt.runSet.latest[key] = latestRun{occurrence: occurrence, handle: handle}
	rt.runSet.active++
	rt.runSet.results += attempts
	rt.runSet.exclusive = rt.runSet.exclusive || exclusive
	plan := execution.Plan{TaskID: c.ID, Test: a.Test(), Parameters: c.Parameters, Schedule: s, Conditions: start,
		Deadline:       time.Duration(a.AttemptLimits().DurationMS) * time.Millisecond,
		Limits:         a.AttemptLimits(),
		Implementation: path.Implementation, Result: path.Result, Context: path.Context, Dialers: path.Dialers, Preconditions: gates}
	return handle, plan, r, nil
}

// addRun publishes a run handle. A full table first drops terminal runs; it
// never evicts an active run. An invalidated table means a transition began.
func (rt *runtime) addRun(r runRecord) (string, error) {
	h, err := rt.runs.add(r)
	if err == (opError{code: capacityExceeded}) {
		rt.runs.removeIf(func(v runRecord) bool { return v.terminal() })
		h, err = rt.runs.add(r)
	}
	if err == (opError{code: handleUnavailable}) {
		return "", fail(transitionInProgress)
	}
	return h, err
}

// execute runs on its own goroutine and holds the generation reservation, so
// lifecycle transitions wait for it. It uses the run context, which CancelRun
// and every lifecycle transition cancel; the engine records the cancelled
// attempt with cancelled_by_host and starts no further attempt. A panic below
// the engine ends the run without crashing the host.
func (rt *runtime) execute(plan execution.Plan, r *reserved, release func()) {
	defer release()
	defer func() {
		rt.runSet.mu.Lock()
		rt.runSet.active--
		rt.runSet.results -= r.results
		if r.exclusive {
			rt.runSet.exclusive = false
		}
		rt.runSet.mu.Unlock()
		r.cancel(nil)
		r.status.finish()
	}()
	defer func() { _ = recover() }()
	_, _ = rt.engine.Run(r.ctx, plan)
}

// cancelRun reserves the generation, so it serializes with lifecycle
// transitions, which cancel every run themselves. A panic below the facade
// returns cleanup_incomplete rather than reporting success.
// requirement-ref: REQ-LIMIT-0008
func (c *Client) cancelRun(runHandle string) (err error) {
	if c == nil {
		return fail(notInitialized)
	}
	rt, err := c.reserve()
	if err != nil {
		return err
	}
	defer c.release()
	defer func() {
		if recover() != nil {
			err = fail(cleanupIncomplete)
		}
	}()
	r, err := rt.runs.lookup(runHandle)
	if err != nil {
		return err
	}
	if r.terminal() {
		return nil
	}
	r.cancel(errRunCancelled)
	if !rt.awaitRun(r.status) {
		return fail(cleanupIncomplete)
	}
	return nil
}

// awaitRun waits until the run's engine has returned, for at most the
// registered cleanup bound on the engine clock.
func (rt *runtime) awaitRun(s *runStatus) bool {
	select {
	case <-s.finished:
		return true
	default:
	}
	expired := make(chan struct{})
	stop := rt.run.clock.AfterFunc(execution.CleanupBound, func() { close(expired) })
	defer stop()
	select {
	case <-s.finished:
		return true
	case <-expired:
		return false
	}
}

func (c *Client) runState(runHandle string) (state string, err error) {
	if c == nil {
		return "", fail(notInitialized)
	}
	rt, err := c.reserve()
	if err != nil {
		return "", err
	}
	defer c.release()
	r, err := rt.runs.lookup(runHandle)
	if err != nil {
		return "", err
	}
	if r.terminal() {
		return runTerminal, nil
	}
	return runActive, nil
}
