package dns_test

import (
	"encoding/binary"
	"fmt"
	"strings"
	"testing"

	"github.com/tv-rain/barometer-go/internal/protocol/dns"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

var question = dns.Question{ID: 1234, Name: "example.org", Type: "A"}

func query(t *testing.T) []byte {
	t.Helper()
	b, e := dns.Query(question)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func response(t *testing.T, records ...loopback.DNSRecord) []byte {
	t.Helper()
	return loopback.DNSResponse(query(t), 0, records)
}
func padded(b []byte, size int) []byte {
	out := append([]byte(nil), b...)
	out[10] = 0
	out[11] = 1
	n := size - len(out) - 11
	out = append(out, 0, 0, 41, 16, 0, 0, 0, 0, 0, byte(n>>8), byte(n))
	out = append(out, 0, 1, byte((n-4)>>8), byte(n-4))
	out = append(out, make([]byte, n-4)...)
	return out
}

// requirement-ref: REQ-LIMIT-0014
// requirement-ref: REQ-DNS-01-0033
func TestWireMessageBoundaries(t *testing.T) {
	base := response(t, loopback.DNSRecord{Type: "A", Value: "1.1.1.1"})
	for _, maxBytes := range []int{dns.MaxUDP, dns.MaxTCP} {
		for _, delta := range []int{-1, 0, 1} {
			if maxBytes+delta > 65535 {
				continue
			}
			b := padded(base, maxBytes+delta)
			_, err := dns.Parse(b, question, maxBytes)
			if (err == nil) != (delta <= 0) {
				t.Fatalf("%d %+d: %v", maxBytes, delta, err)
			}
		}
	}
}

// requirement-ref: REQ-DNS-01-0022
// requirement-ref: REQ-DNS-01-0036
// requirement-ref: REQ-OBS-0013
func TestWireRetainsOnlyRelevantPublicAnswers(t *testing.T) {
	b := response(t, loopback.DNSRecord{Type: "A", Value: "9.9.9.9", Owner: "unrelated.example"}, loopback.DNSRecord{Type: "CNAME", Value: "alias.example.org"}, loopback.DNSRecord{Type: "A", Owner: "alias.example.org", Value: "1.1.1.1"}, loopback.DNSRecord{Type: "A", Owner: "alias.example.org", Value: "1.1.1.1"}, loopback.DNSRecord{Type: "A", Owner: "alias.example.org", Value: "127.0.0.1"}, loopback.DNSRecord{Type: "AAAA", Owner: "alias.example.org", Value: "2606:4700:4700::1111"})
	r, e := dns.Parse(b, question, dns.MaxUDP)
	if e != nil || len(r.Answers) != 2 || r.Answers[0].Type != "CNAME" || r.Answers[1].Value != "1.1.1.1" {
		t.Fatalf("%+v %v", r, e)
	}
}
func TestWireRejectsHostileFrames(t *testing.T) {
	base := response(t, loopback.DNSRecord{Type: "A", Value: "1.1.1.1"})
	mutations := []func([]byte) []byte{
		func(b []byte) []byte { return b[:11] }, func(b []byte) []byte { b[2] &= 0x7f; return b }, func(b []byte) []byte { b[2] |= 2; return b }, func(b []byte) []byte { b[3] |= 0x40; return b },
		func(b []byte) []byte { b[5] = 2; return b }, func(b []byte) []byte { b[7] = 255; return b }, func(b []byte) []byte { return append(b, 0) }, func(b []byte) []byte { b[12] = 0xc0; b[13] = 12; return b }, func(b []byte) []byte { b[12] = 0x40; return b }, func(b []byte) []byte { return b[:len(b)-1] },
	}
	for i, mutate := range mutations {
		if _, e := dns.Parse(mutate(append([]byte(nil), base...)), question, dns.MaxUDP); e == nil {
			t.Fatalf("accepted mutation %d", i)
		}
	}
	for _, mutate := range []func([]byte){func(b []byte) { b[0] ^= 1 }, func(b []byte) { b[13] = 'z' }, func(b []byte) { b[26] = 28 }} {
		b := append([]byte(nil), base...)
		mutate(b)
		if _, e := dns.Parse(b, question, dns.MaxUDP); e == nil {
			t.Fatal("accepted query mismatch")
		}
	}
}

// requirement-ref: REQ-LIMIT-0015
func TestWireAnswerAndCNAMEBounds(t *testing.T) {
	// Use numeric IP labels instead of non-address fixtures.
	for _, count := range []int{15, 16, 17} {
		var records []loopback.DNSRecord
		for i := range count {
			records = append(records, loopback.DNSRecord{Type: "AAAA", Value: fmt.Sprintf("2606:4700:4700::%x", i+1)})
		}
		q := question
		q.Type = "AAAA"
		qb, _ := dns.Query(q)
		b := loopback.DNSResponse(qb, 0, records)
		r, e := dns.Parse(b, q, dns.MaxUDP)
		if (e == nil) != (count <= 16) {
			t.Fatalf("answers %d: %+v %v", count, r, e)
		}
	}
	for _, count := range []int{7, 8, 9} {
		var records []loopback.DNSRecord
		owner := "example.org"
		for i := range count {
			target := strings.Repeat("a", i+1) + ".example.org"
			records = append(records, loopback.DNSRecord{Type: "CNAME", Owner: owner, Value: target})
			owner = target
		}
		records = append(records, loopback.DNSRecord{Type: "A", Owner: owner, Value: "1.1.1.1"})
		_, e := dns.Parse(response(t, records...), question, dns.MaxUDP)
		if (e == nil) != (count <= 8) {
			t.Fatalf("CNAMEs %d: %v", count, e)
		}
	}
	if _, e := dns.Parse(response(t, loopback.DNSRecord{Type: "CNAME", Value: "example.org"}), question, dns.MaxUDP); e == nil {
		t.Fatal("accepted CNAME cycle")
	}
}
func TestWireExtendedRcodeAndEDNSBounds(t *testing.T) {
	b := padded(response(t), 128)
	at := len(response(t))
	b[at+5] = 1
	r, e := dns.Parse(b, question, dns.MaxUDP)
	if e != nil || r.RCode != "OTHER" {
		t.Fatalf("%+v %v", r, e)
	}
	binary.BigEndian.PutUint16(b[at+13:], 65535)
	if _, e = dns.Parse(b, question, dns.MaxUDP); e == nil {
		t.Fatal("accepted malformed EDNS option")
	}
}
func FuzzParse(f *testing.F) {
	q, _ := dns.Query(question)
	f.Add(loopback.DNSResponse(q, 0, nil))
	f.Add([]byte{0xc0, 12})
	f.Fuzz(func(t *testing.T, b []byte) { _, _ = dns.Parse(b, question, dns.MaxTCP) })
}

func TestWireLabelBoundariesCannotAliasQuestion(t *testing.T) {
	// One label containing a literal dot is distinct from two hostname labels.
	q := query(t)
	b := append([]byte(nil), q[:12]...)
	b = append(b, 11)
	b = append(b, []byte("example.org")...)
	b = append(b, 0, 0, 1, 0, 1)
	b[2] |= 0x80
	if _, e := dns.Parse(b, question, dns.MaxUDP); e == nil {
		t.Fatal("literal label separator aliased question")
	}
	// Non-hostname labels on unrelated records remain syntactically valid DNS.
	good := response(t, loopback.DNSRecord{Type: "A", Owner: "_unrelated.example.org", Value: "9.9.9.9"}, loopback.DNSRecord{Type: "A", Value: "1.1.1.1"})
	r, e := dns.Parse(good, question, dns.MaxUDP)
	if e != nil || len(r.Answers) != 1 || r.Answers[0].Value != "1.1.1.1" {
		t.Fatalf("unrelated label %+v %v", r, e)
	}
	bad := response(t, loopback.DNSRecord{Type: "CNAME", Value: "alias.example.org"}, loopback.DNSRecord{Type: "A", Value: "1.1.1.1"})
	if _, e := dns.Parse(bad, question, dns.MaxUDP); e == nil {
		t.Fatal("mixed CNAME/address owner accepted")
	}
}
