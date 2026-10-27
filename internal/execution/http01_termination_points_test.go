package execution

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/measurement/http01"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/networkcontext"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

// http01PointDial routes the i-th dial (0-based) and records every opened
// connection so a test can show no further connection and no open stream.
type http01PointDial struct {
	mu    sync.Mutex
	dials int
	conns []*http01PointConn
	route func(int) (string, bool)
	hit   func()
}

func (d *http01PointDial) Dialer(int) (measurement.Dialer, error) { return d, nil }
func (d *http01PointDial) DialContext(ctx context.Context, _, _ string) (measurement.Conn, error) {
	d.mu.Lock()
	i := d.dials
	d.dials++
	d.mu.Unlock()
	address, interrupt := d.route(i)
	if interrupt {
		d.hit()
		return nil, ctx.Err()
	}
	var nd net.Dialer
	c, e := nd.DialContext(ctx, "tcp", address)
	if e != nil {
		return nil, e
	}
	pc := &http01PointConn{Conn: c}
	d.mu.Lock()
	d.conns = append(d.conns, pc)
	d.mu.Unlock()
	return pc, nil
}
func (d *http01PointDial) state() (int, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, c := range d.conns {
		if !c.closed.Load() {
			return d.dials, false
		}
	}
	return d.dials, true
}

type http01PointConn struct {
	net.Conn
	closed atomic.Bool
}

func (c *http01PointConn) Close() error { c.closed.Store(true); return c.Conn.Close() }

// http01PointResolver interrupts during the i-th lookup (0-based) when
// interruptAt matches. With afterAnswer it first completes the lookup, so the
// interruption lands after DNS and before the gate hands out a TCP address.
type http01PointResolver struct {
	lookups     *atomic.Int32
	interruptAt int
	afterAnswer bool
	hit         func()
}

func (r http01PointResolver) LookupNetIP(ctx context.Context, _, _ string) ([]netip.Addr, error) {
	if int(r.lookups.Add(1)-1) == r.interruptAt {
		r.hit()
		if !r.afterAnswer {
			return nil, ctx.Err()
		}
	}
	return []netip.Addr{netip.MustParseAddr("8.8.8.8")}, nil
}

// http01PointClock is the destination gate's trusted clock. The HTTP-01 staged
// path requires the concrete production gate, so the redirect point is reached
// through its clock: once armed (after the first connection is dialled), the
// gate's next reading is the expiry check at the start of next-URL validation.
type http01PointClock struct {
	inner interface{ Now() time.Time }
	armed *atomic.Bool
	hit   func()
}

func (c http01PointClock) Now() time.Time {
	if c.armed.CompareAndSwap(true, false) {
		c.hit()
	}
	return c.inner.Now()
}

