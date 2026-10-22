package barometer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/admission"
	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/persistence"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/registry"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/taskdecode"
	"github.com/tv-rain/barometer-go/internal/testkit"
	"github.com/tv-rain/barometer-go/internal/testkit/capability"
	"github.com/tv-rain/barometer-go/internal/testkit/ledger"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
	"github.com/tv-rain/barometer-go/internal/testkit/output"
)

// Synthetic run fixtures only. The TCP-01 identifier is used so result
// construction applies its output shape; the runner performs no I/O and is
// not TCP-01 conformance evidence.
var tcp01 = task.TestVersion{ID: "TCP-01", Version: "0.1.0"}

// runFakes are the synthetic run adapters of one harness. The run clock reads
// the same instants as the admission clock but carries a monotonic reading,
// which result timing requires; advance moves both.
type runFakes struct {
	clock   *testkit.WaitClock
	output  *output.Store
	ledger  *ledger.Ledger
	paths   *fakePaths
	runner  *runRunner
	consent *capability.Consent
	device  *capability.Device
}

// fixedJitter always draws the same offset, capped at the maximum.
type fixedJitter time.Duration

func (j fixedJitter) Jitter(max time.Duration) time.Duration { return min(time.Duration(j), max) }

func (h *harness) bindRunAdapters(t *testing.T) {
	t.Helper()
	base := time.Now()
	h.run = runFakes{clock: testkit.NewWaitClock(base.Add(h.clock.Clock.Now().Sub(base))), output: output.New(), ledger: ledger.New(),
		paths: &fakePaths{t: t}, runner: newRunRunner(), consent: capability.NewConsent(platform.ConsentValid), device: capability.NewDevice(platform.DeviceState{})}
	h.adapters.run = runAdapters{clock: h.run.clock, jitter: fixedJitter(0), output: h.run.output, ledger: h.run.ledger, paths: h.run.paths,
		consent: h.run.consent, device: h.run.device, zone: capability.Zone{Loc: time.UTC}}
}

func (h *harness) advance(d time.Duration) {
	_ = h.clock.Advance(d)
	_ = h.run.clock.Advance(d)
}

// runRunner is a scripted synthetic runner. While held, a call waits for
// release or for its context.
type runRunner struct {
	mu      sync.Mutex
	calls   int
	hold    chan struct{}
	fail    map[int]bool
	during  func(call int)
	started chan struct{}
	stuck   chan struct{} // when set, a call ignores its context until closed
}

func newRunRunner() *runRunner {
	return &runRunner{fail: map[int]bool{}, started: make(chan struct{}, 64)}
}

func (r *runRunner) holdCalls() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hold = make(chan struct{})
}

func (r *runRunner) releaseCalls() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.hold != nil {
		close(r.hold)
		r.hold = nil
	}
}

func (r *runRunner) Calls() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls
}

func (r *runRunner) Run(ctx context.Context, in measurement.Input) (measurement.Output, error) {
	r.mu.Lock()
	r.calls++
	n, hold, fail, during, stuck := r.calls, r.hold, r.fail[r.calls], r.during, r.stuck
	r.mu.Unlock()
	r.started <- struct{}{}
	in.Start()
	if during != nil {
		during(n)
	}
	if stuck != nil {
		<-stuck // ignores cancellation, as a runner breaking its contract would
		if ctx.Err() != nil {
			return measurement.Output{}, measurement.ErrRunner
		}
	}
	if hold != nil {
		select {
		case <-hold:
		case <-ctx.Done():
			return measurement.Output{}, ctx.Err()
		}
	}
	if fail || in.Dialer == nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	o, err := observation.DecodeProtocol(observation.Contract, []byte(`{"type":"tcp_connect","status":"succeeded","duration_ms":0,"details":{"address_family":"ipv4"}}`), observation.ProtocolLimits{})
	if err != nil {
		return measurement.Output{}, err
	}
	return measurement.Output{Outcome: "connected", Observations: []observation.Observation{o}, Usage: measurement.Usage{Connections: 1}}, nil
}

// fakePaths binds a direct path with a fixed synthetic Network Context.
type fakePaths struct {
	t     *testing.T
	mu    sync.Mutex
	calls int
	fail  bool
	panic bool
}

var implementation = model.Implementation{LibraryVersion: "0.1.0", Platform: "android"}

func (p *fakePaths) Calls() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.calls
}

