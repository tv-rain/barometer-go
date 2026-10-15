// Package conformance owns the Go-local versioned fixture runner for tests only.
// Production code must not import it. Synthetic profiles confer no authority.
package conformance

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"regexp"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
)

const MaxFixtureBytes = 1 << 20

var ErrFixture = errors.New("conformance: invalid fixture")
var ErrMismatch = errors.New("conformance: expectation mismatch")

// Fixture has no exported fields: only complete validated fixtures can run.
// Profiles are test-author choices, never runtime trust/policy serialization.
type Fixture struct {
	id              string
	requirements    []string
	task            []byte
	context, policy string
	expected        expectation
}

type expectation struct {
	Rejection  string      `json:"rejection"`
	Normalized *Normalized `json:"normalized,omitempty"`
}

// Normalized is an exact, deliberately limited candidate projection, not a
// Measurement Result or a public model. Every field participates in comparison.
type Normalized struct {
	TaskID        string `json:"task_id"`
	Subject       string `json:"subject"`
	Enabled       bool   `json:"enabled"`
	NetworkPath   string `json:"network_path"`
	DurationMS    int    `json:"duration_ms"`
	BytesSent     int    `json:"bytes_sent"`
	BytesReceived int    `json:"bytes_received"`
	Connections   int    `json:"connections"`
	MaxRuns       int    `json:"max_runs"`
	MaxAttempts   int    `json:"max_attempts"`
}

func (f *Fixture) ID() string             { return f.id }
func (f *Fixture) Requirements() []string { return append([]string(nil), f.requirements...) }

func object(fields map[string]*jsondecode.Schema) *jsondecode.Schema {
	s := &jsondecode.Schema{Kind: jsondecode.Object, Fields: map[string]jsondecode.Field{}}
	for name, child := range fields {
		s.Fields[name] = jsondecode.Field{Schema: child, Required: true}
	}
	return s
}

func fixtureSchema() *jsondecode.Schema {
	text := &jsondecode.Schema{Kind: jsondecode.String}
	integer := &jsondecode.Schema{Kind: jsondecode.Integer, Min: "0", Max: "262144"}
	normalized := object(map[string]*jsondecode.Schema{
		"task_id": text, "subject": text, "enabled": {Kind: jsondecode.Boolean}, "network_path": text,
		"duration_ms": integer, "bytes_sent": integer, "bytes_received": integer, "connections": integer,
		"max_runs": integer, "max_attempts": integer,
	})
	expected := object(map[string]*jsondecode.Schema{"rejection": text})
	expected.Fields["normalized"] = jsondecode.Field{Schema: normalized}
	return object(map[string]*jsondecode.Schema{
		"format_version": {Kind: jsondecode.Integer, Min: "1", Max: "1"},
		"case_id":        text, "adapter": text,
		"requirements": {Kind: jsondecode.Array, Element: text, MaxCount: 64},
		"task_base64":  text, "trusted_context": text, "local_policy": text, "expected": expected,
	})
}

var caseID = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)

// This runner pins its initial requirement allowlist to the active public
// catalog at 05c1364af4f9a604b617fca45f060949b557d265. IDs identify assertions;
// their presence alone never establishes coverage of a complete requirement.
var requirementID = regexp.MustCompile(`^REQ-(TASK-0001|JSON-000[1-9]|JSON-001[01])$`)

func Parse(data []byte) (*Fixture, error) {
	if _, err := jsondecode.Decode(data, fixtureSchema(), jsondecode.Limits{MaxBytes: MaxFixtureBytes, MaxDepth: 16, MaxNodes: 4096}); err != nil {
		return nil, ErrFixture
	}
	var wire struct {
		FormatVersion int         `json:"format_version"`
		ID            string      `json:"case_id"`
		Adapter       string      `json:"adapter"`
		Requirements  []string    `json:"requirements"`
		Task          string      `json:"task_base64"`
		Context       string      `json:"trusted_context"`
		Policy        string      `json:"local_policy"`
		Expected      expectation `json:"expected"`
	}
	if json.Unmarshal(data, &wire) != nil || !caseID.MatchString(wire.ID) || wire.Adapter != "common-task-v1" || wire.Context != "remote-fixed-v1" || wire.Policy != "synthetic-local-v1" || len(wire.Requirements) == 0 {
		return nil, ErrFixture
	}
	seen := map[string]bool{}
	for _, id := range wire.Requirements {
		if !requirementID.MatchString(id) || seen[id] {
			return nil, ErrFixture
		}
		seen[id] = true
	}
	// Strict canonical base64 preserves duplicate names, malformed JSON, invalid
	// UTF-8 and original whitespace; no parse/reserialize step touches task bytes.
	if len(wire.Task) > base64.StdEncoding.EncodedLen(65537) {
		return nil, ErrFixture
	}
	taskBytes, err := base64.StdEncoding.Strict().DecodeString(wire.Task)
	if err != nil || len(taskBytes) > 65537 || base64.StdEncoding.EncodeToString(taskBytes) != wire.Task {
		return nil, ErrFixture
	}
	if wire.Expected.Rejection == "" {
		if wire.Expected.Normalized == nil {
			return nil, ErrFixture
		}
	} else if wire.Expected.Normalized != nil || rejectionError(wire.Expected.Rejection) == nil {
		return nil, ErrFixture
	}
	return &Fixture{id: wire.ID, requirements: append([]string(nil), wire.Requirements...), task: taskBytes, context: wire.Context, policy: wire.Policy, expected: wire.Expected}, nil
}
