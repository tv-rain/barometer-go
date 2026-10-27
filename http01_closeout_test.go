package barometer

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

// assertHttp01Rejected admits one HTTP-01 task in a fresh client and requires
// the exact admission class, no task or run handle, no replay commit and no
// contact on the recording transport.
func assertHttp01Rejected(t *testing.T, data, want string) {
	t.Helper()
	h, c, d := httpAdmissionClient(t)
	handle, e := c.AdmitTask(data, h.embedded(t, data))
	if handle != "" || codeOf(e) != want {
		t.Fatalf("admission = %q, %s; want %s", handle, codeOf(e), want)
	}
	if taskHandles(c) != 0 || h.replay.Commits() != 0 || d.calls.Load() != 0 || len(c.rt.runs.entries) != 0 {
		t.Fatal("rejection published state or contacted transport")
	}
}

// http01Checks renders the terminal observation's indexed check results.
func http01Checks(t *testing.T, r httpResult) string {
	t.Helper()
	obs := r.Measurement.Observations
	raw, _ := obs[len(obs)-1].Details["checks"].([]any)
	var out []string
	for _, v := range raw {
		c := v.(map[string]any)
		out = append(out, fmt.Sprintf("%s%d:%t", c["kind"], int(c["index"].(float64)), c["matched"]))
	}
	return strings.Join(out, " ")
}

const http01SemanticsResponse = "HTTP/1.1 200 OK\r\nContent-Length: 14\r\nX-Fixture: header-canary\r\nX-Multi: first\r\nX-Multi: second\r\n\r\nMixedCase-Body"

// Body string operations are case-sensitive, header names are
// case-insensitive, a multivalue header matches when any value matches, and
// every configured check must match. Indexed results are asserted through the
// production runner and serialized result.
// requirement-ref: REQ-HTTP-01-0012
// requirement-ref: REQ-HTTP-01-0013
// requirement-ref: REQ-HTTP-01-0014
// requirement-ref: REQ-HTTP-01-0037
func TestMHttp01CheckSemantics(t *testing.T) {
	for _, tc := range []struct{ name, checks, opt, raw, want, results string }{
		{"body-contains-case", `,"body_checks":[{"type":"contains","value":"mixedcase-body"}]`, `{}`, "", "body_check_failed", "status0:true body0:false"},
		{"body-prefix-case", `,"body_checks":[{"type":"prefix","value":"MIXEDCASE"}]`, `{}`, "", "body_check_failed", "status0:true body0:false"},
		{"body-suffix-case", `,"body_checks":[{"type":"suffix","value":"-BODY"}]`, `{}`, "", "body_check_failed", "status0:true body0:false"},
		{"body-exact-case", `,"body_checks":[{"type":"contains","value":"Case-Bo"},{"type":"prefix","value":"Mixed"},{"type":"suffix","value":"-Body"}]`, `{}`, "", "expected_response", "status0:true body0:true body1:true body2:true"},
		{"header-name-case", `,"header_checks":[{"name":"X-FIXTURE","type":"equals","value":"header-canary"},{"name":"x-fixture","type":"present"},{"name":"X-fIxTuRe","type":"contains","value":"canary"}]`, `{}`, "", "expected_response", "status0:true header0:true header1:true header2:true"},
		{"multivalue-any", `,"header_checks":[{"name":"x-multi","type":"equals","value":"second"},{"name":"x-multi","type":"equals","value":"first"},{"name":"x-multi","type":"contains","value":"eco"}]`, `{}`, "", "expected_response", "status0:true header0:true header1:true header2:true"},
		{"multivalue-none", `,"header_checks":[{"name":"x-multi","type":"equals","value":"third"}]`, `{}`, "", "header_check_failed", "status0:true header0:false"},
		{"conjunction", `,"header_checks":[{"name":"x-fixture","type":"present"},{"name":"x-multi","type":"equals","value":"third"},{"name":"missing","type":"absent"}],"body_checks":[{"type":"contains","value":"MixedCase"},{"type":"suffix","value":"nope"}]`, `{}`, "", "header_check_failed", "status0:true header0:true header1:false header2:true body0:true body1:false"},
		// Truncation leaves the suffix and missing-substring checks undecidable;
		// only the decided prefix result is retained.
		{"partial-undecided", `,"body_checks":[{"type":"prefix","value":"Mi"},{"type":"suffix","value":"Body"},{"type":"contains","value":"zz"}]`, `{"max_body_bytes":4}`, "", "body_limit_reached", "status0:true body0:true"},
		// A decided mismatch wins over truncation and the undecided check is omitted.
		{"partial-mismatch", `,"body_checks":[{"type":"suffix","value":"Body"},{"type":"prefix","value":"Mx"}]`, `{"max_body_bytes":4}`, "", "body_check_failed", "status0:true body1:false"},
		{"status-mismatch-keeps-later", `,"header_checks":[{"name":"x-fixture","type":"present"}],"body_checks":[{"type":"prefix","value":"Mixed"}]`, `{}`, strings.Replace(http01SemanticsResponse, "200 OK", "203 Other", 1), "unexpected_status", "status0:false header0:true body0:true"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := tc.raw
			if raw == "" {
				raw = http01SemanticsResponse
			}
			in := strings.TrimSuffix(httpInput, "}") + tc.checks + "}"
			h, _, _ := httpHarness(t, in, raw, false)
			c := h.initialized(t)
			defer c.Close()
			r := checkHTTP(t, tcpRaw(t, h, c, httpTask("semantics", in, tc.opt)), "completed", "", tc.want)
			if got := http01Checks(t, r); got != tc.results {
				t.Fatalf("checks %q want %q", got, tc.results)
			}
			for _, canary := range []string{"MixedCase", "mixedcase", "second", "third", "x-multi"} {
				if strings.Contains(fmt.Sprint(r.Measurement), canary) {
					t.Fatal("retained check data", canary)
				}
			}
		})
	}
}

