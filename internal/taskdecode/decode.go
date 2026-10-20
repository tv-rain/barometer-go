// Package taskdecode converts strict task JSON to a normalized candidate. It
// performs no admission, scheduling, logging, network or platform operations.
package taskdecode

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/validation"
)

var ErrParameters = errors.New("taskdecode: parameters")

// ErrUnregistered marks a well-formed (test_id, test_version) pair under the
// supported schema_version that no supplied codec selects. It also matches
// jsondecode.ErrVersion. It describes this decoder, not the task's validity.
var ErrUnregistered = errors.New("taskdecode: unregistered test version")

type unregistered struct{}

func (unregistered) Error() string { return ErrUnregistered.Error() }
func (unregistered) Is(target error) bool {
	return target == ErrUnregistered || target == jsondecode.ErrVersion
}

type Definition struct {
	TestID, TestVersion string
	Input, Options      *jsondecode.Schema
}

// TestCodec is a narrow test-owned schema/normalization boundary. Implementations
// must be immutable, pure and bounded, check test-specific semantics, return
// typed domain values, and never perform I/O. This is not runner registration.
type TestCodec interface {
	Definition() Definition
	Normalize(input, options *jsondecode.Value) (task.Parameters, error)
}

type Decoder struct {
	codecs      []TestCodec
	definitions []Definition
}

var testID = regexp.MustCompile(`^[A-Z]{2,4}-[0-9]{2}$`)
var version = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

// New selects only supplied exact test/version pairs. No runtime registry,
// measurement support or empty-object fallback is installed by this package.
// Descriptors and codecs must remain immutable for the decoder's lifetime.
func New(codecs []TestCodec) (*Decoder, error) {
	if len(codecs) == 0 || len(codecs) > 256 {
		return nil, jsondecode.ErrConfig
	}
	d := &Decoder{}
	for _, codec := range codecs {
		if codec == nil {
			return nil, jsondecode.ErrConfig
		}
		def := codec.Definition()
		if !testID.MatchString(def.TestID) || len(def.TestID) > 64 || !version.MatchString(def.TestVersion) || len(def.TestVersion) > 64 || def.Input == nil || def.Options == nil || def.Input.Kind != jsondecode.Object || def.Options.Kind != jsondecode.Object || def.Input.Nullable || def.Options.Nullable {
			return nil, jsondecode.ErrConfig
		}
		for _, prior := range d.definitions {
			if prior.TestID == def.TestID && prior.TestVersion == def.TestVersion {
				return nil, jsondecode.ErrConfig
			}
		}
		// Check descriptor validity without accepting a task or calling any codec.
		for _, s := range []*jsondecode.Schema{inputSchema(def), optionsSchema(def)} {
			if _, err := jsondecode.Decode([]byte(`{}`), s, parseLimits); errors.Is(err, jsondecode.ErrConfig) {
				return nil, err
			}
		}
		d.definitions = append(d.definitions, def)
		d.codecs = append(d.codecs, codec)
	}
	return d, nil
}

// Finite local representation envelope; original complete/nested byte caps are
// registered public limits. Depth matches the existing parser's stack ceiling.
var parseLimits = jsondecode.Limits{MaxBytes: 65536, MaxDepth: 128, MaxNodes: 32768}

// Envelope is one parsed task whose closed top-level schema and exact versions
// are valid (REQ-TRUST-0046 steps 3 and 4). Field values, numeric limits and
// test-specific input and options are not yet validated. It holds the parsed
// representation, never source context, and is not an admitted task.
type Envelope struct {
	decoder *Decoder
	value   *jsondecode.Value
	codec   int
}

// Envelope parses the task and validates its top-level structure with field
// values and registered limits deferred to Fields and input and options
// deferred to Parameters, so a caller can apply the public step order. An
// unsupported schema_version or malformed version tuple returns
// jsondecode.ErrVersion; a well-formed unregistered pair returns ErrUnregistered.
// requirement-ref: REQ-JSON-0008
func (d *Decoder) Envelope(input []byte) (*Envelope, error) {
	if d == nil {
		return nil, jsondecode.ErrConfig
	}
	v, err := jsondecode.Decode(input, shapeSchema, parseLimits)
	if err != nil {
		return nil, err
	}
	schema, id, ver := text(v, "schema_version", ""), text(v, "test_id", ""), text(v, "test_version", "")
	if schema != "0.1.0" {
		return nil, jsondecode.ErrVersion
	}
	for i, def := range d.definitions {
		if def.TestID == id && def.TestVersion == ver {
			return &Envelope{decoder: d, value: v, codec: i}, nil
		}
	}
	if testID.MatchString(id) && version.MatchString(ver) && len(ver) <= 64 {
		return nil, unregistered{}
	}
	return nil, jsondecode.ErrVersion
}

