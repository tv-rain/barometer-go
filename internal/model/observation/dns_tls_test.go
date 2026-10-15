package observation

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

const dnsOK = `{"type":"dns_query","status":"succeeded","duration_ms":500,"details":{"engine":"udp","address_family":"ipv4","query_type":"A","responses":[{"arrival_index":1,"arrival_offset_ms":250,"rcode":"NOERROR","answers":[{"type":"CNAME","value":"alias.example.org"},{"type":"A","value":"1.1.1.1"}]}]}}`
const tlsOK = `{"type":"tls_handshake","status":"succeeded","duration_ms":250,"details":{"address_family":"ipv4","tls_version":"TLS1.3","alpn":"h2","certificate_validation":"trusted"}}`

// requirement-ref: REQ-OBS-0009
// requirement-ref: REQ-OBS-0012
// requirement-ref: REQ-OBS-0013
// requirement-ref: REQ-OBS-0014
// requirement-ref: REQ-OBS-0042
func TestDNSBoundsAndPrivacy(t *testing.T) {
	for _, s := range []string{dnsOK, `{"type":"dns_query","status":"failed","duration_ms":0,"failure":"resolver_error","details":{"engine":"system","query_type":"AAAA","responses":[]}}`, strings.Replace(strings.Replace(dnsOK, `"query_type":"A"`, `"query_type":"AAAA"`, 1), `{"type":"A","value":"1.1.1.1"}`, `{"type":"AAAA","value":"2606:4700:4700::1111"}`, 1)} {
		o, err := DecodeProtocol(Contract, []byte(s), protocolLimits)
		if err != nil {
			t.Fatal(err)
		}
		b, err := json.Marshal(o)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = DecodeProtocol(Contract, b, protocolLimits); err != nil {
			t.Fatal(err)
		}
	}
	cases := map[string]string{
		"unknown engine":           strings.Replace(dnsOK, `"udp"`, `"private"`, 1),
		"unrequested record":       strings.Replace(dnsOK, `"type":"A"`, `"type":"AAAA"`, 1),
		"raw record type":          strings.Replace(dnsOK, `"CNAME"`, `"TXT"`, 1),
		"uppercase name":           strings.Replace(dnsOK, `alias.example.org`, `Alias.example.org`, 1),
		"unicode name":             strings.Replace(dnsOK, `alias.example.org`, `café.example.org`, 1),
		"trailing name dot":        strings.Replace(dnsOK, `alias.example.org`, `alias.example.org.`, 1),
		"name label bound":         strings.Replace(dnsOK, `alias.example.org`, strings.Repeat("a", 64)+".org", 1),
		"non sequential index":     strings.Replace(dnsOK, `"arrival_index":1`, `"arrival_index":2`, 1),
		"zero index":               strings.Replace(dnsOK, `"arrival_index":1`, `"arrival_index":0`, 1),
		"fractional index":         strings.Replace(dnsOK, `"arrival_index":1`, `"arrival_index":1.0`, 1),
		"arrival outside duration": strings.Replace(dnsOK, `"arrival_offset_ms":250`, `"arrival_offset_ms":750`, 1),
		"arrival quantization":     strings.Replace(dnsOK, `"arrival_offset_ms":250`, `"arrival_offset_ms":251`, 1),
		"missing answers":          strings.Replace(dnsOK, `,"answers":[{"type":"CNAME","value":"alias.example.org"},{"type":"A","value":"1.1.1.1"}]`, "", 1),
		"null responses":           `{"type":"dns_query","status":"failed","duration_ms":0,"failure":"timeout","details":{"engine":"system","query_type":"A","responses":null}}`,
		"false limit flag":         strings.Replace(dnsOK, `"engine":"udp"`, `"engine":"udp","response_limit_reached":false`, 1),
		"premature limit flag":     strings.Replace(dnsOK, `"engine":"udp"`, `"engine":"udp","response_limit_reached":true`, 1),
		"rcode alias":              strings.Replace(dnsOK, `"NOERROR"`, `"noerror"`, 1),
		"raw code label":           strings.Replace(dnsOK, `"NOERROR"`, `"BADVERS"`, 1),
		"rcode numeric":            strings.Replace(dnsOK, `"NOERROR"`, `16`, 1),
		"code on success":          strings.Replace(dnsOK, `"NOERROR"`, `"OTHER"`, 1),
		"resolver address":         strings.Replace(dnsOK, `"engine":"udp"`, `"engine":"udp","resolver_address":"private"`, 1),
		"raw response":             strings.Replace(dnsOK, `"rcode":"NOERROR"`, `"rcode":"NOERROR","raw_message":"private"`, 1),
		"ttl":                      strings.Replace(dnsOK, `"value":"1.1.1.1"`, `"value":"1.1.1.1","ttl":300`, 1),
		"duplicate answers":        strings.Replace(dnsOK, `{"type":"CNAME","value":"alias.example.org"}`, `{"type":"A","value":"1.1.1.1"}`, 1),
	}
	for name, s := range cases {
		t.Run(name, func(t *testing.T) { assertProtocolRejected(t, s, protocolLimits) })
	}
	for _, ip := range []string{"10.0.0.1", "127.0.0.1", "169.254.1.1", "100.64.0.1", "0.1.2.3", "192.0.2.1", "198.18.0.1", "203.0.113.1", "198.51.100.1", "224.1.1.1", "240.1.1.1", "001.1.1.1", "::ffff:1.1.1.1", "fe80::1%en0"} {
		assertProtocolRejected(t, strings.Replace(dnsOK, `1.1.1.1`, ip, 1), protocolLimits)
	}
	l := protocolLimits
	l.RetainDNSAnswers = false
	assertProtocolRejected(t, dnsOK, l)
	// Complete original bytes, including whitespace, count toward the DNS bound.
	at := dnsOK + strings.Repeat(" ", 16384-len(dnsOK))
	if _, err := DecodeProtocol(Contract, []byte(at), protocolLimits); err != nil {
		t.Fatal(err)
	}
	assertProtocolRejected(t, at+" ", protocolLimits)
	o, _ := DecodeProtocol(Contract, []byte(dnsOK), protocolLimits)
	before, _ := json.Marshal(o)
	d, _ := o.DNS()
	d.Responses[0].Answers[0].Value = "private"
	d.Responses[0].RCode = "OTHER"
	after, _ := json.Marshal(o)
	if string(before) != string(after) {
		t.Fatal("DNS view alias")
	}
	for _, v := range []any{d, DNSAnswer{Value: "private"}, DNSResponse{RCode: "private"}} {
		if _, err := json.Marshal(v); err == nil {
			t.Fatal("serialized detached DNS view")
		}
	}
}

