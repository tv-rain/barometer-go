package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/netip"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/measurement/dns01"
	"github.com/tv-rain/barometer-go/internal/measurement/http01"
	"github.com/tv-rain/barometer-go/internal/measurement/net05"
	"github.com/tv-rain/barometer-go/internal/measurement/tcp01"
	"github.com/tv-rain/barometer-go/internal/measurement/tls01"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/networkcontext"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/taskdecode"
	"github.com/tv-rain/barometer-go/internal/testkit"
	"github.com/tv-rain/barometer-go/internal/testkit/capability"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

func compositePlan(t *testing.T, steps []string, validated platform.Availability, proxy bool, dial DialerSource) (Plan, *testkit.WaitClock) {
	t.Helper()
	clock := testkit.NewWaitClock(time.Now())
	codec, err := net05.New([]taskdecode.TestCodec{dns01.Codec{}, tcp01.Codec{}, tls01.Codec{}, http01.Codec{}})
	if err != nil {
		t.Fatal(err)
	}
	raw := `{"profile":{"profile_id":"baseline","profile_version":1,"steps":[` + strings.Join(steps, ",") + `]}}`
	def := codec.Definition()
	lim := jsondecode.Limits{MaxBytes: 65536, MaxDepth: 32, MaxNodes: 4000}
	in, err := jsondecode.Decode([]byte(raw), def.Input, lim)
	if err != nil {
		t.Fatal(err)
	}
	opt, err := jsondecode.Decode([]byte(`{}`), def.Options, lim)
	if err != nil {
		t.Fatal(err)
	}
	params, err := codec.Normalize(in, opt)
	if err != nil {
		t.Fatal(err)
	}
	network := capability.NewNetwork(platform.AccessState{Generation: 1, Type: platform.NetworkWiFi, IPv4: platform.AvailabilityYes, Validated: validated})
	var src *networkcontext.Source
	path := "direct"
	if proxy {
		path = "proxy"
		src, err = networkcontext.NewProxy(network, 99, "outline_mobileproxy")
	} else {
		src, err = networkcontext.NewDirect(network)
	}
	if err != nil {
		t.Fatal(err)
	}
	snap, err := src.Capture(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	p := Plan{TaskID: "composite", Test: task.TestVersion{ID: "NET-05", Version: "0.1.0"}, Parameters: params, RequireCompositeValidation: true, Network: task.NetworkRequirements{VPNPolicy: "any"}, Deadline: 60 * time.Second, Limits: task.AttemptLimits{DurationMS: 60000, Connections: 12, BytesSent: 65536, BytesReceived: 262144}, Context: src, Dialers: dial, Implementation: model.Implementation{LibraryVersion: "0.1.0", Platform: "android"}}
	id, version := params.(net05.Parameters).Profile()
	profile := &result.ProfilePlan{ID: id, Version: version, RequireValidation: true}
	var regs []task.TargetRegistration
	for i, s := range params.(net05.Parameters).CompositeSteps() {
		output := result.Output{Test: model.Test{ID: s.Test.ID, Version: "0.1.0", ResultVersion: "0.1.0"}}
		tuple := task.TargetTuple{Address: netip.MustParseAddr("8.8.8.8"), Port: 443, Scheme: "tcp"}
		switch v := s.Parameters.(type) {
		case dns01.Parameters:
			output.DNSMode = v.Engine()
			output.QueryType = v.QueryType()
			output.Limits.RetainDNSAnswers = true
			tuple = task.TargetTuple{Hostname: v.Name()}
		case tls01.Parameters:
			a, port, name, _ := v.Target()
			tuple = task.TargetTuple{Address: a, Port: port, Scheme: "tls", TLSServerName: name, ALPN: v.ALPN()}
		case http01.Parameters:
			tuple = v.Target()
			output.Limits = v.Limits()
			output.HTTPS = true
		}
		regs = append(regs, task.TargetRegistration{ID: fmt.Sprint("reg-", i), Test: s.Test, Role: task.MeasurementTarget, Tuple: tuple, Paths: task.Paths{Direct: true, Proxy: true}, ExpiresAt: task.TimestampFromTime(clock.Now().Add(time.Hour))})
		profile.Steps = append(profile.Steps, result.StepPlan{ID: s.ID, Output: output})
	}
	p.Targets = NewTargets(p.Test, path, regs, clock, nil).WithResolver(net05Resolver{})
	p.Result, err = result.New(result.Binding{TaskID: p.TaskID, Output: result.Output{Test: model.Test{ID: "NET-05", Version: "0.1.0", ResultVersion: "0.1.0"}}, Profile: profile, MaxAttempts: 1, MaxDurationMS: 60000, Implementation: p.Implementation, ContextDecoder: src.Decoder(), Path: snap.Path(), AccessGeneration: snap.Generation()})
	if err != nil {
		t.Fatal(err)
	}
	return p, clock
}
func compositeStep(id, test, input string) string {
	return fmt.Sprintf(`{"step_id":%q,"test_id":%q,"test_version":"0.1.0","input":%s,"options":{}}`, id, test, input)
}
func compositeTCP() []string {
	var s []string
	for i := 0; i < 3; i++ {
		s = append(s, compositeStep(fmt.Sprint("s", i), "TCP-01", `{"ip_address":"8.8.8.8","port":443}`))
	}
	return s
}

type net05Resolver struct{}

func (net05Resolver) LookupNetIP(context.Context, string, string) ([]netip.Addr, error) {
	return []netip.Addr{netip.MustParseAddr("8.8.8.8")}, nil
}

type compositeDial struct {
	calls, closed atomic.Int32
	hook          func(int)
	closeHook     func(int)
	fail          bool
}

func (d *compositeDial) Dialer(int) (measurement.Dialer, error) { return d, nil }
func (d *compositeDial) DialContext(ctx context.Context, _, _ string) (measurement.Conn, error) {
	n := int(d.calls.Add(1))
	if d.hook != nil {
		d.hook(n)
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if d.fail {
		return nil, syscall.ECONNREFUSED
	}
	return &compositeConn{dial: d, n: n}, nil
}

type compositeConn struct {
	dial   *compositeDial
	n      int
	closed bool
}

func (c *compositeConn) Read([]byte) (int, error)    { return 0, io.EOF }
func (c *compositeConn) Write(p []byte) (int, error) { return len(p), nil }
func (c *compositeConn) SetDeadline(time.Time) error { return nil }
func (c *compositeConn) Close() error {
	if !c.closed {
		c.closed = true
		c.dial.closed.Add(1)
		if c.dial.closeHook != nil {
			c.dial.closeHook(c.n)
		}
	}
	return nil
}

type compositeStepHook struct {
	TargetCheck
	targets *Targets
	hook    func()
}

func (g compositeStepHook) Step(test task.TestVersion, tuples []task.TargetTuple) TargetCheck {
	g.hook()
	return g.targets.Step(test, tuples)
}

// Executed synthetic runner/engine boundary matrix for M-NET-05-TERMINATION
// and the GO-SPEC-ISSUE-004 case 3. The second cancellation must never win.
// requirement-ref: REQ-NET-05-0045
// requirement-ref: REQ-NET-05-0047
// requirement-ref: REQ-NET-05-0061
// requirement-ref: REQ-RESULT-0051
// requirement-ref: REQ-RESULT-0052
func TestMNet05Termination(t *testing.T) {
	causes := []struct {
		cause  error
		reason task.ExecutionReason
	}{{context.Canceled, task.CancelledByHost}, {errDeadline, task.AttemptDeadlineExceeded}, {errNetworkChanged, task.NetworkChanged}, {errBudget, task.ResourceLimit}, {pathError{task.PlatformSuspended}, task.PlatformSuspended}, {ErrProxyReconfigured, task.ProxyReconfigured}, {errProxyUnavailable, task.ProxyUnavailable}, {errProxyConnect, task.ProxyConnectFailed}}
	for _, tc := range causes {
		for _, boundary := range []string{"before", "during", "after"} {
			for k := 0; k < 3; k++ {
				t.Run(fmt.Sprintf("%s/%s/%d", tc.reason, boundary, k), func(t *testing.T) {
					ctx, cancel := context.WithCancelCause(context.Background())
					defer cancel(nil)
					interrupt := func() { cancel(tc.cause); cancel(errNetworkChanged) }
					d := &compositeDial{}
					p, clock := compositePlan(t, compositeTCP(), platform.AvailabilityYes, false, d)
					if boundary == "during" {
						d.hook = func(n int) {
							if n == k+1 {
								interrupt()
							}
						}
					}
					if boundary == "after" {
						d.closeHook = func(n int) {
							if n == k+1 {
								interrupt()
							}
						}
					}
					if boundary == "before" {
						count := 0
						g := p.Targets.(*Targets)
						p.Targets = compositeStepHook{g, g, func() {
							if count == k {
								interrupt()
							}
							count++
						}}
					}
					a, _, v, err := (&Engine{clock: clock}).attempt(ctx, p, &net05.Runner{}, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
					if err != nil {
						t.Fatal(err)
					}
					m := v.View().Measurement
					if m == nil {
						t.Fatal("missing partial profile")
					}
					if boundary == "after" && k == 2 {
						if a.Phase() != Completed || m.Outcome != "completed" || len(m.Steps) != 3 || a.Reason() != "" {
							t.Fatal("cleanup replaced final outcome")
						}
						return
					}
					if a.Phase() != Incomplete || a.Reason() != tc.reason || m.Outcome != "incomplete" {
						t.Fatalf("state=%v reason=%s outcome=%s", a.Phase(), a.Reason(), m.Outcome)
					}
					n := k
					if boundary != "before" {
						n++
					}
					if len(m.Steps) != n {
						t.Fatalf("steps=%d want %d", len(m.Steps), n)
					}
					for i, s := range m.Steps {
						want := "completed"
						if boundary == "during" && i == k {
							want = "incomplete"
						}
						if s.State != want || (s.Measurement == nil) != (want == "incomplete") {
							t.Fatalf("step %d: %#v", i, s)
						}
					}
					next := n
					wantStop := ""
					if next < 3 {
						wantStop = fmt.Sprint("s", next)
					}
					if m.StoppedBeforeStep != wantStop {
						t.Fatalf("stopped %s want %s", m.StoppedBeforeStep, wantStop)
					}
					if int(d.calls.Load()) > k+1 || d.closed.Load() != d.calls.Load() && boundary != "during" {
						t.Fatal("later contact or leaked connection")
					}
					raw, _ := v.MarshalJSON()
					if strings.Contains(string(raw), "8.8.8.8") || strings.Contains(string(raw), "network_changed\",\"observations") {
						t.Fatal("target or execution reason leaked into measurement")
					}
				})
			}
		}
	}
}

// requirement-ref: REQ-NET-05-0061
// requirement-ref: REQ-NET-05-0046
func TestMNet05Normalization(t *testing.T) {
	for _, validation := range []platform.Availability{platform.AvailabilityYes, platform.AvailabilityNo, platform.AvailabilityUnknown} {
		d := &compositeDial{fail: true}
		p, clock := compositePlan(t, compositeTCP(), validation, false, d)
		a, _, v, err := (&Engine{clock: clock}).attempt(context.Background(), p, &net05.Runner{}, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
		if err != nil {
			t.Fatal(err)
		}
		want := "completed"
		if validation == platform.AvailabilityNo {
			want = "network_not_validated"
		}
		if validation == platform.AvailabilityUnknown {
			want = "inconclusive"
		}
		m := v.View().Measurement
		if a.Phase() != Completed || a.Reason() != "" || m == nil || m.Outcome != want {
			t.Fatalf("%v: %#v", validation, v.View())
		}
		if validation != platform.AvailabilityYes && d.calls.Load() != 0 {
			t.Fatal("validation gate contacted target")
		}
		if validation == platform.AvailabilityYes {
			for _, s := range m.Steps {
				if s.Measurement.Outcome != "connection_refused" {
					t.Fatal("negative primitive stopped profile")
				}
			}
		}
	}
}

// requirement-ref: REQ-NET-05-0037
// requirement-ref: REQ-LIMIT-0019
func TestMNet05Limits(t *testing.T) {
	for _, limit := range []int{0, 1, 2, 3, 4} {
		d := &compositeDial{}
		p, clock := compositePlan(t, compositeTCP(), platform.AvailabilityYes, false, d)
		p.Limits.Connections = limit
		a, u, v, err := (&Engine{clock: clock}).attempt(context.Background(), p, &net05.Runner{}, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
		if err != nil {
			t.Fatal(err)
		}
		if u.Connections != min(limit, 3) || int(d.calls.Load()) != min(limit, 3) || d.closed.Load() != d.calls.Load() {
			t.Fatal("aggregate connection meter")
		}
		if limit < 3 {
			if a.Phase() != Incomplete || a.Reason() != task.ResourceLimit || v.View().Measurement == nil {
				t.Fatal("missing aggregate interruption")
			}
		} else if a.Phase() != Completed {
			t.Fatal("within bound did not complete")
		}
	}
}

// The three NET-05 cross-cutting rows use one original context and path even
// across independently checked primitive registration scopes.
// requirement-ref: REQ-NET-05-0040
// requirement-ref: REQ-CONTEXT-0025
func TestCcNet05PathContextAndProxyGeneration(t *testing.T) {
	steps := []string{compositeStep("dns", "DNS-01", `{"hostname":"example.com","record_type":"A"}`), compositeStep("tcp", "TCP-01", `{"ip_address":"8.8.8.8","port":443}`), compositeStep("tls", "TLS-01", `{"ip_address":"8.8.8.8","port":443,"server_name":"example.com"}`), compositeStep("http", "HTTP-01", `{"url":"https://8.8.8.8/check","expected_status":200}`)}
	for _, proxy := range []bool{false, true} {
		selected := steps
		if proxy {
			selected = steps[1:]
		} // System DNS is not CONNECT-compatible.
		d := &compositeDial{fail: true}
		p, clock := compositePlan(t, selected, platform.AvailabilityYes, proxy, d)
		a, _, v, err := (&Engine{clock: clock}).attempt(context.Background(), p, &net05.Runner{}, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
		if err != nil {
			t.Fatal(err)
		}
		view := v.View()
		if a.Phase() != Completed || view.Measurement == nil || len(view.Measurement.Steps) != len(selected) {
			t.Fatalf("profile failed: %#v", view)
		}
		raw, _ := v.MarshalJSON()
		var wire map[string]json.RawMessage
		if json.Unmarshal(raw, &wire) != nil {
			t.Fatal("wire")
		}
		if strings.Count(string(raw), `"network_context"`) != 1 {
			t.Fatal("duplicated context")
		}
		if proxy && view.NetworkContext.Access().NetworkChanged {
			t.Fatal("changed path")
		}
		for i, s := range view.Measurement.Steps {
			if s.Test.ID != p.Parameters.(net05.Parameters).CompositeSteps()[i].Test.ID || s.State != "completed" {
				t.Fatal("order/version")
			}
		}
	}
}

// Application HTTP bytes accumulate across steps at the plaintext boundary;
// CONNECT and TLS bytes remain excluded on the real loopback proxy path.
// requirement-ref: REQ-LIMIT-0034
// requirement-ref: REQ-NET-05-0037
func TestMNet05AggregateBytes(t *testing.T) {
	ca, err := loopback.NewTestCA()
	if err != nil {
		t.Fatal(err)
	}
	cert, err := ca.Issue("8.8.8.8", false)
	if err != nil {
		t.Fatal(err)
	}
	for _, proxy := range []bool{false, true} {
		t.Run(fmt.Sprint(proxy), func(t *testing.T) {
			srv, err := loopback.ListenHTTP(loopback.HTTPConfig{Certificate: cert, Raw: []byte("HTTP/1.1 200 OK\r\nContent-Length: 3\r\n\r\nabc")})
			if err != nil {
				t.Fatal(err)
			}
			defer srv.Close()
			var dial DialerSource = loopbackDial{srv.Address()}
			if proxy {
				tunnel, err := loopback.ListenProxy(loopback.ProxyConfig{Resolve: map[string]string{"8.8.8.8:443": srv.Address()}})
				if err != nil {
					t.Fatal(err)
				}
				defer tunnel.Close()
				dial, err = NewProxyDialers(tunnel.Host(), tunnel.Port())
				if err != nil {
					t.Fatal(err)
				}
			}
			steps := []string{compositeStep("first", "HTTP-01", `{"url":"https://8.8.8.8/check","expected_status":200}`), compositeStep("second", "HTTP-01", `{"url":"https://8.8.8.8/check","expected_status":200}`)}
			run := func(limits *task.AttemptLimits) (Attempt, measurement.Usage, model.Result) {
				p, clock := compositePlan(t, steps, platform.AvailabilityYes, proxy, dial)
				if limits != nil {
					p.Limits = *limits
				}
				a, u, v, e := (&Engine{clock: clock}).attempt(context.Background(), p, &net05.Runner{Roots: ca.Pool()}, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
				if e != nil {
					t.Fatal(e)
				}
				if clock.Timers() != 0 {
					t.Fatal("leaked timers")
				}
				return a, u, v.View()
			}
			a, base, _ := run(nil)
			if a.Phase() != Completed || base.Connections != 2 || base.BytesSent == 0 || base.BytesReceived != 2*len("HTTP/1.1 200 OK\r\nContent-Length: 3\r\n\r\nabc") {
				t.Fatalf("baseline meter: %#v", base)
			}
			for _, sent := range []bool{true, false} {
				for _, delta := range []int{-1, 0, 1} {
					limits := task.AttemptLimits{Connections: 12, BytesSent: 65536, BytesReceived: 262144}
					if sent {
						limits.BytesSent = base.BytesSent + delta
					} else {
						limits.BytesReceived = base.BytesReceived + delta
					}
					a, u, v := run(&limits)
					if u.BytesSent > limits.BytesSent || u.BytesReceived > limits.BytesReceived {
						t.Fatal("aggregate overrun")
					}
					if delta < 0 {
						if a.Phase() != Incomplete || a.Reason() != task.ResourceLimit || v.Measurement == nil || v.Measurement.Outcome != "incomplete" || v.Measurement.Steps[0].State != "completed" || v.Measurement.Steps[1].State != "incomplete" {
							t.Fatalf("sent=%v delta=%d state=%v reason=%v measurement=%#v", sent, delta, a.Phase(), a.Reason(), v.Measurement)
						}
					} else if a.Phase() != Completed {
						t.Fatalf("exact/above budget failed: %v %v %v", sent, delta, v)
					}
				}
			}
		})
	}
}

// A stricter task duration ends the aggregate, with the exact equality
// boundary driven by the same injected clock as every primitive operation.
// requirement-ref: REQ-NET-05-0037
func TestMNet05AggregateDeadline(t *testing.T) {
	for _, duration := range []time.Duration{time.Millisecond, 30 * time.Second, 60 * time.Second} {
		d := &compositeDial{}
		p, clock := compositePlan(t, compositeTCP(), platform.AvailabilityYes, false, d)
		p.Deadline = duration
		d.hook = func(n int) {
			if n == 2 {
				clock.Advance(duration - time.Nanosecond)
				clock.Advance(time.Nanosecond)
			}
		}
		a, _, v, err := (&Engine{clock: clock}).attempt(context.Background(), p, &net05.Runner{}, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
		if err != nil {
			t.Fatal(err)
		}
		if a.Phase() != Incomplete || a.Reason() != task.AttemptDeadlineExceeded || d.calls.Load() != 2 || len(v.View().Measurement.Steps) != 2 || clock.Timers() != 0 {
			t.Fatalf("aggregate deadline %s: %#v", duration, v.View())
		}
	}
}

// Another step's initial tuple cannot authorize contact or a redirect link.
// requirement-ref: REQ-NET-05-0019
// requirement-ref: REQ-HTTP-01-0051
func TestNet05StepTargetAndRedirectIsolation(t *testing.T) {
	clock := testkit.NewWaitClock(time.Now())
	test := task.TestVersion{ID: "HTTP-01", Version: "0.1.0"}
	first, _ := http01.NormalizeURL("https://8.8.8.8/first")
	second, _ := http01.NormalizeURL("https://8.8.8.8/second")
	dest, _ := http01.NormalizeURL("https://1.1.1.1/next")
	expires := task.TimestampFromTime(clock.Now().Add(time.Hour))
	regs := []task.TargetRegistration{{ID: "first", Test: test, Role: task.MeasurementTarget, Tuple: first, Paths: task.Paths{Direct: true}, ExpiresAt: expires}, {ID: "second", Test: test, Role: task.MeasurementTarget, Tuple: second, Paths: task.Paths{Direct: true}, ExpiresAt: expires, RedirectIDs: []string{"next"}}, {ID: "next", Test: test, Role: task.RedirectTarget, Tuple: dest, Paths: task.Paths{Direct: true}, ExpiresAt: expires}}
	gate := NewTargets(task.TestVersion{ID: "NET-05", Version: "0.1.0"}, "direct", regs, clock, nil)
	a := gate.Step(test, []task.TargetTuple{first})
	b := gate.Step(test, []task.TargetTuple{second})
	if a == nil || b == nil {
		t.Fatal("missing step gate")
	}
	if _, err := a.Redirect(dest); err == nil {
		t.Fatal("borrowed other step's redirect")
	}
	if _, err := b.Redirect(dest); err != nil {
		t.Fatal("own redirect rejected")
	}
	if _, err := a.Prepare(context.Background(), "tcp", "1.1.1.1:443", func() {}); err == nil {
		t.Fatal("redirect target authorized as initial target")
	}
	changed := first
	changed.ExpectedStatus = 201
	if gate.Step(test, []task.TargetTuple{changed}) != nil {
		t.Fatal("tuple/check mismatch accepted")
	}
	if gate.Step(task.TestVersion{ID: "TLS-01", Version: "0.1.0"}, []task.TargetTuple{first}) != nil {
		t.Fatal("cross-test grant")
	}
}

// A task-level validation constraint is not a NET-05 terminal outcome.
// requirement-ref: REQ-NET-05-0061
func TestNet05TaskValidationBeforeStart(t *testing.T) {
	d := &compositeDial{}
	p, clock := compositePlan(t, compositeTCP(), platform.AvailabilityNo, false, d)
	p.Network.RequireValidated = true
	a, _, v, err := (&Engine{clock: clock}).attempt(context.Background(), p, &net05.Runner{}, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
	if err != nil {
		t.Fatal(err)
	}
	if a.Phase() != NotRun || a.Reason() != task.ConstraintsUnsatisfied || v.View().Measurement != nil || v.View().NetworkContext != nil || d.calls.Load() != 0 {
		t.Fatal("pre-start constraint became a composite observation")
	}
}

type compositeSerialDial struct {
	calls, active, overlap atomic.Int32
	entered                chan struct{}
	release                chan struct{}
}

func (d *compositeSerialDial) Dialer(int) (measurement.Dialer, error) { return d, nil }
func (d *compositeSerialDial) DialContext(ctx context.Context, _, _ string) (measurement.Conn, error) {
	if d.active.Add(1) > 1 {
		d.overlap.Add(1)
	}
	defer d.active.Add(-1)
	if d.calls.Add(1) == 1 {
		close(d.entered)
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-d.release:
		return nil, syscall.ECONNREFUSED
	}
}

// One generation-owned runner serializes entire profiles and cancels waiters
// without starting a step, then can be reused after the active profile closes.
// requirement-ref: REQ-NET-05-0037
func TestNet05WholeProfileSerialization(t *testing.T) {
	d := &compositeSerialDial{entered: make(chan struct{}), release: make(chan struct{})}
	p, clock := compositePlan(t, compositeTCP(), platform.AvailabilityYes, false, d)
	runner := &net05.Runner{}
	engine := &Engine{clock: clock}
	type completed struct {
		a   Attempt
		err error
	}
	first := make(chan completed, 1)
	go func() {
		a, _, _, err := engine.attempt(context.Background(), p, runner, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
		first <- completed{a, err}
	}()
	select {
	case <-d.entered:
	case <-time.After(5 * time.Second):
		close(d.release)
		t.Fatal("first profile did not start")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	second := make(chan completed, 1)
	go func() {
		a, _, _, err := engine.attempt(ctx, p, runner, "AgICAgICAgICAgICAgICAg", 1)
		second <- completed{a, err}
	}()
	cancel()
	select {
	case v := <-second:
		if v.err != nil || v.a.Phase() != NotRun || v.a.Reason() != task.CancelledByHost {
			close(d.release)
			t.Fatalf("waiting profile: %#v", v)
		}
	case <-time.After(5 * time.Second):
		close(d.release)
		t.Fatal("waiter ignored cancellation")
	}
	if d.calls.Load() != 1 {
		close(d.release)
		t.Fatal("waiting profile contacted target")
	}
	close(d.release)
	select {
	case v := <-first:
		if v.err != nil || v.a.Phase() != Completed {
			t.Fatalf("first profile: %#v", v)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("first profile remained active")
	}
	a, _, _, err := engine.attempt(context.Background(), p, runner, "AwMDAwMDAwMDAwMDAwMDAw", 1)
	if err != nil || a.Phase() != Completed || d.calls.Load() != 6 || d.overlap.Load() != 0 {
		t.Fatal("profile lock leaked or operations overlapped")
	}
}

type compositeMixedDial struct {
	address, stage string
	interrupt      func()
	calls          atomic.Int32
}

func (d *compositeMixedDial) Dialer(int) (measurement.Dialer, error) { return d, nil }
func (d *compositeMixedDial) DialContext(ctx context.Context, _, _ string) (measurement.Conn, error) {
	n := d.calls.Add(1)
	if d.stage == "tcp" && n == 1 || d.stage == "http-connect" && n == 2 {
		d.interrupt()
		return nil, ctx.Err()
	}
	var nd net.Dialer
	c, err := nd.DialContext(ctx, "tcp", d.address)
	if err != nil {
		return nil, err
	}
	return &httpBoundaryConn{Conn: c, closeHook: func() {
		if d.stage == "cleanup" && n == 2 {
			d.interrupt()
		}
	}}, nil
}

// The mixed DNS/TCP/HTTP profile also interrupts real shared primitives at
// resolution, TCP, TLS, HTTP body read and final cleanup, for every reason.
// requirement-ref: REQ-NET-05-0061
// requirement-ref: REQ-NET-05-0047
func TestNet05MixedPrimitiveTermination(t *testing.T) {
	ca, err := loopback.NewTestCA()
	if err != nil {
		t.Fatal(err)
	}
	cert, err := ca.Issue("8.8.8.8", false)
	if err != nil {
		t.Fatal(err)
	}
	causes := []struct {
		cause  error
		reason task.ExecutionReason
	}{{context.Canceled, task.CancelledByHost}, {errDeadline, task.AttemptDeadlineExceeded}, {errNetworkChanged, task.NetworkChanged}, {errBudget, task.ResourceLimit}, {pathError{task.PlatformSuspended}, task.PlatformSuspended}, {ErrProxyReconfigured, task.ProxyReconfigured}, {errProxyUnavailable, task.ProxyUnavailable}, {errProxyConnect, task.ProxyConnectFailed}}
	steps := []string{compositeStep("dns", "DNS-01", `{"hostname":"example.com","record_type":"A"}`), compositeStep("tcp", "TCP-01", `{"ip_address":"8.8.8.8","port":443}`), compositeStep("http", "HTTP-01", `{"url":"https://8.8.8.8/check","expected_status":200}`)}
	for _, tc := range causes {
		for _, stage := range []string{"dns", "tcp", "http-connect", "tls", "body", "cleanup"} {
			t.Run(string(tc.reason)+"/"+stage, func(t *testing.T) {
				ctx, cancel := context.WithCancelCause(context.Background())
				defer cancel(nil)
				interrupt := func() { cancel(tc.cause); cancel(errNetworkChanged) }
				raw := "HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n"
				if stage == "body" {
					raw = "HTTP/1.1 200 OK\r\nContent-Length: 100\r\n\r\nx"
				}
				srv, err := loopback.ListenHTTP(loopback.HTTPConfig{Certificate: cert, Raw: []byte(raw), Stall: stage == "body"})
				if err != nil {
					t.Fatal(err)
				}
				defer srv.Close()
				address := srv.Address()
				var stalled *loopback.TLSServer
				if stage == "tls" {
					stalled, err = loopback.ListenTLS(loopback.TLSConfig{Mode: loopback.TLSStall})
					if err != nil {
						t.Fatal(err)
					}
					defer stalled.Close()
					address = stalled.Address()
				}
				d := &compositeMixedDial{address: address, stage: stage, interrupt: interrupt}
				p, clock := compositePlan(t, steps, platform.AvailabilityYes, false, d)
				p.Targets = p.Targets.(*Targets).WithResolver(httpBoundaryResolver{stage: stage, interrupt: interrupt})
				done := make(chan struct{})
				if stage == "tls" || stage == "body" {
					go func() {
						defer close(done)
						wait, stop := context.WithTimeout(ctx, 5*time.Second)
						defer stop()
						var err error
						if stage == "tls" {
							err = stalled.Wait(wait, func(c loopback.TLSCounts) bool { return c.In > 0 })
						} else {
							err = srv.Wait(wait, func(c loopback.HTTPCounts) bool { return len(c.Requests) > 0 && c.Out >= len(raw) })
						}
						if err != nil {
							t.Error("active stage not reached", err)
						}
						interrupt()
					}()
				} else {
					close(done)
				}
				a, _, v, err := (&Engine{clock: clock}).attempt(ctx, p, &net05.Runner{Roots: ca.Pool()}, "AQEBAQEBAQEBAQEBAQEBAQ", 1)
				<-done
				if err != nil {
					t.Fatal(err)
				}
				m := v.View().Measurement
				if m == nil {
					t.Fatal("partial profile missing")
				}
				if stage == "cleanup" {
					if a.Phase() != Completed || m.Outcome != "completed" || len(m.Steps) != 3 || a.Reason() != "" {
						t.Fatal("final cleanup replaced outcome")
					}
					return
				}
				k := 2
				if stage == "dns" {
					k = 0
				}
				if stage == "tcp" {
					k = 1
				}
				if a.Phase() != Incomplete || a.Reason() != tc.reason || m.Outcome != "incomplete" || len(m.Steps) != k+1 || m.Steps[k].State != "incomplete" || m.Steps[k].Measurement != nil {
					t.Fatalf("phase=%v reason=%s measurement=%#v", a.Phase(), a.Reason(), m)
				}
				for _, s := range m.Steps[:k] {
					if s.State != "completed" || s.Measurement == nil {
						t.Fatal("earlier primitive result lost")
					}
				}
				if clock.Timers() != 0 {
					t.Fatal("leaked timers")
				}
			})
		}
	}
}
