// Package tls owns one bounded authenticated TLS handshake over an
// established stream and its failure normalization. It never dials, selects a
// path, resolves a name, retries, resumes a session, persists data or
// constructs result envelopes.
package tls

import (
	"context"
	stdtls "crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"reflect"
	"slices"
	"syscall"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/observation"
)

// OperationBound is the registered protocol-operation maximum (REQ-LIMIT-0008).
const OperationBound = 15 * time.Second

var ErrInterrupted = errors.New("tls: operation interrupted")
var errTimeout = errors.New("tls: operation deadline")

type Clock interface {
	Now() time.Time
	AfterFunc(time.Duration, func()) func() bool
}

// Stream is the established connection the handshake runs on. Its bytes are
// TLS handshake and record bytes; the caller decides how they are metered.
type Stream interface {
	io.ReadWriteCloser
	SetDeadline(time.Time) error
}

// Config is the complete handshake profile. There is no cipher, version,
// client-certificate, resumption or verification control: certificate and
// hostname validation always run against Roots, and nil Roots selects the
// platform's system trust roots through crypto/x509.
type Config struct {
	ServerName string   // normalized DNS name or explicit IP, verified against the certificate
	ALPN       []string // registered offer, in order; may be empty
	Roots      *x509.CertPool
}

// Session is an established, authenticated TLS session. Its Read and Write
// are the plaintext side; Close sends the normal closure and closes the stream.
type Session struct{ conn *stdtls.Conn }

func (s *Session) Read(p []byte) (int, error)  { return s.conn.Read(p) }
func (s *Session) Write(p []byte) (int, error) { return s.conn.Write(p) }
func (s *Session) Close() error                { return s.conn.Close() }

// Handshake performs one TLS handshake on raw and returns its terminal
// tls_handshake observation and, on success, the session. Without a session
// raw is closed. No application data is read or written. An outer
// interruption, or a profile that is not a valid TLS-01 shape, has no
// protocol observation.
// requirement-ref: REQ-TLS-01-0005
// requirement-ref: REQ-TLS-01-0007
// requirement-ref: REQ-TLS-01-0008
// requirement-ref: REQ-TLS-01-0013
// requirement-ref: REQ-TLS-01-0017
// requirement-ref: REQ-TLS-01-0034
// requirement-ref: REQ-TLS-01-0036
// requirement-ref: REQ-LIMIT-0008
// requirement-ref: REQ-OBS-0019
// requirement-ref: REQ-OBS-0020
// requirement-ref: REQ-OBS-0043
// requirement-ref: REQ-OBS-0048
// requirement-ref: REQ-OBS-0049
// requirement-ref: REQ-OBS-0050
func Handshake(ctx context.Context, clock Clock, family string, raw Stream, cfg Config) (*Session, observation.Observation, error) {
	if raw == nil {
		return nil, observation.Observation{}, ErrInterrupted
	}
	if ctx.Err() != nil || clock == nil || family != "ipv4" && family != "ipv6" || cfg.ServerName == "" ||
		observation.ValidateALPNOffer(observation.Contract, cfg.ALPN) != nil {
		_ = raw.Close()
		return nil, observation.Observation{}, ErrInterrupted
	}
	verified := false
	conn := stdtls.Client(stream{raw}, &stdtls.Config{
		ServerName: cfg.ServerName, RootCAs: cfg.Roots, NextProtos: slices.Clone(cfg.ALPN),
		MinVersion: stdtls.VersionTLS12, SessionTicketsDisabled: true, Renegotiation: stdtls.RenegotiateNever,
		// Called only after the standard chain and hostname checks succeeded.
		VerifyConnection: func(stdtls.ConnectionState) error { verified = true; return nil },
	})
	started := clock.Now()
	op, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	stop := clock.AfterFunc(OperationBound, func() { cancel(errTimeout) })
	defer stop()
	err := conn.HandshakeContext(op)
	// An outer condition observed before the outcome wins.
	if ctx.Err() != nil {
		_ = conn.Close()
		return nil, observation.Observation{}, ErrInterrupted
	}
	d := observation.TLSDetails{AddressFamily: family}
	status, failure := "succeeded", ""
	if errors.Is(context.Cause(op), errTimeout) {
		failure = "timeout"
	} else if err != nil {
		failure, d.CertificateValidation, d.Alert = normalize(err, len(cfg.ALPN) > 0)
	}
	if failure != "" {
		status = "failed"
		if d.CertificateValidation == "" {
			d.CertificateValidation = "not_completed"
			if verified {
				d.CertificateValidation = "trusted"
			} else if failure == "inconclusive" {
				d.CertificateValidation = "inconclusive"
			}
		}
	} else {
		d.CertificateValidation = "trusted"
		state := conn.ConnectionState()
		d.TLSVersion, _ = observation.NegotiatedTLSVersion(state.Version, true)
		d.ALPN, _ = observation.NegotiatedALPN(state.NegotiatedProtocol, true)
	}
	duration := max(0, int(clock.Now().Sub(started)/time.Millisecond))
	o, e := observation.NewTLS(observation.Contract, status, duration/250*250, failure, d)
	if e != nil || failure != "" {
		_ = conn.Close()
		if e != nil {
			return nil, observation.Observation{}, ErrInterrupted
		}
		return nil, o, nil
	}
	return &Session{conn}, o, nil
}

