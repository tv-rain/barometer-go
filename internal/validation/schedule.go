package validation

import (
	"errors"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/task"
)

var ErrTiming = errors.New("validation: timing")

// TimingContext is supplied separately by a trusted caller. This value does not
// establish authentication, authorization, replay protection or full admission.
type TimingContext struct {
	Origin     string
	AcceptedAt time.Time
	IssuedAt   *time.Time
	// True only for release-reviewed one-time embedded tasks that the host will
	// not persist across application updates.
	EmbeddedOneTimeException bool
}

// ScheduleAt enforces expiry against explicit trusted time, never time.Now().
// Every failure matches ErrTiming; a not_after at or before acceptance also
// matches ErrExpired and one beyond the 7-day lifetime also matches ErrLimit
// (REQ-LIMIT-0009).
func ScheduleAt(s task.Schedule, ctx TimingContext) error {
	if !validSchedule(s) || ctx.AcceptedAt.IsZero() {
		return ErrTiming
	}
	base := ctx.AcceptedAt
	switch ctx.Origin {
	case "remote":
		if ctx.IssuedAt == nil || ctx.IssuedAt.IsZero() || ctx.IssuedAt.After(ctx.AcceptedAt.Add(300*time.Second)) || s.NotAfter == nil {
			return ErrTiming
		}
		if ctx.IssuedAt.Before(base) {
			base = *ctx.IssuedAt
		}
	case "embedded":
		if ctx.IssuedAt != nil {
			return ErrTiming
		}
		if s.NotAfter == nil && (!ctx.EmbeddedOneTimeException || s.Repeat.Mode != "none" || s.MaxRuns == nil || *s.MaxRuns != 1) {
			return ErrTiming
		}
	default:
		return ErrTiming
	}
	if s.NotAfter != nil && !s.NotAfter.AfterTime(ctx.AcceptedAt) {
		return errExpired
	}
	if s.NotAfter != nil && s.NotAfter.AfterTime(base.Add(604800*time.Second)) {
		return errTimingLimit
	}
	return nil
}

// StartContext describes an already due run/attempt, including locally realized
// jitter. Orchestration owns due-time generation, history, clocks and randomness.
type StartContext struct {
	Now, DueAt                                       time.Time
	LocalZone                                        *time.Location
	CompletedRuns                                    int
	Supported, ConstraintsSatisfied, SafetySatisfied bool
}

// RunEligible checks only the start envelope. A true result is not admission.
func RunEligible(s task.Schedule, ctx StartContext) bool {
	if !startEnvelope(s, ctx) || s.MaxRuns == nil || ctx.CompletedRuns < 0 || ctx.CompletedRuns >= *s.MaxRuns {
		return false
	}
	return s.Repeat.Mode != "none" || ctx.CompletedRuns == 0
}

// RetryEligible belongs to an existing run: retries do not consume max_runs.
// CompletedAttempts includes the first attempt. Measurement outcomes cannot be
// promoted into registered cross-cutting execution reasons.
func RetryEligible(s task.Schedule, ctx StartContext, completedAttempts int, reason task.ExecutionReason) bool {
	if !startEnvelope(s, ctx) || completedAttempts < 1 || completedAttempts >= s.Retry.MaxAttempts || !task.Retryable(reason) {
		return false
	}
	for _, allowed := range s.Retry.On {
		if allowed == reason {
			return true
		}
	}
	return false
}

func startEnvelope(s task.Schedule, ctx StartContext) bool {
	if !validSchedule(s) || !ctx.Supported || !ctx.ConstraintsSatisfied || !ctx.SafetySatisfied || ctx.Now.IsZero() || ctx.DueAt.IsZero() || ctx.Now.Before(ctx.DueAt) {
		return false
	}
	if s.StartsAt != nil && s.StartsAt.AfterTime(ctx.Now) {
		return false
	}
	if s.NotAfter != nil && !s.NotAfter.AfterTime(ctx.Now) {
		return false
	}
	if s.Window == nil {
		return true
	}
	if ctx.LocalZone == nil {
		return false
	}
	now := ctx.Now.In(ctx.LocalZone)
	minute := now.Hour()*60 + now.Minute()
	for _, r := range s.Window.Ranges {
		if r.Start < r.End && minute >= r.Start && minute < r.End {
			return true
		}
		if r.End < r.Start && (minute >= r.Start || minute < r.End) {
			return true
		}
	}
	return false
}
