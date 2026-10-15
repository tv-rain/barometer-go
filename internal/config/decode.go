package config

import (
	"net/netip"
	"net/url"
	"strconv"
	"strings"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
)

// Finite implementation-local bounds for this versionless configuration model.
// requirement-ref: REQ-JSON-0009
const (
	MaxBytes         = 16384
	maxEndpointBytes = 2048
	maxApprovals     = 64
)

var limits = jsondecode.Limits{MaxBytes: MaxBytes, MaxDepth: 8, MaxNodes: 64}

func object(fields map[string]jsondecode.Field) *jsondecode.Schema {
	return &jsondecode.Schema{Kind: jsondecode.Object, Fields: fields}
}
func textSchema() *jsondecode.Schema { return &jsondecode.Schema{Kind: jsondecode.String} }
func required(s *jsondecode.Schema) jsondecode.Field {
	return jsondecode.Field{Schema: s, Required: true}
}
func optional(s *jsondecode.Schema) jsondecode.Field { return jsondecode.Field{Schema: s} }

var proxySchema = object(map[string]jsondecode.Field{
	"host": required(textSchema()), "port": required(&jsondecode.Schema{Kind: jsondecode.Integer, Min: "1", Max: "65535"}), "connection_code": required(textSchema()),
})
var networkSchema = object(map[string]jsondecode.Field{
	"primary_endpoint": required(textSchema()), "fallback_endpoint": optional(textSchema()),
})
var linkageSchema = object(map[string]jsondecode.Field{
	"mode": required(textSchema()), "scope": optional(textSchema()),
	"rotation_period_seconds": optional(&jsondecode.Schema{Kind: jsondecode.Integer, Min: "86400", Max: "604800"}),
})
var initSchema = object(map[string]jsondecode.Field{
	"proxy": optional(proxySchema), "network_context": optional(networkSchema), "result_linkage": optional(linkageSchema),
})

// NewDecoder pins the containing-model contract without inventing a version
// field. An unsupported contract never falls back to this schema.
// requirement-ref: REQ-JSON-0008
func NewDecoder(contract string, a Approvals) (*Decoder, error) {
	if contract != Contract {
		return nil, ErrVersion
	}
	for _, list := range [][]string{a.Endpoints, a.ConnectionCodes, a.LinkageScopes} {
		if len(list) > maxApprovals {
			return nil, ErrApproval
		}
	}
	for _, endpoint := range a.Endpoints {
		if !validEndpoint(endpoint) {
			return nil, ErrApproval
		}
	}
	for _, code := range a.ConnectionCodes {
		if !category(code, 64) {
			return nil, ErrApproval
		}
	}
	for _, scope := range a.LinkageScopes {
		if !category(scope, 128) {
			return nil, ErrApproval
		}
	}
	a.Endpoints = append([]string(nil), a.Endpoints...)
	a.ConnectionCodes = append([]string(nil), a.ConnectionCodes...)
	a.LinkageScopes = append([]string(nil), a.LinkageScopes...)
	return &Decoder{ready: true, approvals: a}, nil
}

