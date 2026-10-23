package execution_test

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"testing"
	"time"

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
