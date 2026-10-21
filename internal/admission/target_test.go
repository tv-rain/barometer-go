package admission_test

import (
	"errors"
	"maps"
	"net/netip"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/admission"
	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/registry"
	"github.com/tv-rain/barometer-go/internal/taskdecode"
)

// Synthetic test-scoped codecs and target rules under the registered TCP-01,
// TLS-01, HTTP-01 and DNS-01 IDs. Their input shape and normalization are local
// fixtures, not the public test schemas or production codecs (DEV-454): each
// decodes literal string fields into one typed tuple, lowercases names and
// writes IP addresses in netip canonical form, keeping IPv4-mapped IPv6 mapped.
type tupleParameters struct {
	Tuple     task.TargetTuple
	Redirects []string
}

func (tupleParameters) TaskParameters() {}

type tupleCodec struct{ id string }

func (c tupleCodec) Definition() taskdecode.Definition {
	text := &jsondecode.Schema{Kind: jsondecode.String, MaxBytes: 2048}
	list := &jsondecode.Schema{Kind: jsondecode.Array, Element: text, MaxCount: 16}
	fields := map[string]jsondecode.Field{}
	for _, name := range []string{"host", "address", "port", "scheme", "url", "resolver", "server_name", "status"} {
		fields[name] = jsondecode.Field{Schema: text}
	}
	for _, name := range []string{"bootstrap", "alpn", "body", "header", "redirects"} {
		fields[name] = jsondecode.Field{Schema: list}
	}
	return taskdecode.Definition{TestID: c.id, TestVersion: "0.1.0",
		Input:   &jsondecode.Schema{Kind: jsondecode.Object, Fields: fields},
		Options: &jsondecode.Schema{Kind: jsondecode.Object, Fields: map[string]jsondecode.Field{}}}
}

var errFixture = errors.New("fixture normalization")

func (tupleCodec) Normalize(input, _ *jsondecode.Value) (task.Parameters, error) {
	text := func(name string) string {
		if v, ok := input.Lookup(name); ok {
			return v.Text()
		}
		return ""
	}
	texts := func(name string) []string {
		v, ok := input.Lookup(name)
		if !ok {
			return nil
		}
		var out []string
		for _, e := range v.Elements() {
			out = append(out, e.Text())
		}
		return out
	}
	number := func(name string, bits int) (uint64, error) {
		if s := text(name); s != "" {
			return strconv.ParseUint(s, 10, bits)
		}
		return 0, nil
	}
	var t task.TargetTuple
	var err error
	t.Hostname, t.ResolverName, t.TLSServerName = strings.ToLower(text("host")), strings.ToLower(text("resolver")), strings.ToLower(text("server_name"))
	t.Scheme, t.ALPN = text("scheme"), texts("alpn")
	if s := text("address"); s != "" {
		if t.Address, err = netip.ParseAddr(s); err != nil {
			return nil, errFixture
		}
	}
	for _, s := range texts("bootstrap") {
		a, err := netip.ParseAddr(s)
		if err != nil {
			return nil, errFixture
		}
		t.BootstrapAddresses = append(t.BootstrapAddresses, a)
	}
	port, err := number("port", 16)
	status, err2 := number("status", 16)
	if err != nil || err2 != nil {
		return nil, errFixture
	}
	t.Port, t.ExpectedStatus = uint16(port), int(status)
	for _, s := range texts("body") {
		kind, value, _ := strings.Cut(s, ":")
		t.BodyChecks = append(t.BodyChecks, task.BodyCheck{Type: kind, Value: value})
	}
	for _, s := range texts("header") {
		parts := strings.SplitN(s, ":", 3)
		for len(parts) < 3 {
			parts = append(parts, "")
		}
		t.HeaderChecks = append(t.HeaderChecks, task.HeaderCheck{Name: parts[0], Type: parts[1], Value: parts[2]})
	}
	if t.URL, err = normalURL(text("url")); err != nil {
		return nil, errFixture
	}
	p := tupleParameters{Tuple: t}
	for _, s := range texts("redirects") {
		u, err := normalURL(s)
		if err != nil || u == "" {
			return nil, errFixture
		}
		p.Redirects = append(p.Redirects, u)
	}
	return p, nil
}

