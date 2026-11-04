package barometer

import (
	"github.com/tv-rain/barometer-go/internal/seam"
)

// The facade package barometermobile reaches bind only through internal/seam,
// so the root gains no export for adapter installation.
func init() { seam.Register(bindSeam) }

// bindSeam installs the facade's trusted adapters on an uninitialized root
// client. The registry and trust roots stay the production defaults.
func bindSeam(client any, s seam.Adapters) error {
	c, ok := client.(*Client)
	if !ok || c == nil {
		return fail(notInitialized)
	}
	return c.bind(adapters{
		clock:            s.Clock,
		policy:           s.Policy,
		deliveries:       s.Deliveries,
		replay:           s.Replay,
		approvals:        s.Approvals,
		linkageKeys:      s.LinkageKeys,
		contextApprovals: s.Context,
		records:          s.Records,
		run: runAdapters{
			resolver:       s.Resolver,
			clock:          s.RunClock,
			jitter:         s.Jitter,
			ledger:         s.Ledger,
			network:        s.Network,
			implementation: s.Implementation,
			consent:        s.Consent,
			device:         s.Device,
			zone:           s.TimeZone,
			directDialers:  s.DirectDialers,
		},
	})
}
