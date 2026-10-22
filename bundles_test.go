package barometer

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/persistence"
	"github.com/tv-rain/barometer-go/internal/persistence/pending"
	recordstore "github.com/tv-rain/barometer-go/internal/testkit/records"
)

// Synthetic facade cases for NextBundle and AcknowledgeBundle (DEV-563) over
// the production pending store and the fault-injectable record port. No
// runtime, platform or durability evidence.

// completedRun runs one single-attempt task to completion.
func completedRun(t *testing.T, h *harness, c *Client, id string) {
	t.Helper()
	rh, err := c.StartRun(h.admit(t, c, runTask(id, "", "", false)))
	if err != nil {
		t.Fatal(err)
	}
	waitTerminal(t, c, rh)
}

func bundleResults(t *testing.T, b string) []resultRecord {
	t.Helper()
	var w struct {
		SchemaVersion string         `json:"schema_version"`
		Results       []resultRecord `json:"results"`
	}
	if err := json.Unmarshal([]byte(b), &w); err != nil || w.SchemaVersion != "0.1.0" {
		t.Fatal("not a Result Bundle")
	}
	return w.Results
}

// A completed run's result is returned in a bundle, with the same bytes on
// every call until acknowledgement; acknowledgement deletes it immediately.
// requirement-ref: REQ-LIMIT-0024
func TestNextBundleAndAcknowledge(t *testing.T) {
	h := runningHarness(t)
	c := h.initialized(t)
	if b, err := c.NextBundle(); b != "" || err != nil {
		t.Fatal("bundle before any result:", err)
	}
	completedRun(t, h, c, "nb-1")
	b1, err := c.NextBundle()
	if err != nil || len(bundleResults(t, b1)) != 1 {
		t.Fatal("no bundle for the completed run:", err)
	}
	completedRun(t, h, c, "nb-2")
	if b, _ := c.NextBundle(); b != b1 {
		t.Fatal("returned bundle changed")
	}
	id := bundleID(t, b1)
	if err := c.AcknowledgeBundle(id); err != nil {
		t.Fatal(err)
	}
	if codeOf(c.AcknowledgeBundle(id)) != "bundle_unavailable" || codeOf(c.AcknowledgeBundle("")) != "bundle_unavailable" ||
		codeOf(c.AcknowledgeBundle("T.not-a-bundle")) != "bundle_unavailable" {
		t.Fatal("deleted or unknown identifier accepted")
	}
	if bytes.Contains(h.run.output.Dump(), []byte(id)) {
		t.Fatal("acknowledged bundle remains in storage")
	}
	b2, err := c.NextBundle()
	if err != nil || bundleID(t, b2) == id || len(bundleResults(t, b2)) != 1 {
		t.Fatal("second result not delivered in a new bundle")
	}
	if c.AcknowledgeBundle(bundleID(t, b2)) != nil {
		t.Fatal("acknowledgement failed")
	}
	if b, err := c.NextBundle(); b != "" || err != nil || h.run.output.Len(persistence.PendingClass) != 0 {
		t.Fatal("pending state remains after acknowledgement")
	}
}

// Storage failures are operational errors: a failed deletion is not
// acknowledgement and a retry deletes the bundle.
func TestBundleStorageFailures(t *testing.T) {
	h := runningHarness(t)
	c := h.initialized(t)
	completedRun(t, h, c, "sf-1")
	h.run.output.Inject(recordstore.FailGet, true)
	if b, err := c.NextBundle(); b != "" || codeOf(err) != "storage_unavailable" {
		t.Fatal("unreadable bundle:", codeOf(err))
	}
	h.run.output.Inject(recordstore.FailGet, false)
	b, err := c.NextBundle()
	if err != nil {
		t.Fatal(err)
	}
	h.run.output.Inject(recordstore.FailBeforeDelete, true)
	if codeOf(c.AcknowledgeBundle(bundleID(t, b))) != "storage_unavailable" {
		t.Fatal("failed deletion reported as acknowledgement")
	}
	h.run.output.Inject(recordstore.FailBeforeDelete, false)
	if again, _ := c.NextBundle(); again != b {
		t.Fatal("bundle lost after a failed deletion")
	}
	if c.AcknowledgeBundle(bundleID(t, b)) != nil {
		t.Fatal("retry failed")
	}
}

