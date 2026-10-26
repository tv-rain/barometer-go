package execution

import (
	"context"
	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/measurement"
	dnsmeasurement "github.com/tv-rain/barometer-go/internal/measurement/dns01"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/networkcontext"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit"
	"github.com/tv-rain/barometer-go/internal/testkit/capability"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
	"io"
	"sync/atomic"
	"testing"
	"time"
)

// Exercise all termination reasons at the engine/production-runner boundary.
// Suspension is injected as an execution cause; this is no native lifecycle
// bridge or device evidence. The root tests separately exercise real CONNECT.
// requirement-ref: REQ-DNS-01-0039
// requirement-ref: REQ-RESULT-0051
// requirement-ref: REQ-RESULT-0052
// requirement-ref: REQ-RESULT-0053
func TestDns01AllTerminationBoundaries(t *testing.T) {
	causes := []struct {
		cause  error
		reason task.ExecutionReason
	}{
		{context.Canceled, task.CancelledByHost}, {errNetworkChanged, task.NetworkChanged}, {errDeadline, task.AttemptDeadlineExceeded},
		{pathError{task.PlatformSuspended}, task.PlatformSuspended}, {errProxyUnavailable, task.ProxyUnavailable}, {errProxyConnect, task.ProxyConnectFailed}, {ErrProxyReconfigured, task.ProxyReconfigured},
	}
	for _, transport := range []string{"udp", "tcp"} {
		for _, tc := range causes {
			for _, stage := range []string{"before start", "generation and family", "connect", "write", "read", "normalize", "window1", "window2", "system", "close"} {
				if transport == "tcp" && (stage == "window1" || stage == "window2" || stage == "system") {
					continue
				}
				t.Run(transport+"/"+string(tc.reason)+"/"+stage, func(t *testing.T) {
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
					codec := dnsmeasurement.Codec{}
					def := codec.Definition()
					input, err := jsondecode.Decode([]byte(dnsBoundaryInput(stage)), def.Input, jsondecode.Limits{MaxBytes: 1024, MaxDepth: 8, MaxNodes: 32})
					if err != nil {
						t.Fatal(err)
					}
					options, err := jsondecode.Decode([]byte(dnsBoundaryOptions(stage, transport)), def.Options, jsondecode.Limits{MaxBytes: 1024, MaxDepth: 8, MaxNodes: 32})
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
					decoder, err := result.New(result.Binding{TaskID: "dns-boundaries", Output: result.Output{Test: model.Test{ID: "DNS-01", Version: "0.1.0", ResultVersion: "0.1.0"}, DNSMode: dnsBoundaryEngine(stage, transport), QueryType: "A", Limits: observation.ProtocolLimits{RetainDNSAnswers: true}}, MaxAttempts: 1, MaxDurationMS: 30000, Implementation: impl, ContextDecoder: cd, Path: path, AccessGeneration: 1})
					if err != nil {
						t.Fatal(err)
					}
					dial := &dnsTerminationDial{stage: stage, transport: transport, interrupt: interrupt, clock: clock}
					targets := dnsTerminationTarget{tcpTerminationTarget{stage: stage, interrupt: interrupt}}
					p := Plan{TaskID: "dns-boundaries", Test: task.TestVersion{ID: "DNS-01", Version: "0.1.0"}, Parameters: params, Network: task.NetworkRequirements{VPNPolicy: "any"}, Deadline: 30 * time.Second, Limits: task.AttemptLimits{Connections: 1, BytesSent: 65536, BytesReceived: 262144}, Implementation: impl, Result: decoder, Context: capture, Dialers: dial, Targets: targets}
					engine := &Engine{clock: dnsBoundaryClock{clock, dial}}
					attempt, usage, value, err := engine.attempt(ctx, p, &dnsmeasurement.Runner{}, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
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
						if attempt.Reason() != "" || v.Measurement == nil || v.Measurement.Outcome != "resolved" {
							t.Fatal("cleanup replaced terminal outcome")
						}
					} else if attempt.Reason() != tc.reason || v.Measurement != nil {
						t.Fatal("lost first reason or invented protocol outcome")
					}
					if want == NotRun && v.NetworkContext != nil {
						t.Fatal("context on not_run")
					}
					if usage.BytesSent > 65536 || usage.BytesReceived > 262144 || dial.calls > 1 || int(dial.closed.Load()) != dial.calls {
						t.Fatalf("traffic or leaked connection: %+v %+v", usage, dial)
					}
					if clock.Timers() != 0 || network.Watching() != 0 {
						t.Fatal("timer or network watch leaked")
					}
				})
			}
		}
	}
}

