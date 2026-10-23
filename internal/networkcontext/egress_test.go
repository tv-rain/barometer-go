package networkcontext_test

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"net/url"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/observation"
	"github.com/tv-rain/barometer-go/internal/networkcontext"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/testkit"
	"github.com/tv-rain/barometer-go/internal/testkit/capability"
)

// Egress acquisition cases (DEV-647) with a scripted fetch: no socket is
// opened here. The fetch stands in for the HTTPS request that
// internal/execution makes through the attempt's path dialer; the request
// itself, TLS and HTTP validation are exercised against synthetic loopback
// endpoints in internal/execution. These are synthetic core contract cases,
// not field or platform evidence.

const (
	primaryURL  = "https://context.example/network-context"
	fallbackURL = "https://backup.example/network-context"
	fullBody    = `{"country_code":"NL","region_code":"NL-NH","asn":64500,"classification_version":"0.1.0"}`
)

var egressApprovals = observation.ContextApprovals{ClassificationVersions: []string{"0.1.0"}, CountryCodes: []string{"NL", "DE"}, RegionCodes: []string{"NL-NH"}}

// endpoints scripts a fetch per endpoint host and records each call.
type endpoints struct {
	mu    sync.Mutex
	calls []string
	reply map[string]func(ctx context.Context) ([]byte, error)
	roots []*x509.CertPool
}

func newEndpoints() *endpoints {
	return &endpoints{reply: map[string]func(context.Context) ([]byte, error){}}
}

func (e *endpoints) body(host, body string) {
	e.reply[host] = func(context.Context) ([]byte, error) { return []byte(body), nil }
}
func (e *endpoints) fail(host string) {
	e.reply[host] = func(context.Context) ([]byte, error) { return nil, errors.New("synthetic endpoint failure") }
}

func (e *endpoints) fetch(ctx context.Context, u *url.URL, roots *x509.CertPool) ([]byte, error) {
	e.mu.Lock()
	e.calls = append(e.calls, u.String())
	e.roots = append(e.roots, roots)
	f := e.reply[u.Hostname()]
	e.mu.Unlock()
	if f == nil {
		return nil, errors.New("unscripted")
	}
	return f(ctx)
}

func (e *endpoints) count() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	n := len(e.calls)
	e.calls = nil
	return n
}

// world is one lookup over a fake platform network and a manual clock.
type world struct {
	t     *testing.T
	clock *testkit.Clock
	net   *capability.Network
	cache *networkcontext.Cache
	look  *networkcontext.Lookup
	ep    *endpoints
}

func newWorld(t *testing.T, fallback string) *world {
	t.Helper()
	w := &world{t: t, clock: testkit.NewClock(time.Unix(1_800_000_000, 0)), net: capability.NewNetwork(platform.AccessState{Generation: 1, Type: platform.NetworkWiFi}),
		cache: networkcontext.NewCache(), ep: newEndpoints()}
	w.relookup(fallback)
	return w
}

func (w *world) relookup(fallback string) {
	w.t.Helper()
	l, err := networkcontext.NewLookup(networkcontext.Endpoints{Primary: primaryURL, Fallback: fallback}, egressApprovals, nil, w.clock, w.cache)
	if err != nil {
		w.t.Fatal(err)
	}
	w.look = l
}

func (w *world) source(proxyGeneration uint64) *networkcontext.Source {
	w.t.Helper()
	var s *networkcontext.Source
	var err error
	if proxyGeneration == 0 {
		s, err = networkcontext.NewDirect(w.net)
	} else {
		s, err = networkcontext.NewProxy(w.net, proxyGeneration, "outline_mobileproxy")
	}
	if err == nil {
		s, err = s.WithLookup(w.look)
	}
	if err != nil {
		w.t.Fatal(err)
	}
	return s
}

