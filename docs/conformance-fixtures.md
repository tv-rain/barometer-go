# Conformance Fixture Format and Runner

## Scope and Authority

DEV-470 adds a Go-local development format and the test-only `internal/testkit/conformance` runner. Format version `1` and adapter `common-task-v1` exercise the existing task decoder with an explicitly synthetic `SYN-01` codec. They do not register a measurement, implement admission or policy enforcement, or create a public task/result/API representation. Other implementations are not required to adopt this format. There is no cross-repository change or merge ordering.

The public baseline is Draft `0.1.0` at immutable revision [`05c1364af4f9a604b617fca45f060949b557d265`](https://github.com/tv-rain/barometer-spec/commit/05c1364af4f9a604b617fca45f060949b557d265): [runtime JSON decoding](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/data-formats/json-decoding.md), [common task fields](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/data-formats/task.md), [schedule](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/data-formats/task-schedule.md), [execution](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/data-formats/task-execution.md), and [registered limits](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/library/registered-limits.md). Public text controls runtime behavior; the fixture wrapper is outside that runtime JSON contract.

## Version 1 Schema

The executable closed schema is `fixtureSchema` in [`fixture.go`](../internal/testkit/conformance/fixture.go). `Parse` applies bounded strict JSON validation followed by semantic validation before returning an immutable fixture. Unknown fields, duplicate decoded names, invalid text, trailing content, missing/null/wrong-type fields, noncanonical integers, unsupported versions/profiles and invalid expectations reject atomically with `ErrFixture`.

Every top-level field below is required:

| Field | Contract |
|---|---|
| `format_version` | Integer `1`; independent of runtime/specification/module versions. |
| `case_id` | Unique within the corpus; lowercase letter followed by at most 63 lowercase letters, digits or hyphens. Used as the Go subtest name. |
| `adapter` | Exact string `common-task-v1`. No latest/compatible fallback. |
| `requirements` | 1–64 distinct IDs from the pinned initial allowlist: `REQ-TASK-0001` and `REQ-JSON-0001` through `REQ-JSON-0011`. This is an assertion reference, not complete coverage. |
| `task_base64` | Canonical padded standard base64 containing 0–65,537 exact task bytes. Empty/malformed/oversized runtime input can be a negative case. |
| `trusted_context` | Exact synthetic profile reference `remote-fixed-v1`, defined below. |
| `local_policy` | Exact synthetic profile reference `synthetic-local-v1`, defined below. |
| `expected` | Closed object with required `rejection` and optional `normalized`; exactly one outcome. |

For success, `rejection` is the empty string and `normalized` is required. For rejection, omit `normalized` and use one exact class: `syntax`, `text`, `duplicate`, `field`, `type`, `required`, `integer`, `range`, `version`, `bounds`, `task`, `timing`, or `parameters`. An unrelated error fails the case. A rejected decode must also return a nil candidate.

`normalized` is a closed candidate projection with all fields required: strings `task_id`, `subject`, `network_path`; boolean `enabled`; and exact integers in 0–262,144 named `duration_ms`, `bytes_sent`, `bytes_received`, `connections`, `max_runs`, `max_attempts`. Comparison is exact for every projected field. The finite integer envelope is tooling representation, not permission for runtime values outside public bounds. This projection does not assert the candidate's complete schedule, safety, labels, provenance or execution behavior.

The complete fixture is capped at 1 MiB, 16 value levels and 4,096 value nodes. Runtime task decoding retains its own 65,536-byte maximum; the additional fixture byte permits maximum-plus-one tests. Base64 preserves duplicate names, escape spellings, malformed JSON, invalid UTF-8 and whitespace. No task parsing or reserialization occurs before the consumer decoder.

## Synthetic Profiles and Adapter

`remote-fixed-v1` constructs a structurally valid `SourceContext` with remote origin, issuer `synthetic-issuer`, policy generation `1`, authenticated assertion `true`, and issuance time `2026-10-15T00:00:00Z`. The accepted clock is the same fixed instant. These values are assertions by an isolated test adapter, not real authentication evidence. The runner reads no wall clock, process timezone, secrets or participant data.

`synthetic-local-v1` constructs a `LocalPolicy` at generation `1`, with the same issuer and deployment/issuer permissions for synthetic `SYN-01` version `0.1.0`: direct path only, low risk, no registrations, label namespaces, consent requirement or additional opt-ins. Its explicit limits are:

| Limits | Values |
|---|---|
| Attempt duration / sent / received / connections | 30,000 ms / 32,768 bytes / 131,072 bytes / 8 |
| Runs / attempts / validity | 1 / 1 / 86,400 seconds |
| Minimum repeat interval / jitter | 21,600 / 300 seconds |
| Minimum / maximum retry delay / maximum jitter | 30,000 / 30,000 / 10,000 ms |
| Result bytes / bundle bytes / results per bundle | 65,536 / 262,144 / 16 |
| Pending bundles / bytes / retention | 32 / 4,194,304 / 86,400 seconds |

The runner checks construction, issuer lookup and generation association. Policy is included as a separate immutable test context; it is not passed to the common decoder, which does not enforce authorization or local-policy intersections. No passing fixture establishes policy enforcement. The runtime models retain their JSON import/export prohibition; these profile references are not a runtime trust or policy wire schema. See [trusted policy models](trusted-policy-models.md).

`common-task-v1` supplies the exact synthetic schema: required string `input.subject`, optional boolean `options.enabled`, with lowercase subject normalization and enabled default `true`. It delegates common task decoding, defaulting and validation to `internal/taskdecode` and supplies the fixed timing context separately. Synthetic parameter normalization is tooling behavior, not a normative measurement contract.

## Run and Extend

Run from the Go repository root:

```sh
go test ./internal/testkit/conformance
make check
```

The existing `make test`, race checks and CI discover the runner tests through `go test ./...`; no extra service, dependency, sibling checkout or native SDK is required. The corpus is loaded in sorted filename order, requires unique case IDs, and runs each parsed case repeatedly and concurrently. Go's test exit status provides the CI pass/fail result; no timestamps or raw input/error text enter diagnostics. Requirement IDs and case IDs are trusted test-author metadata and must remain sanitized.

The corpus in [`testdata`](../internal/testkit/conformance/testdata) includes positive default, explicit-false and explicit-limits cases and rejection cases for unknown fields, escaped duplicate names, trailing JSON, unsupported test versions, integer exponent spelling, null options and invalid UTF-8.

DEV-571 adds REQ-JSON-0009/0010 boundary cases at maximum-minus-one, maximum and maximum-plus-one for the task decoder's local parse limits of 65,536 bytes, 128 value levels and 32,768 value nodes:

| Cases | Below and at the limit | Above the limit |
|---|---|---|
| `bytes-*` | Valid task padded with trailing whitespace to 65,535 and 65,536 bytes; normalizes like `defaults` | 65,537 bytes; `bounds` |
| `depth-*` | Valid task plus an unknown nested-array member at 127 and 128 levels; `field` | 129 levels; `bounds` |
| `nodes-*` | Top-level array of 32,767 and 32,768 values; `type` | 32,769 values; `bounds` |

The closed task schema has no member that can hold 128 levels or 32,768 values, so the depth and node cases at and below the limit cannot be valid tasks. Their schema-level rejection class shows that the parse bound did not fire. The node limit also cannot be reached independently of the byte limit: the smallest value plus its separator takes two bytes, so 32,769 values need at least 65,537 bytes. `nodes-max-plus-one` therefore rejects at the byte bound before node counting, and the node check itself is covered by the configurable-limit unit tests in `internal/jsondecode`.
Runner tests independently verify malformed fixture rejection, bounds, exact-byte preservation, storage isolation and that wrong outputs or wrong rejection classes fail.

Start new cases with a positive control and change one property. Do not use real participant data, production targets, deployment configuration or secrets. Add a file under `testdata` with a unique case ID and review its requirement/assertion relationship. Unknown IDs require deliberate allowlist/baseline review. A profile, adapter or projection change must not silently reinterpret existing version-1 fixtures: add a new exact profile/adapter identifier or bump the format when its structure/meaning changes. Full schema/measurement/admission adapters remain separate implementation tasks.

## Evidence Limits

These cases provide synthetic regression evidence for the named assertions. They do not establish complete requirement coverage, production test schemas, authentication, admission, replay, effective policy enforcement, measurement results, device/platform support, field validation or release support. No production traceability manifest or support-status promotion is added; future mapping must record immutable specification and implementation baselines and reviewed case coverage under the [traceability convention](traceability-manifest.md).
