package execution

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/task"
)

// Reviewed implementation input (DEV-645): the most bytes of a CONNECT
// response head, its status line and header fields with their terminating
// empty line, that the dialer reads. The public text requires a bounded
// exchange (REQ-PROXY-0017) and registers no value. A longer head fails the
// tunnel; it never enlarges an attempt bound.
const maxConnectResponse = 8192

// maxHostname is the longest normalized hostname a CONNECT authority carries.
const maxHostname = 253

// pathError is a path failure that ends the attempt with its reason. The
// proxy dialer returns it; the attempt scope turns it into the attempt's
// cancellation cause, so a runner's outcome built from the failed dial is not
// retained (REQ-PROXY-0029).
type pathError struct{ reason task.ExecutionReason }

func (e pathError) Error() string { return "execution: path failure: " + string(e.reason) }

// pathFailure returns the reason a dial error carries, if it is a path failure.
func pathFailure(err error) (task.ExecutionReason, bool) {
	var pe pathError
	if errors.As(err, &pe) {
		return pe.reason, true
	}
	return "", false
}

var (
	errProxyUnavailable = pathError{task.ProxyUnavailable}
	errProxyConnect     = pathError{task.ProxyConnectFailed}
	// errProxyNetwork refuses a dial HTTP CONNECT cannot carry. Admission
	// admits only proxy-compatible tasks, so it is an implementation failure.
	errProxyNetwork = pathError{task.InternalError}
	errProxyConfig  = errors.New("execution: invalid proxy endpoint")
)

// ProxyDialers is the proxy path's DialerSource: the attempt's snapshot of
// the host-configured local HTTP/1.1 CONNECT endpoint (REQ-EXECUTION-0009).
// Each attempt gets its own dialer and every stream of the attempt opens its
// own tunnel through the same endpoint, so no tunnel is pooled or reused
// across streams or attempts (REQ-PROXY-0020) and a redirect a test follows
// opens a new tunnel on the same snapshot (REQ-PROXY-0018). Nothing is ever
// dialed directly. The engine wraps the dialer in the attempt scope, which
// counts each tunnel as a connection before it is opened and meters only the
// bytes exchanged inside the established tunnel, so the CONNECT exchange is
// not counted (REQ-LIMIT-0033).
type ProxyDialers struct{ endpoint string }

// NewProxyDialers binds the endpoint of a validated host proxy configuration.
// It accepts only a loopback IP address or localhost and a port in 1 to 65535
// (REQ-PROXY-0005, REQ-PROXY-0007); the endpoint is never task input.
func NewProxyDialers(host string, port int) (ProxyDialers, error) {
	if port < 1 || port > 65535 {
		return ProxyDialers{}, errProxyConfig
	}
	if host != "localhost" {
		ip, err := netip.ParseAddr(host)
		if err != nil || ip.Zone() != "" || !ip.IsLoopback() {
			return ProxyDialers{}, errProxyConfig
		}
	}
	return ProxyDialers{endpoint: net.JoinHostPort(host, strconv.Itoa(port))}, nil
}

func (p ProxyDialers) Dialer(int) (measurement.Dialer, error) {
	if p.endpoint == "" {
		return nil, errProxyConfig
	}
	return proxyDialer{endpoint: p.endpoint}, nil
}

type proxyDialer struct{ endpoint string }

// DialContext opens one tunnel for one TCP stream (REQ-PROXY-0017): it
// connects to the loopback endpoint, sends a CONNECT request for the exact
// authority the test selected and returns the tunnel once the endpoint has
// answered with a valid HTTP/1.1 2xx response. The request carries only the
// request line and Host, so no credentials, cookies or other application
// headers are sent. A failed connection to the endpoint is proxy_unavailable;
// a refused, closed, malformed, oversized or interrupted CONNECT is
// proxy_connect_failed. A dial the proxy cannot carry is refused before any
// contact. The dial and the exchange end with ctx.
// requirement-ref: REQ-PROXY-0017
// requirement-ref: REQ-PROXY-0019
func (d proxyDialer) DialContext(ctx context.Context, network, address string) (measurement.Conn, error) {
	authority, ok := connectAuthority(network, address)
	if !ok {
		return nil, errProxyNetwork
	}
	var nd net.Dialer
	conn, err := nd.DialContext(ctx, "tcp", d.endpoint)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, errProxyUnavailable
	}
	// localhost must reach the loopback interface; anything else is not the
	// configured local endpoint.
	if a, ok := conn.RemoteAddr().(*net.TCPAddr); !ok || !a.IP.IsLoopback() {
		_ = conn.Close()
		return nil, errProxyUnavailable
	}
	tunnel, err := connect(ctx, conn, authority)
	if err != nil {
		_ = conn.Close()
		return nil, errProxyConnect
	}
	return tunnel, nil
}