// acquire captures on the direct path (proxyGeneration 0) or the given proxy
// generation and acquires egress through the scripted endpoints.
func (w *world) acquire(proxyGeneration uint64) observation.Egress {
	w.t.Helper()
	snap, err := w.source(proxyGeneration).Capture(context.Background())
	if err != nil {
		w.t.Fatal(err)
	}
	snap, err = snap.Acquire(context.Background(), w.ep.fetch)
	if err != nil {
		w.t.Fatal(err)
	}
	return snap.Egress()
}

func wantEgress(t *testing.T, name string, e observation.Egress, status, source string) {
	t.Helper()
	if e.Status != status || e.LookupSource != source {
		t.Fatalf("%s: egress %s/%s, want %s/%s", name, e.Status, e.LookupSource, status, source)
	}
	if status == "unavailable" && (e.CountryCode != "" || e.RegionCode != "" || e.ASN != nil || e.ClassificationVersion != "") {
		t.Fatalf("%s: unavailable egress carries a classification: %+v", name, e)
	}
}

// CC-CONTEXT-ACQUISITION-AND-FALLBACK: the bounded order of cache, one
// primary request, one fallback request and unavailable; a valid primary
// with unknown values does not trigger the fallback; no endpoint is retried;
// failures are not cached; the cache key is exactly the access-network
// generation, path and proxy generation; freshness and every invalidation
// event. The cache is an in-memory value with no persistence port.
// requirement-ref: REQ-CONTEXT-0020
// requirement-ref: REQ-CONTEXT-0021
// requirement-ref: REQ-CONTEXT-0022
// requirement-ref: REQ-CONTEXT-0023
// requirement-ref: REQ-CONTEXT-0024
func TestCcContextAcquisitionAndFallback(t *testing.T) {
	t.Run("primary success", func(t *testing.T) {
		w := newWorld(t, fallbackURL)
		w.ep.body("context.example", fullBody)
		e := w.acquire(0)
		wantEgress(t, "primary", e, "fresh", "primary")
		if e.CountryCode != "NL" || e.RegionCode != "NL-NH" || e.ASN == nil || *e.ASN != 64500 || e.ClassificationVersion != "0.1.0" || w.ep.count() != 1 {
			t.Fatalf("egress %+v", e)
		}
	})
	t.Run("valid primary with unknown values", func(t *testing.T) {
		w := newWorld(t, fallbackURL)
		w.ep.body("context.example", `{"classification_version":"0.1.0"}`)
		w.ep.body("backup.example", fullBody)
		e := w.acquire(0)
		wantEgress(t, "unknown values", e, "fresh", "primary")
		if e.CountryCode != "" || e.ASN != nil || w.ep.count() != 1 {
			t.Fatal("fallback contacted or a value guessed")
		}
	})
	for name, primary := range map[string]func(*endpoints){
		"unreachable":      func(e *endpoints) { e.fail("context.example") },
		"malformed":        func(e *endpoints) { e.body("context.example", `{"country_code":`) },
		"missing version":  func(e *endpoints) { e.body("context.example", `{"country_code":"NL"}`) },
		"returned address": func(e *endpoints) { e.body("context.example", `{"ip":"192.0.2.1","classification_version":"0.1.0"}`) },
		"unapproved country": func(e *endpoints) {
			e.body("context.example", `{"country_code":"FR","classification_version":"0.1.0"}`)
		},
		"unapproved version": func(e *endpoints) { e.body("context.example", `{"classification_version":"9.9.9"}`) },
		"region without country": func(e *endpoints) {
			e.body("context.example", `{"region_code":"NL-NH","classification_version":"0.1.0"}`)
		},
		"oversized": func(e *endpoints) {
			e.body("context.example", `{"classification_version":"0.1.0"}`+strings.Repeat(" ", 4096))
		},
	} {
		t.Run("fallback after "+name, func(t *testing.T) {
			w := newWorld(t, fallbackURL)
			primary(w.ep)
			w.ep.body("backup.example", `{"country_code":"DE","classification_version":"0.1.0"}`)
			e := w.acquire(0)
			wantEgress(t, name, e, "fresh", "fallback")
			if e.CountryCode != "DE" || w.ep.count() != 2 {
				t.Fatalf("egress %+v", e)
			}
		})
	}
	t.Run("both fail", func(t *testing.T) {
		w := newWorld(t, fallbackURL)
		w.ep.fail("context.example")
		w.ep.fail("backup.example")
		wantEgress(t, "both fail", w.acquire(0), "unavailable", "none")
		if n := w.ep.count(); n != 2 {
			t.Fatalf("%d requests, want one per endpoint", n)
		}
		// A failure is not cached: the next acquisition asks again.
		w.ep.body("context.example", fullBody)
		wantEgress(t, "after failure", w.acquire(0), "fresh", "primary")
	})
	t.Run("no fallback configured", func(t *testing.T) {
		w := newWorld(t, "")
		w.ep.fail("context.example")
		wantEgress(t, "no fallback", w.acquire(0), "unavailable", "none")
		if w.ep.count() != 1 {
			t.Fatal("more than one request")
		}
	})
	t.Run("cache key and freshness", func(t *testing.T) {
		w := newWorld(t, fallbackURL)
		w.ep.body("context.example", fullBody)
		w.acquire(0)
		w.ep.count()
		wantEgress(t, "same key", w.acquire(0), "cached", "cache")
		_ = w.clock.Advance(networkcontext.DefaultFreshness)
		wantEgress(t, "at freshness bound", w.acquire(0), "cached", "cache")
		if w.ep.count() != 0 {
			t.Fatal("cache hit made a request")
		}
		_ = w.clock.Advance(time.Second)
		wantEgress(t, "after freshness bound", w.acquire(0), "fresh", "primary")
		w.ep.count()

		// Proxy and direct contexts are separate entries; each proxy
		// generation is its own key.
		wantEgress(t, "proxy generation 1", w.acquire(1), "fresh", "primary")
		wantEgress(t, "proxy generation 1 again", w.acquire(1), "cached", "cache")
		wantEgress(t, "proxy generation 2", w.acquire(2), "fresh", "primary")
		wantEgress(t, "direct unaffected", w.acquire(0), "cached", "cache")
		w.ep.count()

		// ForgetProxy (proxy replacement or clearing) drops proxy entries only.
		w.cache.ForgetProxy()
		wantEgress(t, "proxy after forget", w.acquire(2), "fresh", "primary")
		wantEgress(t, "direct after forget", w.acquire(0), "cached", "cache")

		// An access-network generation change is a different key, and a
		// newer generation drops the older one's entries.
		w.net.Set(platform.AccessState{Generation: 2, Type: platform.NetworkWiFi})
		wantEgress(t, "new access generation", w.acquire(0), "fresh", "primary")
		if w.cache.Len() != 1 {
			t.Fatalf("%d entries after a generation change", w.cache.Len())
		}

		// Clear (shutdown, reinitialization, policy replacement, clearing)
		// drops every entry, and the old lookup can no longer write.
		w.cache.Clear()
		wantEgress(t, "stale lookup after clear", w.acquire(0), "fresh", "primary")
		if w.cache.Len() != 0 {
			t.Fatal("a lookup from before Clear wrote the cache")
		}
		w.relookup(fallbackURL)
		w.acquire(0)
		wantEgress(t, "new lookup", w.acquire(0), "cached", "cache")
		w.look.Close()
		if w.cache.Len() != 0 {
			t.Fatal("closing the lookup kept its entries")
		}
	})
}

