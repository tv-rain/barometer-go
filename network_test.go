package barometer

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/netip"
	"strconv"
	"sync"
	"testing"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/registry"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

// echoRunner dials a loopback echo server through the attempt's dialer,
// exchanges loopback.Payload and reports a tcp_connect observation in the
// TCP-01 shape. It is synthetic and not TCP-01 evidence.
type echoRunner struct {
	address string
	mu      sync.Mutex
	calls   int
}

func (r *echoRunner) Run(ctx context.Context, in measurement.Input) (measurement.Output, error) {
	r.mu.Lock()
	r.calls++
	r.mu.Unlock()
	c, err := in.Dialer.DialContext(ctx, "tcp", r.address)
	if err != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	defer c.Close()
	buf := make([]byte, len(loopback.Payload))
	if _, err := c.Write(loopback.Payload); err != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	if _, err := io.ReadFull(c, buf); err != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	o, err := observation.DecodeProtocol(observation.Contract, []byte(`{"type":"tcp_connect","status":"succeeded","duration_ms":0,"details":{"address_family":"ipv4"}}`), observation.ProtocolLimits{})
	if err != nil {
		return measurement.Output{}, err
	}
	return measurement.Output{Outcome: "connected", Observations: []observation.Observation{o}}, nil
}

func (r *echoRunner) Calls() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls
}

// directHarness runs TCP-01 tasks with the echo runner through the production
// direct path, which bind derives from the fake platform network.
func directHarness(t *testing.T) (*harness, *echoRunner, *loopback.Server) {
	t.Helper()
	srv, err := loopback.Listen()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = srv.Close() })
	h := runningHarness(t)
	r := &echoRunner{address: srv.Address()}
	reg, err := registry.New([]registry.Entry{{Test: tcp01, Codec: tcpCodec{}, BaselineRisk: task.LowRisk, Targets: loopbackTargets(r.address), Proxy: registry.NoProxy{}, Runner: r}})
	if err != nil {
		t.Fatal("synthetic registry rejected")
	}
	h.adapters.registry = reg
	in := synPolicyInput(2, "reg-a", func(p *task.Permissions) { p.Tests = []task.TestVersion{tcp01} })
	in.Targets[0].Test = tcp01
	in.Targets[0].Tuple = *loopbackTargets(r.address).tuple
	h.adapters.policy, err = task.NewLocalPolicy(in)
	if err != nil {
		t.Fatal(err)
	}
	h.adapters.run.addressPolicy = loopbackAddressPolicy{}
	h.adapters.run.paths = nil
	return h, r, srv
}

// The production direct path (DEV-644): bind derives it from the network
// capability; a run dials a real loopback socket through a fresh direct
// dialer inside the metered attempt scope, and its result carries the access
// context the platform reported with egress unavailable on the direct path.
// A task whose network condition the platform does not satisfy returns
// run_not_due and opens no socket. Without a network capability there is no
// path, and initialization fails closed.
func TestStartRunOnTheDirectPath(t *testing.T) {
	h, r, srv := directHarness(t)
	h.run.network.Set(platform.AccessState{Generation: 7, Type: platform.NetworkCellular, VPN: platform.VPNNotDetected, IPv4: platform.AvailabilityYes, IPv6: platform.AvailabilityNo, Validated: platform.AvailabilityYes})
	c := h.initialized(t)
	if _, ok := c.bound.run.paths.(*networkPaths); !ok {
		t.Fatal("bind did not derive the direct path")
	}
	rh, err := c.StartRun(h.admit(t, c, runTask("direct-1", "", `"network":{"allowed_transports":["cellular"],"vpn_policy":"exclude","require_validated":true}`, false)))
	if err != nil {
		t.Fatal(codeOf(err))
	}
	waitTerminal(t, c, rh)
	recs := h.run.output.Records()
	if len(recs) != 1 || r.Calls() != 1 || srv.Accepted() != 1 {
		t.Fatalf("results %d, runner calls %d, accepted %d", len(recs), r.Calls(), srv.Accepted())
	}
	var res struct {
		Execution      struct{ State string } `json:"execution"`
		NetworkContext struct {
			Access map[string]any    `json:"access"`
			Egress map[string]string `json:"egress"`
		} `json:"network_context"`
	}
	if err := json.Unmarshal(recs[0], &res); err != nil {
		t.Fatal(err)
	}
	a, e := res.NetworkContext.Access, res.NetworkContext.Egress
	if res.Execution.State != "completed" || a["network_type"] != "cellular" || a["vpn_state"] != "not_detected" || a["ipv4_available"] != "yes" ||
		a["ipv6_available"] != "no" || a["validated"] != "yes" || a["network_changed"] != false ||
		len(e) != 3 || e["status"] != "unavailable" || e["lookup_source"] != "none" || e["network_path"] != "direct" {
		t.Fatalf("result %s", recs[0])
	}

	h, r, srv = directHarness(t)
	c = h.initialized(t)
	if _, err := c.StartRun(h.admit(t, c, runTask("direct-2", "", `"network":{"require_validated":true}`, false))); codeOf(err) != "run_not_due" {
		t.Fatal(codeOf(err))
	}
	if r.Calls() != 0 || srv.Accepted() != 0 || len(h.run.output.Records()) != 0 {
		t.Fatal("unsatisfied condition opened a socket or wrote a result")
	}

	h, _, _ = directHarness(t)
	h.adapters.run.network = nil
	if err := h.client(t).Initialize(`{}`); codeOf(err) != "not_initialized" {
		t.Fatal("initialized without a path")
	}
}

// An unreadable access context at reservation means no path: StartRun returns
// run_not_due and reserves nothing.
func TestStartRunWithoutAccessContext(t *testing.T) {
	h, r, _ := directHarness(t)
	c := h.initialized(t)
	th := h.admit(t, c, runTask("direct-3", "", "", false))
	h.run.network.Fail()
	if _, err := c.StartRun(th); codeOf(err) != "run_not_due" {
		t.Fatal(codeOf(err))
	}
	h.run.network.Set(platform.AccessState{Type: platform.NetworkWiFi}) // no generation
	if _, err := c.StartRun(th); codeOf(err) != "run_not_due" {
		t.Fatal(codeOf(err))
	}
	if r.Calls() != 0 || len(h.run.output.Records()) != 0 {
		t.Fatal("ran without access context")
	}
	h.run.network.Set(defaultAccess)
	rh, err := c.StartRun(th)
	if err != nil {
		t.Fatal(codeOf(err))
	}
	waitTerminal(t, c, rh)
	if r.Calls() != 1 {
		t.Fatal("did not run once access context was available")
	}
}

func loopbackTargets(address string) tcpTargets {
	host, port, _ := net.SplitHostPort(address)
	n, _ := strconv.Atoi(port)
	q := task.TargetTuple{Address: netip.MustParseAddr(host), Port: uint16(n)}
	return tcpTargets{tuple: &q}
}

type loopbackAddressPolicy struct{}

func (loopbackAddressPolicy) Allows(a netip.Addr) bool {
	return a.IsLoopback() || task.PublicAddress(a)
}
