package http

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/observation"
	"github.com/tv-rain/barometer-go/internal/testkit"
)

type memoryConn struct {
	*bytes.Reader
	sent   bytes.Buffer
	closed bool
}

func (c *memoryConn) Write(p []byte) (int, error) { return c.sent.Write(p) }
func (c *memoryConn) Close() error                { c.closed = true; return nil }

type scriptedEndpoint struct {
	responses       []string
	opened, checked int
	reject          bool
	class           string
	conns           []*memoryConn
	preceding       []observation.Observation
	outcome         string
}

func (e *scriptedEndpoint) Open(context.Context, *url.URL) (Stream, []observation.Observation, string, error) {
	e.opened++
	if e.outcome != "" {
		return Stream{}, e.preceding, e.outcome, nil
	}
	c := &memoryConn{Reader: bytes.NewReader([]byte(e.responses[min(e.opened-1, len(e.responses)-1)]))}
	e.conns = append(e.conns, c)
	return Stream{c, "ipv4"}, e.preceding, "", nil
}
func (e *scriptedEndpoint) Redirect(*url.URL) (Endpoint, string, error) {
	e.checked++
	if e.reject {
		return nil, e.class, ErrInterrupted
	}
	return e, "registered_https", nil
}
func response(body string) string {
	return fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Length: %d\r\nX-Synthetic: header-canary\r\n\r\n%s", len(body), body)
}
func run(t *testing.T, c Config, responses ...string) (Result, *scriptedEndpoint) {
	t.Helper()
	e := &scriptedEndpoint{responses: responses}
	r, close, err := Exchange(context.Background(), testkit.NewClock(time.Now()), e, c)
	defer close()
	if err != nil {
		t.Fatal(err)
	}
	return r, e
}
func config() Config { return Config{URL: "https://measurement.example/static", MaxBodyBytes: 4096} }
func terminal(t *testing.T, r Result) observation.HTTPDetails {
	t.Helper()
	d, ok := r.Observations[len(r.Observations)-1].HTTP()
	if !ok {
		t.Fatal("missing HTTP")
	}
	return d
}

// requirement-ref: REQ-OBS-0028
// requirement-ref: REQ-OBS-0029
func TestCcObsHttpExchangeBody(t *testing.T) {
	for _, bound := range []int{4096, 65536} {
		for _, length := range []int{0, bound - 1, bound, bound + 1} {
			t.Run(fmt.Sprintf("%d/%d", bound, length), func(t *testing.T) {
				body := strings.Repeat("x", length)
				sum := sha256.Sum256([]byte(body))
				c := config()
				c.MaxBodyBytes = bound
				c.BodyChecks = []BodyCheck{{"sha256", hex.EncodeToString(sum[:])}}
				r, e := run(t, c, response(body))
				d := terminal(t, r)
				if d.Body == nil || d.Body.BytesProcessed != min(length, bound) || d.Body.Truncated != (length > bound) {
					t.Fatalf("body %+v", d.Body)
				}
				if length > bound {
					if r.Outcome != "body_limit_reached" || len(d.Checks) != 0 {
						t.Fatal(r.Outcome, d.Checks)
					}
				} else if r.Outcome != "expected_response" || len(d.Checks) != 1 || !d.Checks[0].Matched {
					t.Fatal(r.Outcome, d.Checks)
				}
				if !e.conns[0].closed {
					t.Fatal("body stream not closed")
				}
				b, err := json.Marshal(r)
				if err != nil || bytes.Contains(b, []byte(hex.EncodeToString(sum[:]))) || bytes.Contains(b, []byte("header-canary")) {
					t.Fatal("retained response data", err)
				}
			})
		}
	}
}

