package barometer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/registry"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit/capability"
)

var net01Test = task.TestVersion{ID: "NET-01", Version: "0.1.0"}

// net01Network is the fake platform port. Read 1 of a run is the reservation
// check at StartRun; script may replace any read, for example to fail or
// block the attempt's own read. It never carries the capability's withheld
// canaries, because the port has no field for them.
type net01Network struct {
	*capability.Network
	mu     sync.Mutex
	reads  int
	script func(ctx context.Context, read int) (platform.AccessState, error, bool)
}

func (n *net01Network) Access(ctx context.Context) (platform.AccessState, error) {
	n.mu.Lock()
	n.reads++
	read, script := n.reads, n.script
	n.mu.Unlock()
	if script != nil {
		if s, err, ok := script(ctx, read); ok {
			return s, err
		}
	}
	return n.Network.Access(ctx)
}

func (n *net01Network) Reads() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.reads
}

// net01Dialers is the recording transport. NET-01 must never ask for a path
// dialer, dial, resolve or reach an egress endpoint, so every use is counted.
type net01Dialers struct{ calls atomic.Int32 }

func (d *net01Dialers) Dialer(int) (measurement.Dialer, error) { d.calls.Add(1); return d, nil }
func (d *net01Dialers) DialContext(context.Context, string, string) (measurement.Conn, error) {
	d.calls.Add(1)
	return nil, errors.New("synthetic: NET-01 opened a connection")
}

// net01Harness runs NET-01 with the production codec, rules, registry, runner
// and direct path. Only the platform network port and transport are fakes.
func net01Harness(t *testing.T) (*harness, *net01Network, *net01Dialers) {
	t.Helper()
	h := runningHarness(t)
	h.adapters.registry = nil
	var err error
	// No target registration exists: NET-01 has no remote target to match.
	in := synPolicyInput(2, "reg-a", func(p *task.Permissions) { p.Tests = []task.TestVersion{net01Test}; p.RegistrationIDs = nil })
	in.Issuers[0].Permissions.RegistrationIDs, in.Issuers[1].Permissions.RegistrationIDs, in.Targets = nil, nil, nil
	h.adapters.policy, err = task.NewLocalPolicy(in)
	if err != nil {
		t.Fatal(err)
	}
	n := &net01Network{Network: h.run.network}
	h.adapters.run.paths, h.adapters.run.network = nil, n
	d := &net01Dialers{}
	h.adapters.run.directDialers = d
	return h, n, d
}

func net01Task(id string) string {
	return strings.Replace(runTask(id, "", "", false), `"test_id":"TCP-01"`, `"test_id":"NET-01"`, 1)
}

// net01Result checks the common NET-01 shape: never a top-level
// network_context or egress, and, for an outcome, exactly outcome, an empty
// observations array and access only where wanted. It returns the access.
func net01Result(t *testing.T, raw []byte, state, reason, outcome string, withAccess bool) map[string]any {
	t.Helper()
	var r map[string]any
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatal(err)
	}
	if _, ok := r["network_context"]; ok || strings.Contains(string(raw), `"egress"`) {
		t.Fatalf("context in NET-01 result: %s", raw)
	}
	for _, canary := range capability.Canaries {
		if strings.Contains(string(raw), canary) {
			t.Fatalf("canary %q in result", canary)
		}
	}
	e := r["execution"].(map[string]any)
	if e["state"] != state || (reason == "") != (e["reason"] == nil) || reason != "" && e["reason"] != reason {
		t.Fatalf("execution: %s", raw)
	}
	m, ok := r["measurement"].(map[string]any)
	if outcome == "" {
		if ok {
			t.Fatalf("interrupted NET-01 has a measurement: %s", raw)
		}
		return nil
	}
	want := []string{"observations", "outcome"}
	if withAccess {
		want = append(want, "access")
	}
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	slices.Sort(want)
	obs, isArray := m["observations"].([]any)
	if !ok || m["outcome"] != outcome || !slices.Equal(keys, want) || !isArray || len(obs) != 0 {
		t.Fatalf("measurement: %s", raw)
	}
	if !withAccess {
		return nil
	}
	a := m["access"].(map[string]any)
	if len(a) != 6 || a["network_changed"] != false {
		t.Fatalf("access: %s", raw)
	}
	return a
}

