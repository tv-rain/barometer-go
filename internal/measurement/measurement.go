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

// ErrNoNetworkState and ErrNetworkState are the two fixed failures of a
// network-state read: no supported platform API, or a failure of an available
// API before a coherent snapshot. Neither carries platform detail.
var (
	ErrNoNetworkState = errors.New("measurement: network state unavailable")
	ErrNetworkState   = errors.New("measurement: network state failed")
)

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

// SessionConn is an attempt connection that can carry a session protocol,
// TLS, whose handshake and record bytes are not test traffic
// (REQ-LIMIT-0034). Session switches it once: raw is the same stream without
// metering, still owned and closed by the attempt, and plaintext wraps the
// session's application side so those bytes count against the attempt
// budget. Afterwards the connection itself refuses reads and writes; a
// second call fails.
type SessionConn interface {
	Conn
	Session() (raw Conn, plaintext func(io.ReadWriter) io.ReadWriter, err error)
}

// DatagramConn preserves UDP truncation through the attempt meter. The
// transport compares the source to the checked peer without exposing addresses.
type DatagramConn interface {
	Conn
	ReadDatagram([]byte) (n int, truncated, matchingSource bool, err error)
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

// StagedDialer exposes minimized DNS/TCP observations from the execution-owned
// gate. It resolves once and pins contact while preserving scope accounting.
type StagedDialer interface {
	Dialer
	ConnectStages(context.Context, string, string, Clock) (Conn, []observation.Observation, string, error)
}

// Clock is the attempt's monotonic time source for bounded protocol operations.
type Clock interface {
	Now() time.Time
	AfterFunc(time.Duration, func()) func() bool
}

// CompositeInput carries trusted attempt capabilities for an admitted profile.
// Step narrows target authority without replacing path, context or budgets.
type CompositeInput struct {
	Validation        string
	RequireValidation bool
	Step              func(task.TestVersion, []task.TargetTuple) (Input, error)
}

type CompositeOutput struct {
	ID            string
	Version       int64
	Validation    string
	Steps         []StepOutput
	StoppedBefore string
}
type StepOutput struct {
	ID     string
	Test   task.TestVersion
	Output *Output // nil only for the single final interrupted step
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
	Composite  *CompositeInput
	Parameters task.Parameters
	// Resolve performs one authorized system lookup, returning only a coarse
	// result. It never exposes platform error text, resolver state or addresses.
	Resolve func(context.Context, string, string) (bool, string, error)
	Clock   Clock
	// CheckNetwork starts the test's own stable-generation/family check.
	// A refusal ends the attempt with the execution owner's reason.
	CheckNetwork func(family string) error
	// RecordOutcome fixes a terminal protocol outcome before cleanup. False
	// means an outer condition already ended the attempt; publish no outcome.
	// After a successful marker, perform only cleanup, with no further I/O.
	RecordOutcome func() bool
	Dialer        Dialer
	// Network reads the access state once through the shared Network
	// Context collector, for a test whose procedure is that read (NET-01).
	// It starts the test and arms the attempt's generation watch; it sends
	// no traffic and acquires no egress context. It is nil for every other
	// test, which then also receives no access values.
	Network func(context.Context) (observation.AccessSnapshot, error)
	// Validation reads NET-03's platform validation and captive-portal state
	// once, after the stability check and before any contact. It returns only
	// the coarse platform_validation value, unknown when the platform cannot
	// report it. A changed generation or an ended attempt is an error, and the
	// execution owner records the reason. It is nil for every other test.
	Validation func(context.Context) (string, error)
	Start      func()
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

// Output carries a terminal primitive outcome or a bounded composite profile.
// Only a composite may retain incomplete steps after an outer interruption. A
// negative DNS, TCP, TLS or HTTP outcome is still a terminal outcome
// (REQ-SCHEDULE-0025). Result construction validates it against the test.
type Output struct {
	Composite *CompositeOutput
	// Access is NET-01's normalized access object; nil for every other test.
	Access *observation.AccessSnapshot
	// PlatformValidation and ResponseClass are NET-03's separate platform
	// and active evidence; both are empty for every other test.
	PlatformValidation, ResponseClass string
	Outcome                           string
	Observations                      []observation.Observation
	Usage                             Usage
}

// Runner executes exactly one attempt of one registered test version. It must
// honor ctx, open connections only through Input.Dialer and return ErrRunner,
// not a raw platform error, when it cannot produce a terminal outcome. A
// returned primitive outcome is terminal: execution records it as completed even when
// ctx ends while connections close (REQ-RESULT-0052), so a runner interrupted
// before its outcome must return ErrRunner instead. Implementations must be
// safe for concurrent use by independent attempts. NET-05 may instead return
// Output.Composite with outcome incomplete; execution supplies the common
// stopping reason and validates the retained partial profile.
type Runner interface {
	Run(ctx context.Context, in Input) (Output, error)
}
