package barometer

import (
	"context"
	"errors"

	"github.com/tv-rain/barometer-go/internal/persistence/pending"
)

// NextBundle returns the oldest unacknowledged, unexpired Result Bundle as
// UTF-8 JSON, or the empty string when none is pending. Repeated calls return
// the same bytes and identifier until the bundle is acknowledged or expires;
// no result is added to a returned bundle. It does not wait for new results,
// upload or acknowledge anything.
func (c *Client) NextBundle() (string, error) { return c.nextBundle() }

// AcknowledgeBundle deletes the pending bundle that bundleID names, once the
// host has established durable delivery. An unknown or already deleted
// identifier returns bundle_unavailable. A failed deletion returns
// storage_unavailable and is not an acknowledgement; a retry deletes it.
func (c *Client) AcknowledgeBundle(bundleID string) error { return c.acknowledgeBundle(bundleID) }

// nextBundle reserves the current generation, so it serializes with
// lifecycle transitions. A panic below the facade returns storage_unavailable.
// requirement-ref: REQ-LIMIT-0024
func (c *Client) nextBundle() (encoded string, err error) {
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
			encoded, err = "", fail(storageUnavailable)
		}
	}()
	ctx, cancel := context.WithTimeout(rt.ctx, rt.deadline)
	defer cancel()
	_, b, ok, err := rt.bundles.Next(ctx)
	if err != nil {
		return "", rt.storageError()
	}
	if !ok {
		return "", nil
	}
	return string(b), nil
}

// acknowledgeBundle reserves the current generation like nextBundle. A panic
// below the facade returns storage_unavailable, never success.
// requirement-ref: REQ-LIMIT-0024
func (c *Client) acknowledgeBundle(bundleID string) (err error) {
	if c == nil {
		return fail(notInitialized)
	}
	rt, err := c.reserve()
	if err != nil {
		return err
	}
	defer c.release()
	defer func() {
		if recover() != nil {
			err = fail(storageUnavailable)
		}
	}()
	ctx, cancel := context.WithTimeout(rt.ctx, rt.deadline)
	defer cancel()
	switch err := rt.bundles.Acknowledge(ctx, bundleID); {
	case err == nil:
		return nil
	case errors.Is(err, pending.ErrUnknown):
		return fail(bundleUnavailable)
	}
	return rt.storageError()
}

// storageError is the code of a failed pending-storage call: a transition
// that cancelled the generation, or unavailable storage.
func (rt *runtime) storageError() error {
	if rt.ctx.Err() != nil {
		return fail(transitionInProgress)
	}
	return fail(storageUnavailable)
}
