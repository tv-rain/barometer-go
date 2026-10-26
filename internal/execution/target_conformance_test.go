package execution_test

import (
	"context"
	"net"
	"net/netip"
	"strconv"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/execution"
	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/testkit/contextendpoint"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

type deniedDialRunner struct {
	begin, ignore bool
	out           measurement.Output
}

func (r deniedDialRunner) Run(ctx context.Context, in measurement.Input) (measurement.Output, error) {
	if r.begin {
		in.Start()
	}
	_, err := in.Dialer.DialContext(ctx, "tcp", "denied.example:443")
	if err == nil {
		return measurement.Output{}, measurement.ErrRunner
	}
	if r.ignore {
		return r.out, nil
	}
	return measurement.Output{}, measurement.ErrRunner
}

// Execution/result evidence for the target-validation rows only. The runner
// deliberately tries to evade preflight or swallow a refusal; zero DNS/proxy
// contact is independently covered by the production gate and proxy tests.
// requirement-ref: REQ-PROXY-0043
// requirement-ref: REQ-PROXY-0046
// requirement-ref: REQ-RESULT-0051
func TestCcProxyProductionTargetRefusal(t *testing.T) {
	for _, begin := range []bool{false, true} {
		for _, ignore := range []bool{false, true} {
			f := newFixture(t, schedule(1))
			p := scriptedProxy(t, loopback.ProxyConfig{})
			f.onProxy(t, p.Port())
			reg := task.TargetRegistration{ID: "initial", Test: tcp01, Role: task.MeasurementTarget, Tuple: task.TargetTuple{Hostname: "denied.example", Port: 443}, Paths: task.Paths{Proxy: true}, ExpiresAt: task.TimestampFromTime(f.clock.Now().Add(time.Hour))}
			f.plan.Targets = execution.NewTargets(tcp01, "proxy", []task.TargetRegistration{reg}, f.clock, nil)
			out := <-f.start(t, f.run, deniedDialRunner{begin: begin, ignore: ignore, out: connected(t).out})
			if out.err != nil {
				t.Fatal(out.err)
			}
			rec := f.records(t)[0]
			want := "not_run"
			if begin {
				want = "incomplete"
			}
			if rec.Execution.State != want || rec.Execution.Reason != "proxy_path_unsupported" || rec.Measurement != nil || p.Counts().Accepted != 0 {
				t.Fatalf("begin=%v ignore=%v: %+v proxy=%+v", begin, ignore, rec, p.Counts())
			}
		}
	}
}

// An approved explicit public IP is checked by the production gate and then
// routed by a synthetic CONNECT proxy to a real loopback echo server. No
// actual public address is contacted and no loopback exception is granted.
// requirement-ref: REQ-TRUST-0016
// requirement-ref: REQ-PROXY-0043
func TestCcProxyProductionAllowedTarget(t *testing.T) {
	srv := echoServer(t)
	_, ps, _ := net.SplitHostPort(srv.Address())
	port, _ := strconv.Atoi(ps)
	authority := net.JoinHostPort("8.8.8.8", ps)
	p := scriptedProxy(t, loopback.ProxyConfig{Resolve: map[string]string{authority: srv.Address()}})
	f := newFixture(t, schedule(1))
	f.onProxy(t, p.Port())
	params := echoOnce(srv)
	params.Address = authority
	f.plan.Parameters = params
	reg := task.TargetRegistration{ID: "initial", Test: tcp01, Role: task.MeasurementTarget, Tuple: task.TargetTuple{Address: netip.MustParseAddr("8.8.8.8"), Port: uint16(port)}, Paths: task.Paths{Proxy: true}, ExpiresAt: task.TimestampFromTime(f.clock.Now().Add(time.Hour))}
	f.plan.Targets = execution.NewTargets(tcp01, "proxy", []task.TargetRegistration{reg}, f.clock, nil)
	out := <-f.start(t, f.run, loopback.Exchange{})
	if out.err != nil {
		t.Fatal(out.err)
	}
	rec := f.records(t)[0]
	if rec.Execution.State != "completed" || len(p.Counts().Requests) != 1 || srv.Accepted() != 1 {
		t.Fatalf("result=%+v proxy=%+v contacts=%d", rec, p.Counts(), srv.Accepted())
	}
}

// A checked public authority is routed to loopback by a synthetic direct
// transport; the production gate never authorizes the loopback address itself.
// requirement-ref: REQ-TRUST-0016
func TestCcDirectProductionAllowedTarget(t *testing.T) {
	srv := echoServer(t)
	_, ps, _ := net.SplitHostPort(srv.Address())
	port, _ := strconv.Atoi(ps)
	authority := net.JoinHostPort("8.8.8.8", ps)
	f := newFixture(t, schedule(1))
	params := echoOnce(srv)
	params.Address = authority
	f.plan.Parameters = params
	f.plan.Dialers = contextendpoint.NewRoutes(map[string]string{authority: srv.Address()})
	reg := task.TargetRegistration{ID: "initial", Test: tcp01, Role: task.MeasurementTarget, Tuple: task.TargetTuple{Address: netip.MustParseAddr("8.8.8.8"), Port: uint16(port)}, Paths: task.Paths{Direct: true}, ExpiresAt: task.TimestampFromTime(f.clock.Now().Add(time.Hour))}
	f.plan.Targets = execution.NewTargets(tcp01, "direct", []task.TargetRegistration{reg}, f.clock, nil)
	out := <-f.start(t, f.run, loopback.Exchange{})
	if out.err != nil {
		t.Fatal(out.err)
	}
	rec := f.records(t)[0]
	if rec.Execution.State != "completed" || srv.Accepted() != 1 {
		t.Fatalf("state=%s contacts=%d", rec.Execution.State, srv.Accepted())
	}
}