func (p *fakePaths) direct(_ context.Context, r pathRequest) (pathPlan, error) {
	p.mu.Lock()
	p.calls++
	fail, panics := p.fail, p.panic
	p.mu.Unlock()
	if panics {
		panic("synthetic path panic")
	}
	if fail {
		return pathPlan{}, errors.New("synthetic path failure")
	}
	cd, err := observation.NewContextDecoder(observation.Contract, observation.ContextApprovals{})
	if err != nil {
		return pathPlan{}, err
	}
	path, err := observation.NewPath(1, 0, "direct", "")
	if err != nil {
		return pathPlan{}, err
	}
	d, err := result.New(result.Binding{TaskID: r.TaskID, Output: result.Output{Test: model.Test{ID: r.Test.ID, Version: r.Test.Version, ResultVersion: result.Contract}},
		MaxAttempts: r.MaxAttempts, MaxDurationMS: r.MaxDurationMS, Implementation: implementation, ContextDecoder: cd, Path: path, AccessGeneration: 1})
	if err != nil {
		return pathPlan{}, err
	}
	return pathPlan{Implementation: implementation, Result: d, Context: contextSource{cd: cd, path: path}, Dialers: &loopback.Dialers{}}, nil
}

type contextSource struct {
	cd   *observation.ContextDecoder
	path observation.Path
}

func (c contextSource) Capture(context.Context) (observation.NetworkContext, error) {
	return c.cd.Decode([]byte(`{"access":{"network_type":"wifi","vpn_state":"unknown","ipv4_available":"yes","ipv6_available":"unknown","validated":"unknown","network_changed":false},"egress":{"status":"unavailable","lookup_source":"none","network_path":"direct"}}`), c.path)
}

// tcpCodec is the SYN-01 codec under the TCP-01 identifier.
type tcpCodec struct{ synCodec }

func (tcpCodec) Definition() taskdecode.Definition {
	d := synCodec{}.Definition()
	d.TestID = tcp01.ID
	return d
}

type tcpTargets struct{}

func (tcpTargets) Targets(task.Parameters) ([]registry.Target, bool) {
	return []registry.Target{{Role: task.MeasurementTarget, Tuple: task.TargetTuple{Hostname: "reg-a.example.test"}}}, true
}

// runningHarness admits TCP-01 tasks under a generation-2 policy and runs them
// with the scripted runner.
func runningHarness(t *testing.T) *harness {
	t.Helper()
	h := newHarness(t)
	in := synPolicyInput(2, "reg-a", func(p *task.Permissions) { p.Tests = []task.TestVersion{tcp01} })
	in.Targets[0].Test = tcp01
	policy, err := task.NewLocalPolicy(in)
	if err != nil {
		t.Fatal("synthetic policy rejected")
	}
	reg, err := registry.New([]registry.Entry{{Test: tcp01, Codec: tcpCodec{}, BaselineRisk: task.LowRisk, Targets: tcpTargets{}, Proxy: registry.NoProxy{}, Runner: h.run.runner}})
	if err != nil {
		t.Fatal("synthetic registry rejected")
	}
	h.adapters.policy, h.adapters.registry = policy, reg
	return h
}

// runTask is a TCP-01 task valid for a day. schedule and execution are the
// members of those objects without braces; consent sets the task's own
// consent requirement.
func runTask(id, schedule, execution string, consent bool) string {
	if schedule != "" {
		schedule = "," + schedule
	}
	if execution != "" {
		execution = "," + execution
	}
	return fmt.Sprintf(`{"schema_version":"0.1.0","task_id":%q,"test_id":"TCP-01","test_version":"0.1.0","input":{},"options":{},"schedule":{"not_after":%q%s},"execution":{"limits":{}%s},"safety":{"risk_level":"low","requires_user_consent":%t}}`,
		id, epoch.Add(24*time.Hour).Format(time.RFC3339), schedule, execution, consent)
}

// admit admits an embedded task and returns its handle.
func (h *harness) admit(t *testing.T, c *Client, taskJSON string) string {
	t.Helper()
	th, err := c.AdmitTask(taskJSON, h.embedded(t, taskJSON))
	if err != nil {
		t.Fatal("admission rejected:", err)
	}
	return th
}

// waitTerminal polls RunState until the run is terminal. The bound is a
// liveness guard, not a timing assertion.
func waitTerminal(t *testing.T, c *Client, rh string) {
	t.Helper()
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); time.Sleep(time.Millisecond) {
		if s, err := c.RunState(rh); err != nil || s == runTerminal {
			if err != nil {
				t.Fatal("run state:", err)
			}
			return
		}
	}
	t.Fatal("run did not become terminal")
}

func waitStarted(t *testing.T, r *runRunner) {
	t.Helper()
	select {
	case <-r.started:
	case <-time.After(5 * time.Second):
		t.Fatal("runner was not invoked")
	}
}

type resultRecord struct {
	RunID     string `json:"run_id"`
	Attempt   int    `json:"attempt"`
	Execution struct {
		State  string `json:"state"`
		Reason string `json:"reason"`
	} `json:"execution"`
}

func records(t *testing.T, s *output.Store) []resultRecord {
	t.Helper()
	var out []resultRecord
	for _, b := range s.Records() {
		var r resultRecord
		if err := json.Unmarshal(b, &r); err != nil {
			t.Fatal(err)
		}
		out = append(out, r)
	}
	return out
}

func key(id string) persistence.ReplayKey {
	return persistence.ReplayKey{IssuerID: "embedded-issuer", TaskID: id}
}

