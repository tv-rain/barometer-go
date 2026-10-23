package execution_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/execution"
	"github.com/tv-rain/barometer-go/internal/measurement"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/networkcontext"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit/capability"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

// Synthetic proxy fixtures. The connection_code is a reviewed-looking
// category; the proxy generation is local metadata that must never reach a
// result.
const (
	connectionCode  = "outline_mobileproxy"
	proxyGeneration = 7311
)

func scriptedProxy(t *testing.T, cfg loopback.ProxyConfig) *loopback.Proxy {
	t.Helper()
	p, err := loopback.ListenProxy(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = p.Close() })
	return p
}

// echoServer answers one payload-sized request with a payload-sized reply.
func echoServer(t *testing.T) *loopback.Server {
	t.Helper()
	return exchangeServer(t, len(loopback.Payload), len(loopback.Payload))
}

// echoOnce makes the Exchange runner send one payload-sized request on one
// connection and read the reply.
func echoOnce(srv *loopback.Server) loopback.ExchangeParameters {
	n := len(loopback.Payload)
	return loopback.ExchangeParameters{Address: srv.Address(), Connections: 1, Exchanges: 1, Request: n, Reply: n}
}

// closedPort returns a loopback port with no listener.
func closedPort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()
	return port
}

func proxyDialers(t *testing.T, port int) execution.ProxyDialers {
	t.Helper()
	d, err := execution.NewProxyDialers("127.0.0.1", port)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// proxyContext captures through the production proxy-path source.
type proxyContext struct {
	mu    sync.Mutex
	calls int
	src   *networkcontext.Source
}

func (c *proxyContext) Capture(ctx context.Context) (networkcontext.Snapshot, error) {
	c.mu.Lock()
	c.calls++
	c.mu.Unlock()
	return c.src.Capture(ctx)
}

func (c *proxyContext) Calls() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls
}

