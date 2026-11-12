package barometer

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
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
	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

var tls01 = task.TestVersion{ID: "TLS-01", Version: "0.1.0"}

const tlsServerName = "measurement.example"
const tlsInput = `{"ip_address":"8.8.8.8","port":443,"server_name":"measurement.example"}`
const tlsOptions = `{"alpn_protocols":["h2","http/1.1"]}`

// These tests admit public literal targets with the production codec, target
// rule and registry, and trust only a synthetic CA through the measurement
// roots seam. The scripted transport maps the checked literal to a loopback
// TLS server and fails if the runner changes its network or authority.
type tlsTransport struct {
	calls     atomic.Int32
	authority string
	dial      func(context.Context) (measurement.Conn, error)
}

func (d *tlsTransport) Dialer(int) (measurement.Dialer, error) { return d, nil }
func (d *tlsTransport) DialContext(ctx context.Context, network, address string) (measurement.Conn, error) {
	d.calls.Add(1)
	expected := d.authority
	if expected == "" {
		expected = tcpAuthority
	}
	if network != "tcp" || address != expected {
		return nil, errors.New("synthetic: unexpected destination")
	}
	return d.dial(ctx)
}

type tlsFixture struct {
	ca   *loopback.TestCA
	cert tls.Certificate
}

func newTLSFixture(t *testing.T) tlsFixture {
	t.Helper()
	ca, err := loopback.NewTestCA()
	if err != nil {
		t.Fatal(err)
	}
	cert, err := ca.Issue(tlsServerName, false)
	if err != nil {
		t.Fatal(err)
	}
	return tlsFixture{ca, cert}
}

func tlsPolicy(t *testing.T, h *harness, ip string, port uint16, name string, alpn []string) {
	t.Helper()
	p := synPolicyInput(2, "reg-a", func(p *task.Permissions) { p.Tests = []task.TestVersion{tls01}; p.AllowUnknownRisk = true })
	p.Targets[0].Test = tls01
	p.Targets[0].Tuple = task.TargetTuple{Address: netip.MustParseAddr(ip), Port: port, Scheme: "tls", TLSServerName: name, ALPN: alpn}
	var err error
	h.adapters.policy, err = task.NewLocalPolicy(p)
	if err != nil {
		t.Fatal(err)
	}
}

func tlsProductionHarness(t *testing.T, f tlsFixture) (*harness, *tlsTransport) {
	t.Helper()
	h := runningHarness(t)
	h.adapters.registry = nil
	h.adapters.measurementRoots = f.ca.Pool()
	tlsPolicy(t, h, tcpPublic, 443, tlsServerName, []string{"h2", "http/1.1"})
	h.adapters.run.paths = nil
	d := &tlsTransport{dial: func(context.Context) (measurement.Conn, error) { return nil, syscall.ECONNREFUSED }}
	h.adapters.run.directDialers = d
	h.adapters.approvals = config.Approvals{ConnectionCodes: []string{"outline_mobileproxy"}}
	return h, d
}

