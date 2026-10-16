package jsondecode

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/tv-rain/barometer-go/internal/testkit"
)

var testLimits = Limits{MaxBytes: 65536, MaxDepth: 32, MaxNodes: 4096}

func integer(min, max string) *Schema { return &Schema{Kind: Integer, Min: min, Max: max} }
func modelSchema() *Schema {
	return &Schema{Kind: Object, Fields: map[string]Field{
		"version":   {&Schema{Kind: String}, true},
		"test_id":   {&Schema{Kind: String}, true},
		"count":     {integer("0", "18446744073709551615"), true},
		"starts_at": {&Schema{Kind: String, Nullable: true}, false},
		"labels":    {&Schema{Kind: Object, Additional: &Schema{Kind: String}}, false},
		"items":     {&Schema{Kind: Array, Element: &Schema{Kind: Object, Fields: map[string]Field{"name": {&Schema{Kind: String}, true}}}}, false},
	}}
}
func selected(s *Schema) Selection {
	return Selection{Fields: []string{"test_id", "version"}, Cases: []Case{{[]string{"SYNTHETIC", "0.1.0"}, s}}}
}

const base = `"version":"0.1.0","test_id":"SYNTHETIC","count":0`

// requirement-ref: REQ-JSON-0001
// requirement-ref: REQ-JSON-0002
// requirement-ref: REQ-JSON-0003
// requirement-ref: REQ-JSON-0004
// requirement-ref: REQ-JSON-0005
// requirement-ref: REQ-JSON-0006
// requirement-ref: REQ-JSON-0007
// requirement-ref: REQ-JSON-0008
func TestRejectionCorpus(t *testing.T) {
	cases := testkit.InvalidJSON()
	add := func(name, requirement, class, input string) {
		cases = append(cases, testkit.RejectionCase{Name: name, Requirement: requirement, Expected: class, Input: []byte(input)})
	}
	add("top-array", "REQ-JSON-0001", "type", `[]`)
	add("top-null", "REQ-JSON-0005", "type", `null`)
	add("top-scalar", "REQ-JSON-0001", "type", `true`)
	add("duplicate-equal", "REQ-JSON-0003", "duplicate", `{`+base+`,"count":0}`)
	add("duplicate-conflict", "REQ-JSON-0003", "duplicate", `{`+base+`,"count":1}`)
	add("duplicate-escaped", "REQ-JSON-0003", "duplicate", `{`+base+`,"co\u0075nt":0}`)
	add("duplicate-label", "REQ-JSON-0003", "duplicate", `{`+base+`,"labels":{"x":"a","x":"a"}}`)
	add("duplicate-array-object", "REQ-JSON-0003", "duplicate", `{`+base+`,"items":[{"name":"a","n\u0061me":"b"}]}`)
	add("unknown", "REQ-JSON-0004", "field", `{`+base+`,"secret-name":"secret-value"}`)
	add("case", "REQ-JSON-0004", "field", `{"Version":"0.1.0","test_id":"SYNTHETIC","count":0}`)
	add("unknown-nested", "REQ-JSON-0004", "field", `{`+base+`,"items":[{"name":"a","extra":0}]}`)
	add("required", "REQ-JSON-0005", "required", `{"version":"0.1.0","test_id":"SYNTHETIC"}`)
	add("null-integer", "REQ-JSON-0005", "type", `{"version":"0.1.0","test_id":"SYNTHETIC","count":null}`)
	for _, raw := range []string{`"0"`, `true`, `[]`, `{}`} {
		add("wrong-type-"+raw, "REQ-JSON-0005", "type", `{"version":"0.1.0","test_id":"SYNTHETIC","count":`+raw+`}`)
	}
	for _, raw := range []string{"-0", "1.0", "1e0", "1E+0", "1.5"} {
		add("integer-"+raw, "REQ-JSON-0006", "integer", `{"version":"0.1.0","test_id":"SYNTHETIC","count":`+raw+`}`)
	}
	for _, raw := range []string{"+1", "01", "1.", "1e", "--1"} {
		add("syntax-"+raw, "REQ-JSON-0001", "syntax", `{"version":"0.1.0","test_id":"SYNTHETIC","count":`+raw+`}`)
	}
	for _, raw := range []string{"-1", "18446744073709551616"} {
		add("range-"+raw, "REQ-JSON-0007", "range", `{"version":"0.1.0","test_id":"SYNTHETIC","count":`+raw+`}`)
	}
	add("surrogate-name", "REQ-JSON-0002", "text", `{`+base+`,"\ud800":"a"}`)
	add("surrogate-nested", "REQ-JSON-0002", "text", `{`+base+`,"labels":{"x":"\udc00"}}`)
	// Regression cases for already rejected input; they assert, not change, behavior.
	for _, raw := range []string{"\x00", "\x01", "\x1f", "\t", "\n", "\r"} {
		add("raw-control-value-"+strconv.Quote(raw), "REQ-JSON-0002", "text", `{`+base+`,"labels":{"x":"a`+raw+`b"}}`)
	}
	add("raw-control-name", "REQ-JSON-0002", "text", `{`+base+`,"labels":{"a`+"\x1f"+`b":"x"}}`)
	for name, raw := range map[string]string{"slash-2": "\xc0\xaf", "nul-2": "\xc0\x80", "ascii-2": "\xc1\xbf", "slash-3": "\xe0\x80\xaf", "slash-4": "\xf0\x80\x80\xaf"} {
		add("overlong-"+name, "REQ-JSON-0002", "text", `{`+base+`,"labels":{"x":"`+raw+`"}}`)
	}
	for name, raw := range map[string]string{"high": "\xed\xa0\x80", "low": "\xed\xb0\x80", "pair": "\xed\xa0\xbd\xed\xb8\x80"} {
		add("cesu-surrogate-"+name, "REQ-JSON-0002", "text", `{`+base+`,"labels":{"x":"`+raw+`"}}`)
	}
	add("cesu-surrogate-name", "REQ-JSON-0002", "text", `{`+base+`,"labels":{"`+"\xed\xa0\x80"+`":"x"}}`)
	for _, raw := range []string{"NaN", "-NaN", "Infinity", "-Infinity", "+Infinity", "inf", "nan"} {
		add("non-finite-"+raw, "REQ-JSON-0001", "syntax", `{"version":"0.1.0","test_id":"SYNTHETIC","count":`+raw+`}`)
	}
	classes := map[string]error{"syntax": ErrSyntax, "text": ErrText, "type": ErrType, "duplicate": ErrDuplicate, "field": ErrField, "required": ErrRequired, "integer": ErrInteger, "range": ErrRange}
	testkit.RunRejections(t, cases, func(input []byte) error {
		v, err := Decode(input, modelSchema(), testLimits)
		if v != nil {
			t.Error("rejection exposed a candidate")
		}
		return err
	}, func(class string, err error) bool { return errors.Is(err, classes[class]) })
	for _, input := range []string{`{}`, `{"version":null,"test_id":"SYNTHETIC"}`, `{"version":1,"test_id":"SYNTHETIC"}`, `{"version":"","test_id":"SYNTHETIC"}`, `{"version":"9.0","test_id":"SYNTHETIC"}`, `{"version":"0.1.0","test_id":"OTHER"}`} {
		v, err := DecodeSelected([]byte(input), selected(modelSchema()), testLimits)
		if v != nil || !errors.Is(err, ErrVersion) {
			t.Errorf("version corpus: %v", err)
		}
	}
}

