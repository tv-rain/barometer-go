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
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/config"
	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
	"syscall"
)

var dnsTest = task.TestVersion{ID: "DNS-01", Version: "0.1.0"}

const dnsInput = `{"hostname":"example.org","record_type":"A","resolver":{"ip_address":"8.8.8.8"}}`

func dnsTask(id, mode string) string {
	data := strings.Replace(runTask(id, "", "", false), `"TCP-01"`, `"DNS-01"`, 1)
	input := dnsInput
	options := `{"resolver_mode":"direct","transport":"` + mode + `"}`
	if mode == "system" {
		input = `{"hostname":"example.org","record_type":"A"}`
		options = `{}`
	}
	return strings.Replace(strings.Replace(data, `"input":{}`, `"input":`+input, 1), `"options":{}`, `"options":`+options, 1)
}

type dnsTransport struct {
	calls atomic.Int32
	reads atomic.Int32
	dial  func(context.Context, string) (measurement.Conn, error)
	// beforeRead, when set, runs before each datagram read with the count of
	// datagrams already read; tests use it to advance the clock between replies.
	beforeRead func(read int32)
}

func (d *dnsTransport) Dialer(int) (measurement.Dialer, error) { return d, nil }
func (d *dnsTransport) DialContext(ctx context.Context, network, address string) (measurement.Conn, error) {
	d.calls.Add(1)
	if address != "8.8.8.8:53" && address != "1.1.1.1:53" || network != "udp" && network != "tcp" {
		return nil, errors.New("synthetic unexpected destination")
	}
	c, e := d.dial(ctx, network)
	if c != nil {
		return dnsReadConn{c, &d.reads, d.beforeRead}, e
	}
	return c, e
}
func dnsHarness(t *testing.T, mode string) (*harness, *dnsTransport) {
	t.Helper()
	h := runningHarness(t)
	h.adapters.registry = nil
	h.adapters.run.paths = nil
	p := synPolicyInput(2, "reg-a", func(p *task.Permissions) {
		p.Tests = []task.TestVersion{dnsTest}
		p.RegistrationIDs = []string{"reg-a"}
		if mode != "system" {
			p.RegistrationIDs = append(p.RegistrationIDs, "resolver-a")
		}
		p.AllowUnknownRisk = true
	})
	p.Targets[0].Test = dnsTest
	p.Targets[0].Tuple = task.TargetTuple{Hostname: "example.org"}
	if mode != "system" {
		r := p.Targets[0]
		r.ID = "resolver-a"
		r.Role = task.ResolverTarget
		r.Tuple = task.TargetTuple{Address: netip.MustParseAddr("8.8.8.8"), Port: 53, Scheme: mode}
		p.Targets = append(p.Targets, r)
	}
	var e error
	h.adapters.policy, e = task.NewLocalPolicy(p)
	if e != nil {
		t.Fatal(e)
	}
	d := &dnsTransport{dial: func(context.Context, string) (measurement.Conn, error) { return nil, errors.New("SYNTHETIC-RAW-ERROR") }}
	h.adapters.run.directDialers = d
	return h, d
}

type dnsEnvelope struct {
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
			Duration              int `json:"duration_ms"`
			Details               struct {
				Engine        string
				QueryType     string `json:"query_type"`
				AddressFamily string `json:"address_family"`
				Responses     []struct {
					Index   int `json:"arrival_index"`
					Offset  int `json:"arrival_offset_ms"`
					RCode   string
					Answers []struct{ Type, Value string }
				}
				Limit bool `json:"response_limit_reached"`
			} `json:"details"`
		}
	}
}

func dnsResult(t *testing.T, raw []byte, state, reason, outcome string) dnsEnvelope {
	t.Helper()
	var r dnsEnvelope
	if e := json.Unmarshal(raw, &r); e != nil {
		t.Fatal(e)
	}
	if r.Execution.State != state || r.Execution.Reason != reason {
		t.Fatalf("state %s", raw)
	}
	if outcome == "" {
		if r.Measurement != nil {
			t.Fatalf("interrupted measurement: %s", raw)
		}
		return r
	}
	if r.Schema != "0.1.0" || r.Test.ID != "DNS-01" || r.Test.Version != "0.1.0" || r.Test.ResultVersion != "0.1.0" || r.Measurement == nil || r.Measurement.Outcome != outcome || len(r.Measurement.Observations) != 1 || r.Measurement.Observations[0].Type != "dns_query" {
		t.Fatalf("measurement %s", raw)
	}
	for _, canary := range []string{"example.org", "8.8.8.8", "SYNTHETIC-RAW-ERROR", `"hostname"`, `"ip_address"`, `"ttl"`, `"query_id"`, `"port"`} {
		if strings.Contains(string(raw), canary) {
			t.Fatalf("retained %s: %s", canary, raw)
		}
	}
	return r
}
func dnsLoopback(t *testing.T, d *dnsTransport, reply func([]byte) [][]byte) *loopback.DNSServer {
	t.Helper()
	s, e := loopback.ListenDNS(reply)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { s.Close() })
	d.dial = func(ctx context.Context, network string) (measurement.Conn, error) {
		var nd net.Dialer
		return nd.DialContext(ctx, network, s.Address())
	}
	return s
}
func oneAnswer(q []byte) [][]byte {
	return [][]byte{loopback.DNSResponse(q, 0, []loopback.DNSRecord{{Type: "A", Value: "1.1.1.1"}})}
}

