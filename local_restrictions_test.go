package barometer

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/testkit/contextendpoint"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

// Local policy restrictions through the root facade: device constraints and
// Network Context eligibility stricter than a task's. Synthetic platform
// state, loopback endpoints and synthetic policies only; no field claim.

// deviceHarness is runningHarness whose local policy sets the given device
// constraints.
func deviceHarness(t *testing.T, minimum *int, charging bool) *harness {
	t.Helper()
	return customHarness(t, nil, func(in *task.PolicyInput) {
		in.Device = task.DeviceConstraints{MinimumBatteryPercent: minimum, ChargingOnly: charging}
	})
}

func percent(n int) *int { return &n }

// CC-EXECUTION-FIELDS-DEVICE-FIELDS: minimum_battery_percent and
// allow_charging_only gate every attempt on synthetic platform state, and
// local policy may enforce stricter constraints: the higher battery minimum
// and the charging requirement of either side apply, an unknown or
// unavailable state never satisfies a local constraint, and no local value
// relaxes the task's.
// requirement-ref: REQ-EXECUTION-0025
// requirement-ref: REQ-EXECUTION-0026
func TestCcExecutionFieldsDeviceFields(t *testing.T) {
	notAfter := `"not_after":"` + at(20*time.Hour) + `"`
	battery := func(n int) string { return `"device":{"minimum_battery_percent":` + strconv.Itoa(n) + `}` }
	charging := `"device":{"allow_charging_only":true}`
	level := func(n int) *platform.DeviceState { return &platform.DeviceState{BatteryPercent: n, BatteryKnown: true} }
	cases := []struct {
		name          string
		localMinimum  *int
		localCharging bool
		execution     string
		state         *platform.DeviceState // nil: the capability fails
		noProvider    bool
		runs          bool
	}{
		{"task battery at its minimum", nil, false, battery(20), level(20), false, true},
		{"task battery below its minimum", nil, false, battery(20), level(19), false, false},
		{"task charging while charging", nil, false, charging, &platform.DeviceState{Charging: true, ChargingKnown: true}, false, true},
		{"local minimum without a task value, below", percent(50), false, "", level(49), false, false},
		{"local minimum without a task value, at", percent(50), false, "", level(50), false, true},
		{"local minimum stricter than the task's", percent(50), false, battery(20), level(30), false, false},
		{"task minimum stricter than the local one", percent(20), false, battery(50), level(30), false, false},
		{"both minimums met", percent(20), false, battery(50), level(50), false, true},
		{"local zero minimum, unknown battery", percent(0), false, "", &platform.DeviceState{Charging: true, ChargingKnown: true}, false, false},
		{"local minimum, unknown battery", percent(10), false, "", &platform.DeviceState{BatteryPercent: 99}, false, false},
		{"local minimum, device unavailable", percent(10), false, "", nil, false, false},
		{"local minimum, no device capability", percent(10), false, "", nil, true, false},
		{"local charging, not charging", nil, true, "", &platform.DeviceState{ChargingKnown: true, BatteryPercent: 100, BatteryKnown: true}, false, false},
		{"local charging, unknown charging", nil, true, "", &platform.DeviceState{Charging: true}, false, false},
		{"local charging while charging", nil, true, "", &platform.DeviceState{Charging: true, ChargingKnown: true}, false, true},
		{"local charging adds to a task battery minimum", nil, true, battery(20), level(80), false, false},
		{"task charging with a local battery minimum", percent(30), false, charging, &platform.DeviceState{Charging: true, ChargingKnown: true, BatteryPercent: 29, BatteryKnown: true}, false, false},
		{"both met", percent(30), true, charging, &platform.DeviceState{Charging: true, ChargingKnown: true, BatteryPercent: 30, BatteryKnown: true}, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := deviceHarness(t, tc.localMinimum, tc.localCharging)
			switch {
			case tc.noProvider:
				h.adapters.run.device = nil
			case tc.state == nil:
				h.run.device.Fail()
			default:
				h.run.device.Set(*tc.state)
			}
			c := h.initialized(t)
			code := startCode(t, c, h.admit(t, c, scheduledTask("device-local", notAfter, tc.execution, false)))
			if tc.runs != (code == "ok") || !tc.runs && code != "run_not_due" {
				t.Fatalf("start = %s, want run=%t", code, tc.runs)
			}
			if !tc.runs && (h.run.runner.Calls() != 0 || len(h.run.output.Records()) != 0) {
				t.Fatal("an attempt or Measurement Result with an unsatisfied device constraint")
			}
		})
	}
	t.Run("no constraint queries no device", func(t *testing.T) {
		h := deviceHarness(t, nil, false)
		c := h.initialized(t)
		if code := startCode(t, c, h.admit(t, c, scheduledTask("device-none", notAfter, "", false))); code != "ok" || h.run.device.Calls() != 0 {
			t.Fatalf("start = %s with %d device queries", code, h.run.device.Calls())
		}
	})
	t.Run("local constraint applies before the retry", func(t *testing.T) {
		// Attempt 1 fails with a retryable reason, then the battery drops
		// below the local minimum; the task states no device constraint, and
		// no second attempt starts.
		h := deviceHarness(t, percent(50), false)
		h.run.device.Set(*level(80))
		h.run.runner.fail[1] = true
		h.run.runner.during = func(n int) {
			if n == 1 {
				h.run.device.Set(*level(49))
			}
		}
		c := h.initialized(t)
		runToEnd(t, c, h.admit(t, c, scheduledTask("device-retry", notAfter+","+fixedRetry("internal_error"), "", false)))
		if h.run.runner.Calls() != 1 || len(h.run.output.Records()) != 1 {
			t.Fatalf("%d attempts after the local constraint failed", h.run.runner.Calls())
		}
	})
}

