package barometer

import (
	"crypto/x509"
	"errors"
	"testing"

	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/testkit/contextendpoint"
)

// Outside Android the seam has no platform loader and keeps Go's system
// roots; a loader result is used as given, and a loader error or an empty
// result reports the root set unavailable instead of falling back.
func TestSeamRootsFailClosed(t *testing.T) {
	if platformRoots != nil {
		t.Fatal("a host build has a platform root loader")
	}
	if roots, ok := seamRoots(nil); roots != nil || !ok {
		t.Fatal("no loader must keep the system roots")
	}
	pool := x509.NewCertPool()
	if roots, ok := seamRoots(func() (*x509.CertPool, error) { return pool, nil }); roots != pool || !ok {
		t.Fatal("a loaded root set was not used")
	}
	for name, load := range map[string]func() (*x509.CertPool, error){
		"error":    func() (*x509.CertPool, error) { return x509.NewCertPool(), errors.New("no store") },
		"nil pool": func() (*x509.CertPool, error) { return nil, nil },
	} {
		if roots, ok := seamRoots(load); roots != nil || ok {
			t.Fatalf("%s: unavailable roots were not reported", name)
		}
	}
}

// Without a root set the production registry holds only the tests that
// verify no certificate; every TLS-dependent test is absent.
func TestProductionRegistryWithoutRoots(t *testing.T) {
	r, err := productionRegistry(nil, false)
	if err != nil || r == nil {
		t.Fatal("registry without roots rejected")
	}
	if len(r.Codecs()) != 3 {
		t.Fatalf("registered %d tests, want DNS-01, TCP-01 and NET-01", len(r.Codecs()))
	}
	for _, id := range []string{"DNS-01", "TCP-01", "NET-01"} {
		if _, ok := r.Lookup(task.TestVersion{ID: id, Version: "0.1.0"}); !ok {
			t.Fatalf("%s missing", id)
		}
	}
	for _, id := range []string{"TLS-01", "HTTP-01", "NET-03", "DNS-03", "DNS-04", "NET-05"} {
		if _, ok := r.Lookup(task.TestVersion{ID: id, Version: "0.1.0"}); ok {
			t.Fatalf("%s registered without a root set", id)
		}
	}
}

// A client bound with an unavailable root set rejects a TLS-dependent task
// as task_unsupported before any contact, and leaves configured egress
// context unavailable without contacting the endpoint, while a task that
// verifies no certificate still runs.
func TestUnavailableRootsFailTLSWorkClosed(t *testing.T) {
	h := runningHarness(t)
	h.adapters.registry = nil
	h.adapters.run.paths = nil
	h.adapters.rootsUnavailable = true
	httpPolicy(t, h, httpInput)
	d := &httpTransport{}
	h.adapters.run.directDialers = d
	c := h.initialized(t)
	t.Cleanup(func() { _ = c.Close() })
	data := httpTask("http-no-roots", httpInput, `{}`)
	if _, err := c.AdmitTask(data, h.embedded(t, data)); codeOf(err) != "task_unsupported" {
		t.Fatalf("HTTP-01 without roots: %v", err)
	}

	h, _, e := contextHarness(t, contextendpoint.Response{Body: nlClassified})
	h.adapters.rootsUnavailable = true
	c = initializeWith(t, h, contextConfig)
	res := startAndWait(t, h, c, runTask("context-no-roots", "", "", false))
	if eg := res.NetworkContext.Egress; res.Execution.State != "completed" || eg["status"] != "unavailable" || e.requests() != 0 {
		t.Fatalf("egress %v with %d requests", eg, e.requests())
	}
}