// normalURL lowercases the scheme and host and keeps every other component.
func normalURL(s string) (string, error) {
	if s == "" {
		return "", nil
	}
	u, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	u.Host = strings.ToLower(u.Host)
	return u.String(), nil
}

// tupleRule reports the complete tuple each synthetic test requires and
// reports false when any required component is absent.
type tupleRule struct{ id string }

func (r tupleRule) Targets(p task.Parameters) ([]registry.Target, bool) {
	tp, ok := p.(tupleParameters)
	if !ok {
		return nil, false
	}
	t := tp.Tuple
	var complete bool
	switch r.id {
	case "TCP-01":
		complete = t.Address.IsValid() && t.Port != 0 && t.Scheme != ""
	case "TLS-01":
		complete = t.Address.IsValid() && t.Port != 0 && t.Scheme != "" && t.TLSServerName != "" && len(t.ALPN) > 0
	case "HTTP-01":
		complete = t.URL != "" && t.ExpectedStatus != 0
	case "DNS-01":
		complete = t.Hostname != "" && t.ResolverName != "" && len(t.BootstrapAddresses) > 0
	}
	if !complete || len(tp.Redirects) > 0 && r.id != "HTTP-01" {
		return nil, false
	}
	out := []registry.Target{{Role: task.MeasurementTarget, Tuple: t}}
	for _, u := range tp.Redirects {
		out = append(out, registry.Target{Role: task.RedirectTarget, Tuple: task.TargetTuple{URL: u}})
	}
	return out, true
}

// tupleProxy lets every synthetic tuple through the proxy, so path cases reach
// target registration.
var tupleProxy = proxyRule(func(p task.Parameters) bool {
	_, ok := p.(tupleParameters)
	return ok
})

var tupleTests = []string{"TCP-01", "TLS-01", "HTTP-01", "DNS-01"}

func tupleEntries() []registry.Entry {
	var out []registry.Entry
	for _, id := range tupleTests {
		out = append(out, registry.Entry{Test: task.TestVersion{ID: id, Version: "0.1.0"}, Codec: tupleCodec{id: id}, BaselineRisk: task.LowRisk, Targets: tupleRule{id: id}, Proxy: tupleProxy})
	}
	return out
}

func addr(s string) netip.Addr { return netip.MustParseAddr(s) }

func tupleRegistration(id, test string, role task.TargetRole, tuple task.TargetTuple) task.TargetRegistration {
	r := registration(id)
	r.Test, r.Role, r.Tuple, r.Paths = task.TestVersion{ID: test, Version: "0.1.0"}, role, tuple, task.Paths{Direct: true}
	return r
}

// tupleRegistrations are the synthetic registrations matched by tupleInputs.
func tupleRegistrations() []task.TargetRegistration {
	http := tupleRegistration("http-status", "HTTP-01", task.MeasurementTarget, task.TargetTuple{URL: "https://www.example.test/status", ExpectedStatus: 204,
		BodyChecks: []task.BodyCheck{{Type: "contains", Value: "ok"}}, HeaderChecks: []task.HeaderCheck{{Name: "content-type", Type: "equals", Value: "text/plain"}}})
	http.RedirectIDs = []string{"http-redirect"}
	return []task.TargetRegistration{
		tupleRegistration("tcp-endpoint", "TCP-01", task.MeasurementTarget, task.TargetTuple{Address: addr("192.0.2.10"), Port: 443, Scheme: "tcp"}),
		tupleRegistration("tls-endpoint", "TLS-01", task.MeasurementTarget, task.TargetTuple{Address: addr("192.0.2.20"), Port: 443, Scheme: "tls", TLSServerName: "tls.example.test", ALPN: []string{"h2", "http/1.1"}}),
		http,
		tupleRegistration("http-redirect", "HTTP-01", task.RedirectTarget, task.TargetTuple{URL: "https://cdn.example.test/status"}),
		tupleRegistration("dns-name", "DNS-01", task.MeasurementTarget, task.TargetTuple{Hostname: "example.test", ResolverName: "resolver.example.test", BootstrapAddresses: []netip.Addr{addr("192.0.2.53"), addr("2001:db8::53")}}),
	}
}

