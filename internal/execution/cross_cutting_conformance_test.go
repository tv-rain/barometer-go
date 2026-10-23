package execution_test

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/config"
	"github.com/tv-rain/barometer-go/internal/execution"
	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/networkcontext"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/taskdecode"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
	"github.com/tv-rain/barometer-go/internal/validation"
)

// Runtime conformance family cases on the engine (DEV-640). They use the
// manual engine clock, the output-port fake and, for budgets, real TCP sockets
// on 127.0.0.1 through the test-only loopback dialer and exchange servers. They
// are not registered measurement, runtime, platform or field evidence.

// exchangeOn runs Exchange against another server within the same attempt.
func exchangeOn(ctx context.Context, in measurement.Input, p loopback.ExchangeParameters) (measurement.Output, error) {
	return loopback.Exchange{}.Run(ctx, measurement.Input{Parameters: p, Dialer: in.Dialer, Start: in.Start})
}

// CC-EXECUTION-FIELDS-LIMITS: attempt limits apply to each attempt. Attempt 1
// consumes the registered default budgets exactly and ends with a retryable
// reason; attempt 2 of the same run starts with full budgets, one connection
// or byte beyond either limit stops it with resource_limit, and its deadline
// runs from its own start.
// requirement-ref: REQ-EXECUTION-0011
func TestCcExecutionFieldsLimits(t *testing.T) {
	full := loopback.ExchangeParameters{Connections: 8, Exchanges: 1, Request: 4096, Reply: 16384}
	fullUsage := measurement.Usage{BytesSent: 32768, BytesReceived: 131072, Connections: 8}
	srv := exchangeServer(t, 4096, 16384)
	full.Address = srv.Address()
	// short is a server whose reply is one byte larger than its request share
	// allows, for the received-bytes boundary.
	short := exchangeServer(t, 4096, 16385)
	cases := []struct {
		name   string
		second func(ctx context.Context, in measurement.Input) (measurement.Output, error)
		want   want
		usage  measurement.Usage
	}{
		{"full budgets again", func(ctx context.Context, in measurement.Input) (measurement.Output, error) {
			return exchangeOn(ctx, in, full)
		}, completedConnected, fullUsage},
		{"a ninth connection", func(ctx context.Context, in measurement.Input) (measurement.Output, error) {
			if _, err := exchangeOn(ctx, in, full); err != nil {
				return measurement.Output{}, err
			}
			if _, err := in.Dialer.DialContext(ctx, "tcp", srv.Address()); err != nil {
				return measurement.Output{}, measurement.ErrRunner
			}
			return connected(t).out, nil
		}, stoppedByBudget, fullUsage},
		{"one byte beyond max_bytes_sent", func(ctx context.Context, in measurement.Input) (measurement.Output, error) {
			seven := full
			seven.Connections = 7
			if _, err := exchangeOn(ctx, in, seven); err != nil {
				return measurement.Output{}, err
			}
			c, err := in.Dialer.DialContext(ctx, "tcp", srv.Address())
			if err != nil {
				return measurement.Output{}, measurement.ErrRunner
			}
			if _, err := c.Write(make([]byte, 4097)); err != nil {
				return measurement.Output{}, measurement.ErrRunner
			}
			return connected(t).out, nil
		}, stoppedByBudget, measurement.Usage{BytesSent: 28672, BytesReceived: 114688, Connections: 8}},
		{"one byte beyond max_bytes_received", func(ctx context.Context, in measurement.Input) (measurement.Output, error) {
			seven := full
			seven.Connections = 7
			if _, err := exchangeOn(ctx, in, seven); err != nil {
				return measurement.Output{}, err
			}
			last := loopback.ExchangeParameters{Address: short.Address(), Connections: 1, Exchanges: 1, Request: 4096, Reply: 16385}
			return exchangeOn(ctx, in, last)
		}, stoppedByBudget, measurement.Usage{BytesSent: 32768, BytesReceived: 131072, Connections: 8}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t, schedule(2, task.InternalError))
			calls := 0
			r := runnerFunc(func(ctx context.Context, in measurement.Input) (measurement.Output, error) {
				calls++
				if calls == 1 {
					if _, err := exchangeOn(ctx, in, full); err != nil {
						return measurement.Output{}, err
					}
					_ = f.clock.Advance(20 * time.Second)
					return measurement.Output{}, measurement.ErrRunner
				}
				return tc.second(ctx, in)
			})
			o := <-f.start(t, f.run, r)
			f.expect(t, o.rep, o.err, nil, []want{{"incomplete", "internal_error", ""}, tc.want}, []time.Duration{35 * time.Second})
			if o.rep.Attempts[0].Usage != fullUsage || o.rep.Attempts[1].Usage != tc.usage {
				t.Fatalf("usage = %+v, %+v", o.rep.Attempts[0].Usage, o.rep.Attempts[1].Usage)
			}
		})
	}
	t.Run("the deadline runs from each attempt's own start", func(t *testing.T) {
		for _, tc := range []struct {
			name  string
			spend time.Duration
			want  want
		}{
			{"attempt 2 one millisecond within its deadline", attemptDeadline - time.Millisecond, completedConnected},
			{"attempt 2 at its deadline", attemptDeadline, want{"incomplete", "attempt_deadline_exceeded", ""}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				f := newFixture(t, schedule(2, task.InternalError))
				calls := 0
				r := runnerFunc(func(ctx context.Context, in measurement.Input) (measurement.Output, error) {
					calls++
					in.Start()
					if calls == 1 {
						// Attempt 1 uses 20 s; a deadline counted from it
						// would already have passed when attempt 2 starts.
						_ = f.clock.Advance(20 * time.Second)
						return measurement.Output{}, measurement.ErrRunner
					}
					_ = f.clock.Advance(tc.spend)
					if ctx.Err() != nil {
						return measurement.Output{}, ctx.Err()
					}
					return connected(t).out, nil
				})
				o := <-f.start(t, f.run, r)
				f.expect(t, o.rep, o.err, nil, []want{{"incomplete", "internal_error", ""}, tc.want}, []time.Duration{35 * time.Second})
			})
		}
	})
}

