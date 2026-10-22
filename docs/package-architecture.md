# Package Architecture

**Status:** Accepted

**Specification baseline:** Barometer Draft `0.1.0`

## Purpose

This document defines the intended package ownership and dependency rules for the Barometer Go core. It guides later scaffolding and implementation; the presence of a package in this map is not evidence that the package, measurement, or platform is implemented, conformant, field-validated, or supported.

The design preserves the project dependency direction:

```text
public specification
        ↓
Barometer Go core
        ↓
native and Flutter binding
        ↓
host application
```

The host remains responsible for enablement, task delivery (for remote tasks, transport of signed packages that the Go core verifies), scheduling, deployment policy, optional proxy operation, result delivery, transparency, and deployment-specific review. The Go core remains responsible for admission, validation, bounded execution, registered measurement behavior, minimized results, cancellation, retention, and clearing.

## Design Constraints

- Keep the initial implementation surface small enough for a small team and mobile environments.
- Treat measurements as observations, not proof of censorship.
- Keep execution state, protocol observations, and derived interpretation separate.
- Keep the core backend-independent and host-orchestrated.
- Reject unknown or unsafe task behavior rather than silently rewriting, clamping, or partially executing it.
- Do not allow task-controlled data to assert provenance, issuer authority, local policy, target registration, proxy configuration, protected-storage capabilities, or participant identity.
- Use ordinary serializable values at the future Flutter boundary. Native objects, callbacks, storage handles, secrets, and Go implementation types do not cross that boundary.
- Prefer purpose-specific packages and interfaces. Do not add abstractions solely to mirror document sections or anticipated future features.

## Package Map

The following paths are the intended ownership map. A path should be created only when the corresponding implementation work begins.

```text
github.com/tv-rain/barometer-go          public Go facade and lifecycle entry points
└── internal
    ├── config                          immutable host configuration, validation and preparation
    ├── model
    │   ├── task                         admitted task and trusted-context domain values
    │   ├── observation                  bounded protocol and network observations
    │   └── result                       result, bundle, and linkage domain values
    ├── taskdecode                       deterministic decoding of serialized tasks
    ├── jsondecode                       shared strict JSON representation and schema decoding
    ├── validation                       structural, semantic, target, safety, and limit validation
    ├── admission                        trust, policy, replay, expiry, and admission decisions
    ├── execution                        attempt orchestration, cancellation, path snapshots, and lifecycle
    ├── measurement                      registered measurement runner contract
    │   ├── dns01
    │   ├── dns03
    │   ├── dns04
    │   ├── tcp01
    │   ├── tls01
    │   ├── http01
    │   ├── net01
    │   ├── net03
    │   └── net05
    ├── protocol
    │   ├── dns                          bounded DNS operations shared by registered tests
    │   ├── tcp                          bounded TCP connection operations
    │   ├── tls                          bounded authenticated TLS operations
    │   └── http                         bounded HTTP and HTTPS operations
    ├── registry                         reviewed test identifiers, versions, and runner composition
    ├── networkcontext                   acquisition, normalization, caching, and path consistency
    ├── result                           result normalization and minimized result construction
    ├── bundle                           bundle construction, size limits, and delivery state transitions
    ├── persistence                      bounded atomic persistence capabilities and clearing coordination
    ├── platform                         normalized platform capabilities required by the core
    └── testkit                          synthetic clocks, stores, transports, and fixtures for tests only
```

This map intentionally has no `pkg`, `common`, `shared`, `helpers`, or `utils` package. A repeated operation becomes a package only when it has one coherent owner and a stable reason to exist.

## Ownership Boundaries

### Initialization Configuration

`internal/config` owns the closed host-initialization schema, pure semantic validation, immutable snapshots and atomic proxy value replacement. It consumes `internal/jsondecode` and narrow clock/protected-key contracts in `internal/platform`, with deployment approvals supplied separately by reviewed internal integration. It does not import task decoding or orchestration, perform I/O or activate a runtime. Its preparation results establish no task provenance, local-policy authority, replay eligibility or execution permission. See [Initialization and Configuration Models](configuration-models.md) for local bounds, capability contracts, synthetic evidence and remaining runtime gates.

