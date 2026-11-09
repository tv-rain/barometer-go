package admission_test

import (
	"bytes"
	"testing"

	"github.com/tv-rain/barometer-go/internal/model/task"
)

// The embedded one-time exception of REQ-LIMIT-0013 holds only for a task
// that the active local policy lists as release-reviewed by issuer and the
// digest of its exact bytes. Every other omission of not_after, including a
// near-miss listing, is task_schema_invalid, and a listing never relaxes the
// one-run, no-repeat or remote rules.
// requirement-ref: REQ-LIMIT-0013
func TestGateReleaseReviewedOneTimeException(t *testing.T) {
	oneTime := taskWith(map[string]any{"schedule.not_after": nil})
	twoRuns := taskWith(map[string]any{"schedule.not_after": nil, "schedule.limits.max_runs": 2})
	repeating := taskWith(map[string]any{"schedule.not_after": nil, "schedule.repeat": map[string]any{"mode": "interval", "interval": 21600, "jitter": 300}, "schedule.limits.max_runs": 1})
	// Semantically equal JSON with other bytes is another task.
	spaced := bytes.Replace(oneTime, []byte(`"schema_version":`), []byte(`"schema_version": `), 1)
	if bytes.Equal(spaced, oneTime) {
		t.Fatal("fixture unchanged")
	}
	cases := []struct {
		name    string
		policy  task.PolicyInput
		present []byte
		remote  bool
		want    task.AdmissionCode
	}{
		{"listed", reviewed(permissive(), "embedded-issuer", oneTime), oneTime, false, 0},
		{"no list", permissive(), oneTime, false, task.SchemaInvalid},
		{"other task listed", reviewed(permissive(), "embedded-issuer", twoRuns), oneTime, false, task.SchemaInvalid},
		{"other bytes of the same task", reviewed(permissive(), "embedded-issuer", oneTime), spaced, false, task.SchemaInvalid},
		{"listed for another issuer", reviewed(permissive(), "remote-issuer", oneTime), oneTime, false, task.SchemaInvalid},
		{"listed with two runs", reviewed(permissive(), "embedded-issuer", twoRuns), twoRuns, false, task.SchemaInvalid},
		{"listed and repeating", reviewed(permissive(), "embedded-issuer", repeating), repeating, false, task.SchemaInvalid},
		{"remote task listed", reviewed(permissive(), "remote-issuer", oneTime), oneTime, true, task.SchemaInvalid},
	}
	for _, c := range cases {
		h := newHarnessWith(t, mustPolicy(t, c.policy))
		var err error
		if c.remote {
			_, err = h.remote(t, 0, c.present)
		} else {
			_, err = h.embedded(t, c.present)
		}
		expect(t, c.name, err, c.want)
		if c.want != 0 && h.normalize.Load() != 0 {
			t.Errorf("%s: codec reached after step 6 failed", c.name)
		}
	}
	// A task with not_after needs no listing.
	h := newHarness(t)
	_, err := h.embedded(t, taskWith(nil))
	expect(t, "with not_after", err, 0)
}

// A release-review entry is trusted policy: one naming an issuer the policy
// does not authorize, a zero digest, a duplicate or too many entries reject
// the whole policy, and the snapshot keeps no alias of the input.
func TestPolicyReleaseReviewEntriesAreValidated(t *testing.T) {
	digest := [32]byte{1}
	for name, list := range map[string][]task.ReleaseReviewedTask{
		"unknown issuer": {{IssuerID: "other-issuer", TaskSHA256: digest}},
		"zero digest":    {{IssuerID: "embedded-issuer"}},
		"duplicate":      {{IssuerID: "embedded-issuer", TaskSHA256: digest}, {IssuerID: "embedded-issuer", TaskSHA256: digest}},
		"too many": func() []task.ReleaseReviewedTask {
			var l []task.ReleaseReviewedTask
			for i := 0; i <= task.MaxReleaseReviewedTasks; i++ {
				l = append(l, task.ReleaseReviewedTask{IssuerID: "embedded-issuer", TaskSHA256: [32]byte{byte(i), byte(i >> 8), 1}})
			}
			return l
		}(),
	} {
		in := permissive()
		in.ReleaseReviewed = list
		if p, err := task.NewLocalPolicy(in); err == nil || p.Valid() {
			t.Errorf("%s: policy accepted", name)
		}
	}
	in := permissive()
	in.ReleaseReviewed = []task.ReleaseReviewedTask{{IssuerID: "embedded-issuer", TaskSHA256: digest}}
	p := mustPolicy(t, in)
	in.ReleaseReviewed[0].TaskSHA256 = [32]byte{2}
	if !p.ReleaseReviewed("embedded-issuer", digest) || p.ReleaseReviewed("embedded-issuer", [32]byte{2}) || p.ReleaseReviewed("remote-issuer", digest) {
		t.Fatal("release-review lookup is not exact or aliases its input")
	}
	if (task.LocalPolicy{}).ReleaseReviewed("embedded-issuer", digest) {
		t.Fatal("an invalid policy lists a task")
	}
}