// StartRun reserves the run in the ledger, starts it on the engine and
// returns a run handle at once; RunState reports active, then terminal.
func TestStartRunRunsOneDueRun(t *testing.T) {
	h := runningHarness(t)
	c := h.initialized(t)
	th := h.admit(t, c, runTask("run-1", "", "", false))
	h.run.runner.holdCalls()
	rh, err := c.StartRun(th)
	if err != nil || !strings.HasPrefix(rh, runHandlePrefix) {
		t.Fatalf("start: %q %v", rh, err)
	}
	waitStarted(t, h.run.runner)
	if s, err := c.RunState(rh); s != runActive || err != nil {
		t.Fatalf("held run: %q %v", s, err)
	}
	if rec, ok := h.run.ledger.Record(key("run-1")); !ok || rec.Runs != 1 || rec.Next != 1 {
		t.Fatalf("run not recorded before it ran: %+v", rec)
	}
	h.run.runner.releaseCalls()
	waitTerminal(t, c, rh)
	got := records(t, h.run.output)
	if len(got) != 1 || got[0].Attempt != 1 || got[0].Execution.State != "completed" || h.run.runner.Calls() != 1 {
		t.Fatalf("results: %+v", got)
	}
	// The run is the task's only run: a later start resolves to it.
	h.advance(time.Hour)
	if again, err := c.StartRun(th); again != rh || err != nil {
		t.Fatal("second start of a one-time task:", codeOf(err))
	}
	if h.run.runner.Calls() != 1 || len(h.run.output.Records()) != 1 {
		t.Fatal("repeated start did work")
	}
	// Once the terminal record is reclaimed, the task is not due.
	c.rt.runs.remove(rh)
	if _, err := c.StartRun(th); codeOf(err) != "run_not_due" {
		t.Fatal("second run of a one-time task:", codeOf(err))
	}
}

// Concurrent starts for the same due occurrence share one run handle and one
// run, whether the run is active or already terminal.
func TestStartRunDuplicatesShareOneRun(t *testing.T) {
	h := runningHarness(t)
	c := h.initialized(t)
	th := h.admit(t, c, runTask("dup-1", "", "", false))
	h.run.runner.holdCalls()
	const callers = 16
	handles := make([]string, callers)
	var wg sync.WaitGroup
	for i := range callers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rh, err := c.StartRun(th)
			if err != nil {
				t.Error("duplicate start:", err)
			}
			handles[i] = rh
		}()
	}
	wg.Wait()
	for _, rh := range handles {
		if rh != handles[0] {
			t.Fatal("duplicate starts returned different handles")
		}
	}
	h.run.runner.releaseCalls()
	waitTerminal(t, c, handles[0])
	if again, err := c.StartRun(th); again != handles[0] || err != nil {
		t.Fatal("terminal duplicate:", codeOf(err))
	}
	if rec, _ := h.run.ledger.Record(key("dup-1")); rec.Runs != 1 || h.run.runner.Calls() != 1 || len(h.run.output.Records()) != 1 || len(c.rt.runs.entries) != 1 {
		t.Fatalf("duplicates allocated extra runs: %+v", rec)
	}
}

// Schedule boundaries, evaluated on the injected clock: starts_at, not_after
// and a local time window (start inclusive, end exclusive, midnight crossing).
func TestStartRunScheduleBoundaries(t *testing.T) {
	at := func(d time.Duration) string { return epoch.Add(d).Format(time.RFC3339) }
	cases := []struct {
		name, schedule string
		offset         time.Duration
		due            bool
	}{
		{"before starts_at", `"starts_at":"` + at(time.Hour) + `"`, time.Hour - time.Second, false},
		{"at starts_at", `"starts_at":"` + at(time.Hour) + `"`, time.Hour, true},
		{"after starts_at", `"starts_at":"` + at(time.Hour) + `"`, time.Hour + time.Second, true},
		{"before not_after", "", 24*time.Hour - time.Second, true},
		{"at not_after", "", 24 * time.Hour, false},
		{"window 11:59", `"time_window":{"ranges":[{"start":"12:00","end":"18:00"}]}`, 23*time.Hour + 59*time.Minute, false},
		{"window 12:00", `"time_window":{"ranges":[{"start":"12:00","end":"18:00"}]}`, 0, true},
		{"window 17:59", `"time_window":{"ranges":[{"start":"12:00","end":"18:00"}]}`, 5*time.Hour + 59*time.Minute, true},
		{"window 18:00", `"time_window":{"ranges":[{"start":"12:00","end":"18:00"}]}`, 6 * time.Hour, false},
		{"night window 21:59", `"time_window":{"ranges":[{"start":"22:00","end":"06:00"}]}`, 9*time.Hour + 59*time.Minute, false},
		{"night window 22:00", `"time_window":{"ranges":[{"start":"22:00","end":"06:00"}]}`, 10 * time.Hour, true},
		{"night window 05:59", `"time_window":{"ranges":[{"start":"22:00","end":"06:00"}]}`, 17*time.Hour + 59*time.Minute, true},
		{"night window 06:00", `"time_window":{"ranges":[{"start":"22:00","end":"06:00"}]}`, 18 * time.Hour, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := runningHarness(t)
			c := h.initialized(t)
			th := h.admit(t, c, runTask("sched-1", tc.schedule, "", false))
			h.advance(tc.offset)
			rh, err := c.StartRun(th)
			if tc.due != (err == nil) || !tc.due && codeOf(err) != "run_not_due" {
				t.Fatalf("got %s, want due=%t", codeOf(err), tc.due)
			}
			if tc.due {
				waitTerminal(t, c, rh)
			} else if rec, _ := h.run.ledger.Record(key("sched-1")); rec.Runs != 0 || h.run.runner.Calls() != 0 {
				t.Fatal("refused start reserved a run")
			}
		})
	}
	t.Run("window without a time zone", func(t *testing.T) {
		h := runningHarness(t)
		h.adapters.run.zone = capability.Zone{}
		c := h.initialized(t)
		th := h.admit(t, c, runTask("sched-2", `"time_window":{"ranges":[{"start":"12:00","end":"18:00"}]}`, "", false))
		if _, err := c.StartRun(th); codeOf(err) != "run_not_due" {
			t.Fatal("unknown time zone satisfied a window:", codeOf(err))
		}
	})
}

