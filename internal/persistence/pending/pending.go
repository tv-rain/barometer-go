// Package pending is the production pending-result store over the
// persistence.Records port. It groups validated attempt results into
// canonical Result Bundles with internal/bundle, holds them until
// acknowledgement or expiry and reports the capacity that run reservation
// checks before it starts work. It owns the record encoding, bounds and
// retention; the port only stores bytes. It obtains the linkage of a new
// bundle from the derivation it is given and performs no transport or
// upload, and it is not a mobile fallback: a platform backend (DEV-455)
// supplies the port.
package pending

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"slices"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/tv-rain/barometer-go/internal/bundle"
	"github.com/tv-rain/barometer-go/internal/jsondecode"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/persistence"
	"github.com/tv-rain/barometer-go/internal/result"
)

// Registered pending maxima (REQ-LIMIT-0021, public-api.md Results and
// Delivery): pending bundles, their aggregate encoded bytes and the
// unacknowledged retention.
const (
	MaxBundles = 32
	MaxBytes   = 4194304
	Retention  = 86400 * time.Second
)

// maxResult is the registered maximum of one encoded Measurement Result
// (REQ-LIMIT-0021); capacity assumes every future result is this large.
const maxResult = 65536

// envelope is the size of a bundle without linkage and without results: the
// fixed members and the 22-byte identifier.
const envelope = len(`{"schema_version":"0.1.0","bundle_id":"","results":[]}`) + 22

var (
	// ErrUnavailable is returned for any storage failure, unreadable or
	// invalid stored state, cancellation or an expired deadline. It carries
	// no detail.
	ErrUnavailable = errors.New("pending: storage unavailable")
	// ErrFull rejects a result that no bundle can hold within the pending
	// limits. The result is not stored and nothing is truncated.
	ErrFull = errors.New("pending: pending storage full")
	// ErrResult rejects a result the bundle producer does not accept.
	ErrResult = errors.New("pending: invalid result")
	// ErrUnknown is returned for an identifier that names no returned,
	// pending bundle.
	ErrUnknown = errors.New("pending: unknown bundle")
	// ErrConfig rejects incomplete or out-of-range store inputs.
	ErrConfig = errors.New("pending: invalid store configuration")
)

// Clock is the store's time source. Its wall reading is persisted as a
// bundle's creation time, so retention survives a restart.
type Clock interface {
	Now() time.Time
}

// Linkage returns the linkage of a bundle created now, or nil when linkage is
// disabled or unavailable; internal/persistence/linkage supplies it. A nil
// Linkage omits linkage from every bundle.
type Linkage func(context.Context) *model.RotatingLinkage

// Limits are reviewed local pending limits. They may only lower the
// registered maxima; zero selects the maximum.
type Limits struct {
	Bundles int
	Bytes   int
}

// Store is safe for concurrent use. Every operation holds a one-slot
// semaphore that a caller waits for only until its context ends, so no
// uninterruptible lock is held across storage I/O. Every operation is bounded
// by the store's deadline as well as the caller's context.
//
// The newest bundle stays open while this store built it: a result is added
// by rebuilding that bundle with a fresh identifier and replacing its record
// in one atomic write. A bundle is sealed, and never changes again, once it
// is returned by Next, holds as many results as the bundle limits allow, or
// is read back from storage after a restart or an uncertain write. Retention
// runs from a bundle's creation, the persisting of its first result. A bundle
// carries the linkage derived when it is created, so results added while it
// is open keep its epoch and token.
type Store struct {
	sem      chan struct{}
	records  persistence.Records
	builder  *bundle.Builder
	link     Linkage
	reserve  int // bytes linkage can add to a bundle
	clock    Clock
	limits   Limits
	deadline time.Duration

	loaded   bool
	entries  []entry                // retained bundles, oldest first
	open     []result.Value         // members of the newest entry while it is open
	openLink *model.RotatingLinkage // the open bundle's linkage, or nil
	next     uint64                 // sequence number of the next new bundle
}

type entry struct {
	seq     uint64
	id      string
	created time.Time
	size    int // encoded bundle bytes
}

// New returns a store that has not read storage yet. Open, or the first
// operation, loads it.
func New(records persistence.Records, builder *bundle.Builder, link Linkage, clock Clock, limits Limits, deadline time.Duration) (*Store, error) {
	if records == nil || builder == nil || clock == nil || deadline <= 0 {
		return nil, ErrConfig
	}
	if limits.Bundles == 0 {
		limits.Bundles = MaxBundles
	}
	if limits.Bytes == 0 {
		limits.Bytes = MaxBytes
	}
	if limits.Bundles < 1 || limits.Bundles > MaxBundles || limits.Bytes < envelope+builder.LinkageReserve()+maxResult || limits.Bytes > MaxBytes {
		return nil, ErrConfig
	}
	return &Store{sem: make(chan struct{}, 1), records: records, builder: builder, link: link, reserve: builder.LinkageReserve(), clock: clock, limits: limits, deadline: deadline}, nil
}

