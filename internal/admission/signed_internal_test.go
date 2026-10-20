package admission

import (
	"bytes"
	"crypto/sha512"
	"math/big"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/testkit/delivery"
)

var signedAt = time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)

// The admission message construction matches the independent testkit one,
// byte for byte, and decoding returns the exact members.
// requirement-ref: REQ-TRUST-0048
// requirement-ref: REQ-TRUST-0049
func TestDecodePackageAndMessage(t *testing.T) {
	s := delivery.NewSigner("remote-key-1", "remote-issuer", 1)
	exact := []byte(`{"synthetic":"task","binary":"` + "\x00\xff" + `"}`)
	p, ok := decodePackage(s.Sign(signedAt, exact).Encode())
	if !ok || p.issuerID != "remote-issuer" || p.keyID != "remote-key-1" || p.issuedText != "2026-10-19T12:00:00Z" || !p.issued.Equal(signedAt) || !bytes.Equal(p.task, exact) {
		t.Fatal("valid package not decoded exactly")
	}
	if !bytes.Equal(signedMessage(p), delivery.Message("remote-key-1", "remote-issuer", "2026-10-19T12:00:00Z", exact)) {
		t.Fatal("message construction differs from REQ-TRUST-0049")
	}
	want := slices.Concat([]byte{0, 0, 0, 24}, []byte("barometer-signed-task-v1"), []byte{0, 0, 0, 12}, []byte("remote-key-1"), []byte{0, 0, 0, 13}, []byte("remote-issuer"), []byte{0, 0, 0, 20}, []byte("2026-10-19T12:00:00Z"), []byte{0, 0, 0, byte(len(exact))}, exact)
	if !bytes.Equal(signedMessage(p), want) {
		t.Fatal("message is not the length-prefixed concatenation")
	}
	if !verifySignature(s.PublicKey(), p) {
		t.Fatal("valid signature rejected")
	}
	// Length prefixes keep boundaries: moving a byte between members changes M.
	q := p
	q.keyID, q.issuerID = "remote-key-1r", "emote-issuer"
	if bytes.Equal(signedMessage(q), signedMessage(p)) || verifySignature(s.PublicKey(), q) {
		t.Fatal("member boundary is ambiguous")
	}
	if formatted := strings.Join([]string{p.String(), p.GoString()}, " "); formatted != "admission.signedPackage admission.signedPackage" {
		t.Fatal("package formatting exposes members")
	}
}

func TestPackageMemberFormats(t *testing.T) {
	for _, s := range []string{"2026-10-19T12:00:00Z", "0001-01-01T00:00:00Z", "9999-12-31T23:59:59Z", "2028-02-29T00:00:00Z"} {
		if _, ok := packageTime(s); !ok {
			t.Errorf("%s rejected", s)
		}
	}
	for _, s := range []string{"", "2026-10-19T12:00:00", "2026-10-19t12:00:00Z", "2026-10-19T12:00:00z", "2026-10-19 12:00:00Z", "2026-10-19T12:00:00.0Z", "2026-10-19T12:00:00.5Z",
		"2026-10-19T12:00:00+00:00", "2026-10-19T14:00:00+02:00", "2026-10-19T23:59:60Z", "2026-02-30T00:00:00Z", "2027-02-29T00:00:00Z", "0000-01-01T00:00:00Z",
		"2026-13-01T00:00:00Z", "2026-10-19T24:00:00Z", "+2026-10-19T12:00:00Z", " 2026-10-19T12:00:00Z", "2026-1-19T12:00:00Z"} {
		if _, ok := packageTime(s); ok {
			t.Errorf("%q accepted", s)
		}
	}
	for _, s := range []string{"a", "key-1", "a.b_c-d", strings.Repeat("k", 64)} {
		if !packageIdentifier(s) {
			t.Errorf("%q rejected", s)
		}
	}
	for _, s := range []string{"", "Key", "key 1", "key/1", "ключ", strings.Repeat("k", 65), "key\x00"} {
		if packageIdentifier(s) {
			t.Errorf("%q accepted", s)
		}
	}
	for in, want := range map[string]string{"AA": "\x00", "AAA": "\x00\x00", "_-8": "\xff\xef"} {
		if b, ok := canonicalBase64URL(in); !ok || string(b) != want {
			t.Errorf("%q not decoded", in)
		}
	}
	// Padding, whitespace (which Go's decoder skips), the standard alphabet,
	// nonzero trailing bits and impossible lengths are not canonical.
	for _, s := range []string{"AA==", "AA=", "A A", "AA\n", "\rAA", "AA\t", "+/", "AB", "AAB", "A", "AAAAA", "AA%"} {
		if _, ok := canonicalBase64URL(s); ok {
			t.Errorf("%q accepted", s)
		}
	}
}

