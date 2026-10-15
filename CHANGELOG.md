# Changelog

All notable changes to Barometer Go will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project follows [Semantic Versioning](https://semver.org/). A repository release version does not by itself establish implementation conformance, mobile-platform support, field validation, or support for a Draft Barometer specification.

## [Unreleased]

### Added

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

No release tag or implementation-support claim is created by these unreleased entries.
