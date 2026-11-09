package barometermobile

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/platform"
)

func runTask(id, test, input, options string) string {
	return fmt.Sprintf(`{"schema_version":"0.1.0","task_id":%q,"test_id":%q,"test_version":"0.1.0","input":%s,"options":%s,"schedule":{"not_after":%q},"execution":{"limits":{}},"safety":{"risk_level":"low","requires_user_consent":false}}`,
		id, test, input, options, time.Now().Add(24*time.Hour).UTC().Format(time.RFC3339))
}

func code(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func TestNewClientRejectsMissingAdaptersAndInvalidTrust(t *testing.T) {
	trust := trustDocument("TCP-01", tcpTarget)
	n := newNatives()
	for name, a := range map[string]*Adapters{
		"nil":        nil,
		"no records": {Clock: n.clock, Network: n.network},
		"no clock":   {Records: n.records, Network: n.network},
		"no network": {Records: n.records, Clock: n.clock},
	} {
		if _, err := newClient(a, trust, options{platform: "android"}); code(err) != "adapters_invalid" {
			t.Errorf("%s: %v", name, err)
		}
	}
	if _, err := newClient(n.adapters(), trust, options{}); code(err) != "adapters_invalid" {
		t.Error("an unsupported platform is accepted:", err)
	}
	if _, err := NewClient(n.adapters(), trust); code(err) != "adapters_invalid" {
		t.Error("the host platform is accepted by NewClient:", err)
	}
	for _, bad := range [][]byte{nil, []byte(`{}`), []byte(strings.Replace(string(trust), "barometer-trust/1", "barometer-trust/2", 1))} {
		if _, err := newClient(n.adapters(), bad, options{platform: "ios"}); code(err) != "trust_invalid" {
			t.Error("invalid trust document accepted:", err)
		}
	}
	// Optional adapters may be absent.
	if _, err := newClient(&Adapters{Records: n.records, Clock: n.clock, Network: n.network}, trust, options{platform: "ios"}); err != nil {
		t.Fatal(err)
	}
	var c *Client
	if c.Initialize(`{}`) == nil || c.Close() != nil {
		t.Fatal("nil client")
	}
}

func TestConstructionPerformsNoIO(t *testing.T) {
	n := newNatives()
	n.records.fail.Store(true)
	n.clock.fail.Store(true)
	n.network.fail.Store(true)
	newTestClient(t, n, trustDocument("TCP-01", tcpTarget), options{})
	if n.records.count() != 0 || n.records.puts.Load() != 0 {
		t.Fatal("construction wrote records")
	}
}

func TestInitializeAndLifecycleThroughTheFacade(t *testing.T) {
	n := newNatives()
	c := newTestClient(t, n, trustDocument("TCP-01", tcpTarget), options{})
	if err := c.Initialize(`{}`); err != nil {
		t.Fatal(code(err))
	}
	if err := c.Initialize(`{}`); code(err) != "already_initialized" {
		t.Fatal(code(err))
	}
	if b, err := c.NextBundle(); err != nil || b != "" {
		t.Fatal("unexpected bundle", b, err)
	}
	if _, err := c.AdmitTask(runTask("t1", "TCP-01", `{"ip_address":"8.8.8.8","port":443}`, `{}`), "0123"); err == nil {
		t.Fatal("an unknown sourceRef was admitted")
	}
	if err := c.ClearMeasurementData(); err != nil {
		t.Fatal(code(err))
	}
	if n.keys.deletes.Load() != 1 {
		t.Fatal("clearing did not delete the linkage key")
	}
	if err := c.Initialize(`{}`); err != nil {
		t.Fatal(code(err))
	}
	if err := c.Close(); err != nil {
		t.Fatal(code(err))
	}
	if _, err := c.InstallEmbeddedTask("embedded-issuer", []byte(`{}`)); code(err) != "client_closed" {
		t.Fatal("install after close:", code(err))
	}
	if err := c.Initialize(`{}`); code(err) != "client_closed" {
		t.Fatal(code(err))
	}
}

func TestInitializeFailsClosedOnStalledStorage(t *testing.T) {
	n := newNatives()
	c := newTestClient(t, n, trustDocument("TCP-01", tcpTarget), options{})
	n.records.gate.stall.Store(true)
	start := time.Now()
	if err := c.Initialize(`{}`); code(err) != "not_initialized" {
		t.Fatal(code(err))
	}
	if took := time.Since(start); took > 2*defaultCallback {
		t.Fatal("initialize waited past the lifecycle deadline:", took)
	}
	n.records.gate.open()
	n.records.gate.stall.Store(false)
	if err := c.Initialize(`{}`); err != nil {
		t.Fatal("initialize did not recover:", code(err))
	}
}

func TestInstallErrorsUseFixedCodes(t *testing.T) {
	n := newNatives()
	c := newTestClient(t, n, trustDocument("TCP-01", tcpTarget), options{})
	if _, err := c.InstallEmbeddedTask("not a valid issuer!", []byte(`{}`)); code(err) != "delivery_rejected" {
		t.Fatal(code(err))
	}
	if _, err := c.InstallRemoteTask(nil); code(err) != "delivery_rejected" {
		t.Fatal(code(err))
	}
	n.clock.fail.Store(true)
	if _, err := c.InstallEmbeddedTask("embedded-issuer", []byte(`{}`)); code(err) != "delivery_unavailable" {
		t.Fatal(code(err))
	}
	n.clock.fail.Store(false)
	for i := range deliveryCapacity {
		if _, err := c.InstallEmbeddedTask("embedded-issuer", []byte(fmt.Sprint(i))); err != nil {
			t.Fatal(code(err))
		}
	}
	if _, err := c.InstallEmbeddedTask("embedded-issuer", []byte(`{}`)); code(err) != "capacity_exceeded" {
		t.Fatal(code(err))
	}
}

// waitTerminal polls RunState; the bound is a liveness guard.
func waitTerminal(t *testing.T, c *Client, rh string) {
	t.Helper()
	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); time.Sleep(5 * time.Millisecond) {
		s, err := c.RunState(rh)
		if err != nil {
			t.Fatal("run state:", code(err))
		}
		if s == "terminal" {
			return
		}
	}
	t.Fatal("run did not become terminal")
}

