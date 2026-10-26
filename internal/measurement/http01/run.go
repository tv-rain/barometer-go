package http01

import (
	"context"
	"crypto/x509"
	"errors"
	"io"
	"net/netip"
	"net/url"
	"sync"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	httpprotocol "github.com/tv-rain/barometer-go/internal/protocol/http"
	"github.com/tv-rain/barometer-go/internal/protocol/tls"
)

// Runner is per-client and serializes equal normalized URLs. Roots selects
// the same reviewed trust policy as TLS-01; nil uses system roots.
type Runner struct {
	Roots  *x509.CertPool
	mu     sync.Mutex
	active map[string]chan struct{}
}

// requirement-ref: REQ-HTTP-01-0005
// requirement-ref: REQ-HTTP-01-0007
// requirement-ref: REQ-HTTP-01-0032
// requirement-ref: REQ-HTTP-01-0038
// requirement-ref: REQ-HTTP-01-0052
func (r *Runner) Run(ctx context.Context, in measurement.Input) (measurement.Output, error) {
	p, ok := in.Parameters.(Parameters)
	d, has := in.Dialer.(measurement.StagedDialer)
	if !ok || !has || r == nil || in.Clock == nil || in.RecordOutcome == nil || in.CheckNetwork == nil || p.tuple.URL == "" || ctx.Err() != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	if !r.acquire(ctx, p.tuple.URL) {
		return measurement.Output{}, measurement.ErrRunner
	}
	defer r.release(p.tuple.URL)
	if in.CheckNetwork("") != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	cfg := httpprotocol.Config{URL: p.tuple.URL, MaxBodyBytes: p.body, MaxRedirects: p.redirects, ExpectedStatus: p.tuple.ExpectedStatus}
	for _, c := range p.tuple.BodyChecks {
		cfg.BodyChecks = append(cfg.BodyChecks, httpprotocol.BodyCheck{Type: c.Type, Value: c.Value})
	}
	for _, c := range p.tuple.HeaderChecks {
		cfg.HeaderChecks = append(cfg.HeaderChecks, httpprotocol.HeaderCheck{Name: c.Name, Type: c.Type, Value: c.Value})
	}
	res, cleanup, err := httpprotocol.Exchange(ctx, in.Clock, endpoint{p: p, in: in, dial: d, roots: r.Roots}, cfg)
	defer cleanup()
	if err != nil || !in.RecordOutcome() {
		return measurement.Output{}, measurement.ErrRunner
	}
	outcome := res.Outcome
	if outcome == "timeout" {
		outcome = "http_timeout"
	}
	return measurement.Output{Outcome: outcome, Observations: res.Observations}, nil
}

type endpoint struct {
	p        Parameters
	in       measurement.Input
	dial     measurement.StagedDialer
	roots    *x509.CertPool
	redirect bool
}
type plaintext struct {
	io.ReadWriter
	close func() error
}

func (c *plaintext) Close() error { return c.close() }
func (e endpoint) Open(ctx context.Context, u *url.URL) (httpprotocol.Stream, []observation.Observation, string, error) {
	network, family := "tcp4", "ipv4"
	if ip, err := netip.ParseAddr(u.Hostname()); err == nil {
		network = "tcp"
		if ip.Is6() {
			family = "ipv6"
		}
	} else if e.p.family == "ipv6" {
		network, family = "tcp6", "ipv6"
	}
	if e.in.CheckNetwork(family) != nil {
		return httpprotocol.Stream{}, nil, "", measurement.ErrRunner
	}
	port := u.Port()
	if port == "" {
		port = "443"
	}
	authority := u.Host
	if u.Port() == "" {
		authority += ":" + port
	}
	c, obs, family, err := e.dial.ConnectStages(ctx, network, authority, e.in.Clock)
	if errors.Is(err, measurement.ErrTargetRejected) && e.redirect {
		return httpprotocol.Stream{}, nil, "", &httpprotocol.DestinationRejected{Class: "prohibited_host"}
	}
	if err != nil {
		return httpprotocol.Stream{}, nil, "", err
	}
	if c == nil {
		if len(obs) == 0 {
			return httpprotocol.Stream{}, nil, "", measurement.ErrRunner
		}
		outcome := "tcp_failure"
		if obs[len(obs)-1].Type() == "dns_query" {
			outcome = "dns_failure"
		}
		return httpprotocol.Stream{}, obs, outcome, nil
	}
	sc, ok := c.(measurement.SessionConn)
	if !ok {
		_ = c.Close()
		return httpprotocol.Stream{}, nil, "", measurement.ErrRunner
	}
	raw, meter, err := sc.Session()
	if err != nil {
		_ = c.Close()
		return httpprotocol.Stream{}, nil, "", err
	}
	session, o, err := tls.Handshake(ctx, e.in.Clock, family, raw, tls.Config{ServerName: u.Hostname(), ALPN: []string{"http/1.1"}, Roots: e.roots})
	if err != nil {
		_ = c.Close()
		return httpprotocol.Stream{}, nil, "", err
	}
	obs = append(obs, o)
	if session == nil {
		return httpprotocol.Stream{}, obs, "tls_failure", nil
	}
	return httpprotocol.Stream{Conn: &plaintext{ReadWriter: meter(session), close: c.Close}, Family: family}, obs, "", nil
}

// requirement-ref: REQ-HTTP-01-0018
// requirement-ref: REQ-HTTP-01-0019
// requirement-ref: REQ-HTTP-01-0051
func (e endpoint) Redirect(u *url.URL) (httpprotocol.Endpoint, string, error) {
	t, ok := NormalizeURL(u.String())
	if !ok {
		return nil, "prohibited_host", measurement.ErrTargetRejected
	}
	t.ExpectedStatus = e.p.tuple.ExpectedStatus
	t.BodyChecks = e.p.Target().BodyChecks
	t.HeaderChecks = e.p.Target().HeaderChecks
	parent, ok := e.dial.(measurement.TargetDialer)
	if !ok {
		return nil, "unregistered_https", measurement.ErrTargetRejected
	}
	d, err := parent.Redirect(t)
	if err != nil {
		return nil, "unregistered_https", err
	}
	staged, ok := d.(measurement.StagedDialer)
	if !ok {
		return nil, "unregistered_https", measurement.ErrTargetRejected
	}
	e.p.tuple = t
	e.dial = staged
	e.redirect = true
	return e, "registered_https", nil
}
func (r *Runner) acquire(ctx context.Context, key string) bool {
	for {
		r.mu.Lock()
		if r.active == nil {
			r.active = map[string]chan struct{}{}
		}
		w, ok := r.active[key]
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
		case <-w:
		}
	}
}
func (r *Runner) release(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	close(r.active[key])
	delete(r.active, key)
}
