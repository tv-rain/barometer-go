package admission_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/registry"
)

const (
	requiredFields = "execution.network.context.required_fields"
	allowedCodes   = "execution.network.context.allowed_country_codes"
)

// Egress Network Context acquisition exists since DEV-647, so a valid
// execution.context requirement is admitted for embedded and remote origin
// and evaluated at each attempt; an invalid context is still schema_invalid
// (REQ-EXECUTION-0021). Admission performs no context lookup: the harness's
// recording transport sees zero calls.
// requirement-ref: REQ-TRUST-0018
// requirement-ref: REQ-TRUST-0036
func TestGateAdmitsEgressContextRequirements(t *testing.T) {
	cases := []struct {
		name  string
		edits map[string]any
		want  task.AdmissionCode
	}{
		{"no network object", nil, 0},
		{"network without context", map[string]any{"execution.network.vpn_policy": "any"}, 0},
		{"empty context", map[string]any{"execution.network.context": map[string]any{}}, 0},
		{"empty required_fields", map[string]any{requiredFields: []any{}}, 0},
		{"required country_code", map[string]any{requiredFields: []any{"country_code"}}, 0},
		{"required region_code", map[string]any{requiredFields: []any{"region_code"}}, 0},
		{"required asn", map[string]any{requiredFields: []any{"asn"}}, 0},
		{"all required fields", map[string]any{requiredFields: []any{"country_code", "region_code", "asn"}}, 0},
		{"country allowlist", map[string]any{allowedCodes: []any{"RU"}}, 0},
		{"required fields and allowlist", map[string]any{requiredFields: []any{"asn"}, allowedCodes: []any{"RU", "KZ"}}, 0},
		// An invalid context, and any allowlist other than country, is rejected.
		{"unknown required field", map[string]any{requiredFields: []any{"city"}}, task.SchemaInvalid},
		{"duplicate required field", map[string]any{requiredFields: []any{"asn", "asn"}}, task.SchemaInvalid},
		{"empty allowlist", map[string]any{allowedCodes: []any{}}, task.SchemaInvalid},
		{"lowercase country code", map[string]any{allowedCodes: []any{"ru"}}, task.SchemaInvalid},
		{"unknown context member", map[string]any{"execution.network.context.endpoint": "https://example.test"}, task.SchemaInvalid},
		{"region allowlist", map[string]any{"execution.network.context.allowed_region_codes": []any{"RU-MOW"}}, task.SchemaInvalid},
		{"asn allowlist", map[string]any{"execution.network.context.allowed_asns": []any{64500}}, task.SchemaInvalid},
		{"provider allowlist", map[string]any{"execution.network.context.allowed_providers": []any{"example"}}, task.SchemaInvalid},
		{"operator allowlist", map[string]any{"execution.network.context.allowed_operators": []any{"example"}}, task.SchemaInvalid},
	}
	for _, c := range cases {
		h := newHarness(t)
		b := taskWith(c.edits)
		_, err := h.embedded(t, b)
		expect(t, "embedded "+c.name, err, c.want)
		_, err = h.remote(t, 0, b)
		expect(t, "remote "+c.name, err, c.want)
	}
}

// A context requirement changes no other admission step: each failure keeps
// the class it has without the requirement (REQ-TRUST-0047).
// requirement-ref: REQ-TRUST-0047
func TestGateEgressContextKeepsOtherSteps(t *testing.T) {
	with := func(extra map[string]any, context bool) []byte {
		edits := map[string]any{}
		if context {
			edits[requiredFields] = []any{"country_code"}
		}
		for k, v := range extra {
			edits[k] = v
		}
		return taskWith(edits)
	}
	for _, c := range []struct {
		name   string
		policy func(*task.PolicyInput)
		extra  map[string]any
	}{
		{"unknown field", nil, map[string]any{"private": 1}},
		{"task_id length", nil, map[string]any{"task_id": strings.Repeat("a", 129)}},
		{"expiry", nil, map[string]any{"schedule.not_after": "2026-10-19T11:00:00Z"}},
		{"field limit", nil, map[string]any{"execution.limits.max_duration_ms": 60001}},
		{"invalid input", nil, map[string]any{"input": map[string]any{}}},
		{"test not permitted", func(in *task.PolicyInput) {
			in.Issuers[0].Permissions.Tests = []task.TestVersion{{ID: "SYN-02", Version: "0.1.0"}}
			in.Issuers[0].Permissions.RegistrationIDs = nil
		}, nil},
		{"policy rollback", func(in *task.PolicyInput) { in.Generation = 3 }, nil},
		{"label namespace", nil, map[string]any{"labels": map[string]any{"private": "a"}}},
		{"path not permitted", func(in *task.PolicyInput) { in.Permissions.Paths = task.Paths{Proxy: true} }, nil},
		{"unregistered target", func(in *task.PolicyInput) { in.Targets[0].Tuple.Hostname = "unregistered.example.test" }, nil},
		{"risk bound", func(in *task.PolicyInput) {
			in.Permissions.MaxRisk, in.Permissions.AllowMediumRisk = task.LowRisk, false
		}, nil},
	} {
		var codes [2]error
		for i, context := range []bool{false, true} {
			in := permissive()
			if c.policy != nil {
				c.policy(&in)
			}
			h := newHarnessWith(t, mustPolicy(t, in))
			_, codes[i] = h.embedded(t, with(c.extra, context))
		}
		if codes[0] == nil || code(codes[0]) != code(codes[1]) {
			t.Errorf("%s: without context %v, with context %v", c.name, codes[0], codes[1])
		}
	}
}

