package execution_test

import (
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/execution"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/persistence"
)

var anchor = time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)

func repeating(maxRuns int) task.Schedule {
	interval, jitter := 21600, 300
	return task.Schedule{Repeat: task.Repeat{Mode: "interval", IntervalSeconds: &interval, JitterSeconds: &jitter}, MaxRuns: &maxRuns, Retry: task.Retry{MaxAttempts: 1}}
}

func oneTime() task.Schedule {
	one := 1
	return task.Schedule{Repeat: task.Repeat{Mode: "none"}, MaxRuns: &one, Retry: task.Retry{MaxAttempts: 1}}
}

// The anchor is starts_at, rounded up to a whole second, or acceptance.
func TestAccountingAnchor(t *testing.T) {
	s := oneTime()
	if rec := execution.NewAccounting(s, anchor); !rec.Anchor.Equal(anchor) || !rec.NextDue.Equal(anchor) || rec.Runs != 0 || rec.Next != 0 || !rec.Valid() {
		t.Fatal("acceptance anchor")
	}
	at := task.Timestamp{Whole: anchor.Add(time.Hour), Fraction: "5"}
	s.StartsAt = &at
	if rec := execution.NewAccounting(s, anchor); !rec.Anchor.Equal(anchor.Add(time.Hour + time.Second)) {
		t.Fatal("fractional starts_at must round up")
	}
}

// Occurrence k is due at anchor + k*interval + jitter(k), drawn once when the
// occurrence becomes current; occurrence 0 has no jitter.
func TestIntervalDueTimes(t *testing.T) {
	s := repeating(3)
	j := fixedJitter(2 * time.Minute)
	rec := execution.NewAccounting(s, anchor)
	b := []struct {
		at  time.Duration
		due bool
	}{{-time.Second, false}, {0, true}}
	for _, c := range b {
		if execution.Due(s, rec, anchor.Add(c.at)) != c.due {
			t.Fatalf("occurrence 0 at %v", c.at)
		}
	}
	rec = execution.Reserve(s, rec, j)
	if rec.Runs != 1 || rec.Next != 1 || !rec.NextDue.Equal(anchor.Add(6*time.Hour+2*time.Minute)) {
		t.Fatal("occurrence 1 due time")
	}
	for _, c := range []struct {
		at  time.Duration
		due bool
	}{{6*time.Hour + 2*time.Minute - time.Second, false}, {6*time.Hour + 2*time.Minute, true}, {6*time.Hour + 3*time.Minute, true}} {
		now := anchor.Add(c.at)
		if got := execution.Due(s, execution.Advance(s, rec, now, j), now); got != c.due {
			t.Fatalf("occurrence 1 at %v: due %t", c.at, got)
		}
	}
	// Jitter is capped at the task's jitter.
	big := execution.Reserve(s, execution.NewAccounting(s, anchor), fixedJitter(time.Hour))
	if !big.NextDue.Equal(anchor.Add(6*time.Hour + 5*time.Minute)) {
		t.Fatal("jitter above the task maximum")
	}
}

// Missed occurrences are skipped, never run late, and consume no run; past
// the last occurrence the schedule has ended.
func TestMissedOccurrencesAreSkipped(t *testing.T) {
	s := repeating(4)
	j := fixedJitter(0)
	rec := execution.Reserve(s, execution.NewAccounting(s, anchor), j)
	now := anchor.Add(12*time.Hour + time.Minute) // occurrence 2 is current; 1 was missed
	rec = execution.Advance(s, rec, now, j)
	if rec.Next != 2 || rec.Runs != 1 || !execution.Due(s, rec, now) || execution.Occurrence(s, rec, now) != 2 {
		t.Fatalf("skip to occurrence 2: next %d runs %d", rec.Next, rec.Runs)
	}
	// Advancing again within the same occurrence keeps its due time.
	if again := execution.Advance(s, rec, now.Add(time.Hour), fixedJitter(time.Minute)); again != rec {
		t.Fatal("jitter redrawn for the current occurrence")
	}
	rec = execution.Advance(s, rec, anchor.Add(24*time.Hour), j)
	if rec.Next != 4 || execution.Due(s, rec, anchor.Add(24*time.Hour)) || !rec.Valid() {
		t.Fatal("schedule did not end after the last occurrence")
	}
}

// Repeat mode none permits one run whatever max_runs says, and the run
// budget holds at limit -1, limit and limit +1.
func TestRunBudget(t *testing.T) {
	s := oneTime()
	rec := execution.Reserve(s, execution.NewAccounting(s, anchor), fixedJitter(0))
	if execution.Due(s, execution.Advance(s, rec, anchor.Add(time.Hour), fixedJitter(0)), anchor.Add(time.Hour)) || execution.Occurrence(s, rec, anchor.Add(48*time.Hour)) != 0 {
		t.Fatal("second run of a one-time task")
	}
	for _, limit := range []int{1, 2, 28} {
		s := repeating(limit)
		rec := execution.NewAccounting(s, anchor)
		runs := 0
		for k := range limit + 1 {
			now := anchor.Add(time.Duration(k) * 6 * time.Hour)
			rec = execution.Advance(s, rec, now, fixedJitter(0))
			if execution.Due(s, rec, now) {
				rec = execution.Reserve(s, rec, fixedJitter(0))
				runs++
			}
		}
		if runs != limit || rec.Runs != limit || !rec.Valid() {
			t.Fatalf("limit %d: %d runs", limit, runs)
		}
	}
	missing := repeating(1)
	missing.MaxRuns = nil
	if execution.Due(missing, execution.NewAccounting(missing, anchor), anchor) {
		t.Fatal("missing max_runs permitted a run")
	}
	if execution.Due(oneTime(), persistence.RunRecord{}, anchor) {
		t.Fatal("invalid record permitted a run")
	}
}
