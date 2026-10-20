package admission_test

import (
	"fmt"
	"testing"

	"github.com/tv-rain/barometer-go/internal/admission"
	"github.com/tv-rain/barometer-go/internal/model/task"
)

var riskNames = map[task.Risk]string{task.LowRisk: "low", task.MediumRisk: "medium", task.HighRisk: "high", task.UnknownRisk: "unknown"}

// declaring returns task edits declaring the level with a consistent factor.
func declaring(level task.Risk) map[string]any {
	if level == task.LowRisk {
		return map[string]any{}
	}
	name := riskNames[level]
	return map[string]any{"safety.risk_level": name, "safety.risk_factors": []any{map[string]any{"type": "target_sensitivity", "level": name}}}
}

// mostRestrictive is the expected effective risk, computed independently of
// the implementation: unknown dominates, otherwise the highest ordered level.
func mostRestrictive(levels ...task.Risk) task.Risk {
	out := task.LowRisk
	for _, l := range levels {
		if l == task.UnknownRisk {
			return task.UnknownRisk
		}
		if l > out {
			out = l
		}
	}
	return out
}

// riskPolicy is the permissive policy with the matched registration's
// reviewed baseline set and the given unknown-risk exceptions.
func riskPolicy(registered task.Risk, issuerUnknown, localUnknown bool) task.PolicyInput {
	in := permissive()
	in.Targets[0].BaselineRisk = registered
	in.Issuers[0].Permissions.AllowUnknownRisk = issuerUnknown
	in.Permissions.AllowUnknownRisk = localUnknown
	return in
}

// Effective risk is the most restrictive of the registered baseline, the
// task-declared level and the locally determined risk of the matched target
// registration. Under a policy that permits every level, each combination is
// admitted with exactly that effective risk, never below the baseline.
// requirement-ref: REQ-SAFETY-0004
// requirement-ref: REQ-SAFETY-0005
// requirement-ref: REQ-SAFETY-0023
// requirement-ref: REQ-SAFETY-0024
// requirement-ref: REQ-SAFETY-0025
// requirement-ref: REQ-SAFETY-0026
func TestGateEffectiveRisk(t *testing.T) {
	levels := []task.Risk{task.LowRisk, task.MediumRisk, task.HighRisk, task.UnknownRisk}
	for _, baseline := range []task.Risk{task.LowRisk, task.MediumRisk, task.UnknownRisk} {
		for _, declared := range levels {
			for _, registered := range levels {
				name := fmt.Sprintf("baseline %s declared %s registered %s", riskNames[baseline], riskNames[declared], riskNames[registered])
				h := newHarness(t)
				rebind(t, h, mustPolicy(t, riskPolicy(registered, true, true)), synEntry(h, baseline))
				g, err := h.embedded(t, taskWith(declaring(declared)))
				expect(t, name, err, 0)
				want := mostRestrictive(baseline, declared, registered)
				got := admission.GatedEffectiveRisk(g)
				if got != want {
					t.Errorf("%s: effective risk %s, want %s", name, riskNames[got], riskNames[want])
				}
				if got != task.UnknownRisk && (baseline == task.UnknownRisk || got < baseline) {
					t.Errorf("%s: effective risk below the registered baseline", name)
				}
			}
		}
	}
}