// net05Policy enables the synthetic NET-05 entry and composite tasks on both
// sides, then applies edit.
func net05Policy(edit func(*task.PolicyInput)) task.PolicyInput {
	net05 := task.TestVersion{ID: "NET-05", Version: "0.1.0"}
	in := permissive()
	reg := registration("reg-net05")
	reg.Test = net05
	in.Targets = append(in.Targets, reg)
	for _, p := range []*task.Permissions{&in.Permissions, &in.Issuers[0].Permissions} {
		p.Tests, p.AllowComposite = append(p.Tests, net05), true
		p.RegistrationIDs = append(p.RegistrationIDs, reg.ID)
	}
	if edit != nil {
		edit(&in)
	}
	return in
}

// NET-05 is rejected when this build does not register it (step 4), when
// composite execution is not enabled on both sides (step 7), and, because
// composite steps cannot yet be matched against registrations (DEV-454), even
// when it is registered and enabled (step 8). All three reject atomically with
// zero network activity. Registered SYN-01 stays admissible beside it.
// requirement-ref: REQ-TRUST-0018
// requirement-ref: REQ-TRUST-0036
// requirement-ref: REQ-TRUST-0047
func TestGateRejectsCompositeProfiles(t *testing.T) {
	net05 := task.TestVersion{ID: "NET-05", Version: "0.1.0"}
	netTask := func(extra map[string]any) []byte {
		edits := map[string]any{"test_id": "NET-05"}
		for k, v := range extra {
			edits[k] = v
		}
		return taskWith(edits)
	}

	// Unregistered: the default registry holds only SYN-01, so the composite
	// is task_unsupported even under a policy that enables it.
	h := newHarnessWith(t, mustPolicy(t, net05Policy(nil)))
	_, err := h.embedded(t, netTask(nil))
	expect(t, "unregistered NET-05", err, task.Unsupported)
	_, err = h.embedded(t, taskWith(map[string]any{"test_id": "NET-05", "test_version": "0.2.0"}))
	expect(t, "unregistered NET-05 version", err, task.Unsupported)
	if h.normalize.Load() != 0 {
		t.Error("codec reached for an unregistered test")
	}
	_, err = h.embedded(t, []byte(baseTask))
	expect(t, "registered SYN-01 beside NET-05 policy", err, 0)

	for _, c := range []struct {
		name  string
		edit  func(*task.PolicyInput)
		extra map[string]any
		want  task.AdmissionCode
	}{
		{"enabled and registered", nil, nil, task.Unsupported},
		{"policy rollback before composite", func(in *task.PolicyInput) { in.Generation = 3 }, nil, task.PolicyRollback},
		{"issuer composite disabled", func(in *task.PolicyInput) {
			p := &in.Issuers[0].Permissions
			p.AllowComposite = false
			p.Tests = slices.DeleteFunc(p.Tests, func(tv task.TestVersion) bool { return tv == net05 })
			p.RegistrationIDs = []string{"reg-a"}
		}, nil, task.IssuerNotAuthorized},
		{"local composite disabled", func(in *task.PolicyInput) {
			p := &in.Permissions
			p.AllowComposite = false
			p.Tests = slices.DeleteFunc(p.Tests, func(tv task.TestVersion) bool { return tv == net05 })
			p.RegistrationIDs = []string{"reg-a", "reg-b"}
		}, nil, task.IssuerNotAuthorized},
		{"invalid input before composite", nil, map[string]any{"input": map[string]any{}}, task.SchemaInvalid},
		{"egress requirement before composite", nil, map[string]any{requiredFields: []any{"asn"}}, task.Unsupported},
		{"composite before path not permitted", func(in *task.PolicyInput) { in.Permissions.Paths = task.Paths{Proxy: true} }, nil, task.Unsupported},
		{"composite before unregistered target", func(in *task.PolicyInput) { in.Targets[2].Tuple.Hostname = "unregistered.example.test" }, nil, task.Unsupported},
		{"composite before risk bound", func(in *task.PolicyInput) {
			in.Permissions.MaxRisk, in.Permissions.AllowMediumRisk = task.LowRisk, false
		}, nil, task.Unsupported},
	} {
		h := newHarness(t)
		rebind(t, h, mustPolicy(t, net05Policy(c.edit)), synEntry(h, task.MediumRisk),
			registry.Entry{Test: net05, Codec: synCodec{calls: h.normalize, id: "NET-05"}, BaselineRisk: task.MediumRisk, Targets: net05Targets, Proxy: registry.NoProxy{}, Composite: true})
		_, err := h.embedded(t, netTask(c.extra))
		expect(t, c.name, err, c.want)
	}
}