func tupleInputs() map[string]map[string]any {
	return map[string]map[string]any{
		"TCP-01":  {"address": "192.0.2.10", "port": "443", "scheme": "tcp"},
		"TLS-01":  {"address": "192.0.2.20", "port": "443", "scheme": "tls", "server_name": "tls.example.test", "alpn": []any{"h2", "http/1.1"}},
		"HTTP-01": {"url": "https://www.example.test/status", "status": "204", "body": []any{"contains:ok"}, "header": []any{"content-type:equals:text/plain"}, "redirects": []any{"https://cdn.example.test/status"}},
		"DNS-01":  {"host": "example.test", "resolver": "resolver.example.test", "bootstrap": []any{"192.0.2.53", "2001:db8::53"}},
	}
}

// tuplePolicy permits the four synthetic tests and every registration on both
// sides, on both paths, so only target registration can reject.
func tuplePolicy(edit func(*task.PolicyInput)) task.PolicyInput {
	in := permissive()
	in.Targets = tupleRegistrations()
	for _, p := range []*task.Permissions{&in.Permissions, &in.Issuers[0].Permissions, &in.Issuers[1].Permissions} {
		p.Tests, p.RegistrationIDs = nil, nil
		for _, id := range tupleTests {
			p.Tests = append(p.Tests, task.TestVersion{ID: id, Version: "0.1.0"})
		}
		for _, r := range in.Targets {
			p.RegistrationIDs = append(p.RegistrationIDs, r.ID)
		}
	}
	if edit != nil {
		edit(&in)
	}
	return in
}

func tupleHarness(t *testing.T, edit func(*task.PolicyInput)) *harness {
	t.Helper()
	h := newHarness(t)
	rebind(t, h, mustPolicy(t, tuplePolicy(edit)), tupleEntries()...)
	return h
}

// tupleTask builds a task for test with its registered input, edited per key;
// a nil value deletes the key.
func tupleTask(test string, edits map[string]any, extra map[string]any) []byte {
	input := tupleInputs()[test]
	for k, v := range edits {
		if v == nil {
			delete(input, k)
		} else {
			input[k] = v
		}
	}
	all := map[string]any{"test_id": test, "input": input}
	maps.Copy(all, extra)
	return taskWith(all)
}

// Every synthetic tuple that exactly matches its registration passes step 8,
// and the carried registration IDs narrow to those the task matched.
// requirement-ref: REQ-TRUST-0013
// requirement-ref: REQ-TRUST-0014
func TestGateTargetRegistrationExactMatch(t *testing.T) {
	h := tupleHarness(t, nil)
	want := map[string][]string{"TCP-01": {"tcp-endpoint"}, "TLS-01": {"tls-endpoint"}, "HTTP-01": {"http-status", "http-redirect"}, "DNS-01": {"dns-name"}}
	for _, test := range tupleTests {
		g, err := h.embedded(t, tupleTask(test, nil, nil))
		if err != nil {
			t.Errorf("%s: exact tuple rejected as %v", test, err)
			continue
		}
		if got := admission.GatedRegistrations(g); !reflect.DeepEqual(got, want[test]) {
			t.Errorf("%s: carried registrations %v, want %v", test, got, want[test])
		}
	}
	// A redirect is optional; an HTTP-01 task without one matches only its
	// measurement registration.
	g, err := h.embedded(t, tupleTask("HTTP-01", map[string]any{"redirects": nil}, nil))
	if err != nil || !reflect.DeepEqual(admission.GatedRegistrations(g), []string{"http-status"}) {
		t.Errorf("HTTP-01 without redirect: %v", err)
	}
}

