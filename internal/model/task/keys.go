package task

import (
	"math/big"
	"time"
)

// Registered trusted-key bounds (REQ-LIMIT-0032).
const (
	MaxKeysPerIssuer = 4
	MaxPolicyKeys    = 32
	MaxKeyValidity   = 31536000 * time.Second
	PublicKeySize    = 32
)

// TrustedKey is one trusted issuer key record (REQ-TRUST-0051). It enters
// Barometer only inside the trusted local policy loaded at initialization;
// tasks, packages and other delivered content cannot add or change it.
// NotBefore and NotAfter are whole UTC seconds, in the fixed format of a
// package issued_at.
type TrustedKey struct {
	trustedOnly
	KeyID               string
	IssuerID            string
	PublicKey           [PublicKeySize]byte
	NotBefore, NotAfter Timestamp
}

// Covers reports whether t lies within the key's validity period, both ends
// inclusive (REQ-TRUST-0052 check 7).
func (k TrustedKey) Covers(t time.Time) bool {
	return !t.Before(k.NotBefore.Whole) && !t.After(k.NotAfter.Whole)
}

// requirement-ref: REQ-TRUST-0051
// requirement-ref: REQ-LIMIT-0032
func (k TrustedKey) valid() bool {
	if !identifier(k.KeyID) || !identifier(k.IssuerID) || !keyTime(k.NotBefore) || !keyTime(k.NotAfter) {
		return false
	}
	if !k.NotBefore.Whole.Before(k.NotAfter.Whole) || k.NotAfter.Whole.Sub(k.NotBefore.Whole) > MaxKeyValidity {
		return false
	}
	return ValidPublicKey(k.PublicKey)
}

// keyTime is a whole UTC second, the precision of the fixed key time format.
func keyTime(t Timestamp) bool { return t.Fraction == "" && trustTimestamp(t) }

// validKeys checks the key records of one policy against its issuers and the
// registered count limits.
func validKeys(keys []TrustedKey, issuers map[string]bool) bool {
	if len(keys) > MaxPolicyKeys {
		return false
	}
	seen := make(map[string]bool, len(keys))
	perIssuer := make(map[string]int, len(issuers))
	for _, k := range keys {
		if !k.valid() || seen[k.KeyID] || !issuers[k.IssuerID] {
			return false
		}
		seen[k.KeyID] = true
		if perIssuer[k.IssuerID]++; perIssuer[k.IssuerID] > MaxKeysPerIssuer {
			return false
		}
	}
	return true
}

// Edwards25519 constants (RFC 8032 section 5.1).
var (
	fieldP = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 255), big.NewInt(19))
	curveD = func() *big.Int {
		d := new(big.Int).ModInverse(big.NewInt(121666), fieldP)
		d.Mul(d, big.NewInt(-121665))
		return d.Mod(d, fieldP)
	}()
	sqrtM1 = new(big.Int).Exp(big.NewInt(2), new(big.Int).Rsh(new(big.Int).Sub(fieldP, big.NewInt(1)), 2), fieldP)
)

// ValidPublicKey reports whether b is the canonical encoding of an Ed25519
// curve point that is not of small order (REQ-TRUST-0051). It runs only during
// policy validation over at most MaxPolicyKeys keys and is not constant time;
// public keys are not secret.
func ValidPublicKey(b [PublicKeySize]byte) bool {
	x, y, ok := decodePoint(b)
	if !ok {
		return false
	}
	// A point has small order exactly when [8]P is the identity (0, 1).
	for range 3 {
		x, y = pointAdd(x, y, x, y)
	}
	return !(x.Sign() == 0 && y.Cmp(big.NewInt(1)) == 0)
}

// decodePoint follows RFC 8032 section 5.1.3 and also rejects the
// non-canonical encodings that the RFC decoding leaves to implementations:
// a y coordinate not below p, and x = 0 with the sign bit set.
func decodePoint(b [PublicKeySize]byte) (x, y *big.Int, ok bool) {
	sign := b[31] >> 7
	le := b
	le[31] &= 0x7f
	be := make([]byte, PublicKeySize)
	for i := range le {
		be[PublicKeySize-1-i] = le[i]
	}
	y = new(big.Int).SetBytes(be)
	if y.Cmp(fieldP) >= 0 {
		return nil, nil, false
	}
	yy := new(big.Int).Mul(y, y)
	u := new(big.Int).Sub(yy, big.NewInt(1))
	u.Mod(u, fieldP)
	v := new(big.Int).Mul(curveD, yy)
	v.Add(v, big.NewInt(1))
	v.Mod(v, fieldP)
	// x^2 = u / v; v is never zero because d is not a square.
	w := new(big.Int).Mul(u, new(big.Int).ModInverse(v, fieldP))
	w.Mod(w, fieldP)
	exp := new(big.Int).Rsh(new(big.Int).Add(fieldP, big.NewInt(3)), 3)
	x = new(big.Int).Exp(w, exp, fieldP)
	xx := new(big.Int).Mul(x, x)
	xx.Mod(xx, fieldP)
	if xx.Cmp(w) != 0 {
		x.Mul(x, sqrtM1)
		x.Mod(x, fieldP)
		xx.Mul(x, x)
		xx.Mod(xx, fieldP)
		if xx.Cmp(w) != 0 {
			return nil, nil, false
		}
	}
	if x.Sign() == 0 {
		if sign == 1 {
			return nil, nil, false
		}
		return x, y, true
	}
	if uint8(x.Bit(0)) != sign {
		x.Sub(fieldP, x)
	}
	return x, y, true
}

// pointAdd is affine addition on -x^2 + y^2 = 1 + d x^2 y^2. The formulas are
// complete for this curve, so they also double.
func pointAdd(x1, y1, x2, y2 *big.Int) (*big.Int, *big.Int) {
	t := new(big.Int).Mul(x1, x2)
	t.Mul(t, y1)
	t.Mul(t, y2)
	t.Mul(t, curveD)
	t.Mod(t, fieldP)
	nx := new(big.Int).Add(new(big.Int).Mul(x1, y2), new(big.Int).Mul(y1, x2))
	ny := new(big.Int).Add(new(big.Int).Mul(y1, y2), new(big.Int).Mul(x1, x2))
	dx := new(big.Int).Add(big.NewInt(1), t)
	dy := new(big.Int).Sub(big.NewInt(1), t)
	nx.Mul(nx, new(big.Int).ModInverse(dx.Mod(dx, fieldP), fieldP))
	ny.Mul(ny, new(big.Int).ModInverse(dy.Mod(dy, fieldP), fieldP))
	return nx.Mod(nx, fieldP), ny.Mod(ny, fieldP)
}
