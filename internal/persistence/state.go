package persistence

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"sync"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/task"
)

// ErrStateUnavailable is returned by OpenState when the stored state cannot
// be read, is invalid or cannot be purged. It carries no detail.
var ErrStateUnavailable = errors.New("persistence: state unavailable")

// State is the production replay store and run ledger over one Records port.
// It owns the record encoding, validation, retention and bounds; the port only
// stores bytes. Exactly one State may be open over a storage scope at a time.
//
// State keeps an index of the retained records' retention ends and accepted
// generations, built by OpenState and kept by every write. The index holds
// every record that is or may be durable: an uncertain write is indexed as if
// it succeeded, so retention, capacity and the highest generation account for
// it. A record the index holds but the port does not is dropped when found.
//
// Every operation first purges the records overdue at its trusted time, run
// records before replay records, and fails rather than proceeding when a
// delete fails. A record is retained no longer than RetainUntil; a run record
// carries the RetainUntil of the replay record of the same key.
type State struct {
	mu       sync.Mutex
	records  Records
	capacity int
	replay   map[[RecordNameSize]byte]replayEntry
	runs     map[[RecordNameSize]byte]time.Time
}

type replayEntry struct {
	retainUntil time.Time
	generation  uint64
}

// OpenState reads, validates and indexes the state in records, deleting
// records overdue at now and run records without a retained replay record. It
// keeps at most capacity replay records, from 1 to MaxClassRecords; a full
// store rejects commits and never evicts a retained record. Any unreadable or
// invalid record fails it with ErrStateUnavailable rather than being dropped.
// requirement-ref: REQ-LIMIT-0023
func OpenState(ctx context.Context, records Records, capacity int, now time.Time) (*State, error) {
	if records == nil || capacity < 1 || capacity > MaxClassRecords || now.IsZero() {
		return nil, ErrStateUnavailable
	}
	s := &State{records: records, capacity: capacity, replay: map[[RecordNameSize]byte]replayEntry{}, runs: map[[RecordNameSize]byte]time.Time{}}
	keys, err := s.list(ctx, ReplayClass)
	if err != nil {
		return nil, ErrStateUnavailable
	}
	for _, k := range keys {
		r, ok, err := s.getReplay(ctx, k.Name)
		if err != nil {
			return nil, ErrStateUnavailable
		}
		if ok {
			s.replay[k.Name] = replayEntry{r.RetainUntil, r.PolicyGeneration}
		}
	}
	if keys, err = s.list(ctx, RunClass); err != nil {
		return nil, ErrStateUnavailable
	}
	for _, k := range keys {
		_, retain, ok, err := s.getRun(ctx, k.Name)
		if err != nil {
			return nil, ErrStateUnavailable
		}
		if !ok {
			continue
		}
		owner, held := s.replay[k.Name]
		if held && !retain.Equal(owner.retainUntil) {
			return nil, ErrStateUnavailable
		}
		if !held {
			// An orphan: its replay record was purged first.
			if s.records.Delete(ctx, k) != nil {
				return nil, ErrStateUnavailable
			}
			continue
		}
		s.runs[k.Name] = retain
	}
	if s.purge(ctx, now) != nil {
		return nil, ErrStateUnavailable
	}
	return s, nil
}

// Replay returns the replay-store view of s.
func (s *State) Replay() ReplayStore { return replayView{s} }

// Runs returns the run-ledger view of s.
func (s *State) Runs() RunLedger { return runView{s} }

// Forget drops the index after Clear has deleted the records it describes,
// so a later operation finds no retention, capacity or generation of a
// cleared record. It writes nothing. Calling it while records remain only
// loses index entries: a record found later is indexed again on access, but
// retention and generation reads miss it until State is reopened.
func (s *State) Forget() {
	s.mu.Lock()
	defer s.mu.Unlock()
	clear(s.replay)
	clear(s.runs)
}

func (*State) String() string   { return "persistence.State" }
func (*State) GoString() string { return "persistence.State" }

type replayView struct{ s *State }
type runView struct{ s *State }

