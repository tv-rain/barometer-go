# Measurement Result Models

**Status:** Internal model and serialization foundation; runtime integration pending  
**Contract:** Draft `0.1.0`  
**Public baseline:** [`barometer-spec` at `a944ed6c389fec9f0d089f0044ead6fe845328b1`](https://github.com/tv-rain/barometer-spec/tree/a944ed6c389fec9f0d089f0044ead6fe845328b1)

## Ownership and Use

`internal/model/result` owns detached typed attempt, execution, timing, measurement, profile and step records. They do not establish admission or validity, and reject direct JSON import/export. `internal/result` owns the explicit, bounded wire boundary. Its `Decoder.Decode` validates serialized results atomically; `Decoder.Build` applies bounded producer preflight and the same complete validation to typed values. Only a successful immutable `Value` can serialize. `Value.View` and encoded bytes are detached copies; changing a returned record cannot change retained output. Protocol observations and context snapshots retain their existing immutable owners.

The controlling documents are [Measurement Result](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md), [Measurement Observations](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md), [Network Context](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/network-context.md), [Registered Limits](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md) and each selected [initial test Output contract](https://github.com/tv-rain/barometer-spec/tree/a944ed6c389fec9f0d089f0044ead6fe845328b1/tests). Examples are fixtures, not alternative schemas.

## Explicit Local Binding

`New` copies a bounded `Binding` and selects an initial exact `(test ID, behavior version, result version)` tuple. It accepts only the nine initial test IDs at behavior/result version `0.1.0`; the common envelope independently selects `schema_version: 0.1.0`. Unsupported combinations fail without interpreting a compatible subset. No runtime registry or runner is installed.

Nil or zero-value decoders return `ErrBinding` before processing input or producer records. Only a decoder returned by successful `New` may construct a result. See [Model Validation and Normalization](model-validation.md) for validation precedence and cross-model acceptance coverage.

The future reviewed containing owner must derive the binding from the accepted task and effective configuration. It supplies the exact task ID, retry and attempt-duration bounds, reviewed producer library version/platform, primitive output parameters and protocol limits. The binding contains no target, URL, host annotation, task input, labels, or raw error. Version text is locally bounded to 64 printable ASCII bytes and must match the reviewed producer exactly; this model does not verify a Git tag or authorize a release. Profile revisions use a positive signed 64-bit integer representation. Such bounds describe this internal foundation, not a new public schema or support commitment.

A separate `ContextDecoder` and `Path` attach imported Network Context to the local path/generation metadata using the existing context owner. Wire values cannot supply generations or review approvals. Completed/incomplete remote-network tests require context; unavailable egress remains an explicit object. A completed result rejects a context marked changed. A `not_run` result never carries context, even after a capture (REQ-RESULT-0053, DEV-561). NET-01 omits the common context and uses the existing access-only snapshot representation, with a separately supplied local access generation.

These are internal structural inputs, not authentication, admission or execution capabilities. Correct task binding, policy authority, producer-release selection, capture/freshness provenance, current generations and the first observed termination reason remain the future containing owner's responsibility. A parsed result cannot prove those runtime facts.

## Common Envelope and Payloads

All fixed objects are closed, including test metadata, implementation, timing, execution and selected payload details. Required fields, null versus absence, exact integer tokens, UTF-8, duplicate names and framing follow the shared strict decoder. Complete input, including whitespace, and normalized output are limited to 65,536 bytes. The parser additionally uses local depth-16 and 32,768-value-node bounds; nested payload passes use depth 14. A non-composite result allows at most 16 observations. No field truncation, unknown-field dropping, task/input copying, digest substitution or arbitrary extension map is provided.

A run ID must be canonical unpadded base64url of 16 nonzero-in-aggregate bytes. `NewRunID` obtains cryptographic randomness for the future scheduled-run owner. A run ID's scope, reuse on retries, fresh identity for repeated runs and exclusion from traffic/logging require runtime integration; decoding cannot prove randomness or historical uniqueness. Attempt count cannot exceed the supplied accepted retry bound.

Time is UTC truncated to a minute; elapsed values are rounded upward to 250 ms, preserving zero. `Timing` requires monotonic-bearing `time.Time` values and rejects negative elapsed time. The local duration retention ceiling is 60,000 ms: an actual overrun emits that lower bound and `duration_capped: true`; false flags and capped values below the ceiling are rejected. Without capping, duration cannot exceed the quantized effective attempt bound. The helper does not enforce an execution deadline or capture platform suspension.

Execution accepts only `completed`, `incomplete` and `not_run`, using the existing closed execution-reason vocabulary. Completed attempts require a terminal measurement and prohibit a reason. Negative DNS/TCP/TLS/HTTP observations remain completed outcomes. Not-run attempts prohibit measurement. Incomplete non-composite attempts omit measurement; NET-05 alone may retain its bounded partial output.

Initial output validation covers:

- TCP-01 exact single TCP observation and matching positive/negative outcome.
- TLS-01 TCP/TLS order and terminal failure/certificate evidence.
- DNS-01 selected engine/query type, bounded DNS responses, normalized answer-set conflict checks and response-count-limit precedence.
- DNS-03/DNS-04 ordered optional bootstrap, TCP, TLS, applicable HTTP and encrypted-DNS stages, with exact terminal stage/outcome combinations. A non-success HTTP response may terminate DoH without a fabricated DNS observation.
- HTTP-01 bounded redirect/reuse-aware protocol order, selected limits and terminal outcome/check evidence; expected responses require every configured indexed check.
- NET-01 empty observations, normalized access-only snapshots and platform outcomes, with no shared egress context.
- NET-03 separate platform state and active response class, HTTP/HTTPS stage selection, no followed redirects, and conditional response evidence.
- NET-05 exact copied profile identity and ordered primitive plan, one outer context, no placeholder or nested result envelopes, no more than eight steps, and at most one final incomplete step without primitive measurement. A remaining step requires the exact next `stopped_before_step`. Primitive negative outcomes do not prevent a completed composite. Required-validation rejection runs no primitives.

The models check retained evidence, not protocol truth. For example, absent DNS answers under the separate reviewed retention gate cannot establish whether the underlying API observed an answer; a successful producer can report resolution without retaining addresses. Raw response contents, configured check values, platform API evidence and unavailable details cannot be reconstructed from minimized observations. Actual acquisition, test-specific input admission, error normalization, traffic behavior, complete measurement conformance and safe runtime generation/termination handling require their own implementation and validation.

## Synthetic Evidence

The following tests are local serialization/model cases, not the broader mixed runtime conformance families in the historical traceability manifests:

| Test | Public obligations exercised |
|---|---|
| `TestIndependentVersionsAndNegativeCompletion` | REQ-RESULT-0001, 0010–0012, 0031–0032, 0038; exact independent versions and completed negative outcomes |
| `TestEnvelopePresenceAndBounds` | REQ-RESULT-0002, 0013–0016, 0019, 0034, 0036–0042; required values, canonical IDs, timing, platform, state and version binding |
| `TestExecutionContextAndAttemptBinding` | REQ-RESULT-0020–0024, 0029, 0037, 0042; complete reasons, context requirements, no protocol error as execution reason, retry bounds |
| `TestUnknownPrivateAndMalformedFields` | REQ-RESULT-0003, 0005, 0012, 0014–0015, 0025, 0027, 0032–0035, 0044–0046; closed objects, private fields, strict framing and sanitized rejection |
| `TestEncodedAndObservationBoundaries` | REQ-RESULT-0004; exact encoded-byte boundary and observation count |
| `TestProducerTimingAndRandomIdentity` | Model/helper portions of REQ-RESULT-0006, 0016–0018, 0036; secure ID construction, monotonic input, submillisecond rounding, zero and cap boundaries |
| `TestInitialPayloadSchemasAndEvidence`, `TestDNSCollectionAndHTTPEvidence`, `TestDoHNonSuccessResponseAndRedirectBounds` | REQ-RESULT-0026–0028, 0041, 0043; initial Output schemas, stage order, negative evidence, DNS collection precedence, HTTP checks and response context |
| `TestCompositeCompletionAndPartialBoundaries`, `TestCompositeVersionPrecisionAndMarkerPresence` | REQ-RESULT-0021, 0029; REQ-NET-05-0040–0047; copied plan identity, ordered steps, partial/final step rules, primitive negatives, required validation, exact revisions beyond floating-point precision and marker presence |
| `TestImmutableViewsAndProducerBuild`, `TestBindingBoundsAndProducerContext` | Atomic validated producer path, immutable snapshots, invalid binding rejection, blocked detached/zero encoding and bounded recursive-input rejection |
| `FuzzResult` | Atomic rejection and normalized decode/encode/producer-roundtrip invariants over all selected initial envelope schemas |

No machine-readable runtime case family is promoted to passed/implemented by these cases. No public Go API, bundle, persistence, delivery, scheduler, platform adapter or measurement runner is introduced. Local tests, race checks and pure-Go mobile compilation cannot establish native lifecycle behavior, field evidence, platform support, or a released conformant library.

## Local Validation Record

On October 15, 2026, the DEV-482 working tree passed `make check` (formatting, host build, unit tests, race detector, static analysis, module integrity and Android/iOS ARM64 pure-Go compilation), `git diff --check`, and the public requirement-reference checker. The checker validates reference structure only; it does not assess coverage or conformance. A final 15-second, two-worker `FuzzResult` run passed after 265,851 executions. These are local synthetic/compiler results, with the runtime and platform limitations stated above; no CI, native binding, device or release result is implied.
