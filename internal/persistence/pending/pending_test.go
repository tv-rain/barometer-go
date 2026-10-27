package pending_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/bundle"
	"github.com/tv-rain/barometer-go/internal/config"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/persistence"
	"github.com/tv-rain/barometer-go/internal/persistence/pending"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit"
	"github.com/tv-rain/barometer-go/internal/testkit/records"
)

// Synthetic not_run results only; no measurement, runtime or platform
// evidence.

var epoch = time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)

type configClock struct{}

func (configClock) Now(context.Context) (time.Time, error) { return epoch, nil }

func builder(t testing.TB) *bundle.Builder {
	t.Helper()
	d, err := config.NewDecoder(config.Contract, config.Approvals{})
	if err != nil {
		t.Fatal(err)
	}
	s, err := d.Decode([]byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	i, err := config.Prepare(s, config.Capabilities{Clock: configClock{}})
	if err != nil {
		t.Fatal(err)
	}
	b, err := bundle.New(bundle.Contract, i)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// value builds a valid not_run result with a fresh run identifier, so every
// stored result is distinguishable.
func value(t testing.TB) result.Value {
	t.Helper()
	test := model.Test{ID: "TCP-01", Version: result.Contract, ResultVersion: result.Contract}
	impl := model.Implementation{LibraryVersion: "0.1.0", Platform: "android"}
	d, err := result.New(result.Binding{TaskID: "campaign", Output: result.Output{Test: test}, MaxAttempts: 1, MaxDurationMS: 1000, Implementation: impl})
	if err != nil {
		t.Fatal(err)
	}
	v, err := d.Build(model.Result{SchemaVersion: result.Contract, TaskID: "campaign", RunID: result.NewRunID(), Attempt: 1, Test: test,
		Implementation: impl, Time: model.Time{StartedAt: "2026-10-22T12:00:00Z"}, Execution: model.Execution{State: "not_run", Reason: "constraints_unsatisfied"}})
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func runID(t testing.TB, v result.Value) string {
	t.Helper()
	return v.View().RunID
}

type fixture struct {
	t       *testing.T
	clock   *testkit.Clock
	records *records.Store
	limits  pending.Limits
	store   *pending.Store
}

func newFixture(t *testing.T, limits pending.Limits) *fixture {
	t.Helper()
	f := &fixture{t: t, clock: testkit.NewClock(epoch), records: records.New(), limits: limits}
	f.reopen()
	return f
}

// reopen models a new owner over the same storage, as after a restart.
func (f *fixture) reopen() {
	f.t.Helper()
	s, err := pending.New(f.records, builder(f.t), nil, f.clock, f.limits, time.Second)
	if err != nil {
		f.t.Fatal(err)
	}
	if err := s.Open(context.Background()); err != nil {
		f.t.Fatal("open failed:", err)
	}
	f.store = s
}

func (f *fixture) put(v result.Value) error { return f.store.Put(context.Background(), v) }

func (f *fixture) mustPut() result.Value {
	f.t.Helper()
	v := value(f.t)
	if err := f.put(v); err != nil {
		f.t.Fatal("put failed:", err)
	}
	return v
}

func (f *fixture) next() (string, []byte, bool) {
	f.t.Helper()
	id, b, ok, err := f.store.Next(context.Background())
	if err != nil {
		f.t.Fatal("next failed:", err)
	}
	return id, b, ok
}

func (f *fixture) ack(id string) error { return f.store.Acknowledge(context.Background(), id) }

func (f *fixture) capacity() int {
	f.t.Helper()
	n, err := f.store.Capacity(context.Background())
	if err != nil {
		f.t.Fatal("capacity failed:", err)
	}
	return n
}

// members returns the run identifiers of an encoded bundle, after checking
// that it is canonical and carries the identifier Next reported.
func members(t testing.TB, id string, b []byte) []string {
	t.Helper()
	var w struct {
		SchemaVersion string `json:"schema_version"`
		BundleID      string `json:"bundle_id"`
		Results       []struct {
			RunID string `json:"run_id"`
		} `json:"results"`
	}
	if err := json.Unmarshal(b, &w); err != nil || w.SchemaVersion != bundle.Contract || w.BundleID != id || bytes.Contains(b, []byte("linkage")) {
		t.Fatal("returned bytes are not the canonical bundle")
	}
	out := make([]string, len(w.Results))
	for i, r := range w.Results {
		out[i] = r.RunID
	}
	return out
}

func TestNewRejectsInvalidInputs(t *testing.T) {
	r, b, c := records.New(), builder(t), testkit.NewClock(epoch)
	for name, mk := range map[string]func() (*pending.Store, error){
		"no records":  func() (*pending.Store, error) { return pending.New(nil, b, nil, c, pending.Limits{}, time.Second) },
		"no builder":  func() (*pending.Store, error) { return pending.New(r, nil, nil, c, pending.Limits{}, time.Second) },
		"no clock":    func() (*pending.Store, error) { return pending.New(r, b, nil, nil, pending.Limits{}, time.Second) },
		"no deadline": func() (*pending.Store, error) { return pending.New(r, b, nil, c, pending.Limits{}, 0) },
		"bundles above max": func() (*pending.Store, error) {
			return pending.New(r, b, nil, c, pending.Limits{Bundles: 33}, time.Second)
		},
		"negative bundles": func() (*pending.Store, error) {
			return pending.New(r, b, nil, c, pending.Limits{Bundles: -1}, time.Second)
		},
		"bytes above max": func() (*pending.Store, error) {
			return pending.New(r, b, nil, c, pending.Limits{Bytes: 4194305}, time.Second)
		},
		"bytes below one result": func() (*pending.Store, error) {
			return pending.New(r, b, nil, c, pending.Limits{Bytes: 65611}, time.Second)
		},
	} {
		if s, err := mk(); s != nil || !errors.Is(err, pending.ErrConfig) {
			t.Errorf("%s: accepted", name)
		}
	}
	if s, err := pending.New(r, b, nil, c, pending.Limits{Bundles: 32, Bytes: 4194304}, time.Second); err != nil || fmt.Sprintf("%v %#v", s, s) != "pending.Store pending.Store" {
		t.Fatal("registered maxima rejected or store formatting exposes state")
	}
}

// Results are grouped into bundles of at most 16; a returned bundle is
// sealed and returned unchanged until acknowledgement; acknowledgement of
// one bundle leaves the others.
func TestGroupingRetrievalAndAcknowledgement(t *testing.T) {
	f := newFixture(t, pending.Limits{})
	if _, _, ok := f.next(); ok {
		t.Fatal("empty store returned a bundle")
	}
	var want []string
	for range 17 {
		want = append(want, runID(t, f.mustPut()))
	}
	if f.records.Len(persistence.PendingClass) != 2 {
		t.Fatal("17 results are not two bundles")
	}
	id1, b1, ok := f.next()
	if !ok || fmt.Sprint(members(t, id1, b1)) != fmt.Sprint(want[:16]) {
		t.Fatal("first bundle does not hold the first 16 results in order")
	}
	// A later result joins the open second bundle, never the returned one.
	want = append(want, runID(t, f.mustPut()))
	if again, b, _ := f.next(); again != id1 || !bytes.Equal(b, b1) {
		t.Fatal("repeated retrieval changed the bundle")
	}
	if err := f.ack("AAAAAAAAAAAAAAAAAAAAAA"); !errors.Is(err, pending.ErrUnknown) {
		t.Fatal("unknown identifier:", err)
	}
	if err := f.ack(""); !errors.Is(err, pending.ErrUnknown) {
		t.Fatal("empty identifier:", err)
	}
	if err := f.ack(id1); err != nil {
		t.Fatal(err)
	}
	if err := f.ack(id1); !errors.Is(err, pending.ErrUnknown) {
		t.Fatal("repeated acknowledgement:", err)
	}
	id2, b2, ok := f.next()
	if !ok || id2 == id1 || fmt.Sprint(members(t, id2, b2)) != fmt.Sprint(want[16:]) {
		t.Fatal("second bundle wrong")
	}
	// The second bundle is now returned, so the next result starts a third.
	want = append(want, runID(t, f.mustPut()))
	if again, b, _ := f.next(); again != id2 || !bytes.Equal(b, b2) {
		t.Fatal("result appended to a returned bundle")
	}
	if f.ack(id2) != nil {
		t.Fatal("acknowledgement failed")
	}
	id3, b3, _ := f.next()
	if fmt.Sprint(members(t, id3, b3)) != fmt.Sprint(want[18:]) {
		t.Fatal("third bundle wrong")
	}
	if f.ack(id3) != nil || f.records.Len(persistence.PendingClass) != 0 {
		t.Fatal("acknowledged bundles remain stored")
	}
	for _, id := range want {
		if bytes.Contains(f.records.Dump(), []byte(id)) {
			t.Fatal("acknowledged result bytes remain in storage")
		}
	}
}

// The identifier of the open bundle was never returned, so it does not
// acknowledge it.
func TestOpenBundleIdentifierIsNotAcknowledged(t *testing.T) {
	f := newFixture(t, pending.Limits{})
	f.mustPut()
	keys, _ := f.records.List(context.Background(), persistence.PendingClass)
	raw, _ := f.records.Value(keys[0])
	var w struct {
		ID string `json:"bundle_id"`
	}
	if json.Unmarshal(raw[18:], &w) != nil || w.ID == "" {
		t.Fatal("stored record unreadable")
	}
	if err := f.ack(w.ID); !errors.Is(err, pending.ErrUnknown) || f.records.Len(persistence.PendingClass) != 1 {
		t.Fatal("open bundle acknowledged:", err)
	}
}

// Retention runs from bundle creation: a bundle is returned before 86,400
// seconds and deleted, never returned, at and after it, on access and on
// opening after a restart.
// requirement-ref: REQ-LIMIT-0024
func TestRetentionBoundary(t *testing.T) {
	for _, tc := range []struct {
		name    string
		age     time.Duration
		kept    bool
		restart bool
	}{
		{"before", pending.Retention - time.Nanosecond, true, false},
		{"at", pending.Retention, false, false},
		{"after", pending.Retention + time.Second, false, false},
		{"before, after restart", pending.Retention - time.Nanosecond, true, true},
		{"at, after restart", pending.Retention, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t, pending.Limits{})
			f.mustPut()
			// A result added later does not extend the bundle's retention.
			_ = f.clock.Advance(time.Hour)
			f.mustPut()
			_ = f.clock.Advance(tc.age - time.Hour)
			if tc.restart {
				f.records.Restart()
				f.reopen()
				if got := f.records.Len(persistence.PendingClass) == 1; got != tc.kept {
					t.Fatal("opening did not enforce retention")
				}
			}
			_, _, ok := f.next()
			if ok != tc.kept || f.records.Len(persistence.PendingClass) == 1 != tc.kept {
				t.Fatalf("returned %v, stored %d", ok, f.records.Len(persistence.PendingClass))
			}
		})
	}
	t.Run("expired returned bundle cannot be acknowledged", func(t *testing.T) {
		f := newFixture(t, pending.Limits{})
		f.mustPut()
		id, _, _ := f.next()
		_ = f.clock.Advance(pending.Retention)
		if err := f.ack(id); !errors.Is(err, pending.ErrUnknown) || f.records.Len(persistence.PendingClass) != 0 {
			t.Fatal("expired bundle acknowledged or kept:", err)
		}
	})
	t.Run("failed expiry deletion never returns the bundle late", func(t *testing.T) {
		f := newFixture(t, pending.Limits{})
		f.mustPut()
		_ = f.clock.Advance(pending.Retention)
		f.records.Inject(records.FailBeforeDelete, true)
		if _, _, _, err := f.store.Next(context.Background()); !errors.Is(err, pending.ErrUnavailable) {
			t.Fatal("expired bundle served while its deletion failed:", err)
		}
		if err := f.put(value(t)); !errors.Is(err, pending.ErrUnavailable) {
			t.Fatal("put proceeded before expired state was deleted:", err)
		}
		f.records.Inject(records.FailBeforeDelete, false)
		if _, _, ok := f.next(); ok || f.records.Len(persistence.PendingClass) != 0 {
			t.Fatal("retry did not delete the expired bundle")
		}
	})
}

// Registered maxima at −1, at and +1: the 16th result joins its bundle and
// the 17th starts another; the 32nd bundle is accepted and the 33rd
// rejected without changing stored state.
// requirement-ref: REQ-LIMIT-0021
// requirement-ref: REQ-LIMIT-0022
func TestCountLimits(t *testing.T) {
	f := newFixture(t, pending.Limits{})
	for i := 1; i <= 17; i++ {
		f.mustPut()
		want := 1
		if i == 17 {
			want = 2
		}
		if f.records.Len(persistence.PendingClass) != want {
			t.Fatalf("after %d results: %d bundles", i, f.records.Len(persistence.PendingClass))
		}
	}
	f = newFixture(t, pending.Limits{})
	for i := 1; i <= pending.MaxBundles; i++ {
		f.mustPut()
		f.reopen() // seals the bundle, so each holds one result
	}
	dump := f.records.Dump()
	if err := f.put(value(t)); !errors.Is(err, pending.ErrFull) || !bytes.Equal(f.records.Dump(), dump) || f.capacity() != 0 {
		t.Fatal("33rd bundle accepted or state changed:", err)
	}
	id, _, _ := f.next()
	if f.ack(id) != nil || f.capacity() == 0 {
		t.Fatal("acknowledgement freed no capacity")
	}
	f.mustPut()
}

// The aggregate byte limit rejects a result that would exceed it and never
// truncates one; Capacity counts only results of the registered maximum
// size in new bundles, so it reaches zero first.
// requirement-ref: REQ-LIMIT-0022
func TestAggregateBytesAndCapacity(t *testing.T) {
	const minimum = 76 + 65536 // one bundle holding one maximum result
	f := newFixture(t, pending.Limits{Bytes: minimum})
	if f.capacity() != 1 {
		t.Fatal("capacity of the smallest store:", f.capacity())
	}
	f.mustPut()
	if f.capacity() != 0 {
		t.Fatal("capacity after a stored result")
	}
	for {
		before := f.records.Dump()
		err := f.put(value(t))
		if errors.Is(err, pending.ErrFull) {
			if !bytes.Equal(before, f.records.Dump()) {
				t.Fatal("rejected result changed storage")
			}
			break
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	keys, _ := f.records.List(context.Background(), persistence.PendingClass)
	total := 0
	for _, k := range keys {
		v, _ := f.records.Value(k)
		total += len(v) - 18
	}
	if total > minimum {
		t.Fatal("aggregate limit exceeded:", total)
	}
	if n := newFixture(t, pending.Limits{}).capacity(); n != 63 {
		t.Fatal("capacity of an empty store at the registered maxima:", n)
	}
	if n := newFixture(t, pending.Limits{Bytes: 76 + 2*65536 + 1}).capacity(); n != 2 {
		t.Fatal("two maximum results in one bundle:", n)
	}
	if n := newFixture(t, pending.Limits{Bytes: 76 + 2*65536}).capacity(); n != 1 {
		t.Fatal("one byte short of two maximum results:", n)
	}
}

// A failed write before durability stores nothing; an uncertain write may
// have stored the result; either way the store rereads storage and seals
// what it finds.
func TestWriteFaults(t *testing.T) {
	for _, tc := range []struct {
		fault   records.Fault
		durable bool
	}{{records.FailBeforePut, false}, {records.CrashDuringPut, false}, {records.FailAfterPut, true}} {
		t.Run(fmt.Sprint(tc.fault), func(t *testing.T) {
			f := newFixture(t, pending.Limits{})
			first := runID(t, f.mustPut())
			f.records.Inject(tc.fault, true)
			second := value(t)
			if err := f.put(second); !errors.Is(err, pending.ErrUnavailable) {
				t.Fatal("failed write reported:", err)
			}
			f.records.Inject(tc.fault, false)
			third := runID(t, f.mustPut())
			id, b, _ := f.next()
			want := []string{first}
			if tc.durable {
				want = append(want, runID(t, second))
			}
			if fmt.Sprint(members(t, id, b)) != fmt.Sprint(want) {
				t.Fatalf("after the failed write the first bundle holds %v", members(t, id, b))
			}
			if f.ack(id) != nil {
				t.Fatal("acknowledgement failed")
			}
			id, b, _ = f.next()
			if fmt.Sprint(members(t, id, b)) != fmt.Sprint([]string{third}) {
				t.Fatal("the result after recovery joined a sealed bundle")
			}
		})
	}
}

// A failed deletion is not acknowledgement: the bundle stays retrievable and
// a retry deletes it. A deletion that removed the record but reported failure
// succeeds on retry.
func TestAcknowledgementFaults(t *testing.T) {
	for _, fault := range []records.Fault{records.FailBeforeDelete, records.FailAfterDelete} {
		t.Run(fmt.Sprint(fault), func(t *testing.T) {
			f := newFixture(t, pending.Limits{})
			f.mustPut()
			f.reopen()
			f.mustPut()
			id, b, _ := f.next()
			f.records.Inject(fault, true)
			if err := f.ack(id); !errors.Is(err, pending.ErrUnavailable) {
				t.Fatal("failed deletion reported:", err)
			}
			f.records.Inject(fault, false)
			if fault == records.FailBeforeDelete {
				if again, b2, _ := f.next(); again != id || !bytes.Equal(b, b2) {
					t.Fatal("bundle not retrievable after a failed deletion")
				}
			}
			if err := f.ack(id); err != nil {
				t.Fatal("retry failed:", err)
			}
			if next, _, ok := f.next(); !ok || next == id || f.records.Len(persistence.PendingClass) != 1 {
				t.Fatal("partial acknowledgement removed the wrong bundles")
			}
		})
	}
	t.Run("record lost under the store", func(t *testing.T) {
		f := newFixture(t, pending.Limits{})
		f.mustPut()
		f.reopen()
		f.mustPut()
		id, _, _ := f.next()
		keys, _ := f.records.List(context.Background(), persistence.PendingClass)
		for _, k := range keys {
			if v, _ := f.records.Value(k); bytes.Contains(v, []byte(id)) {
				_ = f.records.Delete(context.Background(), k)
			}
		}
		if next, _, ok := f.next(); !ok || next == id {
			t.Fatal("a missing record was served")
		}
	})
}

// After a restart every stored bundle is sealed and keeps its bytes,
// identifier and creation time; unreadable stored state fails closed.
func TestRestartRecovery(t *testing.T) {
	f := newFixture(t, pending.Limits{})
	var want []string
	for range 3 {
		want = append(want, runID(t, f.mustPut()))
	}
	f.records.Restart()
	f.reopen()
	id, b, _ := f.next()
	if fmt.Sprint(members(t, id, b)) != fmt.Sprint(want) {
		t.Fatal("restart changed the bundle")
	}
	f.records.Restart()
	f.reopen()
	f.mustPut()
	if again, b2, _ := f.next(); again != id || !bytes.Equal(b, b2) || f.records.Len(persistence.PendingClass) != 2 {
		t.Fatal("a result joined a bundle stored before the restart")
	}

	keys, _ := f.records.List(context.Background(), persistence.PendingClass)
	good, _ := f.records.Value(keys[0])
	for name, raw := range map[string][]byte{
		"empty":            {},
		"short header":     good[:10],
		"wrong format":     append([]byte{2}, good[1:]...),
		"wrong class":      append([]byte{1, byte(persistence.RunClass)}, good[2:]...),
		"zero time":        append(append(append([]byte{}, good[:10]...), make([]byte, 8)...), good[18:]...),
		"truncated bundle": good[:len(good)-1],
		"bad identifier":   bytes.Replace(good, []byte(`"bundle_id":"`), []byte(`"bundle_id":"=`), 1),
		"linkage":          bytes.Replace(good, []byte(`","results":[`), []byte(`","linkage":{},"results":[`), 1),
		"no results":       append(append([]byte{}, good[:bytes.Index(good, []byte(`"results":[`))+11]...), "]}"...),
	} {
		t.Run(name, func(t *testing.T) {
			r := records.New()
			r.Corrupt(keys[0], raw)
			s, _ := pending.New(r, builder(t), nil, f.clock, pending.Limits{}, time.Second)
			if err := s.Open(context.Background()); !errors.Is(err, pending.ErrUnavailable) {
				t.Fatal("invalid stored record accepted")
			}
			if _, err := s.Capacity(context.Background()); !errors.Is(err, pending.ErrUnavailable) {
				t.Fatal("capacity reported over invalid state")
			}
		})
	}
	t.Run("misnamed record", func(t *testing.T) {
		r := records.New()
		k := keys[0]
		k.Name[0] ^= 1
		r.Corrupt(k, good)
		s, _ := pending.New(r, builder(t), nil, f.clock, pending.Limits{}, time.Second)
		if err := s.Open(context.Background()); !errors.Is(err, pending.ErrUnavailable) {
			t.Fatal("misnamed record accepted")
		}
	})
	t.Run("unreadable storage", func(t *testing.T) {
		for _, fault := range []records.Fault{records.FailList, records.FailGet} {
			r := records.New()
			r.Corrupt(keys[0], good)
			r.Inject(fault, true)
			s, _ := pending.New(r, builder(t), nil, f.clock, pending.Limits{}, time.Second)
			if err := s.Open(context.Background()); !errors.Is(err, pending.ErrUnavailable) {
				t.Fatal("unreadable storage opened")
			}
		}
	})
}

// A stalled backend ends at the store deadline, and a caller waiting for the
// store returns when its own context ends.
func TestStallAndCancellation(t *testing.T) {
	r := records.New()
	s, err := pending.New(r, builder(t), nil, testkit.NewClock(epoch), pending.Limits{}, 50*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	r.Inject(records.Stall, true)
	done := make(chan error, 1)
	go func() { done <- s.Open(context.Background()) }()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := s.Capacity(ctx); !errors.Is(err, pending.ErrUnavailable) {
		t.Fatal("cancelled caller not released:", err)
	}
	select {
	case err := <-done:
		if !errors.Is(err, pending.ErrUnavailable) {
			t.Fatal("stalled open reported:", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("stalled open outlived the store deadline")
	}
	r.Inject(records.Stall, false)
	if _, err := s.Capacity(context.Background()); err != nil {
		t.Fatal("store did not recover:", err)
	}
}

// Concurrent writers, retrieval, acknowledgement and retention keep every
// stored result exactly once: acknowledged or still pending.
func TestConcurrentDeliveryAndRetention(t *testing.T) {
	f := newFixture(t, pending.Limits{})
	const writers, each = 4, 24
	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		put   = map[string]bool{}
		acked []string
		stop  = make(chan struct{})
	)
	for range writers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range each {
				v := value(t)
				if err := f.put(v); err == nil {
					mu.Lock()
					put[runID(t, v)] = true
					mu.Unlock()
				} else if !errors.Is(err, pending.ErrFull) {
					t.Error(err)
				}
				_, _ = f.store.Capacity(context.Background())
			}
		}()
	}
	var dwg sync.WaitGroup
	dwg.Add(1)
	go func() {
		defer dwg.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			id, b, ok, err := f.store.Next(context.Background())
			if err != nil {
				t.Error(err)
				return
			}
			if ok && f.ack(id) == nil {
				mu.Lock()
				acked = append(acked, members(t, id, b)...)
				mu.Unlock()
			}
			_ = f.clock.Advance(time.Millisecond)
		}
	}()
	wg.Wait()
	close(stop)
	dwg.Wait()
	seen := map[string]int{}
	for _, id := range acked {
		seen[id]++
	}
	for {
		id, b, ok := f.next()
		if !ok {
			break
		}
		for _, r := range members(t, id, b) {
			seen[r]++
		}
		if f.ack(id) != nil {
			t.Fatal("acknowledgement failed")
		}
	}
	if len(seen) != len(put) {
		t.Fatalf("%d results stored, %d delivered", len(put), len(seen))
	}
	for id, n := range seen {
		if n != 1 || !put[id] {
			t.Fatal("result delivered twice or never stored")
		}
	}
}

// After clearing deletes the stored bundles, Forget drops the index and the
// open bundle, so the store reloads empty storage and returns nothing. It
// does not wait for an operation in progress.
func TestForgetAfterClear(t *testing.T) {
	f := newFixture(t, pending.Limits{})
	f.mustPut()
	if err := persistence.Clear(context.Background(), f.records); err != nil {
		t.Fatal(err)
	}
	if !f.store.Forget() {
		t.Fatal("idle store did not forget")
	}
	if _, _, ok := f.next(); ok || f.capacity() == 0 {
		t.Fatal("forgotten store still holds a cleared bundle")
	}
	f.records.Inject(records.Stall, true)
	v := value(t)
	done := make(chan struct{})
	go func() {
		defer close(done)
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()
		_ = f.store.Put(ctx, v)
	}()
	time.Sleep(50 * time.Millisecond)
	if f.store.Forget() {
		t.Error("forget did not report a busy store")
	}
	<-done
}

// Stored bundles are re-read with the shared strict parser and each result
// with its registered result schema. A hostile stored value leaves the store
// unavailable without writing or deleting a record; an escaped spelling of a
// known name is the same name and is returned byte for byte.
// requirement-ref: REQ-JSON-0003
func TestStrictReread(t *testing.T) {
	f := newFixture(t, pending.Limits{})
	f.mustPut()
	keys, _ := f.records.List(context.Background(), persistence.PendingClass)
	good, _ := f.records.Value(keys[0])
	edit := func(old, new string) []byte {
		if !bytes.Contains(good, []byte(old)) {
			t.Fatal("fixture lacks", old)
		}
		return bytes.Replace(good, []byte(old), []byte(new), 1)
	}
	const state = `"execution":{"state":"not_run"`
	result := good[bytes.Index(good, []byte(`"results":[`))+11 : len(good)-2]
	many := append([]byte{}, good[:bytes.Index(good, []byte(`"results":[`))+11]...)
	for i := range bundle.MaxResults + 1 {
		if i > 0 {
			many = append(many, ',')
		}
		many = append(many, result...)
	}
	many = append(many, "]}"...)
	for name, raw := range map[string][]byte{
		"equal duplicate":            edit(state, state+`,"state":"not_run"`),
		"escaped-equivalent dup":     edit(state, state+`,"state":"not_run"`),
		"escaped unknown name":       edit(state, state+`,"seed":true`),
		"unknown result member":      edit(`"attempt":1`, `"attempt":1,"seed":true`),
		"unregistered test":          edit(`"id":"TCP-01"`, `"id":"TCP-99"`),
		"other result version":       edit(`"schema_version":"0.1.0","task_id"`, `"schema_version":"0.2.0","task_id"`),
		"decimal integer":            edit(`"attempt":1`, `"attempt":1.0`),
		"exponent integer":           edit(`"attempt":1`, `"attempt":1e0`),
		"quoted integer":             edit(`"attempt":1`, `"attempt":"1"`),
		"invalid UTF-8":              edit(`"campaign"`, "\"camp\xffign\""),
		"unpaired surrogate":         edit(`"campaign"`, `"\ud800"`),
		"control character":          edit(`"campaign"`, "\"camp\naign\""),
		"non-object result":          edit(string(result), `null`),
		"trailing comma":             edit(string(result), string(result)+`,`),
		"more results than a bundle": many,
	} {
		t.Run(name, func(t *testing.T) {
			r := records.New()
			r.Corrupt(keys[0], raw)
			s, _ := pending.New(r, builder(t), nil, f.clock, pending.Limits{}, time.Second)
			if err := s.Open(context.Background()); !errors.Is(err, pending.ErrUnavailable) {
				t.Fatal("hostile stored bundle accepted")
			}
			if _, _, _, err := s.Next(context.Background()); !errors.Is(err, pending.ErrUnavailable) {
				t.Fatal("hostile stored bundle returned")
			}
			if stored, _ := r.Value(keys[0]); r.Puts() != 0 || r.Deletes() != 0 || !bytes.Equal(stored, raw) {
				t.Fatal("rejecting a stored bundle changed storage")
			}
		})
	}

	t.Run("escaped known name", func(t *testing.T) {
		raw := edit(state, `"execution":{"state":"not_run"`)
		r := records.New()
		r.Corrupt(keys[0], raw)
		s, _ := pending.New(r, builder(t), nil, f.clock, pending.Limits{}, time.Second)
		id, b, ok, err := s.Next(context.Background())
		if err != nil || !ok || !bytes.Equal(b, raw[18:]) || id == "" {
			t.Fatal("escaped spelling of a known name rejected or altered")
		}
	})

	// A record tampered after loading, keeping its identifier and size, is
	// refused when Next rereads it and stays stored and indexed.
	t.Run("tampered after load", func(t *testing.T) {
		g := newFixture(t, pending.Limits{})
		g.mustPut()
		raw, _ := g.records.Value(keys[0])
		old := []byte(`"reason":"constraints_unsatisfied"`)
		tampered := bytes.Replace(raw, old, []byte(`"state":"xconstraints_unsatisfied"`), 1)
		if len(tampered) != len(raw) || bytes.Equal(tampered, raw) {
			t.Fatal("fixture lacks an equal-length duplicate")
		}
		g.records.Corrupt(keys[0], tampered)
		puts, deletes := g.records.Puts(), g.records.Deletes()
		for range 2 {
			if _, _, _, err := g.store.Next(context.Background()); !errors.Is(err, pending.ErrUnavailable) {
				t.Fatal("tampered bundle returned")
			}
		}
		if stored, _ := g.records.Value(keys[0]); g.records.Puts() != puts || g.records.Deletes() != deletes || !bytes.Equal(stored, tampered) {
			t.Fatal("refusing a tampered bundle changed storage")
		}
		g.records.Corrupt(keys[0], raw)
		if _, b, ok := g.next(); !ok || !bytes.Equal(b, raw[18:]) {
			t.Fatal("the index lost the bundle")
		}
	})
}
