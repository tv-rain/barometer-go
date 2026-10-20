package barometer

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/admission"
	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/persistence"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/registry"
	"github.com/tv-rain/barometer-go/internal/taskdecode"
	"github.com/tv-rain/barometer-go/internal/testkit"
	"github.com/tv-rain/barometer-go/internal/testkit/replay"
)

var epoch = time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)

// fakeClock adapts the manually advanced testkit clock to platform.Clock.
type fakeClock struct{ *testkit.Clock }

func (c fakeClock) Now(context.Context) (time.Time, error) { return c.Clock.Now(), nil }

// SYN-01 is a synthetic test-scoped registry entry, never a production entry.
type synParameters struct{}

func (synParameters) TaskParameters() {}

type synCodec struct{}

func (synCodec) Definition() taskdecode.Definition {
	return taskdecode.Definition{TestID: "SYN-01", TestVersion: "0.1.0",
		Input:   &jsondecode.Schema{Kind: jsondecode.Object, Fields: map[string]jsondecode.Field{}},
		Options: &jsondecode.Schema{Kind: jsondecode.Object, Fields: map[string]jsondecode.Field{}}}
}

func (synCodec) Normalize(_, _ *jsondecode.Value) (task.Parameters, error) {
	return synParameters{}, nil
}

type noTargets struct{}

func (noTargets) Targets(task.Parameters) ([]registry.Target, bool) { return nil, false }

// harness holds the synthetic adapters bound to one client. The policy is the
// zero value, so a later admission would fail closed with
// task_policy_unavailable; the lifecycle shell does not admit tasks.
type harness struct {
	clock      fakeClock
	deliveries *admission.DeliveryStore
	replay     *replay.Store
	adapters   adapters
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{clock: fakeClock{testkit.NewClock(epoch)}}
	var err error
	if h.deliveries, err = admission.NewDeliveryStore(h.clock, admission.DeliveryLimits{Capacity: 8, Lifetime: time.Minute}, 2); err != nil {
		t.Fatal("synthetic delivery store rejected")
	}
	if h.replay, err = replay.New(8); err != nil {
		t.Fatal("synthetic replay store rejected")
	}
	reg, err := registry.New([]registry.Entry{{Test: task.TestVersion{ID: "SYN-01", Version: "0.1.0"}, Codec: synCodec{}, BaselineRisk: task.LowRisk, Targets: noTargets{}, Proxy: registry.NoProxy{}}})
	if err != nil {
		t.Fatal("synthetic registry rejected")
	}
	h.adapters = adapters{clock: h.clock, registry: reg, deliveries: h.deliveries, replay: h.replay, deadline: 200 * time.Millisecond}
	return h
}

func (h *harness) client(t *testing.T) *Client {
	t.Helper()
	c := NewClient()
	if err := c.bind(h.adapters); err != nil {
		t.Fatal("bind rejected:", err)
	}
	return c
}

func (h *harness) initialized(t *testing.T) *Client {
	t.Helper()
	c := h.client(t)
	if err := c.Initialize(`{}`); err != nil {
		t.Fatal("initialize rejected:", err)
	}
	return c
}

func codeOf(err error) string {
	if err == nil {
		return "ok"
	}
	return err.Error()
}

func stateOf(c *Client) state {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}

// gate blocks the clear hook until released, holding the client in the
// transitioning state without any lock held.
type gate struct {
	entered, release chan struct{}
	result           error
}

func newGate() *gate { return &gate{entered: make(chan struct{}), release: make(chan struct{})} }

func (g *gate) clear(ctx context.Context) error {
	close(g.entered)
	select {
	case <-g.release:
	case <-ctx.Done():
	}
	return g.result
}

