// Package barometermobile is the gomobile binding surface of Barometer. It is
// the only package bound with gomobile, standalone and in an umbrella build.
// It is not a supported Go API: Go integrators use the root package.
//
// Only string, []byte, int64, bool, error, and interfaces and struct pointers
// declared here cross this package's exported surface. The native adapters and
// the trusted local document reach it only from reviewed, Barometer-owned
// native code; Dart passes only the serializable lifecycle values.
package barometermobile

import (
	"context"
	"errors"
	"runtime"
	"time"

	barometer "github.com/tv-rain/barometer-go"
	"github.com/tv-rain/barometer-go/internal/admission"
	"github.com/tv-rain/barometer-go/internal/execution"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/persistence"
	"github.com/tv-rain/barometer-go/internal/seam"
	"github.com/tv-rain/barometer-go/internal/trustdoc"
)

// libraryVersion is the implementation identity recorded in results.
const libraryVersion = "0.1.0"

// Reviewed binding inputs: implementation-local bounds, not registered limits.
const (
	replayCapacity   = 1024
	deliveryCapacity = 64
	deliveryLifetime = 10 * time.Minute
	installDeadline  = 5 * time.Second
)

// Facade-only error codes. Lifecycle methods return the root's codes.
var (
	errAdapters  = errors.New("adapters_invalid")
	errTrust     = errors.New("trust_invalid")
	errRejected  = errors.New("delivery_rejected")
	errDelivery  = errors.New("delivery_unavailable")
	errCapacity  = errors.New("capacity_exceeded")
	errClosed    = errors.New("client_closed")
	errNilClient = errors.New("not_initialized")
)

// Client owns one root client bound to the native adapters.
type Client struct {
	root       *barometer.Client
	deliveries *admission.DeliveryStore
	generation uint64
}

// options are test-only inputs that NewClient never sets.
type options struct {
	platform      string
	directDialers execution.DialerSource
}

// NewClient binds adapters and the trusted local document to a new,
// uninitialized client. trust is the "barometer-trust/1" document that
// reviewed native code reads from the application package; it never comes
// from Dart, task data or the network. NewClient performs no I/O and starts
// no measurement. It returns adapters_invalid when a required adapter is
// missing or the platform is not Android or iOS, and trust_invalid when the
// document is invalid.
func NewClient(adapters *Adapters, trust []byte) (*Client, error) {
	platform := ""
	switch runtime.GOOS {
	case "android", "ios":
		platform = runtime.GOOS
	}
	return newClient(adapters, trust, options{platform: platform})
}

func newClient(in *Adapters, trust []byte, o options) (*Client, error) {
	if in == nil || in.Records == nil || in.Clock == nil || in.Network == nil || o.platform == "" {
		return nil, errAdapters
	}
	doc, err := trustdoc.Decode(trust)
	if err != nil {
		return nil, errTrust
	}
	rec := records{in.Records}
	trusted := clock{in.Clock}
	state, err := persistence.NewLazyState(rec, replayCapacity)
	if err != nil {
		return nil, errAdapters
	}
	deliveries, err := admission.NewDeliveryStore(trusted, admission.DeliveryLimits{Capacity: deliveryCapacity, Lifetime: deliveryLifetime}, doc.Policy.Generation())
	if err != nil {
		return nil, errAdapters
	}
	a := seam.Adapters{
		Clock: trusted, Records: rec, Network: network{in.Network},
		RunClock: execution.SystemClock{}, Jitter: execution.LocalJitter{},
		Replay: state.Replay(), Ledger: state.Runs(), Deliveries: deliveries,
		Policy: doc.Policy, Approvals: doc.Approvals, Context: doc.Context,
		Implementation: model.Implementation{LibraryVersion: libraryVersion, Platform: o.platform},
		DirectDialers:  o.directDialers,
	}
	// Optional adapters stay nil interfaces when absent, never typed nils.
	if in.LinkageKeys != nil {
		a.LinkageKeys = linkageKeys{in.LinkageKeys}
	}
	if in.Consent != nil {
		a.Consent = consent{in.Consent}
	}
	if in.Device != nil {
		a.Device = device{in.Device}
	}
	if in.TimeZone != nil {
		a.TimeZone = timeZone{in.TimeZone}
	}
	if in.Resolver != nil {
		a.Resolver = resolver{in.Resolver}
	}
	root := barometer.NewClient()
	if seam.Bind(root, a) != nil {
		return nil, errAdapters
	}
	return &Client{root: root, deliveries: deliveries, generation: doc.Policy.Generation()}, nil
}