// Changing one tuple component at a time, the path or a redirect target is
// target_not_registered with no network activity.
// requirement-ref: REQ-TRUST-0014
// requirement-ref: REQ-TRUST-0015
func TestGateTargetRegistrationSingleComponentMutations(t *testing.T) {
	h := tupleHarness(t, nil)
	proxy := map[string]any{"execution.network_path": "proxy"}
	cases := []struct {
		name, test string
		edits      map[string]any
		extra      map[string]any
	}{
		{"IP address", "TCP-01", map[string]any{"address": "192.0.2.11"}, nil},
		{"IP address family", "TCP-01", map[string]any{"address": "2001:db8::10"}, nil},
		{"port", "TCP-01", map[string]any{"port": "444"}, nil},
		{"scheme", "TCP-01", map[string]any{"scheme": "udp"}, nil},
		{"added hostname", "TCP-01", map[string]any{"host": "example.test"}, nil},
		{"TCP path", "TCP-01", nil, proxy},
		{"TLS address", "TLS-01", map[string]any{"address": "192.0.2.21"}, nil},
		{"TLS port", "TLS-01", map[string]any{"port": "8443"}, nil},
		{"server name", "TLS-01", map[string]any{"server_name": "other.example.test"}, nil},
		{"ALPN value", "TLS-01", map[string]any{"alpn": []any{"h2", "h3"}}, nil},
		{"ALPN dropped", "TLS-01", map[string]any{"alpn": []any{"h2"}}, nil},
		{"ALPN added", "TLS-01", map[string]any{"alpn": []any{"h2", "http/1.1", "h3"}}, nil},
		{"ALPN order", "TLS-01", map[string]any{"alpn": []any{"http/1.1", "h2"}}, nil},
		{"TLS path", "TLS-01", nil, proxy},
		{"URL host", "HTTP-01", map[string]any{"url": "https://api.example.test/status"}, nil},
		{"URL path", "HTTP-01", map[string]any{"url": "https://www.example.test/other"}, nil},
		{"URL port", "HTTP-01", map[string]any{"url": "https://www.example.test:8443/status"}, nil},
		{"URL query", "HTTP-01", map[string]any{"url": "https://www.example.test/status?x=1"}, nil},
		{"expected status", "HTTP-01", map[string]any{"status": "200"}, nil},
		{"body check value", "HTTP-01", map[string]any{"body": []any{"contains:okay"}}, nil},
		{"body check type", "HTTP-01", map[string]any{"body": []any{"prefix:ok"}}, nil},
		{"body check dropped", "HTTP-01", map[string]any{"body": nil}, nil},
		{"header check value", "HTTP-01", map[string]any{"header": []any{"content-type:equals:text/html"}}, nil},
		{"header check name", "HTTP-01", map[string]any{"header": []any{"content-length:equals:text/plain"}}, nil},
		{"redirect target", "HTTP-01", map[string]any{"redirects": []any{"https://cdn.example.test/other"}}, nil},
		{"extra redirect", "HTTP-01", map[string]any{"redirects": []any{"https://cdn.example.test/status", "https://evil.example.test/"}}, nil},
		{"HTTP path", "HTTP-01", nil, proxy},
		{"queried name", "DNS-01", map[string]any{"host": "www.example.test"}, nil},
		{"resolver name", "DNS-01", map[string]any{"resolver": "other-resolver.example.test"}, nil},
		{"bootstrap address", "DNS-01", map[string]any{"bootstrap": []any{"192.0.2.54", "2001:db8::53"}}, nil},
		{"bootstrap dropped", "DNS-01", map[string]any{"bootstrap": []any{"192.0.2.53"}}, nil},
		{"bootstrap order", "DNS-01", map[string]any{"bootstrap": []any{"2001:db8::53", "192.0.2.53"}}, nil},
		{"DNS path", "DNS-01", nil, proxy},
	}
	for _, c := range cases {
		_, err := h.embedded(t, tupleTask(c.test, c.edits, c.extra))
		expect(t, c.test+" "+c.name, err, task.TargetNotRegistered)
	}
}

