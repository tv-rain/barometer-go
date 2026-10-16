// Package result owns the bounded Measurement Result wire boundary. It performs
// no admission, execution, network activity, persistence or interpretation.
package result

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"slices"
	"time"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
)

const Contract = "0.1.0"

var (
	ErrValue       = errors.New("result: invalid value")
	ErrBinding     = errors.New("result: invalid local binding")
	ErrUnavailable = errors.New("result: unavailable result")
)

// Output selects a reviewed initial Output contract and its effective limits.
// The future registered-test owner derives it from validated task parameters;
// result JSON cannot choose the plan, retention gate, or configured checks.
type Output struct {
	Test               model.Test
	Limits             observation.ProtocolLimits
	DNSMode, QueryType string
	Bootstrap, HTTPS   bool
}
type StepPlan struct {
	ID     string
	Output Output
}
type ProfilePlan struct {
	ID                string
	Version           int64
	RequireValidation bool
	Steps             []StepPlan
}

// Binding is a trusted containing-owner input, not a task/admission capability.
// ContextDecoder/Path bind imported context to local metadata, not wire fields.
// CapturedContext requires retention even for not_run when capture occurred.
// Implementation must identify the reviewed released producer, not host input.
type Binding struct {
	TaskID                     string
	Output                     Output
	MaxAttempts, MaxDurationMS int
	Implementation             model.Implementation
	ContextDecoder             *observation.ContextDecoder
	Path                       observation.Path
	AccessGeneration           uint64
	CapturedContext            bool
	Profile                    *ProfilePlan
}
type Decoder struct {
	ready   bool
	binding Binding
}

func New(b Binding) (*Decoder, error) {
	if !ascii(b.TaskID, 128) || b.MaxAttempts < 1 || b.MaxAttempts > 2 || b.MaxDurationMS < 1 || b.MaxDurationMS > 60000 ||
		!ascii(b.Implementation.LibraryVersion, 64) || !oneOf(b.Implementation.Platform, "android", "ios") || !validOutput(b.Output) {
		return nil, ErrBinding
	}
	if b.Output.Test.ID == "NET-05" {
		if b.Profile == nil || !ascii(b.Profile.ID, 64) || b.Profile.Version < 1 || len(b.Profile.Steps) < 1 || len(b.Profile.Steps) > 8 {
			return nil, ErrBinding
		}
		p := *b.Profile
		p.Steps = slices.Clone(p.Steps)
		seen := map[string]bool{}
		for _, s := range p.Steps {
			if !ascii(s.ID, 64) || seen[s.ID] || !oneOf(s.Output.Test.ID, "DNS-01", "TCP-01", "TLS-01", "HTTP-01") || !validOutput(s.Output) {
				return nil, ErrBinding
			}
			seen[s.ID] = true
		}
		b.Profile = &p
	} else if b.Profile != nil {
		return nil, ErrBinding
	}
	return &Decoder{ready: true, binding: b}, nil
}
func validOutput(o Output) bool {
	if o.Test.Version != Contract || o.Test.ResultVersion != Contract || !oneOf(o.Test.ID, "DNS-01", "DNS-03", "DNS-04", "TCP-01", "TLS-01", "HTTP-01", "NET-01", "NET-03", "NET-05") {
		return false
	}
	l := o.Limits
	if l.MaxBodyBytes < 0 || l.MaxBodyBytes > 65536 || l.MaxRedirects < 0 || l.MaxRedirects > 2 || l.HeaderChecks < 0 || l.HeaderChecks > 8 || l.BodyChecks < 0 || l.BodyChecks > 8 || l.MaxDNSResponses < 0 || l.MaxDNSResponses > 3 {
		return false
	}
	if oneOf(o.Test.ID, "DNS-01", "DNS-03", "DNS-04") && !oneOf(o.QueryType, "A", "AAAA") {
		return false
	}
	if o.Test.ID == "DNS-01" && !oneOf(o.DNSMode, "system", "udp", "tcp") {
		return false
	}
	if o.Test.ID == "NET-03" && (l.MaxRedirects != 0 || !l.StatusCheck) {
		return false
	}
	return true
}

