package trustdoc_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
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

// restricted adds the optional local restriction inputs to the valid document.
func restricted(key, members string) string {
	return strings.Replace(document(key), `"keys":[`, members+`,"keys":[`, 1)
}

const digestHex = "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"

const restrictions = `"release_reviewed_tasks":[{"issuer_id":"remote-issuer","task_sha256":"` + digestHex + `"}],
"device":{"minimum_battery_percent":30,"charging_only":true},
"network_context":{"paths":{"direct":true,"proxy":false},"endpoints":["https://context.example.org/v1"]}`

// The optional members decode into the policy; absent members restrict
// nothing and list no release-reviewed task.
func TestDecodeLocalRestrictionInputs(t *testing.T) {
	key := syntheticKey(t)
	doc, err := trustdoc.Decode([]byte(restricted(key, restrictions)))
	if err != nil {
		t.Fatal(err)
	}
	var digest [32]byte
	if raw, err := hex.DecodeString(digestHex); err != nil || copy(digest[:], raw) != 32 {
		t.Fatal("bad digest fixture")
	}
	p := doc.Policy
	if !p.ReleaseReviewed("remote-issuer", digest) || p.ReleaseReviewed("remote-issuer", [32]byte{}) {
		t.Fatal("release-reviewed task not decoded")
	}
	if d := p.Device(); d.MinimumBatteryPercent == nil || *d.MinimumBatteryPercent != 30 || !d.ChargingOnly {
		t.Fatalf("device: %+v", d)
	}
	c, ok := p.ContextEligibility()
	if !ok || !c.Permits("direct") || c.Permits("proxy") || len(c.Endpoints) != 1 {
		t.Fatalf("context eligibility: %+v", c)
	}
	charging := strings.Replace(restrictions, `"minimum_battery_percent":30,`, ``, 1)
	forbidding := strings.Replace(restrictions, `{"direct":true,"proxy":false},"endpoints":["https://context.example.org/v1"]`, `{"direct":false,"proxy":false},"endpoints":[]`, 1)
	for name, in := range map[string]string{"charging only": charging, "forbidden context": forbidding} {
		doc, err := trustdoc.Decode([]byte(restricted(key, in)))
		if err != nil {
			t.Fatal(name, err)
		}
		if d := doc.Policy.Device(); name == "charging only" && (d.MinimumBatteryPercent != nil || !d.ChargingOnly) {
			t.Fatalf("%s: device %+v", name, d)
		}
		if c, ok := doc.Policy.ContextEligibility(); name == "forbidden context" && (!ok || c.Permits("direct") || c.Permits("proxy")) {
			t.Fatalf("%s: context %+v", name, c)
		}
	}
	doc, err = trustdoc.Decode([]byte(document(key)))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := doc.Policy.ContextEligibility(); ok || doc.Policy.Device().Active() || doc.Policy.ReleaseReviewed("remote-issuer", digest) {
		t.Fatal("absent members restrict or list something")
	}
	if s := fmt.Sprintf("%v %+v %#v", doc, doc, doc); strings.Contains(s, digestHex) {
		t.Fatal("formatting leaks trusted values:", s)
	}
}