// contextPolicy rebinds the directHarness policy with local Network Context
// eligibility; nil leaves acquisition to the host configuration.
func contextPolicy(t *testing.T, h *harness, r *echoRunner, eligibility *task.ContextEligibility) {
	t.Helper()
	in := synPolicyInput(2, "reg-a", func(p *task.Permissions) { p.Tests = []task.TestVersion{tcp01} })
	in.Targets[0].Test = tcp01
	in.Targets[0].Tuple = *loopbackTargets(r.address).tuple
	in.Context = eligibility
	p, err := task.NewLocalPolicy(in)
	if err != nil {
		t.Fatal("synthetic policy rejected")
	}
	h.adapters.policy = p
}

// The stricter-local-policy row of
// CC-EXECUTION-FIELDS-NETWORK-FIELDS-NETWORK-CONTEXT-REQUIREMENTS: local
// policy may forbid Network Context acquisition, or narrow it to some
// configured endpoints or to one path. A forbidden acquisition contacts no
// endpoint; a task without context requirements runs with egress unavailable
// and a task that requires a field records network_context_unavailable
// (REQ-EXECUTION-0024, REQ-CONTEXT-0028). Narrowing never adds an endpoint
// the host did not configure, and request counts show that no task value
// changes the endpoints or the path.
// requirement-ref: REQ-EXECUTION-0022
// requirement-ref: REQ-EXECUTION-0023
// requirement-ref: REQ-EXECUTION-0024
func TestLocalPolicyContextEligibility(t *testing.T) {
	requires := `"network":{"context":{"required_fields":["country_code"]}}`
	allowlist := `"network":{"context":{"allowed_country_codes":["NL"]}}`
	direct := task.Paths{Direct: true}
	cases := []struct {
		name        string
		eligibility *task.ContextEligibility
		// status of a task without requirements; "" means no acquisition and
		// network_context_unavailable for a task that requires a field.
		status             string
		primary, fallback  bool // endpoints contacted
		requirementOutcome string
	}{
		{"no local eligibility", nil, "fresh", true, false, "completed"},
		{"both endpoints and the direct path", &task.ContextEligibility{Paths: direct, Endpoints: []string{contextPrimary, contextFallback}}, "fresh", true, false, "completed"},
		{"acquisition forbidden", &task.ContextEligibility{}, "unavailable", false, false, "network_context_unavailable"},
		{"direct path, no endpoint", &task.ContextEligibility{Paths: direct}, "unavailable", false, false, "network_context_unavailable"},
		{"endpoints on the proxy path only", &task.ContextEligibility{Paths: task.Paths{Proxy: true}, Endpoints: []string{contextPrimary, contextFallback}}, "unavailable", false, false, "network_context_unavailable"},
		{"only an unconfigured endpoint", &task.ContextEligibility{Paths: direct, Endpoints: []string{"https://other.example/network-context"}}, "unavailable", false, false, "network_context_unavailable"},
		// The fallback answers 503, so egress stays unavailable, but only
		// the permitted endpoint is asked.
		{"fallback endpoint only", &task.ContextEligibility{Paths: direct, Endpoints: []string{contextFallback}}, "unavailable", false, true, "network_context_unavailable"},
		{"primary endpoint only", &task.ContextEligibility{Paths: direct, Endpoints: []string{contextPrimary}}, "fresh", true, false, "completed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, r, e := contextHarness(t, contextendpoint.Response{Body: nlClassified})
			contextPolicy(t, h, r, tc.eligibility)
			c := initializeWith(t, h, contextConfig)
			res := startAndWait(t, h, c, runTask("eligible-1", "", "", false))
			if eg := res.NetworkContext.Egress; res.Execution.State != "completed" || eg["status"] != tc.status {
				t.Fatalf("result %+v, egress %v", res.Execution, eg)
			}
			for _, execution := range []string{requires, allowlist} {
				res = startAndWait(t, h, c, runTask("eligible-"+strconv.Itoa(len(execution)), "", execution, false))
				got := res.Execution.State
				if got != "completed" {
					got = res.Execution.Reason
				}
				if got != tc.requirementOutcome {
					t.Fatalf("%s: %+v, want %s", execution, res.Execution, tc.requirementOutcome)
				}
			}
			if p, f := len(e.primary.Requests()) > 0, len(e.fallback.Requests()) > 0; p != tc.primary || f != tc.fallback {
				t.Fatalf("primary contacted %t, fallback contacted %t", p, f)
			}
		})
	}
	t.Run("proxy path not eligible", func(t *testing.T) {
		h, r, e := contextHarness(t, contextendpoint.Response{Body: nlClassified})
		contextPolicy(t, h, r, &task.ContextEligibility{Paths: direct, Endpoints: []string{contextPrimary}})
		reg, err := proxyRegistry(r)
		if err != nil {
			t.Fatal(err)
		}
		h.adapters.registry = reg
		proxy, err := loopback.ListenProxy(loopback.ProxyConfig{Resolve: map[string]string{"8.8.8.8:443": e.primary.Address()}})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = proxy.Close() })
		c := initializeWith(t, h, `{"proxy":{"host":"127.0.0.1","port":`+strconv.Itoa(proxy.Port())+`,"connection_code":"outline_mobileproxy"},"network_context":{"primary_endpoint":"`+contextPrimary+`"}}`)
		res := startAndWait(t, h, c, runTask("proxy-1", "", `"network_path":"proxy"`, false))
		if eg := res.NetworkContext.Egress; eg["status"] != "unavailable" || eg["network_path"] != "proxy" {
			t.Fatalf("proxy egress %v", eg)
		}
		res = startAndWait(t, h, c, runTask("proxy-2", "", `"network_path":"proxy",`+requires, false))
		if res.Execution.State != "not_run" || res.Execution.Reason != "network_context_unavailable" {
			t.Fatalf("proxy run requiring context: %+v", res.Execution)
		}
		for _, req := range proxy.Counts().Requests {
			if strings.HasPrefix(req, "CONNECT 8.8.8.8:443 ") {
				t.Fatal("context acquired on a path local policy does not permit")
			}
		}
		// The direct path stays eligible.
		if eg := startAndWait(t, h, c, runTask("direct-1", "", "", false)).NetworkContext.Egress; eg["status"] != "fresh" {
			t.Fatalf("direct egress %v", eg)
		}
	})
}
