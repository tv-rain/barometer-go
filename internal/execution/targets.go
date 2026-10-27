package execution

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/task"
)

// ErrTargetRejected carries no target, registration, address or raw error.
var ErrTargetRejected = measurement.ErrTargetRejected

// AddressPolicy is trusted local policy, never task or runner input. The
// default rejects special-purpose addresses; a reviewed integration may impose
// stricter restrictions. Implementations must be immutable and safe for
// concurrent calls. Synthetic test adapters alone may allow loopback fixtures.
// There is no serialized or exported facade input.
type AddressPolicy interface{ Allows(netip.Addr) bool }

// TargetCheck is the execution-owned per-contact gate. Prepare returns only a
// checked literal IP authority, so a direct transport cannot resolve again.
// start is called before network resolution, never for a local refusal.
type TargetCheck interface {
	Prepare(context.Context, string, string, func()) (string, error)
	Redirect(task.TargetTuple) (TargetCheck, error)
}

// SystemResolver performs an ordinary per-name platform lookup. Implementors
// honor cancellation and retain no Barometer-owned query/result state or
// unrelated resolver activity. Native behavior needs separate validation.
type SystemResolver interface {
	LookupNetIP(context.Context, string, string) ([]netip.Addr, error)
}

// Targets owns only admitted registrations and their redirect links. Each
// redirect child narrows to its full normalized tuple and shares no mutable
// registration state. Expiry/path/address checks are repeated for every dial.
type Targets struct {
	redirect      bool
	test          task.TestVersion
	proxy         bool
	registrations []task.TargetRegistration
	current       []string
	ancestors     []string
	clock         interface{ Now() time.Time }
	policy        AddressPolicy
	resolver      SystemResolver
}

func NewTargets(test task.TestVersion, path string, registrations []task.TargetRegistration, clock interface{ Now() time.Time }, policy AddressPolicy) *Targets {
	t := &Targets{test: test, proxy: path == "proxy", clock: clock, policy: policy, resolver: net.DefaultResolver}
	if policy == nil {
		t.policy = PublicAddresses{}
	}
	if path != "direct" && path != "proxy" || len(registrations) > 256 {
		return t
	}
	for _, r := range registrations {
		t.registrations = append(t.registrations, r.Clone())
		if r.Role != task.RedirectTarget {
			t.current = append(t.current, r.ID)
		}
	}
	return t
}

func (t *Targets) permitted(r task.TargetRegistration) bool {
	return t != nil && t.clock != nil && !t.clock.Now().IsZero() && r.Test == t.test && r.ExpiresAt.AfterTime(t.clock.Now()) &&
		(t.proxy && r.Paths.Proxy || !t.proxy && r.Paths.Direct)
}

// destination derives only connection authority. Query names and TLS SNI are
// not destinations. Complete URL/check/ALPN matching happens at admission or
// at Redirect, not by comparing just a hostname here.
func destination(q task.TargetTuple) (string, uint16) {
	if q.Address.IsValid() {
		return q.Address.String(), q.Port
	}
	if q.URL != "" {
		u, err := url.Parse(q.URL)
		if err != nil || u.Scheme != "https" && u.Scheme != "http" || u.User != nil || u.Fragment != "" {
			return "", 0
		}
		port := 443
		if u.Scheme == "http" {
			port = 80
		}
		if u.Port() != "" {
			port, err = strconv.Atoi(u.Port())
			if err != nil || port < 1 || port > 65535 {
				return "", 0
			}
		}
		return u.Hostname(), uint16(port)
	}
	return q.Hostname, q.Port
}

func (t *Targets) matches(r task.TargetRegistration, network, host string, port uint16) bool {
	if !t.permitted(r) || r.Tuple.Scheme == "udp" && !strings.HasPrefix(network, "udp") || (r.Tuple.Scheme == "tcp" || r.Tuple.Scheme == "tls" || r.Tuple.Scheme == "https" || r.Tuple.Scheme == "http") && !strings.HasPrefix(network, "tcp") {
		return false
	}
	h, p := destination(r.Tuple)
	if p != port {
		return false
	}
	if len(r.Tuple.BootstrapAddresses) > 0 {
		a, err := netip.ParseAddr(host)
		return err == nil && slices.Contains(r.Tuple.BootstrapAddresses, a)
	}
	return h == host
}

