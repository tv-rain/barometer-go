package observation

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

const accessOK = `{"network_type":"unknown","vpn_state":"unknown","ipv4_available":"unknown","ipv6_available":"unknown","validated":"unknown","network_changed":false}`
const contextOK = `{"access":` + accessOK + `,"egress":{"status":"fresh","lookup_source":"primary","network_path":"direct","country_code":"NL","region_code":"NL-NH","asn":4294967295,"classification_version":"0.1.0"}}`

func contextDecoder(t *testing.T) *ContextDecoder {
	t.Helper()
	d, err := NewContextDecoder(Contract, ContextApprovals{ConnectionCodes: []string{"reviewed_proxy"}, ClassificationVersions: []string{"0.1.0"}, CountryCodes: []string{"NL", "US"}, RegionCodes: []string{"NL-NH"}})
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// requirement-ref: REQ-OBS-0006
// requirement-ref: REQ-CONTEXT-FORMAT-0004
// requirement-ref: REQ-CONTEXT-FORMAT-0005
func TestAccessOnlySnapshot(t *testing.T) {
	a, err := DecodeAccess(Contract, []byte(accessOK), 1)
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != accessOK || !a.MatchesGeneration(1) || a.MatchesGeneration(2) {
		t.Fatal("invalid access-only representation or binding")
	}
	changed, err := DecodeAccess(Contract, []byte(strings.Replace(accessOK, `false`, `true`, 1)), 1)
	if err != nil || changed.MatchesGeneration(1) {
		t.Fatal("changed access snapshot eligible")
	}
	if _, err := DecodeAccess("unknown", []byte(accessOK), 1); !errors.Is(err, ErrVersion) {
		t.Fatal(err)
	}
	for _, s := range []string{strings.Replace(accessOK, `false`, `null`, 1), strings.Replace(accessOK, `{`, `{"device_id":"private",`, 1), accessOK + `{}`} {
		if _, err := DecodeAccess(Contract, []byte(s), 1); err == nil {
			t.Fatal("accepted invalid access snapshot")
		}
	}
	if _, err := DecodeAccess(Contract, []byte(accessOK), 0); err == nil {
		t.Fatal("accepted zero generation")
	}
	var zero AccessSnapshot
	if _, err := json.Marshal(zero); err == nil {
		t.Fatal("serialized zero access snapshot")
	}
	before, _ := json.Marshal(a)
	if err := json.Unmarshal([]byte(accessOK), &a); !errors.Is(err, ErrImport) {
		t.Fatal(err)
	}
	after, _ := json.Marshal(a)
	if string(before) != string(after) {
		t.Fatal("failed import changed model")
	}
	n, err := contextDecoder(t).Decode([]byte(contextOK), directPath(t, 1))
	if err != nil {
		t.Fatal(err)
	}
	projected, err := json.Marshal(n.AccessSnapshot())
	if err != nil || string(projected) != accessOK {
		t.Fatal("invalid access projection", err)
	}
}
func directPath(t *testing.T, g uint64) Path {
	t.Helper()
	p, err := NewPath(g, 0, "direct", "")
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// requirement-ref: REQ-CONTEXT-FORMAT-0004
// requirement-ref: REQ-CONTEXT-FORMAT-0007
// requirement-ref: REQ-CONTEXT-FORMAT-0008
// requirement-ref: REQ-CONTEXT-FORMAT-0009
func TestContextUnknownAndFreshnessStates(t *testing.T) {
	d := contextDecoder(t)
	path := directPath(t, 1)
	for _, s := range []string{
		contextOK,
		strings.Replace(strings.Replace(contextOK, `"fresh"`, `"cached"`, 1), `"primary"`, `"cache"`, 1),
		`{"access":` + accessOK + `,"egress":{"status":"fresh","lookup_source":"fallback","network_path":"direct","classification_version":"0.1.0"}}`,
		`{"access":` + accessOK + `,"egress":{"status":"unavailable","lookup_source":"none","network_path":"direct"}}`,
	} {
		n, err := d.Decode([]byte(s), path)
		if err != nil {
			t.Fatal(err)
		}
		if !n.Matches(path) {
			t.Fatal("lost path")
		}
		b, err := json.Marshal(n)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = d.Decode(b, path); err != nil {
			t.Fatal(err)
		}
	}
	// Missing classification is preserved; no inferred country, ASN or fallback.
	n, err := d.Decode([]byte(`{"access":`+accessOK+`,"egress":{"status":"fresh","lookup_source":"primary","network_path":"direct","classification_version":"0.1.0"}}`), path)
	if err != nil {
		t.Fatal(err)
	}
	if n.Egress().CountryCode != "" || n.Egress().ASN != nil {
		t.Fatal("invented classification")
	}
}

// requirement-ref: REQ-CONTEXT-FORMAT-0017
// requirement-ref: REQ-CONTEXT-FORMAT-0019
// requirement-ref: REQ-CONTEXT-FORMAT-0020
// requirement-ref: REQ-CONTEXT-FORMAT-0022
func TestContextRejectionAndNormalization(t *testing.T) {
	d := contextDecoder(t)
	path := directPath(t, 1)
	cases := map[string]string{
		"missing required boolean":   strings.Replace(contextOK, `,"network_changed":false`, "", 1),
		"null boolean":               strings.Replace(contextOK, `false`, `null`, 1),
		"unknown transport":          strings.Replace(contextOK, `"network_type":"unknown"`, `"network_type":"private"`, 1),
		"unknown source":             strings.Replace(contextOK, `"primary"`, `"unknown"`, 1),
		"cache source with fresh":    strings.Replace(contextOK, `"primary"`, `"cache"`, 1),
		"cached with primary":        strings.Replace(contextOK, `"fresh"`, `"cached"`, 1),
		"unavailable classification": strings.Replace(strings.Replace(contextOK, `"fresh"`, `"unavailable"`, 1), `"primary"`, `"none"`, 1),
		"version absent":             strings.Replace(contextOK, `,"classification_version":"0.1.0"`, "", 1),
		"unapproved version":         strings.Replace(contextOK, `"0.1.0"`, `"private"`, 1),
		"unapproved country":         strings.Replace(contextOK, `"NL"`, `"ZZ"`, 1),
		"region mismatch":            strings.Replace(contextOK, `"NL"`, `"US"`, 1),
		"region without country":     strings.Replace(contextOK, `"country_code":"NL",`, "", 1),
		"unapproved region":          strings.Replace(contextOK, `"NL-NH"`, `"NL-XY"`, 1),
		"direct code":                strings.Replace(contextOK, `"network_path":"direct"`, `"network_path":"direct","connection_code":"reviewed_proxy"`, 1),
		"asn zero":                   strings.Replace(contextOK, `4294967295`, `0`, 1),
		"asn over uint32":            strings.Replace(contextOK, `4294967295`, `4294967296`, 1),
		"asn float":                  strings.Replace(contextOK, `4294967295`, `1.0`, 1),
		"duplicate":                  strings.Replace(contextOK, `"asn":4294967295`, `"asn":1,"asn":2`, 1),
		"oversize":                   contextOK + strings.Repeat(" ", 4096),
	}
	for name, s := range cases {
		t.Run(name, func(t *testing.T) { assertContextRejected(t, d, s, path) })
	}
	n, err := d.Decode([]byte(strings.ReplaceAll(contextOK, `NL`, `nl`)), path)
	if err != nil {
		t.Fatal(err)
	}
	if n.Egress().CountryCode != "NL" || n.Egress().RegionCode != "NL-NH" {
		t.Fatal("not normalized")
	}
	if _, err := NewContextDecoder("0.2.0", ContextApprovals{}); !errors.Is(err, ErrVersion) {
		t.Fatal(err)
	}
	var zero NetworkContext
	if _, err := json.Marshal(zero); err == nil {
		t.Fatal("serialized zero")
	}
	if err := json.Unmarshal([]byte(contextOK), &zero); !errors.Is(err, ErrImport) {
		t.Fatal(err)
	}
}

// requirement-ref: REQ-CONTEXT-0017
// requirement-ref: REQ-CONTEXT-FORMAT-0016
func TestContextPathGenerationAndPrivacy(t *testing.T) {
	for _, view := range []any{Access{NetworkType: "private"}, Egress{CountryCode: "private"}} {
		if _, err := json.Marshal(view); err == nil {
			t.Fatal("serialized unvalidated detached view")
		}
	}
	d := contextDecoder(t)
	path := directPath(t, 1)
	n, err := d.Decode([]byte(contextOK), path)
	if err != nil {
		t.Fatal(err)
	}
	other := directPath(t, 2)
	if n.Matches(other) {
		t.Fatal("crossed access generation")
	}
	proxy, _ := NewPath(1, 1, "proxy", "reviewed_proxy")
	if n.Matches(proxy) {
		t.Fatal("crossed path")
	}
	assertContextRejected(t, d, contextOK, proxy)
	proxied := strings.Replace(contextOK, `"network_path":"direct"`, `"network_path":"proxy","connection_code":"reviewed_proxy"`, 1)
	n, err = d.Decode([]byte(proxied), proxy)
	if err != nil {
		t.Fatal(err)
	}
	newProxy, _ := NewPath(1, 2, "proxy", "reviewed_proxy")
	if n.Matches(newProxy) {
		t.Fatal("crossed proxy generation")
	}
	changed, err := d.Decode([]byte(strings.Replace(proxied, `false`, `true`, 1)), proxy)
	if err != nil {
		t.Fatal(err)
	}
	if changed.Matches(proxy) {
		t.Fatal("changed network considered stable")
	}
	before, _ := json.Marshal(n)
	e := n.Egress()
	*e.ASN = 1
	a := n.Access()
	a.NetworkType = "private"
	after, _ := json.Marshal(n)
	if string(before) != string(after) {
		t.Fatal("mutable context alias")
	}
	if _, err := json.Marshal(proxy); err == nil {
		t.Fatal("serialized generation metadata")
	}
	for _, name := range []string{"ip", "endpoint_url", "proxy_host", "network_generation", "proxy_generation", "device_id", "ssid", "resolver_address", "provider", "operator", "context_id"} {
		assertContextRejected(t, d, strings.Replace(contextOK, `{`, `{"`+name+`":"private",`, 1), path)
		assertContextRejected(t, d, strings.Replace(contextOK, `"access":{`, `"access":{"`+name+`":"private",`, 1), path)
		assertContextRejected(t, d, strings.Replace(contextOK, `"egress":{`, `"egress":{"`+name+`":"private",`, 1), path)
		if strings.Contains(string(after), `"`+name+`":`) {
			t.Fatalf("serialized forbidden field %s", name)
		}
	}
	assertContextRejected(t, d, strings.Replace(proxied, `,"connection_code":"reviewed_proxy"`, "", 1), proxy)
	assertContextRejected(t, d, proxied, Path{})
}

func TestContextApprovalsCopiedAndBounded(t *testing.T) {
	a := ContextApprovals{ClassificationVersions: []string{"0.1.0"}, CountryCodes: []string{"NL"}, RegionCodes: []string{"NL-NH"}}
	d, err := NewContextDecoder(Contract, a)
	if err != nil {
		t.Fatal(err)
	}
	a.ClassificationVersions[0] = "private"
	a.CountryCodes[0] = "ZZ"
	a.RegionCodes[0] = "ZZ-X"
	if _, err := d.Decode([]byte(contextOK), directPath(t, 1)); err != nil {
		t.Fatal("approval alias", err)
	}
	for _, a := range []ContextApprovals{{ClassificationVersions: []string{""}}, {CountryCodes: []string{"nl"}}, {RegionCodes: []string{"NL-NH"}}, {ConnectionCodes: []string{"private@invalid"}}, {CountryCodes: []string{"NL", "NL"}}} {
		if _, err := NewContextDecoder(Contract, a); err == nil {
			t.Fatal("accepted invalid approvals")
		}
	}
	for _, p := range []Path{{}, {network: 1, name: "direct", proxy: 1}, {network: 1, name: "proxy", code: "reviewed_proxy"}} {
		if _, err := NewPath(p.network, p.proxy, p.name, p.code); err == nil {
			t.Fatal("invalid path")
		}
	}
}

func assertContextRejected(t *testing.T, d *ContextDecoder, s string, path Path) {
	t.Helper()
	n, err := d.Decode([]byte(s), path)
	if err == nil {
		t.Fatal("accepted invalid context")
	}
	if _, err = json.Marshal(n); err == nil {
		t.Fatal("partial context escaped")
	}
}

func FuzzContext(f *testing.F) {
	f.Add([]byte(contextOK))
	f.Fuzz(func(t *testing.T, input []byte) {
		d := contextDecoder(t)
		path := directPath(t, 1)
		n, err := d.Decode(input, path)
		if err != nil {
			return
		}
		b, err := json.Marshal(n)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = d.Decode(b, path); err != nil {
			t.Fatal(err)
		}
	})
}
