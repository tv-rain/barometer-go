package dns

import (
	"context"
	"errors"
	"io"
	"slices"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/observation"
)

type Clock interface {
	Now() time.Time
	AfterFunc(time.Duration, func()) func() bool
}
type Conn interface {
	io.ReadWriteCloser
	SetDeadline(time.Time) error
}
type DatagramConn interface {
	Conn
	ReadDatagram([]byte) (int, bool, bool, error)
}
type Input struct {
	Clock   Clock
	Dial    func(context.Context, string, string) (Conn, error)
	Resolve func(context.Context, string, string) (bool, string, error)
}

const OperationBound = 15 * time.Second

var ErrInterrupted = errors.New("dns: operation interrupted")
var errTimeout = errors.New("dns: operation timeout")
var errWindow = errors.New("dns: observation window ended")

type Result struct {
	Outcome     string
	Observation observation.Observation
}

// Direct performs one UDP or length-prefixed TCP query, without retries or
// fallback. The caller records its terminal outcome before closing conn.
// The UDP transport compares every datagram source with the checked resolver.
// requirement-ref: REQ-LIMIT-0034
// requirement-ref: REQ-DNS-01-0033
// requirement-ref: REQ-DNS-01-0038
func Direct(ctx context.Context, in Input, q Question, transport, address, family string, collect bool, window time.Duration) (Conn, Result, error) {
	return exchange(ctx, in, q, transport, transport, address, family, collect, window)
}

// Session performs one length-prefixed DNS-over-TLS exchange on the plaintext
// side of an established, authenticated session, recorded with engine dot. The
// caller owns conn and records the terminal outcome before closing it; the
// operation bound or an outer interruption closes conn. There is no retry,
// cleartext fallback or second resolver.
// requirement-ref: REQ-LIMIT-0034
func Session(ctx context.Context, clock Clock, conn Conn, q Question, family string) (Result, error) {
	if conn == nil {
		return Result{}, ErrInterrupted
	}
	dial := func(context.Context, string, string) (Conn, error) { return conn, nil }
	_, result, err := exchange(ctx, Input{Clock: clock, Dial: dial}, q, "tcp", "dot", "", family, false, 0)
	return result, err
}

