package task_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/netip"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/task"
)

func stamp() task.Timestamp {
	return task.TimestampFromTime(time.Date(2026, 10, 15, 12, 0, 0, 123, time.UTC))
}
func source() task.SourceInput {
	ts := stamp()
	return task.SourceInput{Origin: task.Remote, IssuerID: "reviewed-issuer", PolicyGeneration: 2, Authenticated: true, IssuedAt: &ts}
}
func policy() task.PolicyInput {
	p := task.Permissions{
		Tests: []task.TestVersion{{ID: "HTTP-01", Version: "0.1.0"}}, RegistrationIDs: []string{"endpoint", "redirect"}, Paths: task.Paths{Direct: true}, MaxRisk: task.LowRisk,
		LabelNamespaces: []string{"campaign"},
		Limits: task.PolicyLimits{Attempt: task.AttemptLimits{DurationMS: 30000, BytesSent: 32768, BytesReceived: 131072, Connections: 8}, MaxRuns: 1, MaxAttempts: 1, MaxValiditySeconds: 86400,
			MinRepeatIntervalSeconds: 21600, MinRepeatJitterSeconds: 300, MinRetryDelayMS: 30000, MaxRetryDelayMS: 30000, MaxRetryJitterMS: 10000,
			MaxResultBytes: 65536, MaxBundleBytes: 262144, MaxResultsPerBundle: 16, MaxPendingBundles: 32, MaxPendingBytes: 4194304, MaxRetentionSeconds: 86400},
	}
	t := task.TargetRegistration{ID: "endpoint", Test: p.Tests[0], Role: task.MeasurementTarget, Paths: p.Paths, BaselineRisk: task.LowRisk, Owner: "synthetic-owner", Purpose: "synthetic-test", ReviewGeneration: 2, ExpiresAt: stamp(), RedirectIDs: []string{"redirect"},
		Tuple: task.TargetTuple{Hostname: "example.org", URL: "https://example.org/check", Port: 443, Scheme: "https", TLSServerName: "example.org", BootstrapAddresses: []netip.Addr{netip.MustParseAddr("192.0.2.1")}, ALPN: []string{"http/1.1"}, ExpectedStatus: 200, BodyChecks: []task.BodyCheck{{Type: "prefix", Value: "ok"}}, HeaderChecks: []task.HeaderCheck{{Name: "content-type", Type: "contains", Value: "text/plain"}}}}
	r := t
	r.ID = "redirect"
	r.Role = task.RedirectTarget
	r.RedirectIDs = nil
	return task.PolicyInput{Generation: 2, Permissions: p, Issuers: []task.IssuerPermissions{{IssuerID: "reviewed-issuer", Permissions: p}}, Targets: []task.TargetRegistration{t, r}}
}

func TestSourceContextInvariantsAndByteBinding(t *testing.T) {
	in := source()
	s, err := task.NewSourceContext(in)
	if err != nil || !s.Valid() || s.Origin() != task.Remote || !s.Authenticated() || s.IssuerID() != in.IssuerID || s.PolicyGeneration() != 2 {
		t.Fatal("valid source rejected")
	}
	want := *in.IssuedAt
	in.IssuedAt.Fraction = "9"
	got, ok := s.IssuedAt()
	if !ok || got != want {
		t.Fatal("source retains caller pointer")
	}
	r, err := task.NewDeliveryRecord(s, []byte(`{"a":1}`))
	if err != nil || !r.Matches([]byte(`{"a":1}`)) || r.Matches([]byte(`{ "a":1}`)) || r.Matches(nil) {
		t.Fatal("exact byte identity lost")
	}
	buf := []byte(`{"a":1}`)
	r, err = task.NewDeliveryRecord(s, buf)
	buf[0] = ' '
	if err != nil || r.Matches(buf) || !r.Matches([]byte(`{"a":1}`)) {
		t.Fatal("payload mutation changed record")
	}
	for _, b := range [][]byte{nil, bytes.Repeat([]byte{'x'}, task.MaxDeliveryBytes+1)} {
		if r, e := task.NewDeliveryRecord(s, b); e != task.ErrTrustModel || r.Source().Valid() {
			t.Fatal("invalid byte envelope accepted")
		}
	}
	// Records bind over-limit tasks so admission can reject them by size.
	for _, n := range []int{65536, 65537, task.MaxDeliveryBytes} {
		if r, e := task.NewDeliveryRecord(s, bytes.Repeat([]byte{'x'}, n)); e != nil || !r.Matches(bytes.Repeat([]byte{'x'}, n)) {
			t.Fatal("receive byte bound rejected")
		}
	}
	if r, e := task.NewDeliveryRecord(task.SourceContext{}, []byte("x")); e != task.ErrTrustModel || r.Matches([]byte("x")) {
		t.Fatal("zero source bound")
	}
	embedded, e := task.NewSourceContext(task.SourceInput{Origin: task.Embedded, IssuerID: "release", PolicyGeneration: 1})
	if e != nil || !embedded.Valid() || embedded.Authenticated() {
		t.Fatal("reviewed embedded shape rejected")
	}
	if _, has := embedded.IssuedAt(); has {
		t.Fatal("embedded issuance invented")
	}
}

