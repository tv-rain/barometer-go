# Traceability Mappings

- [`measurements.json`](measurements.json) — all 393 allocated measurement units at the verified public PR #7 merge; 108 planned case families, without implementation or support claims.
- [`reports/measurements.json`](reports/measurements.json) — deterministic measurement metadata report retaining implementation/conformance, separately scoped shared-ID and inventory gaps.
- [Measurement requirement and case plan](../docs/measurement-requirements.md) — DEV-473 ownership, complete per-unit/table coverage expectations, reuse and native/host/analyst boundaries.
- [`cross-cutting.json`](cross-cutting.json) — shared-contract mapping of all 573 active shared requirements at the verified public PR #8 merge and explicit Go baseline (DEV-568); planned targets/cases and explicit gaps only.
- [`reports/cross-cutting-a944ed6.json`](reports/cross-cutting-a944ed6.json) — deterministic valid metadata report for that baseline, retaining planned implementation/conformance, registered-test and inventory gaps.
- [`historical/cross-cutting-47d7592.json`](historical/cross-cutting-47d7592.json) — byte-identical preserved DEV-471 mapping at the public PR #6 merge; valid only for its original baselines.
- [`reports/cross-cutting.json`](reports/cross-cutting.json) — unchanged historical report for that preserved mapping.
- [`reports/README.md`](reports/README.md) — tooling provenance, reproduction command and limitations.
- [Cross-cutting requirement and case plan](../docs/cross-cutting-requirements.md) — DEV-471 mapping rebaselined by DEV-568, detailed case families, ownership boundaries and baseline maintenance.
- [Manifest validation and claims](../docs/traceability-manifest.md) — canonical public format, checker invocation and evidence limitations.

This directory records development metadata. It establishes no implemented measurement, released API, device/platform support or field validation.
