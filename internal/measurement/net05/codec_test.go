package net05

import (
	"testing"

	"github.com/tv-rain/barometer-go/internal/model/task"
)

type undeclared struct{}

func (undeclared) TaskParameters() {}

// A step that reports no declared maxima cannot be summed, so the profile
// fails closed even under the composite ceilings (GO-SPEC-ISSUE-021).
// requirement-ref: REQ-LIMIT-0035
func TestDeclaredMaximaFailClosed(t *testing.T) {
	ceiling := task.AttemptLimits{DurationMS: 60000, BytesSent: 65536, BytesReceived: 262144, Connections: 12}
	p := Parameters{id: "baseline", version: 1, steps: []task.PrimitiveStep{{ID: "s", Test: task.TestVersion{ID: "TCP-01", Version: "0.1.0"}, Parameters: undeclared{}}}}
	if p.ValidateLimits(ceiling) {
		t.Fatal("undeclared step admitted")
	}
}
