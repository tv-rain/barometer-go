package loopback

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"io"
	"math/big"
	"net"
	"slices"
	"sync"
	"time"
)

// CertificateCanary is placed in the subject, issuer and SAN-adjacent fields
// of every synthetic certificate, so tests can assert that no certificate
// data is retained.
const CertificateCanary = "barometer-canary-certificate"

// TestCA is a synthetic certificate authority for loopback TLS servers. It is
// never a production trust anchor.
type TestCA struct {
	cert *x509.Certificate
	key  *ecdsa.PrivateKey
	pool *x509.CertPool
}

// NewTestCA creates a fresh synthetic CA.
func NewTestCA() (*TestCA, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(91337), Subject: pkix.Name{CommonName: CertificateCanary + " CA", Organization: []string{CertificateCanary}},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(24 * time.Hour),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	pool.AddCert(cert)
	return &TestCA{cert: cert, key: key, pool: pool}, nil
}

// Pool is the trust pool holding only this CA.
func (c *TestCA) Pool() *x509.CertPool { return c.pool }

// Issue returns a leaf for name, expired when expired is set.
func (c *TestCA) Issue(name string, expired bool) (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}
	notBefore, notAfter := time.Now().Add(-time.Hour), time.Now().Add(24*time.Hour)
	if expired {
		notBefore, notAfter = time.Now().Add(-48*time.Hour), time.Now().Add(-24*time.Hour)
	}
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(424242), Subject: pkix.Name{CommonName: name, Organization: []string{CertificateCanary}}, DNSNames: []string{name},
		NotBefore: notBefore, NotAfter: notAfter,
		KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, c.cert, &key.PublicKey, c.key)
	if err != nil {
		return tls.Certificate{}, err
	}
	return tls.Certificate{Certificate: [][]byte{der, c.cert.Raw}, PrivateKey: key}, nil
}

// TLSMode scripts a loopback TLS server's answer to a ClientHello.
type TLSMode int

const (
	// TLSAccept completes the handshake and then reads plaintext until the
	// client closes.
	TLSAccept TLSMode = iota
	// TLSAlert reads the first record and answers with one fatal alert.
	TLSAlert
	// TLSStall reads and never answers until the client closes.
	TLSStall
	// TLSReset reads the first record and resets the connection.
	TLSReset
	// TLSGarbage reads the first record and answers with non-TLS bytes.
	TLSGarbage
)

// TLSConfig configures a TLSServer. Certificate is presented in TLSAccept
// mode; ALPN lists the server's protocols; Alert is the TLSAlert description.
type TLSConfig struct {
	Mode        TLSMode
	Certificate tls.Certificate
	ALPN        []string
	Alert       uint8
}

// Hello is what a TLSServer observed of one ClientHello in TLSAccept mode.
type Hello struct {
	ServerName string
	ALPN       []string
}

// TLSServer is a scripted TLS endpoint on 127.0.0.1. It counts the bytes on
// each socket and the plaintext the client sends after the handshake.
type TLSServer struct {
	ln        net.Listener
	cfg       TLSConfig
	wg        sync.WaitGroup
	mu        sync.Mutex
	accepted  int
	open      int
	in, out   int
	plaintext int
	hellos    []Hello
	changed   chan struct{}
}

