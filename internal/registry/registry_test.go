package registry

import (
	"errors"
	"strconv"
	"sync"
	"testing"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/taskdecode"
)

// Test-scoped synthetic entries only. Codecs reuse registered test IDs where a
// rule depends on them, but none is a production entry or a runner.
type fixtureParameters struct{ TCP bool }

func (fixtureParameters) TaskParameters() {}

type fixtureCodec struct{ def taskdecode.Definition }

func (c fixtureCodec) Definition() taskdecode.Definition { return c.def }
func (fixtureCodec) Normalize(_, _ *jsondecode.Value) (task.Parameters, error) {
	return fixtureParameters{}, nil
}

func codec(id, ver string) fixtureCodec {
	object := func() *jsondecode.Schema {
		return &jsondecode.Schema{Kind: jsondecode.Object, Fields: map[string]jsondecode.Field{}}
	}
	return fixtureCodec{def: taskdecode.Definition{TestID: id, TestVersion: ver, Input: object(), Options: object()}}
}

// tcpOnly is conditional compatibility in the shape of REQ-PROXY-0026.
type tcpOnly struct{}

func (tcpOnly) Compatible(p task.Parameters) bool {
	f, ok := p.(fixtureParameters)
	return ok && f.TCP
}

// fixedTarget reports one measurement tuple, for fixture parameters only.
type fixedTarget struct{ host string }

func (r fixedTarget) Targets(p task.Parameters) ([]Target, bool) {
	if _, ok := p.(fixtureParameters); !ok {
		return nil, false
	}
	return []Target{{Role: task.MeasurementTarget, Tuple: task.TargetTuple{Hostname: r.host}}}, true
}

func entry(id, ver string) Entry {
	return Entry{Test: task.TestVersion{ID: id, Version: ver}, Codec: codec(id, ver), BaselineRisk: task.LowRisk, Targets: fixedTarget{host: "example.test"}, Proxy: NoProxy{}}
}

func fixtureEntries() []Entry {
	syn := entry("SYN-01", "0.1.0")
	syn2 := entry("SYN-01", "0.2.0")
	syn2.BaselineRisk = task.MediumRisk
	dns := entry("DNS-01", "0.1.0")
	dns.Proxy, dns.Step = tcpOnly{}, true
	net05 := entry("NET-05", "0.1.0")
	net05.BaselineRisk, net05.Composite = task.MediumRisk, true
	return []Entry{syn, syn2, dns, net05}
}

func mustNew(t *testing.T, entries []Entry) *Registry {
	t.Helper()
	r, err := New(entries)
	if err != nil || r == nil {
		t.Fatalf("New() = %v, %v", r, err)
	}
	return r
}

func TestLookupIsExact(t *testing.T) {
	r := mustNew(t, fixtureEntries())
	for _, e := range fixtureEntries() {
		got, ok := r.Lookup(e.Test)
		if !ok || got.Test != e.Test || got.BaselineRisk != e.BaselineRisk || got.Composite != e.Composite || got.Step != e.Step {
			t.Fatalf("Lookup(%v) = %+v, %v", e.Test, got, ok)
		}
		if def := got.Codec.Definition(); def.TestID != e.Test.ID || def.TestVersion != e.Test.Version {
			t.Fatalf("Lookup(%v) codec = %v", e.Test, def)
		}
	}
	for _, miss := range []task.TestVersion{
		{ID: "SYN-01", Version: "0.3.0"},  // no latest-version fallback
		{ID: "SYN-01", Version: "0.1"},    // no prefix match
		{ID: "SYN-01", Version: "0.1.0 "}, // no trimming
		{ID: "SYN-01", Version: "v0.1.0"}, // no alias
		{ID: "syn-01", Version: "0.1.0"},  // no case folding
		{ID: "SYN-0", Version: "0.1.0"},   // no partial ID
		{ID: "SYN-02", Version: "0.1.0"},  // unknown test
		{ID: "SYN-01"},                    // missing version
		{Version: "0.1.0"},                // missing ID
		{ID: "TCP-01", Version: "0.1.0"},  // registered publicly, not here
	} {
		if got, ok := r.Lookup(miss); ok || got != (Entry{}) {
			t.Errorf("Lookup(%q) = %+v, %v; want miss", miss, got, ok)
		}
	}
	var nilRegistry *Registry
	if _, ok := nilRegistry.Lookup(task.TestVersion{ID: "SYN-01", Version: "0.1.0"}); ok || nilRegistry.Codecs() != nil {
		t.Fatal("nil registry must not resolve entries")
	}
}

