package barometer

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"net/netip"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/config"
	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/measurement/dns03"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

var dohTest = task.TestVersion{ID: "DNS-03", Version: "0.1.0"}

const (
	dohName       = "doh.resolver.example"
	dohURL        = "https://doh.resolver.example/dns-query"
	dohInput      = `{"resolver_url":"https://doh.resolver.example/dns-query","hostname":"query.example","record_type":"A","bootstrap_ip":"8.8.8.8"}`
	dohSystem     = `{"resolver_url":"https://doh.resolver.example/dns-query","hostname":"query.example","record_type":"A"}`
	dohAuthority  = "8.8.8.8:443"
	dohQueryBytes = 31 // header, query.example and question tail
)

func dohParams(t *testing.T, input, options string) dns03.Parameters {
	t.Helper()
	c := dns03.Codec{}
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
	return p.(dns03.Parameters)
}

// dohPolicy registers the query name and the complete resolver profile that
// the production target rule extracts from input.
func dohPolicy(t *testing.T, h *harness, input string) {
	t.Helper()
	targets, ok := dns03Targets{}.Targets(dohParams(t, input, `{}`))
	if !ok || len(targets) != 2 {
		t.Fatal("target rule")
	}
	p := synPolicyInput(2, "reg-a", func(p *task.Permissions) {
		p.Tests = []task.TestVersion{dohTest}
		p.RegistrationIDs = []string{"reg-a", "reg-b"}
	})
	resolver := p.Targets[0].Clone()
	resolver.ID, resolver.Role, resolver.Tuple = "reg-b", targets[1].Role, targets[1].Tuple
	p.Targets[0].Test, p.Targets[0].Role, p.Targets[0].Tuple = dohTest, targets[0].Role, targets[0].Tuple
	resolver.Test = dohTest
	p.Targets = append(p.Targets, resolver)
	var e error
	if h.adapters.policy, e = task.NewLocalPolicy(p); e != nil {
		t.Fatal(e)
	}
}

// dohTransport maps the checked resolver authority to the loopback server and
// fails any other destination; lookups counts system bootstrap requests.
type dohTransport struct {
	// authorities, when set, replace the fixed resolver authority.
	authorities []string
	calls       atomic.Int32
	lookups     atomic.Int32
	dial        func(context.Context) (measurement.Conn, error)
}

func (d *dohTransport) Dialer(int) (measurement.Dialer, error) { return d, nil }
func (d *dohTransport) DialContext(ctx context.Context, network, address string) (measurement.Conn, error) {
	d.calls.Add(1)
	if !strings.HasPrefix(network, "tcp") || len(d.authorities) == 0 && address != dohAuthority || len(d.authorities) > 0 && !slices.Contains(d.authorities, address) {
		return nil, errors.New("CANARY-RAW-ERROR")
	}
	return d.dial(ctx)
}

type dohFixture struct {
	ca   *loopback.TestCA
	cert loopback.HTTPConfig
}

func newDohFixture(t *testing.T) dohFixture {
	t.Helper()
	ca, err := loopback.NewTestCA()
	if err != nil {
		t.Fatal(err)
	}
	cert, err := ca.Issue(dohName, false)
	if err != nil {
		t.Fatal(err)
	}
	return dohFixture{ca, loopback.HTTPConfig{Certificate: cert}}
}

func (f dohFixture) reply(fn func([]byte) []byte) loopback.HTTPConfig {
	c := f.cert
	c.DoH = fn
	return c
}

func dohHarness(t *testing.T, f dohFixture, input string) (*harness, *dohTransport) {
	t.Helper()
	h := runningHarness(t)
	h.adapters.registry = nil
	h.adapters.measurementRoots = f.ca.Pool()
	h.adapters.run.paths = nil
	h.adapters.approvals = config.Approvals{ConnectionCodes: []string{"outline_mobileproxy"}}
	dohPolicy(t, h, input)
	d := &dohTransport{dial: func(context.Context) (measurement.Conn, error) { return nil, syscall.ECONNREFUSED }}
	h.adapters.run.directDialers = d
	h.adapters.run.resolver = dnsSystemResolver{func(ctx context.Context, family, name string) ([]netip.Addr, error) {
		d.lookups.Add(1)
		if family != "ip4" || name != dohName {
			return nil, errors.New("CANARY-RAW-ERROR")
		}
		return []netip.Addr{netip.MustParseAddr("8.8.8.8")}, nil
	}}
	return h, d
}

// dohRoute sends every checked dial to address.
func dohRoute(d *dohTransport, address string) {
	d.dial = func(ctx context.Context) (measurement.Conn, error) {
		var nd net.Dialer
		return nd.DialContext(ctx, "tcp", address)
	}
}

func dohServe(t *testing.T, d *dohTransport, cfg loopback.HTTPConfig) *loopback.HTTPServer {
	t.Helper()
	srv, err := loopback.ListenHTTP(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = srv.Close() })
	dohRoute(d, srv.Address())
	return srv
}

func dohServeTLS(t *testing.T, d *dohTransport, cfg loopback.TLSConfig) *loopback.TLSServer {
	t.Helper()
	srv, err := loopback.ListenTLS(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = srv.Close() })
	dohRoute(d, srv.Address())
	return srv
}

func dohTask(id, input, options string) string {
	s := strings.Replace(runTask(id, "", "", false), `"test_id":"TCP-01"`, `"test_id":"DNS-03"`, 1)
	s = strings.Replace(s, `"input":{}`, `"input":`+input, 1)
	return strings.Replace(s, `"options":{}`, `"options":`+options, 1)
}

// dohHTTP frames a raw HTTP/1.1 response carrying header canaries that the
// result must not retain.
func dohHTTP(status int, contentType string, body []byte) []byte {
	head := fmt.Sprintf("HTTP/1.1 %d Synthetic\r\nSet-Cookie: cookie-canary=1\r\nX-Header-Canary: header-canary\r\n", status)
	if contentType != "" {
		head += "Content-Type: " + contentType + "\r\n"
	}
	return append([]byte(head+fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))), body...)
}

func dohReply(code uint16, records ...loopback.DNSRecord) func([]byte) []byte {
	return func(q []byte) []byte {
		return dohHTTP(200, "application/dns-message", loopback.DNSResponse(q, code, records))
	}
}

var dohAnswer = dohReply(0, loopback.DNSRecord{Type: "A", Value: "9.9.9.9"})

// dohSized pads a NOERROR response with one additional record of an
// uninterpreted type to exactly size bytes.
func dohSized(size int) func([]byte) []byte {
	return func(q []byte) []byte {
		r := loopback.DNSResponse(q, 0, []loopback.DNSRecord{{Type: "A", Value: "9.9.9.9"}})
		rdata := size - len(r) - 11
		r = append(r, 0, 0, 16, 0, 1, 0, 0, 0, 60, byte(rdata>>8), byte(rdata))
		r = append(r, make([]byte, rdata)...)
		binary.BigEndian.PutUint16(r[10:], binary.BigEndian.Uint16(r[10:])+1)
		return dohHTTP(200, "application/dns-message", r)
	}
}

// dohReadFails accepts writes and fails every read with an untyped error,
// which no stage class covers.
type dohReadFails struct{}

func (dohReadFails) Read([]byte) (int, error)    { return 0, errors.New("CANARY-RAW-ERROR") }
func (dohReadFails) Write(b []byte) (int, error) { return len(b), nil }
func (dohReadFails) Close() error                { return nil }
func (dohReadFails) SetDeadline(time.Time) error { return nil }

type dohExpect struct {
	state, reason, outcome string
	types                  []string // observation types, in order
	failure                string   // terminal observation failure
	validation, rcode      string
	answers, status        int
}

