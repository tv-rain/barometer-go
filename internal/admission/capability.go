package admission

import (
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/registry"
)

// requiresCompositeSteps rejects a composite codec that supplies no bounded,
// normalized step list. Production NET-05 supplies one; synthetic or incomplete
// entries remain unsupported rather than bypassing primitive admission.
// requirement-ref: REQ-TRUST-0036
func requiresCompositeSteps(e registry.Entry, params task.Parameters) bool {
	if !e.Composite {
		return false
	}
	p, ok := params.(task.CompositeParameters)
	return !ok || len(p.CompositeSteps()) < 1 || len(p.CompositeSteps()) > 8
}
