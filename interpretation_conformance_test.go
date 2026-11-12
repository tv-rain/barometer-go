package barometer

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/tv-rain/barometer-go/internal/result"
)

// The shared INTERPRETATION check of the nine measurement case families.
//
// One prohibited-claims check runs over every text the implementation ships
// about results: the comments and string literals of every production Go
// file (including the result, observation and measurement types and their
// field descriptions), the reference documentation, the README and the
// sanitized examples. A family test supplies its test's prohibited claims and
// the qualifiers its guidance must keep; the same check also rejects verdict
// fields in serialized results and requires the strict result decoder to
// reject an injected verdict field as unknown.
//
// A sentence matching a claim pattern fails unless it is negated (for
// example "does not establish", "is not", "never", "outside"); the guidance
// states every limit in that negated form. The case plans, the traceability
// manifests and evidence, the changelog and the specification-issue records
// quote the prohibited claims as requirements or history, so they are not
// shipped guidance and are excluded. This is a portable textual check;
// judgement about wording in context remains the RV-M-*-INTERPRETATION
// review cases.

type claim struct {
	name    string
	pattern *regexp.Regexp
}

func claims(pairs ...string) []claim {
	var out []claim
	for i := 0; i+1 < len(pairs); i += 2 {
		out = append(out, claim{pairs[i], regexp.MustCompile(pairs[i+1])})
	}
	return out
}

// commonClaims apply to every measurement (REQ-*-INTERPRETATION "definitive
// proof of censorship" and verdict rows).
var commonClaims = claims(
	"definitive proof of censorship", `\b(is|are|provides?|constitutes?|gives?)\b[^.]*\b(definitive|conclusive)?\s*proof of (censorship|blocking)`,
	"result proves censorship", `\b(proves?|demonstrates?|confirms?)\b[^.]*\b(censorship|censored|blocking|blocked)\b`,
	"outcome is a verdict", `\b(censorship|blocking|anomaly|success|censored|blocked)\s+verdict\b`,
)

var negation = regexp.MustCompile(`\b(not|never|no|nor|cannot|can't|neither|without|outside|instead of|rather than)\b`)
var sentenceBreak = regexp.MustCompile(`[.;!?](\s+|$)|\n\s*\n|\n\s*[-*|]\s`)

type corpusText struct {
	source string
	text   string
}

var corpusOnce sync.Once
var corpusCache []corpusText
var corpusErr error

// excludedDocs quote prohibited claims as plans, history or findings.
var excludedDocs = []string{"docs/measurement-requirements.md", "docs/cross-cutting-requirements.md", "docs/specification-issues.md", "docs/traceability-manifest.md", "CHANGELOG.md"}

// interpretationCorpus returns the shipped guidance text: comments and string
// literals of production Go files and of the sanitized examples, and the
// Markdown documentation outside the excluded records.
func interpretationCorpus(t *testing.T) []corpusText {
	t.Helper()
	corpusOnce.Do(func() {
		fset := token.NewFileSet()
		corpusErr = filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				switch path {
				case "build", "tools", "traceability", "scripts", ".git", "internal/testkit":
					return filepath.SkipDir
				}
				return nil
			}
			example := strings.HasPrefix(path, "internal/examples/")
			switch {
			case strings.HasSuffix(path, ".go") && (!strings.HasSuffix(path, "_test.go") || example):
				f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
				if err != nil {
					return err
				}
				for _, g := range f.Comments {
					corpusCache = append(corpusCache, corpusText{path, g.Text()})
				}
				ast.Inspect(f, func(n ast.Node) bool {
					if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING {
						if s, err := strconv.Unquote(lit.Value); err == nil {
							corpusCache = append(corpusCache, corpusText{path, s})
						}
					}
					return true
				})
			case strings.HasSuffix(path, ".md") && !slices.Contains(excludedDocs, filepath.ToSlash(path)):
				b, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				corpusCache = append(corpusCache, corpusText{path, string(b)})
			}
			return nil
		})
	})
	if corpusErr != nil || len(corpusCache) == 0 {
		t.Fatal("interpretation corpus unavailable")
	}
	return corpusCache
}

