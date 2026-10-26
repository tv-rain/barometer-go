package tcp01

import (
	"context"
	"io"
	"net/netip"
	"sync"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	"github.com/tv-rain/barometer-go/internal/protocol/tcp"
)

// Runner is instance-owned by the production registry. Its zero value is ready.
// Equivalent concurrent tuples wait under their own attempt contexts, without
// contact. Entries exist only while a tuple is active and are removed on close.
type Runner struct {
	mu     sync.Mutex
	active map[netip.AddrPort]chan struct{}
}

// requirement-ref: REQ-TCP-01-0007
// requirement-ref: REQ-TCP-01-0017
// requirement-ref: REQ-TCP-01-0019
// requirement-ref: REQ-TCP-01-0022
// requirement-ref: REQ-TCP-01-0024
// requirement-ref: REQ-TCP-01-0026
func (r *Runner) Run(ctx context.Context, in measurement.Input) (measurement.Output, error) {
	p, ok := in.Parameters.(Parameters)
	if !ok || r == nil || in.Dialer == nil || in.Clock == nil || in.CheckNetwork == nil || in.RecordOutcome == nil || ctx.Err() != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	a, port, ok := p.Target()
	if !ok {
		return measurement.Output{}, measurement.ErrRunner
	}
	key := netip.AddrPortFrom(a, port)
	if !r.acquire(ctx, key) {
		return measurement.Output{}, measurement.ErrRunner
	}
	defer r.release(key)
	family := "ipv6"
	if a.Is4() {
		family = "ipv4"
	}
	if in.CheckNetwork(family) != nil || ctx.Err() != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	c, o, err := tcp.Connect(ctx, in.Clock, family, func(op context.Context) (io.ReadWriteCloser, error) {
		return in.Dialer.DialContext(op, "tcp", key.String())
	})
	if err != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	outcome := o.Failure()
	if o.Status() == "succeeded" {
		outcome = "connected"
	}
	// The terminal outcome is fixed before close (GO-SPEC-ISSUE-004).
	if !in.RecordOutcome() {
		if c != nil {
			_ = c.Close()
		}
		return measurement.Output{}, measurement.ErrRunner
	}
	if c != nil {
		_ = c.Close()
	}
	return measurement.Output{Outcome: outcome, Observations: []observation.Observation{o}, Usage: measurement.Usage{Connections: 1}}, nil
}
func (r *Runner) acquire(ctx context.Context, key netip.AddrPort) bool {
	for {
		r.mu.Lock()
		if r.active == nil {
			r.active = make(map[netip.AddrPort]chan struct{})
		}
		wait, ok := r.active[key]
		if !ok && ctx.Err() == nil {
			r.active[key] = make(chan struct{})
			r.mu.Unlock()
			return true
		}
		r.mu.Unlock()
		if !ok {
			return false
		}
		select {
		case <-ctx.Done():
			return false
		case <-wait:
		}
	}
}
func (r *Runner) release(key netip.AddrPort) {
	r.mu.Lock()
	defer r.mu.Unlock()
	close(r.active[key])
	delete(r.active, key)
}
