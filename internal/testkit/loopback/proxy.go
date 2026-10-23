package loopback

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
)

// ProxyMode selects how a scripted Proxy answers a CONNECT request.
type ProxyMode int

const (
	// ProxyRelay answers with Head, sends Ahead and then relays the tunnel to
	// the requested authority, resolved through Resolve for a hostname.
	ProxyRelay ProxyMode = iota
	// ProxyReply writes Head and closes, whatever the request.
	ProxyReply
	// ProxyHold reads the request and never answers; it holds the
	// connection until the client closes it.
	ProxyHold
)

// maxProxyHead bounds the request head the scripted proxy reads.
const maxProxyHead = 8192

// ProxyConfig scripts a Proxy. Head defaults to a 200 response in relay mode.
// Resolve maps a CONNECT authority to the loopback address the proxy relays
// to, which stands in for resolution behind the proxy; an authority not in it
// is relayed only when it is a 127.0.0.1 address.
type ProxyConfig struct {
	Mode    ProxyMode
	Head    []byte
	Ahead   []byte
	Resolve map[string]string
}

// Proxy is a scripted local HTTP/1.1 CONNECT proxy on 127.0.0.1 for tests. It
// is not a production proxy and not MobileProxy. It records every request
// head it reads and counts the CONNECT exchange separately from the tunnel
// bytes it relays, so a test compares a client-side meter with the bytes that
// crossed the tunnel and shows the exchange itself was not metered.
type Proxy struct {
	ln  net.Listener
	cfg ProxyConfig
	wg  sync.WaitGroup

	mu                          sync.Mutex
	conns                       map[net.Conn]struct{}
	accepted, open              int
	requests                    []string
	headIn, headOut             int
	tunnelIn, tunnelOut, relays int
	changed                     chan struct{}
}

// ListenProxy starts a scripted proxy on an ephemeral 127.0.0.1 port.
func ListenProxy(cfg ProxyConfig) (*Proxy, error) {
	if cfg.Mode == ProxyRelay && cfg.Head == nil {
		cfg.Head = []byte("HTTP/1.1 200 Connection established\r\n\r\n")
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	p := &Proxy{ln: ln, cfg: cfg, conns: make(map[net.Conn]struct{}), changed: make(chan struct{})}
	p.wg.Add(1)
	go p.serve()
	return p, nil
}

// Host and Port are the proxy's listening endpoint.
func (p *Proxy) Host() string { return "127.0.0.1" }
func (p *Proxy) Port() int    { return p.ln.Addr().(*net.TCPAddr).Port }

func (p *Proxy) serve() {
	defer p.wg.Done()
	for {
		c, err := p.ln.Accept()
		if err != nil {
			return
		}
		p.mu.Lock()
		p.accepted++
		p.conns[c] = struct{}{}
		p.mu.Unlock()
		p.track(1)
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			defer p.track(-1)
			defer p.drop(c)
			p.handle(c)
		}()
	}
}

