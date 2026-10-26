package barometer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/config"
	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

const tcpPublic = "8.8.8.8"
const tcpAuthority = "8.8.8.8:443"

// These tests admit public literal targets with the production codec, target
// rule and registry. Only the platform transport is scripted: it maps the
// checked literal to loopback, never contacts that public destination, and
// fails if the runner changes its network or authority. No address exception
// or alternate codec is used to admit the loopback fixture.
type tcpTransport struct {
	calls     atomic.Int32
	authority string
	dial      func(context.Context) (measurement.Conn, error)
}

func (d *tcpTransport) Dialer(int) (measurement.Dialer, error) { return d, nil }
func (d *tcpTransport) DialContext(ctx context.Context, network, address string) (measurement.Conn, error) {
	d.calls.Add(1)
	expected := d.authority
	if expected == "" {
		expected = tcpAuthority
	}
	if network != "tcp" || address != expected {
		return nil, errors.New("synthetic: unexpected destination")
	}
	return d.dial(ctx)
}
func tcpProductionHarness(t *testing.T) (*harness, *tcpTransport) {
	t.Helper()
	h := runningHarness(t)
	h.adapters.registry = nil
	p := synPolicyInput(2, "reg-a", func(p *task.Permissions) { p.Tests = []task.TestVersion{tcp01}; p.AllowUnknownRisk = true })
	p.Targets[0].Test = tcp01
	p.Targets[0].Tuple = task.TargetTuple{Address: netip.MustParseAddr(tcpPublic), Port: 443, Scheme: "tcp"}
	var err error
	h.adapters.policy, err = task.NewLocalPolicy(p)
	if err != nil {
		t.Fatal(err)
	}
	h.adapters.run.paths = nil
	d := &tcpTransport{dial: func(context.Context) (measurement.Conn, error) { return nil, syscall.ECONNREFUSED }}
	h.adapters.run.directDialers = d
	h.adapters.approvals = config.Approvals{ConnectionCodes: []string{"outline_mobileproxy"}}
	return h, d
}
func tcpTask(id string) string {
	return strings.Replace(runTask(id, "", "", false), `"input":{}`, `"input":{"ip_address":"8.8.8.8","port":443}`, 1)
}
func tcpRaw(t *testing.T, h *harness, c *Client, data string) []byte {
	t.Helper()
	rh, err := c.StartRun(h.admit(t, c, data))
	if err != nil {
		t.Fatal(codeOf(err))
	}
	waitTerminal(t, c, rh)
	records := h.run.output.Records()
	if len(records) == 0 {
		t.Fatal("no validated result")
	}
	return records[len(records)-1]
}
func tcpResult(t *testing.T, raw []byte, state, reason, outcome string) {
	t.Helper()
	var r struct {
		Execution      struct{ State, Reason string }
		NetworkContext json.RawMessage `json:"network_context"`
		Measurement    *struct {
			Outcome      string
			Observations []struct {
				Type, Status, Failure string
				Duration              int `json:"duration_ms"`
				Details               map[string]any
			}
		}
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatal(err)
	}
	if r.Execution.State != state || r.Execution.Reason != reason {
		t.Fatalf("state: %s", raw)
	}
	if outcome == "" {
		if r.Measurement != nil {
			t.Fatalf("interrupted measurement: %s", raw)
		}
		if state == "not_run" && len(r.NetworkContext) != 0 {
			t.Fatal("not_run has context")
		}
		return
	}
	if r.Measurement == nil || r.Measurement.Outcome != outcome || len(r.Measurement.Observations) != 1 {
		t.Fatalf("measurement: %s", raw)
	}
	o := r.Measurement.Observations[0]
	if o.Type != "tcp_connect" || o.Duration < 0 || o.Duration%250 != 0 || len(o.Details) != 1 || o.Details["address_family"] != "ipv4" {
		t.Fatalf("observation: %s", raw)
	}
	if outcome == "connected" {
		if o.Status != "succeeded" || o.Failure != "" {
			t.Fatal(string(raw))
		}
	} else if o.Status != "failed" || o.Failure != outcome {
		t.Fatal(string(raw))
	}
	for _, canary := range []string{tcpPublic, `"port"`, `"local_address"`, `"payload"`, `CANARY-RAW-ERROR`, `91337`} {
		if strings.Contains(string(raw), canary) {
			t.Fatalf("retained %s", canary)
		}
	}
}