// connect performs the CONNECT exchange on conn. Ending ctx interrupts it by
// expiring the connection's deadline.
func connect(ctx context.Context, conn net.Conn, authority string) (measurement.Conn, error) {
	stop := context.AfterFunc(ctx, func() { _ = conn.SetDeadline(time.Unix(1, 0)) })
	if _, err := io.WriteString(conn, "CONNECT "+authority+" HTTP/1.1\r\nHost: "+authority+"\r\n\r\n"); err != nil {
		stop()
		return nil, err
	}
	r := bufio.NewReader(io.LimitReader(conn, maxConnectResponse))
	err := readConnectResponse(r)
	if !stop() || ctx.Err() != nil {
		return nil, errProxyConnect
	}
	if err != nil {
		return nil, err
	}
	// Bytes the endpoint sent after the response head belong to the tunnel.
	ahead, _ := r.Peek(r.Buffered())
	return &tunnelConn{Conn: conn, ahead: append([]byte(nil), ahead...)}, nil
}

// readConnectResponse accepts only a syntactically valid HTTP/1.1 2xx status
// line followed by well-formed header fields and the empty line, each ended by
// CRLF, within the response bound. A 2xx response to CONNECT has no content,
// so framing header fields are ignored; no part of the response is retained
// (REQ-PROXY-0033).
func readConnectResponse(r *bufio.Reader) error {
	line, err := readLine(r)
	if err != nil || !validStatusLine(line) {
		return errProxyConnect
	}
	for {
		line, err := readLine(r)
		if err != nil {
			return errProxyConnect
		}
		if line == "" {
			return nil
		}
		if !validField(line) {
			return errProxyConnect
		}
	}
}

// readLine returns one CRLF-terminated line without its terminator. A bare LF
// or CR is malformed.
func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	line, ok := strings.CutSuffix(line, "\r\n")
	if !ok || strings.ContainsAny(line, "\r\n") {
		return "", errProxyConnect
	}
	return line, nil
}

// validField checks "name: value" with a token name, no whitespace before the
// colon or obsolete line folding, and a value of visible characters, spaces
// and tabs.
func validField(line string) bool {
	name, value, ok := strings.Cut(line, ":")
	if !ok || name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		if !tchar(name[i]) {
			return false
		}
	}
	for i := 0; i < len(value); i++ {
		if c := value[i]; c != '\t' && (c < ' ' || c == 0x7f) {
			return false
		}
	}
	return true
}

// tchar reports an RFC 9110 token character.
func tchar(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || digit(c) || strings.IndexByte("!#$%&'*+-.^_`|~", c) >= 0
}

// validStatusLine checks "HTTP/1.1 2xx" with an optional reason phrase of
// visible characters, spaces and tabs.
func validStatusLine(line string) bool {
	rest, ok := strings.CutPrefix(line, "HTTP/1.1 2")
	if !ok || len(rest) < 2 || !digit(rest[0]) || !digit(rest[1]) {
		return false
	}
	rest = rest[2:]
	if rest == "" {
		return true
	}
	if rest[0] != ' ' {
		return false
	}
	for i := 1; i < len(rest); i++ {
		if c := rest[i]; c != '\t' && (c < ' ' || c == 0x7f) {
			return false
		}
	}
	return true
}

func digit(c byte) bool { return c >= '0' && c <= '9' }

// connectAuthority returns the CONNECT authority of a TCP dial: the explicit
// IP address and port, or the normalized hostname and port, whose resolution
// then happens behind the proxy (REQ-PROXY-0019). Only "tcp" is carried: an
// address-family-specific or non-stream dial needs an operation HTTP CONNECT
// cannot provide. A hostname is lowercase letters, digits, hyphens and dots
// in non-empty labels, so nothing outside the authority can reach the request.
func connectAuthority(network, address string) (string, bool) {
	if network != "tcp" {
		return "", false
	}
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return "", false
	}
	p, err := strconv.Atoi(port)
	if err != nil || p < 1 || p > 65535 || strconv.Itoa(p) != port {
		return "", false
	}
	if ip, err := netip.ParseAddr(host); err == nil {
		if ip.Zone() != "" {
			return "", false
		}
		return netip.AddrPortFrom(ip, uint16(p)).String(), true
	}
	host = strings.ToLower(host)
	if !hostname(host) {
		return "", false
	}
	return net.JoinHostPort(host, port), true
}

func hostname(s string) bool {
	if s == "" || len(s) > maxHostname {
		return false
	}
	for _, label := range strings.Split(s, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for i := 0; i < len(label); i++ {
			if c := label[i]; !(c >= 'a' && c <= 'z' || digit(c) || c == '-') {
				return false
			}
		}
	}
	return true
}

// tunnelConn is the established tunnel. It first returns the tunnel bytes the
// endpoint sent with its response, then reads from the connection.
type tunnelConn struct {
	net.Conn
	ahead []byte
}

func (c *tunnelConn) Read(p []byte) (int, error) {
	if len(c.ahead) > 0 {
		n := copy(p, c.ahead)
		c.ahead = c.ahead[n:]
		return n, nil
	}
	return c.Conn.Read(p)
}
