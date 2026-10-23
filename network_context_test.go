package barometer

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/tv-rain/barometer-go/internal/config"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/registry"
	"github.com/tv-rain/barometer-go/internal/testkit/contextendpoint"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

// Egress Network Context through the root facade (DEV-647). The endpoints
// are synthetic HTTPS servers on 127.0.0.1 with certificates from a
// synthetic CA, reached on the direct path through the routes test seam and
// on the proxy path through a scripted loopback CONNECT proxy. No real
// endpoint is contacted; these are synthetic cases, not field evidence.

const (
	contextPrimary  = "https://context.example/network-context"
	contextFallback = "https://backup.example/network-context"
	contextConfig   = `{"network_context":{"primary_endpoint":"` + contextPrimary + `","fallback_endpoint":"` + contextFallback + `"}}`
	nlClassified    = `{"country_code":"NL","region_code":"NL-NH","asn":64500,"classification_version":"0.1.0"}`
)

type contextEndpoints struct {
	primary, fallback *contextendpoint.Server
	routes            *contextendpoint.Routes
}

func (e contextEndpoints) requests() int {
	return len(e.primary.Requests()) + len(e.fallback.Requests())
}

// contextHarness is directHarness with approved context endpoints, reviewed
// classification approvals and the synthetic CA as trust roots.
func contextHarness(t *testing.T, primary contextendpoint.Response) (*harness, *echoRunner, contextEndpoints) {
	t.Helper()
	h, r, _ := directHarness(t)
	ca, err := contextendpoint.NewCA()
	if err != nil {
		t.Fatal(err)
	}
	var e contextEndpoints
	if e.primary, err = contextendpoint.Listen(ca, "context.example", primary); err != nil {
		t.Fatal(err)
	}
	if e.fallback, err = contextendpoint.Listen(ca, "backup.example", contextendpoint.Response{Status: 503}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = e.primary.Close(); _ = e.fallback.Close() })
	e.routes = contextendpoint.NewRoutes(map[string]string{"context.example:443": e.primary.Address(), "backup.example:443": e.fallback.Address()})
	h.adapters.approvals = config.Approvals{Endpoints: []string{contextPrimary, contextFallback}, ConnectionCodes: []string{"outline_mobileproxy"}}
	h.adapters.contextApprovals = observation.ContextApprovals{ClassificationVersions: []string{"0.1.0"}, CountryCodes: []string{"NL"}, RegionCodes: []string{"NL-NH"}}
	h.adapters.contextRoots = ca.Pool()
	h.adapters.run.directDialers = e.routes
	return h, r, e
}

type contextResult struct {
	Execution struct {
		State  string `json:"state"`
		Reason string `json:"reason"`
	} `json:"execution"`
	NetworkContext *struct {
		Egress map[string]any `json:"egress"`
	} `json:"network_context"`
}

// startAndWait admits taskJSON on c, runs it and returns its result.
func startAndWait(t *testing.T, h *harness, c *Client, taskJSON string) contextResult {
	t.Helper()
	before := len(h.run.output.Records())
	rh, err := c.StartRun(h.admit(t, c, taskJSON))
	if err != nil {
		t.Fatal(codeOf(err))
	}
	waitTerminal(t, c, rh)
	recs := h.run.output.Records()
	if len(recs) != before+1 {
		t.Fatalf("results = %d", len(recs))
	}
	var res contextResult
	if err := json.Unmarshal(recs[len(recs)-1], &res); err != nil {
		t.Fatal(err)
	}
	return res
}

func initializeWith(t *testing.T, h *harness, configJSON string) *Client {
	t.Helper()
	c := h.client(t)
	if err := c.Initialize(configJSON); err != nil {
		t.Fatal("initialize rejected:", err)
	}
	return c
}

