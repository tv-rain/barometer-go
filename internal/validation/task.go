// Package validation owns pure task validation. It cannot admit or execute tasks.
package validation

import (
	"errors"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/tv-rain/barometer-go/internal/model/task"
)

var ErrTask = errors.New("validation: task")
var taskID = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,127}$`)
var labelKey = regexp.MustCompile(`^[a-z][a-z0-9_.-]{0,63}$`)

// Common checks candidate invariants without making policy/trust decisions.
// It applies the documented max_runs default and canonical label ordering.
// Original encoded byte bounds remain the decoder's responsibility.
func Common(c task.Candidate) (task.Candidate, error) {
	if !taskID.MatchString(c.ID) || c.SchemaVersion != "0.1.0" {
		return task.Candidate{}, ErrTask
	}
	if !validSchedule(c.Schedule) || !validExecution(c.Execution, c.TestID == "NET-01") || !validSafety(c.Safety) {
		return task.Candidate{}, ErrTask
	}
	// requirement-ref: REQ-LABEL-0005
	// requirement-ref: REQ-LABEL-0010
	// requirement-ref: REQ-LABEL-0014
	if len(c.Labels) > 16 {
		return task.Candidate{}, ErrTask
	}
	seen := make(map[string]bool, len(c.Labels))
	for _, l := range c.Labels {
		if !labelKey.MatchString(l.Key) || strings.HasPrefix(l.Key, "barometer.") || len(l.Value) > 128 || !utf8.ValidString(l.Value) || seen[l.Key] {
			return task.Candidate{}, ErrTask
		}
		seen[l.Key] = true
		for _, r := range l.Value {
			if unicode.IsControl(r) {
				return task.Candidate{}, ErrTask
			}
		}
	}
	if c.Schedule.MaxRuns == nil {
		n := 1
		c.Schedule.MaxRuns = &n
	}
	// requirement-ref: REQ-LABEL-0004
	// Copy before sorting: normalization must not rewrite caller-owned values.
	if len(c.Labels) == 0 {
		c.Labels = nil
	} else {
		c.Labels = slices.Clone(c.Labels)
		slices.SortFunc(c.Labels, func(a, b task.Label) int { return strings.Compare(a.Key, b.Key) })
	}
	return c, nil
}

func validSchedule(s task.Schedule) bool {
	if s.StartsAt != nil && s.NotAfter != nil && !s.StartsAt.Before(*s.NotAfter) {
		return false
	}
	if s.Window != nil {
		// requirement-ref: REQ-SCHEDULE-0013
		// requirement-ref: REQ-SCHEDULE-0014
		// requirement-ref: REQ-SCHEDULE-0043
		// local is the only defined timezone; decoding defaults an omitted value to it.
		if s.Window.Timezone != "local" || len(s.Window.Ranges) < 1 || len(s.Window.Ranges) > 4 {
			return false
		}
		var covered [1440]bool
		seen := map[task.TimeRange]bool{}
		for _, r := range s.Window.Ranges {
			if r.Start < 0 || r.Start >= 1440 || r.End < 0 || r.End >= 1440 || seen[r] {
				return false
			}
			seen[r] = true
			for m := r.Start; m != r.End; m = (m + 1) % 1440 {
				if covered[m] {
					return false
				}
				covered[m] = true
			}
		}
	}
	if s.MaxRuns != nil && (*s.MaxRuns < 1 || *s.MaxRuns > 28) {
		return false
	}
	switch s.Repeat.Mode {
	case "none":
		if s.Repeat.IntervalSeconds != nil || s.Repeat.JitterSeconds != nil {
			return false
		}
	case "interval":
		if s.NotAfter == nil || s.MaxRuns == nil || s.Repeat.IntervalSeconds == nil || s.Repeat.JitterSeconds == nil {
			return false
		}
		i, j := *s.Repeat.IntervalSeconds, *s.Repeat.JitterSeconds
		if i < 21600 || i > 604800 || j < 300 || j > 1800 || j > i/4 {
			return false
		}
	default:
		return false
	}
	r := s.Retry
	if r.MaxAttempts == 1 {
		return r.On == nil && r.Backoff == nil
	}
	if r.MaxAttempts != 2 || len(r.On) == 0 || len(r.On) > 4 || r.Backoff == nil {
		return false
	}
	seen := map[task.ExecutionReason]bool{}
	for _, reason := range r.On {
		if !task.Retryable(reason) || seen[reason] {
			return false
		}
		seen[reason] = true
	}
	b := r.Backoff
	if b.InitialDelayMS < 30000 || b.InitialDelayMS > 300000 || b.MaxDelayMS < b.InitialDelayMS || b.MaxDelayMS > 600000 || b.JitterMS < 0 || b.JitterMS > 60000 {
		return false
	}
	switch b.Mode {
	case "fixed":
		return b.Multiplier == nil
	case "exponential":
		return b.Multiplier != nil && *b.Multiplier == 2
	}
	return false
}

func validExecution(e task.Execution, zeroTraffic bool) bool {
	if e.NetworkPath != "direct" && e.NetworkPath != "proxy" {
		return false
	}
	l := e.Limits
	if l.DurationMS < 1 || l.DurationMS > 60000 {
		return false
	}
	if zeroTraffic {
		if l.BytesSent != 0 || l.BytesReceived != 0 || l.Connections != 0 {
			return false
		}
	} else if l.BytesSent < 1 || l.BytesSent > 65536 || l.BytesReceived < 1 || l.BytesReceived > 262144 || l.Connections < 1 || l.Connections > 12 {
		return false
	}
	for _, s := range e.Network.AllowedTransports {
		if s != "wifi" && s != "cellular" {
			return false
		}
	}
	if e.Network.VPNPolicy != "any" && e.Network.VPNPolicy != "exclude" && e.Network.VPNPolicy != "require" {
		return false
	}
	ctx := e.Network.Context
	if !unique(ctx.RequiredFields) || !unique(ctx.AllowedCountryCodes) {
		return false
	}
	for _, f := range ctx.RequiredFields {
		if f != "country_code" && f != "region_code" && f != "asn" {
			return false
		}
	}
	if ctx.AllowedCountryCodes != nil && len(ctx.AllowedCountryCodes) == 0 {
		return false
	}
	for _, code := range ctx.AllowedCountryCodes {
		if !countryCode(code) {
			return false
		}
	}
	if e.Device.MinimumBatteryPercent != nil && (*e.Device.MinimumBatteryPercent < 0 || *e.Device.MinimumBatteryPercent > 100) {
		return false
	}
	return e.ConcurrencyMode == "" || e.ConcurrencyMode == "shared" || e.ConcurrencyMode == "exclusive"
}

// requirement-ref: REQ-SAFETY-0012
// requirement-ref: REQ-SAFETY-0013
// requirement-ref: REQ-SAFETY-0031
func validSafety(s task.Safety) bool {
	rank := map[string]int{"low": 1, "medium": 2, "high": 3, "unknown": 4}
	// Any risk other than low needs at least one factor; an empty list does not explain it.
	if rank[s.RiskLevel] == 0 || (s.RiskLevel != "low" && len(s.Factors) == 0) {
		return false
	}
	for _, f := range s.Factors {
		if rank[f.Level] == 0 {
			return false
		}
		switch f.Type {
		case "network_observability", "traffic_distinctiveness", "target_sensitivity", "data_sensitivity", "legal_or_social_context":
		default:
			return false
		}
		if f.Level == "unknown" && s.RiskLevel != "unknown" {
			return false
		}
		if s.RiskLevel != "unknown" && rank[f.Level] > rank[s.RiskLevel] {
			return false
		}
	}
	return true
}

func unique(values []string) bool {
	seen := map[string]bool{}
	for _, s := range values {
		if seen[s] {
			return false
		}
		seen[s] = true
	}
	return true
}

// ISO 3166-1 alpha-2 assigned country codes; no pseudo-country/unknown matching.
const countries = " AD AE AF AG AI AL AM AO AQ AR AS AT AU AW AX AZ BA BB BD BE BF BG BH BI BJ BL BM BN BO BQ BR BS BT BV BW BY BZ CA CC CD CF CG CH CI CK CL CM CN CO CR CU CV CW CX CY CZ DE DJ DK DM DO DZ EC EE EG EH ER ES ET FI FJ FK FM FO FR GA GB GD GE GF GG GH GI GL GM GN GP GQ GR GS GT GU GW GY HK HM HN HR HT HU ID IE IL IM IN IO IQ IR IS IT JE JM JO JP KE KG KH KI KM KN KP KR KW KY KZ LA LB LC LI LK LR LS LT LU LV LY MA MC MD ME MF MG MH MK ML MM MN MO MP MQ MR MS MT MU MV MW MX MY MZ NA NC NE NF NG NI NL NO NP NR NU NZ OM PA PE PF PG PH PK PL PM PN PR PS PT PW PY QA RE RO RS RU RW SA SB SC SD SE SG SH SI SJ SK SL SM SN SO SR SS ST SV SX SY SZ TC TD TF TG TH TJ TK TL TM TN TO TR TT TV TW TZ UA UG UM US UY UZ VA VC VE VG VI VN VU WF WS YE YT ZA ZM ZW "

func countryCode(s string) bool { return len(s) == 2 && strings.Contains(countries, " "+s+" ") }
