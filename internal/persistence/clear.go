package persistence

import (
	"context"
	"crypto/sha256"
	"errors"
)

// ErrClearIncomplete is returned by Clear when any step fails. Some state may
// already be deleted; the marker stays, so the storage scope stays blocked
// until a later Clear succeeds. It carries no detail.
var ErrClearIncomplete = errors.New("persistence: clearing incomplete")

// Eraser deletes one piece of Barometer-owned state that is not stored over
// the record port, such as the linkage key or, later, the Network Context
// cache (DEV-578). Erasing absent state succeeds. It must not create a
// replacement secret or send a network request.
type Eraser func(context.Context) error

// Forgetter is implemented by Go stores that index records in memory. Forget
// drops that index after Clear has deleted the records it describes, so no
// deleted retention, capacity or generation is still counted.
type Forgetter interface{ Forget() }

// markerKey names the clearing-in-progress marker, the only ClearingClass
// record. Its value is the format byte and the class.
var (
	markerKey   = RecordKey{Class: ClearingClass, Name: sha256.Sum256([]byte("barometer/clearing-marker/v1\x00"))}
	markerValue = []byte{storedFormat, byte(ClearingClass)}
)

// Clear deletes every Barometer-owned record in the storage scope and then
// runs erasers, in this order:
//
//  1. write the clearing-in-progress marker;
//  2. delete every record of every other class, the newest class first, so
//     pending results go before run accounting and run accounting before the
//     replay records it belongs to, then list each class again to confirm it
//     is empty;
//  3. run each eraser;
//  4. delete every marker-class record, the marker last, and confirm none
//     remains.
//
// Any failure stops it with ErrClearIncomplete. The marker stays, and
// Clearing reports it, until a later Clear completes; only an uncertain
// deletion of the marker itself, after everything else is gone, may remove it. Clear is idempotent and
// bounded by ctx. It reads no clock, uploads nothing and opens no connection.
// requirement-ref: REQ-LINKAGE-0023
// requirement-ref: REQ-LINKAGE-0024
// requirement-ref: REQ-TRUST-0025
func Clear(ctx context.Context, records Records, erasers ...Eraser) error {
	if records == nil || ctx.Err() != nil {
		return ErrClearIncomplete
	}
	if records.Put(ctx, markerKey, markerValue) != nil {
		return ErrClearIncomplete
	}
	for _, class := range clearedClasses() {
		if deleteClass(ctx, records, class) != nil {
			return ErrClearIncomplete
		}
	}
	for _, erase := range erasers {
		if erase == nil || erase(ctx) != nil {
			return ErrClearIncomplete
		}
	}
	if deleteMarkers(ctx, records) != nil {
		return ErrClearIncomplete
	}
	return nil
}

// deleteMarkers deletes every ClearingClass record, the marker last, so the
// scope stays marked until nothing else remains.
func deleteMarkers(ctx context.Context, records Records) error {
	keys, err := records.List(ctx, ClearingClass)
	if err != nil || len(keys) > MaxClassRecords {
		return ErrRecordsUnavailable
	}
	for _, k := range keys {
		if k.Class != ClearingClass {
			return ErrRecordsUnavailable
		}
		if k != markerKey && records.Delete(ctx, k) != nil {
			return ErrRecordsUnavailable
		}
	}
	if records.Delete(ctx, markerKey) != nil {
		return ErrRecordsUnavailable
	}
	if keys, err = records.List(ctx, ClearingClass); err != nil || len(keys) > 0 {
		return ErrRecordsUnavailable
	}
	return nil
}

// Clearing reports whether the storage scope holds the clearing marker: a
// Clear started and did not complete, possibly before a restart. A stored
// value other than the marker's still counts as the marker, so the scope
// stays blocked. An error means the answer is unknown.
func Clearing(ctx context.Context, records Records) (bool, error) {
	if records == nil {
		return false, ErrRecordsUnavailable
	}
	_, ok, err := records.Get(ctx, markerKey)
	if err != nil {
		return false, ErrRecordsUnavailable
	}
	return ok, nil
}

// clearedClasses lists every valid class except the marker's, newest first.
func clearedClasses() []RecordClass {
	var classes []RecordClass
	for c := endClass - 1; c.Valid(); c-- {
		if c != ClearingClass {
			classes = append(classes, c)
		}
	}
	return classes
}

func deleteClass(ctx context.Context, records Records, class RecordClass) error {
	for pass := 0; pass < 2; pass++ {
		keys, err := records.List(ctx, class)
		if err != nil || len(keys) > MaxClassRecords {
			return ErrRecordsUnavailable
		}
		if pass == 1 && len(keys) > 0 {
			// A record listed after its deletion succeeded is not cleared.
			return ErrRecordsUnavailable
		}
		for _, k := range keys {
			if k.Class != class || records.Delete(ctx, k) != nil {
				return ErrRecordsUnavailable
			}
		}
	}
	return nil
}