// CC-CONTEXT-HOST-INITIALIZATION: the host configures egress classification
// once at Initialize with reviewed, approved endpoints; without it Barometer
// reports access context only and contacts no endpoint; the configuration
// cannot be changed for an initialized client, and no task or later call can
// add an endpoint; invalid, unapproved or credential-bearing endpoints and
// unknown fields fail initialization; a redirect is not followed.
// requirement-ref: REQ-CONTEXT-0007
// requirement-ref: REQ-CONTEXT-0008
// requirement-ref: REQ-CONTEXT-0009
// requirement-ref: REQ-CONTEXT-0010
func TestCcContextHostInitialization(t *testing.T) {
	h, _, e := contextHarness(t, contextendpoint.Response{Body: nlClassified})
	c := initializeWith(t, h, contextConfig)
	res := startAndWait(t, h, c, runTask("context-1", "", "", false))
	if eg := res.NetworkContext.Egress; res.Execution.State != "completed" || eg["status"] != "fresh" || eg["lookup_source"] != "primary" || eg["country_code"] != "NL" {
		t.Fatalf("egress %v", eg)
	}

	// Absent configuration: access context only, no endpoint contact, even
	// for a task that requires a field it then cannot have.
	h, _, e = contextHarness(t, contextendpoint.Response{Body: nlClassified})
	c = initializeWith(t, h, `{}`)
	res = startAndWait(t, h, c, runTask("context-2", "", "", false))
	if eg := res.NetworkContext.Egress; eg["status"] != "unavailable" || e.requests() != 0 {
		t.Fatalf("egress %v with %d requests", eg, e.requests())
	}
	res = startAndWait(t, h, c, runTask("context-3", "", `"network":{"context":{"required_fields":["country_code"]}}`, false))
	if res.Execution.State != "not_run" || res.Execution.Reason != "network_context_unavailable" || e.requests() != 0 {
		t.Fatalf("result %+v with %d requests", res.Execution, e.requests())
	}

	for name, cfg := range map[string]string{
		"unapproved":      `{"network_context":{"primary_endpoint":"https://other.example/network-context"}}`,
		"http":            `{"network_context":{"primary_endpoint":"http://context.example/network-context"}}`,
		"credentials":     `{"network_context":{"primary_endpoint":"https://user:secret@context.example/network-context"}}`,
		"fragment":        `{"network_context":{"primary_endpoint":"https://context.example/network-context#x"}}`,
		"no primary":      `{"network_context":{"fallback_endpoint":"` + contextFallback + `"}}`,
		"unknown field":   `{"network_context":{"primary_endpoint":"` + contextPrimary + `","discovery":true}}`,
		"fallback list":   `{"network_context":{"primary_endpoint":"` + contextPrimary + `","fallback_endpoint":["` + contextFallback + `"]}}`,
		"loopback":        `{"network_context":{"primary_endpoint":"https://127.0.0.1/network-context"}}`,
		"second fallback": `{"network_context":{"primary_endpoint":"` + contextPrimary + `","fallback_endpoint":"` + contextFallback + `","second_fallback_endpoint":"` + contextFallback + `"}}`,
	} {
		h, _, _ := contextHarness(t, contextendpoint.Response{Body: nlClassified})
		if err := h.client(t).Initialize(cfg); err == nil {
			t.Fatalf("%s: initialized", name)
		}
	}

	// No public method changes the endpoints of an initialized client.
	for i := range reflect.TypeOf(&Client{}).NumMethod() {
		m := reflect.TypeOf(&Client{}).Method(i).Name
		if strings.Contains(m, "Context") || strings.Contains(m, "Endpoint") {
			t.Fatalf("public method %s may change the context configuration", m)
		}
	}

	// A redirect is not followed: the redirect target sees no request.
	h, _, e = contextHarness(t, contextendpoint.Response{Status: 302, Header: map[string]string{"Location": "https://redirect.example/network-context"}})
	redirected, err := contextendpoint.Listen(mustCA(t), "redirect.example", contextendpoint.Response{Body: nlClassified})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = redirected.Close() })
	e.routes.Add("redirect.example:443", redirected.Address())
	c = initializeWith(t, h, `{"network_context":{"primary_endpoint":"`+contextPrimary+`"}}`)
	res = startAndWait(t, h, c, runTask("context-4", "", "", false))
	if eg := res.NetworkContext.Egress; eg["status"] != "unavailable" || len(redirected.Requests()) != 0 || len(e.primary.Requests()) != 1 {
		t.Fatalf("egress %v after a redirect", eg)
	}
}