// M-TCP-01-ADMISSION: actual Client admission and atomic rejection.
// requirement-ref: REQ-TCP-01-0001
// requirement-ref: REQ-TCP-01-0010
// requirement-ref: REQ-TCP-01-0011
// requirement-ref: REQ-TCP-01-0012
// requirement-ref: REQ-TCP-01-0013
// requirement-ref: REQ-TCP-01-0014
// requirement-ref: REQ-JSON-0001
// requirement-ref: REQ-JSON-0002
// requirement-ref: REQ-JSON-0003
// requirement-ref: REQ-JSON-0004
// requirement-ref: REQ-JSON-0005
func TestMTcp01Admission(t *testing.T) {
	for _, ip := range []string{"8.8.8.8", "2606:4700:4700::1111"} {
		for _, port := range []int{1, 65535} {
			t.Run(fmt.Sprintf("valid/%s/%d", ip, port), func(t *testing.T) {
				h, d := tcpProductionHarness(t)
				p := synPolicyInput(2, "reg-a", func(p *task.Permissions) { p.Tests = []task.TestVersion{tcp01}; p.AllowUnknownRisk = true })
				p.Targets[0].Test = tcp01
				p.Targets[0].Tuple = task.TargetTuple{Address: netip.MustParseAddr(ip), Port: uint16(port), Scheme: "tcp"}
				h.adapters.policy, _ = task.NewLocalPolicy(p)
				c := h.initialized(t)
				defer c.Close()
				data := strings.Replace(tcpTask("valid"), `{"ip_address":"8.8.8.8","port":443}`, fmt.Sprintf(`{"ip_address":%q,"port":%d}`, ip, port), 1)
				if _, err := c.AdmitTask(data, h.embedded(t, data)); err != nil {
					t.Fatal(codeOf(err))
				}
				if d.calls.Load() != 0 {
					t.Fatal("admission contacted transport")
				}
			})
		}
	}
	inputs := []string{
		`{}`, `null`, `[]`, `{"ip_address":"8.8.8.8"}`, `{"ip_address":"8.8.8.8","port":0}`, `{"ip_address":"8.8.8.8","port":65536}`,
		`{"ip_address":"8.8.8.8","port":-1}`, `{"ip_address":"8.8.8.8","port":-0}`, `{"ip_address":"8.8.8.8","port":9007199254740993}`, `{"ip_address":"8.8.8.8","port":1.0}`, `{"ip_address":"8.8.8.8","port":1e2}`, `{"ip_address":"8.8.8.8","port":"443"}`, `{"ip_address":"8.8.8.8","port":null}`,
		`{"ip_address":"8.8.8.8","port":443,"port":443}`, `{"ip_address":"8.8.8.8","port":443,"\u0070ort":443}`, `{"ip_address":"8.8.8.8","port":443,"\u0078":1}`,
		`{"ip_address":["8.8.8.8"],"port":443}`, `{"ip_address":"8.8.8.8","port":[80,443]}`, `{"ip_address":"8.8.8.8","port":"1-100"}`,
	}
	for _, ip := range []string{"example.test", "${target}", "8.8.8.08", " 8.8.8.8", "0.0.0.0", "127.0.0.1", "10.0.0.1", "100.64.0.1", "169.254.1.1", "192.0.2.1", "198.18.0.1", "224.0.0.1", "240.0.0.1", "::", "::1", "fe80::1", "fc00::1", "ff02::1", "2001:db8::1", "::ffff:8.8.8.8", "2606:4700:4700:0:0:0:0:1111", "2606:4700:4700::ABCD", "fe80::1%en0"} {
		inputs = append(inputs, fmt.Sprintf(`{"ip_address":%q,"port":443}`, ip))
	}
	for i, input := range inputs {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			h, d := tcpProductionHarness(t)
			c := h.initialized(t)
			defer c.Close()
			data := strings.Replace(tcpTask("invalid"), `{"ip_address":"8.8.8.8","port":443}`, input, 1)
			handle, err := c.AdmitTask(data, h.embedded(t, data))
			if handle != "" || codeOf(err) != "task_schema_invalid" {
				t.Fatalf("%s: %s", input, codeOf(err))
			}
			if taskHandles(c) != 0 || h.replay.Commits() != 0 || len(c.rt.runs.entries) != 0 || d.calls.Load() != 0 || len(h.run.output.Records()) != 0 {
				t.Fatal("non-atomic rejection")
			}
		})
	}
	t.Run("unregistered tuple", func(t *testing.T) {
		h, d := tcpProductionHarness(t)
		c := h.initialized(t)
		defer c.Close()
		data := strings.Replace(tcpTask("wrong-port"), `"port":443`, `"port":444`, 1)
		_, err := c.AdmitTask(data, h.embedded(t, data))
		if codeOf(err) != "target_not_registered" || d.calls.Load() != 0 || h.replay.Commits() != 0 {
			t.Fatal(codeOf(err))
		}
	})
}

