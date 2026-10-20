package admission_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/admission"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/testkit"
	"github.com/tv-rain/barometer-go/internal/testkit/delivery"
)

var epoch = time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)

// fakeClock adapts the manually advanced testkit clock to platform.Clock and
// can be made unavailable.
type fakeClock struct {
	*testkit.Clock
	down atomic.Bool
}

func (c *fakeClock) Now(context.Context) (time.Time, error) {
	if c.down.Load() {
		return time.Time{}, errors.New("synthetic clock unavailable")
	}
	return c.Clock.Now(), nil
}

func newClock() *fakeClock { return &fakeClock{Clock: testkit.NewClock(epoch)} }

const lifetime = 10 * time.Minute

var exact = []byte(`{"synthetic":"task"}`)

func newStore(t *testing.T, c *fakeClock, capacity int) *admission.DeliveryStore {
	t.Helper()
	s, err := admission.NewDeliveryStore(c, admission.DeliveryLimits{Capacity: capacity, Lifetime: lifetime}, 2)
	if err != nil {
		t.Fatal("valid store rejected")
	}
	return s
}

// signer is the synthetic remote issuer key; otherSigner is a trusted key of
// the embedded issuer, used to move packages between issuers.
func signer() delivery.Signer      { return delivery.NewSigner("remote-key-1", "remote-issuer", 1) }
func otherSigner() delivery.Signer { return delivery.NewSigner("embedded-key-1", "embedded-issuer", 2) }

// Key validity used by the synthetic policy: 30 days before epoch to 300
// days after it, within the registered 365-day maximum.
var (
	keyNotBefore = epoch.Add(-30 * 24 * time.Hour)
	keyNotAfter  = epoch.Add(300 * 24 * time.Hour)
)

// signed returns an encoded valid package for b issued at epoch+ahead.
func signed(ahead time.Duration, b []byte) []byte {
	return signer().Sign(epoch.Add(ahead), b).Encode()
}

func code(err error) task.AdmissionCode {
	var ae task.AdmissionError
	if !errors.As(err, &ae) {
		return 0
	}
	return ae.Code()
}

func mustEmbedded(t *testing.T, s *admission.DeliveryStore, b []byte) string {
	t.Helper()
	ref, err := s.InstallEmbedded(context.Background(), "embedded-issuer", 2, b)
	if err != nil {
		t.Fatal("embedded install rejected")
	}
	return ref
}

func TestDeliveryStoreBindsSourceContextToExactBytes(t *testing.T) {
	ctx := context.Background()
	s := newStore(t, newClock(), 4)
	buf := append([]byte(nil), exact...)
	ref := mustEmbedded(t, s, buf)
	buf[0] = ' ' // the store must not alias or retain the caller's bytes
	r, err := s.Resolve(ctx, ref, exact)
	if err != nil {
		t.Fatal("embedded record did not resolve")
	}
	src := r.Source()
	if _, issued := src.IssuedAt(); src.Origin() != task.Embedded || src.IssuerID() != "embedded-issuer" || src.PolicyGeneration() != 2 || !src.Verified() || issued {
		t.Fatal("embedded source context not preserved")
	}

	// Embedded resolution is repeatable until expiry; replay handles redelivery.
	if _, err := s.Resolve(ctx, ref, exact); err != nil {
		t.Fatal("resolution consumed an embedded record")
	}

	// A remote record binds only origin and generation to the package bytes;
	// the issuer and issuance time come from admission's verification.
	pkg := signed(0, exact)
	ref, err = s.InstallRemote(ctx, 2, pkg)
	if err != nil {
		t.Fatal("remote package install rejected")
	}
	r, err = s.Resolve(ctx, ref, pkg)
	if err != nil {
		t.Fatal("remote record did not resolve")
	}
	src = r.Source()
	if _, ok := src.IssuedAt(); src.Origin() != task.Remote || src.IssuerID() != "" || src.PolicyGeneration() != 2 || src.Verified() || ok {
		t.Fatal("remote record established provenance before verification")
	}
	if r.Matches(exact) || !r.Matches(pkg) {
		t.Fatal("remote record not bound to the exact package bytes")
	}
}

