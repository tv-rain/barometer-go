package result

import (
	"bytes"
	"testing"

	"github.com/tv-rain/barometer-go/internal/model/observation"
)

func net05(outcome, state string, steps []string, stop string) string {
	s := `{"outcome":"` + outcome + `","observations":[],"profile":{"profile_id":"baseline","profile_version":1},"network_validation":{"state":"` + state + `","source":"platform"},"steps":[`
	for i, step := range steps {
		if i > 0 {
			s += ","
		}
		s += step
	}
	s += "]"
	if stop != "" {
		s += `,"stopped_before_step":"` + stop + `"`
	}
	return s + "}"
}
func step(id, state, m string) string {
	s := `{"step_id":"` + id + `","test":{"id":"TCP-01","version":"0.1.0","result_version":"0.1.0"},"state":"` + state + `"`
	if m != "" {
		s += `,"measurement":` + m
	}
	return s + "}"
}

// requirement-ref: REQ-TCP-01-0024
// requirement-ref: REQ-TLS-01-0025
// requirement-ref: REQ-DNS-01-0037
// requirement-ref: REQ-DNS-03-0027
// requirement-ref: REQ-DNS-04-0027
// requirement-ref: REQ-HTTP-01-0036
// requirement-ref: REQ-NET-01-0023
// requirement-ref: REQ-NET-03-0027
func TestInitialPayloadSchemasAndEvidence(t *testing.T) {
	cases := []struct{ id, m string }{
		{"TCP-01", payload("connected", tcp("succeeded", ""))},
		{"TLS-01", payload("handshake_succeeded", tcp("succeeded", ""), tls("succeeded", "", "trusted"))},
		{"TLS-01", payload("certificate_untrusted", tcp("succeeded", ""), tls("failed", "certificate_untrusted", "untrusted"))},
		{"TLS-01", payload("tcp_refused", tcp("failed", "connection_refused"))},
		{"DNS-01", payload("resolved", dns("system", "succeeded", "", ""))},
		{"DNS-01", payload("nxdomain", dns("system", "failed", "nxdomain", "NXDOMAIN"))},
		{"DNS-03", payload("response", tcp("succeeded", ""), tls("succeeded", "", "trusted"), http(), dns("doh", "succeeded", "", "NOERROR"))},
		{"DNS-03", payload("tcp_failure", tcp("failed", "connection_refused"))},
		{"DNS-04", payload("dns_refused", tcp("succeeded", ""), tls("succeeded", "", "trusted"), dns("dot", "failed", "refused", "REFUSED"))},
		{"DNS-04", payload("tcp_timeout", tcp("failed", "timeout"))},
		{"HTTP-01", payload("expected_response", tcp("succeeded", ""), tls("succeeded", "", "trusted"), http())},
		{"HTTP-01", payload("tls_failure", tcp("succeeded", ""), tls("failed", "certificate_untrusted", "untrusted"))},
		{"NET-01", `{"outcome":"snapshot","observations":[],"access":` + access() + `}`},
		{"NET-01", payload("platform_unavailable")},
		{"NET-01", payload("inconclusive")},
		{"NET-03", `{"outcome":"validated","observations":[` + tcp("succeeded", "") + `,` + tls("succeeded", "", "trusted") + `,` + http() + `],"platform_validation":"unknown","response_class":"expected"}`},
		{"NET-03", `{"outcome":"connection_failure","observations":[` + tcp("failed", "connection_refused") + `],"platform_validation":"validated"}`},
	}
	for i, tc := range cases {
		t.Run(tc.id+"/"+string(rune('A'+i)), func(t *testing.T) {
			d := decoder(t, binding(tc.id))
			input := envelopeJSON(tc.id, "completed", "", tc.m)
			v, e := d.Decode(input)
			if e != nil {
				t.Fatal(e)
			}
			wire, _ := v.MarshalJSON()
			if _, e := d.Decode(wire); e != nil {
				t.Fatal("roundtrip", e)
			}
			reject(t, d, mutate(t, input, func(v map[string]any) { object(v, "measurement")["outcome"] = "censored" }))
			reject(t, d, mutate(t, input, func(v map[string]any) { object(v, "measurement")["raw_body"] = "secret" }))
		})
	}
	// Unsupported result versions cannot borrow another primitive's schema.
	d := decoder(t, binding("TCP-01"))
	reject(t, d, envelopeJSON("TCP-01", "completed", "", payload("connected", tls("succeeded", "", "trusted"))))
	d = decoder(t, binding("TLS-01"))
	reject(t, d, envelopeJSON("TLS-01", "completed", "", payload("handshake_succeeded", tls("succeeded", "", "trusted"), tcp("succeeded", ""))))
	reject(t, d, envelopeJSON("TLS-01", "completed", "", payload("timeout", tcp("succeeded", ""), tls("failed", "timeout", "not_completed"))))
	reject(t, d, envelopeJSON("TLS-01", "completed", "", payload("handshake_succeeded", tcp("failed", "connection_refused"), tls("succeeded", "", "trusted"))))
	reject(t, d, envelopeJSON("TLS-01", "completed", "", payload("handshake_succeeded", tcp("succeeded", ""))))
	// One primitive cannot retain interrupted protocol observations as terminal.
	reject(t, decoder(t, binding("TCP-01")), envelopeJSON("TCP-01", "completed", "", payload("connected", tcp("incomplete", ""))))
	// NET-01 access-only result never serializes shared egress context.
	d = decoder(t, binding("NET-01"))
	reject(t, d, envelopeJSON("NET-01", "completed", "", payload("snapshot")))
	input := envelopeJSON("NET-01", "completed", "", `{"outcome":"snapshot","observations":[],"access":`+access()+`}`)
	reject(t, d, mutate(t, input, func(v map[string]any) { v["network_context"] = map[string]any{} }))
	reject(t, d, bytes.Replace(input, []byte(`"network_changed":false`), []byte(`"network_changed":true`), 1))
}

