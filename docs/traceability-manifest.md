# Traceability Manifest Adoption

## Authority and Ownership

Use the public [traceability manifest format](https://github.com/tv-rain/barometer-spec/blob/main/traceability/README.md) and [schema](https://github.com/tv-rain/barometer-spec/blob/main/traceability/manifest.schema.json) for future machine-readable requirement mappings. The public format owns field meanings, statuses, and validation rules; this repository owns Go mappings, cases, and sanitized evidence. Do not copy the schema here or make internal review records necessary for public verification.

The [Normative Source Inventory](normative-sources.md) remains the document-level contract inventory, [Requirement References](requirement-references.md) defines source/test citations, and [Package Architecture](package-architecture.md#specification-traceability) records intended ownership. None is a complete manifest or conformance report. Requirements without IDs remain authoritative and must be recorded as allocation/coverage gaps.

## Future Manifest Preparation

Place future production manifests under `traceability/` in this repository when their mapping task is approved. Each manifest selects one full immutable public specification commit/version and one full immutable Go commit. Resolve release tags to commit IDs. Preserve historical manifests when adopting new baselines rather than reusing old results as current evidence.

Map active public requirement IDs to Go file/declaration locations and locally identified conformance cases. Mark intended locations and cases as planned; they need not exist, but cannot count as implementation or passing evidence. Shared cases may cover multiple requirements only with explicit completeness review. Use partial coverage while IDs or mappings are incomplete, and list applicable public sources with unallocated obligations.

An `implemented` status requires existing behavior, all applicable mapped cases passing at the exact selected baselines, review evidence covering the whole obligation, and no unresolved affected issue. Existing source with incomplete evidence is `partial`. Record findings through [Specification Issues](specification-issues.md); public acceptance must be available and applicable adoption recorded before an issue can stop blocking a claim. Support records separately identify core, platform, or field scope and the additional evidence each requires. Pure-Go cross-compilation and native binding smoke checks do not establish device/platform support or field validation.

Evidence reports may be committed after the code they tested: the evidence reference has its own immutable revision, while its specification/implementation revision fields identify the actual tested commits. This avoids requiring a commit to contain its own final hash. Use public, sanitized evidence with explicit requirement IDs, case results, inputs/environment, and limitations. Never commit participant or measurement data, private configuration, credentials, sensitive logs, or personal paths.

## Validation and Current Limits

DEV-468 defines the format and examples only. This adoption creates no production manifest, cases, runtime behavior, or support claim. The public example is illustrative partial metadata with intentional gaps. DEV-469 owns the executable checker, automated validation scenarios, gap reports, and local/CI integration; no traceability-checker command is available or added to `make check` here.

Until the checker exists, review schema structure and the public semantic contract explicitly. Record the selected schema revision and review/check limits. JSON Schema alone cannot verify repository paths, evidence contents, completeness, issue resolution, or semantic conformance. Do not report a manifest as checker-validated merely because JSON parsing or schema validation succeeded.

When changing checked requirement comments, use the public requirement checker with explicitly supplied checkouts, then run `make check` and `git diff --check` as required by this repository. Future production manifests must follow the public strict decoding, immutable-baseline, reference, stale-path, claim, and coverage rules.

## Compatibility and Review Order

Merge coordination decision 0007, then the public format, then this Go guidance. The public `main` links become available after that merge; during review use the companion DEV-468 branch and identify its actual commit. No Go dependency, public API, runtime serialization, specification version, release tag, or support status changes. Future manifests reject unknown format versions and require an explicit mapping/evidence review when their public or Go baseline changes.
