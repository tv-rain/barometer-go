# Specification Issues

## Scope and Baseline

This record holds Go implementation findings that need public specification clarification. It is not a local specification or an approval to choose behavior. Issue identifiers below belong to this Go findings record; they are not public requirement IDs, conformance-case IDs, or ClickUp task IDs.

Sources are pinned to public specification Draft `0.1.0`, revision `95dfeef1367f4c320bca7c7ae91b13ca7608b37f`, as recorded in the [Normative Source Inventory](normative-sources.md#reviewed-baseline). The source-discovery review found no missing contract file. This record is not an exhaustive semantic audit of all requirements.

## Recording and Reviewing Findings

Follow the public [workflow](https://github.com/tv-rain/barometer-spec/blob/main/specification-issues/README.md) and [template](https://github.com/tv-rain/barometer-spec/blob/main/specification-issues/template.md). Retain local finding IDs and historical source links. Cite allocated requirement IDs; for unallocated obligations cite exact immutable sections without inventing an ID. Stop affected implementation pending public disposition; unrelated approved work may continue.

Public maintainers resolve the contract. Merge changed public wording before required internal alignment and Go adoption, then bindings and host integration follow. Record the accepted immutable baseline, version/compatibility decision, implementation revisions, conformance cases/results, and release order before closing an accepted finding. No-text-change clarification requires public disposition citing the existing rule. Proposals and passing local tests are not acceptance.

The open finding below exercises the template without choosing decoding behavior. Merge the coordination decision and public procedure before this consumer update; its workflow links to public `main` become available after that merge. The procedure is development guidance outside the historical runtime baseline above.

## GO-SPEC-ISSUE-001: Duplicate JSON Member Names Outside Labels

**Status:** Public contract merged; shared Go decoder adoption prepared under DEV-477, with downstream model/admission/binding evidence pending. The historical finding below records the pre-resolution baseline and is retained unchanged; the adoption addendum supersedes its implementation-blocking statements.

### Identity, Ownership, and Affected Versions

Local identity remains `GO-SPEC-ISSUE-001`. Public proposal/PR and maintainer assignment are pending. The Go implementation owner prepares the proposal; public maintainers own disposition. No duplicate is identified in this record. Related work is deterministic task decoding; no implementation task or conformance case is created here.

Affected task/trust contracts are Draft `0.1.0` at the immutable public revision above. Exact sections are linked below. Broader duplicate-member semantics have no allocated requirement ID at that baseline; `REQ-TASK-0001` governs unknown fields and does not settle this question. Decoding is unimplemented, so no affected Go release or platform behavior is established; platform impact remains unvalidated.

### Public Evidence

- [Task Description Format](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/data-formats/task.md#top-level-fields) defines required fields and rejects unknown top-level fields but does not state how repeated member names are handled.
- [Labels Format — Keys](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/data-formats/task-labels.md#keys) explicitly rejects duplicate keys even if a parser would retain only one value. That rule is stated for the labels object.
- [Task Trust — Admission Procedure](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/library/task-trust.md#admission-procedure) requires complete validation of the top-level schema and nested task objects before scheduling.
- [Task Trust — Replay, Duplicate, and Rollback Handling](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/library/task-trust.md#replay-duplicate-and-rollback-handling) defines replay identity using a digest of the exact accepted bytes. This settles byte identity for redelivery, not repeated-name decoding semantics.

### Ambiguity and Impact

The reviewed sources do not explicitly extend the labels duplicate-key rule to the top-level task or other nested objects. They do not specify a selection/merge rule for those objects either. A synthetic task fragment such as `{"test_id":"NET-01","test_id":"TCP-01"}` therefore raises an unresolved admission question; it is not a valid complete task example.

This affects deterministic decoding and consistency between host validation and Go admission. The exact-byte replay digest does not resolve which value a decoder should validate. No behavior is implemented or selected by this issue record.

### Impact, Review, and Interim Handling

Conflicting duplicate members could make host and library validation select different measurement targets or behavior. Safety/trust implications therefore require public maintainer and admission/safety review; platform/parser differences need validation. This is a potential impact, not evidence of an implemented vulnerability. No participant collection, deployment, or legal conclusion is introduced by this finding. Deployment-specific gates remain applicable if the resolution changes those flows.

Affected decoding work remains blocked. Preserve the explicit labels rejection rule; do not extrapolate it or select first/last-member precedence elsewhere. Source discovery and unrelated approved work may continue. The fragment above is synthetic; no runtime reproduction or conformance result is claimed.

### Alternatives and Proposed Public Resolution

Alternatives for public review include rejecting duplicate decoded member names throughout tasks, defining explicit deterministic selection semantics, or retaining current text with a maintainer clarification of an existing rule. Whole-task rejection could avoid parser disagreement but its scope and escaped-name equivalence require review. Selection semantics could preserve parser conventions but require host/library consistency and safety analysis. Retaining ambiguity leaves affected decoding blocked.

Exact wording and a preferred alternative are pending maintainer review; none is adopted here. Review must cover top-level/nested objects and equivalent escaped names, distinguish duplicate-member handling from unknown-field rules, and record whether wording changes existing behavior or clarifies it. Allocate IDs and decide versions/compatibility through the public convention, not this Go record.

### Required Resolution

Public maintainers must clarify the scope and handling of duplicate member names in serialized tasks, including nested objects, and define whether differently escaped spellings of the same decoded name count as duplicates. Record the reviewed public revision and version decision here before adopting the behavior in Go. Any subsequent conformance cases belong to the decoding implementation task.

### Completion Evidence

Pending public clarification and later implementation/conformance evidence. The existing labels rejection rule remains authoritative while this broader question is open.

### Ordering and Adoption Evidence

A coordination decision is required only if resolution changes architecture or responsibility boundaries; none is currently proposed. Public disposition/PR, merged revision, and version decision are pending. After public resolution, align required internal admission/validation reviews, implement Go decoding and synthetic cases, then update affected bindings/host validation and record release compatibility. Owners and downstream task/PR links must be assigned during review.

Planned validation covers duplicate names at top level and nested objects, equal/conflicting values, escaped-name equivalence, and independent unknown-field rejection. Expected duplicate-member outcomes depend on the accepted contract; the existing labels rejection rule remains the baseline. Cases must check deterministic host/Go interpretation and absence of execution for rejected tasks. No case has been implemented or run for this finding.

Closure evidence remains pending: accepted public baseline, review alignment, implementation commits, conformance results, compatibility/release order, and any follow-ups. This template alignment does not close or resolve `GO-SPEC-ISSUE-001`.

### Adoption Addendum — October 14, 2026

Public [PR #5](https://github.com/tv-rain/barometer-spec/pull/5) merged the approved [Runtime JSON Decoding](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/data-formats/json-decoding.md) contract and [public disposition](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/specification-issues/strict-json-decoding.md) at immutable revision `05c1364af4f9a604b617fca45f060949b557d265`, retaining Draft `0.1.0`. Public rejection rules now cover every object and escaped-name equivalence. The earlier public-clarification blocker is resolved; the original finding and historical source inventory remain evidence of their original baseline.

DEV-477 aligns internal task-trust/resource validation plans before Go adoption, without marking specialist checklists complete. The shared implementation in `internal/jsondecode` and its [consumer contract and test map](json-decoding.md) adopt REQ-JSON-0001 through REQ-JSON-0011. This is shared infrastructure only: individual task, result, configuration and other schemas, admission/no-network behavior, host/binding agreement, producer migration, and release support are not established. GO-SPEC-ISSUE-001 remains open for applicable downstream evidence or explicit non-applicability; infrastructure tests do not close it.

Merge internal alignment before the dependent Go change. No public wording or version change is needed. Work is prepared on `agent/dev-477-strict-decoding-infrastructure`; implementation commits and human PR links await final approval and handoff. Later consumer claims must pin their own reviewed model and local-bound baselines and retain exact task bytes for integrity/replay checks.
