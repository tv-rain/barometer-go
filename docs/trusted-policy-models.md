# Trusted Context and Local Policy Models

## Authority and Scope

DEV-479 uses Draft `0.1.0` at immutable public `tv-rain/barometer-spec` revision `05c1364af4f9a604b617fca45f060949b557d265`: [Task Trust and Local Policy](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/library/task-trust.md), [Registered Initial Limits](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/library/registered-limits.md), [Safety](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/data-formats/task-safety.md), and [HTTP-01](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/tests/http-01.md) for the represented response-check fields. The [Go facade contract](public-api.md#trusted-initialization-and-admission) defines the separate reviewed adapter boundary.

`internal/model/task` owns these in-process values alongside the existing candidate task model. It remains a dependency leaf using only the standard library, with no I/O, decoding dependency, platform capability, scheduling, admission, execution, persistence or logging. No root-package exports, public JSON schema, third-party dependencies or native/Flutter changes are added. Constructors and typed getters provide the internal handoff surface; no unused adapter or storage interfaces are scaffolded.

These are structurally validated models. Successful construction is not authentication, release review, target approval, effective authorization, policy activation or task admission.

## Source Context and Exact Bytes

An internal reviewed adapter supplies `SourceInput` after establishing provenance. `NewSourceContext` requires a bounded lowercase issuer identifier and positive `uint64` policy generation. Remote origin requires authentication and issuance time; reviewed embedded origin requires authentication to be false and issuance time absent. Unknown origins and invalid timestamp representations reject atomically. Timestamps preserve exact fractional precision using the existing UTC `Timestamp` representation. No current clock is read and no future-offset or task-validity decision is made here.

`SourceContext` stores private immutable values; zero values are invalid and the issuance pointer is copied. It is structurally separate from `Candidate`, task parameters, labels and initialization JSON.

`NewDeliveryRecord` binds a valid context to the SHA-256 digest of 1–131,072 exact received bytes (`task.MaxDeliveryBytes`) without retaining the payload. This implementation-local receive bound is deliberately above the registered 65,536-byte task limit so that admission rejects an oversized task as `task_limit_exceeded` before parsing instead of the record refusing it (DEV-586). `Matches` distinguishes semantically equivalent JSON with different bytes. This value models the exact-byte association required by the future delivery owner. It does not verify integrity/authentication, parse the task, mint a public `sourceRef`, retain replay state or return an admitted task. The adapter must verify the same bytes it binds; admission must use that identity rather than reconstructing bytes from a candidate.

<!-- requirement-ref: REQ-TRUST-0001 -->

The missing-authentication/issuance structural tests support the future [REQ-TRUST-0001](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/library/task-trust.md#REQ-TRUST-0001) boundary only. Trusted-time availability, actual source verification and fail-closed runtime admission are not implemented by these tests or constructors.

## Policy Snapshots

`NewLocalPolicy` validates a complete `PolicyInput` before returning a valid `LocalPolicy`. It copies nested slices on ingress and getters return independent copies; callers must not mutate inputs concurrently with construction. Read-only access after construction is safe concurrently. Failure returns only the invalid zero snapshot and a fixed error, without changing an existing snapshot. No in-place policy update method exists.

The snapshot carries:

- A positive policy generation and deployment permissions.
- Uniquely named issuer permission sets with exact test/version pairs, registration references, direct/proxy flags, maximum ordered risk, explicit medium-risk/composite enablement, an independent unknown-risk exception flag, a consent requirement, label namespaces and numeric ceilings.
- Uniquely named target registrations with test/version and role, typed destination tuples, path permissions, baseline risk, review owner/purpose, review generation, expiry and redirect registration references.

Unknown risk is unordered and never a maximum-risk value. `HighRisk` as a ceiling or `AllowUnknownRisk` is an explicit trusted policy input for later reviewed exception handling, not permission for a task to self-authorize. Medium-risk and NET-05 controls are separate opt-ins. An empty label list permits no labels; namespace entries are literal reviewed names, not wildcard rules. Their task-label matching rules and privacy review must be enforced downstream.

Deployment and issuer permissions remain independent. Construction does not compute a union or an effective grant; a broader issuer permission does not alter a local restriction. Future admission must intersect registered behavior, issuer authority, local policy, platform capability and task restrictions. Consent inputs are requirements, never evidence that consent exists. Configuration approvals for linkage, proxy or context endpoints confer no issuer or target authority.

Numeric fields are explicit with no permissive zero defaults. Attempt ceilings, run/attempt/lifetime bounds, repeat cadence/jitter minima, retry delay/jitter bounds, result/bundle/storage maxima and local result retention are checked against the public registry. Zero traffic ceilings can represent a local-only policy. Protocol-specific and composite limits still belong to their validators. Replay retention is not configurable here and must follow the public replay contract in the future persistence/admission work. No numeric limit is clamped.

## Target Representation and Limits

Registrations reference exact test/version pairs and exact registration IDs. Duplicate IDs/pairs and unresolved registration or redirect references reject. Redirect references must name a redirect-role registration for the same test/version. Review generation is positive metadata; the public contract does not establish its ordering against the containing policy generation. Expiry is structurally required; construction does not decide whether it is current.

`TargetTuple` represents canonical hostname, IP (`netip.Addr`), port, scheme, URL, resolver name, bootstrap addresses, TLS server name, ALPN, expected status, body checks and header checks. Typed tuples retain absent properties without rewriting or resolving them. Basic checks reject malformed/noncanonical hosts, wildcard names, zoned/mapped addresses, credential-bearing or fragmented URLs, incompatible URL/hostname/scheme/port fields, malformed checks, sensitive response-header names and oversized collections/values. URL IP literals and a separately supplied literal address must agree. Body literals preserve bounded UTF-8; header names use lowercase canonical tokens. Constructors never fetch a target or extend authority through resolution or redirects.

A structurally valid tuple may still be incomplete or unsafe for a particular test. Test-owned validation must require that test's complete tuple and enforce options, address families, bootstrap rules, approved checks, risk, redirects, resource constraints and normalized exact matching. Destination/address policy, non-user-specific identifiers, secrets in static URLs, ownership and review validity cannot be established by syntax. Immediate pre-contact DNS/address/redirect validation remains mandatory, including when using a proxy. These model getters perform exact ID lookup only; they are not a target matcher or network authorization decision.

Finite implementation-local limits keep trusted inputs bounded: at most 256 issuers, targets, test/version pairs or registration references per respective collection; 16 label namespaces; 16 bootstrap addresses; 8 ALPN entries of at most 32 printable ASCII bytes; 8 body and 8 header checks with 256-byte literals; 2,048-byte URLs; 253-byte hosts with 63-byte labels. Review owner/purpose are bounded to 256 printable ASCII bytes. Source/expiry fractional timestamp text is bounded to 65,536 digits. These are local representation choices, not additional public schema fields, device measurements or safety approvals.

## Revocation and Serialization Boundaries

`NewRevocations` copies bounded issuer/registration ID lists under a positive generation. At least one list must be nonempty; duplicate/invalid IDs reject. It deliberately does not compare generations with an active policy, install a revocation, decide what happens to queued/running tasks, change expiry, cancel activity or persist anything. Operational revocation remains unresolved in the internal task-trust review; defining that behavior requires the public specification and review sequence before implementation. Constructing this input cannot mutate a `LocalPolicy`.

All trusted input containers and snapshots explicitly reject JSON import and export, including import into an existing valid value. Ordinary task and initialization schemas have no trust/policy fields. Trust-looking text in a permitted task label remains opaque text. The future public facade must resolve adapter-installed delivery records; it must never deserialize a source or policy supplied by Flutter. Internal constructors can still be called by trusted Go code: this is a structural boundary, not protection against a compromised host or adapter. Models must never be placed in results, traffic, logs, analytics or crash reports.

Errors are fixed `ErrTrustModel` or `ErrTrustEncoding` sentinels without payloads, source references, issuer/target values, raw parser errors or verification material. They are operational failures and never Measurement Results.

## Synthetic Evidence and Remaining Work

| Case | Evidence scope |
|---|---|
| `TestSourceContextInvariantsAndByteBinding`, `TestMalformedSourceRejectedAtomically` | Origin/authentication/time shape, copied time, byte-boundary and exact-byte binding; no cryptographic or clock assurance. |
| `TestPolicyRejectsMalformedOrAmbiguousAuthority`, `TestRegistrationTuplesAreExactAndBounded` | Atomic model rejection, ambiguous IDs/references, explicit limits and canonical typed tuple structure. |
| `TestPolicySnapshotsSurviveMutationAndConcurrentReads` | Deep copies on ingress/egress, exact ID lookup and concurrent immutable reads. |
| `TestPolicyKeepsRestrictionsSeparateWithoutGrantingAuthority` | Independent local/issuer restrictions and zero-traffic local-only representation. |
| `TestRevocationInputsHaveNoActivationSideEffects` | Immutable inputs and no policy activation effects. |
| `TestTrustModelsCannotBeImportedOrExportedAsJSON`, `TestSerializedInputCannotSupplyTrustOrPolicy` | JSON boundary rejection, unchanged valid state, task/config injection rejection and opaque labels. |

Validation commands are `make check`, `git diff --check`, and `python3 -B ../barometer-spec/scripts/check_requirements.py --consumer .` against the explicit public checkout. The fixtures are synthetic and perform no network measurement. Complete admission, monotonic policy activation, trust/time availability, protected persistence, replay, operational revocation, native delivery adapters, facade integration and device validation remain pending. There is no full runtime conformance, platform, field or release-support claim.

Local validation on October 15, 2026 passed the complete `make check`: formatting, host build, unit tests, race detection, static analysis, module integrity, diff whitespace and pure-Go Android ARM64/iOS ARM64 compilation. Requirement-reference structure validation passed against the explicit public checkout above; allocation history, coverage and semantic conformance were not assessed by that checker. Native bindings, devices, CI, releases and field behavior were not tested by this change. No sibling repository was modified.
