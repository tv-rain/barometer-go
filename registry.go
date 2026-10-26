package barometer

import (
	tcpmeasurement "github.com/tv-rain/barometer-go/internal/measurement/tcp01"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/registry"
)

// productionRegistry assembles only delivered measurements, per client
// generation. Presence is not released, platform, or field-validated support.
// requirement-ref: REQ-TCP-01-0023
// requirement-ref: REQ-TCP-01-0031
func productionRegistry() (*registry.Registry, error) {
	return registry.New([]registry.Entry{{Test: task.TestVersion{ID: "TCP-01", Version: "0.1.0"}, Codec: tcpmeasurement.Codec{}, BaselineRisk: task.LowRisk, Targets: tcp01Targets{}, Proxy: tcpmeasurement.ProxyRule{}, Runner: &tcpmeasurement.Runner{}, Step: true}})
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