// A task cannot lower the registered baseline: a low declaration on a medium
// baseline is evaluated as medium, so a policy that permits only low rejects
// it, and an unknown baseline stays unknown whatever the task declares.
// requirement-ref: REQ-SAFETY-0024
// requirement-ref: REQ-SAFETY-0025
// requirement-ref: REQ-SAFETY-0010
func TestGateTaskCannotLowerBaseline(t *testing.T) {
	for _, side := range []string{"issuer", "local"} {
		h := newHarness(t)
		rebind(t, h, mustPolicy(t, narrowed(side, func(p *task.Permissions) { p.MaxRisk, p.AllowMediumRisk = task.LowRisk, false })), synEntry(h, task.MediumRisk))
		_, err := h.embedded(t, []byte(baseTask))
		expect(t, side+" low declaration on a medium baseline under a low maximum", err, task.IssuerNotAuthorized)

		h = newHarness(t)
		rebind(t, h, mustPolicy(t, riskPolicy(task.LowRisk, side == "issuer", side == "local")), synEntry(h, task.UnknownRisk))
		_, err = h.embedded(t, []byte(baseTask))
		expect(t, side+" low declaration on an unknown baseline with a one-sided exception", err, task.IssuerNotAuthorized)
	}
	h := newHarness(t)
	rebind(t, h, mustPolicy(t, riskPolicy(task.LowRisk, true, true)), synEntry(h, task.UnknownRisk))
	g, err := h.embedded(t, []byte(baseTask))
	if expect(t, "low declaration on an unknown baseline with the exception", err, 0); admission.GatedEffectiveRisk(g) != task.UnknownRisk {
		t.Error("unknown baseline lowered by the task declaration")
	}
}

// Locally determined risk only raises effective risk, and the raised level is
// then held to the same bounds as the baseline and declaration: the lower
// maximum, medium-risk enablement and the unknown-risk exception, each
// narrowed in the issuer and then in the deployment.
// requirement-ref: REQ-SAFETY-0010
// requirement-ref: REQ-SAFETY-0023
// requirement-ref: REQ-SAFETY-0027
// requirement-ref: REQ-TRUST-0011
func TestGateLocallyDeterminedRiskBounds(t *testing.T) {
	cases := []struct {
		name       string
		registered task.Risk
		edit       func(*task.Permissions)
		want       task.AdmissionCode
	}{
		{"medium registration under a low maximum", task.MediumRisk, func(p *task.Permissions) { p.MaxRisk, p.AllowMediumRisk = task.LowRisk, false }, task.IssuerNotAuthorized},
		{"low registration under a low maximum", task.LowRisk, func(p *task.Permissions) { p.MaxRisk, p.AllowMediumRisk = task.LowRisk, false }, 0},
		{"medium registration without medium enablement", task.MediumRisk, func(p *task.Permissions) { p.MaxRisk, p.AllowMediumRisk = task.MediumRisk, false }, task.IssuerNotAuthorized},
		{"medium registration with medium enablement", task.MediumRisk, func(p *task.Permissions) { p.MaxRisk = task.MediumRisk }, 0},
		{"high registration under a medium maximum", task.HighRisk, func(p *task.Permissions) { p.MaxRisk = task.MediumRisk }, task.IssuerNotAuthorized},
		{"high registration without medium enablement", task.HighRisk, func(p *task.Permissions) { p.AllowMediumRisk = false }, task.IssuerNotAuthorized},
		{"high registration under a high maximum", task.HighRisk, func(*task.Permissions) {}, 0},
		{"unknown registration without the exception", task.UnknownRisk, func(*task.Permissions) {}, task.IssuerNotAuthorized},
	}
	for _, c := range cases {
		for _, side := range []string{"issuer", "local"} {
			in := narrowed(side, c.edit)
			in.Targets[0].BaselineRisk = c.registered
			h := newHarness(t)
			rebind(t, h, mustPolicy(t, in), synEntry(h, task.LowRisk))
			g, err := h.embedded(t, []byte(baseTask))
			expect(t, side+" "+c.name, err, c.want)
			if c.want == 0 && admission.GatedEffectiveRisk(g) != c.registered {
				t.Errorf("%s %s: locally determined risk not carried", side, c.name)
			}
		}
	}
	// Unknown locally determined risk needs the reviewed exception on both sides.
	for _, ex := range []struct{ issuer, local bool }{{false, false}, {true, false}, {false, true}, {true, true}} {
		h := newHarness(t)
		rebind(t, h, mustPolicy(t, riskPolicy(task.UnknownRisk, ex.issuer, ex.local)), synEntry(h, task.LowRisk))
		_, err := h.embedded(t, []byte(baseTask))
		want := task.IssuerNotAuthorized
		if ex.issuer && ex.local {
			want = 0
		}
		expect(t, fmt.Sprintf("unknown registration with exceptions %v", ex), err, want)
	}
}

