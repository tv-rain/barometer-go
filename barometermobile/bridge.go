package barometermobile

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tv-rain/barometer-go/internal/persistence"
	"github.com/tv-rain/barometer-go/internal/platform"

	// The time-zone database is embedded so that IANA identifiers from the
	// native TimeZone adapter resolve the same way on Android and iOS.
	_ "time/tzdata"
)

// Callback deadline bounds. A call with no context deadline gets the default;
// no call waits longer than the maximum.
const (
	defaultCallback = 5 * time.Second
	maxCallback     = 30 * time.Second
	// watchBound is how long Watch may take to register before the facade
	// treats the access network as changed.
	watchBound = time.Second
)

var (
	errCallback = errors.New("barometermobile: callback failed")
	errDeadline = errors.New("barometermobile: callback deadline exceeded")
)

// budget returns how long a callback under ctx may run.
func budget(ctx context.Context) (time.Duration, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	d := defaultCallback
	if dl, ok := ctx.Deadline(); ok {
		if d = time.Until(dl); d <= 0 {
			return 0, context.DeadlineExceeded
		}
	}
	return min(d, maxCallback), nil
}

// millis rounds d up to whole milliseconds, at least one.
func millis(d time.Duration) int64 {
	return max(1, int64((d+time.Millisecond-1)/time.Millisecond))
}

// call runs f on its own goroutine with the remaining deadline and stops
// waiting at that deadline or when ctx ends. A late result is discarded; a
// panic is a failure. f is never started once the caller stopped waiting or
// the deadline passed.
func call[T any](ctx context.Context, f func(timeoutMillis int64) (T, error)) (T, error) {
	return dispatch(ctx, nil, f)
}

// sequencer hands the callbacks of one adapter to native code one at a time,
// in the order Go issued them. A call starts only after every earlier call
// returned from native code or was skipped, so a call issued earlier is
// entered first or never entered.
type sequencer struct {
	mu   sync.Mutex
	tail chan struct{} // closed when the latest issued call is finished
}

func newSequencer() *sequencer {
	s := &sequencer{tail: make(chan struct{})}
	close(s.tail)
	return s
}

// enqueue takes the next place in issue order. The caller waits for prev
// before entering native code and closes mine when finished with it.
func (s *sequencer) enqueue() (prev <-chan struct{}, mine chan struct{}) {
	mine = make(chan struct{})
	s.mu.Lock()
	prev, s.tail = s.tail, mine
	s.mu.Unlock()
	return prev, mine
}

// Call states shared by a waiting caller and the goroutine that enters f.
const (
	callPending int32 = iota
	callEntered
	callAbandoned
)

// dispatch is call, ordered by s when s is not nil. f receives the time left
// to the caller's deadline when it is entered.
func dispatch[T any](ctx context.Context, s *sequencer, f func(timeoutMillis int64) (T, error)) (T, error) {
	var zero T
	d, err := budget(ctx)
	if err != nil {
		return zero, err
	}
	deadline := time.Now().Add(d)
	var prev <-chan struct{}
	var mine chan struct{}
	if s != nil {
		prev, mine = s.enqueue()
	}
	type result struct {
		v   T
		err error
	}
	var state atomic.Int32
	done := make(chan result, 1)
	go func() {
		if s != nil {
			defer close(mine)
			<-prev
		}
		left := time.Until(deadline)
		if left <= 0 || ctx.Err() != nil || !state.CompareAndSwap(callPending, callEntered) {
			return
		}
		defer func() {
			if recover() != nil {
				done <- result{zero, errCallback}
			}
		}()
		v, err := f(millis(left))
		done <- result{v, err}
	}()
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case r := <-done:
		return r.v, r.err
	case <-ctx.Done():
		state.CompareAndSwap(callPending, callAbandoned)
		return zero, ctx.Err()
	case <-t.C:
		state.CompareAndSwap(callPending, callAbandoned)
		return zero, errDeadline
	}
}

// ---- Records

// encodeKey is the native form of a record key.
func encodeKey(k persistence.RecordKey) string {
	return strconv.Itoa(int(k.Class)) + "-" + hex.EncodeToString(k.Name[:])
}

// decodeKey accepts only the exact form encodeKey produces for class.
func decodeKey(s string, class persistence.RecordClass) (persistence.RecordKey, bool) {
	prefix := strconv.Itoa(int(class)) + "-"
	rest, ok := strings.CutPrefix(s, prefix)
	if !ok || len(rest) != 2*persistence.RecordNameSize || strings.ToLower(rest) != rest {
		return persistence.RecordKey{}, false
	}
	k := persistence.RecordKey{Class: class}
	if n, err := hex.Decode(k.Name[:], []byte(rest)); err != nil || n != persistence.RecordNameSize {
		return persistence.RecordKey{}, false
	}
	return k, true
}

// records orders its callbacks: the native store sees them in issue order,
// so a call abandoned at its deadline cannot arrive after a later Clear.
type records struct {
	native Records
	seq    *sequencer
}