// Structural evidence for trust separation; actual provenance/time availability
// and fail-closed admission remain with the future admission owner.
// requirement-ref: REQ-TRUST-0001
func TestMalformedSourceRejectedAtomically(t *testing.T) {
	cases := map[string]func(*task.SourceInput){
		"origin": func(s *task.SourceInput) { s.Origin = 0 }, "generation": func(s *task.SourceInput) { s.PolicyGeneration = 0 },
		"issuer empty": func(s *task.SourceInput) { s.IssuerID = "" }, "issuer case": func(s *task.SourceInput) { s.IssuerID = "Issuer" }, "issuer unicode": func(s *task.SourceInput) { s.IssuerID = "é" }, "issuer long": func(s *task.SourceInput) { s.IssuerID = strings.Repeat("a", 65) },
		"authentication": func(s *task.SourceInput) { s.Authenticated = false }, "missing time": func(s *task.SourceInput) { s.IssuedAt = nil }, "zero time": func(s *task.SourceInput) { s.IssuedAt = &task.Timestamp{} },
		"fraction": func(s *task.SourceInput) { s.IssuedAt.Fraction = "01x" }, "trailing zero": func(s *task.SourceInput) { s.IssuedAt.Fraction = "10" }, "fraction bound": func(s *task.SourceInput) { s.IssuedAt.Fraction = strings.Repeat("1", 65537) },
		"whole fraction": func(s *task.SourceInput) { s.IssuedAt.Whole = s.IssuedAt.Whole.Add(time.Nanosecond) }, "not UTC": func(s *task.SourceInput) { s.IssuedAt.Whole = s.IssuedAt.Whole.In(time.FixedZone("offset", 3600)) },
		"embedded auth": func(s *task.SourceInput) { s.Origin = task.Embedded; s.IssuedAt = nil }, "embedded time": func(s *task.SourceInput) { s.Origin = task.Embedded; s.Authenticated = false },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			in := source()
			mutate(&in)
			s, e := task.NewSourceContext(in)
			if e != task.ErrTrustModel || s.Valid() {
				t.Fatal("invalid source escaped")
			}
		})
	}
	in := source()
	in.IssuerID = strings.Repeat("a", 64)
	in.PolicyGeneration = ^uint64(0)
	if s, e := task.NewSourceContext(in); e != nil || !s.Valid() {
		t.Fatal("exact positive generation/identifier bounds rejected")
	}
}

