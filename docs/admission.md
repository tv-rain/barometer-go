# Admission Design

## Status and Authority

DEV-583 records the intended design of `internal/admission` and the minimal `internal/registry` it consumes. DEV-597 rebaselines it to Draft `0.1.0` at public revision `d93aff93cbc17e7c065fb40d8f6e2e01a697bf35` (public PR #14 merge), which adds REQ-TRUST-0036 to REQ-TRUST-0038 for [GO-SPEC-ISSUE-008](specification-issues.md#go-spec-issue-008-admission-error-classes-for-unsupported-and-unmapped-rejections). The controlling sources are [Task Trust and Local Policy](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md), [Safety](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md) and [Registered Initial Limits](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md), with coordination decision 0002 (trusted task boundary and registered limits). The public text controls behavior. This record fixes step order, error mapping and component ownership so the admission tasks of the Admission and local policy sub-epic (DEV-584 to DEV-594) do not have to guess them.

This record was written before the packages existed and changes no public specification; it now also records what the sub-epic delivered. `internal/registry` now exists with its admission subset (DEV-584, see [Registry](package-architecture.md#registry)). `internal/admission` exists with the delivery-record seam (DEV-585, see [Delivery-Record Seam](#delivery-record-seam)) the gates for steps 1 to 6 and the input and options part of step 8 (DEV-586, see [Implemented Gates](#implemented-gates)), and issuer and policy resolution in step 7 with the path-authority part of step 8 and the issuer and local risk bounds of step 9 (DEV-587, see [Issuer Authorization](#issuer-authorization)), target registration matching in step 8 (DEV-588, see [Target Registration](#target-registration)), and effective risk with the consent requirement in step 9 (DEV-589, see [Effective Risk and Consent](#effective-risk-and-consent)), and the rejection of tasks that require egress Network Context or NET-05 composite steps (DEV-590, see [Unsupported Capabilities](#unsupported-capabilities)), and replay and acceptance in steps 10 and 11 with the `persistence.ReplayStore` contract (DEV-591, see [Atomic Rejection and Replay Persistence](#atomic-rejection-and-replay-persistence)). The root `Client.AdmitTask` runs the complete pipeline through the public facade (DEV-592), local policy is replaced only by quiesce and reinitialization (DEV-593), and DEV-594 adds hostile-input fuzzing, strict JSON cases and a dependency-graph check at the admission boundary ([Hostile-Input Evidence](#hostile-input-evidence)). DEV-599 adds the proxy-compatibility part of step 8 ([Proxy Compatibility](#proxy-compatibility)). Each package is created by the first task that gives it behavior and tests, following the [package-creation rules](package-architecture.md#package-creation-rules). Remote-delivery authentication and issuer/target revocation were deferred behind the remote-delivery gate (DEV-595); DEV-612 adds signed-package verification and key rotation by policy replacement, and the current state of that gate is recorded under [Signed Remote Task Packages](#signed-remote-task-packages). The design needs human review before the dependent tasks move to `to do`.

DEV-612 rebaselines this record to public revision `e97502747c917a16bc2b71337535ff972bbd7d64` (public PR #18 merge), which adopts coordination decision 0011: Barometer verifies signed remote task packages itself against trusted issuer keys in local policy, and the host no longer asserts remote authentication ([Signed Remote Task Packages](#signed-remote-task-packages)). That revision retires REQ-TRUST-0001, REQ-TRUST-0002, REQ-TRUST-0004 to REQ-TRUST-0007, REQ-TRUST-0009, REQ-TRUST-0017 and REQ-TRUST-0037 in favor of REQ-TRUST-0040 to REQ-TRUST-0047, and adds REQ-TRUST-0048 to REQ-TRUST-0054 and REQ-LIMIT-0032. Links below cite that revision; links to an unchanged requirement at `d93aff9` remain valid because its text did not change. DEV-612 did not lift the remote-delivery gate (DEV-595): it left the cryptographic review of the package and verification design, the safety review of residual key exposure and the maximum key validity, and the host integration review of package retention to be performed. Those reviews were performed later and the gate is lifted ([Remote-Delivery Gate](#remote-delivery-gate)).

The public text now names the class of every admission rejection this record previously marked *open*, and makes the first failing step normative ([REQ-TRUST-0047](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0047), successor of REQ-TRUST-0037). The GO-SPEC-ISSUE-008 disposition is adopted in this record and in `task.AdmissionError`; the admission tasks still have to implement and evidence each row.

DEV-599 adopts the [GO-SPEC-ISSUE-009](specification-issues.md#go-spec-issue-009-admission-class-for-proxy-incompatible-task-parameters) disposition at the same public revision `e975027`, which carries [REQ-TRUST-0039](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0039) and [REQ-PROXY-0042](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/proxy.md#REQ-PROXY-0042) (public PR #16, merged before PR #18). Step 8 now evaluates registry proxy compatibility ([Proxy Compatibility](#proxy-compatibility)).

DEV-624 rebaselines this record to public revision `4743b636f29feb1f0f8d042866d04ccda3532938` (public PR #19 merge), which adds [REQ-TRUST-0055](https://github.com/tv-rain/barometer-spec/blob/4743b636f29feb1f0f8d042866d04ccda3532938/library/task-trust.md#REQ-TRUST-0055) from the [key-expiry redelivery record](https://github.com/tv-rain/barometer-spec/blob/4743b636f29feb1f0f8d042866d04ccda3532938/specification-issues/key-expiry-redelivery.md) and changes no other requirement. REQ-TRUST-0055 confirms the existing fail-closed verification of every presented package ([Signed Remote Task Packages](#signed-remote-task-packages)). DEV-624 also makes step 1 decide the delivery binding before it reads trusted time, as the REQ-TRUST-0047 table orders ([Delivery-Record Seam](#delivery-record-seam)), and records the remote-delivery gate decision ([Remote-Delivery Gate](#remote-delivery-gate)). Links to an unchanged requirement at `e975027` or `d93aff9` remain valid.

## Existing Inputs

Admission composes values that already exist and adds no new trust model:

- `task.SourceContext` and `task.DeliveryRecord` ([Trusted Context and Local Policy Models](trusted-policy-models.md)): origin, policy generation, the issuer of an embedded task or of a verified remote package, the issuance time of a verified remote package, and the SHA-256 digest of the exact received bytes.
- `task.LocalPolicy`: the immutable issuer permissions, deployment permissions, target registrations and trusted issuer keys (`task.TrustedKey`) of one policy generation.
- `taskdecode.Decoder` and `validation.Common` ([Task Models](task-models.md)): bounded strict decoding, common validation and exact `(test_id, test_version)` codec selection.
- `task.AdmissionError` ([Execution-State and Error Models](execution-models.md#admission-errors)): the ten REQ-TRUST-0034 classes and `task_unsupported` (REQ-TRUST-0036), carrying only a code.
- `platform.Clock`, reached through `config.Initialization` ([Initialization and Configuration Models](configuration-models.md)).

## Ordered Steps and Error Classes

<!-- requirement-ref: REQ-TRUST-0046 -->
<!-- requirement-ref: REQ-TRUST-0034 -->
<!-- requirement-ref: REQ-TRUST-0036 -->
<!-- requirement-ref: REQ-TRUST-0047 -->
<!-- requirement-ref: REQ-TRUST-0039 -->
<!-- requirement-history: REQ-TRUST-0017 -->
<!-- requirement-history: REQ-TRUST-0037 -->

Admission runs the eleven steps of [REQ-TRUST-0046](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0046) in public order. The first failing step ends admission, and the rejected task receives exactly one class, that of the first failing step ([REQ-TRUST-0047](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0047)); within step 2 a remote package receives the class of the first failing check of [REQ-TRUST-0052](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0052). This table was written for REQ-TRUST-0017 and REQ-TRUST-0037 at `d93aff9`; their successors keep steps 4 to 11 and every class, and move the remote `issued_at` check from step 3 to check 8 of step 2. Classes are those of [REQ-TRUST-0034](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0034) and `task_unsupported` ([REQ-TRUST-0036](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0036)). Where another requirement names the class for a condition, that class applies; the table below follows the REQ-TRUST-0047 condition-to-class table.

| Step | Work | Class on failure |
|---|---|---|
| 1. Receive | Resolve the adapter-installed delivery record for the presented reference in this client instance and compare it with the exact presented bytes (`DeliveryRecord.Matches`): the serialized task for embedded origin, the signed package for remote origin. | Missing, expired, foreign-instance or mismatched record: `task_source_unauthenticated`, decided before trusted time is read. Trusted time unavailable for a held record that matches the presented bytes: `task_trust_unavailable`. |
| 2. Provenance | Embedded: require reviewed embedded origin. Remote: verify the signed package by the nine checks of REQ-TRUST-0052 ([Signed Remote Task Packages](#signed-remote-task-packages)). | Embedded or unknown origin that is not reviewed: `task_source_unauthenticated`. Remote: the class of the first failing check (`task_limit_exceeded`, `task_policy_unavailable`, `task_trust_unavailable` or `task_source_unauthenticated`). |
| 3. Time and size | Embedded: read the trusted clock. Remote: use the acceptance time of check 6, and the `issuer_id` and `issued_at` of the verified package. Reject more than 65,536 task bytes before parsing; parse with the bounded strict decoder. | Clock unavailable: `task_trust_unavailable`. Over the byte limit: `task_limit_exceeded`. Malformed JSON: `task_schema_invalid`. |
| 4. Schema | Validate the closed top-level schema, `schema_version` and exact test version. | `task_schema_invalid`. Well-formed but unregistered `(test_id, test_version)` pair: `task_unsupported`. |
| 5. Identifier | Validate `task_id` characters and length. Campaign scope and non-user-specificity cannot be established from syntax; they stay an issuer review obligation (`RV-CC-TRUST-TASK-IDENTIFIER-RULES`). | Characters: `task_schema_invalid`. Length over 128 bytes: `task_limit_exceeded`. |
| 6. Fields and limits | Validate expiry, schedule, retry, execution, safety and every registered numeric limit (`validation.Common`). | Expired: `task_expired`. Value above a registered or stricter local limit: `task_limit_exceeded`. Other invalid structure: `task_schema_invalid`. Unknown risk factor type rejected under REQ-SAFETY-0017 rather than treated as `unknown` risk, or a valid field that requires a capability this build does not provide (for example egress Network Context for `execution.context`): `task_unsupported`. |
| 7. Resolve | Select the exact registry entry; resolve issuer permissions, target registrations and the active policy generation. | No policy installed: `task_policy_unavailable`. Unknown issuer, or test/version not permitted: `task_issuer_not_authorized`. Source generation lower than the active one: `policy_rollback`. Source generation higher than the active one: `task_policy_unavailable`. Value beyond an issuer or local ceiling: `task_limit_exceeded`. Label outside the permitted namespaces: `task_issuer_not_authorized`. |
| 8. Atomic validation | Run test-specific input/options validation, then registry proxy compatibility for a `proxy` task, then path authority, and match every target tuple, redirect registration, path and composite step against the resolved registrations. | Invalid input or options: `task_schema_invalid` (no public row names this class; [GO-SPEC-ISSUE-010](specification-issues.md#go-spec-issue-010-admission-class-for-step-8-input-and-options-failures)). Test, input or options incompatible with HTTP CONNECT on the proxy path: `task_schema_invalid` (REQ-TRUST-0039), before path authority and target registration. Unmatched or expired registration, or a direct or proxy path the matched target registration does not permit: `target_not_registered`. Direct or proxy path not permitted by issuer authority or local policy: `task_issuer_not_authorized`. Composite steps this build cannot match: `task_unsupported` (REQ-TRUST-0036; the REQ-TRUST-0037 table lists no step-8 row for it). |
| 9. Risk and consent | Compute effective risk and the consent requirement (see [Outputs Consumed by Execution](#outputs-consumed-by-execution)). | Effective risk above the issuer maximum or a local-policy bound, or `unknown` without an explicit reviewed local policy permitting it: `task_issuer_not_authorized`. |
| 10. Replay | Read the highest retained accepted policy generation, then look up `(issuer_id, task_id)` in the replay store. | Active generation lower than the highest retained accepted generation: `policy_rollback`. Different digest: `task_id_conflict`. Store unavailable or corrupt, for remote and embedded origin: `task_trust_unavailable`. Same digest: idempotent result, no error. |
| 11. Accept | Commit the replay record atomically and conditionally, then return the admitted task. | Persistence failure: `task_trust_unavailable`. |

`task_unsupported` describes this build, not the task's validity or safety or the network. A build that provides a capability does not return the class for it; runtime unavailability of a provided capability, such as egress Network Context, remains an unsatisfied execution constraint recorded per attempt ([REQ-CONTEXT-0026](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0026), [REQ-CONTEXT-0028](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0028)).

<!-- requirement-ref: REQ-TRUST-0018 -->

Steps 1 to 9 are pure with respect to Barometer state and the network. They read only the delivery record, clock, registry, policy snapshot and the task bytes. No step performs DNS, Network Context, proxy or measurement activity ([REQ-TRUST-0018](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0018)), so a test can check this with a recording transport that must observe zero calls.

<!-- requirement-ref: REQ-TRUST-0035 -->

Every class is returned as a `task.AdmissionError`, which carries only the code. Decoder, validation, store and platform failures are mapped to a code at the step boundary and never wrapped, so no payload, target, credential, key, signature, participant value or raw platform error reaches the caller ([REQ-TRUST-0035](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0035)). An admission error is never a Measurement Result.

## Consumed Capabilities and Owners

| Capability | Contract used by admission | Owner |
|---|---|---|
| Delivery-record lookup | Resolve a transient reference to one `DeliveryRecord` in this client instance; bounded lifetime and capacity; invalidated on clearing, close and policy replacement. | `internal/admission` (seam and embedded origin, DEV-585; package-only remote records and signed-package verification, DEV-612). Release gate for remote delivery: DEV-595, lifted October 21, 2026 ([Remote-Delivery Gate](#remote-delivery-gate)). |
| Trusted clock | `platform.Clock`; unavailable time fails closed. | `internal/platform`, supplied through `internal/config` (existing). |
| Replay store | Read by replay key; highest retained accepted generation; one atomic conditional write of the replay record; delete on clearing. | `persistence.ReplayStore` contract in `internal/persistence`, defined from admission's needs (DEV-591) and agreed with the State, retention and clearing sub-epic (DEV-577), which owns the production implementation; clearing is DEV-564 and platform storage DEV-455. |
| Registry | Exact `(test_id, test_version)` lookup returning codec, baseline risk, target rule and proxy compatibility. | `internal/registry`, DEV-584. |
| Local policy snapshot | The active immutable `task.LocalPolicy` and its generation. | `internal/model/task` (existing). Activation and replacement by reinitialization: the root `Client` (DEV-593, see [Implemented Subset](public-api.md#implemented-subset)). |

The client lifecycle that serializes admission with other transitions and owns handle tables belongs to the Runtime and execution sub-epic (DEV-576). DEV-600 provides it as a lifecycle shell: the root `Client` binds these capabilities through an unexported adapter seam, builds the `Admitter` at initialization and owns the task and run handle tables (see [Implemented Subset](public-api.md#implemented-subset)). `Client.AdmitTask` (DEV-592) reserves the current generation, runs `Admit` under the generation's context and publishes a task handle only after `Admit` accepts; an exact redelivery resolves to the handle already published in that generation (see [Implemented Subset](public-api.md#implemented-subset)).

## Delivery-Record Seam

<!-- requirement-ref: REQ-TRUST-0041 -->
<!-- requirement-ref: REQ-TRUST-0042 -->
<!-- requirement-ref: REQ-TRUST-0044 -->
<!-- requirement-ref: REQ-TRUST-0008 -->
<!-- requirement-history: REQ-TRUST-0004 -->
<!-- requirement-history: REQ-TRUST-0005 -->
<!-- requirement-history: REQ-TRUST-0006 -->
<!-- requirement-history: REQ-TRUST-0007 -->

DEV-585 implements the delivery-record lookup consumed by step 1 in `internal/admission/delivery.go`, and DEV-612 reduces its remote path to a transport. `admission.DeliveryStore` belongs to one client instance. It is created with a `platform.Clock`, a capacity of 1 to 256 records, a record lifetime of more than zero and at most one hour, and the active policy generation. These bounds are implementation-local, not public limits; the facade selects concrete values within them.

- **Embedded installation.** `InstallEmbedded` builds an embedded-origin `task.SourceContext` (issuer and policy generation, no issuance time) for an exact task packaged with a reviewed application deployment ([REQ-TRUST-0041](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0041)).
- **Remote installation.** `InstallRemote(ctx, generation, signedPackage)` binds remote origin and the policy generation to the exact received package bytes. The host is a transport ([REQ-TRUST-0042](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0042)): it supplies no issuer, issuance time, authentication flag, key or verdict, and installing verifies nothing. `task.NewSourceContext` refuses a remote issuer, so a host-supplied remote issuer cannot be represented ([REQ-TRUST-0044](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0044)). The remote source stays unverified until admission verifies the package at step 2. The `RemoteVerifier` interface of DEV-585 and its synthetic HMAC verifier are removed.
- **Binding.** Both installers bind the source to the bytes through `task.NewDeliveryRecord`, which keeps the SHA-256 digest and never the payload. Each returns a fresh reference of 128 random bits from `crypto/rand`, lowercase hex. Expired records are purged before the capacity check; a full store rejects the install and never evicts a live record. Installation needs trusted time. Install failures are fixed `ErrDelivery*` sentinels for the adapter, not admission classes. Records bind up to 131,072 bytes, above the 65,536-byte task and 98,304-byte package limits, so oversized input reaches its admission check.
- **Resolution.** `Resolve(ctx, ref, presented)` returns the record only when this store issued the reference under the active policy generation, the record is unexpired and it matches the exact presented bytes. A malformed, forged, foreign-instance, invalidated, expired, stale-generation or byte-mismatched reference returns `task_source_unauthenticated`, as in step 1. The binding is checked before trusted time is read (DEV-624): a reference that is unknown, foreign, stale, invalidated, closed or presented with other bytes is `task_source_unauthenticated` even when the clock is unavailable, because [REQ-TRUST-0047](https://github.com/tv-rain/barometer-spec/blob/4743b636f29feb1f0f8d042866d04ccda3532938/library/task-trust.md#REQ-TRUST-0047) places the binding row before the trusted-time row. Only a held record that matches the presented bytes, whose expiry needs trusted time, returns `task_trust_unavailable` when the clock is unavailable. A clock that reads earlier than installation fails closed as an expired record. The record is checked again after the clock is read, so one invalidated, replaced or closed meanwhile does not resolve. An embedded record is not consumed: exact redelivery is handled by the replay store at step 10. A remote record is single-use: the first resolution of its reference drops it, whatever the outcome, so no digest of a signed package outlives the admission that presented it ([REQ-TRUST-0053](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0053)). A host that retries a package installs a new reference.
- **Stale generation.** A record is stale when it was installed while an earlier policy generation was active. This is separate from the source's own `policy_generation`, which step 7 compares with the active policy and which can still yield `policy_rollback` or `task_policy_unavailable`.
- **Invalidation.** `Invalidate` drops every record for data clearing; `ReplacePolicy` drops every record and activates a strictly higher generation; `Close` drops every record and refuses later installs. The facade and lifecycle tasks call them; the store does not observe the client lifecycle itself.
- **Privacy.** References and records are not task fields and have no JSON form. Formatting the store prints only its type name, and errors carry only fixed text or the admission code, so references, issuers, packages and task bytes cannot reach traffic, results, logs or persistent identifiers through this seam ([REQ-TRUST-0008](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0008)). The root facade passes `sourceRef` to `Resolve` unchanged and never formats it (DEV-592); native resolution remains DEV-455.

Synthetic tests in `internal/admission/delivery_test.go` cover forged, missing, malformed, expired, foreign-instance, stale-generation, one-byte-mismatch and closed cases, capacity and lifetime boundaries with the manual test clock, unavailable time, binding before trusted time with the clock unavailable (`TestResolveChecksBindingBeforeTrustedTime`), rejected installs, single-use remote records (`TestRemoteRecordsAreSingleUse`), leakage canaries and concurrent install, resolution and invalidation under the race detector.

## Signed Remote Task Packages

<!-- requirement-ref: REQ-TRUST-0040 -->
<!-- requirement-ref: REQ-TRUST-0043 -->
<!-- requirement-ref: REQ-TRUST-0045 -->
<!-- requirement-ref: REQ-TRUST-0048 -->
<!-- requirement-ref: REQ-TRUST-0049 -->
<!-- requirement-ref: REQ-TRUST-0050 -->
<!-- requirement-ref: REQ-TRUST-0051 -->
<!-- requirement-ref: REQ-TRUST-0052 -->
<!-- requirement-ref: REQ-TRUST-0053 -->
<!-- requirement-ref: REQ-TRUST-0054 -->
<!-- requirement-ref: REQ-TRUST-0055 -->
<!-- requirement-ref: REQ-LIMIT-0032 -->
<!-- requirement-history: REQ-TRUST-0001 -->
<!-- requirement-history: REQ-TRUST-0002 -->
<!-- requirement-history: REQ-TRUST-0009 -->

DEV-612 implements Barometer's verification of signed remote task packages ([REQ-TRUST-0048](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0048) to [REQ-TRUST-0054](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0054)) in `internal/admission/signed.go`, with the trusted key records in `internal/model/task/keys.go`. A remote task is admitted only when its package is signed by a trusted key of its issuer in the active local policy ([REQ-TRUST-0040](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0040), [REQ-TRUST-0043](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0043)). Verification lives in admission, the owner of the ordered decision, rather than in a separate package; it uses only the Go standard library (`crypto/ed25519`, `encoding/base64`, `math/big`) and adds no dependency.

`Admitter.verifyPackage` runs the checks of [REQ-TRUST-0052](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0052) in order on the exact presented package, after step 1 has resolved its record:

| Check | Implementation | Class on failure |
|---|---|---|
| 1 | The package is at most `admission.MaxPackageBytes` (98,304, [REQ-LIMIT-0032](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/registered-limits.md#REQ-LIMIT-0032)) bytes; nothing is decoded otherwise. | `task_limit_exceeded` |
| 2 | The active policy snapshot is valid. | `task_policy_unavailable` |
| 3 | Strict decoding of a closed five-member object (duplicate and unknown members rejected); identifiers of 1 to 64 lowercase ASCII letters, digits, `.`, `_` or `-`; `issued_at` exactly `YYYY-MM-DDTHH:MM:SSZ`, reproduced by formatting the parsed time, so fractions, offsets, lowercase letters, leap seconds and invalid dates fail; `task` and `signature` in canonical unpadded base64url only, checked against the alphabet first because Go's decoder skips newlines, then re-encoded to reject nonzero trailing bits; at least one task byte and exactly 64 signature bytes. | `task_source_unauthenticated` |
| 4 | `key_id` names a key of the active policy, by exact match. | `task_source_unauthenticated` |
| 5 | That key's issuer equals the package `issuer_id`. | `task_source_unauthenticated` |
| 6 | The trusted clock reads; this reading is the acceptance time. | `task_trust_unavailable` |
| 7 | The acceptance time and `issued_at` each lie within the key's `not_before` to `not_after`, both ends inclusive. | `task_source_unauthenticated` |
| 8 | `issued_at` is at most 300 seconds after the acceptance time. | `task_source_unauthenticated` |
| 9 | `ed25519.Verify` over `M` of [REQ-TRUST-0049](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0049): the context `barometer-signed-task-v1`, `key_id`, `issuer_id`, `issued_at` and the task bytes, each with a 4-byte big-endian length prefix. | `task_source_unauthenticated` |

- **Verification strictness ([REQ-TRUST-0050](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0050)).** Go's `crypto/ed25519.Verify` is cofactorless, rejects a scalar `S` not below the group order, and compares the encoding of the recomputed `R` with the signature bytes, so a non-canonical `R` fails. These are properties of the Go 1.26 standard library, which `TestSignatureStrictnessEdgeVectors` checks rather than assumes: it builds signatures from synthetic keys whose equation holds for the decoded values, and shows that the canonical-`R` control is accepted while the non-canonical `R` encodings and `S + L`, `S + 2L` and `S = L` are rejected.
- **Trusted keys ([REQ-TRUST-0051](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0051)).** Public-key validation happens once, when `task.NewLocalPolicy` validates the policy; see [Trusted Issuer Keys](trusted-policy-models.md#trusted-issuer-keys). A policy with an invalid key record is not constructed, so it cannot be activated.
- **What the task contributes.** The task bytes are not parsed or evaluated before check 9 succeeds: rejected packages never reach the task codec in the tests. The decoded bytes then go through steps 3 to 11 exactly as an embedded task's bytes do, and the replay identity is the digest of those task bytes, not of the package. A re-signed package of the same task bytes is therefore an exact redelivery.
- **Source context.** On success the unverified remote source is completed with the package's `issuer_id` and `issued_at` (`SourceContext.WithVerifiedPackage`); no host value contributes. No Go, native or Flutter binding accepts a key, a verdict or an authentication flag ([REQ-TRUST-0045](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0045)); trusted keys enter only inside the policy bound by the unexported adapter seam.
- **No retention ([REQ-TRUST-0053](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0053)).** The decoded package exists only inside one verification call. `Gated` keeps the verified source context and the digest of the task bytes, not the delivery record, and the remote delivery record is dropped at resolution. The replay record holds the issuer, `task_id` and the task-byte digest. Errors carry only the class, never which check failed.
- **Rotation and revocation ([REQ-TRUST-0054](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0054)).** Keys and issuers change only through `replacePolicy` with a strictly higher generation (DEV-593). Replacement invalidates every task handle and delivery reference, so a task signed with a removed key, or by a removed issuer, fails check 4 when its package is redelivered, and a task signed with a retained key resolves to its retained acceptance. A task admitted while its key was valid is not re-verified when the key later expires: its handle stays until its own expiry.
- **Every presentation is verified ([REQ-TRUST-0055](https://github.com/tv-rain/barometer-spec/blob/4743b636f29feb1f0f8d042866d04ccda3532938/library/task-trust.md#REQ-TRUST-0055)).** The REQ-TRUST-0054 exemption covers only an acceptance the instance still holds. Every presented package, including an exact redelivery and a redelivery after policy replacement, reinitialization or clearing, goes through the nine checks at its own acceptance time before replay handling at step 10, so a package whose key is no longer trusted and valid is `task_source_unauthenticated` even when a replay record or a held handle exists, and the rejection writes no replay record and leaves a held handle valid. The replay identity is the digest of the task bytes, so the same task bytes re-signed with a currently valid key are an exact redelivery that resolves to the retained acceptance with its original expiry. Nothing retained (package, signature or key ID) exempts a later presentation. This is the behavior DEV-612 implemented; DEV-624 adds the requirement reference and two cases to `TestCcTrustKeyRotationByPolicyReplacement`.

The interpretation question DEV-612 recorded here, whether REQ-TRUST-0054 lets a task survive policy replacement without re-verification, is closed by [REQ-TRUST-0055](https://github.com/tv-rain/barometer-spec/blob/4743b636f29feb1f0f8d042866d04ccda3532938/library/task-trust.md#REQ-TRUST-0055), which confirms the fail-closed behavior described above ([specification issues](specification-issues.md#interpretation-question)). A remote task whose key validity ends before it is redelivered after policy replacement, reinitialization or clearing therefore stops before its own expiry unless its issuer re-signs it with a currently valid key.

### Remote-Delivery Gate

DEV-612 left the remote-delivery gate (DEV-595) closed until the reviews required by coordination decision 0011 were performed: the cryptographic review of the package encoding, message construction, verification strictness and public-key validation, the safety review of residual key exposure and of the 365-day maximum key validity, and the host integration review of package retention. None of them is performed by this implementation or its tests. The project maintainer performed these reviews under DEV-615 (`barometer-spec-internal` PR #7, merge `4522637`) and recorded no findings, and on October 21, 2026 lifted the DEV-595 gate. Independent external reviews (DEV-618 to DEV-621) run asynchronously and do not gate; their findings return as ordinary findings. No binding or host adoption of the package-only remote path exists yet (DEV-455), so hosts cannot yet pass packages to the Go core. Lifting the gate establishes no runtime, platform, field or release support.

## Implemented Gates

<!-- requirement-ref: REQ-TRUST-0019 -->
<!-- requirement-ref: REQ-TRUST-0027 -->
<!-- requirement-ref: REQ-LIMIT-0004 -->
<!-- requirement-ref: REQ-LIMIT-0005 -->

DEV-586 implements the first admission steps in `internal/admission/admit.go`. `admission.Admitter` belongs to one client instance and binds its `DeliveryStore`, trusted `platform.Clock` and reviewed `registry.Registry`; its task decoder is built from exactly the registry's codecs. `Gate(ctx, sourceRef, presented)` runs the steps below in public order and returns either an unpublished `Gated` value or the code-only `task.AdmissionError` of the first failing step. `Gated` holds the verified source context, the digest of the exact task bytes, the trusted acceptance time, the registry entry and the validated candidate for the later steps (DEV-612 replaced the resolved delivery record, so nothing derived from a signed package is kept). It is not an admitted task, holds no task bytes, formats as its type name and has no serialized form. Nothing is persisted or published, so discarding it, or any rejection, leaves no state behind and leaves an embedded delivery record resolvable; a remote record is single-use.

- **Step 1** resolves the delivery record for the exact bytes (see [Delivery-Record Seam](#delivery-record-seam)).
- **Step 2** accepts reviewed embedded origin, or verifies a remote package by the nine checks of [Signed Remote Task Packages](#signed-remote-task-packages) and continues with the decoded task bytes. Any other source is `task_source_unauthenticated`.
- **Step 3** reads trusted time once for embedded origin; for remote origin the acceptance time is the reading of check 6, and `issued_at`, a whole second, comes from the verified package. More than 65,536 task bytes is `task_limit_exceeded`, before parsing. Delivery records bind up to 131,072 bytes (an implementation-local receive bound) so that an oversized task or package reaches its check rather than failing at step 1.
- **Step 4** parses and validates the closed top-level shape and value kinds, with every field value range, count and byte cap relaxed and `input` and `options` treated as opaque objects. It then requires `schema_version` `0.1.0` and selects the exact `(test_id, test_version)`. A well-formed pair the registry does not hold is `task_unsupported`; a malformed pair or other schema version is `task_schema_invalid`. Shape is validated before selection, so a task that is not valid under the public task schema is `task_schema_invalid` even when its pair is unregistered. An integer above the relaxed bound, necessarily above its registered maximum, is `task_limit_exceeded`.
- **Step 5** checks `task_id` on the exact string: more than 128 bytes is `task_limit_exceeded` whatever the characters; empty, any character outside lowercase ASCII letters, digits, `.`, `_` and `-`, or a leading `.`, `_` or `-` (the public task format pattern) is `task_schema_invalid`. Campaign scope and non-user-specificity remain the issuer review obligation `RV-CC-TRUST-TASK-IDENTIFIER-RULES`.
- **Step 6** applies every common field value, registered limit and default, then expiry against the acceptance time. A value above a registered maximum, including `input` over 49,152 and `options` over 8,192 encoded bytes, labels over 16 entries or 2,048 encoded bytes and `not_after` beyond the seven-day lifetime, is `task_limit_exceeded`. `not_after` at or before acceptance is `task_expired`. An unrecognized risk factor type, rejected under REQ-SAFETY-0017 rather than treated as `unknown` risk, is `task_unsupported`. Every other invalid value, including a value below a registered minimum, is `task_schema_invalid`. Embedded tasks without `not_after` are rejected, because no reviewed configuration for the one-time embedded exception exists yet. Last, a valid task that requires egress Network Context is `task_unsupported` (DEV-590, [Unsupported Capabilities](#unsupported-capabilities)).
- **Step 8, input and options** validates them against the selected test's schemas and normalizes them with its codec; failures are `task_schema_invalid`, or `task_limit_exceeded` for a stricter test byte cap. DEV-587 adds step 7 before it and the path-authority part after it ([Issuer Authorization](#issuer-authorization)); DEV-588 adds the target part ([Target Registration](#target-registration)); DEV-590 rejects composite tasks right after input and options ([Unsupported Capabilities](#unsupported-capabilities)); DEV-599 adds proxy compatibility after them ([Proxy Compatibility](#proxy-compatibility)). No public row of REQ-TRUST-0047 names the class of an input or options failure; Go uses `task_schema_invalid` by its name, and `task_limit_exceeded` for a stricter test byte cap, pending public disposition ([GO-SPEC-ISSUE-010](specification-issues.md#go-spec-issue-010-admission-class-for-step-8-input-and-options-failures)).

Within one step, the order of checks above decides which class is returned when a task fails more than one check; REQ-TRUST-0037 orders only steps. Decoding and validation failures are mapped at the step boundary through refining classes (`jsondecode.ErrLimit`, `validation.ErrLimit`, `validation.ErrExpired`, `validation.ErrUnsupported` and `taskdecode.ErrUnregistered`) that still match their coarse classes, so existing decoder callers are unaffected. The staged decoder is described in [Task Models](task-models.md).

Task-controlled authority claims and recipient selectors are unknown fields and fail at step 4. Placeholder text in identifiers, labels, input and options is kept literal: the library performs no local interpolation of deployment, application-version, cohort, country or network values. Broad eligibility applied by host policy, and its review, remain host obligations the library cannot enforce (`RV-CC-TRUST-TARGETED-DELIVERY`). `Gate` takes the presented bytes as a slice; reading them through a counting reader that stops at 65,537 bytes belongs to the native adapter that receives them (DEV-455). The root facade receives an immutable Go string and adds no check of its own, so the byte limit keeps its step-3 position after delivery resolution (DEV-592).

Synthetic tests in `internal/admission/admit_test.go` cover each step's classes and boundaries, first-failing-step precedence, origins, `task_id` rules, task-controlled authority and selector fields, literal placeholders, canary leakage, ownership, concurrency and the strict JSON rules at the gate (escaped and nested duplicate names, exact integer spellings, exact versions after decoding escapes, framing and nesting depth). Their recording transport is never connected to admission, which has no transport seam; network absence is checked on the dependency graph instead ([Hostile-Input Evidence](#hostile-input-evidence)). `FuzzGate` checks that arbitrary bound bytes yield either a gate outcome or exactly one class, deterministically. The status of each admission case family is recorded in [Hostile-Input Evidence](#hostile-input-evidence).

## Issuer Authorization

<!-- requirement-ref: REQ-TRUST-0010 -->
<!-- requirement-ref: REQ-TRUST-0011 -->
<!-- requirement-ref: REQ-TRUST-0012 -->
<!-- requirement-ref: REQ-LIMIT-0026 -->
<!-- requirement-ref: REQ-LIMIT-0030 -->

DEV-587 implements effective authority ([REQ-TRUST-0010](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0010)) in `internal/admission/authority.go`. `NewAdmitter` also binds the instance's active `task.LocalPolicy` snapshot; an invalid (zero) snapshot models an instance with no installed policy. The root `Client` activates the snapshot at initialization and replaces it only by reinitialization (DEV-593). Issuer permissions and deployment permissions are applied independently and intersected: a permission is effective only when both grant it, a maximum is the lower of the two and a minimum the higher. Nothing is a union or an implicit default, and no value is clamped. The intersection is carried on `Gated` as an unpublished `Authority` value (formatted as its type name, with no serialized form) for the later steps and for execution.

- **Step 7** first requires a valid policy snapshot (`task_policy_unavailable`), then compares the source context's `policy_generation` with the active one: lower is `policy_rollback`, higher is `task_policy_unavailable`. It then resolves the source issuer by exact ID and requires the exact `(test_id, test_version)` in both the issuer's and the deployment's test list; otherwise `task_issuer_not_authorized`. NET-05 additionally needs composite enablement on both sides; the policy model already refuses a NET-05 entry without it, so this check is defense in depth. An enabled NET-05 task is still rejected at step 8 until composite steps can be matched ([Unsupported Capabilities](#unsupported-capabilities)).
- **Step 7, numeric ceilings.** Each task value bounded by a policy ceiling is compared with the intersected ceiling: attempt duration, bytes sent, bytes received and connections; `max_runs`; retry `max_attempts`; the validity lifetime from the earlier of issuance and acceptance, as in expiry validation; repeat interval and jitter minima; and retry initial delay minimum, maximum delay and jitter maximum. A value beyond a ceiling, including an interval, jitter or retry delay below a policy minimum, raises authority ([REQ-TRUST-0012](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0012), [REQ-LIMIT-0030](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0030)) and is `task_limit_exceeded`. A value within a ceiling becomes the effective value. Result, bundle, pending-storage and retention ceilings have no task field; they are carried intersected for result and storage enforcement. The registered maxima were already applied at step 6, and the policy model refuses any host ceiling above them.
- **Step 7, labels.** Every label key must lie in a namespace permitted by the issuer and, separately, by the deployment: equal to a listed namespace or below it at a `.` boundary, as with the reserved `barometer.` namespace ([REQ-LABEL-0011](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0011)). `campaign` permits `campaign` and `campaign.wave`, not `campaignx`. An empty list permits no labels. A key outside is `task_issuer_not_authorized`. The public text does not define namespace matching further; this reading is a Go-local choice that admits no key the literal reviewed names do not cover.
- **Step 8, path authority** runs after input and options and proxy compatibility ([Proxy Compatibility](#proxy-compatibility)): the task's `network_path` must be permitted by both sides, otherwise `task_issuer_not_authorized` ([REQ-TRUST-0037](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0037)). The effective paths narrow to the task's one path. The matched registration's paths are checked by target matching (`target_not_registered`, [Target Registration](#target-registration)).
- **Step 9, issuer and local risk bounds.** `unknown` needs the reviewed unknown-risk exception on both sides. An ordered risk above the lower maximum, or a medium or high risk without medium-risk enablement on both sides, is `task_issuer_not_authorized`. Treating high risk as requiring medium-risk enablement is the fail-closed reading of the opt-in. DEV-589 applies these bounds to the full effective risk ([Effective Risk and Consent](#effective-risk-and-consent)).
- **Carried authority.** The intersected registration IDs, narrowed by target matching to those the task matched, the narrowed ceilings and path, the risk bounds and the consent requirement, which is required when the issuer, the deployment or the task requires it.

Platform capability, the fourth term of the intersection, is applied where a capability is missing (DEV-590, `task_unsupported`, [Unsupported Capabilities](#unsupported-capabilities)).

Synthetic tests in `internal/admission/authority_test.go` narrow every dimension in the issuer and then in the deployment while the other stays permissive, cover limit −1, limit and limit +1 for every task-bounded ceiling with the lower value carried as effective, issuer-versus-deployment conflicts in both directions, generation lower, equal and higher, unknown issuers, remote sources, precedence against steps 4 to 8, and the admission-time rows of `CC-LIMIT-LIMIT-HANDLING`, including attempts to raise a limit through labels, input, options, schedule, execution, safety, an unknown field and host configuration. All use the unused recording transport that must see zero calls. The NET-05 nine-step row of `CC-LIMIT-LIMIT-HANDLING` needs the NET-05 codec (DEV-454) and is not covered. `CC-TRUST-ISSUER-AUTHORIZATION` stays `planned`: its registration-reference row is now covered by target matching ([Target Registration](#target-registration)), and steps 10 and 11 exist (DEV-591), but its rows use NET-01 and NET-05 tasks and a platform-capability dimension, which need production codecs (DEV-454) and NET-05 composite matching ([Admission Case Families](#admission-case-families)). `CC-LIMIT-LIMIT-HANDLING` stays `planned` for its runtime rows (DEV-576).

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

Registrations with credentials, interpolation syntax, wildcard or noncanonical hosts, or IPv4-mapped addresses are already refused when the policy snapshot is built ([Trusted Context and Local Policy Models](trusted-policy-models.md#target-representation-and-limits)). NET-05 composite-step matching remains an unimplemented part of step 8. Composite-step matching needs the NET-05 codec (DEV-454); until then composite tasks are rejected ([Unsupported Capabilities](#unsupported-capabilities)).

Synthetic tests in `internal/admission/target_test.go` use test-scoped codecs and rules under the TCP-01, TLS-01, HTTP-01 and DNS-01 IDs. These are not the public test schemas or production codecs. The tests cover:

- an exact match for each test, with the carried IDs;
- single-component mutations of every tuple component, a redirect target and direct versus proxy path;
- normalized forms that match, against suffix, subdomain, parent-domain, trailing-dot, wildcard, URL path-prefix, URL host-suffix and IPv4-mapped forms that do not;
- expiry at one second after, at and one second before trusted `now`, including an expired redirect registration;
- issuer, deployment, test/version, role, path and redirect-listing failures;
- incomplete tuples and the target bound;
- precedence against steps 6, 4 and the path authority of step 8, and before step 9;
- policy refusal of credential and interpolation registrations.

All use the unused recording transport, which must see zero calls. `CC-TRUST-TARGET-REGISTRATION` stays `planned`. Steps 10 and 11 exist (DEV-591), but its rows use TCP-01, TLS-01, HTTP-01 and DNS-01 tuples, which need production codecs (DEV-454), and its execution-owned resolver and redirect rows need the execution work above ([Admission Case Families](#admission-case-families)).

## Proxy Compatibility

<!-- requirement-ref: REQ-TRUST-0039 -->
<!-- requirement-ref: REQ-PROXY-0026 -->
<!-- requirement-ref: REQ-PROXY-0027 -->
<!-- requirement-ref: REQ-PROXY-0042 -->

DEV-599 implements the proxy-compatibility part of step 8 in `internal/admission/admit.go`, adopting the [GO-SPEC-ISSUE-009](specification-issues.md#go-spec-issue-009-admission-class-for-proxy-incompatible-task-parameters) disposition.

- **Rule from the registered test.** Each registry entry carries a required `registry.ProxyRule`. For a task with `network_path: proxy`, admission passes the codec's normalized parameters to it. When the rule reports the test, input or options incompatible with HTTP CONNECT ([REQ-PROXY-0026](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/proxy.md#REQ-PROXY-0026)), the task is rejected atomically with `task_schema_invalid` ([REQ-TRUST-0039](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0039)) before any network activity ([REQ-PROXY-0027](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/proxy.md#REQ-PROXY-0027)). A `direct` task is not checked. `registry.NoProxy` rejects every proxy task of its test. Compatibility never grants proxy use; the path must still be permitted by issuer, local policy and the matched registration.
- **Order.** The check runs after input and options and before path authority and target registration. A task that is both incompatible and not permitted the proxy path is `task_schema_invalid`. Steps 1 to 7 and input and options keep precedence over it.
- **Not admitted to fail at execution.** An incompatible task never becomes an admitted task that ends in `proxy_path_unsupported` ([REQ-PROXY-0042](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/proxy.md#REQ-PROXY-0042)). That execution state remains for an incompatibility the pre-attempt check of REQ-EXECUTION-0008 finds in an admitted task; that check belongs to execution (DEV-578) and is not implemented.
- **NET-05.** The composite-capability rejection of [Unsupported Capabilities](#unsupported-capabilities) runs before this check, so an enabled NET-05 task is still `task_unsupported` and its registry rule is not consulted. Rejecting a NET-05 profile with an incompatible step as `task_schema_invalid` needs the NET-05 codec (DEV-454).
- **Rules are synthetic.** No production rule for a registered test exists; the per-test compatibility of REQ-PROXY-0026 belongs to the production codecs (DEV-454). `task_unsupported` for a combination REQ-PROXY-0026 permits but this build does not provide through the proxy has no case until a production rule distinguishes it.

`TestGateProxyCompatibility` and `TestGateCompositeBeforeProxyCompatibility` in `internal/admission/proxy_test.go` use synthetic test-scoped rules over the SYN-01 and NET-05 fixtures. They cover a compatible and an incompatible proxy task for embedded and remote origin, the unchecked direct path, a never-compatible rule, precedence over issuer and local path authority, target registration and the step-9 risk bound, the earlier input, options, codec, step-6, step-7 and policy-rollback failures, which leave the rule unconsulted, and the composite rejection before the rule. All use the unused recording transport, which must see zero calls. `CC-PROXY-INITIAL-TEST-COMPATIBILITY` stays `planned`: its plan parameterizes the nine registered tests, which need production codecs and rules (DEV-454).

## Effective Risk and Consent

<!-- requirement-ref: REQ-SAFETY-0010 -->
<!-- requirement-ref: REQ-SAFETY-0020 -->
<!-- requirement-ref: REQ-SAFETY-0021 -->
<!-- requirement-ref: REQ-SAFETY-0022 -->
<!-- requirement-ref: REQ-SAFETY-0023 -->
<!-- requirement-ref: REQ-SAFETY-0024 -->
<!-- requirement-ref: REQ-SAFETY-0025 -->
<!-- requirement-ref: REQ-SAFETY-0029 -->

DEV-589 completes step 9 in `internal/admission/authority.go`. It runs after target matching and before replay.

- **Effective risk** is the most restrictive of three trusted inputs ([REQ-SAFETY-0023](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0023)):
  - the registered baseline of the exact registry entry;
  - the task-declared `risk_level`, already validated at step 6;
  - the locally determined risk: the reviewed `BaselineRisk` of every target registration the task matched at step 8.

  `unknown` from any input makes effective risk `unknown`; otherwise the highest ordered level wins. No input lowers it, so a `low` declaration on a `medium` baseline is evaluated as `medium` ([REQ-SAFETY-0024](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0024), [REQ-SAFETY-0025](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0025)). Using matched registrations as the locally determined input is a Go-local choice: they are the only reviewed, trusted, task-specific risk assessment that exists. No pluggable local risk assessor exists.
- **Bounds.** The [issuer and local risk bounds](#issuer-authorization) apply to effective risk. A rejection is `task_issuer_not_authorized`, an admission error and never a measurement result ([REQ-SAFETY-0030](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0030)).
- **Fail closed.** A matched registration that can no longer be read, or a level outside the reviewed set, counts as `unknown` ([REQ-SAFETY-0029](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0029)). An unrecognized risk factor type is already `task_unsupported` at step 6, even where `unknown` risk is permitted. An `unknown` factor under a known overall level is already `task_schema_invalid` at step 6: [REQ-SAFETY-0019](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0019) allows a trusted local assessment to resolve such a factor, but none exists, so none is applied.
- **Consent requirement** is required when the task, the issuer or the deployment requires it. A task value of `false` adds no requirement and clears none ([REQ-SAFETY-0021](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0021), [REQ-SAFETY-0022](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0022)). Admission does not check consent validity.
- **Recorded values.** Effective risk and the consent requirement are carried on the unpublished `Gated` value. Step 11 (DEV-591) records them on the admitted-task value consumed by execution ([Outputs Consumed by Execution](#outputs-consumed-by-execution)).

Synthetic tests in `internal/admission/safety_test.go` cover every combination of baseline (`low`, `medium`, `unknown`), declaration and matched-registration risk, with the expected effective risk computed independently. They also cover:

- baseline-lowering attempts;
- the maximum, medium-enablement and unknown-exception bounds raised by a matched registration, narrowed in the issuer and then in the deployment;
- the unknown-risk exception on neither, one or both sides;
- unrecognized factor types and unknown factors under a known level;
- precedence of a target failure over a risk failure;
- every combination of task, issuer and deployment consent requirement.

`internal/admission/risk_internal_test.go` covers out-of-range levels and an unreadable matched registration. All use the unused recording transport, which must see zero calls. The `CC-SAFETY-*` families stay `planned`: steps 10 and 11 exist (DEV-591), but each also asserts execution-owned rows such as per-attempt re-evaluation and consent validity, which need runs, a host consent provider and host or device restrictions (DEV-576, DEV-455; [Admission Case Families](#admission-case-families)).

## Unsupported Capabilities

<!-- requirement-ref: REQ-TRUST-0036 -->
<!-- requirement-ref: REQ-EXECUTION-0021 -->
<!-- requirement-ref: REQ-EXECUTION-0023 -->
<!-- requirement-ref: REQ-EXECUTION-0024 -->

DEV-590 makes tasks that need a capability this build does not provide fail closed at admission instead of being accepted and failing later. The checks are in `internal/admission/capability.go`. Each returns `task_unsupported` ([REQ-TRUST-0036](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0036)), describes this build rather than the task or the network, and is removed by the task that delivers the capability.

- **Egress Network Context, step 6.** Acquisition is not provided until Network paths and context (DEV-578). A task whose `execution.network.context` has a non-empty `required_fields` or an `allowed_country_codes` list, which implicitly requires `country_code` ([REQ-EXECUTION-0023](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0023)), is rejected. This is the REQ-TRUST-0037 step-6 condition "a valid field requires a capability the implementation does not provide". It is the last step-6 check, so a structurally invalid context, an expired task or a value above a registered limit keeps its own class, and no later step runs. A context with neither member states no requirement and is admitted; as with an absent context, failure to obtain egress context then does not block execution ([REQ-EXECUTION-0024](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0024)). Admission performs no context lookup, endpoint call or caching. When DEV-578 provides acquisition, this check is removed and an unavailable required field becomes an unsatisfied execution constraint (`network_context_unavailable`, [REQ-CONTEXT-0026](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0026)).
- **NET-05 composite steps, step 8.** NET-05 is rejected as `task_unsupported` at step 4 when the build does not register it, which is the case for every production build today, and as `task_issuer_not_authorized` at step 7 unless composite execution is enabled by both issuer and deployment. A registered and enabled composite is still rejected, because step 8 must match every composite step against the resolved registrations and this needs the NET-05 codec (DEV-454). The check runs after input and options, so invalid composite input stays `task_schema_invalid`, and before path authority, target matching and risk. REQ-TRUST-0037 gives no step-8 row for an unsupported capability; the class follows REQ-TRUST-0036 and the step follows the composite-step matching that REQ-TRUST-0017 places in step 8.

Synthetic tests in `internal/admission/capability_test.go` use the unused recording transport that must see zero calls. They cover:

- positive controls without a network object, without `context`, with an empty `context` and with an empty `required_fields`;
- each required field individually, the allowlist alone and both together, for embedded and remote origin;
- invalid contexts;
- precedence against earlier step-6 failures and against steps 7 to 9;
- unregistered, disabled and enabled NET-05, and precedence within step 8.

These are the admission rejection rows of `CC-CONTEXT-EXECUTION-INTEGRATION`; its acquisition rows stay `planned` for DEV-578.

## Atomic Rejection and Replay Persistence

<!-- requirement-ref: REQ-TRUST-0021 -->
<!-- requirement-ref: REQ-TRUST-0024 -->
<!-- requirement-ref: REQ-TRUST-0038 -->

Rejection is atomic because nothing is written before step 11. Steps 1 to 9 build an immutable, unpublished decision value from owned copies. A failure discards it and leaves the replay store, handle table, schedule accounting and policy unchanged. Validation never clamps, drops or substitutes a value to make a task acceptable.

DEV-591 implements steps 10 and 11 in `internal/admission/state.go`. `Admitter.Admit(ctx, sourceRef, exactTask)` runs `Gate` and then, under the admitter's admission serialization, the steps below. `NewAdmitter` requires a `persistence.ReplayStore`, so no admitter can accept a task without replay state. Acceptance makes a task admitted only after its state is durable:

1. Step 10 reads the highest accepted policy generation among retained replay records. If it is above the active generation, the active policy is a rollback and admission returns `policy_rollback` (see [Rollback Detection Across Restart](#rollback-detection-across-restart)).
2. Step 10 then reads the replay key `(issuer_id, task_id)`. A record with the same SHA-256 digest of the exact bytes is an idempotent redelivery: `Admit` returns an `Admitted` value that carries the original record, reports `Redelivery()`, writes nothing and does not extend expiry or retention. A redelivery must not create a run. A different digest, including a whitespace-only change, returns `task_id_conflict`.
3. For a new key, step 11 builds the replay record and commits it with one atomic write that is conditional on the key still being absent, so concurrent identical deliveries resolve to one record. If another delivery committed first, the same-key rules of step 2 apply to the returned record.
4. Only after the commit succeeds is the task returned as `Admitted`. A failed, interrupted or uncertain commit returns `task_trust_unavailable` and no admitted task. The contract requires a commit to be durable as a whole or not at all. After an uncertain commit that did become durable, a redelivery of the same bytes is an idempotent redelivery.
5. An expired task is rejected at step 6 as `task_expired` before replay state is read, so a retained record never reactivates it, before or after its purge.

The replay record ([REQ-TRUST-0022](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0022)) holds exactly:
- the replay key;
- the SHA-256 digest of the exact accepted bytes;
- the accepted policy generation;
- the acceptance class (the origin);
- the expiry classes needed for cleanup: the task's `not_after`, rounded up to a whole second, and the retention end.

Retention ends 86,400 seconds after the task's expiry, capped at 691,200 seconds after acceptance ([REQ-LIMIT-0023](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0023)). Admission rejects `not_after` beyond seven days after issuance or acceptance, so the cap does not end retention before a valid task expires. A record that would do so is refused, and the task fails closed. Every store method receives trusted time and purges records overdue at that time before use. The record holds no input, option, target, label, participant identifier, result or task bytes ([REQ-TRUST-0023](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0023)), and formats as its type name.

Unavailable or invalid replay state fails closed for remote tasks and, under [REQ-TRUST-0038](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0038), for embedded tasks. This covers a store error, an undecodable record, a record that fails `ReplayRecord.Valid` or does not match its key, and a full store. Admission returns `task_trust_unavailable` and never accepts an embedded task without replay state. Store failures carry no detail and are never wrapped.

`ClearMeasurementData` must delete replay records, so that a later redelivery is evaluated as new within task expiry ([REQ-TRUST-0025](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0025)). Clearing is not implemented here (DEV-564).

`TestCcTrustReplayDuplicateAndRollbackHandling` in `internal/admission/cross_cutting_conformance_test.go` uses the fault-injectable [test replay store](testkit.md#scope-and-ownership) and the manual test clock. It exercises:
- every row of the replay table for embedded and remote origin;
- whitespace-only and one-byte changes;
- expiry and retention boundaries;
- a failure at each store operation, including writes before, during and after durability;
- a corrupt or invalid record, and a full store;
- simulated crash and restart;
- concurrent identical deliveries under the race detector;
- a leakage canary over the stored bytes.

Its post-clearing rows are the planned `CC-TRUST-REPLAY-AFTER-CLEARING` case, which waits for clearing (DEV-564). `internal/persistence` tests cover record construction, retention boundaries and the 691,200-second cap.

### Rollback Detection Across Restart

Decision (DEV-591, agreed for the State, retention and clearing sub-epic DEV-577): the highest accepted policy generation is not persisted as a separate state class. It is derived from the retained replay records through `ReplayStore.HighestGeneration`. A client initialized after a restart with a policy generation lower than one it accepted tasks under rejects every task as `policy_rollback` while such a record is retained. This adds no new persistent state, holds nothing beyond the replay records and is cleared with them.

Limitation: after data clearing, or once the last record from the higher generation passes its retention end (at most 691,200 seconds after acceptance), the library cannot detect that rollback. DEV-593 rejects an earlier or equal policy generation at replacement, and an earlier generation at initialization, against the highest generation the client instance has activated ([REQ-TRUST-0030](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0030), [REQ-TRUST-0031](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0031)). That floor is held in memory: it survives clearing but not a new instance or a restart, where rejecting an earlier generation remains with the host deployment. A durable generation floor would close this gap but adds a state class that clearing must handle; DEV-593 kept it out of scope, and it can be revisited with DEV-564 and DEV-577.

### Admitted-Task Record

Decision (DEV-591): step 11 commits only the replay record. An admitted-task record, from which the executable set could be rebuilt after a restart, would place task content, including targets, in local storage. Its shape depends on the handle tables and execution state owned by the Runtime and execution sub-epic (DEV-576). It is deferred to DEV-576 and DEV-577 and needs privacy review. Until then the admitted task lives only in memory. After a restart, a redelivery of the same bytes resolves through the replay record to the same admitted identity (replay key and digest) with its original expiry and retention. Executable state is not otherwise restored.

## Outputs Consumed by Execution

<!-- requirement-ref: REQ-SAFETY-0023 -->

An accepted task yields an immutable admitted-task value and an opaque local handle. Execution consumes:

- **Admitted-task identity:** replay key `(issuer_id, task_id)`, exact-byte digest, accepted policy generation and origin. Source context never enters traffic or results ([REQ-TRUST-0008](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0008)); `task_id` is emitted unchanged.
- **Validated candidate:** normalized test parameters, schedule, execution constraints, labels and expiry.
- **Effective authority:** the resolved registration IDs, permitted paths and limits after intersecting registered behavior, issuer authority, local policy and task restrictions ([REQ-TRUST-0010](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0010)).
- **Effective risk:** the most restrictive of the registered baseline, task-declared and locally determined risk ([REQ-SAFETY-0023](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0023)). `unknown` stays unordered and is never lowered.
- **Consent requirement:** required when the task, issuer permissions or local policy requires it. A task value of `false` adds no permission ([REQ-SAFETY-0021](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0021)).

Admission records requirements, not eligibility. Safety declarations are evaluated again before every attempt ([REQ-SAFETY-0001](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0001)): execution must check consent status, host and device restrictions, and the policy current at attempt start. Pre-contact address and redirect checks ([REQ-TRUST-0016](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0016)) also remain with execution.

### Consent Handoff

<!-- requirement-ref: REQ-SAFETY-0020 -->
<!-- requirement-ref: REQ-SAFETY-0027 -->

Admission decides only whether consent is required. Whether valid consent covers a measurement is decided at attempt start, because consent can be granted, withdrawn or expire after admission ([REQ-SAFETY-0020](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0020), [REQ-SAFETY-0027](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0027)). No consent-status capability exists yet, and the public specification does not define its host-facing shape. The following obligations are handed off; they are requirements on future work, not implemented behavior.

- **Capability (Mobile and application integration, DEV-455).** The host supplies consent status outside task-controlled data, through trusted initialization or a platform capability. For one admitted task at attempt start it reports exactly one of: valid consent covering this measurement; absent; expired or withdrawn; or unavailable. It carries no user, account, device or consent-record identifier, and no task value selects or changes the answer. Consent UI, storage and signals remain with the host.
- **Consumer (Runtime and execution, DEV-576).** Before reserving an attempt in `StartRun` ([Public API](public-api.md)), execution queries the capability only for a task whose recorded consent requirement is set. Only valid consent permits the attempt. Absent, expired, withdrawn, unavailable, an error or a missing capability means no attempt starts and no Measurement Result is produced. The outcome is a bounded local not-run or operational state, distinct from every network failure reason and never interpreted as censorship ([REQ-SAFETY-0030](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0030)). A task without a recorded requirement does not query the capability.
- **Consumer, implemented (DEV-636).** `Client.StartRun` and the engine apply this through `execution.Gates` before reservation and before each attempt, including a retry, using the internal `platform.Consent` capability ([Attempt-Start Gates](execution-models.md#attempt-start-gates)). The query carries no task value. An unsatisfied check returns `run_not_due` and produces no result. Tests use the synthetic `internal/testkit/capability` provider; no native provider exists.
- **Review points.** The capability's Go and binding shape is a public-API change, and it may need a public specification issue before it is fixed. Both need safety and privacy review. The execution-owned rows of `CC-SAFETY-USER-CONSENT` and `CC-SAFETY-SAFETY-EVALUATION` stay `planned` until this work exists.

## Package-Creation Checklist

### `internal/admission`

- **Responsibility:** the ordered accept, idempotent-recognize or reject decision, which combines decoding, validation and trusted state. It cannot stay in `validation`, which must not mutate replay state, or in the root facade, which must not own policy decisions.
- **Requirements and cases:** the active REQ-TRUST requirements of task trust, including REQ-TRUST-0040 to REQ-TRUST-0055 for signed remote packages, REQ-LIMIT-0032, and the admission-time safety rows, through the `CC-TRUST-*` families, the admission rows of the `CC-SAFETY-*` families and `CC-LIMIT-TASK-AND-IDENTIFIER-LIMITS` in the [cross-cutting plan](cross-cutting-requirements.md).
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

## Hostile-Input Evidence

<!-- requirement-ref: REQ-TRUST-0046 -->
<!-- requirement-ref: REQ-JSON-0003 -->
<!-- requirement-ref: REQ-JSON-0010 -->

DEV-594 consolidates the evidence that untrusted, invalid, replayed or unsupported tasks fail closed before any measurement traffic. It adds tests only; no admission behavior changes. DEV-612 moves the remote rows of these tests from the synthetic HMAC verifier to signed packages and adds the package rows below.

- **Facade fuzzing.** `FuzzAdmitTask` in the root package drives `Client.AdmitTask` with arbitrary task text and a fuzzed source: an embedded record for the exact text, an embedded record for other bytes, a signed remote package with a fuzzed signature and issuance time (each presentation installing its own single-use reference), a forged reference and a record from another instance. Every call either admits, publishing exactly one handle and one replay record with no run, or returns exactly one fixed, unwrapped public admission class and leaves no handle, replay record or run. Results are deterministic, an exact redelivery returns the same handle, and only an exact-bytes record of a reviewed origin or a validly signed package admits. `FuzzGate` gains seeds for escaped duplicate names, non-exact integers and deep nesting.
- **Strict JSON at the admission boundary** (the admission share of [GO-SPEC-ISSUE-001](specification-issues.md#go-spec-issue-001-duplicate-json-member-names-outside-labels)). `TestGateStepClasses` and `TestAdmitTaskStrictJSON`, for both reviewed origins, reject escaped-equivalent duplicate names at the top level and in `input`, `options`, `schedule` and `execution.limits`, escaped unknown names, `-0`, `1.0` and `1e0` for an integer field, an integer beyond the exact float range, a trailing value, a byte-order mark, invalid UTF-8 and nesting beyond the parse depth, and count the complete-input limit on original escaped bytes. An escaped spelling of the exact `schema_version` is the same version after decoding ([REQ-JSON-0004](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/json-decoding.md#REQ-JSON-0004)). Nesting beyond the 128-level parse depth is `task_schema_invalid`: the depth bound is a local implementation bound, not a registered limit. The 32,768-node parse bound cannot be reached within 65,536 bytes, because each value needs at least two bytes, so it is a backstop without an admission-level case.
- **Strict JSON in the signed package envelope** (DEV-566). `TestAdmitTaskSignedPackageStrictJSON` installs and presents package bytes through `Client.AdmitTask`. Equal, conflicting and escaped-equivalent duplicate members, escaped unknown and case-variant member names, a nested value where a string is required, nesting beyond the parse depth, a trailing value, a byte-order mark, invalid UTF-8 in a name or value and a lone surrogate escape are `task_source_unauthenticated`. A package whose `task` member is fully escaped and padded with trailing whitespace is `task_limit_exceeded` one byte over the 98,304-byte limit and admits at exactly the limit, so the limit is counted on original bytes. Escaped spellings of a member name and of a member value decode to the same package and admit. The package has string members only, so integer spellings do not apply to it.
- **No side effects on rejection.** `TestAdmitTaskRejectionsAreAtomic`, `TestAdmitTaskStrictJSON`, `TestAdmitTaskSignedPackageStrictJSON`, `TestAdmitTaskHostileSources` and `FuzzAdmitTask` assert that a rejected call leaves no task handle, replay record or run. Results and bundles do not exist yet, so there is no result state to inspect. The client has no transport, resolver or Network Context seam, so network absence cannot be observed with a recording transport. Instead, `TestAdmissionHasNoNetworkDependency` lists the transitive dependency closure of the root package and of `internal/admission` with `go list -deps` and requires it to contain no `net` package other than `net/netip` and `net/url`, and no `crypto/tls`, `os/exec`, `plugin` or `internal/testkit` package. This is structural evidence for the host build. It does not cover a future trusted adapter seam that could reach the network, and it does not run for mobile builds.
- **Hostile sources.** `TestAdmitTaskHostileSources` presents an empty, forged or foreign-instance reference, a record for other bytes, a bare task presented as remote, a package with an altered task byte, a package signed by an unknown key, a package moved to another issuer, remote `issued_at` 301 seconds ahead, a remote reference presented a second time and an expired record; each is `task_source_unauthenticated` with no state.
- **Signed packages.** `TestCcTrustSignedRemotePackages` covers every signed-package row of the internal validation plan, changing one property of an otherwise valid package per case and asserting the class of the first failing check and that the task codec is not reached; `TestCcTrustSignedPackageErrorsDiscloseNothing` checks that errors carry only the class; `TestCcTrustSignedPackageNotRetained` inspects the replay record and the delivery store after acceptance and rejection; root `TestCcTrustKeyRotationByPolicyReplacement` replaces the policy to remove a key and then an issuer, and, for REQ-TRUST-0055 (DEV-624), rejects a redelivery after a replacement that keeps an expired key without a replay write and resolves the same task bytes re-signed with a valid key to the retained acceptance with its original expiry. `FuzzSignedPackage` presents arbitrary bytes as a package and accepts only bytes that also verify under an independent decoding with `encoding/json` and `crypto/ed25519`.

### Fuzz Validation Record

On October 20, 2026, with Go 1.26.0 on darwin/arm64, two workers per target and minimization capped at 200 executions per input, these finite runs passed (DEV-594):

| Target | Duration | Executions |
|---|---:|---:|
| `FuzzAdmitTask` | 120 s | 78,550 |
| `FuzzGate` | 120 s | 51,328 |

DEV-612 changed the remote case of `FuzzAdmitTask` and added `FuzzSignedPackage`. On October 20, 2026, with Go 1.26.0 on darwin/arm64 and default fuzzing workers, these finite runs passed:

| Target | Duration | Executions |
|---|---:|---:|
| `FuzzSignedPackage` | 60 s | 506,606 |
| `FuzzAdmitTask` | 60 s | 30,022 |

DEV-566 changed no fuzz target in admission. On October 21, 2026, with Go 1.26.0 on darwin/arm64 and two workers, a 60-second `FuzzAdmitTask` regression run passed with 38,345 executions.

DEV-624 changed the order of delivery resolution and no fuzz target. On October 21, 2026, with Go 1.26.0 on darwin/arm64 and default fuzzing workers, these finite regression runs passed:

| Target | Duration | Executions |
|---|---:|---:|
| `FuzzGate` | 20 s | 7,372 |
| `FuzzSignedPackage` | 20 s | 51,064 |
| `FuzzAdmitTask` | 20 s | 32,766 |

Reproduce with `go test . -run='^$' -fuzz='^FuzzAdmitTask$' -fuzztime=120s -fuzzminimizetime=200x -parallel=2` and `go test ./internal/admission -run='^$' -fuzz='^FuzzGate$' -fuzztime=120s -fuzzminimizetime=200x -parallel=2`; the DEV-612 runs used `go test ./internal/admission -run='^$' -fuzz='^FuzzSignedPackage$' -fuzztime=60s` and `go test . -run='^$' -fuzz='^FuzzAdmitTask$' -fuzztime=60s`, and the DEV-624 runs used the `FuzzGate`, `FuzzSignedPackage` and `FuzzAdmitTask` commands above with only `-fuzztime=20s`. `FuzzAdmitTask` builds a fresh synthetic client for each input, so its throughput is far lower than the decoder targets. CI runs only the seed corpora, through `make test` and `make race`; it runs no fuzzing campaign. These finite runs are regression evidence, not exhaustive proof.

### Admission Case Families

Only `CC-TRUST-REPLAY-DUPLICATE-AND-ROLLBACK-HANDLING` (DEV-591), `CC-TRUST-LOCAL-POLICY-LIFECYCLE` (DEV-593), `CC-TRUST-ACCEPTED-TASK-ORIGINS` (DEV-594, rewritten for signed packages by DEV-612) and the four signed-package families added by DEV-612 (`CC-TRUST-SIGNED-REMOTE-PACKAGES`, `CC-TRUST-SIGNED-PACKAGE-DISCLOSURE`, `CC-TRUST-SIGNED-PACKAGE-RETENTION` and `CC-TRUST-KEY-ROTATION`) can be executed in full today. Every other admission-scope family stays `planned` because a row of its case plan needs something that does not exist yet:

| Family | Remains planned because |
|---|---|
| `CC-TRUST-OVERVIEW` | Its rows use NET-01 tasks; no production NET-01 codec is registered (DEV-454). |
| `CC-TRUST-TRUSTED-SOURCE-CONTEXT` | It runs an admitted task against a recording transport and result sink. Runs and the pending-output port exist synthetically since DEV-636, but its rows use registered test codecs and runners (DEV-454) and the production path seam (DEV-578). |
| `CC-TRUST-ISSUER-AUTHORIZATION` | Its rows use NET-01 and NET-05 tasks and a platform-capability dimension; no production codecs (DEV-454) or NET-05 composite matching exist. |
| `CC-TRUST-TARGET-REGISTRATION` | Its rows use TCP-01, TLS-01, HTTP-01 and DNS-01 tuples (DEV-454) and execution-time resolver and redirect checks, which belong to path handling and the registered runners (DEV-578, DEV-454). |
| `CC-TRUST-ADMISSION-PROCEDURE` | It needs transport, DNS and Network Context fakes, production proxy-compatibility rules (DEV-454) and a build variant that provides egress acquisition (DEV-578). |
| `CC-TRUST-TASK-IDENTIFIER-RULES` | Its campaign-identity rows compare task IDs emitted in results across installations with linkage tokens; results do not exist. |
| `CC-TRUST-TARGETED-DELIVERY` | Its linkage rows compare scheduling decisions and host task-selection hooks, which do not exist; run reservation exists since DEV-636, the host task-selection hooks remain DEV-455's. |
| `CC-TRUST-ADMISSION-ERRORS` | It needs a build that provides egress acquisition (DEV-578), result-sink and bundle fakes and captured log output. |
| `CC-TRUST-LOCAL-POLICY-LIFECYCLE-RUNTIME` | It needs the host policy adapter (DEV-455), cancellation of runs on replacement with recorded reasons (DEV-561) and Network Context caching (DEV-578); run reservation exists since DEV-636. |
| `CC-TRUST-REPLAY-AFTER-CLEARING` | It needs the persistent clearing capability (DEV-564, DEV-577). |
| `CC-LIMIT-TASK-AND-IDENTIFIER-LIMITS` | Its rows use a NET-05 `step_id` (DEV-454) and a counting reader at the native adapter (DEV-455). |
| `CC-LIMIT-LIMIT-HANDLING` | Its runtime rows need fake clocks, scripted transports and a resource tracker in execution (DEV-562, DEV-561), and its NET-05 row a NET-05 codec. |
| `CC-INPUT-JSON-STRUCTURE` | Its rows use the registered NET-01 and DNS-01 input schemas (DEV-454). |
| `CC-SAFETY-ENVELOPE`, `CC-SAFETY-USER-CONSENT`, `CC-SAFETY-EFFECTIVE-RISK`, `CC-SAFETY-SAFETY-EVALUATION` | Their per-attempt rows need a host consent provider and host or device restrictions (DEV-455), and rows using registered tests need production codecs (DEV-454). DEV-636 checks consent and device state at attempt start through synthetic capabilities ([Consent Handoff](#consent-handoff)); the families stay planned. |

The admission rows of several planned families are already exercised by `UT.*` cases recorded in the traceability manifest. Splitting those rows into separately executable cases, as DEV-593 did for local policy, is follow-up work; it does not change admission behavior.

## Validation

Admission and registry behavior is validated by the synthetic tests and executed case families described above, with the remaining families planned as recorded in [Hostile-Input Evidence](#hostile-input-evidence). This record is checked with `make check`, `make traceability`, the public requirement-reference checker and `git diff --check`. None of these establishes runtime conformance, platform or field support.