### Root `barometer` Package

The module root is the only initially intended public Go package. It owns:

- initialization and shutdown entry points;
- the small lifecycle facade used by a binding or direct Go integrator;
- serialized task admission and result retrieval entry points;
- cancellation and Barometer-owned data-clearing entry points;
- stable, coarse errors or status values that an integrator must handle.

It does not own measurement logic, JSON schema implementation, trust policy, persistence implementations, platform secrets, collector behavior, or censorship classification. Exporting additional Go packages or interfaces is a public API decision and requires explicit review.

The [Public Go API Boundary](public-api.md) defines the intended facade signatures and trusted adapter, lifecycle, and result-ownership contract. This architecture document fixes ownership and dependency direction; the API design does not export additional packages or create the planned internal package tree.

DEV-600 adds the lifecycle shell described in [Implemented Subset](public-api.md#implemented-subset): `NewClient`, `Initialize`, `Close` and `ClearMeasurementData`, the lifecycle state machine, the per-generation task and run handle tables and the fixed API-local error codes. The root package binds trusted adapters (clock, registry, local policy, delivery store, replay store, configuration approvals and optional linkage keys) through an unexported seam with no production caller, so initialization fails closed until reviewed native wiring exists (DEV-455). It imports `internal/config`, `internal/admission`, `internal/registry`, `internal/persistence`, `internal/platform` and `internal/model/task`, and no network package. Its tests import `internal/testkit`. DEV-592 adds `AdmitTask`, which reserves the current generation, runs `admission.Admitter.Admit` and publishes task handles indexed by replay identity. DEV-593 adds the unexported `replacePolicy` seam, which replaces the local policy only by quiescing the current generation and reinitializing a new one with a strictly higher policy generation ([Policy Replacement](public-api.md#implemented-subset)). DEV-594 adds hostile-input, strict JSON and fuzz tests through `AdmitTask` and no production code ([Hostile-Input Evidence](admission.md#hostile-input-evidence)). DEV-636 adds `StartRun` and `RunState`: the facade owns the per-generation run reservation state (accounting lookup, duplicate resolution, concurrency, the active-run limit and pending-output capacity), builds `execution.Plan` values from admitted tasks and launches `execution.Engine` on one goroutine per active run; it now also imports `internal/execution` and `internal/result`. The run inputs (engine clock, jitter, pending-output port, run ledger, an interim direct-path seam and optional consent, device and time-zone capabilities) are bound through the same unexported seam. Cancellation, bundle, proxy and persistent-clearing entry points remain pending.

### Domain Models

The packages under `internal/model` own in-process values after decoding and normalization:

- `model/task` owns task fields, schedules, execution constraints, safety declarations, trusted source context, local-policy inputs, and admitted-task values;
- `model/observation` owns bounded DNS, TCP, TLS, HTTP, network-state, and execution observations without interpretation verdicts;
- `model/result` owns normalized attempt results, partial-result markers, bundles, network-context attachment, and optional linkage metadata.

Models are dependency leaves. They contain validation-independent value definitions and small invariants needed to keep invalid states unrepresentable, but they do not perform I/O, scheduling, persistence, logging, or platform calls.

The [Observation and Network Context Models](observation-models.md) implementation in `model/observation` owns bounded DNS/TCP/TLS/HTTP values and minimized access/egress snapshots with non-serialized path/generation binding. It depends only on the standard library and the shared strict JSON leaf; no acquisition/cache owner is created. The model adopts the merged public DNS/TLS vocabularies and their internal review alignment; it performs no protocol I/O or error parsing. The planned `networkcontext` package will consume these values for runtime acquisition and caching rather than duplicate their domain invariants.

Serialized wire objects and admitted domain values must not be conflated. In particular, trusted source context, local policy, replay state, proxy configuration, and protected-storage capabilities are never populated from task-controlled fields.

The implemented [Trusted Context and Local Policy Models](trusted-policy-models.md) in `model/task` provide structurally validated source metadata, exact-byte delivery binding, immutable local/issuer/target policy values and revocation inputs. Constructors/getters are internal handoff surfaces; they do not authenticate, admit tasks, compute effective authority, activate policy or enforce revocation. No runtime admission or adapter package is created by this model work.

### Deterministic Task Decoding

The implemented `internal/jsondecode` package owns reusable strict JSON framing, Unicode, duplicate-name detection, exact numeric representation, schema/tuple selection, and bounded sanitized atomic decoding. It is a dependency leaf using only the standard library; task, result, bundle, configuration, and Network Context decoders may consume it without importing one another. It does not own any individual public model schema, domain normalization, admission, or execution. See [Shared Strict JSON Decoding](json-decoding.md) for the immutable adopted contract, implementation limits, and consumer responsibilities.

`internal/taskdecode` owns conversion from the serialized Draft `0.1.0` task representation into a candidate task value. It is responsible for:

- the complete serialized-size bound;
- deterministic JSON handling;
- rejection of unknown fields where required;
- exact field types and supported schema versions;
- distinguishing missing values from explicit zero values where the specification requires it;
- producing structured decoding errors without logging sensitive task content.

It does not decide issuer authority, target approval, replay eligibility, local-policy compatibility, or execution eligibility. Documentation examples are test fixtures, not a second schema.

### Validation

`internal/validation` owns pure validation of a candidate task against the public specification, registered measurement definition, registered limits, approved target information, and effective local safety policy. It covers:

- structural and cross-field rules;
- test-specific input and option validation;
- schedule, retry, execution, result-size, and composite-profile limits;
- target and network-path compatibility;
- risk and consent consistency;
- atomic rejection of an invalid `NET-05` profile before network activity.

Validation returns explicit failures and never clamps an excessive value, drops an unsupported step, substitutes a target, changes a network path, or enables a compatible subset of an invalid task.

### Admission

`internal/admission` owns the ordered decision to accept, idempotently recognize, reject, or defer a task. It combines decoding and validation with trusted state and is responsible for:

- authenticated source context supplied separately from the task;
- issuer authorization and policy generation;
- target registration and non-targeted-delivery rules;
- exact-redelivery idempotency and conflicting-identifier rejection;
- expiry, replay, rollback, and bounded validity checks;
- atomic persistence of accepted replay state before a task becomes executable;
- fail-closed behavior when required trust or persistence state is unavailable.

Admission does not execute network activity and cannot be bypassed by the root facade, a measurement runner, a binding, or the host application. `TestAdmissionHasNoNetworkDependency` checks that the transitive dependency closure of `internal/admission` and of the root package contains no network, TLS, process-execution or plugin package for the host build (DEV-594); the bypass rule remains a review rule.

DEV-585 creates the package with the delivery-record seam only, described in [Delivery-Record Seam](admission.md#delivery-record-seam). `admission.DeliveryStore` is an instance-owned, bounded store of adapter-installed `task.DeliveryRecord` values resolved by an opaque reference. It has an embedded-origin installer and, since DEV-612, a remote installer that binds only the exact signed package bytes; the `RemoteVerifier` seam is removed and admission itself verifies packages against the policy's trusted keys ([Signed Remote Task Packages](admission.md#signed-remote-task-packages)). The test-only `internal/testkit/delivery` package signs synthetic packages. DEV-586 adds `admission.Admitter`, whose `Gate` runs admission steps 1 to 6 and the input and options part of step 8 and returns the class of the first failing step, described in [Implemented Gates](admission.md#implemented-gates). DEV-587 binds the active `task.LocalPolicy` snapshot and adds issuer and policy resolution, path authority and the issuer and local risk bounds, carrying the intersected `Authority`, described in [Issuer Authorization](admission.md#issuer-authorization). DEV-591 adds `Admit`, which runs replay (step 10) and acceptance (step 11) against the `persistence.ReplayStore` contract, described in [Atomic Rejection and Replay Persistence](admission.md#atomic-rejection-and-replay-persistence). The root `Client.AdmitTask` wires `Admit` to the facade (DEV-592). DEV-599 adds registry proxy compatibility to step 8, described in [Proxy Compatibility](admission.md#proxy-compatibility).

### Execution

`internal/execution` owns deterministic execution attempts after admission. It is responsible for:

- schedule eligibility as directed by host orchestration;
- execution constraints and concurrency rules;
- deadlines, byte and connection budgets, cancellation, and bounded cleanup;
- stable direct or proxy path snapshots with no silent fallback;
- coordination of network-context acquisition and network-change detection;
- invoking exactly one registered runner or one admitted sequential profile;
- producing execution state separately from protocol observations;
- coordinating result construction and persistence transitions.

Execution does not authenticate task delivery, widen policy, invent retries, operate a collector, infer censorship, or own protocol-specific interpretation.

The implemented [Execution-State and Error Models](execution-models.md) provide a pure immutable attempt transition value in `internal/execution` and finite admission-error/reason values in `internal/model/task`. The task model remains a standard-library-only dependency leaf.

DEV-635 adds the internal [run orchestration engine](execution-models.md#run-orchestration). `Engine.Run` runs one run of an admitted task: it resolves exactly one registered runner by exact test version, executes up to two attempts sequentially, retries only as the schedule's retry object and `validation.RetryEligible` allow, takes backoff delays from an injected clock and local jitter, and passes each terminal attempt through `internal/result` to a narrow `execution.Output` port. Network Context and the attempt-scoped dialer come from trusted plan seams (`ContextSource`, `DialerSource`) until the Network paths and context work replaces them. The engine does not coordinate deadlines or cancellation beyond passing the host context (DEV-561), enforce byte or connection budgets (DEV-562), handle paths or proxies (DEV-578) or persist results (DEV-563).

DEV-636 adds [run reservation](execution-models.md#run-reservation) to `internal/execution`: the run-accounting functions for the interim due-time rule of GO-SPEC-ISSUE-011 over `persistence.RunRecord`, and `execution.Gates`, the fail-closed attempt-start check of network conditions, device state and consent through `internal/platform`. The engine evaluates a plan's `Preconditions` before every attempt and the `Output` port reports `Capacity`. `Client.StartRun` launches the engine; concurrency across runs stays with the facade.

### Registered Measurements and Protocol Primitives

`internal/measurement` owns the minimal runner contract shared by registered measurements. Each child package owns the behavior for exactly one registered test identifier and version.

The packages under `internal/protocol` own reusable bounded protocol operations. Registered measurement packages may depend on these primitives, domain models, and narrowly required platform capabilities. They must not depend on another registered measurement package.

`NET-05` composes reviewed primitives through its own admitted sequential plan. It must not call another test package as an opaque shortcut, execute steps concurrently, execute a compatible subset, or produce a local censorship verdict.

Protocol primitives return observations and budget usage. They do not construct final result envelopes, select retries, change paths, persist data, or classify censorship.

DEV-635 creates `internal/measurement` with the runner contract only. `Runner.Run` receives the admitted normalized parameters, the attempt context and an attempt-scoped `Dialer`, and returns a terminal outcome, its observations and the runner's `Usage` (application bytes sent and received, connections). `Conn` exposes only read, write, close and a deadline, with no address, socket or platform handle. The contract imports only task and observation models and standard-library packages without network access, so admission, which reads the registry, links no network code. A runner reports an execution failure with the fixed `ErrRunner`; a negative protocol outcome is an observation, not an error. No registered runner exists yet; the only runner is the test-only loopback runner in `internal/testkit/loopback`, which is not TCP-01 evidence.

### Registry

`internal/registry` is the composition point for reviewed test identifiers, supported specification versions, runner implementations, validation metadata, and proxy compatibility. It owns the mapping used by admission and execution.

Adding an identifier to the registry requires a corresponding public registered test, internal review alignment, implementation, and conformance evidence. Registry presence alone must not be presented as released or field-validated support.

DEV-584 implements the admission subset. `registry.New` builds an immutable registry from a bounded list of entries. Each entry pairs an exact `(test_id, test_version)` with its `taskdecode.TestCodec`, registered baseline risk, a target rule that reports the complete typed tuple of every destination or name its normalized parameters would contact (DEV-588), a proxy-compatibility rule over normalized parameters, and whether the test is the NET-05 composite or a permitted NET-05 step. Construction rejects the whole set with one fixed error when any entry is incomplete, duplicated, malformed or inconsistent with its codec. `Lookup` matches exact strings only, with no latest-version fallback, case folding or partial match, and is safe for concurrent reads. `Codecs` returns the same reviewed codec set for building the task decoder. There is no global registry or runtime registration method. DEV-635 adds runner composition: an entry may carry a `measurement.Runner`, and `Runner` returns it only for the exact test ID and version. A missing entry or an entry without a runner reports false, and execution then fails closed before any work; there is no default runner or version fallback. The runner stays optional at construction so admission-only entries remain valid. Only test-scoped synthetic entries exist; production entries and runners are added by later tasks. Admission combines the registered baseline risk with the declared and locally determined risk (DEV-589, see [Effective Risk and Consent](admission.md#effective-risk-and-consent)).

### Network Context

`internal/networkcontext` owns normalized access and egress context, cache rules, primary and fallback acquisition, path consistency, coarse failure states, and privacy filtering. It consumes platform and network capabilities but does not expose raw platform objects or unrestricted endpoint responses.

Network Context cannot receive bundle identifiers, linkage material, participant identifiers, task provenance secrets, or collector credentials. Its observations cannot by themselves establish censorship.

### Result Construction

`internal/result` owns conversion of execution state and bounded observations into the Draft `0.1.0` measurement-result representation. It is responsible for:

- required identifiers and version fields;
- timing and partial-result semantics;
- normalized execution state kept separate from measurement observations;
- attachment of minimized Network Context from the selected path;
- test-specific output validation;
- serialized result-size enforcement;
- exclusion of secrets, raw proxy configuration, and unrestricted diagnostic data.

It does not deliver results, retain transport metadata, infer censorship, or modify the observations it receives to create a preferred interpretation.

The [Measurement Result models](result-models.md) implement detached typed values in `model/result` and the bounded producer/import serialization boundary in `internal/result`. That owner depends only on domain models, shared strict JSON and the standard library; it does not import execution or protocol implementations. The initial Output schemas validate retained evidence against a copied local task/output/profile binding. Runtime admission, actual execution, bundle handling and platform support remain pending.

### Bundle Construction

`internal/bundle` owns bounded grouping of normalized results, bundle identifiers, optional linkage metadata, serialized bundle-size enforcement, and local delivery state transitions. Result delivery remains replaceable and host-controlled.

The package must operate without a mandatory collector. It must not accept account identity, source IP metadata, task selection state, or arbitrary stable identifiers as linkage inputs.

The [Result Bundle and linkage models](bundle-models.md) implement detached projections in `model/result` and the bounded producer in `internal/bundle`. The producer composes validated immutable `internal/result.Value` members and copies policy from prepared `internal/config.Initialization`; it depends on those model boundaries and shared decoder error classes, never execution, transport or concrete platform implementations. Only locally enabled, capability-gated rotating linkage projections may serialize. Runtime derivation, epoch/retention/clearing enforcement, pending-state restoration and delivery transitions remain pending.

### Persistence Ports

`internal/persistence` owns narrow capabilities and transactional semantics for Barometer-owned state, including:

- admitted-task and replay records;
- pending and acknowledged result bundles;
- bounded retention metadata;
- network-context cache state where persistence is permitted;
- linkage epoch and reinstall-consistency state;
- coordinated, idempotent clearing.

Interfaces are defined from the needs of their core consumers and should be split when one consumer does not require the full capability. The core depends on these capabilities, not on files, SQL, shared preferences, Keychain, Keystore, cloud storage, or a specific database.

Persistence implementations must be bounded, atomic where required by admission or clearing, application-scoped, non-synchronizing where specified, and safe under interruption. Raw installation secrets are not values in a generic persistence interface.

DEV-591 creates the package with its first contract, `persistence.ReplayStore` and the `ReplayRecord` value consumed by admission steps 10 and 11 ([Atomic Rejection and Replay Persistence](admission.md#atomic-rejection-and-replay-persistence)). The package holds no implementation. The State, retention and clearing sub-epic (DEV-577) owns the production store and the clearing capability, and Mobile and application integration (DEV-455) owns platform storage. The only implementation is the synthetic store in the test-only `internal/testkit/replay` package.

DEV-636 adds `persistence.RunLedger` and the `RunRecord` value, the run accounting consumed by run reservation: a record is written before its run exists, survives reinitialization and holds no task input, target, label, result or participant value. Its production store, retention and clearing belong to DEV-577 and DEV-563; the only implementation is the test-only `internal/testkit/ledger`.

### Platform Ports

`internal/platform` owns normalized capabilities that cannot be implemented portably in Go, such as:

- coarse network and device state needed for execution eligibility;
- network-change and application-lifecycle signals;
- protected, non-exporting HMAC derivation and key deletion;
- trusted wall-clock or monotonic-time access where platform behavior matters;
- platform data-clear and reinstall-consistency support.

These are implementation-owned ports, not task fields and not arbitrary host or Flutter callbacks. The future native binding owns concrete Android and Apple wiring and exposes only reviewed serialized configuration and lifecycle operations to Flutter or the host.

DEV-636 adds three execution-eligibility contracts: `platform.Consent` (valid, absent, ended or unavailable, queried with no task value; the zero status is unavailable), `platform.Device` (coarse battery and charging state with explicit known flags) and `platform.TimeZone` (the device time zone for local time windows). Every error or unknown value is unsatisfied. Their shape is interim; native implementations and binding review belong to DEV-455.

The exact gomobile and native wiring mechanism is deferred to the platform-integration tasks. Any mechanism must preserve this capability boundary and must not expose protected keys, platform storage objects, unrestricted network handles, or trust assertions through the public binding.

### Test Support

The test-only `internal/testkit/conformance` subpackage owns the [Go-local versioned fixture format and runner](conformance-fixtures.md). It consumes `jsondecode`, `taskdecode`, `model/task` and `validation` for synthetic decoding assertions; runtime packages must not import it. Fixed synthetic context/policy profiles are test adapter inputs, not runtime serialization or authority.

`internal/testkit` owns deterministic synthetic implementations used by unit and conformance tests, including clocks, transports, context providers, stores, cancellation controls, and sanitized fixtures.

Production packages must never import `testkit`. An in-memory linkage or persistence implementation must be explicitly test or desktop scoped and must not become an unnoticed mobile fallback.

## Allowed Dependency Direction

The high-level dependency flow is:

```text
root facade
    ↓
admission ──────────────→ taskdecode ──→ model/task
    │                      ↑  │  │
    ├──────→ registry ─────┘  │  └──────→ jsondecode (standard-library leaf)
    │                         ↓
    └──────────────────→ validation
                              ↓
execution ─────────────→ measurement runners → protocol primitives
    │                         │                       │
    ├─────────────────────────┴──────────────→ model/observation
    ├──────────────→ networkcontext ─────────→ platform capabilities
    ├──────────────→ result ─────────────────→ model/result
    ├──────────────→ bundle ─────────────────→ model/result
    └──────────────→ persistence capabilities

native adapters → reviewed platform capabilities
test code       → production packages + testkit
```

More specifically:

- The root facade may import internal orchestration packages.
- Task decoding may import the shared strict JSON decoder, task models, and pure validation; it invokes common and timing validation before publishing a candidate (see [Task Models](task-models.md)).
- The registry may import task models, task decoding codec types, the shared strict JSON decoder and the `internal/measurement` runner contract, but no registered measurement package. Validation must not import it, because task decoding already imports validation; admission passes registry metadata to validation as values.
- Admission may import task decoding, validation, task models, registry metadata, the shared strict JSON decoder's error classes, the `internal/platform` capability contracts (for the trusted clock) and the narrow persistence capabilities it consumes. It must not import configuration, execution, measurement, protocol, Network Context, result, bundle or concrete adapter packages, and measurement packages must not import it.
- Execution may import admitted task models, the measurement contract, registry composition, Network Context, result construction, bundle construction, platform capabilities, and persistence capabilities.
- Registered measurement packages may import task and observation models, protocol primitives, and narrow capabilities needed by that test. The `internal/measurement` contract package itself imports only task and observation models and standard-library packages without network or filesystem access.
- Protocol packages may import observation models and small budget or cancellation value types owned by their consumer-facing contract.
- Result and bundle packages may import model packages but not execution or concrete network implementations. Bundle construction also consumes result construction and validated linkage configuration from `internal/config`.
- Concrete adapters may import the capability contract they implement; capability packages must not import concrete adapters.
- Tests may import production packages and `testkit`; production packages may not import `testkit`.

`TestDependencyRules` in [`dependency_rules_test.go`](../dependency_rules_test.go) enforces the import-visible rules in this section and the next through `go test ./...`, and therefore `make check` and CI. It parses every Go file regardless of build tags. Production files outside `internal/testkit` must not import `internal/testkit` or its subpackages; internal packages, including their tests, must not import the root package; domain models import only other models, `jsondecode` and standard-library packages without network or filesystem access; `jsondecode` imports only the standard library; `registry` imports only task models, `taskdecode`, `jsondecode`, the `internal/measurement` contract and standard-library packages without network or filesystem access, and measurement packages must not import it; the `internal/measurement` contract imports only task and observation models and such standard-library packages; validation, measurement, protocol, result, bundle and capability-contract packages must not import the owners forbidden below. A synthetic negative test shows that each rule rejects a forbidden import. Behavioral rules, such as no clamping, no proxy-to-direct fallback, no replay mutation and no indirect network activity, are not visible in imports and remain review rules; a passing check does not establish them.

If implementing a rule would require an import cycle, ownership must be reconsidered. A cycle must not be hidden through global registries, service locators, untyped maps, or callback bags.

## Forbidden Dependencies and Shortcuts

- Internal packages must not import the root `barometer` package.
- Domain models must not import admission, validation, execution, protocol, persistence, platform, or adapter packages.
- Validation must not perform network activity, mutate replay state, or depend on documentation examples as executable schemas.
- Measurement packages must not import one another or construct final bundle payloads.
- Protocol primitives must not schedule retries, persist state, select direct versus proxy fallback, or emit censorship classifications.
- Observation packages must not depend on result envelopes or derived interpretation.
- Result and bundle packages must not depend on a collector, analytics SDK, account system, or source-IP metadata.
- Persistence and platform abstractions must not expose raw secrets, unrestricted filesystem paths, native objects, account identifiers, or synchronization-capable storage as generic values.
- The public facade and binding must not accept interfaces or callbacks that allow a host or task to bypass trust, policy, target, path, resource, retention, or clearing enforcement.
- No package may silently broaden a task, clamp values that the specification requires rejecting, execute an invalid subset, or fall back from proxy to direct.
- No platform-specific type may appear in the portable domain models or future Flutter-facing values.
- No package may claim implementation support solely because a Draft specification, registry entry, model, or empty runner exists.

## Cross-Repository Boundary

The public specification is authoritative for observable behavior. `barometer-go` implements that behavior and records implementation findings; it does not redefine the public contract locally.

When implementation exposes an ambiguity or requires a behavior change:

1. record the ambiguity rather than silently choosing incompatible behavior;
2. propose the public specification change in `barometer-spec`;
3. align internal reviews or validation plans in `barometer-spec-internal` where required;
4. implement the reviewed behavior in `barometer-go`;
5. update native and Flutter bindings after the Go behavior is available;
6. document compatibility and release ordering across the affected repositories.

The future native and Flutter binding may depend on the stable Go facade or compiled artifacts. The Go core must not depend on Dart, Flutter, a host application, a mandatory collector, or deployment-specific configuration.

## Package-Creation Rules

Before creating a package from this map, its implementation task must establish:

- one coherent responsibility and a clear reason it cannot remain with an existing owner;
- the public specification requirements and conformance cases it implements;
- the packages it may import and the packages forbidden from importing it;
- the public, internal, test-only, or platform-specific status of its API;
- how cancellation, limits, errors, privacy, and clearing affect it;
- synthetic tests that do not require deployment secrets or participant data.

Empty placeholder packages should not be created merely to reproduce this tree. Package boundaries should be introduced with the first behavior and tests they own.

## Specification Traceability

Use the [Normative Source Inventory](normative-sources.md) for the complete document-level public baseline and [Requirement References](requirement-references.md) for stable references to individual public obligations. The table below remains a package-oriented navigation map, not a complete requirement inventory or conformance claim. Record unresolved public contract findings in [Specification Issues](specification-issues.md).

| Implementation area | Primary Draft `0.1.0` specification inputs |
|---|---|
| Task models and decoding | `data-formats/task*.md` |
| Validation and admission | `library/task-trust.md`, `library/registered-limits.md`, registered test inputs and options |
| Execution | `data-formats/task-schedule.md`, `data-formats/task-execution.md`, `library/proxy.md` |
| Measurements and observations | `tests/*.md`, `data-formats/measurement-observations.md` |
| Network Context | `data-formats/network-context.md`, `library/network-context.md`, `library/proxy.md` |
| Results and bundles | `data-formats/measurement-result.md`, `data-formats/result-bundle.md` |
| Persistence and clearing | `library/task-trust.md`, `library/result-linkage.md`, `library/registered-limits.md` |
| Platform capabilities | `library/network-context.md`, `library/proxy.md`, `library/result-linkage.md`, execution constraints in `data-formats/task-execution.md` |

Internal review documents provide validation scenarios and open questions but are not runtime or normative dependencies. Any behavior required for conformance must be present in the public specification.

## Deferred Decisions

This architecture document does not decide or implement:

- final exported Go method names or binding signatures;
- concrete package files, interfaces, or empty scaffolding;
- Android, Apple, gomobile, or Flutter adapter mechanics;
- a collector, scheduler, delivery backend, or deployment configuration;
- concrete persistence technologies;
- supported Go, Android, or Apple version matrices;
- CI, release, versioning, lint, contribution, README, or repository-agent conventions;
- implementation or support status for any registered measurement.

Those decisions belong to their dedicated tasks and must preserve the boundaries in this document or update the architecture through review.

## Review Checklist

- Every public entry point delegates admission and cannot invoke measurement execution directly on unadmitted task data.
- Trusted source context, local policy, target registration, proxy configuration, and protected-storage capabilities remain outside task-controlled values.
- Models are dependency leaves and keep execution state separate from protocol observations and interpretation.
- Registered tests reuse bounded primitives without importing one another.
- Results remain minimized, versioned, bounded, and free of secrets or censorship verdicts.
- Persistence and platform capabilities are narrow, implementation-owned, and replaceable without introducing a mandatory backend.
- Native and Flutter integration uses serialized public values and cannot bypass Go safety enforcement.
- Clearing reaches all Barometer-owned persistent and in-memory state and returns an explicit incomplete-clearing error when required deletion fails.
- Dependency direction is acyclic and contains no generic utility or service-locator escape hatch.
- Claims remain Draft until implementation, conformance, platform, and field evidence justify a reviewed status change.
