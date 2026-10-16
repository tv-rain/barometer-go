"""Regression tests for CI inventory loss and reference consistency; public tooling owns semantic claim tests."""

import copy
import importlib.util
from pathlib import Path
import subprocess
import tempfile
import unittest

spec = importlib.util.spec_from_file_location("ci_traceability", Path(__file__).with_name("check-traceability.py"))
ci = importlib.util.module_from_spec(spec)
spec.loader.exec_module(ci)


class InventoryTests(unittest.TestCase):
    def setUp(self):
        self.expected = {"a.json": {"requirements": [
            {"id": "REQ-TASK-0001", "implementation_locations": [{"state": "planned"}], "case_ids": ["planned-case"]},
            {"id": "REQ-TASK-0002", "implementation_locations": [{"state": "planned"}], "case_ids": ["other-case"]},
        ]}}
        self.actual = copy.deepcopy(self.expected)

    def test_explicit_planned_gaps_pass(self):
        self.assertEqual(ci.inventory_errors(self.actual, self.expected), [])

    def test_deleting_mapping_fails(self):
        self.actual["a.json"]["requirements"].pop()
        self.assertIn("a.json: missing required mapping REQ-TASK-0002", ci.inventory_errors(self.actual, self.expected))

    def test_same_count_replacement_cannot_hide_loss(self):
        self.actual["a.json"]["requirements"][0]["id"] = "REQ-TASK-0003"
        self.assertIn("a.json: missing required mapping REQ-TASK-0001", ci.inventory_errors(self.actual, self.expected))

    def test_empty_targets_or_cases_fail(self):
        for field in ("implementation_locations", "case_ids"):
            with self.subTest(field=field):
                actual = copy.deepcopy(self.actual)
                actual["a.json"]["requirements"][0][field] = []
                self.assertTrue(ci.inventory_errors(actual, self.expected))

    def test_retired_id_replaced_by_mapped_successor_passes(self):
        self.actual["a.json"]["requirements"][1]["id"] = "REQ-TASK-0003"
        retired = {"a.json": {"REQ-TASK-0002": ["REQ-TASK-0003"]}}
        self.assertEqual(ci.inventory_errors(self.actual, self.expected, retired), [])

    def test_retired_id_without_mapped_successor_fails(self):
        self.actual["a.json"]["requirements"].pop()
        retired = {"a.json": {"REQ-TASK-0002": ["REQ-TASK-0003"]}}
        self.assertIn("a.json: retired REQ-TASK-0002 lacks successor mapping REQ-TASK-0003",
                      ci.inventory_errors(self.actual, self.expected, retired))
        self.actual["a.json"]["requirements"].append(
            {"id": "REQ-TASK-0003", "implementation_locations": [], "case_ids": ["other-case"]})
        self.assertTrue(ci.inventory_errors(self.actual, self.expected, retired))

    def test_retirement_without_successor_or_in_other_manifest_fails(self):
        self.actual["a.json"]["requirements"].pop()
        for retired in ({"a.json": {"REQ-TASK-0002": []}}, {"b.json": {"REQ-TASK-0002": ["REQ-TASK-0001"]}}):
            with self.subTest(retired=retired):
                self.assertIn("a.json: missing required mapping REQ-TASK-0002",
                              ci.inventory_errors(self.actual, self.expected, retired))

    def test_retired_successors_reads_catalog_status(self):
        catalog = {"requirements": [
            {"id": "REQ-TASK-0001", "status": "active", "successors": []},
            {"id": "REQ-TASK-0002", "status": "retired", "successors": ["REQ-TASK-0003"]},
        ]}
        self.assertEqual(ci.retired_successors(catalog), {"REQ-TASK-0002": ["REQ-TASK-0003"]})

    def test_removed_or_unregistered_manifest_fails(self):
        self.assertTrue(ci.inventory_errors({}, self.expected))
        self.actual["new.json"] = {"requirements": []}
        self.assertTrue(ci.inventory_errors(self.actual, self.expected))



