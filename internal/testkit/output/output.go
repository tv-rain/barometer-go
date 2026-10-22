// Package output is a test-only, fault-injectable pending-output store for
// execution tests. It keeps serialized results in memory as its model of
// durable state. It is not a persistence implementation or a mobile fallback,
// and production code must not import it.
package output

import (
	"context"
	"errors"
	"sync"

	"github.com/tv-rain/barometer-go/internal/result"
)

// MaxRecords bounds the store; a test needing more is mis-specified.
const MaxRecords = 16

var (
	ErrInjected = errors.New("output: injected failure")
	ErrFull     = errors.New("output: store full")
)

// Fault selects where an injected failure happens.
type Fault uint8

const (
	None Fault = iota
	// BeforeWrite fails without storing anything.
	BeforeWrite
	// AfterWrite stores the record durably and then fails: an uncertain commit.
	AfterWrite
)

// Store is safe for concurrent use. A fault applies to one Put call, selected
// by its 1-based call number.
type Store struct {
	mu      sync.Mutex
	calls   int
	faultAt int
	fault   Fault
	records [][]byte
	// limit, when positive, lowers the reported capacity below MaxRecords;
	// capacityErr fails every Capacity call.
	limit       int
	capacityErr bool
}

func New() *Store { return &Store{} }

// FailAt injects fault f on the call-th Put. A later call replaces it.
func (s *Store) FailAt(call int, f Fault) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.faultAt, s.fault = call, f
}

func (s *Store) Put(ctx context.Context, v result.Value) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	fault := None
	if s.calls == s.faultAt {
		fault = s.fault
	}
	if fault == BeforeWrite {
		return ErrInjected
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(s.records) >= MaxRecords {
		return ErrFull
	}
	b, err := v.MarshalJSON()
	if err != nil {
		return err
	}
	s.records = append(s.records, b)
	if fault == AfterWrite {
		return ErrInjected
	}
	return nil
}

// Calls reports how many Put calls were made.
func (s *Store) Calls() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

// Records returns copies of the durable serialized results, in write order.
func (s *Store) Records() [][]byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([][]byte, len(s.records))
	for i, r := range s.records {
		out[i] = append([]byte(nil), r...)
	}
	return out
}

// SetLimit lowers the capacity the store reports to limit records in total.
func (s *Store) SetLimit(limit int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.limit = limit
}

// FailCapacity makes every later Capacity call fail.
func (s *Store) FailCapacity() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.capacityErr = true
}

// Capacity reports how many more records the store accepts.
func (s *Store) Capacity(ctx context.Context) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.capacityErr {
		return 0, ErrInjected
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	limit := MaxRecords
	if s.limit > 0 {
		limit = min(s.limit, MaxRecords)
	}
	return max(limit-len(s.records), 0), nil
}
