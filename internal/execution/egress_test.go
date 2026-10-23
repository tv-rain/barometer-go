package execution_test

import (
	"encoding/json"
	"net"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/execution"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/networkcontext"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit"
	"github.com/tv-rain/barometer-go/internal/testkit/contextendpoint"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

// Egress Network Context through the engine (DEV-647). The endpoints are
// synthetic HTTPS servers on 127.0.0.1 with certificates from a synthetic CA;
// the direct path reaches them through contextendpoint.Routes, which stands
// in for name resolution, and the proxy path through a scripted loopback
// CONNECT proxy. No real endpoint is contacted. These are synthetic core
// contract cases, not field or platform evidence, and an observed context or
// endpoint failure is not evidence of censorship.

const (
	primaryName  = "context.example"
	fallbackName = "backup.example"
	nlBody       = `{"country_code":"NL","region_code":"NL-NH","asn":64500,"classification_version":"0.1.0"}`
)

var classificationApprovals = observation.ContextApprovals{ClassificationVersions: []string{"0.1.0"}, CountryCodes: []string{"NL", "DE"}, RegionCodes: []string{"NL-NH"}}

// endpointPair is a primary and a fallback synthetic endpoint, and the
// direct-path routes to them.
type endpointPair struct {
	ca                *contextendpoint.CA
	primary, fallback *contextendpoint.Server
	routes            *contextendpoint.Routes
}

func newEndpoints(t *testing.T, primary, fallback contextendpoint.Response) *endpointPair {
	t.Helper()
	ca, err := contextendpoint.NewCA()
	if err != nil {
		t.Fatal(err)
	}
	e := &endpointPair{ca: ca}
	if e.primary, err = contextendpoint.Listen(ca, primaryName, primary); err != nil {
		t.Fatal(err)
	}
	if e.fallback, err = contextendpoint.Listen(ca, fallbackName, fallback); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = e.primary.Close(); _ = e.fallback.Close() })
	e.routes = contextendpoint.NewRoutes(map[string]string{primaryName + ":443": e.primary.Address(), fallbackName + ":443": e.fallback.Address()})
	return e
}

func ok(body string) contextendpoint.Response { return contextendpoint.Response{Body: body} }

// withEgress binds the fixture's direct path to a lookup over e on the
// fixture's engine clock and returns the lookup's cache. The attempts' dialers
// are e's routes, so the lookup reaches the endpoints by name.
func (f *fixture) withEgress(t *testing.T, e *endpointPair, fallback bool) *networkcontext.Cache {
	t.Helper()
	cache := networkcontext.NewCache()
	f.ctx.direct = bindLookup(t, f.ctx.direct, e, fallback, f.clock, cache)
	f.plan.Result = resultDecoder(t, f.ctx.direct, 0)
	f.plan.Dialers = e.routes
	return cache
}

func bindLookup(t *testing.T, src *networkcontext.Source, e *endpointPair, fallback bool, clock networkcontext.Clock, cache *networkcontext.Cache) *networkcontext.Source {
	t.Helper()
	ep := networkcontext.Endpoints{Primary: "https://" + primaryName + "/network-context"}
	if fallback {
		ep.Fallback = "https://" + fallbackName + "/network-context"
	}
	l, err := networkcontext.NewLookup(ep, classificationApprovals, e.ca.Pool(), clock, cache)
	if err != nil {
		t.Fatal(err)
	}
	src, err = src.WithLookup(l)
	if err != nil {
		t.Fatal(err)
	}
	return src
}

