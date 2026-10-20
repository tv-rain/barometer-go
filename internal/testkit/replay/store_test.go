package replay_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/persistence"
	"github.com/tv-rain/barometer-go/internal/testkit/replay"
)

var now = time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)

func record(t *testing.T, taskID string, generation uint64, valid time.Duration) persistence.ReplayRecord {
	t.Helper()
	r, err := persistence.NewReplayRecord(persistence.ReplayKey{IssuerID: "issuer", TaskID: taskID}, sha256.Sum256([]byte(taskID)), generation, task.Remote, task.TimestampFromTime(now.Add(valid)), now)
	if err != nil {
		t.Fatal("record rejected")
	}
	return r
}

func store(t *testing.T, capacity int) *replay.Store {
	t.Helper()
	s, err := replay.New(capacity)
	if err != nil {
		t.Fatal("store rejected")
	}
	return s
}

func TestStoreCommitsConditionallyAndPurgesOverdueRecords(t *testing.T) {
	ctx := context.Background()
	if _, err := replay.New(0); !errors.Is(err, replay.ErrCapacity) {
		t.Fatal("zero capacity accepted")
	}
	s := store(t, 2)
	a, b, c := record(t, "a", 2, time.Hour), record(t, "b", 3, 2*time.Hour), record(t, "c", 2, time.Hour)
	if got, created, err := s.Commit(ctx, a, now); err != nil || !created || got != a {
		t.Fatal("first commit failed")
	}
	other := a
	other.Digest = sha256.Sum256([]byte("other"))
	if got, created, err := s.Commit(ctx, other, now); err != nil || created || got != a {
		t.Fatal("conditional commit overwrote an existing record")
	}
	if _, _, err := s.Commit(ctx, b, now); err != nil {
		t.Fatal("second commit failed")
	}
	if _, _, err := s.Commit(ctx, c, now); !errors.Is(err, persistence.ErrReplayUnavailable) {
		t.Fatal("full store evicted or exceeded capacity")
	}
	if g, err := s.HighestGeneration(ctx, now); err != nil || g != 3 {
		t.Fatal("highest generation wrong")
	}
	// a is overdue first; at its retention end it is purged before use.
	if _, ok, err := s.Lookup(ctx, a.Key, a.RetainUntil.Add(-time.Second)); err != nil || !ok {
		t.Fatal("retained record missing")
	}
	if _, ok, err := s.Lookup(ctx, a.Key, a.RetainUntil); err != nil || ok || s.Len() != 1 {
		t.Fatal("overdue record not purged")
	}
	if g, err := s.HighestGeneration(ctx, b.RetainUntil); err != nil || g != 0 || s.Len() != 0 {
		t.Fatal("overdue generation retained")
	}
	if s.Commits() != 2 {
		t.Fatal("commit count wrong")
	}
	if _, _, err := s.Commit(ctx, persistence.ReplayRecord{}, now); !errors.Is(err, persistence.ErrReplayRecord) {
		t.Fatal("invalid record committed")
	}
}

func TestStoreFaults(t *testing.T) {
	ctx := context.Background()
	a := record(t, "a", 2, time.Hour)
	for name, tc := range map[string]struct {
		fault           replay.Fault
		durable, staged int
	}{
		"before write":    {replay.FailBeforeWrite, 0, 0},
		"during write":    {replay.CrashDuringWrite, 0, 1},
		"after write":     {replay.FailAfterWrite, 1, 0},
		"lookup only":     {replay.FailLookup, 1, 0},
		"generation only": {replay.FailHighest, 1, 0},
	} {
		s := store(t, 4)
		s.Inject(tc.fault, true)
		_, _, err := s.Commit(ctx, a, now)
		if tc.fault == replay.FailLookup || tc.fault == replay.FailHighest {
			if err != nil {
				t.Errorf("%s: unrelated fault failed the commit", name)
			}
		} else if !errors.Is(err, persistence.ErrReplayUnavailable) {
			t.Errorf("%s: commit fault not reported", name)
		}
		if s.Len() != tc.durable || s.Staged() != tc.staged {
			t.Errorf("%s: durable %d staged %d", name, s.Len(), s.Staged())
		}
		s.Restart()
		if s.Staged() != 0 || s.Len() != tc.durable {
			t.Errorf("%s: restart kept staged state or lost durable state", name)
		}
	}
	s := store(t, 4)
	s.Inject(replay.FailLookup, true)
	if _, _, err := s.Lookup(ctx, a.Key, now); !errors.Is(err, persistence.ErrReplayUnavailable) {
		t.Error("lookup fault not reported")
	}
	s.Inject(replay.FailHighest, true)
	if _, err := s.HighestGeneration(ctx, now); !errors.Is(err, persistence.ErrReplayUnavailable) {
		t.Error("generation fault not reported")
	}
	s.Restart()
	s.Corrupt(a.Key, []byte(`{"IssuerID":"issuer"`))
	if _, _, err := s.Lookup(ctx, a.Key, now); !errors.Is(err, persistence.ErrReplayCorrupt) {
		t.Error("corrupt record not reported")
	}
	if _, err := s.HighestGeneration(ctx, now); !errors.Is(err, persistence.ErrReplayCorrupt) {
		t.Error("corrupt state not reported by generation read")
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, _, err := store(t, 1).Lookup(cancelled, a.Key, now); !errors.Is(err, persistence.ErrReplayUnavailable) {
		t.Error("cancelled lookup succeeded")
	}
	if _, _, err := store(t, 1).Lookup(ctx, a.Key, time.Time{}); !errors.Is(err, persistence.ErrReplayUnavailable) {
		t.Error("lookup without trusted time succeeded")
	}
}
