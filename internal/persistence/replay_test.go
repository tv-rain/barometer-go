package persistence_test

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/persistence"
)

var (
	accepted = time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	key      = persistence.ReplayKey{IssuerID: "embedded-issuer", TaskID: "synthetic-task"}
	digest   = sha256.Sum256([]byte("synthetic"))
)

func notAfter(d time.Duration, fraction string) task.Timestamp {
	return task.Timestamp{Whole: accepted.Add(d), Fraction: fraction}
}

// Retention is the task validity plus 86,400 seconds, capped at 691,200 seconds
// after acceptance; a fractional not_after rounds up so the record never ends
// before the task.
// requirement-ref: REQ-LIMIT-0023
// requirement-ref: REQ-TRUST-0021
func TestReplayRecordRetention(t *testing.T) {
	for name, tc := range map[string]struct {
		notAfter          task.Timestamp
		accepted          time.Time
		expires, retained time.Time
	}{
		"one second":               {notAfter(time.Second, ""), accepted, accepted.Add(time.Second), accepted.Add(time.Second + persistence.ReplayGrace)},
		"seven days":               {notAfter(604800*time.Second, ""), accepted, accepted.Add(604800 * time.Second), accepted.Add(persistence.MaxReplayRetention)},
		"fraction rounds up":       {notAfter(time.Hour, "5"), accepted, accepted.Add(time.Hour + time.Second), accepted.Add(time.Hour + time.Second + persistence.ReplayGrace)},
		"cap binds":                {notAfter(604800*time.Second, ""), accepted.Add(-time.Hour), accepted.Add(604800 * time.Second), accepted.Add(-time.Hour + persistence.MaxReplayRetention)},
		"cap truncates acceptance": {notAfter(604800*time.Second, ""), accepted.Add(-time.Hour + 500*time.Millisecond), accepted.Add(604800 * time.Second), accepted.Add(-time.Hour + persistence.MaxReplayRetention)},
	} {
		r, err := persistence.NewReplayRecord(key, digest, 2, task.Embedded, tc.notAfter, tc.accepted)
		if err != nil {
			t.Errorf("%s: record rejected", name)
			continue
		}
		if !r.ExpiresAt.Equal(tc.expires) || !r.RetainUntil.Equal(tc.retained) {
			t.Errorf("%s: expires %v retained %v", name, r.ExpiresAt, r.RetainUntil)
		}
		if r.Key != key || r.Digest != digest || r.PolicyGeneration != 2 || r.Origin != task.Embedded || !r.Valid() {
			t.Errorf("%s: record fields not carried", name)
		}
		if r.Overdue(r.RetainUntil.Add(-time.Nanosecond)) || !r.Overdue(r.RetainUntil) || !r.Overdue(r.RetainUntil.Add(time.Second)) {
			t.Errorf("%s: retention boundary wrong", name)
		}
	}
}

func TestReplayRecordRejectsInvalidValues(t *testing.T) {
	long := strings.Repeat("a", 129)
	for name, build := range map[string]func() (persistence.ReplayRecord, error){
		"empty issuer": func() (persistence.ReplayRecord, error) {
			return persistence.NewReplayRecord(persistence.ReplayKey{TaskID: "t"}, digest, 2, task.Embedded, notAfter(time.Hour, ""), accepted)
		},
		"long issuer": func() (persistence.ReplayRecord, error) {
			return persistence.NewReplayRecord(persistence.ReplayKey{IssuerID: long[:65], TaskID: "t"}, digest, 2, task.Embedded, notAfter(time.Hour, ""), accepted)
		},
		"empty task_id": func() (persistence.ReplayRecord, error) {
			return persistence.NewReplayRecord(persistence.ReplayKey{IssuerID: "i"}, digest, 2, task.Embedded, notAfter(time.Hour, ""), accepted)
		},
		"long task_id": func() (persistence.ReplayRecord, error) {
			return persistence.NewReplayRecord(persistence.ReplayKey{IssuerID: "i", TaskID: long}, digest, 2, task.Embedded, notAfter(time.Hour, ""), accepted)
		},
		"space in task_id": func() (persistence.ReplayRecord, error) {
			return persistence.NewReplayRecord(persistence.ReplayKey{IssuerID: "i", TaskID: "a b"}, digest, 2, task.Embedded, notAfter(time.Hour, ""), accepted)
		},
		"zero digest": func() (persistence.ReplayRecord, error) {
			return persistence.NewReplayRecord(key, [32]byte{}, 2, task.Embedded, notAfter(time.Hour, ""), accepted)
		},
		"zero generation": func() (persistence.ReplayRecord, error) {
			return persistence.NewReplayRecord(key, digest, 0, task.Embedded, notAfter(time.Hour, ""), accepted)
		},
		"no origin": func() (persistence.ReplayRecord, error) {
			return persistence.NewReplayRecord(key, digest, 2, 0, notAfter(time.Hour, ""), accepted)
		},
		"expired": func() (persistence.ReplayRecord, error) {
			return persistence.NewReplayRecord(key, digest, 2, task.Embedded, notAfter(0, ""), accepted)
		},
		"no acceptance": func() (persistence.ReplayRecord, error) {
			return persistence.NewReplayRecord(key, digest, 2, task.Embedded, notAfter(time.Hour, ""), time.Time{})
		},
		"no expiry": func() (persistence.ReplayRecord, error) {
			return persistence.NewReplayRecord(key, digest, 2, task.Embedded, task.Timestamp{}, accepted)
		},
		// The cap would end retention before the task expires.
		"beyond cap": func() (persistence.ReplayRecord, error) {
			return persistence.NewReplayRecord(key, digest, 2, task.Embedded, notAfter(persistence.MaxReplayRetention, ""), accepted)
		},
	} {
		if r, err := build(); !errors.Is(err, persistence.ErrReplayRecord) || r != (persistence.ReplayRecord{}) {
			t.Errorf("%s: invalid record accepted", name)
		}
	}
	r, _ := persistence.NewReplayRecord(key, digest, 2, task.Remote, notAfter(time.Hour, ""), accepted)
	for name, edit := range map[string]func(*persistence.ReplayRecord){
		"retention not after expiry": func(r *persistence.ReplayRecord) { r.RetainUntil = r.ExpiresAt },
		"retention beyond grace": func(r *persistence.ReplayRecord) {
			r.RetainUntil = r.ExpiresAt.Add(persistence.ReplayGrace + time.Second)
		},
		"fractional expiry": func(r *persistence.ReplayRecord) { r.ExpiresAt = r.ExpiresAt.Add(time.Millisecond) },
		"non-UTC retention": func(r *persistence.ReplayRecord) { r.RetainUntil = r.RetainUntil.In(time.FixedZone("x", 3600)) },
	} {
		c := r
		edit(&c)
		if c.Valid() {
			t.Errorf("%s: invalid record reported valid", name)
		}
	}
	if fmt.Sprintf("%v %+v %#v %s", r, r, r, r) != "persistence.ReplayRecord persistence.ReplayRecord persistence.ReplayRecord persistence.ReplayRecord" {
		t.Error("record formatting exposes the replay key")
	}
}