func TestDNSCollectionAndAnswerBoundaries(t *testing.T) {
	base, _ := DecodeProtocol(Contract, []byte(dnsOK), protocolLimits)
	d, _ := base.DNS()
	answer := d.Responses[0].Answers[1]
	for i := 0; i < 8; i++ {
		d.Responses[0].Answers = append(d.Responses[0].Answers, DNSAnswer{Type: "CNAME", Value: fmt.Sprintf("c%d.example.org", i)})
	}
	if _, err := NewDNS(Contract, "succeeded", 500, "", d, protocolLimits); err == nil {
		t.Fatal("nine CNAMEs accepted")
	}
	d.Responses[0].Answers = d.Responses[0].Answers[:len(d.Responses[0].Answers)-1]
	if _, err := NewDNS(Contract, "succeeded", 500, "", d, protocolLimits); err != nil {
		t.Fatal("eight CNAMEs rejected", err)
	}
	d.Responses[0].Answers = d.Responses[0].Answers[:0]
	for i := 0; i < 16; i++ {
		d.Responses[0].Answers = append(d.Responses[0].Answers, DNSAnswer{Type: "A", Value: fmt.Sprintf("1.1.1.%d", i+1)})
	}
	if _, err := NewDNS(Contract, "succeeded", 500, "", d, protocolLimits); err != nil {
		t.Fatal(err)
	}
	d.Responses[0].Answers = append(d.Responses[0].Answers, answer)
	if _, err := NewDNS(Contract, "succeeded", 500, "", d, protocolLimits); err == nil {
		t.Fatal("17 answer input accepted")
	}
	d, _ = base.DNS()
	d.Responses[0].Answers = append(d.Responses[0].Answers, answer)
	o, err := NewDNS(Contract, "succeeded", 500, "", d, protocolLimits)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := o.DNS()
	if len(got.Responses[0].Answers) != 2 || got.Responses[0].Answers[0].Type != "CNAME" {
		t.Fatal("duplicate removal/order")
	}
	d, _ = base.DNS()
	for i := 2; i <= 3; i++ {
		r := d.Responses[0]
		r.ArrivalIndex = i
		r.ArrivalOffsetMS = 500
		d.Responses = append(d.Responses, r)
	}
	d.ResponseLimitReached = true
	if _, err := NewDNS(Contract, "succeeded", 500, "", d, protocolLimits); err != nil {
		t.Fatal(err)
	}
	d.Engine = "tcp"
	if _, err := NewDNS(Contract, "succeeded", 500, "", d, protocolLimits); err == nil {
		t.Fatal("multi-response TCP accepted")
	}
	d.Engine = "udp"
	l := protocolLimits
	l.MaxDNSResponses = 2
	if _, err := NewDNS(Contract, "succeeded", 500, "", d, l); err == nil {
		t.Fatal("configured count bound bypass")
	}
	d.Responses = d.Responses[:2]
	d.Responses[0].ArrivalOffsetMS = 500
	d.Responses[1].ArrivalOffsetMS = 250
	if _, err := NewDNS(Contract, "succeeded", 500, "", d, l); err == nil {
		t.Fatal("reordered timing accepted")
	}
}