// M-TCP-01-OPTIONS; byte caps are checked on original encoded objects.
// requirement-ref: REQ-TCP-01-0015
// requirement-ref: REQ-TCP-01-0016
// requirement-ref: REQ-LIMIT-0004
func TestMTcp01Options(t *testing.T) {
	for _, options := range []string{`{}`, `null`, `[]`, `false`, `{"timeout":1}`, `{"family":"ipv4"}`, `{"\u0078":0}`, `{"x":0,"x":1}`} {
		h, d := tcpProductionHarness(t)
		c := h.initialized(t)
		data := strings.Replace(tcpTask("options"), `"options":{}`, `"options":`+options, 1)
		handle, err := c.AdmitTask(data, h.embedded(t, data))
		if options == `{}` {
			if err != nil || handle == "" {
				t.Fatal(codeOf(err))
			}
		} else if codeOf(err) != "task_schema_invalid" || handle != "" || h.replay.Commits() != 0 {
			t.Fatal(codeOf(err))
		}
		if d.calls.Load() != 0 {
			t.Fatal("options contacted transport")
		}
		c.Close()
	}
	for _, part := range []struct {
		field string
		limit int
	}{{"input", 49152}, {"options", 8192}} {
		for _, delta := range []int{-1, 0, 1} {
			h, _ := tcpProductionHarness(t)
			c := h.initialized(t)
			content := `{}`
			if part.field == "input" {
				content = `{"ip_address":"8.8.8.8","port":443}`
			}
			padded := content[:len(content)-1] + strings.Repeat(" ", part.limit+delta-len(content)) + "}"
			data := strings.Replace(tcpTask("size"), `"`+part.field+`":`+content, `"`+part.field+`":`+padded, 1)
			_, err := c.AdmitTask(data, h.embedded(t, data))
			if delta <= 0 && err != nil || delta > 0 && codeOf(err) != "task_limit_exceeded" {
				t.Fatalf("%s %d: %s", part.field, delta, codeOf(err))
			}
			c.Close()
		}
	}
}

// M-TCP-01-EXECUTION and OUTPUT: production registry, both paths, real sockets.
// requirement-ref: REQ-TCP-01-0006
// requirement-ref: REQ-TCP-01-0007
// requirement-ref: REQ-TCP-01-0020
// requirement-ref: REQ-TCP-01-0024
// requirement-ref: REQ-TCP-01-0027
func TestMTcp01ExecutionAndOutput(t *testing.T) {
	for _, proxy := range []bool{false, true} {
		t.Run(fmt.Sprint(proxy), func(t *testing.T) {
			h, d := tcpProductionHarness(t)
			srv, err := loopback.Listen()
			if err != nil {
				t.Fatal(err)
			}
			defer srv.Close()
			d.dial = func(ctx context.Context) (measurement.Conn, error) {
				var nd net.Dialer
				return nd.DialContext(ctx, "tcp", srv.Address())
			}
			cfg := `{}`
			data := tcpTask("connect")
			var p *loopback.Proxy
			if proxy {
				p, err = loopback.ListenProxy(loopback.ProxyConfig{Resolve: map[string]string{tcpAuthority: srv.Address()}})
				if err != nil {
					t.Fatal(err)
				}
				defer p.Close()
				cfg = proxyConfig(p.Port())
				data = strings.Replace(data, `"limits":{}`, `"limits":{},"network_path":"proxy"`, 1)
			}
			c := h.client(t)
			if err := c.Initialize(cfg); err != nil {
				t.Fatal(codeOf(err))
			}
			defer c.Close()
			tcpResult(t, tcpRaw(t, h, c, data), "completed", "", "connected")
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if err := srv.WaitAccepted(ctx, 1); err != nil {
				t.Fatal(err)
			}
			if err := srv.WaitOpen(ctx, 0); err != nil {
				t.Fatal(err)
			}
			sent, received := srv.Bytes()
			if srv.Accepted() != 1 || sent != 0 || received != 0 {
				t.Fatal("extra connections or application bytes")
			}
			if proxy {
				if err := p.WaitIdle(ctx); err != nil {
					t.Fatal(err)
				}
				counts := p.Counts()
				if counts.Accepted != 1 || counts.TunnelIn != 0 || counts.TunnelOut != 0 || len(counts.Requests) != 1 || !strings.HasPrefix(counts.Requests[0], "CONNECT "+tcpAuthority+" HTTP/1.1") || d.calls.Load() != 0 {
					t.Fatalf("proxy counts %+v", counts)
				}
			} else if d.calls.Load() != 1 {
				t.Fatal("not one direct dial")
			}
			if _, err := c.NextBundle(); err != nil {
				t.Fatal("result did not survive bundle validation", err)
			}
		})
	}
}

