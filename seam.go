package barometer

import (
	"crypto/x509"

	"github.com/tv-rain/barometer-go/internal/seam"
)

// The facade package barometermobile reaches bind only through internal/seam,
// so the root gains no export for adapter installation.
func init() { seam.Register(bindSeam) }

// bindSeam installs the facade's trusted adapters on an uninitialized root
// client. The registry is the production one; the trust roots are the
// platform root set of seamRoots.
func bindSeam(client any, s seam.Adapters) error {
	c, ok := client.(*Client)
	if !ok || c == nil {
		return fail(notInitialized)
	}
	roots, available := seamRoots(platformRoots)
	return c.bind(adapters{
		measurementRoots: roots,
		contextRoots:     roots,
		rootsUnavailable: !available,
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

// seamRoots returns the measurement and context trust roots of a seam-bound
// client from the platform loader. Without a loader, as outside Android, the
// roots are nil and Go's system roots apply. On Android (decision 0016) the
// roots are the loaded platform set; a loader error or an empty result
// reports the set unavailable and never falls back to Go's defaults. The
// loader is fixed at build time: no adapter, document, configuration or task
// value reaches it.
func seamRoots(load func() (*x509.CertPool, error)) (*x509.CertPool, bool) {
	if load == nil {
		return nil, true
	}
	pool, err := load()
	if err != nil || pool == nil {
		return nil, false
	}
	return pool, true
}
