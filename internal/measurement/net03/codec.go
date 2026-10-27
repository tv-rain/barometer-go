// Package net03 implements the Draft 0.1.0 network validation and captive
// portal test. It reads the platform validation state once and performs one
// bounded validation request that never follows a redirect, keeping the two
// sources of evidence separate.
package net03

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

// PrefixBytes is the registered validation response prefix (REQ-NET-03-0021,
// REQ-LIMIT-0016). No stricter registered profile exists in this build, so the
// 4,096-byte profile maximum is never selected.
const PrefixBytes = 1024

// Parameters contains immutable normalized input, never authority or a path
// grant. The tuple carries the URL, scheme, expected status and optional body
// digest, so registration matches the complete expected response profile.
type Parameters struct {
	tuple  task.TargetTuple
	family string
}

func (Parameters) TaskParameters()  {}
func (Parameters) String() string   { return "net03.Parameters" }
func (Parameters) GoString() string { return "net03.Parameters" }
func (p Parameters) Target() task.TargetTuple {
	t := p.tuple
	t.BodyChecks = slices.Clone(t.BodyChecks)
	return t
}
func (p Parameters) AddressFamily() string { return p.family }

// HTTPS reports whether the validation URL is authenticated with TLS.
func (p Parameters) HTTPS() bool { return p.tuple.Scheme == "https" }
func (p Parameters) Limits() observation.ProtocolLimits {
	return observation.ProtocolLimits{MaxBodyBytes: PrefixBytes, StatusCheck: true, BodyChecks: len(p.tuple.BodyChecks)}
}
func (p Parameters) ValidateLimits(l task.AttemptLimits) bool { return PrefixBytes <= l.BytesReceived }

// Codec provides the closed public NET-03 input and options shape. Redirect
// following, custom headers, credentials and request bodies have no field, so
// a task naming one is rejected as an unknown field (REQ-NET-03-0016).
type Codec struct{}

// requirement-ref: REQ-NET-03-0009
// requirement-ref: REQ-NET-03-0013
// requirement-ref: REQ-NET-03-0015
// requirement-ref: REQ-NET-03-0016
func (Codec) Definition() taskdecode.Definition {
	str := func() *jsondecode.Schema { return &jsondecode.Schema{Kind: jsondecode.String} }
	obj := func(f map[string]jsondecode.Field) *jsondecode.Schema {
		return &jsondecode.Schema{Kind: jsondecode.Object, Fields: f}
	}
	return taskdecode.Definition{TestID: "NET-03", TestVersion: "0.1.0",
		Input: obj(map[string]jsondecode.Field{
			"url":                  {Required: true, Schema: str()},
			"expected_status":      {Required: true, Schema: &jsondecode.Schema{Kind: jsondecode.Integer, Min: "200", Max: "599"}},
			"expected_body_sha256": {Schema: str()},
		}),
		Options: obj(map[string]jsondecode.Field{"address_family": {Schema: str()}})}
}

// NormalizeURL performs only syntax and public-address checks on an http or
// https validation URL. Registration, resolution and per-contact address
// checks remain with admission and execution.
// requirement-ref: REQ-NET-03-0011
func NormalizeURL(raw string) (task.TargetTuple, bool) {
	if len(raw) > 2048 || !utf8.ValidString(raw) || strings.ContainsAny(raw, "{}\r\n\t ") {
		return task.TargetTuple{}, false
	}
	u, e := url.Parse(raw)
	if e != nil || u.Scheme != "https" && u.Scheme != "http" || u.User != nil || u.Fragment != "" || strings.Contains(raw, "#") || u.Opaque != "" || u.Host == "" || strings.HasSuffix(u.Host, ":") {
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
	if u.Scheme == "http" {
		port = 80
	}
	defaultPort := port
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
	if port != defaultPort {
		u.Host += ":" + strconv.Itoa(port)
	}
	if u.Path == "" {
		u.Path = "/"
	}
	if len(u.String()) > 2048 {
		return task.TargetTuple{}, false
	}
	return task.TargetTuple{URL: u.String(), Hostname: host, Scheme: u.Scheme, Port: uint16(port)}, true
}

// redirectStatus reports a status that is always classified redirect.
func redirectStatus(status int) bool {
	return status == 301 || status == 302 || status == 303 || status == 307 || status == 308
}

// requirement-ref: REQ-NET-03-0010
// requirement-ref: REQ-NET-03-0014
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
	s, _ := in.Lookup("expected_status")
	t.ExpectedStatus, _ = strconv.Atoi(s.Text())
	if redirectStatus(t.ExpectedStatus) {
		return bad()
	}
	if v, ok := in.Lookup("expected_body_sha256"); ok {
		if len(v.Text()) != 64 || strings.Trim(v.Text(), "0123456789abcdef") != "" {
			return bad()
		}
		t.BodyChecks = []task.BodyCheck{{Type: "sha256", Value: v.Text()}}
	}
	p := Parameters{tuple: t, family: "auto"}
	if v, ok := opt.Lookup("address_family"); ok {
		p.family = v.Text()
	}
	if p.family != "auto" && p.family != "ipv4" && p.family != "ipv6" {
		return bad()
	}
	if ip, e := netip.ParseAddr(p.tuple.Hostname); e == nil && (p.family == "ipv4" && !ip.Is4() || p.family == "ipv6" && !ip.Is6()) {
		return bad()
	}
	return p, nil
}