// M-TCP-01-NORMALIZATION: wrapped platform errors, no raw string/number retention.
// requirement-ref: REQ-TCP-01-0008
// requirement-ref: REQ-TCP-01-0009
// requirement-ref: REQ-TCP-01-0025
func TestMTcp01Normalization(t *testing.T) {
	for _, tc := range []struct {
		err     error
		outcome string
	}{
		{syscall.ECONNREFUSED, "connection_refused"}, {syscall.ECONNRESET, "connection_reset"}, {syscall.ETIMEDOUT, "timeout"}, {syscall.ENETUNREACH, "network_unreachable"}, {syscall.EHOSTUNREACH, "address_unreachable"}, {syscall.EACCES, "permission_denied"}, {syscall.EPERM, "permission_denied"}, {syscall.EINVAL, "socket_error"}, {errors.New("CANARY-RAW-ERROR-91337"), "inconclusive"},
	} {
		t.Run(tc.outcome+fmt.Sprint(tc.err), func(t *testing.T) {
			h, d := tcpProductionHarness(t)
			d.dial = func(context.Context) (measurement.Conn, error) {
				_ = h.run.clock.Advance(375 * time.Millisecond)
				return nil, fmt.Errorf("CANARY-RAW-ERROR: %w", tc.err)
			}
			c := h.initialized(t)
			defer c.Close()
			tcpResult(t, tcpRaw(t, h, c, tcpTask("normalization")), "completed", "", tc.outcome)
			if d.calls.Load() != 1 {
				t.Fatal("retry")
			}
		})
	}
}

// M-TCP-01-LIMITS: independent operation and attempt deadlines.
// requirement-ref: REQ-TCP-01-0022
func TestMTcp01Limits(t *testing.T) {
	for _, tc := range []struct {
		name, limits, state, reason, outcome string
		advance                              time.Duration
	}{
		{"operation", `"max_duration_ms":60000`, "completed", "", "timeout", 15 * time.Second},
		{"attempt", `"max_duration_ms":1000`, "incomplete", "attempt_deadline_exceeded", "", time.Second},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, d := tcpProductionHarness(t)
			d.dial = func(ctx context.Context) (measurement.Conn, error) {
				_ = h.run.clock.Advance(tc.advance)
				<-ctx.Done()
				return nil, ctx.Err()
			}
			c := h.initialized(t)
			defer c.Close()
			data := strings.Replace(tcpTask("limit"), `"limits":{}`, `"limits":{`+tc.limits+`}`, 1)
			tcpResult(t, tcpRaw(t, h, c, data), tc.state, tc.reason, tc.outcome)
			if d.calls.Load() != 1 {
				t.Fatal("unexpected dial count")
			}
		})
	}
}

// M-TCP-01-LIMITS connection rows through Client: the schema floor of
// max_connections is 1, and at it or above it TCP-01 issues exactly one
// connection. TestTcp01ConnectionMeter covers the refused row below it.
// requirement-ref: REQ-TCP-01-0022
// requirement-ref: REQ-LIMIT-0006
func TestMTcp01ConnectionBudget(t *testing.T) {
	for _, limit := range []int{1, 2} {
		t.Run(fmt.Sprint(limit), func(t *testing.T) {
			h, d := tcpProductionHarness(t)
			srv, err := loopback.Listen()
			if err != nil {
				t.Fatal(err)
			}
			defer srv.Close()
			d.dial = func(ctx context.Context) (measurement.Conn, error) {
				var nd net.Dialer
				return nd.DialContext(ctx, "tcp", srv.Address())
			}
			c := h.initialized(t)
			defer c.Close()
			data := strings.Replace(tcpTask("connections"), `"limits":{}`, fmt.Sprintf(`"limits":{"max_connections":%d}`, limit), 1)
			tcpResult(t, tcpRaw(t, h, c, data), "completed", "", "connected")
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if err := srv.WaitAccepted(ctx, 1); err != nil {
				t.Fatal(err)
			}
			if d.calls.Load() != 1 || srv.Accepted() != 1 {
				t.Fatal("not exactly one connection")
			}
		})
	}
}

