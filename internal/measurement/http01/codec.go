// Package http01 implements bounded Draft 0.1.0 HTTPS response validation.
package http01

import (
	"net/netip"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/taskdecode"
)

// Parameters contains immutable normalized input, never authority or a path grant.
type Parameters struct {
	tuple           task.TargetTuple
	family          string
	body, redirects int
}

func (Parameters) TaskParameters()  {}
func (Parameters) String() string   { return "http01.Parameters" }
func (Parameters) GoString() string { return "http01.Parameters" }
func (p Parameters) Target() task.TargetTuple {
	t := p.tuple
	t.BodyChecks = slices.Clone(t.BodyChecks)
	t.HeaderChecks = slices.Clone(t.HeaderChecks)
	return t
}
func (p Parameters) Limits() observation.ProtocolLimits {
	return observation.ProtocolLimits{MaxBodyBytes: p.body, MaxRedirects: p.redirects, StatusCheck: p.tuple.ExpectedStatus != 0, HeaderChecks: len(p.tuple.HeaderChecks), BodyChecks: len(p.tuple.BodyChecks)}
}
func (p Parameters) ValidateLimits(l task.AttemptLimits) bool { return p.body <= l.BytesReceived }

// Codec provides the closed public HTTP-01 input/options shape.
type Codec struct{}

// requirement-ref: REQ-HTTP-01-0010
// requirement-ref: REQ-HTTP-01-0011
// requirement-ref: REQ-HTTP-01-0013
// requirement-ref: REQ-HTTP-01-0022
// requirement-ref: REQ-HTTP-01-0024
func (Codec) Definition() taskdecode.Definition {
	str := func() *jsondecode.Schema { return &jsondecode.Schema{Kind: jsondecode.String} }
	obj := func(f map[string]jsondecode.Field) *jsondecode.Schema {
		return &jsondecode.Schema{Kind: jsondecode.Object, Fields: f}
	}
	return taskdecode.Definition{TestID: "HTTP-01", TestVersion: "0.1.0", Input: obj(map[string]jsondecode.Field{
		"url": {Required: true, Schema: str()}, "expected_status": {Schema: &jsondecode.Schema{Kind: jsondecode.Integer, Min: "100", Max: "599"}},
		"body_checks":   {Schema: &jsondecode.Schema{Kind: jsondecode.Array, MaxCount: 8, Limit: true, Element: obj(map[string]jsondecode.Field{"type": {Required: true, Schema: str()}, "value": {Required: true, Schema: str()}})}},
		"header_checks": {Schema: &jsondecode.Schema{Kind: jsondecode.Array, MaxCount: 8, Limit: true, Element: obj(map[string]jsondecode.Field{"name": {Required: true, Schema: str()}, "type": {Required: true, Schema: str()}, "value": {Schema: str()}})}},
	}), Options: obj(map[string]jsondecode.Field{"address_family": {Schema: str()}, "max_body_bytes": {Schema: &jsondecode.Schema{Kind: jsondecode.Integer, Min: "1", Max: "65536", Limit: true}}, "follow_redirects": {Schema: &jsondecode.Schema{Kind: jsondecode.Boolean}}, "max_redirects": {Schema: &jsondecode.Schema{Kind: jsondecode.Integer, Min: "1", Max: "2", Limit: true}}})}
}

// NormalizeURL performs only syntax and public-address checks. Registration
// and per-contact resolution/address checks remain with admission/execution.
// requirement-ref: REQ-HTTP-01-0015
// requirement-ref: REQ-HTTP-01-0016
// requirement-ref: REQ-HTTP-01-0017
func NormalizeURL(raw string) (task.TargetTuple, bool) {
	if len(raw) > 2048 || !utf8.ValidString(raw) || strings.ContainsAny(raw, "{}\r\n\t ") {
		return task.TargetTuple{}, false
	}
	u, e := url.Parse(raw)
	if e != nil || u.Scheme != "https" || u.User != nil || u.Fragment != "" || strings.Contains(raw, "#") || u.Opaque != "" || u.Host == "" || strings.HasSuffix(u.Host, ":") {
		return task.TargetTuple{}, false
	}
	host := u.Hostname()
	ip, ie := netip.ParseAddr(host)
	if ie == nil {
		if !task.PublicAddress(ip) || ip.Is4In6() || ip.Zone() != "" {
			return task.TargetTuple{}, false
		}
		host = ip.String()
	} else {
		var ok bool
		host, ok = taskdecode.DNSName(host)
		if !ok {
			return task.TargetTuple{}, false
		}
	}
	port := 443
	if u.Port() != "" {
		n, e := strconv.Atoi(u.Port())
		if e != nil || n < 1 || n > 65535 || strconv.Itoa(n) != u.Port() {
			return task.TargetTuple{}, false
		}
		port = n
	}
	u.Host = host
	if ie == nil && ip.Is6() {
		u.Host = "[" + host + "]"
	}
	if port != 443 {
		u.Host += ":" + strconv.Itoa(port)
	}
	if u.Path == "" {
		u.Path = "/"
	}
	if len(u.String()) > 2048 {
		return task.TargetTuple{}, false
	}
	return task.TargetTuple{URL: u.String(), Hostname: host, Scheme: "https", Port: uint16(port)}, true
}

