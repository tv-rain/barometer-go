package barometer

import (
	"context"

	"github.com/tv-rain/barometer-go/internal/model/task"
)

// replacePolicy is the only path that changes the active local policy and its
// target registrations (REQ-TRUST-0029 to REQ-TRUST-0032). It is unexported,
// like bind: the reviewed host adapter (DEV-455) authenticates and verifies a
// policy update before calling it, and no task, collector reply, target,
// proxy or context endpoint reaches it. It replaces the policy by quiescing
// the current generation and reinitializing a new one; nothing is updated in
// place.
//
// The steps and their failure outcomes:
//
//  1. Validate while the current generation stays active: the policy must be
//     complete and valid, its generation strictly greater than the active
//     one, not below the highest generation retained replay records were
//     accepted under, and the successor generation must build. Failure
//     returns policy_rejected and leaves the previous complete policy active
//     with its handles and delivery references.
//  2. Cancel: cancel the current generation's adapter context, so in-flight
//     admission stops, and invalidate its task and run handles. Tasks still
//     permitted are redelivered and admitted again under the new policy; the
//     others are rejected at admission.
//  3. Invalidate delivery records and activate the new generation in the
//     delivery store, so earlier references never resolve. No Network
//     Context cache exists yet.
//  4. Wait for reserved entry points, up to the lifecycle deadline.
//  5. Reinitialize: activate the prepared generation before any new-policy
//     admission runs.
//
// A failure after step 1 leaves the client uninitialized with no policy, so
// a later Initialize admits nothing (task_policy_unavailable) until a
// complete policy is bound again; it returns not_initialized, or
// cleanup_incomplete and stays blocked while reserved work has not finished.
// Neither outcome activates part of the new policy or an earlier generation.
func (c *Client) replacePolicy(p task.LocalPolicy) (err error) {
	if c == nil {
		return fail(notInitialized)
	}
	c.mu.Lock()
	if c.state != initialized {
		defer c.mu.Unlock()
		if c.state == uninitialized {
			return fail(notInitialized)
		}
		return c.refusal()
	}
	a, current := c.bound, c.rt
	active := max(c.policyFloor, a.policy.Generation())
	c.state = transitioning
	c.mu.Unlock()

	var next *runtime
	quiesced, finished := false, false
	defer func() {
		if recover() != nil {
			finished, err = false, fail(policyRejected)
			if quiesced {
				err = fail(notInitialized)
			}
		}
		c.mu.Lock()
		defer c.mu.Unlock()
		switch {
		case !quiesced:
			c.state = initialized
		case finished:
			c.generation++
			next.generation = c.generation
			c.proxyGeneration = next.proxies.generation
			bound := *a
			bound.policy = p
			c.bound, c.rt, c.state = &bound, next, initialized
			c.policyFloor = max(c.policyFloor, p.Generation())
		default:
			if next != nil {
				next.cancel()
			}
			bound := *a
			bound.policy = task.LocalPolicy{}
			c.bound, c.rt = &bound, nil
			if c.active > 0 {
				c.state, c.blockedBy, err = blocked, cleanupIncomplete, fail(cleanupIncomplete)
			} else {
				c.state, err = uninitialized, fail(notInitialized)
			}
		}
	}()

	// Step 1.
	if next = validSuccessor(a, current, p, active); next == nil {
		return fail(policyRejected)
	}

	// Steps 2 and 3.
	c.mu.Lock()
	done := c.quiesceLocked()
	quiesced = true
	c.mu.Unlock()
	current.invalidate()
	invalidated := a.deliveries.ReplacePolicy(p.Generation()) == nil
	if !invalidated {
		a.deliveries.Invalidate()
	}

	// Step 4.
	if !drain(done, a.lifecycleDeadline()) {
		return fail(cleanupIncomplete)
	}
	if !invalidated {
		return fail(notInitialized)
	}

	// Step 5 happens in the deferred activation.
	finished = true
	return nil
}

// validSuccessor returns the prepared generation for p, or nil when p is not
// a complete valid policy above active, when trusted time or the retained
// replay generations are unavailable, or when retained acceptances are above
// p's generation, so activating p would be a rollback.
func validSuccessor(a *adapters, current *runtime, p task.LocalPolicy, active uint64) *runtime {
	if !p.Valid() || p.Generation() <= active {
		return nil
	}
	ctx, cancel := context.WithTimeout(current.ctx, a.lifecycleDeadline())
	defer cancel()
	now, err := a.clock.Now(ctx)
	if err != nil || now.IsZero() {
		return nil
	}
	highest, err := a.replay.HighestGeneration(ctx, now)
	if err != nil || highest > p.Generation() {
		return nil
	}
	next := successor(a, current.init, p)
	if next != nil {
		// Dynamic proxy changes must survive policy replacement; the
		// initialization object intentionally stays immutable for bundle linkage.
		next.proxies.group.cancel(nil)
		next.proxies = current.proxies.successor(next.ctx)
	}
	return next
}