// CC-RESULT-EXECUTION-OBJECT: execution states and reasons as the engine
// writes them. An interruption before the test starts is not_run and after it
// incomplete; a recorded outcome stays completed; the first termination wins;
// measurement presence follows the state; and a plan the engine refuses
// produces no Measurement Result.
// requirement-ref: REQ-RESULT-0019
// requirement-ref: REQ-RESULT-0020
// requirement-ref: REQ-RESULT-0021
// requirement-ref: REQ-RESULT-0022
// requirement-ref: REQ-RESULT-0051
// requirement-ref: REQ-RESULT-0052
func TestCcResultExecutionObject(t *testing.T) {
	cases := []struct {
		name  string
		steps []step
		want  want
	}{
		{"completed with a positive outcome", []step{connected(t)}, completedConnected},
		{"completed with a negative outcome", []step{refused(t)}, want{"completed", "", "connection_refused"}},
		{"internal error after the test started", []step{failure()}, want{"incomplete", "internal_error", ""}},
		{"host cancellation before the test started", []step{{err: measurement.ErrRunner, cancel: true, unbegun: true}}, want{"not_run", "cancelled_by_host", ""}},
		{"host cancellation after the test started", []step{{err: measurement.ErrRunner, cancel: true}}, want{"incomplete", "cancelled_by_host", ""}},
		{"deadline before the test started", []step{{err: measurement.ErrRunner, advance: attemptDeadline, unbegun: true}}, want{"not_run", "attempt_deadline_exceeded", ""}},
		{"deadline after the test started", []step{{err: measurement.ErrRunner, advance: attemptDeadline}}, want{"incomplete", "attempt_deadline_exceeded", ""}},
		{"cancellation after the outcome was recorded", []step{{out: connected(t).out, cancel: true}}, completedConnected},
		{"deadline after the outcome was recorded", []step{{out: connected(t).out, advance: attemptDeadline}}, completedConnected},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t, schedule(1), tc.steps...)
			rep, err := f.engine.Run(f.run, f.plan)
			// expect also checks that completed results carry a measurement
			// and incomplete and not_run results here carry none, and that
			// not_run results carry no Network Context.
			f.expect(t, rep, err, nil, []want{tc.want}, nil)
			for _, r := range f.records(t) {
				if (r.Execution.State == "completed") != (r.Measurement != nil) || (r.Execution.State == "completed") != (r.Execution.Reason == "") {
					t.Fatalf("state %q with reason %q and measurement %v", r.Execution.State, r.Execution.Reason, r.Measurement != nil)
				}
			}
		})
	}
	t.Run("the first termination condition is the reason", func(t *testing.T) {
		// The deadline fires first, then the host cancels.
		f := newFixture(t, schedule(1), step{err: measurement.ErrRunner, advance: attemptDeadline, cancel: true})
		rep, err := f.engine.Run(f.run, f.plan)
		f.expect(t, rep, err, nil, []want{{"incomplete", "attempt_deadline_exceeded", ""}}, nil)
		// The host cancels first, then the deadline passes.
		f = newFixture(t, schedule(1))
		r := runnerFunc(func(ctx context.Context, in measurement.Input) (measurement.Output, error) {
			in.Start()
			f.runner.cancel()
			_ = f.clock.Advance(attemptDeadline)
			return measurement.Output{}, ctx.Err()
		})
		o := <-f.start(t, f.run, r)
		f.expect(t, o.rep, o.err, nil, []want{{"incomplete", "cancelled_by_host", ""}}, nil)
	})
	t.Run("a refused plan writes no result", func(t *testing.T) {
		// A malformed plan is a validation failure of the caller, reported as
		// an error, never as a Measurement Result.
		f := newFixture(t, schedule(1), connected(t))
		f.plan.Limits.Connections = 13 // above the registered maximum
		if _, err := f.engine.Run(f.run, f.plan); err == nil {
			t.Fatal("a plan beyond the registered limits ran")
		}
		if len(f.store.Records()) != 0 || f.runner.Calls() != 0 {
			t.Fatal("a refused plan produced a result or an attempt")
		}
	})
}

// Network-condition family cases (DEV-644). Each run uses the production
// direct-path context source over the fake platform network and a recording
// loopback dialer source, so a run that does not start shows no dialer issued
// and no runner call. Egress Network Context requirements are not exercised:
// admission rejects them until egress lookup exists (DEV-647).

// networkTask decodes a TCP-01 task whose execution carries the given members
// after limits, so defaults (vpn_policy any) are the decoder's.
func networkTask(t *testing.T, execution string) (task.NetworkRequirements, error) {
	t.Helper()
	d, err := taskdecode.New([]taskdecode.TestCodec{fixtureCodec{}})
	if err != nil {
		t.Fatal(err)
	}
	if execution != "" {
		execution = "," + execution
	}
	in := fmt.Sprintf(`{"schema_version":"0.1.0","task_id":"campaign-1","test_id":"TCP-01","test_version":"0.1.0","input":{},"options":{},"schedule":{"not_after":%q},"execution":{"limits":{}%s},"safety":{"risk_level":"low","requires_user_consent":false}}`,
		now.Add(time.Hour).Format(time.RFC3339), execution)
	c, err := d.Decode([]byte(in), validation.TimingContext{Origin: "embedded", AcceptedAt: now})
	if err != nil {
		return task.NetworkRequirements{}, err
	}
	return c.Execution.Network, nil
}

