package pending_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/bundle"
	"github.com/tv-rain/barometer-go/internal/config"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/persistence"
	"github.com/tv-rain/barometer-go/internal/persistence/pending"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit"
	"github.com/tv-rain/barometer-go/internal/testkit/capability"
	"github.com/tv-rain/barometer-go/internal/testkit/records"
)

// Synthetic linkage projections supplied by a scripted derivation; the
// derivation itself is tested in internal/persistence/linkage. No protected
// storage or platform evidence.

const linkScope = "org.example.production"

// linkedBuilder is a builder whose initialization enables rotating linkage
// with a synthetic key store.
func linkedBuilder(t testing.TB) *bundle.Builder {
	t.Helper()
	d, err := config.NewDecoder(config.Contract, config.Approvals{LinkageScopes: []string{linkScope}, AllowLinkage: true})
	if err != nil {
		t.Fatal(err)
	}
	s, err := d.Decode([]byte(`{"result_linkage":{"mode":"rotating_installation","scope":"` + linkScope + `","rotation_period_seconds":86400}}`))
	if err != nil {
		t.Fatal(err)
	}
	i, err := config.Prepare(s, config.Capabilities{Clock: configClock{}, LinkageKeys: capability.NewLinkageKeys([32]byte{})})
	if err != nil {
		t.Fatal(err)
	}
	b, err := bundle.New(bundle.Contract, i)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// script is a derivation that returns its current projection and counts calls.
type script struct {
	mu    sync.Mutex
	next  *model.RotatingLinkage
	calls int
}

func (s *script) set(l *model.RotatingLinkage) { s.mu.Lock(); s.next = l; s.mu.Unlock() }

func (s *script) derive(context.Context) *model.RotatingLinkage {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	if s.next == nil {
		return nil
	}
	c := *s.next
	return &c
}

func projection(epoch int64, fill byte) *model.RotatingLinkage {
	l := &model.RotatingLinkage{Scope: linkScope, Epoch: epoch}
	for i := range l.Token {
		l.Token[i] = fill
	}
	return l
}

type linkedWire struct {
	BundleID string `json:"bundle_id"`
	Linkage  *struct {
		Mode  string `json:"mode"`
		Scope string `json:"scope"`
		Epoch int64  `json:"epoch"`
		Token string `json:"token"`
	} `json:"linkage"`
	Results []json.RawMessage `json:"results"`
}

func linked(t *testing.T, b []byte) linkedWire {
	t.Helper()
	var w linkedWire
	if err := json.Unmarshal(b, &w); err != nil {
		t.Fatal("not a bundle")
	}
	return w
}

func linkedStore(t *testing.T, r *records.Store, c *testkit.Clock, s *script, limits pending.Limits) *pending.Store {
	t.Helper()
	p, err := pending.New(r, linkedBuilder(t), s.derive, c, limits, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Open(context.Background()); err != nil {
		t.Fatal("open failed:", err)
	}
	return p
}

// A bundle carries the linkage derived when it is created; results added
// while it is open keep that epoch and token, and only a new bundle derives
// again. No derivation, or a projection the builder rejects, omits linkage
// while the result is still stored.
// requirement-ref: REQ-LINKAGE-0015
// requirement-ref: REQ-LINKAGE-0037
func TestBundleLinkage(t *testing.T) {
	r, c, s := records.New(), testkit.NewClock(epoch), &script{}
	s.set(projection(20748, 1))
	p := linkedStore(t, r, c, s, pending.Limits{})
	ctx := context.Background()
	for range 2 {
		if err := p.Put(ctx, value(t)); err != nil {
			t.Fatal(err)
		}
	}
	s.set(projection(20749, 2))
	if err := p.Put(ctx, value(t)); err != nil {
		t.Fatal(err)
	}
	id, b, _, err := p.Next(ctx)
	w := linked(t, b)
	if err != nil || w.BundleID != id || len(w.Results) != 3 || w.Linkage == nil || w.Linkage.Mode != "rotating_installation" ||
		w.Linkage.Scope != linkScope || w.Linkage.Epoch != 20748 || w.Linkage.Token != base64.RawURLEncoding.EncodeToString(projection(0, 1).Token[:]) || s.calls != 1 {
		t.Fatalf("first bundle linkage %+v after %d derivations", w.Linkage, s.calls)
	}
	if err := p.Put(ctx, value(t)); err != nil {
		t.Fatal(err)
	}
	if err := p.Acknowledge(ctx, id); err != nil {
		t.Fatal(err)
	}
	_, b, _, _ = p.Next(ctx)
	if w := linked(t, b); w.Linkage == nil || w.Linkage.Epoch != 20749 || s.calls != 2 {
		t.Fatal("new bundle did not derive the current epoch")
	}

	for name, l := range map[string]*model.RotatingLinkage{
		"unavailable":    nil,
		"other scope":    {Scope: "org.example.staging", Epoch: 1},
		"negative epoch": {Scope: linkScope, Epoch: -1},
	} {
		r, s := records.New(), &script{}
		s.set(l)
		p := linkedStore(t, r, c, s, pending.Limits{})
		if err := p.Put(ctx, value(t)); err != nil {
			t.Fatalf("%s: result not stored: %v", name, err)
		}
		if _, b, _, _ := p.Next(ctx); linked(t, b).Linkage != nil || bytes.Contains(b, []byte("linkage")) {
			t.Fatalf("%s: linkage emitted", name)
		}
	}

	// A store without a derivation never emits linkage.
	unlinked, err := pending.New(records.New(), linkedBuilder(t), nil, c, pending.Limits{}, time.Second)
	if err != nil || unlinked.Put(ctx, value(t)) != nil {
		t.Fatal(err)
	}
	if _, b, _, _ := unlinked.Next(ctx); bytes.Contains(b, []byte("linkage")) {
		t.Fatal("linkage without a derivation")
	}
}

// Linked bundles survive a restart byte for byte, and only the exact linkage
// form the producer writes is accepted from storage.
func TestLinkedBundleRestart(t *testing.T) {
	r, c, s := records.New(), testkit.NewClock(epoch), &script{}
	s.set(projection(math.MaxInt64, 0xff))
	p := linkedStore(t, r, c, s, pending.Limits{})
	if err := p.Put(context.Background(), value(t)); err != nil {
		t.Fatal(err)
	}
	r.Restart()
	id, b, _, err := linkedStore(t, r, c, s, pending.Limits{}).Next(context.Background())
	if w := linked(t, b); err != nil || w.BundleID != id || w.Linkage == nil || w.Linkage.Epoch != math.MaxInt64 {
		t.Fatal("linked bundle not restored")
	}

	keys, _ := r.List(context.Background(), persistence.PendingClass)
	good, _ := r.Value(keys[0])
	token := base64.RawURLEncoding.EncodeToString(projection(0, 0xff).Token[:])
	replace := func(old, new string) []byte { return bytes.Replace(good, []byte(old), []byte(new), 1) }
	for name, raw := range map[string][]byte{
		"other mode":       replace(`"mode":"rotating_installation"`, `"mode":"stable_installation"`),
		"empty scope":      replace(`"scope":"`+linkScope+`"`, `"scope":""`),
		"uppercase scope":  replace(linkScope, "Org.example.production"),
		"escaped scope":    replace(linkScope, "org.example\\u002eproduction"),
		"long scope":       replace(linkScope, string(bytes.Repeat([]byte("a"), 129))),
		"leading zero":     replace(`"epoch":9223372036854775807`, `"epoch":09`),
		"negative epoch":   replace(`"epoch":9223372036854775807`, `"epoch":-1`),
		"overflow epoch":   replace(`"epoch":9223372036854775807`, `"epoch":9223372036854775808`),
		"fraction epoch":   replace(`"epoch":9223372036854775807`, `"epoch":1.0`),
		"short token":      replace(token, token[:42]),
		"padded token":     replace(token, token+"="),
		"noncanonical end": replace(token, token[:42]+"_"),
		"extra member":     replace(`"token":"`+token+`"}`, `"token":"`+token+`","prior":"x"}`),
		"reordered":        replace(`{"mode":"rotating_installation","scope":"`+linkScope+`"`, `{"scope":"`+linkScope+`","mode":"rotating_installation"`),
	} {
		t.Run(name, func(t *testing.T) {
			if bytes.Equal(raw, good) {
				t.Fatal("corruption did not apply")
			}
			r := records.New()
			r.Corrupt(keys[0], raw)
			s, _ := pending.New(r, linkedBuilder(t), nil, c, pending.Limits{}, time.Second)
			if err := s.Open(context.Background()); !errors.Is(err, pending.ErrUnavailable) {
				t.Fatal("invalid stored linkage accepted")
			}
		})
	}
}

// The builder's linkage reserve is the size of the largest linkage member, and
// a linked store keeps it free in every bundle it counts as capacity.
// requirement-ref: REQ-LIMIT-0021
func TestLinkageReserve(t *testing.T) {
	b := linkedBuilder(t)
	v := []result.Value{value(t)}
	plain, _ := b.Build(v, nil)
	full, err := b.Build(v, projection(math.MaxInt64, 0))
	if err != nil {
		t.Fatal(err)
	}
	p, _ := plain.MarshalJSON()
	f, _ := full.MarshalJSON()
	if b.LinkageReserve() != len(f)-len(p) || builder(t).LinkageReserve() != 0 {
		t.Fatal("linkage reserve:", b.LinkageReserve(), len(f)-len(p))
	}
	minimum := 76 + b.LinkageReserve() + 65536
	if _, err := pending.New(records.New(), b, nil, testkit.NewClock(epoch), pending.Limits{Bytes: minimum - 1}, time.Second); !errors.Is(err, pending.ErrConfig) {
		t.Fatal("linked store accepted a byte limit without room for linkage")
	}
	s := &script{}
	if n, err := linkedStore(t, records.New(), testkit.NewClock(epoch), s, pending.Limits{Bytes: minimum}).Capacity(context.Background()); err != nil || n != 1 {
		t.Fatal("capacity of the smallest linked store:", n, err)
	}
	if n, _ := linkedStore(t, records.New(), testkit.NewClock(epoch), s, pending.Limits{}).Capacity(context.Background()); n != 63 {
		t.Fatal("capacity of an empty linked store at the registered maxima:", n)
	}
}
