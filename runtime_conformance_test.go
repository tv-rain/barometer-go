package barometer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/admission"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/registry"
	"github.com/tv-rain/barometer-go/internal/validation"
)

// Runtime conformance family cases (DEV-640). They drive the root facade with
// the synthetic run adapters of runs_test.go: an injected clock, scripted
// consent, device and time-zone capabilities, a test run ledger and a scripted
// runner registered under the TCP-01 identifier. Where a case plan names a
// registered test only as the vehicle for a schedule or execution rule, the
// synthetic runner stands in for it; no registered measurement behavior is
// asserted, and none of these cases is runtime, platform or field evidence.

// at formats an instant relative to the epoch.
func at(d time.Duration) string { return epoch.Add(d).Format(time.RFC3339) }

// fixedRetry is a valid retry object with a second attempt for the reasons.
func fixedRetry(reasons ...string) string {
	return fmt.Sprintf(`"retry":{"max_attempts":2,"retry_on":[%s],"backoff":{"mode":"fixed","initial_delay_ms":30000,"max_delay_ms":30000}}`, quoted(reasons))
}

func quoted(values []string) string {
	q := make([]string, len(values))
	for i, v := range values {
		q[i] = fmt.Sprintf("%q", v)
	}
	return strings.Join(q, ",")
}

// scheduledTask is a TCP-01 task whose schedule object is exactly schedule.
func scheduledTask(id, schedule, execution string, consent bool) string {
	if execution != "" {
		execution = "," + execution
	}
	return fmt.Sprintf(`{"schema_version":"0.1.0","task_id":%q,"test_id":"TCP-01","test_version":"0.1.0","input":{},"options":{},"schedule":{%s},"execution":{"limits":{}%s},"safety":{"risk_level":"low","requires_user_consent":%t}}`,
		id, schedule, execution, consent)
}

// tryAdmit admits taskJSON through an embedded record and returns the handle
// and the fixed admission code, "ok" on acceptance.
func (h *harness) tryAdmit(t *testing.T, c *Client, taskJSON string) (string, string) {
	t.Helper()
	th, err := c.AdmitTask(taskJSON, h.embedded(t, taskJSON))
	if err == nil && th == "" || err != nil && th != "" {
		t.Fatalf("admission returned handle %q with %s", th, codeOf(err))
	}
	return th, codeOf(err)
}

// wantAdmission admits taskJSON and checks the fixed code. A rejection
// creates no replay record and no run.
func (h *harness) wantAdmission(t *testing.T, c *Client, name, taskJSON, want string) string {
	t.Helper()
	th, got := h.tryAdmit(t, c, taskJSON)
	if got != want {
		t.Fatalf("%s: admission = %s, want %s", name, got, want)
	}
	if want != "ok" && h.run.runner.Calls() != 0 {
		t.Fatalf("%s: a rejected task ran", name)
	}
	return th
}

// runToEnd starts a run and waits until it is terminal.
func runToEnd(t *testing.T, c *Client, th string) string {
	t.Helper()
	rh, err := c.StartRun(th)
	if err != nil {
		t.Fatal("start:", codeOf(err))
	}
	waitTerminal(t, c, rh)
	return rh
}

// startCode starts a run, waits for it when one starts, and returns the code.
func startCode(t *testing.T, c *Client, th string) string {
	t.Helper()
	rh, err := c.StartRun(th)
	if err == nil {
		waitTerminal(t, c, rh)
	}
	return codeOf(err)
}

// fullRecord is the part of a written Measurement Result these cases inspect.
type fullRecord struct {
	RunID     string `json:"run_id"`
	Attempt   int    `json:"attempt"`
	Execution struct {
		State  string `json:"state"`
		Reason string `json:"reason"`
	} `json:"execution"`
	Time struct {
		StartedAt string `json:"started_at"`
	} `json:"time"`
	Measurement *struct {
		Outcome string `json:"outcome"`
	} `json:"measurement"`
}

func fullRecords(t *testing.T, h *harness) []fullRecord {
	t.Helper()
	var out []fullRecord
	for _, b := range h.run.output.Records() {
		var r fullRecord
		if err := json.Unmarshal(b, &r); err != nil {
			t.Fatal(err)
		}
		out = append(out, r)
	}
	return out
}

// CC-SCHEDULE-FIELDS: omitted, present and conditional schedule fields, their
// defaults, wrong types and unknown members.
// requirement-ref: REQ-SCHEDULE-0002
func TestCcScheduleFields(t *testing.T) {
	notAfter := `"not_after":"` + at(24*time.Hour) + `"`
	t.Run("omitted optional fields take their defaults", func(t *testing.T) {
		h := runningHarness(t)
		h.run.runner.fail[1] = true
		c := h.initialized(t)
		th := h.wantAdmission(t, c, "not_after only", scheduledTask("fields-1", notAfter, "", false), "ok")
		rh := runToEnd(t, c, th)
		// Omitted retry permits one attempt; omitted repeat and limits make
		// the task one run: a later start resolves to the same run.
		got := fullRecords(t, h)
		if len(got) != 1 || got[0].Execution.State != "incomplete" || got[0].Execution.Reason != "internal_error" || h.run.runner.Calls() != 1 {
			t.Fatalf("results = %+v", got)
		}
		h.advance(7 * time.Hour)
		if again, err := c.StartRun(th); again != rh || err != nil {
			t.Fatal("omitted repeat created a second run:", codeOf(err))
		}
	})
	t.Run("every field present", func(t *testing.T) {
		h := runningHarness(t)
		c := h.initialized(t)
		s := `"starts_at":"` + at(0) + `",` + notAfter + `,"time_window":{"timezone":"local","ranges":[{"start":"12:00","end":"18:00"}]},"repeat":{"mode":"none"},` + fixedRetry("internal_error") + `,"limits":{"max_runs":1}`
		th := h.wantAdmission(t, c, "all fields", scheduledTask("fields-2", s, "", false), "ok")
		runToEnd(t, c, th)
		if len(h.run.output.Records()) != 1 {
			t.Fatal("the run wrote no result")
		}
	})
	cases := []struct{ name, schedule, origin, want string }{
		{"embedded one-time task without not_after", `"repeat":{"mode":"none"},"limits":{"max_runs":1}`, "embedded", "task_schema_invalid"},
		{"remote task without not_after", `"repeat":{"mode":"none"}`, "remote", "task_schema_invalid"},
		{"repeated task without not_after", `"repeat":{"mode":"interval","interval":21600,"jitter":300},"limits":{"max_runs":2}`, "embedded", "task_schema_invalid"},
		{"starts_at number", `"starts_at":5,` + notAfter, "embedded", "task_schema_invalid"},
		{"not_after not a timestamp", `"not_after":"tomorrow"`, "embedded", "task_schema_invalid"},
		{"time_window array", notAfter + `,"time_window":[]`, "embedded", "task_schema_invalid"},
		{"repeat string", notAfter + `,"repeat":"none"`, "embedded", "task_schema_invalid"},
		{"retry number", notAfter + `,"retry":1`, "embedded", "task_schema_invalid"},
		{"limits null", notAfter + `,"limits":null`, "embedded", "task_schema_invalid"},
		{"unknown schedule member", notAfter + `,"priority":"high"`, "embedded", "task_schema_invalid"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := runningHarness(t)
			c := h.initialized(t)
			taskJSON := scheduledTask("fields-x", tc.schedule, "", false)
			bytes, ref := h.deliver(t, tc.origin, taskJSON)
			th, err := c.AdmitTask(bytes, ref)
			if codeOf(err) != tc.want || th != "" {
				t.Fatalf("admission = %s, want %s", codeOf(err), tc.want)
			}
			if _, ok, _ := h.replay.Lookup(context.Background(), key("fields-x"), epoch); ok {
				t.Fatal("a rejected schedule left a replay record")
			}
		})
	}
}

// CC-SCHEDULE-SCHEDULING-MODEL: runs and attempts. Repetition creates runs with
// new identities; a retry is a second attempt of the same run and never a new
// run or a run count.
// requirement-ref: REQ-SCHEDULE-0003
// requirement-ref: REQ-SCHEDULE-0004
// requirement-ref: REQ-SCHEDULE-0005
// requirement-ref: REQ-SCHEDULE-0006
func TestCcScheduleSchedulingModel(t *testing.T) {
	h := runningHarness(t)
	h.run.runner.fail[1] = true
	c := h.initialized(t)
	s := `"not_after":"` + at(20*time.Hour) + `","repeat":{"mode":"interval","interval":21600,"jitter":300},` + fixedRetry("internal_error") + `,"limits":{"max_runs":3}`
	th := h.admit(t, c, scheduledTask("model-1", s, "", false))
	for range 3 {
		runToEnd(t, c, th)
		h.advance(6*time.Hour + 5*time.Minute)
	}
	if code := startCode(t, c, th); code != "run_not_due" {
		t.Fatal("a fourth run:", code)
	}
	got := fullRecords(t, h)
	if len(got) != 4 {
		t.Fatalf("results = %+v", got)
	}
	if got[0].RunID != got[1].RunID || got[0].Attempt != 1 || got[1].Attempt != 2 || got[0].Execution.Reason != "internal_error" || got[1].Execution.State != "completed" {
		t.Fatalf("run 1 = %+v", got[:2])
	}
	if got[2].Attempt != 1 || got[3].Attempt != 1 || got[2].RunID == got[1].RunID || got[3].RunID == got[2].RunID || got[3].RunID == got[0].RunID {
		t.Fatalf("runs 2 and 3 = %+v", got[2:])
	}
	if rec, _ := h.run.ledger.Record(key("model-1")); rec.Runs != 3 || h.run.runner.Calls() != 4 {
		t.Fatalf("run accounting %+v after %d attempts", rec, h.run.runner.Calls())
	}

	// repeat none with the same retry policy: exactly one run of at most two
	// attempts, however often the host calls.
	h = runningHarness(t)
	h.run.runner.fail[1], h.run.runner.fail[2] = true, true
	c = h.initialized(t)
	th = h.admit(t, c, scheduledTask("model-2", `"not_after":"`+at(20*time.Hour)+`","repeat":{"mode":"none"},`+fixedRetry("internal_error"), "", false))
	rh := runToEnd(t, c, th)
	for range 3 {
		h.advance(6*time.Hour + 5*time.Minute)
		if again, err := c.StartRun(th); again != rh || err != nil {
			t.Fatal("a one-time task started a second run:", codeOf(err))
		}
	}
	got = fullRecords(t, h)
	if len(got) != 2 || got[0].RunID != got[1].RunID || h.run.runner.Calls() != 2 {
		t.Fatalf("one-time results = %+v", got)
	}
	if rec, _ := h.run.ledger.Record(key("model-2")); rec.Runs != 1 {
		t.Fatalf("one-time run accounting %+v", rec)
	}
}

