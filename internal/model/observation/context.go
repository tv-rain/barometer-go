package observation

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
)

// ContextApprovals are trusted reviewed classification/proxy categories, never
// endpoint-supplied policy. RegionCodes must be verified ISO 3166-2 assignments
// by the integration owner; this leaf model validates syntax and membership.
// Each list has an implementation-local maximum of 256 entries.
type ContextApprovals struct {
	ConnectionCodes        []string
	ClassificationVersions []string
	CountryCodes           []string
	RegionCodes            []string
}

// Path is local binding metadata. It cannot be imported/exported as JSON.
// Its generations identify an initialized runtime's snapshots, not devices.
type Path struct {
	network, proxy uint64
	name, code     string
}

func NewPath(networkGeneration, proxyGeneration uint64, name, connectionCode string) (Path, error) {
	if networkGeneration == 0 {
		return Path{}, ErrValue
	}
	switch name {
	case "direct":
		if proxyGeneration != 0 || connectionCode != "" {
			return Path{}, ErrValue
		}
	case "proxy":
		if proxyGeneration == 0 || !category(connectionCode) {
			return Path{}, ErrValue
		}
	default:
		return Path{}, ErrValue
	}
	return Path{networkGeneration, proxyGeneration, name, connectionCode}, nil
}
func (Path) MarshalJSON() ([]byte, error) { return nil, ErrImport }
func (*Path) UnmarshalJSON([]byte) error  { return ErrImport }

type Access struct {
	NetworkType    string `json:"network_type"`
	VPNState       string `json:"vpn_state"`
	IPv4Available  string `json:"ipv4_available"`
	IPv6Available  string `json:"ipv6_available"`
	Validated      string `json:"validated"`
	NetworkChanged bool   `json:"network_changed"`
}
type Egress struct {
	Status                string  `json:"status"`
	LookupSource          string  `json:"lookup_source"`
	NetworkPath           string  `json:"network_path"`
	ConnectionCode        string  `json:"connection_code,omitempty"`
	CountryCode           string  `json:"country_code,omitempty"`
	RegionCode            string  `json:"region_code,omitempty"`
	ASN                   *uint32 `json:"asn,omitempty"`
	ClassificationVersion string  `json:"classification_version,omitempty"`
}

// NetworkContext stores a minimized immutable value and its non-serialized
// binding. JSON cannot supply the generation or detach an existing snapshot.
type NetworkContext struct {
	valid  bool
	access Access
	egress Egress
	path   Path
}
type ContextDecoder struct {
	ready     bool
	approvals ContextApprovals
}

// AccessSnapshot permits the access-only NET-01 representation before egress
// path selection. Its access generation is local metadata, never a wire field.
type AccessSnapshot struct {
	valid      bool
	value      Access
	generation uint64
}

func DecodeAccess(contract string, input []byte, networkGeneration uint64) (AccessSnapshot, error) {
	if contract != Contract {
		return AccessSnapshot{}, ErrVersion
	}
	if networkGeneration == 0 {
		return AccessSnapshot{}, ErrValue
	}
	if _, err := jsondecode.Decode(input, accessSchema, jsondecode.Limits{MaxBytes: 1024, MaxDepth: 2, MaxNodes: 16}); err != nil {
		return AccessSnapshot{}, err
	}
	var a Access
	if json.Unmarshal(input, &a) != nil || !validAccess(a) {
		return AccessSnapshot{}, ErrValue
	}
	return AccessSnapshot{true, a, networkGeneration}, nil
}
func (a AccessSnapshot) Value() Access { return a.value }
func (a AccessSnapshot) MatchesGeneration(g uint64) bool {
	return a.valid && !a.value.NetworkChanged && a.generation == g
}
func (a AccessSnapshot) MarshalJSON() ([]byte, error) {
	if !a.valid {
		return nil, ErrUnavailable
	}
	return json.Marshal(accessWire(a.value))
}
func (*AccessSnapshot) UnmarshalJSON([]byte) error { return ErrImport }

func (n NetworkContext) AccessSnapshot() AccessSnapshot {
	return AccessSnapshot{n.valid, n.access, n.path.network}
}

