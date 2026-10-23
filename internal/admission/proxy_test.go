package admission_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/admission"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/registry"
)

// countingProxy is a synthetic test-scoped rule: a SYN-01 subject of
// "no-proxy" is incompatible with HTTP CONNECT and every other SYN-01 value is
// compatible. It counts its calls so a test can show when it is not consulted.
func countingProxy(calls *atomic.Int32) registry.ProxyRule {
	return proxyRule(func(p task.Parameters) bool {
		calls.Add(1)
		s, ok := p.(synParameters)
		return ok && s.Subject != "no-proxy"
	})
}

// proxyTask is the base task on the proxy path with the given subject and
// extra edits.
func proxyTask(subject string, extra map[string]any) []byte {
	edits := map[string]any{"execution.network_path": "proxy", "input.subject": subject}
	for k, v := range extra {
		edits[k] = v
	}
	return taskWith(edits)
}

// A proxy task whose parameters the registered rule reports as incompatible
// with HTTP CONNECT is task_schema_invalid at step 8 (REQ-TRUST-0039), for
// embedded and remote origin, and is never admitted to end in
// proxy_path_unsupported (REQ-PROXY-0042). Compatibility is evaluated after
// input and options and before path authority and target registration; the
// earlier steps keep their classes. A direct task is not checked. The
// harness's unused recording transport sees zero calls: every rejection is
// atomic and before network activity (REQ-PROXY-0027).
// requirement-ref: REQ-TRUST-0039
// requirement-ref: REQ-PROXY-0026
// requirement-ref: REQ-PROXY-0027
// requirement-ref: REQ-PROXY-0042
func TestGateProxyCompatibility(t *testing.T) {
	directOnly := func(p *task.Permissions) { p.Paths = task.Paths{Direct: true} }
	for _, c := range []struct {
		name    string
		policy  func(*task.PolicyInput)
		task    []byte
		want    task.AdmissionCode
		consult bool // whether the rule must be called
	}{
		{"compatible proxy task", nil, proxyTask("example", nil), 0, true},
		{"incompatible proxy task", nil, proxyTask("no-proxy", nil), task.SchemaInvalid, true},
		{"incompatible parameters on the direct path", nil, taskWith(map[string]any{"input.subject": "no-proxy"}), 0, false},
		{"compatible task, proxy not permitted", func(in *task.PolicyInput) { directOnly(&in.Permissions) }, proxyTask("example", nil), task.IssuerNotAuthorized, true},
		{"incompatibility before issuer path authority", func(in *task.PolicyInput) { directOnly(&in.Issuers[0].Permissions) }, proxyTask("no-proxy", nil), task.SchemaInvalid, true},
		{"incompatibility before local path authority", func(in *task.PolicyInput) { directOnly(&in.Permissions) }, proxyTask("no-proxy", nil), task.SchemaInvalid, true},
		{"incompatibility before target registration", func(in *task.PolicyInput) { in.Targets[0].Paths = task.Paths{Direct: true} }, proxyTask("no-proxy", nil), task.SchemaInvalid, true},
		{"compatible task, registration direct only", func(in *task.PolicyInput) { in.Targets[0].Paths = task.Paths{Direct: true} }, proxyTask("example", nil), task.TargetNotRegistered, true},
		{"incompatibility before risk bound", func(in *task.PolicyInput) {
			in.Permissions.MaxRisk, in.Permissions.AllowMediumRisk = task.LowRisk, false
		}, proxyTask("no-proxy", nil), task.SchemaInvalid, true},
		{"invalid input before compatibility", nil, proxyTask("no-proxy", map[string]any{"input.unexpected": "x"}), task.SchemaInvalid, false},
		{"invalid options before compatibility", nil, proxyTask("no-proxy", map[string]any{"options.unexpected": "x"}), task.SchemaInvalid, false},
		{"codec failure before compatibility", nil, proxyTask("reject-me", nil), task.SchemaInvalid, false},
		{"issuer authority (step 7) before compatibility", func(in *task.PolicyInput) {
			in.Issuers[0].Permissions.Tests = []task.TestVersion{{ID: "SYN-02", Version: "0.1.0"}}
			in.Issuers[0].Permissions.RegistrationIDs = nil
		}, proxyTask("no-proxy", nil), task.IssuerNotAuthorized, false},
		{"policy rollback before compatibility", func(in *task.PolicyInput) { in.Generation = 3 }, proxyTask("no-proxy", nil), task.PolicyRollback, false},
		{"egress requirement does not bypass compatibility", nil, proxyTask("no-proxy", map[string]any{requiredFields: []any{"asn"}}), task.SchemaInvalid, true},
	} {
		h := newHarness(t)
		var calls atomic.Int32
		entry := synEntry(h, task.LowRisk)
		entry.Proxy = countingProxy(&calls)
		rebind(t, h, mustPolicy(t, narrowedPolicy(c.policy)), entry)
		g, err := h.embedded(t, c.task)
		expect(t, c.name, err, c.want)
		if consulted := calls.Load() > 0; consulted != c.consult {
			t.Errorf("%s: rule consulted %v, want %v", c.name, consulted, c.consult)
		}
		if c.want == 0 && err == nil {
			path := task.Paths{Direct: true}
			if c.consult {
				path = task.Paths{Proxy: true}
			}
			if admission.GatedPaths(g) != path {
				t.Errorf("%s: effective path not narrowed to the task", c.name)
			}
		}
	}

	// Remote origin takes the same row.
	h := newHarness(t)
	var calls atomic.Int32
	entry := synEntry(h, task.LowRisk)
	entry.Proxy = countingProxy(&calls)
	rebind(t, h, mustPolicy(t, permissive()), entry)
	_, err := h.remote(t, time.Second, proxyTask("no-proxy", nil))
	expect(t, "remote incompatible proxy task", err, task.SchemaInvalid)
	_, err = h.remote(t, time.Second, proxyTask("example", map[string]any{"task_id": "synthetic-remote"}))
	expect(t, "remote compatible proxy task", err, 0)

	// A test whose rule is never compatible rejects every proxy task and no
	// direct task.
	h = newHarness(t)
	entry = synEntry(h, task.LowRisk)
	entry.Proxy = registry.NoProxy{}
	rebind(t, h, mustPolicy(t, permissive()), entry)
	_, err = h.embedded(t, proxyTask("example", nil))
	expect(t, "never-compatible test on the proxy path", err, task.SchemaInvalid)
	_, err = h.embedded(t, []byte(baseTask))
	expect(t, "never-compatible test on the direct path", err, 0)
}

