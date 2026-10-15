package result

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
)

func testRef(id string) model.Test {
	return model.Test{ID: id, Version: Contract, ResultVersion: Contract}
}
func binding(id string) Binding {
	cd, _ := observation.NewContextDecoder(Contract, observation.ContextApprovals{})
	path, _ := observation.NewPath(1, 0, "direct", "")
	b := Binding{TaskID: "campaign-1", Output: Output{Test: testRef(id), DNSMode: "system", QueryType: "A", HTTPS: true}, MaxAttempts: 2, MaxDurationMS: 1000, Implementation: model.Implementation{LibraryVersion: "0.1.0", Platform: "android"}, ContextDecoder: cd, Path: path, AccessGeneration: 1}
	b.Output.Limits = observation.ProtocolLimits{MaxBodyBytes: 4096, StatusCheck: true}
	if id == "NET-05" {
		b.Profile = &ProfilePlan{ID: "baseline", Version: 1, Steps: []StepPlan{{"first", Output{Test: testRef("TCP-01")}}, {"second", Output{Test: testRef("TCP-01")}}}}
	}
	return b
}
func decoder(t testing.TB, b Binding) *Decoder {
	t.Helper()
	d, e := New(b)
	if e != nil {
		t.Fatal(e)
	}
	return d
}
func access() string {
	return `{"network_type":"wifi","vpn_state":"unknown","ipv4_available":"yes","ipv6_available":"unknown","validated":"unknown","network_changed":false}`
}
func context() string {
	return `{"access":` + access() + `,"egress":{"status":"unavailable","lookup_source":"none","network_path":"direct"}}`
}
func tcp(status, failure string) string {
	f := ""
	if failure != "" {
		f = `,"failure":"` + failure + `"`
	}
	return `{"type":"tcp_connect","status":"` + status + `","duration_ms":250` + f + `,"details":{"address_family":"ipv4"}}`
}
func tls(status, failure, cert string) string {
	f := ""
	if failure != "" {
		f = `,"failure":"` + failure + `"`
	}
	return `{"type":"tls_handshake","status":"` + status + `","duration_ms":250` + f + `,"details":{"address_family":"ipv4","certificate_validation":"` + cert + `"}}`
}
func dns(engine, status, failure, rcode string) string {
	f := ""
	if failure != "" {
		f = `,"failure":"` + failure + `"`
	}
	responses := `[]`
	if rcode != "" {
		responses = `[{"arrival_index":1,"arrival_offset_ms":250,"rcode":"` + rcode + `","answers":[]}]`
	}
	return `{"type":"dns_query","status":"` + status + `","duration_ms":250` + f + `,"details":{"engine":"` + engine + `","query_type":"A","responses":` + responses + `}}`
}
func http() string {
	return `{"type":"http_exchange","status":"succeeded","duration_ms":250,"details":{"address_family":"ipv4","protocol":"h2","status_code":200,"checks":[{"kind":"status","index":0,"matched":true}]}}`
}
func payload(outcome string, obs ...string) string {
	return `{"outcome":"` + outcome + `","observations":[` + strings.Join(obs, ",") + `]}`
}
func envelopeJSON(id, state, reason, m string) []byte {
	r := ""
	if reason != "" {
		r = `,"reason":"` + reason + `"`
	}
	ctx := ""
	if id != "NET-01" {
		ctx = `,"network_context":` + context()
	}
	p := ""
	if m != "" {
		p = `,"measurement":` + m
	}
	return []byte(`{"schema_version":"0.1.0","task_id":"campaign-1","run_id":"AQEBAQEBAQEBAQEBAQEBAQ","attempt":1,"test":{"id":"` + id + `","version":"0.1.0","result_version":"0.1.0"},"implementation":{"library_version":"0.1.0","platform":"android"},"time":{"started_at":"2026-10-15T12:34:00Z","precision":"minute","duration_ms":1000},"execution":{"state":"` + state + `"` + r + `}` + ctx + p + `}`)
}
func mutate(t testing.TB, input []byte, fn func(map[string]any)) []byte {
	t.Helper()
	var v map[string]any
	if e := json.Unmarshal(input, &v); e != nil {
		t.Fatal(e)
	}
	fn(v)
	b, e := json.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func object(v map[string]any, k string) map[string]any { return v[k].(map[string]any) }
func reject(t testing.TB, d *Decoder, b []byte) error {
	t.Helper()
	v, e := d.Decode(b)
	if e == nil {
		t.Fatalf("accepted invalid result: %s", b)
	}
	if _, e := v.MarshalJSON(); e == nil {
		t.Fatal("partial value escaped")
	}
	return e
}

// requirement-ref: REQ-RESULT-0001
// requirement-ref: REQ-RESULT-0011
// requirement-ref: REQ-RESULT-0032
func TestIndependentVersionsAndNegativeCompletion(t *testing.T) {
	d := decoder(t, binding("TCP-01"))
	input := envelopeJSON("TCP-01", "completed", "", payload("connection_refused", tcp("failed", "connection_refused")))
	v, e := d.Decode(input)
	if e != nil {
		t.Fatal(e)
	}
	if v.View().Execution.State != "completed" || v.View().Execution.Reason != "" {
		t.Fatal("negative observation became execution failure")
	}
	for _, key := range []string{"schema_version", "version", "result_version", "id"} {
		t.Run(key, func(t *testing.T) {
			b := mutate(t, input, func(v map[string]any) {
				if key == "schema_version" {
					v[key] = "9.0.0"
				} else {
					object(v, "test")[key] = "9.0.0"
				}
			})
			if !errors.Is(reject(t, d, b), jsondecode.ErrVersion) {
				t.Fatal("not a version rejection")
			}
		})
	}
	for _, failure := range []string{"timeout", "connection_reset", "network_unreachable", "address_unreachable", "permission_denied", "socket_error", "inconclusive"} {
		if _, e := d.Decode(envelopeJSON("TCP-01", "completed", "", payload(failure, tcp("failed", failure)))); e != nil {
			t.Fatalf("%s: %v", failure, e)
		}
	}
}

// requirement-ref: REQ-RESULT-0002
// requirement-ref: REQ-RESULT-0016
// requirement-ref: REQ-RESULT-0019
// requirement-ref: REQ-RESULT-0034
func TestEnvelopePresenceAndBounds(t *testing.T) {
	d := decoder(t, binding("TCP-01"))
	input := envelopeJSON("TCP-01", "completed", "", payload("connected", tcp("succeeded", "")))
	cases := map[string]func(map[string]any){
		"task binding":           func(v map[string]any) { v["task_id"] = "another-campaign" },
		"run zero":               func(v map[string]any) { v["run_id"] = "AAAAAAAAAAAAAAAAAAAAAA" },
		"run padding":            func(v map[string]any) { v["run_id"] = "AQEBAQEBAQEBAQEBAQEBAQ==" },
		"run noncanonical":       func(v map[string]any) { v["run_id"] = "AQEBAQEBAQEBAQEBAQEBAR" },
		"attempt zero":           func(v map[string]any) { v["attempt"] = 0 },
		"attempt excess":         func(v map[string]any) { v["attempt"] = 3 },
		"fraction":               func(v map[string]any) { v["attempt"] = 1.5 },
		"missing measurement":    func(v map[string]any) { delete(v, "measurement") },
		"null context":           func(v map[string]any) { v["network_context"] = nil },
		"missing context":        func(v map[string]any) { delete(v, "network_context") },
		"reason completed":       func(v map[string]any) { object(v, "execution")["reason"] = "cancelled_by_host" },
		"empty reason completed": func(v map[string]any) { object(v, "execution")["reason"] = "" },
		"invalid state":          func(v map[string]any) { object(v, "execution")["state"] = "failed" },
		"subminute":              func(v map[string]any) { object(v, "time")["started_at"] = "2026-10-15T12:34:01Z" },
		"fractional time":        func(v map[string]any) { object(v, "time")["started_at"] = "2026-10-15T12:34:00.000Z" },
		"offset":                 func(v map[string]any) { object(v, "time")["started_at"] = "2026-10-15T12:34:00+00:00" },
		"precision":              func(v map[string]any) { object(v, "time")["precision"] = "second" },
		"duration quantum":       func(v map[string]any) { object(v, "time")["duration_ms"] = 999 },
		"duration limit":         func(v map[string]any) { object(v, "time")["duration_ms"] = 1250 },
		"cap false":              func(v map[string]any) { object(v, "time")["duration_capped"] = false },
		"cap not ceiling":        func(v map[string]any) { object(v, "time")["duration_capped"] = true },
		"library mismatch":       func(v map[string]any) { object(v, "implementation")["library_version"] = "99.0.0" },
		"platform":               func(v map[string]any) { object(v, "implementation")["platform"] = "linux" },
		"context changed":        func(v map[string]any) { object(object(v, "network_context"), "access")["network_changed"] = true },
	}
	for name, fn := range cases {
		t.Run(name, func(t *testing.T) { reject(t, d, mutate(t, input, fn)) })
	}
	var root map[string]any
	_ = json.Unmarshal(input, &root)
	for key := range root {
		if key == "network_context" {
			continue
		}
		t.Run("required "+key, func(t *testing.T) { reject(t, d, mutate(t, input, func(v map[string]any) { delete(v, key) })) })
	}
	// Exact numeric representation rejects integral decimal/exponent spellings.
	for _, token := range []string{"1.0", "1e0", "-0"} {
		reject(t, d, bytes.Replace(input, []byte(`"attempt":1`), []byte(`"attempt":`+token), 1))
	}
	v, e := d.Decode(mutate(t, input, func(v map[string]any) {
		object(v, "time")["duration_ms"] = 60000
		object(v, "time")["duration_capped"] = true
	}))
	if e != nil || !v.View().Time.DurationCapped {
		t.Fatal(e)
	}
}

// requirement-ref: REQ-RESULT-0021
// requirement-ref: REQ-RESULT-0023
// requirement-ref: REQ-RESULT-0024
// requirement-ref: REQ-RESULT-0042
func TestExecutionContextAndAttemptBinding(t *testing.T) {
	b := binding("TCP-01")
	b.MaxAttempts = 1
	d := decoder(t, b)
	good := envelopeJSON("TCP-01", "completed", "", payload("connected", tcp("succeeded", "")))
	reject(t, d, mutate(t, good, func(v map[string]any) { v["attempt"] = 2 }))
	for _, state := range []string{"incomplete", "not_run"} {
		for _, reason := range []task.ExecutionReason{task.CancelledByHost, task.AttemptDeadlineExceeded, task.ResourceLimit, task.NetworkChanged, task.PlatformSuspended, task.NetworkContextUnavailable, task.ConstraintsUnsatisfied, task.LocalPolicyRejected, task.ProxyNotConfigured, task.ProxyNotPermitted, task.ProxyPathUnsupported, task.ProxyUnavailable, task.ProxyConnectFailed, task.ProxyReconfigured, task.InternalError} {
			input := envelopeJSON("TCP-01", state, string(reason), "")
			if _, e := d.Decode(input); e != nil {
				t.Fatal(state, reason, e)
			}
			reject(t, d, mutate(t, input, func(v map[string]any) { delete(object(v, "execution"), "reason") }))
			reject(t, d, mutate(t, input, func(v map[string]any) { object(v, "execution")["reason"] = "timeout" }))
			reject(t, d, mutate(t, input, func(v map[string]any) {
				v["measurement"] = map[string]any{"outcome": "connected", "observations": []any{}}
			}))
		}
	}
	notRun := mutate(t, envelopeJSON("TCP-01", "not_run", "local_policy_rejected", ""), func(v map[string]any) { delete(v, "network_context") })
	if _, e := d.Decode(notRun); e != nil {
		t.Fatal(e)
	}
	b.CapturedContext = true
	reject(t, decoder(t, b), notRun)
	reject(t, d, mutate(t, envelopeJSON("TCP-01", "incomplete", "network_changed", ""), func(v map[string]any) { delete(v, "network_context") }))
	// Path metadata is supplied locally and cannot be changed by result JSON.
	reject(t, d, bytes.Replace(good, []byte(`"network_path":"direct"`), []byte(`"network_path":"proxy","connection_code":"vpn"`), 1))
}

// requirement-ref: REQ-RESULT-0005
// requirement-ref: REQ-RESULT-0014
// requirement-ref: REQ-RESULT-0025
// requirement-ref: REQ-RESULT-0044
// requirement-ref: REQ-RESULT-0046
func TestUnknownPrivateAndMalformedFields(t *testing.T) {
	d := decoder(t, binding("TCP-01"))
	input := envelopeJSON("TCP-01", "completed", "", payload("connected", tcp("succeeded", "")))
	for _, path := range []string{"root", "test", "implementation", "time", "execution", "measurement", "access", "egress", "observation", "details"} {
		t.Run(path, func(t *testing.T) {
			reject(t, d, mutate(t, input, func(v map[string]any) {
				target := v
				switch path {
				case "access", "egress":
					target = object(object(v, "network_context"), path)
				case "observation", "details":
					target = object(v, "measurement")["observations"].([]any)[0].(map[string]any)
					if path == "details" {
						target = object(target, "details")
					}
				case "root":
				default:
					target = object(v, path)
				}
				target["private_endpoint"] = "DO_NOT_RETAIN"
			}))
		})
	}
	for _, b := range [][]byte{append(append([]byte{}, input...), []byte(` {}`)...), bytes.Replace(input, []byte(`"attempt":1`), []byte(`"attempt":1,"attempt":1`), 1), bytes.Replace(input, []byte(`campaign-1`), []byte(`\ud800`), 1), append([]byte{0xef, 0xbb, 0xbf}, input...)} {
		reject(t, d, b)
	}
	for _, key := range []string{"input", "options", "schedule", "safety", "labels", "input_digest", "user_id", "ip_address", "raw_error", "annotations"} {
		err := reject(t, d, mutate(t, input, func(v map[string]any) { v[key] = "DO_NOT_RETAIN" }))
		if strings.Contains(err.Error(), "DO_NOT_RETAIN") {
			t.Fatal("unsanitized error")
		}
	}
}

// requirement-ref: REQ-RESULT-0004
func TestEncodedAndObservationBoundaries(t *testing.T) {
	d := decoder(t, binding("TCP-01"))
	input := envelopeJSON("TCP-01", "completed", "", payload("connected", tcp("succeeded", "")))
	exact := append(bytes.Repeat([]byte(" "), 65536-len(input)), input...)
	if _, e := d.Decode(exact); e != nil {
		t.Fatal(e)
	}
	if !errors.Is(reject(t, d, append([]byte(" "), exact...)), jsondecode.ErrBounds) {
		t.Fatal("size bound not enforced")
	}
	seventeen := make([]string, 17)
	for i := range seventeen {
		seventeen[i] = tcp("succeeded", "")
	}
	reject(t, d, envelopeJSON("TCP-01", "completed", "", payload("connected", seventeen...)))
}

// requirement-ref: REQ-RESULT-0016
// requirement-ref: REQ-RESULT-0018
// requirement-ref: REQ-RESULT-0036
func TestProducerTimingAndRandomIdentity(t *testing.T) {
	start := time.Now()
	cases := []struct {
		elapsed time.Duration
		want    int
		capped  bool
	}{{0, 0, false}, {1, 250, false}, {250 * time.Millisecond, 250, false}, {250*time.Millisecond + 1, 500, false}, {60 * time.Second, 60000, false}, {60*time.Second + 1, 60000, true}}
	for _, tc := range cases {
		v, e := Timing(start, start.Add(tc.elapsed))
		if e != nil || v.DurationMS != tc.want || v.DurationCapped != tc.capped {
			t.Fatal(tc, v, e)
		}
	}
	if _, e := Timing(start, start.Add(-1)); e == nil {
		t.Fatal("negative elapsed")
	}
	if _, e := Timing(start.Round(0), start.Add(time.Second)); e == nil {
		t.Fatal("wall-only timing accepted")
	}
	seen := map[string]bool{}
	for range 100 {
		id := NewRunID()
		if len(id) != 22 || id == "AAAAAAAAAAAAAAAAAAAAAA" || seen[id] {
			t.Fatal("run identity")
		}
		seen[id] = true
	}
}

func TestImmutableViewsAndProducerBuild(t *testing.T) {
	d := decoder(t, binding("TCP-01"))
	v, e := d.Decode(envelopeJSON("TCP-01", "completed", "", payload("connected", tcp("succeeded", ""))))
	if e != nil {
		t.Fatal(e)
	}
	original, _ := v.MarshalJSON()
	view := v.View()
	view.TaskID = "changed"
	view.Measurement.Outcome = "changed"
	view.Measurement.Observations[0] = observation.Observation{}
	view.NetworkContext = nil
	again, _ := v.MarshalJSON()
	if !bytes.Equal(original, again) {
		t.Fatal("mutated view changed result")
	}
	again[0] = '!'
	again, _ = v.MarshalJSON()
	if !bytes.Equal(original, again) {
		t.Fatal("encoded buffer alias")
	}
	built, e := d.Build(v.View())
	if e != nil {
		t.Fatal(e)
	}
	b, _ := built.MarshalJSON()
	if !bytes.Equal(original, b) {
		t.Fatal("producer/decoder mismatch")
	}
	if _, e = json.Marshal(v.View()); e == nil {
		t.Fatal("detached result serialized")
	}
	if _, e = json.Marshal(*v.View().Measurement); e == nil {
		t.Fatal("detached measurement serialized")
	}
	var zero Value
	if json.Unmarshal(original, &zero) == nil {
		t.Fatal("unversioned import")
	}
	if _, e := json.Marshal(zero); e == nil {
		t.Fatal("zero value serialized")
	}
	invalid := v.View()
	invalid.Measurement.Observations = make([]observation.Observation, 17)
	if _, e := d.Build(invalid); e == nil {
		t.Fatal("unbounded producer")
	}
	// A malicious recursive detached record is rejected before encoding.
	m := &model.Measurement{Outcome: "completed"}
	m.Steps = []model.Step{{Measurement: m}}
	invalid.Measurement = m
	if _, e := d.Build(invalid); e == nil {
		t.Fatal("recursive producer")
	}
}