// scalar is the clamped Ed25519 secret scalar of a signer (RFC 8032 5.1.5).
func scalar(s delivery.Signer) *big.Int {
	h := sha512.Sum512(s.Seed[:])
	a := h[:32]
	a[0] &= 248
	a[31] &= 127
	a[31] |= 64
	return new(big.Int).SetBytes(reversed(a))
}

func reversed(b []byte) []byte {
	r := slices.Clone(b)
	slices.Reverse(r)
	return r
}

// le32 encodes n as 32 little-endian bytes.
func le32(n *big.Int) []byte { return reversed(n.FillBytes(make([]byte, 32))) }

var groupOrder, _ = new(big.Int).SetString("7237005577332262213973186563042994240857116359379907606001950938285454250989", 10)

// signWithR builds a signature whose R bytes are rEnc for an R of the
// identity point: S = k*a mod L satisfies [S]B = R + [k]A when R is the
// identity, so the only property under test is the encoding of R.
func signWithR(s delivery.Signer, rEnc []byte, m []byte) []byte {
	pub := s.PublicKey()
	h := sha512.Sum512(slices.Concat(rEnc, pub[:], m))
	k := new(big.Int).Mod(new(big.Int).SetBytes(reversed(h[:])), groupOrder)
	sv := new(big.Int).Mul(k, scalar(s))
	sv.Mod(sv, groupOrder)
	return slices.Concat(rEnc, le32(sv))
}

// Edge vectors for REQ-TRUST-0050, built from synthetic keys so the
// verification equation provably holds for the decoded values: only the
// encoding differs between the accepted control and the rejected vector.
// requirement-ref: REQ-TRUST-0050
func TestSignatureStrictnessEdgeVectors(t *testing.T) {
	s := delivery.NewSigner("remote-key-1", "remote-issuer", 1)
	p, ok := decodePackage(s.Sign(signedAt, []byte(`{"synthetic":"task"}`)).Encode())
	if !ok {
		t.Fatal("fixture rejected")
	}
	m := signedMessage(p)
	pub := s.PublicKey()
	withSig := func(sig []byte) signedPackage {
		q := p
		copy(q.signature[:], sig)
		return q
	}
	if !verifySignature(pub, p) {
		t.Fatal("control signature rejected")
	}

	// Non-canonical R: the identity encoded with y = p + 1. The control with
	// the canonical identity encoding is accepted, so the equation holds.
	canonicalR := append([]byte{1}, make([]byte, 31)...)
	nonCanonicalR := append([]byte{0xee}, bytes.Repeat([]byte{0xff}, 30)...)
	nonCanonicalR = append(nonCanonicalR, 0x7f)
	if !verifySignature(pub, withSig(signWithR(s, canonicalR, m))) {
		t.Fatal("canonical-R control rejected; the vector construction is wrong")
	}
	if verifySignature(pub, withSig(signWithR(s, nonCanonicalR, m))) {
		t.Fatal("non-canonical R accepted")
	}
	// The identity with the x sign bit set is also a non-canonical R.
	signR := append([]byte{1}, make([]byte, 30)...)
	signR = append(signR, 0x80)
	if verifySignature(pub, withSig(signWithR(s, signR, m))) {
		t.Fatal("non-canonical R with sign bit accepted")
	}

	// S not below the group order: S + L satisfies the equation modulo L.
	valid := p.signature[:]
	sv := new(big.Int).SetBytes(reversed(valid[32:]))
	for _, bad := range []*big.Int{new(big.Int).Add(sv, groupOrder), new(big.Int).Add(sv, new(big.Int).Lsh(groupOrder, 1))} {
		if verifySignature(pub, withSig(slices.Concat(valid[:32], le32(bad)))) {
			t.Fatal("S not below the group order accepted")
		}
	}
	// S exactly L, and S with the top bits set.
	if verifySignature(pub, withSig(slices.Concat(valid[:32], le32(groupOrder)))) {
		t.Fatal("S equal to the group order accepted")
	}
	high := slices.Clone(valid)
	high[63] |= 0xe0
	if verifySignature(pub, withSig(high)) {
		t.Fatal("S with high bits accepted")
	}
	// Any single bit flip in R, S, the message or the key fails.
	for i := range 64 {
		flipped := slices.Clone(valid)
		flipped[i] ^= 1
		if verifySignature(pub, withSig(flipped)) {
			t.Fatalf("signature with byte %d flipped accepted", i)
		}
	}
	other := pub
	other[0] ^= 1
	if task.ValidPublicKey(other) && verifySignature(other, p) {
		t.Fatal("signature verified under another key")
	}
}
