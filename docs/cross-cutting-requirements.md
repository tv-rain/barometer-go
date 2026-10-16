# Cross-cutting Requirements and Case Plan

## Status, Authority, and Baselines

This planning map, created by DEV-471 and rebaselined by DEV-568, identifies shared format and library obligations, intended implementation owners and concrete conformance case families. The linked public text is the oracle; this document neither copies nor changes its normative behavior. All case families below are **planned**; DEV-569 links existing unit tests as separate partial evidence, described below. No executed family case, complete semantic review, implemented requirement, core/platform/field support or release is claimed. A case family is a parameterized suite: exercise each linked unit independently, not a single representative assertion for the whole section.

The immutable public baseline is `a944ed6c389fec9f0d089f0044ead6fe845328b1` (Draft `0.1.0`), the verified merge of [public PR #8](https://github.com/tv-rain/barometer-spec/pull/8). The [manifest](../traceability/cross-cutting.json) maps all 573 active allocated shared requirements to the 191 case families below; public links are pinned to that merge commit. DEV-568 selected Go baseline `cc7e0a4c9a72e6f5d309af6faa94eaf457dafd65`, preceding its planning-only metadata. The report must retain the separately scoped registered-test omissions and unallocated-source/inventory gaps.

DEV-569 moves the Go baseline to `c5f4e92e06936cb677c7a03e9769ff75b1f51874` and links existing Foundation code and tests as partial evidence. The 115 shared requirements cited by reviewed `requirement-ref` comments are `partial`: each maps the existing declarations that implement part of the obligation and one or more `UT.*` cases. A `UT.*` case selects an existing top-level Go unit test, passed at that baseline, with a conformance record in the [unit-test evidence report](../traceability/evidence/go-unit-c5f4e92.md). That report states the narrow aspect each test exercises; the requirement notes state what remains. The planned families below stay planned and unexecuted and are still mapped, so every requirement retains its missing-conformance gap. The other 458 requirements stay `planned`. A planned location becomes existing only where review found that its file implements part of that obligation; otherwise it stays planned even where the file exists. The [DEV-568 mapping](../traceability/historical/cross-cutting-a944ed6-cc7e0a4.json) at Go baseline `cc7e0a4c9a72e6f5d309af6faa94eaf457dafd65` is preserved byte-for-byte with its unchanged [report](../traceability/reports/cross-cutting-a944ed6.json); the current [report](../traceability/reports/cross-cutting-a944ed6-c5f4e92.json) is a new baseline-dated record.

The earlier DEV-471 mapping at public baseline `47d75925b3829fccfb6954b545c15f695375b76f` and Go baseline `a9136f34fad00ecc2171cbc29e835c5c50b340df` is preserved byte-for-byte as [historical manifest](../traceability/historical/cross-cutting-47d7592.json) with its unchanged [report](../traceability/reports/cross-cutting.json). It mapped 565 requirements to 185 families. It remains valid metadata only for those original baselines and is not a current claim.

## Boundaries and Maintenance

Public PR #8 resolved the DEV-481 [vocabulary finding](specification-issues.md#go-spec-issue-002-missing-registered-observation-vocabularies) by retiring REQ-OBS-0011 and REQ-OBS-0021 in favor of REQ-OBS-0042 and REQ-OBS-0043, and adding REQ-OBS-0044 to REQ-OBS-0051. The successors take over their predecessors' case families; the new requirements use six `CC-OBS-VOCAB-*` families. The retired IDs keep `requirement-history` comments and links to the original PR #6 baseline only. They are not current obligations, and no prior claim is rewritten to a successor. Mapping these requirements establishes no runtime vocabulary adoption evidence; GO-SPEC-ISSUE-002 remains open for that downstream evidence.

The 16 format/library sources plus shared JSON decoding are in scope. Registered test-specific schemas/runners, the full threat-model and measurement-limitations inventories, implementation of these cases, native adapters, host and collector software, and field validation remain outside this mapping task. Their applicability is preserved as report gaps, not silently treated as covered. The [normative source inventory](normative-sources.md) supplies the wider source boundary; this map also includes the subsequently adopted JSON contract.

Portable Go owns its package targets. Native storage and binding obligations map to planned native validation contracts in this repository, consumed by future Android/Apple adapter work; these documents cannot implement platform behavior. Host, collector and public authoring obligations map to planned integration/review records. Mixed core/integration case families must retain both scopes; Go tests alone cannot close them. Synthetic requests and canary values are local fixtures, not measurements or participant data.

Unresolved contract findings use [Specification Issues](specification-issues.md). `GO-SPEC-ISSUE-001` remains open for downstream evidence despite the accepted JSON clarification. Mapping a planned test never resolves it. A newly discovered ambiguity must block only the affected claim and enter that workflow before implementation selects behavior.

Public prerequisite ordering is satisfied: PR #6 merged the approved allocation head `9dbd72af37205a08484435cb561a168ac7b1876b` as merge commit `47d75925b3829fccfb6954b545c15f695375b76f`, and PR #8 merged approved head `5869bc314c877e2ccd3426b82aadcd9c92414f02` as `a944ed6c389fec9f0d089f0044ead6fe845328b1`. Between those baselines, the shared catalog changes only by the REQ-OBS retirements and additions above. Any later public change requires a renewed source/catalog review, an explicit baseline update, and regenerated traceability report. Preserve previously committed historical manifests/reports rather than presenting old evidence as current. Keep partial coverage while wider source, test-specific or semantic inventory gaps remain.

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

Case plan: Review output/schema APIs for ordered bounded normalized stage observations without raw traffic or local censorship classifications.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0001 -->
<!-- requirement-ref: REQ-OBS-0002 -->

- [REQ-OBS-0001](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0001) → `CC-OBS-OVERVIEW`.
- [REQ-OBS-0002](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0002) → `CC-OBS-OVERVIEW`.

### CC-OBS-COMMON-STRUCTURE

Public source: `data-formats/measurement-observations.md`, section **Common Structure**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsCommonStructure`.

Case plan: Test common field presence/types/status-dependent failure, unknown fields and allowed test-specific observation types. Preserve order and omit never-attempted operations.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0003 -->
<!-- requirement-ref: REQ-OBS-0004 -->
<!-- requirement-ref: REQ-OBS-0005 -->
<!-- requirement-ref: REQ-OBS-0006 -->

- [REQ-OBS-0003](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0003) → `CC-OBS-COMMON-STRUCTURE`.
- [REQ-OBS-0004](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0004) → `CC-OBS-COMMON-STRUCTURE`.
- [REQ-OBS-0005](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0005) → `CC-OBS-COMMON-STRUCTURE`.
- [REQ-OBS-0006](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0006) → `CC-OBS-COMMON-STRUCTURE`.

### CC-OBS-DNS-QUERY-FAILURES

Public source: `data-formats/measurement-observations.md`, section **DNS Query / Failures**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsDnsQueryFailures`.

Case plan: Test each normalized DNS failure and observable response-code pairing; a system resolver with no code must not invent one.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0007 -->
<!-- requirement-ref: REQ-OBS-0008 -->

- [REQ-OBS-0007](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0007) → `CC-OBS-DNS-QUERY-FAILURES`.
- [REQ-OBS-0008](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0008) → `CC-OBS-DNS-QUERY-FAILURES`.

### CC-OBS-DNS-QUERY-DETAILS

Public source: `data-formats/measurement-observations.md`, section **DNS Query / Details**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsDnsQueryDetails`.

Case plan: Exercise engine/family/query-type/response-count fields and canary prohibited resolver/raw-message fields.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0009 -->
<!-- requirement-ref: REQ-OBS-0010 -->

- [REQ-OBS-0009](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0009) → `CC-OBS-DNS-QUERY-DETAILS`.
- [REQ-OBS-0010](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0010) → `CC-OBS-DNS-QUERY-DETAILS`.

### CC-OBS-DNS-QUERY-DNS-RESPONSE

Public source: `data-formats/measurement-observations.md`, section **DNS Query / DNS Response**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsDnsQueryDnsResponse`.

Case plan: Exercise sequential arrival indexes, quantized timing, closed normalized rcode values and bounded answer/CNAME counts.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0042 -->
<!-- requirement-history: REQ-OBS-0011 -->

- [REQ-OBS-0042](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0042) → `CC-OBS-DNS-QUERY-DNS-RESPONSE`. Successor of retired [REQ-OBS-0011](https://github.com/tv-rain/barometer-spec/blob/47d75925b3829fccfb6954b545c15f695375b76f/data-formats/measurement-observations.md#REQ-OBS-0011), whose mapping is preserved only in the historical manifest.

### CC-OBS-DNS-QUERY-DNS-ANSWER

Public source: `data-formats/measurement-observations.md`, section **DNS Query / DNS Answer**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsDnsQueryDnsAnswer`.

Case plan: Test canonical A/AAAA/CNAME values, first-occurrence deduplication, TTL omission, reviewed-query limits and no on-device enrichment.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0012 -->
<!-- requirement-ref: REQ-OBS-0013 -->
<!-- requirement-ref: REQ-OBS-0014 -->

- [REQ-OBS-0012](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0012) → `CC-OBS-DNS-QUERY-DNS-ANSWER`.
- [REQ-OBS-0013](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0013) → `CC-OBS-DNS-QUERY-DNS-ANSWER`.
- [REQ-OBS-0014](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0014) → `CC-OBS-DNS-QUERY-DNS-ANSWER`.

### CC-OBS-TCP-CONNECT-FAILURES

Public source: `data-formats/measurement-observations.md`, section **TCP Connect / Failures**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsTcpConnectFailures`.

Case plan: Map each TCP failure conservatively and keep outer cancellation/network/proxy/deadline conditions exclusively in execution.reason.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0015 -->
<!-- requirement-ref: REQ-OBS-0016 -->

- [REQ-OBS-0015](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0015) → `CC-OBS-TCP-CONNECT-FAILURES`.
- [REQ-OBS-0016](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0016) → `CC-OBS-TCP-CONNECT-FAILURES`.

### CC-OBS-TCP-CONNECT-DETAILS

Public source: `data-formats/measurement-observations.md`, section **TCP Connect / Details**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsTcpConnectDetails`.

Case plan: Validate address family and exclude target address/port canaries.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0017 -->
<!-- requirement-ref: REQ-OBS-0018 -->

- [REQ-OBS-0017](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0017) → `CC-OBS-TCP-CONNECT-DETAILS`.
- [REQ-OBS-0018](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0018) → `CC-OBS-TCP-CONNECT-DETAILS`.

### CC-OBS-TLS-HANDSHAKE-FAILURES

Public source: `data-formats/measurement-observations.md`, section **TLS Handshake / Failures**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsTlsHandshakeFailures`.

Case plan: Test all registered TLS failures, preceding TCP failure and no TLS observation before TLS activity.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0019 -->
<!-- requirement-ref: REQ-OBS-0020 -->

- [REQ-OBS-0019](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0019) → `CC-OBS-TLS-HANDSHAKE-FAILURES`.
- [REQ-OBS-0020](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0020) → `CC-OBS-TLS-HANDSHAKE-FAILURES`.

### CC-OBS-TLS-HANDSHAKE-DETAILS

Public source: `data-formats/measurement-observations.md`, section **TLS Handshake / Details**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsTlsHandshakeDetails`.

Case plan: Exercise conditional negotiated fields with closed version/ALPN/alert vocabularies and bounded ALPN; exclude every prohibited certificate/handshake/key/session field.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0043 -->
<!-- requirement-history: REQ-OBS-0021 -->
<!-- requirement-ref: REQ-OBS-0022 -->

- [REQ-OBS-0043](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0043) → `CC-OBS-TLS-HANDSHAKE-DETAILS`. Successor of retired [REQ-OBS-0021](https://github.com/tv-rain/barometer-spec/blob/47d75925b3829fccfb6954b545c15f695375b76f/data-formats/measurement-observations.md#REQ-OBS-0021), whose mapping is preserved only in the historical manifest.
- [REQ-OBS-0022](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0022) → `CC-OBS-TLS-HANDSHAKE-DETAILS`.

### CC-OBS-HTTP-EXCHANGE-FAILURES

Public source: `data-formats/measurement-observations.md`, section **HTTP Exchange / Failures**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsHttpExchangeFailures`.

Case plan: Test registered HTTP failures and preceding DNS/TCP/TLS observations; emit HTTP only once HTTP activity starts.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0023 -->
<!-- requirement-ref: REQ-OBS-0024 -->

- [REQ-OBS-0023](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0023) → `CC-OBS-HTTP-EXCHANGE-FAILURES`.
- [REQ-OBS-0024](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0024) → `CC-OBS-HTTP-EXCHANGE-FAILURES`.

### CC-OBS-HTTP-EXCHANGE-DETAILS

Public source: `data-formats/measurement-observations.md`, section **HTTP Exchange / Details**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsHttpExchangeDetails`.

Case plan: Validate conditional status/redirect/body/checks/content-encoding fields and count limits.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0025 -->

- [REQ-OBS-0025](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0025) → `CC-OBS-HTTP-EXCHANGE-DETAILS`.

### CC-OBS-HTTP-EXCHANGE-REDIRECT

Public source: `data-formats/measurement-observations.md`, section **HTTP Exchange / Redirect**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsHttpExchangeRedirect`.

Case plan: Test redirect counts/classes including uncontacted rejected redirects; exclude full destination data.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0026 -->
<!-- requirement-ref: REQ-OBS-0027 -->

- [REQ-OBS-0026](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0026) → `CC-OBS-HTTP-EXCHANGE-REDIRECT`.
- [REQ-OBS-0027](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0027) → `CC-OBS-HTTP-EXCHANGE-REDIRECT`.

### CC-OBS-HTTP-EXCHANGE-BODY

Public source: `data-formats/measurement-observations.md`, section **HTTP Exchange / Body**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsHttpExchangeBody`.

Case plan: Test processed-byte/truncation boundaries and omit raw body/digests.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0028 -->
<!-- requirement-ref: REQ-OBS-0029 -->

- [REQ-OBS-0028](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0028) → `CC-OBS-HTTP-EXCHANGE-BODY`.
- [REQ-OBS-0029](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0029) → `CC-OBS-HTTP-EXCHANGE-BODY`.

### CC-OBS-HTTP-EXCHANGE-CHECK

Public source: `data-formats/measurement-observations.md`, section **HTTP Exchange / Check**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsHttpExchangeCheck`.

Case plan: Test kind/index/matched associations without retaining expected or actual header/body/digest values.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0030 -->
<!-- requirement-ref: REQ-OBS-0031 -->

- [REQ-OBS-0030](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0030) → `CC-OBS-HTTP-EXCHANGE-CHECK`.
- [REQ-OBS-0031](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0031) → `CC-OBS-HTTP-EXCHANGE-CHECK`.

### CC-OBS-VOCAB-CONTRACT-SELECTION

Public source: `data-formats/measurement-observations.md`, section **Registered Normalized Vocabularies / Contract Selection and Producer Boundary**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsVocabContractSelection`.

Case plan: Test that vocabularies are closed with no runtime registry lookup; decoders reject aliases, wrong case, numeric strings and raw library messages without repairing them into coarse markers; keep missing evidence, malformed data and out-of-subset values distinct.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0044 -->
<!-- requirement-ref: REQ-OBS-0045 -->

- [REQ-OBS-0044](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0044) → `CC-OBS-VOCAB-CONTRACT-SELECTION`.
- [REQ-OBS-0045](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0045) → `CC-OBS-VOCAB-CONTRACT-SELECTION`.

### CC-OBS-VOCAB-DNS-RESPONSE-CODE

Public source: `data-formats/measurement-observations.md`, section **Registered Normalized Vocabularies / DNS Response-Code Normalization**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsVocabDnsResponseCode`.

Case plan: Map each effective code 0-5 and an out-of-subset code including extended bits; omit responses without a reliable code; exclude raw codes/diagnostics; keep OTHER distinct from success/no-data and map unregistered nonzero codes conservatively to inconclusive.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0046 -->
<!-- requirement-ref: REQ-OBS-0047 -->

- [REQ-OBS-0046](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0046) → `CC-OBS-VOCAB-DNS-RESPONSE-CODE`.
- [REQ-OBS-0047](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0047) → `CC-OBS-VOCAB-DNS-RESPONSE-CODE`.

### CC-OBS-VOCAB-TLS-NEGOTIATED-PROPERTIES

Public source: `data-formats/measurement-observations.md`, section **Registered Normalized Vocabularies / TLS Negotiated Properties**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsVocabTlsNegotiatedProperties`.

Case plan: Map negotiated 0x0303/0x0304/other versions and http/1.1, h2, dot and other negotiated ALPN; absent fields when nothing reliable was negotiated; ignore record-layer/offered versions; never copy unregistered identifiers or let coarse markers override failures.

Scope: Synthetic core contract cases. No execution evidence or support claim. REQ-OBS-0049 also maps to `CC-OBS-VOCAB-TLS-OFFERED-ALPN` for offered-identifier admission.

<!-- requirement-ref: REQ-OBS-0048 -->
<!-- requirement-ref: REQ-OBS-0049 -->

- [REQ-OBS-0048](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0048) → `CC-OBS-VOCAB-TLS-NEGOTIATED-PROPERTIES`.
- [REQ-OBS-0049](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0049) → `CC-OBS-VOCAB-TLS-NEGOTIATED-PROPERTIES`.

### CC-OBS-VOCAB-TLS-OFFERED-ALPN

Public source: `data-formats/measurement-observations.md`, section **Registered Normalized Vocabularies / TLS Negotiated Properties**. Intended target: `internal/validation/options.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcObsVocabTlsOfferedAlpn`.

Case plan: Accept only offered http/1.1, h2 and dot; reject unknown, h2c, case variants and other registry entries alongside existing count, length and duplicate bounds.

Scope: Synthetic core contract cases. No execution evidence or support claim. This family covers the offered-identifier half of REQ-OBS-0049; negotiated output is covered by `CC-OBS-VOCAB-TLS-NEGOTIATED-PROPERTIES`.

<!-- requirement-ref: REQ-OBS-0049 -->

- [REQ-OBS-0049](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0049) → `CC-OBS-VOCAB-TLS-OFFERED-ALPN`.

### CC-OBS-VOCAB-TLS-PEER-ALERT

Public source: `data-formats/measurement-observations.md`, section **Registered Normalized Vocabularies / TLS Peer-Alert Normalization**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsVocabTlsPeerAlert`.

Case plan: Map every listed received peer description code and an unlisted code; omit alert when the description is unavailable; exclude local cancellation, deadline, validation failures, sent alerts and post-handshake closure; serialize no raw alert data.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0050 -->

- [REQ-OBS-0050](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0050) → `CC-OBS-VOCAB-TLS-PEER-ALERT`.

### CC-OBS-VOCAB-INCOMPLETE-OBSERVATIONS

Public source: `data-formats/measurement-observations.md`, section **Registered Normalized Vocabularies / Incomplete Observations**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsVocabIncompleteObservations`.

Case plan: For each observation type, omit failure on incomplete status and reject null, empty or registered failure values; keep termination in execution.reason.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0051 -->

- [REQ-OBS-0051](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0051) → `CC-OBS-VOCAB-INCOMPLETE-OBSERVATIONS`.

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

- [REQ-OBS-0032](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0032) → `CC-OBS-VALIDATION-AND-BOUNDS`.
- [REQ-OBS-0034](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0034) → `CC-OBS-VALIDATION-AND-BOUNDS`.
- [REQ-OBS-0035](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0035) → `CC-OBS-VALIDATION-AND-BOUNDS`.
- [REQ-OBS-0036](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0036) → `CC-OBS-VALIDATION-AND-BOUNDS`.
- [REQ-OBS-0037](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0037) → `CC-OBS-VALIDATION-AND-BOUNDS`.
- [REQ-OBS-0038](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0038) → `CC-OBS-VALIDATION-AND-BOUNDS`.
- [REQ-OBS-0039](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0039) → `CC-OBS-VALIDATION-AND-BOUNDS`.
- [REQ-OBS-0040](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0040) → `CC-OBS-VALIDATION-AND-BOUNDS`.
- [REQ-OBS-0041](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0041) → `CC-OBS-VALIDATION-AND-BOUNDS`.

### CC-OBS-PRIVACY-AND-INTERPRETATION

Public source: `data-formats/measurement-observations.md`, section **Privacy and Interpretation**. Intended target: `internal/result/observations.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcObsPrivacyAndInterpretation`.

Case plan: Review downstream examples for source-layer interpretation without attributing intent or definitive interference.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OBS-0033 -->

- [REQ-OBS-0033](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-observations.md#REQ-OBS-0033) → `CC-OBS-PRIVACY-AND-INTERPRETATION`.

### CC-RESULT-OVERVIEW

Public source: `data-formats/measurement-result.md`, section **Overview**. Intended target: `internal/result/result.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcResultOverview`.

Case plan: Feed negative DNS/TCP/TLS/HTTP terminal observations; produce completed measurements without automatic retries.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-RESULT-0001 -->

- [REQ-RESULT-0001](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0001) → `CC-RESULT-OVERVIEW`.

### CC-RESULT-TOP-LEVEL-FIELDS

Public source: `data-formats/measurement-result.md`, section **Top-level Fields**. Intended target: `internal/result/result.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcResultTopLevelFields`.

Case plan: Exercise required/conditional fields, encoded bytes/observation counts and task_id association. Exclude automatic task-field copies and reversible input digests.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-RESULT-0002 -->
<!-- requirement-ref: REQ-RESULT-0003 -->
<!-- requirement-ref: REQ-RESULT-0004 -->
<!-- requirement-ref: REQ-RESULT-0005 -->

- [REQ-RESULT-0002](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0002) → `CC-RESULT-TOP-LEVEL-FIELDS`.
- [REQ-RESULT-0003](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0003) → `CC-RESULT-TOP-LEVEL-FIELDS`.
- [REQ-RESULT-0004](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0004) → `CC-RESULT-TOP-LEVEL-FIELDS`.
- [REQ-RESULT-0005](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0005) → `CC-RESULT-TOP-LEVEL-FIELDS`.

### CC-RESULT-RUN-AND-ATTEMPT-IDENTITY

Public source: `data-formats/measurement-result.md`, section **Run and Attempt Identity**. Intended target: `internal/result/result.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcResultRunAndAttemptIdentity`.

Case plan: Inject a controlled random source; new runs get independent IDs, retries preserve ID/increment attempt, retransmission creates no attempt, and IDs never escape to measurement/context/log sinks.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-RESULT-0006 -->
<!-- requirement-ref: REQ-RESULT-0007 -->
<!-- requirement-ref: REQ-RESULT-0008 -->
<!-- requirement-ref: REQ-RESULT-0009 -->

- [REQ-RESULT-0006](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0006) → `CC-RESULT-RUN-AND-ATTEMPT-IDENTITY`.
- [REQ-RESULT-0007](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0007) → `CC-RESULT-RUN-AND-ATTEMPT-IDENTITY`.
- [REQ-RESULT-0008](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0008) → `CC-RESULT-RUN-AND-ATTEMPT-IDENTITY`.
- [REQ-RESULT-0009](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0009) → `CC-RESULT-RUN-AND-ATTEMPT-IDENTITY`.

### CC-RESULT-TEST-OBJECT

Public source: `data-formats/measurement-result.md`, section **Test Object**. Intended target: `internal/result/result.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcResultTestObject`.

Case plan: Select exact registered test/result-version tuples independently; reject unknown fields and unsupported combinations.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-RESULT-0010 -->
<!-- requirement-ref: REQ-RESULT-0011 -->
<!-- requirement-ref: REQ-RESULT-0012 -->

- [REQ-RESULT-0010](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0010) → `CC-RESULT-TEST-OBJECT`.
- [REQ-RESULT-0011](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0011) → `CC-RESULT-TEST-OBJECT`.
- [REQ-RESULT-0012](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0012) → `CC-RESULT-TEST-OBJECT`.

### CC-RESULT-IMPLEMENTATION-OBJECT

Public source: `data-formats/measurement-result.md`, section **Implementation Object**. Intended target: `internal/result/result.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcResultImplementationObject`.

Case plan: Validate bounded released version and platform enum; exclude host/device/build/locale metadata and arbitrary annotations.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-RESULT-0013 -->
<!-- requirement-ref: REQ-RESULT-0014 -->
<!-- requirement-ref: REQ-RESULT-0015 -->

- [REQ-RESULT-0013](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0013) → `CC-RESULT-IMPLEMENTATION-OBJECT`.
- [REQ-RESULT-0014](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0014) → `CC-RESULT-IMPLEMENTATION-OBJECT`.
- [REQ-RESULT-0015](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0015) → `CC-RESULT-IMPLEMENTATION-OBJECT`.

### CC-RESULT-TIME-OBJECT

Public source: `data-formats/measurement-result.md`, section **Time Object**. Intended target: `internal/result/result.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcResultTimeObject`.

Case plan: Use a monotonic fake clock plus wall-clock jumps; test minute floor, 250ms ceiling including zero, stage rounding and duration_capped conditional behavior.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-RESULT-0016 -->
<!-- requirement-ref: REQ-RESULT-0017 -->
<!-- requirement-ref: REQ-RESULT-0018 -->

- [REQ-RESULT-0016](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0016) → `CC-RESULT-TIME-OBJECT`.
- [REQ-RESULT-0017](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0017) → `CC-RESULT-TIME-OBJECT`.
- [REQ-RESULT-0018](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0018) → `CC-RESULT-TIME-OBJECT`.

### CC-RESULT-EXECUTION-OBJECT

Public source: `data-formats/measurement-result.md`, section **Execution Object**. Intended target: `internal/result/result.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcResultExecutionObject`.

Case plan: Exercise all states/reasons, first termination precedence, measurement presence/absence/partial rules and separation of malformed task errors from results.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-RESULT-0019 -->
<!-- requirement-ref: REQ-RESULT-0020 -->
<!-- requirement-ref: REQ-RESULT-0021 -->
<!-- requirement-ref: REQ-RESULT-0022 -->

- [REQ-RESULT-0019](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0019) → `CC-RESULT-EXECUTION-OBJECT`.
- [REQ-RESULT-0020](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0020) → `CC-RESULT-EXECUTION-OBJECT`.
- [REQ-RESULT-0021](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0021) → `CC-RESULT-EXECUTION-OBJECT`.
- [REQ-RESULT-0022](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0022) → `CC-RESULT-EXECUTION-OBJECT`.

### CC-RESULT-NETWORK-CONTEXT

Public source: `data-formats/measurement-result.md`, section **Network Context**. Intended target: `internal/result/result.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcResultNetworkContext`.

Case plan: Test completed/incomplete remote tests with unavailable egress, pre-snapshot not_run and NET-01 exceptions; reject prohibited context data.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-RESULT-0023 -->
<!-- requirement-ref: REQ-RESULT-0024 -->
<!-- requirement-ref: REQ-RESULT-0025 -->

- [REQ-RESULT-0023](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0023) → `CC-RESULT-NETWORK-CONTEXT`.
- [REQ-RESULT-0024](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0024) → `CC-RESULT-NETWORK-CONTEXT`.
- [REQ-RESULT-0025](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0025) → `CC-RESULT-NETWORK-CONTEXT`.

### CC-RESULT-MEASUREMENT-OBJECT

Public source: `data-formats/measurement-result.md`, section **Measurement Object**. Intended target: `internal/result/result.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcResultMeasurementObject`.

Case plan: Select the registered Output schema, ordered observation list and terminal outcome; reject unregistered extensions/verdicts.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-RESULT-0026 -->
<!-- requirement-ref: REQ-RESULT-0027 -->
<!-- requirement-ref: REQ-RESULT-0028 -->

- [REQ-RESULT-0026](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0026) → `CC-RESULT-MEASUREMENT-OBJECT`.
- [REQ-RESULT-0027](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0027) → `CC-RESULT-MEASUREMENT-OBJECT`.
- [REQ-RESULT-0028](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0028) → `CC-RESULT-MEASUREMENT-OBJECT`.

### CC-RESULT-PARTIAL-RESULTS

Public source: `data-formats/measurement-result.md`, section **Partial Results**. Intended target: `internal/result/result.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcResultPartialResults`.

Case plan: Interrupt each stage of NET-05 and primitive tests; retain only schema-permitted partial data and distinguish skipped from failed operations.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-RESULT-0029 -->
<!-- requirement-ref: REQ-RESULT-0030 -->

- [REQ-RESULT-0029](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0029) → `CC-RESULT-PARTIAL-RESULTS`.
- [REQ-RESULT-0030](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0030) → `CC-RESULT-PARTIAL-RESULTS`.

### CC-RESULT-VERSIONING-AND-EXTENSION

Public source: `data-formats/measurement-result.md`, section **Versioning and Extension**. Intended target: `internal/result/result.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcResultVersioningAndExtension`.

Case plan: Reject unknown known-version fields and unsupported exact versions; quarantine, if provided, cannot analyze a silently reduced object. Reject host/task injected extensions.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-RESULT-0031 -->
<!-- requirement-ref: REQ-RESULT-0032 -->
<!-- requirement-ref: REQ-RESULT-0033 -->

- [REQ-RESULT-0031](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0031) → `CC-RESULT-VERSIONING-AND-EXTENSION`.
- [REQ-RESULT-0032](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0032) → `CC-RESULT-VERSIONING-AND-EXTENSION`.
- [REQ-RESULT-0033](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0033) → `CC-RESULT-VERSIONING-AND-EXTENSION`.

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

- [REQ-RESULT-0034](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0034) → `CC-RESULT-VALIDATION`.
- [REQ-RESULT-0036](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0036) → `CC-RESULT-VALIDATION`.
- [REQ-RESULT-0037](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0037) → `CC-RESULT-VALIDATION`.
- [REQ-RESULT-0038](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0038) → `CC-RESULT-VALIDATION`.
- [REQ-RESULT-0039](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0039) → `CC-RESULT-VALIDATION`.
- [REQ-RESULT-0040](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0040) → `CC-RESULT-VALIDATION`.
- [REQ-RESULT-0041](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0041) → `CC-RESULT-VALIDATION`.
- [REQ-RESULT-0042](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0042) → `CC-RESULT-VALIDATION`.
- [REQ-RESULT-0043](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0043) → `CC-RESULT-VALIDATION`.
- [REQ-RESULT-0044](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0044) → `CC-RESULT-VALIDATION`.

### CC-RESULT-PRIVACY-SAFETY-AND-INTERPRETATION

Public source: `data-formats/measurement-result.md`, section **Privacy, Safety, and Interpretation**. Intended target: `internal/result/result.go`. Planned case location: `internal/result/cross_cutting_conformance_test.go`, `symbol:TestCcResultPrivacySafetyAndInterpretation`.

Case plan: Canary every prohibited identifier/raw-data sink; review aggregate retention and separate versioned derived analysis from source observations.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-RESULT-0035 -->
<!-- requirement-ref: REQ-RESULT-0045 -->
<!-- requirement-ref: REQ-RESULT-0046 -->
<!-- requirement-ref: REQ-RESULT-0047 -->
<!-- requirement-ref: REQ-RESULT-0048 -->
<!-- requirement-ref: REQ-RESULT-0049 -->
<!-- requirement-ref: REQ-RESULT-0050 -->

- [REQ-RESULT-0035](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0035) → `CC-RESULT-PRIVACY-SAFETY-AND-INTERPRETATION`.
- [REQ-RESULT-0045](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0045) → `CC-RESULT-PRIVACY-SAFETY-AND-INTERPRETATION`.
- [REQ-RESULT-0046](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0046) → `CC-RESULT-PRIVACY-SAFETY-AND-INTERPRETATION`.
- [REQ-RESULT-0047](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0047) → `CC-RESULT-PRIVACY-SAFETY-AND-INTERPRETATION`.
- [REQ-RESULT-0048](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0048) → `CC-RESULT-PRIVACY-SAFETY-AND-INTERPRETATION`.
- [REQ-RESULT-0049](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0049) → `CC-RESULT-PRIVACY-SAFETY-AND-INTERPRETATION`.
- [REQ-RESULT-0050](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/measurement-result.md#REQ-RESULT-0050) → `CC-RESULT-PRIVACY-SAFETY-AND-INTERPRETATION`.

### CC-CONTEXT-FORMAT-OVERVIEW

Public source: `data-formats/network-context.md`, section **Overview**. Intended target: `internal/networkcontext/model.go`. Planned case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextFormatOverview`.

Case plan: Keep one access generation and selected path; reject raw IP/persistent context identity.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-CONTEXT-FORMAT-0001 -->

- [REQ-CONTEXT-FORMAT-0001](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0001) → `CC-CONTEXT-FORMAT-OVERVIEW`.

### CC-CONTEXT-FORMAT-TOP-LEVEL-FIELDS

Public source: `data-formats/network-context.md`, section **Top-level Fields**. Intended target: `internal/networkcontext/model.go`. Planned case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextFormatTopLevelFields`.

Case plan: Require access/egress objects and reject unknown top-level fields.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-CONTEXT-FORMAT-0002 -->
<!-- requirement-ref: REQ-CONTEXT-FORMAT-0003 -->

- [REQ-CONTEXT-FORMAT-0002](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0002) → `CC-CONTEXT-FORMAT-TOP-LEVEL-FIELDS`.
- [REQ-CONTEXT-FORMAT-0003](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0003) → `CC-CONTEXT-FORMAT-TOP-LEVEL-FIELDS`.

### CC-CONTEXT-FORMAT-ACCESS-OBJECT

Public source: `data-formats/network-context.md`, section **Access Object**. Intended target: `internal/networkcontext/model.go`. Planned case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextFormatAccessObject`.

Case plan: Exercise exact coarse enums/types and reject every prohibited platform identifier.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-CONTEXT-FORMAT-0004 -->
<!-- requirement-ref: REQ-CONTEXT-FORMAT-0005 -->

- [REQ-CONTEXT-FORMAT-0004](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0004) → `CC-CONTEXT-FORMAT-ACCESS-OBJECT`.
- [REQ-CONTEXT-FORMAT-0005](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0005) → `CC-CONTEXT-FORMAT-ACCESS-OBJECT`.

### CC-CONTEXT-FORMAT-EGRESS-OBJECT

Public source: `data-formats/network-context.md`, section **Egress Object**. Intended target: `internal/networkcontext/model.go`. Planned case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextFormatEgressObject`.

Case plan: Test all status/source/path/connection-code/classification combinations, missing optional classification, and no guessed values.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-CONTEXT-FORMAT-0006 -->
<!-- requirement-ref: REQ-CONTEXT-FORMAT-0007 -->
<!-- requirement-ref: REQ-CONTEXT-FORMAT-0008 -->
<!-- requirement-ref: REQ-CONTEXT-FORMAT-0009 -->

- [REQ-CONTEXT-FORMAT-0006](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0006) → `CC-CONTEXT-FORMAT-EGRESS-OBJECT`.
- [REQ-CONTEXT-FORMAT-0007](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0007) → `CC-CONTEXT-FORMAT-EGRESS-OBJECT`.
- [REQ-CONTEXT-FORMAT-0008](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0008) → `CC-CONTEXT-FORMAT-EGRESS-OBJECT`.
- [REQ-CONTEXT-FORMAT-0009](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0009) → `CC-CONTEXT-FORMAT-EGRESS-OBJECT`.

### CC-CONTEXT-FORMAT-PROVIDER-AND-OPERATOR

Public source: `data-formats/network-context.md`, section **Provider and Operator**. Intended target: `internal/networkcontext/model.go`. Planned case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextFormatProviderAndOperator`.

Case plan: Reject free-form provider/operator data; review analysis labels and any proposed operator extension for required specialist review.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-CONTEXT-FORMAT-0010 -->
<!-- requirement-ref: REQ-CONTEXT-FORMAT-0011 -->
<!-- requirement-ref: REQ-CONTEXT-FORMAT-0012 -->

- [REQ-CONTEXT-FORMAT-0010](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0010) → `CC-CONTEXT-FORMAT-PROVIDER-AND-OPERATOR`.
- [REQ-CONTEXT-FORMAT-0011](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0011) → `CC-CONTEXT-FORMAT-PROVIDER-AND-OPERATOR`.
- [REQ-CONTEXT-FORMAT-0012](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0012) → `CC-CONTEXT-FORMAT-PROVIDER-AND-OPERATOR`.

### CC-CONTEXT-FORMAT-USE-IN-EXECUTION

Public source: `data-formats/network-context.md`, section **Use in Execution**. Intended target: `internal/networkcontext/model.go`. Planned case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextFormatUseInExecution`.

Case plan: Use normalized required fields and country eligibility only; reject region/ASN/provider/operator task allowlists.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-CONTEXT-FORMAT-0013 -->
<!-- requirement-ref: REQ-CONTEXT-FORMAT-0014 -->

- [REQ-CONTEXT-FORMAT-0013](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0013) → `CC-CONTEXT-FORMAT-USE-IN-EXECUTION`.
- [REQ-CONTEXT-FORMAT-0014](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0014) → `CC-CONTEXT-FORMAT-USE-IN-EXECUTION`.

### CC-CONTEXT-FORMAT-USE-IN-RESULTS

Public source: `data-formats/network-context.md`, section **Use in Results**. Intended target: `internal/networkcontext/model.go`. Planned case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextFormatUseInResults`.

Case plan: Test schema-permitted context carriage/deduplication and reject persistent cross-batch identity or listed private details.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-CONTEXT-FORMAT-0015 -->
<!-- requirement-ref: REQ-CONTEXT-FORMAT-0016 -->

- [REQ-CONTEXT-FORMAT-0015](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0015) → `CC-CONTEXT-FORMAT-USE-IN-RESULTS`.
- [REQ-CONTEXT-FORMAT-0016](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0016) → `CC-CONTEXT-FORMAT-USE-IN-RESULTS`.

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

- [REQ-CONTEXT-FORMAT-0017](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0017) → `CC-CONTEXT-FORMAT-VALIDATION`.
- [REQ-CONTEXT-FORMAT-0018](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0018) → `CC-CONTEXT-FORMAT-VALIDATION`.
- [REQ-CONTEXT-FORMAT-0019](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0019) → `CC-CONTEXT-FORMAT-VALIDATION`.
- [REQ-CONTEXT-FORMAT-0020](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0020) → `CC-CONTEXT-FORMAT-VALIDATION`.
- [REQ-CONTEXT-FORMAT-0021](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0021) → `CC-CONTEXT-FORMAT-VALIDATION`.
- [REQ-CONTEXT-FORMAT-0022](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0022) → `CC-CONTEXT-FORMAT-VALIDATION`.
- [REQ-CONTEXT-FORMAT-0023](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/network-context.md#REQ-CONTEXT-FORMAT-0023) → `CC-CONTEXT-FORMAT-VALIDATION`.

### CC-BUNDLE-OVERVIEW

Public source: `data-formats/result-bundle.md`, section **Overview**. Intended target: `internal/bundle/bundle.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcBundleOverview`.

Case plan: Group valid individual results with a single-use bundle ID; rotating correlation is separate and default-off.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-BUNDLE-0001 -->
<!-- requirement-ref: REQ-BUNDLE-0002 -->

- [REQ-BUNDLE-0001](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0001) → `CC-BUNDLE-OVERVIEW`.
- [REQ-BUNDLE-0002](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0002) → `CC-BUNDLE-OVERVIEW`.

### CC-BUNDLE-JSON-STRUCTURE

Public source: `data-formats/result-bundle.md`, section **JSON Structure**. Intended target: `internal/bundle/bundle.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcBundleJsonStructure`.

Case plan: Disable linkage and assert the linkage object is absent.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-BUNDLE-0003 -->

- [REQ-BUNDLE-0003](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0003) → `CC-BUNDLE-JSON-STRUCTURE`.

### CC-BUNDLE-TOP-LEVEL-FIELDS

Public source: `data-formats/result-bundle.md`, section **Top-level Fields**. Intended target: `internal/bundle/bundle.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcBundleTopLevelFields`.

Case plan: Exercise envelope required fields/types, 1..16 results and encoded-byte bound; reject unknown fields.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-BUNDLE-0004 -->
<!-- requirement-ref: REQ-BUNDLE-0005 -->

- [REQ-BUNDLE-0004](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0004) → `CC-BUNDLE-TOP-LEVEL-FIELDS`.
- [REQ-BUNDLE-0005](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0005) → `CC-BUNDLE-TOP-LEVEL-FIELDS`.

### CC-BUNDLE-BUNDLE-IDENTIFIER

Public source: `data-formats/result-bundle.md`, section **Bundle Identifier**. Intended target: `internal/bundle/bundle.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcBundleBundleIdentifier`.

Case plan: Inject random IDs; exact retries preserve identity/content, any membership/content mutation gets a new ID, acknowledgement/expiry deletes state, and no ID enters measurement traffic or targeting.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-BUNDLE-0006 -->
<!-- requirement-ref: REQ-BUNDLE-0007 -->
<!-- requirement-ref: REQ-BUNDLE-0008 -->
<!-- requirement-ref: REQ-BUNDLE-0009 -->

- [REQ-BUNDLE-0006](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0006) → `CC-BUNDLE-BUNDLE-IDENTIFIER`.
- [REQ-BUNDLE-0007](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0007) → `CC-BUNDLE-BUNDLE-IDENTIFIER`.
- [REQ-BUNDLE-0008](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0008) → `CC-BUNDLE-BUNDLE-IDENTIFIER`.
- [REQ-BUNDLE-0009](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0009) → `CC-BUNDLE-BUNDLE-IDENTIFIER`.

### CC-BUNDLE-LINKAGE-OBJECT

Public source: `data-formats/result-bundle.md`, section **Linkage Object**. Intended target: `internal/bundle/bundle.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcBundleLinkageObject`.

Case plan: Validate the exact complete rotating-installation object and local scope, encoded token/epoch, no stable mode or user_id representation.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-BUNDLE-0010 -->
<!-- requirement-ref: REQ-BUNDLE-0011 -->
<!-- requirement-ref: REQ-BUNDLE-0012 -->

- [REQ-BUNDLE-0010](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0010) → `CC-BUNDLE-LINKAGE-OBJECT`.
- [REQ-BUNDLE-0011](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0011) → `CC-BUNDLE-LINKAGE-OBJECT`.
- [REQ-BUNDLE-0012](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0012) → `CC-BUNDLE-LINKAGE-OBJECT`.

### CC-BUNDLE-RESULT-MEMBERSHIP

Public source: `data-formats/result-bundle.md`, section **Result Membership**. Intended target: `internal/bundle/bundle.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcBundleResultMembership`.

Case plan: Bundle mixed completed/negative/not_run/incomplete attempts without changing semantics or adding unregistered context references.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-BUNDLE-0013 -->
<!-- requirement-ref: REQ-BUNDLE-0014 -->

- [REQ-BUNDLE-0013](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0013) → `CC-BUNDLE-RESULT-MEMBERSHIP`.
- [REQ-BUNDLE-0014](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0014) → `CC-BUNDLE-RESULT-MEMBERSHIP`.

### CC-BUNDLE-CONSTRUCTION-AND-DELIVERY

Public source: `data-formats/result-bundle.md`, section **Construction and Delivery**. Intended target: `internal/bundle/bundle.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcBundleConstructionAndDelivery`.

Case plan: Fault-inject validation, persistence, ambiguous delivery and durable acknowledgement; replay identical logical bundles and delete by acknowledgement/expiry, never mutating linkage on retry.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-BUNDLE-0015 -->
<!-- requirement-ref: REQ-BUNDLE-0016 -->

- [REQ-BUNDLE-0015](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0015) → `CC-BUNDLE-CONSTRUCTION-AND-DELIVERY`.
- [REQ-BUNDLE-0016](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0016) → `CC-BUNDLE-CONSTRUCTION-AND-DELIVERY`.

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

- [REQ-BUNDLE-0017](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0017) → `CC-BUNDLE-VALIDATION`.
- [REQ-BUNDLE-0021](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0021) → `CC-BUNDLE-VALIDATION`.
- [REQ-BUNDLE-0022](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0022) → `CC-BUNDLE-VALIDATION`.
- [REQ-BUNDLE-0023](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0023) → `CC-BUNDLE-VALIDATION`.
- [REQ-BUNDLE-0024](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0024) → `CC-BUNDLE-VALIDATION`.
- [REQ-BUNDLE-0025](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0025) → `CC-BUNDLE-VALIDATION`.
- [REQ-BUNDLE-0026](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0026) → `CC-BUNDLE-VALIDATION`.
- [REQ-BUNDLE-0027](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0027) → `CC-BUNDLE-VALIDATION`.
- [REQ-BUNDLE-0028](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0028) → `CC-BUNDLE-VALIDATION`.

### CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS

Public source: `data-formats/result-bundle.md`, section **Privacy and Safety Requirements**. Intended target: `internal/bundle/bundle.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcBundlePrivacyAndSafetyRequirements`.

Case plan: Use synthetic canaries to check prohibited IDs and traffic leakage; review sensitive-token access/retention, aggregation and separation from task selection/transport metadata.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-BUNDLE-0018 -->
<!-- requirement-ref: REQ-BUNDLE-0029 -->
<!-- requirement-ref: REQ-BUNDLE-0030 -->
<!-- requirement-ref: REQ-BUNDLE-0031 -->
<!-- requirement-ref: REQ-BUNDLE-0032 -->
<!-- requirement-ref: REQ-BUNDLE-0033 -->
<!-- requirement-ref: REQ-BUNDLE-0034 -->
<!-- requirement-ref: REQ-BUNDLE-0035 -->
<!-- requirement-ref: REQ-BUNDLE-0036 -->

- [REQ-BUNDLE-0018](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0018) → `CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-BUNDLE-0029](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0029) → `CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-BUNDLE-0030](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0030) → `CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-BUNDLE-0031](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0031) → `CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-BUNDLE-0032](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0032) → `CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-BUNDLE-0033](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0033) → `CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-BUNDLE-0034](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0034) → `CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-BUNDLE-0035](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0035) → `CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-BUNDLE-0036](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0036) → `CC-BUNDLE-PRIVACY-AND-SAFETY-REQUIREMENTS`.

