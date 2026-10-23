// Package networkcontext owns runtime Network Context acquisition. Since
// DEV-644 it captures coarse access context from the platform.Network
// capability and binds it to the direct path with egress unavailable. Egress
// lookup, its cache and the proxy path belong to later Network paths and
// context work (DEV-645, DEV-647). Context values are observations about the
// device network, never evidence of censorship.
package networkcontext

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/tv-rain/barometer-go/internal/model/observation"
	"github.com/tv-rain/barometer-go/internal/platform"
)

var (
	ErrCapability  = errors.New("networkcontext: invalid capability")
	ErrUnavailable = errors.New("networkcontext: access context unavailable")
)

// Direct captures access context for direct-path attempts. It holds no
// context between captures, so nothing it acquires is cached or retained.
type Direct struct {
	network platform.Network
	decoder *observation.ContextDecoder
}

// NewDirect binds the platform network capability. A direct context carries
// no classification, so its decoder needs no approvals.
func NewDirect(n platform.Network) (*Direct, error) {
	if n == nil {
		return nil, ErrCapability
	}
	d, err := observation.NewContextDecoder(observation.Contract, observation.ContextApprovals{})
	if err != nil {
		return nil, ErrCapability
	}
	return &Direct{network: n, decoder: d}, nil
}

// Decoder is the context decoder that results built from this source's
// snapshots bind to.
func (d *Direct) Decoder() *observation.ContextDecoder { return d.decoder }

// Capture reads the access context once and binds it to the direct path of
// its access-network generation. It sends no traffic (REQ-CONTEXT-0018). A
// capability error, a zero generation or an ended context is ErrUnavailable;
// nothing is guessed. Values the platform reports as unknown, and values this
// build does not recognize, stay unknown (REQ-CONTEXT-0034).
// requirement-ref: REQ-CONTEXT-0002
// requirement-ref: REQ-CONTEXT-0003
func (d *Direct) Capture(ctx context.Context) (Snapshot, error) {
	if d == nil || ctx.Err() != nil {
		return Snapshot{}, ErrUnavailable
	}
	s, err := d.network.Access(ctx)
	if err != nil || s.Generation == 0 || ctx.Err() != nil {
		return Snapshot{}, ErrUnavailable
	}
	path, err := observation.NewPath(s.Generation, 0, "direct", "")
	if err != nil {
		return Snapshot{}, ErrUnavailable
	}
	a := Normalize(s)
	snap := Snapshot{valid: true, network: d.network, access: a, path: path, generation: s.Generation}
	for i, changed := range []bool{false, true} {
		a.NetworkChanged = changed
		raw, err := json.Marshal(directWire{Access: accessWire(a), Egress: unavailableDirect})
		if err != nil {
			return Snapshot{}, ErrUnavailable
		}
		if snap.contexts[i], err = d.decoder.Decode(raw, path); err != nil {
			return Snapshot{}, ErrUnavailable
		}
	}
	return snap, nil
}

// Normalize maps a platform reading to the registered access vocabulary
// (REQ-CONTEXT-FORMAT-0004). network_changed is false: a reading is one
// generation, and only an attempt can observe it change.
// requirement-ref: REQ-CONTEXT-FORMAT-0004
func Normalize(s platform.AccessState) observation.Access {
	return observation.Access{
		NetworkType:   pick(int(s.Type), "unknown", "wifi", "cellular", "ethernet", "other"),
		VPNState:      pick(int(s.VPN), "unknown", "detected", "not_detected"),
		IPv4Available: availability(s.IPv4),
		IPv6Available: availability(s.IPv6),
		Validated:     availability(s.Validated),
	}
}

func availability(a platform.Availability) string { return pick(int(a), "unknown", "yes", "no") }

// pick returns names[i], or the first name, which is unknown, for any value
// outside the list.
func pick(i int, names ...string) string {
	if i < 0 || i >= len(names) {
		return names[0]
	}
	return names[i]
}

// Snapshot is one access reading bound to the direct path of its generation.
// Its generation is local metadata and never serialized (REQ-CONTEXT-FORMAT-0016).
type Snapshot struct {
	valid      bool
	network    platform.Network
	access     observation.Access
	path       observation.Path
	generation uint64
	contexts   [2]observation.NetworkContext // stable, then changed
}

// Access returns the normalized access values, with network_changed false.
func (s Snapshot) Access() observation.Access { return s.access }

// Path is the local path binding of the snapshot's generation.
func (s Snapshot) Path() observation.Path { return s.path }

// Generation is the access-network generation the snapshot was read in.
func (s Snapshot) Generation() uint64 { return s.generation }

// Context returns the Network Context to attach to a result: the snapshot's
// access values, network_changed as given, and egress unavailable on the
// direct path until egress lookup exists (DEV-647).
func (s Snapshot) Context(changed bool) (observation.NetworkContext, error) {
	if !s.valid {
		return observation.NetworkContext{}, ErrUnavailable
	}
	if changed {
		return s.contexts[1], nil
	}
	return s.contexts[0], nil
}

// Watch calls changed at most once when the access-network generation is no
// longer the snapshot's, immediately if it already differs; stop ends the
// watch. Without a valid snapshot or a stop function from the capability, it
// reports the change at once, so no attempt runs unwatched.
func (s Snapshot) Watch(changed func()) (stop func()) {
	if !s.valid || changed == nil {
		if changed != nil {
			changed()
		}
		return func() {}
	}
	if stop = s.network.Watch(s.generation, changed); stop == nil {
		changed()
		return func() {}
	}
	return stop
}

type directWire struct {
	Access accessWire `json:"access"`
	Egress egressWire `json:"egress"`
}

type accessWire struct {
	NetworkType    string `json:"network_type"`
	VPNState       string `json:"vpn_state"`
	IPv4Available  string `json:"ipv4_available"`
	IPv6Available  string `json:"ipv6_available"`
	Validated      string `json:"validated"`
	NetworkChanged bool   `json:"network_changed"`
}

type egressWire struct {
	Status       string `json:"status"`
	LookupSource string `json:"lookup_source"`
	NetworkPath  string `json:"network_path"`
}

var unavailableDirect = egressWire{Status: "unavailable", LookupSource: "none", NetworkPath: "direct"}
