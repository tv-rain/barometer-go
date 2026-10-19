package admission_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/admission"
	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/registry"
	"github.com/tv-rain/barometer-go/internal/taskdecode"
	"github.com/tv-rain/barometer-go/internal/testkit"
)

// SYN-01 is a synthetic test-scoped registry entry, never a production entry
// or a runner. Its codec keeps input values literal.
type synParameters struct{ Subject string }

func (synParameters) TaskParameters() {}

type synCodec struct{ calls *atomic.Int32 }

func (synCodec) Definition() taskdecode.Definition {
	text := &jsondecode.Schema{Kind: jsondecode.String}
	return taskdecode.Definition{TestID: "SYN-01", TestVersion: "0.1.0",
		Input:   &jsondecode.Schema{Kind: jsondecode.Object, Fields: map[string]jsondecode.Field{"subject": {Schema: text, Required: true}}},
		Options: &jsondecode.Schema{Kind: jsondecode.Object, Fields: map[string]jsondecode.Field{"note": {Schema: text}}}}
}

func (c synCodec) Normalize(input, _ *jsondecode.Value) (task.Parameters, error) {
	if c.calls != nil {
		c.calls.Add(1)
	}
	subject, _ := input.Lookup("subject")
	if subject.Text() == "reject-me" {
		return nil, errors.New("CANARY-CODEC-ERROR")
	}
	return synParameters{Subject: subject.Text()}, nil
}

// stepClock fails from the given call onward, so a test can make trusted time
// unavailable after delivery resolution, at step 3. The raw error is a canary.
type stepClock struct {
	*fakeClock
	calls, failFrom atomic.Int32
}

func (c *stepClock) Now(ctx context.Context) (time.Time, error) {
	if n := c.calls.Add(1); c.failFrom.Load() > 0 && n >= c.failFrom.Load() {
		return time.Time{}, errors.New("CANARY-CLOCK-ERROR")
	}
	return c.fakeClock.Now(ctx)
}

type harness struct {
	clock     *stepClock
	store     *admission.DeliveryStore
	admitter  *admission.Admitter
	transport *testkit.Transport
	normalize *atomic.Int32
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{clock: &stepClock{fakeClock: newClock()}, normalize: &atomic.Int32{}}
	var err error
	if h.store, err = admission.NewDeliveryStore(h.clock, admission.DeliveryLimits{Capacity: 256, Lifetime: lifetime}, 2); err != nil {
		t.Fatal("store rejected")
	}
	reg, err := registry.New([]registry.Entry{{Test: task.TestVersion{ID: "SYN-01", Version: "0.1.0"}, Codec: synCodec{calls: h.normalize}, BaselineRisk: task.MediumRisk, Proxy: registry.NoProxy{}}})
	if err != nil {
		t.Fatal("registry rejected")
	}
	if h.admitter, err = admission.NewAdmitter(h.store, h.clock, reg); err != nil {
		t.Fatal("admitter rejected")
	}
	// The recording transport is never given to admission; it must stay unused.
	if h.transport, err = testkit.NewTransport(); err != nil {
		t.Fatal("transport rejected")
	}
	t.Cleanup(func() {
		if calls, open := h.transport.Counts(); calls != 0 || open != 0 {
			t.Error("admission caused network activity")
		}
	})
	return h
}

func (h *harness) embedded(t *testing.T, b []byte) (admission.Gated, error) {
	t.Helper()
	ref, err := h.store.InstallEmbedded(context.Background(), "embedded-issuer", 2, b)
	if err != nil {
		t.Fatal("embedded install rejected")
	}
	return h.admitter.Gate(context.Background(), ref, b)
}

func (h *harness) remote(t *testing.T, ahead time.Duration, b []byte) (admission.Gated, error) {
	t.Helper()
	v := verifier()
	v.IssuedAt = task.TimestampFromTime(epoch.Add(ahead))
	ref, err := h.store.InstallRemote(context.Background(), v, v.Seal(b))
	if err != nil {
		t.Fatal("remote install rejected")
	}
	return h.admitter.Gate(context.Background(), ref, b)
}