// CC-SCHEDULE-SCHEDULING-MODEL-START-TIME: starts_at is the earliest start, not
// an exact execution time; null or absent means eligible at acceptance.
// requirement-ref: REQ-SCHEDULE-0007
// requirement-ref: REQ-SCHEDULE-0008
// requirement-ref: REQ-SCHEDULE-0009
func TestCcScheduleSchedulingModelStartTime(t *testing.T) {
	notAfter := `"not_after":"` + at(24*time.Hour) + `"`
	cases := []struct {
		name, schedule string
		offset         time.Duration
		due            bool
	}{
		{"one second before starts_at", `"starts_at":"` + at(2*time.Hour) + `",` + notAfter, 2*time.Hour - time.Second, false},
		{"at starts_at", `"starts_at":"` + at(2*time.Hour) + `",` + notAfter, 2 * time.Hour, true},
		{"one second after starts_at", `"starts_at":"` + at(2*time.Hour) + `",` + notAfter, 2*time.Hour + time.Second, true},
		{"five hours after starts_at", `"starts_at":"` + at(2*time.Hour) + `",` + notAfter, 7 * time.Hour, true},
		{"starts_at null, at acceptance", `"starts_at":null,` + notAfter, 0, true},
		{"starts_at absent, at acceptance", notAfter, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := runningHarness(t)
			c := h.initialized(t)
			th := h.admit(t, c, scheduledTask("start-1", tc.schedule, "", false))
			h.advance(tc.offset)
			code := startCode(t, c, th)
			if tc.due != (code == "ok") || !tc.due && code != "run_not_due" {
				t.Fatalf("start = %s, want due=%t", code, tc.due)
			}
			if !tc.due {
				if h.run.runner.Calls() != 0 || len(h.run.output.Records()) != 0 {
					t.Fatal("a run started before starts_at")
				}
				return
			}
			// The result records the actual start, to minute precision,
			// not starts_at.
			got := fullRecords(t, h)
			want := epoch.Add(tc.offset).Truncate(time.Minute).Format(time.RFC3339)
			if len(got) != 1 || got[0].Time.StartedAt != want {
				t.Fatalf("started_at = %+v, want %s", got, want)
			}
		})
	}
	// A run that was not started at starts_at is not missed: the host may
	// start it later, and the lateness does not end the task.
	h := runningHarness(t)
	c := h.initialized(t)
	th := h.admit(t, c, scheduledTask("start-2", `"starts_at":"`+at(time.Hour)+`",`+notAfter, "", false))
	h.advance(23 * time.Hour)
	if code := startCode(t, c, th); code != "ok" {
		t.Fatal("a late start was refused:", code)
	}
}

// CC-SCHEDULE-SCHEDULING-MODEL-TIME-LIMIT: not_after is the task expiry, distinct
// from the recurring time window; remote expiry is relative to the earlier of
// issuance and acceptance; the embedded one-time exception is a trusted input.
// requirement-ref: REQ-SCHEDULE-0010
// requirement-ref: REQ-SCHEDULE-0011
func TestCcScheduleSchedulingModelTimeLimit(t *testing.T) {
	const week = 604800 * time.Second
	// Each package is issued at the epoch and accepted delay later, within the
	// delivery record lifetime; not_after is relative to the issuance.
	remote := []struct {
		name            string
		delay, notAfter time.Duration
		want            string
	}{
		{"not_after at acceptance", 0, 0, "task_expired"},
		{"not_after one second after acceptance", 0, time.Second, "ok"},
		{"one week after issuance and acceptance", 0, week, "ok"},
		{"one week plus one second", 0, week + time.Second, "task_limit_exceeded"},
		{"one week after the earlier issuance", 30 * time.Second, week, "ok"},
		{"one second beyond a week after the earlier issuance", 30 * time.Second, week + time.Second, "task_limit_exceeded"},
		{"not_after at a later acceptance", 30 * time.Second, 30 * time.Second, "task_expired"},
	}
	for _, tc := range remote {
		t.Run("remote "+tc.name, func(t *testing.T) {
			h := runningHarness(t)
			c := h.initialized(t)
			pkg := h.sign(scheduledTask("limit-1", `"not_after":"`+at(tc.notAfter)+`"`, "", false))
			ref := h.installPackage(t, pkg)
			_ = h.clock.Clock.Advance(tc.delay)
			if _, err := c.AdmitTask(pkg, ref); codeOf(err) != tc.want {
				t.Fatalf("admission = %s, want %s", codeOf(err), tc.want)
			}
		})
	}
	t.Run("not_after ends a time window", func(t *testing.T) {
		// The window opens daily at 22:00 local; not_after at 21:00 on the
		// second day ends the task before the second window opens.
		h := runningHarness(t)
		c := h.initialized(t)
		s := `"not_after":"` + at(33*time.Hour) + `","time_window":{"ranges":[{"start":"22:00","end":"23:00"}]},"repeat":{"mode":"interval","interval":21600,"jitter":300},"limits":{"max_runs":4}`
		th := h.admit(t, c, scheduledTask("limit-2", s, "", false))
		h.advance(10 * time.Hour) // 22:00 on the first day
		if code := startCode(t, c, th); code != "ok" {
			t.Fatal("first window:", code)
		}
		h.advance(24 * time.Hour) // 22:00 on the second day, after not_after
		if code := startCode(t, c, th); code != "run_not_due" {
			t.Fatal("a run inside the window but after not_after:", code)
		}
	})
	t.Run("embedded one-time task without not_after", func(t *testing.T) {
		// The facade takes no release-review input, so the exception is not
		// available through it and the omission fails closed.
		h := runningHarness(t)
		c := h.initialized(t)
		h.wantAdmission(t, c, "facade", scheduledTask("limit-3", `"repeat":{"mode":"none"},"limits":{"max_runs":1}`, "", false), "task_schema_invalid")
		// The validator honors the trusted exception only for one run.
		one, two := 1, 2
		s := task.Schedule{Repeat: task.Repeat{Mode: "none"}, MaxRuns: &one, Retry: task.Retry{MaxAttempts: 1}}
		ctx := validation.TimingContext{Origin: "embedded", AcceptedAt: epoch, EmbeddedOneTimeException: true}
		if err := validation.ScheduleAt(s, ctx); err != nil {
			t.Fatal("reviewed one-time exception rejected:", err)
		}
		ctx.EmbeddedOneTimeException = false
		if validation.ScheduleAt(s, ctx) == nil {
			t.Fatal("omission accepted without the exception")
		}
		ctx.EmbeddedOneTimeException = true
		s.MaxRuns = &two
		if validation.ScheduleAt(s, ctx) == nil {
			t.Fatal("exception accepted for more than one run")
		}
		ctx.Origin = "remote"
		issued := epoch
		ctx.IssuedAt = &issued
		s.MaxRuns = &one
		if validation.ScheduleAt(s, ctx) == nil {
			t.Fatal("exception accepted for a remote task")
		}
	})
}

// switchZone is a time-zone capability the test can change between calls.
type switchZone struct {
	mu  sync.Mutex
	loc *time.Location
}

func (z *switchZone) set(loc *time.Location) {
	z.mu.Lock()
	defer z.mu.Unlock()
	z.loc = loc
}

func (z *switchZone) Location(context.Context) (*time.Location, error) {
	z.mu.Lock()
	defer z.mu.Unlock()
	return z.loc, nil
}

var _ platform.TimeZone = (*switchZone)(nil)

// window is a schedule with a local time window of the given ranges, valid for
// a day, repeated so the window can be tested over the day.
func window(ranges string) string {
	return `"not_after":"` + at(24*time.Hour) + `","time_window":{"timezone":"local","ranges":[` + ranges + `]}`
}

func rng(start, end string) string { return fmt.Sprintf(`{"start":%q,"end":%q}`, start, end) }

// clockAt returns the offset from the epoch (12:00 UTC) to the next instant
// whose UTC wall-clock time is hh:mm.
func clockAt(hh, mm int) time.Duration {
	d := time.Duration(hh)*time.Hour + time.Duration(mm)*time.Minute - 12*time.Hour
	if d < 0 {
		d += 24 * time.Hour
	}
	return d
}

// CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW: runs start only inside a local
// range, start inclusive and end exclusive, including ranges crossing midnight
// and several ranges; a run due outside the window waits for the next range.
// requirement-ref: REQ-SCHEDULE-0012
func TestCcScheduleSchedulingModelTimeWindow(t *testing.T) {
	cases := []struct {
		name, ranges string
		hh, mm       int
		due          bool
	}{
		{"11:59 before 12:00-18:00", rng("12:00", "18:00"), 11, 59, false},
		{"12:00 in 12:00-18:00", rng("12:00", "18:00"), 12, 0, true},
		{"17:59 in 12:00-18:00", rng("12:00", "18:00"), 17, 59, true},
		{"18:00 after 12:00-18:00", rng("12:00", "18:00"), 18, 0, false},
		{"21:59 before 22:00-06:00", rng("22:00", "06:00"), 21, 59, false},
		{"22:00 in 22:00-06:00", rng("22:00", "06:00"), 22, 0, true},
		{"05:59 in 22:00-06:00", rng("22:00", "06:00"), 5, 59, true},
		{"06:00 after 22:00-06:00", rng("22:00", "06:00"), 6, 0, false},
		{"12:00 between two ranges", rng("08:00", "10:00") + "," + rng("14:00", "16:00"), 12, 0, false},
		{"14:00 in the second range", rng("08:00", "10:00") + "," + rng("14:00", "16:00"), 14, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := runningHarness(t)
			c := h.initialized(t)
			th := h.admit(t, c, scheduledTask("window-1", window(tc.ranges), "", false))
			h.advance(clockAt(tc.hh, tc.mm))
			code := startCode(t, c, th)
			if tc.due != (code == "ok") || !tc.due && code != "run_not_due" {
				t.Fatalf("start = %s, want due=%t", code, tc.due)
			}
			if !tc.due && h.run.runner.Calls() != 0 {
				t.Fatal("a run started outside the window")
			}
		})
	}
	t.Run("due outside the window waits for the next range start", func(t *testing.T) {
		h := runningHarness(t)
		c := h.initialized(t)
		th := h.admit(t, c, scheduledTask("window-2", window(rng("14:00", "16:00")), "", false))
		for _, d := range []time.Duration{0, time.Hour, time.Hour + 59*time.Minute} {
			h.advance(d - h.run.clock.Now().Sub(epoch))
			if code := startCode(t, c, th); code != "run_not_due" {
				t.Fatal("started before the range:", code)
			}
		}
		h.advance(time.Minute) // 14:00
		if code := startCode(t, c, th); code != "ok" {
			t.Fatal("not started at the range start:", code)
		}
		if got := fullRecords(t, h); len(got) != 1 || got[0].Time.StartedAt != epoch.Add(2*time.Hour).Format(time.RFC3339) {
			t.Fatalf("results = %+v", got)
		}
	})
	t.Run("a range opening after not_after never starts", func(t *testing.T) {
		h := runningHarness(t)
		c := h.initialized(t)
		th := h.admit(t, c, scheduledTask("window-3", `"not_after":"`+at(3*time.Hour)+`","time_window":{"ranges":[`+rng("16:00", "17:00")+`]}`, "", false))
		for _, d := range []time.Duration{time.Hour, 3*time.Hour - time.Second, 4 * time.Hour} {
			h.advance(d - h.run.clock.Now().Sub(epoch))
			if code := startCode(t, c, th); code != "run_not_due" {
				t.Fatal("started outside its window or after not_after:", code)
			}
		}
	})
}

// CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-FIELDS: range count, uniqueness and
// overlap, the timezone value and its use of the device-reported zone.
// requirement-ref: REQ-SCHEDULE-0013
// requirement-ref: REQ-SCHEDULE-0014
// requirement-ref: REQ-SCHEDULE-0043
func TestCcScheduleSchedulingModelTimeWindowFields(t *testing.T) {
	notAfter := `"not_after":"` + at(24*time.Hour) + `"`
	four := rng("00:00", "02:00") + "," + rng("04:00", "06:00") + "," + rng("08:00", "10:00") + "," + rng("12:00", "14:00")
	cases := []struct{ name, window, want string }{
		{"ranges omitted", `{"timezone":"local"}`, "task_schema_invalid"},
		{"zero ranges", `{"ranges":[]}`, "task_schema_invalid"},
		{"one range", `{"ranges":[` + rng("12:00", "18:00") + `]}`, "ok"},
		{"four ranges", `{"ranges":[` + four + `]}`, "ok"},
		{"five ranges", `{"ranges":[` + four + "," + rng("16:00", "18:00") + `]}`, "task_limit_exceeded"},
		{"identical ranges", `{"ranges":[` + rng("08:00", "12:00") + "," + rng("08:00", "12:00") + `]}`, "task_schema_invalid"},
		{"overlapping ranges", `{"ranges":[` + rng("08:00", "12:00") + "," + rng("11:00", "14:00") + `]}`, "task_schema_invalid"},
		{"overlap after midnight expansion", `{"ranges":[` + rng("22:00", "06:00") + "," + rng("05:00", "07:00") + `]}`, "task_schema_invalid"},
		{"timezone omitted", `{"ranges":[` + rng("12:00", "18:00") + `]}`, "ok"},
		{"timezone UTC", `{"timezone":"UTC","ranges":[` + rng("12:00", "18:00") + `]}`, "task_schema_invalid"},
		{"timezone Europe/Berlin", `{"timezone":"Europe/Berlin","ranges":[` + rng("12:00", "18:00") + `]}`, "task_schema_invalid"},
		{"timezone empty", `{"timezone":"","ranges":[` + rng("12:00", "18:00") + `]}`, "task_schema_invalid"},
		{"timezone Local", `{"timezone":"Local","ranges":[` + rng("12:00", "18:00") + `]}`, "task_schema_invalid"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := runningHarness(t)
			c := h.initialized(t)
			th := h.wantAdmission(t, c, tc.name, scheduledTask("tw-1", notAfter+`,"time_window":`+tc.window, "", false), tc.want)
			if tc.want == "ok" {
				// Accepted windows, including an omitted timezone, use the
				// device-reported zone: 12:00 UTC is inside 12:00-18:00
				// and inside 12:00-14:00.
				if code := startCode(t, c, th); code != "ok" {
					t.Fatal("start inside the window:", code)
				}
			} else if rec, ok := h.run.ledger.Record(key("tw-1")); ok || rec.Runs != 0 {
				t.Fatal("a rejected window created run accounting")
			}
		})
	}
	t.Run("local follows the current device zone", func(t *testing.T) {
		h := runningHarness(t)
		zone := &switchZone{loc: time.UTC}
		h.adapters.run.zone = zone
		c := h.initialized(t)
		th := h.admit(t, c, scheduledTask("tw-2", notAfter+`,"time_window":{"timezone":"local","ranges":[`+rng("12:00", "18:00")+`]}`, "", false))
		h.advance(clockAt(10, 30)) // 10:30Z
		if code := startCode(t, c, th); code != "run_not_due" {
			t.Fatal("10:30 local UTC started:", code)
		}
		zone.set(time.FixedZone("UTC+03:00", 3*60*60)) // the same instant is 13:30 local
		if code := startCode(t, c, th); code != "ok" {
			t.Fatal("13:30 local did not start:", code)
		}
	})
}

// CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-RANGE-FIELDS: HH:mm syntax, start
// inclusive and end exclusive, midnight expansion and overlap after it.
// requirement-ref: REQ-SCHEDULE-0015
// requirement-ref: REQ-SCHEDULE-0016
// requirement-ref: REQ-SCHEDULE-0017
func TestCcScheduleSchedulingModelTimeWindowRangeFields(t *testing.T) {
	syntax := []struct {
		start, end, want string
	}{
		{"00:00", "23:59", "ok"},
		{"09:05", "17:30", "ok"},
		{"24:00", "06:00", "task_schema_invalid"},
		{"7:00", "09:00", "task_schema_invalid"},
		{"12:60", "13:00", "task_schema_invalid"},
		{"12:00:00", "13:00", "task_schema_invalid"},
		{"noon", "13:00", "task_schema_invalid"},
		{"12:00", "", "task_schema_invalid"},
	}
	for _, tc := range syntax {
		t.Run(tc.start+"-"+tc.end, func(t *testing.T) {
			h := runningHarness(t)
			c := h.initialized(t)
			h.wantAdmission(t, c, tc.start, scheduledTask("rf-1", window(rng(tc.start, tc.end)), "", false), tc.want)
		})
	}
	members := map[string]string{
		"start missing":  `{"end":"18:00"}`,
		"end missing":    `{"start":"12:00"}`,
		"unknown member": `{"start":"12:00","end":"18:00","days":["mon"]}`,
		"start number":   `{"start":1200,"end":"18:00"}`,
	}
	for name, r := range members {
		t.Run(name, func(t *testing.T) {
			h := runningHarness(t)
			c := h.initialized(t)
			h.wantAdmission(t, c, name, scheduledTask("rf-2", window(r), "", false), "task_schema_invalid")
		})
	}
	bounds := []struct {
		name, ranges string
		hh, mm       int
		due          bool
	}{
		{"start is inclusive", rng("13:15", "13:45"), 13, 15, true},
		{"end is exclusive", rng("13:15", "13:45"), 13, 45, false},
		{"midnight range before midnight", rng("23:00", "01:00"), 23, 30, true},
		{"midnight range after midnight", rng("23:00", "01:00"), 0, 30, true},
		{"midnight range end", rng("23:00", "01:00"), 1, 0, false},
		{"midnight range outside", rng("23:00", "01:00"), 22, 59, false},
	}
	for _, tc := range bounds {
		t.Run(tc.name, func(t *testing.T) {
			h := runningHarness(t)
			c := h.initialized(t)
			th := h.admit(t, c, scheduledTask("rf-3", window(tc.ranges), "", false))
			h.advance(clockAt(tc.hh, tc.mm))
			if code := startCode(t, c, th); tc.due != (code == "ok") {
				t.Fatalf("start = %s, want due=%t", code, tc.due)
			}
		})
	}
	overlaps := map[string]string{
		"overlap only after midnight expansion": rng("22:00", "02:00") + "," + rng("01:00", "03:00"),
		"adjacent ranges do not overlap":        rng("22:00", "02:00") + "," + rng("02:00", "03:00"),
	}
	for name, ranges := range overlaps {
		t.Run(name, func(t *testing.T) {
			h := runningHarness(t)
			c := h.initialized(t)
			want := "task_schema_invalid"
			if strings.HasPrefix(name, "adjacent") {
				want = "ok"
			}
			h.wantAdmission(t, c, name, scheduledTask("rf-4", window(ranges), "", false), want)
		})
	}
}

