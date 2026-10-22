package admission_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/persistence"
	"github.com/tv-rain/barometer-go/internal/testkit/records"
	"github.com/tv-rain/barometer-go/internal/testkit/replay"
)

// replayBackend is a replay store with the test controls of the synthetic
// replay store, so one suite runs against the synthetic store and against the
// production persistence.State over the fault-injectable record port.
type replayBackend interface {
	persistence.ReplayStore
	Inject(f replay.Fault, on bool)
	Restart()
	Corrupt(key persistence.ReplayKey, raw []byte)
	Commits() int
	Len() int
	Staged() int
	Dump() []byte
}

type backendKind string

const (
	syntheticReplay  backendKind = "synthetic store"
	productionReplay backendKind = "production state"
)

var replayBackends = []backendKind{syntheticReplay, productionReplay}

func newReplayBackend(t testing.TB, kind backendKind, capacity int, now func() time.Time) replayBackend {
	t.Helper()
	if kind == syntheticReplay {
		return mustReplay(t, capacity)
	}
	b := &stateBackend{records: records.New(), capacity: capacity, now: now, faults: map[replay.Fault]bool{}}
	b.reopen()
	if b.state == nil {
		t.Fatal("state rejected")
	}
	return b
}

// stateBackend maps the synthetic store's controls onto the record port:
// write faults become record-port faults, and a restart reopens the state at
// the harness clock. The record port has no separate read paths for lookup
// and generation reads, so those two faults fail the view call itself.
// Corruption is found when the state is opened; an unopenable state is
// unavailable, as a client that failed to initialize would be.
type stateBackend struct {
	mu       sync.Mutex
	records  *records.Store
	state    *persistence.State
	capacity int
	now      func() time.Time
	faults   map[replay.Fault]bool
}

func (b *stateBackend) reopen() {
	b.state, _ = persistence.OpenState(context.Background(), b.records, b.capacity, b.now())
}

func (b *stateBackend) view(f replay.Fault) persistence.ReplayStore {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.state == nil || b.faults[f] {
		return nil
	}
	return b.state.Replay()
}

func (b *stateBackend) Lookup(ctx context.Context, key persistence.ReplayKey, now time.Time) (persistence.ReplayRecord, bool, error) {
	v := b.view(replay.FailLookup)
	if v == nil {
		return persistence.ReplayRecord{}, false, persistence.ErrReplayUnavailable
	}
	return v.Lookup(ctx, key, now)
}

func (b *stateBackend) HighestGeneration(ctx context.Context, now time.Time) (uint64, error) {
	v := b.view(replay.FailHighest)
	if v == nil {
		return 0, persistence.ErrReplayUnavailable
	}
	return v.HighestGeneration(ctx, now)
}

func (b *stateBackend) Commit(ctx context.Context, rec persistence.ReplayRecord, now time.Time) (persistence.ReplayRecord, bool, error) {
	v := b.view(0)
	if v == nil {
		return persistence.ReplayRecord{}, false, persistence.ErrReplayUnavailable
	}
	return v.Commit(ctx, rec, now)
}

var writeFaults = map[replay.Fault]records.Fault{
	replay.FailBeforeWrite:  records.FailBeforePut,
	replay.CrashDuringWrite: records.CrashDuringPut,
	replay.FailAfterWrite:   records.FailAfterPut,
}

func (b *stateBackend) Inject(f replay.Fault, on bool) {
	if rf, ok := writeFaults[f]; ok {
		b.records.Inject(rf, on)
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.faults[f] = on
}

func (b *stateBackend) Restart() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.records.Restart()
	clear(b.faults)
	b.reopen()
}

func (b *stateBackend) Corrupt(key persistence.ReplayKey, raw []byte) {
	b.records.Corrupt(persistence.RecordKey{Class: persistence.ReplayClass, Name: persistence.RecordNameOf(key)}, raw)
	b.mu.Lock()
	defer b.mu.Unlock()
	b.reopen()
}

func (b *stateBackend) Commits() int { return b.records.Writes() }
func (b *stateBackend) Len() int     { return b.records.Len(persistence.ReplayClass) }
func (b *stateBackend) Staged() int  { return b.records.Staged() }
func (b *stateBackend) Dump() []byte { return b.records.Dump() }

// newReplay returns a replay store of the harness's kind that reads the
// harness clock when it is (re)opened.
func (h *harness) newReplay(t testing.TB, capacity int) replayBackend {
	return newReplayBackend(t, h.kind, capacity, func() time.Time { return h.clock.Clock.Now() })
}

// clear runs persistence.Clear over the record port and, on success, has
// the State forget its index, as the root Client's clearing capability does.
func (b *stateBackend) clear(ctx context.Context) error {
	if err := persistence.Clear(ctx, b.records); err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.state != nil {
		b.state.Forget()
	}
	return nil
}

func (b *stateBackend) marked() bool {
	m, err := persistence.Clearing(context.Background(), b.records)
	return m && err == nil
}
