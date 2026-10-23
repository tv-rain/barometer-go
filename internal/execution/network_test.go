package execution_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/networkcontext"
	"github.com/tv-rain/barometer-go/internal/platform"
)

// accessOf decodes the access object of a result's Network Context.
func accessOf(t *testing.T, r record) (access map[string]any, egress map[string]any) {
	t.Helper()
	var nc struct {
		Access map[string]any `json:"access"`
		Egress map[string]any `json:"egress"`
	}
	if err := json.Unmarshal(r.NetworkContext, &nc); err != nil {
		t.Fatal(err)
	}
	return nc.Access, nc.Egress
}

// An access-network generation change during an attempt ends it with
// network_changed: not_run without context before the test starts,
// incomplete with network_changed true after it, and no change to a
// completed attempt whose outcome came first (REQ-RESULT-0051,
// REQ-RESULT-0052). Every watch is stopped when the attempt ends.
// requirement-ref: REQ-CONTEXT-0025
// requirement-ref: REQ-RESULT-0051
// requirement-ref: REQ-RESULT-0052
func TestNetworkChangeDuringAttempt(t *testing.T) {
	moved := defaultAccess
	moved.Generation = 2
	cases := []struct {
		name  string
		begin bool // the runner starts the test before the change
		after bool // the change arrives after the outcome
		want  want
	}{
		{"before the test starts", false, false, want{"not_run", "network_changed", ""}},
		{"after the test starts", true, false, want{"incomplete", "network_changed", ""}},
		{"after the outcome", true, true, completedConnected},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t, schedule(1))
			r := runnerFunc(func(ctx context.Context, in measurement.Input) (measurement.Output, error) {
				if tc.begin {
					in.Start()
				}
				if tc.after {
					return connected(t).out, nil
				}
				f.ctx.net.Set(moved)
				<-ctx.Done()
				return measurement.Output{}, ctx.Err()
			})
			o := <-f.start(t, f.run, r)
			if tc.after {
				f.ctx.net.Set(moved)
			}
			f.expect(t, o.rep, o.err, nil, []want{tc.want}, nil)
			if rec := f.records(t)[0]; rec.Execution.State != "not_run" {
				access, egress := accessOf(t, rec)
				if access["network_changed"] != (tc.want.reason == "network_changed") || egress["status"] != "unavailable" || egress["network_path"] != "direct" {
					t.Fatalf("context %s", rec.NetworkContext)
				}
			}
			if f.ctx.net.Watching() != 0 {
				t.Fatal("watch outlived the attempt")
			}
		})
	}
	t.Run("between capture and watch", func(t *testing.T) {
		// A change before the watch is registered is reported at once.
		f := newFixture(t, schedule(1), connected(t))
		f.plan.Context = changingCapture{f.ctx, moved}
		rep, err := f.engine.Run(f.run, f.plan)
		f.expect(t, rep, err, nil, []want{{"not_run", "network_changed", ""}}, nil)
		if f.runner.Calls() != 0 || len(f.dialers.Issued()) != 0 {
			t.Fatal("attempt continued after the change")
		}
	})
}

// changingCapture captures, then moves the platform to another generation.
type changingCapture struct {
	*contextSource
	moved platform.AccessState
}

func (c changingCapture) Capture(ctx context.Context) (snap networkcontext.Snapshot, err error) {
	snap, err = c.contextSource.Capture(ctx)
	c.net.Set(c.moved)
	return snap, err
}

// D1 (DEV-644): an unsatisfied network condition before the first attempt
// records not_run with constraints_unsatisfied, because the run was reserved;
// before a retry it ends the run without a second attempt or result, as an
// unsatisfied precondition does.
// requirement-ref: REQ-EXECUTION-0031
func TestNetworkConditionsBeforeEachAttempt(t *testing.T) {
	validated := task.NetworkRequirements{VPNPolicy: "any", RequireValidated: true}
	f := newFixture(t, schedule(2, task.InternalError), failure(), connected(t))
	f.plan.Network = validated
	f.ctx.net.Set(ready)
	f.runner.steps[0].during = func() { f.ctx.net.Set(platform.AccessState{Generation: 1, Type: platform.NetworkWiFi}) }
	rep, err := f.engine.Run(f.run, f.plan)
	f.expect(t, rep, err, nil, []want{{"incomplete", "internal_error", ""}}, []time.Duration{35 * time.Second})
	if f.runner.Calls() != 1 || len(f.dialers.Issued()) != 1 {
		t.Fatal("a retry ran on an unsatisfied condition")
	}

	f = newFixture(t, schedule(2, task.InternalError), connected(t))
	f.plan.Network = validated
	rep, err = f.engine.Run(f.run, f.plan)
	f.expect(t, rep, err, nil, []want{{"not_run", "constraints_unsatisfied", ""}}, nil)

	// Without a condition, a failed capture is internal_error, not
	// network_context_unavailable (REQ-EXECUTION-0031).
	f = newFixture(t, schedule(1), connected(t))
	f.ctx.net.Fail()
	rep, err = f.engine.Run(f.run, f.plan)
	f.expect(t, rep, err, nil, []want{{"not_run", "internal_error", ""}}, nil)
}
