package barometer

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/config"
	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

type net05Result struct {
	Schema string `json:"schema_version"`
	Test   struct {
		ID, Version   string
		ResultVersion string `json:"result_version"`
	}
	Execution      struct{ State, Reason string }
	NetworkContext json.RawMessage `json:"network_context"`
	Measurement    *struct {
		Outcome string
		Profile struct {
			ID      string `json:"profile_id"`
			Version int64  `json:"profile_version"`
		}
		Validation struct{ State, Source string } `json:"network_validation"`
		Steps      []struct {
			ID   string `json:"step_id"`
			Test struct {
				ID, Version   string
				ResultVersion string `json:"result_version"`
			}
			State       string
			Measurement *struct{ Outcome string }
		}
		StoppedBefore string `json:"stopped_before_step"`
	}
}

func net05Wire(t *testing.T, raw []byte) net05Result {
	t.Helper()
	var v net05Result
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	for _, canary := range []string{"8.8.8.8", "/check", "header-canary", "body-canary", loopback.CertificateCanary, "CANARY-RAW-ERROR", "broad_restriction_suspected", "allowlist_detected"} {
		if strings.Contains(string(raw), canary) {
			t.Fatalf("retained canary %s", canary)
		}
	}
	if v.Measurement != nil && (v.Measurement.Profile.ID != "baseline" || v.Measurement.Profile.Version != 1 || v.Measurement.Validation.Source != "platform") {
		t.Fatal("profile/context identity")
	}
	if v.Schema != "0.1.0" || v.Test.ID != "NET-05" || v.Test.ResultVersion != "0.1.0" {
		t.Fatal("outer versions")
	}
	if v.Measurement == nil {
		t.Fatalf("missing expected profile measurement: %s", raw)
	}
	for _, s := range v.Measurement.Steps {
		if s.Test.Version != "0.1.0" || s.Test.ResultVersion != "0.1.0" {
			t.Fatal("step versions")
		}
	}
	return v
}
func readyComposite(h *harness) {
	s := defaultAccess
	s.Validated = platform.AvailabilityYes
	h.run.network.Set(s)
}
func net05HTTPHarness(t *testing.T, raw string, stall bool) (*harness, *httpTransport, *loopback.HTTPServer) {
	t.Helper()
	h, d := net05AdmissionHarness(t, nil)
	ca, err := loopback.NewTestCA()
	if err != nil {
		t.Fatal(err)
	}
	cert, err := ca.Issue("8.8.8.8", false)
	if err != nil {
		t.Fatal(err)
	}
	srv, err := loopback.ListenHTTP(loopback.HTTPConfig{Certificate: cert, Raw: []byte(raw), Stall: stall})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = srv.Close() })
	h.adapters.measurementRoots = ca.Pool()
	h.adapters.approvals = config.Approvals{ConnectionCodes: []string{"outline_mobileproxy"}}
	d.dial = func(ctx context.Context, network, address string) (measurement.Conn, error) {
		if address != "8.8.8.8:443" {
			t.Error("unexpected contact")
			return nil, fmt.Errorf("CANARY-RAW-ERROR")
		}
		var nd net.Dialer
		return nd.DialContext(ctx, "tcp", srv.Address())
	}
	readyComposite(h)
	return h, d, srv
}

// M-NET-05-EXECUTION and OUTPUT execute through production admission, runtime,
// shared DNS/TCP/TLS/HTTP operations, strict result validation and pending output.
// requirement-ref: REQ-NET-05-0006
// requirement-ref: REQ-NET-05-0007
// requirement-ref: REQ-NET-05-0009
// requirement-ref: REQ-NET-05-0039
// requirement-ref: REQ-NET-05-0040
// requirement-ref: REQ-NET-05-0044
// requirement-ref: REQ-NET-05-0048
func TestMNet05ExecutionAndOutput(t *testing.T) {
	h, d, _ := net05HTTPHarness(t, responseHTTP("body-canary"), false)
	h.adapters.run.resolver = dnsSystemResolver{func(context.Context, string, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("8.8.8.8")}, nil
	}}
	steps := primitiveSteps()
	steps[0] = profileStep("dns", "DNS-01", `{"hostname":"example.com","record_type":"A"}`, `{}`)
	c := h.initialized(t)
	defer c.Close()
	raw := tcpRaw(t, h, c, net05Task("composite-success", profileInput(steps...), `{}`, ""))
	v := net05Wire(t, raw)
	if v.Execution.State != "completed" || v.Measurement == nil || v.Measurement.Outcome != "completed" || len(v.Measurement.Steps) != 4 || d.calls.Load() != 3 {
		t.Fatalf("execution %s", raw)
	}
	for i, want := range []string{"resolved", "connected", "alpn_no_overlap", "expected_response"} {
		s := v.Measurement.Steps[i]
		if s.State != "completed" || s.Measurement == nil || s.Measurement.Outcome != want {
			t.Fatalf("step %d: %s", i, raw)
		}
	}
	if strings.Count(string(raw), `"network_context"`) != 1 || v.Measurement.StoppedBefore != "" {
		t.Fatal("duplicated context or placeholder")
	}
}

