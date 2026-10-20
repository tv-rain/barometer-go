# Sanitized API Examples

These examples are non-normative, repository-local Foundation examples for [DEV-484 — Publish sanitized API examples](https://app.clickup.com/t/86cbb8f3q). They use internal packages and cannot be imported as a supported external API. All executable code is in `_test.go` files. No production package, facade export, dependency, measurement, persistence adapter or platform support is added.

The public [Draft specifications](https://github.com/tv-rain/barometer-spec/tree/main) remain authoritative. Examples do not define schemas or authorize deployment. The [intended Go facade](../../docs/public-api.md) is a design whose runtime implementation remains pending.

## Run

From the Go repository root:

```sh
go test ./internal/examples -run '^Example_' -count=1 -v
make check
```

Each example includes an exact `Output` assertion, so `go test` both compiles and executes it. The existing CI workflow runs `make check`, whose `test` and `race` targets include this package through `go test ./...`. There is no separate module, build tag, generated snippet or excluded example directory. A local passing check does not establish a CI run; remote CI evidence follows the human-created PR.

## Directory Index

- [`models_test.go`](models_test.go) — configuration preparation, strict task candidate decoding, cancellation-state transitions and bounded result serialization against existing internal code.
- [`lifecycle_test.go`](lifecycle_test.go) — intended client call shapes and host failure handling using a scripted test-only client.

## Coverage and Limits

| Example | What executes | Authoritative contract and remaining work |
|---|---|---|
| `Example_configurationPreparation` | Closed configuration decoding with linkage disabled; missing-clock rejection; synthetic clock capability preparation; rejection of JSON trust assertions. | [Task trust](https://github.com/tv-rain/barometer-spec/blob/main/library/task-trust.md), [linkage configuration](https://github.com/tv-rain/barometer-spec/blob/main/library/result-linkage.md#initialization-configuration), and the local [configuration container design](../../docs/configuration-models.md). Preparation does not activate a client, install authenticated delivery, admit a task or initialize storage. |
| `Example_strictTaskDecoding` | An exact supplied `SYN-01` schema produces a candidate; unknown and duplicate fields reject atomically. | [Strict JSON](https://github.com/tv-rain/barometer-spec/blob/main/data-formats/json-decoding.md), [task format](https://github.com/tv-rain/barometer-spec/blob/main/data-formats/task.md), and [schedule](https://github.com/tv-rain/barometer-spec/blob/main/data-formats/task-schedule.md). The invented empty schema is test-only. Timing/origin are fixed validation inputs, not authentication; expiry uses the fixed acceptance time, not today's clock. Production test schemas, policy, target authorization, consent and replay checks remain separate admission work. |
| `Example_cancellationState` | A running attempt value becomes terminal with `cancelled_by_host`, cannot retry and cannot be overwritten by completion. | [Execution](https://github.com/tv-rain/barometer-spec/blob/main/data-formats/task-execution.md). This is pure state handling; it does not interrupt I/O or demonstrate bounded runtime cleanup. |
| `Example_resultSerialization` | A synthetic `not_run` record passes a local binding, serializes, round-trips canonically and rejects bypasses/attempt overflow. | [Measurement Result](https://github.com/tv-rain/barometer-spec/blob/main/data-formats/measurement-result.md) and [registered limits](https://github.com/tv-rain/barometer-spec/blob/main/library/registered-limits.md). No measurement or platform acquisition occurs. The Android enum, example version, fixed run identifier and timestamp are fixtures; runtime identifiers must use the reviewed random producer and actual implementation metadata. A `not_run` result without captured context contains no measurement. |
| `Example_intendedClientLifecycle` | A scripted initialization failure prevents starts; close is checked separately. | [Intended facade](../../docs/public-api.md). The local interface copies only the required intended signatures. Internal packages cannot import the root package, so the example cannot construct the production lifecycle shell (DEV-600), and there is no adapter-registration or authentication shortcut. |
| `Example_intendedCancellationAndClearing` | Scripted cancellation and clearing errors keep host work paused; explicit retries and repeated clearing demonstrate the call sequence. | [Clear-data contract](https://github.com/tv-rain/barometer-spec/blob/main/library/result-linkage.md#clear-data-contract) and [task trust](https://github.com/tv-rain/barometer-spec/blob/main/library/task-trust.md). The script checks calls/arguments and returns predetermined errors. It deletes nothing and cannot demonstrate idempotency, concurrency, interruption recovery or storage safety. |

The model and facade examples are separate scenarios. The synthetic task candidate is never admitted or connected to the separately fabricated result. They must not be assembled into a runnable measurement flow.

## Host Clearing Responsibilities

The intended host flow pauses admission, starts and delivery, calls `ClearMeasurementData`, and keeps the affected scope paused on any error. Explicit recovery may retry clearing. Success must also be followed by deletion of host-owned copies before the host reports complete removal. A later explicit initialization is required before new Barometer work. `Close` releases ownership and stops work; it does not replace privacy clearing.

Actual core clearing must block new work, cancel and clean up active attempts within bounds, delete results, replay/admission state, protected keys, epochs, sentinels and caches, invalidate handles/references, and clear sensitive memory where supported. It must send no request or upload, create no replacement secret, and report incomplete deletion explicitly. Implementation and interrupted/native storage validation remain future gates; successful scripted calls are evidence only of host control flow.

## Synthetic Data and Interpretation

The examples need no network, collector, deployment endpoint, credential, participant data, account identifier, protected key or personal path. Linkage stays disabled. They print fixed coarse outcomes, never raw task/configuration/result payloads or containing-integration errors. The fixed IDs are workflow fixtures and identify no installation or person. The synthetic safety fields do not waive deployment review or consent requirements.

Passing examples establish only the behavior exercised above. They establish no public runtime availability, measurement implementation, censorship finding, conformance, field validation, Android/Apple support or privacy-deletion guarantee.
