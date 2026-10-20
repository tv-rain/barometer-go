package barometer

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/admission"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/persistence"
	"github.com/tv-rain/barometer-go/internal/registry"
	"github.com/tv-rain/barometer-go/internal/testkit/delivery"
	"github.com/tv-rain/barometer-go/internal/testkit/replay"
)

// synTargets resolves every SYN-01 task to the one synthetic registration.
type synTargets struct{}

func (synTargets) Targets(p task.Parameters) ([]registry.Target, bool) {
	if _, ok := p.(synParameters); !ok {
		return nil, false
	}
	return []registry.Target{{Role: task.MeasurementTarget, Tuple: task.TargetTuple{Hostname: "reg-a.example.test"}}}, true
}

// synPermissions grant SYN-01 everything the registered limits allow.
func synPermissions() task.Permissions {
	return task.Permissions{
		Tests:           []task.TestVersion{{ID: "SYN-01", Version: "0.1.0"}},
		RegistrationIDs: []string{"reg-a"},
		Paths:           task.Paths{Direct: true, Proxy: true},
		MaxRisk:         task.HighRisk,
		AllowMediumRisk: true,
		Limits: task.PolicyLimits{
			Attempt: task.AttemptLimits{DurationMS: 60000, BytesSent: 65536, BytesReceived: 262144, Connections: 12},
			MaxRuns: 28, MaxAttempts: 2, MaxValiditySeconds: 604800,
			MinRepeatIntervalSeconds: 21600, MinRepeatJitterSeconds: 300,
			MinRetryDelayMS: 30000, MaxRetryDelayMS: 600000, MaxRetryJitterMS: 60000,
			MaxResultBytes: 65536, MaxBundleBytes: 262144, MaxResultsPerBundle: 16,
			MaxPendingBundles: 32, MaxPendingBytes: 4194304, MaxRetentionSeconds: 86400,
		},
	}
}

func permit(*task.Permissions) {}

// synPolicy is a synthetic policy of the given generation whose deployment and
// both issuers allow SYN-01 against one registration. Only "reg-a" matches the
// SYN-01 target; any other ID registers a different host. adjust narrows or
// widens the deployment and issuer permissions alike.
func synPolicy(t *testing.T, generation uint64, registration string, adjust func(*task.Permissions)) task.LocalPolicy {
	t.Helper()
	perm := synPermissions()
	perm.RegistrationIDs = []string{registration}
	adjust(&perm)
	policy, err := task.NewLocalPolicy(task.PolicyInput{Generation: generation, Permissions: perm,
		Issuers: []task.IssuerPermissions{{IssuerID: "embedded-issuer", Permissions: perm}, {IssuerID: "remote-issuer", Permissions: perm}},
		Targets: []task.TargetRegistration{{ID: registration, Test: task.TestVersion{ID: "SYN-01", Version: "0.1.0"}, Role: task.MeasurementTarget,
			Tuple: task.TargetTuple{Hostname: registration + ".example.test"}, Paths: task.Paths{Direct: true, Proxy: true}, BaselineRisk: task.LowRisk,
			Owner: "synthetic", Purpose: "synthetic", ReviewGeneration: 1, ExpiresAt: task.TimestampFromTime(epoch.Add(48 * time.Hour))}}})
	if err != nil {
		t.Fatal("synthetic policy rejected")
	}
	return policy
}

// admittingHarness binds a synthetic generation-2 policy in which the
// deployment, the embedded issuer and the remote issuer allow SYN-01 against
// the one registration. The client has no transport seam to record, so
// absence of network activity is checked on the dependency graph instead
// (TestAdmissionHasNoNetworkDependency).
func admittingHarness(t *testing.T) *harness {
	t.Helper()
	h := newHarness(t)
	policy := synPolicy(t, 2, "reg-a", permit)
	reg, err := registry.New([]registry.Entry{{Test: task.TestVersion{ID: "SYN-01", Version: "0.1.0"}, Codec: synCodec{}, BaselineRisk: task.LowRisk, Targets: synTargets{}, Proxy: registry.NoProxy{}}})
	if err != nil {
		t.Fatal("synthetic registry rejected")
	}
	h.adapters.policy, h.adapters.registry = policy, reg
	return h
}