const baseTask = `{"schema_version":"0.1.0","task_id":"synthetic-task","test_id":"SYN-01","test_version":"0.1.0","input":{"subject":"example"},"options":{},"schedule":{"not_after":"2026-10-20T12:00:00Z"},"execution":{"limits":{}},"safety":{"risk_level":"low","requires_user_consent":false}}`

// taskWith applies dotted-path edits to the base task; a nil value deletes.
func taskWith(edits map[string]any) []byte {
	var v map[string]any
	if err := json.Unmarshal([]byte(baseTask), &v); err != nil {
		panic("bad fixture")
	}
	for path, value := range edits {
		parts := strings.Split(path, ".")
		obj := v
		for _, p := range parts[:len(parts)-1] {
			x, ok := obj[p].(map[string]any)
			if !ok {
				x = map[string]any{}
				obj[p] = x
			}
			obj = x
		}
		if value == nil {
			delete(obj, parts[len(parts)-1])
		} else {
			obj[parts[len(parts)-1]] = value
		}
	}
	b, err := json.Marshal(v)
	if err != nil {
		panic("fixture encode")
	}
	return b
}

// padded appends insignificant whitespace to reach exactly n bytes.
func padded(b []byte, n int) []byte {
	return append(append([]byte(nil), b...), bytes.Repeat([]byte(" "), n-len(b))...)
}

func labels(n, valueBytes int) map[string]any {
	m := map[string]any{}
	for i := range n {
		m[fmt.Sprintf("k%02d", i)] = strings.Repeat("v", valueBytes)
	}
	return m
}

// labelBytes returns 16 labels whose object encodes to exactly total bytes
// (2017 to 2193), with every value within the 128-byte label value maximum.
func labelBytes(total int) map[string]any {
	m := labels(16, 0)
	rest := total - len(`{}`) - 15 - 16*len(`"k00":""`)
	for i := 0; rest > 0; i++ {
		n := min(rest, 128)
		m[fmt.Sprintf("k%02d", i)] = strings.Repeat("v", n)
		rest -= n
	}
	return m
}

func expect(t *testing.T, name string, err error, want task.AdmissionCode) {
	t.Helper()
	if want == 0 {
		if err != nil {
			t.Errorf("%s: rejected as %s", name, err)
		}
		return
	}
	if code(err) != want {
		t.Errorf("%s: class %v, want %v", name, err, admissionCode(want))
	}
}

func admissionCode(c task.AdmissionCode) string {
	err, _ := task.NewAdmissionError(c)
	return err.Error()
}

// requirement-ref: REQ-TRUST-0001
// requirement-ref: REQ-TRUST-0004
func TestGateAcceptsReviewedOrigins(t *testing.T) {
	h := newHarness(t)
	b := []byte(baseTask)
	g, err := h.embedded(t, b)
	if err != nil {
		t.Fatal("embedded task rejected")
	}
	c := admission.GatedCandidate(g)
	if c.ID != "synthetic-task" || c.Parameters != (synParameters{Subject: "example"}) || *c.Schedule.MaxRuns != 1 || c.Execution.Limits.DurationMS != 30000 {
		t.Error("candidate not decoded through the registry codec with defaults")
	}
	if !admission.GatedAcceptedAt(g).Equal(epoch) || admission.GatedTest(g) != (task.TestVersion{ID: "SYN-01", Version: "0.1.0"}) || admission.GatedBaselineRisk(g) != task.MediumRisk {
		t.Error("trusted time or registry entry not carried")
	}
	if s := admission.GatedSource(g); s.Origin() != task.Embedded || s.IssuerID() != "embedded-issuer" {
		t.Error("source context not carried")
	}
	for _, ahead := range []time.Duration{-time.Hour, 0, admission.MaxIssuedAhead} {
		g, err := h.remote(t, ahead, b)
		if err != nil || admission.GatedSource(g).Origin() != task.Remote {
			t.Errorf("remote task issued %v ahead rejected", ahead)
		}
	}
	if fmt.Sprintf("%v %+v %#v %s", g, g, g, g) != "admission.Gated admission.Gated admission.Gated admission.Gated" {
		t.Error("gate outcome formatting exposes values")
	}
}

