package barometer

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/execution"
	"github.com/tv-rain/barometer-go/internal/model/task"
)

// waitTimers waits until the run clock has n armed timers. The bound is a
// liveness guard, not a timing assertion.
func (h *harness) waitTimers(t *testing.T, n int) {
	t.Helper()
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); time.Sleep(time.Millisecond) {
		if h.run.clock.Timers() == n {
			return
		}
	}
	t.Fatalf("armed timers = %d, want %d", h.run.clock.Timers(), n)
}

func (h *harness) startHeld(t *testing.T, c *Client, id, schedule string) string {
	t.Helper()
	h.run.runner.holdCalls()
	rh, err := c.StartRun(h.admit(t, c, runTask(id, schedule, "", false)))
	if err != nil {
		t.Fatal(err)
	}
	waitStarted(t, h.run.runner)
	return rh
}

func outputBytes(h *harness) [][]byte { return h.run.output.Records() }

// CancelRun stops the active attempt, which is recorded as incomplete
// cancelled_by_host, prevents the permitted retry and returns once the run is
// terminal. Repeating it on the retained terminal run changes nothing.
// requirement-ref: REQ-RESULT-0020
// requirement-ref: REQ-RESULT-0051
func TestCancelRunStopsTheRun(t *testing.T) {
	h := runningHarness(t)
	c := h.initialized(t)
	retry := `"retry":{"max_attempts":2,"retry_on":["internal_error"],"backoff":{"mode":"fixed","initial_delay_ms":30000,"max_delay_ms":30000}}`
	rh := h.startHeld(t, c, "cancel-1", retry)
	if s, _ := c.RunState(rh); s != runActive {
		t.Fatal("held run is not active")
	}
	if err := c.CancelRun(rh); err != nil {
		t.Fatal("cancel:", err)
	}
	if s, err := c.RunState(rh); err != nil || s != runTerminal {
		t.Fatal("cancelled run is not terminal on return:", s, err)
	}
	got := records(t, h.run.output)
	if len(got) != 1 || got[0].Execution.State != "incomplete" || got[0].Execution.Reason != "cancelled_by_host" || h.run.runner.Calls() != 1 {
		t.Fatalf("cancelled run = %+v, runner calls %d", got, h.run.runner.Calls())
	}
	before := outputBytes(h)
	for range 3 {
		if err := c.CancelRun(rh); err != nil {
			t.Fatal("repeated cancel:", err)
		}
	}
	if after := outputBytes(h); len(after) != len(before) || !bytes.Equal(after[0], before[0]) {
		t.Fatal("repeated cancel changed results")
	}
	if rec, ok := h.run.ledger.Record(key("cancel-1")); !ok || rec.Runs != 1 {
		t.Fatal("cancellation changed run accounting")
	}

	// A completed run is unchanged by a later cancellation.
	h.run.runner.releaseCalls()
	rh2, err := c.StartRun(h.admit(t, c, runTask("cancel-2", "", "", false)))
	if err != nil {
		t.Fatal(err)
	}
	waitTerminal(t, c, rh2)
	if err := c.CancelRun(rh2); err != nil {
		t.Fatal("cancel of a completed run:", err)
	}
	got = records(t, h.run.output)
	if len(got) != 2 || got[1].Execution.State != "completed" {
		t.Fatalf("completed run changed: %+v", got)
	}
}