// dohResult checks the serialized result shape and that no resolver URL,
// hostname or address, bootstrap address, query name, raw DNS message, HTTP
// header or body, certificate, cookie or raw error material is retained
// (REQ-DNS-03-0029).
func dohResult(t *testing.T, raw []byte, want dohExpect) {
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
	if r.Execution.State != want.state || r.Execution.Reason != want.reason {
		t.Fatalf("state: %s", raw)
	}
	for _, canary := range []string{dohName, "dns-query", "query.example", "8.8.8.8", `"resolver_url"`, `"bootstrap_ip"`, `"hostname"`, loopback.CertificateCanary, "424242", "91337",
		"CANARY-RAW-ERROR", "x509", "127.0.0.1", "cookie-canary", "header-canary", "body-canary", "dns-message", "text/html", "Synthetic", `"ttl"`, `"redirect"`, `"checks"`} {
		if strings.Contains(string(raw), canary) {
			t.Fatalf("retained %s: %s", canary, raw)
		}
	}
	if want.outcome == "" {
		if r.Measurement != nil {
			t.Fatalf("interrupted measurement: %s", raw)
		}
		if want.state == "not_run" && len(r.NetworkContext) != 0 {
			t.Fatal("not_run has context")
		}
		return
	}
	m := r.Measurement
	if m == nil || m.Outcome != want.outcome || len(m.Observations) != len(want.types) {
		t.Fatalf("measurement: %s", raw)
	}
	allowed := map[string]map[string]bool{
		"tcp_connect":   {"address_family": true},
		"tls_handshake": {"address_family": true, "tls_version": true, "alpn": true, "certificate_validation": true, "alert": true},
		"http_exchange": {"address_family": true, "protocol": true, "status_code": true, "body": true, "content_encoding": true},
		"dns_query":     {"engine": true, "address_family": true, "query_type": true, "responses": true},
	}
	for i, o := range m.Observations {
		last := i == len(m.Observations)-1
		// A non-2xx HTTP exchange is the terminal stage with a succeeded
		// observation (GO-SPEC-ISSUE-020).
		okStatus := !last && o.Status == "succeeded" || last && o.Failure == want.failure && (want.failure == "") == (o.Status == "succeeded")
		if o.Type != want.types[i] || o.Duration%250 != 0 || !okStatus {
			t.Fatalf("observation %d: %s", i, raw)
		}
		for k := range o.Details {
			if !allowed[o.Type][k] {
				t.Fatalf("detail %s: %s", k, raw)
			}
		}
		str := func(k string) string { s, _ := o.Details[k].(string); return s }
		switch o.Type {
		case "tls_handshake":
			if str("address_family") != "ipv4" || last && str("certificate_validation") != want.validation || !last && (str("certificate_validation") != "trusted" || str("alpn") != "http/1.1") {
				t.Fatalf("tls_handshake: %s", raw)
			}
		case "http_exchange":
			code, _ := o.Details["status_code"].(float64)
			if str("protocol") != "http/1.1" || str("address_family") != "ipv4" || want.status != 0 && int(code) != want.status {
				t.Fatalf("http_exchange: %s", raw)
			}
		case "dns_query":
			engine, responses := "doh", 1
			if i == 0 {
				engine, responses = "system", 0
			}
			list, _ := o.Details["responses"].([]any)
			if str("engine") != engine || str("query_type") != "A" || engine == "system" && str("address_family") != "" || engine == "doh" && str("address_family") != "ipv4" {
				t.Fatalf("dns_query: %s", raw)
			}
			if engine == "doh" && want.rcode == "" {
				responses = 0
			}
			if len(list) != responses {
				t.Fatalf("responses: %s", raw)
			}
			if responses == 1 {
				resp := list[0].(map[string]any)
				answers, _ := resp["answers"].([]any)
				if resp["rcode"] != want.rcode || len(answers) != want.answers || len(resp) != 4 || resp["arrival_offset_ms"] != float64(o.Duration) {
					t.Fatalf("response: %s", raw)
				}
			}
		}
	}
}

var (
	dohBootstrapped = []string{"tcp_connect", "tls_handshake", "http_exchange", "dns_query"}
	dohResolved     = []string{"dns_query", "tcp_connect", "tls_handshake", "http_exchange", "dns_query"}
	dohHTTPStage    = []string{"tcp_connect", "tls_handshake", "http_exchange"}
	dohResponse     = dohExpect{state: "completed", outcome: "response", types: dohBootstrapped, rcode: "NOERROR", answers: 1, status: 200}
)

// dohRequest checks the one ordinary RFC 8484 request: POST to the
// registered path with the DNS media types, identity encoding, a 31-byte
// query with ID 0, and no cookie, credential, user agent or other
// identifying header.
func dohRequest(t *testing.T, c loopback.HTTPCounts) {
	t.Helper()
	if len(c.Requests) != 1 || len(c.Hellos) != 1 || c.Hellos[0].ServerName != dohName || !slices.Equal(c.Hellos[0].ALPN, []string{"http/1.1"}) {
		t.Fatalf("requests %+v hellos %+v", c.Requests, c.Hellos)
	}
	r := c.Requests[0]
	if r.Method != "POST" || r.Target != "/dns-query" || len(r.Body) != dohQueryBytes || binary.BigEndian.Uint16(r.Body) != 0 ||
		r.Header.Get("Content-Type") != "application/dns-message" || r.Header.Get("Accept") != "application/dns-message" || r.Header.Get("Accept-Encoding") != "identity" {
		t.Fatalf("request %+v", r)
	}
	for name := range r.Header {
		if !slices.Contains([]string{"Content-Type", "Accept", "Accept-Encoding", "Content-Length", "Connection"}, name) {
			t.Fatalf("unexpected header %s", name)
		}
	}
}

// M-DNS-03-ADMISSION: actual Client admission against the registered resolver
// profile and query name, with atomic rejection before any lookup or socket.
// requirement-ref: REQ-DNS-03-0004
// requirement-ref: REQ-DNS-03-0009
// requirement-ref: REQ-DNS-03-0010
// requirement-ref: REQ-DNS-03-0011
// requirement-ref: REQ-DNS-03-0013
// requirement-ref: REQ-DNS-03-0014
// requirement-ref: REQ-DNS-03-0015
// requirement-ref: REQ-DNS-03-0038
func TestMDns03Admission(t *testing.T) {
	f := newDohFixture(t)
	long := strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", 53) + ".example"
	exact := "https://" + dohName + "/" + strings.Repeat("p", 2048-len("https://"+dohName+"/"))
	for _, input := range []string{
		dohInput, dohSystem,
		strings.Replace(dohInput, "8.8.8.8", "2606:4700:4700::1111", 1),
		strings.Replace(dohInput, dohName, dohName+":8443", 1),
		strings.Replace(dohInput, "/dns-query", "/dns-query?ct=dns", 1),
		strings.Replace(dohInput, "/dns-query", "", 1),
		strings.Replace(dohInput, `"record_type":"A"`, `"record_type":"AAAA"`, 1),
		strings.Replace(strings.Replace(dohInput, dohName, "DoH.Resolver.EXAMPLE", 1), "query.example", "bücher.example", 1),
		strings.Replace(strings.Replace(dohInput, dohName, long, 1), "query.example", long, 1),
		strings.Replace(dohInput, dohURL, exact, 1),
	} {
		t.Run("valid/"+input[:min(len(input), 120)], func(t *testing.T) {
			h, d := dohHarness(t, f, input)
			c := h.initialized(t)
			defer c.Close()
			data := dohTask("valid", input, `{}`)
			if _, err := c.AdmitTask(data, h.embedded(t, data)); err != nil {
				t.Fatal(codeOf(err))
			}
			if d.calls.Load() != 0 || d.lookups.Load() != 0 {
				t.Fatal("admission contacted transport or resolver")
			}
		})
	}
	field := func(name, value string) string {
		return strings.Replace(dohInput, fmt.Sprintf(`"%s":%s`, name, map[string]string{"resolver_url": `"` + dohURL + `"`, "hostname": `"query.example"`, "record_type": `"A"`, "bootstrap_ip": `"8.8.8.8"`}[name]), fmt.Sprintf(`"%s":%s`, name, value), 1)
	}
	inputs := []string{`{}`, `null`, `[]`, dohSystem[:len(dohSystem)-1] + `,"x":1}`, dohInput[:len(dohInput)-1] + `,"method":"GET"}`, dohInput[:len(dohInput)-1] + `,"headers":{}}`,
		`{"hostname":"query.example","record_type":"A"}`, `{"resolver_url":"` + dohURL + `","record_type":"A"}`, `{"resolver_url":"` + dohURL + `","hostname":"query.example"}`}
	for _, v := range []string{`"TXT"`, `"a"`, `"ANY"`, `""`, `null`} {
		inputs = append(inputs, field("record_type", v))
	}
	for _, u := range []string{"", "http://doh.resolver.example/dns-query", "https://user:secret@doh.resolver.example/dns-query", "https://doh.resolver.example/dns-query#frag", "https://doh.resolver.example/dns-query#",
		"https://doh.resolver.example/dns-query{?dns}", "https://${resolver}/dns-query", "https://8.8.8.8/dns-query", "https://[2606:4700:4700::1111]/dns-query", "https://doh.local/dns-query",
		"https://localhost/dns-query", "https://doh.resolver.example:0/dns-query", "https://doh.resolver.example:65536/dns-query", "https://doh.resolver.example:0443/dns-query",
		"https://doh.resolver.example:/dns-query", "//doh.resolver.example/dns-query", "doh.resolver.example/dns-query", "https:///dns-query", "https://doh.resolver.example/dns query", "mailto:doh@resolver.example"} {
		inputs = append(inputs, field("resolver_url", fmt.Sprintf("%q", u)))
	}
	for _, n := range []string{"", "8.8.8.8", "[::1]", "*.example", "example", "a.example.", "query.local", "localhost", "${hostname}", "{device}.example", " query.example", long + "x"} {
		inputs = append(inputs, field("hostname", fmt.Sprintf("%q", n)))
	}
	for _, ip := range []string{dohName, "${bootstrap}", "8.8.8.08", "0.0.0.0", "127.0.0.1", "10.0.0.1", "100.64.0.1", "169.254.1.1", "192.0.2.1", "::1", "fe80::1", "fc00::1", "2001:db8::1", "::ffff:8.8.8.8", "2606:4700:4700::ABCD", "fe80::1%en0", ""} {
		inputs = append(inputs, field("bootstrap_ip", fmt.Sprintf("%q", ip)))
	}
	for i, input := range inputs {
		t.Run(fmt.Sprint("invalid/", i), func(t *testing.T) {
			h, d := dohHarness(t, f, dohInput)
			c := h.initialized(t)
			defer c.Close()
			data := dohTask("invalid", input, `{}`)
			handle, err := c.AdmitTask(data, h.embedded(t, data))
			if handle != "" || codeOf(err) != "task_schema_invalid" {
				t.Fatalf("%s: %s", input, codeOf(err))
			}
			if taskHandles(c) != 0 || h.replay.Commits() != 0 || d.calls.Load() != 0 || d.lookups.Load() != 0 || len(h.run.output.Records()) != 0 {
				t.Fatal("non-atomic rejection")
			}
		})
	}
	t.Run("url above bound", func(t *testing.T) {
		h, d := dohHarness(t, f, dohInput)
		c := h.initialized(t)
		defer c.Close()
		data := dohTask("oversized", strings.Replace(dohInput, dohURL, exact+"p", 1), `{}`)
		handle, err := c.AdmitTask(data, h.embedded(t, data))
		if handle != "" || codeOf(err) != "task_limit_exceeded" || d.calls.Load() != 0 || h.replay.Commits() != 0 {
			t.Fatal(codeOf(err))
		}
	})
	for _, tc := range []struct{ name, registered, input string }{
		{"other path", dohInput, strings.Replace(dohInput, "/dns-query", "/resolve", 1)},
		{"added query", dohInput, strings.Replace(dohInput, "/dns-query", "/dns-query?user=1", 1)},
		{"other port", dohInput, strings.Replace(dohInput, dohName, dohName+":8443", 1)},
		{"other bootstrap", dohInput, strings.Replace(dohInput, "8.8.8.8", "8.8.4.4", 1)},
		{"bootstrap omitted", dohInput, dohSystem},
		{"bootstrap added", dohSystem, dohInput},
		{"other resolver", dohInput, strings.Replace(dohInput, dohName, "other.example", 1)},
		{"other query name", dohInput, strings.Replace(dohInput, "query.example", "other.example", 1)},
	} {
		t.Run("unregistered/"+tc.name, func(t *testing.T) {
			h, d := dohHarness(t, f, tc.registered)
			c := h.initialized(t)
			defer c.Close()
			data := dohTask("unregistered", tc.input, `{}`)
			_, err := c.AdmitTask(data, h.embedded(t, data))
			if codeOf(err) != "target_not_registered" || d.calls.Load() != 0 || d.lookups.Load() != 0 || h.replay.Commits() != 0 {
				t.Fatal(codeOf(err))
			}
		})
	}
}

