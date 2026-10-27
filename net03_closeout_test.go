package barometer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/measurement/net03"
	"github.com/tv-rain/barometer-go/internal/networkcontext"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/result"
)

// Further input types, unsafe addresses and user-specific variants are
// rejected before contact, and every final non-redirect status is admitted.
// requirement-ref: REQ-NET-03-0009
// requirement-ref: REQ-NET-03-0010
// requirement-ref: REQ-NET-03-0011
// requirement-ref: REQ-NET-03-0043
func TestMNet03AdmissionTypesAndStatuses(t *testing.T) {
	for i, in := range []string{`{"url":1,"expected_status":200}`, `{"url":true,"expected_status":200}`, `{"url":"https://8.8.8.8/check","expected_status":null}`,
		`{"url":"https://8.8.8.8/check","expected_status":true}`, `{"url":"https://8.8.8.8/check","expected_status":[200]}`, `{"url":"https://8.8.8.8/check","expected_status":200,"expected_body_sha256":1}`,
		`{"url":"https://8.8.8.8/check","expected_status":200,"expected_body_sha256":"` + net03Digest + `0"}`, `{"url":"https://8.8.8.8/check","expected_status":200,"expected_body_sha256":"` + strings.Repeat("g", 64) + `"}`,
		`{"url":"https://[fe80::1]/check","expected_status":200}`, `{"url":"https://0.0.0.0/check","expected_status":200}`, `{"url":"https://[::ffff:8.8.8.8]/check","expected_status":200}`,
		`{"url":"https://192.168.1.1/check","expected_status":200}`, `{"url":"check","expected_status":200}`} {
		assertNet03Rejected(t, net03Task(fmt.Sprint("type-", i), in, `{}`), "task_schema_invalid")
	}
	// A user-specific query is not the registered profile.
	assertNet03Rejected(t, net03Task("user", strings.Replace(net03Input, "/check", "/check?user=canary", 1), `{}`), "target_not_registered")
	for _, status := range []int{200, 300, 304, 305, 306, 404, 599} {
		in := fmt.Sprintf(`{"url":"https://8.8.8.8/check","expected_status":%d}`, status)
		h, d, _, _, _ := net03Harness(t, in, "", "unknown", false, false)
		c := h.initialized(t)
		data := net03Task("final", in, `{}`)
		if _, e := c.AdmitTask(data, h.embedded(t, data)); e != nil {
			t.Fatal(status, codeOf(e))
		}
		if d.calls.Load() != 0 {
			t.Fatal("admission I/O")
		}
		_ = c.Close()
	}
}

// An omitted family normalizes to auto. A selected family the network does
// not support stops the attempt before any contact rather than falling back.
// requirement-ref: REQ-NET-03-0015
func TestMNet03OptionsDefaultAndUnsupportedFamily(t *testing.T) {
	if f := net03Params(t, net03Input, `{}`).AddressFamily(); f != "auto" {
		t.Fatal(f)
	}
	host := strings.Replace(net03Input, "8.8.8.8", "service.example", 1)
	h, d, _, _, other := net03Harness(t, host, net03Response(200, ""), "validated", false, false)
	c := h.initialized(t)
	defer c.Close()
	net03Result(t, tcpRaw(t, h, c, net03Task("ipv6", host, `{"address_family":"ipv6"}`)), "incomplete", "constraints_unsatisfied", "")
	if d.calls.Load() != 0 || other.Load() != 0 {
		t.Fatal("contact on an unsupported family")
	}
}

