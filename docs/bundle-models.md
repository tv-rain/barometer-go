# Result Bundle and Linkage Models

**Status:** Internal producer and serialization foundation, integrated with the pending store and the linkage source; native protected storage pending

**Contract:** Draft `0.1.0`

**Public baseline:** [`barometer-spec` at `a944ed6c389fec9f0d089f0044ead6fe845328b1`](https://github.com/tv-rain/barometer-spec/tree/a944ed6c389fec9f0d089f0044ead6fe845328b1)

## Ownership and Construction

`internal/model/result` owns detached `Bundle` and `RotatingLinkage` projections. They reject direct JSON import/export. `internal/bundle` owns the bounded producer boundary; only its immutable validated `Value` can serialize. It consumes immutable `internal/result.Value` instances, so each member has already passed the exact common envelope and registered Output validation selected by the result owner. Zero or invalid members reject the whole bundle without truncation, omission or automatic splitting.

The controlling public contracts are [Result Bundle](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md), [Result Linkage and Private State](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md), [Registered Limits](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md), and the [Measurement Result boundary](result-models.md). No public schema or sibling repository changes are required for this foundation.

`New` selects exactly bundle contract `0.1.0`, independently of each member's version selection. It requires a successfully prepared immutable local initialization and copies its effective linkage policy and protected-key availability. It invokes no clock or key operation. `Build` accepts 1–16 validated results and enforces the complete encoded maximum of 262,144 bytes, including linkage, delimiters and identifiers. It computes that size before allocating the complete output and generating an ID. Member byte copies are individually bounded by the existing 65,536-byte result limit. There is no new dependency or public Go API.

## Identity and Retry Bytes

Each successful `Build` obtains a fresh cryptographically random, nonzero-in-aggregate 16-byte bundle ID encoded as canonical unpadded base64url. There is no caller-supplied ID. This identity groups a logical submission, independently of the existing run ID and attempt number inside each member. It is the collector's potential idempotency key, not a separate participant identifier or an implemented ingestion operation.

The value retains exact encoded bytes. `MarshalJSON` returns a detached copy; `View` returns detached domain records. Copying the value preserves its logical identity and exact retry bytes. Editing a view cannot change the bundle. Every new `Build`, including one with equal membership, produces a new logical bundle. A retry must retain the original value; reconstructing one is a new submission. Changes to results, linkage presence, epoch or token cannot reuse an old identity through this API. Bundling preserves each member's execution state, outcome and Network Context without merging or synthesizing them.

This is an output-only producer boundary. Arbitrary JSON import is rejected, including unknown/private fields, caller IDs and alternate linkage modes. It does not restore pending bundles after process restart. A future persistence task must define and review a bounded exact-byte restoration boundary, identity validation and state lifecycle rather than importing JSON directly into this value. No runtime queue, acknowledgements, retry scheduling, durable duplicate suppression or collector is implemented.

Since DEV-563 the [pending store](package-architecture.md#persistence-ports) holds the producer's exact bytes and returns them through `Client.NextBundle` until `Client.AcknowledgeBundle` or expiry; it checks stored bytes against the form it writes and re-reads them with the shared strict parser, but never restores a `Value`, so the restoration boundary above is still not needed. The pending store enforces 32 pending bundles, 4,194,304 pending bytes and the 86,400-second retention maximum. Retry scheduling, durable duplicate suppression and any collector remain out of scope.

## Optional Rotating Linkage

Linkage is absent by default. `Build(..., nil)` omits it even when initialization permits rotating linkage. An explicit projection is accepted only when trusted initialization both enables rotating linkage and has a protected-key capability. Its scope must match the exact reviewed local scope. Missing capability or disabled local policy rejects an attempted projection; the runtime owner must pass nil when linkage is disabled or unavailable.

The closed representation emits only `mode: rotating_installation`, the configured scope, a nonnegative epoch, and an unpadded base64url 32-byte token. Epochs use a local signed-64-bit nonnegative representation, with exact integer encoding beyond floating-point precision. Token size is fixed by the typed array; no nonzero-token rule is invented. There is no stable mode, account/device identifier, arbitrary extension map, prior token, or mapping between epochs. Detached projections cannot serialize independently.

The projection is a trusted internal containing-owner input, analogous to the result owner's local binding. Structural validation cannot prove that token bytes came from HMAC rather than another source, or that an epoch is current. `internal/persistence/linkage` is that owner ([Persistence Ports](package-architecture.md#persistence-ports)). It derives tokens with the specified domain separation through the internal protected-key capability, takes the epoch from the trusted clock, never reuses an epoch below the highest it has used, and omits linkage when any prerequisite fails; its only inputs are the purpose string, the reviewed deployment scope and the epoch, never task, collector or host-account values. The pending store asks it for linkage once per new bundle and keeps the projection while the bundle is open. `Builder.LinkageReserve` reports the largest linkage member for the configured scope, so the pending store can keep room for it. This package itself does not derive tokens, create secrets, call platform capabilities, track emitted epochs, reset private state or offer an in-memory/mobile fallback.

Linkage remains sensitive pseudonymous data. Runtime enforcement must still keep identifiers out of measurement and Network Context traffic and prevent linkage-directed task selection; neither path receives a bundle or linkage value in this build. Native protected storage, reinstall and backup behavior and the host reset path belong to DEV-455, and no production code binds a key store, so no build emits linkage. This model does not establish any of those guarantees or platform support.

## Synthetic Evidence

| Test | Model obligations exercised |
|---|---|
| `TestContractMembershipAndIdentity` | Exact bundle version, valid/invalid initialization, zero builder/member rejection, 0/1/16/17 membership, canonical nonzero random identity and fresh logical IDs |
| `TestTrustedOptionalLinkage` | Default omission, disabled/missing-key gating, exact local scope, nonnegative/exact epoch encoding, fixed token encoding, blocked JSON import and detached serialization |
| `TestImmutableRetryBytesAndViews` | Detached buffers/views, copied caller inputs, concurrent exact-byte retries, fresh identity after linkage changes |
| `TestMixedResultMembershipIsPreserved` | Byte-for-byte membership, negative completed outcomes versus incomplete/not-run execution, retained context and detached nested observations |
| `TestExactEncodedBoundaries` | Validated DNS composite results producing exactly 262,143/262,144/262,145 encoded bundle bytes with and without linkage; atomic over-limit rejection |
| `TestProducedBundleIsStrictJSON` | Produced output, with and without linkage and at 1, 16 and the largest result, parses with the strict shared parser: one complete value, valid UTF-8 and unique decoded names in every object within the bundle byte bound (DEV-566 producer evidence for [GO-SPEC-ISSUE-001](specification-issues.md#go-spec-issue-001-duplicate-json-member-names-outside-labels)) |
| `FuzzBundleProducer` | Membership, invalid-member and optional-linkage boundaries, atomic errors, bounded valid serialization and exact-byte retries |
| `TestDerivation`, `TestDisabledByDefault`, `TestFailuresOmitLinkage`, `TestEpochFloor`, `TestEraseDuringDerivation`, `TestConcurrentDerivation` (`internal/persistence/linkage`) | Independently computed HMAC-SHA-256 vectors for two epochs, the exact derivation message and no other input, scope separation, linkage off unless enabled with a key store, omission on clock, context and key-store failure, the non-decreasing epoch floor across generations, erasure and a clear during derivation |
| `TestBundleLinkage`, `TestLinkedBundleRestart`, `TestLinkageReserve` (`internal/persistence/pending`) | One derivation per new bundle with the projection kept while it is open, omission of unavailable or rejected linkage, exact-byte restart of linked bundles, rejection of malformed stored linkage, and the linkage byte reserve in limits and capacity |
| `TestCcLinkageRotationLimits`, `TestClearRemovesLinkage` (root) | The linkage rows of `CC-LIMIT-CONTEXT-LINKAGE-AND-RETENTION-LIMITS` through `Initialize` and `NextBundle`, and clearing that deletes the key and epoch state, creates no secret and derives a new token afterwards |

These are local synthetic model cases. They do not promote historical mixed runtime conformance families to implemented/passed, establish CI or device evidence, or claim released measurement, privacy-lifecycle or platform support.

## Local Validation Record

On October 16, 2026, the DEV-485 working tree passed focused bundle/model tests, `make check` (formatting, host build, unit tests, race detector, static analysis, module integrity and Android/iOS ARM64 pure-Go compilation), `git diff --check`, and the public requirement-reference checker. The checker validates reference structure only, without coverage/conformance or allocation-history assessment. A 15-second, two-worker `FuzzBundleProducer` run passed after 624,445 executions. No CI, native binding, device or release evidence is implied.