class ReferenceTests(unittest.TestCase):
    def setUp(self):
        self.catalog = {"requirements": [
            {"id": "REQ-TASK-0001", "status": "active", "successors": []},
            {"id": "REQ-TASK-0002", "status": "retired", "successors": ["REQ-TASK-0003"]},
            {"id": "REQ-TASK-0003", "status": "active", "successors": []},
            {"id": "REQ-TASK-0004", "status": "active", "successors": []},
        ]}
        self.manifests = {"a.json": {
            "requirements": [
                {"id": "REQ-TASK-0001", "implementation_locations": [{"path": "internal/a/a.go"}],
                 "case_ids": ["planned-case", "unit-case"]},
                {"id": "REQ-TASK-0003", "implementation_locations": [{"path": "internal/b/b.go"}],
                 "case_ids": ["planned-case"]},
            ],
            "cases": [
                {"id": "planned-case", "location": {"path": "internal/a/planned_test.go"}},
                {"id": "unit-case", "location": {"path": "internal/a/a_test.go"}},
            ],
        }}

    def errors(self, *references):
        return ci.reference_errors(list(references), self.manifests, self.catalog)

    def test_linked_source_and_test_references_pass(self):
        self.assertEqual(self.errors(("internal/a/a.go", 1, "REQ-TASK-0001"),
                                     ("internal/a/a_test.go", 2, "REQ-TASK-0001"),
                                     ("internal/b/b.go", 3, "REQ-TASK-0003")), [])

    def test_unknown_retired_or_unmapped_ids_fail(self):
        for identity, reason in (("REQ-TASK-9999", "unknown"), ("REQ-TASK-0002", "retired"),
                                 ("REQ-TASK-0004", "absent from the current manifests")):
            with self.subTest(identity=identity):
                errors = self.errors(("internal/a/a.go", 7, identity))
                self.assertEqual(len(errors), 1)
                self.assertIn(f"internal/a/a.go:7: requirement-ref {identity}", errors[0])
                self.assertIn(reason, errors[0])

    def test_unlinked_files_fail(self):
        for path, target in (("internal/c/c.go", "an implementation location"),
                             ("internal/c/c_test.go", "a case location"),
                             # A test file is not linked by an implementation location, and vice versa.
                             ("internal/a/a_test.go", "a case location")):
            with self.subTest(path=path):
                identity = "REQ-TASK-0003" if path == "internal/a/a_test.go" else "REQ-TASK-0001"
                errors = self.errors((path, 1, identity))
                self.assertEqual(errors, [f"{path}:1: requirement-ref {identity} is not linked from the manifests as {target}"])

    def test_case_of_another_requirement_does_not_link(self):
        self.manifests["a.json"]["requirements"][1]["case_ids"].append("unit-case")
        self.assertEqual(self.errors(("internal/a/a_test.go", 1, "REQ-TASK-0003")), [])
        self.manifests["a.json"]["requirements"][1]["case_ids"].remove("unit-case")
        self.assertTrue(self.errors(("internal/a/a_test.go", 1, "REQ-TASK-0003")))

    def test_go_references_scan_working_tree_comments_only(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            subprocess.run(["git", "init", "-q", str(root)], check=True)
            (root / "internal").mkdir()
            (root / "internal/a.go").write_text(
                "package a\n// requirement-ref: REQ-TASK-0001\n// requirement-history: REQ-TASK-0002\n", encoding="utf-8")
            (root / "internal/a_test.go").write_text("package a\n\n// requirement-ref: REQ-TASK-0004\n", encoding="utf-8")
            (root / "notes.md").write_text("<!-- requirement-ref: REQ-TASK-0004 -->\n", encoding="utf-8")
            self.assertEqual(ci.go_references(root), [("internal/a.go", 2, "REQ-TASK-0001"),
                                                      ("internal/a_test.go", 3, "REQ-TASK-0004")])


AUTOMATED_PLAN = ("Synthetic core contract cases; Exercise attempt byte budgets of 0, 1, the registered maximum and "
                  "maximum plus one across three sequential attempts with a scripted transport; assert each attempt "
                  "starts from the full budget and the over-limit attempt stops with the registered reason. "
                  "Requirements: REQ-TASK-0001.")
REVIEW_PLAN = ("Human review case; not automated coverage. Reviewer role: interpretation reviewer. Reviewed artifact: "
               "host guidance. Checklist: (1) no verdict wording. Evidence record: a review record for REQ-TASK-0001.")


class CasePlanTests(unittest.TestCase):
    def setUp(self):
        self.manifests = {"a.json": {
            "requirements": [{"id": "REQ-TASK-0001", "case_ids": ["CC-A", "RV-CC-A", "UT.a.TestA"]}],
            "cases": [
                {"id": "CC-A", "scope": AUTOMATED_PLAN, "location": {"path": "internal/a/a_test.go",
                                                                      "selector": "symbol:TestA"}},
                {"id": "RV-CC-A", "scope": REVIEW_PLAN, "location": {"path": "docs/review.md",
                                                                      "selector": "heading:RV-CC-A"}},
                # Unit-test scopes are bound to their evidence and exempt from the wording rules.
                {"id": "UT.a.TestA", "scope": "Short evidence scope.", "location": {"path": "internal/a/a_test.go",
                                                                                    "selector": "symbol:TestA"}},
            ],
        }}

    def case(self, identity):
        return next(case for case in self.manifests["a.json"]["cases"] if case["id"] == identity)

    def test_measurable_and_typed_review_plans_pass(self):
        self.assertEqual(ci.case_plan_errors(self.manifests), [])
        self.assertEqual(ci.case_counts(self.manifests["a.json"]), {"automated": 1, "review": 1, "unit_test": 1})

    def test_vague_lead_fails_unless_typed_review(self):
        for text in ("Review every row of the table and assert the registered reason for each row and boundary value.",
                     "Exercise rows; inspect retained fields for leakage of every synthetic canary in every result.",
                     "Exercise the boundaries and review interpretation of overhead versus application bytes later.",
                     "Plan device checks for every storage class and interruption point on each supported platform."):
            with self.subTest(text=text):
                self.case("CC-A")["scope"] = "Synthetic core contract cases; " + text * 2
                errors = ci.case_plan_errors(self.manifests)
                self.assertEqual(len(errors), 1)
                self.assertIn("move review work to an RV- case", errors[0])

    def test_short_plan_fails(self):
        self.case("CC-A")["scope"] = "Synthetic core contract cases; Reset budgets per attempt."
        [error] = ci.case_plan_errors(self.manifests)
        self.assertIn("at least 100 are required", error)

    def test_shared_plan_text_fails(self):
        self.manifests["a.json"]["cases"].append(dict(self.case("CC-A"), id="CC-B"))
        self.manifests["a.json"]["requirements"][0]["case_ids"].append("CC-B")
        errors = ci.case_plan_errors(self.manifests)
        self.assertEqual(len(errors), 2)
        self.assertTrue(all("shares its plan text" in error for error in errors))

    def test_named_requirement_must_be_mapped(self):
        self.case("CC-A")["scope"] = AUTOMATED_PLAN.replace("REQ-TASK-0001.", "REQ-TASK-0001, REQ-TASK-0002.")
        self.assertEqual(ci.case_plan_errors(self.manifests),
                         ["a.json: case CC-A names REQ-TASK-0002 but is not mapped to it"])

    def test_review_case_requires_structure(self):
        for change, problem in (
                ({"scope": REVIEW_PLAN.replace("Human review case; not automated coverage. ", "")}, "must open with"),
                ({"scope": REVIEW_PLAN.replace("Checklist: ", "Items: ")}, "lacks 'Checklist:'"),
                ({"scope": REVIEW_PLAN.replace("Reviewer role: interpretation reviewer. Reviewed artifact: host guidance.",
                                               "Reviewed artifact: host guidance. Reviewer role: interpretation reviewer.")},
                 "in that order"),
                ({"location": {"path": "internal/a/a_test.go", "selector": "symbol:TestA"}}, "own Markdown heading"),
                ({"location": {"path": "docs/review.md", "selector": "heading:Other"}}, "own Markdown heading")):
            with self.subTest(problem=problem):
                case = self.case("RV-CC-A")
                original = dict(case)
                case.update(change)
                errors = ci.case_plan_errors(self.manifests)
                case.clear()
                case.update(original)
                self.assertEqual(len(errors), 1)
                self.assertIn(problem, errors[0])

    def test_untyped_review_text_cannot_hide_behind_review_prefix_only(self):
        # An RV- ID with an automated-style plan is still rejected for missing review structure.
        self.case("RV-CC-A")["scope"] = AUTOMATED_PLAN
        self.assertTrue(ci.case_plan_errors(self.manifests))


class PreflightTests(unittest.TestCase):
    root = Path(__file__).resolve().parents[1]

    def errors(self, go_output="go version go1.26.0 linux/amd64", python=(3, 14), modules=("jsonschema",)):
        def run_go():
            if isinstance(go_output, BaseException):
                raise go_output
            return go_output
        return ci.preflight(self.root, run_go=run_go, python_version=python, has_module=lambda name: name in modules)

    def test_ready_environment(self):
        self.assertEqual(self.errors(), [])
        self.assertEqual(self.errors(go_output="go version go1.27.1 darwin/arm64"), [])

    def test_module_requirement_is_read_from_go_mod(self):
        self.assertEqual(ci.required_go(self.root), "1.26.0")

    def test_old_go_fails_fast(self):
        for output in ("go version go1.24.5 darwin/arm64", "go version go1.26rc1 linux/amd64", "go version go1.25 linux/amd64"):
            with self.subTest(output=output):
                [error] = self.errors(go_output=output)
                self.assertIn("requires Go 1.26.0 or later", error)
                self.assertIn("GOTOOLCHAIN=local", error)

    def test_missing_or_unreadable_go_fails_fast(self):
        for failure, text in ((FileNotFoundError(), "No `go` command"),
                              (subprocess.CalledProcessError(1, ["go"]), "`go version` failed"),
                              ("go version devel +abc linux/amd64", "Cannot determine")):
            with self.subTest(text=text):
                [error] = self.errors(go_output=failure)
                self.assertIn(text, error)

    def test_missing_jsonschema_fails_fast(self):
        [error] = self.errors(modules=())
        self.assertIn("cannot import jsonschema", error)
        self.assertIn("docs/traceability-ci.md", error)

    def test_old_python_fails_fast(self):
        [error] = self.errors(python=(3, 9))
        self.assertIn("older than the required 3.10", error)

    def test_all_errors_are_reported_together(self):
        self.assertEqual(len(self.errors(go_output="go version go1.24.0 linux/amd64", python=(3, 9), modules=())), 3)

    def test_version_parsing(self):
        self.assertEqual(ci.version_tuple("1.26.0"), (1, 26, 0))
        self.assertEqual(ci.version_tuple("go1.26"), (1, 26, 0))
        self.assertLess(ci.version_tuple("go1.26rc2"), ci.version_tuple("1.26.0"))
        self.assertIsNone(ci.version_tuple("devel"))


if __name__ == "__main__":
    unittest.main()
