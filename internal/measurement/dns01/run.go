package dns01

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"net/netip"
	"sync"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	"github.com/tv-rain/barometer-go/internal/protocol/dns"
)

// Runner serializes equivalent active tuples in its owning client generation.
// Including different network generations in that serialization is conservative:
// no equivalent operation overlaps, and waiters honor their own deadlines.
type Runner struct {
	mu     sync.Mutex
	active map[Parameters]chan struct{}
}

// requirement-ref: REQ-DNS-01-0009
// requirement-ref: REQ-DNS-01-0010
// requirement-ref: REQ-DNS-01-0039
func (r *Runner) Run(ctx context.Context, in measurement.Input) (measurement.Output, error) {
	p, ok := in.Parameters.(Parameters)
	if !ok || r == nil || p.name == "" || in.Clock == nil || in.CheckNetwork == nil || in.RecordOutcome == nil || ctx.Err() != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	// Options do not distinguish the concurrent hostname/type/mode/resolver key.
	key := p
	key.collect = false
	key.window = 0
	key.transport = ""
	if !r.acquire(ctx, key) {
		return measurement.Output{}, measurement.ErrRunner
	}
	defer r.release(key)
	family := ""
	if p.mode == "direct" {
		family = "ipv6"
		if p.address.Is4() {
			family = "ipv4"
		}
	}
	if in.CheckNetwork(family) != nil || ctx.Err() != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	operation := dns.Input{Clock: in.Clock, Resolve: in.Resolve, Dial: func(ctx context.Context, network, address string) (dns.Conn, error) {
		if in.Dialer == nil {
			return nil, measurement.ErrRunner
		}
		return in.Dialer.DialContext(ctx, network, address)
	}}
	var result dns.Result
	var err error
	var c measurement.Conn
	if p.mode == "system" {
		result, err = dns.System(ctx, operation, p.name, p.query)
	} else {
		var id [2]byte
		if _, err = rand.Read(id[:]); err != nil {
			return measurement.Output{}, measurement.ErrRunner
		}
		c, result, err = dns.Direct(ctx, operation, dns.Question{ID: binary.BigEndian.Uint16(id[:]), Name: p.name, Type: p.query}, p.transport, netip.AddrPortFrom(p.address, p.port).String(), family, p.collect, p.window)
	}
	if c != nil {
		defer c.Close()
	}
	if err != nil || !in.RecordOutcome() {
		return measurement.Output{}, measurement.ErrRunner
	}
	return measurement.Output{Outcome: result.Outcome, Observations: []observation.Observation{result.Observation}}, nil
}
func (r *Runner) acquire(ctx context.Context, key Parameters) bool {
	for {
		r.mu.Lock()
		if r.active == nil {
			r.active = make(map[Parameters]chan struct{})
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
func (r *Runner) release(key Parameters) {
	r.mu.Lock()
	defer r.mu.Unlock()
	close(r.active[key])
	delete(r.active, key)
}
