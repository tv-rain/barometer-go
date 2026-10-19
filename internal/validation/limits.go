package validation

import "errors"

// Refining classes. A returned error that matches one of them also matches
// the coarse ErrTask or ErrTiming it refines, so existing callers keep their
// class. Admission maps them to public classes; nothing here is an admission
// decision.
var (
	// ErrLimit marks a value above a registered maximum (REQ-LIMIT-0002).
	// A value below a registered minimum is not a limit failure.
	ErrLimit = errors.New("validation: limit")
	// ErrExpired marks a not_after at or before trusted acceptance time.
	ErrExpired = errors.New("validation: expired")
	// ErrUnsupported marks a valid value this build does not provide, such as
	// an unrecognized risk factor type rejected under REQ-SAFETY-0017.
	ErrUnsupported = errors.New("validation: unsupported")
)

// classed carries only fixed classes, never input values.
type classed struct{ class, base error }

func (e classed) Error() string        { return e.base.Error() }
func (e classed) Is(target error) bool { return target == e.class || target == e.base }

var (
	errTaskLimit   = classed{ErrLimit, ErrTask}
	errTimingLimit = classed{ErrLimit, ErrTiming}
	errExpired     = classed{ErrExpired, ErrTiming}
	errUnsupported = classed{ErrUnsupported, ErrTask}
)

// MaxTaskIDBytes is the registered task_id maximum (REQ-LIMIT-0004).
const MaxTaskIDBytes = 128

// TaskID checks the task_id rules on the exact string (REQ-TRUST-0019). More
// than 128 bytes is a limit failure whatever the characters; empty or any
// other character or a leading '.', '_' or '-' is invalid. Campaign scope and
// non-user-specificity cannot be established from syntax and remain an issuer
// review obligation.
// requirement-ref: REQ-TRUST-0019
// requirement-ref: REQ-LIMIT-0004
func TaskID(id string) error {
	if len(id) > MaxTaskIDBytes {
		return errTaskLimit
	}
	if !taskID.MatchString(id) {
		return ErrTask
	}
	return nil
}

// limit returns errTaskLimit when value exceeds max and ErrTask when it is
// below min. Callers reject; nothing is clamped.
// requirement-ref: REQ-LIMIT-0002
func limit(value, min, max int) error {
	if value > max {
		return errTaskLimit
	}
	if value < min {
		return ErrTask
	}
	return nil
}
