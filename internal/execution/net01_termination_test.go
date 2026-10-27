package execution

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/measurement/net01"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/networkcontext"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit"
	"github.com/tv-rain/barometer-go/internal/testkit/capability"
)

// net01Capture interrupts at a chosen point around the collector's read.
type net01Capture struct {
	*networkcontext.Source
	before, after func()
}

func (c net01Capture) Capture(ctx context.Context) (networkcontext.Snapshot, error) {
	c.before()
	s, err := c.Source.Capture(ctx)
	c.after()
	return s, err
}

// net01Runner interrupts after the production runner has recorded its outcome,
// while the engine still holds the generation watch (Procedure step 4).
type net01Runner struct{ after func() }

func (r net01Runner) Run(ctx context.Context, in measurement.Input) (measurement.Output, error) {
	out, err := net01.Runner{}.Run(ctx, in)
	r.after()
	return out, err
}

// noDialers fails the test on any use: NET-01 opens no network path.
type noDialers struct{ calls *int }

func (d noDialers) Dialer(int) (measurement.Dialer, error) {
	*d.calls++
	return nil, errors.New("synthetic: NET-01 asked for a dialer")
}

func net01Plan(t *testing.T, network *capability.Network, capture ContextSource, dials *int) Plan {
	t.Helper()
	codec := net01.Codec{}
	def := codec.Definition()
	empty, err := jsondecode.Decode([]byte(`{}`), def.Input, jsondecode.Limits{MaxBytes: 1024, MaxDepth: 8, MaxNodes: 32})
	if err != nil {
		t.Fatal(err)
	}
	params, err := codec.Normalize(empty, empty)
	if err != nil {
		t.Fatal(err)
	}
	if capture == nil {
		source, err := networkcontext.NewDirect(network)
		if err != nil {
			t.Fatal(err)
		}
		capture = source
	}
	cd, _ := observation.NewContextDecoder("0.1.0", observation.ContextApprovals{})
	path, _ := observation.NewPath(1, 0, "direct", "")
	impl := model.Implementation{LibraryVersion: "0.1.0", Platform: "android"}
	decoder, err := result.New(result.Binding{TaskID: "net01-boundaries", Output: result.Output{Test: model.Test{ID: "NET-01", Version: "0.1.0", ResultVersion: "0.1.0"}}, MaxAttempts: 1, MaxDurationMS: 30000, Implementation: impl, ContextDecoder: cd, Path: path, AccessGeneration: 1})
	if err != nil {
		t.Fatal(err)
	}
	return Plan{TaskID: "net01-boundaries", Test: task.TestVersion{ID: "NET-01", Version: "0.1.0"}, Parameters: params, Network: task.NetworkRequirements{VPNPolicy: "any"},
		Deadline: 30 * time.Second, Implementation: impl, Result: decoder, Context: capture, Dialers: noDialers{dials}}
}