// runUnder runs one single-attempt plan with network conditions n while the
// fake platform reports s, and returns the single result. failCapture makes
// the capture fail instead.
func runUnder(t *testing.T, n task.NetworkRequirements, s platform.AccessState, failCapture bool) (*fixture, record) {
	t.Helper()
	f := newFixture(t, schedule(1), connected(t))
	f.plan.Network = n
	f.ctx.net.Set(s)
	if failCapture {
		f.ctx.net.Fail()
	}
	rep, err := f.engine.Run(f.run, f.plan)
	recs := f.records(t)
	if err != nil || len(recs) != 1 || recs[0].RunID != rep.RunID {
		t.Fatalf("run: err %v, %d results", err, len(recs))
	}
	return f, recs[0]
}

// ran reports whether the attempt ran; a run that did not start must have
// recorded not_run with constraints_unsatisfied and issued no network
// operation.
func ran(t *testing.T, f *fixture, r record) bool {
	t.Helper()
	switch {
	case r.Execution.State == "completed" && f.runner.Calls() == 1 && len(f.dialers.Issued()) == 1:
		return true
	case r.Execution.State == "not_run" && r.Execution.Reason == "constraints_unsatisfied" && f.runner.Calls() == 0 && len(f.dialers.Issued()) == 0:
		return false
	}
	t.Fatalf("unexpected attempt: %+v, runner calls %d, dialers %d", r.Execution, f.runner.Calls(), len(f.dialers.Issued()))
	return false
}

// ready is a platform reading that satisfies every access condition.
var ready = platform.AccessState{Generation: 1, Type: platform.NetworkWiFi, VPN: platform.VPNNotDetected, IPv4: platform.AvailabilityYes, IPv6: platform.AvailabilityNo, Validated: platform.AvailabilityYes}

// CC-EXECUTION-FIELDS-NETWORK: a required network condition the platform
// cannot determine is unsatisfied and the attempt is not_run with
// constraints_unsatisfied, with no network operation; unknown never satisfies
// a condition. The same task runs once the platform reports the values. The
// required egress context row stays blocked on DEV-647: until then the static
// gate refuses it, which this case also asserts.
// requirement-ref: REQ-EXECUTION-0031
func TestCcExecutionFieldsNetwork(t *testing.T) {
	cases := []struct {
		name    string
		network string
		unknown func(*platform.AccessState)
	}{
		{"network_type unknown", `"network":{"allowed_transports":["wifi","cellular"]}`, func(s *platform.AccessState) { s.Type = platform.NetworkUnknown }},
		{"vpn_state unknown, exclude", `"network":{"vpn_policy":"exclude"}`, func(s *platform.AccessState) { s.VPN = platform.VPNUnknown }},
		{"vpn_state unknown, require", `"network":{"vpn_policy":"require"}`, func(s *platform.AccessState) { s.VPN = platform.VPNUnknown }},
		{"validated unknown", `"network":{"require_validated":true}`, func(s *platform.AccessState) { s.Validated = platform.AvailabilityUnknown }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			n, err := networkTask(t, tc.network)
			if err != nil {
				t.Fatal(err)
			}
			s := ready
			tc.unknown(&s)
			if runs(t, n, s, false) {
				t.Fatal("ran on an undeterminable condition")
			}
			// A failed capability is as undeterminable as an unknown value.
			if runs(t, n, ready, true) {
				t.Fatal("ran without access context")
			}
			if s := ready; tc.name == "vpn_state unknown, require" {
				s.VPN = platform.VPNDetected
				if !runs(t, n, s, false) {
					t.Fatal("did not run once the platform reported the values")
				}
			} else if !runs(t, n, ready, false) {
				t.Fatal("did not run once the platform reported the values")
			}
		})
	}
	t.Run("required egress context field", func(t *testing.T) {
		e := task.Execution{NetworkPath: "direct", Network: task.NetworkRequirements{VPNPolicy: "any", Context: task.ContextRequirements{RequiredFields: []string{"country_code"}}}}
		if execution.NetworkSatisfiable(e) {
			t.Fatal("egress requirement satisfiable without egress lookup")
		}
	})
}

