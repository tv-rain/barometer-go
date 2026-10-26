package execution

import (
	"context"
	"net"
	"net/netip"
	"syscall"

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
	if network == "udp" || network == "udp4" || network == "udp6" {
		peer, err := netip.ParseAddrPort(address)
		if err != nil {
			return nil, measurement.ErrTargetRejected
		}
		family, local := "udp6", "[::]:0"
		if peer.Addr().Is4() {
			family, local = "udp4", "0.0.0.0:0"
		}
		var lc net.ListenConfig
		packet, err := lc.ListenPacket(ctx, family, local)
		if err != nil {
			return nil, err
		}
		return directUDP{UDPConn: packet.(*net.UDPConn), peer: peer}, nil
	}
	var d net.Dialer
	return d.DialContext(ctx, network, address)
}

// directUDP sends only to the checked peer. Source comparison and truncation
// survive the attempt meter as booleans, without exposing addresses to runners.
type directUDP struct {
	*net.UDPConn
	peer netip.AddrPort
}

func (c directUDP) Write(p []byte) (int, error) { return c.WriteToUDPAddrPort(p, c.peer) }
func (c directUDP) ReadDatagram(p []byte) (int, bool, bool, error) {
	n, _, flags, peer, err := c.ReadMsgUDPAddrPort(p, nil)
	return max(0, n), flags&syscall.MSG_TRUNC != 0, peer == c.peer, err
}
func (c directUDP) Read(p []byte) (int, error) {
	for {
		n, _, matching, err := c.ReadDatagram(p)
		if err != nil || matching {
			return n, err
		}
	}
}