func dnsBoundaryInput(stage string) string {
	if stage == "system" {
		return `{"hostname":"example.org","record_type":"A"}`
	}
	return `{"hostname":"example.org","record_type":"A","resolver":{"ip_address":"8.8.8.8"}}`
}
func dnsBoundaryOptions(stage, transport string) string {
	if stage == "system" {
		return `{}`
	}
	if stage == "window1" || stage == "window2" {
		return `{"resolver_mode":"direct","collect_additional_responses":true,"additional_response_window_ms":500}`
	}
	return `{"resolver_mode":"direct","transport":"` + transport + `"}`
}
func dnsBoundaryEngine(stage, transport string) string {
	if stage == "system" {
		return "system"
	}
	return transport
}

type dnsTerminationTarget struct{ tcpTerminationTarget }

func (t dnsTerminationTarget) Lookup(ctx context.Context, _, _ string, start func()) (bool, string, error) {
	start()
	t.interrupt()
	return false, "", ctx.Err()
}

type dnsTerminationDial struct {
	transport string
	stage     string
	interrupt func()
	clock     *testkit.WaitClock
	calls     int
	closed    atomic.Int32
	armed     atomic.Bool
}

func (d *dnsTerminationDial) Dialer(int) (measurement.Dialer, error) { return d, nil }
func (d *dnsTerminationDial) DialContext(ctx context.Context, _, _ string) (measurement.Conn, error) {
	d.calls++
	if d.stage == "connect" {
		d.interrupt()
		d.closed.Add(1)
		return nil, ctx.Err()
	}
	return &dnsTerminationConn{dial: d, ctx: ctx}, nil
}

type dnsTerminationConn struct {
	dial   *dnsTerminationDial
	ctx    context.Context
	query  []byte
	frame  []byte
	offset int
	reads  int
}

func (c *dnsTerminationConn) Close() error {
	c.dial.closed.Add(1)
	if c.dial.stage == "close" {
		c.dial.interrupt()
	}
	return nil
}
func (c *dnsTerminationConn) SetDeadline(time.Time) error { return nil }
func (c *dnsTerminationConn) Write(p []byte) (int, error) {
	if c.dial.stage == "write" {
		c.dial.interrupt()
		return 0, c.ctx.Err()
	}
	c.query = append([]byte(nil), p...)
	if c.dial.transport == "tcp" {
		c.query = c.query[2:]
	}
	return len(p), nil
}
func (c *dnsTerminationConn) Read(p []byte) (int, error) {
	c.reads++
	stage := c.dial.stage
	if stage == "read" || stage == "window1" && c.reads == 2 || stage == "window2" && c.reads == 3 {
		c.dial.interrupt()
		return 0, io.ErrClosedPipe
	}
	response := loopback.DNSResponse(c.query, 0, []loopback.DNSRecord{{Type: "A", Value: "1.1.1.1"}})
	if c.dial.transport == "tcp" {
		if c.frame == nil {
			c.frame = append([]byte{byte(len(response) >> 8), byte(len(response))}, response...)
		}
		response = c.frame[c.offset:]
	}
	n := copy(p, response)
	c.offset += n
	if stage == "normalize" && (c.dial.transport != "tcp" || c.offset == len(c.frame)) {
		c.dial.armed.Store(true)
	}
	return n, nil
}

type dnsBoundaryClock struct {
	*testkit.WaitClock
	dial *dnsTerminationDial
}

func (c dnsBoundaryClock) Now() time.Time {
	if c.dial.armed.Swap(false) {
		c.dial.interrupt()
	}
	return c.WaitClock.Now()
}
