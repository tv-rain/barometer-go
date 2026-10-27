// Package dns04 implements the Draft 0.1.0 DNS-over-TLS resolution
// measurement. It reuses the shared TCP, TLS and DNS protocol machinery, not
// the standalone TLS-01 input or operation schema.
package dns04

import (
	"net/netip"
	"strconv"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/taskdecode"
)

// Parameters owns a normalized resolver name and port, an optional canonical
// public bootstrap address, a normalized query name and record type and the
// connection family; the zero value is invalid. It grants no authority.
type Parameters struct {
	resolver, name, query, family string
	port                          uint16
	bootstrap                     netip.Addr
}

func (Parameters) TaskParameters()  {}
func (Parameters) String() string   { return "dns04.Parameters" }
func (Parameters) GoString() string { return "dns04.Parameters" }

// Resolver reports the resolver name, used for bootstrap and TLS
// authentication, its port and the optional bootstrap address.
func (p Parameters) Resolver() (string, uint16, netip.Addr) { return p.resolver, p.port, p.bootstrap }
func (p Parameters) Name() string                           { return p.name }
func (p Parameters) QueryType() string                      { return p.query }
func (p Parameters) AddressFamily() string                  { return p.family }

// Bootstrap reports whether the resolver name is resolved through the system
// resolver, that is, whether no bootstrap address was supplied.
func (p Parameters) Bootstrap() bool { return !p.bootstrap.IsValid() }

func (p Parameters) valid() bool {
	return p.resolver != "" && p.name != "" && p.port != 0 && (p.query == "A" || p.query == "AAAA")
}

// Codec supplies closed descriptors and pure normalization; authority stays
// with admission. It uses the common strict JSON stages, never a second parser.
type Codec struct{}

// requirement-ref: REQ-DNS-04-0009
// requirement-ref: REQ-DNS-04-0014
// requirement-ref: REQ-DNS-04-0016
func (Codec) Definition() taskdecode.Definition {
	str := func() *jsondecode.Schema { return &jsondecode.Schema{Kind: jsondecode.String} }
	return taskdecode.Definition{TestID: "DNS-04", TestVersion: "0.1.0",
		Input: &jsondecode.Schema{Kind: jsondecode.Object, Fields: map[string]jsondecode.Field{
			"resolver_name": {Required: true, Schema: str()},
			"port":          {Required: true, Schema: &jsondecode.Schema{Kind: jsondecode.Integer, Min: "1", Max: "65535"}},
			"hostname":      {Required: true, Schema: str()},
			"record_type":   {Required: true, Schema: str()},
			"bootstrap_ip":  {Schema: str()},
		}}, Options: &jsondecode.Schema{Kind: jsondecode.Object, Fields: map[string]jsondecode.Field{
			"address_family": {Schema: str()},
		}}}
}

// Names are normalized as DNS-01 normalizes its query name and never derived
// from an address. A non-standard port is not rejected here: admission
// accepts it only through an exactly matching reviewed resolver profile.
// requirement-ref: REQ-DNS-04-0004
// requirement-ref: REQ-DNS-04-0010
// requirement-ref: REQ-DNS-04-0013
// requirement-ref: REQ-DNS-04-0015
// requirement-ref: REQ-DNS-04-0017
func (c Codec) Normalize(input, options *jsondecode.Value) (task.Parameters, error) {
	d := c.Definition()
	bad := func() (task.Parameters, error) { return nil, taskdecode.ErrParameters }
	if input == nil || options == nil || jsondecode.Validate(input, d.Input) != nil || jsondecode.Validate(options, d.Options) != nil {
		return bad()
	}
	rn, _ := input.Lookup("resolver_name")
	pv, _ := input.Lookup("port")
	hn, _ := input.Lookup("hostname")
	rt, _ := input.Lookup("record_type")
	resolver, rok := taskdecode.DNSName(rn.Text())
	name, nok := taskdecode.DNSName(hn.Text())
	port, err := strconv.Atoi(pv.Text())
	if !rok || !nok || err != nil || port < 1 || port > 65535 || rt.Text() != "A" && rt.Text() != "AAAA" {
		return bad()
	}
	p := Parameters{resolver: resolver, name: name, query: rt.Text(), port: uint16(port), family: "auto"}
	if v, ok := options.Lookup("address_family"); ok {
		p.family = v.Text()
	}
	if p.family != "auto" && p.family != "ipv4" && p.family != "ipv6" {
		return bad()
	}
	if v, ok := input.Lookup("bootstrap_ip"); ok {
		a, e := netip.ParseAddr(v.Text())
		if e != nil || a.String() != v.Text() || a.Is4In6() || !task.PublicAddress(a) || p.family == "ipv4" && !a.Is4() || p.family == "ipv6" && !a.Is6() {
			return bad()
		}
		p.bootstrap = a
	}
	return p, nil
}

// ProxyRule permits DNS-04 only with an approved bootstrap address: CONNECT
// uses that address while TLS authenticates the resolver name. A resolver
// name is never a CONNECT authority. It grants no authority.
type ProxyRule struct{}

// requirement-ref: REQ-PROXY-0044
func (ProxyRule) Compatible(params task.Parameters) bool {
	p, ok := params.(Parameters)
	return ok && p.valid() && !p.Bootstrap()
}
