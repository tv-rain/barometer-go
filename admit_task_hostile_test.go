package barometer

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/admission"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/testkit/delivery"
)

// esc spells r as a JSON escape sequence, so tests can vary the encoded form
// of a name or value without changing its decoded text.
func esc(r rune) string { return `\` + fmt.Sprintf("u%04x", r) }

// admissionCodes are the public admission classes (REQ-TRUST-0033); a facade
// rejection of task input is exactly one of them.
var admissionCodes = map[string]bool{
	"task_source_unauthenticated": true, "task_issuer_not_authorized": true, "task_policy_unavailable": true,
	"task_trust_unavailable": true, "task_expired": true, "task_id_conflict": true, "policy_rollback": true,
	"target_not_registered": true, "task_limit_exceeded": true, "task_schema_invalid": true, "task_unsupported": true,
}

// wantRejected checks that err is exactly one fixed, unwrapped admission class
// and that the call left no admission state: no task handle, no replay record
// and no run. The client has no result, bundle or transport state to inspect;
// network absence is established by TestAdmissionHasNoNetworkDependency.
func wantRejected(t *testing.T, h *harness, c *Client, th string, err error) {
	t.Helper()
	ae, ok := errors.AsType[task.AdmissionError](err)
	if !ok || !admissionCodes[err.Error()] || ae.Error() != err.Error() || errors.Unwrap(err) != nil {
		t.Fatalf("rejection is not exactly one public admission class: %s", codeOf(err))
	}
	if th != "" || taskHandles(c) != 0 || h.replay.Len() != 0 || h.replay.Commits() != 0 || len(c.rt.runs.entries) != 0 {
		t.Fatal("rejection left a handle, a replay record or a run")
	}
}

// The strict JSON rules apply at the public boundary for both reviewed
// origins (GO-SPEC-ISSUE-001 admission adoption, DEV-566): names compare after
// decoding escapes, duplicate names are rejected in nested objects, integers
// are exact and the complete-input limit is counted on original bytes.
func TestAdmitTaskStrictJSON(t *testing.T) {
	valid := synTask("synthetic-task", time.Hour)
	replace := func(old, new string) string {
		if !strings.Contains(valid, old) {
			t.Fatal("fixture edit does not apply:", old)
		}
		return strings.Replace(valid, old, new, 1)
	}
	cases := []struct {
		name, task, want string
	}{
		{"duplicate top-level member", valid[:len(valid)-1] + `,"task_id":"synthetic-task"}`, "task_schema_invalid"},
		{"escaped duplicate task_id", replace(`"task_id":"synthetic-task"`, `"task_id":"synthetic-task","task_`+esc('i')+`d":"synthetic-task"`), "task_schema_invalid"},
		{"escaped unknown top-level name", replace(`"options":{}`, `"options":{},"`+esc('p')+`rivate":1`), "task_schema_invalid"},
		{"authority claim field", replace(`"options":{}`, `"options":{},"authenticated":true`), "task_schema_invalid"},
		{"unknown input member", replace(`"input":{}`, `"input":{"extra":1}`), "task_schema_invalid"},
		{"duplicate input member", replace(`"input":{}`, `"input":{"extra":1,"extra":1}`), "task_schema_invalid"},
		{"escaped duplicate schedule member", replace(`"schedule":{`, `"schedule":{"not_`+esc('a')+`fter":"2026-10-20T13:00:00Z",`), "task_schema_invalid"},
		{"duplicate limit member", replace(`"limits":{}`, `"limits":{"max_connections":1,"max_connections":1}`), "task_schema_invalid"},
		{"non-exact schema_version", replace(`"schema_version":"0.1.0"`, `"schema_version":"0.1.00"`), "task_schema_invalid"},
		{"negative zero integer", replace(`"limits":{}`, `"limits":{"max_connections":-0}`), "task_schema_invalid"},
		{"decimal integer spelling", replace(`"limits":{}`, `"limits":{"max_connections":1.0}`), "task_schema_invalid"},
		{"exponent integer spelling", replace(`"limits":{}`, `"limits":{"max_connections":1e0}`), "task_schema_invalid"},
		{"integer beyond exact float range", replace(`"limits":{}`, `"limits":{"max_connections":9007199254740993}`), "task_limit_exceeded"},
		{"trailing value", valid + ` {}`, "task_schema_invalid"},
		{"byte-order mark", "\xef\xbb\xbf" + valid, "task_schema_invalid"},
		{"invalid UTF-8", replace(`"options":{}`, "\"options\":{},\"x\xff\":1"), "task_schema_invalid"},
		{"nesting beyond the parse depth", replace(`"options":{}`, `"options":{},"private":`+strings.Repeat("[", 200)+strings.Repeat("]", 200)), "task_schema_invalid"},
		{"escaped bytes over the complete-input limit", replace(`"options":{}`, `"options":{},"private":"`+strings.Repeat(esc('a'), admission.MaxTaskBytes/6+1)+`"`), "task_limit_exceeded"},
	}
	for _, origin := range []string{"embedded", "remote"} {
		for _, tc := range cases {
			t.Run(origin+"/"+tc.name, func(t *testing.T) {
				h := admittingHarness(t)
				c := h.initialized(t)
				th, err := c.AdmitTask(h.deliver(t, origin, tc.task))
				if codeOf(err) != tc.want {
					t.Fatalf("got %s, want %s", codeOf(err), tc.want)
				}
				wantRejected(t, h, c, th, err)
			})
		}
	}
	t.Run("escaped spelling of the exact version is the same version", func(t *testing.T) {
		h := admittingHarness(t)
		c := h.initialized(t)
		b := replace(`"schema_version":"0.1.0"`, `"schema_version":"0.1.`+esc('0')+`"`)
		if th, err := c.AdmitTask(b, h.embedded(t, b)); err != nil || th == "" {
			t.Fatal("escaped exact version rejected:", codeOf(err))
		}
	})
}

// The strict JSON rules apply to the signed package envelope at the public
// boundary (GO-SPEC-ISSUE-001, DEV-566). The host installs the package bytes
// and presents them with their reference; every malformed envelope is
// task_source_unauthenticated before the task is parsed, and the package
// limit is counted on the original escaped bytes. Escaped spellings of known
// names and of member values decode to the same package and admit.
func TestAdmitTaskSignedPackageStrictJSON(t *testing.T) {
	h0 := admittingHarness(t)
	taskJSON := synTask("synthetic-task", time.Hour)
	valid := h0.sign(taskJSON)
	replace := func(old, new string) string {
		if !strings.Contains(valid, old) {
			t.Fatal("fixture edit does not apply:", old)
		}
		return strings.Replace(valid, old, new, 1)
	}
	// escapeAll spells every byte of an ASCII string as a JSON escape.
	escapeAll := func(s string) string {
		var b strings.Builder
		for _, r := range s {
			b.WriteString(esc(r))
		}
		return b.String()
	}
	encodedTask := base64.RawURLEncoding.EncodeToString([]byte(taskJSON))
	escapedTask := replace(`"task":"`+encodedTask+`"`, `"task":"`+escapeAll(encodedTask)+`"`)
	pad := func(pkg string, total int) string { return pkg + strings.Repeat(" ", total-len(pkg)) }

	cases := []struct {
		name, pkg, want string
	}{
		{"equal duplicate member", replace(`{`, `{"key_id":"remote-key-1",`), "task_source_unauthenticated"},
		{"conflicting duplicate member", replace(`{`, `{"key_id":"other-key",`), "task_source_unauthenticated"},
		{"escaped duplicate member", replace(`{`, `{"key_`+esc('i')+`d":"remote-key-1",`), "task_source_unauthenticated"},
		{"escaped duplicate task member", replace(`{`, `{"t`+esc('a')+`sk":"`+encodedTask+`",`), "task_source_unauthenticated"},
		{"unknown member", replace(`{`, `{"algorithm":"ed25519",`), "task_source_unauthenticated"},
		{"escaped unknown member", replace(`{`, `{"`+esc('a')+`uthenticated":"true",`), "task_source_unauthenticated"},
		{"case variant of a member name", replace(`"key_id"`, `"Key_id"`), "task_source_unauthenticated"},
		{"member as a nested object", replace(`"key_id":"remote-key-1"`, `"key_id":{"id":"remote-key-1"}`), "task_source_unauthenticated"},
		{"nesting beyond the package depth", replace(`"key_id":"remote-key-1"`, `"key_id":`+strings.Repeat("[", 200)+strings.Repeat("]", 200)), "task_source_unauthenticated"},
		{"trailing value", valid + ` {}`, "task_source_unauthenticated"},
		{"byte-order mark", "\xef\xbb\xbf" + valid, "task_source_unauthenticated"},
		{"invalid UTF-8 in a member name", replace(`{`, "{\"x\xff\":\"1\","), "task_source_unauthenticated"},
		{"invalid UTF-8 in a member value", replace(`"remote-key-1"`, "\"remote-key-1\xff\""), "task_source_unauthenticated"},
		{"lone surrogate escape", replace(`"remote-key-1"`, `"remote-key-1\ud800"`), "task_source_unauthenticated"},
		{"escaped package one byte over the limit", pad(escapedTask, admission.MaxPackageBytes+1), "task_limit_exceeded"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := admittingHarness(t)
			c := h.initialized(t)
			th, err := c.AdmitTask(tc.pkg, h.installPackage(t, tc.pkg))
			if codeOf(err) != tc.want {
				t.Fatalf("got %s, want %s", codeOf(err), tc.want)
			}
			wantRejected(t, h, c, th, err)
		})
	}

	accepted := []struct{ name, pkg string }{
		{"escaped spelling of a member name", replace(`"key_id"`, `"key_`+esc('i')+`d"`)},
		{"escaped spelling of a member value", replace(`"remote-key-1"`, `"remote`+esc('-')+`key-1"`)},
		{"escaped task member padded to exactly the limit", pad(escapedTask, admission.MaxPackageBytes)},
	}
	for _, tc := range accepted {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.pkg) > admission.MaxPackageBytes {
				t.Fatal("positive control exceeds the package limit")
			}
			h := admittingHarness(t)
			c := h.initialized(t)
			if th, err := c.AdmitTask(tc.pkg, h.installPackage(t, tc.pkg)); err != nil || th == "" {
				t.Fatal("equivalent package rejected:", codeOf(err))
			}
		})
	}
}

// Hostile source contexts presented through the facade never admit, and each
// fails at step 1 or 2 before the task bytes are parsed. For remote origin the
// host passes only package bytes; nothing it asserts authenticates a task.
// requirement-ref: REQ-TRUST-0042
// requirement-ref: REQ-TRUST-0052
func TestAdmitTaskHostileSources(t *testing.T) {
	valid := synTask("synthetic-task", time.Hour)
	ctx := context.Background()
	embedded := func(t *testing.T, h *harness, ref string) (string, string) { return valid, ref }
	cases := []struct {
		name    string
		deliver func(t *testing.T, h *harness) (string, string)
	}{
		{"empty reference", func(t *testing.T, h *harness) (string, string) { return embedded(t, h, "") }},
		{"forged reference", func(t *testing.T, h *harness) (string, string) { return embedded(t, h, strings.Repeat("0", 64)) }},
		{"reference from another instance", func(t *testing.T, h *harness) (string, string) {
			other, err := admission.NewDeliveryStore(h.clock, admission.DeliveryLimits{Capacity: 1, Lifetime: time.Minute}, 2)
			if err != nil {
				t.Fatal(err)
			}
			ref, err := other.InstallEmbedded(ctx, "embedded-issuer", 2, []byte(valid))
			if err != nil {
				t.Fatal(err)
			}
			return valid, ref
		}},
		{"reference for other bytes", func(t *testing.T, h *harness) (string, string) {
			return embedded(t, h, h.embedded(t, synTask("other-task", time.Hour)))
		}},
		{"remote task without a signed package", func(t *testing.T, h *harness) (string, string) {
			return valid, h.installPackage(t, valid)
		}},
		{"remote package with an altered task byte", func(t *testing.T, h *harness) (string, string) {
			p := remoteSigner().Sign(h.clock.Clock.Now(), []byte(valid))
			p.Task[len(p.Task)-2] = ' '
			pkg := string(p.Encode())
			return pkg, h.installPackage(t, pkg)
		}},
		{"remote package signed by an unknown key", func(t *testing.T, h *harness) (string, string) {
			pkg := string(delivery.NewSigner("unknown-key", "remote-issuer", 9).Sign(h.clock.Clock.Now(), []byte(valid)).Encode())
			return pkg, h.installPackage(t, pkg)
		}},
		{"remote package moved to another issuer", func(t *testing.T, h *harness) (string, string) {
			s := embeddedSigner()
			p := s.Sign(h.clock.Clock.Now(), []byte(valid))
			p.IssuerID = "remote-issuer"
			pkg := string(s.Resign(p).Encode())
			return pkg, h.installPackage(t, pkg)
		}},
		{"remote package issued 301 seconds ahead", func(t *testing.T, h *harness) (string, string) {
			pkg := string(remoteSigner().Sign(h.clock.Clock.Now().Add(admission.MaxIssuedAhead+time.Second), []byte(valid)).Encode())
			return pkg, h.installPackage(t, pkg)
		}},
		{"remote reference presented again", func(t *testing.T, h *harness) (string, string) {
			pkg := h.sign(valid)
			ref := h.installPackage(t, pkg)
			if _, err := h.deliveries.Resolve(ctx, ref, []byte(pkg)); err != nil {
				t.Fatal(err)
			}
			return pkg, ref
		}},
		{"expired delivery record", func(t *testing.T, h *harness) (string, string) {
			ref := h.embedded(t, valid)
			h.clock.Advance(time.Minute)
			return valid, ref
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := admittingHarness(t)
			c := h.initialized(t)
			th, err := c.AdmitTask(tc.deliver(t, h))
			if codeOf(err) != "task_source_unauthenticated" {
				t.Fatal("hostile source must be task_source_unauthenticated, got", codeOf(err))
			}
			wantRejected(t, h, c, th, err)
		})
	}
}

// FuzzAdmitTask drives the public facade with arbitrary task text and a
// fuzzed source: an embedded record for the exact text, an embedded record for
// other bytes, a remote signed package with a fuzzed signature and issuance
// time, a forged reference, and a record from another instance. Every call
// either admits, publishing exactly one handle and one replay record, or
// returns exactly one public admission class and leaves no admission state.
// Results are deterministic, exact redelivery returns the same handle, and
// only an exact-bytes record of a reviewed origin, or a validly signed package,
// can admit. A remote reference is single-use, so each remote presentation
// installs its own reference for the same package bytes.
func FuzzAdmitTask(f *testing.F) {
	valid := synTask("synthetic-task", time.Hour)
	for _, seed := range []string{
		valid,
		synTask("synthetic-task", -time.Second),
		strings.Replace(valid, `"task_id":"synthetic-task"`, `"task_id":"synthetic-task","task_`+esc('i')+`d":"synthetic-task"`, 1),
		strings.Replace(valid, `"limits":{}`, `"limits":{"max_connections":1e0}`, 1),
		strings.Replace(valid, `"limits":{}`, `"limits":{"max_duration_ms":60001}`, 1),
		strings.Replace(valid, `"SYN-01"`, `"SYN-02"`, 1),
		strings.Replace(valid, `"options":{}`, `"options":{},"authenticated":true`, 1),
		strings.Replace(valid, `"input":{}`, `"input":{"extra":1}`, 1),
		`{`, `[]`, `null`, `{"schema_version":"0.1.0"}`, "",
	} {
		for source := range uint8(5) {
			f.Add(seed, source, "forged-reference", true, int16(0))
		}
	}
	f.Add(valid, uint8(2), "", false, int16(0))
	f.Add(valid, uint8(2), "", true, int16(301))
	f.Add(valid, uint8(2), "", true, int16(300))
	f.Fuzz(func(t *testing.T, taskJSON string, source uint8, forged string, signedOK bool, issuedAhead int16) {
		h := admittingHarness(t)
		// Queued calls must not hit the short harness deadline, which would
		// fail closed as task_trust_unavailable and break the determinism check.
		h.adapters.deadline = defaultLifecycleDeadline
		c := h.initialized(t)
		ctx := context.Background()
		presented := taskJSON
		var ref string
		present := func() string { return ref }
		admissible := false
		switch source % 5 {
		case 0:
			if r, err := h.deliveries.InstallEmbedded(ctx, "embedded-issuer", 2, []byte(taskJSON)); err == nil {
				ref, admissible = r, true
			}
		case 1:
			other := synTask("other-task", time.Hour)
			if r, err := h.deliveries.InstallEmbedded(ctx, "embedded-issuer", 2, []byte(other)); err == nil {
				ref, admissible = r, taskJSON == other
			}
		case 2:
			p := remoteSigner().Sign(h.clock.Clock.Now().Add(time.Duration(issuedAhead)*time.Second), []byte(taskJSON))
			if !signedOK {
				p.Signature[0] ^= 1
			}
			presented = string(p.Encode())
			present = func() string {
				r, err := h.deliveries.InstallRemote(ctx, 2, []byte(presented))
				if err != nil {
					t.Fatal("remote install rejected")
				}
				return r
			}
			admissible = signedOK && taskJSON != "" && time.Duration(issuedAhead)*time.Second <= admission.MaxIssuedAhead
		case 3:
			ref = forged
		case 4:
			other, err := admission.NewDeliveryStore(h.clock, admission.DeliveryLimits{Capacity: 1, Lifetime: time.Minute}, 2)
			if err != nil {
				t.Fatal(err)
			}
			if ref, err = other.InstallEmbedded(ctx, "embedded-issuer", 2, []byte(taskJSON)); err != nil {
				ref = forged
			}
		}
		th, err := c.AdmitTask(presented, present())
		again, err2 := c.AdmitTask(presented, present())
		if err != nil {
			if codeOf(err2) != codeOf(err) {
				t.Fatalf("nondeterministic rejection: %s then %s", codeOf(err), codeOf(err2))
			}
			wantRejected(t, h, c, th, err)
			if again != "" {
				t.Fatal("a repeated rejection published a handle")
			}
			return
		}
		if !admissible {
			t.Fatal("a task was admitted without an exact-bytes record of a reviewed origin or a valid signature")
		}
		if !strings.HasPrefix(th, taskHandlePrefix) || err2 != nil || again != th {
			t.Fatal("exact redelivery did not return the same handle:", codeOf(err2))
		}
		if h.replay.Commits() != 1 || h.replay.Len() != 1 || taskHandles(c) != 1 || len(c.rt.runs.entries) != 0 {
			t.Fatal("acceptance must commit one replay record, publish one handle and allocate no run")
		}
	})
}