// synTask is a valid SYN-01 task that expires validFor after the epoch.
func synTask(id string, validFor time.Duration) string {
	return fmt.Sprintf(`{"schema_version":"0.1.0","task_id":%q,"test_id":"SYN-01","test_version":"0.1.0","input":{},"options":{},"schedule":{"not_after":%q},"execution":{"limits":{}},"safety":{"risk_level":"low","requires_user_consent":false}}`,
		id, epoch.Add(validFor).Format(time.RFC3339))
}

func (h *harness) embedded(t *testing.T, taskJSON string) string {
	t.Helper()
	ref, err := h.deliveries.InstallEmbedded(context.Background(), "embedded-issuer", 2, []byte(taskJSON))
	if err != nil {
		t.Fatal("embedded install rejected")
	}
	return ref
}

func (h *harness) remote(t *testing.T, taskJSON string) string {
	t.Helper()
	v := delivery.Verifier{Key: []byte("synthetic-key"), IssuerID: "remote-issuer", PolicyGeneration: 2, IssuedAt: task.TimestampFromTime(h.clock.Clock.Now()), IssuerAuthenticated: true}
	ref, err := h.deliveries.InstallRemote(context.Background(), v, v.Seal([]byte(taskJSON)))
	if err != nil {
		t.Fatal("remote install rejected")
	}
	return ref
}

func taskHandles(c *Client) int {
	c.rt.tasks.mu.Lock()
	defer c.rt.tasks.mu.Unlock()
	return len(c.rt.tasks.entries)
}

// AdmitTask resolves the delivery record, runs every admission step and
// publishes a task handle only on acceptance, for both reviewed origins.
func TestAdmitTaskAcceptsReviewedOrigins(t *testing.T) {
	for _, origin := range []string{"embedded", "remote"} {
		t.Run(origin, func(t *testing.T) {
			h := admittingHarness(t)
			c := h.initialized(t)
			b := synTask("synthetic-task", time.Hour)
			ref := h.embedded
			if origin == "remote" {
				ref = h.remote
			}
			th, err := c.AdmitTask(b, ref(t, b))
			if err != nil || !strings.HasPrefix(th, taskHandlePrefix) {
				t.Fatal("valid task rejected:", codeOf(err))
			}
			a, err := c.rt.tasks.lookup(th)
			if err != nil || a.Redelivery() || !a.ExpiresAt().Equal(epoch.Add(time.Hour)) {
				t.Fatal("published handle does not resolve to the accepted task")
			}
			if h.replay.Commits() != 1 || h.replay.Len() != 1 || taskHandles(c) != 1 || len(c.rt.runs.entries) != 0 {
				t.Fatal("acceptance must commit one replay record, publish one handle and allocate no run")
			}
		})
	}
}

// An exact redelivery, through the same or a new delivery reference, returns
// the same handle without a new replay record, a run or a later expiry. Any
// change to the bytes under the same task_id is a conflict.
func TestAdmitTaskExactRedeliveryReturnsTheSameHandle(t *testing.T) {
	h := admittingHarness(t)
	c := h.initialized(t)
	ctx := context.Background()
	b := synTask("synthetic-task", time.Hour)
	ref := h.embedded(t, b)
	first, err := c.AdmitTask(b, ref)
	if err != nil {
		t.Fatal(err)
	}
	key := persistence.ReplayKey{IssuerID: "embedded-issuer", TaskID: "synthetic-task"}
	before, _, err := h.replay.Lookup(ctx, key, h.clock.Clock.Now())
	if err != nil {
		t.Fatal(err)
	}
	h.clock.Advance(30 * time.Second)
	for name, r := range map[string]string{"same reference": ref, "new reference": h.embedded(t, b)} {
		again, err := c.AdmitTask(b, r)
		if err != nil || again != first {
			t.Fatalf("%s: redelivery returned %s, %s", name, codeOf(err), map[bool]string{true: "the same handle", false: "a different handle"}[again == first])
		}
	}
	after, _, err := h.replay.Lookup(ctx, key, h.clock.Clock.Now())
	if err != nil || after != before {
		t.Fatal("redelivery changed the replay record or its expiry")
	}
	if h.replay.Commits() != 1 || taskHandles(c) != 1 || len(c.rt.runs.entries) != 0 {
		t.Fatal("redelivery wrote replay state, published a handle or allocated a run")
	}
	changed := synTask("synthetic-task", time.Hour+time.Second)
	if th, err := c.AdmitTask(changed, h.embedded(t, changed)); codeOf(err) != "task_id_conflict" || th != "" {
		t.Fatal("changed bytes under the same task_id must be task_id_conflict, got", codeOf(err))
	}
	if h.replay.Commits() != 1 || taskHandles(c) != 1 {
		t.Fatal("a conflict must leave no state behind")
	}
}