// requirement-ref: REQ-NET-01-0005
// requirement-ref: REQ-NET-01-0010
// requirement-ref: REQ-NET-01-0011
// requirement-ref: REQ-NET-01-0012
// requirement-ref: REQ-NET-01-0013
// requirement-ref: REQ-NET-01-0014
func TestMNet01Admission(t *testing.T) {
	t.Run("empty input without a target registration", func(t *testing.T) {
		h, n, d := net01Harness(t)
		c := h.initialized(t)
		defer c.Close()
		data := net01Task("valid")
		if _, err := c.AdmitTask(data, h.embedded(t, data)); err != nil {
			t.Fatal(codeOf(err))
		}
		if n.Reads() != 0 || d.calls.Load() != 0 || len(h.run.output.Records()) != 0 {
			t.Fatal("admission read platform state, used the transport or wrote a result")
		}
	})
	inputs := []string{`null`, `[]`, `[{}]`, `""`, `"{}"`, `0`, `1.5`, `true`, `false`,
		`{"interface":"wlan0"}`, `{"interface_name":"canary-if-wlan7"}`, `{"interface_index":3}`, `{"ip_address":"198.51.100.77"}`, `{"local_address":"2001:db8::77"}`,
		`{"device_id":"canary-device"}`, `{"ssid":"canary-ssid-7f3a"}`, `{"bssid":"02:00:5e:10:00:77"}`, `{"imsi":"canary-imsi-250017777"}`, `{"hostname":"example.org"}`,
		`{"":null}`, `{"x":1}`}
	for i, input := range append(inputs, "missing") {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			h, n, d := net01Harness(t)
			c := h.initialized(t)
			defer c.Close()
			data := strings.Replace(net01Task("invalid"), `"input":{}`, `"input":`+input, 1)
			if input == "missing" {
				data = strings.Replace(net01Task("invalid"), `"input":{},`, ``, 1)
			}
			handle, err := c.AdmitTask(data, h.embedded(t, data))
			if handle != "" || codeOf(err) != "task_schema_invalid" {
				t.Fatalf("%s: %s", input, codeOf(err))
			}
			if taskHandles(c) != 0 || h.replay.Commits() != 0 || len(c.rt.runs.entries) != 0 || n.Reads() != 0 || d.calls.Load() != 0 || len(h.run.output.Records()) != 0 {
				t.Fatal("non-atomic rejection")
			}
		})
	}
	t.Run("proxy path", func(t *testing.T) {
		h, _, _ := net01Harness(t)
		c := h.initialized(t)
		defer c.Close()
		data := strings.Replace(net01Task("proxy"), `"execution":{"limits":{}`, `"execution":{"network_path":"proxy","limits":{}`, 1)
		if handle, err := c.AdmitTask(data, h.embedded(t, data)); handle != "" || codeOf(err) != "task_schema_invalid" {
			t.Fatal(codeOf(err))
		}
	})
	t.Run("not a NET-05 step", func(t *testing.T) {
		reg, err := productionRegistry(nil)
		if err != nil {
			t.Fatal(err)
		}
		e, ok := reg.Lookup(net01Test)
		_, noTargets := e.Targets.(registry.NoTargets)
		_, noProxy := e.Proxy.(registry.NoProxy)
		if !ok || e.Step || e.Composite || e.BaselineRisk != task.LowRisk || !noTargets || !noProxy {
			t.Fatal("NET-01 registry entry")
		}
	})
}

