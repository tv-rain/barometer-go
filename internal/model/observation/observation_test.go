package observation

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

const tcpOK = `{"type":"tcp_connect","status":"succeeded","duration_ms":250,"details":{"address_family":"ipv4"}}`
const httpOK = `{"type":"http_exchange","status":"succeeded","duration_ms":500,"details":{"address_family":"unknown","protocol":"h2","status_code":200,"body":{"bytes_processed":512,"truncated":false},"checks":[{"kind":"status","index":0,"matched":true},{"kind":"body","index":0,"matched":true}]}}`

var protocolLimits = ProtocolLimits{MaxBodyBytes: 512, MaxRedirects: 2, StatusCheck: true, HeaderChecks: 8, BodyChecks: 8, RetainDNSAnswers: true}

// requirement-ref: REQ-OBS-0003
// requirement-ref: REQ-OBS-0036
// requirement-ref: REQ-OBS-0037
func TestProtocolPresenceAndVocabulary(t *testing.T) {
	for _, s := range []string{tcpOK, httpOK, strings.Replace(tcpOK, `"succeeded"`, `"incomplete"`, 1), strings.Replace(tcpOK, `"succeeded"`, `"failed","failure":"inconclusive"`, 1)} {
		o, err := DecodeProtocol(Contract, []byte(s), protocolLimits)
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(o)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = DecodeProtocol(Contract, encoded, protocolLimits); err != nil {
			t.Fatal(err)
		}
	}
	cases := map[string]string{
		"missing duration":      strings.Replace(tcpOK, `"duration_ms":250,`, "", 1),
		"negative duration":     strings.Replace(tcpOK, `250`, `-250`, 1),
		"quantization":          strings.Replace(tcpOK, `250`, `251`, 1),
		"operation bound":       strings.Replace(tcpOK, `250`, `60250`, 1),
		"float":                 strings.Replace(tcpOK, `250`, `250.0`, 1),
		"exponent":              strings.Replace(tcpOK, `250`, `25e1`, 1),
		"null":                  strings.Replace(tcpOK, `250`, `null`, 1),
		"duplicate":             strings.Replace(tcpOK, `"duration_ms":250`, `"duration_ms":250,"duration_ms":250`, 1),
		"escaped duplicate":     strings.Replace(tcpOK, `"duration_ms":250`, `"duration_ms":250,"duration_\u006ds":250`, 1),
		"failure absent":        strings.Replace(tcpOK, `"succeeded"`, `"failed"`, 1),
		"failure on success":    strings.Replace(tcpOK, `"succeeded"`, `"succeeded","failure":"timeout"`, 1),
		"failure on incomplete": strings.Replace(tcpOK, `"succeeded"`, `"incomplete","failure":"timeout"`, 1),
		"execution reason":      strings.Replace(tcpOK, `"succeeded"`, `"failed","failure":"network_changed"`, 1),
		"unknown family":        strings.Replace(tcpOK, `"ipv4"`, `"unknown"`, 1),
		"unknown status":        strings.Replace(tcpOK, `"succeeded"`, `"unknown"`, 1),
		"wrong DNS details":     strings.Replace(tcpOK, `"tcp_connect"`, `"dns_query"`, 1),
		"wrong TLS details":     strings.Replace(tcpOK, `"tcp_connect"`, `"tls_handshake"`, 1),
		"trailing":              tcpOK + ` {}`,
		"invalid unicode":       strings.Replace(tcpOK, `"ipv4"`, `"\ud800"`, 1),
		"size":                  tcpOK + strings.Repeat(" ", 65536),
	}
	for name, s := range cases {
		t.Run(name, func(t *testing.T) { assertProtocolRejected(t, s, protocolLimits) })
	}
	if _, err := DecodeProtocol("0.2.0", []byte(tcpOK), protocolLimits); !errors.Is(err, ErrVersion) {
		t.Fatal(err)
	}
	var zero Observation
	if _, err := json.Marshal(zero); err == nil {
		t.Fatal("serialized zero value")
	}
	if err := json.Unmarshal([]byte(tcpOK), &zero); !errors.Is(err, ErrImport) {
		t.Fatal(err)
	}
}

