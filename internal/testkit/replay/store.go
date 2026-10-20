// Package replay owns a synthetic, fault-injectable replay store for tests
// only. Production code must not import it. It models durable state in memory
// and is not a persistence implementation or a mobile fallback.
package replay

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/persistence"
)

// Fault selects one injected failure point.
type Fault uint8

const (
	// FailLookup makes Lookup unavailable.
	FailLookup Fault = iota + 1
	// FailHighest makes HighestGeneration unavailable.
	FailHighest
	// FailBeforeWrite fails Commit before anything is written.
	FailBeforeWrite
	// CrashDuringWrite stages the record but fails before it becomes durable.
	// The staged bytes are never visible and Restart discards them.
	CrashDuringWrite
	// FailAfterWrite makes the record durable but reports failure, modeling an
	// uncertain commit.
	FailAfterWrite
)

// Store keeps encoded records, keyed by replay key, as its durable state. It
// is bounded by capacity and safe for concurrent use.
type Store struct {
	mu       sync.Mutex
	capacity int
	durable  map[persistence.ReplayKey][]byte
	staged   map[persistence.ReplayKey][]byte
	faults   map[Fault]bool
	commits  int
}

// ErrCapacity is returned by New for a capacity below one.
var ErrCapacity = errors.New("testkit/replay: capacity must be positive")

func New(capacity int) (*Store, error) {
	if capacity < 1 {
		return nil, ErrCapacity
	}
	return &Store{capacity: capacity, durable: map[persistence.ReplayKey][]byte{}, staged: map[persistence.ReplayKey][]byte{}, faults: map[Fault]bool{}}, nil
}

// Inject enables or disables one fault until changed or until Restart.
func (s *Store) Inject(f Fault, on bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.faults[f] = on
}

// Restart models a process crash and restart: durable records survive, staged
// writes and injected faults do not.
func (s *Store) Restart() {
	s.mu.Lock()
	defer s.mu.Unlock()
	clear(s.staged)
	clear(s.faults)
}

// Corrupt replaces the stored bytes for key, modeling invalid durable state.
func (s *Store) Corrupt(key persistence.ReplayKey, raw []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.durable[key] = bytes.Clone(raw)
}

// Commits counts records that became durable.
func (s *Store) Commits() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.commits
}

// Len counts durable records, including overdue ones not yet purged.
func (s *Store) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.durable)
}

// Staged counts writes that never became durable.
func (s *Store) Staged() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.staged)
}

// Dump returns every durable and staged byte in key order, for leakage checks.
func (s *Store) Dump() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []byte
	for _, m := range []map[persistence.ReplayKey][]byte{s.durable, s.staged} {
		keys := make([]persistence.ReplayKey, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].IssuerID+"\x00"+keys[i].TaskID < keys[j].IssuerID+"\x00"+keys[j].TaskID
		})
		for _, k := range keys {
			out = append(append(out, m[k]...), '\n')
		}
	}
	return out
}

func (s *Store) Lookup(ctx context.Context, key persistence.ReplayKey, now time.Time) (persistence.ReplayRecord, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ready(ctx, now, FailLookup); err != nil {
		return persistence.ReplayRecord{}, false, err
	}
	raw, ok := s.durable[key]
	if !ok {
		return persistence.ReplayRecord{}, false, nil
	}
	r, err := decode(raw)
	if err != nil || r.Key != key {
		return persistence.ReplayRecord{}, false, persistence.ErrReplayCorrupt
	}
	return r, true, nil
}

func (s *Store) HighestGeneration(ctx context.Context, now time.Time) (uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ready(ctx, now, FailHighest); err != nil {
		return 0, err
	}
	var highest uint64
	for _, raw := range s.durable {
		r, err := decode(raw)
		if err != nil {
			return 0, persistence.ErrReplayCorrupt
		}
		highest = max(highest, r.PolicyGeneration)
	}
	return highest, nil
}

func (s *Store) Commit(ctx context.Context, rec persistence.ReplayRecord, now time.Time) (persistence.ReplayRecord, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !rec.Valid() {
		return persistence.ReplayRecord{}, false, persistence.ErrReplayRecord
	}
	if err := s.ready(ctx, now, FailBeforeWrite); err != nil {
		return persistence.ReplayRecord{}, false, err
	}
	if raw, ok := s.durable[rec.Key]; ok {
		existing, err := decode(raw)
		if err != nil || existing.Key != rec.Key {
			return persistence.ReplayRecord{}, false, persistence.ErrReplayCorrupt
		}
		return existing, false, nil
	}
	if len(s.durable) >= s.capacity {
		return persistence.ReplayRecord{}, false, persistence.ErrReplayUnavailable
	}
	raw := encode(rec)
	if s.faults[CrashDuringWrite] {
		s.staged[rec.Key] = raw
		return persistence.ReplayRecord{}, false, persistence.ErrReplayUnavailable
	}
	s.durable[rec.Key] = raw
	s.commits++
	if s.faults[FailAfterWrite] {
		return persistence.ReplayRecord{}, false, persistence.ErrReplayUnavailable
	}
	return rec, true, nil
}

// ready checks cancellation, trusted time and the injected fault, then purges
// overdue records before use. An overdue record that cannot be decoded is
// invalid state and is reported rather than silently dropped.
func (s *Store) ready(ctx context.Context, now time.Time, f Fault) error {
	if ctx.Err() != nil || now.IsZero() || s.faults[f] {
		return persistence.ErrReplayUnavailable
	}
	for k, raw := range s.durable {
		r, err := decode(raw)
		if err != nil {
			return persistence.ErrReplayCorrupt
		}
		if r.Overdue(now) {
			delete(s.durable, k)
		}
	}
	return nil
}

// stored is the synthetic encoding: exactly the record fields.
type stored struct {
	IssuerID, TaskID       string
	Digest                 []byte
	PolicyGeneration       uint64
	Origin                 task.Origin
	ExpiresAt, RetainUntil time.Time
}

func encode(r persistence.ReplayRecord) []byte {
	b, err := json.Marshal(stored{r.Key.IssuerID, r.Key.TaskID, r.Digest[:], r.PolicyGeneration, r.Origin, r.ExpiresAt, r.RetainUntil})
	if err != nil {
		panic("testkit/replay: encode")
	}
	return b
}

func decode(raw []byte) (persistence.ReplayRecord, error) {
	var v stored
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&v); err != nil || len(v.Digest) != len(persistence.ReplayRecord{}.Digest) {
		return persistence.ReplayRecord{}, persistence.ErrReplayCorrupt
	}
	r := persistence.ReplayRecord{Key: persistence.ReplayKey{IssuerID: v.IssuerID, TaskID: v.TaskID}, PolicyGeneration: v.PolicyGeneration,
		Origin: v.Origin, ExpiresAt: v.ExpiresAt.UTC(), RetainUntil: v.RetainUntil.UTC()}
	copy(r.Digest[:], v.Digest)
	if !r.Valid() {
		return persistence.ReplayRecord{}, persistence.ErrReplayCorrupt
	}
	return r, nil
}
