# Foundation Validation Boundary

**Status date:** October 16, 2026

**Status:** Foundation infrastructure delivered, assessed and accepted on October 16, 2026. Runtime harness extensions remain with the deferred implementation tasks below.

## Scope and Authority

The approved [Foundation assessment boundary](https://app.clickup.com/t/86cbb831b) distinguishes reusable test infrastructure and concrete case plans from runtime enforcement and its executable validation. This document records Go-local capabilities and deferred ownership. It does not alter public requirements, close specification findings, promote Draft status, or establish measurement/platform support. Cross-repository status and the assessment package belong to the coordination repository.

The integrated original Foundation code baseline is [c8e3d21a112d7529d1027972a34cf12855cab7fa](https://github.com/tv-rain/barometer-go/commit/c8e3d21a112d7529d1027972a34cf12855cab7fa). This documentation preparation changes no production code, tests, public schemas or historical manifests/reports. The final preparation commit must be recorded separately in the tracker handoff.

## Delivered Infrastructure and Limits

| Capability | Executable evidence / canonical record | What remains outside that evidence |
|---|---|---|
| Raw hostile JSON and exact size boundaries | [Testkit](testkit.md): unknown/duplicate members, malformed text/values, nesting, bounded builders, maximum-minus-one/maximum/maximum-plus-one byte fixtures and requirement-aware rejection assertions | Helpers do not define production schemas or prove consumer rejection without a consumer-specific positive control and assertion. |
| Deterministic time inputs | Testkit manual clock and before/at/after time fixtures; task/schedule model tests | The clock has no timers and does not drive operating-system or context deadlines. |
| Scripted cancellation and body cleanup | Testkit in-memory HTTP transport tests cover pre-cancellation, waiting cancellation, body closure, exhaustion and prevention of copying past cancellation | No real socket, TLS, lower-layer buffering, runtime worker cleanup or native lifecycle evidence. |
| Strict decoding runner | [Versioned fixtures](conformance-fixtures.md) execute positive/negative common-task decoding with fixed synthetic context/policy profiles and exact expected projections/rejection classes | No production measurement schema, authentication, effective policy enforcement, admission or runtime measurement. |
| Model size and serialization boundaries | [Model validation](model-validation.md), [results](result-models.md) and [bundles](bundle-models.md) contain valid/invalid/boundary/round-trip cases and bounded producers | No integrated execution budget, protocol accounting, transport delivery, persistence, retention or clearing. |
| Intended cancellation/clearing usage | [Executable examples](../internal/examples/README.md) exercise pure state transitions and explicit scripted host call/error flow | Scripted calls delete nothing and prove no cleanup timing, idempotent privacy deletion, concurrency or interruption recovery. |
| Planned conformance | [Cross-cutting case families](cross-cutting-requirements.md), [measurement families](measurement-requirements.md), manifests and gap reports | Planned paths/cases and successful metadata validation are not executed conformance or semantic completeness. The manifests retain different immutable historical baselines. |

## Deferred Implementation and Harness Ownership

These tasks remain backlog and are not enabled for agent execution. Their implementation is outside Foundation readiness; recording the ownership and required validation is part of that readiness. Each future implementation task must introduce its required harness extension alongside actual behavior, rather than substituting a test-only success script.

| Follow-up | Intended owner and required evidence |
|---|---|
| [Deadline and cancellation cleanup](https://app.clickup.com/t/124kk1zc4a9) | Execution/protocol owners: injected timers, active-I/O cancellation, first termination, bounded cleanup and race/leak assertions. Native lifecycle validation remains separately scoped. |
| [Effective resource budgets](https://app.clickup.com/t/124kk1zc4ab) | Execution/protocol/result/bundle owners: lower-layer byte and connection accounting, buffering/partial failures, retries/composites and integrated output limits. Existing model size tests remain reusable but insufficient alone. |
| [Retention and pending-result lifecycle](https://app.clickup.com/t/124kk1zc4ac) | Persistence/bundle owners: deterministic expiry, acknowledgement deletion, capacity limits, failed operations, restart/recovery and prevention of late upload. Native durability requires additional platform evidence. |
| [Measurement-data clearing](https://app.clickup.com/t/124kk1zc4ae) | Lifecycle/persistence/state owners: fault injection for every clearing step, blocking new work, concurrent operations, interruption/recovery, repeated calls, invalidation, no network/new secrets and explicit incomplete deletion. Host-owned copies and OS reset/storage wiring require separate evidence. |
| [Residual allocation and semantic mapping review](https://app.clickup.com/t/124kk1zc4af) | Public maintainers allocate/disposition residual normative sources; Go owners adopt reviewed mappings and immutable baselines. Inventory/semantic completeness requires review beyond structural checks. |
| [Downstream strict-JSON adoption](https://app.clickup.com/t/124kk1zc4ag) | Schema, admission, measurement and binding/host consumers: caller-specific adoption and atomic/no-side-effect integration evidence for GO-SPEC-ISSUE-001. |
| [Runtime observation vocabulary adoption](https://app.clickup.com/t/124kk1zc4aj) | DNS/TLS runtime producer and binding owners: actual protocol/API normalization, result integration and compatibility evidence for GO-SPEC-ISSUE-002. |

Native-specific work is linked to the mobile integration epic in the tracker. A Go fixture cannot establish native storage, suspension, routing, permissions, reset or host-owned deletion behavior. Individual assignees remain for workstream planning; role/package ownership and implementation dependencies are recorded in each task.

## Reproduction and Assessment

Use `make check` for the complete ordinary Go validation set, and the separate [traceability command](traceability-ci.md) with explicitly supplied public checkout, pinned dependencies and an output directory outside source repositories. Invoke the selected Go 1.26 toolchain directly through PATH for subprocesses too: the public tooling tests deliberately set `GOTOOLCHAIN=local` when compiling their selector fixture. An older launcher on PATH would otherwise make those subprocesses fail despite automatic toolchain selection by the main module; the driver now detects this, and a missing Python `jsonschema` package, before starting and exits with a prerequisite message. Install the pinned dependencies and run the exact commands in [Reproduce Locally](traceability-ci.md#reproduce-locally).

Use [mobile build commands](mobile-build.md) and the [isolated API signature fixture](public-api.md#validation-and-implementation-gates) only for their documented compiler/binding scopes. Exact-baseline hosted runs, historical local evidence and fresh checks must be labelled separately. No native build is device or runtime conformance evidence.

Preserve historical committed traceability reports. Fresh reports must retain every gap, selected baseline and review limit. The coordination assessment record and tracker validation task retain the actual outcomes and available artifact references; this capability document does not declare a check passed merely because a command is documented.

## Change History

| Date | Change |
|---|---|
| October 16, 2026 | Recorded delivered Foundation tooling, deferred runtime harness ownership and separate-assessment evidence boundaries without changing runtime behavior or historical reports. |
| October 16, 2026 | Recorded Foundation acceptance after the assessment corrections (DEV-568 to DEV-573). Documentation only; deferred ownership, runtime behavior and support status are unchanged. |
