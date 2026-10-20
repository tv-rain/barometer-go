package task

import (
	"crypto/sha256"
	"errors"
	"time"
)

var (
	ErrTrustModel    = errors.New("task: invalid trust model")
	ErrTrustEncoding = errors.New("task: trust values have no JSON representation")
)

type Origin uint8

const (
	Embedded Origin = iota + 1
	Remote
)

// SourceInput is an internal adapter assertion, never a wire schema. Constructing
// it cannot authenticate a delivery or prove release review. A reviewed adapter
// supplies the origin and policy generation, and for embedded origin also the
// issuer (REQ-TRUST-0044). It has no authentication flag, remote issuer or
// issuance time: for remote origin those come only from a signed package that
// admission verifies (REQ-TRUST-0042).
type SourceInput struct {
	trustedOnly
	Origin           Origin
	IssuerID         string
	PolicyGeneration uint64
}

// trustedOnly blocks accidental JSON import/export, including on input records.
// It is not a sandbox against a compromised Go host or an authenticated issuer.
type trustedOnly struct{}

func (trustedOnly) MarshalJSON() ([]byte, error) { return nil, ErrTrustEncoding }
func (trustedOnly) UnmarshalJSON([]byte) error   { return ErrTrustEncoding }

// SourceContext is structurally valid provenance metadata, not an admission
// permit. Zero values are invalid. No mutable aliases escape the constructor.
// A remote context starts unverified, with no issuer or issuance time, and is
// completed only by WithVerifiedPackage.
type SourceContext struct {
	trustedOnly
	valid      bool
	verified   bool
	origin     Origin
	issuer     string
	generation uint64
	issued     Timestamp
	hasIssued  bool
}

// NewSourceContext requires a positive policy generation. Embedded origin
// requires the issuer; remote origin refuses one, because a host-supplied
// remote issuer is not accepted.
// requirement-ref: REQ-TRUST-0044
func NewSourceContext(in SourceInput) (SourceContext, error) {
	if in.PolicyGeneration == 0 {
		return SourceContext{}, ErrTrustModel
	}
	switch in.Origin {
	case Embedded:
		if !identifier(in.IssuerID) {
			return SourceContext{}, ErrTrustModel
		}
		return SourceContext{valid: true, verified: true, origin: Embedded, issuer: in.IssuerID, generation: in.PolicyGeneration}, nil
	case Remote:
		if in.IssuerID != "" {
			return SourceContext{}, ErrTrustModel
		}
		return SourceContext{valid: true, origin: Remote, generation: in.PolicyGeneration}, nil
	}
	return SourceContext{}, ErrTrustModel
}

// WithVerifiedPackage returns the remote context completed with the issuer and
// issuance time of a signed package (REQ-TRUST-0046 step 3). Only admission
// calls it, after the last check of REQ-TRUST-0052 succeeds; calling it
// verifies nothing. The receiver must be a valid unverified remote context and
// issuedAt a whole UTC second.
func (s SourceContext) WithVerifiedPackage(issuerID string, issuedAt Timestamp) (SourceContext, error) {
	if !s.valid || s.verified || s.origin != Remote || !identifier(issuerID) || issuedAt.Fraction != "" || !trustTimestamp(issuedAt) {
		return SourceContext{}, ErrTrustModel
	}
	s.verified, s.issuer, s.issued, s.hasIssued = true, issuerID, issuedAt, true
	return s, nil
}

func (s SourceContext) Valid() bool                 { return s.valid }
func (s SourceContext) Origin() Origin              { return s.origin }
func (s SourceContext) IssuerID() string            { return s.issuer }
func (s SourceContext) PolicyGeneration() uint64    { return s.generation }
func (s SourceContext) IssuedAt() (Timestamp, bool) { return s.issued, s.hasIssued }

// Verified reports whether the issuer is established: for a valid embedded
// context always, and for remote only after WithVerifiedPackage.
func (s SourceContext) Verified() bool { return s.valid && s.verified }

// MaxDeliveryBytes bounds the exact bytes a delivery record can bind. It is an
// implementation-local receive bound, deliberately above the registered
// 65,536-byte task and 98,304-byte signed-package limits, so that admission,
// not the record, rejects an oversized task or package as task_limit_exceeded
// before decoding (REQ-TRUST-0046 step 3, REQ-TRUST-0052 check 1).
const MaxDeliveryBytes = 2 * 65536

// DeliveryRecord binds metadata to byte identity without retaining payloads.
// It does not verify signatures, check time, install a source reference, or
// implement replay. A future reviewed delivery owner must perform those steps.
type DeliveryRecord struct {
	trustedOnly
	source SourceContext
	digest [sha256.Size]byte
}

func NewDeliveryRecord(source SourceContext, exactTask []byte) (DeliveryRecord, error) {
	if !source.Valid() || len(exactTask) == 0 || len(exactTask) > MaxDeliveryBytes {
		return DeliveryRecord{}, ErrTrustModel
	}
	return DeliveryRecord{source: source, digest: sha256.Sum256(exactTask)}, nil
}
func (r DeliveryRecord) Source() SourceContext { return r.source }
func (r DeliveryRecord) Matches(exactTask []byte) bool {
	return r.source.Valid() && len(exactTask) > 0 && len(exactTask) <= MaxDeliveryBytes && r.digest == sha256.Sum256(exactTask)
}

func identifier(s string) bool {
	if len(s) == 0 || len(s) > 64 {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '.' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}
func trustTimestamp(t Timestamp) bool {
	if t.Whole.IsZero() || t.Whole.Location() != time.UTC || t.Whole.Nanosecond() != 0 || t.Whole.Year() < 1 || t.Whole.Year() > 9999 {
		return false
	}
	if len(t.Fraction) > 0 && t.Fraction[len(t.Fraction)-1] == '0' {
		return false
	}
	if len(t.Fraction) > 65536 {
		return false
	}
	for _, c := range t.Fraction {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
