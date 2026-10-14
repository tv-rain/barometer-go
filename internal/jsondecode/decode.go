// Package jsondecode owns bounded strict JSON representation and schema decoding.
// It performs no I/O, admission, defaulting, or domain-policy validation.
package jsondecode

import (
	"errors"
	"math/big"
	"sort"
	"strings"
)

// Errors contain only fixed operational classes, never input or parser details.
// requirement-ref: REQ-JSON-0011
var (
	ErrSyntax    = errors.New("jsondecode: syntax")
	ErrText      = errors.New("jsondecode: text")
	ErrDuplicate = errors.New("jsondecode: duplicate")
	ErrField     = errors.New("jsondecode: field")
	ErrType      = errors.New("jsondecode: type")
	ErrRequired  = errors.New("jsondecode: required")
	ErrInteger   = errors.New("jsondecode: integer")
	ErrRange     = errors.New("jsondecode: range")
	ErrVersion   = errors.New("jsondecode: version")
	ErrBounds    = errors.New("jsondecode: bounds")
	ErrConfig    = errors.New("jsondecode: configuration")
)

type Kind uint8

const (
	Null Kind = iota
	Object
	Array
	String
	Boolean
	Integer
)

// Limits are required trusted local policy. MaxDepth counts all value levels,
// including the root; MaxNodes counts values, not object member names.
// MaxDepth cannot exceed 128, bounding recursive stack use independently of input.
type Limits struct{ MaxBytes, MaxDepth, MaxNodes int }

// Schema describes one explicitly selected representation. Objects are closed
// unless Additional is set for an explicitly dynamic-key object. Integer bounds
// are mandatory finite canonical decimal strings, compared with math/big.
// MaxBytes counts original encoded value bytes; MaxCount bounds members/elements.
// Zero optional bounds mean the required enclosing Limits apply.
// Callers must not mutate schema/selection maps during decoding.
type Schema struct {
	Kind               Kind
	Nullable           bool
	Fields             map[string]Field
	Additional         *Schema
	Element            *Schema
	Min, Max           string
	MaxBytes, MaxCount int
}

type Field struct {
	Schema   *Schema
	Required bool
}

// Selection uses an ordered tuple of required string fields, allowing exact
// schema-version or test-id/test-version pairing without concatenation aliases.
// Versionless or trusted-context selection uses Decode with the chosen schema.
type Selection struct {
	Fields []string
	Cases  []Case
}
type Case struct {
	Values []string
	Schema *Schema
}

// Value exposes only a complete validated representation. Missing Lookup values
// differ from explicit Null. Text returns strings or exact integer tokens.
// Collection accessors copy containers so callers cannot change decoder state.
type Value struct {
	kind       Kind
	text       string
	boolean    bool
	members    map[string]*Value
	order      []string
	elements   []*Value
	start, end int
}

func (v *Value) Kind() Kind                        { return v.kind }
func (v *Value) Text() string                      { return v.text }
func (v *Value) Bool() bool                        { return v.boolean }
func (v *Value) Lookup(name string) (*Value, bool) { x, ok := v.members[name]; return x, ok }
func (v *Value) Names() []string                   { return append([]string(nil), v.order...) }
func (v *Value) Elements() []*Value                { return append([]*Value(nil), v.elements...) }
func (v *Value) EncodedBytes() int                 { return v.end - v.start }

// Decode returns nil on every failure. No partially decoded model escapes.
// requirement-ref: REQ-JSON-0010
func Decode(input []byte, schema *Schema, limits Limits) (*Value, error) {
	if !validLimits(limits) || !validSchema(schema) {
		return nil, ErrConfig
	}
	v, err := parse(input, limits)
	if err != nil {
		return nil, err
	}
	if err = validate(v, schema); err != nil {
		return nil, err
	}
	return v, nil
}

// DecodeSelected parses once under the same resource envelope, selects exactly
// one registered tuple, then validates only its schema. No fallback is attempted.
// requirement-ref: REQ-JSON-0008
func DecodeSelected(input []byte, selection Selection, limits Limits) (*Value, error) {
	if !validLimits(limits) || len(selection.Fields) == 0 || len(selection.Fields) > 16 || len(selection.Cases) == 0 || len(selection.Cases) > 256 {
		return nil, ErrConfig
	}
	seen := map[string]bool{}
	for _, name := range selection.Fields {
		if name == "" || seen[name] {
			return nil, ErrConfig
		}
		seen[name] = true
	}
	for i, c := range selection.Cases {
		if len(c.Values) != len(selection.Fields) || !validSchema(c.Schema) || c.Schema.Kind != Object || c.Schema.Nullable {
			return nil, ErrConfig
		}
		for j, name := range selection.Fields {
			f, ok := c.Schema.Fields[name]
			if !ok || !f.Required || f.Schema.Kind != String || f.Schema.Nullable || c.Values[j] == "" {
				return nil, ErrConfig
			}
		}
		for _, previous := range selection.Cases[:i] {
			if sameTuple(c.Values, previous.Values) {
				return nil, ErrConfig
			}
		}
	}
	v, err := parse(input, limits)
	if err != nil {
		return nil, err
	}
	if v.kind != Object {
		return nil, ErrType
	}
	tuple := make([]string, len(selection.Fields))
	for i, name := range selection.Fields {
		x, ok := v.Lookup(name)
		if !ok || x.kind != String {
			return nil, ErrVersion
		}
		tuple[i] = x.text
	}
	for _, c := range selection.Cases {
		if sameTuple(tuple, c.Values) {
			if err = validate(v, c.Schema); err != nil {
				return nil, err
			}
			return v, nil
		}
	}
	return nil, ErrVersion
}

