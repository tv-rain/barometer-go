package barometer

import (
	"context"
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/persistence"
	"github.com/tv-rain/barometer-go/internal/testkit/delivery"
	"github.com/tv-rain/barometer-go/internal/testkit/replay"
)

func (h *harness) embeddedAt(t *testing.T, generation uint64, taskJSON string) string {
	t.Helper()
	ref, err := h.deliveries.InstallEmbedded(context.Background(), "embedded-issuer", generation, []byte(taskJSON))
	if err != nil {
		t.Fatal("embedded install rejected")
	}
	return ref
}

func activePolicy(c *Client) task.LocalPolicy {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.bound == nil {
		return task.LocalPolicy{}
	}
	return c.bound.policy
}

// requireActive asserts that generation 2 is still the complete active policy:
// the runtime is unchanged, its handle still resolves and its tasks admit.
func requireActive(t *testing.T, h *harness, c *Client, rt *runtime, handle string) {
	t.Helper()
	if stateOf(c) != initialized || c.rt != rt || rt.ctx.Err() != nil || activePolicy(c).Generation() != 2 {
		t.Fatal("the previous complete policy must stay active")
	}
	if _, err := rt.tasks.lookup(handle); err != nil {
		t.Fatal("a handle of the active policy stopped resolving")
	}
	b := synTask(fmt.Sprintf("still-active-%d", h.replay.Commits()), time.Hour)
	if _, err := c.AdmitTask(b, h.embedded(t, b)); err != nil {
		t.Fatal("the previous policy no longer admits:", codeOf(err))
	}
}

// requireUninitialized asserts that a failed replacement left no policy: the
// client is uninitialized, and initializing it again admits nothing until a
// complete policy is bound, never a mix of the two generations.
func requireUninitialized(t *testing.T, h *harness, c *Client) {
	t.Helper()
	if stateOf(c) != uninitialized || c.rt != nil || activePolicy(c).Valid() {
		t.Fatal("a failed replacement must leave the client uninitialized with no policy")
	}
	b := synTask("after-failure", time.Hour)
	if _, err := c.AdmitTask(b, "unknown-reference"); codeOf(err) != "not_initialized" {
		t.Fatal("an uninitialized client must refuse admission, got", codeOf(err))
	}
	if err := c.Initialize(`{}`); err != nil {
		t.Fatal(err)
	}
	for _, g := range []uint64{2, 3} {
		if _, err := c.AdmitTask(b, h.embeddedAt(t, g, b)); codeOf(err) != "task_policy_unavailable" {
			t.Fatalf("generation %d delivery after a failed replacement: got %s", g, codeOf(err))
		}
	}
}

// blockingClock blocks its first armed call until the context ends.
type blockingClock struct {
	fakeClock
	armed   *atomic.Bool
	entered chan struct{}
}

func (c blockingClock) Now(ctx context.Context) (time.Time, error) {
	if c.armed.CompareAndSwap(true, false) {
		close(c.entered)
		<-ctx.Done()
		return time.Time{}, ctx.Err()
	}
	return c.fakeClock.Now(ctx)
}