// CC-CONTEXT-ENDPOINT-CONTRACT, response side: the endpoint object is decoded
// strictly, unknown or forbidden fields are invalid, a response with only the
// classification version is valid, the body is bounded to 4,096 bytes and
// each operation to 5,000 ms on the engine clock. The fetch receives exactly
// the configured URL and the lookup's trust roots. The request side runs in
// internal/execution against synthetic HTTPS endpoints.
// requirement-ref: REQ-CONTEXT-0013
// requirement-ref: REQ-CONTEXT-0014
// requirement-ref: REQ-CONTEXT-0015
// requirement-ref: REQ-CONTEXT-0016
func TestCcContextEndpointContract(t *testing.T) {
	for name, body := range map[string]string{
		"address":         `{"ip":"192.0.2.1","classification_version":"0.1.0"}`,
		"coordinates":     `{"latitude":52.3,"classification_version":"0.1.0"}`,
		"city":            `{"city":"Amsterdam","classification_version":"0.1.0"}`,
		"postal code":     `{"postal_code":"1011","classification_version":"0.1.0"}`,
		"operator name":   `{"isp":"Example","classification_version":"0.1.0"}`,
		"request id":      `{"request_id":"r-1","classification_version":"0.1.0"}`,
		"duplicate":       `{"country_code":"NL","country_code":"DE","classification_version":"0.1.0"}`,
		"asn zero":        `{"asn":0,"classification_version":"0.1.0"}`,
		"asn too large":   `{"asn":4294967296,"classification_version":"0.1.0"}`,
		"asn string":      `{"asn":"64500","classification_version":"0.1.0"}`,
		"empty country":   `{"country_code":"","classification_version":"0.1.0"}`,
		"long country":    `{"country_code":"NLD","classification_version":"0.1.0"}`,
		"region mismatch": `{"country_code":"DE","region_code":"NL-NH","classification_version":"0.1.0"}`,
		"empty version":   `{"classification_version":""}`,
		"not an object":   `["NL"]`,
		"empty":           ``,
	} {
		w := newWorld(t, "")
		w.ep.body("context.example", body)
		wantEgress(t, name, w.acquire(0), "unavailable", "none")
	}

	w := newWorld(t, "")
	w.ep.body("context.example", `{"classification_version":"0.1.0"}`)
	wantEgress(t, "version only", w.acquire(0), "fresh", "primary")
	if len(w.ep.roots) != 1 || w.ep.roots[0] != nil {
		t.Fatal("fetch did not receive the lookup's roots")
	}

	// Exactly 4,096 bytes are accepted; 4,097 are not.
	padded := func(n int) string {
		b := `{"classification_version":"0.1.0"}`
		return b + strings.Repeat(" ", n-len(b))
	}
	for n, want := range map[int]string{networkcontext.MaxResponseBytes: "fresh", networkcontext.MaxResponseBytes + 1: "unavailable"} {
		w := newWorld(t, "")
		w.ep.body("context.example", padded(n))
		if e := w.acquire(0); e.Status != want {
			t.Fatalf("%d bytes: %s", n, e.Status)
		}
	}

	// The operation bound runs on the engine clock: an answer at 4,999 ms is
	// accepted, the operation is stopped at 5,000 ms.
	for d, want := range map[time.Duration]string{networkcontext.OperationTimeout - time.Millisecond: "fresh", networkcontext.OperationTimeout + time.Millisecond: "unavailable"} {
		w := newWorld(t, "")
		w.ep.reply["context.example"] = func(ctx context.Context) ([]byte, error) {
			_ = w.clock.Advance(d)
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return []byte(fullBody), nil
		}
		if e := w.acquire(0); e.Status != want {
			t.Fatalf("answer after %v: %s", d, e.Status)
		}
		if w.clock.Timers() != 0 {
			t.Fatalf("answer after %v: an operation timer was left armed", d)
		}
	}
}