func TestPositiveAndPresence(t *testing.T) {
	for _, input := range []string{
		`{` + base + `}`, " \t\r\n{" + base + "} \n",
		`{"version":"0.1.0","test_id":"SYNTHETIC","co\u0075nt":9007199254740993}`,
		`{"version":"0.1.0","test_id":"SYNTHETIC","count":18446744073709551615}`,
		`{` + base + `,"starts_at":null,"labels":{"x":"\ud83d\ude00","replacement":"�","é":"a","e\u0301":"b"},"items":[{"name":"a"},{"name":"b"}]}`,
		`{` + base + `,"starts_at":"","labels":{},"items":[]}`,
	} {
		v, err := DecodeSelected([]byte(input), selected(modelSchema()), testLimits)
		if err != nil || v == nil {
			t.Fatalf("positive control: %v", err)
		}
	}
	v, _ := Decode([]byte(`{`+base+`}`), modelSchema(), testLimits)
	if _, ok := v.Lookup("starts_at"); ok {
		t.Fatal("missing lost")
	}
	count, ok := v.Lookup("count")
	if !ok || count.Text() != "0" {
		t.Fatal("zero lost")
	}
	v, _ = Decode([]byte(`{`+base+`,"starts_at":null}`), modelSchema(), testLimits)
	x, ok := v.Lookup("starts_at")
	if !ok || x.Kind() != Null {
		t.Fatal("null lost")
	}
	v, _ = Decode([]byte(`{`+base+`,"starts_at":""}`), modelSchema(), testLimits)
	x, ok = v.Lookup("starts_at")
	if !ok || x.Kind() != String || x.Text() != "" {
		t.Fatal("empty lost")
	}
	for _, raw := range []string{"0", "1", "-1", "-9223372036854775808", "9223372036854775807", "9007199254740993"} {
		v, err := Decode([]byte(raw), integer("-9223372036854775808", "9223372036854775807"), testLimits)
		if err != nil || v.Text() != raw {
			t.Fatalf("exact integer: %v", err)
		}
	}
	for _, raw := range []string{"-9223372036854775809", "9223372036854775808"} {
		if _, err := Decode([]byte(raw), integer("-9223372036854775808", "9223372036854775807"), testLimits); !errors.Is(err, ErrRange) {
			t.Fatal(err)
		}
	}
	// Finite domains are not restricted to machine-width integers.
	raw := strings.Repeat("9", 100)
	v, err := Decode([]byte(raw), integer("0", raw), testLimits)
	if err != nil || v.Text() != raw {
		t.Fatal(err)
	}
}

