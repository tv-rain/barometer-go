package task

import (
	"net/netip"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

// These are finite implementation-local collection bounds, not public maxima or
// authority grants. Changing them requires no change to serialized task schemas.
const maxPolicyEntries = 256

type Risk uint8

const (
	LowRisk Risk = iota + 1
	MediumRisk
	HighRisk
	UnknownRisk // unordered; permission requires an explicit separate opt-in
)

type Paths struct{ Direct, Proxy bool }

func (p Paths) valid() bool { return p.Direct || p.Proxy }

type TestVersion struct{ ID, Version string }

var policyTestID = regexp.MustCompile(`^[A-Z]{2,4}-[0-9]{2}$`)
var policyVersion = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

func (t TestVersion) valid() bool {
	return policyTestID.MatchString(t.ID) && len(t.Version) <= 64 && policyVersion.MatchString(t.Version)
}

// PolicyLimits is an explicit ceiling, with no permissive zero-value defaults.
// Zero traffic ceilings permit local-only behavior; they do not authorize a
// network test. Test-specific/protocol maxima remain mandatory downstream.
type PolicyLimits struct {
	Attempt                                             AttemptLimits
	MaxRuns, MaxAttempts, MaxValiditySeconds            int
	MinRepeatIntervalSeconds, MinRepeatJitterSeconds    int
	MinRetryDelayMS, MaxRetryDelayMS, MaxRetryJitterMS  int
	MaxResultBytes, MaxBundleBytes, MaxResultsPerBundle int
	MaxPendingBundles, MaxPendingBytes                  int
	MaxRetentionSeconds                                 int
}

func (l PolicyLimits) valid() bool {
	return between(l.Attempt.DurationMS, 1, 60000) && between(l.Attempt.BytesSent, 0, 65536) && between(l.Attempt.BytesReceived, 0, 262144) && between(l.Attempt.Connections, 0, 12) &&
		between(l.MaxRuns, 1, 28) && between(l.MaxAttempts, 1, 2) && between(l.MaxValiditySeconds, 1, 604800) &&
		between(l.MinRepeatIntervalSeconds, 21600, 604800) && between(l.MinRepeatJitterSeconds, 300, 1800) && l.MinRepeatJitterSeconds <= l.MinRepeatIntervalSeconds/4 &&
		between(l.MinRetryDelayMS, 30000, 300000) && between(l.MaxRetryDelayMS, l.MinRetryDelayMS, 600000) && between(l.MaxRetryJitterMS, 0, 60000) &&
		between(l.MaxResultBytes, 1, 65536) && between(l.MaxBundleBytes, 1, 262144) && between(l.MaxResultsPerBundle, 1, 16) &&
		between(l.MaxPendingBundles, 1, 32) && between(l.MaxPendingBytes, 1, 4194304) && between(l.MaxRetentionSeconds, 1, 86400)
}
func between(v, min, max int) bool { return v >= min && v <= max }

// Permissions are used independently for deployment policy and each issuer.
// Consumers must intersect these and all registered/task/platform restrictions;
// a valid model alone never establishes effective permission.
type Permissions struct {
	trustedOnly
	Tests                                           []TestVersion
	RegistrationIDs                                 []string
	Paths                                           Paths
	MaxRisk                                         Risk
	RequireConsent, AllowMediumRisk, AllowComposite bool
	AllowUnknownRisk                                bool     // reviewed exception only; unknown is never below MaxRisk
	LabelNamespaces                                 []string // exact reviewed prefixes; empty permits no labels
	Limits                                          PolicyLimits
}

func (p Permissions) valid() bool {
	if len(p.Tests) == 0 || len(p.Tests) > maxPolicyEntries || !unique(p.Tests) || !ids(p.RegistrationIDs) || !p.Paths.valid() || !p.Limits.valid() || p.MaxRisk < LowRisk || p.MaxRisk > HighRisk || len(p.LabelNamespaces) > 16 || !unique(p.LabelNamespaces) {
		return false
	}
	if p.MaxRisk == LowRisk && p.AllowMediumRisk {
		return false
	}
	for _, t := range p.Tests {
		if !t.valid() || t.ID == "NET-05" && !p.AllowComposite {
			return false
		}
	}
	for _, s := range p.LabelNamespaces {
		if !identifier(s) {
			return false
		}
	}
	return true
}
func copyPermissions(p Permissions) Permissions {
	p.Tests = slices.Clone(p.Tests)
	p.RegistrationIDs = slices.Clone(p.RegistrationIDs)
	p.LabelNamespaces = slices.Clone(p.LabelNamespaces)
	return p
}

type IssuerPermissions struct {
	trustedOnly
	IssuerID    string
	Permissions Permissions
}

type TargetRole uint8

const (
	MeasurementTarget TargetRole = iota + 1
	ResolverTarget
	ValidationTarget
	RedirectTarget
)

// TargetTuple holds exact canonical values, never wildcard/suffix rules. Absent
// properties remain absent; constructors do not fill, resolve or rewrite them.
// A test-owned validator must still require its complete safety-relevant tuple.
type TargetTuple struct {
	trustedOnly
	Hostname                    string
	Address                     netip.Addr
	Port                        uint16
	Scheme                      string
	URL                         string
	ResolverName, TLSServerName string
	BootstrapAddresses          []netip.Addr
	ALPN                        []string
	ExpectedStatus              int // zero means no status check
	BodyChecks                  []BodyCheck
	HeaderChecks                []HeaderCheck
}
type BodyCheck struct{ Type, Value string }
type HeaderCheck struct{ Name, Type, Value string }

type TargetRegistration struct {
	trustedOnly
	ID               string
	Test             TestVersion
	Role             TargetRole
	Tuple            TargetTuple
	Paths            Paths
	BaselineRisk     Risk
	Owner, Purpose   string
	ReviewGeneration uint64
	ExpiresAt        Timestamp
	RedirectIDs      []string
}

// PolicyInput is supplied by trusted initialization, separately from config JSON.
type PolicyInput struct {
	trustedOnly
	Generation  uint64
	Permissions Permissions
	Issuers     []IssuerPermissions
	Targets     []TargetRegistration
	// Keys are the trusted issuer keys for signed remote task packages
	// (REQ-TRUST-0051). A policy without keys admits no remote task.
	Keys []TrustedKey
}

// LocalPolicy owns a complete immutable snapshot. Constructors check structural
// invariants only: review/authentication, current time and monotonic activation
// must be established by the future trusted policy/admission owners.
type LocalPolicy struct {
	trustedOnly
	valid bool
	input PolicyInput
}

func NewLocalPolicy(in PolicyInput) (LocalPolicy, error) {
	if in.Generation == 0 || !in.Permissions.valid() || len(in.Issuers) == 0 || len(in.Issuers) > maxPolicyEntries || len(in.Targets) > maxPolicyEntries {
		return LocalPolicy{}, ErrTrustModel
	}
	issuers := make(map[string]bool, len(in.Issuers))
	targets := make(map[string]TargetRegistration, len(in.Targets))
	for _, t := range in.Targets {
		if !t.valid() || targets[t.ID].ID != "" {
			return LocalPolicy{}, ErrTrustModel
		}
		targets[t.ID] = t
	}
	for _, i := range in.Issuers {
		if !identifier(i.IssuerID) || issuers[i.IssuerID] || !i.Permissions.valid() {
			return LocalPolicy{}, ErrTrustModel
		}
		issuers[i.IssuerID] = true
		if !references(i.Permissions, targets) {
			return LocalPolicy{}, ErrTrustModel
		}
	}
	if !references(in.Permissions, targets) {
		return LocalPolicy{}, ErrTrustModel
	}
	if !validKeys(in.Keys, issuers) {
		return LocalPolicy{}, ErrTrustModel
	}
	for _, t := range in.Targets {
		for _, id := range t.RedirectIDs {
			r, ok := targets[id]
			if !ok || r.Role != RedirectTarget || r.Test != t.Test {
				return LocalPolicy{}, ErrTrustModel
			}
		}
	}
	return LocalPolicy{valid: true, input: copyPolicy(in)}, nil
}
func (p LocalPolicy) Valid() bool              { return p.valid }
func (p LocalPolicy) Generation() uint64       { return p.input.Generation }
func (p LocalPolicy) Permissions() Permissions { return copyPermissions(p.input.Permissions) }
func (p LocalPolicy) Issuer(id string) (IssuerPermissions, bool) {
	for _, i := range p.input.Issuers {
		if i.IssuerID == id {
			i.Permissions = copyPermissions(i.Permissions)
			return i, true
		}
	}
	return IssuerPermissions{}, false
}

// TrustedKey returns the trusted issuer key named id, by exact key ID.
func (p LocalPolicy) TrustedKey(id string) (TrustedKey, bool) {
	for _, k := range p.input.Keys {
		if k.KeyID == id {
			return k, true
		}
	}
	return TrustedKey{}, false
}

func (p LocalPolicy) Target(id string) (TargetRegistration, bool) {
	for _, t := range p.input.Targets {
		if t.ID == id {
			return copyTarget(t), true
		}
	}
	return TargetRegistration{}, false
}
func references(p Permissions, targets map[string]TargetRegistration) bool {
	for _, id := range p.RegistrationIDs {
		t, ok := targets[id]
		if !ok || !slices.Contains(p.Tests, t.Test) {
			return false
		}
	}
	return true
}
func (t TargetRegistration) valid() bool {
	return identifier(t.ID) && t.Test.valid() && t.Role >= MeasurementTarget && t.Role <= RedirectTarget && t.Paths.valid() && t.BaselineRisk >= LowRisk && t.BaselineRisk <= UnknownRisk &&
		category(t.Owner, 256) && category(t.Purpose, 256) && t.ReviewGeneration > 0 && trustTimestamp(t.ExpiresAt) && ids(t.RedirectIDs) && t.Tuple.valid()
}
func (t TargetTuple) valid() bool {
	if t.Hostname == "" && !t.Address.IsValid() && t.URL == "" {
		return false
	}
	for _, h := range []string{t.Hostname, t.ResolverName, t.TLSServerName} {
		if h != "" && !canonicalHost(h) {
			return false
		}
	}
	if t.Address.Zone() != "" || t.Address.Is4In6() || len(t.BootstrapAddresses) > 16 || !unique(t.BootstrapAddresses) || len(t.ALPN) > 8 || !unique(t.ALPN) || len(t.BodyChecks) > 8 || len(t.HeaderChecks) > 8 {
		return false
	}
	for _, a := range t.BootstrapAddresses {
		if !a.IsValid() || a.Zone() != "" || a.Is4In6() {
			return false
		}
	}
	for _, a := range t.ALPN {
		if !category(a, 32) {
			return false
		}
	}
	if t.Scheme != "" && t.Scheme != "https" && t.Scheme != "tls" && t.Scheme != "tcp" && t.Scheme != "udp" {
		return false
	}
	if t.ExpectedStatus != 0 && !between(t.ExpectedStatus, 100, 599) {
		return false
	}
	if t.URL != "" {
		if len(t.URL) > 2048 || !utf8.ValidString(t.URL) {
			return false
		}
		u, err := url.Parse(t.URL)
		if err != nil || len(t.URL) > 2048 || u.Scheme != "https" || u.User != nil || u.Fragment != "" || u.Host == "" || u.Opaque != "" || !canonicalHost(u.Hostname()) || u.String() != t.URL || strings.ContainsAny(t.URL, "\r\n\t ") || strings.ContainsAny(t.URL, "{}") {
			return false
		}
		if t.Scheme != "" && t.Scheme != "https" || t.Hostname != "" && t.Hostname != u.Hostname() {
			return false
		}
		port := 443
		if u.Port() != "" {
			n, err := strconv.Atoi(u.Port())
			if err != nil || n < 1 || n > 65535 || strconv.Itoa(n) != u.Port() {
				return false
			}
			port = n
		}
		if strings.HasSuffix(u.Host, ":") || t.Port != 0 && int(t.Port) != port {
			return false
		}
		if a, err := netip.ParseAddr(u.Hostname()); err == nil && t.Address.IsValid() && a != t.Address {
			return false
		}
	}
	for _, c := range t.BodyChecks {
		switch c.Type {
		case "contains", "prefix", "suffix":
			if !boundedText(c.Value, 256) {
				return false
			}
		case "sha256":
			if len(c.Value) != 64 || strings.Trim(c.Value, "0123456789abcdef") != "" {
				return false
			}
		default:
			return false
		}
	}
	for _, c := range t.HeaderChecks {
		if !headerName(c.Name) {
			return false
		}
		switch c.Type {
		case "present", "absent":
			if c.Value != "" {
				return false
			}
		case "equals", "contains":
			if !boundedText(c.Value, 256) || strings.ContainsAny(c.Value, "\r\n\x00") {
				return false
			}
		default:
			return false
		}
	}
	return true
}
func canonicalHost(s string) bool {
	if len(s) == 0 || len(s) > 253 {
		return false
	}
	if a, err := netip.ParseAddr(s); err == nil {
		return a.Zone() == "" && !a.Is4In6() && a.String() == s
	}
	if len(s) == 0 || len(s) > 253 || strings.HasSuffix(s, ".") {
		return false
	}
	for _, label := range strings.Split(s, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, c := range label {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				return false
			}
		}
	}
	return true
}
func headerName(s string) bool {
	if len(s) == 0 || len(s) > 64 || strings.ToLower(s) != s {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || strings.ContainsRune("!#$%&'*+-.^_`|~", c)) {
			return false
		}
	}
	switch s {
	case "authorization", "proxy-authorization", "cookie", "set-cookie", "www-authenticate", "proxy-authenticate":
		return false
	}
	return true
}
func category(s string, max int) bool {
	if len(s) == 0 || len(s) > max {
		return false
	}
	for _, c := range s {
		if c < 32 || c > 126 || c == '{' || c == '}' {
			return false
		}
	}
	return true
}

