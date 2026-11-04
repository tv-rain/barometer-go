package barometermobile

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Synthetic native adapters, implemented in Go against the facade interfaces
// as Kotlin or Swift code would implement them. No real platform state, no
// participant data and no uncontrolled network destination is used.

var errSynthetic = errors.New("SYNTHETIC-NATIVE-ERROR")

// gate stalls a callback until released or until the test ends.
type gate struct {
	stall   atomic.Bool
	entered chan struct{}
	release chan struct{}
	once    sync.Once
}

func newGate() *gate { return &gate{entered: make(chan struct{}, 64), release: make(chan struct{})} }

func (g *gate) wait() {
	if g == nil || !g.stall.Load() {
		return
	}
	select {
	case g.entered <- struct{}{}:
	default:
	}
	<-g.release
}

func (g *gate) open() { g.once.Do(func() { close(g.release) }) }

type memRecords struct {
	mu     sync.Mutex
	values map[string][]byte
	gate   *gate
	fail   atomic.Bool
	keys   []string // overrides List output when set
	puts   atomic.Int32
}

func newMemRecords() *memRecords { return &memRecords{values: map[string][]byte{}, gate: newGate()} }

func (r *memRecords) Get(key string, timeoutMillis int64) (*RecordValue, error) {
	r.gate.wait()
	if r.fail.Load() {
		return nil, errSynthetic
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	v, ok := r.values[key]
	if !ok {
		return nil, nil
	}
	return NewRecordValue(v), nil
}

func (r *memRecords) Put(key string, value []byte, timeoutMillis int64) error {
	r.gate.wait()
	if r.fail.Load() {
		return errSynthetic
	}
	r.puts.Add(1)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.values[key] = append([]byte{}, value...)
	return nil
}

func (r *memRecords) Delete(key string, timeoutMillis int64) error {
	r.gate.wait()
	if r.fail.Load() {
		return errSynthetic
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.values, key)
	return nil
}

func (r *memRecords) List(class int64, timeoutMillis int64) (*StringList, error) {
	r.gate.wait()
	if r.fail.Load() {
		return nil, errSynthetic
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	l := NewStringList()
	if r.keys != nil {
		for _, k := range r.keys {
			l.Add(k)
		}
		return l, nil
	}
	var keys []string
	for k := range r.values {
		if strings.HasPrefix(k, itoa(class)+"-") {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		l.Add(k)
	}
	return l, nil
}

func (r *memRecords) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.values)
}

func itoa(n int64) string { b, _ := json.Marshal(n); return string(b) }

type nativeClock struct {
	gate  *gate
	value atomic.Int64 // when non-zero, returned instead of the current time
	fail  atomic.Bool
}

func (c *nativeClock) NowUnixMillis(timeoutMillis int64) (int64, error) {
	c.gate.wait()
	if c.fail.Load() {
		return 0, errSynthetic
	}
	if v := c.value.Load(); v != 0 {
		return v, nil
	}
	return time.Now().UnixMilli(), nil
}

type nativeNetwork struct {
	gate       *gate
	mu         sync.Mutex
	state      *AccessState
	fail       atomic.Bool
	watchGate  *gate
	watchFail  atomic.Bool
	nextID     atomic.Int64
	observers  map[int64]NetworkObserver
	stopped    atomic.Int32
	lastMillis atomic.Int64
}

func newNativeNetwork() *nativeNetwork {
	return &nativeNetwork{gate: newGate(), watchGate: newGate(), observers: map[int64]NetworkObserver{},
		state: NewAccessState(1, "wifi", "not_detected", "yes", "no", "yes", "no")}
}

func (n *nativeNetwork) Access(timeoutMillis int64) (*AccessState, error) {
	n.lastMillis.Store(timeoutMillis)
	n.gate.wait()
	if n.fail.Load() {
		return nil, errSynthetic
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.state, nil
}

func (n *nativeNetwork) Watch(generation int64, observer NetworkObserver) (int64, error) {
	n.watchGate.wait()
	if n.watchFail.Load() {
		return 0, errSynthetic
	}
	id := n.nextID.Add(1)
	n.mu.Lock()
	n.observers[id] = observer
	n.mu.Unlock()
	return id, nil
}

func (n *nativeNetwork) StopWatch(watchID int64) {
	n.stopped.Add(1)
	n.mu.Lock()
	delete(n.observers, watchID)
	n.mu.Unlock()
}

type nativeConsent struct {
	gate  *gate
	value string
}

func (c *nativeConsent) Status(timeoutMillis int64) (string, error) {
	c.gate.wait()
	return c.value, nil
}

type nativeDevice struct {
	gate  *gate
	state *DeviceState
}

func (d *nativeDevice) State(timeoutMillis int64) (*DeviceState, error) {
	d.gate.wait()
	return d.state, nil
}

type nativeZone struct {
	gate *gate
	name string
}

func (z *nativeZone) Identifier(timeoutMillis int64) (string, error) {
	z.gate.wait()
	return z.name, nil
}

type nativeResolver struct {
	gate    *gate
	mu      sync.Mutex
	queries []string
	answer  []string
	err     error
}

func (r *nativeResolver) LookupNetIP(network string, host string, timeoutMillis int64) (*StringList, error) {
	r.gate.wait()
	r.mu.Lock()
	r.queries = append(r.queries, network+" "+host)
	r.mu.Unlock()
	if r.err != nil {
		return nil, r.err
	}
	l := NewStringList()
	for _, a := range r.answer {
		l.Add(a)
	}
	return l, nil
}

type nativeKeys struct {
	gate    *gate
	size    int
	deletes atomic.Int32
}

func (k *nativeKeys) DeriveHMACSHA256(message []byte, timeoutMillis int64) ([]byte, error) {
	k.gate.wait()
	out := make([]byte, k.size)
	for i := range out {
		out[i] = byte(i)
	}
	return out, nil
}

func (k *nativeKeys) Delete(timeoutMillis int64) error {
	k.gate.wait()
	k.deletes.Add(1)
	return nil
}

// natives groups one synthetic native adapter set.
type natives struct {
	records  *memRecords
	clock    *nativeClock
	network  *nativeNetwork
	consent  *nativeConsent
	device   *nativeDevice
	zone     *nativeZone
	resolver *nativeResolver
	keys     *nativeKeys
}

func newNatives() *natives {
	return &natives{records: newMemRecords(), clock: &nativeClock{gate: newGate()}, network: newNativeNetwork(),
		consent: &nativeConsent{gate: newGate(), value: "valid"}, device: &nativeDevice{gate: newGate(), state: NewDeviceState(80, true, true, true)},
		zone: &nativeZone{gate: newGate(), name: "Europe/Berlin"}, resolver: &nativeResolver{gate: newGate()}, keys: &nativeKeys{gate: newGate(), size: 32}}
}

func (n *natives) adapters() *Adapters {
	return &Adapters{Records: n.records, Clock: n.clock, Network: n.network, Consent: n.consent, Device: n.device,
		TimeZone: n.zone, Resolver: n.resolver, LinkageKeys: n.keys}
}

// openAll releases every gate so stalled callbacks finish before the test ends.
func (n *natives) openAll() {
	for _, g := range []*gate{n.records.gate, n.clock.gate, n.network.gate, n.network.watchGate, n.consent.gate, n.device.gate, n.zone.gate, n.resolver.gate, n.keys.gate} {
		g.open()
	}
}

// ---- Synthetic trusted local document

type target struct {
	ID    string
	Test  string
	Role  string
	Tuple map[string]any
}

// trustDocument permits test against targets for the deployment and the
// embedded issuer "embedded-issuer".
func trustDocument(test string, targets ...target) []byte {
	ids := []string{}
	regs := []any{}
	for _, t := range targets {
		ids = append(ids, t.ID)
		regs = append(regs, map[string]any{
			"id": t.ID, "test": map[string]any{"id": t.Test, "version": "0.1.0"}, "role": t.Role, "tuple": t.Tuple,
			"paths": map[string]any{"direct": true, "proxy": true}, "baseline_risk": "low", "owner": "synthetic", "purpose": "synthetic",
			"review_generation": 1, "expires_at": time.Now().Add(48 * time.Hour).UTC().Format("2006-01-02T15:04:05Z"), "redirect_ids": []string{},
		})
	}
	permissions := map[string]any{
		"tests": []any{map[string]any{"id": test, "version": "0.1.0"}}, "registration_ids": ids,
		"paths": map[string]any{"direct": true, "proxy": true}, "max_risk": "high",
		"require_consent": false, "allow_medium_risk": true, "allow_composite": false, "allow_unknown_risk": true,
		"label_namespaces": []string{},
		"limits": map[string]any{
			"attempt":  map[string]any{"duration_ms": 60000, "bytes_sent": 65536, "bytes_received": 262144, "connections": 12},
			"max_runs": 28, "max_attempts": 2, "max_validity_seconds": 604800,
			"min_repeat_interval_seconds": 21600, "min_repeat_jitter_seconds": 300,
			"min_retry_delay_ms": 30000, "max_retry_delay_ms": 600000, "max_retry_jitter_ms": 60000,
			"max_result_bytes": 65536, "max_bundle_bytes": 262144, "max_results_per_bundle": 16,
			"max_pending_bundles": 32, "max_pending_bytes": 4194304, "max_retention_seconds": 86400,
		},
	}
	doc := map[string]any{
		"format": "barometer-trust/1",
		"policy": map[string]any{
			"generation": 2, "permissions": permissions,
			"issuers": []any{map[string]any{"issuer_id": "embedded-issuer", "permissions": permissions}},
			"targets": regs, "keys": []any{},
		},
		"configuration_approvals": map[string]any{"endpoints": []string{}, "connection_codes": []string{}, "linkage_scopes": []string{}, "allow_linkage": false},
		"context_approvals":       map[string]any{"connection_codes": []string{}, "classification_versions": []string{}, "country_codes": []string{}, "region_codes": []string{}},
	}
	b, err := json.Marshal(doc)
	if err != nil {
		panic(err)
	}
	return b
}

var tcpTarget = target{ID: "reg-a", Test: "TCP-01", Role: "measurement", Tuple: map[string]any{"address": "8.8.8.8", "port": 443, "scheme": "tcp"}}

// newTestClient builds a facade client over n with the host-test platform.
func newTestClient(t *testing.T, n *natives, trust []byte, o options) *Client {
	t.Helper()
	if o.platform == "" {
		o.platform = "android"
	}
	c, err := newClient(n.adapters(), trust, o)
	if err != nil {
		t.Fatal("newClient:", err)
	}
	t.Cleanup(func() { n.openAll(); c.Close() })
	return c
}