// CC-CONTEXT-FORMAT-EGRESS-OBJECT: every status and source combination the
// lookup produces on each path, connection_code only on the proxy path and
// also when egress is unavailable, missing optional classifications absent
// rather than guessed, and nothing else in the egress object.
// requirement-ref: REQ-CONTEXT-FORMAT-0006
// requirement-ref: REQ-CONTEXT-FORMAT-0007
// requirement-ref: REQ-CONTEXT-FORMAT-0008
// requirement-ref: REQ-CONTEXT-FORMAT-0009
func TestCcContextFormatEgressObject(t *testing.T) {
	egressOf := func(t *testing.T, w *world, proxy uint64) map[string]any {
		t.Helper()
		snap, err := w.source(proxy).Capture(context.Background())
		if err == nil {
			snap, err = snap.Acquire(context.Background(), w.ep.fetch)
		}
		var nc observation.NetworkContext
		if err == nil {
			nc, err = snap.Context(false)
		}
		raw, err2 := json.Marshal(nc)
		if err != nil || err2 != nil {
			t.Fatal(err, err2)
		}
		var v struct{ Egress map[string]any }
		if json.Unmarshal(raw, &v) != nil {
			t.Fatal("undecodable context")
		}
		return v.Egress
	}
	keys := func(m map[string]any) string {
		var ks []string
		for k := range m {
			ks = append(ks, k)
		}
		slices.Sort(ks)
		return strings.Join(ks, ",")
	}
	for _, proxy := range []uint64{0, 1} {
		path, code := "direct", ""
		if proxy != 0 {
			path, code = "proxy", ",connection_code"
		}
		w := newWorld(t, fallbackURL)
		w.ep.fail("context.example")
		w.ep.fail("backup.example")
		e := egressOf(t, w, proxy)
		if keys(e) != sortedKeys("lookup_source,network_path,status"+code) || e["status"] != "unavailable" || e["lookup_source"] != "none" || e["network_path"] != path {
			t.Fatalf("%s unavailable: %v", path, e)
		}
		w.ep.body("backup.example", `{"classification_version":"0.1.0"}`)
		e = egressOf(t, w, proxy)
		if keys(e) != sortedKeys("classification_version,lookup_source,network_path,status"+code) || e["status"] != "fresh" || e["lookup_source"] != "fallback" {
			t.Fatalf("%s fallback with unknown values: %v", path, e)
		}
		w.ep.body("context.example", fullBody)
		w.cache.Clear()
		w.relookup(fallbackURL)
		e = egressOf(t, w, proxy)
		if keys(e) != sortedKeys("asn,classification_version,country_code,lookup_source,network_path,region_code,status"+code) || e["status"] != "fresh" || e["lookup_source"] != "primary" {
			t.Fatalf("%s primary: %v", path, e)
		}
		e = egressOf(t, w, proxy)
		if e["status"] != "cached" || e["lookup_source"] != "cache" || e["country_code"] != "NL" {
			t.Fatalf("%s cached: %v", path, e)
		}
		if proxy != 0 && e["connection_code"] != "outline_mobileproxy" {
			t.Fatalf("proxy egress without its connection_code: %v", e)
		}
	}
}

