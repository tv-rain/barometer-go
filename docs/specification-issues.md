# Specification Issues

## Scope and Baseline

Current adoption summaries below were reconciled on October 16, 2026 against verified merged Foundation code. Dated findings and preparation-stage paragraphs are historical evidence, including their then-pending statements. The latest disposition and addenda control current adoption status; no historical source baseline is relabelled as current conformance.

This record holds Go implementation findings that need public specification clarification. It is not a local specification or an approval to choose behavior. Issue identifiers below belong to this Go findings record; they are not public requirement IDs, conformance-case IDs, or ClickUp task IDs.

Sources are pinned to public specification Draft `0.1.0`, revision `95dfeef1367f4c320bca7c7ae91b13ca7608b37f`, as recorded in the [Normative Source Inventory](normative-sources.md#reviewed-baseline). The source-discovery review found no missing contract file. This record is not an exhaustive semantic audit of all requirements. Findings from GO-SPEC-ISSUE-003 onward cite their own immutable public baselines.

## Recording and Reviewing Findings

Follow the public [workflow](https://github.com/tv-rain/barometer-spec/blob/main/specification-issues/README.md) and [template](https://github.com/tv-rain/barometer-spec/blob/main/specification-issues/template.md). Retain local finding IDs and historical source links. Cite allocated requirement IDs; for unallocated obligations cite exact immutable sections without inventing an ID. Stop affected implementation pending public disposition; unrelated approved work may continue.

Public maintainers resolve the contract. Merge changed public wording before required internal alignment and Go adoption, then bindings and host integration follow. Record the accepted immutable baseline, version/compatibility decision, implementation revisions, conformance cases/results, and release order before closing an accepted finding. No-text-change clarification requires public disposition citing the existing rule. Proposals and passing local tests are not acceptance.

The open finding below exercises the template without choosing decoding behavior. Merge the coordination decision and public procedure before this consumer update; its workflow links to public `main` become available after that merge. The procedure is development guidance outside the historical runtime baseline above.

## GO-SPEC-ISSUE-001: Duplicate JSON Member Names Outside Labels

**Status:** Public contract, required internal alignment and shared Go decoder adoption are merged. Foundation model consumers are delivered, with their scope recorded in the model documentation. Applicable production-schema, admission, runtime, producer and binding/host evidence remains pending under the [downstream adoption task](https://app.clickup.com/t/124kk1zc4ag). The finding remains open; historical pending statements below do not describe the current Foundation merge state.

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

### Foundation Merge Evidence — October 16, 2026

The accepted public contract remains Draft `0.1.0` at [05c1364af4f9a604b617fca45f060949b557d265](https://github.com/tv-rain/barometer-spec/commit/05c1364af4f9a604b617fca45f060949b557d265). Required internal review-plan alignment was verified merged before Go adoption; its private immutable evidence is retained in the [implementation task](https://app.clickup.com/t/86cbb8f0x), without copying private review material here or inferring completed specialist review.

Shared Go adoption [PR #14](https://github.com/tv-rain/barometer-go/pull/14) merged approved implementation `834e52a5a23e85c4f7e35de498432381979ed0e9` at [e267ac370716f93c850b8e539d17d37f06359941](https://github.com/tv-rain/barometer-go/commit/e267ac370716f93c850b8e539d17d37f06359941). The merge and head are ancestors of the refreshed integrated Foundation baseline. Subsequent configuration, task, observation, result and bundle consumers are recorded in their linked model documentation and [model-validation map](model-validation.md).

This resolves the stale “prepared/uncommitted” Foundation adoption state. It does not resolve every downstream obligation. [Downstream strict-JSON adoption](https://app.clickup.com/t/124kk1zc4ag) must inventory applicable production schemas, admission/no-side-effect integration, producer and binding/host consumers, and record durable evidence or reviewed non-applicability before this finding closes. Existing historical manifests and source inventories remain unchanged.

## GO-SPEC-ISSUE-002: Missing Registered Observation Vocabularies

**Current state:** Public resolution, required internal alignment and DEV-481 Go model adoption are merged and verified. Actual runtime producer, protocol/API, result-integration and applicable binding/host evidence remains pending under the [runtime vocabulary adoption task](https://app.clickup.com/t/124kk1zc4aj). The finding remains open. The original finding and dated preparation-stage statements below are historical evidence.

### Identity and Ownership

- **Disposition:** Open. The public resolution merged on October 15, 2026 at [a944ed6c389fec9f0d089f0044ead6fe845328b1](https://github.com/tv-rain/barometer-spec/commit/a944ed6c389fec9f0d089f0044ead6fe845328b1); runtime and applicable binding/host adoption remains pending as described under **Current state** above. Original disposition (October 15, 2026, historical): no public resolution or downstream adoption.
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

### Foundation Merge Evidence — October 16, 2026

The accepted public contract is [a944ed6c389fec9f0d089f0044ead6fe845328b1](https://github.com/tv-rain/barometer-spec/commit/a944ed6c389fec9f0d089f0044ead6fe845328b1), retaining Draft `0.1.0`. Required internal alignment was verified merged before Go adoption; private PR/commit evidence remains in the [observation model task](https://app.clickup.com/t/86cbb8f2n), without claiming completed specialist review.

Go [PR #22](https://github.com/tv-rain/barometer-go/pull/22) merged approved implementation `84ccd5c2d4003183d21ffe225a4479b6b572c53e` at [dea8351f618d48f916c1fd9f2576765460b84309](https://github.com/tv-rain/barometer-go/commit/dea8351f618d48f916c1fd9f2576765460b84309). The merge and head are ancestors of the refreshed integrated Foundation baseline. Bounded models, accepted fixed vocabulary mappings and their synthetic tests are delivered; the [observation model record](observation-models.md) remains canonical for that scope.

The [runtime vocabulary adoption task](https://app.clickup.com/t/124kk1zc4aj) owns actual DNS/TLS producer normalization, protocol/API and result integration, applicable bindings, compatibility and evidence or reviewed non-applicability. This finding stays open for that work; merged models and passing compiler checks do not establish protocol execution or platform support.

## GO-SPEC-ISSUE-003: Local Host-Configuration Decoding Bounds

**Current state:** Closed October 16, 2026. The public maintainer accepted the no-text-change clarification recorded in [Local Host-Configuration Decoding Bounds — Maintainer Disposition](https://github.com/tv-rain/barometer-spec/blob/main/specification-issues/configuration-bounds.md#maintainer-disposition--october-16-2026). The Go bounds are local and non-normative under REQ-JSON-0009. Go behavior is unchanged. The finding and ordering sections before the disposition addendum keep their original wording as historical evidence.

### Identity, Baselines and Requirements

- **Disposition:** Closed; see the [disposition addendum](#disposition-addendum--october-16-2026). Original disposition (historical): open, raised October 16, 2026 under [DEV-571](https://app.clickup.com/t/124kk1zc521). The Go implementation agent reports under the task owner's instruction; public maintainers own disposition.
- **Public baseline:** Draft `0.1.0` at [83d0b10cffb37a53de120b56825b42eb226568d9](https://github.com/tv-rain/barometer-spec/tree/83d0b10cffb37a53de120b56825b42eb226568d9). [REQ-JSON-0009](https://github.com/tv-rain/barometer-spec/blob/83d0b10cffb37a53de120b56825b42eb226568d9/data-formats/json-decoding.md#REQ-JSON-0009) requires a finite reviewed local input bound for models without a registered complete-input maximum. The Proxy, Network Context and Result Linkage sections register no complete host-configuration maximum or URL length maximum.
- **Implementation baseline:** `internal/config` at `374c8bd7a8db634ddac5f799b2894ce3d96856d5` selects 16,384 encoded bytes, 8 value levels, 64 value nodes, 2,048-byte endpoint URLs and 64-entry trusted approval lists, as recorded in [Initialization and Configuration Models](configuration-models.md#defaults-validation-and-local-bounds). The approval-list bound applies to trusted internal inputs, not a public serialized model.

### Finding and Interim Handling

Whether REQ-JSON-0009's "reviewed" local bound is satisfied by implementation-local review, or whether public maxima should be registered for host interoperability, needs a public disposition. The proposal recommends a no-text-change clarification that the Go bounds are explicitly local and non-normative. Registering public maxima remains the reviewed alternative.

No work is stopped. The bounds stay as implemented and fail atomically before parsing. They are not normative maxima, measured device limits or a guarantee to hosts. Existing synthetic tests in `internal/config` cover the byte bound at the maximum and maximum-plus-one; they are regression evidence only.

### Ordering and Closure

Merge the public record first, then this entry. A coordination decision is not required. If maintainers register public maxima, that wording change follows its own reviewed public PR, then any internal alignment, Go adoption and conformance cases, then bindings and hosts. Closure requires the recorded maintainer disposition and its immutable public revision; under the proposed clarification no Go behavior change follows.

### Disposition Addendum — October 16, 2026

The public record merged through barometer-spec [PR #10](https://github.com/tv-rain/barometer-spec/pull/10) at [d3ed5cbcc02cecac371bc897598b903c70d0fcfc](https://github.com/tv-rain/barometer-spec/commit/d3ed5cbcc02cecac371bc897598b903c70d0fcfc), and this entry merged through Go [PR #32](https://github.com/tv-rain/barometer-go/pull/32) at [1fa1e6edfa019e62e391191294e2fc81376a455c](https://github.com/tv-rain/barometer-go/commit/1fa1e6edfa019e62e391191294e2fc81376a455c). The project owner and public maintainer, Lev, then accepted alternative 1: a no-text-change clarification citing the existing [REQ-JSON-0009](https://github.com/tv-rain/barometer-spec/blob/83d0b10cffb37a53de120b56825b42eb226568d9/data-formats/json-decoding.md#REQ-JSON-0009) rule. The disposition is recorded in the public record's maintainer-disposition section; its merged public revision is recorded as evidence on [DEV-571](https://app.clickup.com/t/124kk1zc521).

- Draft `0.1.0` text, requirement IDs and versions are unchanged. No internal alignment, Go behavior change, conformance case, binding or host change is required.
- The bounds remain the documented local envelope in [Initialization and Configuration Models](configuration-models.md#defaults-validation-and-local-bounds). They are not public maxima, measured device limits or a guarantee to hosts.
- Existing `internal/config` bound tests remain synthetic regression evidence only.
- Registering public maxima would need a new specification issue; none is open.

## GO-SPEC-ISSUE-004: Boundary Between `not_run` and `incomplete`

**Current state:** Open, raised October 16, 2026 under [DEV-573](https://app.clickup.com/t/124kk1zc613) from the [DEV-572](https://app.clickup.com/t/124kk1zc5xp) case-plan triage. No public record exists yet. Public maintainers own disposition.

### Identity, Baselines and Requirements

- **Reporter:** the Go implementation agent, under the task owner's instruction. **Duplicates:** none. GO-SPEC-ISSUE-001 and GO-SPEC-ISSUE-002 do not cover execution states.
- **Shared baseline:** Draft `0.1.0` at [a944ed6c389fec9f0d089f0044ead6fe845328b1](https://github.com/tv-rain/barometer-spec/tree/a944ed6c389fec9f0d089f0044ead6fe845328b1). [REQ-RESULT-0019](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0019) defines `completed`, `incomplete` and `not_run`, with a required reason for the last two. [REQ-RESULT-0021](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0021) gives the state rules. [REQ-RESULT-0023](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0023) and [REQ-RESULT-0024](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0024) make Network Context required for `completed` and `incomplete` remote tests and omissible for `not_run` before a snapshot. [REQ-EXECUTION-0015](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0015) says only that a task "may not run" when network conditions cannot be detected.
- **Test baseline:** registered tests at [59d49364d85e48cc8e69c762d130a7a86dc1fe65](https://github.com/tv-rain/barometer-spec/tree/59d49364d85e48cc8e69c762d130a7a86dc1fe65). Each termination rule maps outer interruptions to an incomplete state: REQ-NET-01-0027, REQ-NET-03-0029, REQ-DNS-01-0039, REQ-DNS-03-0028, REQ-DNS-04-0028, REQ-TCP-01-0026, REQ-TLS-01-0027 and REQ-HTTP-01-0038. The NET-05 outcome rules are [REQ-NET-05-0046](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0046) and [REQ-NET-05-0047](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0047).
- **Implementation:** `tv-rain/barometer-go` at `58e03a1f65600371960f735595f0a3c84550611f`. `internal/execution` and `internal/result` model the three states and validate state/reason/measurement combinations. No runner selects a state for a real interruption yet.

### Finding and Interim Handling

No public rule says when an attempt stops being `not_run` and becomes `incomplete`. Four cases are open:

1. An interruption after admission and context capture but before any test traffic.
2. An interruption after a test's terminal observation is recorded but before its connection is closed.
3. The execution state that accompanies the NET-05 composite outcomes `network_not_validated` and `inconclusive`.
4. The state and reason when a required network condition cannot be detected (REQ-EXECUTION-0015), where neither `constraints_unsatisfied` nor `network_context_unavailable` is clearly specified.

The choice matters because it decides whether Network Context must be retained under REQ-RESULT-0023/0024, so it affects data minimization as well as interpretation.

**Blocked:** selecting the state and reason for these four cases in runner, execution and result code ([DEV-561](https://app.clickup.com/t/124kk1zc4a9)), and the corresponding expected values in the `M-*-TERMINATION` case plans. **Interim handling:** existing models keep validating only the documented combinations. Case plans keep deferring to the public table. No fallback state is chosen. **Unaffected:** interruptions during an active protocol operation, which every termination rule already maps to `incomplete` with no `measurement`.

### Impact, Reviews and Alternatives

- **Safety and privacy:** affects whether a context snapshot is retained for an attempt that sent no traffic. Privacy review is required for any rule that retains more context.
- **Interpretation:** analysts must be able to separate "never started" from "interrupted". Neither state is evidence of interference.
- **Alternatives:** (a) `not_run` until the first test-specific network operation starts, then `incomplete`; (b) `not_run` only before admission completes; (c) per-test definitions. For cases 2 and 3, choose whether a recorded terminal observation makes the attempt `completed`. For case 4, name one reason.
- **Compatibility:** a no-text-change clarification is possible only if maintainers find an existing rule that settles each case. Otherwise the change is a Draft `0.1.0` wording change with requirement-ID review.

### Ordering and Closure

A coordination decision is not expected. Prepare a public record first. Once maintainers accept wording, merge it, align internal reviews, then adopt it in Go execution/result code and termination cases. Bindings and hosts follow. Closure requires the public disposition, the merged revision if text changes, and Go adoption evidence.

## GO-SPEC-ISSUE-005: Which Bytes Count Toward Attempt Byte Limits

**Current state:** Open, raised October 16, 2026 under [DEV-573](https://app.clickup.com/t/124kk1zc613) from the DEV-572 triage. No public record exists yet.

### Identity, Baselines and Requirements

- **Reporter:** the Go implementation agent, under the task owner's instruction. **Duplicates:** none.
- **Baseline:** Draft `0.1.0` at [a944ed6c389fec9f0d089f0044ead6fe845328b1](https://github.com/tv-rain/barometer-spec/tree/a944ed6c389fec9f0d089f0044ead6fe845328b1). [REQ-LIMIT-0006](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0006) counts `max_bytes_sent` as "application bytes supplied to network APIs, excluding lower-layer framing" and `max_bytes_received` as "application bytes accepted from network APIs before parsing stops". [REQ-LIMIT-0003](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0003) says byte limits are not exact on-wire limits because operating-system, DNS, TCP, TLS, HTTP and proxy layers may add overhead. [REQ-LIMIT-0016](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0016) and [REQ-LIMIT-0019](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0019) add separate header, prefix and composite byte bounds.
- **Implementation:** no byte-budget enforcement exists at `58e03a1f65600371960f735595f0a3c84550611f`.

### Finding and Interim Handling

REQ-LIMIT-0003 lists HTTP and DNS among the layers whose overhead is outside the counted data. But for HTTP-01, NET-03 and DNS-03 the HTTP request line and headers are what the test supplies to an HTTP API, and for DNS tests the DNS message is the payload. The text does not say whether these count as application bytes:

- HTTP request lines and headers, and HTTP response headers, already bounded separately at 16,384 bytes;
- DoH HTTP framing and DoT two-byte length prefixes;
- the proxy CONNECT exchange, which REQ-LIMIT-0003 names as proxy overhead, while `max_connections` counts CONNECT tunnels.

Two conforming implementations could therefore stop the same task at different points.

**Blocked:** byte-budget accounting in [DEV-562](https://app.clickup.com/t/124kk1zc4ab) for HTTP-01, NET-03, DNS-03, DNS-04 and NET-05 aggregation, and the expected boundary values in the `CC-LIMIT-PURPOSE` case plan. **Interim handling:** no accounting layer is chosen; plans keep the open question visible in the `RV-CC-LIMIT-PURPOSE` review case. **Unaffected:** connection, duration, body-prefix and header-size bounds, which have their own units.

### Impact, Reviews and Alternatives

- **Safety:** under-counting lets a task send or receive more than a reviewer approved. Over-counting can make small registered targets fail. Resource and traffic-distinctiveness review is required.
- **Alternatives:** (a) count everything the test hands to the TLS/TCP or system HTTP API, including HTTP headers and DNS messages, and exclude only TLS records, TCP/IP framing and the CONNECT exchange; (b) count only the HTTP message body and DNS message payload; (c) per-test accounting tables.
- **Compatibility:** registered numbers stay unchanged under (a) and (b). Either changes which tasks reach the limit, so record a version decision.

### Ordering and Closure

Public record and disposition first, then internal alignment, then Go accounting with scripted-transport cases, then bindings. Closure requires the public disposition and Go adoption evidence.

## GO-SPEC-ISSUE-006: NET-03 Over-Bound Response Classification

**Current state:** Open, raised October 16, 2026 under [DEV-573](https://app.clickup.com/t/124kk1zc613) from the DEV-572 triage. No public record exists yet.

### Identity, Baselines and Requirements

- **Reporter:** the Go implementation agent, under the task owner's instruction. **Duplicates:** none.
- **Baseline:** NET-03 at [59d49364d85e48cc8e69c762d130a7a86dc1fe65](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md). The data row of [REQ-NET-03-0021](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0021) says that on reaching the header or prefix bound the test must "stop reading and classify the response as exceeding the registered validation bound". [REQ-NET-03-0025](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0025) allows only `expected`, `redirect`, `content_mismatch` and `unexpected`. [REQ-NET-03-0027](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0027) has no matching outcome. [REQ-NET-03-0029](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0029) makes a resource limit an incomplete state with no `measurement`.
- **Implementation:** no NET-03 runner exists at `58e03a1f65600371960f735595f0a3c84550611f`.

### Finding and Interim Handling

The over-bound path has three possible readings: an existing value (`unexpected` or `inconclusive`), a missing new value, or an incomplete execution under REQ-NET-03-0029. The readings differ in whether a measurement exists at all.

**Blocked:** the NET-03 header/prefix bound path, and the matching rows of the `M-NET-03-LIMITS` and `M-NET-03-OUTPUT` plans. **Interim handling:** no value is chosen and no response is classified. **Unaffected:** all other NET-03 paths.

### Impact, Reviews and Alternatives

- **Interpretation:** an oversized validation response could be a captive portal or an ordinary large page. Neither reading is evidence of interference. Measurement review is required.
- **Alternatives:** (a) map it to `response_class: unexpected` and outcome `unexpected_response`; (b) add a registered value; (c) treat it as `resource_limit` under REQ-NET-03-0029 and drop the classification wording.
- **Compatibility:** (b) adds a value to Draft output and needs a version decision. (a) and (c) are wording clarifications.

### Ordering and Closure

Public record and disposition first, then Go adoption when the NET-03 runner is implemented. Closure requires the public disposition and Go adoption evidence.

## GO-SPEC-ISSUE-007: Undefined Allowed Values for `risk_factors` and `timezone`

**Current state:** Open; public disposition merged at [5a37e82865a08922c4821c7318656286933421e5](https://github.com/tv-rain/barometer-spec/tree/5a37e82865a08922c4821c7318656286933421e5) and adopted in Go decoding and validation by [DEV-582](https://app.clickup.com/t/124kk1zcaah). Go now rejects `risk_factors: []` for `medium`, `high` and `unknown` risk; `timezone` validation was already conforming and is now evidenced. Admission-path, binding and host evidence remain open; see the [adoption addendum](#adoption-addendum--october-19-2026). The sections before the addenda keep their original wording as historical evidence. Originally raised October 16, 2026 under [DEV-573](https://app.clickup.com/t/124kk1zc613) from the DEV-572 triage. This finding affects merged Foundation code.

### Identity, Baselines and Requirements

- **Reporter:** the Go implementation agent, under the task owner's instruction. **Duplicates:** none. Strict JSON decoding ([GO-SPEC-ISSUE-001](#go-spec-issue-001-duplicate-json-member-names-outside-labels)) fixes member handling, not these value domains.
- **Baseline:** Draft `0.1.0` at [a944ed6c389fec9f0d089f0044ead6fe845328b1](https://github.com/tv-rain/barometer-spec/tree/a944ed6c389fec9f0d089f0044ead6fe845328b1). [REQ-SAFETY-0012](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0012) requires `risk_factors` for `medium`, `high` and `unknown` risk, and [REQ-SAFETY-0013](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0013) defines the factor fields; neither sets a minimum count. [REQ-SCHEDULE-0013](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0013) describes `timezone` as the timezone used to interpret ranges, "`local` is default", and [REQ-SCHEDULE-0014](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0014) defines only `local`.
- **Implementation:** `tv-rain/barometer-go` at `58e03a1f65600371960f735595f0a3c84550611f`:
  - `internal/taskdecode/schema.go` decodes `risk_factors` as an array with no minimum length. `internal/taskdecode/decode.go` builds an empty, non-nil factor list for `[]`. `internal/validation/task.go` (`validSafety`) rejects only a missing list, so a `medium`, `high` or `unknown` task with `risk_factors: []` is accepted.
  - `internal/validation/task.go` accepts only `timezone: "local"` and rejects any other value. Its comment records this as the only documented value.
- **Synthetic reproduction:** on October 16, 2026, a temporary local test using the existing `taskdecode` fixture decoder (not committed) decoded synthetic tasks with `risk_level` `medium`, `high` and `unknown` and `risk_factors: []` without error, and rejected the same tasks with `risk_factors` absent.

### Finding and Interim Handling

Two value domains are unspecified:

- whether an empty `risk_factors` array satisfies "required";
- which `timezone` values besides `local` exist, and whether others must be rejected.

The current Go choices were made without a public disposition. The `timezone` choice is the conservative one. The `risk_factors` choice is not: it lets a higher-risk task arrive with no machine-readable explanation of its risk.

**Blocked:** further adoption that depends on either domain, and any claim that these validators conform. **Interim handling:** this task does not change Go behavior. Maintainers should decide first. The finding records the more conservative option of rejecting `[]` for `medium`, `high` and `unknown`. Changing merged validation needs its own approved task. **Unaffected:** other safety and schedule validation.

### Impact, Reviews and Alternatives

- **Safety:** `risk_factors` explains effective risk to hosts and reviewers. Safety review is required. Accepting `[]` does not lower effective risk, which is still bounded by the registered baseline, but it weakens review.
- **Alternatives for `risk_factors`:** (a) require at least one factor, the recommended conservative reading; (b) allow `[]` explicitly.
- **Alternatives for `timezone`:** (a) close the domain to `local` in Draft `0.1.0`; (b) define an IANA timezone domain with bounds.
- **Compatibility:** (a) for both matches or tightens current Go behavior. (b) for either is a schema extension that needs a version decision.

### Ordering and Closure

Public record and disposition first. If `risk_factors` must contain at least one factor, a separate Go task changes `validSafety` and adds negative tests. Bindings and hosts follow. Closure requires the public disposition and, where applicable, Go adoption evidence.

### Disposition Addendum — October 19, 2026

The public record [`risk_factors` and `timezone` Value Domains](https://github.com/tv-rain/barometer-spec/blob/main/specification-issues/risk-factors-timezone-domains.md) was prepared under [DEV-581](https://app.clickup.com/t/124kk1zcaag) against public Draft `0.1.0` at [9cafe8d1dc48fb40fb6e49b58a8eefe41a9ee306](https://github.com/tv-rain/barometer-spec/tree/9cafe8d1dc48fb40fb6e49b58a8eefe41a9ee306), where the cited safety and schedule text is unchanged from this finding's baseline. The project owner and public maintainer, Lev, accepted alternative (a) for both domains on October 19, 2026. The accepted wording, REQ-SAFETY-0031 and REQ-SCHEDULE-0043, is part of the same public change and retains Draft `0.1.0`. Its merged immutable revision is recorded as evidence on DEV-581.

- **`risk_factors`:** for `risk_level` `medium`, `high` and `unknown`, `risk_factors` must contain at least one factor; `[]` is rejected. `validSafety` in `internal/validation/task.go`, which rejects only a missing list, does not conform.
- **`timezone`:** `local` is the only defined value in Draft `0.1.0`; any other value is rejected. Current validation, which accepts only `local` and defaults an omitted value to `local`, conforms.
- **Public wording:** new obligations [REQ-SAFETY-0031](https://github.com/tv-rain/barometer-spec/blob/main/data-formats/task-safety.md#REQ-SAFETY-0031) (at least one factor for `medium`, `high` and `unknown`) and [REQ-SCHEDULE-0043](https://github.com/tv-rain/barometer-spec/blob/main/data-formats/task-schedule.md#REQ-SCHEDULE-0043) (`local` is the only `timezone` value). REQ-SAFETY-0012, REQ-SCHEDULE-0013 and REQ-SCHEDULE-0014 are unchanged. Go adoption must cite the merged immutable revision, not a branch.
- **Internal alignment:** not required; no internal review or validation plan covers these fields.
- **Go adoption:** [DEV-582](https://app.clickup.com/t/124kk1zcaah) changes `risk_factors` validation, adds negative and positive cases for both domains, and updates traceability, including manifest entries for the two new requirement IDs. Go behavior is unchanged by this addendum.

This finding stays open until DEV-582 records adoption evidence for `risk_factors` and evidence or non-applicability for `timezone`.

### Adoption Addendum — October 19, 2026

[DEV-582](https://app.clickup.com/t/124kk1zcaah) adopts the disposition against public Draft `0.1.0` at [5a37e82865a08922c4821c7318656286933421e5](https://github.com/tv-rain/barometer-spec/tree/5a37e82865a08922c4821c7318656286933421e5), the merge of public PR #13 recorded on DEV-581.

- **`risk_factors`:** `validSafety` in `internal/validation/task.go` now rejects a missing or empty factor list for `medium`, `high` and `unknown` risk ([REQ-SAFETY-0031](https://github.com/tv-rain/barometer-spec/blob/5a37e82865a08922c4821c7318656286933421e5/data-formats/task-safety.md#REQ-SAFETY-0031)). The whole task is rejected with `validation.ErrTask` and no candidate. A `low` task may still omit the list or send it empty. The decoder schema is unchanged, because an empty array remains valid for `low`.
- **`timezone`:** no behavior change. Decoding defaults an omitted value to `local`, and validation rejects every other value ([REQ-SCHEDULE-0043](https://github.com/tv-rain/barometer-spec/blob/5a37e82865a08922c4821c7318656286933421e5/data-formats/task-schedule.md#REQ-SCHEDULE-0043)). The comment that called `local` the "initially implemented" value now cites the requirement.
- **Implementation:** Go commit `4575079b503467bff83c1175329be6cfa64cfb71`. Its new tests fail at the previous `validSafety` on exactly the empty-list cases and pass after the change.
- **Cases:** synthetic positive and negative cases at three levels. `TestCommonRiskFactorAndTimezoneDomains` covers `validation.Common`, `TestRiskFactorAndTimezoneDomains` covers `taskdecode` decoding, and 13 `risk-factors-*` and `timezone-*` fixtures run through the [conformance fixture runner](conformance-fixtures.md), whose requirement allowlist now includes these IDs. Rejected `timezone` values include `UTC`, `Europe/Berlin`, `""`, `Local` and `LOCAL`.
- **Traceability:** the [cross-cutting manifest](../traceability/cross-cutting.json) is rebaselined to public `5a37e82` and Go `4575079` and links this finding as `accepted_adoption_pending`. REQ-SAFETY-0012, REQ-SAFETY-0013, REQ-SAFETY-0031, REQ-SCHEDULE-0013, REQ-SCHEDULE-0014 and REQ-SCHEDULE-0043 are `partial`, with [executed unit-test evidence](../traceability/evidence/go-unit-4575079.md). None is `implemented`.
- **Remaining scope:** admission under local policy, and the planned `CC-SAFETY-RISK-FACTORS` and `CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-FIELDS` family cases, which need the pending admission and runtime work. Binding and host adoption is unassessed until those components exist. External producers and stored tasks are unassessed. No specialist safety review is claimed.

The finding stays open until the remaining scope records evidence or explicit non-applicability.

## GO-SPEC-ISSUE-008: Admission Error Classes for Unsupported and Unmapped Rejections

**Current state:** Open; public disposition merged at [d93aff93cbc17e7c065fb40d8f6e2e01a697bf35](https://github.com/tv-rain/barometer-spec/tree/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35) and adopted in the Go admission design, case plans, traceability and `task.AdmissionError` by [DEV-597](https://app.clickup.com/t/124kk1zcb7t). No admission implementation exists yet, so the admission tasks (DEV-586 to DEV-590) still have to implement and evidence every table row, precedence and embedded replay handling; internal alignment is tracked separately. See the [disposition addendum](#disposition-addendum--october-19-2026-1) and the [adoption addendum](#adoption-addendum--october-19-2026-1). The sections before the addenda keep their original wording as historical evidence. Originally raised October 19, 2026 under [DEV-583](https://app.clickup.com/t/124kk1zcadk) while recording the [admission design](admission.md).

### Identity, Baselines and Requirements

- **Reporter:** the Go implementation agent, under the task owner's instruction. **Duplicates:** none. [GO-SPEC-ISSUE-004](#go-spec-issue-004-boundary-between-not_run-and-incomplete) concerns execution states, not admission classes.
- **Baseline:** Draft `0.1.0` at [5a37e82865a08922c4821c7318656286933421e5](https://github.com/tv-rain/barometer-spec/tree/5a37e82865a08922c4821c7318656286933421e5). [REQ-TRUST-0034](https://github.com/tv-rain/barometer-spec/blob/5a37e82865a08922c4821c7318656286933421e5/library/task-trust.md#REQ-TRUST-0034) lists ten "initial" admission classes, and [REQ-TRUST-0017](https://github.com/tv-rain/barometer-spec/blob/5a37e82865a08922c4821c7318656286933421e5/library/task-trust.md#REQ-TRUST-0017) orders the admission steps without assigning classes to them. [REQ-SAFETY-0017](https://github.com/tv-rain/barometer-spec/blob/5a37e82865a08922c4821c7318656286933421e5/data-formats/task-safety.md#REQ-SAFETY-0017) permits rejecting an unknown risk factor type "as unsupported". [REQ-EXECUTION-0021](https://github.com/tv-rain/barometer-spec/blob/5a37e82865a08922c4821c7318656286933421e5/data-formats/task-execution.md#REQ-EXECUTION-0021), [REQ-EXECUTION-0023](https://github.com/tv-rain/barometer-spec/blob/5a37e82865a08922c4821c7318656286933421e5/data-formats/task-execution.md#REQ-EXECUTION-0023), [REQ-CONTEXT-0026](https://github.com/tv-rain/barometer-spec/blob/5a37e82865a08922c4821c7318656286933421e5/library/network-context.md#REQ-CONTEXT-0026) and [REQ-CONTEXT-0028](https://github.com/tv-rain/barometer-spec/blob/5a37e82865a08922c4821c7318656286933421e5/library/network-context.md#REQ-CONTEXT-0028) treat an unavailable required egress field as an unsatisfied execution constraint recorded per attempt as `network_context_unavailable`.
- **Implementation:** no admission package exists at `dea88adaa87e1191a745643478f8e0c02eddad3c`. `task.AdmissionError` holds the ten codes; `validation.Common` rejects unknown risk factor types with `validation.ErrTask`, which has no admission class yet.

### Finding and Interim Handling

No public class covers a task that is valid under the public schema but uses a feature this implementation does not provide. The [Admission and local policy sub-epic](https://app.clickup.com/t/124kk1zca7y) plans to reject tasks that require egress Network Context until egress acquisition exists ([DEV-590](https://app.clickup.com/t/124kk1zcae4)). The public text instead admits such a task and lets each attempt end with `network_context_unavailable`. REQ-SAFETY-0017's "unsupported" rejection likewise has no class.

The public text also leaves these admission mappings open; the [admission design](admission.md#ordered-steps-and-error-classes) marks each as *open*:

- a missing, expired or mismatched delivery record, and unavailable trusted time: `task_source_unauthenticated` or `task_trust_unavailable`;
- remote `issued_at` more than 300 seconds in the future;
- `task_id` over 128 bytes: `task_schema_invalid` or `task_limit_exceeded`;
- a well-formed but unregistered test or version;
- a source policy generation higher than the active one;
- a path or proxy the issuer or local policy does not permit;
- effective risk above a local-only bound, or `unknown` without a reviewed exception;
- precedence when several steps would fail. The design reports the first failing step in REQ-TRUST-0017 order.

The design also applies the fail-closed replay rule, which the public table states for remote tasks, to embedded tasks.

**Blocked:** DEV-590, and the *open* rows of DEV-586, DEV-587 and DEV-589 and of the `CC-TRUST-ADMISSION-PROCEDURE`, `CC-TRUST-ADMISSION-ERRORS`, `CC-SAFETY-RISK-FACTORS-INITIAL-RISK-FACTOR-TYPES` and `CC-LIMIT-TASK-AND-IDENTIFIER-LIMITS` plans. **Interim handling:** no class is chosen and no Go behavior changes; plans assert only that rejection happens, that the class is one of the ten public classes and that no network activity occurs. **Unaffected:** mappings the public names determine, such as `task_expired`, `task_id_conflict`, `policy_rollback` and `target_not_registered`.

### Impact, Reviews and Alternatives

- **Safety and interoperability:** every option rejects or skips before network activity, so the safety difference is small. Hosts and issuers that branch on admission classes would diverge between implementations. Admitting an unsupported task can also produce a stream of `not_run` results that look like network failures. Safety review of the chosen mapping is required.
- **Alternatives for unsupported features:** (a) reuse `task_schema_invalid`; (b) add a class such as `task_unsupported`; (c) admit and let execution record `network_context_unavailable`, with no admission rejection; this already conforms for egress but not for REQ-SAFETY-0017.
- **Alternatives for the other mappings:** (a) a normative step-to-class table in `library/task-trust.md`; (b) state that classes are advisory and only rejection is normative.
- **Compatibility:** a new class extends the operational vocabulary and needs a version decision. A mapping table without new classes is a clarification.

### Ordering and Closure

Public record and disposition first, then internal alignment if the task-trust review covers error classes, then Go adoption in the admission tasks, then bindings. Closure requires the public disposition and Go adoption evidence.

### Disposition Addendum — October 19, 2026

The public record [Admission Error Classes for Unsupported and Unmapped Rejections](https://github.com/tv-rain/barometer-spec/blob/main/specification-issues/admission-error-classes.md) was prepared under [DEV-596](https://app.clickup.com/t/124kk1zcb6g) against this finding's public baseline, Draft `0.1.0` at [5a37e82865a08922c4821c7318656286933421e5](https://github.com/tv-rain/barometer-spec/tree/5a37e82865a08922c4821c7318656286933421e5). The project owner and public maintainer, Lev, accepted alternative (b) for unsupported features and alternative (a) for the other mappings on October 19, 2026. The accepted wording, REQ-TRUST-0036 to REQ-TRUST-0038, is part of the same public change and retains Draft `0.1.0`. Its merged immutable revision is recorded as evidence on DEV-596.

- **Unsupported features:** [REQ-TRUST-0036](https://github.com/tv-rain/barometer-spec/blob/main/library/task-trust.md#REQ-TRUST-0036) adds `task_unsupported` for a valid task that requires an unregistered test or version, an unrecognized risk factor type (when rejected under REQ-SAFETY-0017) or a capability the implementation does not provide, such as egress Network Context for `execution.context`. The DEV-590 rejection of egress-requiring tasks conforms with this class. `task.AdmissionError` lacks the code and must gain it before admission returns it.
- **Mappings and precedence:** [REQ-TRUST-0037](https://github.com/tv-rain/barometer-spec/blob/main/library/task-trust.md#REQ-TRUST-0037) makes the first failing step in REQ-TRUST-0017 order normative and maps each *open* row of the [admission design](admission.md#ordered-steps-and-error-classes): unbound delivery record and future `issued_at` to `task_source_unauthenticated`; unavailable trusted time to `task_trust_unavailable`; `task_id` over 128 bytes to `task_limit_exceeded`; unregistered test or version, rejected risk factor type and missing capability to `task_unsupported`; higher source policy generation to `task_policy_unavailable`; path or proxy not permitted by issuer authority or local policy, and risk above a local bound or unpermitted `unknown`, to `task_issuer_not_authorized`; path not permitted by the matched target registration to `target_not_registered`.
- **Embedded replay:** [REQ-TRUST-0038](https://github.com/tv-rain/barometer-spec/blob/main/library/task-trust.md#REQ-TRUST-0038) confirms the design's fail-closed replay handling for embedded tasks.
- **Public wording:** REQ-TRUST-0017, REQ-TRUST-0024, REQ-TRUST-0033 and REQ-TRUST-0034 are unchanged. Go adoption must cite the merged immutable revision, not a branch.
- **Internal alignment:** required. The internal task-trust review validation plan names admission classes and limits fail-closed replay to remote tasks; it must cite the merged wording.
- **Go adoption:** after the public merge, a Go change rebaselines the admission design and the `CC-TRUST-ADMISSION-PROCEDURE`, `CC-TRUST-ADMISSION-ERRORS`, `CC-SAFETY-RISK-FACTORS-INITIAL-RISK-FACTOR-TYPES` and `CC-LIMIT-TASK-AND-IDENTIFIER-LIMITS` plans to exact classes, adds `task_unsupported` to `task.AdmissionError` and adds traceability entries for the three new requirement IDs. DEV-586, DEV-587, DEV-589 and DEV-590 then implement and test the rows. Go behavior is unchanged by this addendum.

This finding stays open until the admission tasks record adoption evidence for every table row, precedence and embedded replay handling.

### Adoption Addendum — October 19, 2026

[DEV-597](https://app.clickup.com/t/124kk1zcb7t) adopts the disposition against public Draft `0.1.0` at [d93aff93cbc17e7c065fb40d8f6e2e01a697bf35](https://github.com/tv-rain/barometer-spec/tree/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35), the merge of public PR #14 recorded on DEV-596.

- **Error vocabulary:** `task.AdmissionCode` gains `Unsupported`, formatted as `task_unsupported` ([REQ-TRUST-0036](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0036)), alongside the ten REQ-TRUST-0034 classes. It carries only the code, like the others. Go commit `c49a77fc35d0190992c272ba7447467ab8ff675c`; `TestAdmissionErrorVocabulary` covers the new code and rejects 0 and 12 to 255.
- **Admission design:** [`docs/admission.md`](admission.md#ordered-steps-and-error-classes) is rebaselined to `d93aff9`. Every *open* row now names its [REQ-TRUST-0037](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0037) class: unbound delivery record and future remote `issued_at`, `task_source_unauthenticated`; unavailable trusted time, `task_trust_unavailable`; unregistered test or version, rejected risk factor type and missing capability, `task_unsupported`; `task_id` over 128 bytes, `task_limit_exceeded`; higher source policy generation, `task_policy_unavailable`; path not permitted by issuer authority or local policy and risk above a local bound or unpermitted `unknown`, `task_issuer_not_authorized`; path not permitted by the matched target registration, `target_not_registered`. The first failing step is normative, and embedded replay fails closed with `task_trust_unavailable` ([REQ-TRUST-0038](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0038)).
- **Case plans:** `CC-TRUST-ADMISSION-PROCEDURE`, `CC-TRUST-ADMISSION-ERRORS`, `CC-SAFETY-RISK-FACTORS-INITIAL-RISK-FACTOR-TYPES`, `CC-LIMIT-TASK-AND-IDENTIFIER-LIMITS`, `CC-TRUST-OVERVIEW`, `CC-TRUST-ACCEPTED-TASK-ORIGINS`, `CC-TRUST-TRUSTED-SOURCE-CONTEXT` and `CC-TRUST-ISSUER-AUTHORIZATION` assert exact classes, and the procedure plan adds precedence cases. `CC-TRUST-REPLAY-DUPLICATE-AND-ROLLBACK-HANDLING` adds the embedded fail-closed row.
- **Traceability:** the [cross-cutting manifest](../traceability/cross-cutting.json) is rebaselined to public `d93aff9` and Go `c49a77f` and links this finding as `accepted_adoption_pending`. REQ-TRUST-0036 is `partial`, with [executed unit-test evidence](../traceability/evidence/go-unit-c49a77f.md) for the code only. REQ-TRUST-0037 and REQ-TRUST-0038 are `planned`. None is `implemented`.
- **Remaining scope:** DEV-586, DEV-587, DEV-589 and DEV-590 implement and test the rows, and the replay task (DEV-591) implements embedded fail-closed handling. The internal task-trust review alignment is separate. Binding and host adoption is unassessed until those components exist. No specialist safety review is claimed.

The finding stays open until the admission tasks record adoption evidence for every table row, precedence and embedded replay handling.

### Gate Evidence Addendum — October 19, 2026

[DEV-586](https://app.clickup.com/t/124kk1zcadx) implements the gates for admission steps 1 to 6 at Go `b25f107f393f531adb4a54d89b76364db4dda640` ([Implemented Gates](admission.md#implemented-gates)). Synthetic unit tests, recorded in the [evidence report](../traceability/evidence/go-unit-b25f107.md), exercise these REQ-TRUST-0037 rows: unbound delivery record and remote `issued_at` more than 300 seconds ahead, `task_source_unauthenticated`; unavailable trusted time, `task_trust_unavailable`; unregistered test or version, `task_unsupported`; `task_id` over 128 bytes, `task_limit_exceeded`; and, at step 6, a rejected risk factor type, `task_unsupported`. They also exercise first-failing-step precedence among steps 1 to 6 and the input and options part of step 8. The rows for a missing capability (DEV-590), steps 7 to 9 (DEV-587 to DEV-589) and embedded replay (DEV-591) have no runtime evidence yet. The finding stays open.

### Issuer Authorization Evidence Addendum — October 19, 2026

[DEV-587](https://app.clickup.com/t/124kk1zcady) adds issuer and policy resolution at step 7, path authority at step 8 and the issuer and local risk bounds at step 9 at Go `aab88f59cbcf62208c9ebfa025fa795354dd072d` ([Issuer Authorization](admission.md#issuer-authorization)). Synthetic unit tests, recorded in the [evidence report](../traceability/evidence/go-unit-aab88f5.md), exercise these REQ-TRUST-0037 rows: source `policy_generation` greater than the active generation, `task_policy_unavailable`; a direct or proxy path not permitted by issuer authority or local policy, `task_issuer_not_authorized`; and effective risk above an issuer or local bound, or `unknown` without the reviewed exception, `task_issuer_not_authorized`, evaluated on the registered baseline and declared risk only. They also exercise precedence around steps 7 and 8. The target-registration rows of step 8 (DEV-588), locally determined risk (DEV-589), the missing-capability row (DEV-590) and embedded replay (DEV-591) have no runtime evidence yet. The finding stays open.

### Target Registration Evidence Addendum — October 19, 2026

[DEV-588](https://app.clickup.com/t/124kk1zcae0) adds target registration matching at step 8 at Go `cb672923cdd7668324f6ac4802d6c60edca5572c` ([Target Registration](admission.md#target-registration)). Synthetic unit tests exercise the REQ-TRUST-0037 row for a target or path not permitted by the matched registration, `target_not_registered`, and its precedence after steps 4 and 6 and the step-8 path authority and before step 9. They are recorded in the [evidence report](../traceability/evidence/go-unit-cb67292.md). The public text assigns no class to parameters that are incompatible with the proxy path; registry proxy compatibility is therefore not implemented, and composite steps need the NET-05 codec. Locally determined risk (DEV-589), the missing-capability row (DEV-590) and embedded replay (DEV-591) have no runtime evidence yet. The finding stays open.

### Effective Risk Evidence Addendum — October 20, 2026

[DEV-589](https://app.clickup.com/t/124kk1zcae1) completes step 9 at Go `41c6f06b0ef46b6324f9f795299ac70af05c6086` ([Effective Risk and Consent](admission.md#effective-risk-and-consent)). Synthetic unit tests, recorded in the [evidence report](../traceability/evidence/go-unit-41c6f06.md), exercise the REQ-TRUST-0037 row for effective risk above an issuer or local bound, or `unknown` without the reviewed exception, `task_issuer_not_authorized`, now including the locally determined risk of matched target registrations. They also exercise precedence of the step-8 target row over step 9, and `task_unsupported` for an unrecognized risk factor type where `unknown` risk is permitted. The missing-capability row (DEV-590) and embedded replay (DEV-591) have no runtime evidence yet. The finding stays open.

### Unsupported Capability Evidence Addendum — October 20, 2026

[DEV-590](https://app.clickup.com/t/124kk1zcae4) rejects unsupported capabilities at Go `295dea9139956bc73eb04d521001ae76f0d574b8` ([Unsupported Capabilities](admission.md#unsupported-capabilities)). Synthetic unit tests, recorded in the [evidence report](../traceability/evidence/go-unit-295dea9.md), exercise the REQ-TRUST-0037 step-6 row for a valid field that requires a capability the implementation does not provide, `task_unsupported`, for each egress Network Context requirement, and its precedence against the other step-6 checks and steps 7 to 9. They also exercise `task_unsupported` for a registered NET-05 composite enabled by issuer and deployment, at step 8 after input and options, because composite steps cannot be matched yet. REQ-TRUST-0037 lists no step-8 row for this condition; the class follows REQ-TRUST-0036 and the step follows the composite-step matching REQ-TRUST-0017 places in step 8. Embedded replay (DEV-591) has no runtime evidence yet. The finding stays open.

## GO-SPEC-ISSUE-009: Admission Class for Proxy-Incompatible Task Parameters

**Current state:** Open; public proposal in review under [DEV-598](https://app.clickup.com/t/124kk1zcbfb). No class is chosen and Go behavior is unchanged. Raised October 19, 2026 while implementing target registration matching at admission step 8 under [DEV-588](https://app.clickup.com/t/124kk1zcae0).

### Identity, Baselines and Requirements

- **Reporter:** the Go implementation agent, under the task owner's instruction. **Duplicates:** none. [GO-SPEC-ISSUE-008](#go-spec-issue-008-admission-error-classes-for-unsupported-and-unmapped-rejections) mapped path refusal by issuer authority, local policy and the matched target registration, but not proxy compatibility.
- **Baseline:** Draft `0.1.0` at [d93aff93cbc17e7c065fb40d8f6e2e01a697bf35](https://github.com/tv-rain/barometer-spec/tree/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35).
  - Admission: [REQ-TRUST-0017](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0017) step 8 validates "proxy compatibility". [REQ-PROXY-0026](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0026) lists each test's compatibility and says that NET-01 with `network_path: proxy` and a NET-05 profile with an incompatible step are rejected. [REQ-PROXY-0027](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0027) requires validation before network activity, and [REQ-PROXY-0022](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0022) makes a proxy task eligible only when the complete test configuration is compatible.
  - Classes: the [REQ-TRUST-0037](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0037) table has no row for proxy incompatibility. [REQ-TRUST-0036](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0036) `task_unsupported` describes the implementation, not the task.
  - Execution: [REQ-PROXY-0028](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0028) and [REQ-RESULT-0020](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0020) define the execution state `proxy_path_unsupported` for "the test, input, or options require an operation that HTTP CONNECT cannot provide". [REQ-EXECUTION-0008](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0008) checks compatibility before a proxy attempt starts, and [REQ-EXECUTION-0010](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0010) requires an explicit non-censorship execution state for an incompatible proxy.
- **Implementation:** `tv-rain/barometer-go` at `2a142c111c013982d9a5cd9d09efb5b577b5b74d` (DEV-588 merge; code at `cb672923cdd7668324f6ac4802d6c60edca5572c`). Each `registry.Entry` carries a required `registry.ProxyRule`, but admission never calls it. Step-8 input and options failures use `task_schema_invalid` in the [admission design](admission.md#ordered-steps-and-error-classes); `task.ProxyPathUnsupported` exists as an execution reason. The `CC-PROXY-INITIAL-TEST-COMPATIBILITY` plan asserts atomic rejection before activity without a class.

### Finding and Interim Handling

The public text requires admission to validate proxy compatibility but assigns no admission class to a failure. It also defines `proxy_path_unsupported` as an execution state for the same condition, so it is unclear whether an incompatible task is rejected at admission or admitted and then recorded per attempt. Choosing a class, or choosing to admit, would redefine the contract.

**Blocked:** the registry proxy-compatibility part of admission step 8 and the class assertion of `CC-PROXY-INITIAL-TEST-COMPATIBILITY`. Go adoption and implementation are tracked by [DEV-599](https://app.clickup.com/t/124kk1zcbfc). **Interim handling:** admission does not evaluate `ProxyRule`; a proxy task is still subject to path authority and target registration. No proxy execution exists, so an incompatible task cannot reach the network through the proxy path. **Unaffected:** the rest of step 8, the step-8 path-authority and target-registration rows, and NET-05 composite steps, which already wait for the NET-05 codec.

### Impact, Reviews and Alternatives

- **Safety and interoperability:** every option makes no network contact through an incompatible path, so the direct safety difference is small. Admitting an incompatible task produces repeated non-runs that hosts may mistake for proxy or network failures. Hosts and issuers that branch on admission classes would diverge between implementations. Safety review of the chosen mapping is required.
- **Alternatives:** (a) `task_schema_invalid`, the class Go already uses for step-8 input and options, because compatibility is a public, implementation-independent rule over the task's own fields; (b) `task_unsupported`, which REQ-TRUST-0036 reserves for implementation capability; (c) `target_not_registered`, which concerns registration authority; (d) `task_issuer_not_authorized`, which concerns the authority intersection; (e) a new class, which is precise but extends the vocabulary; (f) admit and record `proxy_path_unsupported` per attempt, which contradicts REQ-TRUST-0017 step 8 and REQ-PROXY-0026 and REQ-PROXY-0027.
- **Proposed:** (a), with `task_unsupported` kept for a compatible combination the implementation does not provide through the proxy, and `proxy_path_unsupported` limited to an incompatibility found by the pre-attempt check that admission did not reject. The public record lists the tradeoffs.
- **Compatibility:** (a) to (d) add no class and no field; (e) needs a version decision.

### Ordering and Closure

Public record and disposition first, then internal task-trust review alignment if its validation plan names the class, then Go adoption in DEV-599, then bindings. Closure requires the public disposition and Go adoption evidence.
