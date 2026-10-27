package barometer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/netip"
	"slices"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/config"
	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/measurement/net03"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/networkcontext"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit/capability"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

var net03Test = task.TestVersion{ID: "NET-03", Version: "0.1.0"}

const net03Input = `{"url":"https://8.8.8.8/check","expected_status":200}`

// net03Digest is the SHA-256 of the fixture body "ok". It is a canary: no
// result may carry it.
var net03Digest = func() string { s := sha256.Sum256([]byte("ok")); return hex.EncodeToString(s[:]) }()

func net03Task(id, input, options string) string {
	s := strings.Replace(runTask(id, "", "", false), `"test_id":"TCP-01"`, `"test_id":"NET-03"`, 1)
	s = strings.Replace(s, `"input":{}`, `"input":`+input, 1)
	return strings.Replace(s, `"options":{}`, `"options":`+options, 1)
}
func net03Params(t *testing.T, input, options string) net03.Parameters {
	t.Helper()
	c := net03.Codec{}
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
	return p.(net03.Parameters)
}

// net03Policy registers exactly one validation profile: the URL, expected
// status and optional digest of input. No redirect target is registered.
func net03Policy(t *testing.T, h *harness, input string) {
	t.Helper()
	p := synPolicyInput(2, "reg-a", func(p *task.Permissions) { p.Tests = []task.TestVersion{net03Test} })
	p.Targets[0].Test = net03Test
	p.Targets[0].Tuple = net03Params(t, input, `{}`).Target()
	var e error
	h.adapters.policy, e = task.NewLocalPolicy(p)
	if e != nil {
		t.Fatal(e)
	}
}

// net03Access is the default harness reading with the platform validation
// and captive-portal flags the case needs.
func net03Access(platformState string) platform.AccessState {
	a := defaultAccess
	switch platformState {
	case "validated":
		a.Validated = platform.AvailabilityYes
	case "not_validated":
		a.Validated = platform.AvailabilityNo
	case "captive_portal":
		a.Validated, a.CaptivePortal = platform.AvailabilityNo, platform.AvailabilityYes
	}
	return a
}

// net03Harness runs NET-03 with the production codec, rules, registry, runner
// and direct path. The fakes are the platform network port, the recording
// transport and a loopback server. The transport reaches the server only for
// the registered validation address; any other destination, such as a
// redirect canary, is counted in other and refused with a raw-error canary.
func net03Harness(t *testing.T, input, raw string, platformState string, plain, stall bool) (*harness, *httpTransport, *loopback.HTTPServer, *net01Network, *atomic.Int32) {
	t.Helper()
	ca, e := loopback.NewTestCA()
	if e != nil {
		t.Fatal(e)
	}
	target := net03Params(t, input, `{}`).Target()
	cert, e := ca.Issue(target.Hostname, false)
	if e != nil {
		t.Fatal(e)
	}
	srv, e := loopback.ListenHTTP(loopback.HTTPConfig{Certificate: cert, Raw: []byte(raw), Stall: stall, Plain: plain})
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { _ = srv.Close() })
	h := runningHarness(t)
	h.adapters.registry = nil
	h.adapters.measurementRoots = ca.Pool()
	h.adapters.approvals = config.Approvals{ConnectionCodes: []string{"outline_mobileproxy"}}
	net03Policy(t, h, input)
	n := &net01Network{Network: h.run.network}
	n.Set(net03Access(platformState))
	h.adapters.run.paths, h.adapters.run.network = nil, n
	other := &atomic.Int32{}
	d := &httpTransport{dial: func(ctx context.Context, _, a string) (measurement.Conn, error) {
		if a != fmt.Sprintf("8.8.8.8:%d", target.Port) {
			other.Add(1)
			return nil, fmt.Errorf("CANARY-RAW-ERROR")
		}
		var nd net.Dialer
		return nd.DialContext(ctx, "tcp", srv.Address())
	}}
	h.adapters.run.directDialers = d
	return h, d, srv, n, other
}

func net03AdmissionClient(t *testing.T) (*harness, *Client, *httpTransport) {
	t.Helper()
	h, d, _, _, _ := net03Harness(t, net03Input, "", "unknown", false, false)
	c := h.initialized(t)
	t.Cleanup(func() { _ = c.Close() })
	return h, c, d
}

