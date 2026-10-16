"""Regression tests for CI inventory loss; public tooling owns semantic claim tests."""

import copy
import importlib.util
from pathlib import Path
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


if __name__ == "__main__":
    unittest.main()
