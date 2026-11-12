package barometer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"net/netip"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/measurement/net05"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/registry"
)

var net05Test = task.TestVersion{ID: "NET-05", Version: "0.1.0"}

func profileStep(id, test, input, options string) string {
	return fmt.Sprintf(`{"step_id":%q,"test_id":%q,"test_version":"0.1.0","input":%s,"options":%s}`, id, test, input, options)
}
func profileInput(steps ...string) string {
	return `{"profile":{"profile_id":"baseline","profile_version":1,"steps":[` + strings.Join(steps, ",") + `]}}`
}
func primitiveSteps() []string {
	return []string{
		profileStep("dns", "DNS-01", `{"hostname":"example.com","record_type":"A","resolver":{"ip_address":"8.8.8.8"}}`, `{"resolver_mode":"direct","transport":"tcp"}`),
		profileStep("tcp", "TCP-01", `{"ip_address":"8.8.8.8","port":443}`, `{}`),
		profileStep("tls", "TLS-01", `{"ip_address":"8.8.8.8","port":443,"server_name":"example.com"}`, `{"alpn_protocols":["h2"]}`),
		profileStep("http", "HTTP-01", httpInput, `{}`),
	}
}
func net05Task(id, input, options, execution string) string {
	s := strings.Replace(runTask(id, "", execution, false), `"test_id":"TCP-01"`, `"test_id":"NET-05"`, 1)
	s = strings.Replace(s, `"input":{}`, `"input":`+input, 1)
	return strings.Replace(s, `"options":{}`, `"options":`+options, 1)
}
func normalizedProfile(t *testing.T, input string) net05.Parameters {
	t.Helper()
	r, e := productionRegistry(nil, true)
	if e != nil {
		t.Fatal(e)
	}
	entry, _ := r.Lookup(net05Test)
	d := entry.Codec.Definition()
	lim := jsondecode.Limits{MaxBytes: 65536, MaxDepth: 64, MaxNodes: 10000}
	in, e := jsondecode.Decode([]byte(input), d.Input, lim)
	if e != nil {
		t.Fatal(e)
	}
	opt, e := jsondecode.Decode([]byte(`{}`), d.Options, lim)
	if e != nil {
		t.Fatal(e)
	}
	p, e := entry.Codec.Normalize(in, opt)
	if e != nil {
		t.Fatal(e)
	}
	return p.(net05.Parameters)
}

// Policy registrations deliberately belong to the primitive, never NET-05.
func net05AdmissionHarness(t *testing.T, edit func(*task.PolicyInput)) (*harness, *httpTransport) {
	t.Helper()
	h := runningHarness(t)
	h.adapters.registry = nil
	h.adapters.run.paths = nil
	in := synPolicyInput(2, "reg-a", func(p *task.Permissions) {
		p.AllowComposite = true
		p.AllowMediumRisk = true
		p.MaxRisk = task.HighRisk
		p.Tests = []task.TestVersion{net05Test, {ID: "DNS-01", Version: "0.1.0"}, {ID: "TCP-01", Version: "0.1.0"}, {ID: "TLS-01", Version: "0.1.0"}, {ID: "HTTP-01", Version: "0.1.0"}}
	})
	template := in.Targets[0]
	in.Targets = nil
	var ids []string
	r, _ := productionRegistry(nil, true)
	for _, step := range normalizedProfile(t, profileInput(primitiveSteps()...)).CompositeSteps() {
		e, _ := r.Lookup(step.Test)
		targets, ok := e.Targets.Targets(step.Parameters)
		if !ok {
			t.Fatal("targets")
		}
		for _, target := range targets {
			reg := template.Clone()
			reg.ID = fmt.Sprint("target-", len(ids))
			reg.Test = step.Test
			reg.Role = target.Role
			reg.Tuple = target.Tuple
			in.Targets = append(in.Targets, reg)
			ids = append(ids, reg.ID)
		}
	}
	in.Permissions.RegistrationIDs = slices.Clone(ids)
	for i := range in.Issuers {
		in.Issuers[i].Permissions.RegistrationIDs = slices.Clone(ids)
	}
	if edit != nil {
		edit(&in)
	}
	var e error
	h.adapters.policy, e = task.NewLocalPolicy(in)
	if e != nil {
		t.Fatal(e)
	}
	d := &httpTransport{}
	h.adapters.run.directDialers = d
	return h, d
}
func net05AdmissionClient(t *testing.T, edit func(*task.PolicyInput)) (*harness, *Client, *httpTransport) {
	t.Helper()
	h, d := net05AdmissionHarness(t, edit)
	c := h.initialized(t)
	t.Cleanup(func() { _ = c.Close() })
	return h, c, d
}
func assertNet05Rejected(t *testing.T, input, options, execution, want string, edit func(*task.PolicyInput)) {
	t.Helper()
	h, c, d := net05AdmissionClient(t, edit)
	data := net05Task("invalid-profile", input, options, execution)
	handle, e := c.AdmitTask(data, h.embedded(t, data))
	if handle != "" || e == nil || want != "" && codeOf(e) != want {
		t.Fatalf("admission = %q, %s; want %s", handle, codeOf(e), want)
	}
	if taskHandles(c) != 0 || h.replay.Commits() != 0 || d.calls.Load() != 0 || len(c.rt.runs.entries) != 0 {
		t.Fatal("rejection published state or contacted transport")
	}
}

