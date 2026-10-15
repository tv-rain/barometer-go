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

## GO-SPEC-ISSUE-002: Missing Registered Observation Vocabularies

**Current state:** Public resolution and required internal alignment are merged and verified. DEV-481 Go model adoption and local validation are prepared; final implementation approval, commit/push and applicable downstream evidence remain pending. The original finding below is historical evidence.

### Identity and Ownership

- **Disposition:** Open, October 15, 2026; no public resolution or downstream adoption.
- **Discovering role:** Go implementation agent for [DEV-481 — Implement observation and Network Context models](https://app.clickup.com/t/86cbb8f2n). Public maintainer/reviewer assignment and public issue/PR are pending human direction. No public review is claimed.
- **Related findings:** GO-SPEC-ISSUE-001 supplies strict JSON infrastructure but does not resolve vocabulary membership. The public `specification-issues` directory contains no matching vocabulary finding at the baseline below.

### Affected Baselines and Requirements

Public `barometer-spec` Draft `0.1.0`, immutable revision `59d49364d85e48cc8e69c762d130a7a86dc1fe65`:

- [REQ-OBS-0011](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/data-formats/measurement-observations.md#REQ-OBS-0011) requires a registered uppercase DNS response-code name.
- [REQ-OBS-0021](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/data-formats/measurement-observations.md#REQ-OBS-0021) requires a registered negotiated TLS version, registered ASCII ALPN value, and registered coarse alert class.
- [REQ-TLS-01-0016](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0016) describes registered offered ALPN identifiers in `options.alpn_protocols`.
- [REQ-OBS-0003](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/data-formats/measurement-observations.md#REQ-OBS-0003) permits a failure on incomplete observations only when registered; no separate incomplete-failure registration is identified.

The Go baseline before this task is `0f1a2020726f6dc0ec5a25fe90ee4c61177ba230`. No DNS/TLS observation implementation, release, binding or platform behavior is established. The partial model work is on `agent/dev-481-observation-context-models`, uncommitted pending review. No measurement is executed by the reproductions below.

### Finding and Synthetic Reproduction

The public observation/test/limit documents require registry membership but do not enumerate these vocabularies, identify a versioned registry source, or define which externally registered values are allowed. Examples containing `NOERROR`, `TLS1.3` or `h2` are examples, not complete registries. The DNS failure list names certain rcodes through outcome mapping, but does not define the complete required response-code vocabulary. A syntactically valid DNS response with another standard response code therefore has no specified membership decision. For TLS, a detail such as `{"alert":"handshake_failure"}` has no specified accepted coarse-class vocabulary. These are incomplete synthetic detail fragments, not valid complete observations.

Treating every standard DNS/TLS/IANA/Go value as accepted would introduce observable behavior absent from the reviewed public contract. Inferring a small allowlist from examples would reject values without a normative basis and would not implement all registered behavior. Copying a raw TLS alert or arbitrary ALPN/version string would defeat the specified normalized, bounded vocabulary.

**Blocked scope:** DNS/TLS observation validators, DNS response-code membership, TLS version/ALPN/alert membership, and any incomplete observation carrying a failure until applicable registration is explicit. Complete DEV-481 acceptance cannot be claimed. **Unaffected approved work:** TCP/HTTP closed-field models and Network Context normalization/binding. The partial protocol decoder currently accepts only TCP/HTTP and rejects DNS/TLS atomically; it forbids failures on incomplete observations. This is a visible implementation limitation, not a public clarification or fallback policy.

### Impact and Required Reviews

Public schema/measurement and safety/privacy maintainers must review exact retained vocabularies and unknown-value behavior. An arbitrary string or raw library error is not a safe substitute for a coarse TLS alert class. Overly narrow membership may discard ordinary protocol evidence; neither rejection nor a retained failure establishes censorship. Review must distinguish protocol failures from common execution termination.

No new destination, collection, persistence, linkage, native API or host data flow is proposed by this finding. No specialist review is claimed complete. Deployment legal/platform gates remain applicable; isolated synthetic model work does not satisfy them. Binding/host compatibility must be assessed once the accepted representation is known.

### Alternatives and Proposed Public Resolution

Public maintainers should choose and document one of:

1. Enumerate finite initial DNS response-code names, TLS version labels, ALPN identifiers and coarse TLS alert classes, with exact unknown/unregistered handling and version ownership.
2. Reference precise immutable external registries and explicitly define the supported subset, normalization, bounds, coarse alert mapping and unknown handling. External protocol registration alone must not imply safe observation retention.
3. Provide a no-text-change public disposition identifying existing authoritative definitions that resolve every membership decision.

Target wording belongs beside REQ-OBS-0011/0021, with a canonical registry link from the TLS options contract; initial incomplete-failure registration should be explicit or clarified as absent. Exact entries and wording remain pending public review. No alternative is selected by this local record. Public maintainers decide requirement-ID lifecycle, Draft version and compatibility; this Go task allocates no public IDs or wire fields.

Proposed validation after acceptance includes every registered value, values outside the registry, unknown/inconclusive cases, status/failure consistency, privacy-safe alert normalization, encoded-size boundaries and exact version selection. Expected membership outcomes remain pending, not fabricated conformance tests.

### Ordering and Public Disposition

No coordination decision is presently proposed because this finding concerns existing model vocabulary rather than architecture or ownership. If public resolution changes those boundaries, obtain that decision first. Prepare public clarification only under an approved cross-repository scope; merge accepted public wording and record its immutable revision/version decision, align required internal measurement/privacy validation records, then complete Go models/tests, bindings and host integration in that order. Release/migration order and blocked version combinations remain to be determined from the resolution.

### Adoption and Closure Evidence

Public record, maintainer disposition, accepted baseline, internal review alignment, DNS/TLS implementation commits, conformance outcomes and any downstream compatibility evidence are all pending. The local TCP/HTTP/context tests do not resolve this finding. Keep it open until an explicit public disposition and the applicable downstream evidence or non-applicability rationale are recorded.

### Proposal Preparation Addendum — October 15, 2026

The task owner approved expanding the scope to prepare a public clarification and dependent internal alignment. The proposed [Registered Observation Vocabularies](https://github.com/tv-rain/barometer-spec/blob/agent/dev-481-observation-vocabularies/specification-issues/observation-vocabularies.md) review package is prepared on `agent/dev-481-observation-vocabularies` from public revision `59d49364d85e48cc8e69c762d130a7a86dc1fe65`. Its initial commit/push is pending, so that branch link is an intended review location, not available immutable evidence. It proposes a fixed DNS subset with `OTHER`, bounded TLS version/ALPN values with `unknown`, coarse peer-alert classes, strict rejection of other serialized values and no failure field on incomplete observations. It proposes retaining Draft `0.1.0`, with explicit compatibility review, replacing the DNS/TLS table requirements and allocating ten new public obligations; these are not yet accepted decisions.

The internal repository has a dependent review/validation draft on the same branch name, based on its recorded internal baseline. It records required reviews and planned synthetic cases, not completed specialist review or accepted alignment. Public wording/version/ID approval, separate final commit/push approval, the human-created public PR and immutable public merge revision remain pending. Reconcile and merge required internal alignment before adopting the public contract in Go. No DNS/TLS implementation is added by this proposal preparation; the finding remains open and the original evidence above remains historical.

### Public Merge Verification — October 15, 2026

Public [PR #8](https://github.com/tv-rain/barometer-spec/pull/8) merged approved head `5869bc314c877e2ccd3426b82aadcd9c92414f02` into `main` at [a944ed6c389fec9f0d089f0044ead6fe845328b1](https://github.com/tv-rain/barometer-spec/commit/a944ed6c389fec9f0d089f0044ead6fe845328b1), retaining Draft `0.1.0`. Refreshed public `origin/main` contains both commits, and the merge tree matches the approved head exactly. The public vocabulary ambiguity is resolved by that accepted revision; GO-SPEC-ISSUE-002 remains open for downstream adoption/evidence.

The internal alignment branch pins that immutable merge revision and reconciles its validation plan and test companions to the unchanged approved wording. Internal merge is still pending and remains a prerequisite to DNS/TLS Go adoption. No DNS/TLS implementation, completed specialist review, binding migration or release support is established by this public merge. The historical finding and earlier pending statements are retained as dated evidence.

### Internal Merge and Go Model Adoption — October 15, 2026

Required internal alignment was verified merged into its remote main branch, with a merge tree identical to the approved alignment head. Private PR and immutable commit evidence are recorded in the task tracker. It pins the accepted public merge `a944ed6c389fec9f0d089f0044ead6fe845328b1`, retaining Draft `0.1.0`. Both repository prerequisites for Go model adoption are satisfied. This records plan alignment, not completed independent specialist review.

DEV-481 now prepares DNS/TLS models and the fixed producer mappings in `internal/model/observation` against that accepted public baseline. The [model documentation and test traceability](observation-models.md) identify the replacement/new requirements, strict serialized-value acceptance, exact absence-versus-coarse behavior, answer bounds and separate trusted answer-retention gate, local IP filter, certificate/peer-alert separation and incomplete-failure prohibition. The prior finding and pending statements remain historical records of the pre-resolution/adoption stages.

Go implementation is uncommitted pending final approval; the accepted finding remains open for durable downstream implementation evidence and applicable runtime/binding/producer validation or explicit non-applicability. Local model tests cannot establish packet/API evidence, measurement execution, native lifecycle, field support or a released library. No additional specification mutation is needed for this adoption.
