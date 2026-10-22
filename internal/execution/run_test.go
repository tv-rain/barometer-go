package execution_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/execution"
	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/registry"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/taskdecode"
	"github.com/tv-rain/barometer-go/internal/testkit"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
	"github.com/tv-rain/barometer-go/internal/testkit/output"
	"github.com/tv-rain/barometer-go/internal/validation"
)

// Synthetic fixtures only. The TCP-01 identifier is used so result
// construction applies its output shape; no entry here is a production runner.
var tcp01 = task.TestVersion{ID: "TCP-01", Version: "0.1.0"}

type fixtureCodec struct{}

func (fixtureCodec) Definition() taskdecode.Definition {
	object := func() *jsondecode.Schema {
		return &jsondecode.Schema{Kind: jsondecode.Object, Fields: map[string]jsondecode.Field{}}
	}
	return taskdecode.Definition{TestID: tcp01.ID, TestVersion: tcp01.Version, Input: object(), Options: object()}
}
func (fixtureCodec) Normalize(_, _ *jsondecode.Value) (task.Parameters, error) {
	return loopback.Parameters{}, nil
}

type noTargets struct{}

func (noTargets) Targets(task.Parameters) ([]registry.Target, bool) { return nil, true }

func newRegistry(t *testing.T, r measurement.Runner) *registry.Registry {
	t.Helper()
	reg, err := registry.New([]registry.Entry{{Test: tcp01, Codec: fixtureCodec{}, BaselineRisk: task.LowRisk, Targets: noTargets{}, Proxy: registry.NoProxy{}, Runner: r}})
	if err != nil {
		t.Fatal(err)
	}
	return reg
}

// step is one scripted runner response.
type step struct {
	out     measurement.Output
	err     error
	advance time.Duration // clock movement while the runner works
	cancel  bool          // simulate host cancellation during the attempt
	unbegun bool          // return before the test starts (REQ-RESULT-0051)
}

// scripted is a synthetic runner; it performs no I/O.
type scripted struct {
	mu     sync.Mutex
	steps  []step
	calls  int
	clock  *testkit.WaitClock
	cancel context.CancelFunc
}

func (s *scripted) Run(ctx context.Context, in measurement.Input) (measurement.Output, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if in.Dialer == nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	st := s.steps[s.calls]
	s.calls++
	if !st.unbegun {
		in.Start()
	}
	if st.advance > 0 {
		_ = s.clock.Advance(st.advance)
	}
	if st.cancel {
		s.cancel()
	}
	return st.out, st.err
}

func (s *scripted) Calls() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

type contextSource struct {
	t     *testing.T
	fail  map[int]bool
	calls int
}

func (c *contextSource) Capture(context.Context) (observation.NetworkContext, error) {
	c.calls++
	if c.fail[c.calls] {
		return observation.NetworkContext{}, errors.New("synthetic capture failure")
	}
	n, err := contextDecoder(c.t).Decode([]byte(`{"access":{"network_type":"wifi","vpn_state":"unknown","ipv4_available":"yes","ipv6_available":"unknown","validated":"unknown","network_changed":false},"egress":{"status":"unavailable","lookup_source":"none","network_path":"direct"}}`), path(c.t))
	if err != nil {
		c.t.Fatal(err)
	}
	return n, nil
}

type fixedJitter time.Duration

func (j fixedJitter) Jitter(max time.Duration) time.Duration { return min(time.Duration(j), max) }

func contextDecoder(t *testing.T) *observation.ContextDecoder {
	d, err := observation.NewContextDecoder(observation.Contract, observation.ContextApprovals{})
	if err != nil {
		t.Fatal(err)
	}
	return d
}
func path(t *testing.T) observation.Path {
	p, err := observation.NewPath(1, 0, "direct", "")
	if err != nil {
		t.Fatal(err)
	}
	return p
}

var implementation = model.Implementation{LibraryVersion: "0.1.0", Platform: "android"}

