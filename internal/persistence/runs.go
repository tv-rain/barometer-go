package persistence

import (
	"context"
	"errors"
	"time"
)

// ErrRunsUnavailable is the single run-ledger failure. It carries no detail.
var ErrRunsUnavailable = errors.New("persistence: run accounting unavailable")

// MaxRunsPerTask is the registered maximum of runs per task (REQ-LIMIT-0011).
const MaxRunsPerTask = 28

// RunRecord is the schedule accounting of one admitted task, keyed by its
// replay key. It holds no task input, targets, labels, results or participant
// value. Anchor is the time of occurrence 0. Runs counts reserved runs; Next is
// the first occurrence that has neither been reserved nor skipped, and NextDue
// its due time with locally drawn jitter included (GO-SPEC-ISSUE-011).
type RunRecord struct {
	Anchor     time.Time
	Runs, Next int
	NextDue    time.Time
}

// Valid reports whether r is structurally a run record.
func (r RunRecord) Valid() bool {
	return !r.Anchor.IsZero() && !r.NextDue.IsZero() && !r.NextDue.Before(r.Anchor) &&
		r.Runs >= 0 && r.Runs <= r.Next && r.Next <= MaxRunsPerTask
}

func (RunRecord) String() string   { return "persistence.RunRecord" }
func (RunRecord) GoString() string { return "persistence.RunRecord" }

// RunLedger is the run-accounting capability consumed by run reservation. A
// run is recorded before it starts, so an interrupted start can only lose a
// run, never add one. Reinitialization does not reset it; clearing deletes it.
// Implementations are bounded, application-scoped, non-synchronizing and safe
// under interruption, and honor cancellation and finite deadlines. The
// production store, its retention and its clearing belong to the State,
// retention and clearing sub-epic (DEV-577, DEV-563).
type RunLedger interface {
	// Lookup returns the record for key; ok is false when none exists. Any
	// error, including an unreadable or invalid record, means accounting is
	// unavailable; it is never treated as an absent record.
	Lookup(ctx context.Context, key ReplayKey) (rec RunRecord, ok bool, err error)
	// Put durably and atomically replaces the record for key. An error means
	// the record may or may not have become durable, but never partially.
	Put(ctx context.Context, key ReplayKey, rec RunRecord) error
}