// Forms the synthetic codec normalizes to a registered value match; suffix,
// subdomain, path-prefix, trailing-dot and IPv4-mapped forms do not.
// requirement-ref: REQ-TRUST-0015
func TestGateTargetRegistrationNormalization(t *testing.T) {
	h := tupleHarness(t, nil)
	cases := []struct {
		name, test string
		edits      map[string]any
		want       task.AdmissionCode
	}{
		{"uppercase name", "DNS-01", map[string]any{"host": "Example.TEST"}, 0},
		{"uppercase resolver", "DNS-01", map[string]any{"resolver": "RESOLVER.example.test"}, 0},
		{"noncanonical IPv6 bootstrap", "DNS-01", map[string]any{"bootstrap": []any{"192.0.2.53", "2001:DB8:0:0::53"}}, 0},
		{"uppercase URL scheme and host", "HTTP-01", map[string]any{"url": "HTTPS://WWW.Example.test/status"}, 0},
		{"uppercase server name", "TLS-01", map[string]any{"server_name": "TLS.example.TEST"}, 0},
		{"suffix collision", "DNS-01", map[string]any{"host": "evil-example.test"}, task.TargetNotRegistered},
		{"subdomain", "DNS-01", map[string]any{"host": "sub.example.test"}, task.TargetNotRegistered},
		{"parent domain", "TLS-01", map[string]any{"server_name": "example.test"}, task.TargetNotRegistered},
		{"trailing dot", "DNS-01", map[string]any{"host": "example.test."}, task.TargetNotRegistered},
		{"wildcard", "DNS-01", map[string]any{"host": "*.example.test"}, task.TargetNotRegistered},
		{"URL path prefix", "HTTP-01", map[string]any{"url": "https://www.example.test/status/extra"}, task.TargetNotRegistered},
		{"URL extra leading path", "HTTP-01", map[string]any{"url": "https://www.example.test/prefix/status"}, task.TargetNotRegistered},
		{"URL host suffix", "HTTP-01", map[string]any{"url": "https://www.example.test.evil.test/status"}, task.TargetNotRegistered},
		{"URL path case", "HTTP-01", map[string]any{"url": "https://www.example.test/Status"}, task.TargetNotRegistered},
		{"IPv4-mapped IPv6 address", "TCP-01", map[string]any{"address": "::ffff:192.0.2.10"}, task.TargetNotRegistered},
		{"IPv4-mapped bootstrap", "DNS-01", map[string]any{"bootstrap": []any{"::ffff:192.0.2.53", "2001:db8::53"}}, task.TargetNotRegistered},
		{"malformed address", "TCP-01", map[string]any{"address": "192.0.2"}, task.SchemaInvalid},
	}
	for _, c := range cases {
		_, err := h.embedded(t, tupleTask(c.test, c.edits, nil))
		expect(t, c.test+" "+c.name, err, c.want)
	}
}

// A registration is current only while its expiry is after trusted now.
// requirement-ref: REQ-TRUST-0013
// requirement-ref: REQ-TRUST-0014
func TestGateTargetRegistrationExpiry(t *testing.T) {
	for _, c := range []struct {
		offset time.Duration
		want   task.AdmissionCode
	}{{time.Second, 0}, {0, task.TargetNotRegistered}, {-time.Second, task.TargetNotRegistered}} {
		h := tupleHarness(t, func(in *task.PolicyInput) {
			for n := range in.Targets {
				in.Targets[n].ExpiresAt = task.TimestampFromTime(epoch.Add(c.offset))
			}
		})
		for _, test := range tupleTests {
			_, err := h.embedded(t, tupleTask(test, nil, nil))
			expect(t, test+" expiry "+c.offset.String(), err, c.want)
		}
	}
	// An expired redirect registration fails the task even though its
	// measurement registration is current.
	h := tupleHarness(t, func(in *task.PolicyInput) {
		for n := range in.Targets {
			if in.Targets[n].ID == "http-redirect" {
				in.Targets[n].ExpiresAt = task.TimestampFromTime(epoch)
			}
		}
	})
	_, err := h.embedded(t, tupleTask("HTTP-01", nil, nil))
	expect(t, "expired redirect", err, task.TargetNotRegistered)
}

