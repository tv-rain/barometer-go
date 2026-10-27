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

DEV-600 adds the lifecycle shell described in [Implemented Subset](public-api.md#implemented-subset): `NewClient`, `Initialize`, `Close` and `ClearMeasurementData`, the lifecycle state machine, the per-generation task and run handle tables and the fixed API-local error codes. The root package binds trusted adapters (clock, registry, local policy, delivery store, replay store, configuration approvals and optional linkage keys) through an unexported seam with no production caller, so initialization fails closed until reviewed native wiring exists (DEV-455). It imports `internal/config`, `internal/admission`, `internal/registry`, `internal/persistence`, `internal/platform` and `internal/model/task`, and no network package. Its tests import `internal/testkit`. DEV-592 adds `AdmitTask`, which reserves the current generation, runs `admission.Admitter.Admit` and publishes task handles indexed by replay identity. DEV-593 adds the unexported `replacePolicy` seam, which replaces the local policy only by quiescing the current generation and reinitializing a new one with a strictly higher policy generation ([Policy Replacement](public-api.md#implemented-subset)). DEV-594 adds hostile-input, strict JSON and fuzz tests through `AdmitTask` and no production code ([Hostile-Input Evidence](admission.md#hostile-input-evidence)). DEV-636 adds `StartRun` and `RunState`: the facade owns the per-generation run reservation state (accounting lookup, duplicate resolution, concurrency, the active-run limit and pending-output capacity), builds `execution.Plan` values from admitted tasks and launches `execution.Engine` on one goroutine per active run; it now also imports `internal/execution` and `internal/result`. The run inputs (engine clock, jitter, pending-output port, run ledger, an interim direct-path seam and optional consent, device and time-zone capabilities) are bound through the same unexported seam. DEV-561 adds `CancelRun`: each run gets its own context under the generation context, and the facade waits for a cancelled run for at most the registered cleanup bound on the engine clock ([Deadlines, Cancellation and Cleanup](execution-models.md#deadlines-cancellation-and-cleanup)). DEV-563 adds `NextBundle` and `AcknowledgeBundle` over the generation's [pending store](#persistence-ports), which is also the engine's output; the root package now also imports `internal/bundle` and `internal/persistence/pending`. DEV-644 replaces the interim path seam's only implementation with the production `directPath`, which `bind` derives from the `platform.Network` capability and the trusted producer implementation identity when no path is bound; the root package now also imports `internal/networkcontext` and, through `internal/execution`, links package `net` ([Network Conditions and Access Context](execution-models.md#network-conditions-and-access-context)). DEV-645 generalizes it to `networkPaths`, which also plans a proxy task through the proxy configured at `Initialize` and builds the proxy pre-attempt check from the admitted path authority and the registered proxy rule ([Proxy Path](execution-models.md#proxy-path)). The root facade also implements `SetProxy` and `ClearProxy` with an independently cancellable proxy generation. A reservation leases its immutable path before platform I/O and retains it through engine cleanup; replacement seals and drains only proxy work before publishing the next endpoint. Neither the client-state nor proxy-bookkeeping mutex is held across platform I/O or cleanup waiting. Policy replacement carries forward the active proxy configuration rather than resurrecting initialization values.

### Domain Models

The packages under `internal/model` own in-process values after decoding and normalization:

- `model/task` owns task fields, schedules, execution constraints, safety declarations, trusted source context, local-policy inputs, and admitted-task values;
- `model/observation` owns bounded DNS, TCP, TLS, HTTP, network-state, and execution observations without interpretation verdicts;
- `model/result` owns normalized attempt results, partial-result markers, bundles, network-context attachment, and optional linkage metadata.

Models are dependency leaves. They contain validation-independent value definitions and small invariants needed to keep invalid states unrepresentable, but they do not perform I/O, scheduling, persistence, logging, or platform calls.

The [Observation and Network Context Models](observation-models.md) implementation in `model/observation` owns bounded DNS/TCP/TLS/HTTP values and minimized access/egress snapshots with non-serialized path/generation binding. It depends only on the standard library and the shared strict JSON leaf. The model adopts the merged public DNS/TLS vocabularies and their internal review alignment; it performs no protocol I/O or error parsing. Since DEV-644 the `networkcontext` package consumes these values for runtime access capture rather than duplicating their domain invariants; caching remains for DEV-647.

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

Admission does not execute network activity and cannot be bypassed by the root facade, a measurement runner, a binding, or the host application. `TestAdmissionHasNoNetworkDependency` checks that the transitive dependency closure of `internal/admission` contains no network, TLS, process-execution or plugin package for the host build (DEV-594). Since DEV-644 the root package links the direct-path dialer, so its closure contains package `net`, and since DEV-647 the Network Context endpoint request, so it also contains `crypto/tls` and `net/http`; since DEV-665 the TLS-01 handshake of `internal/protocol/tls` also imports `crypto/tls` and `net`, without dialing. The test requires that `internal/execution` and `internal/protocol/tls` are the only module packages in it importing a network or TLS package, and that it contains no process-execution, plugin or testkit package. The bypass rule remains a review rule.

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

DEV-635 adds the internal [run orchestration engine](execution-models.md#run-orchestration). `Engine.Run` runs one run of an admitted task: it resolves exactly one registered runner by exact test version, executes up to two attempts sequentially, retries only as the schedule's retry object and `validation.RetryEligible` allow, takes backoff delays from an injected clock and local jitter, and passes each terminal attempt through `internal/result` to a narrow `execution.Output` port. Network Context and the attempt-scoped dialer come from trusted plan seams (`ContextSource`, `DialerSource`), whose direct-path and proxy-path implementations DEV-644 and DEV-645 add. The engine persists nothing itself; results go to the output port (DEV-563).

DEV-561 adds deadline and cancellation coordination to the engine ([Deadlines, Cancellation and Cleanup](execution-models.md#deadlines-cancellation-and-cleanup)): an attempt deadline on the injected clock, an attempt scope that wraps the path's dialer, records the test start and closes every connection when the attempt ends, and execution states selected under the GO-SPEC-ISSUE-004 disposition. The scope is the only place execution touches connections; it imports no network package and needs no address.

DEV-644 adds the direct path ([Network Conditions and Access Context](execution-models.md#network-conditions-and-access-context)). `ContextSource.Capture` returns a `networkcontext.Snapshot`, and the plan carries the task's network conditions, which `execution.NetworkConditions` evaluates against each attempt's snapshot; a generation change reported by the snapshot's watch cancels the attempt with `network_changed`. `execution.DirectDialers` in `direct.go` and `execution.ProxyDialers` in `proxy.go` are the production files that import network packages: the first gives each attempt a zero `net.Dialer`, the second opens one HTTP/1.1 CONNECT tunnel per stream through the loopback proxy endpoint (DEV-645), and the attempt scope wraps either. The plan's optional `PathCheck` runs before each proxy attempt, and a path failure the proxy dialer reports becomes the attempt's cancellation cause through the scope ([Proxy Path](execution-models.md#proxy-path)).

DEV-562 makes the attempt scope the attempt's meter ([Attempt Byte and Connection Budgets](execution-models.md#attempt-byte-and-connection-budgets)). It counts connections before each dial, issues a write only when the remaining sent budget covers it, caps each read at the remaining received budget, and ends the attempt with `resource_limit` when an operation exceeds the effective limits that `execution.Plan` carries from admission. Because every runner connection passes through it, enforcement does not depend on a runner's own reads or reported usage. Per-protocol accounting belongs to the runners and codecs under DEV-454, which use this scope.

DEV-636 adds [run reservation](execution-models.md#run-reservation) to `internal/execution`: the run-accounting functions for the interim due-time rule of GO-SPEC-ISSUE-011 over `persistence.RunRecord`, and `execution.Gates`, the fail-closed attempt-start check of network conditions, device state and consent through `internal/platform`. The engine evaluates a plan's `Preconditions` before every attempt and the `Output` port reports `Capacity`. `Client.StartRun` launches the engine; concurrency across runs stays with the facade.

### Registered Measurements and Protocol Primitives

Execution owns the [shared target gate](execution-models.md#execution-time-target-validation), using admission-owned target snapshots rather than importing admission. A runner receives only the guarded attempt dialer and a typed redirect capability; each redirect remains on the same metered scope and narrowed registration chain. Domain models supply pure tuple equality and address classification, with no resolution or I/O.


`internal/measurement` owns the minimal runner contract shared by registered measurements. Each child package owns the behavior for exactly one registered test identifier and version.

The packages under `internal/protocol` own reusable bounded protocol operations. Registered measurement packages may depend on these primitives, domain models, and narrowly required platform capabilities. They must not depend on another registered measurement package.

`NET-05` composes reviewed primitives through its own admitted sequential plan. It must not call another test package as an opaque shortcut, execute steps concurrently, execute a compatible subset, or produce a local censorship verdict.

Protocol primitives return observations and budget usage. They do not construct final result envelopes, select retries, change paths, persist data, or classify censorship.

`internal/protocol/http` owns bounded HTTP/1.1 GET and fixed-DNS-media-type POST exchanges over caller-established streams. Its `Endpoint` contract owns transport-stage opening and complete redirect-tuple authorization through the attempt's checked dialers; the primitive never opens a socket, resolves a name, selects a path or authenticates TLS itself. HTTPS callers must offer only HTTP/1.1, authenticate the URL host (including IP SAN validation for explicit-IP URLs), switch the attempt connection through `SessionConn.Session`, and wrap the authenticated session's plaintext with its returned meter. The shared TLS primitive accepts DNS names and IP literals for certificate verification; TLS-01's codec still requires its registered DNS server-name input.

HTTP bounds cover the cumulative response status/header blocks (16,384 bytes, including informational responses), at most 65,536 terminal body bytes, eight header and eight body checks, and at most two separately authorized redirects. The operation timer spans request writing, headers and body reading. Cookies, credentials, caches, identifying headers, decompression, protocol upgrade, automatic retries and connection pooling are absent. Request lines, headers, bodies and response read-ahead cross the plaintext meter; TLS records and CONNECT framing remain below it. Body processing never probes beyond its cap: a declared length or observed EOF establishes completion; an exactly-at-cap body without either remains conservatively truncated. Prefix/contains checks may be decided before completion; suffix/digest checks require completion. Only decidable indexed outcomes enter observations.

`Exchange` returns a terminal primitive result and a cleanup function. A registered caller must fix its outcome with `Input.RecordOutcome` before cleanup and must always clean up; intermediate-hop streams close before the next contact. Outer interruption returns no terminal result. A local safety refusal discovered while opening a redirect returns `DestinationRejected`, updates the observed redirect's minimized class and does not invent a next HTTP stage or increment its contacted count. A POST caller can take the bounded transient payload once to parse its DNS message; GET results retain no payload, and neither serialized observations nor formatted diagnostics expose payloads or check values. Registered HTTP/DoH/NET-03 callers and full result integration remain separate work.


DEV-635 creates `internal/measurement` with the runner contract only. `Runner.Run` receives the admitted normalized parameters, the attempt context, an attempt-scoped `Dialer` and, since DEV-561, a `Start` marker for a test whose procedure observes network state before dialing, and returns a terminal outcome, its observations and the runner's `Usage` (application bytes sent and received, connections). Since DEV-562 that report enforces nothing: the attempt scope meters the dialer and its connections. `Conn` exposes only read, write, close and a deadline, with no address, socket or platform handle. The contract imports only task and observation models and standard-library packages without network access, so admission, which reads the registry, links no network code. A runner reports an execution failure with the fixed `ErrRunner`; a negative protocol outcome is an observation, not an error. The production runners are `internal/measurement/tcp01`, using `internal/protocol/tcp`, `internal/measurement/dns01`, using `internal/protocol/dns`, and `internal/measurement/tls01`, using `internal/protocol/tcp` and `internal/protocol/tls`. A connection that carries TLS is switched to plaintext metering through `SessionConn`, which the attempt scope implements. The test-only echo/loopback runners remain separate synthetic runtime fixtures. The runner input also carries the engine clock, a narrow stable-generation/address-family check and a terminal-outcome marker. The check starts each measurement procedure; the marker refuses an already interrupted outcome and preserves a recorded outcome during close, including proxy reconfiguration. `internal/measurement/net01` receives no dialer: its input carries a `Network` read through the shared collector, which starts the test, and its output carries the normalized access snapshot.

The runner's optional system `Resolve` function belongs to the attempt scope and rechecks the admitted name and registration expiry before one `execution.SystemResolver` call. It returns only a success flag or a fixed failure class. The default is Go's ordinary resolver API with a fully qualified name to suppress search-list expansion; hidden resolver traffic is outside byte accounting. No system responses, codes or datagram timing are synthesized. Native selection, hidden retries/CNAME processing, cancellation and error fidelity remain device-validation gaps. The capability retains no Barometer-owned query or result state.

For UDP DNS, `DatagramConn` carries source equality and truncation booleans through the attempt meter. `execution.directUDP` sends to the checked literal peer and compares sources without exposing socket addresses to the measurement. Every returned payload, including discarded datagrams, counts; truncation caused by an exhausted attempt budget stops as `resource_limit`. DNS protocol code consumes models and leaf dependencies, never the measurement contract, registry or execution owner.

### Registry

`internal/registry` is the composition point for reviewed test identifiers, supported specification versions, runner implementations, validation metadata, and proxy compatibility. It owns the mapping used by admission and execution.

Adding an identifier to the registry requires a corresponding public registered test, internal review alignment, implementation, and conformance evidence. Registry presence alone must not be presented as released or field-validated support.

`registry.New` builds an immutable registry from a bounded list of entries, including an empty list. Each entry pairs an exact `(test_id, test_version)` with its `taskdecode.TestCodec`, registered baseline risk, a target rule that reports the complete typed tuple of every destination or name its normalized parameters would contact, a proxy-compatibility rule over normalized parameters, and whether the test is the NET-05 composite or a permitted NET-05 step. Construction rejects the whole set with one fixed error when any entry is incomplete, duplicated, malformed or inconsistent with its codec. `Lookup` matches exact strings only, with no latest-version fallback, case folding or partial match, and is safe for concurrent reads. `Codecs` returns a copied container of the same reviewed codec set for building the task decoder. Codecs, schema descriptors, rules and runners remain immutable for its lifetime. Admission combines the registered baseline risk with the declared and locally determined risk ([Effective Risk and Consent](admission.md#effective-risk-and-consent)).

An entry may carry a `measurement.Runner`, and `Runner` returns it only for the exact test ID and version. A missing entry or an entry without a runner reports false, and execution then fails closed before any work; there is no default runner or version fallback. The runner stays optional at construction so admission-only entries remain valid.

The root package's unexported `productionRegistry` constructor assembles the production entry list and passes it to `registry.New`. Assembly stays in the root so neither the metadata registry nor a measurement package imports the other. There is no global registry or runtime registration method. `Client.bind` selects this constructor when its internal adapter registry is nil; tests may explicitly inject a synthetic registry. Admission and execution receive that same registry for the generation. The production list contains DNS-01, TCP-01, TLS-01 and HTTP-01 `0.1.0`, with low baseline risk and NET-05 step eligibility. It also contains NET-01 `0.1.0` with low baseline risk, `registry.NoTargets` and `registry.NoProxy`, not eligible as a NET-05 step, and NET-05 `0.1.0` with medium baseline risk and a sequential runner that calls the shared DNS/TCP/TLS/HTTP operations directly, without importing sibling measurement packages. The engine gives each step exact initial target tuples and redirect authority, while every step retains the same path, context, cancellation and aggregate meter. A generation-owned runner serializes whole profiles; waiters honor their own attempt deadlines. A trusted internal local-policy flag controls platform validation, required by default; it is absent from task JSON and the public facade. Unknown platform validation never authorizes primitive traffic. No active validation probe is added. Its codec receives an immutable set of primitive codecs from root assembly; it imports neither primitive packages nor the registry. `task.CompositeParameters` supplies an owned ordered step list to admission, which checks each exact primitive entry, permissions and registrations before retaining the union of matched authority. Root-composed target and proxy rules reuse the same primitive entries. The admission limitations are recorded in [NET-05 coverage](measurement-requirements.md#net-05). TCP-01 matches an explicit public IP/port tuple and permits CONNECT. TLS-01 matches the explicit public IP/port, normalized server name and ordered ALPN offer under scheme `tls` and permits CONNECT; its trust anchors come from the internal adapter seam, where nil selects the system roots. DNS-01 matches the normalized query-name registration and, in direct mode, a separate resolver transport/address/port registration; only its direct TCP mode permits CONNECT. Other well-formed unregistered test versions reject as `task_unsupported` without a task handle, replay write or run. An empty decoder still applies strict JSON framing and exact schema-version checks. Other required platform adapters remain absent from `NewClient`, so it still fails initialization closed.

#### Registered Test Package Pattern

Introduce a per-test package only with its first real behavior and tests. TCP-01 lives in `internal/measurement/tcp01`, with its shared operation in `internal/protocol/tcp`. DNS-01 lives in `internal/measurement/dns01`, with bounded DNS operations and wire parsing in `internal/protocol/dns`. TLS-01 lives in `internal/measurement/tls01`, with the bounded handshake in `internal/protocol/tls`. NET-01 lives in `internal/measurement/net01` and uses no protocol package. NET-03 lives in `internal/measurement/net03` and reuses the TLS and HTTP protocol packages with redirects disabled. DNS-04 lives in `internal/measurement/dns04` and reuses the staged TCP connection, the TLS handshake and the length-prefixed exchange of `internal/protocol/dns` through `dns.Session`. Task-supplied DNS names share `taskdecode.DNSName`. Do not create placeholder packages or register the synthetic TCP-01 fixtures already used in tests. Each exact test version supplies the following pieces for explicit assembly in `productionRegistry`:

| Piece | Owner and contract |
|---|---|
| Codec | The test package implements `taskdecode.TestCodec`. `Definition` supplies the exact public ID/version and closed, non-null object descriptors for `input` and `options`. `Normalize` consumes validated `jsondecode.Value` objects, checks remaining test semantics and returns typed `task.Parameters`, or no parameters on failure. It performs no I/O or authority decisions. |
| Target rule | A root adapter implements `registry.TargetRule` by consuming test-owned typed accessors and constructing a complete `registry.Target` tuple for every contacted destination/name. The test package cannot import `registry`. Wrong parameter types or incomplete tuples reject. No resolution, rewriting or inferred registration occurs. |
| Proxy rule | The test package supplies a pure `Compatible(task.Parameters) bool` rule, or root assembly uses `registry.NoProxy` for an incompatible test. Compatibility cannot grant proxy authority or permit direct fallback; admission also applies the common explicit-IP destination restriction. |
| Risk and composition | Root assembly copies the reviewed public baseline risk and composite/step eligibility into the entry. These values cannot come from task-controlled input or reduce locally determined risk. |
| Runner | The test package implements `measurement.Runner` using admitted typed parameters, the attempt context and scoped capabilities. Shared primitives return observations and usage; execution owns deadlines, budgets, paths, retries and result envelopes. |

A target adapter therefore lives beside the production constructor, while its typed target data and semantics remain test-owned. The codec and proxy rule satisfy their interfaces structurally without importing `registry`. `TestDependencyRules` enforces that separation, sibling measurement isolation, protocol ownership and the ban on production imports of `testkit` across all Go files, regardless of build tags. No new public API or runtime extension point is needed.

**Strict JSON adoption.** Use the existing `taskdecode` stages and `internal/jsondecode` ([consumer contract](json-decoding.md#consumer-contract)); do not add an independent permissive parser. Envelope parsing rejects duplicate decoded names, including escaped-equivalent names, before schema selection. Stage 8 validates the selected input/options descriptors and only then invokes normalization. Select one exact ID/version, with no compatible-version retry or empty-object fallback. Descriptors must specify required/optional fields, nullability, exact integer domains, field/count/encoded-byte bounds and registered-limit classification. Preserve missing, null, empty and zero distinctions and original encoded byte sizes; defaults apply only where the public test defines them. Shared task/input/options byte caps still apply, and a stricter test cap must be represented explicitly. Rejection must be atomic and expose only the fixed admission class.

For every production schema, the owning task must provide positive controls, duplicate and escaped-equivalent names, escaped unknown names, malformed types/values, integer spellings and exact endpoints, below/at/above registered bounds, and atomic rejection cases through the actual `Client.AdmitTask` path with no handle, replay write, run or network activity. Add affected-boundary fuzzing and applicable `REQ-JSON-*` references. Inventory any runtime JSON consumer/producer and record its adoption or not-applicable disposition under [DEV-622](https://app.clickup.com/t/124kk1zcge1), including the GO-SPEC-ISSUE-001 adoption evidence. TCP-01 and DNS-01 record their adoption under DEV-622 in [Measurement Requirements](measurement-requirements.md); each later production schema records its own. Traceability baselines and evidence reports remain milestone work.

### Network Context

`internal/networkcontext` owns normalized access and egress context, cache rules, primary and fallback acquisition, path consistency, coarse failure states, and privacy filtering. It consumes platform and network capabilities but does not expose raw platform objects or unrestricted endpoint responses.

Network Context cannot receive bundle identifiers, linkage material, participant identifiers, task provenance secrets, or collector credentials. Its observations cannot by themselves establish censorship.

DEV-644 creates the package with access-context capture for the direct path, and DEV-645 adds the proxy path. `Source.Capture` reads the `platform.Network` capability once, normalizes the reading with `Normalize` and binds it to the source's path in its access-network generation: the direct path, or the proxy path of one proxy snapshot with its `connection_code`; the `Snapshot` provides the access values, the Network Context to attach, and the generation watch. DEV-647 adds egress acquisition: `Lookup` holds one generation's configured endpoints, reviewed classification approvals and trust roots, and `Snapshot.Acquire` applies the bounded cache, primary and fallback order, decodes the endpoint response strictly and binds an approved classification to the snapshot's path. `Cache` is the client's in-memory egress cache, keyed by access-network generation, path and proxy generation, with an epoch per generation, proxy invalidation and an eraser for clearing. The package performs no I/O itself: it receives a `Fetch` from `internal/execution`, the socket owner, bound to a dialer of the attempt's own path ([Egress Network Context](execution-models.md#egress-network-context)). It imports only model packages, `jsondecode`, the `internal/platform` contracts and the standard library, and opens no socket.

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

The [Result Bundle and linkage models](bundle-models.md) implement detached projections in `model/result` and the bounded producer in `internal/bundle`. The producer composes validated immutable `internal/result.Value` members and copies policy from prepared `internal/config.Initialization`; it depends on those model boundaries and shared decoder error classes, never execution, transport or concrete platform implementations. Only locally enabled, capability-gated rotating linkage projections may serialize. Since DEV-563 the [pending store](#persistence-ports) persists and returns the producer's exact bytes; it never restores a `bundle.Value` from storage, so no restoration boundary is added. Linkage derivation (DEV-454), epoch enforcement and clearing (DEV-564) remain pending.

### Persistence Ports

`internal/persistence` owns narrow capabilities and transactional semantics for Barometer-owned state, including:

- admitted-task and replay records;
- pending and acknowledged result bundles;
- bounded retention metadata;
- network-context cache state where persistence is permitted (none in this build: the egress cache is in memory only);
- linkage epoch and reinstall-consistency state;
- coordinated, idempotent clearing.

Interfaces are defined from the needs of their core consumers and should be split when one consumer does not require the full capability. The core depends on these capabilities, not on files, SQL, shared preferences, Keychain, Keystore, cloud storage, or a specific database.

Persistence implementations must be bounded, atomic where required by admission or clearing, application-scoped, non-synchronizing where specified, and safe under interruption. Raw installation secrets are not values in a generic persistence interface.

DEV-591 creates the package with its first contract, `persistence.ReplayStore` and the `ReplayRecord` value consumed by admission steps 10 and 11 ([Atomic Rejection and Replay Persistence](admission.md#atomic-rejection-and-replay-persistence)). The State, retention and clearing sub-epic (DEV-577) owns the production store and the clearing capability, and Mobile and application integration (DEV-455) owns platform storage.

DEV-636 adds `persistence.RunLedger` and the `RunRecord` value, the run accounting consumed by run reservation: a record is written before its run exists, survives reinitialization and holds no task input, target, label, result or participant value. DEV-564 clears it with every other record class.

DEV-641 adds the state record port and the production stores built on it:

- **Record port.** `persistence.Records` stores bounded bytes within one application-scoped storage scope: atomic `Put` and `Delete` of one record, `Get`, and `List` of one record class. A key is a closed `RecordClass` (`ReplayClass`, `RunClass`; DEV-563 adds pending bundles and DEV-564 the clearing marker) and a fixed 32-byte name. A value is at most 262,144 bytes, one encoded Result Bundle ([REQ-LIMIT-0021](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0021)), and a class lists at most 4,096 keys. Every call honors cancellation and its context deadline. An error from `Put` or `Delete` is an uncertain outcome: the earlier or the new state is durable, never a mixture. Values carry no secret, task input, target, label or participant value. A platform backend (DEV-455) implements the port and only stores bytes.
- **Production stores.** `persistence.State`, opened by `OpenState(ctx, records, capacity, now)`, implements `ReplayStore` (`State.Replay`) and `RunLedger` (`State.Runs`) over one `Records`. The Go code owns encoding, validation, retention and bounds. Record names are domain-separated SHA-256 digests of the replay key, so no issuer or task identifier is a platform key. Values are a fixed binary encoding of exactly the record fields, decoded strictly and validated; an unreadable, invalid or misnamed value is invalid state, never absence. `OpenState` validates every stored record, deletes records overdue at `now` and run records without a replay record, and fails closed on any invalid record. It keeps an in-memory index of retention ends and accepted generations; an uncertain write is indexed as if durable, so retention, capacity and `HighestGeneration` account for it.
- **Conditional writes and ownership.** `Commit` is conditional in Go, under the `State` lock: the port has no compare-and-swap. This relies on one open `State` per storage scope; enforcing a single owner is the platform backend's (DEV-455).
- **Bounds.** A `State` keeps 1 to 4,096 replay records, a capacity chosen by its binding; this bound is implementation-local, not a registered limit. A full store rejects commits and never evicts a retained record.

No production code binds a `State` yet; the root adapters stay unbound until DEV-455. The synthetic `internal/testkit/replay` and `internal/testkit/ledger` stores remain for tests, and the [fault-injectable record store](testkit.md#scope-and-ownership) backs tests of the production stores.

DEV-563 adds the pending results and bundles:

- **Record class and bound.** `PendingClass` holds one pending Result Bundle per record. A record value is now at most 262,144 bytes plus a store header of at most 64 bytes (`RecordHeaderSize`), so one encoded bundle and its header fit.
- **Pending store.** `internal/persistence/pending` implements `execution.Output` and the bundle lifecycle over one `Records` port. It is a subpackage because it composes `internal/bundle` and `internal/result`, which the capability contract package `internal/persistence` must not import. `pending.New` takes the port, the generation's `bundle.Builder`, the engine clock, reviewed limits and the lifecycle deadline, and reads nothing; `Open`, or the first operation, reads, validates and indexes every stored bundle and deletes expired ones. Any unreadable, invalid or misnamed record fails it closed with `ErrUnavailable`.
- **Bundles as results arrive.** Results are not stored outside a bundle. The newest bundle stays open while the store that built it is open: a result is added by rebuilding that bundle with `bundle.Build` and a fresh identifier and replacing its record in one atomic `Put`, so a crash leaves the previous or the new bundle, never a mixture. A result the open bundle cannot hold (16 results or 262,144 bytes) starts a new bundle; a result no bundle can hold within 32 bundles and 4,194,304 aggregate bytes is rejected with `ErrFull` and never truncated. A bundle is sealed once `Next` returns it, and every bundle read back from storage, after a restart or an uncertain write, is sealed. The open bundle's identifier is never returned and is not acknowledged.
- **Record format.** A record name is a domain-separated SHA-256 digest of a store sequence number that orders bundles oldest first, so no bundle identifier or result value is a platform key. A value is a format byte, the class, the sequence number and the creation time in Unix nanoseconds, then the exact bundle bytes, which are checked against the form the store writes (no linkage) when read.
- **Retention.** A bundle is deleted, never returned, at its creation plus 86,400 seconds, on every access and when the store is opened. Creation is the persisting of its first result, the interim rule of [GO-SPEC-ISSUE-012](specification-issues.md#go-spec-issue-012-start-of-unacknowledged-bundle-retention). A failed expiry deletion fails the access.
- **Uncertain writes and deletions.** A failed `Put` makes the store reread storage before its next operation, so its index never assumes either outcome. A failed `Acknowledge` deletion keeps the bundle indexed and retrievable; a retry deletes it, and a record already gone is dropped from the index when found.
- **Concurrency.** Every operation holds a one-slot semaphore that a caller waits for only until its context ends, and is bounded by the store deadline, so no uninterruptible lock is held across storage I/O. `persistence.State` still holds a `sync.Mutex` across its port calls, which is outside this task.
- **Ownership.** `Client.Initialize` opens the generation's store before the generation becomes active, so pending storage is purged on initialization. A policy replacement builds its successor's store without I/O while the current generation still owns storage; the successor loads it on first use. One owner per storage scope is enforced by the platform backend (DEV-455).

The root adapters bind the pending `Records` port and optional lower pending limits; the store is not a separate adapter, and no production code binds the port yet.

DEV-564 adds persistent clearing over the same port:

- **Marker class.** `ClearingClass` holds one clearing-in-progress marker under a fixed, domain-separated name. Record classes are a closed, append-only enumeration ending in an unexported sentinel, and stored values carry the class byte, so a new class is added before the sentinel and never renumbers an existing one.
- **Clear.** `persistence.Clear(ctx, records, erasers...)` in `internal/persistence/clear.go` writes the marker, deletes every record of every other class, newest class first (pending bundles, then run accounting, then replay records, so no run record outlives its replay record), lists each class again to confirm it is empty, runs each eraser and finally deletes every marker-class record, the marker last. Any failure returns `ErrClearIncomplete` with no detail and leaves the marker, except that an uncertain deletion of the marker itself, after everything else is gone, may remove it. It is idempotent, bounded by its context, reads no clock and has no transport. `persistence.Clearing` reports whether the marker is stored; a stored value other than the marker's counts as the marker, and an unreadable one is an error.
- **Registration point.** A class added to `RecordClass` is cleared without registration. State that is not stored over the port is an `Eraser`, a context-bounded deletion that creates no replacement; the root package's `storageClearer` builds the list from the adapters (the optional linkage key store and the in-memory Network Context cache). `TestClearingCoversEveryAdapter` fails for an adapter field that is neither cleared there nor recorded as holding no Barometer-owned state, and `TestClearCoversEveryRecordClass` fails when the class set changes without review.
- **In-memory state.** Go stores that index records implement `Forgetter`. `State.Forget`, reachable through either view, drops the replay and run index after a successful clear, so no deleted retention, capacity or generation is counted. `pending.Store.Forget` drops a store's index and open bundle without waiting for an operation in progress. The client calls both only after `Clear` succeeds, when its reservations have drained. Go cannot guarantee that memory is overwritten: strings are immutable and the garbage collector may hold or have copied earlier values.
- **Lifecycle.** `Initialize` reads the marker before opening the pending store and blocks the client with `clearing_incomplete` while it is stored, so a clear interrupted by a failure or a restart blocks every measurement and result call until recovery clearing completes ([Implemented Subset](public-api.md#implemented-subset)).
- **Locking.** `persistence.State` still holds its `sync.Mutex` across port calls. Clearing does not depend on it: the client drains every reserved call before deletion, and `Forget` runs only after `Clear` returns. Replacing that lock is a separate follow-up.

Native reset wiring, operating-system data clearing, keystore deletion, reinstall sentinels and host-owned copies belong to DEV-455.

### Platform Ports

`internal/platform` owns normalized capabilities that cannot be implemented portably in Go, such as:

- coarse network and device state needed for execution eligibility;
- network-change and application-lifecycle signals;
- protected, non-exporting HMAC derivation and key deletion;
- trusted wall-clock or monotonic-time access where platform behavior matters;
- platform data-clear and reinstall-consistency support.

These are implementation-owned ports, not task fields and not arbitrary host or Flutter callbacks. The future native binding owns concrete Android and Apple wiring and exposes only reviewed serialized configuration and lifecycle operations to Flutter or the host.

DEV-636 adds three execution-eligibility contracts: `platform.Consent` (valid, absent, ended or unavailable, queried with no task value; the zero status is unavailable), `platform.Device` (coarse battery and charging state with explicit known flags) and `platform.TimeZone` (the device time zone for local time windows). Every error or unknown value is unsatisfied. Their shape is interim; native implementations and binding review belong to DEV-455.

DEV-644 adds `platform.Network`: `Access` returns a coarse `AccessState` (network type, VPN state, IPv4 and IPv6 availability, validation state and an access-network generation, with unknown as every zero value), and `Watch` reports once that the generation has moved. The struct has no field for any item REQ-CONTEXT-0003 forbids. Its shape and change signal are interim; native implementations and binding review belong to DEV-455.

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

`TestDependencyRules` in [`dependency_rules_test.go`](../dependency_rules_test.go) enforces the import-visible rules in this section and the next through `go test ./...`, and therefore `make check` and CI. It parses every Go file regardless of build tags. Production files outside `internal/testkit` must not import `internal/testkit` or its subpackages; internal packages, including their tests, must not import the root package; domain models import only other models, `jsondecode` and standard-library packages without network or filesystem access; `jsondecode` imports only the standard library; `registry` imports only task models, `taskdecode`, `jsondecode`, the `internal/measurement` contract and standard-library packages without network or filesystem access, and measurement packages must not import it; the `internal/measurement` contract imports only task and observation models and such standard-library packages; validation, measurement, protocol, result, bundle and capability-contract packages must not import the owners forbidden below; since DEV-644 no production package outside `internal/execution`, `internal/protocol` and `internal/testkit` imports a network package, and `internal/networkcontext` imports only model packages, `jsondecode`, the `internal/platform` contracts and the standard library. A synthetic negative test shows that each rule rejects a forbidden import. Behavioral rules, such as no clamping, no proxy-to-direct fallback, no replay mutation and no indirect network activity, are not visible in imports and remain review rules; a passing check does not establish them.

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