// Prepare authorizes before contact, then validates the complete bounded
// direct-resolution set and pins the chosen address. A mixed unsafe set is
// refused atomically; no filtering, retry or second resolution is hidden here.
// requirement-ref: REQ-TRUST-0016
// requirement-ref: REQ-PROXY-0043
// requirement-ref: REQ-PROXY-0046
func (t *Targets) Prepare(ctx context.Context, network, address string, start func()) (string, error) {
	reject := func() (string, error) {
		if t != nil && t.redirect {
			return "", ErrTargetRejected
		}
		return "", pathError{task.LocalPolicyRejected}
	}
	if t == nil || t.clock == nil || t.policy == nil || !t.validChain() || ctx.Err() != nil {
		return reject()
	}
	if network != "tcp" && network != "tcp4" && network != "tcp6" && network != "udp" && network != "udp4" && network != "udp6" {
		return reject()
	}
	host, ps, err := net.SplitHostPort(address)
	if err != nil {
		return reject()
	}
	port, err := strconv.Atoi(ps)
	if err != nil || port < 1 || port > 65535 || strconv.Itoa(port) != ps {
		return reject()
	}
	ip, iperr := netip.ParseAddr(host)
	if t.proxy && iperr != nil {
		return "", pathError{task.ProxyPathUnsupported}
	}
	allowed := false
	for _, r := range t.registrations {
		if slices.Contains(t.current, r.ID) && t.matches(r, network, host, uint16(port)) {
			allowed = true
			break
		}
	}
	if !allowed {
		return reject()
	}
	ips := []netip.Addr{ip}
	if iperr != nil {
		if t.resolver == nil {
			return reject()
		}
		if start != nil {
			start()
		}
		family := "ip"
		if network == "tcp4" || network == "udp4" {
			family = "ip4"
		}
		if network == "tcp6" || network == "udp6" {
			family = "ip6"
		}
		ips, err = t.resolver.LookupNetIP(ctx, family, host)
		if err != nil {
			return "", errResolution
		}
	}
	if len(ips) == 0 || len(ips) > 64 || ctx.Err() != nil {
		return reject()
	}
	for _, a := range ips {
		if !a.IsValid() || a.Zone() != "" || !t.policy.Allows(a.Unmap()) ||
			(network == "tcp4" || network == "udp4") && !a.Unmap().Is4() || (network == "tcp6" || network == "udp6") && !a.Unmap().Is6() {
			return reject()
		}
	}
	// Re-check expiry immediately before handing the checked address to the
	// transport; resolver work may have advanced trusted time.
	allowed = false
	for _, r := range t.registrations {
		if slices.Contains(t.current, r.ID) && t.matches(r, network, host, uint16(port)) {
			allowed = true
			break
		}
	}
	if !allowed || !t.validChain() || ctx.Err() != nil {
		return reject()
	}
	return net.JoinHostPort(ips[0].Unmap().String(), ps), nil
}

// errResolution is coarse and contains no resolver/hostname detail. A runner
// maps this protocol-resolution failure to its registered observation.
var errResolution = measurement.ErrResolutionFailed

// Redirect validates the entire tuple and the link from the current target.
// Rejection does not cancel the attempt: HTTP records redirect_rejected.
// requirement-ref: REQ-HTTP-01-0051
func (t *Targets) Redirect(tuple task.TargetTuple) (TargetCheck, error) {
	if t == nil || !t.validChain() {
		return nil, ErrTargetRejected
	}
	host, _ := destination(tuple)
	if t.proxy {
		if _, err := netip.ParseAddr(host); err != nil {
			return nil, ErrTargetRejected
		}
	}
	for _, parent := range t.registrations {
		if !slices.Contains(t.current, parent.ID) || !t.permitted(parent) {
			continue
		}
		for _, r := range t.registrations {
			if r.Role == task.RedirectTarget && slices.Contains(parent.RedirectIDs, r.ID) && t.permitted(r) && r.Tuple.Equal(tuple) {
				child := *t
				child.current = []string{r.ID}
				child.ancestors = append(slices.Clone(t.ancestors), parent.ID)
				child.redirect = true
				return &child, nil
			}
		}
	}
	return nil, ErrTargetRejected
}

