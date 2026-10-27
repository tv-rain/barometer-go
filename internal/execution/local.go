package execution

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/networkcontext"
	"github.com/tv-rain/barometer-go/internal/result"
)

// localTest reports whether the test is a local network-state observation
// whose procedure is the access read itself (NET-01).
func localTest(p Plan) bool { return p.Test.ID == "NET-01" }

// localAttempt runs one NET-01 attempt. The runner receives no dialer and no
// resolver, and the attempt acquires no egress context, so it opens no network
// path (REQ-NET-01-0021, REQ-NET-01-0024). Stated access conditions are
// evaluated on a gate read before the test starts. The test's own read,
// through the shared collector, starts it and arms the generation watch: a
// change before the outcome is recorded is incomplete with network_changed,
// and the watch is released after the outcome (GO-SPEC-ISSUE-004). No result
// carries the top-level Network Context (REQ-NET-01-0023).
// requirement-ref: REQ-NET-01-0007
// requirement-ref: REQ-NET-01-0023
// requirement-ref: REQ-NET-01-0024
// requirement-ref: REQ-NET-01-0027
func (e *Engine) localAttempt(actx context.Context, cancel context.CancelCauseFunc, p Plan, runner measurement.Runner, pending Attempt, runID string, n int, started time.Time) (Attempt, measurement.Usage, result.Value, error) {
	stop := func(a Attempt, reason task.ExecutionReason) (Attempt, measurement.Usage, result.Value, error) {
		at, v, err := e.stopped(p, a, reason, nil, runID, n, started)
		return at, measurement.Usage{}, v, err
	}
	if AccessRequired(p.Network) {
		snap, err := p.Context.Capture(actx)
		if actx.Err() != nil {
			return stop(pending, stopReason(actx))
		}
		if err != nil || !NetworkConditions(p.Network, snap.Access()) {
			if n > 1 {
				return Attempt{}, measurement.Usage{}, result.Value{}, errSkipped
			}
			return stop(pending, task.ConstraintsUnsatisfied)
		}
	}
	// NET-01 never acquires egress, so a stated egress requirement cannot be met.
	if reason := ContextConditions(p.Network.Context, observation.Egress{}); reason != "" {
		if reason == task.ConstraintsUnsatisfied && n > 1 {
			return Attempt{}, measurement.Usage{}, result.Value{}, errSkipped
		}
		return stop(pending, reason)
	}
	if actx.Err() != nil {
		return stop(pending, stopReason(actx))
	}
	// The scope records the start and the outcome. It has no path dialer and
	// is never handed to the runner, so no connection can be opened.
	s := newScope(actx, nil, p.Limits, cancel)
	var mu sync.Mutex
	var stopWatch func()
	read := func(ctx context.Context) (observation.AccessSnapshot, error) {
		s.Start()
		snap, err := p.Context.Capture(ctx)
		if err != nil {
			if errors.Is(err, networkcontext.ErrNoNetworkState) {
				return observation.AccessSnapshot{}, measurement.ErrNoNetworkState
			}
			return observation.AccessSnapshot{}, measurement.ErrNetworkState
		}
		w := snap.Watch(func() { cancel(errNetworkChanged) })
		mu.Lock()
		if stopWatch != nil {
			mu.Unlock()
			w()
			return observation.AccessSnapshot{}, measurement.ErrRunner
		}
		stopWatch = w
		mu.Unlock()
		c, err := snap.Context(false)
		if err != nil {
			return observation.AccessSnapshot{}, measurement.ErrNetworkState
		}
		return c.AccessSnapshot(), nil
	}
	out, err := runner.Run(actx, measurement.Input{Parameters: p.Parameters, Clock: e.clock, RecordOutcome: s.RecordOutcome, Start: s.Start, Network: read})
	// Step 4: release the watch created for the run, after the outcome.
	mu.Lock()
	if stopWatch != nil {
		stopWatch()
	}
	stopWatch = func() {}
	mu.Unlock()
	s.close()
	running, _ := pending.Start()
	at := pending
	if s.Started() {
		at = running
	}
	if err != nil || !s.terminal.Load() || out.Composite != nil || len(out.Observations) != 0 {
		reason := task.InternalError
		if actx.Err() != nil {
			reason = stopReason(actx)
		}
		return stop(at, reason)
	}
	completed, _ := running.Complete()
	r, err := e.envelope(p, completed, runID, n, started)
	if err != nil {
		return Attempt{}, measurement.Usage{}, result.Value{}, err
	}
	r.Measurement = &model.Measurement{Outcome: out.Outcome, Observations: []observation.Observation{}, Access: out.Access}
	if v, err := p.Result.Build(r); err == nil {
		return completed, measurement.Usage{}, v, nil
	}
	return stop(running, task.InternalError)
}