// requirement-ref: REQ-OBS-0030
func TestCcObsHttpExchangeCheck(t *testing.T) {
	c := config()
	c.ExpectedStatus = 200
	c.HeaderChecks = []HeaderCheck{{"x-synthetic", "equals", "header-canary"}, {"missing", "absent", ""}, {"X-Synthetic", "contains", "canary"}, {"x-synthetic", "present", ""}}
	c.BodyChecks = []BodyCheck{{"prefix", "body"}, {"contains", "canary"}, {"suffix", "canary"}}
	r, _ := run(t, c, response("body-canary"))
	d := terminal(t, r)
	if len(d.Checks) != 8 {
		t.Fatal(d.Checks)
	}
	for _, v := range d.Checks {
		if !v.Matched {
			t.Fatal(v)
		}
	}
	c.MaxBodyBytes = 4
	c.BodyChecks = []BodyCheck{{"prefix", "body"}, {"contains", "body"}, {"suffix", "body"}, {"sha256", strings.Repeat("0", 64)}}
	r, _ = run(t, c, response("body-canary"))
	d = terminal(t, r)
	if r.Outcome != "body_limit_reached" || len(d.Checks) != 7 || !d.Checks[5].Matched || !d.Checks[6].Matched {
		t.Fatal(r.Outcome, d.Checks)
	}
	for _, kind := range []string{"status", "header", "body"} {
		c = config()
		want := ""
		switch kind {
		case "status":
			c.ExpectedStatus = 204
			want = "unexpected_status"
		case "header":
			c.HeaderChecks = []HeaderCheck{{"missing", "present", ""}}
			want = "header_check_failed"
		case "body":
			c.BodyChecks = []BodyCheck{{"contains", "missing"}}
			want = "body_check_failed"
		}
		r, _ = run(t, c, response("body-canary"))
		if r.Outcome != want {
			t.Fatal(kind, r.Outcome)
		}
	}
}

// requirement-ref: REQ-OBS-0026
// requirement-ref: REQ-OBS-0027
// requirement-ref: REQ-LIMIT-0018
func TestCcObsHttpExchangeRedirect(t *testing.T) {
	for _, tc := range []struct {
		location, class string
		reject          bool
	}{{"https://next.example/", "registered_https", false}, {"https://next.example/", "unregistered_https", true}, {"https://127.0.0.1/", "prohibited_host", true}, {"http://next.example/", "non_https", false}, {"https://user:secret@next.example/", "credentials", false}, {"%zz", "invalid", false}, {"", "invalid", false}, {strings.Repeat("x", 2049), "invalid", false}} {
		for _, follow := range []int{0, 2} {
			t.Run(fmt.Sprintf("%s/%d", tc.class, follow), func(t *testing.T) {
				e := &scriptedEndpoint{responses: []string{"HTTP/1.1 302 Found\r\nLocation: " + tc.location + "\r\nContent-Length: 0\r\n\r\n", response("ok")}, reject: tc.reject, class: tc.class}
				c := config()
				c.MaxRedirects = follow
				r, close, err := Exchange(context.Background(), testkit.NewClock(time.Now()), e, c)
				defer close()
				if err != nil {
					t.Fatal(err)
				}
				d := terminal(t, r)
				if d.Redirect == nil || d.Redirect.TerminalLocationClass != tc.class {
					t.Fatal(d.Redirect)
				}
				want := "redirect_observed"
				contacts := 1
				if follow > 0 {
					want = "redirect_rejected"
					if tc.class == "registered_https" {
						want = "expected_response"
						contacts = 2
						if d.Redirect.FollowedCount != 1 {
							t.Fatal(d.Redirect)
						}
					}
				}
				if r.Outcome != want || e.opened != contacts {
					t.Fatal(r.Outcome, e.opened)
				}
				b, _ := json.Marshal(r)
				if bytes.Contains(b, []byte("next.example")) || bytes.Contains(b, []byte("secret")) {
					t.Fatal("redirect leaked")
				}
			})
		}
	}
	c := config()
	c.MaxRedirects = 2
	r, e := run(t, c, "HTTP/1.1 302 Found\r\nLocation: /again\r\nContent-Length: 0\r\n\r\n")
	if e.opened != 3 || r.Outcome != "redirect_limit_reached" || terminal(t, r).Redirect.FollowedCount != 2 {
		t.Fatal(e.opened, r.Outcome)
	}
}

// requirement-ref: REQ-OBS-0023
// requirement-ref: REQ-OBS-0024
func TestCcObsHttpExchangeFailures(t *testing.T) {
	for _, raw := range []string{"bad\r\n\r\n", "HTTP/1.1 999 Bad\r\n\r\n", "HTTP/1.1 200 OK\r\nContent-Length: 3\r\n\r\nx", "HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\n2\r\nx", "HTTP/1.1 101 Switching Protocols\r\n\r\n"} {
		r, _ := run(t, config(), raw)
		if r.Outcome != "protocol_error" {
			t.Fatal(r.Outcome)
		}
	}
	for _, stage := range []string{"dns_failure", "tcp_failure", "tls_failure"} {
		e := &scriptedEndpoint{outcome: stage}
		r, close, err := Exchange(context.Background(), testkit.NewClock(time.Now()), e, config())
		close()
		if err != nil || len(r.Observations) != 0 || r.Outcome != stage {
			t.Fatal("invented HTTP stage")
		}
	}
	c := config()
	c.BodyChecks = []BodyCheck{{"contains", "x"}}
	r, _ := run(t, c, "HTTP/1.1 200 OK\r\nContent-Encoding: gzip\r\nContent-Length: 1\r\n\r\nx")
	if r.Outcome != "unsupported_content_encoding" || terminal(t, r).Body != nil {
		t.Fatal(r.Outcome)
	}
}

