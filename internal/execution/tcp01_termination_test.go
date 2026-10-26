package execution

import (
	"context"
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
