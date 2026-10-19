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
// instance. It implements steps 1 to 6 (DEV-586), issuer and policy resolution
// in step 7, the input, options and path-authority parts of step 8 and the
// issuer and local risk bounds of step 9 (DEV-587), and target registration
// matching in step 8 (DEV-588). Proxy compatibility, composite steps, locally
// determined risk and consent, replay (step 10) and acceptance (step 11) are
// added by later tasks; nothing here persists state,
// publishes a handle or performs DNS, Network Context, proxy or measurement
// activity. It is safe for concurrent use.
type Admitter struct {
	deliveries *DeliveryStore
	clock      platform.Clock
	registry   *registry.Registry
	decoder    *taskdecode.Decoder
	policy     task.LocalPolicy
}

// NewAdmitter binds the instance's delivery store, trusted clock, reviewed
// registry and active policy snapshot. The task decoder is built from exactly
// the registry's codecs. An invalid (zero) policy is accepted and fails every
// task closed at step 7 as task_policy_unavailable; activating and replacing
// policy belongs to the lifecycle owner.
func NewAdmitter(deliveries *DeliveryStore, clock platform.Clock, reg *registry.Registry, policy task.LocalPolicy) (*Admitter, error) {
	if deliveries == nil || clock == nil || reg == nil {
		return nil, ErrAdmitterConfig
	}
	decoder, err := taskdecode.New(reg.Codecs())
	if err != nil {
		return nil, ErrAdmitterConfig
	}
	return &Admitter{deliveries: deliveries, clock: clock, registry: reg, decoder: decoder, policy: policy}, nil
}

// Gated is the unpublished outcome of the implemented gates: owned values for
// the later steps. It is not an admitted task, holds no task bytes and has no
// serialized form; discarding it leaves no state behind.
type Gated struct {
	record     task.DeliveryRecord
	acceptedAt time.Time
	entry      registry.Entry
	candidate  task.Candidate
	authority  Authority
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
// requirement-ref: REQ-TRUST-0010
// requirement-ref: REQ-TRUST-0012
// requirement-ref: REQ-LIMIT-0026
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
	base := now // validity reference, as in validation.ScheduleAt
	if remote {
		// requirement-ref: REQ-LIMIT-0010
		issued, ok := source.IssuedAt()
		if !ok || issued.AfterTime(now.Add(MaxIssuedAhead)) {
			return Gated{}, admissionError(task.SourceUnauthenticated)
		}
		whole := issued.Whole
		timing.Origin, timing.IssuedAt = "remote", &whole
		if whole.Before(base) {
			base = whole
		}
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
	// Step 7: active policy generation, issuer authority and the intersected
	// test, composite, numeric and label bounds.
	authority, class := resolve(a.policy, source, entry.Test, candidate, base)
	if class != 0 {
		return Gated{}, admissionError(class)
	}
	// Step 8: input and options through the registry-selected codec, the path
	// permitted by issuer and local policy, then every target tuple and redirect
	// registration. Proxy compatibility and composite steps are added later.
	if candidate.Parameters, err = envelope.Parameters(); err != nil {
		return Gated{}, admissionError(fieldClass(err))
	}
	if class = authority.authorizePath(candidate.Execution.NetworkPath); class != 0 {
		return Gated{}, admissionError(class)
	}
	if class = authority.matchTargets(a.policy, entry.Targets, entry.Test, candidate.Parameters, now); class != 0 {
		return Gated{}, admissionError(class)
	}
	// Step 9, issuer and local bounds only: the registered baseline and the
	// task-declared risk.
	if class = authority.authorizeRisk(entry.BaselineRisk, candidate.Safety.RiskLevel); class != 0 {
		return Gated{}, admissionError(class)
	}
	return Gated{record: record, acceptedAt: now, entry: entry, candidate: candidate, authority: authority}, nil
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