func mustCA(t *testing.T) *contextendpoint.CA {
	t.Helper()
	ca, err := contextendpoint.NewCA()
	if err != nil {
		t.Fatal(err)
	}
	return ca
}

// The egress cache is in memory and belongs to one initialized generation:
// later runs on the same generation and path reuse it; clearing,
// reinitialization, policy replacement and shutdown drop it, and the next
// run asks the endpoint again (REQ-CONTEXT-0023, REQ-CONTEXT-0024).
// TestClearingCoversEveryAdapter covers its clearing registration.
// requirement-ref: REQ-CONTEXT-0023
// requirement-ref: REQ-CONTEXT-0024
func TestEgressCacheLifecycle(t *testing.T) {
	h, _, e := contextHarness(t, contextendpoint.Response{Body: nlClassified})
	c := initializeWith(t, h, contextConfig)
	cache := c.bound.run.contextCache
	startAndWait(t, h, c, runTask("cache-1", "", "", false))
	if res := startAndWait(t, h, c, runTask("cache-2", "", "", false)); res.NetworkContext.Egress["lookup_source"] != "cache" || len(e.primary.Requests()) != 1 || cache.Len() != 1 {
		t.Fatalf("second run egress %v, %d requests", res.NetworkContext.Egress, len(e.primary.Requests()))
	}

	if err := c.ClearMeasurementData(); err != nil || cache.Len() != 0 {
		t.Fatalf("clearing: %v, %d entries", err, cache.Len())
	}
	if err := c.Initialize(contextConfig); err != nil {
		t.Fatal(err)
	}
	if res := startAndWait(t, h, c, runTask("cache-3", "", "", false)); res.NetworkContext.Egress["lookup_source"] != "primary" {
		t.Fatalf("after reinitialization %v", res.NetworkContext.Egress)
	}

	in := synPolicyInput(3, "reg-a", func(p *task.Permissions) { p.Tests = []task.TestVersion{tcp01} })
	in.Targets[0].Test = tcp01
	next, err := task.NewLocalPolicy(in)
	if err != nil {
		t.Fatal(err)
	}
	startAndWait(t, h, c, runTask("cache-5", "", "", false))
	if cache.Len() != 1 {
		t.Fatal("no entry before policy replacement")
	}
	if err := c.replacePolicy(next); err != nil || cache.Len() != 0 {
		t.Fatalf("policy replacement: %v, %d entries", err, cache.Len())
	}
	b := runTask("cache-4", "", "", false)
	th, err := c.AdmitTask(b, h.embeddedAt(t, 3, b))
	if err != nil {
		t.Fatal(codeOf(err))
	}
	rh, err := c.StartRun(th)
	if err != nil {
		t.Fatal(codeOf(err))
	}
	waitTerminal(t, c, rh)
	recs := h.run.output.Records()
	var res contextResult
	if err := json.Unmarshal(recs[len(recs)-1], &res); err != nil || res.NetworkContext.Egress["lookup_source"] != "primary" {
		t.Fatalf("after policy replacement %s", recs[len(recs)-1])
	}
	// The new generation caches again; the old one's shutdown did not
	// disable it.
	b = runTask("cache-6", "", "", false)
	if th, err = c.AdmitTask(b, h.embeddedAt(t, 3, b)); err != nil {
		t.Fatal(codeOf(err))
	}
	if rh, err = c.StartRun(th); err != nil {
		t.Fatal(codeOf(err))
	}
	waitTerminal(t, c, rh)
	recs = h.run.output.Records()
	if err := json.Unmarshal(recs[len(recs)-1], &res); err != nil || res.NetworkContext.Egress["lookup_source"] != "cache" {
		t.Fatalf("second run after policy replacement %s", recs[len(recs)-1])
	}

	if err := c.Close(); err != nil || cache.Len() != 0 {
		t.Fatalf("close: %v, %d entries", err, cache.Len())
	}
	if len(e.primary.Requests()) != 3 {
		t.Fatalf("%d requests", len(e.primary.Requests()))
	}
}