// Admission coverage; summed declared maxima are covered with the limits
// family, and the absence of profile-version history in
// TestMNet05ProfileVersionWithoutHistory. Runner evidence is separate in net05_runner_test.go.
// requirement-ref: REQ-NET-05-0014
// requirement-ref: REQ-NET-05-0062
// requirement-ref: REQ-NET-05-0016
// requirement-ref: REQ-NET-05-0017
// requirement-ref: REQ-NET-05-0018
// requirement-ref: REQ-NET-05-0019
// requirement-ref: REQ-NET-05-0020
// requirement-ref: REQ-NET-05-0021
// requirement-ref: REQ-NET-05-0022
// requirement-ref: REQ-NET-05-0023
// requirement-ref: REQ-NET-05-0024
// requirement-ref: REQ-NET-05-0025
// requirement-ref: REQ-LIMIT-0019
func TestMNet05Admission(t *testing.T) {
	steps := primitiveSteps()
	for _, n := range []int{1, 4, 8} {
		t.Run(fmt.Sprint("steps-", n), func(t *testing.T) {
			h, c, d := net05AdmissionClient(t, nil)
			var selected []string
			for i := 0; i < n; i++ {
				selected = append(selected, strings.Replace(steps[i%4], `"step_id":"`+[]string{"dns", "tcp", "tls", "http"}[i%4]+`"`, fmt.Sprintf(`"step_id":"s%d"`, i), 1))
			}
			// Eight steps declare 172,034 received bytes, above the default bound.
			data := strings.Replace(net05Task("valid-profile", profileInput(selected...), `{}`, ""), `"limits":{}`, `"limits":{"max_bytes_received":262144}`, 1)
			handle, e := c.AdmitTask(data, h.embedded(t, data))
			if e != nil || handle == "" {
				t.Fatalf("valid: %s", codeOf(e))
			}
			if h.replay.Commits() != 1 || d.calls.Load() != 0 {
				t.Fatal("admission side effects")
			}
			if _, ok := c.bound.registry.Runner(net05Test); !ok {
				t.Fatal("NET-05 runner missing")
			}
		})
	}
	bad := []string{`{}`, `null`, `{"profile":null}`, profileInput(), profileInput(append(append([]string{}, steps...), append(steps, steps[0])...)...), profileInput(steps[0], steps[0])}
	for _, id := range []string{"", strings.Repeat("a", 65), "é", "x\n", "${value}", "{{value}}", "%NETWORK%"} {
		bad = append(bad, strings.Replace(profileInput(steps[1]), `"baseline"`, fmt.Sprintf("%q", id), 1), profileInput(strings.Replace(steps[1], `"tcp"`, fmt.Sprintf("%q", id), 1)))
	}
	for _, v := range []string{"0", "-1", "1.5", `"1"`, "null"} {
		bad = append(bad, strings.Replace(profileInput(steps[1]), `"profile_version":1`, `"profile_version":`+v, 1))
	}
	for _, field := range []string{"schedule", "execution", "safety", "labels", "script", "conditions"} {
		bad = append(bad, profileInput(strings.Replace(steps[1], `"step_id":`, `"`+field+`":{},"step_id":`, 1)))
	}
	bad = append(bad, profileInput(steps[0], steps[1], steps[2], strings.Replace(steps[3], "https://8.8.8.8/check", "https://user:pass@8.8.8.8/check", 1)), profileInput(steps[0], strings.Replace(steps[1], "8.8.8.8", "127.0.0.1", 1)), strings.Replace(profileInput(steps[1]), `"profile_id":`, `"other":0,"profile_id":`, 1), strings.Replace(profileInput(steps[1]), `"profile_version":1`, `"profile_version":1,"profile_version":2`, 1))
	for i, input := range bad {
		t.Run(fmt.Sprint("invalid-", i), func(t *testing.T) { assertNet05Rejected(t, input, `{}`, "", "", nil) })
	}
	for _, n := range []int{1, 64} {
		p := normalizedProfile(t, strings.Replace(profileInput(strings.Replace(steps[1], `"tcp"`, fmt.Sprintf("%q", strings.Repeat("s", n)), 1)), `"baseline"`, fmt.Sprintf("%q", strings.Repeat("p", n)), 1))
		view := p.CompositeSteps()
		view[0].ID = "changed"
		if p.CompositeSteps()[0].ID == "changed" {
			t.Fatal("mutable step container")
		}
	}
	// Original encoded profile bytes, including whitespace, enforce the cap.
	base := profileInput(steps[1])
	var parsed map[string]json.RawMessage
	if e := json.Unmarshal([]byte(base), &parsed); e != nil {
		t.Fatal(e)
	}
	sized := func(n int) string {
		profile := string(parsed["profile"])
		return `{"profile":` + profile[:len(profile)-1] + strings.Repeat(" ", n-len(profile)) + "}}"
	}
	// Both bounds apply and the stricter decides (REQ-NET-05-0025,
	// REQ-LIMIT-0019): with the 12-byte wrapper the common 49,152-byte input
	// bound admits a profile of at most 49,140 bytes.
	t.Run("largest admissible profile", func(t *testing.T) {
		h, c, d := net05AdmissionClient(t, nil)
		input := sized(49140)
		if len(input) != 49152 {
			t.Fatalf("input is %d bytes", len(input))
		}
		data := net05Task("largest-profile", input, `{}`, "")
		handle, e := c.AdmitTask(data, h.embedded(t, data))
		if e != nil || handle == "" {
			t.Fatalf("49,140-byte profile: %s", codeOf(e))
		}
		if h.replay.Commits() != 1 || d.calls.Load() != 0 {
			t.Fatal("admission side effects")
		}
	})
	for _, n := range []int{49141, 49151, 49152, 49153} {
		input := sized(n)
		if n <= 49152 {
			normalizedProfile(t, input)
		}
		// The codec alone accepts up to 49,152 bytes; the common input bound,
		// which includes the wrapper, rejects first with the input-bound class.
		assertNet05Rejected(t, input, `{}`, "", "task_limit_exceeded", nil)
	}
}

// M-NET-05-ADMISSION profile-version row: monotonic increase is the issuer's
// obligation. Barometer validates only a positive integer, keeps no record of
// other tasks' versions and admits an equal or lower version for the same
// profile_id, retaining the issuer's value unchanged.
// requirement-ref: REQ-NET-05-0062
// requirement-ref: REQ-NET-05-0063
func TestMNet05ProfileVersionWithoutHistory(t *testing.T) {
	h, c, d := net05AdmissionClient(t, nil)
	step := primitiveSteps()[1]
	versions := []int64{5, 3, 5, 1, 9223372036854775807, 2}
	for i, version := range versions {
		input := strings.Replace(profileInput(step), `"profile_version":1`, fmt.Sprintf(`"profile_version":%d`, version), 1)
		if id, got := normalizedProfile(t, input).Profile(); id != "baseline" || got != version {
			t.Fatalf("normalized profile = %s/%d; want baseline/%d", id, got, version)
		}
		data := net05Task(fmt.Sprintf("profile-version-%d", i), input, `{}`, "")
		if handle, e := c.AdmitTask(data, h.embedded(t, data)); e != nil || handle == "" {
			t.Fatalf("profile_version %d after %v: %s", version, versions[:i], codeOf(e))
		}
	}
	if taskHandles(c) != len(versions) || h.replay.Commits() != len(versions) || d.calls.Load() != 0 {
		t.Fatal("profile-version admission compared tasks or contacted transport")
	}
}

