package trustdoc_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/trustdoc"
)

// syntheticKey returns a standard-library Ed25519 public key in the document's
// encoding. Synthetic only.
func syntheticKey(t *testing.T) string {
	t.Helper()
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(pub)
}

const limits = `{"attempt":{"duration_ms":60000,"bytes_sent":65536,"bytes_received":262144,"connections":12},
"max_runs":28,"max_attempts":2,"max_validity_seconds":604800,"min_repeat_interval_seconds":21600,"min_repeat_jitter_seconds":300,
"min_retry_delay_ms":30000,"max_retry_delay_ms":600000,"max_retry_jitter_ms":60000,"max_result_bytes":65536,"max_bundle_bytes":262144,
"max_results_per_bundle":16,"max_pending_bundles":32,"max_pending_bytes":4194304,"max_retention_seconds":86400}`

const permissions = `{"tests":[{"id":"TLS-01","version":"0.1.0"}],"registration_ids":["reg-tls"],"paths":{"direct":true,"proxy":false},
"max_risk":"medium","require_consent":true,"allow_medium_risk":true,"allow_composite":false,"allow_unknown_risk":false,
"label_namespaces":["org.example"],"limits":` + limits + `}`

func document(key string) string {
	return `{"format":"barometer-trust/1",
"policy":{"generation":7,"permissions":` + permissions + `,
 "issuers":[{"issuer_id":"remote-issuer","permissions":` + permissions + `}],
 "targets":[{"id":"reg-tls","test":{"id":"TLS-01","version":"0.1.0"},"role":"measurement",
   "tuple":{"address":"2001:db8::1","port":443,"scheme":"tls","tls_server_name":"example.org","alpn":["h2","http/1.1"]},
   "paths":{"direct":true,"proxy":false},"baseline_risk":"low","owner":"synthetic","purpose":"synthetic",
   "review_generation":3,"expires_at":"2030-01-02T03:04:05Z","redirect_ids":[]}],
 "keys":[{"key_id":"key-1","issuer_id":"remote-issuer","public_key":"` + key + `","not_before":"2026-01-01T00:00:00Z","not_after":"2026-12-31T00:00:00Z"}]},
"configuration_approvals":{"endpoints":["https://context.example.org/v1"],"connection_codes":["outline_mobileproxy"],"linkage_scopes":["scope.example"],"allow_linkage":true},
"context_approvals":{"connection_codes":["outline_mobileproxy"],"classification_versions":["1"],"country_codes":["DE"],"region_codes":[]}}`
}

func TestDecodeValidDocument(t *testing.T) {
	key := syntheticKey(t)
	doc, err := trustdoc.Decode([]byte(document(key)))
	if err != nil {
		t.Fatal(err)
	}
	p := doc.Policy
	if !p.Valid() || p.Generation() != 7 {
		t.Fatal("policy not decoded")
	}
	perm := p.Permissions()
	if perm.MaxRisk != task.MediumRisk || !perm.RequireConsent || perm.Paths.Proxy || perm.Limits.MaxPendingBytes != 4194304 || perm.Limits.Attempt.Connections != 12 || perm.LabelNamespaces[0] != "org.example" {
		t.Fatalf("permissions: %+v", perm)
	}
	reg, ok := p.Target("reg-tls")
	if !ok || reg.Tuple.Address.String() != "2001:db8::1" || reg.Tuple.Port != 443 || reg.Tuple.TLSServerName != "example.org" || len(reg.Tuple.ALPN) != 2 || reg.ReviewGeneration != 3 || reg.ExpiresAt.Whole.Year() != 2030 {
		t.Fatalf("target: %+v", reg)
	}
	k, ok := p.TrustedKey("key-1")
	if !ok || base64.StdEncoding.EncodeToString(k.PublicKey[:]) != key || k.IssuerID != "remote-issuer" {
		t.Fatal("key not decoded")
	}
	if _, ok := p.Issuer("remote-issuer"); !ok {
		t.Fatal("issuer not decoded")
	}
	if !doc.Approvals.AllowLinkage || doc.Approvals.Endpoints[0] != "https://context.example.org/v1" || doc.Context.CountryCodes[0] != "DE" {
		t.Fatal("approvals not decoded")
	}
	if s := fmt.Sprintf("%v %+v %#v", doc, doc, doc); strings.Contains(s, "reg-tls") || strings.Contains(s, key) {
		t.Fatal("formatting leaks trusted values:", s)
	}
}

