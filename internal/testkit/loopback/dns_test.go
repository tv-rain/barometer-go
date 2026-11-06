package loopback

import (
	"net"
	"strconv"
	"testing"
)

// holdUDPPort binds UDP on an ephemeral loopback port whose TCP number is
// free, so a TCP listener can take that number while the UDP port collides.
func holdUDPPort(t *testing.T) (*net.UDPConn, int) {
	t.Helper()
	for i := 0; i < 16; i++ {
		u, e := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
		if e != nil {
			t.Fatal(e)
		}
		port := u.LocalAddr().(*net.UDPAddr).Port
		ln, e := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
		if e == nil {
			ln.Close()
			t.Cleanup(func() { u.Close() })
			return u, port
		}
		u.Close()
	}
	t.Fatal("no UDP port with a free TCP number")
	return nil, 0
}

func TestListenDNSPairRetriesUDPCollision(t *testing.T) {
	_, held := holdUDPPort(t)
	var ports []int
	ln, u, e := listenDNSPair(dnsListenAttempts, func() (net.Listener, error) {
		addr := "127.0.0.1:0"
		if len(ports) == 0 {
			addr = net.JoinHostPort("127.0.0.1", strconv.Itoa(held))
		}
		l, e := net.Listen("tcp", addr)
		if e == nil {
			ports = append(ports, l.Addr().(*net.TCPAddr).Port)
		}
		return l, e
	})
	if e != nil {
		t.Fatal(e)
	}
	defer ln.Close()
	defer u.Close()
	if len(ports) != 2 || ports[0] != held {
		t.Fatalf("TCP ports tried = %v, want the held port %d and then one more", ports, held)
	}
	got := ln.Addr().(*net.TCPAddr).Port
	if got == held || u.LocalAddr().(*net.UDPAddr).Port != got {
		t.Fatalf("TCP port %d, UDP %v, held %d", got, u.LocalAddr(), held)
	}
	// The colliding attempt's TCP listener was released.
	l, e := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(held)))
	if e != nil {
		t.Fatalf("held port's TCP listener not closed: %v", e)
	}
	l.Close()
}

func TestListenDNSPairGivesUpAfterBoundedAttempts(t *testing.T) {
	_, held := holdUDPPort(t)
	calls := 0
	_, _, e := listenDNSPair(3, func() (net.Listener, error) {
		calls++
		return net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(held)))
	})
	if e == nil {
		t.Fatal("collision on every attempt succeeded")
	}
	if calls != 3 {
		t.Fatalf("attempts = %d, want 3", calls)
	}
}
