package barometer

import (
	"context"
	"encoding/binary"
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
	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/measurement/dns04"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

var dotTest = task.TestVersion{ID: "DNS-04", Version: "0.1.0"}

const (
	dotName       = "dot.resolver.example"
	dotInput      = `{"resolver_name":"dot.resolver.example","port":853,"hostname":"query.example","record_type":"A","bootstrap_ip":"8.8.8.8"}`
	dotSystem     = `{"resolver_name":"dot.resolver.example","port":853,"hostname":"query.example","record_type":"A"}`
	dotAuthority  = "8.8.8.8:853"
	dotQueryBytes = 31 // header, query.example and question tail
)

func dotParams(t *testing.T, input, options string) dns04.Parameters {
	t.Helper()
	c := dns04.Codec{}
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
	return p.(dns04.Parameters)
}

// dotPolicy registers the query name and the complete resolver profile that
// the production target rule extracts from input.
func dotPolicy(t *testing.T, h *harness, input string) {
	t.Helper()
	targets, ok := dns04Targets{}.Targets(dotParams(t, input, `{}`))
	if !ok || len(targets) != 2 {
		t.Fatal("target rule")
	}
	p := synPolicyInput(2, "reg-a", func(p *task.Permissions) {
		p.Tests = []task.TestVersion{dotTest}
		p.RegistrationIDs = []string{"reg-a", "reg-b"}
	})
	resolver := p.Targets[0].Clone()
	resolver.ID, resolver.Role, resolver.Tuple = "reg-b", targets[1].Role, targets[1].Tuple
	p.Targets[0].Test, p.Targets[0].Role, p.Targets[0].Tuple = dotTest, targets[0].Role, targets[0].Tuple
	resolver.Test = dotTest
	p.Targets = append(p.Targets, resolver)
	var e error
	if h.adapters.policy, e = task.NewLocalPolicy(p); e != nil {
		t.Fatal(e)
	}
}

// dotTransport maps the checked resolver authority to the loopback server and
// fails any other destination; lookups counts system bootstrap requests.
type dotTransport struct {
	calls   atomic.Int32
	lookups atomic.Int32
	dial    func(context.Context) (measurement.Conn, error)
}

func (d *dotTransport) Dialer(int) (measurement.Dialer, error) { return d, nil }
func (d *dotTransport) DialContext(ctx context.Context, network, address string) (measurement.Conn, error) {
	d.calls.Add(1)
	if !strings.HasPrefix(network, "tcp") || address != dotAuthority {
		return nil, errors.New("CANARY-RAW-ERROR")
	}
	return d.dial(ctx)
}

type dotFixture struct {
	ca   *loopback.TestCA
	cert loopback.TLSConfig
}

func newDotFixture(t *testing.T) dotFixture {
	t.Helper()
	ca, err := loopback.NewTestCA()
	if err != nil {
		t.Fatal(err)
	}
	cert, err := ca.Issue(dotName, false)
	if err != nil {
		t.Fatal(err)
	}
	return dotFixture{ca, loopback.TLSConfig{Certificate: cert}}
}

func dotHarness(t *testing.T, f dotFixture, input string) (*harness, *dotTransport) {
	t.Helper()
	h := runningHarness(t)
	h.adapters.registry = nil
	h.adapters.measurementRoots = f.ca.Pool()
	h.adapters.run.paths = nil
	h.adapters.approvals = config.Approvals{ConnectionCodes: []string{"outline_mobileproxy"}}
	dotPolicy(t, h, input)
	d := &dotTransport{dial: func(context.Context) (measurement.Conn, error) { return nil, syscall.ECONNREFUSED }}
	h.adapters.run.directDialers = d
	h.adapters.run.resolver = dnsSystemResolver{func(ctx context.Context, family, name string) ([]netip.Addr, error) {
		d.lookups.Add(1)
		if family != "ip4" || name != dotName {
			return nil, errors.New("CANARY-RAW-ERROR")
		}
		return []netip.Addr{netip.MustParseAddr("8.8.8.8")}, nil
	}}
	return h, d
}

func dotServe(t *testing.T, d *dotTransport, cfg loopback.TLSConfig) *loopback.TLSServer {
	t.Helper()
	srv, err := loopback.ListenTLS(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = srv.Close() })
	d.dial = func(ctx context.Context) (measurement.Conn, error) {
		var nd net.Dialer
		return nd.DialContext(ctx, "tcp", srv.Address())
	}
	return srv
}

func dotTask(id, input, options string) string {
	s := strings.Replace(runTask(id, "", "", false), `"test_id":"TCP-01"`, `"test_id":"DNS-04"`, 1)
	s = strings.Replace(s, `"input":{}`, `"input":`+input, 1)
	return strings.Replace(s, `"options":{}`, `"options":`+options, 1)
}

func dotFrame(r []byte) []byte { return append([]byte{byte(len(r) >> 8), byte(len(r))}, r...) }

func dotReply(code uint16, records ...loopback.DNSRecord) func([]byte) []byte {
	return func(q []byte) []byte { return dotFrame(loopback.DNSResponse(q, code, records)) }
}