// M-DNS-03-OPTIONS: the address family option and its bootstrap
// compatibility; GET, redirects, alternate endpoints, custom headers,
// credentials, cookies and proxy configuration are outside the closed shape.
// requirement-ref: REQ-DNS-03-0016
// requirement-ref: REQ-DNS-03-0017
func TestMDns03Options(t *testing.T) {
	f := newDohFixture(t)
	v6 := strings.Replace(dohInput, "8.8.8.8", "2606:4700:4700::1111", 1)
	for _, tc := range []struct{ input, options string }{
		{dohInput, `{}`}, {dohInput, `{"address_family":"auto"}`}, {dohInput, `{"address_family":"ipv4"}`}, {v6, `{"address_family":"ipv6"}`}, {v6, `{"address_family":"auto"}`},
		{dohSystem, `{"address_family":"ipv4"}`}, {dohSystem, `{"address_family":"ipv6"}`},
	} {
		t.Run("valid/"+tc.options, func(t *testing.T) {
			h, d := dohHarness(t, f, tc.input)
			c := h.initialized(t)
			defer c.Close()
			data := dohTask("options", tc.input, tc.options)
			if handle, err := c.AdmitTask(data, h.embedded(t, data)); err != nil || handle == "" {
				t.Fatal(codeOf(err))
			}
			if d.calls.Load() != 0 {
				t.Fatal("options contacted transport")
			}
		})
	}
	for _, tc := range []struct{ input, options string }{
		{dohInput, `{"address_family":"ipv6"}`}, {v6, `{"address_family":"ipv4"}`},
		{dohInput, `null`}, {dohInput, `[]`}, {dohInput, `{"address_family":null}`}, {dohInput, `{"address_family":""}`}, {dohInput, `{"address_family":"IPv4"}`}, {dohInput, `{"address_family":4}`}, {dohInput, `{"address_family":"dual"}`},
		{dohInput, `{"method":"GET"}`}, {dohInput, `{"follow_redirects":true}`}, {dohInput, `{"max_redirects":1}`}, {dohInput, `{"alternate_endpoints":["https://other.example/dns-query"]}`},
		{dohInput, `{"headers":{"X-Id":"1"}}`}, {dohInput, `{"credentials":"token"}`}, {dohInput, `{"cookies":"a=b"}`}, {dohInput, `{"proxy":"127.0.0.1:8080"}`},
		{dohInput, `{"reuse_connection":true}`}, {dohInput, `{"retries":1}`}, {dohInput, `{"x":0}`},
	} {
		t.Run("invalid/"+tc.options, func(t *testing.T) {
			h, d := dohHarness(t, f, tc.input)
			c := h.initialized(t)
			defer c.Close()
			data := dohTask("options", tc.input, tc.options)
			handle, err := c.AdmitTask(data, h.embedded(t, data))
			if codeOf(err) != "task_schema_invalid" || handle != "" || h.replay.Commits() != 0 || d.calls.Load() != 0 {
				t.Fatal(codeOf(err))
			}
		})
	}
}

// M-DNS-03-EXECUTION and OUTPUT: production registry, real sockets, a real
// handshake and a real HTTP/1.1 exchange. Bootstrap is omitted with an
// approved bootstrap address; one connection, one ClientHello naming the URL
// hostname, one RFC 8484 POST and one response. The proxy path needs the
// bootstrap address, and CONNECT uses it while TLS still authenticates the
// URL hostname.
// requirement-ref: REQ-DNS-03-0005
// requirement-ref: REQ-DNS-03-0006
// requirement-ref: REQ-DNS-03-0007
// requirement-ref: REQ-DNS-03-0011
// requirement-ref: REQ-DNS-03-0012
// requirement-ref: REQ-DNS-03-0024
// requirement-ref: REQ-DNS-03-0025
// requirement-ref: REQ-DNS-03-0026
// requirement-ref: REQ-DNS-03-0029
// requirement-ref: REQ-DNS-03-0035
// requirement-ref: REQ-DNS-03-0036
// requirement-ref: REQ-DNS-03-0037
// requirement-ref: REQ-PROXY-0044
func TestMDns03ExecutionAndOutput(t *testing.T) {
	f := newDohFixture(t)
	for _, tc := range []struct {
		name, input string
		proxy       bool
		want        dohExpect
	}{
		{"bootstrap address", dohInput, false, dohResponse},
		{"system bootstrap", dohSystem, false, dohExpect{state: "completed", outcome: "response", types: dohResolved, rcode: "NOERROR", answers: 1, status: 200}},
		{"proxy", dohInput, true, dohResponse},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, d := dohHarness(t, f, tc.input)
			srv := dohServe(t, d, f.reply(dohAnswer))
			config := `{}`
			data := dohTask("doh", tc.input, `{}`)
			var p *loopback.Proxy
			if tc.proxy {
				var err error
				if p, err = loopback.ListenProxy(loopback.ProxyConfig{Resolve: map[string]string{dohAuthority: srv.Address()}}); err != nil {
					t.Fatal(err)
				}
				defer p.Close()
				config = proxyConfig(p.Port())
				data = strings.Replace(data, `"limits":{}`, `"limits":{},"network_path":"proxy"`, 1)
			}
			c := h.client(t)
			if err := c.Initialize(config); err != nil {
				t.Fatal(codeOf(err))
			}
			defer c.Close()
			dohResult(t, tcpRaw(t, h, c, data), tc.want)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.Wait(ctx, func(c loopback.HTTPCounts) bool { return c.Accepted == 1 && c.Open == 0 }); err != nil {
				t.Fatal(err)
			}
			dohRequest(t, srv.Counts())
			wantLookups := int32(0)
			if tc.input == dohSystem {
				wantLookups = 1
			}
			if d.lookups.Load() != wantLookups {
				t.Fatal("bootstrap lookup count")
			}
			if tc.proxy {
				if err := p.WaitIdle(ctx); err != nil {
					t.Fatal(err)
				}
				pc := p.Counts()
				if pc.Accepted != 1 || len(pc.Requests) != 1 || !strings.HasPrefix(pc.Requests[0], "CONNECT "+dohAuthority+" HTTP/1.1") || d.calls.Load() != 0 {
					t.Fatalf("proxy counts %+v", pc)
				}
			} else if d.calls.Load() != 1 {
				t.Fatal("not one direct dial")
			}
			if _, err := c.NextBundle(); err != nil {
				t.Fatal("result did not survive bundle validation", err)
			}
		})
	}
	// The system bootstrap returns two addresses; the one selected address is
	// refused and no other address is tried.
	t.Run("one selected address without fallback", func(t *testing.T) {
		h, d := dohHarness(t, f, dohSystem)
		h.adapters.run.resolver = dnsSystemResolver{func(context.Context, string, string) ([]netip.Addr, error) {
			d.lookups.Add(1)
			return []netip.Addr{netip.MustParseAddr("8.8.8.8"), netip.MustParseAddr("8.8.4.4")}, nil
		}}
		c := h.initialized(t)
		defer c.Close()
		dohResult(t, tcpRaw(t, h, c, dohTask("one-address", dohSystem, `{}`)), dohExpect{state: "completed", outcome: "tcp_failure", types: []string{"dns_query", "tcp_connect"}, failure: "connection_refused"})
		if d.calls.Load() != 1 || d.lookups.Load() != 1 {
			t.Fatalf("dials %d lookups %d: address or bootstrap fallback", d.calls.Load(), d.lookups.Load())
		}
	})
	t.Run("proxy without bootstrap address", func(t *testing.T) {
		h, d := dohHarness(t, f, dohSystem)
		p, err := loopback.ListenProxy(loopback.ProxyConfig{})
		if err != nil {
			t.Fatal(err)
		}
		defer p.Close()
		c := h.client(t)
		if err := c.Initialize(proxyConfig(p.Port())); err != nil {
			t.Fatal(codeOf(err))
		}
		defer c.Close()
		data := strings.Replace(dohTask("proxy-system", dohSystem, `{}`), `"limits":{}`, `"limits":{},"network_path":"proxy"`, 1)
		handle, err := c.AdmitTask(data, h.embedded(t, data))
		if codeOf(err) != "task_schema_invalid" || handle != "" || d.calls.Load() != 0 || d.lookups.Load() != 0 || p.Counts().Accepted != 0 {
			t.Fatal(codeOf(err), "incompatible path contacted")
		}
	})
}