// onProxy rebinds the fixture to the proxy path of a scripted endpoint port:
// its dialers, its context source and a result decoder bound to that path.
func (f *fixture) onProxy(t *testing.T, port int) *proxyContext {
	t.Helper()
	src, err := networkcontext.NewProxy(capability.NewNetwork(defaultAccess), proxyGeneration, connectionCode)
	if err != nil {
		t.Fatal(err)
	}
	snap, err := src.Capture(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	d, err := result.New(result.Binding{TaskID: "campaign-1", Output: result.Output{Test: model.Test{ID: tcp01.ID, Version: tcp01.Version, ResultVersion: result.Contract}},
		MaxAttempts: 2, MaxDurationMS: 60000, Implementation: implementation, ContextDecoder: src.Decoder(), Path: snap.Path(), AccessGeneration: snap.Generation()})
	if err != nil {
		t.Fatal(err)
	}
	pc := &proxyContext{src: src}
	f.plan.Result, f.plan.Context, f.plan.Dialers = d, pc, proxyDialers(t, port)
	return pc
}

func idle(t *testing.T, p *loopback.Proxy) loopback.ProxyCounts {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := p.WaitIdle(ctx); err != nil {
		t.Fatal("proxy connections were not released")
	}
	return p.Counts()
}

func connectRequest(authority string) string {
	return "CONNECT " + authority + " HTTP/1.1\r\nHost: " + authority + "\r\n\r\n"
}

// assertProxyEgress checks the proxied result's egress and that no proxy
// endpoint, generation or CONNECT material reaches the serialized result.
func assertProxyEgress(t *testing.T, raw []byte, r record, port int) {
	t.Helper()
	_, egress := accessOf(t, r)
	if len(egress) != 4 || egress["status"] != "unavailable" || egress["lookup_source"] != "none" || egress["network_path"] != "proxy" || egress["connection_code"] != connectionCode {
		t.Fatalf("egress = %v", egress)
	}
	s := string(raw)
	for _, leak := range []string{strconv.Itoa(port), strconv.Itoa(proxyGeneration), "127.0.0.1", "CONNECT", "HTTP/1.1", "Forbidden", "canary"} {
		if strings.Contains(s, leak) {
			t.Fatalf("result contains %q: %s", leak, s)
		}
	}
}

// A proxy attempt runs its stream through one tunnel of the configured
// loopback endpoint: the endpoint reads exactly one bounded CONNECT for the
// test's authority with no credentials or cookies, the target is reached only
// through the tunnel, and the result records the proxy path with its
// connection_code. The meter counts the tunnel bytes, which equal the bytes
// the proxy relayed; the CONNECT exchange is not counted (REQ-LIMIT-0033).
// requirement-ref: REQ-PROXY-0017
// requirement-ref: REQ-PROXY-0031
// requirement-ref: REQ-LIMIT-0033
func TestProxyAttemptRunsThroughTunnel(t *testing.T) {
	srv := echoServer(t)
	p := scriptedProxy(t, loopback.ProxyConfig{})
	f := newFixture(t, schedule(1))
	f.onProxy(t, p.Port())
	f.plan.Parameters = echoOnce(srv)
	o := <-f.start(t, f.run, loopback.Exchange{})
	f.expect(t, o.rep, o.err, nil, []want{completedConnected}, nil)
	assertProxyEgress(t, f.store.Records()[0], f.records(t)[0], p.Port())
	waitAccepted(t, srv, 1)
	in, out := serverBytes(t, srv)
	c := idle(t, p)
	n := len(loopback.Payload)
	if c.Accepted != 1 || c.Relays != 1 || len(c.Requests) != 1 || c.Requests[0] != connectRequest(srv.Address()) {
		t.Fatalf("proxy counts = %+v", c)
	}
	if srv.Accepted() != 1 || in != n || out != n || c.TunnelIn != n || c.TunnelOut != n {
		t.Fatalf("target read %d wrote %d, tunnel %d/%d", in, out, c.TunnelIn, c.TunnelOut)
	}
	usage := o.rep.Attempts[0].Usage
	if usage != (measurement.Usage{BytesSent: n, BytesReceived: n, Connections: 1}) || c.HeadIn == 0 || c.HeadOut == 0 {
		t.Fatalf("metered %+v, CONNECT exchange %d/%d", usage, c.HeadIn, c.HeadOut)
	}
}

// Byte limits equal to the tunnel bytes are enough: the CONNECT request and
// response, larger than either limit, are below the metered stream. The
// tunnel still counts toward max_connections before it is opened.
// requirement-ref: REQ-LIMIT-0033
// requirement-ref: REQ-LIMIT-0034
func TestConnectExchangeIsNotMetered(t *testing.T) {
	const request, reply = 16, 24
	srv := exchangeServer(t, request, reply)
	p := scriptedProxy(t, loopback.ProxyConfig{})
	for _, tc := range []struct {
		name   string
		limits task.AttemptLimits
		want   want
	}{
		{"exact budgets", limits(request, reply, 1), completedConnected},
		{"no connection budget", limits(request, reply, 0), stoppedByBudget},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := p.Counts()
			f := newFixture(t, schedule(1))
			f.onProxy(t, p.Port())
			f.plan.Limits = tc.limits
			f.plan.Parameters = loopback.ExchangeParameters{Address: srv.Address(), Connections: 1, Exchanges: 1, Request: request, Reply: reply}
			o := <-f.start(t, f.run, loopback.Exchange{})
			f.expect(t, o.rep, o.err, nil, []want{tc.want}, nil)
			c := idle(t, p)
			if tc.want == stoppedByBudget {
				if c.Accepted != before.Accepted {
					t.Fatal("a tunnel was opened beyond max_connections")
				}
				return
			}
			head := c.HeadIn - before.HeadIn
			if head <= request || c.TunnelIn-before.TunnelIn != request || c.TunnelOut-before.TunnelOut != reply {
				t.Fatalf("CONNECT %d bytes, tunnel %d/%d", head, c.TunnelIn-before.TunnelIn, c.TunnelOut-before.TunnelOut)
			}
		})
	}
}