// requirement-ref: REQ-OBS-0025
func TestCcObsHttpExchangeDetails(t *testing.T) {
	c := config()
	c.ExpectedStatus = 200
	for range 8 {
		c.HeaderChecks = append(c.HeaderChecks, HeaderCheck{"X-Synthetic", "present", ""})
		c.BodyChecks = append(c.BodyChecks, BodyCheck{"contains", "x"})
	}
	r, _ := run(t, c, response("x"))
	if len(terminal(t, r).Checks) != 17 {
		t.Fatal("maximum checks lost")
	}
	d := terminal(t, r)
	for _, family := range []string{"ipv4", "ipv6", "unknown"} {
		for _, protocol := range []string{"http/1.1", "h2", "unknown"} {
			d.AddressFamily = family
			d.Protocol = protocol
			if _, e := makeObservation(0, "", d, c); e != nil {
				t.Fatal(e)
			}
		}
	}
	for _, protocol := range []string{"", "h3"} {
		d.Protocol = protocol
		if _, e := makeObservation(0, "", d, c); e == nil {
			t.Fatal("bad protocol accepted")
		}
	}
	d.Protocol = "http/1.1"
	d.AddressFamily = ""
	if _, e := makeObservation(0, "", d, c); e == nil {
		t.Fatal("missing family accepted")
	}
}

// requirement-ref: REQ-LIMIT-0016
// requirement-ref: REQ-LIMIT-0017
func TestCcLimitHttpAndValidationLimits(t *testing.T) {
	for _, size := range []int{MaxHeaderBytes - 1, MaxHeaderBytes, MaxHeaderBytes + 1} {
		prefix := "HTTP/1.1 200 OK\r\nContent-Length: 0\r\nX-Pad: "
		raw := prefix + strings.Repeat("x", size-len(prefix)-4) + "\r\n\r\n"
		r, _ := run(t, config(), raw)
		if (r.Outcome == "expected_response") != (size <= MaxHeaderBytes) {
			t.Fatal(size, r.Outcome)
		}
	}
	for _, mutate := range []func(*Config){func(c *Config) { c.URL = strings.Repeat("x", 2049) }, func(c *Config) { c.MaxBodyBytes = 0 }, func(c *Config) { c.MaxBodyBytes = 65537 }, func(c *Config) { c.MaxRedirects = 3 }, func(c *Config) { c.ExpectedStatus = 600 }, func(c *Config) { c.HeaderChecks = make([]HeaderCheck, 9) }, func(c *Config) { c.BodyChecks = make([]BodyCheck, 9) }, func(c *Config) { c.BodyChecks = []BodyCheck{{"contains", strings.Repeat("x", 257)}} }, func(c *Config) { c.BodyChecks = []BodyCheck{{"regex", "x"}} }, func(c *Config) { c.HeaderChecks = []HeaderCheck{{"Set-Cookie", "present", ""}} }, func(c *Config) { c.PostBody = []byte("x"); c.MaxRedirects = 1 }} {
		c := config()
		mutate(&c)
		e := &scriptedEndpoint{}
		_, close, err := Exchange(context.Background(), testkit.NewClock(time.Now()), e, c)
		close()
		if err == nil || e.opened != 0 {
			t.Fatal("invalid config contacted target")
		}
	}
	r, _ := run(t, config(), "HTTP/1.1 100 Continue\r\n\r\n"+response("x"))
	if r.Outcome != "expected_response" {
		t.Fatal(r.Outcome)
	}
}

