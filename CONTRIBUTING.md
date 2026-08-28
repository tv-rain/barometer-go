# Contributing to Barometer Go

Barometer Go is intended for safety-sensitive use on consumer mobile devices and potentially hostile networks. Contributions must preserve conservative execution, privacy, resource, and interpretation boundaries as well as ordinary code quality.

## Before Starting

For tracked Barometer work, follow the coordination repository's `docs/task-workflow.md`. Confirm the task outcome, acceptance criteria, dependencies, responsible repository, approval boundaries, and validation requirements before implementation.

Read [Package Architecture](docs/package-architecture.md) and the relevant public specifications before changing behavior. The public specification is authoritative for observable behavior; examples and internal review documents are not alternative schemas.

Work on a task-specific branch and keep changes limited to this repository unless an approved cross-repository plan explicitly requires otherwise. Inspect and preserve unrelated local work before editing.

## Change Scope

- Keep a contribution bounded to one coherent, independently reviewable outcome.
- Do not introduce public API, new exported packages, third-party dependencies, native SDK requirements, infrastructure, collectors, or deployment behavior without explicit reviewed scope.
- Do not create empty internal packages solely to reproduce the planned package tree.
- Record specification ambiguities instead of silently selecting new observable behavior.
- Create or propose a follow-up task for newly discovered work that can be completed independently.
- Keep semantic versioning, tagging, changelog, artifact naming, and release policy changes within their dedicated reviewed work rather than defining them incidentally.

## Safety and Test Data

Use deterministic synthetic or local fixtures. Tests must not require deployment secrets, participant data, measurement data, private endpoints, account identifiers, personal machine paths, or uncontrolled field measurements.

Measurements are observations with multiple possible explanations. Code, tests, comments, and documentation must not turn a protocol outcome into a censorship verdict or imply zero risk, perfect anonymity, conformance, field validation, or platform support without evidence.

## Local Validation

The canonical commands run from the repository root:

```sh
make format
make check
git diff --check
```

`make check` verifies formatting, host compilation, unit tests, standard Go static analysis, and pure-Go cross-compilation for Android ARM64 and iOS ARM64. See the [README](README.md#validate-a-clean-checkout) for individual commands and the limits of the mobile smoke check.

Add focused tests for new behavior and boundary or hostile-input tests where safety, validation, resource use, cancellation, persistence, or clearing is affected. A change is not complete merely because it compiles.

## Review and Handoff

Before requesting review:

- verify every acceptance criterion;
- run the documented validation commands;
- check the final diff for unintended or generated files;
- document known limitations and any required compatibility ordering;
- link durable implementation and validation evidence in the tracked task when applicable;
- avoid claims that exceed the available implementation, conformance, platform, or field evidence.

Never commit credentials, participant data, sensitive observations, private deployment configuration, unrestricted logs, personal paths, chat transcripts, or temporary agent files.
