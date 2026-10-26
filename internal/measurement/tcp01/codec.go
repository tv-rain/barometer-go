// Package tcp01 implements the Draft 0.1.0 TCP connect measurement.
package tcp01

import (
	"net/netip"
	"strconv"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/taskdecode"
)

// Parameters owns an exact canonical public address/port; zero is invalid.
type Parameters struct {
	address netip.Addr
	port    uint16
}

func (Parameters) TaskParameters()  {}
func (Parameters) String() string   { return "tcp01.Parameters" }
func (Parameters) GoString() string { return "tcp01.Parameters" }
func (p Parameters) Target() (netip.Addr, uint16, bool) {
	return p.address, p.port, p.port != 0 && task.PublicAddress(p.address) && !p.address.Is4In6()
}

// Codec supplies closed descriptors and pure normalization; authority stays
// with admission. It uses the common strict JSON stages, never a second parser.
type Codec struct{}

// requirement-ref: REQ-TCP-01-0010
// requirement-ref: REQ-TCP-01-0014
// requirement-ref: REQ-TCP-01-0015
// requirement-ref: REQ-TCP-01-0016
// requirement-ref: REQ-JSON-0008
func (Codec) Definition() taskdecode.Definition {
	return taskdecode.Definition{TestID: "TCP-01", TestVersion: "0.1.0",
		Input: &jsondecode.Schema{Kind: jsondecode.Object, Fields: map[string]jsondecode.Field{
			"ip_address": {Required: true, Schema: &jsondecode.Schema{Kind: jsondecode.String}},
			"port":       {Required: true, Schema: &jsondecode.Schema{Kind: jsondecode.Integer, Min: "1", Max: "65535"}},
		}}, Options: &jsondecode.Schema{Kind: jsondecode.Object}}
}

// requirement-ref: REQ-TCP-01-0001
// requirement-ref: REQ-TCP-01-0005
// requirement-ref: REQ-TCP-01-0011
// requirement-ref: REQ-TCP-01-0012
// requirement-ref: REQ-TCP-01-0013
func (c Codec) Normalize(input, options *jsondecode.Value) (task.Parameters, error) {
	d := c.Definition()
	if input == nil || options == nil || jsondecode.Validate(input, d.Input) != nil || jsondecode.Validate(options, d.Options) != nil {
		return nil, taskdecode.ErrParameters
	}
	ip, _ := input.Lookup("ip_address")
	port, _ := input.Lookup("port")
	a, e := netip.ParseAddr(ip.Text())
	n, ne := strconv.Atoi(port.Text())
	if e != nil || ne != nil || a.String() != ip.Text() || a.Is4In6() || !task.PublicAddress(a) || n < 1 || n > 65535 {
		return nil, taskdecode.ErrParameters
	}
	return Parameters{address: a, port: uint16(n)}, nil
}

// ProxyRule permits this exact explicit-IP connect shape; it grants no authority.
type ProxyRule struct{}

func (ProxyRule) Compatible(p task.Parameters) bool {
	v, ok := p.(Parameters)
	if !ok {
		return false
	}
	_, _, ok = v.Target()
	return ok
}
