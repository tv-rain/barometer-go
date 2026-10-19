# Execution-State and Error Models

## Authority and Scope

DEV-480 consumes Draft `0.1.0` at public revision `59d49364d85e48cc8e69c762d130a7a86dc1fe65`: [Measurement Result](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/data-formats/measurement-result.md#REQ-RESULT-0019), [Task Trust](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/library/task-trust.md#REQ-TRUST-0033) and [Task Schedule](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/data-formats/task-schedule.md#REQ-SCHEDULE-0022). Public text controls behavior; this document records the narrower Go implementation.

`internal/model/task` owns the eleven admission-error codes (the ten initial REQ-TRUST-0034 classes and `task_unsupported`, added by DEV-597 for REQ-TRUST-0036) and all fifteen execution reasons. `internal/execution` owns a pure attempt value. Both use only existing dependencies and introduce no public facade, third-party dependency or new serialized schema. `task.ErrorContract` and `execution.Contract` identify the `0.1.0` vocabularies. Internal enum ordinals are not wire identifiers. Admission errors and attempts reject JSON import and export, including import into an existing value; future result and operational interfaces must use their reviewed contracts.

## Admission Errors

`NewAdmissionError` accepts a closed numeric `AdmissionCode`; zero and unknown codes are rejected. The error stores only that code and formats it as the public operational class. It has no payload, target, platform message, wrapped cause or arbitrary text field. Admission errors are distinct from `ExecutionReason` and cannot be passed to `Attempt.Stop`. A zero error value formats to a fixed invalid-error message and carries no valid code.

Existing decoding/configuration/validation sentinels retain their current contracts. The future admission owner must deliberately map failures at the appropriate stage. This task does not authenticate, authorize, reject deliveries, persist replay state or generate Measurement Results for invalid tasks.

## Attempt Transitions

An attempt is a copied value with private state; callers must retain each returned successor. Failed operations return the unchanged value and a fixed error. Zero is invalid. `NewAttempt` starts pending; it does not prove admission or eligibility.

| Current phase | Start | Complete | Stop with registered reason |
|---|---|---|---|
| Pending | Running | Reject | NotRun |
| Running | Reject | Completed | Incomplete |
| Completed, Incomplete, NotRun, or zero | Reject | Reject | Reject |

Pending and Running are internal phases. Completed, Incomplete and NotRun correspond to the public terminal states. Complete accepts no reason or protocol error; negative DNS, TCP, TLS and HTTP observations can still be terminal measurement outcomes. Stop accepts only a valid `ExecutionReason`, with unknown values, protocol failures and admission classes rejected even after an explicit string conversion. Terminal states cannot be reopened and subsequent conditions cannot replace the first reason.

The future runtime owner must serialize competing events and establish whether execution started, whether a runner produced a terminal outcome, and which outer condition occurred first. This value provides neither synchronization nor cancellation/cleanup, clocks, counters, path snapshots, policy enforcement or resource accounting. It must not be shared as mutable runtime state without an owning synchronization mechanism.

The future result constructor must require measurement data for completed results, forbid it for not-run results, and permit bounded incomplete output only where the test schema allows it (initially NET-05). This attempt model carries no protocol data or partial-result payload and cannot validate those result-envelope obligations.

## Retry Boundary

`RetryReason` returns a candidate only for a terminal incomplete/not-run attempt with `platform_suspended`, `network_context_unavailable`, `proxy_unavailable` or `internal_error`. It reuses `task.Retryable`, preserving existing task decoding and schedule semantics. Completed protocol outcomes cannot produce a retry candidate. Host cancellation, deadline/resource limits, policy rejection, network changes and other path failures remain non-retryable.

The candidate is not permission to execute. `validation.RetryEligible` must still enforce the accepted retry allowlist, attempt count, expiry, time window, realized due time, supported behavior and execution/safety constraints. Retry candidates create no new runs and do not bypass local policy or authorize network work.

## Synthetic Evidence and Remaining Work

- `internal/execution/attempt_test.go` covers every phase/operation pair, all registered reasons before and after start, preservation of the first condition, invalid/protocol/admission reasons, JSON rejection and integration with bounded schedule eligibility.
- `internal/model/task/admission_error_test.go` checks all eleven public codes, the entire invalid numeric domain, bounded error text, sensitive-input rejection and separation from execution reasons.
- Existing task-decoding and schedule tests remain applicable to the four allowed retry reasons.

These tests establish local model invariants. They do not establish admission correctness, runtime concurrency/cancellation, registered measurement conformance, result serialization, device behavior or platform support. The existing traceability manifests remain planned cross-cutting coverage; this task does not promote their broader case families to implemented or passed status.
