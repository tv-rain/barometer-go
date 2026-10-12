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

The host remains responsible for enablement, authenticated task delivery, scheduling, deployment policy, optional proxy operation, result delivery, transparency, and deployment-specific review. The Go core remains responsible for admission, validation, bounded execution, registered measurement behavior, minimized results, cancellation, retention, and clearing.

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
    ├── model
    │   ├── task                         admitted task and trusted-context domain values
    │   ├── observation                  bounded protocol and network observations
    │   └── result                       result, bundle, and linkage domain values
    ├── taskdecode                       deterministic decoding of serialized tasks
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

### Root `barometer` Package

The module root is the only initially intended public Go package. It owns:

- initialization and shutdown entry points;
- the small lifecycle facade used by a binding or direct Go integrator;
- serialized task admission and result retrieval entry points;
- cancellation and Barometer-owned data-clearing entry points;
- stable, coarse errors or status values that an integrator must handle.

It does not own measurement logic, JSON schema implementation, trust policy, persistence implementations, platform secrets, collector behavior, or censorship classification. Exporting additional Go packages or interfaces is a public API decision and requires explicit review.

Exact exported method names and serialized representations are deferred until their dedicated API tasks. This document fixes ownership and dependency direction, not the final API signatures.

### Domain Models

The packages under `internal/model` own in-process values after decoding and normalization:

- `model/task` owns task fields, schedules, execution constraints, safety declarations, trusted source context, local-policy inputs, and admitted-task values;
- `model/observation` owns bounded DNS, TCP, TLS, HTTP, network-state, and execution observations without interpretation verdicts;
- `model/result` owns normalized attempt results, partial-result markers, bundles, network-context attachment, and optional linkage metadata.

Models are dependency leaves. They contain validation-independent value definitions and small invariants needed to keep invalid states unrepresentable, but they do not perform I/O, scheduling, persistence, logging, or platform calls.

Serialized wire objects and admitted domain values must not be conflated. In particular, trusted source context, local policy, replay state, proxy configuration, and protected-storage capabilities are never populated from task-controlled fields.

### Deterministic Task Decoding

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

Admission does not execute network activity and cannot be bypassed by the root facade, a measurement runner, a binding, or the host application.

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

### Registered Measurements and Protocol Primitives

`internal/measurement` owns the minimal runner contract shared by registered measurements. Each child package owns the behavior for exactly one registered test identifier and version.

The packages under `internal/protocol` own reusable bounded protocol operations. Registered measurement packages may depend on these primitives, domain models, and narrowly required platform capabilities. They must not depend on another registered measurement package.

`NET-05` composes reviewed primitives through its own admitted sequential plan. It must not call another test package as an opaque shortcut, execute steps concurrently, execute a compatible subset, or produce a local censorship verdict.

Protocol primitives return observations and budget usage. They do not construct final result envelopes, select retries, change paths, persist data, or classify censorship.

### Registry

`internal/registry` is the composition point for reviewed test identifiers, supported specification versions, runner implementations, validation metadata, and proxy compatibility. It owns the mapping used by admission and execution.

Adding an identifier to the registry requires a corresponding public registered test, internal review alignment, implementation, and conformance evidence. Registry presence alone must not be presented as released or field-validated support.

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

### Bundle Construction

`internal/bundle` owns bounded grouping of normalized results, bundle identifiers, optional linkage metadata, serialized bundle-size enforcement, and local delivery state transitions. Result delivery remains replaceable and host-controlled.

The package must operate without a mandatory collector. It must not accept account identity, source IP metadata, task selection state, or arbitrary stable identifiers as linkage inputs.

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

### Platform Ports

`internal/platform` owns normalized capabilities that cannot be implemented portably in Go, such as:

- coarse network and device state needed for execution eligibility;
- network-change and application-lifecycle signals;
- protected, non-exporting HMAC derivation and key deletion;
- trusted wall-clock or monotonic-time access where platform behavior matters;
- platform data-clear and reinstall-consistency support.

These are implementation-owned ports, not task fields and not arbitrary host or Flutter callbacks. The future native binding owns concrete Android and Apple wiring and exposes only reviewed serialized configuration and lifecycle operations to Flutter or the host.

The exact gomobile and native wiring mechanism is deferred to the platform-integration tasks. Any mechanism must preserve this capability boundary and must not expose protected keys, platform storage objects, unrestricted network handles, or trust assertions through the public binding.

### Test Support

`internal/testkit` owns deterministic synthetic implementations used by unit and conformance tests, including clocks, transports, context providers, stores, cancellation controls, and sanitized fixtures.

Production packages must never import `testkit`. An in-memory linkage or persistence implementation must be explicitly test or desktop scoped and must not become an unnoticed mobile fallback.

## Allowed Dependency Direction

The high-level dependency flow is:

```text
root facade
    ↓
admission ──────────────→ taskdecode → model/task
    │                         ↓
    └──────────────────→ validation → registry metadata
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
- Admission may import task decoding, validation, task models, registry metadata, and the narrow persistence capabilities it consumes.
- Execution may import admitted task models, the measurement contract, registry composition, Network Context, result construction, bundle construction, platform capabilities, and persistence capabilities.
- Registered measurement packages may import task and observation models, protocol primitives, and narrow capabilities needed by that test.
- Protocol packages may import observation models and small budget or cancellation value types owned by their consumer-facing contract.
- Result and bundle packages may import model packages but not execution or concrete network implementations.
- Concrete adapters may import the capability contract they implement; capability packages must not import concrete adapters.
- Tests may import production packages and `testkit`; production packages may not import `testkit`.

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

Use [Requirement References](requirement-references.md) for stable references to individual public obligations. The table below remains a document-level navigation map, not a complete requirement inventory or conformance claim.

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
