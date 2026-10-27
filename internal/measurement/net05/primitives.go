package net05

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"io"
	"net/netip"
	"time"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/protocol/dns"
	"github.com/tv-rain/barometer-go/internal/protocol/tcp"
	tlsprotocol "github.com/tv-rain/barometer-go/internal/protocol/tls"
)

// These narrow parameter views consume the exact-version codec's immutable
// values. NET-05 imports no sibling measurement package and calls no runner.
type dnsParameters interface {
	Name() string
	QueryType() string
	Engine() string
	Resolver() (netip.Addr, uint16)
	AdditionalResponses() (bool, time.Duration)
}
type tcpParameters interface {
	Target() (netip.Addr, uint16, bool)
}
type tlsParameters interface {
	Target() (netip.Addr, uint16, string, bool)
	ALPN() []string
}

func (r *Runner) primitive(ctx context.Context, id string, in measurement.Input) (measurement.Output, error) {
	switch id {
	case "DNS-01":
		return runDNS(ctx, in)
	case "TCP-01":
		return runTCP(ctx, in)
	case "TLS-01":
		return r.runTLS(ctx, in)
	case "HTTP-01":
		return r.runHTTP(ctx, in)
	}
	return measurement.Output{}, measurement.ErrRunner
}
func family(a netip.Addr) string {
	if a.Is4() {
		return "ipv4"
	}
	return "ipv6"
}
func runDNS(ctx context.Context, in measurement.Input) (measurement.Output, error) {
	p, ok := in.Parameters.(dnsParameters)
	if !ok {
		return measurement.Output{}, measurement.ErrRunner
	}
	f := ""
	a, port := p.Resolver()
	if p.Engine() != "system" {
		f = family(a)
	}
	if in.CheckNetwork(f) != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	op := dns.Input{Clock: in.Clock, Resolve: in.Resolve, Dial: func(c context.Context, n, a string) (dns.Conn, error) { return in.Dialer.DialContext(c, n, a) }}
	var result dns.Result
	var err error
	if p.Engine() == "system" {
		result, err = dns.System(ctx, op, p.Name(), p.QueryType())
	} else {
		var id [2]byte
		if _, err = rand.Read(id[:]); err != nil {
			return measurement.Output{}, measurement.ErrRunner
		}
		collect, window := p.AdditionalResponses()
		var conn dns.Conn
		conn, result, err = dns.Direct(ctx, op, dns.Question{ID: binary.BigEndian.Uint16(id[:]), Name: p.Name(), Type: p.QueryType()}, p.Engine(), netip.AddrPortFrom(a, port).String(), f, collect, window)
		if conn != nil {
			defer conn.Close()
		}
	}
	if err != nil || !in.RecordOutcome() {
		return measurement.Output{}, measurement.ErrRunner
	}
	return measurement.Output{Outcome: result.Outcome, Observations: []observation.Observation{result.Observation}}, nil
}
func runTCP(ctx context.Context, in measurement.Input) (measurement.Output, error) {
	p, ok := in.Parameters.(tcpParameters)
	if !ok {
		return measurement.Output{}, measurement.ErrRunner
	}
	a, port, valid := p.Target()
	if !valid || in.CheckNetwork(family(a)) != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	c, o, err := tcp.Connect(ctx, in.Clock, family(a), func(ctx context.Context) (io.ReadWriteCloser, error) {
		return in.Dialer.DialContext(ctx, "tcp", netip.AddrPortFrom(a, port).String())
	})
	if c != nil {
		defer c.Close()
	}
	if err != nil || !in.RecordOutcome() {
		return measurement.Output{}, measurement.ErrRunner
	}
	outcome := o.Failure()
	if o.Status() == "succeeded" {
		outcome = "connected"
	}
	return measurement.Output{Outcome: outcome, Observations: []observation.Observation{o}}, nil
}
func (r *Runner) runTLS(ctx context.Context, in measurement.Input) (measurement.Output, error) {
	p, ok := in.Parameters.(tlsParameters)
	if !ok {
		return measurement.Output{}, measurement.ErrRunner
	}
	a, port, name, valid := p.Target()
	if !valid || in.CheckNetwork(family(a)) != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	c, o, err := tcp.Connect(ctx, in.Clock, family(a), func(ctx context.Context) (io.ReadWriteCloser, error) {
		return in.Dialer.DialContext(ctx, "tcp", netip.AddrPortFrom(a, port).String())
	})
	if c != nil {
		defer c.Close()
	}
	if err != nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	obs := []observation.Observation{o}
	outcome := ""
	if c == nil {
		outcome = map[string]string{"timeout": "tcp_timeout", "connection_refused": "tcp_refused", "connection_reset": "tcp_reset", "inconclusive": "inconclusive"}[o.Failure()]
		if outcome == "" {
			data, _ := json.Marshal(struct {
				Type     string `json:"type"`
				Status   string `json:"status"`
				Duration int    `json:"duration_ms"`
				Failure  string `json:"failure"`
				Details  struct {
					Family string `json:"address_family"`
				} `json:"details"`
			}{"tcp_connect", "failed", o.DurationMS(), "inconclusive", struct {
				Family string `json:"address_family"`
			}{family(a)}})
			obs[0], err = observation.DecodeProtocol(observation.Contract, data, observation.ProtocolLimits{})
			if err != nil {
				return measurement.Output{}, measurement.ErrRunner
			}
			outcome = "inconclusive"
		}
	} else {
		sc, ok := c.(measurement.SessionConn)
		if !ok {
			return measurement.Output{}, measurement.ErrRunner
		}
		raw, _, err := sc.Session()
		if err != nil {
			return measurement.Output{}, measurement.ErrRunner
		}
		session, h, err := tlsprotocol.Handshake(ctx, in.Clock, family(a), raw, tlsprotocol.Config{ServerName: name, ALPN: p.ALPN(), Roots: r.Roots})
		if session != nil {
			defer session.Close()
		}
		if err != nil {
			return measurement.Output{}, measurement.ErrRunner
		}
		obs = append(obs, h)
		outcome = map[string]string{"timeout": "tls_timeout", "alert": "tls_alert", "connection_reset": "tcp_reset", "protocol_error": "protocol_error", "certificate_untrusted": "certificate_untrusted", "certificate_name_mismatch": "certificate_name_mismatch", "certificate_time_invalid": "certificate_time_invalid", "alpn_no_overlap": "alpn_no_overlap", "inconclusive": "inconclusive"}[h.Failure()]
		if h.Status() == "succeeded" {
			outcome = "handshake_succeeded"
		}
	}
	if outcome == "" || !in.RecordOutcome() {
		return measurement.Output{}, measurement.ErrRunner
	}
	return measurement.Output{Outcome: outcome, Observations: obs}, nil
}