// requirement-ref: REQ-DNS-01-0014
// requirement-ref: REQ-DNS-01-0016
// requirement-ref: REQ-DNS-01-0017
// requirement-ref: REQ-DNS-01-0019
// requirement-ref: REQ-DNS-01-0020
// requirement-ref: REQ-DNS-01-0021
// requirement-ref: REQ-DNS-01-0023
// requirement-ref: REQ-DNS-01-0024
func TestMDns01Admission(t *testing.T) {
	for _, input := range []string{`{}`, `null`, `[]`, `{"hostname":"example.org","record_type":"MX","resolver":{"ip_address":"8.8.8.8"}}`, `{"hostname":"example.org","record_type":"A"}`, `{"hostname":"example.org","record_type":"A","resolver":{"ip_address":"127.0.0.1"}}`, `{"hostname":"example.org","record_type":"A","resolver":{"ip_address":"8.8.8.8","port":0}}`, `{"hostname":"example.org","record_type":"A","resolver":{"ip_address":"8.8.8.8","port":65536}}`, `{"hostname":"example.org","record_type":"A","resolver":{"ip_address":"::ffff:8.8.8.8"}}`, dnsInput[:len(dnsInput)-1] + `,"unknown":0}`} {
		h, d := dnsHarness(t, "udp")
		c := h.initialized(t)
		data := strings.Replace(dnsTask("invalid", "udp"), dnsInput, input, 1)
		handle, e := c.AdmitTask(data, h.embedded(t, data))
		if handle != "" || codeOf(e) != "task_schema_invalid" || h.replay.Commits() != 0 || d.calls.Load() != 0 || len(h.run.output.Records()) != 0 {
			t.Fatalf("%s: %s", input, codeOf(e))
		}
		c.Close()
	}
	// Resolver addresses: canonical public literals only, as for TCP-01.
	for _, ip := range []string{"8.8.8.08", " 8.8.8.8", "0.0.0.0", "10.0.0.1", "100.64.0.1", "169.254.1.1", "192.0.2.1", "198.18.0.1", "224.0.0.1", "240.0.0.1", "::", "::1", "fe80::1", "fc00::1", "ff02::1", "2001:db8::1", "2606:4700:4700:0:0:0:0:1111", "2606:4700:4700::ABCD", "fe80::1%en0", "dns.example.org"} {
		h, d := dnsHarness(t, "udp")
		c := h.initialized(t)
		data := strings.Replace(dnsTask("resolver", "udp"), `"ip_address":"8.8.8.8"`, fmt.Sprintf(`"ip_address":%q`, ip), 1)
		handle, e := c.AdmitTask(data, h.embedded(t, data))
		if handle != "" || codeOf(e) != "task_schema_invalid" || h.replay.Commits() != 0 || d.calls.Load() != 0 {
			t.Fatalf("resolver %q: %s", ip, codeOf(e))
		}
		c.Close()
	}
	for _, name := range []string{".", "single", "a.local", "a.localhost", "a.home.arpa", "a.internal", "127.0.0.1", "https://example.org", "*.example.org", "a..org", "-a.org", "a-.org", strings.Repeat("a", 64) + ".org", strings.Repeat("a", 254) + ".org", "a.org.", "a\u200d.org"} {
		h, d := dnsHarness(t, "udp")
		c := h.initialized(t)
		data := strings.Replace(dnsTask("name", "udp"), `example.org`, name, 1)
		handle, e := c.AdmitTask(data, h.embedded(t, data))
		if handle != "" || codeOf(e) != "task_schema_invalid" || h.replay.Commits() != 0 || d.calls.Load() != 0 || len(h.run.output.Records()) != 0 {
			t.Fatalf("name %q: %s", name, codeOf(e))
		}
		c.Close()
	}
	for _, tc := range []struct{ from, to string }{{`"ip_address":"8.8.8.8"`, `"ip_address":"1.1.1.1"`}, {`"ip_address":"8.8.8.8"`, `"ip_address":"8.8.8.8","port":54`}, {`example.org`, `other.example.org`}} {
		h, d := dnsHarness(t, "udp")
		c := h.initialized(t)
		data := strings.Replace(dnsTask("registration", "udp"), tc.from, tc.to, 1)
		_, e := c.AdmitTask(data, h.embedded(t, data))
		if codeOf(e) != "target_not_registered" || d.calls.Load() != 0 {
			t.Fatal(codeOf(e))
		}
		c.Close()
	}
}

// requirement-ref: REQ-DNS-01-0025
// requirement-ref: REQ-DNS-01-0026
// requirement-ref: REQ-LIMIT-0014
func TestMDns01Options(t *testing.T) {
	for _, options := range []string{`{}`, `{"resolver_mode":"wrong"}`, `{"resolver_mode":"direct","transport":"tls"}`, `{"resolver_mode":"direct","collect_additional_responses":true}`, `{"resolver_mode":"direct","collect_additional_responses":true,"additional_response_window_ms":0}`, `{"resolver_mode":"direct","collect_additional_responses":true,"additional_response_window_ms":1001}`, `{"resolver_mode":"direct","collect_additional_responses":true,"additional_response_window_ms":10001}`, `{"resolver_mode":"direct","transport":"tcp","collect_additional_responses":true,"additional_response_window_ms":100}`, `{"resolver_mode":"direct","additional_response_window_ms":100}`, `{"resolver_mode":"direct","collect_additional_responses":null}`, `{"resolver_mode":"direct","unexpected":1}`,
		`{"resolver_mode":1}`, `{"resolver_mode":null}`, `{"resolver_mode":"direct","transport":1}`, `{"resolver_mode":"direct","transport":null}`, `{"resolver_mode":"DIRECT"}`,
		`{"resolver_mode":"direct","collect_additional_responses":"true"}`, `{"resolver_mode":"direct","collect_additional_responses":1}`,
		`{"resolver_mode":"direct","collect_additional_responses":true,"additional_response_window_ms":"100"}`, `{"resolver_mode":"direct","collect_additional_responses":true,"additional_response_window_ms":1.5}`,
		`{"resolver_mode":"direct","collect_additional_responses":true,"additional_response_window_ms":null}`, `{"resolver_mode":"direct","collect_additional_responses":false,"additional_response_window_ms":100}`,
		`{"resolver_mode":"system","transport":"udp"}`, `{"resolver_mode":"system","collect_additional_responses":true,"additional_response_window_ms":100}`, `{"transport":"udp"}`, `[]`, `null`} {
		h, d := dnsHarness(t, "udp")
		c := h.initialized(t)
		data := strings.Replace(dnsTask("options", "udp"), `{"resolver_mode":"direct","transport":"udp"}`, options, 1)
		handle, e := c.AdmitTask(data, h.embedded(t, data))
		want := "task_schema_invalid"
		if strings.Contains(options, ":1001}") || strings.Contains(options, ":10001}") {
			want = "task_limit_exceeded"
		}
		if handle != "" || codeOf(e) != want || d.calls.Load() != 0 || h.replay.Commits() != 0 {
			t.Fatalf("%s: %s", options, codeOf(e))
		}
		c.Close()
	}
	// The window is 1..1,000 ms: REQ-LIMIT-0014 is stricter than the test's
	// option table and controls admission (see the M-DNS-01-OPTIONS plan).
	for _, options := range []string{`{"resolver_mode":"direct","collect_additional_responses":true,"additional_response_window_ms":1}`, `{"resolver_mode":"direct","collect_additional_responses":true,"additional_response_window_ms":1000}`, `{"resolver_mode":"direct","collect_additional_responses":false}`, `{"resolver_mode":"direct"}`} {
		h, d := dnsHarness(t, "udp")
		c := h.initialized(t)
		data := strings.Replace(dnsTask("window", "udp"), `{"resolver_mode":"direct","transport":"udp"}`, options, 1)
		if _, e := c.AdmitTask(data, h.embedded(t, data)); e != nil || d.calls.Load() != 0 {
			t.Fatal(options, codeOf(e))
		}
		c.Close()
	}
	for _, mode := range []string{"system", "udp", "tcp"} {
		h, _ := dnsHarness(t, mode)
		c := h.initialized(t)
		data := dnsTask("valid", mode)
		_, e := c.AdmitTask(data, h.embedded(t, data))
		if e != nil {
			t.Fatal(mode, codeOf(e))
		}
		c.Close()
	}
}

