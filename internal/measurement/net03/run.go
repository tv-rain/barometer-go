package net03

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
	"github.com/tv-rain/barometer-go/internal/model/task"
	httpprotocol "github.com/tv-rain/barometer-go/internal/protocol/http"
	"github.com/tv-rain/barometer-go/internal/protocol/tls"
)

// Runner is per-client and admits one active validation run at a time, so no
// two runs overlap in any network generation; a later run waits for the
// earlier one or for its own context (REQ-NET-03-0021). Roots selects the
// same reviewed trust policy as TLS-01; nil uses system roots.
type Runner struct {
	Roots *x509.CertPool
	once  sync.Once
	slot  chan struct{}
}

// Run follows the NET-03 Procedure: the NET-01 stability check, which starts
// the test, one platform read, one request without credentials, cookies or
// redirect following, the bounded comparison, and close after the outcome is
// recorded. A common interruption at any point before the outcome is the
// execution layer's and returns ErrRunner, never an outcome.
// requirement-ref: REQ-NET-03-0005
// requirement-ref: REQ-NET-03-0006
// requirement-ref: REQ-NET-03-0012
// requirement-ref: REQ-NET-03-0021
// requirement-ref: REQ-NET-03-0028
// requirement-ref: REQ-NET-03-0029
// requirement-ref: REQ-LIMIT-0018
func (r *Runner) Run(ctx context.Context, in measurement.Input) (measurement.Output, error) {
	p, ok := in.Parameters.(Parameters)
	d, has := in.Dialer.(measurement.StagedDialer)
	if !ok || !has || r == nil || in.Clock == nil || in.RecordOutcome == nil || in.CheckNetwork == nil || in.Validation == nil || p.tuple.URL == "" || ctx.Err() != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	r.once.Do(func() { r.slot = make(chan struct{}, 1) })
	select {
	case r.slot <- struct{}{}:
	case <-ctx.Done():
		return measurement.Output{}, measurement.ErrRunner
	}
	defer func() { <-r.slot }()
	if in.CheckNetwork("") != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	platform, err := in.Validation(ctx)
	if err != nil || ctx.Err() != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	cfg := httpprotocol.Config{URL: p.tuple.URL, MaxBodyBytes: PrefixBytes, ExpectedStatus: p.tuple.ExpectedStatus}
	for _, c := range p.tuple.BodyChecks {
		cfg.BodyChecks = append(cfg.BodyChecks, httpprotocol.BodyCheck{Type: c.Type, Value: c.Value})
	}
	res, cleanup, err := httpprotocol.Exchange(ctx, in.Clock, endpoint{p: p, in: in, dial: d, roots: r.Roots}, cfg)
	defer cleanup()
	if err != nil || len(res.Observations) == 0 {
		return measurement.Output{}, measurement.ErrRunner
	}
	class := Classify(res.Outcome, res.Observations[len(res.Observations)-1])
	outcome := Outcome(class, platform, res.Outcome)
	// The outcome is fixed before the connection closes (GO-SPEC-ISSUE-004).
	if !in.RecordOutcome() {
		return measurement.Output{}, measurement.ErrRunner
	}
	return measurement.Output{Outcome: outcome, Observations: res.Observations, PlatformValidation: platform, ResponseClass: class}, nil
}

// Classify returns the response_class of the terminal exchange, or "" when no
// valid HTTP response was received: a header block over the bound or a
// request that ended without a final status (REQ-NET-03-0041). A status equal
// to expected_status with a decidable different digest is content_mismatch;
// every other non-matching response, including a body that reaches the prefix
// before the digest is decidable, is unexpected.
// requirement-ref: REQ-NET-03-0025
// requirement-ref: REQ-NET-03-0026
func Classify(primitive string, last observation.Observation) string {
	h, ok := last.HTTP()
	if !ok || h.StatusCode == nil {
		return ""
	}
	switch primitive {
	case "redirect_observed":
		return "redirect"
	case "expected_response":
		return "expected"
	case "body_check_failed":
		return "content_mismatch"
	}
	return "unexpected"
}

// Outcome selects the NET-03 outcome from the active request first; the
// platform state selects it only when the request produced no response class
// (REQ-NET-03-0042). Neither source rewrites the other.
// requirement-ref: REQ-NET-03-0008
// requirement-ref: REQ-NET-03-0023
// requirement-ref: REQ-NET-03-0027
func Outcome(class, platform, primitive string) string {
	switch class {
	case "expected":
		return "validated"
	case "redirect", "content_mismatch":
		return "captive_portal_suspected"
	case "unexpected":
		return "unexpected_response"
	}
	switch platform {
	case "captive_portal":
		return "captive_portal_suspected"
	case "not_validated":
		return "not_validated"
	}
	switch primitive {
	case "dns_failure", "tls_failure":
		return primitive
	case "tcp_failure":
		return "connection_failure"
	}
	return "inconclusive"
}

type endpoint struct {
	p     Parameters
	in    measurement.Input
	dial  measurement.StagedDialer
	roots *x509.CertPool
}
type plaintext struct {
	io.ReadWriter
	close func() error
}

func (c *plaintext) Close() error { return c.close() }

// Open performs the DNS and TCP stages through the attempt's gate, then TLS
// only for an https URL. A plain http stream is the metered attempt
// connection itself.
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
		if u.Scheme == "http" {
			port = "80"
		}
	}
	authority := u.Host
	if u.Port() == "" {
		authority += ":" + port
	}
	c, obs, family, err := e.dial.ConnectStages(ctx, network, authority, e.in.Clock)
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
	if u.Scheme == "http" {
		return httpprotocol.Stream{Conn: c, Family: family}, obs, "", nil
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

// Redirect only classifies a safely parsed location; NET-03 registers no
// redirect target and never follows one, so it never authorizes a hop and
// performs no I/O. The location itself is not retained (REQ-NET-03-0012).
func (e endpoint) Redirect(u *url.URL) (httpprotocol.Endpoint, string, error) {
	if ip, err := netip.ParseAddr(u.Hostname()); err == nil && !task.PublicAddress(ip) {
		return nil, "prohibited_host", errRedirect
	}
	return nil, "unregistered_https", errRedirect
}

var errRedirect = errors.New("net03: redirect not followed")
