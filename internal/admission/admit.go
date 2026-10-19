package admission

import (
	"context"
	"errors"
	"time"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/registry"
	"github.com/tv-rain/barometer-go/internal/taskdecode"
	"github.com/tv-rain/barometer-go/internal/validation"
)

// Registered admission bounds (REQ-LIMIT-0004, REQ-LIMIT-0010).
const (
	MaxTaskBytes   = 65536
	MaxIssuedAhead = 300 * time.Second
)

// ErrAdmitterConfig is the single construction error. It carries no detail.
var ErrAdmitterConfig = errors.New("admission: invalid admitter configuration")

// Admitter runs the ordered admission steps of REQ-TRUST-0017 for one client
// instance. It currently implements steps 1 to 6 and the input and options
// part of step 8 (DEV-586). Issuer, policy and target resolution (step 7), the
// rest of step 8, risk and consent (step 9), replay (step 10) and acceptance
// (step 11) are added by later tasks; nothing here persists state, publishes a
// handle or performs DNS, Network Context, proxy or measurement activity.
// It is safe for concurrent use.
type Admitter struct {
	deliveries *DeliveryStore
	clock      platform.Clock
	registry   *registry.Registry
	decoder    *taskdecode.Decoder
}

// NewAdmitter binds the instance's delivery store, trusted clock and reviewed
// registry. The task decoder is built from exactly the registry's codecs.
func NewAdmitter(deliveries *DeliveryStore, clock platform.Clock, reg *registry.Registry) (*Admitter, error) {
	if deliveries == nil || clock == nil || reg == nil {
		return nil, ErrAdmitterConfig
	}
	decoder, err := taskdecode.New(reg.Codecs())
	if err != nil {
		return nil, ErrAdmitterConfig
	}
	return &Admitter{deliveries: deliveries, clock: clock, registry: reg, decoder: decoder}, nil
}

// Gated is the unpublished outcome of the implemented gates: owned values for
// the later steps. It is not an admitted task, holds no task bytes and has no
// serialized form; discarding it leaves no state behind.
type Gated struct {
	record     task.DeliveryRecord
	acceptedAt time.Time
	entry      registry.Entry
	candidate  task.Candidate
}

// String and GoString keep source context, labels and parameters out of
// formatted output.
func (Gated) String() string   { return "admission.Gated" }
func (Gated) GoString() string { return "admission.Gated" }

// Gate runs the implemented steps in public order and returns the class of the
// first failing step (REQ-TRUST-0037) as a code-only task.AdmissionError. Step
// failures are mapped at the step boundary and never wrapped (REQ-TRUST-0035).
// Task-controlled authority claims and recipient selectors are unknown fields
// and fail at step 4; no task value is interpolated locally.
// requirement-ref: REQ-TRUST-0001
// requirement-ref: REQ-TRUST-0002
// requirement-ref: REQ-TRUST-0017
// requirement-ref: REQ-TRUST-0018
// requirement-ref: REQ-TRUST-0026
// requirement-ref: REQ-TRUST-0027
// requirement-ref: REQ-TRUST-0033
// requirement-ref: REQ-TRUST-0035
// requirement-ref: REQ-TRUST-0036
// requirement-ref: REQ-TRUST-0037
func (a *Admitter) Gate(ctx context.Context, sourceRef string, exactTask []byte) (Gated, error) {
	if a == nil {
		return Gated{}, admissionError(task.TrustUnavailable)
	}
	// Step 1: the adapter-installed record for these exact bytes.
	record, err := a.deliveries.Resolve(ctx, sourceRef, exactTask)
	if err != nil {
		return Gated{}, err
	}
	// Step 2: authenticated remote origin or reviewed embedded origin.
	// requirement-ref: REQ-TRUST-0004
	source := record.Source()
	remote := source.Origin() == task.Remote
	if !source.Valid() || remote != source.Authenticated() || !remote && source.Origin() != task.Embedded {
		return Gated{}, admissionError(task.SourceUnauthenticated)
	}
	// Step 3: trusted time, issuance freshness, then the byte limit before parsing.
	now, err := a.clock.Now(ctx)
	if err != nil || now.IsZero() {
		return Gated{}, admissionError(task.TrustUnavailable)
	}
	timing := validation.TimingContext{Origin: "embedded", AcceptedAt: now}
	if remote {
		// requirement-ref: REQ-LIMIT-0010
		issued, ok := source.IssuedAt()
		if !ok || issued.AfterTime(now.Add(MaxIssuedAhead)) {
			return Gated{}, admissionError(task.SourceUnauthenticated)
		}
		whole := issued.Whole
		timing.Origin, timing.IssuedAt = "remote", &whole
	}
	// requirement-ref: REQ-LIMIT-0004
	// requirement-ref: REQ-LIMIT-0005
	if len(exactTask) > MaxTaskBytes {
		return Gated{}, admissionError(task.LimitExceeded)
	}
	// Steps 3 and 4: parse, closed top-level schema and exact versions.
	envelope, err := a.decoder.Envelope(exactTask)
	if err != nil {
		if errors.Is(err, taskdecode.ErrUnregistered) {
			return Gated{}, admissionError(task.Unsupported)
		}
		// Only an integer beyond the relaxed shape bound, necessarily above its
		// registered maximum, can be a limit failure here.
		return Gated{}, admissionError(fieldClass(err))
	}
	entry, ok := a.registry.Lookup(envelope.Test())
	if !ok {
		return Gated{}, admissionError(task.Unsupported)
	}
	// Step 5: task_id syntax and length. Campaign scope and non-user-specificity
	// remain an issuer review obligation (RV-CC-TRUST-TASK-IDENTIFIER-RULES).
	// requirement-ref: REQ-TRUST-0019
	if err = validation.TaskID(envelope.TaskID()); err != nil {
		return Gated{}, admissionError(fieldClass(err))
	}
	// Step 6: expiry, schedule, retry, execution, safety and numeric limits.
	candidate, err := envelope.Fields(timing)
	if err != nil {
		return Gated{}, admissionError(fieldClass(err))
	}
	// Step 8, input and options only: the registry-selected codec. Step 7 and
	// the target, proxy and composite parts of step 8 are inserted around it.
	if candidate.Parameters, err = envelope.Parameters(); err != nil {
		return Gated{}, admissionError(fieldClass(err))
	}
	return Gated{record: record, acceptedAt: now, entry: entry, candidate: candidate}, nil
}

// fieldClass maps a decoding or validation failure to its public class. A value
// above a registered maximum is task_limit_exceeded, an expired task
// task_expired, an unrecognized risk factor type task_unsupported, and every
// other failure task_schema_invalid.
func fieldClass(err error) task.AdmissionCode {
	switch {
	case errors.Is(err, validation.ErrExpired):
		return task.Expired
	case errors.Is(err, validation.ErrLimit), errors.Is(err, jsondecode.ErrLimit):
		return task.LimitExceeded
	case errors.Is(err, validation.ErrUnsupported):
		return task.Unsupported
	}
	return task.SchemaInvalid
}
