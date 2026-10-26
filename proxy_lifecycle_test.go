package barometer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/config"
	"github.com/tv-rain/barometer-go/internal/execution"
	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/networkcontext"
	"github.com/tv-rain/barometer-go/internal/registry"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

const replacementProxy = `{"host":"127.0.0.1","port":54322,"connection_code":"outline_mobileproxy"}`

// proxyCoreHarness uses the production path planning and a runner that never
// dials. It exercises lifecycle ordering without requiring native adapters.
func proxyCoreHarness(t *testing.T) *harness {
	t.Helper()
	h := runningHarness(t)
	reg, err := registry.New([]registry.Entry{{Test: tcp01, Codec: tcpCodec{}, BaselineRisk: task.LowRisk, Targets: tcpTargets{}, Proxy: proxyCompatible{}, Runner: h.run.runner}})
	if err != nil {
		t.Fatal(err)
	}
	h.adapters.registry = reg
	h.adapters.approvals = config.Approvals{ConnectionCodes: []string{"outline_mobileproxy"}}
	return h
}

func configuredCore(t *testing.T) (*harness, *Client) {
	t.Helper()
	h := proxyCoreHarness(t)
	c := h.client(t)
	if err := c.Initialize(proxyConfig(54321)); err != nil {
		t.Fatal(err)
	}
	return h, c
}