// Each case fails exactly one implemented step and expects its REQ-TRUST-0037
// class. Later steps are not reached: the codec is never called on rejection.
// requirement-ref: REQ-TRUST-0001
// requirement-ref: REQ-TRUST-0017
// requirement-ref: REQ-TRUST-0018
// requirement-ref: REQ-TRUST-0036
// requirement-ref: REQ-TRUST-0037
// requirement-ref: REQ-SAFETY-0017
// requirement-ref: REQ-LIMIT-0004
// requirement-ref: REQ-LIMIT-0005
// requirement-ref: REQ-LIMIT-0010
func TestGateStepClasses(t *testing.T) {
	ctx := context.Background()
	h := newHarness(t)
	base := []byte(baseTask)
	ref := mustEmbedded(t, h.store, base)
	other := newHarness(t)
	foreign := mustEmbedded(t, other.store, base)
	mismatch := append([]byte(nil), base...)
	mismatch[len(mismatch)-2] = ' '
	unauthenticated := verifier()
	unauthenticated.IssuerAuthenticated = false

	// Steps 1 and 2: no record, a foreign or mismatched record, or no proof.
	for name, run := range map[string]func() error{
		"missing record":   func() error { _, err := h.admitter.Gate(ctx, strings.Repeat("0", 32), base); return err },
		"malformed ref":    func() error { _, err := h.admitter.Gate(ctx, "not-a-reference", base); return err },
		"foreign instance": func() error { _, err := h.admitter.Gate(ctx, foreign, base); return err },
		"byte mismatch":    func() error { _, err := h.admitter.Gate(ctx, ref, mismatch); return err },
		"encryption without issuer authentication": func() error {
			if _, err := h.store.InstallRemote(ctx, unauthenticated, unauthenticated.Seal(base)); err == nil {
				t.Fatal("unauthenticated remote delivery installed")
			}
			_, err := h.admitter.Gate(ctx, strings.Repeat("a", 32), base)
			return err
		},
	} {
		expect(t, name, run(), task.SourceUnauthenticated)
	}
	if h.normalize.Load() != 0 {
		t.Fatal("codec reached on a provenance rejection")
	}
	// Expired record, then trusted time unavailable at step 1 and at step 3.
	stale := mustEmbedded(t, h.store, base)
	if err := h.clock.Clock.Advance(lifetime); err != nil {
		t.Fatal(err)
	}
	_, err := h.admitter.Gate(ctx, stale, base)
	expect(t, "expired record", err, task.SourceUnauthenticated)
	h.clock.down.Store(true)
	_, err = h.admitter.Gate(ctx, ref, base)
	expect(t, "clock unavailable at resolution", err, task.TrustUnavailable)
	h.clock.down.Store(false)

	h = newHarness(t)
	fresh := mustEmbedded(t, h.store, base)
	h.clock.failFrom.Store(h.clock.calls.Load() + 2)
	_, err = h.admitter.Gate(ctx, fresh, base)
	expect(t, "clock unavailable at step 3", err, task.TrustUnavailable)

	h = newHarness(t)
	_, err = h.remote(t, admission.MaxIssuedAhead+time.Second, base)
	expect(t, "remote issued_at 301 seconds ahead", err, task.SourceUnauthenticated)
	_, err = h.remote(t, admission.MaxIssuedAhead+time.Second/2, base)
	expect(t, "remote issued_at 300.5 seconds ahead", err, task.SourceUnauthenticated)

	cases := []struct {
		name string
		task []byte
		want task.AdmissionCode
	}{
		// Step 3: the complete-task limit on original bytes, before parsing.
		{"65535 bytes", padded(base, 65535), 0},
		{"65536 bytes", padded(base, admission.MaxTaskBytes), 0},
		{"65537 bytes", padded(base, admission.MaxTaskBytes+1), task.LimitExceeded},
		{"unknown field crossing the limit", padded(taskWith(map[string]any{"unknown": strings.Repeat("x", 65000)}), admission.MaxTaskBytes+1), task.LimitExceeded},
		{"unknown field within the limit", taskWith(map[string]any{"unknown": strings.Repeat("x", 65000)}), task.SchemaInvalid},
		{"oversized malformed bytes", bytes.Repeat([]byte("!"), admission.MaxTaskBytes+1), task.LimitExceeded},
		{"malformed JSON", []byte(`{"schema_version":`), task.SchemaInvalid},
		{"invalid UTF-8", append([]byte(baseTask[:len(baseTask)-1]+`,"labels":{"k":"`), 0xff, '"', '}', '}'), task.SchemaInvalid},
		{"duplicate member", []byte(baseTask[:len(baseTask)-1] + `,"task_id":"synthetic-task"}`), task.SchemaInvalid},
		// Step 4: closed top-level schema and exact versions.
		{"unknown top-level field", taskWith(map[string]any{"private": "x"}), task.SchemaInvalid},
		{"missing safety", taskWith(map[string]any{"safety": nil}), task.SchemaInvalid},
		{"input not an object", taskWith(map[string]any{"input": "x"}), task.SchemaInvalid},
		{"unsupported schema_version", taskWith(map[string]any{"schema_version": "0.2.0"}), task.SchemaInvalid},
		{"malformed test_version", taskWith(map[string]any{"test_version": "v1"}), task.SchemaInvalid},
		{"malformed test_id", taskWith(map[string]any{"test_id": "syn-01"}), task.SchemaInvalid},
		{"unregistered test_version", taskWith(map[string]any{"test_version": "0.2.0"}), task.Unsupported},
		{"unregistered test_id", taskWith(map[string]any{"test_id": "SYN-02"}), task.Unsupported},
		// Step 6: field values, registered limits and expiry.
		{"expired", taskWith(map[string]any{"schedule.not_after": "2026-10-19T11:59:59Z"}), task.Expired},
		{"expiring at acceptance", taskWith(map[string]any{"schedule.not_after": "2026-10-19T12:00:00Z"}), task.Expired},
		{"lifetime over seven days", taskWith(map[string]any{"schedule.not_after": "2026-10-26T12:00:01Z"}), task.LimitExceeded},
		{"embedded without not_after", taskWith(map[string]any{"schedule.not_after": nil}), task.SchemaInvalid},
		{"duration at maximum", taskWith(map[string]any{"execution.limits.max_duration_ms": 60000}), 0},
		{"duration above maximum", taskWith(map[string]any{"execution.limits.max_duration_ms": 60001}), task.LimitExceeded},
		{"duration below minimum", taskWith(map[string]any{"execution.limits.max_duration_ms": 0}), task.SchemaInvalid},
		{"huge integer", []byte(strings.Replace(baseTask, `"limits":{}`, `"limits":{"max_connections":`+strings.Repeat("9", 5000)+`}`, 1)), task.LimitExceeded},
		{"runs above maximum", taskWith(map[string]any{"schedule.limits.max_runs": 29}), task.LimitExceeded},
		{"battery out of range", taskWith(map[string]any{"execution.device.minimum_battery_percent": 101}), task.SchemaInvalid},
		{"sixteen labels", taskWith(map[string]any{"labels": labels(16, 1)}), 0},
		{"seventeen labels", taskWith(map[string]any{"labels": labels(17, 1)}), task.LimitExceeded},
		{"labels at 2048 encoded bytes", taskWith(map[string]any{"labels": labelBytes(2048)}), 0},
		{"labels at 2049 encoded bytes", taskWith(map[string]any{"labels": labelBytes(2049)}), task.LimitExceeded},
		{"labels over 2048 bytes", taskWith(map[string]any{"labels": labels(16, 128)}), task.LimitExceeded},
		{"input at 49152 bytes", taskWith(map[string]any{"input.subject": strings.Repeat("s", 49152-len(`{"subject":""}`))}), 0},
		{"input over 49152 bytes", taskWith(map[string]any{"input.subject": strings.Repeat("s", 49153-len(`{"subject":""}`))}), task.LimitExceeded},
		{"escaped input over 49152 encoded bytes", []byte(strings.Replace(baseTask, `"subject":"example"`, `"subject":"`+strings.Repeat("\\u0061", 8190)+`"`, 1)), task.LimitExceeded},
		{"options at 8192 bytes", taskWith(map[string]any{"options.note": strings.Repeat("n", 8192-len(`{"note":""}`))}), 0},
		{"options over 8192 bytes", taskWith(map[string]any{"options.note": strings.Repeat("n", 8193-len(`{"note":""}`))}), task.LimitExceeded},
		{"unknown risk factor type", taskWith(map[string]any{"safety.risk_factors": []any{map[string]any{"type": "unregistered_factor", "level": "low"}}}), task.Unsupported},
		{"risk factor above declared level", taskWith(map[string]any{"safety.risk_factors": []any{map[string]any{"type": "data_sensitivity", "level": "high"}}}), task.SchemaInvalid},
		// Step 8, input and options: the registry-selected codec (marked late).
		{"input missing required member", taskWith(map[string]any{"input": map[string]any{}}), task.SchemaInvalid},
		{"unknown option", taskWith(map[string]any{"options.extra": true}), task.SchemaInvalid},
		{"codec rejection", taskWith(map[string]any{"input.subject": "reject-me"}), task.SchemaInvalid},
	}
	for _, c := range cases {
		h := newHarness(t)
		_, err := h.embedded(t, c.task)
		expect(t, c.name, err, c.want)
		if late := strings.HasPrefix(c.name, "input missing") || c.name == "unknown option" || c.name == "codec rejection"; c.want != 0 && !late && h.normalize.Load() != 0 {
			t.Errorf("%s: codec reached after an earlier step failed", c.name)
		}
	}
}

