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
