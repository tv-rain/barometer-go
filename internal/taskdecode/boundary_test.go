package taskdecode

import (
	"strings"
	"testing"

	"github.com/tv-rain/barometer-go/internal/model/task"
)

func TestRegisteredNumericalBoundaries(t *testing.T) {
	d := fixtureDecoder(t)
	for _, tc := range []struct {
		path      string
		low, high int
	}{
		{"execution.limits.max_duration_ms", 1, 60000},
		{"execution.limits.max_bytes_sent", 1, 65536},
		{"execution.limits.max_bytes_received", 1, 262144},
		{"execution.limits.max_connections", 1, 12},
		{"execution.device.minimum_battery_percent", 0, 100},
		{"schedule.limits.max_runs", 1, 28},
	} {
		for _, value := range []int{tc.low, tc.high} {
			if _, err := d.Decode(change(tc.path, value, false), timing()); err != nil {
				t.Fatal("registered numerical endpoint rejected")
			}
		}
	}
	for _, n := range []int{1, 128} {
		if _, err := d.Decode(change("task_id", strings.Repeat("a", n), false), timing()); err != nil {
			t.Fatal("identifier boundary rejected")
		}
	}
	for _, n := range []int{64, 65} {
		c, err := d.Decode(change("labels", map[string]string{strings.Repeat("a", n): "x"}, false), timing())
		if (err == nil) != (n == 64) || (err != nil && c != nil) {
			t.Fatal("label key boundary differs")
		}
	}
	// Byte counts apply to decoded UTF-8 label values as well as encoded objects.
	for _, n := range []int{64, 65} {
		c, err := d.Decode(change("labels", map[string]string{"campaign": strings.Repeat("я", n)}, false), timing())
		if (err == nil) != (n == 64) || (err != nil && c != nil) {
			t.Fatal("UTF-8 label boundary differs")
		}
	}
	for _, mode := range []string{"fixed", "exponential"} {
		for _, jitter := range []int{0, 60000} {
			r := retry(2, mode)
			b := r["backoff"].(map[string]any)
			b["initial_delay_ms"] = 300000
			b["max_delay_ms"] = 300000
			b["jitter_ms"] = jitter
			c, err := d.Decode(change("schedule.retry", r, false), timing())
			if err != nil || c.Schedule.Retry.Backoff.JitterMS != jitter {
				t.Fatal("backoff endpoint or explicit zero jitter differs")
			}
		}
	}
	for _, mode := range []string{"fixed", "exponential"} {
		for _, tc := range []struct {
			field string
			value int
		}{
			{"initial_delay_ms", 29999}, {"initial_delay_ms", 300001}, {"max_delay_ms", 600001}, {"jitter_ms", -1}, {"jitter_ms", 60001}, {"multiplier", 3},
		} {
			r := retry(2, mode)
			r["backoff"].(map[string]any)[tc.field] = tc.value
			if c, err := d.Decode(change("schedule.retry", r, false), timing()); c != nil || err == nil {
				t.Fatal("invalid backoff endpoint accepted")
			}
		}
	}
	s := repeated()
	s["repeat"].(map[string]any)["interval"] = 604800
	s["repeat"].(map[string]any)["jitter"] = 1800
	if _, err := d.Decode(change("schedule", s, false), timing()); err != nil {
		t.Fatal("repeat upper endpoints rejected")
	}
}

func TestExecutionAndSafetyValues(t *testing.T) {
	d := fixtureDecoder(t)
	for _, path := range []string{"direct", "proxy"} {
		if _, err := d.Decode(change("execution.network_path", path, false), timing()); err != nil {
			t.Fatal("supported path rejected")
		}
	}
	for _, policy := range []string{"any", "exclude", "require"} {
		if _, err := d.Decode(change("execution.network.vpn_policy", policy, false), timing()); err != nil {
			t.Fatal("supported VPN policy rejected")
		}
	}
	for _, mode := range []string{"shared", "exclusive"} {
		if _, err := d.Decode(change("execution.concurrency.mode", mode, false), timing()); err != nil {
			t.Fatal("supported concurrency rejected")
		}
	}
	for _, transports := range [][]string{{}, {"wifi"}, {"cellular"}, {"wifi", "cellular"}} {
		if _, err := d.Decode(change("execution.network.allowed_transports", transports, false), timing()); err != nil {
			t.Fatal("supported transports rejected")
		}
	}
	if _, err := d.Decode(change("execution.network.context", map[string]any{"required_fields": []string{"country_code", "region_code", "asn"}, "allowed_country_codes": []string{"RU", "NL", "BQ"}}, false), timing()); err != nil {
		t.Fatal("supported context constraints rejected")
	}
	for _, level := range []string{"low", "medium", "high", "unknown"} {
		c, err := d.Decode(change("safety", map[string]any{"risk_level": level, "requires_user_consent": true, "risk_factors": []any{map[string]any{"type": "network_observability", "level": level}}}, false), timing())
		if err != nil || !c.Safety.RequiresConsent || c.Safety.Factors[0].Level != level {
			t.Fatal("supported safety declaration rejected or rewritten")
		}
	}
	for _, reason := range []task.ExecutionReason{task.PlatformSuspended, task.NetworkContextUnavailable, task.ProxyUnavailable, task.InternalError} {
		if !task.Retryable(reason) {
			t.Fatal("registered retry reason rejected")
		}
	}
}
