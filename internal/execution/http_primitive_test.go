package execution

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	"github.com/tv-rain/barometer-go/internal/model/task"
	httpprotocol "github.com/tv-rain/barometer-go/internal/protocol/http"
	"github.com/tv-rain/barometer-go/internal/protocol/tcp"
	"github.com/tv-rain/barometer-go/internal/protocol/tls"
	"github.com/tv-rain/barometer-go/internal/testkit"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

// httpEndpoint is a test caller, not a registered measurement. It composes
// the production target gate, attempt meter, TCP and TLS primitives without
// bypassing any destination check. Future HTTP-01/DNS-03/NET-03 callers own
// their test-specific tuples, resolution observations and result mapping.
type httpEndpoint struct {
	dial     measurement.TargetDialer
	clock    measurement.Clock
	roots    *x509.CertPool
	tuple    task.TargetTuple
	bodyRead chan struct{}
}
type httpPlain struct {
	io.ReadWriter
	close    func() error
	reads    int
	bodyRead chan struct{}
}

func (c *httpPlain) Close() error { return c.close() }
func (c *httpPlain) Read(p []byte) (int, error) {
	c.reads++
	if c.reads == 2 && c.bodyRead != nil {
		close(c.bodyRead)
	}
	return c.ReadWriter.Read(p)
}
func (e httpEndpoint) Open(ctx context.Context, u *url.URL) (httpprotocol.Stream, []observation.Observation, string, error) {
	port := u.Port()
	if port == "" {
		port = "443"
	}
	var destinationError error
	c, tcpObs, err := tcp.Connect(ctx, e.clock, "ipv4", func(op context.Context) (io.ReadWriteCloser, error) {
		c, err := e.dial.DialContext(op, "tcp", net.JoinHostPort(u.Hostname(), port))
		if errors.Is(err, measurement.ErrTargetRejected) {
			destinationError = &httpprotocol.DestinationRejected{Class: "prohibited_host"}
			return nil, destinationError
		}
		return c, err
	})
	if err != nil {
		return httpprotocol.Stream{}, nil, "", err
	}
	if destinationError != nil {
		return httpprotocol.Stream{}, nil, "", destinationError
	}
	obs := []observation.Observation{tcpObs}
	if c == nil {
		return httpprotocol.Stream{}, obs, "tcp_failure", nil
	}
	sc, ok := c.(measurement.SessionConn)
	if !ok {
		_ = c.Close()
		return httpprotocol.Stream{}, nil, "", measurement.ErrRunner
	}
	raw, meter, err := sc.Session()
	if err != nil {
		_ = c.Close()
		return httpprotocol.Stream{}, nil, "", err
	}
	session, tlsObs, err := tls.Handshake(ctx, e.clock, "ipv4", raw, tls.Config{ServerName: u.Hostname(), ALPN: []string{"http/1.1"}, Roots: e.roots})
	if err != nil {
		return httpprotocol.Stream{}, nil, "", err
	}
	obs = append(obs, tlsObs)
	if session == nil {
		return httpprotocol.Stream{}, obs, "tls_failure", nil
	}
	return httpprotocol.Stream{Conn: &httpPlain{ReadWriter: meter(session), close: c.Close, bodyRead: e.bodyRead}, Family: "ipv4"}, obs, "", nil
}
func (e httpEndpoint) Redirect(u *url.URL) (httpprotocol.Endpoint, string, error) {
	if ip, err := netip.ParseAddr(u.Hostname()); err == nil && !task.PublicAddress(ip) {
		return nil, "prohibited_host", ErrTargetRejected
	}
	tuple := e.tuple
	tuple.URL = u.String()
	tuple.Hostname = u.Hostname()
	tuple.Port = 443
	if u.Port() != "" {
		p, _ := strconv.Atoi(u.Port())
		tuple.Port = uint16(p)
	}
	d, err := e.dial.Redirect(tuple)
	if err != nil {
		return nil, "unregistered_https", err
	}
	e.dial = d
	e.tuple = tuple
	return e, "registered_https", nil
}

