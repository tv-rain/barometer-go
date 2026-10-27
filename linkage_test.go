package barometer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/tv-rain/barometer-go/internal/config"
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
