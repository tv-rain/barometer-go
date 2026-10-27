package execution

import (
	"context"
	"net/netip"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/measurement/dns03"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/networkcontext"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

const (
	dohResolver   = "doh.resolver.example"
	dohBootstrap  = `{"resolver_url":"https://doh.resolver.example/dns-query","hostname":"query.example","record_type":"A","bootstrap_ip":"8.8.8.8"}`
	dohSystem     = `{"resolver_url":"https://doh.resolver.example/dns-query","hostname":"query.example","record_type":"A"}`
	dohQueryBytes = 31 // header, query.example and question tail
)

var dns03Test = task.TestVersion{ID: "DNS-03", Version: "0.1.0"}

func dns03Params(t *testing.T, input string) dns03.Parameters {
	t.Helper()
	codec := dns03.Codec{}
	def := codec.Definition()
	lim := jsondecode.Limits{MaxBytes: 1024, MaxDepth: 8, MaxNodes: 32}
	in, err := jsondecode.Decode([]byte(input), def.Input, lim)
	if err != nil {
		t.Fatal(err)
	}
	opt, err := jsondecode.Decode([]byte(`{}`), def.Options, lim)
	if err != nil {
		t.Fatal(err)
	}
	p, err := codec.Normalize(in, opt)
	if err != nil {
		t.Fatal(err)
	}
	return p.(dns03.Parameters)
}

// dns03Gate registers the query name and the resolver profile in the shape
// of the production target rule, on the given path.
func dns03Gate(clock interface{ Now() time.Time }, p dns03.Parameters, path string, resolver SystemResolver) *Targets {
	u, host, port, bootstrap := p.Resolver()
	profile := task.TargetTuple{URL: u, Hostname: host, ResolverName: host, TLSServerName: host, Port: port, Scheme: "https"}
	if bootstrap.IsValid() {
		profile.BootstrapAddresses = []netip.Addr{bootstrap}
	}
	paths := task.Paths{Direct: path == "direct", Proxy: path == "proxy"}
	expires := task.TimestampFromTime(clock.Now().Add(time.Hour))
	regs := []task.TargetRegistration{
		{ID: "doh-query", Test: dns03Test, Role: task.MeasurementTarget, Tuple: task.TargetTuple{Hostname: p.Name()}, Paths: paths, ExpiresAt: expires},
		{ID: "doh-resolver", Test: dns03Test, Role: task.ResolverTarget, Tuple: profile, Paths: paths, ExpiresAt: expires},
	}
	return NewTargets(dns03Test, path, regs, clock, nil).WithResolver(resolver)
}

func dns03Plan(t *testing.T, proxy bool, limits task.AttemptLimits, p dns03.Parameters, dial DialerSource, gate TargetCheck) Plan {
	t.Helper()
	plan := tls01Plan(t, proxy, 30*time.Second, limits, dial, gate)
	plan.TaskID, plan.Test, plan.Parameters = "dns03-plan", dns03Test, p
	source := plan.Context.(*networkcontext.Source)
	snap, err := source.Capture(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	out := result.Output{Test: model.Test{ID: "DNS-03", Version: "0.1.0", ResultVersion: "0.1.0"}, QueryType: p.QueryType(), Bootstrap: p.Bootstrap()}
	out.Limits.MaxBodyBytes = dns03.MaxBody
	out.Limits.RetainDNSAnswers = true
	plan.Result, err = result.New(result.Binding{TaskID: plan.TaskID, Output: out, MaxAttempts: 1, MaxDurationMS: 30000, Implementation: plan.Implementation, ContextDecoder: source.Decoder(), Path: snap.Path(), AccessGeneration: snap.Generation()})
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

// dohServer answers one RFC 8484 POST with a NOERROR response carrying one
// synthetic A answer.
func dohServer(t *testing.T) (*loopback.TestCA, loopback.HTTPConfig) {
	t.Helper()
	ca, err := loopback.NewTestCA()
	if err != nil {
		t.Fatal(err)
	}
	cert, err := ca.Issue(dohResolver, false)
	if err != nil {
		t.Fatal(err)
	}
	return ca, loopback.HTTPConfig{Certificate: cert, DoH: func(q []byte) []byte {
		r := loopback.DNSResponse(q, 0, []loopback.DNSRecord{{Type: "A", Value: "9.9.9.9"}})
		head := "HTTP/1.1 200 OK\r\nContent-Type: application/dns-message\r\nContent-Length: " + strconv.Itoa(len(r)) + "\r\n\r\n"
		return append([]byte(head), r...)
	}}
}

// HTTP request and response framing and the DNS messages are attempt bytes
// on both paths; TLS handshake and record bytes and CONNECT are not. The
// metered usage equals the plaintext the server read and wrote. A sent or
// received budget equal to the measured request or response completes; one
// byte short ends the attempt as a stricter common limit, before the request
// is complete or while the response is read (REQ-LIMIT-0034).
// requirement-ref: REQ-LIMIT-0034
// requirement-ref: REQ-DNS-03-0005
// requirement-ref: REQ-DNS-03-0023
func TestDns03PlaintextAccounting(t *testing.T) {
	ca, cfg := dohServer(t)
	request, response := 0, 0 // measured on the direct path, then used as the exact bounds
	for _, tc := range []struct {
		name        string
		proxy       bool
		bound, recv int // 0: 4096; 1: the measured request or response; 2: one byte short
	}{{"direct", false, 0, 0}, {"proxy", true, 0, 0}, {"sent equal", false, 1, 0}, {"sent below", false, 2, 0}, {"received equal", false, 0, 1}, {"received below", false, 0, 2}} {
		t.Run(tc.name, func(t *testing.T) {
			srv, err := loopback.ListenHTTP(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer srv.Close()
			var dial DialerSource = loopbackDial{srv.Address()}
			path := "direct"
			if tc.proxy {
				p, err := loopback.ListenProxy(loopback.ProxyConfig{Resolve: map[string]string{"8.8.8.8:443": srv.Address()}})
				if err != nil {
					t.Fatal(err)
				}
				defer p.Close()
				if dial, err = NewProxyDialers("127.0.0.1", p.Port()); err != nil {
					t.Fatal(err)
				}
				path = "proxy"
			}
			sent, received := 4096, 4096
			switch tc.bound {
			case 1:
				sent = request
			case 2:
				sent = request - 1
			}
			switch tc.recv {
			case 1:
				received = response
			case 2:
				received = response - 1
			}
			clock := testkit.NewWaitClock(time.Now())
			params := dns03Params(t, dohBootstrap)
			plan := dns03Plan(t, tc.proxy, task.AttemptLimits{Connections: 1, BytesSent: sent, BytesReceived: received}, params, dial, dns03Gate(clock, params, path, nil))
			engine := &Engine{clock: clock}
			attempt, usage, value, err := engine.attempt(context.Background(), plan, &dns03.Runner{Roots: ca.Pool()}, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.Wait(ctx, func(c loopback.HTTPCounts) bool { return c.Accepted == 1 && c.Open == 0 }); err != nil {
				t.Fatal(err)
			}
			c := srv.Counts()
			if tc.bound == 2 || tc.recv == 2 {
				// A short sent budget stops before the request is complete; a
				// short received budget after the request, while reading.
				if attempt.Phase() != Incomplete || attempt.Reason() != task.ResourceLimit || value.View().Measurement != nil || len(c.Requests) != tc.recv/2 {
					t.Fatalf("phase %v reason %q requests %d", attempt.Phase(), attempt.Reason(), len(c.Requests))
				}
				return
			}
			v := value.View()
			if attempt.Phase() != Completed || v.Measurement == nil || v.Measurement.Outcome != "response" || len(v.Measurement.Observations) != 4 {
				t.Fatalf("phase %v reason %q", attempt.Phase(), attempt.Reason())
			}
			if len(c.Requests) != 1 || len(c.Requests[0].Body) != dohQueryBytes || c.In <= dohQueryBytes || usage != (measurement.Usage{Connections: 1, BytesSent: c.In, BytesReceived: c.Out}) {
				t.Fatalf("usage %+v server %+v", usage, c)
			}
			if c.WireIn <= c.In || c.WireOut <= c.Out {
				t.Fatal("TLS bytes not below the counted boundary")
			}
			if request == 0 {
				request, response = c.In, c.Out
			} else if c.In != request || c.Out != response {
				t.Fatalf("request %d and response %d on %s, want %d and %d", c.In, c.Out, tc.name, request, response)
			}
		})
	}
}

// dns03ConnectedDial interrupts after a successful connect, before the
// ClientHello is sent.
type dns03ConnectedDial struct{ *dns04PointDial }

func (d dns03ConnectedDial) Dialer(int) (measurement.Dialer, error) { return d, nil }
func (d dns03ConnectedDial) DialContext(ctx context.Context, network, address string) (measurement.Conn, error) {
	c, err := d.dns04PointDial.DialContext(ctx, network, address)
	if err == nil {
		d.hit()
	}
	return c, err
}

// M-DNS-03-TERMINATION at every Procedure point under the GO-SPEC-ISSUE-004
// table, with and without a bootstrap address: before the test starts the
// attempt is not_run; during bootstrap, after bootstrap before TCP, during
// TCP connect, after TCP before the ClientHello, during TLS establishment,
// after the POST is read and while the response body is read it is
// incomplete with exactly the first reason, no measurement, no further dial
// or lookup and every connection closed; while closing the session after the
// outcome is recorded it stays completed.
// requirement-ref: REQ-DNS-03-0028
// requirement-ref: REQ-RESULT-0051
// requirement-ref: REQ-RESULT-0052
func TestDns03TerminationPoints(t *testing.T) {
	ca, base := dohServer(t)
	const partial = "HTTP/1.1 200 OK\r\nContent-Type: application/dns-message\r\nContent-Length: 100\r\n\r\na"
	causes := []struct {
		cause  error
		reason task.ExecutionReason
	}{{context.Canceled, task.CancelledByHost}, {errNetworkChanged, task.NetworkChanged}, {errDeadline, task.AttemptDeadlineExceeded}, {pathError{task.PlatformSuspended}, task.PlatformSuspended}, {errProxyUnavailable, task.ProxyUnavailable}, {errProxyConnect, task.ProxyConnectFailed}, {ErrProxyReconfigured, task.ProxyReconfigured}}
	type point struct {
		stage, input   string
		dials, lookups int
	}
	var points []point
	for _, stage := range []string{"before start", "bootstrap", "after bootstrap", "connect", "before tls", "tls", "http", "body", "close"} {
		for _, input := range []string{dohBootstrap, dohSystem} {
			system := input == dohSystem
			if !system && (stage == "bootstrap" || stage == "after bootstrap") {
				continue // the bootstrap observation is omitted
			}
			pt := point{stage: stage, input: input, dials: 1}
			if stage == "before start" || stage == "bootstrap" || stage == "after bootstrap" {
				pt.dials = 0
			}
			if system && stage != "before start" {
				pt.lookups = 1
			}
			points = append(points, pt)
		}
	}
	for _, tc := range causes {
		for _, pt := range points {
			variant := "bootstrap_ip"
			if pt.input == dohSystem {
				variant = "system"
			}
			t.Run(string(tc.reason)+"/"+pt.stage+"/"+variant, func(t *testing.T) {
				clock := testkit.NewWaitClock(time.Now())
				ctx, cancel := context.WithCancelCause(context.Background())
				defer cancel(nil)
				var once sync.Once
				interrupt := func() { once.Do(func() { cancel(tc.cause); cancel(errNetworkChanged) }) }
				var address string
				var tlsSrv *loopback.TLSServer
				var httpSrv *loopback.HTTPServer
				if pt.stage == "tls" {
					s, err := loopback.ListenTLS(loopback.TLSConfig{Mode: loopback.TLSStall})
					if err != nil {
						t.Fatal(err)
					}
					defer s.Close()
					tlsSrv, address = s, s.Address()
				} else {
					cfg := base
					switch pt.stage {
					case "http":
						cfg.DoH = func([]byte) []byte { return nil }
						cfg.Stall = true
					case "body":
						cfg.DoH = func([]byte) []byte { return []byte(partial) }
						cfg.Stall = true
					}
					s, err := loopback.ListenHTTP(cfg)
					if err != nil {
						t.Fatal(err)
					}
					defer s.Close()
					httpSrv, address = s, s.Address()
				}
				dial := &dns04PointDial{address: address, interrupt: pt.stage == "connect", hit: interrupt}
				if pt.stage == "close" {
					dial.onClose = interrupt
				}
				var source DialerSource = dial
				if pt.stage == "before tls" {
					source = dns03ConnectedDial{dial}
				}
				lookups := &atomic.Int32{}
				resolver := http01PointResolver{lookups: lookups, interruptAt: -1, hit: interrupt}
				if pt.stage == "bootstrap" || pt.stage == "after bootstrap" {
					resolver.interruptAt, resolver.afterAnswer = 0, pt.stage == "after bootstrap"
				}
				params := dns03Params(t, pt.input)
				plan := dns03Plan(t, false, task.AttemptLimits{Connections: 1, BytesSent: 4096, BytesReceived: 4096}, params, source, dns03Gate(clock, params, "direct", resolver))
				if pt.stage == "before start" {
					plan.Context = tcpInterruptedCapture{plan.Context.(*networkcontext.Source), interrupt}
				}
				done := make(chan struct{})
				if pt.stage == "tls" || pt.stage == "http" || pt.stage == "body" {
					go func() {
						defer close(done)
						wait, stop := context.WithTimeout(ctx, 5*time.Second)
						defer stop()
						var e error
						if pt.stage == "tls" {
							e = tlsSrv.Wait(wait, func(c loopback.TLSCounts) bool { return c.In > 0 })
						} else if pt.stage == "http" {
							e = httpSrv.Wait(wait, func(c loopback.HTTPCounts) bool { return len(c.Requests) == 1 })
						} else {
							e = httpSrv.Wait(wait, func(c loopback.HTTPCounts) bool { return len(c.Requests) == 1 && c.Out >= len(partial) })
						}
						if e == nil {
							interrupt()
						} else {
							t.Error("stage was not reached", e)
							cancel(context.Canceled)
						}
					}()
				} else {
					close(done)
				}
				engine := &Engine{clock: clock}
				a, _, v, err := engine.attempt(ctx, plan, &dns03.Runner{Roots: ca.Pool()}, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
				if err != nil {
					t.Fatal(err)
				}
				<-done
				view := v.View()
				switch pt.stage {
				case "before start":
					if a.Phase() != NotRun || a.Reason() != tc.reason || view.Measurement != nil || view.NetworkContext != nil {
						t.Fatalf("phase %v reason %s", a.Phase(), a.Reason())
					}
				case "close":
					if a.Phase() != Completed || a.Reason() != "" || view.Measurement == nil || view.Measurement.Outcome != "response" {
						t.Fatalf("cleanup replaced the recorded outcome: phase %v reason %s", a.Phase(), a.Reason())
					}
				default:
					if a.Phase() != Incomplete || a.Reason() != tc.reason || view.Measurement != nil {
						t.Fatalf("phase %v reason %s measurement %v", a.Phase(), a.Reason(), view.Measurement != nil)
					}
				}
				for deadline := time.Now().Add(5 * time.Second); ; time.Sleep(time.Millisecond) {
					n, closed := dial.state()
					if n != pt.dials {
						t.Fatalf("dials %d want %d", n, pt.dials)
					}
					if closed {
						break
					}
					if time.Now().After(deadline) {
						t.Fatal("connection left open")
					}
				}
				if int(lookups.Load()) != pt.lookups {
					t.Fatalf("lookups %d want %d", lookups.Load(), pt.lookups)
				}
				if httpSrv != nil {
					c := httpSrv.Counts()
					if want := map[bool]int{true: 1, false: 0}[pt.stage == "http" || pt.stage == "body" || pt.stage == "close"]; c.Accepted > 1 || len(c.Requests) != want {
						t.Fatalf("second connection or request after interruption: %+v", c)
					}
					if pt.stage == "before tls" && len(c.Hellos) != 0 {
						t.Fatal("ClientHello sent after the interruption")
					}
				}
				if tlsSrv != nil && tlsSrv.Counts().Accepted > 1 {
					t.Fatal("second connection after interruption")
				}
				if clock.Timers() != 0 {
					t.Fatal("timer leaked")
				}
			})
		}
	}
}