type measured struct {
	Execution   struct{ State, Reason string }
	Measurement *struct {
		Outcome      string
		Observations []struct {
			Type, Status string
			Details      map[string]any
		}
	}
	Implementation struct {
		LibraryVersion string `json:"library_version"`
		Platform       string
	}
}

// runOne installs, admits and runs taskJSON and returns its retained result
// from the bundle the facade returns, then acknowledges the bundle.
func runOne(t *testing.T, c *Client, taskJSON string) (measured, string) {
	t.Helper()
	ref, err := c.InstallEmbeddedTask("embedded-issuer", []byte(taskJSON))
	if err != nil {
		t.Fatal("install:", code(err))
	}
	th, err := c.AdmitTask(taskJSON, ref)
	if err != nil {
		t.Fatal("admit:", code(err))
	}
	rh, err := c.StartRun(th)
	if err != nil {
		t.Fatal("start:", code(err))
	}
	waitTerminal(t, c, rh)
	raw, err := c.NextBundle()
	if err != nil || raw == "" {
		t.Fatal("no bundle:", code(err))
	}
	var b struct {
		BundleID string            `json:"bundle_id"`
		Results  []json.RawMessage `json:"results"`
	}
	if err := json.Unmarshal([]byte(raw), &b); err != nil || len(b.Results) != 1 {
		t.Fatalf("bundle: %s", raw)
	}
	var r measured
	if err := json.Unmarshal(b.Results[0], &r); err != nil {
		t.Fatal(err)
	}
	if err := c.AcknowledgeBundle(b.BundleID); err != nil {
		t.Fatal("acknowledge:", code(err))
	}
	return r, string(b.Results[0])
}

// loopbackDialer maps the checked public literal to a loopback listener. It
// never contacts the public destination and fails on any other destination.
type loopbackDialer struct{ addr, want string }

