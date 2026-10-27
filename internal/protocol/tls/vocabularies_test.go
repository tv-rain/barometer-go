package tls_test

import (
	"context"
	stdtls "crypto/tls"
	"encoding/json"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	protocol "github.com/tv-rain/barometer-go/internal/protocol/tls"
	"github.com/tv-rain/barometer-go/internal/testkit"
)

type failingStream struct {
	err    error
	closed bool
}

func (s *failingStream) Read([]byte) (int, error)    { return 0, s.err }
func (s *failingStream) Write(b []byte) (int, error) { return len(b), nil }
func (s *failingStream) Close() error                { s.closed = true; return nil }
func (*failingStream) SetDeadline(time.Time) error   { return nil }

// Typed API fault injection through the real handshake: error text or a local
// alert value cannot impersonate crypto/tls's received-peer-alert evidence.
// requirement-ref: REQ-OBS-0045
// requirement-ref: REQ-OBS-0050
func TestHandshakeDoesNotInferPeerAlertFromErrors(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
	}{
		{"raw text", errors.New("PRIVATE-ERROR-CANARY remote error: tls: bad certificate")},
		{"opaque remote error", &net.OpError{Op: "remote error", Err: errors.New("PRIVATE-ERROR-CANARY tls: alert 42")}},
		{"local alert type", stdtls.AlertError(42)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stream := &failingStream{err: tc.err}
			session, o, err := protocol.Handshake(context.Background(), testkit.NewClock(time.Now()), "ipv4", stream, protocol.Config{ServerName: "example.org"})
			if err != nil || session != nil || !stream.closed || o.Status() != "failed" || o.Failure() != "inconclusive" {
				t.Fatalf("handshake: %v %v", o, err)
			}
			details, ok := o.TLS()
			if !ok || details.Alert != "" || details.CertificateValidation != "inconclusive" || details.TLSVersion != "" || details.ALPN != "" {
				t.Fatalf("fabricated peer/negotiated evidence: %+v", details)
			}
			raw, err := json.Marshal(o)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(raw), "CANARY") || strings.Contains(string(raw), `"alert"`) || strings.Contains(string(raw), "bad certificate") {
				t.Fatal("raw error retained")
			}
		})
	}
}
