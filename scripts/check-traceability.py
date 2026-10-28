"""Run the pinned public tooling and Go-local traceability CI policy offline."""

import argparse
from collections import Counter
import importlib.util
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import tempfile

SPEC_URL = "https://github.com/tv-rain/barometer-spec"
IMPL_URL = "https://github.com/tv-rain/barometer-go"
TOOLING_REVISION = "5d01879cfd1d47396c780befa2567f7e90439948"
# Current source references advance independently of historical evidence/tooling.
REFERENCE_REVISION = "f889a381c84a781176024f3fed3a3d4f5c92b241"
# Independent reviewed inventory; never infer required mappings from the input under test.
INVENTORY_REVISION = "c06ce6b7b73b8520f35ce10982315fcb4d05615b"
BASELINES = {
    "cross-cutting.json": ("f889a381c84a781176024f3fed3a3d4f5c92b241",
                           "249438ee1791cf3a06fa8da82274ac85a8e68199"),
    "measurements.json": ("f889a381c84a781176024f3fed3a3d4f5c92b241",
                          "249438ee1791cf3a06fa8da82274ac85a8e68199"),
}
# Earlier manifests are retained at the immutable commits where they were current, not as
# committed copies (decision 0012); git history and CI report artifacts hold them.
# Same standalone comment form as the public reference checker; history citations are excluded.
REFERENCE = re.compile(r"// requirement-ref: (\S+)")


def git(root, *args):
    return subprocess.check_output(["git", "--no-replace-objects", "-C", str(root), *args],
                                   env=dict(os.environ, GIT_NO_LAZY_FETCH="1"))


def retired_successors(catalog):
    """Return retired catalog IDs and their declared successors."""
    return {item["id"]: item["successors"] for item in catalog["requirements"] if item["status"] == "retired"}


def inventory_errors(manifests, expected, retired=None):
    """Partial evidence is permitted; removal of an allocated mapping is not.

    An allocated ID may leave a manifest only when it is retired at that manifest's
    public baseline and every declared successor keeps a target/case mapping.
    """
    retired = retired or {}
    errors = []
    if set(manifests) != set(expected):
        errors.append("Manifest inventory changed; review CI inventory and baseline policy.")
    for name in sorted(set(manifests) & set(expected)):
        required = {item["id"] for item in expected[name]["requirements"]}
        mapped = {item["id"] for item in manifests[name]["requirements"]
                  if item["implementation_locations"] and item["case_ids"]}
        actual = {item["id"] for item in manifests[name]["requirements"]}
        for identity in sorted(required - actual):
            successors = retired.get(name, {}).get(identity)
            if not successors:
                errors.append(f"{name}: missing required mapping {identity}")
            for successor in sorted(set(successors or ()) - mapped):
                errors.append(f"{name}: retired {identity} lacks successor mapping {successor}")
        # A retained ID with an empty target/case list is also a lost mapping.
        for item in manifests[name]["requirements"]:
            if item["id"] in required and (not item["implementation_locations"] or not item["case_ids"]):
                errors.append(f"{name}: missing target/case mapping {item['id']}")
    return errors


def go_references(root):
    """Return (path, line, ID) for every Go requirement-ref comment in the working tree."""
    listed = git(root, "ls-files", "-z", "--cached", "--others", "--exclude-standard", "--", "*.go")
    references = []
    for path in sorted({p for p in listed.decode().split("\0") if p}):
        if not (root / path).is_file():
            continue
        for number, line in enumerate((root / path).read_text(encoding="utf-8").splitlines(), 1):
            references.extend((path, number, match.group(1)) for match in REFERENCE.finditer(line))
    return references


def reference_findings(references, manifests, catalog, milestone_ids=None):
    """Return (errors, warnings) for Go requirement-ref comments.

    Unknown and retired IDs are always errors. Missing mappings for IDs already allocated
    at the pinned milestones are errors, preserving inventory protection. An active ID first
    allocated after those milestone catalogs may await the next update as an explicit warning
    (decision 0012), as may an unlinked source/test location. Without milestone metadata,
    missing mappings fail closed. Warnings establish neither mapping nor conformance.
    """
    status = {item["id"]: item["status"] for item in catalog["requirements"]}
    records = {}
    for name, manifest in sorted(manifests.items()):
        cases = {case["id"]: case["location"]["path"] for case in manifest["cases"]}
        for item in manifest["requirements"]:
            records.setdefault(item["id"], []).append(
                ({loc["path"] for loc in item["implementation_locations"]},
                 {cases[c] for c in item["case_ids"] if c in cases}))
    errors, warnings = [], []
    for path, number, identity in references:
        where = f"{path}:{number}: requirement-ref {identity}"
        if identity not in status:
            errors.append(f"{where} is unknown in the pinned public catalog")
        elif status[identity] != "active":
            errors.append(f"{where} is {status[identity]} in the pinned public catalog")
        elif identity not in records:
            if milestone_ids is not None and identity not in milestone_ids:
                warnings.append(f"{where} was allocated after the milestone catalogs; "
                                "mapping awaits the next milestone update")
            else:
                errors.append(f"{where} is absent from the current manifests")
        else:
            kind = 1 if path.endswith("_test.go") else 0
            if not any(path in linked[kind] for linked in records[identity]):
                target = "a case location" if kind else "an implementation location"
                warnings.append(f"{where} is not linked from the manifests as {target}")
    return errors, warnings


