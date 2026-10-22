# Execution-State and Error Models

## Authority and Scope

DEV-480 consumes Draft `0.1.0` at public revision `59d49364d85e48cc8e69c762d130a7a86dc1fe65`: [Measurement Result](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/data-formats/measurement-result.md#REQ-RESULT-0019), [Task Trust](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/library/task-trust.md#REQ-TRUST-0033) and [Task Schedule](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/data-formats/task-schedule.md#REQ-SCHEDULE-0022). Public text controls behavior; this document records the narrower Go implementation.

`internal/model/task` owns the eleven admission-error codes (the ten initial REQ-TRUST-0034 classes and `task_unsupported`, added by DEV-597 for REQ-TRUST-0036) and all fifteen execution reasons. `internal/execution` owns a pure attempt value and, since DEV-635, the [run orchestration](#run-orchestration) engine that drives it. Both use only existing dependencies and introduce no public facade, third-party dependency or new serialized schema. `task.ErrorContract` and `execution.Contract` identify the `0.1.0` vocabularies. Internal enum ordinals are not wire identifiers. Admission errors and attempts reject JSON import and export, including import into an existing value; future result and operational interfaces must use their reviewed contracts.

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

## Run Orchestration

DEV-635 adds `execution.Engine`, the internal engine for one run of an admitted task. It is not exported through the root package; `Client.StartRun` (DEV-636, [Run Reservation](#run-reservation)) launches it, and `Client.CancelRun` (DEV-561, [Deadlines, Cancellation and Cleanup](#deadlines-cancellation-and-cleanup)) stops it. Public baseline: Draft `0.1.0` at [`4743b63`](https://github.com/tv-rain/barometer-spec/tree/4743b636f29feb1f0f8d042866d04ccda3532938) for [Retry](https://github.com/tv-rain/barometer-spec/blob/4743b636f29feb1f0f8d042866d04ccda3532938/data-formats/task-schedule.md#retry) and [Schedule and Retry Limits](https://github.com/tv-rain/barometer-spec/blob/4743b636f29feb1f0f8d042866d04ccda3532938/library/registered-limits.md#schedule-and-retry-limits).

`NewEngine` takes the registry, a `Clock` (monotonic `Now`, a cancellable `Wait` and an `AfterFunc` timer), a `Jitter` source and the `Output` port; any missing input fails with `ErrEngine`. `SystemClock` and `LocalJitter` (uniform, from `crypto/rand`; a read failure yields no jitter) are the non-test implementations. A `Plan` carries only trusted local inputs prepared after admission: task ID, exact test version, normalized parameters, the validated schedule, the attempt deadline (the effective `max_duration_ms`, at most `MaxAttemptDuration` = 60 s, or `ErrPlan`), the host-directed eligibility snapshot (`validation.StartContext`, whose `Now` the engine sets), the implementation identity, a `result.Decoder` bound to the same task, a `ContextSource` and a `DialerSource`.

`Run` fails closed before any work, with no result, no context capture, no dialer and no runner call, on an invalid plan (`ErrPlan`) or a missing runner (`ErrNoRunner`). It then draws one run ID (`result.NewRunID`) shared by every attempt of the run. A run whose context has already ended records attempt 1 as `not_run` with the cancellation reason and does nothing else. Otherwise a run that `validation.RunEligible` rejects fails closed (`ErrNotEligible`) with no result, and attempts execute in order, each under its own deadline ([Deadlines, Cancellation and Cleanup](#deadlines-cancellation-and-cleanup)):

1. Capture Network Context. A failure stops the attempt before it starts: `not_run` with `network_context_unavailable`, or with the cancellation reason when the attempt context ended.
2. Obtain a fresh dialer for the attempt number. A failure is `not_run` with `internal_error`.
3. Invoke the registered runner once through the attempt scope. A terminal output becomes `completed` with its measurement and Network Context. A runner error becomes `internal_error`, or the cancellation reason when the attempt context ended, and is `not_run` or `incomplete` depending on whether the test started. Output that result construction rejects for the test is not retained: the attempt becomes `incomplete` with `internal_error`.
4. Build the result through `result.Decoder.Build` and write it to the `Output` port. A `not_run` result never carries Network Context. Writes use a context that the run's cancellation does not end, so a cancelled attempt is still recorded and a racing cancellation cannot discard a committed outcome; the output port bounds its own writes. Any write error, including an uncertain commit, ends the run with `ErrOutput`; the write is not repeated and no further attempt starts.
5. Decide a retry. Only an attempt whose reason is a registered retry candidate (`Attempt.RetryReason`) can retry, and only when `validation.RetryEligible` accepts it before the wait. With at most two attempts, `fixed` and `exponential` backoff both produce `initial_delay_ms`; the engine adds local jitter in `[0, jitter_ms]` and caps the realized delay at `max_delay_ms` (REQ-SCHEDULE-0029). It does not retry when the realized delay would reach `not_after`, waits on the injected clock, and checks `validation.RetryEligible` again afterwards, so a closed time window or expired task ends the run. Completed attempts, including negative protocol outcomes, never retry (REQ-SCHEDULE-0025).

A run context that ends after an attempt prevents every further attempt, including a permitted retry and one already waiting in its backoff. The `Report` lists each terminal attempt written to the port, with its attempt value and the runner's `Usage`; it stays valid when `Run` returns an error.

### Reviewed Bounds

These are reviewed implementation inputs, not new public limits:

| Bound | Value | Source |
|---|---|---|
| Attempts per run | At most `MaxAttempts` = 2; the plan's `max_attempts` must be 1 or 2 | REQ-LIMIT-0011, REQ-SCHEDULE-0023 |
| Workers per run | 1: attempts run sequentially in the caller's goroutine. The engine starts no long-lived goroutine; the attempt deadline timer and the attempt scope's cancellation hook each run one short callback when they fire | Implementation choice; concurrent runs are the caller's responsibility |
| Attempt deadline | The plan's `max_duration_ms`, at most 60,000 ms, on the engine clock, covering context capture, the runner and cleanup | REQ-LIMIT-0006 |
| Cleanup bound | `CleanupBound` = 2,000 ms after cancellation or the attempt deadline, observed by `CancelRun` | REQ-LIMIT-0008 |
| Retry delay | `initial_delay_ms` plus jitter up to `jitter_ms`, capped at `max_delay_ms`, never reaching `not_after` | REQ-SCHEDULE-0028, REQ-SCHEDULE-0029 |

### State Mapping and GO-SPEC-ISSUE-004

Since DEV-561 the engine follows the [GO-SPEC-ISSUE-004 disposition](specification-issues.md#go-spec-issue-004-boundary-between-not_run-and-incomplete) at public [`84f52f8`](https://github.com/tv-rain/barometer-spec/tree/84f52f8872a5af9e6d97894611e3b2eaba18aba8):

- **Test start (REQ-RESULT-0051).** An attempt is `not_run` until the test starts and `incomplete` after, when an outer condition stops it. The test starts at its first dial through the attempt scope or, for a runner that observes network state before dialing, when it calls `measurement.Input.Start`. Context capture, dialer acquisition and the runner's local validation do not start it. A runner that returns a terminal outcome has started.
- **Recorded outcome (REQ-RESULT-0052).** A returned outcome is `completed`, even when a cancellation or the deadline arrives while connections close. A runner interrupted before its outcome returns `measurement.ErrRunner`.
- **Network Context (REQ-RESULT-0053).** No `not_run` result carries Network Context, and `result.Decoder` rejects one that does; the former captured-context binding is removed.
- **Reasons.** An expired attempt deadline is `attempt_deadline_exceeded`. `CancelRun` and lifecycle transitions are host requests and record `cancelled_by_host`. The first cancellation is kept: a deadline that expired before a host cancellation stays the reason, and the reverse. `platform_suspended` (DEV-455), `proxy_reconfigured` (DEV-578) and `resource_limit` (DEV-562) belong to later tasks.

NET-05 composite states (REQ-NET-05-0061) and the unsatisfied-network-condition rule (REQ-EXECUTION-0031) need no engine change: no NET-05 runner exists (DEV-454), and network conditions are unsatisfied at the start gates, which produce no result.

## Deadlines, Cancellation and Cleanup

DEV-561 adds deadline and cancellation coordination. The public bounds are the attempt deadline, the wall-clock limit of the complete attempt including context acquisition and cleanup ([REQ-LIMIT-0006](https://github.com/tv-rain/barometer-spec/blob/84f52f8872a5af9e6d97894611e3b2eaba18aba8/library/registered-limits.md#REQ-LIMIT-0006)), and cleanup that stops within 2,000 ms of it ([REQ-LIMIT-0008](https://github.com/tv-rain/barometer-spec/blob/84f52f8872a5af9e6d97894611e3b2eaba18aba8/library/registered-limits.md#REQ-LIMIT-0008)).

- **Attempt deadline.** Each attempt runs under a child of the run context. An `AfterFunc` timer on the engine clock, armed when the attempt begins and stopped when it ends, cancels it with a deadline cause after `max_duration_ms`. Tests drive it with `testkit.WaitClock`, whose timers fire only when the test advances the clock.
- **Attempt scope.** The runner receives the scope in place of the path's dialer. It records the test start, ties every dial to the attempt context, and owns every connection opened through it: when the attempt context ends it closes them, which interrupts blocked reads and writes, and it refuses later dials. When the runner returns, the scope closes whatever the runner left open. Without lower-layer budgets (DEV-562), operating-system buffering and retransmission below the dialer are outside this guarantee.
- **Run cancellation.** `StartRun` gives each run its own context, a child of the generation context. `CancelRun` cancels it and waits for the run's engine to return, for at most `CleanupBound` on the engine clock; lifecycle transitions cancel the generation context and wait for runs through their reservations, for at most the lifecycle deadline. A run ends with no further attempt.
- **Cleanup failure.** The engine cannot stop a runner that ignores its context, so it never reports cleanup as done before the runner returns. The run stays active, keeps its reservation (active-run count, exclusive flag and pending-output allowance) and starts no attempt, while its connections are already closed and new dials refused. `CancelRun` returns `cleanup_incomplete` and a later call waits again; `Close`, clearing and policy replacement return `cleanup_incomplete` or `clearing_incomplete` and stay blocked and retryable. Failure to clean up is a local condition, not a result field.

There is no separate run deadline: the public text defines none. A run is bounded by at most two attempts, each with its deadline, and a retry delay that never reaches `not_after`.

### Seams and Owners

| Seam | Interim implementation | Owner |
|---|---|---|
| `Output` (pending results) | Test-only `internal/testkit/output` | Production store: DEV-563 |
| `ContextSource` | Test fakes | Network Context acquisition: DEV-578 |
| `DialerSource` / `measurement.Dialer` | Test-only `internal/testkit/loopback` | Path, proxy and budget accounting: DEV-578, DEV-562 |
| `Clock` | `SystemClock`; `testkit.WaitClock` in tests | Platform timer and suspension behavior: DEV-455 |
| `Preconditions` (DEV-636) | `execution.Gates`; scripted fakes in tests | Network conditions: DEV-578 |

Since DEV-636 the engine also evaluates the plan's optional `Preconditions` before every attempt: unsatisfied before the first attempt, `Run` fails closed with `ErrNotEligible` and no result; unsatisfied before a retry, the run ends with the first attempt's result unchanged and no second attempt. The `Output` port gains `Capacity`, which run reservation checks before a run starts.

## Run Reservation

DEV-636 adds the start-time decisions behind `Client.StartRun` ([Implemented Subset](public-api.md#implemented-subset)). The host directs every wakeup; nothing here starts a run on its own.

### Run Accounting

`persistence.RunRecord` is the schedule accounting of one admitted task, keyed by its replay key and held in the `persistence.RunLedger` port: the anchor of occurrence 0, the number of reserved runs, the next occurrence that has neither been reserved nor skipped, and that occurrence's due time with jitter included. The record is written before a run exists, so an interrupted or uncertain write can only lose a run. Reinitialization does not reset it; the production store, its retention and its deletion by clearing belong to DEV-577 and DEV-563, and the only implementation is the test-only `internal/testkit/ledger`.

`execution.NewAccounting`, `Occurrence`, `Advance`, `Due` and `Reserve` implement the interim rule of [GO-SPEC-ISSUE-011](specification-issues.md#go-spec-issue-011-due-times-and-catch-up-for-interval-repetition): the anchor is `starts_at` (rounded up to a whole second) or the acceptance time; occurrence 0 is due at the anchor without jitter; occurrence *k* of an `interval` schedule is due at anchor + *k* × `interval` plus jitter drawn from the engine's `Jitter` source once, capped at the task's `jitter`; the current occurrence is the latest whose unjittered time has passed, and once that reaches `max_runs` the schedule has ended; earlier unreserved occurrences are skipped and consume no run. Repeat mode `none` permits one run. `validation.RunEligible` still decides `starts_at`, `not_after`, the time window and `max_runs` for the due occurrence.

When no run was reserved before a reinitialization, the anchor of a task without `starts_at` is the redelivery's acceptance time, which can only move occurrences later. This and every other part of the rule are interim until public disposition.

### Attempt-Start Gates

`execution.Gates` is the plan's `Preconditions`, evaluated by `StartRun` before reservation and by the engine before each attempt. Every missing capability, error or undetermined value is unsatisfied; nothing is guessed:

- **Network conditions.** Until DEV-578 evaluates them, a `proxy` path, `allowed_transports`, a `vpn_policy` other than `any`, `require_validated` and Network Context requirements are unsatisfied ([REQ-EXECUTION-0031](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/data-formats/task-execution.md#REQ-EXECUTION-0031)), so such a task never runs and never reaches the path seam. Admission still rejects tasks that require egress Network Context as `task_unsupported`.
- **Device.** `minimum_battery_percent` needs a known battery level at or above it, and `allow_charging_only` a known charging state, from `platform.Device`.
- **Consent.** Only for a task whose consent requirement admission recorded, `platform.Consent` must report `ConsentValid` ([Consent Handoff](admission.md#consent-handoff)). The query carries no task value. Absent, ended, unavailable, an error or a missing capability start no attempt and produce no result.

`StartRun` maps an unsatisfied gate to `run_not_due`, so no result is written for it. Concurrency is the facade's, because it spans runs: an `exclusive` run starts only when no other run is active, and no run starts beside an active exclusive run. An absent concurrency member states no requirement and is treated as `shared`.

### Reviewed Inputs

| Input | Value | Basis |
|---|---|---|
| Active runs per client | 4, each on one goroutine | Implementation choice; the registered limits set no ceiling |
| Pending-output check | Free capacity at least the new run's `max_attempts` plus the `max_attempts` of every active run | Backpressure before capacity is exhausted ([Results and Delivery](public-api.md#results-and-delivery)) |
| Reservation lock | One per generation, held only for calls bounded by the generation context and the lifecycle deadline | Serializes the due check, the ledger write and the run handle |

## Synthetic Evidence and Remaining Work

- `internal/execution/attempt_test.go` covers every phase/operation pair, all registered reasons before and after start, preservation of the first condition, invalid/protocol/admission reasons, JSON rejection and integration with bounded schedule eligibility.
- `internal/execution/run_test.go` (DEV-635) covers attempt and retry sequencing on an injected clock with no sleeps: completed and negative outcomes without retry, runner failure with and without a retry object, a reason outside `retry_on`, retry then completion, an exhausted retry budget, the jitter cap at `max_delay_ms`, no retry when the delay reaches `not_after`, re-checked eligibility after the wait, pre-start `not_run` paths without Network Context, invalid runner output, host cancellation, output-port failures before and after a durable write, fail-closed plan, runner and eligibility checks, and real loopback TCP I/O through the attempt-scoped dialer (not TCP-01 evidence).
- `internal/execution/occurrence_test.go` (DEV-636) covers the anchor, interval due times with jitter drawn once and capped, skipped occurrences, the end of the schedule after the last occurrence, one run for repeat mode `none`, and the run budget at limit −1, limit and limit +1. `gates_test.go` covers every unsatisfiable network condition, consent statuses, errors and a missing capability, battery and charging boundaries, and that consent and device state are queried only for a requirement. `TestPreconditionsAreRecheckedBeforeEachAttempt` covers the engine's check before each attempt.
- `runs_test.go` in the root package (DEV-636) covers `StartRun` and `RunState` end to end with synthetic adapters; see the [DEV-636 validation record](public-api.md#dev-636-validation-record).
- `internal/execution/deadline_test.go` (DEV-561) covers the attempt deadline one millisecond before, at and after `max_duration_ms` on the injected clock; expiry and host cancellation before the test starts (in context capture and before the first network operation) as `not_run` without Network Context; cancellation during a dial and during I/O on a real loopback socket as `incomplete`, with the server observing the socket closed; the deadline during real I/O; the first cancellation kept as the reason; release of a socket the runner left open and refusal of a dial after the attempt ended; 200 cancel/complete races, each with one result whose state matches the report; and the deadline bounds of the plan. `TestHostCancellation` covers cancellation before the run, before and after the test start, after the outcome and during the retry backoff. `cancel_run_test.go` in the root package covers `CancelRun` end to end; see the [DEV-561 validation record](public-api.md#dev-561-validation-record).
- `internal/registry/registry_test.go` covers exact runner lookup with no version fallback and fail-closed lookup for entries without a runner; `dependency_rules_test.go` covers the registry and runner-contract import rules.
- `internal/model/task/admission_error_test.go` checks all eleven public codes, the entire invalid numeric domain, bounded error text, sensitive-input rejection and separation from execution reasons.
- Existing task-decoding and schedule tests remain applicable to the four allowed retry reasons.

These tests establish local model and synthetic orchestration invariants, and deadline and cancellation behavior on the host test platform with loopback sockets. They do not establish admission correctness, byte or connection budget enforcement, runtime cancellation or suspension behavior on devices, registered measurement conformance, result serialization, device behavior or platform support. The existing traceability manifests remain planned cross-cutting coverage; this task does not promote their broader case families to implemented or passed status.
