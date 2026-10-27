package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/measurement/http01"
	"github.com/tv-rain/barometer-go/internal/measurement/net05"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/platform"
	tcpprotocol "github.com/tv-rain/barometer-go/internal/protocol/tcp"
	tlsprotocol "github.com/tv-rain/barometer-go/internal/protocol/tls"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

// net05ScriptDial scripts the n-th composite dial (1-based); an absent entry
// fails the test. Every address is recorded for the no-extra-contact checks.
type net05ScriptDial struct {
	t       *testing.T
	calls   atomic.Int32
	script  map[int]func(context.Context) (measurement.Conn, error)
	address []string
}

func (d *net05ScriptDial) Dialer(int) (measurement.Dialer, error) { return d, nil }
func (d *net05ScriptDial) DialContext(ctx context.Context, _, address string) (measurement.Conn, error) {
	n := int(d.calls.Add(1))
	d.address = append(d.address, address)
	f, ok := d.script[n]
	if !ok {
		d.t.Errorf("unexpected dial %d to %s", n, address)
		return nil, syscall.ECONNREFUSED
	}
	return f(ctx)
}

func net05LoopbackConn(address string) func(context.Context) (measurement.Conn, error) {
	return func(ctx context.Context) (measurement.Conn, error) {
		var nd net.Dialer
		return nd.DialContext(ctx, "tcp", address)
	}
}

// M-NET-05-TERMINATION: a primitive protocol timeout (TCP connect and TLS
// handshake operation bounds, driven by the injected clock below the outer
// deadline), a refusal and a TLS certificate failure inside a step each yield
// a completed step carrying the primitive negative outcome, not an
// interruption; later steps still run once, and the composite completes.
// requirement-ref: REQ-NET-05-0046
// requirement-ref: REQ-NET-05-0047
// requirement-ref: REQ-NET-05-0061
func TestMNet05StepNegativeOutcomesStayCompleted(t *testing.T) {
	ca, err := loopback.NewTestCA()
	if err != nil {
		t.Fatal(err)
	}
	other, err := loopback.NewTestCA()
	if err != nil {
		t.Fatal(err)
	}
	untrustedCert, err := other.Issue("example.com", false)
	if err != nil {
		t.Fatal(err)
	}
	stall, err := loopback.ListenTLS(loopback.TLSConfig{Mode: loopback.TLSStall})
	if err != nil {
		t.Fatal(err)
	}
	defer stall.Close()
	untrusted, err := loopback.ListenTLS(loopback.TLSConfig{Certificate: untrustedCert})
	if err != nil {
		t.Fatal(err)
	}
	defer untrusted.Close()
	tcp := `{"ip_address":"8.8.8.8","port":443}`
	tls := `{"ip_address":"8.8.8.8","port":443,"server_name":"example.com"}`
	steps := []string{compositeStep("tcp-timeout", "TCP-01", tcp), compositeStep("tls-timeout", "TLS-01", tls), compositeStep("tls-untrusted", "TLS-01", tls), compositeStep("tcp-refused", "TCP-01", tcp), compositeStep("tcp-connected", "TCP-01", tcp)}
	d := &net05ScriptDial{t: t}
	p, clock := compositePlan(t, steps, platform.AvailabilityYes, false, d)
	stalled := make(chan struct{})
	d.script = map[int]func(context.Context) (measurement.Conn, error){
		1: func(ctx context.Context) (measurement.Conn, error) {
			_ = clock.Advance(tcpprotocol.OperationBound)
			<-ctx.Done()
			return nil, ctx.Err()
		},
		2: func(ctx context.Context) (measurement.Conn, error) {
			go func() {
				defer close(stalled)
				wait, stop := context.WithTimeout(context.Background(), 5*time.Second)
				defer stop()
				if stall.Wait(wait, func(c loopback.TLSCounts) bool { return c.In > 0 }) == nil {
					_ = clock.Advance(tlsprotocol.OperationBound)
				}
			}()
			return net05LoopbackConn(stall.Address())(ctx)
		},
		3: net05LoopbackConn(untrusted.Address()),
		4: func(context.Context) (measurement.Conn, error) { return nil, syscall.ECONNREFUSED },
		5: net05LoopbackConn(untrusted.Address()),
	}
	a, u, v, err := (&Engine{clock: clock}).attempt(context.Background(), p, &net05.Runner{Roots: ca.Pool()}, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
	<-stalled
	if err != nil {
		t.Fatal(err)
	}
	m := v.View().Measurement
	if a.Phase() != Completed || a.Reason() != "" || m == nil || m.Outcome != "completed" || m.StoppedBeforeStep != "" || len(m.Steps) != 5 {
		t.Fatalf("phase=%v reason=%s measurement=%#v", a.Phase(), a.Reason(), m)
	}
	for i, want := range []string{"timeout", "tls_timeout", "certificate_untrusted", "connection_refused", "connected"} {
		s := m.Steps[i]
		if s.State != "completed" || s.Measurement == nil || s.Measurement.Outcome != want {
			t.Fatalf("step %d: %#v", i, s)
		}
	}
	if d.calls.Load() != 5 || u.Connections != 5 || clock.Timers() != 0 {
		t.Fatalf("dials=%d usage=%#v timers=%d", d.calls.Load(), u, clock.Timers())
	}
	wait, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	// The final TCP-01 connect may close before the server's accept loop runs,
	// so wait for the expected accept count with no open socket.
	if stall.Wait(wait, func(c loopback.TLSCounts) bool { return c.Accepted == 1 && c.Open == 0 }) != nil ||
		untrusted.Wait(wait, func(c loopback.TLSCounts) bool { return c.Accepted == 2 && c.Open == 0 }) != nil {
		t.Fatalf("socket leaked or retried: stall %+v untrusted %+v", stall.Counts(), untrusted.Counts())
	}
	if stall.Counts().Accepted != 1 || untrusted.Counts().Accepted != 2 || untrusted.Counts().Plaintext != 0 {
		t.Fatal("retried or carried application data")
	}
	raw, err := v.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(raw), `"network_context"`) != 1 || strings.Count(string(raw), `"execution"`) != 1 || strings.Count(string(raw), `"schema_version"`) != 1 {
		t.Fatalf("nested envelope: %s", raw)
	}
}

