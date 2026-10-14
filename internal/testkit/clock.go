package testkit

import (
	"errors"
	"sync"
	"time"
)

// Clock is a manually advanced, concurrency-safe time source. It creates no
// timers or goroutines. Consumers inject Now into their own time-dependent logic.
// It does not drive context.WithDeadline or operating-system timers.
type Clock struct {
	mu  sync.Mutex
	now time.Time
}

func NewClock(now time.Time) *Clock { return &Clock{now: now} }

func (c *Clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *Clock) Advance(d time.Duration) error {
	if d < 0 {
		return errors.New("testkit: clock cannot move backwards")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
	return nil
}

// TimeBoundary supplies before/at/after values without wall-clock sleeps. The
// consumer chooses the contract's equality behavior; this helper does not.
type TimeBoundary struct {
	Before, At, After time.Time
}

func TimeBoundaries(at time.Time, quantum time.Duration) (TimeBoundary, error) {
	if quantum <= 0 {
		return TimeBoundary{}, errors.New("testkit: time quantum must be positive")
	}
	return TimeBoundary{at.Add(-quantum), at, at.Add(quantum)}, nil
}