var dotAnswer = dotReply(0, loopback.DNSRecord{Type: "A", Value: "9.9.9.9"})

// dotSized pads a NOERROR response with one additional record of an
// uninterpreted type to exactly size bytes.
func dotSized(size int) func([]byte) []byte {
	return func(q []byte) []byte {
		r := loopback.DNSResponse(q, 0, []loopback.DNSRecord{{Type: "A", Value: "9.9.9.9"}})
		rdata := size - len(r) - 11
		r = append(r, 0, 0, 16, 0, 1, 0, 0, 0, 60, byte(rdata>>8), byte(rdata))
		r = append(r, make([]byte, rdata)...)
		binary.BigEndian.PutUint16(r[10:], binary.BigEndian.Uint16(r[10:])+1)
		return dotFrame(r)
	}
}

type dotExpect struct {
	state, reason, outcome string
	types                  []string // observation types, in order
	failure                string   // terminal observation failure
	validation, rcode      string
	answers                int
}

// dotResult checks the serialized result shape and that no resolver name,
// address, port, bootstrap address, query name, certificate, framing or raw
// error material is retained (REQ-DNS-04-0029).
func dotResult(t *testing.T, raw []byte, want dotExpect) {
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
	for _, canary := range []string{dotName, "query.example", "8.8.8.8", "853", `"resolver_name"`, `"bootstrap_ip"`, `"hostname"`, `"port"`, loopback.CertificateCanary, "424242", "91337", "CANARY-RAW-ERROR", "x509", "127.0.0.1", `"alpn"`, `"ttl"`} {
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
		"tls_handshake": {"address_family": true, "tls_version": true, "certificate_validation": true, "alert": true},
		"dns_query":     {"engine": true, "address_family": true, "query_type": true, "responses": true},
	}
	for i, o := range m.Observations {
		last := i == len(m.Observations)-1
		if o.Type != want.types[i] || o.Duration%250 != 0 || !last && o.Status != "succeeded" || last && o.Failure != want.failure || last && (want.failure == "") != (o.Status == "succeeded") {
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
			if str("address_family") != "ipv4" || last && str("certificate_validation") != want.validation || !last && str("certificate_validation") != "trusted" {
				t.Fatalf("tls_handshake: %s", raw)
			}
		case "dns_query":
			engine, responses := "dot", 1
			if i == 0 && len(want.types) > 1 && want.types[1] == "tcp_connect" || len(want.types) == 1 {
				engine, responses = "system", 0
			}
			list, _ := o.Details["responses"].([]any)
			if str("engine") != engine || str("query_type") != "A" || engine == "system" && str("address_family") != "" {
				t.Fatalf("dns_query: %s", raw)
			}
			if engine == "dot" && want.rcode == "" {
				responses = 0
			}
			if len(list) != responses {
				t.Fatalf("responses: %s", raw)
			}
			if responses == 1 {
				resp := list[0].(map[string]any)
				answers, _ := resp["answers"].([]any)
				if resp["rcode"] != want.rcode || len(answers) != want.answers || len(resp) != 4 {
					t.Fatalf("response: %s", raw)
				}
			}
		}
	}
}

var (
	dotBootstrapped = []string{"tcp_connect", "tls_handshake", "dns_query"}
	dotResolved     = []string{"dns_query", "tcp_connect", "tls_handshake", "dns_query"}
	dotResponse     = dotExpect{state: "completed", outcome: "response", types: dotBootstrapped, rcode: "NOERROR", answers: 1}
)

