# Measurement Requirements and Case Plan

## Status, Authority, and Baselines

The last milestone manifest covers 396 allocated public requirement units across the nine Draft `0.1.0` measurements and 108 case families. The measurement manifest is pinned to public `e6257ec64564ad415a2d6c1f8d106fa1763d7de8`, which adds REQ-HTTP-01-0051 (hostname proxy redirect rejection) and REQ-HTTP-01-0052 (HTTP-01 terminal-outcome precedence), and Go `dbf43f6fd2d27e85c6f920a4a864991f59733bbf`, the DEV-622 strict-JSON update after the Measurement web reachability and composite milestone (DEV-670, closing DEV-657), which re-executes that milestone's cases without changing a status: 157 units are `partial`, 239 are `planned`, none is `implemented`, and 37 family records and 66 unit-test cases are `passed` with [evidence](../traceability/evidence/go-unit-dbf43f6.md). The NET-01, NET-03, DNS-03 and DNS-04 suites are not yet mapped and wait for the next milestone. Earlier manifests and evidence are retained in Git history.

DNS-01, TCP-01, TLS-01 and HTTP-01 have production codecs and runners, NET-05 has a production codec and sequential runner, NET-01 has a production codec and zero-traffic runner, NET-03 has a production codec and validation runner, DNS-04 has a production codec and DNS-over-TLS runner, and DNS-03 has a production codec and DNS-over-HTTPS runner; their synthetic family evidence is recorded in [Current NET-01 Execution](#current-net-01-execution), [Current NET-03 Execution](#current-net-03-execution), [Current DNS-04 Execution](#current-dns-04-execution) and [Current DNS-03 Execution](#current-dns-03-execution) until the next milestone. The DNS-01, TCP-01, TLS-01 and HTTP-01 ADMISSION, OPTIONS, EXECUTION, NORMALIZATION, LIMITS, TERMINATION, OUTPUT and VERSIONS families and the NET-05 OPTIONS, NORMALIZATION, OUTPUT, TERMINATION and VERSIONS families are `passed` with synthetic evidence, as recorded below. NET-05 ADMISSION, EXECUTION and LIMITS stay `planned` because rows of their plans are approved deferrals owned by [DEV-658](https://app.clickup.com/t/124kk1zcw4g): profile-version monotonicity, summed declared primitive maxima and fresh or cached NET-03 context ([NET-05](#net-05)); unit-test cases record the rows they execute. Four plan rows record approved dispositions instead of tests: the TLS-01 ALPN length bound, subsumed by registry membership; the HTTP-01 `inconclusive` outcome, exercised at the exchange layer; and, in the planned NET-05 limits plan, the 65,536-byte sent ceiling and the 49,152-byte profile bound (GO-SPEC-ISSUE-018), exercised at admission only. REQ-HTTP-01-0051 is mapped to the HTTP-01 admission and execution families and REQ-HTTP-01-0052 to normalization. The PRECONDITIONS, MINIMIZATION, INTERPRETATION and SAFETY families of these tests, and all `RV-*` cases stay planned. Execution does not establish review completeness, full conformance, platform or field support, or a release.

The [public allocation guide](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/requirements/measurement-allocation.md) records unit boundaries and informative dispositions. The public text owns observable behavior; this document assigns Go work and plans validation. Existing [cross-cutting mappings](cross-cutting-requirements.md) and reports retain their original selected baselines. All shared data-format/library source files and existing catalog entries are unchanged between their public PR #6 baseline and this PR #7 merge. This establishes source continuity, not newly executed conformance or support.

Public prerequisite ordering is satisfied. No internal review change is required for this annotation-only allocation: no public behavior, version, runtime field or private normative obligation changes. Native/host/analyst review and validation remain separately scoped future evidence.

## Ownership and Validation Boundaries

Each family must exercise **every linked unit**, every schema/table row, each conditional relation, numeric/string/count boundary and rejection path. A family is a parameterized suite, not one representative test. Test plans describe synthetic future cases; names/paths do not create packages or executable evidence. All numeric limits and defaults remain controlled by the linked public source and any stricter shared/task/local policy.

Admission and option families plan strict test-specific schemas and pure validation in `internal/validation` plus runner-side guard checks. Source/target authority remains in `internal/admission`; it cannot be accepted from serialized task input. Execution/normalization/limit families plan measurement runners, shared protocol owners and `internal/execution`. Versions belong to `internal/registry` and result schema selection. Output/minimization belong to the runner and `internal/result` using `internal/model/observation` and `internal/model/result`. These are intended owners under [Package Architecture](package-architecture.md), not claims that those packages or declarations exist.

Portable fake-platform tests are necessary for NET-01/NET-03/network generation but cannot establish Android or Apple API, permissions, trust-store, suspension, routing or cleanup behavior. Preconditions/safety require separately recorded native and host review/validation. Interpretation families pair a portable check for prohibited claims with a typed `RV-*` result/host/analyst review case: a Go test cannot establish correctness of an external analyst's interpretation. Local synthetic canaries, deterministic clocks and recording transports are the default; no participant data or uncontrolled measurement is needed.

Unresolved public contract ambiguities must enter [Specification Issues](specification-issues.md) and block affected implementation/claims. In particular, outcome-table testing must not invent precedence for combinations the public contract leaves unclear. GO-SPEC-ISSUE-001 remains open for the pending-store bundle re-read and binding/host evidence; the registered schemas and measurement runtime consumers and producers are recorded in its [Registered-Schema Adoption Addendum](specification-issues.md#registered-schema-adoption-addendum--october-27-2026).

## Current HTTP-01 Execution

The production `HTTP-01/0.1.0` entry uses `internal/measurement/http01` for closed input/options decoding, normalization, exact URL/check target extraction, explicit-IP-only proxy compatibility and one bounded HTTPS GET chain. The public baseline for its outcome order is [`e6257ec64564ad415a2d6c1f8d106fa1763d7de8`](https://github.com/tv-rain/barometer-spec/tree/e6257ec64564ad415a2d6c1f8d106fa1763d7de8), with internal validation alignment merged at `986a207176a9190d1c4c354fd69df55cfab23616`. The implementation-author scoped precedence review and baseline reconciliation are recorded in the internal HTTP-01 companion; broader reviews remain deferred.

The runner composes the execution-owned staged connection capability, authenticated TLS primitive and bounded HTTP/1.1 exchange primitive. One direct hostname lookup is pinned by the target gate before TCP; it emits only an empty-response system DNS observation, without retaining addresses. Direct hostname `auto` currently selects IPv4 without fallback; explicit `ipv6` selects one AAAA lookup. An explicit-IP URL uses its literal family and emits no DNS observation. Proxy URLs require a literal public IP and `address_family: auto`; every redirect uses a complete registered tuple and the same scope, path generation, budgets and cancellation. URL normalization folds the authority case, canonicalizes IPs, omits default port 443 and supplies an empty path as `/`, preserving query and nonempty path bytes.

HTTP uses identity encoding, no credentials/cookies/User-Agent/body, no automatic retry or connection reuse, at most two authorized redirect requests, and the registered header/body bounds. The TLS plaintext wrapper meters request/response framing and read-ahead; TLS records and CONNECT stay below that boundary. `max_body_bytes` exceeding the task receive bound rejects at decoding/admission without clamping. Codec byte-bound errors preserve only the fixed limit class. Equivalent normalized URLs serialize under their attempt contexts. The containing result decoder receives the exact configured body, redirect and check limits; all output is minimized and revalidated before persistence.

`http01_conformance_test.go` contains the eight `TestMHttp01*` family suites plus production direct/proxy, DNS, redirect, schema/minimization and resource cases. `internal/measurement/http01/codec_test.go` covers URL/check length/count boundaries, immutable tuples and fuzz seeds. `internal/execution/http01_termination_test.go:TestHttp01AllTerminationBoundaries` applies eight common reasons to pre-start, DNS, TCP, TLS, HTTP headers, body and recorded-outcome cleanup, without emitting partial HTTP-01 output. The shared primitive suites provide additional body/check, header, encoding and counted-boundary evidence. `http01_closeout_test.go`, `internal/execution/http01_termination_points_test.go` and `internal/protocol/http/http01_inconclusive_test.go` add exact rejection classes, check semantics, option defaults and overrides, outer bounds on both paths, a differently spelled equivalent URL, redirect and hop termination points, version shapes and the `inconclusive` outcome at the exchange layer. The eight families are `passed` at Go `1a37603` ([evidence](../traceability/evidence/go-unit-1a37603.md#M-HTTP-01-ADMISSION)).

These are synthetic Go cases. Full conformance, independent deployment/interpretation review, native resolver/trust-store behavior, host integration, device validation, platform support and release support are not claimed. A response mismatch or transport failure is an observation with ordinary server, network, certificate or configuration explanations, not a censorship verdict.

## Shared Authority and Reuse

The [public reuse matrix](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/requirements/measurement-allocation.md#reused-contracts) links the controlling shared contracts. Reuse the existing cross-cutting case families for strict JSON, task/execution/result envelopes, protocol observations, network context, registered limits, trust and proxy/path behavior, checking their applicability at the selected new baseline. Requirement IDs retain source-local identities; one shared case is not independent evidence for every caller without complete caller-specific coverage review.

NET-05 dispatches exact-version DNS-01/TCP-01/TLS-01/HTTP-01 schemas and calls shared protocol operations directly; its profile/sequence/aggregate/partial-output obligations remain separate. DNS-03/DNS-04/NET-03 share protocol machinery and observation formats, not the standalone TLS-01/HTTP-01 input or operation schema. NET-03 permits its registered HTTP validation behavior, DNS-03 uses POST, and HTTP-01 uses GET. NET-01 remains zero-traffic, without egress. These boundaries are part of each consumer's future integration cases.

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
| DNS-03 / DNS-04 | Shared DNS/TCP/TLS families; DNS-03 additionally consumes HTTP observation/bound families. Bootstrap, resolver authentication and POST/length-prefix exchange cases remain consumer-specific; they run in [Current DNS-03 Execution](#current-dns-03-execution) and [Current DNS-04 Execution](#current-dns-04-execution). |
| NET-03 | Shared access/platform and DNS/TCP/TLS/HTTP observation families, including HTTP/validation bounds. [Protocol output cases](#m-net-03-output) retain NET-03's distinct platform/active evidence and redirect prohibition. |
| NET-05 | Reuse the exact-version [DNS-01 admission](#m-dns-01-admission), [TCP-01 admission](#m-tcp-01-admission), [TLS-01 admission](#m-tls-01-admission), [HTTP-01 admission](#m-http-01-admission), and each primitive's option, execution, normalization and output families below; [context acquisition](cross-cutting-requirements.md#cc-context-acquisition-and-fallback), [composite limits](cross-cutting-requirements.md#cc-limit-net-05-composite-limits), and [partial results](cross-cutting-requirements.md#cc-result-partial-results) remain additional dependencies. [NET-05 execution](#m-net-05-execution) and [output](#m-net-05-output) check caller-specific composition. |

## Termination States after GO-SPEC-ISSUE-004

The `M-*-TERMINATION` plans were written before the [GO-SPEC-ISSUE-004 disposition](specification-issues.md#go-spec-issue-004-boundary-between-not_run-and-incomplete), and several expect `incomplete` at every interruption point. DEV-561 adopts the disposition in Go execution at public [`84f52f8`](https://github.com/tv-rain/barometer-spec/tree/84f52f8872a5af9e6d97894611e3b2eaba18aba8) ([State Mapping](execution-models.md#state-mapping-and-go-spec-issue-004)). The manifest plans carry these expected states, and REQ-NET-05-0061 is mapped to `M-NET-05-TERMINATION`. Applied to the named points of each plan:

| Interruption point | Expected result | Public rule |
|---|---|---|
| Before the test starts: input validation, context capture or acquisition, constraint evaluation, and any point before the test's first test-specific network operation or its own network-state observation | `not_run` with the first condition's reason, no `measurement` and no `network_context` | [REQ-RESULT-0051](https://github.com/tv-rain/barometer-spec/blob/84f52f8872a5af9e6d97894611e3b2eaba18aba8/data-formats/measurement-result.md#REQ-RESULT-0051), [REQ-RESULT-0053](https://github.com/tv-rain/barometer-spec/blob/84f52f8872a5af9e6d97894611e3b2eaba18aba8/data-formats/measurement-result.md#REQ-RESULT-0053) |
| After the test starts and before it records its terminal outcome, including during the test's own network-generation or stability check | `incomplete` with the first condition's reason, as each plan states | REQ-RESULT-0051 and each test's termination rule |
| After the test records its terminal outcome, while connections close and resources are released | `completed` with the recorded outcome; the interruption changes neither state, reason nor `measurement`, and release stays within the attempt's bounds | [REQ-RESULT-0052](https://github.com/tv-rain/barometer-spec/blob/84f52f8872a5af9e6d97894611e3b2eaba18aba8/data-formats/measurement-result.md#REQ-RESULT-0052) |


- **`M-NET-01-TERMINATION`:** before step 1 is `not_run`. NET-01 collection starts the test, so steps 1 to 3 are `incomplete`. Release of callbacks and handles in step 4 is `completed` once the outcome is recorded.
- **`M-NET-03-TERMINATION`:** the NET-01 network-generation stability check starts the test; every later point before the outcome is `incomplete`. "After the first response or failure but before the connection is closed" is `completed` once the outcome is recorded.
- **`M-NET-05-TERMINATION`:** point (a) follows the stable-network check of Procedure step 3, so it stays `incomplete`. A task-level `require_validated` constraint that is unsatisfied before that check is `not_run` with `constraints_unsatisfied`. Point (d) is `completed` once the composite outcome is recorded. `network_not_validated` and `inconclusive` are `completed` and the NET-05 `incomplete` outcome stays with `execution.state: incomplete` ([REQ-NET-05-0061](https://github.com/tv-rain/barometer-spec/blob/84f52f8872a5af9e6d97894611e3b2eaba18aba8/tests/net-05.md#REQ-NET-05-0061)).
- **`M-DNS-01-TERMINATION`:** every listed point follows the recorded network generation, so all stay `incomplete`.
- **`M-DNS-03-TERMINATION`:** after input validation and before bootstrap is `not_run`. After the final `doh` `dns_query` observation is `completed` once the outcome is recorded.
- **`M-DNS-04-TERMINATION`:** before input validation completes is `not_run`. While closing the session is `completed` once the outcome is recorded.
- **`M-TCP-01-TERMINATION`:** during address and port validation is `not_run`. The network-generation and address-family check starts the test. After the outcome is recorded and during the close are `completed`.
- **`M-TLS-01-TERMINATION`:** after input validation (step 1) is `not_run`; step 2 starts the test. After the negotiated properties are recorded and before close is `completed` once the outcome is recorded.
- **`M-HTTP-01-TERMINATION`:** after input validation and before the test's first network-state observation or DNS query is `not_run`. Every other listed point precedes the terminal outcome and stays `incomplete`.

A required network condition that cannot be determined is unsatisfied, so the attempt is `not_run` with `constraints_unsatisfied` ([REQ-EXECUTION-0031](https://github.com/tv-rain/barometer-spec/blob/84f52f8872a5af9e6d97894611e3b2eaba18aba8/data-formats/task-execution.md#REQ-EXECUTION-0031)). Classifying a test's own generation or stability check as its start follows the "own network-state observation" clause of REQ-RESULT-0051; each runner task confirms its procedure's first such step against the public text.

## Source Summary

| Test | Allocated units | Planned families |
|---|---:|---:|
| NET-01 | 38 | 12 |
| NET-03 | 40 | 12 |
| NET-05 | 61 | 12 |
| DNS-01 | 52 | 12 |
| DNS-03 | 39 | 12 |
| DNS-04 | 39 | 12 |
| TCP-01 | 37 | 12 |
| TLS-01 | 38 | 12 |
| HTTP-01 | 52 | 12 |

## Planned Cases and Requirement Dispositions

### NET-01

#### Current NET-01 Execution

`internal/measurement/net01` owns the strict empty-object codec and the runner; the root `registry.go` assembles its exact `0.1.0` entry with low baseline risk, `registry.NoTargets` and `registry.NoProxy`. NET-01 is not a NET-05 step. Admission accepts its empty target list because it has no remote target (REQ-TRUST-0013), and rejects the proxy path (REQ-PROXY-0026). The execution engine runs NET-01 without a path dialer, resolver or egress acquisition: stated access conditions are evaluated on a gate read before the test starts, a stated egress requirement is `not_run` with `network_context_unavailable`, and the test's own read through `networkcontext.Source` starts it and arms the generation watch. The platform port reports `platform.ErrNetworkUnavailable` for a missing network-state API, which becomes `platform_unavailable`; any other failure becomes `platform_error`. The `inconclusive` outcome is not emitted until the public contract defines an inconsistent snapshot. A platform that is already unavailable at `StartRun` leaves no run, because the reservation check reads access state first. No result carries the top-level Network Context.

The suites are portable synthetic evidence with a fake platform port, a deterministic clock and a recording transport. Native Android and Apple network-state APIs, permissions, change signals and suspension delivery (DEV-455), and the PRECONDITIONS, MINIMIZATION, INTERPRETATION and SAFETY families and their `RV-*` cases remain deferred.

| Family | Executable suite |
|---|---|
| `M-NET-01-ADMISSION` | `TestMNet01Admission` |
| `M-NET-01-OPTIONS` | `TestMNet01Options` |
| `M-NET-01-EXECUTION` | `TestMNet01ExecutionAndOutput` |
| `M-NET-01-NORMALIZATION` | `TestMNet01Normalization` (partial: no `inconclusive` row) |
| `M-NET-01-LIMITS` | `TestMNet01Limits`, `TestNet01DeadlineReleasesRead` |
| `M-NET-01-TERMINATION` | `TestMNet01Termination`, `TestNet01AllTerminationBoundaries`, `TestNet01GenerationChange` |
| `M-NET-01-OUTPUT` | `TestMNet01ExecutionAndOutput`, `TestMNet01Versions` |
| `M-NET-01-VERSIONS` | `TestMNet01Versions` |

The manifest records these families at the next milestone.

#### M-NET-01-ADMISSION

Disposition: **passed** for synthetic production admission in `net01_conformance_test.go`, `symbol:TestMNet01Admission`: the empty input object is admitted with no target registration, and wrong root types, a missing input and interface, address, device, Wi-Fi, subscriber and hostname properties are rejected as `task_schema_invalid` with no handle, replay, run, platform read, transport use or result. The proxy path is rejected at admission. No support claim follows.

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

Disposition: **passed** for synthetic production admission in `net01_conformance_test.go`, `symbol:TestMNet01Options`: only the empty options object is admitted, and it selects the one snapshot behavior.

Case plan: Accept the empty options object and reject every added property or wrong type without changing the selected snapshot behavior.

<!-- requirement-ref: REQ-NET-01-0015 -->
[REQ-NET-01-0015](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0015) → `M-NET-01-OPTIONS` (Options).

<!-- requirement-ref: REQ-NET-01-0016 -->
[REQ-NET-01-0016](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0016) → `M-NET-01-OPTIONS` (Options).

#### M-NET-01-PRECONDITIONS

Intended owner: `internal/measurement/net01/run.go`. Planned case: `internal/measurement/net01/conformance_test.go`, `symbol:TestMNet01Preconditions`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/platform/network.go`, `docs/measurement-platform-validation.md`.

Case plan: Run the portable NET-01 gate against a fake platform with a synthetic capability matrix and assert no network operation is attempted. Supported network-state API absent: assert execution.state not_run with a registered reason per the public table and no measurement. Required state API present but the only available implementation path requires interface addresses, SSIDs, BSSIDs, subscriber identifiers or location (iterate each): assert the gate denies before any such access occurs and the fake platform records zero calls to those accessors. Assert the gate requests from the fake platform only the network-state capability. Stricter task execution constraint or local probe policy denial: assert not_run with constraints_unsatisfied or local_policy_rejected respectively. Requirements: REQ-NET-01-0017, REQ-NET-01-0018, REQ-NET-01-0019, REQ-NET-01-0020.

Scope: portable automated component only; judgement-based completion is review case `RV-M-NET-01-PRECONDITIONS`, which is not automated coverage.

Review case `RV-M-NET-01-PRECONDITIONS` (`docs/measurement-deployment-review.md`, `heading:RV-M-NET-01-PRECONDITIONS`): Human review case; not automated coverage. Reviewer role: Mobile platform privacy reviewer (independent of the implementer). Reviewed artifact: Android and Apple NET-01 adapter code, manifest/entitlement declarations and host integration guidance at the reviewed implementation revision. Checklist: (1) Does the Android adapter use a supported network-state API and declare only the permissions required to read network state? (2) Does the Apple adapter use a supported network-state API and declare only the entitlements or capabilities required to read network state? (3) Is it confirmed on each platform that no code path reads interface addresses, SSIDs, BSSIDs, subscriber identifiers or location? (4) Does host guidance tell integrators not to grant broader capabilities for NET-01? (5) Are platform-version differences that change the minimum capability set documented as limitations? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-NET-01-0017 -->
[REQ-NET-01-0017](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0017) → `M-NET-01-PRECONDITIONS`, `RV-M-NET-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-NET-01-0018 -->
[REQ-NET-01-0018](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0018) → `M-NET-01-PRECONDITIONS`, `RV-M-NET-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-NET-01-0019 -->
[REQ-NET-01-0019](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0019) → `M-NET-01-PRECONDITIONS`, `RV-M-NET-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-NET-01-0020 -->
[REQ-NET-01-0020](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0020) → `M-NET-01-PRECONDITIONS`, `RV-M-NET-01-PRECONDITIONS` (Preconditions).

#### M-NET-01-EXECUTION

Disposition: **passed** for synthetic production execution in `net01_conformance_test.go`, `symbol:TestMNet01ExecutionAndOutput`, through `Client` with the production registry and direct path. One platform read through the shared collector starts the test; the recording transport sees no dialer request, connection or egress lookup, and the generation watch is released. Native Android and Apple network-state behavior remains a DEV-455 gap.

Case plan: Use a recording platform provider to assert normalized access reads, unknown values, transition detection and release of handles/callbacks. Install a transport that fails on any use; no primary or fallback egress endpoint may be contacted.

<!-- requirement-ref: REQ-NET-01-0006 -->
[REQ-NET-01-0006](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0006) → `M-NET-01-EXECUTION` (Operation).

<!-- requirement-ref: REQ-NET-01-0007 -->
[REQ-NET-01-0007](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0007) → `M-NET-01-EXECUTION` (Procedure).

#### M-NET-01-NORMALIZATION

Disposition: **partial**, synthetic production cases in `net01_conformance_test.go`, `symbol:TestMNet01Normalization`: every access value, unrecognized platform values as `unknown`, `platform_unavailable` from the platform's no-API signal and `platform_error` from any other failure or a zero generation, with no raw error text. The `inconclusive` row is not executed: the public contract does not define an internally inconsistent snapshot, so the runner never emits `inconclusive` until it does. A snapshot is one platform read and cannot combine generations.

Case plan: Supply coherent, unavailable, failed and inconsistent snapshots; exercise every failure/outcome row and distinguish unknown properties from unavailable/inconsistent snapshots. Verify completed snapshots cannot combine generations.

<!-- requirement-ref: REQ-NET-01-0008 -->
[REQ-NET-01-0008](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0008) → `M-NET-01-NORMALIZATION` (Success Condition).

<!-- requirement-ref: REQ-NET-01-0009 -->
[REQ-NET-01-0009](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0009) → `M-NET-01-NORMALIZATION` (Failure Conditions).

<!-- requirement-ref: REQ-NET-01-0026 -->
[REQ-NET-01-0026](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0026) → `M-NET-01-NORMALIZATION` (Outcomes).

#### M-NET-01-LIMITS

Disposition: **passed** for synthetic cases in `net01_conformance_test.go`, `symbol:TestMNet01Limits`, and `internal/execution/net01_termination_test.go`, `symbol:TestNet01DeadlineReleasesRead`: the outer deadline releases an unanswered read with zero transport use, and concurrent runs across a generation transition each keep one coherent generation or end `incomplete` with `network_changed`.

Case plan: Drive the fake clock to the outer deadline and assert prompt release with zero network operations/bytes. Run concurrent snapshots with a generation transition and assert coherent values or the prescribed incomplete state, never merged state.

<!-- requirement-ref: REQ-NET-01-0021 -->
[REQ-NET-01-0021](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0021) → `M-NET-01-LIMITS` (Resource Bounds).

#### M-NET-01-VERSIONS

Disposition: **passed** for synthetic cases in `net01_conformance_test.go`, `symbol:TestMNet01Versions` (the task's singular VERSION family): the exact emitted task/schema/result combination, task-time rejection of unsupported, malformed and missing `test_version` and `schema_version`, and result rejection of unsupported versions and cross-test identities and shapes.

Case plan: Select the exact registered combination `test.id: NET-01`, `test.version: 0.1.0`, `test.result_version: 0.1.0` with Measurement Result `schema_version: 0.1.0` and verify it is accepted and emitted as exactly that combination. Iterate rejections on each field independently, with the other fields at their registered values: an unregistered but well-formed version (`0.2.0`, `1.0.0`), malformed strings (`0.1`, `v0.1.0`, `0.1.0-draft`, empty string, non-string value) and a missing field, for `schema_version`, `test.version` and `test.result_version`. Iterate cross-test combinations: a result with `test.id: NET-03` carrying the NET-01 `access` measurement shape, and a NET-01 result whose `measurement` follows another registered test's Output shape. Assert unsupported task versions are rejected as task-validation failures before execution through the operational interface and produce no Measurement Result, and that unsupported result combinations are rejected rather than coerced, interpreted best-effort or mapped to another version. Assert the `Draft` status and `0.1.0` registration metadata alone add no supported combination. Requirements: REQ-NET-01-0022.

<!-- requirement-ref: REQ-NET-01-0022 -->
[REQ-NET-01-0022](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0022) → `M-NET-01-VERSIONS` (Output).

#### M-NET-01-OUTPUT

Disposition: **passed** for synthetic production output in `net01_conformance_test.go`, `symbol:TestMNet01ExecutionAndOutput` and `symbol:TestMNet01Versions`: no top-level `network_context` or `egress`, an empty `observations` array, the six-field `access` object for `snapshot`, and the result boundary's conditional `access` rule for `inconclusive` and the failure outcomes.

Case plan: Verify no top-level network_context, no egress, empty observations and the conditional complete access object for snapshot/inconclusive. Exercise all access enums through the canonical shared format suite.

<!-- requirement-ref: REQ-NET-01-0023 -->
[REQ-NET-01-0023](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0023) → `M-NET-01-OUTPUT` (Output).

<!-- requirement-ref: REQ-NET-01-0024 -->
[REQ-NET-01-0024](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0024) → `M-NET-01-OUTPUT` (Output).

<!-- requirement-ref: REQ-NET-01-0025 -->
[REQ-NET-01-0025](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0025) → `M-NET-01-OUTPUT` (Additional Field).

#### M-NET-01-TERMINATION

Disposition: **passed** for synthetic cases in `internal/execution/net01_termination_test.go`, `symbol:TestNet01AllTerminationBoundaries` and `symbol:TestNet01GenerationChange`, and `net01_conformance_test.go`, `symbol:TestMNet01Termination`: before the read is `not_run`; during and after the read is `incomplete` with the first cause; after the outcome is recorded is `completed` with the watch released. Suspension is an injected engine cause, not device evidence.

Case plan: Use a fake platform network-state provider, a deterministic clock and a recording transport that fails on any use, and inject interruptions at each NET-01 Procedure point: before step 1 (read current path/connectivity state through the shared Network Context collector), during step 1 (provider call outstanding), between steps 1 and 2 (raw state read, not yet normalized into `access`), during step 3 (network-generation check) and during step 4 (release of callbacks/handles). Expected states follow the GO-SPEC-ISSUE-004 disposition: before step 1 the attempt is not_run with that reason, no measurement and no network_context; collection starts the test, so steps 1 to 3 are incomplete as asserted below; during step 4, after the outcome is recorded, the attempt stays completed with the recorded outcome and the release still completes. At each point inject separately: a network-generation change, host cancellation, outer attempt-deadline expiry on the fake clock, and platform suspension. At the incomplete points, assert a generation change during collection yields `execution.state: incomplete` with `execution.reason: network_changed` and no `measurement` object (no `outcome`, `observations` or `access`); assert cancellation, deadline expiry and suspension yield the common execution layer reasons `cancelled_by_host`, `attempt_deadline_exceeded` and `platform_suspended` with no `measurement`, never a NET-01 `outcome` value such as `platform_error` or `inconclusive`. Inject two conditions in sequence and assert `reason` names only the first. Assert no top-level `network_context` or `egress` is added on any interrupted path, every provider handle and callback created for the run is released, the recording transport records zero connections and zero bytes, and no retry or fallback (including any egress-classification endpoint) is attempted within the run. Proxy loss/reconfiguration is not exercised because NET-01 opens no network path. Requirements: REQ-NET-01-0027.

<!-- requirement-ref: REQ-NET-01-0027 -->
[REQ-NET-01-0027](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0027) → `M-NET-01-TERMINATION` (Outcomes).

#### M-NET-01-MINIMIZATION

Intended owner: `internal/result/net01.go`. Planned case: `internal/result/net01_conformance_test.go`, `symbol:TestMNet01Minimization`. Disposition: **planned**, with no execution or support evidence.

Case plan: Configure the fake platform provider to expose one unique synthetic canary string or value in each item REQ-NET-01-0028 prohibits: interface name, interface index, local address, public address, route, DNS configuration, SSID, BSSID, Wi-Fi scan data, cell identifier, SIM data, subscriber data, carrier name, signal strength, VPN application identity, network generation, platform object description and raw platform error text (the last via a scripted `platform_error` provider failure). Run NET-01 to `snapshot`, `platform_error`, `inconclusive` and `incomplete`/`network_changed` and assert the serialized Measurement Result, every nested object, the empty `observations` array, any task-validation or error path and any log or diagnostic surface exposed by the result path contain none of the canaries. Assert the only retained measurement fields are `outcome`, `observations` (empty) and, where permitted, `access` with exactly `network_type`, `vpn_state`, `ipv4_available`, `ipv6_available`, `validated` and `network_changed` drawn from their enumerated values; assert no top-level `network_context`, no `egress` object and no extra field. Only synthetic provider data is used. Requirements: REQ-NET-01-0028.

<!-- requirement-ref: REQ-NET-01-0028 -->
[REQ-NET-01-0028](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0028) → `M-NET-01-MINIMIZATION` (Outcomes).

#### M-NET-01-INTERPRETATION

Intended owner: `docs/measurement-interpretation-review.md`. Planned case: `internal/result/net01_conformance_test.go`, `symbol:TestMNet01Interpretation`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/measurement/net01/run.go`, `internal/result/net01.go`.

Case plan: Assert the serialized NET-01 `measurement` contains only `outcome`, `observations` and `access`, that `outcome` takes only `snapshot`, `platform_unavailable`, `platform_error` or `inconclusive`, and that no field names or values express a verdict or interpretation (for example censorship, blocking, anomaly, reachability or operator conclusions). Assert Go doc strings for the NET-01 result and access types, documented field descriptions and every sanitized example or host-facing text string shipped with the implementation contain none of these prohibited claims: that the Internet is reachable; that `validated: yes`, `ipv4_available: yes` or `ipv6_available: yes` establishes Internet reachability; identification of a network operator, subscriber, device or physical location; a country, region, ASN, provider, mobile operator or public IP address derived from NET-01; that `vpn_state: detected` establishes traffic is routed through a VPN; and that a result is definitive proof of censorship. Verify the field descriptions retain the documented qualifiers that availability values are platform-reported and not proven reachability. Requirements: REQ-NET-01-0001, REQ-NET-01-0002, REQ-NET-01-0003, REQ-NET-01-0004, REQ-NET-01-0029, REQ-NET-01-0030, REQ-NET-01-0031.

Scope: portable automated component only; judgement-based completion is review case `RV-M-NET-01-INTERPRETATION`, which is not automated coverage.

Review case `RV-M-NET-01-INTERPRETATION` (`docs/measurement-interpretation-review.md`, `heading:RV-M-NET-01-INTERPRETATION`): Human review case; not automated coverage. Reviewer role: Measurement interpretation reviewer (analyst role independent of the implementer). Reviewed artifact: Go result documentation and host/analyst interpretation guidance for NET-01 at the reviewed implementation revision. Checklist: (1) Does the guidance limit a single NET-01 result to the normalized OS-exposed access properties at that time and whether the network changed during the snapshot? (2) Does the guidance state that country, region, ASN, provider and operator require separate egress or consented context and cannot be inferred from NET-01? (3) Does the guidance name the listed alternative explanations (incomplete or delayed platform state, VPN present without carrying traffic, dual-stack capability not establishing reachability, stale OS validation state)? (4) Does the guidance require NET-03 results, protocol measurements within the same stable network generation, or repeated snapshots after a reported change before any stronger interpretation? (5) Does any guidance that combines NET-01 with other results avoid presenting the combination as proof of censorship or attributing intent? (6) Is the guidance free of implied or paraphrased reachability, location or operator conclusions that the automated phrase checks would not catch? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-NET-01-0001 -->
[REQ-NET-01-0001](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0001) → `M-NET-01-INTERPRETATION`, `RV-M-NET-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-NET-01-0002 -->
[REQ-NET-01-0002](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0002) → `M-NET-01-INTERPRETATION`, `RV-M-NET-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-NET-01-0003 -->
[REQ-NET-01-0003](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0003) → `M-NET-01-INTERPRETATION`, `RV-M-NET-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-NET-01-0004 -->
[REQ-NET-01-0004](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0004) → `M-NET-01-INTERPRETATION`, `RV-M-NET-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-NET-01-0029 -->
[REQ-NET-01-0029](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0029) → `M-NET-01-INTERPRETATION`, `RV-M-NET-01-INTERPRETATION` (Supported Interpretation).

<!-- requirement-ref: REQ-NET-01-0030 -->
[REQ-NET-01-0030](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0030) → `M-NET-01-INTERPRETATION`, `RV-M-NET-01-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

<!-- requirement-ref: REQ-NET-01-0031 -->
[REQ-NET-01-0031](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0031) → `M-NET-01-INTERPRETATION`, `RV-M-NET-01-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

#### M-NET-01-SAFETY

Intended owner: `docs/measurement-deployment-review.md`. Planned case: `internal/admission/measurement_safety_conformance_test.go`, `symbol:TestMNet01Safety`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/execution/measurements.go`, `internal/result/net01.go`.

Case plan: Run NET-01 with a fake platform provider and a recording transport for each registered row and action. Baseline Risk: verify registered `risk_level: low` and `requires_user_consent: false`; exercise a task declaring `low` (accepted), a task declaring `medium` or `high` (effective risk at least the declared level), a task declaring `requires_user_consent: true` with no host-established consent (no execution), and a trusted local policy that rejects or delays (`not_run` with `local_policy_rejected`, no measurement). Risk Factors: for `network_observability` and `traffic_distinctiveness` assert zero connections, zero bytes and no DNS or socket use on the recording transport; for `target_sensitivity` assert no remote target is accepted (non-empty input rejected before execution); for `data_sensitivity` assert output limited to the six coarse `access` classifications with no interface or subscriber identifiers; for `legal_or_social_context` assert no network-visible participation signal (zero traffic). Safety Analysis actions: data minimization (no `egress`, no carrier identity), traffic distinctiveness (no traffic), user-identification risk (synthetic interface names, addresses, SSIDs, BSSIDs, carrier names and subscriber identifiers from the provider are absent), target risk (no target), repeated-execution risk (N repeated runs record zero network operations), residual risk (no `network_context` or `egress` joined to output), and probe final authority to delay, skip or reject. Requirements: REQ-NET-01-0032, REQ-NET-01-0033, REQ-NET-01-0034, REQ-NET-01-0035, REQ-NET-01-0036, REQ-NET-01-0037, REQ-NET-01-0038.

Scope: portable automated component only; judgement-based completion is review case `RV-M-NET-01-SAFETY`, which is not automated coverage.

Review case `RV-M-NET-01-SAFETY` (`docs/measurement-deployment-review.md`, `heading:RV-M-NET-01-SAFETY`): Human review case; not automated coverage. Reviewer role: Deployment safety and privacy reviewer (independent of the implementer). Reviewed artifact: NET-01 native Android and Apple adapter capability declarations, host scheduling and consent integration, and deployment documentation at the reviewed implementation revision. Checklist: (1) Do the native adapters request only the platform capabilities required to read coarse network state, with no permission granting interface addresses, SSIDs, BSSIDs, subscriber identifiers or location? (2) Does the host integration preserve that `requires_user_consent: false` is subject to host and local policy, without treating it as a blanket consent conclusion? (3) Does the deployment documentation address the residual fingerprinting risk of retaining coarse network-state combinations alongside egress context or unrelated telemetry, including retention bounds? (4) Does the host scheduling configuration bound snapshot frequency so repeated execution does not create excessive resource use? (5) Have the deployment assumptions behind the `legal_or_social_context` row been reviewed for each intended deployment context? (6) Does on-device evidence from each native platform show no network traffic during NET-01 execution? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-NET-01-0032 -->
[REQ-NET-01-0032](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0032) → `M-NET-01-SAFETY`, `RV-M-NET-01-SAFETY` (Baseline Risk).

<!-- requirement-ref: REQ-NET-01-0033 -->
[REQ-NET-01-0033](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0033) → `M-NET-01-SAFETY`, `RV-M-NET-01-SAFETY` (Risk Factors).

<!-- requirement-ref: REQ-NET-01-0034 -->
[REQ-NET-01-0034](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0034) → `M-NET-01-SAFETY`, `RV-M-NET-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-NET-01-0035 -->
[REQ-NET-01-0035](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0035) → `M-NET-01-SAFETY`, `RV-M-NET-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-NET-01-0036 -->
[REQ-NET-01-0036](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0036) → `M-NET-01-SAFETY`, `RV-M-NET-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-NET-01-0037 -->
[REQ-NET-01-0037](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0037) → `M-NET-01-SAFETY`, `RV-M-NET-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-NET-01-0038 -->
[REQ-NET-01-0038](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-01.md#REQ-NET-01-0038) → `M-NET-01-SAFETY`, `RV-M-NET-01-SAFETY` (Safety Analysis).

### NET-03

#### Current NET-03 Execution

`internal/measurement/net03` owns the strict codec and the runner; the root `registry.go` assembles its exact `0.1.0` entry with low baseline risk, a target rule that matches the complete normalized URL, expected status and optional body digest, and `registry.NoProxy` (REQ-PROXY-0044). NET-03 is not a NET-05 step. Validation URLs may be `http` or `https`: target tuples and the execution destination gate accept an `http` URL with default port 80, and its tuple names the scheme, so no `https` registration matches it. The codec rejects a redirect or informational `expected_status`.

The runner follows the Procedure. The NET-01 stability check (`CheckNetwork`) starts the test. One platform read through `measurement.Input.Validation` follows; it reads the interim `platform.AccessState.CaptivePortal` and `Validated` flags, maps them to `platform_validation`, records `unknown` when the platform cannot report them, and ends the attempt with `network_changed` when the reading belongs to another generation. One request through the shared HTTP exchange follows, with redirects disabled, a 1,024-byte response prefix, the status check and, when configured, the SHA-256 check; TLS is used only for `https`. A redirect location is classified without contact. The outcome is recorded before the connection closes. A per-client semaphore admits one validation run at a time; a later run waits for it or for its own context.

Response classes and the outcome follow REQ-NET-03-0041 to REQ-NET-03-0043, merged through public PR #35 at `f889a38` (GO-SPEC-ISSUE-006, GO-SPEC-ISSUE-019). A redirect status is `redirect`; a matching status with a decidable different digest is `content_mismatch`; any other parsed response, including a digest undecidable within the prefix, is `unexpected`; a header block over 16,384 bytes or an exchange without a final status has no class. The request decides first; platform `captive_portal` or `not_validated` selects the outcome only without a class; otherwise the DNS, TCP or TLS failure, or `inconclusive` for an HTTP failure. Go `requirement-ref` comments cite them against the reference pin `f889a38`.

The suites are portable synthetic evidence with a fake platform port, a deterministic clock, a recording transport and loopback HTTP and HTTPS servers. Native Android and Apple validation and captive-portal signals (DEV-455), the PRECONDITIONS, MINIMIZATION, INTERPRETATION and SAFETY families and their `RV-*` cases, and NET-05's use of NET-03 context remain deferred.

| Family | Executable suite |
|---|---|
| `M-NET-03-ADMISSION` | `TestMNet03Admission` |
| `M-NET-03-OPTIONS` | `TestMNet03Options` |
| `M-NET-03-EXECUTION` | `TestMNet03Execution` |
| `M-NET-03-NORMALIZATION` | `TestMNet03Normalization` |
| `M-NET-03-LIMITS` | `TestMNet03Limits` |
| `M-NET-03-TERMINATION` | `TestMNet03Termination`, `TestNet03AllTerminationBoundaries` |
| `M-NET-03-OUTPUT` | `TestMNet03Output`, `TestMNet03Execution` |
| `M-NET-03-VERSIONS` | `TestMNet03Versions` |

The manifest records these families and the three new requirement IDs at the next milestone.

#### M-NET-03-ADMISSION

Disposition: **passed** for synthetic production admission in `net03_conformance_test.go`, `symbol:TestMNet03Admission`: registered `https` and `http` profiles with and without a digest are admitted; missing, mistyped and out-of-range fields, unsafe schemes, credentials, fragments, loopback, private and link-local addresses, redirect and informational statuses, malformed digests, unknown and duplicate fields are `task_schema_invalid`; a 2,049-byte URL is `task_limit_exceeded`; a different URL, scheme, status or digest is `target_not_registered`; none produces a handle, replay commit, run or transport contact. A redirect canary destination is never contacted (`TestMNet03Execution`). No support claim follows.

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

<!-- requirement-ref: REQ-NET-03-0043 -->
[REQ-NET-03-0043](https://github.com/tv-rain/barometer-spec/blob/f889a381c84a781176024f3fed3a3d4f5c92b241/tests/net-03.md#REQ-NET-03-0043) → `M-NET-03-ADMISSION` (Input Validation). A redirect or informational `expected_status` is rejected; GO-SPEC-ISSUE-019 records its adoption.

#### M-NET-03-OPTIONS

Disposition: **passed** for synthetic production admission in `net03_conformance_test.go`, `symbol:TestMNet03Options`: omitted, `auto` and `ipv4` families and `ipv6` for a hostname URL are admitted; an unknown or null family, `ipv6` for an IPv4 literal, redirect-following, header, cookie, credential, method, body and body-size options and the proxy path are `task_schema_invalid`; a receive budget below the registered prefix is rejected rather than clamped.

Case plan: Exercise auto/ipv4/ipv6, omitted default and unsupported family; reject redirect-following, custom-header, credential and body options. No option may bypass local policy.

<!-- requirement-ref: REQ-NET-03-0015 -->
[REQ-NET-03-0015](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0015) → `M-NET-03-OPTIONS` (Options).

<!-- requirement-ref: REQ-NET-03-0016 -->
[REQ-NET-03-0016](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0016) → `M-NET-03-OPTIONS` (Options).

#### M-NET-03-PRECONDITIONS

Intended owner: `internal/measurement/net03/run.go`. Planned case: `internal/measurement/net03/conformance_test.go`, `symbol:TestMNet03Preconditions`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/platform/network.go`, `docs/measurement-platform-validation.md`.

Case plan: Run the portable NET-03 gate against a fake platform and synthetic local profile registry with a deterministic clock. Stable network generation unavailable: assert execution.state not_run, no request sent. Validation target or expected response profile not locally available, and present but not current: assert not_run before any request. All preconditions satisfied: assert exactly one request proceeds. Fake platform captive-portal flow that would display, submit or retain user data (iterate each): assert the gate denies, no request is sent and no flow data is retained. Stricter task execution constraint or local probe policy denial: assert not_run with constraints_unsatisfied or local_policy_rejected respectively; other denials use a registered reason per the public table. Requirements: REQ-NET-03-0017, REQ-NET-03-0018, REQ-NET-03-0019, REQ-NET-03-0020.

Scope: portable automated component only; judgement-based completion is review case `RV-M-NET-03-PRECONDITIONS`, which is not automated coverage.

Review case `RV-M-NET-03-PRECONDITIONS` (`docs/measurement-deployment-review.md`, `heading:RV-M-NET-03-PRECONDITIONS`): Human review case; not automated coverage. Reviewer role: Mobile platform privacy reviewer (independent of the implementer). Reviewed artifact: Android and Apple NET-03 adapter code, captive-portal integration and permission declarations at the reviewed implementation revision. Checklist: (1) On Android, can the adapter observe validation state without launching or interacting with a captive-portal UI? (2) On Apple platforms, can the adapter observe validation state without launching or interacting with a captive-portal UI? (3) Is it confirmed that no native path displays, submits or retains user data from a captive-portal flow? (4) Are only the permissions required for validation-state access declared on each platform? (5) Are platform cases where isolation cannot be established documented as reasons the test does not run? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-NET-03-0017 -->
[REQ-NET-03-0017](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0017) → `M-NET-03-PRECONDITIONS`, `RV-M-NET-03-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-NET-03-0018 -->
[REQ-NET-03-0018](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0018) → `M-NET-03-PRECONDITIONS`, `RV-M-NET-03-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-NET-03-0019 -->
[REQ-NET-03-0019](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0019) → `M-NET-03-PRECONDITIONS`, `RV-M-NET-03-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-NET-03-0020 -->
[REQ-NET-03-0020](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0020) → `M-NET-03-PRECONDITIONS`, `RV-M-NET-03-PRECONDITIONS` (Preconditions).

#### M-NET-03-EXECUTION

Disposition: **passed** for synthetic production execution in `net03_conformance_test.go`, `symbol:TestMNet03Execution`, through `Client` with the production registry and direct path, for `https` (TCP, TLS, HTTP) and `http` (TCP, HTTP) independently: the stability check, one platform read, one GET without body, cookie, authorization, cache validator or user agent, one connection, and close. Redirects to registered-looking, private, plaintext and hostname locations are classified and never contacted. Native platform validation behavior remains a DEV-455 gap.

Case plan: Record NET-01 stability check, platform validation capture, one request without credentials/cookies or redirect following, bounded comparison, and close. Test HTTP and HTTPS independently, with TLS only for HTTPS; keep a destination canary behind every redirect.

<!-- requirement-ref: REQ-NET-03-0005 -->
[REQ-NET-03-0005](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0005) → `M-NET-03-EXECUTION` (Operation).

<!-- requirement-ref: REQ-NET-03-0006 -->
[REQ-NET-03-0006](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0006) → `M-NET-03-EXECUTION` (Procedure).

#### M-NET-03-NORMALIZATION

Disposition: **passed** for synthetic production cases in `net03_conformance_test.go`, `symbol:TestMNet03Normalization`: every response class and the no-response, TCP, TLS and DNS failure paths, each with platform state `validated`, `not_validated`, `captive_portal` and `unknown`, select the outcome of the REQ-NET-03-0042 order; `platform_validation` is retained unchanged and `response_class` is present exactly when a valid response was received.

Case plan: Exercise every failure/outcome row and active-response class with the platform state varied independently, including conflicting/stale/unknown state. Preserve both evidence sources; test expected, mismatch, redirect and unrecognized responses without inventing an outcome precedence when the contract is ambiguous.

<!-- requirement-ref: REQ-NET-03-0007 -->
[REQ-NET-03-0007](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0007) → `M-NET-03-NORMALIZATION` (Success Condition).

<!-- requirement-ref: REQ-NET-03-0008 -->
[REQ-NET-03-0008](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0008) → `M-NET-03-NORMALIZATION` (Failure Conditions).

<!-- requirement-ref: REQ-NET-03-0027 -->
[REQ-NET-03-0027](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0027) → `M-NET-03-NORMALIZATION` (Outcomes).

<!-- requirement-ref: REQ-NET-03-0028 -->
[REQ-NET-03-0028](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0028) → `M-NET-03-NORMALIZATION` (Outcomes).

<!-- requirement-ref: REQ-NET-03-0041 -->
[REQ-NET-03-0041](https://github.com/tv-rain/barometer-spec/blob/f889a381c84a781176024f3fed3a3d4f5c92b241/tests/net-03.md#REQ-NET-03-0041) → `M-NET-03-NORMALIZATION` and `M-NET-03-LIMITS` (Outcomes). Valid responses and classes, with a digest undecidable within the prefix classified `unexpected`; GO-SPEC-ISSUE-006 records its adoption.

<!-- requirement-ref: REQ-NET-03-0042 -->
[REQ-NET-03-0042](https://github.com/tv-rain/barometer-spec/blob/f889a381c84a781176024f3fed3a3d4f5c92b241/tests/net-03.md#REQ-NET-03-0042) → `M-NET-03-NORMALIZATION` (Outcomes). The request selects the outcome first and platform state only without a response class; GO-SPEC-ISSUE-019 records its adoption.

#### M-NET-03-LIMITS

Disposition: **passed** for synthetic cases in `net03_conformance_test.go`, `symbol:TestMNet03Limits`: bodies of 0, 1,023, 1,024, 1,025 and 4,096 bytes with and without a digest (the digest undecidable within the prefix is `unexpected`, GO-SPEC-ISSUE-006 option (a)), header blocks of 16,383, 16,384 and 16,385 bytes (over the bound has no class), the exact counted receive budget and one byte less (`resource_limit`), one request with no retry, and a second run that waits without contact while the first is active. Overlapping runs are delayed, which REQ-NET-03-0021 permits.

Case plan: Test 16,384-byte response headers and each applicable registered response-prefix bound at below/equal/above boundaries, outer deadline, zero retries and one active run per generation. Count actual operations and buffered reads separately; reject overlapping runs as specified.

<!-- requirement-ref: REQ-NET-03-0021 -->
[REQ-NET-03-0021](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0021) → `M-NET-03-LIMITS` (Resource Bounds).

#### M-NET-03-VERSIONS

Disposition: **passed** for synthetic cases in `net03_conformance_test.go`, `symbol:TestMNet03Versions`: task-time rejection of unsupported, malformed and missing `test_version` with no contact or run; the emitted combination decodes; the result decoder rejects every unsupported or malformed version, NET-01, NET-05 and HTTP-01 identities, NET-01 and NET-05 shapes, missing or invalid NET-03 fields, a contradicting class, foreign outcomes and an unknown field.

Case plan: Assert the NET-03 registration emits exactly one combination: test.id NET-03, test.version 0.1.0, Measurement Result schema_version 0.1.0 and test.result_version 0.1.0. Reject before execution, as a task-validation failure that produces no Measurement Result, a NET-03 task naming test version 0.2.0, 1.0.0, 0.1, v0.1.0, 0.1.0 with surrounding whitespace, an empty string, a JSON number or a missing version. Reject in the result decoder, without best-effort interpretation or silent field dropping, results with schema_version or test.result_version set to each of the same unsupported or malformed values, results whose test.version and result_version are individually valid but differ from the registered combination, and cross-test results: test.id NET-03 with a NET-01 access-only Output or a NET-05 profile Output, and test.id NET-01 or NET-05 carrying the NET-03 Output (platform_validation and response_class). Assert that an unknown field inside a known 0.1.0 version is rejected, that no version is coerced, defaulted or added, and that the Draft status and registration metadata of tests/net-03.md do not by themselves cause any other version to be accepted. Requirements: REQ-NET-03-0022.

<!-- requirement-ref: REQ-NET-03-0022 -->
[REQ-NET-03-0022](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0022) → `M-NET-03-VERSIONS` (Output).

#### M-NET-03-OUTPUT

Disposition: **passed** for synthetic production output in `net03_conformance_test.go`, `symbol:TestMNet03Output` and `symbol:TestMNet03Execution`: the registered versions, the measurement field set, an expected response retained beside platform `captive_portal` without either rewriting the other, only registered `http_exchange` details, and the minimized redirect object with `followed_count` 0 and each location class. Every result is checked for the URL, address, location, body, digest, header, cookie, raw-error and certificate canaries.

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

Disposition: **passed** for synthetic cases in `internal/execution/net03_termination_test.go`, `symbol:TestNet03AllTerminationBoundaries`, and `net03_conformance_test.go`, `symbol:TestMNet03Termination`: for `https` and `http`, before the test starts is `not_run`; the stability check, the platform read, DNS, after DNS, TCP, TLS (`https` only), headers and the body prefix are `incomplete` with the first of host cancellation, `network_changed`, the attempt deadline, `resource_limit` and injected suspension, with no measurement, second request or open connection; release after the outcome is `completed`. A generation change during the platform read and a bounded protocol timeout are exercised through `Client`. Suspension is an injected engine cause, not device evidence.

Case plan: Drive NET-03 with a synthetic registered validation target, a fake platform validation source, a scripted DNS/TCP/TLS/HTTP transport and a deterministic clock, and inject each REQ-NET-03-0029 interruption at every Procedure point: during the NET-01 network-generation stability check; after it and before the platform validation/captive-portal state read; during that read; before and during the dns_query; before and during the tcp_connect; before and during the tls_handshake for an https URL only (no tls_handshake point exists for an http URL); after the request is sent and while response headers are read; while the bounded response prefix is read; and after the first response or failure but before the connection is closed. Expected states follow the GO-SPEC-ISSUE-004 disposition: the NET-01 network-generation stability check starts the test, so every point from it until the outcome is recorded is incomplete as asserted below; an interruption before it is not_run with no measurement and no network_context; and after the first response or failure but before the connection is closed the attempt stays completed with the recorded outcome. For each incomplete point iterate the five interruptions named by the requirement: a network-generation change (expect execution.state incomplete, reason network_changed), host cancellation (incomplete, cancelled_by_host), expiry of the outer attempt deadline as distinct from the bounded operation timeout (incomplete, attempt_deadline_exceeded), an outer non-time resource bound (incomplete, resource_limit) and platform suspension (incomplete, platform_suspended). Assert every such result has no measurement object at all (no outcome, observations, platform_validation or response_class, since NET-03 is not the NET-05 partial-result exception), carries exactly one reason naming the first termination condition, and that injecting two conditions keeps only the first. Assert that a bounded protocol timeout, DNS/TCP/TLS failure, redirect or content mismatch occurring without an outer interruption instead yields state completed with a measurement, that no retry, second request or redirect request is recorded after any interruption, that the connection and goroutines are released, and that no terminal protocol outcome or captured platform state is invented for an interrupted stage. Requirements: REQ-NET-03-0029.

<!-- requirement-ref: REQ-NET-03-0029 -->
[REQ-NET-03-0029](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0029) → `M-NET-03-TERMINATION` (Outcomes).

#### M-NET-03-MINIMIZATION

Intended owner: `internal/result/net03.go`. Planned case: `internal/result/net03_conformance_test.go`, `symbol:TestMNet03Minimization`. Disposition: **planned**, with no execution or support evidence.

Case plan: Configure a synthetic NET-03 task and scripted transport that place a distinct canary in every item REQ-NET-03-0030 names: the validation URL (scheme, host, path and query canaries), the resolved validation address, a 302 Location redirect URL with host, path, query and fragment canaries, the response body, the body SHA-256 digest (including the expected_body_sha256 input value), HTTP header names and header values (custom response headers), a Set-Cookie value, an authentication-state canary, a platform portal URL and interface name/index from the fake platform validation source, and a raw transport/platform error string for each DNS, TCP, TLS and HTTP failure path. Iterate validated, redirect, content_mismatch, unexpected, dns_failure, connection_failure, tls_failure, not_validated and incomplete paths, serialize the complete result and every nested observation and error path, and assert no canary appears in any byte of the encoding. Assert the measurement contains only outcome, observations, platform_validation and conditional response_class; observations contain only type, status, duration_ms, normalized failure and the registered details (dns_query engine/query_type/responses, tcp_connect address_family, tls_handshake registered fields, http_exchange address_family, protocol, status_code, redirect with followed_count 0 and terminal_location_class only, body bytes_processed/truncated, checks kind/index/matched, content_encoding); and that unknown or extra fields are rejected. Raw errors map to registered failure values or inconclusive. No participant or live measurement data is used. Requirements: REQ-NET-03-0030.

<!-- requirement-ref: REQ-NET-03-0030 -->
[REQ-NET-03-0030](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0030) → `M-NET-03-MINIMIZATION` (Outcomes).

#### M-NET-03-INTERPRETATION

Intended owner: `docs/measurement-interpretation-review.md`. Planned case: `internal/result/net03_conformance_test.go`, `symbol:TestMNet03Interpretation`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/measurement/net03/run.go`, `internal/result/net03.go`.

Case plan: Assert the serialized NET-03 measurement field set is exactly outcome, observations, platform_validation and conditional response_class, with outcome limited to validated, captive_portal_suspected, not_validated, dns_failure, connection_failure, tls_failure, unexpected_response and inconclusive, so no interpretation, verdict, censorship, operator or general-reachability field exists. Assert the Go field descriptions, documented enum descriptions and sanitized example and host-facing text for every outcome contain none of these prohibited claims: that validated proves arbitrary or general Internet destinations are reachable; that captive_portal_suspected, redirect or content_mismatch identifies the operator or purpose of an intercepted page; that not_validated or any unvalidated result is evidence of censorship; that a single result is definitive proof of censorship or a generic blocking, anomaly or success verdict; and that platform_validation overrides or confirms the active request. Exercise synthetic conflicting platform and active observations and assert outcome inconclusive rather than a stronger claim. Requirements: REQ-NET-03-0001, REQ-NET-03-0002, REQ-NET-03-0003, REQ-NET-03-0031, REQ-NET-03-0032, REQ-NET-03-0033.

Scope: portable automated component only; judgement-based completion is review case `RV-M-NET-03-INTERPRETATION`, which is not automated coverage.

Review case `RV-M-NET-03-INTERPRETATION` (`docs/measurement-interpretation-review.md`, `heading:RV-M-NET-03-INTERPRETATION`): Human review case; not automated coverage. Reviewer role: Measurement interpretation reviewer (analyst role independent of the implementer). Reviewed artifact: Go result documentation and host/analyst guidance for NET-03 at the reviewed implementation revision. Checklist: (1) Does the guidance limit a validated outcome to the configured validation endpoint on the observed network path, consistent with the Supported Interpretation? (2) Does the guidance list each Alternative Explanation (endpoint, DNS provider, CDN or certificate misconfiguration or unavailability; subscription, quota or local access-control pages; stale platform validation state; proxy, antivirus, enterprise gateway or VPN alteration; transient network change splitting observations)? (3) Does the guidance treat captive_portal_suspected as consistent with interception without attributing an operator, purpose or intent to the intercepting party? (4) Does the guidance require the REQ-NET-03-0032 evidence (a repeat on the same stable network generation, measurements against independently operated benign endpoints, same-window DNS/TCP/TLS/HTTP layer results, independent reference signals from a comparable network context) before any stronger interpretation? (5) Does the guidance present independent reference signals as triangulation rather than ground truth, noting differences in time, geography, vantage point and method? (6) Is the wording free of language that would let a reader treat a single NET-03 result as definitive proof of censorship? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-NET-03-0001 -->
[REQ-NET-03-0001](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0001) → `M-NET-03-INTERPRETATION`, `RV-M-NET-03-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-NET-03-0002 -->
[REQ-NET-03-0002](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0002) → `M-NET-03-INTERPRETATION`, `RV-M-NET-03-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-NET-03-0003 -->
[REQ-NET-03-0003](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0003) → `M-NET-03-INTERPRETATION`, `RV-M-NET-03-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-NET-03-0031 -->
[REQ-NET-03-0031](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0031) → `M-NET-03-INTERPRETATION`, `RV-M-NET-03-INTERPRETATION` (Supported Interpretation).

<!-- requirement-ref: REQ-NET-03-0032 -->
[REQ-NET-03-0032](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0032) → `M-NET-03-INTERPRETATION`, `RV-M-NET-03-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

<!-- requirement-ref: REQ-NET-03-0033 -->
[REQ-NET-03-0033](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0033) → `M-NET-03-INTERPRETATION`, `RV-M-NET-03-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

#### M-NET-03-SAFETY

Intended owner: `docs/measurement-deployment-review.md`. Planned case: `internal/admission/measurement_safety_conformance_test.go`, `symbol:TestMNet03Safety`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/execution/measurements.go`, `internal/result/net03.go`.

Case plan: Iterate the registered Baseline Risk rows (risk level low; requires user consent false, subject to host and local policy) and each Risk Factors row (network_observability low, traffic_distinctiveness low, target_sensitivity low, data_sensitivity low, legal_or_social_context low) with synthetic task safety declarations: assert effective risk is never below low, that a task declaring risk_level medium or high or any factor at medium or high raises the effective risk to at least that level, that an unknown factor keeps the overall level unknown unless trusted local assessment resolves it, that requires_user_consent true blocks execution until the fake host establishes consent, and that consent false adds no requirement while host or local policy may still decline. Exercise each normative safety action with synthetic targets and a recording transport: data minimization (no response body, full redirect URL, cookie or unrelated header in the result); traffic distinctiveness (exactly one validation request per run, zero retries, and no added repeated requests when the fake platform reports reliable state); user identification (the recorded request carries no identifiers, cache validators, cookies or application authentication); target risk (a target absent from the locally registered benign validation set is rejected and never contacted, and a redirect target is never contacted); and probe authority (fake local policy can delay, skip or reject execution, yielding not_run with local_policy_rejected when rejected). Requirements: REQ-NET-03-0034, REQ-NET-03-0035, REQ-NET-03-0036, REQ-NET-03-0037, REQ-NET-03-0038, REQ-NET-03-0039, REQ-NET-03-0040.

Scope: portable automated component only; judgement-based completion is review case `RV-M-NET-03-SAFETY`, which is not automated coverage.

Review case `RV-M-NET-03-SAFETY` (`docs/measurement-deployment-review.md`, `heading:RV-M-NET-03-SAFETY`): Human review case; not automated coverage. Reviewer role: Deployment safety and privacy reviewer (independent of the implementer, with host-integration and local legal/social context knowledge). Reviewed artifact: NET-03 host integration, native Android and Apple platform-validation adapters, scheduling configuration and validation target set at the reviewed deployment revision. Checklist: (1) Is every configured validation target controlled by the adopter or another explicitly approved operator, safety-reviewed and non-sensitive? (2) Does host scheduling avoid synchronized schedules and repeated validation requests when the platform already provides reliable state? (3) Has the repeated-execution risk of a fixed endpoint and cadence been assessed for the deployment, and is the residual risk that operators observe the destination and tailor behavior to known validation endpoints documented for adopters? (4) Do the native platform-validation adapters avoid exposing or retaining the platform portal URL, interface information or raw platform errors to the core result? (5) Has the host determined whether host or local policy requires consent despite the registered requires_user_consent false baseline? (6) Has a local deployment review of legal or social context been completed for each deployment region? (7) Does observed on-device traffic from the integrated build contain no identifiers, cookies, cache validators or application authentication on the validation request? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-NET-03-0034 -->
[REQ-NET-03-0034](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0034) → `M-NET-03-SAFETY`, `RV-M-NET-03-SAFETY` (Baseline Risk).

<!-- requirement-ref: REQ-NET-03-0035 -->
[REQ-NET-03-0035](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0035) → `M-NET-03-SAFETY`, `RV-M-NET-03-SAFETY` (Risk Factors).

<!-- requirement-ref: REQ-NET-03-0036 -->
[REQ-NET-03-0036](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0036) → `M-NET-03-SAFETY`, `RV-M-NET-03-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-NET-03-0037 -->
[REQ-NET-03-0037](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0037) → `M-NET-03-SAFETY`, `RV-M-NET-03-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-NET-03-0038 -->
[REQ-NET-03-0038](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0038) → `M-NET-03-SAFETY`, `RV-M-NET-03-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-NET-03-0039 -->
[REQ-NET-03-0039](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0039) → `M-NET-03-SAFETY`, `RV-M-NET-03-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-NET-03-0040 -->
[REQ-NET-03-0040](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-03.md#REQ-NET-03-0040) → `M-NET-03-SAFETY`, `RV-M-NET-03-SAFETY` (Safety Analysis).

### NET-05

The production `internal/measurement/net05` codec and root registry composition provide decoding, composite target/proxy rules and a sequential runner. `TestMNet05Admission`, `TestMNet05Options`, `TestMNet05Versions` and `TestMNet05AtomicTargetAndProxyRules` in `net05_conformance_test.go` exercise the production `Client.AdmitTask` path with synthetic policy, registrations and recording transports. `Client.StartRun` uses the production runner, shared protocol operations, a single aggregate meter and the existing strict NET-05 result boundary. Each step receives only its exact primitive version and initial tuples; HTTP redirects cannot borrow another profile step's links.

Coverage is partial. The task owner explicitly approved two deferrals: `profile_version` is a positive signed 64-bit value without cross-task history or monotonicity enforcement (REQ-NET-05-0015); admission does not sum primitive duration, connection or byte maxima or prove complete-envelope fit (REQ-NET-05-0026, REQ-LIMIT-0020). The latter remains a temporary specification deviation pending clarification and implementation. Primitive schema/option bounds still apply, including HTTP body size against the outer receive bound; the common input cap also bounds the entire wrapped profile. Runtime aggregate duration, connection and application-byte enforcement and bounded result emission are exercised synthetically. Platform validation is required by default: known negative validation yields terminal `network_not_validated`, unknown yields terminal `inconclusive`, and neither starts a primitive. Only a trusted internal policy exception permits known negative validation; unknown still never runs. Task `require_validated` remains an earlier constraint and produces no NET-05 measurement when unsatisfied. Validation source is `platform`; no fresh/cached NET-03 acquisition or connectivity probe is implemented. Traffic-pattern reviews, platform and full conformance evidence remain pending. At Go `1a37603` the OPTIONS, NORMALIZATION, OUTPUT, TERMINATION and VERSIONS families are `passed`; ADMISSION, EXECUTION and LIMITS stay `planned` for these deferrals, owned by [DEV-658](https://app.clickup.com/t/124kk1zcw4g), with their executed rows recorded as unit-test cases.

#### M-NET-05-ADMISSION

Case: `net05_conformance_test.go`, `symbol:TestMNet05Admission`. Disposition: **planned**; the family stays planned because two plan rows are approved deferrals owned by [DEV-658](https://app.clickup.com/t/124kk1zcw4g): nonmonotone profile versions (REQ-NET-05-0015) and over-budget summed declared maxima (REQ-NET-05-0026). Executed rows, passed at Go `1a37603` as unit-test cases [`UT.root.TestMNet05Admission`](../traceability/evidence/go-unit-1a37603.md#UT.root.TestMNet05Admission), [`UT.root.TestMNet05AtomicTargetAndProxyRules`](../traceability/evidence/go-unit-1a37603.md#UT.root.TestMNet05AtomicTargetAndProxyRules), [`UT.root.TestMNet05CodecBoundsAndComposition`](../traceability/evidence/go-unit-1a37603.md#UT.root.TestMNet05CodecBoundsAndComposition): profile, identifier, step-count and encoded-size boundaries, closed primitive and step schemas, unsafe or invalid final steps, exact registration scope, path, expiry and issuer failures, composite enablement and whole-profile proxy rejection.

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

Case: `net05_conformance_test.go`, `symbol:TestMNet05Options`, with the evidence in [its record](../traceability/evidence/go-unit-1a37603.md#M-NET-05-OPTIONS). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Accept only empty composite options. Independently vary each primitive option via its exact versioned schema; reject attempts to override sequence, outer execution/safety, primitive limits or no-retry behavior.

<!-- requirement-ref: REQ-NET-05-0028 -->
[REQ-NET-05-0028](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0028) → `M-NET-05-OPTIONS` (Options).

<!-- requirement-ref: REQ-NET-05-0029 -->
[REQ-NET-05-0029](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0029) → `M-NET-05-OPTIONS` (Options).

<!-- requirement-ref: REQ-NET-05-0030 -->
[REQ-NET-05-0030](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0030) → `M-NET-05-OPTIONS` (Options).

#### M-NET-05-PRECONDITIONS

Intended owner: `internal/measurement/net05/run.go`. Planned case: `internal/measurement/net05/conformance_test.go`, `symbol:TestMNet05Preconditions`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/platform/network.go`, `docs/measurement-platform-validation.md`.

Case plan: Run the portable NET-05 gate with a synthetic host policy, synthetic inline profile, fake platform and recording transport. Iterate single-precondition failures and assert execution.state not_run with no profile step started: task source or inline profile untrusted by host policy; any target failing local privacy and safety evaluation; NET-05 neither explicitly enabled by host policy nor triggered by repeated baseline anomalies under that policy; network not stable; validation or captive-portal context unavailable; outer execution envelope unable to accommodate the complete inline profile within primitive limits (for example step durations exceeding max_duration_ms). Assert envelope fit never weakens primitive limits. Inject host-policy withdrawal after step 1 and assert no further steps start. Assert a stricter execution constraint or local policy denial gives constraints_unsatisfied or local_policy_rejected; other denials use a registered reason per the public table. Requirements: REQ-NET-05-0031, REQ-NET-05-0032, REQ-NET-05-0033, REQ-NET-05-0034, REQ-NET-05-0035, REQ-NET-05-0036.

Scope: portable automated component only; judgement-based completion is review case `RV-M-NET-05-PRECONDITIONS`, which is not automated coverage.

Review case `RV-M-NET-05-PRECONDITIONS` (`docs/measurement-deployment-review.md`, `heading:RV-M-NET-05-PRECONDITIONS`): Human review case; not automated coverage. Reviewer role: Measurement safety reviewer (independent of the implementer and profile author). Reviewed artifact: NET-05 traffic-pattern gate logic, host enablement and anomaly-trigger policy guidance, and registered example profiles at the reviewed implementation revision. Checklist: (1) Does the gate define and apply a criterion for unusually dense, synchronized or identifiable traffic patterns before any step starts? (2) Is it established that profiles sharing a schedule across many devices cannot synchronize into an identifiable pattern? (3) Does host guidance require reviewed policy before enabling NET-05 or anomaly-triggered runs? (4) Is the repeated-baseline-anomaly trigger bounded so that it cannot escalate into dense execution? (5) Are residual distinctiveness risks documented as limitations? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-NET-05-0031 -->
[REQ-NET-05-0031](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0031) → `M-NET-05-PRECONDITIONS`, `RV-M-NET-05-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-NET-05-0032 -->
[REQ-NET-05-0032](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0032) → `M-NET-05-PRECONDITIONS`, `RV-M-NET-05-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-NET-05-0033 -->
[REQ-NET-05-0033](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0033) → `M-NET-05-PRECONDITIONS`, `RV-M-NET-05-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-NET-05-0034 -->
[REQ-NET-05-0034](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0034) → `M-NET-05-PRECONDITIONS`, `RV-M-NET-05-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-NET-05-0035 -->
[REQ-NET-05-0035](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0035) → `M-NET-05-PRECONDITIONS`, `RV-M-NET-05-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-NET-05-0036 -->
[REQ-NET-05-0036](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0036) → `M-NET-05-PRECONDITIONS`, `RV-M-NET-05-PRECONDITIONS` (Preconditions).

#### M-NET-05-EXECUTION

Case: `net05_runner_test.go`, `symbol:TestMNet05ExecutionAndOutput`. Disposition: **planned**; the family stays planned because fresh or cached NET-03 validation context is an approved deferral owned by [DEV-658](https://app.clickup.com/t/124kk1zcw4g); only platform validation context is implemented. Executed rows, passed at Go `1a37603` as unit-test cases [`UT.root.TestMNet05ExecutionAndOutput`](../traceability/evidence/go-unit-1a37603.md#UT.root.TestMNet05ExecutionAndOutput), [`UT.root.TestNet05DirectDNSPrimitives`](../traceability/evidence/go-unit-1a37603.md#UT.root.TestNet05DirectDNSPrimitives), [`UT.root.TestNet05TLSPrimitiveSuccess`](../traceability/evidence/go-unit-1a37603.md#UT.root.TestNet05TLSPrimitiveSuccess), [`UT.root.TestNet05StepCountExecution`](../traceability/evidence/go-unit-1a37603.md#UT.root.TestNet05StepCountExecution), [`UT.root.TestMNet05PrimitiveRiskAndMissingCapability`](../traceability/evidence/go-unit-1a37603.md#UT.root.TestMNet05PrimitiveRiskAndMissingCapability), [`UT.internal.execution.TestNet05WholeProfileSerialization`](../traceability/evidence/go-unit-1a37603.md#UT.internal.execution.TestNet05WholeProfileSerialization), [`UT.internal.execution.TestNet05StepTargetAndRedirectIsolation`](../traceability/evidence/go-unit-1a37603.md#UT.internal.execution.TestNet05StepTargetAndRedirectIsolation): whole-profile admission and effective risk before any connection, platform validation context and strictly sequential DNS-01, TCP-01, TLS-01 and HTTP-01 dispatch through shared protocol operations, with no overlap, retry or borrowed redirect authority. Deployment reviews remain deferred.

Case plan: Record admission of the entire profile and effective risk before any connection, current NET-01, permitted fresh/cached/platform NET-03 context and strictly sequential DNS-01/TCP-01/TLS-01/HTTP-01 dispatch. Each primitive must use its own schema and shared protocol operations, with no parallel steps, fallback, retries, target expansion or combined verdict.

<!-- requirement-ref: REQ-NET-05-0006 -->
[REQ-NET-05-0006](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0006) → `M-NET-05-EXECUTION` (Operation).

<!-- requirement-ref: REQ-NET-05-0007 -->
[REQ-NET-05-0007](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0007) → `M-NET-05-EXECUTION` (Operation).

<!-- requirement-ref: REQ-NET-05-0008 -->
[REQ-NET-05-0008](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0008) → `M-NET-05-EXECUTION` (Operation).

<!-- requirement-ref: REQ-NET-05-0009 -->
[REQ-NET-05-0009](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0009) → `M-NET-05-EXECUTION` (Procedure).

#### M-NET-05-NORMALIZATION

Case: `internal/execution/net05_conformance_test.go`, `symbol:TestMNet05Normalization`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-NET-05-NORMALIZATION). Disposition: **passed** at Go `1a37603`; partial evidence only, with no platform or full-conformance claim.

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

Case: `internal/execution/net05_conformance_test.go`, `symbol:TestMNet05Limits`. Disposition: **planned**; the family stays planned because rejecting over-budget declared maxima atomically before contact (REQ-NET-05-0026, REQ-LIMIT-0020) is an approved deferral owned by [DEV-658](https://app.clickup.com/t/124kk1zcw4g). Executed rows, passed at Go `1a37603` as unit-test cases [`UT.root.TestMNet05RegisteredCeilingAdmission`](../traceability/evidence/go-unit-1a37603.md#UT.root.TestMNet05RegisteredCeilingAdmission), [`UT.internal.execution.TestMNet05Limits`](../traceability/evidence/go-unit-1a37603.md#UT.internal.execution.TestMNet05Limits), [`UT.internal.execution.TestMNet05AggregateBytes`](../traceability/evidence/go-unit-1a37603.md#UT.internal.execution.TestMNet05AggregateBytes), [`UT.internal.execution.TestMNet05AggregateDeadline`](../traceability/evidence/go-unit-1a37603.md#UT.internal.execution.TestMNet05AggregateDeadline), [`UT.internal.execution.TestMNet05ConnectionCeiling`](../traceability/evidence/go-unit-1a37603.md#UT.internal.execution.TestMNet05ConnectionCeiling), [`UT.internal.execution.TestMNet05ReceivedCeiling`](../traceability/evidence/go-unit-1a37603.md#UT.internal.execution.TestMNet05ReceivedCeiling), [`UT.internal.execution.TestMNet05DurationCeiling`](../traceability/evidence/go-unit-1a37603.md#UT.internal.execution.TestMNet05DurationCeiling), [`UT.internal.execution.TestNet05WholeProfileSerialization`](../traceability/evidence/go-unit-1a37603.md#UT.internal.execution.TestNet05WholeProfileSerialization): the 60,000 ms, 12-connection, 65,536-byte sent and 262,144-byte received ceilings below, at and above at admission; the duration, connection and received-byte ceilings and stricter task bounds against actual aggregate use on direct and CONNECT paths; and one sequential profile run per runner. The sent ceiling is unreachable at run time and the profile bound is decided by the common input bound (GO-SPEC-ISSUE-018), as the plan records.

Case plan: Test 49,152 encoded profile bytes, 60,000 ms, 12 connections/tunnels, 65,536 application bytes sent and 262,144 received at below/equal/above boundaries and against stricter primitive/task bounds. Reject over-budget declared maxima atomically before contact; enforce actual aggregate use and one sequential profile run per generation. Registered Draft primitives cannot reach the 65,536-byte sent ceiling at run time (no request body and at most eight bounded GET requests), so that ceiling is tested at admission only; the common input bound decides before the equal profile bound (GO-SPEC-ISSUE-018), so the 49,152-byte profile boundary is tested only for rejection.

<!-- requirement-ref: REQ-NET-05-0037 -->
[REQ-NET-05-0037](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0037) → `M-NET-05-LIMITS` (Resource Bounds).

<!-- requirement-ref: REQ-NET-05-0038 -->
[REQ-NET-05-0038](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0038) → `M-NET-05-LIMITS` (Resource Bounds).

#### M-NET-05-VERSIONS

Case: `net05_conformance_test.go`, `symbol:TestMNet05Versions`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-NET-05-VERSIONS). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Select the exact registered combination for NET-05: test.id NET-05, test.version 0.1.0, Measurement Result schema_version 0.1.0 and test.result_version 0.1.0, and for every dispatched primitive step test_id DNS-01, TCP-01, TLS-01 or HTTP-01 with test_version 0.1.0 and emitted step test.result_version 0.1.0. Assert a valid profile produces exactly these strings in the outer envelope and in each steps[].test object. Iterate rejection inputs using a recording transport: outer NET-05 version 0.1.1, 0.2.0, 1.0.0, empty string, v0.1.0 and non-string values; each step test_version set to an unsupported, empty, padded or non-string value while other steps stay valid; a step test_id of NET-05 (nested), NET-01, NET-03, DNS-03 and an unregistered ID; a DNS-01 step whose input/options belong to HTTP-01 (cross-test input); and an emitted result whose schema_version, outer result_version or a step result_version differs from 0.1.0. Assert every unsupported task or step version, malformed version and cross-test combination is a task-validation failure reported outside Measurement Results, that the whole profile is rejected before any primitive step or connection (zero recorded contacts; no compatible subset runs), and that no version is guessed, coerced, defaulted or added. Assert a receiver validator refuses best-effort interpretation of a result carrying an unknown schema, test or result version. Registration metadata and Draft status confer no support claim. Requirements: REQ-NET-05-0039.

<!-- requirement-ref: REQ-NET-05-0039 -->
[REQ-NET-05-0039](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0039) → `M-NET-05-VERSIONS` (Output).

#### M-NET-05-OUTPUT

Case: `net05_runner_test.go`, `symbol:TestMNet05ExecutionAndOutput`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-NET-05-OUTPUT). Disposition: **passed** at Go `1a37603`; the existing `internal/result` boundary validates exact profile and step identity, version and output shape before persistence. No support claim follows from synthetic output evidence.

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

Case: `internal/execution/net05_conformance_test.go`, `symbol:TestMNet05Termination`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-NET-05-TERMINATION). Disposition: **passed** at Go `1a37603`. Each common reason is injected before, during and after each step; final terminal outcomes survive cleanup. Platform suspension is a synthetic event, not native lifecycle evidence.

Case plan: Use a validated synthetic three-step profile (DNS-01, TCP-01, HTTP-01 steps with synthetic .test targets) driven by scripted per-primitive transports, a deterministic clock and a fake platform that can raise a network-generation change, host cancellation, outer attempt-deadline expiry, non-time connection/data bound exhaustion, platform suspension and proxy reconfiguration. Iterate each interruption point: (a) after NET-01/NET-03 context capture and before step 1; (b) before each later step k; (c) inside each started step k while its primitive operation is in flight; (d) after the final step reaches a terminal primitive outcome. Expected states follow the GO-SPEC-ISSUE-004 disposition: point (a) follows the stable-network check of Procedure step 3, so points (a) to (c) are incomplete as asserted below; a task-level require_validated constraint that is unsatisfied before that check is not_run with constraints_unsatisfied and no measurement; and the composite outcomes completed, network_not_validated and inconclusive are retained with execution.state completed while composite outcome incomplete keeps execution.state incomplete with the stopping reason. For each of the common reasons network_changed, resource_limit, attempt_deadline_exceeded, cancelled_by_host, platform_suspended and proxy_reconfigured, assert execution.state is incomplete with exactly that reason (only the first cross-cutting condition is recorded) and measurement.outcome is incomplete; for network_changed and resource_limit assert no network_changed or resource_limit outcome value is emitted, only outcome incomplete plus the matching execution.reason. Assert permitted partial profile output at (a)/(b): steps contains only the earlier steps, each with state completed and its primitive measurement, no placeholder entry for unstarted steps, and stopped_before_step equals the step_id of step k; at (c): step k is the single final entry with state incomplete and no measurement object, and stopped_before_step equals the next profile step_id when one remains and is absent when step k is the last; at (d): outcome completed, execution.state completed, no reason, no stopped_before_step. Assert a primitive protocol timeout, refusal or TLS failure scripted inside a step yields a completed step with the primitive negative outcome rather than an interruption, and that nested step measurement objects never carry their own common envelope, execution object or Network Context while one outer Network Context remains present. Verify after each interruption that the recording transport shows no further connection, retry, fallback to the direct path, new proxy endpoint or mixed proxy generation, that all sockets and timers are released, and that no terminal protocol outcome is invented for the interrupted or unstarted steps. Requirements: REQ-NET-05-0047, REQ-NET-05-0061.

<!-- requirement-ref: REQ-NET-05-0047 -->
[REQ-NET-05-0047](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0047) → `M-NET-05-TERMINATION` (Outcomes).

<!-- requirement-ref: REQ-NET-05-0061 -->
[REQ-NET-05-0061](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/tests/net-05.md#REQ-NET-05-0061) → `M-NET-05-TERMINATION` (Outcomes).

#### M-NET-05-MINIMIZATION

Intended owner: `internal/result/net05.go`. Planned case: `internal/result/net05_conformance_test.go`, `symbol:TestMNet05Minimization`. Disposition: **planned**, with no execution or support evidence.

Case plan: Run a synthetic profile covering DNS-01, TCP-01, TLS-01 and HTTP-01 steps through scripted transports, inserting one unique canary per prohibited item named by the linked unit: an outer task schedule canary, an outer safety-declaration canary and an outer labels canary (nested task envelope, schedule, safety and labels must not be copied); a canary hostname and a canary HTTPS URL path in step inputs (duplicated targets and target URLs); a canary TCP-01 target address from 192.0.2.0/24 (IP addresses beyond bounded DNS answer values); a canary profile field value such as a step option value (profile contents); a canary response-body string and a canary HTTP header value returned by the scripted HTTP server; a canary certificate subject and serial from the synthetic TLS chain (certificate data); and a canary platform error string raised by the fake platform (platform errors). Assert serialized output, including every nested step measurement and observation and any result-validation error path, contains none of these canaries, no nested schedule, execution, safety or labels object, no per-step Network Context and no field named or valued as a combined restriction verdict (for example broad_restriction_suspected or allowlist_detected). Assert the only retained NET-05 fields are outcome, observations, profile.profile_id, profile.profile_version, network_validation.state, network_validation.source, steps[].step_id, steps[].test.id, steps[].test.version, steps[].test.result_version, steps[].state, steps[].measurement (the primitive Output object as registered by that primitive) and stopped_before_step when required; the DNS-01 step may retain bounded normalized DNS answer values (a synthetic 198.51.100.0/24 answer is present only in the DNS answers path). Reject any result carrying an extra field before bundling. No participant or live measurement data is used. Requirements: REQ-NET-05-0048.

<!-- requirement-ref: REQ-NET-05-0048 -->
[REQ-NET-05-0048](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0048) → `M-NET-05-MINIMIZATION` (Outcomes).

#### M-NET-05-INTERPRETATION

Intended owner: `docs/measurement-interpretation-review.md`. Planned case: `internal/result/net05_conformance_test.go`, `symbol:TestMNet05Interpretation`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/measurement/net05/run.go`, `internal/result/net05.go`.

Case plan: Assert the serialized NET-05 result for completed, incomplete, network_not_validated and inconclusive synthetic runs contains no interpretation, verdict, classification, anomaly, censorship, allowlist or broad-restriction field, and that outcome takes only completed, network_not_validated, incomplete or inconclusive, never broad_restriction_suspected, allowlist_detected or another censorship classification, including when every step reports a primitive negative outcome. Assert Go field documentation, the result-type doc comments and sanitized example and host-facing text for NET-05 contain none of these prohibited claims: that the result proves an allowlist policy exists or identifies who operates it; that combined results are classified as a broad restriction, allowlist or censorship event; that a profile tests sensitive, political, circumvention or broad public website lists; that NET-05 replaces outage, captive-portal, routing or independent reference analysis; that correlated observations establish censor intent or policy with certainty; and that external datasets are ground truth. Verify documentation states that one complete execution establishes only the ordered primitive results on one network generation, that correlation with target knowledge, repeated runs, controls and independent signals is outside NET-05 version 0.1.0, and that the listed alternative explanations (simultaneous outage or misconfiguration of external targets, shared hidden CDN/DNS/routing/certificate dependencies, transit failure or congestion, captive portal/subscription/VPN/enterprise policy/local firewall, stale profile definitions or changed endpoint behavior, address-family failure or mobile handover) are present. Requirements: REQ-NET-05-0001, REQ-NET-05-0002, REQ-NET-05-0003, REQ-NET-05-0004, REQ-NET-05-0049, REQ-NET-05-0050, REQ-NET-05-0051, REQ-NET-05-0052.

Scope: portable automated component only; judgement-based completion is review case `RV-M-NET-05-INTERPRETATION`, which is not automated coverage.

Review case `RV-M-NET-05-INTERPRETATION` (`docs/measurement-interpretation-review.md`, `heading:RV-M-NET-05-INTERPRETATION`): Human review case; not automated coverage. Reviewer role: Measurement interpretation reviewer (analyst role independent of the implementer). Reviewed artifact: Go result documentation and host/analyst guidance for NET-05 at the reviewed implementation revision. Checklist: (1) Does the guidance avoid presenting failures across several profile steps as evidence of an allowlist, broad restriction or censorship event without the stronger-evidence criteria? (2) Does the guidance require repeated complete profiles on stable network generations under a conservative schedule before stronger interpretation? (3) Does the guidance require consistent failure across independently operated targets and multiple protocol layers, and acknowledge that targets may share hidden CDN, DNS, routing or certificate dependencies? (4) Does the guidance require continued reachability of suitable local or host-application controls and the absence of a captive-portal or obvious local-network explanation? (5) Does the guidance treat independent reference signals as supporting comparisons with documented differences in time, network and geographic context, not as ground truth? (6) Does the guidance keep correlation and any derived classification outside NET-05 version 0.1.0 and state that intent or policy cannot be established with certainty? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-NET-05-0001 -->
[REQ-NET-05-0001](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0001) → `M-NET-05-INTERPRETATION`, `RV-M-NET-05-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-NET-05-0002 -->
[REQ-NET-05-0002](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0002) → `M-NET-05-INTERPRETATION`, `RV-M-NET-05-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-NET-05-0003 -->
[REQ-NET-05-0003](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0003) → `M-NET-05-INTERPRETATION`, `RV-M-NET-05-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-NET-05-0004 -->
[REQ-NET-05-0004](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0004) → `M-NET-05-INTERPRETATION`, `RV-M-NET-05-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-NET-05-0049 -->
[REQ-NET-05-0049](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0049) → `M-NET-05-INTERPRETATION`, `RV-M-NET-05-INTERPRETATION` (Supported Interpretation).

<!-- requirement-ref: REQ-NET-05-0050 -->
[REQ-NET-05-0050](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0050) → `M-NET-05-INTERPRETATION`, `RV-M-NET-05-INTERPRETATION` (Supported Interpretation).

<!-- requirement-ref: REQ-NET-05-0051 -->
[REQ-NET-05-0051](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0051) → `M-NET-05-INTERPRETATION`, `RV-M-NET-05-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

<!-- requirement-ref: REQ-NET-05-0052 -->
[REQ-NET-05-0052](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0052) → `M-NET-05-INTERPRETATION`, `RV-M-NET-05-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

#### M-NET-05-SAFETY

Intended owner: `docs/measurement-deployment-review.md`. Planned case: `internal/admission/measurement_safety_conformance_test.go`, `symbol:TestMNet05Safety`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/execution/measurements.go`, `internal/result/net05.go`.

Case plan: Iterate the Baseline Risk rows using synthetic targets and recording transports: risk level medium and requires user consent false subject to host and local policy; assert a task declaring low is rejected or evaluated as at least medium, a task declaring high raises effective risk to high, and a host/local policy requiring consent blocks execution without consent despite the false baseline. Iterate each Risk Factors row (network_observability medium, traffic_distinctiveness medium, target_sensitivity low, data_sensitivity medium, legal_or_social_context medium), asserting effective risk is the most restrictive of NET-05, every referenced primitive (DNS-01, TCP-01, TLS-01, HTTP-01, each registered low), the task declaration and local policy, and that an unknown factor type is rejected or treated as unknown rather than ignored. Exercise each normative safety action separately: data minimization (retain primitive classifications and safe step identifiers only, no duplicated raw target lists, response bodies, full addresses or detailed network identity); traffic distinctiveness (recording transport shows strictly sequential steps, no automatic retries, no port variation, no target expansion, and broad lists rejected; no synchronized fleet-wide cadence is introduced by the library); user-identification risk (no per-installation identifier and no unrelated device or application telemetry in output or traffic); target risk (each inline target independently passes schema and local safety policy, and one rejected target rejects the whole profile before contact); repeated-execution risk (an overlapping second profile run on the same network generation is delayed or rejected and scheduling remains host/local controlled); and final probe authority (local policy can delay, skip or reject, producing local_policy_rejected with no measurement traffic). Requirements: REQ-NET-05-0053, REQ-NET-05-0054, REQ-NET-05-0055, REQ-NET-05-0056, REQ-NET-05-0057, REQ-NET-05-0058, REQ-NET-05-0059, REQ-NET-05-0060.

Scope: portable automated component only; judgement-based completion is review case `RV-M-NET-05-SAFETY`, which is not automated coverage.

Review case `RV-M-NET-05-SAFETY` (`docs/measurement-deployment-review.md`, `heading:RV-M-NET-05-SAFETY`): Human review case; not automated coverage. Reviewer role: Deployment safety reviewer (privacy and safety role independent of the implementer). Reviewed artifact: Host integration, native adapter capabilities, NET-05 profile target composition and scheduling policy for the reviewed deployment revision. Checklist: (1) Does every profile target use a benign endpoint with documented ownership and operational expectations, and is target composition free of sensitive, political, circumvention or broad public lists? (2) Is the host and local scheduling policy conservative enough that repeated NET-05 runs do not create a synchronized fleet-wide cadence or an unusually dense, identifiable traffic pattern? (3) Is NET-05 enabled only by reviewed host policy or by repeated baseline anomalies under that policy? (4) Has the deployment assessed residual risks that a network observer may correlate the destination sequence and that combined results may characterize a user's access environment, for the intended deployment contexts? (5) Do native platform capabilities and observed real traffic on target platforms match the sequential, no-retry, no-port-variation behavior asserted by the portable tests? (6) Do host consent handling and legal or social context review cover the coordinated multi-target behavior, without blanket legal conclusions? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-NET-05-0053 -->
[REQ-NET-05-0053](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0053) → `M-NET-05-SAFETY`, `RV-M-NET-05-SAFETY` (Baseline Risk).

<!-- requirement-ref: REQ-NET-05-0054 -->
[REQ-NET-05-0054](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0054) → `M-NET-05-SAFETY`, `RV-M-NET-05-SAFETY` (Risk Factors).

<!-- requirement-ref: REQ-NET-05-0055 -->
[REQ-NET-05-0055](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0055) → `M-NET-05-SAFETY`, `RV-M-NET-05-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-NET-05-0056 -->
[REQ-NET-05-0056](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0056) → `M-NET-05-SAFETY`, `RV-M-NET-05-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-NET-05-0057 -->
[REQ-NET-05-0057](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0057) → `M-NET-05-SAFETY`, `RV-M-NET-05-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-NET-05-0058 -->
[REQ-NET-05-0058](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0058) → `M-NET-05-SAFETY`, `RV-M-NET-05-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-NET-05-0059 -->
[REQ-NET-05-0059](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0059) → `M-NET-05-SAFETY`, `RV-M-NET-05-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-NET-05-0060 -->
[REQ-NET-05-0060](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/net-05.md#REQ-NET-05-0060) → `M-NET-05-SAFETY`, `RV-M-NET-05-SAFETY` (Safety Analysis).

### DNS-01

#### Current DNS-01 Execution

`internal/measurement/dns01` owns the codec, parameters, proxy rule and runner; root `registry.go` composes its exact `0.1.0` entry and separate query-name/resolver target rules. `internal/protocol/dns` owns one bounded query and strict wire parsing. Execution supplies the checked/metered transport, clock, stable-generation check and terminal-outcome marker. Equivalent hostname/type/mode/resolver tuples serialize under their own deadlines, conservatively across network generations too.

Names use nontransitional UTS #46 lookup mapping with STD3, label, hyphen, joiner, bidi and DNS-length checks (`golang.org/x/net/idna` v0.61.0, Unicode 17.0.0; `x/text` v0.43.0). They become lowercase ASCII without a trailing dot and fit 253 bytes. Root/single-label, local/search and private namespaces are refused. Wire queries request A or AAAA with a fresh unpredictable ID, ordinary recursion desired and no identifying metadata.

Runtime JSON inventory: DNS-01 input/options are consumed only by the existing `taskdecode`/`jsondecode` stages, then normalized by the codec. The runner produces typed `observation.NewDNS` values, validated by the existing strict observation and result decoders; there is no independent runtime JSON parser. `TestDns01StrictSchemaBothOrigins`, the codec matrix, result-output mutations and `FuzzDns01Admission` provide caller-specific strict-JSON adoption evidence. The binary DNS parser has separate hostile-wire tests and `FuzzParse`. This is the DNS-01 share of the GO-SPEC-ISSUE-001 [Registered-Schema Adoption Addendum](specification-issues.md#registered-schema-adoption-addendum--october-27-2026); binding/host adoption remains open.

Each direct operation has a 15-second bound and sends one UDP query or one TCP length-prefixed query, without retry or fallback. UDP source, ID and question must match; unrelated packets are discarded without overriding usable replies. Additional collection starts after the first matching reply and stops at three. The shared REQ-LIMIT-0014 maximum of **1,000 ms** is stricter than the test option table's 10,000 ms and controls admission without clamping. Above-maximum `task_limit_exceeded` follows the approved interim handling in [GO-SPEC-ISSUE-016](specification-issues.md#go-spec-issue-016-unnamed-classes-and-order-in-step-8-input-and-options); this is not an exact-class conformance claim.

The parser validates sizes/counts, bounded backward compression pointers, complete frame/section lengths, correlation and EDNS option framing/effective response codes. It retains public requested-type answers and intervening CNAMEs only on the query's chain, removing exact duplicates in original order. Answers are never contacted. TTLs, raw messages, query names, unrelated records and resolver metadata are not retained. Unsafe or over-limit normalization yields `malformed_response` without partially appending answers. Observation validation enforces 16 answers/eight CNAMEs per response and the 16,384-byte encoded ceiling. Conflict outranks resolution; the three-response cap outranks conflict.

System mode invokes one ordinary resolver API call for the fully qualified name. It emits no response records, codes, address-family metadata or datagram timing that the API did not expose. Hidden DNS messages are not metered. Success/no-data is exposed without retained answer values. Go's `IsNotFound` conflates NXDOMAIN and no-data and maps to `inconclusive`; typed timeouts map to `timeout`, other API errors to `resolver_error`. The narrow trusted capability must honor cancellation and retain no Barometer-owned query/result state. Ordinary-resolver selection, search suppression, hidden retries/CNAME handling, cancellation fidelity, resource behavior and error distinctions still require Android/Apple evidence. Synthetic API tests establish no native support.

| Family | Executable synthetic coverage |
|---|---|
| `M-DNS-01-ADMISSION` | `dns01_conformance_test.go:TestMDns01Admission`, `TestDns01NormalizedNamesAndResolverBoundaries`, `TestDns01StrictSchemaBothOrigins` and `internal/measurement/dns01/codec_test.go`: normalized name/type/resolver boundaries, canonical public resolver literals, exact registrations, strict JSON through embedded/signed remote Client admission, atomic rejection and no contact. `FuzzDns01Admission` exercises the same boundary. `TestDns01CNAMEAnswersCauseNoContact` shows a CNAME chain causes no further query or contact. |
| `M-DNS-01-OPTIONS` | `TestMDns01Options`, `TestCodecFieldTypesAndModeRelations`: defaults, mode/transport relations, the conditional 1..1,000 ms window, wrong scalar types and unknowns. `TestDns01StrictSchemaBothOrigins` checks original input/options byte caps below/at/above. |
| `M-DNS-01-EXECUTION` | `TestMDns01ExecutionAndOutput`, `TestDns01SystemAPIOutputAndHiddenBytes`, `internal/protocol/dns/query_test.go:TestQueryCorrelationAndCollectionPrecedence`, `internal/execution/dns01_transport_test.go:TestDnsUDPSourceAndTruncation`: one operation, source/ID/question filtering and ordered replies without retry/fallback or answer contact. `TestMDns01CollectionOrderAndBounds` asserts arrival order and coarse offsets and that disabled collection reads one reply. |
| `M-DNS-01-NORMALIZATION` | `TestMDns01Normalization`, `TestQueryCorrelationAndCollectionPrecedence`, `internal/protocol/dns/wire_test.go`: codes, empty results, malformed frames, source/mismatch failures, duplicate/conflicting replies and count precedence; bounded public-answer filtering. |
| `M-DNS-01-LIMITS` | `TestMDns01Limits`, `TestDns01EquivalentRunsSerialize`, `TestCcLimitPurposeDNSRows` and wire boundary cases: independent operation/attempt deadlines, UDP 4,096 and TCP 65,535 byte messages, three replies, 16 answers/eight CNAMEs, bounded normalization, tuple exclusion and DNS frame/prefix metering. `TestMDns01CollectionOrderAndBounds` leaves a fourth response unread, `TestDns01TupleComponentsDecideSerialization` varies each tuple component, and `TestMDns01ConnectionBudget` meters one socket or connection. |
| `M-DNS-01-TERMINATION` | `TestMDns01Termination`, `internal/execution/dns01_termination_test.go:TestDns01AllTerminationBoundaries`: pre-start `not_run`, started `incomplete`, no interrupted partial measurement, first-cause precedence, no transport operation after the interruption, no direct query from system mode, and recorded outcomes preserved during close. Host, network, deadline, synthetic suspension and proxy causes at connect/write/read/normalization and UDP collection points; no native lifecycle signal claimed. |
| `M-DNS-01-OUTPUT` | `TestMDns01ExecutionAndOutput`, `TestDns01SystemAPIOutputAndHiddenBytes`, `TestDns01ResultVersionsAndClosedOutput`: one minimized mode-correct observation, ordered replies, quantized timing, observable fields only, prohibited canaries absent and production result/bundle validation. |
| `M-DNS-01-VERSIONS` | `TestMDns01Versions`, `TestDns01ResultVersionsAndClosedOutput`: exact task/result/schema versions, missing/wrong types and cross-test output rejected. |

`TestCcLimitPurposeDNSRows` executes the DNS rows of `CC-LIMIT-PURPOSE`: UDP payloads, TCP length prefixes and below/at/above attempt budgets. `TestDns01SystemAPIOutputAndHiddenBytes` covers the hidden-wire exception. `TestCcProxyOverviewDNSRow` rejects UDP/system proxy tasks without contact and runs TCP over one loopback CONNECT tunnel: DNS bytes count; the CONNECT exchange does not. Whole `CC-LIMIT-PURPOSE` remains planned for other callers and review. Tests use synthetic adapters and loopback fixtures only. Review/precondition/deployment judgement, all `RV-*` cases, device validation and full conformance remain pending. The eight families above are `passed` at Go `1a37603` ([evidence](../traceability/evidence/go-unit-1a37603.md)).


#### M-DNS-01-ADMISSION

Owners: `internal/measurement/dns01`, `internal/protocol/dns` and the root `registry.go`. Case: `dns01_conformance_test.go`, `symbol:TestMDns01Admission`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-DNS-01-ADMISSION). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

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

Owners: `internal/measurement/dns01`, `internal/protocol/dns` and the root `registry.go`. Case: `dns01_conformance_test.go`, `symbol:TestMDns01Options`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-DNS-01-OPTIONS). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Cover omitted/system/direct modes, direct-only UDP/TCP, collect_additional_responses and conditional 1..1,000 ms window: the shared registered maximum of 1,000 ms is stricter than the test's option table and controls admission, so 1,001 ms and above are rejected without clamping. Test every missing/null/wrong/unknown option and prohibited combination, with no fallback to another mode, resolver or transport.

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

Owners: `internal/measurement/dns01`, `internal/protocol/dns` and the root `registry.go`. Case: `dns01_conformance_test.go`, `symbol:TestMDns01ExecutionAndOutput`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-DNS-01-EXECUTION). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

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

Owners: `internal/measurement/dns01`, `internal/protocol/dns` and the root `registry.go`. Case: `dns01_conformance_test.go`, `symbol:TestMDns01Normalization`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-DNS-01-NORMALIZATION). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

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

Owners: `internal/measurement/dns01`, `internal/protocol/dns` and the root `registry.go`. Case: `dns01_conformance_test.go`, `symbol:TestMDns01Limits`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-DNS-01-LIMITS). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Exercise UDP 4,096-byte/TCP 65,535-byte messages, three matching responses, 16 normalized entries/eight CNAMEs and 16,384 encoded observation bytes at below/equal/above boundaries; test the additional window and outer deadline independently. Assert no test-layer retries and deduplication by the complete hostname/type/mode/resolver/generation tuple; the implementation may conservatively also serialize tuples that differ only in transport or network generation, never fewer.

<!-- requirement-ref: REQ-DNS-01-0033 -->
[REQ-DNS-01-0033](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0033) → `M-DNS-01-LIMITS` (Resource Bounds).

#### M-DNS-01-VERSIONS

Owners: `internal/measurement/dns01`, `internal/protocol/dns` and the root `registry.go`. Case: `dns01_conformance_test.go`, `symbol:TestMDns01Versions`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-DNS-01-VERSIONS). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Select test_id DNS-01 with task test_version 0.1.0 and expect a result with schema_version 0.1.0, test.id DNS-01, test.version 0.1.0 and test.result_version 0.1.0 as the only registered combination. Reject before any recording-transport traffic, as a task-validation failure rather than a Measurement Result, each task test_version variant 0.2.0, 1.0.0, 0.1, 0.1.0-draft, v0.1.0, an empty string, a non-string value and a missing field. Verify result validation rejects schema_version or test.result_version set to 0.2.0, 0.1 or missing, a DNS-01 test.version paired with a different result_version, and cross-test combinations: test.id DNS-01 carrying another test's Output (a dns_query with engine doh or dot, or NET-05 profile and step fields) and test.id DNS-03 or DNS-04 carrying a DNS-01 direct UDP output. Assert no version is coerced, guessed or added and that the Draft status and registration metadata alone confer no support. Requirements: REQ-DNS-01-0034.

<!-- requirement-ref: REQ-DNS-01-0034 -->
[REQ-DNS-01-0034](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0034) → `M-DNS-01-VERSIONS` (Output).

#### M-DNS-01-OUTPUT

Owners: `internal/measurement/dns01`, `internal/protocol/dns` and the root `registry.go`. Case: `dns01_conformance_test.go`, `symbol:TestMDns01ExecutionAndOutput`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-DNS-01-OUTPUT). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Require exactly one dns_query with mode-correct engine, observable fields only and bounded ordered responses. System API omissions must remain omissions rather than invented codes/datagrams/timing; reuse the shared DNS observation schema for answer normalization and length/count limits.

<!-- requirement-ref: REQ-DNS-01-0035 -->
[REQ-DNS-01-0035](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0035) → `M-DNS-01-OUTPUT` (Output).

<!-- requirement-ref: REQ-DNS-01-0036 -->
[REQ-DNS-01-0036](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0036) → `M-DNS-01-OUTPUT` (Output).

#### M-DNS-01-TERMINATION

Owners: `internal/measurement/dns01`, `internal/protocol/dns` and the root `registry.go`. Case: `dns01_conformance_test.go`, `symbol:TestMDns01Termination`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-DNS-01-TERMINATION). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Drive DNS-01 with a fake platform resolver, a scripted UDP/TCP recording transport, a deterministic clock and a controllable network generation, and inject each interruption at each Procedure stage after the network generation is recorded: after step 2 (generation recorded, no query issued), during the system-mode lookup (step 3), after the direct UDP or TCP query is sent and before any response (step 4), after the first matching UDP response while response-source, transaction ID and question validation runs (step 5), inside the enabled additional-response window with one and with two matching responses already read (step 5), and during response normalization (steps 6-7). Expected states follow the GO-SPEC-ISSUE-004 disposition: every listed point follows the recorded network generation, so each is incomplete as asserted below, an interruption before the generation is recorded is not_run with that reason, no measurement and no network_context, and an interruption after the outcome is recorded, while the socket closes, stays completed with the recorded outcome. Iterate the interruptions network generation change, host cancellation, outer attempt deadline expiry, proxy failure and proxy replacement or clearing where a proxy path is configured, and platform suspension. Assert execution.state is incomplete with execution.reason respectively network_changed, cancelled_by_host, attempt_deadline_exceeded, the registered proxy reason (proxy_unavailable, proxy_connect_failed or proxy_reconfigured) and platform_suspended, that reason names only the first condition, and that no measurement object is present even when matching responses were already read (DNS-01 has no partial output, unlike NET-05). Contrast a bounded resolution timeout, which must remain a completed result with outcome timeout, against attempt_deadline_exceeded. Assert the recording transport shows no further read, no second query and no fallback to another mode, server or transport after the interruption, and that no terminal DNS outcome such as network_changed, timeout or response_limit_reached is invented for an interrupted attempt. Requirements: REQ-DNS-01-0039.

<!-- requirement-ref: REQ-DNS-01-0039 -->
[REQ-DNS-01-0039](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0039) → `M-DNS-01-TERMINATION` (Outcomes).

#### M-DNS-01-MINIMIZATION

Intended owner: `internal/result/dns01.go`. Planned case: `internal/result/dns01_conformance_test.go`, `symbol:TestMDns01Minimization`. Disposition: **planned**, with no execution or support evidence.

Case plan: Run system mode and direct UDP and TCP mode against a fake platform resolver and scripted recording transports, injecting one unique synthetic canary per prohibited item: a canary queried hostname approved only by the synthetic local policy, a canary direct resolver address and a canary non-default port in the fake resolver registry, a canary system resolver configuration (resolver list and search domain) exposed by the fake platform, a canary 16-bit query identifier, a canary byte pattern in the raw DNS message, canary records in the authority and additional sections, a canary TTL on every answer, canary answer ASN and organization strings offered by a fake enrichment source, a canary local socket address, and canary platform error text on a failing system lookup. Iterate success, nxdomain, servfail, malformed_response, query_mismatch, conflicting_responses and response_limit_reached paths. Assert the serialized Measurement Result, every nested dns_query observation and every error path contain none of the canaries and no field beyond engine, address_family, query_type, responses, response_limit_reached and, per response, arrival_index, arrival_offset_ms, rcode and answers with only type and value. Assert retained answers are only the bounded normalized A, AAAA and intervening CNAME values from the scripted responses, deduplicated in first-occurrence order, and that no participant or live measurement data is used. Requirements: REQ-DNS-01-0040.

<!-- requirement-ref: REQ-DNS-01-0040 -->
[REQ-DNS-01-0040](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0040) → `M-DNS-01-MINIMIZATION` (Outcomes).

#### M-DNS-01-INTERPRETATION

Intended owner: `docs/measurement-interpretation-review.md`. Planned case: `internal/result/dns01_conformance_test.go`, `symbol:TestMDns01Interpretation`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/measurement/dns01/run.go`, `internal/result/dns01.go`.

Case plan: Assert the DNS-01 Output schema is a closed field set (outcome, observations and the registered dns_query details) with no verdict, injected, spoofed, legitimate, authoritative, correct, censored, blocked, reachable, anomaly or interference field, and that outcome values are only the registered DNS-01 outcomes. Exercise serialized results for resolved, nxdomain, no_data, conflicting_responses and response_limit_reached, the documented field descriptions and the sanitized conflicting-response example and host-facing text, and assert none states or implies that an unexpected answer was introduced intentionally, that a returned address is reachable, that answers were compared with authoritative ground truth in the run, that the first or a later response is spoofed, injected, legitimate or authoritative, that either differing resolver is correct, that censorship or blocking occurred, or that NXDOMAIN or an empty answer is a successful address resolution. Verify no answer is enriched with ASN, organization, geolocation, reputation or a censorship verdict. Requirements: REQ-DNS-01-0001, REQ-DNS-01-0002, REQ-DNS-01-0003, REQ-DNS-01-0004, REQ-DNS-01-0041, REQ-DNS-01-0042, REQ-DNS-01-0043, REQ-DNS-01-0044.

Scope: portable automated component only; judgement-based completion is review case `RV-M-DNS-01-INTERPRETATION`, which is not automated coverage.

Review case `RV-M-DNS-01-INTERPRETATION` (`docs/measurement-interpretation-review.md`, `heading:RV-M-DNS-01-INTERPRETATION`): Human review case; not automated coverage. Reviewer role: Measurement interpretation reviewer (analyst role independent of the implementer). Reviewed artifact: Go result documentation and host/analyst guidance for DNS-01 at the reviewed implementation revision. Checklist: (1) Does the guidance limit a single execution to the result exposed by the selected resolver path for the configured name, record type, time and network generation? (2) Does the guidance describe conflicting direct UDP responses only as multiple matching datagrams with differing normalized contents, without implying which one was injected or correct? (3) Does the guidance present the listed alternative explanations (authoritative or recursive outage; DNSSEC failure, stale or negative caching, split-horizon or DNS64; misconfigured or changed records; captive portal, VPN, enterprise resolver or local filtering; platform error collapsing; direct resolver outage, filtering, rate limiting or routing failure; legitimate retransmission or anycast change) as considerations before any stronger claim? (4) Does the guidance require the listed stronger evidence (repeated system and direct results on a stable network, separately reviewed authoritative comparison, DNS-03 and DNS-04 observations where permitted, TCP/TLS/HTTPS measurements to approved endpoints, independent comparable observations) before stronger interpretation? (5) Does the guidance treat a difference between resolvers as differing observations rather than proof that either resolver is correct or that censorship occurred? (6) Does the documentation state the limitation that internal system resolver retries and behavior may be opaque? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-DNS-01-0001 -->
[REQ-DNS-01-0001](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0001) → `M-DNS-01-INTERPRETATION`, `RV-M-DNS-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-DNS-01-0002 -->
[REQ-DNS-01-0002](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0002) → `M-DNS-01-INTERPRETATION`, `RV-M-DNS-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-DNS-01-0003 -->
[REQ-DNS-01-0003](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0003) → `M-DNS-01-INTERPRETATION`, `RV-M-DNS-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-DNS-01-0004 -->
[REQ-DNS-01-0004](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0004) → `M-DNS-01-INTERPRETATION`, `RV-M-DNS-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-DNS-01-0041 -->
[REQ-DNS-01-0041](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0041) → `M-DNS-01-INTERPRETATION`, `RV-M-DNS-01-INTERPRETATION` (Supported Interpretation).

<!-- requirement-ref: REQ-DNS-01-0042 -->
[REQ-DNS-01-0042](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0042) → `M-DNS-01-INTERPRETATION`, `RV-M-DNS-01-INTERPRETATION` (Supported Interpretation).

<!-- requirement-ref: REQ-DNS-01-0043 -->
[REQ-DNS-01-0043](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0043) → `M-DNS-01-INTERPRETATION`, `RV-M-DNS-01-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

<!-- requirement-ref: REQ-DNS-01-0044 -->
[REQ-DNS-01-0044](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0044) → `M-DNS-01-INTERPRETATION`, `RV-M-DNS-01-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

#### M-DNS-01-SAFETY

Intended owner: `docs/measurement-deployment-review.md`. Planned case: `internal/admission/measurement_safety_conformance_test.go`, `symbol:TestMDns01Safety`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/execution/measurements.go`, `internal/result/dns01.go`.

Case plan: Iterate the Baseline Risk rows (risk level low; requires user consent false, subject to host and local policy) and each Risk Factors row (network_observability, traffic_distinctiveness, target_sensitivity, data_sensitivity, legal_or_social_context, each low) using synthetic tasks: assert a task declaring risk_level medium, high or unknown is evaluated at that level, the effective risk is never below the registered low baseline, and a task or fake host policy requiring consent prevents execution without fake host consent. Exercise each Safety Analysis action with synthetic targets and recording transports: data minimization, asserting retained data is limited to requested type, coarse outcome, timing and normalized answers; traffic distinctiveness, asserting exactly one ordinary query of type A or AAAA per run with no test-layer retry or repetition; user-identification risk, asserting no installation identifier and no canary entries from a fake unrelated DNS cache; target risk, asserting a non-reviewed hostname or unregistered direct resolver is rejected with zero recorded traffic; repeated-execution risk, asserting the test layer does not self-schedule repeats; and probe authority, asserting a fake local policy can delay, skip or reject (not_run with local_policy_rejected) with no traffic. Requirements: REQ-DNS-01-0045, REQ-DNS-01-0046, REQ-DNS-01-0047, REQ-DNS-01-0048, REQ-DNS-01-0049, REQ-DNS-01-0050, REQ-DNS-01-0051, REQ-DNS-01-0052.

Scope: portable automated component only; judgement-based completion is review case `RV-M-DNS-01-SAFETY`, which is not automated coverage.

Review case `RV-M-DNS-01-SAFETY` (`docs/measurement-deployment-review.md`, `heading:RV-M-DNS-01-SAFETY`): Human review case; not automated coverage. Reviewer role: Deployment safety and privacy reviewer (independent of the implementer). Reviewed artifact: DNS-01 deployment configuration, approved hostname and direct resolver registrations, host scheduling policy and native Android/Apple resolver adapters at the reviewed revision. Checklist: (1) Are all approved hostnames non-sensitive with known operational ownership in each intended deployment context, given that query names can be sensitive in some environments? (2) Is each registered direct resolver safety-reviewed for the deployment context, with its recognizability on the network path considered? (3) Does host scheduling of repeated DNS-01 runs for a fixed name or direct resolver remain conservative enough to avoid a recognizable pattern? (4) Do the native system-resolver paths avoid exposing or retaining unrelated system DNS activity or cache contents on each platform? (5) Are host and local consent policies documented where they require consent despite the false baseline value? (6) Are the residual risks (resolver and network observe the hostname; direct resolver use may be treated differently; answers may indirectly reveal network configuration) disclosed in host-facing documentation? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-DNS-01-0045 -->
[REQ-DNS-01-0045](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0045) → `M-DNS-01-SAFETY`, `RV-M-DNS-01-SAFETY` (Baseline Risk).

<!-- requirement-ref: REQ-DNS-01-0046 -->
[REQ-DNS-01-0046](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0046) → `M-DNS-01-SAFETY`, `RV-M-DNS-01-SAFETY` (Risk Factors).

<!-- requirement-ref: REQ-DNS-01-0047 -->
[REQ-DNS-01-0047](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0047) → `M-DNS-01-SAFETY`, `RV-M-DNS-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-DNS-01-0048 -->
[REQ-DNS-01-0048](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0048) → `M-DNS-01-SAFETY`, `RV-M-DNS-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-DNS-01-0049 -->
[REQ-DNS-01-0049](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0049) → `M-DNS-01-SAFETY`, `RV-M-DNS-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-DNS-01-0050 -->
[REQ-DNS-01-0050](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0050) → `M-DNS-01-SAFETY`, `RV-M-DNS-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-DNS-01-0051 -->
[REQ-DNS-01-0051](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0051) → `M-DNS-01-SAFETY`, `RV-M-DNS-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-DNS-01-0052 -->
[REQ-DNS-01-0052](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-01.md#REQ-DNS-01-0052) → `M-DNS-01-SAFETY`, `RV-M-DNS-01-SAFETY` (Safety Analysis).

### DNS-03

#### Current DNS-03 Execution

`internal/measurement/dns03` owns the strict codec and the runner; the root `registry.go` assembles its exact `0.1.0` entry with medium baseline risk (REQ-DNS-03-0033), a target rule and a proxy rule. DNS-03 is not a NET-05 step. The codec accepts an absolute `https` resolver URL of at most 2,048 bytes whose host is a DNS name, never an address, since the resolver is bootstrapped and authenticated by that name; it rejects credentials, fragments, interpolation and local or special-use names, folds the host case, omits default port 443 and supplies an empty path as `/`, preserving path and query bytes. Whether a query parameter is user-specific is decided by exact registration. The target rule extracts the query name as a measurement target and the complete resolver profile as a resolver target: the normalized URL, its hostname as hostname, resolver name and TLS server name, the port, scheme `https` and, when supplied, the one bootstrap address. Admission matches both exactly. The proxy rule accepts only a task with `bootstrap_ip` (REQ-PROXY-0044); `CONNECT` uses that address while TLS authenticates the URL hostname.

The runner follows the Procedure. Without `bootstrap_ip`, one system lookup of the URL hostname through the staged destination gate records a system `dns_query`; `auto` makes one IPv4 lookup, as HTTP-01 and DNS-04 do, and `ipv6` one IPv6 lookup. With `bootstrap_ip` the family is the address's and the bootstrap observation is omitted. One TCP connection follows, then one TLS handshake that authenticates the URL hostname and offers ALPN `http/1.1`, then one RFC 8484 POST through the shared HTTP/1.1 exchange with redirects disabled, `Content-Type` and `Accept` `application/dns-message`, identity encoding and no cookie, credential or user agent. The query uses DNS ID 0, as RFC 8484 recommends; TLS protects the exchange. A 2xx `application/dns-message` payload of at most 65,535 bytes is parsed through `dns.Message` into a `doh` `dns_query` with at most 16 normalized answers, whose duration and arrival offset are the carrying exchange's. The attempt connection switches to the session at the handshake, so handshake and record bytes are not counted, while HTTP framing and the DNS messages are counted in both directions on the plaintext side (REQ-LIMIT-0034). The outcome is recorded before close (GO-SPEC-ISSUE-004). There is no GET, redirect, alternate endpoint or address, retry or connection reuse, and one run is active per resolver and query tuple.

A non-2xx response is `http_failure` before its encoding or media type is checked, without reading its body or producing a DNS observation; a redirect `Location` is neither classified nor contacted ([GO-SPEC-ISSUE-020](specification-issues.md#go-spec-issue-020-dns-03-non-success-status-and-media-type-precedence)). A 2xx response with another media type is `unexpected_content_type`; any other HTTP failure, such as an unsupported encoding or a protocol error, is `http_failure`. At every stage a registered `timeout` is `timeout` and an `inconclusive` failure `inconclusive`; every other bootstrap, TCP or TLS failure keeps its stage class, including a TLS reset as `tls_failure`. An oversized or unparseable payload is `malformed_dns_response`.

The suites are portable synthetic evidence with a scripted system resolver, a recording transport, a loopback HTTPS server with a scripted DoH responder, a loopback TLS server for failure modes and a synthetic CA, a deterministic clock and, for the proxy path, the loopback `CONNECT` proxy. The HTTP/1.1-only exchange is a recorded limitation: RFC 8484 recommends HTTP/2, which browsers use, so the traffic shape differs from the most common DoH clients. Native trust-store, resolver and platform behavior (DEV-455), the PRECONDITIONS, MINIMIZATION, INTERPRETATION and SAFETY families and their `RV-*` cases remain deferred.

| Family | Executable suite |
|---|---|
| `M-DNS-03-ADMISSION` | `TestMDns03Admission` |
| `M-DNS-03-OPTIONS` | `TestMDns03Options` |
| `M-DNS-03-EXECUTION` | `TestMDns03ExecutionAndOutput`, `TestDns03PlaintextAccounting` |
| `M-DNS-03-NORMALIZATION` | `TestMDns03Normalization`, `TestPOSTStatusPrecedesMediaType` |
| `M-DNS-03-LIMITS` | `TestMDns03Limits`, `TestMDns03TupleConcurrency`, `TestDns03PlaintextAccounting` |
| `M-DNS-03-TERMINATION` | `TestMDns03Termination`, `TestDns03TerminationPoints` |
| `M-DNS-03-OUTPUT` | `TestMDns03ExecutionAndOutput`, `TestMDns03Normalization` |
| `M-DNS-03-VERSIONS` | `TestMDns03Versions` |

The manifest records these families at the next milestone.

#### M-DNS-03-ADMISSION

Disposition: **passed** for synthetic production admission in `dns03_conformance_test.go`, `symbol:TestMDns03Admission`: registered profiles with and without a bootstrap address, an IPv6 bootstrap address, a non-default port, a query string, an empty path, AAAA, case-folded and IDNA names, 253-byte resolver and query names and a 2,048-byte URL are admitted; missing, mistyped and unknown fields, non-`https`, credential-bearing, fragment, template-interpolated, IP-literal, local, special-use, malformed-port and relative URLs, IP-literal, single-label, local, interpolated and over-long query names, and non-canonical, mapped, private, loopback, link-local, documentation and zoned bootstrap addresses are `task_schema_invalid`; a 2,049-byte URL is `task_limit_exceeded`; a different path, query, port, bootstrap address, resolver or query name, or a bootstrap address added or omitted relative to the profile, is `target_not_registered`; none produces a handle, replay commit, lookup or contact. TLS authentication of the URL hostname with a bootstrap address runs in `TestMDns03ExecutionAndOutput`, and redirect non-contact in `TestMDns03Normalization`.

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

Disposition: **passed** for synthetic production admission in `dns03_conformance_test.go`, `symbol:TestMDns03Options`: omitted, `auto`, `ipv4` and `ipv6` families compatible with the bootstrap address, or with system bootstrap, are admitted; an incompatible family, an unknown, null, empty or mistyped family, and GET, redirect following, alternate endpoints, custom headers, credentials, cookies, task proxy configuration, connection reuse, retry and unknown options are `task_schema_invalid` without contact.

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

Disposition: **passed** for synthetic production execution in `dns03_conformance_test.go`, `symbol:TestMDns03ExecutionAndOutput`, and `internal/execution/dns03_termination_test.go`, `symbol:TestDns03PlaintextAccounting`, through `Client` with the production registry on the direct path with and without system bootstrap and on the proxy path with a bootstrap address: at most one lookup, one connection, one ClientHello naming the URL hostname with ALPN `http/1.1`, one POST to the registered path with the DNS media types, identity encoding, a 31-byte query with ID 0 and no other header, and one response, then close; a proxy task without a bootstrap address is rejected at admission without contact.

Case plan: Record optional system bootstrap, one selected resolver TCP connection, authenticated TLS, one RFC 8484 POST with application/dns-message and bounded response validation. Vary returned media type, DNS ID/name/type mismatch, malformed/oversized message and each failed stage; prohibit redirects, endpoint/transport fallback and application retries.

<!-- requirement-ref: REQ-DNS-03-0005 -->
[REQ-DNS-03-0005](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0005) → `M-DNS-03-EXECUTION` (Operation).

<!-- requirement-ref: REQ-DNS-03-0006 -->
[REQ-DNS-03-0006](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0006) → `M-DNS-03-EXECUTION` (Procedure).

#### M-DNS-03-NORMALIZATION

Disposition: **passed** for synthetic production cases in `dns03_conformance_test.go`, `symbol:TestMDns03Normalization`, and `internal/protocol/http/exchange_test.go`, `symbol:TestPOSTStatusPrecedesMediaType`: `response` with answers, empty and with media-type parameters; `bootstrap_failure`; `tcp_failure` for refused, reset and unreachable; `tls_failure` for a reset, an untrusted certificate, a URL-hostname mismatch, an alert and a protocol error; `http_failure` for `503` HTML, `404` with a DNS message and a `307` redirect (status first, no parse and no redirect contact, GO-SPEC-ISSUE-020), an unsupported encoding, a premature body end and a malformed response; `unexpected_content_type` for another and a missing media type; `dns_nxdomain`, `dns_servfail`, `dns_refused`; `malformed_dns_response` for an empty and an invalid message; `query_mismatch` for identifier and type; `timeout` for bootstrap and TCP; and `inconclusive` for an unclassified TCP failure and an unregistered response code, each at its observed stage after ordered successful prior stages and with exactly one request. The post-connect `timeout` rows run in `TestMDns03Limits`.

Case plan: Exercise all terminal outcome rows, including non-success HTTP, content type, DNS response-code failures, malformed/mismatching DNS, per-stage timeout and inconclusive errors. Attribute failure only to the actually observed stage and preserve ordered preceding observations.

<!-- requirement-ref: REQ-DNS-03-0007 -->
[REQ-DNS-03-0007](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0007) → `M-DNS-03-NORMALIZATION` (Success Condition).

<!-- requirement-ref: REQ-DNS-03-0008 -->
[REQ-DNS-03-0008](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0008) → `M-DNS-03-NORMALIZATION` (Failure Conditions).

<!-- requirement-ref: REQ-DNS-03-0027 -->
[REQ-DNS-03-0027](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0027) → `M-DNS-03-NORMALIZATION` (Outcomes).

#### M-DNS-03-LIMITS

Disposition: **passed** for synthetic cases in `dns03_conformance_test.go`, `symbol:TestMDns03Limits` and `symbol:TestMDns03TupleConcurrency`, and `internal/execution/dns03_termination_test.go`, `symbol:TestDns03PlaintextAccounting`: the TLS and HTTP operation bounds yield `timeout` while the outer deadline is open and the outer deadline yields `incomplete`; messages of 65,534 and 65,535 bytes are accepted, 65,536 bytes and a 200,000-byte payload are `malformed_dns_response`, and a stricter received limit below the framed response ends the attempt with `resource_limit`; 16 answers are retained and 17 are `malformed_dns_response`; a sent limit equal to the measured request succeeds and one byte below it ends with `resource_limit` before the request is complete; metered usage equals the server-side plaintext on the direct and proxy paths; one connection and request without retry; and an equivalent resolver and query tuple waits without contact.

Case plan: Test the 65,535-byte DNS message and 16 retained answers at below/equal/above boundaries under stricter shared DNS/HTTP/TLS/task limits. Assert one post-bootstrap resolver connection/request, no retry, full-chain deadline and deduplication by resolver/query tuple.

<!-- requirement-ref: REQ-DNS-03-0023 -->
[REQ-DNS-03-0023](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0023) → `M-DNS-03-LIMITS` (Resource Bounds).

#### M-DNS-03-VERSIONS

Disposition: **passed** for synthetic cases in `dns03_conformance_test.go`, `symbol:TestMDns03Versions`: task-time rejection of unsupported, malformed and missing `test_version` with no lookup, contact or run; the emitted combination decodes; the result decoder rejects every unsupported or malformed version, DNS-04 and DNS-01 identities and bindings, a bootstrap binding without the system observation, foreign outcomes, engines and query types, a missing HTTP stage, reordered stages and unknown fields.

Case plan: Select test.id DNS-03 with test.version 0.1.0, Measurement Result schema_version 0.1.0 and test.result_version 0.1.0 and assert that exact registered combination validates and is the only combination emitted. Iterate rejections before any transport contact on a recording transport: unsupported test.version values (0.1.1, 0.2.0, 1.0.0), malformed versions (empty, 0.1, v0.1.0, 0.1.0-draft, non-string, whitespace-padded), unsupported schema_version and unsupported test.result_version values in the same forms, a mismatched pair where only one of test.version and result_version is 0.1.0, and cross-test combinations such as test.id DNS-03 paired with a DNS-01 or DNS-04 Output object, or test.id DNS-01 carrying a DNS-03 Output object. Assert each rejected task is a task-validation failure reported outside Measurement Results, produces no Measurement Result and no measurement object, and is never coerced, guessed, interpreted best-effort or mapped to another version. Verify a received result with an unsupported version is not parsed as the known 0.1.0 format and that unknown fields within 0.1.0 are rejected. Assert the Draft status and registration metadata alone do not cause any additional version to be accepted. Requirements: REQ-DNS-03-0024.

<!-- requirement-ref: REQ-DNS-03-0024 -->
[REQ-DNS-03-0024](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0024) → `M-DNS-03-VERSIONS` (Output).

#### M-DNS-03-OUTPUT

Disposition: **passed** for synthetic production output in `dns03_conformance_test.go`, `symbol:TestMDns03ExecutionAndOutput` and `symbol:TestMDns03Normalization`: optional system bootstrap, TCP, TLS, HTTP and `doh` DNS observations in order, bootstrap omitted with a bootstrap address, unstarted stages omitted after failure and no DNS observation after a non-2xx or wrong-media-type response, only registered details, and at most 16 normalized answers without TTL. Every result is checked for resolver URL and path, hostname, query name, bootstrap address, certificate, cookie, header, body, media-type, raw-error and loopback canaries.

Case plan: Verify bootstrap DNS only when actually used, then TCP/TLS/HTTP/doh DNS in execution order, stopping at the failed stage. Use shared DNS normalization and HTTP/TLS observations without importing HTTP-01 GET semantics or synthesizing unstarted stages.

<!-- requirement-ref: REQ-DNS-03-0025 -->
[REQ-DNS-03-0025](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0025) → `M-DNS-03-OUTPUT` (Output).

<!-- requirement-ref: REQ-DNS-03-0026 -->
[REQ-DNS-03-0026](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0026) → `M-DNS-03-OUTPUT` (Output).

#### M-DNS-03-TERMINATION

Disposition: **passed** for synthetic cases in `internal/execution/dns03_termination_test.go`, `symbol:TestDns03TerminationPoints`, and `dns03_conformance_test.go`, `symbol:TestMDns03Termination`: before the test starts is `not_run`; during bootstrap, TCP connect, TLS establishment and after the POST is read it is `incomplete` with the first of host cancellation, `network_changed`, the attempt deadline, injected suspension and the proxy reasons, with no measurement, further lookup, dial or request and every connection closed; while closing after the outcome is recorded it stays `completed`. Through `Client`, a generation change and a cancellation after the POST and a refused tunnel are `incomplete` without direct or resolver fallback.

Case plan: Drive DNS-03 with a synthetic locally registered resolver profile, a scripted system resolver, scripted TCP/TLS/HTTP transports, a deterministic clock and a fake platform, in two variants: without bootstrap_ip (system bootstrap dns_query required) and with an approved synthetic bootstrap_ip (first observation omitted). Interrupt at each Procedure point: after input validation and before bootstrap; during the system bootstrap dns_query; between bootstrap and TCP; during tcp_connect; between TCP and TLS; during tls_handshake (including resolver certificate authentication for the URL hostname); during the single HTTPS POST with media type application/dns-message (request send and response read); during HTTP response validation and bounded DNS message parsing; and after the final doh dns_query observation but before result emission. Expected states follow the GO-SPEC-ISSUE-004 disposition: after input validation and before bootstrap the attempt is not_run with that reason, no measurement and no network_context; after the final doh dns_query observation it stays completed with the recorded outcome; every other point is incomplete as asserted below. At every point inject each REQ-DNS-03-0028 condition separately: network generation change (execution.reason network_changed), host cancellation (cancelled_by_host), outer attempt deadline expiry (attempt_deadline_exceeded), proxy failure or reconfiguration where a host proxy path is in effect (proxy_unavailable, proxy_connect_failed, proxy_reconfigured), and platform suspension (platform_suspended). At the incomplete points, assert execution.state is incomplete with exactly that reason, that no measurement object is present (DNS-03 does not permit NET-05-style partial output), and that when two conditions occur the reason names only the first. Verify contrastively that an active-stage protocol timeout is a completed result with outcome timeout and the terminal DNS, TCP, TLS or HTTP observation failed with its registered timeout, distinct from attempt_deadline_exceeded. Assert via the recording transports that after interruption sockets and TLS/HTTP sessions are released, no further connection or request is made, no fallback to another address, endpoint, transport, GET or redirect target occurs, no application-layer retry occurs, and no unstarted stage or terminal DNS outcome is synthesized. Requirements: REQ-DNS-03-0028.

<!-- requirement-ref: REQ-DNS-03-0028 -->
[REQ-DNS-03-0028](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0028) → `M-DNS-03-TERMINATION` (Outcomes).

#### M-DNS-03-MINIMIZATION

Intended owner: `internal/result/dns03.go`. Planned case: `internal/result/dns03_conformance_test.go`, `symbol:TestMDns03Minimization`. Disposition: **planned**, with no execution or support evidence.

Case plan: Configure a synthetic locally registered resolver profile and scripted system resolver, TCP, TLS and HTTP transports seeded with one unique canary per REQ-DNS-03-0029 item: resolver_url (including path canary), resolver hostname, resolver address (returned by the scripted bootstrap resolver and used for TCP), bootstrap_ip, query hostname, raw DNS message bytes (query and response, including a canary query identifier and canary authority and additional section records), HTTP header names and values (including a canary Content-Type value on an unexpected_content_type response), HTTP body bytes beyond the DNS message, certificate data (canary subject, issuer, serial and fingerprint in the synthetic chain), resolver cookies (Set-Cookie canary), and raw transport, TLS, HTTP, DNS and platform error strings. Iterate the success path and each failure class (bootstrap_failure, tcp_failure, tls_failure, http_failure, unexpected_content_type, dns_nxdomain, dns_servfail, dns_refused, malformed_dns_response, query_mismatch, timeout, inconclusive) plus incomplete terminations, with and without bootstrap_ip. Assert serialized result and error paths, including every nested observation, contain none of the canaries in any encoding and no field outside the allowed set: observation type, status, duration_ms, failure, and details limited to dns_query engine, address_family, query_type, responses (arrival_index, arrival_offset_ms, rcode, answers with type and value), response_limit_reached; tcp_connect address_family; tls_handshake address_family, tls_version, alpn, certificate_validation, alert; http_exchange address_family, protocol, status_code, body (bytes_processed, truncated), content_encoding. Verify retained answers in the doh dns_query are only bounded normalized A, AAAA or CNAME values, at most 16 per response, deduplicated, without TTLs. No participant or live measurement data is used. Requirements: REQ-DNS-03-0029.

<!-- requirement-ref: REQ-DNS-03-0029 -->
[REQ-DNS-03-0029](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0029) → `M-DNS-03-MINIMIZATION` (Outcomes).

#### M-DNS-03-INTERPRETATION

Intended owner: `docs/measurement-interpretation-review.md`. Planned case: `internal/result/dns03_conformance_test.go`, `symbol:TestMDns03Interpretation`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/measurement/dns03/run.go`, `internal/result/dns03.go`.

Case plan: Generate synthetic results for every DNS-03 outcome (response, bootstrap_failure, tcp_failure, tls_failure, http_failure, unexpected_content_type, dns_nxdomain, dns_servfail, dns_refused, malformed_dns_response, query_mismatch, timeout, inconclusive) plus incomplete terminations, including a response whose answers differ from a synthetic system DNS answer set. Assert the serialized result schema has no interpretation, verdict, blocking, censorship, interference, authoritative, correctness or anomaly field, and that unknown fields of that kind are rejected. Assert generated Go field descriptions, error and log strings, and the sanitized DNS-03 example and host-facing text contain none of these prohibited claims: that the test configured DoH for the device or host application; that a DoH result is authoritative or correct; that a difference from system DNS establishes interference; that a bootstrap, TCP, TLS, HTTP or DNS failure is DoH blocking or censorship; and that answers carry on-device ASN, organization, geolocation, reputation or censorship-verdict enrichment. Verify each failure is attributed only to the directly observed stage via observation order and the terminal observation. Requirements: REQ-DNS-03-0001, REQ-DNS-03-0002, REQ-DNS-03-0003, REQ-DNS-03-0030, REQ-DNS-03-0031, REQ-DNS-03-0032.

Scope: portable automated component only; judgement-based completion is review case `RV-M-DNS-03-INTERPRETATION`, which is not automated coverage.

Review case `RV-M-DNS-03-INTERPRETATION` (`docs/measurement-interpretation-review.md`, `heading:RV-M-DNS-03-INTERPRETATION`): Human review case; not automated coverage. Reviewer role: Measurement interpretation reviewer (analyst role independent of the implementer). Reviewed artifact: Go result documentation and host/analyst guidance for DNS-03 at the reviewed implementation revision. Checklist: (1) Does the guidance limit one execution to establishing whether each observed bootstrap, TCP, TLS, HTTP and DNS stage completed for the configured resolver and query? (2) Does the guidance attribute each failure to the directly observed stage rather than summarizing it as DoH blocking? (3) Are the Alternative Explanations retained: resolver, authoritative DNS, CDN or certificate outage; bootstrap DNS failure unrelated to the DoH service; resolver rate limiting, policy refusal, maintenance or overload; VPN, proxy, captive portal, enterprise policy or local firewall behavior; IPv4/IPv6 routing differences or transient packet loss; stale or incorrect registered resolver configuration? (4) Does any stronger interpretation require the listed evidence: repeated DoH results on a stable network without aggressive retries, DNS-01 and DNS-04 results for the same safe query, layer-specific TCP, TLS and HTTPS checks against approved endpoints, and independent observations from comparable times and network contexts? (5) Does the guidance state that a difference from system DNS does not by itself establish interference and that a DoH result is not shown to be authoritative or correct? (6) Does the guidance avoid implying that the test configures DoH for the device or host application? (7) Is any derived anomaly or interference analysis kept separate from the source result with its method and version identified? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-DNS-03-0001 -->
[REQ-DNS-03-0001](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0001) → `M-DNS-03-INTERPRETATION`, `RV-M-DNS-03-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-DNS-03-0002 -->
[REQ-DNS-03-0002](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0002) → `M-DNS-03-INTERPRETATION`, `RV-M-DNS-03-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-DNS-03-0003 -->
[REQ-DNS-03-0003](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0003) → `M-DNS-03-INTERPRETATION`, `RV-M-DNS-03-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-DNS-03-0030 -->
[REQ-DNS-03-0030](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0030) → `M-DNS-03-INTERPRETATION`, `RV-M-DNS-03-INTERPRETATION` (Supported Interpretation).

<!-- requirement-ref: REQ-DNS-03-0031 -->
[REQ-DNS-03-0031](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0031) → `M-DNS-03-INTERPRETATION`, `RV-M-DNS-03-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

<!-- requirement-ref: REQ-DNS-03-0032 -->
[REQ-DNS-03-0032](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0032) → `M-DNS-03-INTERPRETATION`, `RV-M-DNS-03-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

#### M-DNS-03-SAFETY

Intended owner: `docs/measurement-deployment-review.md`. Planned case: `internal/admission/measurement_safety_conformance_test.go`, `symbol:TestMDns03Safety`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/execution/measurements.go`, `internal/result/dns03.go`.

Case plan: Using synthetic locally registered resolver endpoints and query names, a fake local policy and host, a deterministic clock and recording TCP/TLS/HTTP transports, iterate the Baseline Risk rows: assert the registered baseline risk level is medium and requires_user_consent is false, and verify a task declaring low is rejected or evaluated as at least medium, a task declaring high yields effective risk high, a task requiring consent without valid host consent is not executed, and host or local policy requiring consent is honored despite the false baseline. Iterate each Risk Factors row (network_observability medium, traffic_distinctiveness medium, target_sensitivity medium, data_sensitivity low, legal_or_social_context medium): assert the task overall risk is at least as restrictive as every factor, an unknown factor yields unknown unless resolved by trusted local assessment, and no factor lowers the baseline. Exercise each normative action: data minimization, asserting no response body, full headers, resolver cookies or address lists are retained; traffic distinctiveness, asserting the recording transport sees exactly one ordinary RFC 8484 POST with application/dns-message and no unusual headers, padding variation or synchronized repetition under the deterministic clock; user-identification risk, asserting the recorded request carries no cookies, authentication, installation identifiers or application headers; target risk, asserting an unregistered or unreviewed resolver endpoint or query name is rejected with zero recorded transport contacts; and probe authority, asserting local policy can delay, skip or reject execution (execution.reason local_policy_rejected per the public execution table) with no network contact. Requirements: REQ-DNS-03-0033, REQ-DNS-03-0034, REQ-DNS-03-0035, REQ-DNS-03-0036, REQ-DNS-03-0037, REQ-DNS-03-0038, REQ-DNS-03-0039.

Scope: portable automated component only; judgement-based completion is review case `RV-M-DNS-03-SAFETY`, which is not automated coverage.

Review case `RV-M-DNS-03-SAFETY` (`docs/measurement-deployment-review.md`, `heading:RV-M-DNS-03-SAFETY`): Human review case; not automated coverage. Reviewer role: Deployment safety and privacy reviewer (independent of the implementer, with host-integration knowledge). Reviewed artifact: DNS-03 deployment configuration, registered resolver profiles and query names, host scheduling and consent integration, and native Android/Apple network adapter behavior at the reviewed implementation revision. Checklist: (1) Have the resolver endpoints and query names completed technical, privacy and safety review and been locally registered for the target deployment? (2) Is the medium baseline justified for the deployment, given that the resolver destination and TLS metadata remain observable and encrypted DNS use may be sensitive there, or should effective risk be raised? (3) Do host scheduling and orchestration avoid repeated connections to a recognizable resolver forming a strong participation signal and avoid synchronized repetition across devices? (4) Does host and local policy determine consent appropriately where requires_user_consent is false but the deployment may require consent? (5) Do the native network stacks add no cookies, authentication, installation identifiers or application headers to the DoH request and leave its traffic pattern ordinary? (6) Are residual risks documented for hosts: the resolver operator sees the query and source network, and the access network sees the resolver destination? (7) Is DNS-03 run only under explicit reviewed host configuration or as an approved anomaly-triggered diagnostic? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-DNS-03-0033 -->
[REQ-DNS-03-0033](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0033) → `M-DNS-03-SAFETY`, `RV-M-DNS-03-SAFETY` (Baseline Risk).

<!-- requirement-ref: REQ-DNS-03-0034 -->
[REQ-DNS-03-0034](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0034) → `M-DNS-03-SAFETY`, `RV-M-DNS-03-SAFETY` (Risk Factors).

<!-- requirement-ref: REQ-DNS-03-0035 -->
[REQ-DNS-03-0035](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0035) → `M-DNS-03-SAFETY`, `RV-M-DNS-03-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-DNS-03-0036 -->
[REQ-DNS-03-0036](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0036) → `M-DNS-03-SAFETY`, `RV-M-DNS-03-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-DNS-03-0037 -->
[REQ-DNS-03-0037](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0037) → `M-DNS-03-SAFETY`, `RV-M-DNS-03-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-DNS-03-0038 -->
[REQ-DNS-03-0038](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0038) → `M-DNS-03-SAFETY`, `RV-M-DNS-03-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-DNS-03-0039 -->
[REQ-DNS-03-0039](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-03.md#REQ-DNS-03-0039) → `M-DNS-03-SAFETY`, `RV-M-DNS-03-SAFETY` (Safety Analysis).

### DNS-04

#### Current DNS-04 Execution

`internal/measurement/dns04` owns the strict codec and the runner; the root `registry.go` assembles its exact `0.1.0` entry with medium baseline risk (REQ-DNS-04-0033), a target rule and a proxy rule. DNS-04 is not a NET-05 step. The target rule extracts the query name as a measurement target and the complete resolver profile as a resolver target: `resolver_name` as hostname, resolver name and TLS server name, the port, scheme `tls` and, when supplied, the one bootstrap address. Admission matches both exactly, so a profile with several bootstrap addresses needs one registration per address, and a non-standard port is admitted only through an exactly matching reviewed profile. The proxy rule accepts only a task with `bootstrap_ip` (REQ-PROXY-0044); `CONNECT` uses that address while TLS authenticates the resolver name.

The runner follows the Procedure. Without `bootstrap_ip`, one system lookup of the resolver name through the staged destination gate records a system `dns_query`; `auto` makes one IPv4 lookup, as HTTP-01 does for a direct hostname, and `ipv6` one IPv6 lookup. With `bootstrap_ip` the family is the address's and the bootstrap observation is omitted. One TCP connection follows, then one TLS handshake that authenticates `resolver_name` with no ALPN offer, since the test registers none, then one length-prefixed exchange through `dns.Session`, recorded as a `dot` `dns_query` with at most 16 normalized answers. The attempt connection switches to the session at the handshake, so handshake and record bytes are not counted, while the DNS message and its two-byte prefix are counted in both directions on the plaintext side (REQ-LIMIT-0034). The outcome is recorded before close (GO-SPEC-ISSUE-004). There is no cleartext fallback, second resolver or retry, and one run is active per resolver and query tuple.

A TLS reset is `tcp_reset`, a TLS timeout `timeout`, and every other TLS failure, including resolver certificate failure, `tls_failure`. A TCP failure or DNS-stage failure without a DNS-04 class, such as `network_unreachable` or a session closed mid-exchange (`resolver_error`), keeps its stage and duration with failure `inconclusive`, as does an unregistered response code.

The suites are portable synthetic evidence with a scripted system resolver, a recording transport, a loopback TLS server with a scripted DoT responder and a synthetic CA, a deterministic clock and, for the proxy path, the loopback `CONNECT` proxy. Native trust-store, resolver and platform behavior (DEV-455), the PRECONDITIONS, MINIMIZATION, INTERPRETATION and SAFETY families and their `RV-*` cases remain deferred.

| Family | Executable suite |
|---|---|
| `M-DNS-04-ADMISSION` | `TestMDns04Admission` |
| `M-DNS-04-OPTIONS` | `TestMDns04Options` |
| `M-DNS-04-EXECUTION` | `TestMDns04ExecutionAndOutput`, `TestDns04PlaintextAccounting` |
| `M-DNS-04-NORMALIZATION` | `TestMDns04Normalization` |
| `M-DNS-04-LIMITS` | `TestMDns04Limits`, `TestMDns04TupleConcurrency`, `TestDns04PlaintextAccounting` |
| `M-DNS-04-TERMINATION` | `TestMDns04Termination`, `TestDns04TerminationPoints` |
| `M-DNS-04-OUTPUT` | `TestMDns04ExecutionAndOutput`, `TestMDns04Normalization` |
| `M-DNS-04-VERSIONS` | `TestMDns04Versions` |

The manifest records these families at the next milestone.

#### M-DNS-04-ADMISSION

Disposition: **passed** for synthetic production admission in `dns04_conformance_test.go`, `symbol:TestMDns04Admission`: registered profiles with and without a bootstrap address, an IPv6 bootstrap address, a non-standard port in a matching reviewed profile, AAAA, case-folded and IDNA names and 253-byte resolver and query names are admitted; missing, mistyped, out-of-range, duplicate and unknown fields, IP-literal, single-label, local, special-use, interpolated and over-long names, and non-canonical, mapped, private, loopback, link-local, documentation and zoned bootstrap addresses are `task_schema_invalid`; a different port, bootstrap address, resolver or query name, or a bootstrap address added or omitted relative to the profile, is `target_not_registered`; none produces a handle, replay commit, lookup or contact. TLS authentication of `resolver_name` with a bootstrap address runs in `TestMDns04ExecutionAndOutput`.

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

Disposition: **passed** for synthetic production admission in `dns04_conformance_test.go`, `symbol:TestMDns04Options`: omitted, `auto`, `ipv4` and `ipv6` families compatible with the bootstrap address, or with system bootstrap, are admitted; an incompatible family, an unknown, null, empty or mistyped family, and opportunistic authentication, cleartext or resolver fallback, connection reuse, ALPN, verification, retry and unknown options are `task_schema_invalid` without contact.

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

Disposition: **passed** for synthetic production execution in `dns04_conformance_test.go`, `symbol:TestMDns04ExecutionAndOutput`, and `internal/execution/dns04_termination_test.go`, `symbol:TestDns04PlaintextAccounting`, through `Client` with the production registry on the direct path with and without system bootstrap and on the proxy path with a bootstrap address: at most one lookup, one connection, one ClientHello naming the resolver with no ALPN offer, one 33-byte framed query and one response, then close; a proxy task without a bootstrap address is rejected at admission without contact.

Case plan: Record optional system bootstrap, one selected TCP connection, authenticated TLS, one length-prefixed DNS query and bounded matching response, then close. Fault the prefix, message length, query match and each stage; require no cleartext or resolver fallback and no application retry.

<!-- requirement-ref: REQ-DNS-04-0005 -->
[REQ-DNS-04-0005](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0005) → `M-DNS-04-EXECUTION` (Operation).

<!-- requirement-ref: REQ-DNS-04-0006 -->
[REQ-DNS-04-0006](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0006) → `M-DNS-04-EXECUTION` (Procedure).

#### M-DNS-04-NORMALIZATION

Disposition: **passed** for synthetic production cases in `dns04_conformance_test.go`, `symbol:TestMDns04Normalization`: `response` with answers and empty, `bootstrap_failure`, `tcp_timeout`, `tcp_refused`, `tcp_reset` before and during TLS, `tls_failure` for an untrusted certificate, a resolver-name mismatch, an alert and a protocol error, `dns_nxdomain`, `dns_servfail`, `dns_refused`, `malformed_dns_response` for a short prefix and an invalid message, `query_mismatch`, and `inconclusive` for an unclassified TCP failure and an unregistered response code, each at its observed stage after ordered successful prior stages. The post-connect `timeout` rows run in `TestMDns04Limits`.

Case plan: Exercise every outcome row, including TCP reset before TLS versus reset during TLS, authenticated negative DNS codes, malformed prefix/message, mismatch, post-connect timeout and inconclusive. Keep terminal failure at the observed stage with prior observations ordered.

<!-- requirement-ref: REQ-DNS-04-0007 -->
[REQ-DNS-04-0007](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0007) → `M-DNS-04-NORMALIZATION` (Success Condition).

<!-- requirement-ref: REQ-DNS-04-0008 -->
[REQ-DNS-04-0008](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0008) → `M-DNS-04-NORMALIZATION` (Failure Conditions).

<!-- requirement-ref: REQ-DNS-04-0027 -->
[REQ-DNS-04-0027](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0027) → `M-DNS-04-NORMALIZATION` (Outcomes).

#### M-DNS-04-LIMITS

Disposition: **passed** for synthetic cases in `dns04_conformance_test.go`, `symbol:TestMDns04Limits` and `symbol:TestMDns04TupleConcurrency`, and `internal/execution/dns04_termination_test.go`, `symbol:TestDns04PlaintextAccounting`: the TLS and DNS operation bounds yield `timeout` while the outer deadline is open; messages of 65,534 and 65,535 bytes are accepted and a stricter received limit one byte below the framed message ends the attempt with `resource_limit`; 16 answers are retained and 17 are `malformed_dns_response`; a sent limit at the framed query succeeds and one byte below it ends with `resource_limit` before the message leaves; one connection without retry; and an equivalent resolver and query tuple waits without contact. A message above 65,535 bytes cannot be framed.

Case plan: Test the 65,535-byte DNS message and 16-answer limit at below/equal/above boundaries under stricter common limits. Assert one resolver connection/DNS exchange after optional bootstrap, bounded chain deadline, no retries and deduplication by resolver/query tuple.

<!-- requirement-ref: REQ-DNS-04-0023 -->
[REQ-DNS-04-0023](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0023) → `M-DNS-04-LIMITS` (Resource Bounds).

#### M-DNS-04-VERSIONS

Disposition: **passed** for synthetic cases in `dns04_conformance_test.go`, `symbol:TestMDns04Versions`: task-time rejection of unsupported, malformed and missing `test_version` with no lookup, contact or run; the emitted combination decodes; the result decoder rejects every unsupported or malformed version, DNS-03 and DNS-01 identities and bindings, a bootstrap binding without the system observation, foreign outcomes, engines and query types, reordered or missing stages and unknown fields.

Case plan: Exercise task acceptance and result emission for DNS-04 with the exact registered combination test_version 0.1.0 (test.version 0.1.0), Measurement Result schema_version 0.1.0 and test.result_version 0.1.0, and assert the emitted result carries exactly test.id DNS-04 with that one combination. Iterate unsupported values (0.1.1, 0.2.0, 1.0.0), malformed values (empty string, 0.1, v0.1.0, 0.1.0 with trailing whitespace, numeric 0.1, null) for test_version, schema_version and result_version independently, and mixed combinations where only one of the three is unsupported. Reject each before execution as a task-validation failure, assert the recording transport saw no bootstrap, TCP, TLS or DNS activity and that no Measurement Result is produced. Inject cross-test combinations: a DNS-04 measurement object validated under test.id DNS-01 or DNS-03, and a DNS-03 or DNS-01 output validated as DNS-04 result_version 0.1.0, and assert rejection rather than best-effort interpretation, coercion or emission of a second combination. Verify registry Draft status and version metadata alone do not mark any other version supported. Requirements: REQ-DNS-04-0024.

<!-- requirement-ref: REQ-DNS-04-0024 -->
[REQ-DNS-04-0024](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0024) → `M-DNS-04-VERSIONS` (Output).

#### M-DNS-04-OUTPUT

Disposition: **passed** for synthetic production output in `dns04_conformance_test.go`, `symbol:TestMDns04ExecutionAndOutput` and `symbol:TestMDns04Normalization`: optional system bootstrap, TCP, TLS and `dot` DNS observations in order, bootstrap omitted with a bootstrap address and unstarted stages omitted after failure, only registered details, and at most 16 normalized answers without TTL or framing. Every result is checked for the resolver name, query name, bootstrap address, port, certificate, raw-error and loopback canaries.

Case plan: Verify optional bootstrap DNS, TCP, TLS, then dot DNS; omit bootstrap when an approved bootstrap_ip is supplied and omit unstarted stages after failure. Reuse the common DNS answer normalization and bounds without retaining framing bytes.

<!-- requirement-ref: REQ-DNS-04-0025 -->
[REQ-DNS-04-0025](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0025) → `M-DNS-04-OUTPUT` (Output).

<!-- requirement-ref: REQ-DNS-04-0026 -->
[REQ-DNS-04-0026](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0026) → `M-DNS-04-OUTPUT` (Output).

#### M-DNS-04-TERMINATION

Disposition: **passed** for synthetic cases in `internal/execution/dns04_termination_test.go`, `symbol:TestDns04TerminationPoints`, and `dns04_conformance_test.go`, `symbol:TestMDns04Termination`: before the test starts is `not_run`; during bootstrap, TCP connect, TLS establishment and after the framed query is written it is `incomplete` with the first of host cancellation, `network_changed`, the attempt deadline, injected suspension and the proxy reasons, with no measurement, further lookup, dial or write and every connection closed; while closing after the outcome is recorded it stays `completed`. Through `Client`, a generation change and a cancellation after the query and a refused tunnel are `incomplete` without direct or resolver fallback.

Case plan: Drive DNS-04 with a synthetic registered resolver profile, a scripted system resolver, a scripted TCP/TLS/DoT transport, a fake platform and a deterministic clock, and inject each interruption at each Procedure point: before input validation completes, during system bootstrap resolution of the resolver name (omitted when an approved bootstrap_ip is supplied), during the TCP connect to the one selected address on the configured port, during TLS establishment and resolver-name certificate validation, after the length-prefixed DNS query is written but before the bounded response is fully read, and while closing the session. Expected states follow the GO-SPEC-ISSUE-004 disposition: before input validation completes the attempt is not_run with that reason, no measurement and no network_context; while closing the session after the outcome is recorded it stays completed with the recorded outcome; every other point is incomplete as asserted below. At each point inject in turn a network generation change, host cancellation, expiry of the outer attempt deadline, platform suspension, and, only where a proxy path is configured, proxy loss and host proxy reconfiguration. At the incomplete points, assert execution.state incomplete with exactly the first applicable reason from the common table (network_changed, cancelled_by_host, attempt_deadline_exceeded, platform_suspended, proxy_unavailable, proxy_connect_failed or proxy_reconfigured), no measurement object, no outcome and no observations, and that no failure value such as timeout, tcp_timeout or network_changed is emitted as a measurement outcome for these interruptions. Verify the contrasting controls: a per-stage operation bound reached while the outer deadline remains open yields execution.state completed with outcome tcp_timeout or timeout, and an oversized length prefix yields completed malformed_dns_response rather than an incomplete state. Assert the recording transport shows the socket closed, no further write after the interruption, no cleartext DNS, no alternate resolver, no second connection and no application-layer retry, and that no partial NET-05 style measurement is retained. Requirements: REQ-DNS-04-0028.

<!-- requirement-ref: REQ-DNS-04-0028 -->
[REQ-DNS-04-0028](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0028) → `M-DNS-04-TERMINATION` (Outcomes).

#### M-DNS-04-MINIMIZATION

Intended owner: `internal/result/dns04.go`. Planned case: `internal/result/dns04_conformance_test.go`, `symbol:TestMDns04Minimization`. Disposition: **planned**, with no execution or support evidence.

Case plan: Exercise DNS-04 against a scripted DoT transport using one unique synthetic canary per prohibited item: a canary resolver_name, a canary resolver address returned by the scripted bootstrap, a canary non-standard port value accepted by a synthetic reviewed profile, a canary bootstrap_ip, a canary query hostname distinct from every scripted answer value, a canary marker embedded in the raw DNS response message bytes (query identifier, flags, authority and additional sections), canary TLS certificate data (subject, issuer, serial, fingerprint), canary query framing bytes in the two-byte length prefix, and a canary raw error string from each scripted stage failure. Iterate success, bootstrap_failure, tcp_refused, tls_failure, dns_refused, malformed_dns_response and query_mismatch runs, and assert serialized output of the full Measurement Result, every nested observation and every error path contains none of the canaries. Assert the measurement contains only outcome and ordered observations whose fields are limited to type, status, duration_ms, failure and details, with details limited to address_family for tcp_connect; address_family, tls_version, certificate_validation and an optional alert for tls_handshake; and engine (system or dot), address_family, query_type and responses with arrival_index, arrival_offset_ms, rcode and answers of type and value for dns_query, with at most 16 normalized answers and no TTL. Reject any extra field. Requirements: REQ-DNS-04-0029.

<!-- requirement-ref: REQ-DNS-04-0029 -->
[REQ-DNS-04-0029](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0029) → `M-DNS-04-MINIMIZATION` (Outcomes).

#### M-DNS-04-INTERPRETATION

Intended owner: `docs/measurement-interpretation-review.md`. Planned case: `internal/result/dns04_conformance_test.go`, `symbol:TestMDns04Interpretation`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/measurement/dns04/run.go`, `internal/result/dns04.go`.

Case plan: Assert serialized DNS-04 results for every registered outcome (response, bootstrap_failure, tcp_timeout, tcp_refused, tcp_reset, tls_failure, dns_nxdomain, dns_servfail, dns_refused, malformed_dns_response, query_mismatch, timeout, inconclusive) contain no interpretation, verdict, censorship, blocking, anomaly or intent field and no field outside the registered Output schema. Assert Go field documentation, outcome descriptions and the sanitized example and host-facing text contain none of these prohibited claims: that the test configures DoT for the device or host application; that a response answer is authoritative or correct; that a failure on the standard DoT port 853 proves intentional blocking; that a transport difference between DNS-04 and DNS-01 or DNS-03 establishes censor intent; that a single execution establishes more than whether each observed bootstrap, TCP, TLS and DNS stage completed; or that dns_refused, tls_failure, tcp_reset or tcp_timeout by itself excludes resolver, authoritative DNS, certificate or routing outage, bootstrap failure, resolver rate limiting or policy refusal, port-specific firewall behavior, VPN, captive portal, enterprise policy or local security software, or IPv4/IPv6 path differences. Requirements: REQ-DNS-04-0001, REQ-DNS-04-0002, REQ-DNS-04-0003, REQ-DNS-04-0030, REQ-DNS-04-0031, REQ-DNS-04-0032.

Scope: portable automated component only; judgement-based completion is review case `RV-M-DNS-04-INTERPRETATION`, which is not automated coverage.

Review case `RV-M-DNS-04-INTERPRETATION` (`docs/measurement-interpretation-review.md`, `heading:RV-M-DNS-04-INTERPRETATION`): Human review case; not automated coverage. Reviewer role: Measurement interpretation reviewer (analyst role independent of the implementer). Reviewed artifact: Go result documentation, host integration guidance and analyst guidance for DNS-04 at the reviewed implementation revision. Checklist: (1) Does the guidance limit one execution to stage completion of the observed bootstrap, TCP, TLS and DoT DNS stages for the configured resolver and query? (2) Does the guidance list resolver, authoritative DNS, certificate or routing outage, unrelated bootstrap failure, resolver rate limiting or maintenance, port-specific firewall behavior, VPN, captive portal, enterprise policy or local security software, and IPv4/IPv6 or stale configuration differences as alternative explanations for each negative outcome? (3) Does the guidance require repeated DoT results on a stable network without aggressive retries before a stronger interpretation? (4) Does the guidance require comparison with DNS-01 and DNS-03 results for the same safe query and with TCP-01 and TLS-01 observations for the approved resolver endpoint? (5) Does the guidance require independent observations from comparable times and network contexts, documenting differences in time, vantage point and method? (6) Does the guidance describe a difference between DNS transports as a transport-specific observation and not as evidence of censor intent? (7) Does the guidance avoid implying that the test configures DoT for the device or that the resolver answer is authoritative or correct? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-DNS-04-0001 -->
[REQ-DNS-04-0001](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0001) → `M-DNS-04-INTERPRETATION`, `RV-M-DNS-04-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-DNS-04-0002 -->
[REQ-DNS-04-0002](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0002) → `M-DNS-04-INTERPRETATION`, `RV-M-DNS-04-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-DNS-04-0003 -->
[REQ-DNS-04-0003](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0003) → `M-DNS-04-INTERPRETATION`, `RV-M-DNS-04-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-DNS-04-0030 -->
[REQ-DNS-04-0030](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0030) → `M-DNS-04-INTERPRETATION`, `RV-M-DNS-04-INTERPRETATION` (Supported Interpretation).

<!-- requirement-ref: REQ-DNS-04-0031 -->
[REQ-DNS-04-0031](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0031) → `M-DNS-04-INTERPRETATION`, `RV-M-DNS-04-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

<!-- requirement-ref: REQ-DNS-04-0032 -->
[REQ-DNS-04-0032](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0032) → `M-DNS-04-INTERPRETATION`, `RV-M-DNS-04-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

#### M-DNS-04-SAFETY

Intended owner: `docs/measurement-deployment-review.md`. Planned case: `internal/admission/measurement_safety_conformance_test.go`, `symbol:TestMDns04Safety`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/execution/measurements.go`, `internal/result/dns04.go`.

Case plan: Exercise DNS-04 with synthetic registered resolver profiles and query names, a recording TCP/TLS/DoT transport and a fake host policy. Iterate the Baseline Risk rows: risk level medium, where a task declaring low is rejected or evaluated as at least medium and tasks declaring medium, high or unknown keep effective risk at least medium; and requires user consent false subject to host and local policy, where a fake host or local policy requiring consent without valid consent blocks execution. Iterate each Risk Factors row (network_observability medium, traffic_distinctiveness medium, target_sensitivity medium, data_sensitivity low, legal_or_social_context medium) and assert a task factor declared below its registered level does not lower effective risk. Exercise each Safety Analysis action: data minimization, asserting only stage outcomes, coarse quantized timing and minimal DNS properties are retained; traffic distinctiveness, asserting one standard exchange with no port-range probing, TLS variant probing or fallback; user-identification risk, asserting the recorded bytes contain no identifier, authentication or unrelated application data canary; target risk, rejecting unregistered or unreviewed resolver endpoints, query names, non-standard ports without review, and local, loopback, link-local, private, interpolated or unknown targets before any traffic; and probe authority, verifying local policy can delay, skip or reject execution with not_run local_policy_rejected and no traffic. Requirements: REQ-DNS-04-0033, REQ-DNS-04-0034, REQ-DNS-04-0035, REQ-DNS-04-0036, REQ-DNS-04-0037, REQ-DNS-04-0038, REQ-DNS-04-0039.

Scope: portable automated component only; judgement-based completion is review case `RV-M-DNS-04-SAFETY`, which is not automated coverage.

Review case `RV-M-DNS-04-SAFETY` (`docs/measurement-deployment-review.md`, `heading:RV-M-DNS-04-SAFETY`): Human review case; not automated coverage. Reviewer role: Deployment safety and privacy reviewer (independent of the implementer and the deploying host team). Reviewed artifact: DNS-04 deployment configuration, registered resolver profiles and query-name list, host scheduling and consent policy, and Android and Apple native network adapter behavior at the reviewed revision. Checklist: (1) Have the resolver endpoints and query names completed technical, privacy and safety review for the intended deployment contexts, including whether encrypted DNS resolvers may be treated as circumvention infrastructure there? (2) Does the host run DNS-04 only under explicit reviewed host configuration or as an approved anomaly-triggered diagnostic? (3) Does the host scheduling frequency avoid creating a recognizable repeated pattern on the dedicated DoT port in the target environments? (4) Has the host or local policy determined whether consent is required in restrictive environments where encrypted DNS use may carry additional legal or social risk, without relying on requires_user_consent false? (5) Do native adapters produce traffic limited to one standard DoT exchange with no platform-added retries, fallback, or identifying data, as shown by captured traffic from a controlled test network? (6) Is the residual risk that the resolver sees the query and source network and the access network sees the resolver destination and transport documented to the deploying host? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-DNS-04-0033 -->
[REQ-DNS-04-0033](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0033) → `M-DNS-04-SAFETY`, `RV-M-DNS-04-SAFETY` (Baseline Risk).

<!-- requirement-ref: REQ-DNS-04-0034 -->
[REQ-DNS-04-0034](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0034) → `M-DNS-04-SAFETY`, `RV-M-DNS-04-SAFETY` (Risk Factors).

<!-- requirement-ref: REQ-DNS-04-0035 -->
[REQ-DNS-04-0035](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0035) → `M-DNS-04-SAFETY`, `RV-M-DNS-04-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-DNS-04-0036 -->
[REQ-DNS-04-0036](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0036) → `M-DNS-04-SAFETY`, `RV-M-DNS-04-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-DNS-04-0037 -->
[REQ-DNS-04-0037](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0037) → `M-DNS-04-SAFETY`, `RV-M-DNS-04-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-DNS-04-0038 -->
[REQ-DNS-04-0038](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0038) → `M-DNS-04-SAFETY`, `RV-M-DNS-04-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-DNS-04-0039 -->
[REQ-DNS-04-0039](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/dns-04.md#REQ-DNS-04-0039) → `M-DNS-04-SAFETY`, `RV-M-DNS-04-SAFETY` (Safety Analysis).

### TCP-01

#### Current TCP-01 Execution

`internal/measurement/tcp01` owns the strict codec, normalized typed parameters, proxy compatibility and runner. The root `registry.go` assembles its exact `0.1.0` entry and target adapter. `internal/protocol/tcp` owns the reusable 15-second connect bound, typed socket-error mapping and quantized observation. Execution supplies the clock, checked/metered dialer, stable-generation/family check and terminal-outcome marker. Equivalent address/port tuples wait within one registry instance until the active connection closes, under their own attempt deadlines. No application bytes, DNS lookup, address expansion, direct fallback or in-attempt retry is issued.

The executable suites below are portable synthetic evidence. The direct fixture maps the checked public literal to loopback inside its test transport; the proxy fixture maps that same CONNECT authority to loopback inside its scripted proxy. Neither fixture contacts the public destination or allows special-purpose input through the codec. Successful results pass production result/bundle validation. Native routing, suspension delivery, host integration, interpretation/safety judgement and all `RV-*` cases remain deferred.

| Family | Executable suite and coverage |
|---|---|
| `M-TCP-01-ADMISSION` | `tcp01_conformance_test.go:TestMTcp01Admission`: public IPv4/IPv6 and port endpoints, exact registration, prohibited address classes, canonical spelling, closed schema, types, duplicate/escaped names, exact integer spelling, and rejection without handle/replay/run/contact/result. `TestTcp01StrictSchemaBothOrigins` checks embedded and signed remote delivery. |
| `M-TCP-01-OPTIONS` | `TestMTcp01Options`: empty-only options and original encoded input/options byte caps at limit −1, limit and limit +1. |
| `M-TCP-01-EXECUTION` | `TestMTcp01ExecutionAndOutput`: one checked, metered direct connection or proxy tunnel, immediate close and zero application bytes, with no retry or fallback. |
| `M-TCP-01-NORMALIZATION` | `TestMTcp01Normalization`: every terminal failure vocabulary, wrapped typed socket errors, coarse unknown errors and 250-ms timing. |
| `M-TCP-01-LIMITS` | `TestMTcp01Limits`: independent operation and outer deadlines. `TestMTcp01ConnectionBudget` and `internal/execution` `TestTcp01ConnectionMeter`: one metered connection, refused before the transport at `max_connections` 0. `TestTcp01ReleasesSocketOnEveryPath`: a late socket is closed exactly once on every ending. `TestMTcp01TupleConcurrency`: equivalent tuples wait without contact, cancellation while waiting is `not_run`, and the tuple is released. The execution suite below checks cleanup and released timers/watches. |
| `M-TCP-01-TERMINATION` | `TestMTcp01Termination`, `TestMTcp01ProxyTermination`, and `internal/execution/tcp01_termination_test.go:TestTcp01AllTerminationBoundaries`: pre-start `not_run`; started `incomplete`; terminal outcome retained through close; host, network, deadline, synthetic suspension and proxy causes, with first-cause precedence and no fabricated TCP outcome. Suspension is an injected engine cause, not device evidence. |
| `M-TCP-01-OUTPUT` | `TestMTcp01ExecutionAndOutput`, `TestMTcp01Normalization`, `TestMTcp01Versions`: exactly one minimized TCP observation, selected family, fixed normalized vocabulary, quantized duration, and rejection of extra fields or cross-test observations. |
| `M-TCP-01-VERSIONS` | `TestMTcp01Versions` (the task's singular VERSION family): exact task/schema/result tuple and strict rejection of unsupported, malformed, absent and cross-test versions. |

`TestTcp01AdmissionRiskAndTimezoneDomains` asserts GO-SPEC-ISSUE-007 with the production codec; `FuzzTcp01Admission` exercises atomic input/options rejection through `Client.AdmitTask`. The codec uses the common strict JSON stages; the protocol observation producer re-decodes its generated closed JSON through the existing observation decoder. These are the TCP-01 consumer/producer inputs to the GO-SPEC-ISSUE-001 [Registered-Schema Adoption Addendum](specification-issues.md#registered-schema-adoption-addendum--october-27-2026); binding/host evidence remains pending. The eight families above are `passed` at Go `1a37603` ([evidence](../traceability/evidence/go-unit-1a37603.md)).

#### M-TCP-01-ADMISSION

Owners: `internal/measurement/tcp01`, `internal/protocol/tcp` and the root `registry.go`. Case: `tcp01_conformance_test.go`, `symbol:TestMTcp01Admission`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-TCP-01-ADMISSION). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

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

Owners: `internal/measurement/tcp01`, `internal/protocol/tcp` and the root `registry.go`. Case: `tcp01_conformance_test.go`, `symbol:TestMTcp01Options`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-TCP-01-OPTIONS). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

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

Owners: `internal/measurement/tcp01`, `internal/protocol/tcp` and the root `registry.go`. Case: `tcp01_conformance_test.go`, `symbol:TestMTcp01ExecutionAndOutput`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-TCP-01-EXECUTION). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Record one socket connect to the exact accepted address/port and immediate close without application bytes. No DNS lookup, alternate address/port, retry or additional protocol may run.

<!-- requirement-ref: REQ-TCP-01-0006 -->
[REQ-TCP-01-0006](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0006) → `M-TCP-01-EXECUTION` (Operation).

<!-- requirement-ref: REQ-TCP-01-0007 -->
[REQ-TCP-01-0007](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0007) → `M-TCP-01-EXECUTION` (Procedure).

#### M-TCP-01-NORMALIZATION

Owners: `internal/measurement/tcp01`, `internal/protocol/tcp` and the root `registry.go`. Case: `tcp01_conformance_test.go`, `symbol:TestMTcp01Normalization`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-TCP-01-NORMALIZATION). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Exercise every failure/outcome row, including refusal/reset/timeout/unreachable/permission and coarse inconclusive errors, using deterministic socket fixtures. Distinguish socket outcomes from common interruption and map to exactly one tcp_connect.

<!-- requirement-ref: REQ-TCP-01-0008 -->
[REQ-TCP-01-0008](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0008) → `M-TCP-01-NORMALIZATION` (Success Condition).

<!-- requirement-ref: REQ-TCP-01-0009 -->
[REQ-TCP-01-0009](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0009) → `M-TCP-01-NORMALIZATION` (Failure Conditions).

<!-- requirement-ref: REQ-TCP-01-0025 -->
[REQ-TCP-01-0025](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0025) → `M-TCP-01-NORMALIZATION` (Outcomes).

#### M-TCP-01-LIMITS

Owners: `internal/measurement/tcp01`, `internal/protocol/tcp` and the root `registry.go`. Case: `tcp01_conformance_test.go`, `symbol:TestMTcp01Limits`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-TCP-01-LIMITS). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Advance the operation and outer deadlines independently; assert one connection, handshake traffic only, no retries and deduplication by address/port tuple. Verify socket release on all failure/cancellation paths.

<!-- requirement-ref: REQ-TCP-01-0022 -->
[REQ-TCP-01-0022](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0022) → `M-TCP-01-LIMITS` (Resource Bounds).

#### M-TCP-01-VERSIONS

Owners: `internal/measurement/tcp01`, `internal/protocol/tcp` and the root `registry.go`. Case: `tcp01_conformance_test.go`, `symbol:TestMTcp01Versions`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-TCP-01-VERSIONS). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Exercise result construction and validation for test.id TCP-01 with test.version 0.1.0, schema_version 0.1.0 and test.result_version 0.1.0, and assert this exact registered combination is the only one emitted and accepted. Iterate rejected variants: schema_version 0.2.0, 1.0.0, 0.1, v0.1.0, empty string and numeric 0.1; test.version 0.1.1 and 0.2.0; test.result_version 0.2.0, 0.1 and missing; cross-test combinations such as test.id DNS-01 or TLS-01 carrying the TCP-01 measurement, and test.id TCP-01 carrying a dns_query or tls_handshake observation instead of tcp_connect; and an unknown field in the test object. Assert an unsupported or malformed version selected for execution is rejected before any socket is opened, and that a received result with an unsupported version is never interpreted best-effort, coerced to 0.1.0, or analyzed after discarding unknown fields (the strict decoder rejects it; retaining an opaque bounded object in a quarantine path is a receiver option in the public text, not a library obligation). Assert the Draft status and registration metadata alone do not add any supported version. Requirements: REQ-TCP-01-0023.

<!-- requirement-ref: REQ-TCP-01-0023 -->
[REQ-TCP-01-0023](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0023) → `M-TCP-01-VERSIONS` (Output).

#### M-TCP-01-OUTPUT

Owners: `internal/measurement/tcp01`, `internal/protocol/tcp` and the root `registry.go`. Case: `tcp01_conformance_test.go`, `symbol:TestMTcp01ExecutionAndOutput`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-TCP-01-OUTPUT). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Require exactly one tcp_connect with the selected family, normalized outcome and quantized duration. Reuse shared TCP observation validation; resolve the target through task identity rather than duplicating destination fields.

<!-- requirement-ref: REQ-TCP-01-0024 -->
[REQ-TCP-01-0024](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0024) → `M-TCP-01-OUTPUT` (Output).

#### M-TCP-01-TERMINATION

Owners: `internal/measurement/tcp01`, `internal/protocol/tcp` and the root `registry.go`. Case: `tcp01_conformance_test.go`, `symbol:TestMTcp01Termination`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-TCP-01-TERMINATION). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Exercise a valid synthetic TCP-01 task against a scripted transport, fake platform and deterministic clock, and interrupt at each step of the TCP-01 Procedure section: during address/port validation, during the stable network-generation and address-family check, after the connect attempt starts but before the socket reports an outcome, after a success or socket outcome is recorded but before close, and during the immediate close. At each point inject separately a network-generation change, host cancellation, expiry of the outer attempt deadline, platform suspension, and (where a proxy path is configured) proxy failure or proxy reconfiguration. Expected states follow the GO-SPEC-ISSUE-004 disposition: assert an interruption during address/port validation is not_run with exactly that reason, no measurement and no network_context; at the network-generation and address-family check, which starts the test, and while the connect is pending, assert execution.state is incomplete with exactly the first applicable reason (network_changed, cancelled_by_host, attempt_deadline_exceeded, platform_suspended, or the matching proxy reason such as proxy_unavailable, proxy_connect_failed or proxy_reconfigured), that no measurement object and no partial tcp_connect observation is emitted (TCP-01 is not NET-05); after a success or socket outcome is recorded and during the immediate close, assert execution.state stays completed with the recorded outcome unchanged, and that network_changed never appears as a measurement outcome or tcp_connect failure. Contrast with expiry of the operation deadline inside the connect step, which must yield state completed, outcome timeout and one failed tcp_connect with failure timeout. Verify on every path that the socket is released, no application bytes are written, and no retry, alternate address/port or invented terminal TCP outcome follows the interruption. Requirements: REQ-TCP-01-0026.

<!-- requirement-ref: REQ-TCP-01-0026 -->
[REQ-TCP-01-0026](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0026) → `M-TCP-01-TERMINATION` (Outcomes).

#### M-TCP-01-MINIMIZATION

Intended owner: `internal/result/tcp01.go`. Planned case: `internal/result/tcp01_conformance_test.go`, `symbol:TestMTcp01Minimization`. Disposition: **planned**, with no execution or support evidence.

Case plan: Exercise a synthetic accepted TCP-01 target through a scripted transport and fake platform that surface one unique canary per item prohibited by REQ-TCP-01-0027: a socket error string (CANARY-SOCKERR-TEXT), a socket error number (91337), the destination address (the synthetic configured address string), the destination port (synthetic port 47391), a local address (CANARY-LOCAL-ADDR), a local port (51873), TCP flags (CANARY-TCP-FLAGS), packet counts (73191), payload data (CANARY-PAYLOAD-BYTES) and retry metadata (CANARY-RETRY-META). Iterate the connected outcome and every failed outcome (timeout, connection_refused, connection_reset, network_unreachable, address_unreachable, permission_denied, socket_error, inconclusive) plus the incomplete error path. Assert the serialized result, error values and all nested observations contain none of the canaries and no field outside the allowed set; the measurement retains only outcome and one tcp_connect observation with type, status, duration_ms as a non-negative multiple of 250, failure only when failed, and details containing only address_family (ipv4 or ipv6). Assert that an injected extra detail or measurement field is rejected rather than dropped. No participant or live measurement data is used. Requirements: REQ-TCP-01-0027.

<!-- requirement-ref: REQ-TCP-01-0027 -->
[REQ-TCP-01-0027](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0027) → `M-TCP-01-MINIMIZATION` (Outcomes).

#### M-TCP-01-INTERPRETATION

Intended owner: `docs/measurement-interpretation-review.md`. Planned case: `internal/result/tcp01_conformance_test.go`, `symbol:TestMTcp01Interpretation`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/measurement/tcp01/run.go`, `internal/result/tcp01.go`.

Case plan: Assert the serialized TCP-01 result schema, the Go field descriptions and documentation strings, and the sanitized example and host-facing text for every outcome (connected, timeout, connection_refused, connection_reset, network_unreachable, address_unreachable, permission_denied, socket_error, inconclusive) contain none of these prohibited claims: that DNS resolution was performed or measured; that TLS, HTTP or another application protocol is usable or unusable; that a timeout or reset establishes intentional filtering; that a single connection result is definitive proof of censorship; or that outcome is a generic censorship, blocking, anomaly or success verdict. Assert no interpretation, verdict, censored, blocked or filtered field exists and that injecting such a field into a measurement is rejected as unknown. Assert text describing a result stays within one execution's claim that a TCP connection to the configured address and port completed or produced a particular observable socket outcome. Requirements: REQ-TCP-01-0002, REQ-TCP-01-0003, REQ-TCP-01-0004, REQ-TCP-01-0028, REQ-TCP-01-0029, REQ-TCP-01-0030.

Scope: portable automated component only; judgement-based completion is review case `RV-M-TCP-01-INTERPRETATION`, which is not automated coverage.

Review case `RV-M-TCP-01-INTERPRETATION` (`docs/measurement-interpretation-review.md`, `heading:RV-M-TCP-01-INTERPRETATION`): Human review case; not automated coverage. Reviewer role: Measurement interpretation reviewer (analyst role independent of the implementer). Reviewed artifact: Go result documentation and host/analyst guidance for TCP-01 at the reviewed implementation revision. Checklist: (1) Does the guidance limit one TCP-01 execution to whether a connection to the configured address and port completed or produced a particular observable socket outcome? (2) Does the guidance present the documented alternative explanations (target service offline, overloaded or not listening; operational firewall reject or drop; routing, peering, NAT, carrier-grade NAT, VPN or local network failure; IPv4/IPv6 path differences; transient packet loss or mobile-network handover; incorrect or stale target configuration) before any filtering hypothesis? (3) Does the guidance require the listed stronger evidence (repeated attempts under a stable network generation and conservative schedule, successful control connections on the same port, related DNS/TLS/HTTPS observations, independent comparable measurements) before stronger conclusions? (4) Does any wording, taken in context, imply that a refusal, reset or timeout indicates intentional filtering or censorship? (5) Does any wording imply that a connected outcome shows TLS, HTTP or application-level usability, or that the test resolved DNS? (6) Does the guidance avoid presenting incomplete or not_run attempts as network interference? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-TCP-01-0002 -->
[REQ-TCP-01-0002](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0002) → `M-TCP-01-INTERPRETATION`, `RV-M-TCP-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-TCP-01-0003 -->
[REQ-TCP-01-0003](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0003) → `M-TCP-01-INTERPRETATION`, `RV-M-TCP-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-TCP-01-0004 -->
[REQ-TCP-01-0004](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0004) → `M-TCP-01-INTERPRETATION`, `RV-M-TCP-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-TCP-01-0028 -->
[REQ-TCP-01-0028](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0028) → `M-TCP-01-INTERPRETATION`, `RV-M-TCP-01-INTERPRETATION` (Supported Interpretation).

<!-- requirement-ref: REQ-TCP-01-0029 -->
[REQ-TCP-01-0029](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0029) → `M-TCP-01-INTERPRETATION`, `RV-M-TCP-01-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

<!-- requirement-ref: REQ-TCP-01-0030 -->
[REQ-TCP-01-0030](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0030) → `M-TCP-01-INTERPRETATION`, `RV-M-TCP-01-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

#### M-TCP-01-SAFETY

Intended owner: `docs/measurement-deployment-review.md`. Planned case: `internal/admission/measurement_safety_conformance_test.go`, `symbol:TestMTcp01Safety`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/execution/measurements.go`, `internal/result/tcp01.go`.

Case plan: Iterate each Baseline Risk row and each Risk Factors row with synthetic tasks, a fake local policy and host, and a recording transport. Baseline: assert the registered risk level is low and requires_user_consent false is evaluated subject to host and local policy, so a fake host that requires consent blocks execution without valid consent. Factors network_observability, traffic_distinctiveness, target_sensitivity, data_sensitivity and legal_or_social_context: assert each registered low level is kept, a task declaring medium, high or unknown raises the effective risk accordingly, and effective risk is never below the registered baseline. Exercise each normative safety action: data minimization (serialized output holds only outcome, address family and quantized duration); traffic distinctiveness (recording transport sees one connect per run, port arrays or ranges rejected, equivalent concurrent runs for one address/port tuple deduplicated or delayed, no retry); user-identification risk (zero application bytes written and no identifier such as run_id in traffic); target risk (reject loopback, unspecified, multicast, link-local, private, documentation, benchmarking and other special-purpose synthetic addresses, hostnames, and targets the fake policy marks unreviewed or sensitive, with no socket opened and no measurement); and probe final authority (fake local policy delaying, skipping or rejecting a valid task prevails). Requirements: REQ-TCP-01-0031, REQ-TCP-01-0032, REQ-TCP-01-0033, REQ-TCP-01-0034, REQ-TCP-01-0035, REQ-TCP-01-0036, REQ-TCP-01-0037.

Scope: portable automated component only; judgement-based completion is review case `RV-M-TCP-01-SAFETY`, which is not automated coverage.

Review case `RV-M-TCP-01-SAFETY` (`docs/measurement-deployment-review.md`, `heading:RV-M-TCP-01-SAFETY`): Human review case; not automated coverage. Reviewer role: Deployment safety reviewer (privacy and safety role independent of the implementer). Reviewed artifact: TCP-01 native socket adapters, host scheduling and consent integration, and deployment target/port review records at the reviewed implementation revision. Checklist: (1) Do the Android and Apple native socket paths send only TCP handshake traffic with no application payload and no attached identifiers on real platform stacks? (2) Does the host scheduling configuration avoid repeated bare connections, port sequences and synchronized schedules that could resemble scanning? (3) Has each deployed target address and port been reviewed per task for target sensitivity and legal or social context in the intended deployment regions? (4) Does the host consent and local policy integration apply its own consent requirement where requires_user_consent is false but host or local policy requires consent? (5) Is the residual risk that the destination and network can observe the source address and connection attempt documented in host-facing deployment material? (6) Does the platform configuration prevent the connection from being routed through an unintended network path such as an unintended VPN or proxy? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-TCP-01-0031 -->
[REQ-TCP-01-0031](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0031) → `M-TCP-01-SAFETY`, `RV-M-TCP-01-SAFETY` (Baseline Risk).

<!-- requirement-ref: REQ-TCP-01-0032 -->
[REQ-TCP-01-0032](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0032) → `M-TCP-01-SAFETY`, `RV-M-TCP-01-SAFETY` (Risk Factors).

<!-- requirement-ref: REQ-TCP-01-0033 -->
[REQ-TCP-01-0033](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0033) → `M-TCP-01-SAFETY`, `RV-M-TCP-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-TCP-01-0034 -->
[REQ-TCP-01-0034](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0034) → `M-TCP-01-SAFETY`, `RV-M-TCP-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-TCP-01-0035 -->
[REQ-TCP-01-0035](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0035) → `M-TCP-01-SAFETY`, `RV-M-TCP-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-TCP-01-0036 -->
[REQ-TCP-01-0036](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0036) → `M-TCP-01-SAFETY`, `RV-M-TCP-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-TCP-01-0037 -->
[REQ-TCP-01-0037](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tcp-01.md#REQ-TCP-01-0037) → `M-TCP-01-SAFETY`, `RV-M-TCP-01-SAFETY` (Safety Analysis).

### TLS-01

Shared DNS/TLS vocabulary packet, handshake and result-integration cases, including HTTP-01 and NET-05 reuse, are recorded in [Runtime Vocabulary Evidence and Compatibility](observation-models.md#runtime-vocabulary-evidence-and-compatibility). These cases do not promote broader measurement or review families, and native/host adoption remains pending.

#### Current TLS-01 Execution

`internal/measurement/tls01` owns the strict codec, normalized typed parameters, proxy compatibility and runner. The root `registry.go` assembles its exact `0.1.0` entry and target adapter, whose registered tuple is the address, port, normalized `server_name` and ordered ALPN offer under scheme `tls`. `internal/protocol/tls` owns the reusable handshake: one ClientHello with the configured SNI and registered ALPN offer, TLS 1.2 or later, certificate chain and hostname validation always enabled, no session cache, tickets, renegotiation, client certificate, cipher or version control, a 15-second operation bound, typed failure mapping and a quantized observation. The runner reuses `internal/protocol/tcp` for the explicit-IP connect, records the outcome before close and sends no application data; the normal TLS closure is a record, not plaintext. Equivalent endpoint and server-name tuples wait within one registry instance. No DNS lookup, second address, alternate SNI or profile, or in-attempt retry is issued.

TLS plaintext accounting follows [REQ-LIMIT-0034](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/library/registered-limits.md#REQ-LIMIT-0034): before the handshake the runner switches its metered attempt connection with `measurement.SessionConn.Session`, so handshake and record bytes on the stream or proxy tunnel are not counted and the session's plaintext side is metered against the same budget ([Attempt Byte and Connection Budgets](execution-models.md#attempt-byte-and-connection-budgets)). TLS-01 sends and reads no plaintext, so its metered bytes are zero.

**Trust policy and time source (REQ-TLS-01-0020).** Production uses Go `crypto/x509` with the system trust roots and the device wall clock. On Android that is the Go system pool read from the platform's certificate files; on Apple platforms Go delegates verification to the platform verifier. Whether user-installed or profile roots are accepted, how platform verifier errors map to `certificate_untrusted`, `certificate_name_mismatch` and `certificate_time_invalid`, and behavior with an incorrect device clock need native evidence and the `RV-M-TLS-01-PRECONDITIONS` review (DEV-455). The loopback suites trust only a synthetic CA through the internal `measurementRoots` adapter seam, which exercises Go's own verifier.

**Mapping decisions.** Typed errors only, with raw text, alert codes and certificate data never retained:

- A received peer alert is recognized only by crypto/tls's typed `remote error`; its description maps through the REQ-OBS-0050 classes. Description 120 (`no_application_protocol`) answering a non-empty ALPN offer is `alpn_no_overlap`, without `details.alert`.
- A record that is not TLS is `protocol_error`; a reset is `connection_reset` (outcome `tcp_reset`). A peer closure without an alert, a peer selecting an unoffered ALPN identifier and other untyped library errors stay `inconclusive`, because crypto/tls exposes them only as error text.
- `certificate_validation` is `trusted` once the standard chain and hostname checks have passed, the specific class for a local verification failure, `inconclusive` for an inconclusive failure before verification, and `not_completed` otherwise. Go checks validity time before hostname before chain.
- The REQ-TLS-01-0026 outcome table covers only the `timeout`, `connection_refused`, `connection_reset` and `inconclusive` TCP failures. Another `tcp_connect` failure (`network_unreachable`, `address_unreachable`, `permission_denied`, `socket_error`) is emitted as `inconclusive` with outcome `inconclusive`, under the REQ-TLS-01-0008 `inconclusive` class. This reading has no public disposition yet.

The executable suites below are portable synthetic evidence. The direct fixture maps the checked public literal to a loopback TLS server inside its test transport; the proxy fixture maps that CONNECT authority to the same server inside its scripted proxy. Neither contacts the public destination. Successful results pass production result/bundle validation. Native trust stores, platform TLS fingerprints, suspension delivery, host integration, the PRECONDITIONS, MINIMIZATION, INTERPRETATION and SAFETY families and all `RV-*` cases remain deferred.

| Family | Executable suite and coverage |
|---|---|
| `M-TLS-01-ADMISSION` | `tls01_conformance_test.go:TestMTls01Admission`: IPv4/IPv6 and port endpoints, deterministic `server_name` normalization through 253 bytes, the exact reviewed tuple including SNI and ALPN order, IP-literal, wildcard, single-label, trailing-dot, special-use and oversized names, prohibited address classes, verification-bypass input fields and closed schema, all rejected without handle, replay, run, contact or result. |
| `M-TLS-01-OPTIONS` | `TestMTls01Options`: omitted, empty and 1–3 registered identifiers in order; unregistered, case-variant, empty, oversized, duplicate and over-count offers; cipher, version, invalid-certificate, alternate SNI, client-certificate, resumption and insecure-verification controls. |
| `M-TLS-01-EXECUTION` | `TestMTls01ExecutionAndOutput` and `internal/execution/tls01_test.go:TestTls01PlaintextAccounting`: one checked direct connection or proxy tunnel, one ClientHello with the configured SNI and ALPN order observed by the server, no plaintext, no retry or fallback; zero metered bytes while handshake bytes cross the socket and tunnel. |
| `M-TLS-01-NORMALIZATION` | `TestMTls01Normalization`: every outcome row from real loopback behavior (success with and without ALPN, reset, stall, alerts of each class, ALPN without overlap, non-TLS bytes, untrusted chain, name mismatch, expired leaf) or a wrapped platform dial error, with the required `certificate_validation` on every TLS observation and 250-ms timing. |
| `M-TLS-01-LIMITS` | `TestMTls01Limits`: the handshake bound against a stalled peer, the outer deadline and cancellation, each releasing the socket; one connection at `max_connections` 1 and 2 with one-byte byte limits. `TestMTls01TupleConcurrency`: an equivalent tuple waits without contact and is `not_run` when cancelled. `TestScopeSessionMetersPlaintext`: raw session bytes are not counted, plaintext is, and the plaintext side keeps the budget. |
| `M-TLS-01-TERMINATION` | `TestMTls01Termination` and `internal/execution/tls01_test.go:TestTls01AllTerminationBoundaries`: before the test starts `not_run`; the network check, connect, before the ClientHello and a stalled handshake `incomplete`; the outcome retained through close; host, network, deadline, synthetic suspension and proxy causes with first-cause precedence and no fabricated TLS outcome. Suspension is an injected engine cause, not device evidence. |
| `M-TLS-01-OUTPUT` | `TestMTls01ExecutionAndOutput`, `TestMTls01Normalization`, `TestMTls01Versions`: `tcp_connect` then `tls_handshake` only when TLS began, minimized details, canaries for the address, server name, certificate subject and serial and raw errors absent. |
| `M-TLS-01-VERSIONS` | `TestMTls01Versions` (the task's singular VERSION family): exact task/schema/result tuple and strict rejection of unsupported, cross-test, reordered, truncated and extra-field outputs. |

The codec uses the common strict JSON stages and shares `taskdecode.DNSName` with DNS-01; the TLS observation producer uses the existing `observation.NewTLS` decoder path. These, with `TestRegisteredSchemasStrictJSON`, are the TLS-01 consumer/producer inputs to the GO-SPEC-ISSUE-001 [Registered-Schema Adoption Addendum](specification-issues.md#registered-schema-adoption-addendum--october-27-2026). `tls01_execution_closeout_test.go` adds the no-implicit-DNS and no-session-resumption rows, and `internal/measurement/tls01` `TestMTls01ALPNBoundsIsolated` isolates the ALPN count bound at the options descriptor. The eight families above are `passed` at Go `1a37603` ([evidence](../traceability/evidence/go-unit-1a37603.md#M-TLS-01-ADMISSION)).

#### M-TLS-01-ADMISSION

Owners: `internal/measurement/tls01`, `internal/protocol/tls` and the root `registry.go`. Case: `tls01_conformance_test.go`, `symbol:TestMTls01Admission`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-TLS-01-ADMISSION). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

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

Owners: `internal/measurement/tls01`, `internal/protocol/tls` and the root `registry.go`. Case: `tls01_conformance_test.go`, `symbol:TestMTls01Options`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-TLS-01-OPTIONS). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Exercise omitted/empty ALPN, 1..8 identifiers, 1..32 ASCII bytes, ordering, registered values and duplicates at all boundaries. Only h2, http/1.1 and dot are registered offer identifiers, so through admission registry membership is the stricter check and subsumes the 1..32-byte length bound, and the 8-identifier count bound is isolated at the closed options descriptor. Reject custom cipher/version profiles, invalid-certificate controls, alternate SNI, client certificates, resumption and insecure verification.

<!-- requirement-ref: REQ-TLS-01-0016 -->
[REQ-TLS-01-0016](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0016) → `M-TLS-01-OPTIONS` (Options).

<!-- requirement-ref: REQ-TLS-01-0017 -->
[REQ-TLS-01-0017](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0017) → `M-TLS-01-OPTIONS` (Options).

#### M-TLS-01-PRECONDITIONS

Intended owner: `internal/measurement/tls01/run.go`. Planned case: `internal/measurement/tls01/conformance_test.go`, `symbol:TestMTls01Preconditions`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/platform/network.go`, `docs/measurement-platform-validation.md`.

Case plan: Run the portable TLS-01 gate against a fake platform, synthetic target registry and recording transport. Stable network generation unavailable, and execution deadline unavailable: assert execution.state not_run with no TCP connection opened. Target not registered as expected to support normal publicly authenticated TLS: assert not_run. Target registered as requiring client authentication, credentials or a sensitive server name (iterate each): assert not_run and no ClientHello sent. All preconditions met: assert exactly one TCP connection and one handshake. Stricter task execution constraint or local policy denial: assert constraints_unsatisfied or local_policy_rejected; other denials use a registered reason per the public table. Requirements: REQ-TLS-01-0018, REQ-TLS-01-0019, REQ-TLS-01-0020, REQ-TLS-01-0021, REQ-TLS-01-0022.

Scope: portable automated component only; judgement-based completion is review case `RV-M-TLS-01-PRECONDITIONS`, which is not automated coverage.

Review case `RV-M-TLS-01-PRECONDITIONS` (`docs/measurement-deployment-review.md`, `heading:RV-M-TLS-01-PRECONDITIONS`): Human review case; not automated coverage. Reviewer role: TLS and mobile platform reviewer (independent of the implementer). Reviewed artifact: Go TLS-01 implementation documentation and Android and Apple adapter trust-store and time-source integration at the reviewed implementation revision. Checklist: (1) Is the trust store used on Android documented, including whether user-installed roots are accepted? (2) Is the trust store used on Apple platforms documented, including whether user-installed or configuration-profile roots are accepted? (3) Is the time source used for certificate validity documented on each platform, including behavior with an incorrect device clock? (4) Do documented trust-store and time-source differences have stated effects on certificate_untrusted and certificate_time_invalid interpretation? (5) Is it confirmed that no native path enables client certificates or credentials? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-TLS-01-0018 -->
[REQ-TLS-01-0018](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0018) → `M-TLS-01-PRECONDITIONS`, `RV-M-TLS-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-TLS-01-0019 -->
[REQ-TLS-01-0019](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0019) → `M-TLS-01-PRECONDITIONS`, `RV-M-TLS-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-TLS-01-0020 -->
[REQ-TLS-01-0020](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0020) → `M-TLS-01-PRECONDITIONS`, `RV-M-TLS-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-TLS-01-0021 -->
[REQ-TLS-01-0021](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0021) → `M-TLS-01-PRECONDITIONS`, `RV-M-TLS-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-TLS-01-0022 -->
[REQ-TLS-01-0022](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0022) → `M-TLS-01-PRECONDITIONS`, `RV-M-TLS-01-PRECONDITIONS` (Preconditions).

#### M-TLS-01-EXECUTION

Owners: `internal/measurement/tls01`, `internal/protocol/tls` and the root `registry.go`. Case: `tls01_conformance_test.go`, `symbol:TestMTls01ExecutionAndOutput`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-TLS-01-EXECUTION). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Record explicit-IP TCP then one normal SNI/ALPN handshake with certificate/hostname validation; retain first failed stage or minimized negotiated properties and close before application data. No implicit DNS, alternate profiles, SNI comparison, sessions or retries.

<!-- requirement-ref: REQ-TLS-01-0005 -->
[REQ-TLS-01-0005](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0005) → `M-TLS-01-EXECUTION` (Operation).

<!-- requirement-ref: REQ-TLS-01-0006 -->
[REQ-TLS-01-0006](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0006) → `M-TLS-01-EXECUTION` (Procedure).

#### M-TLS-01-NORMALIZATION

Owners: `internal/measurement/tls01`, `internal/protocol/tls` and the root `registry.go`. Case: `tls01_conformance_test.go`, `symbol:TestMTls01Normalization`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-TLS-01-NORMALIZATION). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Drive TLS-01 through a scripted transport and deterministic clock and iterate every outcome row. Assert each outcome and its terminal observation: handshake_succeeded with tls_handshake succeeded and certificate_validation trusted; tcp_timeout, tcp_refused with tcp_connect failed timeout or connection_refused and no tls_handshake observation; tcp_reset from tcp_connect failed connection_reset, and separately from tls_handshake failed connection_reset after a succeeded tcp_connect; tls_timeout with tls_handshake failed timeout; tls_alert with tls_handshake failed alert, with details.alert either absent or a registered coarse alert class and no raw alert; protocol_error with tls_handshake failed protocol_error; certificate_untrusted, certificate_name_mismatch, certificate_time_invalid with the same failure value and certificate_validation untrusted, name_mismatch or time_invalid respectively; alpn_no_overlap with tls_handshake failed alpn_no_overlap; inconclusive when the terminal TCP or TLS observation failed inconclusive. Assert the required certificate_validation value is present on every tls_handshake observation and the stage-specific failure is preserved rather than collapsed into a single filtering outcome. Inject network_changed and the outer deadline and assert an incomplete execution state, not a TLS outcome. Assert success only when the handshake completes, the chain is accepted by the documented trust policy and the certificate is valid for server_name. Requirements: REQ-TLS-01-0007, REQ-TLS-01-0008, REQ-TLS-01-0026.

<!-- requirement-ref: REQ-TLS-01-0007 -->
[REQ-TLS-01-0007](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0007) → `M-TLS-01-NORMALIZATION` (Success Condition).

<!-- requirement-ref: REQ-TLS-01-0008 -->
[REQ-TLS-01-0008](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0008) → `M-TLS-01-NORMALIZATION` (Failure Conditions).

<!-- requirement-ref: REQ-TLS-01-0026 -->
[REQ-TLS-01-0026](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0026) → `M-TLS-01-NORMALIZATION` (Outcomes).

#### M-TLS-01-LIMITS

Owners: `internal/measurement/tls01`, `internal/protocol/tls` and the root `registry.go`. Case: `tls01_conformance_test.go`, `symbol:TestMTls01Limits`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-TLS-01-LIMITS). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Assert one connection/handshake, full-chain deadline, no application payload/retries and deduplication by endpoint/server-name tuple. Test cancellation and close while the peer stalls in handshake.

<!-- requirement-ref: REQ-TLS-01-0023 -->
[REQ-TLS-01-0023](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0023) → `M-TLS-01-LIMITS` (Resource Bounds).

#### M-TLS-01-VERSIONS

Owners: `internal/measurement/tls01`, `internal/protocol/tls` and the root `registry.go`. Case: `tls01_conformance_test.go`, `symbol:TestMTls01Versions`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-TLS-01-VERSIONS). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Select the exact registered combination test.id TLS-01, test.version 0.1.0, Measurement Result schema_version 0.1.0 and test.result_version 0.1.0, and Assert a synthetic completed TLS-01 result validates only with that combination. Reject, before execution or interpretation and without coercion or best-effort parsing: schema_version 0.2.0, 1.0.0, 0.1 and empty or non-string values; test.version 0.2.0 or 0.1; test.result_version 0.2.0 or missing; a TLS-01 measurement labelled test.id TCP-01 or HTTP-01; a TLS-01 envelope carrying the TCP-01 outcome connected or any outcome not in the TLS-01 Outcomes table; a TLS-01 result containing an observation type other than tcp_connect or tls_handshake; and unknown fields in the envelope, test object or measurement. Verify a task referencing an unsupported TLS-01 version is a task-validation failure reported outside Measurement Results, that no additional version is added or inferred, and that the Draft status and registration metadata of TLS-01 confer no support claim. Requirements: REQ-TLS-01-0024.

<!-- requirement-ref: REQ-TLS-01-0024 -->
[REQ-TLS-01-0024](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0024) → `M-TLS-01-VERSIONS` (Output).

#### M-TLS-01-OUTPUT

Owners: `internal/measurement/tls01`, `internal/protocol/tls` and the root `registry.go`. Case: `tls01_conformance_test.go`, `symbol:TestMTls01ExecutionAndOutput`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-TLS-01-OUTPUT). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Require TCP then TLS only when handshake activity began, in execution order. Check minimized allowed negotiated properties against common TCP/TLS observations with certificate-validation classes preserved.

<!-- requirement-ref: REQ-TLS-01-0025 -->
[REQ-TLS-01-0025](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0025) → `M-TLS-01-OUTPUT` (Output).

#### M-TLS-01-TERMINATION

Owners: `internal/measurement/tls01`, `internal/protocol/tls` and the root `registry.go`. Case: `tls01_conformance_test.go`, `symbol:TestMTls01Termination`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-TLS-01-TERMINATION). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Exercise a valid TLS-01 task against a synthetic target accepted by a test-only local policy fixture, using a deterministic clock, fake platform and scripted TCP/TLS transport, and interrupt at each Procedure point: after input validation (step 1), after network-generation and address-family confirmation (step 2), while the TCP connect is pending (step 3), after TCP success before the ClientHello is sent, while the TLS handshake is stalled mid-exchange (step 4), and after negotiated properties are recorded but before close (steps 5-6). Expected states follow the GO-SPEC-ISSUE-004 disposition: after input validation (step 1) the attempt is not_run with that reason, no measurement and no network_context; step 2 starts the test; after the negotiated properties are recorded and before close it stays completed with the recorded outcome; every other point is incomplete as asserted below. At each point inject in turn a network-generation change, host cancellation, outer attempt-deadline expiry, proxy reconfiguration (and other registered proxy_* reasons where a proxy path is configured, per the public table) and platform suspension. At the incomplete points, assert execution.state is incomplete with execution.reason exactly network_changed, cancelled_by_host, attempt_deadline_exceeded, proxy_reconfigured (or the matching proxy_* reason) or platform_suspended, naming only the first condition; assert measurement is absent at those points because TLS-01 defines no partial Output (only NET-05 may retain partial measurement); assert no outcome is network_changed or any other invented TLS outcome, no tcp_connect or tls_handshake observation is emitted for the interrupted stage, the socket is closed, no application payload is sent, and no second address, server name, TLS profile or retry is attempted. Verify the contrast cases: a bounded TCP or TLS operation timeout before the outer deadline yields execution.state completed with outcome tcp_timeout or tls_timeout, while expiry of the outer deadline yields attempt_deadline_exceeded with no measurement. Requirements: REQ-TLS-01-0027.

<!-- requirement-ref: REQ-TLS-01-0027 -->
[REQ-TLS-01-0027](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0027) → `M-TLS-01-TERMINATION` (Outcomes).

#### M-TLS-01-MINIMIZATION

Intended owner: `internal/result/tls01.go`. Planned case: `internal/result/tls01_conformance_test.go`, `symbol:TestMTls01Minimization`. Disposition: **planned**, with no execution or support evidence.

Case plan: Exercise TLS-01 with a synthetic task and scripted TCP/TLS transport and fake trust store, and inject one unique synthetic canary into each item REQ-TLS-01-0028 names and that the shared tls_handshake observation contract also prohibits: the destination IP address, the destination port, server_name, a certificate DER blob and a certificate-chain element, certificate subject, issuer and SAN names, certificate serial number, public key and signature, certificate fingerprint, the raw TLS alert code and description, ClientHello fields (client random, extension bytes), ServerHello fields (server random, extension bytes), the offered cipher-suite list, the negotiated cipher suite, the session identifier and session ticket, TLS secrets (key-log material), and the raw library error string. Drive the success path and each failure class (tcp_timeout, tcp_refused, tcp_reset, tls_timeout, tls_alert, protocol_error, certificate_untrusted, certificate_name_mismatch, certificate_time_invalid, alpn_no_overlap, inconclusive) plus incomplete and not_run paths. Assert serialized result and error paths, including every nested observation, contain no canary and no input digest, and contain only the allowed fields: measurement.outcome, and per observation type, status, duration_ms, failure, details.address_family, and for tls_handshake details.tls_version, details.alpn, details.certificate_validation and a registered coarse details.alert class; assert any extra field is rejected. No participant or live measurement data is used. Requirements: REQ-TLS-01-0028.

<!-- requirement-ref: REQ-TLS-01-0028 -->
[REQ-TLS-01-0028](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0028) → `M-TLS-01-MINIMIZATION` (Outcomes).

#### M-TLS-01-INTERPRETATION

Intended owner: `docs/measurement-interpretation-review.md`. Planned case: `internal/result/tls01_conformance_test.go`, `symbol:TestMTls01Interpretation`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/measurement/tls01/run.go`, `internal/result/tls01.go`.

Case plan: Assert the TLS-01 result schema has no interpretation, verdict, anomaly, blocked, censored, filtered or confidence field and that such a field is rejected as unknown. Assert result field descriptions, documented outcome descriptions and the sanitized example and host-facing text contain none of these prohibited claims: that a handshake failure establishes TLS or SNI filtering, censorship, blocking, interference or intent; that a certificate_untrusted, certificate_name_mismatch or certificate_time_invalid outcome establishes interception or tampering rather than target, CDN, trust-store, certificate or device-time misconfiguration; that tcp_timeout, tcp_reset or tls_timeout excludes routing, packet loss, mobile handover, VPN, firewall, proxy, rate limiting or overload; that alpn_no_overlap or protocol_error excludes unsupported TLS version, algorithm or ALPN configuration; that handshake_succeeded establishes HTTP or application-content availability; that a result compares different SNI values; or that a failure excludes a stale address-to-name mapping. Verify outcome text is limited to whether TCP and normal authenticated TLS completed for the configured address, port, server name and ALPN offer. Requirements: REQ-TLS-01-0001, REQ-TLS-01-0002, REQ-TLS-01-0003, REQ-TLS-01-0029, REQ-TLS-01-0030, REQ-TLS-01-0031.

Scope: portable automated component only; judgement-based completion is review case `RV-M-TLS-01-INTERPRETATION`, which is not automated coverage.

Review case `RV-M-TLS-01-INTERPRETATION` (`docs/measurement-interpretation-review.md`, `heading:RV-M-TLS-01-INTERPRETATION`): Human review case; not automated coverage. Reviewer role: Measurement interpretation reviewer (analyst role independent of the implementer). Reviewed artifact: Go result documentation and host/analyst guidance for TLS-01 at the reviewed implementation revision. Checklist: (1) Does the guidance limit a single execution to whether TCP and normal authenticated TLS completed for the configured address, port, server name and ALPN offer? (2) Does the guidance require a TCP-01 result for the same address and port, repeated TLS observations under a stable network generation, successful control handshakes on comparable infrastructure, and HTTP-01 results with independent observations before any stronger interpretation? (3) Does the guidance present every listed alternative explanation (target, load balancer, CDN, certificate or trust-store misconfiguration; expired certificate or incorrect device time; routing, loss, handover, VPN, firewall or proxy behavior; unsupported TLS version, algorithm or ALPN; stale address-to-name mapping; server-side rate limiting or overload) as plausible for negative outcomes? (4) Does the guidance state that different-SNI comparison is outside TLS-01 and requires a separately reviewed measurement, without suggesting hosts combine TLS-01 runs into an SNI comparison? (5) Does the guidance link the cross-cutting Measurement Limitations and keep any derived anomaly analysis separate from the source result? (6) Is any wording that frames a certificate-validation failure as interception or filtering absent from analyst-facing material? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-TLS-01-0001 -->
[REQ-TLS-01-0001](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0001) → `M-TLS-01-INTERPRETATION`, `RV-M-TLS-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-TLS-01-0002 -->
[REQ-TLS-01-0002](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0002) → `M-TLS-01-INTERPRETATION`, `RV-M-TLS-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-TLS-01-0003 -->
[REQ-TLS-01-0003](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0003) → `M-TLS-01-INTERPRETATION`, `RV-M-TLS-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-TLS-01-0029 -->
[REQ-TLS-01-0029](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0029) → `M-TLS-01-INTERPRETATION`, `RV-M-TLS-01-INTERPRETATION` (Supported Interpretation).

<!-- requirement-ref: REQ-TLS-01-0030 -->
[REQ-TLS-01-0030](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0030) → `M-TLS-01-INTERPRETATION`, `RV-M-TLS-01-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

<!-- requirement-ref: REQ-TLS-01-0031 -->
[REQ-TLS-01-0031](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0031) → `M-TLS-01-INTERPRETATION`, `RV-M-TLS-01-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

#### M-TLS-01-SAFETY

Intended owner: `docs/measurement-deployment-review.md`. Planned case: `internal/admission/measurement_safety_conformance_test.go`, `symbol:TestMTls01Safety`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/execution/measurements.go`, `internal/result/tls01.go`.

Case plan: Iterate every Baseline Risk and Risk Factors row of TLS-01 using synthetic targets, a fake local policy and a recording TCP/TLS transport. Baseline Risk rows: Assert registered risk level low and requires_user_consent false, and Verify a fake host or local policy that requires consent or declines execution still blocks the run (not_run with the applicable reason per the public table). Risk Factors rows network_observability, traffic_distinctiveness, target_sensitivity, data_sensitivity and legal_or_social_context, each low: Assert a task declaring risk_level low is evaluated at the registered baseline, a task adding a medium, high or unknown factor raises effective risk to the most restrictive value, and no task declaration lowers the baseline. Exercise each normative safety action: data minimization, asserting no full certificate chain, raw ClientHello or ServerHello, session key or packet trace is retained; traffic distinctiveness, asserting the recording transport sees exactly one TCP connection to the configured IP and port, one standards-compliant ClientHello with the configured server_name and registered ALPN list in order, no malformed, alternate-SNI or comparative handshake, no implicit DNS lookup and no retry; user-identification risk, asserting no client certificate, identifier, cookie or application data byte is sent and the session closes after the handshake; target risk, asserting an unreviewed address, port and server-name tuple is rejected before any traffic; repeated-execution risk, asserting one active run per endpoint and server-name tuple with equivalent runs deduplicated or delayed. Verify the probe can delay, skip or reject a run that the task permits. Requirements: REQ-TLS-01-0032, REQ-TLS-01-0033, REQ-TLS-01-0034, REQ-TLS-01-0035, REQ-TLS-01-0036, REQ-TLS-01-0037, REQ-TLS-01-0038.

Scope: portable automated component only; judgement-based completion is review case `RV-M-TLS-01-SAFETY`, which is not automated coverage.

Review case `RV-M-TLS-01-SAFETY` (`docs/measurement-deployment-review.md`, `heading:RV-M-TLS-01-SAFETY`): Human review case; not automated coverage. Reviewer role: Deployment safety and privacy reviewer (independent of the implementer). Reviewed artifact: TLS-01 native Android and Apple TLS adapter behavior, host scheduling and consent integration, and the deployment target list at the reviewed implementation revision. Checklist: (1) Does the native platform TLS stack used by each adapter produce platform-typical ClientHello behavior, and is any fingerprint difference from the host application's own traffic documented as a residual risk? (2) Has each deployed address, port and server-name tuple been reviewed as non-sensitive for the deployment's legal and social context, including plaintext SNI visibility? (3) Does the host scheduling configuration avoid repeated bare handshakes forming a recognizable pattern for the deployed targets? (4) Do host and local policy consent decisions apply where the deployment context warrants consent despite the registered requires_user_consent false? (5) Are the documented trust policy and time source for each platform adapter recorded and reviewed? (6) Is any task in the deployment that raises effective risk above low reviewed and accompanied by the matching risk factors? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-TLS-01-0032 -->
[REQ-TLS-01-0032](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0032) → `M-TLS-01-SAFETY`, `RV-M-TLS-01-SAFETY` (Baseline Risk).

<!-- requirement-ref: REQ-TLS-01-0033 -->
[REQ-TLS-01-0033](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0033) → `M-TLS-01-SAFETY`, `RV-M-TLS-01-SAFETY` (Risk Factors).

<!-- requirement-ref: REQ-TLS-01-0034 -->
[REQ-TLS-01-0034](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0034) → `M-TLS-01-SAFETY`, `RV-M-TLS-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-TLS-01-0035 -->
[REQ-TLS-01-0035](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0035) → `M-TLS-01-SAFETY`, `RV-M-TLS-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-TLS-01-0036 -->
[REQ-TLS-01-0036](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0036) → `M-TLS-01-SAFETY`, `RV-M-TLS-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-TLS-01-0037 -->
[REQ-TLS-01-0037](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0037) → `M-TLS-01-SAFETY`, `RV-M-TLS-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-TLS-01-0038 -->
[REQ-TLS-01-0038](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/tls-01.md#REQ-TLS-01-0038) → `M-TLS-01-SAFETY`, `RV-M-TLS-01-SAFETY` (Safety Analysis).

### HTTP-01

#### M-HTTP-01-ADMISSION

Owners: `internal/measurement/http01`, `internal/protocol/http`, the shared execution and result boundaries and the root `registry.go`. Case: `http01_conformance_test.go`, `symbol:TestMHttp01Admission`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-HTTP-01-ADMISSION). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

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

<!-- requirement-ref: REQ-HTTP-01-0051 -->
[REQ-HTTP-01-0051](https://github.com/tv-rain/barometer-spec/blob/e6257ec64564ad415a2d6c1f8d106fa1763d7de8/tests/http-01.md#REQ-HTTP-01-0051) → `M-HTTP-01-ADMISSION`, `M-HTTP-01-EXECUTION` (Input Validation). A proxy-path initial URL must be an explicit IP literal; a proxy-path hostname redirect is `redirect_rejected` before resolution, CONNECT or contact.

#### M-HTTP-01-OPTIONS

Owners: `internal/measurement/http01`, `internal/protocol/http`, the shared execution and result boundaries and the root `registry.go`. Case: `http01_conformance_test.go`, `symbol:TestMHttp01Options`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-HTTP-01-OPTIONS). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Test auto/ipv4/ipv6, max_body_bytes default 4096 and 1..65536 with stricter task bounds, follow_redirects default false and conditional max_redirects 1/2. Reject max_redirects when disabled and every custom method/header/body/cookie/credential/proxy/compression/TLS override.

<!-- requirement-ref: REQ-HTTP-01-0024 -->
[REQ-HTTP-01-0024](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0024) → `M-HTTP-01-OPTIONS` (Options).

<!-- requirement-ref: REQ-HTTP-01-0025 -->
[REQ-HTTP-01-0025](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0025) → `M-HTTP-01-OPTIONS` (Options).

#### M-HTTP-01-PRECONDITIONS

Intended owner: `internal/measurement/http01/run.go`. Planned case: `internal/measurement/http01/conformance_test.go`, `symbol:TestMHttp01Preconditions`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/platform/network.go`, `docs/measurement-platform-validation.md`.

Case plan: Run the portable HTTP-01 gate with a synthetic local target registry, deterministic clock, fake platform and recording transport. URL not locally registered, or registered but not current: assert execution.state not_run and no DNS, TCP, TLS or HTTP activity. Expected response properties not registered or not current: assert not_run. A permitted redirect target or response check not registered with the measurement target (iterate each): assert not_run. Stable network generation or task execution deadline unavailable: assert not_run. Target whose local registration does not mark it as a small response suitable for bounded mobile measurement: assert not_run. Fake HTTP stack reporting that credentials, cookies or identifying headers would be attached and cannot be disabled (iterate each): assert not_run and no request sent; when they can be disabled, assert the recorded request carries none of them. Stricter execution constraint or local policy denial: assert constraints_unsatisfied or local_policy_rejected; other denials use a registered reason per the public table. Requirements: REQ-HTTP-01-0026, REQ-HTTP-01-0027, REQ-HTTP-01-0028, REQ-HTTP-01-0029, REQ-HTTP-01-0030, REQ-HTTP-01-0031.

Scope: portable automated component only; judgement-based completion is review case `RV-M-HTTP-01-PRECONDITIONS`, which is not automated coverage.

Review case `RV-M-HTTP-01-PRECONDITIONS` (`docs/measurement-deployment-review.md`, `heading:RV-M-HTTP-01-PRECONDITIONS`): Human review case; not automated coverage. Reviewer role: Mobile platform privacy reviewer (independent of the implementer). Reviewed artifact: Android and Apple HTTP stack configuration used by HTTP-01 and host integration guidance at the reviewed implementation revision. Checklist: (1) On Android, is it confirmed that the HTTP stack attaches no cookies, credentials or identifying headers by default or via host configuration? (2) On Apple platforms, is it confirmed that the HTTP stack attaches no cookies, credentials, cached authentication or identifying headers by default or via host configuration? (3) Is the default User-Agent or equivalent header documented and non-identifying on each platform? (4) Does host guidance prevent shared cookie stores or interceptors from applying to HTTP-01 requests? (5) Are platform cases where suppression cannot be established documented as reasons the test does not run? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-HTTP-01-0026 -->
[REQ-HTTP-01-0026](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0026) → `M-HTTP-01-PRECONDITIONS`, `RV-M-HTTP-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-HTTP-01-0027 -->
[REQ-HTTP-01-0027](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0027) → `M-HTTP-01-PRECONDITIONS`, `RV-M-HTTP-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-HTTP-01-0028 -->
[REQ-HTTP-01-0028](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0028) → `M-HTTP-01-PRECONDITIONS`, `RV-M-HTTP-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-HTTP-01-0029 -->
[REQ-HTTP-01-0029](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0029) → `M-HTTP-01-PRECONDITIONS`, `RV-M-HTTP-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-HTTP-01-0030 -->
[REQ-HTTP-01-0030](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0030) → `M-HTTP-01-PRECONDITIONS`, `RV-M-HTTP-01-PRECONDITIONS` (Preconditions).

<!-- requirement-ref: REQ-HTTP-01-0031 -->
[REQ-HTTP-01-0031](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0031) → `M-HTTP-01-PRECONDITIONS`, `RV-M-HTTP-01-PRECONDITIONS` (Preconditions).

#### M-HTTP-01-EXECUTION

Owners: `internal/measurement/http01`, `internal/protocol/http`, the shared execution and result boundaries and the root `registry.go`. Case: `http01_conformance_test.go`, `symbol:TestMHttp01Execution`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-HTTP-01-EXECUTION). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Record system DNS, TCP, authenticated TLS, ordinary bodyless credential-free GET requesting identity encoding, bounded redirect decision and terminal checks/body close. Validate each next target and resolved address before contact; no downgrade, unsafe redirects, excessive redirect contact or test-layer retries.

<!-- requirement-ref: REQ-HTTP-01-0005 -->
[REQ-HTTP-01-0005](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0005) → `M-HTTP-01-EXECUTION` (Operation).

<!-- requirement-ref: REQ-HTTP-01-0006 -->
[REQ-HTTP-01-0006](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0006) → `M-HTTP-01-EXECUTION` (Operation).

<!-- requirement-ref: REQ-HTTP-01-0007 -->
[REQ-HTTP-01-0007](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0007) → `M-HTTP-01-EXECUTION` (Procedure).

#### M-HTTP-01-NORMALIZATION

Owners: `internal/measurement/http01`, `internal/protocol/http`, the shared execution and result boundaries and the root `registry.go`. Case: `http01_conformance_test.go`, `symbol:TestMHttp01Normalization`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-HTTP-01-NORMALIZATION). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Exercise every outcome row plus configured-check combinations and partial/unevaluated checks, body completion versus truncation and unsupported content encoding. Preserve failure at the observed stage and indexed decidable check outcomes; do not invent verdicts for checks not evaluated or select undocumented precedence. The inconclusive outcome cannot arise through the runner, because crypto/tls never returns a zero-byte read without an error; it is exercised at the bounded HTTP exchange primitive and by decoding the registered result shape.

<!-- requirement-ref: REQ-HTTP-01-0052 -->
[REQ-HTTP-01-0052](https://github.com/tv-rain/barometer-spec/blob/e6257ec64564ad415a2d6c1f8d106fa1763d7de8/tests/http-01.md#REQ-HTTP-01-0052) → `M-HTTP-01-NORMALIZATION` (Outcomes). Overlapping terminal-response conditions use the merged order; preserve decidable checks and common interruption precedence. GO-SPEC-ISSUE-017 records its adoption.

<!-- requirement-ref: REQ-HTTP-01-0008 -->
[REQ-HTTP-01-0008](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0008) → `M-HTTP-01-NORMALIZATION` (Success Condition).

<!-- requirement-ref: REQ-HTTP-01-0009 -->
[REQ-HTTP-01-0009](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0009) → `M-HTTP-01-NORMALIZATION` (Failure Conditions).

<!-- requirement-ref: REQ-HTTP-01-0036 -->
[REQ-HTTP-01-0036](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0036) → `M-HTTP-01-NORMALIZATION` (Outcomes).

<!-- requirement-ref: REQ-HTTP-01-0037 -->
[REQ-HTTP-01-0037](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0037) → `M-HTTP-01-NORMALIZATION` (Outcomes).

#### M-HTTP-01-LIMITS

Owners: `internal/measurement/http01`, `internal/protocol/http`, the shared execution and result boundaries and the root `registry.go`. Case: `http01_conformance_test.go`, `symbol:TestMHttp01Limits`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-HTTP-01-LIMITS). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Test 16,384 response-header bytes, configured body bound and at most two followed redirects within outer connection/byte/time bounds at below/equal/above boundaries. Verify close at the active bound, no test retries and URL deduplication; distinguish application bytes processed from lower-layer buffering/on-wire use.

<!-- requirement-ref: REQ-HTTP-01-0032 -->
[REQ-HTTP-01-0032](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0032) → `M-HTTP-01-LIMITS` (Resource Bounds).

#### M-HTTP-01-VERSIONS

Owners: `internal/measurement/http01`, `internal/protocol/http`, the shared execution and result boundaries and the root `registry.go`. Case: `http01_conformance_test.go`, `symbol:TestMHttp01Versions`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-HTTP-01-VERSIONS). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Exercise task acceptance and result validation with the single registered combination test.id HTTP-01, test.version 0.1.0, schema_version 0.1.0 and test.result_version 0.1.0, and assert that an emitted result carries exactly that combination. Iterate unsupported and malformed values for each field independently: 0.2.0, 1.0.0, 0.1, v0.1.0, 0.1.0 with trailing whitespace, empty string, numeric 0.1, null and missing. Iterate cross-test combinations: test.id HTTP-01 with a measurement shaped as the NET-01 access-only object or carrying NET-05 profile/step fields, NET-03 or TLS-01 test.id with an HTTP-01 measurement object, and lowercase http-01. Assert task-side unsupported versions are task-validation failures that produce no Measurement Result and no network contact on the recording transport, and that a receiver rejects or quarantines unsupported result versions without best-effort parsing, coercion to 0.1.0, dropping unknown fields or adding another version; Draft status confers no support. Requirements: REQ-HTTP-01-0033.

<!-- requirement-ref: REQ-HTTP-01-0033 -->
[REQ-HTTP-01-0033](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0033) → `M-HTTP-01-VERSIONS` (Output).

#### M-HTTP-01-OUTPUT

Owners: `internal/measurement/http01`, `internal/protocol/http`, the shared execution and result boundaries and the root `registry.go`. Case: `http01_conformance_test.go`, `symbol:TestMHttp01Output`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-HTTP-01-OUTPUT). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Validate ordered observations only for performed DNS/TCP/TLS/HTTP operations, including bounded redirect chains and connection-reuse omissions. Terminal HTTP retains only status, minimized redirect class, bounded body metadata and indexed checks; no request/response content.

<!-- requirement-ref: REQ-HTTP-01-0034 -->
[REQ-HTTP-01-0034](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0034) → `M-HTTP-01-OUTPUT` (Output).

<!-- requirement-ref: REQ-HTTP-01-0035 -->
[REQ-HTTP-01-0035](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0035) → `M-HTTP-01-OUTPUT` (Output).

#### M-HTTP-01-TERMINATION

Owners: `internal/measurement/http01`, `internal/protocol/http`, the shared execution and result boundaries and the root `registry.go`. Case: `http01_conformance_test.go`, `symbol:TestMHttp01Termination`, with the supporting tests named in [its evidence](../traceability/evidence/go-unit-1a37603.md#M-HTTP-01-TERMINATION). Disposition: **passed** at Go `1a37603`; partial evidence only, with no implementation or support claim.

Case plan: Drive HTTP-01 with a synthetic registered HTTPS target, a scripted resolver/TCP/TLS/HTTP transport, a fake network-generation source and a deterministic clock, and inject each interruption at each Procedure point: after input validation and before DNS (step 1-2), during the system DNS query (step 3), after DNS before TCP, during TCP connect and during the TLS handshake (step 4), after the GET is sent while awaiting response headers and during header read (step 5), after a redirect response is received with following disabled and with following enabled before next-URL validation (steps 6-7), during each redirect hop's DNS, TCP, TLS and HTTP stages for max_redirects 1 and 2, and during terminal-response body reading below max_body_bytes (step 8). Expected states follow the GO-SPEC-ISSUE-004 disposition: after input validation and before the test's first network-state observation or DNS query the attempt is not_run with that reason, no measurement and no network_context; every other listed point precedes the terminal outcome and is incomplete as asserted below. Inject in turn a network-generation change, host cancellation, outer attempt deadline expiry, a non-time outer connection or data bound, platform suspension and, where the host has configured a proxy path, proxy_unavailable, proxy_connect_failed and proxy_reconfigured. At the incomplete points, assert execution.state is incomplete with exactly the first matching reason from the common table (network_changed, cancelled_by_host, attempt_deadline_exceeded, resource_limit, platform_suspended, proxy_unavailable, proxy_connect_failed, proxy_reconfigured), measurement is absent at those points because HTTP-01 does not register bounded partial output, and no HTTP outcome (including network_changed, http_timeout or inconclusive) or check outcome is invented. Verify a bounded protocol timeout before the outer deadline still yields state completed with outcome http_timeout, that an uninterrupted run yields completed with measurement, that the response body and connections are closed, and that the recording transport shows no retry, fallback stack, further redirect hop or new connection after the interruption. Requirements: REQ-HTTP-01-0038.

<!-- requirement-ref: REQ-HTTP-01-0038 -->
[REQ-HTTP-01-0038](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0038) → `M-HTTP-01-TERMINATION` (Outcomes).

#### M-HTTP-01-MINIMIZATION

Intended owner: `internal/result/http01.go`. Planned case: `internal/result/http01_conformance_test.go`, `symbol:TestMHttp01Minimization`. Disposition: **planned**, with no execution or support evidence.

Case plan: Exercise HTTP-01 against scripted transports with a unique synthetic canary per prohibited item: requested URL, redirect Location URL, hostname, path, query, resolved destination address, request header value, response header name, response header value, request-side configured header_checks value, response body bytes, expected sha256 body digest in body_checks, computed observed body digest, matched contains/prefix/suffix strings, a Set-Cookie value, a credential in a WWW-Authenticate challenge and in redirect userinfo, certificate subject/issuer/serial, a cached response body, and raw transport/TLS/HTTP error text. Cover expected_response, redirect_observed, redirect_rejected, redirect_limit_reached, unexpected_status, header_check_failed, body_check_failed, body_limit_reached, unsupported_content_encoding and DNS/TCP/TLS failure paths. Assert the serialized result, every nested observation and every error/task-validation path contain no canary in any encoding and no field beyond the registered set: outcome, observations[].type, status, duration_ms, failure; dns_query engine, query_type, responses; tcp_connect address_family; tls_handshake address_family, tls_version, alpn, certificate_validation, alert; http_exchange address_family, protocol, status_code, redirect.followed_count, redirect.terminal_location_class, body.bytes_processed, body.truncated, checks[].kind/index/matched and content_encoding. Requirements: REQ-HTTP-01-0039.

<!-- requirement-ref: REQ-HTTP-01-0039 -->
[REQ-HTTP-01-0039](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0039) → `M-HTTP-01-MINIMIZATION` (Outcomes).

#### M-HTTP-01-INTERPRETATION

Intended owner: `docs/measurement-interpretation-review.md`. Planned case: `internal/result/http01_conformance_test.go`, `symbol:TestMHttp01Interpretation`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/measurement/http01/run.go`, `internal/result/http01.go`.

Case plan: Assert the HTTP-01 Output schema, serialized results for every registered outcome and every documented field description and sanitized example or host-facing text contain no field outside outcome and observations, no interpretation, verdict, blocking, censorship, block-page or manipulation field, and none of these prohibited claims: that the test browsed or rendered a page; that it authenticated, submitted user data or executed scripts; that an unexpected status, redirect_observed, redirect_rejected, header_check_failed or body_check_failed outcome establishes a block page, intentional manipulation or blocking; that body.bytes_processed or max_body_bytes is an exact on-wire byte limit; or that a result excludes origin, CDN, DNS, routing, certificate or deployment failure, intentional redirect or changed content, captive portal, enterprise proxy, VPN, antivirus or local filtering, server-side geolocation, rate limiting, load balancing or maintenance, IPv4/IPv6 or transient mobile-network differences, stale expected status, body pattern, digest or header expectation, or legitimate redirect-chain changes. Verify each documented outcome is described only as whether the endpoint or approved terminal redirect target responded and whether bounded status, header and body checks matched. Requirements: REQ-HTTP-01-0001, REQ-HTTP-01-0002, REQ-HTTP-01-0003, REQ-HTTP-01-0040, REQ-HTTP-01-0041, REQ-HTTP-01-0042, REQ-HTTP-01-0043.

Scope: portable automated component only; judgement-based completion is review case `RV-M-HTTP-01-INTERPRETATION`, which is not automated coverage.

Review case `RV-M-HTTP-01-INTERPRETATION` (`docs/measurement-interpretation-review.md`, `heading:RV-M-HTTP-01-INTERPRETATION`): Human review case; not automated coverage. Reviewer role: Measurement interpretation reviewer (analyst role independent of the implementer). Reviewed artifact: Go result documentation, host/analyst guidance and any derived-analysis description for HTTP-01 at the reviewed implementation revision. Checklist: (1) Does guidance limit a single execution to the supported interpretation: the endpoint or approved terminal redirect target responded and its bounded status, header and body properties matched or did not match registered expectations on the observed DNS, TCP, TLS and HTTP path? (2) Does guidance require DNS-01, TCP-01 and TLS-01 observations for related approved targets, a NET-03 result excluding obvious validation or captive-portal problems, repeated observations under a stable network generation, and suitable control endpoints and independent reference signals before any stronger interpretation? (3) Are response differences and redirects routed to manual or separately reviewed classification before they can support a stronger blocking interpretation? (4) Does guidance present each listed alternative explanation as a live possibility for unexpected_status, redirect, header/body check and failure outcomes rather than as something the result rules out? (5) Is a stale registered expectation (status, body pattern, digest or header) treated as a cause to re-review the target registration rather than as evidence of interference? (6) Does any wording elsewhere in guidance imply that an HTTP-01 outcome is proof of censorship, a block page or intentional manipulation? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-HTTP-01-0001 -->
[REQ-HTTP-01-0001](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0001) → `M-HTTP-01-INTERPRETATION`, `RV-M-HTTP-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-HTTP-01-0002 -->
[REQ-HTTP-01-0002](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0002) → `M-HTTP-01-INTERPRETATION`, `RV-M-HTTP-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-HTTP-01-0003 -->
[REQ-HTTP-01-0003](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0003) → `M-HTTP-01-INTERPRETATION`, `RV-M-HTTP-01-INTERPRETATION` (Non-goals).

<!-- requirement-ref: REQ-HTTP-01-0040 -->
[REQ-HTTP-01-0040](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0040) → `M-HTTP-01-INTERPRETATION`, `RV-M-HTTP-01-INTERPRETATION` (Supported Interpretation).

<!-- requirement-ref: REQ-HTTP-01-0041 -->
[REQ-HTTP-01-0041](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0041) → `M-HTTP-01-INTERPRETATION`, `RV-M-HTTP-01-INTERPRETATION` (Supported Interpretation).

<!-- requirement-ref: REQ-HTTP-01-0042 -->
[REQ-HTTP-01-0042](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0042) → `M-HTTP-01-INTERPRETATION`, `RV-M-HTTP-01-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

<!-- requirement-ref: REQ-HTTP-01-0043 -->
[REQ-HTTP-01-0043](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0043) → `M-HTTP-01-INTERPRETATION`, `RV-M-HTTP-01-INTERPRETATION` (Evidence Needed for Stronger Interpretation).

#### M-HTTP-01-SAFETY

Intended owner: `docs/measurement-deployment-review.md`. Planned case: `internal/admission/measurement_safety_conformance_test.go`, `symbol:TestMHttp01Safety`. Disposition: **planned**, with no execution or support evidence. Additional intended owners: `internal/admission/measurements.go`, `internal/execution/measurements.go`, `internal/result/http01.go`.

Case plan: Exercise task safety evaluation and execution for HTTP-01 using synthetic registered targets, a fake local policy and host consent source, and recording transports. Iterate Baseline Risk rows: risk level low (task declaring low, medium, high and unknown yields effective risk low, medium, high and unknown respectively, never below low; a task with a factor more restrictive than its overall risk_level or an unknown factor type is rejected or treated as unknown) and requires user consent false subject to host and local policy (task false with a host policy requiring consent and no consent yields not_run local_policy_rejected with no network contact; task true without valid consent does not execute). Iterate each Risk Factors row: network_observability (recording transport shows only the registered origin over authenticated TLS, no plaintext HTTP), traffic_distinctiveness (one ordinary bodyless GET with identity encoding, no measurement identifiers or unusual headers), target_sensitivity (unregistered, private or local target rejected before DNS), data_sensitivity (body read stops at max_body_bytes and is not retained) and legal_or_social_context (low baseline cannot be lowered by the task). Exercise each Safety Analysis action: data minimization (no body, full headers, matched strings, cookies, full redirect URL or unrelated certificate metadata in output), traffic distinctiveness (no repeated redirect chain; equivalent runs for one normalized URL deduplicated or delayed), user-identification risk (synthetic cookie jar, credential store, user-state cache and app identifier are present yet never sent or used), target risk (unregistered or non-HTTPS redirect from a registered URL yields redirect_rejected and is not contacted) and probe final authority (fake local policy delay, skip and reject each prevent contact). Requirements: REQ-HTTP-01-0044, REQ-HTTP-01-0045, REQ-HTTP-01-0046, REQ-HTTP-01-0047, REQ-HTTP-01-0048, REQ-HTTP-01-0049, REQ-HTTP-01-0050.

Scope: portable automated component only; judgement-based completion is review case `RV-M-HTTP-01-SAFETY`, which is not automated coverage.

Review case `RV-M-HTTP-01-SAFETY` (`docs/measurement-deployment-review.md`, `heading:RV-M-HTTP-01-SAFETY`): Human review case; not automated coverage. Reviewer role: Deployment safety and privacy reviewer (independent of the implementer, with platform-policy and legal input as required). Reviewed artifact: HTTP-01 target and redirect-target registrations, task safety declarations, host scheduling and consent integration, and the native Android/Apple HTTP stack configuration for the reviewed deployment revision. Checklist: (1) Has target and content sensitivity of every registered URL and redirect target been reviewed for the specific deployment's legal and social context? (2) Does the deployment schedule avoid making repeated requests to a fixed validation path recognizable, and is any residual recognizability documented? (3) Is it documented that each redirect target sees the source network and request fingerprint, and that the native TLS behavior may differ from host application traffic? (4) Does the native HTTP stack configuration on each platform disable cookies, credentials, user-state caches and application-specific identifying headers, with the test declined where a default cannot be disabled? (5) Is the possibility that the stack receives more bytes on the wire than the application processes documented for data-cost and observability review? (6) Does the host consent and local policy integration preserve the probe's authority to delay, skip or reject and never lower the registered low baseline? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-HTTP-01-0044 -->
[REQ-HTTP-01-0044](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0044) → `M-HTTP-01-SAFETY`, `RV-M-HTTP-01-SAFETY` (Baseline Risk).

<!-- requirement-ref: REQ-HTTP-01-0045 -->
[REQ-HTTP-01-0045](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0045) → `M-HTTP-01-SAFETY`, `RV-M-HTTP-01-SAFETY` (Risk Factors).

<!-- requirement-ref: REQ-HTTP-01-0046 -->
[REQ-HTTP-01-0046](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0046) → `M-HTTP-01-SAFETY`, `RV-M-HTTP-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-HTTP-01-0047 -->
[REQ-HTTP-01-0047](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0047) → `M-HTTP-01-SAFETY`, `RV-M-HTTP-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-HTTP-01-0048 -->
[REQ-HTTP-01-0048](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0048) → `M-HTTP-01-SAFETY`, `RV-M-HTTP-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-HTTP-01-0049 -->
[REQ-HTTP-01-0049](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0049) → `M-HTTP-01-SAFETY`, `RV-M-HTTP-01-SAFETY` (Safety Analysis).

<!-- requirement-ref: REQ-HTTP-01-0050 -->
[REQ-HTTP-01-0050](https://github.com/tv-rain/barometer-spec/blob/59d49364d85e48cc8e69c762d130a7a86dc1fe65/tests/http-01.md#REQ-HTTP-01-0050) → `M-HTTP-01-SAFETY`, `RV-M-HTTP-01-SAFETY` (Safety Analysis).

## Maintenance and Validation

Before adopting a new public or Go baseline, review source/catalog/issue changes, recheck all measurement IDs and case applicability, and update pinned public links together with the manifest at the milestone cadence of coordination decision 0012. Earlier manifests stay at the commits where they were current and are never presented as current evidence. A newly allocated ID must receive its own reviewed disposition; planned work must not silently become implemented or passed.

The manifest maps every allocated measurement ID at the selected public baseline. Passed families and unit-test cases carry conformance evidence; the other targets and cases are planned and support arrays are empty. Shared IDs remain in the separately scoped cross-cutting manifest; residual unallocated authoritative prose, threat-model/measurement-limitations allocation and full semantic inventory review remain visible. GO-SPEC-ISSUE-001 is linked to test-specific schema/option/version work, GO-SPEC-ISSUE-017 to the HTTP-01 outcome and check rules REQ-HTTP-01-0052 resolved, and GO-SPEC-ISSUE-018 to the NET-05 profile bound; their immutable findings remain available at the Go baseline.

Reproduce the metadata report with `make traceability` as described in [Traceability CI](traceability-ci.md#reproduce-locally). The checker receives explicitly supplied public/Go repositories, reviewed schema/tooling, expected baselines and current mode under [Manifest Validation](traceability-manifest.md). Verify exact ID/case/owner/scope correspondence between this plan and the manifest; the checker does not execute planned cases or prove semantic completeness.

Run public allocation-history/reference checks with explicitly supplied Go/internal consumers, Go `make check`, documentation/link checks and `git diff --check`. Public tooling dependencies remain development-only and are not added to the Go runtime module or its canonical checks. Public merge precedes this Go adoption; no runtime release or support promotion is introduced. Present the validated Go change and wait for separate final approval before commit/push.
