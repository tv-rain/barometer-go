package admission

import (
	"time"

	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/persistence"
)

// Test-only views of the unpublished gate outcome.
func GatedCandidate(g Gated) task.Candidate  { return g.candidate }
func GatedAcceptedAt(g Gated) time.Time      { return g.acceptedAt }
func GatedTest(g Gated) task.TestVersion     { return g.entry.Test }
func GatedSource(g Gated) task.SourceContext { return g.source }
func GatedBaselineRisk(g Gated) task.Risk    { return g.entry.BaselineRisk }

// Test-only views of the effective authority.
func GatedLimits(g Gated) task.PolicyLimits { return g.authority.limits }
func GatedPaths(g Gated) task.Paths         { return g.authority.paths }
func GatedRegistrations(g Gated) []string   { return g.authority.registrationIDs }
func GatedRequiresConsent(g Gated) bool     { return g.authority.requireConsent }
func GatedPolicyGeneration(g Gated) uint64  { return g.authority.generation }
func GatedAuthority(g Gated) Authority      { return g.authority }
func GatedEffectiveRisk(g Gated) task.Risk  { return g.authority.effectiveRisk }

// Test-only views of an admitted task.
func AdmittedRecord(a Admitted) persistence.ReplayRecord { return a.record }
func AdmittedCandidate(a Admitted) task.Candidate        { return a.gated.candidate }
