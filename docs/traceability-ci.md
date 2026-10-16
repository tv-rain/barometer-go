# Traceability CI Checks

DEV-475 adds the `Traceability and conformance smoke` job to the existing CI workflow. It checks working-tree manifests and requirement-reference comments, runs the public checker's synthetic positive/negative suite and the Go conformance fixture smoke tests, and retains full JSON metadata reports as a CI artifact for 14 days. Compilation and synthetic cases establish no runtime, platform or field support.

## Reproduce Locally

Use Go 1.26.0 or later, Git, tar, and Python 3.10 or later with the public tooling's pinned packages. CI uses Python 3.14. Supply an explicit public specification checkout with full history containing commit `a944ed6c389fec9f0d089f0044ead6fe845328b1`; no private repository is required. The script never fetches, installs packages or discovers sibling checkouts. Dependency installation is a separate network operation.

From this repository, with the public checkout at the explicitly selected example location:

```sh
python3 -m venv /tmp/barometer-traceability-venv
git -C ../barometer-spec show a944ed6c389fec9f0d089f0044ead6fe845328b1:scripts/requirements.txt > /tmp/barometer-traceability-requirements.txt
/tmp/barometer-traceability-venv/bin/python -m pip install -r /tmp/barometer-traceability-requirements.txt
make traceability PYTHON=/tmp/barometer-traceability-venv/bin/python SPEC_CHECKOUT=../barometer-spec REPORT_DIR=/tmp/barometer-traceability-reports
make check
```

Both checkouts need the selected immutable commits, including Go inventory commit `c06ce6b7b73b8520f35ce10982315fcb4d05615b` and the manifests' implementation baselines. Reports must be outside both checkouts. The driver exports the public tooling/source snapshot from its pinned Git commit into temporary storage, builds the standard-library declaration helper there, and removes that storage on exit. Working-tree changes to the supplied public checkout cannot select executable tooling or schema. Python packages are development dependencies, with no Go module dependency or change to `make check` prerequisites.

## Checks and Failure Rules

- Validate public catalog integrity and Go working-tree `.md`/`.go` requirement-reference comments against the pinned public source snapshot.
- Run the public tooling tests, including broken references, false implementation/support claims, stale baselines, malformed metadata and the public invalid scenarios, using synthetic local Git repositories.
- Validate each working-tree manifest with the canonical public checker and independently selected specification/implementation commits. Existing paths, selectors, issue/evidence references and support gates are checked against those immutable objects.
- Compare each manifest's allocated IDs with its original mapping at the independent Go inventory commit. Removal of an allocated ID or its target/case mapping fails, even though partial-mode metadata alone permits omissions. Missing or unregistered top-level manifest files fail. New allocations and baseline changes require deliberate review of the driver's inventory and baseline policy.
- Run `go test -count=1 ./internal/testkit/conformance` against the current checkout. The ordinary Go CI job continues to run `make check` independently.

The driver exits nonzero for any failing command, invalid manifest, inventory loss or setup error. It writes complete checker reports, including every permitted gap and review assertion, and prints finding counts plus individual errors. CI uploads reports even when validation fails. No report is substituted for execution of a mapped case.

## Baselines and Evidence Limits

Public tooling/schema/reference scanning are pinned to `a944ed6c389fec9f0d089f0044ead6fe845328b1`. The working-tree cross-cutting manifest retains public baseline `47d75925b3829fccfb6954b545c15f695375b76f` and Go baseline `a9136f34fad00ecc2171cbc29e835c5c50b340df`; the measurement manifest retains public baseline `59d49364d85e48cc8e69c762d130a7a86dc1fe65` and Go baseline `4aafb0e94d5e13411d1f76521ece989a03aee617`. Reports label these checks historical. They validate preserved metadata against its declared objects; they do not promote old evidence into current implementation claims or combine differently based manifests into a single conformance report.

Planned targets/cases and declared unallocated sources remain permitted, visible gaps. The inventory gate prevents loss of previously allocated mappings; it does not infer untagged obligations or require unfinished implementation to pass. Changes to case meaning, semantic completeness, inventory completeness, evidence authenticity and issue closure/adoption still require human review. Adding an ID or preserving a nonempty case list cannot prove coverage. Future source allocations require reviewed pin/inventory updates; this job does not track a moving public branch.

Keep historical committed reports unchanged. New CI outputs include a provenance record with tooling, inventory and consumer revisions. During local development, the consumer revision names HEAD while manifests, reference comments and synthetic tests include uncommitted changes; such a run is local validation, not immutable committed evidence. No measurement, public API, native adapter, release or support status is changed by this task.