// requirement-ref: REQ-NET-01-0015
// requirement-ref: REQ-NET-01-0016
func TestMNet01Options(t *testing.T) {
	for i, options := range []string{`null`, `[]`, `[{}]`, `""`, `"{}"`, `0`, `1.5`, `true`, `false`, `{"x":1}`, `{"resolver_mode":"system"}`, `{"egress":true}`, `{"interface":"wlan0"}`, "missing"} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			h, n, d := net01Harness(t)
			c := h.initialized(t)
			defer c.Close()
			data := strings.Replace(net01Task("options"), `"options":{}`, `"options":`+options, 1)
			if options == "missing" {
				data = strings.Replace(net01Task("options"), `"options":{},`, ``, 1)
			}
			if handle, err := c.AdmitTask(data, h.embedded(t, data)); handle != "" || codeOf(err) != "task_schema_invalid" {
				t.Fatalf("%s: %s", options, codeOf(err))
			}
			if h.replay.Commits() != 0 || n.Reads() != 0 || d.calls.Load() != 0 {
				t.Fatal("non-atomic rejection")
			}
		})
	}
	// The empty object selects the one snapshot behavior.
	h, _, _ := net01Harness(t)
	c := h.initialized(t)
	defer c.Close()
	net01Result(t, tcpRaw(t, h, c, net01Task("empty-options")), "completed", "", "snapshot", true)
}

// NET-01 runs through Client with the production registry: one platform read
// starts it, no path dialer, connection or egress endpoint is used, and the
// result carries the access object directly with no top-level context.
// requirement-ref: REQ-NET-01-0006
// requirement-ref: REQ-NET-01-0007
// requirement-ref: REQ-NET-01-0008
// requirement-ref: REQ-NET-01-0021
// requirement-ref: REQ-NET-01-0023
// requirement-ref: REQ-NET-01-0024
// requirement-ref: REQ-NET-01-0025
// requirement-ref: REQ-NET-01-0026
func TestMNet01ExecutionAndOutput(t *testing.T) {
	h, n, d := net01Harness(t)
	c := h.initialized(t)
	defer c.Close()
	a := net01Result(t, tcpRaw(t, h, c, net01Task("snapshot")), "completed", "", "snapshot", true)
	want := map[string]any{"network_type": "wifi", "vpn_state": "unknown", "ipv4_available": "yes", "ipv6_available": "unknown", "validated": "unknown", "network_changed": false}
	for k, v := range want {
		if a[k] != v {
			t.Fatalf("access %s: %v", k, a[k])
		}
	}
	if n.Reads() != 2 || n.Watching() != 0 || d.calls.Load() != 0 {
		t.Fatalf("reads %d, watches %d, transport %d", n.Reads(), n.Watching(), d.calls.Load())
	}
	t.Run("egress requirement", func(t *testing.T) {
		h, n, d := net01Harness(t)
		c := h.initialized(t)
		defer c.Close()
		data := strings.Replace(net01Task("context-requirement"), `"execution":{"limits":{}`, `"execution":{"limits":{},"network":{"context":{"required_fields":["country_code"]}}`, 1)
		net01Result(t, tcpRaw(t, h, c, data), "not_run", "network_context_unavailable", "", false)
		if n.Reads() != 1 || d.calls.Load() != 0 {
			t.Fatal("egress requirement started NET-01 or looked up egress")
		}
	})
	t.Run("access condition", func(t *testing.T) {
		h, n, d := net01Harness(t)
		c := h.initialized(t)
		defer c.Close()
		data := strings.Replace(net01Task("condition"), `"execution":{"limits":{}`, `"execution":{"limits":{},"network":{"allowed_transports":["wifi"]}`, 1)
		net01Result(t, tcpRaw(t, h, c, data), "completed", "", "snapshot", true)
		n.Set(platform.AccessState{Generation: 1, Type: platform.NetworkCellular})
		n.script = func(_ context.Context, read int) (platform.AccessState, error, bool) {
			// Satisfied at reservation, unsatisfied at the attempt's gate.
			return defaultAccess, nil, read == 4
		}
		net01Result(t, tcpRaw(t, h, c, strings.Replace(data, `"condition"`, `"condition-2"`, 1)), "not_run", "constraints_unsatisfied", "", false)
		if d.calls.Load() != 0 || n.Watching() != 0 {
			t.Fatal("transport or watch")
		}
	})
}