// ipv6 is accepted only for an IPv6 literal (or hostname) URL. The run uses
// the registered IPv6 literal authority; the synthetic transport routes that
// checked authority to the loopback fixture, so this demonstrates address-
// family selection and the ipv6 observation, not real IPv6 reachability.
// requirement-ref: REQ-HTTP-01-0024
// requirement-ref: REQ-HTTP-01-0025
func TestMHttp01OptionsAddressFamilyAndDefaults(t *testing.T) {
	const v6 = `{"url":"https://[2606:4700:4700::1111]/check","expected_status":200}`
	assertHttp01Rejected(t, httpTask("v4-literal-ipv6", httpInput, `{"address_family":"ipv6"}`), "task_schema_invalid")
	assertHttp01Rejected(t, httpTask("v6-literal-ipv4", v6, `{"address_family":"ipv4"}`), "task_schema_invalid")
	if p := httpParams(t, v6, `{"address_family":"ipv6"}`); p.AddressFamily() != "ipv6" {
		t.Fatal(p.AddressFamily())
	}
	if p := httpParams(t, strings.Replace(httpInput, "8.8.8.8", "service.example", 1), `{"address_family":"ipv6"}`); p.AddressFamily() != "ipv6" {
		t.Fatal("hostname ipv6")
	}
	for _, opt := range []string{`{"address_family":"ipv6"}`, `{"address_family":"auto"}`, `{}`} {
		t.Run(opt, func(t *testing.T) {
			h, d, srv := httpHarness(t, v6, responseHTTP(""), false)
			d.dial = func(ctx context.Context, _, a string) (measurement.Conn, error) {
				if a != "[2606:4700:4700::1111]:443" {
					return nil, fmt.Errorf("CANARY-RAW-ERROR")
				}
				var nd net.Dialer
				return nd.DialContext(ctx, "tcp", srv.Address())
			}
			access := defaultAccess
			access.IPv6 = platform.AvailabilityYes
			h.run.network.Set(access)
			c := h.initialized(t)
			defer c.Close()
			r := checkHTTP(t, tcpRaw(t, h, c, httpTask("ipv6", v6, opt)), "completed", "", "expected_response")
			for i, o := range r.Measurement.Observations {
				if o.Type != "tcp_connect" && o.Type != "tls_handshake" && o.Type != "http_exchange" || o.Details["address_family"] != "ipv6" {
					t.Fatal(i, o)
				}
			}
			if strings.Contains(string(tcpRawLast(t, h)), "2606:4700") {
				t.Fatal("retained address")
			}
		})
	}
	// Default max_body_bytes is 4096 in the normalized limits and at run time.
	if httpParams(t, httpInput, `{}`).Limits().MaxBodyBytes != 4096 {
		t.Fatal("default body bound")
	}
	h, _, _ := httpHarness(t, httpInput, responseHTTP(strings.Repeat("a", 4097)), false)
	c := h.initialized(t)
	defer c.Close()
	r := checkHTTP(t, tcpRaw(t, h, c, httpTask("default-body", httpInput, `{}`)), "completed", "", "expected_response")
	b := r.Measurement.Observations[2].Details["body"].(map[string]any)
	if int(b["bytes_processed"].(float64)) != 4096 || b["truncated"] != true {
		t.Fatal(b)
	}
	// The default bound participates in the outer receive-bound admission rule.
	data := strings.Replace(httpTask("default-receive-bound", httpInput, `{}`), `"limits":{}`, `"limits":{"max_bytes_received":4095}`, 1)
	assertHttp01Rejected(t, data, "task_limit_exceeded")
}

