package execution

import (
	"context"
	"net"

	"github.com/tv-rain/barometer-go/internal/measurement"
)

// DirectDialers is the direct path's DialerSource. Each attempt gets its own
// dialer, which is the attempt's path snapshot (REQ-EXECUTION-0009): every
// connection of the attempt uses the device's direct network and no proxy,
// and nothing is shared with another attempt. The engine wraps it in the
// attempt scope, which meters and closes every connection, so the dialer
// itself holds no state and sets no limit.
type DirectDialers struct{}

func (DirectDialers) Dialer(int) (measurement.Dialer, error) { return directDialer{}, nil }

// directDialer is the path transport under the shared target gate. For a
// measurement it receives only the checked literal IP, never a name to resolve
// again. Its zero net.Dialer consults no proxy setting or connection pool.
type directDialer struct{}

func (directDialer) DialContext(ctx context.Context, network, address string) (measurement.Conn, error) {
	var d net.Dialer
	return d.DialContext(ctx, network, address)
}