// Every rejection is atomic: no handle, no replay record and no run, with the
// class of the first failing step.
func TestAdmitTaskRejectionsAreAtomic(t *testing.T) {
	valid := synTask("synthetic-task", time.Hour)
	cases := []struct {
		name string
		run  func(t *testing.T, h *harness, c *Client) (string, error)
		want string
	}{
		{"unknown reference", func(t *testing.T, h *harness, c *Client) (string, error) {
			return c.AdmitTask(valid, "unknown-reference")
		}, "task_source_unauthenticated"},
		{"bytes differ from the record", func(t *testing.T, h *harness, c *Client) (string, error) {
			return c.AdmitTask(synTask("other-task", time.Hour), h.embedded(t, valid))
		}, "task_source_unauthenticated"},
		{"oversized bytes on a foreign reference", func(t *testing.T, h *harness, c *Client) (string, error) {
			return c.AdmitTask(strings.Repeat(" ", admission.MaxTaskBytes+1), h.embedded(t, valid))
		}, "task_source_unauthenticated"},
		{"complete task one byte over the limit", func(t *testing.T, h *harness, c *Client) (string, error) {
			b := valid[:len(valid)-1] + strings.Repeat(" ", admission.MaxTaskBytes+1-len(valid)) + "}"
			return c.AdmitTask(b, h.embedded(t, b))
		}, "task_limit_exceeded"},
		{"schema invalid", func(t *testing.T, h *harness, c *Client) (string, error) {
			b := `{"schema_version":"0.1.0"}`
			return c.AdmitTask(b, h.embedded(t, b))
		}, "task_schema_invalid"},
		{"expired", func(t *testing.T, h *harness, c *Client) (string, error) {
			b := synTask("synthetic-task", -time.Second)
			return c.AdmitTask(b, h.embedded(t, b))
		}, "task_expired"},
		{"unregistered test", func(t *testing.T, h *harness, c *Client) (string, error) {
			b := strings.Replace(valid, `"SYN-01"`, `"SYN-02"`, 1)
			return c.AdmitTask(b, h.embedded(t, b))
		}, "task_unsupported"},
		{"limit exceeded", func(t *testing.T, h *harness, c *Client) (string, error) {
			b := strings.Replace(valid, `"limits":{}`, `"limits":{"max_duration_ms":60001}`, 1)
			return c.AdmitTask(b, h.embedded(t, b))
		}, "task_limit_exceeded"},
		{"unauthorized issuer", func(t *testing.T, h *harness, c *Client) (string, error) {
			ref, err := h.deliveries.InstallEmbedded(context.Background(), "other-issuer", 2, []byte(valid))
			if err != nil {
				t.Fatal(err)
			}
			return c.AdmitTask(valid, ref)
		}, "task_issuer_not_authorized"},
		{"replay state unavailable", func(t *testing.T, h *harness, c *Client) (string, error) {
			h.replay.Inject(replay.FailLookup, true)
			return c.AdmitTask(valid, h.embedded(t, valid))
		}, "task_trust_unavailable"},
		{"commit not durable", func(t *testing.T, h *harness, c *Client) (string, error) {
			h.replay.Inject(replay.CrashDuringWrite, true)
			return c.AdmitTask(valid, h.embedded(t, valid))
		}, "task_trust_unavailable"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := admittingHarness(t)
			c := h.initialized(t)
			th, err := tc.run(t, h, c)
			if codeOf(err) != tc.want || th != "" {
				t.Fatalf("got %s, want %s", codeOf(err), tc.want)
			}
			if _, ok := errors.AsType[task.AdmissionError](err); !ok {
				t.Fatal("admission rejection is not a public admission class")
			}
			if taskHandles(c) != 0 || h.replay.Len() != 0 || len(c.rt.runs.entries) != 0 {
				t.Fatal("rejection left a handle, a replay record or a run")
			}
		})
	}
	t.Run("policy rollback", func(t *testing.T) {
		h := admittingHarness(t)
		c := h.initialized(t)
		rec, err := persistence.NewReplayRecord(persistence.ReplayKey{IssuerID: "embedded-issuer", TaskID: "earlier-task"}, [32]byte{1}, 3, task.Embedded, task.TimestampFromTime(epoch.Add(time.Hour)), epoch)
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := h.replay.Commit(context.Background(), rec, epoch); err != nil {
			t.Fatal(err)
		}
		if th, err := c.AdmitTask(valid, h.embedded(t, valid)); codeOf(err) != "policy_rollback" || th != "" || taskHandles(c) != 0 || h.replay.Len() != 1 {
			t.Fatal("a lower active generation must be policy_rollback with no state change, got", codeOf(err))
		}
	})
}

