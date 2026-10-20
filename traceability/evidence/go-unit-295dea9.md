# Go Unit-Test Evidence at 295dea9

DEV-590 records executed synthetic Go unit tests that partially evidence requirements in the cross-cutting manifest after rejecting tasks that require unsupported capabilities at admission. Each section below is the conformance evidence for one cross-cutting manifest case. The measurement manifest keeps its earlier baseline and its [evidence at c5f4e92](go-unit-c5f4e92.md).

## Execution Record

- Implementation: `https://github.com/tv-rain/barometer-go` commit `295dea9139956bc73eb04d521001ae76f0d574b8` (DEV-590 commit A), exported with `git archive` into an empty directory outside any checkout.
- Specification baseline: `https://github.com/tv-rain/barometer-spec` commit `d93aff93cbc17e7c065fb40d8f6e2e01a697bf35` (Draft `0.1.0`, public PR #14 merge), unchanged from the previous report.
- Environment: go1.26.0 darwin/arm64 (`GOTOOLCHAIN=local`, `GOFLAGS` empty, `CGO_ENABLED=1`), default test timeout, no race detector.
- Command, from the export root: `go test -count=1 -json ./...`
- Outcome: exit status 0; all 175 top-level tests passed, none failed or skipped. Each case below names one passing top-level test; a top-level pass includes all of its subtests.
- Executed on October 20, 2026. Inputs are the synthetic, in-repository fixtures at the implementation commit; no participant, measurement or private data is involved.

## Carried-Forward Cases

The 90 cases carried from the [41c6f06 report](go-unit-41c6f06.md) were re-executed at this baseline and keep their recorded aspects. Between `41c6f06` and `295dea9` the `composite enabled` row of `TestGateIssuerAuthorization` now expects `task_unsupported` at step 8 instead of admission, because composite steps cannot be matched yet; the recorded aspects of that case cover composite enablement at step 7 and are unchanged. The other carried test bodies are unchanged, and no other carried test admits a task with an egress context requirement or an enabled NET-05 entry.

## New Cases

DEV-590 adds 3 cases for the rejection of unsupported capabilities in `internal/admission`: egress Network Context requirements, their precedence within step 6 and against steps 7 to 9, and NET-05 composite profiles. They use the synthetic SYN-01 registry entry, a synthetic NET-05 composite entry with a test-scoped codec, synthetic target registrations and a synthetic local policy.

## Limits

Every case is partial evidence. A passing case shows only the stated aspects of each linked requirement; the notes in the manifest record the remaining obligations. The tests exercise pure-Go model, decoding, validation, admission-gate, result and bundle code on one host with synthetic delivery records, synthetic registry entries and target rules, a synthetic local policy and a fake clock. They do not execute the planned conformance families, the execution-time address and redirect checks of REQ-TRUST-0016, proxy compatibility, composite-step matching, egress Network Context acquisition, a trusted local assessment resolving unknown risk factors, consent validity at attempt start, per-attempt safety re-evaluation, replay or acceptance steps, facade wiring, a measurement runner, native bindings, a device or a field deployment, and they establish no implemented requirement, released API, platform or field support.

## Cases

<a id="UT.internal.admission.TestGateAcceptsReviewedOrigins"></a>

### UT.internal.admission.TestGateAcceptsReviewedOrigins

- Case: `UT.internal.admission.TestGateAcceptsReviewedOrigins`
- Result: passed
- Test: `TestGateAcceptsReviewedOrigins` in `internal/admission/admit_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateAcceptsReviewedOrigins$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateAcceptsReviewedOrigins in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0001: Asserts that a remote task is gated only through a verifier-installed authenticated record with issued_at one hour before, equal to and 300 seconds after the fake trusted clock, and that the outcome carries remote origin.
  - REQ-TRUST-0004: Asserts that an embedded task installed through the embedded-origin installer passes the gates with embedded origin, its issuer, the trusted acceptance time and the registry entry, decoded through the registry-selected synthetic codec with defaults applied.

<a id="UT.internal.admission.TestGateAdmissionLimitHandling"></a>

### UT.internal.admission.TestGateAdmissionLimitHandling

- Case: `UT.internal.admission.TestGateAdmissionLimitHandling`
- Result: passed
- Test: `TestGateAdmissionLimitHandling` in `internal/admission/authority_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateAdmissionLimitHandling$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateAdmissionLimitHandling in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0026: Asserts that complete task JSON, task_id, input, options, label count, max_runs, max_duration_ms, max_bytes_sent, max_bytes_received and max_connections are each accepted at their registered maximum and rejected one unit above it as task_limit_exceeded at admission, with an unused recording transport that sees zero calls. The NET-05 nine-step row and runtime limit handling are not tested.
  - REQ-LIMIT-0030: Asserts that label values naming limits are kept as opaque labels with effective limits unchanged, and that limit-like members in input, options, schedule, execution and safety and an unknown top-level field are task_schema_invalid; also asserts that issuer or deployment policy ceilings above the registered attempt, run, attempt-count and validity maxima are refused by policy construction.

<a id="UT.internal.admission.TestGateAuthorityPrecedence"></a>

### UT.internal.admission.TestGateAuthorityPrecedence

- Case: `UT.internal.admission.TestGateAuthorityPrecedence`
- Result: passed
- Test: `TestGateAuthorityPrecedence` in `internal/admission/authority_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateAuthorityPrecedence$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateAuthorityPrecedence in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0037: Asserts first-failing-step precedence around step 7 and step 8: an unknown top-level field, an expired task and a value above a registered maximum each keep their earlier class with an unknown issuer, an unknown issuer precedes invalid input, a policy ceiling precedes path authority, invalid input precedes path authority, and path authority precedes the risk bound.

<a id="UT.internal.admission.TestGateConsentRequirement"></a>

### UT.internal.admission.TestGateConsentRequirement

- Case: `UT.internal.admission.TestGateConsentRequirement`
- Result: passed
- Test: `TestGateConsentRequirement` in `internal/admission/safety_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateConsentRequirement$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateConsentRequirement in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SAFETY-0021: Asserts that a task requires_user_consent of false, with issuer and deployment consent requirements off, is admitted with no consent requirement recorded on the gated value.
  - REQ-SAFETY-0022: Asserts for all eight combinations of task, issuer and deployment consent requirement that the recorded requirement is set whenever the issuer or the deployment requires consent, whatever the task value. Consent validity at attempt start is not tested.
  - REQ-SAFETY-0028: Asserts that a task value of false never clears an issuer or deployment consent requirement, and that declaring requires_user_consent true does not relax a risk bound that rejects the same task.

<a id="UT.internal.admission.TestGateEffectiveRisk"></a>

### UT.internal.admission.TestGateEffectiveRisk

- Case: `UT.internal.admission.TestGateEffectiveRisk`
- Result: passed
- Test: `TestGateEffectiveRisk` in `internal/admission/safety_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateEffectiveRisk$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateEffectiveRisk in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SAFETY-0004: Asserts with synthetic registry baselines low, medium and unknown that the registered baseline is an input to the effective risk recorded on the gated value.
  - REQ-SAFETY-0005: Asserts that a declared level above the registered baseline raises effective risk and that no declared level reduces it below the baseline.
  - REQ-SAFETY-0023: Asserts for every combination of registered baseline (low, medium, unknown), declared level and matched target-registration baseline (low, medium, high, unknown), under a policy permitting every level, that the recorded effective risk equals the most restrictive input as computed independently in the test, with unknown dominating. Options, schedule or other local assessments are not inputs and are not tested.
  - REQ-SAFETY-0024: Asserts that effective risk is never below the registered baseline for any declared level.
  - REQ-SAFETY-0025: Asserts that a low declaration on a medium baseline is admitted with effective risk medium, never low.
  - REQ-SAFETY-0026: Asserts that a declaration higher than the registered baseline is accepted and becomes the effective risk.

<a id="UT.internal.admission.TestGateEgressContextPrecedence"></a>

### UT.internal.admission.TestGateEgressContextPrecedence

- Case: `UT.internal.admission.TestGateEgressContextPrecedence`
- Result: passed
- Test: `TestGateEgressContextPrecedence` in `internal/admission/capability_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateEgressContextPrecedence$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateEgressContextPrecedence in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0036: Asserts that an unrecognized risk factor type together with an egress context requirement is rejected as task_unsupported.
  - REQ-TRUST-0037: Asserts that an unknown field, task_id over 128 bytes, expiry, a value above a registered limit and an invalid risk factor keep their own class ahead of the step-6 egress check, and that the egress check precedes invalid input, a test/version not permitted, policy rollback, a label outside the permitted namespaces, a path not permitted, an unregistered target and a risk bound, with the registry codec never reached.

<a id="UT.internal.admission.TestGateErrorsCarryNoData"></a>

### UT.internal.admission.TestGateErrorsCarryNoData

- Case: `UT.internal.admission.TestGateErrorsCarryNoData`
- Result: passed
- Test: `TestGateErrorsCarryNoData` in `internal/admission/admit_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateErrorsCarryNoData$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateErrorsCarryNoData in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0033: Asserts that each rejection is a task.AdmissionError whose text is the fixed class string, that repeating an identical failing delivery returns the same error, and that rejection leaves the delivery record resolvable and a later valid task accepted.
  - REQ-TRUST-0035: Embeds canaries in payload, target, labels, issuer, codec error and clock error text and asserts none appears in the error text, %v, %+v or %#v formatting, and that errors.Unwrap returns nil.

<a id="UT.internal.admission.TestGateIssuerAndLocalCeilings"></a>

### UT.internal.admission.TestGateIssuerAndLocalCeilings

- Case: `UT.internal.admission.TestGateIssuerAndLocalCeilings`
- Result: passed
- Test: `TestGateIssuerAndLocalCeilings` in `internal/admission/authority_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateIssuerAndLocalCeilings$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateIssuerAndLocalCeilings in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0030: Asserts that a task value one unit beyond an issuer or deployment ceiling is task_limit_exceeded for every task-bounded ceiling, including repeat interval, jitter and retry delay minima, and is never clamped.
  - REQ-TRUST-0010: Asserts that each numeric ceiling narrowed in the issuer, and separately in the deployment policy, binds while the other side is permissive, and that result and retention ceilings are carried as the lower of the two sides in both directions.
  - REQ-TRUST-0011: Asserts that issuer and deployment ceilings bound attempt duration, bytes sent and received, connections, max_runs, retry max_attempts, validity lifetime, repeat interval and jitter minima, and retry initial delay, maximum delay and jitter.
  - REQ-TRUST-0012: Asserts for each task-bounded ceiling that a value at the bound and one unit below it is accepted, the lower value is carried as the effective value, and one unit beyond is rejected as task_limit_exceeded.

<a id="UT.internal.admission.TestGateIssuerAuthorization"></a>

### UT.internal.admission.TestGateIssuerAuthorization

- Case: `UT.internal.admission.TestGateIssuerAuthorization`
- Result: passed
- Test: `TestGateIssuerAuthorization` in `internal/admission/authority_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateIssuerAuthorization$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateIssuerAuthorization in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0010: Narrows test/version, network path, maximum risk, medium-risk enablement, unknown-risk exception, label namespaces and NET-05 composite enablement in the issuer and then in the deployment policy, with the other side permissive, and asserts each narrowing binds.
  - REQ-TRUST-0011: Asserts that permitted test versions, direct and proxy path use, maximum risk, medium-risk and NET-05 enablement, the unknown-risk exception and label namespaces at a '.' boundary are each enforced as task_issuer_not_authorized when not granted. Registration-reference matching and consent evaluation are not tested.
  - REQ-TRUST-0037: Asserts task_issuer_not_authorized for a direct or proxy path not permitted by issuer or deployment policy and for risk above an issuer or local bound, or unknown risk without the reviewed exception on both sides; locally determined risk is not tested.

<a id="UT.internal.admission.TestGateIssuerLocalConflicts"></a>

### UT.internal.admission.TestGateIssuerLocalConflicts

- Case: `UT.internal.admission.TestGateIssuerLocalConflicts`
- Result: passed
- Test: `TestGateIssuerLocalConflicts` in `internal/admission/authority_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateIssuerLocalConflicts$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateIssuerLocalConflicts in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0010: Asserts that when issuer and deployment disagree the stricter attempt duration binds in both directions, disjoint path grants permit neither path, registration IDs intersect, and consent is required when the issuer or the task requires it, with effective paths narrowed to the task's path.

<a id="UT.internal.admission.TestGateLocallyDeterminedRiskBounds"></a>

### UT.internal.admission.TestGateLocallyDeterminedRiskBounds

- Case: `UT.internal.admission.TestGateLocallyDeterminedRiskBounds`
- Result: passed
- Test: `TestGateLocallyDeterminedRiskBounds` in `internal/admission/safety_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateLocallyDeterminedRiskBounds$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateLocallyDeterminedRiskBounds in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SAFETY-0010: Asserts that unknown locally determined risk from a matched registration is rejected unless the reviewed unknown-risk exception is granted by both issuer and deployment, and admitted when both grant it.
  - REQ-SAFETY-0023: Asserts that a matched registration's reviewed baseline raises effective risk on a low registered baseline and low declaration, and is carried as the effective risk when permitted.
  - REQ-SAFETY-0027: Asserts at admission that effective risk raised by a matched registration must be permitted by local policy: above the lower maximum or without medium-risk enablement it is rejected as task_issuer_not_authorized, narrowed in the issuer and then in the deployment. Consent and host or device conditions are not tested.
  - REQ-TRUST-0011: Asserts that the issuer maximum risk, medium-risk enablement and unknown-risk exception bound locally determined risk, each narrowed in the issuer with the deployment permissive.

<a id="UT.internal.admission.TestGateNoTaskControlledAuthorityOrInterpolation"></a>

### UT.internal.admission.TestGateNoTaskControlledAuthorityOrInterpolation

- Case: `UT.internal.admission.TestGateNoTaskControlledAuthorityOrInterpolation`
- Result: passed
- Test: `TestGateNoTaskControlledAuthorityOrInterpolation` in `internal/admission/admit_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateNoTaskControlledAuthorityOrInterpolation$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateNoTaskControlledAuthorityOrInterpolation in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0002: Asserts that top-level authenticated, issuer_id, policy_generation, trusted and source fields are task_schema_invalid, and that a label reading 'authenticated: true' leaves the gated source context unauthenticated embedded origin.
  - REQ-TRUST-0026: Asserts that recipient selectors (installation, linkage, bundle, run, account, advertising, device, push, IP, location, eligibility and selector fields, top-level and nested) are task_schema_invalid at admission.
  - REQ-TRUST-0027: Asserts that placeholder text for country, application version, cohort and network in input, options and labels is kept literal in the gated candidate and that identical bytes on two instances yield equal candidates. Host-side eligibility policy is not tested.

<a id="UT.internal.admission.TestGatePolicyAndIssuerResolution"></a>

### UT.internal.admission.TestGatePolicyAndIssuerResolution

- Case: `UT.internal.admission.TestGatePolicyAndIssuerResolution`
- Result: passed
- Test: `TestGatePolicyAndIssuerResolution` in `internal/admission/authority_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGatePolicyAndIssuerResolution$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGatePolicyAndIssuerResolution in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0010: Asserts that issuer authority is resolved by exact issuer ID for embedded and authenticated remote sources, that an issuer absent from the policy is task_issuer_not_authorized, and that remote validity is bounded from the earlier of issuance and acceptance.
  - REQ-TRUST-0037: Asserts at step 7 that no installed policy and a source policy_generation above the active one are task_policy_unavailable, one below is policy_rollback, and an unknown issuer is task_issuer_not_authorized, with generation checked before the issuer and the codec not reached.

<a id="UT.internal.admission.TestGatePrecedence"></a>

### UT.internal.admission.TestGatePrecedence

- Case: `UT.internal.admission.TestGatePrecedence`
- Result: passed
- Test: `TestGatePrecedence` in `internal/admission/admit_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGatePrecedence$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGatePrecedence in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0037: Asserts first-failing-step precedence among implemented steps: unregistered version before a 129-byte task_id, unknown field before unregistered version or over-limit labels, task_id characters before a field limit, task_id length before expiry, expiry and unrecognized risk factor type before invalid input, provenance, freshness and trusted time before the byte limit.

<a id="UT.internal.admission.TestGateRejectsCompositeProfiles"></a>

### UT.internal.admission.TestGateRejectsCompositeProfiles

- Case: `UT.internal.admission.TestGateRejectsCompositeProfiles`
- Result: passed
- Test: `TestGateRejectsCompositeProfiles` in `internal/admission/capability_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateRejectsCompositeProfiles$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateRejectsCompositeProfiles in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0018: Asserts that every NET-05 rejection, unregistered, without composite enablement or registered and enabled, causes no call or open connection on the unused recording transport.
  - REQ-TRUST-0036: Asserts that NET-05 absent from the registry, at the registered or another version, is task_unsupported even under a policy that enables composite tasks, and that a registered NET-05 composite enabled by both issuer and deployment is task_unsupported because composite steps cannot be matched, while SYN-01 is admitted under the same policy. Composite-step matching is not provided and is not tested.
  - REQ-TRUST-0037: Asserts that policy rollback and missing composite enablement in the issuer or in the deployment (task_issuer_not_authorized) precede the step-8 composite check, that invalid input and an egress requirement precede it, and that it precedes path authority, target registration and risk bounds.

<a id="UT.internal.admission.TestGateRejectsEgressContextRequirements"></a>

### UT.internal.admission.TestGateRejectsEgressContextRequirements

- Case: `UT.internal.admission.TestGateRejectsEgressContextRequirements`
- Result: passed
- Test: `TestGateRejectsEgressContextRequirements` in `internal/admission/capability_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateRejectsEgressContextRequirements$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateRejectsEgressContextRequirements in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0018: Asserts that rejecting a task with an egress context requirement, for embedded and remote origin, causes no call or open connection on the unused recording transport, so admission performs no context lookup, endpoint call or other network activity.
  - REQ-TRUST-0036: Asserts that a valid task requiring country_code, region_code or asn individually or all three, a country allowlist alone or with an empty required_fields, or required fields with an allowlist, is rejected as task_unsupported for embedded and remote origin; that tasks without a network object, without context, with an empty context or with an empty required_fields are admitted; and that an invalid context (unknown field name, duplicate field, empty allowlist, lowercase code or unknown member) is task_schema_invalid instead. Egress acquisition is not provided and is not tested.
  - REQ-TRUST-0037: Asserts the step-6 row for a valid field that requires a capability the implementation does not provide (task_unsupported), with the registry codec never reached.

<a id="UT.internal.admission.TestGateStepClasses"></a>

### UT.internal.admission.TestGateStepClasses

- Case: `UT.internal.admission.TestGateStepClasses`
- Result: passed
- Test: `TestGateStepClasses` in `internal/admission/admit_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateStepClasses$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateStepClasses in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0004: Asserts on original encoded bytes: complete task 65,535 and 65,536 accepted and 65,537 task_limit_exceeded; input 49,152 accepted and 49,153 rejected, including an escaped value whose decoded length is smaller; options 8,192 and 8,193; 16 and 17 labels; labels of 2,048 and 2,049 encoded bytes.
  - REQ-LIMIT-0005: Asserts that an unknown field taking a task to 65,537 bytes is task_limit_exceeded while the same field within the limit is task_schema_invalid, and that 65,537 bytes of malformed input are task_limit_exceeded before parsing.
  - REQ-LIMIT-0010: Asserts that remote issued_at up to 300 seconds after trusted time passes and 300.5 or 301 seconds is rejected, and that unavailable trusted time fails remote and embedded admission.
  - REQ-SAFETY-0017: Asserts that a task with an unrecognized risk factor type is rejected as task_unsupported rather than silently ignored, and that the same type with an invalid level is task_schema_invalid.
  - REQ-TRUST-0001: Asserts task_source_unauthenticated for a missing, malformed, foreign-instance, byte-mismatched or expired delivery record and for a delivery whose verifier reports encryption without issuer authentication, and task_trust_unavailable when trusted time is unavailable at resolution or at step 3.
  - REQ-TRUST-0017: Asserts the class of each implemented step (1 to 6 and the input/options part of 8) for tasks that each fail one step, and that the synthetic codec is not invoked when a step before 8 fails.
  - REQ-TRUST-0018: Every rejection and acceptance runs with an unused recording transport that must report zero calls and no open bodies; admission holds no transport and its import rule forbids network and file I/O.
  - REQ-TRUST-0036: Asserts task_unsupported, not task_schema_invalid, for a well-formed unregistered test_version or test_id and for an unrecognized risk factor type; a malformed test_id or test_version and an unsupported schema_version are task_schema_invalid.
  - REQ-TRUST-0037: Asserts the REQ-TRUST-0037 classes for the implemented rows: step 1 binding failures and remote issued_at 301 and 300.5 seconds ahead as task_source_unauthenticated, unavailable trusted time as task_trust_unavailable, and unregistered pairs and unrecognized risk factor types as task_unsupported.

<a id="UT.internal.admission.TestGateTargetRegistrationAuthority"></a>

### UT.internal.admission.TestGateTargetRegistrationAuthority

- Case: `UT.internal.admission.TestGateTargetRegistrationAuthority`
- Result: passed
- Test: `TestGateTargetRegistrationAuthority` in `internal/admission/target_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateTargetRegistrationAuthority$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateTargetRegistrationAuthority in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0011: Asserts that a registration, including a redirect registration, omitted from the issuer's or the deployment's registration list is target_not_registered, while one omitted only by another issuer still matches.
  - REQ-TRUST-0013: Asserts target_not_registered for a registration of another test or version, of the validation or resolver role instead of measurement, and for a measurement tuple registered only with the redirect role.
  - REQ-TRUST-0014: Asserts that the matched registration's direct or proxy path permission binds, including for a redirect registration, that a redirect not listed by the matched registration is rejected, and that a proxy-only duplicate tuple does not hide a direct registration.

<a id="UT.internal.admission.TestGateTargetRegistrationExactMatch"></a>

### UT.internal.admission.TestGateTargetRegistrationExactMatch

- Case: `UT.internal.admission.TestGateTargetRegistrationExactMatch`
- Result: passed
- Test: `TestGateTargetRegistrationExactMatch` in `internal/admission/target_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateTargetRegistrationExactMatch$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateTargetRegistrationExactMatch in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0013: Asserts that synthetic TCP-01, TLS-01, HTTP-01 and DNS-01 tasks whose every reported tuple exactly matches a current, permitted registration pass step 8, and that the carried registration IDs narrow to the matched measurement and redirect registrations; acceptance (steps 10 and 11) is not tested.
  - REQ-TRUST-0014: Asserts matching on address, port and scheme (TCP-01), plus TLS server name and ALPN values (TLS-01), URL, expected status, body and header checks and a permitted redirect target (HTTP-01), and hostname, resolver name and bootstrap addresses (DNS-01), using test-scoped synthetic codecs and rules, not the public test schemas.

<a id="UT.internal.admission.TestGateTargetRegistrationExpiry"></a>

### UT.internal.admission.TestGateTargetRegistrationExpiry

- Case: `UT.internal.admission.TestGateTargetRegistrationExpiry`
- Result: passed
- Test: `TestGateTargetRegistrationExpiry` in `internal/admission/target_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateTargetRegistrationExpiry$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateTargetRegistrationExpiry in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0013: Asserts that a registration expiring one second after trusted now matches and one expiring at or one second before trusted now is target_not_registered, for each synthetic test, with the deterministic clock.
  - REQ-TRUST-0014: Asserts that an expired redirect registration fails an HTTP-01 task whose measurement registration is current.

<a id="UT.internal.admission.TestGateTargetRegistrationNormalization"></a>

### UT.internal.admission.TestGateTargetRegistrationNormalization

- Case: `UT.internal.admission.TestGateTargetRegistrationNormalization`
- Result: passed
- Test: `TestGateTargetRegistrationNormalization` in `internal/admission/target_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateTargetRegistrationNormalization$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateTargetRegistrationNormalization in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0015: Asserts that forms the synthetic codec normalizes to a registered value (case of names, URL scheme and host, noncanonical IPv6 text) match, while suffix (evil-example.test), subdomain, parent-domain, trailing-dot, wildcard, URL path-prefix and extra-path, URL host-suffix, URL path-case and IPv4-mapped IPv6 forms are target_not_registered; a malformed address is task_schema_invalid at the codec. Public normalization rules for each registered test are not tested.

<a id="UT.internal.admission.TestGateTargetRegistrationPrecedence"></a>

### UT.internal.admission.TestGateTargetRegistrationPrecedence

- Case: `UT.internal.admission.TestGateTargetRegistrationPrecedence`
- Result: passed
- Test: `TestGateTargetRegistrationPrecedence` in `internal/admission/target_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateTargetRegistrationPrecedence$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateTargetRegistrationPrecedence in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0037: Asserts the step-8 target-registration row and first-failing-step precedence around it: an expired task with an unregistered target is task_expired, an unknown field task_schema_invalid, an issuer path refusal task_issuer_not_authorized, and an unregistered target with risk above the issuer maximum target_not_registered.

<a id="UT.internal.admission.TestGateTargetRegistrationRequiresCompleteTuples"></a>

### UT.internal.admission.TestGateTargetRegistrationRequiresCompleteTuples

- Case: `UT.internal.admission.TestGateTargetRegistrationRequiresCompleteTuples`
- Result: passed
- Test: `TestGateTargetRegistrationRequiresCompleteTuples` in `internal/admission/target_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateTargetRegistrationRequiresCompleteTuples$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateTargetRegistrationRequiresCompleteTuples in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0013: Asserts target_not_registered when the test's rule reports no target, only a redirect or only a resolver, parameters of an unexpected type, or more than the local bound of 64 tuples, and that 64 tuples are accepted.
  - REQ-TRUST-0014: Asserts target_not_registered when a synthetic rule reports an incomplete tuple: TCP-01 without address, port or scheme, TLS-01 without server name or ALPN, HTTP-01 without URL or status, DNS-01 without resolver or bootstrap addresses, or a redirect on a non-HTTP test.

<a id="UT.internal.admission.TestGateTargetRegistrationSingleComponentMutations"></a>

### UT.internal.admission.TestGateTargetRegistrationSingleComponentMutations

- Case: `UT.internal.admission.TestGateTargetRegistrationSingleComponentMutations`
- Result: passed
- Test: `TestGateTargetRegistrationSingleComponentMutations` in `internal/admission/target_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateTargetRegistrationSingleComponentMutations$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateTargetRegistrationSingleComponentMutations in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0014: Changes one component at a time (IP address and family, port, scheme, added hostname, server name, each ALPN change including order, URL host, path, port and query, expected status, body and header checks, redirect target, an extra redirect, queried name, resolver name, bootstrap address including order, and direct versus proxy path) and asserts target_not_registered for each.
  - REQ-TRUST-0015: Asserts that matching compares the complete normalized typed tuple, so no single-component variation of a registered tuple is authorized.

<a id="UT.internal.admission.TestGateTaskCannotLowerBaseline"></a>

### UT.internal.admission.TestGateTaskCannotLowerBaseline

- Case: `UT.internal.admission.TestGateTaskCannotLowerBaseline`
- Result: passed
- Test: `TestGateTaskCannotLowerBaseline` in `internal/admission/safety_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateTaskCannotLowerBaseline$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateTaskCannotLowerBaseline in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SAFETY-0010: Asserts that an unknown registered baseline with a low declaration fails closed as task_issuer_not_authorized when the unknown-risk exception is granted on one side only, and is admitted as unknown only with both.
  - REQ-SAFETY-0024: Asserts that a low declaration cannot lower a medium baseline below an issuer or deployment maximum of low, and cannot lower an unknown baseline.
  - REQ-SAFETY-0025: Asserts that a low declaration on a medium baseline is evaluated as medium and rejected as task_issuer_not_authorized under a low maximum on either side.

<a id="UT.internal.admission.TestGateTaskIdentifierRules"></a>

### UT.internal.admission.TestGateTaskIdentifierRules

- Case: `UT.internal.admission.TestGateTaskIdentifierRules`
- Result: passed
- Test: `TestGateTaskIdentifierRules` in `internal/admission/admit_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateTaskIdentifierRules$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateTaskIdentifierRules in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0004: Asserts 128-byte task_id values pass and 129-byte values are task_limit_exceeded, never task_schema_invalid, including 129 uppercase bytes.
  - REQ-TRUST-0019: Asserts task_id values of 1, 127 and 128 allowed bytes pass; empty, uppercase, space, '/', non-ASCII, leading '.', '_' or '-', placeholder syntax and a numeric task_id are task_schema_invalid. Campaign scope and non-user-specificity are not tested.

<a id="UT.internal.admission.TestGateUnreliableSafetyInputsFailClosed"></a>

### UT.internal.admission.TestGateUnreliableSafetyInputsFailClosed

- Case: `UT.internal.admission.TestGateUnreliableSafetyInputsFailClosed`
- Result: passed
- Test: `TestGateUnreliableSafetyInputsFailClosed` in `internal/admission/safety_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateUnreliableSafetyInputsFailClosed$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateUnreliableSafetyInputsFailClosed in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SAFETY-0017: Asserts that unrecognized factor types (battery_drain, bandwidth, a case variant, a hyphen variant and the empty string) are task_unsupported at admission even when both sides grant the unknown-risk exception, never accepted with the factor ignored.
  - REQ-SAFETY-0019: Asserts that an unknown factor under overall low, medium or high is rejected as task_schema_invalid, because no trusted local assessment exists to resolve it. Resolution through a trusted assessment is not implemented or tested.
  - REQ-SAFETY-0029: Asserts that declarations whose risk cannot be evaluated reliably fail closed before effective-risk evaluation.
  - REQ-TRUST-0037: Asserts that an unregistered target at step 8 returns target_not_registered before an unknown-risk rejection at step 9.

<a id="UT.internal.admission.TestRestrictiveFailsClosed"></a>

### UT.internal.admission.TestRestrictiveFailsClosed

- Case: `UT.internal.admission.TestRestrictiveFailsClosed`
- Result: passed
- Test: `TestRestrictiveFailsClosed` in `internal/admission/risk_internal_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestRestrictiveFailsClosed$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestRestrictiveFailsClosed in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SAFETY-0029: Asserts that the effective-risk combination treats unknown and out-of-range levels, including the zero value, as unknown; that unrecognized declared strings map to unknown; and that a matched registration missing from the policy snapshot is evaluated as unknown and rejected without the unknown-risk exception.

<a id="UT.internal.admission.TestTargetRegistrationRejectsCredentialsAndInterpolation"></a>

### UT.internal.admission.TestTargetRegistrationRejectsCredentialsAndInterpolation

- Case: `UT.internal.admission.TestTargetRegistrationRejectsCredentialsAndInterpolation`
- Result: passed
- Test: `TestTargetRegistrationRejectsCredentialsAndInterpolation` in `internal/admission/target_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestTargetRegistrationRejectsCredentialsAndInterpolation$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestTargetRegistrationRejectsCredentialsAndInterpolation in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0015: Asserts that local policy construction refuses registrations with URL credentials or user info, interpolation syntax in URL, host or ALPN, wildcard or suffix-rule hosts and IPv4-mapped addresses, so no task can match them.

<a id="UT.internal.bundle.TestBundleValidationPrecedence"></a>

### UT.internal.bundle.TestBundleValidationPrecedence

- Case: `UT.internal.bundle.TestBundleValidationPrecedence`
- Result: passed
- Test: `TestBundleValidationPrecedence` in `internal/bundle/validation_test.go` (package `internal/bundle`)
- Reproduce one case: `go test -count=1 -run '^TestBundleValidationPrecedence$' ./internal/bundle`
- Scope: Existing synthetic Go unit test TestBundleValidationPrecedence in package internal/bundle, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-BUNDLE-0002: The 'invalid member rejects all' subtest asserts that a bundle containing a zero (unvalidated) result.Value is rejected with ErrValue and no partial bundle escapes.
  - REQ-BUNDLE-0022: The 'invalid member rejects all' subtest asserts that a zero result.Value alongside a valid result causes ErrValue, with no partial bundle.
  - REQ-LIMIT-0021: Asserts a 17-result bundle is rejected with ErrBounds.

<a id="UT.internal.bundle.TestContractMembershipAndIdentity"></a>

### UT.internal.bundle.TestContractMembershipAndIdentity

- Case: `UT.internal.bundle.TestContractMembershipAndIdentity`
- Result: passed
- Test: `TestContractMembershipAndIdentity` in `internal/bundle/bundle_test.go` (package `internal/bundle`)
- Reproduce one case: `go test -count=1 -run '^TestContractMembershipAndIdentity$' ./internal/bundle`
- Scope: Existing synthetic Go unit test TestContractMembershipAndIdentity in package internal/bundle, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-BUNDLE-0001: Builds 64 single-result bundles and asserts no bundle_id repeats, and asserts each ID strictly decodes to 16 non-zero bytes.
  - REQ-BUNDLE-0002: Asserts that Build rejects a results slice containing a zero result.Value together with a valid one.
  - REQ-BUNDLE-0003: Asserts that an unlinked bundle's wire encoding has exactly three top-level members, with schema_version "0.1.0".
  - REQ-BUNDLE-0004: Asserts that New rejects versions "", "0.1", and "0.2.0"; Build rejects 0 and 17 results and accepts 1 and 16; the output has exactly three top-level members with schema_version "0.1.0"; and bundle_id strictly decodes to 16 non-zero bytes.
  - REQ-BUNDLE-0006: Asserts that 64 consecutive Builds produce distinct bundle_id values.
  - REQ-BUNDLE-0017: Asserts that bundle_id strictly base64url-decodes to 16 bytes, is not all zeroes, and re-encodes identically.
  - REQ-BUNDLE-0021: Asserts that Build rejects 0 and 17 results and accepts 1 and 16 with membership preserved.
  - REQ-BUNDLE-0022: Asserts that Build rejects a results slice containing a zero result.Value.
  - REQ-BUNDLE-0029: Asserts that an unlinked bundle's wire encoding has exactly three top-level members (schema_version, bundle_id, results), so no extra bundle-level fields are present.

<a id="UT.internal.bundle.TestExactEncodedBoundaries"></a>

### UT.internal.bundle.TestExactEncodedBoundaries

- Case: `UT.internal.bundle.TestExactEncodedBoundaries`
- Result: passed
- Test: `TestExactEncodedBoundaries` in `internal/bundle/bundle_test.go` (package `internal/bundle`)
- Reproduce one case: `go test -count=1 -run '^TestExactEncodedBoundaries$' ./internal/bundle`
- Scope: Existing synthetic Go unit test TestExactEncodedBoundaries in package internal/bundle, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-BUNDLE-0004: Asserts that bundles of exactly 262,143 and 262,144 encoded bytes are accepted and that 262,145 bytes is rejected with ErrBounds, with and without linkage.
  - REQ-BUNDLE-0021: With and without linkage, asserts that encoded bundles of 262,143 and 262,144 bytes are produced exactly and that 262,145 bytes is rejected with ErrBounds.
  - REQ-LIMIT-0021: Asserts bundles encoding to 262143 and 262144 bytes are built and a 262145-byte bundle is rejected with ErrBounds, with and without linkage.
  - REQ-LIMIT-0022: Asserts a 262145-byte bundle is rejected with ErrBounds and exposes no partial bundle, while accepted bundles have exactly the target size and valid JSON.

<a id="UT.internal.bundle.TestImmutableRetryBytesAndViews"></a>

### UT.internal.bundle.TestImmutableRetryBytesAndViews

- Case: `UT.internal.bundle.TestImmutableRetryBytesAndViews`
- Result: passed
- Test: `TestImmutableRetryBytesAndViews` in `internal/bundle/bundle_test.go` (package `internal/bundle`)
- Reproduce one case: `go test -count=1 -run '^TestImmutableRetryBytesAndViews$' ./internal/bundle`
- Scope: Existing synthetic Go unit test TestImmutableRetryBytesAndViews in package internal/bundle, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-BUNDLE-0007: Asserts that MarshalJSON bytes stay identical, including under concurrent reads, after the source linkage, result slice, View, and returned buffer are mutated, and that rebuilding with changed or removed linkage yields a different ID.
  - REQ-BUNDLE-0013: Asserts that the bundled result's execution.state remains not_run after the input slice and view are mutated.
  - REQ-BUNDLE-0016: Asserts that rebuilding with a linkage whose epoch and token changed, or with nil linkage, yields a bundle ID different from the original.

<a id="UT.internal.bundle.TestMixedResultMembershipIsPreserved"></a>

### UT.internal.bundle.TestMixedResultMembershipIsPreserved

- Case: `UT.internal.bundle.TestMixedResultMembershipIsPreserved`
- Result: passed
- Test: `TestMixedResultMembershipIsPreserved` in `internal/bundle/bundle_test.go` (package `internal/bundle`)
- Reproduce one case: `go test -count=1 -run '^TestMixedResultMembershipIsPreserved$' ./internal/bundle`
- Scope: Existing synthetic Go unit test TestMixedResultMembershipIsPreserved in package internal/bundle, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-BUNDLE-0002: Bundles negative-completed, incomplete, and not-run results produced by result.New/Build and asserts that each results[i] is byte-identical to that result's MarshalJSON output.
  - REQ-BUNDLE-0013: Bundles a completed timeout result, an incomplete cancelled_by_host result, and a not_run result, then asserts that each bundled entry equals its original encoding and that view mutation does not alter the stored outcome.
  - REQ-BUNDLE-0014: Bundles two results that each carry a decoded network_context and asserts that every bundled entry is byte-identical to the result's own encoding, so the contexts are not merged or deduplicated.

<a id="UT.internal.bundle.TestTrustedOptionalLinkage"></a>

### UT.internal.bundle.TestTrustedOptionalLinkage

- Case: `UT.internal.bundle.TestTrustedOptionalLinkage`
- Result: passed
- Test: `TestTrustedOptionalLinkage` in `internal/bundle/bundle_test.go` (package `internal/bundle`)
- Reproduce one case: `go test -count=1 -run '^TestTrustedOptionalLinkage$' ./internal/bundle`
- Scope: Existing synthetic Go unit test TestTrustedOptionalLinkage in package internal/bundle, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-BUNDLE-0003: Asserts that bundles built with nil linkage on builders with linkage disabled and enabled contain no 'linkage' bytes.
  - REQ-BUNDLE-0004: Asserts that linkage is absent unless supplied and, when present, has mode, scope, epoch, and token fields.
  - REQ-BUNDLE-0006: Asserts that json.Unmarshal of wire input containing a bundle_id into bundle.Value fails, so serialized input cannot supply the identifier.
  - REQ-BUNDLE-0010: For epochs 0, 2948, 2^53+1, and MaxInt64, asserts that the encoded linkage has mode rotating_installation, the configured scope, the given epoch, and a token that strictly decodes to the supplied 32 bytes, and that epoch -1 is rejected.
  - REQ-BUNDLE-0011: Asserts that json.Unmarshal into bundle.Value fails for linkage inputs with mode 'stable' plus user_id, and with previous_token, and that the produced linkage mode is rotating_installation.
  - REQ-BUNDLE-0018: For a builder initialized from configuration '{}', asserts that the bundle contains no 'linkage' bytes and that Build with a linkage projection is rejected.
  - REQ-BUNDLE-0023: Asserts that Build with a linkage projection is rejected, with no partial bundle, for a builder with linkage disabled and for one configured for rotating_installation but lacking LinkageKeys, and that nil linkage yields no 'linkage' bytes.
  - REQ-BUNDLE-0025: Asserts that linkage scopes 'other', empty, and 129 'a' characters are rejected; that the emitted scope equals the configured 'org.example.production'; and that serialized wire input cannot be unmarshalled into a bundle Value.
  - REQ-BUNDLE-0026: Asserts that the emitted linkage.token strictly base64url-decodes to bytes equal to the supplied 32-byte token.
  - REQ-LINKAGE-0006: Asserts that a builder without local linkage configuration, or without a key store, rejects supplied linkage; that a mismatched scope is rejected; and that wire JSON cannot be unmarshalled into a bundle Value.

<a id="UT.internal.config.TestAtomicRejectionCorpus"></a>

### UT.internal.config.TestAtomicRejectionCorpus

- Case: `UT.internal.config.TestAtomicRejectionCorpus`
- Result: passed
- Test: `TestAtomicRejectionCorpus` in `internal/config/config_test.go` (package `internal/config`)
- Reproduce one case: `go test -count=1 -run '^TestAtomicRejectionCorpus$' ./internal/config`
- Scope: Existing synthetic Go unit test TestAtomicRejectionCorpus in package internal/config, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0004: Rejects host configuration with a Proxy key, an unknown top-level field, and other invalid sections, publishing no snapshot.
  - REQ-JSON-0005: Rejects null proxy, network_context and result_linkage sections, an empty proxy object, and a quoted port.
  - REQ-JSON-0006: Rejects configuration proxy ports 1.0, 1e0 and -0.
  - REQ-JSON-0010: Asserts each malformed or invalid configuration returns an error and a zero, invalid Snapshot.

<a id="UT.internal.config.TestClosedSchemasAndTrustSeparation"></a>

### UT.internal.config.TestClosedSchemasAndTrustSeparation

- Case: `UT.internal.config.TestClosedSchemasAndTrustSeparation`
- Result: passed
- Test: `TestClosedSchemasAndTrustSeparation` in `internal/config/config_test.go` (package `internal/config`)
- Reproduce one case: `go test -count=1 -run '^TestClosedSchemasAndTrustSeparation$' ./internal/config`
- Scope: Existing synthetic Go unit test TestClosedSchemasAndTrustSeparation in package internal/config, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0004: Rejects configuration containing trust or version field names at top level and inside proxy, network_context and result_linkage sections.

<a id="UT.internal.config.TestConfigurationValidationPrecedence"></a>

### UT.internal.config.TestConfigurationValidationPrecedence

- Case: `UT.internal.config.TestConfigurationValidationPrecedence`
- Result: passed
- Test: `TestConfigurationValidationPrecedence` in `internal/config/validation_test.go` (package `internal/config`)
- Reproduce one case: `go test -count=1 -run '^TestConfigurationValidationPrecedence$' ./internal/config`
- Scope: Existing synthetic Go unit test TestConfigurationValidationPrecedence in package internal/config, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0010: Asserts oversize, wrong-shape and unapproved configurations return no valid snapshot with a stable error class.
  - REQ-JSON-0011: Asserts the same expected error class, at most 128 bytes, is returned over 20 runs for each invalid configuration case.

<a id="UT.internal.config.TestVersionsApprovalsAndBounds"></a>

### UT.internal.config.TestVersionsApprovalsAndBounds

- Case: `UT.internal.config.TestVersionsApprovalsAndBounds`
- Result: passed
- Test: `TestVersionsApprovalsAndBounds` in `internal/config/config_test.go` (package `internal/config`)
- Reproduce one case: `go test -count=1 -run '^TestVersionsApprovalsAndBounds$' ./internal/config`
- Scope: Existing synthetic Go unit test TestVersionsApprovalsAndBounds in package internal/config, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0008: Asserts config.NewDecoder rejects contracts empty, latest, 0.1, 0.2.0 and 0.1.0 with trailing space with ErrVersion and returns no decoder.
  - REQ-JSON-0009: Asserts configuration input at MaxBytes is accepted and one byte over is rejected with ErrBounds.

<a id="UT.internal.execution.TestAllTerminationReasons"></a>

### UT.internal.execution.TestAllTerminationReasons

- Case: `UT.internal.execution.TestAllTerminationReasons`
- Result: passed
- Test: `TestAllTerminationReasons` in `internal/execution/attempt_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestAllTerminationReasons$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestAllTerminationReasons in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-EXECUTION-0010: Stops pending and started attempts with each proxy_* execution reason (proxy_not_configured, proxy_not_permitted, proxy_path_unsupported, proxy_unavailable, proxy_connect_failed, proxy_reconfigured) and asserts the resulting not_run/incomplete phase retains that explicit reason.
  - REQ-RESULT-0020: Stops pending and running attempts with each of the 15 registered reasons and asserts not_run/incomplete phase with the reason retained.
  - REQ-RESULT-0021: Asserts a second Stop after the first termination returns ErrTransition and leaves the first reason unchanged.
  - REQ-SCHEDULE-0027: Asserts only platform_suspended, network_context_unavailable, proxy_unavailable and internal_error yield a retry candidate, while cancelled_by_host, resource_limit, constraints_unsatisfied, local_policy_rejected, proxy_reconfigured and other reasons do not.

<a id="UT.internal.execution.TestProtocolOutcomesAndInvalidReasons"></a>

### UT.internal.execution.TestProtocolOutcomesAndInvalidReasons

- Case: `UT.internal.execution.TestProtocolOutcomesAndInvalidReasons`
- Result: passed
- Test: `TestProtocolOutcomesAndInvalidReasons` in `internal/execution/attempt_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestProtocolOutcomesAndInvalidReasons$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestProtocolOutcomesAndInvalidReasons in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0001: Asserts protocol-outcome strings (nxdomain, tcp_timeout, connection_refused, tls_validation_failed, unexpected_status, etc.) are rejected as execution reasons and that a started attempt still Completes with no reason and no retry candidate.
  - REQ-RESULT-0020: Asserts Stop rejects unregistered strings including protocol outcomes, an empty string, and 'internal_error ' with ErrReason.
  - REQ-RESULT-0021: Asserts protocol-outcome strings cannot be used as stop reasons and a running attempt still completes.
  - REQ-SCHEDULE-0027: Asserts measurement outcome strings are rejected as reasons and completed attempts never become retryable.

<a id="UT.internal.jsondecode.TestBounds"></a>

### UT.internal.jsondecode.TestBounds

- Case: `UT.internal.jsondecode.TestBounds`
- Result: passed
- Test: `TestBounds` in `internal/jsondecode/decode_test.go` (package `internal/jsondecode`)
- Reproduce one case: `go test -count=1 -run '^TestBounds$' ./internal/jsondecode`
- Scope: Existing synthetic Go unit test TestBounds in package internal/jsondecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0009: Asserts 64-byte input is accepted and 65-byte input rejected with ErrBounds even when it is invalid syntax, nested object MaxBytes counts original whitespace, and MaxCount, depth and node limits reject one-over values.
  - REQ-JSON-0010: Asserts depth 1024 and 128 unclosed arrays are rejected with ErrBounds, and over-limit byte, count, depth and node inputs fail.

<a id="UT.internal.jsondecode.TestConfigurationAndNoFallback"></a>

### UT.internal.jsondecode.TestConfigurationAndNoFallback

- Case: `UT.internal.jsondecode.TestConfigurationAndNoFallback`
- Result: passed
- Test: `TestConfigurationAndNoFallback` in `internal/jsondecode/decode_test.go` (package `internal/jsondecode`)
- Reproduce one case: `go test -count=1 -run '^TestConfigurationAndNoFallback$' ./internal/jsondecode`
- Scope: Existing synthetic Go unit test TestConfigurationAndNoFallback in package internal/jsondecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0008: Asserts a 0.1.0 input missing a field required only by the 0.1.0 schema fails with ErrRequired instead of matching a registered 0.2.0 schema.

<a id="UT.internal.jsondecode.TestErrorsAndDeterminism"></a>

### UT.internal.jsondecode.TestErrorsAndDeterminism

- Case: `UT.internal.jsondecode.TestErrorsAndDeterminism`
- Result: passed
- Test: `TestErrorsAndDeterminism` in `internal/jsondecode/decode_test.go` (package `internal/jsondecode`)
- Reproduce one case: `go test -count=1 -run '^TestErrorsAndDeterminism$' ./internal/jsondecode`
- Scope: Existing synthetic Go unit test TestErrorsAndDeterminism in package internal/jsondecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0011: Asserts errors for inputs containing a synthetic-password name and SYNTHETIC-TOKEN value are at most 32 bytes, contain neither string, and are identical over 50 runs, and required-field errors are stable.

<a id="UT.internal.jsondecode.TestPositiveAndPresence"></a>

### UT.internal.jsondecode.TestPositiveAndPresence

- Case: `UT.internal.jsondecode.TestPositiveAndPresence`
- Result: passed
- Test: `TestPositiveAndPresence` in `internal/jsondecode/decode_test.go` (package `internal/jsondecode`)
- Reproduce one case: `go test -count=1 -run '^TestPositiveAndPresence$' ./internal/jsondecode`
- Scope: Existing synthetic Go unit test TestPositiveAndPresence in package internal/jsondecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0001: Accepts an object surrounded by space, tab, carriage return and line feed.
  - REQ-JSON-0002: Accepts label names é and e plus combining acute as distinct keys, showing no Unicode normalization is applied.
  - REQ-JSON-0003: Accepts canonically equivalent but distinct label keys é and e plus combining acute in the same object.
  - REQ-JSON-0005: Asserts a missing optional member is absent from Lookup, explicit null is retained as Null, an empty string is retained as empty, and zero is retained as 0.
  - REQ-JSON-0007: Asserts 9007199254740993, int64 extremes and a 100-digit value within its declared bound decode with their exact token text, and values just outside int64 bounds are rejected with ErrRange.

<a id="UT.internal.jsondecode.TestRejectionCorpus"></a>

### UT.internal.jsondecode.TestRejectionCorpus

- Case: `UT.internal.jsondecode.TestRejectionCorpus`
- Result: passed
- Test: `TestRejectionCorpus` in `internal/jsondecode/decode_test.go` (package `internal/jsondecode`)
- Reproduce one case: `go test -count=1 -run '^TestRejectionCorpus$' ./internal/jsondecode`
- Scope: Existing synthetic Go unit test TestRejectionCorpus in package internal/jsondecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0001: Rejects empty input, trailing value, trailing content, comment, trailing comma, byte-order mark, truncated input, invalid number syntax, and top-level array or scalar against an object schema.
  - REQ-JSON-0002: Rejects invalid UTF-8, a malformed \u escape, unpaired high and low surrogate escapes, and unpaired surrogates in a member name and a nested label value with ErrText.
  - REQ-JSON-0003: Rejects duplicate count members with equal and conflicting values, count versus co\u0075nt, duplicate label keys, and duplicate name/n\u0061me in an array element object with ErrDuplicate.
  - REQ-JSON-0004: Rejects an unknown top-level field, a case-variant Version field, and an unknown field inside an array element object with ErrField.
  - REQ-JSON-0005: Rejects top-level null, a missing required count, count null, and count given as a string, boolean, array or object.
  - REQ-JSON-0006: Rejects integer field values -0, 1.0, 1e0, 1E+0 and 1.5 with ErrInteger and +1, 01, 1., 1e and --1 with ErrSyntax.
  - REQ-JSON-0007: Rejects count values -1 and 18446744073709551616 outside the declared 0..18446744073709551615 range with ErrRange.
  - REQ-JSON-0008: Asserts DecodeSelected rejects inputs with missing version fields, null, integer, empty or 9.0 versions, and an unregistered test_id with ErrVersion.
  - REQ-LABEL-0010: The 'duplicate-label' case asserts that the JSON parser rejects {"labels":{"x":"a","x":"a"}} with the duplicate error class.

<a id="UT.internal.jsondecode.TestValueContainers"></a>

### UT.internal.jsondecode.TestValueContainers

- Case: `UT.internal.jsondecode.TestValueContainers`
- Result: passed
- Test: `TestValueContainers` in `internal/jsondecode/decode_test.go` (package `internal/jsondecode`)
- Reproduce one case: `go test -count=1 -run '^TestValueContainers$' ./internal/jsondecode`
- Scope: Existing synthetic Go unit test TestValueContainers in package internal/jsondecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0002: Asserts a valid surrogate-pair escape decodes to the single scalar U+1F600 and a literal U+FFFD is preserved.

<a id="UT.internal.model.observation.TestAcceptedProducerMappings"></a>

### UT.internal.model.observation.TestAcceptedProducerMappings

- Case: `UT.internal.model.observation.TestAcceptedProducerMappings`
- Result: passed
- Test: `TestAcceptedProducerMappings` in `internal/model/observation/dns_tls_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestAcceptedProducerMappings$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestAcceptedProducerMappings in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0044: Asserts ValidateALPNOffer with contract 0.2.0 is rejected.
  - REQ-OBS-0045: Asserts DNSResponseCode, NegotiatedTLSVersion, NegotiatedALPN and PeerAlertClass return no value when exposed is false while mapping exposed out-of-subset values to OTHER/unknown/other.
  - REQ-OBS-0046: Asserts DNSResponseCode maps 0..5 to NOERROR..REFUSED and every other uint16 to OTHER when exposed, and returns no value for exposed=false.
  - REQ-OBS-0047: Asserts single-response observations with FORMERR, NOTIMP or OTHER are accepted only with failure inconclusive (SERVFAIL/NXDOMAIN/REFUSED with their specific failures) and are rejected with no_data.
  - REQ-OBS-0048: Asserts NegotiatedTLSVersion returns TLS1.2 for 0x0303, TLS1.3 for 0x0304, unknown for 0, 0x0301, 0x0305 and 65535, and no value when exposed is false.
  - REQ-OBS-0049: Asserts NegotiatedALPN returns http/1.1, h2, dot unchanged and unknown for h3/private, no value when not exposed or empty, and ValidateALPNOffer accepts the three identifiers but rejects unknown, h2c, h3, H2, private, duplicate h2 and nine entries.
  - REQ-OBS-0050: Asserts PeerAlertClass maps all 256 codes to the specified classes with other as default, and returns no value for exposed=false.

<a id="UT.internal.model.observation.TestAccessOnlySnapshot"></a>

### UT.internal.model.observation.TestAccessOnlySnapshot

- Case: `UT.internal.model.observation.TestAccessOnlySnapshot`
- Result: passed
- Test: `TestAccessOnlySnapshot` in `internal/model/observation/context_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestAccessOnlySnapshot$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestAccessOnlySnapshot in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-CONTEXT-FORMAT-0004: Asserts the all-unknown six-field access object decodes and re-serializes byte-identically, and that a null network_changed, an added device_id field, or trailing data is rejected.
  - REQ-CONTEXT-FORMAT-0005: Asserts DecodeAccess rejects an access object with an added device_id field.
  - REQ-OBS-0006: Asserts the access-only object decodes via DecodeAccess, serializes byte-identically, binds to its generation, and can be projected from a full NetworkContext.

<a id="UT.internal.model.observation.TestContextNormalizationStaysWithinASCIICodes"></a>

### UT.internal.model.observation.TestContextNormalizationStaysWithinASCIICodes

- Case: `UT.internal.model.observation.TestContextNormalizationStaysWithinASCIICodes`
- Result: passed
- Test: `TestContextNormalizationStaysWithinASCIICodes` in `internal/model/observation/validation_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestContextNormalizationStaysWithinASCIICodes$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestContextNormalizationStaysWithinASCIICodes in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-CONTEXT-FORMAT-0006: Asserts US/us/uS with matching region decode to US and US-SC and round-trip idempotently, while codes containing U+017F or leading whitespace are rejected with ErrValue.
  - REQ-CONTEXT-FORMAT-0018: Asserts us/us-sc and uS/Us-sC decode to US/US-SC, normalization is idempotent on round trip, and U+017F or whitespace-containing codes are rejected.
  - REQ-CONTEXT-FORMAT-0023: Asserts contexts rejected for non-ASCII codes cannot be marshalled.
  - REQ-JSON-0002: Asserts country/region codes containing U+017F or leading whitespace are rejected rather than case-folded, while ASCII mixed case is normalized.

<a id="UT.internal.model.observation.TestContextPathGenerationAndPrivacy"></a>

### UT.internal.model.observation.TestContextPathGenerationAndPrivacy

- Case: `UT.internal.model.observation.TestContextPathGenerationAndPrivacy`
- Result: passed
- Test: `TestContextPathGenerationAndPrivacy` in `internal/model/observation/context_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestContextPathGenerationAndPrivacy$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestContextPathGenerationAndPrivacy in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-CONTEXT-0017: Asserts a decoded direct context fails Matches for another network generation and for a proxy path, a proxy context fails Matches for a new proxy generation, a changed-network context fails Matches, and direct-path input is rejected when decoded against a proxy Path.
  - REQ-CONTEXT-FORMAT-0001: Asserts injected ip, network_generation, proxy_generation, device_id, ssid and context_id fields are rejected at top level, in access and in egress, that serialized context contains none of them, and that a Path cannot be JSON-serialized.
  - REQ-CONTEXT-FORMAT-0005: Asserts context decoding rejects ip, ssid, resolver_address, operator, provider and device_id fields injected into the access object.
  - REQ-CONTEXT-FORMAT-0016: Asserts context decoding rejects ip, endpoint_url, proxy_host, network_generation, proxy_generation, provider, operator and context_id at every object level, that serialized context omits them, and that json.Marshal of a Path fails.
  - REQ-CONTEXT-FORMAT-0017: Asserts unknown fields such as ip, ssid and context_id are rejected at top level, in access and in egress.

<a id="UT.internal.model.observation.TestContextRejectionAndNormalization"></a>

### UT.internal.model.observation.TestContextRejectionAndNormalization

- Case: `UT.internal.model.observation.TestContextRejectionAndNormalization`
- Result: passed
- Test: `TestContextRejectionAndNormalization` in `internal/model/observation/context_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestContextRejectionAndNormalization$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestContextRejectionAndNormalization in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-CONTEXT-FORMAT-0004: Asserts context decoding rejects a missing network_changed, a null network_changed, and an unregistered network_type value "private".
  - REQ-CONTEXT-FORMAT-0006: Asserts rejection of an unknown lookup_source, absent or unapproved classification_version, unapproved country or region, region inconsistent with country, connection_code on direct, and ASN values 0, 4294967296 and 1.0.
  - REQ-CONTEXT-FORMAT-0007: Asserts an unavailable egress object that still carries country, region, ASN and classification_version is rejected.
  - REQ-CONTEXT-FORMAT-0008: Asserts fresh with lookup_source cache and cached with lookup_source primary are rejected.
  - REQ-CONTEXT-FORMAT-0017: Asserts rejection of network_type "private" and lookup_source "unknown" enum values plus duplicate keys and oversize input.
  - REQ-CONTEXT-FORMAT-0018: Asserts a context with nl/nl-NH codes decodes to CountryCode NL and RegionCode NL-NH.
  - REQ-CONTEXT-FORMAT-0019: Asserts rejection of a region_code NL-NH with country_code US and of a region_code NL-NH with country_code removed.
  - REQ-CONTEXT-FORMAT-0020: Asserts a direct network_path egress object with connection_code reviewed_proxy is rejected.
  - REQ-CONTEXT-FORMAT-0022: Asserts the case "unavailable classification" (status unavailable, lookup_source none, with country/region/ASN/classification_version present) is rejected.
  - REQ-CONTEXT-FORMAT-0023: Asserts, via assertContextRejected for each invalid case, that the returned context fails to serialize, and that a zero NetworkContext cannot be marshalled.

<a id="UT.internal.model.observation.TestContextUnknownAndFreshnessStates"></a>

### UT.internal.model.observation.TestContextUnknownAndFreshnessStates

- Case: `UT.internal.model.observation.TestContextUnknownAndFreshnessStates`
- Result: passed
- Test: `TestContextUnknownAndFreshnessStates` in `internal/model/observation/context_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestContextUnknownAndFreshnessStates$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestContextUnknownAndFreshnessStates in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-CONTEXT-FORMAT-0006: Asserts fresh/primary, cached/cache, fresh/fallback without classification, and unavailable/none egress objects on a direct path decode and round-trip.
  - REQ-CONTEXT-FORMAT-0007: Asserts an egress object with status unavailable, lookup_source none, network_path direct and no classification fields decodes and round-trips.
  - REQ-CONTEXT-FORMAT-0008: Asserts fresh/primary, fresh/fallback and cached/cache egress combinations decode successfully.
  - REQ-CONTEXT-FORMAT-0009: Asserts a fresh/primary egress with only classification_version decodes with empty CountryCode and nil ASN, and that fresh/fallback without classification fields is accepted.

<a id="UT.internal.model.observation.TestDNSBoundsAndPrivacy"></a>

### UT.internal.model.observation.TestDNSBoundsAndPrivacy

- Case: `UT.internal.model.observation.TestDNSBoundsAndPrivacy`
- Result: passed
- Test: `TestDNSBoundsAndPrivacy` in `internal/model/observation/dns_tls_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestDNSBoundsAndPrivacy$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestDNSBoundsAndPrivacy in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0009: Asserts udp/ipv4/A and system/AAAA observations round-trip and rejects an unknown engine, null responses, response_limit_reached false, and response_limit_reached true before the bound.
  - REQ-OBS-0012: Asserts rejection of TXT type, an AAAA answer for an A query, Alias/café/trailing-dot/64-char-label CNAMEs, and answers 10.0.0.1, 127.0.0.1, 192.0.2.1, 001.1.1.1, ::ffff:1.1.1.1, fe80::1%en0 and other non-public addresses.
  - REQ-OBS-0013: Asserts an observation with two identical A answers in one response and an answer carrying ttl are rejected.
  - REQ-OBS-0014: Asserts a DNS observation padded to exactly 16384 bytes is accepted and one byte more is rejected, answers are rejected when RetainDNSAnswers is false, and resolver_address/raw_message fields are rejected.
  - REQ-OBS-0040: Asserts a DNS observation with an injected resolver_address field is rejected.
  - REQ-OBS-0042: Asserts rejection of arrival_index 2, 0 and 1.0, arrival_offset_ms 750 beyond duration and 251, missing answers, and rcode values noerror, BADVERS and 16.
  - REQ-OBS-0045: Asserts rcode noerror (incorrect case), BADVERS (raw label) and numeric 16 are rejected.
  - REQ-OBS-0047: Asserts a succeeded observation whose only response has rcode OTHER is rejected.

<a id="UT.internal.model.observation.TestDNSCollectionAndAnswerBoundaries"></a>

### UT.internal.model.observation.TestDNSCollectionAndAnswerBoundaries

- Case: `UT.internal.model.observation.TestDNSCollectionAndAnswerBoundaries`
- Result: passed
- Test: `TestDNSCollectionAndAnswerBoundaries` in `internal/model/observation/dns_tls_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestDNSCollectionAndAnswerBoundaries$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestDNSCollectionAndAnswerBoundaries in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0009: Asserts three UDP responses with response_limit_reached are accepted, multiple responses for tcp engine are rejected, and a configured MaxDNSResponses of 2 is enforced.
  - REQ-OBS-0013: Asserts NewDNS removes an appended exact duplicate answer leaving two answers with the original CNAME first.
  - REQ-OBS-0014: Asserts NewDNS accepts 16 answers and rejects 17 answer inputs in one response.
  - REQ-OBS-0042: Asserts nine CNAMEs are rejected while eight are accepted, and decreasing arrival offsets across responses are rejected.

<a id="UT.internal.model.observation.TestHTTPBoundsAndConsistency"></a>

### UT.internal.model.observation.TestHTTPBoundsAndConsistency

- Case: `UT.internal.model.observation.TestHTTPBoundsAndConsistency`
- Result: passed
- Test: `TestHTTPBoundsAndConsistency` in `internal/model/observation/observation_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestHTTPBoundsAndConsistency$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestHTTPBoundsAndConsistency in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0025: Asserts rejection of protocol h3, status_code 99 and 600, empty checks, null body and an unregistered redirect terminal class, and acceptance of a redirect with followed_count 2.
  - REQ-OBS-0028: Asserts rejection when truncated is missing, bytes_processed 513 exceeds a 512-byte configured bound, truncated is true with 511 bytes, and when the limit itself exceeds 65536.
  - REQ-OBS-0030: Asserts rejection of status check index 1, index -1, body check index 8, a duplicate status check, and a body check when BodyChecks is configured as 0.
  - REQ-OBS-0037: Asserts an http observation with failure raw_error is rejected.
  - REQ-OBS-0038: Asserts a succeeded http observation with an unmatched check and a failed body_check_failed without an unmatched body check are rejected, while body_check_failed with an unmatched body check is accepted.

<a id="UT.internal.model.observation.TestIncompleteFailureAbsentForEveryType"></a>

### UT.internal.model.observation.TestIncompleteFailureAbsentForEveryType

- Case: `UT.internal.model.observation.TestIncompleteFailureAbsentForEveryType`
- Result: passed
- Test: `TestIncompleteFailureAbsentForEveryType` in `internal/model/observation/dns_tls_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestIncompleteFailureAbsentForEveryType$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestIncompleteFailureAbsentForEveryType in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0051: Asserts incomplete tcp, http, dns and tls observations are accepted without failure and rejected with failure null, "", timeout or inconclusive.

<a id="UT.internal.model.observation.TestProtocolPresenceAndVocabulary"></a>

### UT.internal.model.observation.TestProtocolPresenceAndVocabulary

- Case: `UT.internal.model.observation.TestProtocolPresenceAndVocabulary`
- Result: passed
- Test: `TestProtocolPresenceAndVocabulary` in `internal/model/observation/observation_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestProtocolPresenceAndVocabulary$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestProtocolPresenceAndVocabulary in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0003: Asserts valid tcp/http/incomplete/failed observations round-trip and rejects missing/negative/251/float/null duration_ms, unknown status, failed without failure, failure on succeeded or incomplete, and DNS/TLS type with TCP details.
  - REQ-OBS-0036: Asserts a tcp observation with status succeeded and failure timeout is rejected.
  - REQ-OBS-0037: Asserts a failed tcp observation without failure and one with failure network_changed are rejected, and failed/inconclusive is accepted.
  - REQ-OBS-0044: Asserts DecodeProtocol with contract 0.2.0 returns ErrVersion.
  - REQ-OBS-0051: Asserts an incomplete tcp observation with failure timeout is rejected.

<a id="UT.internal.model.observation.TestProtocolPrivacyAndImmutableViews"></a>

### UT.internal.model.observation.TestProtocolPrivacyAndImmutableViews

- Case: `UT.internal.model.observation.TestProtocolPrivacyAndImmutableViews`
- Result: passed
- Test: `TestProtocolPrivacyAndImmutableViews` in `internal/model/observation/observation_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestProtocolPrivacyAndImmutableViews$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestProtocolPrivacyAndImmutableViews in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0004: Asserts tcp and http observations with injected unknown common or detail fields (raw_packet, headers, destination, etc.) are rejected.
  - REQ-OBS-0018: Asserts tcp observations with an injected destination field (at top level) are rejected by the decoder.
  - REQ-OBS-0031: Asserts http observations with injected headers, body_bytes or body_digest detail fields are rejected and that detached Check/Body/HTTPDetails values cannot be JSON-serialized.
  - REQ-OBS-0040: Asserts tcp and http observations with injected resolver_address, destination or proxy_host fields are rejected.

<a id="UT.internal.model.observation.TestProtocolValidationPrecedenceAndStableEncoding"></a>

### UT.internal.model.observation.TestProtocolValidationPrecedenceAndStableEncoding

- Case: `UT.internal.model.observation.TestProtocolValidationPrecedenceAndStableEncoding`
- Result: passed
- Test: `TestProtocolValidationPrecedenceAndStableEncoding` in `internal/model/observation/validation_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestProtocolValidationPrecedenceAndStableEncoding$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestProtocolValidationPrecedenceAndStableEncoding in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0008: Asserts DecodeProtocol with contract 9.0.0 returns ErrVersion for each protocol observation type.
  - REQ-JSON-0011: Asserts an observation containing a private DO_NOT_RETAIN field returns the same ErrField or ErrVersion class, at most 128 bytes, over 20 runs.
  - REQ-OBS-0004: Asserts an injected unknown common field "private" is rejected with jsondecode.ErrField for all four observation types.
  - REQ-OBS-0044: Asserts contract 9.0.0 returns ErrVersion before field validation for all four observation types.

<a id="UT.internal.model.observation.TestTLSVocabularyAndConsistency"></a>

### UT.internal.model.observation.TestTLSVocabularyAndConsistency

- Case: `UT.internal.model.observation.TestTLSVocabularyAndConsistency`
- Result: passed
- Test: `TestTLSVocabularyAndConsistency` in `internal/model/observation/dns_tls_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestTLSVocabularyAndConsistency$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestTLSVocabularyAndConsistency in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0038: Asserts succeeded TLS with certificate_validation not_completed and alert on a succeeded handshake are rejected, and certificate_untrusted with trusted validation is rejected.
  - REQ-OBS-0043: Asserts TLS1.2/TLS1.3/unknown/absent versions are accepted, TLS1.0, 771 and tls1.3 versions, alpn private and empty, alert on succeeded and an extra certificate field are rejected, and each alert class is accepted only with failure alert.
  - REQ-OBS-0045: Asserts TLS1.0, numeric 771, lowercase tls1.3, alpn private and numeric alert 42 are rejected rather than repaired.
  - REQ-OBS-0049: Asserts a TLS observation with alpn private or empty string is rejected.
  - REQ-OBS-0050: Asserts each alert class plus absent alert is accepted with failed/alert, alert on a succeeded handshake and numeric alert 42 are rejected, and certificate_error with trusted local validation is accepted.

<a id="UT.internal.model.task.TestAdmissionErrorVocabulary"></a>

### UT.internal.model.task.TestAdmissionErrorVocabulary

- Case: `UT.internal.model.task.TestAdmissionErrorVocabulary`
- Result: passed
- Test: `TestAdmissionErrorVocabulary` in `internal/model/task/admission_error_test.go` (package `internal/model/task`)
- Reproduce one case: `go test -count=1 -run '^TestAdmissionErrorVocabulary$' ./internal/model/task`
- Scope: Existing synthetic Go unit test TestAdmissionErrorVocabulary in package internal/model/task, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0034: Asserts each of the ten codes maps to its registered string, is not an execution reason, and that all other uint8 values outside the eleven defined codes are rejected.
  - REQ-TRUST-0035: Asserts admission errors refuse JSON marshal/unmarshal and that a canary containing secret, credentials, participant, target and raw platform text never appears in formatted output.
  - REQ-TRUST-0036: Asserts the Unsupported code maps to the fixed string task_unsupported, is not an execution reason, carries no payload under the same encoding and canary checks, and that values 12 to 255 and 0 are rejected.

<a id="UT.internal.model.task.TestMalformedSourceRejectedAtomically"></a>

### UT.internal.model.task.TestMalformedSourceRejectedAtomically

- Case: `UT.internal.model.task.TestMalformedSourceRejectedAtomically`
- Result: passed
- Test: `TestMalformedSourceRejectedAtomically` in `internal/model/task/trust_test.go` (package `internal/model/task`)
- Reproduce one case: `go test -count=1 -run '^TestMalformedSourceRejectedAtomically$' ./internal/model/task`
- Scope: Existing synthetic Go unit test TestMalformedSourceRejectedAtomically in package internal/model/task, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0001: Asserts NewSourceContext returns ErrTrustModel and an invalid context for a remote source with Authenticated false, nil or zero IssuedAt, invalid origin, zero policy generation, or invalid issuer.

<a id="UT.internal.result.TestBindingBoundsAndProducerContext"></a>

### UT.internal.result.TestBindingBoundsAndProducerContext

- Case: `UT.internal.result.TestBindingBoundsAndProducerContext`
- Result: passed
- Test: `TestBindingBoundsAndProducerContext` in `internal/result/payload_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestBindingBoundsAndProducerContext$' ./internal/result`
- Scope: Existing synthetic Go unit test TestBindingBoundsAndProducerContext in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0011: Asserts New rejects a binding whose Test.Version or Test.ResultVersion alone is 9.0.0.
  - REQ-RESULT-0038: Asserts New rejects a binding with unsupported Test.Version or Test.ResultVersion.

<a id="UT.internal.result.TestCompositeCompletionAndPartialBoundaries"></a>

### UT.internal.result.TestCompositeCompletionAndPartialBoundaries

- Case: `UT.internal.result.TestCompositeCompletionAndPartialBoundaries`
- Result: passed
- Test: `TestCompositeCompletionAndPartialBoundaries` in `internal/result/payload_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestCompositeCompletionAndPartialBoundaries$' ./internal/result`
- Scope: Existing synthetic Go unit test TestCompositeCompletionAndPartialBoundaries in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-05-0040: Accepts a NET-05 result with profile, network_validation and two ordered step measurements, including a negative primitive outcome, and rejects the steps in reverse order.
  - REQ-NET-05-0041: Accepts partial results with stopped_before_step naming the next profile step, rejects missing or wrong stopped_before_step, and rejects a profile_version different from the bound profile.
  - REQ-NET-05-0044: Rejects a step whose test result_version is 9.0.0, rejects an incomplete step that includes a measurement, and accepts completed steps with primitive measurements.
  - REQ-NET-05-0045: Accepts incomplete results with zero, one completed, one incomplete, or completed-then-incomplete steps and matching stopped_before_step, and rejects an incomplete first step followed by a completed step.
  - REQ-NET-05-0046: Accepts completed with all steps completed under completed state, rejects completed for partial step sets, accepts network_not_validated with no steps only when the profile requires validation, and rejects steps run under not_validated validation when required.
  - REQ-RESULT-0021: Accepts NET-05 incomplete results with partial step measurements and rejects inconsistent partial structures.
  - REQ-RESULT-0027: Accepts NET-05 measurements carrying the named profile, network_validation, steps, and stopped_before_step fields.
  - REQ-RESULT-0029: Accepts NET-05 incomplete results retaining zero, one completed, or a trailing incomplete step with stopped_before_step, and rejects inconsistent partial structures.

<a id="UT.internal.result.TestCompositeVersionPrecisionAndMarkerPresence"></a>

### UT.internal.result.TestCompositeVersionPrecisionAndMarkerPresence

- Case: `UT.internal.result.TestCompositeVersionPrecisionAndMarkerPresence`
- Result: passed
- Test: `TestCompositeVersionPrecisionAndMarkerPresence` in `internal/result/payload_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestCompositeVersionPrecisionAndMarkerPresence$' ./internal/result`
- Scope: Existing synthetic Go unit test TestCompositeVersionPrecisionAndMarkerPresence in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0007: Asserts a NET-05 profile_version of 9007199254740993 round-trips without rounding.
  - REQ-NET-05-0041: Rejects a completed NET-05 result containing an empty stopped_before_step.

<a id="UT.internal.result.TestEncodedAndObservationBoundaries"></a>

### UT.internal.result.TestEncodedAndObservationBoundaries

- Case: `UT.internal.result.TestEncodedAndObservationBoundaries`
- Result: passed
- Test: `TestEncodedAndObservationBoundaries` in `internal/result/result_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestEncodedAndObservationBoundaries$' ./internal/result`
- Scope: Existing synthetic Go unit test TestEncodedAndObservationBoundaries in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0009: Asserts a 65536-byte result is accepted and a 65537-byte result is rejected with ErrBounds.
  - REQ-LIMIT-0021: Asserts a 65536-byte TCP-01 result is accepted, 65537 bytes is rejected with ErrBounds, and a result with 17 observations is rejected.
  - REQ-RESULT-0004: For TCP-01, accepts a whitespace-padded input of exactly 65,536 bytes, rejects 65,537 bytes with jsondecode.ErrBounds, and rejects a measurement with 17 observations.
  - REQ-RESULT-0026: Rejects a TCP-01 measurement with 17 observations.

<a id="UT.internal.result.TestEnvelopePresenceAndBounds"></a>

### UT.internal.result.TestEnvelopePresenceAndBounds

- Case: `UT.internal.result.TestEnvelopePresenceAndBounds`
- Result: passed
- Test: `TestEnvelopePresenceAndBounds` in `internal/result/result_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestEnvelopePresenceAndBounds$' ./internal/result`
- Scope: Existing synthetic Go unit test TestEnvelopePresenceAndBounds in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0002: Rejects deletion of each required top-level field, mismatched task_id, zero/padded/noncanonical run_id, attempt 0/3/1.5 and non-canonical numeric spellings, invalid platform and library_version, and malformed time/execution values for a TCP-01 binding.
  - REQ-RESULT-0016: Rejects started_at with seconds, fractional seconds, or +00:00 offset, precision other than minute, duration_ms 999, duration_capped false, and duration_capped true below 60,000, and accepts duration_ms 60000 with duration_capped true.
  - REQ-RESULT-0019: Rejects execution.state 'failed' and completed results carrying reason cancelled_by_host or an empty reason.
  - REQ-RESULT-0021: Rejects a completed TCP-01 result with the measurement removed.
  - REQ-RESULT-0023: Rejects a completed TCP-01 result with network_context missing or null.
  - REQ-RESULT-0034: Rejects deletion of each required top-level field, a completed result without measurement or with reason, and missing network_context for a completed TCP-01 result.
  - REQ-RESULT-0036: Rejects an all-zero run_id, a padded run_id, and a noncanonical run_id encoding.
  - REQ-RESULT-0041: Rejects a completed TCP-01 result whose measurement is removed.

<a id="UT.internal.result.TestExecutionContextAndAttemptBinding"></a>

### UT.internal.result.TestExecutionContextAndAttemptBinding

- Case: `UT.internal.result.TestExecutionContextAndAttemptBinding`
- Result: passed
- Test: `TestExecutionContextAndAttemptBinding` in `internal/result/result_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestExecutionContextAndAttemptBinding$' ./internal/result`
- Scope: Existing synthetic Go unit test TestExecutionContextAndAttemptBinding in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0019: Accepts incomplete and not_run results with each registered reason and rejects them when the reason is removed.
  - REQ-RESULT-0020: Decodes incomplete and not_run TCP-01 results with each of the 15 registered reasons and rejects reason 'timeout'.
  - REQ-RESULT-0021: For TCP-01, rejects incomplete and not_run results that include a measurement or omit the reason, and accepts them with a registered reason and no measurement.
  - REQ-RESULT-0023: Rejects an incomplete (network_changed) TCP-01 result whose network_context was deleted, while accepting results whose egress.status is unavailable.
  - REQ-RESULT-0024: Accepts a not_run TCP-01 result without network_context, rejects it when Binding.CapturedContext is true, and rejects an incomplete result without context.
  - REQ-RESULT-0029: Rejects TCP-01 incomplete results that include a measurement.
  - REQ-RESULT-0034: Rejects incomplete/not_run results without reason or with measurement, and not_run without context when context was captured.
  - REQ-RESULT-0042: For incomplete and not_run TCP-01 results, accepts each of 15 registered reasons and rejects a missing reason and reason 'timeout'.

<a id="UT.internal.result.TestImmutableViewsAndProducerBuild"></a>

### UT.internal.result.TestImmutableViewsAndProducerBuild

- Case: `UT.internal.result.TestImmutableViewsAndProducerBuild`
- Result: passed
- Test: `TestImmutableViewsAndProducerBuild` in `internal/result/result_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestImmutableViewsAndProducerBuild$' ./internal/result`
- Scope: Existing synthetic Go unit test TestImmutableViewsAndProducerBuild in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0004: Asserts Build rejects a producer result with 17 observations and a recursive NET-05 step structure before encoding.

<a id="UT.internal.result.TestIndependentVersionsAndNegativeCompletion"></a>

### UT.internal.result.TestIndependentVersionsAndNegativeCompletion

- Case: `UT.internal.result.TestIndependentVersionsAndNegativeCompletion`
- Result: passed
- Test: `TestIndependentVersionsAndNegativeCompletion` in `internal/result/result_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestIndependentVersionsAndNegativeCompletion$' ./internal/result`
- Scope: Existing synthetic Go unit test TestIndependentVersionsAndNegativeCompletion in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0001: Decodes TCP-01 results whose single tcp_connect observation failed (connection_refused, timeout, connection_reset, network_unreachable, address_unreachable, permission_denied, socket_error, inconclusive) as execution.state completed with no reason.
  - REQ-RESULT-0011: Mutating schema_version, test.id, test.version, or test.result_version to an unsupported value each yields jsondecode.ErrVersion and no usable Value.
  - REQ-RESULT-0032: Mutating schema_version, test.id, test.version, or test.result_version to an unsupported value each yields jsondecode.ErrVersion and no usable Value.
  - REQ-RESULT-0038: Mutating schema_version, test.id, test.version, or test.result_version to an unsupported value each yields jsondecode.ErrVersion and no usable Value.

<a id="UT.internal.result.TestInitialPayloadSchemasAndEvidence"></a>

### UT.internal.result.TestInitialPayloadSchemasAndEvidence

- Case: `UT.internal.result.TestInitialPayloadSchemasAndEvidence`
- Result: passed
- Test: `TestInitialPayloadSchemasAndEvidence` in `internal/result/payload_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestInitialPayloadSchemasAndEvidence$' ./internal/result`
- Scope: Existing synthetic Go unit test TestInitialPayloadSchemasAndEvidence in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-DNS-01-0037: Accepts DNS-01 results whose outcome resolved has a succeeded system dns_query and whose outcome nxdomain has a dns_query failed with nxdomain retaining rcode NXDOMAIN, and rejects the same results when the outcome is replaced by the unregistered value censored.
  - REQ-DNS-03-0027: Accepts DNS-03 results with outcome response (tcp, tls, http, succeeded doh dns_query with NOERROR) and tcp_failure (failed tcp_connect), and rejects them when the outcome is replaced by censored.
  - REQ-DNS-04-0027: Accepts DNS-04 results with outcome dns_refused (dot dns_query failed with refused retaining rcode REFUSED after tcp and tls) and tcp_timeout (tcp_connect failed with timeout), and rejects them when the outcome is replaced by censored.
  - REQ-HTTP-01-0036: Accepts HTTP-01 results with outcome expected_response (terminal succeeded http_exchange) and tls_failure (tls_handshake failed with certificate_untrusted), and rejects them when the outcome is replaced by censored.
  - REQ-NET-01-0023: Accepts a NET-01 snapshot result with an access object and empty observations, rejects snapshot without access, rejects a top-level network_context member, and rejects a snapshot with network_changed true.
  - REQ-NET-03-0027: Accepts NET-03 results validated (expected response class, platform unknown) and connection_failure (failed tcp_connect), and rejects them when the outcome is replaced by censored.
  - REQ-RESULT-0024: Asserts a NET-01 result that adds a network_context object is rejected.
  - REQ-RESULT-0026: Accepts registered outcomes with ordered observations for TCP-01, TLS-01, DNS-01, DNS-03, DNS-04, HTTP-01, NET-01, and NET-03, rejects outcome 'censored', and rejects out-of-order or mismatched observation sequences.
  - REQ-RESULT-0027: For each initial test, asserts a measurement containing an extra raw_body field is rejected.
  - REQ-RESULT-0041: Rejects outcome 'censored' for each initial test, an incomplete tcp_connect observation in a completed TCP-01 result, and TLS-01 outcomes inconsistent with observations.
  - REQ-RESULT-0043: Rejects a TCP-01 result carrying a tls_handshake observation and TLS-01 results with wrong observation order or missing handshake.
  - REQ-RESULT-0046: Rejects a measurement with an added raw_body field for each initial test.
  - REQ-TCP-01-0024: Accepts a TCP-01 connected result with one succeeded tcp_connect observation, rejects a TCP-01 result whose observation is tls_handshake, rejects an incomplete tcp_connect, and rejects unknown measurement fields.
  - REQ-TLS-01-0025: Accepts TLS-01 results with tcp_connect followed by tls_handshake and a tcp_refused result with only a failed tcp_connect, and rejects tls_handshake before tcp_connect, handshake_succeeded without a tls_handshake, and a tls_handshake after a failed tcp_connect.

<a id="UT.internal.result.TestProducerTimingAndRandomIdentity"></a>

### UT.internal.result.TestProducerTimingAndRandomIdentity

- Case: `UT.internal.result.TestProducerTimingAndRandomIdentity`
- Result: passed
- Test: `TestProducerTimingAndRandomIdentity` in `internal/result/result_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestProducerTimingAndRandomIdentity$' ./internal/result`
- Scope: Existing synthetic Go unit test TestProducerTimingAndRandomIdentity in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0006: Generates 100 run IDs with NewRunID and asserts each is 22 characters, not all-zero, and unique within the sample.
  - REQ-RESULT-0016: Asserts Timing maps elapsed 0, 1ns, 250ms, 250ms+1ns, 60s, and 60s+1ns to 0, 250, 250, 500, 60000 (uncapped), and 60000 (capped).
  - REQ-RESULT-0018: Asserts Timing rejects start.Round(0) (stripped monotonic reading) and an end before start, and caps elapsed time above 60 s at 60000 with duration_capped.
  - REQ-RESULT-0036: Asserts NewRunID output is 22 characters and never the all-zero encoding over 100 samples.

<a id="UT.internal.result.TestResultValidationPrecedence"></a>

### UT.internal.result.TestResultValidationPrecedence

- Case: `UT.internal.result.TestResultValidationPrecedence`
- Result: passed
- Test: `TestResultValidationPrecedence` in `internal/result/validation_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestResultValidationPrecedence$' ./internal/result`
- Scope: Existing synthetic Go unit test TestResultValidationPrecedence in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0008: Asserts a result with schema_version 9.0.0 or test result_version 9.0.0 is rejected with ErrVersion.
  - REQ-JSON-0011: Asserts each invalid result case, including one with a private DO_NOT_RETAIN field, returns the expected error class, at most 128 bytes, over 30 runs.
  - REQ-RESULT-0032: Asserts an unknown schema_version is rejected with ErrVersion before unknown fields are evaluated and a test.result_version mismatch is rejected with ErrVersion.

<a id="UT.internal.result.TestUninitializedDecoderRejectsBeforeInput"></a>

### UT.internal.result.TestUninitializedDecoderRejectsBeforeInput

- Case: `UT.internal.result.TestUninitializedDecoderRejectsBeforeInput`
- Result: passed
- Test: `TestUninitializedDecoderRejectsBeforeInput` in `internal/result/validation_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestUninitializedDecoderRejectsBeforeInput$' ./internal/result`
- Scope: Existing synthetic Go unit test TestUninitializedDecoderRejectsBeforeInput in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0010: Asserts nil and zero result decoders return ErrBinding for any input and the returned value cannot be serialized.

<a id="UT.internal.result.TestUnknownPrivateAndMalformedFields"></a>

### UT.internal.result.TestUnknownPrivateAndMalformedFields

- Case: `UT.internal.result.TestUnknownPrivateAndMalformedFields`
- Result: passed
- Test: `TestUnknownPrivateAndMalformedFields` in `internal/result/result_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestUnknownPrivateAndMalformedFields$' ./internal/result`
- Scope: Existing synthetic Go unit test TestUnknownPrivateAndMalformedFields in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0003: Adds an unknown private_endpoint field at the root (and keys such as input, options, labels, user_id, raw_error, annotations) and asserts the result is rejected with a sanitized error.
  - REQ-RESULT-0005: Adds top-level input, options, schedule, safety, labels, and input_digest keys to a valid TCP-01 result and asserts each is rejected without echoing the value in the error.
  - REQ-RESULT-0014: Adds a private_endpoint field inside the implementation object, and an annotations key at the root, and asserts both results are rejected.
  - REQ-RESULT-0025: Adds private_endpoint inside network_context.access, network_context.egress, and other objects, and ip_address at the root, and asserts each result is rejected.
  - REQ-RESULT-0027: Asserts an unknown private_endpoint field inside measurement is rejected.
  - REQ-RESULT-0032: Asserts unknown fields at root and nested objects are rejected rather than discarded.
  - REQ-RESULT-0043: Rejects unknown fields added to an observation and to its details object.
  - REQ-RESULT-0044: Rejects unknown fields at every object level, keys such as raw_error, user_id, ip_address, and input_digest, trailing data, duplicate keys, invalid escapes, and a BOM, with errors not containing the injected value.
  - REQ-RESULT-0046: Rejects results carrying raw_error or an injected private_endpoint value in root, measurement, observation, and details objects.

<a id="UT.internal.taskdecode.TestByteAndLabelBounds"></a>

### UT.internal.taskdecode.TestByteAndLabelBounds

- Case: `UT.internal.taskdecode.TestByteAndLabelBounds`
- Result: passed
- Test: `TestByteAndLabelBounds` in `internal/taskdecode/decode_test.go` (package `internal/taskdecode`)
- Reproduce one case: `go test -count=1 -run '^TestByteAndLabelBounds$' ./internal/taskdecode`
- Scope: Existing synthetic Go unit test TestByteAndLabelBounds in package internal/taskdecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0009: Asserts task inputs of 65536 bytes are accepted and 65537 rejected with ErrBounds, and input, options and labels nested-byte limits count original whitespace bytes.
  - REQ-LABEL-0005: Asserts that encoded labels objects of 2047 and 2048 bytes are accepted and 2049 is rejected, and that 17 labels are rejected with ErrBounds while 16 are accepted.

<a id="UT.internal.taskdecode.TestDefaultsAndDeterminism"></a>

### UT.internal.taskdecode.TestDefaultsAndDeterminism

- Case: `UT.internal.taskdecode.TestDefaultsAndDeterminism`
- Result: passed
- Test: `TestDefaultsAndDeterminism` in `internal/taskdecode/decode_test.go` (package `internal/taskdecode`)
- Reproduce one case: `go test -count=1 -run '^TestDefaultsAndDeterminism$' ./internal/taskdecode`
- Scope: Existing synthetic Go unit test TestDefaultsAndDeterminism in package internal/taskdecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-EXECUTION-0013: Asserts a task omitting execution limits decodes with defaults DurationMS 30000, BytesSent 32768, BytesReceived 131072 and Connections 8.
  - REQ-LABEL-0004: Asserts that decoding a task with "labels":{} yields a candidate deeply equal to the candidate decoded with labels omitted.

<a id="UT.internal.taskdecode.TestExactSchemaSelection"></a>

### UT.internal.taskdecode.TestExactSchemaSelection

- Case: `UT.internal.taskdecode.TestExactSchemaSelection`
- Result: passed
- Test: `TestExactSchemaSelection` in `internal/taskdecode/decode_test.go` (package `internal/taskdecode`)
- Reproduce one case: `go test -count=1 -run '^TestExactSchemaSelection$' ./internal/taskdecode`
- Scope: Existing synthetic Go unit test TestExactSchemaSelection in package internal/taskdecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0008: Asserts a 0.2.0 task selects only the 0.2.0 codec, invalid 0.2.0 input calls no codec, and unsupported or case-variant test_id/test_version tuples are rejected with ErrVersion.

<a id="UT.internal.taskdecode.TestHostConfigurationCannotComeFromTask"></a>

### UT.internal.taskdecode.TestHostConfigurationCannotComeFromTask

- Case: `UT.internal.taskdecode.TestHostConfigurationCannotComeFromTask`
- Result: passed
- Test: `TestHostConfigurationCannotComeFromTask` in `internal/taskdecode/config_boundary_test.go` (package `internal/taskdecode`)
- Reproduce one case: `go test -count=1 -run '^TestHostConfigurationCannotComeFromTask$' ./internal/taskdecode`
- Scope: Existing synthetic Go unit test TestHostConfigurationCannotComeFromTask in package internal/taskdecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LABEL-0001: Asserts that a label whose value is host result_linkage JSON is retained as opaque text and does not yield a valid deployment configuration.
  - REQ-LINKAGE-0006: Asserts that a task containing a top-level result_linkage field is rejected, and that a label containing result_linkage JSON does not produce a valid configuration.
  - REQ-TASK-0001: Asserts tasks with top-level proxy, network_context, result_linkage, local_policy, policy_generation, authenticated, clock or linkage_keys members are rejected with no candidate.

<a id="UT.internal.taskdecode.TestModelValidationPrecedenceAndExactCodec"></a>

### UT.internal.taskdecode.TestModelValidationPrecedenceAndExactCodec

- Case: `UT.internal.taskdecode.TestModelValidationPrecedenceAndExactCodec`
- Result: passed
- Test: `TestModelValidationPrecedenceAndExactCodec` in `internal/taskdecode/validation_test.go` (package `internal/taskdecode`)
- Reproduce one case: `go test -count=1 -run '^TestModelValidationPrecedenceAndExactCodec$' ./internal/taskdecode`
- Scope: Existing synthetic Go unit test TestModelValidationPrecedenceAndExactCodec in package internal/taskdecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0008: Asserts an unsupported test_version is rejected with ErrVersion and a 0.2.0 task with 0.1.0-shaped input fails with ErrField without invoking the 0.1.0 codec.
  - REQ-JSON-0010: Asserts invalid task inputs return a nil candidate and do not invoke the test codec before earlier validation succeeds.
  - REQ-JSON-0011: Asserts each invalid task case returns the expected error class, at most 128 bytes, over 20 runs.

<a id="UT.internal.taskdecode.TestRegisteredNumericalBoundaries"></a>

### UT.internal.taskdecode.TestRegisteredNumericalBoundaries

- Case: `UT.internal.taskdecode.TestRegisteredNumericalBoundaries`
- Result: passed
- Test: `TestRegisteredNumericalBoundaries` in `internal/taskdecode/boundary_test.go` (package `internal/taskdecode`)
- Reproduce one case: `go test -count=1 -run '^TestRegisteredNumericalBoundaries$' ./internal/taskdecode`
- Scope: Existing synthetic Go unit test TestRegisteredNumericalBoundaries in package internal/taskdecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-EXECUTION-0013: Asserts the lowest and highest registered values of max_duration_ms, max_bytes_sent, max_bytes_received and max_connections are accepted.
  - REQ-LABEL-0005: Asserts via Decode that 64-byte label keys are accepted and 65-byte keys rejected, and that 128-byte UTF-8 values (64 'я') are accepted and 130-byte values rejected.

<a id="UT.internal.taskdecode.TestRejectionCorpus"></a>

### UT.internal.taskdecode.TestRejectionCorpus

- Case: `UT.internal.taskdecode.TestRejectionCorpus`
- Result: passed
- Test: `TestRejectionCorpus` in `internal/taskdecode/decode_test.go` (package `internal/taskdecode`)
- Reproduce one case: `go test -count=1 -run '^TestRejectionCorpus$' ./internal/taskdecode`
- Scope: Existing synthetic Go unit test TestRejectionCorpus in package internal/taskdecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-EXECUTION-0013: Asserts max_duration_ms 60001, max_bytes_sent 65537, max_bytes_received 262145 and max_connections 13 are rejected with ErrRange and no candidate.
  - REQ-LABEL-0014: The 'label-control' case asserts that a decoded label value 'a\nb' is rejected with validation.ErrTask.
  - REQ-TASK-0001: Asserts a task with an unknown top-level extra member is rejected with ErrField and no candidate.

<a id="UT.internal.taskdecode.TestRepresentationAndSanitizedErrors"></a>

### UT.internal.taskdecode.TestRepresentationAndSanitizedErrors

- Case: `UT.internal.taskdecode.TestRepresentationAndSanitizedErrors`
- Result: passed
- Test: `TestRepresentationAndSanitizedErrors` in `internal/taskdecode/decode_test.go` (package `internal/taskdecode`)
- Reproduce one case: `go test -count=1 -run '^TestRepresentationAndSanitizedErrors$' ./internal/taskdecode`
- Scope: Existing synthetic Go unit test TestRepresentationAndSanitizedErrors in package internal/taskdecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0003: Rejects task inputs with task_id and task_\u0069d, and with repeated subject or enabled members, returning no candidate.
  - REQ-JSON-0006: Rejects task max_connections values -0, 1.0 and 1e0 with no candidate.
  - REQ-JSON-0011: Asserts a failing test codec's error is surfaced as ErrParameters without the codec's sensitive text.

<a id="UT.internal.taskdecode.TestRiskFactorAndTimezoneDomains"></a>

### UT.internal.taskdecode.TestRiskFactorAndTimezoneDomains

- Case: `UT.internal.taskdecode.TestRiskFactorAndTimezoneDomains`
- Result: passed
- Test: `TestRiskFactorAndTimezoneDomains` in `internal/taskdecode/domain_test.go` (package `internal/taskdecode`)
- Reproduce one case: `go test -count=1 -run '^TestRiskFactorAndTimezoneDomains$' ./internal/taskdecode`
- Scope: Existing synthetic Go unit test TestRiskFactorAndTimezoneDomains in package internal/taskdecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SAFETY-0012: Decodes synthetic SYN-01 tasks and asserts that `medium`, `high` and `unknown` with `risk_factors` omitted return ErrTask and a nil candidate, while `low` with it omitted decodes.
  - REQ-SAFETY-0013: Asserts that one `network_observability` factor at the task's level decodes for each risk level and is preserved in the candidate.
  - REQ-SAFETY-0031: Asserts that `medium`, `high` and `unknown` with `risk_factors: []` return ErrTask and a nil candidate on every repetition, that one factor decodes, and that `low` with `[]` decodes with an empty factor list.
  - REQ-SCHEDULE-0013: Asserts that a `time_window` with `timezone` omitted decodes with timezone `local`, and with `local` decodes unchanged.
  - REQ-SCHEDULE-0014: Asserts that `local` is the decoded and accepted timezone value; device-zone evaluation is not exercised.
  - REQ-SCHEDULE-0043: Asserts that `timezone` values `UTC`, `Europe/Berlin`, `""`, `Local`, `LOCAL` and `local ` return ErrTask and a nil candidate.

<a id="UT.internal.taskdecode.TestSerializedInputCannotSupplyTrustOrPolicy"></a>

### UT.internal.taskdecode.TestSerializedInputCannotSupplyTrustOrPolicy

- Case: `UT.internal.taskdecode.TestSerializedInputCannotSupplyTrustOrPolicy`
- Result: passed
- Test: `TestSerializedInputCannotSupplyTrustOrPolicy` in `internal/taskdecode/trust_boundary_test.go` (package `internal/taskdecode`)
- Reproduce one case: `go test -count=1 -run '^TestSerializedInputCannotSupplyTrustOrPolicy$' ./internal/taskdecode`
- Scope: Existing synthetic Go unit test TestSerializedInputCannotSupplyTrustOrPolicy in package internal/taskdecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0004: Rejects tasks carrying trust/policy field names at top level and under input, options, safety, execution and schedule, rejects the same names in configuration, and keeps a label value resembling a source record as opaque text.
  - REQ-LABEL-0001: Asserts that a label whose value is a serialized trust/policy assertion is retained as unchanged opaque text.
  - REQ-TASK-0001: Asserts tasks with top-level trust/policy-named members such as source_context, origin and issuer_id are rejected with no candidate.

<a id="UT.internal.taskdecode.TestStagedDecoding"></a>

### UT.internal.taskdecode.TestStagedDecoding

- Case: `UT.internal.taskdecode.TestStagedDecoding`
- Result: passed
- Test: `TestStagedDecoding` in `internal/taskdecode/staged_test.go` (package `internal/taskdecode`)
- Reproduce one case: `go test -count=1 -run '^TestStagedDecoding$' ./internal/taskdecode`
- Scope: Existing synthetic Go unit test TestStagedDecoding in package internal/taskdecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0008: Asserts the staged decoder selects the exact (test_id, test_version) after shape validation, returns ErrUnregistered (also ErrVersion) only for a well-formed unregistered pair and ErrVersion for a malformed pair or unsupported schema_version, and validates input against the selected test schema only in Parameters, without invoking the codec on failure.

<a id="UT.internal.testkit.conformance.TestConformanceFixtures"></a>

### UT.internal.testkit.conformance.TestConformanceFixtures

- Case: `UT.internal.testkit.conformance.TestConformanceFixtures`
- Result: passed
- Test: `TestConformanceFixtures` in `internal/testkit/conformance/runner_test.go` (package `internal/testkit/conformance`)
- Reproduce one case: `go test -count=1 -run '^TestConformanceFixtures$' ./internal/testkit/conformance`
- Scope: Existing synthetic Go unit test TestConformanceFixtures in package internal/testkit/conformance, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SAFETY-0012: Fixtures `risk-factors-*-one` and `risk-factors-low-empty` decode through adapter `common-task-v1` to the expected normalized projection; this projection does not include safety fields.
  - REQ-SAFETY-0013: Fixtures `risk-factors-medium-one`, `risk-factors-high-one` and `risk-factors-unknown-one`, each with one valid factor, decode to the expected projection.
  - REQ-SAFETY-0031: Fixtures `risk-factors-medium-empty`, `risk-factors-high-empty` and `risk-factors-unknown-empty` reject with exact class `task` and a nil candidate, repeatedly and concurrently; `risk-factors-low-empty` is accepted.
  - REQ-SCHEDULE-0013: Fixtures `timezone-omitted` and `timezone-local` decode to the expected projection.
  - REQ-SCHEDULE-0014: Fixture `timezone-local` is accepted; device-zone evaluation is not exercised.
  - REQ-SCHEDULE-0043: Fixtures `timezone-utc`, `timezone-iana`, `timezone-empty` and `timezone-local-case` reject with exact class `task` and a nil candidate.

<a id="UT.internal.validation.TestCommonLabelInvariants"></a>

### UT.internal.validation.TestCommonLabelInvariants

- Case: `UT.internal.validation.TestCommonLabelInvariants`
- Result: passed
- Test: `TestCommonLabelInvariants` in `internal/validation/task_test.go` (package `internal/validation`)
- Reproduce one case: `go test -count=1 -run '^TestCommonLabelInvariants$' ./internal/validation`
- Scope: Existing synthetic Go unit test TestCommonLabelInvariants in package internal/validation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LABEL-0004: The 'omitted' (nil) and 'empty' ([]) subtests assert that Common accepts both and returns nil Labels.
  - REQ-LABEL-0005: Asserts acceptance of a 64-byte key with a 128-byte UTF-8 value and of 16 labels, and rejection of a 65-byte key, a 130-byte value, and 17 labels.
  - REQ-LABEL-0010: The 'duplicate' subtest asserts that Common rejects two labels with key 'campaign' and returns a zero candidate.
  - REQ-LABEL-0014: Asserts that values containing 0xff (invalid UTF-8) and '\n' are rejected, and that 64x'я' and '${literal}' are accepted.

<a id="UT.internal.validation.TestCommonNormalizationIsIdempotentAndConservative"></a>

### UT.internal.validation.TestCommonNormalizationIsIdempotentAndConservative

- Case: `UT.internal.validation.TestCommonNormalizationIsIdempotentAndConservative`
- Result: passed
- Test: `TestCommonNormalizationIsIdempotentAndConservative` in `internal/validation/task_test.go` (package `internal/validation`)
- Reproduce one case: `go test -count=1 -run '^TestCommonNormalizationIsIdempotentAndConservative$' ./internal/validation`
- Scope: Existing synthetic Go unit test TestCommonNormalizationIsIdempotentAndConservative in package internal/validation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-EXECUTION-0013: Asserts validation.Common rejects attempt limits with DurationMS 60001 or BytesSent 0 with ErrTask and returns a zero candidate instead of a clamped one.
  - REQ-LABEL-0001: Asserts that Common keeps label values like '${literal}' verbatim, only sorts labels, leaves the rest of the candidate equal to the expected normalized form, and does not alias the input.
  - REQ-LABEL-0014: Asserts that the label values '${literal}' and 'Я' are preserved verbatim through normalization.

<a id="UT.internal.validation.TestCommonRiskFactorAndTimezoneDomains"></a>

### UT.internal.validation.TestCommonRiskFactorAndTimezoneDomains

- Case: `UT.internal.validation.TestCommonRiskFactorAndTimezoneDomains`
- Result: passed
- Test: `TestCommonRiskFactorAndTimezoneDomains` in `internal/validation/task_test.go` (package `internal/validation`)
- Reproduce one case: `go test -count=1 -run '^TestCommonRiskFactorAndTimezoneDomains$' ./internal/validation`
- Scope: Existing synthetic Go unit test TestCommonRiskFactorAndTimezoneDomains in package internal/validation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SAFETY-0012: Asserts that validation.Common rejects `medium`, `high` and `unknown` candidates with no risk factor list with ErrTask and a zero candidate, and accepts a `low` candidate with the list omitted.
  - REQ-SAFETY-0013: Asserts that one factor of type `network_observability` at the task's own level is accepted for each of `low`, `medium`, `high` and `unknown`.
  - REQ-SAFETY-0031: Asserts that `medium`, `high` and `unknown` candidates with an empty, non-nil factor list are rejected with ErrTask and a zero candidate, that one factor makes each valid, and that a `low` candidate with an empty list remains valid.
  - REQ-SCHEDULE-0013: Asserts that a time window with timezone `local` validates and keeps `local`.
  - REQ-SCHEDULE-0014: Asserts that `local` is the timezone value accepted for a time window; the device-reported zone itself is not evaluated by this test.
  - REQ-SCHEDULE-0043: Asserts that time-window timezones `UTC`, `Europe/Berlin`, `""`, `Local` and `LOCAL` are rejected with ErrTask and a zero candidate.

<a id="UT.internal.validation.TestRefiningClasses"></a>

### UT.internal.validation.TestRefiningClasses

- Case: `UT.internal.validation.TestRefiningClasses`
- Result: passed
- Test: `TestRefiningClasses` in `internal/validation/limits_test.go` (package `internal/validation`)
- Reproduce one case: `go test -count=1 -run '^TestRefiningClasses$' ./internal/validation`
- Scope: Existing synthetic Go unit test TestRefiningClasses in package internal/validation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0002: Asserts values above registered duration, connection, run, attempt, label and repeat-jitter maxima are rejected matching ErrLimit, values below minima are rejected without ErrLimit, and a not_after beyond seven days matches ErrLimit; nothing is accepted after clamping.
  - REQ-SAFETY-0017: Asserts that validation.Common rejects an unrecognized risk factor type matching ErrUnsupported and ErrTask, and that an otherwise invalid factor is not reported as unsupported.

<a id="UT.internal.validation.TestTaskIDRulesAndLimitClass"></a>

### UT.internal.validation.TestTaskIDRulesAndLimitClass

- Case: `UT.internal.validation.TestTaskIDRulesAndLimitClass`
- Result: passed
- Test: `TestTaskIDRulesAndLimitClass` in `internal/validation/limits_test.go` (package `internal/validation`)
- Reproduce one case: `go test -count=1 -run '^TestTaskIDRulesAndLimitClass$' ./internal/validation`
- Scope: Existing synthetic Go unit test TestTaskIDRulesAndLimitClass in package internal/validation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0004: Asserts 129-byte task_id values match ErrLimit as well as ErrTask, whatever their characters, and that no shorter invalid identifier matches ErrLimit.
  - REQ-TRUST-0019: Asserts validation.TaskID accepts 1 and 128 allowed bytes and rejects empty, uppercase, space, '/', non-ASCII and leading '-' identifiers with ErrTask.