// ListenTLS starts a scripted TLS server on an ephemeral 127.0.0.1 port.
func ListenTLS(cfg TLSConfig) (*TLSServer, error) {
	if cfg.Mode < TLSAccept || cfg.Mode > TLSGarbage || cfg.Mode == TLSAccept && len(cfg.Certificate.Certificate) == 0 {
		return nil, errors.New("loopback: invalid TLS server configuration")
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	s := &TLSServer{ln: ln, cfg: cfg, changed: make(chan struct{})}
	s.wg.Add(1)
	go s.serve()
	return s, nil
}

func (s *TLSServer) serve() {
	defer s.wg.Done()
	for {
		c, err := s.ln.Accept()
		if err != nil {
			return
		}
		s.mu.Lock()
		s.accepted++
		s.mu.Unlock()
		s.track(1)
		s.wg.Add(1)
		go func(raw net.Conn) {
			defer s.wg.Done()
			defer s.track(-1)
			defer raw.Close()
			s.handle(raw)
		}(c)
	}
}

func (s *TLSServer) handle(raw net.Conn) {
	c := &tlsCounted{Conn: raw, s: s}
	if s.cfg.Mode != TLSAccept {
		if s.cfg.Mode == TLSStall {
			_, _ = io.Copy(io.Discard, c)
			return
		}
		buf := make([]byte, 5)
		if _, err := io.ReadFull(c, buf); err != nil {
			return
		}
		switch s.cfg.Mode {
		case TLSAlert:
			_, _ = c.Write([]byte{21, 3, 3, 0, 2, 2, s.cfg.Alert})
		case TLSReset:
			if t, ok := raw.(*net.TCPConn); ok {
				_ = t.SetLinger(0)
			}
		case TLSGarbage:
			_, _ = c.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		}
		return
	}
	conn := tls.Server(c, &tls.Config{Certificates: []tls.Certificate{s.cfg.Certificate}, NextProtos: s.cfg.ALPN, MinVersion: tls.VersionTLS12,
		GetConfigForClient: func(h *tls.ClientHelloInfo) (*tls.Config, error) {
			s.mu.Lock()
			s.hellos = append(s.hellos, Hello{ServerName: h.ServerName, ALPN: slices.Clone(h.SupportedProtos)})
			s.mu.Unlock()
			return nil, nil
		}})
	if conn.Handshake() != nil {
		return
	}
	n, _ := io.Copy(io.Discard, conn)
	s.mu.Lock()
	s.plaintext += int(n)
	s.signal()
	s.mu.Unlock()
}

type tlsCounted struct {
	net.Conn
	s *TLSServer
}

func (c *tlsCounted) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	c.s.mu.Lock()
	c.s.in += n
	c.s.signal()
	c.s.mu.Unlock()
	return n, err
}

func (c *tlsCounted) Write(p []byte) (int, error) {
	n, err := c.Conn.Write(p)
	c.s.mu.Lock()
	c.s.out += n
	c.s.mu.Unlock()
	return n, err
}

func (s *TLSServer) track(delta int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.open += delta
	s.signal()
}

// signal wakes waiters; the caller holds mu.
func (s *TLSServer) signal() {
	close(s.changed)
	s.changed = make(chan struct{})
}

// Address is the server's 127.0.0.1 host:port.
func (s *TLSServer) Address() string { return s.ln.Addr().String() }

// TLSCounts reports accepted and open connections, socket bytes in each direction,
// plaintext received after the handshake and observed ClientHellos.
type TLSCounts struct {
	Accepted, Open, In, Out, Plaintext int
	Hellos                             []Hello
}

func (s *TLSServer) Counts() TLSCounts {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.counts()
}

func (s *TLSServer) counts() TLSCounts {
	return TLSCounts{Accepted: s.accepted, Open: s.open, In: s.in, Out: s.out, Plaintext: s.plaintext, Hellos: slices.Clone(s.hellos)}
}

// WaitIdle waits until every accepted connection has closed.
func (s *TLSServer) WaitIdle(ctx context.Context) error {
	return s.Wait(ctx, func(c TLSCounts) bool { return c.Open == 0 })
}

// Wait waits until ready reports true for the current counts.
func (s *TLSServer) Wait(ctx context.Context, ready func(TLSCounts) bool) error {
	for {
		s.mu.Lock()
		counts, changed := s.counts(), s.changed
		s.mu.Unlock()
		if ready(counts) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-changed:
		}
	}
}

// Close stops the server and waits for its connections to end.
func (s *TLSServer) Close() error {
	err := s.ln.Close()
	s.wg.Wait()
	return err
}
