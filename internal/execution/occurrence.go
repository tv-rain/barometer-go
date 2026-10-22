package execution

import (
	"time"

	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/persistence"
)

// Run accounting under the interim due-time rule of GO-SPEC-ISSUE-011. The
// host directs wakeups; these functions only decide which occurrence a start
// request may reserve, and never start a run themselves.
//
//   - The anchor is starts_at when present, otherwise the acceptance time.
//   - Occurrence 0 is due at the anchor without jitter. For interval
//     repetition, occurrence k >= 1 is due at anchor + k*interval plus jitter
//     drawn locally once, when the occurrence first becomes current.
//   - The current occurrence is the latest whose unjittered time has passed;
//     once that is max_runs or later, none is. Earlier unreserved occurrences
//     are skipped and never start late; a skipped occurrence consumes no run.
//
// requirement-ref: REQ-SCHEDULE-0007
// requirement-ref: REQ-SCHEDULE-0009
// requirement-ref: REQ-SCHEDULE-0020
// requirement-ref: REQ-SCHEDULE-0021
// requirement-ref: REQ-SCHEDULE-0030

// NewAccounting returns the accounting of a task with no reserved run. A
// fractional starts_at is rounded up, so the anchor is never before it.
func NewAccounting(s task.Schedule, accepted time.Time) persistence.RunRecord {
	anchor := accepted
	if s.StartsAt != nil {
		anchor = s.StartsAt.Whole
		if s.StartsAt.Fraction != "" {
			anchor = anchor.Add(time.Second)
		}
	}
	return persistence.RunRecord{Anchor: anchor, NextDue: anchor}
}

// Occurrence returns the occurrence current at now. At or beyond max_runs no
// occurrence of the schedule is current any more.
func Occurrence(s task.Schedule, rec persistence.RunRecord, now time.Time) int {
	interval, ok := repeatInterval(s)
	if !ok || now.Before(rec.Anchor) {
		return 0
	}
	return int(min(now.Sub(rec.Anchor)/interval, persistence.MaxRunsPerTask))
}

// Advance skips to the occurrence current at now and draws its jitter. It
// returns rec unchanged when that occurrence is already the next one. Past
// the last occurrence, every remaining occurrence is skipped.
func Advance(s task.Schedule, rec persistence.RunRecord, now time.Time, j Jitter) persistence.RunRecord {
	switch k := Occurrence(s, rec, now); {
	case k <= rec.Next:
	case k >= maxRuns(s):
		rec.Next = maxRuns(s)
	default:
		rec.Next, rec.NextDue = k, dueAt(s, rec.Anchor, k, j)
	}
	return rec
}

// Due reports whether the next occurrence exists and its due time has passed.
// The other eligibility conditions are validation.RunEligible's.
func Due(s task.Schedule, rec persistence.RunRecord, now time.Time) bool {
	return rec.Valid() && rec.Next < maxRuns(s) && rec.Runs < maxRuns(s) && !now.Before(rec.NextDue)
}

// Reserve returns the accounting after reserving the next occurrence.
func Reserve(s task.Schedule, rec persistence.RunRecord, j Jitter) persistence.RunRecord {
	rec.Runs++
	rec.Next++
	if _, ok := repeatInterval(s); ok && rec.Next < maxRuns(s) {
		rec.NextDue = dueAt(s, rec.Anchor, rec.Next, j)
	}
	return rec
}

func dueAt(s task.Schedule, anchor time.Time, k int, j Jitter) time.Time {
	interval, ok := repeatInterval(s)
	if !ok || k == 0 {
		return anchor
	}
	jitter := time.Duration(*s.Repeat.JitterSeconds) * time.Second
	return anchor.Add(time.Duration(k)*interval + j.Jitter(jitter))
}

func repeatInterval(s task.Schedule) (time.Duration, bool) {
	if s.Repeat.Mode != "interval" || s.Repeat.IntervalSeconds == nil || s.Repeat.JitterSeconds == nil || *s.Repeat.IntervalSeconds <= 0 {
		return 0, false
	}
	return time.Duration(*s.Repeat.IntervalSeconds) * time.Second, true
}

// maxRuns is the validated max_runs, or zero when it is missing; repeat mode
// none permits one run whatever max_runs says.
func maxRuns(s task.Schedule) int {
	if s.MaxRuns == nil {
		return 0
	}
	if s.Repeat.Mode != "interval" {
		return min(*s.MaxRuns, 1)
	}
	return min(*s.MaxRuns, persistence.MaxRunsPerTask)
}