// anyProxy is a synthetic rule under which every task is proxy-compatible.
type anyProxy struct{}

func (anyProxy) Compatible(task.Parameters) bool { return true }

// customHarness is runningHarness with the registry entry and the deployment
// and issuer permissions adjusted. entry may clear the runner.
func customHarness(t *testing.T, entry func(*registry.Entry), adjust func(in *task.PolicyInput)) *harness {
	t.Helper()
	h := newHarness(t)
	in := synPolicyInput(2, "reg-a", func(p *task.Permissions) { p.Tests = []task.TestVersion{tcp01} })
	in.Targets[0].Test = tcp01
	if adjust != nil {
		adjust(&in)
	}
	policy, err := task.NewLocalPolicy(in)
	if err != nil {
		t.Fatal("synthetic policy rejected")
	}
	e := registry.Entry{Test: tcp01, Codec: tcpCodec{}, BaselineRisk: task.LowRisk, Targets: tcpTargets{}, Proxy: registry.NoProxy{}, Runner: h.run.runner}
	if entry != nil {
		entry(&e)
	}
	reg, err := registry.New([]registry.Entry{e})
	if err != nil {
		t.Fatal("synthetic registry rejected")
	}
	h.adapters.policy, h.adapters.registry = policy, reg
	return h
}

// admitted returns the admission outputs behind a task handle.
func admitted(t *testing.T, c *Client, th string) admission.Admitted {
	t.Helper()
	c.mu.Lock()
	rt := c.rt
	c.mu.Unlock()
	a, err := rt.tasks.lookup(th)
	if err != nil {
		t.Fatal("task handle does not resolve")
	}
	return a
}

// CC-SCHEDULE-SCHEDULING-MODEL-REPEAT: none creates one run, interval creates
// runs at the interval up to max_runs, and interval-only fields are rejected
// for none.
// requirement-ref: REQ-SCHEDULE-0018
func TestCcScheduleSchedulingModelRepeat(t *testing.T) {
	notAfter := `"not_after":"` + at(6*24*time.Hour) + `"`
	runs := func(t *testing.T, schedule string, calls int) (int, *harness) {
		t.Helper()
		h := runningHarness(t)
		c := h.initialized(t)
		th := h.admit(t, c, scheduledTask("repeat-1", schedule, "", false))
		for range calls {
			if code := startCode(t, c, th); code != "ok" && code != "run_not_due" {
				t.Fatal(code)
			}
			h.advance(6*time.Hour + 5*time.Minute)
		}
		rec, _ := h.run.ledger.Record(key("repeat-1"))
		return rec.Runs, h
	}
	for name, s := range map[string]string{"repeat omitted": notAfter, "repeat none": notAfter + `,"repeat":{"mode":"none"}`} {
		t.Run(name, func(t *testing.T) {
			if n, h := runs(t, s, 6); n != 1 || h.run.runner.Calls() != 1 {
				t.Fatalf("%d runs, want 1", n)
			}
		})
	}
	for _, maxRuns := range []int{1, 2, 5} {
		t.Run(fmt.Sprintf("interval with max_runs %d", maxRuns), func(t *testing.T) {
			s := notAfter + fmt.Sprintf(`,"repeat":{"mode":"interval","interval":21600,"jitter":300},"limits":{"max_runs":%d}`, maxRuns)
			n, h := runs(t, s, 10)
			if n != maxRuns || h.run.runner.Calls() != maxRuns {
				t.Fatalf("%d runs, want %d", n, maxRuns)
			}
			if got := fullRecords(t, h); len(got) != maxRuns || maxRuns > 1 && got[0].RunID == got[1].RunID {
				t.Fatalf("results = %+v", got)
			}
		})
	}
	for name, repeat := range map[string]string{
		"interval with none":  `{"mode":"none","interval":21600}`,
		"jitter with none":    `{"mode":"none","jitter":300}`,
		"unregistered mode":   `{"mode":"daily"}`,
		"mode missing":        `{"interval":21600,"jitter":300}`,
		"unknown repeat item": `{"mode":"none","every":"day"}`,
	} {
		t.Run(name, func(t *testing.T) {
			h := runningHarness(t)
			c := h.initialized(t)
			h.wantAdmission(t, c, name, scheduledTask("repeat-2", notAfter+`,"repeat":`+repeat+`,"limits":{"max_runs":2}`, "", false), "task_schema_invalid")
		})
	}
}

// CC-SCHEDULE-SCHEDULING-MODEL-RETRY: a second attempt keeps the run identity and
// the run count; a terminal negative measurement is never retried.
// requirement-ref: REQ-SCHEDULE-0022
// requirement-ref: REQ-SCHEDULE-0023
// requirement-ref: REQ-SCHEDULE-0024
// requirement-ref: REQ-SCHEDULE-0025
func TestCcScheduleSchedulingModelRetry(t *testing.T) {
	h := runningHarness(t)
	h.run.runner.fail[1] = true
	c := h.initialized(t)
	s := `"not_after":"` + at(20*time.Hour) + `","repeat":{"mode":"interval","interval":21600,"jitter":300},` + fixedRetry("internal_error") + `,"limits":{"max_runs":2}`
	th := h.admit(t, c, scheduledTask("retry-1", s, "", false))
	runToEnd(t, c, th)
	if rec, _ := h.run.ledger.Record(key("retry-1")); rec.Runs != 1 {
		t.Fatalf("a retry counted as a run: %+v", rec)
	}
	h.advance(6*time.Hour + 5*time.Minute)
	runToEnd(t, c, th)
	h.advance(6*time.Hour + 5*time.Minute)
	if code := startCode(t, c, th); code != "run_not_due" {
		t.Fatal("the retry consumed no run, so max_runs 2 permits exactly two runs:", code)
	}
	got := fullRecords(t, h)
	if len(got) != 3 || got[0].RunID != got[1].RunID || got[1].Attempt != 2 || got[2].RunID == got[0].RunID || got[2].Attempt != 1 {
		t.Fatalf("results = %+v", got)
	}
	if waits := h.run.clock.Waits(); len(waits) != 1 || waits[0] != 30*time.Second {
		t.Fatalf("retry waits = %v", waits)
	}

	// A negative measurement outcome is a completed attempt, not an execution
	// failure, even when internal_error is retryable.
	h = runningHarness(t)
	h.run.runner.refuse[1] = true
	c = h.initialized(t)
	runToEnd(t, c, h.admit(t, c, scheduledTask("retry-2", `"not_after":"`+at(20*time.Hour)+`",`+fixedRetry("internal_error"), "", false)))
	got = fullRecords(t, h)
	if len(got) != 1 || got[0].Execution.State != "completed" || got[0].Measurement == nil || got[0].Measurement.Outcome != "connection_refused" || h.run.runner.Calls() != 1 {
		t.Fatalf("negative outcome = %+v", got)
	}
}

// retrySchedule is a valid not_after plus the given retry object.
func retrySchedule(retry string) string {
	return `"not_after":"` + at(20*time.Hour) + `","retry":` + retry
}