// M-DNS-04-ADMISSION: actual Client admission against the registered resolver
// profile and query name, with atomic rejection before any lookup or socket.
// requirement-ref: REQ-DNS-04-0004
// requirement-ref: REQ-DNS-04-0009
// requirement-ref: REQ-DNS-04-0010
// requirement-ref: REQ-DNS-04-0013
// requirement-ref: REQ-DNS-04-0014
// requirement-ref: REQ-DNS-04-0015
// requirement-ref: REQ-DNS-04-0038
func TestMDns04Admission(t *testing.T) {
	f := newDotFixture(t)
	long := strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", 53) + ".example"
	for _, input := range []string{
		dotInput, dotSystem,
		strings.Replace(dotInput, "8.8.8.8", "2606:4700:4700::1111", 1),
		strings.Replace(dotInput, "853", "8853", 1),
		strings.Replace(dotInput, `"record_type":"A"`, `"record_type":"AAAA"`, 1),
		strings.Replace(strings.Replace(dotInput, dotName, "DoT.Resolver.EXAMPLE", 1), "query.example", "bücher.example", 1),
		strings.Replace(strings.Replace(dotInput, dotName, long, 1), "query.example", long, 1),
	} {
		t.Run("valid/"+input, func(t *testing.T) {
			h, d := dotHarness(t, f, input)
			c := h.initialized(t)
			defer c.Close()
			data := dotTask("valid", input, `{}`)
			if _, err := c.AdmitTask(data, h.embedded(t, data)); err != nil {
				t.Fatal(codeOf(err))
			}
			if d.calls.Load() != 0 || d.lookups.Load() != 0 {
				t.Fatal("admission contacted transport or resolver")
			}
		})
	}
	field := func(name, value string) string {
		return strings.Replace(dotInput, fmt.Sprintf(`"%s":%s`, name, map[string]string{"resolver_name": `"` + dotName + `"`, "port": "853", "hostname": `"query.example"`, "record_type": `"A"`, "bootstrap_ip": `"8.8.8.8"`}[name]), fmt.Sprintf(`"%s":%s`, name, value), 1)
	}
	inputs := []string{`{}`, `null`, `[]`, dotSystem[:len(dotSystem)-1] + `,"x":1}`, dotInput[:len(dotInput)-1] + `,"transport":"tcp"}`, dotInput[:len(dotInput)-1] + `,"cleartext_fallback":true}`,
		`{"port":853,"hostname":"query.example","record_type":"A"}`, `{"resolver_name":"dot.resolver.example","hostname":"query.example","record_type":"A"}`,
		`{"resolver_name":"dot.resolver.example","port":853,"record_type":"A"}`, `{"resolver_name":"dot.resolver.example","port":853,"hostname":"query.example"}`}
	for _, v := range []string{`0`, `65536`, `"853"`, `853.5`, `null`} {
		inputs = append(inputs, field("port", v))
	}
	for _, v := range []string{`"TXT"`, `"a"`, `"ANY"`, `""`, `null`} {
		inputs = append(inputs, field("record_type", v))
	}
	for _, n := range []string{"", "8.8.8.8", "[::1]", "*.example", "example", "a.example.", "dot.local", "localhost", "dot.internal", "${resolver}", "{device}.example", " dot.example", long + "x"} {
		inputs = append(inputs, field("resolver_name", fmt.Sprintf("%q", n)), field("hostname", fmt.Sprintf("%q", n)))
	}
	for _, ip := range []string{"dot.resolver.example", "${bootstrap}", "8.8.8.08", "0.0.0.0", "127.0.0.1", "10.0.0.1", "100.64.0.1", "169.254.1.1", "192.0.2.1", "::1", "fe80::1", "fc00::1", "2001:db8::1", "::ffff:8.8.8.8", "2606:4700:4700::ABCD", "fe80::1%en0", ""} {
		inputs = append(inputs, field("bootstrap_ip", fmt.Sprintf("%q", ip)))
	}
	for i, input := range inputs {
		t.Run(fmt.Sprint("invalid/", i), func(t *testing.T) {
			h, d := dotHarness(t, f, dotInput)
			c := h.initialized(t)
			defer c.Close()
			data := dotTask("invalid", input, `{}`)
			handle, err := c.AdmitTask(data, h.embedded(t, data))
			if handle != "" || codeOf(err) != "task_schema_invalid" {
				t.Fatalf("%s: %s", input, codeOf(err))
			}
			if taskHandles(c) != 0 || h.replay.Commits() != 0 || d.calls.Load() != 0 || d.lookups.Load() != 0 || len(h.run.output.Records()) != 0 {
				t.Fatal("non-atomic rejection")
			}
		})
	}
	for _, tc := range []struct{ name, registered, input string }{
		{"other port", dotInput, strings.Replace(dotInput, "853", "8853", 1)},
		{"nonstandard port unreviewed", dotInput, strings.Replace(dotInput, "853", "53", 1)},
		{"other bootstrap", dotInput, strings.Replace(dotInput, "8.8.8.8", "8.8.4.4", 1)},
		{"bootstrap omitted", dotInput, dotSystem},
		{"bootstrap added", dotSystem, dotInput},
		{"other resolver", dotInput, strings.Replace(dotInput, dotName, "other.example", 1)},
		{"other query name", dotInput, strings.Replace(dotInput, "query.example", "other.example", 1)},
	} {
		t.Run("unregistered/"+tc.name, func(t *testing.T) {
			h, d := dotHarness(t, f, tc.registered)
			c := h.initialized(t)
			defer c.Close()
			data := dotTask("unregistered", tc.input, `{}`)
			_, err := c.AdmitTask(data, h.embedded(t, data))
			if codeOf(err) != "target_not_registered" || d.calls.Load() != 0 || d.lookups.Load() != 0 || h.replay.Commits() != 0 {
				t.Fatal(codeOf(err))
			}
		})
	}
}

