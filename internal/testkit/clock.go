package testkit

import (
	"context"
	"errors"
	"slices"
	"sync"
	"time"
)

// Clock is a manually advanced, concurrency-safe time source. It creates no
// operating-system timers or goroutines. Consumers inject Now into their own
// time-dependent logic. It does not drive context.WithDeadline; AfterFunc
// timers fire only when Advance reaches them.
type Clock struct {
	mu     sync.Mutex
	now    time.Time
	timers []*timer
}

type timer struct {
	at   time.Time
	f    func()
	done bool // fired or stopped
}

func NewClock(now time.Time) *Clock { return &Clock{now: now} }

func (c *Clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

// Advance moves the clock forward and then runs, in the caller's goroutine and
// in time order, every AfterFunc timer it has reached: a timer fires when the
// clock is at or after its time.
func (c *Clock) Advance(d time.Duration) error {
	if d < 0 {
		return errors.New("testkit: clock cannot move backwards")
	}
	c.mu.Lock()
	c.now = c.now.Add(d)
	var due []*timer
	kept := c.timers[:0]
	for _, t := range c.timers {
		switch {
		case t.done:
		case !t.at.After(c.now):
			t.done = true
			due = append(due, t)
		default:
			kept = append(kept, t)
		}
	}
	c.timers = kept
	c.mu.Unlock()
	slices.SortStableFunc(due, func(a, b *timer) int { return a.at.Compare(b.at) })
	for _, t := range due {
		t.f()
	}
	return nil
}

// AfterFunc arranges for f to run when Advance reaches now+d. A non-positive d
// fires on the next Advance, including Advance(0). stop reports whether it
// prevented f from running.
func (c *Clock) AfterFunc(d time.Duration, f func()) (stop func() bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := &timer{at: c.now.Add(max(d, 0)), f: f}
	c.timers = append(c.timers, t)
	return func() bool {
		c.mu.Lock()
		defer c.mu.Unlock()
		if t.done {
			return false
		}
		t.done = true
		return true
	}
}

// Timers reports how many AfterFunc timers are armed and have neither fired
// nor been stopped, so a test can see that a consumer armed one.
func (c *Clock) Timers() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	n := 0
	for _, t := range c.timers {
		if !t.done {
			n++
		}
	}
	return n
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

// WaitClock adds a deterministic Wait to Clock for engines that wait on an
// injected clock. Wait returns the context error if ctx is done; otherwise it
// records the requested duration and advances the clock by it at once, firing
// the timers it reaches. It creates no operating-system timers or goroutines,
// so tests assert delays without sleeping.
type WaitClock struct {
	*Clock
	mu    sync.Mutex
	waits []time.Duration
}

func NewWaitClock(now time.Time) *WaitClock { return &WaitClock{Clock: NewClock(now)} }

func (c *WaitClock) Wait(ctx context.Context, d time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := c.Advance(d); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.waits = append(c.waits, d)
	return nil
}

// Waits returns a copy of the durations waited, in order.
func (c *WaitClock) Waits() []time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]time.Duration(nil), c.waits...)
}