func newRecords(native Records) records { return records{native, newSequencer()} }

// recordErr returns ctx's error when it ended, else ErrRecordsUnavailable.
func recordErr(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return persistence.ErrRecordsUnavailable
}

func (r records) Get(ctx context.Context, key persistence.RecordKey) ([]byte, bool, error) {
	if !key.Class.Valid() {
		return nil, false, persistence.ErrRecordsUnavailable
	}
	v, err := dispatch(ctx, r.seq, func(ms int64) (*RecordValue, error) { return r.native.Get(encodeKey(key), ms) })
	if err != nil {
		return nil, false, recordErr(ctx)
	}
	if v == nil {
		return nil, false, persistence.ErrRecordsUnavailable
	}
	if v.absent {
		return nil, false, nil
	}
	if len(v.value) > persistence.MaxRecordValue {
		return nil, false, persistence.ErrRecordsUnavailable
	}
	return append([]byte{}, v.value...), true, nil
}

func (r records) Put(ctx context.Context, key persistence.RecordKey, value []byte) error {
	if !key.Class.Valid() || len(value) > persistence.MaxRecordValue {
		return persistence.ErrRecordsUnavailable
	}
	value = append([]byte{}, value...)
	if _, err := dispatch(ctx, r.seq, func(ms int64) (struct{}, error) { return struct{}{}, r.native.Put(encodeKey(key), value, ms) }); err != nil {
		return recordErr(ctx)
	}
	return nil
}

func (r records) Delete(ctx context.Context, key persistence.RecordKey) error {
	if !key.Class.Valid() {
		return persistence.ErrRecordsUnavailable
	}
	if _, err := dispatch(ctx, r.seq, func(ms int64) (struct{}, error) { return struct{}{}, r.native.Delete(encodeKey(key), ms) }); err != nil {
		return recordErr(ctx)
	}
	return nil
}

func (r records) List(ctx context.Context, class persistence.RecordClass) ([]persistence.RecordKey, error) {
	if !class.Valid() {
		return nil, persistence.ErrRecordsUnavailable
	}
	l, err := dispatch(ctx, r.seq, func(ms int64) (*StringList, error) { return r.native.List(int64(class), ms) })
	if err != nil {
		return nil, recordErr(ctx)
	}
	if l == nil {
		return nil, nil
	}
	if l.overflow || len(l.items) > persistence.MaxClassRecords {
		return nil, persistence.ErrRecordsUnavailable
	}
	keys := make([]persistence.RecordKey, 0, len(l.items))
	for _, s := range l.items {
		k, ok := decodeKey(s, class)
		if !ok {
			return nil, persistence.ErrRecordsUnavailable
		}
		keys = append(keys, k)
	}
	return keys, nil
}

// ---- Linkage keys

// linkageKeys orders its callbacks like records, so an abandoned derivation,
// which creates the key on first use, cannot recreate it after a later
// Delete.
type linkageKeys struct {
	native LinkageKeyStore
	seq    *sequencer
}

func newLinkageKeys(native LinkageKeyStore) linkageKeys {
	return linkageKeys{native, newSequencer()}
}

func (l linkageKeys) DeriveHMACSHA256(ctx context.Context, message []byte) ([32]byte, error) {
	var out [32]byte
	message = append([]byte{}, message...)
	b, err := dispatch(ctx, l.seq, func(ms int64) ([]byte, error) { return l.native.DeriveHMACSHA256(message, ms) })
	if err != nil || len(b) != sha256.Size {
		return out, errCallback
	}
	copy(out[:], b)
	return out, nil
}

func (l linkageKeys) Delete(ctx context.Context) error {
	_, err := dispatch(ctx, l.seq, func(ms int64) (struct{}, error) { return struct{}{}, l.native.Delete(ms) })
	if err != nil {
		return errCallback
	}
	return nil
}

// ---- Clock

type clock struct{ native Clock }

func (c clock) Now(ctx context.Context) (time.Time, error) {
	ms, err := call(ctx, c.native.NowUnixMillis)
	if err != nil || ms <= 0 {
		return time.Time{}, errCallback
	}
	return time.UnixMilli(ms).UTC(), nil
}

// ---- Network

type network struct{ native Network }

func networkType(s string) platform.NetworkType {
	switch s {
	case "wifi":
		return platform.NetworkWiFi
	case "cellular":
		return platform.NetworkCellular
	case "ethernet":
		return platform.NetworkEthernet
	case "other":
		return platform.NetworkOther
	}
	return platform.NetworkUnknown
}

func vpnState(s string) platform.VPNState {
	switch s {
	case "detected":
		return platform.VPNDetected
	case "not_detected":
		return platform.VPNNotDetected
	}
	return platform.VPNUnknown
}

func availability(s string) platform.Availability {
	switch s {
	case "yes":
		return platform.AvailabilityYes
	case "no":
		return platform.AvailabilityNo
	}
	return platform.AvailabilityUnknown
}

