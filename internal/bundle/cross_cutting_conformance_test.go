package bundle_test

import (
	"bytes"
	"cmp"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/bundle"
	"github.com/tv-rain/barometer-go/internal/jsondecode"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/persistence"
	"github.com/tv-rain/barometer-go/internal/persistence/pending"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit"
	"github.com/tv-rain/barometer-go/internal/testkit/records"
)

// CC-LIMIT-RESULT-AND-BUNDLE-LIMITS (DEV-563): every encoded, count and
// pending-storage maximum of REQ-LIMIT-0021 at −1, at and +1 where a fixture
// reaches it, through the bundle producer and the production pending store
// over the fault-injectable record port. A result that does not fit starts
// another bundle or is rejected with a bounded error; no stored result is
// truncated. Synthetic core contract cases; no execution, runtime, platform
// or durability evidence.
// requirement-ref: REQ-LIMIT-0021
// requirement-ref: REQ-LIMIT-0022
func TestCcLimitResultAndBundleLimits(t *testing.T) {
	t.Run("encoded Measurement Result: 65,536 bytes", func(t *testing.T) {
		// Produced results stay below the maximum with these fixtures, so
		// the boundary is exercised on the strict import boundary, padded
		// with insignificant whitespace.
		ref := ref("TCP-01")
		d, err := result.New(result.Binding{TaskID: "campaign", Output: result.Output{Test: ref}, MaxAttempts: 1, MaxDurationMS: 1000, Implementation: model.Implementation{LibraryVersion: "0.1.0", Platform: "android"}})
		if err != nil {
			t.Fatal(err)
		}
		input, _ := notRun(t).MarshalJSON()
		for _, size := range []int{65535, 65536, 65537} {
			_, err := d.Decode(append(bytes.Repeat([]byte(" "), size-len(input)), input...))
			if size <= 65536 && err != nil || size > 65536 && !errors.Is(err, jsondecode.ErrBounds) {
				t.Fatalf("%d bytes: %v", size, err)
			}
		}
	})

	t.Run("results per bundle: 16", func(t *testing.T) {
		s := newPendingFixture(t, pending.Limits{})
		var want [][]byte
		for i := 1; i <= 17; i++ {
			want = append(want, s.put(t, notRun(t)))
			if n := s.records.Len(persistence.PendingClass); n != 1+i/17 {
				t.Fatalf("%d results in %d bundles", i, n)
			}
		}
		s.intact(t, want, []int{16, 1})
	})

	t.Run("encoded bundle: 262,144 bytes", func(t *testing.T) {
		for _, target := range []int{bundle.MaxBytes - 1, bundle.MaxBytes, bundle.MaxBytes + 1} {
			s := newPendingFixture(t, pending.Limits{})
			var want [][]byte
			for _, r := range sized(t, builder(t, false), target) {
				want = append(want, s.put(t, r))
			}
			split := []int{6}
			if target > bundle.MaxBytes {
				split = []int{5, 1} // the sixth result starts another bundle
			}
			sizes := s.intact(t, want, split)
			if target <= bundle.MaxBytes && sizes[0] != target {
				t.Fatalf("bundle of %d bytes, want %d", sizes[0], target)
			}
		}
	})

	t.Run("pending bundles: 32", func(t *testing.T) {
		s := newPendingFixture(t, pending.Limits{})
		for i := 1; i <= pending.MaxBundles; i++ {
			s.put(t, notRun(t))
			s.reopen(t) // seals the bundle, so each holds one result
		}
		dump := s.records.Dump()
		if err := s.store.Put(context.Background(), notRun(t)); !errors.Is(err, pending.ErrFull) || !bytes.Equal(dump, s.records.Dump()) {
			t.Fatal("33rd bundle accepted or storage changed:", err)
		}
		if n, err := s.store.Capacity(context.Background()); err != nil || n != 0 {
			t.Fatal("capacity at the bundle maximum:", n, err)
		}
	})

	t.Run("aggregate pending bytes: 4,194,304", func(t *testing.T) {
		b := builder(t, false)
		for _, last := range []int{bundle.MaxBytes - 1, bundle.MaxBytes} {
			s := newPendingFixture(t, pending.Limits{})
			full := sized(t, b, bundle.MaxBytes)
			for i := range 16 {
				items := full
				if i == 15 {
					items = sized(t, b, last)
				}
				for _, r := range items {
					s.put(t, r)
				}
				s.reopen(t)
			}
			total := 0
			for _, size := range s.intact(t, nil, nil) {
				total += size
			}
			if total != pending.MaxBytes-bundle.MaxBytes+last {
				t.Fatalf("aggregate %d bytes", total)
			}
			// Any further result exceeds the aggregate maximum.
			dump := s.records.Dump()
			if err := s.store.Put(context.Background(), notRun(t)); !errors.Is(err, pending.ErrFull) || !bytes.Equal(dump, s.records.Dump()) {
				t.Fatalf("aggregate %d + 1 result accepted: %v", total, err)
			}
			if n, err := s.store.Capacity(context.Background()); err != nil || n != 0 {
				t.Fatal("capacity at the aggregate maximum:", n, err)
			}
		}
	})

	t.Run("capacity reaches zero before storage is exhausted", func(t *testing.T) {
		s := newPendingFixture(t, pending.Limits{})
		big := largeResult(t, 8*3*8*246) // the largest fixture result
		accepted := 0
		for {
			n, err := s.store.Capacity(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if n == 0 {
				break
			}
			if err := s.store.Put(context.Background(), big); err != nil {
				t.Fatalf("a result within reported capacity %d was rejected: %v", n, err)
			}
			accepted++
		}
		if accepted < 63 {
			t.Fatal("capacity below the registered maxima:", accepted)
		}
	})
}

type pendingFixture struct {
	records *records.Store
	clock   *testkit.Clock
	limits  pending.Limits
	store   *pending.Store
}

func newPendingFixture(t *testing.T, limits pending.Limits) *pendingFixture {
	t.Helper()
	f := &pendingFixture{records: records.New(), clock: testkit.NewClock(time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)), limits: limits}
	f.reopen(t)
	return f
}