// M-NET-05-VERSIONS receiver side: the emitted profile validates only with
// schema_version 0.1.0, NET-05 0.1.0 / result 0.1.0 and the exact registered
// primitive tuple in each step; an unknown or malformed schema, test, result
// or step version, a relabelled step, or an unregistered validation state is
// refused instead of being interpreted best-effort.
// requirement-ref: REQ-NET-05-0039
// requirement-ref: REQ-NET-05-0043
// requirement-ref: REQ-NET-05-0044
func TestMNet05ReceiverRejectsVersions(t *testing.T) {
	d := &compositeDial{fail: true}
	p, clock := compositePlan(t, compositeTCP(), platform.AvailabilityYes, false, d)
	a, _, v, err := (&Engine{clock: clock}).attempt(context.Background(), p, &net05.Runner{}, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
	if err != nil || a.Phase() != Completed {
		t.Fatal(err, a.Phase())
	}
	raw, err := v.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var wire struct {
		Schema string `json:"schema_version"`
		Test   struct {
			ID, Version   string
			ResultVersion string `json:"result_version"`
		}
		Measurement struct {
			Validation struct{ State, Source string } `json:"network_validation"`
			Steps      []struct {
				ID   string `json:"step_id"`
				Test struct {
					ID, Version   string
					ResultVersion string `json:"result_version"`
				}
			}
		}
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatal(err)
	}
	if wire.Schema != "0.1.0" || wire.Test.ID != "NET-05" || wire.Test.Version != "0.1.0" || wire.Test.ResultVersion != "0.1.0" || wire.Measurement.Validation.State != "validated" || wire.Measurement.Validation.Source != "platform" || len(wire.Measurement.Steps) != 3 {
		t.Fatalf("emitted identity: %s", raw)
	}
	for i, s := range wire.Measurement.Steps {
		if s.ID != fmt.Sprint("s", i) || s.Test.ID != "TCP-01" || s.Test.Version != "0.1.0" || s.Test.ResultVersion != "0.1.0" {
			t.Fatalf("step %d identity: %s", i, raw)
		}
	}
	if _, err := p.Result.Decode(raw); err != nil {
		t.Fatal("emitted result rejected", err)
	}
	obj := func(v map[string]any, path ...any) map[string]any {
		var cur any = v
		for _, k := range path {
			switch k := k.(type) {
			case string:
				cur = cur.(map[string]any)[k]
			case int:
				cur = cur.([]any)[k]
			}
		}
		return cur.(map[string]any)
	}
	type mutation struct {
		name string
		edit func(map[string]any)
	}
	var mutations []mutation
	for _, bad := range []any{"0.1.1", "0.2.0", "1.0.0", "9.0.0", "0.1", "v0.1.0", "0.1.0 ", "", 0.1, nil} {
		bad := bad
		mutations = append(mutations,
			mutation{fmt.Sprint("schema_version=", bad), func(v map[string]any) { v["schema_version"] = bad }},
			mutation{fmt.Sprint("test.version=", bad), func(v map[string]any) { obj(v, "test")["version"] = bad }},
			mutation{fmt.Sprint("test.result_version=", bad), func(v map[string]any) { obj(v, "test")["result_version"] = bad }},
		)
		for i := 0; i < 3; i++ {
			i := i
			mutations = append(mutations,
				mutation{fmt.Sprintf("steps[%d].test.version=%v", i, bad), func(v map[string]any) { obj(v, "measurement", "steps", i, "test")["version"] = bad }},
				mutation{fmt.Sprintf("steps[%d].test.result_version=%v", i, bad), func(v map[string]any) {
					obj(v, "measurement", "steps", i, "test")["result_version"] = bad
				}},
			)
		}
	}
	for _, id := range []string{"NET-01", "NET-03", "DNS-03", "TCP-01", "TLS-01", "HTTP-01", "net-05", ""} {
		mutations = append(mutations, mutation{"test.id=" + id, func(v map[string]any) { obj(v, "test")["id"] = id }})
	}
	for _, id := range []string{"NET-05", "NET-01", "NET-03", "DNS-03", "DNS-01", "TLS-01", "HTTP-01", "tcp-01", ""} {
		mutations = append(mutations, mutation{"steps[2].test.id=" + id, func(v map[string]any) { obj(v, "measurement", "steps", 2, "test")["id"] = id }})
	}
	for _, field := range []string{"version", "result_version", "id"} {
		mutations = append(mutations,
			mutation{"missing test." + field, func(v map[string]any) { delete(obj(v, "test"), field) }},
			mutation{"missing steps[0].test." + field, func(v map[string]any) { delete(obj(v, "measurement", "steps", 0, "test"), field) }},
		)
	}
	mutations = append(mutations,
		mutation{"missing schema_version", func(v map[string]any) { delete(v, "schema_version") }},
		mutation{"steps[0].test extra", func(v map[string]any) { obj(v, "measurement", "steps", 0, "test")["extra"] = true }},
		mutation{"steps[0] common envelope", func(v map[string]any) { obj(v, "measurement", "steps", 0)["schema_version"] = "0.1.0" }},
		mutation{"steps[0].measurement network_context", func(v map[string]any) {
			obj(v, "measurement", "steps", 0, "measurement")["network_context"] = map[string]any{}
		}},
		mutation{"network_validation.state=other", func(v map[string]any) { obj(v, "measurement", "network_validation")["state"] = "other" }},
		mutation{"network_validation.state=not_validated", func(v map[string]any) { obj(v, "measurement", "network_validation")["state"] = "not_validated" }},
		mutation{"network_validation.state missing", func(v map[string]any) { delete(obj(v, "measurement", "network_validation"), "state") }},
		mutation{"network_validation.source=other", func(v map[string]any) { obj(v, "measurement", "network_validation")["source"] = "other" }},
	)
	for _, m := range mutations {
		var value map[string]any
		if err := json.Unmarshal(raw, &value); err != nil {
			t.Fatal(err)
		}
		m.edit(value)
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := p.Result.Decode(data); err == nil {
			t.Errorf("%s accepted", m.name)
		}
	}
}

// net05Routes maps an authorized literal authority to a loopback server and
// counts every dial, so the aggregate connection meter is observable.
type net05Routes struct {
	routes map[string]string
	calls  atomic.Int32
}

func (d *net05Routes) Dialer(int) (measurement.Dialer, error) { return d, nil }
func (d *net05Routes) DialContext(ctx context.Context, network, address string) (measurement.Conn, error) {
	d.calls.Add(1)
	target, ok := d.routes[address]
	if !ok {
		return nil, syscall.ECONNREFUSED
	}
	var nd net.Dialer
	return nd.DialContext(ctx, network, target)
}

// M-NET-05-LIMITS registered connection ceiling against actual aggregate use.
// Four HTTP-01 steps each follow two registered redirects (three connections
// per step), so the profile uses exactly 12 connections; a following TCP-01
// step needs a 13th. Declared-maxima admission is an approved deferral, so
// such a profile is admitted and the runtime meter alone must stop it.
// requirement-ref: REQ-NET-05-0037
// requirement-ref: REQ-LIMIT-0019
func TestMNet05ConnectionCeiling(t *testing.T) {
	ca, err := loopback.NewTestCA()
	if err != nil {
		t.Fatal(err)
	}
	ips := []string{"8.8.8.8", "9.9.9.9", "1.1.1.1"}
	responses := []string{
		"HTTP/1.1 302 Found\r\nLocation: https://9.9.9.9/check\r\nContent-Length: 0\r\n\r\n",
		"HTTP/1.1 307 Redirect\r\nLocation: https://1.1.1.1/check\r\nContent-Length: 0\r\n\r\n",
		"HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nok",
	}
	httpStep := func(id string) string {
		return fmt.Sprintf(`{"step_id":%q,"test_id":"HTTP-01","test_version":"0.1.0","input":{"url":"https://8.8.8.8/check","expected_status":200},"options":{"follow_redirects":true,"max_redirects":2}}`, id)
	}
	for _, tc := range []struct {
		name        string
		tcpStep     bool
		connections int
		steps       int    // emitted steps
		interrupted bool   // the last emitted step is the interrupted one
		stop        string // stopped_before_step
	}{
		{"below need", false, 11, 4, true, ""},
		{"at ceiling", false, 12, 4, false, ""},
		// The TCP-01 step starts and its 13th connection is refused by the
		// meter: it is the single final incomplete step, with no measurement.
		{"above ceiling", true, 12, 5, true, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			routes := &net05Routes{routes: map[string]string{}}
			for i, ip := range ips {
				cert, err := ca.Issue(ip, false)
				if err != nil {
					t.Fatal(err)
				}
				srv, err := loopback.ListenHTTP(loopback.HTTPConfig{Certificate: cert, Raw: []byte(responses[i])})
				if err != nil {
					t.Fatal(err)
				}
				defer srv.Close()
				routes.routes[ip+":443"] = srv.Address()
			}
			steps := []string{httpStep("h0"), httpStep("h1"), httpStep("h2"), httpStep("h3")}
			if tc.tcpStep {
				steps = append(steps, compositeStep("tcp", "TCP-01", `{"ip_address":"8.8.8.8","port":443}`))
			}
			p, clock := compositePlan(t, steps, platform.AvailabilityYes, false, routes)
			p.Limits.Connections = tc.connections
			// Re-register the shared initial tuple with its two-hop redirect chain.
			test := task.TestVersion{ID: "HTTP-01", Version: "0.1.0"}
			var initial task.TargetTuple
			for _, s := range p.Parameters.(net05.Parameters).CompositeSteps() {
				if v, ok := s.Parameters.(http01.Parameters); ok {
					initial = v.Target()
				}
			}
			expires := task.TimestampFromTime(clock.Now().Add(time.Hour))
			second, _ := http01.NormalizeURL("https://9.9.9.9/check")
			third, _ := http01.NormalizeURL("https://1.1.1.1/check")
			// A redirect hop inherits the step's checks (REQ-HTTP-01-0051).
			second.ExpectedStatus, third.ExpectedStatus = initial.ExpectedStatus, initial.ExpectedStatus
			regs := []task.TargetRegistration{
				{ID: "initial", Test: test, Role: task.MeasurementTarget, Tuple: initial, Paths: task.Paths{Direct: true}, ExpiresAt: expires, RedirectIDs: []string{"second"}},
				{ID: "second", Test: test, Role: task.RedirectTarget, Tuple: second, Paths: task.Paths{Direct: true}, ExpiresAt: expires, RedirectIDs: []string{"third"}},
				{ID: "third", Test: test, Role: task.RedirectTarget, Tuple: third, Paths: task.Paths{Direct: true}, ExpiresAt: expires},
				{ID: "tcp", Test: task.TestVersion{ID: "TCP-01", Version: "0.1.0"}, Role: task.MeasurementTarget, Tuple: task.TargetTuple{Address: netip.MustParseAddr("8.8.8.8"), Port: 443, Scheme: "tcp"}, Paths: task.Paths{Direct: true}, ExpiresAt: expires},
			}
			p.Targets = NewTargets(p.Test, "direct", regs, clock, nil).WithResolver(net05Resolver{})
			a, u, v, err := (&Engine{clock: clock}).attempt(context.Background(), p, &net05.Runner{Roots: ca.Pool()}, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
			if err != nil {
				t.Fatal(err)
			}
			m := v.View().Measurement
			if m == nil || len(m.Steps) != tc.steps {
				t.Fatalf("measurement %#v", m)
			}
			want := min(tc.connections, 12)
			if u.Connections != want || int(routes.calls.Load()) != want {
				t.Fatalf("connections used=%d dialed=%d want %d", u.Connections, routes.calls.Load(), want)
			}
			for i, s := range m.Steps {
				last := i == len(m.Steps)-1
				if tc.interrupted && last {
					if s.State != "incomplete" || s.Measurement != nil {
						t.Fatalf("interrupted step %d: %#v", i, s)
					}
					continue
				}
				if s.State != "completed" || s.Measurement == nil || s.Measurement.Outcome != "expected_response" {
					t.Fatalf("step %d: %#v", i, s)
				}
			}
			if tc.interrupted || tc.stop != "" {
				if a.Phase() != Incomplete || a.Reason() != task.ResourceLimit || m.Outcome != "incomplete" || m.StoppedBeforeStep != tc.stop {
					t.Fatalf("phase=%v reason=%s outcome=%s stop=%q", a.Phase(), a.Reason(), m.Outcome, m.StoppedBeforeStep)
				}
			} else if a.Phase() != Completed || m.Outcome != "completed" {
				t.Fatalf("phase=%v reason=%s outcome=%s", a.Phase(), a.Reason(), m.Outcome)
			}
			if clock.Timers() != 0 {
				t.Fatal("leaked timers")
			}
		})
	}
}

// M-NET-05-LIMITS registered 262,144-byte received ceiling against actual
// aggregate use: four HTTP-01 steps each receive R application bytes with the
// task limit at the ceiling. R = 65,536 reaches the ceiling exactly and
// completes; one byte below completes; one byte above interrupts the fourth
// step with resource_limit while the earlier steps keep their results. The
// 65,536-byte sent ceiling cannot be reached by registered Draft primitives
// (no request body; at most 8 steps of bounded GET requests), so only its
// admission bound is exercised, at the Client.
// requirement-ref: REQ-NET-05-0037
// requirement-ref: REQ-LIMIT-0034
func TestMNet05ReceivedCeiling(t *testing.T) {
	ca, err := loopback.NewTestCA()
	if err != nil {
		t.Fatal(err)
	}
	cert, err := ca.Issue("8.8.8.8", false)
	if err != nil {
		t.Fatal(err)
	}
	step := func(id string) string {
		return fmt.Sprintf(`{"step_id":%q,"test_id":"HTTP-01","test_version":"0.1.0","input":{"url":"https://8.8.8.8/check","expected_status":200},"options":{"max_body_bytes":65536}}`, id)
	}
	for _, delta := range []int{-1, 0, 1} {
		t.Run(fmt.Sprint(delta), func(t *testing.T) {
			size := 65536 + delta
			head := func(n int) string { return fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Length: %d\r\n\r\n", n) }
			body := size - len(head(size))
			body = size - len(head(body))
			raw := head(body) + strings.Repeat("x", body)
			if len(raw) != size {
				t.Fatalf("fixture %d != %d", len(raw), size)
			}
			srv, err := loopback.ListenHTTP(loopback.HTTPConfig{Certificate: cert, Raw: []byte(raw)})
			if err != nil {
				t.Fatal(err)
			}
			defer srv.Close()
			p, clock := compositePlan(t, []string{step("h0"), step("h1"), step("h2"), step("h3")}, platform.AvailabilityYes, false, loopbackDial{srv.Address()})
			if p.Limits.BytesReceived != 262144 {
				t.Fatal("plan is not at the registered ceiling")
			}
			a, u, v, err := (&Engine{clock: clock}).attempt(context.Background(), p, &net05.Runner{Roots: ca.Pool()}, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
			if err != nil {
				t.Fatal(err)
			}
			m := v.View().Measurement
			if u.BytesReceived > 262144 || m == nil || len(m.Steps) != 4 {
				t.Fatalf("usage=%#v measurement=%#v", u, m)
			}
			if delta > 0 {
				if a.Phase() != Incomplete || a.Reason() != task.ResourceLimit || m.Outcome != "incomplete" || m.Steps[3].State != "incomplete" || m.Steps[3].Measurement != nil {
					t.Fatalf("phase=%v reason=%s measurement=%#v", a.Phase(), a.Reason(), m)
				}
				for _, s := range m.Steps[:3] {
					if s.State != "completed" || s.Measurement == nil {
						t.Fatal("earlier step lost")
					}
				}
			} else if a.Phase() != Completed || m.Outcome != "completed" || u.BytesReceived != 4*size {
				t.Fatalf("phase=%v reason=%s usage=%#v", a.Phase(), a.Reason(), u)
			}
			if clock.Timers() != 0 {
				t.Fatal("leaked timers")
			}
		})
	}
}

// M-NET-05-LIMITS registered 60,000 ms ceiling below and at the boundary: with
// the plan deadline at the ceiling, 60,000 ms less one nanosecond of clock
// time lets the profile complete; reaching 60,000 ms interrupts it. A task
// above the ceiling is refused at admission (Client test).
// requirement-ref: REQ-NET-05-0037
func TestMNet05DurationCeiling(t *testing.T) {
	for _, elapsed := range []time.Duration{60*time.Second - time.Nanosecond, 60 * time.Second} {
		d := &compositeDial{}
		p, clock := compositePlan(t, compositeTCP(), platform.AvailabilityYes, false, d)
		if p.Deadline != 60*time.Second || p.Limits.DurationMS != 60000 {
			t.Fatal("plan is not at the registered ceiling")
		}
		d.hook = func(n int) {
			if n == 2 {
				_ = clock.Advance(elapsed)
			}
		}
		a, _, v, err := (&Engine{clock: clock}).attempt(context.Background(), p, &net05.Runner{}, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
		if err != nil {
			t.Fatal(err)
		}
		m := v.View().Measurement
		if elapsed < 60*time.Second {
			if a.Phase() != Completed || m.Outcome != "completed" || d.calls.Load() != 3 {
				t.Fatalf("below ceiling: %v %s %#v", a.Phase(), a.Reason(), m)
			}
		} else if a.Phase() != Incomplete || a.Reason() != task.AttemptDeadlineExceeded || d.calls.Load() != 2 {
			t.Fatalf("at ceiling: %v %s calls=%d", a.Phase(), a.Reason(), d.calls.Load())
		}
		if clock.Timers() != 0 {
			t.Fatal("leaked timers")
		}
	}
}