// M-TCP-01-TERMINATION, applying the GO-SPEC-ISSUE-004 table.
// requirement-ref: REQ-TCP-01-0026
// requirement-ref: REQ-RESULT-0051
// requirement-ref: REQ-RESULT-0052
// requirement-ref: REQ-RESULT-0053
func TestMTcp01Termination(t *testing.T) {
	for _, reason := range []string{"network_changed", "cancelled_by_host"} {
		t.Run(reason, func(t *testing.T) {
			h, d := tcpProductionHarness(t)
			entered := make(chan struct{})
			d.dial = func(ctx context.Context) (measurement.Conn, error) {
				close(entered)
				<-ctx.Done()
				return nil, ctx.Err()
			}
			c := h.initialized(t)
			defer c.Close()
			rh, err := c.StartRun(h.admit(t, c, tcpTask("terminate")))
			if err != nil {
				t.Fatal(codeOf(err))
			}
			select {
			case <-entered:
			case <-time.After(time.Second):
				t.Fatal("no dial")
			}
			if reason == "network_changed" {
				a := defaultAccess
				a.Generation++
				h.run.network.Set(a)
			} else if err := c.CancelRun(rh); err != nil {
				t.Fatal(codeOf(err))
			}
			waitTerminal(t, c, rh)
			tcpResult(t, h.run.output.Records()[0], "incomplete", reason, "")
		})
	}
	t.Run("family check", func(t *testing.T) {
		h, d := tcpProductionHarness(t)
		a := defaultAccess
		a.IPv4 = platform.AvailabilityNo
		h.run.network.Set(a)
		c := h.initialized(t)
		defer c.Close()
		tcpResult(t, tcpRaw(t, h, c, tcpTask("family")), "incomplete", "constraints_unsatisfied", "")
		if d.calls.Load() != 0 {
			t.Fatal("unsupported family dialed")
		}
	})
	t.Run("destination gate", func(t *testing.T) {
		h, d := tcpProductionHarness(t)
		h.adapters.run.addressPolicy = denyTCPAddress{}
		c := h.initialized(t)
		defer c.Close()
		tcpResult(t, tcpRaw(t, h, c, tcpTask("gate")), "incomplete", "local_policy_rejected", "")
		if d.calls.Load() != 0 {
			t.Fatal("destination refusal contacted transport")
		}
	})
	t.Run("after outcome", func(t *testing.T) {
		h, d := tcpProductionHarness(t)
		d.dial = func(context.Context) (measurement.Conn, error) {
			return &tcpClosingConn{close: func() { a := defaultAccess; a.Generation++; h.run.network.Set(a) }}, nil
		}
		c := h.initialized(t)
		defer c.Close()
		tcpResult(t, tcpRaw(t, h, c, tcpTask("close")), "completed", "", "connected")
	})
}

type denyTCPAddress struct{}

func (denyTCPAddress) Allows(netip.Addr) bool { return false }

type tcpClosingConn struct{ close func() }

func (c *tcpClosingConn) Close() error {
	if c.close != nil {
		c.close()
	}
	return nil
}
func (*tcpClosingConn) Read([]byte) (int, error)    { panic("TCP-01 read application data") }
func (*tcpClosingConn) Write([]byte) (int, error)   { panic("TCP-01 wrote application data") }
func (*tcpClosingConn) SetDeadline(time.Time) error { return nil }

// GO-SPEC-ISSUE-007 at actual production-codec admission.
// requirement-ref: REQ-SAFETY-0031
// requirement-ref: REQ-SCHEDULE-0043
func TestTcp01AdmissionRiskAndTimezoneDomains(t *testing.T) {
	for _, level := range []string{"low", "medium", "high", "unknown"} {
		for _, factors := range []string{"", `,"risk_factors":[]`, `,"risk_factors":[{"type":"network_observability","level":"` + level + `"}]`} {
			h, _ := tcpProductionHarness(t)
			c := h.initialized(t)
			data := strings.Replace(tcpTask("risk"), `"risk_level":"low","requires_user_consent":false`, `"risk_level":"`+level+`","requires_user_consent":false`+factors, 1)
			_, err := c.AdmitTask(data, h.embedded(t, data))
			invalid := level != "low" && !strings.Contains(factors, "network_observability")
			if invalid && codeOf(err) != "task_schema_invalid" || !invalid && err != nil {
				t.Fatalf("%s %s: %s", level, factors, codeOf(err))
			}
			if invalid && h.replay.Commits() != 0 {
				t.Fatal("risk rejection persisted")
			}
			c.Close()
		}
	}
	for _, zone := range []string{"omitted", "local", "UTC", "Europe/Berlin", "", "Local", "LOCAL", "local "} {
		h, _ := tcpProductionHarness(t)
		c := h.initialized(t)
		z := `"timezone":` + fmt.Sprintf("%q", zone) + `,`
		if zone == "omitted" {
			z = ""
		}
		data := strings.Replace(tcpTask("zone"), `"schedule":{`, `"schedule":{"time_window":{`+z+`"ranges":[{"start":"08:00","end":"20:00"}]},`, 1)
		handle, err := c.AdmitTask(data, h.embedded(t, data))
		valid := zone == "omitted" || zone == "local"
		if valid && (err != nil || handle == "") || !valid && (codeOf(err) != "task_schema_invalid" || h.replay.Commits() != 0) {
			t.Fatalf("zone %q: %s", zone, codeOf(err))
		}
		c.Close()
	}
}