func resultDecoder(t *testing.T, src *networkcontext.Source, proxy uint64) *result.Decoder {
	t.Helper()
	name, code := "direct", ""
	if proxy != 0 {
		name, code = "proxy", connectionCode
	}
	p, err := observation.NewPath(1, proxy, name, code)
	if err != nil {
		t.Fatal(err)
	}
	d, err := result.New(result.Binding{TaskID: "campaign-1", Output: result.Output{Test: model.Test{ID: tcp01.ID, Version: tcp01.Version, ResultVersion: result.Contract}},
		MaxAttempts: 2, MaxDurationMS: 60000, Implementation: implementation, ContextDecoder: src.Decoder(), Path: p, AccessGeneration: 1})
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func egressOf(t *testing.T, r record) map[string]any {
	t.Helper()
	var nc struct{ Egress map[string]any }
	if err := json.Unmarshal(r.NetworkContext, &nc); err != nil {
		t.Fatalf("network_context %s: %v", r.NetworkContext, err)
	}
	return nc.Egress
}

func runOnce(t *testing.T, f *fixture) []record {
	t.Helper()
	if _, err := f.engine.Run(f.run, f.plan); err != nil {
		t.Fatal(err)
	}
	return f.records(t)
}

// waitFor polls a test-local condition; it never waits on the engine clock.
func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal("condition not reached")
		}
		time.Sleep(time.Millisecond)
	}
}

// CC-CONTEXT-ENDPOINT-CONTRACT, request side: one HTTPS GET with Accept:
// application/json to the fixed URL, carrying only Host and Accept: no
// User-Agent, compression, cookie, authorization, identifier, address,
// target, task detail or connection_code. The response bounds run in
// TestContextEndpointResponses and in internal/networkcontext.
// requirement-ref: REQ-CONTEXT-0011
// requirement-ref: REQ-CONTEXT-0012
// requirement-ref: REQ-CONTEXT-0036
func TestCcContextEndpointContract(t *testing.T) {
	e := newEndpoints(t, ok(nlBody), ok(nlBody))
	f := newFixture(t, schedule(1), connected(t))
	f.withEgress(t, e, true)
	recs := runOnce(t, f)
	if len(recs) != 1 || recs[0].Execution.State != "completed" {
		t.Fatalf("records %+v", recs)
	}
	reqs := e.primary.Requests()
	if len(reqs) != 1 || len(e.fallback.Requests()) != 0 {
		t.Fatalf("%d primary, %d fallback requests", len(reqs), len(e.fallback.Requests()))
	}
	r := reqs[0]
	var names []string
	for k := range r.Header {
		names = append(names, k)
	}
	if r.Method != "GET" || r.Path != "/network-context" || r.RawQuery != "" || r.Host != primaryName ||
		!slices.Equal(names, []string{"Accept"}) || r.Header.Get("Accept") != "application/json" {
		t.Fatalf("request %+v", r)
	}
	raw := string(e.primary.Raw())
	if !strings.HasPrefix(raw, "GET /network-context HTTP/1.1\r\nHost: context.example\r\n") {
		t.Fatalf("request head %q", raw)
	}
	for _, leak := range []string{"User-Agent", "Accept-Encoding", "Cookie", "Authorization", "campaign-1", "TCP-01", connectionCode, "127.0.0.1", "run", "attempt"} {
		if strings.Contains(raw, leak) {
			t.Fatalf("request carries %q: %q", leak, raw)
		}
	}
}

