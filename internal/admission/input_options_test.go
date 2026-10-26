package admission_test

import (
	"errors"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/taskdecode"
)

// classCodec is a synthetic test-scoped SYN-01 codec with the option shapes the
// step-8 class rows need: a field-table range (count), a registered numeric
// limit (window_ms), a value with a stricter test byte cap (tag) and a stricter
// test byte cap on the whole options object. It is not the codec of any
// registered test.
type classCodec struct{ calls *atomic.Int32 }

func (c classCodec) Definition() taskdecode.Definition {
	text := &jsondecode.Schema{Kind: jsondecode.String}
	return taskdecode.Definition{TestID: "SYN-01", TestVersion: "0.1.0",
		Input: &jsondecode.Schema{Kind: jsondecode.Object, Fields: map[string]jsondecode.Field{"subject": {Schema: text, Required: true}}},
		Options: &jsondecode.Schema{Kind: jsondecode.Object, MaxBytes: 64, Limit: true, Fields: map[string]jsondecode.Field{
			"note":      {Schema: text},
			"count":     {Schema: &jsondecode.Schema{Kind: jsondecode.Integer, Min: "1", Max: "5"}},
			"window_ms": {Schema: &jsondecode.Schema{Kind: jsondecode.Integer, Min: "1", Max: "1000", Limit: true}},
			"tag":       {Schema: &jsondecode.Schema{Kind: jsondecode.Object, Additional: &jsondecode.Schema{Kind: jsondecode.Any}, MaxBytes: 16, Limit: true}},
		}}}
}

func (c classCodec) Normalize(input, _ *jsondecode.Value) (task.Parameters, error) {
	c.calls.Add(1)
	subject, _ := input.Lookup("subject")
	if subject.Text() == "ambiguous" {
		return nil, errors.New("CANARY-CODEC-ERROR")
	}
	return synParameters{Subject: subject.Text()}, nil
}

// A task whose input or options fail validation against the selected test is
// task_schema_invalid at step 8, and a stricter test byte cap on options or a
// value within them is task_limit_exceeded, checked before the content it
// bounds (REQ-TRUST-0056). Invalid options are rejected at admission, never
// admitted to be skipped (REQ-OPTIONS-0033). An option above a registered
// numeric limit is task_limit_exceeded, and invalid input precedes over-limit
// options; both are interim handling under GO-SPEC-ISSUE-016. The codec is not
// reached for a schema or limit failure, and the harness transport sees zero
// calls.
// requirement-ref: REQ-TRUST-0056
// requirement-ref: REQ-OPTIONS-0033
// requirement-ref: REQ-TRUST-0047
// requirement-ref: REQ-LIMIT-0002
func TestGateInputOptionsClasses(t *testing.T) {
	for _, c := range []struct {
		name  string
		task  []byte
		want  task.AdmissionCode
		codec bool // whether the codec must be reached
	}{
		{"valid options", taskWith(map[string]any{"options.count": 5, "options.window_ms": 1000, "options.tag": map[string]any{"k": "v"}}), 0, true},
		{"unknown option", taskWith(map[string]any{"options.extra": true}), task.SchemaInvalid, false},
		{"unknown input member", taskWith(map[string]any{"input.extra": "x"}), task.SchemaInvalid, false},
		{"option below its range", taskWith(map[string]any{"options.count": 0}), task.SchemaInvalid, false},
		{"option above its range", taskWith(map[string]any{"options.count": 6}), task.SchemaInvalid, false},
		{"option of the wrong type", taskWith(map[string]any{"options.count": "1"}), task.SchemaInvalid, false},
		{"input the codec cannot normalize", taskWith(map[string]any{"input.subject": "ambiguous"}), task.SchemaInvalid, true},
		{"options at the test cap", taskWith(map[string]any{"options.note": strings.Repeat("n", 64-len(`{"note":""}`))}), 0, true},
		{"options over the test cap", taskWith(map[string]any{"options.note": strings.Repeat("n", 65-len(`{"note":""}`))}), task.LimitExceeded, false},
		{"invalid options over the test cap", taskWith(map[string]any{"options.extra": strings.Repeat("x", 64)}), task.LimitExceeded, false},
		{"option value over its test cap", taskWith(map[string]any{"options.tag": map[string]any{"k": strings.Repeat("v", 16)}}), task.LimitExceeded, false},
		{"option value of the wrong type within its test cap", taskWith(map[string]any{"options.tag": "v"}), task.SchemaInvalid, false},
		{"option value of the wrong type over its test cap", taskWith(map[string]any{"options.tag": strings.Repeat("v", 16)}), task.LimitExceeded, false},
		{"option at a registered limit", taskWith(map[string]any{"options.window_ms": 1000}), 0, true},
		{"option above a registered limit", taskWith(map[string]any{"options.window_ms": 1001}), task.LimitExceeded, false},
		{"invalid input before over-limit options", taskWith(map[string]any{"input.extra": "x", "options.window_ms": 1001}), task.SchemaInvalid, false},
	} {
		h := newHarness(t)
		entry := synEntry(h, task.LowRisk)
		entry.Codec = classCodec{calls: h.normalize}
		rebind(t, h, mustPolicy(t, permissive()), entry)
		_, err := h.embedded(t, c.task)
		expect(t, c.name, err, c.want)
		if reached := h.normalize.Load() > 0; reached != c.codec {
			t.Errorf("%s: codec reached %v, want %v", c.name, reached, c.codec)
		}
	}

	// Remote origin takes the same rows.
	h := newHarness(t)
	entry := synEntry(h, task.LowRisk)
	entry.Codec = classCodec{calls: h.normalize}
	rebind(t, h, mustPolicy(t, permissive()), entry)
	_, err := h.remote(t, 0, taskWith(map[string]any{"options.count": 6}))
	expect(t, "remote option above its range", err, task.SchemaInvalid)
	_, err = h.remote(t, 0, taskWith(map[string]any{"options.window_ms": 1001}))
	expect(t, "remote option above a registered limit", err, task.LimitExceeded)
}