// requirement-ref: REQ-TRUST-0019
// requirement-ref: REQ-LIMIT-0004
func TestGateTaskIdentifierRules(t *testing.T) {
	h := newHarness(t)
	for _, c := range []struct {
		id   string
		want task.AdmissionCode
	}{
		{"a", 0}, {"0", 0}, {"a.b_c-d", 0}, {strings.Repeat("a", 127), 0}, {strings.Repeat("a", 128), 0},
		{"", task.SchemaInvalid}, {strings.Repeat("a", 129), task.LimitExceeded}, {strings.Repeat("A", 129), task.LimitExceeded},
		{"Task", task.SchemaInvalid}, {"task id", task.SchemaInvalid}, {"task/id", task.SchemaInvalid}, {"tâsk", task.SchemaInvalid},
		{"_task", task.SchemaInvalid}, {".task", task.SchemaInvalid}, {"-task", task.SchemaInvalid}, {"${installation}", task.SchemaInvalid},
	} {
		_, err := h.embedded(t, taskWith(map[string]any{"task_id": c.id}))
		expect(t, "task_id "+fmt.Sprintf("%q", c.id), err, c.want)
	}
	_, err := h.embedded(t, taskWith(map[string]any{"task_id": 7}))
	expect(t, "task_id number", err, task.SchemaInvalid)
}

