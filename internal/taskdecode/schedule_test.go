package taskdecode

import (
	"errors"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/validation"
)

func repeated() map[string]any {
	return map[string]any{"not_after": "2026-10-21T00:00:00Z", "repeat": map[string]any{"mode": "interval", "interval": 21600, "jitter": 300}, "limits": map[string]any{"max_runs": 28}}
}
func retry(attempts int, mode string) map[string]any {
	b := map[string]any{"mode": mode, "initial_delay_ms": 30000, "max_delay_ms": 600000}
	if mode == "exponential" {
		b["multiplier"] = 2
	}
	return map[string]any{"max_attempts": attempts, "retry_on": []string{"platform_suspended", "internal_error"}, "backoff": b}
}
func ranges(pairs ...[2]string) []any {
	a := []any{}
	for _, pair := range pairs {
		a = append(a, map[string]any{"start": pair[0], "end": pair[1]})
	}
	return a
}

func TestScheduleRelationships(t *testing.T) {
	d := fixtureDecoder(t)
	for _, mode := range []string{"fixed", "exponential"} {
		s := repeated()
		s["retry"] = retry(2, mode)
		c, err := d.Decode(change("schedule", s, false), timing())
		if err != nil || c.Schedule.Retry.Backoff.JitterMS != 10000 {
			t.Fatal("valid bounded repeat/retry or jitter default rejected")
		}
	}
	for _, mutate := range []struct {
		name  string
		apply func(map[string]any)
	}{
		{"interval-too-low", func(s map[string]any) { s["repeat"].(map[string]any)["interval"] = 21599 }},
		{"interval-too-high", func(s map[string]any) { s["repeat"].(map[string]any)["interval"] = 604801 }},
		{"jitter-too-low", func(s map[string]any) { s["repeat"].(map[string]any)["jitter"] = 299 }},
		{"jitter-too-high", func(s map[string]any) { s["repeat"].(map[string]any)["jitter"] = 1801 }},
		{"jitter-missing", func(s map[string]any) { delete(s["repeat"].(map[string]any), "jitter") }},
		{"run-limit-missing", func(s map[string]any) { delete(s, "limits") }},
		{"expiry-missing", func(s map[string]any) { delete(s, "not_after") }},
		{"retry-empty-on", func(s map[string]any) { r := retry(2, "fixed"); r["retry_on"] = []string{}; s["retry"] = r }},
		{"retry-duplicates", func(s map[string]any) {
			r := retry(2, "fixed")
			r["retry_on"] = []string{"internal_error", "internal_error"}
			s["retry"] = r
		}},
		{"retry-no-backoff", func(s map[string]any) { r := retry(2, "fixed"); delete(r, "backoff"); s["retry"] = r }},
		{"backoff-delay-relationship", func(s map[string]any) {
			r := retry(2, "fixed")
			r["backoff"].(map[string]any)["initial_delay_ms"] = 300000
			r["backoff"].(map[string]any)["max_delay_ms"] = 30000
			s["retry"] = r
		}},
		{"fixed-multiplier", func(s map[string]any) {
			r := retry(2, "fixed")
			r["backoff"].(map[string]any)["multiplier"] = 2
			s["retry"] = r
		}},
		{"exponential-no-multiplier", func(s map[string]any) {
			r := retry(2, "exponential")
			delete(r["backoff"].(map[string]any), "multiplier")
			s["retry"] = r
		}},
		{"backoff-mode", func(s map[string]any) { s["retry"] = retry(2, "linear") }},
		{"attempt-count", func(s map[string]any) { s["retry"] = retry(3, "fixed") }},
		{"backoff-unknown", func(s map[string]any) {
			r := retry(2, "fixed")
			r["backoff"].(map[string]any)["extra"] = true
			s["retry"] = r
		}},
	} {
		t.Run(mutate.name, func(t *testing.T) {
			s := repeated()
			mutate.apply(s)
			c, err := d.Decode(change("schedule", s, false), timing())
			if c != nil || err == nil {
				t.Fatal("invalid schedule relationship accepted")
			}
		})
	}
	for _, reason := range []string{"dns_failure", "tcp_timeout", "tls_failure", "http_error", "target_unavailable", "cancelled_by_host", "proxy_reconfigured", "constraints_unsatisfied", "policy_rejected", "resource_limit"} {
		r := retry(2, "fixed")
		r["retry_on"] = []string{reason}
		if c, err := d.Decode(change("schedule.retry", r, false), timing()); c != nil || !errors.Is(err, validation.ErrTask) {
			t.Fatal("measurement/other non-retry reason accepted")
		}
	}
	for _, reason := range []task.ExecutionReason{task.PlatformSuspended, task.NetworkContextUnavailable, task.ProxyUnavailable, task.InternalError} {
		r := retry(2, "fixed")
		r["retry_on"] = []string{string(reason)}
		if _, err := d.Decode(change("schedule.retry", r, false), timing()); err != nil {
			t.Fatal("registered cross-cutting retry reason rejected")
		}
	}
}

