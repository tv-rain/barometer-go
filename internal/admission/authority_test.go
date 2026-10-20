package admission_test

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/admission"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/registry"
)

// registeredMaxima are the public registered ceilings; the permissive policy
// grants exactly these, so only the registered limits bind.
func registeredMaxima() task.PolicyLimits {
	return task.PolicyLimits{
		Attempt: task.AttemptLimits{DurationMS: 60000, BytesSent: 65536, BytesReceived: 262144, Connections: 12},
		MaxRuns: 28, MaxAttempts: 2, MaxValiditySeconds: 604800,
		MinRepeatIntervalSeconds: 21600, MinRepeatJitterSeconds: 300,
		MinRetryDelayMS: 30000, MaxRetryDelayMS: 600000, MaxRetryJitterMS: 60000,
		MaxResultBytes: 65536, MaxBundleBytes: 262144, MaxResultsPerBundle: 16,
		MaxPendingBundles: 32, MaxPendingBytes: 4194304, MaxRetentionSeconds: 86400,
	}
}

func permissions() task.Permissions {
	return task.Permissions{
		Tests:           []task.TestVersion{{ID: "SYN-01", Version: "0.1.0"}},
		RegistrationIDs: []string{"reg-a"},
		Paths:           task.Paths{Direct: true, Proxy: true},
		MaxRisk:         task.HighRisk,
		AllowMediumRisk: true,
		LabelNamespaces: []string{"t", "cohort", "network", "note", "campaign"},
		Limits:          registeredMaxima(),
	}
}

func registration(id string) task.TargetRegistration {
	return task.TargetRegistration{ID: id, Test: task.TestVersion{ID: "SYN-01", Version: "0.1.0"}, Role: task.MeasurementTarget,
		Tuple: task.TargetTuple{Hostname: id + ".example.test"}, Paths: task.Paths{Direct: true, Proxy: true}, BaselineRisk: task.LowRisk,
		Owner: "synthetic", Purpose: "synthetic", ReviewGeneration: 1, ExpiresAt: task.TimestampFromTime(epoch.Add(48 * time.Hour))}
}

// permissive is a synthetic generation-2 policy in which deployment and both
// test issuers allow everything the registered limits allow for SYN-01.
func permissive() task.PolicyInput {
	local := permissions()
	local.RegistrationIDs = []string{"reg-a", "reg-b"}
	return task.PolicyInput{Generation: 2, Permissions: local,
		Issuers: []task.IssuerPermissions{{IssuerID: "embedded-issuer", Permissions: permissions()}, {IssuerID: "remote-issuer", Permissions: permissions()}},
		Targets: []task.TargetRegistration{registration("reg-a"), registration("reg-b")},
		Keys:    []task.TrustedKey{signer().Key(keyNotBefore, keyNotAfter), otherSigner().Key(keyNotBefore, keyNotAfter)}}
}

func mustPolicy(t testing.TB, in task.PolicyInput) task.LocalPolicy {
	t.Helper()
	p, err := task.NewLocalPolicy(in)
	if err != nil {
		t.Fatal("synthetic policy rejected")
	}
	return p
}

// narrowed returns the permissive policy with one permission set edited: the
// embedded issuer's ("issuer") or the deployment's ("local").
func narrowed(side string, edit func(*task.Permissions)) task.PolicyInput {
	in := permissive()
	if side == "issuer" {
		edit(&in.Issuers[0].Permissions)
	} else {
		edit(&in.Permissions)
	}
	return in
}

// rebind replaces the harness admitter with one over the given registry and
// policy, keeping its store, clock and unused transport.
func rebind(t *testing.T, h *harness, policy task.LocalPolicy, entries ...registry.Entry) {
	t.Helper()
	reg, err := registry.New(entries)
	if err != nil {
		t.Fatal("registry rejected")
	}
	if h.admitter, err = admission.NewAdmitter(h.store, h.clock, reg, policy, h.replay); err != nil {
		t.Fatal("admitter rejected")
	}
}

func synEntry(h *harness, baseline task.Risk) registry.Entry {
	return registry.Entry{Test: task.TestVersion{ID: "SYN-01", Version: "0.1.0"}, Codec: synCodec{calls: h.normalize}, BaselineRisk: baseline, Targets: synTargets, Proxy: registry.NoProxy{}}
}