// Every proxy failure is an explicit execution state, never a protocol
// observation and never a direct dial: a closed endpoint port is
// proxy_unavailable, and a refused, non-HTTP/1.1, SOCKS, malformed, oversized
// or closed CONNECT is proxy_connect_failed. The runner turns the failed dial
// into a socket_error outcome, which is not retained (REQ-PROXY-0029). The
// attempt is incomplete because the dial started the test (REQ-RESULT-0051).
// requirement-ref: REQ-PROXY-0028
// requirement-ref: REQ-PROXY-0029
// requirement-ref: REQ-PROXY-0030
// requirement-ref: REQ-EXECUTION-0010
func TestProxyFailuresAreExecutionStates(t *testing.T) {
	failed := want{"incomplete", "proxy_connect_failed", ""}
	for _, tc := range []struct {
		name string
		head string
		want want
	}{
		{"endpoint unavailable", "", want{"incomplete", "proxy_unavailable", ""}},
		{"non-2xx", "HTTP/1.1 403 Forbidden\r\nX-Canary: canary\r\nContent-Length: 6\r\n\r\ncanary", failed},
		{"HTTP/1.0", "HTTP/1.0 200 OK\r\n\r\n", failed},
		{"three-digit 2xx only", "HTTP/1.1 2000 OK\r\n\r\n", failed},
		{"SOCKS reply", "\x05\x00", failed},
		{"malformed header", "HTTP/1.1 200 OK\r\nno colon\r\n\r\n", failed},
		{"oversized head", "HTTP/1.1 200 OK\r\nX-Canary: " + strings.Repeat("a", 9000) + "\r\n\r\n", failed},
		{"closed without reply", "-", failed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := echoServer(t)
			port := closedPort(t)
			var p *loopback.Proxy
			if tc.head != "" {
				head := []byte(tc.head)
				if tc.head == "-" {
					head = nil
				}
				p = scriptedProxy(t, loopback.ProxyConfig{Mode: loopback.ProxyReply, Head: head})
				port = p.Port()
			}
			f := newFixture(t, schedule(1))
			f.onProxy(t, port)
			f.plan.Parameters = echoOnce(srv)
			o := <-f.start(t, f.run, loopback.Exchange{})
			f.expect(t, o.rep, o.err, nil, []want{tc.want}, nil)
			assertProxyEgress(t, f.store.Records()[0], f.records(t)[0], port)
			if p != nil {
				if c := idle(t, p); c.Accepted != 1 || c.Relays != 0 {
					t.Fatalf("proxy counts = %+v", c)
				}
			}
			if srv.Accepted() != 0 {
				t.Fatal("the target was dialed outside the proxy")
			}
		})
	}
}

// A retry after proxy_unavailable uses a fresh dialer of the same proxy
// snapshot and never the direct path (REQ-PROXY-0030).
// requirement-ref: REQ-PROXY-0030
func TestProxyRetryNeverFallsBackToDirect(t *testing.T) {
	srv := echoServer(t)
	f := newFixture(t, schedule(2, task.ProxyUnavailable))
	f.onProxy(t, closedPort(t))
	f.plan.Parameters = echoOnce(srv)
	o := <-f.start(t, f.run, loopback.Exchange{})
	unavailable := want{"incomplete", "proxy_unavailable", ""}
	f.expect(t, o.rep, o.err, nil, []want{unavailable, unavailable}, []time.Duration{35 * time.Second})
	if srv.Accepted() != 0 {
		t.Fatal("a retry reached the target directly")
	}
}

// Each stream opens its own tunnel: a test that follows a redirect or opens a
// second connection gets a new CONNECT on the same endpoint, and a later
// attempt never reuses an earlier tunnel (REQ-PROXY-0018, REQ-PROXY-0020).
// requirement-ref: REQ-PROXY-0018
// requirement-ref: REQ-PROXY-0020
func TestEveryStreamOpensItsOwnTunnel(t *testing.T) {
	srv := exchangeServer(t, 4, 4)
	p := scriptedProxy(t, loopback.ProxyConfig{})
	f := newFixture(t, schedule(1))
	f.onProxy(t, p.Port())
	f.plan.Parameters = loopback.ExchangeParameters{Address: srv.Address(), Connections: 2, Exchanges: 1, Request: 4, Reply: 4}
	o := <-f.start(t, f.run, loopback.Exchange{})
	f.expect(t, o.rep, o.err, nil, []want{completedConnected}, nil)
	o = <-f.start(t, f.run, loopback.Exchange{})
	if o.err != nil {
		t.Fatal(o.err)
	}
	c := idle(t, p)
	if c.Accepted != 4 || c.Relays != 4 || len(c.Requests) != 4 || o.rep.Attempts[0].Usage.Connections != 2 {
		t.Fatalf("proxy counts = %+v", c)
	}
	waitAccepted(t, srv, 4)
}