func tcpRawLast(t *testing.T, h *harness) []byte {
	t.Helper()
	rows := h.run.output.Records()
	return rows[len(rows)-1]
}

// Every custom request, credential, proxy, compression or TLS override is
// outside the closed options shape.
// requirement-ref: REQ-HTTP-01-0024
// requirement-ref: REQ-HTTP-01-0025
func TestMHttp01OptionsOverridesRejected(t *testing.T) {
	for _, opt := range []string{`{"method":"HEAD"}`, `{"headers":{"X-Id":"1"}}`, `{"user_agent":"x"}`, `{"body":"x"}`, `{"request_body":"x"}`, `{"cookies":true}`, `{"cookie":"a=b"}`, `{"credentials":{"user":"u"}}`, `{"authorization":"x"}`, `{"proxy":{"host":"127.0.0.1"}}`, `{"network_path":"proxy"}`, `{"compression":true}`, `{"accept_encoding":"gzip"}`, `{"tls":{}}`, `{"server_name":"x"}`, `{"alpn_protocols":["h2"]}`, `{"insecure_skip_verify":true}`, `{"follow_redirects":null}`, `{"max_redirects":null}`, `{"max_body_bytes":null}`} {
		assertHttp01Rejected(t, httpTask("override", httpInput, opt), "task_schema_invalid")
	}
}

// Rejected and unsupported test identities and versions on the task side.
// requirement-ref: REQ-HTTP-01-0033
func TestMHttp01TaskVersionShapes(t *testing.T) {
	base := httpTask("version", httpInput, `{}`)
	for _, v := range []string{`0.1`, `null`, `1`, `"0.1.0\t"`} {
		assertHttp01Rejected(t, strings.Replace(base, `"test_version":"0.1.0"`, `"test_version":`+v, 1), "task_schema_invalid")
	}
	assertHttp01Rejected(t, strings.Replace(base, `"test_version":"0.1.0",`, ``, 1), "task_schema_invalid")
	for _, id := range []string{"http-01", "NET-03", "TLS-01", "NET-01", "NET-05"} {
		h, c, d := httpAdmissionClient(t)
		data := strings.Replace(base, `"test_id":"HTTP-01"`, fmt.Sprintf(`"test_id":%q`, id), 1)
		handle, e := c.AdmitTask(data, h.embedded(t, data))
		if handle != "" || e == nil || taskHandles(c) != 0 || h.replay.Commits() != 0 || d.calls.Load() != 0 {
			t.Fatalf("cross-test task %s: %s", id, codeOf(e))
		}
	}
}

