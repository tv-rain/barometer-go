// Package platform owns narrow internal capability contracts consumed by the
// portable core. Concrete reviewed adapters are not implemented here.
package platform

import (
	"context"
	"time"
)

// Clock is installed internally by a reviewed integration. Unavailable or
// invalid time is an error, never an implicit system-clock fallback. Implementors
// honor cancellation and a finite caller deadline without background work.
type Clock interface {
	Now(context.Context) (time.Time, error)
}

// LinkageKeyStore never exports the installation secret. Derivation lazily
// creates protected, application-scoped, non-synchronizing key material; Delete
// is idempotent. Implementors honor finite deadlines and fail when protection
// is unavailable. Ordinary-file or host-provided key fallback is forbidden.
type LinkageKeyStore interface {
	DeriveHMACSHA256(context.Context, []byte) ([32]byte, error)
	Delete(context.Context) error
}

// ConsentStatus is the host-reported consent state for measurements that
// require consent. The zero value is ConsentUnavailable, so an unset status
// never permits an attempt.
type ConsentStatus uint8

const (
	ConsentUnavailable ConsentStatus = iota
	ConsentValid                     // valid consent covers the measurement
	ConsentAbsent
	ConsentEnded // expired or withdrawn
)

// Consent reports consent status outside task-controlled data. The query
// carries no task value, so no task selects or changes the answer, and the
// answer carries no user, account, device or consent-record identifier. Only
// ConsentValid with a nil error permits an attempt. The shape is interim: the
// native capability and its binding are a review point of DEV-455.
type Consent interface {
	Status(context.Context) (ConsentStatus, error)
}

// DeviceState is a coarse device snapshot. A false Known flag means the value
// could not be determined; it never satisfies a requirement.
type DeviceState struct {
	BatteryPercent int
	BatteryKnown   bool
	Charging       bool
	ChargingKnown  bool
}

// Device reports coarse device state for execution constraints. An error means
// the state is unavailable, which never satisfies a requirement.
type Device interface {
	State(context.Context) (DeviceState, error)
}

// TimeZone reports the current device time zone used for local time windows
// (REQ-SCHEDULE-0014). An error or nil location never satisfies a window.
type TimeZone interface {
	Location(context.Context) (*time.Location, error)
}