// Value is immutable; its detached domain view cannot serialize independently.
type Value struct {
	value   model.Result
	encoded []byte
}

func (v Value) MarshalJSON() ([]byte, error) {
	if len(v.encoded) == 0 {
		return nil, ErrUnavailable
	}
	return slices.Clone(v.encoded), nil
}
func (*Value) UnmarshalJSON([]byte) error { return model.ErrEncoding }
func (v Value) View() model.Result        { return clone(v.value) }

// Decode uses the shared strict parser before ordinary JSON projection. Every
// failure returns a zero Value and a fixed class, never parser/input text.
// requirement-ref: REQ-RESULT-0003
// requirement-ref: REQ-RESULT-0004
// requirement-ref: REQ-RESULT-0032
// requirement-ref: REQ-RESULT-0038
// requirement-ref: REQ-RESULT-0044
func (d *Decoder) Decode(input []byte) (Value, error) {
	if d == nil || !d.ready {
		return Value{}, ErrBinding
	}
	b := d.binding
	root, err := jsondecode.DecodeSelected(input, jsondecode.Selection{Fields: []string{"schema_version"}, Cases: []jsondecode.Case{{Values: []string{Contract}, Schema: envelope(b.Output.Test.ID)}}}, jsondecode.Limits{MaxBytes: 65536, MaxDepth: 16, MaxNodes: 32768})
	if err != nil {
		return Value{}, err
	}
	var w resultWire
	if json.Unmarshal(input, &w) != nil {
		return Value{}, ErrValue
	}
	if w.Test != testWireOf(b.Output.Test) {
		return Value{}, jsondecode.ErrVersion
	}
	if w.TaskID != b.TaskID || w.Attempt > b.MaxAttempts || w.Implementation != (implementationWire{b.Implementation.LibraryVersion, b.Implementation.Platform}) {
		return Value{}, ErrValue
	}
	run, err := base64.RawURLEncoding.Strict().DecodeString(w.RunID)
	if err != nil || len(run) != 16 || bytes.Equal(run, make([]byte, 16)) || base64.RawURLEncoding.EncodeToString(run) != w.RunID {
		return Value{}, ErrValue
	}
	t, err := time.Parse("2006-01-02T15:04:05Z", w.Time.StartedAt)
	if err != nil || t.Second() != 0 || t.Format("2006-01-02T15:04:05Z") != w.Time.StartedAt || w.Time.Precision != "minute" || w.Time.DurationMS%250 != 0 {
		return Value{}, ErrValue
	}
	capped := w.Time.DurationCapped != nil
	if capped && (!*w.Time.DurationCapped || w.Time.DurationMS != 60000) || !capped && w.Time.DurationMS > quantize(b.MaxDurationMS) {
		return Value{}, ErrValue
	}
	exec, _ := root.Lookup("execution")
	_, hasReason := exec.Lookup("reason")
	switch w.Execution.State {
	case "completed":
		if hasReason || len(w.Measurement) == 0 {
			return Value{}, ErrValue
		}
	case "not_run":
		if !hasReason || !w.Execution.Reason.Valid() || len(w.Measurement) != 0 {
			return Value{}, ErrValue
		}
	case "incomplete":
		if !hasReason || !w.Execution.Reason.Valid() || len(w.Measurement) != 0 && b.Output.Test.ID != "NET-05" {
			return Value{}, ErrValue
		}
	default:
		return Value{}, ErrValue
	}
	v := model.Result{SchemaVersion: Contract, TaskID: w.TaskID, RunID: w.RunID, Attempt: w.Attempt, Test: b.Output.Test, Implementation: b.Implementation,
		Time: model.Time{StartedAt: w.Time.StartedAt, DurationMS: w.Time.DurationMS, DurationCapped: capped}, Execution: model.Execution{State: w.Execution.State, Reason: w.Execution.Reason}}
	if b.Output.Test.ID == "NET-01" {
		if len(w.NetworkContext) != 0 {
			return Value{}, ErrValue
		}
	} else {
		if len(w.NetworkContext) == 0 && (w.Execution.State != "not_run" || b.CapturedContext) {
			return Value{}, ErrValue
		}
		if len(w.NetworkContext) != 0 {
			n, e := b.ContextDecoder.Decode(w.NetworkContext, b.Path)
			if e != nil {
				return Value{}, ErrValue
			}
			v.NetworkContext = &n
			// A completed attempt cannot use a snapshot marked changed.
			if w.Execution.State == "completed" && n.Access().NetworkChanged {
				return Value{}, ErrValue
			}
		}
	}
	if len(w.Measurement) != 0 {
		m, e := d.measurement(w.Measurement, b.Output, w.Execution.State, w.Time.DurationMS)
		if e != nil {
			return Value{}, e
		}
		v.Measurement = m
	}
	encoded, err := encode(v)
	if err != nil || len(encoded) > 65536 {
		return Value{}, jsondecode.ErrBounds
	}
	return Value{v, encoded}, nil
}

