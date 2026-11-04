package persistence_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/persistence"
	recordstore "github.com/tv-rain/barometer-go/internal/testkit/records"
)

// countingRecords counts List calls, which only opening the state makes
// before any record exists.
type countingRecords struct {
	*recordstore.Store
	lists atomic.Int32
	fail  bool
	block chan struct{}
}

func (c *countingRecords) List(ctx context.Context, class persistence.RecordClass) ([]persistence.RecordKey, error) {
	c.lists.Add(1)
	if c.block != nil {
		select {
		case <-c.block:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if c.fail {
		return nil, persistence.ErrRecordsUnavailable
	}
	return c.Store.List(ctx, class)
}

func TestLazyStateOpensOnFirstUse(t *testing.T) {
	if _, err := persistence.NewLazyState(nil, 8); err == nil {
		t.Fatal("nil records accepted")
	}
	if _, err := persistence.NewLazyState(recordstore.New(), 0); err == nil {
		t.Fatal("zero capacity accepted")
	}
	r := &countingRecords{Store: recordstore.New(), fail: true}
	l, err := persistence.NewLazyState(r, 8)
	if err != nil {
		t.Fatal(err)
	}
	l.Forget() // unopened: no index, no I/O
	if r.lists.Load() != 0 {
		t.Fatal("construction or Forget performed I/O")
	}
	ctx := context.Background()
	now := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
	if _, err := l.Replay().HighestGeneration(ctx, now); !errors.Is(err, persistence.ErrStateUnavailable) {
		t.Fatal("failed open did not fail closed:", err)
	}
	r.fail = false
	if g, err := l.Replay().HighestGeneration(ctx, now); err != nil || g != 0 {
		t.Fatal("open did not recover:", err)
	}
	opened := r.lists.Load()
	key := persistence.ReplayKey{IssuerID: "synthetic-issuer", TaskID: "synthetic-task"}
	if _, ok, err := l.Runs().Lookup(ctx, key, now); err != nil || ok {
		t.Fatal("lookup:", err)
	}
	if r.lists.Load() != opened {
		t.Fatal("the state was opened twice")
	}
	l.Forget()

	// A caller waiting behind a stalled open honors its own deadline.
	r2 := &countingRecords{Store: recordstore.New(), block: make(chan struct{})}
	l2, _ := persistence.NewLazyState(r2, 8)
	first := make(chan error, 1)
	go func() {
		_, err := l2.Replay().HighestGeneration(ctx, now)
		first <- err
	}()
	for r2.lists.Load() == 0 {
		time.Sleep(time.Millisecond)
	}
	short, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
	defer cancel()
	if _, err := l2.Replay().HighestGeneration(short, now); err == nil {
		t.Fatal("waiting caller succeeded behind a stalled open")
	}
	close(r2.block)
	if err := <-first; err != nil {
		t.Fatal(err)
	}
}
