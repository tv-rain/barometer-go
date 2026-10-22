// Package measurement owns the minimal contract between execution and a
// registered measurement runner. It holds no runner, registry or global state.
// A runner returns protocol observations and budget usage only: it never
// constructs result envelopes, chooses retries, changes paths, persists data or
// classifies censorship. Presence of a runner is not evidence of conformance.
package measurement

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/observation"
	"github.com/tv-rain/barometer-go/internal/model/task"
)

// ErrRunner is the single fixed error a runner reports for an execution
// failure. Protocol failures are observations, never runner errors.
var ErrRunner = errors.New("measurement: runner failed")

// Conn is the connection a runner receives. It deliberately exposes no network
// address, socket or platform handle, and keeps the contract free of the net
// package so admission-only code that reads the registry links no network code.
type Conn interface {
	io.ReadWriteCloser
	SetDeadline(time.Time) error
}

// Dialer is the attempt-scoped connection seam. Execution supplies a fresh
// dialer for every attempt; the selected path, proxy handling and byte or
// connection accounting belong to the dialer's owner, not to the runner. A
// runner opens connections only through it.
type Dialer interface {
	DialContext(ctx context.Context, network, address string) (Conn, error)
}

// Input carries the admitted test's normalized parameters and its dialer.
// Parameters were produced by the registered codec; a runner must reject a
// value of an unexpected type with ErrRunner rather than guess.
type Input struct {
	Parameters task.Parameters
	Dialer     Dialer
}

// Usage is the runner's account of application bytes and connections. It is
// reported for budget owners; it does not itself enforce a limit.
type Usage struct {
	BytesSent, BytesReceived, Connections int
}

// Output is a terminal measurement outcome with its bounded observations. A
// negative DNS, TCP, TLS or HTTP outcome is still a terminal outcome
// (REQ-SCHEDULE-0025). Result construction validates it against the test.
type Output struct {
	Outcome      string
	Observations []observation.Observation
	Usage        Usage
}

// Runner executes exactly one attempt of one registered test version. It must
// honor ctx, open connections only through Input.Dialer and return ErrRunner,
// not a raw platform error, when it cannot produce a terminal outcome.
// Implementations must be safe for concurrent use by independent attempts.
type Runner interface {
	Run(ctx context.Context, in Input) (Output, error)
}
