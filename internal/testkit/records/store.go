// Package records owns a synthetic, fault-injectable implementation of the
// persistence.Records port for tests only. Production code must not import
// it. It models durable state in memory and is not a persistence
// implementation or a mobile fallback.
package records

import (
	"bytes"
	"context"
	"sort"
	"sync"

	"github.com/tv-rain/barometer-go/internal/persistence"
)

// Fault selects one injected failure point.
type Fault uint8

const (
	// FailGet makes Get unavailable.
	FailGet Fault = iota + 1
	// FailList makes List unavailable.
	FailList
	// FailBeforePut fails Put before anything is written.
	FailBeforePut
	// CrashDuringPut stages the value but fails before it becomes durable.
	// The staged bytes are never visible and Restart discards them.
	CrashDuringPut
	// FailAfterPut makes the value durable but reports failure, modeling an
	// uncertain write.
	FailAfterPut
	// FailBeforeDelete fails Delete before anything is removed.
	FailBeforeDelete
	// FailAfterDelete removes the record but reports failure.
	FailAfterDelete
	// Stall makes every operation wait for its context to end, then fail
	// with the context's error.
	Stall
)

// MaxRecords bounds the store; a test needing more is mis-specified.
const MaxRecords = 2 * persistence.MaxClassRecords

// Store keeps values keyed by record key as its durable state. It is safe for
// concurrent use.
type Store struct {
	mu      sync.Mutex
	durable map[persistence.RecordKey][]byte
	staged  map[persistence.RecordKey][]byte
	faults  map[Fault]bool
	puts    int
	writes  int
	deletes int
}

func New() *Store {
	return &Store{durable: map[persistence.RecordKey][]byte{}, staged: map[persistence.RecordKey][]byte{}, faults: map[Fault]bool{}}
}

// Inject enables or disables one fault until changed or until Restart.
func (s *Store) Inject(f Fault, on bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.faults[f] = on
}

// Restart models a process crash and restart: durable records survive, staged
// writes and injected faults do not. Stores built over s must be reopened.
func (s *Store) Restart() {
	s.mu.Lock()
	defer s.mu.Unlock()
	clear(s.staged)
	clear(s.faults)
}

// Corrupt replaces the stored bytes for key, modeling invalid durable state.
func (s *Store) Corrupt(key persistence.RecordKey, raw []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.durable[key] = bytes.Clone(raw)
}

// Value returns a copy of the durable bytes for key.
func (s *Store) Value(key persistence.RecordKey) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.durable[key]
	return bytes.Clone(v), ok
}

// Len counts durable records of class.
func (s *Store) Len(class persistence.RecordClass) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for k := range s.durable {
		if k.Class == class {
			n++
		}
	}
	return n
}

// Staged counts writes that never became durable.
func (s *Store) Staged() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.staged)
}

// Writes counts values that became durable.
func (s *Store) Writes() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.writes
}

// Puts and Deletes count calls that reached the store.
func (s *Store) Puts() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.puts
}

func (s *Store) Deletes() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.deletes
}

// Dump returns every durable and staged byte in key order, for leakage checks.
func (s *Store) Dump() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []byte
	for _, m := range []map[persistence.RecordKey][]byte{s.durable, s.staged} {
		for _, k := range sorted(m) {
			out = append(append(out, m[k]...), '\n')
		}
	}
	return out
}

func (s *Store) Get(ctx context.Context, key persistence.RecordKey) ([]byte, bool, error) {
	if err := s.enter(ctx, FailGet); err != nil {
		return nil, false, err
	}
	defer s.mu.Unlock()
	v, ok := s.durable[key]
	return bytes.Clone(v), ok, nil
}

func (s *Store) Put(ctx context.Context, key persistence.RecordKey, value []byte) error {
	if err := s.enter(ctx, FailBeforePut); err != nil {
		return err
	}
	defer s.mu.Unlock()
	s.puts++
	if !key.Class.Valid() || len(value) > persistence.MaxRecordValue {
		return persistence.ErrRecordsUnavailable
	}
	if _, ok := s.durable[key]; !ok && len(s.durable) >= MaxRecords {
		return persistence.ErrRecordsUnavailable
	}
	if s.faults[CrashDuringPut] {
		s.staged[key] = bytes.Clone(value)
		return persistence.ErrRecordsUnavailable
	}
	s.durable[key] = bytes.Clone(value)
	s.writes++
	if s.faults[FailAfterPut] {
		return persistence.ErrRecordsUnavailable
	}
	return nil
}

func (s *Store) Delete(ctx context.Context, key persistence.RecordKey) error {
	if err := s.enter(ctx, FailBeforeDelete); err != nil {
		return err
	}
	defer s.mu.Unlock()
	s.deletes++
	delete(s.durable, key)
	if s.faults[FailAfterDelete] {
		return persistence.ErrRecordsUnavailable
	}
	return nil
}

func (s *Store) List(ctx context.Context, class persistence.RecordClass) ([]persistence.RecordKey, error) {
	if err := s.enter(ctx, FailList); err != nil {
		return nil, err
	}
	defer s.mu.Unlock()
	var keys []persistence.RecordKey
	for _, k := range sorted(s.durable) {
		if k.Class == class {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

// enter checks cancellation, a stall and the operation's fault. On success it
// returns with s.mu held.
func (s *Store) enter(ctx context.Context, f Fault) error {
	s.mu.Lock()
	if s.faults[Stall] {
		s.mu.Unlock()
		<-ctx.Done()
		return ctx.Err()
	}
	if err := ctx.Err(); err != nil {
		s.mu.Unlock()
		return err
	}
	if s.faults[f] {
		s.mu.Unlock()
		return persistence.ErrRecordsUnavailable
	}
	return nil
}

func sorted(m map[persistence.RecordKey][]byte) []persistence.RecordKey {
	keys := make([]persistence.RecordKey, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Class != keys[j].Class {
			return keys[i].Class < keys[j].Class
		}
		return bytes.Compare(keys[i].Name[:], keys[j].Name[:]) < 0
	})
	return keys
}