// assertNet03Rejected admits one NET-03 task in a fresh client and requires
// the exact admission class, no task or run handle, no replay commit and no
// contact on the recording transport.
func assertNet03Rejected(t *testing.T, data, want string) {
	t.Helper()
	h, c, d := net03AdmissionClient(t)
	handle, e := c.AdmitTask(data, h.embedded(t, data))
	if handle != "" || codeOf(e) != want {
		t.Fatalf("admission = %q, %s; want %s", handle, codeOf(e), want)
	}
	if taskHandles(c) != 0 || h.replay.Commits() != 0 || d.calls.Load() != 0 || len(c.rt.runs.entries) != 0 {
		t.Fatal("rejection published state or contacted transport")
	}
}

func net03Response(status int, body string) string {
	return fmt.Sprintf("HTTP/1.1 %d Status\r\nContent-Length: %d\r\nX-Fixture: header-canary\r\nSet-Cookie: session=cookie-canary\r\n\r\n%s", status, len(body), body)
}

// net03Canaries must never appear in a NET-03 result: the validation URL and
// address, redirect locations, body, digest, header names and values, cookie,
// raw errors and certificate details (REQ-NET-03-0030).
var net03Canaries = []string{"8.8.8.8", "/check", "service.example", "portal-canary", "body-canary", "header-canary", "cookie-canary", "X-Fixture", "x-fixture", "Set-Cookie", "CANARY-RAW-ERROR", loopback.CertificateCanary, "https://", "http://", "127.0.0.1", net03Digest}

// net03Result checks a NET-03 result and returns its measurement. An outcome
// has exactly outcome, observations, platform_validation and, when a valid
// response was received, response_class; an interrupted attempt has no
// measurement at all.
func net03Result(t *testing.T, raw []byte, state, reason, outcome string) map[string]any {
	t.Helper()
	var r map[string]any
	if e := json.Unmarshal(raw, &r); e != nil {
		t.Fatal(e)
	}
	for _, c := range append(slices.Clone(net03Canaries), capability.Canaries...) {
		if strings.Contains(string(raw), c) {
			t.Fatalf("retained %q: %s", c, raw)
		}
	}
	e := r["execution"].(map[string]any)
	if e["state"] != state || (reason == "") != (e["reason"] == nil) || reason != "" && e["reason"] != reason {
		t.Fatalf("execution want %s/%s: %s", state, reason, raw)
	}
	m, ok := r["measurement"].(map[string]any)
	if outcome == "" {
		if ok {
			t.Fatalf("interrupted NET-03 has a measurement: %s", raw)
		}
		if _, has := r["network_context"]; state == "not_run" && has {
			t.Fatal("not_run context")
		}
		return nil
	}
	if !ok || m["outcome"] != outcome {
		t.Fatalf("outcome want %s: %s", outcome, raw)
	}
	for k := range m {
		if !slices.Contains([]string{"outcome", "observations", "platform_validation", "response_class"}, k) {
			t.Fatalf("unregistered measurement field %s", k)
		}
	}
	if _, ok := m["platform_validation"].(string); !ok {
		t.Fatalf("platform_validation missing: %s", raw)
	}
	return m
}

func net03Observations(m map[string]any) []map[string]any {
	var out []map[string]any
	for _, o := range m["observations"].([]any) {
		out = append(out, o.(map[string]any))
	}
	return out
}
func net03Types(m map[string]any) string {
	var out []string
	for _, o := range net03Observations(m) {
		out = append(out, o["type"].(string))
	}
	return strings.Join(out, ",")
}
func net03Last(m map[string]any) map[string]any {
	obs := net03Observations(m)
	return obs[len(obs)-1]
}

