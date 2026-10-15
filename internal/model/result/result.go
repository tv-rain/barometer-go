// Package result owns detached attempt-result values. These records do not
// establish validity or admission; internal/result owns their wire boundary.
package result

import (
	"errors"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	"github.com/tv-rain/barometer-go/internal/model/task"
)

var ErrEncoding = errors.New("result: use validated result boundary")

type Test struct{ ID, Version, ResultVersion string }
type Implementation struct{ LibraryVersion, Platform string }
type Time struct {
	StartedAt      string
	DurationMS     int
	DurationCapped bool
}
type Execution struct {
	State  string
	Reason task.ExecutionReason
}
type Profile struct {
	ID      string
	Version int64
}
type NetworkValidation struct{ State, Source string }
type Step struct {
	ID          string
	Test        Test
	State       string
	Measurement *Measurement
}
type Measurement struct {
	Outcome                           string
	Observations                      []observation.Observation
	Access                            *observation.AccessSnapshot
	PlatformValidation, ResponseClass string
	Profile                           *Profile
	NetworkValidation                 *NetworkValidation
	Steps                             []Step
	StoppedBeforeStep                 string
}
type Result struct {
	SchemaVersion, TaskID, RunID string
	Attempt                      int
	Test                         Test
	Implementation               Implementation
	Time                         Time
	Execution                    Execution
	NetworkContext               *observation.NetworkContext
	Measurement                  *Measurement
}

// Detached mutable records must not bypass validation or become a second schema.
func (Result) MarshalJSON() ([]byte, error)      { return nil, ErrEncoding }
func (*Result) UnmarshalJSON([]byte) error       { return ErrEncoding }
func (Measurement) MarshalJSON() ([]byte, error) { return nil, ErrEncoding }
func (*Measurement) UnmarshalJSON([]byte) error  { return ErrEncoding }

func (Test) MarshalJSON() ([]byte, error) { return nil, ErrEncoding }
func (*Test) UnmarshalJSON([]byte) error  { return ErrEncoding }

func (Implementation) MarshalJSON() ([]byte, error) { return nil, ErrEncoding }
func (*Implementation) UnmarshalJSON([]byte) error  { return ErrEncoding }

func (Time) MarshalJSON() ([]byte, error) { return nil, ErrEncoding }
func (*Time) UnmarshalJSON([]byte) error  { return ErrEncoding }

func (Execution) MarshalJSON() ([]byte, error) { return nil, ErrEncoding }
func (*Execution) UnmarshalJSON([]byte) error  { return ErrEncoding }

func (Profile) MarshalJSON() ([]byte, error) { return nil, ErrEncoding }
func (*Profile) UnmarshalJSON([]byte) error  { return ErrEncoding }

func (NetworkValidation) MarshalJSON() ([]byte, error) { return nil, ErrEncoding }
func (*NetworkValidation) UnmarshalJSON([]byte) error  { return ErrEncoding }

func (Step) MarshalJSON() ([]byte, error) { return nil, ErrEncoding }
func (*Step) UnmarshalJSON([]byte) error  { return ErrEncoding }
