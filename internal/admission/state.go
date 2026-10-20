package admission

import (
	"context"
	"crypto/sha256"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/persistence"
)

// Admitted is an accepted task whose replay record is durable. A redelivery of
// the same exact bytes resolves to the same record, so the same admitted task:
// its acceptance, expiry and retention are the original ones. It holds no task
// bytes and has no serialized form. Publishing a handle and scheduling runs
// belong to the lifecycle and execution owners (DEV-592, DEV-576).
type Admitted struct {
	gated      Gated
	record     persistence.ReplayRecord
	redelivery bool
}

// Redelivery reports whether this outcome recognized an earlier acceptance
// rather than creating one. A redelivery must not create a run.
func (a Admitted) Redelivery() bool { return a.redelivery }

// Identity is the replay identity of an admitted task: its replay key and the
// digest of the exact accepted bytes. A redelivery of the same bytes has the
// same identity. It is comparable so that the lifecycle owner can resolve a
// redelivery to the handle it already published (DEV-592). It has no
// serialized form and formats as its type name only.
type Identity struct {
	key    persistence.ReplayKey
	digest [sha256.Size]byte
}

func (Identity) String() string   { return "admission.Identity" }
func (Identity) GoString() string { return "admission.Identity" }

// Identity returns the replay identity this outcome resolved to.
func (a Admitted) Identity() Identity {
	return Identity{key: a.record.Key, digest: a.record.Digest}
}

// ExpiresAt is the end of the admitted task's validity, taken from the
// retained replay record, so a redelivery reports the original expiry.
func (a Admitted) ExpiresAt() time.Time { return a.record.ExpiresAt }

// EvaluatedAt is the trusted time at which this outcome was evaluated.
func (a Admitted) EvaluatedAt() time.Time { return a.gated.acceptedAt }

// String and GoString keep source context, the replay key and parameters out
// of formatted output.
func (Admitted) String() string   { return "admission.Admitted" }
func (Admitted) GoString() string { return "admission.Admitted" }

// Admit runs every admission step. Steps 1 to 9 are Gate. Step 10 checks the
// replay store and step 11 commits the replay record before the task is
// returned as admitted, under the instance's admission serialization:
//
//   - replay state that is unavailable or invalid fails closed with
//     task_trust_unavailable for remote and embedded tasks alike;
//   - an active policy generation lower than the highest retained accepted
//     generation is a policy rollback (policy_rollback);
//   - the same replay key with the same digest is an idempotent redelivery that
//     writes nothing and keeps the original expiry;
//   - the same key with a different digest is task_id_conflict;
//   - a new key is accepted only after its record is durably committed; a failed
//     or uncertain commit returns task_trust_unavailable and no admitted task.
//
// An expired task never reaches step 10: step 6 rejects it as task_expired, so a
// retained record cannot reactivate it.
// requirement-ref: REQ-TRUST-0046
// requirement-ref: REQ-TRUST-0021
// requirement-ref: REQ-TRUST-0024
// requirement-ref: REQ-TRUST-0038
func (a *Admitter) Admit(ctx context.Context, sourceRef string, presented []byte) (Admitted, error) {
	if a == nil {
		return Admitted{}, admissionError(task.TrustUnavailable)
	}
	g, err := a.Gate(ctx, sourceRef, presented)
	if err != nil {
		return Admitted{}, err
	}
	// The replay identity is the exact task bytes: for remote origin the bytes
	// decoded from the verified package, not the package (REQ-TRUST-0053).
	source := g.source
	key := persistence.ReplayKey{IssuerID: source.IssuerID(), TaskID: g.candidate.ID}
	digest := g.digest
	now := g.acceptedAt

	a.accept.Lock()
	defer a.accept.Unlock()
	// Step 10: rollback against retained accepted generations, then the key.
	highest, err := a.replay.HighestGeneration(ctx, now)
	if err != nil {
		return Admitted{}, admissionError(task.TrustUnavailable)
	}
	if highest > g.authority.generation {
		return Admitted{}, admissionError(task.PolicyRollback)
	}
	existing, ok, err := a.replay.Lookup(ctx, key, now)
	if err != nil || ok && (!existing.Valid() || existing.Key != key) {
		return Admitted{}, admissionError(task.TrustUnavailable)
	}
	if ok {
		return redeliver(g, existing, digest)
	}
	// Step 11: the record is durable before the task is admitted.
	notAfter := g.candidate.Schedule.NotAfter
	if notAfter == nil {
		return Admitted{}, admissionError(task.TrustUnavailable)
	}
	rec, err := persistence.NewReplayRecord(key, digest, g.authority.generation, source.Origin(), *notAfter, now)
	if err != nil {
		return Admitted{}, admissionError(task.TrustUnavailable)
	}
	stored, created, err := a.replay.Commit(ctx, rec, now)
	if err != nil || !stored.Valid() || stored.Key != key {
		return Admitted{}, admissionError(task.TrustUnavailable)
	}
	if !created {
		return redeliver(g, stored, digest)
	}
	return Admitted{gated: g, record: rec}, nil
}

// redeliver applies the same-key rows of REQ-TRUST-0024 to a retained record.
func redeliver(g Gated, existing persistence.ReplayRecord, digest [sha256.Size]byte) (Admitted, error) {
	if existing.Digest != digest {
		return Admitted{}, admissionError(task.IDConflict)
	}
	return Admitted{gated: g, record: existing, redelivery: true}, nil
}
