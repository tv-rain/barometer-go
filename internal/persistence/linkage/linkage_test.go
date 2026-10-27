package linkage_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/config"
	"github.com/tv-rain/barometer-go/internal/persistence/linkage"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/testkit/capability"
)

// Synthetic key material, configuration and clocks only. The vectors were
// computed independently of this package with HMAC-SHA-256 over the
// REQ-LINKAGE-0010 message. No protected storage, platform or privacy
// evidence.

const scope = "org.example.production"

var (
	// at is 2026-10-22T12:00:00Z, Unix 1,792,670,400: epoch 2964 of a
	// 604,800-second period.
	at  = time.Unix(1792670400, 0)
	key = func() (k [32]byte) {
		for i := range k {
			k[i] = byte(i)
		}
		return k
	}()
)

type clock struct {
	mu   sync.Mutex
	now  time.Time
	fail bool
}

func (c *clock) Now(context.Context) (time.Time, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.fail {
		return time.Time{}, errors.New("synthetic clock unavailable")
	}
	return c.now, nil
}

func (c *clock) set(t time.Time) { c.mu.Lock(); c.now = t; c.mu.Unlock() }

// initialization prepares host configuration under synthetic approvals.
func initialization(t *testing.T, configJSON string, allow bool, c platform.Clock, keys platform.LinkageKeyStore) *config.Initialization {
	t.Helper()
	d, err := config.NewDecoder(config.Contract, config.Approvals{LinkageScopes: []string{scope, "org.example.staging"}, AllowLinkage: allow})
	if err != nil {
		t.Fatal(err)
	}
	s, err := d.Decode([]byte(configJSON))
	if err != nil {
		t.Fatal("configuration rejected:", err)
	}
	i, err := config.Prepare(s, config.Capabilities{Clock: c, LinkageKeys: keys})
	if err != nil {
		t.Fatal(err)
	}
	return i
}

func rotating(scope string, period int) string {
	return fmt.Sprintf(`{"result_linkage":{"mode":"rotating_installation","scope":%q,"rotation_period_seconds":%d}}`, scope, period)
}

func token(t *testing.T, s string) [32]byte {
	t.Helper()
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil || len(b) != 32 {
		t.Fatal("bad vector")
	}
	return [32]byte(b)
}

// Derivation is deterministic for the same secret, scope and epoch, uses
// exactly the specified message and no other input, and changes with the
// epoch and the scope.
// requirement-ref: REQ-LINKAGE-0010
// requirement-ref: REQ-LINKAGE-0011
// requirement-ref: REQ-LINKAGE-0012
// requirement-ref: REQ-LINKAGE-0015
func TestDerivation(t *testing.T) {
	c, keys := &clock{now: at}, capability.NewLinkageKeys(key)
	derive := linkage.NewSource().For(initialization(t, rotating(scope, 604800), true, c, keys))
	if derive == nil {
		t.Fatal("enabled linkage has no derivation")
	}
	first := derive(context.Background())
	if first == nil || first.Scope != scope || first.Epoch != 2964 || first.Token != token(t, "yqSMAm-CKF5-kfbeaTlTfNqpd_xXKFv9hPfEIV7D23w") {
		t.Fatalf("epoch 2964 projection %+v", first)
	}
	c.set(time.Unix(2965*604800-1, 0))
	if again := derive(context.Background()); again == nil || *again != *first {
		t.Fatal("same epoch derived a different projection")
	}
	c.set(time.Unix(2965*604800, 0))
	next := derive(context.Background())
	if next == nil || next.Epoch != 2965 || next.Token != token(t, "thsXcilOwzBWOQLjQCrEQP1VX4wrrR_Mc2RdTLDDvyw") {
		t.Fatalf("epoch 2965 projection %+v", next)
	}
	want := []string{"barometer-result-linkage-v1\x00org.example.production\x002964", "barometer-result-linkage-v1\x00org.example.production\x002964",
		"barometer-result-linkage-v1\x00org.example.production\x002965"}
	got := keys.Messages()
	if len(got) != len(want) {
		t.Fatalf("%d derivations, want %d", len(got), len(want))
	}
	for i := range want {
		if !bytes.Equal(got[i], []byte(want[i])) {
			t.Fatalf("message %d is %q", i, got[i])
		}
	}

	other := linkage.NewSource().For(initialization(t, rotating("org.example.staging", 604800), true, &clock{now: at}, capability.NewLinkageKeys(key)))
	if l := other(context.Background()); l == nil || l.Epoch != 2964 || l.Token == first.Token {
		t.Fatal("another scope derived the same token")
	}
}

// Linkage is off unless trusted initialization enables it and a protected-key
// capability exists; a disabled source makes no key call.
// requirement-ref: REQ-LINKAGE-0004
// requirement-ref: REQ-LINKAGE-0043
// requirement-ref: REQ-LINKAGE-0044
func TestDisabledByDefault(t *testing.T) {
	c := &clock{now: at}
	keys := capability.NewLinkageKeys(key)
	for name, init := range map[string]*config.Initialization{
		"absent":          initialization(t, `{}`, true, c, keys),
		"mode none":       initialization(t, `{"result_linkage":{"mode":"none"}}`, true, c, keys),
		"policy disables": initialization(t, rotating(scope, 604800), false, c, keys),
		"no key store":    initialization(t, rotating(scope, 604800), true, c, nil),
	} {
		if linkage.NewSource().For(init) != nil {
			t.Errorf("%s: linkage enabled", name)
		}
	}
	var none *linkage.Source
	if none.For(initialization(t, rotating(scope, 604800), true, c, keys)) != nil || linkage.NewSource().For(nil) != nil {
		t.Error("missing source or initialization enabled linkage")
	}
	if len(keys.Messages()) != 0 {
		t.Fatal("disabled linkage reached the key store")
	}
}

