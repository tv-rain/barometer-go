package barometer

import (
	"crypto/x509"
	"net/netip"

	dnsmeasurement "github.com/tv-rain/barometer-go/internal/measurement/dns01"
	"github.com/tv-rain/barometer-go/internal/measurement/dns04"
	httpmeasurement "github.com/tv-rain/barometer-go/internal/measurement/http01"
	"github.com/tv-rain/barometer-go/internal/measurement/net01"
	"github.com/tv-rain/barometer-go/internal/measurement/net03"
	"github.com/tv-rain/barometer-go/internal/measurement/net05"
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
// requirement-ref: REQ-NET-01-0022
// requirement-ref: REQ-NET-01-0032
// requirement-ref: REQ-NET-03-0004
// requirement-ref: REQ-NET-03-0022
// requirement-ref: REQ-NET-03-0034
// requirement-ref: REQ-DNS-04-0033
// requirement-ref: REQ-DNS-04-0039
func productionRegistry(roots *x509.CertPool) (*registry.Registry, error) {
	entries := []registry.Entry{{Test: task.TestVersion{ID: "DNS-01", Version: "0.1.0"}, Codec: dnsmeasurement.Codec{}, BaselineRisk: task.LowRisk, Targets: dns01Targets{}, Proxy: dnsmeasurement.ProxyRule{}, Runner: &dnsmeasurement.Runner{}, Step: true}, {Test: task.TestVersion{ID: "TCP-01", Version: "0.1.0"}, Codec: tcpmeasurement.Codec{}, BaselineRisk: task.LowRisk, Targets: tcp01Targets{}, Proxy: tcpmeasurement.ProxyRule{}, Runner: &tcpmeasurement.Runner{}, Step: true},
		{Test: task.TestVersion{ID: "TLS-01", Version: "0.1.0"}, Codec: tlsmeasurement.Codec{}, BaselineRisk: task.LowRisk, Targets: tls01Targets{}, Proxy: tlsmeasurement.ProxyRule{}, Runner: &tlsmeasurement.Runner{Roots: roots}, Step: true}, {Test: task.TestVersion{ID: "HTTP-01", Version: "0.1.0"}, Codec: httpmeasurement.Codec{}, BaselineRisk: task.LowRisk, Targets: http01Targets{}, Proxy: httpmeasurement.ProxyRule{}, Runner: &httpmeasurement.Runner{Roots: roots}, Step: true}}
	primitives, err := registry.New(entries)
	if err != nil {
		return nil, err
	}
	codec, err := net05.New(primitives.Codecs())
	if err != nil {
		return nil, err
	}
	rules := net05Rules{primitives: primitives}
	// NET-01 has no remote target and is rejected on the proxy path
	// (REQ-PROXY-0026); it is not a NET-05 step.
	entries = append(entries, registry.Entry{Test: task.TestVersion{ID: "NET-01", Version: "0.1.0"}, Codec: net01.Codec{}, BaselineRisk: task.LowRisk, Targets: registry.NoTargets{}, Proxy: registry.NoProxy{}, Runner: net01.Runner{}})
	// NET-03 combines platform state with one endpoint request and has no
	// proxy path semantics (REQ-PROXY-0044); it is not a NET-05 step.
	entries = append(entries, registry.Entry{Test: task.TestVersion{ID: "NET-03", Version: "0.1.0"}, Codec: net03.Codec{}, BaselineRisk: task.LowRisk, Targets: net03Targets{}, Proxy: registry.NoProxy{}, Runner: &net03.Runner{Roots: roots}})
	// DNS-04 is proxy-compatible only with an approved bootstrap address
	// (REQ-PROXY-0044); it is not a NET-05 step.
	entries = append(entries, registry.Entry{Test: task.TestVersion{ID: "DNS-04", Version: "0.1.0"}, Codec: dns04.Codec{}, BaselineRisk: task.MediumRisk, Targets: dns04Targets{}, Proxy: dns04.ProxyRule{}, Runner: &dns04.Runner{Roots: roots}})
	entries = append(entries, registry.Entry{Test: task.TestVersion{ID: "NET-05", Version: "0.1.0"}, Codec: codec, BaselineRisk: task.MediumRisk, Targets: rules, Proxy: rules, Runner: &net05.Runner{Roots: roots}, Composite: true})
	return registry.New(entries)
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

// DNS-04 matches the query name and the complete resolver profile: name,
// port and the optional bootstrap address. The resolver name is the connection
// destination only without a bootstrap address, and always the TLS name.
// requirement-ref: REQ-DNS-04-0004
// requirement-ref: REQ-DNS-04-0010
// requirement-ref: REQ-DNS-04-0011
// requirement-ref: REQ-DNS-04-0013
// requirement-ref: REQ-DNS-04-0038
type dns04Targets struct{}

func (dns04Targets) Targets(params task.Parameters) ([]registry.Target, bool) {
	p, ok := params.(dns04.Parameters)
	resolver, port, bootstrap := p.Resolver()
	if !ok || p.Name() == "" || resolver == "" || port == 0 {
		return nil, false
	}
	profile := task.TargetTuple{Hostname: resolver, ResolverName: resolver, TLSServerName: resolver, Port: port, Scheme: "tls"}
	if bootstrap.IsValid() {
		profile.BootstrapAddresses = []netip.Addr{bootstrap}
	}
	return []registry.Target{{Role: task.MeasurementTarget, Tuple: task.TargetTuple{Hostname: p.Name()}}, {Role: task.ResolverTarget, Tuple: profile}}, true
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

// NET-03 matches the complete normalized validation URL, expected status and
// optional body digest, so only a locally registered profile runs
// (REQ-NET-03-0010, REQ-NET-03-0018). Redirect targets are never registered.
type net03Targets struct{}

// requirement-ref: REQ-NET-03-0010
func (net03Targets) Targets(p task.Parameters) ([]registry.Target, bool) {
	v, ok := p.(net03.Parameters)
	if !ok || v.Target().URL == "" {
		return nil, false
	}
	return []registry.Target{{Role: task.MeasurementTarget, Tuple: v.Target()}}, true
}

// net05Rules composes the same exact primitive entries used by the codec and
// admission. It never grants a primitive's targets NET-05 registration scope.
type net05Rules struct{ primitives *registry.Registry }

// requirement-ref: REQ-NET-05-0019
func (r net05Rules) Targets(params task.Parameters) ([]registry.Target, bool) {
	p, ok := params.(net05.Parameters)
	if !ok {
		return nil, false
	}
	steps := p.CompositeSteps()
	if len(steps) < 1 || len(steps) > 8 {
		return nil, false
	}
	var out []registry.Target
	for _, step := range steps {
		entry, ok := r.primitives.Lookup(step.Test)
		if !ok || !entry.Step {
			return nil, false
		}
		targets, ok := entry.Targets.Targets(step.Parameters)
		if !ok || len(targets) == 0 || len(out)+len(targets) > 64 {
			return nil, false
		}
		out = append(out, targets...)
	}
	return out, true
}

// requirement-ref: REQ-PROXY-0044
// requirement-ref: REQ-TRUST-0039
func (r net05Rules) Compatible(params task.Parameters) bool {
	p, ok := params.(net05.Parameters)
	if !ok {
		return false
	}
	steps := p.CompositeSteps()
	if len(steps) < 1 || len(steps) > 8 {
		return false
	}
	for _, step := range steps {
		entry, ok := r.primitives.Lookup(step.Test)
		if !ok || !entry.Step || !registry.ProxyCompatible(entry.Proxy, entry.Targets, step.Parameters) {
			return false
		}
	}
	return true
}