// sentences splits text into lowercase sentences with collapsed whitespace.
func sentences(text string) []string {
	var out []string
	for _, s := range sentenceBreak.Split(text, -1) {
		s = strings.Join(strings.Fields(strings.ToLower(s)), " ")
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// assertNoClaims fails for every non-negated sentence of the corpus that
// makes one of the claims.
func assertNoClaims(t *testing.T, cs []claim) {
	t.Helper()
	for _, c := range interpretationCorpus(t) {
		for _, f := range claimsIn(c.text, cs) {
			t.Errorf("%s states the prohibited claim %s", c.source, f)
		}
	}
}

// claimsIn returns each claim a non-negated sentence of text makes, with the
// sentence.
func claimsIn(text string, cs []claim) []string {
	var out []string
	for _, s := range sentences(text) {
		if negation.MatchString(s) {
			continue
		}
		for _, cl := range cs {
			if cl.pattern.MatchString(s) {
				out = append(out, strconv.Quote(cl.name)+": "+strconv.Quote(s))
			}
		}
	}
	return out
}

// guidance returns the section of docs/measurement-interpretation.md for one
// test, lowercased with collapsed whitespace.
func guidance(t *testing.T, id string) string {
	t.Helper()
	b, err := os.ReadFile("docs/measurement-interpretation.md")
	if err != nil {
		t.Fatal("interpretation guidance unavailable")
	}
	_, after, ok := strings.Cut(string(b), "\n## "+id+"\n")
	if !ok {
		t.Fatalf("no %s interpretation guidance", id)
	}
	section, _, _ := strings.Cut(after, "\n## ")
	return strings.Join(strings.Fields(strings.ToLower(section)), " ")
}

// assertQualifiers requires each qualifier in the test's guidance section.
func assertQualifiers(t *testing.T, id string, qualifiers ...string) {
	t.Helper()
	g := guidance(t, id)
	for _, q := range qualifiers {
		if !strings.Contains(g, strings.ToLower(q)) {
			t.Errorf("%s guidance lacks the qualifier %q", id, q)
		}
	}
}

// verdictTokens may not appear in any member name or string value of a
// serialized measurement.
var verdictTokens = []string{"verdict", "interpret", "censor", "block", "filter", "anomal", "confiden", "inject", "spoof", "legitimate", "authoritative", "correct", "reachab", "intent", "manipulat", "allowlist", "restriction", "operator", "interfer", "tamper", "malicious", "proof", "country", "asn", "geoloc", "reputation"}

// verdictFields are injected into results to prove the decoder rejects them.
var verdictFields = []string{"verdict", "interpretation", "censored", "blocked", "filtered", "anomaly", "confidence", "reachable", "authoritative"}

func words(v any, out *[]string) {
	switch x := v.(type) {
	case map[string]any:
		for k, child := range x {
			*out = append(*out, k)
			words(child, out)
		}
	case []any:
		for _, child := range x {
			words(child, out)
		}
	case string:
		*out = append(*out, x)
	}
}

// assertNoVerdict checks one serialized result: its outcome, when present, is
// registered, and no member name or string value of its measurement carries
// a verdict token.
func assertNoVerdict(t *testing.T, raw []byte, outcomes []string) {
	t.Helper()
	var r map[string]any
	if json.Unmarshal(raw, &r) != nil {
		t.Fatal("result is not JSON")
	}
	m, ok := r["measurement"].(map[string]any)
	if !ok {
		return
	}
	if o, _ := m["outcome"].(string); !slices.Contains(outcomes, o) {
		t.Errorf("unregistered outcome %q", o)
	}
	var ws []string
	words(m, &ws)
	for _, w := range ws {
		// A registered socket outcome names an unreachable network or
		// address; it states the observed error, not reachability.
		lw := strings.ReplaceAll(strings.ToLower(w), "unreachable", "")
		for _, tok := range verdictTokens {
			if strings.Contains(lw, tok) {
				t.Errorf("measurement carries %q", w)
			}
		}
	}
}

// assertRejected requires the strict decoder to accept raw and to reject each
// mutation of it.
func assertRejected(t *testing.T, d *result.Decoder, raw []byte, mutations ...func(map[string]any)) {
	t.Helper()
	if _, err := d.Decode(raw); err != nil {
		t.Fatal("positive control rejected")
	}
	for i, mutate := range mutations {
		var v map[string]any
		_ = json.Unmarshal(raw, &v)
		mutate(v)
		data, _ := json.Marshal(v)
		if _, err := d.Decode(data); err == nil {
			t.Errorf("mutation %d accepted", i)
		}
	}
}

// verdictInjections add each verdict field to the measurement, to its first
// observation and to that observation's details.
func verdictInjections() []func(map[string]any) {
	var out []func(map[string]any)
	for _, f := range verdictFields {
		out = append(out,
			func(v map[string]any) { v["measurement"].(map[string]any)[f] = true },
			func(v map[string]any) { v[f] = "synthetic" },
			func(v map[string]any) {
				if o, ok := v["measurement"].(map[string]any)["observations"].([]any); ok && len(o) > 0 {
					o[0].(map[string]any)[f] = true
				} else {
					v["measurement"].(map[string]any)[f] = "synthetic"
				}
			},
			func(v map[string]any) {
				if o, ok := v["measurement"].(map[string]any)["observations"].([]any); ok && len(o) > 0 {
					if d, ok := o[0].(map[string]any)["details"].(map[string]any); ok {
						d[f] = true
						return
					}
				}
				v["measurement"].(map[string]any)[f] = 1
			},
		)
	}
	return out
}

// interpretation runs the shared check for one test: serialized results,
// injected verdict fields, the prohibited claims over the shipped text and
// the qualifiers the guidance keeps.
func interpretation(t *testing.T, id string, d *result.Decoder, sample []byte, raws [][]byte, outcomes []string, cs []claim, qualifiers ...string) {
	t.Helper()
	if len(raws) == 0 {
		t.Fatal("no results")
	}
	for _, raw := range raws {
		assertNoVerdict(t, raw, outcomes)
	}
	assertRejected(t, d, sample, verdictInjections()...)
	assertNoClaims(t, append(slices.Clone(commonClaims), cs...))
	assertQualifiers(t, id, qualifiers...)
}

// The shared harnesses detect what they exist to detect: a canary in each
// searched encoding and a whole-number canary, but not digits inside another
// number; an affirmative claim, but not the same claim negated.
func TestPrivacyHarnessNegativeControls(t *testing.T) {
	cs := canarySet{{"text", "CANARY-CONTROL"}, {"number", "47391"}}
	for _, leak := range []string{`{"x":"CANARY-CONTROL"}`, `canary-control`, `"Q0FOQVJZLUNPTlRST0w="`, `43414e4152592d434f4e54524f4c`, `{"port":47391}`, `x=canary-control&y`} {
		if len(cs.found([]byte(leak))) == 0 {
			t.Errorf("leak %q not found", leak)
		}
	}
	if f := cs.found([]byte(`{"duration_ms":473910,"run_id":"47391"}`)); len(f) != 0 {
		t.Errorf("false positive %v", f)
	}
	all := append(slices.Clone(commonClaims), claim{"filtering", regexp.MustCompile(`\btimeout\b[^.]*\bestablishes\b[^.]*\bfiltering\b`)})
	if len(claimsIn("A timeout establishes intentional filtering. This result proves censorship.", all)) != 2 {
		t.Error("affirmative claims not found")
	}
	if f := claimsIn("A timeout does not establish filtering. No result proves censorship; it is not a blocking verdict.", all); len(f) != 0 {
		t.Errorf("negated claims reported: %v", f)
	}
}