// requirement-ref: REQ-NET-03-0004
// requirement-ref: REQ-NET-03-0009
// requirement-ref: REQ-NET-03-0010
// requirement-ref: REQ-NET-03-0011
// requirement-ref: REQ-NET-03-0013
// requirement-ref: REQ-NET-03-0014
// requirement-ref: REQ-NET-03-0043
func TestMNet03Admission(t *testing.T) {
	invalid := []string{`{}`, `null`, `[]`, `{"url":null,"expected_status":200}`, `{"expected_status":200}`, `{"url":"https://8.8.8.8/check"}`,
		`{"url":"ftp://8.8.8.8/check","expected_status":200}`, `{"url":"https://u:p@8.8.8.8/check","expected_status":200}`, `{"url":"https://8.8.8.8/check#x","expected_status":200}`,
		`{"url":"https://127.0.0.1/check","expected_status":200}`, `{"url":"http://10.0.0.1/check","expected_status":200}`, `{"url":"https://169.254.1.1/check","expected_status":200}`, `{"url":"https://[::1]/check","expected_status":200}`,
		`{"url":"https://8.8.8.8/check","expected_status":"200"}`, `{"url":"https://8.8.8.8/check","expected_status":199}`, `{"url":"https://8.8.8.8/check","expected_status":600}`, `{"url":"https://8.8.8.8/check","expected_status":200.5}`,
		`{"url":"https://8.8.8.8/check","expected_status":200,"expected_body_sha256":"ABC"}`, `{"url":"https://8.8.8.8/check","expected_status":200,"expected_body_sha256":"` + strings.ToUpper(net03Digest) + `"}`,
		`{"url":"https://8.8.8.8/check","expected_status":200,"expected_body_sha256":"` + net03Digest[:63] + `"}`, `{"url":"https://8.8.8.8/check","expected_status":200,"expected_body_sha256":null}`,
		`{"url":"https://8.8.8.8/check","expected_status":200,"other":0}`, `{"url":"https://8.8.8.8/check","url":"https://8.8.8.8/check","expected_status":200}`}
	// A redirect or informational status could never establish validation.
	for _, s := range []int{100, 101, 301, 302, 303, 307, 308} {
		invalid = append(invalid, fmt.Sprintf(`{"url":"https://8.8.8.8/check","expected_status":%d}`, s))
	}
	for i, in := range invalid {
		assertNet03Rejected(t, net03Task(fmt.Sprint("bad-", i), in, `{}`), "task_schema_invalid")
	}
	long := `{"url":"https://8.8.8.8/` + strings.Repeat("a", 2049-len("https://8.8.8.8/")) + `","expected_status":200}`
	assertNet03Rejected(t, net03Task("long", long, `{}`), "task_limit_exceeded")
	// The profile is the complete tuple: a different URL, status or digest is
	// not the registered validation target (REQ-NET-03-0010).
	exact := `{"url":"https://8.8.8.8/` + strings.Repeat("a", 2048-len("https://8.8.8.8/")) + `","expected_status":200}`
	for i, in := range []string{strings.Replace(net03Input, "/check", "/other", 1), strings.Replace(net03Input, "200", "204", 1), strings.TrimSuffix(net03Input, "}") + `,"expected_body_sha256":"` + net03Digest + `"}`, strings.Replace(net03Input, "https", "http", 1), exact} {
		assertNet03Rejected(t, net03Task(fmt.Sprint("unregistered-", i), in, `{}`), "target_not_registered")
	}
	for _, in := range []string{net03Input, `{"url":"http://8.8.8.8/check","expected_status":204}`, `{"url":"https://8.8.8.8/check","expected_status":200,"expected_body_sha256":"` + net03Digest + `"}`} {
		h, d, _, _, _ := net03Harness(t, in, "", "unknown", false, false)
		c := h.initialized(t)
		data := net03Task("good", in, `{}`)
		if _, e := c.AdmitTask(data, h.embedded(t, data)); e != nil {
			t.Fatal(in, codeOf(e))
		}
		if d.calls.Load() != 0 {
			t.Fatal("admission I/O")
		}
		_ = c.Close()
	}
}

// requirement-ref: REQ-NET-03-0015
// requirement-ref: REQ-NET-03-0016
// requirement-ref: REQ-NET-03-0040
func TestMNet03Options(t *testing.T) {
	for i, opt := range []string{`{"address_family":"unknown"}`, `{"address_family":null}`, `{"address_family":"ipv6"}`, `{"follow_redirects":true}`, `{"max_redirects":1}`, `{"headers":{}}`, `{"cookies":true}`, `{"credentials":"x"}`, `{"method":"POST"}`, `{"body":""}`, `{"max_body_bytes":4096}`} {
		assertNet03Rejected(t, net03Task(fmt.Sprint("bad-opt-", i), net03Input, opt), "task_schema_invalid")
	}
	// NET-03 has no proxy path semantics (REQ-PROXY-0044).
	assertNet03Rejected(t, strings.Replace(net03Task("proxy", net03Input, `{}`), `"limits":{}`, `"limits":{},"network_path":"proxy"`, 1), "task_schema_invalid")
	host := strings.Replace(net03Input, "8.8.8.8", "service.example", 1)
	for _, tc := range []struct{ input, opt string }{{net03Input, `{}`}, {net03Input, `{"address_family":"auto"}`}, {net03Input, `{"address_family":"ipv4"}`}, {host, `{"address_family":"ipv6"}`}} {
		h, d, _, _, _ := net03Harness(t, tc.input, "", "unknown", false, false)
		c := h.initialized(t)
		data := net03Task("good-opt", tc.input, tc.opt)
		if _, e := c.AdmitTask(data, h.embedded(t, data)); e != nil {
			t.Fatal(tc.opt, codeOf(e))
		}
		if d.calls.Load() != 0 {
			t.Fatal("admission I/O")
		}
		_ = c.Close()
	}
	// The registered prefix must fit the attempt's receive budget; a task
	// cannot lower it below the bound instead (no clamping).
	h, c, _ := net03AdmissionClient(t)
	data := strings.Replace(net03Task("receive-bound", net03Input, `{}`), `"limits":{}`, `"limits":{"max_bytes_received":1023}`, 1)
	if _, e := c.AdmitTask(data, h.embedded(t, data)); e == nil {
		t.Fatal("prefix exceeded receive bound")
	}
}

