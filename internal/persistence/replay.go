// Package persistence owns narrow capability contracts for Barometer-owned
// state, the state record port and the Go stores built on it. It defines what
// core consumers need, not a storage technology: no files, SQL or platform
// storage lives here; a platform backend implements Records and stores bytes.
package persistence

import (
	"context"
	"crypto/sha256"
	"errors"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/task"
)

// Replay-state retention (REQ-LIMIT-0023): the validity of the accepted task
// plus ReplayGrace, capped at MaxReplayRetention after acceptance.
const (
	ReplayGrace        = 86400 * time.Second
	MaxReplayRetention = 691200 * time.Second
)

// Replay-store failures. They carry no detail; admission maps every store
// failure to task_trust_unavailable at the step boundary.
var (
	ErrReplayUnavailable = errors.New("persistence: replay state unavailable")
	ErrReplayCorrupt     = errors.New("persistence: replay state invalid")
	ErrReplayRecord      = errors.New("persistence: invalid replay record")
)

// ReplayKey is the effective replay key (issuer_id, task_id).
type ReplayKey struct{ IssuerID, TaskID string }

// ReplayRecord is the bounded local record of REQ-TRUST-0021 and
// REQ-TRUST-0022. It holds the replay key, the SHA-256 digest of the exact
// accepted bytes, the accepted policy generation, the acceptance class (the
// origin) and the expiry classes needed for cleanup: the task's expiry and the
// end of its retention. It has no field for task input, targets, labels,
// participant identifiers or results (REQ-TRUST-0023), and no task bytes.
// requirement-ref: REQ-TRUST-0021
// requirement-ref: REQ-TRUST-0022
// requirement-ref: REQ-TRUST-0023
type ReplayRecord struct {
	Key              ReplayKey
	Digest           [sha256.Size]byte
	PolicyGeneration uint64
	Origin           task.Origin
	// ExpiresAt is the task's not_after rounded up to a whole UTC second, so
	// the record never ends before the task does.
	ExpiresAt time.Time
	// RetainUntil is the instant from which the record is overdue and purged.
	RetainUntil time.Time
}

// NewReplayRecord builds the record for a task accepted at acceptedAt with the
// given not_after. Retention ends ReplayGrace after the task's expiry, but no
// later than MaxReplayRetention after acceptance.
// requirement-ref: REQ-LIMIT-0023
func NewReplayRecord(key ReplayKey, digest [sha256.Size]byte, generation uint64, origin task.Origin, notAfter task.Timestamp, acceptedAt time.Time) (ReplayRecord, error) {
	if acceptedAt.IsZero() || notAfter.Whole.IsZero() {
		return ReplayRecord{}, ErrReplayRecord
	}
	expires := notAfter.Whole.UTC()
	if notAfter.Fraction != "" {
		expires = expires.Add(time.Second)
	}
	retain := expires.Add(ReplayGrace)
	if limit := acceptedAt.UTC().Truncate(time.Second).Add(MaxReplayRetention); limit.Before(retain) {
		retain = limit
	}
	r := ReplayRecord{Key: key, Digest: digest, PolicyGeneration: generation, Origin: origin, ExpiresAt: expires, RetainUntil: retain}
	if !r.ExpiresAt.After(acceptedAt) || !r.Valid() {
		return ReplayRecord{}, ErrReplayRecord
	}
	return r, nil
}

// Valid reports whether r is structurally a replay record. A store that returns
// an invalid record has invalid replay state.
func (r ReplayRecord) Valid() bool {
	return boundedText(r.Key.IssuerID, 64) && boundedText(r.Key.TaskID, 128) && r.Digest != [sha256.Size]byte{} &&
		r.PolicyGeneration > 0 && (r.Origin == task.Embedded || r.Origin == task.Remote) &&
		wholeUTC(r.ExpiresAt) && wholeUTC(r.RetainUntil) && r.RetainUntil.After(r.ExpiresAt) && !r.RetainUntil.After(r.ExpiresAt.Add(ReplayGrace))
}

// Overdue reports whether retention has ended at now.
func (r ReplayRecord) Overdue(now time.Time) bool { return !now.Before(r.RetainUntil) }

// String and GoString keep the replay key out of formatted output.
func (ReplayRecord) String() string   { return "persistence.ReplayRecord" }
func (ReplayRecord) GoString() string { return "persistence.ReplayRecord" }

// ReplayStore is the replay-state capability consumed by admission steps 10 and
// 11 (REQ-TRUST-0046). Every method receives trusted time and treats records
// overdue at that time as absent, purging them before use; implementations
// never read the system clock. Implementations are bounded, application-scoped,
// non-synchronizing and safe under interruption, and honor cancellation and
// finite deadlines without background work. Any error, including an unreadable
// or invalid record, means replay state is unavailable or invalid; it is never
// treated as an absent record. Clearing is a separate capability owned by the
// State, retention and clearing sub-epic.
type ReplayStore interface {
	// Lookup returns the retained record for key; ok is false when none exists.
	Lookup(ctx context.Context, key ReplayKey, now time.Time) (rec ReplayRecord, ok bool, err error)
	// HighestGeneration returns the highest accepted policy generation among
	// retained records, or zero when none is retained.
	HighestGeneration(ctx context.Context, now time.Time) (uint64, error)
	// Commit durably and atomically writes rec only if no record for rec.Key is
	// retained. On success it returns rec and created true. If a record already
	// exists it writes nothing and returns that record with created false. An
	// error means nothing is known to be committed; the record may or may not
	// have become durable, but never partially.
	Commit(ctx context.Context, rec ReplayRecord, now time.Time) (stored ReplayRecord, created bool, err error)
}

func boundedText(s string, max int) bool {
	if len(s) == 0 || len(s) > max {
		return false
	}
	for i := range len(s) {
		if s[i] <= ' ' || s[i] > '~' {
			return false
		}
	}
	return true
}

func wholeUTC(t time.Time) bool {
	return !t.IsZero() && t.Location() == time.UTC && t.Nanosecond() == 0
}
