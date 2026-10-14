package jsondecode

import (
	"encoding/json"
	"unicode/utf8"
)

type parser struct {
	input      []byte
	pos, nodes int
	limits     Limits
}

// requirement-ref: REQ-JSON-0001
// requirement-ref: REQ-JSON-0002
// requirement-ref: REQ-JSON-0009
func parse(input []byte, limits Limits) (*Value, error) {
	if len(input) > limits.MaxBytes {
		return nil, ErrBounds
	}
	if !utf8.Valid(input) {
		return nil, ErrText
	}
	p := parser{input: input, limits: limits}
	p.space()
	v, err := p.value(1)
	if err != nil {
		return nil, err
	}
	p.space()
	if p.pos != len(input) {
		return nil, ErrSyntax
	}
	return v, nil
}

func (p *parser) space() {
	for p.pos < len(p.input) {
		switch p.input[p.pos] {
		case ' ', '\t', '\r', '\n':
			p.pos++
		default:
			return
		}
	}
}
func (p *parser) take(c byte) bool {
	if p.pos < len(p.input) && p.input[p.pos] == c {
		p.pos++
		return true
	}
	return false
}

func (p *parser) value(depth int) (*Value, error) {
	if depth > p.limits.MaxDepth || p.nodes >= p.limits.MaxNodes {
		return nil, ErrBounds
	}
	p.nodes++
	if p.pos >= len(p.input) {
		return nil, ErrSyntax
	}
	v := &Value{start: p.pos}
	switch p.input[p.pos] {
	case '{':
		p.pos++
		v.kind = Object
		v.members = map[string]*Value{}
		p.space()
		if p.take('}') {
			break
		}
		for {
			name, err := p.string()
			if err != nil {
				return nil, err
			}
			// requirement-ref: REQ-JSON-0003
			if _, exists := v.members[name]; exists {
				return nil, ErrDuplicate
			}
			p.space()
			if !p.take(':') {
				return nil, ErrSyntax
			}
			p.space()
			x, err := p.value(depth + 1)
			if err != nil {
				return nil, err
			}
			v.members[name] = x
			v.order = append(v.order, name)
			p.space()
			if p.take('}') {
				break
			}
			if !p.take(',') {
				return nil, ErrSyntax
			}
			p.space()
		}
	case '[':
		p.pos++
		v.kind = Array
		p.space()
		if p.take(']') {
			break
		}
		for {
			x, err := p.value(depth + 1)
			if err != nil {
				return nil, err
			}
			v.elements = append(v.elements, x)
			p.space()
			if p.take(']') {
				break
			}
			if !p.take(',') {
				return nil, ErrSyntax
			}
			p.space()
		}
	case '"':
		v.kind = String
		s, err := p.string()
		if err != nil {
			return nil, err
		}
		v.text = s
	case 't':
		if !p.literal("true") {
			return nil, ErrSyntax
		}
		v.kind = Boolean
		v.boolean = true
	case 'f':
		if !p.literal("false") {
			return nil, ErrSyntax
		}
		v.kind = Boolean
	case 'n':
		if !p.literal("null") {
			return nil, ErrSyntax
		}
		v.kind = Null
	default:
		v.kind = Integer
		start := p.pos
		p.take('-')
		if !p.take('0') {
			if !p.digit19() {
				return nil, ErrSyntax
			}
			p.digits()
		}
		if p.take('.') {
			if !p.digit() {
				return nil, ErrSyntax
			}
			p.digits()
		}
		if p.take('e') || p.take('E') {
			if !p.take('+') {
				p.take('-')
			}
			if !p.digit() {
				return nil, ErrSyntax
			}
			p.digits()
		}
		v.text = string(p.input[start:p.pos])
	}
	v.end = p.pos
	return v, nil
}

func (p *parser) literal(s string) bool {
	if len(p.input)-p.pos < len(s) || string(p.input[p.pos:p.pos+len(s)]) != s {
		return false
	}
	p.pos += len(s)
	return true
}
func (p *parser) digit() bool {
	if p.pos < len(p.input) && p.input[p.pos] >= '0' && p.input[p.pos] <= '9' {
		p.pos++
		return true
	}
	return false
}
func (p *parser) digit19() bool {
	if p.pos < len(p.input) && p.input[p.pos] >= '1' && p.input[p.pos] <= '9' {
		p.pos++
		return true
	}
	return false
}
func (p *parser) digits() {
	for p.digit() {
	}
}

// Validate Unicode escape pairs before calling the standard string unmarshaller,
// whose replacement of unpaired surrogates is unsuitable for this contract.
func (p *parser) string() (string, error) {
	start := p.pos
	if !p.take('"') {
		return "", ErrSyntax
	}
	for p.pos < len(p.input) {
		c := p.input[p.pos]
		p.pos++
		if c == '"' {
			var s string
			if json.Unmarshal(p.input[start:p.pos], &s) != nil {
				return "", ErrText
			}
			return s, nil
		}
		if c < 0x20 {
			return "", ErrText
		}
		if c != '\\' {
			continue
		}
		if p.pos >= len(p.input) {
			return "", ErrText
		}
		escape := p.input[p.pos]
		p.pos++
		switch escape {
		case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
		case 'u':
			x, ok := p.hex4()
			if !ok {
				return "", ErrText
			}
			if x >= 0xdc00 && x <= 0xdfff {
				return "", ErrText
			}
			if x >= 0xd800 && x <= 0xdbff {
				if !p.take('\\') || !p.take('u') {
					return "", ErrText
				}
				y, ok := p.hex4()
				if !ok || y < 0xdc00 || y > 0xdfff {
					return "", ErrText
				}
			}
		default:
			return "", ErrText
		}
	}
	return "", ErrSyntax
}

func (p *parser) hex4() (uint16, bool) {
	if len(p.input)-p.pos < 4 {
		return 0, false
	}
	var x uint16
	for range 4 {
		c := p.input[p.pos]
		p.pos++
		var n byte
		switch {
		case c >= '0' && c <= '9':
			n = c - '0'
		case c >= 'a' && c <= 'f':
			n = c - 'a' + 10
		case c >= 'A' && c <= 'F':
			n = c - 'A' + 10
		default:
			return 0, false
		}
		x = x*16 + uint16(n)
	}
	return x, true
}
