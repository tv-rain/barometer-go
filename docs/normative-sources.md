# Normative Source Inventory

## Purpose and Authority

This is the Go implementation's document-level inventory of authoritative public specification inputs. It is a navigation and baseline record, not a second specification, a complete requirement-to-code mapping, or conformance evidence. The linked public text controls observable behavior, including field tables, validation rules, limits, failure states, lifecycle, privacy, safety, and interpretation constraints. Requirements without allocated IDs remain authoritative.

Read each applicable document in full before implementing its behavior. A document may contain both normative rules and illustrative examples; its presence here does not make every example or explanatory paragraph normative. Use [Requirement References](requirement-references.md) for individual obligation references and [Package Architecture](package-architecture.md#specification-traceability) for planned ownership.

## Reviewed Baseline

| Property | Recorded value |
|---|---|
| Public repository | [tv-rain/barometer-spec](https://github.com/tv-rain/barometer-spec) |
| Immutable revision | [`95dfeef1367f4c320bca7c7ae91b13ca7608b37f`](https://github.com/tv-rain/barometer-spec/commit/95dfeef1367f4c320bca7c7ae91b13ca7608b37f) |
| Applicable specification version | `0.1.0`, as declared by all 27 contract documents below |
| Specification status | Draft; Phase 1 acceptance does not establish implementation, conformance, platform support, or field validation |
| Inventory date | October 13, 2026 |
| Review scope | Public source discovery, declared versions, authority classification, local references, and identified source ambiguities; not a comprehensive semantic audit |

All public contract links below are pinned to this revision. They do not silently follow `main`. The public specification version is independent of the Go release version, task/result schema versions, individual test versions, result versions, and classification-data versions. Those fields retain their own specified semantics even where their current values coincide.

## Public Runtime Contracts

Every row uses the shared Draft `0.1.0` baseline above. Applicability describes where the Go implementation must consult the source; it does not assert that the area is implemented. Platform and host obligations in these documents still constrain the core's capability boundaries and failure behavior.

### Data Formats — 11 Sources

| Public source | Go applicability |
|---|---|
| [Task Description Format](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/data-formats/task.md) | Serialized task envelope, required fields, exact version selection, unknown-field rejection, task identifiers, and complete-task size. |
| [Input Format](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/data-formats/task-input.md) | Object shape, registered test-specific schemas, target normalization, input bounds, and validation. |
| [Options Format](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/data-formats/task-options.md) | Required options object, test-specific options/defaults, bounds, and validation. |
| [Schedule Format](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/data-formats/task-schedule.md) | Run eligibility, expiry, repetition, jitter, retries, and run/attempt accounting under host orchestration. |
| [Execution Format](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/data-formats/task-execution.md) | Direct/proxy selection, execution constraints, effective resource limits, cancellation, and execution failure reasons. |
| [Safety Format](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/data-formats/task-safety.md) | Untrusted risk/consent declarations and evaluation against registered risk and trusted local policy. |
| [Labels Format](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/data-formats/task-labels.md) | Optional bounded annotations, key/value validation, reserved namespace, duplicate-key rejection, and prohibition on behavioral control or participant identifiers. |
| [Measurement Observation Formats](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/data-formats/measurement-observations.md) | Shared bounded protocol/platform observation structures, status/failure fields, ordering, timing, and minimized details. |
| [Measurement Result Format](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/data-formats/measurement-result.md) | Per-attempt envelope, identifiers and versions, execution state distinct from measurement outcome, timing, partial results, and size. |
| [Network Context Format](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/data-formats/network-context.md) | Normalized access/egress objects, coarse enums, selected-path metadata, and excluded identifying fields. |
| [Result Bundle Format](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/data-formats/result-bundle.md) | Bounded result grouping, bundle identifiers, optional linkage representation, delivery idempotency, local retention, and clearing. |

### Library Contracts — 5 Sources

| Public source | Go applicability |
|---|---|
| [Task Trust and Local Policy](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/library/task-trust.md) | Trusted source context, issuer authorization, target registration, atomic admission, exact-byte replay identity, expiry, fail-closed errors, and policy lifecycle. |
| [Registered Initial Limits](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/library/registered-limits.md) | Canonical numeric defaults/maxima across tasks, schedules, attempts, protocols, composite profiles, results, bundles, context, linkage, and retention. Stricter specified/local limits also apply; excessive task requests are rejected rather than clamped. |
| [Proxy Integration](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/library/proxy.md) | Host-configured HTTP CONNECT proxy, supported test/path combinations, path snapshots and generation changes, privacy filtering, and no silent direct fallback. |
| [Network Context](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/library/network-context.md) | Platform access collection, endpoint configuration/response contract, selected-path acquisition, bounded fallback, cache lifecycle, and execution integration. |
| [Result Linkage and Private State](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/library/result-linkage.md) | Default-off rotating linkage, protected platform capabilities, derivation, rotation, persistence, retention, and complete clearing. No portable or mobile fallback may weaken this contract. |

### Initial Measurements — 9 Sources

Each test document supplies its own inputs, options, preconditions, resource bounds, Output schema/outcomes, interpretation, and privacy/safety constraints in addition to the shared contracts. Test/result version selection is governed by those sources, not inferred from this inventory.

| Public source | Go applicability |
|---|---|
| [DNS-01: DNS resolution](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/tests/dns-01.md) | System or direct DNS resolution and bounded matching-response observations. |
| [DNS-03: DNS-over-HTTPS resolution](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/tests/dns-03.md) | One bounded authenticated HTTPS DNS exchange with a reviewed resolver and query. |
| [DNS-04: DNS-over-TLS resolution](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/tests/dns-04.md) | One bounded authenticated TLS DNS exchange, without cleartext/resolver fallback. |
| [TCP-01: TCP connect](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/tests/tcp-01.md) | One connection to a registered public IP address and port, without application data. |
| [TLS-01: TLS handshake](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/tests/tls-01.md) | One authenticated handshake with an explicit address/server name and bounded negotiated properties. |
| [HTTP-01: Small HTTPS request](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/tests/http-01.md) | Bounded unauthenticated HTTPS GET, registered checks, and validated optional redirects. |
| [NET-01: Network state](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/tests/net-01.md) | Coarse local platform state and transitions; no network traffic or egress classification. |
| [NET-03: Network validation and captive portal](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/tests/net-03.md) | Platform validation context and one bounded request to a reviewed validation endpoint, without redirects. |
| [NET-05: Multi-endpoint reachability profile](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/tests/net-05.md) | Atomic admission and sequential DNS-01/TCP-01/TLS-01/HTTP-01 primitive steps, aggregate bounds, shared context/path, and ordered results without a combined censorship verdict. |

### Cross-cutting Safety and Interpretation — 2 Sources

| Public source | Go applicability |
|---|---|
| [Threat Model](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/threat-model.md) | Security/privacy assumptions, trust boundaries, required mitigations, prohibited behavior, and residual risks across admission, execution, state, and result handling. |
| [Measurement Limitations](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/measurement-limitations.md) | Cross-cutting interpretation, missingness, comparability, and reporting guardrails; results remain observations rather than censorship proof. |

## Other Source Classes

| Source class | Authority and use |
|---|---|
| [Public requirement convention](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/requirements/README.md), [catalog](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/requirements/catalog.json), and [checker documentation](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/scripts/README.md) | Development rules and identity/location/lifecycle metadata. The marked public obligation supplies runtime authority. Catalog/checker success does not establish exhaustive coverage or conformance. |
| Public root/directory READMEs | Discovery indexes, not replacement contracts. Follow their links to the applicable source. |
| Public `tests/template.md` | Authoring structure for future tests, not a registered test or additional executable behavior. |
| JSON, Dart, and other examples inside specifications | Illustrations subordinate to field tables and prose; not executable deployment configuration or an alternative schema. |
| Public `AGENTS.md`, tooling/tests, licenses, trademarks, and assets | Contribution, development, licensing, or branding material; not measurement runtime contracts. |
| Coordination principles, architecture, accepted decisions, and task workflow | Project/process constraints that must be followed. Observable runtime requirements still belong in the public specification; flag a conflict instead of selecting a competing contract. |
| Coordination work plans, backlog, status records, and Phase 1 delivery manifest | Planning, status, and historical delivery evidence. The Phase 1 revision is a historical baseline, not a substitute for this selected revision. |
| Internal registry/prefix allocation, per-test companions, review checklists, validation plans, and open questions | Governance, review, and planning inputs. They cannot add private runtime obligations or establish conformance through checklist presence. Public implementation must not require access to them. |
| Go [package architecture](package-architecture.md), [requirement guidance](requirement-references.md), release/contribution documents, this inventory, and [specification issues](specification-issues.md) | Implementation ownership, process, navigation, and findings. They do not redefine the public contract. Planned packages and proposed resolutions are not implemented behavior. |
| ClickUp task descriptions and evidence | Work coordination and acceptance evidence; cannot override a public specification or accepted decision. |

## Completeness, Issues, and Maintenance

The inventory covers every contract file in the selected public revision's `data-formats/` and `library/` directories, all nine registered initial tests, and the two cross-cutting public contracts. Directory indexes and local link targets were checked for additional sources. No missing contract file was found in that source-discovery scope. The partial requirement catalog was not used as a completeness filter.

The identified decoding ambiguity is recorded in [Specification Issues](specification-issues.md#go-spec-issue-001-duplicate-json-member-names-outside-labels). This record is not a public resolution and does not authorize implementation choices. Future findings must remain visible rather than becoming local rules.

When adopting a new public revision, review added, removed, moved, and changed sources; update the shared immutable baseline and all pinned links together; recheck declared versions/statuses and source classifications; and reassess affected issue records and requirement references. Do not treat an unchanged `0.1.0` label as proof that the source text is unchanged. Record conformance and implementation evidence separately when those tasks exist.

Resolve observable-behavior ambiguities through the existing public-specification review procedure before implementing the affected behavior. A required cross-project decision precedes public changes; internal review alignment and Go adoption follow those reviewed changes. This inventory changes only Go documentation and introduces no runtime compatibility or release ordering requirement.
