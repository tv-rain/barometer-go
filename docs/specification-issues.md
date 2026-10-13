# Specification Issues

## Scope and Baseline

This record holds Go implementation findings that need public specification clarification. It is not a local specification or an approval to choose behavior. Issue identifiers below belong to this Go findings record; they are not public requirement IDs, conformance-case IDs, or ClickUp task IDs.

Sources are pinned to public specification Draft `0.1.0`, revision `95dfeef1367f4c320bca7c7ae91b13ca7608b37f`, as recorded in the [Normative Source Inventory](normative-sources.md#reviewed-baseline). The source-discovery review found no missing contract file. This record is not an exhaustive semantic audit of all requirements.

## GO-SPEC-ISSUE-001: Duplicate JSON Member Names Outside Labels

**Status:** Open; public clarification required before implementation of the affected decoding behavior.

### Public Evidence

- [Task Description Format](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/data-formats/task.md#top-level-fields) defines required fields and rejects unknown top-level fields but does not state how repeated member names are handled.
- [Labels Format — Keys](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/data-formats/task-labels.md#keys) explicitly rejects duplicate keys even if a parser would retain only one value. That rule is stated for the labels object.
- [Task Trust — Admission Procedure](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/library/task-trust.md#admission-procedure) requires complete validation of the top-level schema and nested task objects before scheduling.
- [Task Trust — Replay, Duplicate, and Rollback Handling](https://github.com/tv-rain/barometer-spec/blob/95dfeef1367f4c320bca7c7ae91b13ca7608b37f/library/task-trust.md#replay-duplicate-and-rollback-handling) defines replay identity using a digest of the exact accepted bytes. This settles byte identity for redelivery, not repeated-name decoding semantics.

### Ambiguity and Impact

The reviewed sources do not explicitly extend the labels duplicate-key rule to the top-level task or other nested objects. They do not specify a selection/merge rule for those objects either. A synthetic task fragment such as `{"test_id":"NET-01","test_id":"TCP-01"}` therefore raises an unresolved admission question; it is not a valid complete task example.

This affects deterministic decoding and consistency between host validation and Go admission. The exact-byte replay digest does not resolve which value a decoder should validate. No behavior is implemented or selected by this issue record.

### Required Resolution

Public maintainers must clarify the scope and handling of duplicate member names in serialized tasks, including nested objects, and define whether differently escaped spellings of the same decoded name count as duplicates. Record the reviewed public revision and version decision here before adopting the behavior in Go. Any subsequent conformance cases belong to the decoding implementation task.

### Completion Evidence

Pending public clarification and later implementation/conformance evidence. The existing labels rejection rule remains authoritative while this broader question is open.