func retry(initial, maxDelay, jitter int) map[string]any {
	return map[string]any{"max_attempts": 2, "retry_on": []string{"platform_suspended"},
		"backoff": map[string]any{"mode": "fixed", "initial_delay_ms": initial, "max_delay_ms": maxDelay, "jitter_ms": jitter}}
}

func repeat(interval, jitter int) map[string]any {
	return map[string]any{"mode": "interval", "interval": interval, "jitter": jitter}
}

func notAfter(seconds int) string {
	return epoch.Add(time.Duration(seconds) * time.Second).Format(time.RFC3339)
}

// Every numeric bound is enforced per issuer and per deployment policy: a task
// value at the effective bound or below it is accepted and carried as the
// effective value, one unit beyond is rejected and never clamped.
// requirement-ref: REQ-TRUST-0010
// requirement-ref: REQ-TRUST-0011
// requirement-ref: REQ-TRUST-0012
// requirement-ref: REQ-LIMIT-0030
func TestGateIssuerAndLocalCeilings(t *testing.T) {
	type row struct {
		task  map[string]any
		want  task.AdmissionCode
		check func(task.PolicyLimits) bool
	}
	cases := []struct {
		name string
		edit func(*task.Permissions)
		rows []row
	}{
		{"attempt duration", func(p *task.Permissions) { p.Limits.Attempt.DurationMS = 40000 }, []row{
			{map[string]any{"execution.limits.max_duration_ms": 39999}, 0, func(l task.PolicyLimits) bool { return l.Attempt.DurationMS == 39999 }},
			{map[string]any{"execution.limits.max_duration_ms": 40000}, 0, func(l task.PolicyLimits) bool { return l.Attempt.DurationMS == 40000 }},
			{map[string]any{"execution.limits.max_duration_ms": 40001}, task.LimitExceeded, nil}}},
		{"bytes sent", func(p *task.Permissions) { p.Limits.Attempt.BytesSent = 40000 }, []row{
			{map[string]any{"execution.limits.max_bytes_sent": 39999}, 0, func(l task.PolicyLimits) bool { return l.Attempt.BytesSent == 39999 }},
			{map[string]any{"execution.limits.max_bytes_sent": 40000}, 0, nil},
			{map[string]any{"execution.limits.max_bytes_sent": 40001}, task.LimitExceeded, nil}}},
		{"bytes received", func(p *task.Permissions) { p.Limits.Attempt.BytesReceived = 140000 }, []row{
			{map[string]any{"execution.limits.max_bytes_received": 139999}, 0, func(l task.PolicyLimits) bool { return l.Attempt.BytesReceived == 139999 }},
			{map[string]any{"execution.limits.max_bytes_received": 140000}, 0, nil},
			{map[string]any{"execution.limits.max_bytes_received": 140001}, task.LimitExceeded, nil}}},
		{"connections", func(p *task.Permissions) { p.Limits.Attempt.Connections = 9 }, []row{
			{map[string]any{"execution.limits.max_connections": 8}, 0, func(l task.PolicyLimits) bool { return l.Attempt.Connections == 8 }},
			{map[string]any{"execution.limits.max_connections": 9}, 0, nil},
			{map[string]any{"execution.limits.max_connections": 10}, task.LimitExceeded, nil}}},
		{"runs", func(p *task.Permissions) { p.Limits.MaxRuns = 5 }, []row{
			{map[string]any{"schedule.limits.max_runs": 4}, 0, func(l task.PolicyLimits) bool { return l.MaxRuns == 4 }},
			{map[string]any{"schedule.limits.max_runs": 5}, 0, nil},
			{map[string]any{"schedule.limits.max_runs": 6}, task.LimitExceeded, nil}}},
		{"attempts", func(p *task.Permissions) { p.Limits.MaxAttempts = 1 }, []row{
			{map[string]any{}, 0, func(l task.PolicyLimits) bool { return l.MaxAttempts == 1 }},
			{map[string]any{"schedule.retry": retry(60000, 60000, 0)}, task.LimitExceeded, nil}}},
		{"validity", func(p *task.Permissions) { p.Limits.MaxValiditySeconds = 3600 }, []row{
			{map[string]any{"schedule.not_after": notAfter(3599)}, 0, nil},
			{map[string]any{"schedule.not_after": notAfter(3600)}, 0, nil},
			{map[string]any{"schedule.not_after": notAfter(3601)}, task.LimitExceeded, nil}}},
		{"repeat interval minimum", func(p *task.Permissions) { p.Limits.MinRepeatIntervalSeconds = 43200 }, []row{
			{map[string]any{"schedule.repeat": repeat(43201, 600), "schedule.limits.max_runs": 2}, 0, func(l task.PolicyLimits) bool { return l.MinRepeatIntervalSeconds == 43201 }},
			{map[string]any{"schedule.repeat": repeat(43200, 600), "schedule.limits.max_runs": 2}, 0, nil},
			{map[string]any{"schedule.repeat": repeat(43199, 600), "schedule.limits.max_runs": 2}, task.LimitExceeded, nil}}},
		{"repeat jitter minimum", func(p *task.Permissions) { p.Limits.MinRepeatJitterSeconds = 600 }, []row{
			{map[string]any{"schedule.repeat": repeat(43200, 601), "schedule.limits.max_runs": 2}, 0, func(l task.PolicyLimits) bool { return l.MinRepeatJitterSeconds == 601 }},
			{map[string]any{"schedule.repeat": repeat(43200, 600), "schedule.limits.max_runs": 2}, 0, nil},
			{map[string]any{"schedule.repeat": repeat(43200, 599), "schedule.limits.max_runs": 2}, task.LimitExceeded, nil}}},
		{"retry delay minimum", func(p *task.Permissions) { p.Limits.MinRetryDelayMS = 60000 }, []row{
			{map[string]any{"schedule.retry": retry(60001, 60001, 0)}, 0, func(l task.PolicyLimits) bool { return l.MinRetryDelayMS == 60001 }},
			{map[string]any{"schedule.retry": retry(60000, 60000, 0)}, 0, nil},
			{map[string]any{"schedule.retry": retry(59999, 60000, 0)}, task.LimitExceeded, nil}}},
		{"retry delay maximum", func(p *task.Permissions) { p.Limits.MaxRetryDelayMS = 300000 }, []row{
			{map[string]any{"schedule.retry": retry(60000, 299999, 0)}, 0, func(l task.PolicyLimits) bool { return l.MaxRetryDelayMS == 299999 }},
			{map[string]any{"schedule.retry": retry(60000, 300000, 0)}, 0, nil},
			{map[string]any{"schedule.retry": retry(60000, 300001, 0)}, task.LimitExceeded, nil}}},
		{"retry jitter maximum", func(p *task.Permissions) { p.Limits.MaxRetryJitterMS = 1000 }, []row{
			{map[string]any{"schedule.retry": retry(60000, 60000, 999)}, 0, func(l task.PolicyLimits) bool { return l.MaxRetryJitterMS == 999 }},
			{map[string]any{"schedule.retry": retry(60000, 60000, 1000)}, 0, nil},
			{map[string]any{"schedule.retry": retry(60000, 60000, 1001)}, task.LimitExceeded, nil}}},
	}
	for _, c := range cases {
		for _, side := range []string{"issuer", "local"} {
			for i, r := range c.rows {
				name := fmt.Sprintf("%s %s row %d", side, c.name, i)
				h := newHarnessWith(t, mustPolicy(t, narrowed(side, c.edit)))
				g, err := h.embedded(t, taskWith(r.task))
				expect(t, name, err, r.want)
				if r.want != 0 && h.normalize.Load() != 0 {
					t.Errorf("%s: codec reached after step 7 failed", name)
				}
				if err == nil && r.check != nil && !r.check(admission.GatedLimits(g)) {
					t.Errorf("%s: lower task value not carried as the effective value", name)
				}
			}
		}
	}
	// Ceilings without a task field are carried, intersected, for result and
	// storage enforcement.
	for _, swap := range []bool{false, true} {
		in := permissive()
		a, b := &in.Issuers[0].Permissions.Limits, &in.Permissions.Limits
		if swap {
			a, b = b, a
		}
		a.MaxResultBytes, b.MaxRetentionSeconds = 4096, 3600
		g, err := newHarnessWith(t, mustPolicy(t, in)).embedded(t, []byte(baseTask))
		if l := admission.GatedLimits(g); err != nil || l.MaxResultBytes != 4096 || l.MaxRetentionSeconds != 3600 || l.MaxBundleBytes != 262144 {
			t.Errorf("result and retention ceilings not intersected (swapped: %v)", swap)
		}
	}
}

