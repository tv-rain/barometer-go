package barometer

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/tv-rain/barometer-go/internal/config"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/registry"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

// proxyCompatible is a synthetic rule that accepts every parameter value, so
// a TCP-01-shaped synthetic task may select the proxy path. It is not the
// production TCP-01 rule (DEV-454).
type proxyCompatible struct{}

func (proxyCompatible) Compatible(task.Parameters) bool { return true }

// proxyHarness runs TCP-01-shaped tasks with the echo runner through the
// production path source, with a registry rule that permits the proxy and a
// reviewed connection_code approval. The policy permits both paths.
func proxyHarness(t *testing.T) (*harness, *echoRunner, *loopback.Server) {
	t.Helper()
	h, r, srv := directHarness(t)
	reg, err := registry.New([]registry.Entry{{Test: tcp01, Codec: tcpCodec{}, BaselineRisk: task.LowRisk, Targets: tcpTargets{}, Proxy: proxyCompatible{}, Runner: r}})
	if err != nil {
		t.Fatal("synthetic registry rejected")
	}
	h.adapters.registry = reg
	h.adapters.approvals = config.Approvals{ConnectionCodes: []string{"outline_mobileproxy"}}
	return h, r, srv
}

func proxyConfig(port int) string {
	return fmt.Sprintf(`{"proxy":{"host":"127.0.0.1","port":%d,"connection_code":"outline_mobileproxy"}}`, port)
}

type proxiedResult struct {
	Execution struct {
		State  string `json:"state"`
		Reason string `json:"reason"`
	} `json:"execution"`
	NetworkContext *struct {
		Egress map[string]string `json:"egress"`
	} `json:"network_context"`
}

// runOnce admits taskJSON, starts its run and returns its single result.
func runOnce(t *testing.T, h *harness, configJSON, taskJSON string) (proxiedResult, []byte, *Client) {
	t.Helper()
	c := h.client(t)
	if err := c.Initialize(configJSON); err != nil {
		t.Fatal("initialize rejected:", err)
	}
	rh, err := c.StartRun(h.admit(t, c, taskJSON))
	if err != nil {
		t.Fatal(codeOf(err))
	}
	waitTerminal(t, c, rh)
	recs := h.run.output.Records()
	if len(recs) != 1 {
		t.Fatalf("results = %d", len(recs))
	}
	var res proxiedResult
	if err := json.Unmarshal(recs[0], &res); err != nil {
		t.Fatal(err)
	}
	return res, recs[0], c
}

// A proxy task runs through the proxy configured at Initialize: its stream is
// one tunnel of the loopback endpoint, the target is reached only through it,
// and the result records the proxy path and connection_code and nothing about
// the endpoint or the CONNECT exchange. A direct task on the same client is
// unaffected by the proxy.
// requirement-ref: REQ-PROXY-0012
// requirement-ref: REQ-PROXY-0022
// requirement-ref: REQ-PROXY-0031
// requirement-ref: REQ-PROXY-0032
// requirement-ref: REQ-PROXY-0033
func TestStartRunOnTheProxyPath(t *testing.T) {
	h, r, srv := proxyHarness(t)
	p, err := loopback.ListenProxy(loopback.ProxyConfig{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = p.Close() })
	res, raw, c := runOnce(t, h, proxyConfig(p.Port()), runTask("proxy-1", "", `"network_path":"proxy"`, false))
	e := res.NetworkContext.Egress
	if res.Execution.State != "completed" || len(e) != 4 || e["network_path"] != "proxy" || e["connection_code"] != "outline_mobileproxy" || e["status"] != "unavailable" {
		t.Fatalf("result %s", raw)
	}
	bundle, err := c.NextBundle()
	if err != nil || !strings.Contains(bundle, `"connection_code":"outline_mobileproxy"`) {
		t.Fatal("no bundle with the proxied result:", err)
	}
	for _, leak := range []string{strconv.Itoa(p.Port()), "127.0.0.1", "CONNECT", "HTTP/1.1"} {
		if strings.Contains(string(raw), leak) || strings.Contains(bundle, leak) {
			t.Fatalf("result or bundle contains %q: %s", leak, raw)
		}
	}
	if c := p.Counts(); r.Calls() != 1 || srv.Accepted() != 1 || c.Relays != 1 || len(c.Requests) != 1 || !strings.HasPrefix(c.Requests[0], "CONNECT "+r.address+" HTTP/1.1\r\n") {
		t.Fatalf("runner calls %d, proxy %+v", r.Calls(), c)
	}

	h, _, _ = proxyHarness(t)
	res, raw, _ = runOnce(t, h, proxyConfig(p.Port()), runTask("direct-1", "", "", false))
	if e := res.NetworkContext.Egress; res.Execution.State != "completed" || e["network_path"] != "direct" || len(e) != 3 || p.Counts().Accepted != 1 {
		t.Fatalf("direct result %s with %d proxy accepts", raw, p.Counts().Accepted)
	}
}

// A proxy task without a configured proxy is started and records not_run
// with proxy_not_configured, making no contact and never using the direct
// path; a configured but unreachable endpoint is proxy_unavailable after the
// test starts.
// requirement-ref: REQ-EXECUTION-0008
// requirement-ref: REQ-EXECUTION-0010
// requirement-ref: REQ-PROXY-0028
// requirement-ref: REQ-PROXY-0030
func TestProxyPathFailuresAreRecorded(t *testing.T) {
	h, r, srv := proxyHarness(t)
	res, raw, _ := runOnce(t, h, `{}`, runTask("proxy-2", "", `"network_path":"proxy"`, false))
	if res.Execution.State != "not_run" || res.Execution.Reason != "proxy_not_configured" || res.NetworkContext != nil || r.Calls() != 0 || srv.Accepted() != 0 {
		t.Fatalf("result %s, runner calls %d, direct accepts %d", raw, r.Calls(), srv.Accepted())
	}

	h, r, srv = proxyHarness(t)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()
	res, raw, _ = runOnce(t, h, proxyConfig(port), runTask("proxy-3", "", `"network_path":"proxy"`, false))
	if res.Execution.State != "incomplete" || res.Execution.Reason != "proxy_unavailable" || res.NetworkContext == nil ||
		res.NetworkContext.Egress["network_path"] != "proxy" || r.Calls() != 1 || srv.Accepted() != 0 {
		t.Fatalf("result %s, direct accepts %d", raw, srv.Accepted())
	}
}