// requirement-ref: REQ-JSON-0009
// requirement-ref: REQ-JSON-0010
func TestBounds(t *testing.T) {
	limits := Limits{MaxBytes: 64, MaxDepth: 4, MaxNodes: 8}
	for _, size := range []int{63, 64, 65} {
		input, _ := testkit.PaddedJSON([]byte(`{}`), size)
		v, err := Decode(input, &Schema{Kind: Object}, limits)
		if size <= 64 && (v == nil || err != nil) || size > 64 && (v != nil || !errors.Is(err, ErrBounds)) {
			t.Fatalf("bytes %d: %v", size, err)
		}
	}
	// Bounds win over syntax and unsupported schema selection for oversized input.
	if _, err := DecodeSelected([]byte(strings.Repeat("x", 65)), selected(modelSchema()), limits); !errors.Is(err, ErrBounds) {
		t.Fatal(err)
	}
	s := &Schema{Kind: Object, Fields: map[string]Field{"x": {&Schema{Kind: Object, MaxBytes: 4}, true}}}
	for _, input := range []string{`{"x":{}}`, `{"x":{  }}`} {
		if _, err := Decode([]byte(input), s, limits); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Decode([]byte(`{"x":{   }}`), s, limits); !errors.Is(err, ErrBounds) {
		t.Fatal(err)
	}
	for _, kind := range []Kind{Array, Object} {
		var schema *Schema
		var inputs []string
		if kind == Array {
			schema = &Schema{Kind: Array, Element: &Schema{Kind: Null}, MaxCount: 2}
			inputs = []string{`[null]`, `[null,null]`, `[null,null,null]`}
		} else {
			schema = &Schema{Kind: Object, Additional: &Schema{Kind: Null}, MaxCount: 2}
			inputs = []string{`{"a":null}`, `{"a":null,"b":null}`, `{"a":null,"b":null,"c":null}`}
		}
		for i, input := range inputs {
			_, err := Decode([]byte(input), schema, limits)
			if i < 2 && err != nil || i == 2 && !errors.Is(err, ErrBounds) {
				t.Fatal(err)
			}
		}
	}
	for _, depth := range []int{3, 4, 5} {
		schema := &Schema{Kind: Null}
		for i := 1; i < depth; i++ {
			schema = &Schema{Kind: Array, Element: schema}
		}
		input := []byte(strings.Repeat("[", depth-1) + "null" + strings.Repeat("]", depth-1))
		_, err := Decode(input, schema, limits)
		if depth <= 4 && err != nil || depth > 4 && !errors.Is(err, ErrBounds) {
			t.Fatal(err)
		}
	}
	schema := &Schema{Kind: Array, Element: &Schema{Kind: Null}}
	for _, nodes := range []int{7, 8, 9} {
		input := []byte("[" + strings.TrimSuffix(strings.Repeat("null,", nodes-1), ",") + "]")
		_, err := Decode(input, schema, limits)
		if nodes <= 8 && err != nil || nodes > 8 && !errors.Is(err, ErrBounds) {
			t.Fatal(err)
		}
	}
	deep, _ := testkit.NestedArray(1024)
	if _, err := Decode(deep, schema, testLimits); !errors.Is(err, ErrBounds) {
		t.Fatal(err)
	}
	if _, err := Decode([]byte(strings.Repeat("[", 128)), schema, Limits{1024, 128, 1024}); !errors.Is(err, ErrBounds) {
		t.Fatal(err)
	}
}

// requirement-ref: REQ-JSON-0011
func TestErrorsAndDeterminism(t *testing.T) {
	for _, input := range []string{`{` + base + `,"synthetic-password":"SYNTHETIC-TOKEN"}`, `{` + base + `,"labels":{"synthetic-password":"\ud800"}}`, `{` + base + `,"synthetic-password":1,"synthetic-password":2}`} {
		var previous error
		for range 50 {
			v, err := Decode([]byte(input), modelSchema(), testLimits)
			if v != nil || err == nil || len(err.Error()) > 32 || strings.Contains(err.Error(), "synthetic-password") || strings.Contains(err.Error(), "SYNTHETIC-TOKEN") {
				t.Fatal("unbounded or leaking error")
			}
			if previous != nil && err != previous {
				t.Fatal("unstable failure")
			}
			previous = err
		}
	}
	// Required-field checking must not depend on Go map iteration.
	for range 50 {
		if _, err := Decode([]byte(`{}`), modelSchema(), testLimits); err != ErrRequired {
			t.Fatal(err)
		}
	}
}

func TestConfigurationAndNoFallback(t *testing.T) {
	cycle := &Schema{Kind: Array}
	cycle.Element = cycle
	for _, s := range []*Schema{nil, cycle, {Kind: Integer}, {Kind: Integer, Min: "-0", Max: "1"}, {Kind: Integer, Min: "2", Max: "1"}, {Kind: Array}, {Kind: String, Fields: map[string]Field{"x": {&Schema{Kind: Null}, false}}}} {
		v, err := Decode([]byte(`null`), s, testLimits)
		if v != nil || err != ErrConfig {
			t.Fatal(err)
		}
	}
	for _, limits := range []Limits{{}, {64, 129, 10}, {-1, 4, 8}, {64, 4, 0}} {
		if _, err := Decode([]byte(`{}`), modelSchema(), limits); err != ErrConfig {
			t.Fatal(err)
		}
	}
	s := selected(modelSchema())
	s.Cases = append(s.Cases, s.Cases[0])
	if _, err := DecodeSelected([]byte(`{`+base+`}`), s, testLimits); err != ErrConfig {
		t.Fatal(err)
	}
	s = selected(modelSchema())
	other := modelSchema()
	delete(other.Fields, "count")
	s.Cases = append(s.Cases, Case{[]string{"SYNTHETIC", "0.2.0"}, other})
	if _, err := DecodeSelected([]byte(`{"version":"0.1.0","test_id":"SYNTHETIC"}`), s, testLimits); err != ErrRequired {
		t.Fatal("selected schema was bypassed", err)
	}
	if _, err := DecodeSelected([]byte(`{"version":"0.2.0","test_id":"SYNTHETIC"}`), s, testLimits); err != nil {
		t.Fatal(err)
	}
}

func TestValueContainers(t *testing.T) {
	v, err := Decode([]byte(`{`+base+`,"items":[{"name":"a"}]}`), modelSchema(), testLimits)
	if err != nil {
		t.Fatal(err)
	}
	names := v.Names()
	names[0] = "changed"
	if v.Names()[0] != "version" {
		t.Fatal("names alias")
	}
	x, _ := v.Lookup("items")
	elements := x.Elements()
	elements[0] = nil
	if x.Elements()[0] == nil {
		t.Fatal("elements alias")
	}
	boolValue, err := Decode([]byte(`true`), &Schema{Kind: Boolean}, testLimits)
	if err != nil || !boolValue.Bool() {
		t.Fatal(err)
	}
	textValue, err := Decode([]byte(`"\ud83d\ude00"`), &Schema{Kind: String}, testLimits)
	if err != nil || textValue.Text() != "😀" {
		t.Fatal("surrogate pair was not preserved", err)
	}
	textValue, err = Decode([]byte(`"�"`), &Schema{Kind: String}, testLimits)
	if err != nil || textValue.Text() != "�" {
		t.Fatal("literal replacement scalar was not preserved", err)
	}
	textValue, err = Decode([]byte(`"\"\\\/\b\f\n\r\t"`), &Schema{Kind: String}, testLimits)
	if err != nil || textValue.Text() != "\"\\/\b\f\n\r\t" {
		t.Fatal("valid escapes were not preserved", err)
	}
}

// Schema decoding is deterministic and never exposes partial candidates.
func FuzzDecode(f *testing.F) {
	for _, c := range testkit.InvalidJSON() {
		f.Add(c.Input)
	}
	for _, input := range []string{`{` + base + `}`, `{` + base + `,"labels":{"x":"😀"}}`, `{` + base + `,"count":1}`, `[]`, `-0`} {
		f.Add([]byte(input))
	}
	f.Fuzz(func(t *testing.T, input []byte) {
		limits := Limits{4096, 16, 512}
		schema := modelSchema()
		v, e := DecodeSelected(input, selected(schema), limits)
		w, e2 := DecodeSelected(input, selected(schema), limits)
		if e != e2 || !reflect.DeepEqual(v, w) {
			t.Fatal("non-deterministic decode")
		}
		if e != nil && (v != nil || len(e.Error()) > 32) {
			t.Fatal("failed decode exposed state or unbounded error")
		}
	})
}

// The strict parser may narrow the standard JSON grammar but must never accept
// malformed framing. Use the independent standard validator as an oracle.
func FuzzParse(f *testing.F) {
	for _, c := range testkit.InvalidJSON() {
		f.Add(c.Input)
	}
	for _, input := range []string{`null`, `true`, `false`, `"\"\\\/\b\f\n\r\t"`, `"\ud83d\ude00"`, `{"a":0,"b":[-1,1.25,1E+2]}`, `0`, `-0`, `1e-2`, `1e+2`, `01`, `{"":0}`, `{} []`} {
		f.Add([]byte(input))
	}
	f.Fuzz(func(t *testing.T, input []byte) {
		v, err := parse(input, Limits{4096, 16, 512})
		if err == nil && (v == nil || !json.Valid(input)) {
			t.Fatal("strict parser accepted invalid JSON")
		}
		if err != nil && v != nil {
			t.Fatal("failed parse exposed state")
		}
	})
}