// M-DNS-03-NORMALIZATION: every outcome row from real loopback behavior or a
// wrapped platform error; the terminal failure stays at its observed stage
// with prior observations ordered. A non-2xx status is http_failure before
// the media type and without a DNS parse, and a redirect target is never
// contacted (REQ-DNS-03-0040).
// requirement-ref: REQ-DNS-03-0007
// requirement-ref: REQ-DNS-03-0008
// requirement-ref: REQ-DNS-03-0011
// requirement-ref: REQ-DNS-03-0012
// requirement-ref: REQ-DNS-03-0027
// requirement-ref: REQ-DNS-03-0029
// requirement-ref: REQ-DNS-03-0032
// requirement-ref: REQ-DNS-03-0040
func TestMDns03Normalization(t *testing.T) {
	f := newDohFixture(t)
	other, err := loopback.NewTestCA()
	if err != nil {
		t.Fatal(err)
	}
	untrusted, err := other.Issue(dohName, false)
	if err != nil {
		t.Fatal(err)
	}
	mismatched, err := f.ca.Issue("other.example", false)
	if err != nil {
		t.Fatal(err)
	}
	// A certificate valid only for the bootstrap address does not
	// authenticate the URL hostname (REQ-DNS-03-0011).
	addressOnly, err := f.ca.Issue("8.8.8.8", false)
	if err != nil {
		t.Fatal(err)
	}
	fixed := func(raw string) func([]byte) []byte { return func([]byte) []byte { return []byte(raw) } }
	tcpFail := func(err error) func(context.Context) (measurement.Conn, error) {
		return func(context.Context) (measurement.Conn, error) { return nil, err }
	}
	tls := []string{"tcp_connect", "tls_handshake"}
	httpFailed := func(failure string, status int) dohExpect {
		return dohExpect{state: "completed", outcome: "http_failure", types: dohHTTPStage, failure: failure, status: status}
	}
	dnsFailed := func(outcome, failure, rcode string) dohExpect {
		return dohExpect{state: "completed", outcome: outcome, types: dohBootstrapped, failure: failure, rcode: rcode, status: 200}
	}
	for _, tc := range []struct {
		name, input string
		server      *loopback.HTTPConfig
		tlsServer   *loopback.TLSConfig
		dial        func(context.Context) (measurement.Conn, error)
		lookup      error
		want        dohExpect
	}{
		{name: "response", input: dohInput, server: ptr(f.reply(dohAnswer)), want: dohResponse},
		{name: "response empty", input: dohInput, server: ptr(f.reply(dohReply(0))), want: dohExpect{state: "completed", outcome: "response", types: dohBootstrapped, rcode: "NOERROR", status: 200}},
		{name: "response with parameters", input: dohInput, server: ptr(f.reply(func(q []byte) []byte {
			return dohHTTP(200, "application/dns-message; charset=binary", loopback.DNSResponse(q, 0, nil))
		})), want: dohExpect{state: "completed", outcome: "response", types: dohBootstrapped, rcode: "NOERROR", status: 200}},
		{name: "bootstrap timeout", input: dohSystem, lookup: &net.DNSError{IsTimeout: true, Err: "CANARY-RAW-ERROR"}, want: dohExpect{state: "completed", outcome: "timeout", types: []string{"dns_query"}, failure: "timeout"}},
		{name: "bootstrap not found", input: dohSystem, lookup: &net.DNSError{IsNotFound: true, Err: "CANARY-RAW-ERROR"}, want: dohExpect{state: "completed", outcome: "bootstrap_failure", types: []string{"dns_query"}, failure: "resolver_error"}},
		{name: "tcp timeout", input: dohInput, dial: tcpFail(fmt.Errorf("CANARY-RAW-ERROR: %w", syscall.ETIMEDOUT)), want: dohExpect{state: "completed", outcome: "timeout", types: tls[:1], failure: "timeout"}},
		{name: "tcp refused", input: dohInput, dial: tcpFail(fmt.Errorf("CANARY-RAW-ERROR: %w", syscall.ECONNREFUSED)), want: dohExpect{state: "completed", outcome: "tcp_failure", types: tls[:1], failure: "connection_refused"}},
		{name: "tcp reset", input: dohInput, dial: tcpFail(fmt.Errorf("CANARY-RAW-ERROR: %w", syscall.ECONNRESET)), want: dohExpect{state: "completed", outcome: "tcp_failure", types: tls[:1], failure: "connection_reset"}},
		{name: "tcp unreachable", input: dohInput, dial: tcpFail(fmt.Errorf("CANARY-RAW-ERROR: %w", syscall.ENETUNREACH)), want: dohExpect{state: "completed", outcome: "tcp_failure", types: tls[:1], failure: "network_unreachable"}},
		{name: "tcp unclassified", input: dohInput, dial: tcpFail(errors.New("CANARY-RAW-ERROR")), want: dohExpect{state: "completed", outcome: "inconclusive", types: tls[:1], failure: "inconclusive"}},
		{name: "reset during TLS", input: dohInput, tlsServer: &loopback.TLSConfig{Mode: loopback.TLSReset}, want: dohExpect{state: "completed", outcome: "tls_failure", types: tls, failure: "connection_reset", validation: "not_completed"}},
		{name: "untrusted certificate", input: dohInput, tlsServer: &loopback.TLSConfig{Certificate: untrusted}, want: dohExpect{state: "completed", outcome: "tls_failure", types: tls, failure: "certificate_untrusted", validation: "untrusted"}},
		{name: "resolver name mismatch", input: dohSystem, tlsServer: &loopback.TLSConfig{Certificate: mismatched}, want: dohExpect{state: "completed", outcome: "tls_failure", types: []string{"dns_query", "tcp_connect", "tls_handshake"}, failure: "certificate_name_mismatch", validation: "name_mismatch"}},
		{name: "resolver name mismatch with bootstrap address", input: dohInput, tlsServer: &loopback.TLSConfig{Certificate: mismatched}, want: dohExpect{state: "completed", outcome: "tls_failure", types: tls, failure: "certificate_name_mismatch", validation: "name_mismatch"}},
		{name: "bootstrap address certificate", input: dohInput, tlsServer: &loopback.TLSConfig{Certificate: addressOnly}, want: dohExpect{state: "completed", outcome: "tls_failure", types: tls, failure: "certificate_name_mismatch", validation: "name_mismatch"}},
		{name: "tls alert", input: dohInput, tlsServer: &loopback.TLSConfig{Mode: loopback.TLSAlert, Alert: 40}, want: dohExpect{state: "completed", outcome: "tls_failure", types: tls, failure: "alert", validation: "not_completed"}},
		{name: "tls unclassified", input: dohInput, dial: func(context.Context) (measurement.Conn, error) { return dohReadFails{}, nil }, want: dohExpect{state: "completed", outcome: "inconclusive", types: tls, failure: "inconclusive", validation: "inconclusive"}},
		{name: "tls protocol error", input: dohInput, tlsServer: &loopback.TLSConfig{Mode: loopback.TLSGarbage}, want: dohExpect{state: "completed", outcome: "tls_failure", types: tls, failure: "protocol_error", validation: "not_completed"}},
		{name: "non-success html", input: dohInput, server: ptr(f.reply(fixed(string(dohHTTP(503, "text/html", []byte("body-canary")))))), want: httpFailed("", 503)},
		{name: "non-success dns message", input: dohInput, server: ptr(f.reply(func(q []byte) []byte { return dohHTTP(404, "application/dns-message", loopback.DNSResponse(q, 0, nil)) })), want: httpFailed("", 404)},
		{name: "redirect not followed", input: dohInput, server: ptr(f.reply(fixed("HTTP/1.1 307 Temporary Redirect\r\nLocation: https://other.example/dns-query\r\nContent-Length: 0\r\n\r\n"))), want: httpFailed("", 307)},
		{name: "unsupported encoding", input: dohInput, server: ptr(f.reply(fixed("HTTP/1.1 200 OK\r\nContent-Type: application/dns-message\r\nContent-Encoding: gzip\r\nContent-Length: 4\r\n\r\nbody"))), want: httpFailed("unsupported_content_encoding", 200)},
		{name: "premature body end", input: dohInput, server: ptr(f.reply(fixed("HTTP/1.1 200 OK\r\nContent-Type: application/dns-message\r\nContent-Length: 40\r\n\r\nbody-canary"))), want: httpFailed("protocol_error", 200)},
		{name: "malformed http", input: dohInput, server: ptr(f.reply(fixed("NOT-HTTP body-canary\r\n\r\n"))), want: httpFailed("protocol_error", 0)},
		{name: "unexpected content type", input: dohInput, server: ptr(f.reply(fixed(string(dohHTTP(200, "text/html", []byte("body-canary")))))), want: dohExpect{state: "completed", outcome: "unexpected_content_type", types: dohHTTPStage, failure: "unexpected_content_type", status: 200}},
		{name: "missing content type", input: dohInput, server: ptr(f.reply(func(q []byte) []byte { return dohHTTP(200, "", loopback.DNSResponse(q, 0, nil)) })), want: dohExpect{state: "completed", outcome: "unexpected_content_type", types: dohHTTPStage, failure: "unexpected_content_type", status: 200}},
		{name: "nxdomain", input: dohInput, server: ptr(f.reply(dohReply(3))), want: dnsFailed("dns_nxdomain", "nxdomain", "NXDOMAIN")},
		{name: "servfail", input: dohInput, server: ptr(f.reply(dohReply(2))), want: dnsFailed("dns_servfail", "servfail", "SERVFAIL")},
		{name: "refused", input: dohSystem, server: ptr(f.reply(dohReply(5))), want: dohExpect{state: "completed", outcome: "dns_refused", types: dohResolved, failure: "refused", rcode: "REFUSED", status: 200}},
		{name: "empty message", input: dohInput, server: ptr(f.reply(fixed(string(dohHTTP(200, "application/dns-message", nil))))), want: dnsFailed("malformed_dns_response", "malformed_response", "")},
		{name: "malformed message", input: dohInput, server: ptr(f.reply(func(q []byte) []byte {
			return dohHTTP(200, "application/dns-message", append(q[:2:2], make([]byte, 18)...))
		})), want: dnsFailed("malformed_dns_response", "malformed_response", "")},
		{name: "query identifier mismatch", input: dohInput, server: ptr(f.reply(func(q []byte) []byte {
			r := loopback.DNSResponse(q, 0, nil)
			r[0] ^= 0xff
			return dohHTTP(200, "application/dns-message", r)
		})), want: dnsFailed("query_mismatch", "query_mismatch", "")},
		{name: "query name mismatch", input: dohInput, server: ptr(f.reply(func(q []byte) []byte {
			r := loopback.DNSResponse(q, 0, nil)
			r[13] = 'x' // first byte of the question's first label
			return dohHTTP(200, "application/dns-message", r)
		})), want: dnsFailed("query_mismatch", "query_mismatch", "")},
		{name: "query type mismatch", input: dohInput, server: ptr(f.reply(func(q []byte) []byte {
			r := loopback.DNSResponse(q, 0, nil)
			r[len(r)-3] = 28
			return dohHTTP(200, "application/dns-message", r)
		})), want: dnsFailed("query_mismatch", "query_mismatch", "")},
		{name: "unregistered rcode", input: dohInput, server: ptr(f.reply(dohReply(4))), want: dnsFailed("inconclusive", "inconclusive", "NOTIMP")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, d := dohHarness(t, f, tc.input)
			var srv *loopback.HTTPServer
			if tc.server != nil {
				srv = dohServe(t, d, *tc.server)
			}
			if tc.tlsServer != nil {
				dohServeTLS(t, d, *tc.tlsServer)
			}
			if tc.dial != nil {
				d.dial = tc.dial
			}
			if tc.lookup != nil {
				h.adapters.run.resolver = dnsSystemResolver{func(context.Context, string, string) ([]netip.Addr, error) { d.lookups.Add(1); return nil, tc.lookup }}
			}
			c := h.initialized(t)
			defer c.Close()
			dohResult(t, tcpRaw(t, h, c, dohTask("normalize", tc.input, `{}`)), tc.want)
			if _, err := c.NextBundle(); err != nil {
				t.Fatal("result did not survive bundle validation", err)
			}
			if tc.lookup != nil && d.calls.Load() != 0 {
				t.Fatal("contact after failed bootstrap")
			}
			if srv != nil && (d.calls.Load() != 1 || len(srv.Counts().Requests) != 1) {
				t.Fatal("not exactly one request: retry, fallback or redirect contact")
			}
		})
	}
}

