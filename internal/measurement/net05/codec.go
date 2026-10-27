// Package net05 owns bounded inline profile decoding and sequential execution.
// The codec performs no I/O. Primitive schemas are explicitly composed by the caller.
package net05

import (
	"slices"
	"strconv"
	"strings"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/taskdecode"
)

type Parameters struct {
	id      string
	version int64
	steps   []task.PrimitiveStep
}

func (Parameters) TaskParameters()                        {}
func (Parameters) String() string                         { return "net05.Parameters" }
func (Parameters) GoString() string                       { return "net05.Parameters" }
func (p Parameters) Profile() (string, int64)             { return p.id, p.version }
func (p Parameters) CompositeSteps() []task.PrimitiveStep { return slices.Clone(p.steps) }

// ValidateLimits preserves primitive option bounds against the outer envelope.
// Summed primitive maxima and profile-version history are intentionally deferred;
// see the approved limitations in docs/measurement-requirements.md.
func (p Parameters) ValidateLimits(l task.AttemptLimits) bool {
	for _, s := range p.steps {
		if b, ok := s.Parameters.(interface{ ValidateLimits(task.AttemptLimits) bool }); ok && !b.ValidateLimits(l) {
			return false
		}
	}
	return len(p.steps) >= 1 && len(p.steps) <= 8
}

// Codec keeps immutable primitive codecs, without importing another measurement
// package or the registry. Only explicitly supplied exact versions dispatch.
type Codec struct{ codecs []taskdecode.TestCodec }

func primitive(id string) bool {
	switch id {
	case "DNS-01", "TCP-01", "TLS-01", "HTTP-01":
		return true
	}
	return false
}

func New(codecs []taskdecode.TestCodec) (Codec, error) {
	if _, err := taskdecode.New(codecs); err != nil {
		return Codec{}, err
	}
	var selected []taskdecode.TestCodec
	for _, c := range codecs {
		if primitive(c.Definition().TestID) {
			selected = append(selected, c)
		}
	}
	return Codec{codecs: selected}, nil
}

// requirement-ref: REQ-NET-05-0014
// requirement-ref: REQ-NET-05-0015
// requirement-ref: REQ-NET-05-0016
// requirement-ref: REQ-NET-05-0022
// requirement-ref: REQ-NET-05-0024
// requirement-ref: REQ-NET-05-0025
// requirement-ref: REQ-NET-05-0028
// requirement-ref: REQ-NET-05-0029
func (Codec) Definition() taskdecode.Definition {
	str := &jsondecode.Schema{Kind: jsondecode.String}
	object := func(fields map[string]jsondecode.Field) *jsondecode.Schema {
		return &jsondecode.Schema{Kind: jsondecode.Object, Fields: fields}
	}
	step := object(map[string]jsondecode.Field{
		"step_id": {Required: true, Schema: str}, "test_id": {Required: true, Schema: str}, "test_version": {Required: true, Schema: str},
		"input": {Required: true, Schema: &jsondecode.Schema{Kind: jsondecode.Any, MaxBytes: 49152, Limit: true}}, "options": {Required: true, Schema: &jsondecode.Schema{Kind: jsondecode.Any, MaxBytes: 8192, Limit: true}},
	})
	profile := object(map[string]jsondecode.Field{
		"profile_id":      {Required: true, Schema: str},
		"profile_version": {Required: true, Schema: &jsondecode.Schema{Kind: jsondecode.Integer, Min: "1", Max: "9223372036854775807"}},
		"steps":           {Required: true, Schema: &jsondecode.Schema{Kind: jsondecode.Array, Element: step, MaxCount: 8, Limit: true}},
	})
	profile.MaxBytes, profile.Limit = 49152, true
	return taskdecode.Definition{TestID: "NET-05", TestVersion: "0.1.0", Input: object(map[string]jsondecode.Field{"profile": {Required: true, Schema: profile}}), Options: object(nil)}
}

// identifier retains printable ASCII identifiers without interpolation syntax.
func identifier(s string) bool {
	if len(s) < 1 || len(s) > 64 || strings.ContainsAny(s, "{}$") || strings.Contains(s, "%") {
		return false
	}
	for _, b := range []byte(s) {
		if b < 33 || b > 126 {
			return false
		}
	}
	return true
}

// requirement-ref: REQ-NET-05-0017
// requirement-ref: REQ-NET-05-0018
// requirement-ref: REQ-NET-05-0019
// requirement-ref: REQ-NET-05-0020
// requirement-ref: REQ-NET-05-0021
// requirement-ref: REQ-NET-05-0023
// requirement-ref: REQ-NET-05-0030
func (c Codec) Normalize(input, options *jsondecode.Value) (task.Parameters, error) {
	d := c.Definition()
	if err := jsondecode.Validate(input, d.Input); err != nil {
		return nil, err
	}
	if err := jsondecode.Validate(options, d.Options); err != nil {
		return nil, err
	}
	profile, _ := input.Lookup("profile")
	id, _ := profile.Lookup("profile_id")
	version, _ := profile.Lookup("profile_version")
	steps, _ := profile.Lookup("steps")
	if !identifier(id.Text()) || len(steps.Elements()) == 0 {
		return nil, taskdecode.ErrParameters
	}
	v, err := strconv.ParseInt(version.Text(), 10, 64)
	if err != nil || v < 1 {
		return nil, taskdecode.ErrParameters
	}
	p := Parameters{id: id.Text(), version: v}
	seen := map[string]bool{}
	for _, s := range steps.Elements() {
		sid, _ := s.Lookup("step_id")
		testID, _ := s.Lookup("test_id")
		testVersion, _ := s.Lookup("test_version")
		if !identifier(sid.Text()) || seen[sid.Text()] || !primitive(testID.Text()) {
			return nil, taskdecode.ErrParameters
		}
		seen[sid.Text()] = true
		var codec taskdecode.TestCodec
		for _, candidate := range c.codecs {
			def := candidate.Definition()
			if def.TestID == testID.Text() && def.TestVersion == testVersion.Text() {
				codec = candidate
				break
			}
		}
		if codec == nil {
			return nil, taskdecode.ErrParameters
		}
		in, _ := s.Lookup("input")
		opt, _ := s.Lookup("options")
		def := codec.Definition()
		if err := jsondecode.Validate(in, def.Input); err != nil {
			return nil, err
		}
		if err := jsondecode.Validate(opt, def.Options); err != nil {
			return nil, err
		}
		params, err := codec.Normalize(in, opt)
		if err != nil {
			return nil, err
		}
		if params == nil {
			return nil, taskdecode.ErrParameters
		}
		p.steps = append(p.steps, task.PrimitiveStep{ID: sid.Text(), Test: task.TestVersion{ID: testID.Text(), Version: testVersion.Text()}, Parameters: params})
	}
	return p, nil
}