func TestPOSTPayloadIsTransient(t *testing.T) {
	c := config()
	c.PostBody = []byte("dns-canary")
	r, e := run(t, c, "HTTP/1.1 200 OK\r\nContent-Type: application/dns-message\r\nContent-Length: 3\r\n\r\ndns")
	payload := r.TakeBody()
	if string(payload) != "dns" || r.TakeBody() != nil {
		t.Fatal("payload ownership")
	}
	request := e.conns[0].sent.String()
	if !strings.HasPrefix(request, "POST ") || !strings.HasSuffix(request, "dns-canary") || !strings.Contains(request, "Accept-Encoding: identity") || strings.Contains(request, "User-Agent:") || strings.Contains(request, "Cookie:") {
		t.Fatal("unexpected request")
	}
	if strings.Contains(fmt.Sprintf("%+v %#v", r, r), "canary") {
		t.Fatal("diagnostic leak")
	}
}

// A non-2xx DNS POST ends at its status before the encoding and media-type
// checks, without reading the body or classifying a Location
// (GO-SPEC-ISSUE-020). A 2xx response still checks the media type.
func TestPOSTStatusPrecedesMediaType(t *testing.T) {
	for _, tc := range []struct {
		name, raw, outcome, failure, encoding string
		status                                int
	}{
		{"503 html", "HTTP/1.1 503 Service Unavailable\r\nContent-Type: text/html\r\nContent-Length: 13\r\n\r\nbody-canary!!", "expected_response", "", "identity", 503},
		{"404 dns message", "HTTP/1.1 404 Not Found\r\nContent-Type: application/dns-message\r\nContent-Length: 3\r\n\r\ndns", "expected_response", "", "identity", 404},
		{"302 location", "HTTP/1.1 302 Found\r\nLocation: https://next.example/dns\r\nContent-Length: 0\r\n\r\n", "expected_response", "", "identity", 302},
		{"500 gzip", "HTTP/1.1 500 Internal Server Error\r\nContent-Encoding: gzip\r\nContent-Length: 3\r\n\r\nxyz", "expected_response", "", "", 500},
		{"200 html", "HTTP/1.1 200 OK\r\nContent-Type: text/html\r\nContent-Length: 3\r\n\r\ndns", "unexpected_content_type", "unexpected_content_type", "identity", 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := config()
			c.PostBody = []byte("dns")
			r, e := run(t, c, tc.raw)
			o := r.Observations[len(r.Observations)-1]
			d := terminal(t, r)
			if r.Outcome != tc.outcome || o.Failure() != tc.failure || d.StatusCode == nil || *d.StatusCode != tc.status || d.ContentEncoding != tc.encoding || d.Redirect != nil || e.checked != 0 || e.opened != 1 {
				t.Fatalf("%s %s %+v", r.Outcome, o.Failure(), d)
			}
			if tc.status != 200 && (d.Body != nil || r.TakeBody() != nil) {
				t.Fatal("non-2xx body read")
			}
		})
	}
}

type lateRefusal struct{ scriptedEndpoint }

func (e *lateRefusal) Open(ctx context.Context, u *url.URL) (Stream, []observation.Observation, string, error) {
	if e.opened > 0 {
		return Stream{}, nil, "", &DestinationRejected{Class: "prohibited_host"}
	}
	return e.scriptedEndpoint.Open(ctx, u)
}
func (e *lateRefusal) Redirect(*url.URL) (Endpoint, string, error) { return e, "registered_https", nil }
func TestRedirectRefusedByPerContactGate(t *testing.T) {
	e := &lateRefusal{scriptedEndpoint: scriptedEndpoint{responses: []string{"HTTP/1.1 302 Found\r\nLocation: https://next.example/check\r\nContent-Length: 0\r\n\r\n"}}}
	c := config()
	c.MaxRedirects = 2
	r, close, err := Exchange(context.Background(), testkit.NewClock(time.Now()), e, c)
	defer close()
	if err != nil || r.Outcome != "redirect_rejected" || len(r.Observations) != 1 || e.opened != 1 {
		t.Fatal(err, r.Outcome)
	}
	d := terminal(t, r)
	if d.Redirect.FollowedCount != 0 || d.Redirect.TerminalLocationClass != "prohibited_host" {
		t.Fatal(d.Redirect)
	}
}

