// Package seam carries trusted adapters from the gomobile facade to the root
// package's unexported adapter seam without a root export.
//
// The root package registers its binder from init; the barometermobile facade
// calls Bind. The dependency rules allow only those two packages to import
// this one, so no code outside barometer-go can install adapters and no other
// package of the module installs them either. The values are in-process only:
// Adapters has no serialized form.
package seam

import (
	"errors"
	"sync"

	"github.com/tv-rain/barometer-go/internal/admission"
	"github.com/tv-rain/barometer-go/internal/config"
	"github.com/tv-rain/barometer-go/internal/execution"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/persistence"
	"github.com/tv-rain/barometer-go/internal/platform"
)

// ErrUnavailable is returned by Bind when no binder is registered or the
// binder refused the adapters. It carries no detail.
var ErrUnavailable = errors.New("seam: adapter binding unavailable")

// Adapters are the trusted inputs the facade assembles. Optional capabilities
// are nil when absent; an interface field never holds a typed nil. Trust roots
// are not carried: the root selects the system store.
type Adapters struct {
	Clock       platform.Clock
	Records     persistence.Records
	LinkageKeys platform.LinkageKeyStore // optional; absence disables linkage only
	Network     platform.Network
	Consent     platform.Consent             // optional
	Device      platform.Device              // optional
	TimeZone    platform.TimeZone            // optional
	Resolver    execution.SystemResolver     // optional
	RunClock    execution.Clock              // the engine clock, owned by Go
	Jitter      execution.Jitter             // owned by Go
	Replay      persistence.ReplayStore      // over Records
	Ledger      persistence.RunLedger        // over Records
	Deliveries  *admission.DeliveryStore     // owned by the facade
	Policy      task.LocalPolicy             // from the trusted local document
	Approvals   config.Approvals             // from the trusted local document
	Context     observation.ContextApprovals // from the trusted local document

	Implementation model.Implementation

	// DirectDialers replaces the direct path's dialers. Only facade tests set
	// it, through an unexported constructor, to reach loopback fixtures; the
	// facade's exported constructor never does.
	DirectDialers execution.DialerSource
}

// Binder installs a on client, which must be the root package's *Client.
type Binder func(client any, a Adapters) error

var (
	mu     sync.Mutex
	binder Binder
)

// Register sets the binder once. The root package calls it from init; a second
// registration is a programming error.
func Register(b Binder) {
	mu.Lock()
	defer mu.Unlock()
	if b == nil || binder != nil {
		panic("seam: binder registered twice")
	}
	binder = b
}

// Bind installs a on client through the registered binder.
func Bind(client any, a Adapters) error {
	mu.Lock()
	b := binder
	mu.Unlock()
	if b == nil || b(client, a) != nil {
		return ErrUnavailable
	}
	return nil
}
