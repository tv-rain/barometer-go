package config_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/config"
	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/platform"
)

const proxy = `{"host":"127.0.0.1","port":54321,"connection_code":"outline_mobileproxy"}`
const network = `{"primary_endpoint":"https://context.example/network-context","fallback_endpoint":"https://backup.example/network-context"}`
const linkage = `{"mode":"rotating_installation","scope":"org.example.production","rotation_period_seconds":604800}`
const complete = `{"proxy":` + proxy + `,"network_context":` + network + `,"result_linkage":` + linkage + `}`

func approvals() config.Approvals {
	return config.Approvals{
		Endpoints:       []string{"https://context.example/network-context", "https://backup.example/network-context"},
		ConnectionCodes: []string{"outline_mobileproxy", strings.Repeat("a", 64)},
		LinkageScopes:   []string{"org.example.production", strings.Repeat("a", 128)}, AllowLinkage: true,
	}
}
func decoder(t testing.TB) *config.Decoder {
	t.Helper()
	d, err := config.NewDecoder(config.Contract, approvals())
	if err != nil {
		t.Fatal(err)
	}
	return d
}
func decode(t testing.TB, d *config.Decoder, input string) config.Snapshot {
	t.Helper()
	s, err := d.Decode([]byte(input))
	if err != nil || !s.Valid() {
		t.Fatalf("decode failed: %v", err)
	}
	return s
}

func TestDefaultsAndComplete(t *testing.T) {
	d := decoder(t)
	s := decode(t, d, `{}`)
	if _, ok := s.Proxy(); ok {
		t.Fatal("proxy default")
	}
	if _, ok := s.NetworkContext(); ok {
		t.Fatal("egress default")
	}
	if s.Linkage() != (config.Linkage{Mode: "none"}) {
		t.Fatal("linkage default")
	}
	s = decode(t, d, complete)
	if p, ok := s.Proxy(); !ok || p.Host != "127.0.0.1" || p.Port != 54321 || p.ConnectionCode != "outline_mobileproxy" {
		t.Fatal("proxy")
	}
	if n, ok := s.NetworkContext(); !ok || n.PrimaryEndpoint != "https://context.example/network-context" || n.FallbackEndpoint != "https://backup.example/network-context" {
		t.Fatal("network")
	}
	if s.Linkage() != (config.Linkage{Mode: "rotating_installation", Scope: "org.example.production", RotationPeriodSeconds: 604800}) {
		t.Fatal("linkage")
	}
	for _, host := range []string{"localhost", "127.0.0.2", "::1"} {
		decode(t, d, `{"proxy":`+strings.Replace(proxy, "127.0.0.1", host, 1)+`}`)
	}
	decode(t, d, `{"result_linkage":{"mode":"none"}}`)
	decode(t, d, `{"network_context":{"primary_endpoint":"https://context.example/network-context"}}`)
}