func TestPolicyRejectsMalformedOrAmbiguousAuthority(t *testing.T) {
	cases := map[string]func(*task.PolicyInput){
		"generation": func(p *task.PolicyInput) { p.Generation = 0 }, "missing review": func(p *task.PolicyInput) { p.Targets[0].ReviewGeneration = 0 }, "expiry": func(p *task.PolicyInput) { p.Targets[0].ExpiresAt = task.Timestamp{} },
		"no issuers": func(p *task.PolicyInput) { p.Issuers = nil }, "issuer duplicate": func(p *task.PolicyInput) { p.Issuers = append(p.Issuers, p.Issuers[0]) }, "issuer invalid": func(p *task.PolicyInput) { p.Issuers[0].IssuerID = "*" },
		"issuer unknown registration": func(p *task.PolicyInput) { p.Issuers[0].Permissions.RegistrationIDs = []string{"missing"} }, "local unknown registration": func(p *task.PolicyInput) { p.Permissions.RegistrationIDs = []string{"missing"} },
		"duplicate target": func(p *task.PolicyInput) { p.Targets = append(p.Targets, p.Targets[0]) }, "unknown role": func(p *task.PolicyInput) { p.Targets[0].Role = 0 }, "target risk": func(p *task.PolicyInput) { p.Targets[0].BaselineRisk = 0 },
		"no paths": func(p *task.PolicyInput) { p.Permissions.Paths = task.Paths{} }, "target no paths": func(p *task.PolicyInput) { p.Targets[0].Paths = task.Paths{} }, "risk": func(p *task.PolicyInput) { p.Permissions.MaxRisk = task.UnknownRisk }, "medium inconsistent": func(p *task.PolicyInput) { p.Permissions.AllowMediumRisk = true },
		"duplicate version": func(p *task.PolicyInput) { p.Permissions.Tests = append(p.Permissions.Tests, p.Permissions.Tests[0]) }, "unversioned": func(p *task.PolicyInput) { p.Permissions.Tests[0].Version = "latest" }, "wrong pair": func(p *task.PolicyInput) { p.Targets[0].Test.Version = "0.2.0" },
		"composite disabled": func(p *task.PolicyInput) {
			p.Permissions.Tests = append(p.Permissions.Tests, task.TestVersion{ID: "NET-05", Version: "0.1.0"})
		}, "duplicate labels": func(p *task.PolicyInput) { p.Permissions.LabelNamespaces = []string{"campaign", "campaign"} }, "wildcard label": func(p *task.PolicyInput) { p.Permissions.LabelNamespaces = []string{"*"} },
		"empty limits": func(p *task.PolicyInput) { p.Permissions.Limits = task.PolicyLimits{} }, "duration": func(p *task.PolicyInput) { p.Permissions.Limits.Attempt.DurationMS = 60001 }, "bytes": func(p *task.PolicyInput) { p.Permissions.Limits.Attempt.BytesReceived = 262145 }, "negative": func(p *task.PolicyInput) { p.Permissions.Limits.Attempt.BytesSent = -1 },
		"runs": func(p *task.PolicyInput) { p.Permissions.Limits.MaxRuns = 29 }, "attempts": func(p *task.PolicyInput) { p.Permissions.Limits.MaxAttempts = 3 }, "validity": func(p *task.PolicyInput) { p.Permissions.Limits.MaxValiditySeconds = 604801 }, "cadence": func(p *task.PolicyInput) { p.Permissions.Limits.MinRepeatIntervalSeconds = 21599 }, "jitter": func(p *task.PolicyInput) { p.Permissions.Limits.MinRepeatJitterSeconds = 299 },
		"retry delay": func(p *task.PolicyInput) { p.Permissions.Limits.MaxRetryDelayMS = 29999 }, "retention": func(p *task.PolicyInput) { p.Permissions.Limits.MaxRetentionSeconds = 86401 }, "result": func(p *task.PolicyInput) { p.Permissions.Limits.MaxResultBytes = 65537 },
		"too many": func(p *task.PolicyInput) { p.Targets = make([]task.TargetRegistration, 257) }, "redirect missing": func(p *task.PolicyInput) { p.Targets[0].RedirectIDs = []string{"missing"} }, "redirect self": func(p *task.PolicyInput) { p.Targets[0].RedirectIDs = []string{"endpoint"} }, "redirect role": func(p *task.PolicyInput) { p.Targets[1].Role = task.MeasurementTarget },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			p := policy()
			mutate(&p)
			s, e := task.NewLocalPolicy(p)
			if e != task.ErrTrustModel || s.Valid() || s.Generation() != 0 {
				t.Fatal("invalid policy escaped")
			}
		})
	}
}

