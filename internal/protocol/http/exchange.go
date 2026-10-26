// Package http owns bounded HTTP/1.1 exchanges on checked streams. It does
// not select paths, resolve, dial, authenticate TLS, retry, or build results.
package http

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	stdhttp "net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/observation"
)

const OperationBound = 15 * time.Second
const MaxHeaderBytes = 16384

var ErrInterrupted = errors.New("http: exchange interrupted")
var errTimeout = errors.New("http: operation deadline")
var errHeaders = errors.New("http: invalid response headers")

// DestinationRejected reports a local redirect refusal discovered by Open's
// per-contact gate (for example, a private address in a resolved set). Class
// is prohibited_host or unregistered_https. The fixed error text discloses no
// destination or policy detail. Initial-target refusal is an execution error.
type DestinationRejected struct{ Class string }

func (*DestinationRejected) Error() string { return "http: destination rejected" }

type Clock interface {
	Now() time.Time
	AfterFunc(time.Duration, func()) func() bool
}

// Stream is an already established HTTP/1.1 channel. For HTTPS, Conn must
// be the authenticated session's plaintext side wrapped by the attempt's
// meter, with Close closing the underlying attempt connection. HTTP framing,
// headers and all read-ahead therefore count; TLS and CONNECT do not.
type Stream struct {
	Conn   io.ReadWriteCloser
	Family string
}

// Endpoint is supplied by the caller that owns transport stages. Open must
// check registration, expiry, network/path generation and resolved addresses
// before any contact, using the attempt-scoped dialer. It returns only the
// preceding DNS/TCP/TLS observations actually performed. A failed preceding
// stage returns no stream and its registered outcome, with no HTTP observation.
// Redirect must authorize the COMPLETE next target tuple (including checks),
// return a child restricted to that registration, and reject proxy hostname
// hops before resolution or CONNECT. It performs no network I/O. Rejection
// class is unregistered_https or prohibited_host. It is called even when
// following is disabled, solely to classify a safely parsed target.
type Endpoint interface {
	Open(context.Context, *url.URL) (Stream, []observation.Observation, string, error)
	Redirect(*url.URL) (Endpoint, string, error)
}

type Config struct {
	URL string
	// POST is reserved for one DNS wire message, with fixed DNS media types.
	// Empty PostBody means GET. POST never follows redirects.
	PostBody       []byte
	MaxBodyBytes   int
	MaxRedirects   int // zero disables following (including NET-03 and DNS-03)
	ExpectedStatus int // zero means no check
	HeaderChecks   []HeaderCheck
	BodyChecks     []BodyCheck
}

// Result contains only minimized observations. A POST caller may consume the
// bounded transient DNS payload through TakeBody, then parse and discard it.
// It is never serialized or included by formatted diagnostics.
type Result struct {
	Outcome      string
	Observations []observation.Observation
	body         []byte
}

func (r *Result) TakeBody() []byte { b := r.body; r.body = nil; return b }
func (Result) String() string      { return "http.Result" }
func (Result) GoString() string    { return "http.Result" }

