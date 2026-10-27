// Package capability owns synthetic consent, device, time-zone, network and
// linkage-key capabilities for tests only. Production code must not import it. Each fake
// reports a scripted value, counts its queries and reports no identifier; the
// network fake holds synthetic canaries it cannot report.
package capability

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/tv-rain/barometer-go/internal/platform"
)

// ErrUnavailable is the scripted capability failure.
var ErrUnavailable = errors.New("testkit/capability: unavailable")

// Consent reports a scripted status, or fails while Fail is set.
type Consent struct {
	mu     sync.Mutex
	status platform.ConsentStatus
	fail   bool
	calls  int
}

func NewConsent(s platform.ConsentStatus) *Consent { return &Consent{status: s} }

// Set changes the reported status and clears an injected failure.
func (c *Consent) Set(s platform.ConsentStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.status, c.fail = s, false
}

// Fail makes every later query fail.
func (c *Consent) Fail() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fail = true
}

func (c *Consent) Status(context.Context) (platform.ConsentStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls++
	if c.fail {
		return platform.ConsentUnavailable, ErrUnavailable
	}
	return c.status, nil
}

// Calls reports how many queries were made.
func (c *Consent) Calls() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls
}

// Device reports a scripted state, or fails while Fail is set.
type Device struct {
	mu    sync.Mutex
	state platform.DeviceState
	fail  bool
	calls int
}

func NewDevice(s platform.DeviceState) *Device { return &Device{state: s} }

// Set changes the reported state and clears an injected failure.
func (d *Device) Set(s platform.DeviceState) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.state, d.fail = s, false
}

// Fail makes every later query fail.
func (d *Device) Fail() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.fail = true
}

func (d *Device) State(context.Context) (platform.DeviceState, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.calls++
	if d.fail {
		return platform.DeviceState{}, ErrUnavailable
	}
	return d.state, nil
}

// Calls reports how many queries were made.
func (d *Device) Calls() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.calls
}

// Zone reports a fixed location; a nil location models an unknown zone.
type Zone struct{ Loc *time.Location }

func (z Zone) Location(context.Context) (*time.Location, error) {
	if z.Loc == nil {
		return nil, ErrUnavailable
	}
	return z.Loc, nil
}

// Canaries are synthetic values for every item a real platform can expose but
// access context must never carry (REQ-CONTEXT-0003, REQ-CONTEXT-FORMAT-0005):
// an interface name, local IPv4 and IPv6 addresses, a DNS server address, an
// SSID, a BSSID, a cell identifier, SIM information, a subscriber identifier,
// carrier account information and a physical location. Network holds them
// beside its access state; platform.AccessState has no field for any of them,
// so a test asserts that none appears in what the core produces.
var Canaries = []string{
	"canary-if-wlan7", "198.51.100.77", "2001:db8::77", "203.0.113.53",
	"canary-ssid-7f3a", "02:00:5e:10:00:77", "canary-cell-25001-7777",
	"canary-iccid-8970177", "canary-imsi-250017777", "canary-carrier-account-77",
	"55.7777,37.7777",
}

// Network reports a scripted access state, or fails while Fail is set. Set
// with a new generation notifies every watch on another generation once, on
// the caller's goroutine, outside the fake's lock.
type Network struct {
	mu       sync.Mutex
	state    platform.AccessState
	fail     bool
	calls    int
	next     int
	watchers map[int]networkWatch
	// Withheld are the forbidden platform values beside the state; nothing
	// reads them, because the port cannot carry them.
	Withheld []string
}

type networkWatch struct {
	generation uint64
	changed    func()
}

func NewNetwork(s platform.AccessState) *Network {
	return &Network{state: s, watchers: map[int]networkWatch{}, Withheld: Canaries}
}

// Set changes the reported state, clears an injected failure and delivers a
// generation change to the watches it affects.
func (n *Network) Set(s platform.AccessState) {
	n.mu.Lock()
	n.state, n.fail = s, false
	var fire []func()
	for id, w := range n.watchers {
		if w.generation != s.Generation {
			fire = append(fire, w.changed)
			delete(n.watchers, id)
		}
	}
	n.mu.Unlock()
	for _, f := range fire {
		f()
	}
}

// Fail makes every later query fail.
func (n *Network) Fail() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.fail = true
}

func (n *Network) Access(context.Context) (platform.AccessState, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.calls++
	if n.fail {
		return platform.AccessState{}, ErrUnavailable
	}
	return n.state, nil
}

func (n *Network) Watch(generation uint64, changed func()) func() {
	n.mu.Lock()
	if n.state.Generation != generation {
		n.mu.Unlock()
		changed()
		return func() {}
	}
	id := n.next
	n.next++
	n.watchers[id] = networkWatch{generation, changed}
	n.mu.Unlock()
	return func() {
		n.mu.Lock()
		defer n.mu.Unlock()
		delete(n.watchers, id)
	}
}

// Calls reports how many queries were made.
func (n *Network) Calls() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.calls
}

// Watching reports how many watches are registered and not yet stopped or
// delivered.
func (n *Network) Watching() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	return len(n.watchers)
}
