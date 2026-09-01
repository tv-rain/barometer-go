<p align="left">
  <img src="assets/barometer-logo.svg" alt="Barometer" width="80">
</p>

# Barometer Go

Barometer Go is the platform-neutral core planned for the Barometer mobile network-measurement library. It is intended to admit reviewed tasks, enforce safety and resource bounds, execute registered measurements, and return minimized structured observations without requiring a particular scheduler, collector, backend, or host application.

## Current Status

This repository is at the engineering-foundation stage. It currently contains the Go module bootstrap and an accepted package-architecture plan; it does not yet implement registered measurements, conformance, native adapters, mobile artifacts, or a supported public API.

The Barometer `0.1.0` public specifications are Draft implementation inputs. A specification, package plan, successful Go build, or successful cross-compilation check is not evidence that a measurement or platform is implemented, conformant, field-validated, or supported.

## Project Boundaries

The [Barometer coordination repository](https://github.com/tv-rain/barometer) owns project-wide principles, architecture decisions, delivery planning, and the task workflow. The [public specification](https://github.com/tv-rain/barometer-spec) is authoritative for observable library and measurement behavior. This repository owns only the Go implementation and its repository-local documentation, tests, and release history.

The host application remains responsible for enablement, authenticated task delivery, scheduling, deployment policy, optional proxy operation, result delivery, participant transparency, and deployment-specific review. The Go core must enforce admission, validation, bounded execution, registered behavior, minimized results, cancellation, retention, and clearing without accepting trust assertions from task-controlled data.

See [Package Architecture](docs/package-architecture.md) for the intended package ownership and dependency rules.

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

If a development host is not supported by Go's race detector, run the remaining individual checks locally and rely on the approved CI environment for the race-test result. Do not report the complete `make check` set as passed when the race test was skipped.

## Development

Read [CONTRIBUTING.md](CONTRIBUTING.md) before preparing a change. Automated or agent-assisted work must also follow [AGENTS.md](AGENTS.md).

Use synthetic or local fixtures for development and tests. Do not add credentials, participant data, measurement data, private deployment configuration, personal machine paths, or unrestricted diagnostic logs to the repository.

## License

The code and documentation in this repository are licensed under the [MIT License](LICENSE), unless otherwise noted.

The Barometer name, logo, and other brand assets are not licensed under the MIT License. All rights to the Barometer trademarks and brand assets are reserved by TVR Studios B.V..

See [TRADEMARKS.md](TRADEMARKS.md) for details.
