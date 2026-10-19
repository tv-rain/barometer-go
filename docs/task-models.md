# Task and Schedule Models

## Authority and Scope

DEV-478 adopts the common Draft `0.1.0` task contracts at immutable public `tv-rain/barometer-spec` revision `05c1364af4f9a604b617fca45f060949b557d265`: [Task](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/data-formats/task.md), its linked input, options, execution, schedule, safety and label contracts, [Registered Initial Limits](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/library/registered-limits.md), and the timing requirements of [Task Trust](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/library/task-trust.md). [Runtime JSON Decoding](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/data-formats/json-decoding.md) remains authoritative for representation. These public documents define behavior; this document describes implementation coverage and its limits.

The packages use the standard library and existing strict decoder. No public Go API, external dependency, runtime measurement registration, scheduler, network access, platform adapter or persistence implementation is introduced.

## Ownership and Call Order

- `internal/model/task` owns candidate values: the envelope, typed test-owned parameters, schedules, constraints, safety declarations and labels. It is a dependency leaf. A candidate is never an admitted task.
- `internal/taskdecode` owns the closed common wire schemas, exact schema/test selection, representation bounds, conversion and documented defaults. It calls pure common validation and explicit timing validation before publishing a candidate or invoking test normalization.
- `internal/validation` owns common domain relationships, expiry checks against trusted caller time, and pure start-envelope predicates. It does not authenticate a source, evaluate local safety policy, acquire device state or execute measurements.

`taskdecode.New` receives immutable, locally available `TestCodec` definitions. Each exact `(test_id, test_version)` supplies object schemas for input/options and one pure, bounded normalizer returning owned typed `task.Parameters`. The normalizer owns test-specific defaults, semantics and deterministic normalization; it must not perform I/O or return a nil or typed-nil parameter value. No schema, latest version or empty-object fallback is selected when a tuple is unsupported. An empty or malformed codec collection is a configuration failure. Codec registration does not establish runner implementation or measurement support.

`Decoder.Decode` receives the original bytes and a separate `validation.TimingContext`. It validates representation, common relationships and timing, then invokes only the selected codec. Errors return no candidate. A codec error becomes a fixed `taskdecode: parameters` error; sensitive test-specific errors are not propagated. Callers must separately retain exact received bytes for later integrity/replay handling. Normalized values and JSON reserialization cannot replace those bytes.

## Defaults and Bounds

Registered complete-task, input, options and labels bounds count original encoded UTF-8 bytes, including whitespace. The complete task is bounded to 65,536 bytes before parsing; input/options objects are capped at 49,152/8,192 bytes even if a test descriptor omits or enlarges those caps. Labels are capped at 16 entries and 2,048 encoded bytes, with decoded key/value lengths of 64/128 bytes. Unknown fields and duplicate decoded names reject atomically through the shared decoder. The finite local parsing envelope is 128 value levels and 32,768 value nodes. It does not establish resource behavior on a mobile device.

Missing execution fields receive the registered attempt defaults: 30,000 ms, 32,768 bytes sent, 131,072 bytes received, and 8 connections. Explicit remote-network limits must be positive and within the public maxima; they are never clamped. NET-01 uses zero byte/connection budgets and rejects explicitly positive network budgets. This common special case does not register or implement NET-01.

Missing network path/VPN policy receive `direct`/`any`. Missing repetition/retry receive one run and one attempt. An interval requires explicit bounded interval, jitter, expiry and `max_runs`. Retry with two attempts requires registered unique reasons and backoff; fixed backoff prohibits a multiplier, exponential backoff requires exactly 2, and omitted retry jitter receives 10,000 ms. Explicit zero retry jitter remains zero. There is no invented concurrency default: omitted mode remains absent. Omitted and empty labels are equivalent; label entries are ordered by key. Unreserved labels stay opaque and are never interpolated.

Required fields, nullability and conditional presence are distinct. Only `starts_at` explicitly permits null. An empty country allowlist rejects; an explicitly empty transport list remains empty rather than being widened. Country codes are checked against the included 249 assigned ISO 3166-1 alpha-2 codes. Country allowlists implicitly require country availability when later execution constraints are evaluated. No context lookup happens during decoding.