// Every access value maps to its registered name, including unrecognized
// platform values as unknown, and the failure classes stay distinct. Each
// snapshot is one platform read, so it cannot combine generations.
// requirement-ref: REQ-NET-01-0006
// requirement-ref: REQ-NET-01-0009
// requirement-ref: REQ-NET-01-0026
func TestMNet01Normalization(t *testing.T) {
	for i, tc := range []struct {
		state platform.AccessState
		want  [5]string
	}{
		{platform.AccessState{Generation: 1, Type: platform.NetworkWiFi, VPN: platform.VPNNotDetected, IPv4: platform.AvailabilityYes, IPv6: platform.AvailabilityYes, Validated: platform.AvailabilityYes}, [5]string{"wifi", "not_detected", "yes", "yes", "yes"}},
		{platform.AccessState{Generation: 1, Type: platform.NetworkCellular, VPN: platform.VPNDetected, IPv4: platform.AvailabilityNo, IPv6: platform.AvailabilityYes, Validated: platform.AvailabilityNo}, [5]string{"cellular", "detected", "no", "yes", "no"}},
		{platform.AccessState{Generation: 1, Type: platform.NetworkEthernet}, [5]string{"ethernet", "unknown", "unknown", "unknown", "unknown"}},
		{platform.AccessState{Generation: 1, Type: platform.NetworkOther, IPv4: platform.AvailabilityNo, IPv6: platform.AvailabilityNo, Validated: platform.AvailabilityYes}, [5]string{"other", "unknown", "no", "no", "yes"}},
		{platform.AccessState{Generation: 1}, [5]string{"unknown", "unknown", "unknown", "unknown", "unknown"}},
		{platform.AccessState{Generation: 1, Type: 99, VPN: 99, IPv4: 99, IPv6: 99, Validated: 99}, [5]string{"unknown", "unknown", "unknown", "unknown", "unknown"}},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			h, n, _ := net01Harness(t)
			n.Set(tc.state)
			c := h.initialized(t)
			defer c.Close()
			a := net01Result(t, tcpRaw(t, h, c, net01Task("normalize")), "completed", "", "snapshot", true)
			got := [5]string{a["network_type"].(string), a["vpn_state"].(string), a["ipv4_available"].(string), a["ipv6_available"].(string), a["validated"].(string)}
			if got != tc.want {
				t.Fatalf("got %v", got)
			}
		})
	}
	for name, tc := range map[string]struct {
		state   platform.AccessState
		err     error
		outcome string
	}{
		"no network-state API": {err: platform.ErrNetworkUnavailable, outcome: "platform_unavailable"},
		"wrapped no API":       {err: fmt.Errorf("binding: %w", platform.ErrNetworkUnavailable), outcome: "platform_unavailable"},
		"API failure":          {err: errors.New("canary raw platform error 198.51.100.77"), outcome: "platform_error"},
		"no generation":        {state: platform.AccessState{Type: platform.NetworkWiFi}, outcome: "platform_error"},
	} {
		t.Run(name, func(t *testing.T) {
			h, n, d := net01Harness(t)
			n.script = func(_ context.Context, read int) (platform.AccessState, error, bool) {
				return tc.state, tc.err, read == 2
			}
			c := h.initialized(t)
			defer c.Close()
			raw := tcpRaw(t, h, c, net01Task("failure"))
			net01Result(t, raw, "completed", "", tc.outcome, false)
			if strings.Contains(string(raw), "canary raw") || strings.Contains(string(raw), "inconclusive") || d.calls.Load() != 0 || n.Watching() != 0 {
				t.Fatalf("raw error, inconclusive, transport or watch: %s", raw)
			}
		})
	}
	t.Run("unavailable at reservation", func(t *testing.T) {
		h, n, _ := net01Harness(t)
		n.script = func(context.Context, int) (platform.AccessState, error, bool) {
			return platform.AccessState{}, platform.ErrNetworkUnavailable, true
		}
		c := h.initialized(t)
		defer c.Close()
		if _, err := c.StartRun(h.admit(t, c, net01Task("reserve"))); err == nil || len(h.run.output.Records()) != 0 {
			t.Fatal("a run started without platform state at reservation")
		}
	})
}

