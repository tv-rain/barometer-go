package taskdecode

import (
	"errors"
	"testing"

	"github.com/tv-rain/barometer-go/internal/validation"
)

// requirement-ref: REQ-SAFETY-0012
// requirement-ref: REQ-SAFETY-0013
// requirement-ref: REQ-SAFETY-0031
// requirement-ref: REQ-SCHEDULE-0013
// requirement-ref: REQ-SCHEDULE-0014
// requirement-ref: REQ-SCHEDULE-0043
func TestRiskFactorAndTimezoneDomains(t *testing.T) {
	d := fixtureDecoder(t)
	safety := func(level string, factors []any) map[string]any {
		s := map[string]any{"risk_level": level, "requires_user_consent": true}
		if factors != nil {
			s["risk_factors"] = factors
		}
		return s
	}
	one := func(level string) []any {
		return []any{map[string]any{"type": "network_observability", "level": level}}
	}
	for _, tc := range []struct {
		level   string
		factors []any
		valid   bool
	}{
		{"low", nil, true},
		{"low", []any{}, true},
		{"low", one("low"), true},
		{"medium", nil, false},
		{"medium", []any{}, false},
		{"medium", one("medium"), true},
		{"high", nil, false},
		{"high", []any{}, false},
		{"high", one("high"), true},
		{"unknown", nil, false},
		{"unknown", []any{}, false},
		{"unknown", one("unknown"), true},
	} {
		for range 3 {
			c, err := d.Decode(change("safety", safety(tc.level, tc.factors), false), timing())
			if tc.valid && (err != nil || c == nil || c.Safety.RiskLevel != tc.level || len(c.Safety.Factors) != len(tc.factors)) {
				t.Fatalf("%s risk with %d factors rejected or rewritten", tc.level, len(tc.factors))
			}
			if !tc.valid && (c != nil || !errors.Is(err, validation.ErrTask)) {
				t.Fatalf("%s risk with %v factors accepted", tc.level, tc.factors)
			}
		}
	}
	window := func(zone *string) map[string]any {
		w := map[string]any{"ranges": ranges([2]string{"08:00", "20:00"})}
		if zone != nil {
			w["timezone"] = *zone
		}
		return w
	}
	local := "local"
	for _, zone := range []*string{nil, &local} {
		c, err := d.Decode(change("schedule.time_window", window(zone), false), timing())
		if err != nil || c == nil || c.Schedule.Window.Timezone != "local" {
			t.Fatal("local or omitted timezone rejected or not treated as local")
		}
	}
	for _, zone := range []string{"UTC", "Europe/Berlin", "", "Local", "LOCAL", "local "} {
		if c, err := d.Decode(change("schedule.time_window", window(&zone), false), timing()); c != nil || !errors.Is(err, validation.ErrTask) {
			t.Fatalf("undefined timezone %q accepted", zone)
		}
	}
}