# DEV-572 case-plan specificity. The public case record has no type field, so a human review case is
# typed by its ID prefix, a fixed opening and labelled fields, and is never automated coverage.
REVIEW_PREFIX = "RV-"
REVIEW_OPENING = "Human review case; not automated coverage."
REVIEW_FIELDS = ("Reviewer role: ", "Reviewed artifact: ", "Checklist: ", "Evidence record: ")
# Passing unit-test cases keep the scope their evidence records match.
UNIT_TEST_PREFIX = "UT."
FAMILY_PREFIXES = ("Synthetic core contract cases; ",
                   "Native adapter and device release validation; planned outside portable Go runtime; ")
# A sentence or coordinated clause led by these verbs defers work to judgement rather than a test.
VAGUE_SENTENCE = re.compile(r"(?:^|[.;:]\s+|\b(?:and|or|then)\s+)(Review|Inspect|Plan)\b", re.IGNORECASE)
MINIMUM_PLAN = 100
REQUIREMENT_ID = re.compile(r"\bREQ-(?:[A-Z0-9]+-)+\d{4}\b")


def review_case_errors(case):
    """Return the ways a typed review case lacks its required structure."""
    scope, location = case["scope"], case["location"]
    errors = []
    if not scope.startswith(REVIEW_OPENING + " "):
        errors.append(f"must open with {REVIEW_OPENING!r}")
    positions = [scope.find(field) for field in REVIEW_FIELDS]
    for field, position in zip(REVIEW_FIELDS, positions):
        if position < 0:
            errors.append(f"lacks {field.strip()!r}")
    found = [position for position in positions if position >= 0]
    if found != sorted(found):
        errors.append("must list reviewer role, reviewed artifact, checklist and evidence record in that order")
    if not location["path"].endswith(".md") or location["selector"] != f"heading:{case['id']}":
        errors.append("must be located at its own Markdown heading")
    return errors


def case_plan_errors(manifests):
    """Reject vague or shared automated plans and untyped review work in the current manifests.

    A length floor and wording rules cannot prove a plan complete; completeness stays a review assertion.
    """
    errors = []
    for name, manifest in sorted(manifests.items()):
        linked = {}
        for item in manifest["requirements"]:
            for case_id in item["case_ids"]:
                linked.setdefault(case_id, set()).add(item["id"])
        scopes = Counter(case["scope"] for case in manifest["cases"])
        for case in manifest["cases"]:
            where = f"{name}: case {case['id']}"
            scope = case["scope"]
            if scopes[scope] > 1:
                errors.append(f"{where} shares its plan text with another case")
            if case["id"].startswith(UNIT_TEST_PREFIX):
                continue
            for identity in sorted(set(REQUIREMENT_ID.findall(scope)) - linked.get(case["id"], set())):
                errors.append(f"{where} names {identity} but is not mapped to it")
            if case["id"].startswith(REVIEW_PREFIX):
                errors.extend(f"{where} review case {problem}" for problem in review_case_errors(case))
                continue
            plan = scope
            for prefix in FAMILY_PREFIXES:
                plan = plan.removeprefix(prefix)
            match = VAGUE_SENTENCE.search(plan)
            if match:
                errors.append(f"{where} has a sentence led by {match.group(1)!r}; move review work to an "
                              f"{REVIEW_PREFIX} case")
            if len(plan) < MINIMUM_PLAN:
                errors.append(f"{where} plan has {len(plan)} characters; at least {MINIMUM_PLAN} are required")
    return errors


def case_counts(manifest):
    """Count automated, review and unit-test cases separately; review cases are not automated coverage."""
    kinds = Counter("review" if case["id"].startswith(REVIEW_PREFIX) else
                    "unit_test" if case["id"].startswith(UNIT_TEST_PREFIX) else "automated"
                    for case in manifest["cases"])
    return dict(sorted(kinds.items()))