// interval is a repeated schedule: a 6-hour interval with 300 seconds of
// jitter and the given max_runs.
func interval(maxRuns int) string {
	return fmt.Sprintf(`"repeat":{"mode":"interval","interval":21600,"jitter":300},"limits":{"max_runs":%d}`, maxRuns)
}

// Interval repetition follows the interim rule of GO-SPEC-ISSUE-011:
// occurrence k is due at anchor + k*interval + jitter, a start before that is
// not due, and missed occurrences are skipped rather than run late.
func TestStartRunIntervalOccurrences(t *testing.T) {
	h := runningHarness(t)
	h.adapters.run.jitter = fixedJitter(time.Minute)
	c := h.initialized(t)
	th := h.admit(t, c, runTask("rep-1", interval(4), "", false))
	start := func(want string) string {
		t.Helper()
		rh, err := c.StartRun(th)
		if codeOf(err) != want {
			t.Fatalf("got %s, want %s", codeOf(err), want)
		}
		if err == nil {
			waitTerminal(t, c, rh)
		}
		return rh
	}
	first := start("ok") // occurrence 0, at the anchor, without jitter
	h.advance(6 * time.Hour)
	start("run_not_due") // occurrence 1 is due one minute later
	h.advance(time.Minute - time.Second)
	start("run_not_due")
	h.advance(time.Second)
	if second := start("ok"); second == first {
		t.Fatal("a new occurrence reused the previous run handle")
	}
	// Occurrence 2 is missed; at occurrence 3 only occurrence 3 runs.
	h.advance(12*time.Hour + time.Minute)
	start("ok")
	start("ok") // a repeated start for occurrence 3 returns its handle
	rec, _ := h.run.ledger.Record(key("rep-1"))
	if rec.Runs != 3 || rec.Next != 4 || h.run.runner.Calls() != 3 {
		t.Fatalf("accounting after a skipped occurrence: %+v", rec)
	}
	// max_runs ends the schedule even though not_after is later.
	h.advance(6*time.Hour + time.Minute)
	start("run_not_due")
}

// The run budget at limit -1, limit and limit +1: max_runs runs start and no
// more, whatever the host calls.
func TestStartRunBudget(t *testing.T) {
	for _, maxRuns := range []int{1, 2, 3} {
		t.Run(fmt.Sprint(maxRuns), func(t *testing.T) {
			h := runningHarness(t)
			c := h.initialized(t)
			th := h.admit(t, c, runTask("budget-1", interval(maxRuns), "", false))
			started := 0
			for range maxRuns + 1 {
				if rh, err := c.StartRun(th); err == nil {
					started++
					waitTerminal(t, c, rh)
				} else if codeOf(err) != "run_not_due" {
					t.Fatal(err)
				}
				h.advance(6*time.Hour + 5*time.Minute)
			}
			if rec, _ := h.run.ledger.Record(key("budget-1")); started != maxRuns || rec.Runs != maxRuns || h.run.runner.Calls() != maxRuns {
				t.Fatalf("started %d runs, ledger %+v", started, rec)
			}
		})
	}
}