// Forget on either view forgets the shared State.
func (v replayView) Forget() { v.s.Forget() }
func (v runView) Forget()    { v.s.Forget() }

func (v replayView) Lookup(ctx context.Context, key ReplayKey, now time.Time) (ReplayRecord, bool, error) {
	s := v.s
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ready(ctx, now); err != nil {
		return ReplayRecord{}, false, ErrReplayUnavailable
	}
	name := RecordNameOf(key)
	r, ok, err := s.getReplay(ctx, name)
	switch {
	case errors.Is(err, errInvalid) || ok && r.Key != key:
		return ReplayRecord{}, false, ErrReplayCorrupt
	case err != nil:
		return ReplayRecord{}, false, ErrReplayUnavailable
	case !ok:
		delete(s.replay, name)
		return ReplayRecord{}, false, nil
	case r.Overdue(now):
		if s.records.Delete(ctx, RecordKey{ReplayClass, name}) != nil {
			return ReplayRecord{}, false, ErrReplayUnavailable
		}
		delete(s.replay, name)
		return ReplayRecord{}, false, nil
	}
	s.replay[name] = replayEntry{r.RetainUntil, r.PolicyGeneration}
	return r, true, nil
}

func (v replayView) HighestGeneration(ctx context.Context, now time.Time) (uint64, error) {
	s := v.s
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ready(ctx, now); err != nil {
		return 0, ErrReplayUnavailable
	}
	var highest uint64
	for _, e := range s.replay {
		highest = max(highest, e.generation)
	}
	return highest, nil
}

func (v replayView) Commit(ctx context.Context, rec ReplayRecord, now time.Time) (ReplayRecord, bool, error) {
	s := v.s
	s.mu.Lock()
	defer s.mu.Unlock()
	if !rec.Valid() {
		return ReplayRecord{}, false, ErrReplayRecord
	}
	if err := s.ready(ctx, now); err != nil {
		return ReplayRecord{}, false, ErrReplayUnavailable
	}
	name := RecordNameOf(rec.Key)
	existing, ok, err := s.getReplay(ctx, name)
	switch {
	case errors.Is(err, errInvalid) || ok && existing.Key != rec.Key:
		return ReplayRecord{}, false, ErrReplayCorrupt
	case err != nil:
		return ReplayRecord{}, false, ErrReplayUnavailable
	case ok && !existing.Overdue(now):
		s.replay[name] = replayEntry{existing.RetainUntil, existing.PolicyGeneration}
		return existing, false, nil
	}
	// The record is absent, or overdue and replaced by this write.
	delete(s.replay, name)
	if len(s.replay) >= s.capacity {
		return ReplayRecord{}, false, ErrReplayUnavailable
	}
	s.replay[name] = replayEntry{rec.RetainUntil, rec.PolicyGeneration}
	if s.records.Put(ctx, RecordKey{ReplayClass, name}, encodeReplay(rec)) != nil {
		// Uncertain: the record stays indexed as possibly durable.
		return ReplayRecord{}, false, ErrReplayUnavailable
	}
	return rec, true, nil
}

func (v runView) Lookup(ctx context.Context, key ReplayKey, now time.Time) (RunRecord, bool, error) {
	s := v.s
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ready(ctx, now); err != nil {
		return RunRecord{}, false, ErrRunsUnavailable
	}
	name := RecordNameOf(key)
	r, retain, ok, err := s.getRun(ctx, name)
	switch {
	case err != nil || ok && r.key != key:
		return RunRecord{}, false, ErrRunsUnavailable
	case !ok:
		delete(s.runs, name)
		return RunRecord{}, false, nil
	case !now.Before(retain):
		if s.records.Delete(ctx, RecordKey{RunClass, name}) != nil {
			return RunRecord{}, false, ErrRunsUnavailable
		}
		delete(s.runs, name)
		return RunRecord{}, false, nil
	}
	s.runs[name] = retain
	return r.rec, true, nil
}

