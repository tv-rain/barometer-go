package admission

import (
	"context"
	"crypto/sha256"
	"errors"
	"sync"
	"time"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/persistence"
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

// Admitter runs the ordered admission steps of REQ-TRUST-0046 for one client
// instance. It implements steps 1 to 6 (DEV-586), issuer and policy resolution
// in step 7 and the input, options and path-authority parts of step 8
// (DEV-587), target registration matching in step 8 (DEV-588), registry proxy
// compatibility in step 8 (DEV-599), and effective risk with the consent
// requirement in step 9 (DEV-589). Tasks that require egress Network Context
// (step 6) or NET-05 composite steps (step 8) are rejected as task_unsupported
// until those capabilities exist (DEV-590). Admit adds replay (step 10) and
// acceptance (step 11) through the replay store (DEV-591). Nothing here
// publishes a handle or performs DNS, Network Context, proxy or measurement
// activity. It is safe for concurrent use.
type Admitter struct {
	deliveries *DeliveryStore
	clock      platform.Clock
	registry   *registry.Registry
	decoder    *taskdecode.Decoder
	policy     task.LocalPolicy
	replay     persistence.ReplayStore
	// accept serializes steps 10 and 11 within this instance.
	accept sync.Mutex
}

// NewAdmitter binds the instance's delivery store, trusted clock, reviewed
// registry, active policy snapshot and replay store. The task decoder is built
// from exactly the registry's codecs. An invalid (zero) policy is accepted and
// fails every task closed at step 7 as task_policy_unavailable; activating and
// replacing policy belongs to the lifecycle owner. A replay store is required:
// no task is accepted without replay state (REQ-TRUST-0038).
func NewAdmitter(deliveries *DeliveryStore, clock platform.Clock, reg *registry.Registry, policy task.LocalPolicy, replay persistence.ReplayStore) (*Admitter, error) {
	if deliveries == nil || clock == nil || reg == nil || replay == nil {
		return nil, ErrAdmitterConfig
	}
	decoder, err := taskdecode.New(reg.Codecs())
	if err != nil {
		return nil, ErrAdmitterConfig
	}
	return &Admitter{deliveries: deliveries, clock: clock, registry: reg, decoder: decoder, policy: policy, replay: replay}, nil
}

// Gated is the unpublished outcome of the implemented gates: owned values for
// the later steps, including the effective risk and consent requirement that
// acceptance records on the admitted task. It is not an admitted task, holds no task bytes and has no
// serialized form; discarding it leaves no state behind. It keeps the verified
// source context and the digest of the exact task bytes, never the delivery
// record, so nothing derived from a signed package, its signature or its key
// ID outlives admission (REQ-TRUST-0053).
type Gated struct {
	source     task.SourceContext
	digest     [sha256.Size]byte
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
// first failing step (REQ-TRUST-0047) as a code-only task.AdmissionError. Step
// failures are mapped at the step boundary and never wrapped (REQ-TRUST-0035).
// Task-controlled authority claims and recipient selectors are unknown fields
// and fail at step 4; no task value is interpolated locally. presented is the
// exact serialized task for embedded origin and the exact signed remote task
// package for remote origin (REQ-TRUST-0046 step 1).
// requirement-ref: REQ-TRUST-0040
// requirement-ref: REQ-TRUST-0043
// requirement-ref: REQ-TRUST-0046
// requirement-ref: REQ-TRUST-0018
// requirement-ref: REQ-TRUST-0026
// requirement-ref: REQ-TRUST-0027
// requirement-ref: REQ-TRUST-0033
// requirement-ref: REQ-TRUST-0035
// requirement-ref: REQ-TRUST-0036
// requirement-ref: REQ-TRUST-0047
// requirement-ref: REQ-TRUST-0010
// requirement-ref: REQ-TRUST-0012
// requirement-ref: REQ-LIMIT-0026
func (a *Admitter) Gate(ctx context.Context, sourceRef string, presented []byte) (Gated, error) {
	if a == nil {
		return Gated{}, admissionError(task.TrustUnavailable)
	}
	// Step 1: the adapter-installed record for these exact presented bytes: the
	// serialized task for embedded origin, the signed package for remote.
	record, err := a.deliveries.Resolve(ctx, sourceRef, presented)
	if err != nil {
		return Gated{}, err
	}
	// Step 2: reviewed embedded origin, or a remote signed package verified
	// under REQ-TRUST-0052. Nothing the host asserts authenticates a remote task.
	// requirement-ref: REQ-TRUST-0041
	// requirement-ref: REQ-TRUST-0042
	source, exactTask := record.Source(), presented
	var now time.Time
	switch {
	case source.Valid() && source.Origin() == task.Embedded && source.Verified():
	case source.Valid() && source.Origin() == task.Remote && !source.Verified():
		var class task.AdmissionCode
		if source, exactTask, now, class = a.verifyPackage(ctx, source, presented); class != 0 {
			return Gated{}, admissionError(class)
		}
	default:
		return Gated{}, admissionError(task.SourceUnauthenticated)
	}
	// Step 3: trusted time, then the byte limit before parsing. For remote
	// origin the acceptance time is the reading of check 6, and issuer_id and
	// issued_at come from the verified package.
	if source.Origin() == task.Embedded {
		if now, err = a.clock.Now(ctx); err != nil || now.IsZero() {
			return Gated{}, admissionError(task.TrustUnavailable)
		}
	}
	timing := validation.TimingContext{Origin: "embedded", AcceptedAt: now}
	base := now // validity reference, as in validation.ScheduleAt
	if issued, ok := source.IssuedAt(); ok {
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
	// A valid execution.context requirement needs egress Network Context,
	// which this build does not provide.
	if requiresEgressContext(candidate) {
		return Gated{}, admissionError(task.Unsupported)
	}
	// Step 7: active policy generation, issuer authority and the intersected
	// test, composite, numeric and label bounds.
	authority, class := resolve(a.policy, source, entry.Test, candidate, base)
	if class != 0 {
		return Gated{}, admissionError(class)
	}
	// Step 8: input and options through the registry-selected codec, proxy
	// compatibility, the path permitted by issuer and local policy, then every
	// target tuple and redirect registration (REQ-TRUST-0039).
	if candidate.Parameters, err = envelope.Parameters(); err != nil {
		return Gated{}, admissionError(fieldClass(err))
	}
	// Composite steps cannot be matched yet, so an enabled NET-05 task fails
	// closed here rather than being admitted with its steps unchecked.
	if requiresCompositeSteps(entry) {
		return Gated{}, admissionError(task.Unsupported)
	}
	// A proxy task whose test, input or options HTTP CONNECT cannot carry is
	// invalid for its selected path, whatever authority would permit
	// (REQ-TRUST-0039). It is never admitted to end in proxy_path_unsupported
	// at execution (REQ-PROXY-0042).
	if candidate.Execution.NetworkPath == "proxy" && !entry.Proxy.Compatible(candidate.Parameters) {
		return Gated{}, admissionError(task.SchemaInvalid)
	}
	if class = authority.authorizePath(candidate.Execution.NetworkPath); class != 0 {
		return Gated{}, admissionError(class)
	}
	if class = authority.matchTargets(a.policy, entry.Targets, entry.Test, candidate.Parameters, now); class != 0 {
		return Gated{}, admissionError(class)
	}
	// Step 9: effective risk from the registered baseline, the declared level
	// and the matched registrations, against the intersected risk bounds.
	if class = authority.evaluateRisk(a.policy, entry.BaselineRisk, candidate.Safety.RiskLevel); class != 0 {
		return Gated{}, admissionError(class)
	}
	return Gated{source: source, digest: sha256.Sum256(exactTask), acceptedAt: now, entry: entry, candidate: candidate, authority: authority}, nil
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