// M-DNS-03-LIMITS: per-stage operation bounds while the outer deadline stays
// open, the 65,535-byte message bound, a stricter common received limit,
// the 16-answer bound, and one connection and request with no retry.
// requirement-ref: REQ-DNS-03-0023
// requirement-ref: REQ-LIMIT-0034
func TestMDns03Limits(t *testing.T) {
	f := newDohFixture(t)
	silent := f.reply(func([]byte) []byte { return nil })
	silent.Stall = true
	requested := func(c loopback.HTTPCounts) bool { return len(c.Requests) == 1 }
	t.Run("tls operation", func(t *testing.T) {
		h, d := dohHarness(t, f, dohInput)
		srv := dohServeTLS(t, d, loopback.TLSConfig{Mode: loopback.TLSStall})
		c := h.initialized(t)
		defer c.Close()
		data := strings.Replace(dohTask("limit", dohInput, `{}`), `"limits":{}`, `"limits":{"max_duration_ms":60000}`, 1)
		rh, err := c.StartRun(h.admit(t, c, data))
		if err != nil {
			t.Fatal(codeOf(err))
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Wait(ctx, func(c loopback.TLSCounts) bool { return c.In > 0 }); err != nil {
			t.Fatal(err)
		}
		h.advance(15 * time.Second)
		waitTerminal(t, c, rh)
		dohResult(t, h.run.output.Records()[0], dohExpect{state: "completed", outcome: "timeout", types: []string{"tcp_connect", "tls_handshake"}, failure: "timeout", validation: "not_completed"})
		if err := srv.WaitIdle(ctx); err != nil || d.calls.Load() != 1 || srv.Counts().Accepted != 1 {
			t.Fatal("stalled session not released or retried")
		}
	})
	for _, tc := range []struct {
		name, limits string
		advance      time.Duration
		want         dohExpect
	}{
		{"http operation", `"max_duration_ms":60000`, 15 * time.Second, dohExpect{state: "completed", outcome: "timeout", types: dohHTTPStage, failure: "timeout"}},
		{"attempt deadline", `"max_duration_ms":1000`, time.Second, dohExpect{state: "incomplete", reason: "attempt_deadline_exceeded"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, d := dohHarness(t, f, dohInput)
			srv := dohServe(t, d, silent)
			c := h.initialized(t)
			defer c.Close()
			data := strings.Replace(dohTask("limit", dohInput, `{}`), `"limits":{}`, `"limits":{`+tc.limits+`}`, 1)
			rh, err := c.StartRun(h.admit(t, c, data))
			if err != nil {
				t.Fatal(codeOf(err))
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.Wait(ctx, requested); err != nil {
				t.Fatal(err)
			}
			h.advance(tc.advance)
			waitTerminal(t, c, rh)
			dohResult(t, h.run.output.Records()[0], tc.want)
			if err := srv.Wait(ctx, func(c loopback.HTTPCounts) bool { return c.Open == 0 }); err != nil || d.calls.Load() != 1 || srv.Counts().Accepted != 1 {
				t.Fatal("stalled session not released or retried")
			}
		})
	}
	seventeen := make([]loopback.DNSRecord, 17)
	for i := range seventeen {
		seventeen[i] = loopback.DNSRecord{Type: "A", Value: fmt.Sprintf("9.9.9.%d", i+1)}
	}
	for _, tc := range []struct {
		name, limits string
		reply        func([]byte) []byte
		want         dohExpect
	}{
		{"message below bound", `"max_bytes_received":262144`, dohSized(65534), dohResponse},
		{"message at bound", `"max_bytes_received":262144`, dohSized(65535), dohResponse},
		{"message above bound", `"max_bytes_received":262144`, dohSized(65536), dohExpect{state: "completed", outcome: "malformed_dns_response", types: dohBootstrapped, failure: "malformed_response", status: 200}},
		{"oversized message not read", `"max_bytes_received":262144`, dohSized(200000), dohExpect{state: "completed", outcome: "malformed_dns_response", types: dohBootstrapped, failure: "malformed_response", status: 200}},
		{"message above received limit", `"max_bytes_received":65536`, dohSized(65535), dohExpect{state: "incomplete", reason: "resource_limit"}},
		{"fifteen answers", `"max_bytes_received":4096`, dohReply(0, seventeen[:15]...), dohExpect{state: "completed", outcome: "response", types: dohBootstrapped, rcode: "NOERROR", answers: 15, status: 200}},
		{"sixteen answers", `"max_bytes_received":4096`, dohReply(0, seventeen[:16]...), dohExpect{state: "completed", outcome: "response", types: dohBootstrapped, rcode: "NOERROR", answers: 16, status: 200}},
		{"seventeen answers", `"max_bytes_received":4096`, dohReply(0, seventeen...), dohExpect{state: "completed", outcome: "malformed_dns_response", types: dohBootstrapped, failure: "malformed_response", status: 200}},
		{"one connection", `"max_connections":1`, dohAnswer, dohResponse},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, d := dohHarness(t, f, dohInput)
			srv := dohServe(t, d, f.reply(tc.reply))
			c := h.initialized(t)
			defer c.Close()
			data := strings.Replace(dohTask("bound", dohInput, `{}`), `"limits":{}`, `"limits":{`+tc.limits+`}`, 1)
			dohResult(t, tcpRaw(t, h, c, data), tc.want)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.Wait(ctx, func(c loopback.HTTPCounts) bool { return c.Open == 0 }); err != nil || d.calls.Load() != 1 || srv.Counts().Accepted != 1 || len(srv.Counts().Requests) != 1 {
				t.Fatal("not exactly one connection and request")
			}
		})
	}
}

