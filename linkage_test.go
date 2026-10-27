package barometer

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/config"
	"github.com/tv-rain/barometer-go/internal/persistence"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/testkit/capability"
)

// Synthetic facade cases for rotating result linkage (DEV-675) over the
// production pending store, a test-scoped linkage key store and the
// fault-injectable record port. The key store is not protected storage. No
// runtime, platform, privacy or deployment evidence; no deployment enables
// linkage.

const linkageScope = "org.example.production"

// linkedHarness is a running harness whose trusted approvals admit the
// synthetic linkage scope and whose key store holds a fixed synthetic key.
func linkedHarness(t *testing.T, allow bool) (*harness, *capability.LinkageKeys) {
	t.Helper()
	h := runningHarness(t)
	keys := capability.NewLinkageKeys([32]byte{7})
	h.adapters.linkageKeys = keys
	h.adapters.approvals = config.Approvals{LinkageScopes: []string{linkageScope}, AllowLinkage: allow}
	return h, keys
}

func rotatingConfig(period int) string {
	return fmt.Sprintf(`{"result_linkage":{"mode":"rotating_installation","scope":%q,"rotation_period_seconds":%d}}`, linkageScope, period)
}

type bundleLinkage struct {
	Mode  string `json:"mode"`
	Scope string `json:"scope"`
	Epoch int64  `json:"epoch"`
	Token string `json:"token"`
}

// linkageOf returns the linkage member of the next pending bundle, or nil.
func linkageOf(t *testing.T, c *Client) *bundleLinkage {
	t.Helper()
	b, err := c.NextBundle()
	if err != nil || b == "" {
		t.Fatal("no bundle:", codeOf(err))
	}
	var w struct {
		Linkage *bundleLinkage `json:"linkage"`
	}
	if err := json.Unmarshal([]byte(b), &w); err != nil {
		t.Fatal("not a bundle")
	}
	if w.Linkage == nil && bytes.Contains([]byte(b), []byte("linkage")) {
		t.Fatal("malformed linkage member")
	}
	return w.Linkage
}

// The linkage rows of CC-LIMIT-CONTEXT-LINKAGE-AND-RETENTION-LIMITS:
// rotation periods of 86,399 and 604,801 seconds are rejected before any key
// operation, 86,400 and 604,800 are accepted and their bundles carry the
// epoch of the trusted clock under that period, and absent configuration
// yields no linkage even with a key store and permissive policy bound.
// requirement-ref: REQ-LIMIT-0023
// requirement-ref: REQ-LINKAGE-0003
// requirement-ref: REQ-LINKAGE-0004
// requirement-ref: REQ-LINKAGE-0011
// requirement-ref: REQ-LINKAGE-0017
// requirement-ref: REQ-LINKAGE-0043
func TestCcLinkageRotationLimits(t *testing.T) {
	for _, period := range []int{86399, 604801} {
		h, keys := linkedHarness(t, true)
		c := h.client(t)
		if codeOf(c.Initialize(rotatingConfig(period))) != "not_initialized" || len(keys.Messages()) != 0 {
			t.Fatalf("rotation period %d accepted or reached the key store", period)
		}
	}
	for _, period := range []int{86400, 604800} {
		h, keys := linkedHarness(t, true)
		c := h.client(t)
		if err := c.Initialize(rotatingConfig(period)); err != nil {
			t.Fatalf("rotation period %d rejected: %v", period, codeOf(err))
		}
		completedRun(t, h, c, fmt.Sprintf("period-%d", period))
		l := linkageOf(t, c)
		if l == nil || l.Mode != "rotating_installation" || l.Scope != linkageScope || l.Epoch != epoch.Unix()/int64(period) || len(l.Token) != 43 || len(keys.Messages()) != 1 {
			t.Fatalf("rotation period %d: linkage %+v", period, l)
		}
	}
	for name, cfg := range map[string]string{"absent": `{}`, "mode none": `{"result_linkage":{"mode":"none"}}`} {
		h, keys := linkedHarness(t, true)
		c := h.client(t)
		if err := c.Initialize(cfg); err != nil {
			t.Fatal(err)
		}
		completedRun(t, h, c, "unlinked")
		if linkageOf(t, c) != nil || len(keys.Messages()) != 0 {
			t.Fatalf("%s configuration produced linkage", name)
		}
	}
	// Local policy that disallows linkage normalizes a request to none.
	h, keys := linkedHarness(t, false)
	c := h.client(t)
	if err := c.Initialize(rotatingConfig(604800)); err != nil {
		t.Fatal(err)
	}
	completedRun(t, h, c, "disallowed")
	if linkageOf(t, c) != nil || len(keys.Messages()) != 0 {
		t.Fatal("disallowed linkage produced linkage")
	}
}