func TestEntryMetadata(t *testing.T) {
	r := mustNew(t, fixtureEntries())
	dns, _ := r.Lookup(task.TestVersion{ID: "DNS-01", Version: "0.1.0"})
	if !dns.Proxy.Compatible(fixtureParameters{TCP: true}) || dns.Proxy.Compatible(fixtureParameters{}) || dns.Proxy.Compatible(nil) {
		t.Fatal("conditional proxy rule not preserved")
	}
	syn, _ := r.Lookup(task.TestVersion{ID: "SYN-01", Version: "0.1.0"})
	if syn.Proxy.Compatible(fixtureParameters{TCP: true}) || syn.Composite || syn.Step {
		t.Fatal("NoProxy entry must never be compatible or composite-eligible")
	}
	if got, ok := syn.Targets.Targets(fixtureParameters{}); !ok || len(got) != 1 || got[0].Tuple.Hostname != "example.test" {
		t.Fatal("target rule not preserved")
	}
	if _, ok := syn.Targets.Targets(nil); ok {
		t.Fatal("target rule accepted parameters of an unexpected type")
	}
	syn2, _ := r.Lookup(task.TestVersion{ID: "SYN-01", Version: "0.2.0"})
	if syn.BaselineRisk != task.LowRisk || syn2.BaselineRisk != task.MediumRisk {
		t.Fatal("baseline risk must be per exact version")
	}
	codecs := r.Codecs()
	if len(codecs) != 4 {
		t.Fatalf("Codecs() len = %d", len(codecs))
	}
	if _, err := taskdecode.New(codecs); err != nil {
		t.Fatalf("decoder from registry codecs: %v", err)
	}
}

func TestNewRejectsInvalidSetsAtomically(t *testing.T) {
	mutate := func(i int, f func(*Entry)) []Entry {
		entries := fixtureEntries()
		f(&entries[i])
		return entries
	}
	tooMany := make([]Entry, maxEntries+1)
	for i := range tooMany {
		tooMany[i] = entry("SYN-01", "0.0."+strconv.Itoa(i))
	}
	cases := map[string][]Entry{
		"nil":                    nil,
		"empty":                  {},
		"over bound":             tooMany,
		"duplicate":              append(fixtureEntries(), entry("SYN-01", "0.1.0")),
		"missing codec":          mutate(0, func(e *Entry) { e.Codec = nil }),
		"missing target rule":    mutate(0, func(e *Entry) { e.Targets = nil }),
		"missing proxy rule":     mutate(0, func(e *Entry) { e.Proxy = nil }),
		"missing risk":           mutate(0, func(e *Entry) { e.BaselineRisk = 0 }),
		"out-of-range risk":      mutate(0, func(e *Entry) { e.BaselineRisk = task.UnknownRisk + 1 }),
		"codec ID mismatch":      mutate(0, func(e *Entry) { e.Codec = codec("SYN-02", "0.1.0") }),
		"codec version mismatch": mutate(0, func(e *Entry) { e.Codec = codec("SYN-01", "0.2.0") }),
		"missing test ID":        mutate(0, func(e *Entry) { e.Test.ID, e.Codec = "", codec("", "0.1.0") }),
		"malformed version":      mutate(0, func(e *Entry) { e.Test.Version, e.Codec = "01.0.0", codec("SYN-01", "01.0.0") }),
		"missing input schema":   mutate(0, func(e *Entry) { c := codec("SYN-01", "0.1.0"); c.def.Input = nil; e.Codec = c }),
		"step not a primitive":   mutate(0, func(e *Entry) { e.Step = true }),
		"composite not NET-05":   mutate(0, func(e *Entry) { e.Composite = true }),
		"composite as step":      mutate(3, func(e *Entry) { e.Step = true }),
		"primitive as composite": mutate(2, func(e *Entry) { e.Composite = true }),
	}
	for name, entries := range cases {
		t.Run(name, func(t *testing.T) {
			r, err := New(entries)
			if r != nil || !errors.Is(err, ErrInvalid) || err.Error() != ErrInvalid.Error() {
				t.Fatalf("New() = %v, %v; want nil, ErrInvalid", r, err)
			}
		})
	}
	if _, err := New(tooMany[:maxEntries]); err != nil {
		t.Fatalf("bound must admit %d entries: %v", maxEntries, err)
	}
}

func TestRegistryIsImmutableAfterConstruction(t *testing.T) {
	entries := fixtureEntries()
	r := mustNew(t, entries)
	key := entries[0].Test
	entries[0].BaselineRisk = task.HighRisk
	entries[0].Test.Version = "9.9.9"
	got, ok := r.Lookup(key)
	if !ok || got.BaselineRisk != task.LowRisk || got.Test != key {
		t.Fatalf("caller mutation reached registry: %+v, %v", got, ok)
	}
	got.BaselineRisk = task.HighRisk
	if again, _ := r.Lookup(key); again.BaselineRisk != task.LowRisk {
		t.Fatal("returned entry aliases registry state")
	}
	codecs := r.Codecs()
	codecs[0] = nil
	if again := r.Codecs(); again[0] == nil {
		t.Fatal("Codecs() aliases registry state")
	}
}

func TestConcurrentReads(t *testing.T) {
	r := mustNew(t, fixtureEntries())
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				for _, e := range fixtureEntries() {
					if got, ok := r.Lookup(e.Test); !ok || got.Test != e.Test {
						t.Error("concurrent lookup failed")
						return
					}
				}
				if len(r.Codecs()) != 4 {
					t.Error("concurrent Codecs failed")
					return
				}
			}
		}()
	}
	wg.Wait()
}
