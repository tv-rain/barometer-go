package persistence_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/persistence"
	"github.com/tv-rain/barometer-go/internal/testkit/records"
)

var stateNow = time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)

func replayRecord(t *testing.T, taskID string, generation uint64, valid time.Duration) persistence.ReplayRecord {
	t.Helper()
	r, err := persistence.NewReplayRecord(persistence.ReplayKey{IssuerID: "issuer", TaskID: taskID}, sha256.Sum256([]byte(taskID)), generation, task.Remote, task.TimestampFromTime(stateNow.Add(valid)), stateNow)
	if err != nil {
		t.Fatal("record rejected")
	}
	return r
}

func runRecord() persistence.RunRecord {
	return persistence.RunRecord{Anchor: stateNow.Add(1500 * time.Millisecond), Runs: 1, Next: 2, NextDue: stateNow.Add(time.Hour + 250*time.Millisecond)}
}

func replayKey(k persistence.ReplayKey) persistence.RecordKey {
	return persistence.RecordKey{Class: persistence.ReplayClass, Name: persistence.RecordNameOf(k)}
}

func runKey(k persistence.ReplayKey) persistence.RecordKey {
	return persistence.RecordKey{Class: persistence.RunClass, Name: persistence.RecordNameOf(k)}
}

func open(t *testing.T, r *records.Store, capacity int, now time.Time) *persistence.State {
	t.Helper()
	s, err := persistence.OpenState(context.Background(), r, capacity, now)
	if err != nil {
		t.Fatal("state rejected:", err)
	}
	return s
}

func sameRun(a, b persistence.RunRecord) bool {
	return a.Anchor.Equal(b.Anchor) && a.NextDue.Equal(b.NextDue) && a.Runs == b.Runs && a.Next == b.Next
}

func TestOpenStateRejectsInvalidInputs(t *testing.T) {
	ctx := context.Background()
	for name, tc := range map[string]struct {
		records  persistence.Records
		capacity int
		now      time.Time
	}{
		"no records":    {nil, 1, stateNow},
		"zero capacity": {records.New(), 0, stateNow},
		"over capacity": {records.New(), persistence.MaxClassRecords + 1, stateNow},
		"no time":       {records.New(), 1, time.Time{}},
	} {
		if s, err := persistence.OpenState(ctx, tc.records, tc.capacity, tc.now); s != nil || !errors.Is(err, persistence.ErrStateUnavailable) {
			t.Errorf("%s: accepted", name)
		}
	}
	if s := open(t, records.New(), persistence.MaxClassRecords, stateNow); fmt.Sprintf("%v %+v %#v", s, s, s) != "persistence.State persistence.State persistence.State" {
		t.Error("state formatting exposes its contents")
	}
}

// requirement-ref: REQ-TRUST-0021
// requirement-ref: REQ-TRUST-0022
func TestStateReplayCommitsConditionallyWithinCapacity(t *testing.T) {
	ctx := context.Background()
	r := records.New()
	replay := open(t, r, 2, stateNow).Replay()
	a, b, c := replayRecord(t, "a", 2, time.Hour), replayRecord(t, "b", 3, 2*time.Hour), replayRecord(t, "c", 2, time.Hour)
	if got, created, err := replay.Commit(ctx, a, stateNow); err != nil || !created || got != a {
		t.Fatal("first commit failed")
	}
	other := a
	other.Digest = sha256.Sum256([]byte("other"))
	if got, created, err := replay.Commit(ctx, other, stateNow); err != nil || created || got != a {
		t.Fatal("conditional commit overwrote an existing record")
	}
	if _, _, err := replay.Commit(ctx, b, stateNow); err != nil {
		t.Fatal("second commit failed")
	}
	if _, _, err := replay.Commit(ctx, c, stateNow); !errors.Is(err, persistence.ErrReplayUnavailable) || r.Len(persistence.ReplayClass) != 2 {
		t.Fatal("full store evicted or exceeded capacity")
	}
	if got, ok, err := replay.Lookup(ctx, a.Key, stateNow); err != nil || !ok || got != a {
		t.Fatal("lookup lost the record")
	}
	if _, ok, err := replay.Lookup(ctx, c.Key, stateNow); err != nil || ok {
		t.Fatal("absent record reported")
	}
	if g, err := replay.HighestGeneration(ctx, stateNow); err != nil || g != 3 {
		t.Fatal("highest generation wrong")
	}
	if _, _, err := replay.Commit(ctx, persistence.ReplayRecord{}, stateNow); !errors.Is(err, persistence.ErrReplayRecord) {
		t.Fatal("invalid record committed")
	}
	// Record names are digests: no issuer or task identifier is a key.
	keys, _ := r.List(ctx, persistence.ReplayClass)
	for _, k := range keys {
		if bytes.Contains(k.Name[:], []byte("issuer")) || k.Name != persistence.RecordNameOf(a.Key) && k.Name != persistence.RecordNameOf(b.Key) {
			t.Fatal("unexpected record name")
		}
	}
}

