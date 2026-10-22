package testkit_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/testkit"
)

func TestRawFixturesPreserveHostileBytes(t *testing.T) {
	member, err := testkit.Member("test_id", `"NET-01"`)
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := testkit.Object(member, `"test_\u0069d":"TCP-01"`, `"synthetic_unknown":null`)
	if err != nil {
		t.Fatal(err)
	}
	if string(fixture) != `{"test_id":"NET-01","test_\u0069d":"TCP-01","synthetic_unknown":null}` {
		t.Fatal("raw duplicate/escaped names or unknown field were changed")
	}
	malformed, err := testkit.Object(`"x":`)
	if err != nil || string(malformed) != `{"x":}` {
		t.Fatal("malformed value was normalized")
	}
	depth, err := testkit.NestedArray(1024)
	if err != nil || len(depth) != 2049 {
		t.Fatal("hostile nesting fixture has wrong size")
	}
}

func TestFixtureBoundsAndIndependentCopies(t *testing.T) {
	for _, size := range []int{-1, testkit.MaxFixtureBytes + 1} {
		if _, err := testkit.Bytes(size, 'x'); !errors.Is(err, testkit.ErrFixtureBounds) {
			t.Fatal("unbounded bytes")
		}
	}
	for _, maximum := range []int{0, testkit.MaxFixtureBytes, -1} {
		if _, err := testkit.ByteBoundaries(maximum, 'x'); !errors.Is(err, testkit.ErrFixtureBounds) {
			t.Fatal("unbounded boundary")
		}
	}
	for _, depth := range []int{-1, 1025} {
		if _, err := testkit.NestedArray(depth); !errors.Is(err, testkit.ErrFixtureBounds) {
			t.Fatal("unbounded depth")
		}
	}
	if _, err := testkit.Member(strings.Repeat("a", 4097), "0"); !errors.Is(err, testkit.ErrFixtureBounds) {
		t.Fatal("unbounded member name")
	}
	if _, err := testkit.Member("x", strings.Repeat("0", testkit.MaxFixtureBytes)); !errors.Is(err, testkit.ErrFixtureBounds) {
		t.Fatal("unbounded raw value")
	}
	if _, err := testkit.Object(strings.Repeat("x", testkit.MaxFixtureBytes)); !errors.Is(err, testkit.ErrFixtureBounds) {
		t.Fatal("unbounded object")
	}
	for _, maximum := range []int{1, 65536, 262144} {
		b, err := testkit.ByteBoundaries(maximum, ' ')
		if err != nil || len(b.Below) != maximum-1 || len(b.At) != maximum || len(b.Above) != maximum+1 {
			t.Fatal("incorrect byte boundaries")
		}
		b.At[0] = 'x'
		if b.Above[0] != ' ' {
			t.Fatal("boundary fixtures share storage")
		}
	}
	first, second := testkit.InvalidJSON(), testkit.InvalidJSON()
	first[1].Input[0] = 'x'
	if second[1].Input[0] != '{' {
		t.Fatal("corpus shares mutable storage")
	}
	if len(first) != 11 {
		t.Fatal("missing starter rejection classes")
	}
}

type reporter struct{ messages []string }

func (*reporter) Helper() {}
func (r *reporter) Errorf(format string, args ...any) {
	r.messages = append(r.messages, fmt.Sprintf(format, args...))
}

func TestRejectionDiagnosticsAndClassification(t *testing.T) {
	sentinel := errors.New("expected class")
	c := testkit.RejectionCase{Name: "synthetic", Requirement: "REQ-JSON-0001", Expected: "syntax", Input: []byte("private-payload")}
	matches := func(class string, err error) bool { return class == "syntax" && errors.Is(err, sentinel) }
	r := &reporter{}
	if !testkit.AssertRejected(r, c, fmt.Errorf("wrapper: %w", sentinel), matches) || len(r.messages) != 0 {
		t.Fatal("wrapped expected error failed")
	}
	for _, err := range []error{nil, errors.New("private-error-value")} {
		if testkit.AssertRejected(r, c, err, matches) {
			t.Fatal("accepted input or wrong class passed")
		}
	}
	for _, message := range r.messages {
		if !strings.Contains(message, c.Requirement) || !strings.Contains(message, c.Expected) || strings.Contains(message, "private") {
			t.Fatal("diagnostic omits requirement/class or leaks data")
		}
	}
	if testkit.AssertRejected(r, c, sentinel, nil) {
		t.Fatal("missing matcher passed")
	}
	c.Requirement = ""
	if testkit.AssertRejected(r, c, sentinel, matches) {
		t.Fatal("missing requirement passed")
	}
}