// The first failing step decides the class (REQ-TRUST-0037).
// requirement-ref: REQ-TRUST-0037
func TestGatePrecedence(t *testing.T) {
	long := strings.Repeat("a", 129)
	expired := "2026-10-19T11:00:00Z"
	for _, c := range []struct {
		name string
		task []byte
		want task.AdmissionCode
	}{
		{"unregistered version before 129-byte task_id", taskWith(map[string]any{"test_version": "0.2.0", "task_id": long}), task.Unsupported},
		{"unknown field before unregistered version", taskWith(map[string]any{"test_version": "0.2.0", "private": 1}), task.SchemaInvalid},
		{"unknown field before over-limit labels", taskWith(map[string]any{"private": 1, "labels": labels(17, 1)}), task.SchemaInvalid},
		{"task_id characters before field limit", taskWith(map[string]any{"task_id": "Task", "execution.limits.max_duration_ms": 60001}), task.SchemaInvalid},
		{"task_id length before expiry", taskWith(map[string]any{"task_id": long, "schedule.not_after": expired}), task.LimitExceeded},
		{"expiry before invalid input", taskWith(map[string]any{"schedule.not_after": expired, "input": map[string]any{"other": "x"}}), task.Expired},
		{"unsupported risk factor before invalid input", taskWith(map[string]any{"safety.risk_factors": []any{map[string]any{"type": "unregistered_factor", "level": "low"}}, "input": map[string]any{}}), task.Unsupported},
		{"over-limit input before invalid input member", taskWith(map[string]any{"input": map[string]any{"other": strings.Repeat("x", 49152)}}), task.LimitExceeded},
	} {
		h := newHarness(t)
		_, err := h.embedded(t, c.task)
		expect(t, c.name, err, c.want)
	}
	// Step 2 before 3: an oversized task without authenticated provenance never
	// gets a record, so it fails at step 1, not on size.
	h := newHarness(t)
	v := verifier()
	v.IssuerAuthenticated = false
	big := padded([]byte(baseTask), admission.MaxTaskBytes+1)
	if _, err := h.store.InstallRemote(context.Background(), v, v.Seal(big)); err == nil {
		t.Fatal("unauthenticated remote delivery installed")
	}
	_, err := h.admitter.Gate(context.Background(), strings.Repeat("b", 32), big)
	expect(t, "unauthenticated before byte limit", err, task.SourceUnauthenticated)
	// Freshness and trusted time before size.
	_, err = h.remote(t, admission.MaxIssuedAhead+time.Second, big)
	expect(t, "issued_at before byte limit", err, task.SourceUnauthenticated)
	h = newHarness(t)
	ref := mustEmbedded(t, h.store, big)
	h.clock.failFrom.Store(h.clock.calls.Load() + 2)
	_, err = h.admitter.Gate(context.Background(), ref, big)
	expect(t, "trusted time before byte limit", err, task.TrustUnavailable)
}