// TLS, HTTP and size failures of the primary are invalid responses: the one
// fallback request through the same dialer supplies the context. A redirect
// is not followed, a body of exactly 4,096 bytes is accepted, and the
// operation bound of 5,000 ms on the engine clock stops a stalled primary.
// requirement-ref: REQ-CONTEXT-0010
// requirement-ref: REQ-CONTEXT-0015
// requirement-ref: REQ-CONTEXT-0020
// requirement-ref: REQ-CONTEXT-0021
func TestContextEndpointResponses(t *testing.T) {
	deBody := `{"country_code":"DE","classification_version":"0.1.0"}`
	for _, tc := range []struct {
		name    string
		primary contextendpoint.Response
		setup   func(*testing.T, *endpointPair)
	}{
		{name: "server error", primary: contextendpoint.Response{Status: 500, Body: nlBody}},
		{name: "redirect", primary: contextendpoint.Response{Status: 302, Header: map[string]string{"Location": "https://" + fallbackName + "/network-context"}}},
		{name: "not json", primary: contextendpoint.Response{ContentType: "text/html", Body: nlBody}},
		{name: "compressed", primary: contextendpoint.Response{Header: map[string]string{"Content-Encoding": "gzip"}, Body: nlBody}},
		{name: "empty", primary: ok("")},
		{name: "malformed", primary: ok(`{"country_code":`)},
		{name: "oversized", primary: ok(`{"classification_version":"0.1.0"}` + strings.Repeat(" ", networkcontext.MaxResponseBytes))},
		{name: "unknown CA", primary: ok(nlBody), setup: func(t *testing.T, e *endpointPair) {
			other, _ := contextendpoint.NewCA()
			s, err := contextendpoint.Listen(other, primaryName, ok(nlBody))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = s.Close() })
			e.routes.Add(primaryName+":443", s.Address())
		}},
		{name: "wrong name", primary: ok(nlBody), setup: func(t *testing.T, e *endpointPair) {
			s, err := contextendpoint.Listen(e.ca, "other.example", ok(nlBody))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = s.Close() })
			e.routes.Add(primaryName+":443", s.Address())
		}},
		{name: "unreachable", primary: ok(nlBody), setup: func(t *testing.T, e *endpointPair) {
			e.routes.Add(primaryName+":443", "127.0.0.1:"+strconv.Itoa(closedPort(t)))
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := newEndpoints(t, tc.primary, ok(deBody))
			if tc.setup != nil {
				tc.setup(t, e)
			}
			f := newFixture(t, schedule(1), connected(t))
			f.withEgress(t, e, true)
			recs := runOnce(t, f)
			eg := egressOf(t, recs[0])
			if recs[0].Execution.State != "completed" || eg["status"] != "fresh" || eg["lookup_source"] != "fallback" || eg["country_code"] != "DE" {
				t.Fatalf("egress %v", eg)
			}
			if n := len(e.fallback.Requests()); n != 1 {
				t.Fatalf("%d fallback requests", n)
			}
			if n := len(e.primary.Requests()); n > 1 {
				t.Fatalf("primary retried: %d requests", n)
			}
		})
	}

	t.Run("exactly 4096 bytes", func(t *testing.T) {
		body := `{"classification_version":"0.1.0"}`
		e := newEndpoints(t, ok(body+strings.Repeat(" ", networkcontext.MaxResponseBytes-len(body))), ok(deBody))
		f := newFixture(t, schedule(1), connected(t))
		f.withEgress(t, e, true)
		if eg := egressOf(t, runOnce(t, f)[0]); eg["lookup_source"] != "primary" {
			t.Fatalf("egress %v", eg)
		}
	})

	t.Run("stalled primary", func(t *testing.T) {
		e := newEndpoints(t, contextendpoint.Response{Hold: true}, ok(deBody))
		f := newFixture(t, schedule(1), connected(t))
		f.withEgress(t, e, true)
		done := f.start(t, f.run, f.runner)
		waitFor(t, func() bool { return len(e.primary.Requests()) == 1 })
		_ = f.clock.Advance(networkcontext.OperationTimeout - time.Millisecond)
		if len(e.fallback.Requests()) != 0 {
			t.Fatal("primary stopped before its bound")
		}
		_ = f.clock.Advance(time.Millisecond)
		o := <-done
		if o.err != nil {
			t.Fatal(o.err)
		}
		if eg := egressOf(t, f.records(t)[0]); eg["lookup_source"] != "fallback" {
			t.Fatalf("egress %v", eg)
		}
	})
}