// requirement-ref: REQ-OBS-0045
// requirement-ref: REQ-OBS-0046
// requirement-ref: REQ-OBS-0047
// requirement-ref: REQ-OBS-0048
// requirement-ref: REQ-OBS-0049
// requirement-ref: REQ-OBS-0050
func TestAcceptedProducerMappings(t *testing.T) {
	for code, failure := range map[string]string{"FORMERR": "inconclusive", "SERVFAIL": "servfail", "NXDOMAIN": "nxdomain", "NOTIMP": "inconclusive", "REFUSED": "refused", "OTHER": "inconclusive"} {
		s := strings.Replace(strings.Replace(dnsOK, `"succeeded"`, `"failed","failure":"`+failure+`"`, 1), `"NOERROR"`, `"`+code+`"`, 1)
		if _, err := DecodeProtocol(Contract, []byte(s), protocolLimits); err != nil {
			t.Fatal(code, err)
		}
		assertProtocolRejected(t, strings.Replace(s, `"failure":"`+failure+`"`, `"failure":"no_data"`, 1), protocolLimits)
	}
	for code := 0; code <= 65535; code++ {
		got, present := DNSResponseCode(uint16(code), true)
		want := "OTHER"
		if code < 6 {
			want = []string{"NOERROR", "FORMERR", "SERVFAIL", "NXDOMAIN", "NOTIMP", "REFUSED"}[code]
		}
		if !present || got != want {
			t.Fatalf("DNS code %d", code)
		}
	}
	if got, present := DNSResponseCode(16, false); present || got != "" {
		t.Fatal("invented DNS code")
	}
	for _, code := range []uint16{0, 0x0301, 0x0303, 0x0304, 0x0305, 65535} {
		got, present := NegotiatedTLSVersion(code, true)
		want := "unknown"
		if code == 0x0303 {
			want = "TLS1.2"
		}
		if code == 0x0304 {
			want = "TLS1.3"
		}
		if !present || got != want {
			t.Fatal("TLS version mapping")
		}
	}
	if got, present := NegotiatedTLSVersion(0x0304, false); present || got != "" {
		t.Fatal("invented version")
	}
	for _, p := range []string{"http/1.1", "h2", "dot", "h3", "private"} {
		got, present := NegotiatedALPN(p, true)
		want := p
		if p == "h3" || p == "private" {
			want = "unknown"
		}
		if !present || got != want {
			t.Fatal("ALPN mapping")
		}
	}
	for _, p := range []string{"", "h2"} {
		got, present := NegotiatedALPN(p, false)
		if present || got != "" {
			t.Fatal("invented ALPN")
		}
	}
	if _, present := NegotiatedALPN("", true); present {
		t.Fatal("invented empty ALPN")
	}
	if err := ValidateALPNOffer(Contract, []string{"http/1.1", "h2", "dot"}); err != nil {
		t.Fatal(err)
	}
	for _, offer := range [][]string{{"unknown"}, {"h2c"}, {"h3"}, {"H2"}, {"private"}, {"h2", "h2"}, make([]string, 9)} {
		if ValidateALPNOffer(Contract, offer) == nil {
			t.Fatal("invalid ALPN offer accepted")
		}
	}
	if ValidateALPNOffer("0.2.0", nil) == nil {
		t.Fatal("offer version fallback")
	}
	groups := map[string][]int{"handshake_failure": {40, 71, 86, 120}, "certificate_error": {42, 43, 44, 45, 46, 48, 113, 116}, "protocol_error": {10, 20, 22, 47, 50, 51, 70, 109, 110}, "access_denied": {49, 112, 115}, "internal_error": {80}}
	expected := map[int]string{}
	for group, codes := range groups {
		for _, code := range codes {
			expected[code] = group
		}
	}
	for code := 0; code < 256; code++ {
		want := expected[code]
		if want == "" {
			want = "other"
		}
		got, present := PeerAlertClass(uint8(code), true)
		if !present || got != want {
			t.Fatalf("peer alert %d", code)
		}
	}
	if got, present := PeerAlertClass(42, false); present || got != "" {
		t.Fatal("invented peer alert class")
	}
}

