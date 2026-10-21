# Trusted Context and Local Policy Models

## Authority and Scope

DEV-479 uses Draft `0.1.0` at immutable public `tv-rain/barometer-spec` revision `05c1364af4f9a604b617fca45f060949b557d265`: [Task Trust and Local Policy](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/library/task-trust.md), [Registered Initial Limits](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/library/registered-limits.md), [Safety](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/data-formats/task-safety.md), and [HTTP-01](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/tests/http-01.md) for the represented response-check fields. The [Go facade contract](public-api.md#trusted-initialization-and-admission) defines the separate reviewed adapter boundary.

`internal/model/task` owns these in-process values alongside the existing candidate task model. It remains a dependency leaf using only the standard library, with no I/O, decoding dependency, platform capability, scheduling, admission, execution, persistence or logging. No root-package exports, public JSON schema, third-party dependencies or native/Flutter changes are added. Constructors and typed getters provide the internal handoff surface; no unused adapter or storage interfaces are scaffolded.

These are structurally validated models. Successful construction is not authentication, release review, target approval, effective authorization, policy activation or task admission.

DEV-612 adopts public revision `e97502747c917a16bc2b71337535ff972bbd7d64` for the source context and adds trusted issuer keys to the policy snapshot ([Task Trust and Local Policy](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md), [Registered Initial Limits](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/registered-limits.md)). Remote provenance is now Barometer's verification of a signed package; the host no longer asserts remote authentication ([Signed Remote Task Packages](admission.md#signed-remote-task-packages)).

## Source Context and Exact Bytes

<!-- requirement-ref: REQ-TRUST-0044 -->

An internal reviewed adapter supplies `SourceInput`: origin and positive `uint64` policy generation, and for embedded origin a bounded lowercase issuer identifier ([REQ-TRUST-0044](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0044)). `SourceInput` has no authentication flag and no issuance time, and `NewSourceContext` refuses an issuer for remote origin, so a host cannot supply a remote issuer, issuance time or authentication verdict. A remote context starts unverified, with no issuer or issuance time; `WithVerifiedPackage` completes it with the issuer and whole-second UTC issuance time of a package that admission verified, and refuses an already completed, embedded or zero context. Admission is its only caller; calling it verifies nothing. `Verified` reports whether the issuer is established: always for a valid embedded context, and for remote only after completion. Unknown origins and invalid identifiers or timestamps reject atomically. No current clock is read and no future-offset or task-validity decision is made here.

`SourceContext` stores private immutable values; zero values are invalid. It is structurally separate from `Candidate`, task parameters, labels and initialization JSON.

`NewDeliveryRecord` binds a valid context to the SHA-256 digest of 1–131,072 exact received bytes (`task.MaxDeliveryBytes`) without retaining the payload: a serialized task for embedded origin, a signed package for remote origin. This implementation-local receive bound is deliberately above the registered 65,536-byte task and 98,304-byte package limits so that admission rejects oversized input as `task_limit_exceeded` instead of the record refusing it (DEV-586, DEV-612). `Matches` distinguishes semantically equivalent JSON with different bytes. It does not verify a signature, parse the task, mint a public `sourceRef`, retain replay state or return an admitted task. Admission verifies the same package bytes the record binds, and uses the decoded task bytes rather than reconstructing bytes from a candidate.

<!-- requirement-ref: REQ-TRUST-0043 -->
<!-- requirement-history: REQ-TRUST-0001 -->

The structural tests support [REQ-TRUST-0043](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0043), the successor of REQ-TRUST-0001, only: a host cannot construct a verified remote source. Trusted-time availability, package verification and fail-closed admission are implemented by admission over synthetic delivery records and synthetic keys (DEV-586 to DEV-594 and DEV-612, [Admission Design](admission.md)); the remote-delivery gate (DEV-595) was lifted on October 21, 2026 after the required reviews ([Remote-Delivery Gate](admission.md#remote-delivery-gate)), and no binding or host adoption exists yet (DEV-455).

## Trusted Issuer Keys

<!-- requirement-ref: REQ-TRUST-0051 -->
<!-- requirement-ref: REQ-LIMIT-0032 -->

`PolicyInput.Keys` carries the trusted issuer keys of [REQ-TRUST-0051](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0051). Each `TrustedKey` has a key ID, one issuer, a 32-byte Ed25519 public key held as an array (so it cannot alias caller memory), and `NotBefore` and `NotAfter` as whole UTC seconds. `NewLocalPolicy` rejects the whole policy, with the fixed `ErrTrustModel` and no partial snapshot, when any key record:

- has a key ID or issuer outside 1 to 64 lowercase ASCII letters, digits, `.`, `_` or `-`, or a key ID used by another key of the policy;
- names an issuer the same policy does not authorize;
- has a `NotBefore` or `NotAfter` that is zero, not UTC or not a whole second, a `NotAfter` not later than `NotBefore`, or a period above 365 days (`task.MaxKeyValidity`, [REQ-LIMIT-0032](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/registered-limits.md#REQ-LIMIT-0032));
- would make more than four keys for one issuer or more than 32 in the policy;
- has a public key that is not the canonical encoding of a curve point, or that is a point of small order.

`task.ValidPublicKey` decodes the point as in RFC 8032 section 5.1.3 with `math/big` and also rejects the encodings that the RFC leaves to implementations: a `y` coordinate not below the field prime and `x = 0` with the sign bit set. A point has small order exactly when eight times it is the identity, which the check computes. It runs only during policy validation over at most 32 keys and is not constant time; public keys are not secret. A policy without keys is valid and admits no remote task. `LocalPolicy.TrustedKey` returns a key by exact key ID. `TrustedKey.Covers` is the inclusive validity check of REQ-TRUST-0052 check 7.

Keys enter only inside the trusted policy bound at initialization or by policy replacement. Tasks, packages, collector responses and other delivered content cannot add, replace or extend them, and no Go, native or Flutter binding accepts a key.

## Policy Snapshots

`NewLocalPolicy` validates a complete `PolicyInput` before returning a valid `LocalPolicy`. It copies nested slices on ingress and getters return independent copies; callers must not mutate inputs concurrently with construction. Read-only access after construction is safe concurrently. Failure returns only the invalid zero snapshot and a fixed error, without changing an existing snapshot. No in-place policy update method exists; the root `Client` activates a snapshot at initialization and replaces it only by reinitialization with a strictly higher generation (DEV-593, see [Implemented Subset](public-api.md#implemented-subset)).

The snapshot carries:

- A positive policy generation and deployment permissions.
- Uniquely named issuer permission sets with exact test/version pairs, registration references, direct/proxy flags, maximum ordered risk, explicit medium-risk/composite enablement, an independent unknown-risk exception flag, a consent requirement, label namespaces and numeric ceilings.
- Uniquely named target registrations with test/version and role, typed destination tuples, path permissions, baseline risk, review owner/purpose, review generation, expiry and redirect registration references.

Unknown risk is unordered and never a maximum-risk value. `HighRisk` as a ceiling or `AllowUnknownRisk` is an explicit trusted policy input for later reviewed exception handling, not permission for a task to self-authorize. Medium-risk and NET-05 controls are separate opt-ins. An empty label list permits no labels; namespace entries are literal reviewed names, not wildcard rules. Admission matches a key equal to a namespace or below it at a `.` boundary; the privacy review of permitted namespaces remains a deployment obligation.

Deployment and issuer permissions remain independent. Construction does not compute a union or an effective grant; a broader issuer permission does not alter a local restriction. Admission intersects registered behavior, issuer authority, local policy, platform capability and task restrictions; DEV-587 implements the issuer and local terms ([Issuer Authorization](admission.md#issuer-authorization)), including the Go-local `.`-boundary reading of label namespaces. Consent inputs are requirements, never evidence that consent exists. Configuration approvals for linkage, proxy or context endpoints confer no issuer or target authority.

Numeric fields are explicit with no permissive zero defaults. Attempt ceilings, run/attempt/lifetime bounds, repeat cadence/jitter minima, retry delay/jitter bounds, result/bundle/storage maxima and local result retention are checked against the public registry. Zero traffic ceilings can represent a local-only policy. Protocol-specific and composite limits still belong to their validators. Replay retention is not configurable here and must follow the public replay contract in the future persistence/admission work. No numeric limit is clamped.

## Target Representation and Limits

Registrations reference exact test/version pairs and exact registration IDs. Duplicate IDs/pairs and unresolved registration or redirect references reject. Redirect references must name a redirect-role registration for the same test/version. Review generation is positive metadata; the public contract does not establish its ordering against the containing policy generation. Expiry is structurally required; construction does not decide whether it is current.

`TargetTuple` represents canonical hostname, IP (`netip.Addr`), port, scheme, URL, resolver name, bootstrap addresses, TLS server name, ALPN, expected status, body checks and header checks. Typed tuples retain absent properties without rewriting or resolving them. Basic checks reject malformed/noncanonical hosts, wildcard names, zoned/mapped addresses, credential-bearing or fragmented URLs, incompatible URL/hostname/scheme/port fields, malformed checks, sensitive response-header names and oversized collections/values. URL IP literals and a separately supplied literal address must agree. Body literals preserve bounded UTF-8; header names use lowercase canonical tokens. Constructors never fetch a target or extend authority through resolution or redirects.

A structurally valid tuple may still be incomplete or unsafe for a particular test. Test-owned validation must require that test's complete tuple and enforce options, address families, bootstrap rules, approved checks, risk, redirects, resource constraints and normalized exact matching. Destination/address policy, non-user-specific identifiers, secrets in static URLs, ownership and review validity cannot be established by syntax. Immediate pre-contact DNS/address/redirect validation remains mandatory, including when using a proxy. These model getters perform exact ID lookup only; they are not a target matcher or network authorization decision.

Finite implementation-local limits keep trusted inputs bounded: at most 256 issuers, targets, test/version pairs or registration references per respective collection; 16 label namespaces; 16 bootstrap addresses; 8 ALPN entries of at most 32 printable ASCII bytes; 8 body and 8 header checks with 256-byte literals; 2,048-byte URLs; 253-byte hosts with 63-byte labels. Review owner/purpose are bounded to 256 printable ASCII bytes. Source/expiry fractional timestamp text is bounded to 65,536 digits. These are local representation choices, not additional public schema fields, device measurements or safety approvals.

## Revocation and Serialization Boundaries

`NewRevocations` copies bounded issuer/registration ID lists under a positive generation. At least one list must be nonempty; duplicate/invalid IDs reject. It deliberately does not compare generations with an active policy, install a revocation, decide what happens to queued/running tasks, change expiry, cancel activity or persist anything. Key, issuer and target revocation in version `0.1.0` is policy replacement with a higher generation (REQ-TRUST-0054, see [Signed Remote Task Packages](admission.md#signed-remote-task-packages)); this input is not used for it, and there is no fast revocation channel. Constructing this input cannot mutate a `LocalPolicy`.

All trusted input containers and snapshots explicitly reject JSON import and export, including import into an existing valid value. Ordinary task and initialization schemas have no trust/policy fields. Trust-looking text in a permitted task label remains opaque text. The public facade (`Client.AdmitTask`, DEV-592) resolves adapter-installed delivery records by opaque reference; it never deserializes a source or policy supplied by Flutter. Internal constructors can still be called by trusted Go code: this is a structural boundary, not protection against a compromised host or adapter. Models must never be placed in results, traffic, logs, analytics or crash reports.

Errors are fixed `ErrTrustModel` or `ErrTrustEncoding` sentinels without payloads, source references, issuer/target values, raw parser errors or verification material. They are operational failures and never Measurement Results.

## Synthetic Evidence and Remaining Work

| Case | Evidence scope |
|---|---|
| `TestSourceContextInvariantsAndByteBinding`, `TestMalformedSourceRejectedAtomically` | Origin and generation shape, refusal of a host-supplied remote issuer, completion only of an unverified remote context, byte-boundary and exact-byte binding; no cryptographic or clock assurance. |
| `TestPolicyCarriesBoundedTrustedKeys`, `TestPolicyRejectsMalformedOrAmbiguousKeys`, `TestTrustedKeyCoversInclusiveValidity` | Trusted key records: exact lookup without aliasing, count bounds at and above 4 per issuer and 32 per policy, validity at and above 365 days, malformed identifiers and times, unauthorized issuers, rejected public keys and inclusive validity ends. |
| `TestSmallOrderPointsDecodeAndAreRejected`, `TestNonCanonicalAndOffCurveKeysRejected`, `TestGeneratedKeysAccepted` | The eight small-order points decode and are rejected; non-canonical and off-curve encodings fail decoding; standard-library keys decode, round-trip and are accepted. Synthetic keys only; not a cryptographic review. |
| `TestPolicyRejectsMalformedOrAmbiguousAuthority`, `TestRegistrationTuplesAreExactAndBounded` | Atomic model rejection, ambiguous IDs/references, explicit limits and canonical typed tuple structure. |
| `TestPolicySnapshotsSurviveMutationAndConcurrentReads` | Deep copies on ingress/egress, exact ID lookup and concurrent immutable reads. |
| `TestPolicyKeepsRestrictionsSeparateWithoutGrantingAuthority` | Independent local/issuer restrictions and zero-traffic local-only representation. |
| `TestRevocationInputsHaveNoActivationSideEffects` | Immutable inputs and no policy activation effects. |
| `TestTrustModelsCannotBeImportedOrExportedAsJSON`, `TestSerializedInputCannotSupplyTrustOrPolicy` | JSON boundary rejection, unchanged valid state, task/config injection rejection and opaque labels. |

Validation commands are `make check`, `git diff --check`, and `python3 -B ../barometer-spec/scripts/check_requirements.py --consumer .` against the explicit public checkout. The fixtures are synthetic and perform no network measurement. Admission, trusted-time checks, replay over a synthetic store and facade integration were later delivered by DEV-586 to DEV-594 ([Admission Design](admission.md)). Production protected persistence, the reviews that gate remote delivery (DEV-595), native delivery adapters and device validation remain pending. There is no full runtime conformance, platform, field or release-support claim.

Local validation on October 15, 2026 passed the complete `make check`: formatting, host build, unit tests, race detection, static analysis, module integrity, diff whitespace and pure-Go Android ARM64/iOS ARM64 compilation. Requirement-reference structure validation passed against the explicit public checkout above; allocation history, coverage and semantic conformance were not assessed by that checker. Native bindings, devices, CI, releases and field behavior were not tested by this change. No sibling repository was modified.
