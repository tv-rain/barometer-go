package barometer

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/tv-rain/barometer-go/internal/persistence"
	recordstore "github.com/tv-rain/barometer-go/internal/testkit/records"
)

// Synthetic facade cases for persistent clearing (DEV-564) over the
// production persistence.State and pending store on the fault-injectable
// record port. They establish no runtime, platform, device-durability or
// deletion guarantee.

// failClearing makes every later record deletion fail, so a clear stops
// after writing its marker.
func (h *harness) failClearing() { h.run.output.Inject(recordstore.FailBeforeDelete, true) }

// linkageKeys is a synthetic linkage key store. It holds no key material; it
// counts deletions and derivations so tests can show clearing deletes the key
// and creates no replacement.
type linkageKeys struct {
	mu               sync.Mutex
	deletes, derives int
	fail             bool
}

func (k *linkageKeys) DeriveHMACSHA256(context.Context, []byte) ([32]byte, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.derives++
	return [32]byte{}, errors.New("synthetic linkage keys derive nothing")
}

func (k *linkageKeys) Delete(context.Context) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.fail {
		return errors.New("synthetic linkage key deletion failure")
	}
	k.deletes++
	return nil
}

func (k *linkageKeys) counts() (deletes, derives int) {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.deletes, k.derives
}

// stepRecords counts the record-port steps of a clear once armed and fails
// the failAt-th one, before it takes effect or, with after set, after it.
// Reads are not steps.
type stepRecords struct {
	persistence.Records
	mu            sync.Mutex
	armed, after  bool
	steps, failAt int
}

func (s *stepRecords) arm(failAt int, after bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.armed, s.after, s.steps, s.failAt = true, after, 0, failAt
}

func (s *stepRecords) disarm() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.armed = false
}

// step reports whether this step fails and whether it takes effect first.
func (s *stepRecords) step() (fail, apply bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.armed {
		return false, true
	}
	s.steps++
	if s.steps != s.failAt {
		return false, true
	}
	return true, s.after
}

func (s *stepRecords) Put(ctx context.Context, k persistence.RecordKey, v []byte) error {
	fail, apply := s.step()
	if apply {
		if err := s.Records.Put(ctx, k, v); err != nil || !fail {
			return err
		}
	}
	return persistence.ErrRecordsUnavailable
}

func (s *stepRecords) Delete(ctx context.Context, k persistence.RecordKey) error {
	fail, apply := s.step()
	if apply {
		if err := s.Records.Delete(ctx, k); err != nil || !fail {
			return err
		}
	}
	return persistence.ErrRecordsUnavailable
}

func (s *stepRecords) List(ctx context.Context, c persistence.RecordClass) ([]persistence.RecordKey, error) {
	if fail, _ := s.step(); fail {
		return nil, persistence.ErrRecordsUnavailable
	}
	return s.Records.List(ctx, c)
}

// clearHarness binds the production persistence.State as replay store and
// run ledger over the same record port as pending results, through a
// stepRecords wrapper, plus a synthetic linkage key store.
func clearHarness(t *testing.T) (*harness, *stepRecords, *linkageKeys) {
	t.Helper()
	h := runningHarness(t)
	steps := &stepRecords{Records: h.run.output.Store}
	keys := &linkageKeys{}
	h.adapters.records, h.adapters.linkageKeys = steps, keys
	h.reopenState(t)
	return h, steps, keys
}

// reopenState opens a new State over the harness's record port, as a
// restarted process would.
func (h *harness) reopenState(t *testing.T) {
	t.Helper()
	st, err := persistence.OpenState(context.Background(), h.adapters.records, 64, h.clock.Clock.Now())
	if err != nil {
		t.Fatal("state rejected")
	}
	h.adapters.replay, h.adapters.run.ledger = st.Replay(), st.Runs()
}

// classLens counts the durable records of every class, the marker included.
func classLens(h *harness) map[persistence.RecordClass]int {
	n := map[persistence.RecordClass]int{}
	for c := persistence.RecordClass(1); c.Valid(); c++ {
		n[c] = h.run.output.Len(c)
	}
	return n
}

func emptyScope(h *harness) bool {
	for _, n := range classLens(h) {
		if n != 0 {
			return false
		}
	}
	return true
}

// populated runs one task to completion, leaving a replay record, run
// accounting and a pending bundle, and returns the task's bytes.
func populated(t *testing.T, h *harness, c *Client, id string) string {
	t.Helper()
	completedRun(t, h, c, id)
	n := classLens(h)
	if n[persistence.ReplayClass] != 1 || n[persistence.RunClass] != 1 || n[persistence.PendingClass] != 1 {
		t.Fatalf("expected one record of each class, got %v", n)
	}
	return runTask(id, "", "", false)
}