// requirement-ref: REQ-JSON-0004
// requirement-ref: REQ-JSON-0005
// requirement-ref: REQ-JSON-0010
func TestAtomicRejectionCorpus(t *testing.T) {
	d := decoder(t)
	cases := []string{
		``, `null`, `[]`, `true`, `{`, `{} {}`, `{"proxy":null}`, `{"network_context":null}`, `{"result_linkage":null}`,
		`{"Proxy":` + proxy + `}`, `{"proxy":{}}`, `{"proxy":` + strings.Replace(proxy, `54321`, `"54321"`, 1) + `}`,
		`{"proxy":` + strings.Replace(proxy, `54321`, `1.0`, 1) + `}`, `{"proxy":` + strings.Replace(proxy, `54321`, `1e0`, 1) + `}`,
		`{"proxy":` + strings.Replace(proxy, `54321`, `-0`, 1) + `}`, `{"proxy":` + strings.Replace(proxy, `54321`, `65536`, 1) + `}`,
		`{"proxy":` + strings.Replace(proxy, `54321`, `0`, 1) + `}`,
		`{"proxy":` + strings.Replace(proxy, `127.0.0.1`, `192.168.1.1`, 1) + `}`,
		`{"proxy":` + strings.Replace(proxy, `127.0.0.1`, `proxy.example`, 1) + `}`,
		`{"proxy":` + strings.Replace(proxy, `127.0.0.1`, `[::1]`, 1) + `}`,
		`{"proxy":` + strings.Replace(proxy, `127.0.0.1`, `::1%lo0`, 1) + `}`,
		`{"proxy":` + strings.Replace(proxy, `outline_mobileproxy`, `Outline`, 1) + `}`,
		`{"proxy":` + strings.Replace(proxy, `outline_mobileproxy`, strings.Repeat("a", 65), 1) + `}`,
		`{"proxy":` + strings.Replace(proxy, `outline_mobileproxy`, ``, 1) + `}`,
		`{"network_context":{}}`, `{"network_context":{"fallback_endpoint":"https://backup.example/network-context"}}`,
		`{"network_context":{"primary_endpoint":"http://context.example/"}}`,
		`{"network_context":{"primary_endpoint":"https://user:password@context.example/"}}`,
		`{"network_context":{"primary_endpoint":"https://context.example/#"}}`,
		`{"network_context":{"primary_endpoint":"https://127.0.0.1/"}}`,
		`{"network_context":{"primary_endpoint":"https://10.0.0.1/"}}`,
		`{"network_context":{"primary_endpoint":"https://[::1]/"}}`,
		`{"network_context":{"primary_endpoint":"https://localhost/"}}`,
		`{"network_context":{"primary_endpoint":"https://context.local/"}}`,
		`{"network_context":{"primary_endpoint":"https://context.example:65536/"}}`,
		`{"network_context":{"primary_endpoint":"https://context.example:/"}}`,
		`{"network_context":{"primary_endpoint":"https://context.example/","fallback_endpoint":""}}`,
		`{"result_linkage":{}}`, `{"result_linkage":{"mode":"stable_installation"}}`,
		`{"result_linkage":{"mode":"none","scope":"org.example.production"}}`,
		`{"result_linkage":{"mode":"none","rotation_period_seconds":86400}}`,
		`{"result_linkage":{"mode":"rotating_installation","scope":"org.example.production"}}`,
		`{"result_linkage":` + strings.Replace(linkage, `604800`, `86400.0`, 1) + `}`,
		`{"result_linkage":` + strings.Replace(linkage, `604800`, `86399`, 1) + `}`,
		`{"result_linkage":` + strings.Replace(linkage, `604800`, `604801`, 1) + `}`,
		`{"result_linkage":` + strings.Replace(linkage, `org.example.production`, `INVALID`, 1) + `}`,
		`{"result_linkage":` + strings.Replace(linkage, `org.example.production`, strings.Repeat("a", 129), 1) + `}`,
		`{"proxy":` + proxy + `,"result_linkage":{"mode":"bad"}}`,
		`{"proxy":` + proxy + `,"network_context":{"primary_endpoint":"http://bad.example/"}}`,
		`{"proxy":` + proxy + `,"result_linkage":` + linkage + `,"unknown":true}`,
		"\xef\xbb\xbf{}", "{\"proxy\":\"\xff\"}", `{"\ud800":1}`,
		`{"proxy":` + proxy + `,"pr\u006fxy":` + proxy + `}`,
	}
	for i, input := range cases {
		s, err := d.Decode([]byte(input))
		if err == nil || s.Valid() || s != (config.Snapshot{}) {
			t.Fatalf("case %d leaked candidate: %v", i, err)
		}
		if len(err.Error()) > 32 {
			t.Fatalf("case %d unbounded error", i)
		}
	}
}

func TestClosedSchemasAndTrustSeparation(t *testing.T) {
	d := decoder(t)
	for _, key := range []string{"schema_version", "version", "origin", "issuer_id", "policy_generation", "authenticated", "issued_at", "local_policy", "targets", "clock", "linkage_keys", "sourceRef", "installation_secret", "labels", "task_id"} {
		if s, err := d.Decode([]byte(`{"` + key + `":true}`)); err == nil || s.Valid() {
			t.Fatalf("accepted %s", key)
		}
	}
	for _, item := range []struct{ section, body string }{{"proxy", proxy}, {"network_context", network}, {"result_linkage", linkage}} {
		for _, key := range []string{"unknown", "schema_version", "authenticated", "credentials", "installation_secret"} {
			body := strings.TrimSuffix(item.body, "}") + `,"` + key + `":true}`
			if _, err := d.Decode([]byte(`{"` + item.section + `":` + body + `}`)); !errors.Is(err, jsondecode.ErrField) {
				t.Fatalf("closed %s/%s: %v", item.section, key, err)
			}
		}
	}
	var forged config.Snapshot
	if err := json.Unmarshal([]byte(complete), &forged); err != nil {
		t.Fatal(err)
	}
	if forged.Valid() {
		t.Fatal("standard JSON decoding forged a validated snapshot")
	}
	if _, err := config.Prepare(forged, config.Capabilities{Clock: clock{}}); !errors.Is(err, config.ErrUnavailable) {
		t.Fatal("forged initialization")
	}
}