// Retention is enforced on access and on initialization, before, at and after
// its end on the injected clock; run records go with their replay record.
// requirement-ref: REQ-LIMIT-0023
func TestStateRetentionBoundaries(t *testing.T) {
	ctx := context.Background()
	a := replayRecord(t, "a", 2, time.Hour)
	end := a.RetainUntil
	for name, tc := range map[string]struct {
		at       time.Time
		retained bool
	}{
		"before": {end.Add(-time.Nanosecond), true},
		"at":     {end, false},
		"after":  {end.Add(time.Second), false},
	} {
		// On access.
		r := records.New()
		s := open(t, r, 4, stateNow)
		if _, _, err := s.Replay().Commit(ctx, a, stateNow); err != nil {
			t.Fatal("commit failed")
		}
		if err := s.Runs().Put(ctx, a.Key, runRecord(), stateNow); err != nil {
			t.Fatal("run put failed")
		}
		if g, err := s.Replay().HighestGeneration(ctx, tc.at); err != nil || (g == 2) != tc.retained {
			t.Errorf("%s: generation %d", name, g)
		}
		if _, ok, err := s.Runs().Lookup(ctx, a.Key, tc.at); err != nil || ok != tc.retained {
			t.Errorf("%s: run lookup retained %v", name, ok)
		}
		_, ok, err := s.Replay().Lookup(ctx, a.Key, tc.at)
		if err != nil || ok != tc.retained {
			t.Errorf("%s: replay lookup retained %v", name, ok)
		}
		want := map[bool]int{true: 1, false: 0}[tc.retained]
		if r.Len(persistence.ReplayClass) != want || r.Len(persistence.RunClass) != want {
			t.Errorf("%s: overdue records not deleted", name)
		}

		// On initialization after a restart.
		r = records.New()
		s = open(t, r, 4, stateNow)
		_, _, _ = s.Replay().Commit(ctx, a, stateNow)
		_ = s.Runs().Put(ctx, a.Key, runRecord(), stateNow)
		r.Restart()
		open(t, r, 4, tc.at)
		if r.Len(persistence.ReplayClass) != want || r.Len(persistence.RunClass) != want {
			t.Errorf("%s: initialization kept overdue records", name)
		}
	}
	// A run put after its replay record is overdue has no accounting.
	s := open(t, records.New(), 4, stateNow)
	_, _, _ = s.Replay().Commit(ctx, a, stateNow)
	if err := s.Runs().Put(ctx, a.Key, runRecord(), end); !errors.Is(err, persistence.ErrRunsUnavailable) {
		t.Error("run recorded past its replay retention")
	}
}

func TestStateSurvivesRestart(t *testing.T) {
	ctx := context.Background()
	r := records.New()
	s := open(t, r, 4, stateNow)
	a, b := replayRecord(t, "a", 2, time.Hour), replayRecord(t, "b", 5, 2*time.Hour)
	_, _, _ = s.Replay().Commit(ctx, a, stateNow)
	_, _, _ = s.Replay().Commit(ctx, b, stateNow)
	run := runRecord()
	if err := s.Runs().Put(ctx, a.Key, run, stateNow); err != nil {
		t.Fatal("run put failed")
	}
	r.Restart()
	s = open(t, r, 4, stateNow.Add(time.Minute))
	if got, ok, err := s.Replay().Lookup(ctx, a.Key, stateNow.Add(time.Minute)); err != nil || !ok || got != a {
		t.Fatal("replay record lost across restart")
	}
	if g, err := s.Replay().HighestGeneration(ctx, stateNow.Add(time.Minute)); err != nil || g != 5 {
		t.Fatal("generation lost across restart")
	}
	if got, ok, err := s.Runs().Lookup(ctx, a.Key, stateNow.Add(time.Minute)); err != nil || !ok || !sameRun(got, run) {
		t.Fatal("run record lost across restart")
	}
	if _, ok, err := s.Runs().Lookup(ctx, b.Key, stateNow.Add(time.Minute)); err != nil || ok {
		t.Fatal("absent run record reported")
	}
	// A run record without a replay record is an orphan and is deleted.
	if err := r.Delete(ctx, replayKey(a.Key)); err != nil {
		t.Fatal(err)
	}
	r.Restart()
	s = open(t, r, 4, stateNow.Add(time.Minute))
	if r.Len(persistence.RunClass) != 0 {
		t.Fatal("orphan run record kept")
	}
	if _, ok, err := s.Runs().Lookup(ctx, a.Key, stateNow.Add(time.Minute)); err != nil || ok {
		t.Fatal("orphan run record reported")
	}
	if err := s.Runs().Put(ctx, a.Key, run, stateNow.Add(time.Minute)); !errors.Is(err, persistence.ErrRunsUnavailable) {
		t.Fatal("run recorded without a replay record")
	}
}

