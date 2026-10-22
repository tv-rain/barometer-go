package persistence_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/persistence"
	"github.com/tv-rain/barometer-go/internal/testkit/records"
)

// opLog records the mutating record-port calls a clear makes, in order.
type opLog struct {
	persistence.Records
	mu  sync.Mutex
	ops []string
}

func (l *opLog) log(op string, c persistence.RecordClass) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.ops = append(l.ops, op+classNames[c])
}

var classNames = map[persistence.RecordClass]string{persistence.ReplayClass: " replay", persistence.RunClass: " run",
	persistence.PendingClass: " pending", persistence.ClearingClass: " marker"}

func (l *opLog) Put(ctx context.Context, k persistence.RecordKey, v []byte) error {
	l.log("put", k.Class)
	return l.Records.Put(ctx, k, v)
}

func (l *opLog) Delete(ctx context.Context, k persistence.RecordKey) error {
	l.log("delete", k.Class)
	return l.Records.Delete(ctx, k)
}

func (l *opLog) List(ctx context.Context, c persistence.RecordClass) ([]persistence.RecordKey, error) {
	l.log("list", c)
	return l.Records.List(ctx, c)
}

// seedScope stores one record of every cleared class: a replay record with
// its run record through State, and a pending-class record.
func seedScope(t *testing.T, r *records.Store) *persistence.State {
	t.Helper()
	ctx := context.Background()
	s := open(t, r, 8, stateNow)
	rec := replayRecord(t, "clear-seed", 5, 6*time.Hour)
	if _, _, err := s.Replay().Commit(ctx, rec, stateNow); err != nil {
		t.Fatal(err)
	}
	if err := s.Runs().Put(ctx, rec.Key, runRecord(), stateNow); err != nil {
		t.Fatal(err)
	}
	pk := persistence.RecordKey{Class: persistence.PendingClass, Name: sha256.Sum256([]byte("clear-seed-pending"))}
	if err := r.Put(ctx, pk, []byte("synthetic pending bytes")); err != nil {
		t.Fatal(err)
	}
	return s
}

func scopeEmpty(r *records.Store) bool {
	for c := persistence.RecordClass(1); c.Valid(); c++ {
		if r.Len(c) != 0 {
			return false
		}
	}
	return true
}

// Clear writes the marker first, deletes pending results before run
// accounting and run accounting before replay records, confirms each class
// is empty, runs the erasers and deletes the marker last.
// requirement-ref: REQ-LINKAGE-0023
func TestClearOrder(t *testing.T) {
	r := records.New()
	seedScope(t, r)
	l := &opLog{Records: r}
	var erased []string
	eraser := func(context.Context) error {
		l.mu.Lock()
		defer l.mu.Unlock()
		erased = append(erased, "erase")
		l.ops = append(l.ops, "erase")
		return nil
	}
	if err := persistence.Clear(context.Background(), l, eraser); err != nil {
		t.Fatal(err)
	}
	want := []string{"put marker",
		"list pending", "delete pending", "list pending",
		"list run", "delete run", "list run",
		"list replay", "delete replay", "list replay",
		"erase", "list marker", "delete marker", "list marker"}
	if len(l.ops) != len(want) {
		t.Fatalf("ops %v, want %v", l.ops, want)
	}
	for i := range want {
		if l.ops[i] != want[i] {
			t.Fatalf("ops %v, want %v", l.ops, want)
		}
	}
	if !scopeEmpty(r) || len(erased) != 1 {
		t.Fatal("scope not cleared")
	}
}

// Every class but the marker's is cleared, so a class added to RecordClass
// is cleared without registration.
func TestClearCoversEveryRecordClass(t *testing.T) {
	r := records.New()
	ctx := context.Background()
	n := 0
	for c := persistence.RecordClass(1); c.Valid(); c++ {
		n++
		if err := r.Put(ctx, persistence.RecordKey{Class: c, Name: sha256.Sum256([]byte{byte(c)})}, []byte{1}); err != nil {
			t.Fatal(err)
		}
	}
	if n != 4 || persistence.RecordClass(0).Valid() || persistence.RecordClass(n+1).Valid() {
		t.Fatal("record class set changed: review clearing and this test")
	}
	if err := persistence.Clear(ctx, r); err != nil || !scopeEmpty(r) {
		t.Fatal("a record class survived clearing")
	}
}

