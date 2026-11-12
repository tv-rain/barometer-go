# Measurement Interpretation Review

## Status

This document is the planned location of the INTERPRETATION human review cases of the nine measurements. Their reviewer roles, reviewed artifacts and checklists are in the [measurement case plans](measurement-requirements.md#planned-cases-and-requirement-dispositions); the table below records, for each case, what blocks the review.

No case has been reviewed, and every case stays `planned`. The Go side of each reviewed artifact exists: the result documentation and the [interpretation limits](measurement-interpretation.md) that ship with the implementation. The host and analyst guidance that each case also reviews does not exist yet, because no host application or deployment is in scope, and no implementation revision has been designated for review. A review case also cannot change status before the review-record format decision of [Typed Review Cases](traceability-manifest.md#typed-review-cases): the public format links only `conformance` evidence from a case, so a review record needs that decision first. When a review is performed, its record goes under a heading equal to the case ID, which is the case's manifest location.

Every case has the same reviewer role: a measurement interpretation reviewer, an analyst independent of the implementer.

| Case | Reviewed artifact | Blocker | Reviewer role |
|---|---|---|---|
| `RV-M-NET-01-INTERPRETATION` | Go result documentation and host and analyst guidance for NET-01 | No host or analyst guidance; no revision designated for review | Measurement interpretation reviewer |
| `RV-M-NET-03-INTERPRETATION` | Go result documentation and host and analyst guidance for NET-03 | No host or analyst guidance; no revision designated for review | Measurement interpretation reviewer |
| `RV-M-NET-05-INTERPRETATION` | Go result documentation and host and analyst guidance for NET-05 | No host or analyst guidance; no revision designated for review | Measurement interpretation reviewer |
| `RV-M-DNS-01-INTERPRETATION` | Go result documentation and host and analyst guidance for DNS-01 | No host or analyst guidance; no revision designated for review | Measurement interpretation reviewer |
| `RV-M-DNS-03-INTERPRETATION` | Go result documentation and host and analyst guidance for DNS-03 | No host or analyst guidance; no revision designated for review | Measurement interpretation reviewer |
| `RV-M-DNS-04-INTERPRETATION` | Go result documentation, host integration guidance and analyst guidance for DNS-04 | No host integration or analyst guidance; no revision designated for review | Measurement interpretation reviewer |
| `RV-M-TCP-01-INTERPRETATION` | Go result documentation and host and analyst guidance for TCP-01 | No host or analyst guidance; no revision designated for review | Measurement interpretation reviewer |
| `RV-M-TLS-01-INTERPRETATION` | Go result documentation and host and analyst guidance for TLS-01 | No host or analyst guidance; no revision designated for review | Measurement interpretation reviewer |
| `RV-M-HTTP-01-INTERPRETATION` | Go result documentation, host and analyst guidance and any derived-analysis description for HTTP-01 | No host or analyst guidance or derived-analysis description; no revision designated for review | Measurement interpretation reviewer |
