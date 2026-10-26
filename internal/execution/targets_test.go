package execution

import (
	"context"
	"errors"
	"io"
	"net/netip"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/task"
)

type targetClock struct{ now time.Time }

func (c *targetClock) Now() time.Time { return c.now }

type targetResolver struct {
	answers []netip.Addr
	calls   int
	during  func(context.Context)
	err     error
}

func (r *targetResolver) LookupNetIP(ctx context.Context, _, _ string) ([]netip.Addr, error) {
	r.calls++
	if r.during != nil {
		r.during(ctx)
	}
	return r.answers, r.err
}
func targetRegistration(id, host string, role task.TargetRole, now time.Time) task.TargetRegistration {
	return task.TargetRegistration{ID: id, Test: task.TestVersion{ID: "HTTP-01", Version: "0.1.0"}, Role: role, Tuple: task.TargetTuple{Hostname: host, Port: 443, Scheme: "https", URL: "https://" + host + "/check"}, Paths: task.Paths{Direct: true, Proxy: true}, ExpiresAt: task.TimestampFromTime(now.Add(time.Hour))}
}
func guard(path string, regs ...task.TargetRegistration) (*Targets, *targetClock) {
	c := &targetClock{time.Now()}
	return NewTargets(task.TestVersion{ID: "HTTP-01", Version: "0.1.0"}, path, regs, c, nil), c
}

// Synthetic production-gate evidence for the target-validation rows of
// CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION. No external contact is made.
// requirement-ref: REQ-TRUST-0016
// requirement-ref: REQ-PROXY-0037
func TestTargetGateResolutionAndRegistration(t *testing.T) {
	now := time.Now()
	reg := targetRegistration("initial", "target.example", task.MeasurementTarget, now)
	for _, tc := range []struct {
		name    string
		answers []netip.Addr
		deny    bool
	}{
		{"allowed", []netip.Addr{netip.MustParseAddr("8.8.8.8"), netip.MustParseAddr("9.9.9.9")}, false},
		{"prohibited", []netip.Addr{netip.MustParseAddr("127.0.0.1")}, true},
		{"mixed", []netip.Addr{netip.MustParseAddr("8.8.8.8"), netip.MustParseAddr("10.0.0.1")}, true},
		{"empty", nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, _ := guard("direct", reg)
			r := &targetResolver{answers: tc.answers}
			g.resolver = r
			starts := 0
			got, err := g.Prepare(context.Background(), "tcp", "target.example:443", func() { starts++ })
			if (err != nil) != tc.deny || r.calls != 1 || starts != 1 {
				t.Fatalf("authority=%q err=%v resolver=%d starts=%d", got, err, r.calls, starts)
			}
			if !tc.deny && got != "8.8.8.8:443" {
				t.Fatal("did not pin the checked IP")
			}
		})
	}
	for _, mutate := range []func(*task.TargetRegistration){
		func(r *task.TargetRegistration) {
			r.Tuple.URL = "https://other.example/check"
			r.Tuple.Hostname = "other.example"
		},
		func(r *task.TargetRegistration) { r.ExpiresAt = task.TimestampFromTime(now.Add(-time.Second)) },
		func(r *task.TargetRegistration) { r.Paths = task.Paths{Proxy: true} },
		func(r *task.TargetRegistration) { r.Test.Version = "0.2.0" },
	} {
		q := reg.Clone()
		mutate(&q)
		g, _ := guard("direct", q)
		r := &targetResolver{answers: []netip.Addr{netip.MustParseAddr("8.8.8.8")}}
		g.resolver = r
		_, err := g.Prepare(context.Background(), "tcp", "target.example:443", nil)
		if err == nil || r.calls != 0 {
			t.Fatal("registration rejection resolved or authorized a target")
		}
	}
	// A registration expires while the resolver works: never hand the IP to transport.
	g, c := guard("direct", reg)
	g.resolver = &targetResolver{answers: []netip.Addr{netip.MustParseAddr("8.8.8.8")}, during: func(context.Context) { c.now = now.Add(2 * time.Hour) }}
	if _, err := g.Prepare(context.Background(), "tcp", "target.example:443", nil); err == nil {
		t.Fatal("expiry during resolution bypassed")
	}
	// Cancellation is passed to resolution and prevents contact even if the resolver returns an answer.
	ctx, cancel := context.WithCancel(context.Background())
	g, _ = guard("direct", reg)
	g.resolver = &targetResolver{answers: []netip.Addr{netip.MustParseAddr("8.8.8.8")}, during: func(ctx context.Context) {
		cancel()
		if ctx.Err() == nil {
			t.Fatal("resolver context not cancelled")
		}
	}}
	if _, err := g.Prepare(ctx, "tcp", "target.example:443", nil); err == nil {
		t.Fatal("cancelled resolution authorized contact")
	}
}

