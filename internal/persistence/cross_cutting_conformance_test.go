package persistence_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/tv-rain/barometer-go/internal/persistence"
	"github.com/tv-rain/barometer-go/internal/testkit/records"
)

// Cross-cutting family cases owned by internal/persistence. Synthetic core
// contract cases over the fault-injectable record port; no device-durability,
// native storage or support claim.

// faultAt fails the n-th mutating or listing step once armed, before it takes
// effect or, with after set, after it.
type faultAt struct {
	persistence.Records
	mu       sync.Mutex
	n, steps int
	after    bool
}

func (f *faultAt) step() (fail, apply bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.steps++
	if f.steps != f.n {
		return false, true
	}
	return true, f.after
}

func (f *faultAt) Put(ctx context.Context, k persistence.RecordKey, v []byte) error {
	fail, apply := f.step()
	if apply {
		if err := f.Records.Put(ctx, k, v); err != nil || !fail {
			return err
		}
	}
	return persistence.ErrRecordsUnavailable
}

func (f *faultAt) Delete(ctx context.Context, k persistence.RecordKey) error {
	fail, apply := f.step()
	if apply {
		if err := f.Records.Delete(ctx, k); err != nil || !fail {
			return err
		}
	}
	return persistence.ErrRecordsUnavailable
}

func (f *faultAt) List(ctx context.Context, c persistence.RecordClass) ([]persistence.RecordKey, error) {
	if fail, _ := f.step(); fail {
		return nil, persistence.ErrRecordsUnavailable
	}
	return f.Records.List(ctx, c)
}

// linkageKey is a synthetic linkage key that erase deletes. Clear receives
// only erasers, so it cannot derive a replacement; the root Client case
// TestClearDeletesEveryStateClass counts derivations on a key store.
type linkageKey struct{ present, failErase bool }

func (k *linkageKey) erase(context.Context) error {
	if k.failErase {
		return errors.New("synthetic key deletion failure")
	}
	k.present = false
	return nil
}

// TestCcLinkageClearDataContract executes the Go-owned rows of
// CC-LINKAGE-CLEAR-DATA-CONTRACT for persistent state: each clearing step is
// fault-injected over pending bundles, run accounting, replay records and
// the linkage key; a failure is reported explicitly and leaves the scope
// marked; repeated clearing is idempotent and completes from any partial
// state; clearing takes no key store, derivation or transport. Active
// attempts and bounded cleanup are the root Client cases
// (TestLifecycleTransitionsCancelRuns, TestClearRacesEntryPoints); host UI,
// operating-system reset wiring, native keystores and sentinels are DEV-455.
// requirement-ref: REQ-LINKAGE-0023
// requirement-ref: REQ-LINKAGE-0024
func TestCcLinkageClearDataContract(t *testing.T) {
	ctx := context.Background()
	steps := 0
	for _, after := range []bool{false, true} {
		for n := 1; ; n++ {
			r := records.New()
			seedScope(t, r)
			key := &linkageKey{present: true}
			f := &faultAt{Records: r, n: n, after: after}
			err := persistence.Clear(ctx, f, key.erase)
			if n > f.steps {
				if err != nil || !scopeEmpty(r) || key.present {
					t.Fatalf("after=%v: unfaulted clear incomplete", after)
				}
				steps = f.steps
				break
			}
			name := fmt.Sprintf("after=%v/step=%d", after, n)
			if !errors.Is(err, persistence.ErrClearIncomplete) {
				t.Fatalf("%s: fault reported as cleared", name)
			}
			r.Restart()
			marked, err := persistence.Clearing(ctx, r)
			if err != nil || marked != (r.Len(persistence.ClearingClass) == 1) {
				t.Fatalf("%s: marker state not reported", name)
			}
			if !marked && !(n == 1 && !after) && !(scopeEmpty(r) && !key.present) {
				// Unmarked is allowed only before anything is deleted or
				// once everything is.
				t.Fatalf("%s: a partial clear left the scope unmarked", name)
			}
			for i := range 2 {
				if err := persistence.Clear(ctx, r, key.erase); err != nil || !scopeEmpty(r) || key.present {
					t.Fatalf("%s: repeated clear %d did not complete", name, i)
				}
			}
		}
	}
	if steps < 13 {
		t.Fatalf("only %d clearing steps were fault-injected", steps)
	}

	t.Run("failed key deletion keeps the scope marked", func(t *testing.T) {
		r := records.New()
		seedScope(t, r)
		key := &linkageKey{present: true, failErase: true}
		if !errors.Is(persistence.Clear(ctx, r, key.erase), persistence.ErrClearIncomplete) {
			t.Fatal("failed key deletion reported as cleared")
		}
		if marked, _ := persistence.Clearing(ctx, r); !marked || !key.present {
			t.Fatal("scope not marked after failed key deletion")
		}
		key.failErase = false
		if persistence.Clear(ctx, r, key.erase) != nil || key.present {
			t.Fatal("recovery did not delete the key without replacement")
		}
	})
}