// requirement-ref: REQ-DNS-01-0008
// requirement-ref: REQ-DNS-01-0009
// requirement-ref: REQ-DNS-01-0034
// requirement-ref: REQ-DNS-01-0035
// requirement-ref: REQ-DNS-01-0036
// requirement-ref: REQ-DNS-01-0040
func TestMDns01ExecutionAndOutput(t *testing.T) {
	for _, mode := range []string{"udp", "tcp"} {
		t.Run(mode, func(t *testing.T) {
			h, d := dnsHarness(t, mode)
			s := dnsLoopback(t, d, oneAnswer)
			c := h.initialized(t)
			defer c.Close()
			raw := tcpRaw(t, h, c, dnsTask("query", mode))
			r := dnsResult(t, raw, "completed", "", "resolved")
			o := r.Measurement.Observations[0]
			if o.Details.Engine != mode || len(o.Details.Responses) != 1 || len(o.Details.Responses[0].Answers) != 1 || o.Details.Responses[0].Answers[0].Value != "1.1.1.1" || o.Details.Responses[0].Index != 1 || o.Duration%250 != 0 {
				t.Fatalf("%s", raw)
			}
			queries, _, _ := s.Counts()
			if queries != 1 || d.calls.Load() != 1 {
				t.Fatal("retry or missing query")
			}
			if _, e := c.NextBundle(); e != nil {
				t.Fatal(e)
			}
		})
	}
}

// requirement-ref: REQ-DNS-01-0011
// requirement-ref: REQ-DNS-01-0012
// requirement-ref: REQ-DNS-01-0013
// requirement-ref: REQ-DNS-01-0037
// requirement-ref: REQ-DNS-01-0038
func TestMDns01Normalization(t *testing.T) {
	for _, tc := range []struct {
		code    uint16
		outcome string
	}{{0, "no_data"}, {2, "servfail"}, {3, "nxdomain"}, {5, "refused"}, {1, "inconclusive"}, {4, "inconclusive"}} {
		t.Run(tc.outcome, func(t *testing.T) {
			h, d := dnsHarness(t, "udp")
			dnsLoopback(t, d, func(q []byte) [][]byte { return [][]byte{loopback.DNSResponse(q, tc.code, nil)} })
			c := h.initialized(t)
			defer c.Close()
			dnsResult(t, tcpRaw(t, h, c, dnsTask("rcode", "udp")), "completed", "", tc.outcome)
		})
	}
	for _, count := range []int{2, 3} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			h, d := dnsHarness(t, "udp")
			dnsLoopback(t, d, func(q []byte) [][]byte {
				a := oneAnswer(q)[0]
				b := loopback.DNSResponse(q, 0, []loopback.DNSRecord{{Type: "A", Value: "9.9.9.9"}})
				out := [][]byte{a, b}
				if count == 3 {
					out = append(out, b)
				}
				return out
			})
			c := h.initialized(t)
			defer c.Close()
			data := strings.Replace(dnsTask("collect", "udp"), `"transport":"udp"`, `"transport":"udp","collect_additional_responses":true,"additional_response_window_ms":500`, 1)
			// The manual clock advances only after both responses have been read.
			rh, e := c.StartRun(h.admit(t, c, data))
			if e != nil {
				t.Fatal(codeOf(e))
			}
			if count == 2 {
				waitDNSReads(t, d, h.run.clock, 500*time.Millisecond)
			}
			waitTerminal(t, c, rh)
			outcome := "conflicting_responses"
			if count == 3 {
				outcome = "response_limit_reached"
			}
			r := dnsResult(t, h.run.output.Records()[0], "completed", "", outcome)
			if len(r.Measurement.Observations[0].Details.Responses) != count || r.Measurement.Observations[0].Details.Limit != (count == 3) {
				t.Fatal("collection bounds")
			}
		})
	}
}

type dnsReadConn struct {
	measurement.Conn
	reads      *atomic.Int32
	beforeRead func(int32)
}

func (c dnsReadConn) Read(p []byte) (int, error) {
	n, e := c.Conn.Read(p)
	if n > 0 {
		c.reads.Add(1)
	}
	return n, e
}
func (c dnsReadConn) ReadDatagram(p []byte) (int, bool, bool, error) {
	if c.beforeRead != nil {
		c.beforeRead(c.reads.Load())
	}
	if u, ok := c.Conn.(*net.UDPConn); ok {
		n, _, flags, _, e := u.ReadMsgUDP(p, nil)
		if n > 0 {
			c.reads.Add(1)
		}
		return max(0, n), flags&syscall.MSG_TRUNC != 0, true, e
	}
	n, e := c.Read(p)
	return n, false, true, e
}
func waitDNSReads(t *testing.T, d *dnsTransport, clock *testkit.WaitClock, window time.Duration) {
	t.Helper()
	for deadline := time.Now().Add(3 * time.Second); time.Now().Before(deadline); time.Sleep(time.Millisecond) {
		if d.reads.Load() >= 2 && clock.Timers() >= 3 {
			_ = clock.Advance(window)
			return
		}
	}
	t.Fatal("responses or observation timer missing")
}

// DNS rows of CC-LIMIT-PURPOSE: exact DNS messages and TCP length prefixes,
// including ignored datagrams, count at the metered transport boundary.
// requirement-ref: REQ-LIMIT-0034
// requirement-ref: REQ-DNS-01-0033
func TestCcLimitPurposeDNSRows(t *testing.T) {
	for _, mode := range []string{"udp", "tcp"} {
		for _, direction := range []string{"sent", "received"} {
			for _, delta := range []int{-1, 0, 1} {
				t.Run(fmt.Sprintf("%s/%s/%d", mode, direction, delta), func(t *testing.T) {
					h, d := dnsHarness(t, mode)
					dnsLoopback(t, d, oneAnswer)
					c := h.initialized(t)
					defer c.Close()
					size := 29
					if direction == "received" {
						size = 45
					}
					if mode == "tcp" {
						size += 2
					}
					data := strings.Replace(dnsTask("budget", mode), `"limits":{}`, fmt.Sprintf(`"limits":{"max_bytes_%s":%d}`, direction, size+delta), 1)
					state, reason, outcome := "completed", "", "resolved"
					if delta < 0 {
						state, reason, outcome = "incomplete", "resource_limit", ""
					}
					dnsResult(t, tcpRaw(t, h, c, data), state, reason, outcome)
					if d.calls.Load() != 1 {
						t.Fatal("retry")
					}
				})
			}
		}
	}
}

