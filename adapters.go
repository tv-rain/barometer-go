package barometer

import (
	"context"
	"reflect"
	"time"

	"github.com/tv-rain/barometer-go/internal/admission"
	"github.com/tv-rain/barometer-go/internal/bundle"
	"github.com/tv-rain/barometer-go/internal/config"
	"github.com/tv-rain/barometer-go/internal/execution"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/persistence"
	"github.com/tv-rain/barometer-go/internal/persistence/pending"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/registry"
)

// Reviewed implementation inputs (DEV-600): the default and maximum time a
// lifecycle transition waits for reserved entry points to finish and for the
// persistent-clear capability to return. Expiry reports incomplete work and
// keeps the client blocked; it never enlarges public attempt or cleanup bounds.
const (
	defaultLifecycleDeadline = 5 * time.Second
	maxLifecycleDeadline     = 30 * time.Second
)

// adapters is the implementation-internal seam for trusted, reviewed inputs.
// It has no JSON form and no exported path: only code in this package can bind
// it, and no production code does yet, so Initialize fails closed. Tests bind
// synthetic adapters from internal/testkit. Native wiring belongs to DEV-455.
type adapters struct {
	clock       platform.Clock
	registry    *registry.Registry
	policy      task.LocalPolicy // its Generation is the active policy generation; only replacePolicy changes it
	deliveries  *admission.DeliveryStore
	replay      persistence.ReplayStore
	approvals   config.Approvals
	linkageKeys platform.LinkageKeyStore // optional; absence disables linkage only

	// clearStorage deletes Barometer-owned persistent state for the storage
	// scope (DEV-564). bind derives it from the record port with
	// storageClearer when it is unset; tests may substitute it. Without a
	// record port it stays nil and ClearMeasurementData returns
	// clearing_incomplete.
	clearStorage func(context.Context) error

	// run holds the inputs of run reservation and execution (DEV-636).
	run runAdapters

	// records is the state record port of the storage scope. Each generation
	// opens a pending.Store over it (DEV-563), which is the engine's output
	// and the source of NextBundle; Initialize reads the clearing marker from
	// it and clearing deletes every record class in it. pendingLimits may
	// lower the registered pending maxima; zero selects them.
	records       persistence.Records
	pendingLimits pending.Limits

	// deadline overrides defaultLifecycleDeadline when positive.
	deadline time.Duration
}

// runAdapters are the trusted inputs of StartRun. The clock, jitter source,
// run ledger and path seam are required; initialization fails closed without
// them. The pending-output port is the generation's pending.Store. bind
// derives the path seam from the network capability and the producer
// implementation with newDirectPath when it is unset (DEV-644); tests may
// substitute it. Consent, device state and time zone are optional: a task that
// needs a missing one never runs. The ledger is persistence.State.Runs
// (DEV-641); native network, consent, device and time-zone capabilities and
// the implementation identity belong to DEV-455; tests bind synthetic ones.
type runAdapters struct {
	clock          execution.Clock
	jitter         execution.Jitter
	ledger         persistence.RunLedger
	paths          pathSource
	network        platform.Network
	implementation model.Implementation
	consent        platform.Consent
	device         platform.Device
	zone           platform.TimeZone
	registry       *registry.Registry // the adapters' registry, set by successor
}

// complete reports whether every required run input is present.
func (r runAdapters) complete() bool {
	return !absent(r.clock) && !absent(r.jitter) && !absent(r.ledger) && !absent(r.paths)
}

// storageClearer is the persistent-clear capability over the adapters'
// storage scope, or nil without a record port. It is the registration point
// of clearing: record classes are cleared by persistence.Clear, and every
// other piece of Barometer-owned state an adapter holds is added here as an
// eraser or forgetter. TestClearingCoversEveryAdapter fails for an adapter
// field that is neither listed here nor recorded as holding no such state.
// It captures what it needs, so it reaches the same scope after Close.
func storageClearer(a *adapters) func(context.Context) error {
	if a == nil || absent(a.records) {
		return nil
	}
	records := a.records
	var erasers []persistence.Eraser
	if !absent(a.linkageKeys) {
		// Deleting the key creates no replacement (REQ-LINKAGE-0023 item 8).
		erasers = append(erasers, a.linkageKeys.Delete)
	}
	var forget []persistence.Forgetter
	for _, v := range []any{a.replay, a.run.ledger} {
		if f, ok := v.(persistence.Forgetter); ok && !absent(f) {
			forget = append(forget, f)
		}
	}
	return func(ctx context.Context) error {
		if err := persistence.Clear(ctx, records, erasers...); err != nil {
			return err
		}
		for _, f := range forget {
			f.Forget()
		}
		return nil
	}
}

