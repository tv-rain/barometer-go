package admission

import (
	"slices"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/registry"
)

// maxTaskTargets is a finite implementation-local bound on the destinations
// one task may name, not a public maximum or an authority grant.
const maxTaskTargets = 64

// matchTargets is the target part of step 8. Every destination or name the
// selected test's rule reports for the normalized parameters must match a
// registration that is permitted by both the issuer and local policy, belongs
// to the exact test/version and role, carries an exactly equal tuple, has not
// expired at trusted now and permits the task's path. A redirect-role tuple
// must also be listed by a matched non-redirect registration. Matching compares
// normalized typed values only: there is no substring, suffix, wildcard or
// resolution rule. Any failure, including a rule that reports no measurement
// target or an incomplete tuple, is target_not_registered. The matched IDs
// narrow the carried registration IDs; DNS answers and redirects followed at
// execution must still be checked before contact (REQ-TRUST-0016).
// requirement-ref: REQ-TRUST-0013
// requirement-ref: REQ-TRUST-0014
// requirement-ref: REQ-TRUST-0015
func (a *Authority) matchTargets(policy task.LocalPolicy, rule registry.TargetRule, test task.TestVersion, params task.Parameters, now time.Time) task.AdmissionCode {
	targets, ok := rule.Targets(params)
	if !ok || len(targets) == 0 || len(targets) > maxTaskTargets {
		return task.TargetNotRegistered
	}
	var matched, redirects []string
	measured := false
	// Non-redirect roles first, so redirect permission comes only from
	// registrations this task actually matched.
	for _, t := range targets {
		if t.Role == task.RedirectTarget {
			continue
		}
		r, ok := a.match(policy, test, t, now)
		if !ok {
			return task.TargetNotRegistered
		}
		measured = measured || t.Role == task.MeasurementTarget
		matched = appendNew(matched, r.ID)
		for _, id := range r.RedirectIDs {
			redirects = appendNew(redirects, id)
		}
	}
	if !measured {
		return task.TargetNotRegistered
	}
	for _, t := range targets {
		if t.Role != task.RedirectTarget {
			continue
		}
		r, ok := a.match(policy, test, t, now)
		if !ok || !slices.Contains(redirects, r.ID) {
			return task.TargetNotRegistered
		}
		matched = appendNew(matched, r.ID)
	}
	// Carry only matched initial targets and explicitly linked redirect registrations
	// that were already permitted by both issuer and local policy. Traverse the
	// bounded registration graph so each actual hop still needs its parent's link.
	permitted := a.registrationIDs
	queue := slices.Clone(matched)
	for i := 0; i < len(queue) && i < 256; i++ {
		r, ok := policy.Target(queue[i])
		if !ok {
			continue
		}
		a.targets = append(a.targets, r)
		for _, id := range r.RedirectIDs {
			if slices.Contains(permitted, id) {
				queue = appendNew(queue, id)
			}
		}
	}
	a.registrationIDs = matched
	return 0
}

// match returns the first permitted registration, in policy order, that the
// target satisfies. Registration IDs are unique, so the result is
// deterministic for one policy snapshot.
func (a *Authority) match(policy task.LocalPolicy, test task.TestVersion, t registry.Target, now time.Time) (task.TargetRegistration, bool) {
	for _, id := range a.registrationIDs {
		r, ok := policy.Target(id)
		if !ok || r.Test != test || r.Role != t.Role || !r.ExpiresAt.AfterTime(now) || !pathPermitted(r.Paths, a.paths) || !sameTuple(r.Tuple, t.Tuple) {
			continue
		}
		return r, true
	}
	return task.TargetRegistration{}, false
}

// pathPermitted reports whether a registration permits the task's single
// path, to which authorizePath has already narrowed the effective paths.
func pathPermitted(registered, path task.Paths) bool {
	return path.Direct && registered.Direct || path.Proxy && registered.Proxy
}

// sameTuple is exact equality of every tuple component, including the order of
// bootstrap addresses, ALPN values and checks. An absent component matches
// only an absent component.
func sameTuple(a, b task.TargetTuple) bool { return a.Equal(b) }

func appendNew(ids []string, id string) []string {
	if slices.Contains(ids, id) {
		return ids
	}
	return append(ids, id)
}
