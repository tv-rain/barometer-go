# Public Go API Boundary

**Status:** Defined design; lifecycle shell implemented (DEV-600); run start and polling implemented (DEV-636); run cancellation and attempt deadlines implemented (DEV-561); runtime support pending
**Task:** [DEV-476 — Define the public Go API boundary](https://app.clickup.com/t/86cbb8f0j); lifecycle shell: [DEV-600](https://app.clickup.com/t/124kk1zcdcy); runs: [DEV-636](https://app.clickup.com/t/124kk1zcn0g); cancellation: [DEV-561](https://app.clickup.com/t/124kk1zc4a9)
**Specification baseline:** Draft `0.1.0` at `95dfeef1367f4c320bca7c7ae91b13ca7608b37f`

## Scope and Authority

This document defines the intended root-package facade for initialization, admission, host-triggered execution, cancellation, result retrieval, and measurement-data clearing. It is an implementation-local API design subordinate to the public contracts in the [Normative Source Inventory](normative-sources.md). The design itself introduces no task/result schema change, measurement implementation, conformance claim, release, or platform-support claim. Production exports are added one reviewed task at a time; [Implemented Subset](#implemented-subset) records which exist.

The signatures below are checked in an isolated synthetic module. Fixture bodies deliberately panic if invoked; they validate compiler/binding compatibility only. Do not copy those bodies into production. Implementation of this facade requires admission, execution, persistence, and reviewed platform/trust adapters. No successful no-op or permissive fallback may substitute for those components.

## Responsibility Boundary

| Owner | Responsibilities |
|---|---|
| Host | Enablement; transporting signed remote task packages unchanged, without performing or asserting their authentication; authenticating policy updates through a reviewed integration; scheduling calls within platform restrictions; operating the optional proxy; transporting returned bundles; reporting durable acknowledgement; connecting user-visible clearing; deployment review. |
| Go core | Complete atomic admission; issuer/local-policy/target/replay enforcement; execution eligibility and run/attempt accounting; bounded measurement and cancellation; minimized result/bundle construction; retention and clearing. |
| Barometer-owned native integration | Normalized platform capability acquisition, protected storage, trusted time, and installation of reviewed delivery/policy adapters through implementation-internal seams. |
| Flutter wrapper | Plain serializable configuration and result values plus lifecycle calls. Authentication assertions, key material, native objects, callbacks, and storage handles are excluded. |

The core has no mandatory scheduler, collector, backend, cloud provider, or application dependency. Direct Go integration uses the same reviewed adapter boundary; an out-of-tree dependency-injection API is not defined here.

## Intended Signatures

The module root remains the sole public package. `Client` is an opaque instance with no exported fields. Its Go wrapper is owned by the native binding; Flutter does not receive a Go object.

<!-- public-api-signatures:start -->
```go
package barometer

type Client struct{}

func NewClient() *Client
func (c *Client) Initialize(configJSON string) error
func (c *Client) AdmitTask(taskJSON string, sourceRef string) (string, error)
func (c *Client) StartRun(taskHandle string) (string, error)
func (c *Client) RunState(runHandle string) (string, error)
func (c *Client) CancelRun(runHandle string) error
func (c *Client) NextBundle() (string, error)
func (c *Client) AcknowledgeBundle(bundleID string) error
func (c *Client) SetProxy(proxyJSON string) error
func (c *Client) ClearProxy() error
func (c *Client) ClearMeasurementData() error
func (c *Client) Close() error
```
<!-- public-api-signatures:end -->

All parameters are immutable Go strings; no mutable byte-slice alias crosses the facade. Return pairs use the built-in `error` in second position, compatible with gomobile. Native wrappers translate errors into bounded operational codes and invoke blocking lifecycle operations away from the UI thread. No public `context.Context`, channel, function callback, map, generic value, platform interface, key handle, or storage object is required.

The declaration-only `Client struct{}` represents hidden production state. A zero-value or binding-created instance carries no authority and behaves as uninitialized. Constructors must not begin measurements or create linkage secrets. Copies of a Go value must not establish a second owner of the same persistent state; the eventual implementation must enforce one active owner per storage scope internally.

## Implemented Subset

DEV-600 is the public-API review point for `NewClient`, `Initialize`, `Close` and `ClearMeasurementData`, which the root package now exports with the signatures above, DEV-592 for `AdmitTask`, DEV-636 for `StartRun` and `RunState` and DEV-561 for `CancelRun`. Together they form the lifecycle shell, task admission and host-triggered, cancellable runs. The other documented methods are not exported yet; each later task adds its method as its own review point. `TestPublicSurfaceMatchesDocumentedSignatures` fails if a production export is missing from the signature block, differs from it, or is not in the implemented list, and if `Client` gains an exported field.

| Area | Current behavior |
|---|---|
| States | Uninitialized, initialized, transitioning, blocked (a transition that did not complete) and closed, following the [Lifecycle](#lifecycle-proxy-changes-and-clearing) table. Transitioning and blocked form its "transitioning or clearing incomplete" row. |
| Trusted adapters | An unexported seam binds the clock, registry, `task.LocalPolicy` (whose generation is the active one), `admission.DeliveryStore`, `persistence.ReplayStore`, configuration approvals, the optional linkage key store and, since DEV-636, the run inputs: the engine clock and jitter source, the `execution.Output` pending-output port, the `persistence.RunLedger`, the interim direct-path seam and the optional consent, device-state and time-zone capabilities. A missing required run input fails initialization closed. Nothing outside the root package can bind them, and no production code does yet, so `Initialize` fails closed with `not_initialized`. Tests bind synthetic adapters from `internal/testkit`. Native wiring belongs to DEV-455. Trusted issuer keys are part of `task.LocalPolicy`; there is no separate key, verifier or verdict input (DEV-612). |
| `Initialize` | Requires every mandatory adapter, then decodes the closed `proxy`/`network_context`/`result_linkage` container through `internal/config`, prepares it and builds the admitter; only a complete result becomes a new generation. Any failure returns `not_initialized` and leaves the client uninitialized; no configuration-specific class is added. An invalid local policy is accepted here and fails each later admission with `task_policy_unavailable`, as `NewAdmitter` already defines. The policy and its target registrations stay immutable for the generation; a valid policy whose generation is below the highest this instance has activated fails with `not_initialized` (DEV-593). |
| Serialization | Lifecycle calls mark the client transitioning and release the lock before adapter work, so a losing call returns `transition_in_progress` instead of blocking. `AdmitTask`, `StartRun`, `RunState`, `CancelRun` and future bundle entry points reserve the current generation through an unexported counter; a launched run keeps its reservation until its engine returns. A transition cancels the generation's adapter context, the parent of every run's context, and waits for reservations to finish, up to the lifecycle deadline. A cancelled attempt is recorded as `cancelled_by_host` ([Deadlines, Cancellation and Cleanup](execution-models.md#deadlines-cancellation-and-cleanup)). |
| Handles | Separate task and run tables per generation, holding 128-bit `crypto/rand` values encoded as base64url with a `T.` or `R.` prefix. The prefix keeps them distinct from `task_id` (lowercase) and from `run_id`/`bundle_id` (bare base64url). A full table returns `capacity_exceeded` without eviction; an unknown, other-instance, earlier-generation or invalidated handle returns `handle_unavailable`. Formatting redacts tables and runtime state. `AdmitTask` issues task handles and `StartRun` run handles. A full run table first drops terminal runs; it never evicts an active run. |
| `StartRun` | Reserves the generation, resolves the task handle and, under one reservation lock, applies every start check in order: run accounting from the `persistence.RunLedger` (a missing record is created with the task's anchor); a repeated start for the occurrence that already has a run, which returns that run's handle while its record is kept; the interim due-time rule of [GO-SPEC-ISSUE-011](specification-issues.md#go-spec-issue-011-due-times-and-catch-up-for-interval-repetition), which records a skipped occurrence before anything else; `validation.RunEligible` with the injected engine clock, the device time zone for a window and a registered runner; attempt-start gates ([Run Reservation](execution-models.md#run-reservation)); `shared`/`exclusive` concurrency and the active-run limit; pending-output capacity for every result the new and the active runs may write; and the direct-path seam. Only then is a run handle published and the run recorded in the ledger, before the run exists, and launched on its own goroutine through `execution.Engine`. An unavailable ledger, an uncertain reservation write, unsatisfied consent, device or network conditions, an exclusive conflict, insufficient pending-output capacity and a path failure all return `run_not_due` and start nothing; an uncertain write loses the occurrence rather than risking an extra run. The active-run limit and a full run table return `capacity_exceeded`. A panic below the facade returns `run_not_due`. |
| `RunState` | Returns `active` until the run's engine has returned and `terminal` afterwards, whatever its attempts' outcomes. These are polling values, not Measurement Result execution states. Unknown, other-instance, earlier-generation and reclaimed handles return `handle_unavailable`. |
| Attempt deadline | Since DEV-561, `StartRun` passes the task's effective `max_duration_ms` to the engine, which ends each attempt at that time on the engine clock, including context capture and cleanup, with `attempt_deadline_exceeded`. Attempts dial only through an attempt scope that closes their connections when the attempt ends. An attempt is `not_run` until its test starts and `completed` once its runner returned an outcome (GO-SPEC-ISSUE-004 disposition, [State Mapping](execution-models.md#state-mapping-and-go-spec-issue-004)). |
| `CancelRun` | Reserves the generation and resolves the run handle. A terminal run, including a repeated cancellation of a retained one, returns success without changing its results. Otherwise it cancels the run's own context, which prevents every further attempt and ends the active one as `cancelled_by_host` (`not_run` before the test starts), and waits for the run's engine to return for at most the registered 2,000 ms cleanup bound on the engine clock. If cleanup has not finished it returns `cleanup_incomplete`: the run stays `active`, keeps its active-run, exclusive and pending-output reservation and starts no attempt, its connections are closed and new dials refused, and a later `CancelRun` waits again. Unknown, task, other-instance, earlier-generation and reclaimed handles return `handle_unavailable`; lifecycle refusals return the API-local codes. A panic below the facade returns `cleanup_incomplete`. |
| `AdmitTask` | Reserves the current generation, then runs `Admitter.Admit` (every implemented admission step, including the complete-task byte limit after delivery resolution, and the durable replay commit) under the generation's context, bounded by the lifecycle deadline. Only an accepted task gets a handle. A rejection returns the code-only `task.AdmissionError` class and leaves no handle; lifecycle refusals return the API-local codes. An exact redelivery, through the same or a new delivery reference for embedded origin or a new reference for remote origin, returns the handle already published in this generation, writes no replay record and keeps the original expiry; concurrent identical deliveries share one handle. A redelivery after reinitialization resolves to the retained acceptance under a new handle. If a transition invalidates the table after the replay commit, no handle is published and the call returns `transition_in_progress`; a later redelivery resolves to the same acceptance. A panic below the facade returns `task_trust_unavailable` with no handle. The facade adds no byte check; reading natively received bytes through a counting reader remains with DEV-455. |
| `Close` | Invalidates both handle tables, closes the delivery store (dropping delivery records), cancels the generation context and with it every active run, whose attempt is recorded as `cancelled_by_host`, waits for them like any reservation and releases the runtime and trusted adapters, keeping only the clearing capability for recovery clearing. It deletes no replay or pending state. If reserved work does not finish within the deadline it returns `cleanup_incomplete` and stays blocked and retryable. Enforcing one owner per storage scope stays with DEV-577 and DEV-455; there is no ownership capability to release yet. No context cache exists yet. |
| `ClearMeasurementData` | Blocks new work, invalidates handles and delivery references, cancels active runs (recorded as `cancelled_by_host`) and waits for them up to the lifecycle deadline, then calls the unexported persistent-clear capability. That capability is nil until DEV-564 and DEV-577 provide it, so clearing returns `clearing_incomplete` and keeps activity blocked; it never reports success without deletion. With a capability, success ends uninitialized, or closed if already closed. Before initialization or after close it still returns `clearing_incomplete`. |
| Policy replacement | DEV-593 adds the unexported `replacePolicy` seam, the only path that changes the active policy; like the adapter seam it has no production caller, and the reviewed host adapter (DEV-455) must authenticate and verify an update before calling it. No task, collector reply, target, proxy or context endpoint reaches it, and no policy format is added. Replacement marks the client transitioning, then validates while the current generation stays active: the policy must be complete and valid, its generation strictly above the active one and not below the highest generation of retained replay records, and the successor generation must build. A refusal returns `policy_rejected` and leaves the previous policy, its handles and its delivery references active. It then cancels the current generation's adapter context, so in-flight admission stops, invalidates both handle tables, drops delivery records and moves the delivery store to the new generation, waits for reserved entry points up to the lifecycle deadline and activates the prepared generation. Tasks still permitted are redelivered and resolve to their retained acceptance; a task no longer permitted is rejected at admission, and a delivery under an earlier generation is `policy_rollback`. A failure after validation leaves the client uninitialized with no policy (`not_initialized`), so a later `Initialize` admits nothing until a complete policy is bound, or blocked with `cleanup_incomplete` while reserved work continues; neither activates part of the new policy or an earlier generation. Cancelling the generation context also stops active runs, recorded as `cancelled_by_host`, which the replacement waits for like any reservation. No Network Context cache exists to invalidate (DEV-578). The highest activated generation is held in memory for the instance: it survives clearing but not a new instance or restart. |
| Errors | An unexported error type carries one fixed API-local code with no wrapped cause. Each exported method recovers internal panics to a fixed code and keeps the client blocked when a transition is interrupted. A nil `*Client` fails closed without panicking. |

Reviewed implementation inputs selected by DEV-600:

| Input | Value | Basis |
|---|---|---|
| Task handle table capacity | 256 | Equal to `admission.MaxDeliveryCapacity`, because each admission consumes one delivery record. The registered limits set no ceiling on locally admitted tasks. When the table is full, `AdmitTask` first reclaims tasks whose validity ended at or before the admission time (DEV-592); it never evicts a valid task. |
| Run handle table capacity | 256 | The registered limits set no ceiling on outstanding runs; each task is still limited to 28 runs. |
| Active runs (DEV-636) | 4 across all tasks of one client, each on one goroutine | The registered limits set no ceiling. It bounds run workers; the engine itself starts no goroutines. |
| Lifecycle deadline | 5 seconds by default, at most 30 seconds | Bounds the wait for reserved entry points, for the persistent-clear capability and, since DEV-592, for the trusted adapters during one admission. It reports incomplete work on expiry and cannot enlarge public attempt or cleanup bounds. Admission that exceeds it fails closed with `task_trust_unavailable`. |

The shell and admission perform no network activity: the production files import no network package, and the transitive dependency closure of the root package and of `internal/admission` contains none for the host build (DEV-594, [Hostile-Input Evidence](admission.md#hostile-input-evidence)). Runs reach the network only through the attempt-scoped dialer that the path seam supplies; no production path, runner, pending store, run ledger or capability is bound, so `Initialize` still fails closed in production. They establish no execution, result, bundle, proxy, production policy delivery, persistent-clearing, conformance or platform support, and admission and runs through the facade remain synthetic until reviewed production adapters exist.

## Trusted Initialization and Admission

`Initialize` accepts only the existing serializable `proxy`, `network_context`, and `result_linkage` configuration sections. Their nested schemas, defaults, validation, and privacy restrictions remain those of the pinned public contracts. Omitted linkage is disabled; omitted egress configuration permits access context only. The top-level object is closed and rejects unsupported sections. This implementation-local configuration container is distinct from Task JSON and is not a remotely delivered policy format.

Trusted local policy, including its target registrations and trusted issuer keys, policy generation, trusted clock, replay persistence, normalized platform capabilities, and protected storage are supplied by Barometer-owned reviewed adapters, not by this JSON. Host-selected configuration cannot enlarge adapter-installed authority. Missing required capabilities fail initialization or admission closed; linkage-only storage failure follows the public omission rule and never triggers ordinary-file secret storage.

<!-- requirement-ref: REQ-TRUST-0042 -->
<!-- requirement-ref: REQ-TRUST-0045 -->

`sourceRef` is a transient lookup reference to an adapter-installed delivery record. It is not an authentication assertion, issuer identifier, verification key, signature, policy document, or participant identifier. There is no public registration or minting method for these records.

For embedded origin, `taskJSON` is the exact serialized task. For remote origin, `taskJSON` is the exact signed remote task package of [REQ-TRUST-0048](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0048), and Barometer verifies it against the trusted issuer keys of the active local policy (DEV-612, [Signed Remote Task Packages](admission.md#signed-remote-task-packages)). The method signature is unchanged; only the meaning of the bytes presented for remote origin changed, and the parameter name is not part of the Go API. The host is a transport for remote origin and does not perform or assert authentication ([REQ-TRUST-0042](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0042)). The reviewed delivery path must:

1. For embedded origin, establish reviewed embedded origin and associate the exact UTF-8 task bytes with the issuer and policy generation. For remote origin, associate the exact received package bytes with remote origin and the policy generation only; it supplies no issuer, issuance time, authentication flag, key or verdict.
2. Install the record through an implementation-internal, instance-owned seam with bounded lifetime and capacity.
3. Resolve and compare those exact bytes during `AdmitTask`; reject missing, expired, foreign-instance, stale-policy, or mismatched records before any measurement activity. A remote record is single-use: each presentation of a package needs its own reference, and no digest of the package outlives the admission that presented it.

The native wrapper resolves delivery context internally; Flutter is not asked to supply `sourceRef` or mark arbitrary task bytes authenticated. Ordinary Go callers can present an existing reference, but cannot create authority by choosing its text. Public JSON, labels, endpoint responses, and collector replies cannot populate this registry. References never enter measurement traffic, results, logs, or persistent identifiers. They expire and are invalidated on clearing, close, and policy replacement.

The future adapter must keep registration atomic with respect to exact bytes and policy generation. Random or unguessable references alone are insufficient. Direct Go hosts requiring a delivery integration must use a separately reviewed adapter implementation; this design does not pretend such an adapter currently exists, and no binding accepts a verification key, a signature verdict or an authentication flag ([REQ-TRUST-0045](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0045)). Compromise of a reviewed host/adapter remains the public threat model's residual risk.

`AdmitTask` verifies a remote package before parsing its task, enforces the complete-task byte limit before parsing, preserves exact-byte replay identity, validates the entire task, and returns a local opaque task handle only on acceptance. Admission never starts a measurement or performs egress lookup. Exact redelivery resolves to the same admitted task without allocating another run or extending expiry. Conflicting content, unavailable trust, rollback, unsafe targets, or excessive limits fail atomically using the public admission classes.

## Execution and Cancellation

`StartRun` requests one due run of an admitted task. It checks schedule, expiry, consent/local policy, capability availability, and remaining run budget before reserving a run atomically. The host controls wakeups; the core does not start future repeated runs merely because a task was admitted. Concurrent duplicate starts for the same due occurrence return the same run handle and cannot allocate extra runs. Calls outside an eligible occurrence return a bounded operational error.

An accepted start returns promptly after reserving bounded work. Attempt-level execution includes the specified context/path snapshot, deadline, resource accounting, and permitted retries. Retries belong to the same run and remain constrained by the public schedule; protocol outcomes cannot become retry reasons. No unbounded worker creation or queue is allowed.

`RunState` returns an API-local coarse value: `active` while work or cleanup remains, and `terminal` after the run is finished. These polling values are not Measurement Result execution states or censorship classifications. Attempt results retain their canonical execution state, reason, and protocol observations. A run can be terminal with failed, incomplete, or cancelled attempts.

`CancelRun` prevents further attempts for that run, cancels active work, and waits for bounded cleanup. Repeated cancellation of a retained terminal run succeeds without modifying completed results. A cancellation racing with completion preserves whichever terminal transition already committed. Unknown or invalidated handles return `handle_unavailable`. A cleanup failure returns an explicit coarse error and keeps the affected scope quiesced; success must not be reported while network work continues.

Task and run handles are local control references, distinct from `task_id`, `run_id`, and `bundle_id`. They convey no provenance, are instance/generation scoped, never enter measurement traffic or results, and cannot select participants. Their tables and terminal-record retention must be finite under reviewed local policy; exhaustion returns `capacity_exceeded`, never silent eviction of active/replay state. Native integration must not persist or reuse them across reinitialization.

## Results and Delivery

`NextBundle` is non-blocking polling. It returns the oldest unacknowledged, unexpired canonical Result Bundle as UTF-8 JSON, or the empty string when none is ready. It does not return a custom result envelope, fetch from a collector, upload, or acknowledge data. Admission errors and local status values are never encoded as Measurement Results.

The core builds bundles from completed/terminated attempts without requiring a collector. Bundles obey canonical membership, result validation, identifiers, optional linkage, and byte/count limits. Repeated retrieval returns the same immutable persisted bundle bytes and identifier until acknowledgement or expiry. New results cannot be appended to a previously returned logical bundle. The host retains a copy only for its reviewed delivery purpose and must not log linkage material.

`AcknowledgeBundle` accepts the identifier of a locally pending bundle only after the host establishes durable delivery acknowledgement. It processes deletion immediately; deletion failure is an operational error, not successful acknowledgement. Unknown or already-deleted identifiers return `bundle_unavailable`. The host cannot supply or change a generated bundle identifier, results, or linkage through this method.

Canonical bounds include at most 16 results and 262,144 encoded bytes per bundle, 32 pending bundles and 4,194,304 aggregate pending bytes, and at most 86,400 seconds of unacknowledged retention. Expired bundles are deleted without late upload. Retention remains enforced on state access and initialization after suspension/restart, independently of collector availability. `Close` cannot extend retention. Backpressure must prevent additional execution before pending-output capacity is exhausted; no silent discard of completed results is defined here.

## Lifecycle, Proxy Changes, and Clearing

| State | Allowed behavior |
|---|---|
| Uninitialized | `Initialize`, `ClearMeasurementData`, and `Close`; measurement/result/proxy operations fail closed. |
| Initialized | Admission, execution, polling, acknowledgement, proxy changes, clearing, and close under current immutable policy. |
| Transitioning or clearing incomplete | Block new admission, starts, bundle construction, retrieval, and acknowledgement; allow recovery clearing or close. Never upload. |
| Closed | Repeated close and recovery clearing remain safe; all other calls fail with `client_closed`. Create a new instance to initialize again. |

Calls are safe to invoke concurrently. Lifecycle transitions serialize with admission, run reservation, bundle access, and each other. Operations losing that race return a bounded state error; stale handles cannot reference new-generation work. The implementation must not hold an uninterruptible lock across platform I/O. Reviewed adapter/local policy supplies finite storage/lifecycle deadlines; expiry reports incomplete work and keeps the affected scope blocked. The public attempt/cleanup bounds still control network activity and cannot be enlarged by those deadlines.

Calling `Initialize` on an active instance returns `already_initialized`; configuration and trusted policy do not mutate in place. Policy replacement requires quiescing the old owner, validating the complete authenticated monotonic policy, and initializing a replacement before new-policy work runs. Failure leaves the old complete policy active where safe or leaves the core uninitialized; no partial activation or rollback is permitted. Reinitialization does not reset replay records, schedule accounting, or retention; an explicit privacy clear does.

`SetProxy` and `ClearProxy` preserve the public dynamic-proxy lifecycle: validate replacement before mutation, block old-generation attempts, cancel and clean them up, atomically replace/clear, and invalidate affected context caches. The host keeps its previous proxy alive until successful return. Affected attempts record `proxy_reconfigured` and never continue through a replacement endpoint or direct fallback. No proxy credentials or native proxy object cross this API.

`ClearMeasurementData` implements the complete public clear-data contract: block admission/bundling; cancel work with bounded cleanup; delete bundles, replay/admitted-task state, secrets/keys, linkage epochs, reinstall state, and caches; invalidate handles and delivery references; and clear in-memory sensitive material where supported. It performs no upload or network request and creates no replacement secret. Success leaves the instance uninitialized (or closed if previously closed), requiring a subsequent explicit initialization for new work. Failure returns `clearing_incomplete` and keeps measurement/result activity blocked until recovery succeeds. Repeated clearing, including before initialization, reaches the same Barometer-owned storage scope through the private adapter; it must not report success merely because no in-memory client state exists.

`Close` stops new work, cancels active work, waits for bounded cleanup, invalidates local handles/references and context caches, and releases runtime ownership. It does not substitute for privacy clearing or delete still-valid replay/pending state. Result/retention processing resumes through a new initialized owner; overdue records are purged before use. Failed close remains quiesced and retryable; only completed close enters the closed state. The host connects user-visible deletion promises to `ClearMeasurementData`, not `Close`, and must remove any host-owned copies separately.

## Errors and Bounds

Return errors contain only a fixed operational code, never task bytes, target values, source references, handles, keys, tokens, raw platform errors, or free-form payloads. Public admission classes remain those of Task Trust; API-local classes include `not_initialized`, `already_initialized`, `client_closed`, `transition_in_progress`, `handle_unavailable`, `bundle_unavailable`, `run_not_due`, `capacity_exceeded`, `cleanup_incomplete`, and `clearing_incomplete`. They are local diagnostics and introduce no telemetry/result fields. The unexported policy-replacement seam additionally returns `policy_rejected`, which no exported method returns. Programming misuse, malformed input, or unavailable capabilities must not panic across a language boundary.

All allocations, input parsing, outstanding handles/delivery records, run workers, result queues, and platform operations need finite enforced limits. Task, attempt, result, bundle, replay, and context maxima come from the [registered limits](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/library/registered-limits.md). Additional facade/configuration/adapter capacities and deadlines must be selected in reviewed implementation work and cannot expand those maxima. No runtime readiness is claimed while those implementation inputs remain unspecified.

## Validation and Implementation Gates

Run `python3 scripts/check-public-api.py android` and `python3 scripts/check-public-api.py apple` on hosts with the prerequisites in [Mobile Compilation Smoke Checks](mobile-smoke.md). The helper extracts this document's exact signatures, gives each fixture function a panic body, and runs the existing pinned native build/skip-export check in a temporary module. It adds no production Go source or dependency. A passing result means only those declarations bind and compile for the selected target.

Before runtime implementation, separately validate trust adapter byte/generation binding; forged/missing source rejection; atomic replay and admission; concurrent due-run reservation; bounded cancellation/proxy transitions; polling/retransmission/acknowledgement; suspension/expiry/backpressure; failed/interrupted clearing; zero-value clients; and instance/handle invalidation. These are required future cases, not executed conformance evidence in this task. Complete native storage, lifecycle, and host tests before any platform-support claim.

The [duplicate-member adoption record](specification-issues.md#go-spec-issue-001-duplicate-json-member-names-outside-labels) now identifies the merged public clarification, shared decoder and Foundation model adoption. It remains open for applicable runtime/admission/producer/binding evidence; this API design does not implement those behaviors. The original design task changed only Go design/tooling documentation and fixture validation. Any later deviation from a public contract follows the existing specification-issue workflow before implementation.

## DEV-476 Validation Record

**Validation date:** October 13, 2026
**Scope:** Task review worktree based on Go commit `5dfb51f80f9a2e0582b7d98688ebbc3d2b7ed23b`; documentation and synthetic fixture, not runtime conformance.

| Check | Result and scope |
|---|---|
| `make check` | Passed with Go 1.26.0 on macOS ARM64: formatting, host build, unit/race tests, static analysis, module integrity, diff whitespace, and pure-Go Android/iOS ARM64 compilation. Existing tests exercise the scaffold, not the documented API semantics. |
| `python3 scripts/check-public-api.py android` | Passed: documented signatures generated/compiled into a temporary Android ARM64 AAR, with no skipped exports; JDK 17.0.18, SDK platform 35, NDK 28.2.13676358, API 21 compiler input. |
| `python3 scripts/check-public-api.py apple` | Passed: documented signatures generated/compiled into a temporary iOS ARM64 XCFramework, with no skipped exports; Xcode 27.0 (27A266a), deployment target 13.0 compiler input. |
| Binding tool revision | Both native checks used `golang.org/x/mobile v0.0.0-20260821190718-4776eadac327`, pinned in the existing tooling module. |
| Local documentation links and helper syntax | Passed; no new production dependency or Go export. |
| `git diff --check` | Passed. |

The Apple check initially could not write the compiler's module cache inside the restricted environment; it passed when rerun with cache-write permission. Temporary native artifacts were removed by the existing smoke helper. Neither fixture was executed on a device. Trusted delivery/native storage adapters, selected extra facade capacities/deadlines, runtime behavior, conformance, and device/lifecycle validation remain implementation gates.

## DEV-600 Validation Record

**Validation date:** October 20, 2026
**Scope:** Lifecycle shell on the DEV-600 task branch based on Go commit `ee135ae4d371db9b666bdaa6583e936f146e22d5`; synthetic tests only, not runtime conformance.

| Check | Result and scope |
|---|---|
| `make check` | Passed with Go 1.26.0 on macOS ARM64, including race tests and pure-Go Android/iOS ARM64 compilation. |
| Root lifecycle tests | Passed, also repeated under the race detector. They cover the state × method transition table, nil and zero-value clients, fail-closed initialization for each missing adapter and rejected configuration, concurrent and mixed lifecycle calls, waiting for and expiry of reserved entry points, double close and calls during a transition, handle capacity at limit −1, limit and limit +1 for both tables, stale handles across reinitialization and instances, handle opacity and redaction, panic recovery, and canary checks on error content. |
| `TestPublicSurfaceMatchesDocumentedSignatures` | Passed; a temporary extra exported method was rejected. |
| `TestRootPackageImportsNoNetworkPackages` | Passed. This is an import check; indirect activity through adapters remains a review rule. |
| `python3 scripts/check-public-api.py android` and `apple` | Passed: the documented signature fixture binds for Android ARM64 (JDK 17.0.18, NDK 28.2.13676358) and iOS ARM64. |
| `make mobile-smoke-android` and `make mobile-smoke-apple` | Passed: the production root package, including `Client`, binds with gomobile for both targets. |
| `make traceability` | Passed against the unchanged manifests before the evidence rebaseline, and again after rebaselining the cross-cutting manifest to Go `73faaf5`, where four `UT.root.*` cases make REQ-LINKAGE-0023 and REQ-LINKAGE-0024 `partial`. |
| `git diff --check` | Passed. |

No adapter is wired in production, so no binding or device test exercised a successful initialization. Persistent clearing, admission and the other documented methods remain pending.

## DEV-592 Validation Record

**Validation date:** October 20, 2026
**Scope:** `AdmitTask` on the DEV-592 task branch based on Go commit `766d1328f33cec9c6ca6a8604e720d1dbb133c2c`; synthetic tests only, not runtime conformance.

| Check | Result and scope |
|---|---|
| `make check` | Passed with Go 1.26.0 on macOS ARM64, including race tests and pure-Go Android/iOS ARM64 compilation. |
| Root admission tests | Passed, and the admission and lifecycle tests passed 50 repeated runs under the race detector. Synthetic adapters: a SYN-01 registry entry with one target registration, a generation-2 local policy for an embedded and a remote issuer, the manual test clock, an admission delivery store, the synthetic remote verifier and the test replay store. They cover acceptance for both origins; exact redelivery through the same and a new reference returning the same handle with unchanged replay record and expiry; `task_id_conflict` for changed bytes; atomic rejection for an unknown or mismatched reference, a task one byte over the complete-task limit, invalid schema, expiry, an unregistered test, a numeric limit, an unauthorized issuer, unavailable replay state, a non-durable commit and policy rollback; lifecycle refusals for nil, zero-value, uninitialized, incompletely cleared, transitioning and closed clients; stale references after clearing and store-level policy replacement and redelivery to a new generation; task-table capacity at limit −1, limit and limit +1 with reclamation of expired tasks only; panic recovery; error-content canaries; concurrent identical and distinct deliveries; admission racing close; and cancellation of the generation context on close. An unused recording transport stayed at zero calls. |
| `TestPublicSurfaceMatchesDocumentedSignatures` | Passed with `AdmitTask` added to the implemented list; it failed before the list was updated, showing the new export is gated. |
| `python3 scripts/check-public-api.py android` and `apple` | Passed: the documented signature fixture binds for Android ARM64 (JDK 17.0.18, NDK 28.2.13676358) and iOS ARM64. |
| `make mobile-smoke-android` and `make mobile-smoke-apple` | Passed: the production root package, including `Client.AdmitTask`, binds with gomobile for both targets. |
| `make traceability` | Passed against the unchanged manifests before the evidence rebaseline, and again after rebaselining the cross-cutting manifest to Go `7b098a6`, where four `UT.root.*` cases add facade-level aspects to REQ-TRUST-0017, REQ-TRUST-0024, REQ-TRUST-0035, REQ-TRUST-0037 and REQ-TRUST-0038. |
| `git diff --check` | Passed. |

No adapter is wired in production, so no binding or device test exercised a successful admission. Proxy compatibility at admission, execution and the other documented methods remain pending.

## DEV-593 Validation Record

**Validation date:** October 20, 2026
**Scope:** Policy activation and replacement by reinitialization on the DEV-593 task branch based on Go commit `1122faa64aa2d22ecaa0443b4f2ecf50d91c3c84`; synthetic tests only, not runtime conformance.

| Check | Result and scope |
|---|---|
| `make check` | Passed with Go 1.26.0 on macOS ARM64, including race tests and pure-Go Android/iOS ARM64 compilation. |
| `TestCcTrustLocalPolicyLifecycle` and `TestPolicyReplacementConcurrency` | Passed, and with the admission and lifecycle tests passed 50 repeated runs under the race detector. Synthetic adapters as in the DEV-592 record. They cover: a source check that only the replacement seam assigns the policy, adapters or admitter; unchanged policy after admission; refusal for nil, uninitialized, transitioning and closed clients; refusal with the previous policy, handles and delivery references still active for an equal and a lower generation, an incomplete policy, retained replay acceptances above the new generation, unavailable replay generations and a panicking clock; an uninitialized client with no policy after a panic while cancelling and after the delivery store refuses the new generation, with `task_policy_unavailable` after reinitialization; a blocked client after reserved work outlives the deadline, recovered by close; cancellation of in-flight admission with no replay state; a successful replacement that removes the registration, with cancelled context, invalid handles and delivery records, `target_not_registered` and `policy_rollback`, and a later replacement where redelivery resolves to the retained acceptance; initialization below the highest activated generation failing closed after clearing; policy-shaped labels and unknown task members leaving the policy unchanged; and concurrent replacements and admissions. Mutating the generation, floor, delivery-invalidation, retained-rollback, cancellation and failure-clearing checks each made the tests fail. An unused recording transport stayed at zero calls. |
| `make mobile-smoke-android` and `make mobile-smoke-apple` | Passed: the production root package binds with gomobile for Android ARM64 (JDK 17.0.18, NDK 28.2.13676358) and iOS ARM64. No exported signature changed, so the signature fixture check was not rerun. |
| `make traceability` | Passed against the unchanged manifests before the evidence rebaseline, and again after rebaselining the cross-cutting manifest to Go `954b7d7`, where the executed `CC-TRUST-LOCAL-POLICY-LIFECYCLE` case makes REQ-TRUST-0029 to REQ-TRUST-0032 `partial` and the new `CC-TRUST-LOCAL-POLICY-LIFECYCLE-RUNTIME` case stays planned. |
| `git diff --check` | Passed. |

No adapter is wired in production, so no binding or device test exercised a replacement. Host authentication of policy updates (DEV-455), cancellation of runs (DEV-561), Network Context cache invalidation (DEV-578), collector, target, proxy and context endpoint paths, and a durable generation floor remain pending.

## DEV-594 Validation Record

**Validation date:** October 20, 2026
**Scope:** Admission evidence closeout on the DEV-594 task branch based on Go commit `3bd604bcf54af6c7533831616cce9795a9bb230c`; synthetic tests only, not runtime conformance. No production code or exported signature changes.

| Check | Result and scope |
|---|---|
| `make check` | Passed with Go 1.26.0 on macOS ARM64, including race tests and pure-Go Android/iOS ARM64 compilation. |
| Root and admission tests | Passed. `TestAdmitTaskStrictJSON` and `TestAdmitTaskHostileSources` reject strict JSON violations and hostile sources through `AdmitTask` for both reviewed origins with exactly one public admission class and no handle, replay record or run; `TestCcTrustAcceptedTaskOrigins` executes its family case; `TestGateStepClasses` gains strict JSON rows. |
| `TestAdmissionHasNoNetworkDependency` | Passed. The host-build dependency closure of the root package and of `internal/admission` contains no network, TLS, process-execution or plugin package. Adding a `net/http` import to a root production file made it fail. |
| `FuzzAdmitTask` and `FuzzGate` | Passed finite 120-second runs, recorded in the [admission design](admission.md#fuzz-validation-record). CI runs only the seed corpora. |
| `make traceability` | Passed against the unchanged manifests before the evidence rebaseline, and again after rebaselining the cross-cutting manifest to Go `09e2141`, where `CC-TRUST-ACCEPTED-TASK-ORIGINS` is executed and REQ-TRUST-0005 becomes `partial`. |
| `git diff --check` | Passed. |

No adapter is wired in production, so no binding or device test exercised admission. Production test codecs (DEV-454), binding and host delivery of packages (DEV-455), execution, results and the remaining planned admission families remain pending; remote package verification (DEV-612) and key or issuer removal by policy replacement (DEV-593, DEV-612) are implemented with synthetic evidence, and the remote-delivery gate (DEV-595) is lifted ([Remote-Delivery Gate](admission.md#remote-delivery-gate)).

## DEV-612 Validation Record

**Validation date:** October 20, 2026
**Scope:** Library verification of signed remote task packages at public `e97502747c917a16bc2b71337535ff972bbd7d64`, implemented at Go commit `e663afeb0f9a6f5cc57c914b8f13ff75bb71e216` on the DEV-612 task branch; synthetic tests only, not runtime conformance. `Client.AdmitTask` keeps its signature; for remote origin its presented bytes are the signed package. No identifier is exported.

| Check | Result and scope |
|---|---|
| `make check` | Passed with Go 1.26.0 on macOS ARM64, including race tests and pure-Go Android/iOS ARM64 compilation. |
| Root, admission and model tests | Passed: 229 top-level tests from a clean `git archive` export, recorded in the [evidence report](../traceability/evidence/go-unit-e663afe.md). Synthetic Ed25519 keys from fixed public test seeds; no deployment key. `TestAdmitTaskAcceptsReviewedOrigins`, `TestAdmitTaskStrictJSON` and `TestAdmitTaskHostileSources` present remote tasks as signed packages through the facade; `TestCcTrustKeyRotationByPolicyReplacement` removes a key and then an issuer by policy replacement. |
| `FuzzSignedPackage` and `FuzzAdmitTask` | Passed finite 60-second runs, recorded in the [admission design](admission.md#fuzz-validation-record). CI runs only the seed corpora. |
| `make traceability` | Passed after rebaselining the cross-cutting manifest to public `e975027` and Go `e663afe`, with the DEV-594 mapping preserved as historical. |
| Public requirement-reference checker | Passed against the public checkout at `e975027`; no current reference cites a retired ID. |
| `git diff --check` | Passed. |

No adapter is wired in production, so no binding, native or device test exercised the package path. The cryptographic, safety and host integration reviews were performed by the project maintainer in DEV-615 with no findings and the remote-delivery gate (DEV-595) is lifted ([Remote-Delivery Gate](admission.md#remote-delivery-gate)); binding and host adoption (DEV-455) remain open, so hosts cannot yet pass packages.

## DEV-566 Validation Record

**Validation date:** October 21, 2026
**Scope:** Strict JSON evidence for the existing Go consumers ([GO-SPEC-ISSUE-001](specification-issues.md#existing-consumer-adoption-addendum--october-21-2026)) at public `e97502747c917a16bc2b71337535ff972bbd7d64`, recorded at Go commit `baa3adaa74bd2e3474bc518e257567cbf20b2c00` on the DEV-566 task branch; tests and documentation only, no behavior or signature change, not runtime conformance.

| Check | Result and scope |
|---|---|
| `make check` | Passed with Go 1.26.0 on macOS ARM64, including race tests and pure-Go Android/iOS ARM64 compilation. |
| Root and model tests | Passed: 233 top-level tests from a clean `git archive` export, recorded in the [evidence report](../traceability/evidence/go-unit-baa3ada.md). `TestAdmitTaskSignedPackageStrictJSON` presents hostile signed package envelopes through `Client.AdmitTask`, and `TestInitializeStrictJSON` presents hostile host configuration through `Client.Initialize` with synthetic approvals and a counting linkage key store. |
| `FuzzInitialize` and `FuzzAdmitTask` | Passed finite 60-second runs, recorded in the [configuration models](configuration-models.md#synthetic-evidence) and the [admission design](admission.md#fuzz-validation-record). CI runs only the seed corpora. |
| `make traceability` | Passed after rebaselining the cross-cutting manifest to Go `baa3ada`, with the DEV-612 mapping preserved as historical. |
| Public requirement-reference checker | Passed against the public checkout. |
| `git diff --check` | Passed. |

No adapter is wired in production, so no binding, native or device test exercised these entry points. Registered test schemas and measurement runtime consumers (DEV-622) and binding and host consumers (DEV-623) remain open.

## DEV-636 Validation Record

**Validation date:** October 22, 2026
**Scope:** `StartRun` and `RunState` on the DEV-636 task branch based on Go commit `3707cf0e`; synthetic tests only, not runtime conformance. Interval repetition follows the interim rule of [GO-SPEC-ISSUE-011](specification-issues.md#go-spec-issue-011-due-times-and-catch-up-for-interval-repetition), which awaits public disposition.

| Check | Result and scope |
|---|---|
| `make check` | Passed with Go 1.26.0 on macOS ARM64, including race tests and pure-Go Android/iOS ARM64 compilation. |
| Root run tests (`runs_test.go`) | Passed, and the run, consent and lifecycle tests passed 20 repeated runs under the race detector. Synthetic adapters: a TCP-01-identified registry entry with a scripted runner that performs no I/O, the generation-2 synthetic policy, a run clock with a monotonic reading at the admission clock's instants, fixed jitter, the test pending-output store, the test run ledger, a fixed direct-path seam and scripted consent, device and time-zone capabilities. They cover a run recorded in the ledger before it runs, `active` then `terminal`, and one completed result; 16 concurrent duplicate starts sharing one handle and one run, and the same handle after the run is terminal; `starts_at` before, at and after, `not_after` before and at, day and midnight-crossing window boundaries and a window without a time zone; interval due times with jitter, a skipped occurrence and the end of the schedule; the run budget at limit −1, limit and limit +1; valid, absent, ended, unavailable, failing and missing consent, no query without a requirement, and consent withdrawn before a retry; battery and charging boundaries, unknown state, errors and a missing capability; unsatisfiable network conditions that never reach the path seam; exclusive and shared concurrency and the active-run limit +1 (`capacity_exceeded`); pending-output capacity including the active runs' results and a capacity failure; ledger lookup and write failures and an uncertain reservation that loses the occurrence; unknown, empty, other-instance and stale handles, nil, uninitialized and closed clients, a missing runner, path failures and panics, and a ledger panic after the handle is published; accounting that survives a new instance with redelivery; close and clearing with an active run, which cancel it, wait for it and invalidate its handle; concurrent starts racing close; and error-content and formatting canaries. |
| Execution tests | Passed: `occurrence_test.go`, `gates_test.go` and `TestPreconditionsAreRecheckedBeforeEachAttempt` ([Synthetic Evidence](execution-models.md#synthetic-evidence-and-remaining-work)). |
| `TestPublicSurfaceMatchesDocumentedSignatures` | Passed with `StartRun` and `RunState` added to the implemented list. |
| `python3 scripts/check-public-api.py android` and `apple` | Passed: the documented signature fixture binds for Android ARM64 (JDK 17.0.18, NDK 29.0.14206865) and iOS ARM64. |
| `make mobile-smoke-android` and `make mobile-smoke-apple` | Passed: the production root package, including `Client.StartRun` and `Client.RunState`, binds with gomobile for both targets. |
| `make traceability` | Before the rebaseline, the manifests and case plans were valid and the requirement-reference check reported the new `requirement-ref` comments as not yet linked. Passed after rebaselining the cross-cutting manifest to Go `daeb718`, where 13 `UT.*` cases make REQ-SCHEDULE-0007, REQ-SCHEDULE-0009, REQ-SCHEDULE-0020, REQ-SCHEDULE-0021, REQ-SCHEDULE-0030, REQ-SCHEDULE-0034, REQ-EXECUTION-0027, REQ-EXECUTION-0028 and REQ-SAFETY-0020 `partial` and GO-SPEC-ISSUE-011 is linked as open. |
| `git diff --check` | Passed. |

No adapter is wired in production, so no binding or device test exercised a run. The production pending store (DEV-563), run ledger, its retention and clearing (DEV-577, DEV-564), path selection and network conditions (DEV-578), native consent, device and time-zone capabilities (DEV-455), cancellation semantics and deadlines (DEV-561) and budgets (DEV-562) remain pending.

## DEV-561 Validation Record

**Validation date:** October 22, 2026
**Scope:** `CancelRun`, attempt deadlines, the attempt scope and lifecycle cancellation on the DEV-561 task branch based on Go commit `861a353b`; synthetic tests on the host platform with loopback sockets only, not runtime or device conformance. Execution states follow the [GO-SPEC-ISSUE-004 disposition](specification-issues.md#go-spec-issue-004-boundary-between-not_run-and-incomplete) at public `84f52f8`.

| Check | Result and scope |
|---|---|
| `make check` | Passed with Go 1.26.0 on macOS ARM64, including race tests and pure-Go Android/iOS ARM64 compilation. |
| Root cancellation tests (`cancel_run_test.go`) | Passed, and together with the DEV-636 lifecycle and consent tests passed 20 repeated runs under the race detector, with the DEV-636 synthetic adapters. They cover `CancelRun` on an active run with a permitted retry: one `incomplete` `cancelled_by_host` result, no retry, `terminal` on return and unchanged run accounting; three repeated cancellations and a cancellation of a completed run that leave results byte-identical; empty, unknown, task, altered and earlier-generation handles (`handle_unavailable`), uninitialized, closed, nil and zero clients; a runner that ignores cancellation, where `CancelRun` waits until one millisecond before the cleanup bound, returns `cleanup_incomplete` at it, leaves the run `active` with its exclusive reservation held, and `Close` returns `cleanup_incomplete` until the runner returns; the task's `max_duration_ms` of 1,000 ms ending the attempt at, not before, the deadline with `attempt_deadline_exceeded`; and close, clearing and policy replacement each recording the cancelled attempt as `cancelled_by_host`. |
| Execution tests | Passed, and the package passed 20 repeated runs under the race detector: `deadline_test.go`, `TestHostCancellation` and the updated engine tests ([Synthetic Evidence](execution-models.md#synthetic-evidence-and-remaining-work)). Socket release is observed on the server side of real loopback connections. |
| Result tests | Passed: `TestExecutionContextAndAttemptBinding` rejects Network Context in every `not_run` result for every registered reason (REQ-RESULT-0053). |
| `TestPublicSurfaceMatchesDocumentedSignatures` | Passed with `CancelRun` added to the implemented list. |
| `python3 scripts/check-public-api.py android` and `apple` | Passed: the documented signature fixture binds for Android ARM64 (JDK 17.0.18, NDK 29.0.14206865) and iOS ARM64. |
| `make mobile-smoke-android` and `make mobile-smoke-apple` | Passed: the production root package, including `Client.CancelRun`, binds with gomobile for both targets. |
| `make traceability` | Before the rebaseline, the manifests and case plans were valid and the requirement-reference check reported the new `requirement-ref` comments as not yet linked, and REQ-RESULT-0051 to REQ-RESULT-0053 as unknown at public `4743b63`. The rebaseline to public `84f52f8` follows the implementation commit. |
| `git diff --check` | Passed. |

No adapter is wired in production, so no binding or device test exercised a run or a cancellation. HTTP response bodies are not exercised: no HTTP runner exists (DEV-454). Platform suspension and native lifecycle cleanup (DEV-455), proxy-change cancellation (DEV-578), byte and connection budgets (DEV-562) and persistent clearing (DEV-564) remain pending.
