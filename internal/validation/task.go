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
// Original encoded byte bounds remain the decoder's responsibility. Every
// failure matches ErrTask; a value above a registered maximum also matches
// ErrLimit and an unrecognized risk factor type also matches ErrUnsupported.
func Common(c task.Candidate) (task.Candidate, error) {
	if err := TaskID(c.ID); err != nil {
		return task.Candidate{}, err
	}
	if c.SchemaVersion != "0.1.0" {
		return task.Candidate{}, ErrTask
	}
	if err := scheduleErr(c.Schedule); err != nil {
		return task.Candidate{}, err
	}
	if err := executionErr(c.Execution, c.TestID == "NET-01"); err != nil {
		return task.Candidate{}, err
	}
	if err := safetyErr(c.Safety); err != nil {
		return task.Candidate{}, err
	}
	// requirement-ref: REQ-LABEL-0005
	// requirement-ref: REQ-LABEL-0010
	// requirement-ref: REQ-LABEL-0014
	if len(c.Labels) > 16 {
		return task.Candidate{}, errTaskLimit
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

func validSchedule(s task.Schedule) bool { return scheduleErr(s) == nil }

// scheduleErr reports ErrTask, or errTaskLimit for a value above a registered
// schedule or retry maximum (REQ-LIMIT-0011).
func scheduleErr(s task.Schedule) error {
	if s.StartsAt != nil && s.NotAfter != nil && !s.StartsAt.Before(*s.NotAfter) {
		return ErrTask
	}
	if s.Window != nil {
		// requirement-ref: REQ-SCHEDULE-0013
		// requirement-ref: REQ-SCHEDULE-0014
		// requirement-ref: REQ-SCHEDULE-0043
		// local is the only defined timezone; decoding defaults an omitted value to it.
		if s.Window.Timezone != "local" {
			return ErrTask
		}
		if err := limit(len(s.Window.Ranges), 1, 4); err != nil {
			return err
		}
		var covered [1440]bool
		seen := map[task.TimeRange]bool{}
		for _, r := range s.Window.Ranges {
			if r.Start < 0 || r.Start >= 1440 || r.End < 0 || r.End >= 1440 || seen[r] {
				return ErrTask
			}
			seen[r] = true
			for m := r.Start; m != r.End; m = (m + 1) % 1440 {
				if covered[m] {
					return ErrTask
				}
				covered[m] = true
			}
		}
	}
	if s.MaxRuns != nil {
		if err := limit(*s.MaxRuns, 1, 28); err != nil {
			return err
		}
	}
	switch s.Repeat.Mode {
	case "none":
		if s.Repeat.IntervalSeconds != nil || s.Repeat.JitterSeconds != nil {
			return ErrTask
		}
	case "interval":
		if s.NotAfter == nil || s.MaxRuns == nil || s.Repeat.IntervalSeconds == nil || s.Repeat.JitterSeconds == nil {
			return ErrTask
		}
		i, j := *s.Repeat.IntervalSeconds, *s.Repeat.JitterSeconds
		if err := limit(i, 21600, 604800); err != nil {
			return err
		}
		if err := limit(j, 300, min(1800, i/4)); err != nil {
			return err
		}
	default:
		return ErrTask
	}
	r := s.Retry
	if r.MaxAttempts == 1 {
		if r.On != nil || r.Backoff != nil {
			return ErrTask
		}
		return nil
	}
	if err := limit(r.MaxAttempts, 1, 2); err != nil {
		return err
	}
	if len(r.On) == 0 || len(r.On) > 4 || r.Backoff == nil {
		return ErrTask
	}
	seen := map[task.ExecutionReason]bool{}
	for _, reason := range r.On {
		if !task.Retryable(reason) || seen[reason] {
			return ErrTask
		}
		seen[reason] = true
	}
	b := r.Backoff
	if err := limit(b.InitialDelayMS, 30000, 300000); err != nil {
		return err
	}
	if err := limit(b.MaxDelayMS, b.InitialDelayMS, 600000); err != nil {
		return err
	}
	if err := limit(b.JitterMS, 0, 60000); err != nil {
		return err
	}
	switch b.Mode {
	case "fixed":
		if b.Multiplier != nil {
			return ErrTask
		}
		return nil
	case "exponential":
		if b.Multiplier == nil {
			return ErrTask
		}
		return limit(*b.Multiplier, 2, 2)
	}
	return ErrTask
}

// executionErr reports ErrTask, or errTaskLimit for an attempt limit above
// its registered maximum (REQ-LIMIT-0006). NET-01 requires zero traffic limits.
func executionErr(e task.Execution, zeroTraffic bool) error {
	if e.NetworkPath != "direct" && e.NetworkPath != "proxy" {
		return ErrTask
	}
	l := e.Limits
	if err := limit(l.DurationMS, 1, 60000); err != nil {
		return err
	}
	if zeroTraffic {
		if l.BytesSent != 0 || l.BytesReceived != 0 || l.Connections != 0 {
			return ErrTask
		}
	} else {
		for _, err := range []error{limit(l.BytesSent, 1, 65536), limit(l.BytesReceived, 1, 262144), limit(l.Connections, 1, 12)} {
			if err != nil {
				return err
			}
		}
	}
	for _, s := range e.Network.AllowedTransports {
		if s != "wifi" && s != "cellular" {
			return ErrTask
		}
	}
	if e.Network.VPNPolicy != "any" && e.Network.VPNPolicy != "exclude" && e.Network.VPNPolicy != "require" {
		return ErrTask
	}
	ctx := e.Network.Context
	if !unique(ctx.RequiredFields) || !unique(ctx.AllowedCountryCodes) {
		return ErrTask
	}
	for _, f := range ctx.RequiredFields {
		if f != "country_code" && f != "region_code" && f != "asn" {
			return ErrTask
		}
	}
	if ctx.AllowedCountryCodes != nil && len(ctx.AllowedCountryCodes) == 0 {
		return ErrTask
	}
	for _, code := range ctx.AllowedCountryCodes {
		if !countryCode(code) {
			return ErrTask
		}
	}
	if e.Device.MinimumBatteryPercent != nil && (*e.Device.MinimumBatteryPercent < 0 || *e.Device.MinimumBatteryPercent > 100) {
		return ErrTask
	}
	if e.ConcurrencyMode != "" && e.ConcurrencyMode != "shared" && e.ConcurrencyMode != "exclusive" {
		return ErrTask
	}
	return nil
}

// safetyErr reports ErrTask, or errUnsupported for a risk factor type this
// build does not recognize: it rejects the task as unsupported rather than
// treating that factor's risk as unknown (REQ-SAFETY-0017, REQ-TRUST-0036).
// requirement-ref: REQ-SAFETY-0012
// requirement-ref: REQ-SAFETY-0013
// requirement-ref: REQ-SAFETY-0017
// requirement-ref: REQ-SAFETY-0031
func safetyErr(s task.Safety) error {
	rank := map[string]int{"low": 1, "medium": 2, "high": 3, "unknown": 4}
	// Any risk other than low needs at least one factor; an empty list does not explain it.
	if rank[s.RiskLevel] == 0 || (s.RiskLevel != "low" && len(s.Factors) == 0) {
		return ErrTask
	}
	unsupported := false
	for _, f := range s.Factors {
		if rank[f.Level] == 0 {
			return ErrTask
		}
		switch f.Type {
		case "network_observability", "traffic_distinctiveness", "target_sensitivity", "data_sensitivity", "legal_or_social_context":
		default:
			unsupported = true
		}
		if f.Level == "unknown" && s.RiskLevel != "unknown" {
			return ErrTask
		}
		if s.RiskLevel != "unknown" && rank[f.Level] > rank[s.RiskLevel] {
			return ErrTask
		}
	}
	if unsupported {
		return errUnsupported
	}
	return nil
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
