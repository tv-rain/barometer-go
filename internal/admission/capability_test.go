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

// Egress Network Context acquisition is not provided by this build (DEV-578),
// so a valid execution.context requirement is task_unsupported at step 6
// (REQ-TRUST-0036, REQ-TRUST-0037), for embedded and remote origin. The
// harness's unused recording transport must see zero calls: rejection performs
// no context lookup, endpoint call, DNS or measurement activity.
// requirement-ref: REQ-TRUST-0018
// requirement-ref: REQ-TRUST-0036
// requirement-ref: REQ-TRUST-0037
func TestGateRejectsEgressContextRequirements(t *testing.T) {
	cases := []struct {
		name  string
		edits map[string]any
		want  task.AdmissionCode
	}{
		// Positive controls: no requirement depends on egress context.
		{"no network object", nil, 0},
		{"network without context", map[string]any{"execution.network.vpn_policy": "any"}, 0},
		{"empty context", map[string]any{"execution.network.context": map[string]any{}}, 0},
		{"empty required_fields", map[string]any{requiredFields: []any{}}, 0},
		// Each unsupported field individually, then together.
		{"required country_code", map[string]any{requiredFields: []any{"country_code"}}, task.Unsupported},
		{"required region_code", map[string]any{requiredFields: []any{"region_code"}}, task.Unsupported},
		{"required asn", map[string]any{requiredFields: []any{"asn"}}, task.Unsupported},
		{"all required fields", map[string]any{requiredFields: []any{"country_code", "region_code", "asn"}}, task.Unsupported},
		{"country allowlist", map[string]any{allowedCodes: []any{"RU"}}, task.Unsupported},
		{"country allowlist with empty required_fields", map[string]any{requiredFields: []any{}, allowedCodes: []any{"RU"}}, task.Unsupported},
		{"required fields and allowlist", map[string]any{requiredFields: []any{"asn"}, allowedCodes: []any{"RU", "KZ"}}, task.Unsupported},
		// An invalid context is not a valid field requiring a capability.
		{"unknown required field", map[string]any{requiredFields: []any{"city"}}, task.SchemaInvalid},
		{"duplicate required field", map[string]any{requiredFields: []any{"asn", "asn"}}, task.SchemaInvalid},
		{"empty allowlist", map[string]any{allowedCodes: []any{}}, task.SchemaInvalid},
		{"lowercase country code", map[string]any{allowedCodes: []any{"ru"}}, task.SchemaInvalid},
		{"unknown context member", map[string]any{"execution.network.context.endpoint": "https://example.test"}, task.SchemaInvalid},
	}
	for _, c := range cases {
		h := newHarness(t)
		b := taskWith(c.edits)
		_, err := h.embedded(t, b)
		expect(t, "embedded "+c.name, err, c.want)
		_, err = h.remote(t, 0, b)
		expect(t, "remote "+c.name, err, c.want)
		if c.want != 0 && h.normalize.Load() != 0 {
			t.Errorf("%s: codec reached after step 6 failed", c.name)
		}
	}
}

// The egress check is the last check of step 6: earlier step-6 failures keep
// their class, and every later step is not reached (REQ-TRUST-0037).
// requirement-ref: REQ-TRUST-0036
// requirement-ref: REQ-TRUST-0037
func TestGateEgressContextPrecedence(t *testing.T) {
	egress := func(extra map[string]any) []byte {
		edits := map[string]any{requiredFields: []any{"country_code"}}
		for k, v := range extra {
			edits[k] = v
		}
		return taskWith(edits)
	}
	for _, c := range []struct {
		name   string
		policy func(*task.PolicyInput)
		extra  map[string]any
		want   task.AdmissionCode
	}{
		{"unknown field before egress", nil, map[string]any{"private": 1}, task.SchemaInvalid},
		{"task_id length before egress", nil, map[string]any{"task_id": strings.Repeat("a", 129)}, task.LimitExceeded},
		{"expiry before egress", nil, map[string]any{"schedule.not_after": "2026-10-19T11:00:00Z"}, task.Expired},
		{"field limit before egress", nil, map[string]any{"execution.limits.max_duration_ms": 60001}, task.LimitExceeded},
		{"invalid risk factor before egress", nil, map[string]any{"safety.risk_factors": []any{map[string]any{"type": "data_sensitivity", "level": "high"}}}, task.SchemaInvalid},
		{"unsupported risk factor with egress", nil, map[string]any{"safety.risk_factors": []any{map[string]any{"type": "unregistered_factor", "level": "low"}}}, task.Unsupported},
		{"egress before invalid input", nil, map[string]any{"input": map[string]any{}}, task.Unsupported},
		{"egress before test not permitted", func(in *task.PolicyInput) {
			in.Issuers[0].Permissions.Tests = []task.TestVersion{{ID: "SYN-02", Version: "0.1.0"}}
			in.Issuers[0].Permissions.RegistrationIDs = nil
		}, nil, task.Unsupported},
		{"egress before policy rollback", func(in *task.PolicyInput) { in.Generation = 3 }, nil, task.Unsupported},
		{"egress before label namespace", nil, map[string]any{"labels": map[string]any{"private": "a"}}, task.Unsupported},
		{"egress before path not permitted", func(in *task.PolicyInput) { in.Permissions.Paths = task.Paths{Proxy: true} }, nil, task.Unsupported},
		{"egress before unregistered target", func(in *task.PolicyInput) { in.Targets[0].Tuple.Hostname = "unregistered.example.test" }, nil, task.Unsupported},
		{"egress before risk bound", func(in *task.PolicyInput) {
			in.Permissions.MaxRisk, in.Permissions.AllowMediumRisk = task.LowRisk, false
		}, nil, task.Unsupported},
	} {
		in := permissive()
		if c.policy != nil {
			c.policy(&in)
		}
		h := newHarnessWith(t, mustPolicy(t, in))
		_, err := h.embedded(t, egress(c.extra))
		expect(t, c.name, err, c.want)
		if h.normalize.Load() != 0 {
			t.Errorf("%s: codec reached", c.name)
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
// requirement-ref: REQ-TRUST-0037
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
		{"egress before composite", nil, map[string]any{requiredFields: []any{"asn"}}, task.Unsupported},
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
