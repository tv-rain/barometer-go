package barometer

import (
	"context"
	"sync"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/persistence"
)

// state follows the Lifecycle table in docs/public-api.md. transitioning and
// blocked together form its "transitioning or clearing incomplete" row.
type state uint8

const (
	uninitialized state = iota // the zero value carries no authority
	initialized
	transitioning // a lifecycle call is running without holding the lock
	blocked       // a transition did not complete; only clearing or close recover
	closed
)

// Client is the opaque lifecycle owner. It has no exported fields. A zero-value
// Client is uninitialized and usable; do not copy a Client after first use.
//
// The current implementation provides the lifecycle shell (DEV-600):
// initialization through reviewed internal adapters, bounded handle tables,
// close and the lifecycle part of clearing, and task admission (DEV-592). Runs
// (DEV-636, DEV-561), pending bundles (DEV-563) and persistent clearing
// (DEV-564) followed; proxy changes are added by a later task.
type Client struct {
	mu         sync.Mutex
	state      state
	blockedBy  code // set while state is blocked
	clearing   bool // a recovery clear of a closed client is running
	bound      *adapters
	rt         *runtime
	generation uint64
	// policyFloor is the highest valid policy generation this instance has
	// activated. Initialization below it is a rollback and fails closed.
	policyFloor uint64
	active      int           // reserved entry points of the current generation
	drained     chan struct{} // closed when active reaches zero during a transition
}

// NewClient returns an uninitialized client. It starts no measurement, creates
// no secret and performs no I/O.
func NewClient() *Client { return &Client{} }

// Initialize decodes the closed host configuration container and activates one
// generation over the reviewed internal adapters. The adapter-installed local
// policy and its target registrations stay immutable for that generation. If
// an earlier clear of the storage scope did not complete, including before a
// restart, it returns clearing_incomplete and blocks the client until recovery
// clearing succeeds. Any other failure, including a policy generation below
// one this instance has already activated, returns not_initialized and leaves
// the client uninitialized.
func (c *Client) Initialize(configJSON string) error { return c.initialize(configJSON) }

// AdmitTask runs every admission step for the exact task bytes against the
// adapter-installed delivery record that sourceRef names, and returns a local
// opaque task handle only on acceptance. It starts no measurement and performs
// no network activity. An exact redelivery returns the same handle while the
// current initialization lasts, without a new replay record or a later expiry.
// A rejection returns the code-only public admission class; a call outside the
// initialized state returns an API-local code.
func (c *Client) AdmitTask(taskJSON string, sourceRef string) (string, error) {
	return c.admitTask(taskJSON, sourceRef)
}

// Close stops new work, invalidates handles, delivery references and context
// state and releases runtime ownership. It deletes no replay or pending state
// and does not replace ClearMeasurementData. Repeated close succeeds.
func (c *Client) Close() error { return c.close() }

// ClearMeasurementData blocks new work, invalidates handles and delivery
// references and cancels active runs, then deletes Barometer-owned persistent
// state: pending bundles, run accounting, replay records and the linkage key.
// It uploads nothing, sends no request and creates no replacement secret.
// Success leaves the client uninitialized, or closed if it was closed. Any
// failure returns clearing_incomplete and keeps activity blocked, across a
// restart, until a later clear succeeds.
func (c *Client) ClearMeasurementData() error { return c.clearMeasurementData() }

// refusal is the error for a call that requires an uninitialized client.
func (c *Client) refusal() error {
	switch c.state {
	case initialized:
		return fail(alreadyInitialized)
	case transitioning:
		return fail(transitionInProgress)
	case blocked:
		return fail(c.blockedBy)
	case closed:
		return fail(clientClosed)
	}
	return nil
}

// bind installs trusted adapters on an uninitialized client. It is the only
// installation path and is unexported; there is no production caller yet.
func (c *Client) bind(a adapters) error {
	if c == nil || a.deadline < 0 || a.deadline > maxLifecycleDeadline {
		return fail(notInitialized)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.refusal(); err != nil {
		return err
	}
	if a.clearStorage == nil {
		a.clearStorage = storageClearer(&a)
	}
	c.bound = &a
	return nil
}

// marked reports whether the storage scope holds the clearing marker of an
// incomplete clear. An unreadable marker fails initialization closed.
func marked(a *adapters) (bool, error) {
	if a == nil || absent(a.records) {
		return false, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), a.lifecycleDeadline())
	defer cancel()
	return persistence.Clearing(ctx, a.records)
}

func (c *Client) initialize(configJSON string) (err error) {
	if c == nil {
		return fail(notInitialized)
	}
	c.mu.Lock()
	if err := c.refusal(); err != nil {
		c.mu.Unlock()
		return err
	}
	a := c.bound
	c.state = transitioning
	c.mu.Unlock()

	var rt *runtime
	clearing := false
	defer func() {
		if recover() != nil {
			rt, err = nil, fail(notInitialized)
		}
		c.mu.Lock()
		defer c.mu.Unlock()
		if clearing {
			// An earlier clear did not complete, possibly before a restart:
			// only recovery clearing or close may proceed.
			c.state, c.blockedBy = blocked, clearingIncomplete
			return
		}
		if rt == nil {
			c.state = uninitialized
			return
		}
		c.generation++
		rt.generation = c.generation
		c.rt, c.state = rt, initialized
		if a.policy.Valid() {
			c.policyFloor = max(c.policyFloor, a.policy.Generation())
		}
	}()
	switch found, err := marked(a); {
	case err != nil:
		return fail(notInitialized)
	case found:
		clearing = true
		return fail(clearingIncomplete)
	}
	if a != nil && a.policy.Valid() && a.policy.Generation() < c.policyFloor {
		return fail(notInitialized)
	}
	if rt = prepare(a, []byte(configJSON)); rt == nil {
		return fail(notInitialized)
	}
	return nil
}