func decoder(t *testing.T) *result.Decoder {
	t.Helper()
	d, err := result.New(result.Binding{TaskID: "campaign-1", Output: result.Output{Test: model.Test{ID: tcp01.ID, Version: tcp01.Version, ResultVersion: result.Contract}},
		MaxAttempts: 2, MaxDurationMS: 60000, Implementation: implementation, ContextDecoder: contextDecoder(t), Path: path(t), AccessGeneration: 1})
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func tcpObservation(t *testing.T, failure string) observation.Observation {
	status, f := "succeeded", ""
	if failure != "" {
		status, f = "failed", `,"failure":"`+failure+`"`
	}
	o, err := observation.DecodeProtocol(observation.Contract, []byte(`{"type":"tcp_connect","status":"`+status+`","duration_ms":0`+f+`,"details":{"address_family":"ipv4"}}`), observation.ProtocolLimits{})
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func connected(t *testing.T) step {
	return step{out: measurement.Output{Outcome: "connected", Observations: []observation.Observation{tcpObservation(t, "")}, Usage: measurement.Usage{BytesSent: 1, BytesReceived: 1, Connections: 1}}}
}
func refused(t *testing.T) step {
	return step{out: measurement.Output{Outcome: "connection_refused", Observations: []observation.Observation{tcpObservation(t, "connection_refused")}, Usage: measurement.Usage{Connections: 1}}}
}
func failure() step { return step{err: measurement.ErrRunner} }

var now = time.Now().UTC().Truncate(time.Hour).Add(12 * time.Hour)

// attemptDeadline is the registered default max_duration_ms.
const attemptDeadline = 30 * time.Second

func schedule(maxAttempts int, on ...task.ExecutionReason) task.Schedule {
	maxRuns := 1
	expires := task.TimestampFromTime(now.Add(time.Hour))
	s := task.Schedule{NotAfter: &expires, Repeat: task.Repeat{Mode: "none"}, MaxRuns: &maxRuns, Retry: task.Retry{MaxAttempts: maxAttempts}}
	if maxAttempts == 2 {
		s.Retry.On = on
		s.Retry.Backoff = &task.Backoff{Mode: "fixed", InitialDelayMS: 30000, MaxDelayMS: 60000, JitterMS: 10000}
	}
	return s
}

type fixture struct {
	runner  *scripted
	clock   *testkit.WaitClock
	store   *output.Store
	ctx     *contextSource
	dialers *loopback.Dialers
	engine  *execution.Engine
	plan    execution.Plan
	run     context.Context
}

func newFixture(t *testing.T, s task.Schedule, steps ...step) *fixture {
	t.Helper()
	// A monotonic-bearing base keeps result timing valid on the manual clock.
	clock := testkit.NewWaitClock(time.Now())
	_ = clock.Advance(now.Sub(clock.Now()))
	run, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	f := &fixture{runner: &scripted{steps: steps, clock: clock, cancel: cancel}, clock: clock, store: output.New(), ctx: &contextSource{t: t}, dialers: &loopback.Dialers{}, run: run}
	engine, err := execution.NewEngine(newRegistry(t, f.runner), clock, fixedJitter(5*time.Second), f.store)
	if err != nil {
		t.Fatal(err)
	}
	f.engine = engine
	f.plan = execution.Plan{TaskID: "campaign-1", Test: tcp01, Parameters: loopback.Parameters{}, Schedule: s, Deadline: attemptDeadline,
		Conditions:     validation.StartContext{DueAt: now, Supported: true, ConstraintsSatisfied: true, SafetySatisfied: true},
		Implementation: implementation, Result: decoder(t), Context: f.ctx, Dialers: f.dialers}
	return f
}

type record struct {
	RunID     string `json:"run_id"`
	Attempt   int    `json:"attempt"`
	Execution struct {
		State  string `json:"state"`
		Reason string `json:"reason"`
	} `json:"execution"`
	NetworkContext json.RawMessage `json:"network_context"`
	Measurement    *struct {
		Outcome string `json:"outcome"`
	} `json:"measurement"`
}

func (f *fixture) records(t *testing.T) []record {
	t.Helper()
	var out []record
	for _, b := range f.store.Records() {
		var r record
		if err := json.Unmarshal(b, &r); err != nil {
			t.Fatal(err)
		}
		// A not_run result never carries Network Context.
		if r.Execution.State == "not_run" && len(r.NetworkContext) != 0 {
			t.Fatalf("not_run result retained network_context: %s", b)
		}
		out = append(out, r)
	}
	return out
}

type want struct{ state, reason, outcome string }

func (f *fixture) expect(t *testing.T, rep execution.Report, err error, wantErr error, attempts []want, waits []time.Duration) {
	t.Helper()
	if !errors.Is(err, wantErr) {
		t.Fatalf("Run() error = %v, want %v", err, wantErr)
	}
	recs := f.records(t)
	if len(recs) != len(attempts) {
		t.Fatalf("records = %d, want %d", len(recs), len(attempts))
	}
	for i, w := range attempts {
		r := recs[i]
		got := want{r.Execution.State, r.Execution.Reason, ""}
		if r.Measurement != nil {
			got.outcome = r.Measurement.Outcome
		}
		if got != w || r.Attempt != i+1 || r.RunID != rep.RunID {
			t.Fatalf("record %d = %+v (attempt %d, run %q), want %+v (run %q)", i, got, r.Attempt, r.RunID, w, rep.RunID)
		}
		if w.state != "not_run" && len(r.NetworkContext) == 0 {
			t.Fatalf("record %d lacks network_context", i)
		}
	}
	if got := f.clock.Waits(); len(got) != len(waits) || len(got) > 0 && got[0] != waits[0] {
		t.Fatalf("waits = %v, want %v", got, waits)
	}
}

// One registered runner per attempt; retries only as the retry object allows,
// with the delay taken from injected time and jitter (REQ-SCHEDULE-0022..0029).
// requirement-ref: REQ-SCHEDULE-0025
// requirement-ref: REQ-SCHEDULE-0029
func TestRunSequencesAttemptsAndRetries(t *testing.T) {
	delay := 35 * time.Second // initial 30 s plus 5 s local jitter
	cases := []struct {
		name     string
		schedule task.Schedule
		steps    []step
		attempts []want
		waits    []time.Duration
	}{
		{"completed", schedule(1), nil, []want{{"completed", "", "connected"}}, nil},
		{"negative outcome is not retried", schedule(2, task.InternalError), []step{refused(t)}, []want{{"completed", "", "connection_refused"}}, nil},
		{"runner failure without retry object", schedule(1), []step{failure()}, []want{{"incomplete", "internal_error", ""}}, nil},
		{"reason not in retry_on", schedule(2, task.PlatformSuspended), []step{failure()}, []want{{"incomplete", "internal_error", ""}}, nil},
		{"retry then complete", schedule(2, task.InternalError), []step{failure(), connected(t)}, []want{{"incomplete", "internal_error", ""}, {"completed", "", "connected"}}, []time.Duration{delay}},
		{"retry budget exhausted", schedule(2, task.InternalError), []step{failure(), failure()}, []want{{"incomplete", "internal_error", ""}, {"incomplete", "internal_error", ""}}, []time.Duration{delay}},
		{"invalid runner output", schedule(1), []step{{out: measurement.Output{Outcome: "connected", Observations: []observation.Observation{tcpObservation(t, "connection_refused")}}}}, []want{{"incomplete", "internal_error", ""}}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			steps := tc.steps
			if steps == nil {
				steps = []step{connected(t)}
			}
			f := newFixture(t, tc.schedule, steps...)
			rep, err := f.engine.Run(f.run, f.plan)
			f.expect(t, rep, err, nil, tc.attempts, tc.waits)
			if f.runner.Calls() != len(tc.attempts) || len(rep.Attempts) != len(tc.attempts) || len(f.dialers.Issued()) != len(tc.attempts) {
				t.Fatalf("runner calls = %d, report = %d, dialers = %d; want %d each", f.runner.Calls(), len(rep.Attempts), len(f.dialers.Issued()), len(tc.attempts))
			}
			for i, a := range rep.Attempts {
				if a.Number != i+1 || !a.Attempt.Terminal() {
					t.Fatalf("attempt report %d = %+v", i, a)
				}
			}
		})
	}
}

func TestRetryDelayIsCappedAndBoundedByExpiry(t *testing.T) {
	s := schedule(2, task.InternalError)
	s.Retry.Backoff.JitterMS = 60000 // 30 s + 60 s jitter exceeds max_delay_ms
	f := newFixture(t, s, failure(), connected(t))
	f.engine, _ = execution.NewEngine(newRegistry(t, f.runner), f.clock, fixedJitter(time.Minute), f.store)
	rep, err := f.engine.Run(f.run, f.plan)
	f.expect(t, rep, err, nil, []want{{"incomplete", "internal_error", ""}, {"completed", "", "connected"}}, []time.Duration{time.Minute})

	// A realized delay reaching not_after produces no retry and no wait.
	s = schedule(2, task.InternalError)
	expires := task.TimestampFromTime(now.Add(35 * time.Second))
	s.NotAfter = &expires
	f = newFixture(t, s, failure(), connected(t))
	rep, err = f.engine.Run(f.run, f.plan)
	f.expect(t, rep, err, nil, []want{{"incomplete", "internal_error", ""}}, nil)
}

// Eligibility is re-checked after the wait: a window that closes during the
// backoff ends the run without a second attempt.
func TestRetryRechecksEligibilityAfterWait(t *testing.T) {
	s := schedule(2, task.InternalError)
	minute := now.Hour()*60 + now.Minute()
	s.Window = &task.TimeWindow{Timezone: "local", Ranges: []task.TimeRange{{Start: minute, End: minute + 1}}}
	f := newFixture(t, s, failure(), connected(t))
	f.plan.Conditions.LocalZone = time.UTC
	_ = f.clock.Advance(30 * time.Second) // the 35 s wait crosses the window end
	rep, err := f.engine.Run(f.run, f.plan)
	f.expect(t, rep, err, nil, []want{{"incomplete", "internal_error", ""}}, []time.Duration{35 * time.Second})
	if f.runner.Calls() != 1 {
		t.Fatalf("runner calls = %d after closed window", f.runner.Calls())
	}
}

func TestPreStartFailuresAreNotRun(t *testing.T) {
	// Capture failure: not_run, retryable as network_context_unavailable.
	f := newFixture(t, schedule(2, task.NetworkContextUnavailable), connected(t))
	f.ctx.fail = map[int]bool{1: true}
	rep, err := f.engine.Run(f.run, f.plan)
	f.expect(t, rep, err, nil, []want{{"not_run", "network_context_unavailable", ""}, {"completed", "", "connected"}}, []time.Duration{35 * time.Second})
	if f.runner.Calls() != 1 {
		t.Fatalf("runner invoked for a not_run attempt: %d", f.runner.Calls())
	}

	// Dialer failure: not_run internal_error; the runner is never invoked.
	f = newFixture(t, schedule(1), connected(t))
	f.dialers.FailFor = map[int]bool{1: true}
	rep, err = f.engine.Run(f.run, f.plan)
	f.expect(t, rep, err, nil, []want{{"not_run", "internal_error", ""}}, nil)
	if f.runner.Calls() != 0 {
		t.Fatal("runner invoked without a dialer")
	}
}

// Host cancellation is recorded as cancelled_by_host, prevents every further
// attempt and is written even though the run context has ended.
// requirement-ref: REQ-RESULT-0051
// requirement-ref: REQ-RESULT-0052
func TestHostCancellation(t *testing.T) {
	// Cancelled before the run: attempt 1 is not_run, no other work.
	f := newFixture(t, schedule(2, task.InternalError), connected(t))
	run, cancel := context.WithCancel(context.Background())
	cancel()
	rep, err := f.engine.Run(run, f.plan)
	f.expect(t, rep, err, nil, []want{{"not_run", "cancelled_by_host", ""}}, nil)
	if f.runner.Calls() != 0 || f.ctx.calls != 0 || len(f.dialers.Issued()) != 0 || len(rep.Attempts) != 1 {
		t.Fatal("cancelled run did work")
	}

	// Cancelled after the test started: incomplete and never retried.
	f = newFixture(t, schedule(2, task.InternalError), step{err: measurement.ErrRunner, cancel: true}, connected(t))
	rep, err = f.engine.Run(f.run, f.plan)
	f.expect(t, rep, err, nil, []want{{"incomplete", "cancelled_by_host", ""}}, nil)
	if f.runner.Calls() != 1 || len(rep.Attempts) != 1 {
		t.Fatalf("runner calls = %d, report = %d", f.runner.Calls(), len(rep.Attempts))
	}

	// Cancelled before the test started: not_run without Network Context.
	f = newFixture(t, schedule(1), step{err: measurement.ErrRunner, cancel: true, unbegun: true})
	rep, err = f.engine.Run(f.run, f.plan)
	f.expect(t, rep, err, nil, []want{{"not_run", "cancelled_by_host", ""}}, nil)

	// Cancelled after the outcome was recorded, while connections close:
	// the attempt stays completed and no retry follows.
	done := connected(t)
	done.cancel = true
	f = newFixture(t, schedule(2, task.InternalError), done, connected(t))
	rep, err = f.engine.Run(f.run, f.plan)
	f.expect(t, rep, err, nil, []want{{"completed", "", "connected"}}, nil)

	// Cancelled during the retry backoff: the first result stays, and the
	// retry attempt never starts.
	f = newFixture(t, schedule(2, task.InternalError), failure(), connected(t))
	run, cancel = context.WithCancel(context.Background())
	f.engine, _ = execution.NewEngine(newRegistry(t, f.runner), cancelOnWait{f.clock, cancel}, fixedJitter(5*time.Second), f.store)
	rep, err = f.engine.Run(run, f.plan)
	f.expect(t, rep, err, nil, []want{{"incomplete", "internal_error", ""}}, nil)
	if f.runner.Calls() != 1 || f.ctx.calls != 1 {
		t.Fatalf("retry after cancelled backoff: runner %d, captures %d", f.runner.Calls(), f.ctx.calls)
	}
}

// cancelOnWait cancels the run as soon as the engine starts a backoff wait.
type cancelOnWait struct {
	*testkit.WaitClock
	cancel context.CancelFunc
}

func (c cancelOnWait) Wait(ctx context.Context, d time.Duration) error {
	c.cancel()
	return c.WaitClock.Wait(ctx, d)
}

func TestOutputPortFailuresStopTheRun(t *testing.T) {
	for _, tc := range []struct {
		name    string
		fault   output.Fault
		durable int
	}{
		{"before write", output.BeforeWrite, 0},
		{"uncertain commit", output.AfterWrite, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t, schedule(2, task.InternalError), failure(), connected(t))
			f.store.FailAt(1, tc.fault)
			rep, err := f.engine.Run(f.run, f.plan)
			if !errors.Is(err, execution.ErrOutput) || err.Error() != execution.ErrOutput.Error() {
				t.Fatalf("Run() error = %v", err)
			}
			// No retry after an output failure, and the write is not repeated.
			if f.runner.Calls() != 1 || f.store.Calls() != 1 || len(f.store.Records()) != tc.durable || len(rep.Attempts) != 0 || len(f.clock.Waits()) != 0 {
				t.Fatalf("runner = %d, store = %d, durable = %d, report = %d", f.runner.Calls(), f.store.Calls(), len(f.store.Records()), len(rep.Attempts))
			}
		})
	}
	// A failure on the second attempt keeps the first attempt's report.
	f := newFixture(t, schedule(2, task.InternalError), failure(), connected(t))
	f.store.FailAt(2, output.BeforeWrite)
	rep, err := f.engine.Run(f.run, f.plan)
	if !errors.Is(err, execution.ErrOutput) || len(rep.Attempts) != 1 || len(f.store.Records()) != 1 {
		t.Fatalf("Run() = %+v, %v", rep, err)
	}
}