// CC-EXECUTION-FIELDS-NETWORK-FIELDS: allowed_transports runs only on a
// listed type; values outside the initial list are rejected at validation;
// vpn_policy follows the public table; require_validated runs only on the
// platform's validated yes. The library internal check of REQ-EXECUTION-0016
// is not implemented (GO-SPEC-ISSUE-013), so the row that would use it is
// not executed; Network Context rows stay blocked on DEV-647.
// requirement-ref: REQ-EXECUTION-0016
// requirement-ref: REQ-EXECUTION-0017
func TestCcExecutionFieldsNetworkFields(t *testing.T) {
	types := map[platform.NetworkType]string{platform.NetworkWiFi: "wifi", platform.NetworkCellular: "cellular", platform.NetworkEthernet: "ethernet", platform.NetworkOther: "other", platform.NetworkUnknown: "unknown"}
	for _, list := range [][]string{{"wifi"}, {"cellular"}, {"wifi", "cellular"}} {
		raw, _ := json.Marshal(list)
		n, err := networkTask(t, `"network":{"allowed_transports":`+string(raw)+`}`)
		if err != nil {
			t.Fatal(err)
		}
		for typ, name := range types {
			s := ready
			s.Type = typ
			if got, want := runs(t, n, s, false), slices.Contains(list, name); got != want {
				t.Fatalf("allowed_transports %v on %s: ran %v", list, name, got)
			}
		}
	}
	for _, bad := range []string{`["ethernet"]`, `["other"]`, `["unknown"]`, `["WIFI"]`, `["wifi","bluetooth"]`} {
		if _, err := networkTask(t, `"network":{"allowed_transports":`+bad+`}`); err == nil {
			t.Fatalf("allowed_transports %s accepted", bad)
		}
	}
	for _, policy := range []string{"", `"vpn_policy":"any",`, `"vpn_policy":"exclude",`, `"vpn_policy":"require",`} {
		n, err := networkTask(t, `"network":{`+policy+`"require_validated":false}`)
		if err != nil {
			t.Fatal(err)
		}
		for vpn, want := range vpnTable(n.VPNPolicy) {
			s := ready
			s.VPN = vpn
			if runs(t, n, s, false) != want {
				t.Fatalf("vpn_policy %q with %v", n.VPNPolicy, vpn)
			}
		}
	}
	n, err := networkTask(t, `"network":{"require_validated":true}`)
	if err != nil {
		t.Fatal(err)
	}
	for v, want := range map[platform.Availability]bool{platform.AvailabilityYes: true, platform.AvailabilityNo: false, platform.AvailabilityUnknown: false} {
		s := ready
		s.Validated = v
		if runs(t, n, s, false) != want {
			t.Fatalf("require_validated with %v", v)
		}
	}
}

// vpnTable is the public VPN eligibility table (REQ-EXECUTION-0020).
func vpnTable(policy string) map[platform.VPNState]bool {
	return map[string]map[platform.VPNState]bool{
		"any":     {platform.VPNDetected: true, platform.VPNNotDetected: true, platform.VPNUnknown: true},
		"exclude": {platform.VPNDetected: false, platform.VPNNotDetected: true, platform.VPNUnknown: false},
		"require": {platform.VPNDetected: true, platform.VPNNotDetected: false, platform.VPNUnknown: false},
	}[policy]
}

// CC-EXECUTION-FIELDS-NETWORK-FIELDS-VPN-POLICY: all nine policy and state
// combinations on the engine, omitted vpn_policy as any, unknown carried as
// unknown through normalization and evaluation, and an unregistered policy
// rejected at validation.
// requirement-ref: REQ-EXECUTION-0018
// requirement-ref: REQ-EXECUTION-0019
// requirement-ref: REQ-EXECUTION-0020
func TestCcExecutionFieldsNetworkFieldsVpnPolicy(t *testing.T) {
	for _, policy := range []string{"any", "exclude", "require"} {
		n, err := networkTask(t, `"network":{"vpn_policy":"`+policy+`"}`)
		if err != nil {
			t.Fatal(err)
		}
		for vpn, want := range vpnTable(policy) {
			s := ready
			s.VPN = vpn
			if runs(t, n, s, false) != want {
				t.Fatalf("%s with %v", policy, vpn)
			}
		}
	}
	for _, network := range []string{"", `"network":{}`} {
		n, err := networkTask(t, network)
		if err != nil || n.VPNPolicy != "any" {
			t.Fatalf("omitted vpn_policy decoded as %q", n.VPNPolicy)
		}
	}
	if a := networkcontext.Normalize(platform.AccessState{VPN: platform.VPNUnknown}); a.VPNState != "unknown" ||
		execution.NetworkConditions(task.NetworkRequirements{VPNPolicy: "exclude"}, a) {
		t.Fatal("unknown VPN state mapped to not_detected")
	}
	for _, bad := range []string{"prefer", "ANY", "", "none"} {
		if _, err := networkTask(t, `"network":{"vpn_policy":"`+bad+`"}`); err == nil {
			t.Fatalf("vpn_policy %q accepted", bad)
		}
	}
}

// runs runs one plan under the platform reading and reports whether its
// attempt ran.
func runs(t *testing.T, n task.NetworkRequirements, s platform.AccessState, failCapture bool) bool {
	t.Helper()
	f, r := runUnder(t, n, s, failCapture)
	return ran(t, f, r)
}

// Network path family cases (DEV-645). Proxy runs use the production CONNECT
// dialer and proxy-path context source against the scripted loopback proxy of
// internal/testkit/loopback, which relays to loopback target servers. Runners
// are synthetic; no production TCP-01, TLS-01, HTTP-01 or NET-05 runner
// exists (DEV-454).

// decodeExecution decodes a TCP-01 task whose execution carries the given
// members after limits.
func decodeExecution(t *testing.T, members string) (task.Execution, error) {
	t.Helper()
	d, err := taskdecode.New([]taskdecode.TestCodec{fixtureCodec{}})
	if err != nil {
		t.Fatal(err)
	}
	if members != "" {
		members = "," + members
	}
	in := fmt.Sprintf(`{"schema_version":"0.1.0","task_id":"campaign-1","test_id":"TCP-01","test_version":"0.1.0","input":{},"options":{},"schedule":{"not_after":%q},"execution":{"limits":{}%s},"safety":{"risk_level":"low","requires_user_consent":false}}`,
		now.Add(time.Hour).Format(time.RFC3339), members)
	c, err := d.Decode([]byte(in), validation.TimingContext{Origin: "embedded", AcceptedAt: now})
	if err != nil {
		return task.Execution{}, err
	}
	return c.Execution, nil
}

