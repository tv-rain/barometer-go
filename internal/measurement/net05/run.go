package net05

import (
	"context"
	"crypto/x509"
	"sync"

	"github.com/tv-rain/barometer-go/internal/measurement"
)

// Runner is client-generation owned. It serializes whole profiles, including
// profiles on different access generations, conservatively enforcing one run
// per generation. Waiting is bounded by each caller's attempt context.
// Roots follows the same trusted certificate policy as the primitive runners.
type Runner struct {
	Roots  *x509.CertPool
	mu     sync.Mutex
	active chan struct{}
}

// requirement-ref: REQ-NET-05-0006
// requirement-ref: REQ-NET-05-0009
// requirement-ref: REQ-NET-05-0037
// requirement-ref: REQ-NET-05-0040
// requirement-ref: REQ-NET-05-0045
// requirement-ref: REQ-NET-05-0047
// requirement-ref: REQ-NET-05-0061
func (r *Runner) Run(ctx context.Context, in measurement.Input) (measurement.Output, error) {
	p, ok := in.Parameters.(Parameters)
	if !ok || r == nil || in.Composite == nil || in.Composite.Step == nil || in.CheckNetwork == nil || in.RecordOutcome == nil || in.Clock == nil || len(p.steps) < 1 || len(p.steps) > 8 || !r.acquire(ctx) {
		return measurement.Output{}, measurement.ErrRunner
	}
	defer r.release()
	c := &measurement.CompositeOutput{ID: p.id, Version: p.version, Validation: "unknown", Steps: []measurement.StepOutput{}}
	out := measurement.Output{Composite: c, Outcome: "incomplete"}
	switch in.Composite.Validation {
	case "yes":
		c.Validation = "validated"
	case "no":
		c.Validation = "not_validated"
	}
	stop := func(next int) (measurement.Output, error) {
		if next < len(p.steps) {
			c.StoppedBefore = p.steps[next].ID
		}
		return out, nil
	}
	if in.CheckNetwork("") != nil {
		return stop(0)
	}
	// Unknown platform validation supplies no affirmative connectivity context.
	// No NET-03 or library connectivity request is synthesized.
	if c.Validation == "unknown" || in.Composite.RequireValidation && c.Validation != "validated" {
		if !in.RecordOutcome() {
			return stop(0)
		}
		out.Outcome = "inconclusive"
		if c.Validation == "not_validated" {
			out.Outcome = "network_not_validated"
		}
		c.StoppedBefore = p.steps[0].ID
		return out, nil
	}
	for i, step := range p.steps {
		if ctx.Err() != nil || in.CheckNetwork("") != nil {
			return stop(i)
		}
		child, err := in.Composite.Step(step.Test, primitiveTargets(step.Test.ID, step.Parameters))
		if err != nil || ctx.Err() != nil {
			return stop(i)
		}
		child.Parameters = step.Parameters
		started, recorded := false, false
		check := child.CheckNetwork
		child.CheckNetwork = func(family string) error { started = true; return check(family) }
		marker := child.RecordOutcome
		child.RecordOutcome = func() bool {
			if !marker() {
				return false
			}
			if i == len(p.steps)-1 && !in.RecordOutcome() {
				return false
			}
			recorded = true
			return true
		}
		result, err := r.primitive(ctx, step.Test.ID, child)
		if err != nil || !recorded {
			if started {
				c.Steps = append(c.Steps, measurement.StepOutput{ID: step.ID, Test: step.Test})
				return stop(i + 1)
			}
			return stop(i)
		}
		c.Steps = append(c.Steps, measurement.StepOutput{ID: step.ID, Test: step.Test, Output: &result})
	}
	out.Outcome = "completed"
	return out, nil
}
func (r *Runner) acquire(ctx context.Context) bool {
	for {
		r.mu.Lock()
		if ctx.Err() != nil {
			r.mu.Unlock()
			return false
		}
		wait := r.active
		if wait == nil {
			r.active = make(chan struct{})
			r.mu.Unlock()
			return true
		}
		r.mu.Unlock()
		select {
		case <-ctx.Done():
			return false
		case <-wait:
		}
	}
}
func (r *Runner) release() { r.mu.Lock(); defer r.mu.Unlock(); close(r.active); r.active = nil }