// CC-CONTEXT-EXECUTION-INTEGRATION: acquisition runs inside the attempt
// deadline, cancellation and network-change watch; a missing required field
// is network_context_unavailable and an allowlist mismatch
// constraints_unsatisfied, both without running the test; without a
// requirement an unavailable context does not block the attempt. Platform
// suspension during acquisition needs the native lifecycle signal (DEV-455).
// requirement-ref: REQ-CONTEXT-0025
// requirement-ref: REQ-CONTEXT-0026
// requirement-ref: REQ-CONTEXT-0027
// requirement-ref: REQ-CONTEXT-0028
func TestCcContextExecutionIntegration(t *testing.T) {
	interrupt := map[string]struct {
		act    func(*fixture)
		reason string
	}{
		"host cancellation": {func(f *fixture) { f.runner.cancel() }, "cancelled_by_host"},
		"attempt deadline":  {func(f *fixture) { _ = f.clock.Advance(attemptDeadline) }, "attempt_deadline_exceeded"},
		"network change": {func(f *fixture) {
			s := defaultAccess
			s.Generation = 2
			f.ctx.net.Set(s)
		}, "network_changed"},
	}
	for name, tc := range interrupt {
		t.Run(name, func(t *testing.T) {
			e := newEndpoints(t, contextendpoint.Response{Hold: true}, ok(nlBody))
			f := newFixture(t, schedule(1), connected(t))
			f.withEgress(t, e, true)
			done := f.start(t, f.run, f.runner)
			waitFor(t, func() bool { return len(e.primary.Requests()) == 1 })
			tc.act(f)
			if o := <-done; o.err != nil {
				t.Fatal(o.err)
			}
			recs := f.records(t)
			if len(recs) != 1 || recs[0].Execution.State != "not_run" || recs[0].Execution.Reason != tc.reason || f.runner.Calls() != 0 || len(e.fallback.Requests()) != 0 {
				t.Fatalf("records %+v, runner calls %d, fallback %d", recs, f.runner.Calls(), len(e.fallback.Requests()))
			}
		})
	}

	for _, tc := range []struct {
		name           string
		primary        contextendpoint.Response
		required       []string
		allowed        []string
		state, reason  string
		egressStatus   string
		runnerExpected bool
	}{
		{"required fields present", ok(nlBody), []string{"country_code", "region_code", "asn"}, nil, "completed", "", "fresh", true},
		{"required field unknown", ok(`{"country_code":"NL","classification_version":"0.1.0"}`), []string{"asn"}, nil, "not_run", "network_context_unavailable", "", false},
		{"required field, endpoints failing", contextendpoint.Response{Status: 503}, []string{"country_code"}, nil, "not_run", "network_context_unavailable", "", false},
		{"allowlist match", ok(nlBody), nil, []string{"NL"}, "completed", "", "fresh", true},
		{"allowlist mismatch", ok(nlBody), nil, []string{"DE"}, "not_run", "constraints_unsatisfied", "", false},
		{"allowlist, country unknown", ok(`{"classification_version":"0.1.0"}`), nil, []string{"NL"}, "not_run", "network_context_unavailable", "", false},
		{"no requirement, endpoints failing", contextendpoint.Response{Status: 503}, nil, nil, "completed", "", "unavailable", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fb := tc.primary
			e := newEndpoints(t, tc.primary, fb)
			f := newFixture(t, schedule(1), connected(t))
			f.withEgress(t, e, true)
			f.plan.Network.Context = task.ContextRequirements{RequiredFields: tc.required, AllowedCountryCodes: tc.allowed}
			recs := runOnce(t, f)
			r := recs[0]
			if len(recs) != 1 || r.Execution.State != tc.state || r.Execution.Reason != tc.reason || (f.runner.Calls() == 1) != tc.runnerExpected {
				t.Fatalf("records %+v, runner calls %d", recs, f.runner.Calls())
			}
			if tc.egressStatus != "" && egressOf(t, r)["status"] != tc.egressStatus {
				t.Fatalf("egress %v", egressOf(t, r))
			}
		})
	}

	// network_context_unavailable is retryable when the schedule lists it:
	// the retry acquires again (a failure is not cached) and runs once the
	// field is available.
	t.Run("retry after unavailable context", func(t *testing.T) {
		e := newEndpoints(t, ok(nlBody), contextendpoint.Response{Status: 503})
		e.primary.Queue(contextendpoint.Response{Status: 503})
		f := newFixture(t, schedule(2, task.NetworkContextUnavailable), connected(t))
		f.withEgress(t, e, true)
		f.plan.Network.Context.RequiredFields = []string{"country_code"}
		rep, err := f.engine.Run(f.run, f.plan)
		f.expect(t, rep, err, nil, []want{{"not_run", "network_context_unavailable", ""}, completedConnected}, f.clock.Waits())
		if len(f.clock.Waits()) != 1 || len(e.primary.Requests()) != 2 {
			t.Fatalf("waits %v, %d primary requests", f.clock.Waits(), len(e.primary.Requests()))
		}
	})

	// An allowlist mismatch in a retry ends the run without a result, like
	// an access condition (REQ-EXECUTION-0031).
	t.Run("mismatch prevents a retry", func(t *testing.T) {
		e := newEndpoints(t, ok(nlBody), ok(nlBody))
		e.primary.Queue(ok(`{"country_code":"DE","classification_version":"0.1.0"}`))
		fail := failure()
		f := newFixture(t, schedule(2, task.InternalError), fail, connected(t))
		cache := f.withEgress(t, e, true)
		// The first attempt's context is invalidated while it runs, so the
		// retry acquires again and the primary now reports NL.
		f.runner.steps[0].during = cache.Clear
		f.plan.Network.Context.AllowedCountryCodes = []string{"DE"}
		_, err := f.engine.Run(f.run, f.plan)
		if err != nil || len(f.records(t)) != 1 || f.runner.Calls() != 1 || len(e.primary.Requests()) != 2 {
			t.Fatalf("err %v, %d results, runner calls %d", err, len(f.records(t)), f.runner.Calls())
		}
	})
}