// Connection rows for DNS-01 through Client: the UDP socket or TCP connection
// to the resolver is the attempt's one metered connection. At the schema floor
// of max_connections (1) and above it exactly one reaches the transport; the
// meter's refusal below a limit is shared with TestTcp01ConnectionMeter.
// requirement-ref: REQ-LIMIT-0006
// requirement-ref: REQ-DNS-01-0033
func TestMDns01ConnectionBudget(t *testing.T) {
	for _, mode := range []string{"udp", "tcp"} {
		for _, limit := range []int{1, 2} {
			t.Run(fmt.Sprintf("%s/%d", mode, limit), func(t *testing.T) {
				h, d := dnsHarness(t, mode)
				dnsLoopback(t, d, oneAnswer)
				c := h.initialized(t)
				defer c.Close()
				data := strings.Replace(dnsTask("connections", mode), `"limits":{}`, fmt.Sprintf(`"limits":{"max_connections":%d}`, limit), 1)
				dnsResult(t, tcpRaw(t, h, c, data), "completed", "", "resolved")
				if d.calls.Load() != 1 {
					t.Fatalf("transport dials %d, want 1", d.calls.Load())
				}
			})
		}
	}
}

// DNS row of CC-PROXY-OVERVIEW: system and UDP never enter a CONNECT path;
// TCP uses one tunnel with both DNS length prefixes inside its byte meter.
// requirement-ref: REQ-PROXY-0044
// requirement-ref: REQ-LIMIT-0034
func TestCcProxyOverviewDNSRow(t *testing.T) {
	for _, mode := range []string{"system", "udp", "tcp"} {
		t.Run(mode, func(t *testing.T) {
			h, d := dnsHarness(t, mode)
			s := dnsLoopback(t, d, oneAnswer)
			proxy, e := loopback.ListenProxy(loopback.ProxyConfig{Resolve: map[string]string{"8.8.8.8:53": s.Address()}})
			if e != nil {
				t.Fatal(e)
			}
			defer proxy.Close()
			h.adapters.approvals = config.Approvals{ConnectionCodes: []string{"outline_mobileproxy"}}
			c := h.client(t)
			if e := c.Initialize(proxyConfig(proxy.Port())); e != nil {
				t.Fatal(codeOf(e))
			}
			defer c.Close()
			data := strings.Replace(dnsTask("proxy", mode), `"limits":{}`, `"limits":{"max_bytes_sent":31,"max_bytes_received":47},"network_path":"proxy"`, 1)
			handle, e := c.AdmitTask(data, h.embedded(t, data))
			if mode != "tcp" {
				if codeOf(e) != "task_schema_invalid" || handle != "" || d.calls.Load() != 0 || proxy.Counts().Accepted != 0 {
					t.Fatal(codeOf(e), "incompatible path contacted")
				}
				return
			}
			if e != nil {
				t.Fatal(codeOf(e))
			}
			rh, e := c.StartRun(handle)
			if e != nil {
				t.Fatal(codeOf(e))
			}
			waitTerminal(t, c, rh)
			dnsResult(t, h.run.output.Records()[0], "completed", "", "resolved")
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if e := proxy.WaitIdle(ctx); e != nil {
				t.Fatal(e)
			}
			counts := proxy.Counts()
			if counts.Accepted != 1 || counts.TunnelIn != 31 || counts.TunnelOut != 47 || d.calls.Load() != 0 {
				t.Fatalf("proxy meter %+v", counts)
			}
		})
	}
}

// requirement-ref: REQ-DNS-01-0016
func TestDns01NormalizedNamesAndResolverBoundaries(t *testing.T) {
	long := strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", 61)
	for _, tc := range []struct{ input, normalized string }{{"EXAMPLE.ORG", "example.org"}, {"bücher.example.org", "xn--bcher-kva.example.org"}, {long, long}} {
		h, d := dnsHarness(t, "udp")
		p := synPolicyInput(2, "reg-a", func(p *task.Permissions) { p.Tests = []task.TestVersion{dnsTest} })
		p.Targets[0].Test = dnsTest
		p.Targets[0].Tuple = task.TargetTuple{Hostname: tc.normalized}
		h.adapters.policy, _ = task.NewLocalPolicy(p)
		// System mode needs only the reviewed query-name registration.
		c := h.initialized(t)
		data := strings.Replace(dnsTask("normalized", "system"), "example.org", tc.input, 1)
		if _, e := c.AdmitTask(data, h.embedded(t, data)); e != nil {
			t.Fatalf("%s: %s", tc.input, codeOf(e))
		}
		if d.calls.Load() != 0 {
			t.Fatal("admission traffic")
		}
		c.Close()
	}
	for _, ip := range []string{"8.8.8.8", "2606:4700:4700::1111"} {
		for _, port := range []int{1, 65535} {
			h, _ := dnsHarness(t, "udp")
			p := synPolicyInput(2, "reg-a", func(p *task.Permissions) {
				p.Tests = []task.TestVersion{dnsTest}
				p.RegistrationIDs = []string{"reg-a", "resolver-a"}
			})
			p.Targets[0].Test = dnsTest
			p.Targets[0].Tuple = task.TargetTuple{Hostname: "example.org"}
			r := p.Targets[0]
			r.ID = "resolver-a"
			r.Role = task.ResolverTarget
			r.Tuple = task.TargetTuple{Address: netip.MustParseAddr(ip), Port: uint16(port), Scheme: "udp"}
			p.Targets = append(p.Targets, r)
			h.adapters.policy, _ = task.NewLocalPolicy(p)
			c := h.initialized(t)
			data := strings.Replace(dnsTask("port", "udp"), `"ip_address":"8.8.8.8"`, fmt.Sprintf(`"ip_address":%q,"port":%d`, ip, port), 1)
			if _, e := c.AdmitTask(data, h.embedded(t, data)); e != nil {
				t.Fatal(codeOf(e))
			}
			c.Close()
		}
	}
}

// requirement-ref: REQ-DNS-01-0034
func TestMDns01Versions(t *testing.T) {
	// "missing" removes the member. Every variant is a task-validation
	// failure: no handle, replay record, contact or Measurement Result.
	for _, version := range []string{`"0.1.1"`, `"0.2.0"`, `"1.0.0"`, `"0.1"`, `"v0.1.0"`, `"0.1.0-draft"`, `""`, `null`, `1`, "missing"} {
		h, d := dnsHarness(t, "udp")
		c := h.initialized(t)
		replacement := `"test_version":` + version + `,`
		if version == "missing" {
			replacement = ""
		}
		data := strings.Replace(dnsTask("version", "udp"), `"test_version":"0.1.0",`, replacement, 1)
		handle, e := c.AdmitTask(data, h.embedded(t, data))
		// A well-formed unregistered version is task_unsupported (REQ-TRUST-0047
		// step 4); a malformed or missing one fails the envelope schema.
		want := "task_schema_invalid"
		if version == `"0.1.1"` || version == `"0.2.0"` || version == `"1.0.0"` {
			want = "task_unsupported"
		}
		if code := codeOf(e); code != want || handle != "" || h.replay.Commits() != 0 || d.calls.Load() != 0 || len(h.run.output.Records()) != 0 {
			t.Fatalf("version %s: %s", version, codeOf(e))
		}
		c.Close()
	}
}