// TaskID returns the exact task_id string; Envelope checked only its type.
func (e *Envelope) TaskID() string { return text(e.value, "task_id", "") }

// Test returns the exact registered test ID and version that selected the codec.
func (e *Envelope) Test() task.TestVersion {
	def := e.decoder.definitions[e.codec]
	return task.TestVersion{ID: def.TestID, Version: def.TestVersion}
}

// Fields validates every common field value and registered limit, applies
// defaults and checks expiry against the trusted timing context
// (REQ-TRUST-0046 step 6). Test-specific input and options are checked only
// for their registered byte limits. Failures keep the jsondecode and
// validation classes, including ErrLimit, ErrExpired and ErrUnsupported.
func (e *Envelope) Fields(timing validation.TimingContext) (task.Candidate, error) {
	if err := jsondecode.Validate(e.value, fieldSchema); err != nil {
		return task.Candidate{}, err
	}
	c, err := common(e.value)
	if err != nil {
		return task.Candidate{}, err
	}
	c, err = validation.Common(c)
	if err != nil {
		return task.Candidate{}, err
	}
	if err = validation.ScheduleAt(c.Schedule, timing); err != nil {
		return task.Candidate{}, err
	}
	return c, nil
}

// Parameters validates input and options against the selected test's schemas
// and normalizes them with its codec (the input and options part of step 8).
func (e *Envelope) Parameters() (task.Parameters, error) {
	def := e.decoder.definitions[e.codec]
	input, options := member(e.value, "input"), member(e.value, "options")
	if err := jsondecode.Validate(input, inputSchema(def)); err != nil {
		return nil, err
	}
	if err := jsondecode.Validate(options, optionsSchema(def)); err != nil {
		return nil, err
	}
	p, err := e.decoder.codecs[e.codec].Normalize(input, options)
	if err != nil || p == nil {
		return nil, ErrParameters
	}
	return p, nil
}

// Decode rejects atomically. TimingContext is supplied by a trusted containing
// boundary, never task JSON. The caller separately retains exact received bytes
// for later integrity/replay checks; candidate normalization is not byte identity.
// It runs Envelope, Fields and Parameters in order.
// requirement-ref: REQ-JSON-0010
func (d *Decoder) Decode(input []byte, timing validation.TimingContext) (*task.Candidate, error) {
	e, err := d.Envelope(input)
	if err != nil {
		return nil, err
	}
	c, err := e.Fields(timing)
	if err != nil {
		return nil, err
	}
	if c.Parameters, err = e.Parameters(); err != nil {
		return nil, err
	}
	return &c, nil
}

func member(v *jsondecode.Value, name string) *jsondecode.Value {
	if v == nil {
		return nil
	}
	x, _ := v.Lookup(name)
	return x
}
func text(v *jsondecode.Value, name, fallback string) string {
	if x := member(v, name); x != nil {
		return x.Text()
	}
	return fallback
}
func number(v *jsondecode.Value, name string, fallback int) int {
	if x := member(v, name); x != nil {
		n, _ := strconv.Atoi(x.Text())
		return n
	}
	return fallback
}
func numberPtr(v *jsondecode.Value, name string) *int {
	if member(v, name) == nil {
		return nil
	}
	n := number(v, name, 0)
	return &n
}
func boolean(v *jsondecode.Value, name string) bool {
	if x := member(v, name); x != nil {
		return x.Bool()
	}
	return false
}
func texts(v *jsondecode.Value, name string) []string {
	x := member(v, name)
	if x == nil {
		return nil
	}
	a := make([]string, 0, len(x.Elements()))
	for _, e := range x.Elements() {
		a = append(a, e.Text())
	}
	return a
}