// CC-CONTEXT-OVERVIEW: Network Context is a shared facility. Several runs on
// one access-network generation and path share one bounded acquisition
// instead of each repeating it.
// requirement-ref: REQ-CONTEXT-0001
func TestCcContextOverview(t *testing.T) {
	e := newEndpoints(t, ok(nlBody), ok(nlBody))
	f := newFixture(t, schedule(1), connected(t), connected(t), connected(t))
	f.withEgress(t, e, true)
	for range 3 {
		if _, err := f.engine.Run(f.run, f.plan); err != nil {
			t.Fatal(err)
		}
	}
	recs := f.records(t)
	if len(recs) != 3 || len(e.primary.Requests()) != 1 {
		t.Fatalf("%d results, %d requests", len(recs), len(e.primary.Requests()))
	}
	for i, src := range []string{"primary", "cache", "cache"} {
		if eg := egressOf(t, recs[i]); eg["lookup_source"] != src {
			t.Fatalf("run %d egress %v", i, eg)
		}
	}
}

// onProxyEgress rebinds the fixture to the proxy path of p with a lookup
// over e's names, which only p resolves. The lookup uses clock, so fixtures
// sharing a cache share its time source.
func (f *fixture) onProxyEgress(t *testing.T, e *endpointPair, p int, generation uint64, clock networkcontext.Clock, cache *networkcontext.Cache) {
	t.Helper()
	src, err := networkcontext.NewProxy(f.ctx.net, generation, connectionCode)
	if err != nil {
		t.Fatal(err)
	}
	src = bindLookup(t, src, e, true, clock, cache)
	f.ctx.direct = src
	p0, err := observation.NewPath(1, generation, "proxy", connectionCode)
	if err != nil {
		t.Fatal(err)
	}
	f.plan.Result, err = result.New(result.Binding{TaskID: "campaign-1", Output: result.Output{Test: model.Test{ID: tcp01.ID, Version: tcp01.Version, ResultVersion: result.Contract}},
		MaxAttempts: 2, MaxDurationMS: 60000, Implementation: implementation, ContextDecoder: src.Decoder(), Path: p0, AccessGeneration: 1})
	if err != nil {
		t.Fatal(err)
	}
	f.plan.Dialers = proxyDialers(t, p)
}