// requirement-ref: REQ-PROXY-0043
// requirement-ref: REQ-PROXY-0046
func TestTargetGateProxyRefusesHostnameBeforeStartOrResolution(t *testing.T) {
	reg := targetRegistration("initial", "target.example", task.MeasurementTarget, time.Now())
	g, _ := guard("proxy", reg)
	r := &targetResolver{}
	g.resolver = r
	started := false
	_, err := g.Prepare(context.Background(), "tcp", "target.example:443", func() { started = true })
	reason, ok := pathFailure(err)
	if !ok || reason != task.ProxyPathUnsupported || started || r.calls != 0 {
		t.Fatalf("reason %v starts %v resolutions %d", reason, started, r.calls)
	}
	// Missing authority cannot be supplied by a runner or a transport.
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	d := &recordingTargetDialer{}
	s := newScope(ctx, d, task.AttemptLimits{Connections: 2}, cancel)
	defer s.close()
	s.targets = g
	if _, err := s.DialContext(context.Background(), "tcp", "target.example:443"); err == nil || s.Started() || d.calls != 0 || context.Cause(ctx) == nil {
		t.Fatal("hostname refusal was bypassed")
	}
	// The same refusal cannot be hidden by a runner returning a terminal outcome.
	if reason, ok := cutShort(ctx); !ok || reason != task.ProxyPathUnsupported {
		t.Fatal("refusal did not terminate the attempt")
	}
}

type recordingTargetDialer struct {
	calls     int
	addresses []string
}

func (d *recordingTargetDialer) DialContext(_ context.Context, _, address string) (measurement.Conn, error) {
	d.calls++
	d.addresses = append(d.addresses, address)
	return &targetConn{}, nil
}

type targetConn struct{ closed bool }

func (c *targetConn) Read([]byte) (int, error)    { return 0, io.EOF }
func (c *targetConn) Write(p []byte) (int, error) { return len(p), nil }
func (c *targetConn) Close() error                { c.closed = true; return nil }
func (c *targetConn) SetDeadline(time.Time) error { return nil }

// requirement-ref: REQ-HTTP-01-0051
// requirement-ref: REQ-TRUST-0016
func TestTargetGateRedirectFullTupleAndSharedScope(t *testing.T) {
	now := time.Now()
	initial := targetRegistration("initial", "8.8.8.8", task.MeasurementTarget, now)
	initial.RedirectIDs = []string{"hop"}
	hop := targetRegistration("hop", "9.9.9.9", task.RedirectTarget, now)
	hop.RedirectIDs = []string{"last"}
	last := targetRegistration("last", "1.1.1.1", task.RedirectTarget, now)
	for _, path := range []string{"direct", "proxy"} {
		t.Run(path, func(t *testing.T) {
			g, _ := guard(path, initial, hop, last)
			// Mutating the caller's copy cannot expand the owned snapshot.
			altered := hop.Tuple
			altered.URL = "https://9.9.9.9/unregistered"
			if _, err := g.Redirect(altered); !errors.Is(err, ErrTargetRejected) {
				t.Fatal("same-authority URL bypassed full tuple matching")
			}
			ctx, cancel := context.WithCancelCause(context.Background())
			defer cancel(nil)
			raw := &recordingTargetDialer{}
			s := newScope(ctx, raw, task.AttemptLimits{Connections: 2, BytesSent: 4, BytesReceived: 4}, cancel)
			defer s.close()
			s.targets = g
			// Root dialer cannot directly contact an authorized redirect registration.
			if _, err := g.Prepare(ctx, "tcp", "9.9.9.9:443", nil); err == nil {
				t.Fatal("redirect bypassed its link")
			}
			d, err := s.Redirect(hop.Tuple)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := d.DialContext(ctx, "tcp", "8.8.8.8:443"); !errors.Is(err, ErrTargetRejected) {
				t.Fatal("redirect dialer escaped its narrowed authority")
			}
			conn, err := d.DialContext(ctx, "tcp", "9.9.9.9:443")
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			next, err := d.Redirect(last.Tuple)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := next.DialContext(ctx, "tcp", "1.1.1.1:443"); err != nil {
				t.Fatal(err)
			}
			if raw.calls != 2 || s.Usage().Connections != 2 {
				t.Fatal("redirect hops did not share the meter")
			}
			if _, err := conn.Write([]byte("12345")); err == nil {
				t.Fatal("redirect escaped byte budget")
			}
		})
	}
	// Registered hostname redirects are still forbidden over the proxy path.
	hostname := targetRegistration("hop", "next.example", task.RedirectTarget, now)
	g, _ := guard("proxy", initial, hostname)
	if _, err := g.Redirect(hostname.Tuple); !errors.Is(err, ErrTargetRejected) {
		t.Fatal("registered hostname redirect authorized")
	}
	// Prohibited literal redirect is a local redirect refusal, not TCP failure or outer cancellation.
	bad := targetRegistration("hop", "127.0.0.1", task.RedirectTarget, now)
	g, _ = guard("proxy", initial, bad)
	child, err := g.Redirect(bad.Tuple)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := child.Prepare(context.Background(), "tcp", "127.0.0.1:443", nil); !errors.Is(err, ErrTargetRejected) {
		t.Fatal("prohibited redirect not safely refused")
	}
}

