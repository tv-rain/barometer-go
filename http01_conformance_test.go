package barometer

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/config"
	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/measurement/http01"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

var httpTest = task.TestVersion{ID: "HTTP-01", Version: "0.1.0"}

const httpInput = `{"url":"https://8.8.8.8/check","expected_status":200}`

func httpTask(id, input, options string) string {
	s := strings.Replace(runTask(id, "", "", false), `"test_id":"TCP-01"`, `"test_id":"HTTP-01"`, 1)
	s = strings.Replace(s, `"input":{}`, `"input":`+input, 1)
	return strings.Replace(s, `"options":{}`, `"options":`+options, 1)
}
func httpParams(t *testing.T, input, options string) http01.Parameters {
	t.Helper()
	c := http01.Codec{}
	d := c.Definition()
	l := jsondecode.Limits{MaxBytes: 65536, MaxDepth: 32, MaxNodes: 2000}
	i, e := jsondecode.Decode([]byte(input), d.Input, l)
	if e != nil {
		t.Fatal(e)
	}
	o, e := jsondecode.Decode([]byte(options), d.Options, l)
	if e != nil {
		t.Fatal(e)
	}
	p, e := c.Normalize(i, o)
	if e != nil {
		t.Fatal(e)
	}
	return p.(http01.Parameters)
}
func httpPolicy(t *testing.T, h *harness, input string, redirects ...string) {
	t.Helper()
	p := synPolicyInput(2, "reg-a", func(p *task.Permissions) {
		p.Tests = []task.TestVersion{httpTest}
		p.AllowUnknownRisk = true
		for i := range redirects {
			p.RegistrationIDs = append(p.RegistrationIDs, fmt.Sprint("redirect-", i))
		}
	})
	p.Targets[0].Test = httpTest
	p.Targets[0].Tuple = httpParams(t, input, `{}`).Target()
	parent := 0
	for i, url := range redirects {
		r := p.Targets[0].Clone()
		r.ID = fmt.Sprint("redirect-", i)
		r.Role = task.RedirectTarget
		tuple, ok := http01.NormalizeURL(url)
		if !ok {
			t.Fatal("bad fixture")
		}
		tuple.ExpectedStatus = p.Targets[0].Tuple.ExpectedStatus
		tuple.BodyChecks = p.Targets[0].Tuple.BodyChecks
		tuple.HeaderChecks = p.Targets[0].Tuple.HeaderChecks
		r.Tuple = tuple
		r.RedirectIDs = nil
		p.Targets[parent].RedirectIDs = []string{r.ID}
		p.Targets = append(p.Targets, r)
		parent = len(p.Targets) - 1
	}
	var e error
	h.adapters.policy, e = task.NewLocalPolicy(p)
	if e != nil {
		t.Fatal(e)
	}
}

type httpTransport struct {
	calls atomic.Int32
	dial  func(context.Context, string, string) (measurement.Conn, error)
}

func (d *httpTransport) Dialer(int) (measurement.Dialer, error) { return d, nil }
func (d *httpTransport) DialContext(ctx context.Context, n, a string) (measurement.Conn, error) {
	d.calls.Add(1)
	if d.dial != nil {
		return d.dial(ctx, n, a)
	}
	return nil, syscall.ECONNREFUSED
}
func httpHarness(t *testing.T, input, raw string, stall bool) (*harness, *httpTransport, *loopback.HTTPServer) {
	t.Helper()
	ca, e := loopback.NewTestCA()
	if e != nil {
		t.Fatal(e)
	}
	cert, e := ca.Issue(httpParams(t, input, `{}`).Target().Hostname, false)
	if e != nil {
		t.Fatal(e)
	}
	srv, e := loopback.ListenHTTP(loopback.HTTPConfig{Certificate: cert, Raw: []byte(raw), Stall: stall})
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { _ = srv.Close() })
	h := runningHarness(t)
	h.adapters.registry = nil
	h.adapters.measurementRoots = ca.Pool()
	h.adapters.run.paths = nil
	h.adapters.approvals = config.Approvals{ConnectionCodes: []string{"outline_mobileproxy"}}
	httpPolicy(t, h, input)
	d := &httpTransport{dial: func(ctx context.Context, n, a string) (measurement.Conn, error) {
		if a != "8.8.8.8:443" {
			return nil, fmt.Errorf("CANARY-RAW-ERROR")
		}
		var nd net.Dialer
		return nd.DialContext(ctx, "tcp", srv.Address())
	}}
	h.adapters.run.directDialers = d
	return h, d, srv
}
func httpAdmissionClient(t *testing.T) (*harness, *Client, *httpTransport) {
	t.Helper()
	h := runningHarness(t)
	h.adapters.registry = nil
	h.adapters.run.paths = nil
	httpPolicy(t, h, httpInput)
	d := &httpTransport{}
	h.adapters.run.directDialers = d
	c := h.initialized(t)
	t.Cleanup(func() { _ = c.Close() })
	return h, c, d
}
func responseHTTP(body string) string {
	return fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Length: %d\r\nX-Fixture: header-canary\r\n\r\n%s", len(body), body)
}

