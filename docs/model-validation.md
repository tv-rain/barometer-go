# Model Validation and Normalization

**Status:** Internal Foundation models; admission and runtime integration pending

**Contract:** Draft `0.1.0`

**Public baseline:** [`barometer-spec` at `a944ed6c389fec9f0d089f0044ead6fe845328b1`](https://github.com/tv-rain/barometer-spec/tree/a944ed6c389fec9f0d089f0044ead6fe845328b1)

## Ownership and Validation Order

Each model owner retains its schema and semantic validation. The shared strict JSON leaf handles framing, Unicode, duplicate names, original encoded bounds, types and exact schema selection. It does not select domain defaults or replace test-specific validators. This work adds no public API, generic validator registry, serialized model or runtime admission boundary.

<!-- requirement-ref: REQ-JSON-0011 -->

The [strict decoding contract](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/json-decoding.md#REQ-JSON-0011) requires deterministic behavior for fixed input and the same schema/policy baseline. It prescribes no universal cross-language diagnostic order. Go returns the first failure in these stages:

| Boundary | Validation order |
|---|---|
| Shared JSON | Trusted descriptor/limits; bounded parsing; exact declared tuple when applicable; selected schema. Object schema checks follow input member order, then sorted missing required fields; arrays follow element order. Parser failures retain traversal order. |
| Task | Shared JSON and exact envelope/test tuple; common field projection; common schedule/execution/safety/labels; explicit trusted-time checks; exactly the selected test codec's semantic normalization. |
| Configuration | Initialized decoder; complete structural schema; proxy; primary/fallback context endpoints; linkage. Each section validates values and separate reviewed approvals. |
| Protocol observation | Exact containing contract; trusted protocol limits; shared selected schema; common status/duration; type-specific semantics. |
| Network Context | Initialized decoder; local path shape; shared schema; access and path consistency; egress status, approvals and classification fields. Access-only snapshots first require the exact contract and a positive explicit generation. |
| Measurement Result | Initialized decoder; shared envelope schema/version; exact test/output tuple; bound task/attempt/producer; run identity; timing; execution; context; payload. Typed production checks finite preflight bounds before passing through the decoder. |
| Result Bundle | Initialized builder; result count; optional linkage against initialization; ordered validated members and complete size; fresh random identity and serialization. |
| Trusted source/policy/revocation | Pure typed constructor invariants and exact references; one fixed trust-model failure class. No JSON import/export. |
| Execution attempt | Transition validity before reason validity; errors preserve the original attempt. Terminal state/reason cannot be replaced. |

Operational errors remain fixed bounded classes without input values, supplied field names, raw codec/parser errors or linkage material. Failed decodes/constructors publish no usable partial model. Failed value replacements and attempt transitions preserve the previous value.

## Conservative Normalization

<!-- requirement-ref: REQ-JSON-0008 -->
<!-- requirement-ref: REQ-LABEL-0004 -->
<!-- requirement-ref: REQ-LABEL-0014 -->
<!-- requirement-ref: REQ-CONTEXT-FORMAT-0018 -->

Task decoding applies only the defaults in [Task Models](task-models.md). `validation.Common` checks common invariants and applies the omitted `max_runs` default. It validates label count, uniqueness, syntax, decoded UTF-8 and value bounds, then sorts a copied slice by key. Empty and omitted labels normalize to nil. Values remain literal and case-sensitive, without Unicode normalization, template expansion or enrichment. The decoder separately checks original encoded bytes, including the 2,048-byte labels object limit. A typed candidate cannot recover its original encoded size; `Common` is not an alternate task-ingress or admission boundary.

Common validation leaves test ID, exact version and typed parameters intact. Selected test codecs retain their own defaults and semantic rules; rejection cannot fall back to another version. Limits reject rather than clamp. Proxy paths, empty transport restrictions, consent, risk and retry requirements are not weakened. Timing and eligibility take explicit trusted time/timezone inputs rather than reading the environment.

Network Context accepts ASCII country/region spellings and normalizes case before exact approved-assignment lookup. Non-ASCII spellings reject before Unicode case conversion can turn them into an approved code, such as `uſ` becoming `US`. Configuration may reduce reviewed linkage to `none` under local disablement, retaining no scope or rotation. Producer DNS deduplication remains bounded, preserves first-occurrence order and never truncates distinct answers.

Round trips use each model's intended boundary. Observations, context and results must retain stable canonical bytes after decode/encode/decode. Configuration has no snapshot serializer; tests reconstruct reviewed input fields from copied getters. Trusted models use typed reconstruction and continue rejecting JSON. Bundles have no import/restoration API: copies retain exact retry bytes, while a new build intentionally creates a new random ID. Candidate tasks retain original received bytes separately for delivery identity; normalized candidates are not a signature/replay representation.

## Synthetic Acceptance Evidence

This table combines existing model cases with DEV-483 regressions. It describes model evidence, not execution, admission, platform or field conformance.

| Model family | Valid, invalid and boundary cases | Round trip, normalization and precedence |
|---|---|---|
| Task, input/options, schedule, execution, safety, labels | `internal/taskdecode/{decode,boundary,schedule}_test.go`; `TestCommonLabelInvariants` | `TestModelValidationPrecedenceAndExactCodec`, `TestCommonNormalizationIsIdempotentAndConservative`, `TestDefaultsAndDeterminism`, `TestTimestampFractionIsExact` |
| Initialization, proxy, context endpoints, linkage configuration | `TestAtomicRejectionCorpus`, `TestVersionsApprovalsAndBounds`, `TestSnapshotOwnershipAndReplacement` | `TestConfigurationProjectionRoundTrip`, `TestConfigurationValidationPrecedence` |
| Source, exact-byte delivery, policy, targets, revocation | `internal/model/task/trust_test.go` constructor, tuple, limit, ownership and JSON-rejection tables | `TestTrustedModelReconstruction`; `TestTrustModelsCannotBeImportedOrExportedAsJSON` preserves existing state on attempted import |
| Admission errors and execution attempts | `TestAdmissionErrorVocabulary`, `TestTransitionMatrix`, `TestAllTerminationReasons`, `TestProtocolOutcomesAndInvalidReasons` | `TestStopValidationPrecedence`, `TestAttemptRejectsJSON`; no wire round trip |
| DNS/TCP/TLS/HTTP observations and access/egress context | `internal/model/observation/{observation,dns_tls,context}_test.go` | `TestProtocolValidationPrecedenceAndStableEncoding`, `TestContextNormalizationStaysWithinASCIICodes`; existing access/context generation and producer-mapping cases |
| Measurement Result, nested measurement/profile/steps | `internal/result/{result,payload}_test.go`, including all nine initial Output schemas, NET-05 partial results, byte bounds and recursive producer rejection | `TestResultValidationPrecedence`, `TestUninitializedDecoderRejectsBeforeInput`, `TestImmutableViewsAndProducerBuild`, `FuzzResult` |
| Result Bundle and linkage projection | `TestContractMembershipAndIdentity`, `TestTrustedOptionalLinkage`, `TestExactEncodedBoundaries` | `TestBundleValidationPrecedence`, `TestImmutableRetryBytesAndViews`, `TestMixedResultMembershipIsPreserved`, `FuzzBundleProducer` |

Validation uses `make check`, `git diff --check`, the public requirement-reference checker against the explicit checkout, and finite focused fuzz runs. Reference checking validates citation structure, not coverage or semantic conformance. Historical traceability manifests retain their original baselines and evidence gaps. Production task schemas, complete admission, orchestration, provenance/freshness enforcement, persistence, native adapters, device validation and release support remain outside this task.

## Local Validation Record

On October 16, 2026, the DEV-483 working tree passed `make check`: formatting, host build, unit tests, race detector, static analysis, module integrity and Android/iOS ARM64 pure-Go compilation. `git diff --check` and the public requirement-reference checker also passed. The checker did not assess allocation history, coverage or semantic conformance. No native binding, device, CI, release or field validation was performed.

All six 15-second fuzz runs with two workers per target passed:

| Target | Executions |
|---|---:|
| `FuzzTaskDecode` | 48,373 |
| `FuzzConfiguration` | 101,404 |
| `FuzzProtocol` | 257,913 |
| `FuzzContext` | 419,762 |
| `FuzzResult` | 147,561 |
| `FuzzBundleProducer` | 355,507 |

These finite runs exercise deterministic normalization/errors, bounded atomic failure and each target's existing model invariants. They are regression evidence, not exhaustive proof.