func TestRunFailsClosedBeforeAnyWork(t *testing.T) {
	cases := map[string]struct {
		mutate func(*fixture)
		err    error
	}{
		"missing runner": {func(f *fixture) {
			reg, _ := registry.New([]registry.Entry{{Test: tcp01, Codec: fixtureCodec{}, BaselineRisk: task.LowRisk, Targets: noTargets{}, Proxy: registry.NoProxy{}}})
			f.engine, _ = execution.NewEngine(reg, f.clock, fixedJitter(0), f.store)
		}, execution.ErrNoRunner},
		"unregistered version": {func(f *fixture) { f.plan.Test.Version = "0.2.0" }, execution.ErrNoRunner},
		"missing parameters":   {func(f *fixture) { f.plan.Parameters = nil }, execution.ErrPlan},
		"missing task ID":      {func(f *fixture) { f.plan.TaskID = "" }, execution.ErrPlan},
		"missing result":       {func(f *fixture) { f.plan.Result = nil }, execution.ErrPlan},
		"missing context":      {func(f *fixture) { f.plan.Context = nil }, execution.ErrPlan},
		"missing dialers":      {func(f *fixture) { f.plan.Dialers = nil }, execution.ErrPlan},
		"zero attempts":        {func(f *fixture) { f.plan.Schedule.Retry.MaxAttempts = 0 }, execution.ErrPlan},
		"above registered max": {func(f *fixture) { f.plan.Schedule.Retry.MaxAttempts = 3 }, execution.ErrPlan},
		"not yet due":          {func(f *fixture) { f.plan.Conditions.DueAt = now.Add(time.Second) }, execution.ErrNotEligible},
		"constraints unmet":    {func(f *fixture) { f.plan.Conditions.ConstraintsSatisfied = false }, execution.ErrNotEligible},
		"runs exhausted":       {func(f *fixture) { f.plan.Conditions.CompletedRuns = 1 }, execution.ErrNotEligible},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t, schedule(1), connected(t))
			tc.mutate(f)
			rep, err := f.engine.Run(f.run, f.plan)
			if !errors.Is(err, tc.err) || rep.RunID != "" || len(rep.Attempts) != 0 {
				t.Fatalf("Run() = %+v, %v; want %v", rep, err, tc.err)
			}
			if f.runner.Calls() != 0 || f.store.Calls() != 0 || f.ctx.calls != 0 || len(f.dialers.Issued()) != 0 {
				t.Fatal("rejected run did work")
			}
		})
	}
	var nilEngine *execution.Engine
	if _, err := nilEngine.Run(context.Background(), execution.Plan{}); !errors.Is(err, execution.ErrEngine) {
		t.Fatal("nil engine ran")
	}
	reg := newRegistry(t, &scripted{})
	clock := testkit.NewWaitClock(now)
	for _, args := range [][4]bool{{false, true, true, true}, {true, false, true, true}, {true, true, false, true}, {true, true, true, false}} {
		var (
			r *registry.Registry
			c execution.Clock
			j execution.Jitter
			o execution.Output
		)
		if args[0] {
			r = reg
		}
		if args[1] {
			c = clock
		}
		if args[2] {
			j = fixedJitter(0)
		}
		if args[3] {
			o = output.New()
		}
		if e, err := execution.NewEngine(r, c, j, o); e != nil || !errors.Is(err, execution.ErrEngine) {
			t.Fatalf("NewEngine(%v) = %v, %v", args, e, err)
		}
	}
}