type httpResult struct {
	Schema string `json:"schema_version"`
	Test   struct {
		ID, Version   string
		ResultVersion string `json:"result_version"`
	}
	Execution   struct{ State, Reason string }
	Measurement *struct {
		Outcome      string
		Observations []struct {
			Type, Status, Failure string
			Details               map[string]any
		}
	}
	NetworkContext json.RawMessage `json:"network_context"`
}

func checkHTTP(t *testing.T, raw []byte, state, reason, outcome string) httpResult {
	t.Helper()
	var r httpResult
	if e := json.Unmarshal(raw, &r); e != nil {
		t.Fatal(e)
	}
	if r.Execution.State != state || r.Execution.Reason != reason {
		t.Fatalf("execution: %s", raw)
	}
	if outcome == "" {
		if r.Measurement != nil {
			t.Fatalf("unexpected measurement: %s", raw)
		}
		if state == "not_run" && len(r.NetworkContext) > 0 {
			t.Fatal("not_run context")
		}
		return r
	}
	if r.Measurement == nil || r.Measurement.Outcome != outcome {
		t.Fatalf("outcome want %s: %s", outcome, raw)
	}
	for _, canary := range []string{"8.8.8.8", "/check", "header-canary", "body-canary", "X-Fixture", "x-fixture", "CANARY-RAW-ERROR", loopback.CertificateCanary, "https://", "127.0.0.1"} {
		if strings.Contains(string(raw), canary) {
			t.Fatalf("retained %s: %s", canary, raw)
		}
	}
	return r
}

// requirement-ref: REQ-HTTP-01-0010
// requirement-ref: REQ-HTTP-01-0011
// requirement-ref: REQ-HTTP-01-0013
// requirement-ref: REQ-HTTP-01-0022
func TestMHttp01Admission(t *testing.T) {
	h, d, _ := httpHarness(t, httpInput, responseHTTP(""), false)
	c := h.initialized(t)
	defer c.Close()
	invalid := []string{`{}`, `null`, `[]`, `{"url":null}`, `{"url":"http://8.8.8.8/check"}`, `{"url":"https://u:p@8.8.8.8/check"}`, `{"url":"https://8.8.8.8/check#x"}`, `{"url":"https://127.0.0.1/check"}`, `{"url":"https://8.8.8.8/check","expected_status":600}`, `{"url":"https://8.8.8.8/check","expected_status":"200"}`, `{"url":"https://8.8.8.8/check","other":0}`, `{"url":"https://8.8.8.8/check","url":"https://8.8.8.8/check"}`}
	for _, b := range []string{`{"type":"regex","value":"x"}`, `{"type":"sha256","value":"ABC"}`, `{"type":"contains","value":""}`, `{"type":"prefix","value":null}`} {
		invalid = append(invalid, `{"url":"https://8.8.8.8/check","body_checks":[`+b+`]}`)
	}
	for _, b := range []string{`{"name":"Set-Cookie","type":"present"}`, `{"name":"x","type":"present","value":""}`, `{"name":"x","type":"equals"}`, `{"name":"x","type":"equals","value":""}`, `{"name":"bad header","type":"present"}`} {
		invalid = append(invalid, `{"url":"https://8.8.8.8/check","header_checks":[`+b+`]}`)
	}
	for i, in := range invalid {
		h, c, _ := httpAdmissionClient(t)
		data := httpTask(fmt.Sprint("bad-", i), in, `{}`)
		handle, e := c.AdmitTask(data, h.embedded(t, data))
		if e == nil || handle != "" {
			t.Fatalf("accepted %s", in)
		}
	}
	data := httpTask("unregistered", strings.Replace(httpInput, "/check", "/other", 1), `{}`)
	if _, e := c.AdmitTask(data, h.embedded(t, data)); codeOf(e) != "target_not_registered" {
		t.Fatal(codeOf(e))
	}
	if d.calls.Load() != 0 || h.replay.Commits() != 0 {
		t.Fatal("rejection side effects")
	}
}