// Clearing deletes the linkage key and erases the linkage epoch state and
// creates no replacement; a later bundle derives under a new key, so its
// token differs in the same epoch.
// requirement-ref: REQ-LINKAGE-0023
// requirement-ref: REQ-LINKAGE-0024
func TestClearRemovesLinkage(t *testing.T) {
	h, keys := linkedHarness(t, true)
	c := h.client(t)
	if err := c.Initialize(rotatingConfig(604800)); err != nil {
		t.Fatal(err)
	}
	completedRun(t, h, c, "linked-1")
	before := linkageOf(t, c)
	if before == nil {
		t.Fatal("no linkage before clearing")
	}
	if err := c.ClearMeasurementData(); err != nil {
		t.Fatal("clear failed:", codeOf(err))
	}
	if held, created, deletes := keys.Held(); held || created != 0 || deletes != 1 {
		t.Fatal("clearing left the key or created a replacement")
	}
	if err := c.Initialize(rotatingConfig(604800)); err != nil {
		t.Fatal(err)
	}
	if held, created, _ := keys.Held(); held || created != 0 {
		t.Fatal("initialization created a secret")
	}
	completedRun(t, h, c, "linked-2")
	after := linkageOf(t, c)
	if after == nil || after.Epoch != before.Epoch || after.Token == before.Token {
		t.Fatal("linkage after clearing reused the deleted key")
	}
	if held, created, _ := keys.Held(); !held || created != 1 {
		t.Fatal("a new key was not created when linkage was next required")
	}
}

// overviewScope is the synthetic scope of the CC-LINKAGE-OVERVIEW case plan.
const overviewScope = "org.example.app.test"