// Failed, crashed and uncertain writes: a commit is durable as a whole or not
// at all, and an uncertain one is accounted for as if it succeeded.
func TestStateWriteFaults(t *testing.T) {
	ctx := context.Background()
	a := replayRecord(t, "a", 4, time.Hour)
	for name, tc := range map[string]struct {
		fault   records.Fault
		durable bool
	}{
		"before write": {records.FailBeforePut, false},
		"crash":        {records.CrashDuringPut, false},
		"after write":  {records.FailAfterPut, true},
	} {
		r := records.New()
		s := open(t, r, 4, stateNow)
		r.Inject(tc.fault, true)
		if _, _, err := s.Replay().Commit(ctx, a, stateNow); !errors.Is(err, persistence.ErrReplayUnavailable) {
			t.Errorf("%s: commit fault not reported", name)
		}
		if (r.Len(persistence.ReplayClass) == 1) != tc.durable {
			t.Errorf("%s: durable state wrong", name)
		}
		r.Inject(tc.fault, false)
		if tc.fault != records.FailBeforePut {
			// The possibly durable record still counts for the generation.
			if g, err := s.Replay().HighestGeneration(ctx, stateNow); err != nil || g != 4 {
				t.Errorf("%s: uncertain commit not accounted", name)
			}
		}
		_, ok, err := s.Replay().Lookup(ctx, a.Key, stateNow)
		if err != nil || ok != tc.durable {
			t.Errorf("%s: lookup after fault", name)
		}
		r.Restart()
		s = open(t, r, 4, stateNow)
		if r.Staged() != 0 {
			t.Errorf("%s: staged bytes survived restart", name)
		}
		got, created, err := s.Replay().Commit(ctx, a, stateNow)
		if err != nil || got != a || created == tc.durable {
			t.Errorf("%s: retried commit wrong", name)
		}
	}

	// Run writes.
	for name, tc := range map[string]struct {
		fault   records.Fault
		durable bool
	}{
		"before write": {records.FailBeforePut, false},
		"crash":        {records.CrashDuringPut, false},
		"after write":  {records.FailAfterPut, true},
	} {
		r := records.New()
		s := open(t, r, 4, stateNow)
		_, _, _ = s.Replay().Commit(ctx, a, stateNow)
		r.Inject(tc.fault, true)
		if err := s.Runs().Put(ctx, a.Key, runRecord(), stateNow); !errors.Is(err, persistence.ErrRunsUnavailable) {
			t.Errorf("run %s: fault not reported", name)
		}
		r.Inject(tc.fault, false)
		if _, ok, err := s.Runs().Lookup(ctx, a.Key, stateNow); err != nil || ok != tc.durable {
			t.Errorf("run %s: lookup after fault", name)
		}
	}
	// An invalid run record is never written.
	r := records.New()
	s := open(t, r, 4, stateNow)
	_, _, _ = s.Replay().Commit(ctx, a, stateNow)
	if err := s.Runs().Put(ctx, a.Key, persistence.RunRecord{}, stateNow); !errors.Is(err, persistence.ErrRunsUnavailable) || r.Len(persistence.RunClass) != 0 {
		t.Error("invalid run record written")
	}
}

// A failed delete of an overdue record fails the operation instead of using
// the record; the next operation retries the purge.
func TestStateDeleteFaults(t *testing.T) {
	ctx := context.Background()
	a := replayRecord(t, "a", 2, time.Hour)
	for _, fault := range []records.Fault{records.FailBeforeDelete, records.FailAfterDelete} {
		r := records.New()
		s := open(t, r, 4, stateNow)
		_, _, _ = s.Replay().Commit(ctx, a, stateNow)
		_ = s.Runs().Put(ctx, a.Key, runRecord(), stateNow)
		r.Inject(fault, true)
		if _, _, err := s.Replay().Lookup(ctx, a.Key, a.RetainUntil); !errors.Is(err, persistence.ErrReplayUnavailable) {
			t.Error("failed purge not reported")
		}
		if _, err := s.Replay().HighestGeneration(ctx, a.RetainUntil); err == nil && fault == records.FailBeforeDelete {
			t.Error("generation read with an overdue record kept")
		}
		if _, err := persistence.OpenState(ctx, r, 4, a.RetainUntil); fault == records.FailBeforeDelete && !errors.Is(err, persistence.ErrStateUnavailable) {
			t.Error("initialization ignored a failed purge")
		}
		r.Inject(fault, false)
		if _, ok, err := s.Replay().Lookup(ctx, a.Key, a.RetainUntil); err != nil || ok || r.Len(persistence.ReplayClass) != 0 || r.Len(persistence.RunClass) != 0 {
			t.Error("purge not retried")
		}
	}
}