// CC-PROXY-LIFECYCLE-AND-RENEWAL, CC-CONTEXT-PATH-CONSISTENCY and
// CC-EXECUTION-FIELDS-NETWORK-PATH NET-05 rows: old-generation TCP completes,
// HTTP is interrupted, and the final TCP never contacts either path.
// requirement-ref: REQ-PROXY-0033
// requirement-ref: REQ-NET-05-0040
// requirement-ref: REQ-NET-05-0045
// requirement-ref: REQ-NET-05-0061
func TestCcNet05ProxyLifecycleAndRenewal(t *testing.T) {
	for _, clear := range []bool{false, true} {
		t.Run(fmt.Sprint(clear), func(t *testing.T) {
			h, d, srv := net05HTTPHarness(t, "", true)
			old, err := loopback.ListenProxy(loopback.ProxyConfig{Resolve: map[string]string{"8.8.8.8:443": srv.Address()}})
			if err != nil {
				t.Fatal(err)
			}
			defer old.Close()
			next, err := loopback.ListenProxy(loopback.ProxyConfig{Mode: loopback.ProxyHold})
			if err != nil {
				t.Fatal(err)
			}
			defer next.Close()
			c := h.client(t)
			if err = c.Initialize(proxyConfig(old.Port())); err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			steps := []string{primitiveSteps()[1], primitiveSteps()[3], profileStep("last", "TCP-01", `{"ip_address":"8.8.8.8","port":443}`, `{}`)}
			rh, err := c.StartRun(h.admit(t, c, net05Task("old-profile", profileInput(steps...), `{}`, `"network_path":"proxy"`)))
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err = srv.Wait(ctx, func(s loopback.HTTPCounts) bool { return len(s.Requests) > 0 }); err != nil {
				t.Fatal("HTTP not reached", err)
			}
			if clear {
				err = c.ClearProxy()
			} else {
				err = c.SetProxy(strings.TrimSuffix(strings.TrimPrefix(proxyConfig(next.Port()), `{"proxy":`), `}`))
			}
			if err != nil {
				t.Fatal(err)
			}
			waitTerminal(t, c, rh)
			if err = old.WaitIdle(ctx); err != nil {
				t.Fatal(err)
			}
			recs := h.run.output.Records()
			if len(recs) != 1 {
				t.Fatal("result lost")
			}
			v := net05Wire(t, recs[0])
			if v.Execution.State != "incomplete" || v.Execution.Reason != "proxy_reconfigured" || v.Measurement == nil || v.Measurement.Outcome != "incomplete" || len(v.Measurement.Steps) != 2 || v.Measurement.Steps[0].State != "completed" || v.Measurement.Steps[1].State != "incomplete" || v.Measurement.Steps[1].Measurement != nil || v.Measurement.StoppedBefore != "last" {
				t.Fatalf("partial result: %s", recs[0])
			}
			if old.Counts().Accepted != 2 || next.Counts().Accepted != 0 || d.calls.Load() != 0 {
				t.Fatal("new proxy/direct fallback or remaining step executed")
			}
			if strings.Count(string(recs[0]), `"network_context"`) != 1 || !strings.Contains(string(v.NetworkContext), `"network_path":"proxy"`) {
				t.Fatal("path snapshot lost")
			}
		})
	}
}

// Required platform validation is evaluated before primitives. Task-level
// require_validated remains the earlier not_run/start-reservation constraint.
// requirement-ref: REQ-NET-05-0061
func TestMNet05PlatformValidationGate(t *testing.T) {
	for _, state := range []platform.Availability{platform.AvailabilityNo, platform.AvailabilityUnknown} {
		h, d := net05AdmissionHarness(t, nil)
		a := defaultAccess
		a.Validated = state
		h.run.network.Set(a)
		c := h.initialized(t)
		raw := tcpRaw(t, h, c, net05Task("validation", profileInput(primitiveSteps()[1]), `{}`, ""))
		v := net05Wire(t, raw)
		want := "network_not_validated"
		if state == platform.AvailabilityUnknown {
			want = "inconclusive"
		}
		if v.Execution.State != "completed" || v.Measurement.Outcome != want || len(v.Measurement.Steps) != 0 || d.calls.Load() != 0 {
			t.Fatal("validation gate")
		}
		data := net05Task("task-validation", profileInput(primitiveSteps()[1]), `{}`, `"network":{"require_validated":true}`)
		if rh, e := c.StartRun(h.admit(t, c, data)); rh != "" || codeOf(e) != "run_not_due" {
			t.Fatal("task constraint became measurement")
		}
		c.Close()
	}
}

