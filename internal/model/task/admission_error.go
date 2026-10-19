package task

import "errors"

// ErrorContract identifies the public vocabulary, not a new wire envelope.
const ErrorContract = "0.1.0"

var (
	ErrAdmissionCode     = errors.New("task: invalid admission error code")
	ErrAdmissionEncoding = errors.New("task: admission errors have no JSON representation")
)

// AdmissionCode is separate from ExecutionReason and protocol observations.
// These are the ten initial classes in REQ-TRUST-0034 and task_unsupported
// (REQ-TRUST-0036).
// requirement-ref: REQ-TRUST-0034
// requirement-ref: REQ-TRUST-0036
type AdmissionCode uint8

const (
	SourceUnauthenticated AdmissionCode = iota + 1
	IssuerNotAuthorized
	PolicyUnavailable
	TrustUnavailable
	Expired
	IDConflict
	PolicyRollback
	TargetNotRegistered
	LimitExceeded
	SchemaInvalid
	Unsupported
)

func (c AdmissionCode) code() string {
	switch c {
	case SourceUnauthenticated:
		return "task_source_unauthenticated"
	case IssuerNotAuthorized:
		return "task_issuer_not_authorized"
	case PolicyUnavailable:
		return "task_policy_unavailable"
	case TrustUnavailable:
		return "task_trust_unavailable"
	case Expired:
		return "task_expired"
	case IDConflict:
		return "task_id_conflict"
	case PolicyRollback:
		return "policy_rollback"
	case TargetNotRegistered:
		return "target_not_registered"
	case LimitExceeded:
		return "task_limit_exceeded"
	case SchemaInvalid:
		return "task_schema_invalid"
	case Unsupported:
		return "task_unsupported"
	}
	return ""
}

// AdmissionError cannot carry input bytes, arbitrary messages, or wrapped errors
// (REQ-TRUST-0033..0035). Constructing one does not perform admission.
// requirement-ref: REQ-TRUST-0035
type AdmissionError struct{ code AdmissionCode }

func NewAdmissionError(code AdmissionCode) (AdmissionError, error) {
	if code.code() == "" {
		return AdmissionError{}, ErrAdmissionCode
	}
	return AdmissionError{code: code}, nil
}

func (e AdmissionError) Code() AdmissionCode { return e.code }
func (e AdmissionError) Error() string {
	if code := e.code.code(); code != "" {
		return code
	}
	return "task: invalid admission error"
}

func (AdmissionError) MarshalJSON() ([]byte, error) { return nil, ErrAdmissionEncoding }
func (*AdmissionError) UnmarshalJSON([]byte) error  { return ErrAdmissionEncoding }