// Calls outside the initialized state fail closed with an API-local code.
func TestAdmitTaskLifecycleStates(t *testing.T) {
	valid := synTask("synthetic-task", time.Hour)
	var nilClient *Client
	if th, err := nilClient.AdmitTask(valid, "ref"); codeOf(err) != "not_initialized" || th != "" {
		t.Fatal("nil client must fail closed")
	}
	var zero Client
	if _, err := zero.AdmitTask(valid, "ref"); codeOf(err) != "not_initialized" {
		t.Fatal("zero-value client must fail closed")
	}
	h := admittingHarness(t)
	c := h.client(t)
	ref := h.embedded(t, valid)
	if _, err := c.AdmitTask(valid, ref); codeOf(err) != "not_initialized" {
		t.Fatal("uninitialized client admitted a task")
	}
	if err := c.Initialize(`{}`); err != nil {
		t.Fatal(err)
	}
	if codeOf(c.ClearMeasurementData()) != "clearing_incomplete" {
		t.Fatal("clearing without the capability must be incomplete")
	}
	if _, err := c.AdmitTask(valid, ref); codeOf(err) != "clearing_incomplete" {
		t.Fatal("admission must stay blocked after incomplete clearing, got", codeOf(err))
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := c.AdmitTask(valid, ref); codeOf(err) != "client_closed" {
		t.Fatal("closed client admitted a task")
	}

	h = admittingHarness(t)
	g := newGate()
	h.adapters.clearStorage = g.clear
	c = h.initialized(t)
	ref = h.embedded(t, valid)
	done := make(chan error)
	go func() { done <- c.ClearMeasurementData() }()
	<-g.entered
	if _, err := c.AdmitTask(valid, ref); codeOf(err) != "transition_in_progress" {
		t.Fatal("admission during a transition must be refused, got", codeOf(err))
	}
	close(g.release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if h.replay.Commits() != 0 {
		t.Fatal("a refused call reached admission")
	}
}

// Delivery references stop resolving after clearing and policy replacement,
// and handles do not survive a new generation. A redelivery to the new
// generation resolves to the same admitted task under a new handle.
func TestAdmitTaskAfterInvalidation(t *testing.T) {
	valid := synTask("synthetic-task", time.Hour)
	h := admittingHarness(t)
	h.adapters.clearStorage = func(context.Context) error { return nil }
	c := h.initialized(t)
	ref := h.embedded(t, valid)
	first, err := c.AdmitTask(valid, ref)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.ClearMeasurementData(); err != nil {
		t.Fatal(err)
	}
	if err := c.Initialize(`{}`); err != nil {
		t.Fatal(err)
	}
	if _, err := c.AdmitTask(valid, ref); codeOf(err) != "task_source_unauthenticated" {
		t.Fatal("delivery reference survived clearing, got", codeOf(err))
	}
	if _, err := c.rt.tasks.lookup(first); codeOf(err) != "handle_unavailable" {
		t.Fatal("task handle survived a new generation")
	}
	second, err := c.AdmitTask(valid, h.embedded(t, valid))
	if err != nil || second == first || h.replay.Commits() != 1 {
		t.Fatal("redelivery to a new generation must resolve to the retained acceptance under a new handle")
	}

	stale := h.embedded(t, valid)
	if err := h.deliveries.ReplacePolicy(3); err != nil {
		t.Fatal(err)
	}
	if _, err := c.AdmitTask(valid, stale); codeOf(err) != "task_source_unauthenticated" {
		t.Fatal("delivery reference survived policy replacement, got", codeOf(err))
	}
}

// A full task table drops only tasks whose validity has ended; it never evicts
// a valid task and otherwise returns capacity_exceeded.
func TestAdmitTaskHandleCapacity(t *testing.T) {
	h := admittingHarness(t)
	c := h.initialized(t)
	b := synTask("task-0", time.Hour)
	th, err := c.AdmitTask(b, h.embedded(t, b))
	if err != nil {
		t.Fatal(err)
	}
	live, _ := c.rt.tasks.lookup(th)
	for range taskHandleCapacity - 2 {
		if _, err := c.rt.tasks.add(live); err != nil {
			t.Fatal(err)
		}
	}
	b1 := synTask("task-1", 2*time.Hour)
	if _, err := c.AdmitTask(b1, h.embedded(t, b1)); err != nil || taskHandles(c) != taskHandleCapacity {
		t.Fatal("admission at the capacity limit rejected:", codeOf(err))
	}
	b2 := synTask("task-2", 2*time.Hour)
	if th2, err := c.AdmitTask(b2, h.embedded(t, b2)); codeOf(err) != "capacity_exceeded" || th2 != "" || taskHandles(c) != taskHandleCapacity {
		t.Fatal("admission beyond capacity must be capacity_exceeded without eviction, got", codeOf(err))
	}
	// Once task-0 and its copies have expired, their slots are reclaimed.
	h.clock.Advance(time.Hour)
	th2, err := c.AdmitTask(b2, h.embedded(t, b2))
	if err != nil || taskHandles(c) != 2 {
		t.Fatal("expired tasks were not reclaimed:", codeOf(err))
	}
	if _, err := c.rt.tasks.lookup(th); codeOf(err) != "handle_unavailable" {
		t.Fatal("expired task handle still resolves")
	}
	if again, _ := c.AdmitTask(b2, h.embedded(t, b2)); again != th2 {
		t.Fatal("redelivery after reclamation returned a different handle")
	}
}

// panicClock panics while armed; it models a misbehaving trusted adapter.
type panicClock struct {
	fakeClock
	armed *atomic.Bool
}

func (c panicClock) Now(ctx context.Context) (time.Time, error) {
	if c.armed.Load() {
		panic("CANARY-PANIC")
	}
	return c.fakeClock.Now(ctx)
}

func TestAdmitTaskRecoversPanics(t *testing.T) {
	h := admittingHarness(t)
	armed := &atomic.Bool{}
	h.adapters.clock = panicClock{fakeClock: h.clock, armed: armed}
	c := h.initialized(t)
	b := synTask("synthetic-task", time.Hour)
	ref := h.embedded(t, b)
	armed.Store(true)
	th, err := c.AdmitTask(b, ref)
	if codeOf(err) != "task_trust_unavailable" || th != "" || strings.Contains(fmt.Sprintf("%+v", err), "CANARY") {
		t.Fatal("a panic must become task_trust_unavailable with no handle, got", codeOf(err))
	}
	if c.active != 0 || stateOf(c) != initialized || taskHandles(c) != 0 {
		t.Fatal("a recovered panic must release its reservation and leave no state")
	}
	armed.Store(false)
	if _, err := c.AdmitTask(b, ref); err != nil {
		t.Fatal("client unusable after a recovered panic:", codeOf(err))
	}
}

func TestAdmitTaskErrorsCarryNoInput(t *testing.T) {
	const canary = "canary-7f3a"
	h := admittingHarness(t)
	c := h.initialized(t)
	b := synTask(canary, time.Hour)
	conflict := synTask(canary, 2*time.Hour)
	th, err := c.AdmitTask(b, h.embedded(t, b))
	if err != nil {
		t.Fatal(err)
	}
	var errs []error
	for _, call := range [][2]string{
		{b, canary + "-reference"},
		{`{"` + canary + `":true}`, h.embedded(t, `{"`+canary+`":true}`)},
		{conflict, h.embedded(t, conflict)},
	} {
		_, err := c.AdmitTask(call[0], call[1])
		errs = append(errs, err)
	}
	_ = c.Close()
	_, err = c.AdmitTask(b, canary)
	errs = append(errs, err)
	for _, err := range errs {
		if err == nil {
			t.Fatal("expected an error")
		}
		for _, f := range []string{"%v", "%+v", "%#v", "%s"} {
			out := fmt.Sprintf(f, err)
			if strings.Contains(out, canary) || strings.Contains(out, th) || strings.Contains(out, "reg-a") {
				t.Fatal("error formatting leaked input")
			}
		}
		_, admissionClass := errors.AsType[task.AdmissionError](err)
		_, apiLocal := errors.AsType[opError](err)
		if admissionClass == apiLocal || errors.Unwrap(err) != nil {
			t.Fatal("error is not exactly one fixed, unwrapped code:", err.Error())
		}
	}
}

// Concurrent admissions queue behind admission's own serialization, so this
// test uses the default lifecycle deadline: under the race detector the short
// harness deadline expires while queued, which fails closed as
// task_trust_unavailable rather than testing handle sharing.
func TestAdmitTaskConcurrency(t *testing.T) {
	h := admittingHarness(t)
	h.adapters.deadline = defaultLifecycleDeadline
	c := h.initialized(t)
	b := synTask("synthetic-task", time.Hour)
	ref := h.embedded(t, b)
	const n = 32
	handles := make(chan string, n)
	var wg sync.WaitGroup
	for range n {
		wg.Go(func() {
			th, err := c.AdmitTask(b, ref)
			if err != nil {
				t.Error("concurrent identical delivery rejected:", codeOf(err))
			}
			handles <- th
		})
	}
	wg.Wait()
	close(handles)
	first := <-handles
	for th := range handles {
		if th != first {
			t.Fatal("concurrent identical deliveries received different handles")
		}
	}
	if h.replay.Commits() != 1 || taskHandles(c) != 1 {
		t.Fatal("concurrent identical deliveries committed or published more than once")
	}

	distinct := make(chan string, 4)
	for i := range 4 {
		bi := synTask(fmt.Sprintf("task-%d", i), time.Hour)
		ri := h.embedded(t, bi)
		wg.Go(func() {
			th, err := c.AdmitTask(bi, ri)
			if err != nil {
				t.Error("concurrent distinct delivery rejected:", codeOf(err))
			}
			distinct <- th
		})
	}
	wg.Wait()
	close(distinct)
	seen := map[string]bool{first: true}
	for th := range distinct {
		if seen[th] {
			t.Fatal("distinct tasks share a handle")
		}
		seen[th] = true
	}

	// Admission racing close ends closed with no reservation left.
	results := make(chan string, n)
	for range n {
		wg.Go(func() {
			_, err := c.AdmitTask(b, ref)
			results <- codeOf(err)
		})
	}
	closeErr := c.Close()
	wg.Wait()
	close(results)
	for r := range results {
		switch r {
		case "ok", "client_closed", "transition_in_progress", "task_trust_unavailable", "task_source_unauthenticated":
		default:
			t.Fatal("unexpected result racing close:", r)
		}
	}
	for closeErr != nil {
		closeErr = c.Close()
	}
	if stateOf(c) != closed || c.active != 0 {
		t.Fatal("close did not complete after racing admissions")
	}
}

// Close cancels the generation's adapter context, so a transition does not
// wait on admission blocked in a trusted adapter.
func TestCloseCancelsAdmissionContext(t *testing.T) {
	c := admittingHarness(t).initialized(t)
	rt := c.rt
	if rt.ctx.Err() != nil {
		t.Fatal("fresh generation context already cancelled")
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if rt.ctx.Err() == nil {
		t.Fatal("close left the generation context live")
	}
}
