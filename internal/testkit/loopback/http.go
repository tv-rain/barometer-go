package loopback

import (
	"bufio"
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"slices"
	"sync"
)

// HTTPConfig scripts one synthetic response per connection. Raw includes the
// status line, headers and body, permitting malformed and oversized fixtures.
// Stall leaves the connection open after Raw until the client closes it.
type HTTPConfig struct {
	Certificate tls.Certificate
	Raw         []byte
	Stall       bool
	// Plain serves plaintext HTTP without TLS, for NET-03 http validation
	// URLs; wire and plaintext counters then see the same bytes.
	Plain bool
	// DoH, when set, replaces Raw: it receives the request body (a DNS
	// query for DNS-03) and returns the complete raw response verbatim, so a
	// test scripts status, headers and payload. Nil bytes write nothing.
	DoH func(body []byte) []byte
}
type HTTPRequest struct {
	Method, Target string
	Header         http.Header
	Body           []byte
}
type HTTPCounts struct {
	Accepted, Open, In, Out, WireIn, WireOut int
	Requests                                 []HTTPRequest
	// Hellos records the ClientHello server name and ALPN offer.
	Hellos []Hello
}

// HTTPServer is a loopback-only scripted HTTPS (or, with Plain, HTTP) endpoint with separate TLS
// socket and plaintext counters. Every connection is closed by Close, even
// when a client stalls. Tests never contact an external endpoint.
type HTTPServer struct {
	ln      net.Listener
	cfg     HTTPConfig
	mu      sync.Mutex
	wg      sync.WaitGroup
	conns   map[net.Conn]bool
	counts  HTTPCounts
	closed  bool
	changed chan struct{}
}

func ListenHTTP(cfg HTTPConfig) (*HTTPServer, error) {
	ln, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		return nil, e
	}
	s := &HTTPServer{ln: ln, cfg: cfg, conns: map[net.Conn]bool{}, changed: make(chan struct{})}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			raw, e := ln.Accept()
			if e != nil {
				return
			}
			s.mu.Lock()
			if s.closed {
				s.mu.Unlock()
				_ = raw.Close()
				return
			}
			s.conns[raw] = true
			s.counts.Accepted++
			s.counts.Open++
			s.signalHTTP()
			s.mu.Unlock()
			s.wg.Add(1)
			go s.serveHTTP(raw)
		}
	}()
	return s, nil
}
func (s *HTTPServer) serveHTTP(raw net.Conn) {
	defer s.wg.Done()
	defer func() {
		_ = raw.Close()
		s.mu.Lock()
		delete(s.conns, raw)
		s.counts.Open--
		s.signalHTTP()
		s.mu.Unlock()
	}()
	wire := httpCounted{raw, s, true}
	var conn net.Conn = wire
	if !s.cfg.Plain {
		t := tls.Server(wire, &tls.Config{Certificates: []tls.Certificate{s.cfg.Certificate}, NextProtos: []string{"http/1.1"}, MinVersion: tls.VersionTLS12,
			GetConfigForClient: func(h *tls.ClientHelloInfo) (*tls.Config, error) {
				s.mu.Lock()
				s.counts.Hellos = append(s.counts.Hellos, Hello{ServerName: h.ServerName, ALPN: slices.Clone(h.SupportedProtos)})
				s.mu.Unlock()
				return nil, nil
			}})
		if t.Handshake() != nil {
			return
		}
		conn = t
	}
	plain := httpCounted{conn, s, false}
	r, e := http.ReadRequest(bufio.NewReader(plain))
	if e != nil {
		return
	}
	body, e := io.ReadAll(io.LimitReader(r.Body, 65536))
	if e != nil {
		return
	}
	s.mu.Lock()
	s.counts.Requests = append(s.counts.Requests, HTTPRequest{r.Method, r.RequestURI, r.Header.Clone(), body})
	s.signalHTTP()
	s.mu.Unlock()
	reply := s.cfg.Raw
	if s.cfg.DoH != nil {
		reply = s.cfg.DoH(body)
	}
	_, _ = plain.Write(reply)
	if s.cfg.Stall {
		_, _ = io.Copy(io.Discard, plain)
	}
}

type httpCounted struct {
	net.Conn
	s    *HTTPServer
	wire bool
}

func (c httpCounted) Read(p []byte) (int, error) {
	n, e := c.Conn.Read(p)
	c.s.mu.Lock()
	if c.wire {
		c.s.counts.WireIn += n
	} else {
		c.s.counts.In += n
	}
	c.s.signalHTTP()
	c.s.mu.Unlock()
	return n, e
}
func (c httpCounted) Write(p []byte) (int, error) {
	n, e := c.Conn.Write(p)
	c.s.mu.Lock()
	if c.wire {
		c.s.counts.WireOut += n
	} else {
		c.s.counts.Out += n
	}
	c.s.signalHTTP()
	c.s.mu.Unlock()
	return n, e
}
func (s *HTTPServer) Address() string    { return s.ln.Addr().String() }
func (s *HTTPServer) Counts() HTTPCounts { s.mu.Lock(); defer s.mu.Unlock(); return s.snapshotHTTP() }
func (s *HTTPServer) snapshotHTTP() HTTPCounts {
	c := s.counts
	c.Requests = append([]HTTPRequest(nil), c.Requests...)
	c.Hellos = slices.Clone(c.Hellos)
	return c
}
func (s *HTTPServer) signalHTTP() { close(s.changed); s.changed = make(chan struct{}) }
func (s *HTTPServer) Wait(ctx context.Context, f func(HTTPCounts) bool) error {
	for {
		s.mu.Lock()
		c, ch := s.snapshotHTTP(), s.changed
		s.mu.Unlock()
		if f(c) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ch:
		}
	}
}
func (s *HTTPServer) Close() error {
	e := s.ln.Close()
	// Late accepts observe closed under the same lock and close themselves.
	s.mu.Lock()
	s.closed = true
	for c := range s.conns {
		_ = c.Close()
	}
	s.mu.Unlock()
	s.wg.Wait()
	return e
}
