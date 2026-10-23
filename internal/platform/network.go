package platform

import "context"

// NetworkType is the coarse base transport of the device network. The zero
// value is NetworkUnknown, so an unset type never satisfies a requirement.
type NetworkType uint8

const (
	NetworkUnknown NetworkType = iota
	NetworkWiFi
	NetworkCellular
	NetworkEthernet
	NetworkOther
)

// VPNState is whether the platform reports a VPN on the application's path.
// The zero value is VPNUnknown, which is never treated as VPNNotDetected.
type VPNState uint8

const (
	VPNUnknown VPNState = iota
	VPNDetected
	VPNNotDetected
)

// Availability is a coarse platform-reported yes/no. The zero value is
// AvailabilityUnknown.
type Availability uint8

const (
	AvailabilityUnknown Availability = iota
	AvailabilityYes
	AvailabilityNo
)

// AccessState is one coarse access-context reading (REQ-CONTEXT-0002). It has
// no field for an interface name, address, DNS server, Wi-Fi or cell
// identifier, SIM, subscriber, carrier or location (REQ-CONTEXT-0003), so none
// can reach the core through it. Generation identifies the access network the
// values belong to; it changes whenever that network changes and is local
// metadata that never leaves the core. Zero means no generation is known.
// Validated is the platform's own Internet-validation state; the core never
// substitutes a connectivity check of its own.
type AccessState struct {
	Generation uint64
	Type       NetworkType
	VPN        VPNState
	IPv4       Availability
	IPv6       Availability
	Validated  Availability
}

// Network reports coarse access context for execution constraints and
// results. The query carries no task value, so no task selects or changes the
// answer. An error, a zero generation or an unrecognized value never satisfies
// a requirement; values the platform cannot determine are reported as unknown
// (REQ-CONTEXT-0034).
//
// Watch arranges for changed to be called once, from any goroutine, when the
// access-network generation is no longer generation, immediately if it already
// differs. stop cancels the watch, is idempotent and does not wait for a call
// in progress. Implementors honor finite deadlines and do no background work
// beyond delivering a change. The shape is interim: the native capability, its
// change signal and its binding are a review point of DEV-455.
type Network interface {
	Access(context.Context) (AccessState, error)
	Watch(generation uint64, changed func()) (stop func())
}