// M-TCP-01-VERSIONS: exact task and result tuples, strict receiver rejection.
// requirement-ref: REQ-TCP-01-0023
func TestMTcp01Versions(t *testing.T) {
	for _, version := range []string{`"0.1.1"`, `"0.2.0"`, `"1.0.0"`, `""`, `"v0.1.0"`, `"0.1"`, `0.1`, `null`} {
		h, d := tcpProductionHarness(t)
		c := h.initialized(t)
		data := strings.Replace(tcpTask("version"), `"test_version":"0.1.0"`, `"test_version":`+version, 1)
		handle, err := c.AdmitTask(data, h.embedded(t, data))
		if err == nil || handle != "" || h.replay.Commits() != 0 || d.calls.Load() != 0 {
			t.Fatal("unsupported version admitted", version)
		}
		c.Close()
	}
	h, _ := tcpProductionHarness(t)
	c := h.initialized(t)
	defer c.Close()
	raw := tcpRaw(t, h, c, tcpTask("versions"))
	cd, _ := observation.NewContextDecoder(result.Contract, observation.ContextApprovals{})
	path, _ := observation.NewPath(1, 0, "direct", "")
	decoder, err := result.New(result.Binding{TaskID: "versions", Output: result.Output{Test: model.Test{ID: "TCP-01", Version: "0.1.0", ResultVersion: "0.1.0"}}, MaxAttempts: 1, MaxDurationMS: 60000, Implementation: implementation, ContextDecoder: cd, Path: path, AccessGeneration: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decoder.Decode(raw); err != nil {
		t.Fatal("valid result rejected", err)
	}
	var original map[string]any
	if err := json.Unmarshal(raw, &original); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"schema_version", "version", "result_version"} {
		for _, v := range []any{"0.1.1", "0.2.0", "1.0.0", "0.1", "v0.1.0", "", 0.1, nil} {
			var value map[string]any
			_ = json.Unmarshal(raw, &value)
			if field == "schema_version" {
				value[field] = v
			} else {
				value["test"].(map[string]any)[field] = v
			}
			data, _ := json.Marshal(value)
			if _, err := decoder.Decode(data); err == nil {
				t.Fatalf("accepted %s %v", field, v)
			}
		}
	}
	for _, mutate := range []func(map[string]any){
		func(v map[string]any) { delete(v["test"].(map[string]any), "result_version") },
		func(v map[string]any) { v["test"].(map[string]any)["id"] = "DNS-01" },
		func(v map[string]any) { v["test"].(map[string]any)["id"] = "TLS-01" },
		func(v map[string]any) { v["test"].(map[string]any)["extra"] = true },
		func(v map[string]any) { v["measurement"].(map[string]any)["extra"] = true },
		func(v map[string]any) {
			v["measurement"].(map[string]any)["observations"].([]any)[0].(map[string]any)["type"] = "dns_query"
		},
		func(v map[string]any) {
			v["measurement"].(map[string]any)["observations"].([]any)[0].(map[string]any)["type"] = "tls_handshake"
		},
		func(v map[string]any) { v["measurement"].(map[string]any)["observations"] = []any{} },
		func(v map[string]any) {
			v["measurement"].(map[string]any)["observations"].([]any)[0].(map[string]any)["details"].(map[string]any)["ip_address"] = tcpPublic
		},
	} {
		var v map[string]any
		_ = json.Unmarshal(raw, &v)
		mutate(v)
		data, _ := json.Marshal(v)
		if _, err := decoder.Decode(data); err == nil {
			t.Fatal("invalid output accepted", string(data))
		}
	}
}

// The per-tuple gate holds an equivalent run through immediate close, while
// other task IDs retain distinct results. Cancellation while waiting starts
// no TCP procedure and emits not_run without a Network Context.
// requirement-ref: REQ-TCP-01-0022
func TestMTcp01TupleConcurrency(t *testing.T) {
	h, d := tcpProductionHarness(t)
	entered := make(chan struct{}, 2)
	release := make(chan struct{})
	d.dial = func(ctx context.Context) (measurement.Conn, error) {
		entered <- struct{}{}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-release:
			return nil, syscall.ECONNREFUSED
		}
	}
	c := h.initialized(t)
	defer c.Close()
	first, err := c.StartRun(h.admit(t, c, tcpTask("tuple-first")))
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("first dial missing")
	}
	second, err := c.StartRun(h.admit(t, c, tcpTask("tuple-second")))
	if err != nil {
		t.Fatal(err)
	}
	// CancelRun waits for all cleanup, so the second run has either been
	// cancelled before Runner entry or left its tuple wait without a dial.
	if err := c.CancelRun(second); err != nil {
		t.Fatal(err)
	}
	waitTerminal(t, c, second)
	if d.calls.Load() != 1 {
		t.Fatal("equivalent tuple contacted concurrently")
	}
	recs := h.run.output.Records()
	if len(recs) != 1 {
		t.Fatal("missing cancelled result")
	}
	tcpResult(t, recs[0], "not_run", "cancelled_by_host", "")
	close(release)
	waitTerminal(t, c, first)
	tcpResult(t, h.run.output.Records()[1], "completed", "", "connection_refused")
	// The tuple lease was released after the first result.
	third, err := c.StartRun(h.admit(t, c, tcpTask("tuple-third")))
	if err != nil {
		t.Fatal(err)
	}
	waitTerminal(t, c, third)
	if d.calls.Load() != 2 {
		t.Fatal("tuple lease was not released")
	}
}