// Exercise every NET-01 Procedure point with every common cause at the
// engine/production-runner boundary. Suspension is injected as an execution
// cause; this is no native lifecycle bridge or device evidence.
// requirement-ref: REQ-NET-01-0027
// requirement-ref: REQ-RESULT-0051
// requirement-ref: REQ-RESULT-0052
// requirement-ref: REQ-RESULT-0053
func TestNet01AllTerminationBoundaries(t *testing.T) {
	causes := []struct {
		cause  error
		reason task.ExecutionReason
	}{
		{context.Canceled, task.CancelledByHost}, {errNetworkChanged, task.NetworkChanged}, {errDeadline, task.AttemptDeadlineExceeded}, {pathError{task.PlatformSuspended}, task.PlatformSuspended},
	}
	for _, tc := range causes {
		for _, stage := range []string{"before start", "during read", "after read", "release"} {
			t.Run(string(tc.reason)+"/"+stage, func(t *testing.T) {
				clock := testkit.NewWaitClock(time.Now())
				ctx, cancel := context.WithCancelCause(context.Background())
				defer cancel(nil)
				second := errNetworkChanged
				if tc.cause == errNetworkChanged {
					second = context.Canceled
				}
				interrupt := func() { cancel(tc.cause); cancel(second) } // first condition wins
				none := func() {}
				network := capability.NewNetwork(platform.AccessState{Generation: 1, Type: platform.NetworkWiFi, IPv4: platform.AvailabilityYes})
				source, err := networkcontext.NewDirect(network)
				if err != nil {
					t.Fatal(err)
				}
				capture := net01Capture{source, none, none}
				var runner measurement.Runner = net01.Runner{}
				switch stage {
				case "before start":
					interrupt()
				case "during read":
					capture.before = interrupt
				case "after read":
					capture.after = interrupt
				case "release":
					runner = net01Runner{interrupt}
				}
				dials := 0
				p := net01Plan(t, network, capture, &dials)
				attempt, usage, value, err := (&Engine{clock: clock}).attempt(ctx, p, runner, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
				if err != nil {
					t.Fatal(err)
				}
				want := map[string]Phase{"before start": NotRun, "during read": Incomplete, "after read": Incomplete, "release": Completed}[stage]
				if attempt.Phase() != want {
					t.Fatalf("phase %v, want %v", attempt.Phase(), want)
				}
				v := value.View()
				if stage == "release" {
					if attempt.Reason() != "" || v.Measurement == nil || v.Measurement.Outcome != "snapshot" || v.Measurement.Access == nil {
						t.Fatal("release replaced the recorded outcome")
					}
				} else if attempt.Reason() != tc.reason || v.Measurement != nil {
					t.Fatal("lost first reason or invented a NET-01 outcome")
				}
				if v.NetworkContext != nil {
					t.Fatal("NET-01 result carries top-level network context")
				}
				if usage != (measurement.Usage{}) || dials != 0 {
					t.Fatalf("network path used: %+v, %d dialer requests", usage, dials)
				}
				if clock.Timers() != 0 || network.Watching() != 0 {
					t.Fatal("timer or network watch leaked")
				}
			})
		}
	}
}

// A real generation change while the platform read is in flight is seen by
// the watch armed for the read, never merged into a snapshot; one after the
// outcome is recorded leaves the outcome.
// requirement-ref: REQ-NET-01-0021
// requirement-ref: REQ-NET-01-0027
func TestNet01GenerationChange(t *testing.T) {
	for _, after := range []bool{false, true} {
		network := capability.NewNetwork(platform.AccessState{Generation: 1, Type: platform.NetworkWiFi, IPv4: platform.AvailabilityYes})
		source, err := networkcontext.NewDirect(network)
		if err != nil {
			t.Fatal(err)
		}
		change := func() {
			network.Set(platform.AccessState{Generation: 2, Type: platform.NetworkCellular, IPv4: platform.AvailabilityYes})
		}
		capture := net01Capture{source, func() {}, change}
		var runner measurement.Runner = net01.Runner{}
		if after {
			capture.after = func() {}
			runner = net01Runner{change}
		}
		dials := 0
		attempt, _, value, err := (&Engine{clock: testkit.NewWaitClock(time.Now())}).attempt(context.Background(), net01Plan(t, network, capture, &dials), runner, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
		if err != nil {
			t.Fatal(err)
		}
		v := value.View()
		if after {
			if attempt.Phase() != Completed || v.Measurement == nil || v.Measurement.Access.Value().NetworkType != "wifi" {
				t.Fatal("change after the outcome replaced it")
			}
		} else if attempt.Phase() != Incomplete || attempt.Reason() != task.NetworkChanged || v.Measurement != nil {
			t.Fatalf("phase %v reason %q", attempt.Phase(), attempt.Reason())
		}
		if network.Watching() != 0 || dials != 0 {
			t.Fatal("watch leaked or path used")
		}
	}
}

// The outer deadline ends a read the platform never answers: the attempt is
// incomplete, the read and watch are released and no path is used.
// requirement-ref: REQ-NET-01-0021
func TestNet01DeadlineReleasesRead(t *testing.T) {
	network := &blockingNetwork{Network: capability.NewNetwork(platform.AccessState{Generation: 1}), entered: make(chan struct{})}
	source, err := networkcontext.NewDirect(network)
	if err != nil {
		t.Fatal(err)
	}
	clock := testkit.NewWaitClock(time.Now())
	dials := 0
	p := net01Plan(t, network.Network, source, &dials)
	done := make(chan struct{})
	var attempt Attempt
	go func() {
		defer close(done)
		attempt, _, _, err = (&Engine{clock: clock}).attempt(context.Background(), p, net01.Runner{}, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
	}()
	<-network.entered
	_ = clock.Advance(p.Deadline)
	<-done
	if err != nil || attempt.Phase() != Incomplete || attempt.Reason() != task.AttemptDeadlineExceeded {
		t.Fatalf("phase %v reason %q", attempt.Phase(), attempt.Reason())
	}
	if clock.Timers() != 0 || network.Watching() != 0 || dials != 0 {
		t.Fatal("timer, watch or path leaked")
	}
}

type blockingNetwork struct {
	*capability.Network
	entered chan struct{}
}

func (n *blockingNetwork) Access(ctx context.Context) (platform.AccessState, error) {
	close(n.entered)
	<-ctx.Done()
	return platform.AccessState{}, ctx.Err()
}