func proxyFor(t *testing.T, e *endpointPair, cfg loopback.ProxyConfig) *loopback.Proxy {
	t.Helper()
	cfg.Resolve = map[string]string{primaryName + ":443": e.primary.Address(), fallbackName + ":443": e.fallback.Address()}
	return scriptedProxy(t, cfg)
}

// CC-PROXY-NETWORK-CONTEXT and CC-CONTEXT-PATH-CONSISTENCY: a proxy attempt
// obtains egress only through HTTP CONNECT on its own proxy generation, its
// fallback through the same proxy, never through the direct path; proxy and
// direct contexts are separate cache entries and each proxy generation is
// its own key; a proxy that cannot carry the lookup leaves egress
// unavailable, which blocks only a task that requires it. NET-05 steps
// sharing one snapshot need the NET-05 runner (DEV-454).
// requirement-ref: REQ-PROXY-0023
// requirement-ref: REQ-PROXY-0024
// requirement-ref: REQ-PROXY-0025
// requirement-ref: REQ-CONTEXT-0017
// requirement-ref: REQ-CONTEXT-0018
// requirement-ref: REQ-CONTEXT-0019
func TestCcProxyNetworkContext(t *testing.T) {
	e := newEndpoints(t, contextendpoint.Response{Status: 503}, ok(nlBody))
	p := proxyFor(t, e, loopback.ProxyConfig{})
	cache, clock := networkcontext.NewCache(), testkit.NewClock(now)

	// A direct attempt first fills the direct entry through the routes.
	f := newFixture(t, schedule(1), connected(t))
	f.ctx.direct = bindLookup(t, f.ctx.direct, e, true, clock, cache)
	f.plan.Result, f.plan.Dialers = resultDecoder(t, f.ctx.direct, 0), e.routes
	if eg := egressOf(t, runOnce(t, f)[0]); eg["lookup_source"] != "fallback" || eg["network_path"] != "direct" {
		t.Fatalf("direct egress %v", eg)
	}
	directDials := len(e.routes.Dialed())

	// The proxy attempt does not reuse the direct entry: primary and fallback
	// go through CONNECT on the proxy, and the direct routes see nothing.
	f = newFixture(t, schedule(1), connected(t))
	f.onProxyEgress(t, e, p.Port(), proxyGeneration, clock, cache)
	recs := runOnce(t, f)
	eg := egressOf(t, recs[0])
	if recs[0].Execution.State != "completed" || eg["status"] != "fresh" || eg["lookup_source"] != "fallback" || eg["network_path"] != "proxy" || eg["connection_code"] != connectionCode {
		t.Fatalf("proxy egress %v", eg)
	}
	if got := p.Counts().Requests; len(got) != 2 || !strings.HasPrefix(got[0], "CONNECT "+primaryName+":443 HTTP/1.1\r\n") || !strings.HasPrefix(got[1], "CONNECT "+fallbackName+":443 HTTP/1.1\r\n") {
		t.Fatalf("CONNECT requests %q", got)
	}
	if len(e.routes.Dialed()) != directDials {
		t.Fatal("a proxy attempt used the direct path")
	}

	// The same proxy generation reuses its entry; another generation does not.
	f = newFixture(t, schedule(1), connected(t))
	f.onProxyEgress(t, e, p.Port(), proxyGeneration, clock, cache)
	if eg := egressOf(t, runOnce(t, f)[0]); eg["lookup_source"] != "cache" {
		t.Fatalf("same generation egress %v", eg)
	}
	f = newFixture(t, schedule(1), connected(t))
	f.onProxyEgress(t, e, p.Port(), proxyGeneration+1, clock, cache)
	if eg := egressOf(t, runOnce(t, f)[0]); eg["status"] != "fresh" {
		t.Fatalf("next generation egress %v", eg)
	}
	if len(p.Counts().Requests) != 4 || len(e.routes.Dialed()) != directDials {
		t.Fatalf("CONNECT requests %q", p.Counts().Requests)
	}

	// A proxy that refuses CONNECT: egress unavailable, never direct; a task
	// that requires the context does not run.
	refused := proxyFor(t, e, loopback.ProxyConfig{Mode: loopback.ProxyReply, Head: []byte("HTTP/1.1 403 Forbidden\r\n\r\n")})
	f = newFixture(t, schedule(1), connected(t))
	f.onProxyEgress(t, e, refused.Port(), proxyGeneration+2, clock, cache)
	f.plan.Network.Context.RequiredFields = []string{"country_code"}
	recs = runOnce(t, f)
	if recs[0].Execution.State != "not_run" || recs[0].Execution.Reason != "network_context_unavailable" || f.runner.Calls() != 0 || len(e.routes.Dialed()) != directDials {
		t.Fatalf("records %+v", recs)
	}
}