// requirement-ref: REQ-HTTP-01-0024
// requirement-ref: REQ-HTTP-01-0025
func TestMHttp01Options(t *testing.T) {
	h, d, _ := httpHarness(t, httpInput, responseHTTP(""), false)
	c := h.initialized(t)
	defer c.Close()
	for i, opt := range []string{`{"address_family":"unknown"}`, `{"address_family":null}`, `{"max_body_bytes":0}`, `{"max_body_bytes":65537}`, `{"max_body_bytes":"4"}`, `{"follow_redirects":true}`, `{"max_redirects":1}`, `{"follow_redirects":false,"max_redirects":2}`, `{"follow_redirects":true,"max_redirects":3}`, `{"method":"POST"}`, `{"headers":{}}`, `{"cookies":true}`, `{"insecure_skip_verify":true}`} {
		h, c, _ := httpAdmissionClient(t)
		data := httpTask(fmt.Sprint("bad-opt-", i), httpInput, opt)
		if _, e := c.AdmitTask(data, h.embedded(t, data)); e == nil {
			t.Fatal(opt)
		}
	}
	for i, opt := range []string{`{}`, `{"address_family":"ipv4"}`, `{"max_body_bytes":1}`, `{"max_body_bytes":65536}`, `{"follow_redirects":true,"max_redirects":1}`, `{"follow_redirects":true,"max_redirects":2}`} {
		data := httpTask(fmt.Sprint("good-opt-", i), httpInput, opt)
		if _, e := c.AdmitTask(data, h.embedded(t, data)); e != nil {
			t.Fatal(opt, codeOf(e))
		}
	}
	data := strings.Replace(httpTask("receive-bound", httpInput, `{"max_body_bytes":4096}`), `"limits":{}`, `"limits":{"max_bytes_received":4095}`, 1)
	if _, e := c.AdmitTask(data, h.embedded(t, data)); e == nil {
		t.Fatal("body exceeded receive bound")
	}
	if d.calls.Load() != 0 {
		t.Fatal("admission I/O")
	}
}