// Test/version, role, issuer and local permission, path and redirect listing
// are each required of the matched registration.
// requirement-ref: REQ-TRUST-0013
// requirement-ref: REQ-TRUST-0014
func TestGateTargetRegistrationAuthority(t *testing.T) {
	without := func(id string) func(*task.Permissions) {
		return func(p *task.Permissions) {
			p.RegistrationIDs = slicesWithout(p.RegistrationIDs, id)
		}
	}
	target := func(id string, edit func(*task.TargetRegistration)) func(*task.PolicyInput) {
		return func(in *task.PolicyInput) {
			for n := range in.Targets {
				if in.Targets[n].ID == id {
					edit(&in.Targets[n])
				}
			}
		}
	}
	cases := []struct {
		name, test string
		edit       func(*task.PolicyInput)
		extra      map[string]any
		want       task.AdmissionCode
	}{
		{"issuer lacks registration", "TCP-01", func(in *task.PolicyInput) { without("tcp-endpoint")(&in.Issuers[0].Permissions) }, nil, task.TargetNotRegistered},
		{"local policy lacks registration", "TCP-01", func(in *task.PolicyInput) { without("tcp-endpoint")(&in.Permissions) }, nil, task.TargetNotRegistered},
		{"other issuer lacks registration", "TCP-01", func(in *task.PolicyInput) { without("tcp-endpoint")(&in.Issuers[1].Permissions) }, nil, 0},
		{"issuer lacks redirect registration", "HTTP-01", func(in *task.PolicyInput) { without("http-redirect")(&in.Issuers[0].Permissions) }, nil, task.TargetNotRegistered},
		{"registration for another test", "TCP-01", target("tcp-endpoint", func(r *task.TargetRegistration) { r.Test.ID = "TLS-01" }), nil, task.TargetNotRegistered},
		{"registration for another version", "TCP-01", func(in *task.PolicyInput) {
			for _, p := range []*task.Permissions{&in.Permissions, &in.Issuers[0].Permissions, &in.Issuers[1].Permissions} {
				p.Tests = append(p.Tests, task.TestVersion{ID: "TCP-01", Version: "0.2.0"})
			}
			target("tcp-endpoint", func(r *task.TargetRegistration) { r.Test.Version = "0.2.0" })(in)
		}, nil, task.TargetNotRegistered},
		{"validation role", "TCP-01", target("tcp-endpoint", func(r *task.TargetRegistration) { r.Role = task.ValidationTarget }), nil, task.TargetNotRegistered},
		{"resolver role", "DNS-01", target("dns-name", func(r *task.TargetRegistration) { r.Role = task.ResolverTarget }), nil, task.TargetNotRegistered},
		{"redirect not listed", "HTTP-01", target("http-status", func(r *task.TargetRegistration) { r.RedirectIDs = nil }), nil, task.TargetNotRegistered},
		{"measurement tuple only registered as redirect", "HTTP-01", func(in *task.PolicyInput) {
			target("http-status", func(r *task.TargetRegistration) { r.Role, r.RedirectIDs = task.RedirectTarget, nil })(in)
		}, nil, task.TargetNotRegistered},
		{"proxy-only registration for a direct task", "TCP-01", target("tcp-endpoint", func(r *task.TargetRegistration) { r.Paths = task.Paths{Proxy: true} }), nil, task.TargetNotRegistered},
		{"proxy registration for a proxy task", "TCP-01", target("tcp-endpoint", func(r *task.TargetRegistration) { r.Paths = task.Paths{Proxy: true} }), map[string]any{"execution.network_path": "proxy"}, 0},
		{"redirect not permitting the path", "HTTP-01", func(in *task.PolicyInput) {
			target("http-status", func(r *task.TargetRegistration) { r.Paths = task.Paths{Direct: true, Proxy: true} })(in)
		}, map[string]any{"execution.network_path": "proxy"}, task.TargetNotRegistered},
		// The same tuple registered twice: the first that satisfies every term
		// matches, so a proxy-only duplicate does not hide a direct one.
		{"duplicate tuple on another path", "TCP-01", func(in *task.PolicyInput) {
			dup := tupleRegistration("tcp-proxy", "TCP-01", task.MeasurementTarget, in.Targets[0].Tuple)
			dup.Paths = task.Paths{Proxy: true}
			in.Targets = append([]task.TargetRegistration{dup}, in.Targets...)
			for _, p := range []*task.Permissions{&in.Permissions, &in.Issuers[0].Permissions, &in.Issuers[1].Permissions} {
				p.RegistrationIDs = append([]string{"tcp-proxy"}, p.RegistrationIDs...)
			}
		}, nil, 0},
	}
	for _, c := range cases {
		h := tupleHarness(t, c.edit)
		_, err := h.embedded(t, tupleTask(c.test, nil, c.extra))
		expect(t, c.name, err, c.want)
	}
}