// M-DNS-04-OPTIONS: the address family option and its bootstrap
// compatibility; every fallback, reuse or authentication control is outside
// the closed options shape.
// requirement-ref: REQ-DNS-04-0016
// requirement-ref: REQ-DNS-04-0017
func TestMDns04Options(t *testing.T) {
	f := newDotFixture(t)
	v6 := strings.Replace(dotInput, "8.8.8.8", "2606:4700:4700::1111", 1)
	for _, tc := range []struct{ input, options string }{
		{dotInput, `{}`}, {dotInput, `{"address_family":"auto"}`}, {dotInput, `{"address_family":"ipv4"}`}, {v6, `{"address_family":"ipv6"}`}, {v6, `{"address_family":"auto"}`},
		{dotSystem, `{"address_family":"ipv4"}`}, {dotSystem, `{"address_family":"ipv6"}`},
	} {
		t.Run("valid/"+tc.options, func(t *testing.T) {
			h, d := dotHarness(t, f, tc.input)
			c := h.initialized(t)
			defer c.Close()
			data := dotTask("options", tc.input, tc.options)
			if handle, err := c.AdmitTask(data, h.embedded(t, data)); err != nil || handle == "" {
				t.Fatal(codeOf(err))
			}
			if d.calls.Load() != 0 {
				t.Fatal("options contacted transport")
			}
		})
	}
	for _, tc := range []struct{ input, options string }{
		{dotInput, `{"address_family":"ipv6"}`}, {v6, `{"address_family":"ipv4"}`},
		{dotInput, `null`}, {dotInput, `[]`}, {dotInput, `{"address_family":null}`}, {dotInput, `{"address_family":""}`}, {dotInput, `{"address_family":"IPv4"}`}, {dotInput, `{"address_family":4}`}, {dotInput, `{"address_family":"dual"}`},
		{dotInput, `{"opportunistic":true}`}, {dotInput, `{"authentication":"opportunistic"}`}, {dotInput, `{"cleartext_fallback":true}`}, {dotInput, `{"fallback_resolver":"9.9.9.9"}`},
		{dotInput, `{"reuse_connection":true}`}, {dotInput, `{"alpn_protocols":["dot"]}`}, {dotInput, `{"insecure_skip_verify":true}`}, {dotInput, `{"retries":1}`}, {dotInput, `{"x":0}`},
	} {
		t.Run("invalid/"+tc.options, func(t *testing.T) {
			h, d := dotHarness(t, f, tc.input)
			c := h.initialized(t)
			defer c.Close()
			data := dotTask("options", tc.input, tc.options)
			handle, err := c.AdmitTask(data, h.embedded(t, data))
			if codeOf(err) != "task_schema_invalid" || handle != "" || h.replay.Commits() != 0 || d.calls.Load() != 0 {
				t.Fatal(codeOf(err))
			}
		})
	}
}