// Consent is checked at attempt start only for a task whose consent
// requirement is recorded; anything but valid consent starts no attempt.
func TestStartRunConsent(t *testing.T) {
	cases := map[string]struct {
		edit func(h *harness)
		due  bool
	}{
		"valid":       {func(h *harness) {}, true},
		"absent":      {func(h *harness) { h.run.consent.Set(platform.ConsentAbsent) }, false},
		"ended":       {func(h *harness) { h.run.consent.Set(platform.ConsentEnded) }, false},
		"unavailable": {func(h *harness) { h.run.consent.Set(platform.ConsentUnavailable) }, false},
		"error":       {func(h *harness) { h.run.consent.Fail() }, false},
		"no provider": {func(h *harness) { h.adapters.run.consent = nil }, false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			h := runningHarness(t)
			tc.edit(h)
			c := h.initialized(t)
			th := h.admit(t, c, runTask("consent-1", "", "", true))
			rh, err := c.StartRun(th)
			if tc.due != (err == nil) || !tc.due && codeOf(err) != "run_not_due" {
				t.Fatalf("got %s", codeOf(err))
			}
			if tc.due {
				waitTerminal(t, c, rh)
			} else if h.run.runner.Calls() != 0 || len(h.run.output.Records()) != 0 {
				t.Fatal("an attempt or result without valid consent")
			}
		})
	}
	t.Run("not required", func(t *testing.T) {
		h := runningHarness(t)
		h.run.consent.Set(platform.ConsentAbsent)
		c := h.initialized(t)
		rh, err := c.StartRun(h.admit(t, c, runTask("consent-2", "", "", false)))
		if err != nil {
			t.Fatal(err)
		}
		waitTerminal(t, c, rh)
		if h.run.consent.Calls() != 0 {
			t.Fatal("consent queried for a task without the requirement")
		}
	})
}

// Consent withdrawn after the first attempt prevents the retry: the run ends
// with the first attempt's result only.
func TestConsentIsRecheckedBeforeARetry(t *testing.T) {
	h := runningHarness(t)
	h.run.runner.fail[1] = true
	h.run.runner.during = func(n int) {
		if n == 1 {
			h.run.consent.Set(platform.ConsentEnded)
		}
	}
	c := h.initialized(t)
	retry := `"retry":{"max_attempts":2,"retry_on":["internal_error"],"backoff":{"mode":"fixed","initial_delay_ms":30000,"max_delay_ms":30000}}`
	rh, err := c.StartRun(h.admit(t, c, runTask("consent-3", retry, "", true)))
	if err != nil {
		t.Fatal(err)
	}
	waitTerminal(t, c, rh)
	got := records(t, h.run.output)
	if len(got) != 1 || got[0].Execution.State != "incomplete" || h.run.runner.Calls() != 1 {
		t.Fatalf("retry ran without consent: %+v", got)
	}
}

// Device constraints fail closed: an unknown, unavailable or insufficient
// state never satisfies a requirement.
func TestStartRunDeviceConstraints(t *testing.T) {
	battery, charging := `"device":{"minimum_battery_percent":20}`, `"device":{"allow_charging_only":true}`
	cases := []struct {
		name, execution string
		state           *platform.DeviceState // nil: capability fails
		noProvider, due bool
	}{
		{"battery unknown", battery, &platform.DeviceState{}, false, false},
		{"battery 19", battery, &platform.DeviceState{BatteryPercent: 19, BatteryKnown: true}, false, false},
		{"battery 20", battery, &platform.DeviceState{BatteryPercent: 20, BatteryKnown: true}, false, true},
		{"battery 21", battery, &platform.DeviceState{BatteryPercent: 21, BatteryKnown: true}, false, true},
		{"battery error", battery, nil, false, false},
		{"battery no provider", battery, nil, true, false},
		{"charging unknown", charging, &platform.DeviceState{Charging: true}, false, false},
		{"not charging", charging, &platform.DeviceState{ChargingKnown: true}, false, false},
		{"charging", charging, &platform.DeviceState{Charging: true, ChargingKnown: true}, false, true},
		{"no requirement", "", nil, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := runningHarness(t)
			switch {
			case tc.noProvider:
				h.adapters.run.device = nil
			case tc.state == nil:
				h.run.device.Fail()
			default:
				h.run.device.Set(*tc.state)
			}
			c := h.initialized(t)
			rh, err := c.StartRun(h.admit(t, c, runTask("device-1", "", tc.execution, false)))
			if tc.due != (err == nil) || !tc.due && codeOf(err) != "run_not_due" {
				t.Fatalf("got %s", codeOf(err))
			}
			if tc.due {
				waitTerminal(t, c, rh)
			}
		})
	}
}

// Network conditions this build cannot determine are unsatisfied, so such a
// task never runs and never reaches the path seam.
func TestStartRunNetworkConstraintsAreUnsatisfied(t *testing.T) {
	for name, execution := range map[string]string{
		"allowed transports": `"network":{"allowed_transports":["wifi"]}`,
		"vpn exclude":        `"network":{"vpn_policy":"exclude"}`,
		"vpn require":        `"network":{"vpn_policy":"require"}`,
		"require validated":  `"network":{"require_validated":true}`,
	} {
		t.Run(name, func(t *testing.T) {
			h := runningHarness(t)
			c := h.initialized(t)
			if _, err := c.StartRun(h.admit(t, c, runTask("net-1", "", execution, false))); codeOf(err) != "run_not_due" {
				t.Fatal(codeOf(err))
			}
			if h.run.paths.Calls() != 0 || h.run.runner.Calls() != 0 {
				t.Fatal("unsatisfied network condition reached the path or runner")
			}
		})
	}
}