// Task-controlled authority claims and recipient selectors are unknown fields;
// placeholder text is kept literal and never interpolated locally.
// requirement-ref: REQ-TRUST-0002
// requirement-ref: REQ-TRUST-0026
// requirement-ref: REQ-TRUST-0027
func TestGateNoTaskControlledAuthorityOrInterpolation(t *testing.T) {
	h := newHarness(t)
	for _, path := range []string{"authenticated", "issuer_id", "policy_generation", "trusted", "source", "recipients", "installation_id", "linkage_token",
		"bundle_id", "run_id", "account_id", "advertising_id", "device_id", "push_token", "ip_address", "location", "eligibility", "selector",
		"execution.device.installation_id", "execution.network.ip_address", "schedule.cohort", "safety.consented_cohort"} {
		_, err := h.embedded(t, taskWith(map[string]any{path: "x"}))
		expect(t, "task-controlled "+path, err, task.SchemaInvalid)
	}
	placeholders := taskWith(map[string]any{"input.subject": "{{country}}", "options.note": "${app_version}", "labels": map[string]any{"cohort": "${cohort}", "network": "%NETWORK%"}})
	claim := taskWith(map[string]any{"input.subject": "{{country}}", "options.note": "${app_version}", "labels": map[string]any{"cohort": "${cohort}", "network": "%NETWORK%", "note": "authenticated: true"}})
	var first task.Candidate
	for i, b := range [][]byte{placeholders, placeholders, claim} {
		g, err := newHarness(t).embedded(t, b)
		if err != nil {
			t.Fatal("placeholder text rejected")
		}
		c := admission.GatedCandidate(g)
		if c.Parameters != (synParameters{Subject: "{{country}}"}) || len(c.Labels) < 2 || c.Labels[0] != (task.Label{Key: "cohort", Value: "${cohort}"}) || c.Labels[1] != (task.Label{Key: "network", Value: "%NETWORK%"}) {
			t.Fatal("placeholder text was interpolated or changed")
		}
		if s := admission.GatedSource(g); s.Authenticated() || s.Origin() != task.Embedded {
			t.Fatal("label text changed source context")
		}
		if i == 0 {
			first = c
		} else if i == 1 && !reflect.DeepEqual(first, c) {
			t.Fatal("identical bytes produced different candidates on two instances")
		}
	}
}

