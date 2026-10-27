package dns04

import (
	"context"
	"crypto/rand"
	"crypto/x509"
	"encoding/binary"
	"encoding/json"
	"io"
	"net/netip"
	"strconv"
	"sync"
	"time"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	"github.com/tv-rain/barometer-go/internal/protocol/dns"
	"github.com/tv-rain/barometer-go/internal/protocol/tls"
)

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
	resolver, name, query string
	port                  uint16
	bootstrap             netip.Addr
}

var tcpOutcomes = map[string]string{"timeout": "tcp_timeout", "connection_refused": "tcp_refused", "connection_reset": "tcp_reset", "inconclusive": "inconclusive"}

// A reset during TLS keeps the TCP class; every other TLS failure, including
// a resolver certificate failure, is tls_failure.
var tlsOutcomes = map[string]string{"timeout": "timeout", "connection_reset": "tcp_reset", "inconclusive": "inconclusive"}

var dnsOutcomes = map[string]string{"resolved": "response", "no_data": "response", "nxdomain": "dns_nxdomain", "servfail": "dns_servfail", "refused": "dns_refused",
	"malformed_response": "malformed_dns_response", "query_mismatch": "query_mismatch", "timeout": "timeout", "inconclusive": "inconclusive"}

// Run resolves the resolver name once through the system resolver unless a
// bootstrap address was supplied, opens one TCP connection, authenticates
// TLS for the resolver name, exchanges one length-prefixed DNS message and
// closes the session. Handshake and record bytes are not counted; the DNS
// message and its two-byte length prefix are counted on the session's
// plaintext side (REQ-LIMIT-0034). There is no cleartext fallback, alternate
// resolver or application-layer retry.
// requirement-ref: REQ-DNS-04-0005
// requirement-ref: REQ-DNS-04-0006
// requirement-ref: REQ-DNS-04-0007
// requirement-ref: REQ-DNS-04-0008
// requirement-ref: REQ-DNS-04-0011
// requirement-ref: REQ-DNS-04-0012
// requirement-ref: REQ-DNS-04-0023
// requirement-ref: REQ-DNS-04-0025
// requirement-ref: REQ-DNS-04-0026
// requirement-ref: REQ-DNS-04-0027
// requirement-ref: REQ-DNS-04-0028
// requirement-ref: REQ-DNS-04-0036
// requirement-ref: REQ-LIMIT-0034
func (r *Runner) Run(ctx context.Context, in measurement.Input) (measurement.Output, error) {
	p, ok := in.Parameters.(Parameters)
	d, staged := in.Dialer.(measurement.StagedDialer)
	if !ok || !staged || !p.valid() || r == nil || in.Clock == nil || in.CheckNetwork == nil || in.RecordOutcome == nil || ctx.Err() != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	key := tuple{p.resolver, p.name, p.query, p.port, p.bootstrap}
	if !r.acquire(ctx, key) {
		return measurement.Output{}, measurement.ErrRunner
	}
	defer r.release(key)
	port := strconv.Itoa(int(p.port))
	// Without a bootstrap address, auto makes one IPv4 lookup, as HTTP-01
	// does: one explicit family, no hidden second lookup.
	network, family, address := "tcp4", "ipv4", p.resolver+":"+port
	if p.bootstrap.IsValid() {
		network, address = "tcp", netip.AddrPortFrom(p.bootstrap, p.port).String()
		if p.bootstrap.Is6() {
			family = "ipv6"
		}
	} else if p.family == "ipv6" {
		network, family = "tcp6", "ipv6"
	}
	if in.CheckNetwork(family) != nil || ctx.Err() != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	c, obs, family, err := d.ConnectStages(ctx, network, address, in.Clock)
	if err != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	if c == nil {
		if len(obs) == 0 {
			return measurement.Output{}, measurement.ErrRunner
		}
		last := obs[len(obs)-1]
		outcome, usage := "bootstrap_failure", measurement.Usage{}
		if last.Type() == "tcp_connect" {
			usage.Connections = 1
			if outcome, ok = tcpOutcomes[last.Failure()]; !ok {
				if obs[len(obs)-1], err = inconclusiveConnect(last, family); err != nil {
					return measurement.Output{}, measurement.ErrRunner
				}
				outcome = "inconclusive"
			}
		}
		if !in.RecordOutcome() {
			return measurement.Output{}, measurement.ErrRunner
		}
		return measurement.Output{Outcome: outcome, Observations: obs, Usage: usage}, nil
	}
	sc, ok := c.(measurement.SessionConn)
	if !ok {
		_ = c.Close()
		return measurement.Output{}, measurement.ErrRunner
	}
	raw, meter, err := sc.Session()
	if err != nil {
		_ = c.Close()
		return measurement.Output{}, measurement.ErrRunner
	}
	// TLS authenticates the resolver name even when a bootstrap address was
	// used. No ALPN is offered: the test registers none.
	session, handshake, err := tls.Handshake(ctx, in.Clock, family, raw, tls.Config{ServerName: p.resolver, Roots: r.Roots})
	if err != nil {
		_ = c.Close()
		return measurement.Output{}, measurement.ErrRunner
	}
	obs = append(obs, handshake)
	usage := measurement.Usage{Connections: 1}
	if session == nil {
		outcome, ok := tlsOutcomes[handshake.Failure()]
		if !ok {
			outcome = "tls_failure"
		}
		if !in.RecordOutcome() {
			return measurement.Output{}, measurement.ErrRunner
		}
		return measurement.Output{Outcome: outcome, Observations: obs, Usage: usage}, nil
	}
	stream := &plaintext{ReadWriter: meter(session), raw: raw, close: c.Close}
	var id [2]byte
	if _, err = rand.Read(id[:]); err != nil {
		_ = c.Close()
		return measurement.Output{}, measurement.ErrRunner
	}
	res, err := dns.Session(ctx, in.Clock, stream, dns.Question{ID: binary.BigEndian.Uint16(id[:]), Name: p.name, Type: p.query}, family)
	if err != nil {
		_ = c.Close()
		return measurement.Output{}, measurement.ErrRunner
	}
	outcome, ok := dnsOutcomes[res.Outcome]
	query := res.Observation
	if !ok {
		// A DNS-stage failure without a DNS-04 class, such as a closed
		// session, keeps its stage and duration as inconclusive.
		if query, err = inconclusiveQuery(query); err != nil {
			_ = c.Close()
			return measurement.Output{}, measurement.ErrRunner
		}
		outcome = "inconclusive"
	}
	// The terminal outcome is fixed before close (GO-SPEC-ISSUE-004).
	recorded := in.RecordOutcome()
	_ = c.Close()
	if !recorded {
		return measurement.Output{}, measurement.ErrRunner
	}
	return measurement.Output{Outcome: outcome, Observations: append(obs, query), Usage: usage}, nil
}