func TestRegistrationTuplesAreExactAndBounded(t *testing.T) {
	cases := map[string]func(*task.TargetTuple){
		"empty": func(v *task.TargetTuple) { *v = task.TargetTuple{} }, "wildcard": func(v *task.TargetTuple) { v.Hostname = "*.example.org" }, "noncanonical host": func(v *task.TargetTuple) { v.Hostname = "EXAMPLE.org" }, "trailing dot": func(v *task.TargetTuple) { v.Hostname = "example.org." },
		"credentials": func(v *task.TargetTuple) { v.URL = "https://secret:password@example.org/" }, "fragment": func(v *task.TargetTuple) { v.URL += "#private" }, "scheme": func(v *task.TargetTuple) { v.URL = "http://example.org/" }, "port conflict": func(v *task.TargetTuple) { v.Port = 8443 }, "bad port": func(v *task.TargetTuple) { v.URL = "https://example.org:65536/" }, "host conflict": func(v *task.TargetTuple) { v.Hostname = "other.example.org" },
		"interpolation": func(v *task.TargetTuple) { v.URL = "https://example.org/{device}" }, "address zone": func(v *task.TargetTuple) { v.Address = netip.MustParseAddr("fe80::1%en0") }, "mapped address": func(v *task.TargetTuple) { v.Address = netip.MustParseAddr("::ffff:192.0.2.1") }, "invalid bootstrap": func(v *task.TargetTuple) { v.BootstrapAddresses = []netip.Addr{{}} },
		"duplicate alpn": func(v *task.TargetTuple) { v.ALPN = []string{"h2", "h2"} }, "ALPN bound": func(v *task.TargetTuple) { v.ALPN = []string{strings.Repeat("x", 33)} }, "status": func(v *task.TargetTuple) { v.ExpectedStatus = 600 },
		"unknown check": func(v *task.TargetTuple) { v.BodyChecks = []task.BodyCheck{{Type: "regex", Value: ".*"}} }, "bad hash": func(v *task.TargetTuple) { v.BodyChecks = []task.BodyCheck{{Type: "sha256", Value: "x"}} }, "check bound": func(v *task.TargetTuple) {
			v.BodyChecks = []task.BodyCheck{{Type: "contains", Value: strings.Repeat("x", 257)}}
		},
		"sensitive header": func(v *task.TargetTuple) {
			v.HeaderChecks = []task.HeaderCheck{{Name: "set-cookie", Type: "equals", Value: "secret"}}
		}, "presence value": func(v *task.TargetTuple) {
			v.HeaderChecks = []task.HeaderCheck{{Name: "content-type", Type: "present", Value: "text"}}
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			p := policy()
			mutate(&p.Targets[0].Tuple)
			s, e := task.NewLocalPolicy(p)
			if e != task.ErrTrustModel || s.Valid() {
				t.Fatal("invalid tuple escaped")
			}
		})
	}
	// Full HTTPS tuples are not mandatory for a different registered test. No
	// constructor invents absent properties or registers a production runner.
	p := policy()
	p.Targets[0].Tuple = task.TargetTuple{Address: netip.MustParseAddr("192.0.2.1"), Port: 443, Scheme: "tcp"}
	if _, e := task.NewLocalPolicy(p); e != nil {
		t.Fatal("typed partial tuple shape rejected")
	}
}

func TestPolicySnapshotsSurviveMutationAndConcurrentReads(t *testing.T) {
	in := policy()
	p, e := task.NewLocalPolicy(in)
	if e != nil || !p.Valid() || p.Generation() != 2 {
		t.Fatal("valid policy rejected")
	}
	want := policy()
	in.Generation = 7
	in.Permissions.Tests[0].Version = "9.9.9"
	in.Permissions.RegistrationIDs[0] = "changed"
	in.Permissions.LabelNamespaces[0] = "changed"
	in.Issuers[0].IssuerID = "changed"
	in.Issuers[0].Permissions.Paths.Proxy = true
	in.Targets[0].Tuple.ALPN[0] = "changed"
	in.Targets[0].Tuple.BodyChecks[0].Value = "changed"
	in.Targets[0].Tuple.HeaderChecks[0].Value = "changed"
	in.Targets[0].Tuple.BootstrapAddresses[0] = netip.Addr{}
	in.Targets[0].RedirectIDs[0] = "changed"
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 30 {
				got := p.Permissions()
				if !reflect.DeepEqual(got, want.Permissions) {
					t.Error("local policy changed")
				}
				got.Tests[0].ID = "changed"
				got.RegistrationIDs[0] = "changed"
				got.LabelNamespaces[0] = "changed"
				i, ok := p.Issuer("reviewed-issuer")
				if !ok || !reflect.DeepEqual(i, want.Issuers[0]) {
					t.Error("issuer changed")
				}
				i.Permissions.Tests[0].Version = "changed"
				i.Permissions.RegistrationIDs[0] = "changed"
				i.Permissions.LabelNamespaces[0] = "changed"
				r, ok := p.Target("endpoint")
				if !ok || !reflect.DeepEqual(r, want.Targets[0]) {
					t.Error("target changed")
				}
				r.Tuple.ALPN[0] = "changed"
				r.Tuple.BodyChecks[0].Value = "changed"
				r.Tuple.HeaderChecks[0].Value = "changed"
				r.Tuple.BootstrapAddresses[0] = netip.Addr{}
				r.RedirectIDs[0] = "changed"
			}
		})
	}
	wg.Wait()
	if _, ok := p.Issuer("reviewed"); ok {
		t.Fatal("issuer prefix matched")
	}
	if _, ok := p.Target("endpoint.evil"); ok {
		t.Fatal("target suffix matched")
	}
}