type dnsSystemResolver struct {
	lookup func(context.Context, string, string) ([]netip.Addr, error)
}

func (r dnsSystemResolver) LookupNetIP(ctx context.Context, family, name string) ([]netip.Addr, error) {
	return r.lookup(ctx, family, name)
}

// requirement-ref: REQ-DNS-01-0007
// requirement-ref: REQ-DNS-01-0009
// requirement-ref: REQ-DNS-01-0035
// requirement-ref: REQ-LIMIT-0034
func TestDns01SystemAPIOutputAndHiddenBytes(t *testing.T) {
	for _, tc := range []struct {
		ips     []netip.Addr
		err     error
		outcome string
	}{{[]netip.Addr{netip.MustParseAddr("1.1.1.1")}, nil, "resolved"}, {nil, nil, "no_data"}, {nil, &net.DNSError{IsNotFound: true, Err: "SYNTHETIC-RAW-ERROR"}, "inconclusive"}, {nil, &net.DNSError{IsTimeout: true}, "timeout"}, {nil, errors.New("SYNTHETIC-RAW-ERROR"), "resolver_error"}} {
		t.Run(tc.outcome, func(t *testing.T) {
			h, d := dnsHarness(t, "system")
			calls := 0
			h.adapters.run.resolver = dnsSystemResolver{func(_ context.Context, family, name string) ([]netip.Addr, error) {
				calls++
				if family != "ip4" || name != "example.org." {
					t.Error("unapproved query")
				}
				return tc.ips, tc.err
			}}
			c := h.initialized(t)
			defer c.Close()
			data := strings.Replace(dnsTask("system", "system"), `"limits":{}`, `"limits":{"max_bytes_sent":1,"max_bytes_received":1}`, 1)
			raw := tcpRaw(t, h, c, data)
			r := dnsResult(t, raw, "completed", "", tc.outcome)
			details := r.Measurement.Observations[0].Details
			if details.Engine != "system" || len(details.Responses) != 0 || strings.Contains(string(raw), "arrival_") || strings.Contains(string(raw), "rcode") || strings.Contains(string(raw), "address_family") || d.calls.Load() != 0 || calls != 1 {
				t.Fatalf("invented system metadata or network fallback %s", raw)
			}
		})
	}
}

// requirement-ref: REQ-DNS-01-0039
func TestMDns01Termination(t *testing.T) {
	for _, mode := range []string{"udp", "tcp", "system"} {
		for _, reason := range []string{"network_changed", "cancelled_by_host", "attempt_deadline_exceeded"} {
			t.Run(mode+"/"+reason, func(t *testing.T) {
				h, d := dnsHarness(t, mode)
				entered := make(chan struct{})
				if mode == "system" {
					h.adapters.run.resolver = dnsSystemResolver{func(ctx context.Context, _, _ string) ([]netip.Addr, error) {
						close(entered)
						<-ctx.Done()
						return nil, ctx.Err()
					}}
				} else {
					d.dial = func(ctx context.Context, _ string) (measurement.Conn, error) {
						close(entered)
						<-ctx.Done()
						return nil, ctx.Err()
					}
				}
				c := h.initialized(t)
				defer c.Close()
				data := strings.Replace(dnsTask("terminate", mode), `"limits":{}`, `"limits":{"max_duration_ms":1000}`, 1)
				rh, e := c.StartRun(h.admit(t, c, data))
				if e != nil {
					t.Fatal(codeOf(e))
				}
				select {
				case <-entered:
				case <-time.After(time.Second):
					t.Fatal("no operation")
				}
				switch reason {
				case "network_changed":
					a := defaultAccess
					a.Generation++
					h.run.network.Set(a)
				case "cancelled_by_host":
					if e := c.CancelRun(rh); e != nil {
						t.Fatal(e)
					}
				case "attempt_deadline_exceeded":
					_ = h.run.clock.Advance(time.Second)
				}
				waitTerminal(t, c, rh)
				dnsResult(t, h.run.output.Records()[0], "incomplete", reason, "")
				if h.run.clock.Timers() != 0 {
					t.Fatal("timer leaked")
				}
				// No second query and no fallback between system and direct modes.
				if want := map[bool]int32{true: 0, false: 1}[mode == "system"]; d.calls.Load() != want {
					t.Fatalf("%d direct dials", d.calls.Load())
				}
			})
		}
	}
}

// requirement-ref: REQ-DNS-01-0033
func TestMDns01Limits(t *testing.T) {
	for _, mode := range []string{"udp", "tcp", "system"} {
		t.Run("operation timeout/"+mode, func(t *testing.T) {
			h, d := dnsHarness(t, mode)
			stop := func(ctx context.Context) { _ = h.run.clock.Advance(15 * time.Second); <-ctx.Done() }
			if mode == "system" {
				h.adapters.run.resolver = dnsSystemResolver{func(ctx context.Context, _, _ string) ([]netip.Addr, error) { stop(ctx); return nil, ctx.Err() }}
			} else {
				d.dial = func(ctx context.Context, _ string) (measurement.Conn, error) { stop(ctx); return nil, ctx.Err() }
			}
			c := h.initialized(t)
			defer c.Close()
			data := strings.Replace(dnsTask("operation", mode), `"limits":{}`, `"limits":{"max_duration_ms":60000}`, 1)
			dnsResult(t, tcpRaw(t, h, c, data), "completed", "", "timeout")
		})
	}
	for _, mode := range []string{"udp", "tcp"} {
		maxBytes := 4096
		if mode == "tcp" {
			maxBytes = 65535
		}
		for _, delta := range []int{-1, 0, 1} {
			if mode == "tcp" && delta == 1 {
				continue
			}
			t.Run(fmt.Sprintf("message/%s/%d", mode, delta), func(t *testing.T) {
				h, d := dnsHarness(t, mode)
				dnsLoopback(t, d, func(q []byte) [][]byte {
					b := oneAnswer(q)[0]
					size := maxBytes + delta
					n := size - len(b) - 11
					b[10] = 0
					b[11] = 1
					b = append(b, 0, 0, 41, 16, 0, 0, 0, 0, 0, byte(n>>8), byte(n))
					b = append(b, 0, 1, byte((n-4)>>8), byte(n-4))
					return [][]byte{append(b, make([]byte, n-4)...)}
				})
				c := h.initialized(t)
				defer c.Close()
				outcome := "resolved"
				if delta > 0 {
					outcome = "malformed_response"
				}
				dnsResult(t, tcpRaw(t, h, c, dnsTask("message", mode)), "completed", "", outcome)
			})
		}
	}
}

