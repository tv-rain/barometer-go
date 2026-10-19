package validation_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/validation"
)

// requirement-ref: REQ-TRUST-0019
// requirement-ref: REQ-LIMIT-0004
func TestTaskIDRulesAndLimitClass(t *testing.T) {
	for id, want := range map[string]error{
		"a": nil, strings.Repeat("a", 128): nil, "a.b_c-9": nil,
		"": validation.ErrTask, "A": validation.ErrTask, "a b": validation.ErrTask, "a/b": validation.ErrTask, "é": validation.ErrTask, "-a": validation.ErrTask,
		strings.Repeat("a", 129): validation.ErrLimit, strings.Repeat("A", 129): validation.ErrLimit,
	} {
		err := validation.TaskID(id)
		if want == nil && err != nil || want != nil && !errors.Is(err, want) || err != nil && !errors.Is(err, validation.ErrTask) {
			t.Errorf("task_id of %d bytes: unexpected class", len(id))
		}
		if errors.Is(err, validation.ErrLimit) != (want == validation.ErrLimit) {
			t.Errorf("task_id of %d bytes: limit class mismatch", len(id))
		}
	}
}

// A value above a registered maximum is a limit failure; a value below a
// minimum, or another invalid value, is not. Each refining class still
// matches the coarse class it refines.
// requirement-ref: REQ-LIMIT-0002
// requirement-ref: REQ-SAFETY-0017
func TestRefiningClasses(t *testing.T) {
	ptr := func(n int) *int { return &n }
	cases := []struct {
		name  string
		edit  func(*task.Candidate)
		class error
	}{
		{"duration above", func(c *task.Candidate) { c.Execution.Limits.DurationMS = 60001 }, validation.ErrLimit},
		{"duration below", func(c *task.Candidate) { c.Execution.Limits.DurationMS = 0 }, nil},
		{"connections above", func(c *task.Candidate) { c.Execution.Limits.Connections = 13 }, validation.ErrLimit},
		{"runs above", func(c *task.Candidate) { c.Schedule.MaxRuns = ptr(29) }, validation.ErrLimit},
		{"runs below", func(c *task.Candidate) { c.Schedule.MaxRuns = ptr(0) }, nil},
		{"attempts above", func(c *task.Candidate) { c.Schedule.Retry.MaxAttempts = 3 }, validation.ErrLimit},
		{"labels above", func(c *task.Candidate) {
			for i := range 17 {
				c.Labels = append(c.Labels, task.Label{Key: "k" + strings.Repeat("a", i)})
			}
		}, validation.ErrLimit},
		{"unknown risk factor type", func(c *task.Candidate) {
			c.Safety.RiskLevel, c.Safety.Factors = "medium", []task.RiskFactor{{Type: "unregistered_factor", Level: "medium"}}
		}, validation.ErrUnsupported},
		{"unknown type with invalid level", func(c *task.Candidate) {
			c.Safety.RiskLevel, c.Safety.Factors = "low", []task.RiskFactor{{Type: "unregistered_factor", Level: "high"}}
		}, nil},
		{"jitter above quarter interval", func(c *task.Candidate) {
			na := task.TimestampFromTime(time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC))
			c.Schedule.NotAfter, c.Schedule.MaxRuns = &na, ptr(2)
			c.Schedule.Repeat = task.Repeat{Mode: "interval", IntervalSeconds: ptr(21600), JitterSeconds: ptr(1801)}
		}, validation.ErrLimit},
	}
	for _, tc := range cases {
		c := candidate()
		tc.edit(&c)
		_, err := validation.Common(c)
		if !errors.Is(err, validation.ErrTask) {
			t.Errorf("%s: not rejected as task", tc.name)
			continue
		}
		for _, class := range []error{validation.ErrLimit, validation.ErrUnsupported, validation.ErrExpired} {
			if errors.Is(err, class) != (class == tc.class) {
				t.Errorf("%s: refining class mismatch", tc.name)
			}
		}
	}

	accepted := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	timing := validation.TimingContext{Origin: "embedded", AcceptedAt: accepted}
	for name, c := range map[string]struct {
		notAfter time.Time
		class    error
	}{
		"at acceptance":   {accepted, validation.ErrExpired},
		"before":          {accepted.Add(-time.Second), validation.ErrExpired},
		"seven days":      {accepted.Add(7 * 24 * time.Hour), nil},
		"over seven days": {accepted.Add(7*24*time.Hour + time.Second), validation.ErrLimit},
	} {
		na := task.TimestampFromTime(c.notAfter)
		err := validation.ScheduleAt(task.Schedule{NotAfter: &na, Repeat: task.Repeat{Mode: "none"}, Retry: task.Retry{MaxAttempts: 1}}, timing)
		if c.class == nil && err != nil || c.class != nil && (!errors.Is(err, c.class) || !errors.Is(err, validation.ErrTiming)) {
			t.Errorf("not_after %s: unexpected class", name)
		}
	}
}