func exchange(ctx context.Context, in Input, q Question, transport, engine, address, family string, collect bool, window time.Duration) (conn Conn, result Result, err error) {
	query, e := Query(q)
	if e != nil || in.Clock == nil || in.Dial == nil || transport != "udp" && transport != "tcp" || collect && (transport != "udp" || window <= 0 || window > time.Second) {
		return nil, Result{}, ErrInterrupted
	}
	started := in.Clock.Now()
	op, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	stop := in.Clock.AfterFunc(OperationBound, func() { cancel(errTimeout) })
	defer stop()
	c, e := in.Dial(op, transport, address)
	if c != nil {
		conn = c
	}
	failure := ""
	responses := []observation.DNSResponse{}
	limited := false
	finish := func() (Conn, Result, error) {
		if ctx.Err() != nil {
			if conn != nil {
				_ = conn.Close()
			}
			return nil, Result{}, ErrInterrupted
		}
		if failure == "" && len(responses) == 0 {
			failure = "timeout"
		}
		if failure == "resolver_error" && errors.Is(context.Cause(op), errTimeout) {
			failure = "timeout"
		}
		outcome, status := failure, "failed"
		if len(responses) > 0 && failure == "" {
			status = "succeeded"
			outcome = "no_data"
			first := responses[0]
			switch first.RCode {
			case "NXDOMAIN":
				failure = "nxdomain"
			case "SERVFAIL":
				failure = "servfail"
			case "REFUSED":
				failure = "refused"
			case "NOERROR":
				for _, a := range first.Answers {
					if a.Type == q.Type {
						outcome = "resolved"
					}
				}
			default:
				failure = "inconclusive"
			}
			if failure != "" {
				status = "failed"
				outcome = failure
			}
			conflict := false
			for _, r := range responses[1:] {
				if r.RCode != first.RCode || !sameAnswers(r.Answers, first.Answers) {
					conflict = true
				}
			}
			if conflict {
				outcome = "conflicting_responses"
				status = "succeeded"
				failure = ""
			}
			if limited {
				outcome = "response_limit_reached"
				status = "succeeded"
				failure = ""
			}
		}
		duration := coarse(in.Clock.Now().Sub(started))
		o, e := observation.NewDNS(observation.Contract, status, duration, failure, observation.DNSDetails{Engine: engine, AddressFamily: family, QueryType: q.Type, Responses: responses, ResponseLimitReached: limited}, observation.ProtocolLimits{RetainDNSAnswers: true})
		if e != nil {
			if conn != nil {
				_ = conn.Close()
			}
			return nil, Result{}, ErrInterrupted
		}
		return conn, Result{outcome, o}, nil
	}
	if ctx.Err() != nil {
		return finish()
	}
	if e != nil || conn == nil {
		failure = errorClass(e)
		if errors.Is(context.Cause(op), errTimeout) {
			failure = "timeout"
		}
		return finish()
	}
	// Closing is the cancellation primitive, including blocked writes/reads. The
	// attempt scope independently closes on outer cancellation and budgets.
	unlink := context.AfterFunc(op, func() { _ = c.Close() })
	defer unlink()
	if transport == "tcp" {
		query = append([]byte{byte(len(query) >> 8), byte(len(query))}, query...)
	}
	if n, e := conn.Write(query); e != nil || n != len(query) {
		failure = errorClass(e)
		return finish()
	}
	var stopWindow func() bool
	defer func() {
		if stopWindow != nil {
			stopWindow()
		}
	}()
	lastMismatch := ""
	for {
		var b []byte
		var matchingSource = true
		var e error
		if transport == "tcp" {
			prefix := make([]byte, 2)
			_, e = io.ReadFull(conn, prefix)
			if e == nil {
				n := int(prefix[0])<<8 | int(prefix[1])
				if n < 12 {
					failure = "malformed_response"
					return finish()
				}
				b = make([]byte, n)
				_, e = io.ReadFull(conn, b)
			}
		} else {
			b = make([]byte, MaxUDP+1)
			var n int
			var truncated bool
			if d, ok := conn.(DatagramConn); ok {
				n, truncated, matchingSource, e = d.ReadDatagram(b)
			} else {
				n, e = conn.Read(b)
			}
			if e != nil {
				n = 0
			}
			if n < 0 || n > len(b) {
				failure = "malformed_response"
				return finish()
			}
			if truncated || n > MaxUDP {
				if !matchingSource || len(responses) > 0 {
					continue
				}
				failure = "malformed_response"
				return finish()
			}
			b = b[:n]
		}
		if ctx.Err() != nil {
			return finish()
		}
		if e != nil {
			if errors.Is(context.Cause(op), errWindow) || errors.Is(context.Cause(op), errTimeout) {
				if len(responses) == 0 {
					failure = "timeout"
					if lastMismatch != "" {
						failure = lastMismatch
					}
				}
				return finish()
			}
			failure = errorClass(e)
			if len(responses) > 0 {
				failure = ""
			} else if lastMismatch != "" {
				failure = lastMismatch
			}
			return finish()
		}
		maxBytes := MaxUDP
		if transport == "tcp" {
			maxBytes = MaxTCP
		}
		response, e := Parse(b, q, maxBytes)
		if !matchingSource {
			if e == nil {
				lastMismatch = "unexpected_source"
			}
			continue
		}
		if errors.Is(e, ErrMismatch) {
			lastMismatch = "query_mismatch"
			if transport == "udp" {
				continue
			}
			failure = lastMismatch
			return finish()
		}
		if e != nil {
			if transport == "udp" && len(responses) > 0 {
				continue
			}
			failure = "malformed_response"
			return finish()
		}
		response.ArrivalIndex = len(responses) + 1
		response.ArrivalOffsetMS = coarse(in.Clock.Now().Sub(started))
		responses = append(responses, response)
		if !collect || transport == "tcp" {
			return finish()
		}
		if len(responses) == 3 {
			limited = true
			return finish()
		}
		if stopWindow == nil {
			stopWindow = in.Clock.AfterFunc(window, func() { cancel(errWindow) })
		}
	}
}
func coarse(d time.Duration) int { return max(0, int(d/time.Millisecond)) / 250 * 250 }
func sameAnswers(a, b []observation.DNSAnswer) bool {
	if len(a) != len(b) {
		return false
	}
	for _, v := range a {
		if !slices.Contains(b, v) {
			return false
		}
	}
	return true
}
func errorClass(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	var timeout interface{ Timeout() bool }
	if errors.As(err, &timeout) && timeout.Timeout() {
		return "timeout"
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return "malformed_response"
	}
	return "resolver_error"
}

// System uses only a single checked resolver API call. Hidden wire messages,
// codes, datagrams and per-response timing are never synthesized.
// requirement-ref: REQ-DNS-01-0007
// requirement-ref: REQ-DNS-01-0035
func System(ctx context.Context, in Input, name, query string) (Result, error) {
	if in.Resolve == nil || in.Clock == nil {
		return Result{}, ErrInterrupted
	}
	started := in.Clock.Now()
	op, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	stop := in.Clock.AfterFunc(OperationBound, func() { cancel(errTimeout) })
	defer stop()
	found, failure, e := in.Resolve(op, name, query)
	if ctx.Err() != nil {
		return Result{}, ErrInterrupted
	}
	if errors.Is(context.Cause(op), errTimeout) {
		failure = "timeout"
	} else if e != nil {
		return Result{}, ErrInterrupted
	}
	outcome, status := failure, "failed"
	if failure == "" {
		status = "succeeded"
		outcome = "no_data"
		if found {
			outcome = "resolved"
		}
	}
	o, e := observation.NewDNS(observation.Contract, status, coarse(in.Clock.Now().Sub(started)), failure, observation.DNSDetails{Engine: "system", QueryType: query, Responses: []observation.DNSResponse{}}, observation.ProtocolLimits{})
	if e != nil {
		return Result{}, ErrInterrupted
	}
	return Result{outcome, o}, nil
}
