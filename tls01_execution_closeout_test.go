package barometer

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/netip"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

// M-TLS-01-EXECUTION "no implicit DNS": the trusted system resolver seam is
// installed and fails the test if called. The explicit IP is the only
// connection authority, on the one dial, and the server name never reaches a
// lookup, before or after the handshake.
// requirement-ref: REQ-TLS-01-0005
// requirement-ref: REQ-TLS-01-0006
func TestMTls01NoImplicitDNS(t *testing.T) {
	f := newTLSFixture(t)
	for _, name := range []string{"handshake", "tcp refused"} {
		t.Run(name, func(t *testing.T) {
			h, d := tlsProductionHarness(t, f)
			var lookups atomic.Int32
			h.adapters.run.resolver = dnsSystemResolver{func(context.Context, string, string) ([]netip.Addr, error) {
				lookups.Add(1)
				t.Error("TLS-01 performed a system lookup")
				return nil, nil
			}}
			want := tlsExpect{state: "completed", outcome: "tcp_refused", tcpFailure: "connection_refused"}
			if name == "handshake" {
				tlsServe(t, d, loopback.TLSConfig{Certificate: f.cert, ALPN: []string{"h2"}})
				want = tlsSucceeded
			}
			c := h.initialized(t)
			defer c.Close()
			tlsResult(t, tcpRaw(t, h, c, tlsTask("no-dns")), want)
			// tlsTransport refuses any authority other than the literal IP:port.
			if lookups.Load() != 0 || d.calls.Load() != 1 {
				t.Fatalf("lookups=%d dials=%d", lookups.Load(), d.calls.Load())
			}
		})
	}
}

// tlsSessionServer is a TLS 1.2 endpoint that issues session tickets, so a
// client keeping a session cache would resume on its next connection. It
// records whether each accepted handshake resumed and any plaintext received.
type tlsSessionServer struct {
	ln        net.Listener
	wg        sync.WaitGroup
	mu        sync.Mutex
	resumed   []bool
	plaintext int
}

func listenTLSSessionServer(t *testing.T, cert tls.Certificate) *tlsSessionServer {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := &tlsSessionServer{ln: ln}
	cfg := &tls.Config{Certificates: []tls.Certificate{cert}, NextProtos: []string{"h2"}, MinVersion: tls.VersionTLS12, MaxVersion: tls.VersionTLS12}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			raw, err := ln.Accept()
			if err != nil {
				return
			}
			s.wg.Add(1)
			go func() {
				defer s.wg.Done()
				defer raw.Close()
				_ = raw.SetDeadline(time.Now().Add(5 * time.Second))
				conn := tls.Server(raw, cfg)
				if conn.Handshake() != nil {
					return
				}
				n, _ := io.Copy(io.Discard, conn)
				s.mu.Lock()
				s.resumed = append(s.resumed, conn.ConnectionState().DidResume)
				s.plaintext += int(n)
				s.mu.Unlock()
			}()
		}
	}()
	t.Cleanup(func() { _ = ln.Close(); s.wg.Wait() })
	return s
}

func (s *tlsSessionServer) wait(t *testing.T, n int) ([]bool, int) {
	t.Helper()
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); time.Sleep(time.Millisecond) {
		s.mu.Lock()
		resumed, plaintext := append([]bool(nil), s.resumed...), s.plaintext
		s.mu.Unlock()
		if len(resumed) >= n {
			return resumed, plaintext
		}
	}
	t.Fatalf("server did not complete %d handshakes", n)
	return nil, 0
}

// M-TLS-01-EXECUTION "no sessions" at runtime: against a ticket-issuing TLS 1.2
// endpoint, two sequential attempts each perform a full handshake. The
// control client with a session cache proves the endpoint does resume, so the
// assertion is not vacuous.
// requirement-ref: REQ-TLS-01-0005
// requirement-ref: REQ-TLS-01-0017
func TestMTls01NoSessionResumption(t *testing.T) {
	f := newTLSFixture(t)
	control := listenTLSSessionServer(t, f.cert)
	cache := tls.NewLRUClientSessionCache(1)
	for i := 0; i < 2; i++ {
		conn, err := tls.Dial("tcp", control.ln.Addr().String(), &tls.Config{RootCAs: f.ca.Pool(), ServerName: tlsServerName, NextProtos: []string{"h2"}, ClientSessionCache: cache, MaxVersion: tls.VersionTLS12})
		if err != nil {
			t.Fatal(err)
		}
		_ = conn.Close()
	}
	if resumed, _ := control.wait(t, 2); !resumed[1] {
		t.Fatal("control endpoint does not resume; the test would be vacuous")
	}

	h, d := tlsProductionHarness(t, f)
	srv := listenTLSSessionServer(t, f.cert)
	d.dial = func(ctx context.Context) (measurement.Conn, error) {
		var nd net.Dialer
		return nd.DialContext(ctx, "tcp", srv.ln.Addr().String())
	}
	c := h.initialized(t)
	defer c.Close()
	want := tlsExpect{state: "completed", outcome: "handshake_succeeded", tls: true, validation: "trusted", version: "TLS1.2", alpn: "h2"}
	for _, id := range []string{"session-first", "session-second"} {
		tlsResult(t, tcpRaw(t, h, c, tlsTask(id)), want)
	}
	resumed, plaintext := srv.wait(t, 2)
	if len(resumed) != 2 || resumed[0] || resumed[1] || plaintext != 0 || d.calls.Load() != 2 {
		t.Fatalf("resumed=%v plaintext=%d dials=%d", resumed, plaintext, d.calls.Load())
	}
}