// Exchange performs at most one initial request plus two approved redirects.
// The caller records the terminal outcome before calling the returned cleanup;
// cleanup must always be called. This preserves outcome-before-close ordering.
// An interrupted exchange returns no terminal protocol outcome. Its cleanup
// still closes every opened stream. No body is drained or connection reused.
// requirement-ref: REQ-OBS-0023
// requirement-ref: REQ-OBS-0024
// requirement-ref: REQ-OBS-0025
// requirement-ref: REQ-OBS-0026
// requirement-ref: REQ-OBS-0027
// requirement-ref: REQ-LIMIT-0034
// requirement-ref: REQ-LIMIT-0016
// requirement-ref: REQ-LIMIT-0017
// requirement-ref: REQ-LIMIT-0018
func Exchange(ctx context.Context, clock Clock, endpoint Endpoint, cfg Config) (result Result, cleanup func(), err error) {
	cleanup = func() {}
	u, e := validate(cfg)
	if e != nil || ctx.Err() != nil || clock == nil || endpoint == nil {
		return Result{}, cleanup, ErrInterrupted
	}
	for followed := 0; ; {
		if ctx.Err() != nil {
			return Result{}, cleanup, ErrInterrupted
		}
		stream, preceding, outcome, e := endpoint.Open(ctx, u)
		var rejected *DestinationRejected
		if ctx.Err() == nil && followed > 0 && len(preceding) == 0 && errors.As(e, &rejected) && (rejected.Class == "prohibited_host" || rejected.Class == "unregistered_https") {
			if stream.Conn != nil {
				_ = stream.Conn.Close()
			}
			// The previous response observed this redirect, but the refused
			// target was never contacted and gets no invented protocol stage.
			last := len(result.Observations) - 1
			previous := result.Observations[last]
			d, ok := previous.HTTP()
			if !ok || d.Redirect == nil {
				return Result{}, cleanup, ErrInterrupted
			}
			d.Redirect.TerminalLocationClass = rejected.Class
			o, err := makeObservation(time.Duration(previous.DurationMS())*time.Millisecond, "redirect_rejected", d, cfg)
			if err != nil {
				return Result{}, cleanup, ErrInterrupted
			}
			result.Observations[last] = o
			result.Outcome = "redirect_rejected"
			return result, cleanup, nil
		}
		if followed > 0 && (stream.Conn != nil || len(preceding) > 0) {
			last := len(result.Observations) - 1
			previous := result.Observations[last]
			d, ok := previous.HTTP()
			if !ok || d.Redirect == nil {
				return Result{}, cleanup, ErrInterrupted
			}
			d.Redirect.FollowedCount = followed
			o, err := makeObservation(time.Duration(previous.DurationMS())*time.Millisecond, previous.Failure(), d, cfg)
			if err != nil {
				if stream.Conn != nil {
					_ = stream.Conn.Close()
				}
				return Result{}, cleanup, ErrInterrupted
			}
			result.Observations[last] = o
		}
		result.Observations = append(result.Observations, preceding...)
		if e != nil || ctx.Err() != nil {
			if stream.Conn != nil {
				_ = stream.Conn.Close()
			}
			return Result{}, cleanup, ErrInterrupted
		}
		if stream.Conn == nil {
			if outcome == "" {
				return Result{}, cleanup, ErrInterrupted
			}
			result.Outcome = outcome
			return result, cleanup, nil
		}
		if stream.Family != "ipv4" && stream.Family != "ipv6" && stream.Family != "unknown" {
			_ = stream.Conn.Close()
			return Result{}, cleanup, ErrInterrupted
		}
		started := clock.Now()
		op, cancel := context.WithCancelCause(ctx)
		stop := clock.AfterFunc(OperationBound, func() { cancel(errTimeout) })
		unlink := context.AfterFunc(op, func() { _ = stream.Conn.Close() })
		cleanup = func() { stop(); unlink(); cancel(nil); _ = stream.Conn.Close() }
		d := observation.HTTPDetails{AddressFamily: stream.Family, Protocol: "http/1.1"}
		if followed > 0 {
			d.Redirect = &observation.Redirect{FollowedCount: followed, TerminalLocationClass: "registered_https"}
		}
		failure := ""
		var next *url.URL
		var child Endpoint
		response, e := request(stream.Conn, u, cfg)
		if e != nil {
			failure = ioFailure(e)
		} else {
			d.StatusCode = &response.StatusCode
			if redirectStatus(response.StatusCode) {
				class := "invalid"
				next, class = redirectURL(u, response.Header.Get("Location"))
				if class == "registered_https" {
					child, class, e = endpoint.Redirect(next)
					if e != nil || child == nil {
						if class != "prohibited_host" {
							class = "unregistered_https"
						}
					} else {
						class = "registered_https"
					}
				}
				d.Redirect = &observation.Redirect{FollowedCount: followed, TerminalLocationClass: class}
				switch {
				case cfg.MaxRedirects == 0:
					outcome = "redirect_observed"
				case class != "registered_https":
					failure = "redirect_rejected"
				case followed == cfg.MaxRedirects:
					failure = "redirect_limit_reached"
				default:
					outcome = "follow"
					// The next contact has not happened yet.
				}
			} else {
				var body []byte
				body, failure = evaluate(response, &d, cfg)
				if len(cfg.PostBody) > 0 {
					result.body = body
				}
				outcome = "expected_response"
			}
		}
		// Stop timers before publishing a terminal observation. Outer attempt
		// interruption wins over a protocol timeout or a socket/parser error.
		stop()
		unlink()
		if ctx.Err() != nil {
			cleanup()
			return Result{}, func() {}, ErrInterrupted
		}
		if errors.Is(context.Cause(op), errTimeout) {
			failure = "timeout"
		}
		cancel(nil)
		if failure != "" {
			outcome = failure
		}
		o, e := makeObservation(clock.Now().Sub(started), failure, d, cfg)
		if e != nil {
			cleanup()
			return Result{}, func() {}, ErrInterrupted
		}
		result.Observations = append(result.Observations, o)
		if outcome != "follow" || failure != "" {
			result.Outcome = outcome
			return result, cleanup, nil
		}
		cleanup()
		cleanup = func() {}
		followed++
		u, endpoint = next, child
	}
}

