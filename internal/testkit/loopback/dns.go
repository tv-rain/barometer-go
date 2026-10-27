package loopback

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

// DNSServer is a bounded loopback-only UDP/TCP fixture. Reply is configured
// before ListenDNS and never mutated; up to eight frames are sent per query.
// Close actively releases every accepted stream before waiting for workers.
type DNSServer struct {
	udp                     *net.UDPConn
	tcp                     net.Listener
	reply                   func([]byte) [][]byte
	mu                      sync.Mutex
	wg                      sync.WaitGroup
	closed                  bool
	conns                   map[net.Conn]bool
	queries, sent, received int
}

func ListenDNS(reply func([]byte) [][]byte) (*DNSServer, error) {
	if reply == nil {
		return nil, errors.New("dns fixture: missing reply")
	}
	ln, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		return nil, e
	}
	u, e := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: ln.Addr().(*net.TCPAddr).Port})
	if e != nil {
		ln.Close()
		return nil, e
	}
	s := &DNSServer{udp: u, tcp: ln, reply: reply, conns: make(map[net.Conn]bool)}
	s.wg.Add(2)
	go s.serveUDP()
	go s.serveTCP()
	return s, nil
}
func (s *DNSServer) Address() string { return s.tcp.Addr().String() }
func (s *DNSServer) Counts() (queries, sent, received int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.queries, s.sent, s.received
}
func (s *DNSServer) count(sent, received int, query bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sent += sent
	s.received += received
	if query {
		s.queries++
	}
}
func (s *DNSServer) serveUDP() {
	defer s.wg.Done()
	b := make([]byte, 65536)
	for {
		n, peer, e := s.udp.ReadFromUDP(b)
		if e != nil {
			return
		}
		s.count(0, n, true)
		for i, r := range s.reply(append([]byte(nil), b[:n]...)) {
			if i >= 8 || len(r) > 65535 {
				break
			}
			n, _ := s.udp.WriteToUDP(r, peer)
			s.count(n, 0, false)
		}
	}
}
func (s *DNSServer) serveTCP() {
	defer s.wg.Done()
	for {
		c, e := s.tcp.Accept()
		if e != nil {
			return
		}
		s.mu.Lock()
		if s.closed {
			s.mu.Unlock()
			c.Close()
			return
		}
		s.conns[c] = true
		s.wg.Add(1)
		s.mu.Unlock()
		go s.stream(c)
	}
}
func (s *DNSServer) stream(c net.Conn) {
	defer s.wg.Done()
	defer c.Close()
	defer func() { s.mu.Lock(); delete(s.conns, c); s.mu.Unlock() }()
	_ = c.SetDeadline(time.Now().Add(5 * time.Second))
	prefix := make([]byte, 2)
	n, e := io.ReadFull(c, prefix)
	s.count(0, n, false)
	if e != nil {
		return
	}
	sz := int(binary.BigEndian.Uint16(prefix))
	if sz < 12 {
		return
	}
	q := make([]byte, sz)
	n, e = io.ReadFull(c, q)
	s.count(0, n, true)
	if e != nil {
		return
	}
	responses := s.reply(q)
	if len(responses) == 0 {
		return
	}
	r := responses[0]
	if len(r) > 65535 {
		return
	}
	frame := append([]byte{byte(len(r) >> 8), byte(len(r))}, r...)
	n, _ = c.Write(frame)
	s.count(n, 0, false)
}
func (s *DNSServer) Close() error {
	s.mu.Lock()
	s.closed = true
	for c := range s.conns {
		_ = c.Close()
	}
	s.mu.Unlock()
	_ = s.udp.Close()
	e := s.tcp.Close()
	s.wg.Wait()
	return e
}

// DNSRecord creates only synthetic A/AAAA/CNAME answer records. Empty Owner
// uses the question name. Values and names are never deployment data.
type DNSRecord struct{ Owner, Type, Value string }

// DNSResponse encodes a 12-bit effective RCODE, adding an empty EDNS OPT
// record when its extended bits are nonzero. Larger codes are not wire values.
func DNSResponse(query []byte, code uint16, records []DNSRecord) []byte {
	if code > 4095 {
		return nil
	}
	var p dnsmessage.Parser
	h, e := p.Start(query)
	if e != nil {
		return nil
	}
	q, e := p.Question()
	if e != nil {
		return nil
	}
	b := dnsmessage.NewBuilder(nil, dnsmessage.Header{ID: h.ID, Response: true, RecursionDesired: true, RecursionAvailable: true, RCode: dnsmessage.RCode(code & 15)})
	b.EnableCompression()
	if b.StartQuestions() != nil || b.Question(q) != nil || b.StartAnswers() != nil {
		return nil
	}
	for _, r := range records {
		owner := q.Name
		if r.Owner != "" {
			owner, e = dnsmessage.NewName(r.Owner + ".")
			if e != nil {
				return nil
			}
		}
		rh := dnsmessage.ResourceHeader{Name: owner, Class: dnsmessage.ClassINET, TTL: 60}
		switch r.Type {
		case "A":
			ip := net.ParseIP(r.Value).To4()
			if ip == nil {
				return nil
			}
			var a [4]byte
			copy(a[:], ip)
			e = b.AResource(rh, dnsmessage.AResource{A: a})
		case "AAAA":
			ip := net.ParseIP(r.Value).To16()
			if ip == nil {
				return nil
			}
			var a [16]byte
			copy(a[:], ip)
			e = b.AAAAResource(rh, dnsmessage.AAAAResource{AAAA: a})
		case "CNAME":
			var n dnsmessage.Name
			n, e = dnsmessage.NewName(r.Value + ".")
			if e == nil {
				e = b.CNAMEResource(rh, dnsmessage.CNAMEResource{CNAME: n})
			}
		default:
			return nil
		}
		if e != nil {
			return nil
		}
	}
	if code > 15 {
		if b.StartAdditionals() != nil || b.OPTResource(dnsmessage.ResourceHeader{Name: dnsmessage.MustNewName("."), Type: dnsmessage.TypeOPT, Class: 4096, TTL: uint32(code>>4) << 24}, dnsmessage.OPTResource{}) != nil {
			return nil
		}
	}
	out, e := b.Finish()
	if e != nil {
		return nil
	}
	return out
}