// The outer deadline releases a read the platform does not answer, with zero
// traffic. Concurrent snapshots across a generation transition each keep one
// coherent generation or end incomplete; none merges states.
// requirement-ref: REQ-NET-01-0021
func TestMNet01Limits(t *testing.T) {
	t.Run("deadline", func(t *testing.T) {
		h, n, d := net01Harness(t)
		entered := make(chan struct{})
		n.script = func(ctx context.Context, read int) (platform.AccessState, error, bool) {
			if read != 2 {
				return platform.AccessState{}, nil, false
			}
			close(entered)
			<-ctx.Done()
			return platform.AccessState{}, ctx.Err(), true
		}
		c := h.initialized(t)
		defer c.Close()
		rh, err := c.StartRun(h.admit(t, c, net01Task("deadline")))
		if err != nil {
			t.Fatal(codeOf(err))
		}
		<-entered
		h.advance(30 * time.Second)
		waitTerminal(t, c, rh)
		net01Result(t, h.run.output.Records()[0], "incomplete", "attempt_deadline_exceeded", "", false)
		if d.calls.Load() != 0 || n.Watching() != 0 {
			t.Fatal("transport or watch")
		}
	})
	t.Run("no retry", func(t *testing.T) {
		// A failed platform read is one completed outcome: the run reads once
		// after the reservation check and never retries or opens a path.
		h, n, d := net01Harness(t)
		n.script = func(_ context.Context, read int) (platform.AccessState, error, bool) {
			return platform.AccessState{}, errors.New("synthetic platform failure"), read >= 2
		}
		c := h.initialized(t)
		defer c.Close()
		net01Result(t, tcpRaw(t, h, c, net01Task("no-retry")), "completed", "", "platform_error", false)
		if n.Reads() != 2 || d.calls.Load() != 0 || n.Watching() != 0 || len(h.run.output.Records()) != 1 {
			t.Fatalf("reads %d, transport %d, results %d", n.Reads(), d.calls.Load(), len(h.run.output.Records()))
		}
	})
	t.Run("concurrent transition", func(t *testing.T) {
		h, n, d := net01Harness(t)
		first := platform.AccessState{Generation: 1, Type: platform.NetworkWiFi, VPN: platform.VPNNotDetected, IPv4: platform.AvailabilityYes, IPv6: platform.AvailabilityNo, Validated: platform.AvailabilityYes}
		second := platform.AccessState{Generation: 2, Type: platform.NetworkCellular, VPN: platform.VPNDetected, IPv4: platform.AvailabilityNo, IPv6: platform.AvailabilityYes, Validated: platform.AvailabilityNo}
		n.Set(first)
		c := h.initialized(t)
		defer c.Close()
		var handles []string
		for i := range 4 {
			rh, err := c.StartRun(h.admit(t, c, net01Task(fmt.Sprint("concurrent-", i))))
			if err != nil {
				t.Fatal(codeOf(err))
			}
			handles = append(handles, rh)
			if i == 1 {
				n.Set(second)
			}
		}
		for _, rh := range handles {
			waitTerminal(t, c, rh)
		}
		for _, raw := range h.run.output.Records() {
			var r struct {
				Execution struct{ State, Reason string }
			}
			_ = json.Unmarshal(raw, &r)
			if r.Execution.State == "incomplete" {
				net01Result(t, raw, "incomplete", "network_changed", "", false)
				continue
			}
			a := net01Result(t, raw, "completed", "", "snapshot", true)
			wifi := a["network_type"] == "wifi" && a["vpn_state"] == "not_detected" && a["ipv4_available"] == "yes" && a["ipv6_available"] == "no" && a["validated"] == "yes"
			cell := a["network_type"] == "cellular" && a["vpn_state"] == "detected" && a["ipv4_available"] == "no" && a["ipv6_available"] == "yes" && a["validated"] == "no"
			if !wifi && !cell {
				t.Fatalf("merged snapshot: %s", raw)
			}
		}
		if d.calls.Load() != 0 || n.Watching() != 0 {
			t.Fatal("transport or watch")
		}
	})
}

