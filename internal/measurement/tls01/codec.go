// Package tls01 implements the Draft 0.1.0 TLS handshake measurement.
package tls01

import (
	"net/netip"
	"slices"
	"strconv"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/taskdecode"
)

// Parameters owns an exact canonical public address, port, normalized server
// name and registered ALPN offer; the zero value is invalid.
type Parameters struct {
	address netip.Addr
	port    uint16
	name    string
	alpn    []string
}

func (Parameters) TaskParameters()  {}
func (Parameters) String() string   { return "tls01.Parameters" }
func (Parameters) GoString() string { return "tls01.Parameters" }

// Target reports the explicit destination and the configured server name.
// The name is SNI and certificate hostname only, never connection authority.
func (p Parameters) Target() (netip.Addr, uint16, string, bool) {
	return p.address, p.port, p.name, p.port != 0 && p.name != "" && task.PublicAddress(p.address) && !p.address.Is4In6()
}

// ALPN returns a copy of the offer, in order.
func (p Parameters) ALPN() []string { return slices.Clone(p.alpn) }

// DeclaredMaxima reports the NET-05 step declaration: one connection and no
// application bytes, because handshake bytes are not counted.
func (Parameters) DeclaredMaxima() task.AttemptLimits { return task.AttemptLimits{Connections: 1} }

// Codec supplies closed descriptors and pure normalization; authority stays
// with admission. It uses the common strict JSON stages, never a second parser.
type Codec struct{}

// requirement-ref: REQ-TLS-01-0009
// requirement-ref: REQ-TLS-01-0014
// requirement-ref: REQ-TLS-01-0016
// requirement-ref: REQ-TLS-01-0017
// requirement-ref: REQ-JSON-0008
func (Codec) Definition() taskdecode.Definition {
	str := func() *jsondecode.Schema { return &jsondecode.Schema{Kind: jsondecode.String} }
	return taskdecode.Definition{TestID: "TLS-01", TestVersion: "0.1.0",
		Input: &jsondecode.Schema{Kind: jsondecode.Object, Fields: map[string]jsondecode.Field{
			"ip_address":  {Required: true, Schema: str()},
			"port":        {Required: true, Schema: &jsondecode.Schema{Kind: jsondecode.Integer, Min: "1", Max: "65535"}},
			"server_name": {Required: true, Schema: str()},
		}}, Options: &jsondecode.Schema{Kind: jsondecode.Object, Fields: map[string]jsondecode.Field{
			"alpn_protocols": {Schema: &jsondecode.Schema{Kind: jsondecode.Array, MaxCount: 8, Element: str()}},
		}}}
}

// The server name is normalized deterministically, as DNS-01 normalizes its
// query name, and never derived from the address; admission matches the
// normalized name against the registered tuple.
// requirement-ref: REQ-TLS-01-0004
// requirement-ref: REQ-TLS-01-0010
// requirement-ref: REQ-TLS-01-0011
// requirement-ref: REQ-TLS-01-0012
// requirement-ref: REQ-TLS-01-0015
// requirement-ref: REQ-OBS-0049
func (c Codec) Normalize(input, options *jsondecode.Value) (task.Parameters, error) {
	d := c.Definition()
	if input == nil || options == nil || jsondecode.Validate(input, d.Input) != nil || jsondecode.Validate(options, d.Options) != nil {
		return nil, taskdecode.ErrParameters
	}
	ip, _ := input.Lookup("ip_address")
	port, _ := input.Lookup("port")
	sni, _ := input.Lookup("server_name")
	a, e := netip.ParseAddr(ip.Text())
	n, ne := strconv.Atoi(port.Text())
	name, ok := taskdecode.DNSName(sni.Text())
	if e != nil || ne != nil || !ok || a.String() != ip.Text() || a.Is4In6() || !task.PublicAddress(a) || n < 1 || n > 65535 {
		return nil, taskdecode.ErrParameters
	}
	var alpn []string
	if list, ok := options.Lookup("alpn_protocols"); ok {
		for _, v := range list.Elements() {
			alpn = append(alpn, v.Text())
		}
	}
	if observation.ValidateALPNOffer(observation.Contract, alpn) != nil {
		return nil, taskdecode.ErrParameters
	}
	return Parameters{address: a, port: uint16(n), name: name, alpn: alpn}, nil
}

// ProxyRule permits this exact explicit-IP handshake shape; it grants no
// authority. The CONNECT authority is the address and port, never the name.
type ProxyRule struct{}

func (ProxyRule) Compatible(p task.Parameters) bool {
	v, ok := p.(Parameters)
	if !ok {
		return false
	}
	_, _, _, ok = v.Target()
	return ok
}