// PublicAddresses implements the initial prohibition of local, private,
// documentation, benchmarking and other special-purpose destinations. The
// static ranges follow the IANA IPv4/IPv6 special-purpose registries; no
// runtime lookup or downloaded policy is used. The table needs review when
// those registries change. A globally routable address is not target approval.
// requirement-ref: REQ-TCP-01-0012
type PublicAddresses struct{}

func (PublicAddresses) Allows(a netip.Addr) bool { return task.PublicAddress(a) }

func (Targets) String() string   { return "execution.Targets" }
func (Targets) GoString() string { return "execution.Targets" }

func (t *Targets) validChain() bool {
	if t == nil {
		return false
	}
	if t.test.ID == "DNS-01" {
		for _, r := range t.registrations {
			if slices.Contains(t.current, r.ID) && r.Role == task.MeasurementTarget && !t.permitted(r) {
				return false
			}
		}
	}
	for _, id := range t.ancestors {
		found := false
		for _, r := range t.registrations {
			if r.ID == id && t.permitted(r) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Lookup authorizes exactly the admitted query name again immediately before
// one ordinary system API lookup. Returned addresses are never contacted or
// retained. The API exposes no wire messages, so its traffic is not metered.
// Native resolver selection, cancellation and error fidelity need device proof.
// requirement-ref: REQ-DNS-01-0018
// requirement-ref: REQ-DNS-01-0030
// requirement-ref: REQ-LIMIT-0034
func (t *Targets) Lookup(ctx context.Context, name, query string, start func()) (bool, string, error) {
	reject := func() (bool, string, error) { return false, "", pathError{task.LocalPolicyRejected} }
	if t == nil || t.proxy || t.test != (task.TestVersion{ID: "DNS-01", Version: "0.1.0"}) || t.resolver == nil || !t.validChain() || ctx.Err() != nil || query != "A" && query != "AAAA" {
		return reject()
	}
	allowed := false
	for _, r := range t.registrations {
		if slices.Contains(t.current, r.ID) && r.Role == task.MeasurementTarget && r.Tuple.Hostname == name && r.Tuple.Port == 0 && t.permitted(r) {
			allowed = true
		}
	}
	if !allowed {
		return reject()
	}
	if start != nil {
		start()
	}
	family := "ip4"
	if query == "AAAA" {
		family = "ip6"
	}
	ips, err := t.resolver.LookupNetIP(ctx, family, name+".")
	if ctx.Err() != nil {
		return false, "", measurement.ErrRunner
	}
	if err != nil {
		var de *net.DNSError
		if errors.As(err, &de) {
			if de.IsTimeout {
				return false, "timeout", nil
			}
			if de.IsNotFound {
				// IsNotFound merges NXDOMAIN and no-data; the API cannot distinguish.
				return false, "inconclusive", nil
			}
		}
		return false, "resolver_error", nil
	}
	if len(ips) > 16 {
		return false, "malformed_response", nil
	}
	for _, ip := range ips {
		if !ip.IsValid() || ip.Zone() != "" || query == "A" && !ip.Unmap().Is4() || query == "AAAA" && !ip.Is6() {
			return false, "malformed_response", nil
		}
	}
	return len(ips) > 0, "", nil
}

// WithResolver returns a private copy bound to the trusted per-client resolver.
// Nil selects the ordinary Go system API; no task can supply this capability.
func (t *Targets) WithResolver(r SystemResolver) *Targets {
	if r == nil {
		return t
	}
	c := *t
	c.resolver = r
	return &c
}

// Step narrows the admitted union to one primitive's exact version. The path,
// clock, resolver and local address policy remain the attempt's original ones.
func (t *Targets) Step(test task.TestVersion, tuples []task.TargetTuple) TargetCheck {
	if t == nil || t.test.ID != "NET-05" || len(tuples) < 1 || len(tuples) > 2 {
		return nil
	}
	switch test.ID {
	case "DNS-01", "TCP-01", "TLS-01", "HTTP-01":
	default:
		return nil
	}
	if test.Version != "0.1.0" {
		return nil
	}
	child := *t
	child.test = test
	child.current = nil
	for _, tuple := range tuples {
		found := false
		for _, reg := range child.registrations {
			if reg.Test == test && reg.Role != task.RedirectTarget && reg.Tuple.Equal(tuple) {
				child.current = append(child.current, reg.ID)
				found = true
			}
		}
		if !found {
			return nil
		}
	}
	return &child
}
