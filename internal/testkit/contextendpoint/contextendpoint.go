// Package contextendpoint is a test-only synthetic Network Context endpoint:
// HTTPS servers on 127.0.0.1 whose certificates come from a synthetic CA,
// with scripted responses and capture of every request, and a dialer that
// routes synthetic endpoint names to them. No real endpoint is contacted and
// no real classification is made. Production code must not import it.
package contextendpoint

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"math/big"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/tv-rain/barometer-go/internal/measurement"
)

// CA is a synthetic certificate authority.
type CA struct {
	cert *x509.Certificate
	key  *ecdsa.PrivateKey
	pool *x509.CertPool
}

// NewCA creates a fresh synthetic CA.
func NewCA() (*CA, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "barometer synthetic test CA"},
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
	return &CA{cert: cert, key: key, pool: pool}, nil
}

// Pool is the trust pool holding only this CA.
func (c *CA) Pool() *x509.CertPool { return c.pool }

func (c *CA) issue(name string) (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}
	serial, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		return tls.Certificate{}, err
	}
	tmpl := &x509.Certificate{SerialNumber: serial, Subject: pkix.Name{CommonName: name}, DNSNames: []string{name},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(24 * time.Hour),
		KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, c.cert, &key.PublicKey, c.key)
	if err != nil {
		return tls.Certificate{}, err
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}, nil
}

// Response scripts the server's answer. The zero value is a 200
// application/json response with an empty body.
type Response struct {
	Status      int               // 0 selects 200
	ContentType string            // "" selects application/json
	Header      map[string]string // extra response header fields
	Body        string
	Hold        bool // never answer; wait until the client goes away
}

// Request is one captured request. Remote is the source address the server
// observed, which a real endpoint would classify.
type Request struct {
	Method, Path, RawQuery, Host, Remote string
	Header                               http.Header
}

// Server is one synthetic HTTPS endpoint for a DNS name.
type Server struct {
	ln    net.Listener
	srv   *http.Server
	mu    sync.Mutex
	resp  Response
	queue []Response // answered first, one per request
	reqs  []Request
	raw   []byte
}

// Listen starts a server on an ephemeral 127.0.0.1 port whose certificate
// names name and is issued by ca.
func Listen(ca *CA, name string, r Response) (*Server, error) {
	cert, err := ca.issue(name)
	if err != nil {
		return nil, err
	}
	tcp, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	s := &Server{resp: r}
	// The recorder wraps the TLS listener, so it captures decrypted bytes.
	s.ln = recording{Listener: tls.NewListener(tcp, &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}), s: s}
	s.srv = &http.Server{Handler: http.HandlerFunc(s.handle), ReadHeaderTimeout: 10 * time.Second}
	go func() { _ = s.srv.Serve(s.ln) }()
	return s, nil
}

// Address is the server's 127.0.0.1 address and port.
func (s *Server) Address() string { return s.ln.Addr().String() }

// Set replaces the scripted response.
func (s *Server) Set(r Response) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resp = r
}

// Queue makes the next requests receive rs in order before the scripted
// response applies again.
func (s *Server) Queue(rs ...Response) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queue = append(s.queue, rs...)
}

// Requests returns the captured requests in order.
func (s *Server) Requests() []Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Request(nil), s.reqs...)
}

// Raw returns every decrypted byte the server read, so a test can assert
// what the request carried.
func (s *Server) Raw() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]byte(nil), s.raw...)
}

// Close stops the server.
func (s *Server) Close() error { return s.srv.Close() }

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.reqs = append(s.reqs, Request{Method: r.Method, Path: r.URL.Path, RawQuery: r.URL.RawQuery, Host: r.Host, Remote: r.RemoteAddr, Header: r.Header.Clone()})
	resp := s.resp
	if len(s.queue) > 0 {
		resp, s.queue = s.queue[0], s.queue[1:]
	}
	s.mu.Unlock()
	if resp.Hold {
		<-r.Context().Done()
		return
	}
	for k, v := range resp.Header {
		w.Header().Set(k, v)
	}
	ct := resp.ContentType
	if ct == "" {
		ct = "application/json"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Content-Length", strconv.Itoa(len(resp.Body)))
	status := resp.Status
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	_, _ = w.Write([]byte(resp.Body))
}

type recording struct {
	net.Listener
	s *Server
}

func (l recording) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return recordingConn{Conn: c, s: l.s}, nil
}

type recordingConn struct {
	net.Conn
	s *Server
}

func (c recordingConn) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	c.s.mu.Lock()
	c.s.raw = append(c.s.raw, p[:n]...)
	c.s.mu.Unlock()
	return n, err
}

// Routes is a direct-path stand-in that resolves synthetic names to
// loopback servers. It dials only the authorities it maps and records every
// requested address, so a test can assert which names the direct path
// contacted. It is both a measurement.Dialer and a dialer source.
type Routes struct {
	mu     sync.Mutex
	routes map[string]string
	dialed []string
}

var errUnrouted = errors.New("contextendpoint: unrouted address")

// NewRoutes maps "name:port" authorities to loopback addresses.
func NewRoutes(m map[string]string) *Routes {
	r := &Routes{routes: map[string]string{}}
	for k, v := range m {
		r.routes[k] = v
	}
	return r
}

// Add maps one more authority.
func (r *Routes) Add(authority, address string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes[authority] = address
}

func (r *Routes) DialContext(ctx context.Context, network, address string) (measurement.Conn, error) {
	r.mu.Lock()
	r.dialed = append(r.dialed, address)
	to, ok := r.routes[address]
	r.mu.Unlock()
	if !ok {
		if host, _, err := net.SplitHostPort(address); err != nil || host != "127.0.0.1" {
			return nil, errUnrouted
		}
		to = address
	}
	var d net.Dialer
	return d.DialContext(ctx, network, to)
}

// Dialer returns r itself for every attempt.
func (r *Routes) Dialer(int) (measurement.Dialer, error) { return r, nil }

// Dialed returns every address requested through r, in order.
func (r *Routes) Dialed() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.dialed...)
}
