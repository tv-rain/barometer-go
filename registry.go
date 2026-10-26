package barometer

import (
	dnsmeasurement "github.com/tv-rain/barometer-go/internal/measurement/dns01"
	tcpmeasurement "github.com/tv-rain/barometer-go/internal/measurement/tcp01"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/registry"
)

// productionRegistry assembles only delivered measurements, per client
// generation. Presence is not released, platform, or field-validated support.
// requirement-ref: REQ-TCP-01-0023
// requirement-ref: REQ-TCP-01-0031
func productionRegistry() (*registry.Registry, error) {
	return registry.New([]registry.Entry{{Test: task.TestVersion{ID: "DNS-01", Version: "0.1.0"}, Codec: dnsmeasurement.Codec{}, BaselineRisk: task.LowRisk, Targets: dns01Targets{}, Proxy: dnsmeasurement.ProxyRule{}, Runner: &dnsmeasurement.Runner{}, Step: true}, {Test: task.TestVersion{ID: "TCP-01", Version: "0.1.0"}, Codec: tcpmeasurement.Codec{}, BaselineRisk: task.LowRisk, Targets: tcp01Targets{}, Proxy: tcpmeasurement.ProxyRule{}, Runner: &tcpmeasurement.Runner{}, Step: true}})
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