func TestWindowValidation(t *testing.T) {
	d := fixtureDecoder(t)
	cases := []struct {
		name  string
		pairs [][2]string
		valid bool
	}{
		{"day", [][2]string{{"08:00", "20:00"}}, true},
		{"midnight", [][2]string{{"22:00", "06:00"}}, true},
		{"adjacent", [][2]string{{"22:00", "06:00"}, {"06:00", "08:00"}}, true},
		{"four", [][2]string{{"00:00", "01:00"}, {"02:00", "03:00"}, {"04:00", "05:00"}, {"06:00", "07:00"}}, true},
		{"five", [][2]string{{"00:00", "01:00"}, {"02:00", "03:00"}, {"04:00", "05:00"}, {"06:00", "07:00"}, {"08:00", "09:00"}}, false},
		{"empty", nil, false},
		{"equal-empty-window", [][2]string{{"08:00", "08:00"}}, true},
		{"duplicate", [][2]string{{"08:00", "20:00"}, {"08:00", "20:00"}}, false},
		{"midnight-overlap", [][2]string{{"22:00", "06:00"}, {"05:00", "07:00"}}, false},
		{"enclosed-overlap", [][2]string{{"08:00", "20:00"}, {"10:00", "11:00"}}, false},
		{"invalid-hour", [][2]string{{"24:00", "06:00"}}, false},
		{"invalid-minute", [][2]string{{"22:60", "06:00"}}, false},
		{"short-hour", [][2]string{{"8:00", "20:00"}}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := d.Decode(change("schedule.time_window", map[string]any{"ranges": ranges(tc.pairs...)}, false), timing())
			if (err == nil) != tc.valid || (err != nil && c != nil) {
				t.Fatal("time-window rule differs")
			}
		})
	}
	for _, timezone := range []string{"", "UTC", "Europe/Amsterdam"} {
		if c, err := d.Decode(change("schedule.time_window", map[string]any{"timezone": timezone, "ranges": ranges([2]string{"08:00", "20:00"})}, false), timing()); c != nil || !errors.Is(err, validation.ErrTask) {
			t.Fatal("unsupported timezone accepted")
		}
	}
	if c, err := d.Decode(change("schedule.time_window", map[string]any{"ranges": []any{map[string]any{"start": "08:00", "end": "20:00", "extra": true}}}, false), timing()); c != nil || !errors.Is(err, jsondecode.ErrField) {
		t.Fatal("unknown range field accepted")
	}
}

func TestTrustedTimingBoundary(t *testing.T) {
	d := fixtureDecoder(t)
	for _, tc := range []struct {
		name  string
		ctx   validation.TimingContext
		valid bool
	}{
		{"remote", timing(), true},
		{"no-context", validation.TimingContext{}, false},
		{"remote-no-issuance", validation.TimingContext{Origin: "remote", AcceptedAt: accepted}, false},
		{"remote-future-edge", validation.TimingContext{Origin: "remote", AcceptedAt: accepted, IssuedAt: timePtr(accepted.Add(300 * time.Second))}, true},
		{"remote-future-over", validation.TimingContext{Origin: "remote", AcceptedAt: accepted, IssuedAt: timePtr(accepted.Add(300*time.Second + time.Nanosecond))}, false},
		{"remote-stale", validation.TimingContext{Origin: "remote", AcceptedAt: accepted, IssuedAt: timePtr(accepted.Add(-7 * 24 * time.Hour))}, false},
		{"embedded", validation.TimingContext{Origin: "embedded", AcceptedAt: accepted}, true},
		{"embedded-issued", validation.TimingContext{Origin: "embedded", AcceptedAt: accepted, IssuedAt: &accepted}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, err := d.Decode([]byte(baseJSON), tc.ctx)
			if (err == nil) != tc.valid || (err != nil && c != nil) {
				t.Fatal("trusted timing rule differs")
			}
		})
	}
	noExpiry := change("schedule", map[string]any{}, false)
	ctx := validation.TimingContext{Origin: "embedded", AcceptedAt: accepted, EmbeddedOneTimeException: true}
	if _, err := d.Decode(noExpiry, ctx); err != nil {
		t.Fatal("reviewed embedded one-time exception rejected")
	}
	ctx.EmbeddedOneTimeException = false
	if c, err := d.Decode(noExpiry, ctx); c != nil || err == nil {
		t.Fatal("unreviewed embedded expiry omission accepted")
	}
	ctx.EmbeddedOneTimeException = true
	if c, err := d.Decode(change("schedule", map[string]any{"limits": map[string]any{"max_runs": 2}}, false), ctx); c != nil || err == nil {
		t.Fatal("embedded exception broadened beyond one run")
	}
	for _, key := range []string{"origin", "issued_at", "authenticated", "issuer_id", "policy_generation"} {
		if c, err := d.Decode(change(key, true, false), timing()); c != nil || !errors.Is(err, jsondecode.ErrField) {
			t.Fatal("task supplied trusted context")
		}
	}
}

