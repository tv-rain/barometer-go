package dns03

import (
	"context"
	"crypto/x509"
	"io"
	"net/netip"
	"net/url"
	"strconv"
	"sync"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	"github.com/tv-rain/barometer-go/internal/protocol/dns"
	httpprotocol "github.com/tv-rain/barometer-go/internal/protocol/http"
	"github.com/tv-rain/barometer-go/internal/protocol/tls"
)

// MaxBody reads one byte past the 65,535-byte DNS message bound, so a
// complete message at the bound is distinguished from an oversized one.
const MaxBody = dns.MaxTCP + 1

// Runner is instance-owned by the production registry. Roots selects the same
// reviewed trust policy as TLS-01; nil uses system roots, and only tests
// supply a synthetic CA. Equivalent concurrent resolver and query tuples wait
// under their own attempt contexts, without contact.
type Runner struct {
	Roots *x509.CertPool

	mu     sync.Mutex
	active map[tuple]chan struct{}
}

// tuple is the resolver and query key; the address family option does not
// distinguish it.
type tuple struct {
	url, name, query string
	bootstrap        netip.Addr
}

// Every stage reports its registered timeout as timeout and an unclassified
// failure as inconclusive; any other failure keeps its stage class.
var stageOutcomes = map[string]string{"dns_query": "bootstrap_failure", "tcp_connect": "tcp_failure", "tls_handshake": "tls_failure", "http_exchange": "http_failure"}

var dnsOutcomes = map[string]string{"resolved": "response", "no_data": "response", "nxdomain": "dns_nxdomain", "servfail": "dns_servfail", "refused": "dns_refused",
	"malformed_response": "malformed_dns_response", "query_mismatch": "query_mismatch", "inconclusive": "inconclusive"}

// Run resolves the resolver hostname once through the system resolver unless
// a bootstrap address was supplied, opens one TCP connection, authenticates
// TLS for the URL hostname, sends one RFC 8484 POST through the shared HTTP
// exchange with redirects disabled and parses the bounded DNS payload. A
// non-2xx status is http_failure before the media type is checked and
// without a DNS parse (GO-SPEC-ISSUE-020). Handshake and record bytes are not
// counted; HTTP framing and the DNS messages are counted on the session's
// plaintext side (REQ-LIMIT-0034). There is no GET, redirect, alternate
// endpoint or address, retry or connection reuse.
// requirement-ref: REQ-DNS-03-0005
// requirement-ref: REQ-DNS-03-0006
// requirement-ref: REQ-DNS-03-0007
// requirement-ref: REQ-DNS-03-0008
// requirement-ref: REQ-DNS-03-0011
// requirement-ref: REQ-DNS-03-0012
// requirement-ref: REQ-DNS-03-0023
// requirement-ref: REQ-DNS-03-0025
// requirement-ref: REQ-DNS-03-0026
// requirement-ref: REQ-DNS-03-0027
// requirement-ref: REQ-DNS-03-0028
// requirement-ref: REQ-DNS-03-0036
// requirement-ref: REQ-DNS-03-0037
// requirement-ref: REQ-LIMIT-0034
func (r *Runner) Run(ctx context.Context, in measurement.Input) (measurement.Output, error) {
	p, ok := in.Parameters.(Parameters)
	d, staged := in.Dialer.(measurement.StagedDialer)
	if !ok || !staged || !p.valid() || r == nil || in.Clock == nil || in.CheckNetwork == nil || in.RecordOutcome == nil || ctx.Err() != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	key := tuple{p.url, p.name, p.query, p.bootstrap}
	if !r.acquire(ctx, key) {
		return measurement.Output{}, measurement.ErrRunner
	}
	defer r.release(key)
	// RFC 8484 recommends DNS ID 0 for DoH; TLS protects the exchange, and a
	// fixed ID keeps the message ordinary (REQ-DNS-03-0036).
	q := dns.Question{ID: 0, Name: p.name, Type: p.query}
	query, err := dns.Query(q)
	if err != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	e := &endpoint{p: p, in: in, dial: d, roots: r.Roots}
	res, cleanup, err := httpprotocol.Exchange(ctx, in.Clock, e, httpprotocol.Config{URL: p.url, PostBody: query, MaxBodyBytes: MaxBody})
	// The terminal outcome is fixed before cleanup closes the session
	// (GO-SPEC-ISSUE-004).
	defer cleanup()
	if err != nil || len(res.Observations) == 0 {
		return measurement.Output{}, measurement.ErrRunner
	}
	obs := res.Observations
	last := obs[len(obs)-1]
	outcome, ok := stageOutcomes[last.Type()]
	if !ok {
		return measurement.Output{}, measurement.ErrRunner
	}
	if last.Status() == "failed" {
		if f := last.Failure(); f == "timeout" || f == "inconclusive" || f == "unexpected_content_type" {
			outcome = f
		}
	} else if h, _ := last.HTTP(); last.Type() == "http_exchange" && h.StatusCode != nil && *h.StatusCode >= 200 && *h.StatusCode <= 299 {
		truncated := h.Body == nil || h.Body.Truncated
		message, err := dns.Message(res.TakeBody(), truncated, q, h.AddressFamily, last.DurationMS())
		if outcome, ok = dnsOutcomes[message.Outcome]; err != nil || !ok {
			return measurement.Output{}, measurement.ErrRunner
		}
		obs = append(obs, message.Observation)
	} else if last.Type() != "http_exchange" {
		return measurement.Output{}, measurement.ErrRunner
	}
	usage := measurement.Usage{}
	for _, o := range obs {
		if o.Type() == "tcp_connect" {
			usage.Connections = 1
		}
	}
	if !in.RecordOutcome() {
		return measurement.Output{}, measurement.ErrRunner
	}
	return measurement.Output{Outcome: outcome, Observations: obs, Usage: usage}, nil
}