func TestRunnerCopiesInputs(t *testing.T) {
	sentinel := errors.New("synthetic rejection")
	input := []byte("bounded")
	cases := []testkit.RejectionCase{{Name: "copy", Requirement: "synthetic immutable section", Expected: "invalid", Input: input}}
	testkit.RunRejections(t, cases, func(p []byte) error { p[0] = 'x'; return sentinel }, func(class string, err error) bool { return class == "invalid" && errors.Is(err, sentinel) })
	if string(input) != "bounded" {
		t.Fatal("runner changed caller fixture")
	}
}

func TestManualClockAndDeadlineBoundaries(t *testing.T) {
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	b, err := testkit.TimeBoundaries(at, time.Nanosecond)
	if err != nil || !b.Before.Before(at) || !b.At.Equal(at) || !b.After.After(at) {
		t.Fatal("invalid time boundaries")
	}
	if _, err := testkit.TimeBoundaries(at, 0); err == nil {
		t.Fatal("zero quantum accepted")
	}
	clock := testkit.NewClock(b.Before)
	for _, want := range []time.Time{b.At, b.After} {
		if err := clock.Advance(time.Nanosecond); err != nil || !clock.Now().Equal(want) {
			t.Fatal("clock advance failed")
		}
	}
	if err := clock.Advance(-time.Second); err == nil || !clock.Now().Equal(b.After) {
		t.Fatal("clock moved backwards")
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 100 {
				_ = clock.Advance(time.Nanosecond)
				_ = clock.Now()
			}
		})
	}
	wg.Wait()
	if !clock.Now().Equal(b.After.Add(800 * time.Nanosecond)) {
		t.Fatal("concurrent advances lost")
	}
}

func request(t *testing.T, ctx context.Context) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://synthetic.invalid/", nil)
	if err != nil {
		t.Fatal(err)
	}
	return req
}

func TestTransportScriptAndCleanup(t *testing.T) {
	body := []byte("synthetic-body")
	sentinel := errors.New("synthetic failure")
	transport, err := testkit.NewTransport(testkit.ResponseStep{StatusCode: 200, Body: body}, testkit.ResponseStep{Err: sentinel})
	if err != nil {
		t.Fatal(err)
	}
	body[0] = 'x'
	response, err := transport.RoundTrip(request(t, context.Background()))
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(response.Body)
	if err != nil || string(got) != "synthetic-body" {
		t.Fatal("response script mutated")
	}
	if calls, open := transport.Counts(); calls != 1 || open != 1 {
		t.Fatal("missing call/body counts")
	}
	_ = response.Body.Close()
	_ = response.Body.Close()
	if _, err := response.Body.Read(make([]byte, 1)); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal("read after close succeeded")
	}
	if calls, open := transport.Counts(); calls != 1 || open != 0 {
		t.Fatal("close was not idempotent")
	}
	if _, err := transport.RoundTrip(request(t, context.Background())); !errors.Is(err, sentinel) {
		t.Fatal("scripted error changed")
	}
	if _, err := transport.RoundTrip(request(t, context.Background())); !errors.Is(err, testkit.ErrScriptExhausted) {
		t.Fatal("exhausted script did not fail")
	}
}

type notifyingContext struct {
	context.Context
	entered chan struct{}
	once    sync.Once
}

func (c *notifyingContext) Done() <-chan struct{} {
	c.once.Do(func() { close(c.entered) })
	return c.Context.Done()
}

