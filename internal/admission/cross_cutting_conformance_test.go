package admission_test

import (
	"bytes"
	"context"
	"crypto/sha256"
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
	"github.com/tv-rain/barometer-go/internal/testkit/replay"
)

// baseNotAfter is the base task's expiry.
var baseNotAfter = time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)

func (h *harness) admitEmbedded(t *testing.T, b []byte) (admission.Admitted, error) {
	t.Helper()
	ref, err := h.store.InstallEmbedded(context.Background(), "embedded-issuer", 2, b)
	if err != nil {
		t.Fatal("embedded install rejected")
	}
	return h.admitter.Admit(context.Background(), ref, b)
}

func (h *harness) admitRemote(t *testing.T, b []byte) (admission.Admitted, error) {
	t.Helper()
	v := verifier()
	v.IssuedAt = task.TimestampFromTime(h.clock.Clock.Now())
	ref, err := h.store.InstallRemote(context.Background(), v, v.Seal(b))
	if err != nil {
		t.Fatal("remote install rejected")
	}
	return h.admitter.Admit(context.Background(), ref, b)
}

// restart models a new client instance after a process restart: a fresh
// delivery store and admitter over the surviving replay store, with the given
// active policy generation.
func (h *harness) restart(t *testing.T, generation uint64) {
	t.Helper()
	h.replay.Restart()
	var err error
	if h.store, err = admission.NewDeliveryStore(h.clock, admission.DeliveryLimits{Capacity: 256, Lifetime: lifetime}, generation); err != nil {
		t.Fatal("store rejected")
	}
	in := permissive()
	in.Generation = generation
	reg, err := registry.New([]registry.Entry{synEntry(h, task.MediumRisk)})
	if err != nil {
		t.Fatal("registry rejected")
	}
	if h.admitter, err = admission.NewAdmitter(h.store, h.clock, reg, mustPolicy(t, in), h.replay); err != nil {
		t.Fatal("admitter rejected")
	}
}

func (h *harness) admitAt(t *testing.T, generation uint64, b []byte) (admission.Admitted, error) {
	t.Helper()
	ref, err := h.store.InstallEmbedded(context.Background(), "embedded-issuer", generation, b)
	if err != nil {
		t.Fatal("embedded install rejected")
	}
	return h.admitter.Admit(context.Background(), ref, b)
}

// wantRecord checks that the record holds exactly the REQ-TRUST-0022 values for
// these exact bytes.
func wantRecord(t *testing.T, name string, r persistence.ReplayRecord, issuer string, origin task.Origin, b []byte) {
	t.Helper()
	want := persistence.ReplayRecord{Key: persistence.ReplayKey{IssuerID: issuer, TaskID: "synthetic-task"}, Digest: sha256.Sum256(b),
		PolicyGeneration: 2, Origin: origin, ExpiresAt: baseNotAfter, RetainUntil: baseNotAfter.Add(86400 * time.Second)}
	if r != want {
		t.Errorf("%s: replay record does not hold exactly the accepted key, digest, generation and classes", name)
	}
}

