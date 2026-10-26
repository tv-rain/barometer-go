// Package tcp owns one bounded connect operation and socket-error normalization.
// It never selects a path, retries, persists data, or constructs result envelopes.
package tcp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"syscall"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/observation"
)

// OperationBound is the registered protocol-operation maximum (REQ-LIMIT-0008).
const OperationBound = 15 * time.Second

var ErrInterrupted = errors.New("tcp: operation interrupted")
var errTimeout = errors.New("tcp: operation deadline")

type Clock interface {
	Now() time.Time
	AfterFunc(time.Duration, func()) func() bool
}

// Connect returns a terminal observation and, on success, its open connection.
// dial must use the attempt's metered, destination-checked path. The caller
// owns immediate close (TCP-01) or the next bounded protocol stage. No data
// is read or written. An outer interruption has no protocol observation.
// requirement-ref: REQ-TCP-01-0006
// requirement-ref: REQ-TCP-01-0008
// requirement-ref: REQ-TCP-01-0009
// requirement-ref: REQ-TCP-01-0022
// requirement-ref: REQ-TCP-01-0025
// requirement-ref: REQ-TCP-01-0027
// requirement-ref: REQ-LIMIT-0008
func Connect(ctx context.Context, clock Clock, family string, dial func(context.Context) (io.ReadWriteCloser, error)) (io.ReadWriteCloser, observation.Observation, error) {
	if ctx.Err() != nil || clock == nil || dial == nil || family != "ipv4" && family != "ipv6" {
		return nil, observation.Observation{}, ErrInterrupted
	}
	started := clock.Now()
	op, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	stop := clock.AfterFunc(OperationBound, func() { cancel(errTimeout) })
	defer stop()
	conn, err := dial(op)
	// A terminal operation is recorded before cleanup. An outer condition
	// observed before that point wins; cancellation during caller close does not.
	if ctx.Err() != nil {
		if conn != nil {
			_ = conn.Close()
		}
		return nil, observation.Observation{}, ErrInterrupted
	}
	failure := ""
	if errors.Is(context.Cause(op), errTimeout) {
		failure = "timeout"
	} else if err != nil {
		failure = normalize(err)
	} else if conn == nil {
		return nil, observation.Observation{}, ErrInterrupted
	}
	if failure != "" && conn != nil {
		_ = conn.Close()
		conn = nil
	}
	status := "succeeded"
	if failure != "" {
		status = "failed"
	}
	duration := max(0, int(clock.Now().Sub(started)/time.Millisecond))
	duration = duration / 250 * 250
	// Fixed vocabulary and typed producer; the existing strict observation
	// decoder validates the generated wire form before publishing a value.
	data, e := json.Marshal(struct {
		Type     string `json:"type"`
		Status   string `json:"status"`
		Duration int    `json:"duration_ms"`
		Failure  string `json:"failure,omitempty"`
		Details  struct {
			Family string `json:"address_family"`
		} `json:"details"`
	}{"tcp_connect", status, duration, failure, struct {
		Family string `json:"address_family"`
	}{family}})
	if e != nil {
		if conn != nil {
			_ = conn.Close()
		}
		return nil, observation.Observation{}, ErrInterrupted
	}
	o, e := observation.DecodeProtocol(observation.Contract, data, observation.ProtocolLimits{})
	if e != nil {
		if conn != nil {
			_ = conn.Close()
		}
		return nil, observation.Observation{}, ErrInterrupted
	}
	return conn, o, nil
}

// normalize uses typed errors only. Unknown platform mappings remain
// inconclusive; raw error text/numbers never escape into observations.
func normalize(err error) string {
	switch {
	case errors.Is(err, syscall.ECONNREFUSED):
		return "connection_refused"
	case errors.Is(err, syscall.ECONNRESET):
		return "connection_reset"
	case errors.Is(err, syscall.ENETUNREACH):
		return "network_unreachable"
	case errors.Is(err, syscall.EHOSTUNREACH):
		return "address_unreachable"
	case errors.Is(err, syscall.EACCES), errors.Is(err, syscall.EPERM):
		return "permission_denied"
	case errors.Is(err, syscall.ETIMEDOUT), errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	}
	var timeout interface {
		error
		Timeout() bool
	}
	if errors.As(err, &timeout) && timeout.Timeout() {
		return "timeout"
	}
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return "socket_error"
	}
	return "inconclusive"
}