// InstallEmbeddedTask installs a delivery record that associates the exact
// task bytes with issuerID and returns its opaque sourceRef for AdmitTask.
// The association is a trust input: only reviewed native code that ships the
// task with the application calls it, never Dart.
func (c *Client) InstallEmbeddedTask(issuerID string, taskJSON []byte) (string, error) {
	if c == nil {
		return "", errNilClient
	}
	ctx, cancel := context.WithTimeout(context.Background(), installDeadline)
	defer cancel()
	ref, err := c.deliveries.InstallEmbedded(ctx, issuerID, c.generation, append([]byte{}, taskJSON...))
	return ref, installError(err)
}

// InstallRemoteTask installs a delivery record for the exact signed remote
// task package bytes and returns its opaque sourceRef for AdmitTask. The
// package carries no authority: admission verifies it against the issuer keys
// of the trusted local policy.
func (c *Client) InstallRemoteTask(signedPackage []byte) (string, error) {
	if c == nil {
		return "", errNilClient
	}
	ctx, cancel := context.WithTimeout(context.Background(), installDeadline)
	defer cancel()
	ref, err := c.deliveries.InstallRemote(ctx, c.generation, append([]byte{}, signedPackage...))
	return ref, installError(err)
}

func installError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, admission.ErrDeliveryRejected):
		return errRejected
	case errors.Is(err, admission.ErrDeliveryCapacity):
		return errCapacity
	case errors.Is(err, admission.ErrDeliveryClosed):
		return errClosed
	}
	return errDelivery
}

// The lifecycle methods forward to the root Client with the same string forms
// and error codes; see the root package documentation.

func (c *Client) Initialize(configJSON string) error {
	if c == nil {
		return errNilClient
	}
	return c.root.Initialize(configJSON)
}

func (c *Client) AdmitTask(taskJSON string, sourceRef string) (string, error) {
	if c == nil {
		return "", errNilClient
	}
	return c.root.AdmitTask(taskJSON, sourceRef)
}

func (c *Client) StartRun(taskHandle string) (string, error) {
	if c == nil {
		return "", errNilClient
	}
	return c.root.StartRun(taskHandle)
}

func (c *Client) RunState(runHandle string) (string, error) {
	if c == nil {
		return "", errNilClient
	}
	return c.root.RunState(runHandle)
}

func (c *Client) CancelRun(runHandle string) error {
	if c == nil {
		return errNilClient
	}
	return c.root.CancelRun(runHandle)
}

func (c *Client) NextBundle() (string, error) {
	if c == nil {
		return "", errNilClient
	}
	return c.root.NextBundle()
}

func (c *Client) AcknowledgeBundle(bundleID string) error {
	if c == nil {
		return errNilClient
	}
	return c.root.AcknowledgeBundle(bundleID)
}

func (c *Client) SetProxy(proxyJSON string) error {
	if c == nil {
		return errNilClient
	}
	return c.root.SetProxy(proxyJSON)
}

func (c *Client) ClearProxy() error {
	if c == nil {
		return errNilClient
	}
	return c.root.ClearProxy()
}

func (c *Client) ClearMeasurementData() error {
	if c == nil {
		return errNilClient
	}
	return c.root.ClearMeasurementData()
}

func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	return c.root.Close()
}