// Receiver-side missing fields and cross-test measurement shapes.
// requirement-ref: REQ-HTTP-01-0033
func TestMHttp01ResultVersionShapes(t *testing.T) {
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
	mutate := func(edit func(map[string]any)) []byte {
		var v map[string]any
		if e := json.Unmarshal(raw, &v); e != nil {
			t.Fatal(e)
		}
		edit(v)
		b, _ := json.Marshal(v)
		return b
	}
	access := map[string]any{"network_type": "wifi", "vpn_state": "unknown", "ipv4_available": "yes", "ipv6_available": "unknown", "validated": "unknown", "network_changed": false}
	for name, edit := range map[string]func(map[string]any){
		"missing schema_version": func(v map[string]any) { delete(v, "schema_version") },
		"missing test.version":   func(v map[string]any) { delete(v["test"].(map[string]any), "version") },
		"missing result_version": func(v map[string]any) { delete(v["test"].(map[string]any), "result_version") },
		"missing test.id":        func(v map[string]any) { delete(v["test"].(map[string]any), "id") },
		"NET-03 id":              func(v map[string]any) { v["test"].(map[string]any)["id"] = "NET-03" },
		"NET-01 access field":    func(v map[string]any) { v["measurement"].(map[string]any)["access"] = access },
		"NET-01 access-only":     func(v map[string]any) { v["measurement"] = map[string]any{"access": access} },
		"NET-05 profile": func(v map[string]any) {
			v["measurement"].(map[string]any)["profile"] = map[string]any{"profile_id": "p", "profile_version": 1}
		},
		"NET-05 steps": func(v map[string]any) { v["measurement"].(map[string]any)["steps"] = []any{} },
		"NET-05 validation": func(v map[string]any) {
			v["measurement"].(map[string]any)["network_validation"] = map[string]any{"state": "unknown", "source": "platform"}
		},
	} {
		if _, e := decoder.Decode(mutate(edit)); e == nil {
			t.Fatal("accepted", name)
		}
	}
}

// Application-byte and connection bounds at below/equal/above through the
// Client on the direct and CONNECT proxy paths. Usage is calibrated from the
// loopback server's plaintext counters; equality succeeding on the proxy path
// also shows the CONNECT exchange is not charged to the attempt.
// requirement-ref: REQ-HTTP-01-0032
// requirement-ref: REQ-LIMIT-0034
func TestMHttp01OuterBoundsBoundary(t *testing.T) {
	const body = "0123456789"
	for _, proxy := range []bool{false, true} {
		for _, kind := range []string{"sent", "received", "connections"} {
			t.Run(fmt.Sprint(kind, "/proxy=", proxy), func(t *testing.T) {
				raw, want, opts := responseHTTP(body), "expected_response", `{"max_body_bytes":16}`
				if kind == "connections" {
					raw, want, opts = "HTTP/1.1 302 Found\r\nLocation: https://8.8.8.8/check\r\nContent-Length: 0\r\n\r\n", "redirect_limit_reached", `{"follow_redirects":true,"max_redirects":1}`
				}
				h, d, srv := httpHarness(t, httpInput, raw, false)
				if kind == "connections" {
					httpPolicy(t, h, httpInput, "https://8.8.8.8/check", "https://8.8.8.8/check")
				}
				cfg, path := `{}`, ""
				var p *loopback.Proxy
				if proxy {
					var e error
					p, e = loopback.ListenProxy(loopback.ProxyConfig{Resolve: map[string]string{"8.8.8.8:443": srv.Address()}})
					if e != nil {
						t.Fatal(e)
					}
					defer p.Close()
					cfg, path = proxyConfig(p.Port()), `,"network_path":"proxy"`
				}
				c := h.client(t)
				if e := c.Initialize(cfg); e != nil {
					t.Fatal(codeOf(e))
				}
				defer c.Close()
				task := func(id, limits string) string {
					return strings.Replace(httpTask(id, httpInput, opts), `"limits":{}`, `"limits":{`+limits+`}`+path, 1)
				}
				settle := func() loopback.HTTPCounts {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					if e := srv.Wait(ctx, func(c loopback.HTTPCounts) bool { return c.Open == 0 }); e != nil {
						t.Fatal(e)
					}
					return srv.Counts()
				}
				usage := 2
				switch kind {
				case "sent":
					checkHTTP(t, tcpRaw(t, h, c, task("calibrate", "")), "completed", "", want)
					usage = settle().In
				case "received":
					usage = len(raw)
				}
				if usage <= 0 {
					t.Fatal("calibration")
				}
				field := map[string]string{"sent": "max_bytes_sent", "received": "max_bytes_received", "connections": "max_connections"}[kind]
				for _, delta := range []int{1, 0, -1} {
					before := settle()
					data := task(fmt.Sprint("bound-", delta+1), fmt.Sprintf(`%q:%d`, field, usage+delta))
					if delta < 0 {
						checkHTTP(t, tcpRaw(t, h, c, data), "incomplete", "resource_limit", "")
					} else {
						checkHTTP(t, tcpRaw(t, h, c, data), "completed", "", want)
					}
					after := settle()
					accepted := after.Accepted - before.Accepted
					wantConns := 1
					if kind == "connections" {
						wantConns = min(2, usage+delta)
					}
					if accepted != wantConns {
						t.Fatalf("connections %d want %d", accepted, wantConns)
					}
					if kind == "sent" && delta >= 0 && after.In-before.In != usage {
						t.Fatal("sent usage changed")
					}
				}
				if proxy && d.calls.Load() != 0 {
					t.Fatal("direct fallback")
				}
				if proxy && p.Counts().HeadIn == 0 {
					t.Fatal("CONNECT path not used")
				}
			})
		}
	}
}