func common(v *jsondecode.Value) (task.Candidate, error) {
	c := task.Candidate{SchemaVersion: text(v, "schema_version", ""), ID: text(v, "task_id", ""), TestID: text(v, "test_id", ""), TestVersion: text(v, "test_version", "")}
	s := member(v, "schedule")
	var err error
	c.Schedule.StartsAt, err = timestamp(member(s, "starts_at"))
	if err != nil {
		return task.Candidate{}, err
	}
	c.Schedule.NotAfter, err = timestamp(member(s, "not_after"))
	if err != nil {
		return task.Candidate{}, err
	}
	c.Schedule.Repeat = task.Repeat{Mode: text(member(s, "repeat"), "mode", "none"), IntervalSeconds: numberPtr(member(s, "repeat"), "interval"), JitterSeconds: numberPtr(member(s, "repeat"), "jitter")}
	c.Schedule.MaxRuns = numberPtr(member(s, "limits"), "max_runs")
	if w := member(s, "time_window"); w != nil {
		// requirement-ref: REQ-SCHEDULE-0013
		c.Schedule.Window = &task.TimeWindow{Timezone: text(w, "timezone", "local")}
		for _, r := range member(w, "ranges").Elements() {
			start, e := minute(text(r, "start", ""))
			if e != nil {
				return task.Candidate{}, e
			}
			end, e := minute(text(r, "end", ""))
			if e != nil {
				return task.Candidate{}, e
			}
			c.Schedule.Window.Ranges = append(c.Schedule.Window.Ranges, task.TimeRange{Start: start, End: end})
		}
	}
	r := member(s, "retry")
	c.Schedule.Retry.MaxAttempts = number(r, "max_attempts", 1)
	if on := texts(r, "retry_on"); on != nil {
		c.Schedule.Retry.On = make([]task.ExecutionReason, 0, len(on))
		for _, reason := range on {
			c.Schedule.Retry.On = append(c.Schedule.Retry.On, task.ExecutionReason(reason))
		}
	}
	if b := member(r, "backoff"); b != nil {
		c.Schedule.Retry.Backoff = &task.Backoff{Mode: text(b, "mode", ""), InitialDelayMS: number(b, "initial_delay_ms", 0), MaxDelayMS: number(b, "max_delay_ms", 0), JitterMS: number(b, "jitter_ms", 10000), Multiplier: numberPtr(b, "multiplier")}
	}
	e := member(v, "execution")
	l := member(e, "limits")
	sent, received, connections := 32768, 131072, 8
	if c.TestID == "NET-01" {
		sent, received, connections = 0, 0, 0
	}
	c.Execution = task.Execution{NetworkPath: text(e, "network_path", "direct"), Limits: task.AttemptLimits{DurationMS: number(l, "max_duration_ms", 30000), BytesSent: number(l, "max_bytes_sent", sent), BytesReceived: number(l, "max_bytes_received", received), Connections: number(l, "max_connections", connections)}}
	n := member(e, "network")
	ctx := member(n, "context")
	c.Execution.Network = task.NetworkRequirements{AllowedTransports: texts(n, "allowed_transports"), VPNPolicy: text(n, "vpn_policy", "any"), RequireValidated: boolean(n, "require_validated"), Context: task.ContextRequirements{RequiredFields: texts(ctx, "required_fields"), AllowedCountryCodes: texts(ctx, "allowed_country_codes")}}
	device := member(e, "device")
	c.Execution.Device = task.DeviceRequirements{MinimumBatteryPercent: numberPtr(device, "minimum_battery_percent"), ChargingOnly: boolean(device, "allow_charging_only")}
	c.Execution.ConcurrencyMode = text(member(e, "concurrency"), "mode", "")
	if x := member(member(e, "concurrency"), "mode"); x != nil && x.Text() == "" {
		return task.Candidate{}, validation.ErrTask
	}
	safety := member(v, "safety")
	c.Safety = task.Safety{RiskLevel: text(safety, "risk_level", ""), RequiresConsent: boolean(safety, "requires_user_consent")}
	if factors := member(safety, "risk_factors"); factors != nil {
		c.Safety.Factors = make([]task.RiskFactor, 0, len(factors.Elements()))
		for _, f := range factors.Elements() {
			c.Safety.Factors = append(c.Safety.Factors, task.RiskFactor{Type: text(f, "type", ""), Level: text(f, "level", "")})
		}
	}
	if labels := member(v, "labels"); labels != nil {
		for _, name := range labels.Names() {
			c.Labels = append(c.Labels, task.Label{Key: name, Value: member(labels, name).Text()})
		}
	}
	return c, nil
}

var utcTimestamp = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}[Tt][0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]+)?([Zz]|[+-]00:00)$`)

func timestamp(v *jsondecode.Value) (*task.Timestamp, error) {
	if v == nil || v.Kind() == jsondecode.Null {
		return nil, nil
	}
	if !utcTimestamp.MatchString(v.Text()) {
		return nil, validation.ErrTask
	}
	value := v.Text()
	fraction := ""
	if dot := strings.IndexByte(value, '.'); dot >= 0 {
		tail := value[dot+1:]
		end := strings.IndexAny(tail, "Zz+-")
		fraction = strings.TrimRight(tail[:end], "0")
		value = value[:dot] + tail[end:]
	}
	t, err := time.Parse(time.RFC3339, strings.ToUpper(value))
	if err != nil {
		return nil, validation.ErrTask
	}
	return &task.Timestamp{Whole: t.UTC(), Fraction: fraction}, nil
}
func minute(s string) (int, error) {
	if len(s) != 5 || s[2] != ':' || s[0] < '0' || s[0] > '9' || s[1] < '0' || s[1] > '9' || s[3] < '0' || s[3] > '9' || s[4] < '0' || s[4] > '9' {
		return 0, validation.ErrTask
	}
	h, _ := strconv.Atoi(s[:2])
	m, _ := strconv.Atoi(s[3:])
	if h > 23 || m > 59 {
		return 0, validation.ErrTask
	}
	return h*60 + m, nil
}
