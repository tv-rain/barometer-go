# Traceability Mappings

- [`measurements.json`](measurements.json) — all 393 allocated measurement units at the verified public PR #7 merge; 108 planned case families, with 14 units partially evidenced by DEV-569 unit-test cases at Go `c5f4e92`; no implemented requirement or support claim.
- [`reports/measurements.json`](reports/measurements.json) — unchanged historical DEV-473 report for the [preserved measurement mapping](historical/measurements-59d4936-4aafb0e.json).
- [Measurement requirement and case plan](../docs/measurement-requirements.md) — DEV-473 ownership, complete per-unit/table coverage expectations, reuse and native/host/analyst boundaries.
- [`cross-cutting.json`](cross-cutting.json) — shared-contract mapping of all 573 active shared requirements at the verified public PR #8 merge (DEV-568); 115 partially evidenced by DEV-569 unit-test cases at Go `c5f4e92`, the rest planned, with explicit gaps.
- [`reports/cross-cutting-a944ed6.json`](reports/cross-cutting-a944ed6.json) — unchanged DEV-568 report for the [preserved mapping](historical/cross-cutting-a944ed6-cc7e0a4.json) at Go `cc7e0a4`.
- [`evidence/go-unit-c5f4e92.md`](evidence/go-unit-c5f4e92.md) — DEV-569 executed unit-test evidence at Go `c5f4e92`: one section per passing `UT.*` case, with environment, command, the narrow requirement aspects exercised and limits.
- [`reports/cross-cutting-a944ed6-c5f4e92.json`](reports/cross-cutting-a944ed6-c5f4e92.json) and [`reports/measurements-59d4936-c5f4e92.json`](reports/measurements-59d4936-c5f4e92.json) — DEV-569 valid metadata reports for the current manifests: partial unit-test evidence, with all planned-family conformance gaps retained.
- [`historical/cross-cutting-a944ed6-cc7e0a4.json`](historical/cross-cutting-a944ed6-cc7e0a4.json) and [`historical/measurements-59d4936-4aafb0e.json`](historical/measurements-59d4936-4aafb0e.json) — byte-identical preserved DEV-568 cross-cutting and DEV-473 measurement mappings, valid only for their original baselines.
- [`historical/cross-cutting-47d7592.json`](historical/cross-cutting-47d7592.json) — byte-identical preserved DEV-471 mapping at the public PR #6 merge; valid only for its original baselines.
- [`reports/cross-cutting.json`](reports/cross-cutting.json) — unchanged historical report for that preserved mapping.
- [`reports/README.md`](reports/README.md) — tooling provenance, reproduction command and limitations.
- [Cross-cutting requirement and case plan](../docs/cross-cutting-requirements.md) — DEV-471 mapping rebaselined by DEV-568, detailed case families, ownership boundaries and baseline maintenance.
- [Manifest validation and claims](../docs/traceability-manifest.md) — canonical public format, checker invocation and evidence limitations.

This directory records development metadata. It establishes no implemented measurement, released API, device/platform support or field validation.
