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
// it cannot authenticate a delivery or prove release review. Only a reviewed
// adapter may supply it after establishing provenance for the exact task bytes.
type SourceInput struct {
	trustedOnly
	Origin           Origin
	IssuerID         string
	PolicyGeneration uint64
	Authenticated    bool
	IssuedAt         *Timestamp
}

// trustedOnly blocks accidental JSON import/export, including on input records.
// It is not a sandbox against a compromised Go host or an authenticated issuer.
type trustedOnly struct{}

func (trustedOnly) MarshalJSON() ([]byte, error) { return nil, ErrTrustEncoding }
func (trustedOnly) UnmarshalJSON([]byte) error   { return ErrTrustEncoding }

// SourceContext is structurally valid provenance metadata, not an admission
// permit. Zero values are invalid. No mutable aliases escape the constructor.
type SourceContext struct {
	trustedOnly
	valid         bool
	origin        Origin
	issuer        string
	generation    uint64
	authenticated bool
	issued        Timestamp
	hasIssued     bool
}

func NewSourceContext(in SourceInput) (SourceContext, error) {
	if !identifier(in.IssuerID) || in.PolicyGeneration == 0 {
		return SourceContext{}, ErrTrustModel
	}
	s := SourceContext{valid: true, origin: in.Origin, issuer: in.IssuerID, generation: in.PolicyGeneration, authenticated: in.Authenticated}
	switch in.Origin {
	case Embedded:
		if in.Authenticated || in.IssuedAt != nil {
			return SourceContext{}, ErrTrustModel
		}
	case Remote:
		if !in.Authenticated || in.IssuedAt == nil || !trustTimestamp(*in.IssuedAt) {
			return SourceContext{}, ErrTrustModel
		}
		s.issued, s.hasIssued = *in.IssuedAt, true
	default:
		return SourceContext{}, ErrTrustModel
	}
	return s, nil
}

func (s SourceContext) Valid() bool                 { return s.valid }
func (s SourceContext) Origin() Origin              { return s.origin }
func (s SourceContext) IssuerID() string            { return s.issuer }
func (s SourceContext) PolicyGeneration() uint64    { return s.generation }
func (s SourceContext) Authenticated() bool         { return s.authenticated }
func (s SourceContext) IssuedAt() (Timestamp, bool) { return s.issued, s.hasIssued }

// MaxDeliveryBytes bounds the exact bytes a delivery record can bind. It is an
// implementation-local receive bound, deliberately above the registered 65,536-byte
// task limit so that admission, not the record, rejects an oversized task as
// task_limit_exceeded before parsing (REQ-TRUST-0017 step 3).
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
