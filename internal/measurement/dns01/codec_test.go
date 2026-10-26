package dns01

import (
	"encoding/json"
	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"strings"
	"testing"
)

func normalize(t *testing.T, input, options string) (Parameters, error) {
	t.Helper()
	c := Codec{}
	d := c.Definition()
	limits := jsondecode.Limits{MaxBytes: 65536, MaxDepth: 16, MaxNodes: 1024}
	i, e := jsondecode.Decode([]byte(input), d.Input, limits)
	if e != nil {
		return Parameters{}, e
	}
	o, e := jsondecode.Decode([]byte(options), d.Options, limits)
	if e != nil {
		return Parameters{}, e
	}
	p, e := c.Normalize(i, o)
	if e != nil {
		return Parameters{}, e
	}
	return p.(Parameters), nil
}

// Full closed field/type and conditional matrix, separate from authority tests.
// requirement-ref: REQ-DNS-01-0014
// requirement-ref: REQ-DNS-01-0015
// requirement-ref: REQ-DNS-01-0025
// requirement-ref: REQ-DNS-01-0026
func TestCodecFieldTypesAndModeRelations(t *testing.T) {
	base := `{"hostname":"example.org","record_type":"A","resolver":{"ip_address":"8.8.8.8"}}`
	for _, field := range []string{"hostname", "record_type", "resolver", "resolver.ip_address", "resolver.port"} {
		for _, value := range []any{nil, true, 3, []any{}, map[string]any{}} {
			if field == "resolver.port" && value == 3 {
				continue
			}
			var input map[string]any
			_ = json.Unmarshal([]byte(base), &input)
			if strings.HasPrefix(field, "resolver.") {
				input["resolver"].(map[string]any)[strings.TrimPrefix(field, "resolver.")] = value
			} else {
				input[field] = value
			}
			b, _ := json.Marshal(input)
			if _, e := normalize(t, string(b), `{"resolver_mode":"direct"}`); e == nil {
				t.Fatalf("accepted %s %s", field, b)
			}
		}
	}
	for _, field := range []string{"resolver_mode", "transport", "collect_additional_responses", "additional_response_window_ms"} {
		for _, value := range []any{nil, []any{}, map[string]any{}} {
			options := map[string]any{"resolver_mode": "direct", field: value}
			b, _ := json.Marshal(options)
			if _, e := normalize(t, base, string(b)); e == nil {
				t.Fatalf("accepted %s", b)
			}
		}
	}
	for _, options := range []string{`{"resolver_mode":"direct"}`, `{"resolver_mode":"direct","collect_additional_responses":false}`, `{"resolver_mode":"direct","collect_additional_responses":true,"additional_response_window_ms":1}`, `{"resolver_mode":"direct","collect_additional_responses":true,"additional_response_window_ms":1000}`} {
		p, e := normalize(t, base, options)
		if e != nil || p.port != 53 || p.transport != "udp" {
			t.Fatal(options, e)
		}
	}
	system := `{"hostname":"example.org","record_type":"AAAA"}`
	for _, options := range []string{`{}`, `{"resolver_mode":"system"}`, `{"collect_additional_responses":false}`} {
		p, e := normalize(t, system, options)
		if e != nil || p.Engine() != "system" {
			t.Fatal(options, e)
		}
	}
	for _, options := range []string{`{"transport":"udp"}`, `{"transport":"tcp"}`, `{"collect_additional_responses":true}`, `{"additional_response_window_ms":1}`} {
		if _, e := normalize(t, system, options); e == nil {
			t.Fatal("accepted system options", options)
		}
	}
	if _, e := normalize(t, base, `{}`); e == nil {
		t.Fatal("system resolver input accepted")
	}
	if _, e := (Codec{}).Normalize(nil, nil); e == nil {
		t.Fatal("nil accepted")
	}
}

// requirement-ref: REQ-DNS-01-0016
func TestCodecNormalizedNameBoundaries(t *testing.T) {
	for _, length := range []int{252, 253, 254} {
		name := strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", length-192)
		input := `{"hostname":"` + name + `","record_type":"A"}`
		_, e := normalize(t, input, `{}`)
		if (e == nil) != (length <= 253) {
			t.Fatalf("%d %v", length, e)
		}
	}
	for _, s := range []string{"bücher.example.org", "xn--bcher-kva.example.org", "BÜCHER.example.org"} {
		p, e := normalize(t, `{"hostname":"`+s+`","record_type":"A"}`, `{}`)
		if e != nil || p.name != "xn--bcher-kva.example.org" {
			t.Fatal(s, e)
		}
	}
	for _, s := range []string{"xn--.org", "xn--a.org", "a\u200db.org", "a b.org", "a/b.org", "x@a.org"} {
		if _, e := normalize(t, `{"hostname":"`+s+`","record_type":"A"}`, `{}`); e == nil {
			t.Fatal(s)
		}
	}
}
