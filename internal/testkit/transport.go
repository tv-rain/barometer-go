package testkit

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"sync"
)

var ErrScriptExhausted = errors.New("testkit: transport script exhausted")

// ResponseStep is one in-memory HTTP exchange. WaitForCancel blocks only until
// the request context is cancelled; the test must arrange that cancellation.
// Err is a synthetic error. A step cannot combine error/wait/response modes.
type ResponseStep struct {
	StatusCode    int
	Body          []byte
	Err           error
	WaitForCancel bool
}

// Transport performs no network I/O, retains no requests, and never falls back
// to another transport. Bodies and scripts are copied and bounded at creation.
// Counts let tests check absence of traffic and response-body cleanup.
type Transport struct {
	mu                sync.Mutex
	steps             []ResponseStep
	calls, bodiesOpen int
}

func NewTransport(steps ...ResponseStep) (*Transport, error) {
	if len(steps) > 64 {
		return nil, ErrFixtureBounds
	}
	t := &Transport{}
	total := 0
	for _, step := range steps {
		if len(step.Body) > MaxFixtureBytes-total {
			return nil, ErrFixtureBounds
		}
		if (step.Err != nil || step.WaitForCancel) && (step.StatusCode != 0 || len(step.Body) != 0) || step.Err != nil && step.WaitForCancel {
			return nil, errors.New("testkit: conflicting transport step modes")
		}
		if step.Err == nil && !step.WaitForCancel && (step.StatusCode < 100 || step.StatusCode > 599) {
			return nil, errors.New("testkit: invalid response status")
		}
		total += len(step.Body)
		step.Body = bytes.Clone(step.Body)
		t.steps = append(t.steps, step)
	}
	return t, nil
}

func (t *Transport) Counts() (calls, bodiesOpen int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.calls, t.bodiesOpen
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, errors.New("testkit: missing request")
	}
	// RoundTripper owns closing the request body, including error paths.
	if req.Body != nil {
		defer req.Body.Close()
	}
	if err := req.Context().Err(); err != nil {
		return nil, err
	}
	t.mu.Lock()
	index := t.calls
	t.calls++
	if index >= len(t.steps) {
		t.mu.Unlock()
		return nil, ErrScriptExhausted
	}
	step := t.steps[index]
	if step.Err != nil || step.WaitForCancel {
		t.mu.Unlock()
		if step.WaitForCancel {
			done := req.Context().Done()
			if done == nil {
				return nil, errors.New("testkit: waiting step requires cancellable context")
			}
			<-done
			return nil, req.Context().Err()
		}
		return nil, step.Err
	}
	t.bodiesOpen++
	t.mu.Unlock()
	body := &responseBody{reader: bytes.NewReader(step.Body), owner: t, ctx: req.Context()}
	return &http.Response{StatusCode: step.StatusCode, Header: make(http.Header), Body: body, ContentLength: int64(len(step.Body)), Request: req}, nil
}

type responseBody struct {
	mu     sync.Mutex
	reader *bytes.Reader
	owner  *Transport
	ctx    context.Context
	closed bool
}

func (b *responseBody) Read(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return 0, io.ErrClosedPipe
	}
	if err := b.ctx.Err(); err != nil {
		return 0, err
	}
	return b.reader.Read(p)
}

func (b *responseBody) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.closed {
		b.closed = true
		b.owner.mu.Lock()
		b.owner.bodiesOpen--
		b.owner.mu.Unlock()
	}
	return nil
}