// Real loopback TCP I/O through the attempt-scoped dialer. This exercises
// sockets for later cancellation and budget work; it is not TCP-01 evidence.
func TestLoopbackRunnerOnRealSockets(t *testing.T) {
	srv, err := loopback.Listen()
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	store := output.New()
	dialers := &loopback.Dialers{}
	engine, err := execution.NewEngine(newRegistry(t, loopback.Runner{}), execution.SystemClock{}, execution.LocalJitter{}, store)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	maxRuns := 1
	expires := task.TimestampFromTime(start.Add(time.Hour))
	plan := execution.Plan{TaskID: "campaign-1", Test: tcp01, Parameters: loopback.Parameters{Address: srv.Address()},
		Schedule: task.Schedule{NotAfter: &expires, Repeat: task.Repeat{Mode: "none"}, MaxRuns: &maxRuns, Retry: task.Retry{MaxAttempts: 1}}, Deadline: attemptDeadline,
		Conditions:     validation.StartContext{DueAt: start, Supported: true, ConstraintsSatisfied: true, SafetySatisfied: true},
		Implementation: implementation, Result: decoder(t), Context: &contextSource{t: t}, Dialers: dialers}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	rep, err := engine.Run(ctx, plan)
	if err != nil || len(rep.Attempts) != 1 || rep.Attempts[0].Attempt.Phase() != execution.Completed {
		t.Fatalf("Run() = %+v, %v", rep, err)
	}
	usage := rep.Attempts[0].Usage
	if usage != (measurement.Usage{BytesSent: len(loopback.Payload), BytesReceived: len(loopback.Payload), Connections: 1}) {
		t.Fatalf("usage = %+v", usage)
	}
	if srv.Accepted() != 1 || len(dialers.Issued()) != 1 || dialers.Issued()[0].Dials() != 1 {
		t.Fatalf("accepted = %d, dialers = %d", srv.Accepted(), len(dialers.Issued()))
	}
	recs := store.Records()
	if len(recs) != 1 || !strings.Contains(string(recs[0]), `"outcome":"connected"`) || strings.Contains(string(recs[0]), srv.Address()) {
		t.Fatalf("record = %s", recs)
	}

	// A closed port yields a terminal connect-failure observation, not a retry.
	closed := srv.Address()
	if err := srv.Close(); err != nil {
		t.Fatal(err)
	}
	plan.Parameters = loopback.Parameters{Address: closed}
	rep, err = engine.Run(ctx, plan)
	if err != nil || len(rep.Attempts) != 1 || rep.Attempts[0].Attempt.Phase() != execution.Completed {
		t.Fatalf("Run() on closed port = %+v, %v", rep, err)
	}
	if recs := store.Records(); len(recs) != 2 || !strings.Contains(string(recs[1]), `"outcome":"socket_error"`) {
		t.Fatalf("records = %s", recs)
	}
}