// CC-CONTEXT-CONTEXT-LAYERS-EGRESS-CONTEXT and CC-CONTEXT-FORMAT-USE-IN-RESULTS:
// the egress object carries the classification and its source on the
// attempt's path, with connection_code only on the proxy path; a response
// returning the observed address is rejected; the source address the
// endpoint observed, the endpoint names, the proxy endpoint and the
// generations never reach a result. The core writes no log, and cache keys
// hold only generations and the path name.
// requirement-ref: REQ-CONTEXT-0004
// requirement-ref: REQ-CONTEXT-0005
// requirement-ref: REQ-CONTEXT-0006
// requirement-ref: REQ-CONTEXT-FORMAT-0015
// requirement-ref: REQ-CONTEXT-FORMAT-0016
func TestCcContextContextLayersEgressContext(t *testing.T) {
	e := newEndpoints(t, ok(nlBody), ok(`{"ip":"192.0.2.10","classification_version":"0.1.0"}`))
	p := proxyFor(t, e, loopback.ProxyConfig{})

	f := newFixture(t, schedule(1), connected(t))
	f.withEgress(t, e, true)
	runOnce(t, f)
	g := newFixture(t, schedule(1), connected(t))
	g.onProxyEgress(t, e, p.Port(), proxyGeneration, g.clock, networkcontext.NewCache())
	runOnce(t, g)
	for i, raw := range [][]byte{f.store.Records()[0], g.store.Records()[0]} {
		var r record
		if err := json.Unmarshal(raw, &r); err != nil {
			t.Fatal(err)
		}
		eg := egressOf(t, r)
		_, code := eg["connection_code"]
		if eg["status"] != "fresh" || eg["country_code"] != "NL" || eg["region_code"] != "NL-NH" || eg["asn"] != float64(64500) || eg["classification_version"] != "0.1.0" || code != (i == 1) {
			t.Fatalf("egress %v", eg)
		}
		leaks := []string{primaryName, fallbackName, "network-context", "127.0.0.1", strconv.Itoa(p.Port()), strconv.Itoa(proxyGeneration), "generation"}
		for _, req := range append(e.primary.Requests(), e.fallback.Requests()...) {
			_, port, _ := net.SplitHostPort(req.Remote)
			leaks = append(leaks, req.Remote, port)
		}
		for _, leak := range leaks {
			if strings.Contains(string(raw), leak) {
				t.Fatalf("result contains %q: %s", leak, raw)
			}
		}
	}

	// A response carrying the observed address is rejected.
	e.primary.Set(ok(`{"ip":"192.0.2.10","country_code":"NL","classification_version":"0.1.0"}`))
	h := newFixture(t, schedule(1), connected(t))
	h.withEgress(t, e, true)
	recs := runOnce(t, h)
	if eg := egressOf(t, recs[0]); eg["status"] != "unavailable" || strings.Contains(string(h.store.Records()[0]), "192.0.2.10") {
		t.Fatalf("egress %v", eg)
	}
}