// requirement-ref: REQ-TRUST-0021
// requirement-ref: REQ-TRUST-0022
// requirement-ref: REQ-TRUST-0023
// requirement-ref: REQ-TRUST-0024
// requirement-ref: REQ-TRUST-0038
// requirement-ref: REQ-LIMIT-0023
func TestCcTrustReplayDuplicateAndRollbackHandling(t *testing.T) {
	b := []byte(baseTask)

	t.Run("new key continues admission and commits before acceptance", func(t *testing.T) {
		for _, origin := range []task.Origin{task.Embedded, task.Remote} {
			h := newHarness(t)
			admit, issuer := h.admitEmbedded, "embedded-issuer"
			if origin == task.Remote {
				admit, issuer = h.admitRemote, "remote-issuer"
			}
			a, err := admit(t, b)
			if err != nil || a.Redelivery() || h.replay.Commits() != 1 || h.replay.Len() != 1 {
				t.Fatalf("origin %d: new task not accepted with one durable record", origin)
			}
			wantRecord(t, "accepted", admission.AdmittedRecord(a), issuer, origin, b)
			if admission.AdmittedCandidate(a).ID != "synthetic-task" {
				t.Error("admitted task does not carry the validated candidate")
			}
			if fmt.Sprintf("%v %+v %#v", a, a, a) != "admission.Admitted admission.Admitted admission.Admitted" {
				t.Error("admitted formatting exposes values")
			}
		}
	})

	t.Run("same key and digest is idempotent without extending expiry", func(t *testing.T) {
		h := newHarness(t)
		first, err := h.admitEmbedded(t, b)
		if err != nil {
			t.Fatal("first delivery rejected")
		}
		// Cumulative advances stay before not_after, 24 hours after epoch.
		for _, advance := range []time.Duration{0, time.Hour, 22 * time.Hour} {
			if err := h.clock.Advance(advance); err != nil {
				t.Fatal("clock")
			}
			again, err := h.admitEmbedded(t, b)
			if err != nil || !again.Redelivery() || admission.AdmittedRecord(again) != admission.AdmittedRecord(first) {
				t.Errorf("redelivery after %v not resolved to the same admitted task", advance)
			}
		}
		if h.replay.Commits() != 1 || h.replay.Len() != 1 {
			t.Error("redelivery wrote replay state")
		}
	})

	t.Run("same key and different digest is task_id_conflict", func(t *testing.T) {
		h := newHarness(t)
		if _, err := h.admitEmbedded(t, b); err != nil {
			t.Fatal("first delivery rejected")
		}
		before := h.replay.Dump()
		for name, other := range map[string][]byte{
			"trailing whitespace":   append(append([]byte(nil), b...), ' '),
			"one changed byte":      bytes.Replace(b, []byte(`"example"`), []byte(`"exampla"`), 1),
			"different input":       taskWith(map[string]any{"input.subject": "other"}),
			"reordered same fields": taskWith(nil),
		} {
			if bytes.Equal(other, b) {
				t.Fatalf("%s: fixture equals the accepted bytes", name)
			}
			_, err := h.admitEmbedded(t, other)
			expect(t, name, err, task.IDConflict)
		}
		if h.replay.Commits() != 1 || !bytes.Equal(h.replay.Dump(), before) {
			t.Error("conflicting delivery changed replay state")
		}
		// The same task_id from another issuer is a different replay key.
		if _, err := h.admitRemote(t, b); err != nil {
			t.Error("different issuer treated as the same replay key")
		}
	})

	t.Run("lower policy generation is policy_rollback", func(t *testing.T) {
		h := newHarness(t)
		if _, err := h.admitEmbedded(t, b); err != nil {
			t.Fatal("generation 2 task rejected")
		}
		// A source generation below the active one (step 7).
		other := taskWith(map[string]any{"task_id": "other-task"})
		_, err := h.admitAt(t, 1, other)
		expect(t, "source generation 1 under active 2", err, task.PolicyRollback)
		// After a restart with an earlier active policy, the retained accepted
		// generation detects the rollback for new and redelivered tasks.
		h.restart(t, 1)
		for name, task_ := range map[string][]byte{"new": other, "redelivered": b} {
			_, err := h.admitAt(t, 1, task_)
			expect(t, "rolled-back policy, "+name, err, task.PolicyRollback)
		}
		// The same and a higher active generation still accept.
		h.restart(t, 2)
		if a, err := h.admitAt(t, 2, b); err != nil || !a.Redelivery() {
			t.Error("equal generation after restart rejected")
		}
		h.restart(t, 3)
		if a, err := h.admitAt(t, 3, other); err != nil || a.Redelivery() || admission.AdmittedRecord(a).PolicyGeneration != 3 {
			t.Error("higher generation after restart rejected")
		}
		if h.replay.Commits() != 2 {
			t.Error("rejected task wrote replay state")
		}
	})

	t.Run("expired task is task_expired and never reactivated", func(t *testing.T) {
		h := newHarness(t)
		if _, err := h.admitEmbedded(t, b); err != nil {
			t.Fatal("task rejected")
		}
		// From not_after onward the retained record must not reactivate it, and
		// neither must the record's purge at the end of retention.
		for _, at := range []time.Time{baseNotAfter.Add(-time.Second), baseNotAfter, baseNotAfter.Add(86400*time.Second - time.Second), baseNotAfter.Add(86400 * time.Second), baseNotAfter.Add(691200 * time.Second)} {
			h.clock.Clock = newClock().Clock
			if err := h.clock.Advance(at.Sub(epoch)); err != nil {
				t.Fatal("clock")
			}
			a, err := h.admitEmbedded(t, b)
			if at.Before(baseNotAfter) {
				if err != nil || !a.Redelivery() {
					t.Errorf("redelivery at %v before expiry rejected", at)
				}
				continue
			}
			expect(t, fmt.Sprintf("redelivery at %v", at), err, task.Expired)
		}
		if h.replay.Commits() != 1 {
			t.Error("expired redelivery wrote replay state")
		}
	})

	t.Run("unavailable or invalid replay state fails closed for every origin", func(t *testing.T) {
		faults := map[string]func(h *harness){
			"lookup unavailable":     func(h *harness) { h.replay.Inject(replay.FailLookup, true) },
			"generation unavailable": func(h *harness) { h.replay.Inject(replay.FailHighest, true) },
			"write fails":            func(h *harness) { h.replay.Inject(replay.FailBeforeWrite, true) },
			"crash during write":     func(h *harness) { h.replay.Inject(replay.CrashDuringWrite, true) },
			"uncertain write":        func(h *harness) { h.replay.Inject(replay.FailAfterWrite, true) },
			"corrupt record": func(h *harness) {
				h.replay.Corrupt(persistence.ReplayKey{IssuerID: "embedded-issuer", TaskID: "synthetic-task"}, []byte(`{"IssuerID":`))
				h.replay.Corrupt(persistence.ReplayKey{IssuerID: "remote-issuer", TaskID: "synthetic-task"}, []byte(`{"IssuerID":`))
			},
			"invalid record": func(h *harness) {
				h.replay.Corrupt(persistence.ReplayKey{IssuerID: "other", TaskID: "x"}, []byte(`{"IssuerID":"other","TaskID":"x","Digest":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","PolicyGeneration":1,"Origin":1,"ExpiresAt":"2026-10-20T12:00:00Z","RetainUntil":"2026-10-21T12:00:00Z"}`))
			},
			"capacity exhausted": func(h *harness) {
				s, _ := replay.New(1)
				h.replay = s
				h.restart(t, 2)
				if _, err := h.admitAt(t, 2, taskWith(map[string]any{"task_id": "filler"})); err != nil {
					t.Fatal("filler rejected")
				}
			},
		}
		for name, inject := range faults {
			for _, origin := range []task.Origin{task.Embedded, task.Remote} {
				h := newHarness(t)
				inject(h)
				admit := h.admitEmbedded
				if origin == task.Remote {
					admit = h.admitRemote
				}
				a, err := admit(t, b)
				expect(t, fmt.Sprintf("%s, origin %d", name, origin), err, task.TrustUnavailable)
				if admission.AdmittedRecord(a) != (persistence.ReplayRecord{}) || a.Redelivery() {
					t.Errorf("%s: rejected task returned as admitted", name)
				}
				// Nothing partial survives a restart: a write either became
				// durable as a whole (uncertain write) or left nothing.
				committed := h.replay.Commits()
				h.replay.Restart()
				if h.replay.Staged() != 0 {
					t.Errorf("%s: staged write survived restart", name)
				}
				if name == "corrupt record" || name == "invalid record" || name == "capacity exhausted" {
					continue
				}
				a, err = admit(t, b)
				if err != nil || a.Redelivery() != (name == "uncertain write") || h.replay.Commits() != max(committed, 1) {
					t.Errorf("%s, origin %d: recovery after restart wrong", name, origin)
				}
			}
		}
	})

	t.Run("committed state survives restart", func(t *testing.T) {
		h := newHarness(t)
		first, err := h.admitEmbedded(t, b)
		if err != nil {
			t.Fatal("task rejected")
		}
		h.restart(t, 2)
		again, err := h.admitAt(t, 2, b)
		if err != nil || !again.Redelivery() || admission.AdmittedRecord(again) != admission.AdmittedRecord(first) {
			t.Error("redelivery after restart not resolved to the committed record")
		}
		_, err = h.admitAt(t, 2, append(append([]byte(nil), b...), '\n'))
		expect(t, "conflict after restart", err, task.IDConflict)
	})

	t.Run("concurrent identical deliveries resolve to one record", func(t *testing.T) {
		h := newHarness(t)
		const n = 16
		refs := make([]string, n)
		for i := range refs {
			refs[i] = mustEmbedded(t, h.store, b)
		}
		var wg sync.WaitGroup
		var created, failed atomic.Int32
		for _, ref := range refs {
			wg.Add(1)
			go func() {
				defer wg.Done()
				a, err := h.admitter.Admit(context.Background(), ref, b)
				switch {
				case err != nil:
					failed.Add(1)
				case !a.Redelivery():
					created.Add(1)
				}
			}()
		}
		wg.Wait()
		if failed.Load() != 0 || created.Load() != 1 || h.replay.Commits() != 1 {
			t.Errorf("concurrent deliveries: %d created, %d failed, %d commits", created.Load(), failed.Load(), h.replay.Commits())
		}
	})

	t.Run("stored record holds no task content", func(t *testing.T) {
		h := newHarness(t)
		canary := taskWith(map[string]any{"input.subject": "canary-input-value", "options.note": "canary-option-value",
			"labels": map[string]any{"t.k00": "canary-label-value", "t.participant": "canary-participant-value"}})
		a, err := h.admitEmbedded(t, canary)
		if err != nil {
			t.Fatal("canary task rejected")
		}
		if admission.AdmittedRecord(a).Digest != sha256.Sum256(canary) {
			t.Error("digest is not the SHA-256 of the exact accepted bytes")
		}
		dump := h.replay.Dump()
		for _, s := range []string{"canary-input-value", "canary-option-value", "canary-label-value", "canary-participant-value", "t.k00", "t.participant", "reg-a", "SYN-01", "example.test", "subject", "labels"} {
			if bytes.Contains(dump, []byte(s)) {
				t.Errorf("stored replay state contains %q", s)
			}
		}
	})

	t.Run("nil admitter fails closed", func(t *testing.T) {
		var nilAdmitter *admission.Admitter
		if _, err := nilAdmitter.Admit(context.Background(), "", b); code(err) != task.TrustUnavailable {
			t.Error("nil admitter did not fail closed")
		}
	})
}