// Every malformed or incomplete document rejects as a whole, with no detail.
func TestDecodeRejectsInvalidDocuments(t *testing.T) {
	key := syntheticKey(t)
	valid := document(key)
	replace := func(old, new string) string {
		if !strings.Contains(valid, old) {
			t.Fatalf("fixture lacks %q", old)
		}
		return strings.Replace(valid, old, new, 1)
	}
	cases := map[string]string{
		"empty":                 ``,
		"not an object":         `[]`,
		"other format":          replace(`"barometer-trust/1"`, `"barometer-trust/2"`),
		"unknown member":        replace(`"format"`, `"extra":1,"format"`),
		"duplicate member":      replace(`"format":"barometer-trust/1",`, `"format":"barometer-trust/1","format":"barometer-trust/1",`),
		"missing approvals":     replace(`"configuration_approvals"`, `"configuration_approval"`),
		"missing limit":         strings.Replace(valid, `"max_runs":28,`, ``, -1),
		"zero generation":       replace(`"generation":7`, `"generation":0`),
		"generation overflow":   replace(`"generation":7`, `"generation":9223372036854775808`),
		"fraction generation":   replace(`"generation":7`, `"generation":7.0`),
		"string generation":     replace(`"generation":7`, `"generation":"7"`),
		"unknown risk max":      strings.Replace(valid, `"max_risk":"medium"`, `"max_risk":"unknown"`, -1),
		"unknown role":          replace(`"role":"measurement"`, `"role":"probe"`),
		"non-canonical addr":    replace(`"2001:db8::1"`, `"2001:0db8::1"`),
		"zoned address":         replace(`"2001:db8::1"`, `"fe80::1%en0"`),
		"port out of range":     replace(`"port":443`, `"port":65536`),
		"fractional timestamp":  replace(`"2030-01-02T03:04:05Z"`, `"2030-01-02T03:04:05.5Z"`),
		"offset timestamp":      replace(`"2030-01-02T03:04:05Z"`, `"2030-01-02T05:04:05+02:00"`),
		"short key":             replace(key, key[:40]),
		"unpadded key":          replace(key, strings.TrimRight(key, "=")),
		"url-safe key":          replace(key, strings.NewReplacer("+", "-", "/", "_").Replace(key)+"x"),
		"key of unknown issuer": replace(`"issuer_id":"remote-issuer","public_key"`, `"issuer_id":"other-issuer","public_key"`),
		"unregistered target":   strings.Replace(valid, `"registration_ids":["reg-tls"]`, `"registration_ids":["reg-other"]`, -1),
		"invalid endpoint":      replace(`"https://context.example.org/v1"`, `"http://context.example.org/v1"`),
		"duplicate context":     replace(`"country_codes":["DE"]`, `"country_codes":["DE","DE"]`),
		"no issuers":            replace(`"issuers":[{"issuer_id":"remote-issuer","permissions":`+permissions+`}]`, `"issuers":[]`),
		"null list":             replace(`"redirect_ids":[]`, `"redirect_ids":null`),
		"trailing data":         valid + `{}`,
		"oversized":             replace(`"purpose":"synthetic"`, `"purpose":"`+strings.Repeat("a", trustdoc.MaxBytes)+`"`),
	}
	for name, in := range cases {
		doc, err := trustdoc.Decode([]byte(in))
		if !errors.Is(err, trustdoc.ErrInvalid) || doc.Policy.Valid() {
			t.Errorf("%s: accepted (%v)", name, err)
		}
		if err != nil && err.Error() != trustdoc.ErrInvalid.Error() {
			t.Errorf("%s: error carries detail: %v", name, err)
		}
	}
}
