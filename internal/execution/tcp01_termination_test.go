package execution

import (
	"context"
	"fmt"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/measurement"
	tcpmeasurement "github.com/tv-rain/barometer-go/internal/measurement/tcp01"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/networkcontext"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit"
	"github.com/tv-rain/barometer-go/internal/testkit/capability"
)

// Exercise all termination reasons at the engine/production-runner boundary.
// Suspension is injected as an execution cause; this is no native lifecycle
// bridge or device evidence. The root tests separately exercise real CONNECT.
// requirement-ref: REQ-TCP-01-0026
// requirement-ref: REQ-RESULT-0051
// requirement-ref: REQ-RESULT-0052
// requirement-ref: REQ-RESULT-0053
func TestTcp01AllTerminationBoundaries(t *testing.T) {
	causes := []struct {
		cause  error
		reason task.ExecutionReason
	}{
		{context.Canceled, task.CancelledByHost}, {errNetworkChanged, task.NetworkChanged}, {errDeadline, task.AttemptDeadlineExceeded},
		{pathError{task.PlatformSuspended}, task.PlatformSuspended}, {errProxyUnavailable, task.ProxyUnavailable}, {errProxyConnect, task.ProxyConnectFailed}, {ErrProxyReconfigured, task.ProxyReconfigured},
	}
	for _, tc := range causes {
		for _, stage := range []string{"before start", "generation and family", "connect", "close"} {
			t.Run(string(tc.reason)+"/"+stage, func(t *testing.T) {
				clock := testkit.NewWaitClock(time.Now())
				ctx, cancel := context.WithCancelCause(context.Background())
				defer cancel(nil)
				interrupt := func() { cancel(tc.cause); cancel(errNetworkChanged) } // first condition wins
				network := capability.NewNetwork(platform.AccessState{Generation: 1, Type: platform.NetworkWiFi, IPv4: platform.AvailabilityYes})
				source, err := networkcontext.NewDirect(network)
				if err != nil {
					t.Fatal(err)
				}
				var capture ContextSource = source
				if stage == "before start" {
					capture = tcpInterruptedCapture{source, interrupt}
				}
				codec := tcpmeasurement.Codec{}
				def := codec.Definition()
				input, err := jsondecode.Decode([]byte(`{"ip_address":"8.8.8.8","port":443}`), def.Input, jsondecode.Limits{MaxBytes: 1024, MaxDepth: 8, MaxNodes: 32})
				if err != nil {
					t.Fatal(err)
				}
				options, err := jsondecode.Decode([]byte(`{}`), def.Options, jsondecode.Limits{MaxBytes: 1024, MaxDepth: 8, MaxNodes: 32})
				if err != nil {
					t.Fatal(err)
				}
				params, err := codec.Normalize(input, options)
				if err != nil {
					t.Fatal(err)
				}
				cd, _ := observation.NewContextDecoder("0.1.0", observation.ContextApprovals{})
				path, _ := observation.NewPath(1, 0, "direct", "")
				impl := model.Implementation{LibraryVersion: "0.1.0", Platform: "android"}
				decoder, err := result.New(result.Binding{TaskID: "tcp-boundaries", Output: result.Output{Test: model.Test{ID: "TCP-01", Version: "0.1.0", ResultVersion: "0.1.0"}}, MaxAttempts: 1, MaxDurationMS: 30000, Implementation: impl, ContextDecoder: cd, Path: path, AccessGeneration: 1})
				if err != nil {
					t.Fatal(err)
				}
				dial := &tcpTerminationDial{stage: stage, interrupt: interrupt}
				targets := tcpTerminationTarget{stage: stage, interrupt: interrupt}
				p := Plan{TaskID: "tcp-boundaries", Test: task.TestVersion{ID: "TCP-01", Version: "0.1.0"}, Parameters: params, Network: task.NetworkRequirements{VPNPolicy: "any"}, Deadline: 30 * time.Second, Limits: task.AttemptLimits{Connections: 1}, Implementation: impl, Result: decoder, Context: capture, Dialers: dial, Targets: targets}
				engine := &Engine{clock: clock}
				attempt, usage, value, err := engine.attempt(ctx, p, &tcpmeasurement.Runner{}, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
				if err != nil {
					t.Fatal(err)
				}
				want := Incomplete
				if stage == "before start" {
					want = NotRun
				}
				if stage == "close" {
					want = Completed
				}
				if attempt.Phase() != want {
					t.Fatalf("phase %v, want %v", attempt.Phase(), want)
				}
				v := value.View()
				if stage == "close" {
					if attempt.Reason() != "" || v.Measurement == nil || v.Measurement.Outcome != "connected" {
						t.Fatal("cleanup replaced terminal outcome")
					}
				} else if attempt.Reason() != tc.reason || v.Measurement != nil {
					t.Fatal("lost first reason or invented protocol outcome")
				}
				if want == NotRun && v.NetworkContext != nil {
					t.Fatal("context on not_run")
				}
				if (stage == "before start" || stage == "generation and family") && dial.calls != 0 {
					t.Fatal("dialed after a pre-connect interruption")
				}
				if usage.BytesSent != 0 || usage.BytesReceived != 0 || dial.calls > 1 || dial.closed != dial.calls {
					t.Fatalf("traffic or leaked connection: %+v %+v", usage, dial)
				}
				if clock.Timers() != 0 || network.Watching() != 0 {
					t.Fatal("timer or network watch leaked")
				}
			})
		}
	}
}

// tcp01Plan builds a direct TCP-01 plan for the production runner over a
// scripted dialer, with the given connection limit and attempt deadline.
func tcp01Plan(t *testing.T, connections int, deadline time.Duration, dial DialerSource) Plan {
	t.Helper()
	network := capability.NewNetwork(platform.AccessState{Generation: 1, Type: platform.NetworkWiFi, IPv4: platform.AvailabilityYes})
	source, err := networkcontext.NewDirect(network)
	if err != nil {
		t.Fatal(err)
	}
	codec := tcpmeasurement.Codec{}
	def := codec.Definition()
	input, err := jsondecode.Decode([]byte(`{"ip_address":"8.8.8.8","port":443}`), def.Input, jsondecode.Limits{MaxBytes: 1024, MaxDepth: 8, MaxNodes: 32})
	if err != nil {
		t.Fatal(err)
	}
	options, err := jsondecode.Decode([]byte(`{}`), def.Options, jsondecode.Limits{MaxBytes: 1024, MaxDepth: 8, MaxNodes: 32})
	if err != nil {
		t.Fatal(err)
	}
	params, err := codec.Normalize(input, options)
	if err != nil {
		t.Fatal(err)
	}
	cd, _ := observation.NewContextDecoder("0.1.0", observation.ContextApprovals{})
	path, _ := observation.NewPath(1, 0, "direct", "")
	impl := model.Implementation{LibraryVersion: "0.1.0", Platform: "android"}
	decoder, err := result.New(result.Binding{TaskID: "tcp-plan", Output: result.Output{Test: model.Test{ID: "TCP-01", Version: "0.1.0", ResultVersion: "0.1.0"}}, MaxAttempts: 1, MaxDurationMS: 60000, Implementation: impl, ContextDecoder: cd, Path: path, AccessGeneration: 1})
	if err != nil {
		t.Fatal(err)
	}
	targets := tcpTerminationTarget{interrupt: func() {}}
	return Plan{TaskID: "tcp-plan", Test: task.TestVersion{ID: "TCP-01", Version: "0.1.0"}, Parameters: params, Network: task.NetworkRequirements{VPNPolicy: "any"}, Deadline: deadline, Limits: task.AttemptLimits{Connections: connections}, Implementation: impl, Result: decoder, Context: source, Dialers: dial, Targets: targets}
}

// The production TCP-01 runner dials only through the attempt meter, which
// counts its one connection: with max_connections 0 the dial is refused before
// the transport and the attempt ends with resource_limit; at 1 and above the
// metered usage is exactly one connection and no application bytes.
// requirement-ref: REQ-TCP-01-0022
// requirement-ref: REQ-LIMIT-0006
func TestTcp01ConnectionMeter(t *testing.T) {
	for _, limit := range []int{0, 1, 2} {
		t.Run(fmt.Sprint(limit), func(t *testing.T) {
			dial := &tcpTerminationDial{}
			p := tcp01Plan(t, limit, 30*time.Second, dial)
			engine := &Engine{clock: testkit.NewWaitClock(time.Now())}
			attempt, usage, value, err := engine.attempt(context.Background(), p, &tcpmeasurement.Runner{}, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
			if err != nil {
				t.Fatal(err)
			}
			if limit == 0 {
				if attempt.Phase() != Incomplete || attempt.Reason() != task.ResourceLimit || value.View().Measurement != nil {
					t.Fatalf("phase %v reason %q", attempt.Phase(), attempt.Reason())
				}
				if usage != (measurement.Usage{}) || dial.calls != 0 {
					t.Fatalf("refused connection reached the transport: %+v, %d dials", usage, dial.calls)
				}
				return
			}
			if attempt.Phase() != Completed || value.View().Measurement == nil || value.View().Measurement.Outcome != "connected" {
				t.Fatalf("phase %v reason %q", attempt.Phase(), attempt.Reason())
			}
			if usage != (measurement.Usage{Connections: 1}) || dial.calls != 1 || dial.closed != 1 {
				t.Fatalf("usage %+v, %d dials, %d closed", usage, dial.calls, dial.closed)
			}
		})
	}
}

// M-TCP-01-LIMITS socket release: whichever path ends the connect, a socket
// the platform returns is closed exactly once before the attempt reports, and
// no second dial follows. Late sockets arrive after host cancellation, after
// the attempt deadline and after the 15-second operation bound, and a
// successful socket is closed after its outcome. A refusal returns no socket,
// as the production dialers do, and is not retried.
// requirement-ref: REQ-TCP-01-0022
func TestTcp01ReleasesSocketOnEveryPath(t *testing.T) {
	for _, tc := range []struct {
		name     string
		deadline time.Duration
		phase    Phase
		reason   task.ExecutionReason
		outcome  string
	}{
		{"connected", 30 * time.Second, Completed, "", "connected"},
		{"refused", 30 * time.Second, Completed, "", "connection_refused"},
		{"operation bound", 30 * time.Second, Completed, "", "timeout"},
		{"attempt deadline", 10 * time.Second, Incomplete, task.AttemptDeadlineExceeded, ""},
		{"host cancellation", 30 * time.Second, Incomplete, task.CancelledByHost, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clock := testkit.NewWaitClock(time.Now())
			ctx, cancel := context.WithCancelCause(context.Background())
			defer cancel(nil)
			dial := &releaseDial{}
			dial.dial = func(op context.Context) (measurement.Conn, error) {
				dial.calls++
				if tc.name == "refused" {
					return nil, syscall.ECONNREFUSED
				}
				conn := &releaseConn{}
				dial.conns = append(dial.conns, conn)
				switch tc.name {
				case "operation bound":
					_ = clock.Advance(15 * time.Second)
					<-op.Done()
				case "attempt deadline":
					_ = clock.Advance(10 * time.Second)
					<-op.Done()
				case "host cancellation":
					cancel(context.Canceled)
					<-op.Done()
				}
				return conn, nil
			}
			p := tcp01Plan(t, 1, tc.deadline, dial)
			engine := &Engine{clock: clock}
			attempt, usage, value, err := engine.attempt(ctx, p, &tcpmeasurement.Runner{}, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
			if err != nil {
				t.Fatal(err)
			}
			v := value.View()
			if attempt.Phase() != tc.phase || attempt.Reason() != tc.reason {
				t.Fatalf("phase %v reason %q", attempt.Phase(), attempt.Reason())
			}
			if tc.outcome == "" && v.Measurement != nil || tc.outcome != "" && (v.Measurement == nil || v.Measurement.Outcome != tc.outcome) {
				t.Fatal("unexpected measurement")
			}
			if dial.calls != 1 {
				t.Fatalf("%d dials", dial.calls)
			}
			for _, c := range dial.conns {
				if c.closes.Load() != 1 {
					t.Fatal("socket not released exactly once")
				}
			}
			if usage.BytesSent != 0 || usage.BytesReceived != 0 || usage.Connections != 1 {
				t.Fatalf("usage %+v", usage)
			}
			if clock.Timers() != 0 {
				t.Fatal("timer leaked")
			}
		})
	}
}

type releaseDial struct {
	dial  func(context.Context) (measurement.Conn, error)
	calls int
	conns []*releaseConn
}

func (d *releaseDial) Dialer(int) (measurement.Dialer, error) { return d, nil }
func (d *releaseDial) DialContext(ctx context.Context, _, _ string) (measurement.Conn, error) {
	return d.dial(ctx)
}

type releaseConn struct{ closes atomic.Int32 }

func (c *releaseConn) Close() error              { c.closes.Add(1); return nil }
func (*releaseConn) Read([]byte) (int, error)    { panic("unexpected application read") }
func (*releaseConn) Write([]byte) (int, error)   { panic("unexpected application write") }
func (*releaseConn) SetDeadline(time.Time) error { return nil }

type tcpInterruptedCapture struct {
	source    *networkcontext.Source
	interrupt func()
}

func (c tcpInterruptedCapture) Capture(ctx context.Context) (networkcontext.Snapshot, error) {
	s, e := c.source.Capture(ctx)
	c.interrupt()
	return s, e
}

type tcpTerminationTarget struct {
	stage     string
	interrupt func()
}

func (t tcpTerminationTarget) Prepare(_ context.Context, _, address string, _ func()) (string, error) {
	if t.stage == "generation and family" {
		t.interrupt()
	}
	return address, nil
}
func (tcpTerminationTarget) Redirect(task.TargetTuple) (TargetCheck, error) {
	return nil, measurement.ErrTargetRejected
}

type tcpTerminationDial struct {
	stage         string
	interrupt     func()
	calls, closed int
}

func (d *tcpTerminationDial) Dialer(int) (measurement.Dialer, error) { return d, nil }
func (d *tcpTerminationDial) DialContext(ctx context.Context, _, _ string) (measurement.Conn, error) {
	d.calls++
	if d.stage == "connect" {
		d.interrupt()
		d.closed++
		return nil, ctx.Err()
	}
	return &tcpTerminationConn{dial: d}, nil
}

type tcpTerminationConn struct{ dial *tcpTerminationDial }

func (c *tcpTerminationConn) Close() error {
	c.dial.closed++
	if c.dial.stage == "close" {
		c.dial.interrupt()
	}
	return nil
}
func (*tcpTerminationConn) Read([]byte) (int, error)    { panic("unexpected application read") }
func (*tcpTerminationConn) Write([]byte) (int, error)   { panic("unexpected application write") }
func (*tcpTerminationConn) SetDeadline(time.Time) error { return nil }
