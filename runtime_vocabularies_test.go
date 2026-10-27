package barometer

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

// These are caller/result cases: only the transport and trusted platform inputs
// are synthetic. Production admission, protocol operations, runners and result
// construction execute. They do not establish native or field behavior.
func vocabularyDecoder(t *testing.T, id, mode string) *result.Decoder {
	t.Helper()
	cd, err := observation.NewContextDecoder(result.Contract, observation.ContextApprovals{})
	if err != nil {
		t.Fatal(err)
	}
	path, err := observation.NewPath(1, 0, "direct", "")
	if err != nil {
		t.Fatal(err)
	}
	output := result.Output{Test: model.Test{ID: id, Version: "0.1.0", ResultVersion: "0.1.0"}}
	if id == "DNS-01" {
		output.DNSMode, output.QueryType = mode, "A"
		output.Limits.RetainDNSAnswers = true
	}
	d, err := result.New(result.Binding{TaskID: "vocabulary", Output: output, MaxAttempts: 1, MaxDurationMS: 60000, Implementation: implementation, ContextDecoder: cd, Path: path, AccessGeneration: 1})
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func vocabularyRoundTrip(t *testing.T, d *result.Decoder, raw []byte) {
	t.Helper()
	v, err := d.Decode(raw)
	if err != nil {
		t.Fatal("producer result rejected:", err)
	}
	encoded, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var before, after any
	if json.Unmarshal(raw, &before) != nil || json.Unmarshal(encoded, &after) != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("receiver changed fields or optional-field presence")
	}
}

func vocabularyMutation(t *testing.T, d *result.Decoder, raw []byte, mutate func(map[string]any)) {
	t.Helper()
	var v map[string]any
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	mutate(v)
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Decode(b); err == nil {
		t.Fatalf("receiver repaired invalid serialized vocabulary: %s", b)
	}
}

func vocabularyObservation(v map[string]any, index int) map[string]any {
	return v["measurement"].(map[string]any)["observations"].([]any)[index].(map[string]any)
}

// requirement-ref: REQ-OBS-0042
// requirement-ref: REQ-OBS-0045
// requirement-ref: REQ-OBS-0046
// requirement-ref: REQ-OBS-0047
func TestRuntimeDNSVocabularies(t *testing.T) {
	codes := []struct {
		code                    uint16
		label, outcome, failure string
	}{
		{0, "NOERROR", "no_data", ""}, {1, "FORMERR", "inconclusive", "inconclusive"},
		{2, "SERVFAIL", "servfail", "servfail"}, {3, "NXDOMAIN", "nxdomain", "nxdomain"},
		{4, "NOTIMP", "inconclusive", "inconclusive"}, {5, "REFUSED", "refused", "refused"},
		{6, "OTHER", "inconclusive", "inconclusive"}, {16, "OTHER", "inconclusive", "inconclusive"},
		{18, "OTHER", "inconclusive", "inconclusive"}, {23, "OTHER", "inconclusive", "inconclusive"},
		{3841, "OTHER", "inconclusive", "inconclusive"}, {4095, "OTHER", "inconclusive", "inconclusive"},
	}
	for _, mode := range []string{"udp", "tcp"} {
		for _, tc := range codes {
			t.Run(fmt.Sprintf("%s/%d", mode, tc.code), func(t *testing.T) {
				h, transport := dnsHarness(t, mode)
				server := dnsLoopback(t, transport, func(q []byte) [][]byte { return [][]byte{loopback.DNSResponse(q, tc.code, nil)} })
				c := h.initialized(t)
				defer c.Close()
				raw := tcpRaw(t, h, c, dnsTask("vocabulary", mode))
				r := dnsResult(t, raw, "completed", "", tc.outcome)
				o := r.Measurement.Observations[0]
				if (o.Status == "succeeded") != (tc.failure == "") || o.Failure != tc.failure || len(o.Details.Responses) != 1 || o.Details.Responses[0].RCode != tc.label || len(o.Details.Responses[0].Answers) != 0 {
					t.Fatalf("incorrect normalized response: %s", raw)
				}
				if queries, _, _ := server.Counts(); queries != 1 || transport.calls.Load() != 1 {
					t.Fatal("normalization caused extra contact")
				}
				d := vocabularyDecoder(t, "DNS-01", mode)
				vocabularyRoundTrip(t, d, raw)
				for _, bad := range []any{"BADVERS", "noerror", "16", 16, "PRIVATE-RCODE", nil} {
					vocabularyMutation(t, d, raw, func(v map[string]any) {
						details := vocabularyObservation(v, 0)["details"].(map[string]any)
						details["responses"].([]any)[0].(map[string]any)["rcode"] = bad
					})
				}
			})
		}
	}
}

