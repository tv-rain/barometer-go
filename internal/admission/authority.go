package admission

import (
	"slices"
	"strings"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/task"
)

// Authority is the effective authority of one gated task: registered behavior
// intersected with issuer authorization, deployment policy and the task's own
// restrictions (REQ-TRUST-0010). Every bound only narrows; nothing is a union,
// an implicit default or a clamped value. It is unpublished, holds no task
// bytes and has no serialized form.
type Authority struct {
	generation uint64
	paths      task.Paths
	// Registration IDs permitted by both the issuer and local policy, narrowed
	// by target matching to those the task's destinations matched.
	registrationIDs []string
	// Ceilings after intersection, narrowed to the task's declared values.
	limits                                    task.PolicyLimits
	maxRisk                                   task.Risk
	allowMedium, allowUnknown, allowComposite bool
	requireConsent                            bool
}

func (Authority) String() string   { return "admission.Authority" }
func (Authority) GoString() string { return "admission.Authority" }

// resolve runs step 7 against the active policy snapshot: policy presence,
// source generation, issuer, exact test/version and composite enablement, then
// the task's numeric values and labels against the intersected ceilings.
// requirement-ref: REQ-TRUST-0010
// requirement-ref: REQ-TRUST-0011
// requirement-ref: REQ-TRUST-0012
// requirement-ref: REQ-LIMIT-0030
func resolve(policy task.LocalPolicy, source task.SourceContext, test task.TestVersion, c task.Candidate, base time.Time) (Authority, task.AdmissionCode) {
	if !policy.Valid() {
		return Authority{}, task.PolicyUnavailable
	}
	switch g := source.PolicyGeneration(); {
	case g < policy.Generation():
		return Authority{}, task.PolicyRollback
	case g > policy.Generation():
		return Authority{}, task.PolicyUnavailable
	}
	issuer, ok := policy.Issuer(source.IssuerID())
	if !ok {
		return Authority{}, task.IssuerNotAuthorized
	}
	is, local := issuer.Permissions, policy.Permissions()
	if !slices.Contains(is.Tests, test) || !slices.Contains(local.Tests, test) {
		return Authority{}, task.IssuerNotAuthorized
	}
	a := Authority{
		generation:      policy.Generation(),
		paths:           task.Paths{Direct: is.Paths.Direct && local.Paths.Direct, Proxy: is.Paths.Proxy && local.Paths.Proxy},
		registrationIDs: intersect(is.RegistrationIDs, local.RegistrationIDs),
		limits:          intersectLimits(is.Limits, local.Limits),
		maxRisk:         min(is.MaxRisk, local.MaxRisk),
		allowMedium:     is.AllowMediumRisk && local.AllowMediumRisk,
		allowUnknown:    is.AllowUnknownRisk && local.AllowUnknownRisk,
		allowComposite:  is.AllowComposite && local.AllowComposite,
		requireConsent:  is.RequireConsent || local.RequireConsent || c.Safety.RequiresConsent,
	}
	if test.ID == "NET-05" && !a.allowComposite {
		return Authority{}, task.IssuerNotAuthorized
	}
	if a.limits, ok = narrow(a.limits, c, base); !ok {
		return Authority{}, task.LimitExceeded
	}
	for _, l := range c.Labels {
		if !namespaced(l.Key, is.LabelNamespaces) || !namespaced(l.Key, local.LabelNamespaces) {
			return Authority{}, task.IssuerNotAuthorized
		}
	}
	return a, 0
}

// authorizePath is the authority part of step 8: the task's single network
// path must be permitted by both issuer and local policy (REQ-TRUST-0037).
// The matched registration's paths are checked by matchTargets; registry
// proxy compatibility stays with later step-8 work.
func (a *Authority) authorizePath(path string) task.AdmissionCode {
	switch {
	case path == "direct" && a.paths.Direct:
		a.paths = task.Paths{Direct: true}
	case path == "proxy" && a.paths.Proxy:
		a.paths = task.Paths{Proxy: true}
	default:
		return task.IssuerNotAuthorized
	}
	return 0
}

// authorizeRisk applies the intersected risk bounds to the risk already known
// at admission: the most restrictive of the registered baseline and the
// task-declared level. Unknown is unordered and needs both explicit opt-ins;
// medium and higher need both medium-risk enablements. Locally determined risk
// and consent evaluation remain DEV-589.
// requirement-ref: REQ-TRUST-0011
func (a Authority) authorizeRisk(baseline task.Risk, declared string) task.AdmissionCode {
	risk := baseline
	switch declared {
	case "medium":
		risk = max(risk, task.MediumRisk)
	case "high":
		risk = max(risk, task.HighRisk)
	case "unknown":
		risk = task.UnknownRisk
	}
	switch {
	case risk == task.UnknownRisk:
		if !a.allowUnknown {
			return task.IssuerNotAuthorized
		}
	case risk > a.maxRisk, risk >= task.MediumRisk && !a.allowMedium:
		return task.IssuerNotAuthorized
	}
	return 0
}