// An exclusive run starts only alone and blocks every other start while it is
// active; shared runs are bounded by maxActiveRuns.
func TestStartRunConcurrency(t *testing.T) {
	exclusive, shared := `"concurrency":{"mode":"exclusive"}`, `"concurrency":{"mode":"shared"}`
	h := runningHarness(t)
	c := h.initialized(t)
	h.run.runner.holdCalls()
	a, err := c.StartRun(h.admit(t, c, runTask("conc-a", "", shared, false)))
	if err != nil {
		t.Fatal(err)
	}
	x := h.admit(t, c, runTask("conc-x", "", exclusive, false))
	if _, err := c.StartRun(x); codeOf(err) != "run_not_due" {
		t.Fatal("exclusive run started beside an active run:", codeOf(err))
	}
	h.run.runner.releaseCalls()
	waitTerminal(t, c, a)
	h.run.runner.holdCalls()
	xr, err := c.StartRun(x)
	if err != nil {
		t.Fatal("exclusive run alone:", err)
	}
	if _, err := c.StartRun(h.admit(t, c, runTask("conc-b", "", "", false))); codeOf(err) != "run_not_due" {
		t.Fatal("run started beside an exclusive run:", codeOf(err))
	}
	h.run.runner.releaseCalls()
	waitTerminal(t, c, xr)

	h.run.runner.holdCalls()
	var active []string
	for i := range maxActiveRuns {
		rh, err := c.StartRun(h.admit(t, c, runTask(fmt.Sprintf("conc-s%d", i), "", shared, false)))
		if err != nil {
			t.Fatal(err)
		}
		active = append(active, rh)
	}
	if _, err := c.StartRun(h.admit(t, c, runTask("conc-over", "", shared, false))); codeOf(err) != "capacity_exceeded" {
		t.Fatal("active run limit +1:", codeOf(err))
	}
	if rec, ok := h.run.ledger.Record(key("conc-over")); ok && rec.Runs != 0 {
		t.Fatal("refused start reserved a run")
	}
	h.run.runner.releaseCalls()
	for _, rh := range active {
		waitTerminal(t, c, rh)
	}
}

// Pending-output backpressure: a run starts only when the port has room for
// every result it may write, counting the active runs' possible results.
func TestStartRunPendingOutputBackpressure(t *testing.T) {
	retry := `"retry":{"max_attempts":2,"retry_on":["internal_error"],"backoff":{"mode":"fixed","initial_delay_ms":30000,"max_delay_ms":30000}}`
	h := runningHarness(t)
	h.run.output.SetLimit(1)
	c := h.initialized(t)
	if _, err := c.StartRun(h.admit(t, c, runTask("bp-1", retry, "", false))); codeOf(err) != "run_not_due" {
		t.Fatal("two possible results into one free slot:", codeOf(err))
	}
	h.run.output.SetLimit(2)
	h.run.runner.holdCalls()
	rh, err := c.StartRun(h.admit(t, c, runTask("bp-2", "", "", false)))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.StartRun(h.admit(t, c, runTask("bp-3", retry, "", false))); codeOf(err) != "run_not_due" {
		t.Fatal("capacity held by the active run was counted as free:", codeOf(err))
	}
	h.run.runner.releaseCalls()
	waitTerminal(t, c, rh)
	h.run.output.FailCapacity()
	if _, err := c.StartRun(h.admit(t, c, runTask("bp-4", "", "", false))); codeOf(err) != "run_not_due" {
		t.Fatal("unavailable capacity:", codeOf(err))
	}
	if h.run.runner.Calls() != 1 {
		t.Fatal("a refused start ran")
	}
}