// Client-level interruptions while the attempt's read is outstanding. The
// engine suite TestNet01AllTerminationBoundaries covers every Procedure point
// with every cause, including injected suspension.
// requirement-ref: REQ-NET-01-0027
func TestMNet01Termination(t *testing.T) {
	for _, reason := range []string{"network_changed", "cancelled_by_host", "attempt_deadline_exceeded"} {
		t.Run(reason, func(t *testing.T) {
			h, n, d := net01Harness(t)
			entered, release := make(chan struct{}), make(chan struct{})
			n.script = func(ctx context.Context, read int) (platform.AccessState, error, bool) {
				if read != 2 {
					return platform.AccessState{}, nil, false
				}
				close(entered)
				select {
				case <-ctx.Done():
					return platform.AccessState{}, ctx.Err(), true
				case <-release:
					return defaultAccess, nil, true
				}
			}
			c := h.initialized(t)
			defer c.Close()
			rh, err := c.StartRun(h.admit(t, c, net01Task("terminate")))
			if err != nil {
				t.Fatal(codeOf(err))
			}
			<-entered
			switch reason {
			case "network_changed":
				a := defaultAccess
				a.Generation++
				n.Set(a)
				close(release)
			case "cancelled_by_host":
				if err := c.CancelRun(rh); err != nil {
					t.Fatal(codeOf(err))
				}
			default:
				h.advance(30 * time.Second)
			}
			waitTerminal(t, c, rh)
			net01Result(t, h.run.output.Records()[0], "incomplete", reason, "", false)
			if d.calls.Load() != 0 || n.Watching() != 0 {
				t.Fatal("transport or watch")
			}
		})
	}
}