func TestStateReadFaultsAndCorruption(t *testing.T) {
	ctx := context.Background()
	a, b := replayRecord(t, "a", 2, time.Hour), replayRecord(t, "b", 2, time.Hour)
	setup := func() (*records.Store, *persistence.State) {
		r := records.New()
		s := open(t, r, 4, stateNow)
		_, _, _ = s.Replay().Commit(ctx, a, stateNow)
		_ = s.Runs().Put(ctx, a.Key, runRecord(), stateNow)
		return r, s
	}
	r, s := setup()
	r.Inject(records.FailGet, true)
	if _, _, err := s.Replay().Lookup(ctx, a.Key, stateNow); !errors.Is(err, persistence.ErrReplayUnavailable) {
		t.Error("replay read fault not reported")
	}
	if _, _, err := s.Replay().Commit(ctx, a, stateNow); !errors.Is(err, persistence.ErrReplayUnavailable) {
		t.Error("commit read fault not reported")
	}
	if _, _, err := s.Runs().Lookup(ctx, a.Key, stateNow); !errors.Is(err, persistence.ErrRunsUnavailable) {
		t.Error("run read fault not reported")
	}
	if _, err := persistence.OpenState(ctx, r, 4, stateNow); !errors.Is(err, persistence.ErrStateUnavailable) {
		t.Error("initialization ignored a read fault")
	}
	r.Inject(records.FailGet, false)
	r.Inject(records.FailList, true)
	if _, err := persistence.OpenState(ctx, r, 4, stateNow); !errors.Is(err, persistence.ErrStateUnavailable) {
		t.Error("initialization ignored a list fault")
	}

	good, _ := r.Value(replayKey(a.Key))
	goodRun, _ := r.Value(runKey(a.Key))
	for name, corrupt := range map[string]func(*records.Store){
		"truncated":     func(r *records.Store) { r.Corrupt(replayKey(a.Key), good[:len(good)-1]) },
		"trailing byte": func(r *records.Store) { r.Corrupt(replayKey(a.Key), append(append([]byte{}, good...), 0)) },
		"empty":         func(r *records.Store) { r.Corrupt(replayKey(a.Key), nil) },
		"wrong format":  func(r *records.Store) { v := append([]byte{}, good...); v[0] = 9; r.Corrupt(replayKey(a.Key), v) },
		"wrong class":   func(r *records.Store) { r.Corrupt(replayKey(a.Key), goodRun) },
		"zero generation": func(r *records.Store) {
			v := append([]byte{}, good...)
			clear(v[len(v)-25 : len(v)-17])
			r.Corrupt(replayKey(a.Key), v)
		},
		"wrong name": func(r *records.Store) { r.Corrupt(replayKey(b.Key), good) },
	} {
		r, s := setup()
		corrupt(r)
		key := a.Key
		if name == "wrong name" {
			key = b.Key
		}
		if _, ok, err := s.Replay().Lookup(ctx, key, stateNow); ok || !errors.Is(err, persistence.ErrReplayCorrupt) {
			t.Errorf("%s: lookup did not report invalid state", name)
		}
		if _, _, err := s.Replay().Commit(ctx, replayRecord(t, key.TaskID, 2, time.Hour), stateNow); !errors.Is(err, persistence.ErrReplayCorrupt) {
			t.Errorf("%s: commit did not report invalid state", name)
		}
		if _, err := persistence.OpenState(ctx, r, 4, stateNow); !errors.Is(err, persistence.ErrStateUnavailable) {
			t.Errorf("%s: initialization accepted invalid state", name)
		}
	}
	for name, corrupt := range map[string]func(*records.Store){
		"truncated":   func(r *records.Store) { r.Corrupt(runKey(a.Key), goodRun[:len(goodRun)-1]) },
		"wrong class": func(r *records.Store) { r.Corrupt(runKey(a.Key), good) },
		"runs past next": func(r *records.Store) {
			v := append([]byte{}, goodRun...)
			v[len(v)-18] = 3
			r.Corrupt(runKey(a.Key), v)
		},
		"retention mismatch": func(r *records.Store) {
			v := append([]byte{}, goodRun...)
			v[len(v)-1]++
			r.Corrupt(runKey(a.Key), v)
		},
	} {
		r, s := setup()
		corrupt(r)
		if name != "retention mismatch" {
			if _, ok, err := s.Runs().Lookup(ctx, a.Key, stateNow); ok || !errors.Is(err, persistence.ErrRunsUnavailable) {
				t.Errorf("run %s: lookup did not report invalid state", name)
			}
		}
		if _, err := persistence.OpenState(ctx, r, 4, stateNow); !errors.Is(err, persistence.ErrStateUnavailable) {
			t.Errorf("run %s: initialization accepted invalid state", name)
		}
	}
}

