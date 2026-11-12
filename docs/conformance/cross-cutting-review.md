# Cross-Cutting Review

## Status

This document is a planned location of the cross-cutting human review cases. Their reviewer roles, reviewed artifacts and checklists are in the [cross-cutting case plans](../cross-cutting-requirements.md); the table below records, for the result-linkage review cases and the three safety core review cases planned here, what blocks the review. The other review cases planned at this location are not recorded here yet.

No case has been reviewed, and every case stays `planned`. The artifacts these cases review do not exist yet: no host integration guidance, deployment configuration, deployment retention documentation, analyst guidance or Network Context endpoint operating documentation exists, because no host application, collector or deployment is in scope. A review case also cannot change status before the review-record format decision of [Typed Review Cases](../traceability-manifest.md#typed-review-cases): the public format links only `conformance` evidence from a case, so a review record needs that decision first. When a review is performed, its record goes under a heading equal to the case ID, which is the case's manifest location.

The three safety core cases (`RV-CC-LIMIT-CONTEXT-LINKAGE-AND-RETENTION-LIMITS`, `RV-CC-CONTEXT-CONTEXT-LAYERS-EGRESS-CONTEXT` and `RV-CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION`) do not gate a milestone; an independent external review may perform them once their artifacts exist.

| Case | Reviewed artifact | Blocker | Reviewer role |
|---|---|---|---|
| `RV-CC-LIMIT-CONTEXT-LINKAGE-AND-RETENTION-LIMITS` | Deployment retention documentation for the collector, host analytics, crash reporting and research storage | No deployment, collector or research storage, so no retention documentation | Deployment privacy reviewer, independent of the deployment operator |
| `RV-CC-CONTEXT-CONTEXT-LAYERS-EGRESS-CONTEXT` | Egress-classification endpoint operating documentation, Go result documentation and analyst guidance for egress context | No deployed endpoint or operating documentation; no analyst guidance | Privacy reviewer and measurement interpretation reviewer, independent of the implementer and endpoint operator |
| `RV-CC-PROXY-PRIVACY-SAFETY-AND-INTERPRETATION` | Go result documentation and host and analyst guidance for comparing direct and proxied paths | No host or analyst guidance; no revision designated for review | Measurement interpretation reviewer, an analyst independent of the implementer |
| `RV-CC-LINKAGE-OVERVIEW` | Go linkage API documentation and host integration guidance | No host integration guidance; no revision designated for review | Privacy documentation reviewer, independent of the implementer |
| `RV-CC-LINKAGE-INITIALIZATION-CONFIGURATION` | Deployment `result_linkage` configuration and its documented scope rationale | No deployment configuration | Deployment privacy reviewer, independent of the deployment operator |
| `RV-CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS` | Go linkage documentation, host integration guidance, deployment analysis and retention plan | No host integration guidance; no deployment analysis or retention plan | Privacy and safety reviewer, independent of the implementer and deployment operator |

The other result-linkage review cases are recorded at their own planned locations: `RV-CC-LINKAGE-GO-CORE-AND-MOBILE-BINDING-ARCHITECTURE` and `RV-CC-LINKAGE-APPLE-PLATFORM-STORAGE` in [Native Linkage Validation](native-linkage.md), and `RV-CC-LINKAGE-INGESTION-SEPARATION` in [Host Delivery Review](host-delivery-review.md).
