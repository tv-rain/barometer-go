package records_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/persistence"
	"github.com/tv-rain/barometer-go/internal/testkit/records"
)

var key = persistence.RecordKey{Class: persistence.ReplayClass, Name: [persistence.RecordNameSize]byte{1}}

func TestStoreFaultsAndRestart(t *testing.T) {
	ctx := context.Background()
	for name, tc := range map[string]struct {
		fault   records.Fault
		durable bool
		staged  int
	}{
		"before put": {records.FailBeforePut, false, 0},
		"crash":      {records.CrashDuringPut, false, 1},
		"after put":  {records.FailAfterPut, true, 0},
	} {
		s := records.New()
		s.Inject(tc.fault, true)
		if err := s.Put(ctx, key, []byte("v")); !errors.Is(err, persistence.ErrRecordsUnavailable) {
			t.Errorf("%s: fault not reported", name)
		}
		if _, ok, _ := s.Get(ctx, key); ok != tc.durable || s.Staged() != tc.staged {
			t.Errorf("%s: state wrong", name)
		}
		s.Restart()
		if _, ok, err := s.Get(ctx, key); err != nil || ok != tc.durable || s.Staged() != 0 {
			t.Errorf("%s: restart lost durable or kept staged state", name)
		}
	}
	s := records.New()
	if err := s.Put(ctx, key, []byte("v")); err != nil {
		t.Fatal(err)
	}
	s.Inject(records.FailBeforeDelete, true)
	if s.Delete(ctx, key) == nil || s.Len(persistence.ReplayClass) != 1 {
		t.Error("failed delete removed the record")
	}
	s.Restart()
	s.Inject(records.FailAfterDelete, true)
	if s.Delete(ctx, key) == nil || s.Len(persistence.ReplayClass) != 0 {
		t.Error("uncertain delete kept the record")
	}
	for _, f := range []records.Fault{records.FailGet, records.FailList} {
		s.Inject(f, true)
	}
	if _, _, err := s.Get(ctx, key); err == nil {
		t.Error("get fault not reported")
	}
	if _, err := s.List(ctx, persistence.ReplayClass); err == nil {
		t.Error("list fault not reported")
	}
}

func TestStoreBoundsAndCancellation(t *testing.T) {
	ctx := context.Background()
	s := records.New()
	if s.Put(ctx, key, make([]byte, persistence.MaxRecordValue+1)) == nil || s.Put(ctx, persistence.RecordKey{}, nil) == nil {
		t.Error("oversized value or unknown class stored")
	}
	v := []byte("v")
	_ = s.Put(ctx, key, v)
	v[0] = 'x'
	if got, _, _ := s.Get(ctx, key); string(got) != "v" {
		t.Error("store aliased the caller's bytes")
	}
	if keys, err := s.List(ctx, persistence.RunClass); err != nil || len(keys) != 0 {
		t.Error("list crossed classes")
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, _, err := s.Get(cancelled, key); err == nil {
		t.Error("cancelled get succeeded")
	}
	s.Inject(records.Stall, true)
	deadline, stop := context.WithTimeout(ctx, 10*time.Millisecond)
	defer stop()
	if _, err := s.List(deadline, persistence.ReplayClass); !errors.Is(err, context.DeadlineExceeded) {
		t.Error("stall outlived the deadline")
	}
}