// Decode publishes no usable snapshot until every section has validated.
// Trusted policy, keys, source context and capabilities have no wire fields.
// requirement-ref: REQ-JSON-0004
// requirement-ref: REQ-JSON-0010
func (d *Decoder) Decode(input []byte) (Snapshot, error) {
	if d == nil || !d.ready {
		return Snapshot{}, ErrUnavailable
	}
	v, err := jsondecode.Decode(input, initSchema, limits)
	if err != nil {
		return Snapshot{}, err
	}
	s := Snapshot{owner: d, linkage: Linkage{Mode: "none"}}
	if p, ok := v.Lookup("proxy"); ok {
		s.proxy, err = d.proxy(p)
		if err != nil {
			return Snapshot{}, err
		}
		s.hasProxy = true
	}
	if n, ok := v.Lookup("network_context"); ok {
		s.network = NetworkContext{PrimaryEndpoint: field(n, "primary_endpoint"), FallbackEndpoint: field(n, "fallback_endpoint")}
		for _, name := range []string{"primary_endpoint", "fallback_endpoint"} {
			if value, present := n.Lookup(name); present {
				if !validEndpoint(value.Text()) {
					return Snapshot{}, ErrValue
				}
				if !contains(d.approvals.Endpoints, value.Text()) {
					return Snapshot{}, ErrApproval
				}
			}
		}
		s.hasNetwork = true
	}
	if l, ok := v.Lookup("result_linkage"); ok {
		s.linkage.Mode = field(l, "mode")
		_, scopePresent := l.Lookup("scope")
		_, rotationPresent := l.Lookup("rotation_period_seconds")
		switch s.linkage.Mode {
		case "none":
			if scopePresent || rotationPresent {
				return Snapshot{}, ErrValue
			}
		case "rotating_installation":
			if !scopePresent || !rotationPresent {
				return Snapshot{}, ErrValue
			}
			s.linkage.Scope = field(l, "scope")
			if !category(s.linkage.Scope, 128) {
				return Snapshot{}, ErrValue
			}
			s.linkage.RotationPeriodSeconds, _ = strconv.Atoi(field(l, "rotation_period_seconds"))
			if !contains(d.approvals.LinkageScopes, s.linkage.Scope) {
				return Snapshot{}, ErrApproval
			}
			if !d.approvals.AllowLinkage {
				s.linkage = Linkage{Mode: "none"}
			}
		default:
			return Snapshot{}, ErrValue
		}
	}
	s.valid = true
	return s, nil
}

func (d *Decoder) decodeProxy(input []byte) (Proxy, error) {
	v, err := jsondecode.Decode(input, proxySchema, limits)
	if err != nil {
		return Proxy{}, err
	}
	return d.proxy(v)
}
func (d *Decoder) proxy(v *jsondecode.Value) (Proxy, error) {
	p := Proxy{Host: field(v, "host"), ConnectionCode: field(v, "connection_code")}
	p.Port, _ = strconv.Atoi(field(v, "port"))
	if p.Host != "localhost" {
		ip, err := netip.ParseAddr(p.Host)
		if err != nil || ip.Zone() != "" || !ip.IsLoopback() {
			return Proxy{}, ErrValue
		}
	}
	if !category(p.ConnectionCode, 64) {
		return Proxy{}, ErrValue
	}
	if !contains(d.approvals.ConnectionCodes, p.ConnectionCode) {
		return Proxy{}, ErrApproval
	}
	return p, nil
}
func field(v *jsondecode.Value, name string) string {
	x, ok := v.Lookup(name)
	if !ok {
		return ""
	}
	return x.Text()
}
func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
func category(s string, max int) bool {
	if len(s) == 0 || len(s) > max {
		return false
	}
	for _, c := range []byte(s) {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '.' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

// No resolution or network request occurs here. Exact reviewed URL membership
// handles deployment approval and non-user-specific paths/queries; runtime DNS
// and redirect controls remain the responsibility of the eventual transport.
func validEndpoint(raw string) bool {
	if len(raw) == 0 || len(raw) > maxEndpointBytes {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Opaque != "" || u.Host == "" || u.User != nil || strings.Contains(raw, "#") {
		return false
	}
	host := u.Hostname()
	if host == "" || strings.Contains(host, "%") || strings.HasSuffix(u.Host, ":") {
		return false
	}
	if p := u.Port(); p != "" {
		n, e := strconv.Atoi(p)
		if e != nil || n < 1 || n > 65535 {
			return false
		}
	}
	if ip, e := netip.ParseAddr(host); e == nil {
		return ip.IsGlobalUnicast() && !ip.IsPrivate() && !ip.IsLoopback() && !ip.IsLinkLocalUnicast()
	}
	// DNS names must be fully qualified and use an ordinary ASCII wire spelling.
	host = strings.TrimSuffix(strings.ToLower(host), ".")
	if host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") || !strings.Contains(host, ".") || len(host) > 253 {
		return false
	}
	for _, label := range strings.Split(host, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, c := range []byte(label) {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				return false
			}
		}
	}
	return true
}
