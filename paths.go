package barometer

import (
	"context"
	"errors"

	"github.com/tv-rain/barometer-go/internal/config"
	"github.com/tv-rain/barometer-go/internal/execution"
	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/measurement/dns01"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/networkcontext"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/registry"
	"github.com/tv-rain/barometer-go/internal/result"
)

var (
	errPathUnavailable = errors.New("barometer: network path unavailable")
	errNoProxy         = errors.New("barometer: no proxy configured")
)

// networkPaths is the production path source (DEV-644, DEV-645). At
// reservation it reads the access context once, through the platform network
// capability, and binds the run's result decoder to that snapshot's path:
// direct, or the generation's proxy snapshot. Each attempt then captures its
// own snapshot, evaluates the task's network conditions against it, acquires
// its egress context through a dialer of the same path when the host
// configured egress classification (DEV-647), attaches it to the result, and
// dials through a fresh dialer of the selected path inside the metered
// attempt scope. A proxy task never receives a direct dialer, for its lookup
// or its measurement. Implementation is the trusted producer identity
// of the integration, never host or task input.
type networkPaths struct {
	network        platform.Network
	direct         *networkcontext.Source
	directDialers  execution.DialerSource
	implementation model.Implementation
}

// newNetworkPaths fails without a network capability, so a client that cannot
// read access context has no path and starts no run.
func newNetworkPaths(r runAdapters) (*networkPaths, error) {
	if absent(r.network) {
		return nil, errPathUnavailable
	}
	d, err := networkcontext.NewDirect(r.network)
	if err != nil {
		return nil, errPathUnavailable
	}
	var dialers execution.DialerSource = execution.DirectDialers{}
	if !absent(r.directDialers) {
		dialers = r.directDialers
	}
	return &networkPaths{network: r.network, direct: d, directDialers: dialers, implementation: r.implementation}, nil
}

func (p *networkPaths) plan(ctx context.Context, r pathRequest) (pathPlan, error) {
	if r.NetworkPath != "proxy" {
		return p.bind(ctx, r, p.direct, p.directDialers)
	}
	if !r.Proxy.configured() {
		return p.unconfigured(ctx, r)
	}
	src, err := networkcontext.NewProxy(p.network, r.Proxy.generation, r.Proxy.endpoint.ConnectionCode)
	if err != nil {
		return pathPlan{}, errPathUnavailable
	}
	dialers, err := execution.NewProxyDialers(r.Proxy.endpoint.Host, r.Proxy.endpoint.Port)
	if err != nil {
		return pathPlan{}, errPathUnavailable
	}
	return p.bind(ctx, r, src, dialers)
}

func (p *networkPaths) bind(ctx context.Context, r pathRequest, src *networkcontext.Source, dialers execution.DialerSource) (pathPlan, error) {
	src, err := src.WithLookup(r.Egress)
	if err != nil {
		return pathPlan{}, errPathUnavailable
	}
	snap, err := src.Capture(ctx)
	if err != nil {
		return pathPlan{}, errPathUnavailable
	}
	d, err := p.decoder(r, src.Decoder(), snap.Path(), snap.Generation())
	if err != nil {
		return pathPlan{}, errPathUnavailable
	}
	return pathPlan{Implementation: p.implementation, Access: snap.Access(), Result: d, Context: src, Dialers: dialers}, nil
}

// unconfigured plans a proxy run without a configured proxy. Its attempts
// record proxy_not_configured through the path check before any capture or
// dial, so its context source and dialers only fail. The access reading
// serves the start check alone; the result decoder is bound to no path, so it
// would reject any Network Context.
func (p *networkPaths) unconfigured(ctx context.Context, r pathRequest) (pathPlan, error) {
	snap, err := p.direct.Capture(ctx)
	if err != nil {
		return pathPlan{}, errPathUnavailable
	}
	d, err := p.decoder(r, p.direct.Decoder(), observation.Path{}, snap.Generation())
	if err != nil {
		return pathPlan{}, errPathUnavailable
	}
	return pathPlan{Implementation: p.implementation, Access: snap.Access(), Result: d, Context: noProxy{}, Dialers: noProxy{}}, nil
}

func (p *networkPaths) decoder(r pathRequest, cd *observation.ContextDecoder, path observation.Path, generation uint64) (*result.Decoder, error) {
	output := result.Output{Test: model.Test{ID: r.Test.ID, Version: r.Test.Version, ResultVersion: result.Contract}}
	if params, ok := r.Parameters.(dns01.Parameters); ok {
		output.DNSMode = params.Engine()
		output.QueryType = params.QueryType()
		output.Limits.RetainDNSAnswers = true
	}
	return result.New(result.Binding{TaskID: r.TaskID, Output: output,
		MaxAttempts: r.MaxAttempts, MaxDurationMS: r.MaxDurationMS, Implementation: p.implementation,
		ContextDecoder: cd, Path: path, AccessGeneration: generation})
}

// noProxy is the context source and dialer source of a proxy run without a
// configured proxy. It never captures context and never dials.
type noProxy struct{}

func (noProxy) Capture(context.Context) (networkcontext.Snapshot, error) {
	return networkcontext.Snapshot{}, errNoProxy
}
func (noProxy) Dialer(int) (measurement.Dialer, error) { return nil, errNoProxy }

// proxySnapshot is the immutable path of a leased proxy generation. The
// endpoint and generation are internal and never reach a result.
type proxySnapshot struct {
	endpoint   config.Proxy
	generation uint64 // zero when no proxy is configured
}

func (p proxySnapshot) configured() bool { return p.generation != 0 }
func (p proxySnapshot) String() string   { return "barometer.proxySnapshot" }
func (p proxySnapshot) GoString() string { return "barometer.proxySnapshot" }

// proxy returns the generation's proxy snapshot.
func (rt *runtime) proxy() proxySnapshot {
	if rt.proxies == nil {
		return proxySnapshot{}
	}
	return rt.proxies.snapshot()
}

// proxyCheck is the pre-attempt check of a proxy run (REQ-EXECUTION-0008),
// evaluated by the engine before each attempt. The proxy must be configured,
// the task's admitted path authority, which local policy bounds, must permit
// the proxy, and the registered rule must accept the test's concrete
// parameters. Admission already rejects a task that fails either of the last
// two (REQ-PROXY-0042), and a policy replacement admits every task again, so
// in this build they hold for every admitted task; the check still re-applies
// them rather than assume it.
// requirement-ref: REQ-EXECUTION-0008
// requirement-ref: REQ-PROXY-0022
type proxyCheck struct {
	configured, permitted bool
	rule                  registry.ProxyRule
	targets               registry.TargetRule
	parameters            task.Parameters
}

func (c proxyCheck) Check() task.ExecutionReason {
	switch {
	case !c.configured:
		return task.ProxyNotConfigured
	case !c.permitted:
		return task.ProxyNotPermitted
	case !registry.ProxyCompatible(c.rule, c.targets, c.parameters):
		return task.ProxyPathUnsupported
	}
	return ""
}