// A CONNECT that does not complete within the attempt deadline is
// proxy_connect_failed; a host cancellation during the exchange stays
// cancelled_by_host. Either way the proxy socket is released and nothing is
// dialed directly.
// requirement-ref: REQ-PROXY-0017
// requirement-ref: REQ-PROXY-0028
func TestLateConnect(t *testing.T) {
	for _, tc := range []struct {
		name string
		host bool
		want want
	}{
		{"deadline", false, want{"incomplete", "proxy_connect_failed", ""}},
		{"host cancellation", true, want{"incomplete", "cancelled_by_host", ""}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := echoServer(t)
			p := scriptedProxy(t, loopback.ProxyConfig{Mode: loopback.ProxyHold})
			f := newFixture(t, schedule(1))
			f.onProxy(t, p.Port())
			f.plan.Parameters = echoOnce(srv)
			run, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := f.start(t, run, loopback.Exchange{})
			ctx, stop := context.WithTimeout(context.Background(), 10*time.Second)
			defer stop()
			if err := p.WaitRequests(ctx, 1); err != nil {
				t.Fatal("no CONNECT request")
			}
			if tc.host {
				cancel()
			} else {
				_ = f.clock.Advance(attemptDeadline)
			}
			o := <-done
			f.expect(t, o.rep, o.err, nil, []want{tc.want}, nil)
			idle(t, p)
			if srv.Accepted() != 0 {
				t.Fatal("the target was dialed outside the proxy")
			}
		})
	}
}

// pathCheck scripts the pre-attempt check.
type pathCheck task.ExecutionReason

func (c pathCheck) Check() task.ExecutionReason { return task.ExecutionReason(c) }

// The pre-attempt check (REQ-EXECUTION-0008) records why a proxy attempt
// cannot use the path as not_run, before any context capture or contact, and
// the run ends: none of these reasons is retryable. Admission keeps
// proxy_not_permitted and proxy_path_unsupported from arising for an admitted
// task in this build (REQ-PROXY-0042), so a scripted check stands in for them.
// requirement-ref: REQ-EXECUTION-0008
// requirement-ref: REQ-EXECUTION-0010
// requirement-ref: REQ-PROXY-0042
func TestPathCheckMakesNoContact(t *testing.T) {
	for _, reason := range []task.ExecutionReason{task.ProxyNotConfigured, task.ProxyNotPermitted, task.ProxyPathUnsupported} {
		t.Run(string(reason), func(t *testing.T) {
			p := scriptedProxy(t, loopback.ProxyConfig{})
			f := newFixture(t, schedule(2, task.ProxyUnavailable, task.InternalError))
			pc := f.onProxy(t, p.Port())
			f.plan.Path = pathCheck(reason)
			o := <-f.start(t, f.run, loopback.Exchange{})
			f.expect(t, o.rep, o.err, nil, []want{{"not_run", string(reason), ""}}, nil)
			if pc.Calls() != 0 || p.Counts().Accepted != 0 {
				t.Fatalf("captures %d, proxy accepts %d", pc.Calls(), p.Counts().Accepted)
			}
		})
	}
	// A passing check lets the attempt run.
	srv := echoServer(t)
	p := scriptedProxy(t, loopback.ProxyConfig{})
	f := newFixture(t, schedule(1))
	f.onProxy(t, p.Port())
	f.plan.Path = pathCheck("")
	f.plan.Parameters = echoOnce(srv)
	o := <-f.start(t, f.run, loopback.Exchange{})
	f.expect(t, o.rep, o.err, nil, []want{completedConnected}, nil)
}

// dialOnce opens one tunnel with a fresh proxy dialer.
func dialOnce(t *testing.T, port int, network, address string) (measurement.Conn, error) {
	t.Helper()
	d, err := proxyDialers(t, port).Dialer(1)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return d.DialContext(ctx, network, address)
}