// TestCcTrustLocalPolicyLifecycle is the CC-TRUST-LOCAL-POLICY-LIFECYCLE case:
// the adapter-installed policy is immutable for an initialized generation and
// is replaced only by quiescing and reinitializing with a strictly higher
// generation (REQ-TRUST-0029 to REQ-TRUST-0032).
func TestCcTrustLocalPolicyLifecycle(t *testing.T) {
	t.Run("no in-place policy update path", func(t *testing.T) {
		// Only replacePolicy writes the policy; only bind, close and
		// replacePolicy rebind adapters; nothing reassigns an admitter.
		allowed := map[string]map[string]bool{
			"policy":   {"replacePolicy": true},
			"bound":    {"bind": true, "close": true, "replacePolicy": true},
			"admitter": {},
		}
		files, err := filepath.Glob("*.go")
		if err != nil {
			t.Fatal(err)
		}
		fset := token.NewFileSet()
		for _, name := range files {
			if strings.HasSuffix(name, "_test.go") {
				continue
			}
			f, err := parser.ParseFile(fset, name, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			for _, d := range f.Decls {
				fn, ok := d.(*ast.FuncDecl)
				if !ok || fn.Body == nil {
					continue
				}
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					as, ok := n.(*ast.AssignStmt)
					if !ok {
						return true
					}
					for _, lhs := range as.Lhs {
						if sel, ok := lhs.(*ast.SelectorExpr); ok {
							if funcs, watched := allowed[sel.Sel.Name]; watched && !funcs[fn.Name.Name] {
								t.Errorf("%s assigns %s outside the replacement seam", fn.Name.Name, sel.Sel.Name)
							}
						}
					}
					return true
				})
			}
		}

		// Admission leaves the installed snapshot and generation unchanged.
		h := admittingHarness(t)
		c := h.initialized(t)
		before := activePolicy(c)
		b := synTask("synthetic-task", time.Hour)
		if _, err := c.AdmitTask(b, h.embedded(t, b)); err != nil {
			t.Fatal(err)
		}
		if after := activePolicy(c); after.Generation() != before.Generation() || c.generation != 1 {
			t.Fatal("admission changed the active policy")
		}
		if err := c.Initialize(`{}`); codeOf(err) != "already_initialized" {
			t.Fatal("initialization must not mutate an initialized client")
		}
	})

	t.Run("lifecycle states", func(t *testing.T) {
		var nilClient *Client
		next := synPolicy(t, 3, "reg-a", permit)
		if codeOf(nilClient.replacePolicy(next)) != "not_initialized" || codeOf(NewClient().replacePolicy(next)) != "not_initialized" {
			t.Fatal("replacement needs an initialized client")
		}
		h := admittingHarness(t)
		c := h.initialized(t)
		c.mu.Lock()
		c.state = transitioning
		c.mu.Unlock()
		if codeOf(c.replacePolicy(next)) != "transition_in_progress" {
			t.Fatal("replacement must not race another transition")
		}
		c.mu.Lock()
		c.state = initialized
		c.mu.Unlock()
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
		if codeOf(c.replacePolicy(next)) != "client_closed" {
			t.Fatal("a closed client must refuse replacement")
		}
	})

	t.Run("refused replacements keep the previous policy", func(t *testing.T) {
		cases := []struct {
			name   string
			policy func(t *testing.T) task.LocalPolicy
			setup  func(t *testing.T, h *harness)
		}{
			{"equal generation", func(t *testing.T) task.LocalPolicy { return synPolicy(t, 2, "reg-a", permit) }, nil},
			{"lower generation", func(t *testing.T) task.LocalPolicy { return synPolicy(t, 1, "reg-a", permit) }, nil},
			{"invalid policy", func(t *testing.T) task.LocalPolicy {
				// A generation-3 snapshot with an unresolved registration
				// cannot be constructed; only its invalid zero value exists.
				p, err := task.NewLocalPolicy(task.PolicyInput{Generation: 3, Permissions: synPermissions()})
				if err == nil || p.Valid() {
					t.Fatal("an incomplete policy must not construct")
				}
				return p
			}, nil},
			{"retained acceptances above the new generation", func(t *testing.T) task.LocalPolicy { return synPolicy(t, 3, "reg-a", permit) },
				func(t *testing.T, h *harness) {
					rec, err := persistence.NewReplayRecord(persistence.ReplayKey{IssuerID: "embedded-issuer", TaskID: "later-task"}, sha256.Sum256([]byte("synthetic")), 4, task.Embedded, task.TimestampFromTime(epoch.Add(time.Hour)), epoch)
					if err != nil {
						t.Fatal(err)
					}
					if _, _, err := h.replay.Commit(context.Background(), rec, epoch); err != nil {
						t.Fatal(err)
					}
				}},
			{"replay generations unavailable", func(t *testing.T) task.LocalPolicy { return synPolicy(t, 3, "reg-a", permit) },
				func(t *testing.T, h *harness) { h.replay.Inject(replay.FailHighest, true) }},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				h := admittingHarness(t)
				c := h.initialized(t)
				b := synTask("synthetic-task", time.Hour)
				handle, err := c.AdmitTask(b, h.embedded(t, b))
				if err != nil {
					t.Fatal(err)
				}
				unused := h.embedded(t, synTask("unused-task", time.Hour))
				if tc.setup != nil {
					tc.setup(t, h)
				}
				rt := c.rt
				if err := c.replacePolicy(tc.policy(t)); codeOf(err) != "policy_rejected" {
					t.Fatal("refused replacement returned", codeOf(err))
				}
				h.replay.Inject(replay.FailHighest, false)
				if tc.setup != nil && strings.HasPrefix(tc.name, "retained") {
					// The retained generation-4 acceptance already makes
					// generation 2 a rollback at admission.
					if _, err := c.AdmitTask(b, h.embedded(t, b)); codeOf(err) != "policy_rollback" {
						t.Fatal("generation 2 below a retained acceptance must be a rollback, got", codeOf(err))
					}
					return
				}
				requireActive(t, h, c, rt, handle)
				if _, err := c.AdmitTask(synTask("unused-task", time.Hour), unused); err != nil {
					t.Fatal("a refused replacement invalidated delivery references:", codeOf(err))
				}
			})
		}
	})

	t.Run("a panic during validation keeps the previous policy", func(t *testing.T) {
		h := admittingHarness(t)
		armed := &atomic.Bool{}
		h.adapters.clock = panicClock{fakeClock: h.clock, armed: armed}
		c := h.initialized(t)
		b := synTask("synthetic-task", time.Hour)
		handle, err := c.AdmitTask(b, h.embedded(t, b))
		if err != nil {
			t.Fatal(err)
		}
		rt := c.rt
		armed.Store(true)
		err = c.replacePolicy(synPolicy(t, 3, "reg-a", permit))
		armed.Store(false)
		if codeOf(err) != "policy_rejected" || strings.Contains(fmt.Sprintf("%+v", err), "CANARY") {
			t.Fatal("a panicking adapter must refuse the replacement, got", codeOf(err))
		}
		requireActive(t, h, c, rt, handle)
	})

	t.Run("failure while cancelling leaves the client uninitialized", func(t *testing.T) {
		h := admittingHarness(t)
		c := h.initialized(t)
		cancel := c.rt.cancel
		c.rt.cancel = func() { cancel(); panic("CANARY-PANIC") }
		if err := c.replacePolicy(synPolicy(t, 3, "reg-a", permit)); codeOf(err) != "not_initialized" {
			t.Fatal("a failed cancellation returned", codeOf(err))
		}
		requireUninitialized(t, h, c)
	})

	t.Run("failure while invalidating delivery records leaves the client uninitialized", func(t *testing.T) {
		h := admittingHarness(t)
		c := h.initialized(t)
		stale := h.embedded(t, synTask("stale-task", time.Hour))
		// The delivery store refuses a generation it has already passed.
		if err := h.deliveries.ReplacePolicy(3); err != nil {
			t.Fatal(err)
		}
		other := h.embeddedAt(t, 3, synTask("other-task", time.Hour))
		if err := c.replacePolicy(synPolicy(t, 3, "reg-a", permit)); codeOf(err) != "not_initialized" {
			t.Fatal("a failed delivery invalidation returned", codeOf(err))
		}
		if _, err := h.deliveries.Resolve(context.Background(), other, []byte(synTask("other-task", time.Hour))); err == nil {
			t.Fatal("delivery records survived a failed replacement")
		}
		requireUninitialized(t, h, c)
		if _, err := c.AdmitTask(synTask("stale-task", time.Hour), stale); codeOf(err) != "task_source_unauthenticated" {
			t.Fatal("a stale delivery reference resolved, got", codeOf(err))
		}
	})

	t.Run("work that does not stop keeps the client blocked", func(t *testing.T) {
		h := admittingHarness(t)
		h.adapters.deadline = 20 * time.Millisecond
		c := h.initialized(t)
		if _, err := c.reserve(); err != nil {
			t.Fatal(err)
		}
		if err := c.replacePolicy(synPolicy(t, 3, "reg-a", permit)); codeOf(err) != "cleanup_incomplete" {
			t.Fatal("an expired drain returned", codeOf(err))
		}
		if stateOf(c) != blocked || c.rt != nil || activePolicy(c).Valid() {
			t.Fatal("an expired drain must keep the client blocked with no policy")
		}
		b := synTask("blocked-task", time.Hour)
		if _, err := c.AdmitTask(b, h.embeddedAt(t, 3, b)); codeOf(err) != "cleanup_incomplete" {
			t.Fatal("a blocked client admitted work, got", codeOf(err))
		}
		c.release()
		if err := c.Close(); err != nil || stateOf(c) != closed {
			t.Fatal("close must recover a blocked replacement")
		}
	})

	t.Run("in-flight admission is cancelled", func(t *testing.T) {
		h := admittingHarness(t)
		armed := &atomic.Bool{}
		clock := blockingClock{fakeClock: h.clock, armed: armed, entered: make(chan struct{})}
		h.adapters.clock = clock
		h.adapters.deadline = 5 * time.Second
		c := h.initialized(t)
		b := synTask("in-flight-task", time.Hour)
		ref := h.embedded(t, b)
		armed.Store(true)
		result := make(chan error, 1)
		go func() {
			_, err := c.AdmitTask(b, ref)
			result <- err
		}()
		<-clock.entered
		if err := c.replacePolicy(synPolicy(t, 3, "reg-a", permit)); err != nil {
			t.Fatal("replacement did not cancel in-flight admission:", codeOf(err))
		}
		if err := <-result; codeOf(err) != "task_trust_unavailable" {
			t.Fatal("cancelled admission returned", codeOf(err))
		}
		if h.replay.Commits() != 0 || taskHandles(c) != 0 {
			t.Fatal("cancelled admission left state behind")
		}
	})

	t.Run("successful replacement", func(t *testing.T) {
		h := admittingHarness(t)
		h.adapters.clearStorage = func(context.Context) error { return nil }
		c := h.initialized(t)
		b := synTask("synthetic-task", time.Hour)
		handle, err := c.AdmitTask(b, h.embedded(t, b))
		if err != nil {
			t.Fatal(err)
		}
		unused := h.embedded(t, synTask("queued-task", time.Hour))
		old := c.rt

		// Generation 3 removes the task's registration.
		if err := c.replacePolicy(synPolicy(t, 3, "reg-b", permit)); err != nil {
			t.Fatal("valid replacement rejected:", codeOf(err))
		}
		if stateOf(c) != initialized || c.rt == old || c.generation != 2 || c.rt.generation != 2 || activePolicy(c).Generation() != 3 || c.policyFloor != 3 {
			t.Fatal("replacement must reinitialize a new generation under the new policy")
		}
		if old.ctx.Err() == nil || c.rt.ctx.Err() != nil {
			t.Fatal("replacement must cancel the previous generation only")
		}
		for _, table := range []*taskTable{old.tasks, c.rt.tasks} {
			if _, err := table.lookup(handle); codeOf(err) != "handle_unavailable" {
				t.Fatal("a task handle survived replacement")
			}
		}
		if _, err := c.AdmitTask(synTask("queued-task", time.Hour), unused); codeOf(err) != "task_source_unauthenticated" {
			t.Fatal("a delivery record survived replacement, got", codeOf(err))
		}
		if _, err := c.AdmitTask(b, h.embeddedAt(t, 3, b)); codeOf(err) != "target_not_registered" {
			t.Fatal("a task no longer permitted must be rejected under the new policy, got", codeOf(err))
		}
		if _, err := c.AdmitTask(b, h.embeddedAt(t, 2, b)); codeOf(err) != "policy_rollback" {
			t.Fatal("a generation-2 delivery after replacement must be a rollback, got", codeOf(err))
		}
		if h.replay.Commits() != 1 {
			t.Fatal("rejections under the new policy wrote replay state")
		}

		// Generation 4 permits the task again: a redelivery resolves to the
		// retained acceptance under a new handle.
		if err := c.replacePolicy(synPolicy(t, 4, "reg-a", permit)); err != nil {
			t.Fatal(err)
		}
		again, err := c.AdmitTask(b, h.embeddedAt(t, 4, b))
		if err != nil || again == handle || h.replay.Commits() != 1 {
			t.Fatal("a still-permitted task must resolve to its retained acceptance, got", codeOf(err))
		}
		if a, err := c.rt.tasks.lookup(again); err != nil || !a.Redelivery() {
			t.Fatal("the retained acceptance was not reused")
		}

		// Reinitializing this instance below the highest activated generation
		// is a rollback, even after a privacy clear.
		if err := c.ClearMeasurementData(); err != nil {
			t.Fatal(err)
		}
		for _, g := range []uint64{2, 3} {
			a := h.adapters
			a.policy = synPolicy(t, g, "reg-a", permit)
			if err := c.bind(a); err != nil {
				t.Fatal(err)
			}
			if err := c.Initialize(`{}`); codeOf(err) != "not_initialized" || stateOf(c) != uninitialized {
				t.Fatalf("initializing generation %d after generation 4 must fail closed", g)
			}
		}
		a := h.adapters
		a.policy = synPolicy(t, 4, "reg-a", permit)
		if err := c.bind(a); err != nil {
			t.Fatal(err)
		}
		if err := c.Initialize(`{}`); err != nil {
			t.Fatal("reinitializing the active generation was refused:", codeOf(err))
		}
		next := synTask("next-task", time.Hour)
		if _, err := c.AdmitTask(next, h.embeddedAt(t, 4, next)); err != nil {
			t.Fatal(codeOf(err))
		}
	})

	t.Run("task content cannot change policy", func(t *testing.T) {
		h := admittingHarness(t)
		labelled := func(p *task.Permissions) { p.LabelNamespaces = []string{"campaign"} }
		h.adapters.policy = synPolicy(t, 2, "reg-a", labelled)
		c := h.initialized(t)
		policyShaped := `{\"policy_generation\":9,\"targets\":[{\"id\":\"reg-z\"}],\"issuers\":[\"attacker\"]}`
		accepted := strings.Replace(synTask("labelled-task", time.Hour), `"input":{}`,
			`"labels":{"campaign":"`+policyShaped+`","campaign.policy":"generation 9"},"input":{}`, 1)
		if _, err := c.AdmitTask(accepted, h.embedded(t, accepted)); err != nil {
			t.Fatal("a task with policy-shaped labels was rejected:", codeOf(err))
		}
		for name, field := range map[string]string{
			"policy member":       `"policy":{"generation":9,"targets":[]},`,
			"generation member":   `"policy_generation":9,`,
			"registration member": `"target_registrations":[{"id":"reg-z"}],`,
		} {
			b := strings.Replace(synTask("injected-task", time.Hour), `"input":{}`, field+`"input":{}`, 1)
			if _, err := c.AdmitTask(b, h.embedded(t, b)); codeOf(err) != "task_schema_invalid" {
				t.Fatalf("%s: got %s", name, codeOf(err))
			}
		}
		p := activePolicy(c)
		if p.Generation() != 2 || c.generation != 1 || c.policyFloor != 2 {
			t.Fatal("task content changed the active policy")
		}
		if _, ok := p.Target("reg-z"); ok {
			t.Fatal("task content added a registration")
		}
		if _, ok := p.Issuer("attacker"); ok {
			t.Fatal("task content added an issuer")
		}
	})
}