// Clearing deletes replay records, run accounting, pending bundles and the
// linkage key, creates no replacement secret and leaves the client
// uninitialized; a later Initialize starts from an empty scope.
// requirement-ref: REQ-LINKAGE-0023
// requirement-ref: REQ-LINKAGE-0024
func TestClearDeletesEveryStateClass(t *testing.T) {
	h, _, keys := clearHarness(t)
	c := h.initialized(t)
	populated(t, h, c, "clr-every")
	if err := c.ClearMeasurementData(); err != nil || stateOf(c) != uninitialized {
		t.Fatal("clear failed:", codeOf(err))
	}
	if !emptyScope(h) || len(h.run.output.Dump()) != 0 {
		t.Fatalf("records remain after clearing: %v", classLens(h))
	}
	if deletes, derives := keys.counts(); deletes != 1 || derives != 0 {
		t.Fatalf("linkage key: %d deletions, %d derivations", deletes, derives)
	}
	if _, err := c.reserve(); codeOf(err) != "not_initialized" {
		t.Fatal("cleared client must require explicit initialization")
	}
	if err := c.Initialize(`{}`); err != nil {
		t.Fatal("initialize after clearing:", err)
	}
	if b, err := c.NextBundle(); b != "" || err != nil {
		t.Fatal("a cleared bundle was returned")
	}
	if _, derives := keys.counts(); derives != 0 {
		t.Fatal("clearing or reinitialization created a secret")
	}
}

// Repeated clearing, before initialization and after close, reaches the same
// storage scope; a client with no bound scope never reports success.
func TestRepeatedClearingReachesTheSameScope(t *testing.T) {
	h, _, _ := clearHarness(t)
	c := h.initialized(t)
	populated(t, h, c, "clr-repeat")
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	for i := range 2 {
		if err := c.ClearMeasurementData(); err != nil || stateOf(c) != closed || !emptyScope(h) {
			t.Fatalf("recovery clear %d after close: %s, %v", i, codeOf(err), classLens(h))
		}
	}

	fresh := h.client(t)
	seedPending(h, 0, h.run.clock.Now(), "AAAAAAAAAAAAAAAAAAAAAA")
	for i := range 2 {
		if err := fresh.ClearMeasurementData(); err != nil || stateOf(fresh) != uninitialized || !emptyScope(h) {
			t.Fatalf("clear %d before initialization: %s", i, codeOf(err))
		}
	}

	var zero Client
	if codeOf(zero.ClearMeasurementData()) != "clearing_incomplete" {
		t.Fatal("a client without a storage scope reported clearing success")
	}
}

// Every record-port step of a clear is fault-injected, failing before and
// after it takes effect. A failure returns clearing_incomplete and keeps all
// activity blocked; after a restart, a new client is blocked exactly while
// the marker is stored; recovery clearing then empties the scope.
func TestClearFaultAtEachStep(t *testing.T) {
	for _, after := range []bool{false, true} {
		for step := 1; ; step++ {
			h, steps, _ := clearHarness(t)
			c := h.initialized(t)
			taskJSON := populated(t, h, c, fmt.Sprintf("clr-step-%d", step))
			steps.arm(step, after)
			err := c.ClearMeasurementData()
			total := steps.steps
			steps.disarm()
			if step > total {
				// No step failed: every step has been covered.
				if err != nil || !emptyScope(h) {
					t.Fatalf("after=%v: unfaulted clear failed", after)
				}
				if total < 10 {
					t.Fatalf("after=%v: only %d clearing steps", after, total)
				}
				break
			}
			name := fmt.Sprintf("after=%v/step=%d", after, step)
			if codeOf(err) != "clearing_incomplete" || stateOf(c) != blocked {
				t.Fatalf("%s: got %s in state %v", name, codeOf(err), stateOf(c))
			}
			for call, got := range map[string]error{
				"Initialize":        c.Initialize(`{}`),
				"AdmitTask":         second(c.AdmitTask(taskJSON, h.embedded(t, taskJSON))),
				"StartRun":          second(c.StartRun("T.none")),
				"NextBundle":        second(c.NextBundle()),
				"AcknowledgeBundle": c.AcknowledgeBundle("none"),
			} {
				if codeOf(got) != "clearing_incomplete" {
					t.Fatalf("%s: %s after a failed clear returned %s", name, call, codeOf(got))
				}
			}

			// Restart: durable records survive, the in-memory client does not.
			h.run.output.Restart()
			h.reopenState(t)
			restarted := h.client(t)
			marker := h.run.output.Len(persistence.ClearingClass) == 1
			got := restarted.Initialize(`{}`)
			switch {
			case marker && (codeOf(got) != "clearing_incomplete" || stateOf(restarted) != blocked):
				t.Fatalf("%s: restart with the marker initialized: %s", name, codeOf(got))
			case !marker && got != nil:
				t.Fatalf("%s: restart without the marker failed: %s", name, codeOf(got))
			}
			if marker {
				if _, err := restarted.NextBundle(); codeOf(err) != "clearing_incomplete" {
					t.Fatalf("%s: bundle access after restart: %s", name, codeOf(err))
				}
			}
			if err := restarted.ClearMeasurementData(); err != nil || !emptyScope(h) || stateOf(restarted) != uninitialized {
				t.Fatalf("%s: recovery clearing failed: %s, %v", name, codeOf(err), classLens(h))
			}
			if err := restarted.Initialize(`{}`); err != nil {
				t.Fatalf("%s: initialize after recovery: %s", name, codeOf(err))
			}
		}
	}
}

