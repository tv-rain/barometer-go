// Package ledger owns a synthetic, fault-injectable run ledger for tests only.
// Production code must not import it. It models durable state in memory and
// is not a persistence implementation or a mobile fallback.
package ledger

import (
	"context"
	"errors"
	"sync"

	"github.com/tv-rain/barometer-go/internal/persistence"
)

// Fault selects one injected failure point.
type Fault uint8

const (
	// FailLookup makes Lookup unavailable.
	FailLookup Fault = iota + 1
	// FailBeforeWrite fails Put before anything is written.
	FailBeforeWrite
	// FailAfterWrite makes the record durable but reports failure, modeling an
	// uncertain commit.
	FailAfterWrite
)

// MaxRecords bounds the ledger; a test needing more is mis-specified.
const MaxRecords = 256

var (
	ErrInjected = errors.New("testkit/ledger: injected failure")
	ErrFull     = errors.New("testkit/ledger: ledger full")
)

// Ledger is safe for concurrent use. Records survive client reinitialization
// because the test keeps the same ledger across clients.
type Ledger struct {
	mu      sync.Mutex
	records map[persistence.ReplayKey]persistence.RunRecord
	faults  map[Fault]bool
	puts    int
}

func New() *Ledger {
	return &Ledger{records: map[persistence.ReplayKey]persistence.RunRecord{}, faults: map[Fault]bool{}}
}

// Inject enables or disables one fault until changed.
func (l *Ledger) Inject(f Fault, on bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.faults[f] = on
}

func (l *Ledger) Lookup(ctx context.Context, key persistence.ReplayKey) (persistence.RunRecord, bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.faults[FailLookup] {
		return persistence.RunRecord{}, false, persistence.ErrRunsUnavailable
	}
	if err := ctx.Err(); err != nil {
		return persistence.RunRecord{}, false, err
	}
	r, ok := l.records[key]
	return r, ok, nil
}

func (l *Ledger) Put(ctx context.Context, key persistence.ReplayKey, rec persistence.RunRecord) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.puts++
	if l.faults[FailBeforeWrite] {
		return ErrInjected
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, ok := l.records[key]; !ok && len(l.records) >= MaxRecords {
		return ErrFull
	}
	l.records[key] = rec
	if l.faults[FailAfterWrite] {
		return ErrInjected
	}
	return nil
}

// Record returns the durable record for key.
func (l *Ledger) Record(key persistence.ReplayKey) (persistence.RunRecord, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	r, ok := l.records[key]
	return r, ok
}

// Puts reports how many Put calls were made.
func (l *Ledger) Puts() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.puts
}