func TestStartEligibilityAndRetrySeparation(t *testing.T) {
	d := fixtureDecoder(t)
	s := repeated()
	s["starts_at"] = "2026-10-14T00:00:00Z"
	s["retry"] = retry(2, "fixed")
	s["time_window"] = map[string]any{"ranges": ranges([2]string{"22:00", "06:00"})}
	c, err := d.Decode(change("schedule", s, false), timing())
	if err != nil {
		t.Fatal("eligibility fixture rejected")
	}
	ctx := validation.StartContext{Now: accepted, DueAt: accepted, LocalZone: time.UTC, Supported: true, ConstraintsSatisfied: true, SafetySatisfied: true}
	if !validation.RunEligible(c.Schedule, ctx) {
		t.Fatal("inclusive start or midnight window rejected")
	}
	for _, mutate := range []func(*validation.StartContext){
		func(x *validation.StartContext) { x.Now = accepted.Add(-time.Nanosecond) },
		func(x *validation.StartContext) { x.Now = accepted.Add(7 * 24 * time.Hour) },
		func(x *validation.StartContext) { x.Now = accepted.Add(6 * time.Hour) },
		func(x *validation.StartContext) { x.DueAt = accepted.Add(time.Second) },
		func(x *validation.StartContext) { x.CompletedRuns = 28 },
		func(x *validation.StartContext) { x.CompletedRuns = -1 },
		func(x *validation.StartContext) { x.LocalZone = nil },
		func(x *validation.StartContext) { x.Supported = false },
		func(x *validation.StartContext) { x.ConstraintsSatisfied = false },
		func(x *validation.StartContext) { x.SafetySatisfied = false },
	} {
		x := ctx
		mutate(&x)
		if validation.RunEligible(c.Schedule, x) {
			t.Fatal("unsatisfied start gate accepted")
		}
	}
	// Device-local zone is supplied explicitly; no process timezone is consulted.
	x := ctx
	x.LocalZone = time.FixedZone("synthetic", 7*3600)
	if validation.RunEligible(c.Schedule, x) {
		t.Fatal("device-local window ignored")
	}
	x = ctx
	x.CompletedRuns = 28
	if !validation.RetryEligible(c.Schedule, x, 1, task.InternalError) {
		t.Fatal("retry consumed a new run")
	}
	if validation.RetryEligible(c.Schedule, ctx, 0, task.InternalError) || validation.RetryEligible(c.Schedule, ctx, 2, task.InternalError) || validation.RetryEligible(c.Schedule, ctx, 1, task.ExecutionReason("tcp_timeout")) || validation.RetryEligible(c.Schedule, ctx, 1, task.ProxyUnavailable) {
		t.Fatal("retry attempts/reasons broadened")
	}
	ctx.SafetySatisfied = false
	if validation.RetryEligible(c.Schedule, ctx, 1, task.InternalError) {
		t.Fatal("retry bypassed safety")
	}
	c, err = d.Decode([]byte(baseJSON), timing())
	if err != nil {
		t.Fatal("one-run fixture rejected")
	}
	ctx.SafetySatisfied = true
	ctx.CompletedRuns = 1
	if validation.RunEligible(c.Schedule, ctx) {
		t.Fatal("one-time task repeated")
	}
}

func TestTimestampFractionIsExact(t *testing.T) {
	d := fixtureDecoder(t)
	c, err := d.Decode(change("schedule.starts_at", "2026-10-14T00:00:00.0000000001Z", false), timing())
	if err != nil {
		t.Fatal("sub-nanosecond timestamp rejected")
	}
	ctx := validation.StartContext{Now: accepted, DueAt: accepted, Supported: true, ConstraintsSatisfied: true, SafetySatisfied: true}
	if validation.RunEligible(c.Schedule, ctx) {
		t.Fatal("timestamp fraction truncated into early eligibility")
	}
	ctx.Now = accepted.Add(time.Nanosecond)
	if !validation.RunEligible(c.Schedule, ctx) {
		t.Fatal("exact timestamp fraction ordered incorrectly")
	}
	if c, err := d.Decode(change("schedule.not_after", "2026-10-21T00:00:00.0000000001Z", false), timing()); c != nil || !errors.Is(err, validation.ErrTiming) {
		t.Fatal("fraction truncated to evade lifetime maximum")
	}
	a, _ := d.Decode(change("schedule.starts_at", "2026-10-14T00:00:00.1000+00:00", false), timing())
	b, _ := d.Decode(change("schedule.starts_at", "2026-10-14T00:00:00.1Z", false), timing())
	if a.Schedule.StartsAt.Fraction != b.Schedule.StartsAt.Fraction || !a.Schedule.StartsAt.Whole.Equal(b.Schedule.StartsAt.Whole) {
		t.Fatal("UTC timestamp normalization differs")
	}
	for _, timestamp := range []string{"2026-10-14t00:00:00.1000z", "2026-10-14T00:00:00.1-00:00"} {
		c, err := d.Decode(change("schedule.starts_at", timestamp, false), timing())
		if err != nil || c.Schedule.StartsAt.Fraction != b.Schedule.StartsAt.Fraction || !c.Schedule.StartsAt.Whole.Equal(b.Schedule.StartsAt.Whole) {
			t.Fatal("RFC 3339 UTC spelling normalization differs")
		}
	}
}

func timePtr(t time.Time) *time.Time { return &t }
