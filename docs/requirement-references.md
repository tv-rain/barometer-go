# Requirement References

## Authority and Scope

Use the public [requirement identifier convention](https://github.com/tv-rain/barometer-spec/blob/main/requirements/README.md) when referring to individual normative obligations in code, tests, documentation, and specification issues. Public specification maintainers allocate the IDs. This repository consumes them without defining alternative normative text or adding IDs to runtime task/result schemas.

An ID is stable identity, not a specification version or a support claim. The initial catalog is intentionally partial. Untagged public requirements remain authoritative. The [Normative Source Inventory](normative-sources.md) records the complete document-level baseline; the existing [package-level traceability map](package-architecture.md#specification-traceability) remains useful for locating surrounding requirements. Neither this document's examples nor the source inventory establish a complete requirement-to-case matrix or conformance.

## Reference Form

Use [Traceability Manifest Adoption](traceability-manifest.md) for future machine-readable mappings between these references, immutable baselines, cases/results, issues, and support evidence. The public manifest format complements reference comments; it does not turn a comment into an implementation claim.

Place one standalone comment near the implementing behavior or relevant test:

```go
// requirement-ref: REQ-TASK-0001
```

For checked documentation citations, use a standalone HTML comment outside fenced examples:

```markdown
<!-- requirement-ref: REQ-TASK-0001 -->
```

Also provide a navigable public link and state the applicable baseline for the traceability record. A completed implementation or conformance claim must record the public repository, specification version, immutable public commit or release tag, ID, and implementation/test evidence location. Do not infer a passing test or an implemented package from a reference comment. Historical citations use `requirement-history`, explain the historical context, and identify the immutable revision of the original obligation.

Keep a Go `requirement-ref` only where the code implements, or the test asserts, part of that requirement. The [traceability CI job](traceability-ci.md#checks-and-failure-rules) fails when a Go reference names an unknown or retired ID, an ID absent from the current manifests, or a file the manifests do not link to that ID. Add the declaration as an implementation location, or the test as a mapped case, in the same change. A link does not make the requirement implemented or the case passed.

## Representative Planned Use

These references target Draft `0.1.0` companion specification changes for DEV-466. They demonstrate lookup and intended ownership only; no implementation, runtime tests, or immutable conformance baseline are established here. The linked `main` URLs become available after the public specification branch is merged. During review, validate against the explicitly supplied companion checkout.

<!-- requirement-ref: REQ-TASK-0001 -->

[REQ-TASK-0001](https://github.com/tv-rain/barometer-spec/blob/main/data-formats/task.md#REQ-TASK-0001) belongs with the planned `internal/taskdecode` owner: rejection of unknown top-level task fields. A future case should verify rejection of an otherwise valid task containing an unknown top-level key.

<!-- requirement-history: REQ-TRUST-0001 -->
<!-- requirement-ref: REQ-TRUST-0043 -->

REQ-TRUST-0001 was retired at public revision `e97502747c917a16bc2b71337535ff972bbd7d64`; its successor [REQ-TRUST-0043](https://github.com/tv-rain/barometer-spec/blob/e97502747c917a16bc2b71337535ff972bbd7d64/library/task-trust.md#REQ-TRUST-0043) belongs with the `internal/admission` owner: refusal of remote tasks when signed package verification fails or cannot be performed, or when trusted policy, time or source context is unavailable. `CC-TRUST-SIGNED-REMOTE-PACKAGES` removes each input independently and verifies fail-closed admission (DEV-612).

<!-- requirement-ref: REQ-TCP-01-0001 -->

[REQ-TCP-01-0001](https://github.com/tv-rain/barometer-spec/blob/main/tests/tcp-01.md#REQ-TCP-01-0001) spans planned validation/admission and TCP-01 execution boundaries: invalid input prevents measurement execution and misleading results. Future synthetic cases should observe both the transport-call count and result boundary. This does not create any of those packages or establish their implementation.

## Local Reference Validation

For documentation or code changes containing checked references, use Python 3.9 or later and an explicitly selected public specification checkout containing the convention. From the coordinated workspace's `barometer-go` root:

```sh
python3 -B ../barometer-spec/scripts/check_requirements.py --consumer .
make check
git diff --check
```

For standalone work, supply the actual public checkout path instead of assuming a sibling directory. The checker is optional development tooling owned by the public specification; it is not a runtime dependency and does not add Python or sibling-checkout requirements to `make check`. Its diagnostics cover explicit reference comments and catalog integrity, not semantic conformance or arbitrary Markdown URLs. See the public [checker documentation](https://github.com/tv-rain/barometer-spec/blob/main/scripts/README.md) for catalog-history validation when changing allocations.

## Compatibility and Review

Merge the public convention and catalog before this consumer guidance. No Go API, module dependency, schema, or runtime behavior changes in this adoption. Recheck affected consumers when public source locations or lifecycle state change. Update current references to successor IDs where applicable; preserve historical references with their original immutable baseline. Specification ambiguities still follow the existing cross-repository review procedure.

## CI Integration

The separate [traceability CI job](traceability-ci.md) now runs pinned public tooling, manifest/reference checks and synthetic conformance smoke cases through `make traceability`. Its explicit Python/public-checkout prerequisites remain separate from `make check`. Historical manifests retain their baselines and visible gaps; CI adds no implementation or support claim.