func TestBoundedFramingAndUndecidableChecks(t *testing.T) {
	for _, tc := range []struct {
		raw       string
		bound     int
		truncated bool
	}{
		{"HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\n3\r\nabc\r\n0\r\n\r\n", 4, false},
		{"HTTP/1.1 200 OK\r\nConnection: close\r\n\r\nabc", 4, false},
		{"HTTP/1.1 200 OK\r\nConnection: close\r\n\r\nabc", 3, true},
	} {
		c := config()
		c.MaxBodyBytes = tc.bound
		c.BodyChecks = []BodyCheck{{"suffix", "abc"}}
		r, _ := run(t, c, tc.raw)
		d := terminal(t, r)
		if d.Body.BytesProcessed != 3 || d.Body.Truncated != tc.truncated {
			t.Fatal(d.Body)
		}
		if (r.Outcome == "body_limit_reached") != tc.truncated {
			t.Fatal(r.Outcome)
		}
	}
	c := config()
	c.MaxBodyBytes = 1
	c.BodyChecks = []BodyCheck{{"prefix", "other"}, {"contains", "other"}, {"suffix", "x"}}
	r, _ := run(t, c, response("xyz"))
	d := terminal(t, r)
	if r.Outcome != "body_check_failed" || len(d.Checks) != 1 || d.Checks[0].Matched {
		t.Fatal(r.Outcome, d.Checks)
	}
	c = config()
	c.ExpectedStatus = 200
	c.HeaderChecks = []HeaderCheck{{"X-Synthetic", "equals", "two"}}
	r, _ = run(t, c, "HTTP/1.1 200 OK\r\nContent-Length: 0\r\nX-Synthetic: one\r\nX-Synthetic: two\r\n\r\n")
	if r.Outcome != "expected_response" {
		t.Fatal(r.Outcome)
	}
	// Informational blocks share the cap, even though each block is small.
	r, _ = run(t, config(), strings.Repeat("HTTP/1.1 100 Continue\r\n\r\n", 1000)+response("x"))
	if r.Outcome != "protocol_error" {
		t.Fatal(r.Outcome)
	}
}

func TestObservationRejectsInvalidBoundsAndCheckAssociations(t *testing.T) {
	c := config()
	c.ExpectedStatus = 200
	c.HeaderChecks = make([]HeaderCheck, 8)
	c.BodyChecks = make([]BodyCheck, 8)
	status := 200
	valid := observation.HTTPDetails{AddressFamily: "ipv4", Protocol: "http/1.1", StatusCode: &status, Body: &observation.Body{}}
	for _, mutate := range []func(*observation.HTTPDetails){
		func(d *observation.HTTPDetails) { d.StatusCode = nil },
		func(d *observation.HTTPDetails) { v := 600; d.StatusCode = &v },
		func(d *observation.HTTPDetails) { d.ContentEncoding = "gzip" },
		func(d *observation.HTTPDetails) { d.Body = &observation.Body{BytesProcessed: 4097} },
		func(d *observation.HTTPDetails) { d.Body = &observation.Body{BytesProcessed: -1} },
		func(d *observation.HTTPDetails) {
			d.Checks = []observation.Check{{Kind: "status", Matched: true}, {Kind: "status", Matched: true}}
		},
		func(d *observation.HTTPDetails) {
			d.Checks = []observation.Check{{Kind: "header", Index: 8, Matched: true}}
		},
		func(d *observation.HTTPDetails) {
			d.Checks = []observation.Check{{Kind: "body", Index: 8, Matched: true}}
		},
		func(d *observation.HTTPDetails) { d.Checks = make([]observation.Check, 18) },
	} {
		d := valid
		mutate(&d)
		if _, e := makeObservation(0, "", d, c); e == nil {
			t.Fatal("invalid observation accepted")
		}
	}
}

type timeoutError struct{}

func (timeoutError) Error() string { return "synthetic-raw-error-canary" }
func (timeoutError) Timeout() bool { return true }
func TestTypedTimeoutNormalization(t *testing.T) {
	for _, err := range []error{timeoutError{}, fmt.Errorf("wrapped: %w", timeoutError{}), context.DeadlineExceeded} {
		if ioFailure(err) != "timeout" {
			t.Fatal("typed timeout lost")
		}
	}
	c := config()
	c.URL = "http://measurement.example/check"
	c.PostBody = []byte("dns")
	if _, e := validate(c); e == nil {
		t.Fatal("cleartext DNS POST accepted")
	}
}

func TestNormalizedURLByteBound(t *testing.T) {
	c := config()
	prefix := "https://measurement.example/"
	c.URL = prefix + strings.Repeat("x", 2048-len(prefix))
	if _, e := validate(c); e != nil {
		t.Fatal("exact URL bound rejected")
	}
	c.URL += "x"
	if _, e := validate(c); e == nil {
		t.Fatal("oversized URL accepted")
	}
	c.URL = prefix + strings.Repeat("é", 500)
	if _, e := validate(c); e == nil {
		t.Fatal("escaped URL exceeded normalized bound")
	}
}