// An equivalent resolver and query tuple waits while one run is active;
// cancelling the waiting run starts no DNS-03 procedure.
// requirement-ref: REQ-DNS-03-0023
func TestMDns03TupleConcurrency(t *testing.T) {
	f := newDohFixture(t)
	h, d := dohHarness(t, f, dohInput)
	srv := dohServeTLS(t, d, loopback.TLSConfig{Mode: loopback.TLSStall})
	c := h.initialized(t)
	defer c.Close()
	first, err := c.StartRun(h.admit(t, c, dohTask("tuple-first", dohInput, `{}`)))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Wait(ctx, func(c loopback.TLSCounts) bool { return c.In > 0 }); err != nil {
		t.Fatal(err)
	}
	second, err := c.StartRun(h.admit(t, c, dohTask("tuple-second", dohInput, `{"address_family":"ipv4"}`)))
	if err != nil {
		t.Fatal(err)
	}
	if err := c.CancelRun(second); err != nil {
		t.Fatal(err)
	}
	waitTerminal(t, c, second)
	if d.calls.Load() != 1 {
		t.Fatal("equivalent tuple contacted concurrently")
	}
	rows := h.run.output.Records()
	if len(rows) != 1 {
		t.Fatalf("waiting run produced %d results, want 1", len(rows))
	}
	dohResult(t, rows[0], dohExpect{state: "not_run", reason: "cancelled_by_host"})
	if err := c.CancelRun(first); err != nil {
		t.Fatal(err)
	}
	waitTerminal(t, c, first)
	if rows = h.run.output.Records(); len(rows) != 2 {
		t.Fatalf("cancelled runs produced %d results, want 2", len(rows))
	}
	dohResult(t, rows[1], dohExpect{state: "incomplete", reason: "cancelled_by_host"})
	if err := srv.WaitIdle(ctx); err != nil || srv.Counts().Accepted != 1 {
		t.Fatal("stalled session not released")
	}
}

