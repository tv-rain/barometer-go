package barometermobile

import (
	"context"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/persistence"
)

// lateRecords is a native store whose first Put is delayed before it
// arrives, as when the thread carrying it is scheduled late. Native code
// sees only the order calls arrive in.
type lateRecords struct {
	*memRecords
	delay   chan struct{} // closed to let the first Put arrive
	arrived chan struct{} // closed when the first Put has been applied
	first   chan struct{} // holds one token for the first Put
}

func newLateRecords() *lateRecords {
	r := &lateRecords{memRecords: newMemRecords(), delay: make(chan struct{}), arrived: make(chan struct{}), first: make(chan struct{}, 1)}
	r.first <- struct{}{}
	return r
}

func (r *lateRecords) Put(key string, value []byte, timeoutMillis int64) error {
	select {
	case <-r.first:
		<-r.delay
		defer close(r.arrived)
	default:
	}
	return r.memRecords.Put(key, value, timeoutMillis)
}

// An abandoned Put delayed before it reaches native code never lands after a
// later Clear has completed: either it arrives first and Clear deletes it, or
// it never arrives.
func TestAbandonedPutCannotLandAfterClear(t *testing.T) {
	native := newLateRecords()
	r := newRecords(native)
	key := persistence.RecordKey{Class: persistence.PendingClass, Name: [32]byte{1}}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	if err := r.Put(ctx, key, []byte("late")); err == nil {
		t.Fatal("the delayed Put was not abandoned")
	}
	cancel()

	cleared := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cleared <- persistence.Clear(ctx, r)
	}()
	// Let a Clear that does not wait for the abandoned Put finish first.
	select {
	case err := <-cleared:
		cleared <- err
	case <-time.After(200 * time.Millisecond):
	}
	close(native.delay)
	if err := <-cleared; err != nil {
		t.Fatal("Clear failed:", err)
	}
	select {
	case <-native.arrived:
	case <-time.After(5 * time.Second):
		t.Fatal("the delayed Put never finished")
	}
	if n := native.count(); n != 0 {
		t.Fatalf("%d record(s) stored after Clear completed", n)
	}
}

// A record call whose deadline passes while an earlier call is still in
// native code is never handed over, and calls reach native code in the order
// they were issued.
func TestQueuedRecordCallPastItsDeadlineIsNeverHandedOver(t *testing.T) {
	n := newNatives()
	defer n.openAll()
	r := newRecords(n.records)
	key := persistence.RecordKey{Class: persistence.PendingClass}

	n.records.gate.stall.Store(true)
	held := make(chan struct{})
	go func() {
		defer close(held)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _, _ = r.Get(ctx, key)
	}()
	<-n.records.gate.entered

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if err := r.Put(ctx, key, []byte("v")); err == nil {
		t.Fatal("a Put queued behind a stalled call reported success")
	}
	n.records.gate.stall.Store(false)
	n.records.gate.open()
	<-held

	// A later call is handed over only after the expired Put was skipped.
	if _, ok, err := r.Get(context.Background(), key); err != nil || ok {
		t.Fatal("the expired Put reached native code:", ok, err)
	}
	if p := n.records.puts.Load(); p != 0 {
		t.Fatal("the expired Put reached native code:", p)
	}
}