func sameTuple(a, b []string) bool {
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func validLimits(l Limits) bool {
	return l.MaxBytes > 0 && l.MaxDepth > 0 && l.MaxDepth <= 128 && l.MaxNodes > 0
}

// Validate trusted descriptors with bounded work and reject cycles. Numeric
// endpoints have a finite implementation cap of 4096 decimal digits each.
func validSchema(s *Schema) bool {
	nodes := 0
	path := map[*Schema]bool{}
	var visit func(*Schema, int) bool
	visit = func(s *Schema, depth int) bool {
		nodes++
		if s == nil || nodes > 4096 || depth > 128 || path[s] || s.Kind > Integer || s.MaxBytes < 0 || s.MaxCount < 0 {
			return false
		}
		path[s] = true
		defer delete(path, s)
		if s.Kind != Object && (len(s.Fields) > 0 || s.Additional != nil) || s.Kind != Array && s.Element != nil || s.Kind != Integer && (s.Min != "" || s.Max != "") || s.Kind != Array && s.Kind != Object && s.MaxCount != 0 {
			return false
		}
		switch s.Kind {
		case Object:
			if len(s.Fields) > 4096 {
				return false
			}
			for _, f := range s.Fields {
				if !visit(f.Schema, depth+1) {
					return false
				}
			}
			return s.Additional == nil || visit(s.Additional, depth+1)
		case Array:
			return visit(s.Element, depth+1)
		case Integer:
			if len(strings.TrimPrefix(s.Min, "-")) > 4096 || len(strings.TrimPrefix(s.Max, "-")) > 4096 || !integerToken(s.Min) || !integerToken(s.Max) {
				return false
			}
			lo, _ := new(big.Int).SetString(s.Min, 10)
			hi, _ := new(big.Int).SetString(s.Max, 10)
			return lo.Cmp(hi) <= 0
		}
		return true
	}
	return visit(s, 1)
}

// requirement-ref: REQ-JSON-0004
// requirement-ref: REQ-JSON-0005
func validate(v *Value, s *Schema) error {
	if s.MaxBytes > 0 && v.EncodedBytes() > s.MaxBytes {
		return ErrBounds
	}
	if v.kind == Null && s.Nullable {
		return nil
	}
	if v.kind != s.Kind {
		return ErrType
	}
	switch s.Kind {
	case Object:
		if s.MaxCount > 0 && len(v.members) > s.MaxCount {
			return ErrBounds
		}
		// Input order and sorted descriptor names make failure precedence stable.
		for _, name := range v.order {
			f, ok := s.Fields[name]
			child := f.Schema
			if !ok {
				child = s.Additional
			}
			if child == nil {
				return ErrField
			}
			if err := validate(v.members[name], child); err != nil {
				return err
			}
		}
		names := make([]string, 0, len(s.Fields))
		for name := range s.Fields {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			if s.Fields[name].Required {
				if _, ok := v.members[name]; !ok {
					return ErrRequired
				}
			}
		}
	case Array:
		if s.MaxCount > 0 && len(v.elements) > s.MaxCount {
			return ErrBounds
		}
		for _, x := range v.elements {
			if err := validate(x, s.Element); err != nil {
				return err
			}
		}
	case Integer:
		// requirement-ref: REQ-JSON-0006
		if !integerToken(v.text) {
			return ErrInteger
		}
		// requirement-ref: REQ-JSON-0007
		x, _ := new(big.Int).SetString(v.text, 10)
		lo, _ := new(big.Int).SetString(s.Min, 10)
		hi, _ := new(big.Int).SetString(s.Max, 10)
		if x.Cmp(lo) < 0 || x.Cmp(hi) > 0 {
			return ErrRange
		}
	}
	return nil
}

func integerToken(s string) bool {
	if s == "0" {
		return true
	}
	i := 0
	if len(s) > 0 && s[0] == '-' {
		i++
	}
	if i >= len(s) || s[i] < '1' || s[i] > '9' {
		return false
	}
	for i++; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