// scriptedPreconditions answers each attempt-start check in turn.
type scriptedPreconditions struct {
	answers []bool
	calls   int
}

func (p *scriptedPreconditions) Satisfied(context.Context) bool {
	p.calls++
	return p.answers[p.calls-1]
}

// Preconditions are evaluated before every attempt: unsatisfied before the
// first, the run fails closed with no result; unsatisfied before a retry, the
// run ends with the first attempt's result unchanged.
func TestPreconditionsAreRecheckedBeforeEachAttempt(t *testing.T) {
	f := newFixture(t, schedule(2, task.InternalError), failure(), connected(t))
	f.plan.Preconditions = &scriptedPreconditions{answers: []bool{false}}
	rep, err := f.engine.Run(f.run, f.plan)
	f.expect(t, rep, err, execution.ErrNotEligible, nil, nil)
	if f.runner.Calls() != 0 || f.ctx.calls != 0 {
		t.Fatal("work before an unsatisfied first precondition")
	}

	f = newFixture(t, schedule(2, task.InternalError), failure(), connected(t))
	p := &scriptedPreconditions{answers: []bool{true, false}}
	f.plan.Preconditions = p
	rep, err = f.engine.Run(f.run, f.plan)
	f.expect(t, rep, err, nil, []want{{"incomplete", "internal_error", ""}}, []time.Duration{35 * time.Second})
	if f.runner.Calls() != 1 || p.calls != 2 {
		t.Fatalf("retry ran with unsatisfied preconditions: runner %d, checks %d", f.runner.Calls(), p.calls)
	}

	f = newFixture(t, schedule(2, task.InternalError), failure(), connected(t))
	p = &scriptedPreconditions{answers: []bool{true, true}}
	f.plan.Preconditions = p
	rep, err = f.engine.Run(f.run, f.plan)
	f.expect(t, rep, err, nil, []want{{"incomplete", "internal_error", ""}, {"completed", "", "connected"}}, []time.Duration{35 * time.Second})
	if p.calls != 2 {
		t.Fatal("preconditions not checked before each attempt")
	}
}