// requirement-ref: REQ-TCP-01-0012
func TestDefaultAddressPolicy(t *testing.T) {
	for _, literal := range []string{"0.0.0.0", "10.0.0.1", "100.64.0.1", "127.0.0.1", "169.254.1.1", "172.16.1.1", "192.0.0.9", "192.0.2.1", "192.168.1.1", "198.18.0.1", "198.51.100.1", "203.0.113.1", "224.0.0.1", "255.255.255.255", "::", "::1", "::ffff:127.0.0.1", "64:ff9b::a00:1", "2001:db8::1", "fc00::1", "fe80::1", "ff02::1", "3fff::1"} {
		if (PublicAddresses{}).Allows(netip.MustParseAddr(literal)) {
			t.Fatalf("prohibited class authorized: %s", literal)
		}
	}
	for _, literal := range []string{"8.8.8.8", "9.9.9.9", "2606:4700:4700::1111"} {
		if !(PublicAddresses{}).Allows(netip.MustParseAddr(literal)) {
			t.Fatal("public control rejected")
		}
	}
	if strings.Contains(ErrTargetRejected.Error(), "target.example") {
		t.Fatal("error disclosed target")
	}
}

// The scope cannot contact a transport with mixed DNS answers, and a zero
// connection budget cannot even invoke the resolver.
// requirement-ref: REQ-TRUST-0016
// requirement-ref: REQ-LIMIT-0007
func TestTargetScopeResolutionCannotBypassBounds(t *testing.T) {
	for _, connections := range []int{0, 1} {
		reg := targetRegistration("initial", "target.example", task.MeasurementTarget, time.Now())
		g, _ := guard("direct", reg)
		resolver := &targetResolver{answers: []netip.Addr{netip.MustParseAddr("8.8.8.8"), netip.MustParseAddr("127.0.0.1")}}
		g.resolver = resolver
		ctx, cancel := context.WithCancelCause(context.Background())
		raw := &recordingTargetDialer{}
		s := newScope(ctx, raw, task.AttemptLimits{Connections: connections}, cancel)
		s.targets = g
		if _, err := s.DialContext(context.Background(), "tcp", "target.example:443"); err == nil {
			t.Fatal("unsafe resolution succeeded")
		}
		if raw.calls != 0 || s.Usage().Connections != 0 || resolver.calls != connections {
			t.Fatalf("bound=%d raw=%d reserved=%d resolver=%d", connections, raw.calls, s.Usage().Connections, resolver.calls)
		}
		if connections == 0 && s.Started() {
			t.Fatal("local budget refusal started the test")
		}
		s.close()
		cancel(nil)
	}
}

// Constructor inputs and returned redirect chains must not alias caller data,
// and an ancestor expiring while a hop is pending cannot keep granting access.
// requirement-ref: REQ-TRUST-0016
func TestTargetSnapshotOwnershipAndAncestorExpiry(t *testing.T) {
	now := time.Now()
	initial := targetRegistration("initial", "8.8.8.8", task.MeasurementTarget, now)
	initial.RedirectIDs = []string{"hop"}
	hop := targetRegistration("hop", "9.9.9.9", task.RedirectTarget, now)
	hop.Tuple.ALPN = []string{"http/1.1"}
	expected := hop.Clone()
	g, c := guard("proxy", initial, hop)
	initial.RedirectIDs[0] = "unrelated"
	hop.Tuple.ALPN[0] = "mutated"
	child, err := g.Redirect(expected.Tuple)
	if err != nil {
		t.Fatal("caller mutated owned snapshot", err)
	}
	c.now = now.Add(2 * time.Hour)
	if _, err := child.Prepare(context.Background(), "tcp", "9.9.9.9:443", nil); !errors.Is(err, ErrTargetRejected) {
		t.Fatal("expired chain retained authority")
	}
}

type denyAddresses struct{}

func (denyAddresses) Allows(netip.Addr) bool { return false }

// An endpoint approved by configuration is still subject to the same stricter
// local address policy as the attempt, before the context transport is called.
// requirement-ref: REQ-PROXY-0045
func TestContextFetchUsesLocalAddressPolicy(t *testing.T) {
	raw := &recordingTargetDialer{}
	u, err := url.Parse("https://8.8.8.8/network-context")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := contextFetcher(raw, denyAddresses{})(context.Background(), u, nil); err == nil || raw.calls != 0 {
		t.Fatal("context bypassed local address restriction")
	}
}

// Name-resolution failure stays distinguishable to a protocol runner without
// exposing the resolver's raw error or target details.
func TestTargetResolutionFailureIsCoarse(t *testing.T) {
	reg := targetRegistration("initial", "target.example", task.MeasurementTarget, time.Now())
	g, _ := guard("direct", reg)
	g.resolver = &targetResolver{err: errors.New("RAW-RESOLVER-DETAIL-target.example")}
	_, err := g.Prepare(context.Background(), "tcp", "target.example:443", nil)
	if !errors.Is(err, measurement.ErrResolutionFailed) || strings.Contains(err.Error(), "target.example") {
		t.Fatal("resolver failure was hidden or disclosed")
	}
}