A `medium`, `high` or `unknown` risk level requires at least one risk factor ([REQ-SAFETY-0031](https://github.com/tv-rain/barometer-spec/blob/5a37e82865a08922c4821c7318656286933421e5/data-formats/task-safety.md#REQ-SAFETY-0031)): a missing or empty `risk_factors` list rejects the whole task. A `low` task may omit the list or send it empty. Factor types and levels are validated as before; local policy, not the declaration, still decides whether a task may run.

## Scheduling and Retry Boundary

Expiry must be later than acceptance and at most 604,800 seconds after the earlier of trusted issuance and acceptance. Remote context requires issuance time, bounded to 300 seconds into the future. Embedded context prohibits issuance; missing expiry requires an explicit release-reviewed one-time exception covering non-persistence across application updates. The timing context contains no evidence of authentication by itself; an eventual admission boundary must establish its provenance and trust.

UTC timestamps accept RFC 3339 `T`/`Z` separators in either case and zero offsets (`+00:00` or `-00:00`). They normalize to UTC. Calendar validation currently uses Go's ordinary-second RFC 3339 parser; leap-second timestamps are outside this implementation's accepted representation. Decimal fractions are stored exactly rather than rounded to nanoseconds, with trailing zeros removed for deterministic comparison. Leap-second support requires separate implementation and review; this change does not claim full RFC 3339 conformance.

`local` is the only defined time-window timezone in Draft `0.1.0` ([REQ-SCHEDULE-0043](https://github.com/tv-rain/barometer-spec/blob/5a37e82865a08922c4821c7318656286933421e5/data-formats/task-schedule.md#REQ-SCHEDULE-0043)); an omitted value is `local`, and every other string, including `""` and differently cased spellings, rejects. Window starts are inclusive and ends exclusive; midnight-crossing ranges are expanded for overlap detection. Equal endpoints represent an empty window and never grant all-day eligibility. Duplicate ranges reject, including duplicate empty ranges.

`RunEligible` and `RetryEligible` check an already due run or attempt against time, expiry, windows and caller-established support, execution and safety gates. They receive the current device timezone explicitly and never read the process timezone or wall clock. Repetition respects the run cap; a retry belongs to an existing run and does not consume another run. Retry reasons use a distinct execution-reason type and the four registered cross-cutting reasons. DNS/TCP/TLS/HTTP observations, cancellation, policy decisions, proxy reconfiguration and resource limits cannot become retries through an unknown string.

These predicates do not generate due times, random jitter, backoff timers, history or exact scheduling guarantees. A future orchestration owner must generate independent approved jitter, bound the realized retry delay, supply due times, track attempts/runs and obtain fresh trusted execution/safety decisions. Passing an eligibility predicate is not admission or permission to contact a target.

## Synthetic Evidence

The tests live in `internal/taskdecode` and exercise all three new package owners through the consuming decoder. SYN-01 and its two test versions are synthetic typed codec fixtures, never runtime measurement registrations. The NET-01 fixture uses its public empty input/options shape only to check the common zero-traffic budget rule.

| Contract area | Cases |
|---|---|
| Common schema, presence, unknown fields and atomic failure | `TestRejectionCorpus`, `TestRepresentationAndSanitizedErrors` |
| Exact test/version schemas and configuration | `TestExactSchemaSelection`, `TestDescriptorConfiguration` |
| Deterministic defaults and typed test normalization | `TestDefaultsAndDeterminism`, including concurrent reads |
| Original encoded byte/count and numerical boundaries | `TestByteAndLabelBounds`, `TestRegisteredNumericalBoundaries`, `TestNET01ZeroTrafficLimits` |
| Schedule/retry relationships and observation separation | `TestScheduleRelationships` |
| Window overlap, midnight and endpoint semantics | `TestWindowValidation`, `TestStartEligibilityAndRetrySeparation` |
| Trusted timing, embedded exception and expiry boundaries | `TestTrustedTimingBoundary`, `TestTimestampFractionIsExact` |
| Execution and safety declaration values | `TestExecutionAndSafetyValues` |
| Determinism, bounded errors and nil candidates under mutation | `FuzzTaskDecode` |

Run `make check`, `git diff --check` and the public requirement-reference checker with the explicit specification checkout. A bounded fuzz run can use `go test ./internal/taskdecode -run='^$' -fuzz='^FuzzTaskDecode$' -fuzztime=20s -parallel=2`.

## Local Validation — October 14, 2026

The prepared changes passed `make check`: formatting, host compilation, unit tests, race detection, static analysis, module consistency/verification, diff whitespace and pure-Go Android ARM64/iOS ARM64 compilation. The public requirement-reference checker passed against the explicit selected public checkout; allocation history and semantic conformance were not assessed. The coordination repository remained unchanged.

The final 20-second `FuzzTaskDecode` run with two workers passed with 204,149 executions. This finite run checks deterministic candidates/errors and bounded atomic failure; it is not exhaustive proof or mobile/device evidence. No native artifact, device test, CI result, release, or full runtime-conformance claim is established.

## Remaining Runtime Work

Production test-specific schemas/normalizers, actual measurement runners and their input/option/execution compatibility checks remain with their approved implementation tasks. Complete admission must establish source authentication, reviewed embedded provenance, issuer authority, registered targets, stricter local limits, registered risk, consent, non-targeted delivery, replay and durable state before scheduling or execution. The common safety declaration checks neither authorize unknown/high risk nor establish consent. Structural label/task-ID checks cannot prove absence of participant-derived data; reviewed delivery and privacy policy remain required.

This change supplies candidate models and synthetic common-contract evidence. It does not establish full task admission, test conformance, native/device behavior, field validation, release support, or closure of downstream specification-issue evidence gaps. No cross-repository changes or merge ordering are required beyond the already merged strict-decoding baseline.