// requirement-ref: REQ-HTTP-01-0005
// requirement-ref: REQ-HTTP-01-0007
// requirement-ref: REQ-HTTP-01-0051
func TestMHttp01Execution(t *testing.T) {
	for _, proxy := range []bool{false, true} {
		t.Run(fmt.Sprint(proxy), func(t *testing.T) {
			h, d, srv := httpHarness(t, httpInput, responseHTTP("body-canary"), false)
			cfg := `{}`
			data := httpTask("run", httpInput, `{}`)
			if proxy {
				p, e := loopback.ListenProxy(loopback.ProxyConfig{Resolve: map[string]string{"8.8.8.8:443": srv.Address()}})
				if e != nil {
					t.Fatal(e)
				}
				defer p.Close()
				cfg = proxyConfig(p.Port())
				data = strings.Replace(data, `"limits":{}`, `"limits":{},"network_path":"proxy"`, 1)
			}
			c := h.client(t)
			if e := c.Initialize(cfg); e != nil {
				t.Fatal(codeOf(e))
			}
			defer c.Close()
			r := checkHTTP(t, tcpRaw(t, h, c, data), "completed", "", "expected_response")
			if len(r.Measurement.Observations) != 3 {
				t.Fatal(r.Measurement)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if e := srv.Wait(ctx, func(c loopback.HTTPCounts) bool { return c.Open == 0 && len(c.Requests) == 1 }); e != nil {
				t.Fatal(e)
			}
			req := srv.Counts().Requests[0]
			if req.Method != "GET" || len(req.Body) != 0 || req.Header.Get("Accept-Encoding") != "identity" || req.Header.Get("User-Agent") != "" || req.Header.Get("Cookie") != "" || req.Header.Get("Authorization") != "" {
				t.Fatal("request profile")
			}
			if proxy && d.calls.Load() != 0 {
				t.Fatal("direct fallback")
			}
		})
	}
}

// requirement-ref: REQ-HTTP-01-0033
// requirement-ref: REQ-HTTP-01-0034
// requirement-ref: REQ-HTTP-01-0035
// requirement-ref: REQ-HTTP-01-0039
func TestMHttp01Output(t *testing.T) {
	in := `{"url":"https://8.8.8.8/check","expected_status":200,"body_checks":[{"type":"contains","value":"body-canary"}],"header_checks":[{"name":"X-Fixture","type":"equals","value":"header-canary"}]}`
	h, _, _ := httpHarness(t, in, responseHTTP("body-canary"), false)
	c := h.initialized(t)
	defer c.Close()
	r := checkHTTP(t, tcpRaw(t, h, c, httpTask("output", in, `{}`)), "completed", "", "expected_response")
	if r.Schema != "0.1.0" || r.Test.ID != "HTTP-01" || r.Test.Version != "0.1.0" || r.Test.ResultVersion != "0.1.0" {
		t.Fatal(r.Test)
	}
	for i, typ := range []string{"tcp_connect", "tls_handshake", "http_exchange"} {
		if r.Measurement.Observations[i].Type != typ {
			t.Fatal("order")
		}
	}
	last := r.Measurement.Observations[2].Details
	if len(last["checks"].([]any)) != 3 {
		t.Fatal(last)
	}
}

// requirement-ref: REQ-HTTP-01-0036
// requirement-ref: REQ-HTTP-01-0037
// requirement-ref: REQ-HTTP-01-0052
func TestMHttp01Normalization(t *testing.T) {
	for _, tc := range []struct{ name, raw, checks, opt, want string }{
		{"status", "HTTP/1.1 503 Nope\r\nContent-Length: 0\r\n\r\n", "", `{}`, "unexpected_status"},
		{"header", responseHTTP("ok"), `,"header_checks":[{"name":"missing","type":"present"}]`, `{}`, "header_check_failed"},
		{"body", responseHTTP("no"), `,"body_checks":[{"type":"contains","value":"ok"}]`, `{}`, "body_check_failed"},
		{"encoding-before-status", "HTTP/1.1 503 Nope\r\nContent-Encoding: gzip\r\nContent-Length: 0\r\n\r\n", `,"body_checks":[{"type":"contains","value":"ok"}]`, `{}`, "unsupported_content_encoding"},
		{"read-before-status", "HTTP/1.1 503 Nope\r\nContent-Length: 4\r\n\r\nab", "", `{}`, "protocol_error"},
		{"body-limit", responseHTTP("abcdef"), `,"body_checks":[{"type":"suffix","value":"ok"}]`, `{"max_body_bytes":4}`, "body_limit_reached"},
		{"decided-truncated", responseHTTP("abcdef"), `,"body_checks":[{"type":"prefix","value":"abcd"}]`, `{"max_body_bytes":4}`, "expected_response"},
		{"status-before-limit", "HTTP/1.1 503 Nope\r\nContent-Length: 6\r\n\r\nabcdef", `,"body_checks":[{"type":"suffix","value":"ok"}]`, `{"max_body_bytes":4}`, "unexpected_status"},
		{"malformed", "garbage\r\n\r\n", "", `{}`, "protocol_error"},
		{"redirect-observed", "HTTP/1.1 302 Found\r\nLocation: https://1.1.1.1/check\r\nContent-Length: 0\r\n\r\n", "", `{}`, "redirect_observed"},
		{"redirect-rejected", "HTTP/1.1 302 Found\r\nLocation: http://1.1.1.1/check\r\nContent-Length: 0\r\n\r\n", "", `{"follow_redirects":true,"max_redirects":1}`, "redirect_rejected"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := strings.TrimSuffix(httpInput, "}") + tc.checks + "}"
			h, _, _ := httpHarness(t, in, tc.raw, false)
			c := h.initialized(t)
			defer c.Close()
			checkHTTP(t, tcpRaw(t, h, c, httpTask("normalize", in, tc.opt)), "completed", "", tc.want)
		})
	}
}