// requirement-ref: REQ-OBS-0043
// requirement-ref: REQ-OBS-0048
// requirement-ref: REQ-OBS-0049
func TestRuntimeTLSNegotiatedVocabularies(t *testing.T) {
	f := newTLSFixture(t)
	for _, version := range []struct {
		wire  uint16
		label string
	}{{tls.VersionTLS12, "TLS1.2"}, {tls.VersionTLS13, "TLS1.3"}} {
		for _, alpn := range []string{"http/1.1", "h2", "dot", ""} {
			t.Run(version.label+"/"+alpn, func(t *testing.T) {
				h, transport := tlsProductionHarness(t, f)
				var offer []string
				options := `{}`
				if alpn != "" {
					offer = []string{alpn}
					options = `{"alpn_protocols":["` + alpn + `"]}`
				}
				tlsPolicy(t, h, tcpPublic, 443, tlsServerName, offer)
				server := tlsServe(t, transport, loopback.TLSConfig{Certificate: f.cert, ALPN: offer, MaxVersion: version.wire})
				c := h.initialized(t)
				defer c.Close()
				raw := tcpRaw(t, h, c, strings.Replace(tlsTask("vocabulary"), tlsOptions, options, 1))
				tlsResult(t, raw, tlsExpect{state: "completed", outcome: "handshake_succeeded", tls: true, validation: "trusted", version: version.label, alpn: alpn})
				if server.Counts().Plaintext != 0 || transport.calls.Load() != 1 {
					t.Fatal("extra contact or application data")
				}
				d := vocabularyDecoder(t, "TLS-01", "")
				vocabularyRoundTrip(t, d, raw)
				var v map[string]any
				if err := json.Unmarshal(raw, &v); err != nil {
					t.Fatal(err)
				}
				details := vocabularyObservation(v, 1)["details"].(map[string]any)
				for _, key := range []string{"alert", "alpn"} {
					if key == "alpn" && alpn != "" {
						continue
					}
					if _, present := details[key]; present {
						t.Fatalf("absent %s was retained", key)
					}
				}
				for _, field := range []string{"tls_version", "alpn"} {
					for _, bad := range []any{"PRIVATE-TLS-CANARY", "TLS1.0", "H2", 17, nil} {
						vocabularyMutation(t, d, raw, func(v map[string]any) { vocabularyObservation(v, 1)["details"].(map[string]any)[field] = bad })
					}
				}
			})
		}
	}
}

// Every explicitly grouped description in the public table is received as a
// real TLS record. Description 120 is tested without an ALPN offer here; the
// existing normalization suite separately verifies alpn_no_overlap precedence.
// requirement-ref: REQ-OBS-0043
// requirement-ref: REQ-OBS-0050
func TestRuntimeTLSPeerAlertVocabularies(t *testing.T) {
	f := newTLSFixture(t)
	groups := []struct {
		codes []uint8
		class string
	}{
		{[]uint8{40, 71, 86, 120}, "handshake_failure"},
		{[]uint8{42, 43, 44, 45, 46, 48, 113, 116}, "certificate_error"},
		{[]uint8{10, 20, 22, 47, 50, 51, 70, 109, 110}, "protocol_error"},
		{[]uint8{49, 112, 115}, "access_denied"},
		{[]uint8{80}, "internal_error"},
		{[]uint8{1, 200, 255}, "other"},
	}
	for _, group := range groups {
		for _, code := range group.codes {
			t.Run(fmt.Sprint(code), func(t *testing.T) {
				h, transport := tlsProductionHarness(t, f)
				tlsPolicy(t, h, tcpPublic, 443, tlsServerName, nil)
				tlsServe(t, transport, loopback.TLSConfig{Mode: loopback.TLSAlert, Alert: code})
				c := h.initialized(t)
				defer c.Close()
				raw := tcpRaw(t, h, c, strings.Replace(tlsTask("vocabulary"), tlsOptions, `{}`, 1))
				tlsResult(t, raw, tlsExpect{state: "completed", outcome: "tls_alert", tls: true, tlsFailure: "alert", validation: "not_completed", alert: group.class})
				if transport.calls.Load() != 1 {
					t.Fatal("alert caused a retry")
				}
				d := vocabularyDecoder(t, "TLS-01", "")
				vocabularyRoundTrip(t, d, raw)
				var v map[string]any
				if err := json.Unmarshal(raw, &v); err != nil {
					t.Fatal(err)
				}
				details := vocabularyObservation(v, 1)["details"].(map[string]any)
				for _, key := range []string{"tls_version", "alpn"} {
					if _, ok := details[key]; ok {
						t.Fatalf("unnegotiated %s retained", key)
					}
				}
				for _, bad := range []any{"bad_certificate", "PRIVATE-ALERT-CANARY", code, fmt.Sprint(code), nil} {
					vocabularyMutation(t, d, raw, func(v map[string]any) { vocabularyObservation(v, 1)["details"].(map[string]any)["alert"] = bad })
				}
			})
		}
	}
}

