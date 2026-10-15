// Package execution owns attempt lifecycle transitions. This pure value model
// does not admit tasks, schedule work, enforce budgets or execute measurements.
package execution

import (
	"errors"

	"github.com/tv-rain/barometer-go/internal/model/task"
)

const Contract = "0.1.0"

var (
	ErrTransition = errors.New("execution: invalid transition")
	ErrReason     = errors.New("execution: invalid reason")
	ErrEncoding   = errors.New("execution: attempts have no JSON representation")
)

// Phase is in-process lifecycle state, not a result wire enum. Zero is invalid.
type Phase uint8

const (
	Pending Phase = iota + 1
	Running
	Completed
	Incomplete
	NotRun
)

// Attempt is immutable: operations return a new value, or the unchanged value
// on error. A terminal attempt cannot be reopened or have its reason replaced.
// The owner must serialize competing events; this value is not a runtime lock.
type Attempt struct {
	phase  Phase
	reason task.ExecutionReason
}

func NewAttempt() Attempt                      { return Attempt{phase: Pending} }
func (a Attempt) Phase() Phase                 { return a.phase }
func (a Attempt) Reason() task.ExecutionReason { return a.reason }
func (a Attempt) Terminal() bool {
	return a.phase == Completed || a.phase == Incomplete || a.phase == NotRun
}

// Start records that execution began after the caller established admission and
// eligibility. It cannot itself establish either condition.
func (a Attempt) Start() (Attempt, error) {
	if a.phase != Pending {
		return a, ErrTransition
	}
	a.phase = Running
	return a, nil
}

// Complete records a runner's terminal measurement outcome, including negative
// DNS/TCP/TLS/HTTP observations (REQ-RESULT-0021). It accepts no error/reason.
// Result construction must separately require and validate the measurement.
// requirement-ref: REQ-RESULT-0021
func (a Attempt) Complete() (Attempt, error) {
	if a.phase != Running {
		return a, ErrTransition
	}
	a.phase = Completed
	return a, nil
}

// Stop records the first observed outer termination condition. Before Start it
// yields not_run; after Start it yields incomplete. No protocol error, raw
// platform error, admission error, or arbitrary message is accepted.
func (a Attempt) Stop(reason task.ExecutionReason) (Attempt, error) {
	if a.phase != Pending && a.phase != Running {
		return a, ErrTransition
	}
	if !reason.Valid() {
		return a, ErrReason
	}
	if a.phase == Pending {
		a.phase = NotRun
	} else {
		a.phase = Incomplete
	}
	a.reason = reason
	return a, nil
}

// RetryReason exposes only a terminal, registered retry candidate. The caller
// must still apply validation.RetryEligible for policy, timing and count bounds.
// requirement-ref: REQ-SCHEDULE-0027
func (a Attempt) RetryReason() (task.ExecutionReason, bool) {
	if (a.phase == Incomplete || a.phase == NotRun) && task.Retryable(a.reason) {
		return a.reason, true
	}
	return "", false
}

// Internal phases must not accidentally become a new unversioned wire schema.
// The future result owner maps terminal phases to the existing result contract.
func (Attempt) MarshalJSON() ([]byte, error) { return nil, ErrEncoding }
func (*Attempt) UnmarshalJSON([]byte) error  { return ErrEncoding }
