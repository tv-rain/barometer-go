<p align="left">
  <img src="assets/barometer-logo.svg" alt="Barometer" width="80">
</p>

# Barometer Go

Barometer Go is the platform-neutral core planned for the Barometer mobile network-measurement library. It is intended to admit reviewed tasks, enforce safety and resource bounds, execute registered measurements, and return minimized structured observations without requiring a particular scheduler, collector, backend, or host application.

## Current Status

This repository is at the engineering-foundation stage. It currently contains the Go module bootstrap, an accepted package-architecture plan, [shared strict JSON decoding infrastructure](docs/json-decoding.md), and [common task and schedule models](docs/task-models.md). Common task decoding and pure validation produce candidates using explicitly supplied test schemas; production test-specific schemas, complete admission, registered measurements, full runtime conformance, native adapters, mobile release artifacts, and a supported public API remain pending.

The Barometer `0.1.0` public specifications are Draft implementation inputs. A specification, package plan, successful Go build, or successful cross-compilation check is not evidence that a measurement or platform is implemented, conformant, field-validated, or supported.

The [conformance fixture format and runner](docs/conformance-fixtures.md) add versioned, strictly validated synthetic task-decoding cases with separate context/policy profiles and exact expected candidate projections or rejection classes. Admission, policy enforcement and measurement execution remain pending.

Reusable [negative and boundary-test helpers](docs/testkit.md) provide bounded synthetic fixtures, rejection diagnostics, and deterministic time/transport controls for future tests. They do not implement runtime schemas or measurements.

[Initialization and configuration models](docs/configuration-models.md) add internal immutable host configuration, atomic value replacement, trusted deployment approval inputs and narrow clock/protected-key capability contracts. Configuration preparation does not activate a client or replace the pending admission, execution and native adapter work.

[Trusted context and local policy models](docs/trusted-policy-models.md) provide separate internal source metadata, exact-byte delivery binding, immutable issuer/target policy snapshots and revocation inputs. They have no JSON import/export path and do not implement authentication, effective authorization, admission, replay, policy activation or revocation enforcement.

## Versions and Releases

Repository versioning, compatibility, changelog, tag, artifact-naming, and experimental-release rules are defined in [Versioning and Releases](docs/versioning-and-releases.md). Notable changes are recorded in the [Changelog](CHANGELOG.md).

A Barometer Go release version is independent from a public specification version and from the evidence-backed support status of any measurement or platform. The documented release process requires those facts to be stated separately.

## Project Boundaries

The [Barometer coordination repository](https://github.com/tv-rain/barometer) owns project-wide principles, architecture decisions, delivery planning, and the task workflow. The [public specification](https://github.com/tv-rain/barometer-spec) is authoritative for observable library and measurement behavior. This repository owns only the Go implementation and its repository-local documentation, tests, and release history.

The host application remains responsible for enablement, authenticated task delivery, scheduling, deployment policy, optional proxy operation, result delivery, participant transparency, and deployment-specific review. The Go core must enforce admission, validation, bounded execution, registered behavior, minimized results, cancellation, retention, and clearing without accepting trust assertions from task-controlled data.

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