// A remote record is single-use: its first resolution drops it whatever the
// outcome, so no package digest outlives the admission that presented it.
// requirement-ref: REQ-TRUST-0053
func TestRemoteRecordsAreSingleUse(t *testing.T) {
	ctx := context.Background()
	c := newClock()
	s := newStore(t, c, 8)
	pkg := signed(0, exact)
	install := func() string {
		ref, err := s.InstallRemote(ctx, 2, pkg)
		if err != nil {
			t.Fatal("remote install rejected")
		}
		return ref
	}
	ref := install()
	if _, err := s.Resolve(ctx, ref, pkg); err != nil {
		t.Fatal("first resolution failed")
	}
	if _, err := s.Resolve(ctx, ref, pkg); code(err) != task.SourceUnauthenticated {
		t.Fatal("remote record resolved twice")
	}
	ref = install()
	if _, err := s.Resolve(ctx, ref, exact); code(err) != task.SourceUnauthenticated {
		t.Fatal("mismatched bytes resolved")
	}
	if _, err := s.Resolve(ctx, ref, pkg); code(err) != task.SourceUnauthenticated {
		t.Fatal("remote record survived a mismatched resolution")
	}
	ref = install()
	c.down.Store(true)
	if _, err := s.Resolve(ctx, ref, pkg); code(err) != task.TrustUnavailable {
		t.Fatal("unavailable time not reported")
	}
	c.down.Store(false)
	if _, err := s.Resolve(ctx, ref, pkg); code(err) != task.SourceUnauthenticated {
		t.Fatal("remote record survived a resolution without trusted time")
	}
	// Each retry installs a new reference; resolution frees capacity.
	for range 16 {
		if _, err := s.Resolve(ctx, install(), pkg); err != nil {
			t.Fatal("fresh reference did not resolve")
		}
	}
}

func TestDeliveryReferencesAreOpaque(t *testing.T) {
	s := newStore(t, newClock(), 8)
	seen := map[string]bool{}
	for range 8 {
		ref := mustEmbedded(t, s, exact)
		if len(ref) != 32 || strings.Trim(ref, "0123456789abcdef") != "" || seen[ref] {
			t.Fatal("reference is not a fresh 128-bit lowercase hex value")
		}
		if strings.Contains(ref, "embedded") || strings.Contains(ref, "synthetic") {
			t.Fatal("reference derived from record content")
		}
		seen[ref] = true
	}
}

func TestResolveRejectsUntrustedReferences(t *testing.T) {
	ctx := context.Background()
	c := newClock()
	s := newStore(t, c, 16)
	other := newStore(t, newClock(), 4)
	foreign := mustEmbedded(t, other, exact)
	oneByte := append([]byte(nil), exact...)
	oneByte[len(oneByte)-2] ^= 1

	cases := []struct {
		name string
		ref  func(*testing.T) string
		task []byte
	}{
		{"empty reference", func(t *testing.T) string { return "" }, exact},
		{"forged well-formed reference", func(t *testing.T) string { return strings.Repeat("ab", 16) }, exact},
		{"uppercase reference", func(t *testing.T) string { return strings.ToUpper(mustEmbedded(t, s, exact)) }, exact},
		{"extended reference", func(t *testing.T) string { return mustEmbedded(t, s, exact) + "0" }, exact},
		{"truncated reference", func(t *testing.T) string { return mustEmbedded(t, s, exact)[1:] }, exact},
		{"foreign instance", func(t *testing.T) string { return foreign }, exact},
		{"one-byte mismatch", func(t *testing.T) string { return mustEmbedded(t, s, exact) }, oneByte},
		{"semantically equal bytes", func(t *testing.T) string { return mustEmbedded(t, s, exact) }, []byte(`{ "synthetic":"task"}`)},
		{"empty bytes", func(t *testing.T) string { return mustEmbedded(t, s, exact) }, nil},
		{"invalidated", func(t *testing.T) string { r := mustEmbedded(t, s, exact); s.Invalidate(); return r }, exact},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ref := tc.ref(t)
			if _, err := s.Resolve(ctx, ref, tc.task); code(err) != task.SourceUnauthenticated {
				t.Fatalf("got %v, want task_source_unauthenticated", err)
			}
		})
	}
	if _, err := other.Resolve(ctx, foreign, exact); err != nil {
		t.Fatal("issuing instance lost its record")
	}
}

func TestResolveRejectsStalePolicyGeneration(t *testing.T) {
	ctx := context.Background()
	s := newStore(t, newClock(), 4)
	ref := mustEmbedded(t, s, exact)
	if s.ReplacePolicy(2) == nil || s.ReplacePolicy(1) == nil {
		t.Fatal("non-increasing policy generation accepted")
	}
	if _, err := s.Resolve(ctx, ref, exact); err != nil {
		t.Fatal("rejected replacement invalidated records")
	}
	if err := s.ReplacePolicy(3); err != nil {
		t.Fatal("higher policy generation rejected")
	}
	if _, err := s.Resolve(ctx, ref, exact); code(err) != task.SourceUnauthenticated {
		t.Fatal("record from replaced generation resolved")
	}
	// The source's own generation is checked later by admission (rollback);
	// the store accepts records installed under the new active generation.
	if _, err := s.Resolve(ctx, mustEmbedded(t, s, exact), exact); err != nil {
		t.Fatal("record installed after replacement rejected")
	}
}