// Response-class boundaries of REQ-NET-03-0041 not covered by the family
// suites: each redirect status, non-redirect 3xx statuses, encodings without a
// digest, malformed framing, a body read timeout, and the failed exchange of a
// response that is not a valid HTTP response.
// requirement-ref: REQ-NET-03-0025
// requirement-ref: REQ-NET-03-0041
// requirement-ref: REQ-NET-03-0042
func TestMNet03ResponseClassBoundaries(t *testing.T) {
	digest := strings.TrimSuffix(net03Input, "}") + `,"expected_body_sha256":"` + net03Digest + `"}`
	for _, tc := range []struct {
		name, input, raw, class, outcome string
	}{
		{"301", net03Input, "HTTP/1.1 301 Moved\r\nLocation: https://1.1.1.1/portal-canary\r\nContent-Length: 0\r\n\r\n", "redirect", "captive_portal_suspected"},
		{"303", net03Input, "HTTP/1.1 303 See Other\r\nLocation: https://1.1.1.1/portal-canary\r\nContent-Length: 0\r\n\r\n", "redirect", "captive_portal_suspected"},
		{"308", net03Input, "HTTP/1.1 308 Permanent\r\nLocation: https://1.1.1.1/portal-canary\r\nContent-Length: 0\r\n\r\n", "redirect", "captive_portal_suspected"},
		{"300", net03Input, "HTTP/1.1 300 Multiple\r\nLocation: https://1.1.1.1/portal-canary\r\nContent-Length: 0\r\n\r\n", "unexpected", "unexpected_response"},
		{"304", net03Input, "HTTP/1.1 304 Not Modified\r\n\r\n", "unexpected", "unexpected_response"},
		{"expected-304", strings.Replace(net03Input, "200", "304", 1), "HTTP/1.1 304 Not Modified\r\n\r\n", "expected", "validated"},
		{"encoding-no-digest", net03Input, "HTTP/1.1 200 OK\r\nContent-Encoding: gzip\r\nContent-Length: 2\r\n\r\nok", "expected", "validated"},
		{"malformed-chunk", digest, "HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\nzz\r\nok\r\n0\r\n\r\n", "unexpected", "unexpected_response"},
		{"no-status", net03Input, "garbage\r\n\r\n", "", "inconclusive"},
		{"header-over-bound", net03Input, "HTTP/1.1 200 OK\r\nX-Padding: " + strings.Repeat("a", 16384) + "\r\n\r\n", "", "inconclusive"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, d, srv, _, other := net03Harness(t, tc.input, tc.raw, "unknown", false, false)
			c := h.initialized(t)
			defer c.Close()
			raw := tcpRaw(t, h, c, net03Task("class", tc.input, `{}`))
			m := net03Result(t, raw, "completed", "", tc.outcome)
			class, has := m["response_class"]
			if has != (tc.class != "") || has && class != tc.class {
				t.Fatalf("class: %s", raw)
			}
			last := net03Last(m)
			if tc.class == "" && (last["status"] != "failed" || last["failure"] == nil) {
				t.Fatalf("no failed exchange: %s", raw)
			}
			if d.calls.Load() != 1 || other.Load() != 0 || len(srv.Counts().Requests) != 1 {
				t.Fatal("additional contact")
			}
		})
	}
	// The body read times out after the headers were parsed: once every
	// served byte was read, the client's next read on the plain connection
	// waits for the missing body bytes when the bounded operation timeout
	// fires.
	t.Run("body-timeout", func(t *testing.T) {
		in := `{"url":"http://8.8.8.8/check","expected_status":200}`
		served := "HTTP/1.1 200 OK\r\nContent-Length: 10\r\n\r\nok"
		h, d, _, _, _ := net03Harness(t, in, served, "captive_portal", true, true)
		counter, dial := &net03ReadCounter{want: int64(len(served))}, d.dial
		d.dial = func(ctx context.Context, n, a string) (measurement.Conn, error) {
			conn, e := dial(ctx, n, a)
			if e != nil {
				return conn, e
			}
			counter.Conn = conn
			return counter, nil
		}
		c := h.initialized(t)
		defer c.Close()
		rh, e := c.StartRun(h.admit(t, c, net03Task("body-timeout", in, `{}`)))
		if e != nil {
			t.Fatal(e)
		}
		for deadline := time.Now().Add(5 * time.Second); !counter.waiting.Load(); time.Sleep(time.Millisecond) {
			if time.Now().After(deadline) {
				t.Fatal("body read not reached")
			}
		}
		h.advance(15 * time.Second)
		waitTerminal(t, c, rh)
		rows := h.run.output.Records()
		m := net03Result(t, rows[0], "completed", "", "unexpected_response")
		last := net03Last(m)
		if m["response_class"] != "unexpected" || m["platform_validation"] != "captive_portal" || last["failure"] != "timeout" || d.calls.Load() != 1 {
			t.Fatal(m)
		}
	})
}

// net03ReadCounter reports a client read that starts after want bytes were
// already read from the connection.
type net03ReadCounter struct {
	measurement.Conn
	want    int64
	total   atomic.Int64
	waiting atomic.Bool
}

func (c *net03ReadCounter) Read(p []byte) (int, error) {
	if c.total.Load() >= c.want {
		c.waiting.Store(true)
	}
	n, err := c.Conn.Read(p)
	c.total.Add(int64(n))
	return n, err
}