// requirement-ref: REQ-DNS-01-0038
func TestDNSCollectionAndHTTPEvidence(t *testing.T) {
	b := binding("DNS-01")
	b.Output.DNSMode = "udp"
	d := decoder(t, b)
	// Different retained codes establish a conflict, with count-limit precedence.
	obs := `{"type":"dns_query","status":"succeeded","duration_ms":500,"details":{"engine":"udp","query_type":"A","responses":[{"arrival_index":1,"arrival_offset_ms":250,"rcode":"NOERROR","answers":[]},{"arrival_index":2,"arrival_offset_ms":500,"rcode":"NXDOMAIN","answers":[]}]}}`
	input := envelopeJSON("DNS-01", "completed", "", payload("conflicting_responses", obs))
	if _, e := d.Decode(input); e != nil {
		t.Fatal(e)
	}
	reject(t, d, bytes.Replace(input, []byte(`"outcome":"conflicting_responses"`), []byte(`"outcome":"resolved"`), 1))
	b.Output.Limits.MaxDNSResponses = 2
	d = decoder(t, b)
	obs = string(bytes.Replace([]byte(obs), []byte(`"engine":"udp"`), []byte(`"engine":"udp","response_limit_reached":true`), 1))
	if _, e := d.Decode(envelopeJSON("DNS-01", "completed", "", payload("response_limit_reached", obs))); e != nil {
		t.Fatal(e)
	}
	reject(t, d, envelopeJSON("DNS-01", "completed", "", payload("conflicting_responses", obs)))
	b = binding("HTTP-01")
	b.Output.Limits.BodyChecks = 1
	d = decoder(t, b)
	input = envelopeJSON("HTTP-01", "completed", "", payload("expected_response", tcp("succeeded", ""), tls("succeeded", "", "trusted"), http()))
	reject(t, d, input) // Configured body check is missing, not silently "success".
	b = binding("NET-03")
	b.Output.HTTPS = false
	d = decoder(t, b)
	input = envelopeJSON("NET-03", "completed", "", `{"outcome":"validated","observations":[`+tcp("succeeded", "")+`,`+http()+`],"platform_validation":"unknown","response_class":"expected"}`)
	if _, e := d.Decode(input); e != nil {
		t.Fatal(e)
	}
	reject(t, d, mutate(t, input, func(v map[string]any) { delete(object(v, "measurement"), "response_class") }))
	reject(t, d, mutate(t, input, func(v map[string]any) { object(v, "measurement")["response_class"] = "redirect" }))
	reject(t, d, envelopeJSON("NET-03", "completed", "", `{"outcome":"connection_failure","observations":[`+tcp("failed", "connection_refused")+`],"platform_validation":"unknown","response_class":"unexpected"}`))
}