func (d loopbackDialer) Dialer(int) (measurement.Dialer, error) { return d, nil }
func (d loopbackDialer) DialContext(ctx context.Context, network, address string) (measurement.Conn, error) {
	if network != "tcp" || address != d.want {
		return nil, errors.New("synthetic: unexpected destination")
	}
	var nd net.Dialer
	return nd.DialContext(ctx, "tcp", d.addr)
}

// TCP-01 runs through the facade on the host with the production registry,
// codec and runner. Only the direct dialer is replaced, through the test-only
// option, to reach a loopback listener.
func TestTCP01RunsThroughTheFacade(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	accepted := make(chan struct{}, 1)
	go func() {
		if conn, err := l.Accept(); err == nil {
			conn.Close()
			accepted <- struct{}{}
		}
	}()
	n := newNatives()
	c := newTestClient(t, n, trustDocument("TCP-01", tcpTarget), options{directDialers: loopbackDialer{addr: l.Addr().String(), want: "8.8.8.8:443"}})
	if err := c.Initialize(`{}`); err != nil {
		t.Fatal(code(err))
	}
	r, raw := runOne(t, c, runTask("tcp-1", "TCP-01", `{"ip_address":"8.8.8.8","port":443}`, `{}`))
	if r.Execution.State != "completed" || r.Measurement == nil || r.Measurement.Outcome != "connected" || len(r.Measurement.Observations) != 1 || r.Measurement.Observations[0].Type != "tcp_connect" {
		t.Fatalf("result: %s", raw)
	}
	if r.Implementation.Platform != "android" || r.Implementation.LibraryVersion != libraryVersion {
		t.Fatalf("implementation identity: %s", raw)
	}
	select {
	case <-accepted:
	case <-time.After(5 * time.Second):
		t.Fatal("no loopback connection")
	}
	if strings.Contains(raw, "8.8.8.8") || strings.Contains(raw, "127.0.0.1") {
		t.Fatalf("retained a target address: %s", raw)
	}
	if n.records.count() == 0 {
		t.Fatal("replay and run state did not reach the native record store")
	}
}

// DNS-01 runs through the facade with the system resolver mode, which uses the
// native SystemResolver adapter and opens no socket.
func TestDNS01RunsThroughTheFacade(t *testing.T) {
	for _, tc := range []struct {
		answer  []string
		err     error
		outcome string
	}{{[]string{"93.184.215.14"}, nil, "resolved"}, {nil, nil, "no_data"}, {nil, errSynthetic, "resolver_error"}} {
		t.Run(tc.outcome, func(t *testing.T) {
			n := newNatives()
			n.resolver.answer, n.resolver.err = tc.answer, tc.err
			c := newTestClient(t, n, trustDocument("DNS-01", target{ID: "reg-a", Test: "DNS-01", Role: "measurement", Tuple: map[string]any{"hostname": "example.org"}}),
				options{platform: "ios", directDialers: loopbackDialer{want: "none"}})
			if err := c.Initialize(`{}`); err != nil {
				t.Fatal(code(err))
			}
			r, raw := runOne(t, c, runTask("dns-1", "DNS-01", `{"hostname":"example.org","record_type":"A"}`, `{}`))
			if r.Execution.State != "completed" || r.Measurement == nil || r.Measurement.Outcome != tc.outcome || r.Measurement.Observations[0].Type != "dns_query" {
				t.Fatalf("result: %s", raw)
			}
			if r.Implementation.Platform != "ios" {
				t.Fatalf("implementation identity: %s", raw)
			}
			if len(n.resolver.queries) != 1 || n.resolver.queries[0] != "ip4 example.org." {
				t.Fatal("unexpected resolver queries:", n.resolver.queries)
			}
			for _, canary := range []string{"example.org", "93.184.215.14", "SYNTHETIC"} {
				if strings.Contains(raw, canary) {
					t.Fatalf("retained %s: %s", canary, raw)
				}
			}
		})
	}
}