func TestResolveExpiryBoundaries(t *testing.T) {
	ctx := context.Background()
	c := newClock()
	s := newStore(t, c, 4)
	ref := mustEmbedded(t, s, exact)
	b, err := testkit.TimeBoundaries(epoch.Add(lifetime), time.Nanosecond)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Advance(b.Before.Sub(epoch)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Resolve(ctx, ref, exact); err != nil {
		t.Fatal("record rejected before expiry")
	}
	if err := c.Advance(b.At.Sub(b.Before)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Resolve(ctx, ref, exact); code(err) != task.SourceUnauthenticated {
		t.Fatal("record resolved at expiry")
	}

	// A trusted clock reading earlier than installation fails closed.
	moved := &fakeClock{Clock: testkit.NewClock(epoch.Add(time.Minute))}
	s = newStore(t, moved, 4)
	ref = mustEmbedded(t, s, exact)
	moved.Clock = testkit.NewClock(epoch)
	if _, err := s.Resolve(ctx, ref, exact); code(err) != task.SourceUnauthenticated {
		t.Fatal("record resolved with clock before installation")
	}

	// Unavailable trusted time is unavailable trust, not a missing record.
	c = newClock()
	s = newStore(t, c, 4)
	ref = mustEmbedded(t, s, exact)
	c.down.Store(true)
	if _, err := s.Resolve(ctx, ref, exact); code(err) != task.TrustUnavailable {
		t.Fatal("unavailable clock not task_trust_unavailable")
	}
	if _, err := s.InstallEmbedded(ctx, "embedded-issuer", 2, exact); !errors.Is(err, admission.ErrDeliveryUnavailable) {
		t.Fatal("install succeeded without trusted time")
	}
	c.down.Store(false)
	if _, err := s.Resolve(ctx, ref, exact); err != nil {
		t.Fatal("record lost while clock was unavailable")
	}
}

func TestDeliveryStoreCapacityBoundaries(t *testing.T) {
	ctx := context.Background()
	c := newClock()
	const capacity = 3
	s := newStore(t, c, capacity)
	refs := make([]string, 0, capacity)
	for range capacity {
		refs = append(refs, mustEmbedded(t, s, exact))
	}
	if _, err := s.InstallEmbedded(ctx, "embedded-issuer", 2, exact); !errors.Is(err, admission.ErrDeliveryCapacity) {
		t.Fatal("install beyond capacity accepted")
	}
	for _, ref := range refs {
		if _, err := s.Resolve(ctx, ref, exact); err != nil {
			t.Fatal("full store evicted a live record")
		}
	}
	if err := c.Advance(lifetime); err != nil {
		t.Fatal(err)
	}
	for range capacity {
		mustEmbedded(t, s, exact) // expired records free capacity
	}
}

func TestDeliveryStoreConfigurationBounds(t *testing.T) {
	c := newClock()
	ok := []admission.DeliveryLimits{{Capacity: 1, Lifetime: time.Nanosecond}, {Capacity: admission.MaxDeliveryCapacity, Lifetime: admission.MaxDeliveryLifetime}}
	for _, l := range ok {
		if _, err := admission.NewDeliveryStore(c, l, 1); err != nil {
			t.Fatal("boundary configuration rejected")
		}
	}
	bad := map[string]func() (*admission.DeliveryStore, error){
		"zero capacity": func() (*admission.DeliveryStore, error) {
			return admission.NewDeliveryStore(c, admission.DeliveryLimits{0, lifetime}, 1)
		},
		"capacity over max": func() (*admission.DeliveryStore, error) {
			return admission.NewDeliveryStore(c, admission.DeliveryLimits{admission.MaxDeliveryCapacity + 1, lifetime}, 1)
		},
		"zero lifetime": func() (*admission.DeliveryStore, error) {
			return admission.NewDeliveryStore(c, admission.DeliveryLimits{1, 0}, 1)
		},
		"lifetime over max": func() (*admission.DeliveryStore, error) {
			return admission.NewDeliveryStore(c, admission.DeliveryLimits{1, admission.MaxDeliveryLifetime + 1}, 1)
		},
		"nil clock": func() (*admission.DeliveryStore, error) {
			return admission.NewDeliveryStore(nil, admission.DeliveryLimits{1, lifetime}, 1)
		},
		"zero generation": func() (*admission.DeliveryStore, error) {
			return admission.NewDeliveryStore(c, admission.DeliveryLimits{1, lifetime}, 0)
		},
	}
	for name, f := range bad {
		if s, err := f(); s != nil || !errors.Is(err, admission.ErrDeliveryConfig) {
			t.Errorf("%s: accepted", name)
		}
	}
}

func TestInstallRejectsUnprovenDeliveries(t *testing.T) {
	ctx := context.Background()
	s := newStore(t, newClock(), 16)
	cases := map[string]func() (string, error){
		"embedded invalid issuer":  func() (string, error) { return s.InstallEmbedded(ctx, "Issuer", 2, exact) },
		"embedded zero generation": func() (string, error) { return s.InstallEmbedded(ctx, "embedded-issuer", 0, exact) },
		"embedded empty task":      func() (string, error) { return s.InstallEmbedded(ctx, "embedded-issuer", 2, nil) },
		"embedded oversized task": func() (string, error) {
			return s.InstallEmbedded(ctx, "embedded-issuer", 2, make([]byte, task.MaxDeliveryBytes+1))
		},
		"remote empty package": func() (string, error) { return s.InstallRemote(ctx, 2, nil) },
		"remote oversized package": func() (string, error) {
			return s.InstallRemote(ctx, 2, make([]byte, task.MaxDeliveryBytes+1))
		},
		"remote zero policy generation": func() (string, error) { return s.InstallRemote(ctx, 0, signed(0, exact)) },
	}
	for name, f := range cases {
		if ref, err := f(); ref != "" || !errors.Is(err, admission.ErrDeliveryRejected) {
			t.Errorf("%s: installed", name)
		}
	}
}

func TestClosedStoreRefusesInstallAndResolution(t *testing.T) {
	ctx := context.Background()
	s := newStore(t, newClock(), 4)
	ref := mustEmbedded(t, s, exact)
	s.Close()
	s.Close()
	if _, err := s.Resolve(ctx, ref, exact); code(err) != task.SourceUnauthenticated {
		t.Fatal("record resolved after close")
	}
	if _, err := s.InstallEmbedded(ctx, "embedded-issuer", 2, exact); !errors.Is(err, admission.ErrDeliveryClosed) {
		t.Fatal("install accepted after close")
	}
	if err := s.ReplacePolicy(3); !errors.Is(err, admission.ErrDeliveryClosed) {
		t.Fatal("policy replacement accepted after close")
	}
}

func TestDeliveryValuesDoNotLeak(t *testing.T) {
	ctx := context.Background()
	const issuer = "canary-issuer-7f3a"
	payload := []byte(`{"canary":"payload-9b1c"}`)
	s := newStore(t, newClock(), 4)
	ref, err := s.InstallEmbedded(ctx, issuer, 2, payload)
	if err != nil {
		t.Fatal("canary install rejected")
	}
	record, err := s.Resolve(ctx, ref, payload)
	if err != nil {
		t.Fatal("canary record did not resolve")
	}
	var outputs []string
	if _, err := s.Resolve(ctx, ref, exact); err != nil {
		outputs = append(outputs, err.Error())
	}
	if _, err := s.InstallEmbedded(ctx, "Bad "+issuer, 2, payload); err != nil {
		outputs = append(outputs, err.Error())
	}
	outputs = append(outputs, fmt.Sprint(s), fmt.Sprintf("%+v", s), fmt.Sprintf("%#v", s))
	for _, out := range outputs {
		if strings.Contains(out, ref) || strings.Contains(out, issuer) || strings.Contains(out, "payload-9b1c") {
			t.Fatal("reference, issuer or payload leaked into output")
		}
	}
	if _, err := json.Marshal(record); err == nil {
		t.Fatal("delivery record serialized")
	}
	if _, err := json.Marshal(record.Source()); err == nil {
		t.Fatal("source context serialized")
	}
	if b, _ := json.Marshal(s); bytes.Contains(b, []byte(ref)) {
		t.Fatal("store serialized a reference")
	}
}

func TestDeliveryStoreConcurrentUse(t *testing.T) {
	ctx := context.Background()
	c := newClock()
	const capacity = 32
	s := newStore(t, c, capacity)
	var wg sync.WaitGroup
	var generation atomic.Uint64
	generation.Store(2)
	for w := range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 200 {
				var ref string
				var err error
				if (w+i)%2 == 0 {
					ref, err = s.InstallEmbedded(ctx, "embedded-issuer", 2, exact)
				} else {
					ref, err = s.InstallRemote(ctx, 2, exact)
				}
				if err != nil && !errors.Is(err, admission.ErrDeliveryCapacity) {
					t.Error("unexpected install error")
					return
				}
				if err == nil {
					if _, err := s.Resolve(ctx, ref, exact); err != nil && code(err) != task.SourceUnauthenticated {
						t.Error("unexpected resolve error")
						return
					}
				}
				switch i % 50 {
				case 17:
					s.Invalidate()
				case 31:
					_ = s.ReplacePolicy(generation.Add(1))
				case 43:
					_ = c.Advance(time.Second)
				}
			}
		}()
	}
	wg.Wait()
	// The store never exceeds its capacity under concurrent installs.
	installed := 0
	for {
		if _, err := s.InstallEmbedded(ctx, "embedded-issuer", 2, exact); err != nil {
			break
		}
		installed++
		if installed > capacity {
			t.Fatal("capacity exceeded")
		}
	}
}