// Run-ledger failures fail closed. An uncertain reservation loses the
// occurrence rather than risking an extra run.
func TestStartRunLedgerFailures(t *testing.T) {
	for name, fault := range map[string]ledger.Fault{"lookup": ledger.FailLookup, "before write": ledger.FailBeforeWrite} {
		t.Run(name, func(t *testing.T) {
			h := runningHarness(t)
			c := h.initialized(t)
			th := h.admit(t, c, runTask("ledger-1", "", "", false))
			h.run.ledger.Inject(fault, true)
			if _, err := c.StartRun(th); codeOf(err) != "run_not_due" {
				t.Fatal(codeOf(err))
			}
			h.run.ledger.Inject(fault, false)
			if h.run.runner.Calls() != 0 || len(c.rt.runs.entries) != 0 {
				t.Fatal("a run started without durable accounting")
			}
			rh, err := c.StartRun(th)
			if err != nil {
				t.Fatal("recovered ledger:", err)
			}
			waitTerminal(t, c, rh)
		})
	}
	t.Run("uncertain reservation", func(t *testing.T) {
		h := runningHarness(t)
		c := h.initialized(t)
		th := h.admit(t, c, runTask("ledger-2", `"starts_at":"`+epoch.Add(time.Hour).Format(time.RFC3339)+`"`, "", false))
		// A start before starts_at only creates the record.
		if _, err := c.StartRun(th); codeOf(err) != "run_not_due" {
			t.Fatal(codeOf(err))
		}
		h.advance(time.Hour)
		h.run.ledger.Inject(ledger.FailAfterWrite, true)
		if _, err := c.StartRun(th); codeOf(err) != "run_not_due" {
			t.Fatal("uncertain reservation:", codeOf(err))
		}
		h.run.ledger.Inject(ledger.FailAfterWrite, false)
		if rec, _ := h.run.ledger.Record(key("ledger-2")); rec.Runs != 1 {
			t.Fatal("reservation not durable in the fake")
		}
		if _, err := c.StartRun(th); codeOf(err) != "run_not_due" || h.run.runner.Calls() != 0 || len(c.rt.runs.entries) != 0 {
			t.Fatal("an uncertain reservation started or allowed a run:", codeOf(err))
		}
	})
}

// Unknown, stale and other-instance handles, lifecycle refusals, a missing
// runner, path failures and panics all start nothing.
func TestStartRunRefusals(t *testing.T) {
	h := runningHarness(t)
	c := h.initialized(t)
	th := h.admit(t, c, runTask("ref-1", "", "", false))
	for name, handle := range map[string]string{"unknown": "T.unknown", "empty": "", "run handle": runHandlePrefix + "x"} {
		if _, err := c.StartRun(handle); codeOf(err) != "handle_unavailable" {
			t.Fatal(name, codeOf(err))
		}
		if _, err := c.RunState(handle); codeOf(err) != "handle_unavailable" {
			t.Fatal(name, "state:", codeOf(err))
		}
	}
	if _, err := runningHarness(t).initialized(t).StartRun(th); codeOf(err) != "handle_unavailable" {
		t.Fatal("other instance:", codeOf(err))
	}
	var nilClient *Client
	if _, err := nilClient.StartRun(th); codeOf(err) != "not_initialized" {
		t.Fatal("nil client")
	}
	if _, err := nilClient.RunState("R.x"); codeOf(err) != "not_initialized" {
		t.Fatal("nil client state")
	}
	if _, err := NewClient().StartRun(th); codeOf(err) != "not_initialized" {
		t.Fatal("uninitialized client")
	}
	h.run.paths.fail = true
	if _, err := c.StartRun(th); codeOf(err) != "run_not_due" {
		t.Fatal("path failure:", codeOf(err))
	}
	h.run.paths.fail, h.run.paths.panic = false, true
	if _, err := c.StartRun(th); codeOf(err) != "run_not_due" {
		t.Fatal("path panic:", codeOf(err))
	}
	h.run.paths.panic = false
	if h.run.runner.Calls() != 0 {
		t.Fatal("refused starts ran")
	}

	// SYN-01 is registered without a runner: admitted, but never run.
	s := admittingHarness(t)
	sc := s.initialized(t)
	if _, err := sc.StartRun(s.admit(t, sc, synTask("syn-1", time.Hour))); codeOf(err) != "run_not_due" {
		t.Fatal("missing runner:", codeOf(err))
	}

	rh, err := c.StartRun(th)
	if err != nil {
		t.Fatal(err)
	}
	waitTerminal(t, c, rh)
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := c.StartRun(th); codeOf(err) != "client_closed" {
		t.Fatal("closed client:", codeOf(err))
	}
	if _, err := c.RunState(rh); codeOf(err) != "client_closed" {
		t.Fatal("closed client state:", codeOf(err))
	}
}

// Run accounting survives reinitialization: a redelivered one-time task gets
// a new task handle, its old run handle is stale and it does not run again.
func TestRunAccountingSurvivesReinitialization(t *testing.T) {
	h := runningHarness(t)
	c := h.initialized(t)
	taskJSON := runTask("reinit-1", "", "", false)
	rh, err := c.StartRun(h.admit(t, c, taskJSON))
	if err != nil {
		t.Fatal(err)
	}
	waitTerminal(t, c, rh)
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	// A new instance over the same replay store and run ledger; close
	// dropped the previous instance's delivery store.
	var err2 error
	if h.deliveries, err2 = admission.NewDeliveryStore(h.clock, admission.DeliveryLimits{Capacity: 8, Lifetime: time.Minute}, 2); err2 != nil {
		t.Fatal(err2)
	}
	h.adapters.deliveries = h.deliveries
	c2 := h.initialized(t)
	th := h.admit(t, c2, taskJSON)
	if _, err := c2.RunState(rh); codeOf(err) != "handle_unavailable" {
		t.Fatal("stale run handle:", codeOf(err))
	}
	if _, err := c2.StartRun(th); codeOf(err) != "run_not_due" {
		t.Fatal("reinitialization reset run accounting:", codeOf(err))
	}
	if h.run.runner.Calls() != 1 {
		t.Fatal("task ran twice")
	}
}