func (v runView) Put(ctx context.Context, key ReplayKey, rec RunRecord, now time.Time) error {
	s := v.s
	s.mu.Lock()
	defer s.mu.Unlock()
	if !rec.Valid() {
		return ErrRunsUnavailable
	}
	if err := s.ready(ctx, now); err != nil {
		return ErrRunsUnavailable
	}
	// The run record takes the retention of the replay record of its key; a
	// key without a retained replay record has no accounting.
	name := RecordNameOf(key)
	owner, ok, err := s.getReplay(ctx, name)
	if err != nil || !ok || owner.Key != key || owner.Overdue(now) {
		return ErrRunsUnavailable
	}
	if _, held := s.runs[name]; !held && len(s.runs) >= MaxClassRecords {
		return ErrRunsUnavailable
	}
	s.runs[name] = owner.RetainUntil
	if s.records.Put(ctx, RecordKey{RunClass, name}, encodeRun(key, rec, owner.RetainUntil)) != nil {
		// Uncertain: the record stays indexed as possibly durable.
		return ErrRunsUnavailable
	}
	return nil
}

// ready checks cancellation and trusted time, then purges overdue records.
func (s *State) ready(ctx context.Context, now time.Time) error {
	if ctx.Err() != nil || now.IsZero() {
		return ErrRecordsUnavailable
	}
	return s.purge(ctx, now)
}

// purge deletes every indexed record overdue at now, run records first so no
// run record outlives its replay record. A failed delete keeps the entry and
// fails the caller; the next operation retries it.
func (s *State) purge(ctx context.Context, now time.Time) error {
	for name, retain := range s.runs {
		if !now.Before(retain) {
			if err := s.records.Delete(ctx, RecordKey{RunClass, name}); err != nil {
				return err
			}
			delete(s.runs, name)
		}
	}
	for name, e := range s.replay {
		if !now.Before(e.retainUntil) {
			if _, run := s.runs[name]; run {
				if err := s.records.Delete(ctx, RecordKey{RunClass, name}); err != nil {
					return err
				}
				delete(s.runs, name)
			}
			if err := s.records.Delete(ctx, RecordKey{ReplayClass, name}); err != nil {
				return err
			}
			delete(s.replay, name)
		}
	}
	return nil
}

func (s *State) list(ctx context.Context, class RecordClass) ([]RecordKey, error) {
	keys, err := s.records.List(ctx, class)
	if err != nil || len(keys) > MaxClassRecords {
		return nil, ErrRecordsUnavailable
	}
	for _, k := range keys {
		if k.Class != class {
			return nil, ErrRecordsUnavailable
		}
	}
	return keys, nil
}

// errInvalid marks a stored value that is unreadable, invalid or stored under
// the wrong name.
var errInvalid = errors.New("persistence: invalid stored record")

func (s *State) getReplay(ctx context.Context, name [RecordNameSize]byte) (ReplayRecord, bool, error) {
	raw, ok, err := s.records.Get(ctx, RecordKey{ReplayClass, name})
	if err != nil || !ok {
		return ReplayRecord{}, false, err
	}
	r, err := decodeReplay(raw)
	if err != nil || RecordNameOf(r.Key) != name {
		return ReplayRecord{}, false, errInvalid
	}
	return r, true, nil
}

type storedRun struct {
	key ReplayKey
	rec RunRecord
}

func (s *State) getRun(ctx context.Context, name [RecordNameSize]byte) (storedRun, time.Time, bool, error) {
	raw, ok, err := s.records.Get(ctx, RecordKey{RunClass, name})
	if err != nil || !ok {
		return storedRun{}, time.Time{}, false, err
	}
	r, retain, err := decodeRun(raw)
	if err != nil || RecordNameOf(r.key) != name {
		return storedRun{}, time.Time{}, false, errInvalid
	}
	return r, retain, true, nil
}

// RecordNameOf is the record name of a replay key: a domain-separated SHA-256
// digest, so the port never sees the issuer or task identifier as a key.
func RecordNameOf(k ReplayKey) [RecordNameSize]byte {
	b := append([]byte("barometer/state-record/v1\x00"), byte(len(k.IssuerID)))
	b = append(append(b, k.IssuerID...), byte(len(k.TaskID)))
	return sha256.Sum256(append(b, k.TaskID...))
}

