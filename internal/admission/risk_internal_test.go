package admission

import (
	"testing"

	"github.com/tv-rain/barometer-go/internal/model/task"
)

// Levels outside the reviewed set, including the zero value of a missing
// input, are unknown and never lower effective risk.
// requirement-ref: REQ-SAFETY-0029
func TestRestrictiveFailsClosed(t *testing.T) {
	for _, c := range []struct{ a, b, want task.Risk }{
		{task.LowRisk, task.LowRisk, task.LowRisk},
		{task.LowRisk, task.HighRisk, task.HighRisk},
		{task.HighRisk, task.MediumRisk, task.HighRisk},
		{task.HighRisk, task.UnknownRisk, task.UnknownRisk},
		{task.UnknownRisk, task.LowRisk, task.UnknownRisk},
		{0, task.LowRisk, task.UnknownRisk},
		{task.LowRisk, task.UnknownRisk + 1, task.UnknownRisk},
	} {
		if got := restrictive(c.a, c.b); got != c.want {
			t.Errorf("restrictive(%d, %d) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
	for _, level := range []string{"", "none", "LOW", "Unknown"} {
		if declaredRisk(level) != task.UnknownRisk {
			t.Errorf("declared level %q not unknown", level)
		}
	}
	// A matched registration missing from the policy counts as unknown.
	a := Authority{registrationIDs: []string{"absent"}, maxRisk: task.HighRisk, allowMedium: true}
	if a.evaluateRisk(task.LocalPolicy{}, task.LowRisk, "low") != task.IssuerNotAuthorized {
		t.Error("missing registration did not fail closed")
	}
	a.allowUnknown = true
	if a.evaluateRisk(task.LocalPolicy{}, task.LowRisk, "low") != 0 || a.effectiveRisk != task.UnknownRisk {
		t.Error("missing registration not evaluated as unknown")
	}
}
