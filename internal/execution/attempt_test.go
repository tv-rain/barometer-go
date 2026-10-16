package execution_test

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/execution"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/validation"
)

// REQ-RESULT-0019..0022: valid transitions, reason presence and immutable
// terminal states. Pending/Running are internal, never serialized result states.
func TestTransitionMatrix(t *testing.T) {
	pending := execution.NewAttempt()
	running, err := pending.Start()
	if err != nil {
		t.Fatal(err)
	}
	completed, err := running.Complete()
	if err != nil {
		t.Fatal(err)
	}
	incomplete, err := running.Stop(task.CancelledByHost)
	if err != nil {
		t.Fatal(err)
	}
	notRun, err := pending.Stop(task.LocalPolicyRejected)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name                  string
		in                    execution.Attempt
		start, complete, stop execution.Phase
	}{
		{"zero", execution.Attempt{}, 0, 0, 0},
		{"pending", pending, execution.Running, 0, execution.NotRun},
		{"running", running, 0, execution.Completed, execution.Incomplete},
		{"completed", completed, 0, 0, 0},
		{"incomplete", incomplete, 0, 0, 0},
		{"not_run", notRun, 0, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, op := range []struct {
				name string
				want execution.Phase
				call func() (execution.Attempt, error)
			}{
				{"start", tc.start, tc.in.Start},
				{"complete", tc.complete, tc.in.Complete},
				{"stop", tc.stop, func() (execution.Attempt, error) { return tc.in.Stop(task.InternalError) }},
			} {
				t.Run(op.name, func(t *testing.T) {
					got, err := op.call()
					if op.want == 0 {
						if !errors.Is(err, execution.ErrTransition) || got != tc.in {
							t.Fatalf("invalid transition changed state: %v %v", got, err)
						}
						return
					}
					if err != nil || got.Phase() != op.want {
						t.Fatalf("transition: %v %v", got, err)
					}
					wantReason := task.ExecutionReason("")
					if op.name == "stop" {
						wantReason = task.InternalError
					}
					if got.Reason() != wantReason {
						t.Fatal("wrong reason")
					}
					if got.Terminal() != (op.name != "start") {
						t.Fatal("wrong terminal flag")
					}
				})
			}
		})
	}
}

// REQ-RESULT-0020 and REQ-SCHEDULE-0027: every registered reason, with only
// the four allowed retry classes. No string-prefix or truthy matching.
func TestAllTerminationReasons(t *testing.T) {
	for _, tc := range []struct {
		reason task.ExecutionReason
		retry  bool
	}{
		{task.CancelledByHost, false}, {task.AttemptDeadlineExceeded, false},
		{task.ResourceLimit, false}, {task.NetworkChanged, false},
		{task.PlatformSuspended, true}, {task.NetworkContextUnavailable, true},
		{task.ConstraintsUnsatisfied, false}, {task.LocalPolicyRejected, false},
		{task.ProxyNotConfigured, false}, {task.ProxyNotPermitted, false},
		{task.ProxyPathUnsupported, false}, {task.ProxyUnavailable, true},
		{task.ProxyConnectFailed, false}, {task.ProxyReconfigured, false},
		{task.InternalError, true},
	} {
		t.Run(string(tc.reason), func(t *testing.T) {
			for _, started := range []bool{false, true} {
				a := execution.NewAttempt()
				want := execution.NotRun
				if started {
					a, _ = a.Start()
					want = execution.Incomplete
				}
				stopped, err := a.Stop(tc.reason)
				if err != nil || stopped.Phase() != want || stopped.Reason() != tc.reason {
					t.Fatal("reason rejected or changed")
				}
				r, ok := stopped.RetryReason()
				if ok != tc.retry || (ok && r != tc.reason) || (!ok && r != "") {
					t.Fatal("wrong retry classification")
				}
				// First condition wins, even if a later condition is retryable.
				again, err := stopped.Stop(task.InternalError)
				if !errors.Is(err, execution.ErrTransition) || again != stopped {
					t.Fatal("first reason overwritten")
				}
			}
		})
	}
}

