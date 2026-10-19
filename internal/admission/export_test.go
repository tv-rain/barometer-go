package admission

import (
	"time"

	"github.com/tv-rain/barometer-go/internal/model/task"
)

// Test-only views of the unpublished gate outcome.
func GatedCandidate(g Gated) task.Candidate  { return g.candidate }
func GatedAcceptedAt(g Gated) time.Time      { return g.acceptedAt }
func GatedTest(g Gated) task.TestVersion     { return g.entry.Test }
func GatedSource(g Gated) task.SourceContext { return g.record.Source() }
func GatedBaselineRisk(g Gated) task.Risk    { return g.entry.BaselineRisk }