// Unknown, task, stale and earlier-generation handles return
// handle_unavailable; calls outside the initialized state return the
// lifecycle codes.
func TestCancelRunHandles(t *testing.T) {
	h := runningHarness(t)
	h.adapters.clearStorage = func(context.Context) error { return nil }
	c := h.initialized(t)
	th := h.admit(t, c, runTask("cancel-h", "", "", false))
	rh, err := c.StartRun(th)
	if err != nil {
		t.Fatal(err)
	}
	waitTerminal(t, c, rh)
	for _, bad := range []string{"", "R.unknown", th, rh + "x"} {
		if codeOf(c.CancelRun(bad)) != "handle_unavailable" {
			t.Fatalf("CancelRun(%q) = %v", bad, c.CancelRun(bad))
		}
	}
	if err := c.ClearMeasurementData(); err != nil {
		t.Fatal(err)
	}
	if codeOf(c.CancelRun(rh)) != "not_initialized" {
		t.Fatal("cancel on an uninitialized client")
	}
	if err := c.Initialize("{}"); err != nil {
		t.Fatal(err)
	}
	if codeOf(c.CancelRun(rh)) != "handle_unavailable" {
		t.Fatal("earlier-generation handle resolved")
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if codeOf(c.CancelRun(rh)) != "client_closed" {
		t.Fatal("cancel on a closed client")
	}
	var nilClient *Client
	if codeOf(nilClient.CancelRun(rh)) != "not_initialized" || codeOf(NewClient().CancelRun(rh)) != "not_initialized" {
		t.Fatal("nil or zero client")
	}
}

// Cleanup that does not finish within the registered bound returns
// cleanup_incomplete: the run stays active, keeps its reservation and starts
// no attempt, and a later call waits again. Close reports the same scope as
// incomplete and stays retryable.
// requirement-ref: REQ-LIMIT-0008
func TestCancelRunCleanupIncomplete(t *testing.T) {
	h := runningHarness(t)
	h.adapters.deadline = 50 * time.Millisecond
	c := h.initialized(t)
	stuck := make(chan struct{})
	h.run.runner.stuck = stuck
	rh, err := c.StartRun(h.admit(t, c, runTask("stuck-1", "", "", false)))
	if err != nil {
		t.Fatal(err)
	}
	waitStarted(t, h.run.runner)
	h.waitTimers(t, 1) // the attempt deadline

	result := make(chan error, 1)
	go func() { result <- c.CancelRun(rh) }()
	h.waitTimers(t, 2) // plus the cleanup bound
	_ = h.run.clock.Advance(execution.CleanupBound - time.Millisecond)
	select {
	case err := <-result:
		t.Fatal("cancel returned before the cleanup bound:", err)
	default:
	}
	_ = h.run.clock.Advance(time.Millisecond)
	if err := <-result; codeOf(err) != "cleanup_incomplete" {
		t.Fatal("cancel past the cleanup bound:", err)
	}
	if s, _ := c.RunState(rh); s != runActive {
		t.Fatal("run with incomplete cleanup is not active")
	}
	if _, err := c.StartRun(h.admit(t, c, runTask("stuck-2", "", `"concurrency":{"mode":"exclusive"}`, false))); codeOf(err) != "run_not_due" {
		t.Fatal("exclusive run started beside a run whose cleanup is incomplete:", err)
	}
	if codeOf(c.Close()) != "cleanup_incomplete" {
		t.Fatal("close with incomplete cleanup")
	}

	close(stuck)
	if err := c.Close(); err != nil {
		t.Fatal("retried close:", err)
	}
	got := records(t, h.run.output)
	if len(got) != 1 || got[0].Execution.State != "incomplete" || got[0].Execution.Reason != "cancelled_by_host" || h.run.runner.Calls() != 1 {
		t.Fatalf("records = %+v", got)
	}
}

// The attempt deadline comes from the task's max_duration_ms on the engine
// clock: one millisecond before it the run is active, at it the attempt ends
// with attempt_deadline_exceeded.
// requirement-ref: REQ-LIMIT-0006
func TestStartRunEnforcesTheAttemptDeadline(t *testing.T) {
	h := runningHarness(t)
	c := h.initialized(t)
	h.run.runner.holdCalls()
	rh, err := c.StartRun(h.admit(t, c, strings.Replace(runTask("deadline-1", "", "", false), `"limits":{}`, `"limits":{"max_duration_ms":1000}`, 1)))
	if err != nil {
		t.Fatal(err)
	}
	waitStarted(t, h.run.runner)
	h.waitTimers(t, 1)
	_ = h.run.clock.Advance(999 * time.Millisecond)
	if s, _ := c.RunState(rh); s != runActive || len(h.run.output.Records()) != 0 {
		t.Fatal("attempt ended before its deadline")
	}
	_ = h.run.clock.Advance(time.Millisecond)
	waitTerminal(t, c, rh)
	got := records(t, h.run.output)
	if len(got) != 1 || got[0].Execution.State != "incomplete" || got[0].Execution.Reason != string(task.AttemptDeadlineExceeded) {
		t.Fatalf("records = %+v", got)
	}
}

// Close, clearing and policy replacement cancel the active run, wait for it
// and leave its attempt recorded as cancelled_by_host.
// requirement-ref: REQ-LINKAGE-0023
func TestLifecycleTransitionsCancelRuns(t *testing.T) {
	transitions := map[string]func(*testing.T, *Client) error{
		"close": func(_ *testing.T, c *Client) error { return c.Close() },
		"clear": func(_ *testing.T, c *Client) error { return c.ClearMeasurementData() },
		"policy replacement": func(t *testing.T, c *Client) error {
			in := synPolicyInput(3, "reg-a", func(p *task.Permissions) { p.Tests = []task.TestVersion{tcp01} })
			in.Targets[0].Test = tcp01
			next, err := task.NewLocalPolicy(in)
			if err != nil {
				t.Fatal("synthetic policy rejected")
			}
			return c.replacePolicy(next)
		},
	}
	for name, transition := range transitions {
		t.Run(name, func(t *testing.T) {
			h := runningHarness(t)
			h.adapters.clearStorage = func(context.Context) error { return nil }
			c := h.initialized(t)
			h.startHeld(t, c, "life-cancel", "")
			if err := transition(t, c); err != nil {
				t.Fatal("transition:", err)
			}
			got := records(t, h.run.output)
			if len(got) != 1 || got[0].Execution.State != "incomplete" || got[0].Execution.Reason != "cancelled_by_host" {
				t.Fatalf("records = %+v", got)
			}
		})
	}
}