// pathSelection asserts the decoded path selection: omitted is direct, only
// direct and proxy are accepted, and the task carries no endpoint, category,
// credential, proxy identity or fallback.
func pathSelection(t *testing.T) {
	t.Helper()
	for members, want := range map[string]string{"": "direct", `"network_path":"direct"`: "direct", `"network_path":"proxy"`: "proxy"} {
		if e, err := decodeExecution(t, members); err != nil || e.NetworkPath != want {
			t.Fatalf("%s decoded as %q, %v", members, e.NetworkPath, err)
		}
	}
	for _, members := range []string{
		`"network_path":"socks"`, `"network_path":"PROXY"`, `"network_path":""`, `"network_path":null`,
		`"network_path":"proxy","proxy":{"host":"127.0.0.1","port":54321}`, `"network_path":"proxy","proxy_id":"p1"`,
		`"network_path":"proxy","endpoint":"127.0.0.1:54321"`, `"network_path":"proxy","connection_code":"outline_mobileproxy"`,
		`"network_path":"proxy","proxy_authorization":"Basic Y2FuYXJ5"`, `"network_path":"proxy","fallback":"direct"`,
	} {
		if _, err := decodeExecution(t, members); err == nil {
			t.Fatalf("%s accepted", members)
		}
	}
}

// deniedPreconditions asserts each pre-attempt denial is not_run with its
// reason, with no context capture, endpoint contact or runner call.
func deniedPreconditions(t *testing.T) {
	t.Helper()
	for _, reason := range []task.ExecutionReason{task.ProxyNotConfigured, task.ProxyNotPermitted, task.ProxyPathUnsupported} {
		srv := echoServer(t)
		p := scriptedProxy(t, loopback.ProxyConfig{})
		f := newFixture(t, schedule(1))
		pc := f.onProxy(t, p.Port())
		f.plan.Path = pathCheck(reason)
		f.plan.Parameters = echoOnce(srv)
		o := <-f.start(t, f.run, loopback.Exchange{})
		f.expect(t, o.rep, o.err, nil, []want{{"not_run", string(reason), ""}}, nil)
		if pc.Calls() != 0 || p.Counts().Accepted != 0 || srv.Accepted() != 0 {
			t.Fatalf("%s: captures %d, proxy accepts %d, target accepts %d", reason, pc.Calls(), p.Counts().Accepted, srv.Accepted())
		}
	}
}

// proxyRun runs one single-attempt Exchange plan through the proxy endpoint
// at port against srv and returns its fixture and outcome.
func proxyRun(t *testing.T, port int, srv *loopback.Server) (*fixture, outcome) {
	t.Helper()
	f := newFixture(t, schedule(1))
	f.onProxy(t, port)
	f.plan.Parameters = echoOnce(srv)
	return f, <-f.start(t, f.run, loopback.Exchange{})
}

// pathFailures injects every path failure and asserts the explicit state, the
// proxy egress and zero contact with the target outside the proxy.
func pathFailures(t *testing.T) {
	t.Helper()
	for name, tc := range map[string]struct {
		head   string
		reason string
	}{
		"unavailable": {"", "proxy_unavailable"},
		"refused":     {"HTTP/1.1 407 Proxy Authentication Required\r\nProxy-Authenticate: Basic\r\n\r\n", "proxy_connect_failed"},
		"SOCKS":       {"\x05\xff", "proxy_connect_failed"},
		"HTTP/1.0":    {"HTTP/1.0 200 OK\r\n\r\n", "proxy_connect_failed"},
		"truncated":   {"HTTP/1.1 200 OK\r\n", "proxy_connect_failed"},
	} {
		srv := echoServer(t)
		port := closedPort(t)
		if tc.head != "" {
			port = scriptedProxy(t, loopback.ProxyConfig{Mode: loopback.ProxyReply, Head: []byte(tc.head)}).Port()
		}
		f, o := proxyRun(t, port, srv)
		f.expect(t, o.rep, o.err, nil, []want{{"incomplete", tc.reason, ""}}, nil)
		assertProxyEgress(t, f.store.Records()[0], f.records(t)[0], port)
		if srv.Accepted() != 0 {
			t.Fatalf("%s: the target was dialed outside the proxy", name)
		}
	}
}

// CC-EXECUTION-FIELDS-NETWORK-PATH: omitted network_path is direct and proxy
// is selected explicitly; each denied proxy precondition and each injected
// path failure is an explicit state with zero direct fallback. The NET-05 row,
// one path snapshot across every profile step, needs the NET-05 runner
// (DEV-454), so the family stays planned; a synthetic runner's streams within
// one attempt share its snapshot, which TestEveryStreamOpensItsOwnTunnel
// shows.
// requirement-ref: REQ-EXECUTION-0006
// requirement-ref: REQ-EXECUTION-0007
// requirement-ref: REQ-EXECUTION-0008
// requirement-ref: REQ-EXECUTION-0009
// requirement-ref: REQ-EXECUTION-0010
func TestCcExecutionFieldsNetworkPath(t *testing.T) {
	pathSelection(t)
	deniedPreconditions(t)
	pathFailures(t)
	srv := echoServer(t)
	p := scriptedProxy(t, loopback.ProxyConfig{})
	f, o := proxyRun(t, p.Port(), srv)
	f.expect(t, o.rep, o.err, nil, []want{completedConnected}, nil)
	assertProxyEgress(t, f.store.Records()[0], f.records(t)[0], p.Port())
}

// countingCheck counts pre-attempt path checks.
type countingCheck struct {
	mu    sync.Mutex
	calls int
}

func (c *countingCheck) Check() task.ExecutionReason {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls++
	return ""
}

// countingPreconditions counts attempt-start checks.
type countingPreconditions struct{ countingCheck }

func (c *countingPreconditions) Satisfied(context.Context) bool {
	c.Check()
	return true
}

