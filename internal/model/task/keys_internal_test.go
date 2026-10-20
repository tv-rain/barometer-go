package task

import (
	"crypto/ed25519"
	"encoding/hex"
	"math/big"
	"testing"
)

// smallOrder are the canonical encodings of the eight points of order 1, 2, 4
// and 8 on edwards25519.
var smallOrder = []string{
	"0100000000000000000000000000000000000000000000000000000000000000",
	"ecffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff7f",
	"0000000000000000000000000000000000000000000000000000000000000000",
	"0000000000000000000000000000000000000000000000000000000000000080",
	"c7176a703d4dd84fba3c0b760d10670f2a2053fa2c39ccc64ec7fd7792ac037a",
	"c7176a703d4dd84fba3c0b760d10670f2a2053fa2c39ccc64ec7fd7792ac03fa",
	"26e8958fc2b227b045c3f489f2ef98f0d5dfac05d3c63339b13802886d53fc05",
	"26e8958fc2b227b045c3f489f2ef98f0d5dfac05d3c63339b13802886d53fc85",
}

func encoding(t *testing.T, s string) [PublicKeySize]byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != PublicKeySize {
		t.Fatal("bad fixture")
	}
	return [PublicKeySize]byte(b)
}

// The small-order encodings must decode as curve points, so their rejection is
// the small-order check and not a decoding failure; each is a distinct point
// of order dividing 8.
// requirement-ref: REQ-TRUST-0051
func TestSmallOrderPointsDecodeAndAreRejected(t *testing.T) {
	seen := map[string]bool{}
	for _, s := range smallOrder {
		b := encoding(t, s)
		x, y, ok := decodePoint(b)
		if !ok {
			t.Fatalf("%s does not decode", s)
		}
		seen[x.String()+","+y.String()] = true
		px, py := x, y
		for range 3 {
			px, py = pointAdd(px, py, px, py)
		}
		if px.Sign() != 0 || py.Cmp(big.NewInt(1)) != 0 {
			t.Fatalf("%s is not of small order", s)
		}
		if ValidPublicKey(b) {
			t.Fatalf("%s accepted", s)
		}
	}
	if len(seen) != 8 {
		t.Fatal("small-order fixtures are not eight distinct points")
	}
}

// Encodings that decode under lenient rules but are not canonical, and
// coordinates that are not on the curve, are rejected by decoding.
func TestNonCanonicalAndOffCurveKeysRejected(t *testing.T) {
	for _, s := range []string{
		"0100000000000000000000000000000000000000000000000000000000000080", // identity, x = 0 with sign bit
		"ecffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", // y = p - 1, x = 0 with sign bit
		"edffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff7f", // y = p
		"eeffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff7f", // y = p + 1, non-canonical identity
		"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff7f", // y = 2^255 - 1
	} {
		if _, _, ok := decodePoint(encoding(t, s)); ok {
			t.Fatalf("%s decoded", s)
		}
	}
	offCurve := 0
	for y := byte(2); y < 64; y++ {
		var b [PublicKeySize]byte
		b[0] = y
		if _, _, ok := decodePoint(b); !ok {
			offCurve++
			if ValidPublicKey(b) {
				t.Fatal("off-curve key accepted")
			}
		}
	}
	if offCurve == 0 {
		t.Fatal("no off-curve fixture exercised")
	}
}

// Keys produced by the standard library from fixed seeds are canonical points
// of large order, and decoding then re-encoding them is the identity.
func TestGeneratedKeysAccepted(t *testing.T) {
	for label := range 32 {
		seed := make([]byte, ed25519.SeedSize)
		for i := range seed {
			seed[i] = byte(label*7 + i)
		}
		var b [PublicKeySize]byte
		copy(b[:], ed25519.NewKeyFromSeed(seed).Public().(ed25519.PublicKey))
		x, y, ok := decodePoint(b)
		if !ok || !ValidPublicKey(b) {
			t.Fatal("generated key rejected")
		}
		le := y.FillBytes(make([]byte, PublicKeySize))
		var enc [PublicKeySize]byte
		for i := range le {
			enc[i] = le[PublicKeySize-1-i]
		}
		enc[31] |= byte(x.Bit(0)) << 7
		if enc != b {
			t.Fatal("decoding does not round-trip")
		}
	}
}