// A peer's unoffered selection is rejected by crypto/tls before any successful
// negotiation; it must not become a successful observation with alpn: unknown.
// requirement-ref: REQ-OBS-0049
func TestRuntimeTLSRejectsUnofferedALPN(t *testing.T) {
	// Verify the hostile fixture actually reaches the ALPN validation branch.
	server, err := loopback.ListenTLS(loopback.TLSConfig{Mode: loopback.TLSUnofferedALPN})
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	rawConn, err := net.DialTimeout("tcp", server.Address(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	conn := tls.Client(rawConn, &tls.Config{ServerName: tlsServerName, NextProtos: []string{"h2"}, MinVersion: tls.VersionTLS12})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = conn.HandshakeContext(ctx)
	conn.Close()
	if err == nil || !strings.Contains(err.Error(), "unadvertised ALPN protocol") {
		t.Fatalf("fixture failed outside ALPN selection: %v", err)
	}
	f := newTLSFixture(t)
	h, transport := tlsProductionHarness(t, f)
	tlsServe(t, transport, loopback.TLSConfig{Mode: loopback.TLSUnofferedALPN})
	c := h.initialized(t)
	defer c.Close()
	raw := tcpRaw(t, h, c, tlsTask("vocabulary"))
	tlsResult(t, raw, tlsExpect{state: "completed", outcome: "inconclusive", tls: true, tlsFailure: "inconclusive", validation: "inconclusive"})
	if bytes.Contains(raw, []byte("PRIVATE-ALPN-CANARY")) || bytes.Contains(raw, []byte(`"alpn"`)) || bytes.Contains(raw, []byte(`"tls_version"`)) || bytes.Contains(raw, []byte(`"alert"`)) {
		t.Fatal("invalid negotiation retained")
	}
	vocabularyRoundTrip(t, vocabularyDecoder(t, "TLS-01", ""), raw)
}

// Existing HTTP and composite consumers must preserve the shared primitive
// labels while selecting their own terminal outcome.
// requirement-ref: REQ-OBS-0046
// requirement-ref: REQ-OBS-0047
// requirement-ref: REQ-OBS-0050
func TestRuntimeVocabularyReuse(t *testing.T) {
	t.Run("HTTP-01", func(t *testing.T) {
		h, transport, _ := httpHarness(t, httpInput, responseHTTP(""), false)
		server, err := loopback.ListenTLS(loopback.TLSConfig{Mode: loopback.TLSAlert, Alert: 200})
		if err != nil {
			t.Fatal(err)
		}
		defer server.Close()
		transport.dial = func(ctx context.Context, network, address string) (measurement.Conn, error) {
			if network != "tcp" || address != "8.8.8.8:443" {
				return nil, fmt.Errorf("unexpected synthetic destination")
			}
			var d net.Dialer
			return d.DialContext(ctx, "tcp", server.Address())
		}
		c := h.initialized(t)
		defer c.Close()
		raw := tcpRaw(t, h, c, httpTask("vocabulary", httpInput, `{}`))
		checkHTTP(t, raw, "completed", "", "tls_failure")
		var v map[string]any
		if err := json.Unmarshal(raw, &v); err != nil {
			t.Fatal(err)
		}
		o := vocabularyObservation(v, 1)
		if o["type"] != "tls_handshake" || o["failure"] != "alert" || o["details"].(map[string]any)["alert"] != "other" || transport.calls.Load() != 1 {
			t.Fatalf("HTTP changed shared TLS evidence: %s", raw)
		}
		vocabularyRoundTrip(t, vocabularyDecoder(t, "HTTP-01", ""), raw)
	})
	t.Run("NET-05", func(t *testing.T) {
		h, transport := net05AdmissionHarness(t, nil)
		dnsServer, err := loopback.ListenDNS(func(q []byte) [][]byte { return [][]byte{loopback.DNSResponse(q, 16, nil)} })
		if err != nil {
			t.Fatal(err)
		}
		defer dnsServer.Close()
		tlsServer, err := loopback.ListenTLS(loopback.TLSConfig{Mode: loopback.TLSAlert, Alert: 200})
		if err != nil {
			t.Fatal(err)
		}
		defer tlsServer.Close()
		transport.dial = func(ctx context.Context, network, address string) (measurement.Conn, error) {
			destination := dnsServer.Address()
			if address == "8.8.8.8:443" {
				destination = tlsServer.Address()
			} else if address != "8.8.8.8:53" {
				return nil, fmt.Errorf("unexpected synthetic destination")
			}
			if network != "tcp" {
				return nil, fmt.Errorf("unexpected synthetic transport")
			}
			var d net.Dialer
			return d.DialContext(ctx, network, destination)
		}
		readyComposite(h)
		c := h.initialized(t)
		defer c.Close()
		steps := primitiveSteps()
		raw := tcpRaw(t, h, c, net05Task("vocabulary", profileInput(steps[0], steps[2]), `{}`, ""))
		r := net05Wire(t, raw)
		if r.Execution.State != "completed" || r.Measurement.Outcome != "completed" || len(r.Measurement.Steps) != 2 || r.Measurement.Steps[0].Measurement.Outcome != "inconclusive" || r.Measurement.Steps[1].Measurement.Outcome != "tls_alert" || transport.calls.Load() != 2 {
			t.Fatalf("composite outcomes: %s", raw)
		}
		var v map[string]any
		if err := json.Unmarshal(raw, &v); err != nil {
			t.Fatal(err)
		}
		nested := v["measurement"].(map[string]any)["steps"].([]any)
		dnsObservation := vocabularyObservation(nested[0].(map[string]any), 0)
		response := dnsObservation["details"].(map[string]any)["responses"].([]any)[0].(map[string]any)
		tlsObservation := vocabularyObservation(nested[1].(map[string]any), 1)
		if response["rcode"] != "OTHER" || tlsObservation["details"].(map[string]any)["alert"] != "other" {
			t.Fatalf("composite changed shared evidence: %s", raw)
		}
		cd, err := observation.NewContextDecoder(result.Contract, observation.ContextApprovals{})
		if err != nil {
			t.Fatal(err)
		}
		path, err := observation.NewPath(1, 0, "direct", "")
		if err != nil {
			t.Fatal(err)
		}
		d, err := result.New(result.Binding{TaskID: "vocabulary", Output: result.Output{Test: model.Test{ID: "NET-05", Version: "0.1.0", ResultVersion: "0.1.0"}}, MaxAttempts: 1, MaxDurationMS: 60000, Implementation: implementation, ContextDecoder: cd, Path: path, AccessGeneration: 1, Profile: &result.ProfilePlan{ID: "baseline", Version: 1, Steps: []result.StepPlan{
			{ID: "dns", Output: result.Output{Test: model.Test{ID: "DNS-01", Version: "0.1.0", ResultVersion: "0.1.0"}, DNSMode: "tcp", QueryType: "A", Limits: observation.ProtocolLimits{RetainDNSAnswers: true}}},
			{ID: "tls", Output: result.Output{Test: model.Test{ID: "TLS-01", Version: "0.1.0", ResultVersion: "0.1.0"}}},
		}}})
		if err != nil {
			t.Fatal(err)
		}
		vocabularyRoundTrip(t, d, raw)
	})
}
