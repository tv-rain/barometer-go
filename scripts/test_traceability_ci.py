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


if __name__ == "__main__":
    unittest.main()