func httpFixture(t *testing.T, raw string, stall, proxy bool, limits task.AttemptLimits) (*scope, httpEndpoint, *loopback.HTTPServer, *loopback.Proxy, *testkit.Clock, context.Context, context.CancelCauseFunc) {
	t.Helper()
	ca, err := loopback.NewTestCA()
	if err != nil {
		t.Fatal(err)
	}
	cert, err := ca.Issue("8.8.8.8", false)
	if err != nil {
		t.Fatal(err)
	}
	srv, err := loopback.ListenHTTP(loopback.HTTPConfig{Certificate: cert, Raw: []byte(raw), Stall: stall})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = srv.Close() })
	var dial measurement.Dialer = loopbackDial{srv.Address()}
	var p *loopback.Proxy
	if proxy {
		p, err = loopback.ListenProxy(loopback.ProxyConfig{Resolve: map[string]string{"8.8.8.8:443": srv.Address()}})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = p.Close() })
		pd, err := NewProxyDialers("127.0.0.1", p.Port())
		if err != nil {
			t.Fatal(err)
		}
		dial, err = pd.Dialer(1)
		if err != nil {
			t.Fatal(err)
		}
	}
	clock := testkit.NewClock(time.Now())
	ctx, cancel := context.WithCancelCause(context.Background())
	s := newScope(ctx, dial, limits, cancel)
	t.Cleanup(func() { cancel(nil); s.close() })
	reg := targetRegistration("initial", "8.8.8.8", task.MeasurementTarget, clock.Now())
	path := "direct"
	if proxy {
		path = "proxy"
	}
	s.targets = NewTargets(reg.Test, path, []task.TargetRegistration{reg}, clock, nil)
	return s, httpEndpoint{dial: s, clock: clock, roots: ca.Pool(), tuple: reg.Tuple}, srv, p, clock, ctx, cancel
}
func httpConfig() httpprotocol.Config {
	return httpprotocol.Config{URL: "https://8.8.8.8/check", MaxBodyBytes: 4096}
}

// HTTP rows at primitive level. Registered callers remain separate tasks.
// Exact plaintext budgets include request/status lines, headers and bodies;
// handshake, TLS records and the proxy CONNECT exchange stay outside them.
// requirement-ref: REQ-LIMIT-0033
// requirement-ref: REQ-LIMIT-0034
func TestCcLimitPurposeHTTPRows(t *testing.T) {
	body := "synthetic-body"
	raw := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
	for _, proxy := range []bool{false, true} {
		for _, post := range []bool{false, true} {
			t.Run(fmt.Sprintf("proxy=%t/post=%t", proxy, post), func(t *testing.T) {
				response := raw
				if post {
					response = strings.Replace(raw, "Content-Length:", "Content-Type: application/dns-message\r\nContent-Length:", 1)
				}
				var exact measurement.Usage
				for _, mode := range []string{"baseline", "exact", "sent-short", "received-short"} {
					limits := task.AttemptLimits{BytesSent: 32768, BytesReceived: 131072, Connections: 1}
					if mode != "baseline" {
						limits.BytesSent = exact.BytesSent
						limits.BytesReceived = exact.BytesReceived
					}
					if mode == "sent-short" {
						limits.BytesSent--
					}
					if mode == "received-short" {
						limits.BytesReceived--
					}
					s, e, srv, p, _, ctx, _ := httpFixture(t, response, false, proxy, limits)
					cfg := httpConfig()
					if post {
						cfg.PostBody = []byte("dns-wire-canary")
					}
					r, close, err := httpprotocol.Exchange(ctx, e.clock, e, cfg)
					close()
					s.close()
					if mode == "sent-short" || mode == "received-short" {
						if err == nil || !errors.Is(context.Cause(ctx), errBudget) {
							t.Fatal(mode, err, context.Cause(ctx))
						}
						continue
					}
					if err != nil || r.Outcome != "expected_response" || len(r.Observations) != 3 {
						t.Fatal(mode, err, r.Outcome, len(r.Observations))
					}
					wait, done := context.WithTimeout(context.Background(), 2*time.Second)
					defer done()
					if err := srv.Wait(wait, func(c loopback.HTTPCounts) bool { return c.Accepted == 1 && c.Open == 0 }); err != nil {
						t.Fatal(err)
					}
					counts := srv.Counts()
					u := s.Usage()
					if u.BytesSent != counts.In || u.BytesReceived != len(response) || u.Connections != 1 || counts.WireIn <= u.BytesSent || counts.WireOut <= u.BytesReceived {
						t.Fatalf("%s meter %+v server %+v", mode, u, counts)
					}
					if mode == "baseline" {
						exact = u
					} else if u != exact {
						t.Fatal("exact bound changed usage")
					}
					if len(counts.Requests) != 1 || counts.Requests[0].Header.Get("Accept-Encoding") != "identity" || counts.Requests[0].Header.Get("User-Agent") != "" {
						t.Fatal("request profile")
					}
					if post && string(r.TakeBody()) != body {
						t.Fatal("POST payload lost")
					}
					if p != nil {
						if err := p.WaitIdle(wait); err != nil {
							t.Fatal(err)
						}
						pc := p.Counts()
						if pc.Accepted != 1 || pc.TunnelIn != counts.WireIn || pc.TunnelOut != counts.WireOut || pc.HeadIn == 0 || pc.HeadOut == 0 {
							t.Fatalf("proxy %+v", pc)
						}
					}
				}
			})
		}
	}
}