// CC-EXECUTION-ENVELOPE: constraints and the path check are evaluated
// immediately before every attempt, including a retry, and the selected path
// changes how the stream is established but not the registered operation:
// the same synthetic runner against the same target records the same
// measurement on both paths, and only the egress path differs.
// requirement-ref: REQ-EXECUTION-0001
// requirement-ref: REQ-EXECUTION-0002
func TestCcExecutionEnvelope(t *testing.T) {
	f := newFixture(t, schedule(2, task.ProxyUnavailable))
	f.onProxy(t, closedPort(t))
	pre, check := &countingPreconditions{}, &countingCheck{}
	f.plan.Preconditions, f.plan.Path = pre, check
	f.plan.Parameters = echoOnce(echoServer(t))
	o := <-f.start(t, f.run, loopback.Exchange{})
	unavailable := want{"incomplete", "proxy_unavailable", ""}
	f.expect(t, o.rep, o.err, nil, []want{unavailable, unavailable}, []time.Duration{35 * time.Second})
	if pre.calls != 2 || check.calls != 2 {
		t.Fatalf("preconditions %d, path checks %d", pre.calls, check.calls)
	}

	measurementOf := func(raw []byte) string {
		var r struct {
			Measurement json.RawMessage `json:"measurement"`
		}
		if err := json.Unmarshal(raw, &r); err != nil {
			t.Fatal(err)
		}
		return string(r.Measurement)
	}
	srv := echoServer(t)
	direct := newFixture(t, schedule(1))
	direct.plan.Parameters = echoOnce(srv)
	od := <-direct.start(t, direct.run, loopback.Exchange{})
	direct.expect(t, od.rep, od.err, nil, []want{completedConnected}, nil)
	p := scriptedProxy(t, loopback.ProxyConfig{})
	proxied, op := proxyRun(t, p.Port(), srv)
	proxied.expect(t, op.rep, op.err, nil, []want{completedConnected}, nil)
	d, pr := direct.store.Records()[0], proxied.store.Records()[0]
	if measurementOf(d) != measurementOf(pr) || od.rep.Attempts[0].Usage != op.rep.Attempts[0].Usage {
		t.Fatalf("direct %s, proxied %s", measurementOf(d), measurementOf(pr))
	}
	if _, e := accessOf(t, direct.records(t)[0]); e["network_path"] != "direct" {
		t.Fatalf("direct egress %v", e)
	}
	assertProxyEgress(t, pr, proxied.records(t)[0], p.Port())
}

// CC-PROXY-HTTP-CONNECT-BEHAVIOR: each stream opens one tunnel with exactly
// one bounded CONNECT for the exact authority, an IP address and port or a
// normalized hostname resolved behind the proxy, with no credentials, cookies
// or cleartext target request. Only a valid HTTP/1.1 2xx within the deadline
// establishes the tunnel; TLS inside it keeps its SNI and certificate
// validation; a second stream, as a followed redirect opens, gets a fresh
// tunnel; cancellation and the deadline close the exchange.
// requirement-ref: REQ-PROXY-0017
// requirement-ref: REQ-PROXY-0018
// requirement-ref: REQ-PROXY-0019
// requirement-ref: REQ-PROXY-0020
func TestCcProxyHttpConnectBehavior(t *testing.T) {
	target := newTLSTarget(t)
	_, port, _ := net.SplitHostPort(target.Address())
	p := scriptedProxy(t, loopback.ProxyConfig{Resolve: map[string]string{tlsName + ":" + port: target.Address()}})
	for _, address := range []string{target.Address(), "TARGET.example.test:" + port} {
		c, err := dialOnce(t, p.Port(), "tcp", address)
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err = tlsEcho(ctx, c, tlsName, target.roots)
		cancel()
		if err != nil {
			t.Fatalf("TLS through the tunnel to %s: %v", address, err)
		}
	}
	// Certificate validation is unchanged inside the tunnel.
	c, err := dialOnce(t, p.Port(), "tcp", target.Address())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var unknown x509.HostnameError
	if err := tlsEcho(ctx, c, "other.example.test", target.roots); !errors.As(err, &unknown) {
		t.Fatalf("mismatched name accepted through the tunnel: %v", err)
	}
	counts := idle(t, p)
	wantReqs := []string{connectRequest(target.Address()), connectRequest(tlsName + ":" + port), connectRequest(target.Address())}
	if strings.Join(counts.Requests, "") != strings.Join(wantReqs, "") || counts.Accepted != 3 {
		t.Fatalf("requests = %q", counts.Requests)
	}
	if sni := target.SNI(); len(sni) != 3 || sni[0] != tlsName || sni[1] != tlsName || sni[2] != "other.example.test" {
		t.Fatalf("SNI = %q", sni)
	}
	for _, r := range counts.Requests {
		lower := strings.ToLower(r)
		for _, forbidden := range []string{"authorization", "cookie", "user-agent", "get ", "post "} {
			if strings.Contains(lower, forbidden) {
				t.Fatalf("CONNECT request carries %q: %q", forbidden, r)
			}
		}
	}
	// Valid 2xx variants establish a tunnel; anything else does not.
	for head, ok := range map[string]bool{
		"HTTP/1.1 200 Connection established\r\n\r\n": true, "HTTP/1.1 204\r\n\r\n": true, "HTTP/1.1 299 Synthetic\r\nX-A: b\r\n\r\n": true,
		"HTTP/1.1 302 Found\r\nLocation: http://example.test/\r\n\r\n": false, "HTTP/1.1 200OK\r\n\r\n": false, "HTTP/2 200\r\n\r\n": false,
		"HTTP/1.1 200 OK\r\n bad: fold\r\n\r\n": false, "HTTP/1.1 200 OK\n\n": false, "": false,
	} {
		rp := scriptedProxy(t, loopback.ProxyConfig{Mode: loopback.ProxyReply, Head: []byte(head)})
		c, err := dialOnce(t, rp.Port(), "tcp", "127.0.0.1:9")
		if (err == nil) != ok {
			t.Fatalf("%q: established %v", head, err == nil)
		}
		if c != nil {
			_ = c.Close()
		}
	}
	// Fresh tunnels per stream and per attempt, the deadline, cancellation
	// and the endpoint failures.
	t.Run("streams", TestEveryStreamOpensItsOwnTunnel)
	t.Run("late", TestLateConnect)
	t.Run("failures", func(t *testing.T) { pathFailures(t) })
}