func boundedText(s string, max int) bool { return len(s) > 0 && len(s) <= max && utf8.ValidString(s) }
func ids(v []string) bool {
	if len(v) > maxPolicyEntries || !unique(v) {
		return false
	}
	for _, s := range v {
		if !identifier(s) {
			return false
		}
	}
	return true
}
func unique[T comparable](v []T) bool {
	seen := make(map[T]bool, len(v))
	for _, x := range v {
		if seen[x] {
			return false
		}
		seen[x] = true
	}
	return true
}
func copyTarget(t TargetRegistration) TargetRegistration {
	t.RedirectIDs = slices.Clone(t.RedirectIDs)
	t.Tuple.BootstrapAddresses = slices.Clone(t.Tuple.BootstrapAddresses)
	t.Tuple.ALPN = slices.Clone(t.Tuple.ALPN)
	t.Tuple.BodyChecks = slices.Clone(t.Tuple.BodyChecks)
	t.Tuple.HeaderChecks = slices.Clone(t.Tuple.HeaderChecks)
	return t
}
func copyPolicy(in PolicyInput) PolicyInput {
	in.Permissions = copyPermissions(in.Permissions)
	in.Issuers = slices.Clone(in.Issuers)
	in.Targets = slices.Clone(in.Targets)
	in.Keys = slices.Clone(in.Keys)
	for n := range in.Issuers {
		in.Issuers[n].Permissions = copyPermissions(in.Issuers[n].Permissions)
	}
	for n := range in.Targets {
		in.Targets[n] = copyTarget(in.Targets[n])
	}
	return in
}

