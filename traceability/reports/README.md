# Traceability Reports

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