// When issuer and deployment disagree, the stricter value binds in both
// directions; neither side widens the other.
// requirement-ref: REQ-TRUST-0010
func TestGateIssuerLocalConflicts(t *testing.T) {
	for _, c := range []struct{ issuer, local int }{{50000, 40000}, {40000, 50000}} {
		in := permissive()
		in.Issuers[0].Permissions.Limits.Attempt.DurationMS = c.issuer
		in.Permissions.Limits.Attempt.DurationMS = c.local
		h := newHarnessWith(t, mustPolicy(t, in))
		_, err := h.embedded(t, taskWith(map[string]any{"execution.limits.max_duration_ms": 40000}))
		expect(t, fmt.Sprint("duration at stricter bound ", c), err, 0)
		_, err = h.embedded(t, taskWith(map[string]any{"execution.limits.max_duration_ms": 40001}))
		expect(t, fmt.Sprint("duration above stricter bound ", c), err, task.LimitExceeded)
	}
	// Paths: issuer direct only, local proxy only permits neither.
	in := permissive()
	in.Issuers[0].Permissions.Paths = task.Paths{Direct: true}
	in.Permissions.Paths = task.Paths{Proxy: true}
	h := newHarnessWith(t, mustPolicy(t, in))
	for _, path := range []string{"direct", "proxy"} {
		_, err := h.embedded(t, taskWith(map[string]any{"execution.network_path": path}))
		expect(t, "disjoint paths "+path, err, task.IssuerNotAuthorized)
	}
	// Registration references and label namespaces intersect; consent is
	// required when any side requires it.
	in = permissive()
	in.Issuers[0].Permissions.RequireConsent = true
	g, err := newHarnessWith(t, mustPolicy(t, in)).embedded(t, []byte(baseTask))
	if err != nil || !slices.Equal(admission.GatedRegistrations(g), []string{"reg-a"}) || !admission.GatedRequiresConsent(g) || admission.GatedPolicyGeneration(g) != 2 {
		t.Error("registrations, consent or generation not intersected")
	}
	g, err = newHarness(t).embedded(t, taskWith(map[string]any{"safety.requires_user_consent": true}))
	if err != nil || !admission.GatedRequiresConsent(g) {
		t.Error("task consent requirement dropped")
	}
	g, err = newHarness(t).embedded(t, []byte(baseTask))
	if err != nil || admission.GatedRequiresConsent(g) || admission.GatedPaths(g) != (task.Paths{Direct: true}) {
		t.Error("effective consent or path not narrowed to the task")
	}
	if fmt.Sprintf("%v %+v %#v", admission.GatedAuthority(g), admission.GatedAuthority(g), admission.GatedAuthority(g)) != "admission.Authority admission.Authority admission.Authority" {
		t.Error("authority formatting exposes values")
	}
}