// requirement-ref: REQ-NET-05-0040
// requirement-ref: REQ-NET-05-0044
// requirement-ref: REQ-NET-05-0045
// requirement-ref: REQ-NET-05-0046
func TestCompositeCompletionAndPartialBoundaries(t *testing.T) {
	b := binding("NET-05")
	d := decoder(t, b)
	positive := payload("connected", tcp("succeeded", ""))
	negative := payload("timeout", tcp("failed", "timeout"))
	first := step("first", "completed", negative)
	second := step("second", "completed", positive)
	complete := envelopeJSON("NET-05", "completed", "", net05("completed", "validated", []string{first, second}, ""))
	v, e := d.Decode(complete)
	if e != nil {
		t.Fatal(e)
	}
	view := v.View()
	view.Measurement.Profile.ID = "changed"
	view.Measurement.NetworkValidation.State = "changed"
	view.Measurement.Steps[0].Measurement.Outcome = "changed"
	if v.View().Measurement.Steps[0].Measurement.Outcome != "timeout" {
		t.Fatal("step alias")
	}
	if _, e := d.Build(v.View()); e != nil {
		t.Fatal(e)
	}
	for _, steps := range [][]string{nil, {first}, {step("first", "incomplete", "")}, {first, step("second", "incomplete", "")}} {
		stop := ""
		if len(steps) < 2 {
			if len(steps) == 0 {
				stop = "first"
			} else {
				stop = "second"
			}
		}
		input := envelopeJSON("NET-05", "incomplete", "attempt_deadline_exceeded", net05("incomplete", "validated", steps, stop))
		if _, e := d.Decode(input); e != nil {
			t.Fatal(string(input), e)
		}
		reject(t, d, mutate(t, input, func(v map[string]any) { object(v, "measurement")["outcome"] = "completed" }))
	}
	invalid := []struct{ state, m string }{
		{"completed", net05("completed", "validated", []string{second, first}, "")},
		{"completed", net05("completed", "validated", []string{first}, "second")},
		{"incomplete", net05("incomplete", "validated", []string{first, second}, "")},
		{"incomplete", net05("incomplete", "validated", []string{step("first", "incomplete", ""), second}, "")},
		{"incomplete", net05("incomplete", "validated", []string{first}, "")},
		{"incomplete", net05("incomplete", "validated", []string{first}, "first")},
		{"incomplete", net05("incomplete", "validated", []string{first, step("second", "incomplete", positive)}, "")},
		{"completed", net05("network_not_validated", "not_validated", nil, "first")},
	}
	for _, tc := range invalid {
		reject(t, d, envelopeJSON("NET-05", tc.state, func() string {
			if tc.state == "incomplete" {
				return "resource_limit"
			}
			return ""
		}(), tc.m))
	}
	reject(t, d, mutate(t, complete, func(v map[string]any) { object(object(v, "measurement"), "profile")["profile_version"] = 2 }))
	reject(t, d, mutate(t, complete, func(v map[string]any) {
		steps := object(v, "measurement")["steps"].([]any)
		object(steps[0].(map[string]any), "test")["result_version"] = "9.0.0"
	}))
	b.Profile.RequireValidation = true
	d = decoder(t, b)
	denied := envelopeJSON("NET-05", "completed", "", net05("network_not_validated", "not_validated", nil, "first"))
	if _, e := d.Decode(denied); e != nil {
		t.Fatal(e)
	}
	reject(t, d, bytes.Replace(complete, []byte(`"state":"validated"`), []byte(`"state":"not_validated"`), 1))
	// New copied the trusted profile. Later mutation cannot change accepted IDs.
	b.Profile.Steps[0].ID = "replaced"
	if _, e := d.Decode(denied); e != nil {
		t.Fatal("binding alias", e)
	}
}

func TestBindingBoundsAndProducerContext(t *testing.T) {
	for _, fn := range []func(*Binding){
		func(b *Binding) { b.TaskID = "" }, func(b *Binding) { b.MaxDurationMS = 60001 }, func(b *Binding) { b.MaxAttempts = 3 },
		func(b *Binding) { b.Output.Test.Version = "9.0.0" }, func(b *Binding) { b.Output.Test.ResultVersion = "9.0.0" },
		func(b *Binding) { b.Output.Limits = observation.ProtocolLimits{BodyChecks: 9} },
		func(b *Binding) { b.Implementation.Platform = "web" }, func(b *Binding) { b.Implementation.LibraryVersion = "" },
	} {
		b := binding("TCP-01")
		fn(&b)
		if _, e := New(b); e == nil {
			t.Fatal("invalid binding accepted")
		}
	}
	b := binding("NET-05")
	b.Profile.Steps[1].ID = b.Profile.Steps[0].ID
	if _, e := New(b); e == nil {
		t.Fatal("duplicate step")
	}
	b = binding("NET-05")
	b.Profile.Steps[1].Output.Test = testRef("NET-05")
	if _, e := New(b); e == nil {
		t.Fatal("nested composite")
	}
	var d *Decoder
	if _, e := d.Decode([]byte(`{}`)); e == nil {
		t.Fatal("nil decoder")
	}
}

