package barometer

import (
	"fmt"
	"strings"
	"testing"
)

// strictSchema is one registered production test schema exercised through
// Client.AdmitTask. DNS-01 and TCP-01 keep their own strict-schema tests.
type strictSchema struct {
	name           string
	input, options string // accepted baseline
	// member, integer, optMember and optInteger are member texts present in
	// the baseline: a fixed and an integer input member and a fixed and an
	// integer options member. nested is a member inside a value a later stage
	// decodes (a NET-05 step).
	member, integer, nested, optMember, optInteger string
	harness                                        func(t *testing.T) (*harness, func() int32)
	task                                           func(id, input, options string) string
}

func strictSchemas() []strictSchema {
	replaceFields := func(base, input, options string) string {
		return strings.Replace(strings.Replace(base, `"input":{}`, `"input":`+input, 1), `"options":{}`, `"options":`+options, 1)
	}
	return []strictSchema{
		{
			name: "TLS-01", input: tlsInput, options: tlsOptions,
			member: `"server_name":"measurement.example"`, integer: `"port":443`, optMember: `"alpn_protocols":["h2","http/1.1"]`,
			harness: func(t *testing.T) (*harness, func() int32) {
				h, d := tlsProductionHarness(t, newTLSFixture(t))
				return h, d.calls.Load
			},
			task: func(id, input, options string) string {
				s := strings.Replace(tlsTask(id), `"input":`+tlsInput, `"input":`+input, 1)
				return strings.Replace(s, `"options":`+tlsOptions, `"options":`+options, 1)
			},
		},
		{
			name: "HTTP-01", input: httpInput, options: `{"follow_redirects":false,"max_body_bytes":1024}`,
			member: `"url":"https://8.8.8.8/check"`, integer: `"expected_status":200`, optMember: `"follow_redirects":false`, optInteger: `"max_body_bytes":1024`,
			harness: func(t *testing.T) (*harness, func() int32) {
				h, d, _ := httpHarness(t, httpInput, responseHTTP(""), false)
				return h, d.calls.Load
			},
			task: httpTask,
		},
		{
			name: "NET-01", input: `{}`, options: `{}`,
			harness: func(t *testing.T) (*harness, func() int32) {
				h, n, d := net01Harness(t)
				return h, func() int32 { return int32(n.Reads()) + d.calls.Load() }
			},
			task: func(id, input, options string) string { return replaceFields(net01Task(id), input, options) },
		},
		{
			name: "NET-03", input: net03Input, options: `{"address_family":"auto"}`,
			member: `"url":"https://8.8.8.8/check"`, integer: `"expected_status":200`, optMember: `"address_family":"auto"`,
			harness: func(t *testing.T) (*harness, func() int32) {
				h, d, _, n, _ := net03Harness(t, net03Input, "", "unknown", false, false)
				return h, func() int32 { return int32(n.Reads()) + d.calls.Load() }
			},
			task: net03Task,
		},
		{
			name: "NET-05", input: profileInput(primitiveSteps()...), options: `{}`,
			member: `"profile_id":"baseline"`, integer: `"profile_version":1`, nested: `"port":443`,
			harness: func(t *testing.T) (*harness, func() int32) {
				h, d := net05AdmissionHarness(t, nil)
				return h, d.calls.Load
			},
			task: func(id, input, options string) string { return net05Task(id, input, options, "") },
		},
		{
			name: "DNS-03", input: dohInput, options: `{"address_family":"auto"}`,
			member: `"hostname":"query.example"`, optMember: `"address_family":"auto"`,
			harness: func(t *testing.T) (*harness, func() int32) {
				h, d := dohHarness(t, newDohFixture(t), dohInput)
				return h, func() int32 { return d.calls.Load() + d.lookups.Load() }
			},
			task: dohTask,
		},
		{
			name: "DNS-04", input: dotInput, options: `{"address_family":"auto"}`,
			member: `"hostname":"query.example"`, integer: `"port":853`, optMember: `"address_family":"auto"`,
			harness: func(t *testing.T) (*harness, func() int32) {
				h, d := dotHarness(t, newDotFixture(t), dotInput)
				return h, func() int32 { return d.calls.Load() + d.lookups.Load() }
			},
			task: dotTask,
		},
	}
}

// escapeName spells the first character of a member name as a \u escape,
// which decodes to the same name.
func escapeName(member string) string {
	return fmt.Sprintf(`"\u%04x%s`, member[1], member[2:])
}

func appendMember(object, member string) string {
	if object == "{}" {
		return "{" + member + "}"
	}
	return object[:len(object)-1] + "," + member + "}"
}

func memberName(member string) string { return member[:strings.Index(member, ":")] }

// hostile returns the rejected input and options variants of one schema.
func (s strictSchema) hostile() (inputs, options []string) {
	inputs = []string{appendMember(s.input, `"x":1`)}
	if s.member != "" {
		name := memberName(s.member)
		inputs = append(inputs, appendMember(s.input, s.member), appendMember(s.input, escapeName(s.member)))
		for _, v := range []string{`null`, `[]`, `{}`, `true`} {
			inputs = append(inputs, strings.Replace(s.input, s.member, name+":"+v, 1))
		}
	}
	inputs = append(inputs, respell(s.input, s.integer)...)
	if s.nested != "" {
		for _, extra := range []string{s.nested, escapeName(s.nested), `"x":1`} {
			inputs = append(inputs, strings.Replace(s.input, s.nested, s.nested+","+extra, 1))
		}
	}
	options = []string{appendMember(s.options, `"x":1`)}
	if s.optMember != "" {
		options = append(options, appendMember(s.options, s.optMember), appendMember(s.options, escapeName(s.optMember)))
	}
	return inputs, append(options, respell(s.options, s.optInteger)...)
}