// Each non-numeric authority dimension is narrowed in the issuer and then in
// the deployment policy while the other side stays permissive.
// requirement-ref: REQ-TRUST-0010
// requirement-ref: REQ-TRUST-0011
// requirement-ref: REQ-TRUST-0047
func TestGateIssuerAuthorization(t *testing.T) {
	net05 := task.TestVersion{ID: "NET-05", Version: "0.1.0"}
	high := []any{map[string]any{"type": "data_sensitivity", "level": "high"}}
	unknown := []any{map[string]any{"type": "target_sensitivity", "level": "unknown"}}
	cases := []struct {
		name     string
		edit     func(*task.Permissions)
		task     map[string]any
		baseline task.Risk
		want     task.AdmissionCode
	}{
		{"permissive", func(*task.Permissions) {}, nil, task.MediumRisk, 0},
		{"test/version not permitted", func(p *task.Permissions) {
			p.Tests = []task.TestVersion{{ID: "SYN-01", Version: "0.2.0"}}
			p.RegistrationIDs = nil
		}, nil, task.MediumRisk, task.IssuerNotAuthorized},
		{"other test only", func(p *task.Permissions) {
			p.Tests = []task.TestVersion{{ID: "SYN-02", Version: "0.1.0"}}
			p.RegistrationIDs = nil
		}, nil, task.MediumRisk, task.IssuerNotAuthorized},
		{"proxy only for a direct task", func(p *task.Permissions) { p.Paths = task.Paths{Proxy: true} }, nil, task.MediumRisk, task.IssuerNotAuthorized},
		{"direct only for a proxy task", func(p *task.Permissions) { p.Paths = task.Paths{Direct: true} }, map[string]any{"execution.network_path": "proxy"}, task.MediumRisk, task.IssuerNotAuthorized},
		{"proxy permitted for a proxy task", func(p *task.Permissions) { p.Paths = task.Paths{Proxy: true} }, map[string]any{"execution.network_path": "proxy"}, task.MediumRisk, 0},
		{"maximum risk low for a medium baseline", func(p *task.Permissions) { p.MaxRisk, p.AllowMediumRisk = task.LowRisk, false }, nil, task.MediumRisk, task.IssuerNotAuthorized},
		{"maximum risk low for a low task", func(p *task.Permissions) { p.MaxRisk, p.AllowMediumRisk = task.LowRisk, false }, nil, task.LowRisk, 0},
		{"maximum risk low for a medium task", func(p *task.Permissions) { p.MaxRisk, p.AllowMediumRisk = task.LowRisk, false },
			map[string]any{"safety.risk_level": "medium", "safety.risk_factors": []any{map[string]any{"type": "data_sensitivity", "level": "medium"}}}, task.LowRisk, task.IssuerNotAuthorized},
		{"medium-risk enablement off", func(p *task.Permissions) { p.MaxRisk, p.AllowMediumRisk = task.MediumRisk, false }, nil, task.MediumRisk, task.IssuerNotAuthorized},
		{"maximum risk medium for a high task", func(p *task.Permissions) { p.MaxRisk = task.MediumRisk }, map[string]any{"safety.risk_level": "high", "safety.risk_factors": high}, task.MediumRisk, task.IssuerNotAuthorized},
		{"maximum risk high for a high task", func(*task.Permissions) {}, map[string]any{"safety.risk_level": "high", "safety.risk_factors": high}, task.MediumRisk, 0},
		{"medium enablement off for a high task", func(p *task.Permissions) { p.AllowMediumRisk = false }, map[string]any{"safety.risk_level": "high", "safety.risk_factors": high}, task.LowRisk, task.IssuerNotAuthorized},
		{"unknown risk without exception", func(*task.Permissions) {}, map[string]any{"safety.risk_level": "unknown", "safety.risk_factors": unknown}, task.MediumRisk, task.IssuerNotAuthorized},
		{"unknown baseline without exception", func(*task.Permissions) {}, nil, task.UnknownRisk, task.IssuerNotAuthorized},
		{"label namespace not permitted", func(p *task.Permissions) { p.LabelNamespaces = []string{"campaign"} }, map[string]any{"labels": map[string]any{"cohort": "a"}}, task.MediumRisk, task.IssuerNotAuthorized},
		{"no namespaces and a label", func(p *task.Permissions) { p.LabelNamespaces = nil }, map[string]any{"labels": map[string]any{"campaign": "a"}}, task.MediumRisk, task.IssuerNotAuthorized},
		{"no namespaces and no labels", func(p *task.Permissions) { p.LabelNamespaces = nil }, nil, task.MediumRisk, 0},
		{"label below namespace at a dot", func(p *task.Permissions) { p.LabelNamespaces = []string{"campaign"} }, map[string]any{"labels": map[string]any{"campaign": "a", "campaign.wave": "b"}}, task.MediumRisk, 0},
		{"label sharing a namespace prefix", func(p *task.Permissions) { p.LabelNamespaces = []string{"campaign"} }, map[string]any{"labels": map[string]any{"campaignx": "a"}}, task.MediumRisk, task.IssuerNotAuthorized},
		{"label above a dotted namespace", func(p *task.Permissions) { p.LabelNamespaces = []string{"campaign.wave"} }, map[string]any{"labels": map[string]any{"campaign": "a"}}, task.MediumRisk, task.IssuerNotAuthorized},
		{"composite disabled", func(p *task.Permissions) { p.AllowComposite = false }, map[string]any{"test_id": "NET-05"}, task.MediumRisk, task.IssuerNotAuthorized},
		// Enablement passes step 7, but composite steps cannot be matched yet,
		// so step 8 fails closed (DEV-590).
		{"composite enabled", func(*task.Permissions) {}, map[string]any{"test_id": "NET-05"}, task.MediumRisk, task.Unsupported},
	}
	for _, c := range cases {
		for _, side := range []string{"issuer", "local"} {
			in := permissive()
			// NET-05 needs explicit composite enablement on both sides; the
			// "composite disabled" row then removes it from one side.
			reg := registration("reg-net05")
			reg.Test = net05
			in.Targets = append(in.Targets, reg)
			for _, p := range []*task.Permissions{&in.Permissions, &in.Issuers[0].Permissions} {
				p.Tests, p.AllowComposite = append(p.Tests, net05), true
				p.RegistrationIDs = append(p.RegistrationIDs, reg.ID)
			}
			if side == "issuer" {
				c.edit(&in.Issuers[0].Permissions)
			} else {
				c.edit(&in.Permissions)
			}
			for _, p := range []*task.Permissions{&in.Permissions, &in.Issuers[0].Permissions} {
				if !p.AllowComposite {
					p.Tests = slices.DeleteFunc(p.Tests, func(tv task.TestVersion) bool { return tv == net05 })
					p.RegistrationIDs = slices.DeleteFunc(p.RegistrationIDs, func(id string) bool { return id == reg.ID })
				}
			}
			h := newHarness(t)
			rebind(t, h, mustPolicy(t, in), synEntry(h, c.baseline),
				registry.Entry{Test: net05, Codec: synCodec{calls: h.normalize, id: "NET-05"}, BaselineRisk: c.baseline, Targets: net05Targets, Proxy: registry.NoProxy{}, Composite: true})
			_, err := h.embedded(t, taskWith(c.task))
			expect(t, side+" "+c.name, err, c.want)
		}
	}
	// Unknown risk needs the reviewed exception on both sides.
	for _, both := range []bool{false, true} {
		in := permissive()
		in.Issuers[0].Permissions.AllowUnknownRisk, in.Permissions.AllowUnknownRisk = true, both
		h := newHarnessWith(t, mustPolicy(t, in))
		_, err := h.embedded(t, taskWith(map[string]any{"safety.risk_level": "unknown", "safety.risk_factors": unknown}))
		want := task.IssuerNotAuthorized
		if both {
			want = 0
		}
		expect(t, fmt.Sprint("unknown risk exception on both sides: ", both), err, want)
	}
}