// REQ-RESULT-0021, REQ-SCHEDULE-0025: protocol observations do not terminate
// an attempt. Explicit casts from untrusted strings still fail closed.
func TestProtocolOutcomesAndInvalidReasons(t *testing.T) {
	for _, reason := range []string{"", "dns_error", "nxdomain", "tcp_timeout", "connection_refused", "reset", "tls_validation_failed", "redirect", "unexpected_status", "response_mismatch", "task_schema_invalid", "internal_error ", "secret-token participant target.example"} {
		for _, started := range []bool{false, true} {
			a := execution.NewAttempt()
			if started {
				a, _ = a.Start()
			}
			got, err := a.Stop(task.ExecutionReason(reason))
			if !errors.Is(err, execution.ErrReason) || got != a {
				t.Fatal("invalid reason accepted or changed state")
			}
			if _, ok := got.RetryReason(); ok {
				t.Fatal("nonterminal retry")
			}
			if started {
				got, err = got.Complete()
				if err != nil || got.Phase() != execution.Completed || got.Reason() != "" {
					t.Fatal("negative observation prevented completion")
				}
				if _, ok := got.RetryReason(); ok {
					t.Fatal("completed outcome became retryable")
				}
			}
		}
	}
}

// REQ-SCHEDULE-0023..0029: a retry candidate is not permission to retry.
func TestRetryRequiresScheduleEligibility(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	maxRuns := 1
	expires := task.TimestampFromTime(now.Add(time.Hour))
	s := task.Schedule{NotAfter: &expires, Repeat: task.Repeat{Mode: "none"}, MaxRuns: &maxRuns,
		Retry: task.Retry{MaxAttempts: 2, On: []task.ExecutionReason{task.InternalError}, Backoff: &task.Backoff{Mode: "fixed", InitialDelayMS: 30000, MaxDelayMS: 30000}}}
	ctx := validation.StartContext{Now: now, DueAt: now, Supported: true, ConstraintsSatisfied: true, SafetySatisfied: true, CompletedRuns: 1}
	a, _ := execution.NewAttempt().Start()
	a, _ = a.Stop(task.InternalError)
	r, ok := a.RetryReason()
	if !ok || !validation.RetryEligible(s, ctx, 1, r) {
		t.Fatal("allowed retry rejected")
	}
	for _, mutate := range []func(*task.Schedule, *validation.StartContext, *int){
		func(_ *task.Schedule, _ *validation.StartContext, count *int) { *count = 2 },
		func(s *task.Schedule, _ *validation.StartContext, _ *int) {
			s.Retry.On = []task.ExecutionReason{task.ProxyUnavailable}
		},
		func(_ *task.Schedule, ctx *validation.StartContext, _ *int) { ctx.Now = now.Add(time.Hour) },
		func(_ *task.Schedule, ctx *validation.StartContext, _ *int) { ctx.ConstraintsSatisfied = false },
		func(_ *task.Schedule, ctx *validation.StartContext, _ *int) { ctx.SafetySatisfied = false },
		func(_ *task.Schedule, ctx *validation.StartContext, _ *int) { ctx.DueAt = now.Add(time.Second) },
	} {
		ss, cc, count := s, ctx, 1
		mutate(&ss, &cc, &count)
		if validation.RetryEligible(ss, cc, count, r) {
			t.Fatal("retry escaped schedule bounds")
		}
	}
}

func TestAttemptRejectsJSON(t *testing.T) {
	a, _ := execution.NewAttempt().Stop(task.InternalError)
	if b, err := json.Marshal(a); !errors.Is(err, execution.ErrEncoding) || len(b) != 0 {
		t.Fatal("attempt serialized")
	}
	before := a
	if err := json.Unmarshal([]byte(`{"phase":3,"reason":"participant-secret"}`), &a); !errors.Is(err, execution.ErrEncoding) || a != before {
		t.Fatal("JSON changed attempt")
	}
}

// requirement-ref: REQ-JSON-0011
func TestStopValidationPrecedence(t *testing.T) {
	terminal, err := execution.NewAttempt().Stop(task.CancelledByHost)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		in   execution.Attempt
		want error
	}{
		{"invalid phase and reason", execution.Attempt{}, execution.ErrTransition},
		{"terminal phase and reason", terminal, execution.ErrTransition},
		{"pending invalid reason", execution.NewAttempt(), execution.ErrReason},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for range 20 {
				got, err := tc.in.Stop(task.ExecutionReason("DO_NOT_RETAIN"))
				if got != tc.in || !errors.Is(err, tc.want) || len(err.Error()) > 128 {
					t.Fatalf("precedence or atomicity: %v", err)
				}
			}
		})
	}
}