func slicesWithout(ids []string, id string) []string {
	var out []string
	for _, s := range ids {
		if s != id {
			out = append(out, s)
		}
	}
	return out
}

// An incomplete tuple, parameters of an unexpected type, no measurement target
// and more than the local target bound are target_not_registered.
// requirement-ref: REQ-TRUST-0013
// requirement-ref: REQ-TRUST-0014
func TestGateTargetRegistrationRequiresCompleteTuples(t *testing.T) {
	h := tupleHarness(t, nil)
	for _, c := range []struct {
		name, test string
		edits      map[string]any
	}{
		{"TCP without port", "TCP-01", map[string]any{"port": nil}},
		{"TCP without address", "TCP-01", map[string]any{"address": nil}},
		{"TCP without scheme", "TCP-01", map[string]any{"scheme": nil}},
		{"TLS without server name", "TLS-01", map[string]any{"server_name": nil}},
		{"TLS without ALPN", "TLS-01", map[string]any{"alpn": nil}},
		{"HTTP without status", "HTTP-01", map[string]any{"status": nil}},
		{"HTTP without URL", "HTTP-01", map[string]any{"url": nil}},
		{"DNS without resolver", "DNS-01", map[string]any{"resolver": nil}},
		{"DNS without bootstrap", "DNS-01", map[string]any{"bootstrap": nil}},
		{"redirect on a non-HTTP test", "TCP-01", map[string]any{"redirects": []any{"https://cdn.example.test/status"}}},
	} {
		_, err := h.embedded(t, tupleTask(c.test, c.edits, nil))
		expect(t, c.name, err, task.TargetNotRegistered)
	}

	// Rules over SYN-01 parameters exercise the remaining outcomes.
	reg := registry.Target{Role: task.MeasurementTarget, Tuple: task.TargetTuple{Hostname: "reg-a.example.test"}}
	redirectOnly := registry.Target{Role: task.RedirectTarget, Tuple: reg.Tuple}
	many := func(n int) fixedTargets {
		out := make(fixedTargets, n)
		for i := range out {
			out[i] = reg
		}
		return out
	}
	for _, c := range []struct {
		name  string
		rule  registry.TargetRule
		codec taskdecode.TestCodec
		want  task.AdmissionCode
	}{
		{"no targets", fixedTargets{}, synCodec{}, task.TargetNotRegistered},
		{"redirect only", fixedTargets{redirectOnly}, synCodec{}, task.TargetNotRegistered},
		{"unexpected parameter type", tupleRule{id: "SYN-01"}, synCodec{}, task.TargetNotRegistered},
		{"at the target bound", many(64), synCodec{}, 0},
		{"over the target bound", many(65), synCodec{}, task.TargetNotRegistered},
	} {
		h := newHarness(t)
		rebind(t, h, mustPolicy(t, permissive()), registry.Entry{Test: task.TestVersion{ID: "SYN-01", Version: "0.1.0"}, Codec: c.codec, BaselineRisk: task.LowRisk, Targets: c.rule, Proxy: registry.NoProxy{}})
		_, err := h.embedded(t, []byte(baseTask))
		expect(t, c.name, err, c.want)
	}
	// A registered resolver alone is not a measurement destination.
	in := permissive()
	in.Targets[0].Role = task.ResolverTarget
	resolverOnly := fixedTargets{{Role: task.ResolverTarget, Tuple: reg.Tuple}}
	for _, c := range []struct {
		name string
		rule fixedTargets
		want task.AdmissionCode
	}{
		{"registered resolver only", resolverOnly, task.TargetNotRegistered},
		{"registered resolver with a measurement target", append(resolverOnly, registry.Target{Role: task.MeasurementTarget, Tuple: task.TargetTuple{Hostname: "reg-b.example.test"}}), 0},
	} {
		local := in
		local.Permissions.RegistrationIDs = []string{"reg-a", "reg-b"}
		local.Issuers = []task.IssuerPermissions{{IssuerID: "embedded-issuer", Permissions: local.Permissions}}
		local.Keys = local.Keys[1:] // keep only the embedded issuer's key
		h := newHarness(t)
		rebind(t, h, mustPolicy(t, local), registry.Entry{Test: task.TestVersion{ID: "SYN-01", Version: "0.1.0"}, Codec: synCodec{}, BaselineRisk: task.LowRisk, Targets: c.rule, Proxy: registry.NoProxy{}})
		_, err := h.embedded(t, []byte(baseTask))
		expect(t, c.name, err, c.want)
	}
}

