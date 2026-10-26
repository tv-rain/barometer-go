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

**Status:** Public contract, required internal alignment and shared Go decoder adoption are merged. Foundation model consumers are delivered, with their scope recorded in the model documentation. The admission share is delivered by DEV-594 ([Admission Adoption Addendum](#admission-adoption-addendum--october-20-2026)). DEV-566 records evidence for the remaining existing consumers and an inventory of every JSON consumer ([Existing-Consumer Adoption Addendum](#existing-consumer-adoption-addendum--october-21-2026)). Registered production schemas and measurement runtime consumers ([DEV-622](https://app.clickup.com/t/124kk1zcge1)) now include TCP-01 and DNS-01, whose caller-specific evidence is recorded in [Measurement Requirements](measurement-requirements.md). Remaining production callers and binding/host consumers ([DEV-623](https://app.clickup.com/t/124kk1zcge2)) remain pending. The finding remains open; historical pending statements below do not describe the current Foundation merge state.

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

### Admission Adoption Addendum — October 20, 2026

DEV-594 delivers the admission share of the [downstream adoption task](https://app.clickup.com/t/124kk1zc4ag) against the accepted public contract at [05c1364af4f9a604b617fca45f060949b557d265](https://github.com/tv-rain/barometer-spec/commit/05c1364af4f9a604b617fca45f060949b557d265), unchanged at the current public baseline `d93aff93cbc17e7c065fb40d8f6e2e01a697bf35`.

- **Inventory.** Admission has one JSON consumer: `internal/taskdecode`, called by `internal/admission` at step 4 (`Envelope`, through `jsondecode.Decode` with 65,536-byte, 128-level and 32,768-node parse limits) and step 8 (`Fields` and `Parameters`, through `jsondecode.Validate` on the selected test's schemas). The root `Client.AdmitTask` passes the presented string as bytes without decoding it. Delivery records, source context and local policy have no JSON form, so they are not applicable.
- **Evidence.** `TestGateStepClasses` (gate) and `TestAdmitTaskStrictJSON` (public facade, both reviewed origins) reject escaped-equivalent and equal duplicate names at the top level and in `input`, `options`, `schedule` and `execution.limits`; escaped unknown names; `-0`, `1.0` and `1e0` for an integer field; an integer beyond the exact float range; trailing values, a byte-order mark and invalid UTF-8; and nesting beyond the parse depth. They count the complete-input limit on original escaped bytes and accept an escaped spelling of the exact `schema_version`. Every rejection leaves no handle, replay record or run, and `FuzzAdmitTask` checks the same property on arbitrary input ([Hostile-Input Evidence](admission.md#hostile-input-evidence)).
- **Not covered.** At DEV-594 the registered production test schemas did not exist, so their `input` and `options` decoding was exercised only through synthetic schemas; the TCP-01 and DNS-01 schemas now have caller-specific evidence in [Measurement Requirements](measurement-requirements.md). Producer, runtime and binding/host consumers remain under the downstream adoption task.

This finding stays open until the remaining consumers record evidence or reviewed non-applicability.

### Existing-Consumer Adoption Addendum — October 21, 2026

DEV-566 records the strict-JSON evidence for the JSON consumers that exist in this repository after DEV-594, against the accepted public contract at [05c1364af4f9a604b617fca45f060949b557d265](https://github.com/tv-rain/barometer-spec/commit/05c1364af4f9a604b617fca45f060949b557d265). The contract text is unchanged at the current cross-cutting baseline `e97502747c917a16bc2b71337535ff972bbd7d64`; the public disposition record there only adds its merge record. DEV-566 adds tests and records only; no decoding behavior changes.

**Inventory.** Every package that imports `internal/jsondecode` or `encoding/json` outside tests, and every public entry point that accepts JSON text:

| Consumer | Role | Disposition | Evidence |
|---|---|---|---|
| `internal/jsondecode` | Shared strict parser and schema validation | Delivered (DEV-477) | [Shared Strict JSON Decoding](json-decoding.md) |
| `internal/taskdecode`, called by `internal/admission` steps 4 and 8; `Client.AdmitTask` | Task envelope, `input` and `options` | Delivered (DEV-594) | [Admission Adoption Addendum](#admission-adoption-addendum--october-20-2026) |
| `internal/admission` signed package decoding (`signed.go`, DEV-612); `Client.AdmitTask` for remote origin | Signed remote package envelope | Adopted here | `TestAdmitTaskSignedPackageStrictJSON`; internal `TestCcTrustSignedRemotePackages` and `FuzzSignedPackage` ([Hostile-Input Evidence](admission.md#hostile-input-evidence)) |
| `internal/config`, called by `Client.Initialize` | Host configuration container | Model delivered (DEV-486); facade adopted here | `TestAtomicRejectionCorpus`, `FuzzConfiguration`; `TestInitializeStrictJSON`, `FuzzInitialize` ([Synthetic Evidence](configuration-models.md#synthetic-evidence)) |
| `internal/model/observation` | Protocol observations and Network Context: decoder and producer | Delivered (DEV-481) | [Observation models](observation-models.md); producer output re-decodes in `TestProtocolValidationPrecedenceAndStableEncoding` and `TestContextNormalizationStaysWithinASCIICodes` |
| `internal/result` | Measurement Results: decoder and producer | Delivered (DEV-482) | [Result models](result-models.md); producer output re-decodes in `FuzzResult` and `TestImmutableViewsAndProducerBuild` |
| `internal/bundle` | Result bundle: producer only, no decoder | Producer evidence adopted here | `TestProducedBundleIsStrictJSON` ([Synthetic Evidence](bundle-models.md#synthetic-evidence)) |
| `internal/execution`, `internal/persistence`, `internal/registry`, `internal/validation`, `internal/platform`; delivery records, source context and local policy | No serialized JSON form; `execution.Attempt` refuses JSON encoding | Not applicable | — |
| `internal/testkit` | Test-only fixtures, conformance runner, synthetic signer and replay store | Not applicable | `TestAdmissionHasNoNetworkDependency` requires `internal/testkit` to be absent from the production closure |
| Registered production test schemas and measurement runtime | TCP-01, DNS-01 and TLS-01 codecs through strict task stages; TCP, DNS and TLS observation producers re-decode their generated closed JSON | TCP-01, DNS-01 and TLS-01 portable adoption; other schemas deferred | [Current TCP-01 Execution](measurement-requirements.md#current-tcp-01-execution); [Current DNS-01 Execution](measurement-requirements.md#current-dns-01-execution); [Current TLS-01 Execution](measurement-requirements.md#current-tls-01-execution); [DEV-622](https://app.clickup.com/t/124kk1zcge1) |
| Native, Flutter and host validation | Not implemented; the gomobile smoke build is not binding evidence | Deferred | [DEV-623](https://app.clickup.com/t/124kk1zcge2) |

`internal/result` and `internal/model/observation` decode a second time with `encoding/json` into closed wire structures. That second pass runs only after `jsondecode.DecodeSelected` has accepted the same bytes and rejected duplicate, unknown and case-variant names, so its case-insensitive matching and last-member precedence cannot select a different value.

**Evidence.**

- **Signed package envelope.** `TestAdmitTaskSignedPackageStrictJSON` presents package bytes through `Client.AdmitTask`. Equal, conflicting and escaped-equivalent duplicate members, escaped unknown and case-variant names, a nested value, nesting beyond the parse depth, a trailing value, a byte-order mark, invalid UTF-8 in a name or value and a lone surrogate escape are `task_source_unauthenticated`. A fully escaped `task` member padded one byte over the 98,304-byte limit is `task_limit_exceeded` and admits at exactly the limit. Escaped spellings of a member name and a member value admit. Each rejection leaves no handle, replay record or run. The package has string members only, so integer spellings do not apply.
- **Host configuration.** `TestInitializeStrictJSON` presents configuration through `Client.Initialize` with trusted synthetic approvals. Duplicate and escaped-equivalent names at the top level and in `proxy`, `network_context` and `result_linkage`, escaped unknown and case-variant names, `-0`, decimal, exponent and quoted spellings for `port` and `rotation_period_seconds`, an integer beyond the exact float range, a trailing value, a byte-order mark, invalid UTF-8, a lone surrogate, nesting depth and escaped content one byte over the 16,384-byte limit are `not_initialized`. Each leaves the client uninitialized with no runtime generation and no linkage key call, and a later valid initialization succeeds. Escaped spellings of a name and an approved value, and escaped content at exactly the limit, are accepted. `FuzzInitialize` requires the facade outcome to match the decoder for the same bytes and checks the same no-side-effect property.
- **Producers.** The result, observation and configuration producers re-decode through their strict decoders in the tests listed above. The bundle has no decoder, so `TestProducedBundleIsStrictJSON` parses produced bundles with the strict shared parser, with and without linkage, at 1 and 16 results and at the largest result size.

**Not covered.** The 64-node configuration parse bound and the 6-node package bound are exercised only by the shared decoder tests, not at these facades. Producer evidence shows that the tested outputs satisfy the strict parser and re-decode; it is not a proof for every reachable value. Registered production test schemas, measurement runtime consumers and binding/host consumers remain under DEV-622 and DEV-623, so this finding stays open until they record evidence or reviewed non-applicability.

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

**Current state:** Open; public disposition accepted October 22, 2026 under [DEV-579](https://app.clickup.com/t/124kk1zcaae) in the public record [Boundary Between `not_run` and `incomplete`](https://github.com/tv-rain/barometer-spec/blob/main/specification-issues/not-run-incomplete-boundary.md); wording merged at public [`84f52f8`](https://github.com/tv-rain/barometer-spec/tree/84f52f8872a5af9e6d97894611e3b2eaba18aba8) and internal alignment merged; Go execution and result adoption delivered under [DEV-561](https://app.clickup.com/t/124kk1zc4a9); binding and host adoption pending. See the [disposition addendum](#disposition-addendum--october-22-2026), the [merge addendum](#merge-addendum--october-22-2026) and the [Go adoption addendum](#go-adoption-addendum--october-22-2026). The sections before it keep their original wording as historical evidence. Originally raised October 16, 2026 under [DEV-573](https://app.clickup.com/t/124kk1zc613) from the [DEV-572](https://app.clickup.com/t/124kk1zc5xp) case-plan triage.

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

### Disposition Addendum — October 22, 2026

The public maintainer accepted the resolution recommended in the public record on October 22, 2026, in the DEV-579 task review. The wording changes public text and is merged together with the record; the merged immutable revision is recorded here after merge. Draft `0.1.0` is retained, and no field, state or reason is added.

- **Case 1:** an attempt is `not_run` until the test starts, then `incomplete` if a cross-cutting condition stops it ([REQ-RESULT-0051](https://github.com/tv-rain/barometer-spec/blob/main/data-formats/measurement-result.md#REQ-RESULT-0051), alternative (a)). The test starts at its first test-specific network operation or, when that comes first, its own network-state observation, such as NET-01 collection or the NET-05 stable-network check. Admission, validation, context capture and acquisition, including traffic to a context endpoint, and constraint evaluation do not start it.
- **Case 2:** once the test records its terminal outcome, the attempt is `completed`; a later interruption or cleanup failure does not change the state, reason or `measurement` ([REQ-RESULT-0052](https://github.com/tv-rain/barometer-spec/blob/main/data-formats/measurement-result.md#REQ-RESULT-0052)).
- **Case 3:** NET-05 `network_not_validated` and `inconclusive` are retained with `execution.state: completed`; the NET-05 `incomplete` outcome stays with `incomplete`; an unsatisfied task-level `require_validated` constraint before the test starts is `not_run` with `constraints_unsatisfied` ([REQ-NET-05-0061](https://github.com/tv-rain/barometer-spec/blob/main/tests/net-05.md#REQ-NET-05-0061)).
- **Case 4:** a required network condition that cannot be determined is unsatisfied: `not_run` with `constraints_unsatisfied` ([REQ-EXECUTION-0031](https://github.com/tv-rain/barometer-spec/blob/main/data-formats/task-execution.md#REQ-EXECUTION-0031), replacing REQ-EXECUTION-0015). `network_context_unavailable` stays limited to required Network Context fields.
- **Network Context:** a `not_run` result must not contain `network_context`, even when a snapshot was captured ([REQ-RESULT-0053](https://github.com/tv-rain/barometer-spec/blob/main/data-formats/measurement-result.md#REQ-RESULT-0053), replacing REQ-RESULT-0024). This reduces retention; no additional privacy review is required.
- **Go effect:** no Go behavior changes in this task. `internal/result` currently enforces the retired REQ-RESULT-0024: it rejects a `not_run` result without `network_context` when the binding records a captured context. Adoption must make the decoder reject `network_context` in every `not_run` result, select states under REQ-RESULT-0051 and REQ-RESULT-0052 in runner and execution code (DEV-561), fill the expected values in the `M-*-TERMINATION` case plans, and rebaseline the traceability mapping so that current `requirement-ref` comments cite REQ-RESULT-0053 and REQ-EXECUTION-0031 instead of the retired IDs. Until that rebaseline, existing references remain valid for the pinned public baseline.
- **Internal alignment:** recorded in the internal NET-05 review, whose `inconclusive` wording is aligned with REQ-NET-05-0061.
- **Remaining before closure:** the merged public revision, the internal alignment merge, and Go adoption evidence. Binding and host adoption follow. No runtime, platform or field support is claimed.

### Merge Addendum — October 22, 2026

The sections above keep their wording as historical evidence, including their then-pending statements and links to public `main`.

- **Merged public revision:** [public PR #22](https://github.com/tv-rain/barometer-spec/pull/22) merged the accepted wording and catalog changes at [`84f52f8872a5af9e6d97894611e3b2eaba18aba8`](https://github.com/tv-rain/barometer-spec/tree/84f52f8872a5af9e6d97894611e3b2eaba18aba8). This is the immutable baseline for adoption: [REQ-RESULT-0051](https://github.com/tv-rain/barometer-spec/blob/84f52f8872a5af9e6d97894611e3b2eaba18aba8/data-formats/measurement-result.md#REQ-RESULT-0051), [REQ-RESULT-0052](https://github.com/tv-rain/barometer-spec/blob/84f52f8872a5af9e6d97894611e3b2eaba18aba8/data-formats/measurement-result.md#REQ-RESULT-0052), [REQ-RESULT-0053](https://github.com/tv-rain/barometer-spec/blob/84f52f8872a5af9e6d97894611e3b2eaba18aba8/data-formats/measurement-result.md#REQ-RESULT-0053), [REQ-EXECUTION-0031](https://github.com/tv-rain/barometer-spec/blob/84f52f8872a5af9e6d97894611e3b2eaba18aba8/data-formats/task-execution.md#REQ-EXECUTION-0031) and [REQ-NET-05-0061](https://github.com/tv-rain/barometer-spec/blob/84f52f8872a5af9e6d97894611e3b2eaba18aba8/tests/net-05.md#REQ-NET-05-0061); REQ-RESULT-0024 and REQ-EXECUTION-0015 are retired there. Draft `0.1.0` is retained.
- **Internal alignment:** merged through internal review repository PR #10 at `0414e543db2fea2bb156daaa823e0cc37c759ae0`.
- **Go state:** unchanged. The traceability mapping stays on public `4743b63`, where the retired IDs are still active, so current references remain valid until the rebaseline to `84f52f8` during adoption.
- **Remaining before closure:** Go adoption evidence, as listed under the disposition addendum. Binding and host adoption follow. No runtime, platform or field support is claimed.

### Go Adoption Addendum — October 22, 2026

DEV-561 adopts the disposition in Go execution and result code, as described in [State Mapping and GO-SPEC-ISSUE-004](execution-models.md#state-mapping-and-go-spec-issue-004):

- **Case 1 (REQ-RESULT-0051):** the engine records the test start at the first dial through the attempt scope or at the runner's explicit `measurement.Input.Start`. Before it, an outer condition is `not_run`; after it, `incomplete`. Context capture, dialer acquisition, local validation and run cancellation before the first attempt are `not_run`. Evidence: `TestHostCancellation`, `TestAttemptDeadlineBeforeTheTestStarts` and `TestCancellationAtEveryStage`.
- **Case 2 (REQ-RESULT-0052):** a runner's returned outcome is `completed` even when a cancellation or the deadline arrives while connections close, and results are written with a context the cancellation does not end. Evidence: `TestHostCancellation`, `TestAttemptScopeReleasesSockets` and `TestCancellationRacesWithCompletion`.
- **Case 3 (REQ-NET-05-0061):** no Go change is possible yet; no NET-05 runner exists (DEV-454). The [termination expectations](measurement-requirements.md#termination-states-after-go-spec-issue-004) record the expected states.
- **Case 4 (REQ-EXECUTION-0031):** already followed: undeterminable network conditions are unsatisfied at the start gates (DEV-636), which start no attempt and write no result.
- **Network Context (REQ-RESULT-0053):** `result.Decoder` rejects `network_context` in every `not_run` result and `Binding.CapturedContext` is removed. Evidence: `TestExecutionContextAndAttemptBinding`.
- **Case plans:** the eight `M-*-TERMINATION` plans in the measurement manifest keep their text at its unchanged baseline; their expected states under the disposition are recorded in [Termination States after GO-SPEC-ISSUE-004](measurement-requirements.md#termination-states-after-go-spec-issue-004) until the measurement manifest is rebaselined.
- **Traceability:** the cross-cutting manifest is rebaselined to public `84f52f8` together with this adoption, so the `requirement-ref` comments cite REQ-RESULT-0051, REQ-RESULT-0052 and REQ-RESULT-0053 instead of the retired REQ-RESULT-0024.
- **Remaining before closure:** binding and host adoption (DEV-455) and NET-05 runner evidence (DEV-657); TCP-01 and DNS-01 termination evidence is recorded at the measurement milestone ([Measurement Requirements](measurement-requirements.md#termination-states-after-go-spec-issue-004)). No runtime, platform or field support is claimed.

## GO-SPEC-ISSUE-005: Which Bytes Count Toward Attempt Byte Limits

**Current state:** Open; public disposition accepted October 22, 2026 under [DEV-580](https://app.clickup.com/t/124kk1zcaaf) in the public record [Bytes Counted Toward Attempt Byte Limits](https://github.com/tv-rain/barometer-spec/blob/main/specification-issues/attempt-byte-accounting.md); wording merged at public [`5d01879`](https://github.com/tv-rain/barometer-spec/tree/5d01879cfd1d47396c780befa2567f7e90439948) and internal alignment merged; generic stream-layer Go adoption delivered under [DEV-562](https://app.clickup.com/t/124kk1zc4ab) and the proxy path that keeps the `CONNECT` exchange below the metered scope under [DEV-645](https://app.clickup.com/t/124kk1zcrje); per-protocol adoption for TCP-01 and DNS-01 delivered under [DEV-656](https://app.clickup.com/t/124kk1zcw4e) and for the remaining tests pending. See the [disposition addendum](#disposition-addendum--october-22-2026-1), the [merge addendum](#merge-addendum--october-22-2026-1), the [generic-layer adoption addendum](#generic-layer-adoption-addendum--october-22-2026) and the [proxy-path addendum](#proxy-path-addendum--october-23-2026). The sections before it keep their original wording as historical evidence. Originally raised October 16, 2026 under [DEV-573](https://app.clickup.com/t/124kk1zc613) from the DEV-572 triage.

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

### Disposition Addendum — October 22, 2026

The public maintainer accepted alternative (a), as recommended in the public record, on October 22, 2026, in the DEV-580 task review. The wording changes public text and is merged together with the record; the merged immutable revision is recorded here after merge. Draft `0.1.0` is retained, and no registered number, field, state or reason changes.

- **Counted:** every byte Barometer writes to or reads from a TCP stream, the plaintext side of a TLS session or a UDP datagram payload for the test, including HTTP request lines, headers, bodies and HTTP/2 framing, response status lines and headers, DNS messages and retransmissions Barometer sends, DNS-over-TCP and DoT length prefixes, the complete DoH exchange, followed redirects, and every NET-05 primitive step toward the aggregate limits ([REQ-LIMIT-0034](https://github.com/tv-rain/barometer-spec/blob/main/library/registered-limits.md#REQ-LIMIT-0034)). Bytes read ahead into a buffer count when read; each read requests no more than the remaining budget, and a write that would exceed the remaining sent budget is not issued.
- **Not counted:** TLS handshake and records, TCP/IP and UDP/IP headers, operating-system retransmissions, proxy transport framing, the `CONNECT` request and response with the loopback proxy endpoint, resolution through a platform resolver API that does not expose its messages, and Network Context acquisition ([REQ-LIMIT-0033](https://github.com/tv-rain/barometer-spec/blob/main/library/registered-limits.md#REQ-LIMIT-0033), [REQ-EXECUTION-0032](https://github.com/tv-rain/barometer-spec/blob/main/data-formats/task-execution.md#REQ-EXECUTION-0032)).
- **Narrower bounds:** header, body, prefix and DNS message bytes also count toward the attempt limits.
- **Requirement IDs:** REQ-LIMIT-0003 is retired with successors REQ-LIMIT-0033 and REQ-LIMIT-0034, and REQ-EXECUTION-0014 with successor REQ-EXECUTION-0032. REQ-LIMIT-0006 and REQ-EXECUTION-0012 keep their IDs with editorial pointers.
- **Go effect:** no Go behavior changes in this task. Adoption must implement byte-budget accounting at the counted boundary for HTTP-01, NET-03, DNS-01, DNS-03, DNS-04 and the NET-05 aggregate ([DEV-562](https://app.clickup.com/t/124kk1zc4ab) and the per-protocol work under DEV-454), fill the expected boundary values in the `CC-LIMIT-PURPOSE` case plan and its `RV-CC-LIMIT-PURPOSE` review case, add scripted-transport cases, and rebaseline the traceability mapping so that current `requirement-ref` comments cite REQ-LIMIT-0033, REQ-LIMIT-0034 and REQ-EXECUTION-0032 instead of the retired IDs. The mapping stays on public `4743b63`, where the retired IDs are still active, so existing references remain valid until that rebaseline.
- **Internal alignment:** recorded in the internal registered-limits review and NET-05 review, aligned with the counted boundary.
- **Remaining before closure:** the merged public revision, the internal alignment merge, and Go adoption evidence. Binding and host adoption follow. No runtime, platform or field support is claimed.

### Merge Addendum — October 22, 2026

The sections above keep their wording as historical evidence, including their then-pending statements and links to public `main`.

- **Merged public revision:** [public PR #24](https://github.com/tv-rain/barometer-spec/pull/24) merged the accepted wording and catalog changes at [`5d01879cfd1d47396c780befa2567f7e90439948`](https://github.com/tv-rain/barometer-spec/tree/5d01879cfd1d47396c780befa2567f7e90439948). This is the immutable baseline for adoption: [REQ-LIMIT-0033](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/library/registered-limits.md#REQ-LIMIT-0033), [REQ-LIMIT-0034](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/library/registered-limits.md#REQ-LIMIT-0034) and [REQ-EXECUTION-0032](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/data-formats/task-execution.md#REQ-EXECUTION-0032); REQ-LIMIT-0003 and REQ-EXECUTION-0014 are retired there. Draft `0.1.0` is retained.
- **Internal alignment:** merged through internal review repository PR #11 at `dc421a80b3fb19bd07b17a0a4181390ad6e50d2b`.
- **Go state:** unchanged. The traceability mapping stays on public `4743b63`, where the retired IDs are still active, so current references remain valid until the rebaseline to `5d01879` during adoption.
- **Remaining before closure:** Go adoption evidence, as listed under the disposition addendum. Binding and host adoption follow. No runtime, platform or field support is claimed.

### Generic-Layer Adoption Addendum — October 22, 2026

DEV-562 adopts the disposition for the generic stream layer, as described in [Attempt Byte and Connection Budgets](execution-models.md#attempt-byte-and-connection-budgets). Every runner connection is the attempt scope's metered `measurement.Conn`, which is where test bytes cross a TCP stream at REQ-LIMIT-0034's boundary:

- **Sent and received (REQ-LIMIT-0034):** a write that would exceed the remaining `max_bytes_sent` is not issued, a partial write counts the bytes accepted, each read requests no more than the remaining `max_bytes_received`, and bytes read ahead count when read. Evidence: `TestSentBudgetBoundaries`, `TestReceivedBudgetBoundaries`, `TestBudgetAcrossConnectionReuse`, `TestReadAheadIsCountedAndCapped` and `TestPartialTransfersAreCountedExactly`; the loopback cases compare the metered counts with the bytes the server read and wrote.
- **Stop (REQ-LIMIT-0027, REQ-LIMIT-0028):** an operation the budget cannot cover ends the attempt as `incomplete` with `resource_limit`, closes every connection and refuses further I/O without another dialer. Evidence: `TestExhaustedBudgetDispatchesNoFurtherIO` and `TestExhaustionReleasesBlockedSockets`.
- **Connections (REQ-LIMIT-0006):** every dial, including a failed one, counts before it is issued. Evidence: `TestConnectionBudgetBoundaries`, `TestFailedConnectCountsTowardConnections` and `TestZeroLimitsPermitNoTraffic`.
- **Not counted (REQ-LIMIT-0033):** TLS records, IP headers, retransmissions, proxy framing and the loopback-proxy `CONNECT` exchange lie below the stream a runner receives, and Network Context acquisition is outside the scope.
- **Traceability:** the cross-cutting manifest is rebaselined to public `5d01879` together with this adoption, so the `requirement-ref` comments cite REQ-LIMIT-0033, REQ-LIMIT-0034 and REQ-EXECUTION-0032 instead of the retired REQ-LIMIT-0003 and REQ-EXECUTION-0014.
- **Per-protocol evidence delivered:** TCP-01 opens one metered connection and sends no application bytes (`TestTcp01ConnectionMeter`, `TestMTcp01ConnectionBudget`); DNS-01 counts exact DNS messages and both TCP length prefixes at the metered boundary, with its one UDP socket or TCP connection counted toward `max_connections` (`TestCcLimitPurposeDNSRows`, `TestCcProxyOverviewDNSRow`, `TestMDns01ConnectionBudget`), under [DEV-656](https://app.clickup.com/t/124kk1zcw4e). TLS-01 meters the plaintext side of its session instead of the handshake and record bytes, on the direct path and through the proxy tunnel (`TestTls01PlaintextAccounting`, `TestScopeSessionMetersPlaintext`, `TestMTls01Limits`).
- **Remaining before closure:** per-protocol evidence that HTTP-01, NET-03, DNS-03, DNS-04 and the NET-05 aggregate route their counted bytes (headers, DNS messages, DoH framing, DoT prefixes, plaintext through the TLS session boundary and UDP payloads) through this scope, with the remaining `CC-LIMIT-PURPOSE` protocol rows (DEV-657, DEV-658); and binding and host adoption. No runtime, platform or field support is claimed.

### Proxy-Path Addendum — October 23, 2026

DEV-645 adds the proxy path ([Proxy Path](execution-models.md#proxy-path)) at public `5d01879`. `execution.ProxyDialers` performs the `CONNECT` exchange on the loopback connection before it returns the tunnel, and the attempt scope wraps the proxy dialer, so the exchange lies below the metered stream ([REQ-LIMIT-0033](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/library/registered-limits.md#REQ-LIMIT-0033)) while each tunnel counts toward `max_connections` before it is opened.

- **Evidence:** `TestProxyAttemptRunsThroughTunnel` compares the metered counts with the bytes the loopback target and the scripted proxy relayed, and the proxy's own count of the `CONNECT` exchange; `TestConnectExchangeIsNotMetered` completes an attempt whose byte limits equal the tunnel bytes, smaller than the `CONNECT` request, and opens no tunnel under a zero connection limit. Synthetic runners and the test-only scripted proxy; both passed at the DEV-648 milestone ([evidence at d788a80](../traceability/evidence/go-unit-d788a80.md#UT.internal.execution.TestProxyAttemptRunsThroughTunnel)).
- **Remaining before closure:** per-protocol evidence with the remaining `CC-LIMIT-PURPOSE` protocol rows (DEV-657, DEV-658), and binding and host adoption. No runtime, platform or field support is claimed.

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

**Current state:** Open; public disposition merged at [5a37e82865a08922c4821c7318656286933421e5](https://github.com/tv-rain/barometer-spec/tree/5a37e82865a08922c4821c7318656286933421e5) and adopted in Go decoding and validation by [DEV-582](https://app.clickup.com/t/124kk1zcaah). Go now rejects `risk_factors: []` for `medium`, `high` and `unknown` risk; `timezone` validation was already conforming and is now evidenced. Admission applies this validation at step 6. `TestTcp01AdmissionRiskAndTimezoneDomains` now asserts both domains through `Client.AdmitTask` with the production TCP-01 codec. Other planned family cases and binding and host evidence remain open. See the [adoption addendum](#adoption-addendum--october-19-2026) and the [current-state addendum](#current-state-addendum--october-21-2026). The sections before the addenda keep their original wording as historical evidence. Originally raised October 16, 2026 under [DEV-573](https://app.clickup.com/t/124kk1zc613) from the DEV-572 triage. This finding affects merged Foundation code.

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

### Current-State Addendum — October 21, 2026

Recorded under DEV-624 against public `4743b636f29feb1f0f8d042866d04ccda3532938`, where the cited safety and schedule text is unchanged.

- **Admission:** the admission pipeline now exists ([Admission Design](admission.md)). Step 6 applies `validation.Common`, so the REQ-SAFETY-0031 and REQ-SCHEDULE-0043 rules apply to every task presented for admission. `TestTcp01AdmissionRiskAndTimezoneDomains` asserts `risk_factors: []` for `medium`, `high` and `unknown` risk and non-`local` `timezone` values through production TCP-01 admission, including positive domain controls. Binding and host evidence remain pending.
- **Remaining scope:** the complete planned `CC-SAFETY-RISK-FACTORS` family; the planned `CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-FIELDS` family, owned by execution (DEV-576); and binding and host adoption (DEV-455). The production TCP-01 admission-domain rows are covered as described above. External producers and stored tasks remain unassessed. No specialist safety review is claimed.

The closure criterion, evidence or explicit non-applicability for the remaining scope, is not met. The finding stays open.

## GO-SPEC-ISSUE-008: Admission Error Classes for Unsupported and Unmapped Rejections

**Current state:** Open; public disposition merged at [d93aff93cbc17e7c065fb40d8f6e2e01a697bf35](https://github.com/tv-rain/barometer-spec/tree/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35) and adopted in the Go admission design, case plans, traceability and `task.AdmissionError` by [DEV-597](https://app.clickup.com/t/124kk1zcb7t). The admission tasks DEV-586 to DEV-592 implement the rows through the public facade, with unit-test evidence recorded in their addenda; the proxy-compatibility row is tracked separately as [GO-SPEC-ISSUE-009](#go-spec-issue-009-admission-class-for-proxy-incompatible-task-parameters) and most admission case families stay planned. Internal task-trust review alignment is merged, DEV-624 corrects the step-1 precedence between the delivery binding and trusted time, and the remaining case families and binding and host adoption remain open. The human maintainer confirmed the class-mapping safety review on October 21, 2026 at the [DEV-575 assessment](https://github.com/tv-rain/barometer/blob/main/docs/admission-assessment.md#reviews). This is maintainer review, not independent specialist review. See the [disposition addendum](#disposition-addendum--october-19-2026-1), the [adoption addendum](#adoption-addendum--october-19-2026-1), the [replay and facade evidence addendum](#replay-and-facade-evidence-addendum--october-20-2026) and the [current-state addendum](#current-state-addendum--october-21-2026-1). The sections before the addenda keep their original wording as historical evidence. Originally raised October 19, 2026 under [DEV-583](https://app.clickup.com/t/124kk1zcadk) while recording the [admission design](admission.md).

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

### Replay and Facade Evidence Addendum — October 20, 2026

[DEV-591](https://app.clickup.com/t/124kk1zcae5) executes `CC-TRUST-REPLAY-DUPLICATE-AND-ROLLBACK-HANDLING`, including the REQ-TRUST-0038 row: with a replay-storage fault injected, an embedded task fails closed with `task_trust_unavailable` and is never accepted, assigned a handle or made executable. [DEV-592](https://app.clickup.com/t/124kk1zcae7) runs every step through `Client.AdmitTask` and returns the code-only class of the first failing step. [DEV-594](https://app.clickup.com/t/124kk1zcaea) adds facade-level strict JSON, hostile-source and fuzz evidence, in which every rejection is exactly one public class with no handle, replay record or run ([Hostile-Input Evidence](admission.md#hostile-input-evidence)). `CC-TRUST-ADMISSION-PROCEDURE` and `CC-TRUST-ADMISSION-ERRORS` stay planned for production proxy-compatibility rules, transport and DNS fakes, result-sink and bundle assertions, and captured log output (DEV-454); egress acquisition is delivered by DEV-647. Whether the unit-test evidence suffices to close this finding is left to review. The finding stays open.

### Current-State Addendum — October 21, 2026

Recorded under DEV-624 against public `4743b636f29feb1f0f8d042866d04ccda3532938`, where the REQ-TRUST-0036, REQ-TRUST-0038 and [REQ-TRUST-0047](https://github.com/tv-rain/barometer-spec/blob/4743b636f29feb1f0f8d042866d04ccda3532938/library/task-trust.md#REQ-TRUST-0047) text is unchanged from `e975027`.

- **Internal alignment:** the internal task-trust review validation plan was aligned with the admission error classes under DEV-596 (`barometer-spec-internal` PR #5, merge `56ee602b858a0db63cc0b8d1d8a3dc9b3c35e66d`).
- **Step-1 precedence correction:** `DeliveryStore.Resolve` read trusted time before it checked the delivery binding, so an unknown, foreign-instance, stale-generation, invalidated, closed-store or byte-mismatched reference presented while the clock was unavailable returned `task_trust_unavailable`. The REQ-TRUST-0047 table places the binding row at step 1, before the trusted-time row, so the first failing step was misreported. DEV-624 checks the binding first: such a reference is now `task_source_unauthenticated` whatever the clock state, and only a held record that matches the presented bytes, whose expiry needs trusted time, is `task_trust_unavailable`. Remote records stay single-use, embedded records are not consumed and resolution stays free of network activity ([Delivery-Record Seam](admission.md#delivery-record-seam)). `TestResolveChecksBindingBeforeTrustedTime` covers each case; with the previous order its binding cases fail.
- **Superseded traceability note:** the manifest note for this finding that began "no admission implementation returns the classes yet" predates DEV-586 to DEV-594 and is replaced in the DEV-624 rebaseline.
- **Step-8 input and options:** no public row names the class of an input or options failure; this is recorded separately as [GO-SPEC-ISSUE-010](#go-spec-issue-010-admission-class-for-step-8-input-and-options-failures).
- **Remaining scope:** `CC-TRUST-ADMISSION-PROCEDURE`, which needs transport and DNS fakes and production proxy-compatibility rules (DEV-454); `CC-TRUST-ADMISSION-ERRORS`, which needs result-sink and bundle fakes and captured log output (DEV-454); and binding and host adoption, unassessed until those components exist (DEV-455).

The closure criteria are not met. The finding stays open.

### Egress Acquisition Addendum — October 23, 2026

DEV-647 adds egress Network Context acquisition and removes the admission egress check: a valid task with `required_fields` or `allowed_country_codes` is admitted, and its requirements are evaluated before each attempt ([Egress Network Context](execution-models.md#egress-network-context)). `UT.internal.admission.TestGateAdmitsEgressContextRequirements` and `UT.internal.admission.TestGateEgressContextKeepsOtherSteps` passed at the DEV-648 milestone ([evidence at d788a80](../traceability/evidence/go-unit-d788a80.md#UT.internal.admission.TestGateAdmitsEgressContextRequirements)). The other `task_unsupported` rows are unchanged.

- **Remaining scope:** `CC-TRUST-ADMISSION-PROCEDURE`, which needs transport and DNS fakes and production proxy-compatibility rules (DEV-454); `CC-TRUST-ADMISSION-ERRORS`, which needs result-sink and bundle fakes and captured log output (DEV-454); and binding and host adoption (DEV-455). The finding stays open.

## GO-SPEC-ISSUE-009: Admission Class for Proxy-Incompatible Task Parameters

**Current state:** Open; public disposition merged at [59b192d498bf011e331928d192702e1205a89f35](https://github.com/tv-rain/barometer-spec/tree/59b192d498bf011e331928d192702e1205a89f35) (public PR #16, [DEV-598](https://app.clickup.com/t/124kk1zcbfb)) and adopted in admission step 8, the admission design, case plans and traceability at public `e975027` by [DEV-599](https://app.clickup.com/t/124kk1zcbfc). Internal task-trust review alignment is merged (`barometer-spec-internal` PR #8). The execution-time pre-attempt check exists since [DEV-645](https://app.clickup.com/t/124kk1zcrje). Production compatibility rules and NET-05 whole-profile rejection remain open. The human maintainer confirmed the class-mapping safety review on October 21, 2026 at the [DEV-575 assessment](https://github.com/tv-rain/barometer/blob/main/docs/admission-assessment.md#reviews). This is maintainer review, not independent specialist review. See the [disposition addendum](#disposition-addendum--october-19-2026-2), the [adoption addendum](#adoption-addendum--october-21-2026), the [current-state addendum](#current-state-addendum--october-21-2026-2) and the [execution-check addendum](#execution-check-addendum--october-23-2026). The sections before the addenda keep their original wording as historical evidence. Raised October 19, 2026 while implementing target registration matching at admission step 8 under [DEV-588](https://app.clickup.com/t/124kk1zcae0).

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

### Disposition Addendum — October 19, 2026

The public record [Admission Class for Proxy-Incompatible Task Parameters](https://github.com/tv-rain/barometer-spec/blob/59b192d498bf011e331928d192702e1205a89f35/specification-issues/proxy-compatibility-admission-class.md) was prepared under DEV-598. The project owner and public maintainer, Lev, accepted alternative (a) on October 19, 2026; the accepted wording merged with public PR #16 at `59b192d`, retaining Draft `0.1.0`, and is unchanged in the current baseline `e975027`.

- **Class:** [REQ-TRUST-0039](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0039) rejects a task with `network_path: proxy` whose test, input or options are not compatible with HTTP CONNECT under REQ-PROXY-0026, including a NET-05 profile with any incompatible step, at step 8 with `task_schema_invalid`. Within step 8, compatibility is evaluated after input and options and before path authority and target registration. `task_unsupported` stays the class for a combination REQ-PROXY-0026 permits but an implementation does not provide through the proxy.
- **Execution state:** [REQ-PROXY-0042](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/proxy.md#REQ-PROXY-0042) states that an incompatible task is not admitted to fail at execution; `proxy_path_unsupported` records only an incompatibility found by the pre-attempt check of REQ-EXECUTION-0008 in an admitted task.
- **Requirement IDs:** REQ-TRUST-0039 and REQ-PROXY-0042 are new. The later signed remote task change (`e975027`) relinks REQ-TRUST-0039's step reference to REQ-TRUST-0046; its text is otherwise unchanged.
- **Reviews:** internal task-trust review alignment is required by the disposition and has not been recorded. The required safety review of the mapping is not represented as completed.

### Adoption Addendum — October 21, 2026

[DEV-599](https://app.clickup.com/t/124kk1zcbfc) adopts the disposition against public Draft `0.1.0` at [e97502747c917a16bc2b71337535ff972bbd7d64](https://github.com/tv-rain/barometer-spec/tree/e97502747c917a16bc2b71337535ff972bbd7d64), the baseline of the current traceability manifest. The internal alignment was explicitly left as a separate gap by the task owner; it does not block the Go adoption of public text.

- **Admission:** Go commit `d309b5b08ad27449523990d3eb1dca080209b757`. `Admitter.Gate` passes the codec's normalized parameters of a `proxy` task to the registered test's `registry.ProxyRule` after input and options and before path authority, target registration and the step-9 risk bounds, and rejects an incompatible task with `task_schema_invalid`. `direct` tasks are not checked. See [Proxy Compatibility](admission.md#proxy-compatibility).
- **Evidence:** `TestGateProxyCompatibility` and `TestGateCompositeBeforeProxyCompatibility` use synthetic test-scoped rules; all 235 top-level tests passed at `d309b5b` ([evidence](../traceability/evidence/go-unit-d309b5b.md)). The recording transport sees zero calls on every case.
- **Case plans and traceability:** `CC-PROXY-INITIAL-TEST-COMPATIBILITY` and `CC-TRUST-ADMISSION-PROCEDURE` assert `task_schema_invalid` and the step-8 precedence; both stay `planned`. The [cross-cutting manifest](../traceability/cross-cutting.json) maps REQ-TRUST-0039 and REQ-PROXY-0042 and links this finding as `accepted_adoption_pending`; REQ-TRUST-0039, REQ-PROXY-0042, REQ-PROXY-0026 and REQ-PROXY-0027 are `partial`. None is `implemented`.
- **NET-05:** the composite-capability rejection (DEV-590) still runs first, so an enabled NET-05 task is `task_unsupported` and its rule is not consulted. Rejecting a profile with an incompatible step as `task_schema_invalid` needs the NET-05 codec (DEV-454).
- **Remaining scope:** production compatibility rules for the registered tests and NET-05 (DEV-454), a case for `task_unsupported` on a permitted but unprovided combination once a production rule distinguishes one, the execution-time pre-attempt check and `proxy_path_unsupported` (DEV-578), internal task-trust review alignment, and the safety review of the mapping. Binding and host adoption is unassessed. This finding stays open until those are recorded.

### Current-State Addendum — October 21, 2026

Recorded under DEV-624 against public `4743b636f29feb1f0f8d042866d04ccda3532938`, where REQ-TRUST-0039 and REQ-PROXY-0042 are unchanged from `e975027`.

- **Internal alignment:** the internal task-trust review was aligned with REQ-TRUST-0039 and REQ-PROXY-0042 under DEV-598 (`barometer-spec-internal` PR #8, merge `409967428593200514b23523f8a132a0e24afb9a`), after the Go adoption.
- **Remaining scope:** production compatibility rules for the registered tests (DEV-454); rejection of a NET-05 profile with an incompatible step as a whole, which needs the NET-05 codec (DEV-454); a case for `task_unsupported` on a permitted but unprovided combination once a production rule distinguishes one; and binding and host adoption (DEV-455).

The closure criteria are not met. The finding stays open.

### Execution-Check Addendum — October 23, 2026

DEV-645 adds the pre-attempt check of [REQ-EXECUTION-0008](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/data-formats/task-execution.md#REQ-EXECUTION-0008) at public `5d01879` ([Proxy Path](execution-models.md#proxy-path)). Before each proxy attempt the engine re-applies the registered proxy rule to the admitted parameters, after the configured-proxy and policy conditions, and an incompatibility records `not_run` with `proxy_path_unsupported` and makes no contact, as [REQ-PROXY-0042](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/library/proxy.md#REQ-PROXY-0042) limits that state.

- **Reachability:** admission rejects every incompatible task, registry rules are immutable and a policy replacement admits every task again, so in this build an admitted task cannot fail the check. `TestPathCheckMakesNoContact` and `TestCcExecutionFieldsNetworkPath` produce the state with a scripted check.
- **Remaining scope:** production compatibility rules for the registered tests after TCP-01 and DNS-01, which DEV-661 and DEV-662 delivered, and NET-05 whole-profile rejection (DEV-657, DEV-658); a case for `task_unsupported` on a permitted but unprovided combination once a production rule distinguishes one; and binding and host adoption (DEV-455). The finding stays open.

## GO-SPEC-ISSUE-010: Admission Class for Step-8 Input and Options Failures

**Current state:** Accepted and adopted. The public disposition (alternative (a)) merged at [`12062471875bbfbc55eafc94dc5a7ea4cd07ca5e`](https://github.com/tv-rain/barometer-spec/tree/12062471875bbfbc55eafc94dc5a7ea4cd07ca5e) (public PR #30, [DEV-663](https://app.clickup.com/t/124kk1zcw4r)); internal alignment and the maintainer safety review of the mapping are merged; Go adopts it with synthetic cases and unchanged behavior. Classes and order the public text leaves unnamed continue in [GO-SPEC-ISSUE-016](#go-spec-issue-016-unnamed-classes-and-order-in-step-8-input-and-options). See the [adoption addendum](#adoption-addendum--october-26-2026). The sections before it keep their original wording as historical evidence. Raised October 21, 2026 under DEV-624 from the observation recorded during DEV-598.

### Identity, Baselines and Requirements

- **Reporter:** the Go implementation agent, under the task owner's instruction. **Duplicates:** none. [GO-SPEC-ISSUE-008](#go-spec-issue-008-admission-error-classes-for-unsupported-and-unmapped-rejections) mapped the rows its proposal listed and did not list this one. [GO-SPEC-ISSUE-009](#go-spec-issue-009-admission-class-for-proxy-incompatible-task-parameters) mapped proxy compatibility only; its public record notes, as a related observation not proposed there, that no public requirement names the class of a step-8 input or options failure ([Admission Class for Proxy-Incompatible Task Parameters](https://github.com/tv-rain/barometer-spec/blob/59b192d498bf011e331928d192702e1205a89f35/specification-issues/proxy-compatibility-admission-class.md#alternatives-and-proposed-public-resolution)).
- **Baseline:** Draft `0.1.0` at [4743b636f29feb1f0f8d042866d04ccda3532938](https://github.com/tv-rain/barometer-spec/tree/4743b636f29feb1f0f8d042866d04ccda3532938).
  - Admission: step 8 of [REQ-TRUST-0046](https://github.com/tv-rain/barometer-spec/blob/4743b636f29feb1f0f8d042866d04ccda3532938/library/task-trust.md#REQ-TRUST-0046) validates "all input, options, proxy compatibility, target tuples, redirect registrations, and composite steps atomically". The [REQ-TRUST-0047](https://github.com/tv-rain/barometer-spec/blob/4743b636f29feb1f0f8d042866d04ccda3532938/library/task-trust.md#REQ-TRUST-0047) table has two step-8 rows, for a path refused by issuer authority or local policy and by the matched target registration, and [REQ-TRUST-0039](https://github.com/tv-rain/barometer-spec/blob/4743b636f29feb1f0f8d042866d04ccda3532938/library/task-trust.md#REQ-TRUST-0039) names `task_schema_invalid` for proxy incompatibility only. REQ-TRUST-0047 applies a class another requirement names for a condition; none names one for invalid input or options.
  - Input and options: [REQ-INPUT-0006](https://github.com/tv-rain/barometer-spec/blob/4743b636f29feb1f0f8d042866d04ccda3532938/data-formats/task-input.md#REQ-INPUT-0006) and [REQ-OPTIONS-0009](https://github.com/tv-rain/barometer-spec/blob/4743b636f29feb1f0f8d042866d04ccda3532938/data-formats/task-options.md#REQ-OPTIONS-0009) reject unknown properties, [REQ-INPUT-0014](https://github.com/tv-rain/barometer-spec/blob/4743b636f29feb1f0f8d042866d04ccda3532938/data-formats/task-input.md#REQ-INPUT-0014) rejects an input that cannot be normalized safely, and [REQ-OPTIONS-0019](https://github.com/tv-rain/barometer-spec/blob/4743b636f29feb1f0f8d042866d04ccda3532938/data-formats/task-options.md#REQ-OPTIONS-0019) says that when options are unsupported, invalid or unsafe the probe "should reject or skip the task with an appropriate task or execution status", without naming a class or choosing between rejection and an execution status.
  - Classes: [REQ-TRUST-0034](https://github.com/tv-rain/barometer-spec/blob/4743b636f29feb1f0f8d042866d04ccda3532938/library/task-trust.md#REQ-TRUST-0034) and [REQ-TRUST-0036](https://github.com/tv-rain/barometer-spec/blob/4743b636f29feb1f0f8d042866d04ccda3532938/library/task-trust.md#REQ-TRUST-0036).
- **Implementation:** `tv-rain/barometer-go` at `3aece5ef33f6c151a7f92c25a2fdf44116519bad` (DEV-599 merge). `Admitter.Gate` validates input and options against the selected test's schemas and codec and returns `task_schema_invalid` on failure, or `task_limit_exceeded` when a test-specific byte cap stricter than the registered `input` or `options` maximum is exceeded ([Implemented Gates](admission.md#implemented-gates)). The registered 49,152-byte `input` and 8,192-byte `options` maxima are checked earlier, at step 6. No production test codec exists (DEV-622); the synthetic test-scoped codecs exercise the path.

### Finding and Interim Handling

The class of a step-8 input or options failure is unmapped. Go selected `task_schema_invalid` by the class name, and `task_limit_exceeded` for a stricter test byte cap, before REQ-TRUST-0037 and its successor REQ-TRUST-0047 made classes normative; REQ-TRUST-0039 later placed proxy compatibility after input and options and gave it the same class, which suggests but does not state the mapping. REQ-OPTIONS-0019 also leaves open whether unsupported options are rejected at admission or skipped with an execution status.

**Blocked:** nothing in current approved work; an exact-class assertion for these rows in public conformance cases stays provisional. **Interim handling:** no behavior change. Go keeps `task_schema_invalid` and `task_limit_exceeded` for these rows, and the Go admission design marks them as unmapped. **Unaffected:** the proxy-compatibility, path-authority, target-registration and composite rows of step 8, and the step-6 `input` and `options` byte maxima, which are registered limits.

### Impact, Reviews and Alternatives

- **Safety and interoperability:** every alternative rejects or skips before network activity, so the direct safety difference is small. Hosts and issuers that branch on admission classes could diverge between implementations, and skipping with an execution status would admit a task that cannot run. Safety review of the chosen mapping is required.
- **Alternatives:** (a) a REQ-TRUST-0047 row mapping invalid input or options to `task_schema_invalid` and a stricter test-specific limit to `task_limit_exceeded`, matching current Go behavior; (b) `task_schema_invalid` for every step-8 input or options failure, including test-specific byte caps; (c) `task_unsupported` for an option value a test defines but an implementation does not provide, with (a) or (b) for invalid values; (d) admit and record an execution status, as REQ-OPTIONS-0019 permits, which conflicts with step 8 rejecting atomically.
- **Compatibility:** (a) to (c) add no class or field and would be a Draft `0.1.0` clarification; (c) narrows the use of an existing class and needs its own review; (d) changes admission outcomes.

### Ordering and Closure

Public record and disposition first, then internal task-trust review alignment if its validation plan names the class, then Go adoption with exact-class cases, then bindings. Closure requires the public disposition, its immutable revision if wording changes, and Go adoption evidence or recorded non-applicability.

### Adoption Addendum — October 26, 2026

- **Public:** [Admission Class for Step-8 Input and Options Failures](https://github.com/tv-rain/barometer-spec/blob/12062471875bbfbc55eafc94dc5a7ea4cd07ca5e/specification-issues/step-8-input-options-admission-class.md), merged through public PR #30 at `12062471875bbfbc55eafc94dc5a7ea4cd07ca5e`, retaining Draft `0.1.0`. [REQ-TRUST-0056](https://github.com/tv-rain/barometer-spec/blob/12062471875bbfbc55eafc94dc5a7ea4cd07ca5e/library/task-trust.md#REQ-TRUST-0056) makes an input or options validation failure at step 8 `task_schema_invalid`, and a stricter test-defined byte limit, checked before the content it bounds, `task_limit_exceeded`. [REQ-OPTIONS-0033](https://github.com/tv-rain/barometer-spec/blob/12062471875bbfbc55eafc94dc5a7ea4cd07ca5e/data-formats/task-options.md#REQ-OPTIONS-0033) replaces the retired REQ-OPTIONS-0019: invalid or prohibited options are rejected at admission and never admitted to be skipped.
- **Internal:** `barometer-spec-internal` PR #14, merged at `bf2fce978f526fcc80aac2cf9d568352a45799f3`, aligned the task-trust review validation plan. PR #15, merged at `7868ed6ca180ef9ab6330d4256070a80d4402cd7`, records the safety review of the mapping by the project maintainer on October 26, 2026: passed, with one follow-up finding carried into GO-SPEC-ISSUE-016. It is a maintainer review, not an independent specialist review.
- **Go:** behavior is unchanged; `Admitter.Gate` already returns these classes ([Implemented Gates](admission.md#implemented-gates)). `TestGateInputOptionsClasses` asserts each class through the gate with a synthetic test-scoped codec, for embedded and remote origin: unknown input and option members, values outside a field range or of the wrong type and a codec normalization failure are `task_schema_invalid`; options or an option value over a stricter test byte cap, including invalid content over the cap, are `task_limit_exceeded`; the codec is not reached for a schema or limit failure, and the recording transport sees zero calls. The existing `TestGateProxyCompatibility` cases keep invalid input and options before proxy compatibility. The current reference-source pin moves to `1206247`; current comments cite REQ-TRUST-0056 and REQ-OPTIONS-0033, and the retired REQ-OPTIONS-0019 keeps a `requirement-history` citation. Traceability manifests stay at their milestone baselines until the next milestone update.
- **Not applicable yet:** no Draft `0.1.0` test defines a stricter `input` or `options` byte limit, so the byte-limit rows have synthetic evidence only. The TCP-01 and DNS-01 production codecs use the staged contract of the codec pattern and define no stricter byte limit.
- **Remaining scope:** binding and host adoption (DEV-455). The unnamed classes and order are tracked in GO-SPEC-ISSUE-016. The finding is closed for the Go core since the adoption merged at `bb5df9e`.

## GO-SPEC-ISSUE-011: Due Times and Catch-Up for Interval Repetition

**Current state:** Open, raised October 22, 2026 under DEV-636 while implementing `StartRun`. The public record [Interval Due Times and Catch-Up](https://github.com/tv-rain/barometer-spec/blob/9804ab34f21863723a94caba0541fd4e64870fce/specification-issues/interval-due-times-catch-up.md) is open with no disposition; public maintainers own the disposition, and [DEV-651](https://app.clickup.com/t/124kk1zcujd) owns its follow-up. Go implements the conservative interim rule below, as approved by the task owner. See the [public record addendum](#public-record-addendum--october-22-2026).

### Identity, Baselines and Requirements

- **Reporter:** the Go implementation agent, under the task owner's instruction. **Duplicates:** none.
- **Baseline:** Draft `0.1.0` at [5d01879cfd1d47396c780befa2567f7e90439948](https://github.com/tv-rain/barometer-spec/tree/5d01879cfd1d47396c780befa2567f7e90439948).
  - [REQ-SCHEDULE-0019](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/data-formats/task-schedule.md#REQ-SCHEDULE-0019) defines `interval` as the seconds "between scheduled runs", and [REQ-SCHEDULE-0020](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/data-formats/task-schedule.md#REQ-SCHEDULE-0020) says a repeated run "becomes due according to the interval" and may start later.
  - [REQ-SCHEDULE-0021](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/data-formats/task-schedule.md#REQ-SCHEDULE-0021) draws jitter "independently for each due run" and forbids it to move a run before `starts_at`, after `not_after` or outside a time window.
  - [REQ-SCHEDULE-0007](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/data-formats/task-schedule.md#REQ-SCHEDULE-0007), [REQ-SCHEDULE-0009](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/data-formats/task-schedule.md#REQ-SCHEDULE-0009), [REQ-SCHEDULE-0030](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/data-formats/task-schedule.md#REQ-SCHEDULE-0030), [REQ-SCHEDULE-0032](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/data-formats/task-schedule.md#REQ-SCHEDULE-0032), [REQ-SCHEDULE-0034](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/data-formats/task-schedule.md#REQ-SCHEDULE-0034) and [REQ-SCHEDULE-0035](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/data-formats/task-schedule.md#REQ-SCHEDULE-0035) give the start time, the run limit and the eligibility conjunction, with final authority to delay or skip at the probe.
- **Implementation:** `tv-rain/barometer-go` on the DEV-636 task branch, based on `3707cf0`. `execution.Occurrence` computes the current occurrence ([Run Reservation](execution-models.md#run-reservation)).

### Finding and Interim Handling

The contract does not say what the interval is measured from (the first run's due time, `starts_at`, acceptance, or the actual start of the previous run), whether the first run of a repeated task is jittered, or what happens to occurrences that become due while the host does not call: whether they are run late, one after another, or skipped, and whether a skipped occurrence counts toward `max_runs`. These choices change how many runs a task makes and how close together they can be, which is observable behavior and a traffic-pattern concern.

**Blocked:** nothing; the task owner approved implementing the interim rule while this finding is open. **Interim rule:**

1. The anchor is `starts_at` when present, otherwise the time the task was accepted by this instance's admission of it. When no run has been reserved before reinitialization, a redelivery's acceptance time is used, which can only move occurrences later.
2. Occurrence 0 is due at the anchor without jitter (REQ-SCHEDULE-0009). For `interval`, occurrence *k* ≥ 1 is due at anchor + *k* × `interval` + *j*ₖ, where *j*ₖ is drawn locally from `crypto/rand` in [0, `jitter`] once, when the occurrence first becomes current, and recorded with the run accounting.
3. The current occurrence is the latest *k* whose unjittered time anchor + *k* × `interval` has passed. Once that *k* reaches `max_runs`, no occurrence is current and the schedule has ended. Earlier occurrences that never started are skipped and do not start late. A skipped occurrence consumes no run, so a task makes at most `max_runs` runs and can make fewer.
4. The current occurrence starts only after its jittered due time and while every other eligibility condition holds. A due time at or after `not_after` never starts. A call outside a time window leaves the occurrence current, so it can start when the window opens before the next occurrence becomes current.
5. At most one run is reserved per occurrence. A repeated start for the occurrence that already has a run returns that run's handle while its record is kept.

### Impact, Reviews and Alternatives

- **Safety:** the interim rule never starts more runs than `max_runs`, and never runs missed occurrences in a burst. Two runs can still start close together when a window or constraint delays one occurrence until the next becomes current. Its cost is fewer runs when the host wakes rarely. Safety review of the chosen public rule is required, because catch-up behavior can make measurement traffic more distinctive.
- **Alternatives:** (a) the interim rule; (b) anchor each run to the actual start of the previous run, so no occurrence is ever skipped but the schedule drifts with host wakeups; (c) a fixed grid with catch-up that runs each missed occurrence once, one per call; (d) jitter occurrence 0 as well, delaying the first run by up to `jitter` after `starts_at`.
- **Compatibility:** each alternative is a Draft `0.1.0` clarification with no new field; (b) and (c) produce more runs than (a) for the same host behavior.

### Ordering and Closure

Public record and disposition first, then internal schedule-review alignment if its validation plan names due times, then Go adoption with due-time cases, then bindings. Closure requires the public disposition, its immutable revision if wording changes, and Go adoption evidence or recorded non-applicability.

### Public Record Addendum — October 22, 2026

The sections above keep their wording as historical evidence, including the statement that no public record existed and the reference to the DEV-636 task branch, which merged as [PR #64](https://github.com/tv-rain/barometer-go/pull/64) at `861a353b6da47e122ce06a73f8e8eafc56d6a01b`.

- **Public record:** DEV-639 published [Interval Due Times and Catch-Up](https://github.com/tv-rain/barometer-spec/blob/9804ab34f21863723a94caba0541fd4e64870fce/specification-issues/interval-due-times-catch-up.md) through [public PR #26](https://github.com/tv-rain/barometer-spec/pull/26), merged at `9804ab34f21863723a94caba0541fd4e64870fce`. It is open, lists the alternatives above without selecting one, changes no specification text and flags the safety review of the chosen rule as not completed.
- **Ownership:** after safety core closeout, [DEV-651](https://app.clickup.com/t/124kk1zcujd) owns the public disposition, its safety review and Go adoption.
- **Go state:** unchanged; the interim rule stays in force. The planned `CC-SCHEDULE-SCHEDULING-MODEL-REPEAT-INTERVAL-FIELDS` family, whose delayed due runs depend on this rule, stays `planned` ([Runtime Case Families](execution-models.md#runtime-case-families)).
- **Remaining before closure:** the public disposition, its immutable revision if wording changes, and Go adoption evidence. No runtime, platform or field support is claimed.

## GO-SPEC-ISSUE-012: Start of Unacknowledged Bundle Retention

**Current state:** Open, raised October 22, 2026 under DEV-563 while implementing the pending-result store. No public record exists yet; public maintainers own the disposition. Go implements the conservative interim rule below, which the task owner approved with the DEV-563 plan.

### Identity, Baselines and Requirements

- **Ownership:** [DEV-651](https://app.clickup.com/t/124kk1zcujd) owns the public record, disposition, privacy review and Go adoption after safety core closeout.
- **Reporter:** the Go implementation agent, under the task owner's instruction. **Duplicates:** none.
- **Baseline:** Draft `0.1.0` at [9804ab34f21863723a94caba0541fd4e64870fce](https://github.com/tv-rain/barometer-spec/tree/9804ab34f21863723a94caba0541fd4e64870fce).
  - [REQ-LIMIT-0023](https://github.com/tv-rain/barometer-spec/blob/9804ab34f21863723a94caba0541fd4e64870fce/library/registered-limits.md#REQ-LIMIT-0023) sets "Unacknowledged local bundle retention" to at most 86,400 seconds, and [REQ-LIMIT-0024](https://github.com/tv-rain/barometer-spec/blob/9804ab34f21863723a94caba0541fd4e64870fce/library/registered-limits.md#REQ-LIMIT-0024) deletes expired pending bundles rather than uploading them late.
  - [REQ-BUNDLE-0015](https://github.com/tv-rain/barometer-spec/blob/9804ab34f21863723a94caba0541fd4e64870fce/data-formats/result-bundle.md#REQ-BUNDLE-0015) step 8 deletes the local bundle "after at most 86,400 seconds without acknowledgement", after steps 1 to 5 finish the selected attempts, generate the identifier and persist the bundle.
- **Implementation:** `tv-rain/barometer-go` on the DEV-563 task branch, based on `9253fdd`. `internal/persistence/pending` ([Pending Results and Bundles](package-architecture.md#persistence-ports)).

### Finding and Interim Handling

The contract does not say when the 86,400 seconds start: when an attempt result is completed or first persisted, when the logical bundle holding it is finally persisted with its identifier (step 5), or when the bundle is first offered for delivery. Results wait locally before their bundle is complete, so the later readings let a result stay on the device longer than 86,400 seconds, and the reading changes when data is deleted, which is observable behavior and a privacy concern. The contract also defines no separate retention for results not yet in a bundle.

**Blocked:** nothing; the interim rule is the most conservative reading. **Interim rule:**

1. Every pending result is persisted inside a bundle record from its first write; there is no separate unbundled result state.
2. A bundle's retention starts when its first result is persisted, on the engine clock's wall reading, and is stored with the bundle so it survives a restart. Adding a result to the open bundle, sealing it, returning it from `NextBundle` or reinitializing does not move it.
3. A bundle is deleted, and never returned, once 86,400 seconds have passed since that start: at the start time plus 86,400 seconds exactly, on every pending-store access and when initialization opens the store. A failed deletion fails the access rather than returning the bundle.

Every result is therefore retained for at most 86,400 seconds after it was persisted, which follows its completion only by the time of the write. A wall clock moved backwards can delay expiry; the same limitation applies to replay retention.

### Impact, Reviews and Alternatives

- **Privacy:** the interim rule never retains a result longer than any other reading allows. Its cost is that results added late to a bundle are deleted sooner than 86,400 seconds after their own completion. A privacy review of the chosen public rule is required, because the retention start decides how long measurement results stay on the device.
- **Alternatives:** (a) the interim rule; (b) start at the final persistence of the complete logical bundle (step 5), so a result can wait up to the time its bundle stays open plus 86,400 seconds; (c) start when the bundle is first offered to the host; (d) start per result at its completion, which needs per-result expiry inside a bundle and contradicts immutable bundle bytes.
- **Compatibility:** each alternative is a Draft `0.1.0` clarification with no new field.

### Ordering and Closure

Public record and disposition first, then internal alignment if a retention validation plan names the start point, then Go adoption with retention cases, then bindings. Closure requires the public disposition, its immutable revision if wording changes, and Go adoption evidence or recorded non-applicability.

## GO-SPEC-ISSUE-013: Library Internal Check for `require_validated`

**Current state:** Open, raised October 23, 2026 under DEV-644 while implementing network-condition evaluation. The public record [Library Internal Check for `require_validated`](https://github.com/tv-rain/barometer-spec/blob/32062c4bafcc29f5050ca6b78423f2acaefdde56/specification-issues/require-validated-internal-check.md) is open with no disposition; public maintainers own the disposition, and [DEV-651](https://app.clickup.com/t/124kk1zcujd) owns its follow-up. Go implements the conservative interim rule below, which the task owner approved with the DEV-644 plan. See the [public record addendum](#public-record-addendum--october-23-2026).

### Identity, Baselines and Requirements

- **Reporter:** the Go implementation agent, under the task owner's instruction. **Duplicates:** none.
- **Baseline:** Draft `0.1.0` at [5d01879cfd1d47396c780befa2567f7e90439948](https://github.com/tv-rain/barometer-spec/tree/5d01879cfd1d47396c780befa2567f7e90439948).
  - [REQ-EXECUTION-0016](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/data-formats/task-execution.md#REQ-EXECUTION-0016) describes `require_validated` as "Require confirmed Internet connectivity, system method will be used to check when available, lib internal check will be called otherwise."
  - [REQ-EXECUTION-0031](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/data-formats/task-execution.md#REQ-EXECUTION-0031) makes a network condition that cannot be determined, including validation required by `require_validated`, unsatisfied: the task does not run, Barometer does not guess, and a recorded attempt is `not_run` with `constraints_unsatisfied`.
  - The access object's `validated` field is the "coarse platform Internet-validation state" ([REQ-CONTEXT-FORMAT-0004](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0004)).
- **Implementation:** `tv-rain/barometer-go` on the DEV-644 task branch, based on `0869b97`. `execution.NetworkConditions` ([Network Conditions and Access Context](execution-models.md#network-conditions-and-access-context)).

### Finding and Interim Handling

REQ-EXECUTION-0016 says the library runs its own connectivity check when the platform offers no validation method, while REQ-EXECUTION-0031 says an undeterminable validation state is unsatisfied. The contract defines no such check: no endpoint, protocol, traffic bound, timing, privacy property, result field or relationship to the attempt's limits. A library-originated probe sent before a measurement would be network traffic outside the attempt's registered test and could make Barometer's behavior more distinctive on a monitored network.

**Blocked:** the row of `CC-EXECUTION-FIELDS-NETWORK-FIELDS` that uses "the fake library internal check". **Interim rule:**

1. `require_validated` is satisfied only by the platform's own `validated: yes`.
2. `no`, `unknown`, a capability error and an unreadable access context are unsatisfied, as REQ-EXECUTION-0031 requires: `StartRun` returns `run_not_due`, and before the first attempt of a reserved run the attempt is `not_run` with `constraints_unsatisfied`.
3. No library connectivity check is implemented, and no traffic is sent to decide validation.

### Impact, Reviews and Alternatives

- **Safety and privacy:** the interim rule sends no traffic outside the registered test. Its cost is that a task requiring validation never runs on a platform that reports no validation state. Any library check needs a safety and privacy review before implementation; this was accepted as a limitation of the DEV-578 sub-epic.
- **Alternatives:** (a) remove the internal-check clause from REQ-EXECUTION-0016 and keep REQ-EXECUTION-0031 as the only rule; (b) specify a bounded, reviewed library check, with its endpoint, traffic and attempt-limit accounting, and its result representation; (c) let local policy enable a reviewed check per deployment.
- **Compatibility:** (a) is a Draft `0.1.0` clarification with no new field; (b) and (c) add behavior and need a version or compatibility decision.

### Ordering and Closure

Public record and disposition first, then internal alignment of the execution validation plan, then Go adoption with the row's case, then bindings. Closure requires the public disposition, its immutable revision if wording changes, and Go adoption evidence or recorded non-applicability.

### Public Record Addendum — October 23, 2026

The sections above keep their wording as historical evidence, including the statement that no public record existed and the reference to the DEV-644 task branch, which merged as [PR #70](https://github.com/tv-rain/barometer-go/pull/70) at `c4b3761085c8bc9703db32baf7c3327d9697d18b`.

- **Public record:** DEV-649 published [Library Internal Check for `require_validated`](https://github.com/tv-rain/barometer-spec/blob/32062c4bafcc29f5050ca6b78423f2acaefdde56/specification-issues/require-validated-internal-check.md) through [public PR #27](https://github.com/tv-rain/barometer-spec/pull/27), merged at `32062c4bafcc29f5050ca6b78423f2acaefdde56`. It is open, lists the alternatives above without selecting one and changes no specification text.
- **Ownership:** after safety core closeout, [DEV-651](https://app.clickup.com/t/124kk1zcujd) owns the disposition, the safety and privacy review that any library check needs, and Go adoption.
- **Go state:** unchanged; the interim rule stays in force. The planned `CC-EXECUTION-FIELDS-NETWORK-FIELDS` family, whose library-internal-check row depends on this finding, stays `planned` ([Runtime Case Families](execution-models.md#runtime-case-families)).

## GO-SPEC-ISSUE-014: Equal Start and End in a Time-Window Range

**Current state:** Open local finding, recorded under DEV-652 from the runtime closeout limitation. No public record or disposition exists. Raising this finding publicly is a separate maintainer decision; this record changes no behavior.

### Identity, Baselines and Requirements

- **Reporter:** the Go implementation agent, following the safety core closeout correction task. **Duplicates:** none identified in this Go findings record. Public maintainers own any contract disposition.
- **Public baseline:** Draft `0.1.0` at `5d01879cfd1d47396c780befa2567f7e90439948`, [Task Schedule — Time Window](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/data-formats/task-schedule.md#time-window): REQ-SCHEDULE-0013 requires unique, non-overlapping ranges; REQ-SCHEDULE-0015 defines an inclusive start and exclusive end; REQ-SCHEDULE-0016 requires `HH:mm`; REQ-SCHEDULE-0017 defines crossing midnight when the end is earlier than the start. None explicitly defines equality.
- **Go baseline:** `de59ede07160658f61049cb4e90cdff9dd1da9d3`. `internal/validation/task.go` (`scheduleErr`) accepts an otherwise valid range with equal endpoints, and `internal/validation/schedule.go` (`startEnvelope`) matches only strictly increasing or strictly decreasing endpoints. Source inspection confirms the [runtime assessment limitation](https://github.com/tv-rain/barometer/blob/main/docs/runtime-assessment.md#limitations-carried-forward); no new conformance evidence is claimed.

### Finding and Current Handling

A synthetic range `{"start":"12:00","end":"12:00"}` passes schedule validation but contributes no eligible minute. A window consisting only of that range never permits a run or retry; another nonempty range can still match. Equality is neither rejected nor treated as a full day. This is the existing Go behavior, not a public clarification or a newly selected interim rule.

### Impact, Review and Ordering

The current behavior adds no measurement traffic, but an issuer expecting a full-day range gets no run. Treating equality as a full day would expand eligible traffic; rejection would change admission compatibility. Alternatives for public review are an explicitly empty range, rejection, or an explicitly full-day range. Any change needs public disposition and safety review before Go adoption, with binding and producer compatibility considered. Keep behavior unchanged pending that separate decision; no public issue is raised by this task and no finding is closed.

## GO-SPEC-ISSUE-015: Resolved-Address Validation Behind a Hostname CONNECT Proxy

**Current state:** Public disposition and internal alignment merged; scoped review confirmed complete by the project owner on October 26, 2026 with instruction to proceed in Go. The shared Go gate and synthetic cases are adopted under [DEV-659](https://app.clickup.com/t/124kk1zcw4j), merged at `9ad1166`, and the TCP-01 and DNS-01 runners use it; binding and host integration remain pending. The finding is not closed and no platform, field or release support is claimed.

### Finding and Historical Evidence

At public Draft `0.1.0` revision `32062c4bafcc29f5050ca6b78423f2acaefdde56`, REQ-TRUST-0016 required checks before contact while the then-active REQ-PROXY-0019 and REQ-PROXY-0026 permitted hostname CONNECT with resolution behind the proxy. Go revision `3c9a3ab9d5e2a3fa30d987b8398d3dde48d4469c` supplied no proxy-side destination addresses. Two proxy scripts could produce identical successful responses while selecting different allowed/prohibited addresses. The original example was reasoning, not executed conformance evidence or a production exposure claim.

The finding was published through [Go PR #78](https://github.com/tv-rain/barometer-go/pull/78), and the original public proposal through [public PR #28](https://github.com/tv-rain/barometer-spec/pull/28). Publication alone selected no rule. Local DNS would not establish the proxy's answer and could disclose target names; checking after CONNECT would be too late.

### Accepted Authority and Ordering

- **Coordination:** decision 0013, [PR #32](https://github.com/tv-rain/barometer/pull/32), merged at [`61691f49773c68b660e84939555070c8ee1990ff`](https://github.com/tv-rain/barometer/blob/61691f49773c68b660e84939555070c8ee1990ff/docs/decisions/0013-explicit-ip-only-proxy-connect.md).
- **Public:** [explicit-IP-only disposition](https://github.com/tv-rain/barometer-spec/blob/bb805f07a36630db7250d5ab63006ed5e351f7a7/specification-issues/proxy-resolved-address-validation.md), [PR #29](https://github.com/tv-rain/barometer-spec/pull/29), merged after coordination at [`bb805f07a36630db7250d5ab63006ed5e351f7a7`](https://github.com/tv-rain/barometer-spec/tree/bb805f07a36630db7250d5ab63006ed5e351f7a7), retaining Draft `0.1.0` and explicitly narrowing compatibility.
- **Internal:** `barometer-spec-internal` PR #13, merged at `fc5fdf8733ad990f8363dcf4748edafa40b08dde`, aligned `proxy-target-validation-review.md` and the affected companions. The owner subsequently stated that the required review is done and instructed Go implementation; this confirmation is recorded on DEV-659. It is not represented as independent interpretation review or an invented set of review findings.
- **Adoption order:** coordination, public contract, internal alignment and scoped review precede this Go change. Binding/host and producer compatibility assessment follows Go evidence before affected integration/release.

REQ-PROXY-0043, REQ-PROXY-0044, REQ-PROXY-0045 and REQ-PROXY-0046 replace REQ-PROXY-0019, -0026, -0023 and -0042 respectively. REQ-CONTEXT-0043 replaces REQ-CONTEXT-0020. REQ-HTTP-01-0051 defines hostname proxy redirect rejection. REQ-TRUST-0016 and REQ-PROXY-0037 retain their pre-contact safety obligations; current Go references use successors while prior milestone manifests remain pinned to their original revisions.

### Go Adoption and Evidence Boundary

The [execution target gate](execution-models.md#execution-time-target-validation) binds admitted registrations to the attempt and checks exact initial authority/transport, expiry, path and local address policy before each contact. Direct hostname resolution is checked and pinned to a literal IP, with unsafe mixed sets refused. The proxy path refuses hostname authorities before DNS or proxy contact. Typed redirect dialers check complete tuples and authorized parent links while sharing the attempt's budgets, cancellation and cleanup. Existing coarse result/admission states are retained; local refusal is not a protocol failure or censorship verdict. Proxy context endpoint eligibility is checked before bounded primary/fallback acquisition, and missing required context prevents the test from starting.

Synthetic evidence locations are `internal/execution/targets_test.go`, `internal/execution/target_conformance_test.go`, `internal/execution/egress_test.go`, `internal/networkcontext/egress_test.go`, and the admission/registry tests. They exercise the production shared gate with scripted resolution and transports and approved explicit IP authorities routed by a synthetic proxy to loopback; no external target is contacted. Legacy budget/lifecycle fixtures use an explicit test authorizer so they do not claim target-validation evidence. No production loopback exemption or public policy override is introduced by those fixtures.

The shared target-validation portion of `CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION` has executable synthetic gate cases; the whole mixed family remains planned. Production HTTP-01/NET-05 codecs and runners, their use of the shared redirect boundary and protocol-specific observation construction remain with their measurement tasks. Native/binding/host adoption, independent interpretation review, field validation and release claims remain pending. A malicious proxy may still disobey an explicit-IP request under the existing compromised-component threat limitation. A future hostname enforcement capability requires a separate reviewed public contract.

Final validation results, approved Go commit/branch links and remaining work are recorded once on DEV-659. Closure requires applicable evidence and downstream adoption or explicit non-applicability; this finding is not closed merely because the contract and shared gate exist.

## GO-SPEC-ISSUE-016: Unnamed Classes and Order in Step-8 Input and Options

**Current state:** Open, raised October 26, 2026 under [DEV-663](https://app.clickup.com/t/124kk1zcw4r) from the maintainer safety review of the GO-SPEC-ISSUE-010 mapping and from its Go adoption. No public record exists yet; public maintainers own disposition. Go behavior is unchanged.

### Identity, Baselines and Requirements

- **Reporter:** the Go implementation agent, under the task owner's instruction. **Duplicates:** none. [GO-SPEC-ISSUE-010](#go-spec-issue-010-admission-class-for-step-8-input-and-options-failures) mapped invalid input and options and stricter test byte limits; this finding covers what that mapping leaves unnamed.
- **Baseline:** Draft `0.1.0` at [`12062471875bbfbc55eafc94dc5a7ea4cd07ca5e`](https://github.com/tv-rain/barometer-spec/tree/12062471875bbfbc55eafc94dc5a7ea4cd07ca5e): [REQ-TRUST-0056](https://github.com/tv-rain/barometer-spec/blob/12062471875bbfbc55eafc94dc5a7ea4cd07ca5e/library/task-trust.md#REQ-TRUST-0056), [REQ-OPTIONS-0033](https://github.com/tv-rain/barometer-spec/blob/12062471875bbfbc55eafc94dc5a7ea4cd07ca5e/data-formats/task-options.md#REQ-OPTIONS-0033), [REQ-TRUST-0047](https://github.com/tv-rain/barometer-spec/blob/12062471875bbfbc55eafc94dc5a7ea4cd07ca5e/library/task-trust.md#REQ-TRUST-0047), [REQ-TRUST-0036](https://github.com/tv-rain/barometer-spec/blob/12062471875bbfbc55eafc94dc5a7ea4cd07ca5e/library/task-trust.md#REQ-TRUST-0036) and [REQ-LIMIT-0002](https://github.com/tv-rain/barometer-spec/blob/12062471875bbfbc55eafc94dc5a7ea4cd07ca5e/library/registered-limits.md#REQ-LIMIT-0002).
- **Implementation:** `Admitter.Gate`, `taskdecode.Envelope.Parameters` and `jsondecode.Validate` on this branch. TCP-01 and DNS-01 production codecs use the same staged contract as the synthetic test-scoped codecs.

### Finding and Interim Handling

1. **Local-policy option restrictions.** REQ-OPTIONS-0033 sends options local policy does not permit to the class REQ-TRUST-0047 lists for that condition, and the public options document lets a probe enforce stricter local safety restrictions. REQ-TRUST-0047 lists local-policy refusal only for a direct or proxy path at step 8 and for the risk bound at step 9, so a stricter local restriction on another option value has no named class. Go has no such local option restriction yet. The maintainer review suggested `task_issuer_not_authorized`, as for the step-8 local-policy path row.
2. **Registered numeric limits on input or option values.** REQ-TRUST-0056 names `task_limit_exceeded` only for byte limits, so a value above a registered numeric limit, such as a DNS-01 `additional_response_window_ms` above the REQ-LIMIT-0014 maximum, reads as a value the schema does not permit, `task_schema_invalid`. REQ-LIMIT-0002 requires rejection without naming a class. **Interim handling:** Go returns `task_limit_exceeded` for a limit-flagged bound, consistent with the class name and with step 6 for execution limits, as the project owner approved under DEV-663; `TestGateInputOptionsClasses` records it.
3. **Order among several failing values.** REQ-TRUST-0056 orders a byte limit before the content it bounds and REQ-TRUST-0039 orders input and options before proxy compatibility, but nothing orders input against options or one member against another. A task with invalid input and over-limit options could receive either class from different implementations. **Interim handling:** Go validates input before options and members in their input order, so the first failing value decides; `TestGateInputOptionsClasses` records invalid input before over-limit options.
4. **Capability gaps reported by a codec.** A codec normalization failure is always `task_schema_invalid`, so a production codec cannot report a valid option that needs an unprovided capability as `task_unsupported` (REQ-TRUST-0036). This is a Go codec-contract gap, not a public ambiguity. The codec pattern delivered by DEV-660 does not add such a path, and neither TCP-01 nor DNS-01 needs one; it stays open for the first codec that does, under the Web reachability and composite sub-epic ([DEV-657](https://app.clickup.com/t/124kk1zcw4f)).

**Blocked:** nothing in current approved work; exact-class conformance assertions for items 1 to 3 stay provisional, and the DNS-01 codec (DEV-662) should not assert item 2 as conformant before disposition. **Unaffected:** the rows REQ-TRUST-0056 names, proxy compatibility, path authority, target registration and composite rows.

### Impact and Ordering

- **Safety:** every case is rejected at admission before network activity; only the class differs. Hosts and issuers that branch on classes could diverge between implementations.
- **Alternatives:** for item 1, `task_issuer_not_authorized`, `task_schema_invalid` or `task_unsupported`; for item 2, `task_limit_exceeded` or `task_schema_invalid`; for item 3, a public order (input before options and member order, or limit failures before content failures across the step) or explicit implementation choice.
- **Ordering:** public record and disposition first, then internal task-trust review alignment, then Go adoption with exact-class cases, then bindings. Item 4 follows independently, with the first codec that needs it (DEV-657).

## GO-SPEC-ISSUE-017: HTTP-01 Terminal Response Outcome Precedence

**Current state:** The owner selected the proposed order; [public PR #33](https://github.com/tv-rain/barometer-spec/pull/33) merged REQ-HTTP-01-0052 at `e6257ec64564ad415a2d6c1f8d106fa1763d7de8`, retaining Draft `0.1.0`. [Internal PR #17](https://github.com/tv-rain/barometer-spec-internal/pull/17) merged validation alignment at `986a207176a9190d1c4c354fd69df55cfab23616`; baseline reconciliation and scoped implementation-author review are prepared in the internal companion. Go HTTP-01 adoption and synthetic cases are prepared under [DEV-667](https://app.clickup.com/t/124kk1zcw4x), pending final approval and merge. The finding remains open for durable adoption/compatibility evidence; broader reviews and platform/field/release support are not claimed.

### Identity, Baselines and Finding

The Go implementation agent reports this finding; public maintainers own disposition. No duplicate was identified in the Go or public findings. Public baseline: Draft `0.1.0` at [`2b843611ddc6fe3718359034574cd2989b3cbf9d`](https://github.com/tv-rain/barometer-spec/tree/2b843611ddc6fe3718359034574cd2989b3cbf9d), HTTP-01 and Measurement Result versions `0.1.0`. REQ-HTTP-01-0007, -0008, -0009, -0036 and -0037 define procedure, checks and outcomes without explicitly ordering overlapping terminal-response failures. REQ-HTTP-01-0038 and the GO-SPEC-ISSUE-004 common termination rules remain separate and authoritative. REQ-OBS-0023, -0028, -0030 and -0031 constrain failure vocabulary, body metadata and minimized check results.

At Go [`8576bbc10d68c4a5c2328df34043d31124b06bad`](https://github.com/tv-rain/barometer-go/tree/8576bbc10d68c4a5c2328df34043d31124b06bad), `internal/protocol/http/checks.go` (`evaluate`) and `exchange.go` (`Exchange`) select unsupported encoding or body-read errors over an earlier failed status check; otherwise status precedes header, header precedes body and a decidable mismatch precedes undecidable checks at the body bound. The public instruction to stop at the first failed stage does not explicitly order these conditions within the HTTP stage. The HTTP-01 normalization case plan prohibits undocumented precedence.

A synthetic expected status 200 plus body `contains: ok`, with response status 503 and `Content-Encoding: gzip`, follows the current source path to `unsupported_content_encoding`, retaining a failed status check and no body check. A 503 response whose declared body ends prematurely instead selects `protocol_error`; a truncated body with an undecidable suffix retains `unexpected_status`. These are source-inspected examples, not newly executed reproductions or conformance evidence. The public proposal supplies alternatives, proposed wording and expected combined-failure cases.

### Impact, Interim Handling and Ordering

Different summaries could affect analysis and cross-implementation comparison even with identical indexed checks. No new traffic, retained content, target, retry, identifier, resource allowance or API is proposed. Ordinary server/content/encoding/transfer failures remain alternatives to censorship. The owner selected the order, and scoped implementation-author technical/safety/privacy review records unchanged traffic, bounds, minimization and common termination in the reconciled internal companion. Independent and whole-deployment reviews remain separate; native, binding, host and release support are not established.

**Adoption:** the public ambiguity is resolved by REQ-HTTP-01-0052. The production HTTP-01 runner reuses the bounded primitive's selected order, preserves decidable checks and fixes its terminal marker before cleanup. Synthetic overlapping-failure cases and common interruption boundaries are listed in [Current HTTP-01 Execution](measurement-requirements.md#current-http-01-execution). No new fallback or censorship classification is introduced.

Required order remains public contract, internal alignment, Go adoption, then applicable binding/host validation. Public and semantic internal alignment merges are recorded above; the internal immutable-baseline reconciliation and Go implementation await final commit/push approval. Closure needs durable adoption evidence and applicable downstream validation or explicit non-applicability. A prepared or passing synthetic suite alone is not closure.

## Signed Remote Task Package Adoption

**Status:** Public contract and internal review alignment are merged; the Go core adopts them in DEV-612 with synthetic evidence only. The cryptographic, safety and host integration reviews were performed by the project maintainer in DEV-615 with no findings, and the remote-delivery gate (DEV-595) is lifted. The interpretation question below is closed by REQ-TRUST-0055, adopted by DEV-624. Binding and host adoption remain open (DEV-455).

This is an adoption record, not a Go finding: the change came from coordination decision 0011 (DEV-609), not from an implementation ambiguity. Its public record is [Library-Verified Signed Remote Task Packages](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/specification-issues/signed-remote-task-packages.md).

### Baselines and Requirements

- **Public baseline:** Draft `0.1.0` at immutable revision `e97502747c917a16bc2b71337535ff972bbd7d64` (public PR #18 merge, DEV-610).
- **Internal alignment:** `barometer-spec-internal` task-trust review at the PR #6 merge (DEV-611), which adds the signed-package rows of the validation plan.
- **Adopted requirements:** REQ-TRUST-0040 to REQ-TRUST-0054 and REQ-LIMIT-0032. The retired REQ-TRUST-0001, REQ-TRUST-0002, REQ-TRUST-0004 to REQ-TRUST-0007, REQ-TRUST-0009, REQ-TRUST-0017 and REQ-TRUST-0037 are replaced by their declared successors in `requirement-ref` comments; Markdown records cite them only as `requirement-history`.

### Adoption Evidence

- **Implementation:** `internal/admission/signed.go` (package decoding, message construction, the nine ordered checks), `internal/admission/delivery.go` (package-only, single-use remote records), `internal/model/task/trust.go` (no host-supplied remote issuer, issuance time or authentication flag) and `internal/model/task/keys.go` (trusted key records and public-key validation), described in [Signed Remote Task Packages](admission.md#signed-remote-task-packages) and [Trusted Issuer Keys](trusted-policy-models.md#trusted-issuer-keys). The `RemoteVerifier` seam and the synthetic HMAC verifier are removed; tests use synthetic Ed25519 keys from `internal/testkit/delivery`.
- **Cases:** `CC-TRUST-SIGNED-REMOTE-PACKAGES`, `CC-TRUST-SIGNED-PACKAGE-DISCLOSURE`, `CC-TRUST-SIGNED-PACKAGE-RETENTION`, `CC-TRUST-KEY-ROTATION`, the rewritten `CC-TRUST-ACCEPTED-TASK-ORIGINS`, unit-test cases for key validation, package formats and Ed25519 edge vectors, and the fuzz targets `FuzzSignedPackage` and `FuzzAdmitTask`. Results are recorded in the traceability evidence for this change.

### Interpretation Question

The paragraph below keeps its DEV-612 wording as historical evidence.

REQ-TRUST-0054 states that a task admitted while its signing key was trusted and valid is not re-verified when that key's validity later ends. The existing replacement procedure (DEV-593) invalidates every task handle and relies on redelivery for tasks the new policy still permits, and every redelivered package is verified at its new acceptance time. A task whose key validity ended between admission and a policy replacement is therefore rejected on redelivery, and an exact redelivery within one generation after key expiry is rejected while the existing handle stays valid. The implementation keeps this fail-closed behavior. Whether REQ-TRUST-0054 intends such a task to survive replacement without re-verification is left to the public specification; it is recorded here and not resolved by local choice.

**Closed by reference, October 21, 2026 (DEV-624).** The public maintainer accepted the fail-closed reading in [Re-verification of Redelivered Remote Tasks After Key Expiry](https://github.com/tv-rain/barometer-spec/blob/4743b636f29feb1f0f8d042866d04ccda3532938/specification-issues/key-expiry-redelivery.md), merged with public PR #19 at `4743b636f29feb1f0f8d042866d04ccda3532938` as [REQ-TRUST-0055](https://github.com/tv-rain/barometer-spec/blob/4743b636f29feb1f0f8d042866d04ccda3532938/library/task-trust.md#REQ-TRUST-0055): the REQ-TRUST-0054 exemption covers only an acceptance the instance still holds, and every presentation of a package is verified at its own acceptance time before replay handling and rejected with `task_source_unauthenticated` when its key is not trusted and valid, without invalidating a held acceptance. The behavior described above conforms and is unchanged. DEV-624 adds the `requirement-ref` on `Admitter.verifyPackage`, cases in `TestCcTrustKeyRotationByPolicyReplacement` for a redelivery after a replacement that keeps an expired key (rejected, no replay write) and for the same task bytes re-signed with a valid key (the retained acceptance with its original expiry), and maps REQ-TRUST-0055 to `CC-TRUST-KEY-ROTATION` in traceability ([Signed Remote Task Packages](admission.md#signed-remote-task-packages)).

### Remaining Boundaries

The DEV-612 list of required reviews is superseded by the maintainer decision of October 21, 2026:

- The cryptographic review of the package encoding, signed-message construction, Ed25519 verification strictness and public-key validation, the safety review of residual key exposure, the 365-day maximum key validity and four keys per issuer, and the host integration review of package retention were performed by the project maintainer under DEV-615 (`barometer-spec-internal` PR #7, merge `4522637fc91890acae9de6b4da9bd189d9f3318e`) and passed with no findings. The remote-delivery gate (DEV-595) is lifted.
- Independent external reviews of the same scopes (DEV-618 to DEV-621) run asynchronously and do not gate; their findings return as ordinary findings under the specification issue workflow.
- No native, Flutter or host binding of the package-only remote path exists yet (DEV-455), so hosts cannot yet pass packages to the Go core, and how hosts avoid retaining received packages is not yet exercised in binding or host code.
- No runtime, platform, field or release support is claimed.