// Stored values: a format byte, the class, the replay key as two
// length-prefixed fields, then fixed-width big-endian fields. Decoding
// consumes the value exactly and validates the record.
const storedFormat = 1

func appendKey(b []byte, class RecordClass, k ReplayKey) []byte {
	b = append(b, storedFormat, byte(class), byte(len(k.IssuerID)))
	b = append(append(b, k.IssuerID...), byte(len(k.TaskID)))
	return append(b, k.TaskID...)
}

func encodeReplay(r ReplayRecord) []byte {
	b := append(appendKey(nil, ReplayClass, r.Key), r.Digest[:]...)
	b = binary.BigEndian.AppendUint64(b, r.PolicyGeneration)
	b = append(b, byte(r.Origin))
	b = binary.BigEndian.AppendUint64(b, uint64(r.ExpiresAt.Unix()))
	return binary.BigEndian.AppendUint64(b, uint64(r.RetainUntil.Unix()))
}

func decodeReplay(raw []byte) (ReplayRecord, error) {
	d := decoder{raw: raw}
	k := d.key(ReplayClass)
	var r ReplayRecord
	r.Key = k
	copy(r.Digest[:], d.take(sha256.Size))
	r.PolicyGeneration = d.uint64()
	r.Origin = task.Origin(d.byte())
	r.ExpiresAt = d.seconds()
	r.RetainUntil = d.seconds()
	if !d.done() || !r.Valid() {
		return ReplayRecord{}, errInvalid
	}
	return r, nil
}

func encodeRun(k ReplayKey, r RunRecord, retain time.Time) []byte {
	b := binary.BigEndian.AppendUint64(appendKey(nil, RunClass, k), uint64(r.Anchor.UnixNano()))
	b = append(b, byte(r.Runs), byte(r.Next))
	b = binary.BigEndian.AppendUint64(b, uint64(r.NextDue.UnixNano()))
	return binary.BigEndian.AppendUint64(b, uint64(retain.Unix()))
}

func decodeRun(raw []byte) (storedRun, time.Time, error) {
	d := decoder{raw: raw}
	k := d.key(RunClass)
	r := RunRecord{Anchor: d.nanos()}
	r.Runs, r.Next = int(d.byte()), int(d.byte())
	r.NextDue = d.nanos()
	retain := d.seconds()
	if !d.done() || !r.Valid() {
		return storedRun{}, time.Time{}, errInvalid
	}
	return storedRun{k, r}, retain, nil
}

// decoder reads a stored value; any short or malformed read marks it bad.
type decoder struct {
	raw []byte
	bad bool
}

func (d *decoder) take(n int) []byte {
	if d.bad || n > len(d.raw) {
		d.bad = true
		return make([]byte, n)
	}
	b := d.raw[:n]
	d.raw = d.raw[n:]
	return b
}

func (d *decoder) byte() byte       { return d.take(1)[0] }
func (d *decoder) uint64() uint64   { return binary.BigEndian.Uint64(d.take(8)) }
func (d *decoder) done() bool       { return !d.bad && len(d.raw) == 0 }
func (d *decoder) nanos() time.Time { return time.Unix(0, int64(d.uint64())).UTC() }

func (d *decoder) seconds() time.Time {
	v := int64(d.uint64())
	if v <= 0 {
		d.bad = true
	}
	return time.Unix(v, 0).UTC()
}

func (d *decoder) key(class RecordClass) ReplayKey {
	if d.byte() != storedFormat || RecordClass(d.byte()) != class {
		d.bad = true
	}
	issuer := string(d.take(int(d.byte())))
	id := string(d.take(int(d.byte())))
	k := ReplayKey{IssuerID: issuer, TaskID: id}
	if !boundedText(k.IssuerID, 64) || !boundedText(k.TaskID, 128) {
		d.bad = true
	}
	return k
}