// respell replaces an integer member of object with non-canonical spellings.
func respell(object, member string) []string {
	if member == "" {
		return nil
	}
	name, v := memberName(member), member[strings.Index(member, ":")+1:]
	var out []string
	for _, spelling := range []string{`-0`, v + `.0`, v + `e0`, v + `E+0`, `"` + v + `"`} {
		out = append(out, strings.Replace(object, member, name+":"+spelling, 1))
	}
	return out
}

// Every registered production schema other than DNS-01 and TCP-01 rejects
// duplicate and escaped-equivalent names, escaped unknown names, wrong-type
// values and non-canonical integer spellings in input and options, for both
// reviewed origins. A rejection publishes no handle, replay record, run or
// result and makes no network or platform contact; escaped known names and
// the baseline are accepted.
// requirement-ref: REQ-JSON-0003
// requirement-ref: REQ-JSON-0004
// requirement-ref: REQ-JSON-0005
// requirement-ref: REQ-JSON-0006
// requirement-ref: REQ-JSON-0010
func TestRegisteredSchemasStrictJSON(t *testing.T) {
	for _, s := range strictSchemas() {
		inputs, options := s.hostile()
		type variant struct {
			input, options string
			valid          bool
		}
		variants := []variant{{s.input, s.options, true}}
		if s.member != "" {
			variants = append(variants, variant{strings.Replace(s.input, s.member, escapeName(s.member), 1), s.options, true})
		}
		if s.optMember != "" {
			variants = append(variants, variant{s.input, strings.Replace(s.options, s.optMember, escapeName(s.optMember), 1), true})
		}
		for _, in := range inputs {
			variants = append(variants, variant{in, s.options, false})
		}
		for _, opt := range options {
			variants = append(variants, variant{s.input, opt, false})
		}
		for _, origin := range []string{"embedded", "remote"} {
			for i, v := range variants {
				t.Run(fmt.Sprintf("%s/%s/%d", s.name, origin, i), func(t *testing.T) {
					h, contacts := s.harness(t)
					c := h.initialized(t)
					defer c.Close()
					handle, err := c.AdmitTask(h.deliver(t, origin, s.task("strict", v.input, v.options)))
					if v.valid && (err != nil || handle == "") {
						t.Fatalf("%s %s: %s", v.input, v.options, codeOf(err))
					}
					if !v.valid {
						if codeOf(err) != "task_schema_invalid" || handle != "" {
							t.Fatalf("%s %s: %s", v.input, v.options, codeOf(err))
						}
						assertNoAdmissionEffects(t, h, c)
					}
					if contacts() != 0 {
						t.Fatal("contact during admission")
					}
				})
			}
		}
	}
}

// The registered 49,152-byte input and 8,192-byte options bounds apply to the
// original encoded value of each production schema, including whitespace.
// requirement-ref: REQ-JSON-0009
// requirement-ref: REQ-JSON-0010
func TestRegisteredSchemasStrictJSONBounds(t *testing.T) {
	for _, s := range strictSchemas() {
		for _, field := range []string{"input", "options"} {
			for _, delta := range []int{-1, 0, 1} {
				t.Run(fmt.Sprintf("%s/%s/%d", s.name, field, delta), func(t *testing.T) {
					size, content := 49152, s.input
					if field == "options" {
						size, content = 8192, s.options
					}
					padded := content[:len(content)-1] + strings.Repeat(" ", size+delta-len(content)) + "}"
					input, options := s.input, s.options
					if field == "input" {
						input = padded
					} else {
						options = padded
					}
					h, contacts := s.harness(t)
					c := h.initialized(t)
					defer c.Close()
					data := s.task("bytes", input, options)
					handle, err := c.AdmitTask(data, h.embedded(t, data))
					if delta <= 0 && (err != nil || handle == "") {
						t.Fatal(codeOf(err))
					}
					if delta > 0 {
						if codeOf(err) != "task_limit_exceeded" || handle != "" {
							t.Fatal(codeOf(err))
						}
						assertNoAdmissionEffects(t, h, c)
					}
					if contacts() != 0 {
						t.Fatal("contact during admission")
					}
				})
			}
		}
	}
}

func assertNoAdmissionEffects(t *testing.T, h *harness, c *Client) {
	t.Helper()
	if taskHandles(c) != 0 || h.replay.Commits() != 0 || len(c.rt.runs.entries) != 0 || len(h.run.output.Records()) != 0 {
		t.Fatal("non-atomic rejection")
	}
}

// Arbitrary input and options for any of these schemas either admit or reject
// atomically, and admission never contacts the network or platform.
// requirement-ref: REQ-JSON-0010
func FuzzRegisteredSchemaAdmission(f *testing.F) {
	schemas := strictSchemas()
	for i, s := range schemas {
		f.Add(uint8(i), s.input, s.options)
		inputs, options := s.hostile()
		f.Add(uint8(i), inputs[len(inputs)-1], options[len(options)-1])
	}
	f.Add(uint8(0), `null`, `{}`)
	f.Fuzz(func(t *testing.T, sel uint8, input, options string) {
		if len(input)+len(options) > 60000 {
			t.Skip()
		}
		s := schemas[int(sel)%len(schemas)]
		h, contacts := s.harness(t)
		h.adapters.deadline = defaultLifecycleDeadline
		c := h.initialized(t)
		defer c.Close()
		data := s.task("fuzz", input, options)
		handle, err := c.AdmitTask(data, h.embedded(t, data))
		if contacts() != 0 || len(h.run.output.Records()) != 0 || len(c.rt.runs.entries) != 0 {
			t.Fatal("admission executed")
		}
		if err != nil && (handle != "" || taskHandles(c) != 0 || h.replay.Commits() != 0) {
			t.Fatal("non-atomic rejection")
		}
	})
}
