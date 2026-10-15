package taskdecode

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/testkit"
	"github.com/tv-rain/barometer-go/internal/validation"
)

// SYN-01 is a synthetic schema-selection fixture, never a registered runner.
type fixtureParameters struct {
	Subject string
	Enabled bool
}

func (fixtureParameters) TaskParameters() {}

type fixtureCodec struct {
	def   Definition
	calls *atomic.Int32
	fail  bool
}

func (c fixtureCodec) Definition() Definition { return c.def }
func (c fixtureCodec) Normalize(input, options *jsondecode.Value) (task.Parameters, error) {
	if c.calls != nil {
		c.calls.Add(1)
	}
	if c.fail {
		return nil, errors.New("sensitive codec error must not escape")
	}
	subject := "subject"
	if c.def.TestVersion == "0.2.0" {
		subject = "target"
	}
	p := fixtureParameters{Subject: strings.ToLower(text(input, subject, "")), Enabled: true}
	if v := member(options, "enabled"); v != nil {
		p.Enabled = v.Bool()
	}
	return p, nil
}
func fixtureDefinition(id, ver string) Definition {
	key := "subject"
	if ver == "0.2.0" {
		key = "target"
	}
	input := object(map[string]jsondecode.Field{key: field(&jsondecode.Schema{Kind: jsondecode.String}, true)})
	options := object(map[string]jsondecode.Field{"enabled": field(&jsondecode.Schema{Kind: jsondecode.Boolean}, false)})
	if id == "NET-01" {
		input = object(nil)
		options = object(nil)
	}
	return Definition{TestID: id, TestVersion: ver, Input: input, Options: options}
}
func fixtureDecoder(t testing.TB, codecs ...TestCodec) *Decoder {
	t.Helper()
	if len(codecs) == 0 {
		codecs = []TestCodec{fixtureCodec{def: fixtureDefinition("SYN-01", "0.1.0")}}
	}
	d, err := New(codecs)
	if err != nil {
		t.Fatal("fixture decoder configuration failed")
	}
	return d
}

var accepted = time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)

func timing() validation.TimingContext {
	issued := accepted
	return validation.TimingContext{Origin: "remote", AcceptedAt: accepted, IssuedAt: &issued}
}

const baseJSON = `{"schema_version":"0.1.0","task_id":"synthetic-task","test_id":"SYN-01","test_version":"0.1.0","input":{"subject":"EXAMPLE"},"options":{},"schedule":{"not_after":"2026-10-15T00:00:00Z"},"execution":{"limits":{}},"safety":{"risk_level":"low","requires_user_consent":false}}`

func fixtureMap() map[string]any {
	var v map[string]any
	if err := json.Unmarshal([]byte(baseJSON), &v); err != nil {
		panic("bad fixture")
	}
	return v
}
func change(path string, value any, remove bool) []byte {
	v := fixtureMap()
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
	if remove {
		delete(obj, parts[len(parts)-1])
	} else {
		obj[parts[len(parts)-1]] = value
	}
	b, err := json.Marshal(v)
	if err != nil {
		panic("fixture encode")
	}
	return b
}

func TestDefaultsAndDeterminism(t *testing.T) {
	d := fixtureDecoder(t)
	a, err := d.Decode([]byte(baseJSON), timing())
	if err != nil {
		t.Fatal("valid fixture rejected")
	}
	if a.Execution.NetworkPath != "direct" || a.Execution.Network.VPNPolicy != "any" || a.Execution.ConcurrencyMode != "" || a.Execution.Limits != (task.AttemptLimits{DurationMS: 30000, BytesSent: 32768, BytesReceived: 131072, Connections: 8}) || a.Schedule.Repeat.Mode != "none" || a.Schedule.Retry.MaxAttempts != 1 || *a.Schedule.MaxRuns != 1 {
		t.Fatal("documented defaults differ")
	}
	if a.Parameters != (fixtureParameters{Subject: "example", Enabled: true}) {
		t.Fatal("test-owned normalization differs")
	}
	b, err := d.Decode(change("options.enabled", false, false), timing())
	if err != nil || b.Parameters.(fixtureParameters).Enabled {
		t.Fatal("explicit false became default")
	}
	b, err = d.Decode(change("schedule.starts_at", nil, false), timing())
	if err != nil || !reflect.DeepEqual(a, b) {
		t.Fatal("nullable immediate start differs from omission")
	}
	b, err = d.Decode(change("labels", map[string]string{}, false), timing())
	if err != nil || !reflect.DeepEqual(a, b) {
		t.Fatal("empty/omitted labels differ")
	}
	b, err = d.Decode(change("labels", map[string]string{"z": "last", "a": "first"}, false), timing())
	if err != nil || b.Labels[0].Key != "a" {
		t.Fatal("labels not deterministically ordered")
	}
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			b, err := d.Decode([]byte(baseJSON), timing())
			if err != nil || !reflect.DeepEqual(a, b) {
				t.Error("concurrent normalization differs")
			}
		})
	}
	wg.Wait()
}