// Every failed prerequisite omits linkage rather than falling back: an
// unavailable or pre-1970 clock, a cancelled context and a key-store failure.
// requirement-ref: REQ-LINKAGE-0014
// requirement-ref: REQ-LINKAGE-0037
// requirement-ref: REQ-LINKAGE-0038
func TestFailuresOmitLinkage(t *testing.T) {
	c, keys := &clock{now: at}, capability.NewLinkageKeys(key)
	derive := linkage.NewSource().For(initialization(t, rotating(scope, 86400), true, c, keys))
	c.fail = true
	if derive(context.Background()) != nil {
		t.Error("unavailable clock derived linkage")
	}
	c.fail = false
	c.set(time.Unix(-1, 0))
	if derive(context.Background()) != nil {
		t.Error("negative Unix time derived linkage")
	}
	c.set(at)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if derive(ctx) != nil {
		t.Error("cancelled context derived linkage")
	}
	if len(keys.Messages()) != 0 {
		t.Fatal("failed prerequisite reached the key store")
	}
	keys.Fail()
	if derive(context.Background()) != nil {
		t.Error("key-store failure derived linkage")
	}
	keys.Recover()
	if derive(context.Background()) == nil {
		t.Error("linkage not available again after recovery")
	}
}

// A clock moved back into an earlier epoch omits linkage until a
// non-decreasing epoch returns; an earlier epoch is never reused. A new
// scope or period starts its own floor, and so does Erase.
// requirement-ref: REQ-LINKAGE-0014
// requirement-ref: REQ-LINKAGE-0023
func TestEpochFloor(t *testing.T) {
	c, keys := &clock{now: at}, capability.NewLinkageKeys(key)
	source := linkage.NewSource()
	derive := source.For(initialization(t, rotating(scope, 86400), true, c, keys))
	if l := derive(context.Background()); l == nil || l.Epoch != 20748 {
		t.Fatal("no linkage at epoch 20748")
	}
	c.set(at.Add(-24 * time.Hour))
	if derive(context.Background()) != nil {
		t.Fatal("earlier epoch reused")
	}
	// A later generation of the same client keeps the floor.
	if source.For(initialization(t, rotating(scope, 86400), true, c, keys))(context.Background()) != nil {
		t.Fatal("reinitialization reopened an earlier epoch")
	}
	c.set(at)
	if l := derive(context.Background()); l == nil || l.Epoch != 20748 {
		t.Fatal("linkage not restored at the floor epoch")
	}
	c.set(at.Add(-24 * time.Hour))
	if source.Erase(context.Background()) != nil {
		t.Fatal("erase failed")
	}
	if l := derive(context.Background()); l == nil || l.Epoch != 20747 {
		t.Fatal("erase kept the floor")
	}
	// Another rotation period, which needs a new initialization, has its
	// own epoch numbering and so its own floor.
	c.set(at.Add(-48 * time.Hour))
	if l := source.For(initialization(t, rotating(scope, 172800), true, c, keys))(context.Background()); l == nil || l.Epoch != 10373 {
		t.Fatal("another rotation period inherited the floor")
	}
	if fmt.Sprintf("%v %#v", source, source) != "linkage.Source linkage.Source" {
		t.Fatal("source formatting exposes state")
	}
}

// A clear that runs while a derivation is in progress discards its token, so
// no token from a deleted key is emitted.
// requirement-ref: REQ-LINKAGE-0023
func TestEraseDuringDerivation(t *testing.T) {
	keys := capability.NewLinkageKeys(key)
	source := linkage.NewSource()
	derive := source.For(initialization(t, rotating(scope, 604800), true, &clock{now: at}, keys))
	keys.OnDerive = func() { _ = source.Erase(context.Background()); _ = keys.Delete(context.Background()) }
	if derive(context.Background()) != nil {
		t.Fatal("derivation interrupted by clearing emitted linkage")
	}
	keys.OnDerive = nil
	l := derive(context.Background())
	if held, created, _ := keys.Held(); l == nil || !held || created != 1 || l.Token == token(t, "yqSMAm-CKF5-kfbeaTlTfNqpd_xXKFv9hPfEIV7D23w") {
		t.Fatal("derivation after clearing did not use a new key")
	}
}

// Concurrent derivations in one epoch agree and never lower the floor.
func TestConcurrentDerivation(t *testing.T) {
	derive := linkage.NewSource().For(initialization(t, rotating(scope, 604800), true, &clock{now: at}, capability.NewLinkageKeys(key)))
	var wg sync.WaitGroup
	out := make([]*[32]byte, 16)
	for i := range out {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if l := derive(context.Background()); l != nil {
				out[i] = &l.Token
			}
		}()
	}
	wg.Wait()
	for _, tok := range out {
		if tok == nil || *tok != token(t, "yqSMAm-CKF5-kfbeaTlTfNqpd_xXKFv9hPfEIV7D23w") {
			t.Fatal("concurrent derivation disagreed or failed")
		}
	}
}
