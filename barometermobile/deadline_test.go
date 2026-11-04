package barometermobile

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/persistence"
	"github.com/tv-rain/barometer-go/internal/platform"
)

// slack bounds how long after its deadline a stalled callback may hold its
// caller. It is a liveness bound, not a timing assertion.
const slack = time.Second

func stalledCtx(t *testing.T) (context.Context, func()) {
	t.Helper()
	return context.WithTimeout(context.Background(), 100*time.Millisecond)
}

// expectBounded runs f and fails unless it returns an error within the
// context deadline plus slack.
func expectBounded(t *testing.T, name string, f func(context.Context) error) {
	t.Helper()
	ctx, cancel := stalledCtx(t)
	defer cancel()
	start := time.Now()
	err := f(ctx)
	if err == nil {
		t.Errorf("%s: stalled callback reported success", name)
	}
	if took := time.Since(start); took > 100*time.Millisecond+slack {
		t.Errorf("%s: waited %v past the deadline", name, took)
	}
	if strings.Contains(err.Error(), "SYNTHETIC") {
		t.Errorf("%s: native error text crossed the facade", name)
	}
}

// Every blocking callback gets the caller's remaining deadline, and the
// facade stops waiting at it and treats the call as failed.
func TestStalledCallbacksFailAtTheDeadline(t *testing.T) {
	n := newNatives()
	defer n.openAll()
	for _, g := range []*gate{n.records.gate, n.clock.gate, n.network.gate, n.consent.gate, n.device.gate, n.zone.gate, n.resolver.gate, n.keys.gate} {
		g.stall.Store(true)
	}
	key := persistence.RecordKey{Class: persistence.PendingClass}
	r := newRecords(n.records)
	expectBounded(t, "Records.Get", func(ctx context.Context) error { _, _, err := r.Get(ctx, key); return err })
	expectBounded(t, "Records.Put", func(ctx context.Context) error { return r.Put(ctx, key, []byte("v")) })
	expectBounded(t, "Records.Delete", func(ctx context.Context) error { return r.Delete(ctx, key) })
	expectBounded(t, "Records.List", func(ctx context.Context) error { _, err := r.List(ctx, persistence.ReplayClass); return err })
	expectBounded(t, "Clock", func(ctx context.Context) error { _, err := clock{n.clock}.Now(ctx); return err })
	expectBounded(t, "Network.Access", func(ctx context.Context) error { _, err := network{n.network}.Access(ctx); return err })
	expectBounded(t, "Consent", func(ctx context.Context) error { _, err := consent{n.consent}.Status(ctx); return err })
	expectBounded(t, "Device", func(ctx context.Context) error { _, err := device{n.device}.State(ctx); return err })
	expectBounded(t, "TimeZone", func(ctx context.Context) error { _, err := timeZone{n.zone}.Location(ctx); return err })
	expectBounded(t, "Resolver", func(ctx context.Context) error {
		_, err := resolver{n.resolver}.LookupNetIP(ctx, "ip4", "example.org.")
		return err
	})
	expectBounded(t, "LinkageKeys.Derive", func(ctx context.Context) error {
		_, err := newLinkageKeys(n.keys).DeriveHMACSHA256(ctx, []byte("m"))
		return err
	})
	expectBounded(t, "LinkageKeys.Delete", func(ctx context.Context) error { return newLinkageKeys(n.keys).Delete(ctx) })
	// A record port failure is the port's fixed error or the context's.
	ctx, cancel := stalledCtx(t)
	defer cancel()
	if err := r.Put(ctx, key, []byte("v")); !errors.Is(err, persistence.ErrRecordsUnavailable) && !errors.Is(err, context.DeadlineExceeded) {
		t.Error("abandoned Put is not reported as uncertain:", err)
	}
}