// requirement-ref: REQ-TRUST-0039
// requirement-ref: REQ-PROXY-0044
// requirement-ref: REQ-NET-05-0019
func TestMNet05AtomicTargetAndProxyRules(t *testing.T) {
	steps := primitiveSteps()
	all := profileInput(steps...)
	for _, edit := range []func(*task.PolicyInput){
		func(in *task.PolicyInput) { in.Targets[len(in.Targets)-1].Tuple.URL = "https://8.8.8.8/other" },
		func(in *task.PolicyInput) { in.Targets[len(in.Targets)-1].Test = net05Test },
		func(in *task.PolicyInput) { in.Targets[1].Paths = task.Paths{Proxy: true} },
		func(in *task.PolicyInput) { in.Targets[4].ExpiresAt = task.TimestampFromTime(epoch) },
		func(in *task.PolicyInput) {
			in.Issuers[0].Permissions.RegistrationIDs = in.Issuers[0].Permissions.RegistrationIDs[:4]
		},
	} {
		assertNet05Rejected(t, all, `{}`, "", "target_not_registered", edit)
	}
	for _, issuer := range []bool{false, true} {
		assertNet05Rejected(t, all, `{}`, "", "task_issuer_not_authorized", func(in *task.PolicyInput) {
			p := &in.Permissions
			if issuer {
				p = &in.Issuers[0].Permissions
			}
			p.Tests = slices.DeleteFunc(slices.Clone(p.Tests), func(v task.TestVersion) bool { return v.ID == "HTTP-01" })
			p.RegistrationIDs = slices.Clone(p.RegistrationIDs[:4])
		})
		assertNet05Rejected(t, all, `{}`, "", "task_issuer_not_authorized", func(in *task.PolicyInput) {
			p := &in.Permissions
			if issuer {
				p = &in.Issuers[0].Permissions
			}
			p.AllowComposite = false
			p.Tests = slices.DeleteFunc(slices.Clone(p.Tests), func(v task.TestVersion) bool { return v == net05Test })
		})
	}
	// A valid final DNS system/UDP step makes the entire proxy profile invalid,
	// even when authority/targets would also refuse it (GO-SPEC-ISSUE-009).
	for _, dns := range []string{
		profileStep("last", "DNS-01", `{"hostname":"example.com","record_type":"A"}`, `{}`),
		profileStep("last", "DNS-01", `{"hostname":"example.com","record_type":"A","resolver":{"ip_address":"8.8.8.8"}}`, `{"resolver_mode":"direct","transport":"udp"}`),
	} {
		assertNet05Rejected(t, profileInput(steps[1], steps[2], steps[3], dns), `{}`, `"network_path":"proxy"`, "task_schema_invalid", func(in *task.PolicyInput) {
			in.Permissions.Paths = task.Paths{Direct: true}
			in.Targets[0].Tuple.Hostname = "other.example.com"
		})
	}
	hostnameHTTP := strings.Replace(steps[3], "https://8.8.8.8/check", "https://example.com/check", 1)
	assertNet05Rejected(t, profileInput(steps[1], hostnameHTTP), `{}`, `"network_path":"proxy"`, "task_schema_invalid", nil)
	h, c, d := net05AdmissionClient(t, nil)
	data := net05Task("compatible-proxy", all, `{}`, `"network_path":"proxy"`)
	if _, e := c.AdmitTask(data, h.embedded(t, data)); e != nil {
		t.Fatalf("compatible proxy: %s", codeOf(e))
	}
	if d.calls.Load() != 0 {
		t.Fatal("proxy admission contacted transport")
	}
	// A target's elevated risk cannot be hidden inside a medium-risk composite.
	assertNet05Rejected(t, all, `{}`, "", "task_issuer_not_authorized", func(in *task.PolicyInput) {
		in.Permissions.MaxRisk = task.MediumRisk
		in.Targets[4].BaselineRisk = task.HighRisk
	})
}

// requirement-ref: REQ-NET-05-0028
// requirement-ref: REQ-NET-05-0029
// requirement-ref: REQ-NET-05-0030
func TestMNet05Options(t *testing.T) {
	all := profileInput(primitiveSteps()...)
	for _, opt := range []string{`null`, `[]`, `{"sequence":"parallel"}`, `{"retry":true}`, `{"execution":{}}`, `{"safety":{}}`} {
		assertNet05Rejected(t, all, opt, "", "task_schema_invalid", nil)
	}
	for _, s := range []string{
		profileStep("last", "TCP-01", `{"ip_address":"8.8.8.8","port":443}`, `{"retry":true}`),
		profileStep("last", "TLS-01", `{"ip_address":"8.8.8.8","port":443,"server_name":"example.com"}`, `{"insecure_skip_verify":true}`),
		profileStep("last", "DNS-01", `{"hostname":"example.com","record_type":"A"}`, `{"resolver_mode":"invalid"}`),
		profileStep("last", "HTTP-01", httpInput, `{"max_body_bytes":0}`),
	} {
		assertNet05Rejected(t, profileInput(primitiveSteps()[1], s), `{}`, "", "task_schema_invalid", nil)
	}
	assertNet05Rejected(t, profileInput(profileStep("http", "HTTP-01", httpInput, `{"max_body_bytes":65537}`)), `{}`, "", "task_limit_exceeded", nil)
	input := profileInput(profileStep("http", "HTTP-01", httpInput, `{"max_body_bytes":4096}`))
	h, c, _ := net05AdmissionClient(t, nil)
	data := strings.Replace(net05Task("primitive-bound", input, `{}`, ""), `"limits":{}`, `"limits":{"max_bytes_received":4095}`, 1)
	if handle, e := c.AdmitTask(data, h.embedded(t, data)); handle != "" || codeOf(e) != "task_limit_exceeded" {
		t.Fatal("primitive bound not enforced")
	}
	// Composite options cannot carry limits, retries, ordering or safety.
	for _, opt := range []string{`{"limits":{}}`, `{"limits":{"max_connections":13}}`, `{"max_duration_ms":120000}`, `{"retries":1}`, `{"max_attempts":2}`, `{"parallel":true}`, `{"order":["http","dns"]}`, `{"fallback":"direct"}`, `{"network_path":"direct"}`, `{"":0}`} {
		assertNet05Rejected(t, all, opt, "", "task_schema_invalid", nil)
	}
	// A step cannot override primitive limits, retry, ordering, path or safety
	// beside its input/options; the closed step object rejects the member.
	for _, member := range []string{`"limits":{}`, `"limits":{"max_connections":13,"max_duration_ms":120000}`, `"retries":1`, `"max_attempts":2`, `"order":0`, `"parallel":true`, `"network_path":"direct"`, `"risk_level":"low"`} {
		assertNet05Rejected(t, profileInput(primitiveSteps()[1], strings.Replace(primitiveSteps()[3], `"step_id":`, member+`,"step_id":`, 1)), `{}`, "", "task_schema_invalid", nil)
	}
	// Each primitive option is checked by that primitive's exact versioned
	// schema, independently and in a later step after a valid one; limit and
	// retry members inside step options are unknown primitive options.
	tcp := `{"ip_address":"8.8.8.8","port":443}`
	tls := `{"ip_address":"8.8.8.8","port":443,"server_name":"example.com"}`
	dns := `{"hostname":"example.com","record_type":"A","resolver":{"ip_address":"8.8.8.8"}}`
	for _, tc := range []struct{ test, input, options, want string }{
		{"TCP-01", tcp, `{"limits":{"max_connections":13}}`, "task_schema_invalid"},
		{"TCP-01", tcp, `{"max_duration_ms":120000}`, "task_schema_invalid"},
		{"TCP-01", tcp, `{"timeout_ms":1}`, "task_schema_invalid"},
		{"TCP-01", tcp, `null`, "task_schema_invalid"},
		{"TLS-01", tls, `{"alpn_protocols":["h3"]}`, "task_schema_invalid"},
		{"TLS-01", tls, `{"alpn_protocols":["h2","h2"]}`, "task_schema_invalid"},
		{"TLS-01", tls, `{"alpn_protocols":"h2"}`, "task_schema_invalid"},
		{"TLS-01", tls, `{"session_resumption":true}`, "task_schema_invalid"},
		{"TLS-01", tls, `{"retries":1}`, "task_schema_invalid"},
		{"DNS-01", dns, `{"resolver_mode":"direct","transport":"doh"}`, "task_schema_invalid"},
		{"DNS-01", dns, `{"resolver_mode":"direct","transport":"tcp","collect_additional_responses":"yes"}`, "task_schema_invalid"},
		{"DNS-01", dns, `{"resolver_mode":"direct","transport":"tcp","additional_response_window_ms":1}`, "task_schema_invalid"},
		{"DNS-01", dns, `{"resolver_mode":"direct","transport":"tcp","max_attempts":2}`, "task_schema_invalid"},
		{"HTTP-01", httpInput, `{"max_redirects":3}`, "task_limit_exceeded"},
		{"HTTP-01", httpInput, `{"follow_redirects":"yes"}`, "task_schema_invalid"},
		{"HTTP-01", httpInput, `{"address_family":"ipv5"}`, "task_schema_invalid"},
		{"HTTP-01", httpInput, `{"max_body_bytes":"4096"}`, "task_schema_invalid"},
		{"HTTP-01", httpInput, `{"limits":{"max_bytes_received":262145}}`, "task_schema_invalid"},
		{"HTTP-01", httpInput, `{"retry":true}`, "task_schema_invalid"},
	} {
		assertNet05Rejected(t, profileInput(primitiveSteps()[1], profileStep("last", tc.test, tc.input, tc.options)), `{}`, "", tc.want, nil)
	}
	// A narrower primitive option within the outer envelope is accepted.
	h, c, d := net05AdmissionClient(t, nil)
	data = net05Task("primitive-option", input, `{}`, "")
	if handle, e := c.AdmitTask(data, h.embedded(t, data)); e != nil || handle == "" || d.calls.Load() != 0 {
		t.Fatalf("narrower primitive option: %s", codeOf(e))
	}
}

