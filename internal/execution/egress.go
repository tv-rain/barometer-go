package execution

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"mime"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"time"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/networkcontext"
)

var errContextResponse = errors.New("execution: invalid context endpoint response")

// ContextFetch binds approved Network Context endpoints to a path dialer and
// the default address policy. Proxy hostname endpoints are refused before
// resolution or contact; the engine also applies the attempt's trusted policy.
// Lookup traffic remains outside the measurement byte/connection meter.
func ContextFetch(d measurement.Dialer) networkcontext.Fetch {
	return contextFetcher(d, PublicAddresses{})
}

func contextFetcher(d measurement.Dialer, policy AddressPolicy) networkcontext.Fetch {
	return func(ctx context.Context, u *url.URL, roots *x509.CertPool) ([]byte, error) {
		return fetchContext(ctx, d, u, roots, policy)
	}
}

// fetchContext makes one HTTPS GET to u. The request carries only the
// request line, Host and Accept: no User-Agent, cookie, authorization,
// identifier, address, target or connection_code (REQ-CONTEXT-0012). There
// is no redirect, compression, connection reuse or environment proxy
// setting, and the connection is closed afterwards. Only a 200 response with
// an uncompressed application/json body of at most
// networkcontext.MaxResponseBytes is valid; reading stops at the bound
// (REQ-CONTEXT-0015). The status line and headers are bounded separately by
// networkcontext.MaxHeaderBytes. Everything ends with ctx.
// requirement-ref: REQ-CONTEXT-0010
// requirement-ref: REQ-CONTEXT-0011
// requirement-ref: REQ-CONTEXT-0012
// requirement-ref: REQ-CONTEXT-0015
func fetchContext(ctx context.Context, d measurement.Dialer, u *url.URL, roots *x509.CertPool, policy AddressPolicy) ([]byte, error) {
	if d == nil || u == nil || u.Scheme != "https" || u.Hostname() == "" {
		return nil, errContextResponse
	}
	port := u.Port()
	if port == "" {
		port = "443"
	}
	host := u.Hostname()
	ip, parseErr := netip.ParseAddr(host)
	ips := []netip.Addr{ip}
	if parseErr != nil {
		if _, proxy := d.(proxyDialer); proxy {
			return nil, errContextResponse
		}
		var err error
		ips, err = net.DefaultResolver.LookupNetIP(ctx, "ip", host)
		if err != nil {
			return nil, errContextResponse
		}
	}
	if len(ips) == 0 || len(ips) > 64 || ctx.Err() != nil {
		return nil, errContextResponse
	}
	for _, a := range ips {
		if policy == nil || !policy.Allows(a.Unmap()) {
			return nil, errContextResponse
		}
	}
	raw, err := d.DialContext(ctx, "tcp", net.JoinHostPort(ips[0].Unmap().String(), port))
	if err != nil {
		return nil, errContextResponse
	}
	conn := tls.Client(asNetConn(raw), &tls.Config{ServerName: u.Hostname(), RootCAs: roots, MinVersion: tls.VersionTLS12, NextProtos: []string{"http/1.1"}})
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { _ = raw.SetDeadline(time.Unix(1, 0)) })
	defer stop()
	if conn.HandshakeContext(ctx) != nil {
		return nil, errContextResponse
	}
	path := u.EscapedPath()
	if path == "" {
		path = "/"
	}
	req := &http.Request{Method: http.MethodGet, URL: &url.URL{Path: u.Path, RawPath: path, RawQuery: u.RawQuery}, Host: u.Host,
		Proto: "HTTP/1.1", ProtoMajor: 1, ProtoMinor: 1,
		// An empty User-Agent suppresses the default one.
		Header: http.Header{"Accept": {"application/json"}, "User-Agent": {""}}}
	if req.Write(conn) != nil {
		return nil, errContextResponse
	}
	in := &io.LimitedReader{R: conn, N: networkcontext.MaxHeaderBytes}
	resp, err := http.ReadResponse(bufio.NewReader(in), req)
	if err != nil {
		return nil, errContextResponse
	}
	defer resp.Body.Close()
	media, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if resp.StatusCode != http.StatusOK || err != nil || media != "application/json" ||
		resp.Header.Get("Content-Encoding") != "" || resp.ContentLength > networkcontext.MaxResponseBytes {
		return nil, errContextResponse
	}
	// The buffered reader may already hold the start of the body; from here
	// the body has its own bound.
	in.N = networkcontext.MaxResponseBytes + 1
	body, err := io.ReadAll(io.LimitReader(resp.Body, networkcontext.MaxResponseBytes+1))
	if err != nil || len(body) > networkcontext.MaxResponseBytes || ctx.Err() != nil {
		return nil, errContextResponse
	}
	return body, nil
}

// asNetConn returns the dialer's connection as a net.Conn for TLS. Path
// dialers return one already; a connection that is not exposes no address and
// applies every deadline to both directions.
func asNetConn(c measurement.Conn) net.Conn {
	if n, ok := c.(net.Conn); ok {
		return n
	}
	return plainConn{c}
}

type plainConn struct{ measurement.Conn }

func (plainConn) LocalAddr() net.Addr                  { return noAddr{} }
func (plainConn) RemoteAddr() net.Addr                 { return noAddr{} }
func (c plainConn) SetReadDeadline(t time.Time) error  { return c.SetDeadline(t) }
func (c plainConn) SetWriteDeadline(t time.Time) error { return c.SetDeadline(t) }

type noAddr struct{}

func (noAddr) Network() string { return "tcp" }
func (noAddr) String() string  { return "" }