// reserve admits one entry point under the current generation. Admission, run
// reservation and bundle access call it before using the runtime and release
// afterwards, so lifecycle transitions serialize with them.
func (c *Client) reserve() (*runtime, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	switch c.state {
	case initialized:
		c.active++
		return c.rt, nil
	case uninitialized:
		return nil, fail(notInitialized)
	case transitioning:
		return nil, fail(transitionInProgress)
	case blocked:
		return nil, fail(c.blockedBy)
	}
	return nil, fail(clientClosed)
}

func (c *Client) release() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.active--; c.active == 0 && c.drained != nil {
		close(c.drained)
		c.drained = nil
	}
}

// admitTask reserves the current generation, so it serializes with lifecycle
// transitions, and publishes a handle only after Admit has durably accepted the
// task. A panic below the facade becomes task_trust_unavailable with no handle.
// The step order of REQ-TRUST-0047, including the complete-task byte limit
// after delivery resolution, is the admitter's; the facade adds no check.
func (c *Client) admitTask(taskJSON, sourceRef string) (handle string, err error) {
	if c == nil {
		return "", fail(notInitialized)
	}
	rt, err := c.reserve()
	if err != nil {
		return "", err
	}
	defer c.release()
	defer func() {
		if recover() != nil {
			handle, err = "", trustUnavailable()
		}
	}()
	ctx, cancel := context.WithTimeout(rt.ctx, rt.deadline)
	defer cancel()
	admitted, err := rt.admitter.Admit(ctx, sourceRef, []byte(taskJSON))
	if err != nil {
		return "", err
	}
	// The replay record is durable. If a transition invalidated the table
	// meanwhile, no handle is published; a redelivery to a later generation
	// resolves to the same admitted task.
	handle, err = rt.tasks.publish(admitted, admitted.EvaluatedAt())
	if err == (opError{code: handleUnavailable}) {
		return "", fail(transitionInProgress)
	}
	return handle, err
}

// trustUnavailable is the admission class for a failure that admission cannot
// attribute to the task.
func trustUnavailable() error {
	e, err := task.NewAdmissionError(task.TrustUnavailable)
	if err != nil {
		return fail(notInitialized)
	}
	return e
}

// quiesceLocked starts a transition and returns a channel to wait on when
// reserved entry points are still running. The caller holds c.mu.
func (c *Client) quiesceLocked() <-chan struct{} {
	c.state = transitioning
	if c.active == 0 {
		c.drained = nil
		return nil
	}
	c.drained = make(chan struct{})
	return c.drained
}

// drain waits without holding the lock, for at most the lifecycle deadline.
func drain(done <-chan struct{}, d time.Duration) bool {
	if done == nil {
		return true
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-done:
		return true
	case <-t.C:
		return false
	}
}

func (c *Client) close() (err error) {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	if c.state == closed {
		c.mu.Unlock()
		return nil
	}
	if c.state == transitioning {
		c.mu.Unlock()
		return fail(transitionInProgress)
	}
	a, rt := c.bound, c.rt
	done := c.quiesceLocked()
	c.mu.Unlock()

	finished := false
	defer func() {
		if recover() != nil {
			finished, err = false, fail(cleanupIncomplete)
		}
		c.mu.Lock()
		defer c.mu.Unlock()
		if !finished {
			c.state, c.blockedBy = blocked, cleanupIncomplete
			return
		}
		// Only the clearing capability survives, so recovery clearing can
		// still reach the storage scope after close.
		c.rt, c.bound, c.state = nil, nil, closed
		if a != nil {
			c.bound = &adapters{clearStorage: a.clearStorage, deadline: a.deadline}
		}
	}()
	rt.invalidate()
	if a != nil && a.deliveries != nil {
		a.deliveries.Close()
	}
	if !drain(done, a.lifecycleDeadline()) {
		return fail(cleanupIncomplete)
	}
	finished = true
	return nil
}

func (c *Client) clearMeasurementData() (err error) {
	if c == nil {
		return fail(clearingIncomplete)
	}
	c.mu.Lock()
	if c.state == transitioning || c.clearing {
		c.mu.Unlock()
		return fail(transitionInProgress)
	}
	wasClosed := c.state == closed
	a, rt := c.bound, c.rt
	var done <-chan struct{}
	if wasClosed {
		c.clearing = true
	} else {
		done = c.quiesceLocked()
	}
	c.mu.Unlock()

	cleared := false
	defer func() {
		if recover() != nil {
			cleared, err = false, fail(clearingIncomplete)
		}
		c.mu.Lock()
		defer c.mu.Unlock()
		c.rt = nil
		switch {
		case wasClosed:
			c.clearing = false
		case cleared:
			c.state = uninitialized
		default:
			c.state, c.blockedBy = blocked, clearingIncomplete
		}
	}()
	rt.invalidate()
	if a != nil && a.deliveries != nil {
		a.deliveries.Invalidate()
	}
	if !drain(done, a.lifecycleDeadline()) || a == nil || a.clearStorage == nil {
		return fail(clearingIncomplete)
	}
	ctx, cancel := context.WithTimeout(context.Background(), a.lifecycleDeadline())
	defer cancel()
	if a.clearStorage(ctx) != nil {
		return fail(clearingIncomplete)
	}
	// Drop the cleared generation's in-memory results as far as Go allows;
	// its reservations have drained, so the store is idle.
	if rt != nil {
		rt.bundles.Forget()
	}
	cleared = true
	return nil
}