// Step 7 resolves the active policy generation and the source issuer before
// any authority comparison.
// requirement-ref: REQ-TRUST-0010
// requirement-ref: REQ-TRUST-0047
func TestGatePolicyAndIssuerResolution(t *testing.T) {
	ctx := context.Background()
	b := []byte(baseTask)
	gate := func(h *harness, issuer string, generation uint64) error {
		ref, err := h.store.InstallEmbedded(ctx, issuer, generation, b)
		if err != nil {
			t.Fatal("embedded install rejected")
		}
		_, err = h.admitter.Gate(ctx, ref, b)
		return err
	}
	h := newHarness(t)
	expect(t, "matching generation", gate(h, "embedded-issuer", 2), 0)
	expect(t, "lower generation", gate(h, "embedded-issuer", 1), task.PolicyRollback)
	expect(t, "higher generation", gate(h, "embedded-issuer", 3), task.PolicyUnavailable)
	expect(t, "unknown issuer", gate(h, "other-issuer", 2), task.IssuerNotAuthorized)
	expect(t, "issuer name prefix", gate(h, "embedded", 2), task.IssuerNotAuthorized)
	expect(t, "rollback before unknown issuer", gate(h, "other-issuer", 1), task.PolicyRollback)
	if h.normalize.Load() != 1 {
		t.Error("codec reached after a step-7 rejection")
	}
	none := newHarnessWith(t, task.LocalPolicy{})
	expect(t, "no policy installed", gate(none, "embedded-issuer", 2), task.PolicyUnavailable)
	expect(t, "no policy before unknown issuer", gate(none, "other-issuer", 1), task.PolicyUnavailable)

	// Remote tasks resolve the issuer of the verified package the same way. A
	// policy without the issuer cannot hold its keys, so its packages fail
	// verification at step 2 before issuer authorization.
	_, err := h.remote(t, 0, b)
	expect(t, "remote issuer", err, 0)
	in := permissive()
	in.Issuers, in.Keys = in.Issuers[:1], in.Keys[1:]
	_, err = newHarnessWith(t, mustPolicy(t, in)).remote(t, 0, b)
	expect(t, "remote issuer not in policy", err, task.SourceUnauthenticated)
	ref, err := h.store.InstallRemote(ctx, 1, signed(0, b))
	if err != nil {
		t.Fatal("remote install rejected")
	}
	_, err = h.admitter.Gate(ctx, ref, signed(0, b))
	expect(t, "remote rollback", err, task.PolicyRollback)
	// Remote validity is bounded from the earlier of issuance and acceptance.
	in = permissive()
	in.Issuers[1].Permissions.Limits.MaxValiditySeconds = 3600
	h = newHarnessWith(t, mustPolicy(t, in))
	_, err = h.remote(t, -time.Minute, taskWith(map[string]any{"schedule.not_after": notAfter(3540)}))
	expect(t, "remote validity from issuance", err, 0)
	_, err = h.remote(t, -time.Minute, taskWith(map[string]any{"schedule.not_after": notAfter(3541)}))
	expect(t, "remote validity beyond issuance bound", err, task.LimitExceeded)
}