func TestLifecycleTransitionTable(t *testing.T) {
	type setup func(t *testing.T, h *harness) *Client
	setups := map[string]setup{
		"zero value":       func(*testing.T, *harness) *Client { return &Client{} },
		"uninitialized":    func(t *testing.T, h *harness) *Client { return h.client(t) },
		"initialized":      func(t *testing.T, h *harness) *Client { return h.initialized(t) },
		"clear incomplete": func(t *testing.T, h *harness) *Client { c := h.initialized(t); _ = c.ClearMeasurementData(); return c },
		"close incomplete": func(t *testing.T, h *harness) *Client {
			c := h.initialized(t)
			if _, err := c.reserve(); err != nil {
				t.Fatal(err)
			}
			if err := c.Close(); codeOf(err) != "cleanup_incomplete" {
				t.Fatal("expected cleanup_incomplete, got", codeOf(err))
			}
			c.release()
			return c
		},
		"closed": func(t *testing.T, h *harness) *Client { c := h.initialized(t); _ = c.Close(); return c },
	}
	type outcome struct {
		code  string
		state state
	}
	cases := []struct {
		from, method string
		want         outcome
	}{
		{"zero value", "Initialize", outcome{"not_initialized", uninitialized}},
		{"zero value", "reserve", outcome{"not_initialized", uninitialized}},
		{"zero value", "Clear", outcome{"clearing_incomplete", blocked}},
		{"zero value", "Close", outcome{"ok", closed}},
		{"uninitialized", "Initialize", outcome{"ok", initialized}},
		{"uninitialized", "reserve", outcome{"not_initialized", uninitialized}},
		{"uninitialized", "Clear", outcome{"clearing_incomplete", blocked}},
		{"uninitialized", "Close", outcome{"ok", closed}},
		{"initialized", "Initialize", outcome{"already_initialized", initialized}},
		{"initialized", "reserve", outcome{"ok", initialized}},
		{"initialized", "Clear", outcome{"clearing_incomplete", blocked}},
		{"initialized", "Close", outcome{"ok", closed}},
		{"clear incomplete", "Initialize", outcome{"clearing_incomplete", blocked}},
		{"clear incomplete", "reserve", outcome{"clearing_incomplete", blocked}},
		{"clear incomplete", "Clear", outcome{"clearing_incomplete", blocked}},
		{"clear incomplete", "Close", outcome{"ok", closed}},
		{"close incomplete", "Initialize", outcome{"cleanup_incomplete", blocked}},
		{"close incomplete", "reserve", outcome{"cleanup_incomplete", blocked}},
		{"close incomplete", "Clear", outcome{"clearing_incomplete", blocked}},
		{"close incomplete", "Close", outcome{"ok", closed}},
		{"closed", "Initialize", outcome{"client_closed", closed}},
		{"closed", "reserve", outcome{"client_closed", closed}},
		{"closed", "Clear", outcome{"clearing_incomplete", closed}},
		{"closed", "Close", outcome{"ok", closed}},
	}
	for _, tc := range cases {
		t.Run(tc.from+"/"+tc.method, func(t *testing.T) {
			h := newHarness(t)
			c := setups[tc.from](t, h)
			var err error
			switch tc.method {
			case "Initialize":
				err = c.Initialize(`{}`)
			case "reserve":
				if _, err = c.reserve(); err == nil {
					c.release()
				}
			case "Clear":
				err = c.ClearMeasurementData()
			case "Close":
				err = c.Close()
			}
			if got := (outcome{codeOf(err), stateOf(c)}); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestTransitioningRefusesEveryCall(t *testing.T) {
	for _, from := range []string{"initialized", "closed"} {
		t.Run(from, func(t *testing.T) {
			h := newHarness(t)
			g := newGate()
			h.adapters.clearStorage = g.clear
			c := h.initialized(t)
			if from == "closed" {
				if err := c.Close(); err != nil {
					t.Fatal(err)
				}
			}
			done := make(chan error)
			go func() { done <- c.ClearMeasurementData() }()
			<-g.entered
			want := map[string]string{"Initialize": "transition_in_progress", "Clear": "transition_in_progress", "reserve": "transition_in_progress", "Close": "transition_in_progress"}
			if from == "closed" {
				// A recovery clear keeps a closed client closed.
				want = map[string]string{"Initialize": "client_closed", "Clear": "transition_in_progress", "reserve": "client_closed", "Close": "ok"}
			}
			_, reserveErr := c.reserve()
			got := map[string]string{"Initialize": codeOf(c.Initialize(`{}`)), "Clear": codeOf(c.ClearMeasurementData()), "reserve": codeOf(reserveErr), "Close": codeOf(c.Close())}
			for k := range want {
				if got[k] != want[k] {
					t.Errorf("%s during transition: got %s, want %s", k, got[k], want[k])
				}
			}
			close(g.release)
			if err := <-done; err != nil {
				t.Fatal("clear with succeeding hook:", err)
			}
			wantState := uninitialized
			if from == "closed" {
				wantState = closed
			}
			if s := stateOf(c); s != wantState {
				t.Fatalf("state %v, want %v", s, wantState)
			}
		})
	}
}

func TestNilAndZeroValueClients(t *testing.T) {
	var nilClient *Client
	if codeOf(nilClient.Initialize(`{}`)) != "not_initialized" || nilClient.Close() != nil || codeOf(nilClient.ClearMeasurementData()) != "clearing_incomplete" {
		t.Fatal("nil client must fail closed without panicking")
	}
	var zero Client
	if codeOf(zero.Initialize(`{}`)) != "not_initialized" {
		t.Fatal("zero-value client has no adapters and must not initialize")
	}
	if codeOf(NewClient().Initialize(`{}`)) != "not_initialized" {
		t.Fatal("NewClient has no production adapters and must fail closed")
	}
	if _, err := zero.reserve(); codeOf(err) != "not_initialized" {
		t.Fatal("zero-value client carries no authority")
	}
}

func TestInitializeFailsClosed(t *testing.T) {
	var typedNilClock *fakeClock
	var typedNilReplay *replay.Store
	edits := map[string]func(*adapters){
		"no clock":          func(a *adapters) { a.clock = nil },
		"typed nil clock":   func(a *adapters) { a.clock = typedNilClock },
		"no registry":       func(a *adapters) { a.registry = nil },
		"no delivery store": func(a *adapters) { a.deliveries = nil },
		"no replay store":   func(a *adapters) { a.replay = nil },
		"typed nil replay":  func(a *adapters) { a.replay = typedNilReplay },
	}
	for name, edit := range edits {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t)
			edit(&h.adapters)
			c := h.client(t)
			if err := c.Initialize(`{}`); codeOf(err) != "not_initialized" || stateOf(c) != uninitialized {
				t.Fatalf("missing adapter: got %s in state %v", codeOf(err), stateOf(c))
			}
		})
	}
	configs := map[string]string{
		"not json":            `not json`,
		"empty":               ``,
		"task json":           `{"task_id":"t","test":{"id":"SYN-01","version":"0.1.0"}}`,
		"trust assertion":     `{"authenticated":true}`,
		"policy section":      `{"local_policy":{"generation":9}}`,
		"duplicate member":    `{"proxy":{},"proxy":{}}`,
		"unapproved endpoint": `{"network_context":{"primary_endpoint":"https://ctx.example.test/"}}`,
		"unapproved linkage":  `{"result_linkage":{"mode":"rotating_installation","scope":"s","rotation_period_seconds":604800}}`,
		"trailing data":       `{} {}`,
	}
	for name, cfg := range configs {
		t.Run(name, func(t *testing.T) {
			c := newHarness(t).client(t)
			if err := c.Initialize(cfg); codeOf(err) != "not_initialized" || stateOf(c) != uninitialized {
				t.Fatalf("got %s in state %v", codeOf(err), stateOf(c))
			}
			// A failed attempt publishes nothing and leaves a later valid
			// initialization possible.
			if err := c.Initialize(`{}`); err != nil {
				t.Fatal("valid initialization after failure:", err)
			}
		})
	}
}

func TestBindIsRestrictedToUninitializedClients(t *testing.T) {
	h := newHarness(t)
	c := h.initialized(t)
	if codeOf(c.bind(h.adapters)) != "already_initialized" {
		t.Fatal("adapters must not change under an initialized generation")
	}
	_ = c.Close()
	if codeOf(c.bind(h.adapters)) != "client_closed" {
		t.Fatal("closed client accepted adapters")
	}
	bad := h.adapters
	bad.deadline = maxLifecycleDeadline + time.Nanosecond
	if codeOf(NewClient().bind(bad)) != "not_initialized" {
		t.Fatal("deadline above the reviewed maximum accepted")
	}
}

func TestConcurrentLifecycleCalls(t *testing.T) {
	h := newHarness(t)
	c := h.client(t)
	const n = 32
	var wg sync.WaitGroup
	results := make(chan string, n)
	for range n {
		wg.Add(1)
		go func() { defer wg.Done(); results <- codeOf(c.Initialize(`{}`)) }()
	}
	wg.Wait()
	close(results)
	successes := 0
	for r := range results {
		switch r {
		case "ok":
			successes++
		case "already_initialized", "transition_in_progress":
		default:
			t.Fatal("unexpected concurrent initialize result", r)
		}
	}
	if successes != 1 || c.generation != 1 {
		t.Fatalf("got %d successful initializations and generation %d", successes, c.generation)
	}

	// Mixed concurrent calls end in exactly one closed state, never a panic.
	results = make(chan string, 3*n)
	for range n {
		wg.Add(3)
		go func() {
			defer wg.Done()
			if _, err := c.reserve(); err == nil {
				c.release()
			}
		}()
		go func() { defer wg.Done(); results <- "close:" + codeOf(c.Close()) }()
		go func() { defer wg.Done(); results <- "clear:" + codeOf(c.ClearMeasurementData()) }()
	}
	wg.Wait()
	close(results)
	for r := range results {
		switch r {
		case "close:ok", "close:transition_in_progress", "close:cleanup_incomplete", "clear:clearing_incomplete", "clear:transition_in_progress":
		default:
			t.Fatal("unexpected concurrent result", r)
		}
	}
	for c.Close() != nil {
	}
	if stateOf(c) != closed || c.active != 0 {
		t.Fatalf("final state %v with %d reservations", stateOf(c), c.active)
	}
}

func TestTransitionsWaitForReservedEntryPoints(t *testing.T) {
	h := newHarness(t)
	h.adapters.deadline = 5 * time.Second
	c := h.initialized(t)
	if _, err := c.reserve(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error)
	go func() { done <- c.Close() }()
	// The waiting transition holds no lock: new entry points are refused
	// rather than blocked.
	for {
		_, err := c.reserve()
		if codeOf(err) == "transition_in_progress" {
			break
		}
		if err == nil {
			c.release()
		}
		time.Sleep(time.Millisecond)
	}
	c.release()
	if err := <-done; err != nil {
		t.Fatal("close after the reserved entry point finished:", err)
	}
	if stateOf(c) != closed {
		t.Fatal("close did not complete")
	}
}

func TestExpiredDrainKeepsTheClientBlocked(t *testing.T) {
	h := newHarness(t)
	h.adapters.deadline = 20 * time.Millisecond
	c := h.initialized(t)
	if _, err := c.reserve(); err != nil {
		t.Fatal(err)
	}
	if codeOf(c.Close()) != "cleanup_incomplete" || stateOf(c) != blocked {
		t.Fatal("close must report incomplete cleanup and stay blocked")
	}
	if codeOf(c.ClearMeasurementData()) != "clearing_incomplete" {
		t.Fatal("clear must not succeed while an entry point is still running")
	}
	c.release()
	if c.Close() != nil || c.Close() != nil || stateOf(c) != closed {
		t.Fatal("close must be retryable and repeatable once work has finished")
	}
}

func TestCloseInvalidatesWithoutDeletingReplayState(t *testing.T) {
	h := newHarness(t)
	c := h.initialized(t)
	ctx := context.Background()
	rec, err := persistence.NewReplayRecord(persistence.ReplayKey{IssuerID: "embedded-issuer", TaskID: "task-1"}, sha256.Sum256([]byte("synthetic")), 2, task.Embedded, task.TimestampFromTime(epoch.Add(time.Hour)), epoch)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := h.replay.Commit(ctx, rec, epoch); err != nil {
		t.Fatal(err)
	}
	ref, err := h.deliveries.InstallEmbedded(ctx, "embedded-issuer", 2, []byte(`{"synthetic":"task"}`))
	if err != nil {
		t.Fatal(err)
	}
	rt := c.rt
	th, err := rt.tasks.add(admission.Admitted{})
	if err != nil {
		t.Fatal(err)
	}
	rh, err := rt.runs.add(runRecord{taskHandle: th})
	if err != nil {
		t.Fatal(err)
	}

	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := rt.tasks.lookup(th); codeOf(err) != "handle_unavailable" {
		t.Fatal("task handle survived close")
	}
	if _, err := rt.runs.lookup(rh); codeOf(err) != "handle_unavailable" {
		t.Fatal("run handle survived close")
	}
	if _, err := h.deliveries.Resolve(ctx, ref, []byte(`{"synthetic":"task"}`)); err == nil {
		t.Fatal("delivery record survived close")
	}
	if _, err := h.deliveries.InstallEmbedded(ctx, "embedded-issuer", 2, []byte(`{}`)); err == nil {
		t.Fatal("closed client's delivery store accepted a record")
	}
	if c.rt != nil || c.bound == nil || c.bound.deliveries != nil || c.bound.replay != nil || c.bound.clock != nil {
		t.Fatal("close must release runtime state and trusted adapters")
	}
	if h.replay.Len() != 1 {
		t.Fatal("close deleted replay state")
	}
	// Clearing without the persistent-clear capability deletes nothing and
	// does not report success.
	if codeOf(c.ClearMeasurementData()) != "clearing_incomplete" || h.replay.Len() != 1 {
		t.Fatal("clearing must not report success without deleting persistent state")
	}
}

func TestClearInvalidatesHandlesAndDeliveryReferences(t *testing.T) {
	h := newHarness(t)
	c := h.initialized(t)
	ctx := context.Background()
	ref, err := h.deliveries.InstallEmbedded(ctx, "embedded-issuer", 2, []byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	rt := c.rt
	th, _ := rt.tasks.add(admission.Admitted{})
	if codeOf(c.ClearMeasurementData()) != "clearing_incomplete" || stateOf(c) != blocked {
		t.Fatal("clear without the capability must stay blocked")
	}
	if _, err := rt.tasks.lookup(th); codeOf(err) != "handle_unavailable" {
		t.Fatal("handle survived clearing")
	}
	if _, err := h.deliveries.Resolve(ctx, ref, []byte(`{}`)); err == nil {
		t.Fatal("delivery reference survived clearing")
	}
	if _, err := c.reserve(); codeOf(err) != "clearing_incomplete" {
		t.Fatal("new work must stay blocked after incomplete clearing")
	}
}

func TestStaleHandlesAcrossReinitialization(t *testing.T) {
	h := newHarness(t)
	h.adapters.clearStorage = func(context.Context) error { return nil }
	c := h.initialized(t)
	first := c.rt
	th, _ := first.tasks.add(admission.Admitted{})
	rh, _ := first.runs.add(runRecord{taskHandle: th})
	if err := c.ClearMeasurementData(); err != nil || stateOf(c) != uninitialized {
		t.Fatal("clear with a succeeding capability must end uninitialized")
	}
	if _, err := c.reserve(); codeOf(err) != "not_initialized" {
		t.Fatal("cleared client must require explicit initialization")
	}
	if err := c.Initialize(`{}`); err != nil {
		t.Fatal(err)
	}
	second := c.rt
	if second == first || second.generation != first.generation+1 {
		t.Fatal("reinitialization must create a new generation")
	}
	if _, err := second.tasks.lookup(th); codeOf(err) != "handle_unavailable" {
		t.Fatal("task handle from an earlier generation resolved")
	}
	if _, err := second.runs.lookup(rh); codeOf(err) != "handle_unavailable" {
		t.Fatal("run handle from an earlier generation resolved")
	}
	other := newHarness(t).initialized(t)
	if _, err := other.rt.tasks.lookup(th); codeOf(err) != "handle_unavailable" {
		t.Fatal("handle resolved in another instance")
	}
	nh, _ := second.tasks.add(admission.Admitted{})
	if nh == th {
		t.Fatal("handle reused across generations")
	}
}

func TestHandleTableCapacity(t *testing.T) {
	for name, capacity := range map[string]int{"task": taskHandleCapacity, "run": runHandleCapacity} {
		for _, n := range []int{capacity - 1, capacity, capacity + 1} {
			t.Run(fmt.Sprintf("%s/%d", name, n), func(t *testing.T) {
				table := newHandleTable[runRecord]("X.", capacity)
				seen := map[string]bool{}
				for i := range n {
					h, err := table.add(runRecord{})
					if i >= capacity {
						if codeOf(err) != "capacity_exceeded" || h != "" {
							t.Fatal("full table must return capacity_exceeded")
						}
						continue
					}
					if err != nil || seen[h] {
						t.Fatal("add within capacity failed or repeated a handle")
					}
					seen[h] = true
				}
				// Exhaustion never evicts an existing handle.
				for h := range seen {
					if _, err := table.lookup(h); err != nil {
						t.Fatal("existing handle evicted")
					}
				}
				if n > capacity {
					for h := range seen {
						table.remove(h)
						break
					}
					if _, err := table.add(runRecord{}); err != nil {
						t.Fatal("freed slot not reusable")
					}
				}
			})
		}
	}
	if taskHandleCapacity > admission.MaxDeliveryCapacity || runHandleCapacity < 1 || taskHandleCapacity < 1 {
		t.Fatal("handle capacities outside the reviewed inputs")
	}
}

func TestHandlesAreOpaqueAndDistinctFromIdentifiers(t *testing.T) {
	c := newHarness(t).initialized(t)
	th, _ := c.rt.tasks.add(admission.Admitted{})
	rh, _ := c.rt.runs.add(runRecord{taskHandle: th})
	for _, h := range []string{th, rh} {
		raw := h[2:]
		if b, err := base64.RawURLEncoding.Strict().DecodeString(raw); err != nil || len(b) != 16 {
			t.Fatal("handle body is not 128 random bits")
		}
		// run_id and bundle_id are bare base64url; task_id is lowercase.
		if _, err := base64.RawURLEncoding.Strict().DecodeString(h); err == nil || strings.ToLower(h) == h {
			t.Fatal("handle could be mistaken for an identifier")
		}
	}
	if !strings.HasPrefix(th, taskHandlePrefix) || !strings.HasPrefix(rh, runHandlePrefix) {
		t.Fatal("handle kinds are not distinct")
	}
	for _, v := range []any{c.rt.tasks, c.rt.runs, c.rt, c} {
		for _, f := range []string{"%v", "%+v", "%#v", "%s"} {
			if out := fmt.Sprintf(f, v); strings.Contains(out, th[2:]) || strings.Contains(out, rh[2:]) {
				t.Fatal("formatting leaked a handle")
			}
		}
	}
	if _, err := c.rt.tasks.lookup(rh); codeOf(err) != "handle_unavailable" {
		t.Fatal("run handle resolved in the task table")
	}
	for _, h := range []string{"", "T.", "T.AAAAAAAAAAAAAAAAAAAAAA", strings.Repeat("T", 4096)} {
		if _, err := c.rt.tasks.lookup(h); codeOf(err) != "handle_unavailable" {
			t.Fatal("unknown handle resolved")
		}
	}
	c.rt.tasks.invalidate()
	if _, err := c.rt.tasks.add(admission.Admitted{}); codeOf(err) != "handle_unavailable" {
		t.Fatal("invalidated table accepted a handle")
	}
}

func TestPanicsAreRecoveredToFixedCodes(t *testing.T) {
	h := newHarness(t)
	h.adapters.clearStorage = func(context.Context) error { panic("CANARY-PANIC") }
	c := h.initialized(t)
	err := c.ClearMeasurementData()
	if codeOf(err) != "clearing_incomplete" || stateOf(c) != blocked || strings.Contains(fmt.Sprintf("%+v", err), "CANARY") {
		t.Fatal("panicking clear capability must become clearing_incomplete and keep the client blocked")
	}
	if c.Close() != nil || stateOf(c) != closed {
		t.Fatal("close must recover a blocked client")
	}
	if codeOf(c.ClearMeasurementData()) != "clearing_incomplete" || stateOf(c) != closed {
		t.Fatal("panicking recovery clear after close must keep the client closed")
	}
}

func TestErrorsCarryOnlyFixedCodes(t *testing.T) {
	fixed := map[string]bool{"not_initialized": true, "already_initialized": true, "client_closed": true, "transition_in_progress": true, "handle_unavailable": true,
		"bundle_unavailable": true, "run_not_due": true, "capacity_exceeded": true, "cleanup_incomplete": true, "clearing_incomplete": true}
	for c := notInitialized; c <= clearingIncomplete; c++ {
		if !fixed[c.String()] {
			t.Fatal("unregistered code", c.String())
		}
	}
	if len(fixed) != int(clearingIncomplete) {
		t.Fatal("code vocabulary out of step with docs/public-api.md")
	}
	const canary = "CANARY-7f3a"
	h := newHarness(t)
	h.adapters.clearStorage = func(context.Context) error { return errors.New(canary + " raw platform error") }
	c := h.client(t)
	var errs []error
	errs = append(errs,
		c.Initialize(`{"`+canary+`":"https://`+canary+`.example.test/"}`),
		c.Initialize(`{"network_context":{"primary_endpoint":"https://`+canary+`.example.test/"}}`))
	_ = c.Initialize(`{}`)
	th, _ := c.rt.tasks.add(admission.Admitted{})
	_, lookupErr := c.rt.tasks.lookup(canary + th)
	errs = append(errs, lookupErr, c.Initialize(`{}`), c.ClearMeasurementData())
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	errs = append(errs, c.Initialize(canary), c.ClearMeasurementData())
	for _, err := range errs {
		if err == nil {
			t.Fatal("expected an error")
		}
		for _, f := range []string{"%v", "%+v", "%#v", "%s"} {
			out := fmt.Sprintf(f, err)
			if strings.Contains(out, canary) || strings.Contains(out, th) {
				t.Fatal("error formatting leaked input")
			}
		}
		if !fixed[err.Error()] || errors.Unwrap(err) != nil {
			t.Fatal("error is not a fixed, unwrapped code:", err.Error())
		}
		// Identical failures produce identical errors.
		var op opError
		if !errors.As(err, &op) || op != (opError{code: op.code}) {
			t.Fatal("error carries more than a code")
		}
	}
}

var _ platform.Clock = fakeClock{}