// requirement-ref: REQ-NET-03-0005
// requirement-ref: REQ-NET-03-0006
// requirement-ref: REQ-NET-03-0012
// requirement-ref: REQ-NET-03-0028
// requirement-ref: REQ-NET-03-0038
func TestMNet03Execution(t *testing.T) {
	for _, tc := range []struct {
		name, input, types string
		plain              bool
	}{
		{"https", net03Input, "tcp_connect,tls_handshake,http_exchange", false},
		{"http", `{"url":"http://8.8.8.8/check","expected_status":204}`, "tcp_connect,http_exchange", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			status := 200
			if tc.plain {
				status = 204
			}
			h, d, srv, n, other := net03Harness(t, tc.input, net03Response(status, ""), "validated", tc.plain, false)
			c := h.initialized(t)
			defer c.Close()
			m := net03Result(t, tcpRaw(t, h, c, net03Task("run", tc.input, `{}`)), "completed", "", "validated")
			if net03Types(m) != tc.types || m["platform_validation"] != "validated" || m["response_class"] != "expected" {
				t.Fatal(m)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if e := srv.Wait(ctx, func(c loopback.HTTPCounts) bool { return c.Open == 0 && len(c.Requests) == 1 }); e != nil {
				t.Fatal(e)
			}
			req := srv.Counts().Requests[0]
			if req.Method != "GET" || req.Target != "/check" || len(req.Body) != 0 || req.Header.Get("User-Agent") != "" || req.Header.Get("Cookie") != "" || req.Header.Get("Authorization") != "" || req.Header.Get("If-None-Match") != "" || req.Header.Get("If-Modified-Since") != "" {
				t.Fatal("request profile")
			}
			// Reservation check, the attempt's capture and NET-03's own
			// platform read; exactly one connection and no other contact.
			if n.Reads() != 3 || d.calls.Load() != 1 || other.Load() != 0 {
				t.Fatalf("reads %d dials %d other %d", n.Reads(), d.calls.Load(), other.Load())
			}
		})
	}
	// A redirect target is classified and never contacted, whatever its class.
	for _, location := range []string{"https://1.1.1.1/portal-canary", "https://10.0.0.1/portal-canary", "http://1.1.1.1/portal-canary", "https://service.example/portal-canary?q=portal-canary"} {
		t.Run("redirect "+location, func(t *testing.T) {
			raw := "HTTP/1.1 302 Found\r\nLocation: " + location + "\r\nContent-Length: 0\r\n\r\n"
			h, d, srv, _, other := net03Harness(t, net03Input, raw, "validated", false, false)
			c := h.initialized(t)
			defer c.Close()
			m := net03Result(t, tcpRaw(t, h, c, net03Task("redirect", net03Input, `{}`)), "completed", "", "captive_portal_suspected")
			if m["response_class"] != "redirect" || d.calls.Load() != 1 || other.Load() != 0 || len(srv.Counts().Requests) != 1 {
				t.Fatal("redirect target contacted or misclassified")
			}
		})
	}
}