// Malformed, ambiguous or out-of-range restriction inputs reject the whole
// document, with no detail.
func TestDecodeRejectsInvalidLocalRestrictions(t *testing.T) {
	key := syntheticKey(t)
	edit := func(old, new string) string {
		if !strings.Contains(restrictions, old) {
			t.Fatalf("fixture lacks %q", old)
		}
		return restricted(key, strings.Replace(restrictions, old, new, 1))
	}
	endpoints := make([]string, 65)
	for i := range endpoints {
		endpoints[i] = fmt.Sprintf(`"https://context%d.example.org/v1"`, i)
	}
	entries := make([]string, 257)
	for i := range entries {
		entries[i] = fmt.Sprintf(`{"issuer_id":"remote-issuer","task_sha256":"%064x"}`, i+1)
	}
	cases := map[string]string{
		"uppercase digest":      edit(digestHex, strings.ToUpper(digestHex)),
		"short digest":          edit(digestHex, digestHex[:62]),
		"long digest":           edit(digestHex, digestHex+"00"),
		"non-hex digest":        edit(digestHex, "zz"+digestHex[2:]),
		"zero digest":           edit(digestHex, strings.Repeat("0", 64)),
		"unlisted issuer":       edit(`"issuer_id":"remote-issuer"`, `"issuer_id":"other-issuer"`),
		"duplicate entry":       edit(`"task_sha256":"`+digestHex+`"}]`, `"task_sha256":"`+digestHex+`"},{"issuer_id":"remote-issuer","task_sha256":"`+digestHex+`"}]`),
		"too many entries":      edit(`[{"issuer_id":"remote-issuer","task_sha256":"`+digestHex+`"}]`, `[`+strings.Join(entries, ",")+`]`),
		"entry unknown member":  edit(`"task_sha256"`, `"origin":"embedded","task_sha256"`),
		"null list":             edit(`"release_reviewed_tasks":[`, `"release_reviewed_tasks":null,"x":[`),
		"battery above 100":     edit(`"minimum_battery_percent":30`, `"minimum_battery_percent":101`),
		"negative battery":      edit(`"minimum_battery_percent":30`, `"minimum_battery_percent":-1`),
		"fractional battery":    edit(`"minimum_battery_percent":30`, `"minimum_battery_percent":30.5`),
		"string battery":        edit(`"minimum_battery_percent":30`, `"minimum_battery_percent":"30"`),
		"missing charging":      edit(`,"charging_only":true`, ``),
		"device unknown member": edit(`"charging_only":true`, `"charging_only":true,"allow_charging_only":false`),
		"empty device":          edit(`{"minimum_battery_percent":30,"charging_only":true}`, `{}`),
		"null device":           edit(`{"minimum_battery_percent":30,"charging_only":true}`, `null`),
		"http endpoint":         edit(`"https://context.example.org/v1"`, `"http://context.example.org/v1"`),
		"endpoint with user":    edit(`"https://context.example.org/v1"`, `"https://user@context.example.org/v1"`),
		"endpoint fragment":     edit(`"https://context.example.org/v1"`, `"https://context.example.org/v1#x"`),
		"duplicate endpoint":    edit(`["https://context.example.org/v1"]`, `["https://context.example.org/v1","https://context.example.org/v1"]`),
		"too many endpoints":    edit(`["https://context.example.org/v1"]`, `[`+strings.Join(endpoints, ",")+`]`),
		"missing paths":         edit(`"paths":{"direct":true,"proxy":false},`, ``),
		"missing endpoints":     edit(`,"endpoints":["https://context.example.org/v1"]`, ``),
		"context unknown field": edit(`"endpoints"`, `"required_fields":["country_code"],"endpoints"`),
		"duplicate device":      edit(`"device":`, `"device":{"charging_only":false},"device":`),
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

// FuzzDecode checks that arbitrary bytes either decode to a valid policy or
// fail with the fixed ErrInvalid, deterministically and without panicking.
func FuzzDecode(f *testing.F) {
	key := base64.StdEncoding.EncodeToString(make([]byte, 32))
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err == nil {
		key = base64.StdEncoding.EncodeToString(pub)
	}
	f.Add([]byte(document(key)))
	f.Add([]byte(restricted(key, restrictions)))
	f.Add([]byte(`{"format":"barometer-trust/1"}`))
	f.Fuzz(func(t *testing.T, b []byte) {
		doc, err := trustdoc.Decode(b)
		again, err2 := trustdoc.Decode(b)
		if (err == nil) != (err2 == nil) || doc.Policy.Valid() != again.Policy.Valid() {
			t.Fatal("non-deterministic decoding")
		}
		if err != nil {
			if err != trustdoc.ErrInvalid || doc.Policy.Valid() {
				t.Fatal("rejection is not the fixed ErrInvalid with no policy")
			}
			return
		}
		if !doc.Policy.Valid() {
			t.Fatal("accepted document without a valid policy")
		}
	})
}