func (*Store) String() string   { return "pending.Store" }
func (*Store) GoString() string { return "pending.Store" }

// Open reads, validates and indexes every stored bundle, sealing them all,
// and deletes those whose retention has ended. Any unreadable or invalid
// record fails it with ErrUnavailable rather than being dropped.
// requirement-ref: REQ-LIMIT-0024
func (s *Store) Open(ctx context.Context) error {
	ctx, done, err := s.enter(ctx)
	if err != nil {
		return err
	}
	defer done()
	s.loaded = false
	return s.ready(ctx)
}

// Forget drops the in-memory index and the open bundle's results after Clear
// has deleted the stored bundles, so the store reloads storage before its
// next operation. It waits for no operation: when one is in progress it
// leaves the store unchanged and returns false. Dropping references does not
// overwrite memory the garbage collector already copied or has not reclaimed.
func (s *Store) Forget() bool {
	select {
	case s.sem <- struct{}{}:
	default:
		return false
	}
	defer func() { <-s.sem }()
	clear(s.open)
	s.entries, s.open, s.openLink, s.loaded = nil, nil, nil, false
	return true
}

// Put adds one validated result to the open bundle or, when it cannot hold
// it, to a new bundle. A result no bundle can hold within the limits is
// rejected with ErrFull; it is never truncated. ErrUnavailable means the
// result may or may not be durable; the store then rereads storage before
// its next operation.
// requirement-ref: REQ-LIMIT-0021
// requirement-ref: REQ-LIMIT-0022
func (s *Store) Put(ctx context.Context, v result.Value) error {
	ctx, done, err := s.enter(ctx)
	if err != nil {
		return err
	}
	defer done()
	if err := s.ready(ctx); err != nil {
		return err
	}
	if s.open != nil {
		last := s.entries[len(s.entries)-1]
		b, err := s.builder.Build(append(slices.Clip(s.open), v), s.openLink)
		switch {
		case err == nil:
			encoded, _ := b.MarshalJSON()
			if s.total()-last.size+len(encoded) > s.limits.Bytes {
				return ErrFull
			}
			if err := s.write(ctx, entry{last.seq, b.ID(), last.created, len(encoded)}, encoded); err != nil {
				return err
			}
			s.open = append(slices.Clip(s.open), v)
			return nil
		case !errors.Is(err, jsondecode.ErrBounds):
			return ErrResult
		}
		// The open bundle cannot hold the result: seal it.
		s.open, s.openLink = nil, nil
	}
	if len(s.entries) >= s.limits.Bundles {
		return ErrFull
	}
	l := s.derive(ctx)
	b, err := s.builder.Build([]result.Value{v}, l)
	if err != nil && l != nil {
		// Linkage the builder rejects is omitted, never repaired.
		l = nil
		b, err = s.builder.Build([]result.Value{v}, nil)
	}
	if err != nil {
		return ErrResult
	}
	encoded, _ := b.MarshalJSON()
	if s.total()+len(encoded) > s.limits.Bytes {
		return ErrFull
	}
	// The creation time is kept at the precision it is stored with.
	e := entry{s.next, b.ID(), time.Unix(0, s.clock.Now().UnixNano()), len(encoded)}
	if err := s.write(ctx, e, encoded); err != nil {
		return err
	}
	s.next++
	s.open, s.openLink = []result.Value{v}, l
	return nil
}

// Capacity reports how many more results of the registered maximum size the
// store can certainly accept in new bundles. It ignores free room in the open
// bundle, which Next may seal, so a result it counts is never rejected
// because a bundle was returned in between; it reaches zero before storage is
// exhausted.
func (s *Store) Capacity(ctx context.Context) (int, error) {
	ctx, done, err := s.enter(ctx)
	if err != nil {
		return 0, err
	}
	defer done()
	if err := s.ready(ctx); err != nil {
		return 0, err
	}
	free, n := s.limits.Bytes-s.total(), 0
	for range s.limits.Bundles - len(s.entries) {
		size, count := envelope+s.reserve, 0
		for count < bundle.MaxResults && size+maxResult+min(count, 1) <= bundle.MaxBytes && size+maxResult+min(count, 1) <= free {
			size += maxResult + min(count, 1)
			count++
		}
		if count == 0 {
			break
		}
		free -= size
		n += count
	}
	return n, nil
}