func (a *adapters) lifecycleDeadline() time.Duration {
	if a == nil || a.deadline <= 0 {
		return defaultLifecycleDeadline
	}
	return a.deadline
}

// runtime is the state of one initialized generation. It is dropped, never
// reused, when the generation ends.
type runtime struct {
	generation uint64
	init       *config.Initialization
	admitter   *admission.Admitter
	tasks      *taskTable
	runs       *handleTable[runRecord]
	runSet     *runSet
	run        runAdapters
	engine     *execution.Engine
	bundles    *pending.Store

	// ctx bounds adapter calls made for this generation. invalidate cancels
	// it, so a transition does not wait on a stalled trusted adapter.
	ctx      context.Context
	cancel   context.CancelFunc
	deadline time.Duration
}

// String and GoString keep handles, adapters and configuration out of
// formatted output, including %s and %v of a containing Client.
func (r *runtime) String() string   { return "barometer.runtime" }
func (r *runtime) GoString() string { return "barometer.runtime" }

func (a *adapters) String() string   { return "barometer.adapters" }
func (a *adapters) GoString() string { return "barometer.adapters" }

func (r *runtime) invalidate() {
	if r != nil {
		r.cancel()
		r.tasks.invalidate()
		r.runs.invalidate()
	}
}

// prepare builds a complete generation or nothing. It decodes only the closed
// host configuration container and performs no network activity.
func prepare(a *adapters, configJSON []byte) *runtime {
	if a == nil || absent(a.clock) || a.registry == nil || a.deliveries == nil || absent(a.replay) || !a.run.complete() || absent(a.records) {
		return nil
	}
	decoder, err := config.NewDecoder(config.Contract, a.approvals)
	if err != nil {
		return nil
	}
	snapshot, err := decoder.Decode(configJSON)
	if err != nil {
		return nil
	}
	var keys platform.LinkageKeyStore
	if !absent(a.linkageKeys) {
		keys = a.linkageKeys
	}
	init, err := config.Prepare(snapshot, config.Capabilities{Clock: a.clock, LinkageKeys: keys})
	if err != nil {
		return nil
	}
	rt := successor(a, init, a.policy)
	if rt == nil {
		return nil
	}
	// Initialization reads and purges pending storage before any use, so
	// bundles that expired while no client was initialized are deleted now.
	ctx, cancel := context.WithTimeout(rt.ctx, rt.deadline)
	defer cancel()
	if rt.bundles.Open(ctx) != nil {
		rt.cancel()
		return nil
	}
	return rt
}

// successor builds the generation that admits under policy, with fresh handle
// tables and a fresh adapter context, or nothing. It does not activate it and
// performs no storage I/O: during policy replacement the current generation
// still owns pending storage, so the successor's store loads it on first use.
func successor(a *adapters, init *config.Initialization, policy task.LocalPolicy) *runtime {
	admitter, err := admission.NewAdmitter(a.deliveries, a.clock, a.registry, policy, a.replay)
	if err != nil {
		return nil
	}
	builder, err := bundle.New(bundle.Contract, init)
	if err != nil {
		return nil
	}
	store, err := pending.New(a.records, builder, a.run.clock, a.pendingLimits, a.lifecycleDeadline())
	if err != nil {
		return nil
	}
	engine, err := execution.NewEngine(a.registry, a.run.clock, a.run.jitter, store)
	if err != nil {
		return nil
	}
	run := a.run
	run.registry = a.registry
	ctx, cancel := context.WithCancel(context.Background())
	return &runtime{init: init, admitter: admitter, tasks: newTaskTable(),
		runs:   newHandleTable[runRecord](runHandlePrefix, runHandleCapacity),
		runSet: newRunSet(), run: run, engine: engine, bundles: store,
		ctx: ctx, cancel: cancel, deadline: a.lifecycleDeadline()}
}

// absent also treats typed nil values in an interface as missing.
func absent(v any) bool {
	if v == nil {
		return true
	}
	r := reflect.ValueOf(v)
	switch r.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return r.IsNil()
	}
	return false
}