func second[T any](_ T, err error) error { return err }

// A failed linkage-key deletion is an incomplete clear: the marker stays and
// a retry completes it.
func TestClearLinkageKeyFailure(t *testing.T) {
	h, _, keys := clearHarness(t)
	c := h.initialized(t)
	populated(t, h, c, "clr-key")
	keys.mu.Lock()
	keys.fail = true
	keys.mu.Unlock()
	if codeOf(c.ClearMeasurementData()) != "clearing_incomplete" || h.run.output.Len(persistence.ClearingClass) != 1 {
		t.Fatal("failed key deletion reported as cleared")
	}
	keys.mu.Lock()
	keys.fail = false
	keys.mu.Unlock()
	if err := c.ClearMeasurementData(); err != nil || !emptyScope(h) {
		t.Fatal("retry did not complete clearing")
	}
}

// An unreadable marker fails initialization closed without blocking.
func TestInitializeWithUnreadableMarker(t *testing.T) {
	h, _, _ := clearHarness(t)
	h.run.output.Inject(recordstore.FailGet, true)
	c := h.client(t)
	if codeOf(c.Initialize(`{}`)) != "not_initialized" || stateOf(c) != uninitialized {
		t.Fatal("initialized without reading the clearing marker")
	}
}

// Clearing racing admission, StartRun, NextBundle and AcknowledgeBundle
// leaves nothing behind: every entry point either finished before the clear
// quiesced the client or was refused.
func TestClearRacesEntryPoints(t *testing.T) {
	h, _, _ := clearHarness(t)
	c := h.initialized(t)
	populated(t, h, c, "clr-race-0")
	b, err := c.NextBundle()
	if err != nil {
		t.Fatal(err)
	}
	id := bundleID(t, b)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := range 4 {
		taskJSON := runTask(fmt.Sprintf("clr-race-%d", i+1), "", "", false)
		ref := h.embedded(t, taskJSON)
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if th, err := c.AdmitTask(taskJSON, ref); err == nil {
				_, _ = c.StartRun(th)
			}
			_, _ = c.NextBundle()
			_ = c.AcknowledgeBundle(id)
		}()
	}
	close(start)
	err = c.ClearMeasurementData()
	wg.Wait()
	if err != nil {
		t.Fatal("clear failed:", codeOf(err))
	}
	if !emptyScope(h) {
		t.Fatalf("a racing entry point left state behind: %v", classLens(h))
	}
}

// Every adapter field is either cleared through storageClearer or recorded
// here as holding no Barometer-owned persistent or sensitive state, so a new
// state-holding adapter cannot be forgotten by clearing.
func TestClearingCoversEveryAdapter(t *testing.T) {
	covered := map[string]string{
		"clock":         "exempt: trusted time source",
		"registry":      "exempt: reviewed static registrations",
		"policy":        "exempt: installed local policy; the activated generation floor survives clearing in memory",
		"deliveries":    "lifecycle: delivery references are invalidated before deletion",
		"replay":        "cleared: ReplayClass records; Forgetter drops the State index",
		"approvals":     "exempt: reviewed configuration approvals",
		"linkageKeys":   "cleared: eraser deletes the protected key",
		"clearStorage":  "the clearing capability itself",
		"run":           "see runAdapters",
		"records":       "cleared: every record class",
		"pendingLimits": "exempt: limits",
		"deadline":      "exempt: lifecycle deadline",
	}
	runCovered := map[string]string{
		"clock":          "exempt: engine clock",
		"jitter":         "exempt: jitter source",
		"ledger":         "cleared: RunClass records; Forgetter drops the State index",
		"paths":          "exempt: path seam holds no stored state; directPath caches no context",
		"network":        "exempt: host-reported capability",
		"implementation": "exempt: producer identity",
		"consent":        "exempt: host-reported capability",
		"device":         "exempt: host-reported capability",
		"zone":           "exempt: host-reported capability",
		"registry":       "exempt: copy of the adapters' registry",
	}
	for typ, want := range map[reflect.Type]map[string]string{reflect.TypeOf(adapters{}): covered, reflect.TypeOf(runAdapters{}): runCovered} {
		if typ.NumField() != len(want) {
			t.Errorf("%s has %d fields, %d classified", typ, typ.NumField(), len(want))
		}
		for i := range typ.NumField() {
			if _, ok := want[typ.Field(i).Name]; !ok {
				t.Errorf("%s.%s is not classified for clearing", typ, typ.Field(i).Name)
			}
		}
	}

	// The production State's views are the forgetters storageClearer finds.
	h, _, _ := clearHarness(t)
	for _, v := range []any{h.adapters.replay, h.adapters.run.ledger} {
		if _, ok := v.(persistence.Forgetter); !ok {
			t.Fatal("production State view is not a Forgetter")
		}
	}
}
