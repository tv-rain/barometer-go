// Package registry holds the reviewed mapping from an exact registered test ID
// and version to the codec, baseline risk, target rule, proxy compatibility and
// composite eligibility admission needs, and to the runner execution invokes.
// It performs no I/O, holds no participant data and has no runtime
// registration. Presence of an entry is not evidence of released, conformant
// or field-validated support.
package registry

import (
	"errors"
	"net/netip"
	"net/url"

	"github.com/tv-rain/barometer-go/internal/measurement"
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

// Target is one destination or name a task would contact, with the role its
// local registration must have.
type Target struct {
	Role  task.TargetRole
	Tuple task.TargetTuple
}

// TargetRule is the test-owned extraction of every destination or name that
// normalized parameters of its test would contact, each as that test's
// complete safety-relevant tuple (REQ-TRUST-0014) in the normalized typed form
// registrations use. Implementations must be immutable, pure and bounded, must
// report false for parameters of an unexpected type or an incomplete tuple, and
// must not resolve, fetch or rewrite a target. A rule never grants authority:
// admission matches every tuple exactly against local registrations.
type TargetRule interface {
	Targets(task.Parameters) ([]Target, bool)
}

// NoTargets is the rule of a test with no remote target, such as NET-01, whose
// target is the local operating-system network-state API (REQ-NET-01-0005).
// It reports an empty list for any parameters. Admission accepts an empty
// list only from this rule; a test with this rule is never a NET-05 step.
type NoTargets struct{}

func (NoTargets) Targets(task.Parameters) ([]Target, bool) { return nil, true }

// Entry is one reviewed test version. Codec, Targets, Proxy and Runner must
// remain immutable for the registry's lifetime. Runner is optional so that an
// admission-only entry stays valid; execution fails closed without one.
type Entry struct {
	Test         task.TestVersion
	Codec        taskdecode.TestCodec
	BaselineRisk task.Risk
	Targets      TargetRule
	Proxy        ProxyRule
	Runner       measurement.Runner
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
// copies. An empty set is valid and resolves no test. Any invalid, incomplete
// or duplicate entry rejects the whole set.
func New(entries []Entry) (*Registry, error) {
	if len(entries) > maxEntries {
		return nil, ErrInvalid
	}
	r := &Registry{entries: make([]Entry, 0, len(entries)), index: make(map[task.TestVersion]int, len(entries))}
	codecs := make([]taskdecode.TestCodec, 0, len(entries))
	for _, e := range entries {
		if e.Codec == nil || e.Targets == nil || e.Proxy == nil || e.BaselineRisk < task.LowRisk || e.BaselineRisk > task.UnknownRisk {
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

// Runner returns the runner composed for the exact test ID and version. A
// missing entry or an entry without a runner reports false; there is no
// fallback to another version or a default runner.
func (r *Registry) Runner(test task.TestVersion) (measurement.Runner, bool) {
	e, ok := r.Lookup(test)
	if !ok || e.Runner == nil {
		return nil, false
	}
	return e.Runner, true
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

// ProxyCompatible combines the registered compatibility rule with the common
// explicit-IP-only destination rule. Query names and TLS SNI are not CONNECT
// authorities; redirect registrations are checked when a hop is followed.
// requirement-ref: REQ-PROXY-0044
func ProxyCompatible(rule ProxyRule, targets TargetRule, params task.Parameters) bool {
	if rule == nil || targets == nil || !rule.Compatible(params) {
		return false
	}
	ts, ok := targets.Targets(params)
	if !ok || len(ts) > 64 {
		// Extraction failure retains the target-registration failure class at admission.
		return true
	}
	for _, t := range ts {
		if t.Role == task.RedirectTarget {
			continue
		}
		q := t.Tuple
		if q.URL != "" && t.Role == task.MeasurementTarget {
			u, err := url.Parse(q.URL)
			if err != nil {
				return false
			}
			if a, err := netip.ParseAddr(u.Hostname()); err != nil || a.Zone() != "" {
				return false
			}
			continue
		}
		if len(q.BootstrapAddresses) > 0 || q.Address.IsValid() {
			continue
		}
		host := q.Hostname
		if q.URL != "" {
			u, err := url.Parse(q.URL)
			if err != nil {
				return false
			}
			host = u.Hostname()
		} else if q.Port == 0 {
			continue
		}
		ip, err := netip.ParseAddr(host)
		if err != nil || ip.Zone() != "" {
			return false
		}
	}
	return true
}
