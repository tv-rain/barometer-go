package execution

import (
	"context"
	"net/netip"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/measurement/dns04"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/networkcontext"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

const (
	dotResolver    = "dot.resolver.example"
	dotBootstrap   = `{"resolver_name":"dot.resolver.example","port":853,"hostname":"query.example","record_type":"A","bootstrap_ip":"8.8.8.8"}`
	dotSystem      = `{"resolver_name":"dot.resolver.example","port":853,"hostname":"query.example","record_type":"A"}`
	dotQueryBytes  = 31 // header, query.example and question tail
	dotFramedQuery = 2 + dotQueryBytes
)

var dns04Test = task.TestVersion{ID: "DNS-04", Version: "0.1.0"}

func dns04Params(t *testing.T, input string) dns04.Parameters {
	t.Helper()
	codec := dns04.Codec{}
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
	return p.(dns04.Parameters)
}

// dns04Gate registers the query name and the resolver profile in the shape
// of the production target rule, on the given path.
func dns04Gate(clock interface{ Now() time.Time }, p dns04.Parameters, path string, resolver SystemResolver) *Targets {
	name, port, bootstrap := p.Resolver()
	profile := task.TargetTuple{Hostname: name, ResolverName: name, TLSServerName: name, Port: port, Scheme: "tls"}
	if bootstrap.IsValid() {
		profile.BootstrapAddresses = []netip.Addr{bootstrap}
	}
	paths := task.Paths{Direct: path == "direct", Proxy: path == "proxy"}
	expires := task.TimestampFromTime(clock.Now().Add(time.Hour))
	regs := []task.TargetRegistration{
		{ID: "dot-query", Test: dns04Test, Role: task.MeasurementTarget, Tuple: task.TargetTuple{Hostname: p.Name()}, Paths: paths, ExpiresAt: expires},
		{ID: "dot-resolver", Test: dns04Test, Role: task.ResolverTarget, Tuple: profile, Paths: paths, ExpiresAt: expires},
	}
	return NewTargets(dns04Test, path, regs, clock, nil).WithResolver(resolver)
}

func dns04Plan(t *testing.T, proxy bool, limits task.AttemptLimits, p dns04.Parameters, dial DialerSource, gate TargetCheck) Plan {
	t.Helper()
	plan := tls01Plan(t, proxy, 30*time.Second, limits, dial, gate)
	plan.TaskID, plan.Test, plan.Parameters = "dns04-plan", dns04Test, p
	source := plan.Context.(*networkcontext.Source)
	snap, err := source.Capture(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	out := result.Output{Test: model.Test{ID: "DNS-04", Version: "0.1.0", ResultVersion: "0.1.0"}, QueryType: p.QueryType(), Bootstrap: p.Bootstrap()}
	out.Limits.RetainDNSAnswers = true
	plan.Result, err = result.New(result.Binding{TaskID: plan.TaskID, Output: out, MaxAttempts: 1, MaxDurationMS: 30000, Implementation: plan.Implementation, ContextDecoder: source.Decoder(), Path: snap.Path(), AccessGeneration: snap.Generation()})
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

// dotAnswer frames one NOERROR response with a single synthetic A answer and
// records the sizes the server read and wrote.
func dotAnswer(query, response *atomic.Int32) func([]byte) []byte {
	return func(q []byte) []byte {
		r := loopback.DNSResponse(q, 0, []loopback.DNSRecord{{Type: "A", Value: "9.9.9.9"}})
		query.Store(int32(len(q)))
		response.Store(int32(len(r)))
		return append([]byte{byte(len(r) >> 8), byte(len(r))}, r...)
	}
}

func dotCertificate(t *testing.T) (*loopback.TestCA, loopback.TLSConfig) {
	t.Helper()
	ca, err := loopback.NewTestCA()
	if err != nil {
		t.Fatal(err)
	}
	cert, err := ca.Issue(dotResolver, false)
	if err != nil {
		t.Fatal(err)
	}
	return ca, loopback.TLSConfig{Certificate: cert}
}

// The DNS message and its two-byte length prefix are attempt bytes on both
// paths; TLS handshake and record bytes are not. A sent budget one byte short
// of the framed query ends the attempt before the message leaves, as a
// stricter common limit (REQ-LIMIT-0034).
// requirement-ref: REQ-LIMIT-0034
// requirement-ref: REQ-DNS-04-0005
// requirement-ref: REQ-DNS-04-0023
func TestDns04PlaintextAccounting(t *testing.T) {
	ca, cfg := dotCertificate(t)
	for _, tc := range []struct {
		name  string
		proxy bool
		sent  int
	}{{"direct", false, 4096}, {"proxy", true, 4096}, {"sent equal", false, dotFramedQuery}, {"sent below", false, dotFramedQuery - 1}} {
		t.Run(tc.name, func(t *testing.T) {
			var query, response atomic.Int32
			cfg.DoT = dotAnswer(&query, &response)
			srv, err := loopback.ListenTLS(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer srv.Close()
			var dial DialerSource = loopbackDial{srv.Address()}
			path := "direct"
			if tc.proxy {
				p, err := loopback.ListenProxy(loopback.ProxyConfig{Resolve: map[string]string{"8.8.8.8:853": srv.Address()}})
				if err != nil {
					t.Fatal(err)
				}
				defer p.Close()
				if dial, err = NewProxyDialers("127.0.0.1", p.Port()); err != nil {
					t.Fatal(err)
				}
				path = "proxy"
			}
			clock := testkit.NewWaitClock(time.Now())
			params := dns04Params(t, dotBootstrap)
			plan := dns04Plan(t, tc.proxy, task.AttemptLimits{Connections: 1, BytesSent: tc.sent, BytesReceived: 4096}, params, dial, dns04Gate(clock, params, path, nil))
			engine := &Engine{clock: clock}
			attempt, usage, value, err := engine.attempt(context.Background(), plan, &dns04.Runner{Roots: ca.Pool()}, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.WaitIdle(ctx); err != nil {
				t.Fatal(err)
			}
			if tc.sent < dotFramedQuery {
				if attempt.Phase() != Incomplete || attempt.Reason() != task.ResourceLimit || value.View().Measurement != nil || srv.Counts().Plaintext != 0 {
					t.Fatalf("phase %v reason %q plaintext %d", attempt.Phase(), attempt.Reason(), srv.Counts().Plaintext)
				}
				return
			}
			v := value.View()
			if attempt.Phase() != Completed || v.Measurement == nil || v.Measurement.Outcome != "response" || len(v.Measurement.Observations) != 3 {
				t.Fatalf("phase %v reason %q", attempt.Phase(), attempt.Reason())
			}
			if query.Load() != dotQueryBytes || usage != (measurement.Usage{Connections: 1, BytesSent: 2 + dotQueryBytes, BytesReceived: 2 + int(response.Load())}) {
				t.Fatalf("usage %+v query %d response %d", usage, query.Load(), response.Load())
			}
			c := srv.Counts()
			if c.Accepted != 1 || c.Plaintext != 2+dotQueryBytes || len(c.Hellos) != 1 || c.Hellos[0].ServerName != dotResolver || len(c.Hellos[0].ALPN) != 0 {
				t.Fatalf("server %+v", c)
			}
		})
	}
}

// dns04PointConn reports its close so the closing point can interrupt after
// the outcome is recorded.
type dns04PointConn struct {
	measurement.Conn
	closed atomic.Bool
	hit    func()
}

func (c *dns04PointConn) Close() error {
	if c.closed.CompareAndSwap(false, true) && c.hit != nil {
		c.hit()
	}
	return c.Conn.Close()
}

type dns04PointDial struct {
	mu        sync.Mutex
	dials     int
	conns     []*dns04PointConn
	address   string
	interrupt bool
	onClose   func()
	hit       func()
}

func (d *dns04PointDial) Dialer(int) (measurement.Dialer, error) { return d, nil }
func (d *dns04PointDial) DialContext(ctx context.Context, network, _ string) (measurement.Conn, error) {
	d.mu.Lock()
	d.dials++
	d.mu.Unlock()
	if d.interrupt {
		d.hit()
		return nil, ctx.Err()
	}
	c, err := loopbackDial{d.address}.DialContext(ctx, "tcp", "")
	if err != nil {
		return nil, err
	}
	pc := &dns04PointConn{Conn: c, hit: d.onClose}
	d.mu.Lock()
	d.conns = append(d.conns, pc)
	d.mu.Unlock()
	return pc, nil
}
func (d *dns04PointDial) state() (int, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, c := range d.conns {
		if !c.closed.Load() {
			return d.dials, false
		}
	}
	return d.dials, true
}

// M-DNS-04-TERMINATION at every Procedure point under the GO-SPEC-ISSUE-004
// table: before the test starts the attempt is not_run; during bootstrap,
// TCP connect, TLS establishment and after the length-prefixed query is
// written it is incomplete with exactly the first reason, no measurement,
// no further dial or lookup and every connection closed; while closing the
// session after the outcome is recorded it stays completed.
// requirement-ref: REQ-DNS-04-0028
// requirement-ref: REQ-RESULT-0051
// requirement-ref: REQ-RESULT-0052
func TestDns04TerminationPoints(t *testing.T) {
	ca, base := dotCertificate(t)
	causes := []struct {
		cause  error
		reason task.ExecutionReason
	}{{context.Canceled, task.CancelledByHost}, {errNetworkChanged, task.NetworkChanged}, {errDeadline, task.AttemptDeadlineExceeded}, {pathError{task.PlatformSuspended}, task.PlatformSuspended}, {errProxyUnavailable, task.ProxyUnavailable}, {errProxyConnect, task.ProxyConnectFailed}, {ErrProxyReconfigured, task.ProxyReconfigured}}
	points := []struct {
		stage, input   string
		dials, lookups int
	}{
		{"before start", dotBootstrap, 0, 0},
		{"bootstrap", dotSystem, 0, 1},
		{"connect", dotBootstrap, 1, 0},
		{"tls", dotSystem, 1, 1},
		{"dns", dotBootstrap, 1, 0},
		{"close", dotBootstrap, 1, 0},
	}
	for _, tc := range causes {
		for _, pt := range points {
			t.Run(string(tc.reason)+"/"+pt.stage, func(t *testing.T) {
				clock := testkit.NewWaitClock(time.Now())
				ctx, cancel := context.WithCancelCause(context.Background())
				defer cancel(nil)
				var once sync.Once
				interrupt := func() { once.Do(func() { cancel(tc.cause); cancel(errNetworkChanged) }) }
				cfg := base
				if pt.stage == "tls" {
					cfg = loopback.TLSConfig{Mode: loopback.TLSStall}
				}
				cfg.DoT = func(q []byte) []byte {
					if pt.stage == "dns" {
						return nil
					}
					r := loopback.DNSResponse(q, 0, nil)
					return append([]byte{byte(len(r) >> 8), byte(len(r))}, r...)
				}
				srv, err := loopback.ListenTLS(cfg)
				if err != nil {
					t.Fatal(err)
				}
				defer srv.Close()
				dial := &dns04PointDial{address: srv.Address(), interrupt: pt.stage == "connect", hit: interrupt}
				if pt.stage == "close" {
					dial.onClose = interrupt
				}
				lookups := &atomic.Int32{}
				resolver := http01PointResolver{lookups: lookups, interruptAt: -1, hit: interrupt}
				if pt.stage == "bootstrap" {
					resolver.interruptAt = 0
				}
				params := dns04Params(t, pt.input)
				plan := dns04Plan(t, false, task.AttemptLimits{Connections: 1, BytesSent: 4096, BytesReceived: 4096}, params, dial, dns04Gate(clock, params, "direct", resolver))
				source := plan.Context.(*networkcontext.Source)
				if pt.stage == "before start" {
					plan.Context = tcpInterruptedCapture{source, interrupt}
				}
				done := make(chan struct{})
				if pt.stage == "tls" || pt.stage == "dns" {
					go func() {
						defer close(done)
						wait, stop := context.WithTimeout(ctx, 5*time.Second)
						defer stop()
						var e error
						if pt.stage == "tls" {
							e = srv.Wait(wait, func(c loopback.TLSCounts) bool { return c.In > 0 })
						} else {
							e = srv.Wait(wait, func(c loopback.TLSCounts) bool { return c.Plaintext == 2+dotQueryBytes })
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
				a, _, v, err := engine.attempt(ctx, plan, &dns04.Runner{Roots: ca.Pool()}, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
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
				if c := srv.Counts(); c.Accepted > 1 || pt.stage != "dns" && pt.stage != "close" && c.Plaintext != 0 {
					t.Fatalf("second connection or DNS traffic after interruption: %+v", c)
				}
				if clock.Timers() != 0 {
					t.Fatal("timer leaked")
				}
			})
		}
	}
}