// M-NET-05-LIMITS registered ceilings at admission: 60,000 ms, 12 connections
// or tunnels, 65,536 bytes sent and 262,144 received, each below, at and
// above the ceiling with the other limits omitted, and over-budget summed
// declared maxima rejected atomically before contact. Declared-maxima
// boundaries and actual aggregate use are driven in internal/execution.
// requirement-ref: REQ-NET-05-0026
// requirement-ref: REQ-NET-05-0037
// requirement-ref: REQ-NET-05-0038
// requirement-ref: REQ-LIMIT-0020
// requirement-ref: REQ-LIMIT-0035
func TestMNet05RegisteredCeilingAdmission(t *testing.T) {
	all := profileInput(primitiveSteps()...)
	for _, tc := range []struct {
		key     string
		ceiling int
	}{{"max_duration_ms", 60000}, {"max_connections", 12}, {"max_bytes_sent", 65536}, {"max_bytes_received", 262144}} {
		for _, delta := range []int{-1, 0, 1} {
			t.Run(fmt.Sprintf("%s/%d", tc.key, tc.ceiling+delta), func(t *testing.T) {
				data := strings.Replace(net05Task("ceiling", all, `{}`, ""), `"limits":{}`, fmt.Sprintf(`"limits":{%q:%d}`, tc.key, tc.ceiling+delta), 1)
				h, c, d := net05AdmissionClient(t, nil)
				handle, e := c.AdmitTask(data, h.embedded(t, data))
				if d.calls.Load() != 0 {
					t.Fatal("admission contacted transport")
				}
				if delta > 0 {
					if handle != "" || codeOf(e) != "task_limit_exceeded" || taskHandles(c) != 0 || h.replay.Commits() != 0 {
						t.Fatalf("above ceiling: %q %s", handle, codeOf(e))
					}
				} else if e != nil || handle == "" {
					t.Fatalf("within ceiling: %s", codeOf(e))
				}
			})
		}
	}
	// Summed declared maxima (GO-SPEC-ISSUE-021): the four primitive steps
	// declare 4 connections and 86,017 received bytes, and four HTTP-01 steps
	// following two redirects plus one TCP-01 step declare 13 connections.
	tcp := profileStep("extra", "TCP-01", `{"ip_address":"8.8.8.8","port":443}`, `{}`)
	var redirecting []string
	for i := 0; i < 4; i++ {
		redirecting = append(redirecting, profileStep(fmt.Sprint("h", i), "HTTP-01", httpInput, `{"follow_redirects":true,"max_redirects":2}`))
	}
	for _, tc := range []struct {
		name, input, limits string
		over                bool
	}{
		{"connections-at-task-bound", all, `"max_connections":4`, false},
		{"connections-over-task-bound", all, `"max_connections":3`, true},
		{"received-at-task-bound", all, `"max_bytes_received":86017`, false},
		{"received-over-task-bound", all, `"max_bytes_received":86016`, true},
		{"connections-at-composite", profileInput(redirecting...), `"max_connections":12,"max_bytes_received":262144`, false},
		{"connections-over-composite", profileInput(append(redirecting, tcp)...), `"max_connections":12,"max_bytes_received":262144`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := strings.Replace(net05Task("declared", tc.input, `{}`, ""), `"limits":{}`, `"limits":{`+tc.limits+`}`, 1)
			h, c, d := net05AdmissionClient(t, nil)
			handle, e := c.AdmitTask(data, h.embedded(t, data))
			if d.calls.Load() != 0 {
				t.Fatal("admission contacted transport")
			}
			if tc.over {
				if handle != "" || codeOf(e) != "task_limit_exceeded" || taskHandles(c) != 0 || h.replay.Commits() != 0 || len(c.rt.runs.entries) != 0 {
					t.Fatalf("over declared bound: %q %s", handle, codeOf(e))
				}
			} else if e != nil || handle == "" {
				t.Fatalf("within declared bound: %s", codeOf(e))
			}
		})
	}
}