// requirement-ref: REQ-JSON-0008
func TestExactSchemaSelection(t *testing.T) {
	var oldCalls, newCalls atomic.Int32
	d := fixtureDecoder(t, fixtureCodec{def: fixtureDefinition("SYN-01", "0.1.0"), calls: &oldCalls}, fixtureCodec{def: fixtureDefinition("SYN-01", "0.2.0"), calls: &newCalls})
	if c, err := d.Decode(change("test_version", "0.2.0", false), timing()); c != nil || !errors.Is(err, jsondecode.ErrField) || oldCalls.Load() != 0 || newCalls.Load() != 0 {
		t.Fatal("selected schema failure fell back or published candidate")
	}
	v := fixtureMap()
	v["test_version"] = "0.2.0"
	v["input"] = map[string]any{"target": "NEW"}
	b, _ := json.Marshal(v)
	c, err := d.Decode(b, timing())
	if err != nil || c.Parameters.(fixtureParameters).Subject != "new" || oldCalls.Load() != 0 || newCalls.Load() != 1 {
		t.Fatal("wrong exact codec selected")
	}
	for _, pair := range [][2]string{{"SYN-01", "0.3.0"}, {"SYN-02", "0.1.0"}, {"syn-01", "0.1.0"}} {
		v := fixtureMap()
		v["test_id"] = pair[0]
		v["test_version"] = pair[1]
		b, _ := json.Marshal(v)
		if c, err := d.Decode(b, timing()); c != nil || !errors.Is(err, jsondecode.ErrVersion) {
			t.Fatal("unsupported exact tuple accepted")
		}
	}
}

