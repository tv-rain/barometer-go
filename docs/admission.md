# Admission Design

## Status and Authority

DEV-583 records the intended design of `internal/admission` and the minimal `internal/registry` it consumes. DEV-597 rebaselines it to Draft `0.1.0` at public revision `d93aff93cbc17e7c065fb40d8f6e2e01a697bf35` (public PR #14 merge), which adds REQ-TRUST-0036 to REQ-TRUST-0038 for [GO-SPEC-ISSUE-008](specification-issues.md#go-spec-issue-008-admission-error-classes-for-unsupported-and-unmapped-rejections). The controlling sources are [Task Trust and Local Policy](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md), [Safety](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md) and [Registered Initial Limits](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md), with coordination decision 0002 (trusted task boundary and registered limits). The public text controls behavior. This record fixes step order, error mapping and component ownership so the admission tasks of the Admission and local policy sub-epic (DEV-584 to DEV-594) do not have to guess them.

This is a design record only. No package, interface, behavior, requirement status or public specification changes here. `internal/registry` now exists with its admission subset (DEV-584, see [Registry](package-architecture.md#registry)). `internal/admission` exists with the delivery-record seam (DEV-585, see [Delivery-Record Seam](#delivery-record-seam)) the gates for steps 1 to 6 and the input and options part of step 8 (DEV-586, see [Implemented Gates](#implemented-gates)), and issuer and policy resolution in step 7 with the path-authority part of step 8 and the issuer and local risk bounds of step 9 (DEV-587, see [Issuer Authorization](#issuer-authorization)), and target registration matching in step 8 (DEV-588, see [Target Registration](#target-registration)); proxy compatibility and composite steps in step 8, locally determined risk and consent in step 9, and steps 10 and 11 do not exist yet. Each package is created by the first task that gives it behavior and tests, following the [package-creation rules](package-architecture.md#package-creation-rules). Remote-delivery authentication and issuer/target revocation stay deferred (DEV-595) and must exist before remote task delivery is used outside controlled development. The design needs human review before the dependent tasks move to `to do`.

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
| 7. Resolve | Select the exact registry entry; resolve issuer permissions, target registrations and the active policy generation. | No policy installed: `task_policy_unavailable`. Unknown issuer, or test/version not permitted: `task_issuer_not_authorized`. Source generation lower than the active one: `policy_rollback`. Source generation higher than the active one: `task_policy_unavailable`. Value beyond an issuer or local ceiling: `task_limit_exceeded`. Label outside the permitted namespaces: `task_issuer_not_authorized`. |
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
| Registry | Exact `(test_id, test_version)` lookup returning codec, baseline risk, target rule and proxy compatibility. | `internal/registry`, DEV-584. |
| Local policy snapshot | The active immutable `task.LocalPolicy` and its generation. | `internal/model/task` (existing). Activation and replacement by reinitialization: DEV-593. |

The client lifecycle that serializes admission with other transitions and owns handle tables belongs to the Runtime and execution sub-epic (DEV-576). The `Client.AdmitTask` wiring (DEV-592) depends on it.

## Delivery-Record Seam

<!-- requirement-ref: REQ-TRUST-0004 -->
<!-- requirement-ref: REQ-TRUST-0005 -->
<!-- requirement-ref: REQ-TRUST-0006 -->
<!-- requirement-ref: REQ-TRUST-0007 -->
<!-- requirement-ref: REQ-TRUST-0008 -->

DEV-585 implements the delivery-record lookup consumed by step 1 in `internal/admission/delivery.go`. `admission.DeliveryStore` belongs to one client instance. It is created with a `platform.Clock`, a capacity of 1 to 256 records, a record lifetime of more than zero and at most one hour, and the active policy generation. These bounds are implementation-local, not public limits; the facade selects concrete values within them.

- **Installation.** `InstallEmbedded` builds an embedded-origin `task.SourceContext` (issuer, policy generation, `authenticated: false`, no `issued_at`) for an exact task packaged with a reviewed application deployment ([REQ-TRUST-0004](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0004)). `InstallRemote` installs a record only for the exact bytes a `RemoteVerifier` authenticated and only when it returns an authenticated remote source with `issued_at`. Both bind the source to the bytes through `task.NewDeliveryRecord`, which keeps the SHA-256 digest and never the payload ([REQ-TRUST-0006](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0006), [REQ-TRUST-0007](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0007)). Each returns a fresh reference of 128 random bits from `crypto/rand`, lowercase hex. Expired records are purged before the capacity check; a full store rejects the install and never evicts a live record. Installation needs trusted time. Install failures are fixed `ErrDelivery*` sentinels for the adapter, not admission classes.
- **No production remote verifier.** Nothing in production implements `RemoteVerifier`; remote authentication and revocation remain DEV-595. The synthetic verifier in `internal/testkit/delivery` models an authenticated channel with an HMAC envelope. Its transport-encryption-without-issuer-authentication mode returns no source, so no record is installed ([REQ-TRUST-0005](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0005)).
- **Resolution.** `Resolve(ctx, ref, exactTask)` returns the record only when this store issued the reference under the active policy generation, the record is unexpired and it matches the exact presented bytes. A malformed, forged, foreign-instance, invalidated, expired, stale-generation or byte-mismatched reference returns `task_source_unauthenticated`, as in step 1. A trusted clock that is unavailable returns `task_trust_unavailable`, and one that reads earlier than installation fails closed as an expired record. Resolution does not consume the record: exact redelivery is handled by the replay store at step 10.
- **Stale generation.** A record is stale when it was installed while an earlier policy generation was active. This is separate from the source's own `policy_generation`, which step 7 compares with the active policy and which can still yield `policy_rollback` or `task_policy_unavailable`.
- **Invalidation.** `Invalidate` drops every record for data clearing; `ReplacePolicy` drops every record and activates a strictly higher generation; `Close` drops every record and refuses later installs. The facade and lifecycle tasks call them; the store does not observe the client lifecycle itself.
- **Privacy.** References and records are not task fields and have no JSON form. Formatting the store prints only its type name, and errors carry only fixed text or the admission code, so references, issuers and task bytes cannot reach traffic, results, logs or persistent identifiers through this seam ([REQ-TRUST-0008](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0008)). Root-facade and native resolution of `sourceRef` remain DEV-592 and DEV-455.

Synthetic tests in `internal/admission/delivery_test.go` cover forged, missing, malformed, expired, foreign-instance, stale-generation, one-byte-mismatch and closed cases, capacity and lifetime boundaries with the manual test clock, unavailable time, unproven remote deliveries, leakage canaries and concurrent install, resolution and invalidation under the race detector. They exercise the seam only. The `CC-TRUST-ACCEPTED-TASK-ORIGINS` and `CC-TRUST-TRUSTED-SOURCE-CONTEXT` families need the admission pipeline and stay `planned`.

## Implemented Gates

<!-- requirement-ref: REQ-TRUST-0019 -->
<!-- requirement-ref: REQ-TRUST-0027 -->
<!-- requirement-ref: REQ-LIMIT-0004 -->
<!-- requirement-ref: REQ-LIMIT-0005 -->

DEV-586 implements the first admission steps in `internal/admission/admit.go`. `admission.Admitter` belongs to one client instance and binds its `DeliveryStore`, trusted `platform.Clock` and reviewed `registry.Registry`; its task decoder is built from exactly the registry's codecs. `Gate(ctx, sourceRef, exactTask)` runs the steps below in public order and returns either an unpublished `Gated` value or the code-only `task.AdmissionError` of the first failing step. `Gated` holds the resolved record, trusted acceptance time, registry entry and validated candidate for the later steps. It is not an admitted task, holds no task bytes, formats as its type name and has no serialized form. Nothing is persisted or published, so discarding it, or any rejection, leaves no state behind and leaves the delivery record resolvable.

- **Step 1** resolves the delivery record for the exact bytes (see [Delivery-Record Seam](#delivery-record-seam)).
- **Step 2** accepts only authenticated remote origin or unauthenticated embedded origin. The store cannot hold any other combination, so this check is defense in depth.
- **Step 3** reads trusted time once; that reading is the acceptance time. Remote `issued_at` more than 300 seconds after it, including a fractional excess, is `task_source_unauthenticated`. Then more than 65,536 original bytes is `task_limit_exceeded`, before parsing. Delivery records bind up to 131,072 bytes (an implementation-local receive bound) so that an oversized task reaches this check rather than failing at step 1.
- **Step 4** parses and validates the closed top-level shape and value kinds, with every field value range, count and byte cap relaxed and `input` and `options` treated as opaque objects. It then requires `schema_version` `0.1.0` and selects the exact `(test_id, test_version)`. A well-formed pair the registry does not hold is `task_unsupported`; a malformed pair or other schema version is `task_schema_invalid`. Shape is validated before selection, so a task that is not valid under the public task schema is `task_schema_invalid` even when its pair is unregistered. An integer above the relaxed bound, necessarily above its registered maximum, is `task_limit_exceeded`.
- **Step 5** checks `task_id` on the exact string: more than 128 bytes is `task_limit_exceeded` whatever the characters; empty, any character outside lowercase ASCII letters, digits, `.`, `_` and `-`, or a leading `.`, `_` or `-` (the public task format pattern) is `task_schema_invalid`. Campaign scope and non-user-specificity remain the issuer review obligation `RV-CC-TRUST-TASK-IDENTIFIER-RULES`.
- **Step 6** applies every common field value, registered limit and default, then expiry against the acceptance time. A value above a registered maximum, including `input` over 49,152 and `options` over 8,192 encoded bytes, labels over 16 entries or 2,048 encoded bytes and `not_after` beyond the seven-day lifetime, is `task_limit_exceeded`. `not_after` at or before acceptance is `task_expired`. An unrecognized risk factor type, rejected under REQ-SAFETY-0017 rather than treated as `unknown` risk, is `task_unsupported`. Every other invalid value, including a value below a registered minimum, is `task_schema_invalid`. Embedded tasks without `not_after` are rejected, because no reviewed configuration for the one-time embedded exception exists yet.
- **Step 8, input and options** validates them against the selected test's schemas and normalizes them with its codec; failures are `task_schema_invalid`, or `task_limit_exceeded` for a stricter test byte cap. DEV-587 adds step 7 before it and the path-authority part after it ([Issuer Authorization](#issuer-authorization)); DEV-588 adds the target part ([Target Registration](#target-registration)); the proxy-compatibility and composite parts are not implemented yet.

Within one step, the order of checks above decides which class is returned when a task fails more than one check; REQ-TRUST-0037 orders only steps. Decoding and validation failures are mapped at the step boundary through refining classes (`jsondecode.ErrLimit`, `validation.ErrLimit`, `validation.ErrExpired`, `validation.ErrUnsupported` and `taskdecode.ErrUnregistered`) that still match their coarse classes, so existing decoder callers are unaffected. The staged decoder is described in [Task Models](task-models.md).

Task-controlled authority claims and recipient selectors are unknown fields and fail at step 4. Placeholder text in identifiers, labels, input and options is kept literal: the library performs no local interpolation of deployment, application-version, cohort, country or network values. Broad eligibility applied by host policy, and its review, remain host obligations the library cannot enforce (`RV-CC-TRUST-TARGETED-DELIVERY`). `Gate` takes the presented bytes as a slice; reading them through a counting reader that stops at 65,537 bytes belongs to the facade or native adapter that receives them (DEV-592, DEV-455).

Synthetic tests in `internal/admission/admit_test.go` cover each step's classes and boundaries, first-failing-step precedence, origins, `task_id` rules, task-controlled authority and selector fields, literal placeholders, canary leakage, ownership and concurrency, with an unused recording transport that must see zero calls. `FuzzGate` checks that arbitrary bound bytes yield either a gate outcome or exactly one class, deterministically. They exercise the implemented steps only; the planned `CC-TRUST-*` and `CC-LIMIT-TASK-AND-IDENTIFIER-LIMITS` families also need the later steps and stay `planned`.

## Issuer Authorization

<!-- requirement-ref: REQ-TRUST-0010 -->
<!-- requirement-ref: REQ-TRUST-0011 -->
<!-- requirement-ref: REQ-TRUST-0012 -->
<!-- requirement-ref: REQ-LIMIT-0026 -->
<!-- requirement-ref: REQ-LIMIT-0030 -->

DEV-587 implements effective authority ([REQ-TRUST-0010](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0010)) in `internal/admission/authority.go`. `NewAdmitter` also binds the instance's active `task.LocalPolicy` snapshot; an invalid (zero) snapshot models an instance with no installed policy. Activating and replacing policy remains with DEV-593. Issuer permissions and deployment permissions are applied independently and intersected: a permission is effective only when both grant it, a maximum is the lower of the two and a minimum the higher. Nothing is a union or an implicit default, and no value is clamped. The intersection is carried on `Gated` as an unpublished `Authority` value (formatted as its type name, with no serialized form) for the later steps and for execution.

- **Step 7** first requires a valid policy snapshot (`task_policy_unavailable`), then compares the source context's `policy_generation` with the active one: lower is `policy_rollback`, higher is `task_policy_unavailable`. It then resolves the source issuer by exact ID and requires the exact `(test_id, test_version)` in both the issuer's and the deployment's test list; otherwise `task_issuer_not_authorized`. NET-05 additionally needs composite enablement on both sides; the policy model already refuses a NET-05 entry without it, so this check is defense in depth.
- **Step 7, numeric ceilings.** Each task value bounded by a policy ceiling is compared with the intersected ceiling: attempt duration, bytes sent, bytes received and connections; `max_runs`; retry `max_attempts`; the validity lifetime from the earlier of issuance and acceptance, as in expiry validation; repeat interval and jitter minima; and retry initial delay minimum, maximum delay and jitter maximum. A value beyond a ceiling, including an interval, jitter or retry delay below a policy minimum, raises authority ([REQ-TRUST-0012](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0012), [REQ-LIMIT-0030](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0030)) and is `task_limit_exceeded`. A value within a ceiling becomes the effective value. Result, bundle, pending-storage and retention ceilings have no task field; they are carried intersected for result and storage enforcement. The registered maxima were already applied at step 6, and the policy model refuses any host ceiling above them.
- **Step 7, labels.** Every label key must lie in a namespace permitted by the issuer and, separately, by the deployment: equal to a listed namespace or below it at a `.` boundary, as with the reserved `barometer.` namespace ([REQ-LABEL-0011](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0011)). `campaign` permits `campaign` and `campaign.wave`, not `campaignx`. An empty list permits no labels. A key outside is `task_issuer_not_authorized`. The public text does not define namespace matching further; this reading is a Go-local choice that admits no key the literal reviewed names do not cover.
- **Step 8, path authority** runs after input and options: the task's `network_path` must be permitted by both sides, otherwise `task_issuer_not_authorized` ([REQ-TRUST-0037](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0037)). The effective paths narrow to the task's one path. The matched registration's paths are checked by target matching (`target_not_registered`, [Target Registration](#target-registration)); registry proxy compatibility is not implemented yet.
- **Step 9, issuer and local risk bounds** apply to the risk known without local determination: the more restrictive of the registered baseline and the task-declared level. `unknown` from either needs the reviewed unknown-risk exception on both sides. An ordered risk above the lower maximum, or a medium or high risk without medium-risk enablement on both sides, is `task_issuer_not_authorized`. Treating high risk as requiring medium-risk enablement is the fail-closed reading of the opt-in. Locally determined risk, risk factors and consent evaluation remain DEV-589.
- **Carried authority.** The intersected registration IDs, narrowed by target matching to those the task matched, the narrowed ceilings and path, the risk bounds and the consent requirement, which is required when the issuer, the deployment or the task requires it.

Platform capability, the fourth term of the intersection, is applied where a capability is missing (DEV-590, `task_unsupported`); this task adds no capability check.

Synthetic tests in `internal/admission/authority_test.go` narrow every dimension in the issuer and then in the deployment while the other stays permissive, cover limit −1, limit and limit +1 for every task-bounded ceiling with the lower value carried as effective, issuer-versus-deployment conflicts in both directions, generation lower, equal and higher, unknown issuers, remote sources, precedence against steps 4 to 8, and the admission-time rows of `CC-LIMIT-LIMIT-HANDLING`, including attempts to raise a limit through labels, input, options, schedule, execution, safety, an unknown field and host configuration. All use the unused recording transport that must see zero calls. The NET-05 nine-step row of `CC-LIMIT-LIMIT-HANDLING` needs the NET-05 codec (DEV-454) and is not covered. `CC-TRUST-ISSUER-AUTHORIZATION` stays `planned`: its registration-reference row is now covered by target matching ([Target Registration](#target-registration)), but full acceptance needs steps 10 and 11. `CC-LIMIT-LIMIT-HANDLING` stays `planned` for its runtime rows (DEV-576).

## Target Registration

<!-- requirement-ref: REQ-TRUST-0013 -->
<!-- requirement-ref: REQ-TRUST-0014 -->
<!-- requirement-ref: REQ-TRUST-0015 -->
<!-- requirement-ref: REQ-TRUST-0016 -->

DEV-588 implements the target part of step 8 ([REQ-TRUST-0013](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0013) to [REQ-TRUST-0015](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0015)) in `internal/admission/target.go`. It runs after input and options and path authority, and before the step-9 risk bounds.

- **Tuples come from the registered test.** Each registry entry carries a required `registry.TargetRule`. Given the codec's normalized parameters, it reports every destination or name the task would contact as a role and a `task.TargetTuple` in normalized typed form. The rule owns that test's complete safety-relevant tuple, for example address, port and scheme for a TCP connection, plus server name and ALPN for TLS. It reports false for parameters of an unexpected type or an incomplete tuple. A rule is pure and bounded and never resolves, fetches or rewrites a target. Task values are never interpreted as registration IDs.
- **Matching.** Every reported tuple must match one registration that satisfies all of the following:
  - its ID is permitted by both the issuer and local policy;
  - it is for the exact test/version and the same role;
  - its tuple is exactly equal to the task's, compared on every component, including the order of bootstrap addresses, ALPN values and checks; an absent component matches only an absent one;
  - its expiry is after trusted `now`;
  - it permits the task's single path.

  There is no substring, suffix, wildcard, case-folding or resolution rule at matching time. Normalization is the codec's job, so two inputs match only when the codec normalizes both to the same typed value.
- **Redirects.** A redirect-role tuple must also be listed in `RedirectIDs` of a non-redirect registration this task matched. Like every other registration, it must be permitted by both the issuer and local policy. That conservative reading means an issuer must list redirect registrations explicitly.
- **Failure class.** Any failure is `target_not_registered`: an unmatched, expired or wrong-role registration, an unlisted redirect, a path the matched registration does not permit, no measurement-role tuple, a rule that reports false, or more than 64 tuples. 64 is a finite implementation-local bound, not a public maximum. Path refusal by issuer authority or local policy is still `task_issuer_not_authorized` and happens first ([REQ-TRUST-0037](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0037)).
- **Carried authority.** The carried registration IDs narrow to the matched ones, so execution receives only registrations this task actually matched.
- **Execution-time obligation.** DNS resolution and redirects do not extend authority ([REQ-TRUST-0016](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0016)). Admission performs no DNS, redirect or network activity, so it cannot check resolved addresses or redirect locations. Immediately before every contact, including on the proxy path, execution must:
  - check each resolved address and redirect target against the matched registrations and local address policy;
  - re-check that the registrations have not expired;
  - make no contact on failure.

  This obligation belongs to Measurement implementation (DEV-454) and Network paths and context (DEV-578) and has no implementation yet.

Registrations with credentials, interpolation syntax, wildcard or noncanonical hosts, or IPv4-mapped addresses are already refused when the policy snapshot is built ([Trusted Context and Local Policy Models](trusted-policy-models.md#target-representation-and-limits)). Registry proxy compatibility and NET-05 composite steps remain unimplemented parts of step 8. The public text does not assign a class to proxy-incompatible parameters, and composite steps need the NET-05 codec (DEV-454).

Synthetic tests in `internal/admission/target_test.go` use test-scoped codecs and rules under the TCP-01, TLS-01, HTTP-01 and DNS-01 IDs. These are not the public test schemas or production codecs. The tests cover:

- an exact match for each test, with the carried IDs;
- single-component mutations of every tuple component, a redirect target and direct versus proxy path;
- normalized forms that match, against suffix, subdomain, parent-domain, trailing-dot, wildcard, URL path-prefix, URL host-suffix and IPv4-mapped forms that do not;
- expiry at one second after, at and one second before trusted `now`, including an expired redirect registration;
- issuer, deployment, test/version, role, path and redirect-listing failures;
- incomplete tuples and the target bound;
- precedence against steps 6, 4 and the path authority of step 8, and before step 9;
- policy refusal of credential and interpolation registrations.

All use the unused recording transport, which must see zero calls. `CC-TRUST-TARGET-REGISTRATION` stays `planned`. Its plan asserts acceptance, which needs steps 10 and 11. Its execution-owned resolver and redirect rows need the execution work above.

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
- **Imports:** `internal/model/task`, `internal/taskdecode`, `internal/validation`, `internal/registry`, `internal/jsondecode` (error classes only), `internal/platform` (clock) and the narrow persistence interface it consumes. Execution, measurement, protocol, network-context, result, bundle, testkit and the root package are forbidden.
- **Imported by:** the root facade and, for admitted-task values only, `internal/execution`. Validation, models, measurement and protocol packages must not import it.
- **API status:** internal. No exported root symbol, serialized form or binding change without a reviewed public-API task.
- **Cancellation, limits, errors, privacy, clearing:** bounded by the task byte limit and finite record-table capacity; no network activity. Errors are code-only `task.AdmissionError` values, one per rejected task. No source context, payload or target is logged. Clearing deletes replay and admitted-task records and invalidates delivery references and handles.
- **Synthetic tests:** fake delivery records, fake clock, an in-memory test store under `internal/testkit` with fault injection, and a recording transport. No secrets or participant data.

### `internal/registry`

- **Responsibility:** the exact mapping from reviewed `(test_id, test_version)` to codec, baseline risk, target rule and proxy compatibility used by admission. Later tasks add runner composition for execution. It cannot stay in `taskdecode`, which receives codecs but does not decide which tests are reviewed.
- **Requirements and cases:** the registered-version rows of each measurement plan, for example REQ-NET-01-0022, and REQ-SAFETY-0004 (baseline risk), through the cases linked from [Measurement Requirements](measurement-requirements.md) and `CC-SAFETY-SAFETY-MODEL`.
- **Imports:** `internal/model/task` and `internal/taskdecode` codec types. Admission, execution, persistence, platform and testkit are forbidden. Measurement packages may be added only when runner composition is introduced.
- **Imported by:** admission and execution. Validation receives registry metadata as values from admission and must not import it, because `taskdecode` already imports validation. Measurement and protocol packages must not import it.
- **API status:** internal and immutable after construction. There is no runtime registration method, global variable or service locator.
- **Cancellation, limits, errors, privacy, clearing:** no I/O, cancellation or persistent state. Construction rejects duplicates and malformed entries with a fixed error. It holds no participant data and needs no clearing.
- **Synthetic tests:** exact lookup, missing tuples, duplicate rejection and the absence of latest-version fallback. Registry presence is not evidence of support.

## Validation

Admission and registry behavior is validated by the planned case families above when those packages exist. This record is checked with `make check`, `make traceability`, the public requirement-reference checker and `git diff --check`. None of these establishes runtime conformance, platform or field support.
