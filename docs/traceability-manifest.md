# Traceability Manifest Adoption

## Authority and Ownership

Use the public [traceability manifest format](https://github.com/tv-rain/barometer-spec/blob/main/traceability/README.md) and [schema](https://github.com/tv-rain/barometer-spec/blob/main/traceability/manifest.schema.json) for machine-readable requirement mappings. The public format owns field meanings, statuses, and validation rules; this repository owns Go mappings, cases, and sanitized evidence. Do not copy the schema here or make internal review records necessary for public verification.

The [Normative Source Inventory](normative-sources.md) remains the document-level contract inventory, [Requirement References](requirement-references.md) defines source/test citations, and [Package Architecture](package-architecture.md#specification-traceability) records intended ownership. None is a complete manifest or conformance report. Requirements without IDs remain authoritative and must be recorded as allocation/coverage gaps.

## Manifest Preparation and Maintenance

Place approved mapping manifests under `traceability/` in this repository. DEV-471 adds the [cross-cutting manifest](../traceability/cross-cutting.json) and [requirement/case plan](cross-cutting-requirements.md), pinned to the verified public PR #6 merge. Each manifest selects one full immutable public specification commit/version and one full immutable Go commit. Resolve release tags to commit IDs. Preserve historical manifests when adopting new baselines rather than reusing old results as current evidence.

DEV-473 adds the [measurement manifest](../traceability/measurements.json), [case plan](measurement-requirements.md) and [report](../traceability/reports/measurements.json), pinned to public PR #7 merge `59d49364d85e48cc8e69c762d130a7a86dc1fe65` and Go baseline `4aafb0e94d5e13411d1f76521ece989a03aee617`. All 393 allocated measurement units and 108 families are planned; no implementation, executed conformance or support is claimed. The 565 shared IDs remain separately mapped by DEV-471 at its original baselines. Shared public source text and existing catalog entries are unchanged through PR #7; this source continuity is not new conformance evidence. Reproduction and residual gaps are recorded with the [reports](../traceability/reports/README.md#measurement-report-dev-473).

Map active public requirement IDs to Go file/declaration locations and locally identified conformance cases. Mark intended locations and cases as planned; they need not exist, but cannot count as implementation or passing evidence. Shared cases may cover multiple requirements only with explicit completeness review. Use partial coverage while IDs or mappings are incomplete, and list applicable public sources with unallocated obligations.

An `implemented` status requires existing behavior, all applicable mapped cases passing at the exact selected baselines, review evidence covering the whole obligation, and no unresolved affected issue. Existing source with incomplete evidence is `partial`. Record findings through [Specification Issues](specification-issues.md); public acceptance must be available and applicable adoption recorded before an issue can stop blocking a claim. Support records separately identify core, platform, or field scope and the additional evidence each requires. Pure-Go cross-compilation and native binding smoke checks do not establish device/platform support or field validation.

Evidence reports may be committed after the code they tested: the evidence reference has its own immutable revision, while its specification/implementation revision fields identify the actual tested commits. This avoids requiring a commit to contain its own final hash. Use public, sanitized evidence with explicit requirement IDs, case results, inputs/environment, and limitations. Never commit participant or measurement data, private configuration, credentials, sensitive logs, or personal paths.

## Validation and Current Limits

DEV-468 defines the format and examples. DEV-469 provides the public [traceability checker and commands](https://github.com/tv-rain/barometer-spec/blob/main/scripts/README.md#traceability-checker), deterministic text/JSON reports, automated synthetic validation scenarios, and public tooling CI. DEV-471 now supplies a planned mapping of 565 shared requirements at immutable public/Go baselines. Its detailed case plans are not executed conformance cases or a new runtime/support claim. The public example remains illustrative partial metadata with intentional gaps.

Run the checker from a reviewed public tooling checkout, supplying each repository's canonical URL/local checkout mapping, explicit expected specification/implementation commits, current or historical mode, and the reviewed schema. Build the public standard-library Go declaration helper outside the checkout and provide it for existing Go symbol selectors. The checker reads immutable Git objects, does not discover siblings or fetch automatically, and records the schema digest. Preserve the tooling commit and report alongside the tested baselines. Additional public evidence repositories/revisions must be supplied explicitly; the reference revision may follow the code revision it describes.

Exit `0` means valid metadata with permitted gaps reported; `1` means invalid metadata or unjustified claims; `2` means invalid invocation/setup. Review assertions about evidence authenticity/content, semantic completeness, issue disposition/adoption, and inventory completeness remain visible and require human review. JSON/schema success alone cannot establish conformance or support. The checker does not execute the mapped cases. DEV-471 documents explicit baseline selection and report reproduction without adding Python, sibling repositories or public tooling to `make check`; public tooling CI validates the checker itself using synthetic repositories.

When changing checked requirement comments, use the public requirement checker with explicitly supplied checkouts, then run `make check` and `git diff --check` as required by this repository. Future production manifests must follow the public strict decoding, immutable-baseline, reference, stale-path, claim, and coverage rules.

## Compatibility and Review Order

The DEV-468 format and DEV-469 public tooling are adopted. DEV-471 consumes the verified public PR #6 merge, which contains the shared requirement allocations. That public prerequisite is merged before this Go mapping. The Python checker dependencies and Go parser helper belong to the public development tooling, not this runtime module. No Go dependency, public API, runtime serialization, specification version, release tag, or support status changes. Future manifests reject unknown format versions and require an explicit mapping/evidence review when their public or Go baseline changes.

## CI Integration

The separate [traceability CI job](traceability-ci.md) now runs pinned public tooling, manifest/reference checks and synthetic conformance smoke cases through `make traceability`. Its explicit Python/public-checkout prerequisites remain separate from `make check`. Historical manifests retain their baselines and visible gaps; CI adds no implementation or support claim.
