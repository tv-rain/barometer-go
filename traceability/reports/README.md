# Traceability Reports

- [`measurements.json`](measurements.json) — DEV-473 valid partial metadata: 393 missing implementation claims, 393 missing conformance records, 565 shared catalog IDs outside this manifest's scope, and 19 unallocated-inventory findings. No runtime conformance or support claim.
- [`cross-cutting.json`](cross-cutting.json) — valid partial metadata, not runtime conformance. It reports 565 missing implementation claims, 565 missing conformance records, one out-of-scope catalog mapping and 19 unallocated-inventory findings (18 source paths plus the general inventory-review finding).

## Provenance and Reproduction

Generated October 15, 2026 using the reviewed public checker and schema from `tv-rain/barometer-spec` commit `05c1364af4f9a604b617fca45f060949b557d265`. Its schema digest is recorded in the JSON report. The selected public source baseline is the PR #6 merge `47d75925b3829fccfb6954b545c15f695375b76f`. Checker and schema files remain identical to the tooling revision above; the public changes add annotations/catalog metadata only. Python 3.14 with the public tooling's pinned requirements was used. No mapped case was executed by this checker.

From the Go repository in the coordinated workspace, with the public tooling dependencies installed, supply exactly these two checkouts and selected immutable baselines:

```sh
python3 -B ../barometer-spec/scripts/check_traceability.py traceability/cross-cutting.json \
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
python3 -B ../barometer-spec/scripts/check_traceability.py traceability/measurements.json \
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
