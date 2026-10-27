package admission

import (
	"slices"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/task"
)

// matchComposite matches each primitive under its own exact test/version. Each
// step starts from the original authority intersection, so an earlier step
// cannot remove registrations needed by a later step. Nothing is published on
// failure; redirects remain linked to their own primitive registrations.
// requirement-ref: REQ-NET-05-0018
// requirement-ref: REQ-NET-05-0019
// requirement-ref: REQ-NET-05-0009
// requirement-ref: REQ-TRUST-0014
func (a *Admitter) matchComposite(authority *Authority, source task.SourceContext, params task.Parameters, now time.Time, baseline task.Risk) (task.Risk, task.AdmissionCode) {
	p, ok := params.(task.CompositeParameters)
	if !ok {
		return baseline, task.Unsupported
	}
	issuer, ok := a.policy.Issuer(source.IssuerID())
	if !ok {
		return baseline, task.IssuerNotAuthorized
	}
	var ids []string
	var targets []task.TargetRegistration
	for _, step := range p.CompositeSteps() {
		e, ok := a.registry.Lookup(step.Test)
		if !ok || !e.Step || e.Composite {
			return baseline, task.Unsupported
		}
		if !slices.Contains(issuer.Permissions.Tests, step.Test) || !slices.Contains(a.policy.Permissions().Tests, step.Test) {
			return baseline, task.IssuerNotAuthorized
		}
		perStep := *authority
		perStep.targets = nil
		if code := perStep.matchTargets(a.policy, e.Targets, step.Test, step.Parameters, now); code != 0 {
			return baseline, code
		}
		for _, id := range perStep.registrationIDs {
			ids = appendNew(ids, id)
		}
		for _, target := range perStep.targets {
			if !slices.ContainsFunc(targets, func(r task.TargetRegistration) bool { return r.ID == target.ID }) {
				targets = append(targets, target)
			}
		}
		baseline = restrictive(baseline, e.BaselineRisk)
	}
	authority.registrationIDs, authority.targets = ids, targets
	return baseline, 0
}
