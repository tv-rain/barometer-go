# Traceability Mappings

- [`cross-cutting.json`](cross-cutting.json) — shared-contract mapping of all 591 active shared requirements at public `5d01879` and Go `d788a80`: 308 partially evidenced by unit-test and conformance family cases, the rest planned, with explicit gaps. [Case plan](../docs/cross-cutting-requirements.md).
- [`measurements.json`](measurements.json) — all 394 allocated measurement units at public `5d01879` and Go `178bda0`: 16 of 108 case families passed for TCP-01 and DNS-01 and 68 units partially evidenced, the rest planned. [Case plan](../docs/measurement-requirements.md).
- [`evidence/`](evidence/) — executed test evidence reports. The current manifests reference [`go-unit-d788a80.md`](evidence/go-unit-d788a80.md) (cross-cutting) and [`go-unit-178bda0.md`](evidence/go-unit-178bda0.md) (measurements); the other reports are evidence for earlier manifests.

Manifests and evidence are updated at milestones, as described in [Manifest Validation](../docs/traceability-manifest.md). Earlier manifests and checker reports are retained in git history at the commits where they were current, not as committed copies; `git log -- traceability/` lists them. CI regenerates reports on every run and uploads them as artifacts ([Traceability CI](../docs/traceability-ci.md)).

This directory records development metadata. It establishes no implemented measurement, released API, device/platform support or field validation.