// Rejections carry only the fixed class string, are deterministic, change no
// state and leave the delivery record usable.
// requirement-ref: REQ-TRUST-0033
// requirement-ref: REQ-TRUST-0035
func TestGateErrorsCarryNoData(t *testing.T) {
	canaries := []string{"CANARY-PAYLOAD", "CANARY-TARGET", "CANARY-LABEL", "CANARY-CODEC-ERROR", "CANARY-CLOCK-ERROR", "embedded-issuer", "remote-issuer"}
	failing := [][]byte{
		taskWith(map[string]any{"private": "CANARY-PAYLOAD", "input.subject": "https://CANARY-TARGET.example/"}),
		taskWith(map[string]any{"labels": map[string]any{"Bad": "CANARY-LABEL"}}),
		taskWith(map[string]any{"input.subject": "reject-me"}),
		taskWith(map[string]any{"test_version": "0.2.0", "input.subject": "CANARY-PAYLOAD"}),
		taskWith(map[string]any{"schedule.not_after": "2026-10-18T00:00:00Z", "labels": map[string]any{"k": "CANARY-LABEL"}}),
		padded(taskWith(map[string]any{"input.subject": "CANARY-PAYLOAD"}), admission.MaxTaskBytes+1),
	}
	check := func(name string, err error) {
		t.Helper()
		var ae task.AdmissionError
		if !errors.As(err, &ae) || err.Error() != admissionCode(ae.Code()) || errors.Unwrap(err) != nil {
			t.Errorf("%s: not a code-only admission error", name)
			return
		}
		formatted := fmt.Sprintf("%v %+v %#v %s", err, err, err, err)
		for _, canary := range canaries {
			if strings.Contains(formatted, canary) {
				t.Errorf("%s: error exposes a canary", name)
			}
		}
	}
	h := newHarness(t)
	for i, b := range failing {
		ref := mustEmbedded(t, h.store, b)
		_, first := h.admitter.Gate(context.Background(), ref, b)
		_, again := h.admitter.Gate(context.Background(), ref, b)
		check(fmt.Sprint("case ", i), first)
		if first == nil || first.Error() != again.Error() {
			t.Errorf("case %d: repeated delivery changed the result", i)
		}
		if _, err := h.store.Resolve(context.Background(), ref, b); err != nil {
			t.Errorf("case %d: rejection consumed or invalidated the record", i)
		}
	}
	ref := mustEmbedded(t, h.store, []byte(baseTask))
	h.clock.failFrom.Store(h.clock.calls.Load() + 2)
	_, err := h.admitter.Gate(context.Background(), ref, []byte(baseTask))
	check("clock failure", err)
	h.clock.failFrom.Store(0)
	if _, err := h.admitter.Gate(context.Background(), ref, []byte(baseTask)); err != nil {
		t.Error("valid task rejected after earlier rejections")
	}
}

// The gate owns its outcome: caller bytes are not modified or aliased, and
// concurrent gates of one instance are independent.
func TestGateOwnershipAndConcurrency(t *testing.T) {
	h := newHarness(t)
	b := taskWith(map[string]any{"labels": map[string]any{"campaign": "spring"}})
	before := append([]byte(nil), b...)
	ref := mustEmbedded(t, h.store, b)
	g, err := h.admitter.Gate(context.Background(), ref, b)
	if err != nil || !bytes.Equal(b, before) {
		t.Fatal("gate rejected or modified caller bytes")
	}
	for i := range b {
		b[i] = 'x'
	}
	if c := admission.GatedCandidate(g); c.Labels[0].Value != "spring" || c.Parameters != (synParameters{Subject: "example"}) {
		t.Fatal("gate outcome aliases caller bytes")
	}
	var wg sync.WaitGroup
	var failures atomic.Int32
	for i := range 32 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b := []byte(baseTask)
			want := task.SchemaInvalid
			if i%2 == 0 {
				want = 0
			} else {
				b = taskWith(map[string]any{"private": i})
			}
			ref, err := h.store.InstallEmbedded(context.Background(), "embedded-issuer", 2, b)
			if err != nil {
				failures.Add(1)
				return
			}
			if _, err := h.admitter.Gate(context.Background(), ref, b); code(err) != want {
				failures.Add(1)
			}
		}()
	}
	wg.Wait()
	if failures.Load() != 0 {
		t.Fatal("concurrent gates interfered")
	}
}