// Steps keep their public order around the new step-7 and step-8 checks.
// requirement-ref: REQ-TRUST-0047
func TestGateAuthorityPrecedence(t *testing.T) {
	ctx := context.Background()
	unknownIssuer := func(b []byte) error {
		h := newHarness(t)
		ref, err := h.store.InstallEmbedded(ctx, "other-issuer", 2, b)
		if err != nil {
			t.Fatal("embedded install rejected")
		}
		_, err = h.admitter.Gate(ctx, ref, b)
		return err
	}
	expect(t, "unknown field before unknown issuer", unknownIssuer(taskWith(map[string]any{"private": 1})), task.SchemaInvalid)
	expect(t, "expiry before unknown issuer", unknownIssuer(taskWith(map[string]any{"schedule.not_after": "2026-10-19T11:00:00Z"})), task.Expired)
	expect(t, "registered limit before unknown issuer", unknownIssuer(taskWith(map[string]any{"execution.limits.max_duration_ms": 60001})), task.LimitExceeded)
	expect(t, "unknown issuer before invalid input", unknownIssuer(taskWith(map[string]any{"input": map[string]any{}})), task.IssuerNotAuthorized)

	in := permissive()
	in.Permissions.Paths = task.Paths{Proxy: true}
	in.Permissions.Limits.Attempt.DurationMS = 40000
	in.Permissions.MaxRisk, in.Permissions.AllowMediumRisk = task.LowRisk, false
	for _, c := range []struct {
		name string
		task map[string]any
		want task.AdmissionCode
	}{
		{"ceiling before path", map[string]any{"execution.limits.max_duration_ms": 40001}, task.LimitExceeded},
		{"invalid input before path", map[string]any{"input": map[string]any{}}, task.SchemaInvalid},
		{"path before risk", nil, task.IssuerNotAuthorized},
	} {
		h := newHarnessWith(t, mustPolicy(t, in))
		_, err := h.embedded(t, taskWith(c.task))
		expect(t, c.name, err, c.want)
	}
}

