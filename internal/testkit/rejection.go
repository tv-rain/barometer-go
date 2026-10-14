package testkit

import "testing"

// Reporter is the small testing interface used by rejection assertions. Payloads
// and raw errors are deliberately excluded from failure diagnostics.
type Reporter interface {
	Helper()
	Errorf(string, ...any)
}

// AssertRejected checks both rejection and the consumer's expected error class.
// A caller may use errors.Is/As in matches; matching raw sensitive strings is not
// necessary. Missing metadata, matcher, or oversized fixtures fail the test.
func AssertRejected(t Reporter, c RejectionCase, err error, matches func(string, error) bool) bool {
	t.Helper()
	if c.Name == "" || c.Requirement == "" || c.Expected == "" || matches == nil || len(c.Input) > MaxFixtureBytes {
		t.Errorf("testkit: invalid rejection case metadata, matcher, or fixture bounds")
		return false
	}
	if err == nil {
		t.Errorf("%s: requirement %s: expected rejection %s; input accepted", c.Name, c.Requirement, c.Expected)
		return false
	}
	if !matches(c.Expected, err) {
		t.Errorf("%s: requirement %s: expected rejection %s; different error class", c.Name, c.Requirement, c.Expected)
		return false
	}
	return true
}

// RunRejections gives model/protocol tests a common runner while leaving decoding
// and error classification with their owner. Each invocation receives a copy.
func RunRejections(t *testing.T, cases []RejectionCase, reject func([]byte) error, matches func(string, error) bool) {
	t.Helper()
	if reject == nil {
		t.Fatal("testkit: missing rejection function")
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Name == "" || c.Requirement == "" || c.Expected == "" || matches == nil || len(c.Input) > MaxFixtureBytes {
				t.Fatal("testkit: invalid rejection case metadata, matcher, or fixture bounds")
			}
			AssertRejected(t, c, reject(append([]byte(nil), c.Input...)), matches)
		})
	}
}
