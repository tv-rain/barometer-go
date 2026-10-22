package barometer

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"

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

// runRecord is the run-table value. A record without a status was never
// launched and is terminal.
type runRecord struct {
	taskHandle string
	status     *runStatus
}

func (r runRecord) terminal() bool { return r.status == nil || r.status.done.Load() }

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

// removeIf drops every entry for which done reports true.
func (t *handleTable[T]) removeIf(done func(T) bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for h, v := range t.entries {
		if done(v) {
			delete(t.entries, h)
		}
	}
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

// taskTable is the task-handle table plus an index from replay identity to
// handle, so an exact redelivery resolves to the handle already published in
// this generation. Publication is serialized, so concurrent identical
// deliveries share one handle.
type taskTable struct {
	publishing sync.Mutex
	byIdentity map[admission.Identity]string
	*handleTable[admission.Admitted]
}

func newTaskTable() *taskTable {
	return &taskTable{byIdentity: make(map[admission.Identity]string),
		handleTable: newHandleTable[admission.Admitted](taskHandlePrefix, taskHandleCapacity)}
}

// publish returns the handle bound to a's identity, or binds a new one. A full
// table first drops tasks whose validity ended at or before now; it never
// evicts a task that is still valid and otherwise returns capacity_exceeded.
func (t *taskTable) publish(a admission.Admitted, now time.Time) (string, error) {
	t.publishing.Lock()
	defer t.publishing.Unlock()
	id := a.Identity()
	if h, ok := t.byIdentity[id]; ok {
		if _, err := t.lookup(h); err == nil {
			return h, nil
		}
		delete(t.byIdentity, id)
	}
	h, err := t.add(a)
	if err == (opError{code: capacityExceeded}) {
		t.removeIf(func(v admission.Admitted) bool { return !v.ExpiresAt().After(now) })
		for known, kh := range t.byIdentity {
			if _, err := t.lookup(kh); err != nil {
				delete(t.byIdentity, known)
			}
		}
		h, err = t.add(a)
	}
	if err != nil {
		return "", err
	}
	t.byIdentity[id] = h
	return h, nil
}

// invalidate drops every handle and identity binding and refuses later
// additions.
func (t *taskTable) invalidate() {
	t.publishing.Lock()
	defer t.publishing.Unlock()
	t.handleTable.invalidate()
	clear(t.byIdentity)
}

func (t *taskTable) String() string   { return "barometer.taskTable" }
func (t *taskTable) GoString() string { return "barometer.taskTable" }