// Network absence for these rows is structural: admission and the root facade
// have no transport, resolver or Network Context dependency
// (TestAdmissionHasNoNetworkDependency in the root package), so no recording
// transport can be connected to them.
// requirement-ref: REQ-TRUST-0001
// requirement-ref: REQ-TRUST-0004
// requirement-ref: REQ-TRUST-0005
func TestCcTrustAcceptedTaskOrigins(t *testing.T) {
	b := []byte(baseTask)
	ctx := context.Background()
	noState := func(t *testing.T, name string, h *harness, a admission.Admitted, err error, want task.AdmissionCode) {
		t.Helper()
		if code(err) != want {
			t.Errorf("%s: got class %d, want %d", name, code(err), want)
		}
		if admission.AdmittedRecord(a) != (persistence.ReplayRecord{}) || h.replay.Len() != 0 || h.replay.Commits() != 0 || h.replay.Staged() != 0 {
			t.Errorf("%s: rejection wrote a replay or admitted-task record", name)
		}
	}

	t.Run("same bytes are accepted through each reviewed origin", func(t *testing.T) {
		h := newHarness(t)
		source, err := task.NewSourceContext(task.SourceInput{Origin: task.Embedded, IssuerID: "embedded-issuer", PolicyGeneration: 2})
		if _, issued := source.IssuedAt(); err != nil || source.Authenticated() || issued {
			t.Fatal("embedded origin must carry authenticated false and no issued_at")
		}
		if _, err := h.admitEmbedded(t, b); err != nil {
			t.Fatal("embedded origin rejected:", code(err))
		}
		h = newHarness(t)
		if _, err := h.admitRemote(t, b); err != nil {
			t.Fatal("authenticated remote origin with issued_at at the trusted clock rejected:", code(err))
		}
	})

	t.Run("remote origin without issuer authentication", func(t *testing.T) {
		h := newHarness(t)
		v := verifier()
		v.IssuedAt = task.TimestampFromTime(h.clock.Clock.Now())
		v.IssuerAuthenticated = false
		ref, installErr := h.store.InstallRemote(ctx, v, v.Seal(b))
		if installErr == nil || ref != "" {
			t.Fatal("transport encryption without issuer authentication produced a record")
		}
		if _, err := task.NewSourceContext(task.SourceInput{Origin: task.Remote, IssuerID: "remote-issuer", PolicyGeneration: 2, IssuedAt: &v.IssuedAt}); err == nil {
			t.Fatal("remote source with authenticated false was constructed")
		}
		a, err := h.admitter.Admit(ctx, ref, b)
		noState(t, "unauthenticated remote", h, a, err, task.SourceUnauthenticated)
	})

	t.Run("remote record bound to bytes differing by one byte", func(t *testing.T) {
		h := newHarness(t)
		v := verifier()
		v.IssuedAt = task.TimestampFromTime(h.clock.Clock.Now())
		ref, err := h.store.InstallRemote(ctx, v, v.Seal(b))
		if err != nil {
			t.Fatal(err)
		}
		other := bytes.Clone(b)
		other[len(other)-2] = ' '
		a, err := h.admitter.Admit(ctx, ref, other)
		noState(t, "byte mismatch", h, a, err, task.SourceUnauthenticated)
		tampered := v.Seal(b)
		tampered[len(tampered)-2] ^= 1
		if ref, err := h.store.InstallRemote(ctx, v, tampered); err == nil || ref != "" {
			t.Fatal("an envelope failing integrity produced a record")
		}
	})

	t.Run("remote record without issued_at cannot exist", func(t *testing.T) {
		h := newHarness(t)
		if _, err := task.NewSourceContext(task.SourceInput{Origin: task.Remote, IssuerID: "remote-issuer", PolicyGeneration: 2, Authenticated: true}); err == nil {
			t.Fatal("remote source without issued_at was constructed")
		}
		a, err := h.admitter.Admit(ctx, "", b)
		noState(t, "no issued_at, hence no record", h, a, err, task.SourceUnauthenticated)
	})

	t.Run("trusted clock unavailable", func(t *testing.T) {
		h := newHarness(t)
		v := verifier()
		v.IssuedAt = task.TimestampFromTime(h.clock.Clock.Now())
		ref, err := h.store.InstallRemote(ctx, v, v.Seal(b))
		if err != nil {
			t.Fatal(err)
		}
		h.clock.down.Store(true)
		a, err := h.admitter.Admit(ctx, ref, b)
		noState(t, "clock unavailable", h, a, err, task.TrustUnavailable)
	})

	t.Run("no delivery record", func(t *testing.T) {
		for _, ref := range []string{"", strings.Repeat("0", 64), "not-a-reference"} {
			h := newHarness(t)
			a, err := h.admitter.Admit(ctx, ref, b)
			noState(t, "no record", h, a, err, task.SourceUnauthenticated)
		}
	})
}