// M-DNS-04-EXECUTION and OUTPUT: production registry, real sockets and a real
// handshake. Bootstrap is omitted with an approved bootstrap address; one
// connection, one ClientHello naming the resolver with no ALPN offer, one
// framed query and one response. The proxy path needs the bootstrap address,
// and CONNECT uses it while TLS still authenticates the resolver name.
// requirement-ref: REQ-DNS-04-0005
// requirement-ref: REQ-DNS-04-0006
// requirement-ref: REQ-DNS-04-0007
// requirement-ref: REQ-DNS-04-0011
// requirement-ref: REQ-DNS-04-0012
// requirement-ref: REQ-DNS-04-0024
// requirement-ref: REQ-DNS-04-0025
// requirement-ref: REQ-DNS-04-0026
// requirement-ref: REQ-DNS-04-0029
// requirement-ref: REQ-DNS-04-0036
// requirement-ref: REQ-DNS-04-0037
// requirement-ref: REQ-PROXY-0044
func TestMDns04ExecutionAndOutput(t *testing.T) {
	f := newDotFixture(t)
	for _, tc := range []struct {
		name, input string
		proxy       bool
		want        dotExpect
	}{
		{"bootstrap address", dotInput, false, dotResponse},
		{"system bootstrap", dotSystem, false, dotExpect{state: "completed", outcome: "response", types: dotResolved, rcode: "NOERROR", answers: 1}},
		{"proxy", dotInput, true, dotResponse},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, d := dotHarness(t, f, tc.input)
			cfg := f.cert
			cfg.DoT = dotAnswer
			srv := dotServe(t, d, cfg)
			config := `{}`
			data := dotTask("dot", tc.input, `{}`)
			var p *loopback.Proxy
			if tc.proxy {
				var err error
				if p, err = loopback.ListenProxy(loopback.ProxyConfig{Resolve: map[string]string{dotAuthority: srv.Address()}}); err != nil {
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
			dotResult(t, tcpRaw(t, h, c, data), tc.want)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.Wait(ctx, func(c loopback.TLSCounts) bool { return c.Accepted == 1 && c.Open == 0 }); err != nil {
				t.Fatal(err)
			}
			counts := srv.Counts()
			if counts.Accepted != 1 || counts.Plaintext != 2+dotQueryBytes || len(counts.Hellos) != 1 || counts.Hellos[0].ServerName != dotName || len(counts.Hellos[0].ALPN) != 0 {
				t.Fatalf("server %+v", counts)
			}
			wantLookups := int32(0)
			if tc.input == dotSystem {
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
				if pc.Accepted != 1 || len(pc.Requests) != 1 || !strings.HasPrefix(pc.Requests[0], "CONNECT "+dotAuthority+" HTTP/1.1") || d.calls.Load() != 0 {
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
	t.Run("proxy without bootstrap address", func(t *testing.T) {
		h, d := dotHarness(t, f, dotSystem)
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
		data := strings.Replace(dotTask("proxy-system", dotSystem, `{}`), `"limits":{}`, `"limits":{},"network_path":"proxy"`, 1)
		handle, err := c.AdmitTask(data, h.embedded(t, data))
		if codeOf(err) != "task_schema_invalid" || handle != "" || d.calls.Load() != 0 || d.lookups.Load() != 0 || p.Counts().Accepted != 0 {
			t.Fatal(codeOf(err), "incompatible path contacted")
		}
	})
}

// M-DNS-04-NORMALIZATION: every outcome row from real loopback behavior or a
// wrapped platform error; the terminal failure stays at its observed stage
// with prior observations ordered.
// requirement-ref: REQ-DNS-04-0007
// requirement-ref: REQ-DNS-04-0008
// requirement-ref: REQ-DNS-04-0027
// requirement-ref: REQ-DNS-04-0029
// requirement-ref: REQ-DNS-04-0035
func TestMDns04Normalization(t *testing.T) {
	f := newDotFixture(t)
	other, err := loopback.NewTestCA()
	if err != nil {
		t.Fatal(err)
	}
	untrusted, err := other.Issue(dotName, false)
	if err != nil {
		t.Fatal(err)
	}
	mismatched, err := f.ca.Issue("other.example", false)
	if err != nil {
		t.Fatal(err)
	}
	reply := func(fn func([]byte) []byte) loopback.TLSConfig { c := f.cert; c.DoT = fn; return c }
	tcpFail := func(err error) func(context.Context) (measurement.Conn, error) {
		return func(context.Context) (measurement.Conn, error) { return nil, err }
	}
	tls := []string{"tcp_connect", "tls_handshake"}
	for _, tc := range []struct {
		name, input string
		server      *loopback.TLSConfig
		dial        func(context.Context) (measurement.Conn, error)
		lookup      error
		want        dotExpect
	}{
		{name: "response", input: dotInput, server: ptr(reply(dotAnswer)), want: dotResponse},
		{name: "response empty", input: dotInput, server: ptr(reply(dotReply(0))), want: dotExpect{state: "completed", outcome: "response", types: dotBootstrapped, rcode: "NOERROR"}},
		{name: "bootstrap timeout", input: dotSystem, lookup: &net.DNSError{IsTimeout: true, Err: "CANARY-RAW-ERROR"}, want: dotExpect{state: "completed", outcome: "bootstrap_failure", types: []string{"dns_query"}, failure: "timeout"}},
		{name: "bootstrap not found", input: dotSystem, lookup: &net.DNSError{IsNotFound: true, Err: "CANARY-RAW-ERROR"}, want: dotExpect{state: "completed", outcome: "bootstrap_failure", types: []string{"dns_query"}, failure: "resolver_error"}},
		{name: "tcp timeout", input: dotInput, dial: tcpFail(fmt.Errorf("CANARY-RAW-ERROR: %w", syscall.ETIMEDOUT)), want: dotExpect{state: "completed", outcome: "tcp_timeout", types: tls[:1], failure: "timeout"}},
		{name: "tcp refused", input: dotInput, dial: tcpFail(fmt.Errorf("CANARY-RAW-ERROR: %w", syscall.ECONNREFUSED)), want: dotExpect{state: "completed", outcome: "tcp_refused", types: tls[:1], failure: "connection_refused"}},
		{name: "tcp reset before TLS", input: dotInput, dial: tcpFail(fmt.Errorf("CANARY-RAW-ERROR: %w", syscall.ECONNRESET)), want: dotExpect{state: "completed", outcome: "tcp_reset", types: tls[:1], failure: "connection_reset"}},
		{name: "tcp unclassified", input: dotInput, dial: tcpFail(fmt.Errorf("CANARY-RAW-ERROR: %w", syscall.ENETUNREACH)), want: dotExpect{state: "completed", outcome: "inconclusive", types: tls[:1], failure: "inconclusive"}},
		{name: "tcp reset during TLS", input: dotInput, server: &loopback.TLSConfig{Mode: loopback.TLSReset}, want: dotExpect{state: "completed", outcome: "tcp_reset", types: tls, failure: "connection_reset", validation: "not_completed"}},
		{name: "untrusted certificate", input: dotInput, server: &loopback.TLSConfig{Certificate: untrusted}, want: dotExpect{state: "completed", outcome: "tls_failure", types: tls, failure: "certificate_untrusted", validation: "untrusted"}},
		{name: "resolver name mismatch", input: dotInput, server: &loopback.TLSConfig{Certificate: mismatched}, want: dotExpect{state: "completed", outcome: "tls_failure", types: tls, failure: "certificate_name_mismatch", validation: "name_mismatch"}},
		{name: "tls alert", input: dotInput, server: &loopback.TLSConfig{Mode: loopback.TLSAlert, Alert: 40}, want: dotExpect{state: "completed", outcome: "tls_failure", types: tls, failure: "alert", validation: "not_completed"}},
		{name: "tls protocol error", input: dotInput, server: &loopback.TLSConfig{Mode: loopback.TLSGarbage}, want: dotExpect{state: "completed", outcome: "tls_failure", types: tls, failure: "protocol_error", validation: "not_completed"}},
		{name: "nxdomain", input: dotInput, server: ptr(reply(dotReply(3))), want: dotExpect{state: "completed", outcome: "dns_nxdomain", types: dotBootstrapped, failure: "nxdomain", rcode: "NXDOMAIN"}},
		{name: "servfail", input: dotInput, server: ptr(reply(dotReply(2))), want: dotExpect{state: "completed", outcome: "dns_servfail", types: dotBootstrapped, failure: "servfail", rcode: "SERVFAIL"}},
		{name: "refused", input: dotSystem, server: ptr(reply(dotReply(5))), want: dotExpect{state: "completed", outcome: "dns_refused", types: dotResolved, failure: "refused", rcode: "REFUSED"}},
		{name: "short length prefix", input: dotInput, server: ptr(reply(func([]byte) []byte { return []byte{0, 11, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11} })), want: dotExpect{state: "completed", outcome: "malformed_dns_response", types: dotBootstrapped, failure: "malformed_response"}},
		{name: "malformed message", input: dotInput, server: ptr(reply(func(q []byte) []byte { return dotFrame(append(q[:2:2], make([]byte, 18)...)) })), want: dotExpect{state: "completed", outcome: "malformed_dns_response", types: dotBootstrapped, failure: "malformed_response"}},
		{name: "query mismatch", input: dotInput, server: ptr(reply(func(q []byte) []byte {
			r := loopback.DNSResponse(q, 0, nil)
			r[0] ^= 0xff
			return dotFrame(r)
		})), want: dotExpect{state: "completed", outcome: "query_mismatch", types: dotBootstrapped, failure: "query_mismatch"}},
		{name: "unregistered rcode", input: dotInput, server: ptr(reply(dotReply(4))), want: dotExpect{state: "completed", outcome: "inconclusive", types: dotBootstrapped, failure: "inconclusive", rcode: "NOTIMP"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, d := dotHarness(t, f, tc.input)
			if tc.server != nil {
				dotServe(t, d, *tc.server)
			}
			if tc.dial != nil {
				d.dial = tc.dial
			}
			if tc.lookup != nil {
				h.adapters.run.resolver = dnsSystemResolver{func(context.Context, string, string) ([]netip.Addr, error) { d.lookups.Add(1); return nil, tc.lookup }}
			}
			c := h.initialized(t)
			defer c.Close()
			dotResult(t, tcpRaw(t, h, c, dotTask("normalize", tc.input, `{}`)), tc.want)
			if _, err := c.NextBundle(); err != nil {
				t.Fatal("result did not survive bundle validation", err)
			}
			if tc.lookup != nil && d.calls.Load() != 0 {
				t.Fatal("contact after failed bootstrap")
			}
		})
	}
}

func ptr[T any](v T) *T { return &v }

// M-DNS-04-LIMITS: per-stage operation bounds while the outer deadline stays
// open, the 65,535-byte message bound under a stricter common received
// limit, the 16-answer bound, and one connection with no retry.
// requirement-ref: REQ-DNS-04-0023
// requirement-ref: REQ-LIMIT-0034
func TestMDns04Limits(t *testing.T) {
	f := newDotFixture(t)
	for _, tc := range []struct {
		name, limits string
		stall        loopback.TLSConfig
		reached      func(loopback.TLSCounts) bool
		advance      time.Duration
		want         dotExpect
	}{
		{"tls operation", `"max_duration_ms":60000`, loopback.TLSConfig{Mode: loopback.TLSStall}, func(c loopback.TLSCounts) bool { return c.In > 0 }, 15 * time.Second,
			dotExpect{state: "completed", outcome: "timeout", types: []string{"tcp_connect", "tls_handshake"}, failure: "timeout", validation: "not_completed"}},
		{"dns operation", `"max_duration_ms":60000`, loopback.TLSConfig{Certificate: f.cert.Certificate, DoT: func([]byte) []byte { return nil }}, func(c loopback.TLSCounts) bool { return c.Plaintext == 2+dotQueryBytes }, 15 * time.Second,
			dotExpect{state: "completed", outcome: "timeout", types: dotBootstrapped, failure: "timeout"}},
		{"attempt deadline", `"max_duration_ms":1000`, loopback.TLSConfig{Certificate: f.cert.Certificate, DoT: func([]byte) []byte { return nil }}, func(c loopback.TLSCounts) bool { return c.Plaintext == 2+dotQueryBytes }, time.Second,
			dotExpect{state: "incomplete", reason: "attempt_deadline_exceeded"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, d := dotHarness(t, f, dotInput)
			srv := dotServe(t, d, tc.stall)
			c := h.initialized(t)
			defer c.Close()
			data := strings.Replace(dotTask("limit", dotInput, `{}`), `"limits":{}`, `"limits":{`+tc.limits+`}`, 1)
			rh, err := c.StartRun(h.admit(t, c, data))
			if err != nil {
				t.Fatal(codeOf(err))
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.Wait(ctx, tc.reached); err != nil {
				t.Fatal(err)
			}
			h.advance(tc.advance)
			waitTerminal(t, c, rh)
			dotResult(t, h.run.output.Records()[0], tc.want)
			if err := srv.WaitIdle(ctx); err != nil || d.calls.Load() != 1 || srv.Counts().Accepted != 1 {
				t.Fatal("stalled session not released or retried")
			}
		})
	}
	sixteen := make([]loopback.DNSRecord, 17)
	for i := range sixteen {
		sixteen[i] = loopback.DNSRecord{Type: "A", Value: fmt.Sprintf("9.9.9.%d", i+1)}
	}
	for _, tc := range []struct {
		name, limits string
		reply        func([]byte) []byte
		want         dotExpect
	}{
		{"message below bound", `"max_bytes_received":65536`, dotSized(65534), dotResponse},
		{"message at bound", `"max_bytes_received":65537`, dotSized(65535), dotResponse},
		{"message above received limit", `"max_bytes_received":65536`, dotSized(65535), dotExpect{state: "incomplete", reason: "resource_limit"}},
		{"sixteen answers", `"max_bytes_received":4096`, dotReply(0, sixteen[:16]...), dotExpect{state: "completed", outcome: "response", types: dotBootstrapped, rcode: "NOERROR", answers: 16}},
		{"seventeen answers", `"max_bytes_received":4096`, dotReply(0, sixteen...), dotExpect{state: "completed", outcome: "malformed_dns_response", types: dotBootstrapped, failure: "malformed_response"}},
		{"one connection", `"max_connections":1,"max_bytes_sent":33`, dotAnswer, dotResponse},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, d := dotHarness(t, f, dotInput)
			cfg := f.cert
			cfg.DoT = tc.reply
			srv := dotServe(t, d, cfg)
			c := h.initialized(t)
			defer c.Close()
			data := strings.Replace(dotTask("bound", dotInput, `{}`), `"limits":{}`, `"limits":{`+tc.limits+`}`, 1)
			dotResult(t, tcpRaw(t, h, c, data), tc.want)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.WaitIdle(ctx); err != nil || d.calls.Load() != 1 || srv.Counts().Accepted != 1 {
				t.Fatal("not exactly one connection")
			}
		})
	}
}

// An equivalent resolver and query tuple waits while one run is active;
// cancelling the waiting run starts no DNS-04 procedure.
// requirement-ref: REQ-DNS-04-0023
func TestMDns04TupleConcurrency(t *testing.T) {
	f := newDotFixture(t)
	h, d := dotHarness(t, f, dotInput)
	srv := dotServe(t, d, loopback.TLSConfig{Mode: loopback.TLSStall})
	c := h.initialized(t)
	defer c.Close()
	first, err := c.StartRun(h.admit(t, c, dotTask("tuple-first", dotInput, `{}`)))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Wait(ctx, func(c loopback.TLSCounts) bool { return c.In > 0 }); err != nil {
		t.Fatal(err)
	}
	second, err := c.StartRun(h.admit(t, c, dotTask("tuple-second", dotInput, `{"address_family":"ipv4"}`)))
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
	dotResult(t, rows[0], dotExpect{state: "not_run", reason: "cancelled_by_host"})
	if err := c.CancelRun(first); err != nil {
		t.Fatal(err)
	}
	waitTerminal(t, c, first)
	if rows = h.run.output.Records(); len(rows) != 2 {
		t.Fatalf("cancelled runs produced %d results, want 2", len(rows))
	}
	dotResult(t, rows[1], dotExpect{state: "incomplete", reason: "cancelled_by_host"})
	if err := srv.WaitIdle(ctx); err != nil || srv.Counts().Accepted != 1 {
		t.Fatal("stalled session not released")
	}
}

// M-DNS-04-TERMINATION through Client after the length-prefixed query is
// written, and a refused proxy tunnel, applying the GO-SPEC-ISSUE-004 table;
// internal/execution covers every reason at every Procedure point.
// requirement-ref: REQ-DNS-04-0028
// requirement-ref: REQ-RESULT-0051
func TestMDns04Termination(t *testing.T) {
	f := newDotFixture(t)
	for _, reason := range []string{"network_changed", "cancelled_by_host"} {
		t.Run(reason, func(t *testing.T) {
			h, d := dotHarness(t, f, dotInput)
			srv := dotServe(t, d, loopback.TLSConfig{Certificate: f.cert.Certificate, DoT: func([]byte) []byte { return nil }})
			c := h.initialized(t)
			defer c.Close()
			rh, err := c.StartRun(h.admit(t, c, dotTask("terminate", dotInput, `{}`)))
			if err != nil {
				t.Fatal(codeOf(err))
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.Wait(ctx, func(c loopback.TLSCounts) bool { return c.Plaintext == 2+dotQueryBytes }); err != nil {
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
			dotResult(t, h.run.output.Records()[0], dotExpect{state: "incomplete", reason: reason})
			if err := srv.WaitIdle(ctx); err != nil || srv.Counts().Accepted != 1 || srv.Counts().Plaintext != 2+dotQueryBytes {
				t.Fatal("socket not closed, retried or written after the interruption")
			}
		})
	}
	t.Run("proxy refused tunnel", func(t *testing.T) {
		h, d := dotHarness(t, f, dotInput)
		p, err := loopback.ListenProxy(loopback.ProxyConfig{Mode: loopback.ProxyReply, Head: []byte("HTTP/1.1 502 Bad Gateway\r\n\r\n")})
		if err != nil {
			t.Fatal(err)
		}
		defer p.Close()
		data := strings.Replace(dotTask("proxy-termination", dotInput, `{}`), `"limits":{}`, `"limits":{},"network_path":"proxy"`, 1)
		_, raw, c := runOnce(t, h, proxyConfig(p.Port()), data)
		defer c.Close()
		dotResult(t, raw, dotExpect{state: "incomplete", reason: "proxy_connect_failed"})
		if d.calls.Load() != 0 || d.lookups.Load() != 0 {
			t.Fatal("fell back to the direct path or the system resolver")
		}
	})
}

// M-DNS-04-VERSIONS: exact task and result tuples, strict receiver rejection
// and no cross-test interpretation.
// requirement-ref: REQ-DNS-04-0024
func TestMDns04Versions(t *testing.T) {
	f := newDotFixture(t)
	for _, version := range []string{`"0.1.1"`, `"0.2.0"`, `"1.0.0"`, `""`, `"v0.1.0"`, `"0.1"`, `"0.1.0 "`, `0.1`, `null`, "missing"} {
		h, d := dotHarness(t, f, dotInput)
		c := h.initialized(t)
		replacement := `"test_version":` + version + `,`
		if version == "missing" {
			replacement = ""
		}
		data := strings.Replace(dotTask("version", dotInput, `{}`), `"test_version":"0.1.0",`, replacement, 1)
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
	h, d := dotHarness(t, f, dotInput)
	cfg := f.cert
	cfg.DoT = dotAnswer
	dotServe(t, d, cfg)
	c := h.initialized(t)
	defer c.Close()
	raw := tcpRaw(t, h, c, dotTask("versions", dotInput, `{}`))
	cd, _ := observation.NewContextDecoder(result.Contract, observation.ContextApprovals{})
	path, _ := observation.NewPath(1, 0, "direct", "")
	output := result.Output{Test: model.Test{ID: "DNS-04", Version: "0.1.0", ResultVersion: "0.1.0"}, QueryType: "A"}
	output.Limits.RetainDNSAnswers = true
	decoder, err := result.New(result.Binding{TaskID: "versions", Output: output, MaxAttempts: 1, MaxDurationMS: 60000, Implementation: implementation, ContextDecoder: cd, Path: path, AccessGeneration: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decoder.Decode(raw); err != nil {
		t.Fatal("valid result rejected", err)
	}
	// The same bytes are not a DNS-03 or DNS-01 result, and a bootstrap
	// binding requires the system observation first.
	for _, o := range []result.Output{
		{Test: model.Test{ID: "DNS-03", Version: "0.1.0", ResultVersion: "0.1.0"}, QueryType: "A"},
		{Test: model.Test{ID: "DNS-01", Version: "0.1.0", ResultVersion: "0.1.0"}, QueryType: "A", DNSMode: "tcp"},
		{Test: model.Test{ID: "DNS-04", Version: "0.1.0", ResultVersion: "0.1.0"}, QueryType: "A", Bootstrap: true},
	} {
		o.Limits.RetainDNSAnswers = true
		other, err := result.New(result.Binding{TaskID: "versions", Output: o, MaxAttempts: 1, MaxDurationMS: 60000, Implementation: implementation, ContextDecoder: cd, Path: path, AccessGeneration: 1})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := other.Decode(raw); err == nil {
			t.Fatalf("DNS-04 output accepted as %+v", o.Test)
		}
	}
	for _, field := range []string{"schema_version", "version", "result_version"} {
		for _, v := range []any{"0.1.1", "0.2.0", "1.0.0", "0.1", "v0.1.0", "", "0.1.0 ", 0.1, nil} {
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
		func(v map[string]any) { v["test"].(map[string]any)["id"] = "DNS-03" },
		func(v map[string]any) { v["test"].(map[string]any)["id"] = "DNS-01" },
		func(v map[string]any) { v["test"].(map[string]any)["extra"] = true },
		func(v map[string]any) { v["measurement"].(map[string]any)["extra"] = true },
		func(v map[string]any) { v["measurement"].(map[string]any)["outcome"] = "resolved" },
		func(v map[string]any) { v["measurement"].(map[string]any)["outcome"] = "blocked" },
		func(v map[string]any) { obs(v, 2)["details"].(map[string]any)["engine"] = "doh" },
		func(v map[string]any) { obs(v, 2)["details"].(map[string]any)["engine"] = "tcp" },
		func(v map[string]any) { obs(v, 2)["details"].(map[string]any)["query_type"] = "AAAA" },
		func(v map[string]any) { obs(v, 2)["details"].(map[string]any)["resolver_name"] = dotName },
		func(v map[string]any) { obs(v, 1)["details"].(map[string]any)["server_name"] = dotName },
		func(v map[string]any) { obs(v, 0)["details"].(map[string]any)["port"] = 853 },
		func(v map[string]any) {
			m := v["measurement"].(map[string]any)
			m["observations"] = []any{obs(v, 0), obs(v, 2)}
		},
		func(v map[string]any) {
			m := v["measurement"].(map[string]any)
			m["observations"] = append([]any{obs(v, 2)}, m["observations"].([]any)...)
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