// Next returns the identifier and exact stored bytes of the oldest retained
// bundle, or ok false when none is pending, and seals it. Repeated calls
// return the same bytes until acknowledgement or expiry. A record the index
// holds but storage does not is dropped when found.
// requirement-ref: REQ-LIMIT-0024
func (s *Store) Next(ctx context.Context) (id string, encoded []byte, ok bool, err error) {
	ctx, done, err := s.enter(ctx)
	if err != nil {
		return "", nil, false, err
	}
	defer done()
	if err := s.ready(ctx); err != nil {
		return "", nil, false, err
	}
	for len(s.entries) > 0 {
		e := s.entries[0]
		raw, found, err := s.records.Get(ctx, key(e.seq))
		if err != nil {
			return "", nil, false, ErrUnavailable
		}
		if !found {
			s.drop(0)
			continue
		}
		stored, encoded, err := decode(raw)
		if err != nil || !stored.same(e) {
			return "", nil, false, ErrUnavailable
		}
		if len(s.entries) == 1 {
			s.open, s.openLink = nil, nil
		}
		return e.id, encoded, true, nil
	}
	return "", nil, false, nil
}

// Acknowledge deletes the returned bundle id names. ErrUnknown means no
// retained, returned bundle has that identifier; ErrUnavailable means the
// deletion failed and the bundle may still be stored, so a retry deletes it.
// requirement-ref: REQ-LIMIT-0024
func (s *Store) Acknowledge(ctx context.Context, id string) error {
	ctx, done, err := s.enter(ctx)
	if err != nil {
		return err
	}
	defer done()
	if err := s.ready(ctx); err != nil {
		return err
	}
	for i, e := range s.entries {
		if e.id != id {
			continue
		}
		if s.open != nil && i == len(s.entries)-1 {
			// The open bundle's identifier was never returned.
			return ErrUnknown
		}
		if s.records.Delete(ctx, key(e.seq)) != nil {
			return ErrUnavailable
		}
		s.drop(i)
		return nil
	}
	return ErrUnknown
}

// derive returns the linkage of a new bundle, or nil without a derivation.
func (s *Store) derive(ctx context.Context) *model.RotatingLinkage {
	if s.link == nil {
		return nil
	}
	return s.link(ctx)
}

// enter acquires the semaphore within the caller's context and the store
// deadline and returns the bounded context and its release.
func (s *Store) enter(ctx context.Context) (context.Context, func(), error) {
	ctx, cancel := context.WithTimeout(ctx, s.deadline)
	select {
	case s.sem <- struct{}{}:
	case <-ctx.Done():
		cancel()
		return nil, nil, ErrUnavailable
	}
	return ctx, func() { <-s.sem; cancel() }, nil
}

// ready loads storage when needed and purges expired bundles. A failed
// delete keeps the entry and fails the caller; the next operation retries it.
func (s *Store) ready(ctx context.Context) error {
	if ctx.Err() != nil {
		return ErrUnavailable
	}
	if !s.loaded {
		if err := s.load(ctx); err != nil {
			return err
		}
	}
	now := s.clock.Now()
	for i := 0; i < len(s.entries); {
		if now.Before(s.entries[i].created.Add(Retention)) {
			i++
			continue
		}
		if s.records.Delete(ctx, key(s.entries[i].seq)) != nil {
			return ErrUnavailable
		}
		s.drop(i)
	}
	return nil
}

func (s *Store) load(ctx context.Context) error {
	keys, err := s.records.List(ctx, persistence.PendingClass)
	if err != nil || len(keys) > persistence.MaxClassRecords {
		return ErrUnavailable
	}
	entries := make([]entry, 0, len(keys))
	for _, k := range keys {
		raw, found, err := s.records.Get(ctx, k)
		if err != nil || k.Class != persistence.PendingClass {
			return ErrUnavailable
		}
		if !found {
			continue
		}
		e, _, err := decode(raw)
		if err != nil || key(e.seq) != k {
			return ErrUnavailable
		}
		entries = append(entries, e)
	}
	slices.SortFunc(entries, func(a, b entry) int {
		switch {
		case a.seq < b.seq:
			return -1
		case a.seq > b.seq:
			return 1
		}
		return 0
	})
	s.entries, s.open, s.openLink, s.next, s.loaded = entries, nil, nil, 0, true
	if len(entries) > 0 {
		s.next = entries[len(entries)-1].seq + 1
	}
	return nil
}

// write stores e and replaces its index entry. A failed write is uncertain:
// the store rereads storage before its next operation.
func (s *Store) write(ctx context.Context, e entry, encoded []byte) error {
	if s.records.Put(ctx, key(e.seq), encode(e, encoded)) != nil {
		s.loaded = false
		return ErrUnavailable
	}
	if n := len(s.entries); n > 0 && s.entries[n-1].seq == e.seq {
		s.entries[n-1] = e
	} else {
		s.entries = append(s.entries, e)
	}
	return nil
}

func (e entry) same(o entry) bool {
	return e.seq == o.seq && e.id == o.id && e.size == o.size && e.created.Equal(o.created)
}