// RevocationInput names principals/registrations only. It deliberately defines no
// wire format, activation time, cancellation, replay or persistence semantics.
// Operational revocation remains an open public/internal review question.
type RevocationInput struct {
	trustedOnly
	PolicyGeneration           uint64
	IssuerIDs, RegistrationIDs []string
}
type Revocations struct {
	trustedOnly
	valid bool
	input RevocationInput
}

func NewRevocations(in RevocationInput) (Revocations, error) {
	if in.PolicyGeneration == 0 || !ids(in.IssuerIDs) || !ids(in.RegistrationIDs) || len(in.IssuerIDs)+len(in.RegistrationIDs) == 0 {
		return Revocations{}, ErrTrustModel
	}
	in.IssuerIDs = slices.Clone(in.IssuerIDs)
	in.RegistrationIDs = slices.Clone(in.RegistrationIDs)
	return Revocations{valid: true, input: in}, nil
}
func (r Revocations) Valid() bool { return r.valid }
func (r Revocations) Input() RevocationInput {
	in := r.input
	in.IssuerIDs = slices.Clone(in.IssuerIDs)
	in.RegistrationIDs = slices.Clone(in.RegistrationIDs)
	return in
}

// Equal compares every normalized safety-relevant tuple component.
func (a TargetTuple) Equal(b TargetTuple) bool {
	return a.Hostname == b.Hostname && a.Address == b.Address && a.Port == b.Port && a.Scheme == b.Scheme && a.URL == b.URL &&
		a.ResolverName == b.ResolverName && a.TLSServerName == b.TLSServerName && a.ExpectedStatus == b.ExpectedStatus &&
		slices.Equal(a.BootstrapAddresses, b.BootstrapAddresses) && slices.Equal(a.ALPN, b.ALPN) && slices.Equal(a.BodyChecks, b.BodyChecks) && slices.Equal(a.HeaderChecks, b.HeaderChecks)
}

// Clone returns owned copies of the registration's collection values.
func (t TargetRegistration) Clone() TargetRegistration { return copyTarget(t) }
