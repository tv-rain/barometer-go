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

    def test_removed_or_unregistered_manifest_fails(self):
        self.assertTrue(ci.inventory_errors({}, self.expected))
        self.actual["new.json"] = {"requirements": []}
        self.assertTrue(ci.inventory_errors(self.actual, self.expected))


if __name__ == "__main__":
    unittest.main()
