package admission

import "github.com/tv-rain/barometer-go/internal/registry"

// This build does not provide NET-05 composite-step validation (DEV-454). A
// task that requires it is rejected as task_unsupported instead of being
// admitted and failing later (REQ-TRUST-0036). The check is removed by the
// task that delivers the capability and does not depend on network state.
// Egress Network Context requirements are admitted since DEV-647 and
// evaluated at each attempt against the egress context of its own path.

// requiresCompositeSteps is the composite part of step 8: every NET-05 step
// must be matched against the resolved registrations, which this build cannot
// do, so an explicitly enabled composite task is still rejected.
// requirement-ref: REQ-TRUST-0036
func requiresCompositeSteps(e registry.Entry) bool { return e.Composite }
