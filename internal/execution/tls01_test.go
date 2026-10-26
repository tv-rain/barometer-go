package execution

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/measurement"
	tlsmeasurement "github.com/tv-rain/barometer-go/internal/measurement/tls01"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/networkcontext"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit"
	"github.com/tv-rain/barometer-go/internal/testkit/capability"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

const tlsName = "measurement.example"

// tls01Plan builds a TLS-01 plan for the production codec over dial. On the
// proxy path the result decoder and context are bound to the proxy source.
func tls01Plan(t *testing.T, proxy bool, deadline time.Duration, limits task.AttemptLimits, dial DialerSource, targets TargetCheck) Plan {
	t.Helper()
	network := capability.NewNetwork(platform.AccessState{Generation: 1, Type: platform.NetworkWiFi, IPv4: platform.AvailabilityYes})
	var source *networkcontext.Source
	var err error
	if proxy {
		source, err = networkcontext.NewProxy(network, 7311, "outline_mobileproxy")
	} else {
		source, err = networkcontext.NewDirect(network)
	}
	if err != nil {
		t.Fatal(err)
	}
	snap, err := source.Capture(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	codec := tlsmeasurement.Codec{}
	def := codec.Definition()
	lim := jsondecode.Limits{MaxBytes: 1024, MaxDepth: 8, MaxNodes: 32}
	input, err := jsondecode.Decode([]byte(`{"ip_address":"8.8.8.8","port":443,"server_name":"`+tlsName+`"}`), def.Input, lim)
	if err != nil {
		t.Fatal(err)
	}
	options, err := jsondecode.Decode([]byte(`{"alpn_protocols":["h2","http/1.1"]}`), def.Options, lim)
	if err != nil {
		t.Fatal(err)
	}
	params, err := codec.Normalize(input, options)
	if err != nil {
		t.Fatal(err)
	}
	impl := model.Implementation{LibraryVersion: "0.1.0", Platform: "android"}
	decoder, err := result.New(result.Binding{TaskID: "tls-plan", Output: result.Output{Test: model.Test{ID: "TLS-01", Version: "0.1.0", ResultVersion: "0.1.0"}}, MaxAttempts: 1, MaxDurationMS: 60000, Implementation: impl, ContextDecoder: source.Decoder(), Path: snap.Path(), AccessGeneration: snap.Generation()})
	if err != nil {
		t.Fatal(err)
	}
	return Plan{TaskID: "tls-plan", Test: task.TestVersion{ID: "TLS-01", Version: "0.1.0"}, Parameters: params, Network: task.NetworkRequirements{VPNPolicy: "any"}, Deadline: deadline, Limits: limits, Implementation: impl, Result: decoder, Context: source, Dialers: dial, Targets: targets}
}

type loopbackDial struct{ address string }

func (d loopbackDial) Dialer(int) (measurement.Dialer, error) { return d, nil }
func (d loopbackDial) DialContext(ctx context.Context, network, _ string) (measurement.Conn, error) {
	var nd net.Dialer
	return nd.DialContext(ctx, network, d.address)
}

// TLS handshake and record bytes cross the socket and the proxy tunnel but
// are not attempt bytes: TLS-01 sends no plaintext, so the meter reports zero
// sent and received bytes and one connection on both paths, and byte limits
// of zero are enough (REQ-LIMIT-0034).
// requirement-ref: REQ-LIMIT-0034
// requirement-ref: REQ-TLS-01-0023
// requirement-ref: REQ-TLS-01-0025
func TestTls01PlaintextAccounting(t *testing.T) {
	ca, err := loopback.NewTestCA()
	if err != nil {
		t.Fatal(err)
	}
	cert, err := ca.Issue(tlsName, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, proxy := range []bool{false, true} {
		t.Run(map[bool]string{false: "direct", true: "proxy"}[proxy], func(t *testing.T) {
			srv, err := loopback.ListenTLS(loopback.TLSConfig{Certificate: cert, ALPN: []string{"h2"}})
			if err != nil {
				t.Fatal(err)
			}
			defer srv.Close()
			var dial DialerSource = loopbackDial{srv.Address()}
			var p *loopback.Proxy
			if proxy {
				p, err = loopback.ListenProxy(loopback.ProxyConfig{Resolve: map[string]string{"8.8.8.8:443": srv.Address()}})
				if err != nil {
					t.Fatal(err)
				}
				defer p.Close()
				if dial, err = NewProxyDialers("127.0.0.1", p.Port()); err != nil {
					t.Fatal(err)
				}
			}
			plan := tls01Plan(t, proxy, 30*time.Second, task.AttemptLimits{Connections: 1}, dial, tcpTerminationTarget{interrupt: func() {}})
			engine := &Engine{clock: testkit.NewWaitClock(time.Now())}
			attempt, usage, value, err := engine.attempt(context.Background(), plan, &tlsmeasurement.Runner{Roots: ca.Pool()}, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
			if err != nil {
				t.Fatal(err)
			}
			v := value.View()
			if attempt.Phase() != Completed || v.Measurement == nil || v.Measurement.Outcome != "handshake_succeeded" || len(v.Measurement.Observations) != 2 {
				t.Fatalf("phase %v reason %q", attempt.Phase(), attempt.Reason())
			}
			d, _ := v.Measurement.Observations[1].TLS()
			if d.ALPN != "h2" || d.TLSVersion != "TLS1.3" || d.CertificateValidation != "trusted" {
				t.Fatalf("details %+v", d)
			}
			if usage != (measurement.Usage{Connections: 1}) {
				t.Fatalf("usage %+v", usage)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.WaitIdle(ctx); err != nil {
				t.Fatal(err)
			}
			c := srv.Counts()
			if c.Accepted != 1 || c.In == 0 || c.Out == 0 || c.Plaintext != 0 || len(c.Hellos) != 1 || c.Hellos[0].ServerName != tlsName ||
				len(c.Hellos[0].ALPN) != 2 || c.Hellos[0].ALPN[0] != "h2" || c.Hellos[0].ALPN[1] != "http/1.1" {
				t.Fatalf("server %+v", c)
			}
			if proxy {
				if err := p.WaitIdle(ctx); err != nil {
					t.Fatal(err)
				}
				pc := p.Counts()
				if pc.Accepted != 1 || pc.TunnelIn != c.In || pc.TunnelOut != c.Out {
					t.Fatalf("tunnel %+v, server %+v", pc, c)
				}
			}
		})
	}
}

// Session moves metering from the stream to the session's plaintext side
// once: raw bytes are not counted, plaintext bytes are, the metered
// connection refuses further I/O, and the plaintext side keeps the budget.
// requirement-ref: REQ-LIMIT-0034
// requirement-ref: REQ-LIMIT-0027
func TestScopeSessionMetersPlaintext(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	client, server := net.Pipe()
	defer server.Close()
	go func() { _, _ = io.Copy(io.Discard, server) }()
	s := newScope(ctx, pipeDialer{client}, task.AttemptLimits{BytesSent: 8, BytesReceived: 8, Connections: 1}, cancel)
	s.targets = tcpTerminationTarget{interrupt: func() {}}
	defer s.close()
	conn, err := s.DialContext(ctx, "tcp", "8.8.8.8:443")
	if err != nil {
		t.Fatal(err)
	}
	sc, ok := conn.(measurement.SessionConn)
	if !ok {
		t.Fatal("attempt connection cannot carry a session")
	}
	raw, plaintext, err := sc.Session()
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := sc.Session(); err == nil {
		t.Fatal("second session accepted")
	}
	if _, err := conn.Write([]byte("x")); err == nil {
		t.Fatal("metered stream accepted I/O after Session")
	}
	if _, err := raw.Write(make([]byte, 64)); err != nil {
		t.Fatal(err)
	}
	if u := s.Usage(); u.BytesSent != 0 {
		t.Fatalf("raw session bytes metered: %+v", u)
	}
	var buf bytes.Buffer
	side := plaintext(&buf)
	if n, err := side.Write([]byte("hello")); n != 5 || err != nil {
		t.Fatal(n, err)
	}
	if u := s.Usage(); u.BytesSent != 5 {
		t.Fatalf("plaintext not metered: %+v", u)
	}
	if n, err := side.Write([]byte("four")); n != 0 || !errors.Is(err, errBudget) {
		t.Fatalf("over-budget plaintext write issued: %d %v", n, err)
	}
	if buf.String() != "hello" || !errors.Is(context.Cause(ctx), errBudget) {
		t.Fatal("budget did not end the attempt")
	}
	if _, err := raw.Write([]byte("x")); err == nil {
		t.Fatal("raw stream usable after the attempt ended")
	}
}

type pipeDialer struct{ conn net.Conn }

func (d pipeDialer) DialContext(context.Context, string, string) (measurement.Conn, error) {
	return d.conn, nil
}

// TLS-01 interruptions at each Procedure point follow GO-SPEC-ISSUE-004:
// before the network check is not_run; from the check until the outcome is
// recorded is incomplete with the first reason and no measurement; while the
// session closes after the outcome is recorded, the result stays completed.
// requirement-ref: REQ-TLS-01-0027
// requirement-ref: REQ-RESULT-0051
// requirement-ref: REQ-RESULT-0052
// requirement-ref: REQ-RESULT-0053
func TestTls01AllTerminationBoundaries(t *testing.T) {
	ca, err := loopback.NewTestCA()
	if err != nil {
		t.Fatal(err)
	}
	cert, err := ca.Issue(tlsName, false)
	if err != nil {
		t.Fatal(err)
	}
	causes := []struct {
		cause  error
		reason task.ExecutionReason
	}{
		{context.Canceled, task.CancelledByHost}, {errNetworkChanged, task.NetworkChanged}, {errDeadline, task.AttemptDeadlineExceeded},
		{pathError{task.PlatformSuspended}, task.PlatformSuspended}, {errProxyUnavailable, task.ProxyUnavailable}, {errProxyConnect, task.ProxyConnectFailed}, {ErrProxyReconfigured, task.ProxyReconfigured},
	}
	for _, tc := range causes {
		for _, stage := range []string{"before start", "generation and family", "connect", "before hello", "handshake", "close"} {
			t.Run(string(tc.reason)+"/"+stage, func(t *testing.T) {
				mode := loopback.TLSAccept
				if stage == "handshake" {
					mode = loopback.TLSStall
				}
				srv, err := loopback.ListenTLS(loopback.TLSConfig{Mode: mode, Certificate: cert})
				if err != nil {
					t.Fatal(err)
				}
				defer srv.Close()
				clock := testkit.NewWaitClock(time.Now())
				ctx, cancel := context.WithCancelCause(context.Background())
				defer cancel(nil)
				interrupt := func() { cancel(tc.cause); cancel(errNetworkChanged) } // first condition wins
				plan := tls01Plan(t, false, 30*time.Second, task.AttemptLimits{Connections: 1}, &tlsStageDial{stage: stage, address: srv.Address(), interrupt: interrupt, srv: srv}, tcpTerminationTarget{stage: stage, interrupt: interrupt})
				if stage == "before start" {
					plan.Context = tcpInterruptedCapture{plan.Context.(*networkcontext.Source), interrupt}
				}
				engine := &Engine{clock: clock}
				attempt, usage, value, err := engine.attempt(ctx, plan, &tlsmeasurement.Runner{Roots: ca.Pool()}, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
				if err != nil {
					t.Fatal(err)
				}
				want := Incomplete
				switch stage {
				case "before start":
					want = NotRun
				case "close":
					want = Completed
				}
				v := value.View()
				if attempt.Phase() != want {
					t.Fatalf("phase %v, want %v (%s)", attempt.Phase(), want, attempt.Reason())
				}
				if stage == "close" {
					if attempt.Reason() != "" || v.Measurement == nil || v.Measurement.Outcome != "handshake_succeeded" {
						t.Fatal("cleanup replaced terminal outcome")
					}
				} else if attempt.Reason() != tc.reason || v.Measurement != nil {
					t.Fatalf("reason %q, measurement %v", attempt.Reason(), v.Measurement != nil)
				}
				if want == NotRun && v.NetworkContext != nil {
					t.Fatal("context on not_run")
				}
				if usage.BytesSent != 0 || usage.BytesReceived != 0 || usage.Connections > 1 {
					t.Fatalf("usage %+v", usage)
				}
				ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel2()
				if err := srv.WaitIdle(ctx2); err != nil {
					t.Fatal("socket not released")
				}
				c := srv.Counts()
				if c.Accepted > 1 || c.Plaintext != 0 || (stage == "before start" || stage == "generation and family") && c.Accepted != 0 {
					t.Fatalf("server %+v", c)
				}
				if clock.Timers() != 0 {
					t.Fatal("timer leaked")
				}
			})
		}
	}
}

// tlsStageDial dials the loopback server and injects the interruption at the
// named point: during the connect, after TCP success before the ClientHello,
// while the peer stalls mid-handshake, or during close after the outcome.
type tlsStageDial struct {
	stage, address string
	interrupt      func()
	srv            *loopback.TLSServer
}

func (d *tlsStageDial) Dialer(int) (measurement.Dialer, error) { return d, nil }
func (d *tlsStageDial) DialContext(ctx context.Context, network, _ string) (measurement.Conn, error) {
	if d.stage == "connect" {
		d.interrupt()
		<-ctx.Done()
		return nil, ctx.Err()
	}
	var nd net.Dialer
	c, err := nd.DialContext(ctx, network, d.address)
	if err != nil {
		return nil, err
	}
	switch d.stage {
	case "before hello":
		d.interrupt()
	case "handshake":
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if d.srv.Wait(ctx, func(c loopback.TLSCounts) bool { return c.In > 0 }) == nil {
				d.interrupt()
			}
		}()
	case "close":
		return closeHook{Conn: c, hook: d.interrupt}, nil
	}
	return c, nil
}

type closeHook struct {
	net.Conn
	hook func()
}

func (c closeHook) Close() error { c.hook(); return c.Conn.Close() }
