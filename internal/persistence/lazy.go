package persistence

import (
	"context"
	"time"
)

// LazyState opens a State over its Records port on first use, so the binding
// that creates it performs no I/O. Opening runs under the first caller's
// context and trusted time; a failed open is retried by the next operation.
// It is the replay store and run ledger of one storage scope, like State.
type LazyState struct {
	sem      chan struct{}
	records  Records
	capacity int
	state    *State
}

// NewLazyState returns an unopened state over records with the replay
// capacity of OpenState.
func NewLazyState(records Records, capacity int) (*LazyState, error) {
	if records == nil || capacity < 1 || capacity > MaxClassRecords {
		return nil, ErrStateUnavailable
	}
	return &LazyState{sem: make(chan struct{}, 1), records: records, capacity: capacity}, nil
}

// open returns the opened State. The one-slot semaphore is waited for only
// until ctx ends, so a stalled first open does not hold later callers past
// their own deadlines.
func (l *LazyState) open(ctx context.Context, now time.Time) (*State, error) {
	select {
	case l.sem <- struct{}{}:
	case <-ctx.Done():
		return nil, ErrStateUnavailable
	}
	defer func() { <-l.sem }()
	if l.state == nil {
		s, err := OpenState(ctx, l.records, l.capacity, now)
		if err != nil {
			return nil, ErrStateUnavailable
		}
		l.state = s
	}
	return l.state, nil
}

// Replay returns the replay-store view.
func (l *LazyState) Replay() ReplayStore { return lazyReplay{l} }

// Runs returns the run-ledger view.
func (l *LazyState) Runs() RunLedger { return lazyRuns{l} }

// Forget forgets the opened State, if any. An unopened state holds no index.
func (l *LazyState) Forget() {
	l.sem <- struct{}{}
	s := l.state
	<-l.sem
	if s != nil {
		s.Forget()
	}
}

func (*LazyState) String() string   { return "persistence.LazyState" }
func (*LazyState) GoString() string { return "persistence.LazyState" }

type lazyReplay struct{ l *LazyState }
type lazyRuns struct{ l *LazyState }

func (v lazyReplay) Forget() { v.l.Forget() }
func (v lazyRuns) Forget()   { v.l.Forget() }

func (v lazyReplay) Lookup(ctx context.Context, key ReplayKey, now time.Time) (ReplayRecord, bool, error) {
	s, err := v.l.open(ctx, now)
	if err != nil {
		return ReplayRecord{}, false, err
	}
	return s.Replay().Lookup(ctx, key, now)
}

func (v lazyReplay) HighestGeneration(ctx context.Context, now time.Time) (uint64, error) {
	s, err := v.l.open(ctx, now)
	if err != nil {
		return 0, err
	}
	return s.Replay().HighestGeneration(ctx, now)
}

func (v lazyReplay) Commit(ctx context.Context, rec ReplayRecord, now time.Time) (ReplayRecord, bool, error) {
	s, err := v.l.open(ctx, now)
	if err != nil {
		return ReplayRecord{}, false, err
	}
	return s.Replay().Commit(ctx, rec, now)
}

func (v lazyRuns) Lookup(ctx context.Context, key ReplayKey, now time.Time) (RunRecord, bool, error) {
	s, err := v.l.open(ctx, now)
	if err != nil {
		return RunRecord{}, false, err
	}
	return s.Runs().Lookup(ctx, key, now)
}

func (v lazyRuns) Put(ctx context.Context, key ReplayKey, rec RunRecord, now time.Time) error {
	s, err := v.l.open(ctx, now)
	if err != nil {
		return err
	}
	return s.Runs().Put(ctx, key, rec, now)
}
