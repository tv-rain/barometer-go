package http

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/observation"
	"github.com/tv-rain/barometer-go/internal/testkit"
)

// http01StalledBodyConn delivers the response head, then reports a body read
// that makes no progress and returns no error. The production HTTP-01 runner
// always reads through crypto/tls, whose Read never returns (0, nil) for a
// non-empty buffer, so this inconclusive body read is reachable only at the
// exchange layer that the runner delegates terminal classification to.
type http01StalledBodyConn struct {
	head   string
	sent   strings.Builder
	reads  int
	closed bool
}

func (c *http01StalledBodyConn) Read(p []byte) (int, error) {
	c.reads++
	if c.head == "" {
		return 0, nil
	}
	n := copy(p, c.head)
	c.head = c.head[n:]
	return n, nil
}
func (c *http01StalledBodyConn) Write(p []byte) (int, error) { return c.sent.Write(p) }
func (c *http01StalledBodyConn) Close() error                { c.closed = true; return nil }

type http01StalledEndpoint struct {
	conn   *http01StalledBodyConn
	opened int
}

func (e *http01StalledEndpoint) Open(context.Context, *url.URL) (Stream, []observation.Observation, string, error) {
	e.opened++
	return Stream{e.conn, "ipv4"}, nil, "", nil
}
func (e *http01StalledEndpoint) Redirect(*url.URL) (Endpoint, string, error) {
	return nil, "unregistered_https", ErrInterrupted
}

// An inconclusive body read takes precedence over an already evaluated status
// mismatch, keeps only the checks decided before termination, invents no body
// check, performs no further read or connection, and closes the stream.
// requirement-ref: REQ-HTTP-01-0036
// requirement-ref: REQ-HTTP-01-0037
// requirement-ref: REQ-HTTP-01-0052
func TestMHttp01NormalizationInconclusive(t *testing.T) {
	conn := &http01StalledBodyConn{head: "HTTP/1.1 200 OK\r\nContent-Length: 10\r\nX-Fixture: header-canary\r\n\r\n"}
	e := &http01StalledEndpoint{conn: conn}
	c := Config{URL: "https://measurement.example/static", MaxBodyBytes: 4096, ExpectedStatus: 204,
		HeaderChecks: []HeaderCheck{{"x-fixture", "present", ""}}, BodyChecks: []BodyCheck{{"contains", "body-canary"}}}
	r, cleanup, err := Exchange(context.Background(), testkit.NewClock(time.Now()), e, c)
	if err != nil {
		t.Fatal(err)
	}
	reads := conn.reads
	cleanup()
	if r.Outcome != "inconclusive" || len(r.Observations) != 1 || r.Observations[0].Failure() != "inconclusive" {
		t.Fatal(r.Outcome, r.Observations)
	}
	d := terminal(t, r)
	if len(d.Checks) != 2 || d.Checks[0] != (observation.Check{Kind: "status", Matched: false}) || d.Checks[1] != (observation.Check{Kind: "header", Matched: true}) {
		t.Fatal("decided checks", d.Checks)
	}
	if d.Body == nil || d.Body.BytesProcessed != 0 || d.Body.Truncated {
		t.Fatal("body metadata", d.Body)
	}
	if e.opened != 1 || conn.reads != reads || !conn.closed {
		t.Fatal("retry, extra read or open stream")
	}
	b, _ := json.Marshal(r)
	if strings.Contains(string(b), "header-canary") || strings.Contains(string(b), "body-canary") {
		t.Fatal("retained response data")
	}
}