// requirement-ref: REQ-TASK-0001
func TestRejectionCorpus(t *testing.T) {
	type entry struct {
		name, path string
		value      any
		remove     bool
		expected   error
	}
	entries := []entry{
		{"unknown-top", "extra", true, false, jsondecode.ErrField},
		{"schema-version", "schema_version", "0.2.0", false, jsondecode.ErrVersion},
		{"version-type", "test_version", 1, false, jsondecode.ErrVersion},
		{"identifier-empty", "task_id", "", false, validation.ErrTask},
		{"identifier-case", "task_id", "Task", false, validation.ErrTask},
		{"identifier-first", "task_id", "_task", false, validation.ErrTask},
		{"identifier-long", "task_id", strings.Repeat("a", 129), false, validation.ErrTask},
		{"schedule-expired", "schedule.not_after", "2026-10-14T00:00:00Z", false, validation.ErrTiming},
		{"schedule-over-week", "schedule.not_after", "2026-10-21T00:00:00.000000001Z", false, validation.ErrTiming},
		{"schedule-start-equal", "schedule.starts_at", "2026-10-15T00:00:00Z", false, validation.ErrTask},
		{"schedule-local-offset", "schedule.starts_at", "2026-10-14T01:00:00+01:00", false, validation.ErrTask},
		{"schedule-invalid-date", "schedule.starts_at", "2026-02-30T00:00:00Z", false, validation.ErrTask},
		{"schedule-comma-fraction", "schedule.starts_at", "2026-10-14T00:00:00,5Z", false, validation.ErrTask},
		{"schedule-null-expiry", "schedule.not_after", nil, false, jsondecode.ErrType},
		{"schedule-no-expiry", "schedule.not_after", nil, true, validation.ErrTiming},
		{"repeat-unknown", "schedule.repeat.mode", "weekly", false, validation.ErrTask},
		{"repeat-none-extra", "schedule.repeat.interval", 21600, false, validation.ErrTask},
		{"repeat-missing-fields", "schedule.repeat.mode", "interval", false, validation.ErrTask},
		{"retry-empty", "schedule.retry", map[string]any{}, false, jsondecode.ErrRequired},
		{"retry-one-extra", "schedule.retry", map[string]any{"max_attempts": 1, "retry_on": []string{}}, false, validation.ErrTask},
		{"retry-two-missing", "schedule.retry", map[string]any{"max_attempts": 2}, false, validation.ErrTask},
		{"runs-zero", "schedule.limits.max_runs", 0, false, jsondecode.ErrRange},
		{"runs-high", "schedule.limits.max_runs", 29, false, jsondecode.ErrRange},
		{"duration-zero", "execution.limits.max_duration_ms", 0, false, jsondecode.ErrRange},
		{"duration-high", "execution.limits.max_duration_ms", 60001, false, jsondecode.ErrRange},
		{"bytes-sent-zero", "execution.limits.max_bytes_sent", 0, false, validation.ErrTask},
		{"bytes-sent-high", "execution.limits.max_bytes_sent", 65537, false, jsondecode.ErrRange},
		{"bytes-received-high", "execution.limits.max_bytes_received", 262145, false, jsondecode.ErrRange},
		{"connections-high", "execution.limits.max_connections", 13, false, jsondecode.ErrRange},
		{"negative-limit", "execution.limits.max_connections", -1, false, jsondecode.ErrRange},
		{"quoted-limit", "execution.limits.max_connections", "8", false, jsondecode.ErrType},
		{"network-path", "execution.network_path", "fallback", false, validation.ErrTask},
		{"vpn-policy", "execution.network.vpn_policy", "unknown", false, validation.ErrTask},
		{"transport", "execution.network.allowed_transports", []string{"ethernet"}, false, validation.ErrTask},
		{"context-field", "execution.network.context.required_fields", []string{"provider"}, false, validation.ErrTask},
		{"context-duplicates", "execution.network.context.required_fields", []string{"asn", "asn"}, false, validation.ErrTask},
		{"country-empty", "execution.network.context.allowed_country_codes", []string{}, false, validation.ErrTask},
		{"country-case", "execution.network.context.allowed_country_codes", []string{"ru"}, false, validation.ErrTask},
		{"country-unassigned", "execution.network.context.allowed_country_codes", []string{"ZZ"}, false, validation.ErrTask},
		{"country-duplicate", "execution.network.context.allowed_country_codes", []string{"RU", "RU"}, false, validation.ErrTask},
		{"battery-high", "execution.device.minimum_battery_percent", 101, false, jsondecode.ErrRange},
		{"concurrency", "execution.concurrency.mode", "parallel", false, validation.ErrTask},
		{"concurrency-empty", "execution.concurrency.mode", "", false, validation.ErrTask},
		{"risk-none", "safety.risk_level", "none", false, validation.ErrTask},
		{"risk-no-factors", "safety.risk_level", "medium", false, validation.ErrTask},
		{"risk-downgrade", "safety.risk_factors", []any{map[string]any{"type": "traffic_distinctiveness", "level": "high"}}, false, validation.ErrTask},
		{"risk-unknown-factor", "safety.risk_factors", []any{map[string]any{"type": "traffic_distinctiveness", "level": "unknown"}}, false, validation.ErrTask},
		{"risk-factor-type", "safety.risk_factors", []any{map[string]any{"type": "unexpected", "level": "low"}}, false, validation.ErrTask},
		{"label-case", "labels", map[string]any{"Campaign": "x"}, false, validation.ErrTask},
		{"label-reserved", "labels", map[string]any{"barometer.example": "x"}, false, validation.ErrTask},
		{"label-value-type", "labels", map[string]any{"campaign": true}, false, jsondecode.ErrType},
		{"label-control", "labels", map[string]any{"campaign": "a\nb"}, false, validation.ErrTask},
		{"label-length", "labels", map[string]any{"campaign": strings.Repeat("x", 129)}, false, validation.ErrTask},
	}
	for _, path := range []string{"schema_version", "task_id", "test_id", "test_version", "input", "options", "schedule", "execution", "safety", "execution.limits", "safety.risk_level", "safety.requires_user_consent"} {
		expected := jsondecode.ErrRequired
		if path == "schema_version" || path == "test_id" || path == "test_version" {
			expected = jsondecode.ErrVersion
		}
		entries = append(entries, entry{"missing-" + path, path, nil, true, expected})
	}
	for _, path := range []string{"input", "options", "schedule", "execution", "safety", "execution.limits", "labels"} {
		entries = append(entries, entry{"null-" + path, path, nil, false, jsondecode.ErrType})
	}
	for _, path := range []string{"schedule", "schedule.repeat", "schedule.retry", "schedule.limits", "execution", "execution.limits", "execution.network", "execution.network.context", "execution.device", "execution.concurrency", "safety", "input", "options"} {
		entries = append(entries, entry{"unknown-" + path, path + ".unrecognized", true, false, jsondecode.ErrField})
	}
	var calls atomic.Int32
	d := fixtureDecoder(t, fixtureCodec{def: fixtureDefinition("SYN-01", "0.1.0"), calls: &calls})
	for _, e := range entries {
		t.Run(e.name, func(t *testing.T) {
			b := change(e.path, e.value, e.remove)
			candidate, err := d.Decode(b, timing())
			testkit.AssertRejected(t, testkit.RejectionCase{Name: e.name, Requirement: "Draft 0.1.0 task common schema", Expected: "registered rejection", Input: b}, err, func(_ string, err error) bool { return errors.Is(err, e.expected) })
			if candidate != nil {
				t.Fatal("partial candidate escaped")
			}
		})
	}
	if calls.Load() != 0 {
		t.Fatal("invalid common task reached test normalization")
	}
}

