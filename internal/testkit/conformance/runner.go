package conformance

import (
	"errors"
	"strings"
	"time"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/taskdecode"
	"github.com/tv-rain/barometer-go/internal/validation"
)

type parameters struct {
	subject string
	enabled bool
}

func (parameters) TaskParameters() {}

type codec struct{}

func (codec) Definition() taskdecode.Definition {
	return taskdecode.Definition{TestID: "SYN-01", TestVersion: "0.1.0",
		Input:   object(map[string]*jsondecode.Schema{"subject": {Kind: jsondecode.String}}),
		Options: &jsondecode.Schema{Kind: jsondecode.Object, Fields: map[string]jsondecode.Field{"enabled": {Schema: &jsondecode.Schema{Kind: jsondecode.Boolean}}}},
	}
}
func (codec) Normalize(input, options *jsondecode.Value) (task.Parameters, error) {
	s, _ := input.Lookup("subject")
	p := parameters{subject: strings.ToLower(s.Text()), enabled: true}
	if v, ok := options.Lookup("enabled"); ok {
		p.enabled = v.Bool()
	}
	return p, nil
}

func rejectionError(class string) error {
	switch class {
	case "syntax":
		return jsondecode.ErrSyntax
	case "text":
		return jsondecode.ErrText
	case "duplicate":
		return jsondecode.ErrDuplicate
	case "field":
		return jsondecode.ErrField
	case "type":
		return jsondecode.ErrType
	case "required":
		return jsondecode.ErrRequired
	case "integer":
		return jsondecode.ErrInteger
	case "range":
		return jsondecode.ErrRange
	case "version":
		return jsondecode.ErrVersion
	case "bounds":
		return jsondecode.ErrBounds
	case "task":
		return validation.ErrTask
	case "timing":
		return validation.ErrTiming
	case "parameters":
		return taskdecode.ErrParameters
	}
	return nil
}

// environment constructs synthetic immutable models separately from task bytes.
// This is an isolated test adapter, never an import path for runtime trust.
func environment() (task.SourceContext, task.LocalPolicy, validation.TimingContext, error) {
	at := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	issued := task.TimestampFromTime(at)
	pending, err := task.NewSourceContext(task.SourceInput{Origin: task.Remote, PolicyGeneration: 1})
	if err != nil {
		return pending, task.LocalPolicy{}, validation.TimingContext{}, ErrFixture
	}
	source, err := pending.WithVerifiedPackage("synthetic-issuer", issued)
	if err != nil {
		return source, task.LocalPolicy{}, validation.TimingContext{}, ErrFixture
	}
	p := task.Permissions{Tests: []task.TestVersion{{ID: "SYN-01", Version: "0.1.0"}}, Paths: task.Paths{Direct: true}, MaxRisk: task.LowRisk,
		Limits: task.PolicyLimits{Attempt: task.AttemptLimits{DurationMS: 30000, BytesSent: 32768, BytesReceived: 131072, Connections: 8}, MaxRuns: 1, MaxAttempts: 1, MaxValiditySeconds: 86400,
			MinRepeatIntervalSeconds: 21600, MinRepeatJitterSeconds: 300, MinRetryDelayMS: 30000, MaxRetryDelayMS: 30000, MaxRetryJitterMS: 10000,
			MaxResultBytes: 65536, MaxBundleBytes: 262144, MaxResultsPerBundle: 16, MaxPendingBundles: 32, MaxPendingBytes: 4194304, MaxRetentionSeconds: 86400}}
	policy, err := task.NewLocalPolicy(task.PolicyInput{Generation: 1, Permissions: p, Issuers: []task.IssuerPermissions{{IssuerID: "synthetic-issuer", Permissions: p}}})
	if err != nil {
		return source, policy, validation.TimingContext{}, ErrFixture
	}
	return source, policy, validation.TimingContext{Origin: "remote", AcceptedAt: at, IssuedAt: &at}, nil
}

// Run checks the declared candidate projection or exact rejection class. It
// returns fixed diagnostics without task bytes, targets or consumer error text.
// Policy is constructed/checked as context, not enforced as effective authority.
func Run(f *Fixture) error {
	if f == nil || f.context != "remote-fixed-v1" || f.policy != "synthetic-local-v1" {
		return ErrFixture
	}
	source, policy, timing, err := environment()
	if err != nil {
		return ErrFixture
	}
	if _, ok := policy.Issuer(source.IssuerID()); !ok || policy.Generation() != source.PolicyGeneration() {
		return ErrFixture
	}
	decoder, err := taskdecode.New([]taskdecode.TestCodec{codec{}})
	if err != nil {
		return ErrFixture
	}
	candidate, err := decoder.Decode(append([]byte(nil), f.task...), timing)
	if f.expected.Rejection != "" {
		if candidate != nil || !errors.Is(err, rejectionError(f.expected.Rejection)) {
			return ErrMismatch
		}
		return nil
	}
	if err != nil || candidate == nil {
		return ErrMismatch
	}
	p, ok := candidate.Parameters.(parameters)
	if !ok || candidate.Schedule.MaxRuns == nil {
		return ErrMismatch
	}
	limits := candidate.Execution.Limits
	got := Normalized{TaskID: candidate.ID, Subject: p.subject, Enabled: p.enabled, NetworkPath: candidate.Execution.NetworkPath,
		DurationMS: limits.DurationMS, BytesSent: limits.BytesSent, BytesReceived: limits.BytesReceived, Connections: limits.Connections,
		MaxRuns: *candidate.Schedule.MaxRuns, MaxAttempts: candidate.Schedule.Retry.MaxAttempts}
	if f.expected.Normalized == nil || got != *f.expected.Normalized {
		return ErrMismatch
	}
	return nil
}
