package execution

import (
	"context"
	"net"
	"net/netip"
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

type httpBoundaryDial struct {
	address, stage string
	interrupt      func()
}

func (d httpBoundaryDial) Dialer(int) (measurement.Dialer, error) { return d, nil }
func (d httpBoundaryDial) DialContext(ctx context.Context, network, address string) (measurement.Conn, error) {
	if d.stage == "connect" {
		d.interrupt()
		return nil, ctx.Err()
	}
	var nd net.Dialer
	c, e := nd.DialContext(ctx, "tcp", d.address)
	if e != nil {
		return nil, e
	}
	return &httpBoundaryConn{Conn: c, closeHook: func() {
		if d.stage == "cleanup" {
			d.interrupt()
		}
	}}, nil
}

type httpBoundaryConn struct {
	net.Conn
	closeHook func()
}

func (c *httpBoundaryConn) Close() error { c.closeHook(); return c.Conn.Close() }

type httpBoundaryResolver struct {
	stage     string
	interrupt func()
}

func (r httpBoundaryResolver) LookupNetIP(ctx context.Context, _, _ string) ([]netip.Addr, error) {
	if r.stage == "dns" {
		r.interrupt()
		return nil, ctx.Err()
	}
	return []netip.Addr{netip.MustParseAddr("8.8.8.8")}, nil
}

// Every common reason is injected against the production HTTP-01 runner and
// engine. Proxy/suspension causes here are synthetic state events; actual
// CONNECT and lifecycle-adapter validation remain separately bounded evidence.
// requirement-ref: REQ-HTTP-01-0038
// requirement-ref: REQ-RESULT-0051
// requirement-ref: REQ-RESULT-0052
// requirement-ref: REQ-RESULT-0053
func TestHttp01AllTerminationBoundaries(t *testing.T) {
	ca, e := loopback.NewTestCA()
	if e != nil {
		t.Fatal(e)
	}
	cert, e := ca.Issue("service.example", false)
	if e != nil {
		t.Fatal(e)
	}
	causes := []struct {
		cause  error
		reason task.ExecutionReason
	}{{context.Canceled, task.CancelledByHost}, {errNetworkChanged, task.NetworkChanged}, {errDeadline, task.AttemptDeadlineExceeded}, {errBudget, task.ResourceLimit}, {pathError{task.PlatformSuspended}, task.PlatformSuspended}, {errProxyUnavailable, task.ProxyUnavailable}, {errProxyConnect, task.ProxyConnectFailed}, {ErrProxyReconfigured, task.ProxyReconfigured}}
	for _, tc := range causes {
		for _, stage := range []string{"before start", "dns", "connect", "tls", "headers", "body", "cleanup"} {
			t.Run(string(tc.reason)+"/"+stage, func(t *testing.T) {
				clock := testkit.NewWaitClock(time.Now())
				ctx, cancel := context.WithCancelCause(context.Background())
				defer cancel(nil)
				interrupt := func() { cancel(tc.cause); cancel(errNetworkChanged) }
				raw := "HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n"
				stall := stage == "headers" || stage == "body"
				if stage == "headers" {
					raw = ""
				}
				if stage == "body" {
					raw = "HTTP/1.1 200 OK\r\nContent-Length: 100\r\n\r\na"
				}
				srv, err := loopback.ListenHTTP(loopback.HTTPConfig{Certificate: cert, Raw: []byte(raw), Stall: stall})
				if err != nil {
					t.Fatal(err)
				}
				defer srv.Close()
				address := srv.Address()
				var tlsServer *loopback.TLSServer
				if stage == "tls" {
					tlsServer, err = loopback.ListenTLS(loopback.TLSConfig{Mode: loopback.TLSStall})
					if err != nil {
						t.Fatal(err)
					}
					defer tlsServer.Close()
					address = tlsServer.Address()
				}
				codec := http01.Codec{}
				def := codec.Definition()
				lim := jsondecode.Limits{MaxBytes: 4096, MaxDepth: 16, MaxNodes: 64}
				in, err := jsondecode.Decode([]byte(`{"url":"https://service.example/check","expected_status":200}`), def.Input, lim)
				if err != nil {
					t.Fatal(err)
				}
				opt, err := jsondecode.Decode([]byte(`{}`), def.Options, lim)
				if err != nil {
					t.Fatal(err)
				}
				params, err := codec.Normalize(in, opt)
				if err != nil {
					t.Fatal(err)
				}
				hp := params.(http01.Parameters)
				reg := task.TargetRegistration{ID: "http-a", Test: task.TestVersion{ID: "HTTP-01", Version: "0.1.0"}, Role: task.MeasurementTarget, Tuple: hp.Target(), Paths: task.Paths{Direct: true}, ExpiresAt: task.TimestampFromTime(clock.Now().Add(time.Hour))}
				gate := NewTargets(reg.Test, "direct", []task.TargetRegistration{reg}, clock, nil).WithResolver(httpBoundaryResolver{stage, interrupt})
				limits := task.AttemptLimits{Connections: 4, BytesSent: 32768, BytesReceived: 131072}
				p := tls01Plan(t, false, 30*time.Second, limits, httpBoundaryDial{address, stage, interrupt}, gate)
				p.TaskID = "http-boundaries"
				p.Test = reg.Test
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
				if stage == "before start" {
					p.Context = tcpInterruptedCapture{source, interrupt}
				}
				done := make(chan struct{})
				if stage == "tls" || stage == "headers" || stage == "body" {
					go func() {
						defer close(done)
						wait, stop := context.WithTimeout(ctx, 5*time.Second)
						defer stop()
						var e error
						if stage == "tls" {
							e = tlsServer.Wait(wait, func(c loopback.TLSCounts) bool { return c.In > 0 })
						} else {
							e = srv.Wait(wait, func(c loopback.HTTPCounts) bool { return len(c.Requests) > 0 && c.Out >= len(raw) })
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
				want := Incomplete
				if stage == "before start" {
					want = NotRun
				}
				if stage == "cleanup" {
					want = Completed
				}
				if a.Phase() != want {
					t.Fatalf("phase %v want %v", a.Phase(), want)
				}
				view := v.View()
				if want == Completed {
					if a.Reason() != "" || view.Measurement == nil || view.Measurement.Outcome != "expected_response" {
						t.Fatal("cleanup replaced outcome")
					}
				} else if a.Reason() != tc.reason || view.Measurement != nil {
					t.Fatalf("reason %s want %s or invented partial output", a.Reason(), tc.reason)
				}
				if want == NotRun && view.NetworkContext != nil {
					t.Fatal("not_run context")
				}
			})
		}
	}
}
