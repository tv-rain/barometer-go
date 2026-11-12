package barometer

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/tv-rain/barometer-go/internal/model/observation"
	"github.com/tv-rain/barometer-go/internal/result"
)

// The shared MINIMIZATION harness of the nine measurement case families.
//
// A family test gives every item its linked requirement prohibits one unique
// synthetic canary, drives each outcome path through the production codec,
// runner, result construction and pending store with scripted transports, and
// then hands every surface the result path exposes to this harness: each
// serialized result, the bundle returned by NextBundle, the pending store's
// durable and staged bytes, and the text of every error returned by the
// facade, including task-validation rejections. The library has no log or
// other diagnostic surface, so these are the complete result and error paths.
//
// A canary is searched for in its raw spelling, case-insensitively, and in its
// JSON-escaped, percent-encoded, hexadecimal and base64 encodings. A canary
// made only of digits is matched as a whole number, so a duration or a
// quantity that merely contains the digits does not count. Random run IDs are
// redacted first. Canaries are synthetic; no participant or live measurement
// data is used.

// canary is one prohibited item and the unique synthetic value carrying it.
type canary struct{ item, value string }

type canarySet []canary

var digitsOnly = regexp.MustCompile(`^[0-9]+$`)
var runIDField = regexp.MustCompile(`"run_id":"[A-Za-z0-9_-]*"`)

// encodings returns the spellings in which a string canary is searched for.
func encodings(v string) []string {
	quoted, _ := json.Marshal(v)
	forms := []string{v, string(quoted[1 : len(quoted)-1]), url.QueryEscape(v), url.PathEscape(v), hex.EncodeToString([]byte(v)),
		base64.StdEncoding.EncodeToString([]byte(v)), base64.RawStdEncoding.EncodeToString([]byte(v)), base64.RawURLEncoding.EncodeToString([]byte(v))}
	slices.Sort(forms)
	return slices.Compact(forms)
}

// absent fails when any canary reaches the surface in any searched encoding.
// Diagnostics name the item and surface, never the surrounding bytes.
func (cs canarySet) absent(t *testing.T, surface string, data []byte) {
	t.Helper()
	for _, item := range cs.found(data) {
		t.Errorf("%s retained the %s canary", surface, item)
	}
}

// found returns the items whose canary appears in data.
func (cs canarySet) found(data []byte) []string {
	text := runIDField.ReplaceAllString(string(data), `"run_id":""`)
	lower := strings.ToLower(text)
	var out []string
	for _, c := range cs {
		if digitsOnly.MatchString(c.value) {
			if regexp.MustCompile(`(^|[^0-9])` + c.value + `([^0-9]|$)`).MatchString(text) {
				out = append(out, c.item)
			}
			continue
		}
		for _, form := range encodings(c.value) {
			if strings.Contains(lower, strings.ToLower(form)) {
				out = append(out, c.item)
				break
			}
		}
	}
	return out
}

// unique fails when two items share a canary or one text canary contains
// another, so a match always identifies one item. Binary canaries may nest:
// a certificate's DER contains its public key and signature.
func (cs canarySet) unique(t *testing.T) {
	t.Helper()
	binary := func(c canary) bool {
		return strings.HasSuffix(c.item, "(hex)") || strings.HasSuffix(c.item, "(base64)")
	}
	for i, a := range cs {
		for j, b := range cs {
			if i != j && (a.value == b.value || !binary(a) && !binary(b) && strings.Contains(a.value, b.value)) {
				t.Fatalf("canary for %s is not unique", a.item)
			}
		}
	}
}

// envelopeFields is the closed set of Measurement Result members.
var envelopeFields = []string{"schema_version", "task_id", "run_id", "attempt", "test", "implementation", "time", "execution", "network_context", "measurement"}

// fieldPaths records every member path below v. Object members are joined
// with "." and array elements are written as "[]".
func fieldPaths(v any, prefix string, out map[string]bool) {
	switch x := v.(type) {
	case map[string]any:
		for k, child := range x {
			p := k
			if prefix != "" {
				p = prefix + "." + k
			}
			out[p] = true
			fieldPaths(child, p, out)
		}
	case []any:
		for _, child := range x {
			fieldPaths(child, prefix+"[]", out)
		}
	}
}

// allowedPaths expands each allowed leaf path to every enclosing path.
func allowedPaths(leaves []string) map[string]bool {
	out := map[string]bool{}
	for _, l := range leaves {
		parts := strings.Split(l, ".")
		for i := range parts {
			p := strings.Join(parts[:i+1], ".")
			out[p] = true
			out[strings.TrimSuffix(p, "[]")] = true
		}
	}
	return out
}

// minimized checks one serialized result: only envelope members, no member
// named in forbidden, and a measurement, when present, whose every member path
// is in allowed. It returns the decoded measurement or nil.
func minimized(t *testing.T, raw []byte, allowed []string, forbidden ...string) map[string]any {
	t.Helper()
	var r map[string]any
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatal("result is not JSON")
	}
	for k := range r {
		if !slices.Contains(envelopeFields, k) || slices.Contains(forbidden, k) {
			t.Errorf("result carries member %q", k)
		}
	}
	m, _ := r["measurement"].(map[string]any)
	if m == nil {
		return nil
	}
	ok := allowedPaths(allowed)
	paths := map[string]bool{}
	fieldPaths(m, "", paths)
	for p := range paths {
		if !ok[p] {
			t.Errorf("measurement carries %s outside the allowed set", p)
		}
	}
	return m
}

// surfaces collects every result and error surface of one harness run.
type surfaces struct {
	results [][]byte
	errs    []string
	extra   map[string][]byte
}

func (s *surfaces) err(err error) {
	if err != nil {
		s.errs = append(s.errs, err.Error())
	}
}