// CC-EXECUTION-FIELDS-NETWORK-FIELDS-NETWORK-CONTEXT-REQUIREMENTS and
// CC-CONTEXT-FORMAT-USE-IN-EXECUTION: requirements use the normalized field
// names and a country allowlist only; region, ASN, provider and operator
// allowlists and task-supplied endpoints are rejected by the task schema;
// requirements apply to the attempt's own path, and the task cannot change
// which endpoints are asked: every request goes to the configured names.
// Stricter local-policy context eligibility (REQ-EXECUTION-0023, last rule)
// is not provided by this build's local policy model.
// requirement-ref: REQ-EXECUTION-0021
// requirement-ref: REQ-EXECUTION-0022
// requirement-ref: REQ-EXECUTION-0023
// requirement-ref: REQ-EXECUTION-0024
// requirement-ref: REQ-CONTEXT-FORMAT-0013
// requirement-ref: REQ-CONTEXT-FORMAT-0014
func TestCcExecutionFieldsNetworkFieldsNetworkContextRequirements(t *testing.T) {
	n, err := networkTask(t, `"network":{"context":{"required_fields":["country_code","region_code","asn"],"allowed_country_codes":["NL","DE"]}}`)
	if err != nil || !slices.Equal(n.Context.RequiredFields, []string{"country_code", "region_code", "asn"}) || !slices.Equal(n.Context.AllowedCountryCodes, []string{"NL", "DE"}) {
		t.Fatalf("decoded %+v, %v", n.Context, err)
	}
	for _, bad := range []string{
		`"network":{"context":{"allowed_region_codes":["NL-NH"]}}`,
		`"network":{"context":{"allowed_asns":[64500]}}`,
		`"network":{"context":{"allowed_providers":["example"]}}`,
		`"network":{"context":{"allowed_operators":["example"]}}`,
		`"network":{"context":{"primary_endpoint":"https://context.example/"}}`,
		`"network":{"context":{"required_fields":["city"]}}`,
		`"network":{"context":{"allowed_country_codes":["nl"]}}`,
	} {
		if _, err := networkTask(t, bad); err == nil {
			t.Fatalf("accepted %s", bad)
		}
	}

	asn := uint32(64500)
	full := observation.Egress{Status: "fresh", CountryCode: "NL", RegionCode: "NL-NH", ASN: &asn}
	for _, tc := range []struct {
		req  task.ContextRequirements
		e    observation.Egress
		want task.ExecutionReason
	}{
		{task.ContextRequirements{}, observation.Egress{Status: "unavailable"}, ""},
		{task.ContextRequirements{RequiredFields: []string{"country_code", "region_code", "asn"}}, full, ""},
		{task.ContextRequirements{RequiredFields: []string{"region_code"}}, observation.Egress{Status: "fresh", CountryCode: "NL"}, task.NetworkContextUnavailable},
		{task.ContextRequirements{RequiredFields: []string{"asn"}}, observation.Egress{Status: "cached", CountryCode: "NL"}, task.NetworkContextUnavailable},
		{task.ContextRequirements{RequiredFields: []string{"provider"}}, full, task.NetworkContextUnavailable},
		{task.ContextRequirements{AllowedCountryCodes: []string{"NL"}}, full, ""},
		{task.ContextRequirements{AllowedCountryCodes: []string{"DE"}}, full, task.ConstraintsUnsatisfied},
		{task.ContextRequirements{AllowedCountryCodes: []string{"NL"}}, observation.Egress{Status: "unavailable"}, task.NetworkContextUnavailable},
	} {
		if got := execution.ContextConditions(tc.req, tc.e); got != tc.want {
			t.Fatalf("%+v against %+v: %q, want %q", tc.req, tc.e, got, tc.want)
		}
	}

	// The task's requirements do not change the requests: only the
	// configured names are dialed, once each at most.
	e := newEndpoints(t, contextendpoint.Response{Status: 503}, ok(nlBody))
	f := newFixture(t, schedule(1), connected(t))
	f.withEgress(t, e, true)
	f.plan.Network.Context = task.ContextRequirements{RequiredFields: []string{"asn"}, AllowedCountryCodes: []string{"NL"}}
	runOnce(t, f)
	if got := e.routes.Dialed(); !slices.Equal(got, []string{primaryName + ":443", fallbackName + ":443"}) {
		t.Fatalf("dialed %q", got)
	}
}
