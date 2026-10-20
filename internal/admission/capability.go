package admission

import (
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/registry"
)

// This build does not provide egress Network Context acquisition (Network
// paths and context, DEV-578) or NET-05 composite-step validation (DEV-454).
// A task that requires either is rejected as task_unsupported instead of being
// admitted and failing later (REQ-TRUST-0036). Each check is removed by the
// task that delivers its capability; neither depends on network state.

// requiresEgressContext is the capability part of step 6: a required egress
// field or a country allowlist (which implicitly requires country_code,
// REQ-EXECUTION-0023) needs egress acquisition. A context without either
// states no requirement and is admitted (REQ-EXECUTION-0024).
// requirement-ref: REQ-TRUST-0036
func requiresEgressContext(c task.Candidate) bool {
	ctx := c.Execution.Network.Context
	return len(ctx.RequiredFields) > 0 || len(ctx.AllowedCountryCodes) > 0
}

// requiresCompositeSteps is the composite part of step 8: every NET-05 step
// must be matched against the resolved registrations, which this build cannot
// do, so an explicitly enabled composite task is still rejected.
// requirement-ref: REQ-TRUST-0036
func requiresCompositeSteps(e registry.Entry) bool { return e.Composite }
