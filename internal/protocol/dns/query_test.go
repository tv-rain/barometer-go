package dns_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/protocol/dns"
	"github.com/tv-rain/barometer-go/internal/testkit"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

type packet struct {
	kind   string
	source bool
}
type script struct {
	packets       []packet
	query         []byte
	writes, reads int
	closes        atomic.Int32
	clock         *testkit.Clock
}

func (c *script) Write(b []byte) (int, error) {
	c.writes++
	c.query = append([]byte(nil), b...)
	return len(b), nil
}
func (c *script) Read(b []byte) (int, error) { n, _, _, e := c.ReadDatagram(b); return n, e }
func (c *script) ReadDatagram(b []byte) (int, bool, bool, error) {
	c.reads++
	if len(c.packets) == 0 {
		_ = c.clock.Advance(15 * time.Second)
		return 0, false, true, errors.New("synthetic closed")
	}
	p := c.packets[0]
	c.packets = c.packets[1:]
	r := loopback.DNSResponse(c.query, 0, []loopback.DNSRecord{{Type: "A", Value: "1.1.1.1"}})
	switch p.kind {
	case "mismatch":
		r[0] ^= 1
	case "malformed":
		r = r[:len(r)-1]
	case "foreign name":
		r[13] = 'z'
	}
	return copy(b, r), false, p.source, nil
}
func (c *script) SetDeadline(time.Time) error { return nil }
func (c *script) Close() error                { c.closes.Add(1); return nil }

// requirement-ref: REQ-DNS-01-0010
// requirement-ref: REQ-DNS-01-0013
// requirement-ref: REQ-DNS-01-0038
func TestQueryCorrelationAndCollectionPrecedence(t *testing.T) {
	for _, tc := range []struct {
		name    string
		packets []packet
		collect bool
		want    string
	}{
		{"query mismatch", []packet{{"mismatch", true}}, false, "query_mismatch"},
		{"wrong question", []packet{{"foreign name", true}}, false, "query_mismatch"},
		{"unexpected source", []packet{{"valid", false}}, false, "unexpected_source"},
		{"malformed", []packet{{"malformed", true}}, false, "malformed_response"},
		{"later usable", []packet{{"mismatch", true}, {"valid", true}}, false, "resolved"},
		{"later approved source", []packet{{"valid", false}, {"valid", true}}, false, "resolved"},
		{"unrelated source", []packet{{"mismatch", false}}, false, "timeout"},
		{"identical answers", []packet{{"valid", true}, {"valid", true}}, true, "resolved"},
		{"count takes precedence", []packet{{"valid", true}, {"valid", true}, {"valid", true}}, true, "response_limit_reached"},
		{"malformed after usable", []packet{{"valid", true}, {"malformed", true}}, true, "resolved"},
		{"no result", nil, false, "timeout"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clock := testkit.NewClock(time.Now())
			c := &script{packets: tc.packets, clock: clock}
			calls := 0
			in := dns.Input{Clock: clock, Dial: func(context.Context, string, string) (dns.Conn, error) { calls++; return c, nil }}
			conn, out, e := dns.Direct(context.Background(), in, question, "udp", "synthetic:53", "ipv4", tc.collect, 500*time.Millisecond)
			if e != nil || out.Outcome != tc.want {
				t.Fatalf("%s %v", out.Outcome, e)
			}
			if conn != nil {
				conn.Close()
			}
			if calls != 1 || c.writes != 1 || clock.Timers() != 0 {
				t.Fatal("retry or timer leak")
			}
		})
	}
}