func TestRepresentationAndSanitizedErrors(t *testing.T) {
	d := fixtureDecoder(t)
	for _, replacement := range []string{`-0`, `1.0`, `1e0`, `9007199254740993`, `999999999999999999999999999999999999`} {
		b := strings.Replace(baseJSON, `"limits":{}`, `"limits":{"max_connections":`+replacement+`}`, 1)
		if c, err := d.Decode([]byte(b), timing()); c != nil || err == nil {
			t.Fatal("invalid exact integer accepted")
		}
	}
	for _, b := range []string{baseJSON + ` {}`, strings.Replace(baseJSON, `"task_id":"synthetic-task"`, `"task_id":"synthetic-task","task_\u0069d":"other"`, 1), strings.Replace(baseJSON, `"subject":"EXAMPLE"`, `"subject":"EXAMPLE","subject":"EXAMPLE"`, 1), strings.Replace(baseJSON, `"options":{}`, `"options":{"enabled":true,"enabled":false}`, 1)} {
		if c, err := d.Decode([]byte(b), timing()); c != nil || err == nil {
			t.Fatal("duplicate/framing failure accepted")
		}
	}
	d = fixtureDecoder(t, fixtureCodec{def: fixtureDefinition("SYN-01", "0.1.0"), fail: true})
	if c, err := d.Decode([]byte(baseJSON), timing()); c != nil || !errors.Is(err, ErrParameters) || strings.Contains(err.Error(), "sensitive") {
		t.Fatal("codec error leaked or was not atomic")
	}
}

