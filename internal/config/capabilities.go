package config

import (
	"reflect"

	"github.com/tv-rain/barometer-go/internal/platform"
)

// Capabilities has no JSON representation or registration path from a task.
// Native implementations and trusted provenance are separate reviewed work.
type Capabilities struct {
	Clock       platform.Clock
	LinkageKeys platform.LinkageKeyStore
}

// Initialization combines validated values and private capability references.
// It is a preparation result, not an initialized client or an admission permit.
// Delivery verification, local policy, replay and target authorization must still
// be installed and checked by their eventual runtime consumers.
type Initialization struct {
	config Snapshot
	clock  platform.Clock
	keys   platform.LinkageKeyStore
}

func Prepare(s Snapshot, c Capabilities) (*Initialization, error) {
	if !s.valid || missing(c.Clock) {
		return nil, ErrUnavailable
	}
	var keys platform.LinkageKeyStore
	if !missing(c.LinkageKeys) {
		keys = c.LinkageKeys
	}
	return &Initialization{config: s, clock: c.Clock, keys: keys}, nil
}
func (i *Initialization) Configuration() Snapshot { return i.config }
func (i *Initialization) Clock() platform.Clock   { return i.clock }

// Missing protected storage disables linkage alone. Runtime derivation failure
// must also omit linkage, never fail over to insecure key storage.
func (i *Initialization) LinkageKeys() (platform.LinkageKeyStore, bool) {
	if i.config.linkage.Mode != "rotating_installation" || i.keys == nil {
		return nil, false
	}
	return i.keys, true
}

func missing(v any) bool {
	if v == nil {
		return true
	}
	r := reflect.ValueOf(v)
	switch r.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return r.IsNil()
	}
	return false
}