// requirement-ref: REQ-HTTP-01-0032
func TestMHttp01Limits(t *testing.T) {
	for _, size := range []int{0, 3, 4, 5, 4095, 4096, 4097, 65535, 65536, 65537} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			bound := 4
			if size > 5 {
				bound = 4096
			}
			if size > 4097 {
				bound = 65536
			}
			h, _, _ := httpHarness(t, httpInput, responseHTTP(strings.Repeat("a", size)), false)
			c := h.initialized(t)
			defer c.Close()
			r := checkHTTP(t, tcpRaw(t, h, c, httpTask("bound", httpInput, fmt.Sprintf(`{"max_body_bytes":%d}`, bound))), "completed", "", "expected_response")
			b := r.Measurement.Observations[2].Details["body"].(map[string]any)
			if int(b["bytes_processed"].(float64)) != min(size, bound) || b["truncated"] != (size > bound) {
				t.Fatal(b)
			}
		})
	}
}

// requirement-ref: REQ-HTTP-01-0038
// requirement-ref: REQ-RESULT-0051
// requirement-ref: REQ-RESULT-0052
func TestMHttp01Termination(t *testing.T) {
	for _, kind := range []string{"host", "deadline", "protocol"} {
		t.Run(kind, func(t *testing.T) {
			h, _, srv := httpHarness(t, httpInput, "", true)
			c := h.initialized(t)
			defer c.Close()
			data := httpTask("termination", httpInput, `{}`)
			rh, e := c.StartRun(h.admit(t, c, data))
			if e != nil {
				t.Fatal(e)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if e := srv.Wait(ctx, func(c loopback.HTTPCounts) bool { return len(c.Requests) == 1 }); e != nil {
				t.Fatal(e)
			}
			state, reason, outcome := "incomplete", "cancelled_by_host", ""
			switch kind {
			case "host":
				if e := c.CancelRun(rh); e != nil {
					t.Fatal(e)
				}
			case "deadline":
				h.advance(30 * time.Second)
				reason = "attempt_deadline_exceeded"
			case "protocol":
				h.advance(15 * time.Second)
				state, reason, outcome = "completed", "", "http_timeout"
			}
			waitTerminal(t, c, rh)
			rows := h.run.output.Records()
			if len(rows) != 1 {
				t.Fatal("missing result")
			}
			checkHTTP(t, rows[0], state, reason, outcome)
		})
	}
}

// requirement-ref: REQ-HTTP-01-0033
func TestMHttp01Versions(t *testing.T) {
	h, d, _ := httpHarness(t, httpInput, responseHTTP(""), false)
	c := h.initialized(t)
	defer c.Close()
	for i, v := range []string{"0.2.0", "1.0.0", "0.1", "v0.1.0", "0.1.0 ", ""} {
		data := strings.Replace(httpTask(fmt.Sprint("version-", i), httpInput, `{}`), `"test_version":"0.1.0"`, fmt.Sprintf(`"test_version":%q`, v), 1)
		if _, e := c.AdmitTask(data, h.embedded(t, data)); e == nil {
			t.Fatal(v)
		}
	}
	if d.calls.Load() != 0 {
		t.Fatal("unsupported version I/O")
	}
}

// The staged connection must not perform a second lookup or expose answers.
// requirement-ref: REQ-HTTP-01-0007
// requirement-ref: REQ-HTTP-01-0034
func TestMHttp01DNSStages(t *testing.T) {
	for _, kind := range []string{"success", "failure", "unsafe"} {
		t.Run(kind, func(t *testing.T) {
			in := strings.Replace(httpInput, "8.8.8.8", "service.example", 1)
			h, d, _ := httpHarness(t, in, responseHTTP(""), false)
			var calls atomic.Int32
			h.adapters.run.resolver = dnsSystemResolver{func(ctx context.Context, network, name string) ([]netip.Addr, error) {
				calls.Add(1)
				if network != "ip4" || name != "service.example" {
					t.Error("wrong resolver request")
				}
				if kind == "failure" {
					return nil, &net.DNSError{IsTimeout: true, Err: "CANARY-RAW-ERROR"}
				}
				if kind == "unsafe" {
					return []netip.Addr{netip.MustParseAddr("8.8.8.8"), netip.MustParseAddr("127.0.0.1")}, nil
				}
				return []netip.Addr{netip.MustParseAddr("8.8.8.8")}, nil
			}}
			c := h.initialized(t)
			defer c.Close()
			want, state, reason := "expected_response", "completed", ""
			if kind == "failure" {
				want = "dns_failure"
			}
			if kind == "unsafe" {
				want, state, reason = "", "incomplete", "local_policy_rejected"
			}
			raw := tcpRaw(t, h, c, httpTask("dns", in, `{}`))
			r := checkHTTP(t, raw, state, reason, want)
			if strings.Contains(string(raw), "service.example") {
				t.Fatal("retained hostname")
			}
			if calls.Load() != 1 {
				t.Fatal("second lookup")
			}
			if kind != "success" && d.calls.Load() != 0 {
				t.Fatal("contact after failed DNS/gate")
			}
			if kind == "success" {
				if len(r.Measurement.Observations) != 4 || r.Measurement.Observations[0].Type != "dns_query" {
					t.Fatal("DNS sequence")
				}
			}
		})
	}
}

// requirement-ref: REQ-HTTP-01-0036
func TestMHttp01PrecedingFailures(t *testing.T) {
	for _, kind := range []string{"tcp", "tls"} {
		t.Run(kind, func(t *testing.T) {
			h, d, _ := httpHarness(t, httpInput, responseHTTP(""), false)
			if kind == "tcp" {
				d.dial = nil
			} else {
				h.adapters.measurementRoots = nil
			}
			c := h.initialized(t)
			defer c.Close()
			r := checkHTTP(t, tcpRaw(t, h, c, httpTask("preceding", httpInput, `{}`)), "completed", "", kind+"_failure")
			n := 1
			if kind == "tls" {
				n = 2
			}
			if len(r.Measurement.Observations) != n {
				t.Fatal("invented stage")
			}
		})
	}
}

// requirement-ref: REQ-HTTP-01-0018
// requirement-ref: REQ-HTTP-01-0032
// requirement-ref: REQ-HTTP-01-0051
func TestMHttp01Redirects(t *testing.T) {
	for _, kind := range []string{"follow", "limit1", "limit2", "unregistered", "proxy-hostname"} {
		t.Run(kind, func(t *testing.T) {
			target := "https://8.8.8.8/next"
			if kind == "proxy-hostname" {
				target = "https://service.example/next"
			}
			raw := fmt.Sprintf("HTTP/1.1 302 Found\r\nLocation: %s\r\nContent-Length: 0\r\n\r\n", target)
			h, d, srv := httpHarness(t, httpInput, raw, false)
			httpPolicy(t, h, httpInput, target)
			bound := 1
			want := "redirect_limit_reached"
			if kind == "limit1" {
				httpPolicy(t, h, httpInput, target, target)
			}
			if kind == "limit2" {
				bound = 2
				httpPolicy(t, h, httpInput, target, target, target)
			}
			if kind == "unregistered" {
				httpPolicy(t, h, httpInput)
				want = "redirect_rejected"
			}
			if kind == "proxy-hostname" {
				want = "redirect_rejected"
			}
			if kind == "follow" {
				ca, e := loopback.NewTestCA()
				if e != nil {
					t.Fatal(e)
				}
				cert, e := ca.Issue("8.8.8.8", false)
				if e != nil {
					t.Fatal(e)
				}
				first, e := loopback.ListenHTTP(loopback.HTTPConfig{Certificate: cert, Raw: []byte(raw)})
				if e != nil {
					t.Fatal(e)
				}
				defer first.Close()
				last, e := loopback.ListenHTTP(loopback.HTTPConfig{Certificate: cert, Raw: []byte(responseHTTP(""))})
				if e != nil {
					t.Fatal(e)
				}
				defer last.Close()
				h.adapters.measurementRoots = ca.Pool()
				d.dial = func(ctx context.Context, _, a string) (measurement.Conn, error) {
					if a != "8.8.8.8:443" {
						t.Error("unchecked authority")
					}
					address := first.Address()
					if d.calls.Load() > 1 {
						address = last.Address()
					}
					var nd net.Dialer
					return nd.DialContext(ctx, "tcp", address)
				}
				want = "expected_response"
			}
			cfg := `{}`
			data := httpTask("redirect", httpInput, fmt.Sprintf(`{"follow_redirects":true,"max_redirects":%d}`, bound))
			if kind == "proxy-hostname" {
				p, e := loopback.ListenProxy(loopback.ProxyConfig{Resolve: map[string]string{"8.8.8.8:443": srv.Address()}})
				if e != nil {
					t.Fatal(e)
				}
				defer p.Close()
				cfg = proxyConfig(p.Port())
				data = strings.Replace(data, `"limits":{}`, `"limits":{},"network_path":"proxy"`, 1)
			}
			c := h.client(t)
			if e := c.Initialize(cfg); e != nil {
				t.Fatal(e)
			}
			defer c.Close()
			r := checkHTTP(t, tcpRaw(t, h, c, data), "completed", "", want)
			expected := 2
			if kind == "limit2" {
				expected = 3
			}
			if kind == "unregistered" || kind == "proxy-hostname" {
				expected = 1
			}
			if len(r.Measurement.Observations) != 3*expected {
				t.Fatal("redirect stage count")
			}
			if kind != "proxy-hostname" && int(d.calls.Load()) != expected {
				t.Fatal("extra hop")
			}
		})
	}
}

// requirement-ref: REQ-HTTP-01-0032
// requirement-ref: REQ-LIMIT-0034
func TestMHttp01OuterBudgets(t *testing.T) {
	for _, kind := range []string{"sent", "received", "connections"} {
		t.Run(kind, func(t *testing.T) {
			raw := responseHTTP(strings.Repeat("a", 6000))
			if kind == "connections" {
				raw = "HTTP/1.1 302 Found\r\nLocation: https://8.8.8.8/check\r\nContent-Length: 0\r\n\r\n"
			}
			h, _, srv := httpHarness(t, httpInput, raw, false)
			opts, limits := `{}`, `"max_bytes_sent":1`
			if kind == "received" {
				limits = `"max_bytes_received":4096`
			}
			if kind == "connections" {
				limits = `"max_connections":1`
				opts = `{"follow_redirects":true,"max_redirects":1}`
				httpPolicy(t, h, httpInput, "https://8.8.8.8/check")
			}
			c := h.initialized(t)
			defer c.Close()
			data := strings.Replace(httpTask("budget", httpInput, opts), `"limits":{}`, `"limits":{`+limits+`}`, 1)
			checkHTTP(t, tcpRaw(t, h, c, data), "incomplete", "resource_limit", "")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if e := srv.Wait(ctx, func(c loopback.HTTPCounts) bool { return c.Open == 0 }); e != nil {
				t.Fatal(e)
			}
		})
	}
}

// requirement-ref: REQ-HTTP-01-0032
func TestMHttp01EquivalentURLsSerialize(t *testing.T) {
	h, d, srv := httpHarness(t, httpInput, "", true)
	c := h.initialized(t)
	defer c.Close()
	first, e := c.StartRun(h.admit(t, c, httpTask("first", httpInput, `{}`)))
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if e := srv.Wait(ctx, func(c loopback.HTTPCounts) bool { return len(c.Requests) == 1 }); e != nil {
		t.Fatal(e)
	}
	second, e := c.StartRun(h.admit(t, c, httpTask("second", httpInput, `{}`)))
	if e != nil {
		t.Fatal(e)
	}
	if e := c.CancelRun(second); e != nil {
		t.Fatal(e)
	}
	waitTerminal(t, c, second)
	checkHTTP(t, h.run.output.Records()[0], "not_run", "cancelled_by_host", "")
	if d.calls.Load() != 1 {
		t.Fatal("concurrent equivalent URL")
	}
	if e := c.CancelRun(first); e != nil {
		t.Fatal(e)
	}
	waitTerminal(t, c, first)
	checkHTTP(t, h.run.output.Records()[1], "incomplete", "cancelled_by_host", "")
}

// requirement-ref: REQ-HTTP-01-0033
// requirement-ref: REQ-HTTP-01-0039
func TestMHttp01ResultVersionsAndClosedOutput(t *testing.T) {
	h, _, _ := httpHarness(t, httpInput, responseHTTP(""), false)
	c := h.initialized(t)
	defer c.Close()
	raw := tcpRaw(t, h, c, httpTask("versions", httpInput, `{}`))
	cd, _ := observation.NewContextDecoder(result.Contract, observation.ContextApprovals{})
	path, _ := observation.NewPath(1, 0, "direct", "")
	decoder, e := result.New(result.Binding{TaskID: "versions", Output: result.Output{Test: model.Test{ID: "HTTP-01", Version: "0.1.0", ResultVersion: "0.1.0"}, Limits: httpParams(t, httpInput, `{}`).Limits(), HTTPS: true}, MaxAttempts: 1, MaxDurationMS: 60000, Implementation: implementation, ContextDecoder: cd, Path: path, AccessGeneration: 1})
	if e != nil {
		t.Fatal(e)
	}
	if _, e := decoder.Decode(raw); e != nil {
		t.Fatal(e)
	}
	for _, field := range []string{"schema_version", "version", "result_version"} {
		for _, v := range []any{"0.2.0", "1.0.0", "0.1", "v0.1.0", "0.1.0 ", "", 0.1, nil} {
			var value map[string]any
			_ = json.Unmarshal(raw, &value)
			if field == "schema_version" {
				value[field] = v
			} else {
				value["test"].(map[string]any)[field] = v
			}
			data, _ := json.Marshal(value)
			if _, e := decoder.Decode(data); e == nil {
				t.Fatalf("accepted %s %v", field, v)
			}
		}
	}
	for _, field := range []string{"body_content", "expected_digest", "url", "header_value", "debug_log", "censorship"} {
		var value map[string]any
		_ = json.Unmarshal(raw, &value)
		last := value["measurement"].(map[string]any)["observations"].([]any)[2].(map[string]any)
		last["details"].(map[string]any)[field] = "CANARY-RAW-ERROR"
		data, _ := json.Marshal(value)
		if _, e := decoder.Decode(data); e == nil {
			t.Fatal("accepted unknown sensitive field", field)
		}
	}
	for _, id := range []string{"NET-01", "NET-05", "TLS-01", "http-01"} {
		var v map[string]any
		_ = json.Unmarshal(raw, &v)
		v["test"].(map[string]any)["id"] = id
		data, _ := json.Marshal(v)
		if _, e := decoder.Decode(data); e == nil {
			t.Fatal("cross-test result accepted")
		}
	}
}

// requirement-ref: REQ-HTTP-01-0032
func TestMHttp01HeaderBoundary(t *testing.T) {
	for _, size := range []int{16383, 16384, 16385} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			prefix := "HTTP/1.1 200 OK\r\nContent-Length: 0\r\nX-Padding: "
			raw := prefix + strings.Repeat("a", size-len(prefix)-4) + "\r\n\r\n"
			h, _, _ := httpHarness(t, httpInput, raw, false)
			c := h.initialized(t)
			defer c.Close()
			want := "expected_response"
			if size > 16384 {
				want = "protocol_error"
			}
			checkHTTP(t, tcpRaw(t, h, c, httpTask("header-bound", httpInput, `{}`)), "completed", "", want)
		})
	}
}

// requirement-ref: REQ-HTTP-01-0051
func TestMHttp01ProxyAdmission(t *testing.T) {
	for _, tc := range []struct{ input, opt string }{{strings.Replace(httpInput, "8.8.8.8", "service.example", 1), `{}`}, {httpInput, `{"address_family":"ipv4"}`}} {
		h, _, _ := httpHarness(t, tc.input, responseHTTP(""), false)
		c := h.initialized(t)
		data := strings.Replace(httpTask("proxy-shape", tc.input, tc.opt), `"limits":{}`, `"limits":{},"network_path":"proxy"`, 1)
		if _, e := c.AdmitTask(data, h.embedded(t, data)); codeOf(e) != "task_schema_invalid" {
			t.Fatal(codeOf(e))
		}
		_ = c.Close()
	}
}