// CC-SCHEDULE-SCHEDULING-MODEL-RETRY-RETRY-FIELDS: permitted and forbidden retry
// reasons, conditional fields, fixed and exponential delays, and jitter capped
// at max_delay_ms and never extending beyond expiry.
// requirement-ref: REQ-SCHEDULE-0026
// requirement-ref: REQ-SCHEDULE-0027
// requirement-ref: REQ-SCHEDULE-0028
// requirement-ref: REQ-SCHEDULE-0029
func TestCcScheduleSchedulingModelRetryRetryFields(t *testing.T) {
	backoff := `"backoff":{"mode":"fixed","initial_delay_ms":30000,"max_delay_ms":30000}`
	retry := func(on string) string { return `{"max_attempts":2,"retry_on":[` + on + `],` + backoff + `}` }
	cases := []struct{ name, retry, want string }{
		{"platform_suspended", retry(`"platform_suspended"`), "ok"},
		{"network_context_unavailable", retry(`"network_context_unavailable"`), "ok"},
		{"proxy_unavailable", retry(`"proxy_unavailable"`), "ok"},
		{"internal_error", retry(`"internal_error"`), "ok"},
		{"all four reasons", retry(`"platform_suspended","network_context_unavailable","proxy_unavailable","internal_error"`), "ok"},
		{"cancelled_by_host", retry(`"cancelled_by_host"`), "task_schema_invalid"},
		{"constraints_unsatisfied", retry(`"constraints_unsatisfied"`), "task_schema_invalid"},
		{"proxy_reconfigured", retry(`"proxy_reconfigured"`), "task_schema_invalid"},
		{"resource_limit", retry(`"resource_limit"`), "task_schema_invalid"},
		{"attempt_deadline_exceeded", retry(`"attempt_deadline_exceeded"`), "task_schema_invalid"},
		{"measurement outcome", retry(`"connection_refused"`), "task_schema_invalid"},
		{"duplicate reason", retry(`"internal_error","internal_error"`), "task_schema_invalid"},
		{"empty retry_on", retry(``), "task_schema_invalid"},
		{"one attempt", `{"max_attempts":1}`, "ok"},
		{"one attempt with retry_on", `{"max_attempts":1,"retry_on":["internal_error"]}`, "task_schema_invalid"},
		{"one attempt with backoff", `{"max_attempts":1,` + backoff + `}`, "task_schema_invalid"},
		{"two attempts without retry_on", `{"max_attempts":2,` + backoff + `}`, "task_schema_invalid"},
		{"two attempts without backoff", `{"max_attempts":2,"retry_on":["internal_error"]}`, "task_schema_invalid"},
		{"max_attempts missing", `{"retry_on":["internal_error"],` + backoff + `}`, "task_schema_invalid"},
		{"fixed with multiplier", `{"max_attempts":2,"retry_on":["internal_error"],"backoff":{"mode":"fixed","initial_delay_ms":30000,"multiplier":2,"max_delay_ms":30000}}`, "task_schema_invalid"},
		{"exponential with multiplier 2", `{"max_attempts":2,"retry_on":["internal_error"],"backoff":{"mode":"exponential","initial_delay_ms":30000,"multiplier":2,"max_delay_ms":60000}}`, "ok"},
		{"exponential without multiplier", `{"max_attempts":2,"retry_on":["internal_error"],"backoff":{"mode":"exponential","initial_delay_ms":30000,"max_delay_ms":60000}}`, "task_schema_invalid"},
		{"exponential with multiplier 3", `{"max_attempts":2,"retry_on":["internal_error"],"backoff":{"mode":"exponential","initial_delay_ms":30000,"multiplier":3,"max_delay_ms":60000}}`, "task_limit_exceeded"},
		{"unregistered backoff mode", `{"max_attempts":2,"retry_on":["internal_error"],"backoff":{"mode":"linear","initial_delay_ms":30000,"max_delay_ms":30000}}`, "task_schema_invalid"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := runningHarness(t)
			c := h.initialized(t)
			h.wantAdmission(t, c, tc.name, scheduledTask("rf-1", retrySchedule(tc.retry), "", false), tc.want)
		})
	}
	// The realized delay: initial delay plus local jitter, capped at
	// max_delay_ms; fixed and exponential produce one delay with two attempts.
	delays := []struct {
		name    string
		backoff string
		jitter  time.Duration
		want    time.Duration
	}{
		{"fixed, no jitter drawn", `{"mode":"fixed","initial_delay_ms":30000,"max_delay_ms":60000}`, 0, 30 * time.Second},
		{"fixed, default jitter cap 10 s", `{"mode":"fixed","initial_delay_ms":30000,"max_delay_ms":60000}`, time.Hour, 40 * time.Second},
		{"exponential, jitter within max", `{"mode":"exponential","initial_delay_ms":30000,"multiplier":2,"max_delay_ms":60000,"jitter_ms":20000}`, 15 * time.Second, 45 * time.Second},
		{"jitter capped at max_delay_ms", `{"mode":"fixed","initial_delay_ms":30000,"max_delay_ms":45000,"jitter_ms":60000}`, time.Minute, 45 * time.Second},
	}
	for _, tc := range delays {
		t.Run(tc.name, func(t *testing.T) {
			h := runningHarness(t)
			h.adapters.run.jitter = fixedJitter(tc.jitter)
			h.run.runner.fail[1] = true
			c := h.initialized(t)
			runToEnd(t, c, h.admit(t, c, scheduledTask("rf-2", retrySchedule(`{"max_attempts":2,"retry_on":["internal_error"],"backoff":`+tc.backoff+`}`), "", false)))
			if waits := h.run.clock.Waits(); len(waits) != 1 || waits[0] != tc.want || h.run.runner.Calls() != 2 {
				t.Fatalf("waits = %v, want %v", waits, tc.want)
			}
		})
	}
	t.Run("a delay reaching not_after is not taken", func(t *testing.T) {
		h := runningHarness(t)
		h.adapters.run.jitter = fixedJitter(5 * time.Second)
		h.run.runner.fail[1] = true
		c := h.initialized(t)
		s := `"not_after":"` + at(35*time.Second) + `","retry":{"max_attempts":2,"retry_on":["internal_error"],"backoff":` + `{"mode":"fixed","initial_delay_ms":30000,"max_delay_ms":60000}}`
		runToEnd(t, c, h.admit(t, c, scheduledTask("rf-3", s, "", false)))
		if len(h.run.clock.Waits()) != 0 || h.run.runner.Calls() != 1 || len(h.run.output.Records()) != 1 {
			t.Fatalf("waits %v, %d attempts", h.run.clock.Waits(), h.run.runner.Calls())
		}
	})
}

// CC-SCHEDULE-VALIDATION: every listed schedule constraint is enforced at
// admission; an exact redelivery adds no schedule and keeps the accepted
// expiry; invalid or expired schedules are admission failures, not results.
// requirement-ref: REQ-SCHEDULE-0033
// requirement-ref: REQ-SCHEDULE-0036
// requirement-ref: REQ-SCHEDULE-0037
// requirement-ref: REQ-SCHEDULE-0038
// requirement-ref: REQ-SCHEDULE-0039
// requirement-ref: REQ-SCHEDULE-0040
// requirement-ref: REQ-SCHEDULE-0041
// requirement-ref: REQ-SCHEDULE-0042
func TestCcScheduleValidation(t *testing.T) {
	notAfter := `"not_after":"` + at(20*time.Hour) + `"`
	interval := `"repeat":{"mode":"interval","interval":21600,"jitter":300}`
	backoff := `"backoff":{"mode":"fixed","initial_delay_ms":30000,"max_delay_ms":30000}`
	cases := []struct{ name, schedule, want string }{
		// REQ-SCHEDULE-0033: unknown members at every level.
		{"unknown schedule member", notAfter + `,"deadline":"x"`, "task_schema_invalid"},
		{"unknown time_window member", notAfter + `,"time_window":{"ranges":[` + rng("12:00", "18:00") + `],"days":["mon"]}`, "task_schema_invalid"},
		{"unknown range member", notAfter + `,"time_window":{"ranges":[{"start":"12:00","end":"18:00","weekday":1}]}`, "task_schema_invalid"},
		{"unknown repeat member", notAfter + `,"repeat":{"mode":"none","count":2}`, "task_schema_invalid"},
		{"unknown retry member", notAfter + `,"retry":{"max_attempts":1,"on":"x"}`, "task_schema_invalid"},
		{"unknown backoff member", notAfter + `,"retry":{"max_attempts":2,"retry_on":["internal_error"],"backoff":{"mode":"fixed","initial_delay_ms":30000,"max_delay_ms":30000,"factor":1}}`, "task_schema_invalid"},
		{"unknown limits member", notAfter + `,"limits":{"max_runs":1,"max_attempts":1}`, "task_schema_invalid"},
		// REQ-SCHEDULE-0036: RFC 3339 UTC timestamps, starts_at before not_after.
		{"not_after with an offset", `"not_after":"` + epoch.Add(20*time.Hour).In(time.FixedZone("", 7200)).Format(time.RFC3339) + `"`, "task_schema_invalid"},
		{"starts_at with an offset", `"starts_at":"` + epoch.In(time.FixedZone("", 3600)).Format(time.RFC3339) + `",` + notAfter, "task_schema_invalid"},
		{"timestamp without a zone", `"not_after":"` + epoch.Add(20*time.Hour).Format("2006-01-02T15:04:05") + `"`, "task_schema_invalid"},
		{"starts_at equal to not_after", `"starts_at":"` + at(20*time.Hour) + `",` + notAfter, "task_schema_invalid"},
		{"starts_at after not_after", `"starts_at":"` + at(21*time.Hour) + `",` + notAfter, "task_schema_invalid"},
		{"starts_at before not_after", `"starts_at":"` + at(time.Hour) + `",` + notAfter, "ok"},
		// REQ-SCHEDULE-0037: ranges unique and non-overlapping after expansion.
		{"duplicate ranges", notAfter + `,"time_window":{"ranges":[` + rng("01:00", "02:00") + "," + rng("01:00", "02:00") + `]}`, "task_schema_invalid"},
		{"ranges overlapping after expansion", notAfter + `,"time_window":{"ranges":[` + rng("23:00", "02:00") + "," + rng("01:30", "03:00") + `]}`, "task_schema_invalid"},
		// REQ-SCHEDULE-0038: repeat defaults to none; interval fields rejected for none.
		{"repeat omitted", notAfter, "ok"},
		{"interval field with none", notAfter + `,"repeat":{"mode":"none","interval":21600,"jitter":300}`, "task_schema_invalid"},
		// REQ-SCHEDULE-0039: repeated tasks need bounded expiry, max_runs and jitter.
		{"repeated without not_after", interval + `,"limits":{"max_runs":2}`, "task_schema_invalid"},
		{"repeated without max_runs", notAfter + `,` + interval, "task_schema_invalid"},
		{"repeated without jitter", notAfter + `,"repeat":{"mode":"interval","interval":21600},"limits":{"max_runs":2}`, "task_schema_invalid"},
		{"repeated with every bound", notAfter + `,` + interval + `,"limits":{"max_runs":2}`, "ok"},
		// REQ-SCHEDULE-0040: retry relationships.
		{"retry_on without a second attempt", notAfter + `,"retry":{"max_attempts":1,"retry_on":["internal_error"]}`, "task_schema_invalid"},
		{"second attempt without backoff", notAfter + `,"retry":{"max_attempts":2,"retry_on":["internal_error"]}`, "task_schema_invalid"},
		{"retry on a measurement outcome", notAfter + `,"retry":{"max_attempts":2,"retry_on":["tls_failure"],` + backoff + `}`, "task_schema_invalid"},
		// REQ-SCHEDULE-0042: an expired schedule is an admission failure.
		{"not_after already passed", `"not_after":"` + at(-time.Second) + `"`, "task_expired"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := runningHarness(t)
			c := h.initialized(t)
			th := h.wantAdmission(t, c, tc.name, scheduledTask("val-1", tc.schedule, "", false), tc.want)
			if tc.want != "ok" {
				if th != "" || len(h.run.output.Records()) != 0 || h.run.output.Calls() != 0 {
					t.Fatal("an admission failure produced a handle or a Measurement Result")
				}
			}
		})
	}
	t.Run("exact redelivery adds no schedule and keeps the expiry", func(t *testing.T) {
		h := runningHarness(t)
		c := h.initialized(t)
		taskJSON := scheduledTask("val-2", `"not_after":"`+at(7*time.Hour)+`",`+interval+`,"limits":{"max_runs":3}`, "", false)
		first := h.admit(t, c, taskJSON)
		runToEnd(t, c, first)
		h.advance(time.Hour)
		again, err := c.AdmitTask(taskJSON, h.embedded(t, taskJSON))
		if err != nil || again != first {
			t.Fatal("redelivery:", codeOf(err))
		}
		if a := admitted(t, c, again); !a.ExpiresAt().Equal(epoch.Add(7 * time.Hour)) {
			t.Fatal("redelivery moved the expiry")
		}
		// Occurrence 1 is due at 6h; the redelivery did not restart the
		// schedule or reset the run count.
		h.advance(5*time.Hour + 5*time.Minute)
		runToEnd(t, c, again)
		h.advance(6 * time.Hour)
		if code := startCode(t, c, again); code != "run_not_due" {
			t.Fatal("a run after the accepted expiry:", code)
		}
		if rec, _ := h.run.ledger.Record(key("val-2")); rec.Runs != 2 || h.run.runner.Calls() != 2 {
			t.Fatalf("run accounting after redelivery: %+v", rec)
		}
	})
}