func ioFailure(err error) string {
	var timeout interface{ Timeout() bool }
	if errors.Is(err, context.DeadlineExceeded) || errors.As(err, &timeout) && timeout.Timeout() {
		return "timeout"
	}
	return "protocol_error"
}

func request(conn io.ReadWriter, u *url.URL, cfg Config) (*stdhttp.Response, error) {
	method := "GET"
	var payload io.Reader
	if len(cfg.PostBody) > 0 {
		method = "POST"
		payload = bytes.NewReader(cfg.PostBody)
	}
	req, err := stdhttp.NewRequest(method, u.String(), payload)
	if err != nil {
		return nil, errHeaders
	}
	req.Close = true
	req.Header.Set("User-Agent", "") // suppress Go's identifying default
	req.Header.Set("Accept-Encoding", "identity")
	if method == "POST" {
		req.Header.Set("Content-Type", "application/dns-message")
		req.Header.Set("Accept", "application/dns-message")
	}
	if err := req.Write(conn); err != nil {
		return nil, err
	}
	reader := bufio.NewReader(conn)
	// Bound the actual header blocks, including informational responses, not
	// body bytes read ahead by the buffered reader. All read-ahead is metered.
	remaining := MaxHeaderBytes
	for {
		header := make([]byte, 0, 512)
		for {
			if remaining == 0 {
				return nil, errHeaders
			}
			b, err := reader.ReadByte()
			if err != nil {
				return nil, err
			}
			remaining--
			header = append(header, b)
			if bytes.HasSuffix(header, []byte("\r\n\r\n")) {
				break
			}
		}
		parsed := bufio.NewReader(io.MultiReader(bytes.NewReader(header), reader))
		response, err := stdhttp.ReadResponse(parsed, req)
		if err != nil {
			return nil, err
		}
		if response.Proto != "HTTP/1.1" || response.StatusCode < 100 || response.StatusCode > 599 || response.StatusCode == 101 {
			return nil, errHeaders
		}
		if response.StatusCode < 200 {
			continue
		}
		return response, nil
	}
}

func redirectStatus(status int) bool {
	return status == 301 || status == 302 || status == 303 || status == 307 || status == 308
}

func makeObservation(duration time.Duration, failure string, d observation.HTTPDetails, cfg Config) (observation.Observation, error) {
	status := "succeeded"
	if failure != "" {
		status = "failed"
	}
	return observation.NewHTTP(observation.Contract, status, max(0, int(duration/time.Millisecond))/250*250, failure, d, observation.ProtocolLimits{MaxBodyBytes: cfg.MaxBodyBytes, MaxRedirects: cfg.MaxRedirects, StatusCheck: cfg.ExpectedStatus != 0, HeaderChecks: len(cfg.HeaderChecks), BodyChecks: len(cfg.BodyChecks)})
}

func redirectURL(base *url.URL, raw string) (*url.URL, string) {
	if raw == "" || len(raw) > 2048 {
		return nil, "invalid"
	}
	u, e := url.Parse(raw)
	if e != nil {
		return nil, "invalid"
	}
	u = base.ResolveReference(u)
	if u.Scheme != "https" {
		return nil, "non_https"
	}
	if u.User != nil {
		return nil, "credentials"
	}
	if u.Fragment != "" || u.Hostname() == "" || len(u.String()) > 2048 || strings.ContainsAny(u.Host, "\r\n\t ") {
		return nil, "invalid"
	}
	return u, "registered_https"
}
