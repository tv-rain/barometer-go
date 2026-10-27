package barometer

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/measurement/dns03"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

// M-DNS-04-ADMISSION: a case-folded resolver and query name, and the
// Unicode form of a registered A-label query name, match the profile and
// query name registered in canonical form.
// requirement-ref: REQ-DNS-04-0004
// requirement-ref: REQ-DNS-04-0009
// requirement-ref: REQ-DNS-04-0010
func TestMDns04NormalizedAdmission(t *testing.T) {
	f := newDotFixture(t)
	idna := strings.Replace(dotInput, "query.example", "xn--bcher-kva.example", 1)
	for _, tc := range []struct{ name, registered, input string }{
		{"case folded", dotInput, strings.Replace(strings.Replace(dotInput, dotName, "DoT.Resolver.EXAMPLE", 1), "query.example", "QUERY.Example", 1)},
		{"unicode query name", idna, strings.Replace(idna, "xn--bcher-kva.example", "bücher.example", 1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, d := dotHarness(t, f, tc.registered)
			c := h.initialized(t)
			defer c.Close()
			data := dotTask("normalized", tc.input, `{}`)
			if handle, err := c.AdmitTask(data, h.embedded(t, data)); err != nil || handle == "" {
				t.Fatal(codeOf(err))
			}
			if d.calls.Load() != 0 || d.lookups.Load() != 0 {
				t.Fatal("admission contacted transport or resolver")
			}
		})
	}
}

// dotClosing accepts one TLS connection with the fixture certificate, reads
// the framed query, writes reply and closes, so the stream can end inside
// the length prefix or the message. A nil reply resets the connection.
type dotClosing struct {
	accepted, plaintext atomic.Int32
	done                chan struct{}
}

func dotServeClosing(t *testing.T, d *dotTransport, cert tls.Certificate, reply func([]byte) []byte) *dotClosing {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	s := &dotClosing{done: make(chan struct{})}
	go func() {
		defer close(s.done)
		raw, err := ln.Accept()
		if err != nil {
			return
		}
		s.accepted.Add(1)
		conn := tls.Server(raw, &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12})
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if conn.Handshake() != nil {
			return
		}
		prefix := make([]byte, 2)
		if _, err := io.ReadFull(conn, prefix); err != nil {
			return
		}
		q := make([]byte, int(prefix[0])<<8|int(prefix[1]))
		n, _ := io.ReadFull(conn, q)
		s.plaintext.Add(int32(2 + n))
		if reply == nil {
			_ = raw.(*net.TCPConn).SetLinger(0)
			_ = raw.Close()
			return
		}
		_, _ = conn.Write(reply(q[:n]))
	}()
	d.dial = func(ctx context.Context) (measurement.Conn, error) {
		var nd net.Dialer
		return nd.DialContext(ctx, "tcp", ln.Addr().String())
	}
	return s
}

// M-DNS-04-EXECUTION fault injection: the stream ends inside the length
// prefix, inside the message or right after the query, or the prefix
// announces more than is sent; each is a completed malformed_dns_response at
// the DoT stage, not an incomplete state, and a reset after the query keeps
// the DoT stage as inconclusive. With these faults, a query mismatch,
// a short message and a TLS failure, the attempt makes one dial and one
// connection, writes nothing after the framed query, and never falls back to
// cleartext, another resolver or a retry.
// requirement-ref: REQ-DNS-04-0005
// requirement-ref: REQ-DNS-04-0006
// requirement-ref: REQ-DNS-04-0008
// requirement-ref: REQ-DNS-04-0028
func TestMDns04ExchangeFaults(t *testing.T) {
	f := newDotFixture(t)
	malformed := dotExpect{state: "completed", outcome: "malformed_dns_response", types: dotBootstrapped, failure: "malformed_response"}
	for _, tc := range []struct {
		name  string
		reply func([]byte) []byte
		want  dotExpect
	}{
		{"closed inside prefix", func([]byte) []byte { return []byte{0} }, malformed},
		{"closed after query", func([]byte) []byte { return nil }, malformed},
		{"closed inside message", func(q []byte) []byte { return dotAnswer(q)[:20] }, malformed},
		{"prefix beyond stream", func(q []byte) []byte {
			r := dotAnswer(q)
			r[0], r[1] = 0xff, 0xff
			return r
		}, malformed},
		{"reset after query", nil, dotExpect{state: "completed", outcome: "inconclusive", types: dotBootstrapped, failure: "inconclusive"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, d := dotHarness(t, f, dotInput)
			srv := dotServeClosing(t, d, f.cert.Certificate, tc.reply)
			c := h.initialized(t)
			defer c.Close()
			dotResult(t, tcpRaw(t, h, c, dotTask("fault", dotInput, `{}`)), tc.want)
			if _, err := c.NextBundle(); err != nil {
				t.Fatal("result did not survive bundle validation", err)
			}
			<-srv.done
			if d.calls.Load() != 1 || d.lookups.Load() != 0 || srv.accepted.Load() != 1 || srv.plaintext.Load() != 2+dotQueryBytes {
				t.Fatalf("dials %d lookups %d accepted %d plaintext %d", d.calls.Load(), d.lookups.Load(), srv.accepted.Load(), srv.plaintext.Load())
			}
		})
	}
	other, err := loopback.NewTestCA()
	if err != nil {
		t.Fatal(err)
	}
	untrusted, err := other.Issue(dotName, false)
	if err != nil {
		t.Fatal(err)
	}
	reply := func(fn func([]byte) []byte) loopback.TLSConfig { c := f.cert; c.DoT = fn; return c }
	for _, tc := range []struct {
		name      string
		server    loopback.TLSConfig
		plaintext int
		want      dotExpect
	}{
		{"query mismatch", reply(func(q []byte) []byte {
			r := loopback.DNSResponse(q, 0, nil)
			r[0] ^= 0xff
			return dotFrame(r)
		}), 2 + dotQueryBytes, dotExpect{state: "completed", outcome: "query_mismatch", types: dotBootstrapped, failure: "query_mismatch"}},
		{"short message", reply(func([]byte) []byte { return []byte{0, 11, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11} }), 2 + dotQueryBytes, malformed},
		{"tls failure", loopback.TLSConfig{Certificate: untrusted}, 0, dotExpect{state: "completed", outcome: "tls_failure", types: []string{"tcp_connect", "tls_handshake"}, failure: "certificate_untrusted", validation: "untrusted"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, d := dotHarness(t, f, dotInput)
			srv := dotServe(t, d, tc.server)
			c := h.initialized(t)
			defer c.Close()
			dotResult(t, tcpRaw(t, h, c, dotTask("fault", dotInput, `{}`)), tc.want)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.WaitIdle(ctx); err != nil {
				t.Fatal(err)
			}
			if n := srv.Counts(); d.calls.Load() != 1 || d.lookups.Load() != 0 || n.Accepted != 1 || n.Plaintext != tc.plaintext {
				t.Fatalf("dials %d lookups %d server %+v", d.calls.Load(), d.lookups.Load(), n)
			}
		})
	}
}

