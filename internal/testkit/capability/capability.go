// Package capability owns synthetic consent, device and time-zone capabilities
// for tests only. Production code must not import it. Each fake reports a
// scripted value, counts its queries and carries no identifier.
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
