package barometer

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"

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
	r, e := productionRegistry(nil)
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
	r, _ := productionRegistry(nil)
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

// Admission coverage; profile history and summed maxima are approved
// deferrals. Runner evidence is separate in net05_runner_test.go.
// requirement-ref: REQ-NET-05-0014
// requirement-ref: REQ-NET-05-0015
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
func TestMNet05Admission(t *testing.T) {
	steps := primitiveSteps()
	for _, n := range []int{1, 4, 8} {
		t.Run(fmt.Sprint("steps-", n), func(t *testing.T) {
			h, c, d := net05AdmissionClient(t, nil)
			var selected []string
			for i := 0; i < n; i++ {
				selected = append(selected, strings.Replace(steps[i%4], `"step_id":"`+[]string{"dns", "tcp", "tls", "http"}[i%4]+`"`, fmt.Sprintf(`"step_id":"s%d"`, i), 1))
			}
			data := net05Task("valid-profile", profileInput(selected...), `{}`, "")
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
	for _, n := range []int{49151, 49152, 49153} {
		profile := string(parsed["profile"])
		profile = profile[:len(profile)-1] + strings.Repeat(" ", n-len(profile)) + "}"
		input := `{"profile":` + profile + `}`
		if n <= 49152 {
			normalizedProfile(t, input)
		}
		// Common input has the same byte ceiling and includes the wrapper;
		// it can reject before the profile-specific cap is reached.
		assertNet05Rejected(t, input, `{}`, "", "task_limit_exceeded", nil)
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
	r, _ := productionRegistry(nil)
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
		original, _ := productionRegistry(nil)
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