// Every injected record-port fault stops Clear with ErrClearIncomplete. The
// marker is reported for as long as it is stored, including after a restart,
// and a later Clear completes from any partial state.
func TestClearFaults(t *testing.T) {
	ctx := context.Background()
	for name, f := range map[string]records.Fault{
		"marker write fails":      records.FailBeforePut,
		"marker write crashes":    records.CrashDuringPut,
		"marker write uncertain":  records.FailAfterPut,
		"listing fails":           records.FailList,
		"deletion fails":          records.FailBeforeDelete,
		"deletion uncertain":      records.FailAfterDelete,
		"storage stalls":          records.Stall,
		"reads fail during clear": records.FailGet,
	} {
		t.Run(name, func(t *testing.T) {
			r := records.New()
			s := seedScope(t, r)
			r.Inject(f, true)
			cctx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
			err := persistence.Clear(cctx, r)
			cancel()
			r.Restart()
			marker := r.Len(persistence.ClearingClass) == 1
			if f == records.FailGet {
				// Clear does not read: it completes.
				if err != nil || !scopeEmpty(r) {
					t.Fatal("clear without reads failed")
				}
				return
			}
			if !errors.Is(err, persistence.ErrClearIncomplete) {
				t.Fatal("fault reported as cleared")
			}
			got, err := persistence.Clearing(ctx, r)
			if err != nil || got != marker {
				t.Fatalf("Clearing = %v, %v with marker stored %v", got, err, marker)
			}
			if err := persistence.Clear(ctx, r); err != nil || !scopeEmpty(r) {
				t.Fatal("recovery clear did not complete")
			}
			if got, err := persistence.Clearing(ctx, r); got || err != nil {
				t.Fatal("marker reported after completed clearing")
			}
			s.Forget()
		})
	}

	t.Run("eraser fails", func(t *testing.T) {
		r := records.New()
		seedScope(t, r)
		failing := func(context.Context) error { return errors.New("synthetic eraser failure") }
		if !errors.Is(persistence.Clear(ctx, r, failing), persistence.ErrClearIncomplete) {
			t.Fatal("eraser failure reported as cleared")
		}
		if got, _ := persistence.Clearing(ctx, r); !got {
			t.Fatal("marker missing after an eraser failure")
		}
		if persistence.Clear(ctx, r, nil) == nil {
			t.Fatal("nil eraser accepted")
		}
	})

	t.Run("invalid inputs", func(t *testing.T) {
		cancelled, cancel := context.WithCancel(ctx)
		cancel()
		if persistence.Clear(ctx, nil) == nil || persistence.Clear(cancelled, records.New()) == nil {
			t.Fatal("clear without storage or with a cancelled context succeeded")
		}
		if _, err := persistence.Clearing(ctx, nil); err == nil {
			t.Fatal("marker read without storage")
		}
		r := records.New()
		r.Inject(records.FailGet, true)
		if _, err := persistence.Clearing(ctx, r); err == nil {
			t.Fatal("unreadable marker reported as absent")
		}
	})
}

// A stored value other than the marker's still blocks the scope.
func TestClearingTreatsAnyMarkerValueAsPresent(t *testing.T) {
	r := records.New()
	if err := persistence.Clear(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	r.Inject(records.FailBeforeDelete, true)
	_ = persistence.Clear(context.Background(), r)
	r.Restart()
	keys, _ := r.List(context.Background(), persistence.ClearingClass)
	if len(keys) != 1 {
		t.Fatal("marker not stored")
	}
	r.Corrupt(keys[0], []byte("garbage"))
	if got, err := persistence.Clearing(context.Background(), r); !got || err != nil {
		t.Fatal("corrupt marker not treated as present")
	}
}

// After Clear, Forget drops the State index: no cleared generation, capacity
// or retention is still counted, and the cleared key commits as new.
// requirement-ref: REQ-TRUST-0025
func TestStateForgetAfterClear(t *testing.T) {
	ctx := context.Background()
	r := records.New()
	s := open(t, r, 1, stateNow)
	rec := replayRecord(t, "forget", 7, 6*time.Hour)
	if _, _, err := s.Replay().Commit(ctx, rec, stateNow); err != nil {
		t.Fatal(err)
	}
	if err := persistence.Clear(ctx, r); err != nil {
		t.Fatal(err)
	}
	s.Replay().(persistence.Forgetter).Forget()
	s.Runs().(persistence.Forgetter).Forget()
	if g, err := s.Replay().HighestGeneration(ctx, stateNow); err != nil || g != 0 {
		t.Fatal("cleared generation still counted:", g)
	}
	other := replayRecord(t, "after-forget", 7, 6*time.Hour)
	if _, created, err := s.Replay().Commit(ctx, other, stateNow); err != nil || !created {
		t.Fatal("cleared record still holds capacity")
	}
}