// CC-SCHEDULE-VALIDATION-ELIGIBILITY-RULES: each eligibility predicate denied
// on its own, for a first attempt and for a retry, starts nothing; the probe
// keeps final authority to delay a due run.
// requirement-ref: REQ-SCHEDULE-0034
// requirement-ref: REQ-SCHEDULE-0035
func TestCcScheduleValidationEligibilityRules(t *testing.T) {
	notAfter := `"not_after":"` + at(20*time.Hour) + `"`
	initial := []struct {
		name                string
		harness             func(t *testing.T) *harness
		schedule, execution string
		consent             bool
		offset              time.Duration
		edit                func(h *harness)
	}{
		{"task not supported (no runner)", func(t *testing.T) *harness {
			return customHarness(t, func(e *registry.Entry) { e.Runner = nil }, nil)
		}, notAfter, "", false, 0, nil},
		{"before starts_at", runningHarness, `"starts_at":"` + at(time.Hour) + `",` + notAfter, "", false, 0, nil},
		{"at not_after", runningHarness, notAfter, "", false, 20 * time.Hour, nil},
		{"outside the time window", runningHarness, notAfter + `,"time_window":{"ranges":[` + rng("14:00", "16:00") + `]}`, "", false, 0, nil},
		{"execution constraint unsatisfied", runningHarness, notAfter, `"device":{"minimum_battery_percent":50}`, false, 0,
			func(h *harness) { h.run.device.Set(platform.DeviceState{BatteryPercent: 49, BatteryKnown: true}) }},
		{"safety requirement unsatisfied", runningHarness, notAfter, "", true, 0,
			func(h *harness) { h.run.consent.Set(platform.ConsentAbsent) }},
		{"probe declines: no pending-output capacity", runningHarness, notAfter, "", false, 0,
			func(h *harness) { h.run.output.FailCapacity() }},
	}
	for _, tc := range initial {
		t.Run("first attempt: "+tc.name, func(t *testing.T) {
			h := tc.harness(t)
			if tc.edit != nil {
				tc.edit(h)
			}
			c := h.initialized(t)
			th := h.admit(t, c, scheduledTask("elig-1", tc.schedule, tc.execution, tc.consent))
			h.advance(tc.offset)
			if code := startCode(t, c, th); code != "run_not_due" {
				t.Fatal("started:", code)
			}
			if h.run.runner.Calls() != 0 || h.run.paths.Calls() != 0 || len(h.run.output.Records()) != 0 {
				t.Fatal("an ineligible run reached the path, the runner or the output")
			}
			if rec, _ := h.run.ledger.Record(key("elig-1")); rec.Runs != 0 {
				t.Fatal("an ineligible start consumed a run")
			}
		})
	}
	t.Run("first attempt: max_runs reached", func(t *testing.T) {
		h := runningHarness(t)
		c := h.initialized(t)
		th := h.admit(t, c, scheduledTask("elig-2", notAfter+`,"repeat":{"mode":"interval","interval":21600,"jitter":300},"limits":{"max_runs":1}`, "", false))
		runToEnd(t, c, th)
		h.advance(6*time.Hour + 5*time.Minute)
		if code := startCode(t, c, th); code != "run_not_due" || h.run.runner.Calls() != 1 {
			t.Fatal("a run beyond max_runs:", code)
		}
	})
	// Retry attempts: attempt 1 fails with a retryable reason, then one
	// predicate changes before the retry; no second attempt starts.
	retry := []struct {
		name                string
		schedule, execution string
		consent             bool
		during              func(h *harness)
	}{
		{"constraint fails before the retry", notAfter + "," + fixedRetry("internal_error"), `"device":{"minimum_battery_percent":50}`, false,
			func(h *harness) { h.run.device.Set(platform.DeviceState{BatteryPercent: 10, BatteryKnown: true}) }},
		{"consent ends before the retry", notAfter + "," + fixedRetry("internal_error"), "", true,
			func(h *harness) { h.run.consent.Set(platform.ConsentEnded) }},
		{"retry would fall after not_after", `"not_after":"` + at(20*time.Second) + `",` + fixedRetry("internal_error"), "", false, nil},
	}
	for _, tc := range retry {
		t.Run("retry attempt: "+tc.name, func(t *testing.T) {
			h := runningHarness(t)
			h.run.device.Set(platform.DeviceState{BatteryPercent: 80, BatteryKnown: true})
			h.run.runner.fail[1] = true
			if tc.during != nil {
				h.run.runner.during = func(n int) {
					if n == 1 {
						tc.during(h)
					}
				}
			}
			c := h.initialized(t)
			runToEnd(t, c, h.admit(t, c, scheduledTask("elig-3", tc.schedule, tc.execution, tc.consent)))
			if h.run.runner.Calls() != 1 || len(h.run.output.Records()) != 1 {
				t.Fatalf("%d attempts after the predicate failed", h.run.runner.Calls())
			}
		})
	}
	t.Run("retry attempt: window closes before the retry", func(t *testing.T) {
		h := runningHarness(t)
		h.run.runner.fail[1] = true
		c := h.initialized(t)
		th := h.admit(t, c, scheduledTask("elig-4", notAfter+`,"time_window":{"ranges":[`+rng("12:00", "12:01")+`]},`+fixedRetry("internal_error"), "", false))
		h.advance(40 * time.Second) // the retry 30 s later falls at 12:01:10
		runToEnd(t, c, th)
		if h.run.runner.Calls() != 1 || len(h.run.output.Records()) != 1 {
			t.Fatalf("%d attempts; a retry started outside the window", h.run.runner.Calls())
		}
	})
}

// noTrustedTime is a trusted clock that is unavailable.
type noTrustedTime struct{}

func (noTrustedTime) Now(context.Context) (time.Time, error) {
	return time.Time{}, errors.New("synthetic unavailable trusted time")
}

