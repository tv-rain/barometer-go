package barometer

import (
	"context"
	"errors"

	"github.com/tv-rain/barometer-go/internal/execution"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/networkcontext"
	"github.com/tv-rain/barometer-go/internal/result"
)

var errPathUnavailable = errors.New("barometer: direct path unavailable")

// directPath is the production direct path (DEV-644). At reservation it reads
// the access context once, through the platform network capability, and binds
// the run's result decoder to that snapshot's direct path. Each attempt then
// captures its own snapshot, evaluates the task's network conditions against
// it, attaches it to the result with egress unavailable, and dials through a
// fresh direct dialer inside the metered attempt scope. Implementation is the
// trusted producer identity of the integration, never host or task input.
type directPath struct {
	context        *networkcontext.Direct
	implementation model.Implementation
}

// newDirectPath fails without a network capability, so a client that cannot
// read access context has no path and starts no run.
func newDirectPath(r runAdapters) (*directPath, error) {
	if absent(r.network) {
		return nil, errPathUnavailable
	}
	d, err := networkcontext.NewDirect(r.network)
	if err != nil {
		return nil, errPathUnavailable
	}
	return &directPath{context: d, implementation: r.implementation}, nil
}

func (p *directPath) direct(ctx context.Context, r pathRequest) (pathPlan, error) {
	snap, err := p.context.Capture(ctx)
	if err != nil {
		return pathPlan{}, errPathUnavailable
	}
	d, err := result.New(result.Binding{TaskID: r.TaskID, Output: result.Output{Test: model.Test{ID: r.Test.ID, Version: r.Test.Version, ResultVersion: result.Contract}},
		MaxAttempts: r.MaxAttempts, MaxDurationMS: r.MaxDurationMS, Implementation: p.implementation,
		ContextDecoder: p.context.Decoder(), Path: snap.Path(), AccessGeneration: snap.Generation()})
	if err != nil {
		return pathPlan{}, errPathUnavailable
	}
	return pathPlan{Implementation: p.implementation, Access: snap.Access(), Result: d, Context: p.context, Dialers: execution.DirectDialers{}}, nil
}
