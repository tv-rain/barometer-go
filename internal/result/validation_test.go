package result

import (
	"bytes"
	"errors"
	"testing"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	model "github.com/tv-rain/barometer-go/internal/model/result"
)

// requirement-ref: REQ-JSON-0010
func TestUninitializedDecoderRejectsBeforeInput(t *testing.T) {
	for _, d := range []*Decoder{nil, {}} {
		for _, input := range [][]byte{nil, []byte(`{}`), envelopeJSON("TCP-01", "completed", "", payload("connected", tcp("succeeded", "")))} {
			v, err := d.Decode(input)
			if !errors.Is(err, ErrBinding) {
				t.Fatalf("uninitialized decoder: %v", err)
			}
			if _, err := v.MarshalJSON(); err == nil {
				t.Fatal("usable result escaped")
			}
		}
		if _, err := d.Build(model.Result{}); !errors.Is(err, ErrBinding) {
			t.Fatalf("uninitialized producer: %v", err)
		}
	}
}

// requirement-ref: REQ-JSON-0008
// requirement-ref: REQ-JSON-0011
func TestResultValidationPrecedence(t *testing.T) {
	d := decoder(t, binding("TCP-01"))
	base := envelopeJSON("TCP-01", "completed", "", payload("connected", tcp("succeeded", "")))
	for _, tc := range []struct {
		name  string
		input []byte
		want  error
	}{
		{"complete size before syntax", bytes.Repeat([]byte("!"), 65537), jsondecode.ErrBounds},
		{"syntax before version", []byte(`{"schema_version":"9.0.0"`), jsondecode.ErrSyntax},
		{"version before fields", []byte(`{"schema_version":"9.0.0","private":"DO_NOT_RETAIN"}`), jsondecode.ErrVersion},
		{"structure before binding", mutate(t, base, func(v map[string]any) { v["task_id"] = "different"; v["attempt"] = "1" }), jsondecode.ErrType},
		{"test version before values", mutate(t, base, func(v map[string]any) { object(v, "test")["result_version"] = "9.0.0"; v["task_id"] = "different" }), jsondecode.ErrVersion},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for range 30 {
				if err := reject(t, d, tc.input); !errors.Is(err, tc.want) || len(err.Error()) > 128 {
					t.Fatalf("precedence: %v", err)
				}
			}
		})
	}
}