// reopen opens a new store over the same storage, sealing every bundle. It
// uses the client's default 5-second lifecycle deadline: strictly re-reading
// 4,194,304 pending bytes under the race detector on a shared runner can
// exceed one second.
func (f *pendingFixture) reopen(t *testing.T) {
	t.Helper()
	s, err := pending.New(f.records, builder(t, false), nil, f.clock, f.limits, 5*time.Second)
	if err != nil || s.Open(context.Background()) != nil {
		t.Fatal("pending store did not open")
	}
	f.store = s
}

func (f *pendingFixture) put(t *testing.T, r result.Value) []byte {
	t.Helper()
	if err := f.store.Put(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	b, _ := r.MarshalJSON()
	return b
}

// intact checks that the stored bundles, oldest first, hold want byte for
// byte, split as given, and returns their encoded sizes. nil skips a check.
func (f *pendingFixture) intact(t *testing.T, want [][]byte, split []int) []int {
	t.Helper()
	keys, err := f.records.List(context.Background(), persistence.PendingClass)
	if err != nil {
		t.Fatal(err)
	}
	type stored struct {
		seq     uint64
		size    int
		results []json.RawMessage
	}
	var all []stored
	for _, k := range keys {
		raw, _ := f.records.Value(k)
		var w struct {
			Results []json.RawMessage `json:"results"`
		}
		if err := json.Unmarshal(raw[18:], &w); err != nil || len(raw)-18 > bundle.MaxBytes {
			t.Fatal("stored bundle invalid or oversized")
		}
		all = append(all, stored{binary.BigEndian.Uint64(raw[2:10]), len(raw) - 18, w.Results})
	}
	slices.SortFunc(all, func(a, b stored) int { return cmp.Compare(a.seq, b.seq) })
	var got [][]byte
	var counts, sizes []int
	for _, s := range all {
		counts = append(counts, len(s.results))
		sizes = append(sizes, s.size)
		for _, r := range s.results {
			got = append(got, r)
		}
	}
	if split != nil && !slices.Equal(counts, split) {
		t.Fatalf("bundles hold %v results, want %v", counts, split)
	}
	if want != nil && !slices.EqualFunc(got, want, bytes.Equal) {
		t.Fatal("stored results differ from the accepted results")
	}
	return sizes
}

// sized returns six results whose bundle, built without linkage, is exactly
// target bytes, following TestExactEncodedBoundaries.
func sized(t *testing.T, b *bundle.Builder, target int) []result.Value {
	t.Helper()
	base := largeResult(t, 0)
	items := []result.Value{base, base, base, base, base, base}
	v, err := b.Build(items, nil)
	if err != nil {
		t.Fatal(err)
	}
	extra := target - len(encoded(t, v))
	for i := range items {
		n := min(extra, 8*3*8*246)
		extra -= n
		items[i] = largeResult(t, n)
	}
	if extra != 0 {
		t.Fatal("insufficient fixture capacity")
	}
	return items
}