// bundleTop decodes the next pending bundle's top-level members and
// acknowledges it, so the next run starts a new bundle.
func bundleTop(t *testing.T, c *Client) map[string]json.RawMessage {
	t.Helper()
	b, err := c.NextBundle()
	if err != nil || b == "" {
		t.Fatal("no bundle:", codeOf(err))
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal([]byte(b), &top); err != nil {
		t.Fatal("not a bundle")
	}
	if err := c.AcknowledgeBundle(bundleID(t, b)); err != nil {
		t.Fatal("acknowledge failed:", codeOf(err))
	}
	return top
}

// advanceBoth moves the trusted clock and the run clock together.
func advanceBoth(t *testing.T, h *harness, d time.Duration) {
	t.Helper()
	if h.clock.Advance(d) != nil || h.run.clock.Advance(d) != nil {
		t.Fatal("clock advance failed")
	}
}

// expectedToken is HMAC-SHA-256 over the REQ-LINKAGE-0010 message, computed
// here independently of internal/persistence/linkage.
func expectedToken(key [32]byte, scope string, epoch int64) string {
	m := hmac.New(sha256.New, key[:])
	fmt.Fprintf(m, "barometer-result-linkage-v1\x00%s\x00%d", scope, epoch)
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

// CC-LINKAGE-OVERVIEW: without result_linkage and with mode "none", three
// bundles across two 86,400-second epochs carry no linkage, distinct
// bundle_ids and no other repeated bundle-level value, and the key store sees
// no derivation and creates no secret. With rotating_installation the linkage
// object has exactly mode, scope, epoch and token; bundles of one epoch share
// the token, the next epoch derives a new one, and each token is the
// independently computed HMAC of the encoded scope and epoch.
// requirement-ref: REQ-LINKAGE-0001
// requirement-ref: REQ-LINKAGE-0043
func TestCcLinkageOverview(t *testing.T) {
	for name, cfg := range map[string]string{"absent": `{}`, "mode none": `{"result_linkage":{"mode":"none"}}`} {
		h, keys := linkedHarness(t, true)
		c := h.client(t)
		if err := c.Initialize(cfg); err != nil {
			t.Fatal(err)
		}
		seen := map[string]bool{}
		for i := range 3 {
			if i == 2 {
				// 2026-10-20T12:00Z plus 13 hours is in the next 86,400-second epoch.
				advanceBoth(t, h, 13*time.Hour)
			}
			completedRun(t, h, c, fmt.Sprintf("overview-%d", i))
			for member, v := range bundleTop(t, c) {
				switch member {
				case "linkage":
					t.Fatalf("%s: bundle carries linkage", name)
				case "schema_version", "results":
					continue
				}
				if seen[member+"="+string(v)] {
					t.Fatalf("%s: bundle-level %s repeats", name, member)
				}
				seen[member+"="+string(v)] = true
			}
		}
		if held, created, _ := keys.Held(); len(keys.Messages()) != 0 || created != 0 || !held {
			t.Fatalf("%s: key store used or a secret created", name)
		}
	}

	h, keys := linkedHarness(t, true)
	h.adapters.approvals.LinkageScopes = []string{overviewScope}
	c := h.client(t)
	cfg := `{"result_linkage":{"mode":"rotating_installation","scope":"` + overviewScope + `","rotation_period_seconds":86400}}`
	if err := c.Initialize(cfg); err != nil {
		t.Fatal(err)
	}
	var tokens []string
	for i := range 3 {
		if i == 2 {
			advanceBoth(t, h, 13*time.Hour)
		}
		completedRun(t, h, c, fmt.Sprintf("linked-overview-%d", i))
		var l map[string]json.RawMessage
		if err := json.Unmarshal(bundleTop(t, c)["linkage"], &l); err != nil || len(l) != 4 {
			t.Fatalf("linkage object has members other than mode, scope, epoch and token: %v", l)
		}
		var v bundleLinkage
		raw, _ := json.Marshal(l)
		if err := json.Unmarshal(raw, &v); err != nil || v.Mode != "rotating_installation" || v.Scope != overviewScope {
			t.Fatal("linkage mode or scope")
		}
		want := h.clock.Clock.Now().Unix() / 86400
		if v.Epoch != want || v.Token != expectedToken([32]byte{7}, overviewScope, want) {
			t.Fatalf("bundle %d: epoch %d token %q, want epoch %d", i, v.Epoch, v.Token, want)
		}
		tokens = append(tokens, v.Token)
	}
	if tokens[0] != tokens[1] || tokens[1] == tokens[2] || len(keys.Messages()) != 3 {
		t.Fatal("tokens did not follow the epoch")
	}
}

// CC-LINKAGE-INITIALIZATION-CONFIGURATION through Initialize: the mode,
// scope and rotation matrix, absent configuration, an unreviewed scope, tasks
// and the proxy setter carrying linkage fields, and a changed configuration
// that takes effect only after clearing and reinitialization. Local policy
// that disables linkage is asserted by TestCcLinkageRotationLimits.
// requirement-ref: REQ-LINKAGE-0003
// requirement-ref: REQ-LINKAGE-0004
// requirement-ref: REQ-LINKAGE-0005
// requirement-ref: REQ-LINKAGE-0006
// requirement-ref: REQ-LINKAGE-0017
func TestCcLinkageInitializationConfiguration(t *testing.T) {
	long := string(bytes.Repeat([]byte("a"), 128))
	link := func(members string) string { return `{"result_linkage":{` + members + `}}` }
	rot := func(scope string, period int) string {
		return link(fmt.Sprintf(`"mode":"rotating_installation","scope":%q,"rotation_period_seconds":%d`, scope, period))
	}
	// accepted maps each configuration to its expected scope, "" for none.
	accepted := map[string]string{
		link(`"mode":"none"`):     "",
		rot("a", 86400):           "a",
		rot(long, 86400):          long,
		rot(linkageScope, 86400):  linkageScope,
		rot(linkageScope, 604800): linkageScope,
	}
	rejected := map[string]string{
		"other mode":         link(`"mode":"stable_installation"`),
		"wrong-case none":    link(`"mode":"None"`),
		"wrong-case mode":    link(`"mode":"Rotating_installation","scope":"a","rotation_period_seconds":86400`),
		"missing mode":       link(`"scope":"a","rotation_period_seconds":86400`),
		"missing scope":      link(`"mode":"rotating_installation","rotation_period_seconds":86400`),
		"missing rotation":   link(`"mode":"rotating_installation","scope":"a"`),
		"empty scope":        rot("", 86400),
		"129-char scope":     rot(long+"a", 86400),
		"uppercase scope":    rot("Org.example.production", 86400),
		"space in scope":     rot("org example", 86400),
		"slash in scope":     rot("org/example", 86400),
		"non-ASCII scope":    rot("org.exämple", 86400),
		"unreviewed scope":   rot("org.example.unreviewed", 86400),
		"rotation 86,399":    rot(linkageScope, 86399),
		"rotation 604,801":   rot(linkageScope, 604801),
		"none with scope":    link(`"mode":"none","scope":"a"`),
		"none with rotation": link(`"mode":"none","rotation_period_seconds":86400`),
	}
	for name, cfg := range rejected {
		h, keys := linkedHarness(t, true)
		h.adapters.approvals.LinkageScopes = []string{linkageScope, "a", long}
		if codeOf(h.client(t).Initialize(cfg)) != "not_initialized" || len(keys.Messages()) != 0 {
			t.Errorf("%s accepted or reached the key store", name)
		}
	}
	for cfg, scope := range accepted {
		h, keys := linkedHarness(t, true)
		h.adapters.approvals.LinkageScopes = []string{linkageScope, "a", long}
		c := h.client(t)
		if err := c.Initialize(cfg); err != nil {
			t.Fatalf("%s rejected: %v", cfg, codeOf(err))
		}
		completedRun(t, h, c, "accepted")
		l := linkageOf(t, c)
		if scope == "" && (l != nil || len(keys.Messages()) != 0) || scope != "" && (l == nil || l.Scope != scope) {
			t.Fatalf("%s: linkage %+v", cfg, l)
		}
	}

	// Tasks cannot carry linkage configuration, and the runtime proxy setter
	// cannot change it: both are rejected and the next bundle keeps the
	// initialized scope and 86,400-second epoch.
	h, keys := linkedHarness(t, true)
	c := h.client(t)
	if err := c.Initialize(rotatingConfig(86400)); err != nil {
		t.Fatal(err)
	}
	extra := `,"result_linkage":{"mode":"rotating_installation","scope":"org.example.other","rotation_period_seconds":604800}`
	task := runTask("task-linkage", "", "", false)
	for name, body := range map[string]string{
		"top level": task[:len(task)-1] + extra + "}",
		"options":   strings.Replace(task, `"options":{}`, `"options":{`+extra[1:]+`}`, 1),
	} {
		if _, err := c.AdmitTask(body, h.embedded(t, body)); err == nil {
			t.Fatalf("task with linkage in %s admitted", name)
		}
	}
	if c.SetProxy(`{"result_linkage":{"mode":"none"}}`) == nil {
		t.Fatal("proxy setter accepted linkage configuration")
	}
	completedRun(t, h, c, "after-attempts")
	if l := linkageOf(t, c); l == nil || l.Scope != linkageScope || l.Epoch != epoch.Unix()/86400 {
		t.Fatalf("linkage changed at runtime: %+v", l)
	}
	for i := 0; i < reflect.TypeOf(c).NumMethod(); i++ {
		if name := reflect.TypeOf(c).Method(i).Name; strings.Contains(strings.ToLower(name), "linkage") {
			t.Fatalf("Client exposes %s", name)
		}
	}
	// A changed configuration applies only after reinitialization.
	if codeOf(c.Initialize(rotatingConfig(604800))) != "already_initialized" {
		t.Fatal("reinitialization without clearing accepted")
	}
	if err := c.ClearMeasurementData(); err != nil {
		t.Fatal(err)
	}
	if err := c.Initialize(rotatingConfig(604800)); err != nil {
		t.Fatal(err)
	}
	completedRun(t, h, c, "reinitialized")
	if l := linkageOf(t, c); l == nil || l.Epoch != epoch.Unix()/604800 || len(keys.Messages()) != 2 {
		t.Fatalf("reinitialized linkage %+v", l)
	}
}

// Rows of CC-LINKAGE-STORAGE-FAILURE at the facade: an unavailable key store
// omits linkage while the result is still delivered in a bundle
// with its single-use bundle_id; a token derived before a failure is not
// reused after it; and a key store that fails deletion makes clearing report
// an explicit error without keeping pending results.
// requirement-ref: REQ-LINKAGE-0037
// requirement-ref: REQ-LINKAGE-0038
func TestLinkageStorageFailure(t *testing.T) {
	h, keys := linkedHarness(t, true)
	c := h.client(t)
	if err := c.Initialize(rotatingConfig(604800)); err != nil {
		t.Fatal(err)
	}
	keys.Fail()
	completedRun(t, h, c, "unavailable")
	b, err := c.NextBundle()
	if err != nil || len(bundleResults(t, b)) != 1 || bytes.Contains([]byte(b), []byte("linkage")) {
		t.Fatal("unavailable key store: linkage emitted or result not delivered")
	}
	if c.AcknowledgeBundle(bundleID(t, b)) != nil {
		t.Fatal("acknowledge failed")
	}
	keys.Recover()
	completedRun(t, h, c, "available")
	if bundleTop(t, c)["linkage"] == nil {
		t.Fatal("no linkage after recovery")
	}
	keys.Fail()
	completedRun(t, h, c, "failed-again")
	if bundleTop(t, c)["linkage"] != nil {
		t.Fatal("a token derived before the failure was reused")
	}

	// Pending results are deleted even when the key cannot be.
	completedRun(t, h, c, "pending")
	if h.run.output.Len(persistence.PendingClass) == 0 {
		t.Fatal("no pending result to clear")
	}
	if c.ClearMeasurementData() == nil || h.run.output.Len(persistence.PendingClass) != 0 {
		t.Fatal("clearing with a failing key store reported success or kept results")
	}

}

// The synthetic installation secret held by the test key store never appears,
// raw or encoded, in any value or error the Client API returns or in
// library-owned storage; the derived token appears only as the bundle's
// linkage.token. Every exported Client method takes and returns only strings
// and errors, so no key, key handle or storage object crosses the API, and the
// internal key-store contract returns derived tokens, never the secret.
// requirement-ref: REQ-LINKAGE-0008
// requirement-ref: REQ-LINKAGE-0019
// requirement-ref: REQ-LINKAGE-0046
func TestLinkageSecretNotExposed(t *testing.T) {
	var secret [32]byte
	for i := range secret {
		secret[i] = byte(0xa5 ^ i*29)
	}
	h := runningHarness(t)
	keys := capability.NewLinkageKeys(secret)
	h.adapters.linkageKeys = keys
	h.adapters.approvals = config.Approvals{LinkageScopes: []string{linkageScope}, AllowLinkage: true}
	c := h.client(t)

	var out []string
	note := func(s string, err error) {
		out = append(out, s)
		if err != nil {
			out = append(out, err.Error())
		}
	}
	note("", c.Initialize(rotatingConfig(604800)))
	note("", c.Initialize(rotatingConfig(604800)))
	task := runTask("secret", "", "", false)
	th, err := c.AdmitTask(task, h.embedded(t, task))
	note(th, err)
	note(c.AdmitTask("{}", "unknown"))
	rh, err := c.StartRun(th)
	note(rh, err)
	waitTerminal(t, c, rh)
	note(c.RunState(rh))
	note(c.RunState("unknown"))
	note("", c.SetProxy(`{"result_linkage":{}}`))
	b, err := c.NextBundle()
	if err != nil {
		t.Fatal(err)
	}
	var w struct {
		Linkage bundleLinkage `json:"linkage"`
	}
	if json.Unmarshal([]byte(b), &w) != nil || w.Linkage.Token == "" || strings.Count(b, w.Linkage.Token) != 1 {
		t.Fatal("token missing or outside linkage.token")
	}
	note("", c.AcknowledgeBundle("unknown"))
	note("", c.AcknowledgeBundle(bundleID(t, b)))
	stored := string(h.run.output.Dump())
	note("", c.ClearMeasurementData())
	note("", c.Close())

	forms := []string{string(secret[:]), hex.EncodeToString(secret[:]), strings.ToUpper(hex.EncodeToString(secret[:])),
		base64.StdEncoding.EncodeToString(secret[:]), base64.RawStdEncoding.EncodeToString(secret[:]),
		base64.URLEncoding.EncodeToString(secret[:]), base64.RawURLEncoding.EncodeToString(secret[:]), fmt.Sprint(secret[:])}
	for _, s := range append(out, b, stored) {
		for _, f := range forms {
			if strings.Contains(s, f) {
				t.Fatal("installation secret exposed")
			}
		}
	}
	for _, s := range out {
		if strings.Contains(s, w.Linkage.Token) {
			t.Fatal("token exposed outside the bundle")
		}
	}

	str, errType := reflect.TypeOf(""), reflect.TypeOf((*error)(nil)).Elem()
	ct := reflect.TypeOf(c)
	for i := 0; i < ct.NumMethod(); i++ {
		m := ct.Method(i).Type
		for j := 1; j < m.NumIn(); j++ {
			if m.In(j) != str {
				t.Fatalf("Client.%s takes a %v", ct.Method(i).Name, m.In(j))
			}
		}
		for j := 0; j < m.NumOut(); j++ {
			if m.Out(j) != str && m.Out(j) != errType {
				t.Fatalf("Client.%s returns a %v", ct.Method(i).Name, m.Out(j))
			}
		}
	}
	kt := reflect.TypeOf((*platform.LinkageKeyStore)(nil)).Elem()
	derive, _ := kt.MethodByName("DeriveHMACSHA256")
	if kt.NumMethod() != 2 || derive.Type.NumOut() != 2 || derive.Type.Out(0) != reflect.TypeOf([32]byte{}) {
		t.Fatal("key-store contract changed")
	}
}
