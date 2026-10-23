package barometer

// code is one API-local operational class from docs/public-api.md "Errors and
// Bounds". Public admission classes are returned as task.AdmissionError by the
// admission entry point and are not duplicated here.
type code uint8

const (
	notInitialized code = iota + 1
	alreadyInitialized
	clientClosed
	transitionInProgress
	handleUnavailable
	bundleUnavailable
	runNotDue
	capacityExceeded
	cleanupIncomplete
	clearingIncomplete
	// storageUnavailable is a pending-storage failure of NextBundle or
	// AcknowledgeBundle (DEV-563). A failed deletion is not acknowledgement.
	storageUnavailable
	// policyRejected is returned only by the unexported policy-replacement
	// seam: the replacement was refused and the previous policy is active.
	policyRejected
	// proxyRejected leaves the previous proxy configuration and work unchanged.
	proxyRejected
)

func (c code) String() string {
	switch c {
	case notInitialized:
		return "not_initialized"
	case alreadyInitialized:
		return "already_initialized"
	case clientClosed:
		return "client_closed"
	case transitionInProgress:
		return "transition_in_progress"
	case handleUnavailable:
		return "handle_unavailable"
	case bundleUnavailable:
		return "bundle_unavailable"
	case runNotDue:
		return "run_not_due"
	case capacityExceeded:
		return "capacity_exceeded"
	case cleanupIncomplete:
		return "cleanup_incomplete"
	case clearingIncomplete:
		return "clearing_incomplete"
	case storageUnavailable:
		return "storage_unavailable"
	case policyRejected:
		return "policy_rejected"
	case proxyRejected:
		return "proxy_rejected"
	}
	return "operation_failed"
}

// opError carries only a fixed code: no input bytes, targets, source
// references, handles, adapter errors or wrapped causes. It has no Unwrap.
type opError struct{ code code }

func (e opError) Error() string { return e.code.String() }

func fail(c code) error { return opError{code: c} }