// namespaced reports whether a label key lies in one of the reviewed literal
// namespaces: equal to it or below it at a '.' boundary, as with the reserved
// "barometer." namespace (REQ-LABEL-0011). An empty list permits no labels.
func namespaced(key string, namespaces []string) bool {
	for _, ns := range namespaces {
		if key == ns || strings.HasPrefix(key, ns+".") {
			return true
		}
	}
	return false
}

func intersect(a, b []string) []string {
	var out []string
	for _, s := range a {
		if slices.Contains(b, s) {
			out = append(out, s)
		}
	}
	return out
}

// intersectLimits keeps the stricter value of each ceiling: the lower maximum
// and the higher minimum. An intersection no task can satisfy fails closed.
func intersectLimits(a, b task.PolicyLimits) task.PolicyLimits {
	return task.PolicyLimits{
		Attempt: task.AttemptLimits{
			DurationMS:    min(a.Attempt.DurationMS, b.Attempt.DurationMS),
			BytesSent:     min(a.Attempt.BytesSent, b.Attempt.BytesSent),
			BytesReceived: min(a.Attempt.BytesReceived, b.Attempt.BytesReceived),
			Connections:   min(a.Attempt.Connections, b.Attempt.Connections),
		},
		MaxRuns:                  min(a.MaxRuns, b.MaxRuns),
		MaxAttempts:              min(a.MaxAttempts, b.MaxAttempts),
		MaxValiditySeconds:       min(a.MaxValiditySeconds, b.MaxValiditySeconds),
		MinRepeatIntervalSeconds: max(a.MinRepeatIntervalSeconds, b.MinRepeatIntervalSeconds),
		MinRepeatJitterSeconds:   max(a.MinRepeatJitterSeconds, b.MinRepeatJitterSeconds),
		MinRetryDelayMS:          max(a.MinRetryDelayMS, b.MinRetryDelayMS),
		MaxRetryDelayMS:          min(a.MaxRetryDelayMS, b.MaxRetryDelayMS),
		MaxRetryJitterMS:         min(a.MaxRetryJitterMS, b.MaxRetryJitterMS),
		MaxResultBytes:           min(a.MaxResultBytes, b.MaxResultBytes),
		MaxBundleBytes:           min(a.MaxBundleBytes, b.MaxBundleBytes),
		MaxResultsPerBundle:      min(a.MaxResultsPerBundle, b.MaxResultsPerBundle),
		MaxPendingBundles:        min(a.MaxPendingBundles, b.MaxPendingBundles),
		MaxPendingBytes:          min(a.MaxPendingBytes, b.MaxPendingBytes),
		MaxRetentionSeconds:      min(a.MaxRetentionSeconds, b.MaxRetentionSeconds),
	}
}

// narrow checks every task value that a ceiling bounds and returns the
// ceilings narrowed to those values. A value beyond a ceiling, including a
// repeat interval, jitter or retry delay below a policy minimum, raises
// authority and is rejected, never clamped. Result, bundle and retention
// ceilings have no task field; they are carried for result and storage
// enforcement. base is the trusted validity reference used by expiry.
func narrow(l task.PolicyLimits, c task.Candidate, base time.Time) (task.PolicyLimits, bool) {
	e, s := c.Execution.Limits, c.Schedule
	if e.DurationMS > l.Attempt.DurationMS || e.BytesSent > l.Attempt.BytesSent || e.BytesReceived > l.Attempt.BytesReceived || e.Connections > l.Attempt.Connections {
		return l, false
	}
	l.Attempt = e
	if s.MaxRuns == nil || *s.MaxRuns > l.MaxRuns || s.Retry.MaxAttempts > l.MaxAttempts {
		return l, false
	}
	l.MaxRuns, l.MaxAttempts = *s.MaxRuns, s.Retry.MaxAttempts
	if s.NotAfter != nil {
		if s.NotAfter.AfterTime(base.Add(time.Duration(l.MaxValiditySeconds) * time.Second)) {
			return l, false
		}
	}
	if s.Repeat.Mode == "interval" {
		i, j := s.Repeat.IntervalSeconds, s.Repeat.JitterSeconds
		if i == nil || j == nil || *i < l.MinRepeatIntervalSeconds || *j < l.MinRepeatJitterSeconds {
			return l, false
		}
		l.MinRepeatIntervalSeconds, l.MinRepeatJitterSeconds = *i, *j
	}
	if b := s.Retry.Backoff; b != nil {
		if b.InitialDelayMS < l.MinRetryDelayMS || b.MaxDelayMS > l.MaxRetryDelayMS || b.JitterMS > l.MaxRetryJitterMS {
			return l, false
		}
		l.MinRetryDelayMS, l.MaxRetryDelayMS, l.MaxRetryJitterMS = b.InitialDelayMS, b.MaxDelayMS, b.JitterMS
	}
	return l, true
}
