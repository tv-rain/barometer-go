<p align="left">
  <img src="assets/barometer-logo.svg" alt="Barometer" width="80">
</p>

# Barometer Go

Barometer Go is the platform-neutral core planned for the Barometer mobile network-measurement library. It is intended to admit reviewed tasks, enforce safety and resource bounds, execute registered measurements, and return minimized structured observations without requiring a particular scheduler, collector, backend, or host application.

## Current Status

This repository is at the engineering-foundation stage. It currently contains the Go module bootstrap, an accepted package-architecture plan, [shared strict JSON decoding infrastructure](docs/json-decoding.md), and [common task and schedule models](docs/task-models.md). Common task decoding and pure validation produce candidates using explicitly supplied test schemas, and the [admission pipeline](docs/admission.md) is implemented with synthetic evidence; the remaining production test-specific schemas and registered measurements, production persistence adapters, full runtime conformance, native adapters and bindings, mobile release artifacts, and a supported public API remain pending.

A shared [execution target gate](docs/execution-models.md#execution-time-target-validation) checks admitted registrations, expiry, path and local address policy before contact. Direct resolution is checked and pinned; hostname CONNECT is refused, including redirects and proxy context lookups. Synthetic evidence does not establish full conformance, platform or field support.

DNS-01, TCP-01, TLS-01 and HTTP-01 have production measurement entries. DNS-01 performs one system lookup or one bounded direct UDP/TCP query; TCP-01 performs one bounded connection; TLS-01 performs one bounded authenticated handshake on one explicit-IP connection; HTTP-01 performs one bounded identity-encoded HTTPS GET plus at most two separately authorized redirects, with minimized indexed checks. Direct transports and HTTP CONNECT tunnels use destination checks and attempt metering; DNS system lookups expose no wire bytes and are incompatible with CONNECT, as is direct UDP. [DNS-01](docs/measurement-requirements.md#current-dns-01-execution), [TCP-01](docs/measurement-requirements.md#current-tcp-01-execution) and [TLS-01](docs/measurement-requirements.md#current-tls-01-execution) execution notes record synthetic admission, operation, termination and output evidence and the remaining native gaps. This does not establish full conformance, platform or field support. [HTTP-01 execution notes](docs/measurement-requirements.md#current-http-01-execution) record direct/proxy, outcome-precedence and termination evidence and the current IPv4 selection for direct hostname `auto`; other registered measurement runners remain pending. NET-01 reads coarse access state once through the platform network port, with no network traffic or egress lookup, and records it without the top-level Network Context; it does not yet emit `inconclusive`, and native network-state behavior remains pending ([NET-01 execution notes](docs/measurement-requirements.md#current-net-01-execution)). NET-03 reads the platform validation and captive-portal state once and sends one bounded `http` or `https` validation GET that never follows a redirect, keeping the platform and active evidence separate; its response classes and outcome order follow the selected rules prepared publicly as REQ-NET-03-0041 to REQ-NET-03-0043, and native validation signals remain pending ([NET-03 execution notes](docs/measurement-requirements.md#current-net-03-execution)). NET-05 admits the whole profile atomically and executes its DNS/TCP/TLS/HTTP steps sequentially through shared protocol primitives, using one metered attempt scope and retaining bounded partial results. Platform validation is required by default; unknown validation produces an inconclusive result without primitive traffic. Profile-version history and summed primitive-maxima admission checks remain approved deferrals; [NET-05 coverage](docs/measurement-requirements.md#net-05) records the synthetic evidence and limitations.

The Barometer `0.1.0` public specifications are Draft implementation inputs. A specification, package plan, successful Go build, or successful cross-compilation check is not evidence that a measurement or platform is implemented, conformant, field-validated, or supported.

The [conformance fixture format and runner](docs/conformance-fixtures.md) add versioned, strictly validated synthetic task-decoding cases with separate context/policy profiles and exact expected candidate projections or rejection classes. The fixtures exercise decoding only; admission and local policy enforcement are implemented separately with synthetic evidence ([Admission Design](docs/admission.md)), and measurement execution remains pending.

Reusable [negative and boundary-test helpers](docs/testkit.md) provide bounded synthetic fixtures, rejection diagnostics, and deterministic time/transport controls for future tests. They do not implement runtime schemas or measurements.

[Initialization and configuration models](docs/configuration-models.md) add internal immutable host configuration, atomic value replacement, trusted deployment approval inputs and narrow clock/protected-key capability contracts. Configuration preparation does not activate a client or replace the pending admission, execution and native adapter work.

[Trusted context and local policy models](docs/trusted-policy-models.md) provide separate internal source metadata, exact-byte delivery binding, immutable issuer/target policy snapshots and revocation inputs. They have no JSON import/export path and do not implement authentication, effective authorization, admission, replay, policy activation or revocation enforcement.

[Execution-state and error models](docs/execution-models.md) add bounded internal admission errors, pure attempt transitions and the complete execution-reason vocabulary. Admission returns these errors ([Admission Design](docs/admission.md)). An internal [run orchestration engine](docs/execution-models.md#run-orchestration) invokes one registered runner per attempt through the [runner contract](docs/package-architecture.md#registered-measurements-and-protocol-primitives), sequences retries per the schedule and passes each attempt result to an output port, exercised only with synthetic and loopback test runners; `Client.StartRun` launches it after [run reservation](docs/execution-models.md#run-reservation). Each attempt ends at its `max_duration_ms` on an injected clock, and an attempt scope closes its connections when it ends ([Deadlines, Cancellation and Cleanup](docs/execution-models.md#deadlines-cancellation-and-cleanup)) and meters them against the attempt byte and connection limits ([Attempt Byte and Connection Budgets](docs/execution-models.md#attempt-byte-and-connection-budgets)); per-protocol accounting, production runners and persistence remain pending.

A narrow [state record port](docs/package-architecture.md#persistence-ports) stores Barometer-owned state as bounded bytes; the production replay store and run ledger, `persistence.State`, implement encoding, validation and retention over it in Go and are tested only against a fault-injectable test store. No platform backend exists and nothing binds the stores yet (DEV-455). The pending-result store (DEV-563) holds results and bundles over the same port, and `ClearMeasurementData` deletes every record class through it behind a durable clearing marker (DEV-564); native reset wiring remains DEV-455.

[Observation and Network Context models](docs/observation-models.md) add bounded internal DNS/TCP/TLS/HTTP observations and immutable minimized access/egress snapshots bound to local path generations. The models adopt the merged public vocabulary contract and internal alignment recorded in [Specification Issues](docs/specification-issues.md#go-spec-issue-002-missing-registered-observation-vocabularies). Since DEV-644, runs on the direct path capture coarse access context through an internal platform capability, evaluate transport, VPN and validation conditions against it and attach it to results with egress unavailable ([Network Conditions and Access Context](docs/execution-models.md#network-conditions-and-access-context)). Egress lookup and caching and the proxy path exist since DEV-647 and DEV-645 ([Proxy Path](docs/execution-models.md#proxy-path)). TCP-01, DNS-01, TLS-01 and HTTP-01 run through `Client` with the production registry and execution-time destination checks; the remaining registered measurements are pending ([Measurement Requirements](docs/measurement-requirements.md)).

[Measurement Result models](docs/result-models.md) add internal typed attempt values, strict bounded serialization, independently selected versions, initial test-output validation and NET-05 partial-result handling. Runtime execution, bundling and the pending-result store exist ([Execution Models](docs/execution-models.md)); platform persistence and delivery remain pending (DEV-455).

[Result Bundle and linkage models](docs/bundle-models.md) add bounded internal bundle production from validated results, random logical-bundle identifiers, immutable retry bytes and optional rotating-linkage projections gated by trusted initialization. Token derivation, epoch lifecycle, persistence, delivery and clearing remain pending.

[Model validation and normalization](docs/model-validation.md) records the Foundation models' validation order and acceptance coverage. Common candidate validation preserves label invariants and deterministic ordering, context case normalization stays within ASCII codes, and uninitialized result decoders reject before processing input.

Executable, non-normative [sanitized API examples](internal/examples/README.md) cover existing Foundation models and intended lifecycle failure handling. Client lifecycle and clearing examples use an explicit test-only script; runtime implementation remains pending.

The root package exports a [`Client` lifecycle shell](docs/public-api.md#implemented-subset): `NewClient`, `Initialize`, `Close`, `ClearMeasurementData`, `AdmitTask`, `StartRun`, `RunState`, `CancelRun`, `NextBundle` and `AcknowledgeBundle`, with bounded handle tables and fixed operational error codes. `AdmitTask` runs the implemented admission steps and returns a task handle only on acceptance. Local policy is immutable per generation and replaced only by quiescing and reinitializing with a higher generation, through an unexported seam. No production adapter is wired, so `Initialize` fails closed. Clearing reports `clearing_incomplete` until persistent-state deletion exists. `StartRun` reserves one due run of an admitted task under the interim due-time rule of GO-SPEC-ISSUE-011 and `RunState` polls it. `CancelRun` stops a run and waits for its cleanup within the registered bound. Results are held in bounded pending bundles that `NextBundle` returns and `AcknowledgeBundle` deletes, expiring after at most 86,400 seconds under the interim rule of GO-SPEC-ISSUE-012. `SetProxy` and `ClearProxy` change the configured proxy generation. No runtime or platform support is claimed.

The [Foundation validation boundary](docs/foundation-validation.md) records which test infrastructure exists and which runtime harness extensions remain with later implementation tasks. Foundation was assessed and accepted on October 16, 2026 (see the coordination repository's assessment record); acceptance does not promote runtime, measurement or platform support.

## Versions and Releases

Repository versioning, compatibility, changelog, tag, artifact-naming, and experimental-release rules are defined in [Versioning and Releases](docs/versioning-and-releases.md). Notable changes are recorded in the [Changelog](CHANGELOG.md).

A Barometer Go release version is independent from a public specification version and from the evidence-backed support status of any measurement or platform. The documented release process requires those facts to be stated separately.

## Project Boundaries

The [Barometer coordination repository](https://github.com/tv-rain/barometer) owns project-wide principles, architecture decisions, delivery planning, and the task workflow. The [public specification](https://github.com/tv-rain/barometer-spec) is authoritative for observable library and measurement behavior. This repository owns only the Go implementation and its repository-local documentation, tests, and release history.

The host application remains responsible for enablement, task delivery (remote tasks arrive as signed packages that the Go core verifies against trusted keys in local policy), scheduling, deployment policy, optional proxy operation, result delivery, participant transparency, and deployment-specific review. The Go core must enforce admission, validation, bounded execution, registered behavior, minimized results, cancellation, retention, and clearing without accepting trust assertions from task-controlled data.

See [Package Architecture](docs/package-architecture.md) for the intended package ownership and dependency rules. The [Public Go API Boundary](docs/public-api.md) defines the intended facade and trust/lifecycle contract; its isolated signature fixture is not a production API or runtime implementation.

## Requirements

- Go 1.26.0 or a later compatible Go release, as declared by [`go.mod`](go.mod).
- Make for the canonical command aliases.
- A host supported by Go's race detector, with CGO and a working C compiler, is required for `make race` and the complete `make check` validation set. The approved CI environment provides this capability.
- No deployment credentials, participant data, private configuration, mobile SDK, or external service is required for the current validation commands.

## Validate a Clean Checkout

Run the complete local validation set from the repository root:

```sh
make check
```

The separate [traceability CI checks](docs/traceability-ci.md) use `make traceability` with an explicit public checkout, Python environment and report directory. They validate metadata/references and run synthetic conformance smoke cases while preserving unresolved gaps. They require Go 1.26.0 or later first on `PATH` and Python 3.10 or later with the pinned public tooling dependencies; the command fails fast with a prerequisite message otherwise. Follow [Reproduce Locally](docs/traceability-ci.md#reproduce-locally) for the install and run commands.

The individual commands are:

| Command | Purpose |
|---|---|
| `make build` | Compile all Go packages for the host platform. |
| `make format` | Apply `gofmt` to repository Go files. |
| `make format-check` | Fail if any Go file requires formatting without changing files. |
| `make test` | Run all Go tests. |
| `make race` | Run all Go tests with Go's race detector on supported hosts. |
| `make lint` | Run the standard Go static analyzer with `go vet`. |
| `make integrity` | Check module-file consistency, verify downloaded modules, and reject whitespace errors in the working diff. |
| `make mobile-smoke` | Cross-compile the current pure-Go packages for Android ARM64 and iOS ARM64. |
| `make check` | Run formatting, build, unit-test, race-test, static-analysis, repository-integrity, and mobile-compilation checks. |

`make mobile-smoke` is a compiler portability check only. It does not run on a device, build a native adapter, invoke gomobile, validate the future exported API boundary, or establish Android or Apple support.

The separate `make mobile-smoke-android` and `make mobile-smoke-apple` commands generate and compile gomobile bindings for the root package using a smoke-only exported function. They require native SDKs and are not part of `make check`. See [Mobile Compilation Smoke Checks](docs/mobile-smoke.md) for prerequisites, the minimal boundary decision, pinned tooling, negative tests, and evidence limitations. Their temporary artifacts are not release artifacts or platform-support evidence.

If a development host is not supported by Go's race detector, run the remaining individual checks locally and rely on the approved CI environment for the race-test result. Do not report the complete `make check` set as passed when the race test was skipped.

## Development

Read [CONTRIBUTING.md](CONTRIBUTING.md) before preparing a change. Automated or agent-assisted work must also follow [AGENTS.md](AGENTS.md).

Use synthetic or local fixtures for development and tests. Do not add credentials, participant data, measurement data, private deployment configuration, personal machine paths, or unrestricted diagnostic logs to the repository.

## License

The code and documentation in this repository are licensed under the [MIT License](LICENSE), unless otherwise noted.

The Barometer name, logo, and other brand assets are not licensed under the MIT License. All rights to the Barometer trademarks and brand assets are reserved by TVR Studios B.V..

See [TRADEMARKS.md](TRADEMARKS.md) for details.