// Lifecycle transitions serialize with run reservation: a transition cancels
// the generation context, waits for active runs and invalidates run handles.
func TestRunsAndLifecycleTransitions(t *testing.T) {
	for name, transition := range map[string]func(*Client) error{"close": (*Client).Close, "clear": (*Client).ClearMeasurementData} {
		t.Run(name, func(t *testing.T) {
			h := runningHarness(t)
			h.adapters.clearStorage = func(context.Context) error { return nil }
			c := h.initialized(t)
			h.run.runner.holdCalls()
			rh, err := c.StartRun(h.admit(t, c, runTask("life-1", "", "", false)))
			if err != nil {
				t.Fatal(err)
			}
			waitStarted(t, h.run.runner)
			if err := transition(c); err != nil {
				t.Fatal("transition with an active run:", err)
			}
			c.mu.Lock()
			active := c.active
			c.mu.Unlock()
			if active != 0 {
				t.Fatal("transition returned while a run held its reservation")
			}
			if _, err := c.RunState(rh); err == nil {
				t.Fatal("run handle survived the transition")
			}
		})
	}
	t.Run("concurrent starts and close", func(t *testing.T) {
		h := runningHarness(t)
		c := h.initialized(t)
		var handles []string
		for i := range 8 {
			handles = append(handles, h.admit(t, c, runTask(fmt.Sprintf("race-%d", i), "", "", false)))
		}
		var wg sync.WaitGroup
		for _, th := range handles {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := c.StartRun(th)
				switch codeOf(err) {
				case "ok", "run_not_due", "capacity_exceeded", "transition_in_progress", "client_closed", "handle_unavailable":
				default:
					t.Error("unexpected start result:", codeOf(err))
				}
			}()
		}
		closeErr := c.Close()
		wg.Wait()
		if closeErr != nil && codeOf(closeErr) != "transition_in_progress" {
			t.Fatal("close:", closeErr)
		}
		if err := c.Close(); err != nil {
			t.Fatal("retried close:", err)
		}
		// No run outlives close, and none counts more than once.
		if c.active != 0 || h.run.runner.Calls() > len(handles) {
			t.Fatal("a run outlived close or ran twice")
		}
		for i := range handles {
			if rec, ok := h.run.ledger.Record(key(fmt.Sprintf("race-%d", i))); ok && rec.Runs > 1 {
				t.Fatal("extra run")
			}
		}
	})
}

// Errors from the run entry points carry only fixed codes.
func TestRunErrorsCarryOnlyFixedCodes(t *testing.T) {
	const canary = "CANARY-run-9c1d"
	h := runningHarness(t)
	c := h.initialized(t)
	th := h.admit(t, c, runTask("canary-1", "", `"network":{"require_validated":true}`, false))
	var errs []error
	for _, call := range []func() error{
		func() error { _, err := c.StartRun(canary); return err },
		func() error { _, err := c.StartRun(th + canary); return err },
		func() error { _, err := c.RunState(canary); return err },
		func() error { _, err := c.StartRun(th); return err },
	} {
		errs = append(errs, call())
	}
	for _, err := range errs {
		var op opError
		if err == nil || !errors.As(err, &op) || errors.Unwrap(err) != nil {
			t.Fatal("not a fixed code:", err)
		}
		for _, f := range []string{"%v", "%+v", "%#v", "%s"} {
			if out := fmt.Sprintf(f, err); strings.Contains(out, canary) || strings.Contains(out, th) {
				t.Fatal("error formatting leaked input")
			}
		}
	}
	if out := fmt.Sprintf("%v %+v %#v", c, c.rt.runSet, c.rt); strings.Contains(out, th) {
		t.Fatal("runtime formatting leaked a handle")
	}
}

// panickingLedger panics on the reservation write, after the run handle is
// published.
type panickingLedger struct{ *ledger.Ledger }

func (l panickingLedger) Put(ctx context.Context, k persistence.ReplayKey, rec persistence.RunRecord) error {
	if rec.Runs > 0 {
		panic("synthetic ledger panic")
	}
	return l.Ledger.Put(ctx, k, rec)
}

// A panic after the run handle is published withdraws the handle: no record
// stays active without a run.
func TestStartRunPanicWithdrawsTheHandle(t *testing.T) {
	h := runningHarness(t)
	h.adapters.run.ledger = panickingLedger{h.run.ledger}
	c := h.initialized(t)
	if _, err := c.StartRun(h.admit(t, c, runTask("panic-1", "", "", false))); codeOf(err) != "run_not_due" {
		t.Fatal(codeOf(err))
	}
	if len(c.rt.runs.entries) != 0 || h.run.runner.Calls() != 0 || c.active != 0 {
		t.Fatal("a panicking reservation left a run handle, a run or a reservation")
	}
}
