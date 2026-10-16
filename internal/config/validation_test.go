package config_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/tv-rain/barometer-go/internal/config"
	"github.com/tv-rain/barometer-go/internal/jsondecode"
)

// requirement-ref: REQ-JSON-0010
// requirement-ref: REQ-JSON-0011
func TestConfigurationValidationPrecedence(t *testing.T) {
	d := decoder(t)
	for _, tc := range []struct {
		name, input string
		want        error
	}{
		{"size before syntax", strings.Repeat("!", config.MaxBytes+1), jsondecode.ErrBounds},
		{"shape before semantics", `{"proxy":{"host":"public.example","port":1,"connection_code":"unapproved"},"result_linkage":null}`, jsondecode.ErrType},
		{"proxy before endpoint", `{"network_context":{"primary_endpoint":"invalid"},"proxy":{"host":"localhost","port":1,"connection_code":"unapproved"}}`, config.ErrApproval},
		{"endpoint before linkage", `{"result_linkage":{"mode":"unsupported"},"network_context":{"primary_endpoint":"https://unapproved.example/context"}}`, config.ErrApproval},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for range 20 {
				v, err := d.Decode([]byte(tc.input))
				if v.Valid() || !errors.Is(err, tc.want) || len(err.Error()) > 128 {
					t.Fatalf("precedence: %v", err)
				}
			}
		})
	}
}

// Versionless configuration has no snapshot serializer. Reconstruct the input
// through its reviewed field names and copied views, never private state.
func TestConfigurationProjectionRoundTrip(t *testing.T) {
	for _, input := range []string{`{}`, complete, `{"result_linkage":{"mode":"none"}}`} {
		t.Run(input, func(t *testing.T) {
			d := decoder(t)
			first := decode(t, d, input)
			wire := map[string]any{}
			if p, ok := first.Proxy(); ok {
				wire["proxy"] = map[string]any{"host": p.Host, "port": p.Port, "connection_code": p.ConnectionCode}
			}
			if n, ok := first.NetworkContext(); ok {
				x := map[string]any{"primary_endpoint": n.PrimaryEndpoint}
				if n.FallbackEndpoint != "" {
					x["fallback_endpoint"] = n.FallbackEndpoint
				}
				wire["network_context"] = x
			}
			l := first.Linkage()
			if l.Mode == "rotating_installation" {
				wire["result_linkage"] = map[string]any{"mode": l.Mode, "scope": l.Scope, "rotation_period_seconds": l.RotationPeriodSeconds}
			}
			encoded, err := json.Marshal(wire)
			if err != nil {
				t.Fatal(err)
			}
			second := decode(t, d, string(encoded))
			if !reflect.DeepEqual(first, second) {
				t.Fatal("configuration projection changed")
			}
		})
	}
	// A policy-disabled request stays disabled when projected back to input.
	a := approvals()
	a.AllowLinkage = false
	d, err := config.NewDecoder(config.Contract, a)
	if err != nil {
		t.Fatal(err)
	}
	first := decode(t, d, complete)
	if first.Linkage() != (config.Linkage{Mode: "none"}) {
		t.Fatal("disabled linkage survived normalization")
	}
	if got := decode(t, d, `{"result_linkage":{"mode":"none"}}`).Linkage(); got != first.Linkage() {
		t.Fatal("disabled linkage changed on round trip")
	}
}
