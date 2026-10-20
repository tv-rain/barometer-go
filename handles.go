package barometer

import (
	"crypto/rand"
	"encoding/base64"
	"sync"

	"github.com/tv-rain/barometer-go/internal/admission"
)

// Reviewed implementation inputs (DEV-600). The registered limits set no
// ceiling on locally admitted tasks or outstanding runs; these capacities bound
// the facade tables only. The task table matches admission.MaxDeliveryCapacity,
// because each admission consumes one adapter-installed delivery record.
const (
	taskHandleCapacity = admission.MaxDeliveryCapacity
	runHandleCapacity  = 256
)

// Handle prefixes keep handles distinct from identifiers: '.' is outside the
// base64url alphabet of run_id and bundle_id, and uppercase is outside task_id.
const (
	taskHandlePrefix = "T."
	runHandlePrefix  = "R."
)

// runRecord is the run-table value; run reservation (DEV-561) extends it.
type runRecord struct{ taskHandle string }

// handleTable maps unguessable local references to values owned by one client
// instance and one initialization generation. A new generation gets a new
// table, so handles from an earlier generation or another instance never
// resolve. Handles are not derived from identifiers, never persisted and never
// formatted by String or GoString.
type handleTable[T any] struct {
	mu          sync.Mutex
	prefix      string
	capacity    int
	invalidated bool
	entries     map[string]T
}

func newHandleTable[T any](prefix string, capacity int) *handleTable[T] {
	return &handleTable[T]{prefix: prefix, capacity: capacity, entries: make(map[string]T)}
}

// add never evicts: a full table returns capacity_exceeded.
func (t *handleTable[T]) add(v T) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.invalidated {
		return "", fail(handleUnavailable)
	}
	if len(t.entries) >= t.capacity {
		return "", fail(capacityExceeded)
	}
	for {
		var b [16]byte
		_, _ = rand.Read(b[:]) // crypto/rand.Read never returns an error.
		h := t.prefix + base64.RawURLEncoding.EncodeToString(b[:])
		if _, taken := t.entries[h]; !taken {
			t.entries[h] = v
			return h, nil
		}
	}
}

func (t *handleTable[T]) lookup(h string) (T, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	v, ok := t.entries[h]
	if !ok {
		var zero T
		return zero, fail(handleUnavailable)
	}
	return v, nil
}

func (t *handleTable[T]) remove(h string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, ok := t.entries[h]
	delete(t.entries, h)
	return ok
}

// invalidate drops every handle and refuses later additions.
func (t *handleTable[T]) invalidate() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.invalidated = true
	clear(t.entries)
}

func (t *handleTable[T]) String() string   { return "barometer.handleTable" }
func (t *handleTable[T]) GoString() string { return "barometer.handleTable" }
