package task_test

import (
	"crypto/ed25519"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/task"
)

func publicKey(label byte) [task.PublicKeySize]byte {
	seed := make([]byte, ed25519.SeedSize)
	for i := range seed {
		seed[i] = label ^ byte(i)
	}
	var k [task.PublicKeySize]byte
	copy(k[:], ed25519.NewKeyFromSeed(seed).Public().(ed25519.PublicKey))
	return k
}

func trustedKey(id, issuer string, label byte) task.TrustedKey {
	nb := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return task.TrustedKey{KeyID: id, IssuerID: issuer, PublicKey: publicKey(label),
		NotBefore: task.Timestamp{Whole: nb}, NotAfter: task.Timestamp{Whole: nb.Add(task.MaxKeyValidity)}}
}

func keyedPolicy() task.PolicyInput {
	in := policy()
	in.Keys = []task.TrustedKey{trustedKey("key-1", "reviewed-issuer", 1)}
	return in
}

// requirement-ref: REQ-TRUST-0051
// requirement-ref: REQ-LIMIT-0032
func TestPolicyCarriesBoundedTrustedKeys(t *testing.T) {
	in := keyedPolicy()
	p, err := task.NewLocalPolicy(in)
	if err != nil {
		t.Fatal(err)
	}
	in.Keys[0].KeyID = "changed"
	in.Keys[0].PublicKey[0] ^= 1
	k, ok := p.TrustedKey("key-1")
	if !ok || k.IssuerID != "reviewed-issuer" || k.PublicKey != publicKey(1) {
		t.Fatal("trusted key lost or aliased")
	}
	k.PublicKey[0] ^= 1
	if again, _ := p.TrustedKey("key-1"); again.PublicKey != publicKey(1) {
		t.Fatal("trusted key getter aliased")
	}
	if _, ok := p.TrustedKey("changed"); ok {
		t.Fatal("unknown key ID resolved")
	}
	if _, ok := p.TrustedKey("KEY-1"); ok {
		t.Fatal("key lookup is not exact")
	}
	// A policy without keys stays valid; it admits no remote task.
	if p, err := task.NewLocalPolicy(policy()); err != nil || !p.Valid() {
		t.Fatal("policy without keys rejected")
	}
	if _, ok := (task.LocalPolicy{}).TrustedKey("key-1"); ok {
		t.Fatal("zero policy resolved a key")
	}

	// Count bounds: four keys per issuer and 32 in one policy are accepted.
	in = keyedPolicy()
	in.Keys = nil
	for n := range 4 {
		in.Keys = append(in.Keys, trustedKey(fmt.Sprintf("key-%d", n), "reviewed-issuer", byte(n)))
	}
	if _, err := task.NewLocalPolicy(in); err != nil {
		t.Fatal("four keys for one issuer rejected")
	}
	in = manyIssuers(8)
	if _, err := task.NewLocalPolicy(in); err != nil || len(in.Keys) != task.MaxPolicyKeys {
		t.Fatal("32 keys in one policy rejected")
	}
}

// manyIssuers returns a policy with n issuers holding four keys each.
func manyIssuers(n int) task.PolicyInput {
	in := keyedPolicy()
	in.Keys = nil
	base := in.Issuers[0]
	in.Issuers = nil
	for i := range n {
		issuer := base
		issuer.IssuerID = fmt.Sprintf("issuer-%d", i)
		in.Issuers = append(in.Issuers, issuer)
		for k := range 4 {
			in.Keys = append(in.Keys, trustedKey(fmt.Sprintf("key-%d-%d", i, k), issuer.IssuerID, byte(i*4+k)))
		}
	}
	return in
}

