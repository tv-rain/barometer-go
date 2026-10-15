package conformance

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func fixtureBytes(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/defaults.json")
	if err != nil {
		t.Fatal("cannot load positive control")
	}
	return b
}

func mutate(t *testing.T, edit func(map[string]any)) []byte {
	t.Helper()
	var wire map[string]any
	if json.Unmarshal(fixtureBytes(t), &wire) != nil {
		t.Fatal("invalid control")
	}
	edit(wire)
	b, err := json.Marshal(wire)
	if err != nil {
		t.Fatal("cannot encode mutation")
	}
	return b
}

// Requirement IDs live in each checked fixture; they identify the narrow
// assertion, not complete coverage of every obligation bearing that ID.
func TestConformanceFixtures(t *testing.T) {
	paths, err := filepath.Glob("testdata/*.json")
	if err != nil || len(paths) < 2 {
		t.Fatal("missing corpus")
	}
	seen := map[string]bool{}
	positive, negative := 0, 0
	for _, path := range paths {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal("cannot load corpus")
		}
		f, err := Parse(b)
		if err != nil {
			t.Fatal("invalid corpus fixture")
		}
		if seen[f.ID()] {
			t.Fatal("duplicate case identity")
		}
		seen[f.ID()] = true
		if f.expected.Rejection == "" {
			positive++
		} else {
			negative++
		}
		t.Run(f.ID(), func(t *testing.T) {
			for i := 0; i < 3; i++ {
				if err := Run(f); err != nil {
					t.Fatalf("requirements %v: %v", f.Requirements(), err)
				}
			}
			t.Run("parallel", func(t *testing.T) {
				for i := 0; i < 4; i++ {
					t.Run("read", func(t *testing.T) {
						t.Parallel()
						if Run(f) != nil {
							t.Fatal("concurrent result differs")
						}
					})
				}
			})
		})
	}
	if positive == 0 || negative == 0 {
		t.Fatal("corpus requires acceptance and rejection controls")
	}
}

func TestStrictFixtureValidation(t *testing.T) {
	cases := map[string]func(map[string]any){
		"version":               func(w map[string]any) { w["format_version"] = 2 },
		"integer version":       func(w map[string]any) { w["format_version"] = json.RawMessage(`1.0`) },
		"missing":               func(w map[string]any) { delete(w, "local_policy") },
		"unknown":               func(w map[string]any) { w["unknown"] = true },
		"null":                  func(w map[string]any) { w["trusted_context"] = nil },
		"context":               func(w map[string]any) { w["trusted_context"] = "production" },
		"policy":                func(w map[string]any) { w["local_policy"] = "production" },
		"adapter":               func(w map[string]any) { w["adapter"] = "latest" },
		"id":                    func(w map[string]any) { w["case_id"] = "bad\nname" },
		"no requirements":       func(w map[string]any) { w["requirements"] = []string{} },
		"unknown requirement":   func(w map[string]any) { w["requirements"] = []string{"REQ-JSON-9999"} },
		"duplicate requirement": func(w map[string]any) { w["requirements"] = []string{"REQ-JSON-0001", "REQ-JSON-0001"} },
		"too many requirements": func(w map[string]any) { w["requirements"] = make([]string, 65) },
		"base64":                func(w map[string]any) { w["task_base64"] = "?" },
		"base64 newline":        func(w map[string]any) { w["task_base64"] = w["task_base64"].(string) + "\n" },
		"task bounds":           func(w map[string]any) { w["task_base64"] = base64.StdEncoding.EncodeToString(make([]byte, 65538)) },
		"no expectation":        func(w map[string]any) { w["expected"] = map[string]any{"rejection": ""} },
		"both expectations":     func(w map[string]any) { w["expected"].(map[string]any)["rejection"] = "field" },
		"unknown class":         func(w map[string]any) { w["expected"] = map[string]any{"rejection": "any-error"} },
		"unknown output":        func(w map[string]any) { w["expected"].(map[string]any)["normalized"].(map[string]any)["extra"] = true },
		"missing output field": func(w map[string]any) {
			delete(w["expected"].(map[string]any)["normalized"].(map[string]any), "enabled")
		},
	}
	for name, edit := range cases {
		t.Run(name, func(t *testing.T) {
			if f, err := Parse(mutate(t, edit)); f != nil || !errors.Is(err, ErrFixture) {
				t.Fatal("invalid fixture accepted")
			}
		})
	}
	control := fixtureBytes(t)
	for _, b := range [][]byte{nil, []byte(`null`), append(append([]byte(nil), control...), []byte(` {}`)...), bytes.Replace(control, []byte(`"format_version": 1`), []byte(`"format_version":1,"format_\u0076ersion":1`), 1), bytes.Repeat([]byte{' '}, MaxFixtureBytes+1), []byte(`{"format_version":1,"case_id":"\ud800"}`)} {
		if f, err := Parse(b); f != nil || err != ErrFixture {
			t.Fatal("raw invalid fixture accepted")
		}
	}
}

func TestMismatchDetectionAndBytePreservation(t *testing.T) {
	for _, edit := range []func(map[string]any){
		func(w map[string]any) {
			w["expected"].(map[string]any)["normalized"].(map[string]any)["enabled"] = false
		},
		func(w map[string]any) { w["expected"] = map[string]any{"rejection": "field"} },
		func(w map[string]any) {
			w["task_base64"] = base64.StdEncoding.EncodeToString([]byte(`{`))
			w["expected"] = map[string]any{"rejection": "field"}
		},
	} {
		f, err := Parse(mutate(t, edit))
		if err != nil {
			t.Fatal("mutation is not structurally valid")
		}
		if Run(f) != ErrMismatch {
			t.Fatal("wrong expectation passed")
		}
	}
	raw := []byte(" \xff{\"x\":1,\"x\":2} ")
	b := mutate(t, func(w map[string]any) {
		w["task_base64"] = base64.StdEncoding.EncodeToString(raw)
		w["expected"] = map[string]any{"rejection": "text"}
	})
	f, err := Parse(b)
	if err != nil || !bytes.Equal(raw, f.task) {
		t.Fatal("exact bytes changed")
	}
	for i := range b {
		b[i] = 0
	}
	refs := f.Requirements()
	refs[0] = "changed"
	if Run(f) != nil || reflect.DeepEqual(refs, f.Requirements()) {
		t.Fatal("fixture retains mutable caller storage")
	}
	if Run(nil) != ErrFixture {
		t.Fatal("nil fixture accepted")
	}
	for _, err := range []error{ErrFixture, ErrMismatch} {
		if len(err.Error()) > 64 || strings.Contains(err.Error(), string(raw)) {
			t.Fatal("unbounded diagnostics")
		}
	}
}

func TestSyntheticEnvironment(t *testing.T) {
	source, policy, timing, err := environment()
	if err != nil || !source.Valid() || !policy.Valid() || source.PolicyGeneration() != policy.Generation() {
		t.Fatal("invalid synthetic models")
	}
	if _, ok := policy.Issuer(source.IssuerID()); !ok {
		t.Fatal("missing synthetic issuer")
	}
	if timing.AcceptedAt.Year() != 2026 || timing.IssuedAt == nil || !timing.AcceptedAt.Equal(*timing.IssuedAt) {
		t.Fatal("clock is not fixed")
	}
	if _, err := json.Marshal(source); err == nil {
		t.Fatal("runtime context serialization enabled")
	}
	if _, err := json.Marshal(policy); err == nil {
		t.Fatal("runtime policy serialization enabled")
	}
}