// collect adds the pending store's results and bytes and the encoded bundle.
func (s *surfaces) collect(h *harness, c *Client) {
	s.results = append(s.results, h.run.output.Records()...)
	if s.extra == nil {
		s.extra = map[string][]byte{}
	}
	s.extra["pending store"] = append(s.extra["pending store"], h.run.output.Dump()...)
	b, err := c.NextBundle()
	s.err(err)
	s.extra["bundle"] = append(s.extra["bundle"], b...)
}

// check scans every collected surface for every canary.
func (s *surfaces) check(t *testing.T, cs canarySet) {
	t.Helper()
	for _, r := range s.results {
		cs.absent(t, "result", r)
	}
	for name, b := range s.extra {
		cs.absent(t, name, b)
	}
	cs.absent(t, "error", []byte(strings.Join(s.errs, "\n")))
}

// rejectedClean admits taskJSON, which must be rejected, and scans the error.
// The task carries canaries in members the codec refuses, so the rejection is
// the task-validation error path.
func rejectedClean(t *testing.T, h *harness, c *Client, cs canarySet, taskJSON string) {
	t.Helper()
	handle, err := c.AdmitTask(taskJSON, h.embedded(t, taskJSON))
	if err == nil || handle != "" {
		t.Fatal("canary task admitted")
	}
	cs.absent(t, "task-validation error", []byte(err.Error()))
}

// withUnknownInput adds an unknown input member whose name and value are canaries.
func withUnknownInput(taskJSON string, cs canarySet) string {
	var b bytes.Buffer
	for _, c := range cs {
		if !digitsOnly.MatchString(c.value) {
			b.WriteString(c.value)
		}
	}
	member, _ := json.Marshal(b.String())
	pair := string(member) + `:` + string(member)
	if strings.Contains(taskJSON, `"input":{}`) {
		return strings.Replace(taskJSON, `"input":{}`, `"input":{`+pair+`}`, 1)
	}
	return strings.Replace(taskJSON, `"input":{`, `"input":{`+pair+`,`, 1)
}

// primitiveDecoder is the strict result decoder of one direct-path result
// with the given output selection, for injected-field checks.
func primitiveDecoder(t *testing.T, taskID string, o result.Output) *result.Decoder {
	t.Helper()
	cd, _ := observation.NewContextDecoder(result.Contract, observation.ContextApprovals{})
	path, _ := observation.NewPath(1, 0, "direct", "")
	d, err := result.New(result.Binding{TaskID: taskID, Output: o, MaxAttempts: 1, MaxDurationMS: 60000, Implementation: implementation, ContextDecoder: cd, Path: path, AccessGeneration: 1})
	if err != nil {
		t.Fatal("synthetic binding rejected")
	}
	return d
}

// binaryCanaries searches for bytes the test does not choose, such as a
// certificate or a handshake random, in hexadecimal and base64.
func binaryCanaries(item string, b []byte) canarySet {
	return canarySet{{item + " (hex)", hex.EncodeToString(b)}, {item + " (base64)", base64.StdEncoding.EncodeToString(b)}}
}

// recordingConn keeps the first bytes the client writes and reads, so a test
// can derive canaries from handshake values it cannot choose.
type recordingConn struct {
	net.Conn
	mu      sync.Mutex
	out, in []byte
}

const recordLimit = 16384

func (c *recordingConn) Write(p []byte) (int, error) {
	c.mu.Lock()
	if len(c.out) < recordLimit {
		c.out = append(c.out, p[:min(len(p), recordLimit-len(c.out))]...)
	}
	c.mu.Unlock()
	return c.Conn.Write(p)
}

func (c *recordingConn) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	c.mu.Lock()
	if len(c.in) < recordLimit {
		c.in = append(c.in, p[:min(n, recordLimit-len(c.in))]...)
	}
	c.mu.Unlock()
	return n, err
}

// helloCanaries derives canaries from the plaintext TLS hello messages the
// connection carried: the ClientHello random, legacy session identifier,
// offered cipher-suite list and extension bytes, and the ServerHello random,
// session identifier echo and extension bytes. A message that did not arrive
// adds nothing.
func (c *recordingConn) helloCanaries() canarySet {
	c.mu.Lock()
	defer c.mu.Unlock()
	var cs canarySet
	hello := func(side string, b []byte, typ byte) {
		// record header (5), handshake type (1) and length (3), version (2)
		if len(b) < 44 || b[0] != 0x16 || b[5] != typ {
			return
		}
		cs = append(cs, binaryCanaries(side+" random", b[11:43])...)
		sid := int(b[43])
		rest := b[44:]
		if sid > len(rest) {
			return
		}
		if sid > 0 {
			cs = append(cs, binaryCanaries(side+" session identifier", rest[:sid])...)
		}
		rest = rest[sid:]
		if typ == 1 && len(rest) >= 2 {
			n := int(rest[0])<<8 | int(rest[1])
			if 2+n <= len(rest) {
				cs = append(cs, binaryCanaries("offered cipher suites", rest[2:2+n])...)
				rest = rest[2+n:]
				if len(rest) >= 1 && 1+int(rest[0]) <= len(rest) {
					rest = rest[1+int(rest[0]):]
				}
			}
		} else if typ == 2 && len(rest) >= 3 {
			rest = rest[3:] // negotiated suite and compression
		}
		if len(rest) >= 2 {
			n := int(rest[0])<<8 | int(rest[1])
			if n >= 16 && 2+n <= len(rest) {
				cs = append(cs, binaryCanaries(side+" extension bytes", rest[2:2+n])...)
			}
		}
	}
	hello("ClientHello", c.out, 1)
	hello("ServerHello", c.in, 2)
	return cs
}