func (s *Store) drop(i int) {
	if s.open != nil && i == len(s.entries)-1 {
		s.open, s.openLink = nil, nil
	}
	s.entries = slices.Delete(s.entries, i, i+1)
}

func (s *Store) total() int {
	n := 0
	for _, e := range s.entries {
		n += e.size
	}
	return n
}

// key is the record key of a bundle: a domain-separated digest of its
// sequence number, so no bundle identifier or result value is a platform key.
func key(seq uint64) persistence.RecordKey {
	b := binary.BigEndian.AppendUint64([]byte("barometer/pending-bundle/v1\x00"), seq)
	return persistence.RecordKey{Class: persistence.PendingClass, Name: sha256.Sum256(b)}
}

// Stored values: a format byte, the class, the sequence number and the
// creation time in Unix nanoseconds, big-endian, then the exact bundle bytes.
const (
	storedFormat = 1
	headerSize   = 18
	idPrefix     = `{"schema_version":"0.1.0","bundle_id":"`
	idSize       = 22
	resultsStart = `,"results":[`
	linkageStart = `,"linkage":{"mode":"rotating_installation","scope":"`
)

func encode(e entry, encoded []byte) []byte {
	b := make([]byte, 0, headerSize+len(encoded))
	b = append(b, storedFormat, byte(persistence.PendingClass))
	b = binary.BigEndian.AppendUint64(b, e.seq)
	b = binary.BigEndian.AppendUint64(b, uint64(e.created.UnixNano()))
	return append(b, encoded...)
}

// decode validates a stored value strictly: the header, then a bundle in the
// exact form this store writes, with or without linkage, within the bundle
// limit. Linkage is accepted in any reviewed scope syntax, because a bundle
// keeps the linkage of the configuration it was created under.
func decode(raw []byte) (entry, []byte, error) {
	if len(raw) < headerSize || raw[0] != storedFormat || raw[1] != byte(persistence.PendingClass) {
		return entry{}, nil, ErrUnavailable
	}
	seq := binary.BigEndian.Uint64(raw[2:])
	nanos := int64(binary.BigEndian.Uint64(raw[10:]))
	encoded := raw[headerSize:]
	if nanos <= 0 || len(encoded) > bundle.MaxBytes || !bytes.HasPrefix(encoded, []byte(idPrefix)) {
		return entry{}, nil, ErrUnavailable
	}
	id := string(encoded[len(idPrefix):min(len(encoded), len(idPrefix)+idSize)])
	rest, closed := bytes.CutPrefix(encoded[len(idPrefix)+len(id):], []byte(`"`))
	rest, linked := skipLinkage(rest)
	if len(id) != idSize || !base64url(id) || !closed || !linked || !bytes.HasPrefix(rest, []byte(resultsStart)) ||
		len(rest) <= len(resultsStart)+2 || !bytes.HasSuffix(rest, []byte("]}")) ||
		!utf8.Valid(encoded) || !json.Valid(encoded) {
		return entry{}, nil, ErrUnavailable
	}
	return entry{seq, id, time.Unix(0, nanos), len(encoded)}, slices.Clone(encoded), nil
}

// skipLinkage returns b after a leading linkage member in the exact form
// internal/bundle writes, b itself when it has none, and false for a malformed
// one: a 1–128 byte scope of lowercase letters, digits, dot, underscore or
// hyphen, a canonical nonnegative 64-bit epoch and a 43-byte base64url token.
func skipLinkage(b []byte) ([]byte, bool) {
	if !bytes.HasPrefix(b, []byte(`,"linkage":`)) {
		return b, true
	}
	if !bytes.HasPrefix(b, []byte(linkageStart)) {
		return nil, false
	}
	b = b[len(linkageStart):]
	scope, b, ok := bytes.Cut(b, []byte(`","epoch":`))
	if !ok || len(scope) < 1 || len(scope) > 128 || bytes.ContainsFunc(scope, func(r rune) bool {
		return !('a' <= r && r <= 'z' || '0' <= r && r <= '9' || r == '.' || r == '_' || r == '-')
	}) {
		return nil, false
	}
	epoch, b, ok := bytes.Cut(b, []byte(`,"token":"`))
	if _, err := strconv.ParseInt(string(epoch), 10, 64); !ok || err != nil || epoch[0] == '-' || epoch[0] == '+' || len(epoch) > 1 && epoch[0] == '0' {
		return nil, false
	}
	if len(b) < 45 || !base64url(string(b[:43])) || string(b[43:45]) != `"}` {
		return nil, false
	}
	if _, err := base64.RawURLEncoding.Strict().DecodeString(string(b[:43])); err != nil {
		return nil, false
	}
	return b[45:], true
}

func base64url(s string) bool {
	for _, c := range []byte(s) {
		if !('A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' || c == '-' || c == '_') {
			return false
		}
	}
	return true
}
