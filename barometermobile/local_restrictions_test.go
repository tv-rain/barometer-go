package barometermobile

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
)

// withPolicyMembers adds members to the policy of a trusted local document.
func withPolicyMembers(t *testing.T, trust []byte, members map[string]any) []byte {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal(trust, &doc); err != nil {
		t.Fatal(err)
	}
	policy := doc["policy"].(map[string]any)
	for k, v := range members {
		policy[k] = v
	}
	b, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// The release-review list reaches admission only through the trusted local
// document: the embedded one-time task it lists by exact bytes is admitted
// without not_after, and the same task under a document without the list is
// task_schema_invalid.
func TestTrustDocumentReleaseReviewThroughFacade(t *testing.T) {
	oneTime := `{"schema_version":"0.1.0","task_id":"reviewed-1","test_id":"TCP-01","test_version":"0.1.0","input":{"ip_address":"8.8.8.8","port":443},"options":{},"schedule":{"repeat":{"mode":"none"},"limits":{"max_runs":1}},"execution":{"limits":{}},"safety":{"risk_level":"low","requires_user_consent":false}}`
	digest := sha256.Sum256([]byte(oneTime))
	listed := withPolicyMembers(t, trustDocument("TCP-01", tcpTarget), map[string]any{
		"release_reviewed_tasks": []any{map[string]any{"issuer_id": "embedded-issuer", "task_sha256": hex.EncodeToString(digest[:])}},
	})
	for name, tc := range map[string]struct {
		trust []byte
		want  string
	}{
		"listed":   {listed, ""},
		"unlisted": {trustDocument("TCP-01", tcpTarget), "task_schema_invalid"},
	} {
		t.Run(name, func(t *testing.T) {
			c := newTestClient(t, newNatives(), tc.trust, options{directDialers: loopbackDialer{want: "none"}})
			if err := c.Initialize(`{}`); err != nil {
				t.Fatal(code(err))
			}
			ref, err := c.InstallEmbeddedTask("embedded-issuer", []byte(oneTime))
			if err != nil {
				t.Fatal(code(err))
			}
			if _, err := c.AdmitTask(oneTime, ref); code(err) != tc.want {
				t.Fatalf("admission = %q, want %q", code(err), tc.want)
			}
		})
	}
	// An entry for an issuer the document does not authorize rejects the
	// whole document.
	bad := withPolicyMembers(t, trustDocument("TCP-01", tcpTarget), map[string]any{
		"release_reviewed_tasks": []any{map[string]any{"issuer_id": "other-issuer", "task_sha256": hex.EncodeToString(digest[:])}},
	})
	if _, err := newClient(newNatives().adapters(), bad, options{platform: "android"}); err == nil {
		t.Fatal("a document listing an unauthorized issuer was accepted")
	}
}

// Local device constraints in the trusted local document gate a run that
// states no device constraint of its own. The native device reports 80
// percent while charging.
func TestTrustDocumentDeviceConstraintsThroughFacade(t *testing.T) {
	for _, tc := range []struct {
		minimum int
		want    string
	}{{81, "run_not_due"}, {80, ""}} {
		trust := withPolicyMembers(t, trustDocument("TCP-01", tcpTarget), map[string]any{
			"device": map[string]any{"minimum_battery_percent": tc.minimum, "charging_only": true},
		})
		c := newTestClient(t, newNatives(), trust, options{directDialers: loopbackDialer{want: "none"}})
		if err := c.Initialize(`{}`); err != nil {
			t.Fatal(code(err))
		}
		data := runTask("device-1", "TCP-01", `{"ip_address":"8.8.8.8","port":443}`, `{}`)
		ref, err := c.InstallEmbeddedTask("embedded-issuer", []byte(data))
		if err != nil {
			t.Fatal(code(err))
		}
		th, err := c.AdmitTask(data, ref)
		if err != nil {
			t.Fatal(code(err))
		}
		rh, err := c.StartRun(th)
		if code(err) != tc.want {
			t.Fatalf("local minimum %d: start = %q, want %q", tc.minimum, code(err), tc.want)
		}
		if err == nil {
			waitTerminal(t, c, rh)
		}
	}
}