// Procedure points not covered by TestHttp01AllTerminationBoundaries: after
// DNS before TCP; a redirect response with following disabled and enabled,
// interrupted before next-URL validation; and each redirect hop's DNS, TCP,
// TLS and HTTP stage for max_redirects 1 and 2. Every common cause yields
// incomplete with exactly that reason, no measurement, no further dial or
// lookup, and every opened connection closed.
// requirement-ref: REQ-HTTP-01-0038
// requirement-ref: REQ-RESULT-0051
// requirement-ref: REQ-RESULT-0052
func TestHttp01RedirectAndHopTerminationPoints(t *testing.T) {
	ca, e := loopback.NewTestCA()
	if e != nil {
		t.Fatal(e)
	}
	cert, e := ca.Issue("service.example", false)
	if e != nil {
		t.Fatal(e)
	}
	type point struct {
		name      string
		max, hop  int
		stage     string
		follow    bool
		wantDials int
		lookups   int
	}
	points := []point{
		{name: "after-dns", stage: "after-dns", wantDials: 0, lookups: 1},
		{name: "redirect-disabled", stage: "redirect", wantDials: 1, lookups: 1},
		{name: "redirect-enabled", stage: "redirect", follow: true, max: 1, wantDials: 1, lookups: 1},
	}
	for _, max := range []int{1, 2} {
		for hop := 1; hop <= max; hop++ {
			for _, stage := range []string{"dns", "connect", "tls", "headers"} {
				dials := hop + 1
				if stage == "dns" {
					dials = hop
				}
				points = append(points, point{name: fmt.Sprintf("max%d/hop%d/%s", max, hop, stage), max: max, hop: hop, stage: stage, follow: true, wantDials: dials, lookups: hop + 1})
			}
		}
	}
	causes := []struct {
		cause  error
		reason task.ExecutionReason
	}{{context.Canceled, task.CancelledByHost}, {errNetworkChanged, task.NetworkChanged}, {errDeadline, task.AttemptDeadlineExceeded}, {errBudget, task.ResourceLimit}, {pathError{task.PlatformSuspended}, task.PlatformSuspended}, {errProxyUnavailable, task.ProxyUnavailable}, {errProxyConnect, task.ProxyConnectFailed}, {ErrProxyReconfigured, task.ProxyReconfigured}}
	const redirect = "HTTP/1.1 302 Found\r\nLocation: https://service.example/next\r\nContent-Length: 0\r\n\r\n"
	for _, tc := range causes {
		for _, pt := range points {
			t.Run(string(tc.reason)+"/"+pt.name, func(t *testing.T) {
				clock := testkit.NewWaitClock(time.Now())
				ctx, cancel := context.WithCancelCause(context.Background())
				defer cancel(nil)
				var once sync.Once
				interrupt := func() { once.Do(func() { cancel(tc.cause) }) }
				redirects, err := loopback.ListenHTTP(loopback.HTTPConfig{Certificate: cert, Raw: []byte(redirect)})
				if err != nil {
					t.Fatal(err)
				}
				defer redirects.Close()
				stalled, err := loopback.ListenHTTP(loopback.HTTPConfig{Certificate: cert, Stall: true})
				if err != nil {
					t.Fatal(err)
				}
				defer stalled.Close()
				handshake, err := loopback.ListenTLS(loopback.TLSConfig{Mode: loopback.TLSStall})
				if err != nil {
					t.Fatal(err)
				}
				defer handshake.Close()
				armed := &atomic.Bool{}
				dial := &http01PointDial{hit: interrupt, route: func(i int) (string, bool) {
					if i == 0 && pt.stage == "redirect" {
						armed.Store(true)
					}
					if i < pt.hop || pt.hop == 0 {
						return redirects.Address(), false
					}
					switch pt.stage {
					case "connect":
						return "", true
					case "tls":
						return handshake.Address(), false
					}
					return stalled.Address(), false
				}}
				codec := http01.Codec{}
				def := codec.Definition()
				lim := jsondecode.Limits{MaxBytes: 4096, MaxDepth: 16, MaxNodes: 64}
				in, err := jsondecode.Decode([]byte(`{"url":"https://service.example/check","expected_status":200}`), def.Input, lim)
				if err != nil {
					t.Fatal(err)
				}
				options := `{}`
				if pt.follow {
					options = fmt.Sprintf(`{"follow_redirects":true,"max_redirects":%d}`, pt.max)
				}
				opt, err := jsondecode.Decode([]byte(options), def.Options, lim)
				if err != nil {
					t.Fatal(err)
				}
				params, err := codec.Normalize(in, opt)
				if err != nil {
					t.Fatal(err)
				}
				hp := params.(http01.Parameters)
				test := task.TestVersion{ID: "HTTP-01", Version: "0.1.0"}
				expires := task.TimestampFromTime(clock.Now().Add(time.Hour))
				next, ok := hp.RedirectURL("https://service.example/next")
				if !ok {
					t.Fatal("fixture")
				}
				next.ExpectedStatus = 200
				regs := []task.TargetRegistration{
					{ID: "http-a", Test: test, Role: task.MeasurementTarget, Tuple: hp.Target(), Paths: task.Paths{Direct: true}, ExpiresAt: expires, RedirectIDs: []string{"r1"}},
					{ID: "r1", Test: test, Role: task.RedirectTarget, Tuple: next, Paths: task.Paths{Direct: true}, ExpiresAt: expires, RedirectIDs: []string{"r2"}},
					{ID: "r2", Test: test, Role: task.RedirectTarget, Tuple: next, Paths: task.Paths{Direct: true}, ExpiresAt: expires},
				}
				lookups := &atomic.Int32{}
				resolver := http01PointResolver{lookups: lookups, interruptAt: -1, hit: interrupt}
				if pt.stage == "dns" {
					resolver.interruptAt = pt.hop
				}
				if pt.stage == "after-dns" {
					resolver.interruptAt, resolver.afterAnswer = 0, true
				}
				gate := NewTargets(test, "direct", regs, http01PointClock{clock, armed, interrupt}, nil).WithResolver(resolver)
				limits := task.AttemptLimits{Connections: 4, BytesSent: 32768, BytesReceived: 131072}
				p := tls01Plan(t, false, 30*time.Second, limits, dial, gate)
				p.TaskID = "http-points"
				p.Test = test
				p.Parameters = params
				source := p.Context.(*networkcontext.Source)
				snap, err := source.Capture(ctx)
				if err != nil {
					t.Fatal(err)
				}
				p.Result, err = result.New(result.Binding{TaskID: p.TaskID, Output: result.Output{Test: model.Test{ID: "HTTP-01", Version: "0.1.0", ResultVersion: "0.1.0"}, Limits: hp.Limits(), HTTPS: true}, MaxAttempts: 1, MaxDurationMS: 30000, Implementation: p.Implementation, ContextDecoder: source.Decoder(), Path: snap.Path(), AccessGeneration: snap.Generation()})
				if err != nil {
					t.Fatal(err)
				}
				done := make(chan struct{})
				if pt.stage == "tls" || pt.stage == "headers" {
					go func() {
						defer close(done)
						wait, stop := context.WithTimeout(ctx, 5*time.Second)
						defer stop()
						var e error
						if pt.stage == "tls" {
							e = handshake.Wait(wait, func(c loopback.TLSCounts) bool { return c.In > 0 })
						} else {
							e = stalled.Wait(wait, func(c loopback.HTTPCounts) bool { return len(c.Requests) > 0 })
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
				a, _, v, err := engine.attempt(ctx, p, &http01.Runner{Roots: ca.Pool()}, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
				if err != nil {
					t.Fatal(err)
				}
				<-done
				view := v.View()
				if a.Phase() != Incomplete || a.Reason() != tc.reason || view.Measurement != nil {
					t.Fatalf("phase %v reason %s measurement %v", a.Phase(), a.Reason(), view.Measurement != nil)
				}
				for deadline := time.Now().Add(5 * time.Second); ; time.Sleep(time.Millisecond) {
					n, closed := dial.state()
					if n != pt.wantDials {
						t.Fatalf("dials %d want %d", n, pt.wantDials)
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
				requests := pt.hop
				if pt.hop == 0 {
					requests = pt.wantDials
				}
				if got := len(redirects.Counts().Requests); got != requests {
					t.Fatalf("redirect requests %d want %d", got, requests)
				}
			})
		}
	}
}