// CC-LIMIT-SCHEDULE-AND-RETRY-LIMITS: every registered schedule and retry bound
// and default, the issuance skew and trusted-time absence for remote tasks, the
// embedded one-time exception and forbidden measurement-outcome retries.
// requirement-ref: REQ-LIMIT-0009
// requirement-ref: REQ-LIMIT-0010
// requirement-ref: REQ-LIMIT-0011
// requirement-ref: REQ-LIMIT-0012
// requirement-ref: REQ-LIMIT-0013
func TestCcLimitScheduleAndRetryLimits(t *testing.T) {
	notAfter := `"not_after":"` + at(6*24*time.Hour) + `"`
	repeat := func(interval, jitter, runs int) string {
		return fmt.Sprintf(`%s,"repeat":{"mode":"interval","interval":%d,"jitter":%d},"limits":{"max_runs":%d}`, notAfter, interval, jitter, runs)
	}
	retry := func(initial, maxDelay int, jitter string) string {
		return fmt.Sprintf(`%s,"retry":{"max_attempts":2,"retry_on":["internal_error"],"backoff":{"mode":"fixed","initial_delay_ms":%d,"max_delay_ms":%d%s}}`, notAfter, initial, maxDelay, jitter)
	}
	four := rng("00:00", "02:00") + "," + rng("04:00", "06:00") + "," + rng("08:00", "10:00") + "," + rng("12:00", "14:00")
	// Following the admission classes (GO-SPEC-ISSUE-008), a value above a
	// registered maximum is task_limit_exceeded and a value below a minimum
	// or outside a relationship is task_schema_invalid; a future issuance is
	// task_source_unauthenticated (docs/admission.md, step 1).
	cases := []struct{ name, schedule, want string }{
		{"runs 0", notAfter + `,"limits":{"max_runs":0}`, "task_schema_invalid"},
		{"runs 1", notAfter + `,"limits":{"max_runs":1}`, "ok"},
		{"runs 28", repeat(21600, 300, 28), "ok"},
		{"runs 29", repeat(21600, 300, 29), "task_limit_exceeded"},
		{"ranges 4", notAfter + `,"time_window":{"ranges":[` + four + `]}`, "ok"},
		{"ranges 5", notAfter + `,"time_window":{"ranges":[` + four + "," + rng("16:00", "18:00") + `]}`, "task_limit_exceeded"},
		{"interval 21599", repeat(21599, 300, 2), "task_schema_invalid"},
		{"interval 21600", repeat(21600, 300, 2), "ok"},
		{"interval 604800", `"not_after":"` + at(604800*time.Second) + `","repeat":{"mode":"interval","interval":604800,"jitter":300},"limits":{"max_runs":2}`, "ok"},
		{"interval 604801", repeat(604801, 300, 2), "task_limit_exceeded"},
		{"jitter 299", repeat(21600, 299, 2), "task_schema_invalid"},
		{"jitter 300", repeat(21600, 300, 2), "ok"},
		{"jitter 1800", repeat(21600, 1800, 2), "ok"},
		{"jitter 1801", repeat(21600, 1801, 2), "task_limit_exceeded"},
		{"attempts 2", retry(30000, 30000, ""), "ok"},
		{"attempts 3", notAfter + `,"retry":{"max_attempts":3,"retry_on":["internal_error"],"backoff":{"mode":"fixed","initial_delay_ms":30000,"max_delay_ms":30000}}`, "task_limit_exceeded"},
		{"initial delay 29999", retry(29999, 30000, ""), "task_schema_invalid"},
		{"initial delay 30000", retry(30000, 30000, ""), "ok"},
		{"initial delay 300000", retry(300000, 300000, ""), "ok"},
		{"initial delay 300001", retry(300001, 600000, ""), "task_limit_exceeded"},
		{"max delay below initial", retry(60000, 59999, ""), "task_schema_invalid"},
		{"max delay 600000", retry(30000, 600000, ""), "ok"},
		{"max delay 600001", retry(30000, 600001, ""), "task_limit_exceeded"},
		{"retry jitter 0", retry(30000, 30000, `,"jitter_ms":0`), "ok"},
		{"retry jitter 60000", retry(30000, 90000, `,"jitter_ms":60000`), "ok"},
		{"retry jitter 60001", retry(30000, 90000, `,"jitter_ms":60001`), "task_limit_exceeded"},
		{"measurement-outcome retry", notAfter + `,"retry":{"max_attempts":2,"retry_on":["dns_failure"],"backoff":{"mode":"fixed","initial_delay_ms":30000,"max_delay_ms":30000}}`, "task_schema_invalid"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := runningHarness(t)
			c := h.initialized(t)
			h.wantAdmission(t, c, tc.name, scheduledTask("lim-1", tc.schedule, "", false), tc.want)
		})
	}
	t.Run("defaults: one run, one attempt, 10 s retry jitter", func(t *testing.T) {
		h := runningHarness(t)
		h.run.runner.fail[1] = true
		c := h.initialized(t)
		th := h.admit(t, c, scheduledTask("lim-2", notAfter, "", false))
		runToEnd(t, c, th)
		if h.run.runner.Calls() != 1 {
			t.Fatal("omitted retry permitted a second attempt")
		}
		h = runningHarness(t)
		h.adapters.run.jitter = fixedJitter(time.Hour)
		h.run.runner.fail[1] = true
		c = h.initialized(t)
		runToEnd(t, c, h.admit(t, c, scheduledTask("lim-3", retry(30000, 600000, ""), "", false)))
		if w := h.run.clock.Waits(); len(w) != 1 || w[0] != 40*time.Second {
			t.Fatalf("default retry jitter: waits %v", w)
		}
	})
	t.Run("remote expiry relative to issuance or acceptance", func(t *testing.T) {
		for name, tc := range map[string]struct {
			notAfter time.Duration
			want     string
		}{"one week": {604800 * time.Second, "ok"}, "one week plus one second": {604801 * time.Second, "task_limit_exceeded"}, "missing": {-1, "task_schema_invalid"}} {
			h := runningHarness(t)
			c := h.initialized(t)
			s := `"repeat":{"mode":"none"}`
			if tc.notAfter >= 0 {
				s = `"not_after":"` + at(tc.notAfter) + `"`
			}
			bytes, ref := h.deliver(t, "remote", scheduledTask("lim-4", s, "", false))
			if _, err := c.AdmitTask(bytes, ref); codeOf(err) != tc.want {
				t.Fatalf("%s: %s, want %s", name, codeOf(err), tc.want)
			}
		}
	})
	t.Run("issuance skew", func(t *testing.T) {
		for _, tc := range []struct {
			ahead time.Duration
			want  string
		}{{300 * time.Second, "ok"}, {301 * time.Second, "task_source_unauthenticated"}} {
			h := runningHarness(t)
			c := h.initialized(t)
			pkg := string(remoteSigner().Sign(epoch.Add(tc.ahead), []byte(scheduledTask("lim-5", notAfter, "", false))).Encode())
			if _, err := c.AdmitTask(pkg, h.installPackage(t, pkg)); codeOf(err) != tc.want {
				t.Fatalf("issued %v ahead: %s, want %s", tc.ahead, codeOf(err), tc.want)
			}
		}
	})
	t.Run("no trusted time", func(t *testing.T) {
		h := runningHarness(t)
		h.adapters.clock = noTrustedTime{}
		c := h.initialized(t)
		bytes, ref := h.deliver(t, "remote", scheduledTask("lim-6", notAfter, "", false))
		if _, err := c.AdmitTask(bytes, ref); codeOf(err) != "task_trust_unavailable" {
			t.Fatal(codeOf(err))
		}
	})
	t.Run("embedded one-time exception", func(t *testing.T) {
		one, two := 1, 2
		s := task.Schedule{Repeat: task.Repeat{Mode: "none"}, MaxRuns: &one, Retry: task.Retry{MaxAttempts: 1}}
		ctx := validation.TimingContext{Origin: "embedded", AcceptedAt: epoch, EmbeddedOneTimeException: true}
		if validation.ScheduleAt(s, ctx) != nil {
			t.Fatal("reviewed one-time exception rejected")
		}
		s.MaxRuns = &two
		if validation.ScheduleAt(s, ctx) == nil {
			t.Fatal("exception accepted for two runs")
		}
		interval, jitter := 21600, 300
		s.MaxRuns, s.Repeat = &one, task.Repeat{Mode: "interval", IntervalSeconds: &interval, JitterSeconds: &jitter}
		if validation.ScheduleAt(s, ctx) == nil {
			t.Fatal("exception accepted for a repeating task")
		}
		// A repeating embedded task uses the remote lifetime: one week.
		s.MaxRuns = &two
		ctx.EmbeddedOneTimeException = false
		week := task.TimestampFromTime(epoch.Add(604800 * time.Second))
		over := task.TimestampFromTime(epoch.Add(604801 * time.Second))
		if s.NotAfter = &week; validation.ScheduleAt(s, ctx) != nil {
			t.Fatal("embedded repeating task with a one-week lifetime rejected")
		}
		if s.NotAfter = &over; validation.ScheduleAt(s, ctx) == nil {
			t.Fatal("embedded repeating task beyond one week accepted")
		}
	})
}

// consentHarness requires consent in local policy (the deployment
// permissions), in the issuer permissions, or neither.
func consentHarness(t *testing.T, local, issuer bool) *harness {
	t.Helper()
	return customHarness(t, nil, func(in *task.PolicyInput) {
		in.Permissions.RequireConsent = local
		for i := range in.Issuers {
			in.Issuers[i].Permissions.RequireConsent = issuer
		}
	})
}

