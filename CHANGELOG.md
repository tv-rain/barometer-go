# Changelog

All notable changes to Barometer Go will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project follows [Semantic Versioning](https://semver.org/). A repository release version does not by itself establish implementation conformance, mobile-platform support, field validation, or support for a Draft Barometer specification.

## [Unreleased]

### Changed

- Made every traceability case plan measurable or explicitly a human review case (DEV-572): templated measurement termination, minimization, version, interpretation and safety plans are replaced with per-measurement specifics; thin cross-cutting plans gain inputs, boundary values, expected outcomes and requirement IDs; judgement-based remainders become 48 typed `RV-*` review cases that are not automated coverage; and the CI traceability driver rejects vague, shared or untyped plans, with negative tests. Baselines, requirement statuses and allocations are unchanged, the DEV-569 manifests are preserved as historical, and no case is executed and no support claimed.

- Reconciled the GO-SPEC-ISSUE-002 disposition with its current state: the public vocabulary resolution merged at public `a944ed6`, while the finding stays open pending runtime and applicable binding/host adoption evidence. Documentation only; no runtime, traceability or support status changes.

- Linked existing Foundation code and unit tests as partial traceability evidence at Go baseline `c5f4e92`: 129 IDs cited by reviewed `requirement-ref` comments are `partial` with existing locations and passing `UT.*` unit-test cases backed by a committed evidence report, 12 comments that did not demonstrate their requirement were removed, prior manifests are preserved as historical, and CI now fails on unknown, retired, unmapped or unlinked Go references. Planned family cases stay unexecuted; no implemented requirement, runtime, platform or field support is claimed.

- Rebaselined the cross-cutting traceability mapping to public specification `a944ed6`: successors of retired REQ-OBS-0011/0021 and new observation-vocabulary requirements are mapped to planned case families, the prior mapping is preserved as a byte-identical historical manifest, and the CI inventory gate permits removal only of catalog-retired IDs whose successors are mapped. No implementation, conformance, historical report or support status changes.

### Added

- Open specification findings GO-SPEC-ISSUE-004 to GO-SPEC-ISSUE-007 (DEV-573): the `not_run`/`incomplete` boundary, the byte-limit accounting layer, NET-03 over-bound classification, and undefined `risk_factors`/`timezone` value domains, including that current task validation accepts an empty `risk_factors` array. Records only; no runtime, traceability or support change.

- Foundation validation-boundary documentation links delivered synthetic tooling to owned downstream runtime harness tasks; current specification-adoption summaries distinguish merged Foundation code from pending runtime/binding evidence. No runtime behavior, historical traceability report or support status changes.

- Executable sanitized Foundation examples for configuration preparation, strict task decoding, cancellation-state handling and result serialization, plus scripted test-only intended client lifecycle and clearing calls. Existing CI validation runs their output assertions; runtime lifecycle and clearing remain pending.

- Pinned public traceability tooling in CI, independent mapping-loss checks, requirement-reference validation and synthetic conformance smoke cases, with reproducible local commands and full gap reports; no runtime or support promotion.

- Foundation model validation-order and normalization documentation, with combined-error precedence, canonical round-trip and typed-reconstruction regressions across model owners.

- Internal bounded Result Bundle production from validated Measurement Results, random single-use logical IDs, immutable exact retry bytes and trusted-initialization-gated rotating-linkage projections, with synthetic privacy, size-boundary and fuzz tests. Derivation, epoch lifecycle, transport, persistence, retention and clearing remain pending.

- Internal Measurement Result models and strict bounded serialization with independent envelope/test/result versions, exact local task/context binding, initial output schemas, NET-05 partial-step validation and synthetic boundary/fuzz tests. Runtime execution, bundling, persistence, delivery and platform support remain pending.

- Internal bounded DNS/TCP/TLS/HTTP observation models and immutable minimized Network Context snapshots with strict decoding, path/generation binding and synthetic privacy/boundary tests. Closed public vocabularies, typed producer normalization and reviewed-query answer retention are implemented; acquisition, execution and result integration remain pending.

- Internal finite admission errors, pure attempt lifecycle transitions and complete execution-reason models, with first-condition preservation, bounded retry classification and synthetic transition/boundary tests. Runtime admission, orchestration, budget enforcement and result serialization remain pending.

- Versioned Go-local conformance fixtures and a deterministic test-only runner for common task decoding, with exact-byte synthetic inputs, separate trusted-context/local-policy profiles, strict wrapper validation and positive/negative cases. Full admission, policy enforcement and measurement conformance remain pending.

- Internal trusted-source context and exact-byte delivery binding, immutable local/issuer/registered-target policy models, bounded revocation inputs and synthetic boundary tests. Trusted models reject JSON import/export; authentication, admission, replay, policy activation and revocation enforcement remain pending.
- Internal immutable initialization/configuration models for proxy, Network Context endpoints and optional linkage, with strict bounded decoding, separate deployment approvals, atomic proxy value replacement and clock/protected-key capability preparation. Client activation, native adapters and runtime admission remain pending.
- Common internal task and schedule models with exact test-schema selection, deterministic defaults, strict common-field/limit validation, trusted-time expiry checks, bounded retry/repetition eligibility, and synthetic boundary/fuzz tests. Candidates do not establish complete admission or measurement support.
- Shared internal strict JSON decoding with explicit schema/version selection, exact bounded integers, Unicode and duplicate-name checks, finite resource envelopes, sanitized atomic errors, and synthetic negative/boundary/fuzz tests. Individual public model schemas and runtime admission remain pending.

- Internal test-only negative/boundary fixtures, requirement-aware rejection assertions, and deterministic clock/HTTP transport controls, with synthetic reuse examples; no runtime decoding or measurement implementation.
- Defined public Go facade signatures, trusted admission boundary, lifecycle and result ownership, with isolated native binding fixture validation; no production API or runtime implementation.
- Adoption guidance for the public traceability-manifest format and evidence/status gates; no production mapping, checker, or implementation-support claim.
- Contribution guidance for public specification-issue review and ordered adoption, with expanded evidence for the unresolved duplicate-member finding; no decoding behavior is selected.
- Document-level inventory of all 27 Draft `0.1.0` public runtime contracts at a pinned specification revision, authority classifications, and an unresolved task-decoding specification issue; no implementation or conformance claim.
- Android and Apple gomobile compilation smoke checks with pinned tooling, a build-tagged synthetic boundary, and incompatible-export/dependency regression checks; these do not establish mobile platform support.
- Guidance for stable public requirement references and specification-baseline evidence in documentation, code, and tests.
- Initial Go module and repository foundation.
- Repository guidance, canonical validation commands, continuous integration, race testing, and repository-integrity checks.
- MIT licensing for code and documentation, with separate trademark terms for the Barometer name and brand assets.
- Versioning, compatibility, changelog, tagging, artifact-naming, and experimental-release conventions.

### Fixed

- Common candidate validation now rejects duplicate, excessive or invalid UTF-8 labels and normalizes a copied label slice consistently with decoded tasks. Network Context rejects non-ASCII country/region spellings before case conversion can map them onto reviewed ASCII codes. Zero-value result decoders now reject with a binding error before processing input or producer records.

No release tag or implementation-support claim is created by these unreleased entries.
