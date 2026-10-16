# Traceability Reports

- [`measurements.json`](measurements.json) — historical DEV-473 valid partial metadata for the [preserved measurement mapping](../historical/measurements-59d4936-4aafb0e.json): 393 missing implementation claims, 393 missing conformance records, 565 shared catalog IDs outside this manifest's scope, and 19 unallocated-inventory findings. No runtime conformance or support claim.
- [`cross-cutting-a944ed6.json`](cross-cutting-a944ed6.json) — historical DEV-568 valid partial metadata for the [preserved cross-cutting mapping](../historical/cross-cutting-a944ed6-cc7e0a4.json) at public baseline `a944ed6`: 573 missing implementation claims, 573 missing conformance records, 393 registered-test IDs outside this manifest's scope and 19 unallocated-inventory findings. No runtime conformance or support claim.
- [`cross-cutting-a944ed6-c5f4e92.json`](cross-cutting-a944ed6-c5f4e92.json) — DEV-569 valid partial metadata for the current cross-cutting manifest at public `a944ed6` and Go `c5f4e92`: 458 missing implementation claims, 573 missing conformance records, 393 registered-test IDs outside this manifest's scope and 19 unallocated-inventory findings. No runtime conformance or support claim.
- [`measurements-59d4936-c5f4e92.json`](measurements-59d4936-c5f4e92.json) — DEV-569 valid partial metadata for the current measurement manifest at public `59d4936` and Go `c5f4e92`: 379 missing implementation claims, 393 missing conformance records, 565 shared IDs outside this manifest's scope and 19 unallocated-inventory findings. No runtime conformance or support claim.
- [`cross-cutting-a944ed6-c5f4e92-dev572.json`](cross-cutting-a944ed6-c5f4e92-dev572.json) and [`measurements-59d4936-c5f4e92-dev572.json`](measurements-59d4936-c5f4e92-dev572.json) — DEV-572 valid partial metadata for the current manifests after the case-plan rewrite, at the same baselines and with the same gap counts as the DEV-569 reports, which now describe the [historical DEV-569 manifests](../README.md). No runtime conformance or support claim.
- [`cross-cutting.json`](cross-cutting.json) — historical DEV-471 report for the [preserved PR #6 mapping](../historical/cross-cutting-47d7592.json); valid partial metadata, not runtime conformance. It reports 565 missing implementation claims, 565 missing conformance records, one out-of-scope catalog mapping and 19 unallocated-inventory findings (18 source paths plus the general inventory-review finding).

## Provenance and Reproduction

Generated October 15, 2026 using the reviewed public checker and schema from `tv-rain/barometer-spec` commit `05c1364af4f9a604b617fca45f060949b557d265`. Its schema digest is recorded in the JSON report. The selected public source baseline is the PR #6 merge `47d75925b3829fccfb6954b545c15f695375b76f`. Checker and schema files remain identical to the tooling revision above; the public changes add annotations/catalog metadata only. Python 3.14 with the public tooling's pinned requirements was used. No mapped case was executed by this checker.

From the Go repository in the coordinated workspace, with the public tooling dependencies installed, supply exactly these two checkouts and selected immutable baselines:

```sh
python3 -B ../barometer-spec/scripts/check_traceability.py traceability/historical/cross-cutting-47d7592.json \
  --repository https://github.com/tv-rain/barometer-spec=../barometer-spec \
  --repository https://github.com/tv-rain/barometer-go=. \
  --expected-specification 47d75925b3829fccfb6954b545c15f695375b76f \
  --expected-implementation a9136f34fad00ecc2171cbc29e835c5c50b340df \
  --mode current \
  --schema ../barometer-spec/traceability/manifest.schema.json \
  --format json --output traceability/reports/cross-cutting.json
```

The report reads immutable source/catalog objects at the verified public merge and explicitly selected Go baseline; it does not inspect uncommitted runtime source or execute the planned cases. The [planning map](../../docs/cross-cutting-requirements.md) explains each family. All 565 shared IDs have targets/cases, while the one measurement-only catalog omission and residual unallocated source/inventory gaps remain visible. When adopting another source or implementation baseline, review affected mappings and regenerate the report with independently verified expected commits. Never treat an unchanged Draft version as an unchanged source.

Successful metadata validation permits reported gaps. Inventory completeness, case applicability, semantic coverage, issue disposition and evidence authenticity still require human review. This report establishes no runtime, platform, release or field support and does not resolve `GO-SPEC-ISSUE-001`.

## Measurement Report (DEV-473)

Generated October 15, 2026 with Python 3.14 and the public tooling's pinned dependencies. Checker and schema are selected from public merge `59d49364d85e48cc8e69c762d130a7a86dc1fe65`; their files are unchanged from the earlier reviewed tooling. The schema SHA-256 is `2b8c3ce0aeecd55f602534b8e1318addd2ffb83bbc5e013cb3720e7c40ac9c93`. The public prerequisite is the verified [PR #7 merge](https://github.com/tv-rain/barometer-spec/pull/7), whose tree matches approved head `9f0f457014881669cf58af5fad9735c423151fcc`. The Go baseline is `4aafb0e94d5e13411d1f76521ece989a03aee617`, preceding this planning-only metadata. All 108 case locations are planned, so no Go declaration helper or case execution is involved.

From the Go repository with the public checker's pinned dependencies available, supply these exact baselines and two checkouts:

```sh
python3 -B ../barometer-spec/scripts/check_traceability.py traceability/historical/measurements-59d4936-4aafb0e.json \
  --repository https://github.com/tv-rain/barometer-spec=../barometer-spec \
  --repository https://github.com/tv-rain/barometer-go=. \
  --expected-specification 59d49364d85e48cc8e69c762d130a7a86dc1fe65 \
  --expected-implementation 4aafb0e94d5e13411d1f76521ece989a03aee617 \
  --mode current \
  --schema ../barometer-spec/traceability/manifest.schema.json \
  --format json --output traceability/reports/measurements.json
```

The report is valid metadata with no errors or skipped checks. It retains 393 `MISSING_IMPLEMENTATION`, 393 `MISSING_CONFORMANCE`, 565 `MISSING_REQUIREMENT`, and 19 `UNALLOCATED_REQUIREMENTS` gaps. The 565 omitted IDs are the shared requirements already mapped separately by DEV-471 at its original baselines; the two manifests must not be summed as a single-baseline conformance report. All shared public source files and existing catalog records are unchanged through PR #7. The 19 inventory gaps cover 18 declared source paths plus unverified inventory completeness. Measurement annotation coverage does not close residual shared/threat-model/limitations allocation or semantic review.

The [measurement plan](../../docs/measurement-requirements.md) records every measurement unit's disposition, owner, planned case family and reuse boundary. A family must exercise every linked unit and complete schema/bound/outcome vector; one example is insufficient. Native, host, analyst and deployment review remain separately required where applicable. Support/evidence arrays are empty; the open downstream JSON issue is retained for test-specific schema, option and version work. The earlier cross-cutting manifest/report is preserved unchanged. A deterministic report and successful repository checks do not execute these planned cases, prove semantic completeness, close the issue, promote Draft status or establish platform/field support.

## Cross-cutting Report at Go cc7e0a4 (DEV-568)

Generated October 16, 2026 with Python 3.14, the public tooling's pinned dependencies and Go 1.26.0. Checker and schema are selected from public merge `a944ed6c389fec9f0d089f0044ead6fe845328b1`, which is also the selected source baseline: the verified [PR #8 merge](https://github.com/tv-rain/barometer-spec/pull/8), whose tree matches approved head `5869bc314c877e2ccd3426b82aadcd9c92414f02`. The schema SHA-256 is `2b8c3ce0aeecd55f602534b8e1318addd2ffb83bbc5e013cb3720e7c40ac9c93`. The Go baseline is `cc7e0a4c9a72e6f5d309af6faa94eaf457dafd65`, preceding this planning-only metadata. All case locations are planned, so no case is executed.

From the Go repository with the public checker's pinned dependencies available, supply these exact baselines and two checkouts:

```sh
python3 -B ../barometer-spec/scripts/check_traceability.py traceability/historical/cross-cutting-a944ed6-cc7e0a4.json \
  --repository https://github.com/tv-rain/barometer-spec=../barometer-spec \
  --repository https://github.com/tv-rain/barometer-go=. \
  --expected-specification a944ed6c389fec9f0d089f0044ead6fe845328b1 \
  --expected-implementation cc7e0a4c9a72e6f5d309af6faa94eaf457dafd65 \
  --mode current \
  --schema ../barometer-spec/traceability/manifest.schema.json \
  --format json --output traceability/reports/cross-cutting-a944ed6.json
```

The report is valid metadata with no errors or skipped checks. It retains 573 `MISSING_IMPLEMENTATION`, 573 `MISSING_CONFORMANCE`, 393 `MISSING_REQUIREMENT` and 19 `UNALLOCATED_REQUIREMENTS` gaps. The 393 omitted IDs are the registered test-specific units mapped separately by DEV-473 at an earlier baseline. Together the two manifests account for all 966 active catalog IDs at `a944ed6`, but they must not be summed as a single-baseline conformance report. Retired REQ-OBS-0011 and REQ-OBS-0021 are absent; their successors REQ-OBS-0042 and REQ-OBS-0043 are mapped, and the prior mapping is preserved in the [historical manifest](../historical/cross-cutting-47d7592.json). The historical cross-cutting report above remains unchanged and canonical only for its original baselines.

A deterministic report does not execute planned cases, prove semantic completeness, close GO-SPEC-ISSUE-001 or GO-SPEC-ISSUE-002, promote Draft status or establish platform/field support.

## Unit-Test Evidence Reports (DEV-569)

Generated October 16, 2026 with Python 3.14, the public tooling's pinned dependencies and Go 1.26.0. Checker, schema and Go declaration helper are selected from public merge `a944ed6c389fec9f0d089f0044ead6fe845328b1`; the schema SHA-256 is `2b8c3ce0aeecd55f602534b8e1318addd2ffb83bbc5e013cb3720e7c40ac9c93`. Both current manifests use Go baseline `c5f4e92e06936cb677c7a03e9769ff75b1f51874`, the merge of DEV-568. Their `UT.*` cases select top-level unit tests that passed there, and their conformance evidence references sections of the [unit-test evidence report](../evidence/go-unit-c5f4e92.md) in its later evidence commit. The previous manifests are preserved as [historical manifests](../README.md), and the reports above remain unchanged.

Existing symbol locations require the Go declaration helper. Build it outside both checkouts with a Go 1.26 toolchain first on `PATH`, then, from the Go repository with the public checker's pinned dependencies available:

```sh
GOTOOLCHAIN=local go build -o /tmp/barometer-go-selector ../barometer-spec/scripts/go_selector.go
python3 -B ../barometer-spec/scripts/check_traceability.py traceability/cross-cutting.json \
  --repository https://github.com/tv-rain/barometer-spec=../barometer-spec \
  --repository https://github.com/tv-rain/barometer-go=. \
  --expected-specification a944ed6c389fec9f0d089f0044ead6fe845328b1 \
  --expected-implementation c5f4e92e06936cb677c7a03e9769ff75b1f51874 \
  --mode current \
  --schema ../barometer-spec/traceability/manifest.schema.json \
  --go-selector /tmp/barometer-go-selector \
  --format json --output traceability/reports/cross-cutting-a944ed6-c5f4e92.json
python3 -B ../barometer-spec/scripts/check_traceability.py traceability/measurements.json \
  --repository https://github.com/tv-rain/barometer-spec=../barometer-spec \
  --repository https://github.com/tv-rain/barometer-go=. \
  --expected-specification 59d49364d85e48cc8e69c762d130a7a86dc1fe65 \
  --expected-implementation c5f4e92e06936cb677c7a03e9769ff75b1f51874 \
  --mode current \
  --schema ../barometer-spec/traceability/manifest.schema.json \
  --go-selector /tmp/barometer-go-selector \
  --format json --output traceability/reports/measurements-59d4936-c5f4e92.json
```

Both reports are valid metadata with no errors or skipped checks. The cross-cutting report retains 458 `MISSING_IMPLEMENTATION`, 573 `MISSING_CONFORMANCE`, 393 `MISSING_REQUIREMENT` and 19 `UNALLOCATED_REQUIREMENTS` gaps; the measurement report retains 379, 393, 565 and 19. The 115 cross-cutting and 14 measurement requirements with existing locations are `partial`. Every requirement keeps a missing-conformance gap because its planned family cases are mapped but not executed. The 58 cross-cutting and 5 measurement `UT.*` cases cover 60 distinct tests; the checker verifies their locations, evidence references and result/baseline consistency, while the content and sufficiency of each evidence section remain review assertions.

The unit tests exercise pure-Go models, decoders, validators and builders with synthetic fixtures on one host. This evidence does not execute a measurement, runtime producer, native binding, device or deployment, prove semantic completeness, close GO-SPEC-ISSUE-001 or GO-SPEC-ISSUE-002, promote Draft status or establish platform/field support. The two manifests keep different public baselines and must not be summed as a single-baseline conformance report.

## Case-Plan Specificity Reports (DEV-572)

Generated October 16, 2026 with Python 3.14, the public tooling's pinned dependencies and Go 1.26.0, using the same exported public checker, schema and Go declaration helper from `a944ed6c389fec9f0d089f0044ead6fe845328b1` (schema SHA-256 `2b8c3ce0aeecd55f602534b8e1318addd2ffb83bbc5e013cb3720e7c40ac9c93`). The public and Go baselines are unchanged from DEV-569, because the passing `UT.*` evidence is pinned to Go `c5f4e92e06936cb677c7a03e9769ff75b1f51874` and DEV-572 changes only case plans and planned review cases. Reproduce with the DEV-569 commands above, writing `traceability/reports/cross-cutting-a944ed6-c5f4e92-dev572.json` and `traceability/reports/measurements-59d4936-c5f4e92-dev572.json`.

Both reports are valid metadata with no errors or skipped checks, and they retain the DEV-569 gap counts: 458 `MISSING_IMPLEMENTATION`, 573 `MISSING_CONFORMANCE`, 393 `MISSING_REQUIREMENT` and 19 `UNALLOCATED_REQUIREMENTS` for cross-cutting, and 379, 393, 565 and 19 for measurements. The manifests now hold 189 automated, 25 review and 58 unit-test cross-cutting cases and 108 automated, 23 review and 5 unit-test measurement cases. Every `RV-*` case is planned and [is not automated coverage](../../docs/traceability-manifest.md#typed-review-cases). The CI driver's separate specificity rule, not this checker, rejects vague, shared or untyped plans.

Plan text cannot execute a case or prove its completeness. These reports establish no runtime, platform or field support and do not resolve GO-SPEC-ISSUE-001 or GO-SPEC-ISSUE-002.