func TestTransportCancellationWithoutSleeps(t *testing.T) {
	transport, err := testkit.NewTransport(testkit.ResponseStep{WaitForCancel: true})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	notify := &notifyingContext{Context: ctx, entered: make(chan struct{})}
	req := request(t, notify)
	done := make(chan error, 1)
	go func() { _, err := transport.RoundTrip(req); done <- err }()
	<-notify.entered
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatal("pending request did not cancel")
	}
	if calls, open := transport.Counts(); calls != 1 || open != 0 {
		t.Fatal("cancellation leaked a body")
	}
	if _, err := transport.RoundTrip(request(t, ctx)); !errors.Is(err, context.Canceled) {
		t.Fatal("pre-cancelled request accepted")
	}
	if calls, _ := transport.Counts(); calls != 1 {
		t.Fatal("pre-cancelled request consumed a step")
	}
}

type closingBody struct {
	io.Reader
	closed bool
}

func (b *closingBody) Close() error { b.closed = true; return nil }

func TestTransportClosesRequestBodyOnErrors(t *testing.T) {
	transport, _ := testkit.NewTransport()
	req := request(t, context.Background())
	body := &closingBody{Reader: bytes.NewReader(nil)}
	req.Body = body
	_, _ = transport.RoundTrip(req)
	if !body.closed {
		t.Fatal("request body leaked on exhaustion")
	}
}

func TestTransportConstructionBounds(t *testing.T) {
	for _, steps := range [][]testkit.ResponseStep{
		make([]testkit.ResponseStep, 65),
		{{StatusCode: 200, Body: make([]byte, testkit.MaxFixtureBytes+1)}},
		{{StatusCode: 200, Body: make([]byte, testkit.MaxFixtureBytes)}, {StatusCode: 200, Body: []byte{0}}},
		{{StatusCode: 200, WaitForCancel: true}},
		{{Err: errors.New("synthetic"), WaitForCancel: true}},
		{{StatusCode: 0}},
	} {
		if _, err := testkit.NewTransport(steps...); err == nil {
			t.Fatal("invalid or unbounded transport accepted")
		}
	}
}

func TestWaitClockAdvancesWithoutTimers(t *testing.T) {
	start := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	c := testkit.NewWaitClock(start)
	if err := c.Wait(context.Background(), 30*time.Second); err != nil || !c.Now().Equal(start.Add(30*time.Second)) {
		t.Fatalf("Wait() = %v, now %v", err, c.Now())
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := c.Wait(ctx, time.Minute); !errors.Is(err, context.Canceled) || !c.Now().Equal(start.Add(30*time.Second)) {
		t.Fatalf("cancelled Wait() = %v, now %v", err, c.Now())
	}
	if err := c.Wait(context.Background(), -time.Second); err == nil {
		t.Fatal("negative wait accepted")
	}
	waits := c.Waits()
	if len(waits) != 1 || waits[0] != 30*time.Second {
		t.Fatalf("Waits() = %v", waits)
	}
	waits[0] = 0
	if c.Waits()[0] != 30*time.Second {
		t.Fatal("Waits() aliases clock state")
	}
}

// AfterFunc timers fire only when Advance reaches them, at or after their
// time, in time order; a stopped timer never fires.
func TestClockTimersFireOnAdvance(t *testing.T) {
	start := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	c := testkit.NewWaitClock(start)
	var fired []string
	c.AfterFunc(2*time.Second, func() { fired = append(fired, "late") })
	c.AfterFunc(time.Second, func() { fired = append(fired, "early") })
	stop := c.AfterFunc(time.Second, func() { fired = append(fired, "stopped") })
	if !stop() || stop() {
		t.Fatal("stop did not report prevention once")
	}
	if c.Timers() != 2 {
		t.Fatalf("Timers() = %d", c.Timers())
	}
	_ = c.Advance(time.Second - time.Nanosecond)
	if len(fired) != 0 {
		t.Fatal("timer fired before its time")
	}
	if err := c.Wait(context.Background(), 2*time.Second); err != nil {
		t.Fatal(err)
	}
	if len(fired) != 2 || fired[0] != "early" || fired[1] != "late" || c.Timers() != 0 {
		t.Fatalf("fired = %v, armed %d", fired, c.Timers())
	}
	at := c.AfterFunc(0, func() { fired = append(fired, "now") })
	_ = c.Advance(0)
	if len(fired) != 3 || at() {
		t.Fatal("zero-delay timer did not fire on the next advance")
	}
}
