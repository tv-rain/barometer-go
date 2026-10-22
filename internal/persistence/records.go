package persistence

import (
	"context"
	"errors"
)

// Record-port bounds. A record value holds at most one encoded Result Bundle
// (262,144 bytes, REQ-LIMIT-0021) and a store header of at most
// RecordHeaderSize bytes, the largest state class over the port; replay and
// run records are far smaller. A class lists at most MaxClassRecords keys.
const (
	RecordNameSize   = 32
	RecordHeaderSize = 64
	MaxRecordValue   = 262144 + RecordHeaderSize
	MaxClassRecords  = 4096
)

// ErrRecordsUnavailable is returned by a Records implementation for any
// failure. It carries no detail.
var ErrRecordsUnavailable = errors.New("persistence: record storage unavailable")

// RecordClass is the closed set of Barometer-owned state classes stored over
// the record port. Clear deletes every class, so a class added here is
// cleared without further registration. Stored values carry the class byte:
// add a class directly before endClass and never renumber one.
type RecordClass uint8

const (
	// ReplayClass holds replay records (REQ-TRUST-0021).
	ReplayClass RecordClass = iota + 1
	// RunClass holds run accounting records.
	RunClass
	// PendingClass holds pending Result Bundles, one per record.
	PendingClass
	// ClearingClass holds the single clearing-in-progress marker (DEV-564).
	ClearingClass

	endClass // not a class; keep last
)

// Valid reports whether c is a known class.
func (c RecordClass) Valid() bool { return c >= ReplayClass && c < endClass }

// RecordKey names one record. Name is a fixed-size digest chosen by the Go
// store, so no issuer, task or other value appears in a platform key.
type RecordKey struct {
	Class RecordClass
	Name  [RecordNameSize]byte
}

// Records is the state record port: bounded bytes keyed within one
// application-scoped storage scope. Encoding, validation, retention, bounds
// and conditional writes belong to the Go stores built on it; an
// implementation only stores bytes. Values carry no secret, task input,
// target, label or participant value. One Go store owns a storage scope at a
// time; enforcing that is the platform backend's (DEV-455).
//
// Every method honors cancellation and the context's finite deadline and
// returns ErrRecordsUnavailable, or the context's error, on any failure.
// Implementations never synchronize records off the device, do no background
// work and never read the clock.
type Records interface {
	// Get returns a copy of the value stored for key; ok is false when none
	// is stored. Absence is reported only when known.
	Get(ctx context.Context, key RecordKey) (value []byte, ok bool, err error)
	// Put atomically stores value for key, replacing any earlier value. An
	// error means the outcome is uncertain: afterwards either the earlier
	// value or value is durable, never a mixture.
	Put(ctx context.Context, key RecordKey, value []byte) error
	// Delete atomically removes key; deleting an absent key succeeds. An error
	// means the record may or may not still be stored, never partially.
	Delete(ctx context.Context, key RecordKey) error
	// List returns every key of class, in any order, including keys whose
	// values a crash may have left unreadable.
	List(ctx context.Context, class RecordClass) ([]RecordKey, error)
}