func sortedKeys(csv string) string {
	ks := strings.Split(csv, ",")
	slices.Sort(ks)
	return strings.Join(ks, ",")
}

// Without a lookup, acquisition sends nothing and egress stays unavailable;
// NewLookup rejects an incomplete configuration.
func TestAcquireWithoutLookup(t *testing.T) {
	n := capability.NewNetwork(platform.AccessState{Generation: 1})
	d, _ := networkcontext.NewDirect(n)
	snap, err := d.Capture(context.Background())
	if err != nil || snap.Acquires() {
		t.Fatal("snapshot without a lookup acquires")
	}
	called := false
	out, err := snap.Acquire(context.Background(), func(context.Context, *url.URL, *x509.CertPool) ([]byte, error) {
		called = true
		return nil, nil
	})
	if err != nil || called || out.Egress().Status != "unavailable" {
		t.Fatal("acquired without a lookup")
	}
	clock, cache := testkit.NewClock(time.Unix(0, 0)), networkcontext.NewCache()
	for name, e := range map[string]networkcontext.Endpoints{
		"no primary":   {},
		"http":         {Primary: "http://context.example/"},
		"credentials":  {Primary: "https://u:p@context.example/"},
		"fragment":     {Primary: "https://context.example/#f"},
		"bad fallback": {Primary: primaryURL, Fallback: "ftp://backup.example/"},
	} {
		if _, err := networkcontext.NewLookup(e, egressApprovals, nil, clock, cache); err == nil {
			t.Fatal(name)
		}
	}
	if _, err := networkcontext.NewLookup(networkcontext.Endpoints{Primary: primaryURL}, egressApprovals, nil, nil, cache); err == nil {
		t.Fatal("lookup without a clock")
	}
}
