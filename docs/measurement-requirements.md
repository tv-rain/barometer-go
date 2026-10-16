# Measurement Requirements and Case Plan

## Status, Authority, and Baselines

This DEV-473 planning map covers 393 public requirement units across all nine Draft `0.1.0` measurements, with 108 intended case families. All owners and cases are **planned**. No runtime implementation, executed conformance, review-completeness evidence, platform/field support or release is claimed.

The immutable public baseline is `59d49364d85e48cc8e69c762d130a7a86dc1fe65`, the verified merge of [public PR #7](https://github.com/tv-rain/barometer-spec/pull/7). Its tree matches approved allocation head `9f0f457014881669cf58af5fad9735c423151fcc` exactly. The Go baseline is `4aafb0e94d5e13411d1f76521ece989a03aee617`. It precedes this planning metadata because every location/case is planned and no executed evidence is claimed. The [measurement manifest](../traceability/measurements.json) and [deterministic metadata report](../traceability/reports/measurements.json) use those explicit baselines. The report retains 393 missing implementation claims, 393 missing conformance records, 565 shared IDs outside this manifest's scope, and residual inventory/allocation gaps.

DEV-568 deliberately keeps this pin while the [cross-cutting plan](cross-cutting-requirements.md) moves to public baseline `a944ed6c389fec9f0d089f0044ead6fe845328b1`. Between `59d4936` and `a944ed6`, no registered test-specific catalog record is added, retired or changed. The only change to a registered test source is a two-sentence cross-reference after the REQ-TLS-01-0016 options table. It defers the offered ALPN identifiers and the negotiated-output and peer-alert normalization to the shared observation vocabularies. Those obligations are REQ-OBS-0049 and REQ-OBS-0050, mapped in the cross-cutting manifest, including the offered-identifier family `CC-OBS-VOCAB-TLS-OFFERED-ALPN`. They do not change the ID, owner or `M-TLS-01-OPTIONS` case of REQ-TLS-01-0016. Because the two manifests have different baselines, they must not be combined into one current conformance report. The next measurement-scope source change requires a reviewed rebaseline of this manifest.

The [public allocation guide](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/requirements/measurement-allocation.md) records unit boundaries and informative dispositions. The public text owns observable behavior; this document assigns Go work and plans validation. Existing [cross-cutting mappings](cross-cutting-requirements.md) and reports retain their original selected baselines. All shared data-format/library source files and existing catalog entries are unchanged between their public PR #6 baseline and this PR #7 merge. This establishes source continuity, not newly executed conformance or support.

Public prerequisite ordering is satisfied. No internal review change is required for this annotation-only allocation: no public behavior, version, runtime field or private normative obligation changes. Native/host/analyst review and validation remain separately scoped future evidence.

## Ownership and Validation Boundaries

Each family must exercise **every linked unit**, every schema/table row, each conditional relation, numeric/string/count boundary and rejection path. A family is a parameterized suite, not one representative test. Test plans describe synthetic future cases; names/paths do not create packages or executable evidence. All numeric limits and defaults remain controlled by the linked public source and any stricter shared/task/local policy.

Admission and option families plan strict test-specific schemas and pure validation in `internal/validation` plus runner-side guard checks. Source/target authority remains in `internal/admission`; it cannot be accepted from serialized task input. Execution/normalization/limit families plan measurement runners, shared protocol owners and `internal/execution`. Versions belong to `internal/registry` and result schema selection. Output/minimization belong to the runner and `internal/result` using `internal/model/observation` and `internal/model/result`. These are intended owners under [Package Architecture](package-architecture.md), not claims that those packages or declarations exist.

Portable fake-platform tests are necessary for NET-01/NET-03/network generation but cannot establish Android or Apple API, permissions, trust-store, suspension, routing or cleanup behavior. Preconditions/safety require separately recorded native and host review/validation. Interpretation families are planned result/host/analyst review contracts: a Go test cannot establish correctness of an external analyst's interpretation. Local synthetic canaries, deterministic clocks and recording transports are the default; no participant data or uncontrolled measurement is needed.

Unresolved public contract ambiguities must enter [Specification Issues](specification-issues.md) and block affected implementation/claims. In particular, outcome-table testing must not invent precedence for combinations the public contract leaves unclear. GO-SPEC-ISSUE-001 remains open for downstream consumer/model, admission, binding and producer evidence; mapping these measurements does not close it.

## Shared Authority and Reuse

The [public reuse matrix](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/requirements/measurement-allocation.md#reused-contracts) links the controlling shared contracts. Reuse the existing cross-cutting case families for strict JSON, task/execution/result envelopes, protocol observations, network context, registered limits, trust and proxy/path behavior, checking their applicability at the selected new baseline. Requirement IDs retain source-local identities; one shared case is not independent evidence for every caller without complete caller-specific coverage review.

NET-05 dispatches exact-version DNS-01/TCP-01/TLS-01/HTTP-01 schemas and runners; its profile/sequence/aggregate/partial-output obligations remain separate. DNS-03/DNS-04/NET-03 share protocol machinery and observation formats, not the standalone TLS-01/HTTP-01 input or operation schema. NET-03 permits its registered HTTP validation behavior, DNS-03 uses POST, and HTTP-01 uses GET. NET-01 remains zero-traffic, without egress. These boundaries are part of each consumer's future integration cases.

Coverage remains partial in the machine-readable manifest. Shared IDs continue to be covered by the separately scoped cross-cutting mapping, while threat-model/measurement-limitations allocation, residual inventory review, actual implementation/conformance and native/host/analyst evidence remain explicit gaps. An exit-zero metadata report cannot close those gaps or promote Draft status.

The following links select reusable case families and their public requirement IDs, without adding those shared IDs to this separately scoped manifest. They are integration prerequisites, not executed evidence. Each consumer still exercises its own stricter registered behavior and every applicable shared unit.

| Consumer | Reusable lower-layer mapping |
|---|---|
| All nine measurements | [Strict JSON framing](cross-cutting-requirements.md#cc-json-unit-0001) through [sanitized decoding failures](cross-cutting-requirements.md#cc-json-unit-0011), [input validation](cross-cutting-requirements.md#cc-input-validation), [option precedence](cross-cutting-requirements.md#cc-options-execution-precedence), [admission procedure](cross-cutting-requirements.md#cc-trust-admission-procedure), [target registration](cross-cutting-requirements.md#cc-trust-target-registration), [execution model](cross-cutting-requirements.md#cc-execution-execution-model), [effective risk](cross-cutting-requirements.md#cc-safety-effective-risk), [attempt limits](cross-cutting-requirements.md#cc-limit-attempt-limits), [result test/version](cross-cutting-requirements.md#cc-result-test-object), and [result privacy/interpretation](cross-cutting-requirements.md#cc-result-privacy-safety-and-interpretation). Remote-path behavior applies only where the registered test generates traffic. |
| NET-01 | [Access schema](cross-cutting-requirements.md#cc-context-format-access-object), [platform collection](cross-cutting-requirements.md#cc-context-platform-collection), and [context use in results](cross-cutting-requirements.md#cc-context-format-use-in-results), narrowed by NET-01's zero-traffic/no-egress and top-level-context omission. |
| DNS-01 | [DNS failure mapping](cross-cutting-requirements.md#cc-obs-dns-query-failures), [query details](cross-cutting-requirements.md#cc-obs-dns-query-details), [responses](cross-cutting-requirements.md#cc-obs-dns-query-dns-response), [answers](cross-cutting-requirements.md#cc-obs-dns-query-dns-answer), and [DNS bounds](cross-cutting-requirements.md#cc-limit-dns-limits). |
| TCP-01 | [TCP failures](cross-cutting-requirements.md#cc-obs-tcp-connect-failures), [details](cross-cutting-requirements.md#cc-obs-tcp-connect-details), and [path consistency](cross-cutting-requirements.md#cc-context-path-consistency). |
| TLS-01 | Shared TCP families plus [TLS failures](cross-cutting-requirements.md#cc-obs-tls-handshake-failures) and [details](cross-cutting-requirements.md#cc-obs-tls-handshake-details); explicit-IP/SNI/ALPN measurement constraints remain local. |
| HTTP-01 | Shared DNS/TCP/TLS families plus [HTTP failures](cross-cutting-requirements.md#cc-obs-http-exchange-failures), [details](cross-cutting-requirements.md#cc-obs-http-exchange-details), [redirect](cross-cutting-requirements.md#cc-obs-http-exchange-redirect), [body](cross-cutting-requirements.md#cc-obs-http-exchange-body), [checks](cross-cutting-requirements.md#cc-obs-http-exchange-check), and [HTTP bounds](cross-cutting-requirements.md#cc-limit-http-and-validation-limits). |
| DNS-03 / DNS-04 | Shared DNS/TCP/TLS families; DNS-03 additionally consumes HTTP observation/bound families. Bootstrap, resolver authentication and POST/length-prefix exchange cases remain consumer-specific. |
| NET-03 | Shared access/platform and DNS/TCP/TLS/HTTP observation families, including HTTP/validation bounds. [Protocol output cases](#m-net-03-output) retain NET-03's distinct platform/active evidence and redirect prohibition. |
| NET-05 | Reuse the exact-version [DNS-01 admission](#m-dns-01-admission), [TCP-01 admission](#m-tcp-01-admission), [TLS-01 admission](#m-tls-01-admission), [HTTP-01 admission](#m-http-01-admission), and each primitive's option, execution, normalization and output families below; [context acquisition](cross-cutting-requirements.md#cc-context-acquisition-and-fallback), [composite limits](cross-cutting-requirements.md#cc-limit-net-05-composite-limits), and [partial results](cross-cutting-requirements.md#cc-result-partial-results) remain additional dependencies. [NET-05 execution](#m-net-05-execution) and [output](#m-net-05-output) check caller-specific composition. |

## Source Summary

| Test | Allocated units | Planned families |
|---|---:|---:|
| NET-01 | 38 | 12 |
| NET-03 | 40 | 12 |
| NET-05 | 60 | 12 |
| DNS-01 | 52 | 12 |
| DNS-03 | 39 | 12 |
| DNS-04 | 39 | 12 |
| TCP-01 | 37 | 12 |
| TLS-01 | 38 | 12 |
| HTTP-01 | 50 | 12 |

## Planned Cases and Requirement Dispositions

### NET-01

#### M-NET-01-ADMISSION

Intended owner: `internal/validation/net01.go`. Planned case: `internal/validation/net01_conformance_test.go`, `symbol:TestMNet01Admission`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/taskdecode/net01.go`, `internal/admission/measurements.go`, `internal/measurement/net01/run.go`.

Case plan: Accept only empty input; vary wrong root types, null, missing input and every extra property, including interface names and identifiers. Assert rejection before platform collection, traffic, or Measurement Result construction.

<!-- requirement-ref: REQ-NET-01-0005 -->
[REQ-NET-01-0005](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0005) → `M-NET-01-ADMISSION` (Target).

<!-- requirement-ref: REQ-NET-01-0010 -->
[REQ-NET-01-0010](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0010) → `M-NET-01-ADMISSION` (Inputs).

<!-- requirement-ref: REQ-NET-01-0011 -->
[REQ-NET-01-0011](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0011) → `M-NET-01-ADMISSION` (Inputs).

<!-- requirement-ref: REQ-NET-01-0012 -->
[REQ-NET-01-0012](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0012) → `M-NET-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-NET-01-0013 -->
[REQ-NET-01-0013](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0013) → `M-NET-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-NET-01-0014 -->
[REQ-NET-01-0014](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0014) → `M-NET-01-ADMISSION` (Input Validation).

#### M-NET-01-OPTIONS

Intended owner: `internal/validation/net01.go`. Planned case: `internal/validation/net01_conformance_test.go`, `symbol:TestMNet01Options`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/taskdecode/net01.go`, `internal/admission/measurements.go`, `internal/measurement/net01/run.go`.

Case plan: Accept the empty options object and reject every added property or wrong type without changing the selected snapshot behavior.

<!-- requirement-ref: REQ-NET-01-0015 -->
[REQ-NET-01-0015](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0015) → `M-NET-01-OPTIONS` (Options).

<!-- requirement-ref: REQ-NET-01-0016 -->
[REQ-NET-01-0016](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0016) → `M-NET-01-OPTIONS` (Options).

#### M-NET-01-PRECONDITIONS

Intended owner: `internal/measurement/net01/run.go`. Planned case: `internal/measurement/net01/conformance_test.go`, `symbol:TestMNet01Preconditions`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/platform/network.go`, `docs/measurement-platform-validation.md`.

Case plan: Supply a fake platform capability matrix, with required state APIs absent, and separately require forbidden address/SSID/subscriber/location access. Deny the unsafe collection path; review minimum capabilities on Android and Apple separately.

Scope: portable gate tests plus separately required native/host validation; no platform claim.

<!-- requirement-ref: REQ-NET-01-0017 -->
[REQ-NET-01-0017](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0017) → `M-NET-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-NET-01-0018 -->
[REQ-NET-01-0018](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0018) → `M-NET-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-NET-01-0019 -->
[REQ-NET-01-0019](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0019) → `M-NET-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-NET-01-0020 -->
[REQ-NET-01-0020](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0020) → `M-NET-01-PRECONDITIONS` (Preconditions).

#### M-NET-01-EXECUTION

Intended owner: `internal/measurement/net01/run.go`. Planned case: `internal/measurement/net01/conformance_test.go`, `symbol:TestMNet01Execution`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Use a recording platform provider to assert normalized access reads, unknown values, transition detection and release of handles/callbacks. Install a transport that fails on any use; no primary or fallback egress endpoint may be contacted.

<!-- requirement-ref: REQ-NET-01-0006 -->
[REQ-NET-01-0006](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0006) → `M-NET-01-EXECUTION` (Operation).

<!-- requirement-ref: REQ-NET-01-0007 -->
[REQ-NET-01-0007](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0007) → `M-NET-01-EXECUTION` (Procedure).

#### M-NET-01-NORMALIZATION

Intended owner: `internal/measurement/net01/run.go`. Planned case: `internal/measurement/net01/conformance_test.go`, `symbol:TestMNet01Normalization`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/result/net01.go`.

Case plan: Supply coherent, unavailable, failed and inconsistent snapshots; exercise every failure/outcome row and distinguish unknown properties from unavailable/inconsistent snapshots. Verify completed snapshots cannot combine generations.

<!-- requirement-ref: REQ-NET-01-0008 -->
[REQ-NET-01-0008](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0008) → `M-NET-01-NORMALIZATION` (Success Condition).

<!-- requirement-ref: REQ-NET-01-0009 -->
[REQ-NET-01-0009](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0009) → `M-NET-01-NORMALIZATION` (Failure Conditions).

<!-- requirement-ref: REQ-NET-01-0026 -->
[REQ-NET-01-0026](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0026) → `M-NET-01-NORMALIZATION` (Outcomes).

#### M-NET-01-LIMITS

Intended owner: `internal/measurement/net01/run.go`. Planned case: `internal/measurement/net01/conformance_test.go`, `symbol:TestMNet01Limits`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Drive the fake clock to the outer deadline and assert prompt release with zero network operations/bytes. Run concurrent snapshots with a generation transition and assert coherent values or the prescribed incomplete state, never merged state.

<!-- requirement-ref: REQ-NET-01-0021 -->
[REQ-NET-01-0021](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0021) → `M-NET-01-LIMITS` (Resource Bounds).

#### M-NET-01-VERSIONS

Intended owner: `internal/registry/measurements.go`. Planned case: `internal/registry/measurement_conformance_test.go`, `symbol:TestMNet01Versions`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/result/net01.go`.

Case plan: Select exact registered test, Measurement Result schema and test result versions. Vary unsupported/malformed/cross-test versions and schema combinations; reject before execution rather than guessing, coercing or adding another version. Registration metadata and Draft status confer no support.

<!-- requirement-ref: REQ-NET-01-0022 -->
[REQ-NET-01-0022](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0022) → `M-NET-01-VERSIONS` (Output).

#### M-NET-01-OUTPUT

Intended owner: `internal/result/net01.go`. Planned case: `internal/result/net01_conformance_test.go`, `symbol:TestMNet01Output`. Disposition: **planned**, with no execution or support evidence.

Case plan: Verify no top-level network_context, no egress, empty observations and the conditional complete access object for snapshot/inconclusive. Exercise all access enums through the canonical shared format suite.

<!-- requirement-ref: REQ-NET-01-0023 -->
[REQ-NET-01-0023](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0023) → `M-NET-01-OUTPUT` (Output).

<!-- requirement-ref: REQ-NET-01-0024 -->
[REQ-NET-01-0024](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0024) → `M-NET-01-OUTPUT` (Output).

<!-- requirement-ref: REQ-NET-01-0025 -->
[REQ-NET-01-0025](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0025) → `M-NET-01-OUTPUT` (Additional Field).

#### M-NET-01-TERMINATION

Intended owner: `internal/measurement/net01/run.go`. Planned case: `internal/measurement/net01/conformance_test.go`, `symbol:TestMNet01Termination`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Interrupt before/during/after every operation through network-generation change, cancellation, attempt deadline, proxy loss/reconfiguration where applicable, resource bounds and suspension as applicable to each linked source. Assert the exact common execution state/reason and measurement presence/absence; keep NET-05 permitted partial profile output distinct from primitive output. Release resources and prohibit fallback or invented terminal protocol outcomes.

<!-- requirement-ref: REQ-NET-01-0027 -->
[REQ-NET-01-0027](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0027) → `M-NET-01-TERMINATION` (Outcomes).

#### M-NET-01-MINIMIZATION

Intended owner: `internal/result/net01.go`. Planned case: `internal/result/net01_conformance_test.go`, `symbol:TestMNet01Minimization`. Disposition: **planned**, with no execution or support evidence.

Case plan: Inject unique synthetic canaries into every prohibited target, identifier, address, DNS/TLS/HTTP/platform/error field named by each linked unit. Inspect serialized result/error paths and all nested observations, asserting no retained canary or extra field; preserve only explicitly allowed bounded normalized DNS answer values. No participant or live measurement data is used.

<!-- requirement-ref: REQ-NET-01-0028 -->
[REQ-NET-01-0028](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0028) → `M-NET-01-MINIMIZATION` (Outcomes).

#### M-NET-01-INTERPRETATION

Intended owner: `docs/measurement-interpretation-review.md`. Planned case: `docs/measurement-interpretation-review.md`, `heading:M-NET-01-INTERPRETATION`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/measurement/net01/run.go`, `internal/result/net01.go`.

Case plan: Review result/host documentation against every linked unit: coarse OS properties cannot establish Internet reachability, geography/operator identity, effective VPN routing or censorship. Stronger interpretation requires the listed validation/protocol/repeated context and consideration of stale/incomplete platform state.

Scope: separately reviewed integration/interpretation evidence; portable tests cannot complete this family alone.

<!-- requirement-ref: REQ-NET-01-0001 -->
[REQ-NET-01-0001](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0001) → `M-NET-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-NET-01-0002 -->
[REQ-NET-01-0002](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0002) → `M-NET-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-NET-01-0003 -->
[REQ-NET-01-0003](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0003) → `M-NET-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-NET-01-0004 -->
[REQ-NET-01-0004](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0004) → `M-NET-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-NET-01-0029 -->
[REQ-NET-01-0029](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0029) → `M-NET-01-INTERPRETATION` (Supported Interpretation).

<!-- requirement-ref: REQ-NET-01-0030 -->
[REQ-NET-01-0030](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0030) → `M-NET-01-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

<!-- requirement-ref: REQ-NET-01-0031 -->
[REQ-NET-01-0031](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0031) → `M-NET-01-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

#### M-NET-01-SAFETY

Intended owner: `docs/measurement-deployment-review.md`. Planned case: `docs/measurement-deployment-review.md`, `heading:M-NET-01-SAFETY`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/execution/measurements.go`, `internal/result/net01.go`.

Case plan: Review every baseline/factor row and separately exercise each normative safety action using synthetic targets and recording transports. Keep effective risk at least the registered baseline and preserve host/local-policy qualification and final probe authority; no identifiers, unsafe target expansion or distinctive scheduling is introduced. Native capabilities, actual traffic patterns, host scheduling/consent and deployment review require separate scoped evidence; explanatory rationales are not new runtime data.

Scope: separately reviewed integration/interpretation evidence; portable tests cannot complete this family alone.

<!-- requirement-ref: REQ-NET-01-0032 -->
[REQ-NET-01-0032](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0032) → `M-NET-01-SAFETY` (Baseline Risk).

<!-- requirement-ref: REQ-NET-01-0033 -->
[REQ-NET-01-0033](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0033) → `M-NET-01-SAFETY` (Risk Factors).

<!-- requirement-ref: REQ-NET-01-0034 -->
[REQ-NET-01-0034](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0034) → `M-NET-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-NET-01-0035 -->
[REQ-NET-01-0035](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0035) → `M-NET-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-NET-01-0036 -->
[REQ-NET-01-0036](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0036) → `M-NET-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-NET-01-0037 -->
[REQ-NET-01-0037](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0037) → `M-NET-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-NET-01-0038 -->
[REQ-NET-01-0038](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0038) → `M-NET-01-SAFETY` (Safety Analysis).

### NET-03

#### M-NET-03-ADMISSION

Intended owner: `internal/validation/net03.go`. Planned case: `internal/validation/net03_conformance_test.go`, `symbol:TestMNet03Admission`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/taskdecode/net03.go`, `internal/admission/measurements.go`, `internal/measurement/net03/run.go`.

Case plan: Validate HTTP and HTTPS registered endpoints, required expected_status and optional lowercase SHA-256; vary all types, presence, URL length and digest syntax. Reject unsafe/user-specific targets and unknown fields before contact; use a redirect canary destination to prove it is never contacted.

<!-- requirement-ref: REQ-NET-03-0004 -->
[REQ-NET-03-0004](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0004) → `M-NET-03-ADMISSION` (Target).

<!-- requirement-ref: REQ-NET-03-0009 -->
[REQ-NET-03-0009](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0009) → `M-NET-03-ADMISSION` (Inputs).

<!-- requirement-ref: REQ-NET-03-0010 -->
[REQ-NET-03-0010](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0010) → `M-NET-03-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-NET-03-0011 -->
[REQ-NET-03-0011](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0011) → `M-NET-03-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-NET-03-0012 -->
[REQ-NET-03-0012](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0012) → `M-NET-03-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-NET-03-0013 -->
[REQ-NET-03-0013](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0013) → `M-NET-03-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-NET-03-0014 -->
[REQ-NET-03-0014](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0014) → `M-NET-03-ADMISSION` (Input Validation).

#### M-NET-03-OPTIONS

Intended owner: `internal/validation/net03.go`. Planned case: `internal/validation/net03_conformance_test.go`, `symbol:TestMNet03Options`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/taskdecode/net03.go`, `internal/admission/measurements.go`, `internal/measurement/net03/run.go`.

Case plan: Exercise auto/ipv4/ipv6, omitted default and unsupported family; reject redirect-following, custom-header, credential and body options. No option may bypass local policy.

<!-- requirement-ref: REQ-NET-03-0015 -->
[REQ-NET-03-0015](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0015) → `M-NET-03-OPTIONS` (Options).

<!-- requirement-ref: REQ-NET-03-0016 -->
[REQ-NET-03-0016](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0016) → `M-NET-03-OPTIONS` (Options).

#### M-NET-03-PRECONDITIONS

Intended owner: `internal/measurement/net03/run.go`. Planned case: `internal/measurement/net03/conformance_test.go`, `symbol:TestMNet03Preconditions`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/platform/network.go`, `docs/measurement-platform-validation.md`.

Case plan: Vary stable-generation availability, profile freshness/local availability and platform captive-flow behavior. Reject any flow that would display, submit or retain user data; review native platform permission and UI isolation separately.

Scope: portable gate tests plus separately required native/host validation; no platform claim.

<!-- requirement-ref: REQ-NET-03-0017 -->
[REQ-NET-03-0017](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0017) → `M-NET-03-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-NET-03-0018 -->
[REQ-NET-03-0018](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0018) → `M-NET-03-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-NET-03-0019 -->
[REQ-NET-03-0019](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0019) → `M-NET-03-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-NET-03-0020 -->
[REQ-NET-03-0020](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0020) → `M-NET-03-PRECONDITIONS` (Preconditions).

#### M-NET-03-EXECUTION

Intended owner: `internal/measurement/net03/run.go`. Planned case: `internal/measurement/net03/conformance_test.go`, `symbol:TestMNet03Execution`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Record NET-01 stability check, platform validation capture, one request without credentials/cookies or redirect following, bounded comparison, and close. Test HTTP and HTTPS independently, with TLS only for HTTPS; keep a destination canary behind every redirect.

<!-- requirement-ref: REQ-NET-03-0005 -->
[REQ-NET-03-0005](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0005) → `M-NET-03-EXECUTION` (Operation).

<!-- requirement-ref: REQ-NET-03-0006 -->
[REQ-NET-03-0006](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0006) → `M-NET-03-EXECUTION` (Procedure).

#### M-NET-03-NORMALIZATION

Intended owner: `internal/measurement/net03/run.go`. Planned case: `internal/measurement/net03/conformance_test.go`, `symbol:TestMNet03Normalization`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/result/net03.go`.

Case plan: Exercise every failure/outcome row and active-response class with the platform state varied independently, including conflicting/stale/unknown state. Preserve both evidence sources; test expected, mismatch, redirect and unrecognized responses without inventing an outcome precedence when the contract is ambiguous.

<!-- requirement-ref: REQ-NET-03-0007 -->
[REQ-NET-03-0007](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0007) → `M-NET-03-NORMALIZATION` (Success Condition).

<!-- requirement-ref: REQ-NET-03-0008 -->
[REQ-NET-03-0008](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0008) → `M-NET-03-NORMALIZATION` (Failure Conditions).

<!-- requirement-ref: REQ-NET-03-0027 -->
[REQ-NET-03-0027](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0027) → `M-NET-03-NORMALIZATION` (Outcomes).

<!-- requirement-ref: REQ-NET-03-0028 -->
[REQ-NET-03-0028](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0028) → `M-NET-03-NORMALIZATION` (Outcomes).

#### M-NET-03-LIMITS

Intended owner: `internal/measurement/net03/run.go`. Planned case: `internal/measurement/net03/conformance_test.go`, `symbol:TestMNet03Limits`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Test 16,384-byte response headers and each applicable registered response-prefix bound at below/equal/above boundaries, outer deadline, zero retries and one active run per generation. Count actual operations and buffered reads separately; reject overlapping runs as specified.

<!-- requirement-ref: REQ-NET-03-0021 -->
[REQ-NET-03-0021](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0021) → `M-NET-03-LIMITS` (Resource Bounds).

#### M-NET-03-VERSIONS

Intended owner: `internal/registry/measurements.go`. Planned case: `internal/registry/measurement_conformance_test.go`, `symbol:TestMNet03Versions`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/result/net03.go`.

Case plan: Select exact registered test, Measurement Result schema and test result versions. Vary unsupported/malformed/cross-test versions and schema combinations; reject before execution rather than guessing, coercing or adding another version. Registration metadata and Draft status confer no support.

<!-- requirement-ref: REQ-NET-03-0022 -->
[REQ-NET-03-0022](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0022) → `M-NET-03-VERSIONS` (Output).

#### M-NET-03-OUTPUT

Intended owner: `internal/result/net03.go`. Planned case: `internal/result/net03_conformance_test.go`, `symbol:TestMNet03Output`. Disposition: **planned**, with no execution or support evidence.

Case plan: Check required platform_validation, conditional response_class and its complete enum, minimized redirect metadata, and protocol observations actually begun. No redirect location is retained or contacted; platform and active evidence remain distinct.

<!-- requirement-ref: REQ-NET-03-0023 -->
[REQ-NET-03-0023](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0023) → `M-NET-03-OUTPUT` (Output).

<!-- requirement-ref: REQ-NET-03-0024 -->
[REQ-NET-03-0024](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0024) → `M-NET-03-OUTPUT` (Additional Fields).

<!-- requirement-ref: REQ-NET-03-0025 -->
[REQ-NET-03-0025](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0025) → `M-NET-03-OUTPUT` (Additional Fields).

<!-- requirement-ref: REQ-NET-03-0026 -->
[REQ-NET-03-0026](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0026) → `M-NET-03-OUTPUT` (Additional Fields).

#### M-NET-03-TERMINATION

Intended owner: `internal/measurement/net03/run.go`. Planned case: `internal/measurement/net03/conformance_test.go`, `symbol:TestMNet03Termination`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Interrupt before/during/after every operation through network-generation change, cancellation, attempt deadline, proxy loss/reconfiguration where applicable, resource bounds and suspension as applicable to each linked source. Assert the exact common execution state/reason and measurement presence/absence; keep NET-05 permitted partial profile output distinct from primitive output. Release resources and prohibit fallback or invented terminal protocol outcomes.

<!-- requirement-ref: REQ-NET-03-0029 -->
[REQ-NET-03-0029](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0029) → `M-NET-03-TERMINATION` (Outcomes).

#### M-NET-03-MINIMIZATION

Intended owner: `internal/result/net03.go`. Planned case: `internal/result/net03_conformance_test.go`, `symbol:TestMNet03Minimization`. Disposition: **planned**, with no execution or support evidence.

Case plan: Inject unique synthetic canaries into every prohibited target, identifier, address, DNS/TLS/HTTP/platform/error field named by each linked unit. Inspect serialized result/error paths and all nested observations, asserting no retained canary or extra field; preserve only explicitly allowed bounded normalized DNS answer values. No participant or live measurement data is used.

<!-- requirement-ref: REQ-NET-03-0030 -->
[REQ-NET-03-0030](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0030) → `M-NET-03-MINIMIZATION` (Outcomes).

#### M-NET-03-INTERPRETATION

Intended owner: `docs/measurement-interpretation-review.md`. Planned case: `docs/measurement-interpretation-review.md`, `heading:M-NET-03-INTERPRETATION`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/measurement/net03/run.go`, `internal/result/net03.go`.

Case plan: Review interpretation for a single configured endpoint and the stated stronger-evidence criteria, including stale platform state and endpoint/VPN/portal failures. General Internet reachability, portal-operator intent and censorship cannot be inferred from one validation result.

Scope: separately reviewed integration/interpretation evidence; portable tests cannot complete this family alone.

<!-- requirement-ref: REQ-NET-03-0001 -->
[REQ-NET-03-0001](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0001) → `M-NET-03-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-NET-03-0002 -->
[REQ-NET-03-0002](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0002) → `M-NET-03-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-NET-03-0003 -->
[REQ-NET-03-0003](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0003) → `M-NET-03-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-NET-03-0031 -->
[REQ-NET-03-0031](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0031) → `M-NET-03-INTERPRETATION` (Supported Interpretation).

<!-- requirement-ref: REQ-NET-03-0032 -->
[REQ-NET-03-0032](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0032) → `M-NET-03-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

<!-- requirement-ref: REQ-NET-03-0033 -->
[REQ-NET-03-0033](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0033) → `M-NET-03-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

#### M-NET-03-SAFETY

Intended owner: `docs/measurement-deployment-review.md`. Planned case: `docs/measurement-deployment-review.md`, `heading:M-NET-03-SAFETY`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/execution/measurements.go`, `internal/result/net03.go`.

Case plan: Review every baseline/factor row and separately exercise each normative safety action using synthetic targets and recording transports. Keep effective risk at least the registered baseline and preserve host/local-policy qualification and final probe authority; no identifiers, unsafe target expansion or distinctive scheduling is introduced. Native capabilities, actual traffic patterns, host scheduling/consent and deployment review require separate scoped evidence; explanatory rationales are not new runtime data.

Scope: separately reviewed integration/interpretation evidence; portable tests cannot complete this family alone.

<!-- requirement-ref: REQ-NET-03-0034 -->
[REQ-NET-03-0034](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0034) → `M-NET-03-SAFETY` (Baseline Risk).

<!-- requirement-ref: REQ-NET-03-0035 -->
[REQ-NET-03-0035](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0035) → `M-NET-03-SAFETY` (Risk Factors).

<!-- requirement-ref: REQ-NET-03-0036 -->
[REQ-NET-03-0036](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0036) → `M-NET-03-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-NET-03-0037 -->
[REQ-NET-03-0037](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0037) → `M-NET-03-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-NET-03-0038 -->
[REQ-NET-03-0038](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0038) → `M-NET-03-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-NET-03-0039 -->
[REQ-NET-03-0039](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0039) → `M-NET-03-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-NET-03-0040 -->
[REQ-NET-03-0040](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0040) → `M-NET-03-SAFETY` (Safety Analysis).

### NET-05

#### M-NET-05-ADMISSION

Intended owner: `internal/validation/net05.go`. Planned case: `internal/validation/net05_conformance_test.go`, `symbol:TestMNet05Admission`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/taskdecode/net05.go`, `internal/admission/measurements.go`, `internal/measurement/net05/run.go`.

Case plan: Validate the outer profile, profile_id/version, each step_id/test_id/exact version/input/options and all nested closed schemas. Test 0/1/8/9 steps, identifier lengths 0/1/64/65, duplicate IDs, nonpositive/nonmonotone versions, unsupported/nested tests, forbidden nested envelopes and interpolation. Put an invalid or unsafe final step after valid earlier steps and assert zero primitive operations.

<!-- requirement-ref: REQ-NET-05-0005 -->
[REQ-NET-05-0005](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0005) → `M-NET-05-ADMISSION` (Target).

<!-- requirement-ref: REQ-NET-05-0014 -->
[REQ-NET-05-0014](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0014) → `M-NET-05-ADMISSION` (Inputs).

<!-- requirement-ref: REQ-NET-05-0015 -->
[REQ-NET-05-0015](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0015) → `M-NET-05-ADMISSION` (Inputs).

<!-- requirement-ref: REQ-NET-05-0016 -->
[REQ-NET-05-0016](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0016) → `M-NET-05-ADMISSION` (Inputs).

<!-- requirement-ref: REQ-NET-05-0017 -->
[REQ-NET-05-0017](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0017) → `M-NET-05-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-NET-05-0018 -->
[REQ-NET-05-0018](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0018) → `M-NET-05-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-NET-05-0019 -->
[REQ-NET-05-0019](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0019) → `M-NET-05-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-NET-05-0020 -->
[REQ-NET-05-0020](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0020) → `M-NET-05-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-NET-05-0021 -->
[REQ-NET-05-0021](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0021) → `M-NET-05-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-NET-05-0022 -->
[REQ-NET-05-0022](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0022) → `M-NET-05-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-NET-05-0023 -->
[REQ-NET-05-0023](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0023) → `M-NET-05-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-NET-05-0024 -->
[REQ-NET-05-0024](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0024) → `M-NET-05-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-NET-05-0025 -->
[REQ-NET-05-0025](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0025) → `M-NET-05-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-NET-05-0026 -->
[REQ-NET-05-0026](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0026) → `M-NET-05-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-NET-05-0027 -->
[REQ-NET-05-0027](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0027) → `M-NET-05-ADMISSION` (Input Validation).

#### M-NET-05-OPTIONS

Intended owner: `internal/validation/net05.go`. Planned case: `internal/validation/net05_conformance_test.go`, `symbol:TestMNet05Options`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/taskdecode/net05.go`, `internal/admission/measurements.go`, `internal/measurement/net05/run.go`.

Case plan: Accept only empty composite options. Independently vary each primitive option via its exact versioned schema; reject attempts to override sequence, outer execution/safety, primitive limits or no-retry behavior.

<!-- requirement-ref: REQ-NET-05-0028 -->
[REQ-NET-05-0028](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0028) → `M-NET-05-OPTIONS` (Options).

<!-- requirement-ref: REQ-NET-05-0029 -->
[REQ-NET-05-0029](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0029) → `M-NET-05-OPTIONS` (Options).

<!-- requirement-ref: REQ-NET-05-0030 -->
[REQ-NET-05-0030](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0030) → `M-NET-05-OPTIONS` (Options).

#### M-NET-05-PRECONDITIONS

Intended owner: `internal/measurement/net05/run.go`. Planned case: `internal/measurement/net05/conformance_test.go`, `symbol:TestMNet05Preconditions`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/platform/network.go`, `docs/measurement-platform-validation.md`.

Case plan: Vary trusted source/profile, target review, explicit host enablement/anomaly policy, network stability, required validation and remaining envelope capacity. Review dense/synchronized pattern risk; ensure denial before any unsafe profile step and policy withdrawal stops the remaining sequence.

Scope: portable gate tests plus separately required native/host validation; no platform claim.

<!-- requirement-ref: REQ-NET-05-0031 -->
[REQ-NET-05-0031](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0031) → `M-NET-05-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-NET-05-0032 -->
[REQ-NET-05-0032](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0032) → `M-NET-05-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-NET-05-0033 -->
[REQ-NET-05-0033](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0033) → `M-NET-05-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-NET-05-0034 -->
[REQ-NET-05-0034](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0034) → `M-NET-05-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-NET-05-0035 -->
[REQ-NET-05-0035](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0035) → `M-NET-05-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-NET-05-0036 -->
[REQ-NET-05-0036](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0036) → `M-NET-05-PRECONDITIONS` (Preconditions).

#### M-NET-05-EXECUTION

Intended owner: `internal/measurement/net05/run.go`. Planned case: `internal/measurement/net05/conformance_test.go`, `symbol:TestMNet05Execution`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Record admission of the entire profile and effective risk before any connection, current NET-01, permitted fresh/cached/platform NET-03 context and strictly sequential DNS-01/TCP-01/TLS-01/HTTP-01 dispatch. Each primitive must use its own schema/runner, with no parallel steps, fallback, retries, target expansion or combined verdict.

<!-- requirement-ref: REQ-NET-05-0006 -->
[REQ-NET-05-0006](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0006) → `M-NET-05-EXECUTION` (Operation).

<!-- requirement-ref: REQ-NET-05-0007 -->
[REQ-NET-05-0007](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0007) → `M-NET-05-EXECUTION` (Operation).

<!-- requirement-ref: REQ-NET-05-0008 -->
[REQ-NET-05-0008](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0008) → `M-NET-05-EXECUTION` (Operation).

<!-- requirement-ref: REQ-NET-05-0009 -->
[REQ-NET-05-0009](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0009) → `M-NET-05-EXECUTION` (Procedure).

#### M-NET-05-NORMALIZATION

Intended owner: `internal/measurement/net05/run.go`. Planned case: `internal/measurement/net05/conformance_test.go`, `symbol:TestMNet05Normalization`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/result/net05.go`.

Case plan: Exercise every composite failure/outcome row using positive and negative terminal primitive results. All completed primitive results yield completed composite execution; missing validation under requiring policy runs no primitive. Represent network/resource termination at the common execution layer without rewriting primitive failures.

<!-- requirement-ref: REQ-NET-05-0010 -->
[REQ-NET-05-0010](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0010) → `M-NET-05-NORMALIZATION` (Success Condition).

<!-- requirement-ref: REQ-NET-05-0011 -->
[REQ-NET-05-0011](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0011) → `M-NET-05-NORMALIZATION` (Failure Conditions).

<!-- requirement-ref: REQ-NET-05-0012 -->
[REQ-NET-05-0012](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0012) → `M-NET-05-NORMALIZATION` (Failure Conditions).

<!-- requirement-ref: REQ-NET-05-0013 -->
[REQ-NET-05-0013](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0013) → `M-NET-05-NORMALIZATION` (Failure Conditions).

<!-- requirement-ref: REQ-NET-05-0046 -->
[REQ-NET-05-0046](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0046) → `M-NET-05-NORMALIZATION` (Outcomes).

#### M-NET-05-LIMITS

Intended owner: `internal/measurement/net05/run.go`. Planned case: `internal/measurement/net05/conformance_test.go`, `symbol:TestMNet05Limits`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Test 49,152 encoded profile bytes, 60,000 ms, 12 connections/tunnels, 65,536 application bytes sent and 262,144 received at below/equal/above boundaries and against stricter primitive/task bounds. Reject over-budget declared maxima atomically before contact; enforce actual aggregate use and one sequential profile run per generation.

<!-- requirement-ref: REQ-NET-05-0037 -->
[REQ-NET-05-0037](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0037) → `M-NET-05-LIMITS` (Resource Bounds).

<!-- requirement-ref: REQ-NET-05-0038 -->
[REQ-NET-05-0038](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0038) → `M-NET-05-LIMITS` (Resource Bounds).

#### M-NET-05-VERSIONS

Intended owner: `internal/registry/measurements.go`. Planned case: `internal/registry/measurement_conformance_test.go`, `symbol:TestMNet05Versions`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/result/net05.go`.

Case plan: Select exact registered test, Measurement Result schema and test result versions. Vary unsupported/malformed/cross-test versions and schema combinations; reject before execution rather than guessing, coercing or adding another version. Registration metadata and Draft status confer no support.

<!-- requirement-ref: REQ-NET-05-0039 -->
[REQ-NET-05-0039](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0039) → `M-NET-05-VERSIONS` (Output).

#### M-NET-05-OUTPUT

Intended owner: `internal/result/net05.go`. Planned case: `internal/result/net05_conformance_test.go`, `symbol:TestMNet05Output`. Disposition: **planned**, with no execution or support evidence.

Case plan: Validate exact profile identity, required network_validation state/source, ordered bounded started steps, exact primitive/result versions and conditional primitive measurement. Test interruption before each step and inside the last started step: at most one final incomplete step, no placeholders, exact stopped_before_step and one outer Network Context.

<!-- requirement-ref: REQ-NET-05-0040 -->
[REQ-NET-05-0040](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0040) → `M-NET-05-OUTPUT` (Output).

<!-- requirement-ref: REQ-NET-05-0041 -->
[REQ-NET-05-0041](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0041) → `M-NET-05-OUTPUT` (Additional Fields).

<!-- requirement-ref: REQ-NET-05-0042 -->
[REQ-NET-05-0042](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0042) → `M-NET-05-OUTPUT` (Additional Fields).

<!-- requirement-ref: REQ-NET-05-0043 -->
[REQ-NET-05-0043](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0043) → `M-NET-05-OUTPUT` (Additional Fields).

<!-- requirement-ref: REQ-NET-05-0044 -->
[REQ-NET-05-0044](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0044) → `M-NET-05-OUTPUT` (Additional Fields).

<!-- requirement-ref: REQ-NET-05-0045 -->
[REQ-NET-05-0045](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0045) → `M-NET-05-OUTPUT` (Additional Fields).

#### M-NET-05-TERMINATION

Intended owner: `internal/measurement/net05/run.go`. Planned case: `internal/measurement/net05/conformance_test.go`, `symbol:TestMNet05Termination`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Interrupt before/during/after every operation through network-generation change, cancellation, attempt deadline, proxy loss/reconfiguration where applicable, resource bounds and suspension as applicable to each linked source. Assert the exact common execution state/reason and measurement presence/absence; keep NET-05 permitted partial profile output distinct from primitive output. Release resources and prohibit fallback or invented terminal protocol outcomes.

<!-- requirement-ref: REQ-NET-05-0047 -->
[REQ-NET-05-0047](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0047) → `M-NET-05-TERMINATION` (Outcomes).

#### M-NET-05-MINIMIZATION

Intended owner: `internal/result/net05.go`. Planned case: `internal/result/net05_conformance_test.go`, `symbol:TestMNet05Minimization`. Disposition: **planned**, with no execution or support evidence.

Case plan: Inject unique synthetic canaries into every prohibited target, identifier, address, DNS/TLS/HTTP/platform/error field named by each linked unit. Inspect serialized result/error paths and all nested observations, asserting no retained canary or extra field; preserve only explicitly allowed bounded normalized DNS answer values. No participant or live measurement data is used.

<!-- requirement-ref: REQ-NET-05-0048 -->
[REQ-NET-05-0048](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0048) → `M-NET-05-MINIMIZATION` (Outcomes).

#### M-NET-05-INTERPRETATION

Intended owner: `docs/measurement-interpretation-review.md`. Planned case: `docs/measurement-interpretation-review.md`, `heading:M-NET-05-INTERPRETATION`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/measurement/net05/run.go`, `internal/result/net05.go`.

Case plan: Review every claim boundary and stronger-evidence criterion across stable repeated profiles, independent target operators/layers, controls and external signals. No automatic allowlist/anomaly/censorship verdict, certainty about intent, or treatment of reference datasets as ground truth; analyst correlation remains outside NET-05.

Scope: separately reviewed integration/interpretation evidence; portable tests cannot complete this family alone.

<!-- requirement-ref: REQ-NET-05-0001 -->
[REQ-NET-05-0001](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0001) → `M-NET-05-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-NET-05-0002 -->
[REQ-NET-05-0002](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0002) → `M-NET-05-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-NET-05-0003 -->
[REQ-NET-05-0003](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0003) → `M-NET-05-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-NET-05-0004 -->
[REQ-NET-05-0004](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0004) → `M-NET-05-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-NET-05-0049 -->
[REQ-NET-05-0049](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0049) → `M-NET-05-INTERPRETATION` (Supported Interpretation).

<!-- requirement-ref: REQ-NET-05-0050 -->
[REQ-NET-05-0050](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0050) → `M-NET-05-INTERPRETATION` (Supported Interpretation).

<!-- requirement-ref: REQ-NET-05-0051 -->
[REQ-NET-05-0051](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0051) → `M-NET-05-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

<!-- requirement-ref: REQ-NET-05-0052 -->
[REQ-NET-05-0052](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0052) → `M-NET-05-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

#### M-NET-05-SAFETY

Intended owner: `docs/measurement-deployment-review.md`. Planned case: `docs/measurement-deployment-review.md`, `heading:M-NET-05-SAFETY`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/execution/measurements.go`, `internal/result/net05.go`.

Case plan: Review every baseline/factor row and separately exercise each normative safety action using synthetic targets and recording transports. Keep effective risk at least the registered baseline and preserve host/local-policy qualification and final probe authority; no identifiers, unsafe target expansion or distinctive scheduling is introduced. Native capabilities, actual traffic patterns, host scheduling/consent and deployment review require separate scoped evidence; explanatory rationales are not new runtime data.

Scope: separately reviewed integration/interpretation evidence; portable tests cannot complete this family alone.

<!-- requirement-ref: REQ-NET-05-0053 -->
[REQ-NET-05-0053](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0053) → `M-NET-05-SAFETY` (Baseline Risk).

<!-- requirement-ref: REQ-NET-05-0054 -->
[REQ-NET-05-0054](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0054) → `M-NET-05-SAFETY` (Risk Factors).

<!-- requirement-ref: REQ-NET-05-0055 -->
[REQ-NET-05-0055](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0055) → `M-NET-05-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-NET-05-0056 -->
[REQ-NET-05-0056](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0056) → `M-NET-05-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-NET-05-0057 -->
[REQ-NET-05-0057](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0057) → `M-NET-05-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-NET-05-0058 -->
[REQ-NET-05-0058](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0058) → `M-NET-05-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-NET-05-0059 -->
[REQ-NET-05-0059](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0059) → `M-NET-05-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-NET-05-0060 -->
[REQ-NET-05-0060](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0060) → `M-NET-05-SAFETY` (Safety Analysis).

### DNS-01

#### M-DNS-01-ADMISSION

Intended owner: `internal/validation/dns01.go`. Planned case: `internal/validation/dns01_conformance_test.go`, `symbol:TestMDns01Admission`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/taskdecode/dns01.go`, `internal/admission/measurements.go`, `internal/measurement/dns01/run.go`.

Case plan: Vary IDNA-normalized hostname boundaries through 253 ASCII bytes, A/AAAA enums, local/root/single-label/private names and unknown fields. Independently test resolver presence by mode, canonical public address/port syntax, default port and exact local registration. Invalid input never resolves or emits a network-interference result; observed CNAMEs never cause application connections.

<!-- requirement-ref: REQ-DNS-01-0005 -->
[REQ-DNS-01-0005](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0005) → `M-DNS-01-ADMISSION` (Target).

<!-- requirement-ref: REQ-DNS-01-0006 -->
[REQ-DNS-01-0006](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0006) → `M-DNS-01-ADMISSION` (Target).

<!-- requirement-ref: REQ-DNS-01-0014 -->
[REQ-DNS-01-0014](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0014) → `M-DNS-01-ADMISSION` (Inputs).

<!-- requirement-ref: REQ-DNS-01-0015 -->
[REQ-DNS-01-0015](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0015) → `M-DNS-01-ADMISSION` (Inputs).

<!-- requirement-ref: REQ-DNS-01-0016 -->
[REQ-DNS-01-0016](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0016) → `M-DNS-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-DNS-01-0017 -->
[REQ-DNS-01-0017](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0017) → `M-DNS-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-DNS-01-0018 -->
[REQ-DNS-01-0018](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0018) → `M-DNS-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-DNS-01-0019 -->
[REQ-DNS-01-0019](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0019) → `M-DNS-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-DNS-01-0020 -->
[REQ-DNS-01-0020](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0020) → `M-DNS-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-DNS-01-0021 -->
[REQ-DNS-01-0021](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0021) → `M-DNS-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-DNS-01-0022 -->
[REQ-DNS-01-0022](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0022) → `M-DNS-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-DNS-01-0023 -->
[REQ-DNS-01-0023](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0023) → `M-DNS-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-DNS-01-0024 -->
[REQ-DNS-01-0024](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0024) → `M-DNS-01-ADMISSION` (Input Validation).

#### M-DNS-01-OPTIONS

Intended owner: `internal/validation/dns01.go`. Planned case: `internal/validation/dns01_conformance_test.go`, `symbol:TestMDns01Options`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/taskdecode/dns01.go`, `internal/admission/measurements.go`, `internal/measurement/dns01/run.go`.

Case plan: Cover omitted/system/direct modes, direct-only UDP/TCP, collect_additional_responses and conditional 1..10,000 ms window. Test every missing/null/wrong/unknown option and prohibited combination, with no fallback to another mode, resolver or transport.

<!-- requirement-ref: REQ-DNS-01-0025 -->
[REQ-DNS-01-0025](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0025) → `M-DNS-01-OPTIONS` (Options).

<!-- requirement-ref: REQ-DNS-01-0026 -->
[REQ-DNS-01-0026](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0026) → `M-DNS-01-OPTIONS` (Options).

#### M-DNS-01-PRECONDITIONS

Intended owner: `internal/measurement/dns01/run.go`. Planned case: `internal/measurement/dns01/conformance_test.go`, `symbol:TestMDns01Preconditions`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/platform/network.go`, `docs/measurement-platform-validation.md`.

Case plan: Vary generation stability, hostname ownership/review, resolver registration, availability of a narrow resolver API and explicit UDP window/count bounds. Reject APIs exposing unrelated DNS activity; host review and platform visibility require separate evidence.

Scope: portable gate tests plus separately required native/host validation; no platform claim.

<!-- requirement-ref: REQ-DNS-01-0027 -->
[REQ-DNS-01-0027](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0027) → `M-DNS-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-DNS-01-0028 -->
[REQ-DNS-01-0028](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0028) → `M-DNS-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-DNS-01-0029 -->
[REQ-DNS-01-0029](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0029) → `M-DNS-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-DNS-01-0030 -->
[REQ-DNS-01-0030](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0030) → `M-DNS-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-DNS-01-0031 -->
[REQ-DNS-01-0031](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0031) → `M-DNS-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-DNS-01-0032 -->
[REQ-DNS-01-0032](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0032) → `M-DNS-01-PRECONDITIONS` (Preconditions).

#### M-DNS-01-EXECUTION

Intended owner: `internal/measurement/dns01/run.go`. Planned case: `internal/measurement/dns01/conformance_test.go`, `symbol:TestMDns01Execution`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Record one ordinary system lookup or one direct UDP/TCP query. In UDP, vary source, transaction ID and question, unrelated datagrams, multiple matching datagrams, collection enablement and generation changes; preserve matching arrival order and coarse offsets without changing resolver configuration or contacting another name/server/transport.

<!-- requirement-ref: REQ-DNS-01-0007 -->
[REQ-DNS-01-0007](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0007) → `M-DNS-01-EXECUTION` (Operation).

<!-- requirement-ref: REQ-DNS-01-0008 -->
[REQ-DNS-01-0008](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0008) → `M-DNS-01-EXECUTION` (Operation).

<!-- requirement-ref: REQ-DNS-01-0009 -->
[REQ-DNS-01-0009](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0009) → `M-DNS-01-EXECUTION` (Operation).

<!-- requirement-ref: REQ-DNS-01-0010 -->
[REQ-DNS-01-0010](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0010) → `M-DNS-01-EXECUTION` (Procedure).

#### M-DNS-01-NORMALIZATION

Intended owner: `internal/measurement/dns01/run.go`. Planned case: `internal/measurement/dns01/conformance_test.go`, `symbol:TestMDns01Normalization`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/result/dns01.go`.

Case plan: Exercise all failure/outcome rows and precedence with usable, unrelated, mismatching, identical and conflicting responses. Conflict outranks resolved; a count-truncated collection yields response_limit_reached as specified, with the flag. Test NXDOMAIN/empty answers without classifying them as successful address resolution and preserve opaque system error limits.

<!-- requirement-ref: REQ-DNS-01-0011 -->
[REQ-DNS-01-0011](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0011) → `M-DNS-01-NORMALIZATION` (Success Condition).

<!-- requirement-ref: REQ-DNS-01-0012 -->
[REQ-DNS-01-0012](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0012) → `M-DNS-01-NORMALIZATION` (Success Condition).

<!-- requirement-ref: REQ-DNS-01-0013 -->
[REQ-DNS-01-0013](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0013) → `M-DNS-01-NORMALIZATION` (Failure Conditions).

<!-- requirement-ref: REQ-DNS-01-0037 -->
[REQ-DNS-01-0037](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0037) → `M-DNS-01-NORMALIZATION` (Outcomes).

<!-- requirement-ref: REQ-DNS-01-0038 -->
[REQ-DNS-01-0038](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0038) → `M-DNS-01-NORMALIZATION` (Outcomes).

#### M-DNS-01-LIMITS

Intended owner: `internal/measurement/dns01/run.go`. Planned case: `internal/measurement/dns01/conformance_test.go`, `symbol:TestMDns01Limits`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Exercise UDP 4,096-byte/TCP 65,535-byte messages, three matching responses, 16 normalized entries/eight CNAMEs and 16,384 encoded observation bytes at below/equal/above boundaries; test the additional window and outer deadline independently. Assert no test-layer retries and deduplication by the complete hostname/type/mode/resolver/generation tuple.

<!-- requirement-ref: REQ-DNS-01-0033 -->
[REQ-DNS-01-0033](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0033) → `M-DNS-01-LIMITS` (Resource Bounds).

#### M-DNS-01-VERSIONS

Intended owner: `internal/registry/measurements.go`. Planned case: `internal/registry/measurement_conformance_test.go`, `symbol:TestMDns01Versions`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/result/dns01.go`.

Case plan: Select exact registered test, Measurement Result schema and test result versions. Vary unsupported/malformed/cross-test versions and schema combinations; reject before execution rather than guessing, coercing or adding another version. Registration metadata and Draft status confer no support.

<!-- requirement-ref: REQ-DNS-01-0034 -->
[REQ-DNS-01-0034](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0034) → `M-DNS-01-VERSIONS` (Output).

#### M-DNS-01-OUTPUT

Intended owner: `internal/result/dns01.go`. Planned case: `internal/result/dns01_conformance_test.go`, `symbol:TestMDns01Output`. Disposition: **planned**, with no execution or support evidence.

Case plan: Require exactly one dns_query with mode-correct engine, observable fields only and bounded ordered responses. System API omissions must remain omissions rather than invented codes/datagrams/timing; reuse the shared DNS observation schema for answer normalization and length/count limits.

<!-- requirement-ref: REQ-DNS-01-0035 -->
[REQ-DNS-01-0035](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0035) → `M-DNS-01-OUTPUT` (Output).

<!-- requirement-ref: REQ-DNS-01-0036 -->
[REQ-DNS-01-0036](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0036) → `M-DNS-01-OUTPUT` (Output).

#### M-DNS-01-TERMINATION

Intended owner: `internal/measurement/dns01/run.go`. Planned case: `internal/measurement/dns01/conformance_test.go`, `symbol:TestMDns01Termination`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Interrupt before/during/after every operation through network-generation change, cancellation, attempt deadline, proxy loss/reconfiguration where applicable, resource bounds and suspension as applicable to each linked source. Assert the exact common execution state/reason and measurement presence/absence; keep NET-05 permitted partial profile output distinct from primitive output. Release resources and prohibit fallback or invented terminal protocol outcomes.

<!-- requirement-ref: REQ-DNS-01-0039 -->
[REQ-DNS-01-0039](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0039) → `M-DNS-01-TERMINATION` (Outcomes).

#### M-DNS-01-MINIMIZATION

Intended owner: `internal/result/dns01.go`. Planned case: `internal/result/dns01_conformance_test.go`, `symbol:TestMDns01Minimization`. Disposition: **planned**, with no execution or support evidence.

Case plan: Inject unique synthetic canaries into every prohibited target, identifier, address, DNS/TLS/HTTP/platform/error field named by each linked unit. Inspect serialized result/error paths and all nested observations, asserting no retained canary or extra field; preserve only explicitly allowed bounded normalized DNS answer values. No participant or live measurement data is used.

<!-- requirement-ref: REQ-DNS-01-0040 -->
[REQ-DNS-01-0040](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0040) → `M-DNS-01-MINIMIZATION` (Outcomes).

#### M-DNS-01-INTERPRETATION

Intended owner: `docs/measurement-interpretation-review.md`. Planned case: `docs/measurement-interpretation-review.md`, `heading:M-DNS-01-INTERPRETATION`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/measurement/dns01/run.go`, `internal/result/dns01.go`.

Case plan: Review scope to the chosen resolver/time/generation and observed conflicting datagrams. Neither the first nor later answer is automatically correct, injected or authoritative; resolver differences and valid negative answers require the listed controls, independent evidence and alternative DNS explanations.

Scope: separately reviewed integration/interpretation evidence; portable tests cannot complete this family alone.

<!-- requirement-ref: REQ-DNS-01-0001 -->
[REQ-DNS-01-0001](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0001) → `M-DNS-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-DNS-01-0002 -->
[REQ-DNS-01-0002](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0002) → `M-DNS-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-DNS-01-0003 -->
[REQ-DNS-01-0003](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0003) → `M-DNS-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-DNS-01-0004 -->
[REQ-DNS-01-0004](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0004) → `M-DNS-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-DNS-01-0041 -->
[REQ-DNS-01-0041](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0041) → `M-DNS-01-INTERPRETATION` (Supported Interpretation).

<!-- requirement-ref: REQ-DNS-01-0042 -->
[REQ-DNS-01-0042](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0042) → `M-DNS-01-INTERPRETATION` (Supported Interpretation).

<!-- requirement-ref: REQ-DNS-01-0043 -->
[REQ-DNS-01-0043](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0043) → `M-DNS-01-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

<!-- requirement-ref: REQ-DNS-01-0044 -->
[REQ-DNS-01-0044](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0044) → `M-DNS-01-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

#### M-DNS-01-SAFETY

Intended owner: `docs/measurement-deployment-review.md`. Planned case: `docs/measurement-deployment-review.md`, `heading:M-DNS-01-SAFETY`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/execution/measurements.go`, `internal/result/dns01.go`.

Case plan: Review every baseline/factor row and separately exercise each normative safety action using synthetic targets and recording transports. Keep effective risk at least the registered baseline and preserve host/local-policy qualification and final probe authority; no identifiers, unsafe target expansion or distinctive scheduling is introduced. Native capabilities, actual traffic patterns, host scheduling/consent and deployment review require separate scoped evidence; explanatory rationales are not new runtime data.

Scope: separately reviewed integration/interpretation evidence; portable tests cannot complete this family alone.

<!-- requirement-ref: REQ-DNS-01-0045 -->
[REQ-DNS-01-0045](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0045) → `M-DNS-01-SAFETY` (Baseline Risk).

<!-- requirement-ref: REQ-DNS-01-0046 -->
[REQ-DNS-01-0046](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0046) → `M-DNS-01-SAFETY` (Risk Factors).

<!-- requirement-ref: REQ-DNS-01-0047 -->
[REQ-DNS-01-0047](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0047) → `M-DNS-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-DNS-01-0048 -->
[REQ-DNS-01-0048](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0048) → `M-DNS-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-DNS-01-0049 -->
[REQ-DNS-01-0049](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0049) → `M-DNS-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-DNS-01-0050 -->
[REQ-DNS-01-0050](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0050) → `M-DNS-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-DNS-01-0051 -->
[REQ-DNS-01-0051](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0051) → `M-DNS-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-DNS-01-0052 -->
[REQ-DNS-01-0052](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0052) → `M-DNS-01-SAFETY` (Safety Analysis).

### DNS-03

#### M-DNS-03-ADMISSION

Intended owner: `internal/validation/dns03.go`. Planned case: `internal/validation/dns03_conformance_test.go`, `symbol:TestMDns03Admission`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/taskdecode/dns03.go`, `internal/admission/measurements.go`, `internal/measurement/dns03/run.go`.

Case plan: Vary resolver_url HTTPS/2,048-byte syntax, reviewed hostname/A-or-AAAA and optional canonical bootstrap address against the exact registered resolver. Test credentials/fragments/user parameters/local/interpolated/unknown targets. Bootstrap IP must not change certificate hostname authentication; redirect targets are never contacted.

<!-- requirement-ref: REQ-DNS-03-0004 -->
[REQ-DNS-03-0004](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0004) → `M-DNS-03-ADMISSION` (Target).

<!-- requirement-ref: REQ-DNS-03-0009 -->
[REQ-DNS-03-0009](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0009) → `M-DNS-03-ADMISSION` (Inputs).

<!-- requirement-ref: REQ-DNS-03-0010 -->
[REQ-DNS-03-0010](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0010) → `M-DNS-03-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-DNS-03-0011 -->
[REQ-DNS-03-0011](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0011) → `M-DNS-03-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-DNS-03-0012 -->
[REQ-DNS-03-0012](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0012) → `M-DNS-03-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-DNS-03-0013 -->
[REQ-DNS-03-0013](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0013) → `M-DNS-03-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-DNS-03-0014 -->
[REQ-DNS-03-0014](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0014) → `M-DNS-03-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-DNS-03-0015 -->
[REQ-DNS-03-0015](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0015) → `M-DNS-03-ADMISSION` (Input Validation).

#### M-DNS-03-OPTIONS

Intended owner: `internal/validation/dns03.go`. Planned case: `internal/validation/dns03_conformance_test.go`, `symbol:TestMDns03Options`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/taskdecode/dns03.go`, `internal/admission/measurements.go`, `internal/measurement/dns03/run.go`.

Case plan: Exercise auto/ipv4/ipv6 and compatibility with bootstrap_ip, including missing/default/invalid/unknown cases. Reject GET, redirects, alternate endpoints, custom headers, credentials/cookies and task proxy configuration.

<!-- requirement-ref: REQ-DNS-03-0016 -->
[REQ-DNS-03-0016](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0016) → `M-DNS-03-OPTIONS` (Options).

<!-- requirement-ref: REQ-DNS-03-0017 -->
[REQ-DNS-03-0017](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0017) → `M-DNS-03-OPTIONS` (Options).

#### M-DNS-03-PRECONDITIONS

Intended owner: `internal/measurement/dns03/run.go`. Planned case: `internal/measurement/dns03/conformance_test.go`, `symbol:TestMDns03Preconditions`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/platform/network.go`, `docs/measurement-platform-validation.md`.

Case plan: Vary resolver/query technical/privacy/safety approval, local profile availability, explicit host enablement or approved anomaly-trigger policy and stable generation/deadline. Assert no configuration download or query before gates pass; deployment risk review requires separate evidence.

Scope: portable gate tests plus separately required native/host validation; no platform claim.

<!-- requirement-ref: REQ-DNS-03-0018 -->
[REQ-DNS-03-0018](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0018) → `M-DNS-03-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-DNS-03-0019 -->
[REQ-DNS-03-0019](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0019) → `M-DNS-03-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-DNS-03-0020 -->
[REQ-DNS-03-0020](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0020) → `M-DNS-03-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-DNS-03-0021 -->
[REQ-DNS-03-0021](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0021) → `M-DNS-03-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-DNS-03-0022 -->
[REQ-DNS-03-0022](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0022) → `M-DNS-03-PRECONDITIONS` (Preconditions).

#### M-DNS-03-EXECUTION

Intended owner: `internal/measurement/dns03/run.go`. Planned case: `internal/measurement/dns03/conformance_test.go`, `symbol:TestMDns03Execution`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Record optional system bootstrap, one selected resolver TCP connection, authenticated TLS, one RFC 8484 POST with application/dns-message and bounded response validation. Vary returned media type, DNS ID/name/type mismatch, malformed/oversized message and each failed stage; prohibit redirects, endpoint/transport fallback and application retries.

<!-- requirement-ref: REQ-DNS-03-0005 -->
[REQ-DNS-03-0005](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0005) → `M-DNS-03-EXECUTION` (Operation).

<!-- requirement-ref: REQ-DNS-03-0006 -->
[REQ-DNS-03-0006](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0006) → `M-DNS-03-EXECUTION` (Procedure).

#### M-DNS-03-NORMALIZATION

Intended owner: `internal/measurement/dns03/run.go`. Planned case: `internal/measurement/dns03/conformance_test.go`, `symbol:TestMDns03Normalization`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/result/dns03.go`.

Case plan: Exercise all terminal outcome rows, including non-success HTTP, content type, DNS response-code failures, malformed/mismatching DNS, per-stage timeout and inconclusive errors. Attribute failure only to the actually observed stage and preserve ordered preceding observations.

<!-- requirement-ref: REQ-DNS-03-0007 -->
[REQ-DNS-03-0007](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0007) → `M-DNS-03-NORMALIZATION` (Success Condition).

<!-- requirement-ref: REQ-DNS-03-0008 -->
[REQ-DNS-03-0008](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0008) → `M-DNS-03-NORMALIZATION` (Failure Conditions).

<!-- requirement-ref: REQ-DNS-03-0027 -->
[REQ-DNS-03-0027](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0027) → `M-DNS-03-NORMALIZATION` (Outcomes).

#### M-DNS-03-LIMITS

Intended owner: `internal/measurement/dns03/run.go`. Planned case: `internal/measurement/dns03/conformance_test.go`, `symbol:TestMDns03Limits`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Test the 65,535-byte DNS message and 16 retained answers at below/equal/above boundaries under stricter shared DNS/HTTP/TLS/task limits. Assert one post-bootstrap resolver connection/request, no retry, full-chain deadline and deduplication by resolver/query tuple.

<!-- requirement-ref: REQ-DNS-03-0023 -->
[REQ-DNS-03-0023](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0023) → `M-DNS-03-LIMITS` (Resource Bounds).

#### M-DNS-03-VERSIONS

Intended owner: `internal/registry/measurements.go`. Planned case: `internal/registry/measurement_conformance_test.go`, `symbol:TestMDns03Versions`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/result/dns03.go`.

Case plan: Select exact registered test, Measurement Result schema and test result versions. Vary unsupported/malformed/cross-test versions and schema combinations; reject before execution rather than guessing, coercing or adding another version. Registration metadata and Draft status confer no support.

<!-- requirement-ref: REQ-DNS-03-0024 -->
[REQ-DNS-03-0024](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0024) → `M-DNS-03-VERSIONS` (Output).

#### M-DNS-03-OUTPUT

Intended owner: `internal/result/dns03.go`. Planned case: `internal/result/dns03_conformance_test.go`, `symbol:TestMDns03Output`. Disposition: **planned**, with no execution or support evidence.

Case plan: Verify bootstrap DNS only when actually used, then TCP/TLS/HTTP/doh DNS in execution order, stopping at the failed stage. Use shared DNS normalization and HTTP/TLS observations without importing HTTP-01 GET semantics or synthesizing unstarted stages.

<!-- requirement-ref: REQ-DNS-03-0025 -->
[REQ-DNS-03-0025](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0025) → `M-DNS-03-OUTPUT` (Output).

<!-- requirement-ref: REQ-DNS-03-0026 -->
[REQ-DNS-03-0026](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0026) → `M-DNS-03-OUTPUT` (Output).

#### M-DNS-03-TERMINATION

Intended owner: `internal/measurement/dns03/run.go`. Planned case: `internal/measurement/dns03/conformance_test.go`, `symbol:TestMDns03Termination`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Interrupt before/during/after every operation through network-generation change, cancellation, attempt deadline, proxy loss/reconfiguration where applicable, resource bounds and suspension as applicable to each linked source. Assert the exact common execution state/reason and measurement presence/absence; keep NET-05 permitted partial profile output distinct from primitive output. Release resources and prohibit fallback or invented terminal protocol outcomes.

<!-- requirement-ref: REQ-DNS-03-0028 -->
[REQ-DNS-03-0028](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0028) → `M-DNS-03-TERMINATION` (Outcomes).

#### M-DNS-03-MINIMIZATION

Intended owner: `internal/result/dns03.go`. Planned case: `internal/result/dns03_conformance_test.go`, `symbol:TestMDns03Minimization`. Disposition: **planned**, with no execution or support evidence.

Case plan: Inject unique synthetic canaries into every prohibited target, identifier, address, DNS/TLS/HTTP/platform/error field named by each linked unit. Inspect serialized result/error paths and all nested observations, asserting no retained canary or extra field; preserve only explicitly allowed bounded normalized DNS answer values. No participant or live measurement data is used.

<!-- requirement-ref: REQ-DNS-03-0029 -->
[REQ-DNS-03-0029](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0029) → `M-DNS-03-MINIMIZATION` (Outcomes).

#### M-DNS-03-INTERPRETATION

Intended owner: `docs/measurement-interpretation-review.md`. Planned case: `docs/measurement-interpretation-review.md`, `heading:M-DNS-03-INTERPRETATION`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/measurement/dns03/run.go`, `internal/result/dns03.go`.

Case plan: Review layer-specific completion for the configured resolver/query and every stronger-evidence criterion. Reject automatic DoH-blocking claims from bootstrap/TCP/TLS/HTTP/DNS failure or resolver differences; retain independent outage, rate-limit, route, family and VPN explanations.

Scope: separately reviewed integration/interpretation evidence; portable tests cannot complete this family alone.

<!-- requirement-ref: REQ-DNS-03-0001 -->
[REQ-DNS-03-0001](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0001) → `M-DNS-03-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-DNS-03-0002 -->
[REQ-DNS-03-0002](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0002) → `M-DNS-03-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-DNS-03-0003 -->
[REQ-DNS-03-0003](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0003) → `M-DNS-03-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-DNS-03-0030 -->
[REQ-DNS-03-0030](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0030) → `M-DNS-03-INTERPRETATION` (Supported Interpretation).

<!-- requirement-ref: REQ-DNS-03-0031 -->
[REQ-DNS-03-0031](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0031) → `M-DNS-03-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

<!-- requirement-ref: REQ-DNS-03-0032 -->
[REQ-DNS-03-0032](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0032) → `M-DNS-03-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

#### M-DNS-03-SAFETY

Intended owner: `docs/measurement-deployment-review.md`. Planned case: `docs/measurement-deployment-review.md`, `heading:M-DNS-03-SAFETY`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/execution/measurements.go`, `internal/result/dns03.go`.

Case plan: Review every baseline/factor row and separately exercise each normative safety action using synthetic targets and recording transports. Keep effective risk at least the registered baseline and preserve host/local-policy qualification and final probe authority; no identifiers, unsafe target expansion or distinctive scheduling is introduced. Native capabilities, actual traffic patterns, host scheduling/consent and deployment review require separate scoped evidence; explanatory rationales are not new runtime data.

Scope: separately reviewed integration/interpretation evidence; portable tests cannot complete this family alone.

<!-- requirement-ref: REQ-DNS-03-0033 -->
[REQ-DNS-03-0033](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0033) → `M-DNS-03-SAFETY` (Baseline Risk).

<!-- requirement-ref: REQ-DNS-03-0034 -->
[REQ-DNS-03-0034](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0034) → `M-DNS-03-SAFETY` (Risk Factors).

<!-- requirement-ref: REQ-DNS-03-0035 -->
[REQ-DNS-03-0035](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0035) → `M-DNS-03-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-DNS-03-0036 -->
[REQ-DNS-03-0036](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0036) → `M-DNS-03-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-DNS-03-0037 -->
[REQ-DNS-03-0037](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0037) → `M-DNS-03-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-DNS-03-0038 -->
[REQ-DNS-03-0038](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0038) → `M-DNS-03-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-DNS-03-0039 -->
[REQ-DNS-03-0039](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0039) → `M-DNS-03-SAFETY` (Safety Analysis).

### DNS-04

#### M-DNS-04-ADMISSION

Intended owner: `internal/validation/dns04.go`. Planned case: `internal/validation/dns04_conformance_test.go`, `symbol:TestMDns04Admission`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/taskdecode/dns04.go`, `internal/admission/measurements.go`, `internal/measurement/dns04/run.go`.

Case plan: Vary registered resolver_name/port/bootstrap tuple, normalized 253-byte resolver and query names, A/AAAA and canonical public bootstrap syntax. Require reviewed support for nonstandard ports, reject unknown/local/private/interpolated targets, and assert TLS always authenticates resolver_name even with bootstrap_ip; no cleartext DNS is sent on the DoT connection.

<!-- requirement-ref: REQ-DNS-04-0004 -->
[REQ-DNS-04-0004](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0004) → `M-DNS-04-ADMISSION` (Target).

<!-- requirement-ref: REQ-DNS-04-0009 -->
[REQ-DNS-04-0009](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0009) → `M-DNS-04-ADMISSION` (Inputs).

<!-- requirement-ref: REQ-DNS-04-0010 -->
[REQ-DNS-04-0010](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0010) → `M-DNS-04-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-DNS-04-0011 -->
[REQ-DNS-04-0011](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0011) → `M-DNS-04-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-DNS-04-0012 -->
[REQ-DNS-04-0012](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0012) → `M-DNS-04-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-DNS-04-0013 -->
[REQ-DNS-04-0013](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0013) → `M-DNS-04-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-DNS-04-0014 -->
[REQ-DNS-04-0014](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0014) → `M-DNS-04-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-DNS-04-0015 -->
[REQ-DNS-04-0015](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0015) → `M-DNS-04-ADMISSION` (Input Validation).

#### M-DNS-04-OPTIONS

Intended owner: `internal/validation/dns04.go`. Planned case: `internal/validation/dns04_conformance_test.go`, `symbol:TestMDns04Options`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/taskdecode/dns04.go`, `internal/admission/measurements.go`, `internal/measurement/dns04/run.go`.

Case plan: Exercise auto/ipv4/ipv6 with bootstrap compatibility, omitted default and invalid/unknown cases. Assert no opportunistic authentication, cleartext fallback, cross-run connection reuse or alternate resolver fallback.

<!-- requirement-ref: REQ-DNS-04-0016 -->
[REQ-DNS-04-0016](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0016) → `M-DNS-04-OPTIONS` (Options).

<!-- requirement-ref: REQ-DNS-04-0017 -->
[REQ-DNS-04-0017](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0017) → `M-DNS-04-OPTIONS` (Options).

#### M-DNS-04-PRECONDITIONS

Intended owner: `internal/measurement/dns04/run.go`. Planned case: `internal/measurement/dns04/conformance_test.go`, `symbol:TestMDns04Preconditions`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/platform/network.go`, `docs/measurement-platform-validation.md`.

Case plan: Vary technical/privacy/safety review, local registered profile, explicit host/anomaly policy, stable generation and deadline. No resolver contact before all applicable gates pass; native/deployment policy still needs scoped review.

Scope: portable gate tests plus separately required native/host validation; no platform claim.

<!-- requirement-ref: REQ-DNS-04-0018 -->
[REQ-DNS-04-0018](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0018) → `M-DNS-04-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-DNS-04-0019 -->
[REQ-DNS-04-0019](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0019) → `M-DNS-04-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-DNS-04-0020 -->
[REQ-DNS-04-0020](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0020) → `M-DNS-04-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-DNS-04-0021 -->
[REQ-DNS-04-0021](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0021) → `M-DNS-04-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-DNS-04-0022 -->
[REQ-DNS-04-0022](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0022) → `M-DNS-04-PRECONDITIONS` (Preconditions).

#### M-DNS-04-EXECUTION

Intended owner: `internal/measurement/dns04/run.go`. Planned case: `internal/measurement/dns04/conformance_test.go`, `symbol:TestMDns04Execution`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Record optional system bootstrap, one selected TCP connection, authenticated TLS, one length-prefixed DNS query and bounded matching response, then close. Fault the prefix, message length, query match and each stage; require no cleartext or resolver fallback and no application retry.

<!-- requirement-ref: REQ-DNS-04-0005 -->
[REQ-DNS-04-0005](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0005) → `M-DNS-04-EXECUTION` (Operation).

<!-- requirement-ref: REQ-DNS-04-0006 -->
[REQ-DNS-04-0006](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0006) → `M-DNS-04-EXECUTION` (Procedure).

#### M-DNS-04-NORMALIZATION

Intended owner: `internal/measurement/dns04/run.go`. Planned case: `internal/measurement/dns04/conformance_test.go`, `symbol:TestMDns04Normalization`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/result/dns04.go`.

Case plan: Exercise every outcome row, including TCP reset before TLS versus reset during TLS, authenticated negative DNS codes, malformed prefix/message, mismatch, post-connect timeout and inconclusive. Keep terminal failure at the observed stage with prior observations ordered.

<!-- requirement-ref: REQ-DNS-04-0007 -->
[REQ-DNS-04-0007](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0007) → `M-DNS-04-NORMALIZATION` (Success Condition).

<!-- requirement-ref: REQ-DNS-04-0008 -->
[REQ-DNS-04-0008](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0008) → `M-DNS-04-NORMALIZATION` (Failure Conditions).

<!-- requirement-ref: REQ-DNS-04-0027 -->
[REQ-DNS-04-0027](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0027) → `M-DNS-04-NORMALIZATION` (Outcomes).

#### M-DNS-04-LIMITS

Intended owner: `internal/measurement/dns04/run.go`. Planned case: `internal/measurement/dns04/conformance_test.go`, `symbol:TestMDns04Limits`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Test the 65,535-byte DNS message and 16-answer limit at below/equal/above boundaries under stricter common limits. Assert one resolver connection/DNS exchange after optional bootstrap, bounded chain deadline, no retries and deduplication by resolver/query tuple.

<!-- requirement-ref: REQ-DNS-04-0023 -->
[REQ-DNS-04-0023](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0023) → `M-DNS-04-LIMITS` (Resource Bounds).

#### M-DNS-04-VERSIONS

Intended owner: `internal/registry/measurements.go`. Planned case: `internal/registry/measurement_conformance_test.go`, `symbol:TestMDns04Versions`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/result/dns04.go`.

Case plan: Select exact registered test, Measurement Result schema and test result versions. Vary unsupported/malformed/cross-test versions and schema combinations; reject before execution rather than guessing, coercing or adding another version. Registration metadata and Draft status confer no support.

<!-- requirement-ref: REQ-DNS-04-0024 -->
[REQ-DNS-04-0024](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0024) → `M-DNS-04-VERSIONS` (Output).

#### M-DNS-04-OUTPUT

Intended owner: `internal/result/dns04.go`. Planned case: `internal/result/dns04_conformance_test.go`, `symbol:TestMDns04Output`. Disposition: **planned**, with no execution or support evidence.

Case plan: Verify optional bootstrap DNS, TCP, TLS, then dot DNS; omit bootstrap when an approved bootstrap_ip is supplied and omit unstarted stages after failure. Reuse the common DNS answer normalization and bounds without retaining framing bytes.

<!-- requirement-ref: REQ-DNS-04-0025 -->
[REQ-DNS-04-0025](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0025) → `M-DNS-04-OUTPUT` (Output).

<!-- requirement-ref: REQ-DNS-04-0026 -->
[REQ-DNS-04-0026](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0026) → `M-DNS-04-OUTPUT` (Output).

#### M-DNS-04-TERMINATION

Intended owner: `internal/measurement/dns04/run.go`. Planned case: `internal/measurement/dns04/conformance_test.go`, `symbol:TestMDns04Termination`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Interrupt before/during/after every operation through network-generation change, cancellation, attempt deadline, proxy loss/reconfiguration where applicable, resource bounds and suspension as applicable to each linked source. Assert the exact common execution state/reason and measurement presence/absence; keep NET-05 permitted partial profile output distinct from primitive output. Release resources and prohibit fallback or invented terminal protocol outcomes.

<!-- requirement-ref: REQ-DNS-04-0028 -->
[REQ-DNS-04-0028](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0028) → `M-DNS-04-TERMINATION` (Outcomes).

#### M-DNS-04-MINIMIZATION

Intended owner: `internal/result/dns04.go`. Planned case: `internal/result/dns04_conformance_test.go`, `symbol:TestMDns04Minimization`. Disposition: **planned**, with no execution or support evidence.

Case plan: Inject unique synthetic canaries into every prohibited target, identifier, address, DNS/TLS/HTTP/platform/error field named by each linked unit. Inspect serialized result/error paths and all nested observations, asserting no retained canary or extra field; preserve only explicitly allowed bounded normalized DNS answer values. No participant or live measurement data is used.

<!-- requirement-ref: REQ-DNS-04-0029 -->
[REQ-DNS-04-0029](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0029) → `M-DNS-04-MINIMIZATION` (Outcomes).

#### M-DNS-04-INTERPRETATION

Intended owner: `docs/measurement-interpretation-review.md`. Planned case: `docs/measurement-interpretation-review.md`, `heading:M-DNS-04-INTERPRETATION`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/measurement/dns04/run.go`, `internal/result/dns04.go`.

Case plan: Review each configured stage and the stated stronger-evidence criteria; transport differences do not establish censor intent. Include resolver/certificate/bootstrap/port-policy/routing/time/family failures and independent reference context.

Scope: separately reviewed integration/interpretation evidence; portable tests cannot complete this family alone.

<!-- requirement-ref: REQ-DNS-04-0001 -->
[REQ-DNS-04-0001](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0001) → `M-DNS-04-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-DNS-04-0002 -->
[REQ-DNS-04-0002](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0002) → `M-DNS-04-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-DNS-04-0003 -->
[REQ-DNS-04-0003](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0003) → `M-DNS-04-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-DNS-04-0030 -->
[REQ-DNS-04-0030](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0030) → `M-DNS-04-INTERPRETATION` (Supported Interpretation).

<!-- requirement-ref: REQ-DNS-04-0031 -->
[REQ-DNS-04-0031](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0031) → `M-DNS-04-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

<!-- requirement-ref: REQ-DNS-04-0032 -->
[REQ-DNS-04-0032](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0032) → `M-DNS-04-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

#### M-DNS-04-SAFETY

Intended owner: `docs/measurement-deployment-review.md`. Planned case: `docs/measurement-deployment-review.md`, `heading:M-DNS-04-SAFETY`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/execution/measurements.go`, `internal/result/dns04.go`.

Case plan: Review every baseline/factor row and separately exercise each normative safety action using synthetic targets and recording transports. Keep effective risk at least the registered baseline and preserve host/local-policy qualification and final probe authority; no identifiers, unsafe target expansion or distinctive scheduling is introduced. Native capabilities, actual traffic patterns, host scheduling/consent and deployment review require separate scoped evidence; explanatory rationales are not new runtime data.

Scope: separately reviewed integration/interpretation evidence; portable tests cannot complete this family alone.

<!-- requirement-ref: REQ-DNS-04-0033 -->
[REQ-DNS-04-0033](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0033) → `M-DNS-04-SAFETY` (Baseline Risk).

<!-- requirement-ref: REQ-DNS-04-0034 -->
[REQ-DNS-04-0034](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0034) → `M-DNS-04-SAFETY` (Risk Factors).

<!-- requirement-ref: REQ-DNS-04-0035 -->
[REQ-DNS-04-0035](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0035) → `M-DNS-04-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-DNS-04-0036 -->
[REQ-DNS-04-0036](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0036) → `M-DNS-04-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-DNS-04-0037 -->
[REQ-DNS-04-0037](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0037) → `M-DNS-04-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-DNS-04-0038 -->
[REQ-DNS-04-0038](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0038) → `M-DNS-04-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-DNS-04-0039 -->
[REQ-DNS-04-0039](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0039) → `M-DNS-04-SAFETY` (Safety Analysis).

### TCP-01

#### M-TCP-01-ADMISSION

Intended owner: `internal/validation/tcp01.go`. Planned case: `internal/validation/tcp01_conformance_test.go`, `symbol:TestMTcp01Admission`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/taskdecode/tcp01.go`, `internal/admission/measurements.go`, `internal/measurement/tcp01/run.go`.

Case plan: Vary canonical public IPv4/IPv6, ports 1/65535 and outside bounds, exact local registration, prohibited/special-purpose ranges, hostnames, arrays, port ranges, interpolation and unknown fields. Preserve REQ-TCP-01-0001: rejection starts no socket and produces no network-interference Measurement Result.

<!-- requirement-ref: REQ-TCP-01-0005 -->
[REQ-TCP-01-0005](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0005) → `M-TCP-01-ADMISSION` (Target).

<!-- requirement-ref: REQ-TCP-01-0010 -->
[REQ-TCP-01-0010](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0010) → `M-TCP-01-ADMISSION` (Inputs).

<!-- requirement-ref: REQ-TCP-01-0011 -->
[REQ-TCP-01-0011](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0011) → `M-TCP-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-TCP-01-0012 -->
[REQ-TCP-01-0012](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0012) → `M-TCP-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-TCP-01-0013 -->
[REQ-TCP-01-0013](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0013) → `M-TCP-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-TCP-01-0014 -->
[REQ-TCP-01-0014](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0014) → `M-TCP-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-TCP-01-0001 -->
[REQ-TCP-01-0001](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0001) → `M-TCP-01-ADMISSION` (Input Validation).

#### M-TCP-01-OPTIONS

Intended owner: `internal/validation/tcp01.go`. Planned case: `internal/validation/tcp01_conformance_test.go`, `symbol:TestMTcp01Options`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/taskdecode/tcp01.go`, `internal/admission/measurements.go`, `internal/measurement/tcp01/run.go`.

Case plan: Accept empty options only; reject extras and wrong types. Derive the family solely from the validated explicit IP address.

<!-- requirement-ref: REQ-TCP-01-0015 -->
[REQ-TCP-01-0015](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0015) → `M-TCP-01-OPTIONS` (Options).

<!-- requirement-ref: REQ-TCP-01-0016 -->
[REQ-TCP-01-0016](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0016) → `M-TCP-01-OPTIONS` (Options).

#### M-TCP-01-PRECONDITIONS

Intended owner: `internal/measurement/tcp01/run.go`. Planned case: `internal/measurement/tcp01/conformance_test.go`, `symbol:TestMTcp01Preconditions`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/platform/network.go`, `docs/measurement-platform-validation.md`.

Case plan: Vary family support, registered address/port approval, stable generation/deadline and intended path capability. Reject before contact when routing would use an unintended path, including stricter host/task policy.

Scope: portable gate tests plus separately required native/host validation; no platform claim.

<!-- requirement-ref: REQ-TCP-01-0017 -->
[REQ-TCP-01-0017](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0017) → `M-TCP-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-TCP-01-0018 -->
[REQ-TCP-01-0018](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0018) → `M-TCP-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-TCP-01-0019 -->
[REQ-TCP-01-0019](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0019) → `M-TCP-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-TCP-01-0020 -->
[REQ-TCP-01-0020](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0020) → `M-TCP-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-TCP-01-0021 -->
[REQ-TCP-01-0021](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0021) → `M-TCP-01-PRECONDITIONS` (Preconditions).

#### M-TCP-01-EXECUTION

Intended owner: `internal/measurement/tcp01/run.go`. Planned case: `internal/measurement/tcp01/conformance_test.go`, `symbol:TestMTcp01Execution`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Record one socket connect to the exact accepted address/port and immediate close without application bytes. No DNS lookup, alternate address/port, retry or additional protocol may run.

<!-- requirement-ref: REQ-TCP-01-0006 -->
[REQ-TCP-01-0006](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0006) → `M-TCP-01-EXECUTION` (Operation).

<!-- requirement-ref: REQ-TCP-01-0007 -->
[REQ-TCP-01-0007](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0007) → `M-TCP-01-EXECUTION` (Procedure).

#### M-TCP-01-NORMALIZATION

Intended owner: `internal/measurement/tcp01/run.go`. Planned case: `internal/measurement/tcp01/conformance_test.go`, `symbol:TestMTcp01Normalization`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/result/tcp01.go`.

Case plan: Exercise every failure/outcome row, including refusal/reset/timeout/unreachable/permission and coarse inconclusive errors, using deterministic socket fixtures. Distinguish socket outcomes from common interruption and map to exactly one tcp_connect.

<!-- requirement-ref: REQ-TCP-01-0008 -->
[REQ-TCP-01-0008](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0008) → `M-TCP-01-NORMALIZATION` (Success Condition).

<!-- requirement-ref: REQ-TCP-01-0009 -->
[REQ-TCP-01-0009](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0009) → `M-TCP-01-NORMALIZATION` (Failure Conditions).

<!-- requirement-ref: REQ-TCP-01-0025 -->
[REQ-TCP-01-0025](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0025) → `M-TCP-01-NORMALIZATION` (Outcomes).

#### M-TCP-01-LIMITS

Intended owner: `internal/measurement/tcp01/run.go`. Planned case: `internal/measurement/tcp01/conformance_test.go`, `symbol:TestMTcp01Limits`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Advance the operation and outer deadlines independently; assert one connection, handshake traffic only, no retries and deduplication by address/port tuple. Verify socket release on all failure/cancellation paths.

<!-- requirement-ref: REQ-TCP-01-0022 -->
[REQ-TCP-01-0022](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0022) → `M-TCP-01-LIMITS` (Resource Bounds).

#### M-TCP-01-VERSIONS

Intended owner: `internal/registry/measurements.go`. Planned case: `internal/registry/measurement_conformance_test.go`, `symbol:TestMTcp01Versions`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/result/tcp01.go`.

Case plan: Select exact registered test, Measurement Result schema and test result versions. Vary unsupported/malformed/cross-test versions and schema combinations; reject before execution rather than guessing, coercing or adding another version. Registration metadata and Draft status confer no support.

<!-- requirement-ref: REQ-TCP-01-0023 -->
[REQ-TCP-01-0023](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0023) → `M-TCP-01-VERSIONS` (Output).

#### M-TCP-01-OUTPUT

Intended owner: `internal/result/tcp01.go`. Planned case: `internal/result/tcp01_conformance_test.go`, `symbol:TestMTcp01Output`. Disposition: **planned**, with no execution or support evidence.

Case plan: Require exactly one tcp_connect with the selected family, normalized outcome and quantized duration. Reuse shared TCP observation validation; resolve the target through task identity rather than duplicating destination fields.

<!-- requirement-ref: REQ-TCP-01-0024 -->
[REQ-TCP-01-0024](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0024) → `M-TCP-01-OUTPUT` (Output).

#### M-TCP-01-TERMINATION

Intended owner: `internal/measurement/tcp01/run.go`. Planned case: `internal/measurement/tcp01/conformance_test.go`, `symbol:TestMTcp01Termination`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Interrupt before/during/after every operation through network-generation change, cancellation, attempt deadline, proxy loss/reconfiguration where applicable, resource bounds and suspension as applicable to each linked source. Assert the exact common execution state/reason and measurement presence/absence; keep NET-05 permitted partial profile output distinct from primitive output. Release resources and prohibit fallback or invented terminal protocol outcomes.

<!-- requirement-ref: REQ-TCP-01-0026 -->
[REQ-TCP-01-0026](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0026) → `M-TCP-01-TERMINATION` (Outcomes).

#### M-TCP-01-MINIMIZATION

Intended owner: `internal/result/tcp01.go`. Planned case: `internal/result/tcp01_conformance_test.go`, `symbol:TestMTcp01Minimization`. Disposition: **planned**, with no execution or support evidence.

Case plan: Inject unique synthetic canaries into every prohibited target, identifier, address, DNS/TLS/HTTP/platform/error field named by each linked unit. Inspect serialized result/error paths and all nested observations, asserting no retained canary or extra field; preserve only explicitly allowed bounded normalized DNS answer values. No participant or live measurement data is used.

<!-- requirement-ref: REQ-TCP-01-0027 -->
[REQ-TCP-01-0027](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0027) → `M-TCP-01-MINIMIZATION` (Outcomes).

#### M-TCP-01-INTERPRETATION

Intended owner: `docs/measurement-interpretation-review.md`. Planned case: `docs/measurement-interpretation-review.md`, `heading:M-TCP-01-INTERPRETATION`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/measurement/tcp01/run.go`, `internal/result/tcp01.go`.

Case plan: Review the limit to one configured endpoint connect observation, the stronger-evidence list and service/firewall/routing/family/transient explanations. A connection refusal/reset/timeout is never a definitive censorship finding or proof of higher-layer usability.

Scope: separately reviewed integration/interpretation evidence; portable tests cannot complete this family alone.

<!-- requirement-ref: REQ-TCP-01-0002 -->
[REQ-TCP-01-0002](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0002) → `M-TCP-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-TCP-01-0003 -->
[REQ-TCP-01-0003](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0003) → `M-TCP-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-TCP-01-0004 -->
[REQ-TCP-01-0004](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0004) → `M-TCP-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-TCP-01-0028 -->
[REQ-TCP-01-0028](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0028) → `M-TCP-01-INTERPRETATION` (Supported Interpretation).

<!-- requirement-ref: REQ-TCP-01-0029 -->
[REQ-TCP-01-0029](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0029) → `M-TCP-01-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

<!-- requirement-ref: REQ-TCP-01-0030 -->
[REQ-TCP-01-0030](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0030) → `M-TCP-01-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

#### M-TCP-01-SAFETY

Intended owner: `docs/measurement-deployment-review.md`. Planned case: `docs/measurement-deployment-review.md`, `heading:M-TCP-01-SAFETY`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/execution/measurements.go`, `internal/result/tcp01.go`.

Case plan: Review every baseline/factor row and separately exercise each normative safety action using synthetic targets and recording transports. Keep effective risk at least the registered baseline and preserve host/local-policy qualification and final probe authority; no identifiers, unsafe target expansion or distinctive scheduling is introduced. Native capabilities, actual traffic patterns, host scheduling/consent and deployment review require separate scoped evidence; explanatory rationales are not new runtime data.

Scope: separately reviewed integration/interpretation evidence; portable tests cannot complete this family alone.

<!-- requirement-ref: REQ-TCP-01-0031 -->
[REQ-TCP-01-0031](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0031) → `M-TCP-01-SAFETY` (Baseline Risk).

<!-- requirement-ref: REQ-TCP-01-0032 -->
[REQ-TCP-01-0032](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0032) → `M-TCP-01-SAFETY` (Risk Factors).

<!-- requirement-ref: REQ-TCP-01-0033 -->
[REQ-TCP-01-0033](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0033) → `M-TCP-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-TCP-01-0034 -->
[REQ-TCP-01-0034](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0034) → `M-TCP-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-TCP-01-0035 -->
[REQ-TCP-01-0035](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0035) → `M-TCP-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-TCP-01-0036 -->
[REQ-TCP-01-0036](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0036) → `M-TCP-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-TCP-01-0037 -->
[REQ-TCP-01-0037](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0037) → `M-TCP-01-SAFETY` (Safety Analysis).

### TLS-01

#### M-TLS-01-ADMISSION

Intended owner: `internal/validation/tls01.go`. Planned case: `internal/validation/tls01_conformance_test.go`, `symbol:TestMTls01Admission`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/taskdecode/tls01.go`, `internal/admission/measurements.go`, `internal/measurement/tls01/run.go`.

Case plan: Vary canonical public IP, port, deterministic normalized DNS server_name through 253 bytes, exact reviewed tuple and unknown fields. Reject literal/wildcard/derived SNI, private/special/interpolated targets and every certificate/hostname-verification bypass before socket activity.

<!-- requirement-ref: REQ-TLS-01-0004 -->
[REQ-TLS-01-0004](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0004) → `M-TLS-01-ADMISSION` (Target).

<!-- requirement-ref: REQ-TLS-01-0009 -->
[REQ-TLS-01-0009](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0009) → `M-TLS-01-ADMISSION` (Inputs).

<!-- requirement-ref: REQ-TLS-01-0010 -->
[REQ-TLS-01-0010](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0010) → `M-TLS-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-TLS-01-0011 -->
[REQ-TLS-01-0011](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0011) → `M-TLS-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-TLS-01-0012 -->
[REQ-TLS-01-0012](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0012) → `M-TLS-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-TLS-01-0013 -->
[REQ-TLS-01-0013](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0013) → `M-TLS-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-TLS-01-0014 -->
[REQ-TLS-01-0014](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0014) → `M-TLS-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-TLS-01-0015 -->
[REQ-TLS-01-0015](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0015) → `M-TLS-01-ADMISSION` (Input Validation).

#### M-TLS-01-OPTIONS

Intended owner: `internal/validation/tls01.go`. Planned case: `internal/validation/tls01_conformance_test.go`, `symbol:TestMTls01Options`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/taskdecode/tls01.go`, `internal/admission/measurements.go`, `internal/measurement/tls01/run.go`.

Case plan: Exercise omitted/empty ALPN, 1..8 identifiers, 1..32 ASCII bytes, ordering, registered values and duplicates at all boundaries. Reject custom cipher/version profiles, invalid-certificate controls, alternate SNI, client certificates, resumption and insecure verification.

<!-- requirement-ref: REQ-TLS-01-0016 -->
[REQ-TLS-01-0016](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0016) → `M-TLS-01-OPTIONS` (Options).

<!-- requirement-ref: REQ-TLS-01-0017 -->
[REQ-TLS-01-0017](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0017) → `M-TLS-01-OPTIONS` (Options).

#### M-TLS-01-PRECONDITIONS

Intended owner: `internal/measurement/tls01/run.go`. Planned case: `internal/measurement/tls01/conformance_test.go`, `symbol:TestMTls01Preconditions`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/platform/network.go`, `docs/measurement-platform-validation.md`.

Case plan: Vary stable generation/deadline, normal publicly authenticated target, documented trust policy/time source and forbidden client authentication/credentials/sensitive names. Review trust-store and time-source behavior separately for native platforms.

Scope: portable gate tests plus separately required native/host validation; no platform claim.

<!-- requirement-ref: REQ-TLS-01-0018 -->
[REQ-TLS-01-0018](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0018) → `M-TLS-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-TLS-01-0019 -->
[REQ-TLS-01-0019](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0019) → `M-TLS-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-TLS-01-0020 -->
[REQ-TLS-01-0020](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0020) → `M-TLS-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-TLS-01-0021 -->
[REQ-TLS-01-0021](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0021) → `M-TLS-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-TLS-01-0022 -->
[REQ-TLS-01-0022](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0022) → `M-TLS-01-PRECONDITIONS` (Preconditions).

#### M-TLS-01-EXECUTION

Intended owner: `internal/measurement/tls01/run.go`. Planned case: `internal/measurement/tls01/conformance_test.go`, `symbol:TestMTls01Execution`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Record explicit-IP TCP then one normal SNI/ALPN handshake with certificate/hostname validation; retain first failed stage or minimized negotiated properties and close before application data. No implicit DNS, alternate profiles, SNI comparison, sessions or retries.

<!-- requirement-ref: REQ-TLS-01-0005 -->
[REQ-TLS-01-0005](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0005) → `M-TLS-01-EXECUTION` (Operation).

<!-- requirement-ref: REQ-TLS-01-0006 -->
[REQ-TLS-01-0006](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0006) → `M-TLS-01-EXECUTION` (Procedure).

#### M-TLS-01-NORMALIZATION

Intended owner: `internal/measurement/tls01/run.go`. Planned case: `internal/measurement/tls01/conformance_test.go`, `symbol:TestMTls01Normalization`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/result/tls01.go`.

Case plan: Exercise every outcome row, including TCP versus TLS reset, certificate trust/name/time failures, alerts, protocol error, ALPN no-overlap, per-stage timeout and coarse inconclusive. Inspect the required certificate_validation value and stage-specific failure rather than collapsing failures into filtering.

<!-- requirement-ref: REQ-TLS-01-0007 -->
[REQ-TLS-01-0007](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0007) → `M-TLS-01-NORMALIZATION` (Success Condition).

<!-- requirement-ref: REQ-TLS-01-0008 -->
[REQ-TLS-01-0008](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0008) → `M-TLS-01-NORMALIZATION` (Failure Conditions).

<!-- requirement-ref: REQ-TLS-01-0026 -->
[REQ-TLS-01-0026](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0026) → `M-TLS-01-NORMALIZATION` (Outcomes).

#### M-TLS-01-LIMITS

Intended owner: `internal/measurement/tls01/run.go`. Planned case: `internal/measurement/tls01/conformance_test.go`, `symbol:TestMTls01Limits`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Assert one connection/handshake, full-chain deadline, no application payload/retries and deduplication by endpoint/server-name tuple. Test cancellation and close while the peer stalls in handshake.

<!-- requirement-ref: REQ-TLS-01-0023 -->
[REQ-TLS-01-0023](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0023) → `M-TLS-01-LIMITS` (Resource Bounds).

#### M-TLS-01-VERSIONS

Intended owner: `internal/registry/measurements.go`. Planned case: `internal/registry/measurement_conformance_test.go`, `symbol:TestMTls01Versions`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/result/tls01.go`.

Case plan: Select exact registered test, Measurement Result schema and test result versions. Vary unsupported/malformed/cross-test versions and schema combinations; reject before execution rather than guessing, coercing or adding another version. Registration metadata and Draft status confer no support.

<!-- requirement-ref: REQ-TLS-01-0024 -->
[REQ-TLS-01-0024](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0024) → `M-TLS-01-VERSIONS` (Output).

#### M-TLS-01-OUTPUT

Intended owner: `internal/result/tls01.go`. Planned case: `internal/result/tls01_conformance_test.go`, `symbol:TestMTls01Output`. Disposition: **planned**, with no execution or support evidence.

Case plan: Require TCP then TLS only when handshake activity began, in execution order. Check minimized allowed negotiated properties against common TCP/TLS observations with certificate-validation classes preserved.

<!-- requirement-ref: REQ-TLS-01-0025 -->
[REQ-TLS-01-0025](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0025) → `M-TLS-01-OUTPUT` (Output).

#### M-TLS-01-TERMINATION

Intended owner: `internal/measurement/tls01/run.go`. Planned case: `internal/measurement/tls01/conformance_test.go`, `symbol:TestMTls01Termination`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Interrupt before/during/after every operation through network-generation change, cancellation, attempt deadline, proxy loss/reconfiguration where applicable, resource bounds and suspension as applicable to each linked source. Assert the exact common execution state/reason and measurement presence/absence; keep NET-05 permitted partial profile output distinct from primitive output. Release resources and prohibit fallback or invented terminal protocol outcomes.

<!-- requirement-ref: REQ-TLS-01-0027 -->
[REQ-TLS-01-0027](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0027) → `M-TLS-01-TERMINATION` (Outcomes).

#### M-TLS-01-MINIMIZATION

Intended owner: `internal/result/tls01.go`. Planned case: `internal/result/tls01_conformance_test.go`, `symbol:TestMTls01Minimization`. Disposition: **planned**, with no execution or support evidence.

Case plan: Inject unique synthetic canaries into every prohibited target, identifier, address, DNS/TLS/HTTP/platform/error field named by each linked unit. Inspect serialized result/error paths and all nested observations, asserting no retained canary or extra field; preserve only explicitly allowed bounded normalized DNS answer values. No participant or live measurement data is used.

<!-- requirement-ref: REQ-TLS-01-0028 -->
[REQ-TLS-01-0028](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0028) → `M-TLS-01-MINIMIZATION` (Outcomes).

#### M-TLS-01-INTERPRETATION

Intended owner: `docs/measurement-interpretation-review.md`. Planned case: `docs/measurement-interpretation-review.md`, `heading:M-TLS-01-INTERPRETATION`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/measurement/tls01/run.go`, `internal/result/tls01.go`.

Case plan: Review only authenticated completion for the specified endpoint/name/ALPN, with controls and repeated/independent evidence before stronger interpretation. Different-SNI comparison requires separately reviewed behavior; certificate/time/route/CDN/ALPN/family failures do not establish TLS/SNI filtering.

Scope: separately reviewed integration/interpretation evidence; portable tests cannot complete this family alone.

<!-- requirement-ref: REQ-TLS-01-0001 -->
[REQ-TLS-01-0001](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0001) → `M-TLS-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-TLS-01-0002 -->
[REQ-TLS-01-0002](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0002) → `M-TLS-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-TLS-01-0003 -->
[REQ-TLS-01-0003](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0003) → `M-TLS-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-TLS-01-0029 -->
[REQ-TLS-01-0029](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0029) → `M-TLS-01-INTERPRETATION` (Supported Interpretation).

<!-- requirement-ref: REQ-TLS-01-0030 -->
[REQ-TLS-01-0030](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0030) → `M-TLS-01-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

<!-- requirement-ref: REQ-TLS-01-0031 -->
[REQ-TLS-01-0031](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0031) → `M-TLS-01-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

#### M-TLS-01-SAFETY

Intended owner: `docs/measurement-deployment-review.md`. Planned case: `docs/measurement-deployment-review.md`, `heading:M-TLS-01-SAFETY`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/execution/measurements.go`, `internal/result/tls01.go`.

Case plan: Review every baseline/factor row and separately exercise each normative safety action using synthetic targets and recording transports. Keep effective risk at least the registered baseline and preserve host/local-policy qualification and final probe authority; no identifiers, unsafe target expansion or distinctive scheduling is introduced. Native capabilities, actual traffic patterns, host scheduling/consent and deployment review require separate scoped evidence; explanatory rationales are not new runtime data.

Scope: separately reviewed integration/interpretation evidence; portable tests cannot complete this family alone.

<!-- requirement-ref: REQ-TLS-01-0032 -->
[REQ-TLS-01-0032](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0032) → `M-TLS-01-SAFETY` (Baseline Risk).

<!-- requirement-ref: REQ-TLS-01-0033 -->
[REQ-TLS-01-0033](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0033) → `M-TLS-01-SAFETY` (Risk Factors).

<!-- requirement-ref: REQ-TLS-01-0034 -->
[REQ-TLS-01-0034](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0034) → `M-TLS-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-TLS-01-0035 -->
[REQ-TLS-01-0035](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0035) → `M-TLS-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-TLS-01-0036 -->
[REQ-TLS-01-0036](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0036) → `M-TLS-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-TLS-01-0037 -->
[REQ-TLS-01-0037](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0037) → `M-TLS-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-TLS-01-0038 -->
[REQ-TLS-01-0038](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0038) → `M-TLS-01-SAFETY` (Safety Analysis).

### HTTP-01

#### M-HTTP-01-ADMISSION

Intended owner: `internal/validation/http01.go`. Planned case: `internal/validation/http01_conformance_test.go`, `symbol:TestMHttp01Admission`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/taskdecode/http01.go`, `internal/admission/measurements.go`, `internal/measurement/http01/run.go`.

Case plan: Vary HTTPS URL/2,048-byte bounds, optional status, body/header arrays 0..8 and value bounds through 256 UTF-8 bytes; check all nested types/enums/conditional values, SHA-256 syntax and unknown fields. Exercise case-sensitive body operations, case-insensitive header names, multivalue any-match and all-checks conjunction. Reject sensitive headers, unsafe URLs/addresses and unregistered redirect targets before contact.

<!-- requirement-ref: REQ-HTTP-01-0004 -->
[REQ-HTTP-01-0004](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0004) → `M-HTTP-01-ADMISSION` (Target).

<!-- requirement-ref: REQ-HTTP-01-0010 -->
[REQ-HTTP-01-0010](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0010) → `M-HTTP-01-ADMISSION` (Inputs).

<!-- requirement-ref: REQ-HTTP-01-0011 -->
[REQ-HTTP-01-0011](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0011) → `M-HTTP-01-ADMISSION` (Inputs).

<!-- requirement-ref: REQ-HTTP-01-0012 -->
[REQ-HTTP-01-0012](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0012) → `M-HTTP-01-ADMISSION` (Inputs).

<!-- requirement-ref: REQ-HTTP-01-0013 -->
[REQ-HTTP-01-0013](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0013) → `M-HTTP-01-ADMISSION` (Inputs).

<!-- requirement-ref: REQ-HTTP-01-0014 -->
[REQ-HTTP-01-0014](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0014) → `M-HTTP-01-ADMISSION` (Inputs).

<!-- requirement-ref: REQ-HTTP-01-0015 -->
[REQ-HTTP-01-0015](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0015) → `M-HTTP-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-HTTP-01-0016 -->
[REQ-HTTP-01-0016](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0016) → `M-HTTP-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-HTTP-01-0017 -->
[REQ-HTTP-01-0017](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0017) → `M-HTTP-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-HTTP-01-0018 -->
[REQ-HTTP-01-0018](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0018) → `M-HTTP-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-HTTP-01-0019 -->
[REQ-HTTP-01-0019](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0019) → `M-HTTP-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-HTTP-01-0020 -->
[REQ-HTTP-01-0020](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0020) → `M-HTTP-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-HTTP-01-0021 -->
[REQ-HTTP-01-0021](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0021) → `M-HTTP-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-HTTP-01-0022 -->
[REQ-HTTP-01-0022](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0022) → `M-HTTP-01-ADMISSION` (Input Validation).

<!-- requirement-ref: REQ-HTTP-01-0023 -->
[REQ-HTTP-01-0023](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0023) → `M-HTTP-01-ADMISSION` (Input Validation).

#### M-HTTP-01-OPTIONS

Intended owner: `internal/validation/http01.go`. Planned case: `internal/validation/http01_conformance_test.go`, `symbol:TestMHttp01Options`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/taskdecode/http01.go`, `internal/admission/measurements.go`, `internal/measurement/http01/run.go`.

Case plan: Test auto/ipv4/ipv6, max_body_bytes default 4096 and 1..65536 with stricter task bounds, follow_redirects default false and conditional max_redirects 1/2. Reject max_redirects when disabled and every custom method/header/body/cookie/credential/proxy/compression/TLS override.

<!-- requirement-ref: REQ-HTTP-01-0024 -->
[REQ-HTTP-01-0024](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0024) → `M-HTTP-01-OPTIONS` (Options).

<!-- requirement-ref: REQ-HTTP-01-0025 -->
[REQ-HTTP-01-0025](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0025) → `M-HTTP-01-OPTIONS` (Options).

#### M-HTTP-01-PRECONDITIONS

Intended owner: `internal/measurement/http01/run.go`. Planned case: `internal/measurement/http01/conformance_test.go`, `symbol:TestMHttp01Preconditions`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/platform/network.go`, `docs/measurement-platform-validation.md`.

Case plan: Vary current local registration for URL, expected properties and every permitted redirect/check, stable generation/deadline and suitability for bounded mobile response. Reject stacks that cannot suppress credentials/cookies/identifying defaults; review platform defaults separately.

Scope: portable gate tests plus separately required native/host validation; no platform claim.

<!-- requirement-ref: REQ-HTTP-01-0026 -->
[REQ-HTTP-01-0026](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0026) → `M-HTTP-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-HTTP-01-0027 -->
[REQ-HTTP-01-0027](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0027) → `M-HTTP-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-HTTP-01-0028 -->
[REQ-HTTP-01-0028](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0028) → `M-HTTP-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-HTTP-01-0029 -->
[REQ-HTTP-01-0029](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0029) → `M-HTTP-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-HTTP-01-0030 -->
[REQ-HTTP-01-0030](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0030) → `M-HTTP-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-HTTP-01-0031 -->
[REQ-HTTP-01-0031](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0031) → `M-HTTP-01-PRECONDITIONS` (Preconditions).

#### M-HTTP-01-EXECUTION

Intended owner: `internal/measurement/http01/run.go`. Planned case: `internal/measurement/http01/conformance_test.go`, `symbol:TestMHttp01Execution`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Record system DNS, TCP, authenticated TLS, ordinary bodyless credential-free GET requesting identity encoding, bounded redirect decision and terminal checks/body close. Validate each next target and resolved address before contact; no downgrade, unsafe redirects, excessive redirect contact or test-layer retries.

<!-- requirement-ref: REQ-HTTP-01-0005 -->
[REQ-HTTP-01-0005](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0005) → `M-HTTP-01-EXECUTION` (Operation).

<!-- requirement-ref: REQ-HTTP-01-0006 -->
[REQ-HTTP-01-0006](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0006) → `M-HTTP-01-EXECUTION` (Operation).

<!-- requirement-ref: REQ-HTTP-01-0007 -->
[REQ-HTTP-01-0007](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0007) → `M-HTTP-01-EXECUTION` (Procedure).

#### M-HTTP-01-NORMALIZATION

Intended owner: `internal/measurement/http01/run.go`. Planned case: `internal/measurement/http01/conformance_test.go`, `symbol:TestMHttp01Normalization`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/result/http01.go`.

Case plan: Exercise every outcome row plus configured-check combinations and partial/unevaluated checks, body completion versus truncation and unsupported content encoding. Preserve failure at the observed stage and indexed decidable check outcomes; do not invent verdicts for checks not evaluated or select undocumented precedence.

<!-- requirement-ref: REQ-HTTP-01-0008 -->
[REQ-HTTP-01-0008](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0008) → `M-HTTP-01-NORMALIZATION` (Success Condition).

<!-- requirement-ref: REQ-HTTP-01-0009 -->
[REQ-HTTP-01-0009](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0009) → `M-HTTP-01-NORMALIZATION` (Failure Conditions).

<!-- requirement-ref: REQ-HTTP-01-0036 -->
[REQ-HTTP-01-0036](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0036) → `M-HTTP-01-NORMALIZATION` (Outcomes).

<!-- requirement-ref: REQ-HTTP-01-0037 -->
[REQ-HTTP-01-0037](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0037) → `M-HTTP-01-NORMALIZATION` (Outcomes).

#### M-HTTP-01-LIMITS

Intended owner: `internal/measurement/http01/run.go`. Planned case: `internal/measurement/http01/conformance_test.go`, `symbol:TestMHttp01Limits`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Test 16,384 response-header bytes, configured body bound and at most two followed redirects within outer connection/byte/time bounds at below/equal/above boundaries. Verify close at the active bound, no test retries and URL deduplication; distinguish application bytes processed from lower-layer buffering/on-wire use.

<!-- requirement-ref: REQ-HTTP-01-0032 -->
[REQ-HTTP-01-0032](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0032) → `M-HTTP-01-LIMITS` (Resource Bounds).

#### M-HTTP-01-VERSIONS

Intended owner: `internal/registry/measurements.go`. Planned case: `internal/registry/measurement_conformance_test.go`, `symbol:TestMHttp01Versions`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/result/http01.go`.

Case plan: Select exact registered test, Measurement Result schema and test result versions. Vary unsupported/malformed/cross-test versions and schema combinations; reject before execution rather than guessing, coercing or adding another version. Registration metadata and Draft status confer no support.

<!-- requirement-ref: REQ-HTTP-01-0033 -->
[REQ-HTTP-01-0033](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0033) → `M-HTTP-01-VERSIONS` (Output).

#### M-HTTP-01-OUTPUT

Intended owner: `internal/result/http01.go`. Planned case: `internal/result/http01_conformance_test.go`, `symbol:TestMHttp01Output`. Disposition: **planned**, with no execution or support evidence.

Case plan: Validate ordered observations only for performed DNS/TCP/TLS/HTTP operations, including bounded redirect chains and connection-reuse omissions. Terminal HTTP retains only status, minimized redirect class, bounded body metadata and indexed checks; no request/response content.

<!-- requirement-ref: REQ-HTTP-01-0034 -->
[REQ-HTTP-01-0034](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0034) → `M-HTTP-01-OUTPUT` (Output).

<!-- requirement-ref: REQ-HTTP-01-0035 -->
[REQ-HTTP-01-0035](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0035) → `M-HTTP-01-OUTPUT` (Output).

#### M-HTTP-01-TERMINATION

Intended owner: `internal/measurement/http01/run.go`. Planned case: `internal/measurement/http01/conformance_test.go`, `symbol:TestMHttp01Termination`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/execution/measurements.go`.

Case plan: Interrupt before/during/after every operation through network-generation change, cancellation, attempt deadline, proxy loss/reconfiguration where applicable, resource bounds and suspension as applicable to each linked source. Assert the exact common execution state/reason and measurement presence/absence; keep NET-05 permitted partial profile output distinct from primitive output. Release resources and prohibit fallback or invented terminal protocol outcomes.

<!-- requirement-ref: REQ-HTTP-01-0038 -->
[REQ-HTTP-01-0038](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0038) → `M-HTTP-01-TERMINATION` (Outcomes).

#### M-HTTP-01-MINIMIZATION

Intended owner: `internal/result/http01.go`. Planned case: `internal/result/http01_conformance_test.go`, `symbol:TestMHttp01Minimization`. Disposition: **planned**, with no execution or support evidence.

Case plan: Inject unique synthetic canaries into every prohibited target, identifier, address, DNS/TLS/HTTP/platform/error field named by each linked unit. Inspect serialized result/error paths and all nested observations, asserting no retained canary or extra field; preserve only explicitly allowed bounded normalized DNS answer values. No participant or live measurement data is used.

<!-- requirement-ref: REQ-HTTP-01-0039 -->
[REQ-HTTP-01-0039](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0039) → `M-HTTP-01-MINIMIZATION` (Outcomes).

#### M-HTTP-01-INTERPRETATION

Intended owner: `docs/measurement-interpretation-review.md`. Planned case: `docs/measurement-interpretation-review.md`, `heading:M-HTTP-01-INTERPRETATION`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/measurement/http01/run.go`, `internal/result/http01.go`.

Case plan: Review one endpoint/approved terminal response against registered expectations and all stronger-evidence criteria. Body bytes processed are not an exact wire limit; redirects/content changes require manual or separately reviewed classification with CDN/portal/route/family/stale-expectation alternatives.

Scope: separately reviewed integration/interpretation evidence; portable tests cannot complete this family alone.

<!-- requirement-ref: REQ-HTTP-01-0001 -->
[REQ-HTTP-01-0001](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0001) → `M-HTTP-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-HTTP-01-0002 -->
[REQ-HTTP-01-0002](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0002) → `M-HTTP-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-HTTP-01-0003 -->
[REQ-HTTP-01-0003](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0003) → `M-HTTP-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-HTTP-01-0040 -->
[REQ-HTTP-01-0040](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0040) → `M-HTTP-01-INTERPRETATION` (Supported Interpretation).

<!-- requirement-ref: REQ-HTTP-01-0041 -->
[REQ-HTTP-01-0041](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0041) → `M-HTTP-01-INTERPRETATION` (Supported Interpretation).

<!-- requirement-ref: REQ-HTTP-01-0042 -->
[REQ-HTTP-01-0042](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0042) → `M-HTTP-01-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

<!-- requirement-ref: REQ-HTTP-01-0043 -->
[REQ-HTTP-01-0043](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0043) → `M-HTTP-01-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

#### M-HTTP-01-SAFETY

Intended owner: `docs/measurement-deployment-review.md`. Planned case: `docs/measurement-deployment-review.md`, `heading:M-HTTP-01-SAFETY`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/execution/measurements.go`, `internal/result/http01.go`.

Case plan: Review every baseline/factor row and separately exercise each normative safety action using synthetic targets and recording transports. Keep effective risk at least the registered baseline and preserve host/local-policy qualification and final probe authority; no identifiers, unsafe target expansion or distinctive scheduling is introduced. Native capabilities, actual traffic patterns, host scheduling/consent and deployment review require separate scoped evidence; explanatory rationales are not new runtime data.

Scope: separately reviewed integration/interpretation evidence; portable tests cannot complete this family alone.

<!-- requirement-ref: REQ-HTTP-01-0044 -->
[REQ-HTTP-01-0044](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0044) → `M-HTTP-01-SAFETY` (Baseline Risk).

<!-- requirement-ref: REQ-HTTP-01-0045 -->
[REQ-HTTP-01-0045](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0045) → `M-HTTP-01-SAFETY` (Risk Factors).

<!-- requirement-ref: REQ-HTTP-01-0046 -->
[REQ-HTTP-01-0046](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0046) → `M-HTTP-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-HTTP-01-0047 -->
[REQ-HTTP-01-0047](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0047) → `M-HTTP-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-HTTP-01-0048 -->
[REQ-HTTP-01-0048](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0048) → `M-HTTP-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-HTTP-01-0049 -->
[REQ-HTTP-01-0049](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0049) → `M-HTTP-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-HTTP-01-0050 -->
[REQ-HTTP-01-0050](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0050) → `M-HTTP-01-SAFETY` (Safety Analysis).

## Maintenance and Validation

Before adopting a new public or Go baseline, review source/catalog/issue changes, recheck all measurement IDs and case applicability, and update pinned public links together with the manifest and regenerated report. Preserve historical manifests/reports rather than presenting their old evidence as current. A newly allocated ID must receive its own reviewed disposition; planned work must not silently become implemented or passed.

The manifest maps every allocated measurement ID at the selected public baseline. Its targets/cases are planned and support/evidence arrays are empty. Shared IDs remain in the separately scoped cross-cutting manifest; residual unallocated authoritative prose, threat-model/measurement-limitations allocation and full semantic inventory review remain visible. The applicable open specification issue is linked to test-specific schema/option/version work; its immutable finding remains available at the Go baseline.

Reproduce the metadata report using [report provenance and commands](../traceability/reports/README.md#measurement-report-dev-473). The checker receives explicitly supplied public/Go repositories, reviewed schema/tooling, expected baselines and current mode under [Manifest Validation](traceability-manifest.md). Verify exact ID/case/owner/scope correspondence between this plan and the manifest; the checker does not execute planned cases or prove semantic completeness.

Run public allocation-history/reference checks with explicitly supplied Go/internal consumers, Go `make check`, documentation/link checks and `git diff --check`. Public tooling dependencies remain development-only and are not added to the Go runtime module or its canonical checks. Public merge precedes this Go adoption; no runtime release or support promotion is introduced. Present the validated Go change and wait for separate final approval before commit/push.