// Admission-time rows of CC-LIMIT-LIMIT-HANDLING: each registered maximum is
// accepted at the limit and rejected one unit above it with zero network
// activity, and no task-controlled field raises an effective limit. The
// NET-05 nine-step row needs the NET-05 codec (DEV-454) and is not covered.
// requirement-ref: REQ-LIMIT-0026
// requirement-ref: REQ-LIMIT-0030
func TestGateAdmissionLimitHandling(t *testing.T) {
	base := []byte(baseTask)
	for _, c := range []struct {
		name          string
		atLimit, over []byte
	}{
		{"complete task JSON", padded(base, 65536), padded(base, 65537)},
		{"task_id", taskWith(map[string]any{"task_id": strings.Repeat("a", 128)}), taskWith(map[string]any{"task_id": strings.Repeat("a", 129)})},
		{"input", taskWith(map[string]any{"input.subject": strings.Repeat("s", 49152-len(`{"subject":""}`))}), taskWith(map[string]any{"input.subject": strings.Repeat("s", 49153-len(`{"subject":""}`))})},
		{"options", taskWith(map[string]any{"options.note": strings.Repeat("n", 8192-len(`{"note":""}`))}), taskWith(map[string]any{"options.note": strings.Repeat("n", 8193-len(`{"note":""}`))})},
		{"labels", taskWith(map[string]any{"labels": labels(16, 1)}), taskWith(map[string]any{"labels": labels(17, 1)})},
		{"runs", taskWith(map[string]any{"schedule.limits.max_runs": 28}), taskWith(map[string]any{"schedule.limits.max_runs": 29})},
		{"max_duration_ms", taskWith(map[string]any{"execution.limits.max_duration_ms": 60000}), taskWith(map[string]any{"execution.limits.max_duration_ms": 60001})},
		{"max_bytes_sent", taskWith(map[string]any{"execution.limits.max_bytes_sent": 65536}), taskWith(map[string]any{"execution.limits.max_bytes_sent": 65537})},
		{"max_bytes_received", taskWith(map[string]any{"execution.limits.max_bytes_received": 262144}), taskWith(map[string]any{"execution.limits.max_bytes_received": 262145})},
		{"max_connections", taskWith(map[string]any{"execution.limits.max_connections": 12}), taskWith(map[string]any{"execution.limits.max_connections": 13})},
	} {
		h := newHarness(t)
		_, err := h.embedded(t, c.atLimit)
		expect(t, c.name+" at limit", err, 0)
		_, err = h.embedded(t, c.over)
		expect(t, c.name+" one above", err, task.LimitExceeded)
	}
	// Attempts to raise a limit through task-controlled values.
	want := registeredMaxima()
	want.Attempt = task.AttemptLimits{DurationMS: 30000, BytesSent: 32768, BytesReceived: 131072, Connections: 8}
	want.MaxRuns, want.MaxAttempts = 1, 1
	for _, c := range []struct {
		name string
		task map[string]any
		want task.AdmissionCode
	}{
		{"label", map[string]any{"labels": map[string]any{"t.max_duration_ms": "120000", "t.limits": "unlimited"}}, 0},
		{"input member", map[string]any{"input.max_bytes_received": 1048576}, task.SchemaInvalid},
		{"option", map[string]any{"options.max_connections": 64}, task.SchemaInvalid},
		{"schedule member", map[string]any{"schedule.limits.max_attempts": 4}, task.SchemaInvalid},
		{"execution member", map[string]any{"execution.limits.max_result_bytes": 1048576}, task.SchemaInvalid},
		{"safety member", map[string]any{"safety.max_risk": "high"}, task.SchemaInvalid},
		{"unknown field", map[string]any{"limits": map[string]any{"max_duration_ms": 120000}}, task.SchemaInvalid},
	} {
		g, err := newHarness(t).embedded(t, taskWith(c.task))
		expect(t, "raise through "+c.name, err, c.want)
		if err == nil && !reflect.DeepEqual(admission.GatedLimits(g), want) {
			t.Errorf("raise through %s: effective limits changed", c.name)
		}
	}
	// Host configuration cannot grant a value above a registered maximum.
	for name, edit := range map[string]func(*task.PolicyLimits){
		"duration":    func(l *task.PolicyLimits) { l.Attempt.DurationMS = 60001 },
		"bytes sent":  func(l *task.PolicyLimits) { l.Attempt.BytesSent = 65537 },
		"received":    func(l *task.PolicyLimits) { l.Attempt.BytesReceived = 262145 },
		"connections": func(l *task.PolicyLimits) { l.Attempt.Connections = 13 },
		"runs":        func(l *task.PolicyLimits) { l.MaxRuns = 29 },
		"attempts":    func(l *task.PolicyLimits) { l.MaxAttempts = 3 },
		"validity":    func(l *task.PolicyLimits) { l.MaxValiditySeconds = 604801 },
	} {
		for _, side := range []string{"issuer", "local"} {
			if _, err := task.NewLocalPolicy(narrowed(side, func(p *task.Permissions) { edit(&p.Limits) })); err == nil {
				t.Errorf("%s %s above the registered maximum accepted by policy", side, name)
			}
		}
	}
}