// Each step narrows to its exact initial tuples before it receives a dialer.
// In particular an HTTP redirect may follow only this step's parent links,
// never the links of a different HTTP step admitted in the same profile.
func primitiveTargets(id string, params task.Parameters) []task.TargetTuple {
	switch id {
	case "DNS-01":
		p, ok := params.(dnsParameters)
		if !ok {
			return nil
		}
		out := []task.TargetTuple{{Hostname: p.Name()}}
		if p.Engine() != "system" {
			a, port := p.Resolver()
			out = append(out, task.TargetTuple{Address: a, Port: port, Scheme: p.Engine()})
		}
		return out
	case "TCP-01":
		p, ok := params.(tcpParameters)
		if !ok {
			return nil
		}
		a, port, ok := p.Target()
		if !ok {
			return nil
		}
		return []task.TargetTuple{{Address: a, Port: port, Scheme: "tcp"}}
	case "TLS-01":
		p, ok := params.(tlsParameters)
		if !ok {
			return nil
		}
		a, port, name, ok := p.Target()
		if !ok {
			return nil
		}
		return []task.TargetTuple{{Address: a, Port: port, Scheme: "tls", TLSServerName: name, ALPN: p.ALPN()}}
	case "HTTP-01":
		p, ok := params.(httpParameters)
		if ok {
			return []task.TargetTuple{p.Target()}
		}
	}
	return nil
}
