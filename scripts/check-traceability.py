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
TOOLING_REVISION = "d93aff93cbc17e7c065fb40d8f6e2e01a697bf35"
# Independent reviewed inventory; never infer required mappings from the input under test.
INVENTORY_REVISION = "c06ce6b7b73b8520f35ce10982315fcb4d05615b"
BASELINES = {
    "cross-cutting.json": ("d93aff93cbc17e7c065fb40d8f6e2e01a697bf35",
                           "cb672923cdd7668324f6ac4802d6c60edca5572c"),
    "measurements.json": ("59d49364d85e48cc8e69c762d130a7a86dc1fe65",
                          "c5f4e92e06936cb677c7a03e9769ff75b1f51874"),
}
# Revision whose manifests DEV-569 preserved before linking unit-test evidence.
DEV568_REVISION = "c5f4e92e06936cb677c7a03e9769ff75b1f51874"
# Revision whose manifests DEV-572 preserved before making case plans measurable.
DEV569_REVISION = "60d64a606bdd19fa5d326c68a6157e7760b28ff1"
# Revision whose cross-cutting manifest DEV-582 preserved before rebaselining to public 5a37e82.
DEV572_REVISION = "c2b269042eaa5dff85027b46ccbea4dec9eb73a0"
# Revision whose cross-cutting manifest DEV-583 preserved before rewriting the admission case plans.
DEV582_REVISION = "a3e167c91c112e2da8b809be26fd76440cc74c95"
# Revision whose cross-cutting manifest DEV-597 preserved before rebaselining to public d93aff9.
DEV583_REVISION = "6aa8a744351b6faf3020de5433041f16e0b411f1"
# Revision whose cross-cutting manifest DEV-586 preserved before rebaselining to Go b25f107.
DEV597_REVISION = "112812735d3b6c20a5afd2276e5320ef31432114"
# Revision whose cross-cutting manifest DEV-587 preserved before rebaselining to Go aab88f5.
DEV586_REVISION = "e039022bd8318831bb628152d31e6a5b02a1b1f2"
# Revision whose cross-cutting manifest DEV-588 preserved before rebaselining to Go cb67292.
DEV587_REVISION = "922d2457453a55e6b45cc0627dd8939047e97d69"
# Superseded manifests: byte-identical to the original at their recorded source revision and
# checked at their original baselines.
HISTORICAL = {
    "historical/cross-cutting-47d7592.json": ("cross-cutting.json", INVENTORY_REVISION,
                                              "47d75925b3829fccfb6954b545c15f695375b76f",
                                              "a9136f34fad00ecc2171cbc29e835c5c50b340df"),
    "historical/cross-cutting-a944ed6-cc7e0a4.json": ("cross-cutting.json", DEV568_REVISION,
                                                      "a944ed6c389fec9f0d089f0044ead6fe845328b1",
                                                      "cc7e0a4c9a72e6f5d309af6faa94eaf457dafd65"),
    "historical/measurements-59d4936-4aafb0e.json": ("measurements.json", DEV568_REVISION,
                                                     "59d49364d85e48cc8e69c762d130a7a86dc1fe65",
                                                     "4aafb0e94d5e13411d1f76521ece989a03aee617"),
    "historical/cross-cutting-a944ed6-c5f4e92.json": ("cross-cutting.json", DEV569_REVISION,
                                                      "a944ed6c389fec9f0d089f0044ead6fe845328b1",
                                                      "c5f4e92e06936cb677c7a03e9769ff75b1f51874"),
    "historical/measurements-59d4936-c5f4e92.json": ("measurements.json", DEV569_REVISION,
                                                     "59d49364d85e48cc8e69c762d130a7a86dc1fe65",
                                                     "c5f4e92e06936cb677c7a03e9769ff75b1f51874"),
    "historical/cross-cutting-a944ed6-c5f4e92-dev572.json": ("cross-cutting.json", DEV572_REVISION,
                                                             "a944ed6c389fec9f0d089f0044ead6fe845328b1",
                                                             "c5f4e92e06936cb677c7a03e9769ff75b1f51874"),
    "historical/cross-cutting-5a37e82-4575079.json": ("cross-cutting.json", DEV582_REVISION,
                                                      "5a37e82865a08922c4821c7318656286933421e5",
                                                      "4575079b503467bff83c1175329be6cfa64cfb71"),
    "historical/cross-cutting-5a37e82-4575079-dev583.json": ("cross-cutting.json", DEV583_REVISION,
                                                             "5a37e82865a08922c4821c7318656286933421e5",
                                                             "4575079b503467bff83c1175329be6cfa64cfb71"),
    "historical/cross-cutting-d93aff9-c49a77f.json": ("cross-cutting.json", DEV597_REVISION,
                                                      "d93aff93cbc17e7c065fb40d8f6e2e01a697bf35",
                                                      "c49a77fc35d0190992c272ba7447467ab8ff675c"),
    "historical/cross-cutting-d93aff9-b25f107.json": ("cross-cutting.json", DEV586_REVISION,
                                                      "d93aff93cbc17e7c065fb40d8f6e2e01a697bf35",
                                                      "b25f107f393f531adb4a54d89b76364db4dda640"),
    "historical/cross-cutting-d93aff9-aab88f5.json": ("cross-cutting.json", DEV587_REVISION,
                                                      "d93aff93cbc17e7c065fb40d8f6e2e01a697bf35",
                                                      "aab88f59cbcf62208c9ebfa025fa795354dd072d"),
}
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


