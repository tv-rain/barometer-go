package taskdecode

import (
	"testing"

	"github.com/tv-rain/barometer-go/internal/config"
)

// Host configuration and trusted capability names do not become task fields.
// requirement-ref: REQ-TASK-0001
func TestHostConfigurationCannotComeFromTask(t *testing.T) {
	d := fixtureDecoder(t)
	for _, name := range []string{"proxy", "network_context", "result_linkage", "local_policy", "policy_generation", "authenticated", "clock", "linkage_keys"} {
		candidate, err := d.Decode(change(name, map[string]any{"enabled": true}, false), timing())
		if err == nil || candidate != nil {
			t.Fatalf("task populated configuration field %s", name)
		}
	}
	// An unreserved label remains opaque text even if it looks like host JSON.
	const hostJSON = `{"result_linkage":{"mode":"rotating_installation","scope":"unreviewed","rotation_period_seconds":86400}}`
	candidate, err := d.Decode(change("labels", map[string]any{"host_config": hostJSON}, false), timing())
	if err != nil || len(candidate.Labels) != 1 || candidate.Labels[0].Value != hostJSON {
		t.Fatal("opaque label handling changed")
	}
	configuration, err := config.NewDecoder(config.Contract, config.Approvals{})
	if err != nil {
		t.Fatal(err)
	}
	if s, err := configuration.Decode([]byte(candidate.Labels[0].Value)); err == nil || s.Valid() {
		t.Fatal("task label acquired deployment approval")
	}
}
