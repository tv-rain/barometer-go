// Package registry holds the reviewed mapping from an exact registered test ID
// and version to the codec, baseline risk, proxy compatibility and composite
// eligibility admission needs. It performs no I/O, holds no participant data
// and has no runtime registration. Presence of an entry is not evidence of
// released, conformant or field-validated support.
package registry

import (
	"errors"

	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/taskdecode"
)

// ErrInvalid is the single construction error. It carries no entry detail.
var ErrInvalid = errors.New("registry: invalid entries")

// Finite implementation-local bound, matching the task decoder's codec bound.
const maxEntries = 256

// ProxyRule reports whether normalized parameters of its test are compatible
// with the HTTP CONNECT proxy path. Implementations must be immutable, pure and
// bounded and must reject parameters of an unexpected type. Compatibility is
// one term of the authority intersection; it never grants proxy use.
type ProxyRule interface {
	Compatible(task.Parameters) bool
}

// NoProxy is the rule for a test that is never proxy-compatible.
type NoProxy struct{}

func (NoProxy) Compatible(task.Parameters) bool { return false }

// Entry is one reviewed test version. Codec and Proxy must remain immutable
// for the registry's lifetime.
type Entry struct {
	Test         task.TestVersion
	Codec        taskdecode.TestCodec
	BaselineRisk task.Risk
	Proxy        ProxyRule
	// Composite marks the NET-05 composite, which also requires explicit
	// composite permission. Step marks a primitive permitted as a NET-05 step.
	Composite, Step bool
}

// Registry is immutable after New and safe for concurrent reads.
type Registry struct {
	entries []Entry
	index   map[task.TestVersion]int
}

// stepTest reports the primitives permitted as NET-05 profile steps
// (REQ-NET-05-0016).
func stepTest(id string) bool {
	switch id {
	case "DNS-01", "TCP-01", "TLS-01", "HTTP-01":
		return true
	}
	return false
}

// New validates every entry before returning a registry built from owned
// copies. Any invalid, incomplete or duplicate entry rejects the whole set.
func New(entries []Entry) (*Registry, error) {
	if len(entries) == 0 || len(entries) > maxEntries {
		return nil, ErrInvalid
	}
	r := &Registry{entries: make([]Entry, 0, len(entries)), index: make(map[task.TestVersion]int, len(entries))}
	codecs := make([]taskdecode.TestCodec, 0, len(entries))
	for _, e := range entries {
		if e.Codec == nil || e.Proxy == nil || e.BaselineRisk < task.LowRisk || e.BaselineRisk > task.UnknownRisk {
			return nil, ErrInvalid
		}
		if def := e.Codec.Definition(); def.TestID != e.Test.ID || def.TestVersion != e.Test.Version {
			return nil, ErrInvalid
		}
		if e.Composite && (e.Step || e.Test.ID != "NET-05") || e.Step && !stepTest(e.Test.ID) {
			return nil, ErrInvalid
		}
		if _, dup := r.index[e.Test]; dup {
			return nil, ErrInvalid
		}
		r.index[e.Test] = len(r.entries)
		r.entries = append(r.entries, e)
		codecs = append(codecs, e.Codec)
	}
	// The decoder owns identifier, version and schema-descriptor validity.
	if _, err := taskdecode.New(codecs); err != nil {
		return nil, ErrInvalid
	}
	return r, nil
}

// Lookup matches the exact test ID and version strings. There is no latest-
// version fallback, case folding or partial match.
func (r *Registry) Lookup(test task.TestVersion) (Entry, bool) {
	if r == nil {
		return Entry{}, false
	}
	i, ok := r.index[test]
	if !ok {
		return Entry{}, false
	}
	return r.entries[i], true
}

// Codecs returns a fresh copy of the registered codecs, in construction order,
// for building the task decoder from the same reviewed set.
func (r *Registry) Codecs() []taskdecode.TestCodec {
	if r == nil {
		return nil
	}
	codecs := make([]taskdecode.TestCodec, len(r.entries))
	for i, e := range r.entries {
		codecs[i] = e.Codec
	}
	return codecs
}
