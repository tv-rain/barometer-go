package examples

import (
	"errors"
	"fmt"
)

// lifecycleClient is a local subset of docs/public-api.md's intended facade.
// There is no production Client to construct or inject through this interface.
type lifecycleClient interface {
	Initialize(string) error
	CancelRun(string) error
	ClearMeasurementData() error
	Close() error
}

type scriptedCall struct {
	method, argument string
	err              error
}

// scriptedClient ONLY checks host call order/arguments and returns scripted
// errors. It has no client state, storage, secrets, workers, or deletion logic.
// A nil scripted error proves no runtime behavior. Keep this in _test.go.
type scriptedClient struct{ calls []scriptedCall }

var _ lifecycleClient = (*scriptedClient)(nil)

func (c *scriptedClient) call(method, argument string) error {
	if len(c.calls) == 0 || c.calls[0].method != method || c.calls[0].argument != argument {
		panic("unexpected test-only lifecycle call")
	}
	err := c.calls[0].err
	c.calls = c.calls[1:]
	return err
}

func (c *scriptedClient) Initialize(s string) error   { return c.call("Initialize", s) }
func (c *scriptedClient) CancelRun(s string) error    { return c.call("CancelRun", s) }
func (c *scriptedClient) ClearMeasurementData() error { return c.call("ClearMeasurementData", "") }
func (c *scriptedClient) Close() error                { return c.call("Close", "") }

// clearLocalData is host-side control flow, not an implementation of clearing.
// The caller must keep admission, starts and uploads paused on failure and
// separately delete any host-owned copies before reporting complete removal.
func clearLocalData(client lifecycleClient) string {
	if err := client.ClearMeasurementData(); err != nil {
		return "incomplete: keep measurement and delivery paused"
	}
	return "Barometer clear returned success; remove host-owned copies next"
}

func Example_intendedClientLifecycle() {
	fake := &scriptedClient{calls: []scriptedCall{
		{method: "Initialize", argument: `{}`, err: errors.New("not_initialized")},
		{method: "Close"},
	}}
	var client lifecycleClient = fake
	if err := client.Initialize(`{}`); err != nil {
		// No admission/start after initialization failure; no raw error logging.
		fmt.Println("initialization unavailable: do not start measurements")
	}
	if err := client.Close(); err != nil {
		fmt.Println("close incomplete: keep work paused")
	}
	fmt.Println("script consumed:", len(fake.calls) == 0)
	// Output:
	// initialization unavailable: do not start measurements
	// script consumed: true
}

func Example_intendedCancellationAndClearing() {
	// synthetic-run is a test placeholder, never a source/trust assertion.
	fake := &scriptedClient{calls: []scriptedCall{
		{method: "CancelRun", argument: "synthetic-run", err: errors.New("cleanup_incomplete")},
		{method: "ClearMeasurementData", err: errors.New("clearing_incomplete")},
		{method: "ClearMeasurementData"},
		{method: "ClearMeasurementData"},
	}}
	var client lifecycleClient = fake
	if err := client.CancelRun("synthetic-run"); err != nil {
		fmt.Println("cancellation incomplete: keep work paused")
	}
	fmt.Println(clearLocalData(client))
	// A host may explicitly retry clearing while the affected scope stays paused.
	fmt.Println(clearLocalData(client))
	fmt.Println(clearLocalData(client))
	fmt.Println("script consumed:", len(fake.calls) == 0)
	// Output:
	// cancellation incomplete: keep work paused
	// incomplete: keep measurement and delivery paused
	// Barometer clear returned success; remove host-owned copies next
	// Barometer clear returned success; remove host-owned copies next
	// script consumed: true
}