func TestNewAdmitterRejectsIncompleteConfiguration(t *testing.T) {
	h := newHarness(t)
	reg, _ := registry.New([]registry.Entry{{Test: task.TestVersion{ID: "SYN-01", Version: "0.1.0"}, Codec: synCodec{}, BaselineRisk: task.LowRisk, Proxy: registry.NoProxy{}}})
	for name, build := range map[string]func() (*admission.Admitter, error){
		"no store":    func() (*admission.Admitter, error) { return admission.NewAdmitter(nil, h.clock, reg) },
		"no clock":    func() (*admission.Admitter, error) { return admission.NewAdmitter(h.store, nil, reg) },
		"no registry": func() (*admission.Admitter, error) { return admission.NewAdmitter(h.store, h.clock, nil) },
	} {
		if a, err := build(); a != nil || !errors.Is(err, admission.ErrAdmitterConfig) {
			t.Errorf("%s: incomplete admitter accepted", name)
		}
	}
	var nilAdmitter *admission.Admitter
	if _, err := nilAdmitter.Gate(context.Background(), "", nil); code(err) != task.TrustUnavailable {
		t.Error("nil admitter did not fail closed")
	}
}

// FuzzGate checks that arbitrary bound bytes produce either a gate outcome or
// exactly one code-only admission class, deterministically and without panics.
func FuzzGate(f *testing.F) {
	for _, seed := range [][]byte{[]byte(baseTask), taskWith(map[string]any{"private": 1}), taskWith(map[string]any{"test_version": "0.2.0"}), []byte(`{`), []byte(`[]`), []byte(`{"schema_version":"0.1.0"}`)} {
		f.Add(seed)
	}
	h := &harness{clock: &stepClock{fakeClock: newClock()}}
	store, err := admission.NewDeliveryStore(h.clock, admission.DeliveryLimits{Capacity: 256, Lifetime: lifetime}, 2)
	if err != nil {
		f.Fatal("store rejected")
	}
	reg, err := registry.New([]registry.Entry{{Test: task.TestVersion{ID: "SYN-01", Version: "0.1.0"}, Codec: synCodec{}, BaselineRisk: task.LowRisk, Proxy: registry.NoProxy{}}})
	if err != nil {
		f.Fatal("registry rejected")
	}
	admitter, err := admission.NewAdmitter(store, h.clock, reg)
	if err != nil {
		f.Fatal("admitter rejected")
	}
	f.Fuzz(func(t *testing.T, b []byte) {
		ref, err := store.InstallEmbedded(context.Background(), "embedded-issuer", 2, b)
		if err != nil {
			if len(b) > 0 && len(b) <= task.MaxDeliveryBytes {
				store.Invalidate()
				if ref, err = store.InstallEmbedded(context.Background(), "embedded-issuer", 2, b); err != nil {
					t.Fatal("bounded install rejected")
				}
			} else {
				return
			}
		}
		_, first := admitter.Gate(context.Background(), ref, b)
		_, again := admitter.Gate(context.Background(), ref, b)
		if first != nil && (code(first) == 0 || first.Error() != admissionCode(code(first))) || fmt.Sprint(first) != fmt.Sprint(again) {
			t.Fatal("gate returned a non-class or nondeterministic result")
		}
		if first == nil && len(b) > admission.MaxTaskBytes {
			t.Fatal("over-limit task passed the gates")
		}
	})
}