// endpoint owns the DNS-03 transport stages for the one resolver request.
type endpoint struct {
	p     Parameters
	in    measurement.Input
	dial  measurement.StagedDialer
	roots *x509.CertPool
}

// plaintext is the metered application side of the session. Closing it
// closes the attempt connection.
type plaintext struct {
	io.ReadWriter
	close func() error
}

func (s *plaintext) Close() error { return s.close() }

// Open connects to the bootstrap address or, without one, to the resolver
// hostname through the staged destination gate (auto makes one IPv4 lookup,
// as HTTP-01 does), then authenticates TLS for the URL hostname even when a
// bootstrap address was used. It offers ALPN http/1.1 only: the shared
// exchange is HTTP/1.1.
func (e *endpoint) Open(ctx context.Context, _ *url.URL) (httpprotocol.Stream, []observation.Observation, string, error) {
	port := strconv.Itoa(int(e.p.port))
	network, family, address := "tcp4", "ipv4", e.p.host+":"+port
	if e.p.bootstrap.IsValid() {
		network, address = "tcp", netip.AddrPortFrom(e.p.bootstrap, e.p.port).String()
		if e.p.bootstrap.Is6() {
			family = "ipv6"
		}
	} else if e.p.family == "ipv6" {
		network, family = "tcp6", "ipv6"
	}
	if e.in.CheckNetwork(family) != nil || ctx.Err() != nil {
		return httpprotocol.Stream{}, nil, "", measurement.ErrRunner
	}
	c, obs, family, err := e.dial.ConnectStages(ctx, network, address, e.in.Clock)
	if err != nil {
		return httpprotocol.Stream{}, nil, "", err
	}
	if c == nil {
		if len(obs) == 0 {
			return httpprotocol.Stream{}, nil, "", measurement.ErrRunner
		}
		return httpprotocol.Stream{}, obs, stageOutcomes[obs[len(obs)-1].Type()], nil
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
	session, handshake, err := tls.Handshake(ctx, e.in.Clock, family, raw, tls.Config{ServerName: e.p.host, ALPN: []string{"http/1.1"}, Roots: e.roots})
	if err != nil {
		_ = c.Close()
		return httpprotocol.Stream{}, nil, "", err
	}
	obs = append(obs, handshake)
	if session == nil {
		_ = c.Close()
		return httpprotocol.Stream{}, obs, "tls_failure", nil
	}
	return httpprotocol.Stream{Conn: &plaintext{ReadWriter: meter(session), close: c.Close}, Family: family}, obs, "", nil
}

// Redirect is never reached: a DNS POST does not classify a Location, and
// redirect targets are never contacted (REQ-DNS-03-0012).
func (e *endpoint) Redirect(*url.URL) (httpprotocol.Endpoint, string, error) {
	return nil, "unregistered_https", measurement.ErrTargetRejected
}

func (r *Runner) acquire(ctx context.Context, key tuple) bool {
	for {
		r.mu.Lock()
		if r.active == nil {
			r.active = make(map[tuple]chan struct{})
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
func (r *Runner) release(key tuple) {
	r.mu.Lock()
	defer r.mu.Unlock()
	close(r.active[key])
	delete(r.active, key)
}
