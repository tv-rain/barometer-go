// Package admission owns the ordered decision to accept, idempotently
// recognize or reject a task: the delivery-record seam, verification of signed
// remote task packages and the admission pipeline. Nothing here performs
// network activity or establishes runtime conformance.
package admission

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/platform"
)

// Implementation-local bounds for the delivery store. They are not public
// specification limits; the facade selects concrete values within them.
const (
	MaxDeliveryCapacity = 256
	MaxDeliveryLifetime = time.Hour
)

// Install-side errors are fixed sentinels for the reviewed adapter. They carry
// no reference, task bytes, issuer or verifier detail and are never admission
// classes or Measurement Results.
var (
	ErrDeliveryConfig      = errors.New("admission: invalid delivery store configuration")
	ErrDeliveryRejected    = errors.New("admission: delivery record rejected")
	ErrDeliveryCapacity    = errors.New("admission: delivery store full")
	ErrDeliveryUnavailable = errors.New("admission: delivery store unavailable")
	ErrDeliveryClosed      = errors.New("admission: delivery store closed")
)

// DeliveryLimits are the capacity and lifetime of one store.
type DeliveryLimits struct {
	Capacity int
	Lifetime time.Duration
}

// DeliveryStore holds the adapter-installed delivery records of one client
// instance. References are opaque, random, bounded in number and lifetime,
// and resolve only in the store that issued them. Records keep the SHA-256
// digest of the exact bytes, never the payload. A remote record is single-use:
// the first resolution of its reference drops it. The store is safe for
// concurrent use and must not be copied.
type DeliveryStore struct {
	clock  platform.Clock
	limits DeliveryLimits

	mu         sync.Mutex
	closed     bool
	generation uint64
	records    map[string]deliveryEntry
}

type deliveryEntry struct {
	record     task.DeliveryRecord
	generation uint64
	installed  time.Time
	expires    time.Time
}

// refBytes of crypto/rand output form one reference, hex-encoded.
const refBytes = 16

// NewDeliveryStore returns an empty store for the active policy generation.
func NewDeliveryStore(clock platform.Clock, limits DeliveryLimits, generation uint64) (*DeliveryStore, error) {
	if clock == nil || generation == 0 || limits.Capacity < 1 || limits.Capacity > MaxDeliveryCapacity || limits.Lifetime <= 0 || limits.Lifetime > MaxDeliveryLifetime {
		return nil, ErrDeliveryConfig
	}
	return &DeliveryStore{clock: clock, limits: limits, generation: generation, records: make(map[string]deliveryEntry, limits.Capacity)}, nil
}

// InstallEmbedded installs a record for an exact task packaged with a reviewed
// application deployment and covered by its release review (REQ-TRUST-0041).
// The caller is the reviewed adapter; calling it does not perform that review.
func (s *DeliveryStore) InstallEmbedded(ctx context.Context, issuerID string, generation uint64, exactTask []byte) (string, error) {
	source, err := task.NewSourceContext(task.SourceInput{Origin: task.Embedded, IssuerID: issuerID, PolicyGeneration: generation})
	if err != nil {
		return "", ErrDeliveryRejected
	}
	return s.install(ctx, source, exactTask)
}

// InstallRemote installs a record binding remote origin and the policy
// generation to the exact received signed package bytes. The host is a
// transport (REQ-TRUST-0042): it supplies no issuer, issuance time,
// authentication flag, key or verdict, and installing verifies nothing.
// Admission verifies the package against the active policy's trusted keys.
// requirement-ref: REQ-TRUST-0042
// requirement-ref: REQ-TRUST-0044
func (s *DeliveryStore) InstallRemote(ctx context.Context, generation uint64, signedPackage []byte) (string, error) {
	source, err := task.NewSourceContext(task.SourceInput{Origin: task.Remote, PolicyGeneration: generation})
	if err != nil {
		return "", ErrDeliveryRejected
	}
	return s.install(ctx, source, signedPackage)
}

