// Package dns03 implements the Draft 0.1.0 DNS-over-HTTPS resolution
// measurement. It reuses the shared TCP, TLS, HTTP and DNS protocol
// machinery, not the standalone TLS-01 or HTTP-01 input or operation schema.
package dns03

import (
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/taskdecode"
)

// Parameters owns a normalized resolver URL with its hostname and port, an
// optional canonical public bootstrap address, a normalized query name and
// record type and the connection family; the zero value is invalid. It
// grants no authority.
type Parameters struct {
	url, host, name, query, family string
	port                           uint16
	bootstrap                      netip.Addr
}

func (Parameters) TaskParameters()  {}
func (Parameters) String() string   { return "dns03.Parameters" }
func (Parameters) GoString() string { return "dns03.Parameters" }

// Resolver reports the normalized resolver URL, its hostname, used for
// bootstrap and TLS authentication, its port and the optional bootstrap
// address.
func (p Parameters) Resolver() (string, string, uint16, netip.Addr) {
	return p.url, p.host, p.port, p.bootstrap
}
func (p Parameters) Name() string          { return p.name }
func (p Parameters) QueryType() string     { return p.query }
func (p Parameters) AddressFamily() string { return p.family }

// Bootstrap reports whether the resolver hostname is resolved through the
// system resolver, that is, whether no bootstrap address was supplied.
func (p Parameters) Bootstrap() bool { return !p.bootstrap.IsValid() }

func (p Parameters) valid() bool {
	return p.url != "" && p.host != "" && p.name != "" && p.port != 0 && (p.query == "A" || p.query == "AAAA")
}

// Codec supplies closed descriptors and pure normalization; authority stays
// with admission. It uses the common strict JSON stages, never a second parser.
type Codec struct{}

// requirement-ref: REQ-DNS-03-0009
// requirement-ref: REQ-DNS-03-0014
// requirement-ref: REQ-DNS-03-0016
func (Codec) Definition() taskdecode.Definition {
	str := func() *jsondecode.Schema { return &jsondecode.Schema{Kind: jsondecode.String} }
	return taskdecode.Definition{TestID: "DNS-03", TestVersion: "0.1.0",
		Input: &jsondecode.Schema{Kind: jsondecode.Object, Fields: map[string]jsondecode.Field{
			"resolver_url": {Required: true, Schema: str()},
			"hostname":     {Required: true, Schema: str()},
			"record_type":  {Required: true, Schema: str()},
			"bootstrap_ip": {Schema: str()},
		}}, Options: &jsondecode.Schema{Kind: jsondecode.Object, Fields: map[string]jsondecode.Field{
			"address_family": {Schema: str()},
		}}}
}

// The query name is normalized as DNS-01 normalizes it. The closed option
// shape admits no GET, redirect, alternate endpoint, header, credential,
// cookie or proxy control.
// requirement-ref: REQ-DNS-03-0004
// requirement-ref: REQ-DNS-03-0013
// requirement-ref: REQ-DNS-03-0015
// requirement-ref: REQ-DNS-03-0017
func (c Codec) Normalize(input, options *jsondecode.Value) (task.Parameters, error) {
	d := c.Definition()
	bad := func() (task.Parameters, error) { return nil, taskdecode.ErrParameters }
	if input == nil || options == nil || jsondecode.Validate(input, d.Input) != nil || jsondecode.Validate(options, d.Options) != nil {
		return bad()
	}
	ru, _ := input.Lookup("resolver_url")
	hn, _ := input.Lookup("hostname")
	rt, _ := input.Lookup("record_type")
	if len(ru.Text()) > 2048 {
		return nil, jsondecode.ErrLimit
	}
	raw, host, port, uok := NormalizeURL(ru.Text())
	name, nok := taskdecode.DNSName(hn.Text())
	if !uok || !nok || rt.Text() != "A" && rt.Text() != "AAAA" {
		return bad()
	}
	p := Parameters{url: raw, host: host, port: port, name: name, query: rt.Text(), family: "auto"}
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

// NormalizeURL accepts an absolute https resolver URL whose host is a DNS
// name, never an address: the resolver is resolved or bootstrapped and
// authenticated by that name. It rejects credentials, fragments and
// interpolation, folds the host case, omits default port 443 and supplies an
// empty path as /, preserving the path and query bytes otherwise. Whether a
// query is user-specific is decided by exact registration, not here.
// requirement-ref: REQ-DNS-03-0009
// requirement-ref: REQ-DNS-03-0013
func NormalizeURL(raw string) (string, string, uint16, bool) {
	if len(raw) > 2048 || !utf8.ValidString(raw) || strings.ContainsAny(raw, "{}\r\n\t ") {
		return "", "", 0, false
	}
	u, e := url.Parse(raw)
	if e != nil || u.Scheme != "https" || u.User != nil || u.Fragment != "" || strings.Contains(raw, "#") || u.Opaque != "" || u.Host == "" || strings.HasSuffix(u.Host, ":") {
		return "", "", 0, false
	}
	host, ok := taskdecode.DNSName(u.Hostname())
	if !ok {
		return "", "", 0, false
	}
	port := 443
	if u.Port() != "" {
		n, e := strconv.Atoi(u.Port())
		if e != nil || n < 1 || n > 65535 || strconv.Itoa(n) != u.Port() {
			return "", "", 0, false
		}
		port = n
	}
	u.Host = host
	if port != 443 {
		u.Host += ":" + strconv.Itoa(port)
	}
	if u.Path == "" {
		u.Path = "/"
	}
	if len(u.String()) > 2048 {
		return "", "", 0, false
	}
	return u.String(), host, uint16(port), true
}

// ProxyRule permits DNS-03 only with an approved bootstrap address: CONNECT
// uses that address while TLS authenticates the resolver hostname. A
// resolver hostname is never a CONNECT authority. It grants no authority.
type ProxyRule struct{}

// requirement-ref: REQ-PROXY-0044
func (ProxyRule) Compatible(params task.Parameters) bool {
	p, ok := params.(Parameters)
	return ok && p.valid() && !p.Bootstrap()
}