// Initialization reads pending storage: unreadable storage fails it closed,
// and bundles that expired while no client was initialized are deleted
// before use. Close neither deletes nor extends pending state.
func TestBundleRetentionAcrossClose(t *testing.T) {
	h := runningHarness(t)
	c := h.initialized(t)
	completedRun(t, h, c, "rc-1")
	b, _ := c.NextBundle()
	if c.Close() != nil || h.run.output.Len(persistence.PendingClass) != 1 {
		t.Fatal("close deleted pending state")
	}
	h.run.output.Restart()
	h.run.output.Inject(recordstore.FailList, true)
	if codeOf(h.initialized2(t)) != "not_initialized" {
		t.Fatal("initialized over unreadable pending storage")
	}
	h.run.output.Inject(recordstore.FailList, false)
	c2 := h.initialized(t)
	if again, _ := c2.NextBundle(); again != b {
		t.Fatal("bundle not recovered after restart")
	}
	if c2.Close() != nil {
		t.Fatal("close failed")
	}
	h.advance(pending.Retention)
	c3 := h.initialized(t)
	if h.run.output.Len(persistence.PendingClass) != 0 {
		t.Fatal("initialization kept an expired bundle")
	}
	if b, err := c3.NextBundle(); b != "" || err != nil {
		t.Fatal("expired bundle returned late")
	}
}

// initialized2 initializes a fresh client and returns the error.
func (h *harness) initialized2(t *testing.T) error {
	t.Helper()
	return h.client(t).Initialize(`{}`)
}

// Bundle entry points follow the lifecycle table: they fail closed before
// initialization, after close and while a transition is incomplete.
func TestBundleLifecycle(t *testing.T) {
	var nilClient *Client
	if b, err := nilClient.NextBundle(); b != "" || codeOf(err) != "not_initialized" || codeOf(nilClient.AcknowledgeBundle("x")) != "not_initialized" {
		t.Fatal("nil client")
	}
	h := runningHarness(t)
	c := h.client(t)
	if _, err := c.NextBundle(); codeOf(err) != "not_initialized" || codeOf(c.AcknowledgeBundle("x")) != "not_initialized" {
		t.Fatal("uninitialized client")
	}
	if c.Initialize(`{}`) != nil {
		t.Fatal("initialize failed")
	}
	g := newGate()
	h.adapters.clearStorage = g.clear
	c2 := h.initialized(t)
	done := make(chan error, 1)
	go func() { done <- c2.ClearMeasurementData() }()
	<-g.entered
	if _, err := c2.NextBundle(); codeOf(err) != "transition_in_progress" || codeOf(c2.AcknowledgeBundle("x")) != "transition_in_progress" {
		t.Fatal("bundle access during a transition")
	}
	g.result = errors.New("synthetic clear failure")
	close(g.release)
	<-done
	if _, err := c2.NextBundle(); codeOf(err) != "clearing_incomplete" || codeOf(c2.AcknowledgeBundle("x")) != "clearing_incomplete" {
		t.Fatal("bundle access while clearing is incomplete")
	}
	if c.Close() != nil {
		t.Fatal("close failed")
	}
	if _, err := c.NextBundle(); codeOf(err) != "client_closed" || codeOf(c.AcknowledgeBundle("x")) != "client_closed" {
		t.Fatal("closed client")
	}
}

// A stalled pending backend does not block bundle calls beyond the lifecycle
// deadline, and a transition is not blocked by them.
func TestBundleCallsAreBounded(t *testing.T) {
	h := runningHarness(t)
	c := h.initialized(t)
	completedRun(t, h, c, "st-1")
	b, _ := c.NextBundle()
	h.run.output.Inject(recordstore.Stall, true)
	start := time.Now()
	if _, err := c.NextBundle(); codeOf(err) != "storage_unavailable" {
		t.Fatal("stalled storage:", codeOf(err))
	}
	if codeOf(c.AcknowledgeBundle(bundleID(t, b))) != "storage_unavailable" {
		t.Fatal("stalled acknowledgement")
	}
	if time.Since(start) > 5*time.Second {
		t.Fatal("bundle calls outlived the lifecycle deadline")
	}
	h.run.output.Inject(recordstore.Stall, false)
	if c.Close() != nil {
		t.Fatal("close blocked by bundle calls")
	}
}
