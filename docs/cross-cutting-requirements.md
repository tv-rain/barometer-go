# Cross-cutting Requirements and Case Plan

## Status, Authority, and Baselines

This planning map identifies shared format and library obligations, intended implementation owners and concrete conformance case families. The linked public text is the oracle; this document neither copies nor changes its normative behavior. A case family is a parameterized suite, and planned families are not conformance. No complete semantic review, implemented requirement, core/platform/field support or release is claimed.

The [manifest](../traceability/cross-cutting.json) is pinned to public Draft `0.1.0` at `5d01879cfd1d47396c780befa2567f7e90439948` and Go `d788a80754e9f3f55852601bf8df16db4b1a1808`, the closeout milestone of Network paths and context (DEV-578). It maps all 591 active shared requirements (308 `partial`, 283 `planned`) to the case families below and to `UT.*` unit-test cases; its evidence records point to the [evidence report](../traceability/evidence/go-unit-d788a80.md). Executed families are listed in [Runtime Case Families](execution-models.md#runtime-case-families) and in each family's Scope line below.

Baselines and evidence are updated at the milestone cadence of the coordination [decision 0012](https://github.com/tv-rain/barometer/blob/main/docs/decisions/0012-lean-task-process.md): at the closeout of a sub-epic or on request, not after every task. Between updates, case-plan notes and `requirement-ref` comments may run ahead of the manifest. Earlier mappings are retained at the commits where they were current; `git log -- traceability/cross-cutting.json` lists them, and the [changelog](../CHANGELOG.md) records what each change added.

## Boundaries and Maintenance

The current target-validation notes adopt public `bb805f07a36630db7250d5ab63006ed5e351f7a7` for REQ-PROXY-0043 through REQ-PROXY-0046, REQ-CONTEXT-0043 and REQ-HTTP-01-0051. Their predecessors are retired; current comments/links use successors. New gate tests do not rewrite earlier milestone evidence or promote mixed-family status.

The step-8 input and options notes adopt public `12062471875bbfbc55eafc94dc5a7ea4cd07ca5e` for REQ-TRUST-0056 and REQ-OPTIONS-0033 ([GO-SPEC-ISSUE-010](specification-issues.md#go-spec-issue-010-admission-class-for-step-8-input-and-options-failures)). REQ-OPTIONS-0019 is retired there with successor REQ-OPTIONS-0033; its mapping keeps a `requirement-history` comment and its original baseline link only.

Public PR #8 resolved the DEV-481 [vocabulary finding](specification-issues.md#go-spec-issue-002-missing-registered-observation-vocabularies) by retiring REQ-OBS-0011 and REQ-OBS-0021 in favor of REQ-OBS-0042 and REQ-OBS-0043, and adding REQ-OBS-0044 to REQ-OBS-0051. The successors take over their predecessors' case families; the new requirements use six `CC-OBS-VOCAB-*` families. The retired IDs keep `requirement-history` comments and links to the original PR #6 baseline only. They are not current obligations, and no prior claim is rewritten to a successor. Mapping these requirements establishes no runtime vocabulary adoption evidence; GO-SPEC-ISSUE-002 remains open for that downstream evidence.

The 16 format/library sources plus shared JSON decoding are in scope. Registered test-specific schemas/runners, the full threat-model and measurement-limitations inventories, implementation of these cases, native adapters, host and collector software, and field validation remain outside this mapping task. Their applicability is preserved as report gaps, not silently treated as covered. The [normative source inventory](normative-sources.md) supplies the wider source boundary; this map also includes the subsequently adopted JSON contract.

Portable Go owns its package targets. Native storage and binding obligations map to planned native validation contracts in this repository, consumed by future Android/Apple adapter work; these documents cannot implement platform behavior. Host, collector and public authoring obligations map to planned integration/review records. Mixed core/integration case families must retain both scopes; Go tests alone cannot close them. Synthetic requests and canary values are local fixtures, not measurements or participant data.

Unresolved contract findings use [Specification Issues](specification-issues.md). `GO-SPEC-ISSUE-001` remains open for downstream evidence despite the accepted JSON clarification. `GO-SPEC-ISSUE-007` is adopted in Go decoding and validation and remains open for an admission-level case, the planned family cases and binding and host evidence. `GO-SPEC-ISSUE-008` is adopted in the admission design, case plans, `task.AdmissionError` and the admission steps and remains open for the planned admission families, the safety review of the mapping and binding and host evidence. `GO-SPEC-ISSUE-009` is adopted in admission step 8, the case plans and the execution-time pre-attempt check, and remains open for production compatibility rules and the safety review of the mapping. `GO-SPEC-ISSUE-010` records that the class of a step-8 input or options failure is unmapped and awaits public disposition. `GO-SPEC-ISSUE-011` records that interval due times and catch-up are unspecified; Go implements an approved conservative interim rule pending public disposition. Mapping a planned test never resolves it. A newly discovered ambiguity must block only the affected claim and enter that workflow before implementation selects behavior.

Public prerequisite ordering is satisfied: PR #6 merged the approved allocation head `9dbd72af37205a08484435cb561a168ac7b1876b` as merge commit `47d75925b3829fccfb6954b545c15f695375b76f`, and PR #8 merged approved head `5869bc314c877e2ccd3426b82aadcd9c92414f02` as `a944ed6c389fec9f0d089f0044ead6fe845328b1`. Between those baselines, the shared catalog changes only by the REQ-OBS retirements and additions above. Later public changes require renewed source/catalog review and current requirement references and case-plan notes. Update immutable manifest baselines and evidence reports at the milestone cadence above; retain earlier mappings at their original commits rather than copying them or presenting old evidence as current. Keep partial coverage while wider source, test-specific or semantic inventory gaps remain.

Before any future implemented claim, replace planned locations with actual symbols/cases, execute the complete applicable suites at the selected baselines, attach sanitized immutable conformance and completeness-review evidence, and resolve affected specification issues. Platform and field claims require their separately scoped evidence. Do not promote statuses based on compilation or case-plan presence.

## Source Allocation Summary

| Public source | New IDs in public PR #6 |
|---|---:|
| `data-formats/measurement-observations.md` | 41 |
| `data-formats/measurement-result.md` | 50 |
| `data-formats/network-context.md` | 23 |
| `data-formats/result-bundle.md` | 36 |
| `data-formats/task-execution.md` | 30 |
| `data-formats/task-input.md` | 34 |
| `data-formats/task-labels.md` | 29 |
| `data-formats/task-options.md` | 32 |
| `data-formats/task-safety.md` | 30 |
| `data-formats/task-schedule.md` | 42 |
| `data-formats/task.md` | 5 |
| `library/network-context.md` | 42 |
| `library/proxy.md` | 41 |
| `library/registered-limits.md` | 31 |
| `library/result-linkage.md` | 52 |
| `library/task-trust.md` | 34 |
| **Total added** | **552** |

The 13 previously allocated shared IDs brought the PR #6 plan to 565 requirements; the existing TCP-01 ID belongs to the separately scoped measurement mapping. PR #8 adds 10 `data-formats/measurement-observations.md` IDs and retires two, giving 573 active shared requirements at the current baseline. Registered test-specific IDs remain in the [measurement mapping](measurement-requirements.md).

## Inventory and Planned Cases

### CC-OBS-OVERVIEW

Public source: `data-formats/measurement-observations.md`, section **Overview**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsOverview`.

Case plan: Exercise TCP-01, TLS-01 and HTTP-01 style attempts against scripted transports that perform a known sequence of DNS, TCP, TLS and HTTP operations on a deterministic clock, each stage carrying synthetic raw-traffic canaries (DNS message bytes, TCP payload, TLS ClientHello/ServerHello bytes, HTTP header values and body bytes). Assert the observations array follows the injected execution order, omits stages that were never attempted, and stays within the registered observation and encoded-size limits; assert each serialized observation carries only the registered common fields (type, status, duration_ms, failure, details) and the registered details schema for its type, and that no canary byte sequence appears anywhere in the serialized result. Reject decoding of a synthetic observation that adds an unregistered field carrying raw bytes, a debug-log string or a censorship classification label. Assert the measurement outcome comes only from the registered test outcome rules, with no observation-level verdict. Requirements: REQ-OBS-0001, REQ-OBS-0002.

Scope: Synthetic core contract cases. No execution evidence or support claim.

Review case `RV-CC-OBS-OVERVIEW` (`docs/conformance/cross-cutting-review.md`, `heading:RV-CC-OBS-OVERVIEW`): Human review case; not automated coverage. Reviewer role: Measurement interpretation reviewer (analyst role independent of the implementer). Reviewed artifact: Go observation types, exported documentation and host/analyst guidance for measurement observations at the reviewed implementation revision. Checklist: (1) Does the documentation describe observations as normalized, bounded stage records of one Measurement Result rather than packet traces or debug logs? (2) Does no exported API, type name or doc comment present an observation as an independent censorship classification? (3) Does the guidance state that the registered test, not the observation, defines how observations determine the test-specific outcome? (4) Are any documented diagnostic or logging paths kept separate from observation output and free of raw traffic? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-OBS-0001 -->
<!-- requirement-ref: REQ-OBS-0002 -->

- [REQ-OBS-0001](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0001) → `CC-OBS-OVERVIEW`, `RV-CC-OBS-OVERVIEW`.
- [REQ-OBS-0002](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0002) → `CC-OBS-OVERVIEW`, `RV-CC-OBS-OVERVIEW`.

### CC-OBS-COMMON-STRUCTURE

Public source: `data-formats/measurement-observations.md`, section **Common Structure**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsCommonStructure`.

Case plan: Test common field presence/types/status-dependent failure, unknown fields and allowed test-specific observation types. Preserve order and omit never-attempted operations.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0003 -->
<!-- requirement-ref: REQ-OBS-0004 -->
<!-- requirement-ref: REQ-OBS-0005 -->
<!-- requirement-ref: REQ-OBS-0006 -->

- [REQ-OBS-0003](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0003) → `CC-OBS-COMMON-STRUCTURE`.
- [REQ-OBS-0004](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0004) → `CC-OBS-COMMON-STRUCTURE`.
- [REQ-OBS-0005](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0005) → `CC-OBS-COMMON-STRUCTURE`.
- [REQ-OBS-0006](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0006) → `CC-OBS-COMMON-STRUCTURE`.

### CC-OBS-DNS-QUERY-FAILURES

Public source: `data-formats/measurement-observations.md`, section **DNS Query / Failures**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsDnsQueryFailures`.

Case plan: Test each normalized DNS failure and observable response-code pairing; a system resolver with no code must not invent one.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0007 -->
<!-- requirement-ref: REQ-OBS-0008 -->

- [REQ-OBS-0007](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0007) → `CC-OBS-DNS-QUERY-FAILURES`.
- [REQ-OBS-0008](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0008) → `CC-OBS-DNS-QUERY-FAILURES`.

### CC-OBS-DNS-QUERY-DETAILS

Public source: `data-formats/measurement-observations.md`, section **DNS Query / Details**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsDnsQueryDetails`.

Case plan: Exercise engine/family/query-type/response-count fields and canary prohibited resolver/raw-message fields.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0009 -->
<!-- requirement-ref: REQ-OBS-0010 -->

- [REQ-OBS-0009](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0009) → `CC-OBS-DNS-QUERY-DETAILS`.
- [REQ-OBS-0010](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0010) → `CC-OBS-DNS-QUERY-DETAILS`.

### CC-OBS-DNS-QUERY-DNS-RESPONSE

Public source: `data-formats/measurement-observations.md`, section **DNS Query / DNS Response**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsDnsQueryDnsResponse`.

Case plan: Exercise sequential arrival indexes, quantized timing, closed normalized rcode values and bounded answer/CNAME counts.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0042 -->
<!-- requirement-history: REQ-OBS-0011 -->

- [REQ-OBS-0042](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0042) → `CC-OBS-DNS-QUERY-DNS-RESPONSE`. Successor of retired [REQ-OBS-0011](https://github.com/tv-rain/barometer-spec/blob/47d75925b3829fccfb6954b545c15f695375b76f/data-formats/measurement-observations.md#REQ-OBS-0011), whose mapping is preserved only in the historical manifest.

### CC-OBS-DNS-QUERY-DNS-ANSWER

Public source: `data-formats/measurement-observations.md`, section **DNS Query / DNS Answer**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsDnsQueryDnsAnswer`.

Case plan: Test canonical A/AAAA/CNAME values, first-occurrence deduplication, TTL omission, reviewed-query limits and no on-device enrichment.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0012 -->
<!-- requirement-ref: REQ-OBS-0013 -->
<!-- requirement-ref: REQ-OBS-0014 -->

- [REQ-OBS-0012](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0012) → `CC-OBS-DNS-QUERY-DNS-ANSWER`.
- [REQ-OBS-0013](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0013) → `CC-OBS-DNS-QUERY-DNS-ANSWER`.
- [REQ-OBS-0014](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0014) → `CC-OBS-DNS-QUERY-DNS-ANSWER`.

### CC-OBS-TCP-CONNECT-FAILURES

Public source: `data-formats/measurement-observations.md`, section **TCP Connect / Failures**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsTcpConnectFailures`.

Case plan: Map each TCP failure conservatively and keep outer cancellation/network/proxy/deadline conditions exclusively in execution.reason.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0015 -->
<!-- requirement-ref: REQ-OBS-0016 -->

- [REQ-OBS-0015](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0015) → `CC-OBS-TCP-CONNECT-FAILURES`.
- [REQ-OBS-0016](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0016) → `CC-OBS-TCP-CONNECT-FAILURES`.

### CC-OBS-TCP-CONNECT-DETAILS

Public source: `data-formats/measurement-observations.md`, section **TCP Connect / Details**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsTcpConnectDetails`.

Case plan: Exercise tcp_connect observations from a scripted transport for synthetic IPv4 (192.0.2.10) and IPv6 (2001:db8::10) destinations on synthetic port 8443, for succeeded and failed (connection_refused, timeout) operations. Assert details contains exactly address_family with ipv4 or ipv6 matching the attempted destination family, and that neither the destination address in any textual or byte form nor the port number appears anywhere in the serialized observation or result. Iterate decoder inputs and reject: details without address_family, address_family values unknown, IPv4, ipv4-mapped and empty string, and details carrying an extra address or port field. Accept the two valid values. Requirements: REQ-OBS-0017, REQ-OBS-0018.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0017 -->
<!-- requirement-ref: REQ-OBS-0018 -->

- [REQ-OBS-0017](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0017) → `CC-OBS-TCP-CONNECT-DETAILS`.
- [REQ-OBS-0018](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0018) → `CC-OBS-TCP-CONNECT-DETAILS`.

### CC-OBS-TLS-HANDSHAKE-FAILURES

Public source: `data-formats/measurement-observations.md`, section **TLS Handshake / Failures**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsTlsHandshakeFailures`.

Case plan: Drive a scripted transport through synthetic TCP and TLS stages and assert the serialized observation list. Iterate every registered tls_handshake failure value (timeout, connection_reset, alert, protocol_error, certificate_untrusted, certificate_name_mismatch, certificate_time_invalid, alpn_no_overlap, inconclusive): each yields a tcp_connect observation with status succeeded followed by one tls_handshake observation with status failed and that exact failure value. Reject an unregistered failure string (for example tls_reset) in a tls_handshake observation at validation. Inject a TCP-stage failure (timeout, then connection_refused) and assert it is emitted only as a preceding tcp_connect observation with status failed and the TCP failure value, with no tls_handshake observation present because TLS activity never began. Inject a successful TCP connect followed by cancellation before any TLS bytes are written and assert no tls_handshake observation is emitted. Requirements: REQ-OBS-0019, REQ-OBS-0020.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0019 -->
<!-- requirement-ref: REQ-OBS-0020 -->

- [REQ-OBS-0019](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0019) → `CC-OBS-TLS-HANDSHAKE-FAILURES`.
- [REQ-OBS-0020](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0020) → `CC-OBS-TLS-HANDSHAKE-FAILURES`.

### CC-OBS-TLS-HANDSHAKE-DETAILS

Public source: `data-formats/measurement-observations.md`, section **TLS Handshake / Details**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsTlsHandshakeDetails`.

Case plan: Exercise conditional negotiated fields with closed version/ALPN/alert vocabularies and bounded ALPN; exclude every prohibited certificate/handshake/key/session field.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0043 -->
<!-- requirement-history: REQ-OBS-0021 -->
<!-- requirement-ref: REQ-OBS-0022 -->

- [REQ-OBS-0043](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0043) → `CC-OBS-TLS-HANDSHAKE-DETAILS`. Successor of retired [REQ-OBS-0021](https://github.com/tv-rain/barometer-spec/blob/47d75925b3829fccfb6954b545c15f695375b76f/data-formats/measurement-observations.md#REQ-OBS-0021), whose mapping is preserved only in the historical manifest.
- [REQ-OBS-0022](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0022) → `CC-OBS-TLS-HANDSHAKE-DETAILS`.

### CC-OBS-HTTP-EXCHANGE-FAILURES

Public source: `data-formats/measurement-observations.md`, section **HTTP Exchange / Failures**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsHttpExchangeFailures`.

Case plan: Test registered HTTP failures and preceding DNS/TCP/TLS observations; emit HTTP only once HTTP activity starts.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0023 -->
<!-- requirement-ref: REQ-OBS-0024 -->

- [REQ-OBS-0023](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0023) → `CC-OBS-HTTP-EXCHANGE-FAILURES`.
- [REQ-OBS-0024](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0024) → `CC-OBS-HTTP-EXCHANGE-FAILURES`.

### CC-OBS-HTTP-EXCHANGE-DETAILS

Public source: `data-formats/measurement-observations.md`, section **HTTP Exchange / Details**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsHttpExchangeDetails`.

Case plan: Construct synthetic http_exchange observations and validate their details object against the public table. Required fields: omit address_family or protocol and assert rejection; iterate address_family over ipv4, ipv6 and unknown and protocol over http/1.1, h2 and unknown and assert acceptance, and reject out-of-vocabulary values (for example h3). Conditional fields: status_code accepted only as an integer valid HTTP status code and absent when no valid status was received; redirect present only when at least one redirect was observed and only as a complete minimized redirect object; body present only when body processing began and only as a complete body summary; content_encoding accepted only as identity, unsupported or unknown. Checks count: accept a checks array of 17 entries (one status, 8 header, 8 body) and reject 18 entries, a second status check, a ninth header check and a ninth body check; assert checks is absent when no checks were configured. Requirements: REQ-OBS-0025.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0025 -->

- [REQ-OBS-0025](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0025) → `CC-OBS-HTTP-EXCHANGE-DETAILS`.

### CC-OBS-HTTP-EXCHANGE-REDIRECT

Public source: `data-formats/measurement-observations.md`, section **HTTP Exchange / Redirect**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsHttpExchangeRedirect`.

Case plan: Test redirect counts/classes including uncontacted rejected redirects; exclude full destination data.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0026 -->
<!-- requirement-ref: REQ-OBS-0027 -->

- [REQ-OBS-0026](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0026) → `CC-OBS-HTTP-EXCHANGE-REDIRECT`.
- [REQ-OBS-0027](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0027) → `CC-OBS-HTTP-EXCHANGE-REDIRECT`.

### CC-OBS-HTTP-EXCHANGE-BODY

Public source: `data-formats/measurement-observations.md`, section **HTTP Exchange / Body**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsHttpExchangeBody`.

Case plan: Exercise HTTP-01 against a scripted HTTP server on a deterministic clock with identity-encoded synthetic bodies carrying canary bytes, using the registered terminal-body bound of 4,096 bytes (default) and a configured 65,536 bytes (maximum). Iterate body lengths 0, 4,095, 4,096 and 4,097 under the default bound and 65,535, 65,536 and 65,537 under the maximum. Assert bytes_processed is non-negative and never exceeds the configured bound, equals the full body length when the body completed within the bound with truncated false, and equals the bound with truncated true when the bound stopped processing before the body completed (the exactly-at-bound case follows the public table). Configure a synthetic expected body digest and assert only the check result (kind body, index, matched) is reported, while the body bytes, the computed digest and the expected digest are absent from the serialized result. Reject decoding of a body object with bytes_processed above the configured bound, negative bytes_processed, missing truncated, or an extra body-content or digest field. Requirements: REQ-OBS-0028, REQ-OBS-0029.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0028 -->
<!-- requirement-ref: REQ-OBS-0029 -->

- [REQ-OBS-0028](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0028) → `CC-OBS-HTTP-EXCHANGE-BODY`.
- [REQ-OBS-0029](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0029) → `CC-OBS-HTTP-EXCHANGE-BODY`.

### CC-OBS-HTTP-EXCHANGE-CHECK

Public source: `data-formats/measurement-observations.md`, section **HTTP Exchange / Check**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsHttpExchangeCheck`.

Case plan: Test kind/index/matched associations without retaining expected or actual header/body/digest values.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0030 -->
<!-- requirement-ref: REQ-OBS-0031 -->

- [REQ-OBS-0030](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0030) → `CC-OBS-HTTP-EXCHANGE-CHECK`.
- [REQ-OBS-0031](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0031) → `CC-OBS-HTTP-EXCHANGE-CHECK`.

### CC-OBS-VOCAB-CONTRACT-SELECTION

Public source: `data-formats/measurement-observations.md`, section **Registered Normalized Vocabularies / Contract Selection and Producer Boundary**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsVocabContractSelection`.

Case plan: Test that vocabularies are closed with no runtime registry lookup; decoders reject aliases, wrong case, numeric strings and raw library messages without repairing them into coarse markers; keep missing evidence, malformed data and out-of-subset values distinct.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0044 -->
<!-- requirement-ref: REQ-OBS-0045 -->

- [REQ-OBS-0044](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0044) → `CC-OBS-VOCAB-CONTRACT-SELECTION`.
- [REQ-OBS-0045](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0045) → `CC-OBS-VOCAB-CONTRACT-SELECTION`.

### CC-OBS-VOCAB-DNS-RESPONSE-CODE

Public source: `data-formats/measurement-observations.md`, section **Registered Normalized Vocabularies / DNS Response-Code Normalization**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsVocabDnsResponseCode`.

Case plan: Map each effective code 0-5 and an out-of-subset code including extended bits; omit responses without a reliable code; exclude raw codes/diagnostics; keep OTHER distinct from success/no-data and map unregistered nonzero codes conservatively to inconclusive.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0046 -->
<!-- requirement-ref: REQ-OBS-0047 -->

- [REQ-OBS-0046](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0046) → `CC-OBS-VOCAB-DNS-RESPONSE-CODE`.
- [REQ-OBS-0047](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0047) → `CC-OBS-VOCAB-DNS-RESPONSE-CODE`.

### CC-OBS-VOCAB-TLS-NEGOTIATED-PROPERTIES

Public source: `data-formats/measurement-observations.md`, section **Registered Normalized Vocabularies / TLS Negotiated Properties**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsVocabTlsNegotiatedProperties`.

Case plan: Map negotiated 0x0303/0x0304/other versions and http/1.1, h2, dot and other negotiated ALPN; absent fields when nothing reliable was negotiated; ignore record-layer/offered versions; never copy unregistered identifiers or let coarse markers override failures.

Scope: Synthetic core contract cases. No execution evidence or support claim. REQ-OBS-0049 also maps to `CC-OBS-VOCAB-TLS-OFFERED-ALPN` for offered-identifier admission.

<!-- requirement-ref: REQ-OBS-0048 -->
<!-- requirement-ref: REQ-OBS-0049 -->

- [REQ-OBS-0048](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0048) → `CC-OBS-VOCAB-TLS-NEGOTIATED-PROPERTIES`.
- [REQ-OBS-0049](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0049) → `CC-OBS-VOCAB-TLS-NEGOTIATED-PROPERTIES`.

### CC-OBS-VOCAB-TLS-OFFERED-ALPN

Public source: `data-formats/measurement-observations.md`, section **Registered Normalized Vocabularies / TLS Negotiated Properties**. Intended target: `internal/validation/options.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcObsVocabTlsOfferedAlpn`.

Case plan: Accept only offered http/1.1, h2 and dot; reject unknown, h2c, case variants and other registry entries alongside existing count, length and duplicate bounds.

Scope: Synthetic core contract cases. No execution evidence or support claim. This family covers the offered-identifier half of REQ-OBS-0049; negotiated output is covered by `CC-OBS-VOCAB-TLS-NEGOTIATED-PROPERTIES`.

<!-- requirement-ref: REQ-OBS-0049 -->

- [REQ-OBS-0049](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0049) → `CC-OBS-VOCAB-TLS-OFFERED-ALPN`.

### CC-OBS-VOCAB-TLS-PEER-ALERT

Public source: `data-formats/measurement-observations.md`, section **Registered Normalized Vocabularies / TLS Peer-Alert Normalization**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsVocabTlsPeerAlert`.

Case plan: Map every listed received peer description code and an unlisted code; omit alert when the description is unavailable; exclude local cancellation, deadline, validation failures, sent alerts and post-handshake closure; serialize no raw alert data.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0050 -->

- [REQ-OBS-0050](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0050) → `CC-OBS-VOCAB-TLS-PEER-ALERT`.

### CC-OBS-VOCAB-INCOMPLETE-OBSERVATIONS

Public source: `data-formats/measurement-observations.md`, section **Registered Normalized Vocabularies / Incomplete Observations**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsVocabIncompleteObservations`.

Case plan: For each observation type, omit failure on incomplete status and reject null, empty or registered failure values; keep termination in execution.reason.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0051 -->

- [REQ-OBS-0051](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0051) → `CC-OBS-VOCAB-INCOMPLETE-OBSERVATIONS`.

### CC-OBS-VALIDATION-AND-BOUNDS

Public source: `data-formats/measurement-observations.md`, section **Validation and Bounds**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsValidationAndBounds`.

Case plan: Parameterize each invariant, including one terminal failed/incomplete primitive observation, allowed types, bounds, status/detail consistency and inconclusive for unmapped raw errors.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0032 -->
<!-- requirement-ref: REQ-OBS-0034 -->
<!-- requirement-ref: REQ-OBS-0035 -->
<!-- requirement-ref: REQ-OBS-0036 -->
<!-- requirement-ref: REQ-OBS-0037 -->
<!-- requirement-ref: REQ-OBS-0038 -->
<!-- requirement-ref: REQ-OBS-0039 -->
<!-- requirement-ref: REQ-OBS-0040 -->
<!-- requirement-ref: REQ-OBS-0041 -->

- [REQ-OBS-0032](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0032) → `CC-OBS-VALIDATION-AND-BOUNDS`.
- [REQ-OBS-0034](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0034) → `CC-OBS-VALIDATION-AND-BOUNDS`.
- [REQ-OBS-0035](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0035) → `CC-OBS-VALIDATION-AND-BOUNDS`.
- [REQ-OBS-0036](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0036) → `CC-OBS-VALIDATION-AND-BOUNDS`.
- [REQ-OBS-0037](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0037) → `CC-OBS-VALIDATION-AND-BOUNDS`.
- [REQ-OBS-0038](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0038) → `CC-OBS-VALIDATION-AND-BOUNDS`.
- [REQ-OBS-0039](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0039) → `CC-OBS-VALIDATION-AND-BOUNDS`.
- [REQ-OBS-0040](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0040) → `CC-OBS-VALIDATION-AND-BOUNDS`.
- [REQ-OBS-0041](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0041) → `CC-OBS-VALIDATION-AND-BOUNDS`.

### CC-OBS-PRIVACY-AND-INTERPRETATION

Public source: `data-formats/measurement-observations.md`, section **Privacy and Interpretation**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsPrivacyAndInterpretation`.

Case plan: Exercise a scripted TLS-01 style attempt where the TCP connect succeeds and the TLS handshake then fails with injected failures (timeout, alert, certificate_untrusted). Assert the serialized observations are exactly a succeeded tcp_connect followed by one failed tls_handshake carrying the injected registered failure, that the result records the terminal outcome registered by the test rather than any additional field naming a cause, actor or intent, and that the output is byte-identical whether the scripted failure is labelled as a synthetic outage or a synthetic interference scenario. Requirements: REQ-OBS-0033.

Scope: Synthetic core contract cases. No execution evidence or support claim.

Review case `RV-CC-OBS-PRIVACY-AND-INTERPRETATION` (`docs/conformance/cross-cutting-review.md`, `heading:RV-CC-OBS-PRIVACY-AND-INTERPRETATION`): Human review case; not automated coverage. Reviewer role: Measurement interpretation reviewer (analyst role independent of the implementer). Reviewed artifact: Go result documentation, examples and downstream analyst guidance describing stage-ordered observations at the reviewed implementation revision. Checklist: (1) Does each example of a failed later stage after a successful earlier stage claim only that the earlier stage completed and the later stage did not complete normally under that test configuration? (2) Does no example or guidance attribute intent, an actor or definitive interference to an observation order or stage failure? (3) Does the guidance list non-censorship explanations (for example outages, congestion, routing, server or CDN failures, proxy behavior) for stage failures? (4) Does the guidance recommend repeated measurements or independent observations before stronger conclusions? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-OBS-0033 -->

- [REQ-OBS-0033](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-observations.md#REQ-OBS-0033) → `CC-OBS-PRIVACY-AND-INTERPRETATION`, `RV-CC-OBS-PRIVACY-AND-INTERPRETATION`.

### CC-RESULT-OVERVIEW

Public source: `data-formats/measurement-result.md`, section **Overview**. Intended target: `internal/result/result.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcResultOverview`.

Case plan: Feed negative DNS/TCP/TLS/HTTP terminal observations; produce completed measurements without automatic retries.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-RESULT-0001 -->

- [REQ-RESULT-0001](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0001) → `CC-RESULT-OVERVIEW`.

### CC-RESULT-TOP-LEVEL-FIELDS

Public source: `data-formats/measurement-result.md`, section **Top-level Fields**. Intended target: `internal/result/result.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcResultTopLevelFields`.

Case plan: Exercise required/conditional fields, encoded bytes/observation counts and task_id association. Exclude automatic task-field copies and reversible input digests.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-RESULT-0002 -->
<!-- requirement-ref: REQ-RESULT-0003 -->
<!-- requirement-ref: REQ-RESULT-0004 -->
<!-- requirement-ref: REQ-RESULT-0005 -->

- [REQ-RESULT-0002](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0002) → `CC-RESULT-TOP-LEVEL-FIELDS`.
- [REQ-RESULT-0003](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0003) → `CC-RESULT-TOP-LEVEL-FIELDS`.
- [REQ-RESULT-0004](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0004) → `CC-RESULT-TOP-LEVEL-FIELDS`.
- [REQ-RESULT-0005](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0005) → `CC-RESULT-TOP-LEVEL-FIELDS`.

### CC-RESULT-RUN-AND-ATTEMPT-IDENTITY

Public source: `data-formats/measurement-result.md`, section **Run and Attempt Identity**. Intended target: `internal/result/result.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcResultRunAndAttemptIdentity`.

Case plan: Inject a controlled random source; new runs get independent IDs, retries preserve ID/increment attempt, retransmission creates no attempt, and IDs never escape to measurement/context/log sinks.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-RESULT-0006 -->
<!-- requirement-ref: REQ-RESULT-0007 -->
<!-- requirement-ref: REQ-RESULT-0008 -->
<!-- requirement-ref: REQ-RESULT-0009 -->

- [REQ-RESULT-0006](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0006) → `CC-RESULT-RUN-AND-ATTEMPT-IDENTITY`.
- [REQ-RESULT-0007](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0007) → `CC-RESULT-RUN-AND-ATTEMPT-IDENTITY`.
- [REQ-RESULT-0008](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0008) → `CC-RESULT-RUN-AND-ATTEMPT-IDENTITY`.
- [REQ-RESULT-0009](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0009) → `CC-RESULT-RUN-AND-ATTEMPT-IDENTITY`.

### CC-RESULT-TEST-OBJECT

Public source: `data-formats/measurement-result.md`, section **Test Object**. Intended target: `internal/result/result.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcResultTestObject`.

Case plan: Select exact registered test/result-version tuples independently; reject unknown fields and unsupported combinations.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-RESULT-0010 -->
<!-- requirement-ref: REQ-RESULT-0011 -->
<!-- requirement-ref: REQ-RESULT-0012 -->

- [REQ-RESULT-0010](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0010) → `CC-RESULT-TEST-OBJECT`.
- [REQ-RESULT-0011](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0011) → `CC-RESULT-TEST-OBJECT`.
- [REQ-RESULT-0012](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0012) → `CC-RESULT-TEST-OBJECT`.

### CC-RESULT-IMPLEMENTATION-OBJECT

Public source: `data-formats/measurement-result.md`, section **Implementation Object**. Intended target: `internal/result/result.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcResultImplementationObject`.

Case plan: Validate bounded released version and platform enum; exclude host/device/build/locale metadata and arbitrary annotations.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-RESULT-0013 -->
<!-- requirement-ref: REQ-RESULT-0014 -->
<!-- requirement-ref: REQ-RESULT-0015 -->

- [REQ-RESULT-0013](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0013) → `CC-RESULT-IMPLEMENTATION-OBJECT`.
- [REQ-RESULT-0014](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0014) → `CC-RESULT-IMPLEMENTATION-OBJECT`.
- [REQ-RESULT-0015](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0015) → `CC-RESULT-IMPLEMENTATION-OBJECT`.

### CC-RESULT-TIME-OBJECT

Public source: `data-formats/measurement-result.md`, section **Time Object**. Intended target: `internal/result/result.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcResultTimeObject`.

Case plan: Use a monotonic fake clock plus wall-clock jumps; test minute floor, 250ms ceiling including zero, stage rounding and duration_capped conditional behavior.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-RESULT-0016 -->
<!-- requirement-ref: REQ-RESULT-0017 -->
<!-- requirement-ref: REQ-RESULT-0018 -->

- [REQ-RESULT-0016](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0016) → `CC-RESULT-TIME-OBJECT`.
- [REQ-RESULT-0017](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0017) → `CC-RESULT-TIME-OBJECT`.
- [REQ-RESULT-0018](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0018) → `CC-RESULT-TIME-OBJECT`.

### CC-RESULT-EXECUTION-OBJECT

Public source: `data-formats/measurement-result.md`, section **Execution Object**. Intended target: `internal/result/result.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcResultExecutionObject`.

Case plan: Exercise all states/reasons, first termination precedence, measurement presence/absence/partial rules and separation of malformed task errors from results. Interrupt an attempt before and after its test starts and assert not_run before it and incomplete after it (REQ-RESULT-0051), and interrupt it after the test records its terminal outcome and assert the result stays completed with an unchanged state, reason and measurement (REQ-RESULT-0052).

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-RESULT-0019 -->
<!-- requirement-ref: REQ-RESULT-0020 -->
<!-- requirement-ref: REQ-RESULT-0021 -->
<!-- requirement-ref: REQ-RESULT-0022 -->
<!-- requirement-ref: REQ-RESULT-0051 -->
<!-- requirement-ref: REQ-RESULT-0052 -->

- [REQ-RESULT-0019](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0019) → `CC-RESULT-EXECUTION-OBJECT`.
- [REQ-RESULT-0020](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0020) → `CC-RESULT-EXECUTION-OBJECT`.
- [REQ-RESULT-0021](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0021) → `CC-RESULT-EXECUTION-OBJECT`.
- [REQ-RESULT-0022](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0022) → `CC-RESULT-EXECUTION-OBJECT`.
- [REQ-RESULT-0051](https://github.com/tv-rain/barometer-spec/blob/84f52f8872a5af9e6d97894611e3b2eaba18aba8/data-formats/measurement-result.md#REQ-RESULT-0051) → `CC-RESULT-EXECUTION-OBJECT`.
- [REQ-RESULT-0052](https://github.com/tv-rain/barometer-spec/blob/84f52f8872a5af9e6d97894611e3b2eaba18aba8/data-formats/measurement-result.md#REQ-RESULT-0052) → `CC-RESULT-EXECUTION-OBJECT`.

### CC-RESULT-NETWORK-CONTEXT

Public source: `data-formats/measurement-result.md`, section **Network Context**. Intended target: `internal/result/result.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcResultNetworkContext`.

Case plan: Test completed/incomplete remote tests with unavailable egress and NET-01 exceptions; reject network_context in every not_run result, including one whose snapshot was captured before it stopped (REQ-RESULT-0053); reject prohibited context data.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-RESULT-0023 -->
<!-- requirement-ref: REQ-RESULT-0025 -->
<!-- requirement-ref: REQ-RESULT-0053 -->

- [REQ-RESULT-0023](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0023) → `CC-RESULT-NETWORK-CONTEXT`.
- [REQ-RESULT-0025](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0025) → `CC-RESULT-NETWORK-CONTEXT`.
- [REQ-RESULT-0053](https://github.com/tv-rain/barometer-spec/blob/84f52f8872a5af9e6d97894611e3b2eaba18aba8/data-formats/measurement-result.md#REQ-RESULT-0053) → `CC-RESULT-NETWORK-CONTEXT`. Successor of the retired REQ-RESULT-0024.

### CC-RESULT-MEASUREMENT-OBJECT

Public source: `data-formats/measurement-result.md`, section **Measurement Object**. Intended target: `internal/result/result.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcResultMeasurementObject`.

Case plan: Select the registered Output schema, ordered observation list and terminal outcome; reject unregistered extensions/verdicts.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-RESULT-0026 -->
<!-- requirement-ref: REQ-RESULT-0027 -->
<!-- requirement-ref: REQ-RESULT-0028 -->

- [REQ-RESULT-0026](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0026) → `CC-RESULT-MEASUREMENT-OBJECT`.
- [REQ-RESULT-0027](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0027) → `CC-RESULT-MEASUREMENT-OBJECT`.
- [REQ-RESULT-0028](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0028) → `CC-RESULT-MEASUREMENT-OBJECT`.

### CC-RESULT-PARTIAL-RESULTS

Public source: `data-formats/measurement-result.md`, section **Partial Results**. Intended target: `internal/result/result.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcResultPartialResults`.

Case plan: Interrupt each stage of NET-05 and primitive tests; retain only schema-permitted partial data and distinguish skipped from failed operations.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-RESULT-0029 -->
<!-- requirement-ref: REQ-RESULT-0030 -->

- [REQ-RESULT-0029](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0029) → `CC-RESULT-PARTIAL-RESULTS`.
- [REQ-RESULT-0030](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0030) → `CC-RESULT-PARTIAL-RESULTS`.

### CC-RESULT-VERSIONING-AND-EXTENSION

Public source: `data-formats/measurement-result.md`, section **Versioning and Extension**. Intended target: `internal/result/result.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcResultVersioningAndExtension`.

Case plan: Reject unknown known-version fields and unsupported exact versions; quarantine, if provided, cannot analyze a silently reduced object. Reject host/task injected extensions.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-RESULT-0031 -->
<!-- requirement-ref: REQ-RESULT-0032 -->
<!-- requirement-ref: REQ-RESULT-0033 -->

- [REQ-RESULT-0031](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0031) → `CC-RESULT-VERSIONING-AND-EXTENSION`.
- [REQ-RESULT-0032](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0032) → `CC-RESULT-VERSIONING-AND-EXTENSION`.
- [REQ-RESULT-0033](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0033) → `CC-RESULT-VERSIONING-AND-EXTENSION`.

### CC-RESULT-VALIDATION

Public source: `data-formats/measurement-result.md`, section **Validation**. Intended target: `internal/result/result.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcResultValidation`.

Case plan: Parameterize all common validity rules, including all-zero/incorrect run IDs, retry bound, timestamps, duration cap, terminal outcomes and prohibited data before bundling.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-RESULT-0034 -->
<!-- requirement-ref: REQ-RESULT-0036 -->
<!-- requirement-ref: REQ-RESULT-0037 -->
<!-- requirement-ref: REQ-RESULT-0038 -->
<!-- requirement-ref: REQ-RESULT-0039 -->
<!-- requirement-ref: REQ-RESULT-0040 -->
<!-- requirement-ref: REQ-RESULT-0041 -->
<!-- requirement-ref: REQ-RESULT-0042 -->
<!-- requirement-ref: REQ-RESULT-0043 -->
<!-- requirement-ref: REQ-RESULT-0044 -->

- [REQ-RESULT-0034](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0034) → `CC-RESULT-VALIDATION`.
- [REQ-RESULT-0036](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0036) → `CC-RESULT-VALIDATION`.
- [REQ-RESULT-0037](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0037) → `CC-RESULT-VALIDATION`.
- [REQ-RESULT-0038](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0038) → `CC-RESULT-VALIDATION`.
- [REQ-RESULT-0039](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0039) → `CC-RESULT-VALIDATION`.
- [REQ-RESULT-0040](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0040) → `CC-RESULT-VALIDATION`.
- [REQ-RESULT-0041](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0041) → `CC-RESULT-VALIDATION`.
- [REQ-RESULT-0042](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0042) → `CC-RESULT-VALIDATION`.
- [REQ-RESULT-0043](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0043) → `CC-RESULT-VALIDATION`.
- [REQ-RESULT-0044](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0044) → `CC-RESULT-VALIDATION`.

### CC-RESULT-PRIVACY-SAFETY-AND-INTERPRETATION

Public source: `data-formats/measurement-result.md`, section **Privacy, Safety, and Interpretation**. Intended target: `internal/result/result.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcResultPrivacySafetyAndInterpretation`.

Case plan: Inject distinct synthetic canaries through the fake platform, host configuration and scripted transports for every prohibited class: person, account, application-user, advertising, device, SIM, subscriber, push and stable installation identifiers; task input and label values; raw packets, HTTP bodies, HTTP header values, certificate chains, cookies, credentials, system log lines, stack traces (from an injected internal panic) and unrelated resolver state. Run completed, incomplete and not_run attempts and assert no canary, in raw or base64url/hex form, appears in the serialized Measurement Result. Assert the result for a completed negative outcome (for example a TCP timeout) contains only the registered outcome and observations, and reject decoding of a result carrying an unregistered top-level field holding a synthetic derived anomaly or interference analysis object. Requirements: REQ-RESULT-0035, REQ-RESULT-0045, REQ-RESULT-0046, REQ-RESULT-0047, REQ-RESULT-0048, REQ-RESULT-0049, REQ-RESULT-0050.

Scope: Synthetic core contract cases. No execution evidence or support claim.

Review case `RV-CC-RESULT-PRIVACY-SAFETY-AND-INTERPRETATION` (`docs/conformance/cross-cutting-review.md`, `heading:RV-CC-RESULT-PRIVACY-SAFETY-AND-INTERPRETATION`): Human review case; not automated coverage. Reviewer role: Privacy and measurement interpretation reviewer (independent of the implementer). Reviewed artifact: Go result documentation, host integration guidance and analyst guidance for Measurement Results at the reviewed implementation revision. Checklist: (1) Does the guidance state that coarse platform, time, Network Context, connection code, task and run metadata can become identifying in combination and require bounded retention and aggregate reporting? (2) Does the guidance state that a matching task or run identifier establishes only measurement workflow membership, not that results belong to one person? (3) Does the guidance describe a completed negative outcome as one observation on one path and context that does not prove censorship or identify intent? (4) Does the guidance require derived anomaly or interference analysis to remain separate from the source result, identify its method and version, and preserve links to the source observations? (5) Is any documented retention or export path for results bounded and free of per-person aggregation? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-RESULT-0035 -->
<!-- requirement-ref: REQ-RESULT-0045 -->
<!-- requirement-ref: REQ-RESULT-0046 -->
<!-- requirement-ref: REQ-RESULT-0047 -->
<!-- requirement-ref: REQ-RESULT-0048 -->
<!-- requirement-ref: REQ-RESULT-0049 -->
<!-- requirement-ref: REQ-RESULT-0050 -->

- [REQ-RESULT-0035](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0035) → `CC-RESULT-PRIVACY-SAFETY-AND-INTERPRETATION`, `RV-CC-RESULT-PRIVACY-SAFETY-AND-INTERPRETATION`.
- [REQ-RESULT-0045](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0045) → `CC-RESULT-PRIVACY-SAFETY-AND-INTERPRETATION`, `RV-CC-RESULT-PRIVACY-SAFETY-AND-INTERPRETATION`.
- [REQ-RESULT-0046](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0046) → `CC-RESULT-PRIVACY-SAFETY-AND-INTERPRETATION`, `RV-CC-RESULT-PRIVACY-SAFETY-AND-INTERPRETATION`.
- [REQ-RESULT-0047](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0047) → `CC-RESULT-PRIVACY-SAFETY-AND-INTERPRETATION`, `RV-CC-RESULT-PRIVACY-SAFETY-AND-INTERPRETATION`.
- [REQ-RESULT-0048](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0048) → `CC-RESULT-PRIVACY-SAFETY-AND-INTERPRETATION`, `RV-CC-RESULT-PRIVACY-SAFETY-AND-INTERPRETATION`.
- [REQ-RESULT-0049](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0049) → `CC-RESULT-PRIVACY-SAFETY-AND-INTERPRETATION`, `RV-CC-RESULT-PRIVACY-SAFETY-AND-INTERPRETATION`.
- [REQ-RESULT-0050](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/measurement-result.md#REQ-RESULT-0050) → `CC-RESULT-PRIVACY-SAFETY-AND-INTERPRETATION`, `RV-CC-RESULT-PRIVACY-SAFETY-AND-INTERPRETATION`.

### CC-CONTEXT-FORMAT-OVERVIEW

Public source: `data-formats/network-context.md`, section **Overview**. Intended target: `internal/networkcontext/model.go`. Planned case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextFormatOverview`.

Case plan: Exercise context capture with a fake platform that reports access-network generation G1, then switches to generation G2 during the attempt, and a fake egress endpoint, for both direct and proxy selected paths. Assert the result carries exactly one Network Context with one access object taken from the attempt's snapshot (network_changed true after the injected switch) and an egress object whose network_path equals the selected path, with no second access snapshot or path. Inject canaries for the public egress address (203.0.113.7, 2001:db8::7), local addresses (192.168.0.2, fe80::2) and persistent network identifiers (a synthetic platform network handle and generation token) and assert none appears in the serialized context. Reject decoding of a context carrying any additional address or identifier field. Requirements: REQ-CONTEXT-FORMAT-0001.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-CONTEXT-FORMAT-0001 -->

- [REQ-CONTEXT-FORMAT-0001](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0001) → `CC-CONTEXT-FORMAT-OVERVIEW`.

### CC-CONTEXT-FORMAT-TOP-LEVEL-FIELDS

Public source: `data-formats/network-context.md`, section **Top-level Fields**. Intended target: `internal/networkcontext/model.go`. Planned case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextFormatTopLevelFields`.

Case plan: Iterate synthetic Network Context JSON under the strict decoding contract: accept an object with valid access and egress objects, including egress status unavailable with lookup_source none; reject objects with access missing, egress missing, either set to null or to a non-object, and objects with one additional unknown top-level field. Assert every encoded context produced by the library, including when egress lookup fails, contains both access and egress and no other top-level key. Requirements: REQ-CONTEXT-FORMAT-0002, REQ-CONTEXT-FORMAT-0003.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-CONTEXT-FORMAT-0002 -->
<!-- requirement-ref: REQ-CONTEXT-FORMAT-0003 -->

- [REQ-CONTEXT-FORMAT-0002](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0002) → `CC-CONTEXT-FORMAT-TOP-LEVEL-FIELDS`.
- [REQ-CONTEXT-FORMAT-0003](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0003) → `CC-CONTEXT-FORMAT-TOP-LEVEL-FIELDS`.

### CC-CONTEXT-FORMAT-ACCESS-OBJECT

Public source: `data-formats/network-context.md`, section **Access Object**. Intended target: `internal/networkcontext/model.go`. Case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextFormatAccessObject`.

Case plan: Iterate every registered access value and accept each: network_type wifi, cellular, ethernet, other, unknown; vpn_state detected, not_detected, unknown; ipv4_available, ipv6_available and validated yes, no, unknown; network_changed true and false. Reject case variants (WIFI, Yes), unregistered strings, network_changed as the string false or as null, and objects missing each required field in turn. Configure the fake platform to expose canaries for interface names, addresses, Wi-Fi identifiers, DNS server addresses, cell information, SIM information, carrier names, subscriber identifiers and location, and assert none appears in the encoded access object or result; reject decoding of an access object carrying any extra field for those items. Requirements: REQ-CONTEXT-FORMAT-0004, REQ-CONTEXT-FORMAT-0005.

Scope: Synthetic core contract cases. Passed at Go `cdbd160` ([evidence](../traceability/evidence/go-unit-cdbd160.md#CC-CONTEXT-FORMAT-ACCESS-OBJECT)); partial evidence only, with no native capability or support claim.

<!-- requirement-ref: REQ-CONTEXT-FORMAT-0004 -->
<!-- requirement-ref: REQ-CONTEXT-FORMAT-0005 -->

- [REQ-CONTEXT-FORMAT-0004](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0004) → `CC-CONTEXT-FORMAT-ACCESS-OBJECT`.
- [REQ-CONTEXT-FORMAT-0005](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0005) → `CC-CONTEXT-FORMAT-ACCESS-OBJECT`.

### CC-CONTEXT-FORMAT-EGRESS-OBJECT

Public source: `data-formats/network-context.md`, section **Egress Object**. Intended target: `internal/networkcontext/model.go`. Case location: `internal/networkcontext/egress_test.go`, `symbol:TestCcContextFormatEgressObject`.

Case plan: Test all status/source/path/connection-code/classification combinations, missing optional classification, and no guessed values.

Scope: Synthetic core contract cases. Executed by DEV-647 with a scripted fetch: unavailable, fresh primary, fresh fallback and cached egress on the direct and proxy paths, `connection_code` only on the proxy path, and missing classifications absent rather than guessed. Passed at Go `d788a80` ([evidence](../traceability/evidence/go-unit-d788a80.md#CC-CONTEXT-FORMAT-EGRESS-OBJECT)); partial evidence only. No native, field or support claim.

<!-- requirement-ref: REQ-CONTEXT-FORMAT-0006 -->
<!-- requirement-ref: REQ-CONTEXT-FORMAT-0007 -->
<!-- requirement-ref: REQ-CONTEXT-FORMAT-0008 -->
<!-- requirement-ref: REQ-CONTEXT-FORMAT-0009 -->

- [REQ-CONTEXT-FORMAT-0006](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0006) → `CC-CONTEXT-FORMAT-EGRESS-OBJECT`.
- [REQ-CONTEXT-FORMAT-0007](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0007) → `CC-CONTEXT-FORMAT-EGRESS-OBJECT`.
- [REQ-CONTEXT-FORMAT-0008](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0008) → `CC-CONTEXT-FORMAT-EGRESS-OBJECT`.
- [REQ-CONTEXT-FORMAT-0009](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0009) → `CC-CONTEXT-FORMAT-EGRESS-OBJECT`.

### CC-CONTEXT-FORMAT-PROVIDER-AND-OPERATOR

Public source: `data-formats/network-context.md`, section **Provider and Operator**. Intended target: `internal/networkcontext/model.go`. Planned case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextFormatProviderAndOperator`.

Case plan: Configure the fake egress endpoint and fake platform to return a synthetic provider organization name and a synthetic mobile operator/carrier name as canaries alongside a valid asn. Assert the encoded Network Context contains only the registered numeric asn and no provider or operator text in any field, and reject decoding of a context carrying a provider or operator field at the top level, in access or in egress. Requirements: REQ-CONTEXT-FORMAT-0010, REQ-CONTEXT-FORMAT-0011, REQ-CONTEXT-FORMAT-0012.

Scope: Synthetic core contract cases. No execution evidence or support claim.

Review case `RV-CC-CONTEXT-FORMAT-PROVIDER-AND-OPERATOR` (`docs/conformance/cross-cutting-review.md`, `heading:RV-CC-CONTEXT-FORMAT-PROVIDER-AND-OPERATOR`): Human review case; not automated coverage. Reviewer role: Privacy and measurement interpretation reviewer (independent of the implementer). Reviewed artifact: Go Network Context documentation, analyst guidance and any proposed format extension touching provider or operator data at the reviewed implementation revision. Checklist: (1) Does any guidance on deriving a provider organization from asn name a documented mapping dataset? (2) Does the guidance present a derived provider label as contextual rather than a verified access-provider identity? (3) Is mobile operator information absent from the baseline implementation and documentation? (4) If an operator extension is proposed, does it record separate consent, platform, privacy, safety and legal review before adoption? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-CONTEXT-FORMAT-0010 -->
<!-- requirement-ref: REQ-CONTEXT-FORMAT-0011 -->
<!-- requirement-ref: REQ-CONTEXT-FORMAT-0012 -->

- [REQ-CONTEXT-FORMAT-0010](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0010) → `CC-CONTEXT-FORMAT-PROVIDER-AND-OPERATOR`, `RV-CC-CONTEXT-FORMAT-PROVIDER-AND-OPERATOR`.
- [REQ-CONTEXT-FORMAT-0011](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0011) → `CC-CONTEXT-FORMAT-PROVIDER-AND-OPERATOR`, `RV-CC-CONTEXT-FORMAT-PROVIDER-AND-OPERATOR`.
- [REQ-CONTEXT-FORMAT-0012](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0012) → `CC-CONTEXT-FORMAT-PROVIDER-AND-OPERATOR`, `RV-CC-CONTEXT-FORMAT-PROVIDER-AND-OPERATOR`.

### CC-CONTEXT-FORMAT-USE-IN-EXECUTION

Public source: `data-formats/network-context.md`, section **Use in Execution**. Intended target: `internal/networkcontext/model.go`. Case location: `internal/execution/egress_test.go`, `symbol:TestCcExecutionFieldsNetworkFieldsNetworkContextRequirements`.

Case plan: Use normalized required fields and country eligibility only; reject region/ASN/provider/operator task allowlists.

Scope: Synthetic core contract cases. Executed by DEV-647: normalized required fields and a country allowlist are decoded and evaluated, and region, ASN, provider and operator allowlists are rejected by the task schema in `TestCcExecutionFieldsNetworkFieldsNetworkContextRequirements`; the same rejections at admission are the separate unit-test case `UT.internal.admission.TestGateAdmitsEgressContextRequirements`. Passed at Go `d788a80` ([evidence](../traceability/evidence/go-unit-d788a80.md#CC-CONTEXT-FORMAT-USE-IN-EXECUTION)); partial evidence only. No native, field or support claim.

<!-- requirement-ref: REQ-CONTEXT-FORMAT-0013 -->
<!-- requirement-ref: REQ-CONTEXT-FORMAT-0014 -->

- [REQ-CONTEXT-FORMAT-0013](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0013) → `CC-CONTEXT-FORMAT-USE-IN-EXECUTION`.
- [REQ-CONTEXT-FORMAT-0014](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0014) → `CC-CONTEXT-FORMAT-USE-IN-EXECUTION`.

### CC-CONTEXT-FORMAT-USE-IN-RESULTS

Public source: `data-formats/network-context.md`, section **Use in Results**. Intended target: `internal/networkcontext/model.go`. Case location: `internal/execution/egress_test.go`, `symbol:TestCcContextContextLayersEgressContext`.

Case plan: Test schema-permitted context carriage/deduplication and reject persistent cross-batch identity or listed private details.

Scope: Synthetic core contract cases. Executed by DEV-647: each result carries its own context with no deduplication or cross-batch identifier, and no result contains an endpoint name or URL, the observed or local address, the proxy endpoint or a generation. Passed at Go `d788a80` ([evidence](../traceability/evidence/go-unit-d788a80.md#CC-CONTEXT-FORMAT-USE-IN-RESULTS)); partial evidence only. No native, field or support claim.

<!-- requirement-ref: REQ-CONTEXT-FORMAT-0015 -->
<!-- requirement-ref: REQ-CONTEXT-FORMAT-0016 -->

- [REQ-CONTEXT-FORMAT-0015](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0015) → `CC-CONTEXT-FORMAT-USE-IN-RESULTS`.
- [REQ-CONTEXT-FORMAT-0016](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0016) → `CC-CONTEXT-FORMAT-USE-IN-RESULTS`.

### CC-CONTEXT-FORMAT-VALIDATION

Public source: `data-formats/network-context.md`, section **Validation**. Intended target: `internal/networkcontext/model.go`. Planned case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextFormatValidation`.

Case plan: Mutate each field/conditional relationship; invalid context satisfies no execution requirement and is never attached as valid.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-CONTEXT-FORMAT-0017 -->
<!-- requirement-ref: REQ-CONTEXT-FORMAT-0018 -->
<!-- requirement-ref: REQ-CONTEXT-FORMAT-0019 -->
<!-- requirement-ref: REQ-CONTEXT-FORMAT-0020 -->
<!-- requirement-ref: REQ-CONTEXT-FORMAT-0021 -->
<!-- requirement-ref: REQ-CONTEXT-FORMAT-0022 -->
<!-- requirement-ref: REQ-CONTEXT-FORMAT-0023 -->

- [REQ-CONTEXT-FORMAT-0017](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0017) → `CC-CONTEXT-FORMAT-VALIDATION`.
- [REQ-CONTEXT-FORMAT-0018](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0018) → `CC-CONTEXT-FORMAT-VALIDATION`.
- [REQ-CONTEXT-FORMAT-0019](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0019) → `CC-CONTEXT-FORMAT-VALIDATION`.
- [REQ-CONTEXT-FORMAT-0020](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0020) → `CC-CONTEXT-FORMAT-VALIDATION`.
- [REQ-CONTEXT-FORMAT-0021](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0021) → `CC-CONTEXT-FORMAT-VALIDATION`.
- [REQ-CONTEXT-FORMAT-0022](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0022) → `CC-CONTEXT-FORMAT-VALIDATION`.
- [REQ-CONTEXT-FORMAT-0023](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0023) → `CC-CONTEXT-FORMAT-VALIDATION`.

### CC-BUNDLE-OVERVIEW

Public source: `data-formats/result-bundle.md`, section **Overview**. Intended target: `internal/bundle/bundle.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcBundleOverview`.

Case plan: Group valid individual results with a single-use bundle ID; rotating correlation is separate and default-off.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-BUNDLE-0001 -->
<!-- requirement-ref: REQ-BUNDLE-0002 -->

- [REQ-BUNDLE-0001](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0001) → `CC-BUNDLE-OVERVIEW`.
- [REQ-BUNDLE-0002](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0002) → `CC-BUNDLE-OVERVIEW`.

### CC-BUNDLE-JSON-STRUCTURE

Public source: `data-formats/result-bundle.md`, section **JSON Structure**. Intended target: `internal/bundle/bundle.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcBundleJsonStructure`.

Case plan: Build bundles with linkage disabled (default configuration and an explicit disable) from a synthetic not_run TCP-01 result with reason constraints_unsatisfied and from a completed result. Assert the encoded bundle has exactly the top-level keys schema_version, bundle_id and results, with the linkage key absent (not null and not an empty object), and that the public disabled-linkage example decodes successfully. Reject decoding of a disabled-linkage bundle that carries linkage as null or an empty object, per the strict decoding contract. Requirements: REQ-BUNDLE-0003.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-BUNDLE-0003 -->

- [REQ-BUNDLE-0003](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0003) → `CC-BUNDLE-JSON-STRUCTURE`.

### CC-BUNDLE-TOP-LEVEL-FIELDS

Public source: `data-formats/result-bundle.md`, section **Top-level Fields**. Intended target: `internal/bundle/bundle.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcBundleTopLevelFields`.

Case plan: Exercise envelope required fields/types, 1..16 results and encoded-byte bound; reject unknown fields.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-BUNDLE-0004 -->
<!-- requirement-ref: REQ-BUNDLE-0005 -->

- [REQ-BUNDLE-0004](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0004) → `CC-BUNDLE-TOP-LEVEL-FIELDS`.
- [REQ-BUNDLE-0005](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0005) → `CC-BUNDLE-TOP-LEVEL-FIELDS`.

### CC-BUNDLE-BUNDLE-IDENTIFIER

Public source: `data-formats/result-bundle.md`, section **Bundle Identifier**. Intended target: `internal/bundle/bundle.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcBundleBundleIdentifier`.

Case plan: Inject random IDs; exact retries preserve identity/content, any membership/content mutation gets a new ID, acknowledgement/expiry deletes state, and no ID enters measurement traffic or targeting.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-BUNDLE-0006 -->
<!-- requirement-ref: REQ-BUNDLE-0007 -->
<!-- requirement-ref: REQ-BUNDLE-0008 -->
<!-- requirement-ref: REQ-BUNDLE-0009 -->

- [REQ-BUNDLE-0006](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0006) → `CC-BUNDLE-BUNDLE-IDENTIFIER`.
- [REQ-BUNDLE-0007](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0007) → `CC-BUNDLE-BUNDLE-IDENTIFIER`.
- [REQ-BUNDLE-0008](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0008) → `CC-BUNDLE-BUNDLE-IDENTIFIER`.
- [REQ-BUNDLE-0009](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0009) → `CC-BUNDLE-BUNDLE-IDENTIFIER`.

### CC-BUNDLE-LINKAGE-OBJECT

Public source: `data-formats/result-bundle.md`, section **Linkage Object**. Intended target: `internal/bundle/bundle.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcBundleLinkageObject`.

Case plan: Validate the exact complete rotating-installation object and local scope, encoded token/epoch, no stable mode or user_id representation.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-BUNDLE-0010 -->
<!-- requirement-ref: REQ-BUNDLE-0011 -->
<!-- requirement-ref: REQ-BUNDLE-0012 -->

- [REQ-BUNDLE-0010](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0010) → `CC-BUNDLE-LINKAGE-OBJECT`.
- [REQ-BUNDLE-0011](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0011) → `CC-BUNDLE-LINKAGE-OBJECT`.
- [REQ-BUNDLE-0012](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0012) → `CC-BUNDLE-LINKAGE-OBJECT`.

### CC-BUNDLE-RESULT-MEMBERSHIP

Public source: `data-formats/result-bundle.md`, section **Result Membership**. Intended target: `internal/bundle/bundle.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcBundleResultMembership`.

Case plan: Bundle mixed completed/negative/not_run/incomplete attempts without changing semantics or adding unregistered context references.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-BUNDLE-0013 -->
<!-- requirement-ref: REQ-BUNDLE-0014 -->

- [REQ-BUNDLE-0013](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0013) → `CC-BUNDLE-RESULT-MEMBERSHIP`.
- [REQ-BUNDLE-0014](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0014) → `CC-BUNDLE-RESULT-MEMBERSHIP`.

### CC-BUNDLE-CONSTRUCTION-AND-DELIVERY

Public source: `data-formats/result-bundle.md`, section **Construction and Delivery**. Intended target: `internal/bundle/bundle.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcBundleConstructionAndDelivery`.

Case plan: Fault-inject validation, persistence, ambiguous delivery and durable acknowledgement; replay identical logical bundles and delete by acknowledgement/expiry, never mutating linkage on retry.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-BUNDLE-0015 -->
<!-- requirement-ref: REQ-BUNDLE-0016 -->

- [REQ-BUNDLE-0015](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0015) → `CC-BUNDLE-CONSTRUCTION-AND-DELIVERY`.
- [REQ-BUNDLE-0016](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0016) → `CC-BUNDLE-CONSTRUCTION-AND-DELIVERY`.

### CC-BUNDLE-VALIDATION

Public source: `data-formats/result-bundle.md`, section **Validation**. Intended target: `internal/bundle/bundle.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcBundleValidation`.

Case plan: Parameterize ID/all-zero/encoding/count/size/result-schema/linkage/state/pending-storage invariants before submission.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-BUNDLE-0017 -->
<!-- requirement-ref: REQ-BUNDLE-0021 -->
<!-- requirement-ref: REQ-BUNDLE-0022 -->
<!-- requirement-ref: REQ-BUNDLE-0023 -->
<!-- requirement-ref: REQ-BUNDLE-0024 -->
<!-- requirement-ref: REQ-BUNDLE-0025 -->
<!-- requirement-ref: REQ-BUNDLE-0026 -->
<!-- requirement-ref: REQ-BUNDLE-0027 -->
<!-- requirement-ref: REQ-BUNDLE-0028 -->

- [REQ-BUNDLE-0017](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0017) → `CC-BUNDLE-VALIDATION`.
- [REQ-BUNDLE-0021](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0021) → `CC-BUNDLE-VALIDATION`.
- [REQ-BUNDLE-0022](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0022) → `CC-BUNDLE-VALIDATION`.
- [REQ-BUNDLE-0023](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0023) → `CC-BUNDLE-VALIDATION`.
- [REQ-BUNDLE-0024](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0024) → `CC-BUNDLE-VALIDATION`.
- [REQ-BUNDLE-0025](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0025) → `CC-BUNDLE-VALIDATION`.
- [REQ-BUNDLE-0026](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0026) → `CC-BUNDLE-VALIDATION`.
- [REQ-BUNDLE-0027](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0027) → `CC-BUNDLE-VALIDATION`.
- [REQ-BUNDLE-0028](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0028) → `CC-BUNDLE-VALIDATION`.

### CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS

Public source: `data-formats/result-bundle.md`, section **Privacy and Safety Requirements**. Intended target: `internal/bundle/bundle.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcBundlePrivacyAndSafetyRequirements`.

Case plan: Assert a bundle built under the default configuration omits linkage. Inject canaries through the fake platform and host for advertising IDs, vendor IDs, device serials, hardware IDs, push tokens, account IDs, email addresses, phone numbers, SIM and subscriber identifiers and host application user IDs, plus hashed forms of each, and assert none appears in the encoded bundle. With linkage enabled, record all measurement traffic (DNS queries, TCP and TLS payloads, HTTP request lines, headers and bodies, HTTP CONNECT requests) and all Network Context endpoint requests, and assert neither bundle_id nor linkage.token appears in raw or base64url form. Run the same task set with linkage disabled, enabled with token T1 and enabled with token T2 and assert the delivered and executed task set is identical. Derive tokens from two different synthetic random installation secrets with identical host and platform identifiers and assert the tokens differ, and from one secret with different host and platform identifiers and assert the token is unchanged. Requirements: REQ-BUNDLE-0018, REQ-BUNDLE-0029, REQ-BUNDLE-0030, REQ-BUNDLE-0031, REQ-BUNDLE-0032, REQ-BUNDLE-0033, REQ-BUNDLE-0034, REQ-BUNDLE-0035, REQ-BUNDLE-0036.

Scope: Synthetic core contract cases. No execution evidence or support claim.

Review case `RV-CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS` (`docs/conformance/cross-cutting-review.md`, `heading:RV-CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`): Human review case; not automated coverage. Reviewer role: Privacy and collector-operations reviewer (independent of the implementer). Reviewed artifact: Go bundle and linkage documentation plus host and collector integration guidance at the reviewed implementation revision. Checklist: (1) Does the collector guidance require linkage tokens to be treated as sensitive pseudonymous data rather than anonymous data? (2) Does the guidance advise against retaining source IP address, precise arrival time or transport logs together with a linkage token, and recommend separating transport handling from research storage where feasible? (3) Does the guidance require raw-token access to be restricted, audited and limited in retention? (4) Does the analysis and publication guidance require aggregation of linkage, time, region, ASN, network type and connection_code combinations? (5) Does the documentation state that hashing a host or platform identifier does not make it acceptable as a token source? (6) Does no documented host or collector path use linkage for task selection or delivery? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-BUNDLE-0018 -->
<!-- requirement-ref: REQ-BUNDLE-0029 -->
<!-- requirement-ref: REQ-BUNDLE-0030 -->
<!-- requirement-ref: REQ-BUNDLE-0031 -->
<!-- requirement-ref: REQ-BUNDLE-0032 -->
<!-- requirement-ref: REQ-BUNDLE-0033 -->
<!-- requirement-ref: REQ-BUNDLE-0034 -->
<!-- requirement-ref: REQ-BUNDLE-0035 -->
<!-- requirement-ref: REQ-BUNDLE-0036 -->

- [REQ-BUNDLE-0018](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0018) → `CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-BUNDLE-0029](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0029) → `CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-BUNDLE-0030](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0030) → `CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-BUNDLE-0031](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0031) → `CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-BUNDLE-0032](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0032) → `CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-BUNDLE-0033](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0033) → `CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-BUNDLE-0034](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0034) → `CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-BUNDLE-0035](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0035) → `CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-BUNDLE-0036](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0036) → `CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`.

### CC-BUNDLE-INTERPRETATION

Public source: `data-formats/result-bundle.md`, section **Interpretation**. Intended target: `internal/bundle/bundle.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcBundleInterpretation`.

Case plan: Exercise bundle construction with a fake clock, fake secure storage and a synthetic installation secret: two bundles with the same secret, scope and epoch carry matching tokens; then trigger each locally reproducible cause in turn (linkage disabled, epoch rotation, explicit privacy reset, application data clearing via empty storage, key loss and storage failure via fake storage errors) and assert the token is absent or changed while the bundle schema is otherwise unchanged, with no added field or counter marking a new installation, person or account. Requirements: REQ-BUNDLE-0019, REQ-BUNDLE-0020.

Scope: Synthetic core contract cases. No execution evidence or support claim.

Review case `RV-CC-BUNDLE-INTERPRETATION` (`docs/conformance/cross-cutting-review.md`, `heading:RV-CC-BUNDLE-INTERPRETATION`): Human review case; not automated coverage. Reviewer role: Measurement interpretation reviewer (analyst role independent of the implementer). Reviewed artifact: Go linkage documentation and analyst guidance with correlation examples at the reviewed implementation revision. Checklist: (1) Do the examples treat matching tokens only as the same locally stored installation secret, deployment scope and rotation epoch? (2) Does no example claim that matching tokens identify the same person, device owner, account, physical location or access network? (3) Does the guidance list disabled linkage, epoch rotation, data clearing, reinstall detection, key loss, storage failure, device migration and privacy reset as causes of a missing or changed token? (4) Does no example count a missing or changed token as a new person? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-BUNDLE-0019 -->
<!-- requirement-ref: REQ-BUNDLE-0020 -->

- [REQ-BUNDLE-0019](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0019) → `CC-BUNDLE-INTERPRETATION`, `RV-CC-BUNDLE-INTERPRETATION`.
- [REQ-BUNDLE-0020](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/result-bundle.md#REQ-BUNDLE-0020) → `CC-BUNDLE-INTERPRETATION`, `RV-CC-BUNDLE-INTERPRETATION`.

### CC-EXECUTION-ENVELOPE

Public source: `data-formats/task-execution.md`, section **Envelope**. Intended target: `internal/execution/attempt.go`. Case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcExecutionEnvelope`.

Case plan: Recheck constraints immediately before attempts and preserve selected registered protocol semantics through direct/proxy transport.

Scope: Synthetic core contract cases. Runs in `TestCcExecutionEnvelope` with synthetic runners, the production CONNECT dialer and the scripted loopback proxy of `internal/testkit/loopback`. Passed at Go `d788a80` ([evidence](../traceability/evidence/go-unit-d788a80.md#CC-EXECUTION-ENVELOPE)); partial evidence only. No native capability, production runner or support claim.

<!-- requirement-ref: REQ-EXECUTION-0001 -->
<!-- requirement-ref: REQ-EXECUTION-0002 -->

- [REQ-EXECUTION-0001](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0001) → `CC-EXECUTION-ENVELOPE`.
- [REQ-EXECUTION-0002](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0002) → `CC-EXECUTION-ENVELOPE`.

### CC-EXECUTION-EXECUTION-MODEL

Public source: `data-formats/task-execution.md`, section **Execution Model**. Intended target: `internal/execution/attempt.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcExecutionExecutionModel`.

Case plan: Use a deterministic clock and a synthetic task with retry.max_attempts 2 and retry_on platform_suspended, with a fake constraint evaluator whose result is scripted per call. Assert the evaluator is invoked immediately before each execution attempt, including the retry attempt: script satisfied for attempt 1 (which ends with platform_suspended) and unsatisfied for attempt 2, and assert attempt 2 is delayed or skipped according to the schedule while attempt 1's result is unchanged. Script unsatisfied for the first attempt of a run and assert the run is delayed or skipped, never executed. Assert across all scenarios that the count of runs equals the count of schedule-due runs: constraint evaluation, delay or skip never increments the run counter or creates a new run identity. Requirements: REQ-EXECUTION-0003, REQ-EXECUTION-0004.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-EXECUTION-0003 -->
<!-- requirement-ref: REQ-EXECUTION-0004 -->

- [REQ-EXECUTION-0003](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0003) → `CC-EXECUTION-EXECUTION-MODEL`.
- [REQ-EXECUTION-0004](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0004) → `CC-EXECUTION-EXECUTION-MODEL`.

### CC-EXECUTION-FIELDS

Public source: `data-formats/task-execution.md`, section **Fields**. Intended target: `internal/execution/attempt.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcExecutionFields`.

Case plan: Decode synthetic execution objects through the strict task decoder. Reject an execution object without limits. Accept limits as an empty object (registered defaults applied: max_duration_ms 30000, max_bytes_sent 32768, max_bytes_received 131072, max_connections 8) and assert omitted network_path decodes as direct. Accept network_path direct and proxy and reject any other value. Accept network, device and concurrency when absent or present as typed objects of their defined schemas, and reject each when supplied as a non-object (string, array, null) or with an unknown field. Reject an unknown top-level execution field. Requirements: REQ-EXECUTION-0005.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-EXECUTION-0005 -->

- [REQ-EXECUTION-0005](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0005) → `CC-EXECUTION-FIELDS`.

### CC-EXECUTION-FIELDS-NETWORK-PATH

Public source: `data-formats/task-execution.md`, section **Fields / Network Path**. Intended target: `internal/execution/attempt.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcExecutionFieldsNetworkPath`.

Case plan: Test direct default and proxy selection; deny each proxy precondition, snapshot one generation across NET-05, and inject each path failure while asserting zero direct fallback.

Its selection, precondition and path-failure rows run in `TestCcExecutionFieldsNetworkPath` with the production CONNECT dialer against the scripted loopback proxy; its NET-05 row, one path snapshot across every profile step, needs the NET-05 runner (DEV-454), so the family stays planned.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-EXECUTION-0006 -->
<!-- requirement-ref: REQ-EXECUTION-0007 -->
<!-- requirement-ref: REQ-EXECUTION-0008 -->
<!-- requirement-ref: REQ-EXECUTION-0009 -->
<!-- requirement-ref: REQ-EXECUTION-0010 -->

- [REQ-EXECUTION-0006](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0006) → `CC-EXECUTION-FIELDS-NETWORK-PATH`.
- [REQ-EXECUTION-0007](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0007) → `CC-EXECUTION-FIELDS-NETWORK-PATH`.
- [REQ-EXECUTION-0008](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0008) → `CC-EXECUTION-FIELDS-NETWORK-PATH`.
- [REQ-EXECUTION-0009](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0009) → `CC-EXECUTION-FIELDS-NETWORK-PATH`.
- [REQ-EXECUTION-0010](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0010) → `CC-EXECUTION-FIELDS-NETWORK-PATH`.

### CC-EXECUTION-FIELDS-LIMITS

Public source: `data-formats/task-execution.md`, section **Fields / Limits**. Intended target: `internal/execution/attempt.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcExecutionFieldsLimits`.

Case plan: Exercise an accepted task with limits max_duration_ms 30000, max_bytes_sent 32768, max_bytes_received 131072 and max_connections 8 on a deterministic clock and scripted transport, with a retry bound permitting a second attempt of the same run. In attempt 1 consume exactly 8 connections, 32,768 application bytes sent and 131,072 received and terminate it with a cross-cutting reason that the public retry rules make retryable. Assert attempt 2 (same run_id, attempt 2) starts with full budgets: it can open 8 connections and send and receive the full byte limits, a 9th connection or one byte beyond either byte limit stops it with resource_limit, and its duration deadline is measured from its own start rather than from attempt 1. Requirements: REQ-EXECUTION-0011.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-EXECUTION-0011 -->

- [REQ-EXECUTION-0011](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0011) → `CC-EXECUTION-FIELDS-LIMITS`.

### CC-EXECUTION-FIELDS-LIMITS-FIELDS

Public source: `data-formats/task-execution.md`, section **Fields / Limits / Fields**. Intended target: `internal/execution/attempt.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcExecutionFieldsLimitsFields`.

Case plan: Exercise omitted defaults, lower positive requests and above-local/registered rejection; count bytes at the transport boundary, including HTTP headers, DNS messages and length prefixes, stop reads and writes promptly at the counted limits and document buffering below that boundary.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-EXECUTION-0012 -->
<!-- requirement-ref: REQ-EXECUTION-0013 -->
<!-- requirement-ref: REQ-EXECUTION-0032 -->

- [REQ-EXECUTION-0012](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0012) → `CC-EXECUTION-FIELDS-LIMITS-FIELDS`.
- [REQ-EXECUTION-0013](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0013) → `CC-EXECUTION-FIELDS-LIMITS-FIELDS`.
- [REQ-EXECUTION-0032](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/data-formats/task-execution.md#REQ-EXECUTION-0032) → `CC-EXECUTION-FIELDS-LIMITS-FIELDS`. Successor of the retired REQ-EXECUTION-0014.

### CC-EXECUTION-FIELDS-NETWORK

Public source: `data-formats/task-execution.md`, section **Fields / Network**. Intended target: `internal/execution/attempt.go`. Case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcExecutionFieldsNetwork`.

Case plan: Configure the fake platform so required network information is unavailable: network_type unknown with allowed_transports [wifi, cellular], vpn_state unknown under vpn_policy exclude and require, and context required_fields [country_code] with the egress classification unavailable. Assert for each case that no network operation is issued on the recording transport and the attempt does not run (not_run with constraints_unsatisfied for each undeterminable network condition, network_context_unavailable only for the missing required context field), and that unknown is never treated as satisfying a constraint. Assert the same task runs when the fake platform reports the required values. Requirements: REQ-EXECUTION-0031.

DEV-644 and DEV-647: `TestCcExecutionFieldsNetwork` runs the network-type, VPN-state and validation rows with the test-only fake platform network, and a required egress field that no lookup can supply records `network_context_unavailable` with no request or test run.

Scope: Synthetic core contract cases. Executed with the fake platform network. Passed at Go `d788a80` ([evidence](../traceability/evidence/go-unit-d788a80.md#CC-EXECUTION-FIELDS-NETWORK)); partial evidence only. No native, field or support claim.

<!-- requirement-ref: REQ-EXECUTION-0031 -->

- [REQ-EXECUTION-0031](https://github.com/tv-rain/barometer-spec/blob/84f52f8872a5af9e6d97894611e3b2eaba18aba8/data-formats/task-execution.md#REQ-EXECUTION-0031) → `CC-EXECUTION-FIELDS-NETWORK`. Successor of the retired REQ-EXECUTION-0015.

### CC-EXECUTION-FIELDS-NETWORK-FIELDS

Public source: `data-formats/task-execution.md`, section **Fields / Network / Fields**. Intended target: `internal/execution/attempt.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcExecutionFieldsNetworkFields`.

Case plan: Iterate allowed_transports [wifi], [cellular] and [wifi, cellular] against fake-platform network_type wifi, cellular, ethernet, other and unknown, and assert the task runs only when the reported type is in the list. Reject a task whose allowed_transports contains a value outside the initial list (wifi, cellular). Iterate vpn_policy any, exclude and require (and the omitted default any) against vpn_state detected, not_detected and unknown and assert run or do-not-run exactly per the public VPN eligibility table. Exercise require_validated true with validated yes, no and unknown, using the fake system method when available and the fake library internal check otherwise, and assert execution only on confirmed validation. Exercise context required_fields and allowed_country_codes with present, missing and non-matching values. Assert every do-not-run case issues no network operation. Requirements: REQ-EXECUTION-0016, REQ-EXECUTION-0017.

DEV-644: `TestCcExecutionFieldsNetworkFields` runs the allowed-transport, VPN-policy and system-validation rows with the test-only fake platform network; its Network Context rows run in `TestCcExecutionFieldsNetworkFieldsNetworkContextRequirements` since DEV-647; the family stays planned because its library-internal-check row is not executable while [GO-SPEC-ISSUE-013](specification-issues.md#go-spec-issue-013-library-internal-check-for-require_validated) is open.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-EXECUTION-0016 -->
<!-- requirement-ref: REQ-EXECUTION-0017 -->

- [REQ-EXECUTION-0016](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0016) → `CC-EXECUTION-FIELDS-NETWORK-FIELDS`.
- [REQ-EXECUTION-0017](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0017) → `CC-EXECUTION-FIELDS-NETWORK-FIELDS`.

### CC-EXECUTION-FIELDS-NETWORK-FIELDS-VPN-POLICY

Public source: `data-formats/task-execution.md`, section **Fields / Network / Fields / VPN Policy**. Intended target: `internal/execution/attempt.go`. Case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcExecutionFieldsNetworkFieldsVpnPolicy`.

Case plan: Use a fake platform that reports VPN state as detected, not_detected or unknown and evaluate each vpn_policy value (any, exclude, require) for all nine combinations. Expected: any runs for detected, not_detected and unknown; exclude runs only for not_detected and does not run for detected or unknown; require runs only for detected and does not run for not_detected or unknown. Assert omitted vpn_policy behaves as any. Assert the unknown state is carried through evaluation as unknown and is never mapped to not_detected, so exclude with unknown does not run. Reject an unregistered vpn_policy value (for example prefer) at validation. Requirements: REQ-EXECUTION-0018, REQ-EXECUTION-0019, REQ-EXECUTION-0020.

Scope: Synthetic core contract cases. Passed at Go `cdbd160` on the engine with the test-only fake platform network ([evidence](../traceability/evidence/go-unit-cdbd160.md#CC-EXECUTION-FIELDS-NETWORK-FIELDS-VPN-POLICY)); partial evidence only, with no native capability or support claim.

<!-- requirement-ref: REQ-EXECUTION-0018 -->
<!-- requirement-ref: REQ-EXECUTION-0019 -->
<!-- requirement-ref: REQ-EXECUTION-0020 -->

- [REQ-EXECUTION-0018](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0018) → `CC-EXECUTION-FIELDS-NETWORK-FIELDS-VPN-POLICY`.
- [REQ-EXECUTION-0019](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0019) → `CC-EXECUTION-FIELDS-NETWORK-FIELDS-VPN-POLICY`.
- [REQ-EXECUTION-0020](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0020) → `CC-EXECUTION-FIELDS-NETWORK-FIELDS-VPN-POLICY`.

### CC-EXECUTION-FIELDS-NETWORK-FIELDS-NETWORK-CONTEXT-REQUIREMENTS

Public source: `data-formats/task-execution.md`, section **Fields / Network / Fields / Network Context Requirements**. Intended target: `internal/execution/attempt.go`. Case location: `internal/execution/egress_test.go`, `symbol:TestCcExecutionFieldsNetworkFieldsNetworkContextRequirements`.

Case plan: Exercise required_fields/allowlists, missing/invalid/unknown context, optional context failure and stricter policy. Count requests to prove no endpoint/path override by tasks.

Scope: Planned as a whole. `TestCcExecutionFieldsNetworkFieldsNetworkContextRequirements` and `TestCcContextExecutionIntegration` cover required fields and allowlists, missing, unknown and unavailable context, optional context failure and request counts showing that tasks cannot override endpoints or paths. The stricter-local-policy row needs local-policy context eligibility, which the local policy model does not define (DEV-455). Mixed-family evidence is not promoted. No support claim.

<!-- requirement-ref: REQ-EXECUTION-0021 -->
<!-- requirement-ref: REQ-EXECUTION-0022 -->
<!-- requirement-ref: REQ-EXECUTION-0023 -->
<!-- requirement-ref: REQ-EXECUTION-0024 -->

- [REQ-EXECUTION-0021](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0021) → `CC-EXECUTION-FIELDS-NETWORK-FIELDS-NETWORK-CONTEXT-REQUIREMENTS`.
- [REQ-EXECUTION-0022](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0022) → `CC-EXECUTION-FIELDS-NETWORK-FIELDS-NETWORK-CONTEXT-REQUIREMENTS`.
- [REQ-EXECUTION-0023](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0023) → `CC-EXECUTION-FIELDS-NETWORK-FIELDS-NETWORK-CONTEXT-REQUIREMENTS`.
- [REQ-EXECUTION-0024](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0024) → `CC-EXECUTION-FIELDS-NETWORK-FIELDS-NETWORK-CONTEXT-REQUIREMENTS`.

### CC-EXECUTION-FIELDS-DEVICE-FIELDS

Public source: `data-formats/task-execution.md`, section **Fields / Device / Fields**. Intended target: `internal/execution/attempt.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcExecutionFieldsDeviceFields`.

Case plan: Exercise battery/charging constraints using synthetic platform state, with local policy stricter than task values.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-EXECUTION-0025 -->
<!-- requirement-ref: REQ-EXECUTION-0026 -->

- [REQ-EXECUTION-0025](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0025) → `CC-EXECUTION-FIELDS-DEVICE-FIELDS`.
- [REQ-EXECUTION-0026](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0026) → `CC-EXECUTION-FIELDS-DEVICE-FIELDS`.

### CC-EXECUTION-FIELDS-CONCURRENCY

Public source: `data-formats/task-execution.md`, section **Fields / Concurrency**. Intended target: `internal/execution/attempt.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcExecutionFieldsConcurrency`.

Case plan: Schedule synthetic tasks on a deterministic clock with scripted transports that record start and end times: two shared tasks S1 and S2 and one exclusive task E, submitted in the orders S1,S2,E and E,S1,S2 and with E submitted while S1 is running. Assert S1 and S2 may overlap, that E's execution interval never overlaps any other measurement's interval, and that no shared task starts while E is active. Assert that when E cannot run alone it either waits or ends not_run per the public constraint-evaluation rules, without starting any network operation while another measurement is active. Requirements: REQ-EXECUTION-0027.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-EXECUTION-0027 -->

- [REQ-EXECUTION-0027](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0027) → `CC-EXECUTION-FIELDS-CONCURRENCY`.

### CC-EXECUTION-CONSTRAINT-EVALUATION

Public source: `data-formats/task-execution.md`, section **Constraint Evaluation**. Intended target: `internal/execution/attempt.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcExecutionConstraintEvaluation`.

Case plan: Independently fail every applicable runtime constraint and local/OS/user veto; no measurement starts.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-EXECUTION-0028 -->
<!-- requirement-ref: REQ-EXECUTION-0029 -->
<!-- requirement-ref: REQ-EXECUTION-0030 -->

- [REQ-EXECUTION-0028](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0028) → `CC-EXECUTION-CONSTRAINT-EVALUATION`.
- [REQ-EXECUTION-0029](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0029) → `CC-EXECUTION-CONSTRAINT-EVALUATION`.
- [REQ-EXECUTION-0030](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-execution.md#REQ-EXECUTION-0030) → `CC-EXECUTION-CONSTRAINT-EVALUATION`.

### CC-INPUT-ENVELOPE

Public source: `data-formats/task-input.md`, section **Envelope**. Intended target: `internal/validation/input.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcInputEnvelope`.

Case plan: Register two synthetic local test implementations with the same test_id and different test_version values whose input schemas define different fields (version A requires field a, version B requires field b). Submit a task for version A with input containing a and assert acceptance; submit the same input under version B and assert rejection; submit input containing b under version A and assert rejection as an unknown field. Assert schema selection uses the exact (test_id, test_version) pair: an unregistered test_version, or an unregistered test_id with otherwise valid input, is rejected as a task-validation failure. Assert the generic task decoder accepts input only as an object and defines no input fields of its own. Requirements: REQ-INPUT-0001.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-INPUT-0001 -->

- [REQ-INPUT-0001](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0001) → `CC-INPUT-ENVELOPE`.

### CC-INPUT-INPUT-MODEL

Public source: `data-formats/task-input.md`, section **Input Model**. Intended target: `internal/validation/input.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcInputInputModel`.

Case plan: Remove the locally registered input schema and make a fake schema downloader fail if called; reject before scheduling with zero network calls.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-INPUT-0002 -->

- [REQ-INPUT-0002](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0002) → `CC-INPUT-INPUT-MODEL`.

### CC-INPUT-JSON-STRUCTURE

Public source: `data-formats/task-input.md`, section **JSON Structure**. Intended target: `internal/validation/input.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcInputJsonStructure`.

Case plan: Decode otherwise valid synthetic tasks through task admission, varying only the top-level `input` value. Accept `input: {}` for NET-01, whose registered input schema accepts an empty object, and assert the task is admitted. Reject `input: {}` for DNS-01, whose registered input schema requires `hostname` and `record_type`, with a bounded non-measurement admission error. For both NET-01 and DNS-01, iterate the non-object values `null`, `[]`, `[{}]`, `""`, `"{}"`, `0`, `1.5`, `true` and `false` as top-level `input`, and assert each is rejected at admission with a non-measurement admission error, no run or attempt is created and no Measurement Result is emitted. Verify a DNS-01 task with `input: {"hostname": "example.org", "record_type": "A"}` decodes to a JSON object value. Requirements: REQ-INPUT-0003, REQ-INPUT-0004.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-INPUT-0003 -->
<!-- requirement-ref: REQ-INPUT-0004 -->

- [REQ-INPUT-0003](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0003) → `CC-INPUT-JSON-STRUCTURE`.
- [REQ-INPUT-0004](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0004) → `CC-INPUT-JSON-STRUCTURE`.

### CC-INPUT-INPUT-SCHEMA-REQUIREMENTS

Public source: `data-formats/task-input.md`, section **Input Schema Requirements**. Intended target: `internal/validation/input.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcInputInputSchemaRequirements`.

Case plan: Parameterize required/type/empty/string/array/range/protocol/normalization/additional-field constraints from a registered schema. Exercise encoded-byte boundary and missing required subject without environment defaults.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-INPUT-0005 -->
<!-- requirement-ref: REQ-INPUT-0006 -->
<!-- requirement-ref: REQ-INPUT-0007 -->
<!-- requirement-ref: REQ-INPUT-0008 -->

- [REQ-INPUT-0005](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0005) → `CC-INPUT-INPUT-SCHEMA-REQUIREMENTS`.
- [REQ-INPUT-0006](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0006) → `CC-INPUT-INPUT-SCHEMA-REQUIREMENTS`.
- [REQ-INPUT-0007](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0007) → `CC-INPUT-INPUT-SCHEMA-REQUIREMENTS`.
- [REQ-INPUT-0008](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0008) → `CC-INPUT-INPUT-SCHEMA-REQUIREMENTS`.

### CC-INPUT-VALIDATION

Public source: `data-formats/task-input.md`, section **Validation**. Intended target: `internal/validation/input.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcInputValidation`.

Case plan: Inject each listed input failure independently, including an unavailable schema and denied target; reject before scheduling, return a task error and emit no measurement.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-INPUT-0009 -->
<!-- requirement-ref: REQ-INPUT-0010 -->
<!-- requirement-ref: REQ-INPUT-0011 -->
<!-- requirement-ref: REQ-INPUT-0012 -->
<!-- requirement-ref: REQ-INPUT-0027 -->
<!-- requirement-ref: REQ-INPUT-0028 -->
<!-- requirement-ref: REQ-INPUT-0029 -->
<!-- requirement-ref: REQ-INPUT-0030 -->
<!-- requirement-ref: REQ-INPUT-0031 -->
<!-- requirement-ref: REQ-INPUT-0032 -->
<!-- requirement-ref: REQ-INPUT-0033 -->
<!-- requirement-ref: REQ-INPUT-0034 -->

- [REQ-INPUT-0009](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0009) → `CC-INPUT-VALIDATION`.
- [REQ-INPUT-0010](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0010) → `CC-INPUT-VALIDATION`.
- [REQ-INPUT-0011](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0011) → `CC-INPUT-VALIDATION`.
- [REQ-INPUT-0012](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0012) → `CC-INPUT-VALIDATION`.
- [REQ-INPUT-0027](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0027) → `CC-INPUT-VALIDATION`.
- [REQ-INPUT-0028](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0028) → `CC-INPUT-VALIDATION`.
- [REQ-INPUT-0029](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0029) → `CC-INPUT-VALIDATION`.
- [REQ-INPUT-0030](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0030) → `CC-INPUT-VALIDATION`.
- [REQ-INPUT-0031](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0031) → `CC-INPUT-VALIDATION`.
- [REQ-INPUT-0032](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0032) → `CC-INPUT-VALIDATION`.
- [REQ-INPUT-0033](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0033) → `CC-INPUT-VALIDATION`.
- [REQ-INPUT-0034](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0034) → `CC-INPUT-VALIDATION`.

### CC-INPUT-NORMALIZATION

Public source: `data-formats/task-input.md`, section **Normalization**. Intended target: `internal/validation/input.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcInputNormalization`.

Case plan: Repeat normalization with equivalent supported inputs; reject ambiguous subject changes. Preserve only the normalized fields allowed by the exact Output schema.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-INPUT-0013 -->
<!-- requirement-ref: REQ-INPUT-0014 -->
<!-- requirement-ref: REQ-INPUT-0015 -->

- [REQ-INPUT-0013](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0013) → `CC-INPUT-NORMALIZATION`.
- [REQ-INPUT-0014](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0014) → `CC-INPUT-NORMALIZATION`.
- [REQ-INPUT-0015](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0015) → `CC-INPUT-NORMALIZATION`.

### CC-INPUT-TARGET-SAFETY

Public source: `data-formats/task-input.md`, section **Target Safety**. Intended target: `internal/validation/input.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcInputTargetSafety`.

Case plan: Use syntactically valid but unregistered/prohibited local, private, link-local and credential-bearing targets. Recheck redirects before contact; count zero contacts to denied destinations.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-INPUT-0016 -->
<!-- requirement-ref: REQ-INPUT-0017 -->
<!-- requirement-ref: REQ-INPUT-0018 -->

- [REQ-INPUT-0016](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0016) → `CC-INPUT-TARGET-SAFETY`.
- [REQ-INPUT-0017](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0017) → `CC-INPUT-TARGET-SAFETY`.
- [REQ-INPUT-0018](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0018) → `CC-INPUT-TARGET-SAFETY`.

### CC-INPUT-SENSITIVE-AND-DYNAMIC-INPUT

Public source: `data-formats/task-input.md`, section **Sensitive and Dynamic Input**. Intended target: `internal/validation/input.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcInputSensitiveAndDynamicInput`.

Case plan: Attempt every prohibited sensitive input class using synthetic canaries; verify no runtime substitution or automatic access to host/device data.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-INPUT-0019 -->
<!-- requirement-ref: REQ-INPUT-0020 -->

- [REQ-INPUT-0019](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0019) → `CC-INPUT-SENSITIVE-AND-DYNAMIC-INPUT`.
- [REQ-INPUT-0020](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0020) → `CC-INPUT-SENSITIVE-AND-DYNAMIC-INPUT`.

### CC-INPUT-MULTIPLE-TARGETS

Public source: `data-formats/task-input.md`, section **Multiple Targets**. Intended target: `internal/validation/input.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcInputMultipleTargets`.

Case plan: Reject an unregistered generic hostnames array; exercise the exact registered array item/count/encoded-size boundaries for any supported multi-target schema.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-INPUT-0021 -->
<!-- requirement-ref: REQ-INPUT-0022 -->
<!-- requirement-ref: REQ-INPUT-0023 -->

- [REQ-INPUT-0021](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0021) → `CC-INPUT-MULTIPLE-TARGETS`.
- [REQ-INPUT-0022](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0022) → `CC-INPUT-MULTIPLE-TARGETS`.
- [REQ-INPUT-0023](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0023) → `CC-INPUT-MULTIPLE-TARGETS`.

### CC-INPUT-RESULT-PROPAGATION

Public source: `data-formats/task-input.md`, section **Result Propagation**. Intended target: `internal/validation/input.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcInputResultPropagation`.

Case plan: Construct output from a task containing canary inputs; retain only fields authorized by the selected Output schema and local policy.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-INPUT-0024 -->

- [REQ-INPUT-0024](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0024) → `CC-INPUT-RESULT-PROPAGATION`.

### CC-INPUT-IMPLEMENTATION-AUTHORITY

Public source: `data-formats/task-input.md`, section **Implementation Authority**. Intended target: `internal/validation/input.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcInputImplementationAuthority`.

Case plan: Try to widen input schemas or weaken local restrictions using task input; verify the registered schema and stricter local policy prevail.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-INPUT-0025 -->
<!-- requirement-ref: REQ-INPUT-0026 -->

- [REQ-INPUT-0025](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0025) → `CC-INPUT-IMPLEMENTATION-AUTHORITY`.
- [REQ-INPUT-0026](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-input.md#REQ-INPUT-0026) → `CC-INPUT-IMPLEMENTATION-AUTHORITY`.

### CC-LABEL-ENVELOPE

Public source: `data-formats/task-labels.md`, section **Envelope**. Intended target: `internal/taskdecode/labels.go`. Planned case location: `internal/taskdecode/cross_cutting_conformance_test.go`, `symbol:TestCcLabelEnvelope`.

Case plan: Execute the same synthetic task against a recording transport and deterministic clock twice, once with labels {"campaign": "a"} and once with labels {"campaign": "b", "measurement_group": "baseline"}. Assert identical network operations, schedule due times, execution-constraint decisions, limits and safety-policy outcomes; only the labels carried through to output may differ. Supply labels whose keys resemble typed control fields (for example network_path with value proxy, max_runs with value 28, vpn_policy with value require) and assert the typed execution and schedule values remain those of the typed fields, with no behavior change. Requirements: REQ-LABEL-0001, REQ-LABEL-0002.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LABEL-0001 -->
<!-- requirement-ref: REQ-LABEL-0002 -->

- [REQ-LABEL-0001](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0001) → `CC-LABEL-ENVELOPE`.
- [REQ-LABEL-0002](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0002) → `CC-LABEL-ENVELOPE`.

### CC-LABEL-JSON-STRUCTURE

Public source: `data-formats/task-labels.md`, section **JSON Structure**. Intended target: `internal/taskdecode/labels.go`. Planned case location: `internal/taskdecode/cross_cutting_conformance_test.go`, `symbol:TestCcLabelJsonStructure`.

Case plan: Decode synthetic tasks with labels omitted and with labels {} and assert they are equivalent in the decoded task and in downstream behavior. Accept a flat object with string keys and string values (for example {"campaign": "x"}). Reject labels as an array, string, number or null, a nested object value, an array value, and non-string values (number, boolean, null). Requirements: REQ-LABEL-0003, REQ-LABEL-0004.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LABEL-0003 -->
<!-- requirement-ref: REQ-LABEL-0004 -->

- [REQ-LABEL-0003](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0003) → `CC-LABEL-JSON-STRUCTURE`.
- [REQ-LABEL-0004](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0004) → `CC-LABEL-JSON-STRUCTURE`.

### CC-LABEL-CONSTRAINTS

Public source: `data-formats/task-labels.md`, section **Constraints**. Intended target: `internal/taskdecode/labels.go`. Planned case location: `internal/taskdecode/cross_cutting_conformance_test.go`, `symbol:TestCcLabelConstraints`.

Case plan: Exercise 16 entries, 64-byte keys, 128-byte values and 2048 encoded bytes at/below/above bounds, counting JSON syntax and multibyte text.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LABEL-0005 -->
<!-- requirement-ref: REQ-LABEL-0006 -->
<!-- requirement-ref: REQ-LABEL-0007 -->

- [REQ-LABEL-0005](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0005) → `CC-LABEL-CONSTRAINTS`.
- [REQ-LABEL-0006](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0006) → `CC-LABEL-CONSTRAINTS`.
- [REQ-LABEL-0007](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0007) → `CC-LABEL-CONSTRAINTS`.

### CC-LABEL-KEYS

Public source: `data-formats/task-labels.md`, section **Keys**. Intended target: `internal/taskdecode/labels.go`. Planned case location: `internal/taskdecode/cross_cutting_conformance_test.go`, `symbol:TestCcLabelKeys`.

Case plan: Test the exact key alphabet, initial character, case sensitivity, length, and equal/conflicting/escaped duplicate names.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LABEL-0008 -->
<!-- requirement-ref: REQ-LABEL-0009 -->
<!-- requirement-ref: REQ-LABEL-0010 -->

- [REQ-LABEL-0008](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0008) → `CC-LABEL-KEYS`.
- [REQ-LABEL-0009](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0009) → `CC-LABEL-KEYS`.
- [REQ-LABEL-0010](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0010) → `CC-LABEL-KEYS`.

### CC-LABEL-KEYS-RESERVED-NAMESPACE

Public source: `data-formats/task-labels.md`, section **Keys / Reserved Namespace**. Intended target: `internal/taskdecode/labels.go`. Planned case location: `internal/taskdecode/cross_cutting_conformance_test.go`, `symbol:TestCcLabelKeysReservedNamespace`.

Case plan: Reject an unknown barometer. key; accept a structurally valid unreserved key without assigning it control semantics.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LABEL-0011 -->
<!-- requirement-ref: REQ-LABEL-0012 -->

- [REQ-LABEL-0011](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0011) → `CC-LABEL-KEYS-RESERVED-NAMESPACE`.
- [REQ-LABEL-0012](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0012) → `CC-LABEL-KEYS-RESERVED-NAMESPACE`.

### CC-LABEL-VALUES

Public source: `data-formats/task-labels.md`, section **Values**. Intended target: `internal/taskdecode/labels.go`. Planned case location: `internal/taskdecode/cross_cutting_conformance_test.go`, `symbol:TestCcLabelValues`.

Case plan: Test UTF-8 and case preservation, control-character rejection, wrong types, literal expression text and absence of automatic numeric/boolean coercion.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LABEL-0013 -->
<!-- requirement-ref: REQ-LABEL-0014 -->
<!-- requirement-ref: REQ-LABEL-0015 -->
<!-- requirement-ref: REQ-LABEL-0016 -->
<!-- requirement-ref: REQ-LABEL-0017 -->

- [REQ-LABEL-0013](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0013) → `CC-LABEL-VALUES`.
- [REQ-LABEL-0014](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0014) → `CC-LABEL-VALUES`.
- [REQ-LABEL-0015](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0015) → `CC-LABEL-VALUES`.
- [REQ-LABEL-0016](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0016) → `CC-LABEL-VALUES`.
- [REQ-LABEL-0017](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0017) → `CC-LABEL-VALUES`.

### CC-LABEL-PRIVACY-AND-SAFETY

Public source: `data-formats/task-labels.md`, section **Privacy and Safety**. Intended target: `internal/taskdecode/labels.go`. Planned case location: `internal/taskdecode/cross_cutting_conformance_test.go`, `symbol:TestCcLabelPrivacyAndSafety`.

Case plan: Use synthetic canaries for every prohibited identifier/content class and interpolation request; preserve permitted literal-or-reject handling and stricter label removal/rejection policy.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LABEL-0018 -->
<!-- requirement-ref: REQ-LABEL-0019 -->
<!-- requirement-ref: REQ-LABEL-0020 -->
<!-- requirement-ref: REQ-LABEL-0021 -->
<!-- requirement-ref: REQ-LABEL-0022 -->
<!-- requirement-ref: REQ-LABEL-0023 -->

- [REQ-LABEL-0018](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0018) → `CC-LABEL-PRIVACY-AND-SAFETY`.
- [REQ-LABEL-0019](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0019) → `CC-LABEL-PRIVACY-AND-SAFETY`.
- [REQ-LABEL-0020](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0020) → `CC-LABEL-PRIVACY-AND-SAFETY`.
- [REQ-LABEL-0021](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0021) → `CC-LABEL-PRIVACY-AND-SAFETY`.
- [REQ-LABEL-0022](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0022) → `CC-LABEL-PRIVACY-AND-SAFETY`.
- [REQ-LABEL-0023](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0023) → `CC-LABEL-PRIVACY-AND-SAFETY`.

### CC-LABEL-PROCESSING

Public source: `data-formats/task-labels.md`, section **Processing**. Intended target: `internal/taskdecode/labels.go`. Planned case location: `internal/taskdecode/cross_cutting_conformance_test.go`, `symbol:TestCcLabelProcessing`.

Case plan: Vary every apparent behavior/destination label and verify execution is unchanged; accept unknown unreserved metadata and exclude labels from current results.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LABEL-0024 -->
<!-- requirement-ref: REQ-LABEL-0025 -->
<!-- requirement-ref: REQ-LABEL-0026 -->
<!-- requirement-ref: REQ-LABEL-0027 -->

- [REQ-LABEL-0024](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0024) → `CC-LABEL-PROCESSING`.
- [REQ-LABEL-0025](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0025) → `CC-LABEL-PROCESSING`.
- [REQ-LABEL-0026](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0026) → `CC-LABEL-PROCESSING`.
- [REQ-LABEL-0027](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0027) → `CC-LABEL-PROCESSING`.

### CC-LABEL-VALIDATION

Public source: `data-formats/task-labels.md`, section **Validation**. Intended target: `internal/taskdecode/labels.go`. Planned case location: `internal/taskdecode/cross_cutting_conformance_test.go`, `symbol:TestCcLabelValidation`.

Case plan: Mutate each validity predicate independently; return task-validation errors without emitting measurements.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LABEL-0028 -->
<!-- requirement-ref: REQ-LABEL-0029 -->

- [REQ-LABEL-0028](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0028) → `CC-LABEL-VALIDATION`.
- [REQ-LABEL-0029](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-labels.md#REQ-LABEL-0029) → `CC-LABEL-VALIDATION`.

### CC-OPTIONS-ENVELOPE

Public source: `data-formats/task-options.md`, section **Envelope**. Intended target: `internal/validation/options.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcOptionsEnvelope`.

Case plan: Use identical task envelopes with different exact test versions; select each version-specific options schema without changing the outer execution envelope.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OPTIONS-0001 -->
<!-- requirement-ref: REQ-OPTIONS-0002 -->

- [REQ-OPTIONS-0001](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0001) → `CC-OPTIONS-ENVELOPE`.
- [REQ-OPTIONS-0002](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0002) → `CC-OPTIONS-ENVELOPE`.

### CC-OPTIONS-JSON-STRUCTURE

Public source: `data-formats/task-options.md`, section **JSON Structure**. Intended target: `internal/validation/options.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcOptionsJsonStructure`.

Case plan: Decode otherwise valid synthetic NET-01 and DNS-01 tasks through task admission, varying only `options`. Reject a task with the `options` member omitted, for both tests, with a bounded non-measurement admission error, even though every DNS-01 option property is optional. Accept `options: {}` for NET-01 (no test-specific options defined) and for DNS-01, and assert the DNS-01 effective configuration equals its documented defaults: `resolver_mode` `system`, `transport` `udp`, `collect_additional_responses` `false`. For both tests, iterate the top-level values `null`, `[]`, `[{}]`, `""`, `"{}"`, `0`, `1.5`, `true` and `false` and assert each is rejected at admission with no run, attempt or Measurement Result created. Verify a DNS-01 task with `options: {"resolver_mode": "system"}` is accepted, so individual properties remain optional while the object is required. Requirements: REQ-OPTIONS-0003, REQ-OPTIONS-0004, REQ-OPTIONS-0005.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OPTIONS-0003 -->
<!-- requirement-ref: REQ-OPTIONS-0004 -->
<!-- requirement-ref: REQ-OPTIONS-0005 -->

- [REQ-OPTIONS-0003](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0003) → `CC-OPTIONS-JSON-STRUCTURE`.
- [REQ-OPTIONS-0004](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0004) → `CC-OPTIONS-JSON-STRUCTURE`.
- [REQ-OPTIONS-0005](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0005) → `CC-OPTIONS-JSON-STRUCTURE`.

### CC-OPTIONS-OPTIONS-SCHEMA

Public source: `data-formats/task-options.md`, section **Options Schema**. Intended target: `internal/validation/options.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcOptionsOptionsSchema`.

Case plan: Remove the local schema; parameterize required fields, types, defaults, enum/range/count/encoded-byte bounds and input relationships; reject unknown properties without network schema acquisition.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OPTIONS-0006 -->
<!-- requirement-ref: REQ-OPTIONS-0007 -->
<!-- requirement-ref: REQ-OPTIONS-0008 -->
<!-- requirement-ref: REQ-OPTIONS-0009 -->
<!-- requirement-ref: REQ-OPTIONS-0010 -->

- [REQ-OPTIONS-0006](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0006) → `CC-OPTIONS-OPTIONS-SCHEMA`.
- [REQ-OPTIONS-0007](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0007) → `CC-OPTIONS-OPTIONS-SCHEMA`.
- [REQ-OPTIONS-0008](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0008) → `CC-OPTIONS-OPTIONS-SCHEMA`.
- [REQ-OPTIONS-0009](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0009) → `CC-OPTIONS-OPTIONS-SCHEMA`.
- [REQ-OPTIONS-0010](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0010) → `CC-OPTIONS-OPTIONS-SCHEMA`.

### CC-OPTIONS-VALIDATION

Public source: `data-formats/task-options.md`, section **Validation**. Intended target: `internal/validation/options.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcOptionsValidation`.

Case plan: Inject each listed unsupported, unavailable, inconsistent or out-of-envelope option case; reject before scheduling with no Measurement Result and no version fallback.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OPTIONS-0011 -->
<!-- requirement-ref: REQ-OPTIONS-0012 -->
<!-- requirement-ref: REQ-OPTIONS-0013 -->
<!-- requirement-ref: REQ-OPTIONS-0014 -->
<!-- requirement-ref: REQ-OPTIONS-0015 -->
<!-- requirement-ref: REQ-OPTIONS-0025 -->
<!-- requirement-ref: REQ-OPTIONS-0026 -->
<!-- requirement-ref: REQ-OPTIONS-0027 -->
<!-- requirement-ref: REQ-OPTIONS-0028 -->
<!-- requirement-ref: REQ-OPTIONS-0029 -->
<!-- requirement-ref: REQ-OPTIONS-0030 -->
<!-- requirement-ref: REQ-OPTIONS-0031 -->
<!-- requirement-ref: REQ-OPTIONS-0032 -->

- [REQ-OPTIONS-0011](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0011) → `CC-OPTIONS-VALIDATION`.
- [REQ-OPTIONS-0012](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0012) → `CC-OPTIONS-VALIDATION`.
- [REQ-OPTIONS-0013](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0013) → `CC-OPTIONS-VALIDATION`.
- [REQ-OPTIONS-0014](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0014) → `CC-OPTIONS-VALIDATION`.
- [REQ-OPTIONS-0015](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0015) → `CC-OPTIONS-VALIDATION`.
- [REQ-OPTIONS-0025](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0025) → `CC-OPTIONS-VALIDATION`.
- [REQ-OPTIONS-0026](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0026) → `CC-OPTIONS-VALIDATION`.
- [REQ-OPTIONS-0027](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0027) → `CC-OPTIONS-VALIDATION`.
- [REQ-OPTIONS-0028](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0028) → `CC-OPTIONS-VALIDATION`.
- [REQ-OPTIONS-0029](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0029) → `CC-OPTIONS-VALIDATION`.
- [REQ-OPTIONS-0030](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0030) → `CC-OPTIONS-VALIDATION`.
- [REQ-OPTIONS-0031](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0031) → `CC-OPTIONS-VALIDATION`.
- [REQ-OPTIONS-0032](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0032) → `CC-OPTIONS-VALIDATION`.

### CC-OPTIONS-EXECUTION-PRECEDENCE

Public source: `data-formats/task-options.md`, section **Execution Precedence**. Intended target: `internal/validation/options.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcOptionsExecutionPrecedence`.

Case plan: Set a protocol timeout beyond the outer attempt deadline; exercise validation and outer cancellation without converting the outer interruption to a protocol outcome.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OPTIONS-0016 -->
<!-- requirement-ref: REQ-OPTIONS-0017 -->

- [REQ-OPTIONS-0016](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0016) → `CC-OPTIONS-EXECUTION-PRECEDENCE`.
- [REQ-OPTIONS-0017](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0017) → `CC-OPTIONS-EXECUTION-PRECEDENCE`.

### CC-OPTIONS-NO-SILENT-REWRITING

Public source: `data-formats/task-options.md`, section **No Silent Rewriting**. Intended target: `internal/validation/options.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcOptionsNoSilentRewriting`.

Case plan: Supply unsupported/unsafe explicit options and omitted defaults separately; observe rejection at admission with the REQ-TRUST-0056 class, never skipping, and no clamping, coercion or substitution.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OPTIONS-0018 -->
<!-- requirement-history: REQ-OPTIONS-0019 -->
<!-- requirement-ref: REQ-OPTIONS-0033 -->
<!-- requirement-ref: REQ-OPTIONS-0020 -->

- [REQ-OPTIONS-0018](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0018) → `CC-OPTIONS-NO-SILENT-REWRITING`.
- [REQ-OPTIONS-0019](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0019) (retired at `1206247`) → `CC-OPTIONS-NO-SILENT-REWRITING`.
- [REQ-OPTIONS-0033](https://github.com/tv-rain/barometer-spec/blob/12062471875bbfbc55eafc94dc5a7ea4cd07ca5e/data-formats/task-options.md#REQ-OPTIONS-0033) → `CC-OPTIONS-NO-SILENT-REWRITING`.
- [REQ-OPTIONS-0020](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0020) → `CC-OPTIONS-NO-SILENT-REWRITING`.

### CC-OPTIONS-SECURITY-AND-SAFETY

Public source: `data-formats/task-options.md`, section **Security and Safety**. Intended target: `internal/validation/options.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcOptionsSecurityAndSafety`.

Case plan: Attempt each listed option-based code/file/secret/interpolation/control-channel bypass; preserve registered behavior and stricter local safety policy.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OPTIONS-0021 -->
<!-- requirement-ref: REQ-OPTIONS-0022 -->
<!-- requirement-ref: REQ-OPTIONS-0023 -->
<!-- requirement-ref: REQ-OPTIONS-0024 -->

- [REQ-OPTIONS-0021](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0021) → `CC-OPTIONS-SECURITY-AND-SAFETY`.
- [REQ-OPTIONS-0022](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0022) → `CC-OPTIONS-SECURITY-AND-SAFETY`.
- [REQ-OPTIONS-0023](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0023) → `CC-OPTIONS-SECURITY-AND-SAFETY`.
- [REQ-OPTIONS-0024](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-options.md#REQ-OPTIONS-0024) → `CC-OPTIONS-SECURITY-AND-SAFETY`.

### CC-SAFETY-ENVELOPE

Public source: `data-formats/task-safety.md`, section **Envelope**. Intended target: `internal/validation/safety.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcSafetyEnvelope`.

Case plan: Admit a medium-risk synthetic task (registered baseline low) under a fake local policy that permits medium risk and schedules three attempts. Before each attempt change the trusted local inputs: attempt 1 permitted and started; attempt 2 with the local maximum lowered to low, not started; attempt 3 with the permission restored, started. Assert the declaration is evaluated before every attempt, not only at admission. Assert a fake host restriction can delay or skip an attempt even when every declared condition is satisfied. Count runs and attempts before and after each evaluation and assert none is created. Admission rows: a task declaring low risk with requires_user_consent false gains no permission over the same task declaring medium with consent required, and declaration fields cannot alter provenance, issuer authorization, target registration or policy. The per-attempt rows are execution-owned. Requirements: REQ-SAFETY-0001, REQ-SAFETY-0002, REQ-SAFETY-0003.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SAFETY-0001 -->
<!-- requirement-ref: REQ-SAFETY-0002 -->
<!-- requirement-ref: REQ-SAFETY-0003 -->

- [REQ-SAFETY-0001](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0001) → `CC-SAFETY-ENVELOPE`.
- [REQ-SAFETY-0002](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0002) → `CC-SAFETY-ENVELOPE`.
- [REQ-SAFETY-0003](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0003) → `CC-SAFETY-ENVELOPE`.

### CC-SAFETY-SAFETY-MODEL

Public source: `data-formats/task-safety.md`, section **Safety Model**. Intended target: `internal/validation/safety.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcSafetySafetyModel`.

Case plan: Use a synthetic registry with baselines low (NET-01) and medium (a synthetic medium-baseline test). For each baseline, declare low, medium, high and unknown with consistent risk_factors. Assert effective risk is never below the baseline: a low declaration on the medium baseline is rejected or evaluated as medium. Higher declarations raise effective risk to the declared level. Options or a schedule that a fake local assessment marks as increasing risk raise it further. Count runs and attempts with a fake scheduler before and after safety evaluation and assert evaluation adds none. Requirements: REQ-SAFETY-0004, REQ-SAFETY-0005, REQ-SAFETY-0006.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SAFETY-0004 -->
<!-- requirement-ref: REQ-SAFETY-0005 -->
<!-- requirement-ref: REQ-SAFETY-0006 -->

- [REQ-SAFETY-0004](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0004) → `CC-SAFETY-SAFETY-MODEL`.
- [REQ-SAFETY-0005](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0005) → `CC-SAFETY-SAFETY-MODEL`.
- [REQ-SAFETY-0006](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0006) → `CC-SAFETY-SAFETY-MODEL`.

### CC-SAFETY-FIELDS

Public source: `data-formats/task-safety.md`, section **Fields**. Intended target: `internal/validation/safety.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcSafetyFields`.

Case plan: Validate the safety object field by field. Accept {risk_level: low, requires_user_consent: false} and the public complete example. Reject: missing risk_level; missing requires_user_consent; risk_level as 1, null, true or an array; requires_user_consent as "true", 1 or null; risk_factors as an object, a string or null; a risk factor given as a string; an unknown member in safety (risk_score) and in a factor (weight); and a duplicate member name. Each rejection is task_schema_invalid, produces no candidate and causes zero transport calls. Requirements: REQ-SAFETY-0007.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SAFETY-0007 -->

- [REQ-SAFETY-0007](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0007) → `CC-SAFETY-FIELDS`.

### CC-SAFETY-RISK-LEVEL

Public source: `data-formats/task-safety.md`, section **Risk Level**. Intended target: `internal/validation/safety.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcSafetyRiskLevel`.

Case plan: Admit synthetic NET-01 tasks (registered baseline `low`) under a fake local policy and deterministic clock, varying `safety.risk_level` with consistent `risk_factors` where required. Iterate `low`, `medium`, `high` and `unknown` and assert each is decoded as a supported value. Reject `none`, `""`, `"LOW"`, `"Unknown"`, `null`, `0` and a missing `risk_level` at admission with a bounded non-measurement admission error. Exercise `unknown` (with a `risk_factors` entry of level `unknown`) under a local policy with no unknown-risk permission and assert no execution attempt starts and no Measurement Result is produced (fail closed); repeat with a policy explicitly marked reviewed and permitting `unknown` and assert the run becomes eligible. Exercise `high` under a default policy and assert it is not executed, then under a reviewed policy explicitly permitting `high` and assert it becomes eligible. Assert no serialized task, admission status or result value represents a risk level of `none` or any risk-free value. Requirements: REQ-SAFETY-0008, REQ-SAFETY-0009, REQ-SAFETY-0010, REQ-SAFETY-0011.

Scope: Synthetic core contract cases. No execution evidence or support claim.

Review case `RV-CC-SAFETY-RISK-LEVEL` (`docs/conformance/cross-cutting-review.md`, `heading:RV-CC-SAFETY-RISK-LEVEL`): Human review case; not automated coverage. Reviewer role: Safety and privacy reviewer (independent of the implementer). Reviewed artifact: Go public API documentation, host integration guidance and any host-facing risk-level descriptions at the reviewed implementation revision. Checklist: (1) Does any documentation or host-facing text describe a measurement or task as risk-free or as having no risk? (2) Is `low` described as a conservative classification under assessed deployment assumptions rather than a guarantee that execution is safe in every environment? (3) Is `unknown` documented as failing closed by default, runnable only under an explicit, reviewed local policy? (4) Is `high` documented as normally rejected unless explicitly permitted by a reviewed local policy? (5) Does guidance avoid presenting risk levels as guarantees to hosts or end users? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-SAFETY-0008 -->
<!-- requirement-ref: REQ-SAFETY-0009 -->
<!-- requirement-ref: REQ-SAFETY-0010 -->
<!-- requirement-ref: REQ-SAFETY-0011 -->

- [REQ-SAFETY-0008](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0008) → `CC-SAFETY-RISK-LEVEL`, `RV-CC-SAFETY-RISK-LEVEL`.
- [REQ-SAFETY-0009](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0009) → `CC-SAFETY-RISK-LEVEL`, `RV-CC-SAFETY-RISK-LEVEL`.
- [REQ-SAFETY-0010](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0010) → `CC-SAFETY-RISK-LEVEL`, `RV-CC-SAFETY-RISK-LEVEL`.
- [REQ-SAFETY-0011](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0011) → `CC-SAFETY-RISK-LEVEL`, `RV-CC-SAFETY-RISK-LEVEL`.

### CC-SAFETY-RISK-FACTORS

Public source: `data-formats/task-safety.md`, section **Risk Factors**. Intended target: `internal/validation/safety.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcSafetyRiskFactors`.

Case plan: Admit synthetic tasks under a fake local policy, varying `risk_level` and the presence of `risk_factors`. Accept `risk_level: low` with `risk_factors` omitted, and accept `low` with `risk_factors: [{"type": "network_observability", "level": "low"}]`. Reject `medium`, `high` and `unknown` each with `risk_factors` omitted, with a bounded non-measurement admission error and no run or attempt created. Reject `medium`, `high` and `unknown` each with `risk_factors: []` with the same bounded non-measurement admission error and no run or attempt created, and accept `low` with `risk_factors: []`. Accept `medium` with `[{"type": "traffic_distinctiveness", "level": "medium"}]`, `high` with `[{"type": "target_sensitivity", "level": "high"}]` and `unknown` with `[{"type": "data_sensitivity", "level": "unknown"}]` as structurally valid declarations (execution permission for `high` and `unknown` still depends on local policy). Requirements: REQ-SAFETY-0012, REQ-SAFETY-0031.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SAFETY-0012 -->
<!-- requirement-ref: REQ-SAFETY-0031 -->

- [REQ-SAFETY-0012](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0012) → `CC-SAFETY-RISK-FACTORS`.
- [REQ-SAFETY-0031](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0031) → `CC-SAFETY-RISK-FACTORS`.

### CC-SAFETY-RISK-FACTORS-RISK-FACTOR-FIELDS

Public source: `data-formats/task-safety.md`, section **Risk Factors / Risk Factor Fields**. Intended target: `internal/validation/safety.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcSafetyRiskFactorsRiskFactorFields`.

Case plan: Admit synthetic tasks with `risk_level: unknown` (so every factor level is consistent) containing one risk factor, varying the factor object. Reject a factor missing `type` and a factor missing `level`, each with a bounded non-measurement admission error. Iterate factor `level` over `low`, `medium`, `high` and `unknown` and assert each is accepted; reject `none`, `""`, `"Medium"`, `null`, `1` and `true`. Iterate `type` over each initial registered type (`network_observability`, `traffic_distinctiveness`, `target_sensitivity`, `data_sensitivity`, `legal_or_social_context`) with `level: low` and assert each is accepted. Verify the decoded factor exposes exactly the `type` and `level` values supplied. Requirements: REQ-SAFETY-0013, REQ-SAFETY-0014.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SAFETY-0013 -->
<!-- requirement-ref: REQ-SAFETY-0014 -->

- [REQ-SAFETY-0013](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0013) → `CC-SAFETY-RISK-FACTORS-RISK-FACTOR-FIELDS`.
- [REQ-SAFETY-0014](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0014) → `CC-SAFETY-RISK-FACTORS-RISK-FACTOR-FIELDS`.

### CC-SAFETY-RISK-FACTORS-INITIAL-RISK-FACTOR-TYPES

Public source: `data-formats/task-safety.md`, section **Risk Factors / Initial Risk Factor Types**. Intended target: `internal/validation/safety.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcSafetyRiskFactorsInitialRiskFactorTypes`.

Case plan: With risk_level unknown and one factor of level unknown, iterate the five initial types and assert each is accepted and exposed unchanged. Submit unknown types (battery_drain, Network_observability, network-observability, an empty string and a 65-byte string) and assert each task is either rejected before scheduling with task_unsupported, never task_schema_invalid, or treated as unknown risk, and is never accepted with the factor ignored; the current validator rejects them, so admission must return task_unsupported. Submit execution concerns as factor types (bandwidth, battery, runtime, network_availability, device_state) and assert they are rejected as unknown types with task_unsupported, while the corresponding constraints in the execution object are accepted. Requirements: REQ-SAFETY-0015, REQ-SAFETY-0016, REQ-SAFETY-0017, REQ-TRUST-0036.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SAFETY-0015 -->
<!-- requirement-ref: REQ-SAFETY-0016 -->
<!-- requirement-ref: REQ-SAFETY-0017 -->
<!-- requirement-ref: REQ-TRUST-0036 -->

- [REQ-SAFETY-0015](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0015) → `CC-SAFETY-RISK-FACTORS-INITIAL-RISK-FACTOR-TYPES`.
- [REQ-SAFETY-0016](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0016) → `CC-SAFETY-RISK-FACTORS-INITIAL-RISK-FACTOR-TYPES`.
- [REQ-SAFETY-0017](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0017) → `CC-SAFETY-RISK-FACTORS-INITIAL-RISK-FACTOR-TYPES`.
- [REQ-TRUST-0036](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0036) → `CC-SAFETY-RISK-FACTORS-INITIAL-RISK-FACTOR-TYPES`.

### CC-SAFETY-RISK-FACTORS-RISK-CONSISTENCY

Public source: `data-formats/task-safety.md`, section **Risk Factors / Risk Consistency**. Intended target: `internal/validation/safety.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcSafetyRiskFactorsRiskConsistency`.

Case plan: Iterate overall risk_level low, medium and high against a single factor at each of low, medium and high. Accept every pair where the overall level is at least the factor level, and reject low with medium, low with high and medium with high (overall with factor) as task_schema_invalid. Mixed factors low, medium and high are accepted under overall high and rejected under overall medium. With a factor of level unknown and no trusted local assessment, accept only overall unknown and reject low, medium and high. With a fake trusted assessment that resolves the unknown factor to medium, overall medium is accepted only through that trusted input, and the same bytes are rejected once the assessment is removed. Overall unknown with only known factors is accepted. Requirements: REQ-SAFETY-0018, REQ-SAFETY-0019.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SAFETY-0018 -->
<!-- requirement-ref: REQ-SAFETY-0019 -->

- [REQ-SAFETY-0018](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0018) → `CC-SAFETY-RISK-FACTORS-RISK-CONSISTENCY`.
- [REQ-SAFETY-0019](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0019) → `CC-SAFETY-RISK-FACTORS-RISK-CONSISTENCY`.

### CC-SAFETY-USER-CONSENT

Public source: `data-formats/task-safety.md`, section **User Consent**. Intended target: `internal/validation/safety.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcSafetyUserConsent`.

Case plan: Combine task requires_user_consent true and false with a fake host consent provider reporting valid, absent, expired and unavailable consent, and with the local consent requirement off and on. Task true with valid consent: an attempt may start. Task true with absent, expired or unavailable consent: no attempt starts and no Measurement Result is produced. Task false with the local requirement on and absent consent: no attempt starts. Task false with the local requirement off: no consent requirement is added, and risk and target checks still apply. Admission records the consent requirement as required when the task, issuer permissions or local policy requires it, and a task value of false never clears an issuer or local requirement. Requirements: REQ-SAFETY-0020, REQ-SAFETY-0021, REQ-SAFETY-0022.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SAFETY-0020 -->
<!-- requirement-ref: REQ-SAFETY-0021 -->
<!-- requirement-ref: REQ-SAFETY-0022 -->

- [REQ-SAFETY-0020](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0020) → `CC-SAFETY-USER-CONSENT`.
- [REQ-SAFETY-0021](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0021) → `CC-SAFETY-USER-CONSENT`.
- [REQ-SAFETY-0022](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0022) → `CC-SAFETY-USER-CONSENT`.

### CC-SAFETY-EFFECTIVE-RISK

Public source: `data-formats/task-safety.md`, section **Effective Risk**. Intended target: `internal/validation/safety.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcSafetyEffectiveRisk`.

Case plan: Vary the registered baseline (low, medium), the task declaration (low, medium, high, unknown) and a fake local assessment (none, medium, high) independently. Assert effective risk equals the most restrictive of the three, with unknown dominating as unordered. Baseline medium with a low declaration is either rejected as invalid or admitted with effective risk of at least medium, and is never admitted as low. A high declaration on a low baseline gives effective risk high, and a local policy without high permission prevents execution. The admitted-task value exposes effective risk, and no task field lowers it. Requirements: REQ-SAFETY-0023, REQ-SAFETY-0024, REQ-SAFETY-0025, REQ-SAFETY-0026.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SAFETY-0023 -->
<!-- requirement-ref: REQ-SAFETY-0024 -->
<!-- requirement-ref: REQ-SAFETY-0025 -->
<!-- requirement-ref: REQ-SAFETY-0026 -->

- [REQ-SAFETY-0023](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0023) → `CC-SAFETY-EFFECTIVE-RISK`.
- [REQ-SAFETY-0024](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0024) → `CC-SAFETY-EFFECTIVE-RISK`.
- [REQ-SAFETY-0025](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0025) → `CC-SAFETY-EFFECTIVE-RISK`.
- [REQ-SAFETY-0026](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0026) → `CC-SAFETY-EFFECTIVE-RISK`.

### CC-SAFETY-SAFETY-EVALUATION

Public source: `data-formats/task-safety.md`, section **Safety Evaluation**. Intended target: `internal/validation/safety.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcSafetySafetyEvaluation`.

Case plan: Start from a task meeting all five public safety conditions and assert an attempt may start. Deny one condition at a time: an invalid declaration, an unsupported factor, effective risk not permitted by local policy, invalid required consent, and a fake host or device restriction. For each assert no attempt starts, no Measurement Result or bundle entry is produced, and the outcome is a non-measurement admission error or a local not-run state distinct from every network failure reason. Make the safety assessment unavailable (a fake local assessor returning an error) and assert execution fails closed. Give the task unknown fields and labels that try to relax local privacy, consent or safety settings and assert nothing is relaxed. Assert rejection outputs carry no censorship, blocking or interference interpretation. Admission rows cover the declaration, factor support and effective-risk conditions; consent and host or device conditions are execution-owned. Requirements: REQ-SAFETY-0027, REQ-SAFETY-0028, REQ-SAFETY-0029, REQ-SAFETY-0030.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SAFETY-0027 -->
<!-- requirement-ref: REQ-SAFETY-0028 -->
<!-- requirement-ref: REQ-SAFETY-0029 -->
<!-- requirement-ref: REQ-SAFETY-0030 -->

- [REQ-SAFETY-0027](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0027) → `CC-SAFETY-SAFETY-EVALUATION`.
- [REQ-SAFETY-0028](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0028) → `CC-SAFETY-SAFETY-EVALUATION`.
- [REQ-SAFETY-0029](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0029) → `CC-SAFETY-SAFETY-EVALUATION`.
- [REQ-SAFETY-0030](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-safety.md#REQ-SAFETY-0030) → `CC-SAFETY-SAFETY-EVALUATION`.

### CC-SCHEDULE-ENVELOPE

Public source: `data-formats/task-schedule.md`, section **Envelope**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleEnvelope`.

Case plan: Drive the scheduler with a deterministic clock and a fake platform that can report the application as suspended. Admit a NET-01 task with `starts_at` T and `not_after` T+24h. Keep the fake platform suspended from T-1h to T+2h and assert no attempt starts during suspension and the run is not marked failed or missed at T. On resume at T+2h assert the run becomes eligible and its attempt starts after T+2h, and assert the result `started_at` reflects the actual start, not `starts_at`. Repeat with suspension lasting until after `not_after` and assert no run or attempt starts after resume. Assert no API return value or status describes `starts_at` as a guaranteed or exact execution time. Requirements: REQ-SCHEDULE-0001.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0001 -->

- [REQ-SCHEDULE-0001](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0001) → `CC-SCHEDULE-ENVELOPE`.

### CC-SCHEDULE-FIELDS

Public source: `data-formats/task-schedule.md`, section **Fields**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleFields`.

Case plan: Exercise omitted/present and conditional schedule fields, defaults, wrong types and unknown properties.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0002 -->

- [REQ-SCHEDULE-0002](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0002) → `CC-SCHEDULE-FIELDS`.

### CC-SCHEDULE-SCHEDULING-MODEL

Public source: `data-formats/task-schedule.md`, section **Scheduling Model**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleSchedulingModel`.

Case plan: Use a deterministic clock and a synthetic task with repeat.mode interval, interval 21600, jitter 300, max_runs 3, a not_after after the third due run and retry.max_attempts 2 with retry_on internal_error and backoff mode fixed, initial_delay_ms 30000, max_delay_ms 30000. Script attempt 1 of run 1 to end with internal_error and every other attempt to complete. Assert three distinct run identities, run 1 holding attempts 1 and 2 and runs 2 and 3 holding attempt 1 only, the run counter reaching 3 and attempt numbering scoped to each run. Assert a retry attempt shares its run identity and never creates a new run, and that each interval-due run receives a new run identity. With repeat.mode none and the same retry policy, assert exactly one run whose attempt count is at most 2. Requirements: REQ-SCHEDULE-0003, REQ-SCHEDULE-0004, REQ-SCHEDULE-0005, REQ-SCHEDULE-0006.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0003 -->
<!-- requirement-ref: REQ-SCHEDULE-0004 -->
<!-- requirement-ref: REQ-SCHEDULE-0005 -->
<!-- requirement-ref: REQ-SCHEDULE-0006 -->

- [REQ-SCHEDULE-0003](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0003) → `CC-SCHEDULE-SCHEDULING-MODEL`.
- [REQ-SCHEDULE-0004](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0004) → `CC-SCHEDULE-SCHEDULING-MODEL`.
- [REQ-SCHEDULE-0005](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0005) → `CC-SCHEDULE-SCHEDULING-MODEL`.
- [REQ-SCHEDULE-0006](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0006) → `CC-SCHEDULE-SCHEDULING-MODEL`.

### CC-SCHEDULE-SCHEDULING-MODEL-START-TIME

Public source: `data-formats/task-schedule.md`, section **Scheduling Model / Start Time**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleSchedulingModelStartTime`.

Case plan: Test before/at/after starts_at plus permitted null/absence; do not run early or require exact timing.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0007 -->
<!-- requirement-ref: REQ-SCHEDULE-0008 -->
<!-- requirement-ref: REQ-SCHEDULE-0009 -->

- [REQ-SCHEDULE-0007](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0007) → `CC-SCHEDULE-SCHEDULING-MODEL-START-TIME`.
- [REQ-SCHEDULE-0008](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0008) → `CC-SCHEDULE-SCHEDULING-MODEL-START-TIME`.
- [REQ-SCHEDULE-0009](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0009) → `CC-SCHEDULE-SCHEDULING-MODEL-START-TIME`.

### CC-SCHEDULE-SCHEDULING-MODEL-TIME-LIMIT

Public source: `data-formats/task-schedule.md`, section **Scheduling Model / Time Limit**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleSchedulingModelTimeLimit`.

Case plan: Test acceptance/issuance-relative expiry at boundaries and the reviewed embedded one-time exception.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0010 -->
<!-- requirement-ref: REQ-SCHEDULE-0011 -->

- [REQ-SCHEDULE-0010](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0010) → `CC-SCHEDULE-SCHEDULING-MODEL-TIME-LIMIT`.
- [REQ-SCHEDULE-0011](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0011) → `CC-SCHEDULE-SCHEDULING-MODEL-TIME-LIMIT`.

### CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW

Public source: `data-formats/task-schedule.md`, section **Scheduling Model / Time Window**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleSchedulingModelTimeWindow`.

Case plan: Drive the scheduler with a deterministic clock and fake platform timezone UTC, admitting NET-01 tasks whose runs are otherwise eligible. With `time_window` `{"timezone": "local", "ranges": [{"start": "12:00", "end": "18:00"}]}`, advance the clock to local 11:59, 12:00, 17:59 and 18:00 and assert a run starts only at 12:00 and 17:59 (start inclusive, end exclusive) and not at 11:59 or 18:00. With the midnight-crossing range `22:00`-`06:00`, assert no start at 21:59 and 06:00 and a start at 22:00 and 05:59. With two ranges `08:00`-`10:00` and `14:00`-`16:00`, assert no start at 12:00 and a start at 14:00. Assert a run that becomes due outside the window is delayed to the next allowed range start before `not_after` rather than started early. Requirements: REQ-SCHEDULE-0012.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0012 -->

- [REQ-SCHEDULE-0012](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0012) → `CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW`.

### CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-FIELDS

Public source: `data-formats/task-schedule.md`, section **Scheduling Model / Time Window / Fields**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleSchedulingModelTimeWindowFields`.

Case plan: Validate synthetic `time_window` objects at admission and evaluate accepted ones with a deterministic clock and a fake platform timezone. Reject `ranges` omitted and `ranges: []` (0 ranges), accept 1 range and 4 non-overlapping ranges (`00:00`-`02:00`, `04:00`-`06:00`, `08:00`-`10:00`, `12:00`-`14:00`) and reject 5 ranges, with a bounded non-measurement admission error. Reject two identical ranges `08:00`-`12:00`, overlapping ranges `08:00`-`12:00` and `11:00`-`14:00`, and a midnight-crossing overlap `22:00`-`06:00` with `05:00`-`07:00`. Accept `timezone` omitted and assert it is treated as `local`. Reject `timezone` values `UTC`, `Europe/Berlin`, `""` and `Local` with a bounded non-measurement admission error and no run or attempt created. With range `12:00`-`18:00` and `timezone: local`, set the fake device timezone to UTC at fixed instant 10:30Z (local 10:30, no start), then switch it to UTC+03:00 at the same instant (local 13:30) and assert a run starts, demonstrating `local` uses the current device-reported timezone. Requirements: REQ-SCHEDULE-0013, REQ-SCHEDULE-0014, REQ-SCHEDULE-0043.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0013 -->
<!-- requirement-ref: REQ-SCHEDULE-0014 -->
<!-- requirement-ref: REQ-SCHEDULE-0043 -->

- [REQ-SCHEDULE-0013](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0013) → `CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-FIELDS`.
- [REQ-SCHEDULE-0014](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0014) → `CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-FIELDS`.
- [REQ-SCHEDULE-0043](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0043) → `CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-FIELDS`.

### CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-RANGE-FIELDS

Public source: `data-formats/task-schedule.md`, section **Scheduling Model / Time Window / Range Fields**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleSchedulingModelTimeWindowRangeFields`.

Case plan: Test HH:mm syntax, inclusive start/exclusive end, midnight expansion and overlapping ranges after expansion.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0015 -->
<!-- requirement-ref: REQ-SCHEDULE-0016 -->
<!-- requirement-ref: REQ-SCHEDULE-0017 -->

- [REQ-SCHEDULE-0015](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0015) → `CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-RANGE-FIELDS`.
- [REQ-SCHEDULE-0016](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0016) → `CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-RANGE-FIELDS`.
- [REQ-SCHEDULE-0017](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0017) → `CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-RANGE-FIELDS`.

### CC-SCHEDULE-SCHEDULING-MODEL-REPEAT

Public source: `data-formats/task-schedule.md`, section **Scheduling Model / Repeat**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleSchedulingModelRepeat`.

Case plan: Distinguish none and interval, disallow interval-only fields for none and generate the bounded number of runs.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0018 -->

- [REQ-SCHEDULE-0018](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0018) → `CC-SCHEDULE-SCHEDULING-MODEL-REPEAT`.

### CC-SCHEDULE-SCHEDULING-MODEL-REPEAT-INTERVAL-FIELDS

Public source: `data-formats/task-schedule.md`, section **Scheduling Model / Repeat / Interval Fields**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleSchedulingModelRepeatIntervalFields`.

Case plan: Exercise interval/jitter bounds and delayed due runs with injected random offsets; offsets are locally generated per run and stay inside start/expiry/window constraints.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0019 -->
<!-- requirement-ref: REQ-SCHEDULE-0020 -->
<!-- requirement-ref: REQ-SCHEDULE-0021 -->

- [REQ-SCHEDULE-0019](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0019) → `CC-SCHEDULE-SCHEDULING-MODEL-REPEAT-INTERVAL-FIELDS`.
- [REQ-SCHEDULE-0020](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0020) → `CC-SCHEDULE-SCHEDULING-MODEL-REPEAT-INTERVAL-FIELDS`.
- [REQ-SCHEDULE-0021](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0021) → `CC-SCHEDULE-SCHEDULING-MODEL-REPEAT-INTERVAL-FIELDS`.

### CC-SCHEDULE-SCHEDULING-MODEL-RETRY

Public source: `data-formats/task-schedule.md`, section **Scheduling Model / Retry**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleSchedulingModelRetry`.

Case plan: Keep run identity/max_runs accounting stable across the second attempt; a terminal negative measurement creates no automatic retry.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0022 -->
<!-- requirement-ref: REQ-SCHEDULE-0023 -->
<!-- requirement-ref: REQ-SCHEDULE-0024 -->
<!-- requirement-ref: REQ-SCHEDULE-0025 -->

- [REQ-SCHEDULE-0022](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0022) → `CC-SCHEDULE-SCHEDULING-MODEL-RETRY`.
- [REQ-SCHEDULE-0023](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0023) → `CC-SCHEDULE-SCHEDULING-MODEL-RETRY`.
- [REQ-SCHEDULE-0024](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0024) → `CC-SCHEDULE-SCHEDULING-MODEL-RETRY`.
- [REQ-SCHEDULE-0025](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0025) → `CC-SCHEDULE-SCHEDULING-MODEL-RETRY`.

### CC-SCHEDULE-SCHEDULING-MODEL-RETRY-RETRY-FIELDS

Public source: `data-formats/task-schedule.md`, section **Scheduling Model / Retry / Retry Fields**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleSchedulingModelRetryRetryFields`.

Case plan: Enumerate permitted and forbidden reasons; exercise attempts=1/2 conditional fields, fixed/exponential delays and jitter at expiry and delay bounds.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0026 -->
<!-- requirement-ref: REQ-SCHEDULE-0027 -->
<!-- requirement-ref: REQ-SCHEDULE-0028 -->
<!-- requirement-ref: REQ-SCHEDULE-0029 -->

- [REQ-SCHEDULE-0026](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0026) → `CC-SCHEDULE-SCHEDULING-MODEL-RETRY-RETRY-FIELDS`.
- [REQ-SCHEDULE-0027](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0027) → `CC-SCHEDULE-SCHEDULING-MODEL-RETRY-RETRY-FIELDS`.
- [REQ-SCHEDULE-0028](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0028) → `CC-SCHEDULE-SCHEDULING-MODEL-RETRY-RETRY-FIELDS`.
- [REQ-SCHEDULE-0029](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0029) → `CC-SCHEDULE-SCHEDULING-MODEL-RETRY-RETRY-FIELDS`.

### CC-SCHEDULE-SCHEDULING-MODEL-LIMITS-FIELDS

Public source: `data-formats/task-schedule.md`, section **Scheduling Model / Limits / Fields**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleSchedulingModelLimitsFields`.

Case plan: Decode synthetic schedules: omitted max_runs with repeat.mode none defaults to 1; max_runs 0 rejected, 1 accepted, 28 accepted, 29 rejected; non-integer max_runs rejected. With repeat.mode interval, omitted max_runs is rejected and max_runs 28 accepted. Execute with a deterministic clock, repeat.mode interval (interval 21600, jitter 300), max_runs 2, a valid not_after and retry.max_attempts 2 with retry_on platform_suspended and backoff mode fixed, initial_delay_ms 30000, max_delay_ms 30000, scripting a retry in run 1: assert two runs complete and the retry does not count toward max_runs. Configure max_runs 28 with a not_after reached after 2 runs, and separately max_runs 2 with a later not_after, and assert the schedule terminates at whichever limit is reached first. Requirements: REQ-SCHEDULE-0030, REQ-SCHEDULE-0031, REQ-SCHEDULE-0032.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0030 -->
<!-- requirement-ref: REQ-SCHEDULE-0031 -->
<!-- requirement-ref: REQ-SCHEDULE-0032 -->

- [REQ-SCHEDULE-0030](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0030) → `CC-SCHEDULE-SCHEDULING-MODEL-LIMITS-FIELDS`.
- [REQ-SCHEDULE-0031](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0031) → `CC-SCHEDULE-SCHEDULING-MODEL-LIMITS-FIELDS`.
- [REQ-SCHEDULE-0032](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0032) → `CC-SCHEDULE-SCHEDULING-MODEL-LIMITS-FIELDS`.

### CC-SCHEDULE-VALIDATION

Public source: `data-formats/task-schedule.md`, section **Validation**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleValidation`.

Case plan: Mutate each listed schedule constraint; reject expired/invalid schedules, and redeliver identical tasks without extending expiry or adding schedules.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0033 -->
<!-- requirement-ref: REQ-SCHEDULE-0036 -->
<!-- requirement-ref: REQ-SCHEDULE-0037 -->
<!-- requirement-ref: REQ-SCHEDULE-0038 -->
<!-- requirement-ref: REQ-SCHEDULE-0039 -->
<!-- requirement-ref: REQ-SCHEDULE-0040 -->
<!-- requirement-ref: REQ-SCHEDULE-0041 -->
<!-- requirement-ref: REQ-SCHEDULE-0042 -->

- [REQ-SCHEDULE-0033](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0033) → `CC-SCHEDULE-VALIDATION`.
- [REQ-SCHEDULE-0036](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0036) → `CC-SCHEDULE-VALIDATION`.
- [REQ-SCHEDULE-0037](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0037) → `CC-SCHEDULE-VALIDATION`.
- [REQ-SCHEDULE-0038](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0038) → `CC-SCHEDULE-VALIDATION`.
- [REQ-SCHEDULE-0039](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0039) → `CC-SCHEDULE-VALIDATION`.
- [REQ-SCHEDULE-0040](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0040) → `CC-SCHEDULE-VALIDATION`.
- [REQ-SCHEDULE-0041](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0041) → `CC-SCHEDULE-VALIDATION`.
- [REQ-SCHEDULE-0042](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0042) → `CC-SCHEDULE-VALIDATION`.

### CC-SCHEDULE-VALIDATION-ELIGIBILITY-RULES

Public source: `data-formats/task-schedule.md`, section **Validation / Eligibility Rules**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleValidationEligibilityRules`.

Case plan: Deny each eligibility predicate independently for initial and retry attempts; assert no early/out-of-window/expired execution and preserve stricter local veto.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0034 -->
<!-- requirement-ref: REQ-SCHEDULE-0035 -->

- [REQ-SCHEDULE-0034](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0034) → `CC-SCHEDULE-VALIDATION-ELIGIBILITY-RULES`.
- [REQ-SCHEDULE-0035](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task-schedule.md#REQ-SCHEDULE-0035) → `CC-SCHEDULE-VALIDATION-ELIGIBILITY-RULES`.

### CC-TASK-OVERVIEW

Public source: `data-formats/task.md`, section **Overview**. Intended target: `internal/taskdecode/decode.go`. Planned case location: `internal/taskdecode/cross_cutting_conformance_test.go`, `symbol:TestCcTaskOverview`.

Case plan: Submit synthetic tasks to admission with a recording result sink and deterministic clock: a valid NET-01 task, a task with invalid `input`, a task with an expired `not_after`, and a task carrying an extra top-level `results` member containing a canary result-like object. Assert the valid task is admitted and that admission alone writes zero Measurement Results to the sink; assert each invalid task returns a bounded non-measurement admission error and also writes zero Measurement Results. Assert the canary result content never appears in any later Measurement Result, log or status output, and that the only Measurement Result for the valid task is produced after its execution attempt. Requirements: REQ-TASK-0002.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-TASK-0002 -->

- [REQ-TASK-0002](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task.md#REQ-TASK-0002) → `CC-TASK-OVERVIEW`.

### CC-TASK-ENVELOPE

Public source: `data-formats/task.md`, section **Envelope**. Intended target: `internal/taskdecode/decode.go`. Planned case location: `internal/taskdecode/cross_cutting_conformance_test.go`, `symbol:TestCcTaskEnvelope`.

Case plan: Decode a minimal supported task, then omit each required envelope field, add an unknown field, violate task identifier syntax/byte bounds and exceed the complete encoded-byte bound. Check campaign-wide identity is preserved and no result is decoded as task input.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-TASK-0001 -->
<!-- requirement-ref: REQ-TASK-0003 -->
<!-- requirement-ref: REQ-TASK-0004 -->
<!-- requirement-ref: REQ-TASK-0005 -->

- [REQ-TASK-0001](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task.md#REQ-TASK-0001) → `CC-TASK-ENVELOPE`.
- [REQ-TASK-0003](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task.md#REQ-TASK-0003) → `CC-TASK-ENVELOPE`.
- [REQ-TASK-0004](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task.md#REQ-TASK-0004) → `CC-TASK-ENVELOPE`.
- [REQ-TASK-0005](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task.md#REQ-TASK-0005) → `CC-TASK-ENVELOPE`.

### CC-TASK-DESIGN-PRINCIPLES-STABLE-STRUCTURE

Public source: `data-formats/task.md`, section **Design Principles / Stable Structure**. Intended target: `docs/conformance/host-delivery-review.md`. Planned case location: `docs/conformance/host-delivery-review.md`, `heading:RV-CC-TASK-DESIGN-PRINCIPLES-STABLE-STRUCTURE`.

Case plan: no automated case; review case `RV-CC-TASK-DESIGN-PRINCIPLES-STABLE-STRUCTURE`: Human review case; not automated coverage. Reviewer role: Specification maintainer reviewer (independent of the change author). Reviewed artifact: Proposed task schema changes and their specification diffs at the reviewed specification revision. Checklist: (1) Does the proposal add a new top-level task field? (2) If a new top-level field is proposed, does the change record why the extension cannot be placed inside an existing nested object (`input`, `options`, `schedule`, `execution`, `safety` or `labels`)? (3) Are extensions that fit an existing nested object placed there? (4) Is the change frequency of the top-level schema kept rare, with the change recorded in the specification change history? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

Scope: Host, collector or specification review; no runtime enforcement claim. No execution evidence or support claim.

<!-- requirement-ref: REQ-TASK-0006 -->

- [REQ-TASK-0006](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/task.md#REQ-TASK-0006) → `RV-CC-TASK-DESIGN-PRINCIPLES-STABLE-STRUCTURE`.

### CC-CONTEXT-OVERVIEW

Public source: `library/network-context.md`, section **Overview**. Intended target: `internal/networkcontext/acquire.go`. Case location: `internal/execution/egress_test.go`, `symbol:TestCcContextOverview`.

Case plan: Run multiple tests on one generation/path; acquire shared bounded context rather than independently repeating per test.

Scope: Synthetic core contract cases. Executed by DEV-647: three runs on one access-network generation and path share one endpoint request through the shared egress cache, against synthetic HTTPS endpoints. Passed at Go `d788a80` ([evidence](../traceability/evidence/go-unit-d788a80.md#CC-CONTEXT-OVERVIEW)); partial evidence only. No native, field or support claim.

<!-- requirement-ref: REQ-CONTEXT-0001 -->

- [REQ-CONTEXT-0001](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0001) → `CC-CONTEXT-OVERVIEW`.

### CC-CONTEXT-CONTEXT-LAYERS-ACCESS-CONTEXT

Public source: `library/network-context.md`, section **Context Layers / Access Context**. Intended target: `internal/networkcontext/acquire.go`. Case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextContextLayersAccessContext`.

Case plan: Feed the Go core access context through a fake platform that also offers canary values for every forbidden item: interface name, local IPv4 and IPv6 addresses, DNS server address, SSID, BSSID, cell identifier, SIM information, subscriber identifier, carrier account information and physical location. Iterate `network_type` over `wifi`, `cellular`, `ethernet`, `other` and `unknown`, `vpn_state` over `detected`, `not_detected` and `unknown`, `ipv4_available`, `ipv6_available` and `validated` over `yes`, `no` and `unknown`, and `network_changed` over `true` and `false`. Assert the serialized access object contains exactly these six fields with the supplied normalized values. Assert serialized output of the access object, the Measurement Result, logs, cache entries and egress-endpoint requests contains none of the canary values. Reject an access input carrying any additional field. Requirements: REQ-CONTEXT-0002, REQ-CONTEXT-0003.

Scope: Synthetic core contract cases. Passed at Go `cdbd160` with the test-only fake platform network ([evidence](../traceability/evidence/go-unit-cdbd160.md#CC-CONTEXT-CONTEXT-LAYERS-ACCESS-CONTEXT)); the core writes no log; since DEV-647 the egress cache holds classifications keyed only by generations and the path name, and the endpoint request carries only the request line, `Host` and `Accept` (`TestCcContextEndpointContract` in `internal/execution`), so no access value reaches either. Partial evidence only, with no native capability or support claim.

<!-- requirement-ref: REQ-CONTEXT-0002 -->
<!-- requirement-ref: REQ-CONTEXT-0003 -->

- [REQ-CONTEXT-0002](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0002) → `CC-CONTEXT-CONTEXT-LAYERS-ACCESS-CONTEXT`.
- [REQ-CONTEXT-0003](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0003) → `CC-CONTEXT-CONTEXT-LAYERS-ACCESS-CONTEXT`.

### CC-CONTEXT-CONTEXT-LAYERS-EGRESS-CONTEXT

Public source: `library/network-context.md`, section **Context Layers / Egress Context**. Intended target: `internal/networkcontext/acquire.go`. Case location: `internal/execution/egress_test.go`, `symbol:TestCcContextContextLayersEgressContext`.

Case plan: Run synthetic attempts against a local synthetic classification endpoint that records the observed source address and returns `{"country_code": "NL", "region_code": "NL-NH", "asn": 64500, "classification_version": "0.1.0"}`. For a `direct` attempt, assert the egress object contains `status`, `lookup_source`, `network_path: direct`, the classification fields and `classification_version`, and no `connection_code`. For a `proxy` attempt through a scripted loopback CONNECT proxy with `connection_code` `outline_mobileproxy`, assert the lookup goes through the proxy and the egress object includes that `connection_code`. Configure the endpoint to additionally return a field carrying the observed source address and assert the response is rejected and egress reports `status: unavailable`. Assert the recorded source address never appears in serialized results, cache keys, logs or status output. Requirements: REQ-CONTEXT-0004, REQ-CONTEXT-0005, REQ-CONTEXT-0006.

Scope: Synthetic core contract cases. Executed by DEV-647 against a synthetic HTTPS endpoint that records the observed source address, on the direct path and through the scripted loopback CONNECT proxy: the classification and `connection_code` rows, rejection of a response returning an address, and absence of the observed address from results. The core writes no log, and cache keys hold only generations and the path name. The review case stays open and non-gating. Passed at Go `d788a80` ([evidence](../traceability/evidence/go-unit-d788a80.md#CC-CONTEXT-CONTEXT-LAYERS-EGRESS-CONTEXT)); partial evidence only. No native, field or support claim.

Review case `RV-CC-CONTEXT-CONTEXT-LAYERS-EGRESS-CONTEXT` (`docs/conformance/cross-cutting-review.md`, `heading:RV-CC-CONTEXT-CONTEXT-LAYERS-EGRESS-CONTEXT`): Human review case; not automated coverage. Reviewer role: Privacy reviewer and measurement interpretation reviewer (independent of the implementer and endpoint operator). Reviewed artifact: Egress-classification endpoint operating documentation and Go result documentation and analyst guidance for egress context at the reviewed revision. Checklist: (1) Does the endpoint operator documentation state that the source IP is classified without being returned and that its retention is minimized or avoided? (2) Is the endpoint logging configuration consistent with that minimization statement? (3) Does result and analyst guidance describe country and region as IP classification, not the user's physical location? (4) Does guidance state that ASN describes the observed egress network and may not identify the access operator, especially with VPN, proxy, carrier-grade NAT or circumvention paths? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-CONTEXT-0004 -->
<!-- requirement-ref: REQ-CONTEXT-0005 -->
<!-- requirement-ref: REQ-CONTEXT-0006 -->

- [REQ-CONTEXT-0004](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0004) → `CC-CONTEXT-CONTEXT-LAYERS-EGRESS-CONTEXT`, `RV-CC-CONTEXT-CONTEXT-LAYERS-EGRESS-CONTEXT`.
- [REQ-CONTEXT-0005](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0005) → `CC-CONTEXT-CONTEXT-LAYERS-EGRESS-CONTEXT`, `RV-CC-CONTEXT-CONTEXT-LAYERS-EGRESS-CONTEXT`.
- [REQ-CONTEXT-0006](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0006) → `CC-CONTEXT-CONTEXT-LAYERS-EGRESS-CONTEXT`, `RV-CC-CONTEXT-CONTEXT-LAYERS-EGRESS-CONTEXT`.

### CC-CONTEXT-HOST-INITIALIZATION

Public source: `library/network-context.md`, section **Host Initialization**. Intended target: `internal/networkcontext/acquire.go`. Case location: `network_context_test.go`, `symbol:TestCcContextHostInitialization`.

Case plan: Validate endpoint URLs/fields/independent approvals, absence fallback to access-only, immutable configuration and rejection of redirects/discovery/credentials/per-user URLs.

Scope: Synthetic core contract cases. Executed by DEV-647 through the root facade with synthetic adapters: approved endpoints, access-only context and no endpoint contact without configuration, rejected unapproved, non-HTTPS, credential, fragment, loopback and unknown-field configurations, no public method to change the endpoints, and a redirect that is not followed. Endpoint syntax and approval are also covered by `internal/config` tests. Passed at Go `d788a80` ([evidence](../traceability/evidence/go-unit-d788a80.md#CC-CONTEXT-HOST-INITIALIZATION)); partial evidence only. No native, field or support claim.

<!-- requirement-ref: REQ-CONTEXT-0007 -->
<!-- requirement-ref: REQ-CONTEXT-0008 -->
<!-- requirement-ref: REQ-CONTEXT-0009 -->
<!-- requirement-ref: REQ-CONTEXT-0010 -->

- [REQ-CONTEXT-0007](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0007) → `CC-CONTEXT-HOST-INITIALIZATION`.
- [REQ-CONTEXT-0008](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0008) → `CC-CONTEXT-HOST-INITIALIZATION`.
- [REQ-CONTEXT-0009](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0009) → `CC-CONTEXT-HOST-INITIALIZATION`.
- [REQ-CONTEXT-0010](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0010) → `CC-CONTEXT-HOST-INITIALIZATION`.

### CC-CONTEXT-ENDPOINT-CONTRACT

Public source: `library/network-context.md`, section **Endpoint Contract**. Intended target: `internal/networkcontext/acquire.go`. Case location: `internal/execution/egress_test.go`, `symbol:TestCcContextEndpointContract`.

Case plan: Capture one fixed HTTPS GET without identifier/target/category/cookie/auth data. Validate exact response schema, 4096-byte/5000ms limits, forbidden fields and successful unknown classifications.

Scope: Synthetic core contract cases. Executed by DEV-647: the request side (one `GET` with only `Host` and `Accept`, asserted on the decrypted bytes) in `internal/execution`, and the response side (strict schema, forbidden fields, a version-only response, 4,096 and 4,097 bytes, 4,999 and 5,001 ms on the engine clock) in `TestCcContextEndpointContract` of `internal/networkcontext/egress_test.go`; HTTP, TLS and size failures run in `TestContextEndpointResponses`. Passed at Go `d788a80` ([evidence](../traceability/evidence/go-unit-d788a80.md#CC-CONTEXT-ENDPOINT-CONTRACT)); partial evidence only. No native, field or support claim.

<!-- requirement-ref: REQ-CONTEXT-0011 -->
<!-- requirement-ref: REQ-CONTEXT-0012 -->
<!-- requirement-ref: REQ-CONTEXT-0013 -->
<!-- requirement-ref: REQ-CONTEXT-0014 -->
<!-- requirement-ref: REQ-CONTEXT-0015 -->
<!-- requirement-ref: REQ-CONTEXT-0016 -->

- [REQ-CONTEXT-0011](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0011) → `CC-CONTEXT-ENDPOINT-CONTRACT`.
- [REQ-CONTEXT-0012](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0012) → `CC-CONTEXT-ENDPOINT-CONTRACT`.
- [REQ-CONTEXT-0013](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0013) → `CC-CONTEXT-ENDPOINT-CONTRACT`.
- [REQ-CONTEXT-0014](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0014) → `CC-CONTEXT-ENDPOINT-CONTRACT`.
- [REQ-CONTEXT-0015](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0015) → `CC-CONTEXT-ENDPOINT-CONTRACT`.
- [REQ-CONTEXT-0016](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0016) → `CC-CONTEXT-ENDPOINT-CONTRACT`.

### CC-CONTEXT-PATH-CONSISTENCY

Public source: `library/network-context.md`, section **Path Consistency**. Intended target: `internal/networkcontext/acquire.go`. Case location: `internal/execution/egress_test.go`, `symbol:TestCcProxyNetworkContext`.

Case plan: Run direct/proxy/NET-05 with generation changes; assert one matching snapshot and no direct request or cache substitution during proxy-only execution.

Scope: Planned as a whole. `TestCcProxyNetworkContext` covers direct and proxy lookups on their own paths, separate direct and proxy cache entries, a proxy generation per key and no direct request during proxy-only execution; `TestCcContextExecutionIntegration` covers a generation change during acquisition. The row requiring all NET-05 steps to share one context and path snapshot needs the NET-05 runner (DEV-454). Mixed-family evidence is not promoted. No support claim.

<!-- requirement-ref: REQ-CONTEXT-0017 -->
<!-- requirement-ref: REQ-CONTEXT-0018 -->
<!-- requirement-ref: REQ-CONTEXT-0019 -->

- [REQ-CONTEXT-0017](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0017) → `CC-CONTEXT-PATH-CONSISTENCY`.
- [REQ-CONTEXT-0018](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0018) → `CC-CONTEXT-PATH-CONSISTENCY`.
- [REQ-CONTEXT-0019](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0019) → `CC-CONTEXT-PATH-CONSISTENCY`.

### CC-CONTEXT-ACQUISITION-AND-FALLBACK

Public source: `library/network-context.md`, section **Acquisition and Fallback**. Intended target: `internal/networkcontext/acquire.go`. Case location: `internal/networkcontext/egress_test.go`, `symbol:TestCcContextAcquisitionAndFallback`.

Case plan: Count primary/fallback calls, test valid-but-empty primary, malformed/TLS/HTTP/timeout/size failures, exact cache-key hits/misses, freshness and every invalidation event. Verify no persistent cache.

Scope: Synthetic core contract cases. Executed by DEV-647: request counts, a valid primary without classifications, malformed, unapproved, oversized and failing responses, exact cache-key hits and misses, reuse at 300 and not at 301 seconds and every invalidation event with a scripted fetch and the injected clock; TLS, HTTP and timeout failures against synthetic HTTPS endpoints in `TestContextEndpointResponses`; shutdown, reinitialization, policy replacement, clearing and proxy changes through the facade in `TestEgressCacheLifecycle` and `TestEgressCacheProxyChange`. The cache is in memory only, with no persistence port. Passed at Go `d788a80` ([evidence](../traceability/evidence/go-unit-d788a80.md#CC-CONTEXT-ACQUISITION-AND-FALLBACK)); partial evidence only. No native, field or support claim.

<!-- requirement-ref: REQ-CONTEXT-0043 -->
<!-- requirement-ref: REQ-CONTEXT-0021 -->
<!-- requirement-ref: REQ-CONTEXT-0022 -->
<!-- requirement-ref: REQ-CONTEXT-0023 -->
<!-- requirement-ref: REQ-CONTEXT-0024 -->

- [REQ-CONTEXT-0043](https://github.com/tv-rain/barometer-spec/blob/bb805f07a36630db7250d5ab63006ed5e351f7a7/library/network-context.md#REQ-CONTEXT-0043) → `CC-CONTEXT-ACQUISITION-AND-FALLBACK`.
- [REQ-CONTEXT-0021](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0021) → `CC-CONTEXT-ACQUISITION-AND-FALLBACK`.
- [REQ-CONTEXT-0022](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0022) → `CC-CONTEXT-ACQUISITION-AND-FALLBACK`.
- [REQ-CONTEXT-0023](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0023) → `CC-CONTEXT-ACQUISITION-AND-FALLBACK`.
- [REQ-CONTEXT-0024](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0024) → `CC-CONTEXT-ACQUISITION-AND-FALLBACK`.

### CC-CONTEXT-EXECUTION-INTEGRATION

Public source: `library/network-context.md`, section **Execution Integration**. Intended target: `internal/networkcontext/acquire.go`. Case location: `internal/execution/egress_test.go`, `symbol:TestCcContextExecutionIntegration`.

Case plan: Inject cancellation/deadline/suspension/network changes during acquisition; distinguish missing required context, allowlist mismatch and optional egress failure. Requirements: REQ-CONTEXT-0025, REQ-CONTEXT-0026, REQ-CONTEXT-0027, REQ-CONTEXT-0028.

Scope: Planned as a whole. `TestCcContextExecutionIntegration` covers host cancellation, the attempt deadline and a network change during acquisition, a missing required field, an allowlist mismatch, an unknown country, optional egress failure and retries. Its former admission rejection row and the REQ-TRUST-0036 mapping were removed because egress acquisition exists since DEV-647 and admission admits these tasks (`UT.internal.admission.TestGateAdmitsEgressContextRequirements`). The suspension-during-acquisition row needs a platform suspension signal (DEV-455). Mixed-family evidence is not promoted. No support claim.

<!-- requirement-ref: REQ-CONTEXT-0025 -->
<!-- requirement-ref: REQ-CONTEXT-0026 -->
<!-- requirement-ref: REQ-CONTEXT-0027 -->
<!-- requirement-ref: REQ-CONTEXT-0028 -->

- [REQ-CONTEXT-0025](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0025) → `CC-CONTEXT-EXECUTION-INTEGRATION`.
- [REQ-CONTEXT-0026](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0026) → `CC-CONTEXT-EXECUTION-INTEGRATION`.
- [REQ-CONTEXT-0027](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0027) → `CC-CONTEXT-EXECUTION-INTEGRATION`.
- [REQ-CONTEXT-0028](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0028) → `CC-CONTEXT-EXECUTION-INTEGRATION`.

### CC-CONTEXT-PROVIDER-AND-OPERATOR-INTERPRETATION

Public source: `library/network-context.md`, section **Provider and Operator Interpretation**. Intended target: `internal/networkcontext/acquire.go`. Planned case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextProviderAndOperatorInterpretation`.

Case plan: Configure a fake platform with canary device locale, timezone, language, SIM country, application store region and user-profile values, and a synthetic classification endpoint. With the endpoint returning only `{"classification_version": "0.1.0"}`, assert the egress object omits `country_code`, `region_code` and `asn` and no provider or operator value appears anywhere in serialized output, so nothing is inferred from the canaries. With the endpoint returning `asn` 64500, assert the ASN and `classification_version` are retained and no free-form provider name field is produced. Configure the endpoint to also return a free-form operator name and assert the response is rejected. Iterate `network_type` `wifi` and `cellular` and assert no mobile operator or SIM carrier field is collected. Requirements: REQ-CONTEXT-0029, REQ-CONTEXT-0030, REQ-CONTEXT-0031.

Scope: Synthetic core contract cases. No execution evidence or support claim.

Review case `RV-CC-CONTEXT-PROVIDER-AND-OPERATOR-INTERPRETATION` (`docs/conformance/cross-cutting-review.md`, `heading:RV-CC-CONTEXT-PROVIDER-AND-OPERATOR-INTERPRETATION`): Human review case; not automated coverage. Reviewer role: Measurement interpretation reviewer (analyst role independent of the implementer). Reviewed artifact: Analysis documentation describing ASN-to-organization mapping at the reviewed revision. Checklist: (1) Does analysis map ASN to an organization only using the recorded `classification_version` or another documented dataset? (2) Is the mapping dataset and its version documented for each analysis? (3) Is a mapped organization presented as contextual rather than as a verified access-provider identity? (4) Does any analysis derive provider or operator values from locale, timezone, language, SIM country, application store or user profile? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-CONTEXT-0029 -->
<!-- requirement-ref: REQ-CONTEXT-0030 -->
<!-- requirement-ref: REQ-CONTEXT-0031 -->

- [REQ-CONTEXT-0029](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0029) → `CC-CONTEXT-PROVIDER-AND-OPERATOR-INTERPRETATION`, `RV-CC-CONTEXT-PROVIDER-AND-OPERATOR-INTERPRETATION`.
- [REQ-CONTEXT-0030](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0030) → `CC-CONTEXT-PROVIDER-AND-OPERATOR-INTERPRETATION`, `RV-CC-CONTEXT-PROVIDER-AND-OPERATOR-INTERPRETATION`.
- [REQ-CONTEXT-0031](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0031) → `CC-CONTEXT-PROVIDER-AND-OPERATOR-INTERPRETATION`, `RV-CC-CONTEXT-PROVIDER-AND-OPERATOR-INTERPRETATION`.

### CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS

Public source: `library/network-context.md`, section **Privacy and Safety Requirements**. Intended target: `internal/networkcontext/acquire.go`. Planned case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextPrivacyAndSafetyRequirements`.

Case plan: Run synthetic attempts with canary installation, account, device and session identifiers, `task_id`, `test_id`, measurement target hostname, `connection_code`, cookies and a local address present in host and task state, against a recording classification endpoint. Assert each request is a single `GET` to the configured fixed URL with `Accept: application/json` and contains none of the canaries in path, query, headers or body, and carries no `Cookie` or `Authorization` header. Configure the endpoint to return a field carrying the observed IP address and assert the response is rejected. Feed fake platform canaries for GPS location, Wi-Fi identifiers, interface names, cell identifiers, SIM identifiers, phone numbers, subscriber identifiers and carrier account data and assert none appear in any serialized output. Assert a response without `region_code` is accepted with `region_code` absent. Assert that when `asn` is present no provider name field is emitted. Inject endpoint failure (unreachable primary and fallback) and assert egress reports `status: unavailable` with `lookup_source: none`, and the result carries no censorship-indicating outcome attributable to the failure. Requirements: REQ-CONTEXT-0032, REQ-CONTEXT-0035, REQ-CONTEXT-0036, REQ-CONTEXT-0037, REQ-CONTEXT-0038, REQ-CONTEXT-0039, REQ-CONTEXT-0040, REQ-CONTEXT-0041, REQ-CONTEXT-0042.

Scope: Synthetic core contract cases. No execution evidence or support claim.

Review case `RV-CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS` (`docs/conformance/cross-cutting-review.md`, `heading:RV-CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`): Human review case; not automated coverage. Reviewer role: Privacy and safety reviewer (independent of the implementer and endpoint operator). Reviewed artifact: Host deployment configuration, endpoint privacy review records, endpoint operator logging configuration and public-reporting methodology at the reviewed revision. Checklist: (1) Are the configured primary and fallback endpoints each covered by a recorded privacy review? (2) Does the endpoint operator disable or minimize raw access logs and avoid combining source IP and request time with unrelated user or application data? (3) Does the deployed endpoint return only coarse classifications and never the observed IP address? (4) Is region collection enabled only where its necessity and reliability are documented? (5) Does public reporting aggregate context into sufficiently broad groups rather than raw combinations of region, ASN, network type, connection code and time? (6) Does reporting avoid treating endpoint or classification failure as evidence of censorship? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-CONTEXT-0032 -->
<!-- requirement-ref: REQ-CONTEXT-0035 -->
<!-- requirement-ref: REQ-CONTEXT-0036 -->
<!-- requirement-ref: REQ-CONTEXT-0037 -->
<!-- requirement-ref: REQ-CONTEXT-0038 -->
<!-- requirement-ref: REQ-CONTEXT-0039 -->
<!-- requirement-ref: REQ-CONTEXT-0040 -->
<!-- requirement-ref: REQ-CONTEXT-0041 -->
<!-- requirement-ref: REQ-CONTEXT-0042 -->

- [REQ-CONTEXT-0032](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0032) → `CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-CONTEXT-0035](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0035) → `CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-CONTEXT-0036](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0036) → `CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-CONTEXT-0037](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0037) → `CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-CONTEXT-0038](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0038) → `CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-CONTEXT-0039](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0039) → `CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-CONTEXT-0040](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0040) → `CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-CONTEXT-0041](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0041) → `CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-CONTEXT-0042](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0042) → `CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`.

### CC-CONTEXT-PLATFORM-COLLECTION

Public source: `library/network-context.md`, section **Platform Collection**. Intended target: `internal/networkcontext/acquire.go`. Planned case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextPlatformCollection`.

Case plan: Test normalized coarse Android/Apple adapter outputs and unsupported=unknown behavior; reject raw platform objects or sensitive transport details.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-CONTEXT-0033 -->
<!-- requirement-ref: REQ-CONTEXT-0034 -->

- [REQ-CONTEXT-0033](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0033) → `CC-CONTEXT-PLATFORM-COLLECTION`.
- [REQ-CONTEXT-0034](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/network-context.md#REQ-CONTEXT-0034) → `CC-CONTEXT-PLATFORM-COLLECTION`.

### CC-PROXY-OVERVIEW

Public source: `library/proxy.md`, section **Overview**. Intended target: `internal/execution/proxy.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcProxyOverview`.

Case plan: Configure a scripted local proxy listener on `127.0.0.1` and accept the configuration `{"host": "127.0.0.1", "port": 54321, "connection_code": "outline_mobileproxy"}` with no test identifiers or test capability declaration. Run a TCP-01 proxy attempt and assert the library opens one TCP connection to the loopback endpoint and sends an HTTP/1.1 `CONNECT` request line for the target authority, and nothing else is used as proxy transport. Reject configurations adding a transport selector such as `"type": "socks5"`, a declared supported-tests list, or a non-loopback `host` such as `192.0.2.1`, leaving the previous configuration unchanged. Script the listener to reply with a SOCKS handshake reply or a non-HTTP/1.1 status line and assert the attempt fails per the public outcome table without direct fallback. Reject a DNS-01 task with `network_path: proxy` using UDP transport, which needs a non-stream capability. Requirements: REQ-PROXY-0001.

Its configuration, CONNECT and reply rows run in `TestCcProxyOverview` with the production CONNECT dialer against the scripted loopback proxy. The DNS-01 row runs in `dns01_conformance_test.go:TestCcProxyOverviewDNSRow` with the production codec, registry and proxy rule: UDP/system proxy tasks reject without contact and direct TCP runs through one metered CONNECT tunnel. The family has executable synthetic coverage; its manifest awaits the milestone update. The CONNECT dialer also refuses a non-stream dial without contact (`TestConnectAuthority`).

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-PROXY-0001 -->

- [REQ-PROXY-0001](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0001) → `CC-PROXY-OVERVIEW`.

### CC-PROXY-HOST-INTEGRATION

Public source: `library/proxy.md`, section **Host Integration**. Intended target: `proxy.go`. Case location: `proxy_lifecycle_test.go`, `symbol:TestProxyHostConfigurationStaysSeparate`.

Case plan: Exercise the Go binding-facing initialization, replacement and clearing entry points using only ordinary serializable values: initialize with a proxy object {host "127.0.0.1", port 54321, connection_code "outline_mobileproxy"}, replace it with {host "127.0.0.1", port 55001, connection_code "outline_mobileproxy"}, then clear it. Assert after each call that the active configuration equals exactly the supplied object (or no proxy after clearing) and that a recording fake proxy listener receives new proxy-task tunnels only on the port of the current configuration. Run a concurrent reader during replacement and assert it observes either the complete previous or the complete new object, never a mixed host/port/connection_code combination, so replacement is atomic. Assert the exported initialization, setProxy-equivalent and clearProxy-equivalent signatures take only strings, integers and maps/JSON bytes, with no callback, interface, socket or object parameter. Initialize with both a proxy object and a separate top-level network_context object {primary_endpoint, fallback_endpoint} and assert both are configured independently; a proxy object carrying primary_endpoint or fallback_endpoint is rejected, and replacing or clearing the proxy leaves the configured egress endpoints unchanged. Requirements: REQ-PROXY-0002, REQ-PROXY-0003, REQ-PROXY-0004.

Scope: Planned as a whole. `TestProxyHostConfigurationStaysSeparate` covers the serializable Go configuration row and separation of egress endpoints; `TestPublicSurfaceMatchesDocumentedSignatures` checks the Go signatures. Android, Apple and Flutter method mapping and independent binding review remain DEV-455. Mixed-family evidence is not promoted from the Go subset. No binding or support claim.

Review case `RV-CC-PROXY-HOST-INTEGRATION` (`docs/conformance/cross-cutting-review.md`, `heading:RV-CC-PROXY-HOST-INTEGRATION`): Human review case; not automated coverage. Reviewer role: Mobile binding reviewer (independent of the binding implementer). Reviewed artifact: Android, Apple and Flutter binding source and integration documentation for proxy initialization, replacement and clearing at the reviewed revision. Checklist: (1) Do the Android, Apple and Flutter bindings accept the proxy only as ordinary serializable values (host, port, connection_code)? (2) Is the host free of any requirement to supply native callbacks, interfaces, sockets or complex objects to configure, replace or clear the proxy? (3) Does each binding's method naming map to initialization, atomic replacement and clearing with semantics equivalent to the Go core contract? (4) Do the bindings and documentation configure the egress-classification endpoints separately at initialization rather than as properties of the proxy object? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-PROXY-0002 -->
<!-- requirement-ref: REQ-PROXY-0003 -->
<!-- requirement-ref: REQ-PROXY-0004 -->

- [REQ-PROXY-0002](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0002) → `CC-PROXY-HOST-INTEGRATION`, `RV-CC-PROXY-HOST-INTEGRATION`.
- [REQ-PROXY-0003](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0003) → `CC-PROXY-HOST-INTEGRATION`, `RV-CC-PROXY-HOST-INTEGRATION`.
- [REQ-PROXY-0004](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0004) → `CC-PROXY-HOST-INTEGRATION`, `RV-CC-PROXY-HOST-INTEGRATION`.

### CC-PROXY-PROXY-CONFIGURATION

Public source: `library/proxy.md`, section **Proxy Configuration**. Intended target: `internal/execution/proxy.go`. Case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcProxyProxyConfiguration`.

Case plan: Exercise loopback/localhost, port and connection_code boundaries and forbidden fields/transports; reject non-loopback endpoints.

Scope: Synthetic core contract cases. Runs in `TestCcProxyProxyConfiguration` with synthetic runners, the production CONNECT dialer and the scripted loopback proxy of `internal/testkit/loopback`. Passed at Go `d788a80` ([evidence](../traceability/evidence/go-unit-d788a80.md#CC-PROXY-PROXY-CONFIGURATION)); partial evidence only. No native capability, production runner or support claim.

<!-- requirement-ref: REQ-PROXY-0005 -->
<!-- requirement-ref: REQ-PROXY-0006 -->
<!-- requirement-ref: REQ-PROXY-0007 -->

- [REQ-PROXY-0005](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0005) → `CC-PROXY-PROXY-CONFIGURATION`.
- [REQ-PROXY-0006](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0006) → `CC-PROXY-PROXY-CONFIGURATION`.
- [REQ-PROXY-0007](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0007) → `CC-PROXY-PROXY-CONFIGURATION`.

### CC-PROXY-CONFIGURATION-VALIDATION

Public source: `library/proxy.md`, section **Configuration Validation**. Intended target: `proxy.go`. Case location: `proxy_lifecycle_test.go`, `symbol:TestCcProxyConfigurationValidation`.

Case plan: Attempt malformed replacement while an old proxy is active; preserve old state. Test category stability/privacy and internally derived compatibility.

Scope: Synthetic core contract cases. The standalone replacement cases run with synthetic adapters and the production strict configuration decoder in `TestCcProxyConfigurationValidation`: invalid input leaves both the endpoint and the active run unchanged, approved coarse categories survive endpoint renewal, and task capabilities cannot be supplied by proxy configuration. Registry-derived compatibility is also covered by `internal/admission/proxy_test.go`. Passed at Go `d788a80` ([evidence](../traceability/evidence/go-unit-d788a80.md#CC-PROXY-CONFIGURATION-VALIDATION)); partial evidence only. No native, runtime-support or field claim.

<!-- requirement-ref: REQ-PROXY-0008 -->
<!-- requirement-ref: REQ-PROXY-0009 -->
<!-- requirement-ref: REQ-PROXY-0010 -->

- [REQ-PROXY-0008](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0008) → `CC-PROXY-CONFIGURATION-VALIDATION`.
- [REQ-PROXY-0009](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0009) → `CC-PROXY-CONFIGURATION-VALIDATION`.
- [REQ-PROXY-0010](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0010) → `CC-PROXY-CONFIGURATION-VALIDATION`.

### CC-PROXY-LIFECYCLE-AND-RENEWAL

Public source: `library/proxy.md`, section **Lifecycle and Renewal**. Intended target: `proxy.go`. Case location: `proxy_lifecycle_test.go`, `symbol:TestProxyChangesOnlyCancelProxyRuns`.

Case plan: Race set/clear with attempts and all NET-05 steps; block old-generation admission, cancel/close and bound cleanup before returning, invalidate caches and never mix generations or use direct fallback.

Scope: Planned as a whole. `TestProxyChangesOnlyCancelProxyRuns`, `TestProxyLifecycleSerialization`, `TestProxyCleanupIncomplete`, `TestProxyStartRace`, `TestProxyAgainstLifecycleRace` and `TestProxyChangeDrainsPathReservation` cover the Go cancellation, serialization and cleanup rows; `TestProxyLoopbackRenewal` supplies the two-endpoint socket scenario and `TestProxyLoopbackClosesEstablishedTunnel` covers a tunnel blocked on application I/O; both passed with loopback sockets in the Go `d788a80` milestone run. Whole-profile NET-05 steps remain DEV-454; egress cache invalidation on set and clear is covered by `TestEgressCacheProxyChange` (DEV-647). Mixed-family evidence is not promoted from the Go subset. No support claim.

<!-- requirement-ref: REQ-PROXY-0011 -->
<!-- requirement-ref: REQ-PROXY-0012 -->
<!-- requirement-ref: REQ-PROXY-0013 -->
<!-- requirement-ref: REQ-PROXY-0014 -->
<!-- requirement-ref: REQ-PROXY-0015 -->
<!-- requirement-ref: REQ-PROXY-0016 -->

- [REQ-PROXY-0011](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0011) → `CC-PROXY-LIFECYCLE-AND-RENEWAL`.
- [REQ-PROXY-0012](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0012) → `CC-PROXY-LIFECYCLE-AND-RENEWAL`.
- [REQ-PROXY-0013](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0013) → `CC-PROXY-LIFECYCLE-AND-RENEWAL`.
- [REQ-PROXY-0014](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0014) → `CC-PROXY-LIFECYCLE-AND-RENEWAL`.
- [REQ-PROXY-0015](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0015) → `CC-PROXY-LIFECYCLE-AND-RENEWAL`.
- [REQ-PROXY-0016](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0016) → `CC-PROXY-LIFECYCLE-AND-RENEWAL`.

### CC-PROXY-HTTP-CONNECT-BEHAVIOR

Public source: `library/proxy.md`, section **HTTP CONNECT Behavior**. Intended target: `internal/execution/proxy.go`. Case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcProxyHttpConnectBehavior`.

Case plan: Capture synthetic CONNECT exchanges for explicit IP and hostname cases, 2xx/non-2xx/malformed/timeouts, redirects and cancellation. Assert exact authority, no credentials/cleartext target GET, fresh tunnels and unchanged TLS semantics.

Scope: Synthetic core contract cases. Runs in `TestCcProxyHttpConnectBehavior` with synthetic runners, the production CONNECT dialer and the scripted loopback proxy of `internal/testkit/loopback`. Passed at Go `d788a80` ([evidence](../traceability/evidence/go-unit-d788a80.md#CC-PROXY-HTTP-CONNECT-BEHAVIOR)); partial evidence only. No native capability, production runner or support claim.

<!-- requirement-ref: REQ-PROXY-0017 -->
<!-- requirement-ref: REQ-PROXY-0018 -->
<!-- requirement-ref: REQ-PROXY-0043 -->
<!-- requirement-ref: REQ-PROXY-0020 -->

- [REQ-PROXY-0017](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0017) → `CC-PROXY-HTTP-CONNECT-BEHAVIOR`.
- [REQ-PROXY-0018](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0018) → `CC-PROXY-HTTP-CONNECT-BEHAVIOR`.
- [REQ-PROXY-0043](https://github.com/tv-rain/barometer-spec/blob/bb805f07a36630db7250d5ab63006ed5e351f7a7/library/proxy.md#REQ-PROXY-0043) → `CC-PROXY-HTTP-CONNECT-BEHAVIOR`.
- [REQ-PROXY-0020](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0020) → `CC-PROXY-HTTP-CONNECT-BEHAVIOR`.

### CC-PROXY-TASK-SELECTION

Public source: `library/proxy.md`, section **Task Selection**. Intended target: `internal/execution/proxy.go`. Case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcProxyTaskSelection`.

Case plan: Exercise direct default and proxy preconditions; task cannot carry endpoint, category or proxy identity.

Scope: Synthetic core contract cases. Runs in `TestCcProxyTaskSelection` with synthetic runners, the production CONNECT dialer and the scripted loopback proxy of `internal/testkit/loopback`. Passed at Go `d788a80` ([evidence](../traceability/evidence/go-unit-d788a80.md#CC-PROXY-TASK-SELECTION)); partial evidence only. No native capability, production runner or support claim.

<!-- requirement-ref: REQ-PROXY-0021 -->
<!-- requirement-ref: REQ-PROXY-0022 -->

- [REQ-PROXY-0021](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0021) → `CC-PROXY-TASK-SELECTION`.
- [REQ-PROXY-0022](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0022) → `CC-PROXY-TASK-SELECTION`.

### CC-PROXY-NETWORK-CONTEXT

Public source: `library/proxy.md`, section **Network Context**. Intended target: `internal/execution/proxy.go`. Case location: `internal/execution/egress_test.go`, `symbol:TestCcProxyNetworkContext`.

Case plan: Fail primary context endpoint and verify same-generation proxied fallback, exact cache keys and required-context gating with zero direct requests.

Scope: Synthetic core contract cases. Executed by DEV-647 against synthetic HTTPS endpoints through the scripted loopback CONNECT proxy: a failing primary and a same-generation proxied fallback, exact cache keys per proxy generation, required-context gating when the proxy refuses the lookup, and zero direct requests; `TestEgressCacheProxyChange` covers proxy replacement through the facade. Passed at Go `d788a80` ([evidence](../traceability/evidence/go-unit-d788a80.md#CC-PROXY-NETWORK-CONTEXT)); partial evidence only. No native, field or support claim.

<!-- requirement-ref: REQ-PROXY-0045 -->
<!-- requirement-ref: REQ-PROXY-0024 -->
<!-- requirement-ref: REQ-PROXY-0025 -->

- [REQ-PROXY-0045](https://github.com/tv-rain/barometer-spec/blob/bb805f07a36630db7250d5ab63006ed5e351f7a7/library/proxy.md#REQ-PROXY-0045) → `CC-PROXY-NETWORK-CONTEXT`.
- [REQ-PROXY-0024](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0024) → `CC-PROXY-NETWORK-CONTEXT`.
- [REQ-PROXY-0025](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0025) → `CC-PROXY-NETWORK-CONTEXT`.

### CC-PROXY-INITIAL-TEST-COMPATIBILITY

Public source: `library/proxy.md`, section **Initial Test Compatibility**. Intended target: `internal/execution/proxy.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcProxyInitialTestCompatibility`.

Case plan: Parameterize all nine registered tests and relevant input/options combinations, including each incompatible NET-05 step, on network_path: proxy. Assert that every incompatible task is rejected at admission step 8 with task_schema_invalid (REQ-TRUST-0039) before path authority and target registration, atomically, with no compatible subset run and zero transport, DNS and Network Context calls, and that no incompatible task is admitted to end in proxy_path_unsupported (REQ-PROXY-0042). The same parameters on network_path: direct are not rejected for compatibility, and each compatible combination passes the check. Requirements: REQ-PROXY-0026, REQ-PROXY-0027, REQ-TRUST-0039, REQ-PROXY-0042.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-PROXY-0044 -->
<!-- requirement-ref: REQ-PROXY-0027 -->
<!-- requirement-ref: REQ-TRUST-0039 -->
<!-- requirement-ref: REQ-PROXY-0046 -->

- [REQ-PROXY-0044](https://github.com/tv-rain/barometer-spec/blob/bb805f07a36630db7250d5ab63006ed5e351f7a7/library/proxy.md#REQ-PROXY-0044) → `CC-PROXY-INITIAL-TEST-COMPATIBILITY`.
- [REQ-PROXY-0027](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0027) → `CC-PROXY-INITIAL-TEST-COMPATIBILITY`.
- [REQ-TRUST-0039](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0039) → `CC-PROXY-INITIAL-TEST-COMPATIBILITY`.
- [REQ-PROXY-0046](https://github.com/tv-rain/barometer-spec/blob/bb805f07a36630db7250d5ab63006ed5e351f7a7/library/proxy.md#REQ-PROXY-0046) → `CC-PROXY-INITIAL-TEST-COMPATIBILITY`.

### CC-PROXY-EXECUTION-STATES

Public source: `library/proxy.md`, section **Execution States**. Intended target: `internal/execution/proxy.go`. Case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcProxyExecutionStates`.

Case plan: Trigger each path failure and assert normalized state, no censorship attribution, no direct retry, and any new generation needs a separately scheduled attempt.

Scope: Synthetic core contract cases have executable tests for all pre-attempt, CONNECT failure and reconfiguration states, no direct fallback, and a separate run on a new endpoint in `TestCcProxyExecutionStates`. Root integration is exercised by `TestProxyChangesOnlyCancelProxyRuns` and `TestProxyLoopbackRenewal`. Its loopback rows passed in the milestone run. Passed at Go `d788a80` ([evidence](../traceability/evidence/go-unit-d788a80.md#CC-PROXY-EXECUTION-STATES)); partial evidence only. No native or support claim.

<!-- requirement-ref: REQ-PROXY-0028 -->
<!-- requirement-ref: REQ-PROXY-0029 -->
<!-- requirement-ref: REQ-PROXY-0030 -->
<!-- requirement-ref: REQ-PROXY-0046 -->

- [REQ-PROXY-0028](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0028) → `CC-PROXY-EXECUTION-STATES`.
- [REQ-PROXY-0029](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0029) → `CC-PROXY-EXECUTION-STATES`.
- [REQ-PROXY-0030](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0030) → `CC-PROXY-EXECUTION-STATES`.
- [REQ-PROXY-0046](https://github.com/tv-rain/barometer-spec/blob/bb805f07a36630db7250d5ab63006ed5e351f7a7/library/proxy.md#REQ-PROXY-0046) → `CC-PROXY-EXECUTION-STATES`.

### CC-PROXY-RESULT-METADATA

Public source: `library/proxy.md`, section **Result Metadata**. Intended target: `internal/execution/proxy.go`. Case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcProxyResultMetadata`.

Case plan: Use a scripted transport and a fake local HTTP CONNECT proxy configured with synthetic canary values (host "127.0.0.1", a distinctive canary port, connection_code "outline_mobileproxy") to produce four attempts: direct success, direct failure, proxied success after a scripted 2xx CONNECT response, and proxied failures recorded as proxy_connect_failed (scripted non-2xx CONNECT response carrying canary response headers and a canary body) and proxy_unavailable (closed listener port). Assert every proxied result records network_context.egress.network_path "proxy" and connection_code "outline_mobileproxy", including when no network_context endpoints are configured and egress status is unavailable, and every direct result records network_path "direct" with connection_code absent. Assert the serialized results and bundle contain no proxy host or port field or value, no proxy generation value, no CONNECT request line or authority echo, none of the canary proxy response headers or body, and no Outline transport configuration canary passed to the host fake; assert no CONNECT status class field is present unless the pinned result schema defines one. Requirements: REQ-PROXY-0031, REQ-PROXY-0032, REQ-PROXY-0033.

Scope: Synthetic core contract cases. Runs in `TestCcProxyResultMetadata` with synthetic runners, the production CONNECT dialer and the scripted loopback proxy of `internal/testkit/loopback`. Passed at Go `d788a80` ([evidence](../traceability/evidence/go-unit-d788a80.md#CC-PROXY-RESULT-METADATA)); partial evidence only. No native capability, production runner or support claim.

<!-- requirement-ref: REQ-PROXY-0031 -->
<!-- requirement-ref: REQ-PROXY-0032 -->
<!-- requirement-ref: REQ-PROXY-0033 -->

- [REQ-PROXY-0031](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0031) → `CC-PROXY-RESULT-METADATA`.
- [REQ-PROXY-0032](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0032) → `CC-PROXY-RESULT-METADATA`.
- [REQ-PROXY-0033](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0033) → `CC-PROXY-RESULT-METADATA`.

### CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION

Public source: `library/proxy.md`, section **Privacy, Safety, and Interpretation**. Intended target: `internal/execution/proxy.go`. Case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcProxyPrivacySafetyAndInterpretation`.

Case plan: Submit synthetic tasks that attempt to choose or alter the proxy endpoint through execution, input, options, labels or unknown fields (proxy, host, port, endpoint, connection_code) and assert each is rejected at admission with a bounded admission error before any network activity, while the host-configured proxy remains unchanged. Configure a valid proxy with trusted local policy that disallows proxy use and assert a network_path "proxy" task ends not_run with reason proxy_not_permitted and opens no connection. Drive proxy_unavailable, proxy_connect_failed and proxy_reconfigured with a scripted proxy and a recording direct transport and assert zero direct dials to the target and no automatic retry through the direct path. Iterate loopback, private, link-local and other targets prohibited by target validation on both direct and proxy paths and assert identical rejection; assert connection_code in results is copied byte-for-byte from host configuration, never extended with session, installation or endpoint material, and cannot be set by a task. Produce a direct failure followed by a proxied success for the same synthetic target and assert the library emits two independent results with no derived comparison, censorship or path-verdict field. Requirements: REQ-PROXY-0034, REQ-PROXY-0035, REQ-PROXY-0036, REQ-PROXY-0037, REQ-PROXY-0038, REQ-PROXY-0039, REQ-PROXY-0040, REQ-PROXY-0041.

Scope: Planned as a whole. `TestProxyTaskCannotReplaceHostConfiguration` covers task-controlled endpoint fields and preservation of the host proxy; `TestProxyChangesOnlyCancelProxyRuns` covers reconfiguration result minimization. `TestCcProxyResultMetadata` and `TestCcProxyExecutionStates` cover minimized loopback results and no fallback. The production execution/path gate has synthetic target-validation evidence in `TestTargetGateResolutionAndRegistration`, `TestTargetGateProxyRefusesHostnameBeforeStartOrResolution`, `TestTargetGateRedirectFullTupleAndSharedScope`, `TestCcProxyProductionTargetRefusal` and `TestCcProxyProductionAllowedTarget` and `TestCcDirectProductionAllowedTarget`. `TestProxyContextHostnameEligibility` and `TestProxyHostnameContextUnavailable` cover the narrowed context boundary. Registered HTTP/NET-05 codec/runner integration and independent interpretation review remain open (DEV-454); gate tests do not implement those runners. Mixed-family evidence is not promoted from these Go rows. No support claim.

<!-- requirement-ref: REQ-PROXY-0043 -->
<!-- requirement-ref: REQ-PROXY-0046 -->
<!-- requirement-ref: REQ-HTTP-01-0051 -->

The shared explicit-IP and typed redirect checks above map the successor contract to this existing family; the HTTP observation/parser and NET-05 whole-profile rows remain with their production runners.

Review case `RV-CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION` (`docs/conformance/cross-cutting-review.md`, `heading:RV-CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION`): Human review case; not automated coverage. Reviewer role: Measurement interpretation reviewer (analyst role independent of the implementer). Reviewed artifact: Go result documentation and host/analyst guidance for direct versus proxied path comparison at the reviewed implementation revision. Checklist: (1) Does the guidance state that a successful proxied result does not prove the direct path is censored? (2) Does the guidance list outages, DNS-path differences, endpoint changes, proxy behavior and transient network conditions as alternative explanations for a direct failure followed by proxied success? (3) Does the guidance require repeated observations, stable network context, equivalent test inputs and independent supporting signals before any stronger interpretation? (4) Does the guidance state that hostname CONNECT is unavailable in the current contract, that this causes deliberate missingness, and that rewriting a hostname URL to an IP URL cannot manufacture an equivalent comparison? (5) Does host integration guidance require connection_code values coarse enough not to identify a user, account, server, session or installation? (6) Does the guidance avoid describing local-policy proxy rejections or proxy execution states as evidence of censorship? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-PROXY-0034 -->
<!-- requirement-ref: REQ-PROXY-0035 -->
<!-- requirement-ref: REQ-PROXY-0036 -->
<!-- requirement-ref: REQ-PROXY-0037 -->
<!-- requirement-ref: REQ-PROXY-0038 -->
<!-- requirement-ref: REQ-PROXY-0039 -->
<!-- requirement-ref: REQ-PROXY-0040 -->
<!-- requirement-ref: REQ-PROXY-0041 -->

- [REQ-PROXY-0034](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0034) → `CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION`, `RV-CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION`.
- [REQ-PROXY-0035](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0035) → `CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION`, `RV-CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION`.
- [REQ-PROXY-0036](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0036) → `CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION`, `RV-CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION`.
- [REQ-PROXY-0037](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0037) → `CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION`, `RV-CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION`.
- [REQ-PROXY-0038](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0038) → `CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION`, `RV-CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION`.
- [REQ-PROXY-0039](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0039) → `CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION`, `RV-CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION`.
- [REQ-PROXY-0040](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0040) → `CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION`, `RV-CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION`.
- [REQ-PROXY-0041](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/proxy.md#REQ-PROXY-0041) → `CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION`, `RV-CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION`.

### CC-LIMIT-PURPOSE

Public source: `library/registered-limits.md`, section **Purpose**. Intended target: `internal/validation/limits.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcLimitPurpose`.

Case plan: Registry and stricter values: with no local override, assert a task requesting max_bytes_sent 65,536 is accepted with that exact effective value and 65,537 is rejected at admission; with trusted local policy setting a stricter max_bytes_sent of 16,384, assert 16,384 is accepted unchanged and 16,385 is rejected, and repeat for max_duration_ms (60,000 accepted, 60,001 rejected) and max_connections (12 accepted, 13 rejected). Assert no rejected task is accepted with a reduced (clamped) effective value and no accepted task's effective limits differ from the requested values. Counted boundary (REQ-LIMIT-0033, REQ-LIMIT-0034): bytes count where the attempt exchanges test traffic with a TCP stream, the plaintext side of a TLS session or a UDP datagram payload, including HTTP request lines, headers and bodies, response status lines and headers, DNS messages and each DNS retransmission sent, DNS-over-TCP and DoT two-byte length prefixes, the complete DoH exchange and bytes read ahead into a buffer; TLS records, TCP/IP and UDP/IP headers, retransmissions, proxy transport framing, the CONNECT exchange with the loopback proxy endpoint and Network Context acquisition are not counted. Use a recording transport that reports bytes at the counted boundary and below it (TLS records, CONNECT exchange) as separate counters. With default max_bytes_sent 32,768, assert 32,768 counted bytes plus 40,000 uncounted bytes completes without a byte-limit stop, a write that would make 32,769 counted bytes is not issued and stops the attempt with reason resource_limit, and 1,000 counted bytes plus 100,000 uncounted bytes is not stopped; repeat for max_bytes_received at the default 131,072 (131,072 accepted, a further read stops it, and no read requests more than the remaining budget). For HTTP-01, NET-03, DNS-01, DNS-03 and DNS-04 runners, assert that request and response headers, DNS messages and length prefixes are included in the counted bytes. Assert only the counted-boundary counter is compared with the limit. Requirements: REQ-LIMIT-0001, REQ-LIMIT-0002, REQ-LIMIT-0033, REQ-LIMIT-0034.

DNS rows execute in `dns01_conformance_test.go:TestCcLimitPurposeDNSRows`, `TestCcProxyOverviewDNSRow` and `TestDns01SystemAPIOutputAndHiddenBytes`: UDP payloads, TCP DNS messages and both length prefixes count; exact budgets pass, a one-byte short budget stops as `resource_limit`, the CONNECT exchange stays outside the meter, and hidden system messages are excluded. Other protocol callers and review keep this whole family planned; the milestone manifest is unchanged.

Scope: Synthetic core contract cases. No execution evidence or support claim.

Review case `RV-CC-LIMIT-PURPOSE` (`docs/conformance/cross-cutting-review.md`, `heading:RV-CC-LIMIT-PURPOSE`): Human review case; not automated coverage. Reviewer role: Specification and measurement-safety reviewer (independent of the implementer). Reviewed artifact: Go limit-accounting documentation and per-test byte-accounting mapping at the reviewed implementation revision. Checklist: (1) For each registered test, does the documented mapping route every counted byte (HTTP request lines and headers, response status lines and headers, DNS messages, DoH framing, DoT and DNS-over-TCP length prefixes, TLS plaintext, UDP payloads) through the metered attempt scope, consistently with the counted boundary in registered-limits.md? (2) Is the proxy CONNECT exchange with the loopback proxy endpoint kept below the counted boundary, while CONNECT tunnels count toward max_connections? (3) Does the documentation avoid presenting byte limits as exact on-wire limits and document buffering below the counted boundary? (4) Where a public test specification defines a value stricter than the registry, does the implementation document and apply the stricter value? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-LIMIT-0001 -->
<!-- requirement-ref: REQ-LIMIT-0002 -->
<!-- requirement-ref: REQ-LIMIT-0033 -->
<!-- requirement-ref: REQ-LIMIT-0034 -->

- [REQ-LIMIT-0001](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0001) → `CC-LIMIT-PURPOSE`, `RV-CC-LIMIT-PURPOSE`.
- [REQ-LIMIT-0002](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0002) → `CC-LIMIT-PURPOSE`, `RV-CC-LIMIT-PURPOSE`.
- [REQ-LIMIT-0033](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/library/registered-limits.md#REQ-LIMIT-0033) → `CC-LIMIT-PURPOSE`, `RV-CC-LIMIT-PURPOSE`. Successor of the retired REQ-LIMIT-0003.
- [REQ-LIMIT-0034](https://github.com/tv-rain/barometer-spec/blob/5d01879cfd1d47396c780befa2567f7e90439948/library/registered-limits.md#REQ-LIMIT-0034) → `CC-LIMIT-PURPOSE`, `RV-CC-LIMIT-PURPOSE`. Successor of the retired REQ-LIMIT-0003.

### CC-LIMIT-TASK-AND-IDENTIFIER-LIMITS

Public source: `library/registered-limits.md`, section **Task and Identifier Limits**. Intended target: `internal/validation/limits.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcLimitTaskAndIdentifierLimits`.

Case plan: Parameterize each row of the task and identifier table at below, equal and above its maximum, measured on original encoded bytes: complete task 65,535 and 65,536 bytes accepted, 65,537 rejected before parsing as task_limit_exceeded; task_id 127 and 128 ASCII bytes accepted, 129 rejected as task_limit_exceeded, never task_schema_invalid; input JSON 49,151 and 49,152 encoded bytes accepted, 49,153 rejected; options JSON 8,191 and 8,192 accepted, 8,193 rejected; labels with 16 entries accepted and 17 rejected, and 2,048 encoded bytes accepted and 2,049 rejected; a generic identifier such as a NET-05 step_id of 64 ASCII bytes accepted and 65 rejected. Pad with insignificant whitespace and with escape sequences so encoded and decoded lengths differ, and assert the encoded length is counted. An unknown field that takes a task from 65,536 to 65,537 bytes is rejected as task_limit_exceeded; an unknown field within the limit is rejected as task_schema_invalid. Feed 1 MiB through a counting reader and assert no more than 65,537 bytes are read. Requirements: REQ-LIMIT-0004, REQ-LIMIT-0005.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LIMIT-0004 -->
<!-- requirement-ref: REQ-LIMIT-0005 -->

- [REQ-LIMIT-0004](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0004) → `CC-LIMIT-TASK-AND-IDENTIFIER-LIMITS`.
- [REQ-LIMIT-0005](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0005) → `CC-LIMIT-TASK-AND-IDENTIFIER-LIMITS`.

### CC-LIMIT-ATTEMPT-LIMITS

Public source: `library/registered-limits.md`, section **Attempt Limits**. Intended target: `internal/execution/budget.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcLimitAttemptLimits`.

Case plan: Exercise every default/maximum and NET-01 zero-traffic exception, 15000ms protocol ceiling, complete-attempt deadline and 2000ms cleanup budget with fake clocks/transports.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LIMIT-0006 -->
<!-- requirement-ref: REQ-LIMIT-0007 -->
<!-- requirement-ref: REQ-LIMIT-0008 -->

- [REQ-LIMIT-0006](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0006) → `CC-LIMIT-ATTEMPT-LIMITS`.
- [REQ-LIMIT-0007](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0007) → `CC-LIMIT-ATTEMPT-LIMITS`.
- [REQ-LIMIT-0008](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0008) → `CC-LIMIT-ATTEMPT-LIMITS`.

### CC-LIMIT-SCHEDULE-AND-RETRY-LIMITS

Public source: `library/registered-limits.md`, section **Schedule and Retry Limits**. Intended target: `internal/validation/schedule.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcLimitScheduleAndRetryLimits`.

Case plan: Parameterize every bound/default, issuance skew and trusted-time absence, embedded one-time exception and forbidden measurement-outcome retries.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LIMIT-0009 -->
<!-- requirement-ref: REQ-LIMIT-0010 -->
<!-- requirement-ref: REQ-LIMIT-0011 -->
<!-- requirement-ref: REQ-LIMIT-0012 -->
<!-- requirement-ref: REQ-LIMIT-0013 -->

- [REQ-LIMIT-0009](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0009) → `CC-LIMIT-SCHEDULE-AND-RETRY-LIMITS`.
- [REQ-LIMIT-0010](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0010) → `CC-LIMIT-SCHEDULE-AND-RETRY-LIMITS`.
- [REQ-LIMIT-0011](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0011) → `CC-LIMIT-SCHEDULE-AND-RETRY-LIMITS`.
- [REQ-LIMIT-0012](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0012) → `CC-LIMIT-SCHEDULE-AND-RETRY-LIMITS`.
- [REQ-LIMIT-0013](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0013) → `CC-LIMIT-SCHEDULE-AND-RETRY-LIMITS`.

### CC-LIMIT-DNS-LIMITS

Public source: `library/registered-limits.md`, section **DNS Limits**. Intended target: `internal/protocol/dns/limits.go`. Planned case location: `internal/protocol/dns/cross_cutting_conformance_test.go`, `symbol:TestCcLimitDnsLimits`.

Case plan: Parameterize every DNS byte/count/window/answer/CNAME bound; report limit outcome without partial append or prohibited raw DNS fields.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LIMIT-0014 -->
<!-- requirement-ref: REQ-LIMIT-0015 -->

- [REQ-LIMIT-0014](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0014) → `CC-LIMIT-DNS-LIMITS`.
- [REQ-LIMIT-0015](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0015) → `CC-LIMIT-DNS-LIMITS`.

### CC-LIMIT-HTTP-AND-VALIDATION-LIMITS

Public source: `library/registered-limits.md`, section **HTTP and Validation Limits**. Intended target: `internal/protocol/http/limits.go`. Planned case location: `internal/protocol/http/cross_cutting_conformance_test.go`, `symbol:TestCcLimitHttpAndValidationLimits`.

Case plan: Parameterize URL/header/body/check/matcher/redirect/ALPN bounds and identity encoding; test undecidable bounded checks and NET-03 redirect non-following.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LIMIT-0016 -->
<!-- requirement-ref: REQ-LIMIT-0017 -->
<!-- requirement-ref: REQ-LIMIT-0018 -->

- [REQ-LIMIT-0016](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0016) → `CC-LIMIT-HTTP-AND-VALIDATION-LIMITS`.
- [REQ-LIMIT-0017](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0017) → `CC-LIMIT-HTTP-AND-VALIDATION-LIMITS`.
- [REQ-LIMIT-0018](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0018) → `CC-LIMIT-HTTP-AND-VALIDATION-LIMITS`.

### CC-LIMIT-NET-05-COMPOSITE-LIMITS

Public source: `library/registered-limits.md`, section **NET-05 Composite Limits**. Intended target: `internal/validation/profile.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcLimitNet05CompositeLimits`.

Case plan: Parameterize profile/step/identity/aggregate budgets, nested limits and sum-of-declared-maxima overflow; reject the full profile before traffic.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LIMIT-0019 -->
<!-- requirement-ref: REQ-LIMIT-0020 -->

- [REQ-LIMIT-0019](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0019) → `CC-LIMIT-NET-05-COMPOSITE-LIMITS`.
- [REQ-LIMIT-0020](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0020) → `CC-LIMIT-NET-05-COMPOSITE-LIMITS`.

### CC-LIMIT-RESULT-AND-BUNDLE-LIMITS

Public source: `library/registered-limits.md`, section **Result and Bundle Limits**. Intended target: `internal/bundle/bounds.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcLimitResultAndBundleLimits`.

Case plan: Exercise every encoded/count/pending-storage maximum; split bundles or return bounded persistence error without truncating valid results.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LIMIT-0021 -->
<!-- requirement-ref: REQ-LIMIT-0022 -->

- [REQ-LIMIT-0021](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0021) → `CC-LIMIT-RESULT-AND-BUNDLE-LIMITS`.
- [REQ-LIMIT-0022](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0022) → `CC-LIMIT-RESULT-AND-BUNDLE-LIMITS`.

### CC-LIMIT-CONTEXT-LINKAGE-AND-RETENTION-LIMITS

Public source: `library/registered-limits.md`, section **Context, Linkage, and Retention Limits**. Intended target: `internal/persistence/retention.go`. Planned case location: `internal/persistence/cross_cutting_conformance_test.go`, `symbol:TestCcLimitContextLinkageAndRetentionLimits`.

Case plan: Use a deterministic clock, fake Network Context endpoints, fake private-state storage and a recording uploader. Network Context endpoint response: 4,096 bytes accepted, 4,097 bytes not accepted as context (fallback or unavailable egress per network-context.md). Endpoint operation: response at 4,999 ms accepted, at 5,001 ms the operation is stopped. In-memory context freshness: with the default 300 seconds, cached context reused at 300 seconds and not reused at 301 seconds; where freshness is configurable, 900 seconds accepted and 901 rejected. Linkage rotation period: 86,399 rejected, 86,400 and 604,800 accepted, 604,801 rejected, and absent configuration yields no linkage. Unacknowledged bundle retention: with 86,400 seconds, a pending bundle is uploadable before expiry and, once older than 86,400 seconds, is deleted and the recording uploader sees no upload attempt; a configured retention of 86,401 is rejected. Accepted-task replay state: for a task with validity V, the replay record exists until V plus 86,400 seconds and is gone afterwards, and never exists beyond 691,200 seconds. Durable acknowledgement: after acknowledgement processing returns, assert the acknowledged bundle is absent from storage and that no canary result bytes remain in any library-owned storage. Requirements: REQ-LIMIT-0023, REQ-LIMIT-0024, REQ-LIMIT-0025.

Scope: Planned as a whole. The context-endpoint rows run since DEV-647: 4,096 and 4,097 bytes and 4,999 and 5,001 ms in `TestCcContextEndpointContract` (`internal/networkcontext/egress_test.go`) and `TestContextEndpointResponses` (`internal/execution/egress_test.go`), and reuse at 300 and not at 301 seconds in `TestCcContextAcquisitionAndFallback`; the configurable-freshness row does not apply because this build's freshness is fixed at the default. The pending-bundle retention rows are covered by the DEV-563 pending-store tests. The linkage rows remain open (DEV-454). Mixed-family evidence is not promoted. No support claim.

Review case `RV-CC-LIMIT-CONTEXT-LINKAGE-AND-RETENTION-LIMITS` (`docs/conformance/cross-cutting-review.md`, `heading:RV-CC-LIMIT-CONTEXT-LINKAGE-AND-RETENTION-LIMITS`): Human review case; not automated coverage. Reviewer role: Deployment privacy reviewer (independent of the deployment operator). Reviewed artifact: Deployment retention documentation for collector, host analytics, crash reporting and research storage at the reviewed revision. Checklist: (1) Does each of collector, host analytics, crash reporting and research storage have an explicit documented retention period? (2) Does any documented period rely on the 86,400-second local bundle retention or the 691,200-second replay-state cap as permission to retain data for that duration or longer? (3) Is each documented period justified independently of the library's local maxima? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-LIMIT-0023 -->
<!-- requirement-ref: REQ-LIMIT-0024 -->
<!-- requirement-ref: REQ-LIMIT-0025 -->

- [REQ-LIMIT-0023](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0023) → `CC-LIMIT-CONTEXT-LINKAGE-AND-RETENTION-LIMITS`, `RV-CC-LIMIT-CONTEXT-LINKAGE-AND-RETENTION-LIMITS`.
- [REQ-LIMIT-0024](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0024) → `CC-LIMIT-CONTEXT-LINKAGE-AND-RETENTION-LIMITS`, `RV-CC-LIMIT-CONTEXT-LINKAGE-AND-RETENTION-LIMITS`.
- [REQ-LIMIT-0025](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0025) → `CC-LIMIT-CONTEXT-LINKAGE-AND-RETENTION-LIMITS`, `RV-CC-LIMIT-CONTEXT-LINKAGE-AND-RETENTION-LIMITS`.

### CC-LIMIT-LIMIT-HANDLING

Public source: `library/registered-limits.md`, section **Limit Handling**. Intended target: `internal/execution/budget.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcLimitLimitHandling`.

Case plan: Admission: with a recording transport, submit tasks that each exceed one registered maximum by one unit (complete Task JSON 65,537 bytes, task_id 129 bytes, input 49,153 bytes, options 8,193 bytes, 17 labels, 29 runs per task, max_duration_ms 60,001, max_bytes_sent 65,537, max_bytes_received 262,145, max_connections 13, a NET-05 profile with 9 primitive steps) and assert each is rejected with a bounded admission error such as task_limit_exceeded and zero network activity, while the at-maximum variant is accepted. Runtime: with fake clocks and scripted transports, drive the outer deadline (reason attempt_deadline_exceeded), received bytes beyond max_bytes_received and a connection beyond max_connections (reason resource_limit), and an HTTP-01 body reaching its bound (registered body_limit_reached observation) and assert the active operation stops with that registered outcome. Assert a resource tracker shows every socket, body, timer, callback and temporary file closed after each stop, and assert serialized results contain no limit values or raw counters beyond fields the applicable result schema defines. Iterate attempts to raise a limit through labels, input, options, schedule, execution, safety and an unknown field, and through host configuration above a registered maximum, and assert rejection with effective limits unchanged. Requirements: REQ-LIMIT-0026, REQ-LIMIT-0027, REQ-LIMIT-0028, REQ-LIMIT-0029, REQ-LIMIT-0030, REQ-LIMIT-0031.

Scope: Synthetic core contract cases. No execution evidence or support claim.

Review case `RV-CC-LIMIT-LIMIT-HANDLING` (`docs/conformance/cross-cutting-review.md`, `heading:RV-CC-LIMIT-LIMIT-HANDLING`): Human review case; not automated coverage. Reviewer role: Specification governance and threat-model reviewer. Reviewed artifact: Change history of the Go limit constants and the corresponding public registered-limits.md versions at the reviewed revision. Checklist: (1) Do the implementation's registered maxima match registered-limits.md at the pinned specification version? (2) Is every increase of a registered maximum in the history accompanied by a new reviewed public specification version? (3) Is a threat-model review recorded for every such increase? (4) Does no documented host or deployment configuration path allow raising a value above the registered maximum? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-LIMIT-0026 -->
<!-- requirement-ref: REQ-LIMIT-0027 -->
<!-- requirement-ref: REQ-LIMIT-0028 -->
<!-- requirement-ref: REQ-LIMIT-0029 -->
<!-- requirement-ref: REQ-LIMIT-0030 -->
<!-- requirement-ref: REQ-LIMIT-0031 -->

- [REQ-LIMIT-0026](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0026) → `CC-LIMIT-LIMIT-HANDLING`, `RV-CC-LIMIT-LIMIT-HANDLING`.
- [REQ-LIMIT-0027](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0027) → `CC-LIMIT-LIMIT-HANDLING`, `RV-CC-LIMIT-LIMIT-HANDLING`.
- [REQ-LIMIT-0028](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0028) → `CC-LIMIT-LIMIT-HANDLING`, `RV-CC-LIMIT-LIMIT-HANDLING`.
- [REQ-LIMIT-0029](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0029) → `CC-LIMIT-LIMIT-HANDLING`, `RV-CC-LIMIT-LIMIT-HANDLING`.
- [REQ-LIMIT-0030](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0030) → `CC-LIMIT-LIMIT-HANDLING`, `RV-CC-LIMIT-LIMIT-HANDLING`.
- [REQ-LIMIT-0031](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0031) → `CC-LIMIT-LIMIT-HANDLING`, `RV-CC-LIMIT-LIMIT-HANDLING`.

### CC-LINKAGE-OVERVIEW

Public source: `library/result-linkage.md`, section **Overview**. Intended target: `internal/bundle/linkage.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcLinkageOverview`.

Case plan: Initialize once without result_linkage and once with mode "none", then build at least three bundles across two rotation epochs using a deterministic clock and a fake LinkageKeyStore. Assert no bundle contains a linkage object, each bundle_id is distinct and single-use, no bundle-level value other than schema constants and task-defined fields repeats across bundles, and the fake key store records zero DeriveHMACSHA256 calls and no secret creation. Enable rotating_installation with synthetic scope "org.example.app.test" and assert the linkage object contains only mode, scope, epoch and token, with no person, account, device or access-network field. Requirements: REQ-LINKAGE-0001.

Scope: Synthetic core contract cases. No execution evidence or support claim.

Review case `RV-CC-LINKAGE-OVERVIEW` (`docs/conformance/cross-cutting-review.md`, `heading:RV-CC-LINKAGE-OVERVIEW`): Human review case; not automated coverage. Reviewer role: Privacy documentation reviewer (independent of the implementer). Reviewed artifact: Go linkage API documentation and host integration guidance at the reviewed implementation revision. Checklist: (1) Does the documentation describe linkage as pseudonymous and never as anonymous? (2) Does the documentation state that linkage does not identify a person, host application account, device or access network? (3) Does the documentation state that the default mode is no cross-bundle linkage? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-LINKAGE-0001 -->

- [REQ-LINKAGE-0001](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0001) → `CC-LINKAGE-OVERVIEW`, `RV-CC-LINKAGE-OVERVIEW`.

### CC-LINKAGE-GOALS-AND-NON-GOALS

Public source: `library/result-linkage.md`, section **Goals and Non-goals**. Intended target: `internal/bundle/linkage.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcLinkageGoalsAndNonGoals`.

Case plan: Attempt cross-app/deployment/reinstall/reset/account restoration and targeted delivery; no stable identity continuity is permitted.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LINKAGE-0002 -->

- [REQ-LINKAGE-0002](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0002) → `CC-LINKAGE-GOALS-AND-NON-GOALS`.

### CC-LINKAGE-INITIALIZATION-CONFIGURATION

Public source: `library/result-linkage.md`, section **Initialization Configuration**. Intended target: `internal/bundle/linkage.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcLinkageInitializationConfiguration`.

Case plan: Iterate result_linkage initialization values: mode "none" and "rotating_installation" accepted; other strings and wrong case rejected. For rotating_installation, missing scope or missing rotation_period_seconds rejected; scope of 1 and 128 allowed characters accepted, empty and 129 characters rejected, uppercase letters, spaces, "/" and non-ASCII rejected; rotation_period_seconds 86,399 rejected, 86,400 and 604,800 accepted, 604,801 rejected. For mode "none", scope or rotation_period_seconds present is rejected. Absent result_linkage behaves as mode "none" with no linkage object and no key derivation. Assert tasks, scripted server responses and runtime task configuration carrying result_linkage, scope or rotation fields cannot enable linkage, change scope or lengthen rotation; that no runtime setter changes linkage configuration and a changed configuration takes effect only after reinitialization; and that trusted local safety policy disabling linkage yields no linkage object even when initialization requests rotating_installation. Requirements: REQ-LINKAGE-0003, REQ-LINKAGE-0004, REQ-LINKAGE-0005, REQ-LINKAGE-0006.

Scope: Synthetic core contract cases. No execution evidence or support claim.

Review case `RV-CC-LINKAGE-INITIALIZATION-CONFIGURATION` (`docs/conformance/cross-cutting-review.md`, `heading:RV-CC-LINKAGE-INITIALIZATION-CONFIGURATION`): Human review case; not automated coverage. Reviewer role: Deployment privacy reviewer (independent of the deployment operator). Reviewed artifact: Deployment result_linkage configuration and its documented scope rationale at the reviewed revision. Checklist: (1) Does the configured scope identify one reviewed Barometer deployment, such as one application environment? (2) Is the scope fixed for the deployment rather than varying per user or installation? (3) Is the scope free of, and not derived from, any user, account, installation, device, access network, proxy endpoint, experiment participant or other high-cardinality value? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-LINKAGE-0003 -->
<!-- requirement-ref: REQ-LINKAGE-0004 -->
<!-- requirement-ref: REQ-LINKAGE-0005 -->
<!-- requirement-ref: REQ-LINKAGE-0006 -->

- [REQ-LINKAGE-0003](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0003) → `CC-LINKAGE-INITIALIZATION-CONFIGURATION`, `RV-CC-LINKAGE-INITIALIZATION-CONFIGURATION`.
- [REQ-LINKAGE-0004](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0004) → `CC-LINKAGE-INITIALIZATION-CONFIGURATION`, `RV-CC-LINKAGE-INITIALIZATION-CONFIGURATION`.
- [REQ-LINKAGE-0005](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0005) → `CC-LINKAGE-INITIALIZATION-CONFIGURATION`, `RV-CC-LINKAGE-INITIALIZATION-CONFIGURATION`.
- [REQ-LINKAGE-0006](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0006) → `CC-LINKAGE-INITIALIZATION-CONFIGURATION`, `RV-CC-LINKAGE-INITIALIZATION-CONFIGURATION`.

### CC-LINKAGE-INSTALLATION-SECRET

Public source: `library/result-linkage.md`, section **Installation Secret**. Intended target: `internal/bundle/linkage.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcLinkageInstallationSecret`.

Case plan: Use fake protected-key capabilities and controlled randomness to check fresh 256-bit key creation, no host-derived secret, protected ownership and no export/backup/network/log/analytics sinks.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LINKAGE-0007 -->
<!-- requirement-ref: REQ-LINKAGE-0008 -->
<!-- requirement-ref: REQ-LINKAGE-0009 -->

- [REQ-LINKAGE-0007](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0007) → `CC-LINKAGE-INSTALLATION-SECRET`.
- [REQ-LINKAGE-0008](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0008) → `CC-LINKAGE-INSTALLATION-SECRET`.
- [REQ-LINKAGE-0009](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0009) → `CC-LINKAGE-INSTALLATION-SECRET`.

### CC-LINKAGE-TOKEN-DERIVATION

Public source: `library/result-linkage.md`, section **Token Derivation**. Intended target: `internal/bundle/linkage.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcLinkageTokenDerivation`.

Case plan: Use independent HMAC known-answer vectors for exact purpose/separator/scope/decimal epoch message, floor boundaries, base64url and memory cleanup; suppress linkage on invalid/rollback clock until nondecreasing validity.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LINKAGE-0010 -->
<!-- requirement-ref: REQ-LINKAGE-0011 -->
<!-- requirement-ref: REQ-LINKAGE-0012 -->
<!-- requirement-ref: REQ-LINKAGE-0013 -->
<!-- requirement-ref: REQ-LINKAGE-0014 -->

- [REQ-LINKAGE-0010](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0010) → `CC-LINKAGE-TOKEN-DERIVATION`.
- [REQ-LINKAGE-0011](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0011) → `CC-LINKAGE-TOKEN-DERIVATION`.
- [REQ-LINKAGE-0012](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0012) → `CC-LINKAGE-TOKEN-DERIVATION`.
- [REQ-LINKAGE-0013](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0013) → `CC-LINKAGE-TOKEN-DERIVATION`.
- [REQ-LINKAGE-0014](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0014) → `CC-LINKAGE-TOKEN-DERIVATION`.

### CC-LINKAGE-ROTATION-SEMANTICS

Public source: `library/result-linkage.md`, section **Rotation Semantics**. Intended target: `internal/bundle/linkage.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcLinkageRotationSemantics`.

Case plan: Compare same/different scopes/epochs across boundaries; no old-new token mapping or continuity request, and no period extension beyond reviewed bounds.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LINKAGE-0015 -->
<!-- requirement-ref: REQ-LINKAGE-0016 -->
<!-- requirement-ref: REQ-LINKAGE-0017 -->

- [REQ-LINKAGE-0015](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0015) → `CC-LINKAGE-ROTATION-SEMANTICS`.
- [REQ-LINKAGE-0016](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0016) → `CC-LINKAGE-ROTATION-SEMANTICS`.
- [REQ-LINKAGE-0017](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0017) → `CC-LINKAGE-ROTATION-SEMANTICS`.

### CC-LINKAGE-GO-CORE-AND-MOBILE-BINDING-ARCHITECTURE

Public source: `library/result-linkage.md`, section **Go Core and Mobile Binding Architecture**. Intended target: `docs/conformance/native-linkage.md`. Planned case location: `docs/conformance/native-linkage.md`, `heading:CC-LINKAGE-GO-CORE-AND-MOBILE-BINDING-ARCHITECTURE`.

Case plan: Run automated native device-validation suites on Android and Apple release-configuration builds and through the Flutter package. Initialize through Flutter with only the plain result_linkage map (mode "rotating_installation", synthetic scope, rotation_period_seconds 604,800), create a bundle, then call clearMeasurementData, and assert both calls succeed with equivalent semantics on Android and Apple. Attempt to pass a callback, native interface, key handle, byte array or platform storage object inside or alongside result_linkage and assert rejection. Restart the process within one epoch and assert the token is unchanged, demonstrating the persistent platform adapter is connected rather than an in-memory store. Inject protected-storage unavailability and assert bundles omit the linkage object instead of silently using an in-memory implementation. In a Go desktop test build, assert the ephemeral in-memory implementation is explicitly marked and loses linkage after process exit. Requirements: REQ-LINKAGE-0018, REQ-LINKAGE-0019, REQ-LINKAGE-0020, REQ-LINKAGE-0021, REQ-LINKAGE-0022.

Scope: Native adapter and device release validation; planned outside portable Go runtime. No execution evidence or support claim.

Review case `RV-CC-LINKAGE-GO-CORE-AND-MOBILE-BINDING-ARCHITECTURE` (`docs/conformance/native-linkage.md`, `heading:RV-CC-LINKAGE-GO-CORE-AND-MOBILE-BINDING-ARCHITECTURE`): Human review case; not automated coverage. Reviewer role: Mobile binding architecture reviewer (independent of the binding implementer). Reviewed artifact: Go core, gomobile export surface, Android and Apple binding source and Flutter package API at the reviewed revision. Checklist: (1) Do measurement and bundle logic remain in the Go core? (2) Is the LinkageKeyStore-equivalent capability internal, never accepted from Flutter or the host application, and absent from the gomobile and Flutter exported surface? (3) Is the raw installation secret absent from the capability interface, with derivation performed without exporting the key where the platform supports it? (4) Do the Android and Apple bindings own the concrete storage adapter and connect it during native initialization? (5) Is the ephemeral in-memory implementation unreachable as an unnoticed fallback in mobile builds? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-LINKAGE-0018 -->
<!-- requirement-ref: REQ-LINKAGE-0019 -->
<!-- requirement-ref: REQ-LINKAGE-0020 -->
<!-- requirement-ref: REQ-LINKAGE-0021 -->
<!-- requirement-ref: REQ-LINKAGE-0022 -->

- [REQ-LINKAGE-0018](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0018) → `CC-LINKAGE-GO-CORE-AND-MOBILE-BINDING-ARCHITECTURE`, `RV-CC-LINKAGE-GO-CORE-AND-MOBILE-BINDING-ARCHITECTURE`.
- [REQ-LINKAGE-0019](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0019) → `CC-LINKAGE-GO-CORE-AND-MOBILE-BINDING-ARCHITECTURE`, `RV-CC-LINKAGE-GO-CORE-AND-MOBILE-BINDING-ARCHITECTURE`.
- [REQ-LINKAGE-0020](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0020) → `CC-LINKAGE-GO-CORE-AND-MOBILE-BINDING-ARCHITECTURE`, `RV-CC-LINKAGE-GO-CORE-AND-MOBILE-BINDING-ARCHITECTURE`.
- [REQ-LINKAGE-0021](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0021) → `CC-LINKAGE-GO-CORE-AND-MOBILE-BINDING-ARCHITECTURE`, `RV-CC-LINKAGE-GO-CORE-AND-MOBILE-BINDING-ARCHITECTURE`.
- [REQ-LINKAGE-0022](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0022) → `CC-LINKAGE-GO-CORE-AND-MOBILE-BINDING-ARCHITECTURE`, `RV-CC-LINKAGE-GO-CORE-AND-MOBILE-BINDING-ARCHITECTURE`.

### CC-LINKAGE-CLEAR-DATA-CONTRACT

Public source: `library/result-linkage.md`, section **Clear-data Contract**. Intended target: `internal/persistence/clear.go`. Case location: `internal/persistence/cross_cutting_conformance_test.go`, `symbol:TestCcLinkageClearDataContract`.

Case plan: Go-owned rows only. Seed a storage scope with pending bundles, run accounting, replay records and a synthetic linkage key, and fault-inject each clearing step before and after it takes effect. Assert an explicit incomplete error for every failure, a durable clearing marker that persists across a simulated restart while any deletion step may have run, idempotent recovery clearing from every partial state that leaves every record class and the key deleted, and that clearing derives no replacement key and has no transport. Cancellation of active attempts with bounded cleanup is asserted by the root UT cases. Host user-interface and operating-system reset wiring, native keystores, reinstall sentinels and host-owned copies are planned in CC-LINKAGE-CLEAR-DATA-CONTRACT-HOST (DEV-455). Requirements: REQ-LINKAGE-0023, REQ-LINKAGE-0024.

DEV-600 unit-test evidence: the `UT.root.*` cases in the [evidence report](../traceability/evidence/go-unit-73faaf5.md) cover only the lifecycle part of clearing in the root `Client` shell. DEV-564 limits this family to its Go-owned rows, moves the host, operating-system and native rows and REQ-LINKAGE-0025 and REQ-LINKAGE-0026 to `CC-LINKAGE-CLEAR-DATA-CONTRACT-HOST`, and executes it against `persistence.Clear` ([evidence](../traceability/evidence/go-unit-e372219.md#CC-LINKAGE-CLEAR-DATA-CONTRACT)).

Scope: Synthetic core contract cases. Passed at Go `e372219` over the test-only record store; partial evidence only, with no device-durability, native storage or support claim.

<!-- requirement-ref: REQ-LINKAGE-0023 -->
<!-- requirement-ref: REQ-LINKAGE-0024 -->

- [REQ-LINKAGE-0023](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0023) → `CC-LINKAGE-CLEAR-DATA-CONTRACT`.
- [REQ-LINKAGE-0024](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0024) → `CC-LINKAGE-CLEAR-DATA-CONTRACT`.

### CC-LINKAGE-CLEAR-DATA-CONTRACT-HOST

Public source: `library/result-linkage.md`, section **Clear-data Contract**. Intended target: native adapters and host integration (DEV-455). Planned case location: `docs/conformance/native-linkage.md`, `heading:CC-LINKAGE-CLEAR-DATA-CONTRACT-HOST`.

Case plan: With linkage enabled on Android and Apple test builds using a synthetic scope, start an active attempt and hold pending and acknowledged bundles, then trigger every host action that promises to remove Barometer data (clear local data, reset privacy data, delete downloaded or offline data, and logout where the product describes it as removing local user data) and assert each calls clearMeasurementData; separately clear application data through the operating system without calling Barometer. After each, assert the protected key, linkage epoch state, reinstall sentinel, Network Context cache and Barometer-owned files are deleted or invalidated, no host-owned copy remains, no upload or network request occurred, and no replacement secret exists until linkage is next required. Interrupt deletion and assert explicit incomplete clearing and recovery. Requirements: REQ-LINKAGE-0023, REQ-LINKAGE-0024, REQ-LINKAGE-0025, REQ-LINKAGE-0026.

Scope: Native adapter and device release validation; planned outside portable Go runtime. Split from `CC-LINKAGE-CLEAR-DATA-CONTRACT` by DEV-564. No execution evidence or support claim.

<!-- requirement-ref: REQ-LINKAGE-0023 -->
<!-- requirement-ref: REQ-LINKAGE-0024 -->
<!-- requirement-ref: REQ-LINKAGE-0025 -->
<!-- requirement-ref: REQ-LINKAGE-0026 -->

- [REQ-LINKAGE-0023](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0023) → `CC-LINKAGE-CLEAR-DATA-CONTRACT-HOST`.
- [REQ-LINKAGE-0024](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0024) → `CC-LINKAGE-CLEAR-DATA-CONTRACT-HOST`.
- [REQ-LINKAGE-0025](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0025) → `CC-LINKAGE-CLEAR-DATA-CONTRACT-HOST`.
- [REQ-LINKAGE-0026](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0026) → `CC-LINKAGE-CLEAR-DATA-CONTRACT-HOST`.

### CC-LINKAGE-ANDROID-STORAGE

Public source: `library/result-linkage.md`, section **Android Storage**. Intended target: `docs/conformance/native-linkage.md`. Planned case location: `docs/conformance/native-linkage.md`, `heading:CC-LINKAGE-ANDROID-STORAGE`.

Case plan: Run an automated Android device-validation matrix on supported device and OS versions with linkage enabled using a synthetic scope; before each case record the current-epoch token T1. Assert the application-scoped Android Keystore HMAC-SHA-256 key is non-exportable where the platform supports it, associated non-secret state is in the no-backup directory (Context.getNoBackupFilesDir()), no Barometer state exists in shared preferences, Auto Backup or device-to-device-eligible files, external storage or host analytics storage, and backup rules explicitly exclude all Barometer state with no cloud or cross-device synchronization enabled. Iterate explicit clearMeasurementData; system Settings Clear storage or Clear data; uninstall then reinstall; backup and restore to the same and to a different device where supported; device-to-device migration; and interrupted deletion. After each case assert T1 is never emitted again within the same epoch and no case depends on token continuity. For orphan recovery, delete only the key or only the no-backup state and assert initialization detects the generation mismatch, deletes the orphaned material and creates no new secret until linkage is next required. Requirements: REQ-LINKAGE-0027, REQ-LINKAGE-0028, REQ-LINKAGE-0029, REQ-LINKAGE-0030, REQ-LINKAGE-0031.

Scope: Native adapter and device release validation; planned outside portable Go runtime. No execution evidence or support claim.

<!-- requirement-ref: REQ-LINKAGE-0027 -->
<!-- requirement-ref: REQ-LINKAGE-0028 -->
<!-- requirement-ref: REQ-LINKAGE-0029 -->
<!-- requirement-ref: REQ-LINKAGE-0030 -->
<!-- requirement-ref: REQ-LINKAGE-0031 -->

- [REQ-LINKAGE-0027](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0027) → `CC-LINKAGE-ANDROID-STORAGE`.
- [REQ-LINKAGE-0028](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0028) → `CC-LINKAGE-ANDROID-STORAGE`.
- [REQ-LINKAGE-0029](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0029) → `CC-LINKAGE-ANDROID-STORAGE`.
- [REQ-LINKAGE-0030](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0030) → `CC-LINKAGE-ANDROID-STORAGE`.
- [REQ-LINKAGE-0031](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0031) → `CC-LINKAGE-ANDROID-STORAGE`.

### CC-LINKAGE-APPLE-PLATFORM-STORAGE

Public source: `library/result-linkage.md`, section **Apple Platform Storage**. Intended target: `docs/conformance/native-linkage.md`. Planned case location: `docs/conformance/native-linkage.md`, `heading:CC-LINKAGE-APPLE-PLATFORM-STORAGE`.

Case plan: Run an automated Apple device-validation matrix with linkage enabled using a synthetic scope; before each case record the current-epoch token T1. Assert the Keychain item uses kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly when background derivation is configured (or kSecAttrAccessibleWhenUnlockedThisDeviceOnly for a foreground-only configuration), kSecAttrSynchronizable false and an application-private access group, and that a separate test application or extension cannot read it. Assert the non-secret reinstall sentinel is in the application sandbox container and not in Keychain, iCloud key-value storage, CloudKit, a shared app group or NSUbiquitousKeyValueStore. Exercise the sentinel truth table: both present, existing state continues; sentinel absent with surviving Keychain item, item deleted, sentinel created and no new secret until linkage is required; sentinel present with Keychain item absent or invalid, inconsistent state removed and no new secret until required. Iterate explicit clearMeasurementData, uninstall then reinstall, encrypted and unencrypted backup/restore, migration to another device, iCloud Keychain enabled and disabled, missing sentinel, missing Keychain item and interrupted deletion, and assert T1 is never accepted or emitted again after reinstall, restore or migration, with linkage reset when a restored container is detected. Requirements: REQ-LINKAGE-0032, REQ-LINKAGE-0033, REQ-LINKAGE-0034, REQ-LINKAGE-0035, REQ-LINKAGE-0036.

Scope: Native adapter and device release validation; planned outside portable Go runtime. No execution evidence or support claim.

Review case `RV-CC-LINKAGE-APPLE-PLATFORM-STORAGE` (`docs/conformance/native-linkage.md`, `heading:RV-CC-LINKAGE-APPLE-PLATFORM-STORAGE`): Human review case; not automated coverage. Reviewer role: Apple platform storage reviewer (independent of the binding implementer). Reviewed artifact: Apple binding storage design notes and device-validation results across the tested OS versions at the reviewed revision. Checklist: (1) Is the choice between kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly and kSecAttrAccessibleWhenUnlockedThisDeviceOnly justified by whether the integration derives tokens in the background? (2) For each tested OS version, is it documented whether a restored container can be reliably distinguished from the original installation? (3) Where it cannot be reliably distinguished, does the integration reset linkage rather than preserve it? (4) Is the Keychain item not shared with unrelated applications, extensions, analytics components or host account services in the shipped entitlement configuration? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-LINKAGE-0032 -->
<!-- requirement-ref: REQ-LINKAGE-0033 -->
<!-- requirement-ref: REQ-LINKAGE-0034 -->
<!-- requirement-ref: REQ-LINKAGE-0035 -->
<!-- requirement-ref: REQ-LINKAGE-0036 -->

- [REQ-LINKAGE-0032](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0032) → `CC-LINKAGE-APPLE-PLATFORM-STORAGE`, `RV-CC-LINKAGE-APPLE-PLATFORM-STORAGE`.
- [REQ-LINKAGE-0033](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0033) → `CC-LINKAGE-APPLE-PLATFORM-STORAGE`, `RV-CC-LINKAGE-APPLE-PLATFORM-STORAGE`.
- [REQ-LINKAGE-0034](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0034) → `CC-LINKAGE-APPLE-PLATFORM-STORAGE`, `RV-CC-LINKAGE-APPLE-PLATFORM-STORAGE`.
- [REQ-LINKAGE-0035](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0035) → `CC-LINKAGE-APPLE-PLATFORM-STORAGE`, `RV-CC-LINKAGE-APPLE-PLATFORM-STORAGE`.
- [REQ-LINKAGE-0036](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0036) → `CC-LINKAGE-APPLE-PLATFORM-STORAGE`, `RV-CC-LINKAGE-APPLE-PLATFORM-STORAGE`.

### CC-LINKAGE-STORAGE-FAILURE

Public source: `library/result-linkage.md`, section **Storage Failure**. Intended target: `internal/bundle/linkage.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcLinkageStorageFailure`.

Case plan: Inject unavailable/locked/inconsistent/invalid storage; omit linkage with permitted bundle grouping, preserve deletion, forbid weak persistence/host-visible/sticky-token fallback and sanitize diagnostics.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LINKAGE-0037 -->
<!-- requirement-ref: REQ-LINKAGE-0038 -->
<!-- requirement-ref: REQ-LINKAGE-0039 -->

- [REQ-LINKAGE-0037](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0037) → `CC-LINKAGE-STORAGE-FAILURE`.
- [REQ-LINKAGE-0038](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0038) → `CC-LINKAGE-STORAGE-FAILURE`.
- [REQ-LINKAGE-0039](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0039) → `CC-LINKAGE-STORAGE-FAILURE`.

### CC-LINKAGE-INGESTION-SEPARATION

Public source: `library/result-linkage.md`, section **Ingestion Separation**. Intended target: `docs/conformance/host-delivery-review.md`. Planned case location: `docs/conformance/host-delivery-review.md`, `heading:RV-CC-LINKAGE-INGESTION-SEPARATION`.

Case plan: no automated case; review case `RV-CC-LINKAGE-INGESTION-SEPARATION`: Human review case; not automated coverage. Reviewer role: Collector and deployment privacy/safety reviewer (independent of the collector operator). Reviewed artifact: Synthetic deployment delivery-flow description and collector ingress and research-storage configuration at the reviewed revision. Checklist: (1) Is the network-facing ingress component separated from research storage so transport metadata is removed before bundles reach analysis? (2) Does ingress avoid or minimize source-IP logging? (3) Does ingress remove forwarding headers and other transport identifiers before forwarding? (4) Does the bundle endpoint avoid application account authentication where the deployment permits? (5) Is bundle_id retained only as long as needed for idempotency and retry handling? (6) Are size, rate and schema limits enforced without building a persistent device profile? (7) If bundles are delivered through an API authenticated to a user account, is a separate privacy and safety review recorded and is the integration not described as anonymous? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

Scope: Host, collector or specification review; no runtime enforcement claim. No execution evidence or support claim.

<!-- requirement-ref: REQ-LINKAGE-0040 -->
<!-- requirement-ref: REQ-LINKAGE-0041 -->
<!-- requirement-ref: REQ-LINKAGE-0042 -->

- [REQ-LINKAGE-0040](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0040) → `RV-CC-LINKAGE-INGESTION-SEPARATION`.
- [REQ-LINKAGE-0041](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0041) → `RV-CC-LINKAGE-INGESTION-SEPARATION`.
- [REQ-LINKAGE-0042](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0042) → `RV-CC-LINKAGE-INGESTION-SEPARATION`.

### CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS

Public source: `library/result-linkage.md`, section **Privacy and Safety Requirements**. Intended target: `internal/bundle/linkage.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcLinkagePrivacyAndSafetyRequirements`.

Case plan: Use a fake LinkageKeyStore holding a synthetic secret canary, a deterministic clock, and recording fakes for logs, diagnostics, crash/analytics hooks, the Network Context endpoints, measurement targets and the local proxy. Assert initialization without result_linkage emits no linkage and creates no key, and that tasks carrying linkage-enabling fields cannot enable it. With rotating_installation enabled, assert the secret canary never appears in any API return value, error, log, diagnostic or hook output, and the token appears only in bundle linkage.token. Assert Network Context requests contain no bundle_id, scope, epoch or token, and target and proxy traffic (including CONNECT requests) contains no linkage material. Assert task admission and scheduling decisions are identical for linkage disabled and for two different tokens. Assert the previous epoch's token is not retained in memory after rotation and that clearMeasurementData removes cached tokens so the next bundle in the same epoch carries a new token. Requirements: REQ-LINKAGE-0043, REQ-LINKAGE-0044, REQ-LINKAGE-0045, REQ-LINKAGE-0046, REQ-LINKAGE-0047, REQ-LINKAGE-0048, REQ-LINKAGE-0049, REQ-LINKAGE-0050, REQ-LINKAGE-0051, REQ-LINKAGE-0052.

Scope: Synthetic core contract cases. No execution evidence or support claim.

Review case `RV-CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS` (`docs/conformance/cross-cutting-review.md`, `heading:RV-CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`): Human review case; not automated coverage. Reviewer role: Privacy and safety reviewer (independent of the implementer and deployment operator). Reviewed artifact: Go linkage documentation, host integration guidance and deployment analysis/retention plan at the reviewed revision. Checklist: (1) Does all documentation describe the token as installation-scoped and never as a user identifier? (2) Does the deployment plan exclude linkage use for per-installation task selection, retaliation, access control, advertising, product analytics and account enforcement? (3) Is raw token retention bounded to a documented research need? (4) Are combinations of the token with precise timestamps, source IP addresses, rare ASN and region combinations or host account data prevented, or separately reviewed? (5) Does the host expose an explicit user-reachable reset path wherever linkage is enabled? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-LINKAGE-0043 -->
<!-- requirement-ref: REQ-LINKAGE-0044 -->
<!-- requirement-ref: REQ-LINKAGE-0045 -->
<!-- requirement-ref: REQ-LINKAGE-0046 -->
<!-- requirement-ref: REQ-LINKAGE-0047 -->
<!-- requirement-ref: REQ-LINKAGE-0048 -->
<!-- requirement-ref: REQ-LINKAGE-0049 -->
<!-- requirement-ref: REQ-LINKAGE-0050 -->
<!-- requirement-ref: REQ-LINKAGE-0051 -->
<!-- requirement-ref: REQ-LINKAGE-0052 -->

- [REQ-LINKAGE-0043](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0043) → `CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-LINKAGE-0044](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0044) → `CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-LINKAGE-0045](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0045) → `CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-LINKAGE-0046](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0046) → `CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-LINKAGE-0047](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0047) → `CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-LINKAGE-0048](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0048) → `CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-LINKAGE-0049](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0049) → `CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-LINKAGE-0050](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0050) → `CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-LINKAGE-0051](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0051) → `CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-LINKAGE-0052](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/result-linkage.md#REQ-LINKAGE-0052) → `CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`, `RV-CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`.

### CC-TRUST-OVERVIEW

Public source: `library/task-trust.md`, section **Overview**. Intended target: `internal/admission/admit.go`. Planned case location: `internal/admission/cross_cutting_conformance_test.go`, `symbol:TestCcTrustOverview`.

Case plan: Admit synthetic NET-01 tasks with a fake delivery-record store, local policy, registry, deterministic clock and recording transport. A baseline task with an embedded delivery record for its exact bytes, an issuer permitted NET-01 0.1.0 and a matching target registration is accepted. Remove one prerequisite at a time (the delivery record, the issuer permission, the target registration, the registry entry for the exact version, and a numeric value raised above its registered limit) and assert rejection with, respectively, task_source_unauthenticated, task_issuer_not_authorized, target_not_registered, task_unsupported and task_limit_exceeded, and zero transport calls. Add task-controlled authority claims: top-level authenticated, issuer_id, policy_generation and trusted fields are rejected as task_schema_invalid, and a permitted label with the text "authenticated: true" stays opaque, producing the same authority, effective risk, targets and policy decision as the same task without it. With a synthetic registry baseline of medium, a task declaring low cannot lower effective risk. Assert no task field changes the active policy generation or registrations. Requirements: REQ-TRUST-0040, REQ-TRUST-0003.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-TRUST-0040 -->
<!-- requirement-ref: REQ-TRUST-0003 -->

- [REQ-TRUST-0040](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0040) → `CC-TRUST-OVERVIEW`.
- [REQ-TRUST-0003](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0003) → `CC-TRUST-OVERVIEW`.

### CC-TRUST-ACCEPTED-TASK-ORIGINS

Public source: `library/task-trust.md`, section **Accepted Task Origins**. Intended target: `internal/admission/admit.go`. Case location: `internal/admission/cross_cutting_conformance_test.go`, `symbol:TestCcTrustAcceptedTaskOrigins`, executed by DEV-594.

Case plan: Admit the same synthetic task bytes through delivery records of each origin: embedded with its issuer and no issued_at is accepted, and remote as a signed package from a trusted key of its issuer, issued at the fake clock, is accepted with a replay digest equal to the digest of the task bytes. For remote origin, remove one input at a time: a remote source with a host-supplied issuer cannot be constructed; the bare task presented as remote without a signed package, and a package with an added authenticated member, are rejected as task_source_unauthenticated; a record presented with package bytes differing by one byte is rejected as task_source_unauthenticated; an unavailable trusted clock is rejected as task_trust_unavailable; no delivery record at all is rejected as task_source_unauthenticated. For every rejection assert no replay or admitted-task record is written. Admission has no transport seam, so zero transport calls are established by checking that the dependency closure of admission contains no network package. Requirements: REQ-TRUST-0041, REQ-TRUST-0042, REQ-TRUST-0043, REQ-TRUST-0044.

Scope: Synthetic core contract cases. Passed at Go `09e2141` and, rewritten for signed packages by DEV-612, at Go `e663afe`, and re-executed at Go `baa3ada`, `d309b5b`, `19b5810`, `c9516c6`, `daeb718`, `8156c1a` and `2a94d00` ([evidence](../traceability/evidence/go-unit-2a94d00.md#CC-TRUST-ACCEPTED-TASK-ORIGINS)); partial evidence only, with synthetic keys, structural rather than runtime network absence and no support claim.

<!-- requirement-ref: REQ-TRUST-0041 -->
<!-- requirement-ref: REQ-TRUST-0042 -->
<!-- requirement-ref: REQ-TRUST-0043 -->
<!-- requirement-ref: REQ-TRUST-0044 -->

- [REQ-TRUST-0041](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0041) → `CC-TRUST-ACCEPTED-TASK-ORIGINS`.
- [REQ-TRUST-0042](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0042) → `CC-TRUST-ACCEPTED-TASK-ORIGINS`.
- [REQ-TRUST-0043](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0043) → `CC-TRUST-ACCEPTED-TASK-ORIGINS`.
- [REQ-TRUST-0044](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0044) → `CC-TRUST-ACCEPTED-TASK-ORIGINS`.

### CC-TRUST-TRUSTED-SOURCE-CONTEXT

Public source: `library/task-trust.md`, section **Trusted Source Context**. Intended target: `internal/admission/admit.go`. Planned case location: `internal/admission/cross_cutting_conformance_test.go`, `symbol:TestCcTrustTrustedSourceContext`.

Case plan: Construct source contexts through the admission seam, varying one property at a time. origin: embedded and remote accepted, any other value rejected. issuer_id: for embedded, 1 and 64 allowed bytes accepted and empty, 65 bytes, uppercase, space and "/" rejected; for remote, any host-supplied value rejected. policy_generation: 1 and 2^64-1 accepted, 0 rejected. There is no authentication flag to set. issued_at: for remote, taken only from a verified package, so equal to the fake clock and 300 seconds ahead are accepted and 301 seconds ahead is rejected as task_source_unauthenticated; for embedded, absent. Assert task JSON and initialization JSON have no field that sets any property (unknown fields rejected) and the trust models reject JSON import. Admit and run a task with canary issuer_id and generation values against a recording transport and fake result sink, and assert the canaries are absent from every transmitted byte, Measurement Result and bundle. Assert the facade signature fixture accepts no raw key, signature verdict or authenticated flag, and that trusted keys enter only through the trusted local policy. Requirements: REQ-TRUST-0008, REQ-TRUST-0044, REQ-TRUST-0045.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-TRUST-0044 -->
<!-- requirement-ref: REQ-TRUST-0008 -->
<!-- requirement-ref: REQ-TRUST-0045 -->

- [REQ-TRUST-0044](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0044) → `CC-TRUST-TRUSTED-SOURCE-CONTEXT`.
- [REQ-TRUST-0008](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0008) → `CC-TRUST-TRUSTED-SOURCE-CONTEXT`.
- [REQ-TRUST-0045](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0045) → `CC-TRUST-TRUSTED-SOURCE-CONTEXT`.

### CC-TRUST-ISSUER-AUTHORIZATION

Public source: `library/task-trust.md`, section **Issuer Authorization**. Intended target: `internal/admission/admit.go`. Planned case location: `internal/admission/cross_cutting_conformance_test.go`, `symbol:TestCcTrustIssuerAuthorization`.

Case plan: Start from a fake configuration where registered behavior, issuer permission, local deployment permission, platform capability and task restrictions all allow a NET-01 0.1.0 direct-path task, which is accepted. Narrow one dimension at a time while the other four stay permissive, and assert rejection with zero transport calls: issuer lacks the test/version pair (task_issuer_not_authorized); issuer lacks the registration reference (target_not_registered); issuer permits only the proxy path for a direct task (task_issuer_not_authorized); issuer maximum risk low for a medium task (task_issuer_not_authorized); medium-risk enablement off; NET-05 enablement off for a NET-05 task; issuer attempt ceiling below the task max_duration_ms (task_limit_exceeded); task label namespace not permitted; and each of these narrowed in local deployment permission instead of the issuer. For each bound named in the public list, a task value equal to the effective bound is accepted, a value one unit above is rejected and never clamped, and a lower value is accepted and carried as the effective value. Requirements: REQ-TRUST-0010, REQ-TRUST-0011, REQ-TRUST-0012.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-TRUST-0010 -->
<!-- requirement-ref: REQ-TRUST-0011 -->
<!-- requirement-ref: REQ-TRUST-0012 -->

- [REQ-TRUST-0010](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0010) → `CC-TRUST-ISSUER-AUTHORIZATION`.
- [REQ-TRUST-0011](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0011) → `CC-TRUST-ISSUER-AUTHORIZATION`.
- [REQ-TRUST-0012](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0012) → `CC-TRUST-ISSUER-AUTHORIZATION`.

### CC-TRUST-TARGET-REGISTRATION

Public source: `library/task-trust.md`, section **Target Registration**. Intended target: `internal/admission/admit.go`. Planned case location: `internal/admission/cross_cutting_conformance_test.go`, `symbol:TestCcTrustTargetRegistration`.

Case plan: Register synthetic tuples for TCP-01 (address, port), TLS-01 (address, port, server name, ALPN), HTTP-01 (URL, expected status, checks, redirect registrations) and DNS-01 (resolver name, bootstrap addresses). For each, an exactly matching task is accepted. Change one tuple component at a time (hostname, IP address, port, scheme, URL path, resolver name, bootstrap address, server name, an ALPN value, expected status, a body or header check, a redirect target, direct versus proxy path) and assert target_not_registered with zero transport calls. Normalization: a form that the public rules normalize to the registered value matches, while a suffix (evil-example.test against example.test), a subdomain, a URL with an extra path prefix and an IPv4-mapped IPv6 address do not. Expiry one second after the fake clock is accepted; expiry at or before it is rejected. Registrations with URL credentials or interpolation syntax fail policy construction. Execution-owned rows: with a scripted resolver returning an unregistered address and a scripted redirect to an unregistered location, assert no connection to either. Requirements: REQ-TRUST-0013, REQ-TRUST-0014, REQ-TRUST-0015, REQ-TRUST-0016.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-TRUST-0013 -->
<!-- requirement-ref: REQ-TRUST-0014 -->
<!-- requirement-ref: REQ-TRUST-0015 -->
<!-- requirement-ref: REQ-TRUST-0016 -->

- [REQ-TRUST-0013](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0013) → `CC-TRUST-TARGET-REGISTRATION`.
- [REQ-TRUST-0014](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0014) → `CC-TRUST-TARGET-REGISTRATION`.
- [REQ-TRUST-0015](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0015) → `CC-TRUST-TARGET-REGISTRATION`.
- [REQ-TRUST-0016](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0016) → `CC-TRUST-TARGET-REGISTRATION`.

### CC-TRUST-ADMISSION-PROCEDURE

Public source: `library/task-trust.md`, section **Admission Procedure**. Intended target: `internal/admission/admit.go`. Planned case location: `internal/admission/cross_cutting_conformance_test.go`, `symbol:TestCcTrustAdmissionProcedure`.

Case plan: Instrument admission with a fake delivery-record store, clock, registry, policy, replay store and recording transport that log the order of calls. A valid task runs the eleven public steps in order, and its handle is returned only after the replay write commits. Build tasks that each fail exactly one step and assert the listed class: missing, expired, foreign-instance or byte-mismatched delivery record, task_source_unauthenticated; a remote task whose signed package fails verification, task_source_unauthenticated; unavailable trusted clock, task_trust_unavailable; a remote package issued 301 seconds after acceptance, task_source_unauthenticated, while 300 seconds is accepted; 65,537 bytes, task_limit_exceeded; unknown top-level field, task_schema_invalid; well-formed but unregistered test_version, task_unsupported; uppercase task_id, task_schema_invalid; 129-byte task_id, task_limit_exceeded; expired task, task_expired; execution.context requiring egress Network Context in a build without egress acquisition, task_unsupported; no installed policy, task_policy_unavailable; source policy_generation one above the active generation, task_policy_unavailable, and one below, policy_rollback; unknown issuer, task_issuer_not_authorized; a proxy task whose test, input or options are incompatible with HTTP CONNECT, task_schema_invalid; direct or proxy path not permitted by issuer authority or local policy, task_issuer_not_authorized; path not permitted by the matched target registration, and an unregistered target, target_not_registered; effective risk above the issuer maximum or a local-policy bound, and unknown risk without a reviewed local permission, task_issuer_not_authorized; replay conflict, task_id_conflict; replay write failure, task_trust_unavailable. For each assert exactly one class, that no later step's capability is called, that no replay or admitted-task record exists, that no handle is returned and that the transport, DNS and Network Context fakes see zero calls. Precedence follows the first failing step: an unknown top-level field with an unknown issuer returns task_schema_invalid (step 4 before 7), an unregistered test_version with a 129-byte task_id returns task_unsupported (step 4 before 5), an expired task with an unregistered target returns task_expired (step 6 before 8), a 65,537-byte task presented as remote without a signed package returns task_source_unauthenticated (step 2 before 3), and within step 8 invalid options with a proxy-incompatible test return task_schema_invalid with the compatibility rule not consulted, while a proxy-incompatible task whose path is not permitted by issuer or local policy, or by its matched registration, returns task_schema_invalid (compatibility before path authority and target registration). A valid task that requires egress context performs no context lookup during admission in a build that provides egress acquisition. Requirements: REQ-TRUST-0046, REQ-TRUST-0018, REQ-TRUST-0047, REQ-TRUST-0039. Step-8 input and options failures are task_schema_invalid, and a stricter test byte cap, checked before the content it bounds, is task_limit_exceeded (REQ-TRUST-0056).

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-TRUST-0046 -->
<!-- requirement-ref: REQ-TRUST-0018 -->
<!-- requirement-ref: REQ-TRUST-0047 -->
<!-- requirement-ref: REQ-TRUST-0039 -->
<!-- requirement-ref: REQ-TRUST-0056 -->

- [REQ-TRUST-0046](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0046) → `CC-TRUST-ADMISSION-PROCEDURE`.
- [REQ-TRUST-0018](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0018) → `CC-TRUST-ADMISSION-PROCEDURE`.
- [REQ-TRUST-0047](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0047) → `CC-TRUST-ADMISSION-PROCEDURE`.
- [REQ-TRUST-0039](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0039) → `CC-TRUST-ADMISSION-PROCEDURE`.
- [REQ-TRUST-0056](https://github.com/tv-rain/barometer-spec/blob/12062471875bbfbc55eafc94dc5a7ea4cd07ca5e/library/task-trust.md#REQ-TRUST-0056) → `CC-TRUST-ADMISSION-PROCEDURE`.

### CC-TRUST-TASK-IDENTIFIER-RULES

Public source: `library/task-trust.md`, section **Task Identifier Rules**. Intended target: `internal/admission/admit.go`. Planned case location: `internal/admission/cross_cutting_conformance_test.go`, `symbol:TestCcTrustTaskIdentifierRules`.

Case plan: Iterate task_id values: 1 and 128 bytes of allowed characters (lowercase ASCII letters, digits, ".", "_", "-") accepted; empty and 129 bytes rejected; uppercase, space, "/" and non-ASCII characters rejected, all before any network activity. Replay key: deliver the same task_id from two admitted issuers and assert both are accepted as distinct replay keys; deliver the same task_id from one issuer with identical bytes and assert idempotent redelivery, and with different bytes assert task_id_conflict. Campaign-wide identity: admit identical task bytes on two fake installations with different linkage tokens and assert each accepts the identical task_id and emits it unchanged in results, with no installation-specific rewriting or interpolation. Requirements: REQ-TRUST-0019, REQ-TRUST-0020.

Scope: Synthetic core contract cases. No execution evidence or support claim.

Review case `RV-CC-TRUST-TASK-IDENTIFIER-RULES` (`docs/conformance/cross-cutting-review.md`, `heading:RV-CC-TRUST-TASK-IDENTIFIER-RULES`): Human review case; not automated coverage. Reviewer role: Task issuer campaign reviewer (independent of the task author). Reviewed artifact: Issuer task-generation procedure and a sample of issued campaign task definitions at the reviewed revision. Checklist: (1) Does each task_id identify one campaign task definition within the issuer namespace? (2) Is each task_id identical for every installation receiving the same task definition? (3) Is no task_id derived from a person, account, installation, device, advertising ID, push token, linkage token, IP address or other participant-specific value? (4) Is no task_id used to select an individual installation or reconstruct cross-bundle linkage? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-TRUST-0019 -->
<!-- requirement-ref: REQ-TRUST-0020 -->

- [REQ-TRUST-0019](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0019) → `CC-TRUST-TASK-IDENTIFIER-RULES`, `RV-CC-TRUST-TASK-IDENTIFIER-RULES`.
- [REQ-TRUST-0020](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0020) → `CC-TRUST-TASK-IDENTIFIER-RULES`, `RV-CC-TRUST-TASK-IDENTIFIER-RULES`.

### CC-TRUST-REPLAY-DUPLICATE-AND-ROLLBACK-HANDLING

Public source: `library/task-trust.md`, section **Replay, Duplicate, and Rollback Handling**. Intended target: `internal/admission/state.go`. Case location: `internal/admission/cross_cutting_conformance_test.go`, `symbol:TestCcTrustReplayDuplicateAndRollbackHandling`, executed by DEV-591.

Case plan: Exercise every replay-table row with a deterministic clock and fake persistent storage: a new replay key continues admission; the same key with identical bytes is idempotent redelivery with no duplicate runs and unchanged expiry; the same key with one changed byte (including whitespace only) is rejected as task_id_conflict; a lower policy generation than the active one is rejected as policy_rollback; an expired task is rejected as task_expired and never reactivated on later redelivery; unavailable or corrupted replay state fails closed with task_trust_unavailable for a remote task and, with the same storage fault injected, for an embedded task, which is never accepted, assigned a handle or made executable without replay state. Assert the persisted record contains only the (issuer_id, task_id) key, a SHA-256 digest equal to an independently computed digest of the exact accepted bytes, the accepted policy generation and acceptance/expiry classes with retention ending 86,400 seconds after task expiry, and that synthetic canary input, target, label, participant and result strings are absent from stored bytes. Post-clearing redelivery rows are planned separately in CC-TRUST-REPLAY-AFTER-CLEARING. Requirements: REQ-TRUST-0021, REQ-TRUST-0022, REQ-TRUST-0023, REQ-TRUST-0024, REQ-TRUST-0038, REQ-LIMIT-0023.

Scope: Synthetic core contract cases. Passed at Go `9283870` and re-executed at Go `73faaf5`, `7b098a6`, `954b7d7`, `09e2141`, `e663afe`, `baa3ada`, `d309b5b`, `19b5810`, `c9516c6`, `daeb718`, `8156c1a` and `2a94d00` ([evidence](../traceability/evidence/go-unit-2a94d00.md#CC-TRUST-REPLAY-DUPLICATE-AND-ROLLBACK-HANDLING)); partial evidence only, with no production store, device-durability or support claim.

<!-- requirement-ref: REQ-TRUST-0021 -->
<!-- requirement-ref: REQ-TRUST-0022 -->
<!-- requirement-ref: REQ-TRUST-0023 -->
<!-- requirement-ref: REQ-TRUST-0024 -->
<!-- requirement-ref: REQ-TRUST-0038 -->
<!-- requirement-ref: REQ-LIMIT-0023 -->

- [REQ-TRUST-0021](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0021) → `CC-TRUST-REPLAY-DUPLICATE-AND-ROLLBACK-HANDLING`.
- [REQ-TRUST-0022](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0022) → `CC-TRUST-REPLAY-DUPLICATE-AND-ROLLBACK-HANDLING`.
- [REQ-TRUST-0023](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0023) → `CC-TRUST-REPLAY-DUPLICATE-AND-ROLLBACK-HANDLING`.
- [REQ-TRUST-0024](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0024) → `CC-TRUST-REPLAY-DUPLICATE-AND-ROLLBACK-HANDLING`.
- [REQ-TRUST-0038](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0038) → `CC-TRUST-REPLAY-DUPLICATE-AND-ROLLBACK-HANDLING`.
- [REQ-LIMIT-0023](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/registered-limits.md#REQ-LIMIT-0023) → `CC-TRUST-REPLAY-DUPLICATE-AND-ROLLBACK-HANDLING`.

### CC-TRUST-REPLAY-AFTER-CLEARING

Public source: `library/task-trust.md`, section **Replay, Duplicate, and Rollback Handling**. Intended target: `internal/admission/state.go` with the clearing capability (DEV-564). Case location: `internal/admission/cross_cutting_conformance_test.go`, `symbol:TestCcTrustReplayAfterClearing`.

Case plan: After clearMeasurementData deletes replay state with a deterministic clock and fake persistent storage, redeliver previously accepted exact task bytes: an unexpired task is evaluated as new, receives a fresh replay record with unchanged expiry derived from its not_after and creates no record for the cleared acceptance; an expired task is still rejected as task_expired; a failed or incomplete deletion is reported explicitly and leaves no task admitted without replay state. Requirements: REQ-TRUST-0024, REQ-TRUST-0025.

Scope: Synthetic core contract cases, split from `CC-TRUST-REPLAY-DUPLICATE-AND-ROLLBACK-HANDLING` by DEV-591. Passed at Go `e372219` against the production `persistence.State` over the test-only record store ([evidence](../traceability/evidence/go-unit-e372219.md#CC-TRUST-REPLAY-AFTER-CLEARING)); partial evidence only, with no device-durability or support claim.

<!-- requirement-ref: REQ-TRUST-0024 -->
<!-- requirement-ref: REQ-TRUST-0025 -->

- [REQ-TRUST-0024](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0024) → `CC-TRUST-REPLAY-AFTER-CLEARING`.
- [REQ-TRUST-0025](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0025) → `CC-TRUST-REPLAY-AFTER-CLEARING`.

### CC-TRUST-TARGETED-DELIVERY

Public source: `library/task-trust.md`, section **Targeted Delivery**. Intended target: `internal/admission/admit.go`. Planned case location: `internal/admission/cross_cutting_conformance_test.go`, `symbol:TestCcTrustTargetedDelivery`.

Case plan: Submit synthetic tasks attempting recipient selection by linkage or installation token, bundle_id or run_id, account, advertising, device, SIM, subscriber or push identifiers, exact IP address or precise location, or prior per-installation outcomes, through any selector, label, eligibility or unknown field, and assert each is rejected at admission before network activity. Submit tasks whose task_id, labels, targets or payload contain placeholder syntax for deployment, application version, cohort, country or network values and assert the library performs no local interpolation: values stay literal or the task is rejected. With linkage disabled and with two different synthetic tokens, deliver the same task set and assert identical admission and scheduling decisions, with no token exposed to task-selection or host-policy hooks. Requirements: REQ-TRUST-0026, REQ-TRUST-0027, REQ-TRUST-0028.

Scope: Synthetic core contract cases. No execution evidence or support claim.

Review case `RV-CC-TRUST-TARGETED-DELIVERY` (`docs/conformance/cross-cutting-review.md`, `heading:RV-CC-TRUST-TARGETED-DELIVERY`): Human review case; not automated coverage. Reviewer role: Host deployment policy reviewer (independent of the deployment operator). Reviewed artifact: Host task-eligibility policy and its reviewed criteria at the reviewed revision. Checklist: (1) Is eligibility limited to broad deployment, application-version, consented cohort, country or network criteria? (2) Are the criteria non-user-specific and the groups sufficiently broad? (3) Has the policy itself been reviewed and is that review recorded? (4) Does any deployment need participant-specific longitudinal research, and if so is it excluded from claims of compliance with this baseline? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-TRUST-0026 -->
<!-- requirement-ref: REQ-TRUST-0027 -->
<!-- requirement-ref: REQ-TRUST-0028 -->

- [REQ-TRUST-0026](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0026) → `CC-TRUST-TARGETED-DELIVERY`, `RV-CC-TRUST-TARGETED-DELIVERY`.
- [REQ-TRUST-0027](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0027) → `CC-TRUST-TARGETED-DELIVERY`, `RV-CC-TRUST-TARGETED-DELIVERY`.
- [REQ-TRUST-0028](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0028) → `CC-TRUST-TARGETED-DELIVERY`, `RV-CC-TRUST-TARGETED-DELIVERY`.

### CC-TRUST-LOCAL-POLICY-LIFECYCLE

Public source: `library/task-trust.md`, section **Local Policy Lifecycle**. Intended target: the root `Client` (`client.go`, `policy.go`). Case location: `policy_test.go`, `symbol:TestCcTrustLocalPolicyLifecycle`, executed by DEV-593.

Case plan: Initialize with policy generation 2 and admit a task. Assert from the root package source that only the replacement seam assigns the active policy, adapters or admitter, that admission leaves the active generation unchanged and that Initialize on an initialized client returns already_initialized. Drive replacement by quiesce and reinitialization with fault injection at each implemented step. Validation refuses generation 2 (equal) and 1 (lower), an incomplete policy, retained replay acceptances above the new generation, unavailable replay generations and a panicking trusted clock with policy_rejected. After quiescing, a panic while cancelling the current generation and a delivery store that refuses the new generation leave the client uninitialized, and reserved work outliving the lifecycle deadline leaves it blocked with cleanup_incomplete until close. After each validation failure assert the complete generation-2 policy, its handles and its delivery references stay active and its tasks still admit; after each later failure assert no policy remains and that a reinitialized client rejects generation-2 and generation-3 deliveries with task_policy_unavailable, never mixing registrations. Assert in-flight admission is cancelled with no replay record or handle. On a successful generation-3 replacement that removes the task's registration, assert the previous generation's context is cancelled, its task handles and delivery records no longer resolve, a generation-3 redelivery is rejected as target_not_registered and a generation-2 delivery as policy_rollback; after a generation-4 replacement restoring the registration, a redelivery resolves to the retained acceptance under a new handle, and after clearing, initializing generation 2 or 3 fails closed while generation 4 initializes. Send policy-shaped content through permitted task labels and unknown task members and assert the active generation, issuers and registrations are unchanged. Run concurrent replacements and admissions and assert one highest winning generation. Host authentication of policy updates, run cancellation, Network Context cache invalidation and collector, target, proxy and context endpoint paths are planned separately in CC-TRUST-LOCAL-POLICY-LIFECYCLE-RUNTIME. Requirements: REQ-TRUST-0029, REQ-TRUST-0030, REQ-TRUST-0031, REQ-TRUST-0032.

Scope: Synthetic core contract cases. Passed at Go `954b7d7` and re-executed at Go `09e2141`, `e663afe`, `baa3ada`, `d309b5b`, `19b5810`, `c9516c6`, `daeb718`, `8156c1a` and `2a94d00` ([evidence](../traceability/evidence/go-unit-2a94d00.md#CC-TRUST-LOCAL-POLICY-LIFECYCLE)); partial evidence only, with no production policy adapter, durable generation floor or support claim.

<!-- requirement-ref: REQ-TRUST-0029 -->
<!-- requirement-ref: REQ-TRUST-0030 -->
<!-- requirement-ref: REQ-TRUST-0031 -->
<!-- requirement-ref: REQ-TRUST-0032 -->

### CC-TRUST-LOCAL-POLICY-LIFECYCLE-RUNTIME

Public source: `library/task-trust.md`, section **Local Policy Lifecycle**. Intended target: the root `Client` with the host policy adapter (DEV-455), run reservation (DEV-561) and Network Context caching (DEV-647). Planned case location: `policy_test.go`, `symbol:TestCcTrustLocalPolicyLifecycleRuntime`.

Case plan: With the reviewed host policy adapter (DEV-455), run reservation (DEV-561) and Network Context caching (DEV-647) in place, reject a policy update whose authentication or integrity check fails before the replacement seam is called and assert the generation-2 policy stays active. On a generation-3 replacement that removes a registration, assert a queued and a running run of the affected task are cancelled before any generation-3 attempt starts and that no attempt continues under the removed registration. Populate access and egress Network Context snapshots under generation 2 and assert none is used by a generation-3 attempt. Return policy-shaped content from collector reply, target response, proxy and context endpoint fakes and assert the active generation, issuers and registrations are unchanged. Requirements: REQ-TRUST-0029, REQ-TRUST-0030, REQ-TRUST-0031, REQ-TRUST-0032.

Scope: Synthetic core contract cases, split from `CC-TRUST-LOCAL-POLICY-LIFECYCLE` by DEV-593. No execution evidence or support claim.

<!-- requirement-ref: REQ-TRUST-0029 -->
<!-- requirement-ref: REQ-TRUST-0030 -->
<!-- requirement-ref: REQ-TRUST-0031 -->
<!-- requirement-ref: REQ-TRUST-0032 -->

- [REQ-TRUST-0029](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0029) → `CC-TRUST-LOCAL-POLICY-LIFECYCLE`, `CC-TRUST-LOCAL-POLICY-LIFECYCLE-RUNTIME`.
- [REQ-TRUST-0030](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0030) → `CC-TRUST-LOCAL-POLICY-LIFECYCLE`, `CC-TRUST-LOCAL-POLICY-LIFECYCLE-RUNTIME`.
- [REQ-TRUST-0031](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0031) → `CC-TRUST-LOCAL-POLICY-LIFECYCLE`, `CC-TRUST-LOCAL-POLICY-LIFECYCLE-RUNTIME`.
- [REQ-TRUST-0032](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0032) → `CC-TRUST-LOCAL-POLICY-LIFECYCLE`, `CC-TRUST-LOCAL-POLICY-LIFECYCLE-RUNTIME`.

### CC-TRUST-ADMISSION-ERRORS

Public source: `library/task-trust.md`, section **Admission Errors**. Intended target: `internal/admission/admit.go`. Planned case location: `internal/admission/cross_cutting_conformance_test.go`, `symbol:TestCcTrustAdmissionErrors`.

Case plan: Trigger each public class through a synthetic delivery: task_source_unauthenticated, task_issuer_not_authorized, task_policy_unavailable, task_trust_unavailable, task_expired, task_id_conflict, policy_rollback, target_not_registered, task_limit_exceeded, task_schema_invalid and task_unsupported. Trigger task_unsupported with three otherwise valid tasks: a well-formed but unregistered (test_id, test_version) pair, a risk factor type the build does not recognize and rejects, and an execution.context value requiring egress Network Context in a build without egress acquisition; assert none returns task_schema_invalid, and that with a fake build that provides egress acquisition the same egress task is admitted. Assert each returns a task.AdmissionError whose code is the expected class and whose text is the fixed class string, and that no Measurement Result, bundle entry or result-sink call is produced. Embed unique canaries in the task payload, a target URL, a label, a fake credential, verification-key and signature bytes in the delivery record, and the raw error text of a failing fake store and clock. Assert no canary appears in the error text, in %v or %+v formatting, in any errors.Unwrap chain or in captured log output. Assert repeating an identical failing delivery returns an identical error. Requirements: REQ-TRUST-0033, REQ-TRUST-0034, REQ-TRUST-0035, REQ-TRUST-0036.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-TRUST-0033 -->
<!-- requirement-ref: REQ-TRUST-0034 -->
<!-- requirement-ref: REQ-TRUST-0035 -->
<!-- requirement-ref: REQ-TRUST-0036 -->

- [REQ-TRUST-0033](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0033) → `CC-TRUST-ADMISSION-ERRORS`.
- [REQ-TRUST-0034](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0034) → `CC-TRUST-ADMISSION-ERRORS`.
- [REQ-TRUST-0035](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0035) → `CC-TRUST-ADMISSION-ERRORS`.
- [REQ-TRUST-0036](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/library/task-trust.md#REQ-TRUST-0036) → `CC-TRUST-ADMISSION-ERRORS`.

### CC-TRUST-SIGNED-REMOTE-PACKAGES

Public source: `library/task-trust.md`, section **Signed Remote Task Packages**, and `library/registered-limits.md`. Intended target: `internal/admission/signed.go`. Case location: `internal/admission/signed_test.go`, `symbol:TestCcTrustSignedRemotePackages`, executed by DEV-612.

Case plan: Present signed remote task packages built from synthetic Ed25519 keys through a fake delivery store, a policy holding two keys of the remote issuer and one of another issuer, and a fake clock, changing one property of an otherwise valid package per case. Assert acceptance for a valid package, a second key of the issuer, a package padded with whitespace to exactly 98,304 bytes, acceptance exactly at not_after, issued_at exactly at not_before and issued_at 300 seconds ahead. Assert task_limit_exceeded for 98,305 bytes; task_policy_unavailable with no installed policy; task_source_unauthenticated for malformed packages (not JSON, a bare task, unknown, duplicate, escaped-duplicate or missing members, wrong member types, padded, standard-alphabet or newline-containing base64url, nonzero trailing bits, an empty task, a 63- or 65-byte signature, fractional, offset, lowercase, leap-second or invalid-date issued_at, invalid identifiers), an unknown key_id, a key of another issuer or a package re-wrapped under another issuer, acceptance one second after not_after, issued_at one second before not_before or after not_after, issued_at 301 seconds ahead, an altered task byte, issued_at, key_id or signature, and a signature over another context or over the bare task; and task_trust_unavailable when trusted time is unavailable at check 6. Assert exactly one class, from the first failing check, for packages failing several checks, and that no rejected package reaches the task codec. A validly signed package whose task is invalid or over 65,536 bytes fails its later step instead. Requirements: REQ-TRUST-0043, REQ-TRUST-0046, REQ-TRUST-0047, REQ-TRUST-0048, REQ-TRUST-0049, REQ-TRUST-0050, REQ-TRUST-0052, REQ-LIMIT-0032.

Scope: Synthetic core contract cases. Passed at Go `e663afe` and re-executed at Go `baa3ada`, `d309b5b`, `19b5810`, `c9516c6`, `daeb718`, `8156c1a` and `2a94d00` ([evidence](../traceability/evidence/go-unit-2a94d00.md#CC-TRUST-SIGNED-REMOTE-PACKAGES)); partial evidence only, with synthetic keys, no cryptographic review and no support claim.

<!-- requirement-ref: REQ-LIMIT-0032 -->
<!-- requirement-ref: REQ-TRUST-0043 -->
<!-- requirement-ref: REQ-TRUST-0046 -->
<!-- requirement-ref: REQ-TRUST-0047 -->
<!-- requirement-ref: REQ-TRUST-0048 -->
<!-- requirement-ref: REQ-TRUST-0049 -->
<!-- requirement-ref: REQ-TRUST-0050 -->
<!-- requirement-ref: REQ-TRUST-0052 -->

- [REQ-LIMIT-0032](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/registered-limits.md#REQ-LIMIT-0032) → `CC-TRUST-SIGNED-REMOTE-PACKAGES`.
- [REQ-TRUST-0043](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0043) → `CC-TRUST-SIGNED-REMOTE-PACKAGES`.
- [REQ-TRUST-0046](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0046) → `CC-TRUST-SIGNED-REMOTE-PACKAGES`.
- [REQ-TRUST-0047](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0047) → `CC-TRUST-SIGNED-REMOTE-PACKAGES`.
- [REQ-TRUST-0048](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0048) → `CC-TRUST-SIGNED-REMOTE-PACKAGES`.
- [REQ-TRUST-0049](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0049) → `CC-TRUST-SIGNED-REMOTE-PACKAGES`.
- [REQ-TRUST-0050](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0050) → `CC-TRUST-SIGNED-REMOTE-PACKAGES`.
- [REQ-TRUST-0052](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0052) → `CC-TRUST-SIGNED-REMOTE-PACKAGES`.

### CC-TRUST-SIGNED-PACKAGE-DISCLOSURE

Public source: `library/task-trust.md`, sections **Signed Remote Task Packages** and **Admission Errors**. Intended target: `internal/admission/signed.go`. Case location: `internal/admission/signed_test.go`, `symbol:TestCcTrustSignedPackageErrorsDiscloseNothing`, executed by DEV-612.

Case plan: Reject six synthetic signed packages that each fail a different verification check (a bad signature, an issuer moved after signing, an unknown key, malformed JSON, issuance one hour ahead and 98,305 bytes), with canaries in the task payload, key IDs, issuers, signature and public-key prefixes and the raw clock error. Assert that each error formats under %v, %+v, %#v and Error() without any canary or the words ed25519, base64, check, key or signature, and that the six errors produce only the two fixed class strings task_source_unauthenticated and task_limit_exceeded, so the failing check cannot be distinguished beyond its class. Requirements: REQ-TRUST-0035, REQ-TRUST-0052.

Scope: Synthetic core contract cases. Passed at Go `e663afe` and re-executed at Go `baa3ada`, `d309b5b`, `19b5810`, `c9516c6`, `daeb718`, `8156c1a` and `2a94d00` ([evidence](../traceability/evidence/go-unit-2a94d00.md#CC-TRUST-SIGNED-PACKAGE-DISCLOSURE)); partial evidence only, with synthetic keys, no cryptographic review and no support claim.

<!-- requirement-ref: REQ-TRUST-0035 -->
<!-- requirement-ref: REQ-TRUST-0052 -->

- [REQ-TRUST-0035](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0035) → `CC-TRUST-SIGNED-PACKAGE-DISCLOSURE`.
- [REQ-TRUST-0052](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0052) → `CC-TRUST-SIGNED-PACKAGE-DISCLOSURE`.

### CC-TRUST-SIGNED-PACKAGE-RETENTION

Public source: `library/task-trust.md`, section **Signed Remote Task Packages**, and `library/registered-limits.md`. Intended target: `internal/admission/signed.go`. Case location: `internal/admission/signed_test.go`, `symbol:TestCcTrustSignedPackageNotRetained`, executed by DEV-612.

Case plan: Admit a synthetic signed package through a fake delivery store and the test replay store, then assert the replay record holds the remote issuer, the task_id and the SHA-256 digest of the exact task bytes, not of the package, and that neither the record nor the dumped replay store contains the key ID, the encoded signature, the package bytes or the package digest in hex or base64. Assert the remote delivery record no longer resolves after admission, that a package re-signed later over the same task bytes and presented through a new reference resolves to the same admitted identity as a redelivery, and that a package with a corrupted signature leaves no replay record, staged write or resolvable delivery record. Requirements: REQ-TRUST-0053.

Scope: Synthetic core contract cases. Passed at Go `e663afe` and re-executed at Go `baa3ada`, `d309b5b`, `19b5810`, `c9516c6`, `daeb718`, `8156c1a` and `2a94d00` ([evidence](../traceability/evidence/go-unit-2a94d00.md#CC-TRUST-SIGNED-PACKAGE-RETENTION)); partial evidence only, with synthetic keys, no cryptographic review and no support claim.

<!-- requirement-ref: REQ-TRUST-0053 -->

- [REQ-TRUST-0053](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0053) → `CC-TRUST-SIGNED-PACKAGE-RETENTION`.

### CC-TRUST-KEY-ROTATION

Public source: `library/task-trust.md`, section **Local Policy Lifecycle**. Intended target: the root `Client` (`policy.go`) and `internal/admission/signed.go`. Case location: `policy_test.go`, `symbol:TestCcTrustKeyRotationByPolicyReplacement`, executed by DEV-612.

Case plan: Initialize a client with a generation-2 policy holding two keys of the remote issuer and admit one synthetic task signed with each. Replace it with generation 3 holding only the second key and assert every task handle and queued delivery reference is invalidated, that the task signed with the removed key and a new task signed with it are rejected as task_source_unauthenticated on redelivery, and that the task signed with the retained key resolves to its retained acceptance without a new replay record. Replace it with generation 4 without the remote issuer and assert its handles are invalidated and its packages rejected with no replay write. Assert that a policy whose key names an issuer it does not authorize is not constructed and the active policy is unchanged. With a key valid until one hour after acceptance, admit a task valid for three hours, advance the clock two hours, and assert the handle still resolves with its original expiry while a new presentation of the same package is rejected as task_source_unauthenticated and the handle still resolves after that rejection. With the same key validity, replace the policy after two hours with generation 3 keeping the expired key and assert the handle is invalidated and a redelivery of the same package is rejected as task_source_unauthenticated with no replay write. With an expiring key and a second valid key of the same issuer, advance two hours and assert the same task bytes re-signed with the valid key resolve to the held handle, then after a generation-3 replacement assert the original package is rejected as task_source_unauthenticated and the re-signed package resolves to the retained acceptance as a redelivery with its original expiry and no new replay record. Requirements: REQ-TRUST-0030, REQ-TRUST-0051, REQ-TRUST-0054, REQ-TRUST-0055.

Scope: Synthetic core contract cases. Passed at Go `e663afe` and re-executed at Go `baa3ada`, `d309b5b`, `19b5810`, `c9516c6`, `daeb718`, `8156c1a` and `2a94d00` ([evidence](../traceability/evidence/go-unit-2a94d00.md#CC-TRUST-KEY-ROTATION)); partial evidence only, with synthetic keys, no cryptographic review and no support claim.

<!-- requirement-ref: REQ-TRUST-0030 -->
<!-- requirement-ref: REQ-TRUST-0051 -->
<!-- requirement-ref: REQ-TRUST-0054 -->
<!-- requirement-ref: REQ-TRUST-0055 -->

- [REQ-TRUST-0030](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0030) → `CC-TRUST-KEY-ROTATION`.
- [REQ-TRUST-0051](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0051) → `CC-TRUST-KEY-ROTATION`.
- [REQ-TRUST-0054](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0054) → `CC-TRUST-KEY-ROTATION`.
- [REQ-TRUST-0055](https://github.com/tv-rain/barometer-spec/blob/4743b636f29feb1f0f8d042866d04ccda3532938/library/task-trust.md#REQ-TRUST-0055) → `CC-TRUST-KEY-ROTATION`.

### CC-JSON-UNIT-0001

Public source: `data-formats/json-decoding.md`, section **Unit 0001**. Intended target: `internal/jsondecode/decode.go`. Planned case location: `internal/jsondecode/cross_cutting_conformance_test.go`, `symbol:TestCcJsonUnit0001`.

Case plan: Accept exactly one schema-typed JSON value with permitted whitespace; reject empty, BOM, extra values/content, comments, trailing commas and non-object roots for object schemas.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-JSON-0001 -->

- [REQ-JSON-0001](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/json-decoding.md#REQ-JSON-0001) → `CC-JSON-UNIT-0001`.

### CC-JSON-UNIT-0002

Public source: `data-formats/json-decoding.md`, section **Unit 0002**. Intended target: `internal/jsondecode/decode.go`. Planned case location: `internal/jsondecode/cross_cutting_conformance_test.go`, `symbol:TestCcJsonUnit0002`.

Case plan: Test malformed UTF-8, escapes and unpaired surrogates; accept valid scalar pairs and preserve distinct non-normalized/case-sensitive strings and member names.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-JSON-0002 -->

- [REQ-JSON-0002](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/json-decoding.md#REQ-JSON-0002) → `CC-JSON-UNIT-0002`.

### CC-JSON-UNIT-0003

Public source: `data-formats/json-decoding.md`, section **Unit 0003**. Intended target: `internal/jsondecode/decode.go`. Planned case location: `internal/jsondecode/cross_cutting_conformance_test.go`, `symbol:TestCcJsonUnit0003`.

Case plan: Reject equal, conflicting and escaped-equivalent duplicate names at every nesting level; allow same names in distinct objects and distinct Unicode scalar sequences.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-JSON-0003 -->

- [REQ-JSON-0003](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/json-decoding.md#REQ-JSON-0003) → `CC-JSON-UNIT-0003`.

### CC-JSON-UNIT-0004

Public source: `data-formats/json-decoding.md`, section **Unit 0004**. Intended target: `internal/jsondecode/decode.go`. Planned case location: `internal/jsondecode/cross_cutting_conformance_test.go`, `symbol:TestCcJsonUnit0004`.

Case plan: Test exact escaped fixed names, case variants and unknown closed-schema fields while allowing only explicitly defined dynamic keys.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-JSON-0004 -->

- [REQ-JSON-0004](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/json-decoding.md#REQ-JSON-0004) → `CC-JSON-UNIT-0004`.

### CC-JSON-UNIT-0005

Public source: `data-formats/json-decoding.md`, section **Unit 0005**. Intended target: `internal/jsondecode/decode.go`. Planned case location: `internal/jsondecode/cross_cutting_conformance_test.go`, `symbol:TestCcJsonUnit0005`.

Case plan: Parameterize missing/null/empty/zero and every wrong JSON type; apply only explicitly allowed null/default behavior without coercion.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-JSON-0005 -->

- [REQ-JSON-0005](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/json-decoding.md#REQ-JSON-0005) → `CC-JSON-UNIT-0005`.

### CC-JSON-UNIT-0006

Public source: `data-formats/json-decoding.md`, section **Unit 0006**. Intended target: `internal/jsondecode/decode.go`. Planned case location: `internal/jsondecode/cross_cutting_conformance_test.go`, `symbol:TestCcJsonUnit0006`.

Case plan: Test every integer lexical form including -0, decimals, exponents, signs and leading zeroes; accept canonical integral tokens only and then apply domain rules.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-JSON-0006 -->

- [REQ-JSON-0006](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/json-decoding.md#REQ-JSON-0006) → `CC-JSON-UNIT-0006`.

### CC-JSON-UNIT-0007

Public source: `data-formats/json-decoding.md`, section **Unit 0007**. Intended target: `internal/jsondecode/decode.go`. Planned case location: `internal/jsondecode/cross_cutting_conformance_test.go`, `symbol:TestCcJsonUnit0007`.

Case plan: Iterate integer tokens for bounded fields at their exact endpoints and one beyond: port 0, 1, 65535, 65536; rotation_period_seconds 86,399, 86,400, 604,800, 604,801; max_bytes_received 262,144 and 262,145; assert in-range endpoints decode to exactly their decimal value and out-of-range values are rejected. Inject machine-width and floating-point boundary tokens into each integer field: 4294967296 and 4294967297 (2^32, 2^32+1), 9007199254740992 and 9007199254740993 (2^53, 2^53+1), 9223372036854775807 and 9223372036854775808 (2^63-1, 2^63), 18446744073709551617 (2^64+1) and -1 for non-negative fields; assert each is rejected by range or representation, never wrapped, rounded, truncated or clamped to an in-range value. For each bounded integer field, assert both declared domain endpoints are accepted. For each integer field without a finite public upper bound, assert the documented representation maximum is accepted exactly and maximum plus one is rejected. Requirements: REQ-JSON-0007.

Scope: Synthetic core contract cases. No execution evidence or support claim.

Review case `RV-CC-JSON-UNIT-0007` (`docs/conformance/cross-cutting-review.md`, `heading:RV-CC-JSON-UNIT-0007`): Human review case; not automated coverage. Reviewer role: Data-format reviewer (independent of the decoder implementer). Reviewed artifact: Go decoder integer-representation documentation and the field inventory of public integer fields at the reviewed revision. Checklist: (1) Does the documented exact representation for every bounded integer field cover the field's entire declared domain? (2) Is every integer field lacking a finite public upper bound identified, with a documented and reviewed finite exact representation? (3) Does the field inventory used by the automated case include every integer field in the pinned public schemas? Evidence record: A `review` evidence record naming the linked requirement IDs, the reviewed artifact revision, each checklist answer and any finding entered in Specification Issues.

<!-- requirement-ref: REQ-JSON-0007 -->

- [REQ-JSON-0007](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/json-decoding.md#REQ-JSON-0007) → `CC-JSON-UNIT-0007`, `RV-CC-JSON-UNIT-0007`.

### CC-JSON-UNIT-0008

Public source: `data-formats/json-decoding.md`, section **Unit 0008**. Intended target: `internal/jsondecode/decode.go`. Planned case location: `internal/jsondecode/cross_cutting_conformance_test.go`, `symbol:TestCcJsonUnit0008`.

Case plan: Vary required schema/test version fields and trusted versionless context; reject malformed/unsupported selection without trying another schema or injecting version fields.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-JSON-0008 -->

- [REQ-JSON-0008](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/json-decoding.md#REQ-JSON-0008) → `CC-JSON-UNIT-0008`.

### CC-JSON-UNIT-0009

Public source: `data-formats/json-decoding.md`, section **Unit 0009**. Intended target: `internal/jsondecode/decode.go`. Planned case location: `internal/jsondecode/cross_cutting_conformance_test.go`, `symbol:TestCcJsonUnit0009`.

Case plan: Exercise complete original-byte and nested byte/count envelopes including whitespace/unknown fields; unknown versions cannot expand limits and unbounded models remain disabled.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-JSON-0009 -->

- [REQ-JSON-0009](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/json-decoding.md#REQ-JSON-0009) → `CC-JSON-UNIT-0009`.

### CC-JSON-UNIT-0010

Public source: `data-formats/json-decoding.md`, section **Unit 0010**. Intended target: `internal/jsondecode/decode.go`. Planned case location: `internal/jsondecode/cross_cutting_conformance_test.go`, `symbol:TestCcJsonUnit0010`.

Case plan: Inject malformed/deep/large data under finite resource budgets; assert no usable partial model, admission, scheduling or I/O after errors and no trust claim from structural success.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-JSON-0010 -->

- [REQ-JSON-0010](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/json-decoding.md#REQ-JSON-0010) → `CC-JSON-UNIT-0010`.

### CC-JSON-UNIT-0011

Public source: `data-formats/json-decoding.md`, section **Unit 0011**. Intended target: `internal/jsondecode/decode.go`. Planned case location: `internal/jsondecode/cross_cutting_conformance_test.go`, `symbol:TestCcJsonUnit0011`.

Case plan: Inject sensitive canaries including names/targets/keys/tokens and raw parser/platform errors; assert bounded deterministic sanitized operational errors without Measurement Results.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-JSON-0011 -->

- [REQ-JSON-0011](https://github.com/tv-rain/barometer-spec/blob/d93aff93cbc17e7c065fb40d8f6e2e01a697bf35/data-formats/json-decoding.md#REQ-JSON-0011) → `CC-JSON-UNIT-0011`.

