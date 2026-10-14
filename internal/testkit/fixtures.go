// Package testkit provides bounded synthetic fixtures and deterministic controls
// for tests. Production packages must not import it. It implements no runtime
// schema, admission policy, measurement, or platform fallback.
package testkit

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
)

// MaxFixtureBytes bounds each fixture allocation. This is a test tooling limit,
// not a runtime limit or permission to accept larger public models.
const MaxFixtureBytes = 1 << 20

var ErrFixtureBounds = errors.New("testkit: fixture bounds exceeded")

// Member encodes a name and preserves the caller's raw value, including invalid
// JSON. Use only synthetic values. Object deliberately preserves duplicate names.
func Member(name, rawValue string) (string, error) {
	// Bound JSON name expansion before encoding or concatenating caller data.
	if len(name) > 4096 || len(rawValue) > MaxFixtureBytes-6*len(name)-3 {
		return "", ErrFixtureBounds
	}
	key, _ := json.Marshal(name)
	return string(key) + ":" + rawValue, nil
}

// Object joins raw members without parsing or normalizing them. Raw members also
// allow escaped duplicate names that a map-based builder would erase.
func Object(members ...string) ([]byte, error) {
	size := 2
	for i, member := range members {
		if i > 0 {
			size++
		}
		if len(member) > MaxFixtureBytes-size {
			return nil, ErrFixtureBounds
		}
		size += len(member)
	}
	return []byte("{" + strings.Join(members, ",") + "}"), nil
}

// Bytes builds an exact encoded-byte fixture without allocating above the cap.
func Bytes(size int, fill byte) ([]byte, error) {
	if size < 0 || size > MaxFixtureBytes {
		return nil, ErrFixtureBounds
	}
	return bytes.Repeat([]byte{fill}, size), nil
}

// PaddedJSON preserves a synthetic value and adds trailing JSON whitespace to an
// exact encoded size. It does not validate the value. This isolates byte limits
// from syntax failures and preserves the original encoded bytes for the decoder.
func PaddedJSON(value []byte, size int) ([]byte, error) {
	if size < len(value) || size > MaxFixtureBytes {
		return nil, ErrFixtureBounds
	}
	fixture := bytes.Repeat([]byte{' '}, size)
	copy(fixture, value)
	return fixture, nil
}

// ByteBoundary holds independent below/at/above fixtures for an inclusive maximum.
type ByteBoundary struct {
	Below, At, Above []byte
}

func ByteBoundaries(maximum int, fill byte) (ByteBoundary, error) {
	if maximum < 1 || maximum >= MaxFixtureBytes {
		return ByteBoundary{}, ErrFixtureBounds
	}
	below, _ := Bytes(maximum-1, fill)
	at, _ := Bytes(maximum, fill)
	above, _ := Bytes(maximum+1, fill)
	return ByteBoundary{below, at, above}, nil
}

// NestedArray constructs hostile depth fixtures with bounded work and storage.
// The cap is test tooling policy; it does not invent a runtime nesting limit.
func NestedArray(depth int) ([]byte, error) {
	if depth < 0 || depth > 1024 {
		return nil, ErrFixtureBounds
	}
	return []byte(strings.Repeat("[", depth) + "0" + strings.Repeat("]", depth)), nil
}

// RejectionCase uses trusted, synthetic metadata. Requirement is an allocated
// public ID or an immutable public section reference for an unallocated rule.
// Expected is a consumer-defined bounded error class, never a raw parser message.
type RejectionCase struct {
	Name, Requirement, Expected string
	Input                       []byte
}

// InvalidJSON returns fresh fixtures for schema-independent text failures. It is
// a starter corpus, not exhaustive conformance evidence. Object-type, version,
// unknown-field, duplicate-name, and range cases belong to the selected schema.
func InvalidJSON() []RejectionCase {
	// requirement-ref: REQ-JSON-0001
	// requirement-ref: REQ-JSON-0002
	return []RejectionCase{
		{"empty", "REQ-JSON-0001", "syntax", nil},
		{"trailing-value", "REQ-JSON-0001", "syntax", []byte("{} {}")},
		{"trailing-content", "REQ-JSON-0001", "syntax", []byte("{}x")},
		{"comment", "REQ-JSON-0001", "syntax", []byte("{/* synthetic */}")},
		{"trailing-comma", "REQ-JSON-0001", "syntax", []byte("{\"x\":0,}")},
		{"byte-order-mark", "REQ-JSON-0001", "syntax", []byte("\xef\xbb\xbf{}")},
		{"truncated", "REQ-JSON-0001", "syntax", []byte("{\"x\":")},
		{"invalid-utf8", "REQ-JSON-0002", "text", []byte{'"', 0xff, '"'}},
		{"malformed-escape", "REQ-JSON-0002", "text", []byte(`"\uZZZZ"`)},
		{"unpaired-high-surrogate", "REQ-JSON-0002", "text", []byte(`"\ud800"`)},
		{"unpaired-low-surrogate", "REQ-JSON-0002", "text", []byte(`"\udc00"`)},
	}
}
