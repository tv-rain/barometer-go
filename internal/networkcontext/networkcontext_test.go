package networkcontext_test

import (
	"context"
	"testing"

	"github.com/tv-rain/barometer-go/internal/networkcontext"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/testkit/capability"
)

// noStop is a capability that returns no stop function from Watch.
type noStop struct{ *capability.Network }

func (noStop) Watch(uint64, func()) func() { return nil }

// Capture fails closed: a missing capability, a capability error, a zero
// generation and an ended context produce no snapshot, and nothing is guessed.
func TestCaptureFailsClosed(t *testing.T) {
	if _, err := networkcontext.NewDirect(nil); err == nil {
		t.Fatal("nil capability accepted")
	}
	n := capability.NewNetwork(platform.AccessState{Generation: 1})
	d, err := networkcontext.NewDirect(n)
	if err != nil {
		t.Fatal(err)
	}
	n.Fail()
	if _, err := d.Capture(context.Background()); err == nil {
		t.Fatal("capability error captured")
	}
	n.Set(platform.AccessState{Type: platform.NetworkWiFi})
	if _, err := d.Capture(context.Background()); err == nil {
		t.Fatal("zero generation captured")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	n.Set(platform.AccessState{Generation: 1})
	calls := n.Calls()
	if _, err := d.Capture(ctx); err == nil || n.Calls() != calls {
		t.Fatal("ended context captured or queried the capability")
	}
	var zero networkcontext.Snapshot
	if _, err := zero.Context(false); err == nil {
		t.Fatal("zero snapshot produced a context")
	}
}

// Watch reports a change once and stops cleanly; a snapshot that cannot be
// watched reports a change at once, so no attempt runs unwatched.
func TestSnapshotWatch(t *testing.T) {
	n := capability.NewNetwork(platform.AccessState{Generation: 4})
	d, _ := networkcontext.NewDirect(n)
	snap, err := d.Capture(context.Background())
	if err != nil || snap.Generation() != 4 {
		t.Fatal(err)
	}
	fired := 0
	stop := snap.Watch(func() { fired++ })
	n.Set(platform.AccessState{Generation: 4, Type: platform.NetworkCellular})
	if fired != 0 {
		t.Fatal("same generation reported as a change")
	}
	n.Set(platform.AccessState{Generation: 5})
	n.Set(platform.AccessState{Generation: 6})
	stop()
	stop()
	if fired != 1 || n.Watching() != 0 {
		t.Fatalf("fired %d, watching %d", fired, n.Watching())
	}
	stop = snap.Watch(func() { fired++ }) // generation already moved
	stop()
	if fired != 2 {
		t.Fatal("an earlier change was not reported")
	}
	var zero networkcontext.Snapshot
	zero.Watch(func() { fired++ })()
	if fired != 3 {
		t.Fatal("unwatchable zero snapshot reported no change")
	}
	m := noStop{capability.NewNetwork(platform.AccessState{Generation: 1})}
	d, _ = networkcontext.NewDirect(m)
	if snap, err = d.Capture(context.Background()); err != nil {
		t.Fatal(err)
	}
	snap.Watch(func() { fired++ })()
	if fired != 4 {
		t.Fatal("capability without a stop function reported no change")
	}
}

// Suspension is local execution metadata: only a reported yes is suspension,
// it never reaches the Network Context, and a re-read after a change is
// suspension only when the platform reports it then. A failed re-read, a zero
// generation, an unknown state and an ended context are not suspension.
func TestSnapshotSuspension(t *testing.T) {
	ctx := context.Background()
	for v, want := range map[platform.Availability]bool{platform.AvailabilityYes: true, platform.AvailabilityNo: false, platform.AvailabilityUnknown: false, 9: false} {
		n := capability.NewNetwork(platform.AccessState{Generation: 1, Suspended: v})
		d, err := networkcontext.NewDirect(n)
		if err != nil {
			t.Fatal(err)
		}
		snap, err := d.Capture(ctx)
		if err != nil || snap.Suspended() != want {
			t.Fatalf("suspended %d: %v %v", v, snap.Suspended(), err)
		}
		plain := capability.NewNetwork(platform.AccessState{Generation: 1})
		p, _ := networkcontext.NewDirect(plain)
		other, _ := p.Capture(ctx)
		a, _ := snap.Context(true)
		b, _ := other.Context(true)
		if a.Access() != b.Access() {
			t.Fatal("suspension changed the Network Context")
		}
	}
	if (networkcontext.Snapshot{}).Suspended() || (networkcontext.Snapshot{}).SuspendedNow(ctx) {
		t.Fatal("an invalid snapshot is suspended")
	}

	n := capability.NewNetwork(platform.AccessState{Generation: 1})
	d, _ := networkcontext.NewDirect(n)
	snap, _ := d.Capture(ctx)
	if snap.SuspendedNow(ctx) {
		t.Fatal("an active platform is suspended")
	}
	n.Set(platform.AccessState{Generation: 2, Suspended: platform.AvailabilityYes})
	if !snap.SuspendedNow(ctx) {
		t.Fatal("a reported suspension was not seen")
	}
	ended, cancel := context.WithCancel(ctx)
	cancel()
	if snap.SuspendedNow(ended) {
		t.Fatal("an ended context read suspension")
	}
	n.Set(platform.AccessState{Suspended: platform.AvailabilityYes})
	if snap.SuspendedNow(ctx) {
		t.Fatal("a zero generation is suspended")
	}
	n.Fail()
	if snap.SuspendedNow(ctx) {
		t.Fatal("a failed read is suspended")
	}
}