// A run whose consent callback stalls does not start an attempt: the stalled
// answer is unavailable, which never satisfies the requirement.
func TestStalledConsentNeverPermitsAnAttempt(t *testing.T) {
	n := newNatives()
	n.consent.gate.stall.Store(true)
	c := newTestClient(t, n, trustDocument("TCP-01", tcpTarget), options{directDialers: loopbackDialer{want: "none"}})
	if err := c.Initialize(`{}`); err != nil {
		t.Fatal(code(err))
	}
	data := strings.Replace(runTask("consent-1", "TCP-01", `{"ip_address":"8.8.8.8","port":443}`, `{}`), `"requires_user_consent":false`, `"requires_user_consent":true`, 1)
	ref, err := c.InstallEmbeddedTask("embedded-issuer", []byte(data))
	if err != nil {
		t.Fatal(code(err))
	}
	th, err := c.AdmitTask(data, ref)
	if err != nil {
		t.Fatal(code(err))
	}
	start := time.Now()
	if _, err := c.StartRun(th); code(err) != "run_not_due" {
		t.Fatal("a run started without an available consent answer:", code(err))
	}
	if took := time.Since(start); took > defaultCallback+slack {
		t.Fatal("start waited past its deadline:", took)
	}
	if b, err := c.NextBundle(); err != nil || b != "" {
		t.Fatal("a result was produced:", b)
	}
}

// The suspension value maps to the core port: "suspended" and "active" are
// known, anything else is unknown and does not block execution. A run that
// comes due while native code reports suspension is deferred with
// run_not_due and reserves nothing; after resume, which advances the
// generation, the same task starts.
func TestAccessStateSuspension(t *testing.T) {
	n := newNatives()
	ctx := context.Background()
	for v, want := range map[string]platform.Availability{"suspended": platform.AvailabilityYes, "active": platform.AvailabilityNo,
		"unknown": platform.AvailabilityUnknown, "": platform.AvailabilityUnknown, "SUSPENDED": platform.AvailabilityUnknown, "yes": platform.AvailabilityUnknown} {
		n.network.mu.Lock()
		n.network.state = NewAccessState(1, "wifi", "not_detected", "yes", "no", "yes", "no", v)
		n.network.mu.Unlock()
		s, err := network{n.network}.Access(ctx)
		if err != nil || s.Suspended != want {
			t.Errorf("suspension %q: %v %v", v, s.Suspended, err)
		}
	}

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	go func() {
		if conn, err := l.Accept(); err == nil {
			conn.Close()
		}
	}()
	c := newTestClient(t, n, trustDocument("TCP-01", tcpTarget), options{directDialers: loopbackDialer{addr: l.Addr().String(), want: "8.8.8.8:443"}})
	if err := c.Initialize(`{}`); err != nil {
		t.Fatal(code(err))
	}
	taskJSON := runTask("tcp-suspended", "TCP-01", `{"ip_address":"8.8.8.8","port":443}`, `{}`)
	ref, err := c.InstallEmbeddedTask("embedded-issuer", []byte(taskJSON))
	if err != nil {
		t.Fatal("install:", code(err))
	}
	th, err := c.AdmitTask(taskJSON, ref)
	if err != nil {
		t.Fatal("admit:", code(err))
	}
	n.network.mu.Lock()
	n.network.state = NewAccessState(2, "wifi", "not_detected", "yes", "no", "yes", "no", "suspended")
	n.network.mu.Unlock()
	if rh, err := c.StartRun(th); err == nil || code(err) != "run_not_due" || rh != "" {
		t.Fatalf("a run started while suspended: %q %s", rh, code(err))
	}
	if raw, err := c.NextBundle(); err == nil && raw != "" {
		t.Fatalf("a suspended start wrote a result: %s", raw)
	}
	n.network.mu.Lock()
	n.network.state = NewAccessState(3, "wifi", "not_detected", "yes", "no", "yes", "no", "active")
	n.network.mu.Unlock()
	rh, err := c.StartRun(th)
	if err != nil {
		t.Fatal("no run after resume:", code(err))
	}
	waitTerminal(t, c, rh)
}