// Target matching is part of step 8: it follows expiry and path authority and
// precedes the step-9 risk bounds.
// requirement-ref: REQ-TRUST-0047
func TestGateTargetRegistrationPrecedence(t *testing.T) {
	unregistered := map[string]any{"address": "192.0.2.99"}
	cases := []struct {
		name  string
		edit  func(*task.PolicyInput)
		extra map[string]any
		want  task.AdmissionCode
	}{
		{"expired task with an unregistered target", nil, map[string]any{"schedule.not_after": "2026-10-19T11:59:59Z"}, task.Expired},
		{"unknown field with an unregistered target", nil, map[string]any{"private": 1}, task.SchemaInvalid},
		{"issuer path refusal with an unregistered target", func(in *task.PolicyInput) { in.Issuers[0].Permissions.Paths = task.Paths{Proxy: true} }, nil, task.IssuerNotAuthorized},
		{"unregistered target with a risk above the issuer maximum", func(in *task.PolicyInput) {
			in.Issuers[0].Permissions.MaxRisk, in.Issuers[0].Permissions.AllowMediumRisk = task.LowRisk, false
		}, map[string]any{"safety.risk_level": "medium", "safety.risk_factors": []any{map[string]any{"type": "data_sensitivity", "level": "medium"}}}, task.TargetNotRegistered},
	}
	for _, c := range cases {
		h := tupleHarness(t, c.edit)
		_, err := h.embedded(t, tupleTask("TCP-01", unregistered, c.extra))
		expect(t, c.name, err, c.want)
	}
}

// Registrations carrying URL credentials or interpolation syntax never become
// part of a policy snapshot, so no task can match them.
// requirement-ref: REQ-TRUST-0015
func TestTargetRegistrationRejectsCredentialsAndInterpolation(t *testing.T) {
	for name, tuple := range map[string]task.TargetTuple{
		"URL credentials":     {URL: "https://user:secret@www.example.test/status"},
		"URL user":            {URL: "https://user@www.example.test/status"},
		"URL interpolation":   {URL: "https://www.example.test/{country}"},
		"host interpolation":  {Hostname: "{host}.example.test"},
		"ALPN interpolation":  {Hostname: "example.test", ALPN: []string{"{alpn}"}},
		"wildcard host":       {Hostname: "*.example.test"},
		"suffix-rule host":    {Hostname: ".example.test"},
		"IPv4-mapped address": {Address: addr("::ffff:192.0.2.10")},
	} {
		in := tuplePolicy(func(in *task.PolicyInput) { in.Targets[0].Tuple = tuple })
		if _, err := task.NewLocalPolicy(in); !errors.Is(err, task.ErrTrustModel) {
			t.Errorf("%s: registration accepted", name)
		}
	}
}