// M-DNS-04-VERSIONS: results produced by the DNS-03 and DNS-01 runners
// validate under their own binding but are rejected under the DNS-04
// 0.1.0 binding, with their own test.id and relabelled as DNS-04.
// requirement-ref: REQ-DNS-04-0024
func TestMDns04CrossTestOutputs(t *testing.T) {
	cd, _ := observation.NewContextDecoder(result.Contract, observation.ContextApprovals{})
	path, _ := observation.NewPath(1, 0, "direct", "")
	decoder := func(o result.Output) *result.Decoder {
		o.Limits.RetainDNSAnswers = true
		d, err := result.New(result.Binding{TaskID: "versions", Output: o, MaxAttempts: 1, MaxDurationMS: 60000, Implementation: implementation, ContextDecoder: cd, Path: path, AccessGeneration: 1})
		if err != nil {
			t.Fatal(err)
		}
		return d
	}
	doh := func(t *testing.T) []byte {
		f := newDohFixture(t)
		h, d := dohHarness(t, f, dohInput)
		dohServe(t, d, f.reply(dohAnswer))
		c := h.initialized(t)
		defer c.Close()
		return tcpRaw(t, h, c, dohTask("versions", dohInput, `{}`))
	}
	udp := func(t *testing.T) []byte {
		h, d := dnsHarness(t, "tcp")
		dnsLoopback(t, d, oneAnswer)
		c := h.initialized(t)
		defer c.Close()
		return tcpRaw(t, h, c, dnsTask("versions", "tcp"))
	}
	dot := decoder(result.Output{Test: model.Test{ID: "DNS-04", Version: "0.1.0", ResultVersion: "0.1.0"}, QueryType: "A"})
	for _, tc := range []struct {
		name string
		raw  func(*testing.T) []byte
		own  result.Output
	}{
		{"DNS-03", doh, result.Output{Test: model.Test{ID: "DNS-03", Version: "0.1.0", ResultVersion: "0.1.0"}, QueryType: "A", Limits: observation.ProtocolLimits{MaxBodyBytes: dns03.MaxBody}}},
		{"DNS-01", udp, result.Output{Test: model.Test{ID: "DNS-01", Version: "0.1.0", ResultVersion: "0.1.0"}, QueryType: "A", DNSMode: "tcp"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := tc.raw(t)
			if _, err := decoder(tc.own).Decode(raw); err != nil {
				t.Fatal("own result rejected", err)
			}
			if _, err := dot.Decode(raw); err == nil {
				t.Fatalf("%s output accepted as DNS-04", tc.name)
			}
			var v map[string]any
			_ = json.Unmarshal(raw, &v)
			v["test"].(map[string]any)["id"] = "DNS-04"
			relabelled, _ := json.Marshal(v)
			if _, err := dot.Decode(relabelled); err == nil {
				t.Fatalf("%s output relabelled as DNS-04 accepted", tc.name)
			}
		})
	}
}
