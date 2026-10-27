// Package networkcontext owns runtime Network Context acquisition. It
// captures coarse access context from the platform.Network capability and
// binds it to the attempt's network path, direct or the host-configured
// proxy. When the host configures egress classification, an attempt acquires
// egress context through its own path dialer, using an in-memory cache
// (Lookup, Cache). Context values and endpoint failures are observations
// about the device network, never evidence of censorship (REQ-CONTEXT-0042).
package networkcontext

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tv-rain/barometer-go/internal/model/observation"
	"github.com/tv-rain/barometer-go/internal/platform"
)

var (
	ErrCapability  = errors.New("networkcontext: invalid capability")
	ErrUnavailable = errors.New("networkcontext: access context unavailable")
	// ErrNoNetworkState is the ErrUnavailable a capture returns when the
	// platform reports no supported network-state API, which NET-01 records
	// as platform_unavailable. errors.Is matches it to ErrUnavailable.
	ErrNoNetworkState = fmt.Errorf("%w: no network-state API", ErrUnavailable)
)

// Source captures access context for the attempts of one network path. It
// holds no context between captures; egress classifications are cached only
// in its lookup's Cache, never in the source. A proxy source carries the proxy generation and connection_code
// of the snapshot it was built for; neither is task input, and only
// connection_code reaches a result (REQ-PROXY-0031, REQ-PROXY-0033).
type Source struct {
	network platform.Network
	decoder *observation.ContextDecoder
	lookup  *Lookup // nil when the host configured no egress classification
	proxy   uint64  // proxy generation; zero on the direct path
	code    string  // connection_code; empty on the direct path
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
	if _, err := observation.NewPath(1, generation, pathName(code), code); err != nil {
		return nil, ErrCapability
	}
	s := &Source{network: n, proxy: generation, code: code}
	return s, s.bindDecoder()
}

// WithLookup returns a copy of the source whose attempts acquire egress
// context through l; a nil l leaves egress unavailable. Its decoder also
// accepts the classifications l's reviewed approvals permit.
func (d *Source) WithLookup(l *Lookup) (*Source, error) {
	if d == nil {
		return nil, ErrCapability
	}
	s := *d
	s.lookup = l
	return &s, s.bindDecoder()
}