func (p *Proxy) handle(c net.Conn) {
	head, err := readHead(c)
	p.mu.Lock()
	p.headIn += len(head)
	if err == nil {
		p.requests = append(p.requests, string(head))
		close(p.changed)
		p.changed = make(chan struct{})
	}
	p.mu.Unlock()
	if err != nil {
		return
	}
	switch p.cfg.Mode {
	case ProxyHold:
		_, _ = io.Copy(io.Discard, c)
		return
	case ProxyReply:
		p.reply(c, p.cfg.Head)
		return
	}
	target, ok := p.target(head)
	if !ok {
		p.reply(c, []byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		return
	}
	var d net.Dialer
	up, err := d.Dial("tcp", target)
	if err != nil {
		p.reply(c, []byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		return
	}
	defer up.Close()
	p.reply(c, p.cfg.Head)
	p.mu.Lock()
	p.relays++
	p.mu.Unlock()
	if len(p.cfg.Ahead) > 0 {
		n, _ := c.Write(p.cfg.Ahead)
		p.count(0, n)
	}
	done := make(chan struct{}, 2)
	go func() { // client to target
		buf := make([]byte, 4096)
		for {
			n, err := c.Read(buf)
			if n > 0 {
				p.count(n, 0)
				if _, werr := up.Write(buf[:n]); werr != nil {
					break
				}
			}
			if err != nil {
				break
			}
		}
		_ = up.(*net.TCPConn).CloseWrite()
		done <- struct{}{}
	}()
	go func() { // target to client
		buf := make([]byte, 4096)
		for {
			n, err := up.Read(buf)
			if n > 0 {
				if _, werr := c.Write(buf[:n]); werr != nil {
					break
				}
				p.count(0, n)
			}
			if err != nil {
				break
			}
		}
		_ = c.Close()
		done <- struct{}{}
	}()
	<-done
	_ = c.Close()
	_ = up.Close()
	<-done
}

func (p *Proxy) reply(c net.Conn, head []byte) {
	n, _ := c.Write(head)
	p.mu.Lock()
	p.headOut += n
	p.mu.Unlock()
}

func (p *Proxy) count(in, out int) {
	p.mu.Lock()
	p.tunnelIn += in
	p.tunnelOut += out
	p.mu.Unlock()
}

// target returns the relay address of a "CONNECT authority HTTP/1.1" head.
func (p *Proxy) target(head []byte) (string, bool) {
	line, _, _ := strings.Cut(string(head), "\r\n")
	f := strings.Split(line, " ")
	if len(f) != 3 || f[0] != "CONNECT" || f[2] != "HTTP/1.1" {
		return "", false
	}
	if t, ok := p.cfg.Resolve[f[1]]; ok {
		return t, true
	}
	if host, _, err := net.SplitHostPort(f[1]); err == nil && host == "127.0.0.1" {
		return f[1], true
	}
	return "", false
}

// readHead reads a request head through its empty line, one byte at a time
// so no tunnel byte is consumed.
func readHead(c net.Conn) ([]byte, error) {
	var head []byte
	b := make([]byte, 1)
	for len(head) < maxProxyHead {
		if _, err := c.Read(b); err != nil {
			return head, err
		}
		head = append(head, b[0])
		if bytes.HasSuffix(head, []byte("\r\n\r\n")) {
			return head, nil
		}
	}
	return head, errors.New("loopback: request head too long")
}

func (p *Proxy) drop(c net.Conn) {
	_ = c.Close()
	p.mu.Lock()
	delete(p.conns, c)
	p.mu.Unlock()
}

func (p *Proxy) track(delta int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.open += delta
	close(p.changed)
	p.changed = make(chan struct{})
}

// WaitIdle blocks until every accepted connection has been closed and its
// counts are final, or ctx ends.
func (p *Proxy) WaitIdle(ctx context.Context) error {
	for {
		p.mu.Lock()
		open, changed := p.open, p.changed
		p.mu.Unlock()
		if open == 0 {
			return nil
		}
		select {
		case <-changed:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// WaitRequests blocks until the proxy has read n request heads or ctx ends.
func (p *Proxy) WaitRequests(ctx context.Context, n int) error {
	for {
		p.mu.Lock()
		got, changed := len(p.requests), p.changed
		p.mu.Unlock()
		if got >= n {
			return nil
		}
		select {
		case <-changed:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// ProxyCounts are the scripted proxy's own counts.
type ProxyCounts struct {
	Accepted, Relays    int
	HeadIn, HeadOut     int // the CONNECT exchange
	TunnelIn, TunnelOut int // bytes relayed inside established tunnels
	Requests            []string
}

// Counts reports what the proxy accepted, read and wrote.
func (p *Proxy) Counts() ProxyCounts {
	p.mu.Lock()
	defer p.mu.Unlock()
	return ProxyCounts{Accepted: p.accepted, Relays: p.relays, HeadIn: p.headIn, HeadOut: p.headOut,
		TunnelIn: p.tunnelIn, TunnelOut: p.tunnelOut, Requests: append([]string(nil), p.requests...)}
}

// Close stops accepting, closes every connection and waits for them.
func (p *Proxy) Close() error {
	err := p.ln.Close()
	p.mu.Lock()
	for c := range p.conns {
		_ = c.Close()
	}
	p.mu.Unlock()
	p.wg.Wait()
	return err
}
