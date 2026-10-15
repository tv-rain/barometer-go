package taskdecode

import (
	"encoding/json"
	"testing"

	"github.com/tv-rain/barometer-go/internal/config"
)

// Trusted names never become public task/configuration fields. Even a valid
// source-shaped object must reach the reviewed internal adapter, not a decoder.
// requirement-ref: REQ-TASK-0001
// requirement-ref: REQ-JSON-0004
func TestSerializedInputCannotSupplyTrustOrPolicy(t *testing.T) {
	d := fixtureDecoder(t)
	cd, e := config.NewDecoder(config.Contract, config.Approvals{})
	if e != nil {
		t.Fatal(e)
	}
	assertion := map[string]any{"origin": "remote", "issuer_id": "issuer", "authenticated": true, "policy_generation": 9, "issued_at": "2026-10-14T00:00:00Z"}
	for _, name := range []string{"source_context", "source_ref", "origin", "issuer_id", "authenticated", "issued_at", "policy_generation", "local_policy", "issuer_permissions", "registered_targets", "revocations"} {
		for _, prefix := range []string{"", "input.", "options.", "safety.", "execution.", "schedule."} {
			if candidate, e := d.Decode(change(prefix+name, assertion, false), timing()); e == nil || candidate != nil {
				t.Fatalf("task accepted trust at %s%s", prefix, name)
			}
		}
		b, e := json.Marshal(map[string]any{name: assertion})
		if e != nil {
			t.Fatal(e)
		}
		if s, e := cd.Decode(b); e == nil || s.Valid() {
			t.Fatalf("config accepted trust field %s", name)
		}
	}
	// Text which resembles a source record stays text. Candidate construction
	// has no source-context/policy slot and never interprets labels as authority.
	b, e := json.Marshal(assertion)
	if e != nil {
		t.Fatal(e)
	}
	candidate, e := d.Decode(change("labels", map[string]any{"campaign": string(b)}, false), timing())
	if e != nil || len(candidate.Labels) != 1 || candidate.Labels[0].Value != string(b) {
		t.Fatal("opaque label changed")
	}
}