def reference_errors(references, manifests, catalog):
    """Every Go requirement-ref must name an active ID linked from a current manifest to its file.

    Source files must be an implementation location of that ID; test files must be the location
    of a case mapped to that ID. A comment alone never establishes implementation or evidence.
    """
    status = {item["id"]: item["status"] for item in catalog["requirements"]}
    records = {}
    for name, manifest in sorted(manifests.items()):
        cases = {case["id"]: case["location"]["path"] for case in manifest["cases"]}
        for item in manifest["requirements"]:
            records.setdefault(item["id"], []).append(
                ({loc["path"] for loc in item["implementation_locations"]},
                 {cases[c] for c in item["case_ids"] if c in cases}))
    errors = []
    for path, number, identity in references:
        where = f"{path}:{number}: requirement-ref {identity}"
        if identity not in status:
            errors.append(f"{where} is unknown in the pinned public catalog")
        elif status[identity] != "active":
            errors.append(f"{where} is {status[identity]} in the pinned public catalog")
        elif identity not in records:
            errors.append(f"{where} is absent from the current manifests")
        else:
            kind = 1 if path.endswith("_test.go") else 0
            if not any(path in linked[kind] for linked in records[identity]):
                target = "a case location" if kind else "an implementation location"
                errors.append(f"{where} is not linked from the manifests as {target}")
    return errors


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
        sys.path.insert(0, str(tools / "scripts"))
        import check_traceability as checker

        helper = Path(temporary) / "go-selector"
        subprocess.run(["go", "build", "-o", str(helper), str(tools / "scripts/go_selector.go")], check=True)
        commands = [
            [sys.executable, "-B", str(tools / "scripts/check_requirements.py"), "--consumer", str(root)],
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
        for name, (original, source, spec_revision, impl_revision) in HISTORICAL.items():
            raw = (root / "traceability" / name).read_bytes()
            # Preservation means no in-place rewrite of the reviewed original.
            if raw != git(root, "show", f"{source}:traceability/{original}"):
                print(f"{name}: differs from {original} at {source}.", flush=True)
                failed = True
            report = checker.check(raw, {SPEC_URL: spec, IMPL_URL: root}, spec_revision,
                                   impl_revision, "historical", tools / "traceability/manifest.schema.json", helper)
            (reports / name.replace("/", "-")).write_text(json.dumps(report, sort_keys=True, indent=2) + "\n",
                                                          encoding="utf-8")
            counts = Counter(f"{f['level']}:{f['code']}" for f in report["findings"])
            print(f"{name}: valid={report['valid']}; findings={dict(sorted(counts.items()))}", flush=True)
            for finding in report["findings"]:
                if finding["level"] == "error":
                    print(f"{finding['code']} {finding['pointer']} {finding['related_id']}", flush=True)
            failed |= not report["valid"]
        expected = {name: checker.strict_json(git(root, "show", f"{INVENTORY_REVISION}:traceability/{name}"))
                    for name in BASELINES}
        retired = {name: retired_successors(checker.strict_json(
                       git(spec, "show", f"{spec_revision}:requirements/catalog.json")))
                   for name, (spec_revision, _) in BASELINES.items()}
        actual_names = {p.relative_to(root / "traceability").as_posix()
                        for p in (root / "traceability").rglob("*.json")
                        if "reports" not in p.relative_to(root / "traceability").parts}
        if actual_names != set(BASELINES) | set(HISTORICAL):
            print("Manifest inventory changed; review CI inventory and baseline policy.", flush=True)
            failed = True
        if len(parsed) == len(BASELINES):
            errors = inventory_errors(parsed, expected, retired)
            for error in errors:
                print(error, flush=True)
            failed |= bool(errors)
            # Reference comments are checked against the working-tree current manifests.
            errors = reference_errors(go_references(root), parsed,
                                      checker.strict_json((tools / "requirements/catalog.json").read_bytes()))
            for error in errors:
                print(error, flush=True)
            print(f"Go requirement-ref consistency: {len(errors)} error(s).", flush=True)
            failed |= bool(errors)
            # Plan specificity applies to the current manifests; historical manifests stay unchanged.
            errors = case_plan_errors(parsed)
            for error in errors:
                print(error, flush=True)
            for name, manifest in sorted(parsed.items()):
                print(f"{name}: cases={case_counts(manifest)}", flush=True)
            print(f"Case-plan specificity: {len(errors)} error(s).", flush=True)
            failed |= bool(errors)
        (reports / "provenance.json").write_text(json.dumps({
            "tooling_revision": TOOLING_REVISION,
            "inventory_revision": INVENTORY_REVISION,
            "consumer_revision": git(root, "rev-parse", "HEAD").decode().strip(),
            "limits": "Historical metadata checks and current working-tree references/synthetic tests; no runtime or support claim.",
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