func NewContextDecoder(contract string, a ContextApprovals) (*ContextDecoder, error) {
	if contract != Contract {
		return nil, ErrVersion
	}
	for _, list := range [][]string{a.ConnectionCodes, a.ClassificationVersions, a.CountryCodes, a.RegionCodes} {
		if len(list) > 256 {
			return nil, ErrValue
		}
		seen := map[string]bool{}
		for _, s := range list {
			if seen[s] {
				return nil, ErrValue
			}
			seen[s] = true
		}
	}
	for _, s := range a.ConnectionCodes {
		if !category(s) {
			return nil, ErrValue
		}
	}
	for _, s := range a.ClassificationVersions {
		if !asciiToken(s, 64) {
			return nil, ErrValue
		}
	}
	for _, s := range a.CountryCodes {
		if !countrySyntax(s) {
			return nil, ErrValue
		}
	}
	for _, s := range a.RegionCodes {
		if !regionSyntax(s) || !slices.Contains(a.CountryCodes, s[:2]) {
			return nil, ErrValue
		}
	}
	a.ConnectionCodes = slices.Clone(a.ConnectionCodes)
	a.ClassificationVersions = slices.Clone(a.ClassificationVersions)
	a.CountryCodes = slices.Clone(a.CountryCodes)
	a.RegionCodes = slices.Clone(a.RegionCodes)
	return &ContextDecoder{true, a}, nil
}

// Decode binds normalized context to trusted local path metadata atomically.
// ISO assignments, non-user-specific versions and proxy categories must have
// been reviewed by the integration owner before supplying approvals.
// requirement-ref: REQ-CONTEXT-FORMAT-0001
// requirement-ref: REQ-CONTEXT-FORMAT-0017
// requirement-ref: REQ-CONTEXT-FORMAT-0023
func (d *ContextDecoder) Decode(input []byte, path Path) (NetworkContext, error) {
	if d == nil || !d.ready {
		return NetworkContext{}, ErrUnavailable
	}
	if _, err := NewPath(path.network, path.proxy, path.name, path.code); err != nil {
		return NetworkContext{}, err
	}
	v, err := jsondecode.Decode(input, contextSchema, jsondecode.Limits{MaxBytes: 4096, MaxDepth: 5, MaxNodes: 64})
	if err != nil {
		return NetworkContext{}, err
	}
	var wire struct {
		Access Access `json:"access"`
		Egress Egress `json:"egress"`
	}
	if json.Unmarshal(input, &wire) != nil {
		return NetworkContext{}, ErrValue
	}
	a, e := wire.Access, wire.Egress
	if !validAccess(a) || e.NetworkPath != path.name {
		return NetworkContext{}, ErrValue
	}
	eValue, _ := v.Lookup("egress")
	_, codePresent := eValue.Lookup("connection_code")
	if path.name == "direct" && codePresent || path.name == "proxy" && (!codePresent || e.ConnectionCode != path.code || !slices.Contains(d.approvals.ConnectionCodes, e.ConnectionCode)) {
		return NetworkContext{}, ErrValue
	}
	switch e.Status {
	case "unavailable":
		if e.LookupSource != "none" {
			return NetworkContext{}, ErrValue
		}
		for _, name := range []string{"country_code", "region_code", "asn", "classification_version"} {
			if _, ok := eValue.Lookup(name); ok {
				return NetworkContext{}, ErrValue
			}
		}
	case "fresh", "cached":
		if e.Status == "fresh" && !oneOf(e.LookupSource, "primary", "fallback") || e.Status == "cached" && e.LookupSource != "cache" {
			return NetworkContext{}, ErrValue
		}
		if !slices.Contains(d.approvals.ClassificationVersions, e.ClassificationVersion) {
			return NetworkContext{}, ErrValue
		}
		if _, ok := eValue.Lookup("country_code"); ok {
			// ISO codes use ASCII. Reject other scripts before case conversion
			// can map them onto an approved ASCII assignment (for example ſ/S).
			// requirement-ref: REQ-CONTEXT-FORMAT-0006
			if !asciiToken(e.CountryCode, 2) {
				return NetworkContext{}, ErrValue
			}
			e.CountryCode = strings.ToUpper(e.CountryCode)
			if !slices.Contains(d.approvals.CountryCodes, e.CountryCode) {
				return NetworkContext{}, ErrValue
			}
		}
		if _, ok := eValue.Lookup("region_code"); ok {
			if !asciiToken(e.RegionCode, 6) {
				return NetworkContext{}, ErrValue
			}
			e.RegionCode = strings.ToUpper(e.RegionCode)
			if !slices.Contains(d.approvals.RegionCodes, e.RegionCode) || e.CountryCode == "" || !strings.HasPrefix(e.RegionCode, e.CountryCode+"-") {
				return NetworkContext{}, ErrValue
			}
		}
	default:
		return NetworkContext{}, ErrValue
	}
	return NetworkContext{true, a, e, path}, nil
}