// requirement-ref: REQ-OBS-0043
// requirement-ref: REQ-OBS-0050
func TestTLSVocabularyAndConsistency(t *testing.T) {
	o, err := DecodeProtocol(Contract, []byte(tlsOK), protocolLimits)
	if err != nil {
		t.Fatal(err)
	}
	d, _ := o.TLS()
	for _, version := range []string{"TLS1.2", "TLS1.3", "unknown", ""} {
		d.TLSVersion = version
		if _, err := NewTLS(Contract, "succeeded", 250, "", d); err != nil {
			t.Fatal(err)
		}
	}
	for _, s := range []string{
		strings.Replace(tlsOK, `TLS1.3`, `TLS1.0`, 1), strings.Replace(tlsOK, `"TLS1.3"`, `771`, 1), strings.Replace(tlsOK, `TLS1.3`, `tls1.3`, 1), strings.Replace(tlsOK, `"h2"`, `"private"`, 1), strings.Replace(tlsOK, `"h2"`, `""`, 1), strings.Replace(tlsOK, `"trusted"`, `"not_completed"`, 1), strings.Replace(tlsOK, `"trusted"`, `"trusted","alert":"other"`, 1), strings.Replace(tlsOK, `"trusted"`, `"trusted","certificate":"private"`, 1),
	} {
		assertProtocolRejected(t, s, protocolLimits)
	}
	d, _ = o.TLS()
	d.TLSVersion = ""
	d.ALPN = ""
	d.CertificateValidation = "not_completed"
	for _, class := range []string{"handshake_failure", "certificate_error", "protocol_error", "access_denied", "internal_error", "other", ""} {
		d.Alert = class
		a, err := NewTLS(Contract, "failed", 250, "alert", d)
		if err != nil {
			t.Fatal(err)
		}
		b, _ := json.Marshal(a)
		if _, err := DecodeProtocol(Contract, b, protocolLimits); err != nil {
			t.Fatal(err)
		}
	}
	d.Alert = "42"
	if _, err := NewTLS(Contract, "failed", 250, "alert", d); err == nil {
		t.Fatal("numeric alert label accepted")
	}
	d.Alert = "certificate_error"
	d.CertificateValidation = "trusted"
	if _, err := NewTLS(Contract, "failed", 250, "alert", d); err != nil {
		t.Fatal("peer certificate alert rewrote local validation")
	}
	d.Alert = ""
	d.CertificateValidation = "untrusted"
	if _, err := NewTLS(Contract, "failed", 250, "certificate_untrusted", d); err != nil {
		t.Fatal(err)
	}
	d.CertificateValidation = "trusted"
	if _, err := NewTLS(Contract, "failed", 250, "certificate_untrusted", d); err == nil {
		t.Fatal("local certificate mismatch accepted")
	}
	if _, err := json.Marshal(d); err == nil {
		t.Fatal("TLS detached serialization")
	}
}

// requirement-ref: REQ-OBS-0051
func TestIncompleteFailureAbsentForEveryType(t *testing.T) {
	for _, s := range []string{tcpOK, httpOK, dnsOK, tlsOK} {
		s = strings.Replace(s, `"status":"succeeded"`, `"status":"incomplete"`, 1)
		if _, err := DecodeProtocol(Contract, []byte(s), protocolLimits); err != nil {
			t.Fatal(err)
		}
		for _, failure := range []string{`null`, `""`, `"timeout"`, `"inconclusive"`} {
			assertProtocolRejected(t, strings.Replace(s, `"status":"incomplete"`, `"status":"incomplete","failure":`+failure, 1), protocolLimits)
		}
	}
}