// CC-PROXY-CONFIGURATION-VALIDATION: complete strict validation is atomic;
// invalid replacements neither change nor cancel the active generation.
// Coarse categories require trusted approval; compatibility comes from the
// registered rule, with no host-supplied test capability list.
// requirement-ref: REQ-PROXY-0005
// requirement-ref: REQ-PROXY-0006
// requirement-ref: REQ-PROXY-0007
// requirement-ref: REQ-PROXY-0008
// requirement-ref: REQ-PROXY-0009
// requirement-ref: REQ-PROXY-0010
func TestCcProxyConfigurationValidation(t *testing.T) {
	h, c := configuredCore(t)
	h.run.runner.holdCalls()
	rh, err := c.StartRun(h.admit(t, c, runTask("invalid-proxy", "", `"network_path":"proxy"`, false)))
	if err != nil {
		t.Fatal(err)
	}
	waitStarted(t, h.run.runner)
	before := c.rt.proxy()
	for _, input := range []string{
		``, `null`, `{}`, `[]`, replacementProxy + `{}`,
		strings.Repeat(" ", config.MaxBytes+1),
		strings.Replace(replacementProxy, `127.0.0.1`, `198.51.100.1`, 1),
		strings.Replace(replacementProxy, `54322`, `0`, 1),
		strings.Replace(replacementProxy, `54322`, `65536`, 1),
		strings.Replace(replacementProxy, `54322`, `1.0`, 1),
		strings.Replace(replacementProxy, `outline_mobileproxy`, `Not_coarse`, 1),
		strings.Replace(replacementProxy, `outline_mobileproxy`, `unapproved_session`, 1),
		strings.Replace(replacementProxy, `outline_mobileproxy`, strings.Repeat("a", 65), 1),
		strings.Replace(replacementProxy, `outline_mobileproxy`, ``, 1),
		strings.Replace(replacementProxy, `}`, `,"password":"canary-secret"}`, 1),
		strings.Replace(replacementProxy, `}`, `,"tests":["TCP-01"]}`, 1),
		strings.Replace(replacementProxy, `}`, `,"host":"localhost"}`, 1),
		strings.Replace(replacementProxy, `}`, `,"primary_endpoint":"https://egress.example.test"}`, 1),
	} {
		if codeOf(c.SetProxy(input)) != "proxy_rejected" || c.rt.proxy() != before {
			t.Fatal("invalid replacement changed configuration")
		}
		if state, err := c.RunState(rh); err != nil || state != runActive {
			t.Fatal("invalid replacement cancelled the active run")
		}
	}
	if err := c.SetProxy(replacementProxy); err != nil {
		t.Fatal(err)
	}
	got := records(t, h.run.output)
	if len(got) != 1 || got[0].Execution.Reason != "proxy_reconfigured" {
		t.Fatalf("cancel result %+v", got)
	}
	if after := c.rt.proxy(); after.generation <= before.generation || after.endpoint.Port != 54322 || after.endpoint.ConnectionCode != before.endpoint.ConnectionCode {
		t.Fatal("replacement was not atomic or changed category")
	}
	h.run.runner.releaseCalls()
	// Registry-derived compatibility still applies after renewal.
	if _, err := c.rt.tasks.lookup(h.admit(t, c, runTask("after-invalid", "", `"network_path":"proxy"`, false))); err != nil {
		t.Fatal(err)
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}

// requirement-ref: REQ-PROXY-0011
// requirement-ref: REQ-PROXY-0012
// requirement-ref: REQ-PROXY-0013
// requirement-ref: REQ-PROXY-0015
// requirement-ref: REQ-PROXY-0030
func TestProxyChangesOnlyCancelProxyRuns(t *testing.T) {
	for _, clear := range []bool{false, true} {
		t.Run(fmt.Sprint("clear=", clear), func(t *testing.T) {
			h, c := configuredCore(t)
			h.run.runner.holdCalls()
			direct, err := c.StartRun(h.admit(t, c, runTask("kept-direct", "", "", false)))
			if err != nil {
				t.Fatal(err)
			}
			waitStarted(t, h.run.runner)
			retry := `"retry":{"max_attempts":2,"retry_on":["internal_error"],"backoff":{"mode":"fixed","initial_delay_ms":30000,"max_delay_ms":30000}}`
			proxy, err := c.StartRun(h.admit(t, c, runTask("old-proxy", retry, `"network_path":"proxy"`, false)))
			if err != nil {
				t.Fatal(err)
			}
			waitStarted(t, h.run.runner)
			if clear {
				err = c.ClearProxy()
			} else {
				err = c.SetProxy(replacementProxy)
			}
			if err != nil {
				t.Fatal(err)
			}
			if state, _ := c.RunState(direct); state != runActive {
				t.Fatal("direct run was cancelled")
			}
			if state, _ := c.RunState(proxy); state != runTerminal {
				t.Fatal("old proxy generation still in use")
			}
			got := records(t, h.run.output)
			if len(got) != 1 || got[0].Execution.State != "incomplete" || got[0].Execution.Reason != "proxy_reconfigured" {
				t.Fatalf("results %+v", got)
			}
			h.run.runner.releaseCalls()
			waitTerminal(t, c, direct)
			if err := c.ClearProxy(); err != nil {
				t.Fatal(err)
			}
			if err := c.ClearProxy(); err != nil {
				t.Fatal(err)
			}
			if c.rt.proxy().configured() {
				t.Fatal("clear left a proxy")
			}
			missing, err := c.StartRun(h.admit(t, c, runTask("cleared-proxy", "", `"network_path":"proxy"`, false)))
			if err != nil {
				t.Fatal(err)
			}
			waitTerminal(t, c, missing)
			clearedResults := records(t, h.run.output)
			if len(clearedResults) != 3 || clearedResults[2].Execution.Reason != "proxy_not_configured" || h.run.runner.Calls() != 2 {
				t.Fatal("cleared proxy was used or fell back to direct")
			}
			if err := c.SetProxy(replacementProxy); err != nil {
				t.Fatal(err)
			}
			next, err := c.StartRun(h.admit(t, c, runTask("new-proxy", "", `"network_path":"proxy"`, false)))
			if err != nil {
				t.Fatal(err)
			}
			waitTerminal(t, c, next)
			if h.run.runner.Calls() != 3 {
				t.Fatal("old run retried or a direct run disappeared")
			}
			for _, raw := range h.run.output.Records() {
				for _, forbidden := range []string{"generation", "54321", "54322", "127.0.0.1", "CONNECT"} {
					if strings.Contains(string(raw), forbidden) {
						t.Fatal("result leaked proxy internals")
					}
				}
			}
			if err := c.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// Proxy operations fail closed outside the initialized state.
func TestProxyLifecycleStates(t *testing.T) {
	var nilClient *Client
	for _, c := range []*Client{nilClient, NewClient(), &Client{}} {
		if codeOf(c.SetProxy(replacementProxy)) != "not_initialized" || codeOf(c.ClearProxy()) != "not_initialized" {
			t.Fatal("uninitialized proxy operation")
		}
	}
	c := NewClient()
	_ = c.Close()
	if codeOf(c.SetProxy(replacementProxy)) != "client_closed" || codeOf(c.ClearProxy()) != "client_closed" {
		t.Fatal("closed proxy operation")
	}
}

// A manual reservation deterministically holds cleanup. Every losing lifecycle
// call fails promptly rather than waiting on a lock held across cleanup/I/O.
func TestProxyLifecycleSerialization(t *testing.T) {
	for _, clear := range []bool{false, true} {
		_, c := configuredCore(t)
		_, g, err := c.rt.proxies.acquire()
		if err != nil {
			t.Fatal(err)
		}
		result := make(chan error, 1)
		go func() {
			if clear {
				result <- c.ClearProxy()
			} else {
				result <- c.SetProxy(replacementProxy)
			}
		}()
		select {
		case <-g.ctx.Done():
		case <-time.After(time.Second):
			t.Fatal("change did not cancel the generation")
		}
		for _, call := range []func() error{c.Close, c.ClearMeasurementData, c.ClearProxy, func() error { return c.Initialize(`{}`) }, func() error { return c.SetProxy(replacementProxy) }, func() error { return c.replacePolicy(task.LocalPolicy{}) }} {
			if codeOf(call()) != "transition_in_progress" {
				t.Fatal("lifecycle operation crossed proxy change")
			}
		}
		if _, _, err := c.rt.proxies.acquire(); codeOf(err) != "transition_in_progress" {
			t.Fatal("new reservation entered old generation")
		}
		select {
		case err := <-result:
			t.Fatal("returned with an old reservation", err)
		default:
		}
		c.rt.proxies.release(g)
		if err := <-result; err != nil {
			t.Fatal(err)
		}
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

// One injected cleanup budget covers the generation. Failure installs nothing
// and leaves a fail-closed client for recovery clearing or close.
// requirement-ref: REQ-LIMIT-0008
func TestProxyCleanupIncomplete(t *testing.T) {
	for _, clear := range []bool{false, true} {
		h := proxyCoreHarness(t)
		h.adapters.deadline = defaultLifecycleDeadline
		c := h.client(t)
		if err := c.Initialize(proxyConfig(54321)); err != nil {
			t.Fatal(err)
		}
		stuck := make(chan struct{})
		h.run.runner.stuck = stuck
		rh, err := c.StartRun(h.admit(t, c, runTask("stuck-proxy", "", `"network_path":"proxy"`, false)))
		if err != nil {
			t.Fatal(err)
		}
		waitStarted(t, h.run.runner)
		h.waitTimers(t, 1)
		before := c.rt.proxy()
		result := make(chan error, 1)
		go func() {
			if clear {
				result <- c.ClearProxy()
			} else {
				result <- c.SetProxy(replacementProxy)
			}
		}()
		h.waitTimers(t, 2)
		_ = h.run.clock.Advance(execution.CleanupBound - time.Millisecond)
		select {
		case <-result:
			t.Fatal("cleanup returned before bound")
		default:
		}
		_ = h.run.clock.Advance(time.Millisecond)
		if codeOf(<-result) != "cleanup_incomplete" || stateOf(c) != blocked || c.rt.proxy() != before {
			t.Fatal("incomplete cleanup activated a replacement")
		}
		if codeOf(c.ClearProxy()) != "cleanup_incomplete" {
			t.Fatal("blocked proxy scope accepted work")
		}
		r, _ := c.rt.runs.lookup(rh)
		close(stuck)
		select {
		case <-r.status.finished:
		case <-time.After(time.Second):
			t.Fatal("runner failed to finish")
		}
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
		got := records(t, h.run.output)
		if len(got) != 1 || got[0].Execution.Reason != "proxy_reconfigured" {
			t.Fatalf("results %+v", got)
		}
	}
}

// Simultaneous starts and set/clear operations can choose one generation or
// refuse the transition. Every accepted run keeps its snapshot for its lifetime.
func TestProxyStartRace(t *testing.T) {
	for n := 0; n < 40; n++ {
		h, c := configuredCore(t)
		th := h.admit(t, c, runTask("race-proxy", "", `"network_path":"proxy"`, false))
		start := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		var rh string
		var runErr, changeErr error
		go func() { defer wg.Done(); <-start; rh, runErr = c.StartRun(th) }()
		go func() {
			defer wg.Done()
			<-start
			if n%2 == 0 {
				changeErr = c.ClearProxy()
			} else {
				changeErr = c.SetProxy(replacementProxy)
			}
		}()
		close(start)
		wg.Wait()
		if changeErr != nil {
			t.Fatal(changeErr)
		}
		if runErr == nil {
			waitTerminal(t, c, rh)
		} else if codeOf(runErr) != "transition_in_progress" && codeOf(runErr) != "run_not_due" {
			t.Fatal(runErr)
		}
		for _, raw := range h.run.output.Records() {
			var res proxiedResult
			if json.Unmarshal(raw, &res) != nil {
				t.Fatal("bad result")
			}
			if res.Execution.State == "completed" && res.NetworkContext.Egress["network_path"] != "proxy" {
				t.Fatal("fallback during race")
			}
		}
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestProxyPolicyReplacementKeepsCurrentEndpoint(t *testing.T) {
	for _, clear := range []bool{false, true} {
		_, c := configuredCore(t)
		var err error
		if clear {
			err = c.ClearProxy()
		} else {
			err = c.SetProxy(replacementProxy)
		}
		if err != nil {
			t.Fatal(err)
		}
		before := c.rt.proxy()
		in := synPolicyInput(3, "reg-a", func(p *task.Permissions) { p.Tests = []task.TestVersion{tcp01} })
		in.Targets[0].Test = tcp01
		p, err := task.NewLocalPolicy(in)
		if err != nil {
			t.Fatal(err)
		}
		if err := c.replacePolicy(p); err != nil {
			t.Fatal(err)
		}
		after := c.rt.proxy()
		if after.configured() != before.configured() || after.endpoint != before.endpoint || after.configured() && after.generation <= before.generation {
			t.Fatal("policy replacement resurrected initialization proxy")
		}
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestProxyGenerationSurvivesReinitialization(t *testing.T) {
	_, c := configuredCore(t)
	if err := c.SetProxy(replacementProxy); err != nil {
		t.Fatal(err)
	}
	before := c.rt.proxy().generation
	if err := c.ClearProxy(); err != nil {
		t.Fatal(err)
	}
	if err := c.ClearMeasurementData(); err != nil {
		t.Fatal(err)
	}
	if err := c.Initialize(proxyConfig(54321)); err != nil {
		t.Fatal(err)
	}
	if c.rt.proxy().generation <= before {
		t.Fatal("proxy generation reused within the client instance")
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}

// The serializable Go facade keeps egress endpoints outside the proxy object.
// Native and Flutter method mapping remains a separate binding review.
// requirement-ref: REQ-PROXY-0002
// requirement-ref: REQ-PROXY-0003
// requirement-ref: REQ-PROXY-0004
func TestProxyHostConfigurationStaysSeparate(t *testing.T) {
	h := proxyCoreHarness(t)
	h.adapters.approvals.Endpoints = []string{"https://primary.example.test/context", "https://fallback.example.test/context"}
	c := h.client(t)
	input := strings.TrimSuffix(proxyConfig(54321), `}`) + `,"network_context":{"primary_endpoint":"https://primary.example.test/context","fallback_endpoint":"https://fallback.example.test/context"}}`
	if err := c.Initialize(input); err != nil {
		t.Fatal(err)
	}
	before, ok := c.rt.init.Configuration().NetworkContext()
	if !ok {
		t.Fatal("egress endpoints missing")
	}
	if err := c.SetProxy(replacementProxy); err != nil {
		t.Fatal(err)
	}
	if err := c.ClearProxy(); err != nil {
		t.Fatal(err)
	}
	if after, ok := c.rt.init.Configuration().NetworkContext(); !ok || after != before {
		t.Fatal("proxy change altered egress configuration")
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}

type beforeProxyAttempt struct{ entered chan struct{} }

func (p beforeProxyAttempt) Capture(ctx context.Context) (networkcontext.Snapshot, error) {
	close(p.entered)
	<-ctx.Done()
	return networkcontext.Snapshot{}, ctx.Err()
}

type proxyCapturePaths struct {
	inner   pathSource
	entered chan struct{}
}

func (p proxyCapturePaths) plan(ctx context.Context, r pathRequest) (pathPlan, error) {
	plan, err := p.inner.plan(ctx, r)
	plan.Context = beforeProxyAttempt{p.entered}
	return plan, err
}

// Reconfiguration during context acquisition is not_run, without Network
// Context or a protocol outcome, rather than an incomplete started test.
// requirement-ref: REQ-PROXY-0015
// requirement-ref: REQ-RESULT-0051
// requirement-ref: REQ-RESULT-0053
func TestProxyReconfiguredBeforeTestStarts(t *testing.T) {
	h := proxyCoreHarness(t)
	entered := make(chan struct{})
	h.adapters.run.paths = proxyCapturePaths{inner: h.run.paths, entered: entered}
	c := h.client(t)
	if err := c.Initialize(proxyConfig(54321)); err != nil {
		t.Fatal(err)
	}
	rh, err := c.StartRun(h.admit(t, c, runTask("before-proxy-start", "", `"network_path":"proxy"`, false)))
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("attempt did not capture context")
	}
	if err := c.ClearProxy(); err != nil {
		t.Fatal(err)
	}
	waitTerminal(t, c, rh)
	raw := h.run.output.Records()
	if len(raw) != 1 {
		t.Fatal("missing result")
	}
	var res proxiedResult
	if json.Unmarshal(raw[0], &res) != nil || res.Execution.State != "not_run" || res.Execution.Reason != "proxy_reconfigured" || res.NetworkContext != nil || h.run.runner.Calls() != 0 {
		t.Fatalf("invalid pre-start result %s", raw[0])
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}

// Task-controlled proxy fields cannot reach the host's active configuration.
// The target-validation rows of the broader privacy family need production
// codecs and proxy rules; this test does not claim that family is executed.
// requirement-ref: REQ-PROXY-0034
// requirement-ref: REQ-PROXY-0038
func TestProxyTaskCannotReplaceHostConfiguration(t *testing.T) {
	h, c := configuredCore(t)
	before := c.rt.proxy()
	base := runTask("task-proxy-fields", "", `"network_path":"proxy"`, false)
	for _, key := range []string{"proxy", "host", "port", "endpoint", "connection_code"} {
		for _, input := range []string{
			strings.TrimSuffix(base, `}`) + fmt.Sprintf(`,%q:"canary"}`, key),
			strings.Replace(base, `"input":{}`, fmt.Sprintf(`"input":{%q:"canary"}`, key), 1),
			strings.Replace(base, `"options":{}`, fmt.Sprintf(`"options":{%q:"canary"}`, key), 1),
			strings.Replace(base, `"limits":{}`, fmt.Sprintf(`"limits":{},%q:"canary"`, key), 1),
		} {
			_, err := c.AdmitTask(input, h.embedded(t, input))
			h.deliveries.Invalidate() // discard each rejected fixture's delivery slot
			if codeOf(err) != "task_schema_invalid" || c.rt.proxy() != before {
				t.Fatal("task changed host proxy")
			}
		}
	}
	if h.run.runner.Calls() != 0 || h.run.paths.Calls() != 0 {
		t.Fatal("rejected input reached the path")
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}

// blockedProxyPaths stands in for platform I/O that observes cancellation
// but needs an explicit release to finish cleanup.
type blockedProxyPaths struct {
	entered   chan struct{}
	cancelled chan struct{}
	release   chan struct{}
}

func (p *blockedProxyPaths) plan(ctx context.Context, _ pathRequest) (pathPlan, error) {
	close(p.entered)
	<-ctx.Done()
	close(p.cancelled)
	<-p.release
	return pathPlan{}, ctx.Err()
}

func TestProxyChangeDrainsPathReservation(t *testing.T) {
	h := proxyCoreHarness(t)
	p := &blockedProxyPaths{entered: make(chan struct{}), cancelled: make(chan struct{}), release: make(chan struct{})}
	h.adapters.run.paths = p
	c := h.client(t)
	if err := c.Initialize(proxyConfig(54321)); err != nil {
		t.Fatal(err)
	}
	th := h.admit(t, c, runTask("blocked-path", "", `"network_path":"proxy"`, false))
	startResult := make(chan error, 1)
	go func() { _, err := c.StartRun(th); startResult <- err }()
	select {
	case <-p.entered:
	case <-time.After(time.Second):
		t.Fatal("path reservation never entered")
	}
	changeResult := make(chan error, 1)
	go func() { changeResult <- c.SetProxy(replacementProxy) }()
	select {
	case <-p.cancelled:
	case <-time.After(time.Second):
		t.Fatal("platform I/O not cancelled")
	}
	// The state lock is available while platform I/O is still in progress.
	if codeOf(c.Close()) != "transition_in_progress" {
		t.Fatal("lifecycle crossed proxy replacement")
	}
	select {
	case <-changeResult:
		t.Fatal("returned before platform cleanup")
	default:
	}
	close(p.release)
	if codeOf(<-startResult) != "run_not_due" {
		t.Fatal("cancelled reservation launched work")
	}
	if err := <-changeResult; err != nil {
		t.Fatal(err)
	}
	if h.run.runner.Calls() != 0 || len(h.run.output.Records()) != 0 {
		t.Fatal("cancelled reservation produced an attempt")
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}

// Concurrent lifecycle operations may win or refuse the proxy transition;
// none may partially activate a proxy or resurrect an earlier configuration.
func TestProxyAgainstLifecycleRace(t *testing.T) {
	for _, operation := range []string{"initialize", "close", "clear", "policy"} {
		t.Run(operation, func(t *testing.T) {
			for n := 0; n < 15; n++ {
				_, c := configuredCore(t)
				in := synPolicyInput(3, "reg-a", func(p *task.Permissions) { p.Tests = []task.TestVersion{tcp01} })
				in.Targets[0].Test = tcp01
				p, err := task.NewLocalPolicy(in)
				if err != nil {
					t.Fatal(err)
				}
				start := make(chan struct{})
				var wg sync.WaitGroup
				wg.Add(2)
				var changeErr, lifecycleErr error
				go func() {
					defer wg.Done()
					<-start
					if n%2 == 0 {
						changeErr = c.SetProxy(replacementProxy)
					} else {
						changeErr = c.ClearProxy()
					}
				}()
				go func() {
					defer wg.Done()
					<-start
					switch operation {
					case "initialize":
						lifecycleErr = c.Initialize(`{}`)
					case "close":
						lifecycleErr = c.Close()
					case "clear":
						lifecycleErr = c.ClearMeasurementData()
					case "policy":
						lifecycleErr = c.replacePolicy(p)
					}
				}()
				close(start)
				wg.Wait()
				for _, e := range []error{changeErr, lifecycleErr} {
					switch codeOf(e) {
					case "ok", "already_initialized", "not_initialized", "client_closed", "transition_in_progress":
					default:
						t.Fatal(e)
					}
				}
				if stateOf(c) == initialized && changeErr == nil {
					current := c.rt.proxy()
					if n%2 == 0 && current.endpoint.Port != 54322 || n%2 == 1 && current.configured() {
						t.Fatal("partial proxy activation across lifecycle race")
					}
				}
				if err := c.Close(); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

// Real loopback evidence: cancel a CONNECT exchange on the old endpoint,
// wait for its socket to close, then start a separately admitted run on the
// new endpoint. Neither renewal nor clearing uses the target's direct path.
// requirement-ref: REQ-PROXY-0013
// requirement-ref: REQ-PROXY-0014
// requirement-ref: REQ-PROXY-0015
func TestProxyLoopbackRenewal(t *testing.T) {
	for _, clear := range []bool{false, true} {
		h, _, target := proxyHarness(t)
		old, err := loopback.ListenProxy(loopback.ProxyConfig{Mode: loopback.ProxyHold})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = old.Close() })
		next, err := loopback.ListenProxy(loopback.ProxyConfig{})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = next.Close() })
		c := h.client(t)
		if err := c.Initialize(proxyConfig(old.Port())); err != nil {
			t.Fatal(err)
		}
		rh, err := c.StartRun(h.admit(t, c, runTask("held-connect", "", `"network_path":"proxy"`, false)))
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := old.WaitRequests(ctx, 1); err != nil {
			t.Fatal(err)
		}
		if clear {
			err = c.ClearProxy()
		} else {
			err = c.SetProxy(strings.TrimSuffix(strings.TrimPrefix(proxyConfig(next.Port()), `{"proxy":`), `}`))
		}
		if err != nil {
			t.Fatal(err)
		}
		if err := old.WaitIdle(ctx); err != nil {
			t.Fatal("old CONNECT still open", err)
		}
		waitTerminal(t, c, rh)
		if target.Accepted() != 0 || next.Counts().Accepted != 0 {
			t.Fatal("old attempt continued via replacement or direct path")
		}
		got := records(t, h.run.output)
		if len(got) != 1 || got[0].Execution.Reason != "proxy_reconfigured" {
			t.Fatalf("results %+v", got)
		}
		if clear {
			if err := c.SetProxy(strings.TrimSuffix(strings.TrimPrefix(proxyConfig(next.Port()), `{"proxy":`), `}`)); err != nil {
				t.Fatal(err)
			}
		}
		rh, err = c.StartRun(h.admit(t, c, runTask("renewed-connect", "", `"network_path":"proxy"`, false)))
		if err != nil {
			t.Fatal(err)
		}
		waitTerminal(t, c, rh)
		if next.Counts().Relays != 1 || old.Counts().Accepted != 1 || target.Accepted() != 1 {
			t.Fatal("separately scheduled attempt did not use only the new endpoint")
		}
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

// heldTunnelRunner blocks on a read without sending application bytes. The
// target waits for its first request, so only cancellation closes the socket.
type heldTunnelRunner struct {
	address string
	entered chan struct{}
}

func (r heldTunnelRunner) Run(ctx context.Context, in measurement.Input) (measurement.Output, error) {
	conn, err := in.Dialer.DialContext(ctx, "tcp", r.address)
	if err != nil {
		return measurement.Output{}, err
	}
	defer conn.Close()
	close(r.entered)
	_, err = conn.Read(make([]byte, 1))
	return measurement.Output{}, err
}

// Both a pending CONNECT and an established tunnel must be closed. This
// complements TestProxyLoopbackRenewal's pending-exchange and renewal rows.
func TestProxyLoopbackClosesEstablishedTunnel(t *testing.T) {
	for _, clear := range []bool{false, true} {
		h, _, target := proxyHarness(t)
		entered := make(chan struct{})
		reg, err := registry.New([]registry.Entry{{Test: tcp01, Codec: tcpCodec{}, BaselineRisk: task.LowRisk, Targets: tcpTargets{}, Proxy: proxyCompatible{}, Runner: heldTunnelRunner{address: target.Address(), entered: entered}}})
		if err != nil {
			t.Fatal(err)
		}
		h.adapters.registry = reg
		old, err := loopback.ListenProxy(loopback.ProxyConfig{})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = old.Close() })
		c := h.client(t)
		if err := c.Initialize(proxyConfig(old.Port())); err != nil {
			t.Fatal(err)
		}
		rh, err := c.StartRun(h.admit(t, c, runTask("held-tunnel", "", `"network_path":"proxy"`, false)))
		if err != nil {
			t.Fatal(err)
		}
		select {
		case <-entered:
		case <-time.After(time.Second):
			t.Fatal("tunnel did not open")
		}
		if clear {
			err = c.ClearProxy()
		} else {
			err = c.SetProxy(replacementProxy)
		}
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := old.WaitIdle(ctx); err != nil {
			t.Fatal("tunnel remained open", err)
		}
		waitTerminal(t, c, rh)
		if err := target.WaitAccepted(ctx, 1); err != nil {
			t.Fatal("target did not accept tunnel", err)
		}
		got := records(t, h.run.output)
		if len(got) != 1 || got[0].Execution.Reason != "proxy_reconfigured" || target.Accepted() != 1 {
			t.Fatalf("tunnel result %+v", got)
		}
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