func validAccess(a Access) bool {
	return oneOf(a.NetworkType, "wifi", "cellular", "ethernet", "other", "unknown") && oneOf(a.VPNState, "detected", "not_detected", "unknown") && oneOf(a.IPv4Available, "yes", "no", "unknown") && oneOf(a.IPv6Available, "yes", "no", "unknown") && oneOf(a.Validated, "yes", "no", "unknown")
}
func (n NetworkContext) Access() Access { return n.access }
func (n NetworkContext) Egress() Egress {
	e := n.egress
	if e.ASN != nil {
		v := *e.ASN
		e.ASN = &v
	}
	return e
}

// Matches is a pure binding check. Runtime generation capture, cache expiry and
// cancellation are outside this model and must be enforced by their owners.
// requirement-ref: REQ-CONTEXT-0017
func (n NetworkContext) Matches(path Path) bool {
	return n.valid && !n.access.NetworkChanged && n.path == path
}
func (n NetworkContext) MarshalJSON() ([]byte, error) {
	if !n.valid {
		return nil, ErrUnavailable
	}
	return json.Marshal(struct {
		Access accessWire `json:"access"`
		Egress egressWire `json:"egress"`
	}{accessWire(n.access), egressWire(n.egress)})
}
func (*NetworkContext) UnmarshalJSON([]byte) error { return ErrImport }

type accessWire Access
type egressWire Egress

func (Access) MarshalJSON() ([]byte, error) { return nil, ErrImport }
func (Egress) MarshalJSON() ([]byte, error) { return nil, ErrImport }

func category(s string) bool {
	if len(s) == 0 || len(s) > 64 {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_' || c == '-' || c == '.') {
			return false
		}
	}
	return true
}
func asciiToken(s string, max int) bool {
	if len(s) == 0 || len(s) > max {
		return false
	}
	for _, c := range s {
		if c < 33 || c > 126 {
			return false
		}
	}
	return true
}
func countrySyntax(s string) bool {
	return len(s) == 2 && s[0] >= 'A' && s[0] <= 'Z' && s[1] >= 'A' && s[1] <= 'Z'
}
func regionSyntax(s string) bool {
	if len(s) < 4 || len(s) > 6 || !countrySyntax(s[:2]) || s[2] != '-' {
		return false
	}
	for _, c := range s[3:] {
		if !(c >= 'A' && c <= 'Z' || c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}

var accessSchema = object(map[string]jsondecode.Field{
	"network_type": required(textSchema()), "vpn_state": required(textSchema()), "ipv4_available": required(textSchema()), "ipv6_available": required(textSchema()), "validated": required(textSchema()), "network_changed": required(boolean()),
})
var egressSchema = object(map[string]jsondecode.Field{
	"status": required(textSchema()), "lookup_source": required(textSchema()), "network_path": required(textSchema()), "connection_code": optional(textSchema()), "country_code": optional(textSchema()), "region_code": optional(textSchema()), "asn": optional(&jsondecode.Schema{Kind: jsondecode.Integer, Min: "1", Max: "4294967295"}), "classification_version": optional(textSchema()),
})
var contextSchema = object(map[string]jsondecode.Field{"access": required(accessSchema), "egress": required(egressSchema)})