// Production input/options fuzzing uses the real Client path with valid
// trusted delivery. Any rejected mutation is atomic and never contacts a socket.
// requirement-ref: REQ-JSON-0001
// requirement-ref: REQ-JSON-0002
// requirement-ref: REQ-TCP-01-0001
func FuzzTcp01Admission(f *testing.F) {
	for _, input := range []string{`{"ip_address":"8.8.8.8","port":443}`, `{"ip_address":"8.8.8.8","port":443,"\u0070ort":443}`, `{"ip_address":"::1","port":1}`, `null`, `[]`, `{`} {
		f.Add(input, `{}`)
	}
	f.Fuzz(func(t *testing.T, input, options string) {
		if len(input)+len(options) > 60000 {
			t.Skip()
		}
		h, d := tcpProductionHarness(t)
		h.adapters.deadline = defaultLifecycleDeadline
		c := h.initialized(t)
		defer c.Close()
		data := strings.Replace(tcpTask("fuzz-tcp"), `{"ip_address":"8.8.8.8","port":443}`, input, 1)
		data = strings.Replace(data, `"options":{}`, `"options":`+options, 1)
		ref := h.embedded(t, data)
		handle, err := c.AdmitTask(data, ref)
		if d.calls.Load() != 0 || len(c.rt.runs.entries) != 0 || len(h.run.output.Records()) != 0 {
			t.Fatal("admission executed")
		}
		if err != nil && (handle != "" || taskHandles(c) != 0 || h.replay.Commits() != 0) {
			t.Fatal("partial acceptance")
		}
	})
}

// Proxy failures are execution interruptions, never negative TCP outcomes.
// requirement-ref: REQ-TCP-01-0026
func TestMTcp01ProxyTermination(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mode   loopback.ProxyMode
		head   []byte
		reason string
	}{
		{"refused tunnel", loopback.ProxyReply, []byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"), "proxy_connect_failed"},
		{"outer deadline", loopback.ProxyHold, nil, "proxy_connect_failed"},
		{"reconfigured", loopback.ProxyHold, nil, "proxy_reconfigured"},
		{"cancelled", loopback.ProxyHold, nil, "cancelled_by_host"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, d := tcpProductionHarness(t)
			p, err := loopback.ListenProxy(loopback.ProxyConfig{Mode: tc.mode, Head: tc.head})
			if err != nil {
				t.Fatal(err)
			}
			defer p.Close()
			c := h.client(t)
			if err := c.Initialize(proxyConfig(p.Port())); err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			data := strings.Replace(tcpTask("proxy-termination"), `"limits":{}`, `"limits":{"max_duration_ms":1000},"network_path":"proxy"`, 1)
			rh, err := c.StartRun(h.admit(t, c, data))
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if err := p.WaitRequests(ctx, 1); err != nil {
				t.Fatal(err)
			}
			if tc.name == "outer deadline" {
				h.advance(time.Second)
			}
			if tc.name == "reconfigured" {
				if err := c.ClearProxy(); err != nil {
					t.Fatal(err)
				}
			}
			if tc.name == "cancelled" {
				if err := c.CancelRun(rh); err != nil {
					t.Fatal(err)
				}
			}
			waitTerminal(t, c, rh)
			tcpResult(t, h.run.output.Records()[0], "incomplete", tc.reason, "")
			if err := p.WaitIdle(ctx); err != nil {
				t.Fatal(err)
			}
			if d.calls.Load() != 0 || p.Counts().Accepted != 1 || p.Counts().TunnelIn != 0 {
				t.Fatal("fallback, retry or application traffic")
			}
		})
	}
	t.Run("unavailable", func(t *testing.T) {
		h, d := tcpProductionHarness(t)
		p, err := loopback.ListenProxy(loopback.ProxyConfig{})
		if err != nil {
			t.Fatal(err)
		}
		port := p.Port()
		p.Close()
		data := strings.Replace(tcpTask("unavailable"), `"limits":{}`, `"limits":{},"network_path":"proxy"`, 1)
		res, raw, c := runOnce(t, h, proxyConfig(port), data)
		defer c.Close()
		if res.Execution.State != "incomplete" || d.calls.Load() != 0 {
			t.Fatal(string(raw))
		}
		tcpResult(t, raw, "incomplete", "proxy_unavailable", "")
	})
}