func tlsServe(t *testing.T, d *tlsTransport, cfg loopback.TLSConfig) *loopback.TLSServer {
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

func tlsTask(id string) string {
	s := strings.Replace(runTask(id, "", "", false), `"test_id":"TCP-01"`, `"test_id":"TLS-01"`, 1)
	s = strings.Replace(s, `"input":{}`, `"input":`+tlsInput, 1)
	return strings.Replace(s, `"options":{}`, `"options":`+tlsOptions, 1)
}

type tlsObservation struct {
	Type, Status, Failure string
	Duration              int `json:"duration_ms"`
	Details               map[string]any
}

type tlsExpect struct {
	state, reason, outcome string
	tcpFailure             string // empty when tcp_connect succeeded
	tls                    bool   // a tls_handshake observation is present
	tlsFailure             string
	validation             string
	alert                  string
	version, alpn          string
}

// tlsResult checks the serialized result shape and that no target,
// certificate or raw error material is retained (REQ-TLS-01-0028).
func tlsResult(t *testing.T, raw []byte, want tlsExpect) {
	t.Helper()
	var r struct {
		Execution      struct{ State, Reason string }
		NetworkContext json.RawMessage `json:"network_context"`
		Measurement    *struct {
			Outcome      string
			Observations []tlsObservation
		}
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatal(err)
	}
	if r.Execution.State != want.state || r.Execution.Reason != want.reason {
		t.Fatalf("state: %s", raw)
	}
	for _, canary := range []string{tcpPublic, tlsServerName, "other.example", `"port"`, `"server_name"`, `"alpn_protocols"`, loopback.CertificateCanary, "424242", "91337", "CANARY-RAW-ERROR", "remote error", "x509", "127.0.0.1"} {
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
	n := 1
	if want.tls {
		n = 2
	}
	if m == nil || m.Outcome != want.outcome || len(m.Observations) != n {
		t.Fatalf("measurement: %s", raw)
	}
	tcp := m.Observations[0]
	if tcp.Type != "tcp_connect" || tcp.Duration%250 != 0 || len(tcp.Details) != 1 || tcp.Details["address_family"] != "ipv4" || tcp.Failure != want.tcpFailure ||
		(want.tcpFailure == "") != (tcp.Status == "succeeded") {
		t.Fatalf("tcp_connect: %s", raw)
	}
	if !want.tls {
		return
	}
	o := m.Observations[1]
	allowed := map[string]bool{"address_family": true, "tls_version": true, "alpn": true, "certificate_validation": true, "alert": true}
	for k := range o.Details {
		if !allowed[k] {
			t.Fatalf("detail %s: %s", k, raw)
		}
	}
	str := func(k string) string { s, _ := o.Details[k].(string); return s }
	if o.Type != "tls_handshake" || o.Duration%250 != 0 || o.Failure != want.tlsFailure || (want.tlsFailure == "") != (o.Status == "succeeded") ||
		str("address_family") != "ipv4" || str("certificate_validation") != want.validation || str("alert") != want.alert ||
		str("tls_version") != want.version || str("alpn") != want.alpn {
		t.Fatalf("tls_handshake: %s", raw)
	}
}

var tlsSucceeded = tlsExpect{state: "completed", outcome: "handshake_succeeded", tls: true, validation: "trusted", version: "TLS1.3", alpn: "h2"}

// M-TLS-01-ADMISSION: actual Client admission and atomic rejection before any
// socket activity.
// requirement-ref: REQ-TLS-01-0004
// requirement-ref: REQ-TLS-01-0009
// requirement-ref: REQ-TLS-01-0010
// requirement-ref: REQ-TLS-01-0011
// requirement-ref: REQ-TLS-01-0012
// requirement-ref: REQ-TLS-01-0013
// requirement-ref: REQ-TLS-01-0014
// requirement-ref: REQ-TLS-01-0015
func TestMTls01Admission(t *testing.T) {
	f := newTLSFixture(t)
	long := strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", 53) + ".example"
	for _, tc := range []struct {
		ip           string
		port         int
		input        string
		registeredAs string
	}{
		{"8.8.8.8", 1, "measurement.example", "measurement.example"},
		{"8.8.8.8", 65535, "measurement.example", "measurement.example"},
		{"2606:4700:4700::1111", 443, "measurement.example", "measurement.example"},
		{"8.8.8.8", 443, "MEASUREMENT.Example", "measurement.example"},
		{"8.8.8.8", 443, "bücher.example", "xn--bcher-kva.example"},
		{"8.8.8.8", 443, long, long},
	} {
		t.Run(fmt.Sprintf("valid/%s/%d/%s", tc.ip, tc.port, tc.input), func(t *testing.T) {
			h, d := tlsProductionHarness(t, f)
			tlsPolicy(t, h, tc.ip, uint16(tc.port), tc.registeredAs, []string{"h2", "http/1.1"})
			c := h.initialized(t)
			defer c.Close()
			data := strings.Replace(tlsTask("valid"), tlsInput, fmt.Sprintf(`{"ip_address":%q,"port":%d,"server_name":%q}`, tc.ip, tc.port, tc.input), 1)
			if _, err := c.AdmitTask(data, h.embedded(t, data)); err != nil {
				t.Fatal(codeOf(err))
			}
			if d.calls.Load() != 0 {
				t.Fatal("admission contacted transport")
			}
		})
	}
	inputs := []string{
		`{}`, `null`, `[]`, `{"ip_address":"8.8.8.8","port":443}`, `{"ip_address":"8.8.8.8","server_name":"measurement.example"}`, `{"port":443,"server_name":"measurement.example"}`,
		`{"ip_address":"8.8.8.8","port":0,"server_name":"measurement.example"}`, `{"ip_address":"8.8.8.8","port":65536,"server_name":"measurement.example"}`, `{"ip_address":"8.8.8.8","port":"443","server_name":"measurement.example"}`,
		`{"ip_address":"8.8.8.8","port":443,"server_name":"measurement.example","server_name":"measurement.example"}`,
		`{"ip_address":"8.8.8.8","port":443,"server_name":"measurement.example","x":1}`,
		`{"ip_address":"8.8.8.8","port":443,"server_name":["measurement.example"]}`, `{"ip_address":"8.8.8.8","port":443,"server_name":null}`,
		`{"ip_address":"8.8.8.8","port":443,"server_name":"measurement.example","insecure_skip_verify":true}`,
		`{"ip_address":"8.8.8.8","port":443,"server_name":"measurement.example","verify_hostname":false}`,
	}
	for _, name := range []string{"", "8.8.8.8", "2606:4700:4700::1111", "[::1]", "*.example.com", "*", "example", "measurement.example.", ".example", "a..example", "-a.example", "a_b.example",
		"localhost", "measurement.local", "measurement.test", "${server_name}", " measurement.example", long + "x", strings.Repeat("a", 64) + ".example"} {
		inputs = append(inputs, fmt.Sprintf(`{"ip_address":"8.8.8.8","port":443,"server_name":%q}`, name))
	}
	for _, ip := range []string{"example.test", "${target}", "8.8.8.08", "0.0.0.0", "127.0.0.1", "10.0.0.1", "100.64.0.1", "169.254.1.1", "192.0.2.1", "198.18.0.1", "224.0.0.1", "::1", "fe80::1", "fc00::1", "2001:db8::1", "::ffff:8.8.8.8", "2606:4700:4700::ABCD", "fe80::1%en0"} {
		inputs = append(inputs, fmt.Sprintf(`{"ip_address":%q,"port":443,"server_name":"measurement.example"}`, ip))
	}
	for i, input := range inputs {
		t.Run(fmt.Sprint("invalid/", i), func(t *testing.T) {
			h, d := tlsProductionHarness(t, f)
			c := h.initialized(t)
			defer c.Close()
			data := strings.Replace(tlsTask("invalid"), tlsInput, input, 1)
			handle, err := c.AdmitTask(data, h.embedded(t, data))
			if handle != "" || codeOf(err) != "task_schema_invalid" {
				t.Fatalf("%s: %s", input, codeOf(err))
			}
			if taskHandles(c) != 0 || h.replay.Commits() != 0 || len(c.rt.runs.entries) != 0 || d.calls.Load() != 0 || len(h.run.output.Records()) != 0 {
				t.Fatal("non-atomic rejection")
			}
		})
	}
	for _, tc := range []struct{ name, input, options string }{
		{"other port", strings.Replace(tlsInput, "443", "444", 1), tlsOptions},
		{"other address", strings.Replace(tlsInput, "8.8.8.8", "8.8.4.4", 1), tlsOptions},
		{"other server name", strings.Replace(tlsInput, tlsServerName, "other.example", 1), tlsOptions},
		{"other ALPN order", tlsInput, `{"alpn_protocols":["http/1.1","h2"]}`},
		{"no ALPN", tlsInput, `{}`},
	} {
		t.Run("unregistered/"+tc.name, func(t *testing.T) {
			h, d := tlsProductionHarness(t, f)
			c := h.initialized(t)
			defer c.Close()
			data := strings.Replace(strings.Replace(tlsTask("unregistered"), tlsInput, tc.input, 1), tlsOptions, tc.options, 1)
			_, err := c.AdmitTask(data, h.embedded(t, data))
			if codeOf(err) != "target_not_registered" || d.calls.Load() != 0 || h.replay.Commits() != 0 {
				t.Fatal(codeOf(err))
			}
		})
	}
}

// M-TLS-01-OPTIONS: ALPN offers at every boundary; every control outside the
// registered option is rejected.
// requirement-ref: REQ-TLS-01-0016
// requirement-ref: REQ-TLS-01-0017
// requirement-ref: REQ-OBS-0049
func TestMTls01Options(t *testing.T) {
	f := newTLSFixture(t)
	for _, tc := range []struct {
		options string
		alpn    []string
	}{
		{`{}`, nil}, {`{"alpn_protocols":[]}`, nil}, {`{"alpn_protocols":["h2"]}`, []string{"h2"}}, {`{"alpn_protocols":["dot"]}`, []string{"dot"}}, {`{"alpn_protocols":["http/1.1"]}`, []string{"http/1.1"}},
		{`{"alpn_protocols":["http/1.1","h2","dot"]}`, []string{"http/1.1", "h2", "dot"}},
	} {
		t.Run("valid/"+tc.options, func(t *testing.T) {
			h, d := tlsProductionHarness(t, f)
			tlsPolicy(t, h, tcpPublic, 443, tlsServerName, tc.alpn)
			c := h.initialized(t)
			defer c.Close()
			data := strings.Replace(tlsTask("options"), tlsOptions, tc.options, 1)
			if handle, err := c.AdmitTask(data, h.embedded(t, data)); err != nil || handle == "" {
				t.Fatal(codeOf(err))
			}
			if d.calls.Load() != 0 {
				t.Fatal("options contacted transport")
			}
		})
	}
	for _, options := range []string{
		`null`, `[]`, `false`, `{"alpn_protocols":null}`, `{"alpn_protocols":"h2"}`, `{"alpn_protocols":[1]}`, `{"alpn_protocols":[""]}`, `{"alpn_protocols":["unknown"]}`, `{"alpn_protocols":["h2c"]}`,
		`{"alpn_protocols":["H2"]}`, `{"alpn_protocols":["h3"]}`, `{"alpn_protocols":["spdy/3"]}`, `{"alpn_protocols":["` + strings.Repeat("a", 33) + `"]}`, `{"alpn_protocols":["h2","h2"]}`,
		`{"alpn_protocols":["h2","http/1.1","dot","h2","http/1.1","dot","h2","http/1.1","dot"]}`, `{"alpn_protocols":["h2"],"alpn_protocols":["h2"]}`,
		`{"cipher_suites":["TLS_AES_128_GCM_SHA256"]}`, `{"tls_version":"TLS1.2"}`, `{"min_version":"TLS1.3"}`, `{"insecure_skip_verify":true}`, `{"allow_invalid_certificates":true}`,
		`{"server_name":"other.example"}`, `{"sni":"other.example"}`, `{"client_certificate":"x"}`, `{"session_resumption":true}`, `{"session_ticket":"x"}`, `{"x":0}`,
	} {
		t.Run("invalid/"+options, func(t *testing.T) {
			h, d := tlsProductionHarness(t, f)
			c := h.initialized(t)
			defer c.Close()
			data := strings.Replace(tlsTask("options"), `"options":`+tlsOptions, `"options":`+options, 1)
			handle, err := c.AdmitTask(data, h.embedded(t, data))
			if codeOf(err) != "task_schema_invalid" || handle != "" || h.replay.Commits() != 0 || d.calls.Load() != 0 {
				t.Fatal(codeOf(err))
			}
		})
	}
}

// M-TLS-01-EXECUTION and OUTPUT: production registry, both paths, real
// sockets and a real handshake. One connection to the explicit address, one
// ClientHello with the configured SNI and ALPN order, no application data.
// requirement-ref: REQ-TLS-01-0005
// requirement-ref: REQ-TLS-01-0006
// requirement-ref: REQ-TLS-01-0007
// requirement-ref: REQ-TLS-01-0023
// requirement-ref: REQ-TLS-01-0025
// requirement-ref: REQ-TLS-01-0035
// requirement-ref: REQ-TLS-01-0036
func TestMTls01ExecutionAndOutput(t *testing.T) {
	f := newTLSFixture(t)
	for _, proxy := range []bool{false, true} {
		t.Run(fmt.Sprint("proxy=", proxy), func(t *testing.T) {
			h, d := tlsProductionHarness(t, f)
			srv := tlsServe(t, d, loopback.TLSConfig{Certificate: f.cert, ALPN: []string{"h2"}})
			cfg := `{}`
			data := tlsTask("handshake")
			var p *loopback.Proxy
			if proxy {
				var err error
				p, err = loopback.ListenProxy(loopback.ProxyConfig{Resolve: map[string]string{tcpAuthority: srv.Address()}})
				if err != nil {
					t.Fatal(err)
				}
				defer p.Close()
				cfg = proxyConfig(p.Port())
				data = strings.Replace(data, `"limits":{}`, `"limits":{},"network_path":"proxy"`, 1)
			}
			c := h.client(t)
			if err := c.Initialize(cfg); err != nil {
				t.Fatal(codeOf(err))
			}
			defer c.Close()
			tlsResult(t, tcpRaw(t, h, c, data), tlsSucceeded)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.Wait(ctx, func(c loopback.TLSCounts) bool { return c.Accepted == 1 && c.Open == 0 }); err != nil {
				t.Fatal(err)
			}
			counts := srv.Counts()
			if counts.Accepted != 1 || counts.Plaintext != 0 || len(counts.Hellos) != 1 || counts.Hellos[0].ServerName != tlsServerName ||
				strings.Join(counts.Hellos[0].ALPN, ",") != "h2,http/1.1" {
				t.Fatalf("server %+v", counts)
			}
			if proxy {
				if err := p.WaitIdle(ctx); err != nil {
					t.Fatal(err)
				}
				pc := p.Counts()
				if pc.Accepted != 1 || len(pc.Requests) != 1 || !strings.HasPrefix(pc.Requests[0], "CONNECT "+tcpAuthority+" HTTP/1.1") || d.calls.Load() != 0 {
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
}

// M-TLS-01-NORMALIZATION: every outcome row from real loopback behavior or a
// wrapped platform error; stage-specific failures are preserved.
// requirement-ref: REQ-TLS-01-0007
// requirement-ref: REQ-TLS-01-0008
// requirement-ref: REQ-TLS-01-0026
// requirement-ref: REQ-TLS-01-0028
// requirement-ref: REQ-OBS-0019
// requirement-ref: REQ-OBS-0043
// requirement-ref: REQ-OBS-0048
// requirement-ref: REQ-OBS-0049
// requirement-ref: REQ-OBS-0052
// requirement-ref: REQ-OBS-0050
func TestMTls01Normalization(t *testing.T) {
	f := newTLSFixture(t)
	other, err := loopback.NewTestCA()
	if err != nil {
		t.Fatal(err)
	}
	untrusted, err := other.Issue(tlsServerName, false)
	if err != nil {
		t.Fatal(err)
	}
	mismatch, err := f.ca.Issue("other.example", false)
	if err != nil {
		t.Fatal(err)
	}
	expired, err := f.ca.Issue(tlsServerName, true)
	if err != nil {
		t.Fatal(err)
	}
	failedTLS := func(outcome, failure, validation, alert string) tlsExpect {
		return tlsExpect{state: "completed", outcome: outcome, tls: true, tlsFailure: failure, validation: validation, alert: alert}
	}
	for _, tc := range []struct {
		name    string
		server  *loopback.TLSConfig
		dialErr error
		stall   bool
		options string
		want    tlsExpect
	}{
		{name: "succeeded", server: &loopback.TLSConfig{Certificate: f.cert, ALPN: []string{"h2"}}, want: tlsSucceeded},
		{name: "succeeded without ALPN", server: &loopback.TLSConfig{Certificate: f.cert}, want: tlsExpect{state: "completed", outcome: "handshake_succeeded", tls: true, validation: "trusted", version: "TLS1.3"}},
		{name: "tcp refused", dialErr: syscall.ECONNREFUSED, want: tlsExpect{state: "completed", outcome: "tcp_refused", tcpFailure: "connection_refused"}},
		{name: "tcp reset", dialErr: syscall.ECONNRESET, want: tlsExpect{state: "completed", outcome: "tcp_reset", tcpFailure: "connection_reset"}},
		{name: "tcp timeout", dialErr: syscall.ETIMEDOUT, want: tlsExpect{state: "completed", outcome: "tcp_timeout", tcpFailure: "timeout"}},
		{name: "tcp inconclusive", dialErr: errors.New("CANARY-RAW-ERROR"), want: tlsExpect{state: "completed", outcome: "inconclusive", tcpFailure: "inconclusive"}},
		{name: "tcp unreachable without a TLS-01 class", dialErr: syscall.ENETUNREACH, want: tlsExpect{state: "completed", outcome: "inconclusive", tcpFailure: "inconclusive"}},
		{name: "tls reset", server: &loopback.TLSConfig{Mode: loopback.TLSReset}, want: failedTLS("tcp_reset", "connection_reset", "not_completed", "")},
		{name: "tls timeout", server: &loopback.TLSConfig{Mode: loopback.TLSStall}, stall: true, want: failedTLS("tls_timeout", "timeout", "not_completed", "")},
		{name: "alert handshake_failure", server: &loopback.TLSConfig{Mode: loopback.TLSAlert, Alert: 40}, want: failedTLS("tls_alert", "alert", "not_completed", "handshake_failure")},
		{name: "alert certificate_error", server: &loopback.TLSConfig{Mode: loopback.TLSAlert, Alert: 42}, want: failedTLS("tls_alert", "alert", "not_completed", "certificate_error")},
		{name: "alert protocol_error", server: &loopback.TLSConfig{Mode: loopback.TLSAlert, Alert: 70}, want: failedTLS("tls_alert", "alert", "not_completed", "protocol_error")},
		{name: "alert access_denied", server: &loopback.TLSConfig{Mode: loopback.TLSAlert, Alert: 49}, want: failedTLS("tls_alert", "alert", "not_completed", "access_denied")},
		{name: "alert internal_error", server: &loopback.TLSConfig{Mode: loopback.TLSAlert, Alert: 80}, want: failedTLS("tls_alert", "alert", "not_completed", "internal_error")},
		{name: "alert other", server: &loopback.TLSConfig{Mode: loopback.TLSAlert, Alert: 200}, want: failedTLS("tls_alert", "alert", "not_completed", "other")},
		{name: "alert 120 without offer", server: &loopback.TLSConfig{Mode: loopback.TLSAlert, Alert: 120}, options: `{}`, want: failedTLS("tls_alert", "alert", "not_completed", "handshake_failure")},
		{name: "alpn no overlap", server: &loopback.TLSConfig{Certificate: f.cert, ALPN: []string{"dot"}}, want: failedTLS("alpn_no_overlap", "alpn_no_overlap", "not_completed", "")},
		{name: "protocol error", server: &loopback.TLSConfig{Mode: loopback.TLSGarbage}, want: failedTLS("protocol_error", "protocol_error", "not_completed", "")},
		{name: "untrusted", server: &loopback.TLSConfig{Certificate: untrusted, ALPN: []string{"h2"}}, want: failedTLS("certificate_untrusted", "certificate_untrusted", "untrusted", "")},
		{name: "name mismatch", server: &loopback.TLSConfig{Certificate: mismatch, ALPN: []string{"h2"}}, want: failedTLS("certificate_name_mismatch", "certificate_name_mismatch", "name_mismatch", "")},
		{name: "time invalid", server: &loopback.TLSConfig{Certificate: expired, ALPN: []string{"h2"}}, want: failedTLS("certificate_time_invalid", "certificate_time_invalid", "time_invalid", "")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, d := tlsProductionHarness(t, f)
			if tc.options == `{}` {
				tlsPolicy(t, h, tcpPublic, 443, tlsServerName, nil)
			}
			var srv *loopback.TLSServer
			if tc.server != nil {
				srv = tlsServe(t, d, *tc.server)
			} else {
				d.dial = func(context.Context) (measurement.Conn, error) {
					_ = h.run.clock.Advance(375 * time.Millisecond)
					return nil, fmt.Errorf("CANARY-RAW-ERROR: %w", tc.dialErr)
				}
			}
			if tc.stall {
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					if srv.Wait(ctx, func(c loopback.TLSCounts) bool { return c.In > 0 }) == nil {
						h.advance(tlsOperationBound)
					}
				}()
			}
			c := h.initialized(t)
			defer c.Close()
			data := strings.Replace(tlsTask("normalization"), `"limits":{}`, `"limits":{"max_duration_ms":60000}`, 1)
			if tc.options != "" {
				data = strings.Replace(data, tlsOptions, tc.options, 1)
			}
			tlsResult(t, tcpRaw(t, h, c, data), tc.want)
			if d.calls.Load() != 1 {
				t.Fatal("retry or second connection")
			}
			if srv != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := srv.WaitIdle(ctx); err != nil || srv.Counts().Accepted != 1 || srv.Counts().Plaintext != 0 {
					t.Fatal("socket leaked, extra connection or application data")
				}
			}
		})
	}
}

const tlsOperationBound = 15 * time.Second

// M-TLS-01-LIMITS: the handshake bound and the outer attempt deadline are
// distinct, the stalled handshake is released, and the attempt uses one
// connection at the schema floor. One-byte byte limits are enough because
// handshake and record bytes are not attempt bytes (REQ-LIMIT-0034).
// requirement-ref: REQ-TLS-01-0023
// requirement-ref: REQ-LIMIT-0006
// requirement-ref: REQ-LIMIT-0008
// requirement-ref: REQ-LIMIT-0034
func TestMTls01Limits(t *testing.T) {
	f := newTLSFixture(t)
	for _, tc := range []struct {
		name, limits string
		advance      time.Duration
		want         tlsExpect
	}{
		{"operation", `"max_duration_ms":60000`, tlsOperationBound, tlsExpect{state: "completed", outcome: "tls_timeout", tls: true, tlsFailure: "timeout", validation: "not_completed"}},
		{"attempt", `"max_duration_ms":1000`, time.Second, tlsExpect{state: "incomplete", reason: "attempt_deadline_exceeded"}},
		{"cancelled", `"max_duration_ms":60000`, 0, tlsExpect{state: "incomplete", reason: "cancelled_by_host"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, d := tlsProductionHarness(t, f)
			srv := tlsServe(t, d, loopback.TLSConfig{Mode: loopback.TLSStall})
			c := h.initialized(t)
			defer c.Close()
			data := strings.Replace(tlsTask("limit"), `"limits":{}`, `"limits":{`+tc.limits+`}`, 1)
			rh, err := c.StartRun(h.admit(t, c, data))
			if err != nil {
				t.Fatal(codeOf(err))
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.Wait(ctx, func(c loopback.TLSCounts) bool { return c.In > 0 }); err != nil {
				t.Fatal(err)
			}
			if tc.advance > 0 {
				h.advance(tc.advance)
			} else if err := c.CancelRun(rh); err != nil {
				t.Fatal(codeOf(err))
			}
			waitTerminal(t, c, rh)
			tlsResult(t, h.run.output.Records()[0], tc.want)
			if err := srv.WaitIdle(ctx); err != nil || d.calls.Load() != 1 || srv.Counts().Accepted != 1 {
				t.Fatal("stalled handshake not released or retried")
			}
		})
	}
	for _, limit := range []int{1, 2} {
		t.Run(fmt.Sprint("connections/", limit), func(t *testing.T) {
			h, d := tlsProductionHarness(t, f)
			srv := tlsServe(t, d, loopback.TLSConfig{Certificate: f.cert, ALPN: []string{"h2"}})
			c := h.initialized(t)
			defer c.Close()
			data := strings.Replace(tlsTask("connections"), `"limits":{}`, fmt.Sprintf(`"limits":{"max_connections":%d,"max_bytes_sent":1,"max_bytes_received":1}`, limit), 1)
			tlsResult(t, tcpRaw(t, h, c, data), tlsSucceeded)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.WaitIdle(ctx); err != nil || d.calls.Load() != 1 || srv.Counts().Accepted != 1 {
				t.Fatal("not exactly one connection")
			}
		})
	}
}

// An equivalent endpoint and server-name tuple waits while one run is active;
// cancelling the waiting run starts no TLS procedure.
// requirement-ref: REQ-TLS-01-0023
func TestMTls01TupleConcurrency(t *testing.T) {
	f := newTLSFixture(t)
	h, d := tlsProductionHarness(t, f)
	srv := tlsServe(t, d, loopback.TLSConfig{Mode: loopback.TLSStall})
	c := h.initialized(t)
	defer c.Close()
	first, err := c.StartRun(h.admit(t, c, tlsTask("tuple-first")))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Wait(ctx, func(c loopback.TLSCounts) bool { return c.In > 0 }); err != nil {
		t.Fatal(err)
	}
	second, err := c.StartRun(h.admit(t, c, tlsTask("tuple-second")))
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
	tlsResult(t, rows[0], tlsExpect{state: "not_run", reason: "cancelled_by_host"})
	if err := c.CancelRun(first); err != nil {
		t.Fatal(err)
	}
	waitTerminal(t, c, first)
	rows = h.run.output.Records()
	if len(rows) != 2 {
		t.Fatalf("cancelled runs produced %d results, want 2", len(rows))
	}
	tlsResult(t, rows[1], tlsExpect{state: "incomplete", reason: "cancelled_by_host"})
	if err := srv.WaitIdle(ctx); err != nil || srv.Counts().Accepted != 1 {
		t.Fatal("stalled session not released")
	}
}

// M-TLS-01-TERMINATION through Client, applying the GO-SPEC-ISSUE-004 table;
// internal/execution covers every reason at every Procedure point.
// requirement-ref: REQ-TLS-01-0027
// requirement-ref: REQ-RESULT-0051
func TestMTls01Termination(t *testing.T) {
	f := newTLSFixture(t)
	for _, reason := range []string{"network_changed", "cancelled_by_host"} {
		t.Run(reason, func(t *testing.T) {
			h, d := tlsProductionHarness(t, f)
			srv := tlsServe(t, d, loopback.TLSConfig{Mode: loopback.TLSStall})
			c := h.initialized(t)
			defer c.Close()
			rh, err := c.StartRun(h.admit(t, c, tlsTask("terminate")))
			if err != nil {
				t.Fatal(codeOf(err))
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.Wait(ctx, func(c loopback.TLSCounts) bool { return c.In > 0 }); err != nil {
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
			tlsResult(t, h.run.output.Records()[0], tlsExpect{state: "incomplete", reason: reason})
			if err := srv.WaitIdle(ctx); err != nil {
				t.Fatal("socket not closed")
			}
		})
	}
	t.Run("proxy refused tunnel", func(t *testing.T) {
		h, d := tlsProductionHarness(t, f)
		p, err := loopback.ListenProxy(loopback.ProxyConfig{Mode: loopback.ProxyReply, Head: []byte("HTTP/1.1 502 Bad Gateway\r\n\r\n")})
		if err != nil {
			t.Fatal(err)
		}
		defer p.Close()
		data := strings.Replace(tlsTask("proxy-termination"), `"limits":{}`, `"limits":{},"network_path":"proxy"`, 1)
		_, raw, c := runOnce(t, h, proxyConfig(p.Port()), data)
		defer c.Close()
		tlsResult(t, raw, tlsExpect{state: "incomplete", reason: "proxy_connect_failed"})
		if d.calls.Load() != 0 {
			t.Fatal("fell back to the direct path")
		}
	})
}

// M-TLS-01-VERSIONS: exact task and result tuples, strict receiver rejection.
// requirement-ref: REQ-TLS-01-0024
func TestMTls01Versions(t *testing.T) {
	f := newTLSFixture(t)
	// "missing" removes the member. A well-formed unregistered version is
	// task_unsupported (REQ-TRUST-0047 step 4); a malformed or missing one
	// fails the envelope schema. Either is a task-validation failure with no
	// handle, replay record, contact or Measurement Result.
	for _, version := range []string{`"0.1.1"`, `"0.2.0"`, `"1.0.0"`, `""`, `"v0.1.0"`, `"0.1"`, `"0.1.0 "`, `0.1`, `1`, `null`, "missing"} {
		h, d := tlsProductionHarness(t, f)
		c := h.initialized(t)
		replacement := `"test_version":` + version + `,`
		if version == "missing" {
			replacement = ""
		}
		data := strings.Replace(tlsTask("version"), `"test_version":"0.1.0",`, replacement, 1)
		want := "task_schema_invalid"
		if version == `"0.1.1"` || version == `"0.2.0"` || version == `"1.0.0"` {
			want = "task_unsupported"
		}
		handle, err := c.AdmitTask(data, h.embedded(t, data))
		if codeOf(err) != want || handle != "" || h.replay.Commits() != 0 || d.calls.Load() != 0 || len(h.run.output.Records()) != 0 {
			t.Fatalf("version %s: %s, want %s", version, codeOf(err), want)
		}
		c.Close()
	}
	h, d := tlsProductionHarness(t, f)
	tlsServe(t, d, loopback.TLSConfig{Certificate: f.cert, ALPN: []string{"h2"}})
	c := h.initialized(t)
	defer c.Close()
	raw := tcpRaw(t, h, c, tlsTask("versions"))
	cd, _ := observation.NewContextDecoder(result.Contract, observation.ContextApprovals{})
	path, _ := observation.NewPath(1, 0, "direct", "")
	decoder, err := result.New(result.Binding{TaskID: "versions", Output: result.Output{Test: model.Test{ID: "TLS-01", Version: "0.1.0", ResultVersion: "0.1.0"}}, MaxAttempts: 1, MaxDurationMS: 60000, Implementation: implementation, ContextDecoder: cd, Path: path, AccessGeneration: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decoder.Decode(raw); err != nil {
		t.Fatal("valid result rejected", err)
	}
	for _, field := range []string{"schema_version", "version", "result_version"} {
		for _, v := range []any{"0.1.1", "0.2.0", "1.0.0", "0.1", "v0.1.0", "", 0.1, nil} {
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
		func(v map[string]any) { delete(v["test"].(map[string]any), "result_version") },
		func(v map[string]any) { v["test"].(map[string]any)["id"] = "TCP-01" },
		func(v map[string]any) { v["test"].(map[string]any)["id"] = "HTTP-01" },
		func(v map[string]any) { v["test"].(map[string]any)["extra"] = true },
		func(v map[string]any) { v["extra"] = true },
		func(v map[string]any) { v["measurement"].(map[string]any)["extra"] = true },
		func(v map[string]any) { v["measurement"].(map[string]any)["outcome"] = "connected" },
		func(v map[string]any) { v["measurement"].(map[string]any)["outcome"] = "blocked" },
		func(v map[string]any) { v["measurement"].(map[string]any)["outcome"] = "tls_alert" },
		func(v map[string]any) { obs(v, 1)["type"] = "http_exchange" },
		func(v map[string]any) { obs(v, 0)["type"] = "dns_query" },
		func(v map[string]any) {
			m := v["measurement"].(map[string]any)
			m["observations"] = m["observations"].([]any)[:1]
		},
		func(v map[string]any) {
			m := v["measurement"].(map[string]any)
			m["observations"] = []any{obs(v, 1), obs(v, 0)}
		},
		func(v map[string]any) { obs(v, 1)["details"].(map[string]any)["certificate_validation"] = "untrusted" },
		func(v map[string]any) { delete(obs(v, 1)["details"].(map[string]any), "certificate_validation") },
		func(v map[string]any) { obs(v, 1)["details"].(map[string]any)["alpn"] = "h2c" },
		func(v map[string]any) { obs(v, 1)["details"].(map[string]any)["tls_version"] = "TLS1.0" },
		func(v map[string]any) { obs(v, 1)["details"].(map[string]any)["alert"] = "handshake_failure" },
		func(v map[string]any) { obs(v, 1)["details"].(map[string]any)["server_name"] = tlsServerName },
		func(v map[string]any) {
			obs(v, 1)["details"].(map[string]any)["cipher_suite"] = "TLS_AES_128_GCM_SHA256"
		},
		func(v map[string]any) { obs(v, 1)["details"].(map[string]any)["certificate"] = "MII" },
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

const tlsCanaryName = "canary-sni-71.example"
const tlsCanaryAuthority = "8.8.4.4:47392"

// tlsCanaryFixture is a fixture whose names, serials and subjects are
// canaries, with canaries for the bytes of its chain: DER, public key,
// signature and fingerprint.
type tlsCanaryFixture struct {
	tlsFixture
	untrusted, mismatch, expired tls.Certificate
	canaries                     canarySet
}

func newTLSCanaryFixture(t *testing.T) tlsCanaryFixture {
	t.Helper()
	ca, err := loopback.NewTestCA()
	if err != nil {
		t.Fatal(err)
	}
	other, err := loopback.NewTestCA()
	if err != nil {
		t.Fatal(err)
	}
	f := tlsCanaryFixture{tlsFixture: tlsFixture{ca: ca}}
	issue := func(c *loopback.TestCA, name string, expired bool) tls.Certificate {
		cert, err := c.Issue(name, expired)
		if err != nil {
			t.Fatal(err)
		}
		return cert
	}
	f.cert, f.untrusted, f.mismatch, f.expired = issue(ca, tlsCanaryName, false), issue(other, tlsCanaryName, false), issue(ca, "canary-mismatch-72.example", false), issue(ca, tlsCanaryName, true)
	f.canaries = canarySet{
		{"destination address", "8.8.4.4"}, {"destination port", "47392"}, {"server_name", tlsCanaryName}, {"certificate SAN name", "canary-mismatch-72.example"},
		{"certificate subject and issuer", loopback.CertificateCanary}, {"leaf serial number", "424242"}, {"issuer serial number", "91337"},
		{"raw alert code", "alert(200)"}, {"raw library error", "remote error"}, {"raw x509 error", "x509:"}, {"raw TLS error", "tls: "},
		{"raw transport error", "CANARY-RAW-ERROR"}, {"ALPN offer list", `"h2","http/1.1"`},
	}
	for i, cert := range []tls.Certificate{f.cert, f.untrusted, f.mismatch} {
		leaf, err := x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			t.Fatal(err)
		}
		chain, err := x509.ParseCertificate(cert.Certificate[1])
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(leaf.Raw)
		n := fmt.Sprint(" ", i)
		f.canaries = append(f.canaries, binaryCanaries("certificate DER"+n, leaf.Raw)...)
		if i < 2 { // the mismatched leaf shares the first chain's issuer
			f.canaries = append(f.canaries, binaryCanaries("chain element DER"+n, chain.Raw)...)
		}
		f.canaries = append(f.canaries, binaryCanaries("public key"+n, leaf.RawSubjectPublicKeyInfo)...)
		f.canaries = append(f.canaries, binaryCanaries("signature"+n, leaf.Signature)...)
		f.canaries = append(f.canaries, binaryCanaries("fingerprint"+n, sum[:])...)
	}
	return f
}

func tlsCanaryTask(id string) string {
	s := strings.Replace(tlsTask(id), tlsInput, `{"ip_address":"8.8.4.4","port":47392,"server_name":"`+tlsCanaryName+`"}`, 1)
	return strings.Replace(s, `"limits":{}`, `"limits":{"max_duration_ms":60000}`, 1)
}

var tlsAllowed = []string{"outcome", "observations[].type", "observations[].status", "observations[].duration_ms", "observations[].failure", "observations[].details.address_family",
	"observations[].details.tls_version", "observations[].details.alpn", "observations[].details.certificate_validation", "observations[].details.alert"}

type tlsCanaryCase struct {
	name      string
	server    *loopback.TLSConfig
	dialErr   error
	stall     bool
	data      func(string) string
	interrupt string
	want      tlsExpect
}

// tlsCanaryRun runs one TLS-01 path against the canary fixture and returns
// its surfaces and the hello canaries of its connection.
func tlsCanaryRun(t *testing.T, f tlsCanaryFixture, id string, tc tlsCanaryCase) (*surfaces, canarySet) {
	t.Helper()
	h, d := tlsProductionHarness(t, f.tlsFixture)
	tlsPolicy(t, h, "8.8.4.4", 47392, tlsCanaryName, []string{"h2", "http/1.1"})
	d.authority = tlsCanaryAuthority
	var srv *loopback.TLSServer
	var rec *recordingConn
	if tc.server != nil {
		srv = tlsServe(t, d, *tc.server)
		dial := d.dial
		d.dial = func(ctx context.Context) (measurement.Conn, error) {
			c, err := dial(ctx)
			if err != nil {
				return nil, err
			}
			rec = &recordingConn{Conn: c.(net.Conn)}
			return rec, nil
		}
	} else {
		d.dial = func(context.Context) (measurement.Conn, error) {
			return nil, fmt.Errorf("CANARY-RAW-ERROR 8.8.4.4:47392 %s: %w", tlsCanaryName, tc.dialErr)
		}
	}
	if tc.stall {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if srv.Wait(ctx, func(c loopback.TLSCounts) bool { return c.In > 0 }) == nil {
				h.advance(tlsOperationBound)
			}
		}()
	}
	c := h.initialized(t)
	defer c.Close()
	data := tlsCanaryTask(id)
	if tc.data != nil {
		data = tc.data(data)
	}
	s := &surfaces{}
	rh, err := c.StartRun(h.admit(t, c, data))
	if err != nil {
		t.Fatal(codeOf(err))
	}
	if tc.interrupt != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Wait(ctx, func(c loopback.TLSCounts) bool { return c.In > 0 }); err != nil {
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
	rejectedClean(t, h, c, f.canaries, withUnknownInput(tlsCanaryTask("rejected"), f.canaries))
	var hello canarySet
	if rec != nil {
		hello = rec.helloCanaries()
	}
	return s, hello
}

// M-TLS-01-MINIMIZATION: canaries for every REQ-TLS-01-0028 item over the
// success path, every failure class and the incomplete and not_run paths.
// Handshake secrets and session tickets are created inside the TLS library
// and never cross the transport in plaintext, so no canary can be injected;
// the closed field assertion covers them.
// requirement-ref: REQ-TLS-01-0028
func TestMTls01Minimization(t *testing.T) {
	f := newTLSCanaryFixture(t)
	f.canaries.unique(t)
	failed := func(outcome, failure, validation, alert string) tlsExpect {
		return tlsExpect{state: "completed", outcome: outcome, tls: true, tlsFailure: failure, validation: validation, alert: alert}
	}
	tcpFailed := func(outcome, failure string) tlsExpect {
		return tlsExpect{state: "completed", outcome: outcome, tcpFailure: failure}
	}
	cases := []tlsCanaryCase{
		{name: "handshake_succeeded", server: &loopback.TLSConfig{Certificate: f.cert, ALPN: []string{"h2"}}, want: tlsSucceeded},
		{name: "tcp_timeout", dialErr: syscall.ETIMEDOUT, want: tcpFailed("tcp_timeout", "timeout")},
		{name: "tcp_refused", dialErr: syscall.ECONNREFUSED, want: tcpFailed("tcp_refused", "connection_refused")},
		{name: "tcp_reset", dialErr: syscall.ECONNRESET, want: tcpFailed("tcp_reset", "connection_reset")},
		{name: "tls_timeout", server: &loopback.TLSConfig{Mode: loopback.TLSStall}, stall: true, want: failed("tls_timeout", "timeout", "not_completed", "")},
		{name: "tls_alert", server: &loopback.TLSConfig{Mode: loopback.TLSAlert, Alert: 200}, want: failed("tls_alert", "alert", "not_completed", "other")},
		{name: "protocol_error", server: &loopback.TLSConfig{Mode: loopback.TLSGarbage}, want: failed("protocol_error", "protocol_error", "not_completed", "")},
		{name: "certificate_untrusted", server: &loopback.TLSConfig{Certificate: f.untrusted, ALPN: []string{"h2"}}, want: failed("certificate_untrusted", "certificate_untrusted", "untrusted", "")},
		{name: "certificate_name_mismatch", server: &loopback.TLSConfig{Certificate: f.mismatch, ALPN: []string{"h2"}}, want: failed("certificate_name_mismatch", "certificate_name_mismatch", "name_mismatch", "")},
		{name: "certificate_time_invalid", server: &loopback.TLSConfig{Certificate: f.expired, ALPN: []string{"h2"}}, want: failed("certificate_time_invalid", "certificate_time_invalid", "time_invalid", "")},
		{name: "alpn_no_overlap", server: &loopback.TLSConfig{Certificate: f.cert, ALPN: []string{"dot"}}, want: failed("alpn_no_overlap", "alpn_no_overlap", "not_completed", "")},
		{name: "inconclusive", dialErr: errors.New("CANARY-RAW-ERROR"), want: tcpFailed("inconclusive", "inconclusive")},
		{name: "incomplete", server: &loopback.TLSConfig{Mode: loopback.TLSStall}, interrupt: "cancel", want: tlsExpect{state: "incomplete", reason: "cancelled_by_host"}},
		{name: "not_run", data: func(s string) string {
			return strings.Replace(s, `"execution":{"limits":{"max_duration_ms":60000}`, `"execution":{"limits":{"max_duration_ms":60000},"network":{"context":{"required_fields":["country_code"]}}`, 1)
		}, want: tlsExpect{state: "not_run", reason: "network_context_unavailable"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, hello := tlsCanaryRun(t, f, "minimization", tc)
			if tc.server != nil && tc.server.Mode == loopback.TLSAccept && len(hello) == 0 {
				t.Fatal("no hello canaries derived")
			}
			s.check(t, append(slices.Clone(f.canaries), hello...))
			if len(s.results) != 1 {
				t.Fatal("not one result")
			}
			if m := minimized(t, s.results[0], tlsAllowed); (m == nil) != (tc.want.outcome == "") {
				t.Fatal("measurement presence")
			}
			tlsResult(t, s.results[0], tc.want)
		})
	}
	t.Run("injected fields", func(t *testing.T) {
		s, _ := tlsCanaryRun(t, f, "minimization", cases[0])
		obs := func(v map[string]any, i int) map[string]any {
			return v["measurement"].(map[string]any)["observations"].([]any)[i].(map[string]any)
		}
		var inject []func(map[string]any)
		for _, field := range []string{"server_name", "ip_address", "port", "certificate", "certificate_chain", "subject", "issuer", "san", "serial_number", "public_key",
			"signature", "fingerprint", "alert_code", "alert_description", "client_random", "server_random", "cipher_suites", "cipher_suite", "session_id", "session_ticket",
			"key_log", "error", "input_digest"} {
			inject = append(inject, func(v map[string]any) { obs(v, 1)["details"].(map[string]any)[field] = "synthetic" },
				func(v map[string]any) { obs(v, 1)[field] = "synthetic" }, func(v map[string]any) { v["measurement"].(map[string]any)[field] = "synthetic" })
		}
		inject = append(inject, func(v map[string]any) { obs(v, 1)["details"].(map[string]any)["alert"] = "alert(200)" })
		assertRejected(t, primitiveDecoder(t, "minimization", result.Output{Test: model.Test{ID: "TLS-01", Version: "0.1.0", ResultVersion: "0.1.0"}}), s.results[0], inject...)
	})
}

// M-TLS-01-INTERPRETATION: the shared prohibited-claims check with the
// TLS-01 claims, over results of the success and failure outcomes.
// requirement-ref: REQ-TLS-01-0001
// requirement-ref: REQ-TLS-01-0002
// requirement-ref: REQ-TLS-01-0003
// requirement-ref: REQ-TLS-01-0029
// requirement-ref: REQ-TLS-01-0030
// requirement-ref: REQ-TLS-01-0031
func TestMTls01Interpretation(t *testing.T) {
	f := newTLSCanaryFixture(t)
	var raws [][]byte
	for _, tc := range []tlsCanaryCase{
		{server: &loopback.TLSConfig{Certificate: f.cert, ALPN: []string{"h2"}}},
		{dialErr: syscall.ECONNREFUSED}, {dialErr: syscall.ECONNRESET}, {dialErr: syscall.ETIMEDOUT}, {dialErr: errors.New("synthetic")},
		{server: &loopback.TLSConfig{Mode: loopback.TLSAlert, Alert: 40}}, {server: &loopback.TLSConfig{Mode: loopback.TLSGarbage}},
		{server: &loopback.TLSConfig{Certificate: f.untrusted}}, {server: &loopback.TLSConfig{Certificate: f.mismatch}}, {server: &loopback.TLSConfig{Certificate: f.expired}},
		{server: &loopback.TLSConfig{Certificate: f.cert, ALPN: []string{"dot"}}},
	} {
		s, _ := tlsCanaryRun(t, f, "interpretation", tc)
		raws = append(raws, s.results...)
	}
	d := primitiveDecoder(t, "interpretation", result.Output{Test: model.Test{ID: "TLS-01", Version: "0.1.0", ResultVersion: "0.1.0"}})
	interpretation(t, "TLS-01", d, raws[0], raws, []string{"handshake_succeeded", "tcp_timeout", "tcp_refused", "tcp_reset", "tls_timeout", "tls_alert", "protocol_error",
		"certificate_untrusted", "certificate_name_mismatch", "certificate_time_invalid", "alpn_no_overlap", "inconclusive"}, claims(
		"handshake failure establishes filtering", `\b(handshake|tls)\b[^.]*\bfail\w*\b[^.]*\b(establish\w*|indicat\w*|shows?|proves?|means?)\b[^.]*\b(filter\w*|censor\w*|block\w*|interfer\w*|intent\w*)`,
		"certificate outcome establishes interception", `\bcertificate\w*\b[^.]*\b(establish\w*|indicat\w*|shows?|proves?|means?)\b[^.]*\b(intercept\w*|tamper\w*|mitm|man-in-the-middle)`,
		"timeout or reset excludes other causes", `\b(tcp_timeout|tcp_reset|tls_timeout)\b[^.]*\bexclud\w*`,
		"alpn or protocol error excludes configuration", `\b(alpn_no_overlap|protocol_error)\b[^.]*\bexclud\w*`,
		"handshake_succeeded establishes HTTP availability", `\bhandshake_succeeded\b[^.]*\b(establish\w*|shows?|proves?|means?)\b[^.]*\b(http|application|content)`,
		"TLS-01 compares SNI values", `\btls-01\b[^.]*\bcompar\w*\b[^.]*\b(sni|server names?)\b`,
	),
		"establishes whether tcp and normal authenticated tls completed for the configured address, port, server name and alpn offer",
		"does not establish tls or sni filtering, censorship, blocking, interference or intent", "does not establish interception or tampering",
		"does not establish http or application-content availability", "does not compare different sni values", "does not exclude a stale address-to-name mapping",
		"expired certificate or incorrect device time", "server-side rate limiting or overload")
}

// M-TLS-01-PRECONDITIONS: the portable gates, rejection before any TCP
// connection or ClientHello. A registration has no client-authentication,
// credential or sensitive-name marking: a target needing any of them is not
// registered, and the client holds no client certificate or credential
// (M-TLS-01-SAFETY); the review of that registration is RV-M-TLS-01-PRECONDITIONS.
// requirement-ref: REQ-TLS-01-0018
// requirement-ref: REQ-TLS-01-0019
// requirement-ref: REQ-TLS-01-0020
// requirement-ref: REQ-TLS-01-0021
// requirement-ref: REQ-TLS-01-0022
func TestMTls01Preconditions(t *testing.T) {
	f := tls01SafetyFixture(t)
	preconditionGates(t, f)
	t.Run("not registered for publicly authenticated TLS", func(t *testing.T) {
		// A TCP-01 registration of the same address and port does not
		// register the TLS tuple, nor does another server name or ALPN offer.
		f2 := f
		f2.start = func(t *testing.T) (*harness, func() int32) {
			h, count := f.start(t)
			editPolicy(t, h, func(in *task.PolicyInput) {
				in.Targets[0].Tuple = task.TargetTuple{Address: netip.MustParseAddr(tcpPublic), Port: 443, Scheme: "tcp"}
			})
			return h, count
		}
		f2.unregistered(t, tlsTask("tcp-registration"))
		f.unregistered(t, strings.Replace(tlsTask("name"), tlsServerName, "sensitive.example", 1), strings.Replace(tlsTask("alpn"), tlsOptions, `{"alpn_protocols":["h2"]}`, 1))
	})
	t.Run("all preconditions met", TestMTls01ExecutionAndOutput)
}