func (s *DeliveryStore) install(ctx context.Context, source task.SourceContext, exactTask []byte) (string, error) {
	record, err := task.NewDeliveryRecord(source, exactTask)
	if err != nil {
		return "", ErrDeliveryRejected
	}
	now, err := s.clock.Now(ctx)
	if err != nil || now.IsZero() {
		return "", ErrDeliveryUnavailable
	}
	var raw [refBytes]byte
	rand.Read(raw[:])
	ref := hex.EncodeToString(raw[:])

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return "", ErrDeliveryClosed
	}
	s.purge(now)
	if len(s.records) >= s.limits.Capacity {
		return "", ErrDeliveryCapacity
	}
	if _, taken := s.records[ref]; taken {
		return "", ErrDeliveryUnavailable
	}
	s.records[ref] = deliveryEntry{record: record, generation: s.generation, installed: now, expires: now.Add(s.limits.Lifetime)}
	return ref, nil
}

// purge removes expired records. The caller holds s.mu.
func (s *DeliveryStore) purge(now time.Time) {
	for ref, e := range s.records {
		if !e.live(now) {
			delete(s.records, ref)
		}
	}
}

// live is false at and after expiry and, failing closed, when the trusted
// clock reads earlier than installation.
func (e deliveryEntry) live(now time.Time) bool {
	return !now.Before(e.installed) && now.Before(e.expires)
}

// Resolve returns the record for ref when it was issued by this store under
// the active policy generation, is unexpired and matches the exact presented
// bytes. A missing, foreign-instance, expired, stale-generation or mismatched
// record is task_source_unauthenticated; unavailable trusted time is
// task_trust_unavailable (step 1 of docs/admission.md). Resolution performs no
// network activity. It does not consume an embedded record. It drops a remote
// record on its first resolution, whatever the outcome, so no digest of a
// signed package outlives the admission that presented it (REQ-TRUST-0053); a
// retried package needs a newly installed reference.
// requirement-ref: REQ-TRUST-0053
func (s *DeliveryStore) Resolve(ctx context.Context, ref string, presented []byte) (task.DeliveryRecord, error) {
	if !wellFormedRef(ref) {
		return task.DeliveryRecord{}, admissionError(task.SourceUnauthenticated)
	}
	now, err := s.clock.Now(ctx)
	s.mu.Lock()
	e, ok := s.records[ref]
	if ok && e.record.Source().Origin() == task.Remote {
		delete(s.records, ref)
	}
	if err != nil || now.IsZero() {
		s.mu.Unlock()
		return task.DeliveryRecord{}, admissionError(task.TrustUnavailable)
	}
	if ok && !e.live(now) {
		delete(s.records, ref)
		ok = false
	}
	ok = ok && !s.closed && e.generation == s.generation
	s.mu.Unlock()
	if !ok || !e.record.Matches(presented) {
		return task.DeliveryRecord{}, admissionError(task.SourceUnauthenticated)
	}
	return e.record, nil
}

// Invalidate drops every record. Data clearing calls it; the store stays
// usable for later installs.
func (s *DeliveryStore) Invalidate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	clear(s.records)
}

// ReplacePolicy drops every record and activates a strictly higher policy
// generation, so records installed under an earlier generation never resolve.
func (s *DeliveryStore) ReplacePolicy(generation uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrDeliveryClosed
	}
	if generation <= s.generation {
		return ErrDeliveryConfig
	}
	clear(s.records)
	s.generation = generation
	return nil
}

// Close drops every record and refuses later installs. It is idempotent.
func (s *DeliveryStore) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	clear(s.records)
}

// String and GoString keep references and digests out of formatted output.
func (s *DeliveryStore) String() string   { return "admission.DeliveryStore" }
func (s *DeliveryStore) GoString() string { return "admission.DeliveryStore" }

func wellFormedRef(ref string) bool {
	if len(ref) != 2*refBytes {
		return false
	}
	for i := 0; i < len(ref); i++ {
		if c := ref[i]; !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

func admissionError(code task.AdmissionCode) error {
	err, _ := task.NewAdmissionError(code)
	return err
}