// suspension maps the native suspension state; any other value is unknown,
// which does not block execution.
func suspension(s string) platform.Availability {
	switch s {
	case "suspended":
		return platform.AvailabilityYes
	case "active":
		return platform.AvailabilityNo
	}
	return platform.AvailabilityUnknown
}

func (n network) Access(ctx context.Context) (platform.AccessState, error) {
	a, err := call(ctx, n.native.Access)
	if err != nil || a == nil {
		return platform.AccessState{}, errCallback
	}
	s := platform.AccessState{Type: networkType(a.networkType), VPN: vpnState(a.vpn), IPv4: availability(a.ipv4),
		IPv6: availability(a.ipv6), Validated: availability(a.validated), CaptivePortal: availability(a.cpo), Suspended: suspension(a.suspended)}
	if a.generation > 0 {
		s.Generation = uint64(a.generation)
	}
	return s, nil
}

// observer delivers at most one change and none after its watch stopped.
type observer struct {
	once    sync.Once
	stopped atomic.Bool
	changed func()
}

// Changed is called by native code.
func (o *observer) Changed() { o.fire() }

func (o *observer) fire() {
	if !o.stopped.Load() {
		o.once.Do(o.changed)
	}
}

// Watch registers a native watch. A registration that fails, or does not
// return within watchBound, is treated as a change, so the run observing it
// fails closed; a registration that returns late is stopped.
func (n network) Watch(generation uint64, changed func()) func() {
	o := &observer{changed: changed}
	if generation > math.MaxInt64 {
		o.fire()
		return func() {}
	}
	type result struct {
		id  int64
		err error
	}
	done := make(chan result, 1)
	go func() {
		defer func() {
			if recover() != nil {
				done <- result{0, errCallback}
			}
		}()
		id, err := n.native.Watch(int64(generation), o)
		done <- result{id, err}
	}()
	t := time.NewTimer(watchBound)
	defer t.Stop()
	select {
	case r := <-done:
		if r.err != nil {
			o.fire()
			return func() { o.stopped.Store(true) }
		}
		var stop sync.Once
		return func() {
			o.stopped.Store(true)
			stop.Do(func() { go n.stopWatch(r.id) })
		}
	case <-t.C:
		o.fire()
		go func() {
			if r := <-done; r.err == nil {
				n.stopWatch(r.id)
			}
		}()
		return func() { o.stopped.Store(true) }
	}
}

func (n network) stopWatch(id int64) {
	defer func() { _ = recover() }()
	n.native.StopWatch(id)
}

// ---- Consent, device, time zone

type consent struct{ native Consent }

func (c consent) Status(ctx context.Context) (platform.ConsentStatus, error) {
	s, err := call(ctx, c.native.Status)
	if err != nil {
		return platform.ConsentUnavailable, errCallback
	}
	switch s {
	case "valid":
		return platform.ConsentValid, nil
	case "absent":
		return platform.ConsentAbsent, nil
	case "ended":
		return platform.ConsentEnded, nil
	}
	return platform.ConsentUnavailable, errCallback
}

type device struct{ native Device }

func (d device) State(ctx context.Context) (platform.DeviceState, error) {
	s, err := call(ctx, d.native.State)
	if err != nil || s == nil {
		return platform.DeviceState{}, errCallback
	}
	out := platform.DeviceState{Charging: s.charging && s.chargingKnown, ChargingKnown: s.chargingKnown}
	if s.batteryKnown && s.batteryPercent >= 0 && s.batteryPercent <= 100 {
		out.BatteryPercent, out.BatteryKnown = int(s.batteryPercent), true
	}
	return out, nil
}

type timeZone struct{ native TimeZone }

func (z timeZone) Location(ctx context.Context) (*time.Location, error) {
	name, err := call(ctx, z.native.Identifier)
	// Only IANA names: "Local" would select the process zone and "" UTC.
	if err != nil || name == "" || name == "Local" || len(name) > 64 {
		return nil, errCallback
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, errCallback
	}
	return loc, nil
}

// ---- System resolver

type resolver struct{ native SystemResolver }

// maxAddresses bounds one lookup's answer.
const maxAddresses = 64

func (r resolver) LookupNetIP(ctx context.Context, network, host string) ([]netip.Addr, error) {
	if network != "ip" && network != "ip4" && network != "ip6" {
		return nil, errCallback
	}
	l, err := call(ctx, func(ms int64) (*StringList, error) { return r.native.LookupNetIP(network, host, ms) })
	if err != nil || l == nil || l.overflow || len(l.items) > maxAddresses {
		return nil, errCallback
	}
	out := make([]netip.Addr, 0, len(l.items))
	for _, s := range l.items {
		a, err := netip.ParseAddr(s)
		if err != nil || a.Zone() != "" {
			return nil, errCallback
		}
		out = append(out, a)
	}
	return out, nil
}
