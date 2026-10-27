package net05

import (
	"context"
	"crypto/x509"
	"errors"
	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	"github.com/tv-rain/barometer-go/internal/model/task"
	httpprotocol "github.com/tv-rain/barometer-go/internal/protocol/http"
	"github.com/tv-rain/barometer-go/internal/protocol/tls"
	"io"
	"net/netip"
	"net/url"
)

type httpParameters interface {
	Target() task.TargetTuple
	Limits() observation.ProtocolLimits
	AddressFamily() string
	RedirectURL(string) (task.TargetTuple, bool)
}

func (r *Runner) runHTTP(ctx context.Context, in measurement.Input) (measurement.Output, error) {
	p, ok := in.Parameters.(httpParameters)
	d, has := in.Dialer.(measurement.StagedDialer)
	if !ok || !has || in.CheckNetwork("") != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	tuple := p.Target()
	limits := p.Limits()
	cfg := httpprotocol.Config{URL: tuple.URL, MaxBodyBytes: limits.MaxBodyBytes, MaxRedirects: limits.MaxRedirects, ExpectedStatus: tuple.ExpectedStatus}
	for _, c := range tuple.BodyChecks {
		cfg.BodyChecks = append(cfg.BodyChecks, httpprotocol.BodyCheck{Type: c.Type, Value: c.Value})
	}
	for _, c := range tuple.HeaderChecks {
		cfg.HeaderChecks = append(cfg.HeaderChecks, httpprotocol.HeaderCheck{Name: c.Name, Type: c.Type, Value: c.Value})
	}
	result, cleanup, err := httpprotocol.Exchange(ctx, in.Clock, endpoint{p: p, tuple: tuple, in: in, dial: d, roots: r.Roots}, cfg)
	defer cleanup()
	if err != nil || !in.RecordOutcome() {
		return measurement.Output{}, measurement.ErrRunner
	}
	outcome := result.Outcome
	if outcome == "timeout" {
		outcome = "http_timeout"
	}
	return measurement.Output{Outcome: outcome, Observations: result.Observations}, nil
}

type endpoint struct {
	p        httpParameters
	tuple    task.TargetTuple
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
	} else if e.p.AddressFamily() == "ipv6" {
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
	t, ok := e.p.RedirectURL(u.String())
	if !ok {
		return nil, "prohibited_host", measurement.ErrTargetRejected
	}
	t.ExpectedStatus = e.tuple.ExpectedStatus
	t.BodyChecks = e.tuple.BodyChecks
	t.HeaderChecks = e.tuple.HeaderChecks
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
	e.tuple = t
	e.dial = staged
	e.redirect = true
	return e, "registered_https", nil
}