// Every malformed or ambiguous key record rejects the whole policy, with the
// fixed error and no partial snapshot.
// requirement-ref: REQ-TRUST-0051
// requirement-ref: REQ-LIMIT-0032
func TestPolicyRejectsMalformedOrAmbiguousKeys(t *testing.T) {
	second := func(p *task.PolicyInput, k task.TrustedKey) { p.Keys = append(p.Keys, k) }
	cases := map[string]func(*task.PolicyInput){
		"key id empty":      func(p *task.PolicyInput) { p.Keys[0].KeyID = "" },
		"key id case":       func(p *task.PolicyInput) { p.Keys[0].KeyID = "Key-1" },
		"key id long":       func(p *task.PolicyInput) { p.Keys[0].KeyID = strings.Repeat("k", 65) },
		"key id space":      func(p *task.PolicyInput) { p.Keys[0].KeyID = "key 1" },
		"duplicate key id":  func(p *task.PolicyInput) { second(p, trustedKey("key-1", "reviewed-issuer", 2)) },
		"duplicate across":  func(p *task.PolicyInput) { second(p, trustedKey("key-1", "other-issuer", 2)) },
		"issuer unknown":    func(p *task.PolicyInput) { p.Keys[0].IssuerID = "absent-issuer" },
		"issuer empty":      func(p *task.PolicyInput) { p.Keys[0].IssuerID = "" },
		"zero public key":   func(p *task.PolicyInput) { p.Keys[0].PublicKey = [task.PublicKeySize]byte{} },
		"identity key":      func(p *task.PolicyInput) { p.Keys[0].PublicKey = [task.PublicKeySize]byte{1} },
		"non-canonical key": func(p *task.PolicyInput) { p.Keys[0].PublicKey = noncanonical() },
		"equal validity":    func(p *task.PolicyInput) { p.Keys[0].NotAfter = p.Keys[0].NotBefore },
		"reversed validity": func(p *task.PolicyInput) {
			p.Keys[0].NotBefore, p.Keys[0].NotAfter = p.Keys[0].NotAfter, p.Keys[0].NotBefore
		},
		"validity over maximum": func(p *task.PolicyInput) {
			p.Keys[0].NotAfter.Whole = p.Keys[0].NotAfter.Whole.Add(time.Second)
		},
		"zero not before":    func(p *task.PolicyInput) { p.Keys[0].NotBefore = task.Timestamp{} },
		"fraction not after": func(p *task.PolicyInput) { p.Keys[0].NotAfter.Fraction = "5" },
		"sub-second not before": func(p *task.PolicyInput) {
			p.Keys[0].NotBefore.Whole = p.Keys[0].NotBefore.Whole.Add(time.Millisecond)
		},
		"not UTC": func(p *task.PolicyInput) {
			p.Keys[0].NotBefore.Whole = p.Keys[0].NotBefore.Whole.In(time.FixedZone("offset", 3600))
		},
		"fifth key for issuer": func(p *task.PolicyInput) {
			for n := 2; n <= 5; n++ {
				second(p, trustedKey(fmt.Sprintf("key-%d", n), "reviewed-issuer", byte(n)))
			}
		},
		"33rd key": func(p *task.PolicyInput) {
			*p = manyIssuers(8)
			p.Issuers = append(p.Issuers, task.IssuerPermissions{IssuerID: "issuer-8", Permissions: p.Issuers[0].Permissions})
			second(p, trustedKey("key-8-0", "issuer-8", 99))
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			in := keyedPolicy()
			in.Issuers = append(in.Issuers, task.IssuerPermissions{IssuerID: "other-issuer", Permissions: in.Issuers[0].Permissions})
			mutate(&in)
			p, err := task.NewLocalPolicy(in)
			if err != task.ErrTrustModel || p.Valid() || p.Generation() != 0 {
				t.Fatal("invalid key record escaped")
			}
			if _, ok := p.TrustedKey("key-1"); ok {
				t.Fatal("rejected policy exposes a key")
			}
		})
	}
}

// noncanonical encodes y = p + 1, an alias of the identity's y coordinate.
func noncanonical() [task.PublicKeySize]byte {
	k := [task.PublicKeySize]byte{0xee}
	for i := 1; i < 31; i++ {
		k[i] = 0xff
	}
	k[31] = 0x7f
	return k
}

// Validity covers both ends inclusively (REQ-TRUST-0052 check 7).
func TestTrustedKeyCoversInclusiveValidity(t *testing.T) {
	k := trustedKey("key-1", "reviewed-issuer", 1)
	nb, na := k.NotBefore.Whole, k.NotAfter.Whole
	for _, c := range []struct {
		at   time.Time
		want bool
	}{{nb, true}, {na, true}, {nb.Add(-time.Second), false}, {na.Add(time.Second), false}, {nb.Add(-time.Nanosecond), false}, {na.Add(time.Nanosecond), false}, {nb.Add(time.Hour), true}} {
		if k.Covers(c.at) != c.want {
			t.Fatalf("Covers(%v) = %v", c.at, !c.want)
		}
	}
}