// Task dispatch portion of VERSIONS. Result-version and emitted-result cases
// remain with the runner/result integration; no measurement is executed here.
// requirement-ref: REQ-NET-05-0039
func TestMNet05Versions(t *testing.T) {
	steps := primitiveSteps()
	for _, ver := range []string{`"0.1.1"`, `"0.2.0"`, `"1.0.0"`, `""`, `"v0.1.0"`, `"0.1.0 "`, `1`, `null`} {
		for _, s := range steps {
			assertNet05Rejected(t, profileInput(strings.Replace(s, `"test_version":"0.1.0"`, `"test_version":`+ver, 1)), `{}`, "", "", nil)
		}
		h, c, _ := net05AdmissionClient(t, nil)
		data := strings.Replace(net05Task("bad-version", profileInput(steps...), `{}`, ""), `"test_version":"0.1.0"`, `"test_version":`+ver, 1)
		if handle, e := c.AdmitTask(data, h.embedded(t, data)); e == nil || handle != "" || h.replay.Commits() != 0 {
			t.Fatal("outer version accepted")
		}
	}
	for _, id := range []string{"NET-05", "NET-01", "NET-03", "DNS-03", "SYN-01", "tcp-01"} {
		assertNet05Rejected(t, profileInput(steps[1], strings.Replace(steps[3], "HTTP-01", id, 1)), `{}`, "", "task_schema_invalid", nil)
	}
	assertNet05Rejected(t, profileInput(profileStep("cross-test", "DNS-01", httpInput, `{}`)), `{}`, "", "task_schema_invalid", nil)
	p := normalizedProfile(t, profileInput(steps...))
	for i, s := range p.CompositeSteps() {
		if s.Test.Version != "0.1.0" || s.ID != []string{"dns", "tcp", "tls", "http"}[i] {
			t.Fatal("dispatch changed version or order")
		}
	}
	// No historical floor: equal and lower positive revisions are accepted.
	h, c, _ := net05AdmissionClient(t, nil)
	for i, v := range []int{2, 2, 1} {
		data := strings.Replace(net05Task(fmt.Sprint("history-", i), profileInput(steps...), `{}`, ""), `"profile_version":1`, fmt.Sprintf(`"profile_version":%d`, v), 1)
		if _, e := c.AdmitTask(data, h.embedded(t, data)); e != nil {
			t.Fatal("unexpected history enforcement")
		}
	}
}

func TestMNet05CodecBoundsAndComposition(t *testing.T) {
	r, _ := productionRegistry(nil, true)
	entry, _ := r.Lookup(net05Test)
	rules := entry.Targets.(net05Rules)
	if targets, ok := rules.Targets(net05.Parameters{}); ok || targets != nil || rules.Compatible(net05.Parameters{}) {
		t.Fatal("zero profile accepted")
	}
	if targets, ok := rules.Targets(nil); ok || targets != nil || rules.Compatible(nil) {
		t.Fatal("wrong parameter type accepted")
	}
	// A nested primitive keeps its common options cap, including encoded spaces.
	for _, n := range []int{8192, 8193} {
		options := "{" + strings.Repeat(" ", n-2) + "}"
		input := profileInput(profileStep("tcp", "TCP-01", `{"ip_address":"8.8.8.8","port":443}`, options))
		if n == 8192 {
			normalizedProfile(t, input)
		} else {
			assertNet05Rejected(t, input, `{}`, "", "task_limit_exceeded", nil)
		}
	}
	// The codec's own profile cap is checked independently of the narrower
	// wrapped common input cap.
	input := profileInput(primitiveSteps()[1])
	input = strings.Replace(input, `"profile_id":`, strings.Repeat(" ", 49152)+`"profile_id":`, 1)
	d := entry.Codec.Definition()
	if _, e := jsondecode.Decode([]byte(input), d.Input, jsondecode.Limits{MaxBytes: 65536, MaxDepth: 64, MaxNodes: 10000}); e == nil {
		t.Fatal("oversized profile accepted")
	}
	// A last primitive cannot disappear simply because its version is absent.
	last := strings.Replace(primitiveSteps()[3], `"test_version":"0.1.0",`, "", 1)
	assertNet05Rejected(t, profileInput(primitiveSteps()[1], last), `{}`, "", "task_schema_invalid", nil)
	for _, field := range []string{"input", "options"} {
		var step map[string]json.RawMessage
		if e := json.Unmarshal([]byte(primitiveSteps()[1]), &step); e != nil {
			t.Fatal(e)
		}
		delete(step, field)
		raw, _ := json.Marshal(step)
		assertNet05Rejected(t, profileInput(string(raw)), `{}`, "", "task_schema_invalid", nil)
	}
}

// Every primitive's registered baseline contributes independently of the outer
// NET-05 baseline and target-registration risk.
func TestMNet05PrimitiveRiskAndMissingCapability(t *testing.T) {
	for _, missing := range []bool{false, true} {
		h, _ := net05AdmissionHarness(t, func(in *task.PolicyInput) { in.Permissions.MaxRisk = task.MediumRisk })
		original, _ := productionRegistry(nil, true)
		var entries []registry.Entry
		for _, id := range []string{"DNS-01", "TCP-01", "TLS-01", "HTTP-01", "NET-05"} {
			e, _ := original.Lookup(task.TestVersion{ID: id, Version: "0.1.0"})
			if id == "HTTP-01" {
				if missing {
					continue
				}
				e.BaselineRisk = task.HighRisk
			}
			entries = append(entries, e)
		}
		var err error
		h.adapters.registry, err = registry.New(entries)
		if err != nil {
			t.Fatal(err)
		}
		c := h.initialized(t)
		t.Cleanup(func() { _ = c.Close() })
		data := net05Task("primitive-risk", profileInput(primitiveSteps()...), `{}`, "")
		handle, err := c.AdmitTask(data, h.embedded(t, data))
		want := "task_issuer_not_authorized"
		if missing {
			want = "task_unsupported"
		}
		if handle != "" || codeOf(err) != want || h.replay.Commits() != 0 {
			t.Fatalf("primitive admission = %q, %s; want %s", handle, codeOf(err), want)
		}
	}
}

// net05CanarySteps is a synthetic profile of DNS-01, TCP-01, TLS-01 and
// HTTP-01 steps whose inputs and options carry canaries: a canary hostname,
// a canary target address and port, a canary server name, a canary HTTPS URL
// path and a canary step option value. The second DNS step's system lookup
// fails with the canary platform error. The first DNS step queries a direct
// resolver whose scripted response carries the DNS-01 message canaries.
func net05CanarySteps() []string {
	return []string{
		profileStep("dns", "DNS-01", `{"hostname":"canary-n5host-151.example","record_type":"A","resolver":{"ip_address":"8.8.4.4"}}`, `{"resolver_mode":"direct","transport":"tcp"}`),
		profileStep("dns-fail", "DNS-01", `{"hostname":"canary-n5fail-156.example","record_type":"A"}`, `{}`),
		profileStep("tcp", "TCP-01", `{"ip_address":"8.8.4.4","port":47391}`, `{}`),
		profileStep("tls", "TLS-01", `{"ip_address":"8.8.4.4","port":8443,"server_name":"canary-n5sni-152.example"}`, `{}`),
		profileStep("http", "HTTP-01", `{"url":"https://8.8.4.4/canary-n5path-153","expected_status":200}`, `{"max_body_bytes":40961}`),
	}
}