// The CONNECT authority is the exact explicit IP address and port, or the
// normalized hostname and port, whose resolution happens behind the proxy
// (REQ-PROXY-0019). A dial HTTP CONNECT cannot carry, or an address that is
// not a valid authority, is refused without contacting the endpoint.
// requirement-ref: REQ-PROXY-0019
// requirement-ref: REQ-PROXY-0020
func TestConnectAuthority(t *testing.T) {
	srv, err := loopback.Listen()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = srv.Close() })
	_, port, _ := net.SplitHostPort(srv.Address())
	p := scriptedProxy(t, loopback.ProxyConfig{Resolve: map[string]string{"target.example.test:" + port: srv.Address()}})
	for _, tc := range []struct{ address, authority string }{
		{srv.Address(), srv.Address()},
		{"Target.Example.TEST:" + port, "target.example.test:" + port},
	} {
		before := len(p.Counts().Requests)
		c, err := dialOnce(t, p.Port(), "tcp", tc.address)
		if err != nil {
			t.Fatalf("%s: %v", tc.address, err)
		}
		if _, err := c.Write(loopback.Payload); err != nil {
			t.Fatal(err)
		}
		echo := make([]byte, len(loopback.Payload))
		if _, err := io.ReadFull(c, echo); err != nil || string(echo) != string(loopback.Payload) {
			t.Fatalf("tunnel echo %q, %v", echo, err)
		}
		_ = c.Close()
		reqs := p.Counts().Requests
		if len(reqs) != before+1 || reqs[before] != connectRequest(tc.authority) {
			t.Fatalf("requests = %q", reqs)
		}
	}
	// An IPv6 literal keeps its brackets; the scripted proxy refuses it.
	if _, err := dialOnce(t, p.Port(), "tcp", "[2001:db8::1]:443"); err == nil {
		t.Fatal("unrelayed authority succeeded")
	}
	if reqs := p.Counts().Requests; reqs[len(reqs)-1] != connectRequest("[2001:db8::1]:443") {
		t.Fatalf("IPv6 authority %q", reqs[len(reqs)-1])
	}
	before := p.Counts().Accepted
	for _, d := range []struct{ network, address string }{
		{"udp", srv.Address()},
		{"tcp4", srv.Address()},
		{"tcp6", "[::1]:443"},
		{"tcp", "example.test"},
		{"tcp", "example.test:0"},
		{"tcp", "example.test:65536"},
		{"tcp", "example.test:0443"},
		{"tcp", "exa_mple.test:443"},
		{"tcp", "a..b:443"},
		{"tcp", "-a.test:443"},
		{"tcp", "[fe80::1%en0]:443"},
		{"tcp", "example.test\r\nCookie: a:443"},
		{"tcp", strings.Repeat("a.", 127) + "ab:443"},
	} {
		if _, err := dialOnce(t, p.Port(), d.network, d.address); err == nil {
			t.Fatalf("%s %q accepted", d.network, d.address)
		}
	}
	idle(t, p)
	if p.Counts().Accepted != before {
		t.Fatal("a refused dial contacted the endpoint")
	}
}

// Tunnel bytes the endpoint sends with its 2xx response reach the stream
// first; header fields of the response, including framing fields, are
// ignored and retained nowhere.
func TestBytesAfterResponseHeadBelongToTheTunnel(t *testing.T) {
	p := scriptedProxy(t, loopback.ProxyConfig{Mode: loopback.ProxyReply, Head: []byte("HTTP/1.1 200\r\nContent-Length: 99\r\n\r\nearly")})
	c, err := dialOnce(t, p.Port(), "tcp", "127.0.0.1:9")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	got, _ := io.ReadAll(c)
	if string(got) != "early" {
		t.Fatalf("tunnel bytes %q", got)
	}
}

// The endpoint is trusted host configuration: only a loopback address or
// localhost and a port in range (REQ-PROXY-0005, REQ-PROXY-0007).
// requirement-ref: REQ-PROXY-0005
// requirement-ref: REQ-PROXY-0007
func TestProxyEndpointIsLoopback(t *testing.T) {
	for _, ok := range []struct {
		host string
		port int
	}{{"127.0.0.1", 1}, {"127.0.0.2", 65535}, {"::1", 54321}, {"localhost", 54321}} {
		if _, err := execution.NewProxyDialers(ok.host, ok.port); err != nil {
			t.Fatalf("%s:%d rejected", ok.host, ok.port)
		}
	}
	for _, bad := range []struct {
		host string
		port int
	}{{"192.0.2.1", 54321}, {"10.0.0.1", 54321}, {"::ffff:192.0.2.1", 54321}, {"fe80::1%lo0", 54321}, {"proxy.example.test", 54321}, {"", 54321}, {"127.0.0.1", 0}, {"127.0.0.1", 65536}} {
		if _, err := execution.NewProxyDialers(bad.host, bad.port); err == nil {
			t.Fatalf("%s:%d accepted", bad.host, bad.port)
		}
	}
	if _, err := (execution.ProxyDialers{}).Dialer(1); err == nil {
		t.Fatal("zero proxy dialers dialed")
	}
}