// Proxy egress goes only through the configured proxy; replacing or clearing
// the proxy drops the previous generation's proxy entries and keeps the
// direct entry, and the next proxy run asks again through the new proxy
// (REQ-PROXY-0016, REQ-PROXY-0023, REQ-PROXY-0024).
// requirement-ref: REQ-PROXY-0016
func TestEgressCacheProxyChange(t *testing.T) {
	h, r, e := contextHarness(t, contextendpoint.Response{Body: nlClassified})
	reg, err := proxyRegistry(r)
	if err != nil {
		t.Fatal(err)
	}
	h.adapters.registry = reg
	resolve := map[string]string{"context.example:443": e.primary.Address(), "backup.example:443": e.fallback.Address()}
	first, err := loopback.ListenProxy(loopback.ProxyConfig{Resolve: resolve})
	if err != nil {
		t.Fatal(err)
	}
	second, err := loopback.ListenProxy(loopback.ProxyConfig{Resolve: resolve})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = first.Close(); _ = second.Close() })

	c := initializeWith(t, h, `{"proxy":{"host":"127.0.0.1","port":`+strconv.Itoa(first.Port())+`,"connection_code":"outline_mobileproxy"},"network_context":{"primary_endpoint":"`+contextPrimary+`"}}`)
	cache := c.bound.run.contextCache
	startAndWait(t, h, c, runTask("direct-1", "", "", false))
	directDials := len(e.routes.Dialed())
	res := startAndWait(t, h, c, runTask("proxy-1", "", `"network_path":"proxy"`, false))
	if eg := res.NetworkContext.Egress; eg["status"] != "fresh" || eg["network_path"] != "proxy" || eg["connection_code"] != "outline_mobileproxy" || cache.Len() != 2 {
		t.Fatalf("proxy egress %v, %d entries", eg, cache.Len())
	}
	if n := len(e.routes.Dialed()); n != directDials {
		t.Fatalf("a proxy run dialed the direct path: %q", e.routes.Dialed())
	}
	if !strings.HasPrefix(first.Counts().Requests[0], "CONNECT context.example:443 ") {
		t.Fatalf("CONNECT requests %q", first.Counts().Requests)
	}

	if err := c.SetProxy(`{"host":"127.0.0.1","port":` + strconv.Itoa(second.Port()) + `,"connection_code":"outline_mobileproxy"}`); err != nil || cache.Len() != 1 {
		t.Fatalf("SetProxy: %v, %d entries", err, cache.Len())
	}
	res = startAndWait(t, h, c, runTask("proxy-2", "", `"network_path":"proxy"`, false))
	if eg := res.NetworkContext.Egress; eg["lookup_source"] != "primary" || len(second.Counts().Requests) == 0 || !strings.HasPrefix(second.Counts().Requests[0], "CONNECT context.example:443 ") {
		t.Fatalf("after SetProxy %v, %q", eg, second.Counts().Requests)
	}
	if res := startAndWait(t, h, c, runTask("direct-2", "", "", false)); res.NetworkContext.Egress["lookup_source"] != "cache" {
		t.Fatalf("direct entry lost: %v", res.NetworkContext.Egress)
	}
	if err := c.ClearProxy(); err != nil || cache.Len() != 1 {
		t.Fatalf("ClearProxy: %v, %d entries", err, cache.Len())
	}
}

func proxyRegistry(r *echoRunner) (*registry.Registry, error) {
	return registry.New([]registry.Entry{{Test: tcp01, Codec: tcpCodec{}, BaselineRisk: task.LowRisk, Targets: tcpTargets{}, Proxy: proxyCompatible{}, Runner: r}})
}