// A differently spelled but equivalent URL normalizes to the same key and is
// serialized behind the active run rather than contacted concurrently.
// requirement-ref: REQ-HTTP-01-0032
func TestMHttp01EquivalentSpellingSerialized(t *testing.T) {
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
	spelled := strings.Replace(httpInput, "https://8.8.8.8/check", "HTTPS://8.8.8.8:443/check", 1)
	if httpParams(t, spelled, `{}`).Target().URL != httpParams(t, httpInput, `{}`).Target().URL {
		t.Fatal("fixture is not equivalent")
	}
	second, e := c.StartRun(h.admit(t, c, httpTask("second", spelled, `{}`)))
	if e != nil {
		t.Fatal(e)
	}
	// The second run is waiting on the per-URL key: no dial, no request.
	time.Sleep(20 * time.Millisecond)
	if d.calls.Load() != 1 || len(srv.Counts().Requests) != 1 {
		t.Fatal("concurrent equivalent URL")
	}
	if e := c.CancelRun(second); e != nil {
		t.Fatal(e)
	}
	waitTerminal(t, c, second)
	checkHTTP(t, h.run.output.Records()[0], "not_run", "cancelled_by_host", "")
	if e := c.CancelRun(first); e != nil {
		t.Fatal(e)
	}
	waitTerminal(t, c, first)
	checkHTTP(t, h.run.output.Records()[1], "incomplete", "cancelled_by_host", "")
	if d.calls.Load() != 1 {
		t.Fatal("serialized run contacted after cancellation")
	}
}

// The runner cannot produce an inconclusive body read over crypto/tls (see
// internal/protocol/http TestMHttp01NormalizationInconclusive). The bound
// HTTP-01 result decoder still accepts the registered inconclusive shape,
// with only the decided status check retained, and rejects an outcome whose
// terminal failure differs.
// requirement-ref: REQ-HTTP-01-0036
func TestMHttp01InconclusiveResultShape(t *testing.T) {
	in := strings.TrimSuffix(httpInput, "}") + `,"body_checks":[{"type":"contains","value":"x"}]}`
	h, _, _ := httpHarness(t, in, responseHTTP("x"), false)
	c := h.initialized(t)
	defer c.Close()
	raw := tcpRaw(t, h, c, httpTask("inconclusive", in, `{}`))
	cd, _ := observation.NewContextDecoder(result.Contract, observation.ContextApprovals{})
	path, _ := observation.NewPath(1, 0, "direct", "")
	decoder, e := result.New(result.Binding{TaskID: "inconclusive", Output: result.Output{Test: model.Test{ID: "HTTP-01", Version: "0.1.0", ResultVersion: "0.1.0"}, Limits: httpParams(t, in, `{}`).Limits(), HTTPS: true}, MaxAttempts: 1, MaxDurationMS: 60000, Implementation: implementation, ContextDecoder: cd, Path: path, AccessGeneration: 1})
	if e != nil {
		t.Fatal(e)
	}
	shape := func(outcome, failure string) []byte {
		var v map[string]any
		_ = json.Unmarshal(raw, &v)
		m := v["measurement"].(map[string]any)
		m["outcome"] = outcome
		obs := m["observations"].([]any)
		last := obs[len(obs)-1].(map[string]any)
		last["status"], last["failure"] = "failed", failure
		d := last["details"].(map[string]any)
		d["body"] = map[string]any{"bytes_processed": 0, "truncated": false}
		d["checks"] = d["checks"].([]any)[:1] // status only: the body check was not decided
		b, _ := json.Marshal(v)
		return b
	}
	if _, e := decoder.Decode(shape("inconclusive", "inconclusive")); e != nil {
		t.Fatal("registered inconclusive shape rejected", e)
	}
	if _, e := decoder.Decode(shape("inconclusive", "timeout")); e == nil {
		t.Fatal("outcome/failure mismatch accepted")
	}
}