func TestCallbacksReceiveTheRemainingDeadline(t *testing.T) {
	n := newNatives()
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	if _, err := (network{n.network}).Access(ctx); err != nil {
		t.Fatal(err)
	}
	if ms := n.network.lastMillis.Load(); ms < 1 || ms > 250 {
		t.Fatal("timeoutMillis is not the remaining deadline:", ms)
	}
	if _, err := (network{n.network}).Access(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ms := n.network.lastMillis.Load(); ms != defaultCallback.Milliseconds() {
		t.Fatal("a call without a deadline does not get the default bound:", ms)
	}
	expired, cancel2 := context.WithTimeout(context.Background(), -time.Second)
	defer cancel2()
	before := n.network.lastMillis.Load()
	if _, err := (network{n.network}).Access(expired); err == nil {
		t.Fatal("an expired context reached the callback")
	}
	if n.network.lastMillis.Load() != before {
		t.Fatal("an expired context invoked the callback")
	}
}

type panicking struct{ Clock }

func (panicking) NowUnixMillis(int64) (int64, error) { panic("synthetic") }

func TestCallbackErrorsAndUnknownValuesAreUnavailable(t *testing.T) {
	ctx := context.Background()
	n := newNatives()

	n.clock.value.Store(-5)
	if _, err := (clock{n.clock}).Now(ctx); err == nil {
		t.Error("a non-positive time is available")
	}
	n.clock.value.Store(0)
	n.clock.fail.Store(true)
	if _, err := (clock{n.clock}).Now(ctx); err == nil {
		t.Error("a clock error is available")
	}
	if _, err := (clock{panicking{}}).Now(ctx); err == nil {
		t.Error("a panicking callback is available")
	}

	n.network.state = NewAccessState(-3, "satellite", "maybe", "true", "", "YES", "no")
	s, err := network{n.network}.Access(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if s != (platform.AccessState{CaptivePortal: platform.AvailabilityNo}) {
		t.Errorf("unknown access values are not unknown: %+v", s)
	}
	n.network.fail.Store(true)
	if _, err := (network{n.network}).Access(ctx); err == nil {
		t.Error("a network error is available")
	}

	for _, v := range []string{"", "VALID", "unknown", "granted"} {
		if st, err := (consent{&nativeConsent{value: v}}).Status(ctx); err == nil || st != platform.ConsentUnavailable {
			t.Errorf("consent %q is not unavailable", v)
		}
	}
	for v, want := range map[string]platform.ConsentStatus{"valid": platform.ConsentValid, "absent": platform.ConsentAbsent, "ended": platform.ConsentEnded} {
		if st, err := (consent{&nativeConsent{value: v}}).Status(ctx); err != nil || st != want {
			t.Errorf("consent %q: %v %v", v, st, err)
		}
	}

	for _, ds := range []*DeviceState{NewDeviceState(101, true, false, false), NewDeviceState(-1, true, true, false), NewDeviceState(50, false, true, false)} {
		st, err := device{&nativeDevice{state: ds}}.State(ctx)
		if err != nil || st.BatteryKnown || st.Charging || st.ChargingKnown {
			t.Errorf("unknown device values are known: %+v", st)
		}
	}
	if _, err := (device{&nativeDevice{}}).State(ctx); err == nil {
		t.Error("a nil device state is available")
	}

	for _, name := range []string{"", "Local", "Mars/Olympus", "../etc/passwd"} {
		if _, err := (timeZone{&nativeZone{name: name}}).Location(ctx); err == nil {
			t.Errorf("time zone %q is available", name)
		}
	}
	if loc, err := (timeZone{&nativeZone{name: "Asia/Tokyo"}}).Location(ctx); err != nil || loc.String() != "Asia/Tokyo" {
		t.Error("an IANA zone did not load from the embedded database")
	}

	if _, err := newLinkageKeys(&nativeKeys{size: 31}).DeriveHMACSHA256(ctx, nil); err == nil {
		t.Error("a short derivation is available")
	}

	for _, answer := range [][]string{{"1.1.1.1%en0"}, {"example.org"}, {"01.1.1.1"}} {
		if _, err := (resolver{&nativeResolver{answer: answer}}).LookupNetIP(ctx, "ip4", "example.org."); err == nil {
			t.Errorf("resolver answer %q is available", answer)
		}
	}
	if _, err := (resolver{&nativeResolver{answer: []string{"1.1.1.1"}}}).LookupNetIP(ctx, "tcp", "example.org."); err == nil {
		t.Error("an unsupported lookup network reached the callback")
	}
}

func TestRecordKeysAndValuesAreChecked(t *testing.T) {
	ctx := context.Background()
	n := newNatives()
	r := newRecords(n.records)
	key := persistence.RecordKey{Class: persistence.RunClass, Name: [32]byte{0xab, 1}}
	if err := r.Put(ctx, key, []byte("value")); err != nil {
		t.Fatal(err)
	}
	enc := encodeKey(key)
	if enc != "2-ab01"+strings.Repeat("00", 30) {
		t.Fatal("unexpected native key form:", enc)
	}
	keys, err := r.List(ctx, persistence.RunClass)
	if err != nil || len(keys) != 1 || keys[0] != key {
		t.Fatal("round trip failed", keys, err)
	}
	if v, ok, err := r.Get(ctx, key); err != nil || !ok || string(v) != "value" {
		t.Fatal("get failed")
	}
	if _, ok, err := r.Get(ctx, persistence.RecordKey{Class: persistence.RunClass}); err != nil || ok {
		t.Fatal("absent key is not absent")
	}
	for _, bad := range []string{"3-ab01" + strings.Repeat("00", 30), "2-AB01" + strings.Repeat("00", 30), "2-ab01", "2-" + strings.Repeat("zz", 32), "02-ab01" + strings.Repeat("00", 30)} {
		n.records.keys = []string{bad}
		if _, err := r.List(ctx, persistence.RunClass); err == nil {
			t.Errorf("key %q from another class or form is accepted", bad)
		}
	}
	n.records.keys = nil
	if err := r.Put(ctx, key, make([]byte, persistence.MaxRecordValue+1)); err == nil {
		t.Error("an oversized value reached the callback")
	}
	if _, err := r.List(ctx, persistence.RecordClass(9)); err == nil {
		t.Error("an unknown class reached the callback")
	}
	l := NewStringList()
	for range maxListEntries + 1 {
		l.Add("x")
	}
	if !l.overflow {
		t.Error("an unbounded list is accepted")
	}
	n.records.fail.Store(true)
	if _, _, err := r.Get(ctx, key); !errors.Is(err, persistence.ErrRecordsUnavailable) {
		t.Error("a native error is not ErrRecordsUnavailable:", err)
	}
}

func TestWatchFailsClosed(t *testing.T) {
	n := newNatives()
	defer n.openAll()
	net := network{n.network}

	var fired atomic.Int32
	stop := net.Watch(1, func() { fired.Add(1) })
	n.network.mu.Lock()
	var obs NetworkObserver
	for _, o := range n.network.observers {
		obs = o
	}
	n.network.mu.Unlock()
	obs.Changed()
	obs.Changed()
	if fired.Load() != 1 {
		t.Fatal("a change was not delivered exactly once")
	}
	stop()
	stop()

	stop = net.Watch(1, func() { fired.Add(1) })
	n.network.mu.Lock()
	var registered []NetworkObserver
	for _, o := range n.network.observers {
		registered = append(registered, o)
	}
	n.network.mu.Unlock()
	stop()
	for _, o := range registered {
		o.Changed()
	}
	if fired.Load() != 1 {
		t.Fatal("a change was delivered after stop")
	}

	n.network.watchFail.Store(true)
	net.Watch(1, func() { fired.Add(1) })()
	if fired.Load() != 2 {
		t.Fatal("a failed registration is not a change")
	}
	n.network.watchFail.Store(false)

	n.network.watchGate.stall.Store(true)
	start := time.Now()
	stopped := n.network.stopped.Load()
	stop = net.Watch(1, func() { fired.Add(1) })
	if fired.Load() != 3 || time.Since(start) > watchBound+slack {
		t.Fatal("a stalled registration is not a change at its bound")
	}
	stop()
	n.network.watchGate.open()
	for deadline := time.Now().Add(5 * time.Second); n.network.stopped.Load() == stopped; time.Sleep(time.Millisecond) {
		if time.Now().After(deadline) {
			t.Fatal("a late registration was not stopped")
		}
	}
}