// plaintext is the metered application side of the session. Closing it
// closes the attempt connection; deadlines apply to the raw stream.
type plaintext struct {
	io.ReadWriter
	raw   measurement.Conn
	close func() error
}

func (s *plaintext) Close() error                  { return s.close() }
func (s *plaintext) SetDeadline(t time.Time) error { return s.raw.SetDeadline(t) }

// inconclusiveConnect keeps the observed stage and duration of a TCP failure
// that DNS-04 has no class for, without its unsupported failure value.
func inconclusiveConnect(o observation.Observation, family string) (observation.Observation, error) {
	type details struct {
		Family string `json:"address_family"`
	}
	data, err := json.Marshal(struct {
		Type     string  `json:"type"`
		Status   string  `json:"status"`
		Duration int     `json:"duration_ms"`
		Failure  string  `json:"failure"`
		Details  details `json:"details"`
	}{"tcp_connect", "failed", o.DurationMS(), "inconclusive", details{family}})
	if err != nil {
		return observation.Observation{}, err
	}
	return observation.DecodeProtocol(observation.Contract, data, observation.ProtocolLimits{})
}

func inconclusiveQuery(o observation.Observation) (observation.Observation, error) {
	d, ok := o.DNS()
	if !ok {
		return observation.Observation{}, measurement.ErrRunner
	}
	return observation.NewDNS(observation.Contract, "failed", o.DurationMS(), "inconclusive", d, observation.ProtocolLimits{RetainDNSAnswers: true})
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
