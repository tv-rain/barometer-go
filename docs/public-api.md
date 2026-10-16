# Public Go API Boundary

**Status:** Defined design; runtime implementation pending
**Task:** [DEV-476 — Define the public Go API boundary](https://app.clickup.com/t/86cbb8f0j)
**Specification baseline:** Draft `0.1.0` at `95dfeef1367f4c320bca7c7ae91b13ca7608b37f`

## Scope and Authority

This document defines the intended root-package facade for initialization, admission, host-triggered execution, cancellation, result retrieval, and measurement-data clearing. It is an implementation-local API design subordinate to the public contracts in the [Normative Source Inventory](normative-sources.md). It introduces no production exports, task/result schema change, measurement implementation, conformance claim, release, or platform-support claim.

The signatures below are checked in an isolated synthetic module. Fixture bodies deliberately panic if invoked; they validate compiler/binding compatibility only. Do not copy those bodies into production. Implementation of this facade requires admission, execution, persistence, and reviewed platform/trust adapters. No successful no-op or permissive fallback may substitute for those components.

## Responsibility Boundary

| Owner | Responsibilities |
|---|---|
| Host | Enablement; authenticating delivery and policy updates through a reviewed integration; scheduling calls within platform restrictions; operating the optional proxy; transporting returned bundles; reporting durable acknowledgement; connecting user-visible clearing; deployment review. |
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

## Trusted Initialization and Admission

`Initialize` accepts only the existing serializable `proxy`, `network_context`, and `result_linkage` configuration sections. Their nested schemas, defaults, validation, and privacy restrictions remain those of the pinned public contracts. Omitted linkage is disabled; omitted egress configuration permits access context only. The top-level object is closed and rejects unsupported sections. This implementation-local configuration container is distinct from Task JSON and is not a remotely delivered policy format.

Trusted local policy, target registrations, policy generation, delivery verifier, trusted clock, replay persistence, normalized platform capabilities, and protected storage are supplied by Barometer-owned reviewed adapters, not by this JSON. Host-selected configuration cannot enlarge adapter-installed authority. Missing required capabilities fail initialization or admission closed; linkage-only storage failure follows the public omission rule and never triggers ordinary-file secret storage.

`sourceRef` is a transient lookup reference to an adapter-installed delivery record. It is not an authentication assertion, issuer identifier, verification key, signature, policy document, or participant identifier. There is no public registration or minting method for these records. The reviewed delivery path must:

1. Establish reviewed embedded origin or authenticate remote issuer and integrity before creating a record.
2. Associate the exact UTF-8 task bytes with the trusted origin, issuer, generation, authentication outcome, and authenticated issuance time required by the public trust contract.
3. Install the record through an implementation-internal, instance-owned seam with bounded lifetime and capacity.
4. Resolve and compare those exact bytes during `AdmitTask`; reject missing, expired, foreign-instance, stale-policy, or mismatched records before any measurement activity.

The native wrapper resolves delivery context internally; Flutter is not asked to supply `sourceRef` or mark arbitrary task bytes authenticated. Ordinary Go callers can present an existing reference, but cannot create authority by choosing its text. Public JSON, labels, endpoint responses, and collector replies cannot populate this registry. References never enter measurement traffic, results, logs, or persistent identifiers. They expire and are invalidated on clearing, close, and policy replacement.

The future adapter must keep verification and registration atomic with respect to exact bytes and policy generation. Random or unguessable references alone are insufficient. Direct Go hosts requiring a delivery integration must use a separately reviewed adapter implementation; this design does not pretend such an adapter currently exists or expose a raw `authenticated: true` shortcut. Compromise of a reviewed host/adapter remains the public threat model's residual risk.

`AdmitTask` enforces the complete-task byte limit before parsing, preserves exact-byte replay identity, validates the entire task, and returns a local opaque task handle only on acceptance. Admission never starts a measurement or performs egress lookup. Exact redelivery resolves to the same admitted task without allocating another run or extending expiry. Conflicting content, unavailable trust, rollback, unsafe targets, or excessive limits fail atomically using the public admission classes.

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

Return errors contain only a fixed operational code, never task bytes, target values, source references, handles, keys, tokens, raw platform errors, or free-form payloads. Public admission classes remain those of Task Trust; API-local classes include `not_initialized`, `already_initialized`, `client_closed`, `transition_in_progress`, `handle_unavailable`, `bundle_unavailable`, `run_not_due`, `capacity_exceeded`, `cleanup_incomplete`, and `clearing_incomplete`. They are local diagnostics and introduce no telemetry/result fields. Programming misuse, malformed input, or unavailable capabilities must not panic across a language boundary.

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