func FuzzResult(f *testing.F) {
	ids := []string{"TCP-01", "TLS-01", "DNS-01", "DNS-03", "DNS-04", "HTTP-01", "NET-01", "NET-03", "NET-05"}
	decoders := make([]*Decoder, len(ids))
	for i, id := range ids {
		decoders[i] = decoder(f, binding(id))
	}
	f.Add(uint8(0), envelopeJSON("TCP-01", "completed", "", payload("connection_refused", tcp("failed", "connection_refused"))))
	f.Add(uint8(6), envelopeJSON("NET-01", "completed", "", `{"outcome":"snapshot","observations":[],"access":`+access()+`}`))
	f.Add(uint8(8), envelopeJSON("NET-05", "incomplete", "resource_limit", net05("incomplete", "validated", nil, "first")))
	for i, id := range ids {
		f.Add(uint8(i), envelopeJSON(id, "not_run", "local_policy_rejected", ""))
	}
	f.Fuzz(func(t *testing.T, index uint8, input []byte) {
		d := decoders[int(index)%len(decoders)]
		v, e := d.Decode(input)
		if e != nil {
			if _, e := v.MarshalJSON(); e == nil {
				t.Fatal("partial result")
			}
			return
		}
		wire, e := v.MarshalJSON()
		if e != nil || len(wire) > 65536 {
			t.Fatal("invalid successful serialization", e)
		}
		next, e := d.Decode(wire)
		if e != nil {
			t.Fatal("successful result cannot redecode", e)
		}
		again, _ := next.MarshalJSON()
		if !bytes.Equal(wire, again) {
			t.Fatal("normalization not stable")
		}
		produced, e := d.Build(v.View())
		if e != nil {
			t.Fatal("producer roundtrip", e)
		}
		again, _ = produced.MarshalJSON()
		if !bytes.Equal(wire, again) {
			t.Fatal("producer changed result")
		}
	})
}

func TestDoHNonSuccessResponseAndRedirectBounds(t *testing.T) {
	b := binding("DNS-03")
	b.Output.Limits.StatusCheck = false
	d := decoder(t, b)
	response := `{"type":"http_exchange","status":"succeeded","duration_ms":250,"details":{"address_family":"ipv4","protocol":"h2","status_code":503}}`
	input := envelopeJSON("DNS-03", "completed", "", payload("http_failure", tcp("succeeded", ""), tls("succeeded", "", "trusted"), response))
	if _, e := d.Decode(input); e != nil {
		t.Fatal("non-success DoH response", e)
	}
	reject(t, d, bytes.Replace(input, []byte(`"status_code":503`), []byte(`"status_code":200`), 1))
	b = binding("HTTP-01")
	d = decoder(t, b)
	repeated := envelopeJSON("HTTP-01", "completed", "", payload("expected_response", tcp("succeeded", ""), tls("succeeded", "", "trusted"), http(), http()))
	reject(t, d, repeated)
}

func TestCompositeVersionPrecisionAndMarkerPresence(t *testing.T) {
	b := binding("NET-05")
	b.Profile.Version = 9007199254740993
	d := decoder(t, b)
	input := envelopeJSON("NET-05", "incomplete", "resource_limit", net05("incomplete", "validated", nil, "first"))
	input = bytes.Replace(input, []byte(`"profile_version":1`), []byte(`"profile_version":9007199254740993`), 1)
	v, e := d.Decode(input)
	if e != nil {
		t.Fatal(e)
	}
	wire, _ := v.MarshalJSON()
	if !bytes.Contains(wire, []byte(`"profile_version":9007199254740993`)) {
		t.Fatal("profile revision rounded")
	}
	if _, e := d.Build(v.View()); e != nil {
		t.Fatal(e)
	}
	b = binding("NET-05")
	d = decoder(t, b)
	complete := envelopeJSON("NET-05", "completed", "", net05("completed", "validated", []string{step("first", "completed", payload("connected", tcp("succeeded", ""))), step("second", "completed", payload("connected", tcp("succeeded", "")))}, ""))
	reject(t, d, mutate(t, complete, func(v map[string]any) { object(v, "measurement")["stopped_before_step"] = "" }))
}