### CC-BUNDLE-INTERPRETATION

Public source: `data-formats/result-bundle.md`, section **Interpretation**. Intended target: `internal/bundle/bundle.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcBundleInterpretation`.

Case plan: Review analysis examples for installation/scope/epoch-only correlation; resets and missing tokens cannot be counted as new people.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-BUNDLE-0019 -->
<!-- requirement-ref: REQ-BUNDLE-0020 -->

- [REQ-BUNDLE-0019](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0019) → `CC-BUNDLE-INTERPRETATION`.
- [REQ-BUNDLE-0020](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/result-bundle.md#REQ-BUNDLE-0020) → `CC-BUNDLE-INTERPRETATION`.

### CC-EXECUTION-ENVELOPE

Public source: `data-formats/task-execution.md`, section **Envelope**. Intended target: `internal/execution/attempt.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcExecutionEnvelope`.

Case plan: Recheck constraints immediately before attempts and preserve selected registered protocol semantics through direct/proxy transport.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-EXECUTION-0001 -->
<!-- requirement-ref: REQ-EXECUTION-0002 -->

- [REQ-EXECUTION-0001](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0001) → `CC-EXECUTION-ENVELOPE`.
- [REQ-EXECUTION-0002](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0002) → `CC-EXECUTION-ENVELOPE`.

### CC-EXECUTION-EXECUTION-MODEL

Public source: `data-formats/task-execution.md`, section **Execution Model**. Intended target: `internal/execution/attempt.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcExecutionExecutionModel`.

Case plan: Evaluate each retry independently and assert no extra run is created by constraint evaluation.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-EXECUTION-0003 -->
<!-- requirement-ref: REQ-EXECUTION-0004 -->

- [REQ-EXECUTION-0003](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0003) → `CC-EXECUTION-EXECUTION-MODEL`.
- [REQ-EXECUTION-0004](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0004) → `CC-EXECUTION-EXECUTION-MODEL`.

### CC-EXECUTION-FIELDS

Public source: `data-formats/task-execution.md`, section **Fields**. Intended target: `internal/execution/attempt.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcExecutionFields`.

Case plan: Validate required limits object and optional typed network/device/concurrency/path configuration.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-EXECUTION-0005 -->

- [REQ-EXECUTION-0005](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0005) → `CC-EXECUTION-FIELDS`.

### CC-EXECUTION-FIELDS-NETWORK-PATH

Public source: `data-formats/task-execution.md`, section **Fields / Network Path**. Intended target: `internal/execution/attempt.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcExecutionFieldsNetworkPath`.

Case plan: Test direct default and proxy selection; deny each proxy precondition, snapshot one generation across NET-05, and inject each path failure while asserting zero direct fallback.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-EXECUTION-0006 -->
<!-- requirement-ref: REQ-EXECUTION-0007 -->
<!-- requirement-ref: REQ-EXECUTION-0008 -->
<!-- requirement-ref: REQ-EXECUTION-0009 -->
<!-- requirement-ref: REQ-EXECUTION-0010 -->

- [REQ-EXECUTION-0006](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0006) → `CC-EXECUTION-FIELDS-NETWORK-PATH`.
- [REQ-EXECUTION-0007](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0007) → `CC-EXECUTION-FIELDS-NETWORK-PATH`.
- [REQ-EXECUTION-0008](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0008) → `CC-EXECUTION-FIELDS-NETWORK-PATH`.
- [REQ-EXECUTION-0009](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0009) → `CC-EXECUTION-FIELDS-NETWORK-PATH`.
- [REQ-EXECUTION-0010](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0010) → `CC-EXECUTION-FIELDS-NETWORK-PATH`.

### CC-EXECUTION-FIELDS-LIMITS

Public source: `data-formats/task-execution.md`, section **Fields / Limits**. Intended target: `internal/execution/attempt.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcExecutionFieldsLimits`.

Case plan: Reset budgets per attempt.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-EXECUTION-0011 -->

- [REQ-EXECUTION-0011](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0011) → `CC-EXECUTION-FIELDS-LIMITS`.

### CC-EXECUTION-FIELDS-LIMITS-FIELDS

Public source: `data-formats/task-execution.md`, section **Fields / Limits / Fields**. Intended target: `internal/execution/attempt.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcExecutionFieldsLimitsFields`.

Case plan: Exercise omitted defaults, lower positive requests and above-local/registered rejection; count application bytes and stop reads/writes at bounds while documenting lower-layer buffering.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-EXECUTION-0012 -->
<!-- requirement-ref: REQ-EXECUTION-0013 -->
<!-- requirement-ref: REQ-EXECUTION-0014 -->

- [REQ-EXECUTION-0012](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0012) → `CC-EXECUTION-FIELDS-LIMITS-FIELDS`.
- [REQ-EXECUTION-0013](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0013) → `CC-EXECUTION-FIELDS-LIMITS-FIELDS`.
- [REQ-EXECUTION-0014](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0014) → `CC-EXECUTION-FIELDS-LIMITS-FIELDS`.

### CC-EXECUTION-FIELDS-NETWORK

Public source: `data-formats/task-execution.md`, section **Fields / Network**. Intended target: `internal/execution/attempt.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcExecutionFieldsNetwork`.

Case plan: Simulate unavailable required platform network information; do not assume eligibility.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-EXECUTION-0015 -->

- [REQ-EXECUTION-0015](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0015) → `CC-EXECUTION-FIELDS-NETWORK`.

### CC-EXECUTION-FIELDS-NETWORK-FIELDS

Public source: `data-formats/task-execution.md`, section **Fields / Network / Fields**. Intended target: `internal/execution/attempt.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcExecutionFieldsNetworkFields`.

Case plan: Test wifi/cellular transport filtering and each typed network constraint.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-EXECUTION-0016 -->
<!-- requirement-ref: REQ-EXECUTION-0017 -->

- [REQ-EXECUTION-0016](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0016) → `CC-EXECUTION-FIELDS-NETWORK-FIELDS`.
- [REQ-EXECUTION-0017](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0017) → `CC-EXECUTION-FIELDS-NETWORK-FIELDS`.

### CC-EXECUTION-FIELDS-NETWORK-FIELDS-VPN-POLICY

Public source: `data-formats/task-execution.md`, section **Fields / Network / Fields / VPN Policy**. Intended target: `internal/execution/attempt.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcExecutionFieldsNetworkFieldsVpnPolicy`.

Case plan: Exercise all nine policy/state combinations, preserving unknown independently of not_detected.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-EXECUTION-0018 -->
<!-- requirement-ref: REQ-EXECUTION-0019 -->
<!-- requirement-ref: REQ-EXECUTION-0020 -->

- [REQ-EXECUTION-0018](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0018) → `CC-EXECUTION-FIELDS-NETWORK-FIELDS-VPN-POLICY`.
- [REQ-EXECUTION-0019](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0019) → `CC-EXECUTION-FIELDS-NETWORK-FIELDS-VPN-POLICY`.
- [REQ-EXECUTION-0020](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0020) → `CC-EXECUTION-FIELDS-NETWORK-FIELDS-VPN-POLICY`.

### CC-EXECUTION-FIELDS-NETWORK-FIELDS-NETWORK-CONTEXT-REQUIREMENTS

Public source: `data-formats/task-execution.md`, section **Fields / Network / Fields / Network Context Requirements**. Intended target: `internal/execution/attempt.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcExecutionFieldsNetworkFieldsNetworkContextRequirements`.

Case plan: Exercise required_fields/allowlists, missing/invalid/unknown context, optional context failure and stricter policy. Count requests to prove no endpoint/path override by tasks.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-EXECUTION-0021 -->
<!-- requirement-ref: REQ-EXECUTION-0022 -->
<!-- requirement-ref: REQ-EXECUTION-0023 -->
<!-- requirement-ref: REQ-EXECUTION-0024 -->

- [REQ-EXECUTION-0021](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0021) → `CC-EXECUTION-FIELDS-NETWORK-FIELDS-NETWORK-CONTEXT-REQUIREMENTS`.
- [REQ-EXECUTION-0022](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0022) → `CC-EXECUTION-FIELDS-NETWORK-FIELDS-NETWORK-CONTEXT-REQUIREMENTS`.
- [REQ-EXECUTION-0023](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0023) → `CC-EXECUTION-FIELDS-NETWORK-FIELDS-NETWORK-CONTEXT-REQUIREMENTS`.
- [REQ-EXECUTION-0024](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0024) → `CC-EXECUTION-FIELDS-NETWORK-FIELDS-NETWORK-CONTEXT-REQUIREMENTS`.

### CC-EXECUTION-FIELDS-DEVICE-FIELDS

Public source: `data-formats/task-execution.md`, section **Fields / Device / Fields**. Intended target: `internal/execution/attempt.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcExecutionFieldsDeviceFields`.

Case plan: Exercise battery/charging constraints using synthetic platform state, with local policy stricter than task values.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-EXECUTION-0025 -->
<!-- requirement-ref: REQ-EXECUTION-0026 -->

- [REQ-EXECUTION-0025](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0025) → `CC-EXECUTION-FIELDS-DEVICE-FIELDS`.
- [REQ-EXECUTION-0026](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0026) → `CC-EXECUTION-FIELDS-DEVICE-FIELDS`.

### CC-EXECUTION-FIELDS-CONCURRENCY

Public source: `data-formats/task-execution.md`, section **Fields / Concurrency**. Intended target: `internal/execution/attempt.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcExecutionFieldsConcurrency`.

Case plan: Interleave shared and exclusive attempts and assert an exclusive attempt runs alone.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-EXECUTION-0027 -->

- [REQ-EXECUTION-0027](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0027) → `CC-EXECUTION-FIELDS-CONCURRENCY`.

### CC-EXECUTION-CONSTRAINT-EVALUATION

Public source: `data-formats/task-execution.md`, section **Constraint Evaluation**. Intended target: `internal/execution/attempt.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcExecutionConstraintEvaluation`.

Case plan: Independently fail every applicable runtime constraint and local/OS/user veto; no measurement starts.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-EXECUTION-0028 -->
<!-- requirement-ref: REQ-EXECUTION-0029 -->
<!-- requirement-ref: REQ-EXECUTION-0030 -->

- [REQ-EXECUTION-0028](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0028) → `CC-EXECUTION-CONSTRAINT-EVALUATION`.
- [REQ-EXECUTION-0029](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0029) → `CC-EXECUTION-CONSTRAINT-EVALUATION`.
- [REQ-EXECUTION-0030](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-execution.md#REQ-EXECUTION-0030) → `CC-EXECUTION-CONSTRAINT-EVALUATION`.

### CC-INPUT-ENVELOPE

Public source: `data-formats/task-input.md`, section **Envelope**. Intended target: `internal/validation/input.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcInputEnvelope`.

Case plan: Select exact test ID/version; distinguish schemas for two versions with different fields.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-INPUT-0001 -->

- [REQ-INPUT-0001](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0001) → `CC-INPUT-ENVELOPE`.

### CC-INPUT-INPUT-MODEL

Public source: `data-formats/task-input.md`, section **Input Model**. Intended target: `internal/validation/input.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcInputInputModel`.

Case plan: Remove the locally registered input schema and make a fake schema downloader fail if called; reject before scheduling with zero network calls.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-INPUT-0002 -->

- [REQ-INPUT-0002](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0002) → `CC-INPUT-INPUT-MODEL`.

### CC-INPUT-JSON-STRUCTURE

Public source: `data-formats/task-input.md`, section **JSON Structure**. Intended target: `internal/validation/input.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcInputJsonStructure`.

Case plan: Accept empty object only for a test permitting it; reject every non-object JSON type.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-INPUT-0003 -->
<!-- requirement-ref: REQ-INPUT-0004 -->

- [REQ-INPUT-0003](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0003) → `CC-INPUT-JSON-STRUCTURE`.
- [REQ-INPUT-0004](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0004) → `CC-INPUT-JSON-STRUCTURE`.

### CC-INPUT-INPUT-SCHEMA-REQUIREMENTS

Public source: `data-formats/task-input.md`, section **Input Schema Requirements**. Intended target: `internal/validation/input.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcInputInputSchemaRequirements`.

Case plan: Parameterize required/type/empty/string/array/range/protocol/normalization/additional-field constraints from a registered schema. Exercise encoded-byte boundary and missing required subject without environment defaults.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-INPUT-0005 -->
<!-- requirement-ref: REQ-INPUT-0006 -->
<!-- requirement-ref: REQ-INPUT-0007 -->
<!-- requirement-ref: REQ-INPUT-0008 -->

- [REQ-INPUT-0005](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0005) → `CC-INPUT-INPUT-SCHEMA-REQUIREMENTS`.
- [REQ-INPUT-0006](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0006) → `CC-INPUT-INPUT-SCHEMA-REQUIREMENTS`.
- [REQ-INPUT-0007](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0007) → `CC-INPUT-INPUT-SCHEMA-REQUIREMENTS`.
- [REQ-INPUT-0008](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0008) → `CC-INPUT-INPUT-SCHEMA-REQUIREMENTS`.

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

- [REQ-INPUT-0009](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0009) → `CC-INPUT-VALIDATION`.
- [REQ-INPUT-0010](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0010) → `CC-INPUT-VALIDATION`.
- [REQ-INPUT-0011](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0011) → `CC-INPUT-VALIDATION`.
- [REQ-INPUT-0012](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0012) → `CC-INPUT-VALIDATION`.
- [REQ-INPUT-0027](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0027) → `CC-INPUT-VALIDATION`.
- [REQ-INPUT-0028](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0028) → `CC-INPUT-VALIDATION`.
- [REQ-INPUT-0029](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0029) → `CC-INPUT-VALIDATION`.
- [REQ-INPUT-0030](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0030) → `CC-INPUT-VALIDATION`.
- [REQ-INPUT-0031](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0031) → `CC-INPUT-VALIDATION`.
- [REQ-INPUT-0032](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0032) → `CC-INPUT-VALIDATION`.
- [REQ-INPUT-0033](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0033) → `CC-INPUT-VALIDATION`.
- [REQ-INPUT-0034](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0034) → `CC-INPUT-VALIDATION`.

### CC-INPUT-NORMALIZATION

Public source: `data-formats/task-input.md`, section **Normalization**. Intended target: `internal/validation/input.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcInputNormalization`.

Case plan: Repeat normalization with equivalent supported inputs; reject ambiguous subject changes. Preserve only the normalized fields allowed by the exact Output schema.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-INPUT-0013 -->
<!-- requirement-ref: REQ-INPUT-0014 -->
<!-- requirement-ref: REQ-INPUT-0015 -->

- [REQ-INPUT-0013](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0013) → `CC-INPUT-NORMALIZATION`.
- [REQ-INPUT-0014](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0014) → `CC-INPUT-NORMALIZATION`.
- [REQ-INPUT-0015](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0015) → `CC-INPUT-NORMALIZATION`.

### CC-INPUT-TARGET-SAFETY

Public source: `data-formats/task-input.md`, section **Target Safety**. Intended target: `internal/validation/input.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcInputTargetSafety`.

Case plan: Use syntactically valid but unregistered/prohibited local, private, link-local and credential-bearing targets. Recheck redirects before contact; count zero contacts to denied destinations.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-INPUT-0016 -->
<!-- requirement-ref: REQ-INPUT-0017 -->
<!-- requirement-ref: REQ-INPUT-0018 -->

- [REQ-INPUT-0016](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0016) → `CC-INPUT-TARGET-SAFETY`.
- [REQ-INPUT-0017](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0017) → `CC-INPUT-TARGET-SAFETY`.
- [REQ-INPUT-0018](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0018) → `CC-INPUT-TARGET-SAFETY`.

### CC-INPUT-SENSITIVE-AND-DYNAMIC-INPUT

Public source: `data-formats/task-input.md`, section **Sensitive and Dynamic Input**. Intended target: `internal/validation/input.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcInputSensitiveAndDynamicInput`.

Case plan: Attempt every prohibited sensitive input class using synthetic canaries; verify no runtime substitution or automatic access to host/device data.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-INPUT-0019 -->
<!-- requirement-ref: REQ-INPUT-0020 -->

- [REQ-INPUT-0019](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0019) → `CC-INPUT-SENSITIVE-AND-DYNAMIC-INPUT`.
- [REQ-INPUT-0020](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0020) → `CC-INPUT-SENSITIVE-AND-DYNAMIC-INPUT`.

### CC-INPUT-MULTIPLE-TARGETS

Public source: `data-formats/task-input.md`, section **Multiple Targets**. Intended target: `internal/validation/input.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcInputMultipleTargets`.

Case plan: Reject an unregistered generic hostnames array; exercise the exact registered array item/count/encoded-size boundaries for any supported multi-target schema.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-INPUT-0021 -->
<!-- requirement-ref: REQ-INPUT-0022 -->
<!-- requirement-ref: REQ-INPUT-0023 -->

- [REQ-INPUT-0021](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0021) → `CC-INPUT-MULTIPLE-TARGETS`.
- [REQ-INPUT-0022](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0022) → `CC-INPUT-MULTIPLE-TARGETS`.
- [REQ-INPUT-0023](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0023) → `CC-INPUT-MULTIPLE-TARGETS`.

### CC-INPUT-RESULT-PROPAGATION

Public source: `data-formats/task-input.md`, section **Result Propagation**. Intended target: `internal/validation/input.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcInputResultPropagation`.

Case plan: Construct output from a task containing canary inputs; retain only fields authorized by the selected Output schema and local policy.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-INPUT-0024 -->

- [REQ-INPUT-0024](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0024) → `CC-INPUT-RESULT-PROPAGATION`.

### CC-INPUT-IMPLEMENTATION-AUTHORITY

Public source: `data-formats/task-input.md`, section **Implementation Authority**. Intended target: `internal/validation/input.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcInputImplementationAuthority`.

Case plan: Try to widen input schemas or weaken local restrictions using task input; verify the registered schema and stricter local policy prevail.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-INPUT-0025 -->
<!-- requirement-ref: REQ-INPUT-0026 -->

- [REQ-INPUT-0025](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0025) → `CC-INPUT-IMPLEMENTATION-AUTHORITY`.
- [REQ-INPUT-0026](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-input.md#REQ-INPUT-0026) → `CC-INPUT-IMPLEMENTATION-AUTHORITY`.

### CC-LABEL-ENVELOPE

Public source: `data-formats/task-labels.md`, section **Envelope**. Intended target: `internal/taskdecode/labels.go`. Planned case location: `internal/taskdecode/cross_cutting_conformance_test.go`, `symbol:TestCcLabelEnvelope`.

Case plan: Compare executions with different unreserved labels; labels cannot alter typed control fields.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LABEL-0001 -->
<!-- requirement-ref: REQ-LABEL-0002 -->

- [REQ-LABEL-0001](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0001) → `CC-LABEL-ENVELOPE`.
- [REQ-LABEL-0002](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0002) → `CC-LABEL-ENVELOPE`.

### CC-LABEL-JSON-STRUCTURE

Public source: `data-formats/task-labels.md`, section **JSON Structure**. Intended target: `internal/taskdecode/labels.go`. Planned case location: `internal/taskdecode/cross_cutting_conformance_test.go`, `symbol:TestCcLabelJsonStructure`.

Case plan: Compare absent and empty labels; reject non-flat or non-string key/value representations.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LABEL-0003 -->
<!-- requirement-ref: REQ-LABEL-0004 -->

- [REQ-LABEL-0003](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0003) → `CC-LABEL-JSON-STRUCTURE`.
- [REQ-LABEL-0004](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0004) → `CC-LABEL-JSON-STRUCTURE`.

### CC-LABEL-CONSTRAINTS

Public source: `data-formats/task-labels.md`, section **Constraints**. Intended target: `internal/taskdecode/labels.go`. Planned case location: `internal/taskdecode/cross_cutting_conformance_test.go`, `symbol:TestCcLabelConstraints`.

Case plan: Exercise 16 entries, 64-byte keys, 128-byte values and 2048 encoded bytes at/below/above bounds, counting JSON syntax and multibyte text.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LABEL-0005 -->
<!-- requirement-ref: REQ-LABEL-0006 -->
<!-- requirement-ref: REQ-LABEL-0007 -->

- [REQ-LABEL-0005](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0005) → `CC-LABEL-CONSTRAINTS`.
- [REQ-LABEL-0006](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0006) → `CC-LABEL-CONSTRAINTS`.
- [REQ-LABEL-0007](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0007) → `CC-LABEL-CONSTRAINTS`.

### CC-LABEL-KEYS

Public source: `data-formats/task-labels.md`, section **Keys**. Intended target: `internal/taskdecode/labels.go`. Planned case location: `internal/taskdecode/cross_cutting_conformance_test.go`, `symbol:TestCcLabelKeys`.

Case plan: Test the exact key alphabet, initial character, case sensitivity, length, and equal/conflicting/escaped duplicate names.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LABEL-0008 -->
<!-- requirement-ref: REQ-LABEL-0009 -->
<!-- requirement-ref: REQ-LABEL-0010 -->

- [REQ-LABEL-0008](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0008) → `CC-LABEL-KEYS`.
- [REQ-LABEL-0009](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0009) → `CC-LABEL-KEYS`.
- [REQ-LABEL-0010](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0010) → `CC-LABEL-KEYS`.

### CC-LABEL-KEYS-RESERVED-NAMESPACE

Public source: `data-formats/task-labels.md`, section **Keys / Reserved Namespace**. Intended target: `internal/taskdecode/labels.go`. Planned case location: `internal/taskdecode/cross_cutting_conformance_test.go`, `symbol:TestCcLabelKeysReservedNamespace`.

Case plan: Reject an unknown barometer. key; accept a structurally valid unreserved key without assigning it control semantics.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LABEL-0011 -->
<!-- requirement-ref: REQ-LABEL-0012 -->

- [REQ-LABEL-0011](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0011) → `CC-LABEL-KEYS-RESERVED-NAMESPACE`.
- [REQ-LABEL-0012](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0012) → `CC-LABEL-KEYS-RESERVED-NAMESPACE`.

### CC-LABEL-VALUES

Public source: `data-formats/task-labels.md`, section **Values**. Intended target: `internal/taskdecode/labels.go`. Planned case location: `internal/taskdecode/cross_cutting_conformance_test.go`, `symbol:TestCcLabelValues`.

Case plan: Test UTF-8 and case preservation, control-character rejection, wrong types, literal expression text and absence of automatic numeric/boolean coercion.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LABEL-0013 -->
<!-- requirement-ref: REQ-LABEL-0014 -->
<!-- requirement-ref: REQ-LABEL-0015 -->
<!-- requirement-ref: REQ-LABEL-0016 -->
<!-- requirement-ref: REQ-LABEL-0017 -->

- [REQ-LABEL-0013](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0013) → `CC-LABEL-VALUES`.
- [REQ-LABEL-0014](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0014) → `CC-LABEL-VALUES`.
- [REQ-LABEL-0015](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0015) → `CC-LABEL-VALUES`.
- [REQ-LABEL-0016](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0016) → `CC-LABEL-VALUES`.
- [REQ-LABEL-0017](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0017) → `CC-LABEL-VALUES`.

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

- [REQ-LABEL-0018](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0018) → `CC-LABEL-PRIVACY-AND-SAFETY`.
- [REQ-LABEL-0019](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0019) → `CC-LABEL-PRIVACY-AND-SAFETY`.
- [REQ-LABEL-0020](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0020) → `CC-LABEL-PRIVACY-AND-SAFETY`.
- [REQ-LABEL-0021](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0021) → `CC-LABEL-PRIVACY-AND-SAFETY`.
- [REQ-LABEL-0022](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0022) → `CC-LABEL-PRIVACY-AND-SAFETY`.
- [REQ-LABEL-0023](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0023) → `CC-LABEL-PRIVACY-AND-SAFETY`.

### CC-LABEL-PROCESSING

Public source: `data-formats/task-labels.md`, section **Processing**. Intended target: `internal/taskdecode/labels.go`. Planned case location: `internal/taskdecode/cross_cutting_conformance_test.go`, `symbol:TestCcLabelProcessing`.

Case plan: Vary every apparent behavior/destination label and verify execution is unchanged; accept unknown unreserved metadata and exclude labels from current results.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LABEL-0024 -->
<!-- requirement-ref: REQ-LABEL-0025 -->
<!-- requirement-ref: REQ-LABEL-0026 -->
<!-- requirement-ref: REQ-LABEL-0027 -->

- [REQ-LABEL-0024](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0024) → `CC-LABEL-PROCESSING`.
- [REQ-LABEL-0025](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0025) → `CC-LABEL-PROCESSING`.
- [REQ-LABEL-0026](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0026) → `CC-LABEL-PROCESSING`.
- [REQ-LABEL-0027](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0027) → `CC-LABEL-PROCESSING`.

### CC-LABEL-VALIDATION

Public source: `data-formats/task-labels.md`, section **Validation**. Intended target: `internal/taskdecode/labels.go`. Planned case location: `internal/taskdecode/cross_cutting_conformance_test.go`, `symbol:TestCcLabelValidation`.

Case plan: Mutate each validity predicate independently; return task-validation errors without emitting measurements.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LABEL-0028 -->
<!-- requirement-ref: REQ-LABEL-0029 -->

- [REQ-LABEL-0028](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0028) → `CC-LABEL-VALIDATION`.
- [REQ-LABEL-0029](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-labels.md#REQ-LABEL-0029) → `CC-LABEL-VALIDATION`.

### CC-OPTIONS-ENVELOPE

Public source: `data-formats/task-options.md`, section **Envelope**. Intended target: `internal/validation/options.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcOptionsEnvelope`.

Case plan: Use identical task envelopes with different exact test versions; select each version-specific options schema without changing the outer execution envelope.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OPTIONS-0001 -->
<!-- requirement-ref: REQ-OPTIONS-0002 -->

- [REQ-OPTIONS-0001](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0001) → `CC-OPTIONS-ENVELOPE`.
- [REQ-OPTIONS-0002](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0002) → `CC-OPTIONS-ENVELOPE`.

### CC-OPTIONS-JSON-STRUCTURE

Public source: `data-formats/task-options.md`, section **JSON Structure**. Intended target: `internal/validation/options.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcOptionsJsonStructure`.

Case plan: Test required object presence, empty-object defaults and every non-object type.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OPTIONS-0003 -->
<!-- requirement-ref: REQ-OPTIONS-0004 -->
<!-- requirement-ref: REQ-OPTIONS-0005 -->

- [REQ-OPTIONS-0003](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0003) → `CC-OPTIONS-JSON-STRUCTURE`.
- [REQ-OPTIONS-0004](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0004) → `CC-OPTIONS-JSON-STRUCTURE`.
- [REQ-OPTIONS-0005](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0005) → `CC-OPTIONS-JSON-STRUCTURE`.

### CC-OPTIONS-OPTIONS-SCHEMA

Public source: `data-formats/task-options.md`, section **Options Schema**. Intended target: `internal/validation/options.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcOptionsOptionsSchema`.

Case plan: Remove the local schema; parameterize required fields, types, defaults, enum/range/count/encoded-byte bounds and input relationships; reject unknown properties without network schema acquisition.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OPTIONS-0006 -->
<!-- requirement-ref: REQ-OPTIONS-0007 -->
<!-- requirement-ref: REQ-OPTIONS-0008 -->
<!-- requirement-ref: REQ-OPTIONS-0009 -->
<!-- requirement-ref: REQ-OPTIONS-0010 -->

- [REQ-OPTIONS-0006](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0006) → `CC-OPTIONS-OPTIONS-SCHEMA`.
- [REQ-OPTIONS-0007](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0007) → `CC-OPTIONS-OPTIONS-SCHEMA`.
- [REQ-OPTIONS-0008](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0008) → `CC-OPTIONS-OPTIONS-SCHEMA`.
- [REQ-OPTIONS-0009](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0009) → `CC-OPTIONS-OPTIONS-SCHEMA`.
- [REQ-OPTIONS-0010](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0010) → `CC-OPTIONS-OPTIONS-SCHEMA`.

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

- [REQ-OPTIONS-0011](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0011) → `CC-OPTIONS-VALIDATION`.
- [REQ-OPTIONS-0012](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0012) → `CC-OPTIONS-VALIDATION`.
- [REQ-OPTIONS-0013](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0013) → `CC-OPTIONS-VALIDATION`.
- [REQ-OPTIONS-0014](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0014) → `CC-OPTIONS-VALIDATION`.
- [REQ-OPTIONS-0015](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0015) → `CC-OPTIONS-VALIDATION`.
- [REQ-OPTIONS-0025](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0025) → `CC-OPTIONS-VALIDATION`.
- [REQ-OPTIONS-0026](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0026) → `CC-OPTIONS-VALIDATION`.
- [REQ-OPTIONS-0027](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0027) → `CC-OPTIONS-VALIDATION`.
- [REQ-OPTIONS-0028](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0028) → `CC-OPTIONS-VALIDATION`.
- [REQ-OPTIONS-0029](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0029) → `CC-OPTIONS-VALIDATION`.
- [REQ-OPTIONS-0030](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0030) → `CC-OPTIONS-VALIDATION`.
- [REQ-OPTIONS-0031](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0031) → `CC-OPTIONS-VALIDATION`.
- [REQ-OPTIONS-0032](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0032) → `CC-OPTIONS-VALIDATION`.

### CC-OPTIONS-EXECUTION-PRECEDENCE

Public source: `data-formats/task-options.md`, section **Execution Precedence**. Intended target: `internal/validation/options.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcOptionsExecutionPrecedence`.

Case plan: Set a protocol timeout beyond the outer attempt deadline; exercise validation and outer cancellation without converting the outer interruption to a protocol outcome.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OPTIONS-0016 -->
<!-- requirement-ref: REQ-OPTIONS-0017 -->

- [REQ-OPTIONS-0016](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0016) → `CC-OPTIONS-EXECUTION-PRECEDENCE`.
- [REQ-OPTIONS-0017](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0017) → `CC-OPTIONS-EXECUTION-PRECEDENCE`.

### CC-OPTIONS-NO-SILENT-REWRITING

Public source: `data-formats/task-options.md`, section **No Silent Rewriting**. Intended target: `internal/validation/options.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcOptionsNoSilentRewriting`.

Case plan: Supply unsupported/unsafe explicit options and omitted defaults separately; observe rejection or permitted skipping without clamping, coercion or substitution.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OPTIONS-0018 -->
<!-- requirement-ref: REQ-OPTIONS-0019 -->
<!-- requirement-ref: REQ-OPTIONS-0020 -->

- [REQ-OPTIONS-0018](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0018) → `CC-OPTIONS-NO-SILENT-REWRITING`.
- [REQ-OPTIONS-0019](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0019) → `CC-OPTIONS-NO-SILENT-REWRITING`.
- [REQ-OPTIONS-0020](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0020) → `CC-OPTIONS-NO-SILENT-REWRITING`.

### CC-OPTIONS-SECURITY-AND-SAFETY

Public source: `data-formats/task-options.md`, section **Security and Safety**. Intended target: `internal/validation/options.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcOptionsSecurityAndSafety`.

Case plan: Attempt each listed option-based code/file/secret/interpolation/control-channel bypass; preserve registered behavior and stricter local safety policy.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-OPTIONS-0021 -->
<!-- requirement-ref: REQ-OPTIONS-0022 -->
<!-- requirement-ref: REQ-OPTIONS-0023 -->
<!-- requirement-ref: REQ-OPTIONS-0024 -->

- [REQ-OPTIONS-0021](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0021) → `CC-OPTIONS-SECURITY-AND-SAFETY`.
- [REQ-OPTIONS-0022](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0022) → `CC-OPTIONS-SECURITY-AND-SAFETY`.
- [REQ-OPTIONS-0023](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0023) → `CC-OPTIONS-SECURITY-AND-SAFETY`.
- [REQ-OPTIONS-0024](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-options.md#REQ-OPTIONS-0024) → `CC-OPTIONS-SECURITY-AND-SAFETY`.

### CC-SAFETY-ENVELOPE

Public source: `data-formats/task-safety.md`, section **Envelope**. Intended target: `internal/validation/safety.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcSafetyEnvelope`.

Case plan: Change trusted local consent/risk permission between attempts; re-evaluate declarations before each attempt without creating runs.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SAFETY-0001 -->
<!-- requirement-ref: REQ-SAFETY-0002 -->
<!-- requirement-ref: REQ-SAFETY-0003 -->

- [REQ-SAFETY-0001](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0001) → `CC-SAFETY-ENVELOPE`.
- [REQ-SAFETY-0002](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0002) → `CC-SAFETY-ENVELOPE`.
- [REQ-SAFETY-0003](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0003) → `CC-SAFETY-ENVELOPE`.

### CC-SAFETY-SAFETY-MODEL

Public source: `data-formats/task-safety.md`, section **Safety Model**. Intended target: `internal/validation/safety.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcSafetySafetyModel`.

Case plan: Combine registered, configured and local risk; a declaration cannot reduce baseline risk or create runs/attempts.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SAFETY-0004 -->
<!-- requirement-ref: REQ-SAFETY-0005 -->
<!-- requirement-ref: REQ-SAFETY-0006 -->

- [REQ-SAFETY-0004](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0004) → `CC-SAFETY-SAFETY-MODEL`.
- [REQ-SAFETY-0005](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0005) → `CC-SAFETY-SAFETY-MODEL`.
- [REQ-SAFETY-0006](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0006) → `CC-SAFETY-SAFETY-MODEL`.

### CC-SAFETY-FIELDS

Public source: `data-formats/task-safety.md`, section **Fields**. Intended target: `internal/validation/safety.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcSafetyFields`.

Case plan: Exercise required risk_level/consent fields and conditional risk_factors with wrong types and unknown fields.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SAFETY-0007 -->

- [REQ-SAFETY-0007](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0007) → `CC-SAFETY-FIELDS`.

### CC-SAFETY-RISK-LEVEL

Public source: `data-formats/task-safety.md`, section **Risk Level**. Intended target: `internal/validation/safety.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcSafetyRiskLevel`.

Case plan: Test all registered levels, reject none, and exercise unknown with absent and explicit reviewed permission. Review UI/docs for risk-free guarantees.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SAFETY-0008 -->
<!-- requirement-ref: REQ-SAFETY-0009 -->
<!-- requirement-ref: REQ-SAFETY-0010 -->
<!-- requirement-ref: REQ-SAFETY-0011 -->

- [REQ-SAFETY-0008](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0008) → `CC-SAFETY-RISK-LEVEL`.
- [REQ-SAFETY-0009](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0009) → `CC-SAFETY-RISK-LEVEL`.
- [REQ-SAFETY-0010](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0010) → `CC-SAFETY-RISK-LEVEL`.
- [REQ-SAFETY-0011](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0011) → `CC-SAFETY-RISK-LEVEL`.

### CC-SAFETY-RISK-FACTORS

Public source: `data-formats/task-safety.md`, section **Risk Factors**. Intended target: `internal/validation/safety.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcSafetyRiskFactors`.

Case plan: Require factors for medium/high/unknown and allow omission for low.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SAFETY-0012 -->

- [REQ-SAFETY-0012](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0012) → `CC-SAFETY-RISK-FACTORS`.

### CC-SAFETY-RISK-FACTORS-RISK-FACTOR-FIELDS

Public source: `data-formats/task-safety.md`, section **Risk Factors / Risk Factor Fields**. Intended target: `internal/validation/safety.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcSafetyRiskFactorsRiskFactorFields`.

Case plan: Test required type/level fields and exact risk enums for each factor.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SAFETY-0013 -->
<!-- requirement-ref: REQ-SAFETY-0014 -->

- [REQ-SAFETY-0013](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0013) → `CC-SAFETY-RISK-FACTORS-RISK-FACTOR-FIELDS`.
- [REQ-SAFETY-0014](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0014) → `CC-SAFETY-RISK-FACTORS-RISK-FACTOR-FIELDS`.

### CC-SAFETY-RISK-FACTORS-INITIAL-RISK-FACTOR-TYPES

Public source: `data-formats/task-safety.md`, section **Risk Factors / Initial Risk Factor Types**. Intended target: `internal/validation/safety.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcSafetyRiskFactorsInitialRiskFactorTypes`.

Case plan: Test every known type and an unknown type; require rejection or conservative unknown handling, and keep resource/device requirements in execution.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SAFETY-0015 -->
<!-- requirement-ref: REQ-SAFETY-0016 -->
<!-- requirement-ref: REQ-SAFETY-0017 -->

- [REQ-SAFETY-0015](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0015) → `CC-SAFETY-RISK-FACTORS-INITIAL-RISK-FACTOR-TYPES`.
- [REQ-SAFETY-0016](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0016) → `CC-SAFETY-RISK-FACTORS-INITIAL-RISK-FACTOR-TYPES`.
- [REQ-SAFETY-0017](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0017) → `CC-SAFETY-RISK-FACTORS-INITIAL-RISK-FACTOR-TYPES`.

### CC-SAFETY-RISK-FACTORS-RISK-CONSISTENCY

Public source: `data-formats/task-safety.md`, section **Risk Factors / Risk Consistency**. Intended target: `internal/validation/safety.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcSafetyRiskFactorsRiskConsistency`.

Case plan: Exercise the risk ordering matrix and an unknown factor with/without trusted resolution; reject or conservatively elevate understated risk as permitted.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SAFETY-0018 -->
<!-- requirement-ref: REQ-SAFETY-0019 -->

- [REQ-SAFETY-0018](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0018) → `CC-SAFETY-RISK-FACTORS-RISK-CONSISTENCY`.
- [REQ-SAFETY-0019](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0019) → `CC-SAFETY-RISK-FACTORS-RISK-CONSISTENCY`.

### CC-SAFETY-USER-CONSENT

Public source: `data-formats/task-safety.md`, section **User Consent**. Intended target: `internal/validation/safety.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcSafetyUserConsent`.

Case plan: Test task consent true/false against trusted host consent and stricter local requirements; a false task flag grants no permission.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SAFETY-0020 -->
<!-- requirement-ref: REQ-SAFETY-0021 -->
<!-- requirement-ref: REQ-SAFETY-0022 -->

- [REQ-SAFETY-0020](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0020) → `CC-SAFETY-USER-CONSENT`.
- [REQ-SAFETY-0021](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0021) → `CC-SAFETY-USER-CONSENT`.
- [REQ-SAFETY-0022](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0022) → `CC-SAFETY-USER-CONSENT`.

### CC-SAFETY-EFFECTIVE-RISK

Public source: `data-formats/task-safety.md`, section **Effective Risk**. Intended target: `internal/validation/safety.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcSafetyEffectiveRisk`.

Case plan: Vary registered/task/local risk independently, including a task below baseline; assert the most restrictive effective risk or permitted rejection.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SAFETY-0023 -->
<!-- requirement-ref: REQ-SAFETY-0024 -->
<!-- requirement-ref: REQ-SAFETY-0025 -->
<!-- requirement-ref: REQ-SAFETY-0026 -->

- [REQ-SAFETY-0023](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0023) → `CC-SAFETY-EFFECTIVE-RISK`.
- [REQ-SAFETY-0024](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0024) → `CC-SAFETY-EFFECTIVE-RISK`.
- [REQ-SAFETY-0025](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0025) → `CC-SAFETY-EFFECTIVE-RISK`.
- [REQ-SAFETY-0026](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0026) → `CC-SAFETY-EFFECTIVE-RISK`.

### CC-SAFETY-SAFETY-EVALUATION

Public source: `data-formats/task-safety.md`, section **Safety Evaluation**. Intended target: `internal/validation/safety.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcSafetySafetyEvaluation`.

Case plan: Independently deny each required safety/consent/device predicate or make assessment unavailable; prevent execution and distinguish rejection from network outcomes.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SAFETY-0027 -->
<!-- requirement-ref: REQ-SAFETY-0028 -->
<!-- requirement-ref: REQ-SAFETY-0029 -->
<!-- requirement-ref: REQ-SAFETY-0030 -->

- [REQ-SAFETY-0027](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0027) → `CC-SAFETY-SAFETY-EVALUATION`.
- [REQ-SAFETY-0028](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0028) → `CC-SAFETY-SAFETY-EVALUATION`.
- [REQ-SAFETY-0029](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0029) → `CC-SAFETY-SAFETY-EVALUATION`.
- [REQ-SAFETY-0030](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-safety.md#REQ-SAFETY-0030) → `CC-SAFETY-SAFETY-EVALUATION`.

### CC-SCHEDULE-ENVELOPE

Public source: `data-formats/task-schedule.md`, section **Envelope**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleEnvelope`.

Case plan: Delay eligible work under simulated suspension; never promise exact firing times.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0001 -->

- [REQ-SCHEDULE-0001](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0001) → `CC-SCHEDULE-ENVELOPE`.

### CC-SCHEDULE-FIELDS

Public source: `data-formats/task-schedule.md`, section **Fields**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleFields`.

Case plan: Exercise omitted/present and conditional schedule fields, defaults, wrong types and unknown properties.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0002 -->

- [REQ-SCHEDULE-0002](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0002) → `CC-SCHEDULE-FIELDS`.

### CC-SCHEDULE-SCHEDULING-MODEL

Public source: `data-formats/task-schedule.md`, section **Scheduling Model**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleSchedulingModel`.

Case plan: Distinguish repeated runs from retry attempts using deterministic identities and counters.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0003 -->
<!-- requirement-ref: REQ-SCHEDULE-0004 -->
<!-- requirement-ref: REQ-SCHEDULE-0005 -->
<!-- requirement-ref: REQ-SCHEDULE-0006 -->

- [REQ-SCHEDULE-0003](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0003) → `CC-SCHEDULE-SCHEDULING-MODEL`.
- [REQ-SCHEDULE-0004](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0004) → `CC-SCHEDULE-SCHEDULING-MODEL`.
- [REQ-SCHEDULE-0005](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0005) → `CC-SCHEDULE-SCHEDULING-MODEL`.
- [REQ-SCHEDULE-0006](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0006) → `CC-SCHEDULE-SCHEDULING-MODEL`.

### CC-SCHEDULE-SCHEDULING-MODEL-START-TIME

Public source: `data-formats/task-schedule.md`, section **Scheduling Model / Start Time**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleSchedulingModelStartTime`.

Case plan: Test before/at/after starts_at plus permitted null/absence; do not run early or require exact timing.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0007 -->
<!-- requirement-ref: REQ-SCHEDULE-0008 -->
<!-- requirement-ref: REQ-SCHEDULE-0009 -->

- [REQ-SCHEDULE-0007](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0007) → `CC-SCHEDULE-SCHEDULING-MODEL-START-TIME`.
- [REQ-SCHEDULE-0008](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0008) → `CC-SCHEDULE-SCHEDULING-MODEL-START-TIME`.
- [REQ-SCHEDULE-0009](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0009) → `CC-SCHEDULE-SCHEDULING-MODEL-START-TIME`.

### CC-SCHEDULE-SCHEDULING-MODEL-TIME-LIMIT

Public source: `data-formats/task-schedule.md`, section **Scheduling Model / Time Limit**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleSchedulingModelTimeLimit`.

Case plan: Test acceptance/issuance-relative expiry at boundaries and the reviewed embedded one-time exception.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0010 -->
<!-- requirement-ref: REQ-SCHEDULE-0011 -->

- [REQ-SCHEDULE-0010](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0010) → `CC-SCHEDULE-SCHEDULING-MODEL-TIME-LIMIT`.
- [REQ-SCHEDULE-0011](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0011) → `CC-SCHEDULE-SCHEDULING-MODEL-TIME-LIMIT`.

### CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW

Public source: `data-formats/task-schedule.md`, section **Scheduling Model / Time Window**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleSchedulingModelTimeWindow`.

Case plan: Start attempts only in allowed local windows.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0012 -->

- [REQ-SCHEDULE-0012](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0012) → `CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW`.

### CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-FIELDS

Public source: `data-formats/task-schedule.md`, section **Scheduling Model / Time Window / Fields**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleSchedulingModelTimeWindowFields`.

Case plan: Test local timezone changes, range count/uniqueness/overlap and default timezone.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0013 -->
<!-- requirement-ref: REQ-SCHEDULE-0014 -->

- [REQ-SCHEDULE-0013](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0013) → `CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-FIELDS`.
- [REQ-SCHEDULE-0014](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0014) → `CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-FIELDS`.

### CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-RANGE-FIELDS

Public source: `data-formats/task-schedule.md`, section **Scheduling Model / Time Window / Range Fields**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleSchedulingModelTimeWindowRangeFields`.

Case plan: Test HH:mm syntax, inclusive start/exclusive end, midnight expansion and overlapping ranges after expansion.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0015 -->
<!-- requirement-ref: REQ-SCHEDULE-0016 -->
<!-- requirement-ref: REQ-SCHEDULE-0017 -->

- [REQ-SCHEDULE-0015](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0015) → `CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-RANGE-FIELDS`.
- [REQ-SCHEDULE-0016](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0016) → `CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-RANGE-FIELDS`.
- [REQ-SCHEDULE-0017](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0017) → `CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-RANGE-FIELDS`.

### CC-SCHEDULE-SCHEDULING-MODEL-REPEAT

Public source: `data-formats/task-schedule.md`, section **Scheduling Model / Repeat**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleSchedulingModelRepeat`.

Case plan: Distinguish none and interval, disallow interval-only fields for none and generate the bounded number of runs.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0018 -->

- [REQ-SCHEDULE-0018](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0018) → `CC-SCHEDULE-SCHEDULING-MODEL-REPEAT`.

### CC-SCHEDULE-SCHEDULING-MODEL-REPEAT-INTERVAL-FIELDS

Public source: `data-formats/task-schedule.md`, section **Scheduling Model / Repeat / Interval Fields**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleSchedulingModelRepeatIntervalFields`.

Case plan: Exercise interval/jitter bounds and delayed due runs with injected random offsets; offsets are locally generated per run and stay inside start/expiry/window constraints.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0019 -->
<!-- requirement-ref: REQ-SCHEDULE-0020 -->
<!-- requirement-ref: REQ-SCHEDULE-0021 -->

- [REQ-SCHEDULE-0019](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0019) → `CC-SCHEDULE-SCHEDULING-MODEL-REPEAT-INTERVAL-FIELDS`.
- [REQ-SCHEDULE-0020](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0020) → `CC-SCHEDULE-SCHEDULING-MODEL-REPEAT-INTERVAL-FIELDS`.
- [REQ-SCHEDULE-0021](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0021) → `CC-SCHEDULE-SCHEDULING-MODEL-REPEAT-INTERVAL-FIELDS`.

### CC-SCHEDULE-SCHEDULING-MODEL-RETRY

Public source: `data-formats/task-schedule.md`, section **Scheduling Model / Retry**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleSchedulingModelRetry`.

Case plan: Keep run identity/max_runs accounting stable across the second attempt; a terminal negative measurement creates no automatic retry.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0022 -->
<!-- requirement-ref: REQ-SCHEDULE-0023 -->
<!-- requirement-ref: REQ-SCHEDULE-0024 -->
<!-- requirement-ref: REQ-SCHEDULE-0025 -->

- [REQ-SCHEDULE-0022](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0022) → `CC-SCHEDULE-SCHEDULING-MODEL-RETRY`.
- [REQ-SCHEDULE-0023](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0023) → `CC-SCHEDULE-SCHEDULING-MODEL-RETRY`.
- [REQ-SCHEDULE-0024](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0024) → `CC-SCHEDULE-SCHEDULING-MODEL-RETRY`.
- [REQ-SCHEDULE-0025](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0025) → `CC-SCHEDULE-SCHEDULING-MODEL-RETRY`.

### CC-SCHEDULE-SCHEDULING-MODEL-RETRY-RETRY-FIELDS

Public source: `data-formats/task-schedule.md`, section **Scheduling Model / Retry / Retry Fields**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleSchedulingModelRetryRetryFields`.

Case plan: Enumerate permitted and forbidden reasons; exercise attempts=1/2 conditional fields, fixed/exponential delays and jitter at expiry and delay bounds.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0026 -->
<!-- requirement-ref: REQ-SCHEDULE-0027 -->
<!-- requirement-ref: REQ-SCHEDULE-0028 -->
<!-- requirement-ref: REQ-SCHEDULE-0029 -->

- [REQ-SCHEDULE-0026](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0026) → `CC-SCHEDULE-SCHEDULING-MODEL-RETRY-RETRY-FIELDS`.
- [REQ-SCHEDULE-0027](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0027) → `CC-SCHEDULE-SCHEDULING-MODEL-RETRY-RETRY-FIELDS`.
- [REQ-SCHEDULE-0028](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0028) → `CC-SCHEDULE-SCHEDULING-MODEL-RETRY-RETRY-FIELDS`.
- [REQ-SCHEDULE-0029](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0029) → `CC-SCHEDULE-SCHEDULING-MODEL-RETRY-RETRY-FIELDS`.

### CC-SCHEDULE-SCHEDULING-MODEL-LIMITS-FIELDS

Public source: `data-formats/task-schedule.md`, section **Scheduling Model / Limits / Fields**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleSchedulingModelLimitsFields`.

Case plan: Test max_runs default and bounds, repeat requirement and retries excluded from the run count.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0030 -->
<!-- requirement-ref: REQ-SCHEDULE-0031 -->
<!-- requirement-ref: REQ-SCHEDULE-0032 -->

- [REQ-SCHEDULE-0030](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0030) → `CC-SCHEDULE-SCHEDULING-MODEL-LIMITS-FIELDS`.
- [REQ-SCHEDULE-0031](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0031) → `CC-SCHEDULE-SCHEDULING-MODEL-LIMITS-FIELDS`.
- [REQ-SCHEDULE-0032](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0032) → `CC-SCHEDULE-SCHEDULING-MODEL-LIMITS-FIELDS`.

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

- [REQ-SCHEDULE-0033](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0033) → `CC-SCHEDULE-VALIDATION`.
- [REQ-SCHEDULE-0036](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0036) → `CC-SCHEDULE-VALIDATION`.
- [REQ-SCHEDULE-0037](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0037) → `CC-SCHEDULE-VALIDATION`.
- [REQ-SCHEDULE-0038](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0038) → `CC-SCHEDULE-VALIDATION`.
- [REQ-SCHEDULE-0039](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0039) → `CC-SCHEDULE-VALIDATION`.
- [REQ-SCHEDULE-0040](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0040) → `CC-SCHEDULE-VALIDATION`.
- [REQ-SCHEDULE-0041](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0041) → `CC-SCHEDULE-VALIDATION`.
- [REQ-SCHEDULE-0042](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0042) → `CC-SCHEDULE-VALIDATION`.

### CC-SCHEDULE-VALIDATION-ELIGIBILITY-RULES

Public source: `data-formats/task-schedule.md`, section **Validation / Eligibility Rules**. Intended target: `internal/execution/schedule.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcScheduleValidationEligibilityRules`.

Case plan: Deny each eligibility predicate independently for initial and retry attempts; assert no early/out-of-window/expired execution and preserve stricter local veto.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-SCHEDULE-0034 -->
<!-- requirement-ref: REQ-SCHEDULE-0035 -->

- [REQ-SCHEDULE-0034](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0034) → `CC-SCHEDULE-VALIDATION-ELIGIBILITY-RULES`.
- [REQ-SCHEDULE-0035](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task-schedule.md#REQ-SCHEDULE-0035) → `CC-SCHEDULE-VALIDATION-ELIGIBILITY-RULES`.

### CC-TASK-OVERVIEW

Public source: `data-formats/task.md`, section **Overview**. Intended target: `internal/taskdecode/decode.go`. Planned case location: `internal/taskdecode/cross_cutting_conformance_test.go`, `symbol:TestCcTaskOverview`.

Case plan: Verify task admission produces no Measurement Result from supplied task data.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-TASK-0002 -->

- [REQ-TASK-0002](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task.md#REQ-TASK-0002) → `CC-TASK-OVERVIEW`.

### CC-TASK-ENVELOPE

Public source: `data-formats/task.md`, section **Envelope**. Intended target: `internal/taskdecode/decode.go`. Planned case location: `internal/taskdecode/cross_cutting_conformance_test.go`, `symbol:TestCcTaskEnvelope`.

Case plan: Decode a minimal supported task, then omit each required envelope field, add an unknown field, violate task identifier syntax/byte bounds and exceed the complete encoded-byte bound. Check campaign-wide identity is preserved and no result is decoded as task input.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-TASK-0001 -->
<!-- requirement-ref: REQ-TASK-0003 -->
<!-- requirement-ref: REQ-TASK-0004 -->
<!-- requirement-ref: REQ-TASK-0005 -->

- [REQ-TASK-0001](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task.md#REQ-TASK-0001) → `CC-TASK-ENVELOPE`.
- [REQ-TASK-0003](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task.md#REQ-TASK-0003) → `CC-TASK-ENVELOPE`.
- [REQ-TASK-0004](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task.md#REQ-TASK-0004) → `CC-TASK-ENVELOPE`.
- [REQ-TASK-0005](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task.md#REQ-TASK-0005) → `CC-TASK-ENVELOPE`.

### CC-TASK-DESIGN-PRINCIPLES-STABLE-STRUCTURE

Public source: `data-formats/task.md`, section **Design Principles / Stable Structure**. Intended target: `docs/conformance/host-delivery-review.md`. Planned case location: `docs/conformance/host-delivery-review.md`, `heading:CC-TASK-DESIGN-PRINCIPLES-STABLE-STRUCTURE`.

Case plan: Review proposed schema extensions against the stable nested-object evolution rule; this is a documentation review, not a runtime test.

Scope: Host, collector or specification review; no runtime enforcement claim. No execution evidence or support claim.

<!-- requirement-ref: REQ-TASK-0006 -->

- [REQ-TASK-0006](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/task.md#REQ-TASK-0006) → `CC-TASK-DESIGN-PRINCIPLES-STABLE-STRUCTURE`.

### CC-CONTEXT-OVERVIEW

Public source: `library/network-context.md`, section **Overview**. Intended target: `internal/networkcontext/acquire.go`. Planned case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextOverview`.

Case plan: Run multiple tests on one generation/path; acquire shared bounded context rather than independently repeating per test.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-CONTEXT-0001 -->

- [REQ-CONTEXT-0001](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0001) → `CC-CONTEXT-OVERVIEW`.

### CC-CONTEXT-CONTEXT-LAYERS-ACCESS-CONTEXT

Public source: `library/network-context.md`, section **Context Layers / Access Context**. Intended target: `internal/networkcontext/acquire.go`. Planned case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextContextLayersAccessContext`.

Case plan: Expose only normalized coarse platform fields and inspect every listed forbidden identifier sink.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-CONTEXT-0002 -->
<!-- requirement-ref: REQ-CONTEXT-0003 -->

- [REQ-CONTEXT-0002](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0002) → `CC-CONTEXT-CONTEXT-LAYERS-ACCESS-CONTEXT`.
- [REQ-CONTEXT-0003](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0003) → `CC-CONTEXT-CONTEXT-LAYERS-ACCESS-CONTEXT`.

### CC-CONTEXT-CONTEXT-LAYERS-EGRESS-CONTEXT

Public source: `library/network-context.md`, section **Context Layers / Egress Context**. Intended target: `internal/networkcontext/acquire.go`. Planned case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextContextLayersEgressContext`.

Case plan: Use a synthetic endpoint observing source addresses; retain only coarse classification, and review endpoint logging and geographic/operator interpretation.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-CONTEXT-0004 -->
<!-- requirement-ref: REQ-CONTEXT-0005 -->
<!-- requirement-ref: REQ-CONTEXT-0006 -->

- [REQ-CONTEXT-0004](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0004) → `CC-CONTEXT-CONTEXT-LAYERS-EGRESS-CONTEXT`.
- [REQ-CONTEXT-0005](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0005) → `CC-CONTEXT-CONTEXT-LAYERS-EGRESS-CONTEXT`.
- [REQ-CONTEXT-0006](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0006) → `CC-CONTEXT-CONTEXT-LAYERS-EGRESS-CONTEXT`.

### CC-CONTEXT-HOST-INITIALIZATION

Public source: `library/network-context.md`, section **Host Initialization**. Intended target: `internal/networkcontext/acquire.go`. Planned case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextHostInitialization`.

Case plan: Validate endpoint URLs/fields/independent approvals, absence fallback to access-only, immutable configuration and rejection of redirects/discovery/credentials/per-user URLs.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-CONTEXT-0007 -->
<!-- requirement-ref: REQ-CONTEXT-0008 -->
<!-- requirement-ref: REQ-CONTEXT-0009 -->
<!-- requirement-ref: REQ-CONTEXT-0010 -->

- [REQ-CONTEXT-0007](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0007) → `CC-CONTEXT-HOST-INITIALIZATION`.
- [REQ-CONTEXT-0008](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0008) → `CC-CONTEXT-HOST-INITIALIZATION`.
- [REQ-CONTEXT-0009](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0009) → `CC-CONTEXT-HOST-INITIALIZATION`.
- [REQ-CONTEXT-0010](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0010) → `CC-CONTEXT-HOST-INITIALIZATION`.

### CC-CONTEXT-ENDPOINT-CONTRACT

Public source: `library/network-context.md`, section **Endpoint Contract**. Intended target: `internal/networkcontext/acquire.go`. Planned case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextEndpointContract`.

Case plan: Capture one fixed HTTPS GET without identifier/target/category/cookie/auth data. Validate exact response schema, 4096-byte/5000ms limits, forbidden fields and successful unknown classifications.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-CONTEXT-0011 -->
<!-- requirement-ref: REQ-CONTEXT-0012 -->
<!-- requirement-ref: REQ-CONTEXT-0013 -->
<!-- requirement-ref: REQ-CONTEXT-0014 -->
<!-- requirement-ref: REQ-CONTEXT-0015 -->
<!-- requirement-ref: REQ-CONTEXT-0016 -->

- [REQ-CONTEXT-0011](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0011) → `CC-CONTEXT-ENDPOINT-CONTRACT`.
- [REQ-CONTEXT-0012](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0012) → `CC-CONTEXT-ENDPOINT-CONTRACT`.
- [REQ-CONTEXT-0013](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0013) → `CC-CONTEXT-ENDPOINT-CONTRACT`.
- [REQ-CONTEXT-0014](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0014) → `CC-CONTEXT-ENDPOINT-CONTRACT`.
- [REQ-CONTEXT-0015](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0015) → `CC-CONTEXT-ENDPOINT-CONTRACT`.
- [REQ-CONTEXT-0016](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0016) → `CC-CONTEXT-ENDPOINT-CONTRACT`.

### CC-CONTEXT-PATH-CONSISTENCY

Public source: `library/network-context.md`, section **Path Consistency**. Intended target: `internal/networkcontext/acquire.go`. Planned case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextPathConsistency`.

Case plan: Run direct/proxy/NET-05 with generation changes; assert one matching snapshot and no direct request or cache substitution during proxy-only execution.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-CONTEXT-0017 -->
<!-- requirement-ref: REQ-CONTEXT-0018 -->
<!-- requirement-ref: REQ-CONTEXT-0019 -->

- [REQ-CONTEXT-0017](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0017) → `CC-CONTEXT-PATH-CONSISTENCY`.
- [REQ-CONTEXT-0018](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0018) → `CC-CONTEXT-PATH-CONSISTENCY`.
- [REQ-CONTEXT-0019](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0019) → `CC-CONTEXT-PATH-CONSISTENCY`.

### CC-CONTEXT-ACQUISITION-AND-FALLBACK

Public source: `library/network-context.md`, section **Acquisition and Fallback**. Intended target: `internal/networkcontext/acquire.go`. Planned case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextAcquisitionAndFallback`.

Case plan: Count primary/fallback calls, test valid-but-empty primary, malformed/TLS/HTTP/timeout/size failures, exact cache-key hits/misses, freshness and every invalidation event. Verify no persistent cache.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-CONTEXT-0020 -->
<!-- requirement-ref: REQ-CONTEXT-0021 -->
<!-- requirement-ref: REQ-CONTEXT-0022 -->
<!-- requirement-ref: REQ-CONTEXT-0023 -->
<!-- requirement-ref: REQ-CONTEXT-0024 -->

- [REQ-CONTEXT-0020](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0020) → `CC-CONTEXT-ACQUISITION-AND-FALLBACK`.
- [REQ-CONTEXT-0021](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0021) → `CC-CONTEXT-ACQUISITION-AND-FALLBACK`.
- [REQ-CONTEXT-0022](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0022) → `CC-CONTEXT-ACQUISITION-AND-FALLBACK`.
- [REQ-CONTEXT-0023](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0023) → `CC-CONTEXT-ACQUISITION-AND-FALLBACK`.
- [REQ-CONTEXT-0024](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0024) → `CC-CONTEXT-ACQUISITION-AND-FALLBACK`.

### CC-CONTEXT-EXECUTION-INTEGRATION

Public source: `library/network-context.md`, section **Execution Integration**. Intended target: `internal/networkcontext/acquire.go`. Planned case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextExecutionIntegration`.

Case plan: Inject cancellation/deadline/suspension/network changes during acquisition; distinguish missing required context, allowlist mismatch and optional egress failure.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-CONTEXT-0025 -->
<!-- requirement-ref: REQ-CONTEXT-0026 -->
<!-- requirement-ref: REQ-CONTEXT-0027 -->
<!-- requirement-ref: REQ-CONTEXT-0028 -->

- [REQ-CONTEXT-0025](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0025) → `CC-CONTEXT-EXECUTION-INTEGRATION`.
- [REQ-CONTEXT-0026](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0026) → `CC-CONTEXT-EXECUTION-INTEGRATION`.
- [REQ-CONTEXT-0027](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0027) → `CC-CONTEXT-EXECUTION-INTEGRATION`.
- [REQ-CONTEXT-0028](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0028) → `CC-CONTEXT-EXECUTION-INTEGRATION`.

### CC-CONTEXT-PROVIDER-AND-OPERATOR-INTERPRETATION

Public source: `library/network-context.md`, section **Provider and Operator Interpretation**. Intended target: `internal/networkcontext/acquire.go`. Planned case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextProviderAndOperatorInterpretation`.

Case plan: Reject inferred provider/operator data from locale/SIM/profile; retain only permitted ASN classifications and review analysis mappings.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-CONTEXT-0029 -->
<!-- requirement-ref: REQ-CONTEXT-0030 -->
<!-- requirement-ref: REQ-CONTEXT-0031 -->

- [REQ-CONTEXT-0029](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0029) → `CC-CONTEXT-PROVIDER-AND-OPERATOR-INTERPRETATION`.
- [REQ-CONTEXT-0030](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0030) → `CC-CONTEXT-PROVIDER-AND-OPERATOR-INTERPRETATION`.
- [REQ-CONTEXT-0031](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0031) → `CC-CONTEXT-PROVIDER-AND-OPERATOR-INTERPRETATION`.

### CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS

Public source: `library/network-context.md`, section **Privacy and Safety Requirements**. Intended target: `internal/networkcontext/acquire.go`. Planned case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextPrivacyAndSafetyRequirements`.

Case plan: Canary prohibited data in requests/results and review endpoint approval/log minimization/region necessity/aggregate reporting without censorship claims.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-CONTEXT-0032 -->
<!-- requirement-ref: REQ-CONTEXT-0035 -->
<!-- requirement-ref: REQ-CONTEXT-0036 -->
<!-- requirement-ref: REQ-CONTEXT-0037 -->
<!-- requirement-ref: REQ-CONTEXT-0038 -->
<!-- requirement-ref: REQ-CONTEXT-0039 -->
<!-- requirement-ref: REQ-CONTEXT-0040 -->
<!-- requirement-ref: REQ-CONTEXT-0041 -->
<!-- requirement-ref: REQ-CONTEXT-0042 -->

- [REQ-CONTEXT-0032](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0032) → `CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-CONTEXT-0035](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0035) → `CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-CONTEXT-0036](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0036) → `CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-CONTEXT-0037](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0037) → `CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-CONTEXT-0038](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0038) → `CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-CONTEXT-0039](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0039) → `CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-CONTEXT-0040](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0040) → `CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-CONTEXT-0041](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0041) → `CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-CONTEXT-0042](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0042) → `CC-CONTEXT-PRIVACY-AND-SAFETY-REQUIREMENTS`.

### CC-CONTEXT-PLATFORM-COLLECTION

Public source: `library/network-context.md`, section **Platform Collection**. Intended target: `internal/networkcontext/acquire.go`. Planned case location: `internal/networkcontext/cross_cutting_conformance_test.go`, `symbol:TestCcContextPlatformCollection`.

Case plan: Test normalized coarse Android/Apple adapter outputs and unsupported=unknown behavior; reject raw platform objects or sensitive transport details.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-CONTEXT-0033 -->
<!-- requirement-ref: REQ-CONTEXT-0034 -->

- [REQ-CONTEXT-0033](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0033) → `CC-CONTEXT-PLATFORM-COLLECTION`.
- [REQ-CONTEXT-0034](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/network-context.md#REQ-CONTEXT-0034) → `CC-CONTEXT-PLATFORM-COLLECTION`.

### CC-PROXY-OVERVIEW

Public source: `library/proxy.md`, section **Overview**. Intended target: `internal/execution/proxy.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcProxyOverview`.

Case plan: Reject transports beyond the registered single local HTTP/1.1 CONNECT stream capability.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-PROXY-0001 -->

- [REQ-PROXY-0001](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0001) → `CC-PROXY-OVERVIEW`.

### CC-PROXY-HOST-INTEGRATION

Public source: `library/proxy.md`, section **Host Integration**. Intended target: `internal/execution/proxy.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcProxyHostIntegration`.

Case plan: Exercise initialize/replace/clear via serializable values, and keep egress endpoint configuration separate; review bindings for prohibited native objects/callbacks.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-PROXY-0002 -->
<!-- requirement-ref: REQ-PROXY-0003 -->
<!-- requirement-ref: REQ-PROXY-0004 -->

- [REQ-PROXY-0002](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0002) → `CC-PROXY-HOST-INTEGRATION`.
- [REQ-PROXY-0003](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0003) → `CC-PROXY-HOST-INTEGRATION`.
- [REQ-PROXY-0004](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0004) → `CC-PROXY-HOST-INTEGRATION`.

### CC-PROXY-PROXY-CONFIGURATION

Public source: `library/proxy.md`, section **Proxy Configuration**. Intended target: `internal/execution/proxy.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcProxyProxyConfiguration`.

Case plan: Exercise loopback/localhost, port and connection_code boundaries and forbidden fields/transports; reject non-loopback endpoints.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-PROXY-0005 -->
<!-- requirement-ref: REQ-PROXY-0006 -->
<!-- requirement-ref: REQ-PROXY-0007 -->

- [REQ-PROXY-0005](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0005) → `CC-PROXY-PROXY-CONFIGURATION`.
- [REQ-PROXY-0006](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0006) → `CC-PROXY-PROXY-CONFIGURATION`.
- [REQ-PROXY-0007](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0007) → `CC-PROXY-PROXY-CONFIGURATION`.

### CC-PROXY-CONFIGURATION-VALIDATION

Public source: `library/proxy.md`, section **Configuration Validation**. Intended target: `internal/execution/proxy.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcProxyConfigurationValidation`.

Case plan: Attempt malformed replacement while an old proxy is active; preserve old state. Test category stability/privacy and internally derived compatibility.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-PROXY-0008 -->
<!-- requirement-ref: REQ-PROXY-0009 -->
<!-- requirement-ref: REQ-PROXY-0010 -->

- [REQ-PROXY-0008](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0008) → `CC-PROXY-CONFIGURATION-VALIDATION`.
- [REQ-PROXY-0009](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0009) → `CC-PROXY-CONFIGURATION-VALIDATION`.
- [REQ-PROXY-0010](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0010) → `CC-PROXY-CONFIGURATION-VALIDATION`.

### CC-PROXY-LIFECYCLE-AND-RENEWAL

Public source: `library/proxy.md`, section **Lifecycle and Renewal**. Intended target: `internal/execution/proxy.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcProxyLifecycleAndRenewal`.

Case plan: Race set/clear with attempts and all NET-05 steps; block old-generation admission, cancel/close and bound cleanup before returning, invalidate caches and never mix generations or use direct fallback.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-PROXY-0011 -->
<!-- requirement-ref: REQ-PROXY-0012 -->
<!-- requirement-ref: REQ-PROXY-0013 -->
<!-- requirement-ref: REQ-PROXY-0014 -->
<!-- requirement-ref: REQ-PROXY-0015 -->
<!-- requirement-ref: REQ-PROXY-0016 -->

- [REQ-PROXY-0011](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0011) → `CC-PROXY-LIFECYCLE-AND-RENEWAL`.
- [REQ-PROXY-0012](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0012) → `CC-PROXY-LIFECYCLE-AND-RENEWAL`.
- [REQ-PROXY-0013](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0013) → `CC-PROXY-LIFECYCLE-AND-RENEWAL`.
- [REQ-PROXY-0014](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0014) → `CC-PROXY-LIFECYCLE-AND-RENEWAL`.
- [REQ-PROXY-0015](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0015) → `CC-PROXY-LIFECYCLE-AND-RENEWAL`.
- [REQ-PROXY-0016](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0016) → `CC-PROXY-LIFECYCLE-AND-RENEWAL`.

### CC-PROXY-HTTP-CONNECT-BEHAVIOR

Public source: `library/proxy.md`, section **HTTP CONNECT Behavior**. Intended target: `internal/execution/proxy.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcProxyHttpConnectBehavior`.

Case plan: Capture synthetic CONNECT exchanges for explicit IP and hostname cases, 2xx/non-2xx/malformed/timeouts, redirects and cancellation. Assert exact authority, no credentials/cleartext target GET, fresh tunnels and unchanged TLS semantics.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-PROXY-0017 -->
<!-- requirement-ref: REQ-PROXY-0018 -->
<!-- requirement-ref: REQ-PROXY-0019 -->
<!-- requirement-ref: REQ-PROXY-0020 -->

- [REQ-PROXY-0017](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0017) → `CC-PROXY-HTTP-CONNECT-BEHAVIOR`.
- [REQ-PROXY-0018](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0018) → `CC-PROXY-HTTP-CONNECT-BEHAVIOR`.
- [REQ-PROXY-0019](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0019) → `CC-PROXY-HTTP-CONNECT-BEHAVIOR`.
- [REQ-PROXY-0020](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0020) → `CC-PROXY-HTTP-CONNECT-BEHAVIOR`.

### CC-PROXY-TASK-SELECTION

Public source: `library/proxy.md`, section **Task Selection**. Intended target: `internal/execution/proxy.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcProxyTaskSelection`.

Case plan: Exercise direct default and proxy preconditions; task cannot carry endpoint, category or proxy identity.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-PROXY-0021 -->
<!-- requirement-ref: REQ-PROXY-0022 -->

- [REQ-PROXY-0021](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0021) → `CC-PROXY-TASK-SELECTION`.
- [REQ-PROXY-0022](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0022) → `CC-PROXY-TASK-SELECTION`.

### CC-PROXY-NETWORK-CONTEXT

Public source: `library/proxy.md`, section **Network Context**. Intended target: `internal/execution/proxy.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcProxyNetworkContext`.

Case plan: Fail primary context endpoint and verify same-generation proxied fallback, exact cache keys and required-context gating with zero direct requests.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-PROXY-0023 -->
<!-- requirement-ref: REQ-PROXY-0024 -->
<!-- requirement-ref: REQ-PROXY-0025 -->

- [REQ-PROXY-0023](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0023) → `CC-PROXY-NETWORK-CONTEXT`.
- [REQ-PROXY-0024](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0024) → `CC-PROXY-NETWORK-CONTEXT`.
- [REQ-PROXY-0025](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0025) → `CC-PROXY-NETWORK-CONTEXT`.

### CC-PROXY-INITIAL-TEST-COMPATIBILITY

Public source: `library/proxy.md`, section **Initial Test Compatibility**. Intended target: `internal/execution/proxy.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcProxyInitialTestCompatibility`.

Case plan: Parameterize all nine registered tests and relevant input/options combinations, including each incompatible NET-05 step; reject atomically before activity.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-PROXY-0026 -->
<!-- requirement-ref: REQ-PROXY-0027 -->

- [REQ-PROXY-0026](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0026) → `CC-PROXY-INITIAL-TEST-COMPATIBILITY`.
- [REQ-PROXY-0027](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0027) → `CC-PROXY-INITIAL-TEST-COMPATIBILITY`.

### CC-PROXY-EXECUTION-STATES

Public source: `library/proxy.md`, section **Execution States**. Intended target: `internal/execution/proxy.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcProxyExecutionStates`.

Case plan: Trigger each path failure and assert normalized state, no censorship attribution, no direct retry, and any new generation needs a separately scheduled attempt.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-PROXY-0028 -->
<!-- requirement-ref: REQ-PROXY-0029 -->
<!-- requirement-ref: REQ-PROXY-0030 -->

- [REQ-PROXY-0028](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0028) → `CC-PROXY-EXECUTION-STATES`.
- [REQ-PROXY-0029](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0029) → `CC-PROXY-EXECUTION-STATES`.
- [REQ-PROXY-0030](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0030) → `CC-PROXY-EXECUTION-STATES`.

### CC-PROXY-RESULT-METADATA

Public source: `library/proxy.md`, section **Result Metadata**. Intended target: `internal/execution/proxy.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcProxyResultMetadata`.

Case plan: Produce direct/proxy success/failure results and inspect normalized path/category while excluding endpoint/generation/CONNECT/raw transport configuration.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-PROXY-0031 -->
<!-- requirement-ref: REQ-PROXY-0032 -->
<!-- requirement-ref: REQ-PROXY-0033 -->

- [REQ-PROXY-0031](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0031) → `CC-PROXY-RESULT-METADATA`.
- [REQ-PROXY-0032](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0032) → `CC-PROXY-RESULT-METADATA`.
- [REQ-PROXY-0033](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0033) → `CC-PROXY-RESULT-METADATA`.

### CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION

Public source: `library/proxy.md`, section **Privacy, Safety, and Interpretation**. Intended target: `internal/execution/proxy.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcProxyPrivacySafetyAndInterpretation`.

Case plan: Attempt task-driven endpoint/target/policy/identity bypasses; review path-comparison claims against repeated equivalent inputs and independent evidence.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-PROXY-0034 -->
<!-- requirement-ref: REQ-PROXY-0035 -->
<!-- requirement-ref: REQ-PROXY-0036 -->
<!-- requirement-ref: REQ-PROXY-0037 -->
<!-- requirement-ref: REQ-PROXY-0038 -->
<!-- requirement-ref: REQ-PROXY-0039 -->
<!-- requirement-ref: REQ-PROXY-0040 -->
<!-- requirement-ref: REQ-PROXY-0041 -->

- [REQ-PROXY-0034](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0034) → `CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION`.
- [REQ-PROXY-0035](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0035) → `CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION`.
- [REQ-PROXY-0036](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0036) → `CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION`.
- [REQ-PROXY-0037](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0037) → `CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION`.
- [REQ-PROXY-0038](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0038) → `CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION`.
- [REQ-PROXY-0039](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0039) → `CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION`.
- [REQ-PROXY-0040](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0040) → `CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION`.
- [REQ-PROXY-0041](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/proxy.md#REQ-PROXY-0041) → `CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION`.

### CC-LIMIT-PURPOSE

Public source: `library/registered-limits.md`, section **Purpose**. Intended target: `internal/validation/limits.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcLimitPurpose`.

Case plan: Apply stricter local/registered values; exceed limits without silent clamping. Review application-byte versus lower-layer overhead interpretation.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LIMIT-0001 -->
<!-- requirement-ref: REQ-LIMIT-0002 -->
<!-- requirement-ref: REQ-LIMIT-0003 -->

- [REQ-LIMIT-0001](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0001) → `CC-LIMIT-PURPOSE`.
- [REQ-LIMIT-0002](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0002) → `CC-LIMIT-PURPOSE`.
- [REQ-LIMIT-0003](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0003) → `CC-LIMIT-PURPOSE`.

### CC-LIMIT-TASK-AND-IDENTIFIER-LIMITS

Public source: `library/registered-limits.md`, section **Task and Identifier Limits**. Intended target: `internal/validation/limits.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcLimitTaskAndIdentifierLimits`.

Case plan: Parameterize every table row at below/equal/above bounds using original encoded bytes; include whitespace/unknown fields and stop parsing at task maximum.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LIMIT-0004 -->
<!-- requirement-ref: REQ-LIMIT-0005 -->

- [REQ-LIMIT-0004](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0004) → `CC-LIMIT-TASK-AND-IDENTIFIER-LIMITS`.
- [REQ-LIMIT-0005](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0005) → `CC-LIMIT-TASK-AND-IDENTIFIER-LIMITS`.

### CC-LIMIT-ATTEMPT-LIMITS

Public source: `library/registered-limits.md`, section **Attempt Limits**. Intended target: `internal/execution/budget.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcLimitAttemptLimits`.

Case plan: Exercise every default/maximum and NET-01 zero-traffic exception, 15000ms protocol ceiling, complete-attempt deadline and 2000ms cleanup budget with fake clocks/transports.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LIMIT-0006 -->
<!-- requirement-ref: REQ-LIMIT-0007 -->
<!-- requirement-ref: REQ-LIMIT-0008 -->

- [REQ-LIMIT-0006](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0006) → `CC-LIMIT-ATTEMPT-LIMITS`.
- [REQ-LIMIT-0007](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0007) → `CC-LIMIT-ATTEMPT-LIMITS`.
- [REQ-LIMIT-0008](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0008) → `CC-LIMIT-ATTEMPT-LIMITS`.

### CC-LIMIT-SCHEDULE-AND-RETRY-LIMITS

Public source: `library/registered-limits.md`, section **Schedule and Retry Limits**. Intended target: `internal/validation/schedule.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcLimitScheduleAndRetryLimits`.

Case plan: Parameterize every bound/default, issuance skew and trusted-time absence, embedded one-time exception and forbidden measurement-outcome retries.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LIMIT-0009 -->
<!-- requirement-ref: REQ-LIMIT-0010 -->
<!-- requirement-ref: REQ-LIMIT-0011 -->
<!-- requirement-ref: REQ-LIMIT-0012 -->
<!-- requirement-ref: REQ-LIMIT-0013 -->

- [REQ-LIMIT-0009](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0009) → `CC-LIMIT-SCHEDULE-AND-RETRY-LIMITS`.
- [REQ-LIMIT-0010](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0010) → `CC-LIMIT-SCHEDULE-AND-RETRY-LIMITS`.
- [REQ-LIMIT-0011](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0011) → `CC-LIMIT-SCHEDULE-AND-RETRY-LIMITS`.
- [REQ-LIMIT-0012](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0012) → `CC-LIMIT-SCHEDULE-AND-RETRY-LIMITS`.
- [REQ-LIMIT-0013](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0013) → `CC-LIMIT-SCHEDULE-AND-RETRY-LIMITS`.

### CC-LIMIT-DNS-LIMITS

Public source: `library/registered-limits.md`, section **DNS Limits**. Intended target: `internal/protocol/dns/limits.go`. Planned case location: `internal/protocol/dns/cross_cutting_conformance_test.go`, `symbol:TestCcLimitDnsLimits`.

Case plan: Parameterize every DNS byte/count/window/answer/CNAME bound; report limit outcome without partial append or prohibited raw DNS fields.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LIMIT-0014 -->
<!-- requirement-ref: REQ-LIMIT-0015 -->

- [REQ-LIMIT-0014](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0014) → `CC-LIMIT-DNS-LIMITS`.
- [REQ-LIMIT-0015](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0015) → `CC-LIMIT-DNS-LIMITS`.

### CC-LIMIT-HTTP-AND-VALIDATION-LIMITS

Public source: `library/registered-limits.md`, section **HTTP and Validation Limits**. Intended target: `internal/protocol/http/limits.go`. Planned case location: `internal/protocol/http/cross_cutting_conformance_test.go`, `symbol:TestCcLimitHttpAndValidationLimits`.

Case plan: Parameterize URL/header/body/check/matcher/redirect/ALPN bounds and identity encoding; test undecidable bounded checks and NET-03 redirect non-following.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LIMIT-0016 -->
<!-- requirement-ref: REQ-LIMIT-0017 -->
<!-- requirement-ref: REQ-LIMIT-0018 -->

- [REQ-LIMIT-0016](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0016) → `CC-LIMIT-HTTP-AND-VALIDATION-LIMITS`.
- [REQ-LIMIT-0017](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0017) → `CC-LIMIT-HTTP-AND-VALIDATION-LIMITS`.
- [REQ-LIMIT-0018](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0018) → `CC-LIMIT-HTTP-AND-VALIDATION-LIMITS`.

### CC-LIMIT-NET-05-COMPOSITE-LIMITS

Public source: `library/registered-limits.md`, section **NET-05 Composite Limits**. Intended target: `internal/validation/profile.go`. Planned case location: `internal/validation/cross_cutting_conformance_test.go`, `symbol:TestCcLimitNet05CompositeLimits`.

Case plan: Parameterize profile/step/identity/aggregate budgets, nested limits and sum-of-declared-maxima overflow; reject the full profile before traffic.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LIMIT-0019 -->
<!-- requirement-ref: REQ-LIMIT-0020 -->

- [REQ-LIMIT-0019](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0019) → `CC-LIMIT-NET-05-COMPOSITE-LIMITS`.
- [REQ-LIMIT-0020](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0020) → `CC-LIMIT-NET-05-COMPOSITE-LIMITS`.

### CC-LIMIT-RESULT-AND-BUNDLE-LIMITS

Public source: `library/registered-limits.md`, section **Result and Bundle Limits**. Intended target: `internal/bundle/bounds.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcLimitResultAndBundleLimits`.

Case plan: Exercise every encoded/count/pending-storage maximum; split bundles or return bounded persistence error without truncating valid results.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LIMIT-0021 -->
<!-- requirement-ref: REQ-LIMIT-0022 -->

- [REQ-LIMIT-0021](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0021) → `CC-LIMIT-RESULT-AND-BUNDLE-LIMITS`.
- [REQ-LIMIT-0022](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0022) → `CC-LIMIT-RESULT-AND-BUNDLE-LIMITS`.

### CC-LIMIT-CONTEXT-LINKAGE-AND-RETENTION-LIMITS

Public source: `library/registered-limits.md`, section **Context, Linkage, and Retention Limits**. Intended target: `internal/persistence/retention.go`. Planned case location: `internal/persistence/cross_cutting_conformance_test.go`, `symbol:TestCcLimitContextLinkageAndRetentionLimits`.

Case plan: Exercise every endpoint/cache/rotation/retention/replay maximum with deterministic time; durable ack deletes immediately, expiry prevents late upload, and deployment retention needs separate review.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LIMIT-0023 -->
<!-- requirement-ref: REQ-LIMIT-0024 -->
<!-- requirement-ref: REQ-LIMIT-0025 -->

- [REQ-LIMIT-0023](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0023) → `CC-LIMIT-CONTEXT-LINKAGE-AND-RETENTION-LIMITS`.
- [REQ-LIMIT-0024](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0024) → `CC-LIMIT-CONTEXT-LINKAGE-AND-RETENTION-LIMITS`.
- [REQ-LIMIT-0025](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0025) → `CC-LIMIT-CONTEXT-LINKAGE-AND-RETENTION-LIMITS`.

### CC-LIMIT-LIMIT-HANDLING

Public source: `library/registered-limits.md`, section **Limit Handling**. Intended target: `internal/execution/budget.go`. Planned case location: `internal/execution/cross_cutting_conformance_test.go`, `symbol:TestCcLimitLimitHandling`.

Case plan: Inject admission and runtime limit violations; close each resource class, emit only registered outcomes/fields, deny task bypasses and review increases against version/threat-model gates.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LIMIT-0026 -->
<!-- requirement-ref: REQ-LIMIT-0027 -->
<!-- requirement-ref: REQ-LIMIT-0028 -->
<!-- requirement-ref: REQ-LIMIT-0029 -->
<!-- requirement-ref: REQ-LIMIT-0030 -->
<!-- requirement-ref: REQ-LIMIT-0031 -->

- [REQ-LIMIT-0026](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0026) → `CC-LIMIT-LIMIT-HANDLING`.
- [REQ-LIMIT-0027](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0027) → `CC-LIMIT-LIMIT-HANDLING`.
- [REQ-LIMIT-0028](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0028) → `CC-LIMIT-LIMIT-HANDLING`.
- [REQ-LIMIT-0029](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0029) → `CC-LIMIT-LIMIT-HANDLING`.
- [REQ-LIMIT-0030](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0030) → `CC-LIMIT-LIMIT-HANDLING`.
- [REQ-LIMIT-0031](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/registered-limits.md#REQ-LIMIT-0031) → `CC-LIMIT-LIMIT-HANDLING`.

### CC-LINKAGE-OVERVIEW

Public source: `library/result-linkage.md`, section **Overview**. Intended target: `internal/bundle/linkage.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcLinkageOverview`.

Case plan: Initialize without linkage; verify no cross-bundle token and review pseudonymous terminology.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LINKAGE-0001 -->

- [REQ-LINKAGE-0001](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0001) → `CC-LINKAGE-OVERVIEW`.

### CC-LINKAGE-GOALS-AND-NON-GOALS

Public source: `library/result-linkage.md`, section **Goals and Non-goals**. Intended target: `internal/bundle/linkage.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcLinkageGoalsAndNonGoals`.

Case plan: Attempt cross-app/deployment/reinstall/reset/account restoration and targeted delivery; no stable identity continuity is permitted.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LINKAGE-0002 -->

- [REQ-LINKAGE-0002](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0002) → `CC-LINKAGE-GOALS-AND-NON-GOALS`.

### CC-LINKAGE-INITIALIZATION-CONFIGURATION

Public source: `library/result-linkage.md`, section **Initialization Configuration**. Intended target: `internal/bundle/linkage.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcLinkageInitializationConfiguration`.

Case plan: Exercise mode-dependent fields, scope syntax/non-user-specific review, period bounds, default none and trusted reinitialization/local-policy control.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LINKAGE-0003 -->
<!-- requirement-ref: REQ-LINKAGE-0004 -->
<!-- requirement-ref: REQ-LINKAGE-0005 -->
<!-- requirement-ref: REQ-LINKAGE-0006 -->

- [REQ-LINKAGE-0003](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0003) → `CC-LINKAGE-INITIALIZATION-CONFIGURATION`.
- [REQ-LINKAGE-0004](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0004) → `CC-LINKAGE-INITIALIZATION-CONFIGURATION`.
- [REQ-LINKAGE-0005](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0005) → `CC-LINKAGE-INITIALIZATION-CONFIGURATION`.
- [REQ-LINKAGE-0006](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0006) → `CC-LINKAGE-INITIALIZATION-CONFIGURATION`.

### CC-LINKAGE-INSTALLATION-SECRET

Public source: `library/result-linkage.md`, section **Installation Secret**. Intended target: `internal/bundle/linkage.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcLinkageInstallationSecret`.

Case plan: Use fake protected-key capabilities and controlled randomness to check fresh 256-bit key creation, no host-derived secret, protected ownership and no export/backup/network/log/analytics sinks.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LINKAGE-0007 -->
<!-- requirement-ref: REQ-LINKAGE-0008 -->
<!-- requirement-ref: REQ-LINKAGE-0009 -->

- [REQ-LINKAGE-0007](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0007) → `CC-LINKAGE-INSTALLATION-SECRET`.
- [REQ-LINKAGE-0008](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0008) → `CC-LINKAGE-INSTALLATION-SECRET`.
- [REQ-LINKAGE-0009](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0009) → `CC-LINKAGE-INSTALLATION-SECRET`.

### CC-LINKAGE-TOKEN-DERIVATION

Public source: `library/result-linkage.md`, section **Token Derivation**. Intended target: `internal/bundle/linkage.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcLinkageTokenDerivation`.

Case plan: Use independent HMAC known-answer vectors for exact purpose/separator/scope/decimal epoch message, floor boundaries, base64url and memory cleanup; suppress linkage on invalid/rollback clock until nondecreasing validity.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LINKAGE-0010 -->
<!-- requirement-ref: REQ-LINKAGE-0011 -->
<!-- requirement-ref: REQ-LINKAGE-0012 -->
<!-- requirement-ref: REQ-LINKAGE-0013 -->
<!-- requirement-ref: REQ-LINKAGE-0014 -->

- [REQ-LINKAGE-0010](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0010) → `CC-LINKAGE-TOKEN-DERIVATION`.
- [REQ-LINKAGE-0011](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0011) → `CC-LINKAGE-TOKEN-DERIVATION`.
- [REQ-LINKAGE-0012](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0012) → `CC-LINKAGE-TOKEN-DERIVATION`.
- [REQ-LINKAGE-0013](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0013) → `CC-LINKAGE-TOKEN-DERIVATION`.
- [REQ-LINKAGE-0014](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0014) → `CC-LINKAGE-TOKEN-DERIVATION`.

### CC-LINKAGE-ROTATION-SEMANTICS

Public source: `library/result-linkage.md`, section **Rotation Semantics**. Intended target: `internal/bundle/linkage.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcLinkageRotationSemantics`.

Case plan: Compare same/different scopes/epochs across boundaries; no old-new token mapping or continuity request, and no period extension beyond reviewed bounds.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LINKAGE-0015 -->
<!-- requirement-ref: REQ-LINKAGE-0016 -->
<!-- requirement-ref: REQ-LINKAGE-0017 -->

- [REQ-LINKAGE-0015](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0015) → `CC-LINKAGE-ROTATION-SEMANTICS`.
- [REQ-LINKAGE-0016](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0016) → `CC-LINKAGE-ROTATION-SEMANTICS`.
- [REQ-LINKAGE-0017](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0017) → `CC-LINKAGE-ROTATION-SEMANTICS`.

### CC-LINKAGE-GO-CORE-AND-MOBILE-BINDING-ARCHITECTURE

Public source: `library/result-linkage.md`, section **Go Core and Mobile Binding Architecture**. Intended target: `docs/conformance/native-linkage.md`. Planned case location: `docs/conformance/native-linkage.md`, `heading:CC-LINKAGE-GO-CORE-AND-MOBILE-BINDING-ARCHITECTURE`.

Case plan: Inspect Go/native/Flutter boundary for internal non-exporting capabilities and plain configuration/lifecycle only; reject hidden mobile in-memory fallbacks.

Scope: Native adapter and device release validation; planned outside portable Go runtime. No execution evidence or support claim.

<!-- requirement-ref: REQ-LINKAGE-0018 -->
<!-- requirement-ref: REQ-LINKAGE-0019 -->
<!-- requirement-ref: REQ-LINKAGE-0020 -->
<!-- requirement-ref: REQ-LINKAGE-0021 -->
<!-- requirement-ref: REQ-LINKAGE-0022 -->

- [REQ-LINKAGE-0018](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0018) → `CC-LINKAGE-GO-CORE-AND-MOBILE-BINDING-ARCHITECTURE`.
- [REQ-LINKAGE-0019](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0019) → `CC-LINKAGE-GO-CORE-AND-MOBILE-BINDING-ARCHITECTURE`.
- [REQ-LINKAGE-0020](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0020) → `CC-LINKAGE-GO-CORE-AND-MOBILE-BINDING-ARCHITECTURE`.
- [REQ-LINKAGE-0021](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0021) → `CC-LINKAGE-GO-CORE-AND-MOBILE-BINDING-ARCHITECTURE`.
- [REQ-LINKAGE-0022](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0022) → `CC-LINKAGE-GO-CORE-AND-MOBILE-BINDING-ARCHITECTURE`.

### CC-LINKAGE-CLEAR-DATA-CONTRACT

Public source: `library/result-linkage.md`, section **Clear-data Contract**. Intended target: `internal/persistence/clear.go`. Planned case location: `internal/persistence/cross_cutting_conformance_test.go`, `symbol:TestCcLinkageClearDataContract`.

Case plan: Fault-inject each clearing step with active attempts and pending/acked bundles, caches, replay/linkage/sentinel/key state. Assert bounded cleanup, idempotence, explicit incomplete error, no upload/network/new secret, and host UI/OS reset wiring.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LINKAGE-0023 -->
<!-- requirement-ref: REQ-LINKAGE-0024 -->
<!-- requirement-ref: REQ-LINKAGE-0025 -->
<!-- requirement-ref: REQ-LINKAGE-0026 -->

- [REQ-LINKAGE-0023](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0023) → `CC-LINKAGE-CLEAR-DATA-CONTRACT`.
- [REQ-LINKAGE-0024](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0024) → `CC-LINKAGE-CLEAR-DATA-CONTRACT`.
- [REQ-LINKAGE-0025](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0025) → `CC-LINKAGE-CLEAR-DATA-CONTRACT`.
- [REQ-LINKAGE-0026](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0026) → `CC-LINKAGE-CLEAR-DATA-CONTRACT`.

### CC-LINKAGE-ANDROID-STORAGE

Public source: `library/result-linkage.md`, section **Android Storage**. Intended target: `docs/conformance/native-linkage.md`. Planned case location: `docs/conformance/native-linkage.md`, `heading:CC-LINKAGE-ANDROID-STORAGE`.

Case plan: Plan device checks for non-exporting application Keystore and no-backup state, matching generations/orphan recovery; run explicit/system clear, reinstall, backup/restore/migration and interruption matrix without token continuity.

Scope: Native adapter and device release validation; planned outside portable Go runtime. No execution evidence or support claim.

<!-- requirement-ref: REQ-LINKAGE-0027 -->
<!-- requirement-ref: REQ-LINKAGE-0028 -->
<!-- requirement-ref: REQ-LINKAGE-0029 -->
<!-- requirement-ref: REQ-LINKAGE-0030 -->
<!-- requirement-ref: REQ-LINKAGE-0031 -->

- [REQ-LINKAGE-0027](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0027) → `CC-LINKAGE-ANDROID-STORAGE`.
- [REQ-LINKAGE-0028](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0028) → `CC-LINKAGE-ANDROID-STORAGE`.
- [REQ-LINKAGE-0029](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0029) → `CC-LINKAGE-ANDROID-STORAGE`.
- [REQ-LINKAGE-0030](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0030) → `CC-LINKAGE-ANDROID-STORAGE`.
- [REQ-LINKAGE-0031](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0031) → `CC-LINKAGE-ANDROID-STORAGE`.

### CC-LINKAGE-APPLE-PLATFORM-STORAGE

Public source: `library/result-linkage.md`, section **Apple Platform Storage**. Intended target: `docs/conformance/native-linkage.md`. Planned case location: `docs/conformance/native-linkage.md`, `heading:CC-LINKAGE-APPLE-PLATFORM-STORAGE`.

Case plan: Plan device checks for required accessibility/private non-sync Keychain, sentinel truth table and restored-container ambiguity; run uninstall, backup/restore, migration, iCloud settings and interrupted deletion matrix, preferring reset.

Scope: Native adapter and device release validation; planned outside portable Go runtime. No execution evidence or support claim.

<!-- requirement-ref: REQ-LINKAGE-0032 -->
<!-- requirement-ref: REQ-LINKAGE-0033 -->
<!-- requirement-ref: REQ-LINKAGE-0034 -->
<!-- requirement-ref: REQ-LINKAGE-0035 -->
<!-- requirement-ref: REQ-LINKAGE-0036 -->

- [REQ-LINKAGE-0032](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0032) → `CC-LINKAGE-APPLE-PLATFORM-STORAGE`.
- [REQ-LINKAGE-0033](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0033) → `CC-LINKAGE-APPLE-PLATFORM-STORAGE`.
- [REQ-LINKAGE-0034](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0034) → `CC-LINKAGE-APPLE-PLATFORM-STORAGE`.
- [REQ-LINKAGE-0035](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0035) → `CC-LINKAGE-APPLE-PLATFORM-STORAGE`.
- [REQ-LINKAGE-0036](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0036) → `CC-LINKAGE-APPLE-PLATFORM-STORAGE`.

### CC-LINKAGE-STORAGE-FAILURE

Public source: `library/result-linkage.md`, section **Storage Failure**. Intended target: `internal/bundle/linkage.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcLinkageStorageFailure`.

Case plan: Inject unavailable/locked/inconsistent/invalid storage; omit linkage with permitted bundle grouping, preserve deletion, forbid weak persistence/host-visible/sticky-token fallback and sanitize diagnostics.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-LINKAGE-0037 -->
<!-- requirement-ref: REQ-LINKAGE-0038 -->
<!-- requirement-ref: REQ-LINKAGE-0039 -->

- [REQ-LINKAGE-0037](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0037) → `CC-LINKAGE-STORAGE-FAILURE`.
- [REQ-LINKAGE-0038](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0038) → `CC-LINKAGE-STORAGE-FAILURE`.
- [REQ-LINKAGE-0039](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0039) → `CC-LINKAGE-STORAGE-FAILURE`.

### CC-LINKAGE-INGESTION-SEPARATION

Public source: `library/result-linkage.md`, section **Ingestion Separation**. Intended target: `docs/conformance/host-delivery-review.md`. Planned case location: `docs/conformance/host-delivery-review.md`, `heading:CC-LINKAGE-INGESTION-SEPARATION`.

Case plan: Review a synthetic deployment flow for source-IP/header/account separation, rate/schema/idempotency bounds and required privacy/safety review for account-authenticated delivery.

Scope: Host, collector or specification review; no runtime enforcement claim. No execution evidence or support claim.

<!-- requirement-ref: REQ-LINKAGE-0040 -->
<!-- requirement-ref: REQ-LINKAGE-0041 -->
<!-- requirement-ref: REQ-LINKAGE-0042 -->

- [REQ-LINKAGE-0040](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0040) → `CC-LINKAGE-INGESTION-SEPARATION`.
- [REQ-LINKAGE-0041](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0041) → `CC-LINKAGE-INGESTION-SEPARATION`.
- [REQ-LINKAGE-0042](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0042) → `CC-LINKAGE-INGESTION-SEPARATION`.

### CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS

Public source: `library/result-linkage.md`, section **Privacy and Safety Requirements**. Intended target: `internal/bundle/linkage.go`. Planned case location: `internal/bundle/cross_cutting_conformance_test.go`, `symbol:TestCcLinkagePrivacyAndSafetyRequirements`.

Case plan: Inspect all default/control/export/use/context/target/log/retention/reset boundaries with synthetic canaries and review re-identification combinations.

Scope: Synthetic core contract cases. No execution evidence or support claim.

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

- [REQ-LINKAGE-0043](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0043) → `CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-LINKAGE-0044](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0044) → `CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-LINKAGE-0045](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0045) → `CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-LINKAGE-0046](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0046) → `CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-LINKAGE-0047](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0047) → `CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-LINKAGE-0048](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0048) → `CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-LINKAGE-0049](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0049) → `CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-LINKAGE-0050](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0050) → `CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-LINKAGE-0051](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0051) → `CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`.
- [REQ-LINKAGE-0052](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/result-linkage.md#REQ-LINKAGE-0052) → `CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`.

### CC-TRUST-OVERVIEW

Public source: `library/task-trust.md`, section **Overview**. Intended target: `internal/admission/admit.go`. Planned case location: `internal/admission/cross_cutting_conformance_test.go`, `symbol:TestCcTrustOverview`.

Case plan: Supply forged task-controlled authority fields; only separately established provenance, registered behavior and trusted policy can authorize admission.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-TRUST-0002 -->
<!-- requirement-ref: REQ-TRUST-0003 -->

- [REQ-TRUST-0002](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0002) → `CC-TRUST-OVERVIEW`.
- [REQ-TRUST-0003](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0003) → `CC-TRUST-OVERVIEW`.

### CC-TRUST-ACCEPTED-TASK-ORIGINS

Public source: `library/task-trust.md`, section **Accepted Task Origins**. Intended target: `internal/admission/admit.go`. Planned case location: `internal/admission/cross_cutting_conformance_test.go`, `symbol:TestCcTrustAcceptedTaskOrigins`.

Case plan: Exercise reviewed embedded origin and authenticated remote origin, then independently remove authentication, integrity, required time and context; fail closed.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-TRUST-0001 -->
<!-- requirement-ref: REQ-TRUST-0004 -->
<!-- requirement-ref: REQ-TRUST-0005 -->

- [REQ-TRUST-0001](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0001) → `CC-TRUST-ACCEPTED-TASK-ORIGINS`.
- [REQ-TRUST-0004](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0004) → `CC-TRUST-ACCEPTED-TASK-ORIGINS`.
- [REQ-TRUST-0005](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0005) → `CC-TRUST-ACCEPTED-TASK-ORIGINS`.

### CC-TRUST-TRUSTED-SOURCE-CONTEXT

Public source: `library/task-trust.md`, section **Trusted Source Context**. Intended target: `internal/admission/admit.go`. Planned case location: `internal/admission/cross_cutting_conformance_test.go`, `symbol:TestCcTrustTrustedSourceContext`.

Case plan: Validate each origin/issuer/generation/authentication/issued_at relationship and separate trust boundary; reject task/Flutter supplied authentication and prevent source metadata in traffic/results.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-TRUST-0006 -->
<!-- requirement-ref: REQ-TRUST-0007 -->
<!-- requirement-ref: REQ-TRUST-0008 -->
<!-- requirement-ref: REQ-TRUST-0009 -->

- [REQ-TRUST-0006](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0006) → `CC-TRUST-TRUSTED-SOURCE-CONTEXT`.
- [REQ-TRUST-0007](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0007) → `CC-TRUST-TRUSTED-SOURCE-CONTEXT`.
- [REQ-TRUST-0008](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0008) → `CC-TRUST-TRUSTED-SOURCE-CONTEXT`.
- [REQ-TRUST-0009](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0009) → `CC-TRUST-TRUSTED-SOURCE-CONTEXT`.

### CC-TRUST-ISSUER-AUTHORIZATION

Public source: `library/task-trust.md`, section **Issuer Authorization**. Intended target: `internal/admission/admit.go`. Planned case location: `internal/admission/cross_cutting_conformance_test.go`, `symbol:TestCcTrustIssuerAuthorization`.

Case plan: Narrow each authority dimension in turn; effective authority is the intersection and task data cannot widen any dimension.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-TRUST-0010 -->
<!-- requirement-ref: REQ-TRUST-0011 -->
<!-- requirement-ref: REQ-TRUST-0012 -->

- [REQ-TRUST-0010](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0010) → `CC-TRUST-ISSUER-AUTHORIZATION`.
- [REQ-TRUST-0011](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0011) → `CC-TRUST-ISSUER-AUTHORIZATION`.
- [REQ-TRUST-0012](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0012) → `CC-TRUST-ISSUER-AUTHORIZATION`.

### CC-TRUST-TARGET-REGISTRATION

Public source: `library/task-trust.md`, section **Target Registration**. Intended target: `internal/admission/admit.go`. Planned case location: `internal/admission/cross_cutting_conformance_test.go`, `symbol:TestCcTrustTargetRegistration`.

Case plan: Mutate each safety-relevant tuple component, normalization collision and registration expiry; require exact typed local matching and recheck every resolved/redirect address before contact.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-TRUST-0013 -->
<!-- requirement-ref: REQ-TRUST-0014 -->
<!-- requirement-ref: REQ-TRUST-0015 -->
<!-- requirement-ref: REQ-TRUST-0016 -->

- [REQ-TRUST-0013](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0013) → `CC-TRUST-TARGET-REGISTRATION`.
- [REQ-TRUST-0014](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0014) → `CC-TRUST-TARGET-REGISTRATION`.
- [REQ-TRUST-0015](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0015) → `CC-TRUST-TARGET-REGISTRATION`.
- [REQ-TRUST-0016](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0016) → `CC-TRUST-TARGET-REGISTRATION`.

### CC-TRUST-ADMISSION-PROCEDURE

Public source: `library/task-trust.md`, section **Admission Procedure**. Intended target: `internal/admission/admit.go`. Planned case location: `internal/admission/cross_cutting_conformance_test.go`, `symbol:TestCcTrustAdmissionProcedure`.

Case plan: Instrument ordered admission with fake context/transport; independently fail each gate and verify atomic rejection before scheduling or network calls.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-TRUST-0017 -->
<!-- requirement-ref: REQ-TRUST-0018 -->

- [REQ-TRUST-0017](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0017) → `CC-TRUST-ADMISSION-PROCEDURE`.
- [REQ-TRUST-0018](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0018) → `CC-TRUST-ADMISSION-PROCEDURE`.

### CC-TRUST-TASK-IDENTIFIER-RULES

Public source: `library/task-trust.md`, section **Task Identifier Rules**. Intended target: `internal/admission/admit.go`. Planned case location: `internal/admission/cross_cutting_conformance_test.go`, `symbol:TestCcTrustTaskIdentifierRules`.

Case plan: Test exact issuer-scoped replay key, identifier syntax/length and campaign-wide identity; review no participant-specific derivation or targeting.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-TRUST-0019 -->
<!-- requirement-ref: REQ-TRUST-0020 -->

- [REQ-TRUST-0019](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0019) → `CC-TRUST-TASK-IDENTIFIER-RULES`.
- [REQ-TRUST-0020](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0020) → `CC-TRUST-TASK-IDENTIFIER-RULES`.

### CC-TRUST-REPLAY-DUPLICATE-AND-ROLLBACK-HANDLING

Public source: `library/task-trust.md`, section **Replay, Duplicate, and Rollback Handling**. Intended target: `internal/admission/state.go`. Planned case location: `internal/admission/cross_cutting_conformance_test.go`, `symbol:TestCcTrustReplayDuplicateAndRollbackHandling`.

Case plan: Exercise every replay decision with exact bytes/digests, conflicting reuse, rollback, expiry and unavailable state. Inspect minimal bounded persisted record and reset behavior.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-TRUST-0021 -->
<!-- requirement-ref: REQ-TRUST-0022 -->
<!-- requirement-ref: REQ-TRUST-0023 -->
<!-- requirement-ref: REQ-TRUST-0024 -->
<!-- requirement-ref: REQ-TRUST-0025 -->

- [REQ-TRUST-0021](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0021) → `CC-TRUST-REPLAY-DUPLICATE-AND-ROLLBACK-HANDLING`.
- [REQ-TRUST-0022](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0022) → `CC-TRUST-REPLAY-DUPLICATE-AND-ROLLBACK-HANDLING`.
- [REQ-TRUST-0023](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0023) → `CC-TRUST-REPLAY-DUPLICATE-AND-ROLLBACK-HANDLING`.
- [REQ-TRUST-0024](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0024) → `CC-TRUST-REPLAY-DUPLICATE-AND-ROLLBACK-HANDLING`.
- [REQ-TRUST-0025](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0025) → `CC-TRUST-REPLAY-DUPLICATE-AND-ROLLBACK-HANDLING`.

### CC-TRUST-TARGETED-DELIVERY

Public source: `library/task-trust.md`, section **Targeted Delivery**. Intended target: `internal/admission/admit.go`. Planned case location: `internal/admission/cross_cutting_conformance_test.go`, `symbol:TestCcTrustTargetedDelivery`.

Case plan: Attempt each prohibited recipient selector and interpolation path; review broad-group host policy and deny linkage-influenced task selection.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-TRUST-0026 -->
<!-- requirement-ref: REQ-TRUST-0027 -->
<!-- requirement-ref: REQ-TRUST-0028 -->

- [REQ-TRUST-0026](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0026) → `CC-TRUST-TARGETED-DELIVERY`.
- [REQ-TRUST-0027](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0027) → `CC-TRUST-TARGETED-DELIVERY`.
- [REQ-TRUST-0028](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0028) → `CC-TRUST-TARGETED-DELIVERY`.

### CC-TRUST-LOCAL-POLICY-LIFECYCLE

Public source: `library/task-trust.md`, section **Local Policy Lifecycle**. Intended target: `internal/admission/state.go`. Planned case location: `internal/admission/cross_cutting_conformance_test.go`, `symbol:TestCcTrustLocalPolicyLifecycle`.

Case plan: Fault-inject each authenticated generation update step; preserve complete old policy or uninitialized state, invalidate dependent caches and cancel forbidden work before reinitialized execution.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-TRUST-0029 -->
<!-- requirement-ref: REQ-TRUST-0030 -->
<!-- requirement-ref: REQ-TRUST-0031 -->
<!-- requirement-ref: REQ-TRUST-0032 -->

- [REQ-TRUST-0029](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0029) → `CC-TRUST-LOCAL-POLICY-LIFECYCLE`.
- [REQ-TRUST-0030](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0030) → `CC-TRUST-LOCAL-POLICY-LIFECYCLE`.
- [REQ-TRUST-0031](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0031) → `CC-TRUST-LOCAL-POLICY-LIFECYCLE`.
- [REQ-TRUST-0032](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0032) → `CC-TRUST-LOCAL-POLICY-LIFECYCLE`.

### CC-TRUST-ADMISSION-ERRORS

Public source: `library/task-trust.md`, section **Admission Errors**. Intended target: `internal/admission/admit.go`. Planned case location: `internal/admission/cross_cutting_conformance_test.go`, `symbol:TestCcTrustAdmissionErrors`.

Case plan: Trigger each error class and inject sensitive canaries/raw platform errors; return bounded deterministic operational diagnostics without Measurement Results.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-TRUST-0033 -->
<!-- requirement-ref: REQ-TRUST-0034 -->
<!-- requirement-ref: REQ-TRUST-0035 -->

- [REQ-TRUST-0033](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0033) → `CC-TRUST-ADMISSION-ERRORS`.
- [REQ-TRUST-0034](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0034) → `CC-TRUST-ADMISSION-ERRORS`.
- [REQ-TRUST-0035](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/library/task-trust.md#REQ-TRUST-0035) → `CC-TRUST-ADMISSION-ERRORS`.

### CC-JSON-UNIT-0001

Public source: `data-formats/json-decoding.md`, section **Unit 0001**. Intended target: `internal/jsondecode/decode.go`. Planned case location: `internal/jsondecode/cross_cutting_conformance_test.go`, `symbol:TestCcJsonUnit0001`.

Case plan: Accept exactly one schema-typed JSON value with permitted whitespace; reject empty, BOM, extra values/content, comments, trailing commas and non-object roots for object schemas.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-JSON-0001 -->

- [REQ-JSON-0001](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/json-decoding.md#REQ-JSON-0001) → `CC-JSON-UNIT-0001`.

### CC-JSON-UNIT-0002

Public source: `data-formats/json-decoding.md`, section **Unit 0002**. Intended target: `internal/jsondecode/decode.go`. Planned case location: `internal/jsondecode/cross_cutting_conformance_test.go`, `symbol:TestCcJsonUnit0002`.

Case plan: Test malformed UTF-8, escapes and unpaired surrogates; accept valid scalar pairs and preserve distinct non-normalized/case-sensitive strings and member names.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-JSON-0002 -->

- [REQ-JSON-0002](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/json-decoding.md#REQ-JSON-0002) → `CC-JSON-UNIT-0002`.

### CC-JSON-UNIT-0003

Public source: `data-formats/json-decoding.md`, section **Unit 0003**. Intended target: `internal/jsondecode/decode.go`. Planned case location: `internal/jsondecode/cross_cutting_conformance_test.go`, `symbol:TestCcJsonUnit0003`.

Case plan: Reject equal, conflicting and escaped-equivalent duplicate names at every nesting level; allow same names in distinct objects and distinct Unicode scalar sequences.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-JSON-0003 -->

- [REQ-JSON-0003](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/json-decoding.md#REQ-JSON-0003) → `CC-JSON-UNIT-0003`.

### CC-JSON-UNIT-0004

Public source: `data-formats/json-decoding.md`, section **Unit 0004**. Intended target: `internal/jsondecode/decode.go`. Planned case location: `internal/jsondecode/cross_cutting_conformance_test.go`, `symbol:TestCcJsonUnit0004`.

Case plan: Test exact escaped fixed names, case variants and unknown closed-schema fields while allowing only explicitly defined dynamic keys.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-JSON-0004 -->

- [REQ-JSON-0004](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/json-decoding.md#REQ-JSON-0004) → `CC-JSON-UNIT-0004`.

### CC-JSON-UNIT-0005

Public source: `data-formats/json-decoding.md`, section **Unit 0005**. Intended target: `internal/jsondecode/decode.go`. Planned case location: `internal/jsondecode/cross_cutting_conformance_test.go`, `symbol:TestCcJsonUnit0005`.

Case plan: Parameterize missing/null/empty/zero and every wrong JSON type; apply only explicitly allowed null/default behavior without coercion.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-JSON-0005 -->

- [REQ-JSON-0005](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/json-decoding.md#REQ-JSON-0005) → `CC-JSON-UNIT-0005`.

### CC-JSON-UNIT-0006

Public source: `data-formats/json-decoding.md`, section **Unit 0006**. Intended target: `internal/jsondecode/decode.go`. Planned case location: `internal/jsondecode/cross_cutting_conformance_test.go`, `symbol:TestCcJsonUnit0006`.

Case plan: Test every integer lexical form including -0, decimals, exponents, signs and leading zeroes; accept canonical integral tokens only and then apply domain rules.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-JSON-0006 -->

- [REQ-JSON-0006](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/json-decoding.md#REQ-JSON-0006) → `CC-JSON-UNIT-0006`.

### CC-JSON-UNIT-0007

Public source: `data-formats/json-decoding.md`, section **Unit 0007**. Intended target: `internal/jsondecode/decode.go`. Planned case location: `internal/jsondecode/cross_cutting_conformance_test.go`, `symbol:TestCcJsonUnit0007`.

Case plan: Test exact integers around floating-point and machine-width boundaries and finite representation endpoints; reject overflow/rounding/clamping and review full domain coverage.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-JSON-0007 -->

- [REQ-JSON-0007](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/json-decoding.md#REQ-JSON-0007) → `CC-JSON-UNIT-0007`.

### CC-JSON-UNIT-0008

Public source: `data-formats/json-decoding.md`, section **Unit 0008**. Intended target: `internal/jsondecode/decode.go`. Planned case location: `internal/jsondecode/cross_cutting_conformance_test.go`, `symbol:TestCcJsonUnit0008`.

Case plan: Vary required schema/test version fields and trusted versionless context; reject malformed/unsupported selection without trying another schema or injecting version fields.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-JSON-0008 -->

- [REQ-JSON-0008](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/json-decoding.md#REQ-JSON-0008) → `CC-JSON-UNIT-0008`.

### CC-JSON-UNIT-0009

Public source: `data-formats/json-decoding.md`, section **Unit 0009**. Intended target: `internal/jsondecode/decode.go`. Planned case location: `internal/jsondecode/cross_cutting_conformance_test.go`, `symbol:TestCcJsonUnit0009`.

Case plan: Exercise complete original-byte and nested byte/count envelopes including whitespace/unknown fields; unknown versions cannot expand limits and unbounded models remain disabled.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-JSON-0009 -->

- [REQ-JSON-0009](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/json-decoding.md#REQ-JSON-0009) → `CC-JSON-UNIT-0009`.

### CC-JSON-UNIT-0010

Public source: `data-formats/json-decoding.md`, section **Unit 0010**. Intended target: `internal/jsondecode/decode.go`. Planned case location: `internal/jsondecode/cross_cutting_conformance_test.go`, `symbol:TestCcJsonUnit0010`.

Case plan: Inject malformed/deep/large data under finite resource budgets; assert no usable partial model, admission, scheduling or I/O after errors and no trust claim from structural success.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-JSON-0010 -->

- [REQ-JSON-0010](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/json-decoding.md#REQ-JSON-0010) → `CC-JSON-UNIT-0010`.

### CC-JSON-UNIT-0011

Public source: `data-formats/json-decoding.md`, section **Unit 0011**. Intended target: `internal/jsondecode/decode.go`. Planned case location: `internal/jsondecode/cross_cutting_conformance_test.go`, `symbol:TestCcJsonUnit0011`.

Case plan: Inject sensitive canaries including names/targets/keys/tokens and raw parser/platform errors; assert bounded deterministic sanitized operational errors without Measurement Results.

Scope: Synthetic core contract cases. No execution evidence or support claim.

<!-- requirement-ref: REQ-JSON-0011 -->

- [REQ-JSON-0011](https://github.com/tv-rain/barometer-spec/blob/a944ed6c389fec9f0d089f0044ead6fe845328b1/data-formats/json-decoding.md#REQ-JSON-0011) → `CC-JSON-UNIT-0011`.