// net05Canaries are the REQ-NET-05-0048 items. The outer task carries
// schedule and labels canaries; its safety declaration has only registered
// enumerated members, so the check proves no safety member is copied.
var net05Canaries = canarySet{
	{"outer schedule", "07:13:29"}, {"outer labels key", "canary-label-key-154"}, {"outer labels value", "canary-label-value-155"},
	{"step target hostname", "canary-n5host-151.example"}, {"failing step hostname", "canary-n5fail-156.example"}, {"step target URL path", "canary-n5path-153"},
	{"step target address", "8.8.4.4"}, {"step target port", "47391"}, {"TLS step server name", "canary-n5sni-152.example"}, {"profile option value", "40961"},
	{"response body", "CANARY-N5BODY"}, {"response header value", "canary-n5hval-157"}, {"certificate subject and issuer", loopback.CertificateCanary},
	{"leaf serial number", "424242"}, {"issuer serial number", "91337"}, {"platform error", "CANARY-N5-PLATFORM-ERROR"},
	{"raw TCP error", "CANARY-N5-SOCKERR"}, {"DNS answer TTL", "86413"}, {"DNS authority record", "canary-authority-74.example"},
	{"DNS additional record", "canary-additional-75.example"}, {"raw DNS message pattern", "CANARY-RAWDNS-BYTES"}, {"combined restriction verdict", "broad_restriction_suspected"}, {"allowlist verdict", "allowlist_detected"},
}

func net05CanaryTask(id, execution string) string {
	s := net05Task(id, profileInput(net05CanarySteps()...), `{}`, execution)
	s = strings.Replace(s, `"schedule":{`, `"schedule":{"starts_at":"2026-10-19T07:13:29Z",`, 1)
	return strings.Replace(s, `"safety":{`, `"labels":{"canary-label-key-154":"canary-label-value-155"},"safety":{`, 1)
}

// net05CanaryHarness registers exactly the canary profile's targets and
// routes each checked authority to its scripted endpoint.
func net05CanaryHarness(t *testing.T, stall bool) (*harness, *httpTransport, *loopback.HTTPServer) {
	t.Helper()
	h := runningHarness(t)
	h.adapters.registry = nil
	h.adapters.run.paths = nil
	in := synPolicyInput(2, "reg-a", func(p *task.Permissions) {
		p.AllowComposite, p.AllowMediumRisk, p.MaxRisk = true, true, task.HighRisk
		p.Tests = []task.TestVersion{net05Test, {ID: "DNS-01", Version: "0.1.0"}, {ID: "TCP-01", Version: "0.1.0"}, {ID: "TLS-01", Version: "0.1.0"}, {ID: "HTTP-01", Version: "0.1.0"}}
	})
	template := in.Targets[0]
	in.Targets = nil
	var ids []string
	r, _ := productionRegistry(nil, true)
	for _, step := range normalizedProfile(t, profileInput(net05CanarySteps()...)).CompositeSteps() {
		e, _ := r.Lookup(step.Test)
		targets, ok := e.Targets.Targets(step.Parameters)
		if !ok {
			t.Fatal("targets")
		}
		for _, target := range targets {
			reg := template.Clone()
			reg.ID, reg.Test, reg.Role, reg.Tuple = fmt.Sprint("target-", len(ids)), step.Test, target.Role, target.Tuple
			in.Targets = append(in.Targets, reg)
			ids = append(ids, reg.ID)
		}
	}
	in.Permissions.RegistrationIDs = slices.Clone(ids)
	in.Permissions.LabelNamespaces = []string{"canary-label-key-154"}
	for i := range in.Issuers {
		in.Issuers[i].Permissions.RegistrationIDs = slices.Clone(ids)
		in.Issuers[i].Permissions.LabelNamespaces = []string{"canary-label-key-154"}
	}
	var err error
	if h.adapters.policy, err = task.NewLocalPolicy(in); err != nil {
		t.Fatal(err)
	}
	ca, err := loopback.NewTestCA()
	if err != nil {
		t.Fatal(err)
	}
	server := func(name string, cfg loopback.HTTPConfig) *loopback.HTTPServer {
		cert, err := ca.Issue(name, false)
		if err != nil {
			t.Fatal(err)
		}
		cfg.Certificate = cert
		srv, err := loopback.ListenHTTP(cfg)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = srv.Close() })
		return srv
	}
	tlsSrv := server("canary-n5sni-152.example", loopback.HTTPConfig{})
	body := "CANARY-N5BODY"
	httpSrv := server("8.8.4.4", loopback.HTTPConfig{Stall: stall, Raw: []byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nX-Canary-N5: canary-n5hval-157\r\nContent-Length: %d\r\n\r\n%s", len(body), body))})
	h.adapters.measurementRoots = ca.Pool()
	dnsSrv, err := loopback.ListenDNS(func(q []byte) [][]byte { return [][]byte{dnsCanaryResponse(q, 0, dnsmessage.RCodeSuccess, "1.1.1.1")} })
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { dnsSrv.Close() })
	h.adapters.run.resolver = dnsSystemResolver{func(_ context.Context, _, name string) ([]netip.Addr, error) {
		return nil, errors.New("CANARY-N5-PLATFORM-ERROR " + name)
	}}
	d := &httpTransport{dial: func(ctx context.Context, network, address string) (measurement.Conn, error) {
		var nd net.Dialer
		switch address {
		case "8.8.4.4:53":
			return nd.DialContext(ctx, network, dnsSrv.Address())
		case "8.8.4.4:47391":
			return &tcpCanaryConn{}, nil
		case "8.8.4.4:8443":
			return nd.DialContext(ctx, "tcp", tlsSrv.Address())
		case "8.8.4.4:443":
			return nd.DialContext(ctx, "tcp", httpSrv.Address())
		}
		return nil, errors.New("CANARY-N5-SOCKERR unexpected destination")
	}}
	h.adapters.run.directDialers = d
	readyComposite(h)
	return h, d, httpSrv
}

