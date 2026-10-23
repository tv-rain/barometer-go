// Package networkcontext owns runtime Network Context acquisition. It
// captures coarse access context from the platform.Network capability and
// binds it to the attempt's network path, direct or the host-configured
// proxy, with egress unavailable. Egress lookup and its cache belong to later
// Network paths and context work (DEV-647). Context values are observations
// about the device network, never evidence of censorship.
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

// Source captures access context for the attempts of one network path. It
// holds no context between captures, so nothing it acquires is cached or
// retained. A proxy source carries the proxy generation and connection_code
// of the snapshot it was built for; neither is task input, and only
// connection_code reaches a result (REQ-PROXY-0031, REQ-PROXY-0033).
type Source struct {
	network platform.Network
	decoder *observation.ContextDecoder
	proxy   uint64 // proxy generation; zero on the direct path
	code    string // connection_code; empty on the direct path
}

// NewDirect binds the platform network capability to the direct path. A
// direct context carries no classification or category, so its decoder needs
// no approvals.
func NewDirect(n platform.Network) (*Source, error) {
	return newSource(n, 0, "")
}

// NewProxy binds the platform network capability to one proxy snapshot: its
// non-zero generation and the host-configured connection_code, which the
// host configuration has already validated against its reviewed approvals.
func NewProxy(n platform.Network, generation uint64, connectionCode string) (*Source, error) {
	if generation == 0 || connectionCode == "" {
		return nil, ErrCapability
	}
	return newSource(n, generation, connectionCode)
}

func newSource(n platform.Network, generation uint64, code string) (*Source, error) {
	if n == nil {
		return nil, ErrCapability
	}
	var approvals observation.ContextApprovals
	if code != "" {
		approvals.ConnectionCodes = []string{code}
	}
	d, err := observation.NewContextDecoder(observation.Contract, approvals)
	if err != nil {
		return nil, ErrCapability
	}
	if _, err := observation.NewPath(1, generation, pathName(code), code); err != nil {
		return nil, ErrCapability
	}
	return &Source{network: n, decoder: d, proxy: generation, code: code}, nil
}

func pathName(code string) string {
	if code == "" {
		return "direct"
	}
	return "proxy"
}

// Decoder is the context decoder that results built from this source's
// snapshots bind to.
func (d *Source) Decoder() *observation.ContextDecoder { return d.decoder }

// Capture reads the access context once and binds it to the source's path in
// its access-network generation. It sends no traffic (REQ-CONTEXT-0018). A
// capability error, a zero generation or an ended context is ErrUnavailable;
// nothing is guessed. Values the platform reports as unknown, and values this
// build does not recognize, stay unknown (REQ-CONTEXT-0034).
// requirement-ref: REQ-CONTEXT-0002
// requirement-ref: REQ-CONTEXT-0003
func (d *Source) Capture(ctx context.Context) (Snapshot, error) {
	if d == nil || ctx.Err() != nil {
		return Snapshot{}, ErrUnavailable
	}
	s, err := d.network.Access(ctx)
	if err != nil || s.Generation == 0 || ctx.Err() != nil {
		return Snapshot{}, ErrUnavailable
	}
	path, err := observation.NewPath(s.Generation, d.proxy, pathName(d.code), d.code)
	if err != nil {
		return Snapshot{}, ErrUnavailable
	}
	a := Normalize(s)
	snap := Snapshot{valid: true, network: d.network, access: a, path: path, generation: s.Generation}
	egress := egressWire{Status: "unavailable", LookupSource: "none", NetworkPath: pathName(d.code), ConnectionCode: d.code}
	for i, changed := range []bool{false, true} {
		a.NetworkChanged = changed
		raw, err := json.Marshal(contextWire{Access: accessWire(a), Egress: egress})
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

// Snapshot is one access reading bound to its source's path in its generation.
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
// snapshot's path, with its connection_code on the proxy path, until egress
// lookup exists (DEV-647).
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

type contextWire struct {
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
	Status         string `json:"status"`
	LookupSource   string `json:"lookup_source"`
	NetworkPath    string `json:"network_path"`
	ConnectionCode string `json:"connection_code,omitempty"`
}