func TestStateHonorsCancellationAndDeadlines(t *testing.T) {
	a := replayRecord(t, "a", 2, time.Hour)
	r := records.New()
	s := open(t, r, 4, stateNow)
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := s.Replay().Commit(cancelled, a, stateNow); !errors.Is(err, persistence.ErrReplayUnavailable) || r.Puts() != 0 {
		t.Error("cancelled commit proceeded")
	}
	if _, _, err := s.Runs().Lookup(cancelled, a.Key, stateNow); !errors.Is(err, persistence.ErrRunsUnavailable) {
		t.Error("cancelled run lookup proceeded")
	}
	if _, _, err := s.Replay().Lookup(context.Background(), a.Key, time.Time{}); !errors.Is(err, persistence.ErrReplayUnavailable) {
		t.Error("lookup without trusted time proceeded")
	}
	r.Inject(records.Stall, true)
	ctx, stop := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer stop()
	start := time.Now()
	if _, _, err := s.Replay().Commit(ctx, a, stateNow); !errors.Is(err, persistence.ErrReplayUnavailable) || time.Since(start) > 5*time.Second {
		t.Error("stalled storage outlived the deadline")
	}
	if _, err := persistence.OpenState(ctx, r, 4, stateNow); !errors.Is(err, persistence.ErrStateUnavailable) {
		t.Error("stalled initialization succeeded")
	}
}

func TestStateConcurrentCallers(t *testing.T) {
	ctx := context.Background()
	r := records.New()
	s := open(t, r, 64, stateNow)
	same := replayRecord(t, "same", 2, time.Hour)
	var wg sync.WaitGroup
	var mu sync.Mutex
	created := 0
	for i := range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, c, err := s.Replay().Commit(ctx, same, stateNow); err == nil && c {
				mu.Lock()
				created++
				mu.Unlock()
			}
			own := replayRecord(t, fmt.Sprintf("t%d", i), uint64(i+1), time.Hour)
			if _, _, err := s.Replay().Commit(ctx, own, stateNow); err != nil {
				t.Error("commit failed")
			}
			if err := s.Runs().Put(ctx, own.Key, runRecord(), stateNow); err != nil {
				t.Error("run put failed")
			}
			_, _, _ = s.Runs().Lookup(ctx, own.Key, stateNow)
			_, _ = s.Replay().HighestGeneration(ctx, stateNow)
		}()
	}
	wg.Wait()
	if created != 1 || r.Len(persistence.ReplayClass) != 17 || r.Len(persistence.RunClass) != 16 {
		t.Fatalf("created %d replay %d runs %d", created, r.Len(persistence.ReplayClass), r.Len(persistence.RunClass))
	}
	if g, err := s.Replay().HighestGeneration(ctx, stateNow); err != nil || g != 16 {
		t.Fatal("generation wrong after concurrent commits")
	}
}

// Stored values hold exactly the record fields: no task bytes or other values.
// requirement-ref: REQ-TRUST-0023
func TestStateStoresOnlyRecordFields(t *testing.T) {
	ctx := context.Background()
	r := records.New()
	s := open(t, r, 4, stateNow)
	a := replayRecord(t, "a", 2, time.Hour)
	_, _, _ = s.Replay().Commit(ctx, a, stateNow)
	_ = s.Runs().Put(ctx, a.Key, runRecord(), stateNow)
	replay, _ := r.Value(replayKey(a.Key))
	run, _ := r.Value(runKey(a.Key))
	keyBytes := 3 + len(a.Key.IssuerID) + 1 + len(a.Key.TaskID)
	if len(replay) != keyBytes+32+8+1+8+8 || len(run) != keyBytes+8+2+8+8 {
		t.Fatalf("stored sizes %d %d", len(replay), len(run))
	}
}