// Every response class with each platform state, and each preceding-stage
// failure with each platform state, select the REQ-NET-03-0042 outcome. The
// platform value is retained unchanged and response_class is present exactly
// when a valid response was received.
// requirement-ref: REQ-NET-03-0008
// requirement-ref: REQ-NET-03-0023
// requirement-ref: REQ-NET-03-0025
// requirement-ref: REQ-NET-03-0027
// requirement-ref: REQ-NET-03-0041
// requirement-ref: REQ-NET-03-0042
func TestMNet03Normalization(t *testing.T) {
	digest := strings.TrimSuffix(net03Input, "}") + `,"expected_body_sha256":"` + net03Digest + `"}`
	host := strings.Replace(net03Input, "8.8.8.8", "service.example", 1)
	cases := []struct {
		name, input, raw, class, stage string
		types                          string
	}{
		{"expected", net03Input, net03Response(200, "body-canary"), "expected", "", "tcp_connect,tls_handshake,http_exchange"},
		{"expected-digest", digest, net03Response(200, "ok"), "expected", "", "tcp_connect,tls_handshake,http_exchange"},
		{"redirect", net03Input, "HTTP/1.1 307 Temporary\r\nLocation: https://1.1.1.1/portal-canary\r\nContent-Length: 0\r\n\r\n", "redirect", "", "tcp_connect,tls_handshake,http_exchange"},
		{"content-mismatch", digest, net03Response(200, "body-canary"), "content_mismatch", "", "tcp_connect,tls_handshake,http_exchange"},
		{"status-2xx", net03Input, net03Response(204, ""), "unexpected", "", "tcp_connect,tls_handshake,http_exchange"},
		{"status-5xx", digest, net03Response(503, "ok"), "unexpected", "", "tcp_connect,tls_handshake,http_exchange"},
		{"encoding", digest, "HTTP/1.1 200 OK\r\nContent-Encoding: gzip\r\nContent-Length: 2\r\n\r\nok", "unexpected", "", "tcp_connect,tls_handshake,http_exchange"},
		{"premature-body", net03Input, "HTTP/1.1 200 OK\r\nContent-Length: 4\r\n\r\nok", "unexpected", "", "tcp_connect,tls_handshake,http_exchange"},
		{"no-response", net03Input, "garbage\r\n\r\n", "", "", "tcp_connect,tls_handshake,http_exchange"},
		{"tcp", net03Input, "", "", "tcp", "tcp_connect"},
		{"tls", net03Input, "", "", "tls", "tcp_connect,tls_handshake"},
		{"dns", host, "", "", "dns", "dns_query"},
	}
	for _, tc := range cases {
		for _, platformState := range []string{"validated", "not_validated", "captive_portal", "unknown"} {
			t.Run(tc.name+"/"+platformState, func(t *testing.T) {
				h, d, _, _, _ := net03Harness(t, tc.input, tc.raw, platformState, false, false)
				switch tc.stage {
				case "tcp":
					d.dial = func(context.Context, string, string) (measurement.Conn, error) { return nil, syscall.ECONNREFUSED }
				case "tls":
					h.adapters.measurementRoots = nil
				case "dns":
					h.adapters.run.resolver = dnsSystemResolver{func(context.Context, string, string) ([]netip.Addr, error) {
						return nil, &net.DNSError{IsNotFound: true, Err: "CANARY-RAW-ERROR"}
					}}
				}
				want := net03.Outcome(tc.class, platformState, map[string]string{"tcp": "tcp_failure", "tls": "tls_failure", "dns": "dns_failure"}[tc.stage])
				table := map[string]string{"expected": "validated", "redirect": "captive_portal_suspected", "content_mismatch": "captive_portal_suspected", "unexpected": "unexpected_response"}
				if tc.class != "" && want != table[tc.class] || tc.class == "" && platformState == "captive_portal" && want != "captive_portal_suspected" || tc.class == "" && platformState == "not_validated" && want != "not_validated" {
					t.Fatalf("selection %s", want)
				}
				c := h.initialized(t)
				defer c.Close()
				m := net03Result(t, tcpRaw(t, h, c, net03Task("normalize", tc.input, `{}`)), "completed", "", want)
				class, has := m["response_class"]
				if m["platform_validation"] != platformState || has != (tc.class != "") || has && class != tc.class || net03Types(m) != tc.types {
					t.Fatalf("evidence: %v", m)
				}
			})
		}
	}
	// Without a response class or platform state, the stage failure itself.
	for stage, want := range map[string]string{"tcp_failure": "connection_failure", "dns_failure": "dns_failure", "tls_failure": "tls_failure", "timeout": "inconclusive", "protocol_error": "inconclusive"} {
		if got := net03.Outcome("", "validated", stage); got != want {
			t.Fatalf("%s: %s", stage, got)
		}
	}
}