// A platform read that fails is unknown, and the request still decides.
// requirement-ref: REQ-NET-03-0005
// requirement-ref: REQ-NET-03-0006
func TestMNet03PlatformReadFailure(t *testing.T) {
	h, d, _, n, _ := net03Harness(t, net03Input, net03Response(200, ""), "captive_portal", false, false)
	n.script = func(_ context.Context, read int) (platform.AccessState, error, bool) {
		if read != 3 {
			return platform.AccessState{}, nil, false
		}
		return platform.AccessState{}, errors.New("CANARY-RAW-ERROR"), true
	}
	c := h.initialized(t)
	defer c.Close()
	m := net03Result(t, tcpRaw(t, h, c, net03Task("read-failure", net03Input, `{}`)), "completed", "", "validated")
	if m["platform_validation"] != "unknown" || m["response_class"] != "expected" || d.calls.Load() != 1 {
		t.Fatal(m)
	}
}

// Whole foreign measurement shapes and task-side version failures.
// requirement-ref: REQ-NET-03-0022
func TestMNet03VersionShapes(t *testing.T) {
	// A well-formed unregistered version is unsupported; a malformed or
	// missing one fails the task schema. Neither runs or contacts anything.
	for v, want := range map[string]string{`"0.2.0"`: "task_unsupported", `"1.0.0"`: "task_unsupported", `"0.1"`: "task_schema_invalid", `"v0.1.0"`: "task_schema_invalid",
		`"0.1.0 "`: "task_schema_invalid", `""`: "task_schema_invalid", `0.1`: "task_schema_invalid", "missing": "task_schema_invalid"} {
		data := net03Task("version", net03Input, `{}`)
		if v == "missing" {
			data = strings.Replace(data, `"test_version":"0.1.0",`, ``, 1)
		} else {
			data = strings.Replace(data, `"test_version":"0.1.0"`, `"test_version":`+v, 1)
		}
		assertNet03Rejected(t, data, want)
	}
	h, _, _, _, _ := net03Harness(t, net03Input, net03Response(200, ""), "validated", false, false)
	c := h.initialized(t)
	defer c.Close()
	raw := tcpRaw(t, h, c, net03Task("shapes", net03Input, `{}`))
	src, err := networkcontext.NewDirect(h.run.network)
	if err != nil {
		t.Fatal(err)
	}
	snap, err := src.Capture(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	decoder, err := result.New(result.Binding{TaskID: "shapes", Output: measurementOutput(net03Test, net03Params(t, net03Input, `{}`)), MaxAttempts: 1, MaxDurationMS: 60000, Implementation: implementation, ContextDecoder: src.Decoder(), Path: snap.Path(), AccessGeneration: snap.Generation()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decoder.Decode(raw); err != nil {
		t.Fatal("valid result rejected", err)
	}
	for name, shape := range map[string]map[string]any{
		"NET-01 access-only": {"outcome": "available", "observations": []any{}, "access": map[string]any{"network_type": "wifi"}},
		"NET-05 profile":     {"outcome": "completed", "observations": []any{}, "profile": map[string]any{"id": "p"}},
	} {
		var v map[string]any
		_ = json.Unmarshal(raw, &v)
		v["measurement"] = shape
		data, _ := json.Marshal(v)
		if _, err := decoder.Decode(data); err == nil {
			t.Fatalf("accepted %s", name)
		}
	}
}

// Reading stops at the registered prefix: with a 16,384-byte body, a received
// budget of exactly the response head plus 1,024 bytes completes with the
// truncated prefix, and one byte less stops the attempt with resource_limit,
// so no byte past the prefix is counted as read.
// requirement-ref: REQ-NET-03-0021
// requirement-ref: REQ-NET-03-0041
func TestMNet03PrefixStopsReading(t *testing.T) {
	in := `{"url":"http://8.8.8.8/check","expected_status":200}`
	response := net03Response(200, strings.Repeat("a", 16384))
	head := len(response) - 16384
	for _, budget := range []int{head + net03.PrefixBytes, head + net03.PrefixBytes - 1} {
		t.Run(fmt.Sprint(budget), func(t *testing.T) {
			h, d, srv, _, _ := net03Harness(t, in, response, "unknown", true, false)
			c := h.initialized(t)
			defer c.Close()
			data := strings.Replace(net03Task("prefix-read", in, `{}`), `"limits":{}`, fmt.Sprintf(`"limits":{"max_bytes_received":%d}`, budget), 1)
			raw := tcpRaw(t, h, c, data)
			if budget < head+net03.PrefixBytes {
				net03Result(t, raw, "incomplete", "resource_limit", "")
			} else {
				m := net03Result(t, raw, "completed", "", "validated")
				b := net03Last(m)["details"].(map[string]any)["body"].(map[string]any)
				if m["response_class"] != "expected" || b["bytes_processed"] != float64(net03.PrefixBytes) || b["truncated"] != true {
					t.Fatal(m)
				}
			}
			if d.calls.Load() != 1 || len(srv.Counts().Requests) != 1 {
				t.Fatal("retry or second request")
			}
		})
	}
}