// requirement-ref: REQ-OBS-0025
// requirement-ref: REQ-OBS-0028
// requirement-ref: REQ-OBS-0030
// requirement-ref: REQ-OBS-0038
func TestHTTPBoundsAndConsistency(t *testing.T) {
	cases := map[string]string{
		"missing false":               strings.Replace(httpOK, `,"truncated":false`, "", 1),
		"body over configured bound":  strings.Replace(httpOK, `512`, `513`, 1),
		"body truncated before bound": strings.Replace(strings.Replace(httpOK, `512`, `511`, 1), `false`, `true`, 1),
		"status too low":              strings.Replace(httpOK, `200`, `99`, 1),
		"status too high":             strings.Replace(httpOK, `200`, `600`, 1),
		"unknown protocol":            strings.Replace(httpOK, `"h2"`, `"h3"`, 1),
		"status index":                strings.Replace(httpOK, `"index":0`, `"index":1`, 1),
		"negative index":              strings.Replace(httpOK, `"index":0`, `"index":-1`, 1),
		"body index out of range":     strings.Replace(httpOK, `"kind":"body","index":0`, `"kind":"body","index":8`, 1),
		"check mismatch on success":   strings.Replace(httpOK, `"matched":true`, `"matched":false`, 1),
		"unregistered failure":        strings.Replace(httpOK, `"succeeded"`, `"failed","failure":"raw_error"`, 1),
		"failure lacking evidence":    strings.Replace(httpOK, `"succeeded"`, `"failed","failure":"body_check_failed"`, 1),
		"duplicate check":             strings.Replace(httpOK, `"kind":"body"`, `"kind":"status"`, 1),
		"empty checks":                strings.Replace(httpOK, `[{"kind":"status","index":0,"matched":true},{"kind":"body","index":0,"matched":true}]`, `[]`, 1),
		"null body":                   strings.Replace(httpOK, `{"bytes_processed":512,"truncated":false}`, `null`, 1),
	}
	for name, s := range cases {
		t.Run(name, func(t *testing.T) { assertProtocolRejected(t, s, protocolLimits) })
	}
	fail := strings.Replace(strings.Replace(httpOK, `"succeeded"`, `"failed","failure":"body_check_failed"`, 1), `"kind":"body","index":0,"matched":true`, `"kind":"body","index":0,"matched":false`, 1)
	if _, err := DecodeProtocol(Contract, []byte(fail), protocolLimits); err != nil {
		t.Fatal(err)
	}
	l := protocolLimits
	l.BodyChecks = 0
	assertProtocolRejected(t, httpOK, l)
	l = protocolLimits
	l.MaxBodyBytes = 65537
	assertProtocolRejected(t, httpOK, l)
	redirect := `{"type":"http_exchange","status":"succeeded","duration_ms":0,"details":{"address_family":"unknown","protocol":"unknown","status_code":302,"redirect":{"followed_count":2,"terminal_location_class":"registered_https"}}}`
	if _, err := DecodeProtocol(Contract, []byte(redirect), protocolLimits); err != nil {
		t.Fatal(err)
	}
	l = protocolLimits
	l.MaxRedirects = 1
	assertProtocolRejected(t, redirect, l)
	assertProtocolRejected(t, strings.Replace(redirect, `"registered_https"`, `"https://private.invalid"`, 1), protocolLimits)
}

// requirement-ref: REQ-OBS-0040
// requirement-ref: REQ-OBS-0031
func TestProtocolPrivacyAndImmutableViews(t *testing.T) {
	for _, view := range []any{TCPDetails{AddressFamily: "private"}, HTTPDetails{Protocol: "private"}, Redirect{TerminalLocationClass: "private"}, Body{}, Check{Kind: "private"}} {
		if _, err := json.Marshal(view); err == nil {
			t.Fatal("serialized unvalidated detached view")
		}
	}
	for _, field := range []string{"raw_packet", "body_bytes", "body_digest", "headers", "certificate", "resolver_address", "proxy_host", "device_id", "destination", "location"} {
		for _, s := range []string{strings.Replace(tcpOK, `{`, `{"`+field+`":"private",`, 1), strings.Replace(httpOK, `"details":{`, `"details":{"`+field+`":"private",`, 1)} {
			assertProtocolRejected(t, s, protocolLimits)
		}
	}
	input := []byte(httpOK)
	o, err := DecodeProtocol(Contract, input, protocolLimits)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(o)
	for i := range input {
		input[i] = 0
	}
	d, _ := o.HTTP()
	*d.StatusCode = 500
	d.Body.BytesProcessed = 0
	d.Checks[0].Matched = false
	after, _ := json.Marshal(o)
	if string(before) != string(after) {
		t.Fatal("mutable alias escaped")
	}
	if !strings.Contains(string(after), `"truncated":false`) {
		t.Fatal("required false lost")
	}
}

func assertProtocolRejected(t *testing.T, s string, l ProtocolLimits) {
	t.Helper()
	o, err := DecodeProtocol(Contract, []byte(s), l)
	if err == nil {
		t.Fatal("accepted invalid observation")
	}
	if strings.Contains(err.Error(), "private") {
		t.Fatal("input in error")
	}
	if _, err = json.Marshal(o); err == nil {
		t.Fatal("partial model escaped")
	}
}

func FuzzProtocol(f *testing.F) {
	f.Add([]byte(tcpOK))
	f.Add([]byte(httpOK))
	f.Add([]byte(dnsOK))
	f.Add([]byte(tlsOK))
	f.Fuzz(func(t *testing.T, input []byte) {
		o, err := DecodeProtocol(Contract, input, protocolLimits)
		if err != nil {
			return
		}
		b, err := json.Marshal(o)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = DecodeProtocol(Contract, b, protocolLimits); err != nil {
			t.Fatal(err)
		}
	})
}
