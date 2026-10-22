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

`NewEngine` takes the registry, a `Clock` (monotonic `Now`, a cancellable `Wait` and an `AfterFunc` timer), a `Jitter` source and the `Output` port; any missing input fails with `ErrEngine`. `SystemClock` and `LocalJitter` (uniform, from `crypto/rand`; a read failure yields no jitter) are the non-test implementations. A `Plan` carries only trusted local inputs prepared after admission: task ID, exact test version, normalized parameters, the validated schedule, the attempt deadline (the effective `max_duration_ms`, at most `MaxAttemptDuration` = 60 s, or `ErrPlan`), the effective attempt limits after admission (DEV-562, [Attempt Byte and Connection Budgets](#attempt-byte-and-connection-budgets)), the host-directed eligibility snapshot (`validation.StartContext`, whose `Now` the engine sets), the implementation identity, a `result.Decoder` bound to the same task, a `ContextSource` and a `DialerSource`.

`Run` fails closed before any work, with no result, no context capture, no dialer and no runner call, on an invalid plan (`ErrPlan`) or a missing runner (`ErrNoRunner`). It then draws one run ID (`result.NewRunID`) shared by every attempt of the run. A run whose context has already ended records attempt 1 as `not_run` with the cancellation reason and does nothing else. Otherwise a run that `validation.RunEligible` rejects fails closed (`ErrNotEligible`) with no result, and attempts execute in order, each under its own deadline ([Deadlines, Cancellation and Cleanup](#deadlines-cancellation-and-cleanup)):

1. Capture Network Context. A failure stops the attempt before it starts: `not_run` with `network_context_unavailable`, or with the cancellation reason when the attempt context ended.
2. Obtain a fresh dialer for the attempt number. A failure is `not_run` with `internal_error`.
3. Invoke the registered runner once through the attempt scope. A terminal output becomes `completed` with its measurement and Network Context. A runner error becomes `internal_error`, or the cancellation reason when the attempt context ended, and is `not_run` or `incomplete` depending on whether the test started. An attempt whose budget the scope exhausted is `incomplete` with `resource_limit`, even when the runner still returns an outcome. Output that result construction rejects for the test is not retained: the attempt becomes `incomplete` with `internal_error`.
4. Build the result through `result.Decoder.Build` and write it to the `Output` port. A `not_run` result never carries Network Context. Writes use a context that the run's cancellation does not end, so a cancelled attempt is still recorded and a racing cancellation cannot discard a committed outcome; the output port bounds its own writes. Any write error, including an uncertain commit, ends the run with `ErrOutput`; the write is not repeated and no further attempt starts.
5. Decide a retry. Only an attempt whose reason is a registered retry candidate (`Attempt.RetryReason`) can retry, and only when `validation.RetryEligible` accepts it before the wait. With at most two attempts, `fixed` and `exponential` backoff both produce `initial_delay_ms`; the engine adds local jitter in `[0, jitter_ms]` and caps the realized delay at `max_delay_ms` (REQ-SCHEDULE-0029). It does not retry when the realized delay would reach `not_after`, waits on the injected clock, and checks `validation.RetryEligible` again afterwards, so a closed time window or expired task ends the run. Completed attempts, including negative protocol outcomes, never retry (REQ-SCHEDULE-0025).

A run context that ends after an attempt prevents every further attempt, including a permitted retry and one already waiting in its backoff. The `Report` lists each terminal attempt written to the port, with its attempt value and, since DEV-562, the scope's metered `Usage` rather than the runner's report; it stays valid when `Run` returns an error.

### Reviewed Bounds

These are reviewed implementation inputs, not new public limits:

| Bound | Value | Source |
|---|---|---|
| Attempts per run | At most `MaxAttempts` = 2; the plan's `max_attempts` must be 1 or 2 | REQ-LIMIT-0011, REQ-SCHEDULE-0023 |
| Workers per run | 1: attempts run sequentially in the caller's goroutine. The engine starts no long-lived goroutine; the attempt deadline timer and the attempt scope's cancellation hook each run one short callback when they fire | Implementation choice; concurrent runs are the caller's responsibility |
| Attempt deadline | The plan's `max_duration_ms`, at most 60,000 ms, on the engine clock, covering context capture, the runner and cleanup | REQ-LIMIT-0006 |
| Cleanup bound | `CleanupBound` = 2,000 ms after cancellation or the attempt deadline, observed by `CancelRun` | REQ-LIMIT-0008 |
| Attempt byte and connection budgets | The plan's effective `max_bytes_sent`, `max_bytes_received` and `max_connections`, each from 0 to its registered maximum (`MaxBytesSent` = 65,536, `MaxBytesReceived` = 262,144, `MaxConnections` = 12), or `ErrPlan`; never clamped | REQ-LIMIT-0006, REQ-LIMIT-0034 |
| Retry delay | `initial_delay_ms` plus jitter up to `jitter_ms`, capped at `max_delay_ms`, never reaching `not_after` | REQ-SCHEDULE-0028, REQ-SCHEDULE-0029 |

### State Mapping and GO-SPEC-ISSUE-004

Since DEV-561 the engine follows the [GO-SPEC-ISSUE-004 disposition](specification-issues.md#go-spec-issue-004-boundary-between-not_run-and-incomplete) at public [`84f52f8`](https://github.com/tv-rain/barometer-spec/tree/84f52f8872a5af9e6d97894611e3b2eaba18aba8):

- **Test start (REQ-RESULT-0051).** An attempt is `not_run` until the test starts and `incomplete` after, when an outer condition stops it. The test starts at its first dial through the attempt scope or, for a runner that observes network state before dialing, when it calls `measurement.Input.Start`. Context capture, dialer acquisition and the runner's local validation do not start it. A runner that returns a terminal outcome has started.
- **Recorded outcome (REQ-RESULT-0052).** A returned outcome is `completed`, even when a cancellation or the deadline arrives while connections close. A runner interrupted before its outcome returns `measurement.ErrRunner`.
- **Network Context (REQ-RESULT-0053).** No `not_run` result carries Network Context, and `result.Decoder` rejects one that does; the former captured-context binding is removed.
- **Reasons.** An expired attempt deadline is `attempt_deadline_exceeded`. `CancelRun` and lifecycle transitions are host requests and record `cancelled_by_host`. The first cancellation is kept: a deadline that expired before a host cancellation stays the reason, and the reverse. An exhausted attempt byte or connection budget is `resource_limit` (DEV-562). `platform_suspended` (DEV-455) and `proxy_reconfigured` (DEV-578) belong to later tasks.

NET-05 composite states (REQ-NET-05-0061) and the unsatisfied-network-condition rule (REQ-EXECUTION-0031) need no engine change: no NET-05 runner exists (DEV-454), and network conditions are unsatisfied at the start gates, which produce no result.

## Deadlines, Cancellation and Cleanup

DEV-561 adds deadline and cancellation coordination. The public bounds are the attempt deadline, the wall-clock limit of the complete attempt including context acquisition and cleanup ([REQ-LIMIT-0006](https://github.com/tv-rain/barometer-spec/blob/84f52f8872a5af9e6d97894611e3b2eaba18aba8/library/registered-limits.md#REQ-LIMIT-0006)), and cleanup that stops within 2,000 ms of it ([REQ-LIMIT-0008](https://github.com/tv-rain/barometer-spec/blob/84f52f8872a5af9e6d97894611e3b2eaba18aba8/library/registered-limits.md#REQ-LIMIT-0008)).

- **Attempt deadline.** Each attempt runs under a child of the run context. An `AfterFunc` timer on the engine clock, armed when the attempt begins and stopped when it ends, cancels it with a deadline cause after `max_duration_ms`. Tests drive it with `testkit.WaitClock`, whose timers fire only when the test advances the clock.
- **Attempt scope.** The runner receives the scope in place of the path's dialer. It records the test start, ties every dial to the attempt context, and owns every connection opened through it: when the attempt context ends it closes them, which interrupts blocked reads and writes, and it refuses later dials. When the runner returns, the scope closes whatever the runner left open. Operating-system buffering and retransmission below the dialer are outside this guarantee and outside the counted budgets ([REQ-LIMIT-0033](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/library/registered-limits.md#REQ-LIMIT-0033)).
- **Run cancellation.** `StartRun` gives each run its own context, a child of the generation context. `CancelRun` cancels it and waits for the run's engine to return, for at most `CleanupBound` on the engine clock; lifecycle transitions cancel the generation context and wait for runs through their reservations, for at most the lifecycle deadline. A run ends with no further attempt.
- **Cleanup failure.** The engine cannot stop a runner that ignores its context, so it never reports cleanup as done before the runner returns. The run stays active, keeps its reservation (active-run count, exclusive flag and pending-output allowance) and starts no attempt, while its connections are already closed and new dials refused. `CancelRun` returns `cleanup_incomplete` and a later call waits again; `Close`, clearing and policy replacement return `cleanup_incomplete` or `clearing_incomplete` and stay blocked and retryable. Failure to clean up is a local condition, not a result field.

There is no separate run deadline: the public text defines none. A run is bounded by at most two attempts, each with its deadline, and a retry delay that never reaches `not_after`.

## Attempt Byte and Connection Budgets

DEV-562 meters the attempt scope, adopting the [GO-SPEC-ISSUE-005 disposition](specification-issues.md#go-spec-issue-005-which-bytes-count-toward-attempt-byte-limits) for the generic stream layer at public [`5d01879`](https://github.com/tv-rain/barometer-spec/tree/5d01879cfd1d47396c780befa2567f7e90439948). Attempt byte limits count where the attempt exchanges test bytes with a transport ([REQ-LIMIT-0034](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/library/registered-limits.md#REQ-LIMIT-0034)); for every runner that boundary is the `measurement.Conn` the scope returns, so the scope counts there and a runner's own read caps and `Usage` report decide nothing.

- **Effective limits.** `Plan.Limits` carries the attempt limits that admission narrowed to the task's values within the intersected issuer and local ceilings, so they are already the stricter of the task, local and registered values. The engine rejects a plan whose byte or connection limit is negative or above its registered maximum with `ErrPlan`, and never clamps. Zero permits no traffic ([REQ-LIMIT-0007](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/library/registered-limits.md#REQ-LIMIT-0007)).
- **Connections.** Every dial through the scope counts toward `max_connections` before it reaches the path's dialer, whether or not it connects; a refused connect has already sent traffic. A dial beyond the limit is not issued.
- **Sent bytes.** A write is issued only when the remaining sent budget covers all of it, so a write that would exceed `max_bytes_sent` sends nothing. A partial write counts only the bytes the stream accepted.
- **Received bytes.** Each read requests no more than the remaining received budget, so bytes a runner reads ahead into a buffer count when they are read and the counter never exceeds `max_bytes_received`. A read with no budget left is refused, so a runner that reads exactly the limit completes and one that needs one more byte is stopped.
- **Stop.** An operation the remaining budget cannot cover returns an error to the runner and cancels the attempt context with a budget cause. The scope then closes every connection and refuses further dials, reads and writes, as for the deadline and host cancellation, and the attempt is `incomplete` with `resource_limit` ([REQ-LIMIT-0027](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/library/registered-limits.md#REQ-LIMIT-0027), [REQ-LIMIT-0028](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/library/registered-limits.md#REQ-LIMIT-0028)). The stop is the first termination, so an outcome the runner returns afterwards is not recorded as `completed`. The first cancellation cause is kept, as for the deadline and host cancellation. `resource_limit` is not a retry candidate, and every attempt gets a fresh scope and budget.
- **No fallback.** Budget handling never requests another dialer or path; a refused operation ends the attempt.

Not counted at this layer, as the public text requires: TLS handshake and records, TCP/IP and UDP/IP headers, operating-system retransmissions and buffering, proxy transport framing and the `CONNECT` exchange with the loopback proxy endpoint, which belong below the stream a runner receives, and Network Context acquisition. Whether a runner's HTTP headers, DNS messages, DoH framing and DoT length prefixes cross this boundary as counted test bytes, how a TLS runner's plaintext is metered, UDP datagram accounting and the NET-05 aggregate limits are per-protocol evidence for the runners and codecs under DEV-454, which must route their traffic through this scope; the path and proxy seam that places the `CONNECT` exchange below it belongs to DEV-578.

### Seams and Owners

| Seam | Interim implementation | Owner |
|---|---|---|
| `Output` (pending results) | Test-only `internal/testkit/output` | Production store: DEV-563 |
| `ContextSource` | Test fakes | Network Context acquisition: DEV-578 |
| `DialerSource` / `measurement.Dialer` | Test-only `internal/testkit/loopback`; metered by the attempt scope since DEV-562 | Path and proxy: DEV-578; per-protocol accounting: DEV-454 |
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
- `internal/execution/budget_test.go` (DEV-562) covers `max_bytes_sent`, `max_bytes_received` and `max_connections` at limit −1, limit and limit +1 on real loopback sockets, comparing the metered counts with the bytes and connections the exchange server observed; a write over the remaining budget never reaching the server; byte accumulation over a reused connection; a failed connect counting as a connection; zero limits with no dial; read-ahead capped by the budget rather than the buffer; a fresh budget for a retried attempt and no retry after `resource_limit`; no further dial, read or write and no other dialer after exhaustion, with a later outcome not recorded as `completed`; release of a read blocked on another socket; exact partial reads and writes on a scripted connection; and the plan's limit bounds.
- `internal/registry/registry_test.go` covers exact runner lookup with no version fallback and fail-closed lookup for entries without a runner; `dependency_rules_test.go` covers the registry and runner-contract import rules.
- `internal/model/task/admission_error_test.go` checks all eleven public codes, the entire invalid numeric domain, bounded error text, sensitive-input rejection and separation from execution reasons.
- Existing task-decoding and schedule tests remain applicable to the four allowed retry reasons.
- `runtime_conformance_test.go` in the root package and `internal/execution/cross_cutting_conformance_test.go` (DEV-640) execute the runtime case families listed under [Runtime Case Families](#runtime-case-families).

These tests establish local model and synthetic orchestration invariants, and deadline, cancellation and generic stream budget behavior on the host test platform with loopback sockets. They do not establish admission correctness, per-protocol byte accounting, runtime cancellation or suspension behavior on devices, registered measurement conformance, result serialization, device behavior or platform support. The existing traceability manifests remain planned cross-cutting coverage; this task does not promote their broader case families to implemented or passed status.

## Runtime Case Families

DEV-640 classifies the 49 runtime-scope cross-cutting families for the runtime and execution closeout (DEV-576): the `CC-EXECUTION-*`, `CC-SCHEDULE-*` and `CC-LIMIT-*` families, their `RV-` review cases, and the run-dependent `CC-RESULT-EXECUTION-OBJECT`, `CC-RESULT-TIME-OBJECT`, `CC-SAFETY-USER-CONSENT`, `CC-OPTIONS-EXECUTION-PRECEDENCE`, `CC-CONTEXT-FORMAT-USE-IN-EXECUTION`, `CC-CONTEXT-EXECUTION-INTEGRATION`, `CC-PROXY-EXECUTION-STATES` and `CC-TRUST-LOCAL-POLICY-LIFECYCLE-RUNTIME`. A family is executed only when every row of its case plan runs with synthetic adapters, the injected clock and the test-only loopback runner. Mixed families are not split, and planned cases are not conformance.

Where a plan names a registered test only as the vehicle for a schedule or execution rule (for example the NET-01 tasks of `CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW`), the synthetic runner registered under the TCP-01 identifier stands in for it, and no behavior of the named test is asserted. This substitution was approved for DEV-640.

### Executed

These 18 families are executed with synthetic evidence:

| Family | Test |
|---|---|
| `CC-RESULT-EXECUTION-OBJECT` | `TestCcResultExecutionObject` (`internal/execution`) |
| `CC-EXECUTION-FIELDS` | `TestCcExecutionFields` (root) |
| `CC-EXECUTION-FIELDS-LIMITS` | `TestCcExecutionFieldsLimits` (`internal/execution`, real loopback sockets) |
| `CC-EXECUTION-FIELDS-CONCURRENCY` | `TestCcExecutionFieldsConcurrency` (root) |
| `CC-SAFETY-USER-CONSENT` | `TestCcSafetyUserConsent` (root) |
| `CC-SCHEDULE-FIELDS` | `TestCcScheduleFields` (root) |
| `CC-SCHEDULE-SCHEDULING-MODEL` | `TestCcScheduleSchedulingModel` (root) |
| `CC-SCHEDULE-SCHEDULING-MODEL-START-TIME` | `TestCcScheduleSchedulingModelStartTime` (root) |
| `CC-SCHEDULE-SCHEDULING-MODEL-TIME-LIMIT` | `TestCcScheduleSchedulingModelTimeLimit` (root) |
| `CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW` | `TestCcScheduleSchedulingModelTimeWindow` (root) |
| `CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-FIELDS` | `TestCcScheduleSchedulingModelTimeWindowFields` (root) |
| `CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-RANGE-FIELDS` | `TestCcScheduleSchedulingModelTimeWindowRangeFields` (root) |
| `CC-SCHEDULE-SCHEDULING-MODEL-REPEAT` | `TestCcScheduleSchedulingModelRepeat` (root) |
| `CC-SCHEDULE-SCHEDULING-MODEL-RETRY` | `TestCcScheduleSchedulingModelRetry` (root) |
| `CC-SCHEDULE-SCHEDULING-MODEL-RETRY-RETRY-FIELDS` | `TestCcScheduleSchedulingModelRetryRetryFields` (root) |
| `CC-SCHEDULE-VALIDATION` | `TestCcScheduleValidation` (root) |
| `CC-SCHEDULE-VALIDATION-ELIGIBILITY-RULES` | `TestCcScheduleValidationEligibilityRules` (root) |
| `CC-LIMIT-SCHEDULE-AND-RETRY-LIMITS` | `TestCcLimitScheduleAndRetryLimits` (root) |

Two rows run below the facade. The embedded one-time exception (REQ-LIMIT-0013) is a trusted release-review input that `Client.AdmitTask` does not take, so through the facade an embedded task without `not_after` fails closed with `task_schema_invalid`; the cases assert that, and assert the exception itself through `validation.ScheduleAt`. Where a rejection class is not fixed by the case plan, the cases assert the class of the [admission design](admission.md): a value above a registered maximum is `task_limit_exceeded`, a value below a minimum or outside a relationship is `task_schema_invalid`, and a future `issued_at` is `task_source_unauthenticated`.

### Planned

Every other runtime-scope family stays `planned` because a row of its case plan needs something that does not exist yet:

| Family | Remains planned because | Owner |
|---|---|---|
| `CC-RESULT-TIME-OBJECT` | Stage rounding needs protocol stages from registered runners. | DEV-454 |
| `CC-OPTIONS-EXECUTION-PRECEDENCE` | It needs per-test options with a protocol timeout. | DEV-454 |
| `CC-EXECUTION-FIELDS-LIMITS-FIELDS` | Its rows count HTTP headers, DNS messages and length prefixes. | DEV-454 |
| `CC-LIMIT-PURPOSE` | Its rows need HTTP-01, NET-03, DNS-01, DNS-03 and DNS-04 runners at the counted boundary. | DEV-454 |
| `CC-LIMIT-TASK-AND-IDENTIFIER-LIMITS` | Its rows use production input and options schemas and the NET-05 `step_id`. | DEV-454 |
| `CC-LIMIT-ATTEMPT-LIMITS` | Its rows need the NET-01 zero-traffic exception and the 15,000 ms protocol ceiling. | DEV-454 |
| `CC-LIMIT-DNS-LIMITS` | It needs DNS runners. | DEV-454 |
| `CC-LIMIT-HTTP-AND-VALIDATION-LIMITS` | It needs HTTP-01 and NET-03 runners. | DEV-454 |
| `CC-LIMIT-NET-05-COMPOSITE-LIMITS` | It needs the NET-05 codec and composite budgets. | DEV-454 |
| `CC-LIMIT-LIMIT-HANDLING` | Its rows need an HTTP-01 body bound, a NET-05 profile and production codecs. | DEV-454 |
| `CC-EXECUTION-ENVELOPE` | It checks protocol semantics through the direct and proxy transports. | DEV-578 |
| `CC-EXECUTION-FIELDS-NETWORK-PATH` | It needs proxy path selection and path failures. | DEV-578 |
| `CC-EXECUTION-FIELDS-NETWORK` | A task must run once the platform reports the required network values; network conditions are unsatisfied until they are evaluated. | DEV-578 |
| `CC-EXECUTION-FIELDS-NETWORK-FIELDS` | It needs network-type, VPN and validation evaluation. | DEV-578 |
| `CC-EXECUTION-FIELDS-NETWORK-FIELDS-VPN-POLICY` | It needs VPN-state evaluation. | DEV-578 |
| `CC-EXECUTION-FIELDS-NETWORK-FIELDS-NETWORK-CONTEXT-REQUIREMENTS` | It needs Network Context acquisition and endpoints. | DEV-578 |
| `CC-CONTEXT-FORMAT-USE-IN-EXECUTION` | It needs context requirement evaluation. | DEV-578 |
| `CC-CONTEXT-EXECUTION-INTEGRATION` | It needs context acquisition during execution. | DEV-578 |
| `CC-PROXY-EXECUTION-STATES` | It needs the proxy path and its failures. | DEV-578 |
| `CC-SCHEDULE-ENVELOPE` | It needs a platform suspension signal. | DEV-455 |
| `CC-EXECUTION-EXECUTION-MODEL` | Its retry rows need a `platform_suspended` attempt outcome from a platform signal. | DEV-455 |
| `CC-SCHEDULE-SCHEDULING-MODEL-LIMITS-FIELDS` | Its retry row needs a `platform_suspended` attempt outcome from a platform signal. | DEV-455 |
| `CC-EXECUTION-CONSTRAINT-EVALUATION` | Its rows need operating-system and user vetoes. | DEV-455 |
| `CC-EXECUTION-FIELDS-DEVICE-FIELDS` | Its rows need local device policy stricter than the task; local policy has no device constraints yet. | DEV-455 |
| `CC-TRUST-LOCAL-POLICY-LIFECYCLE-RUNTIME` | It needs the host policy adapter and Network Context caches. | DEV-455, DEV-578 |
| `CC-LIMIT-RESULT-AND-BUNDLE-LIMITS` | It needs pending storage and its maxima. | DEV-577 |
| `CC-LIMIT-CONTEXT-LINKAGE-AND-RETENTION-LIMITS` | It needs context endpoints, linkage and pending-bundle retention. | DEV-577, DEV-578 |
| `CC-SCHEDULE-SCHEDULING-MODEL-REPEAT-INTERVAL-FIELDS` | Delayed due runs depend on the open [GO-SPEC-ISSUE-011](specification-issues.md#go-spec-issue-011-due-times-and-catch-up-for-interval-repetition). | DEV-453 |
| `RV-CC-LIMIT-PURPOSE` | Human review of the per-test accounting mapping, which needs the registered runners. | DEV-454; review is non-gating |
| `RV-CC-LIMIT-CONTEXT-LINKAGE-AND-RETENTION-LIMITS` | Human review of deployment retention documentation. | DEV-456; review is non-gating |
| `RV-CC-LIMIT-LIMIT-HANDLING` | Human review of the limit history by an independent reviewer. | DEV-618; review is non-gating |