// Risk declarations that cannot be evaluated reliably fail closed before step
// 9: an unrecognized factor type is unsupported even where unknown risk is
// permitted, and an unknown factor under a known overall level is invalid,
// because no trusted local assessment exists to resolve it. A target failure
// at step 8 takes precedence over a risk failure at step 9.
// requirement-ref: REQ-SAFETY-0017
// requirement-ref: REQ-SAFETY-0019
// requirement-ref: REQ-SAFETY-0029
// requirement-ref: REQ-TRUST-0037
func TestGateUnreliableSafetyInputsFailClosed(t *testing.T) {
	permitsUnknown := mustPolicy(t, riskPolicy(task.LowRisk, true, true))
	for _, typ := range []string{"battery_drain", "bandwidth", "Network_observability", "network-observability", ""} {
		_, err := newHarnessWith(t, permitsUnknown).embedded(t, taskWith(map[string]any{"safety.risk_level": "unknown",
			"safety.risk_factors": []any{map[string]any{"type": typ, "level": "unknown"}}}))
		expect(t, "unrecognized factor type "+typ, err, task.Unsupported)
	}
	for _, level := range []string{"low", "medium", "high"} {
		_, err := newHarnessWith(t, permitsUnknown).embedded(t, taskWith(map[string]any{"safety.risk_level": level,
			"safety.risk_factors": []any{map[string]any{"type": "target_sensitivity", "level": "unknown"}}}))
		expect(t, "unknown factor under overall "+level, err, task.SchemaInvalid)
	}
	in := permissive()
	in.Targets[0].Tuple.Hostname = "other.example.test"
	in.Targets[1].Tuple.Hostname = "other-b.example.test"
	_, err := newHarnessWith(t, mustPolicy(t, in)).embedded(t, taskWith(declaring(task.UnknownRisk)))
	expect(t, "unregistered target before unknown risk", err, task.TargetNotRegistered)
}

// The consent requirement is required when the task, the issuer or the
// deployment requires it. A task value of false adds no requirement and
// never clears an issuer or local one; consent validity is not checked here.
// requirement-ref: REQ-SAFETY-0021
// requirement-ref: REQ-SAFETY-0022
// requirement-ref: REQ-SAFETY-0028
func TestGateConsentRequirement(t *testing.T) {
	for _, taskRequires := range []bool{false, true} {
		for _, issuer := range []bool{false, true} {
			for _, local := range []bool{false, true} {
				in := permissive()
				in.Issuers[0].Permissions.RequireConsent, in.Permissions.RequireConsent = issuer, local
				name := fmt.Sprintf("task %v issuer %v local %v", taskRequires, issuer, local)
				g, err := newHarnessWith(t, mustPolicy(t, in)).embedded(t, taskWith(map[string]any{"safety.requires_user_consent": taskRequires}))
				expect(t, name, err, 0)
				if got := admission.GatedRequiresConsent(g); got != (taskRequires || issuer || local) {
					t.Errorf("%s: consent required %v", name, got)
				}
			}
		}
	}
	// A low-risk task without a consent requirement gains no permission over
	// the same task at medium risk with consent required.
	in := permissive()
	in.Permissions.MaxRisk, in.Permissions.AllowMediumRisk = task.LowRisk, false
	in.Permissions.RequireConsent = true
	h := newHarnessWith(t, mustPolicy(t, in))
	_, err := h.embedded(t, []byte(baseTask))
	expect(t, "low declaration on a medium baseline", err, task.IssuerNotAuthorized)
	_, err = h.embedded(t, taskWith(map[string]any{"safety.requires_user_consent": true}))
	expect(t, "consent declaration does not relax risk", err, task.IssuerNotAuthorized)
}
