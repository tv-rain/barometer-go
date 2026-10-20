// Package delivery builds synthetic signed remote task packages for tests
// only. Production code must not import it. Its keys derive from fixed public
// test seeds, are not deployment keys and confer no authority. It replaces the
// retired HMAC test verifier: remote authentication is now the library's
// verification of a signed package (REQ-TRUST-0042), not a host assertion.
package delivery

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/task"
)

// IssuedLayout is the fixed package time format.
const IssuedLayout = "2006-01-02T15:04:05Z"

// Signer holds one synthetic issuer key.
type Signer struct {
	KeyID, IssuerID string
	Seed            [ed25519.SeedSize]byte
}

// NewSigner derives a deterministic synthetic key from a one-byte label.
func NewSigner(keyID, issuerID string, label byte) Signer {
	s := Signer{KeyID: keyID, IssuerID: issuerID}
	for i := range s.Seed {
		s.Seed[i] = label ^ byte(i)
	}
	return s
}

func (s Signer) private() ed25519.PrivateKey { return ed25519.NewKeyFromSeed(s.Seed[:]) }

// PublicKey is the 32-byte public key encoding.
func (s Signer) PublicKey() [task.PublicKeySize]byte {
	var k [task.PublicKeySize]byte
	copy(k[:], s.private().Public().(ed25519.PublicKey))
	return k
}

// Key is the trusted key record for this signer, valid over [notBefore, notAfter].
func (s Signer) Key(notBefore, notAfter time.Time) task.TrustedKey {
	return task.TrustedKey{KeyID: s.KeyID, IssuerID: s.IssuerID, PublicKey: s.PublicKey(),
		NotBefore: task.Timestamp{Whole: notBefore.UTC().Truncate(time.Second)}, NotAfter: task.Timestamp{Whole: notAfter.UTC().Truncate(time.Second)}}
}

// Package holds the decoded members of one package. Tests change a member
// after signing to build hostile packages.
type Package struct {
	IssuerID, KeyID, IssuedAt string
	Task, Signature           []byte
}

// Message is M of REQ-TRUST-0049, built independently of the admission code.
func Message(keyID, issuerID, issuedAt string, exactTask []byte) []byte {
	var m []byte
	for _, part := range [][]byte{[]byte("barometer-signed-task-v1"), []byte(keyID), []byte(issuerID), []byte(issuedAt), exactTask} {
		m = binary.BigEndian.AppendUint32(m, uint32(len(part)))
		m = append(m, part...)
	}
	return m
}

// Sign returns a valid package for exactTask issued at issuedAt.
func (s Signer) Sign(issuedAt time.Time, exactTask []byte) Package {
	p := Package{IssuerID: s.IssuerID, KeyID: s.KeyID, IssuedAt: issuedAt.UTC().Format(IssuedLayout), Task: append([]byte(nil), exactTask...)}
	return s.Resign(p)
}

// Resign replaces the signature with one over the current members.
func (s Signer) Resign(p Package) Package {
	p.Signature = ed25519.Sign(s.private(), Message(p.KeyID, p.IssuerID, p.IssuedAt, p.Task))
	return p
}

// Encode returns the package JSON with canonical unpadded base64url members.
func (p Package) Encode() []byte {
	b, err := json.Marshal(map[string]string{
		"issuer_id": p.IssuerID, "key_id": p.KeyID, "issued_at": p.IssuedAt,
		"task": base64.RawURLEncoding.EncodeToString(p.Task), "signature": base64.RawURLEncoding.EncodeToString(p.Signature),
	})
	if err != nil {
		panic(err)
	}
	return b
}
