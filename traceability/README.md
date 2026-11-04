# Traceability Mappings

- [`cross-cutting.json`](cross-cutting.json) — shared-contract mapping of all 599 active shared requirements at public `d537657` and Go `7376e2a`: 348 partially evidenced by unit-test and conformance family cases, the rest planned, with explicit gaps. [Case plan](../docs/cross-cutting-requirements.md).
- [`measurements.json`](measurements.json) — all 400 allocated measurement units at public `d537657` and Go `7376e2a`: 70 of 108 case families passed for NET-01, NET-03, NET-05, DNS-01, DNS-03, DNS-04, TCP-01, TLS-01 and HTTP-01 and 251 units partially evidenced, the rest planned. [Case plan](../docs/measurement-requirements.md).
- [`evidence/`](evidence/) — executed test evidence reports. Both current manifests reference [`go-unit-7376e2a.md`](evidence/go-unit-7376e2a.md); the other reports are evidence for earlier manifests.

Manifests and evidence are updated at milestones, as described in [Manifest Validation](../docs/traceability-manifest.md). Earlier manifests and checker reports are retained in git history at the commits where they were current, not as committed copies; `git log -- traceability/` lists them. CI regenerates reports on every run and uploads them as artifacts ([Traceability CI](../docs/traceability-ci.md)).

This directory records development metadata. It establishes no implemented measurement, released API, device/platform support or field validation.
