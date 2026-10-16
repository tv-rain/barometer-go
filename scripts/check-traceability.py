"""Run the pinned public tooling and Go-local traceability CI policy offline."""

import argparse
from collections import Counter
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile

SPEC_URL = "https://github.com/tv-rain/barometer-spec"
IMPL_URL = "https://github.com/tv-rain/barometer-go"
TOOLING_REVISION = "a944ed6c389fec9f0d089f0044ead6fe845328b1"
# Independent reviewed inventory; never infer required mappings from the input under test.
INVENTORY_REVISION = "c06ce6b7b73b8520f35ce10982315fcb4d05615b"
BASELINES = {
    "cross-cutting.json": ("47d75925b3829fccfb6954b545c15f695375b76f",
                           "a9136f34fad00ecc2171cbc29e835c5c50b340df"),
    "measurements.json": ("59d49364d85e48cc8e69c762d130a7a86dc1fe65",
                          "4aafb0e94d5e13411d1f76521ece989a03aee617"),
}


def git(root, *args):
    return subprocess.check_output(["git", "--no-replace-objects", "-C", str(root), *args],
                                   env=dict(os.environ, GIT_NO_LAZY_FETCH="1"))


def inventory_errors(manifests, expected):
    """Partial evidence is permitted; removal of an allocated mapping is not."""
    errors = []
    if set(manifests) != set(expected):
        errors.append("Manifest inventory changed; review CI inventory and baseline policy.")
    for name in sorted(set(manifests) & set(expected)):
        required = {item["id"] for item in expected[name]["requirements"]}
        actual = {item["id"] for item in manifests[name]["requirements"]}
        for identity in sorted(required - actual):
            errors.append(f"{name}: missing required mapping {identity}")
        # A retained ID with an empty target/case list is also a lost mapping.
        for item in manifests[name]["requirements"]:
            if item["id"] in required and (not item["implementation_locations"] or not item["case_ids"]):
                errors.append(f"{name}: missing target/case mapping {item['id']}")
    return errors


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
        expected = {name: checker.strict_json(git(root, "show", f"{INVENTORY_REVISION}:traceability/{name}"))
                    for name in BASELINES}
        actual_names = {p.name for p in (root / "traceability").glob("*.json")}
        if actual_names != set(BASELINES):
            print("Manifest inventory changed; review CI inventory and baseline policy.", flush=True)
            failed = True
        if len(parsed) == len(BASELINES):
            errors = inventory_errors(parsed, expected)
            for error in errors:
                print(error, flush=True)
            failed |= bool(errors)
        (reports / "provenance.json").write_text(json.dumps({
            "tooling_revision": TOOLING_REVISION,
            "inventory_revision": INVENTORY_REVISION,
            "consumer_revision": git(root, "rev-parse", "HEAD").decode().strip(),
            "limits": "Historical metadata checks and current working-tree references/synthetic tests; no runtime or support claim.",
        }, sort_keys=True, indent=2) + "\n", encoding="utf-8")
        print("Reports retain all permitted gaps and human review limits. No runtime/platform/field support established.", flush=True)
        return int(failed)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--spec-checkout", type=Path, required=True)
    parser.add_argument("--report-dir", type=Path, required=True)
    args = parser.parse_args()
    try:
        return run(args)
    except (OSError, ValueError, KeyError, ImportError, subprocess.CalledProcessError) as exc:
        print(f"Traceability setup/input failure: {type(exc).__name__}", file=sys.stderr)
        return 2


if __name__ == "__main__":
    sys.exit(main())
