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

// ErrTargetRejected is a local safety refusal, never a transport observation.
var ErrTargetRejected = errors.New("measurement: target rejected")

// ErrResolutionFailed lets a runner distinguish the resolution stage without
// receiving a hostname, resolver detail or raw platform error.
var ErrResolutionFailed = errors.New("measurement: resolution failed")

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

// TargetDialer keeps redirect chains bound to the preceding registered hop.
type TargetDialer interface {
	Dialer
	Redirect(task.TargetTuple) (TargetDialer, error)
}

// Clock is the attempt's monotonic time source for bounded protocol operations.
type Clock interface {
	Now() time.Time
	AfterFunc(time.Duration, func()) func() bool
}

// Input carries the admitted test's normalized parameters, its dialer and the
// test-start marker. Parameters were produced by the registered codec; a
// runner must reject a value of an unexpected type with ErrRunner rather than
// guess.
//
// The test starts at its first test-specific network operation or, when that
// comes first, its own network-state observation (REQ-RESULT-0051). A dial
// through Dialer marks the start at resolution or permitted contact. A local
// safety/budget refusal does not start it. A runner whose procedure observes
// network state before dialing calls Start first. Start is idempotent and safe
// for concurrent use; validation and other local work before it never start
// the test.
type Input struct {
	Parameters task.Parameters
	Clock      Clock
	// CheckNetwork starts the test's own stable-generation/family check.
	// A refusal ends the attempt with the execution owner's reason.
	CheckNetwork func(family string) error
	// RecordOutcome fixes a terminal protocol outcome before cleanup. False
	// means an outer condition already ended the attempt; publish no outcome.
	// After a successful marker, perform only cleanup, with no further I/O.
	RecordOutcome func() bool
	Dialer        Dialer
	Start         func()
	// Redirect validates a complete normalized redirect tuple and returns a
	// dialer restricted to that hop, sharing this attempt's budgets and path.
	// A rejection is handled as the registered test's redirect_rejected outcome.
	Redirect func(task.TargetTuple) (TargetDialer, error)
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
// not a raw platform error, when it cannot produce a terminal outcome. A
// returned outcome is terminal: execution records it as completed even when
// ctx ends while connections close (REQ-RESULT-0052), so a runner interrupted
// before its outcome must return ErrRunner instead. Implementations must be
// safe for concurrent use by independent attempts.
type Runner interface {
	Run(ctx context.Context, in Input) (Output, error)
}