func (d *Source) bindDecoder() error {
	var approvals observation.ContextApprovals
	if d.lookup != nil {
		approvals = d.lookup.approvals
	}
	if d.code != "" {
		approvals.ConnectionCodes = []string{d.code}
	}
	dec, err := observation.NewContextDecoder(observation.Contract, approvals)
	if err != nil {
		return ErrCapability
	}
	d.decoder = dec
	return nil
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
// capability error, a zero generation or an ended context is ErrUnavailable,
// and ErrNoNetworkState when the platform has no network-state API; nothing
// is guessed. Values the platform reports as unknown, and values this
// build does not recognize, stay unknown (REQ-CONTEXT-0034).
// requirement-ref: REQ-CONTEXT-0002
// requirement-ref: REQ-CONTEXT-0003
func (d *Source) Capture(ctx context.Context) (Snapshot, error) {
	if d == nil || ctx.Err() != nil {
		return Snapshot{}, ErrUnavailable
	}
	s, err := d.network.Access(ctx)
	if err != nil && ctx.Err() == nil && errors.Is(err, platform.ErrNetworkUnavailable) {
		return Snapshot{}, ErrNoNetworkState
	}
	if err != nil || s.Generation == 0 || ctx.Err() != nil {
		return Snapshot{}, ErrUnavailable
	}
	path, err := observation.NewPath(s.Generation, d.proxy, pathName(d.code), d.code)
	if err != nil {
		return Snapshot{}, ErrUnavailable
	}
	snap := Snapshot{valid: true, network: d.network, source: d, access: Normalize(s), path: path, generation: s.Generation}
	if snap.bind(egressWire{Status: "unavailable", LookupSource: "none"}) != nil {
		return Snapshot{}, ErrUnavailable
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
	source     *Source
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

// bind builds the stable and changed contexts of the snapshot with egress e on
// the snapshot's path, with its connection_code on the proxy path.
func (s *Snapshot) bind(e egressWire) error {
	e.NetworkPath, e.ConnectionCode = pathName(s.source.code), s.source.code
	a := s.access
	for i, changed := range []bool{false, true} {
		a.NetworkChanged = changed
		raw, err := json.Marshal(contextWire{Access: accessWire(a), Egress: e})
		if err != nil {
			return ErrUnavailable
		}
		if s.contexts[i], err = s.source.decoder.Decode(raw, s.path); err != nil {
			return ErrUnavailable
		}
	}
	return nil
}

// Acquire returns the snapshot with its egress context acquired through
// fetch, which must be bound to a dialer of the snapshot's own path and generation: the
// direct path, or the proxy generation's CONNECT dialer, never the other
// (REQ-CONTEXT-0017, REQ-PROXY-0023). The cache is keyed by the snapshot's
// access-network generation, path and proxy generation (REQ-CONTEXT-0020,
// REQ-PROXY-0024). Without a lookup, or when no endpoint produces a valid,
// approved classification, egress is unavailable and that is not an error
// (REQ-CONTEXT-0027). A cached classification is status cached from the
// cache; a new one is fresh from the primary or fallback
// (REQ-CONTEXT-FORMAT-0008). Acquisition ends with ctx, so it stays inside
// the attempt deadline and cancellation (REQ-CONTEXT-0025).
// requirement-ref: REQ-CONTEXT-0017
// requirement-ref: REQ-CONTEXT-0019
// requirement-ref: REQ-CONTEXT-0025
// requirement-ref: REQ-CONTEXT-FORMAT-0007
// requirement-ref: REQ-CONTEXT-FORMAT-0008
// requirement-ref: REQ-CONTEXT-FORMAT-0009
// requirement-ref: REQ-PROXY-0045
func (s Snapshot) Acquire(ctx context.Context, fetch Fetch) (Snapshot, error) {
	if !s.valid {
		return Snapshot{}, ErrUnavailable
	}
	if s.source.lookup == nil || fetch == nil {
		return s, nil
	}
	key := cacheKey{network: s.generation, proxy: s.source.proxy, path: pathName(s.source.code)}
	// A classification the reviewed approvals reject is an invalid response,
	// so the fallback is tried; nothing unapproved reaches a result.
	valid := func(c classification) bool {
		probe := s
		return probe.bind(classified(c, "fresh", "primary")) == nil
	}
	c, source, ok := s.source.lookup.acquire(ctx, key, fetch, valid)
	if !ok {
		return s, nil
	}
	status := "fresh"
	if source == "cache" {
		status = "cached"
	}
	out := s
	if out.bind(classified(c, status, source)) != nil {
		return s, nil
	}
	return out, nil
}

func classified(c classification, status, source string) egressWire {
	return egressWire{Status: status, LookupSource: source, CountryCode: c.country, RegionCode: c.region, ASN: c.asn, ClassificationVersion: c.version}
}

// Acquires reports whether the snapshot's source has a lookup, so Acquire
// may contact an endpoint; without one Acquire needs no fetch.
func (s Snapshot) Acquires() bool { return s.valid && s.source.lookup != nil }

// Egress returns the egress values the snapshot's contexts carry.
func (s Snapshot) Egress() observation.Egress { return s.contexts[0].Egress() }

// Context returns the Network Context to attach to a result: the snapshot's
// access values, network_changed as given, and its egress on the snapshot's
// path, with its connection_code on the proxy path.
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
	Status                string  `json:"status"`
	LookupSource          string  `json:"lookup_source"`
	NetworkPath           string  `json:"network_path"`
	ConnectionCode        string  `json:"connection_code,omitempty"`
	CountryCode           string  `json:"country_code,omitempty"`
	RegionCode            string  `json:"region_code,omitempty"`
	ASN                   *uint32 `json:"asn,omitempty"`
	ClassificationVersion string  `json:"classification_version,omitempty"`
}
