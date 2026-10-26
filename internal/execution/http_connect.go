package execution

import (
	"context"
	"errors"
	"io"
	"net"
	"net/netip"
	"time"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	"github.com/tv-rain/barometer-go/internal/protocol/tcp"
)

// ConnectStages uses the same destination gate and meter as DialContext, with
// separate DNS/TCP observations. No resolved address crosses into the runner.
// The explicit family makes one system lookup: no hidden retry or second lookup.
func (s *scope) ConnectStages(ctx context.Context, network, address string, clock measurement.Clock) (measurement.Conn, []observation.Observation, string, error) {
	return s.connectStages(ctx, network, address, clock, s.targets)
}
func (d targetDialer) ConnectStages(ctx context.Context, network, address string, clock measurement.Clock) (measurement.Conn, []observation.Observation, string, error) {
	return d.scope.connectStages(ctx, network, address, clock, d.check)
}

var errStageFailed = errors.New("execution: preceding stage failed")

func (s *scope) connectStages(ctx context.Context, network, address string, clock measurement.Clock, check TargetCheck) (measurement.Conn, []observation.Observation, string, error) {
	t, ok := check.(*Targets)
	if !ok || t == nil || t.resolver == nil || clock == nil || (network != "tcp" && network != "tcp4" && network != "tcp6") {
		return nil, nil, "", measurement.ErrRunner
	}
	trace := &lookupTrace{inner: t.resolver, clock: clock}
	gate := t.WithResolver(trace)
	family := "ipv4"
	host, _, _ := net.SplitHostPort(address)
	ip, _ := netip.ParseAddr(host)
	if network == "tcp6" || ip.Is6() {
		family = "ipv6"
	}
	var tcpObs observation.Observation
	var tcpDone bool
	c, err := s.dialVia(ctx, network, address, gate, func(op context.Context, checked string) (measurement.Conn, error) {
		conn, o, e := tcp.Connect(op, clock, family, func(cctx context.Context) (io.ReadWriteCloser, error) {
			c, err := s.inner.DialContext(cctx, network, checked)
			if reason, ok := pathFailure(err); ok {
				s.pathFailed(reason, err)
			}
			return c, err
		})
		if e != nil {
			return nil, e
		}
		tcpObs = o
		tcpDone = true
		if conn == nil {
			return nil, errStageFailed
		}
		return conn.(measurement.Conn), nil
	})
	var obs []observation.Observation
	if trace.called && !errors.Is(err, ErrTargetRejected) {
		status, failure := "succeeded", ""
		if trace.failure != "" {
			status, failure = "failed", trace.failure
		}
		query := "A"
		if network == "tcp6" {
			query = "AAAA"
		}
		o, e := observation.NewDNS(observation.Contract, status, trace.duration, failure, observation.DNSDetails{Engine: "system", QueryType: query, Responses: []observation.DNSResponse{}}, observation.ProtocolLimits{})
		if e != nil {
			return nil, nil, "", measurement.ErrRunner
		}
		obs = append(obs, o)
	}
	if tcpDone {
		obs = append(obs, tcpObs)
	}
	if errors.Is(err, errStageFailed) || errors.Is(err, measurement.ErrResolutionFailed) {
		err = nil
	}
	return c, obs, family, err
}

type lookupTrace struct {
	inner    SystemResolver
	clock    measurement.Clock
	called   bool
	duration int
	failure  string
}

func (t *lookupTrace) LookupNetIP(ctx context.Context, network, host string) ([]netip.Addr, error) {
	t.called = true
	start := t.clock.Now()
	op, cancel := context.WithCancel(ctx)
	defer cancel()
	stop := t.clock.AfterFunc(15*time.Second, cancel)
	defer stop()
	ips, err := t.inner.LookupNetIP(op, network, host)
	t.duration = max(0, int(t.clock.Now().Sub(start)/time.Millisecond)) / 250 * 250
	if err != nil || len(ips) == 0 || op.Err() != nil {
		t.failure = "resolver_error"
		var de *net.DNSError
		if errors.As(err, &de) && de.IsTimeout || op.Err() != nil && ctx.Err() == nil {
			t.failure = "timeout"
		}
		return nil, measurement.ErrResolutionFailed
	}
	return ips, nil
}