func TestPolicyKeepsRestrictionsSeparateWithoutGrantingAuthority(t *testing.T) {
	in := policy()
	// Target review generations are separate metadata. The public contract does
	// not establish an ordering between them and the containing policy counter.
	in.Targets[0].ReviewGeneration = 9
	in.Permissions.RequireConsent = true
	in.Permissions.Limits.MaxRuns = 1
	in.Issuers[0].Permissions.MaxRisk = task.HighRisk
	in.Issuers[0].Permissions.AllowUnknownRisk = true
	in.Issuers[0].Permissions.AllowMediumRisk = true
	in.Issuers[0].Permissions.Paths.Proxy = true
	in.Issuers[0].Permissions.Limits.MaxRuns = 28
	p, e := task.NewLocalPolicy(in)
	if e != nil {
		t.Fatal(e)
	}
	i, ok := p.Issuer("reviewed-issuer")
	if !ok || i.Permissions.Limits.MaxRuns != 28 || !i.Permissions.AllowUnknownRisk {
		t.Fatal("issuer input lost")
	}
	local := p.Permissions()
	if !local.RequireConsent || local.Limits.MaxRuns != 1 || local.Paths.Proxy || local.AllowUnknownRisk || local.MaxRisk != task.LowRisk {
		t.Fatal("issuer widened local policy")
	}
	// A local-only test can carry no destination registrations and zero network
	// budgets. The model does not silently fill registered network defaults.
	in = policy()
	in.Targets = nil
	in.Permissions.Tests = []task.TestVersion{{ID: "NET-01", Version: "0.1.0"}}
	in.Permissions.RegistrationIDs = nil
	in.Permissions.Limits.Attempt = task.AttemptLimits{DurationMS: 1000}
	in.Issuers[0].Permissions = in.Permissions
	p, e = task.NewLocalPolicy(in)
	if e != nil || p.Permissions().Limits.Attempt.Connections != 0 {
		t.Fatal("local-only policy widened")
	}
}

func TestRevocationInputsHaveNoActivationSideEffects(t *testing.T) {
	in := task.RevocationInput{PolicyGeneration: 3, IssuerIDs: []string{"reviewed-issuer"}, RegistrationIDs: []string{"endpoint"}}
	r, e := task.NewRevocations(in)
	if e != nil || !r.Valid() {
		t.Fatal("valid revocation rejected")
	}
	in.IssuerIDs[0] = "changed"
	in.RegistrationIDs[0] = "changed"
	got := r.Input()
	got.IssuerIDs[0] = "again"
	got.RegistrationIDs[0] = "again"
	if r.Input().IssuerIDs[0] != "reviewed-issuer" || r.Input().RegistrationIDs[0] != "endpoint" {
		t.Fatal("revocation aliased")
	}
	for _, in := range []task.RevocationInput{{}, {PolicyGeneration: 3}, {PolicyGeneration: 3, IssuerIDs: []string{"issuer", "issuer"}}, {PolicyGeneration: 3, RegistrationIDs: []string{"*"}}, {PolicyGeneration: 3, IssuerIDs: make([]string, 257)}} {
		if r, e := task.NewRevocations(in); e != task.ErrTrustModel || r.Valid() {
			t.Fatal("invalid revocation escaped")
		}
	}
	p, e := task.NewLocalPolicy(policy())
	if e != nil {
		t.Fatal(e)
	}
	if _, ok := p.Target("endpoint"); !ok {
		t.Fatal("model construction activated revocation")
	}
}

func TestTrustModelsCannotBeImportedOrExportedAsJSON(t *testing.T) {
	s, e := task.NewSourceContext(source())
	if e != nil {
		t.Fatal(e)
	}
	p, e := task.NewLocalPolicy(policy())
	if e != nil {
		t.Fatal(e)
	}
	r, e := task.NewDeliveryRecord(s, []byte("exact"))
	if e != nil {
		t.Fatal(e)
	}
	rev, e := task.NewRevocations(task.RevocationInput{PolicyGeneration: 3, IssuerIDs: []string{"issuer"}})
	if e != nil {
		t.Fatal(e)
	}
	for _, v := range []any{source(), s, r, policy(), p, policy().Permissions, policy().Issuers[0], policy().Targets[0], policy().Targets[0].Tuple, rev, rev.Input()} {
		if b, e := json.Marshal(v); !errors.Is(e, task.ErrTrustEncoding) || len(b) != 0 {
			t.Fatal("trust material exported")
		}
		ptr := reflect.New(reflect.TypeOf(v))
		ptr.Elem().Set(reflect.ValueOf(v))
		before := ptr.Elem().Interface()
		for _, payload := range []string{`{}`, `null`, `{"authenticated":true,"policy_generation":999}`} {
			if e := json.Unmarshal([]byte(payload), ptr.Interface()); !errors.Is(e, task.ErrTrustEncoding) || !reflect.DeepEqual(before, ptr.Elem().Interface()) {
				t.Fatal("JSON altered trust state")
			}
		}
	}
}