var net05Allowed = func() []string {
	out := []string{"outcome", "observations", "profile.profile_id", "profile.profile_version", "network_validation.state", "network_validation.source",
		"steps[].step_id", "steps[].test.id", "steps[].test.version", "steps[].test.result_version", "steps[].state", "stopped_before_step"}
	seen := map[string]bool{}
	for _, set := range [][]string{tcpAllowed, tlsAllowed, dnsAllowed, httpAllowed} {
		for _, p := range set {
			if !seen[p] {
				seen[p] = true
				out = append(out, "steps[].measurement."+p)
			}
		}
	}
	return out
}()

type net05Run struct {
	s   *surfaces
	m   map[string]any
	raw []byte
}

// net05CanaryRun runs the canary profile once. validated sets the platform
// validation state; cancel interrupts the run while the HTTP step's server
// holds its response.
func net05CanaryRun(t *testing.T, id string, validated platform.Availability, cancel bool) net05Run {
	t.Helper()
	h, _, srv := net05CanaryHarness(t, cancel)
	a := defaultAccess
	a.Validated = validated
	h.run.network.Set(a)
	c := h.initialized(t)
	defer c.Close()
	s := &surfaces{}
	rh, err := c.StartRun(h.admit(t, c, net05CanaryTask(id, "")))
	if err != nil {
		t.Fatal(codeOf(err))
	}
	if cancel {
		ctx, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		if err := srv.Wait(ctx, func(c loopback.HTTPCounts) bool { return len(c.Requests) == 1 }); err != nil {
			t.Fatal(err)
		}
		if err := c.CancelRun(rh); err != nil {
			t.Fatal(codeOf(err))
		}
	}
	waitTerminal(t, c, rh)
	_, err = c.RunState(rh)
	s.err(err)
	s.collect(h, c)
	rejectedClean(t, h, c, net05Canaries, withUnknownInput(net05CanaryTask("rejected", ""), net05Canaries))
	if len(s.results) != 1 {
		t.Fatal("not one result")
	}
	var r map[string]any
	if json.Unmarshal(s.results[0], &r) != nil {
		t.Fatal("result is not JSON")
	}
	m, _ := r["measurement"].(map[string]any)
	return net05Run{s, m, s.results[0]}
}

