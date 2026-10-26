package barometer

import "github.com/tv-rain/barometer-go/internal/registry"

// productionRegistry composes only delivered production measurements. Keep
// assembly here: registry owns metadata, and measurement packages must not
// import it. There is no global state or runtime registration.
//
// No production codec or runner has been delivered yet. The empty set rejects
// every test version; synthetic entries belong exclusively in tests.
func productionRegistry() (*registry.Registry, error) {
	return registry.New(nil)
}
