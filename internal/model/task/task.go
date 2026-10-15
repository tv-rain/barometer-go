// Package task owns candidate task values. Candidates are not admitted tasks.
// It is a dependency leaf: no decoding, I/O, platform calls or orchestration.
package task

import (
	"fmt"
	"strings"
	"time"
)

// Parameters is implemented by the selected test's typed, normalized values.
// The common envelope neither interprets these values nor switches test schemas.
type Parameters interface{ TaskParameters() }

type Candidate struct {
	SchemaVersion, ID, TestID, TestVersion string
	Parameters                             Parameters
	Schedule                               Schedule
	Execution                              Execution
	Safety                                 Safety
	Labels                                 []Label
}

type Label struct{ Key, Value string }

// Optional pointers preserve presence until validation applies documented defaults.
type Schedule struct {
	StartsAt, NotAfter *Timestamp
	Window             *TimeWindow
	Repeat             Repeat
	Retry              Retry
	MaxRuns            *int
}

// Timestamp retains exact decimal fractions without time.Time's nanosecond
// truncation. Whole is the UTC whole second; Fraction has no trailing zeros.
type Timestamp struct {
	Whole    time.Time
	Fraction string
}

func TimestampFromTime(t time.Time) Timestamp {
	return Timestamp{Whole: t.UTC().Truncate(time.Second), Fraction: strings.TrimRight(fmt.Sprintf("%09d", t.Nanosecond()), "0")}
}
func (t Timestamp) Before(other Timestamp) bool {
	if !t.Whole.Equal(other.Whole) {
		return t.Whole.Before(other.Whole)
	}
	return t.Fraction < other.Fraction
}
func (t Timestamp) BeforeTime(other time.Time) bool { return t.Before(TimestampFromTime(other)) }
func (t Timestamp) AfterTime(other time.Time) bool  { return TimestampFromTime(other).Before(t) }

type TimeWindow struct {
	Timezone string
	Ranges   []TimeRange
}

// End is exclusive; End < Start crosses midnight. Equal endpoints cover no time.
type TimeRange struct{ Start, End int } // minutes since local midnight
type Repeat struct {
	Mode                           string
	IntervalSeconds, JitterSeconds *int
}
type Retry struct {
	MaxAttempts int
	On          []ExecutionReason
	Backoff     *Backoff
}
type Backoff struct {
	Mode                                 string
	InitialDelayMS, MaxDelayMS, JitterMS int
	Multiplier                           *int
}

type Execution struct {
	NetworkPath     string
	Limits          AttemptLimits
	Network         NetworkRequirements
	Device          DeviceRequirements
	ConcurrencyMode string // absent remains absent: the contract specifies no default
}
type AttemptLimits struct{ DurationMS, BytesSent, BytesReceived, Connections int }
type NetworkRequirements struct {
	AllowedTransports []string
	VPNPolicy         string
	RequireValidated  bool
	Context           ContextRequirements
}
type ContextRequirements struct{ RequiredFields, AllowedCountryCodes []string }
type DeviceRequirements struct {
	MinimumBatteryPercent *int
	ChargingOnly          bool
}
type Safety struct {
	RiskLevel       string
	Factors         []RiskFactor
	RequiresConsent bool
}
type RiskFactor struct{ Type, Level string }

// ExecutionReason is deliberately distinct from a measurement outcome.
type ExecutionReason string

const (
	PlatformSuspended         ExecutionReason = "platform_suspended"
	NetworkContextUnavailable ExecutionReason = "network_context_unavailable"
	ProxyUnavailable          ExecutionReason = "proxy_unavailable"
	InternalError             ExecutionReason = "internal_error"
)

// Retryable never infers an execution failure from DNS/TCP/TLS/HTTP observations.
func Retryable(reason ExecutionReason) bool {
	switch reason {
	case PlatformSuspended, NetworkContextUnavailable, ProxyUnavailable, InternalError:
		return true
	}
	return false
}