// Concurrent replacements and admissions never activate a lower generation,
// never run two transitions at once and never panic.
func TestPolicyReplacementConcurrency(t *testing.T) {
	h := admittingHarness(t)
	h.adapters.deadline = 5 * time.Second
	c := h.initialized(t)
	policies := make([]task.LocalPolicy, 0, 8)
	for g := uint64(3); g < 11; g++ {
		policies = append(policies, synPolicy(t, g, "reg-a", permit))
	}
	var wg sync.WaitGroup
	var replaced atomic.Int32
	for _, p := range policies {
		wg.Add(1)
		go func() {
			defer wg.Done()
			switch err := c.replacePolicy(p); codeOf(err) {
			case "ok":
				replaced.Add(1)
			case "transition_in_progress", "policy_rejected":
			default:
				t.Error("unexpected replacement result:", codeOf(err))
			}
		}()
	}
	for i := range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b := synTask(fmt.Sprintf("concurrent-%d", i), time.Hour)
			ref, err := h.deliveries.InstallEmbedded(context.Background(), "embedded-issuer", 2, []byte(b))
			if err != nil {
				return // the store moved to a later generation meanwhile
			}
			_, err = c.AdmitTask(b, ref)
			switch r := codeOf(err); r {
			case "ok", "transition_in_progress", "task_source_unauthenticated", "task_trust_unavailable", "policy_rollback":
			default:
				t.Error("unexpected admission result:", r)
			}
		}()
	}
	wg.Wait()
	if replaced.Load() < 1 || stateOf(c) != initialized || c.active != 0 {
		t.Fatal("concurrent replacement must end initialized under one winning generation")
	}
	if g := activePolicy(c).Generation(); g < 3 || g != c.policyFloor {
		t.Fatal("the active generation is not the highest activated one")
	}
}

