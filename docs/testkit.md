# Negative and Boundary-Test Helpers

## Scope and Ownership

`internal/testkit` owns reusable synthetic fixture construction, rejection assertions, a manually advanced clock, and an in-memory HTTP transport. Only Go test files may import it; production packages must never import it. It depends only on the Go standard library and introduces no public API, runtime decoder, domain model, platform adapter, or storage fallback.

These helpers support later model and protocol cases without creating a competing schema or harness. The consumer supplies its own decoder/validator, exact error-class matcher, positive controls, and requirement references. The tests in `internal/testkit/reuse_test.go` demonstrate the same runner against two deliberately synthetic consumers; neither implements a Barometer contract or establishes conformance.

The separate test-only [`internal/testkit/conformance`](conformance-fixtures.md) subpackage owns the versioned fixture wrapper and task-decoding adapter. Unlike the parent helpers, it consumes existing internal decoding and model packages. Production code must not import either package.

The test-only [`internal/testkit/delivery`](admission.md#signed-remote-task-packages) subpackage builds synthetic signed remote task packages for admission tests (DEV-612; it replaced the earlier HMAC verifier). Its Ed25519 keys derive from fixed public test seeds and are not deployment keys; it builds the signed message independently of the admission code, can re-sign altered packages and returns trusted key records for synthetic policies. It imports only task models and the standard library, and production code must not import it.

The test-only `internal/testkit/replay` subpackage owns a bounded, fault-injectable replay store that implements `persistence.ReplayStore` for admission tests ([Atomic Rejection and Replay Persistence](admission.md#atomic-rejection-and-replay-persistence)). It keeps encoded records in memory as its model of durable state and can inject:
- a failure at lookup, at the generation read, before a write, during a write (staged but never durable) and after a durable write (an uncertain commit);
- corrupt records;
- a simulated restart that keeps durable records and discards staged writes.

It purges overdue records using the trusted time it is given and exposes its stored bytes for leakage checks. It is not a persistence implementation or a mobile fallback, and production code must not import it.

The test-only `internal/testkit/output` subpackage (DEV-635) is a bounded, fault-injectable store implementing the `execution.Output` port. It keeps serialized results in memory as its model of durable state, can fail one selected `Put` before writing or after a durable write (an uncertain commit), refuses writes on a cancelled context and reports call counts and copies of stored bytes. It is not a persistence implementation or a mobile fallback; the production store is DEV-563.

The test-only `internal/testkit/ledger` subpackage (DEV-636) is a bounded, fault-injectable `persistence.RunLedger`. It can fail lookups, fail a write before anything is stored, or store durably and report failure (an uncertain commit), and reports stored records and write counts. A test keeps one ledger across clients to model accounting that survives reinitialization. The `internal/testkit/capability` subpackage (DEV-636) provides scripted consent, device-state and time-zone capabilities that count their queries and carry no identifier. `output.Store` gains a reported capacity limit and an injectable capacity failure. None is a production implementation or a mobile fallback.

The test-only `internal/testkit/loopback` subpackage (DEV-635) provides an echo server on an ephemeral `127.0.0.1` port, an attempt-scoped `measurement.Dialer` over `net.Dialer` that counts dials, a per-attempt dialer source with injectable failures, and a runner that connects, sends a fixed synthetic payload, reads the echo and reports a `tcp_connect` observation and its usage. It performs real loopback socket I/O so later tasks can exercise cancellation and budgets on real sockets. It reports the TCP-01 output shape only so result construction accepts it; it is not a TCP-01 implementation or conformance evidence, contacts no other host and is registered only by tests.

## Public Baseline and References

Fixture categories follow Draft `0.1.0` at public revision [`05c1364af4f9a604b617fca45f060949b557d265`](https://github.com/tv-rain/barometer-spec/commit/05c1364af4f9a604b617fca45f060949b557d265):

| Category | Controlling public reference | Helper or consumer responsibility |
|---|---|---|
| Malformed/trailing JSON, invalid text | [Runtime JSON decoding](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/data-formats/json-decoding.md), `REQ-JSON-0001` / `REQ-JSON-0002` | `InvalidJSON` starter corpus; extend for the selected model. |
| Duplicate and escaped-equivalent names | Same document, `REQ-JSON-0003` | `Object` preserves raw members, including repeated names; rejection remains the consumer's responsibility. |
| Unknown fields | Same document, `REQ-JSON-0004`; [task top-level fields](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/data-formats/task.md#top-level-fields), `REQ-TASK-0001` | Add a raw unknown member to an otherwise valid selected model. Dynamic-key objects retain their schema rules. |
| Malformed values and integer spellings | JSON decoding, `REQ-JSON-0005` through `REQ-JSON-0007` | Supply missing/null/wrong-type values and raw integer tokens such as `-0`, `1.0`, or `1e0`; do not normalize them in the fixture. |
| Unsupported versions | JSON decoding, `REQ-JSON-0008` | Supply a version rejected by the selected schema, without adding a version to versionless models. |
| Encoded sizes and hostile inputs | JSON decoding, `REQ-JSON-0009` / `REQ-JSON-0010`; [registered limits](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/library/registered-limits.md) | `ByteBoundaries`, `PaddedJSON`, and `NestedArray`; consumers determine registered/local limits and rejection. |
| Deadline/expiry boundaries | [Attempt limits](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/library/registered-limits.md#attempt-limits), [schedule](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/data-formats/task-schedule.md) | `TimeBoundaries` and `Clock`; consumers choose equality semantics from the contract. |

The older [source inventory](normative-sources.md) and [Go duplicate-member finding](specification-issues.md#go-spec-issue-001-duplicate-json-member-names-outside-labels) describe a historical baseline. The newer public contract explicitly rejects repeated decoded names throughout runtime JSON. This task references that contract for fixture intent; it does not update the global inventory, implement decoding, or claim adoption/conformance of production behavior.

## Adding Cases

Construct an otherwise valid synthetic model first and test acceptance as a positive control. Change one property per negative case, preserving every other precondition. `Member(name, rawValue)` encodes only the name; raw values are intentionally not validated. `Object` preserves raw member order, duplicates, escape spellings, and malformed values:

```go
input, err := testkit.Object(
    `"test_id":"NET-01"`,
    `"test_\u0069d":"TCP-01"`,
)
if err != nil { t.Fatal(err) }
// This is a fragment demonstrating duplicate preservation, not a complete task.
```

Use `RunRejections(t, cases, reject, matches)` to execute a table. Each `RejectionCase` contains a stable case name, allocated requirement ID (or exact immutable section for an unallocated obligation), expected bounded error class, and synthetic input. `reject` returns the consumer's error; `matches` compares the class using `errors.Is`/`errors.As` or another typed predicate. `AssertRejected` can also be used inside tests that check atomic failure, transport calls, or result publication separately. Diagnostics identify the requirement and expected class and never print input bytes or raw error messages. Case metadata is trusted test author content and must contain no sensitive values.

The runner copies inputs and tests both rejection and the expected class. It cannot itself establish absence of a partially decoded model, side effects, leaks, or sensitive errors from the consumer. Assert those separately. A case that returns an unrelated error must fail even though it rejected the input.

## Bounds and Deterministic Controls

The [Foundation validation boundary](foundation-validation.md) records delivered helper capabilities and owned downstream harness extensions. Foundation completion does not establish runtime deadline, effective lower-layer accounting, retention or clearing coverage.

All construction checks happen before generated allocations. Individual fixtures and aggregate scripted response bodies are capped at 1 MiB; encoded member names are limited to 4,096 source bytes, nesting generation to 1,024 levels, and transports to 64 steps. These are test tooling caps, not runtime specification limits. Larger stress tests require separately reviewed tooling. `Bytes` may deliberately produce invalid data; `PaddedJSON` adds trailing JSON whitespace to an existing value for an exact encoded-byte length without changing its content. Byte fixtures at maximum minus one, maximum, and maximum plus one have independent storage.

`Clock.Now` is safe for concurrent use. `Advance` refuses negative durations. Inject `Now` into the consumer and use `TimeBoundaries(at, quantum)` for before/at/after checks without sleeps. The clock has no timers and does not advance `context.WithDeadline`, host clocks, or operating-system timers; timer-driven execution harnesses remain separate work. `WaitClock` (DEV-635) adds a deterministic `Wait` for engines that wait on an injected clock: it returns the context error when the context is done, otherwise records the duration and advances the clock at once, so delay assertions need no sleeps.

`Transport` implements `http.RoundTripper` using ordered synthetic response/error/cancellation steps. It performs no network I/O and retains no request payloads or destinations. Concurrent requests take steps in lock-acquisition order; tests requiring a fixed order must synchronize dispatch. It closes request bodies on every non-nil request path. Response bodies honor cancellation, require explicit close, reject reads after close, and close idempotently. `Counts` reports calls and still-open response bodies; reading to EOF does not count as closing. Pre-cancelled requests consume no step. Exhaustion returns a fixed error with no fallback.

A `WaitForCancel` step requires a cancellable request context. Arrange cancellation with an explicit test synchronization point and a test timeout as a backstop; background contexts fail immediately. This transport does not model sockets, TLS handshakes, real buffering, lower-layer byte accounting, timing, or native platform behavior.

## Validation and Evidence

Run `go test ./internal/testkit` for focused helper tests and `make check` for the repository validation set. Validate allocated references using the explicitly supplied public requirement checker as described in [Requirement References](requirement-references.md). No third-party dependency, public schema change, release tag, production traceability manifest, native compilation requirement, or sibling-repository edit is introduced. Helper tests and portability checks establish only the reported synthetic/tooling scope.