// The queued second tuple does no protocol operation before the first releases.
// requirement-ref: REQ-DNS-01-0033
func TestDns01EquivalentRunsSerialize(t *testing.T) {
	h, d := dnsHarness(t, "udp")
	entered := make(chan struct{}, 2)
	release := make(chan struct{})
	d.dial = func(ctx context.Context, _ string) (measurement.Conn, error) {
		entered <- struct{}{}
		select {
		case <-release:
		case <-ctx.Done():
		}
		return nil, errors.New("synthetic")
	}
	c := h.initialized(t)
	defer c.Close()
	one, e := c.StartRun(h.admit(t, c, dnsTask("first", "udp")))
	if e != nil {
		t.Fatal(e)
	}
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("no first operation")
	}
	two, e := c.StartRun(h.admit(t, c, dnsTask("second", "udp")))
	if e != nil {
		t.Fatal(e)
	}
	// A second contact while the first remains held violates the tuple bound.
	select {
	case <-entered:
		close(release)
		t.Fatal("equivalent operation overlapped")
	case <-time.After(20 * time.Millisecond):
	}
	close(release)
	waitTerminal(t, c, one)
	waitTerminal(t, c, two)
	if len(h.run.output.Records()) != 2 || d.calls.Load() != 2 {
		t.Fatal("serialized run lost or retried")
	}
	for _, raw := range h.run.output.Records() {
		dnsResult(t, raw, "completed", "", "resolver_error")
	}
}

// M-DNS-01-LIMITS tuple rows: only an equivalent hostname/type/mode/resolver
// tuple waits. Changing the hostname, record type or resolver lets the second
// operation proceed while the first is held. Transport and network generation
// are deliberately not part of the key, so UDP and TCP queries for the same
// tuple also wait: a conservative superset of the public tuple.
// requirement-ref: REQ-DNS-01-0033
func TestDns01TupleComponentsDecideSerialization(t *testing.T) {
	for _, tc := range []struct {
		name, from, to string
		overlap        bool
	}{
		{"hostname", `"hostname":"example.org"`, `"hostname":"other.example.org"`, true},
		{"record type", `"record_type":"A"`, `"record_type":"AAAA"`, true},
		{"resolver", `"ip_address":"8.8.8.8"`, `"ip_address":"1.1.1.1"`, true},
		{"transport only", `"transport":"udp"`, `"transport":"tcp"`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, d := dnsHarness(t, "udp")
			p := synPolicyInput(2, "reg-a", func(p *task.Permissions) {
				p.Tests = []task.TestVersion{dnsTest}
				p.RegistrationIDs = []string{"reg-a", "reg-b", "resolver-a", "resolver-b", "resolver-c"}
				p.AllowUnknownRisk = true
			})
			base := p.Targets[0]
			base.Test = dnsTest
			names, resolvers := base, base
			names.Tuple = task.TargetTuple{Hostname: "example.org"}
			other := base
			other.ID = "reg-b"
			other.Tuple = task.TargetTuple{Hostname: "other.example.org"}
			p.Targets = []task.TargetRegistration{names, other}
			for _, r := range []struct{ id, ip, scheme string }{{"resolver-a", "8.8.8.8", "udp"}, {"resolver-b", "1.1.1.1", "udp"}, {"resolver-c", "8.8.8.8", "tcp"}} {
				resolvers.ID = r.id
				resolvers.Role = task.ResolverTarget
				resolvers.Tuple = task.TargetTuple{Address: netip.MustParseAddr(r.ip), Port: 53, Scheme: r.scheme}
				p.Targets = append(p.Targets, resolvers)
			}
			var e error
			if h.adapters.policy, e = task.NewLocalPolicy(p); e != nil {
				t.Fatal(e)
			}
			entered := make(chan struct{}, 2)
			release := make(chan struct{})
			d.dial = func(ctx context.Context, _ string) (measurement.Conn, error) {
				entered <- struct{}{}
				select {
				case <-release:
				case <-ctx.Done():
				}
				return nil, errors.New("synthetic")
			}
			c := h.initialized(t)
			defer c.Close()
			one, e := c.StartRun(h.admit(t, c, dnsTask("first", "udp")))
			if e != nil {
				t.Fatal(e)
			}
			select {
			case <-entered:
			case <-time.After(time.Second):
				t.Fatal("no first operation")
			}
			two, e := c.StartRun(h.admit(t, c, strings.Replace(dnsTask("second", "udp"), tc.from, tc.to, 1)))
			if e != nil {
				t.Fatal(e)
			}
			overlapped := false
			select {
			case <-entered:
				overlapped = true
			case <-time.After(20 * time.Millisecond):
			}
			close(release)
			waitTerminal(t, c, one)
			waitTerminal(t, c, two)
			if overlapped != tc.overlap || d.calls.Load() != 2 {
				t.Fatalf("overlap %v, want %v; %d dials", overlapped, tc.overlap, d.calls.Load())
			}
		})
	}
}

// Collection rows of M-DNS-01-EXECUTION, -OUTPUT and -LIMITS through Client:
// matching UDP responses keep their arrival order and coarse offsets; with
// collection disabled only the first is read; a fourth response is never read
// once three are retained.
// requirement-ref: REQ-DNS-01-0010
// requirement-ref: REQ-DNS-01-0033
// requirement-ref: REQ-DNS-01-0035
func TestMDns01CollectionOrderAndBounds(t *testing.T) {
	second := func(q []byte) []byte {
		return loopback.DNSResponse(q, 0, []loopback.DNSRecord{{Type: "A", Value: "1.0.0.1"}})
	}
	const collect = `"transport":"udp","collect_additional_responses":true,"additional_response_window_ms":500`
	t.Run("ordered offsets", func(t *testing.T) {
		h, d := dnsHarness(t, "udp")
		dnsLoopback(t, d, func(q []byte) [][]byte { return [][]byte{oneAnswer(q)[0], second(q)} })
		d.beforeRead = func(read int32) {
			if read == 1 {
				_ = h.run.clock.Advance(250 * time.Millisecond)
			}
		}
		c := h.initialized(t)
		defer c.Close()
		rh, e := c.StartRun(h.admit(t, c, strings.Replace(dnsTask("order", "udp"), `"transport":"udp"`, collect, 1)))
		if e != nil {
			t.Fatal(codeOf(e))
		}
		waitDNSReads(t, d, h.run.clock, 500*time.Millisecond)
		waitTerminal(t, c, rh)
		r := dnsResult(t, h.run.output.Records()[0], "completed", "", "conflicting_responses")
		got := r.Measurement.Observations[0].Details.Responses
		if len(got) != 2 || got[0].Index != 1 || got[1].Index != 2 || got[0].Offset != 0 || got[1].Offset != 250 ||
			got[0].Answers[0].Value != "1.1.1.1" || got[1].Answers[0].Value != "1.0.0.1" {
			t.Fatalf("responses %+v", got)
		}
	})
	t.Run("collection disabled", func(t *testing.T) {
		h, d := dnsHarness(t, "udp")
		dnsLoopback(t, d, func(q []byte) [][]byte { return [][]byte{oneAnswer(q)[0], second(q)} })
		c := h.initialized(t)
		defer c.Close()
		r := dnsResult(t, tcpRaw(t, h, c, dnsTask("single", "udp")), "completed", "", "resolved")
		got := r.Measurement.Observations[0].Details.Responses
		if len(got) != 1 || got[0].Answers[0].Value != "1.1.1.1" || r.Measurement.Observations[0].Details.Limit || d.reads.Load() != 1 {
			t.Fatalf("responses %+v, %d reads", got, d.reads.Load())
		}
	})
	t.Run("fourth response unread", func(t *testing.T) {
		h, d := dnsHarness(t, "udp")
		s := dnsLoopback(t, d, func(q []byte) [][]byte {
			a := oneAnswer(q)[0]
			return [][]byte{a, a, a, second(q)}
		})
		c := h.initialized(t)
		defer c.Close()
		rh, e := c.StartRun(h.admit(t, c, strings.Replace(dnsTask("bound", "udp"), `"transport":"udp"`, collect, 1)))
		if e != nil {
			t.Fatal(codeOf(e))
		}
		waitTerminal(t, c, rh)
		r := dnsResult(t, h.run.output.Records()[0], "completed", "", "response_limit_reached")
		got := r.Measurement.Observations[0].Details.Responses
		if len(got) != 3 || !r.Measurement.Observations[0].Details.Limit || d.reads.Load() != 3 {
			t.Fatalf("responses %+v, %d reads", got, d.reads.Load())
		}
		for i, response := range got {
			if response.Index != i+1 || response.Answers[0].Value != "1.1.1.1" {
				t.Fatalf("response %d %+v", i, response)
			}
		}
		if queries, _, _ := s.Counts(); queries != 1 || d.calls.Load() != 1 {
			t.Fatal("retry or second query")
		}
	})
}