def run(args):
    root = Path(__file__).resolve().parents[1]
    spec = args.spec_checkout.resolve()
    reports = args.report_dir.resolve()
    if reports == root or root in reports.parents or reports == spec or spec in reports.parents:
        raise ValueError("Report directory must be outside both source checkouts.")
    reports.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix="barometer-traceability-") as temporary:
        tools = Path(temporary) / "public"
        tools.mkdir()
        # Export immutable public tooling/source. Local checkout modifications cannot change it.
        archive = git(spec, "archive", "--format=tar", TOOLING_REVISION)
        subprocess.run(["tar", "-xf", "-", "-C", str(tools)], input=archive, check=True)
        references = Path(temporary) / "references"
        references.mkdir()
        # Export only reviewed source for reference scanning; execute tooling solely
        # from TOOLING_REVISION, never from the source checkout or this source pin.
        archive = git(spec, "archive", "--format=tar", REFERENCE_REVISION)
        subprocess.run(["tar", "-xf", "-", "-C", str(references)], input=archive, check=True)
        sys.path.insert(0, str(tools / "scripts"))
        import check_traceability as checker

        helper = Path(temporary) / "go-selector"
        subprocess.run(["go", "build", "-o", str(helper), str(tools / "scripts/go_selector.go")], check=True)
        commands = [
            [sys.executable, "-B", str(tools / "scripts/check_requirements.py"),
             "--spec-root", str(references), "--baseline", str(tools / "requirements/catalog.json"),
             "--consumer", str(root)],
            [sys.executable, "-B", "-m", "unittest", "discover", "-s", str(tools / "scripts"), "-p", "test_*.py", "-v"],
            [sys.executable, "-B", "-m", "unittest", "discover", "-s", str(root / "scripts"), "-p", "test_traceability_ci.py", "-v"],
            ["go", "test", "-count=1", "./internal/testkit/conformance"],
        ]
        failed = False
        for command in commands:
            failed |= subprocess.run(command, cwd=root).returncode != 0
        parsed = {}
        for name, (spec_revision, impl_revision) in BASELINES.items():
            path = root / "traceability" / name
            raw = path.read_bytes()
            report = checker.check(raw, {SPEC_URL: spec, IMPL_URL: root}, spec_revision,
                                   impl_revision, "historical", tools / "traceability/manifest.schema.json", helper)
            (reports / name).write_text(json.dumps(report, sort_keys=True, indent=2) + "\n", encoding="utf-8")
            counts = Counter(f"{f['level']}:{f['code']}" for f in report["findings"])
            print(f"{name}: valid={report['valid']}; findings={dict(sorted(counts.items()))}", flush=True)
            # Preserve every diagnostic, gap and review limit in the report; print errors in CI.
            for finding in report["findings"]:
                if finding["level"] == "error":
                    print(f"{finding['code']} {finding['pointer']} {finding['related_id']}", flush=True)
            failed |= not report["valid"]
            if report["valid"]:
                parsed[name] = checker.strict_json(raw)
        expected = {name: checker.strict_json(git(root, "show", f"{INVENTORY_REVISION}:traceability/{name}"))
                    for name in BASELINES}
        milestone_catalogs = {name: checker.strict_json(
            git(spec, "show", f"{spec_revision}:requirements/catalog.json"))
            for name, (spec_revision, _) in BASELINES.items()}
        retired = {name: retired_successors(catalog) for name, catalog in milestone_catalogs.items()}
        milestone_ids = {item["id"] for catalog in milestone_catalogs.values()
                         for item in catalog["requirements"]}
        actual_names = {p.relative_to(root / "traceability").as_posix()
                        for p in (root / "traceability").rglob("*.json")
                        if "reports" not in p.relative_to(root / "traceability").parts}
        if actual_names != set(BASELINES):
            print("Manifest inventory changed; review CI inventory and baseline policy.", flush=True)
            failed = True
        if len(parsed) == len(BASELINES):
            errors = inventory_errors(parsed, expected, retired)
            for error in errors:
                print(error, flush=True)
            failed |= bool(errors)
            # Reference comments are checked against the working-tree current manifests.
            errors, warnings = reference_findings(go_references(root), parsed,
                                                  checker.strict_json((references / "requirements/catalog.json").read_bytes()),
                                                  milestone_ids)
            for error in errors:
                print(error, flush=True)
            for warning in warnings:
                print(f"warning: {warning}", flush=True)
            print(f"Go requirement-ref consistency: {len(errors)} error(s), {len(warnings)} pending mapping/location "
                  "reference(s) awaiting the milestone manifest update.", flush=True)
            failed |= bool(errors)
            # Plan specificity applies to the current manifests.
            errors = case_plan_errors(parsed)
            for error in errors:
                print(error, flush=True)
            for name, manifest in sorted(parsed.items()):
                print(f"{name}: cases={case_counts(manifest)}", flush=True)
            print(f"Case-plan specificity: {len(errors)} error(s).", flush=True)
            failed |= bool(errors)
        (reports / "provenance.json").write_text(json.dumps({
            "tooling_revision": TOOLING_REVISION,
            "reference_revision": REFERENCE_REVISION,
            "manifest_baselines": BASELINES,
            "inventory_revision": INVENTORY_REVISION,
            "consumer_revision": git(root, "rev-parse", "HEAD").decode().strip(),
            "limits": "Current manifest metadata checks and working-tree references/synthetic tests; no runtime or support claim.",
        }, sort_keys=True, indent=2) + "\n", encoding="utf-8")
        print("Reports retain all permitted gaps and human review limits. No runtime/platform/field support established.", flush=True)
        return int(failed)


