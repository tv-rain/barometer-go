package barometer

import (
	"crypto/x509"

	dnsmeasurement "github.com/tv-rain/barometer-go/internal/measurement/dns01"
	httpmeasurement "github.com/tv-rain/barometer-go/internal/measurement/http01"
	tcpmeasurement "github.com/tv-rain/barometer-go/internal/measurement/tcp01"
	tlsmeasurement "github.com/tv-rain/barometer-go/internal/measurement/tls01"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/registry"
)

// productionRegistry assembles only delivered measurements, per client
// generation. Presence is not released, platform, or field-validated support.
// roots are TLS-01's trust anchors; nil selects the system roots.
// requirement-ref: REQ-TCP-01-0023
// requirement-ref: REQ-TCP-01-0031
// requirement-ref: REQ-TLS-01-0024
// requirement-ref: REQ-TLS-01-0032
func productionRegistry(roots *x509.CertPool) (*registry.Registry, error) {
	return registry.New([]registry.Entry{{Test: task.TestVersion{ID: "DNS-01", Version: "0.1.0"}, Codec: dnsmeasurement.Codec{}, BaselineRisk: task.LowRisk, Targets: dns01Targets{}, Proxy: dnsmeasurement.ProxyRule{}, Runner: &dnsmeasurement.Runner{}, Step: true}, {Test: task.TestVersion{ID: "TCP-01", Version: "0.1.0"}, Codec: tcpmeasurement.Codec{}, BaselineRisk: task.LowRisk, Targets: tcp01Targets{}, Proxy: tcpmeasurement.ProxyRule{}, Runner: &tcpmeasurement.Runner{}, Step: true},
		{Test: task.TestVersion{ID: "TLS-01", Version: "0.1.0"}, Codec: tlsmeasurement.Codec{}, BaselineRisk: task.LowRisk, Targets: tls01Targets{}, Proxy: tlsmeasurement.ProxyRule{}, Runner: &tlsmeasurement.Runner{Roots: roots}, Step: true}, {Test: task.TestVersion{ID: "HTTP-01", Version: "0.1.0"}, Codec: httpmeasurement.Codec{}, BaselineRisk: task.LowRisk, Targets: http01Targets{}, Proxy: httpmeasurement.ProxyRule{}, Runner: &httpmeasurement.Runner{Roots: roots}, Step: true}})
}

type tcp01Targets struct{}

// requirement-ref: REQ-TCP-01-0018
func (tcp01Targets) Targets(p task.Parameters) ([]registry.Target, bool) {
	v, ok := p.(tcpmeasurement.Parameters)
	if !ok {
		return nil, false
	}
	a, port, ok := v.Target()
	if !ok {
		return nil, false
	}
	return []registry.Target{{Role: task.MeasurementTarget, Tuple: task.TargetTuple{Address: a, Port: port, Scheme: "tcp"}}}, true
}

// The complete reviewed tuple is the address, port, server name and ALPN
// offer. The server name is SNI, not a connection authority.
// requirement-ref: REQ-TLS-01-0010
// requirement-ref: REQ-TLS-01-0037
type tls01Targets struct{}

func (tls01Targets) Targets(p task.Parameters) ([]registry.Target, bool) {
	v, ok := p.(tlsmeasurement.Parameters)
	if !ok {
		return nil, false
	}
	a, port, name, ok := v.Target()
	if !ok {
		return nil, false
	}
	return []registry.Target{{Role: task.MeasurementTarget, Tuple: task.TargetTuple{Address: a, Port: port, Scheme: "tls", TLSServerName: name, ALPN: v.ALPN()}}}, true
}

// DNS query names are registrations, never CONNECT authorities. Direct mode
// independently matches the explicit resolver's transport/address/port tuple.
// requirement-ref: REQ-DNS-01-0020
// requirement-ref: REQ-DNS-01-0028
// requirement-ref: REQ-DNS-01-0029
type dns01Targets struct{}

func (dns01Targets) Targets(params task.Parameters) ([]registry.Target, bool) {
	p, ok := params.(dnsmeasurement.Parameters)
	if !ok || p.Name() == "" {
		return nil, false
	}
	out := []registry.Target{{Role: task.MeasurementTarget, Tuple: task.TargetTuple{Hostname: p.Name()}}}
	if p.Engine() != "system" {
		a, port := p.Resolver()
		out = append(out, registry.Target{Role: task.ResolverTarget, Tuple: task.TargetTuple{Address: a, Port: port, Scheme: p.Engine()}})
	}
	return out, true
}

// HTTP-01 matches the complete normalized URL and expected checks.
type http01Targets struct{}

func (http01Targets) Targets(p task.Parameters) ([]registry.Target, bool) {
	v, ok := p.(httpmeasurement.Parameters)
	if !ok || v.Target().URL == "" {
		return nil, false
	}
	return []registry.Target{{Role: task.MeasurementTarget, Tuple: v.Target()}}, true
}