// M-DNS-01-ADMISSION and -EXECUTION: a CNAME chain in the answer is retained
// as observed data only; no connection or query follows it.
// requirement-ref: REQ-DNS-01-0008
// requirement-ref: REQ-DNS-01-0022
func TestDns01CNAMEAnswersCauseNoContact(t *testing.T) {
	for _, mode := range []string{"udp", "tcp"} {
		t.Run(mode, func(t *testing.T) {
			h, d := dnsHarness(t, mode)
			s := dnsLoopback(t, d, func(q []byte) [][]byte {
				return [][]byte{loopback.DNSResponse(q, 0, []loopback.DNSRecord{{Type: "CNAME", Value: "cdn.example.net"}, {Owner: "cdn.example.net", Type: "A", Value: "1.1.1.1"}})}
			})
			c := h.initialized(t)
			defer c.Close()
			r := dnsResult(t, tcpRaw(t, h, c, dnsTask("cname", mode)), "completed", "", "resolved")
			answers := r.Measurement.Observations[0].Details.Responses[0].Answers
			if len(answers) != 2 || answers[0].Type != "CNAME" || answers[0].Value != "cdn.example.net" || answers[1].Value != "1.1.1.1" {
				t.Fatalf("answers %+v", answers)
			}
			if queries, _, _ := s.Counts(); queries != 1 || d.calls.Load() != 1 {
				t.Fatal("CNAME target queried or contacted")
			}
		})
	}
}

// The result schema stays bound to the admitted mode/type/version.
// requirement-ref: REQ-DNS-01-0034
// requirement-ref: REQ-DNS-01-0035
func TestDns01ResultVersionsAndClosedOutput(t *testing.T) {
	h, d := dnsHarness(t, "udp")
	dnsLoopback(t, d, oneAnswer)
	c := h.initialized(t)
	defer c.Close()
	raw := tcpRaw(t, h, c, dnsTask("versions", "udp"))
	cd, _ := observation.NewContextDecoder(result.Contract, observation.ContextApprovals{})
	path, _ := observation.NewPath(1, 0, "direct", "")
	decoder, e := result.New(result.Binding{TaskID: "versions", Output: result.Output{Test: model.Test{ID: "DNS-01", Version: "0.1.0", ResultVersion: "0.1.0"}, DNSMode: "udp", QueryType: "A", Limits: observation.ProtocolLimits{RetainDNSAnswers: true}}, MaxAttempts: 1, MaxDurationMS: 60000, Implementation: implementation, ContextDecoder: cd, Path: path, AccessGeneration: 1})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = decoder.Decode(raw); e != nil {
		t.Fatal(e)
	}
	for _, field := range []string{"schema_version", "version", "result_version"} {
		for _, value := range []any{"0.1.1", "0.2.0", "1.0.0", "0.1", "v0.1.0", "", nil, 1} {
			var r map[string]any
			_ = json.Unmarshal(raw, &r)
			if field == "schema_version" {
				r[field] = value
			} else {
				r["test"].(map[string]any)[field] = value
			}
			b, _ := json.Marshal(r)
			if _, e := decoder.Decode(b); e == nil {
				t.Fatalf("accepted %s %v", field, value)
			}
		}
	}
	for _, mutate := range []func(map[string]any){func(r map[string]any) { delete(r, "schema_version") }, func(r map[string]any) { delete(r["test"].(map[string]any), "result_version") }, func(r map[string]any) { r["test"].(map[string]any)["id"] = "DNS-03" }, func(r map[string]any) { r["test"].(map[string]any)["id"] = "DNS-04" }, func(r map[string]any) { r["test"].(map[string]any)["id"] = "NET-05" }, func(r map[string]any) { r["test"].(map[string]any)["extra"] = true }, func(r map[string]any) { r["measurement"].(map[string]any)["profile"] = map[string]any{} },
		func(r map[string]any) { r["measurement"].(map[string]any)["steps"] = []any{} }, func(r map[string]any) { r["measurement"].(map[string]any)["stopped_before_step"] = "step-1" }, func(r map[string]any) { r["measurement"].(map[string]any)["observations"] = []any{} }, func(r map[string]any) {
			o := r["measurement"].(map[string]any)["observations"].([]any)[0].(map[string]any)
			o["details"].(map[string]any)["engine"] = "doh"
		}, func(r map[string]any) {
			o := r["measurement"].(map[string]any)["observations"].([]any)[0].(map[string]any)
			o["details"].(map[string]any)["engine"] = "dot"
		}} {
		var r map[string]any
		_ = json.Unmarshal(raw, &r)
		mutate(r)
		b, _ := json.Marshal(r)
		if _, e := decoder.Decode(b); e == nil {
			t.Fatal("accepted cross-test or malformed output")
		}
	}
}