// CC-SAFETY-USER-CONSENT: the task, issuer or local consent requirement and the
// host-reported consent status at attempt start.
// requirement-ref: REQ-SAFETY-0020
// requirement-ref: REQ-SAFETY-0021
// requirement-ref: REQ-SAFETY-0022
func TestCcSafetyUserConsent(t *testing.T) {
	notAfter := `"not_after":"` + at(20*time.Hour) + `"`
	statuses := []struct {
		name   string
		status platform.ConsentStatus
		runs   bool
	}{
		{"valid", platform.ConsentValid, true},
		{"absent", platform.ConsentAbsent, false},
		{"expired or withdrawn", platform.ConsentEnded, false},
		{"unavailable", platform.ConsentUnavailable, false},
	}
	requirements := []struct {
		name                 string
		task, local, issuer  bool
		required, runsAnyway bool
	}{
		{"task requires", true, false, false, true, false},
		{"local policy requires, task false", false, true, false, true, false},
		{"issuer requires, task false", false, false, true, true, false},
		{"nothing requires", false, false, false, false, true},
	}
	for _, rq := range requirements {
		for _, st := range statuses {
			t.Run(rq.name+", consent "+st.name, func(t *testing.T) {
				h := consentHarness(t, rq.local, rq.issuer)
				h.run.consent.Set(st.status)
				c := h.initialized(t)
				th := h.admit(t, c, scheduledTask("consent-1", notAfter, "", rq.task))
				if got := admitted(t, c, th).ConsentRequired(); got != rq.required {
					t.Fatalf("recorded consent requirement = %t, want %t", got, rq.required)
				}
				code := startCode(t, c, th)
				runs := rq.runsAnyway || st.runs
				if runs != (code == "ok") {
					t.Fatalf("start = %s, want run=%t", code, runs)
				}
				if !runs && (h.run.runner.Calls() != 0 || len(h.run.output.Records()) != 0) {
					t.Fatal("an attempt or Measurement Result without valid consent")
				}
				if !rq.required && h.run.consent.Calls() != 0 {
					t.Fatal("consent queried although nothing requires it")
				}
			})
		}
	}
	t.Run("no consent capability", func(t *testing.T) {
		h := runningHarness(t)
		h.adapters.run.consent = nil
		c := h.initialized(t)
		if code := startCode(t, c, h.admit(t, c, scheduledTask("consent-2", notAfter, "", true))); code != "run_not_due" || h.run.runner.Calls() != 0 {
			t.Fatal("a consent-requiring task ran without a consent capability:", code)
		}
	})
	t.Run("without a consent requirement, risk and target checks still apply", func(t *testing.T) {
		h := consentHarness(t, false, false)
		c := h.initialized(t)
		high := strings.Replace(scheduledTask("consent-3", notAfter, "", false), `"risk_level":"low"`, `"risk_level":"high"`, 1)
		if _, code := h.tryAdmit(t, c, high); code == "ok" {
			t.Fatal("a high-risk task was admitted without its review")
		}
		h = customHarness(t, func(e *registry.Entry) { e.Targets = otherTarget{} }, nil)
		c = h.initialized(t)
		h.wantAdmission(t, c, "unregistered target", scheduledTask("consent-4", notAfter, "", false), "target_not_registered")
	})
	t.Run("a valid consent status is rechecked at each attempt", func(t *testing.T) {
		h := runningHarness(t)
		h.run.runner.fail[1] = true
		h.run.runner.during = func(n int) {
			if n == 1 {
				h.run.consent.Set(platform.ConsentAbsent)
			}
		}
		c := h.initialized(t)
		runToEnd(t, c, h.admit(t, c, scheduledTask("consent-5", notAfter+","+fixedRetry("internal_error"), "", true)))
		if h.run.runner.Calls() != 1 {
			t.Fatal("a retry ran after consent became absent")
		}
	})
}

// otherTarget names a host no registration covers.
type otherTarget struct{}

func (otherTarget) Targets(task.Parameters) ([]registry.Target, bool) {
	return []registry.Target{{Role: task.MeasurementTarget, Tuple: task.TargetTuple{Hostname: "unregistered.example.test"}}}, true
}

// executionTask is a TCP-01 task whose execution object is exactly execution.
func executionTask(id, execution string) string {
	return fmt.Sprintf(`{"schema_version":"0.1.0","task_id":%q,"test_id":"TCP-01","test_version":"0.1.0","input":{},"options":{},"schedule":{"not_after":%q},"execution":%s,"safety":{"risk_level":"low","requires_user_consent":false}}`,
		id, at(20*time.Hour), execution)
}

// CC-EXECUTION-FIELDS: the strict execution object: required limits with
// registered defaults, the network path, typed optional objects and unknown
// members.
// requirement-ref: REQ-EXECUTION-0005
func TestCcExecutionFields(t *testing.T) {
	h := customHarness(t, func(e *registry.Entry) { e.Proxy = anyProxy{} }, nil)
	c := h.initialized(t)
	th := h.wantAdmission(t, c, "empty limits", executionTask("exec-1", `{"limits":{}}`), "ok")
	a := admitted(t, c, th)
	if got := a.AttemptLimits(); got != (task.AttemptLimits{DurationMS: 30000, BytesSent: 32768, BytesReceived: 131072, Connections: 8}) {
		t.Fatalf("registered defaults not applied: %+v", got)
	}
	if got := a.Candidate().Execution.NetworkPath; got != "direct" {
		t.Fatalf("omitted network_path = %q, want direct", got)
	}
	cases := []struct{ name, execution, want string }{
		{"limits missing", `{}`, "task_schema_invalid"},
		{"network_path direct", `{"limits":{},"network_path":"direct"}`, "ok"},
		{"network_path proxy", `{"limits":{},"network_path":"proxy"}`, "ok"},
		{"network_path vpn", `{"limits":{},"network_path":"vpn"}`, "task_schema_invalid"},
		{"typed network", `{"limits":{},"network":{"allowed_transports":["wifi"]}}`, "ok"},
		{"typed device", `{"limits":{},"device":{"minimum_battery_percent":20}}`, "ok"},
		{"typed concurrency", `{"limits":{},"concurrency":{"mode":"shared"}}`, "ok"},
		{"unknown execution member", `{"limits":{},"priority":1}`, "task_schema_invalid"},
	}
	for _, member := range []string{"network", "device", "concurrency"} {
		for kind, v := range map[string]string{"string": `"x"`, "array": `[]`, "null": `null`} {
			cases = append(cases, struct{ name, execution, want string }{member + " as " + kind, `{"limits":{},"` + member + `":` + v + `}`, "task_schema_invalid"})
		}
		cases = append(cases, struct{ name, execution, want string }{member + " with an unknown member", `{"limits":{},"` + member + `":{"unknown":1}}`, "task_schema_invalid"})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := customHarness(t, func(e *registry.Entry) { e.Proxy = anyProxy{} }, nil)
			c := h.initialized(t)
			th := h.wantAdmission(t, c, tc.name, executionTask("exec-2", tc.execution), tc.want)
			if tc.name == "network_path proxy" && admitted(t, c, th).Candidate().Execution.NetworkPath != "proxy" {
				t.Fatal("network_path proxy not kept")
			}
		})
	}
}

// CC-EXECUTION-FIELDS-CONCURRENCY: shared runs may overlap; an exclusive run
// never overlaps another measurement, and while it cannot run alone it waits
// without starting any network operation.
// requirement-ref: REQ-EXECUTION-0027
func TestCcExecutionFieldsConcurrency(t *testing.T) {
	shared, exclusive := `"concurrency":{"mode":"shared"}`, `"concurrency":{"mode":"exclusive"}`
	notAfter := `"not_after":"` + at(20*time.Hour) + `"`
	type tasks struct{ s1, s2, e string }
	setup := func(t *testing.T) (*harness, *Client, tasks) {
		h := runningHarness(t)
		c := h.initialized(t)
		return h, c, tasks{
			s1: h.admit(t, c, scheduledTask("conc-s1", notAfter, shared, false)),
			s2: h.admit(t, c, scheduledTask("conc-s2", notAfter, shared, false)),
			e:  h.admit(t, c, scheduledTask("conc-e", notAfter, exclusive, false)),
		}
	}
	// refused checks that a start that cannot run now waits: no run, no
	// path, no runner call and no result.
	refused := func(t *testing.T, h *harness, c *Client, th string) {
		t.Helper()
		calls, paths, results := h.run.runner.Calls(), h.run.paths.Calls(), len(h.run.output.Records())
		if _, err := c.StartRun(th); codeOf(err) != "run_not_due" {
			t.Fatal("started beside an exclusive run:", codeOf(err))
		}
		if h.run.runner.Calls() != calls || h.run.paths.Calls() != paths || len(h.run.output.Records()) != results {
			t.Fatal("a waiting start reached the path, the runner or the output")
		}
	}
	t.Run("S1, S2, E", func(t *testing.T) {
		h, c, ts := setup(t)
		h.run.runner.holdCalls()
		r1, err1 := c.StartRun(ts.s1)
		r2, err2 := c.StartRun(ts.s2)
		if err1 != nil || err2 != nil {
			t.Fatal("shared runs:", codeOf(err1), codeOf(err2))
		}
		waitStarted(t, h.run.runner)
		waitStarted(t, h.run.runner) // both shared runs are active at once
		refused(t, h, c, ts.e)
		h.run.runner.releaseCalls()
		waitTerminal(t, c, r1)
		waitTerminal(t, c, r2)
		if code := startCode(t, c, ts.e); code != "ok" {
			t.Fatal("exclusive run alone:", code)
		}
	})
	t.Run("E, S1, S2", func(t *testing.T) {
		h, c, ts := setup(t)
		h.run.runner.holdCalls()
		re, err := c.StartRun(ts.e)
		if err != nil {
			t.Fatal(codeOf(err))
		}
		waitStarted(t, h.run.runner)
		refused(t, h, c, ts.s1)
		refused(t, h, c, ts.s2)
		h.run.runner.releaseCalls()
		waitTerminal(t, c, re)
		for _, th := range []string{ts.s1, ts.s2} {
			if code := startCode(t, c, th); code != "ok" {
				t.Fatal("shared run after the exclusive run:", code)
			}
		}
	})
	t.Run("E submitted while S1 runs", func(t *testing.T) {
		h, c, ts := setup(t)
		h.run.runner.holdCalls()
		r1, err := c.StartRun(ts.s1)
		if err != nil {
			t.Fatal(codeOf(err))
		}
		waitStarted(t, h.run.runner)
		refused(t, h, c, ts.e)
		h.run.runner.releaseCalls()
		waitTerminal(t, c, r1)
		h.run.runner.holdCalls()
		re, err := c.StartRun(ts.e)
		if err != nil {
			t.Fatal(codeOf(err))
		}
		waitStarted(t, h.run.runner)
		refused(t, h, c, ts.s2)
		h.run.runner.releaseCalls()
		waitTerminal(t, c, re)
	})
}
