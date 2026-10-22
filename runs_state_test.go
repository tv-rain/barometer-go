package barometer

import (
	"context"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/admission"
	"github.com/tv-rain/barometer-go/internal/persistence"
	staterecords "github.com/tv-rain/barometer-go/internal/testkit/records"
)

// stateHarness binds the production replay store and run ledger, one
// persistence.State over the fault-injectable record port, in place of the
// synthetic stores.
type stateHarness struct {
	*harness
	records *staterecords.Store
	state   *persistence.State
}

func productionStateHarness(t *testing.T) *stateHarness {
	t.Helper()
	h := &stateHarness{harness: runningHarness(t), records: staterecords.New()}
	h.reopen(t)
	return h
}

// reopen models a process restart: the record port keeps durable state, the
// State and the delivery store are rebuilt.
func (h *stateHarness) reopen(t *testing.T) {
	t.Helper()
	h.records.Restart()
	var err error
	if h.state, err = persistence.OpenState(context.Background(), h.records, 8, h.clock.Clock.Now()); err != nil {
		t.Fatal("state rejected")
	}
	if h.deliveries, err = admission.NewDeliveryStore(h.clock, admission.DeliveryLimits{Capacity: 8, Lifetime: time.Minute}, 2); err != nil {
		t.Fatal(err)
	}
	h.adapters.deliveries, h.adapters.replay, h.adapters.run.ledger = h.deliveries, h.state.Replay(), h.state.Runs()
}

func (h *stateHarness) runRecord(t *testing.T, id string) (persistence.RunRecord, bool) {
	t.Helper()
	rec, ok, err := h.state.Runs().Lookup(context.Background(), key(id), h.clock.Clock.Now())
	if err != nil {
		t.Fatal("run ledger unavailable")
	}
	return rec, ok
}

func TestStartRunOnProductionState(t *testing.T) {
	h := productionStateHarness(t)
	h.adapters.run.jitter = fixedJitter(time.Minute)
	c := h.initialized(t)
	taskJSON := runTask("state-1", interval(4), "", false)
	th := h.admit(t, c, taskJSON)
	start := func(c *Client, th, want string) string {
		t.Helper()
		rh, err := c.StartRun(th)
		if codeOf(err) != want {
			t.Fatalf("got %s, want %s", codeOf(err), want)
		}
		if err == nil {
			waitTerminal(t, c, rh)
		}
		return rh
	}
	first := start(c, th, "ok")
	if again := start(c, th, "ok"); again != first {
		t.Fatal("a repeated start did not resolve to the run")
	}
	if rec, ok := h.runRecord(t, "state-1"); !ok || rec.Runs != 1 || rec.Next != 1 {
		t.Fatalf("run not recorded: %+v", rec)
	}
	h.advance(6*time.Hour + time.Minute)
	start(c, th, "ok")

	// A process restart keeps the replay and run records: the redelivered
	// task resolves to its replay record and its accounting is not reset.
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	h.reopen(t)
	c = h.initialized(t)
	th = h.admit(t, c, taskJSON)
	start(c, th, "run_not_due")
	h.advance(6 * time.Hour)
	start(c, th, "ok")
	if rec, ok := h.runRecord(t, "state-1"); !ok || rec.Runs != 3 || h.run.runner.Calls() != 3 {
		t.Fatalf("accounting across restart: %+v, %d calls", rec, h.run.runner.Calls())
	}

	// Retention: the run record goes with the replay record, one day after
	// not_after.
	h.advance(48 * time.Hour)
	if _, ok := h.runRecord(t, "state-1"); ok || h.records.Len(persistence.RunClass) != 0 || h.records.Len(persistence.ReplayClass) != 0 {
		t.Fatal("run or replay record outlived its retention")
	}
}

func TestStartRunProductionLedgerFaults(t *testing.T) {
	for name, fault := range map[string]staterecords.Fault{"read": staterecords.FailGet, "before write": staterecords.FailBeforePut, "crash": staterecords.CrashDuringPut} {
		t.Run(name, func(t *testing.T) {
			h := productionStateHarness(t)
			c := h.initialized(t)
			th := h.admit(t, c, runTask("fault-1", "", "", false))
			h.records.Inject(fault, true)
			if _, err := c.StartRun(th); codeOf(err) != "run_not_due" {
				t.Fatal(codeOf(err))
			}
			h.records.Inject(fault, false)
			if h.run.runner.Calls() != 0 || len(c.rt.runs.entries) != 0 {
				t.Fatal("a run started without durable accounting")
			}
			rh, err := c.StartRun(th)
			if err != nil {
				t.Fatal("recovered ledger:", err)
			}
			waitTerminal(t, c, rh)
		})
	}
	t.Run("uncertain reservation", func(t *testing.T) {
		h := productionStateHarness(t)
		c := h.initialized(t)
		th := h.admit(t, c, runTask("fault-2", `"starts_at":"`+epoch.Add(time.Hour).Format(time.RFC3339)+`"`, "", false))
		if _, err := c.StartRun(th); codeOf(err) != "run_not_due" {
			t.Fatal(codeOf(err))
		}
		h.advance(time.Hour)
		h.records.Inject(staterecords.FailAfterPut, true)
		if _, err := c.StartRun(th); codeOf(err) != "run_not_due" {
			t.Fatal("uncertain reservation:", codeOf(err))
		}
		h.records.Inject(staterecords.FailAfterPut, false)
		// The run is lost, never added, including after a restart.
		for range 2 {
			if rec, _ := h.runRecord(t, "fault-2"); rec.Runs != 1 {
				t.Fatal("reservation not durable")
			}
			if _, err := c.StartRun(th); codeOf(err) != "run_not_due" || h.run.runner.Calls() != 0 {
				t.Fatal("an uncertain reservation started or allowed a run:", codeOf(err))
			}
			if err := c.Close(); err != nil {
				t.Fatal(err)
			}
			h.reopen(t)
			c = h.initialized(t)
			th = h.admit(t, c, runTask("fault-2", `"starts_at":"`+epoch.Add(time.Hour).Format(time.RFC3339)+`"`, "", false))
		}
	})
}
