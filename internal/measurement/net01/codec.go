// Package net01 implements the Draft 0.1.0 local network-state observation.
// It reads coarse access state through the shared Network Context collector,
// sends no traffic and acquires no egress context.
package net01

import (
	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/taskdecode"
)

// Parameters is the only normalized value of NET-01: there is no target,
// interface, address or device identifier to carry.
type Parameters struct{}

func (Parameters) TaskParameters()  {}
func (Parameters) String() string   { return "net01.Parameters" }
func (Parameters) GoString() string { return "net01.Parameters" }

// Codec accepts only the empty input and options objects, through the common
// strict JSON stages.
type Codec struct{}

// requirement-ref: REQ-NET-01-0010
// requirement-ref: REQ-NET-01-0011
// requirement-ref: REQ-NET-01-0015
// requirement-ref: REQ-NET-01-0016
func (Codec) Definition() taskdecode.Definition {
	return taskdecode.Definition{TestID: "NET-01", TestVersion: "0.1.0",
		Input: &jsondecode.Schema{Kind: jsondecode.Object}, Options: &jsondecode.Schema{Kind: jsondecode.Object}}
}

// requirement-ref: REQ-NET-01-0012
// requirement-ref: REQ-NET-01-0013
// requirement-ref: REQ-NET-01-0014
func (c Codec) Normalize(input, options *jsondecode.Value) (task.Parameters, error) {
	d := c.Definition()
	if input == nil || options == nil || jsondecode.Validate(input, d.Input) != nil || jsondecode.Validate(options, d.Options) != nil {
		return nil, taskdecode.ErrParameters
	}
	return Parameters{}, nil
}