// normalize uses typed errors only and returns the failure, a certificate
// validation class when local verification decided the failure, and a coarse
// peer-alert class. Untyped library errors stay inconclusive; raw error text,
// alert codes and certificate data never escape into observations.
func normalize(err error, offered bool) (failure, validation, alert string) {
	var verify *stdtls.CertificateVerificationError
	if errors.As(err, &verify) {
		var name x509.HostnameError
		var invalid x509.CertificateInvalidError
		switch {
		case errors.As(verify.Err, &name):
			return "certificate_name_mismatch", "name_mismatch", ""
		case errors.As(verify.Err, &invalid) && invalid.Reason == x509.Expired:
			return "certificate_time_invalid", "time_invalid", ""
		}
		return "certificate_untrusted", "untrusted", ""
	}
	if code, ok := peerAlert(err); ok {
		// no_application_protocol answers an ALPN offer with no overlap.
		if code == 120 && offered {
			return "alpn_no_overlap", "", ""
		}
		alert, _ = observation.PeerAlertClass(code, true)
		return "alert", "", alert
	}
	var record stdtls.RecordHeaderError
	switch {
	case errors.As(err, &record):
		return "protocol_error", "", ""
	case errors.Is(err, syscall.ECONNRESET):
		return "connection_reset", "", ""
	case errors.Is(err, syscall.ETIMEDOUT):
		return "timeout", "", ""
	}
	var timeout interface {
		error
		Timeout() bool
	}
	if errors.As(err, &timeout) && timeout.Timeout() {
		return "timeout", "", ""
	}
	return "inconclusive", "", ""
}

// peerAlert recognizes only crypto/tls's typed record of a received alert: a
// net.OpError "remote error" wrapping the package's alert description byte.
// A locally sent alert or a closure without one is not a peer alert.
func peerAlert(err error) (uint8, bool) {
	var op *net.OpError
	if !errors.As(err, &op) || op.Op != "remote error" || op.Err == nil {
		return 0, false
	}
	v := reflect.ValueOf(op.Err)
	if v.Kind() != reflect.Uint8 || v.Type().PkgPath() != "crypto/tls" || v.Type().Name() != "alert" {
		return 0, false
	}
	return uint8(v.Uint()), true
}

// stream presents raw as a net.Conn without exposing any address. Every
// deadline applies to both directions.
type stream struct{ Stream }

func (stream) LocalAddr() net.Addr                  { return noAddr{} }
func (stream) RemoteAddr() net.Addr                 { return noAddr{} }
func (s stream) SetReadDeadline(t time.Time) error  { return s.SetDeadline(t) }
func (s stream) SetWriteDeadline(t time.Time) error { return s.SetDeadline(t) }

type noAddr struct{}

func (noAddr) Network() string { return "tcp" }
func (noAddr) String() string  { return "" }
