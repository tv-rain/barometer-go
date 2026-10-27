package execution

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/measurement/net03"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/networkcontext"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

// net03PointRunner wraps the production runner to interrupt at the runner's
// own Procedure points: during the stability check that starts the test,
// after it, before or after the platform read, and after the outcome is
// recorded.
type net03PointRunner struct {
	inner measurement.Runner
	stage string
	hit   func()
}

func (r net03PointRunner) Run(ctx context.Context, in measurement.Input) (measurement.Output, error) {
	check, read := in.CheckNetwork, in.Validation
	in.CheckNetwork = func(family string) error {
		if family == "" && r.stage == "during-stability" {
			r.hit()
		}
		err := check(family)
		if family == "" && r.stage == "stability" {
			r.hit()
		}
		return err
	}
	in.Validation = func(ctx context.Context) (string, error) {
		if r.stage == "during-read" {
			r.hit()
		}
		v, err := read(ctx)
		if r.stage == "after-read" {
			r.hit()
		}
		return v, err
	}
	out, err := r.inner.Run(ctx, in)
	if r.stage == "release" {
		r.hit()
	}
	return out, err
}

// net03AfterConnect interrupts once the TCP connection is established: before
// the tls_handshake for an https URL and before the request is sent for an
// http URL.
type net03AfterConnect struct{ *http01PointDial }

func (d net03AfterConnect) Dialer(int) (measurement.Dialer, error) { return d, nil }
func (d net03AfterConnect) DialContext(ctx context.Context, network, address string) (measurement.Conn, error) {
	c, err := d.http01PointDial.DialContext(ctx, network, address)
	if err == nil {
		d.hit()
	}
	return c, err
}