// CC-PROXY-TASK-SELECTION: a task selects only direct or proxy and carries no
// proxy identity, endpoint or connection_code; a proxy task runs only when a
// proxy is configured, local policy permits it and the test is compatible,
// and otherwise records which precondition failed.
// requirement-ref: REQ-PROXY-0021
// requirement-ref: REQ-PROXY-0022
func TestCcProxyTaskSelection(t *testing.T) {
	pathSelection(t)
	deniedPreconditions(t)
}

// CC-PROXY-RESULT-METADATA: direct results record network_path direct
// without connection_code; proxied results, successful or failed and with
// egress unavailable, record network_path proxy and the configured
// connection_code. No result carries the proxy host, port or generation, a
// CONNECT request or authority echo, the proxy's response headers or body, or
// a CONNECT status class.
// requirement-ref: REQ-PROXY-0031
// requirement-ref: REQ-PROXY-0032
// requirement-ref: REQ-PROXY-0033
func TestCcProxyResultMetadata(t *testing.T) {
	srv := echoServer(t)
	direct := func(address string, w want) {
		f := newFixture(t, schedule(1))
		f.plan.Parameters = loopback.ExchangeParameters{Address: address, Connections: 1, Exchanges: 1, Request: len(loopback.Payload), Reply: len(loopback.Payload)}
		o := <-f.start(t, f.run, loopback.Exchange{})
		f.expect(t, o.rep, o.err, nil, []want{w}, nil)
		if _, e := accessOf(t, f.records(t)[0]); len(e) != 3 || e["network_path"] != "direct" {
			t.Fatalf("direct egress %v", e)
		}
	}
	direct(srv.Address(), completedConnected)
	direct("127.0.0.1:"+strconv.Itoa(closedPort(t)), want{"completed", "", "socket_error"})

	ok := scriptedProxy(t, loopback.ProxyConfig{})
	refused := scriptedProxy(t, loopback.ProxyConfig{Mode: loopback.ProxyReply,
		Head: []byte("HTTP/1.1 403 Forbidden\r\nX-Canary: canary-header\r\nContent-Length: 11\r\n\r\ncanary-body")})
	for _, tc := range []struct {
		port int
		want want
	}{
		{ok.Port(), completedConnected},
		{refused.Port(), want{"incomplete", "proxy_connect_failed", ""}},
		{closedPort(t), want{"incomplete", "proxy_unavailable", ""}},
	} {
		f, o := proxyRun(t, tc.port, srv)
		f.expect(t, o.rep, o.err, nil, []want{tc.want}, nil)
		raw := f.store.Records()[0]
		assertProxyEgress(t, raw, f.records(t)[0], tc.port)
		for _, leak := range []string{srv.Address(), "canary-header", "canary-body", "X-Canary", "403", "status_class"} {
			if strings.Contains(string(raw), leak) {
				t.Fatalf("result contains %q: %s", leak, raw)
			}
		}
	}
}