// requirement-ref: REQ-LIMIT-0008
// requirement-ref: REQ-OBS-0023
func TestHTTPBodyCancellationAndDeadline(t *testing.T) {
	for _, proxy := range []bool{false, true} {
		for _, cause := range []string{"host", "outer-deadline", "operation"} {
			t.Run(fmt.Sprintf("%t/%s", proxy, cause), func(t *testing.T) {
				s, e, srv, _, clock, ctx, cancel := httpFixture(t, "HTTP/1.1 200 OK\r\nContent-Length: 20\r\n\r\nx", true, proxy, task.AttemptLimits{BytesSent: 32768, BytesReceived: 131072, Connections: 1})
				e.bodyRead = make(chan struct{})
				type answer struct {
					r   httpprotocol.Result
					err error
				}
				done := make(chan answer, 1)
				go func() {
					r, close, err := httpprotocol.Exchange(ctx, clock, e, httpConfig())
					close()
					done <- answer{r, err}
				}()
				wait, stop := context.WithTimeout(context.Background(), 2*time.Second)
				defer stop()
				if err := srv.Wait(wait, func(c loopback.HTTPCounts) bool { return c.Out > 0 }); err != nil {
					t.Fatal(err)
				}
				select {
				case <-e.bodyRead:
				case <-wait.Done():
					t.Fatal("body read did not begin")
				}
				switch cause {
				case "host":
					cancel(context.Canceled)
				case "outer-deadline":
					cancel(context.DeadlineExceeded)
				case "operation":
					if err := clock.Advance(15 * time.Second); err != nil {
						t.Fatal(err)
					}
				}
				select {
				case a := <-done:
					if cause == "operation" {
						if a.err != nil || a.r.Outcome != "timeout" {
							t.Fatal(a.err, a.r.Outcome)
						}
					} else if a.err == nil || a.r.Outcome != "" {
						t.Fatal("outer interruption became protocol outcome")
					}
				case <-wait.Done():
					t.Fatal("body did not unblock within cleanup bound")
				}
				if err := srv.Wait(wait, func(c loopback.HTTPCounts) bool { return c.Open == 0 }); err != nil {
					t.Fatal(err)
				}
				if s.Usage().Connections != 1 {
					t.Fatal("unexpected connection count")
				}
			})
		}
	}
}

// The redirect child is checked before contact on both paths, including
// hostname proxy rejection before resolution or CONNECT. No synthetic runner
// registers HTTP-01, DNS-03 or NET-03 in production.
// requirement-ref: REQ-HTTP-01-0018
// requirement-ref: REQ-HTTP-01-0051
func TestHTTPRedirectUsesProductionDestinationGate(t *testing.T) {
	for _, proxy := range []bool{false, true} {
		for _, host := range []string{"9.9.9.9", "next.example", "127.0.0.1", "unregistered.example"} {
			t.Run(fmt.Sprintf("%t/%s", proxy, host), func(t *testing.T) {
				raw := "HTTP/1.1 302 Found\r\nLocation: https://" + host + "/check\r\nContent-Length: 0\r\n\r\n"
				s, e, _, _, clock, ctx, _ := httpFixture(t, raw, false, proxy, task.AttemptLimits{BytesSent: 32768, BytesReceived: 131072, Connections: 3})
				initial := targetRegistration("initial", "8.8.8.8", task.MeasurementTarget, clock.Now())
				initial.RedirectIDs = []string{"next"}
				hop := targetRegistration("next", host, task.RedirectTarget, clock.Now())
				regs := []task.TargetRegistration{initial, hop}
				if host == "unregistered.example" {
					regs = regs[:1]
				}
				path := "direct"
				if proxy {
					path = "proxy"
				}
				gate := NewTargets(initial.Test, path, regs, clock, nil)
				resolver := &targetResolver{answers: nil}
				gate.resolver = resolver
				s.targets = gate
				cfg := httpConfig()
				cfg.MaxRedirects = 2
				r, close, err := httpprotocol.Exchange(ctx, clock, e, cfg)
				close()
				if proxy && host == "next.example" || host == "unregistered.example" {
					if err != nil || r.Outcome != "redirect_rejected" || resolver.calls != 0 {
						t.Fatal(err, r.Outcome, resolver.calls)
					}
				}
				if host == "127.0.0.1" {
					if err != nil || r.Outcome != "redirect_rejected" {
						t.Fatal(err, r.Outcome)
					}
				}
				if host != "9.9.9.9" && s.Usage().Connections > 1 {
					t.Fatal("rejected target contacted")
				}
			})
		}
	}
}

