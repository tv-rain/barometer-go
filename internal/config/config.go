// Package config owns immutable host configuration snapshots and their decoding.
// It performs no I/O, admission, scheduling, secret creation or activation.
package config

import "errors"

var (
	ErrValue       = errors.New("config: value")
	ErrApproval    = errors.New("config: approval")
	ErrVersion     = errors.New("config: version")
	ErrUnavailable = errors.New("config: unavailable")
)

// Contract is selected by the trusted integration, never by configuration JSON.
const Contract = "0.1.0"

// Proxy, NetworkContext and Linkage are copied values, not mutable shared state.
type Proxy struct {
	Host           string
	Port           int
	ConnectionCode string
}
type NetworkContext struct{ PrimaryEndpoint, FallbackEndpoint string }
type Linkage struct {
	Mode, Scope           string
	RotationPeriodSeconds int
}

// Snapshot has no exported fields or mutable container aliases. Its zero value
// is not a successfully decoded configuration.
type Snapshot struct {
	owner      *Decoder
	valid      bool
	proxy      Proxy
	hasProxy   bool
	network    NetworkContext
	hasNetwork bool
	linkage    Linkage
}

func (s Snapshot) Valid() bool                            { return s.valid }
func (s Snapshot) Proxy() (Proxy, bool)                   { return s.proxy, s.hasProxy }
func (s Snapshot) NetworkContext() (NetworkContext, bool) { return s.network, s.hasNetwork }
func (s Snapshot) Linkage() Linkage                       { return s.linkage }

// Approvals is trusted, deployment-reviewed input supplied by an internal adapter.
// Exact allowlists establish review, which syntax checks alone cannot establish.
// Entries must be non-user-specific. They confer no task-admission authority.
// NewDecoder copies these bounded lists; later caller mutations have no effect.
type Approvals struct {
	Endpoints, ConnectionCodes, LinkageScopes []string
	AllowLinkage                              bool
}

type Decoder struct {
	ready     bool
	approvals Approvals
}

// WithProxy returns a complete replacement snapshot only after validation. On
// failure it returns the unchanged snapshot and a fixed error. A future runtime
// must still quiesce old-generation work before publishing a successful change.
func (d *Decoder) WithProxy(current Snapshot, input []byte) (Snapshot, error) {
	if d == nil || !d.ready || !current.valid || current.owner != d {
		return current, ErrUnavailable
	}
	p, err := d.decodeProxy(input)
	if err != nil {
		return current, err
	}
	current.proxy, current.hasProxy = p, true
	return current, nil
}

// WithoutProxy is a value transformation; it does not cancel active work.
func (s Snapshot) WithoutProxy() Snapshot {
	s.proxy, s.hasProxy = Proxy{}, false
	return s
}
