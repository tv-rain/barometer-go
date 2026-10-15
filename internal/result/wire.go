package result

import (
	"encoding/json"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
)

type testWire struct {
	ID            string `json:"id"`
	Version       string `json:"version"`
	ResultVersion string `json:"result_version"`
}

func testWireOf(t model.Test) testWire { return testWire{t.ID, t.Version, t.ResultVersion} }

type implementationWire struct {
	LibraryVersion string `json:"library_version"`
	Platform       string `json:"platform"`
}
type timeWire struct {
	StartedAt      string `json:"started_at"`
	Precision      string `json:"precision"`
	DurationMS     int    `json:"duration_ms"`
	DurationCapped *bool  `json:"duration_capped,omitempty"`
}
type executionWire struct {
	State  string               `json:"state"`
	Reason task.ExecutionReason `json:"reason,omitempty"`
}
type resultWire struct {
	SchemaVersion  string             `json:"schema_version"`
	TaskID         string             `json:"task_id"`
	RunID          string             `json:"run_id"`
	Attempt        int                `json:"attempt"`
	Test           testWire           `json:"test"`
	Implementation implementationWire `json:"implementation"`
	Time           timeWire           `json:"time"`
	Execution      executionWire      `json:"execution"`
	NetworkContext json.RawMessage    `json:"network_context,omitempty"`
	Measurement    json.RawMessage    `json:"measurement,omitempty"`
}
type profileWire struct {
	ID      string `json:"profile_id"`
	Version int64  `json:"profile_version"`
}
type validationWire struct {
	State  string `json:"state"`
	Source string `json:"source"`
}
type stepWire struct {
	ID          string          `json:"step_id"`
	Test        testWire        `json:"test"`
	State       string          `json:"state"`
	Measurement json.RawMessage `json:"measurement,omitempty"`
}
type measurementWire struct {
	Outcome            string            `json:"outcome"`
	Observations       []json.RawMessage `json:"observations"`
	Access             json.RawMessage   `json:"access,omitempty"`
	PlatformValidation string            `json:"platform_validation,omitempty"`
	ResponseClass      string            `json:"response_class,omitempty"`
	Profile            *profileWire      `json:"profile,omitempty"`
	NetworkValidation  *validationWire   `json:"network_validation,omitempty"`
	Steps              *[]stepWire       `json:"steps,omitempty"`
	StoppedBeforeStep  string            `json:"stopped_before_step,omitempty"`
}

func encode(v model.Result) ([]byte, error) {
	w := resultWire{SchemaVersion: v.SchemaVersion, TaskID: v.TaskID, RunID: v.RunID, Attempt: v.Attempt, Test: testWireOf(v.Test), Implementation: implementationWire{v.Implementation.LibraryVersion, v.Implementation.Platform},
		Time: timeWire{StartedAt: v.Time.StartedAt, Precision: "minute", DurationMS: v.Time.DurationMS}, Execution: executionWire{v.Execution.State, v.Execution.Reason}}
	if v.Time.DurationCapped {
		flag := true
		w.Time.DurationCapped = &flag
	}
	var err error
	if v.NetworkContext != nil {
		w.NetworkContext, err = json.Marshal(*v.NetworkContext)
		if err != nil {
			return nil, err
		}
	}
	if v.Measurement != nil {
		w.Measurement, err = encodeMeasurement(v.Measurement)
		if err != nil {
			return nil, err
		}
	}
	return json.Marshal(w)
}
func encodeMeasurement(m *model.Measurement) ([]byte, error) {
	w := measurementWire{Outcome: m.Outcome, Observations: make([]json.RawMessage, 0, len(m.Observations)), PlatformValidation: m.PlatformValidation, ResponseClass: m.ResponseClass, StoppedBeforeStep: m.StoppedBeforeStep}
	var err error
	for _, o := range m.Observations {
		b, e := json.Marshal(o)
		if e != nil {
			return nil, e
		}
		w.Observations = append(w.Observations, b)
	}
	if m.Access != nil {
		w.Access, err = json.Marshal(*m.Access)
		if err != nil {
			return nil, err
		}
	}
	if m.Profile != nil {
		w.Profile = &profileWire{m.Profile.ID, m.Profile.Version}
	}
	if m.NetworkValidation != nil {
		w.NetworkValidation = &validationWire{m.NetworkValidation.State, m.NetworkValidation.Source}
	}
	if m.Steps != nil {
		steps := make([]stepWire, 0, len(m.Steps))
		for _, s := range m.Steps {
			sw := stepWire{ID: s.ID, Test: testWireOf(s.Test), State: s.State}
			if s.Measurement != nil {
				sw.Measurement, err = encodeMeasurement(s.Measurement)
				if err != nil {
					return nil, err
				}
			}
			steps = append(steps, sw)
		}
		w.Steps = &steps
	}
	return json.Marshal(w)
}
func preflight(v model.Result) bool {
	for _, s := range []string{v.SchemaVersion, v.TaskID, v.RunID, v.Test.ID, v.Test.Version, v.Test.ResultVersion, v.Implementation.LibraryVersion, v.Implementation.Platform, v.Time.StartedAt, v.Execution.State, string(v.Execution.Reason)} {
		if len(s) > 128 {
			return false
		}
	}
	return boundedMeasurement(v.Measurement, 0)
}
func boundedMeasurement(m *model.Measurement, depth int) bool {
	if m == nil {
		return true
	}
	if depth > 1 || len(m.Observations) > 16 || len(m.Steps) > 8 || len(m.Outcome) > 64 || len(m.PlatformValidation) > 32 || len(m.ResponseClass) > 32 || len(m.StoppedBeforeStep) > 64 {
		return false
	}
	if m.Profile != nil && len(m.Profile.ID) > 64 || m.NetworkValidation != nil && (len(m.NetworkValidation.State) > 32 || len(m.NetworkValidation.Source) > 32) {
		return false
	}
	for _, s := range m.Steps {
		if len(s.ID) > 64 || len(s.Test.ID) > 64 || len(s.Test.Version) > 64 || len(s.Test.ResultVersion) > 64 || len(s.State) > 32 || !boundedMeasurement(s.Measurement, depth+1) {
			return false
		}
	}
	return true
}

// decodeObservations preserves order and delegates exact shared schemas and
// normalized detail invariants to their existing owner.
func decodeObservations(raw []json.RawMessage, o Output, duration int) ([]observation.Observation, error) {
	observations := make([]observation.Observation, 0, len(raw))
	for _, b := range raw {
		v, err := observation.DecodeProtocol(Contract, b, o.Limits)
		if err != nil {
			return nil, err
		}
		if v.DurationMS() > duration {
			return nil, ErrValue
		}
		observations = append(observations, v)
	}
	return observations, nil
}