// NET-05 composite compatibility needs the NET-05 codec (DEV-454). Until then
// an enabled composite task fails closed at step 8 as task_unsupported before
// its registry rule is consulted, so a proxy NET-05 task is not classified as
// incompatible by a placeholder rule.
// requirement-ref: REQ-TRUST-0036
// requirement-ref: REQ-TRUST-0039
func TestGateCompositeBeforeProxyCompatibility(t *testing.T) {
	net05 := task.TestVersion{ID: "NET-05", Version: "0.1.0"}
	h := newHarness(t)
	var calls atomic.Int32
	rebind(t, h, mustPolicy(t, net05Policy(nil)), synEntry(h, task.MediumRisk),
		registry.Entry{Test: net05, Codec: synCodec{calls: h.normalize, id: "NET-05"}, BaselineRisk: task.MediumRisk, Targets: net05Targets,
			Proxy: proxyRule(func(task.Parameters) bool { calls.Add(1); return false }), Composite: true})
	_, err := h.embedded(t, taskWith(map[string]any{"test_id": "NET-05", "execution.network_path": "proxy"}))
	expect(t, "proxy NET-05 task", err, task.Unsupported)
	if calls.Load() != 0 {
		t.Error("composite proxy rule consulted before the composite capability check")
	}
}

// narrowedPolicy returns the permissive policy with edit applied.
func narrowedPolicy(edit func(*task.PolicyInput)) task.PolicyInput {
	in := permissive()
	if edit != nil {
		edit(&in)
	}
	return in
}