func TestVersionsApprovalsAndBounds(t *testing.T) {
	for _, version := range []string{"", "latest", "0.1", "0.2.0", "0.1.0 "} {
		if d, err := config.NewDecoder(version, approvals()); d != nil || !errors.Is(err, config.ErrVersion) {
			t.Fatal("version fallback")
		}
	}
	d := decoder(t)
	decode(t, d, strings.Repeat(" ", config.MaxBytes-2)+`{}`)
	if s, err := d.Decode([]byte(strings.Repeat(" ", config.MaxBytes-1) + `{}`)); s.Valid() || !errors.Is(err, jsondecode.ErrBounds) {
		t.Fatal("byte bound")
	}
	if _, err := d.Decode([]byte(strings.Repeat(`{"a":`, 10) + `0` + strings.Repeat(`}`, 10))); err == nil {
		t.Fatal("depth")
	}
	for _, port := range []string{"1", "65535"} {
		decode(t, d, `{"proxy":`+strings.Replace(proxy, "54321", port, 1)+`}`)
	}
	for _, rotation := range []string{"86400", "604800"} {
		decode(t, d, `{"result_linkage":`+strings.Replace(linkage, "604800", rotation, 1)+`}`)
	}
	decode(t, d, `{"proxy":`+strings.Replace(proxy, "outline_mobileproxy", strings.Repeat("a", 64), 1)+`}`)
	decode(t, d, `{"result_linkage":`+strings.Replace(linkage, "org.example.production", strings.Repeat("a", 128), 1)+`}`)
	a := approvals()
	a.AllowLinkage = false
	disabled, err := config.NewDecoder(config.Contract, a)
	if err != nil {
		t.Fatal(err)
	}
	if s := decode(t, disabled, complete); s.Linkage() != (config.Linkage{Mode: "none"}) {
		t.Fatal("local policy failed to disable linkage")
	}
	unapproved, _ := config.NewDecoder(config.Contract, config.Approvals{})
	for _, input := range []string{complete, `{"proxy":` + proxy + `}`, `{"network_context":` + network + `}`, `{"result_linkage":` + linkage + `}`} {
		if _, err := unapproved.Decode([]byte(input)); !errors.Is(err, config.ErrApproval) {
			t.Fatalf("approval bypass: %v", err)
		}
	}
	decode(t, unapproved, `{}`)
	for _, a := range []config.Approvals{
		{Endpoints: []string{"http://bad.example/"}}, {ConnectionCodes: []string{"INVALID"}}, {LinkageScopes: []string{""}},
		{Endpoints: make([]string, 65)}, {ConnectionCodes: make([]string, 65)}, {LinkageScopes: make([]string, 65)},
	} {
		if d, err := config.NewDecoder(config.Contract, a); d != nil || !errors.Is(err, config.ErrApproval) {
			t.Fatal("invalid trusted input")
		}
	}
	var nilDecoder *config.Decoder
	if s, err := nilDecoder.Decode([]byte(`{}`)); s.Valid() || !errors.Is(err, config.ErrUnavailable) {
		t.Fatal("nil decoder")
	}
	var zeroDecoder config.Decoder
	if s, err := zeroDecoder.Decode([]byte(`{}`)); s.Valid() || !errors.Is(err, config.ErrUnavailable) {
		t.Fatal("zero decoder bypassed containing-contract selection")
	}
}

