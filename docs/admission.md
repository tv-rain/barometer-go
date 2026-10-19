# Admission Design

## Status and Authority

DEV-583 records the intended design of `internal/admission` and the minimal `internal/registry` it consumes. DEV-597 rebaselines it to Draft `0.1.0` at public revision `d93aff93cbc17e7c065fb40d8f6e2e01a697bf35` (public PR #14 merge), which adds REQ-TRUST-0036 to REQ-TRUST-0038 for [GO-SPEC-ISSUE-008](specification-issues.md#go-spec-issue-008-admission-error-classes-for-unsupported-and-unmapped-rejections). The controlling sources are [Task Trust and Local Policy](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md), [Safety](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md) and [Registered Initial Limits](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md), with coordination decision 0002 (trusted task boundary and registered limits). The public text controls behavior. This record fixes step order, error mapping and component ownership so the admission tasks of the Admission and local policy sub-epic (DEV-584 to DEV-594) do not have to guess them.

This is a design record only. No package, interface, behavior, requirement status or public specification changes here. Neither package exists yet; each is created by the first task that gives it behavior and tests, following the [package-creation rules](package-architecture.md#package-creation-rules). Remote-delivery authentication and issuer/target revocation stay deferred (DEV-595) and must exist before remote task delivery is used outside controlled development. The design needs human review before the dependent tasks move to `to do`.

The public text now names the class of every admission rejection this record previously marked *open*, and makes the first failing step normative ([REQ-TRUST-0037](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0037)). The GO-SPEC-ISSUE-008 disposition is adopted in this record and in `task.AdmissionError`; the admission tasks still have to implement and evidence each row.

## Existing Inputs

Admission composes values that already exist and adds no new trust model:

- `task.SourceContext` and `task.DeliveryRecord` ([Trusted Context and Local Policy Models](trusted-policy-models.md)): origin, issuer, policy generation, authentication outcome, issuance time and the SHA-256 digest of the exact received bytes.
- `task.LocalPolicy`: the immutable issuer permissions, deployment permissions and target registrations of one policy generation.
- `taskdecode.Decoder` and `validation.Common` ([Task Models](task-models.md)): bounded strict decoding, common validation and exact `(test_id, test_version)` codec selection.
- `task.AdmissionError` ([Execution-State and Error Models](execution-models.md#admission-errors)): the ten REQ-TRUST-0034 classes and `task_unsupported` (REQ-TRUST-0036), carrying only a code.
- `platform.Clock`, reached through `config.Initialization` ([Initialization and Configuration Models](configuration-models.md)).

## Ordered Steps and Error Classes

<!-- requirement-ref: REQ-TRUST-0017 -->
<!-- requirement-ref: REQ-TRUST-0034 -->
<!-- requirement-ref: REQ-TRUST-0036 -->
<!-- requirement-ref: REQ-TRUST-0037 -->

Admission runs the eleven steps of [REQ-TRUST-0017](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0017) in public order. The first failing step ends admission, and the rejected task receives exactly one class, that of the first failing step ([REQ-TRUST-0037](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0037)). Classes are those of [REQ-TRUST-0034](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0034) and `task_unsupported` ([REQ-TRUST-0036](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0036)). Where another requirement names the class for a condition, that class applies; the table below follows the REQ-TRUST-0037 condition-to-class table.

| Step | Work | Class on failure |
|---|---|---|
| 1. Receive | Resolve the adapter-installed delivery record for the presented reference in this client instance and compare it with the exact presented bytes (`DeliveryRecord.Matches`). | Missing, expired, foreign-instance or mismatched record: `task_source_unauthenticated`. |
| 2. Provenance | Require `remote` with `authenticated: true`, or reviewed `embedded` origin. | `task_source_unauthenticated`. |
| 3. Time and size | Read the trusted clock; require remote `issued_at` no more than 300 seconds in the future; reject more than 65,536 bytes before parsing; parse with the bounded strict decoder. | Clock unavailable: `task_trust_unavailable`. Remote `issued_at` more than 300 seconds in the future: `task_source_unauthenticated`. Over the byte limit: `task_limit_exceeded`. Malformed JSON: `task_schema_invalid`. |
| 4. Schema | Validate the closed top-level schema, `schema_version` and exact test version. | `task_schema_invalid`. Well-formed but unregistered `(test_id, test_version)` pair: `task_unsupported`. |
| 5. Identifier | Validate `task_id` characters and length. Campaign scope and non-user-specificity cannot be established from syntax; they stay an issuer review obligation (`RV-CC-TRUST-TASK-IDENTIFIER-RULES`). | Characters: `task_schema_invalid`. Length over 128 bytes: `task_limit_exceeded`. |
| 6. Fields and limits | Validate expiry, schedule, retry, execution, safety and every registered numeric limit (`validation.Common`). | Expired: `task_expired`. Value above a registered or stricter local limit: `task_limit_exceeded`. Other invalid structure: `task_schema_invalid`. Unknown risk factor type rejected under REQ-SAFETY-0017 rather than treated as `unknown` risk, or a valid field that requires a capability this build does not provide (for example egress Network Context for `execution.context`): `task_unsupported`. |
| 7. Resolve | Select the exact registry entry; resolve issuer permissions, target registrations and the active policy generation. | No policy installed: `task_policy_unavailable`. Unknown issuer, or test/version not permitted: `task_issuer_not_authorized`. Source generation lower than the active one: `policy_rollback`. Source generation higher than the active one: `task_policy_unavailable`. |
| 8. Atomic validation | Run test-specific input/options validation and match every target tuple, redirect registration, path and composite step against the resolved registrations. | Invalid input or options: `task_schema_invalid`. Unmatched or expired registration, or a direct or proxy path the matched target registration does not permit: `target_not_registered`. Direct or proxy path not permitted by issuer authority or local policy: `task_issuer_not_authorized`. |
| 9. Risk and consent | Compute effective risk and the consent requirement (see [Outputs Consumed by Execution](#outputs-consumed-by-execution)). | Effective risk above the issuer maximum or a local-policy bound, or `unknown` without an explicit reviewed local policy permitting it: `task_issuer_not_authorized`. |
| 10. Replay | Look up `(issuer_id, task_id)` in the replay store. | Different digest: `task_id_conflict`. Store unavailable or corrupt, for remote and embedded origin: `task_trust_unavailable`. Same digest: idempotent result, no error. |
| 11. Accept | Persist the replay and admitted-task records atomically, then publish the handle. | Persistence failure: `task_trust_unavailable`. |

`task_unsupported` describes this build, not the task's validity or safety or the network. A build that provides a capability does not return the class for it; runtime unavailability of a provided capability, such as egress Network Context, remains an unsatisfied execution constraint recorded per attempt ([REQ-CONTEXT-0026](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0026), [REQ-CONTEXT-0028](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0028)).

<!-- requirement-ref: REQ-TRUST-0018 -->

Steps 1 to 9 are pure with respect to Barometer state and the network. They read only the delivery record, clock, registry, policy snapshot and the task bytes. No step performs DNS, Network Context, proxy or measurement activity ([REQ-TRUST-0018](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0018)), so a test can check this with a recording transport that must observe zero calls.

<!-- requirement-ref: REQ-TRUST-0035 -->

Every class is returned as a `task.AdmissionError`, which carries only the code. Decoder, validation, store and platform failures are mapped to a code at the step boundary and never wrapped, so no payload, target, credential, key, signature, participant value or raw platform error reaches the caller ([REQ-TRUST-0035](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0035)). An admission error is never a Measurement Result.

## Consumed Capabilities and Owners

| Capability | Contract used by admission | Owner |
|---|---|---|
| Delivery-record lookup | Resolve a transient reference to one `DeliveryRecord` in this client instance; bounded lifetime and capacity; invalidated on clearing, close and policy replacement. | `internal/admission` (seam, embedded origin and test verifier), DEV-585. Remote authentication and revocation: DEV-595. |
| Trusted clock | `platform.Clock`; unavailable time fails closed. | `internal/platform`, supplied through `internal/config` (existing). |
| Replay store | Read by replay key; one atomic conditional write of the replay and admitted-task records; delete on clearing. | `internal/persistence` port, State, retention and clearing sub-epic (DEV-577). Admission defines only the narrow interface it consumes. |
| Registry | Exact `(test_id, test_version)` lookup returning codec, baseline risk and proxy compatibility. | `internal/registry`, DEV-584. |
| Local policy snapshot | The active immutable `task.LocalPolicy` and its generation. | `internal/model/task` (existing). Activation and replacement by reinitialization: DEV-593. |

The client lifecycle that serializes admission with other transitions and owns handle tables belongs to the Runtime and execution sub-epic (DEV-576). The `Client.AdmitTask` wiring (DEV-592) depends on it.

## Atomic Rejection and Replay Persistence

<!-- requirement-ref: REQ-TRUST-0021 -->
<!-- requirement-ref: REQ-TRUST-0024 -->
<!-- requirement-ref: REQ-TRUST-0038 -->

Rejection is atomic because nothing is written before step 11. Steps 1 to 9 build an immutable, unpublished decision value from owned copies. A failure discards it and leaves the replay store, handle table, schedule accounting and policy unchanged. Validation never clamps, drops or substitutes a value to make a task acceptable.

Acceptance makes a task executable only after its state is durable:

1. Under the client's admission serialization, step 10 reads the replay key. A record with the same digest returns the existing handle without allocating a run or extending expiry. A different digest returns `task_id_conflict`.
2. For a new key, step 11 writes one transaction containing the replay record ([REQ-TRUST-0022](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0022): replay key, SHA-256 of the exact accepted bytes, accepted policy generation, acceptance and expiry classes) and the admitted-task record. The write is conditional on the key still being absent, so concurrent identical deliveries resolve to one record.
3. Only after the transaction commits is the handle published to the in-memory table. A failed or uncertain commit returns `task_trust_unavailable` and publishes nothing.
4. After a restart, the executable set is rebuilt from committed admitted-task records only. A record whose task has expired is purged and never reactivated.

The replay record holds no input, target, label, participant identifier or result ([REQ-TRUST-0023](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0023)). `ClearMeasurementData` deletes both record kinds, so a later redelivery is evaluated as new, within task expiry ([REQ-TRUST-0025](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0025)). Unavailable or invalid replay state fails closed for remote tasks and, under [REQ-TRUST-0038](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0038), for embedded tasks: admission returns `task_trust_unavailable` and never accepts an embedded task without replay state. Persisting the active policy generation across restart, needed to detect rollback after a restart, is an open decision with the State, retention and clearing sub-epic.

## Outputs Consumed by Execution

<!-- requirement-ref: REQ-SAFETY-0023 -->

An accepted task yields an immutable admitted-task value and an opaque local handle. Execution consumes:

- **Admitted-task identity:** replay key `(issuer_id, task_id)`, exact-byte digest, accepted policy generation and origin. Source context never enters traffic or results ([REQ-TRUST-0008](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0008)); `task_id` is emitted unchanged.
- **Validated candidate:** normalized test parameters, schedule, execution constraints, labels and expiry.
- **Effective authority:** the resolved registration IDs, permitted paths and limits after intersecting registered behavior, issuer authority, local policy and task restrictions ([REQ-TRUST-0010](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0010)).
- **Effective risk:** the most restrictive of the registered baseline, task-declared and locally determined risk ([REQ-SAFETY-0023](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0023)). `unknown` stays unordered and is never lowered.
- **Consent requirement:** required when the task, issuer permissions or local policy requires it. A task value of `false` adds no permission ([REQ-SAFETY-0021](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0021)).

Admission records requirements, not eligibility. Safety declarations are evaluated again before every attempt ([REQ-SAFETY-0001](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0001)): execution must check consent status, host and device restrictions, and the policy current at attempt start. No capability for consent status at attempt start exists yet. Pre-contact address and redirect checks ([REQ-TRUST-0016](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0016)) also remain with execution.

## Package-Creation Checklist

### `internal/admission`

- **Responsibility:** the ordered accept, idempotent-recognize or reject decision, which combines decoding, validation and trusted state. It cannot stay in `validation`, which must not mutate replay state, or in the root facade, which must not own policy decisions.
- **Requirements and cases:** REQ-TRUST-0001 to REQ-TRUST-0038 and the admission-time safety rows, through the `CC-TRUST-*` families, the admission rows of the `CC-SAFETY-*` families and `CC-LIMIT-TASK-AND-IDENTIFIER-LIMITS` in the [cross-cutting plan](cross-cutting-requirements.md).
- **Imports:** `internal/model/task`, `internal/taskdecode`, `internal/validation`, `internal/registry`, `internal/platform` (clock) and the narrow persistence interface it consumes. Execution, measurement, protocol, network-context, result, bundle, testkit and the root package are forbidden.
- **Imported by:** the root facade and, for admitted-task values only, `internal/execution`. Validation, models, measurement and protocol packages must not import it.
- **API status:** internal. No exported root symbol, serialized form or binding change without a reviewed public-API task.
- **Cancellation, limits, errors, privacy, clearing:** bounded by the task byte limit and finite record-table capacity; no network activity. Errors are code-only `task.AdmissionError` values, one per rejected task. No source context, payload or target is logged. Clearing deletes replay and admitted-task records and invalidates delivery references and handles.
- **Synthetic tests:** fake delivery records, fake clock, an in-memory test store under `internal/testkit` with fault injection, and a recording transport. No secrets or participant data.

### `internal/registry`

- **Responsibility:** the exact mapping from reviewed `(test_id, test_version)` to codec, baseline risk and proxy compatibility used by admission. Later tasks add runner composition for execution. It cannot stay in `taskdecode`, which receives codecs but does not decide which tests are reviewed.
- **Requirements and cases:** the registered-version rows of each measurement plan, for example REQ-NET-01-0022, and REQ-SAFETY-0004 (baseline risk), through the cases linked from [Measurement Requirements](measurement-requirements.md) and `CC-SAFETY-SAFETY-MODEL`.
- **Imports:** `internal/model/task` and `internal/taskdecode` codec types. Admission, execution, persistence, platform and testkit are forbidden. Measurement packages may be added only when runner composition is introduced.
- **Imported by:** admission, validation (metadata only) and execution. Measurement and protocol packages must not import it.
- **API status:** internal and immutable after construction. There is no runtime registration method, global variable or service locator.
- **Cancellation, limits, errors, privacy, clearing:** no I/O, cancellation or persistent state. Construction rejects duplicates and malformed entries with a fixed error. It holds no participant data and needs no clearing.
- **Synthetic tests:** exact lookup, missing tuples, duplicate rejection and the absence of latest-version fallback. Registry presence is not evidence of support.

## Validation

Admission and registry behavior is validated by the planned case families above when those packages exist. This record is checked with `make check`, `make traceability`, the public requirement-reference checker and `git diff --check`. None of these establishes runtime conformance, platform or field support.
