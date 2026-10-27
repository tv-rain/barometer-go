package net01

import (
	"context"
	"errors"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/observation"
)

// Runner performs one network-state read per attempt. It holds no state, so
// its zero value is ready and safe for concurrent attempts.
//
// The outcome is snapshot, platform_unavailable or platform_error. The
// registered inconclusive outcome is never produced: the public contract does
// not yet define which coarse combinations are internally inconsistent, so
// every coherent read is a snapshot. A generation change, cancellation,
// deadline or suspension before the outcome is recorded is the execution
// layer's, never a NET-01 outcome (REQ-NET-01-0027).
type Runner struct{}

// requirement-ref: REQ-NET-01-0006
// requirement-ref: REQ-NET-01-0007
// requirement-ref: REQ-NET-01-0008
// requirement-ref: REQ-NET-01-0009
// requirement-ref: REQ-NET-01-0021
// requirement-ref: REQ-NET-01-0026
func (Runner) Run(ctx context.Context, in measurement.Input) (measurement.Output, error) {
	if _, ok := in.Parameters.(Parameters); !ok || in.Network == nil || in.RecordOutcome == nil || ctx.Err() != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	access, err := in.Network(ctx)
	if ctx.Err() != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	out := measurement.Output{Observations: []observation.Observation{}}
	switch {
	case errors.Is(err, measurement.ErrNoNetworkState):
		out.Outcome = "platform_unavailable"
	case err != nil:
		out.Outcome = "platform_error"
	case access.Value().NetworkChanged:
		return measurement.Output{}, measurement.ErrRunner
	default:
		out.Outcome, out.Access = "snapshot", &access
	}
	// The outcome is fixed before the watch and handles are released
	// (GO-SPEC-ISSUE-004).
	if !in.RecordOutcome() {
		return measurement.Output{}, measurement.ErrRunner
	}
	return out, nil
}
