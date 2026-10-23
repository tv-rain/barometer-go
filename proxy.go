package barometer

import (
	"context"
	"sync"
	"time"

	"github.com/tv-rain/barometer-go/internal/config"
	"github.com/tv-rain/barometer-go/internal/execution"
)

// proxyState owns the independently cancellable proxy generations. A lease
// covers both run reservation (including platform I/O) and the launched run.
// Its mutex protects bookkeeping only; cancellation and waiting happen outside it.
type proxyState struct {
	mu         sync.Mutex
	endpoint   config.Proxy
	configured bool
	generation uint64
	group      *proxyGroup
}

type proxyGroup struct {
	ctx    context.Context
	cancel context.CancelCauseFunc
	active int
	sealed bool
	done   chan struct{}
}

func newProxyGroup(parent context.Context) *proxyGroup {
	ctx, cancel := context.WithCancelCause(parent)
	return &proxyGroup{ctx: ctx, cancel: cancel, done: make(chan struct{})}
}

func newProxyState(parent context.Context, endpoint config.Proxy, configured bool, generation uint64) *proxyState {
	return &proxyState{endpoint: endpoint, configured: configured, generation: generation, group: newProxyGroup(parent)}
}

func (p *proxyState) snapshotLocked() proxySnapshot {
	if !p.configured {
		return proxySnapshot{}
	}
	return proxySnapshot{endpoint: p.endpoint, generation: p.generation}
}

func (p *proxyState) snapshot() proxySnapshot {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.snapshotLocked()
}

// acquire refuses a generation once replacement has sealed it. A StartRun
// that reserved the client before the transition cannot slip into the old path.
func (p *proxyState) acquire() (proxySnapshot, *proxyGroup, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.group.sealed || p.group.ctx.Err() != nil {
		return proxySnapshot{}, nil, fail(transitionInProgress)
	}
	p.group.active++
	return p.snapshotLocked(), p.group, nil
}

func (p *proxyState) release(g *proxyGroup) {
	p.mu.Lock()
	defer p.mu.Unlock()
	g.active--
	if g.sealed && g.active == 0 {
		close(g.done)
	}
}

func (p *proxyState) seal() *proxyGroup {
	p.mu.Lock()
	defer p.mu.Unlock()
	g := p.group
	g.sealed = true
	if g.active == 0 {
		close(g.done)
	}
	return g
}

func (p *proxyState) install(parent context.Context, endpoint config.Proxy, configured bool) uint64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.endpoint, p.configured = endpoint, configured
	p.generation++
	p.group = newProxyGroup(parent)
	return p.generation
}

func (p *proxyState) successor(parent context.Context) *proxyState {
	p.mu.Lock()
	defer p.mu.Unlock()
	return newProxyState(parent, p.endpoint, p.configured, p.generation+1)
}

func (p *proxyState) String() string   { return "barometer.proxyState" }
func (p *proxyState) GoString() string { return "barometer.proxyState" }

// SetProxy validates a complete host-supplied proxy object before stopping
// proxy work, then cancels and drains the previous generation and atomically
// installs the replacement. Invalid input returns proxy_rejected unchanged.
// The host must keep the old service available until successful return.
func (c *Client) SetProxy(proxyJSON string) error { return c.changeProxy(proxyJSON, false) }

// ClearProxy cancels and drains the previous proxy generation before clearing
// its configuration. Direct runs are unaffected. It is safe to repeat.
func (c *Client) ClearProxy() error { return c.changeProxy("", true) }

// requirement-ref: REQ-PROXY-0008
// requirement-ref: REQ-PROXY-0011
// requirement-ref: REQ-PROXY-0012
// requirement-ref: REQ-PROXY-0013
// requirement-ref: REQ-PROXY-0015
func (c *Client) changeProxy(input string, clear bool) (err error) {
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
	rt, a := c.rt, c.bound
	c.state = transitioning
	c.mu.Unlock()
	quiesced, finished := false, false
	var nextGeneration uint64
	defer func() {
		if recover() != nil {
			err = fail(proxyRejected)
			if quiesced {
				err = fail(cleanupIncomplete)
			}
		}
		c.mu.Lock()
		defer c.mu.Unlock()
		if quiesced && !finished {
			c.state, c.blockedBy = blocked, cleanupIncomplete
		} else {
			if finished {
				c.proxyGeneration = nextGeneration
			}
			c.state = initialized
		}
	}()
	var endpoint config.Proxy
	if !clear {
		if len(input) > config.MaxBytes {
			return fail(proxyRejected)
		}
		// Reuse the strict configuration decoder, including the trusted
		// connection-code allowlist. No unvalidated input reaches runtime state.
		d, e := config.NewDecoder(config.Contract, a.approvals)
		if e != nil {
			return fail(proxyRejected)
		}
		base, e := d.Decode([]byte(`{}`))
		if e != nil {
			return fail(proxyRejected)
		}
		next, e := d.WithProxy(base, []byte(input))
		if e != nil {
			return fail(proxyRejected)
		}
		endpoint, _ = next.Proxy()
	}
	// All proxy reservations hold a lease before any path I/O. Sealing and
	// cancelling therefore covers late reservations as well as running attempts.
	g := rt.proxies.seal()
	quiesced = true
	g.cancel(execution.ErrProxyReconfigured)
	if !rt.awaitProxy(g.done) {
		return fail(cleanupIncomplete)
	}
	nextGeneration = rt.proxies.install(rt.ctx, endpoint, !clear)
	// The drained generation's egress contexts are invalid (REQ-PROXY-0016);
	// the new generation's key differs from them in any case.
	if a.run.contextCache != nil {
		a.run.contextCache.ForgetProxy()
	}
	finished = true
	return nil
}

// awaitProxy waits for the whole generation, using one cleanup budget (not
// one per run), capped also by the configured real-time lifecycle deadline.
func (rt *runtime) awaitProxy(done <-chan struct{}) bool {
	select {
	case <-done:
		return true
	default:
	}
	expired := make(chan struct{})
	stop := rt.run.clock.AfterFunc(execution.CleanupBound, func() { close(expired) })
	defer stop()
	timer := time.NewTimer(min(rt.deadline, execution.CleanupBound))
	defer timer.Stop()
	select {
	case <-done:
		return true
	case <-expired:
		return false
	case <-timer.C:
		return false
	}
}
