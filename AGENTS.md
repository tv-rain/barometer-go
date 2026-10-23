# Barometer Go Repository Instructions

## Scope

These instructions apply to the `barometer-go` repository. Follow the Barometer coordination repository's project-wide `AGENTS.md` as well when working from the coordinated workspace. If the two sets of instructions differ, this file controls repository-local implementation details while accepted cross-project safety, privacy, architecture, and workflow decisions continue to apply.

Before changing this repository, read:

- `README.md`;
- `CONTRIBUTING.md`;
- `docs/package-architecture.md`;
- the coordination repository's `docs/task-workflow.md` for tracked work;
- the relevant project principles, architecture decisions, and public specifications for the behavior being changed.

## Authority and Repository Boundary

The public `barometer-spec` repository is authoritative for observable library and measurement behavior. This repository implements that contract; it must not silently redefine, weaken, or extend it.

The `barometer-go` repository owns Go source, repository-local design, tests, build instructions, implementation evidence, and its independent release history. Do not place public specifications, internal governance records, native or Flutter integration code, host-application code, deployment configuration, or collector infrastructure here.

Keep changes within this independent Git repository. Do not treat the coordination workspace and its child repositories as one worktree or one commit.

## Architecture

Preserve the ownership and dependency rules in `docs/package-architecture.md`.

- Keep the module root as the only initially intended public Go package.
- Do not export a new package, interface, method, or serialized representation without an explicitly reviewed public-API task.
- Create an internal package only when its first real behavior and tests establish one coherent owner; do not scaffold the package map with empty placeholders.
- Keep domain models free of I/O, scheduling, persistence, logging, and platform calls.
- Keep task decoding, validation, admission, execution, observations, result construction, and interpretation separate.
- Do not hide dependency cycles through global registries, service locators, untyped maps, generic callback collections, or utility packages.
- Do not add a mandatory scheduler, collector, backend, cloud service, analytics SDK, or host-application dependency.
- Treat new third-party dependencies, native toolchains, mobile SDK requirements, and public API changes as explicit review points rather than incidental implementation details.

When implementation exposes an ambiguity or requires behavior that differs from the Draft public specification, stop and record the issue. Follow the documented cross-repository order: decision when required, public specification, internal review alignment, Go implementation, then native or Flutter integration.

## Safety and Privacy

Assume hostile networks, possible surveillance, and participants in repressive environments. User safety takes priority over measurement coverage.

- Treat every measurement as an observation, not proof of censorship.
- Keep execution state, protocol observations, and derived interpretation distinct.
- Reject invalid or unsafe tasks atomically; do not silently clamp values, drop invalid steps, substitute targets, broaden behavior, or fall back from proxy to direct.
- Keep authenticated source context, issuer authority, local policy, target registration, proxy configuration, protected-storage capability, and participant identity outside task-controlled values.
- Preserve explicit cancellation, time, byte, connection, result-size, retention, and clearing behavior.
- Collect and retain only fields required by the reviewed specification.
- Never add user, account, advertising, device, SIM, subscriber, precise-location, or unrelated application identifiers without an explicit reviewed specification and required approvals.
- Never claim zero risk, perfect anonymity, perfect detection, or implementation/platform support without the required evidence.

Use sanitized synthetic or local fixtures by default. Tests and examples must not require deployment secrets, participant data, private endpoints, sensitive logs, personal machine paths, or uncontrolled network measurement.

## Working Practices

Before editing:

1. Confirm the task is eligible and ready under the coordination task workflow.
2. Inspect the current branch, working tree, relevant history, and existing files.
3. Preserve unrelated changes.
4. Identify the controlling public requirements and applicable architecture rules.
5. Confirm that dependencies and required approvals are satisfied.

Keep each change bounded to one coherent, independently reviewable outcome. Propose a related follow-up task instead of silently absorbing unrelated work. Do not change specifications or sibling repositories as part of a Go task unless the approved scope explicitly requires the cross-repository change and its ordering is documented.

Use precise, conservative language in code, tests, and documentation. Draft behavior is not released support, synthetic validation is not field validation, and successful compilation is not platform support.

Keep `docs/` current-state and update traceability manifests only at milestones, as `CONTRIBUTING.md` describes; record task history in the changelog, commits and the tracker.

## Canonical Commands

Run commands from the repository root:

- `make build` compiles all Go packages.
- `make format` applies Go formatting.
- `make format-check` checks formatting without editing files.
- `make test` runs the test suite.
- `make race` runs the test suite with Go's race detector on supported hosts.
- `make lint` runs standard Go static analysis.
- `make integrity` checks module consistency and verification and rejects whitespace errors in the working diff.
- `make mobile-smoke` cross-compiles the current pure-Go packages for the selected mobile targets.
- `make check` runs the complete non-mutating local validation set.

The `mobile-smoke` command checks pure-Go portability. The separate `mobile-smoke-android` and `mobile-smoke-apple` commands compile gomobile bindings and require native toolchains. Follow [Mobile Compilation Smoke Checks](docs/mobile-smoke.md), including its negative tests, when changing the binding smoke checks or their tooling. None of these commands establishes device, lifecycle, or platform support.

The race check requires a host supported by Go's race detector, CGO, and a working C compiler. If the local host cannot run it, execute the remaining individual checks and use the approved CI result for race-test evidence; do not claim that the complete `make check` set passed locally.

Before handing off a change, run validation in proportion to risk and at minimum run `make check` and `git diff --check`. Confirm that the working tree contains no generated build artifacts, secrets, participant data, private configuration, personal paths, chat transcripts, or temporary agent files.