type httpRoutes map[string]string

func (d httpRoutes) DialContext(ctx context.Context, network, address string) (measurement.Conn, error) {
	local, ok := d[address]
	if !ok {
		return nil, measurement.ErrTargetRejected
	}
	var dial net.Dialer
	return dial.DialContext(ctx, network, local)
}

// requirement-ref: REQ-HTTP-01-0018
// requirement-ref: REQ-OBS-0026
// requirement-ref: REQ-LIMIT-0034
func TestHTTPApprovedRedirectChain(t *testing.T) {
	for _, proxy := range []bool{false, true} {
		t.Run(fmt.Sprint(proxy), func(t *testing.T) {
			ca, err := loopback.NewTestCA()
			if err != nil {
				t.Fatal(err)
			}
			clock := testkit.NewClock(time.Now())
			routes := httpRoutes{}
			var servers []*loopback.HTTPServer
			var regs []task.TargetRegistration
			ips := []string{"8.8.8.8", "9.9.9.9", "1.1.1.1"}
			responses := []string{
				"HTTP/1.1 302 Found\r\nLocation: https://9.9.9.9/check\r\nContent-Length: 0\r\n\r\n",
				"HTTP/1.1 307 Redirect\r\nLocation: https://1.1.1.1/check\r\nContent-Length: 0\r\n\r\n",
				"HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nok",
			}
			for i, ip := range ips {
				cert, err := ca.Issue(ip, false)
				if err != nil {
					t.Fatal(err)
				}
				srv, err := loopback.ListenHTTP(loopback.HTTPConfig{Certificate: cert, Raw: []byte(responses[i])})
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = srv.Close() })
				servers = append(servers, srv)
				routes[ip+":443"] = srv.Address()
				role := task.RedirectTarget
				if i == 0 {
					role = task.MeasurementTarget
				}
				reg := targetRegistration(fmt.Sprint(i), ip, role, clock.Now())
				if i < 2 {
					reg.RedirectIDs = []string{fmt.Sprint(i + 1)}
				}
				regs = append(regs, reg)
			}
			var dial measurement.Dialer = routes
			path := "direct"
			if proxy {
				path = "proxy"
				p, err := loopback.ListenProxy(loopback.ProxyConfig{Resolve: routes})
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = p.Close() })
				source, err := NewProxyDialers("127.0.0.1", p.Port())
				if err != nil {
					t.Fatal(err)
				}
				dial, err = source.Dialer(1)
				if err != nil {
					t.Fatal(err)
				}
			}
			ctx, cancel := context.WithCancelCause(context.Background())
			defer cancel(nil)
			s := newScope(ctx, dial, task.AttemptLimits{BytesSent: 32768, BytesReceived: 131072, Connections: 3}, cancel)
			defer s.close()
			s.targets = NewTargets(regs[0].Test, path, regs, clock, nil)
			e := httpEndpoint{dial: s, clock: clock, roots: ca.Pool(), tuple: regs[0].Tuple}
			cfg := httpConfig()
			cfg.MaxRedirects = 2
			r, cleanup, err := httpprotocol.Exchange(ctx, clock, e, cfg)
			defer cleanup()
			if err != nil || r.Outcome != "expected_response" || len(r.Observations) != 9 {
				t.Fatal(err, r.Outcome, len(r.Observations))
			}
			cleanup()
			for i, want := range []int{1, 2, 2} {
				d, ok := r.Observations[i*3+2].HTTP()
				if !ok || d.Redirect == nil || d.Redirect.FollowedCount != want {
					t.Fatalf("hop %d: %+v", i, d.Redirect)
				}
			}
			wait, stop := context.WithTimeout(context.Background(), 2*time.Second)
			defer stop()
			sent, received := 0, 0
			for _, srv := range servers {
				if err := srv.Wait(wait, func(c loopback.HTTPCounts) bool { return c.Accepted == 1 && c.Open == 0 }); err != nil {
					t.Fatal(err)
				}
				c := srv.Counts()
				sent += c.In
				received += c.Out
			}
			if s.Usage() != (measurement.Usage{BytesSent: sent, BytesReceived: received, Connections: 3}) {
				t.Fatal("redirect accounting", s.Usage(), sent, received)
			}
		})
	}
}