// The proxy egress representation decodes back through the strict context
// decoder bound to the proxy path.
func TestProxyContextRoundTrips(t *testing.T) {
	src, err := networkcontext.NewProxy(capability.NewNetwork(defaultAccess), proxyGeneration, connectionCode)
	if err != nil {
		t.Fatal(err)
	}
	snap, err := src.Capture(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	nc, err := snap.Context(false)
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(nc)
	if err != nil || !strings.Contains(string(b), `"connection_code":"`+connectionCode+`"`) || strings.Contains(string(b), strconv.Itoa(proxyGeneration)) {
		t.Fatalf("context %s, %v", b, err)
	}
	for _, bad := range []struct {
		g    uint64
		code string
	}{{0, connectionCode}, {1, ""}, {1, "Not-A-Category"}} {
		if _, err := networkcontext.NewProxy(capability.NewNetwork(defaultAccess), bad.g, bad.code); err == nil {
			t.Fatalf("proxy source %+v accepted", bad)
		}
	}
}

// tlsTarget is a loopback TLS echo server presenting a synthetic self-signed
// certificate for target.example.test. It records the SNI of each handshake.
type tlsTarget struct {
	ln    net.Listener
	roots *x509.CertPool
	mu    sync.Mutex
	sni   []string
	wg    sync.WaitGroup
}

const tlsName = "target.example.test"

func newTLSTarget(t *testing.T) *tlsTarget {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: tlsName}, DNSNames: []string{tlsName},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, BasicConstraintsValid: true, IsCA: true}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(cert)
	raw, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := &tlsTarget{roots: roots}
	s.ln = tls.NewListener(raw, &tls.Config{Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}}, MinVersion: tls.VersionTLS12,
		GetConfigForClient: func(h *tls.ClientHelloInfo) (*tls.Config, error) {
			s.mu.Lock()
			s.sni = append(s.sni, h.ServerName)
			s.mu.Unlock()
			return nil, nil
		}})
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			c, err := s.ln.Accept()
			if err != nil {
				return
			}
			s.wg.Add(1)
			go func() {
				defer s.wg.Done()
				defer c.Close()
				buf := make([]byte, len(loopback.Payload))
				if n, err := io.ReadFull(c, buf); err == nil {
					_, _ = c.Write(buf[:n])
				}
			}()
		}
	}()
	t.Cleanup(func() { _ = s.ln.Close(); s.wg.Wait() })
	return s
}

func (s *tlsTarget) Address() string { return s.ln.Addr().String() }

func (s *tlsTarget) SNI() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.sni...)
}

// netConn presents a tunnel as a net.Conn to crypto/tls, which needs
// addresses and split deadlines the runner contract deliberately omits.
type netConn struct{ measurement.Conn }

func (netConn) LocalAddr() net.Addr                  { return &net.TCPAddr{} }
func (netConn) RemoteAddr() net.Addr                 { return &net.TCPAddr{} }
func (c netConn) SetReadDeadline(t time.Time) error  { return c.SetDeadline(t) }
func (c netConn) SetWriteDeadline(t time.Time) error { return c.SetDeadline(t) }

// tlsEcho performs a TLS handshake for serverName inside conn, validating the
// certificate against roots, and exchanges one payload.
func tlsEcho(ctx context.Context, conn measurement.Conn, serverName string, roots *x509.CertPool) error {
	c := tls.Client(netConn{conn}, &tls.Config{ServerName: serverName, RootCAs: roots, MinVersion: tls.VersionTLS12})
	defer c.Close()
	if err := c.HandshakeContext(ctx); err != nil {
		return err
	}
	if _, err := c.Write(loopback.Payload); err != nil {
		return err
	}
	echo := make([]byte, len(loopback.Payload))
	if _, err := io.ReadFull(c, echo); err != nil {
		return err
	}
	if string(echo) != string(loopback.Payload) {
		return errors.New("tls echo mismatch")
	}
	return nil
}