MINIMUM_PYTHON = (3, 10)
INSTALL_HINT = "Install the pinned public tooling dependencies as shown in docs/traceability-ci.md."


def version_tuple(text):
    """Parse a Go version such as 1.26.0, go1.26 or go1.26rc1; prereleases sort before the release."""
    match = re.fullmatch(r"(?:go)?(\d+)\.(\d+)(?:\.(\d+))?((?:rc|beta)\d+)?", text)
    if not match:
        return None
    major, minor, patch, prerelease = match.groups()
    return int(major), int(minor), -1 if prerelease else int(patch or 0)


def required_go(root):
    """Return the go directive from go.mod."""
    match = re.search(r"^go (\S+)$", (root / "go.mod").read_text(encoding="utf-8"), re.MULTILINE)
    if not match or not version_tuple(match.group(1)):
        raise ValueError("go.mod has no parseable go directive.")
    return match.group(1)


def go_toolchain_error(output, required):
    """Explain a `go version` result older than the module's go directive, or return None."""
    match = re.match(r"go version (go\S+) ", output)
    found = version_tuple(match.group(1)) if match else None
    if found is None:
        return "Cannot determine the Go version on PATH; `go version` printed unexpected output."
    if found < version_tuple(required):
        return (f"{match.group(1)} is first on PATH, but this module requires Go {required} or later. The public "
                "tooling builds with GOTOOLCHAIN=local, so automatic toolchain switching does not apply. Put a "
                f"Go {required}+ toolchain first on PATH.")
    return None


def preflight(root, run_go=None, python_version=None, has_module=None):
    """Return setup errors detectable before any traceability work starts."""
    errors = []
    if run_go is None:
        def run_go():
            # Run outside the module so an old toolchain reports its version instead of a go.mod error.
            return subprocess.run(["go", "version"], cwd=tempfile.gettempdir(), capture_output=True, text=True,
                                  env=dict(os.environ, GOTOOLCHAIN="local"), check=True).stdout
    try:
        error = go_toolchain_error(run_go(), required_go(root))
    except FileNotFoundError:
        error = "No `go` command found on PATH."
    except subprocess.CalledProcessError:
        error = "`go version` failed for the Go toolchain on PATH."
    if error:
        errors.append(error)
    version = python_version or sys.version_info[:2]
    if tuple(version) < MINIMUM_PYTHON:
        errors.append(f"Python {version[0]}.{version[1]} ({sys.executable}) is older than the required "
                      f"{MINIMUM_PYTHON[0]}.{MINIMUM_PYTHON[1]}. Select another interpreter with PYTHON=.")
    if has_module is None:
        def has_module(name):
            return importlib.util.find_spec(name) is not None
    if not has_module("jsonschema"):
        errors.append(f"Python {sys.executable} cannot import jsonschema. {INSTALL_HINT}")
    return errors


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--spec-checkout", type=Path, required=True)
    parser.add_argument("--report-dir", type=Path, required=True)
    args = parser.parse_args()
    errors = preflight(Path(__file__).resolve().parents[1])
    for error in errors:
        print(f"Traceability prerequisite missing: {error}", file=sys.stderr)
    if errors:
        return 2
    try:
        return run(args)
    except (OSError, ValueError, KeyError, ImportError, subprocess.CalledProcessError) as exc:
        print(f"Traceability setup/input failure: {type(exc).__name__}", file=sys.stderr)
        return 2


if __name__ == "__main__":
    sys.exit(main())