// Trusted keys and issuers are rotated or removed only by a replacement policy
// with a higher generation. Replacement invalidates every handle and delivery
// reference; a redelivered package is verified again under the new policy, so
// a task signed with a removed key, or by a removed issuer, is no longer
// admitted, while a task signed with a retained key resolves to its retained
// acceptance. A task admitted while its key was valid is not re-verified when
// the key later expires: its handle stays until its own expiry, but a new
// presentation of its package is verified at that time and rejected, also
// after a replacement that keeps the key, while the same task bytes re-signed
// with a valid key resolve to the retained acceptance (REQ-TRUST-0055).
// requirement-ref: REQ-TRUST-0054
// requirement-ref: REQ-TRUST-0055
// requirement-ref: REQ-TRUST-0030
// requirement-ref: REQ-TRUST-0051
func TestCcTrustKeyRotationByPolicyReplacement(t *testing.T) {
	rotated := delivery.NewSigner("remote-key-2", "remote-issuer", 3)
	withKeys := func(t *testing.T, generation uint64, issuers []string, keys ...task.TrustedKey) task.LocalPolicy {
		t.Helper()
		in := synPolicyInput(generation, "reg-a", permit)
		var kept []task.IssuerPermissions
		for _, i := range in.Issuers {
			if slices.Contains(issuers, i.IssuerID) {
				kept = append(kept, i)
			}
		}
		in.Issuers, in.Keys = kept, keys
		p, err := task.NewLocalPolicy(in)
		if err != nil {
			t.Fatal("synthetic policy rejected")
		}
		return p
	}
	both := []string{"embedded-issuer", "remote-issuer"}
	present := func(t *testing.T, h *harness, c *Client, generation uint64, s delivery.Signer, taskJSON string) (string, error) {
		t.Helper()
		pkg := string(s.Sign(h.clock.Clock.Now(), []byte(taskJSON)).Encode())
		ref, err := h.deliveries.InstallRemote(context.Background(), generation, []byte(pkg))
		if err != nil {
			t.Fatal("remote install rejected")
		}
		return c.AdmitTask(pkg, ref)
	}

	t.Run("removing a key and then its issuer", func(t *testing.T) {
		h := admittingHarness(t)
		h.adapters.policy = withKeys(t, 2, both, remoteSigner().Key(keyNotBefore, keyNotAfter), rotated.Key(keyNotBefore, keyNotAfter))
		c := h.initialized(t)
		old, retained := synTask("old-key-task", time.Hour), synTask("new-key-task", time.Hour)
		oldHandle, err := present(t, h, c, 2, remoteSigner(), old)
		if err != nil {
			t.Fatal(codeOf(err))
		}
		newHandle, err := present(t, h, c, 2, rotated, retained)
		if err != nil {
			t.Fatal(codeOf(err))
		}
		queued := string(remoteSigner().Sign(h.clock.Clock.Now(), []byte(synTask("queued-task", time.Hour))).Encode())
		queuedRef, _ := h.deliveries.InstallRemote(context.Background(), 2, []byte(queued))

		// Generation 3 rotates: remote-key-1 is removed, remote-key-2 kept.
		if err := c.replacePolicy(withKeys(t, 3, both, rotated.Key(keyNotBefore, keyNotAfter))); err != nil {
			t.Fatal("valid replacement rejected:", codeOf(err))
		}
		for _, handle := range []string{oldHandle, newHandle} {
			if _, err := c.rt.tasks.lookup(handle); codeOf(err) != "handle_unavailable" {
				t.Fatal("a task handle survived replacement")
			}
		}
		if _, err := c.AdmitTask(queued, queuedRef); codeOf(err) != "task_source_unauthenticated" {
			t.Fatal("a queued delivery survived replacement, got", codeOf(err))
		}
		if _, err := present(t, h, c, 3, remoteSigner(), old); codeOf(err) != "task_source_unauthenticated" {
			t.Fatal("a task signed with a removed key was admitted, got", codeOf(err))
		}
		if _, err := present(t, h, c, 3, remoteSigner(), synTask("fresh-old-key-task", time.Hour)); codeOf(err) != "task_source_unauthenticated" {
			t.Fatal("a new task signed with a removed key was admitted, got", codeOf(err))
		}
		again, err := present(t, h, c, 3, rotated, retained)
		if err != nil || again == newHandle {
			t.Fatal("a task signed with a retained key must be admitted again, got", codeOf(err))
		}
		if a, err := c.rt.tasks.lookup(again); err != nil || !a.Redelivery() || h.replay.Commits() != 2 {
			t.Fatal("the retained acceptance was not reused")
		}

		// Generation 4 removes the remote issuer and so all of its keys.
		if err := c.replacePolicy(withKeys(t, 4, []string{"embedded-issuer"}, embeddedSigner().Key(keyNotBefore, keyNotAfter))); err != nil {
			t.Fatal("valid replacement rejected:", codeOf(err))
		}
		if _, err := c.rt.tasks.lookup(again); codeOf(err) != "handle_unavailable" {
			t.Fatal("a task handle of a removed issuer survived replacement")
		}
		if _, err := present(t, h, c, 4, rotated, retained); codeOf(err) != "task_source_unauthenticated" {
			t.Fatal("a task of a removed issuer was admitted, got", codeOf(err))
		}
		if h.replay.Commits() != 2 {
			t.Fatal("rejections under a replacement policy wrote replay state")
		}
	})

	t.Run("a policy whose key names a removed issuer is rejected", func(t *testing.T) {
		h := admittingHarness(t)
		c := h.initialized(t)
		in := synPolicyInput(3, "reg-a", permit)
		in.Issuers = in.Issuers[:1]
		if _, err := task.NewLocalPolicy(in); err == nil {
			t.Fatal("a key for an issuer the policy does not authorize was accepted")
		}
		if activePolicy(c).Generation() != 2 {
			t.Fatal("the active policy changed")
		}
	})

	t.Run("key expiry after admission", func(t *testing.T) {
		h := admittingHarness(t)
		h.adapters.policy = withKeys(t, 2, both, remoteSigner().Key(keyNotBefore, epoch.Add(time.Hour)))
		c := h.initialized(t)
		b := synTask("expiring-key-task", 3*time.Hour)
		pkg := string(remoteSigner().Sign(epoch, []byte(b)).Encode())
		ref, _ := h.deliveries.InstallRemote(context.Background(), 2, []byte(pkg))
		handle, err := c.AdmitTask(pkg, ref)
		if err != nil {
			t.Fatal(codeOf(err))
		}
		if err := h.clock.Advance(2 * time.Hour); err != nil {
			t.Fatal(err)
		}
		if a, err := c.rt.tasks.lookup(handle); err != nil || !a.ExpiresAt().Equal(epoch.Add(3*time.Hour)) {
			t.Fatal("an admitted task was re-verified when its key expired")
		}
		ref, _ = h.deliveries.InstallRemote(context.Background(), 2, []byte(pkg))
		if _, err := c.AdmitTask(pkg, ref); codeOf(err) != "task_source_unauthenticated" {
			t.Fatal("a package presented after its key expired was accepted, got", codeOf(err))
		}
		if a, err := c.rt.tasks.lookup(handle); err != nil || !a.ExpiresAt().Equal(epoch.Add(3*time.Hour)) {
			t.Fatal("a rejected presentation invalidated the held acceptance")
		}
	})

	// REQ-TRUST-0055: after a replacement the instance no longer holds the
	// task, so its redelivery is verified at its own acceptance time even
	// though the new policy keeps the key and a replay record exists.
	t.Run("redelivery after a replacement that keeps an expired key", func(t *testing.T) {
		h := admittingHarness(t)
		expiring := remoteSigner().Key(keyNotBefore, epoch.Add(time.Hour))
		h.adapters.policy = withKeys(t, 2, both, expiring)
		c := h.initialized(t)
		b := synTask("expired-key-redelivery", 3*time.Hour)
		pkg := string(remoteSigner().Sign(epoch, []byte(b)).Encode())
		ref, _ := h.deliveries.InstallRemote(context.Background(), 2, []byte(pkg))
		handle, err := c.AdmitTask(pkg, ref)
		if err != nil {
			t.Fatal(codeOf(err))
		}
		if err := h.clock.Advance(2 * time.Hour); err != nil {
			t.Fatal(err)
		}
		if err := c.replacePolicy(withKeys(t, 3, both, expiring)); err != nil {
			t.Fatal("valid replacement rejected:", codeOf(err))
		}
		if _, err := c.rt.tasks.lookup(handle); codeOf(err) != "handle_unavailable" {
			t.Fatal("a task handle survived replacement")
		}
		ref, _ = h.deliveries.InstallRemote(context.Background(), 3, []byte(pkg))
		if _, err := c.AdmitTask(pkg, ref); codeOf(err) != "task_source_unauthenticated" {
			t.Fatal("a redelivery signed with an expired key was admitted, got", codeOf(err))
		}
		if h.replay.Commits() != 1 {
			t.Fatal("a rejected redelivery wrote replay state")
		}
	})

	// REQ-TRUST-0055 with REQ-TRUST-0024: the replay identity is the task
	// bytes, so a package re-signed with a currently valid key is an exact
	// redelivery with the original expiry, within the generation and after a
	// replacement.
	t.Run("a package re-signed with a valid key after key expiry", func(t *testing.T) {
		h := admittingHarness(t)
		h.adapters.policy = withKeys(t, 2, both, remoteSigner().Key(keyNotBefore, epoch.Add(time.Hour)), rotated.Key(keyNotBefore, keyNotAfter))
		c := h.initialized(t)
		b := synTask("re-signed-task", 3*time.Hour)
		pkg := string(remoteSigner().Sign(epoch, []byte(b)).Encode())
		ref, _ := h.deliveries.InstallRemote(context.Background(), 2, []byte(pkg))
		handle, err := c.AdmitTask(pkg, ref)
		if err != nil {
			t.Fatal(codeOf(err))
		}
		if err := h.clock.Advance(2 * time.Hour); err != nil {
			t.Fatal(err)
		}
		again, err := present(t, h, c, 2, rotated, b)
		if err != nil || again != handle {
			t.Fatal("a re-signed package did not resolve to the held acceptance, got", codeOf(err))
		}
		if err := c.replacePolicy(withKeys(t, 3, both, remoteSigner().Key(keyNotBefore, epoch.Add(time.Hour)), rotated.Key(keyNotBefore, keyNotAfter))); err != nil {
			t.Fatal("valid replacement rejected:", codeOf(err))
		}
		ref, _ = h.deliveries.InstallRemote(context.Background(), 3, []byte(pkg))
		if _, err := c.AdmitTask(pkg, ref); codeOf(err) != "task_source_unauthenticated" {
			t.Fatal("the original package was admitted after its key expired, got", codeOf(err))
		}
		after, err := present(t, h, c, 3, rotated, b)
		if err != nil {
			t.Fatal("a re-signed package was rejected after replacement:", codeOf(err))
		}
		a, err := c.rt.tasks.lookup(after)
		if err != nil || !a.Redelivery() || !a.ExpiresAt().Equal(epoch.Add(3*time.Hour)) || h.replay.Commits() != 1 {
			t.Fatal("the re-signed package did not resolve to the retained acceptance with its original expiry")
		}
	})
}