// Build is the producer path; typed inputs pass the same complete decoder.
// Bounded preflight happens before encoding; recursive/dynamic payloads cannot
// allocate unbounded output or silently truncate required observations.
func (d *Decoder) Build(v model.Result) (Value, error) {
	if d == nil || !d.ready {
		return Value{}, ErrBinding
	}
	if !preflight(v) {
		return Value{}, jsondecode.ErrBounds
	}
	b, err := encode(v)
	if err != nil {
		return Value{}, ErrValue
	}
	return d.Decode(b)
}

// NewRunID is used once per scheduled run by the future execution owner. Retry
// identity reuse, non-reuse across runs and exclusion from traffic remain runtime
// obligations; validating an imported ID cannot establish its randomness.
// requirement-ref: REQ-RESULT-0006
func NewRunID() string {
	for {
		var b [16]byte
		_, _ = rand.Read(b[:])
		if b != ([16]byte{}) {
			return base64.RawURLEncoding.EncodeToString(b[:])
		}
	}
}

// Timing uses the caller's monotonic-bearing time.Time pair; wall-only input is
// rejected. The local duration retention ceiling is 60,000 ms. Capping retains
// that lower bound only when elapsed time actually exceeds the ceiling.
// requirement-ref: REQ-RESULT-0016
// requirement-ref: REQ-RESULT-0018
func Timing(start, end time.Time) (model.Time, error) {
	if start == start.Round(0) || end == end.Round(0) {
		return model.Time{}, ErrValue
	}
	elapsed := end.Sub(start)
	if elapsed < 0 {
		return model.Time{}, ErrValue
	}
	t := model.Time{StartedAt: start.UTC().Truncate(time.Minute).Format("2006-01-02T15:04:05Z")}
	if elapsed > 60*time.Second {
		t.DurationMS = 60000
		t.DurationCapped = true
	} else {
		t.DurationMS = int(elapsed / time.Millisecond)
		if elapsed%time.Millisecond != 0 {
			t.DurationMS++
		}
		t.DurationMS = quantize(t.DurationMS)
	}
	return t, nil
}
func quantize(n int) int { return (n + 249) / 250 * 250 }
func ascii(s string, max int) bool {
	if len(s) == 0 || len(s) > max {
		return false
	}
	for _, c := range s {
		if c < 32 || c > 126 {
			return false
		}
	}
	return true
}
func oneOf(s string, values ...string) bool { return slices.Contains(values, s) }
func clone(v model.Result) model.Result {
	if v.NetworkContext != nil {
		n := *v.NetworkContext
		v.NetworkContext = &n
	}
	v.Measurement = cloneMeasurement(v.Measurement)
	return v
}
func cloneMeasurement(m *model.Measurement) *model.Measurement {
	if m == nil {
		return nil
	}
	c := *m
	c.Observations = slices.Clone(m.Observations)
	if c.Access != nil {
		a := *c.Access
		c.Access = &a
	}
	if c.Profile != nil {
		p := *c.Profile
		c.Profile = &p
	}
	if c.NetworkValidation != nil {
		n := *c.NetworkValidation
		c.NetworkValidation = &n
	}
	c.Steps = slices.Clone(c.Steps)
	for i := range c.Steps {
		c.Steps[i].Measurement = cloneMeasurement(c.Steps[i].Measurement)
	}
	return &c
}
