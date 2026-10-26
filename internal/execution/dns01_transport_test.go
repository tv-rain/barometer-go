package execution

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/measurement"
)

// The production UDP transport observes source equality without handing a raw
// peer address to the measurement, and preserves the kernel truncation flag.
// requirement-ref: REQ-DNS-01-0010
func TestDnsUDPSourceAndTruncation(t *testing.T) {
	server, e := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if e != nil {
		t.Fatal(e)
	}
	defer server.Close()
	rogue, e := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if e != nil {
		t.Fatal(e)
	}
	defer rogue.Close()
	c, e := (directDialer{}).DialContext(context.Background(), "udp", server.LocalAddr().String())
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	_ = c.SetDeadline(time.Now().Add(time.Second))
	_ = server.SetDeadline(time.Now().Add(time.Second))
	if _, e = c.Write([]byte("synthetic query")); e != nil {
		t.Fatal(e)
	}
	b := make([]byte, 64)
	_, peer, e := server.ReadFromUDP(b)
	if e != nil {
		t.Fatal(e)
	}
	d := c.(measurement.DatagramConn)
	if _, e = rogue.WriteToUDP([]byte("foreign"), peer); e != nil {
		t.Fatal(e)
	}
	n, truncated, matching, e := d.ReadDatagram(b)
	if e != nil || matching || truncated || n != 7 {
		t.Fatalf("source %d %v %v %v", n, truncated, matching, e)
	}
	if _, e = server.WriteToUDP(make([]byte, 128), peer); e != nil {
		t.Fatal(e)
	}
	n, truncated, matching, e = d.ReadDatagram(b)
	if e != nil || !matching || !truncated || n != 64 {
		t.Fatalf("truncation %d %v %v %v", n, truncated, matching, e)
	}
}