// requirement-ref: REQ-HTTP-01-0020
// requirement-ref: REQ-HTTP-01-0021
// requirement-ref: REQ-HTTP-01-0025
func (c Codec) Normalize(in, opt *jsondecode.Value) (task.Parameters, error) {
	d := c.Definition()
	bad := func() (task.Parameters, error) { return nil, taskdecode.ErrParameters }
	if in == nil || opt == nil || jsondecode.Validate(in, d.Input) != nil || jsondecode.Validate(opt, d.Options) != nil {
		return bad()
	}
	v, _ := in.Lookup("url")
	if len(v.Text()) > 2048 {
		return nil, jsondecode.ErrLimit
	}
	t, ok := NormalizeURL(v.Text())
	if !ok {
		return bad()
	}
	p := Parameters{tuple: t, family: "auto", body: 4096}
	if v, ok := in.Lookup("expected_status"); ok {
		p.tuple.ExpectedStatus, _ = strconv.Atoi(v.Text())
	}
	if v, ok := in.Lookup("body_checks"); ok {
		for _, x := range v.Elements() {
			a, _ := x.Lookup("type")
			b, _ := x.Lookup("value")
			if (a.Text() == "contains" || a.Text() == "prefix" || a.Text() == "suffix") && len(b.Text()) > 256 {
				return nil, jsondecode.ErrLimit
			}
			p.tuple.BodyChecks = append(p.tuple.BodyChecks, task.BodyCheck{Type: a.Text(), Value: b.Text()})
		}
	}
	if v, ok := in.Lookup("header_checks"); ok {
		for _, x := range v.Elements() {
			a, _ := x.Lookup("name")
			b, _ := x.Lookup("type")
			val, has := x.Lookup("value")
			kind := b.Text()
			if (kind == "present" || kind == "absent") && has || (kind == "equals" || kind == "contains") && !has {
				return bad()
			}
			value := ""
			if has {
				value = val.Text()
			}
			if (kind == "equals" || kind == "contains") && len(value) > 256 {
				return nil, jsondecode.ErrLimit
			}
			p.tuple.HeaderChecks = append(p.tuple.HeaderChecks, task.HeaderCheck{Name: strings.ToLower(a.Text()), Type: kind, Value: value})
		}
	}
	// Validate the complete check tuple before registration matching.
	if !validChecks(p.tuple) {
		return bad()
	}
	if v, ok := opt.Lookup("address_family"); ok {
		p.family = v.Text()
	}
	if p.family != "auto" && p.family != "ipv4" && p.family != "ipv6" {
		return bad()
	}
	if ip, e := netip.ParseAddr(p.tuple.Hostname); e == nil && (p.family == "ipv4" && !ip.Is4() || p.family == "ipv6" && !ip.Is6()) {
		return bad()
	}
	if v, ok := opt.Lookup("max_body_bytes"); ok {
		p.body, _ = strconv.Atoi(v.Text())
	}
	follow, has := opt.Lookup("follow_redirects")
	r, hasR := opt.Lookup("max_redirects")
	if has && follow.Bool() {
		if !hasR {
			return bad()
		}
		p.redirects, _ = strconv.Atoi(r.Text())
	} else if hasR {
		return bad()
	}
	return p, nil
}
func validChecks(t task.TargetTuple) bool {
	for _, c := range t.BodyChecks {
		switch c.Type {
		case "contains", "prefix", "suffix":
			if len(c.Value) == 0 || len(c.Value) > 256 || !utf8.ValidString(c.Value) {
				return false
			}
		case "sha256":
			if len(c.Value) != 64 || strings.Trim(c.Value, "0123456789abcdef") != "" {
				return false
			}
		default:
			return false
		}
	}
	for _, c := range t.HeaderChecks {
		if len(c.Name) == 0 || len(c.Name) > 256 {
			return false
		}
		for _, b := range []byte(c.Name) {
			if !(b >= 'a' && b <= 'z' || b >= '0' && b <= '9' || strings.ContainsRune("!#$%&'*+-.^_`|~", rune(b))) {
				return false
			}
		}
		switch c.Name {
		case "authorization", "proxy-authorization", "www-authenticate", "proxy-authenticate", "cookie", "set-cookie", "authentication-info", "proxy-authentication-info":
			return false
		}
		switch c.Type {
		case "present", "absent":
			if c.Value != "" {
				return false
			}
		case "equals", "contains":
			if len(c.Value) == 0 || len(c.Value) > 256 || !utf8.ValidString(c.Value) || strings.ContainsAny(c.Value, "\r\n\x00") {
				return false
			}
		default:
			return false
		}
	}
	return true
}

type ProxyRule struct{}

// requirement-ref: REQ-HTTP-01-0051
func (ProxyRule) Compatible(params task.Parameters) bool {
	p, ok := params.(Parameters)
	if !ok || p.family != "auto" {
		return false
	}
	_, e := netip.ParseAddr(p.tuple.Hostname)
	return e == nil
}