// Every NET-03 Procedure point with every common cause, at the engine and
// production-runner boundary. Before the test starts the attempt is not_run;
// from the NET-01 stability check until the outcome is recorded it is
// incomplete with exactly the first reason and no measurement; once the
// outcome is recorded, release leaves it completed (GO-SPEC-ISSUE-004). No
// second request or redirect contact follows an interruption, and every
// opened connection, timer and goroutine is released. Suspension is injected
// as an execution cause; this is no native lifecycle evidence.
// requirement-ref: REQ-NET-03-0029
// requirement-ref: REQ-RESULT-0051
// requirement-ref: REQ-RESULT-0052
// requirement-ref: REQ-RESULT-0053
func TestNet03AllTerminationBoundaries(t *testing.T) {
	baseline := runtime.NumGoroutine()
	defer func() {
		// Every runner, exchange and fixture goroutine is released.
		for deadline := time.Now().Add(5 * time.Second); runtime.NumGoroutine() > baseline; time.Sleep(time.Millisecond) {
			if time.Now().After(deadline) {
				t.Errorf("goroutines %d, baseline %d", runtime.NumGoroutine(), baseline)
				return
			}
		}
	}()
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
	}{{context.Canceled, task.CancelledByHost}, {errNetworkChanged, task.NetworkChanged}, {errDeadline, task.AttemptDeadlineExceeded}, {errBudget, task.ResourceLimit}, {pathError{task.PlatformSuspended}, task.PlatformSuspended}}
	type point struct {
		stage   string
		https   bool
		dials   int
		lookups int
	}
	var points []point
	for _, https := range []bool{true, false} {
		points = append(points, point{"before-start", https, 0, 0}, point{"during-stability", https, 0, 0}, point{"stability", https, 0, 0}, point{"during-read", https, 0, 0}, point{"after-read", https, 0, 0},
			point{"dns", https, 0, 1}, point{"after-dns", https, 0, 1}, point{"connect", https, 1, 1}, point{"after-connect", https, 1, 1}, point{"headers", https, 1, 1}, point{"prefix", https, 1, 1}, point{"release", https, 1, 1})
		if https {
			// A tls_handshake point exists only for an https URL.
			points = append(points, point{"tls", true, 1, 1})
		}
	}
	for _, tc := range causes {
		for _, pt := range points {
			scheme := map[bool]string{true: "https", false: "http"}[pt.https]
			t.Run(string(tc.reason)+"/"+scheme+"/"+pt.stage, func(t *testing.T) {
				clock := testkit.NewWaitClock(time.Now())
				ctx, cancel := context.WithCancelCause(context.Background())
				defer cancel(nil)
				var once sync.Once
				second := errNetworkChanged
				if tc.cause == errNetworkChanged {
					second = context.Canceled
				}
				interrupt := func() { once.Do(func() { cancel(tc.cause); cancel(second) }) } // first condition wins
				plain := !pt.https
				served, err := loopback.ListenHTTP(loopback.HTTPConfig{Certificate: cert, Plain: plain, Raw: []byte("HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n")})
				if err != nil {
					t.Fatal(err)
				}
				defer served.Close()
				stalled, err := loopback.ListenHTTP(loopback.HTTPConfig{Certificate: cert, Plain: plain, Stall: true})
				if err != nil {
					t.Fatal(err)
				}
				defer stalled.Close()
				partial, err := loopback.ListenHTTP(loopback.HTTPConfig{Certificate: cert, Plain: plain, Stall: true, Raw: []byte("HTTP/1.1 200 OK\r\nContent-Length: 10\r\n\r\nok")})
				if err != nil {
					t.Fatal(err)
				}
				defer partial.Close()
				handshake, err := loopback.ListenTLS(loopback.TLSConfig{Mode: loopback.TLSStall})
				if err != nil {
					t.Fatal(err)
				}
				defer handshake.Close()
				dial := &http01PointDial{hit: interrupt, route: func(int) (string, bool) {
					switch pt.stage {
					case "connect":
						return "", true
					case "tls":
						return handshake.Address(), false
					case "headers":
						return stalled.Address(), false
					case "prefix":
						return partial.Address(), false
					}
					return served.Address(), false
				}}
				codec := net03.Codec{}
				def := codec.Definition()
				lim := jsondecode.Limits{MaxBytes: 4096, MaxDepth: 16, MaxNodes: 64}
				in, err := jsondecode.Decode([]byte(`{"url":"`+scheme+`://service.example/check","expected_status":200}`), def.Input, lim)
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
				np := params.(net03.Parameters)
				test := task.TestVersion{ID: "NET-03", Version: "0.1.0"}
				regs := []task.TargetRegistration{{ID: "net03-a", Test: test, Role: task.MeasurementTarget, Tuple: np.Target(), Paths: task.Paths{Direct: true}, ExpiresAt: task.TimestampFromTime(clock.Now().Add(time.Hour))}}
				lookups := &atomic.Int32{}
				resolver := http01PointResolver{lookups: lookups, interruptAt: -1, hit: interrupt}
				switch pt.stage {
				case "dns":
					resolver.interruptAt = 0
				case "after-dns":
					resolver.interruptAt, resolver.afterAnswer = 0, true
				}
				gate := NewTargets(test, "direct", regs, clock, nil).WithResolver(resolver)
				var dialers DialerSource = dial
				if pt.stage == "after-connect" {
					dialers = net03AfterConnect{dial}
				}
				p := tls01Plan(t, false, 30*time.Second, task.AttemptLimits{Connections: 4, BytesSent: 32768, BytesReceived: 131072}, dialers, gate)
				p.TaskID, p.Test, p.Parameters = "net03-points", test, params
				source := p.Context.(*networkcontext.Source)
				snap, err := source.Capture(ctx)
				if err != nil {
					t.Fatal(err)
				}
				p.Result, err = result.New(result.Binding{TaskID: p.TaskID, Output: result.Output{Test: model.Test{ID: "NET-03", Version: "0.1.0", ResultVersion: "0.1.0"}, Limits: np.Limits(), HTTPS: np.HTTPS()}, MaxAttempts: 1, MaxDurationMS: 30000, Implementation: p.Implementation, ContextDecoder: source.Decoder(), Path: snap.Path(), AccessGeneration: snap.Generation()})
				if err != nil {
					t.Fatal(err)
				}
				done := make(chan struct{})
				go func() {
					defer close(done)
					wait, stop := context.WithTimeout(ctx, 5*time.Second)
					defer stop()
					var e error
					switch pt.stage {
					case "tls":
						e = handshake.Wait(wait, func(c loopback.TLSCounts) bool { return c.In > 0 })
					case "headers":
						e = stalled.Wait(wait, func(c loopback.HTTPCounts) bool { return len(c.Requests) > 0 })
					case "prefix":
						e = partial.Wait(wait, func(c loopback.HTTPCounts) bool { return c.Out > 0 })
					default:
						return
					}
					if e == nil {
						interrupt()
					} else {
						t.Error("stage was not reached", e)
						cancel(context.Canceled)
					}
				}()
				if pt.stage == "before-start" {
					interrupt()
				}
				runner := net03PointRunner{inner: &net03.Runner{Roots: ca.Pool()}, stage: pt.stage, hit: interrupt}
				a, _, v, err := (&Engine{clock: clock}).attempt(ctx, p, runner, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
				if err != nil {
					t.Fatal(err)
				}
				<-done
				view := v.View()
				switch pt.stage {
				case "before-start":
					if a.Phase() != NotRun || a.Reason() != tc.reason || view.Measurement != nil || view.NetworkContext != nil {
						t.Fatalf("phase %v reason %s", a.Phase(), a.Reason())
					}
				case "release":
					if a.Phase() != Completed || a.Reason() != "" || view.Measurement == nil || view.Measurement.Outcome != "validated" || view.Measurement.PlatformValidation != "unknown" {
						t.Fatal("release replaced the recorded outcome")
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
				if clock.Timers() != 0 {
					t.Fatal("timer leaked")
				}
			})
		}
	}
}