// requirement-ref: REQ-NET-03-0021
// requirement-ref: REQ-LIMIT-0016
// requirement-ref: REQ-LIMIT-0034
// requirement-ref: REQ-NET-03-0041
func TestMNet03Limits(t *testing.T) {
	// The 1,024-byte prefix: a digest decidable within it compares; one that
	// is not is the over-bound unexpected class. Without a digest, truncation
	// alone does not change the class.
	for _, size := range []int{0, 1023, 1024, 1025, 4096} {
		for _, withDigest := range []bool{false, true} {
			t.Run(fmt.Sprintf("%d/%t", size, withDigest), func(t *testing.T) {
				body := strings.Repeat("a", size)
				in := net03Input
				if withDigest {
					s := sha256.Sum256([]byte(body))
					in = strings.TrimSuffix(net03Input, "}") + `,"expected_body_sha256":"` + hex.EncodeToString(s[:]) + `"}`
				}
				h, _, srv, _, _ := net03Harness(t, in, net03Response(200, body), "unknown", false, false)
				c := h.initialized(t)
				defer c.Close()
				want, class := "validated", "expected"
				if withDigest && size > net03.PrefixBytes {
					want, class = "unexpected_response", "unexpected"
				}
				raw := tcpRaw(t, h, c, net03Task("prefix", in, `{}`))
				m := net03Result(t, raw, "completed", "", want)
				b := net03Last(m)["details"].(map[string]any)["body"].(map[string]any)
				if m["response_class"] != class || int(b["bytes_processed"].(float64)) != min(size, net03.PrefixBytes) || b["truncated"] != (size > net03.PrefixBytes) {
					t.Fatalf("prefix: %s", raw)
				}
				if len(srv.Counts().Requests) != 1 {
					t.Fatal("second request")
				}
			})
		}
	}
	// A header block over 16,384 bytes is not a valid response: no class,
	// and the outcome follows the no-class rule.
	for _, size := range []int{16383, 16384, 16385} {
		for _, platformState := range []string{"unknown", "captive_portal"} {
			t.Run(fmt.Sprint("headers/", size, "/", platformState), func(t *testing.T) {
				prefix := "HTTP/1.1 200 OK\r\nContent-Length: 0\r\nX-Padding: "
				raw := prefix + strings.Repeat("a", size-len(prefix)-4) + "\r\n\r\n"
				h, _, _, _, _ := net03Harness(t, net03Input, raw, platformState, false, false)
				c := h.initialized(t)
				defer c.Close()
				want := "validated"
				if size > 16384 {
					want = map[string]string{"unknown": "inconclusive", "captive_portal": "captive_portal_suspected"}[platformState]
				}
				m := net03Result(t, tcpRaw(t, h, c, net03Task("headers", net03Input, `{}`)), "completed", "", want)
				if _, has := m["response_class"]; has != (size <= 16384) {
					t.Fatal(m)
				}
			})
		}
	}
	// Request and response lines and headers count at the attempt's counted
	// boundary: the exact received budget completes, one byte less stops the
	// attempt with resource_limit and no measurement.
	response := net03Response(200, "")
	response = strings.Replace(response, "\r\n\r\n", "\r\nX-Pad: "+strings.Repeat("p", 1100-len(response)-9)+"\r\n\r\n", 1)
	for _, budget := range []int{len(response), len(response) - 1} {
		t.Run(fmt.Sprint("received/", budget), func(t *testing.T) {
			h, _, _, _, _ := net03Harness(t, net03Input, response, "unknown", false, false)
			c := h.initialized(t)
			defer c.Close()
			data := strings.Replace(net03Task("budget", net03Input, `{}`), `"limits":{}`, fmt.Sprintf(`"limits":{"max_bytes_received":%d}`, budget), 1)
			state, reason, outcome := "completed", "", "validated"
			if budget < len(response) {
				state, reason, outcome = "incomplete", "resource_limit", ""
			}
			net03Result(t, tcpRaw(t, h, c, data), state, reason, outcome)
		})
	}
	// One active validation run: a second run waits and sends nothing while
	// the first holds the network, and no run retries or sends a second
	// request.
	t.Run("one active run", func(t *testing.T) {
		h, d, srv, _, _ := net03Harness(t, net03Input, "", "unknown", false, true)
		c := h.initialized(t)
		defer c.Close()
		first, e := c.StartRun(h.admit(t, c, net03Task("first", net03Input, `{}`)))
		if e != nil {
			t.Fatal(codeOf(e))
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if e := srv.Wait(ctx, func(c loopback.HTTPCounts) bool { return len(c.Requests) == 1 }); e != nil {
			t.Fatal(e)
		}
		second, e := c.StartRun(h.admit(t, c, net03Task("second", net03Input, `{}`)))
		if e != nil {
			t.Fatal(codeOf(e))
		}
		time.Sleep(50 * time.Millisecond)
		if d.calls.Load() != 1 || len(srv.Counts().Requests) != 1 {
			t.Fatal("overlapping validation runs")
		}
		if e := c.CancelRun(first); e != nil {
			t.Fatal(e)
		}
		waitTerminal(t, c, first)
		if e := srv.Wait(ctx, func(c loopback.HTTPCounts) bool { return len(c.Requests) == 2 }); e != nil {
			t.Fatal(e)
		}
		if e := c.CancelRun(second); e != nil {
			t.Fatal(e)
		}
		waitTerminal(t, c, second)
		for _, raw := range h.run.output.Records() {
			net03Result(t, raw, "incomplete", "cancelled_by_host", "")
		}
		if d.calls.Load() != 2 {
			t.Fatal("retry or extra connection")
		}
	})
}

// requirement-ref: REQ-NET-03-0022
// requirement-ref: REQ-NET-03-0023
// requirement-ref: REQ-NET-03-0024
// requirement-ref: REQ-NET-03-0026
// requirement-ref: REQ-NET-03-0030
// requirement-ref: REQ-NET-03-0036
func TestMNet03Output(t *testing.T) {
	h, _, _, _, _ := net03Harness(t, net03Input, net03Response(200, "body-canary"), "captive_portal", false, false)
	c := h.initialized(t)
	raw := tcpRaw(t, h, c, net03Task("output", net03Input, `{}`))
	_ = c.Close()
	// The active response decides; the captured platform state is retained
	// unchanged beside it, and neither rewrites the other.
	m := net03Result(t, raw, "completed", "", "validated")
	var r struct {
		Schema string `json:"schema_version"`
		Test   struct {
			ID, Version   string
			ResultVersion string `json:"result_version"`
		}
	}
	if json.Unmarshal(raw, &r) != nil || r.Schema != "0.1.0" || r.Test.ID != "NET-03" || r.Test.Version != "0.1.0" || r.Test.ResultVersion != "0.1.0" {
		t.Fatalf("versions: %s", raw)
	}
	if m["platform_validation"] != "captive_portal" || m["response_class"] != "expected" {
		t.Fatal(m)
	}
	details := net03Last(m)["details"].(map[string]any)
	for k := range details {
		if !slices.Contains([]string{"address_family", "protocol", "status_code", "body", "checks", "content_encoding"}, k) {
			t.Fatalf("unregistered http_exchange detail %s", k)
		}
	}
	// Redirect metadata is the common minimized object only: never followed,
	// with the location reduced to its class.
	for location, class := range map[string]string{"https://1.1.1.1/portal-canary": "unregistered_https", "https://10.0.0.1/portal-canary": "prohibited_host", "http://1.1.1.1/portal-canary": "non_https", "https://u:p@1.1.1.1/portal-canary": "credentials", "": "invalid"} {
		h, _, _, _, _ := net03Harness(t, net03Input, "HTTP/1.1 302 Found\r\nLocation: "+location+"\r\nContent-Length: 0\r\n\r\n", "validated", false, false)
		c := h.initialized(t)
		m := net03Result(t, tcpRaw(t, h, c, net03Task("redirect", net03Input, `{}`)), "completed", "", "captive_portal_suspected")
		_ = c.Close()
		r := net03Last(m)["details"].(map[string]any)["redirect"].(map[string]any)
		if len(r) != 2 || r["followed_count"] != float64(0) || r["terminal_location_class"] != class || m["response_class"] != "redirect" {
			t.Fatalf("%s: %v", location, r)
		}
	}
}

// requirement-ref: REQ-NET-03-0022
func TestMNet03Versions(t *testing.T) {
	for i, v := range []string{`"0.2.0"`, `"1.0.0"`, `"0.1"`, `"v0.1.0"`, `"0.1.0 "`, `""`, `0.1`, "missing"} {
		data := net03Task(fmt.Sprint("version-", i), net03Input, `{}`)
		if v == "missing" {
			data = strings.Replace(data, `"test_version":"0.1.0",`, ``, 1)
		} else {
			data = strings.Replace(data, `"test_version":"0.1.0"`, `"test_version":`+v, 1)
		}
		h, c, d := net03AdmissionClient(t)
		if _, e := c.AdmitTask(data, h.embedded(t, data)); e == nil {
			t.Fatal(v)
		}
		if d.calls.Load() != 0 || len(c.rt.runs.entries) != 0 {
			t.Fatal("unsupported version I/O or run")
		}
	}
	// The emitted combination decodes; the result decoder rejects every
	// other version, a cross-test identity or shape and an unregistered field
	// rather than coercing it.
	h, _, _, _, _ := net03Harness(t, net03Input, net03Response(200, ""), "validated", false, false)
	c := h.initialized(t)
	defer c.Close()
	raw := tcpRaw(t, h, c, net03Task("versions", net03Input, `{}`))
	src, err := networkcontext.NewDirect(h.run.network)
	if err != nil {
		t.Fatal(err)
	}
	snap, err := src.Capture(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	decoder, err := result.New(result.Binding{TaskID: "versions", Output: measurementOutput(net03Test, net03Params(t, net03Input, `{}`)), MaxAttempts: 1, MaxDurationMS: 60000, Implementation: implementation, ContextDecoder: src.Decoder(), Path: snap.Path(), AccessGeneration: snap.Generation()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decoder.Decode(raw); err != nil {
		t.Fatal("valid result rejected", err)
	}
	reject := func(name string, mutate func(map[string]any)) {
		t.Helper()
		var v map[string]any
		_ = json.Unmarshal(raw, &v)
		mutate(v)
		data, _ := json.Marshal(v)
		if _, err := decoder.Decode(data); err == nil {
			t.Fatalf("accepted %s: %s", name, data)
		}
	}
	for _, field := range []string{"schema_version", "version", "result_version"} {
		for _, val := range []any{"0.2.0", "1.0.0", "0.1", "v0.1.0", " 0.1.0", "", 0.1, nil} {
			reject(fmt.Sprint(field, val), func(v map[string]any) {
				if field == "schema_version" {
					v[field] = val
				} else {
					v["test"].(map[string]any)[field] = val
				}
			})
		}
	}
	m := func(v map[string]any) map[string]any { return v["measurement"].(map[string]any) }
	for name, mutate := range map[string]func(map[string]any){
		"NET-01 id":            func(v map[string]any) { v["test"].(map[string]any)["id"] = "NET-01" },
		"NET-05 id":            func(v map[string]any) { v["test"].(map[string]any)["id"] = "NET-05" },
		"HTTP-01 id":           func(v map[string]any) { v["test"].(map[string]any)["id"] = "HTTP-01" },
		"NET-01 access shape":  func(v map[string]any) { m(v)["access"] = map[string]any{} },
		"NET-05 profile shape": func(v map[string]any) { m(v)["profile"] = map[string]any{} },
		"no platform state":    func(v map[string]any) { delete(m(v), "platform_validation") },
		"platform value":       func(v map[string]any) { m(v)["platform_validation"] = "portal" },
		"no response class":    func(v map[string]any) { delete(m(v), "response_class") },
		"response class value": func(v map[string]any) { m(v)["response_class"] = "intercepted" },
		"class contradicts":    func(v map[string]any) { m(v)["response_class"] = "redirect" },
		"HTTP-01 outcome":      func(v map[string]any) { m(v)["outcome"] = "expected_response" },
		"network_changed":      func(v map[string]any) { m(v)["outcome"] = "network_changed" },
		"unknown field":        func(v map[string]any) { m(v)["url"] = "x" },
	} {
		reject(name, mutate)
	}
}

// Client-level interruptions: a generation change during NET-03's own
// platform read, host cancellation and the outer deadline while the response
// is outstanding are incomplete with no measurement; a bounded protocol
// timeout is a completed outcome. The engine suite
// TestNet03AllTerminationBoundaries covers every Procedure point.
// requirement-ref: REQ-NET-03-0029
// requirement-ref: REQ-RESULT-0051
// requirement-ref: REQ-RESULT-0052
func TestMNet03Termination(t *testing.T) {
	t.Run("network_changed during platform read", func(t *testing.T) {
		h, d, _, n, _ := net03Harness(t, net03Input, net03Response(200, ""), "validated", false, false)
		n.script = func(ctx context.Context, read int) (platform.AccessState, error, bool) {
			if read != 3 {
				return platform.AccessState{}, nil, false
			}
			a := net03Access("validated")
			a.Generation++
			n.Set(a)
			return a, nil, true
		}
		c := h.initialized(t)
		defer c.Close()
		net03Result(t, tcpRaw(t, h, c, net03Task("changed", net03Input, `{}`)), "incomplete", "network_changed", "")
		if d.calls.Load() != 0 {
			t.Fatal("contact after the generation changed")
		}
	})
	for _, kind := range []string{"host", "deadline", "protocol"} {
		t.Run(kind, func(t *testing.T) {
			h, d, srv, _, _ := net03Harness(t, net03Input, "", "not_validated", false, true)
			c := h.initialized(t)
			defer c.Close()
			rh, e := c.StartRun(h.admit(t, c, net03Task("termination", net03Input, `{}`)))
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
				// No valid response: the platform state selects the outcome.
				h.advance(15 * time.Second)
				state, reason, outcome = "completed", "", "not_validated"
			}
			waitTerminal(t, c, rh)
			rows := h.run.output.Records()
			if len(rows) != 1 {
				t.Fatal("missing result")
			}
			net03Result(t, rows[0], state, reason, outcome)
			if e := srv.Wait(ctx, func(c loopback.HTTPCounts) bool { return c.Open == 0 }); e != nil || d.calls.Load() != 1 {
				t.Fatal("connection left open or second contact")
			}
		})
	}
}