// net05Decoder is the strict decoder of the canary profile's results.
func net05Decoder(t *testing.T, id string) *result.Decoder {
	t.Helper()
	p := normalizedProfile(t, profileInput(net05CanarySteps()...))
	plan := &result.ProfilePlan{ID: "baseline", Version: 1}
	for _, s := range p.CompositeSteps() {
		plan.Steps = append(plan.Steps, result.StepPlan{ID: s.ID, Output: measurementOutput(s.Test, s.Parameters)})
	}
	cd, _ := observation.NewContextDecoder(result.Contract, observation.ContextApprovals{})
	path, _ := observation.NewPath(1, 0, "direct", "")
	d, err := result.New(result.Binding{TaskID: id, Output: result.Output{Test: model.Test{ID: "NET-05", Version: "0.1.0", ResultVersion: "0.1.0"}}, MaxAttempts: 1, MaxDurationMS: 60000,
		Implementation: implementation, ContextDecoder: cd, Path: path, AccessGeneration: 1, Profile: plan})
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// M-NET-05-MINIMIZATION: one canary per REQ-NET-05-0048 item through a
// synthetic DNS-01, TCP-01, TLS-01 and HTTP-01 profile. Only the registered
// NET-05 and primitive fields remain; the DNS-01 answer is retained only in
// the DNS answers path; no outer envelope member, per-step Network Context or
// combined verdict is copied; an injected extra field is rejected.
// requirement-ref: REQ-NET-05-0048
func TestMNet05Minimization(t *testing.T) {
	net05Canaries.unique(t)
	run := net05CanaryRun(t, "minimization", platform.AvailabilityYes, false)
	run.s.check(t, net05Canaries)
	m := minimized(t, run.raw, net05Allowed)
	if m == nil || m["outcome"] != "completed" {
		t.Fatalf("outcome %v", m["outcome"])
	}
	steps := m["steps"].([]any)
	want := []string{"resolved", "resolver_error", "connected", "handshake_succeeded", "expected_response"}
	for i, s := range steps {
		step := s.(map[string]any)
		if got := step["measurement"].(map[string]any)["outcome"]; got != want[i] {
			t.Fatalf("step %d outcome %v", i, got)
		}
	}
	if strings.Count(string(run.raw), "1.1.1.1") != 1 || !strings.Contains(string(run.raw), `"answers":[{"type":"A","value":"1.1.1.1"}]`) {
		t.Fatalf("the DNS answer is retained outside the DNS answers path: %s", run.raw)
	}
	for _, member := range []string{`"schedule"`, `"execution":{"state"`, `"safety"`, `"labels"`, `"risk_level"`, `"requires_user_consent"`, `"input"`, `"options"`} {
		if n := strings.Count(string(run.raw), member); member == `"execution":{"state"` && n > 1 || member != `"execution":{"state"` && n != 0 {
			t.Errorf("result copies %s", member)
		}
	}
	if strings.Count(string(run.raw), `"network_context"`) > 1 {
		t.Error("per-step Network Context")
	}
	t.Run("incomplete", func(t *testing.T) {
		r := net05CanaryRun(t, "minimization", platform.AvailabilityYes, true)
		r.s.check(t, net05Canaries)
		minimized(t, r.raw, net05Allowed)
	})
	t.Run("injected fields", func(t *testing.T) {
		step := func(v map[string]any, i int) map[string]any {
			return v["measurement"].(map[string]any)["steps"].([]any)[i].(map[string]any)
		}
		var inject []func(map[string]any)
		for _, field := range []string{"schedule", "execution", "safety", "labels", "network_context", "input", "options", "target", "url", "error", "broad_restriction_suspected", "allowlist_detected", "classification"} {
			inject = append(inject, func(v map[string]any) { v["measurement"].(map[string]any)[field] = map[string]any{} },
				func(v map[string]any) { step(v, 0)[field] = map[string]any{} },
				func(v map[string]any) { step(v, 4)["measurement"].(map[string]any)[field] = "synthetic" },
				func(v map[string]any) {
					v["measurement"].(map[string]any)["profile"].(map[string]any)[field] = "synthetic"
				})
		}
		inject = append(inject, func(v map[string]any) { v["measurement"].(map[string]any)["outcome"] = "broad_restriction_suspected" },
			func(v map[string]any) { v["measurement"].(map[string]any)["outcome"] = "allowlist_detected" })
		assertRejected(t, net05Decoder(t, "minimization"), run.raw, inject...)
	})
}

// M-NET-05-INTERPRETATION: the shared prohibited-claims check with the
// NET-05 claims over completed, incomplete, network_not_validated and
// inconclusive runs. The completed run includes a primitive negative outcome
// and the outcome stays completed, never a combined classification.
// requirement-ref: REQ-NET-05-0001
// requirement-ref: REQ-NET-05-0002
// requirement-ref: REQ-NET-05-0003
// requirement-ref: REQ-NET-05-0004
// requirement-ref: REQ-NET-05-0049
// requirement-ref: REQ-NET-05-0050
// requirement-ref: REQ-NET-05-0051
// requirement-ref: REQ-NET-05-0052
func TestMNet05Interpretation(t *testing.T) {
	var raws [][]byte
	for _, tc := range []struct {
		validated platform.Availability
		cancel    bool
		outcome   string
	}{{platform.AvailabilityYes, false, "completed"}, {platform.AvailabilityYes, true, ""}, {platform.AvailabilityNo, false, "network_not_validated"}, {platform.AvailabilityUnknown, false, "inconclusive"}} {
		r := net05CanaryRun(t, "interpretation", tc.validated, tc.cancel)
		if tc.outcome != "" && (r.m == nil || r.m["outcome"] != tc.outcome) {
			t.Fatalf("outcome %v, want %s", r.m, tc.outcome)
		}
		raws = append(raws, r.raw)
	}
	interpretation(t, "NET-05", net05Decoder(t, "interpretation"), raws[0], raws, []string{"completed", "network_not_validated", "incomplete", "inconclusive"}, claims(
		"NET-05 proves an allowlist or identifies its operator", `\bnet-05\b[^.]*\b(proves?|establish\w*|identif\w*|detects?)\b[^.]*\b(allowlist|operat\w*)`,
		"combined results classified as restriction", `\b(combined|correlated)\s+results?\b[^.]*\b(is|are)\s+(classified|a)\b[^.]*\b(broad restriction|allowlist|censorship event)`,
		"profile tests sensitive lists", `\bprofiles?\b[^.]*\btests?\b[^.]*\b(sensitive|political|circumvention|broad public website)`,
		"NET-05 replaces other analysis", `\bnet-05\b[^.]*\breplaces?\b[^.]*\b(outage|captive-portal|routing|reference)`,
		"correlation establishes intent with certainty", `\bcorrelat\w*\b[^.]*\b(establish\w*|proves?)\b[^.]*\b(intent|policy)\b[^.]*\bcertain`,
		"external datasets are ground truth", `\bexternal datasets?\b[^.]*\b(is|are)\s+ground truth`,
	),
		"establishes only the ordered primitive results the inline profile produced on one network generation",
		"does not prove that an allowlist policy exists or identify who operates it", "does not classify the combined results as a broad restriction, allowlist or censorship event",
		"is not for sensitive, political, circumvention or broad public website lists", "does not replace outage, captive-portal, routing or independent reference analysis",
		"is outside net-05 version `0.1.0`", "cannot establish censor intent or policy with certainty", "external datasets are supporting reference signals, not ground truth",
		"simultaneous outage or misconfiguration affecting external profile targets", "shared hidden cdn, dns, routing or certificate dependencies",
		"international transit failure, routing isolation, congestion or packet loss", "captive portal, subscription state, vpn, enterprise policy or local firewall",
		"stale inline profile definitions or changed endpoint behavior", "address-family-specific failure or mobile-network handover")
}

// M-NET-05-PRECONDITIONS: the portable gates, rejection before any profile
// step starts. No anomaly-triggered path exists: NET-05 runs only when issuer
// and deployment both enable composite execution. The traffic-pattern
// criterion is RV-M-NET-05-PRECONDITIONS; the portable part is that profiles
// are bounded to eight declared steps and never overlap (M-NET-05-SAFETY).
// requirement-ref: REQ-NET-05-0031
// requirement-ref: REQ-NET-05-0032
// requirement-ref: REQ-NET-05-0033
// requirement-ref: REQ-NET-05-0034
// requirement-ref: REQ-NET-05-0035
// requirement-ref: REQ-NET-05-0036
func TestMNet05Preconditions(t *testing.T) {
	f := net05SafetyFixture(t)
	preconditionGates(t, f)
	t.Run("untrusted task source", func(t *testing.T) {
		s := net05Serve(t)
		c := s.h.initialized(t)
		defer c.Close()
		data := f.task("untrusted")
		ref, err := s.h.deliveries.InstallEmbedded(context.Background(), "untrusted-issuer", 2, []byte(data))
		if err != nil {
			t.Fatal(err)
		}
		if th, err := c.AdmitTask(data, ref); th != "" || err == nil || s.d.calls.Load()+s.lookups.Load() != 0 {
			t.Fatal("profile from an untrusted issuer admitted")
		}
	})
	t.Run("target fails local evaluation", TestMNet05AtomicTargetAndProxyRules)
	t.Run("not enabled by the deployment", f.notEnabled)
	t.Run("validation context unavailable", TestMNet05PlatformValidationGate)
	t.Run("envelope fit", TestMNet05RegisteredCeilingAdmission)
	t.Run("host policy withdrawn after step 1", func(t *testing.T) {
		s := net05Serve(t)
		entered := make(chan struct{}, 4)
		inner := s.d.dial
		s.d.dial = func(ctx context.Context, n, a string) (measurement.Conn, error) {
			entered <- struct{}{}
			<-ctx.Done()
			return inner(ctx, n, a)
		}
		c := s.h.initialized(t)
		defer c.Close()
		rh, err := c.StartRun(s.h.admit(t, c, f.task("withdrawn")))
		if err != nil {
			t.Fatal(codeOf(err))
		}
		select {
		case <-entered:
		case <-time.After(5 * time.Second):
			t.Fatal("step 2 did not start")
		}
		if s.lookups.Load() != 1 {
			t.Fatal("step 1 did not complete before step 2")
		}
		in := policyInputOf(t, s.h.adapters.policy)
		in.Generation = 3
		in.Permissions.Tests = []task.TestVersion{{ID: "SYN-01", Version: "0.1.0"}}
		in.Permissions.RegistrationIDs, in.Permissions.AllowComposite = nil, false
		next, err := task.NewLocalPolicy(in)
		if err != nil {
			t.Fatal(err)
		}
		if err := c.replacePolicy(next); err != nil {
			t.Fatal(codeOf(err))
		}
		if _, err := c.RunState(rh); err == nil {
			waitTerminal(t, c, rh)
		}
		if s.d.calls.Load() != 1 || len(s.srv.Counts().Requests) != 0 {
			t.Fatalf("steps started after withdrawal: %d dials", s.d.calls.Load())
		}
		records := s.h.run.output.Records()
		if len(records) != 1 {
			t.Fatalf("%d results", len(records))
		}
		if state, reason := executionOf(t, records[0]); state != "incomplete" || reason != "cancelled_by_host" {
			t.Fatalf("withdrawn profile: %s %s", state, reason)
		}
	})
}