func TestByteAndLabelBounds(t *testing.T) {
	d := fixtureDecoder(t)
	for _, n := range []int{65535, 65536, 65537} {
		b := []byte(baseJSON + strings.Repeat(" ", n-len(baseJSON)))
		c, err := d.Decode(b, timing())
		if n <= 65536 {
			if err != nil || c == nil {
				t.Fatal("registered complete byte boundary rejected")
			}
		} else if c != nil || !errors.Is(err, jsondecode.ErrBounds) {
			t.Fatal("complete byte bound not enforced")
		}
	}
	for _, n := range []int{49151, 49152, 49153} {
		// Original encoded input bytes include the quoted subject and braces.
		raw := `{"subject":"` + strings.Repeat("a", n-len(`{"subject":""}`)) + `"}`
		b := strings.Replace(baseJSON, `{"subject":"EXAMPLE"}`, raw, 1)
		c, err := d.Decode([]byte(b), timing())
		if (n <= 49152) != (err == nil) || (err != nil && c != nil) {
			t.Fatal("input byte boundary differs")
		}
	}
	for _, n := range []int{8191, 8192, 8193} {
		raw := `{` + strings.Repeat(" ", n-2) + `}`
		b := strings.Replace(baseJSON, `"options":{}`, `"options":`+raw, 1)
		c, err := d.Decode([]byte(b), timing())
		if (n <= 8192) != (err == nil) || (err != nil && c != nil) {
			t.Fatal("original option whitespace not counted")
		}
	}
	for _, n := range []int{2047, 2048, 2049} {
		raw := `{` + strings.Repeat(" ", n-2) + `}`
		b := strings.TrimSuffix(baseJSON, "}") + `,"labels":` + raw + `}`
		c, err := d.Decode([]byte(b), timing())
		if (n <= 2048) != (err == nil) || (err != nil && c != nil) {
			t.Fatal("label byte boundary differs")
		}
	}
	labels := map[string]string{}
	for i := 0; i < 17; i++ {
		labels["label"+string(rune('a'+i))] = "x"
	}
	if c, err := d.Decode(change("labels", labels, false), timing()); c != nil || !errors.Is(err, jsondecode.ErrBounds) {
		t.Fatal("label count bound not enforced")
	}
	delete(labels, "labelq")
	if _, err := d.Decode(change("labels", labels, false), timing()); err != nil {
		t.Fatal("maximum label count rejected")
	}
}

func TestDescriptorConfiguration(t *testing.T) {
	def := fixtureDefinition("SYN-01", "0.1.0")
	for _, codecs := range [][]TestCodec{nil, {nil}, {fixtureCodec{def: def}, fixtureCodec{def: def}}, {fixtureCodec{def: Definition{TestID: "SYN-01", TestVersion: "0.1.0"}}}} {
		if d, err := New(codecs); d != nil || !errors.Is(err, jsondecode.ErrConfig) {
			t.Fatal("bad codec configuration accepted")
		}
	}
	bad := def
	bad.Input = &jsondecode.Schema{Kind: jsondecode.Array, Element: &jsondecode.Schema{Kind: jsondecode.String}}
	if d, err := New([]TestCodec{fixtureCodec{def: bad}}); d != nil || !errors.Is(err, jsondecode.ErrConfig) {
		t.Fatal("non-object parameters accepted")
	}
	bad = def
	bad.Input = object(map[string]jsondecode.Field{"value": field(&jsondecode.Schema{Kind: jsondecode.Integer}, true)})
	if d, err := New([]TestCodec{fixtureCodec{def: bad}}); d != nil || !errors.Is(err, jsondecode.ErrConfig) {
		t.Fatal("unbounded parameter integer accepted")
	}
}

func TestNET01ZeroTrafficLimits(t *testing.T) {
	d := fixtureDecoder(t, fixtureCodec{def: fixtureDefinition("NET-01", "0.1.0")})
	v := fixtureMap()
	v["test_id"] = "NET-01"
	v["input"] = map[string]any{}
	b, _ := json.Marshal(v)
	c, err := d.Decode(b, timing())
	if err != nil || c.Execution.Limits != (task.AttemptLimits{DurationMS: 30000}) {
		t.Fatal("NET-01 zero traffic defaults differ")
	}
	v["execution"] = map[string]any{"limits": map[string]any{"max_bytes_sent": 1}}
	b, _ = json.Marshal(v)
	if c, err := d.Decode(b, timing()); c != nil || err == nil {
		t.Fatal("NET-01 acquired network budget")
	}
}

func FuzzTaskDecode(f *testing.F) {
	d := fixtureDecoder(f)
	f.Add([]byte(baseJSON))
	f.Add(change("schedule.repeat.mode", "interval", false))
	f.Add([]byte(`{}`))
	f.Fuzz(func(t *testing.T, b []byte) {
		c, err := d.Decode(b, timing())
		second, other := d.Decode(b, timing())
		if !reflect.DeepEqual(c, second) || !errors.Is(err, other) {
			t.Fatal("decoding not deterministic")
		}
		if err != nil && (c != nil || len(err.Error()) > 64) {
			t.Fatal("unbounded or partial failure")
		}
	})
}