// CC-PROXY-PROXY-CONFIGURATION: the proxy object accepts a loopback address
// or localhost, a port from 1 to 65535 and a bounded lowercase connection
// category, rejects every other field, transport, credential or a
// non-loopback endpoint, and an invalid replacement leaves the previous
// configuration unchanged. The dialer accepts only the same endpoints.
// requirement-ref: REQ-PROXY-0005
// requirement-ref: REQ-PROXY-0006
// requirement-ref: REQ-PROXY-0007
func TestCcProxyProxyConfiguration(t *testing.T) {
	code64 := strings.Repeat("a", 64)
	d, err := config.NewDecoder(config.Contract, config.Approvals{ConnectionCodes: []string{"outline_mobileproxy", code64, "a.b-c_1"}})
	if err != nil {
		t.Fatal(err)
	}
	proxyJSON := func(host, port, code, extra string) []byte {
		return []byte(fmt.Sprintf(`{"proxy":{"host":%q,"port":%s,"connection_code":%q%s}}`, host, port, code, extra))
	}
	for _, in := range [][]byte{
		proxyJSON("127.0.0.1", "1", "outline_mobileproxy", ""), proxyJSON("127.255.255.254", "65535", "outline_mobileproxy", ""),
		proxyJSON("::1", "54321", code64, ""), proxyJSON("localhost", "54321", "a.b-c_1", ""),
	} {
		s, err := d.Decode(in)
		p, ok := s.Proxy()
		if err != nil || !ok {
			t.Fatalf("%s rejected: %v", in, err)
		}
		if _, err := execution.NewProxyDialers(p.Host, p.Port); err != nil {
			t.Fatalf("dialer rejected %s", in)
		}
	}
	for _, in := range [][]byte{
		proxyJSON("192.0.2.1", "54321", "outline_mobileproxy", ""), proxyJSON("0.0.0.0", "54321", "outline_mobileproxy", ""),
		proxyJSON("::", "54321", "outline_mobileproxy", ""), proxyJSON("LOCALHOST", "54321", "outline_mobileproxy", ""),
		proxyJSON("proxy.example.test", "54321", "outline_mobileproxy", ""), proxyJSON("127.0.0.1:54321", "54321", "outline_mobileproxy", ""),
		proxyJSON("fe80::1%lo0", "54321", "outline_mobileproxy", ""), proxyJSON("", "54321", "outline_mobileproxy", ""),
		proxyJSON("127.0.0.1", "0", "outline_mobileproxy", ""), proxyJSON("127.0.0.1", "65536", "outline_mobileproxy", ""),
		proxyJSON("127.0.0.1", `"54321"`, "outline_mobileproxy", ""), proxyJSON("127.0.0.1", "54321.5", "outline_mobileproxy", ""),
		proxyJSON("127.0.0.1", "54321", code64+"a", ""), proxyJSON("127.0.0.1", "54321", "Outline", ""),
		proxyJSON("127.0.0.1", "54321", "", ""), proxyJSON("127.0.0.1", "54321", "user@example", ""),
		proxyJSON("127.0.0.1", "54321", "unapproved", ""),
		proxyJSON("127.0.0.1", "54321", "outline_mobileproxy", `,"username":"u"`), proxyJSON("127.0.0.1", "54321", "outline_mobileproxy", `,"password":"p"`),
		proxyJSON("127.0.0.1", "54321", "outline_mobileproxy", `,"type":"socks5"`), proxyJSON("127.0.0.1", "54321", "outline_mobileproxy", `,"tls":true`),
		proxyJSON("127.0.0.1", "54321", "outline_mobileproxy", `,"id":"p1"`), proxyJSON("127.0.0.1", "54321", "outline_mobileproxy", `,"display_name":"x"`),
		proxyJSON("127.0.0.1", "54321", "outline_mobileproxy", `,"pac":"x"`), proxyJSON("127.0.0.1", "54321", "outline_mobileproxy", `,"chain":[]`),
		[]byte(`{"proxy":[{"host":"127.0.0.1","port":1,"connection_code":"outline_mobileproxy"}]}`),
	} {
		if _, err := d.Decode(in); err == nil {
			t.Fatalf("%s accepted", in)
		}
	}
	current, err := d.Decode(proxyJSON("127.0.0.1", "54321", "outline_mobileproxy", ""))
	if err != nil {
		t.Fatal(err)
	}
	next, err := d.WithProxy(current, []byte(`{"host":"192.0.2.1","port":1,"connection_code":"outline_mobileproxy"}`))
	if p, _ := next.Proxy(); err == nil || p != (config.Proxy{Host: "127.0.0.1", Port: 54321, ConnectionCode: "outline_mobileproxy"}) {
		t.Fatalf("invalid replacement changed the configuration: %+v", p)
	}
	TestProxyEndpointIsLoopback(t)
}

// CC-PROXY-OVERVIEW: the configuration names no tests or capabilities; a
// proxy attempt opens one TCP connection to the loopback endpoint and sends an
// HTTP/1.1 CONNECT request line, with no other proxy transport; a SOCKS or
// non-HTTP/1.1 reply fails the attempt without direct fallback; a
// configuration that adds a transport selector, a supported-tests list or a
// non-loopback host is rejected and the previous one stays active. The
// DNS-01-over-UDP row needs the DNS-01 codec and proxy rule (DEV-454), so the
// family stays planned; the CONNECT dialer itself refuses a non-stream dial
// without contact (TestConnectAuthority).
// requirement-ref: REQ-PROXY-0001
func TestCcProxyOverview(t *testing.T) {
	d, err := config.NewDecoder(config.Contract, config.Approvals{ConnectionCodes: []string{"outline_mobileproxy"}})
	if err != nil {
		t.Fatal(err)
	}
	p := scriptedProxy(t, loopback.ProxyConfig{})
	base := fmt.Sprintf(`"host":"127.0.0.1","port":%d,"connection_code":"outline_mobileproxy"`, p.Port())
	current, err := d.Decode([]byte(`{"proxy":{` + base + `}}`))
	if err != nil {
		t.Fatal(err)
	}
	for _, extra := range []string{`,"type":"socks5"`, `,"supported_tests":["TCP-01"]`, `,"capabilities":["tcp"]`} {
		if next, err := d.WithProxy(current, []byte(`{`+base+extra+`}`)); err == nil || next != current {
			t.Fatalf("%s accepted", extra)
		}
	}
	if next, err := d.WithProxy(current, []byte(`{"host":"192.0.2.1","port":54321,"connection_code":"outline_mobileproxy"}`)); err == nil || next != current {
		t.Fatal("non-loopback replacement accepted")
	}
	cfg, _ := current.Proxy()
	srv := echoServer(t)
	f, o := proxyRun(t, cfg.Port, srv)
	f.expect(t, o.rep, o.err, nil, []want{completedConnected}, nil)
	c := idle(t, p)
	if c.Accepted != 1 || len(c.Requests) != 1 || !strings.HasPrefix(c.Requests[0], "CONNECT "+srv.Address()+" HTTP/1.1\r\n") {
		t.Fatalf("proxy counts = %+v", c)
	}
	for _, head := range []string{"\x05\x00", "HTTP/1.0 200 Connection established\r\n\r\n"} {
		srv := echoServer(t)
		rp := scriptedProxy(t, loopback.ProxyConfig{Mode: loopback.ProxyReply, Head: []byte(head)})
		f, o := proxyRun(t, rp.Port(), srv)
		f.expect(t, o.rep, o.err, nil, []want{{"incomplete", "proxy_connect_failed", ""}}, nil)
		if srv.Accepted() != 0 {
			t.Fatal("the target was dialed outside the proxy")
		}
	}
}