// The internal policy exception is affirmative and cannot be selected by a
// task. Known negative platform context remains explicit; unknown never runs.
// requirement-ref: REQ-NET-05-0043
func TestNet05TrustedValidationPolicy(t *testing.T) {
	for _, validation := range []platform.Availability{platform.AvailabilityNo, platform.AvailabilityUnknown} {
		h, d := net05AdmissionHarness(t, nil)
		h.adapters.run.net05AllowUnvalidated = true
		s := defaultAccess
		s.Validated = validation
		h.run.network.Set(s)
		c := h.initialized(t)
		raw := tcpRaw(t, h, c, net05Task("local-validation-policy", profileInput(primitiveSteps()[1]), `{}`, ""))
		v := net05Wire(t, raw)
		if validation == platform.AvailabilityNo {
			if v.Measurement.Outcome != "completed" || len(v.Measurement.Steps) != 1 || d.calls.Load() != 1 {
				t.Fatal("trusted policy exception ignored")
			}
		} else if v.Measurement.Outcome != "inconclusive" || d.calls.Load() != 0 {
			t.Fatal("unknown validation authorized contact")
		}
		c.Close()
	}
}

func TestNet05TLSPrimitiveSuccess(t *testing.T) {
	h, d := net05AdmissionHarness(t, func(in *task.PolicyInput) {
		for i := range in.Targets {
			if in.Targets[i].Test.ID == "TLS-01" {
				in.Targets[i].Tuple.ALPN = nil
			}
		}
	})
	ca, err := loopback.NewTestCA()
	if err != nil {
		t.Fatal(err)
	}
	cert, err := ca.Issue("example.com", false)
	if err != nil {
		t.Fatal(err)
	}
	srv, err := loopback.ListenHTTP(loopback.HTTPConfig{Certificate: cert})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	h.adapters.measurementRoots = ca.Pool()
	d.dial = func(ctx context.Context, _, _ string) (measurement.Conn, error) {
		var nd net.Dialer
		return nd.DialContext(ctx, "tcp", srv.Address())
	}
	readyComposite(h)
	c := h.initialized(t)
	defer c.Close()
	step := profileStep("tls", "TLS-01", `{"ip_address":"8.8.8.8","port":443,"server_name":"example.com"}`, `{}`)
	raw := tcpRaw(t, h, c, net05Task("tls-success", profileInput(step), `{}`, ""))
	v := net05Wire(t, raw)
	if v.Execution.State != "completed" || v.Measurement.Outcome != "completed" || v.Measurement.Steps[0].Measurement.Outcome != "handshake_succeeded" || d.calls.Load() != 1 {
		t.Fatalf("TLS primitive: %s", raw)
	}
}
func TestNet05DirectDNSPrimitives(t *testing.T) {
	for _, transport := range []string{"udp", "tcp"} {
		t.Run(transport, func(t *testing.T) {
			h, d := net05AdmissionHarness(t, func(in *task.PolicyInput) {
				for i := range in.Targets {
					if in.Targets[i].Role == task.ResolverTarget {
						in.Targets[i].Tuple.Scheme = transport
					}
				}
			})
			server, err := loopback.ListenDNS(oneAnswer)
			if err != nil {
				t.Fatal(err)
			}
			defer server.Close()
			d.dial = func(ctx context.Context, network, address string) (measurement.Conn, error) {
				if address != "8.8.8.8:53" {
					t.Error("wrong resolver")
				}
				var nd net.Dialer
				return nd.DialContext(ctx, network, server.Address())
			}
			readyComposite(h)
			c := h.initialized(t)
			defer c.Close()
			step := profileStep("dns", "DNS-01", `{"hostname":"example.com","record_type":"A","resolver":{"ip_address":"8.8.8.8"}}`, `{"resolver_mode":"direct","transport":"`+transport+`"}`)
			raw := tcpRaw(t, h, c, net05Task("dns-primitive", profileInput(step), `{}`, ""))
			v := net05Wire(t, raw)
			if v.Execution.State != "completed" || v.Measurement.Outcome != "completed" || v.Measurement.Steps[0].Measurement.Outcome != "resolved" || d.calls.Load() != 1 {
				t.Fatalf("DNS primitive: %s", raw)
			}
		})
	}
}

// Accepted profiles execute all one/eight steps, including terminal negatives.
// requirement-ref: REQ-NET-05-0037
func TestNet05StepCountExecution(t *testing.T) {
	for _, count := range []int{1, 8} {
		h, d := net05AdmissionHarness(t, nil)
		readyComposite(h)
		c := h.initialized(t)
		var steps []string
		for i := 0; i < count; i++ {
			steps = append(steps, profileStep(fmt.Sprint("s", i), "TCP-01", `{"ip_address":"8.8.8.8","port":443}`, `{}`))
		}
		raw := tcpRaw(t, h, c, net05Task("all-steps", profileInput(steps...), `{}`, ""))
		v := net05Wire(t, raw)
		if v.Measurement.Outcome != "completed" || len(v.Measurement.Steps) != count || int(d.calls.Load()) != count {
			t.Fatal("subset executed")
		}
		c.Close()
	}
}
