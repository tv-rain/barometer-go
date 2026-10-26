// Package dns01 implements the Draft 0.1.0 bounded DNS measurement.
package dns01

import (
	"net/netip"
	"strconv"
	"strings"
	"time"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/taskdecode"
	"golang.org/x/net/idna"
)

// Parameters is immutable and contains no authority grant.
type Parameters struct {
	name, query, mode, transport string
	address                      netip.Addr
	port                         uint16
	collect                      bool
	window                       time.Duration
}

func (Parameters) TaskParameters()     {}
func (Parameters) String() string      { return "dns01.Parameters" }
func (Parameters) GoString() string    { return "dns01.Parameters" }
func (p Parameters) Name() string      { return p.name }
func (p Parameters) QueryType() string { return p.query }
func (p Parameters) Engine() string {
	if p.mode == "system" {
		return "system"
	}
	return p.transport
}
func (p Parameters) Resolver() (netip.Addr, uint16) { return p.address, p.port }

// Explicit nontransitional UTS #46 lookup mapping with STD3, context, bidi and
// DNS-length checks. Dependency and Unicode tables are pinned by go.mod.
var names = idna.New(idna.MapForLookup(), idna.Transitional(false), idna.StrictDomainName(true), idna.ValidateLabels(true), idna.CheckHyphens(true), idna.CheckJoiners(true), idna.BidiRule(), idna.VerifyDNSLength(true))

func normalizeName(s string) (string, bool) {
	if len(s) == 0 || len(s) > 4096 || strings.HasSuffix(s, ".") {
		return "", false
	}
	n, e := names.ToASCII(s)
	n = strings.ToLower(n)
	if e != nil || len(n) > 253 || !strings.Contains(n, ".") || strings.HasSuffix(n, ".") {
		return "", false
	}
	if _, e := netip.ParseAddr(n); e == nil {
		return "", false
	}
	for _, suffix := range []string{"local", "localhost", "internal", "lan", "home", "home.arpa", "onion", "invalid", "test"} {
		if n == suffix || strings.HasSuffix(n, "."+suffix) {
			return "", false
		}
	}
	return n, true
}

type Codec struct{}

// requirement-ref: REQ-DNS-01-0014
// requirement-ref: REQ-DNS-01-0015
// requirement-ref: REQ-DNS-01-0023
// requirement-ref: REQ-DNS-01-0025
func (Codec) Definition() taskdecode.Definition {
	str := func() *jsondecode.Schema { return &jsondecode.Schema{Kind: jsondecode.String} }
	return taskdecode.Definition{TestID: "DNS-01", TestVersion: "0.1.0", Input: &jsondecode.Schema{Kind: jsondecode.Object, Fields: map[string]jsondecode.Field{
		"hostname": {Required: true, Schema: str()}, "record_type": {Required: true, Schema: str()},
		"resolver": {Schema: &jsondecode.Schema{Kind: jsondecode.Object, Fields: map[string]jsondecode.Field{"ip_address": {Required: true, Schema: str()}, "port": {Schema: &jsondecode.Schema{Kind: jsondecode.Integer, Min: "1", Max: "65535"}}}}},
	}}, Options: &jsondecode.Schema{Kind: jsondecode.Object, Fields: map[string]jsondecode.Field{
		"resolver_mode": {Schema: str()}, "transport": {Schema: str()}, "collect_additional_responses": {Schema: &jsondecode.Schema{Kind: jsondecode.Boolean}}, "additional_response_window_ms": {Schema: &jsondecode.Schema{Kind: jsondecode.Integer, Min: "1", Max: "1000", Limit: true}},
	}}}
}

// requirement-ref: REQ-DNS-01-0016
// requirement-ref: REQ-DNS-01-0017
// requirement-ref: REQ-DNS-01-0019
// requirement-ref: REQ-DNS-01-0021
// requirement-ref: REQ-DNS-01-0026
// requirement-ref: REQ-LIMIT-0014
func (c Codec) Normalize(input, options *jsondecode.Value) (task.Parameters, error) {
	d := c.Definition()
	bad := func() (task.Parameters, error) { return nil, taskdecode.ErrParameters }
	if input == nil || options == nil || jsondecode.Validate(input, d.Input) != nil || jsondecode.Validate(options, d.Options) != nil {
		return bad()
	}
	n, _ := input.Lookup("hostname")
	q, _ := input.Lookup("record_type")
	name, ok := normalizeName(n.Text())
	if !ok || q.Text() != "A" && q.Text() != "AAAA" {
		return bad()
	}
	p := Parameters{name: name, query: q.Text(), mode: "system", transport: "udp"}
	if v, ok := options.Lookup("resolver_mode"); ok {
		p.mode = v.Text()
	}
	if p.mode != "system" && p.mode != "direct" {
		return bad()
	}
	tr, hasTransport := options.Lookup("transport")
	if hasTransport {
		p.transport = tr.Text()
	}
	co, hasCollect := options.Lookup("collect_additional_responses")
	if hasCollect {
		p.collect = co.Bool()
	}
	wi, hasWindow := options.Lookup("additional_response_window_ms")
	resolver, hasResolver := input.Lookup("resolver")
	if p.mode == "system" {
		if hasResolver || hasTransport || p.collect || hasWindow {
			return bad()
		}
		return p, nil
	}
	if !hasResolver || p.transport != "udp" && p.transport != "tcp" || p.collect && p.transport != "udp" || hasWindow != p.collect {
		return bad()
	}
	if hasWindow {
		v, _ := strconv.Atoi(wi.Text())

		p.window = time.Duration(v) * time.Millisecond
	}
	ip, _ := resolver.Lookup("ip_address")
	a, e := netip.ParseAddr(ip.Text())
	if e != nil || a.String() != ip.Text() || a.Is4In6() || !task.PublicAddress(a) {
		return bad()
	}
	p.address = a
	p.port = 53
	if v, ok := resolver.Lookup("port"); ok {
		num, _ := strconv.Atoi(v.Text())
		p.port = uint16(num)
	}
	return p, nil
}

type ProxyRule struct{}

// requirement-ref: REQ-PROXY-0044
func (ProxyRule) Compatible(params task.Parameters) bool {
	p, ok := params.(Parameters)
	return ok && p.mode == "direct" && p.transport == "tcp"
}
