package tls01

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"io"
	"net/netip"
	"sync"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	"github.com/tv-rain/barometer-go/internal/protocol/tcp"
	"github.com/tv-rain/barometer-go/internal/protocol/tls"
)

// Runner is instance-owned by the production registry. Roots are the trust
// anchors of the documented trust policy; nil selects the system roots, and
// only tests supply a synthetic CA. Equivalent concurrent endpoint and
// server-name tuples wait under their own attempt contexts, without contact.
type Runner struct {
	Roots *x509.CertPool

	mu     sync.Mutex
	active map[tuple]chan struct{}
}

type tuple struct {
	endpoint netip.AddrPort
	name     string
}

// TCP failures outside the TLS-01 table have no more specific TLS-01 class.
var tcpOutcomes = map[string]string{"timeout": "tcp_timeout", "connection_refused": "tcp_refused", "connection_reset": "tcp_reset", "inconclusive": "inconclusive"}

var tlsOutcomes = map[string]string{"timeout": "tls_timeout", "alert": "tls_alert", "connection_reset": "tcp_reset", "protocol_error": "protocol_error",
	"certificate_untrusted": "certificate_untrusted", "certificate_name_mismatch": "certificate_name_mismatch", "certificate_time_invalid": "certificate_time_invalid",
	"alpn_no_overlap": "alpn_no_overlap", "inconclusive": "inconclusive"}

// Run opens one connection to the explicit address, performs one handshake
// and closes the session without application data. The TLS stream is
// switched to plaintext metering before the handshake, so handshake and
// record bytes are not counted (REQ-LIMIT-0034).
// requirement-ref: REQ-TLS-01-0005
// requirement-ref: REQ-TLS-01-0006
// requirement-ref: REQ-TLS-01-0008
// requirement-ref: REQ-TLS-01-0018
// requirement-ref: REQ-TLS-01-0023
// requirement-ref: REQ-TLS-01-0025
// requirement-ref: REQ-TLS-01-0026
// requirement-ref: REQ-TLS-01-0027
// requirement-ref: REQ-TLS-01-0035
// requirement-ref: REQ-LIMIT-0034
func (r *Runner) Run(ctx context.Context, in measurement.Input) (measurement.Output, error) {
	p, ok := in.Parameters.(Parameters)
	if !ok || r == nil || in.Dialer == nil || in.Clock == nil || in.CheckNetwork == nil || in.RecordOutcome == nil || ctx.Err() != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	a, port, name, ok := p.Target()
	if !ok {
		return measurement.Output{}, measurement.ErrRunner
	}
	key := tuple{netip.AddrPortFrom(a, port), name}
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
	c, connect, err := tcp.Connect(ctx, in.Clock, family, func(op context.Context) (io.ReadWriteCloser, error) {
		return in.Dialer.DialContext(op, "tcp", key.endpoint.String())
	})
	if err != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	if c == nil {
		outcome, ok := tcpOutcomes[connect.Failure()]
		if !ok {
			if connect, err = inconclusiveConnect(connect, family); err != nil {
				return measurement.Output{}, measurement.ErrRunner
			}
			outcome = "inconclusive"
		}
		if !in.RecordOutcome() {
			return measurement.Output{}, measurement.ErrRunner
		}
		return measurement.Output{Outcome: outcome, Observations: []observation.Observation{connect}, Usage: measurement.Usage{Connections: 1}}, nil
	}
	sc, ok := c.(measurement.SessionConn)
	if !ok {
		_ = c.Close()
		return measurement.Output{}, measurement.ErrRunner
	}
	raw, _, err := sc.Session()
	if err != nil {
		_ = c.Close()
		return measurement.Output{}, measurement.ErrRunner
	}
	session, handshake, err := tls.Handshake(ctx, in.Clock, family, raw, tls.Config{ServerName: name, ALPN: p.ALPN(), Roots: r.Roots})
	if err != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	outcome := tlsOutcomes[handshake.Failure()]
	if handshake.Status() == "succeeded" {
		outcome = "handshake_succeeded"
	}
	// The terminal outcome is fixed before close (GO-SPEC-ISSUE-004).
	recorded := in.RecordOutcome()
	if session != nil {
		_ = session.Close()
	}
	if !recorded || outcome == "" {
		return measurement.Output{}, measurement.ErrRunner
	}
	return measurement.Output{Outcome: outcome, Observations: []observation.Observation{connect, handshake}, Usage: measurement.Usage{Connections: 1}}, nil
}

// inconclusiveConnect keeps the observed stage and duration of a TCP failure
// that TLS-01 has no class for, without its unsupported failure value.
func inconclusiveConnect(o observation.Observation, family string) (observation.Observation, error) {
	data, err := json.Marshal(struct {
		Type     string `json:"type"`
		Status   string `json:"status"`
		Duration int    `json:"duration_ms"`
		Failure  string `json:"failure"`
		Details  struct {
			Family string `json:"address_family"`
		} `json:"details"`
	}{"tcp_connect", "failed", o.DurationMS(), "inconclusive", struct {
		Family string `json:"address_family"`
	}{family}})
	if err != nil {
		return observation.Observation{}, err
	}
	return observation.DecodeProtocol(observation.Contract, data, observation.ProtocolLimits{})
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
