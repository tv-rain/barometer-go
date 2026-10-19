package taskdecode

import (
	"errors"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/validation"
)

// Envelope defers field values and limits to Fields and input and options to
// Parameters, so a caller can apply the public step order. Decode composes
// the three stages.
// requirement-ref: REQ-JSON-0008
func TestStagedDecoding(t *testing.T) {
	var calls atomic.Int32
	d := fixtureDecoder(t, fixtureCodec{def: fixtureDefinition("SYN-01", "0.1.0"), calls: &calls})
	over := change("execution.limits.max_duration_ms", 60001, false)
	e, err := d.Envelope(over)
	if err != nil || e.TaskID() != "synthetic-task" || e.Test().ID != "SYN-01" || e.Test().Version != "0.1.0" {
		t.Fatal("envelope applied a field limit or lost the selection")
	}
	if _, err = e.Fields(timing()); !errors.Is(err, jsondecode.ErrLimit) || !errors.Is(err, jsondecode.ErrRange) {
		t.Fatal("field limit not reported as a limit")
	}
	if _, err = d.Envelope(change("execution.limits.max_duration_ms", "x", false)); !errors.Is(err, jsondecode.ErrType) {
		t.Fatal("envelope accepted a wrong value kind")
	}
	badInput := change("input", map[string]any{"other": "x"}, false)
	if e, err = d.Envelope(badInput); err != nil {
		t.Fatal("envelope validated test-specific input")
	}
	if c, err := e.Fields(timing()); err != nil || c.Parameters != nil {
		t.Fatal("fields validated test-specific input or set parameters")
	}
	if _, err = e.Parameters(); !errors.Is(err, jsondecode.ErrField) || calls.Load() != 0 {
		t.Fatal("parameters skipped the selected input schema")
	}
	for input, want := range map[string]error{
		strings.Replace(baseJSON, `"test_version":"0.1.0"`, `"test_version":"0.3.0"`, 1):     ErrUnregistered,
		strings.Replace(baseJSON, `"test_id":"SYN-01"`, `"test_id":"SYN-02"`, 1):             ErrUnregistered,
		strings.Replace(baseJSON, `"test_version":"0.1.0"`, `"test_version":"0.1"`, 1):       jsondecode.ErrVersion,
		strings.Replace(baseJSON, `"schema_version":"0.1.0"`, `"schema_version":"0.2.0"`, 1): jsondecode.ErrVersion,
	} {
		_, err := d.Envelope([]byte(input))
		if !errors.Is(err, want) || !errors.Is(err, jsondecode.ErrVersion) || errors.Is(err, ErrUnregistered) != (want == ErrUnregistered) || err.Error() != want.Error() {
			t.Error("version selection class mismatch")
		}
	}
	expired := timing()
	expired.AcceptedAt = expired.AcceptedAt.AddDate(0, 0, 2)
	if _, err = d.Decode(badInput, expired); !errors.Is(err, validation.ErrExpired) {
		t.Fatal("decode did not apply expiry before test-specific input")
	}
}