func TestSnapshotOwnershipAndReplacement(t *testing.T) {
	a := approvals()
	d, err := config.NewDecoder(config.Contract, a)
	if err != nil {
		t.Fatal(err)
	}
	a.Endpoints[0] = "https://unreviewed.example/"
	a.ConnectionCodes[0] = "unreviewed"
	a.LinkageScopes[0] = "unreviewed"
	input := []byte(complete)
	s, err := d.Decode(input)
	if err != nil {
		t.Fatal(err)
	}
	for i := range input {
		input[i] = ' '
	}
	original := s
	p, _ := s.Proxy()
	p.Host = "192.168.0.1"
	n, _ := s.NetworkContext()
	n.PrimaryEndpoint = "https://bad.example/"
	l := s.Linkage()
	l.Scope = "different"
	if s != original {
		t.Fatal("accessors alias snapshot")
	}
	for _, bad := range []string{`{}`, `null`, strings.Replace(proxy, "54321", "0", 1), strings.TrimSuffix(proxy, "}") + `,"secret":"sensitive"}`} {
		replacement, err := d.WithProxy(s, []byte(bad))
		if err == nil || replacement != original || s != original {
			t.Fatal("partial replacement")
		}
	}
	replaced, err := d.WithProxy(s, []byte(strings.Replace(proxy, "54321", "55001", 1)))
	if err != nil {
		t.Fatal(err)
	}
	if p, _ := replaced.Proxy(); p.Port != 55001 {
		t.Fatal("replacement")
	}
	if s != original {
		t.Fatal("old snapshot mutated")
	}
	if replaced.Linkage() != s.Linkage() {
		t.Fatal("replacement changed linkage")
	}
	if n, ok := replaced.NetworkContext(); !ok || n != nValue(s) {
		t.Fatal("replacement changed endpoint")
	}
	cleared := replaced.WithoutProxy()
	if _, ok := cleared.Proxy(); ok {
		t.Fatal("clear")
	}
	if _, ok := replaced.Proxy(); !ok {
		t.Fatal("clear mutated old snapshot")
	}
	if _, err := d.WithProxy(config.Snapshot{}, []byte(proxy)); !errors.Is(err, config.ErrUnavailable) {
		t.Fatal("zero snapshot")
	}
	var nilDecoder *config.Decoder
	if result, err := nilDecoder.WithProxy(s, []byte(proxy)); result != s || !errors.Is(err, config.ErrUnavailable) {
		t.Fatal("nil replacement")
	}
	var zeroDecoder config.Decoder
	if result, err := zeroDecoder.WithProxy(s, []byte(proxy)); result != s || !errors.Is(err, config.ErrUnavailable) {
		t.Fatal("zero decoder replacement")
	}
	foreign := decoder(t)
	if result, err := foreign.WithProxy(s, []byte(proxy)); result != s || !errors.Is(err, config.ErrUnavailable) {
		t.Fatal("replacement mixed trusted initialization contexts")
	}
}

func TestEndpointApprovalAndLocalLimits(t *testing.T) {
	// A fixed generic query may be reviewed; user-specific endpoint selection
	// cannot be authorized merely by passing URL syntax validation.
	for _, endpoint := range []string{
		"https://context.example/network-context?format=json",
		"https://context.example:443/network-context",
		"https://8.8.8.8/network-context",
		"https://[2606:4700:4700::1111]/network-context",
		"https://context.example/" + strings.Repeat("a", 2048-len("https://context.example/")),
	} {
		d, err := config.NewDecoder(config.Contract, config.Approvals{Endpoints: []string{endpoint}})
		if err != nil {
			t.Fatal("valid reviewed endpoint rejected")
		}
		encoded, _ := json.Marshal(endpoint)
		decode(t, d, `{"network_context":{"primary_endpoint":`+string(encoded)+`}}`)
	}
	for _, endpoint := range []string{
		"https://context.example/" + strings.Repeat("a", 2049-len("https://context.example/")),
		"https://[fe80::1]/", "https://[fc00::1]/", "https://169.254.1.1/", "https://224.0.0.1/", "https://0.0.0.0/",
		"https://context.example:0/", "https://context.example:bad/", "https://context.example:/",
		"https://context.example/#", "https://user@context.example/", "https://sub.localhost/", "https://-bad.example/",
		"https://bad..example/", "https:///missing-host", "https://context.example/\n",
	} {
		if d, err := config.NewDecoder(config.Contract, config.Approvals{Endpoints: []string{endpoint}}); d != nil || !errors.Is(err, config.ErrApproval) {
			t.Fatal("invalid trusted endpoint accepted")
		}
	}
	d := decoder(t)
	if _, err := d.Decode([]byte(`{"network_context":{"primary_endpoint":"https://context.example/network-context","fallback_endpoint":"https://unreviewed.example/"}}`)); !errors.Is(err, config.ErrApproval) {
		t.Fatal("fallback approval bypass")
	}
	members := make([]string, 64)
	for i := range members {
		members[i] = `"` + strings.Repeat("a", i+1) + `":0`
	}
	if _, err := d.Decode([]byte(`{` + strings.Join(members, ",") + `}`)); !errors.Is(err, jsondecode.ErrBounds) {
		t.Fatal("node envelope")
	}
	a := approvals()
	for len(a.Endpoints) < 64 {
		a.Endpoints = append(a.Endpoints, a.Endpoints[0])
	}
	if _, err := config.NewDecoder(config.Contract, a); err != nil {
		t.Fatal("approval count boundary")
	}
}
func nValue(s config.Snapshot) config.NetworkContext { n, _ := s.NetworkContext(); return n }