// M-DNS-03-TERMINATION through Client after the POST is read, and a refused
// proxy tunnel, applying the GO-SPEC-ISSUE-004 table; internal/execution
// covers every reason at every Procedure point.
// requirement-ref: REQ-DNS-03-0028
// requirement-ref: REQ-RESULT-0051
func TestMDns03Termination(t *testing.T) {
	f := newDohFixture(t)
	silent := f.reply(func([]byte) []byte { return nil })
	silent.Stall = true
	for _, reason := range []string{"network_changed", "cancelled_by_host"} {
		t.Run(reason, func(t *testing.T) {
			h, d := dohHarness(t, f, dohInput)
			srv := dohServe(t, d, silent)
			c := h.initialized(t)
			defer c.Close()
			rh, err := c.StartRun(h.admit(t, c, dohTask("terminate", dohInput, `{}`)))
			if err != nil {
				t.Fatal(codeOf(err))
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.Wait(ctx, func(c loopback.HTTPCounts) bool { return len(c.Requests) == 1 }); err != nil {
				t.Fatal(err)
			}
			if reason == "network_changed" {
				a := defaultAccess
				a.Generation++
				h.run.network.Set(a)
			} else if err := c.CancelRun(rh); err != nil {
				t.Fatal(codeOf(err))
			}
			waitTerminal(t, c, rh)
			dohResult(t, h.run.output.Records()[0], dohExpect{state: "incomplete", reason: reason})
			if err := srv.Wait(ctx, func(c loopback.HTTPCounts) bool { return c.Open == 0 }); err != nil || srv.Counts().Accepted != 1 || len(srv.Counts().Requests) != 1 {
				t.Fatal("socket not closed, retried or written after the interruption")
			}
		})
	}
	t.Run("proxy refused tunnel", func(t *testing.T) {
		h, d := dohHarness(t, f, dohInput)
		p, err := loopback.ListenProxy(loopback.ProxyConfig{Mode: loopback.ProxyReply, Head: []byte("HTTP/1.1 502 Bad Gateway\r\n\r\n")})
		if err != nil {
			t.Fatal(err)
		}
		defer p.Close()
		data := strings.Replace(dohTask("proxy-termination", dohInput, `{}`), `"limits":{}`, `"limits":{},"network_path":"proxy"`, 1)
		_, raw, c := runOnce(t, h, proxyConfig(p.Port()), data)
		defer c.Close()
		dohResult(t, raw, dohExpect{state: "incomplete", reason: "proxy_connect_failed"})
		if d.calls.Load() != 0 || d.lookups.Load() != 0 {
			t.Fatal("fell back to the direct path or the system resolver")
		}
	})
}

// M-DNS-03-VERSIONS: exact task and result tuples, strict receiver rejection
// and no cross-test interpretation.
// requirement-ref: REQ-DNS-03-0024
func TestMDns03Versions(t *testing.T) {
	f := newDohFixture(t)
	for _, version := range []string{`"0.1.1"`, `"0.2.0"`, `"1.0.0"`, `""`, `"v0.1.0"`, `"0.1"`, `"0.1.0-draft"`, `"0.1.0 "`, `0.1`, `null`, "missing"} {
		h, d := dohHarness(t, f, dohInput)
		c := h.initialized(t)
		replacement := `"test_version":` + version + `,`
		if version == "missing" {
			replacement = ""
		}
		data := strings.Replace(dohTask("version", dohInput, `{}`), `"test_version":"0.1.0",`, replacement, 1)
		want := "task_schema_invalid"
		if version == `"0.1.1"` || version == `"0.2.0"` || version == `"1.0.0"` {
			want = "task_unsupported"
		}
		handle, err := c.AdmitTask(data, h.embedded(t, data))
		if codeOf(err) != want || handle != "" || h.replay.Commits() != 0 || d.calls.Load() != 0 || d.lookups.Load() != 0 || len(h.run.output.Records()) != 0 {
			t.Fatalf("version %s: %s, want %s", version, codeOf(err), want)
		}
		c.Close()
	}
	h, d := dohHarness(t, f, dohInput)
	dohServe(t, d, f.reply(dohAnswer))
	c := h.initialized(t)
	defer c.Close()
	raw := tcpRaw(t, h, c, dohTask("versions", dohInput, `{}`))
	cd, _ := observation.NewContextDecoder(result.Contract, observation.ContextApprovals{})
	path, _ := observation.NewPath(1, 0, "direct", "")
	binding := func(o result.Output) *result.Decoder {
		o.Limits.RetainDNSAnswers = true
		d, err := result.New(result.Binding{TaskID: "versions", Output: o, MaxAttempts: 1, MaxDurationMS: 60000, Implementation: implementation, ContextDecoder: cd, Path: path, AccessGeneration: 1})
		if err != nil {
			t.Fatal(err)
		}
		return d
	}
	output := result.Output{Test: model.Test{ID: "DNS-03", Version: "0.1.0", ResultVersion: "0.1.0"}, QueryType: "A"}
	output.Limits.MaxBodyBytes = dns03.MaxBody
	decoder := binding(output)
	if _, err := decoder.Decode(raw); err != nil {
		t.Fatal("valid result rejected", err)
	}
	// The same bytes are not a DNS-04 or DNS-01 result, and a bootstrap
	// binding requires the system observation first.
	bootstrap := output
	bootstrap.Bootstrap = true
	for _, o := range []result.Output{
		{Test: model.Test{ID: "DNS-04", Version: "0.1.0", ResultVersion: "0.1.0"}, QueryType: "A"},
		{Test: model.Test{ID: "DNS-01", Version: "0.1.0", ResultVersion: "0.1.0"}, QueryType: "A", DNSMode: "tcp"},
		bootstrap,
	} {
		if _, err := binding(o).Decode(raw); err == nil {
			t.Fatalf("DNS-03 output accepted as %+v", o.Test)
		}
	}
	for _, field := range []string{"schema_version", "version", "result_version"} {
		for _, v := range []any{"0.1.1", "0.2.0", "1.0.0", "0.1", "v0.1.0", "0.1.0-draft", "", "0.1.0 ", 0.1, nil} {
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
	obs := func(v map[string]any, i int) map[string]any {
		return v["measurement"].(map[string]any)["observations"].([]any)[i].(map[string]any)
	}
	for i, mutate := range []func(map[string]any){
		func(v map[string]any) { v["test"].(map[string]any)["id"] = "DNS-04" },
		func(v map[string]any) { v["test"].(map[string]any)["id"] = "DNS-01" },
		func(v map[string]any) { v["test"].(map[string]any)["extra"] = true },
		func(v map[string]any) { v["measurement"].(map[string]any)["extra"] = true },
		func(v map[string]any) { v["measurement"].(map[string]any)["outcome"] = "resolved" },
		func(v map[string]any) { v["measurement"].(map[string]any)["outcome"] = "blocked" },
		func(v map[string]any) { obs(v, 3)["details"].(map[string]any)["engine"] = "dot" },
		func(v map[string]any) { obs(v, 3)["details"].(map[string]any)["engine"] = "system" },
		func(v map[string]any) { obs(v, 3)["details"].(map[string]any)["query_type"] = "AAAA" },
		func(v map[string]any) { obs(v, 3)["details"].(map[string]any)["resolver_url"] = dohURL },
		func(v map[string]any) { obs(v, 2)["details"].(map[string]any)["headers"] = map[string]any{} },
		func(v map[string]any) { obs(v, 1)["details"].(map[string]any)["server_name"] = dohName },
		func(v map[string]any) { obs(v, 0)["details"].(map[string]any)["port"] = 443 },
		func(v map[string]any) {
			m := v["measurement"].(map[string]any)
			m["observations"] = []any{obs(v, 0), obs(v, 1), obs(v, 3)}
		},
		func(v map[string]any) {
			m := v["measurement"].(map[string]any)
			m["observations"] = append([]any{obs(v, 3)}, m["observations"].([]any)...)
		},
	} {
		var v map[string]any
		_ = json.Unmarshal(raw, &v)
		mutate(v)
		data, _ := json.Marshal(v)
		if _, err := decoder.Decode(data); err == nil {
			t.Fatalf("invalid output %d accepted: %s", i, data)
		}
	}
}

const (
	dohCanaryName  = "canary-doh-121.example"
	dohCanaryInput = `{"resolver_url":"https://canary-doh-121.example/canary-dohpath-122","hostname":"canary-dohq-123.example","record_type":"A","bootstrap_ip":"8.8.4.4"}`
	dohCanarySys   = `{"resolver_url":"https://canary-doh-121.example/canary-dohpath-122","hostname":"canary-dohq-123.example","record_type":"A"}`
)

// dohCanaries are the REQ-DNS-03-0029 items; the scripted bootstrap resolver
// returns the canary resolver address, distinct from the canary bootstrap_ip.
var dohCanaries = canarySet{
	{"resolver_url path", "canary-dohpath-122"}, {"resolver hostname", dohCanaryName}, {"resolver address", "8.8.4.5"}, {"bootstrap_ip", "8.8.4.4"},
	{"query hostname", "canary-dohq-123.example"}, {"query identifier", "48879"}, {"DNS authority record", "canary-authority-74.example"},
	{"DNS additional record", "canary-additional-75.example"}, {"DNS additional address", "203.0.113.75"}, {"raw DNS message pattern", "CANARY-RAWDNS-BYTES"}, {"answer TTL", "86413"},
	{"HTTP header name", "x-canary-doh-124"}, {"HTTP header value", "canary-dohval-125"}, {"Set-Cookie value", "canary-dohcookie-126"}, {"Content-Type value", "application/canary-ctype-127"},
	{"HTTP body beyond the DNS message", "CANARY-TRAILING-BODY"}, {"HTTP status text", "CANARY-DOH-REASON"},
	{"certificate subject and issuer", loopback.CertificateCanary}, {"leaf serial number", "424242"}, {"issuer serial number", "91337"},
	{"raw transport error", "CANARY-RAW-ERROR"}, {"raw x509 error", "x509:"}, {"raw TLS error", "tls: "},
}

func dohCanaryHTTP(status int, contentType string, body []byte) []byte {
	head := fmt.Sprintf("HTTP/1.1 %d CANARY-DOH-REASON\r\nX-Canary-Doh-124: canary-dohval-125\r\nSet-Cookie: canary-dohcookie-126=1\r\nContent-Type: %s\r\nContent-Length: %d\r\n\r\n", status, contentType, len(body))
	return append([]byte(head), body...)
}

var dohAllowed = []string{"outcome", "observations[].type", "observations[].status", "observations[].duration_ms", "observations[].failure",
	"observations[].details.engine", "observations[].details.address_family", "observations[].details.query_type", "observations[].details.response_limit_reached",
	"observations[].details.responses[].arrival_index", "observations[].details.responses[].arrival_offset_ms", "observations[].details.responses[].rcode",
	"observations[].details.responses[].answers[].type", "observations[].details.responses[].answers[].value",
	"observations[].details.tls_version", "observations[].details.alpn", "observations[].details.certificate_validation", "observations[].details.alert",
	"observations[].details.protocol", "observations[].details.status_code", "observations[].details.body.bytes_processed", "observations[].details.body.truncated",
	"observations[].details.content_encoding"}

type dohCanaryCase struct {
	name, input, outcome string
	reply                func([]byte) []byte
	tlsServer            *loopback.TLSConfig
	dial                 error
	lookup               error
	cancel               bool
}

func dohCanaryCases(t *testing.T, f dohFixture) []dohCanaryCase {
	other, err := loopback.NewTestCA()
	if err != nil {
		t.Fatal(err)
	}
	untrusted, err := other.Issue(dohCanaryName, false)
	if err != nil {
		t.Fatal(err)
	}
	dns := func(code dnsmessage.RCode, id uint16, answers ...string) func([]byte) []byte {
		return func(q []byte) []byte {
			return dohCanaryHTTP(200, "application/dns-message", dnsCanaryResponse(q, id, code, answers...))
		}
	}
	return []dohCanaryCase{
		{name: "response with bootstrap_ip", input: dohCanaryInput, outcome: "response", reply: dns(dnsmessage.RCodeSuccess, 0, "9.9.9.9", "9.9.9.9", "1.1.1.1")},
		{name: "response with system bootstrap", input: dohCanarySys, outcome: "response", reply: dns(dnsmessage.RCodeSuccess, 0, "9.9.9.9")},
		{name: "bootstrap_failure", input: dohCanarySys, outcome: "bootstrap_failure", lookup: &net.DNSError{IsNotFound: true, Err: "CANARY-RAW-ERROR", Name: dohCanaryName, Server: "127.0.0.1:53"}},
		{name: "tcp_failure", input: dohCanaryInput, outcome: "tcp_failure", dial: fmt.Errorf("CANARY-RAW-ERROR 8.8.4.4:443: %w", syscall.ECONNREFUSED)},
		{name: "tls_failure", input: dohCanaryInput, outcome: "tls_failure", tlsServer: &loopback.TLSConfig{Certificate: untrusted}},
		{name: "http_failure", input: dohCanaryInput, outcome: "http_failure", reply: func(q []byte) []byte {
			return dohCanaryHTTP(503, "text/html", append(dnsCanaryResponse(q, 0, dnsmessage.RCodeSuccess), "CANARY-TRAILING-BODY"...))
		}},
		{name: "unexpected_content_type", input: dohCanaryInput, outcome: "unexpected_content_type", reply: func(q []byte) []byte {
			return dohCanaryHTTP(200, "application/canary-ctype-127", dnsCanaryResponse(q, 0, dnsmessage.RCodeSuccess, "9.9.9.9"))
		}},
		{name: "dns_nxdomain", input: dohCanaryInput, outcome: "dns_nxdomain", reply: dns(dnsmessage.RCodeNameError, 0)},
		{name: "dns_servfail", input: dohCanaryInput, outcome: "dns_servfail", reply: dns(dnsmessage.RCodeServerFailure, 0)},
		{name: "dns_refused", input: dohCanarySys, outcome: "dns_refused", reply: dns(dnsmessage.RCodeRefused, 0)},
		{name: "malformed_dns_response", input: dohCanaryInput, outcome: "malformed_dns_response", reply: func(q []byte) []byte {
			return dohCanaryHTTP(200, "application/dns-message", append(dnsCanaryResponse(q, 0, dnsmessage.RCodeSuccess, "9.9.9.9"), "CANARY-TRAILING-BODY"...))
		}},
		{name: "query_mismatch", input: dohCanaryInput, outcome: "query_mismatch", reply: dns(dnsmessage.RCodeSuccess, 48879, "9.9.9.9")},
		{name: "timeout", input: dohCanaryInput, outcome: "timeout", dial: fmt.Errorf("CANARY-RAW-ERROR 8.8.4.4:443: %w", syscall.ETIMEDOUT)},
		{name: "inconclusive", input: dohCanaryInput, outcome: "inconclusive", reply: dns(dnsmessage.RCodeNotImplemented, 0)},
		{name: "incomplete", input: dohCanaryInput, cancel: true},
	}
}

// dohCanaryRun runs one DNS-03 path against the canary resolver profile and
// returns its surfaces and the canaries of the raw DNS messages it carried.
func dohCanaryRun(t *testing.T, f dohFixture, id string, tc dohCanaryCase) (*surfaces, canarySet) {
	t.Helper()
	h, d := dohHarness(t, f, tc.input)
	d.authorities = []string{"8.8.4.4:443", "8.8.4.5:443"}
	h.adapters.run.resolver = dnsSystemResolver{func(_ context.Context, _, name string) ([]netip.Addr, error) {
		if tc.lookup != nil {
			return nil, tc.lookup
		}
		if name != dohCanaryName {
			return nil, errors.New("synthetic: unexpected name")
		}
		return []netip.Addr{netip.MustParseAddr("8.8.4.5")}, nil
	}}
	var mu sync.Mutex
	var messages canarySet
	var srv *loopback.HTTPServer
	switch {
	case tc.reply != nil || tc.cancel:
		cfg := f.cert
		cfg.DoH = func(q []byte) []byte {
			mu.Lock()
			messages = append(messages, binaryCanaries(fmt.Sprint("raw DNS query ", len(messages)), q)...)
			mu.Unlock()
			if tc.cancel {
				return nil
			}
			r := tc.reply(q)
			mu.Lock()
			messages = append(messages, binaryCanaries(fmt.Sprint("raw HTTP response ", len(messages)), r)...)
			mu.Unlock()
			return r
		}
		cfg.Stall = tc.cancel
		srv = dohServe(t, d, cfg)
	case tc.tlsServer != nil:
		dohServeTLS(t, d, *tc.tlsServer)
	case tc.dial != nil:
		d.dial = func(context.Context) (measurement.Conn, error) { return nil, tc.dial }
	}
	c := h.initialized(t)
	defer c.Close()
	s := &surfaces{}
	rh, err := c.StartRun(h.admit(t, c, dohTask(id, tc.input, `{}`)))
	if err != nil {
		t.Fatal(codeOf(err))
	}
	if tc.cancel {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Wait(ctx, func(c loopback.HTTPCounts) bool { return len(c.Requests) == 1 }); err != nil {
			t.Fatal(err)
		}
		if err := c.CancelRun(rh); err != nil {
			t.Fatal(codeOf(err))
		}
	}
	waitTerminal(t, c, rh)
	_, err = c.RunState(rh)
	s.err(err)
	s.collect(h, c)
	rejectedClean(t, h, c, dohCanaries, withUnknownInput(dohTask("rejected", tc.input, `{}`), dohCanaries))
	if len(s.results) != 1 {
		t.Fatal("not one result")
	}
	var r struct {
		Execution   struct{ State string }
		Measurement *struct{ Outcome string }
	}
	_ = json.Unmarshal(s.results[0], &r)
	if tc.cancel && (r.Execution.State != "incomplete" || r.Measurement != nil) || !tc.cancel && (r.Measurement == nil || r.Measurement.Outcome != tc.outcome) {
		t.Fatalf("%s: unexpected outcome", tc.name)
	}
	mu.Lock()
	defer mu.Unlock()
	return s, slices.Clone(messages)
}

// M-DNS-03-MINIMIZATION: canaries for every REQ-DNS-03-0029 item over the
// success path with and without bootstrap_ip, each failure class and an
// incomplete termination; only the registered observation fields remain and
// retained answers are bounded, deduplicated A values without TTLs.
// requirement-ref: REQ-DNS-03-0029
func TestMDns03Minimization(t *testing.T) {
	dohCanaries.unique(t)
	f := newDohFixture(t)
	cert, err := f.ca.Issue(dohCanaryName, false)
	if err != nil {
		t.Fatal(err)
	}
	f.cert.Certificate = cert
	for _, tc := range dohCanaryCases(t, f) {
		t.Run(tc.name, func(t *testing.T) {
			s, messages := dohCanaryRun(t, f, "minimization", tc)
			if tc.reply != nil && len(messages) == 0 {
				t.Fatal("no message canaries")
			}
			s.check(t, append(slices.Clone(dohCanaries), messages...))
			m := minimized(t, s.results[0], dohAllowed)
			if tc.name == "response with bootstrap_ip" {
				obs := m["observations"].([]any)
				answers := obs[len(obs)-1].(map[string]any)["details"].(map[string]any)["responses"].([]any)[0].(map[string]any)["answers"].([]any)
				if len(answers) != 2 || answers[0].(map[string]any)["value"] != "9.9.9.9" || answers[1].(map[string]any)["value"] != "1.1.1.1" {
					t.Fatal("answers not bounded, deduplicated and ordered")
				}
			}
		})
	}
	t.Run("injected fields", func(t *testing.T) {
		s, _ := dohCanaryRun(t, f, "minimization", dohCanaryCases(t, f)[0])
		obs := func(v map[string]any, i int) map[string]any {
			o := v["measurement"].(map[string]any)["observations"].([]any)
			return o[(i+len(o))%len(o)].(map[string]any)
		}
		var inject []func(map[string]any)
		for _, field := range []string{"resolver_url", "hostname", "ip_address", "bootstrap_ip", "query_id", "ttl", "message", "headers", "content_type", "cookie", "certificate", "error"} {
			inject = append(inject, func(v map[string]any) { obs(v, -1)["details"].(map[string]any)[field] = "synthetic" },
				func(v map[string]any) { obs(v, -2)["details"].(map[string]any)[field] = "synthetic" },
				func(v map[string]any) {
					obs(v, -1)["details"].(map[string]any)["responses"].([]any)[0].(map[string]any)["answers"].([]any)[0].(map[string]any)[field] = 86413
				})
		}
		d := primitiveDecoder(t, "minimization", measurementOutput(dohTest, dohParams(t, dohCanaryInput, `{}`)))
		assertRejected(t, d, s.results[0], inject...)
	})
}

// M-DNS-03-INTERPRETATION: the shared prohibited-claims check with the
// DNS-03 claims over every outcome, including a response whose answers
// differ from a synthetic system answer set, and each failure attributed to
// the observed stage by observation order and the terminal observation.
// requirement-ref: REQ-DNS-03-0001
// requirement-ref: REQ-DNS-03-0002
// requirement-ref: REQ-DNS-03-0003
// requirement-ref: REQ-DNS-03-0030
// requirement-ref: REQ-DNS-03-0031
// requirement-ref: REQ-DNS-03-0032
func TestMDns03Interpretation(t *testing.T) {
	f := newDohFixture(t)
	cert, err := f.ca.Issue(dohCanaryName, false)
	if err != nil {
		t.Fatal(err)
	}
	f.cert.Certificate = cert
	stage := map[string]string{"bootstrap_failure": "dns_query", "tcp_failure": "tcp_connect", "timeout": "tcp_connect", "tls_failure": "tls_handshake", "http_failure": "http_exchange",
		"unexpected_content_type": "http_exchange", "dns_nxdomain": "dns_query", "dns_servfail": "dns_query", "dns_refused": "dns_query", "malformed_dns_response": "dns_query",
		"query_mismatch": "dns_query", "inconclusive": "dns_query"}
	var raws [][]byte
	for _, tc := range dohCanaryCases(t, f) {
		s, _ := dohCanaryRun(t, f, "interpretation", tc)
		raws = append(raws, s.results...)
		var r struct {
			Measurement *struct {
				Observations []struct{ Type, Status string }
			}
		}
		_ = json.Unmarshal(s.results[0], &r)
		if want, ok := stage[tc.outcome]; ok {
			obs := r.Measurement.Observations
			last := obs[len(obs)-1]
			if last.Type != want {
				t.Errorf("%s attributed to %s", tc.outcome, last.Type)
			}
			for _, o := range obs[:len(obs)-1] {
				if o.Status != "succeeded" {
					t.Errorf("%s: a failure before the terminal stage", tc.outcome)
				}
			}
		}
	}
	// The DoH answer 9.9.9.9 differs from the synthetic system answer set
	// {1.0.0.1}; the result carries no comparison, interference or
	// correctness field either way.
	d := primitiveDecoder(t, "interpretation", measurementOutput(dohTest, dohParams(t, dohCanaryInput, `{}`)))
	interpretation(t, "DNS-03", d, raws[0], raws, []string{"response", "bootstrap_failure", "tcp_failure", "tls_failure", "http_failure", "unexpected_content_type",
		"dns_nxdomain", "dns_servfail", "dns_refused", "malformed_dns_response", "query_mismatch", "timeout", "inconclusive"}, claims(
		"DNS-03 configures DoH for the device or host", `\b(dns-03|the test|doh test)\b[^.]*\bconfigur\w*\b[^.]*\b(device|host application|operating system)\b`,
		"DoH result is authoritative or correct", `\b(doh|dns-03)\b[^.]*\b(answer|result|response)s?\b[^.]*\b(is|are)\s+(authoritative|correct)\b`,
		"difference from system DNS establishes interference", `\bdifferen\w*\b[^.]*\bsystem dns\b[^.]*\b(establish\w*|indicat\w*|shows?|proves?|means?)\b[^.]*\binterfer\w*`,
		"stage failure is DoH blocking", `\b(bootstrap|tcp|tls|http|dns)\s+failure\b[^.]*\b(is|indicates?|means?|establish\w*)\b[^.]*\b(doh blocking|blocking|censorship)\b`,
		"answers enriched", `\banswers?\b[^.]*\b(enrich\w*|annotat\w*)\b[^.]*\b(asn|organization|geolocation|reputation|verdict)`,
	),
		"establishes whether each observed bootstrap, tcp, tls, http and dns stage completed for the configured resolver and query",
		"does not configure doh for the device or the host application", "a doh answer is not authoritative or correct",
		"a difference from system dns does not establish interference", "is not summarized as doh blocking or censorship",
		"attributed only to the directly observed stage", "answers carry no on-device asn, organization, geolocation, reputation or verdict enrichment",
		"stale or incorrect registered resolver configuration")
}