// requirement-ref: REQ-NET-01-0022
// requirement-ref: REQ-NET-01-0025
func TestMNet01Versions(t *testing.T) {
	for _, field := range []string{"test_version", "schema_version"} {
		for _, version := range []string{`"0.1.1"`, `"0.2.0"`, `"1.0.0"`, `""`, `"v0.1.0"`, `"0.1"`, `"0.1.0-draft"`, `0.1`, `null`, "missing"} {
			h, n, d := net01Harness(t)
			c := h.initialized(t)
			data := strings.Replace(net01Task("version"), `"`+field+`":"0.1.0"`, `"`+field+`":`+version, 1)
			if version == "missing" {
				data = strings.Replace(net01Task("version"), `"`+field+`":"0.1.0",`, ``, 1)
			}
			handle, err := c.AdmitTask(data, h.embedded(t, data))
			if code := codeOf(err); code != "task_schema_invalid" && code != "task_unsupported" || handle != "" || h.replay.Commits() != 0 || n.Reads() != 0 || d.calls.Load() != 0 || len(h.run.output.Records()) != 0 {
				t.Fatal("unsupported version admitted", field, version, codeOf(err))
			}
			c.Close()
		}
	}
	h, _, _ := net01Harness(t)
	c := h.initialized(t)
	defer c.Close()
	raw := tcpRaw(t, h, c, net01Task("versions"))
	var v struct {
		SchemaVersion string                                      `json:"schema_version"`
		Test          struct{ ID, Version, ResultVersion string } `json:"test"`
	}
	if json.Unmarshal(raw, &v) != nil || v.SchemaVersion != "0.1.0" || v.Test.ID != "NET-01" || v.Test.Version != "0.1.0" {
		t.Fatalf("emitted combination: %s", raw)
	}
	cd, _ := observation.NewContextDecoder(result.Contract, observation.ContextApprovals{})
	path, _ := observation.NewPath(1, 0, "direct", "")
	decoder, err := result.New(result.Binding{TaskID: "versions", Output: result.Output{Test: model.Test{ID: "NET-01", Version: "0.1.0", ResultVersion: "0.1.0"}}, MaxAttempts: 1, MaxDurationMS: 60000, Implementation: implementation, ContextDecoder: cd, Path: path, AccessGeneration: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decoder.Decode(raw); err != nil {
		t.Fatal("valid result rejected", err)
	}
	for _, field := range []string{"schema_version", "version", "result_version"} {
		for _, val := range []any{"0.1.1", "0.2.0", "1.0.0", "0.1", "v0.1.0", "0.1.0-draft", "", 0.1, nil} {
			var value map[string]any
			_ = json.Unmarshal(raw, &value)
			if field == "schema_version" {
				value[field] = val
			} else {
				value["test"].(map[string]any)[field] = val
			}
			data, _ := json.Marshal(value)
			if _, err := decoder.Decode(data); err == nil {
				t.Fatalf("accepted %s %v", field, val)
			}
		}
	}
	// The result boundary keeps the registered conditional access rule for
	// inconclusive, which this runner never emits (REQ-NET-01-0025).
	for _, tc := range []struct {
		outcome string
		access  bool
		valid   bool
	}{{"inconclusive", true, true}, {"inconclusive", false, true}, {"platform_unavailable", false, true}, {"platform_error", false, true},
		{"platform_unavailable", true, false}, {"platform_error", true, false}, {"snapshot", false, false}, {"network_changed", false, false}} {
		var v map[string]any
		_ = json.Unmarshal(raw, &v)
		m := v["measurement"].(map[string]any)
		m["outcome"] = tc.outcome
		if !tc.access {
			delete(m, "access")
		}
		data, _ := json.Marshal(v)
		if _, err := decoder.Decode(data); (err == nil) != tc.valid {
			t.Fatalf("%s with access %v: %v", tc.outcome, tc.access, err)
		}
	}
	tcpObservation := map[string]any{"type": "tcp_connect", "status": "succeeded", "duration_ms": 0, "details": map[string]any{"address_family": "ipv4"}}
	access := func(v map[string]any) map[string]any {
		return v["measurement"].(map[string]any)["access"].(map[string]any)
	}
	for _, mutate := range []func(map[string]any){
		func(v map[string]any) { delete(v, "schema_version") },
		func(v map[string]any) { delete(v["test"].(map[string]any), "version") },
		func(v map[string]any) { delete(v["test"].(map[string]any), "result_version") },
		func(v map[string]any) { v["test"].(map[string]any)["id"] = "NET-03" },
		func(v map[string]any) { v["test"].(map[string]any)["id"] = "TCP-01" },
		func(v map[string]any) { v["measurement"].(map[string]any)["outcome"] = "connected" },
		func(v map[string]any) { v["measurement"].(map[string]any)["observations"] = []any{tcpObservation} },
		func(v map[string]any) { v["measurement"].(map[string]any)["platform_validation"] = "validated" },
		// A complete TCP-01 measurement under the NET-01 identity.
		func(v map[string]any) {
			v["measurement"] = map[string]any{"outcome": "connected", "observations": []any{tcpObservation}}
		},
		// The access object is complete and closed over its registered values.
		func(v map[string]any) { delete(access(v), "validated") },
		func(v map[string]any) { access(v)["network_type"] = "satellite" },
		func(v map[string]any) { access(v)["vpn_state"] = "yes" },
		func(v map[string]any) { access(v)["ipv4_available"] = true },
		func(v map[string]any) { access(v)["ipv6_available"] = "partial" },
		func(v map[string]any) { access(v)["validated"] = nil },
		func(v map[string]any) { access(v)["network_changed"] = "false" },
		func(v map[string]any) { delete(v["measurement"].(map[string]any), "access") },
		func(v map[string]any) {
			v["measurement"].(map[string]any)["access"].(map[string]any)["network_changed"] = true
		},
		func(v map[string]any) {
			v["measurement"].(map[string]any)["access"].(map[string]any)["interface"] = "canary-if-wlan7"
		},
		func(v map[string]any) { v["measurement"].(map[string]any)["egress"] = map[string]any{} },
		func(v map[string]any) { v["network_context"] = map[string]any{} },
	} {
		var v map[string]any
		_ = json.Unmarshal(raw, &v)
		mutate(v)
		data, _ := json.Marshal(v)
		if _, err := decoder.Decode(data); err == nil {
			t.Fatal("invalid output accepted", string(data))
		}
	}
}
