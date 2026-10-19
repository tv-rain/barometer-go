package validation_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/validation"
)

func candidate() task.Candidate {
	return task.Candidate{
		SchemaVersion: "0.1.0", ID: "synthetic-task", TestID: "SYN-01", TestVersion: "0.2.0",
		Schedule:  task.Schedule{Repeat: task.Repeat{Mode: "none"}, Retry: task.Retry{MaxAttempts: 1}},
		Execution: task.Execution{NetworkPath: "proxy", Limits: task.AttemptLimits{DurationMS: 1, BytesSent: 1, BytesReceived: 1, Connections: 1}, Network: task.NetworkRequirements{VPNPolicy: "require"}},
		Safety:    task.Safety{RiskLevel: "low", RequiresConsent: true},
	}
}

// requirement-ref: REQ-LABEL-0004
// requirement-ref: REQ-LABEL-0005
// requirement-ref: REQ-LABEL-0010
// requirement-ref: REQ-LABEL-0014
func TestCommonLabelInvariants(t *testing.T) {
	for _, tc := range []struct {
		name   string
		labels []task.Label
		valid  bool
	}{
		{"omitted", nil, true},
		{"empty", []task.Label{}, true},
		{"boundary", []task.Label{{Key: strings.Repeat("a", 64), Value: strings.Repeat("я", 64)}}, true},
		{"opaque", []task.Label{{Key: "campaign", Value: "${literal}"}}, true},
		{"long key", []task.Label{{Key: strings.Repeat("a", 65)}}, false},
		{"long value", []task.Label{{Key: "campaign", Value: strings.Repeat("я", 65)}}, false},
		{"duplicate", []task.Label{{Key: "campaign", Value: "a"}, {Key: "campaign", Value: "a"}}, false},
		{"invalid UTF-8", []task.Label{{Key: "campaign", Value: string([]byte{0xff})}}, false},
		{"control", []task.Label{{Key: "campaign", Value: "a\nb"}}, false},
		{"reserved", []task.Label{{Key: "barometer.private"}}, false},
		{"count boundary", labelCount(16), true},
		{"count excess", labelCount(17), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := candidate()
			c.Labels = tc.labels
			for range 20 {
				v, err := validation.Common(c)
				if tc.valid {
					if err != nil {
						t.Fatal(err)
					}
					if len(tc.labels) == 0 && v.Labels != nil {
						t.Fatal("empty labels were not normalized")
					}
				} else if !errors.Is(err, validation.ErrTask) || !reflect.DeepEqual(v, task.Candidate{}) {
					t.Fatalf("invalid model escaped: %v", err)
				}
			}
		})
	}
}

func labelCount(n int) []task.Label {
	labels := make([]task.Label, n)
	for i := range labels {
		labels[i] = task.Label{Key: string(rune('a' + i)), Value: "x"}
	}
	return labels
}

// requirement-ref: REQ-LABEL-0001
// requirement-ref: REQ-EXECUTION-0013
func TestCommonNormalizationIsIdempotentAndConservative(t *testing.T) {
	c := candidate()
	c.Labels = []task.Label{{Key: "z", Value: "${literal}"}, {Key: "a", Value: "Я"}}
	first, err := validation.Common(c)
	if err != nil {
		t.Fatal(err)
	}
	want := candidate()
	n := 1
	want.Schedule.MaxRuns = &n
	want.Labels = []task.Label{{Key: "a", Value: "Я"}, {Key: "z", Value: "${literal}"}}
	if !reflect.DeepEqual(first, want) || c.Schedule.MaxRuns != nil || c.Labels[0].Key != "z" {
		t.Fatal("normalization changed scope, input, version, consent, or opaque text")
	}
	second, err := validation.Common(first)
	if err != nil || !reflect.DeepEqual(first, second) {
		t.Fatal("normalization is not idempotent")
	}
	first.Labels[0].Value = "changed"
	if second.Labels[0].Value != "Я" || c.Labels[1].Value != "Я" {
		t.Fatal("normalized labels alias the input")
	}
	for _, limits := range []task.AttemptLimits{{DurationMS: 60001, BytesSent: 1, BytesReceived: 1, Connections: 1}, {DurationMS: 1, BytesSent: 0, BytesReceived: 1, Connections: 1}} {
		invalid := candidate()
		invalid.Execution.Limits = limits
		v, err := validation.Common(invalid)
		if !errors.Is(err, validation.ErrTask) || !reflect.DeepEqual(v, task.Candidate{}) {
			t.Fatal("invalid limit repaired")
		}
	}
}

// requirement-ref: REQ-SAFETY-0012
// requirement-ref: REQ-SAFETY-0013
// requirement-ref: REQ-SAFETY-0031
// requirement-ref: REQ-SCHEDULE-0013
// requirement-ref: REQ-SCHEDULE-0014
// requirement-ref: REQ-SCHEDULE-0043
func TestCommonRiskFactorAndTimezoneDomains(t *testing.T) {
	factor := func(level string) []task.RiskFactor {
		return []task.RiskFactor{{Type: "network_observability", Level: level}}
	}
	window := func(zone string) *task.TimeWindow {
		return &task.TimeWindow{Timezone: zone, Ranges: []task.TimeRange{{Start: 480, End: 1200}}}
	}
	type safetyCase struct {
		name   string
		safety task.Safety
		valid  bool
	}
	cases := []safetyCase{
		{"low omitted", task.Safety{RiskLevel: "low"}, true},
		{"low empty", task.Safety{RiskLevel: "low", Factors: []task.RiskFactor{}}, true},
		{"low one", task.Safety{RiskLevel: "low", Factors: factor("low")}, true},
	}
	for _, level := range []string{"medium", "high", "unknown"} {
		cases = append(cases,
			safetyCase{level + " omitted", task.Safety{RiskLevel: level}, false},
			safetyCase{level + " empty", task.Safety{RiskLevel: level, Factors: []task.RiskFactor{}}, false},
			safetyCase{level + " one", task.Safety{RiskLevel: level, Factors: factor(level)}, true},
		)
	}
	for _, tc := range cases {
		t.Run("risk factors "+tc.name, func(t *testing.T) {
			c := candidate()
			c.Safety = tc.safety
			v, err := validation.Common(c)
			if tc.valid && err != nil {
				t.Fatal(err)
			}
			if !tc.valid && (!errors.Is(err, validation.ErrTask) || !reflect.DeepEqual(v, task.Candidate{})) {
				t.Fatalf("risk factor declaration escaped: %v", err)
			}
		})
	}
	for _, tc := range []struct {
		zone  string
		valid bool
	}{
		{"local", true}, {"UTC", false}, {"Europe/Berlin", false}, {"", false}, {"Local", false}, {"LOCAL", false},
	} {
		t.Run("timezone "+tc.zone, func(t *testing.T) {
			c := candidate()
			c.Schedule.Window = window(tc.zone)
			v, err := validation.Common(c)
			if tc.valid && (err != nil || v.Schedule.Window.Timezone != "local") {
				t.Fatalf("local timezone rejected or rewritten: %v", err)
			}
			if !tc.valid && (!errors.Is(err, validation.ErrTask) || !reflect.DeepEqual(v, task.Candidate{})) {
				t.Fatalf("undefined timezone escaped: %v", err)
			}
		})
	}
}