// The schema is the same for embedded tasks and verified signed remote tasks.
// requirement-ref: REQ-JSON-0003
// requirement-ref: REQ-JSON-0004
// requirement-ref: REQ-JSON-0006
func TestTcp01StrictSchemaBothOrigins(t *testing.T) {
	for _, origin := range []string{"embedded", "remote"} {
		for _, tc := range []struct {
			input string
			valid bool
		}{
			{`{"ip_address":"8.8.8.8","port":443}`, true},
			{`{"\u0069p_address":"8.8.8.8","\u0070ort":443}`, true},
			{`{"ip_address":"8.8.8.8","port":443,"\u0070ort":443}`, false},
			{`{"ip_address":"8.8.8.8","port":443,"\u0078":true}`, false},
			{`{"ip_address":"8.8.8.8","port":1e2}`, false},
			{`{"ip_address":"8.8.8.8","port":-0}`, false},
		} {
			t.Run(origin+tc.input, func(t *testing.T) {
				h, d := tcpProductionHarness(t)
				c := h.initialized(t)
				defer c.Close()
				data := strings.Replace(tcpTask("strict"), `{"ip_address":"8.8.8.8","port":443}`, tc.input, 1)
				presented, ref := h.deliver(t, origin, data)
				handle, err := c.AdmitTask(presented, ref)
				if tc.valid && (err != nil || handle == "") || !tc.valid && (codeOf(err) != "task_schema_invalid" || handle != "" || h.replay.Commits() != 0 || taskHandles(c) != 0) {
					t.Fatal(codeOf(err))
				}
				if d.calls.Load() != 0 {
					t.Fatal("contact during admission")
				}
			})
		}
	}
}

// The family comes only from the canonical input; unknown/unavailable access
// capability cannot silently select another family or another address.
// requirement-ref: REQ-TCP-01-0017
// requirement-ref: REQ-TCP-01-0019
func TestMTcp01AddressFamilies(t *testing.T) {
	for _, ipv6 := range []bool{false, true} {
		for _, available := range []platform.Availability{platform.AvailabilityYes, platform.AvailabilityNo, platform.AvailabilityUnknown} {
			h, d := tcpProductionHarness(t)
			ip := tcpPublic
			family := "ipv4"
			if ipv6 {
				ip = "2606:4700:4700::1111"
				family = "ipv6"
			}
			p := synPolicyInput(2, "reg-a", func(p *task.Permissions) { p.Tests = []task.TestVersion{tcp01} })
			p.Targets[0].Test = tcp01
			p.Targets[0].Tuple = task.TargetTuple{Address: netip.MustParseAddr(ip), Port: 443, Scheme: "tcp"}
			var err error
			h.adapters.policy, err = task.NewLocalPolicy(p)
			if err != nil {
				t.Fatal(err)
			}
			a := defaultAccess
			if ipv6 {
				a.IPv6 = available
			} else {
				a.IPv4 = available
			}
			h.run.network.Set(a)
			d.authority = netip.AddrPortFrom(netip.MustParseAddr(ip), 443).String()
			c := h.initialized(t)
			data := strings.Replace(tcpTask("family-matrix"), `"ip_address":"8.8.8.8"`, fmt.Sprintf(`"ip_address":%q`, ip), 1)
			raw := tcpRaw(t, h, c, data)
			if available != platform.AvailabilityYes {
				tcpResult(t, raw, "incomplete", "constraints_unsatisfied", "")
				if d.calls.Load() != 0 {
					t.Fatal("family refusal contacted transport")
				}
			} else {
				var r struct {
					Execution   struct{ State string }
					Measurement struct {
						Outcome      string
						Observations []struct {
							Details struct {
								Family string `json:"address_family"`
							}
						}
					}
				}
				if err := json.Unmarshal(raw, &r); err != nil {
					t.Fatal(err)
				}
				if r.Execution.State != "completed" || r.Measurement.Outcome != "connection_refused" || len(r.Measurement.Observations) != 1 || r.Measurement.Observations[0].Details.Family != family || d.calls.Load() != 1 || strings.Contains(string(raw), ip) {
					t.Fatal(string(raw))
				}
			}
			c.Close()
		}
	}
}

// Two distinct admitted tasks for the same tuple must both finish, but their
// connects cannot overlap. The first is held while the second run is live.
// requirement-ref: REQ-TCP-01-0022
func TestMTcp01EquivalentRunsSerialize(t *testing.T) {
	h, d := tcpProductionHarness(t)
	entered := make(chan struct{}, 2)
	release := make(chan struct{})
	d.dial = func(ctx context.Context) (measurement.Conn, error) {
		entered <- struct{}{}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-release:
			return nil, syscall.ECONNREFUSED
		}
	}
	c := h.initialized(t)
	defer c.Close()
	first, err := c.StartRun(h.admit(t, c, tcpTask("serialize-first")))
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("first dial missing")
	}
	second, err := c.StartRun(h.admit(t, c, tcpTask("serialize-second")))
	if err != nil {
		t.Fatal(err)
	}
	// A second contact while the first remains held violates the tuple bound.
	select {
	case <-entered:
		close(release)
		t.Fatal("equivalent connections overlapped")
	case <-time.After(20 * time.Millisecond):
	}
	close(release)
	waitTerminal(t, c, first)
	waitTerminal(t, c, second)
	records := h.run.output.Records()
	if len(records) != 2 || d.calls.Load() != 2 {
		t.Fatal("serialized run lost or retried")
	}
	for _, raw := range records {
		tcpResult(t, raw, "completed", "", "connection_refused")
	}
}