// requirement-ref: REQ-JSON-0001
// requirement-ref: REQ-JSON-0002
// requirement-ref: REQ-JSON-0003
// requirement-ref: REQ-JSON-0004
// requirement-ref: REQ-JSON-0005
// requirement-ref: REQ-DNS-01-0023
func TestDns01StrictSchemaBothOrigins(t *testing.T) {
	for _, origin := range []string{"embedded", "remote"} {
		for _, tc := range []struct {
			input string
			valid bool
		}{
			{dnsInput, true}, {`{"\u0068ostname":"example.org","record_type":"A","resolver":{"ip_address":"8.8.8.8","\u0070ort":53}}`, true},
			{`{"hostname":"example.org","hostname":"example.org","record_type":"A","resolver":{"ip_address":"8.8.8.8"}}`, false},
			{`{"hostname":"example.org","\u0068ostname":"example.org","record_type":"A","resolver":{"ip_address":"8.8.8.8"}}`, false},
			{`{"hostname":"example.org","record_type":"A","resolver":{"ip_address":"8.8.8.8","port":53,"\u0070ort":53}}`, false},
			{`{"hostname":"example.org","record_type":"A","resolver":{"ip_address":"8.8.8.8","\u0078":1}}`, false},
			{`{"hostname":"example.org","record_type":"A","resolver":{"ip_address":"8.8.8.8","port":5.3e1}}`, false},
			{`{"hostname":"example.org","record_type":"A","resolver":{"ip_address":"8.8.8.8","port":53.0}}`, false},
			{`{"hostname":"example.org","record_type":"A","resolver":{"ip_address":"8.8.8.8","port":-0}}`, false},
		} {
			h, d := dnsHarness(t, "udp")
			c := h.initialized(t)
			data := strings.Replace(dnsTask("strict", "udp"), dnsInput, tc.input, 1)
			presented, ref := h.deliver(t, origin, data)
			handle, e := c.AdmitTask(presented, ref)
			if tc.valid && (e != nil || handle == "") || !tc.valid && (codeOf(e) != "task_schema_invalid" || handle != "" || h.replay.Commits() != 0 || taskHandles(c) != 0) {
				t.Fatalf("%s %s: %s", origin, tc.input, codeOf(e))
			}
			if d.calls.Load() != 0 {
				t.Fatal("admission contact")
			}
			c.Close()
		}
	}
	for _, field := range []string{"input", "options"} {
		size, content := 49152, dnsInput
		if field == "options" {
			size, content = 8192, `{"resolver_mode":"direct","transport":"udp"}`
		}
		for _, delta := range []int{-1, 0, 1} {
			h, d := dnsHarness(t, "udp")
			c := h.initialized(t)
			pad := content[:len(content)-1] + strings.Repeat(" ", size+delta-len(content)) + "}"
			data := strings.Replace(dnsTask("bytes", "udp"), `"`+field+`":`+content, `"`+field+`":`+pad, 1)
			_, e := c.AdmitTask(data, h.embedded(t, data))
			if delta <= 0 && e != nil || delta > 0 && codeOf(e) != "task_limit_exceeded" || d.calls.Load() != 0 {
				t.Fatal(field, delta, codeOf(e))
			}
			c.Close()
		}
	}
}
func FuzzDns01Admission(f *testing.F) {
	for _, input := range []string{dnsInput, `null`, `{}`, `{"hostname":"bücher.example.org","record_type":"AAAA"}`} {
		f.Add(input, `{"resolver_mode":"direct"}`)
	}
	f.Fuzz(func(t *testing.T, input, options string) {
		if len(input)+len(options) > 60000 {
			t.Skip()
		}
		h, d := dnsHarness(t, "udp")
		h.adapters.deadline = defaultLifecycleDeadline
		c := h.initialized(t)
		defer c.Close()
		data := strings.Replace(dnsTask("fuzz-dns", "udp"), dnsInput, input, 1)
		data = strings.Replace(data, `"options":{"resolver_mode":"direct","transport":"udp"}`, `"options":`+options, 1)
		handle, e := c.AdmitTask(data, h.embedded(t, data))
		if d.calls.Load() != 0 || len(h.run.output.Records()) != 0 || len(c.rt.runs.entries) != 0 {
			t.Fatal("admission executed")
		}
		if e != nil && (handle != "" || taskHandles(c) != 0 || h.replay.Commits() != 0) {
			t.Fatal("non-atomic rejection")
		}
	})
}

// requirement-ref: REQ-DNS-01-0014
// requirement-ref: REQ-DNS-01-0035
func TestDns01AAAAExecution(t *testing.T) {
	for _, mode := range []string{"udp", "tcp"} {
		t.Run(mode, func(t *testing.T) {
			h, d := dnsHarness(t, mode)
			dnsLoopback(t, d, func(q []byte) [][]byte {
				return [][]byte{loopback.DNSResponse(q, 0, []loopback.DNSRecord{{Type: "AAAA", Value: "2606:4700:4700::1111"}})}
			})
			c := h.initialized(t)
			defer c.Close()
			data := strings.Replace(dnsTask("aaaa", mode), `"record_type":"A"`, `"record_type":"AAAA"`, 1)
			raw := tcpRaw(t, h, c, data)
			r := dnsResult(t, raw, "completed", "", "resolved")
			details := r.Measurement.Observations[0].Details
			if details.QueryType != "AAAA" || len(details.Responses[0].Answers) != 1 || details.Responses[0].Answers[0].Type != "AAAA" {
				t.Fatalf("%s", raw)
			}
			if _, e := c.NextBundle(); e != nil {
				t.Fatal(e)
			}
		})
	}
}

// requirement-ref: REQ-TRUST-0016
// requirement-ref: REQ-DNS-01-0018
// requirement-ref: REQ-DNS-01-0020
func TestDns01RechecksBothRegistrationExpiries(t *testing.T) {
	for _, expired := range []string{"name", "resolver"} {
		t.Run(expired, func(t *testing.T) {
			h, d := dnsHarness(t, "udp")
			p := synPolicyInput(2, "reg-a", func(p *task.Permissions) {
				p.Tests = []task.TestVersion{dnsTest}
				p.RegistrationIDs = []string{"reg-a", "resolver-a"}
			})
			p.Targets[0].Test = dnsTest
			p.Targets[0].Tuple = task.TargetTuple{Hostname: "example.org"}
			r := p.Targets[0]
			r.ID = "resolver-a"
			r.Role = task.ResolverTarget
			r.Tuple = task.TargetTuple{Address: netip.MustParseAddr("8.8.8.8"), Port: 53, Scheme: "udp"}
			p.Targets = append(p.Targets, r)
			index := 0
			if expired == "resolver" {
				index = 1
			}
			p.Targets[index].ExpiresAt = task.TimestampFromTime(epoch.Add(time.Second))
			h.adapters.policy, _ = task.NewLocalPolicy(p)
			c := h.initialized(t)
			defer c.Close()
			th := h.admit(t, c, dnsTask("expiry", "udp"))
			h.advance(2 * time.Second)
			rh, e := c.StartRun(th)
			if e != nil {
				t.Fatal(codeOf(e))
			}
			waitTerminal(t, c, rh)
			dnsResult(t, h.run.output.Records()[0], "incomplete", "local_policy_rejected", "")
			if d.calls.Load() != 0 {
				t.Fatal("expired registration contacted")
			}
		})
	}
}
