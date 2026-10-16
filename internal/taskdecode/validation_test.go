package taskdecode

import (
	"errors"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/validation"
)

// requirement-ref: REQ-JSON-0008
// requirement-ref: REQ-JSON-0011
func TestModelValidationPrecedenceAndExactCodec(t *testing.T) {
	var oldCalls, newCalls atomic.Int32
	d := fixtureDecoder(t,
		fixtureCodec{def: fixtureDefinition("SYN-01", "0.1.0"), calls: &oldCalls},
		fixtureCodec{def: fixtureDefinition("SYN-01", "0.2.0"), calls: &newCalls, fail: true},
	)
	for _, tc := range []struct {
		name, input string
		want        error
		calls       int32
	}{
		{"syntax before selection", `{"test_version":"9.0.0"`, jsondecode.ErrSyntax, 0},
		{"selection before shape", `{"schema_version":"0.1.0","test_id":"SYN-01","test_version":"9.0.0","private":"DO_NOT_RETAIN"}`, jsondecode.ErrVersion, 0},
		{"shape before common", string(change("task_id", nil, false)), jsondecode.ErrType, 0},
		{"common before timing", string(change("safety.risk_level", "unsupported", false)), validation.ErrTask, 0},
		{"timing before codec", baseJSON, validation.ErrTiming, 0},
		{"no older schema fallback", strings.Replace(baseJSON, `"test_version":"0.1.0"`, `"test_version":"0.2.0"`, 1), jsondecode.ErrField, 0},
		{"selected semantic rejection", strings.Replace(strings.Replace(baseJSON, `"test_version":"0.1.0"`, `"test_version":"0.2.0"`, 1), `"subject":`, `"target":`, 1), ErrParameters, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for range 20 {
				oldCalls.Store(0)
				newCalls.Store(0)
				ctx := timing()
				if tc.name != "selected semantic rejection" {
					ctx.AcceptedAt = ctx.AcceptedAt.AddDate(0, 0, 2)
				}
				v, err := d.Decode([]byte(tc.input), ctx)
				if !errors.Is(err, tc.want) || v != nil || len(err.Error()) > 128 {
					t.Fatalf("precedence: %v", err)
				}
				if oldCalls.Load() != 0 || newCalls.Load() != tc.calls {
					t.Fatal("wrong codec invoked or earlier validation bypassed")
				}
			}
		})
	}
	valid := []byte(baseJSON[:len(baseJSON)-1] + `,"labels":{"z":"${literal}","a":"Я"}}`)
	v, err := d.Decode(valid, timing())
	if err != nil {
		t.Fatal(err)
	}
	normalized, err := validation.Common(*v)
	if err != nil || !reflect.DeepEqual(*v, normalized) {
		t.Fatal("common validation changed decoded model")
	}
	if v.TestVersion != "0.1.0" || v.Parameters != (fixtureParameters{Subject: "example", Enabled: true}) {
		t.Fatal("test-specific values changed")
	}
}