func TestConcurrentImmutableDecoding(t *testing.T) {
	d := decoder(t)
	expected := decode(t, d, complete)
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 50 {
				s, err := d.Decode([]byte(complete))
				if err != nil || s != expected {
					t.Error("nondeterministic decode")
				}
				if _, err := d.WithProxy(expected, []byte(proxy)); err != nil {
					t.Error(err)
				}
			}
		})
	}
	wg.Wait()
}

type clock struct{}

func (clock) Now(context.Context) (time.Time, error) { return time.Unix(1000, 0), nil }

type nilClock struct{}

func (*nilClock) Now(context.Context) (time.Time, error) { panic("not called during preparation") }

type keys struct{ calls int }

func (k *keys) DeriveHMACSHA256(context.Context, []byte) ([32]byte, error) {
	k.calls++
	return [32]byte{}, nil
}
func (k *keys) Delete(context.Context) error { k.calls++; return nil }

func TestCapabilityPreparation(t *testing.T) {
	s := decode(t, decoder(t), complete)
	for _, c := range []config.Capabilities{{}, {Clock: (*nilClock)(nil)}} {
		if init, err := config.Prepare(s, c); init != nil || !errors.Is(err, config.ErrUnavailable) {
			t.Fatal("missing trusted clock")
		}
	}
	k := &keys{}
	caps := config.Capabilities{Clock: clock{}, LinkageKeys: k}
	i, err := config.Prepare(s, caps)
	if err != nil || i.Configuration() != s {
		t.Fatal("prepare")
	}
	caps.Clock = nil
	caps.LinkageKeys = nil
	if i.Clock() == nil {
		t.Fatal("capability holder aliased")
	}
	if got, ok := i.LinkageKeys(); !ok || got != k {
		t.Fatal("linkage capability")
	}
	if k.calls != 0 {
		t.Fatal("preparation performed key operation")
	}
	for _, store := range []platform.LinkageKeyStore{nil, (*keys)(nil)} {
		i, err := config.Prepare(s, config.Capabilities{Clock: clock{}, LinkageKeys: store})
		if err != nil {
			t.Fatal("linkage-only failure blocked preparation")
		}
		if got, ok := i.LinkageKeys(); ok || got != nil {
			t.Fatal("unavailable protected storage")
		}
	}
	i, err = config.Prepare(decode(t, decoder(t), `{}`), config.Capabilities{Clock: clock{}, LinkageKeys: k})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := i.LinkageKeys(); ok {
		t.Fatal("default enabled linkage")
	}
	if _, err := config.Prepare(config.Snapshot{}, config.Capabilities{Clock: clock{}}); !errors.Is(err, config.ErrUnavailable) {
		t.Fatal("unvalidated config")
	}
}

func FuzzConfiguration(f *testing.F) {
	d := decoder(f)
	for _, seed := range []string{`{}`, complete, `{"proxy":null}`, `{"result_linkage":{"mode":"none"}}`} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, input []byte) {
		a, e1 := d.Decode(input)
		b, e2 := d.Decode(input)
		if !reflect.DeepEqual(a, b) || (e1 == nil) != (e2 == nil) {
			t.Fatal("nondeterministic result")
		}
		if e1 != nil {
			if a != (config.Snapshot{}) || e1.Error() != e2.Error() || len(e1.Error()) > 32 {
				t.Fatal("non-atomic or unstable error")
			}
		} else if !a.Valid() {
			t.Fatal("invalid success")
		}
	})
}
