package barometer_test

import (
	"bytes"
	"go/ast"
	"go/build/constraint"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// implementedAPI is the subset of docs/public-api.md's intended signatures
// that production code exports. Each later task adds its documented method
// here as its own public-API review point.
var implementedAPI = []string{
	"type Client",
	"func NewClient() *Client",
	"func (c *Client) Initialize(configJSON string) error",
	"func (c *Client) AdmitTask(taskJSON string, sourceRef string) (string, error)",
	"func (c *Client) StartRun(taskHandle string) (string, error)",
	"func (c *Client) RunState(runHandle string) (string, error)",
	"func (c *Client) ClearMeasurementData() error",
	"func (c *Client) Close() error",
}

// productionFiles parses the root package's non-test files that a normal build
// includes, so the mobile-smoke-only fixture is excluded.
func productionFiles(t *testing.T) (*token.FileSet, []*ast.File) {
	t.Helper()
	paths, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	var files []*ast.File
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}
		if excludedByDefault(f) {
			continue
		}
		files = append(files, f)
	}
	return fset, files
}

func excludedByDefault(f *ast.File) bool {
	for _, group := range f.Comments {
		if group.Pos() >= f.Package {
			break
		}
		for _, c := range group.List {
			if expr, err := constraint.Parse(c.Text); err == nil {
				return !expr.Eval(func(string) bool { return false })
			}
		}
	}
	return false
}

func documentedSignatures(t *testing.T) []string {
	t.Helper()
	doc, err := os.ReadFile("docs/public-api.md")
	if err != nil {
		t.Fatal(err)
	}
	_, section, _ := strings.Cut(string(doc), "<!-- public-api-signatures:start -->")
	section, _, _ = strings.Cut(section, "<!-- public-api-signatures:end -->")
	m := regexp.MustCompile("(?s)```go\n(.*?)\n```").FindStringSubmatch(section)
	if m == nil {
		t.Fatal("documented signature block not found")
	}
	var out []string
	for _, line := range strings.Split(m[1], "\n") {
		if strings.HasPrefix(line, "func ") {
			out = append(out, line)
		} else if strings.HasPrefix(line, "type ") {
			name, _, _ := strings.Cut(strings.TrimPrefix(line, "type "), " ")
			out = append(out, "type "+name)
		}
	}
	return out
}

func TestPublicSurfaceMatchesDocumentedSignatures(t *testing.T) {
	fset, files := productionFiles(t)
	var exported []string
	for _, f := range files {
		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if !d.Name.IsExported() {
					continue
				}
				if d.Recv != nil {
					recv := d.Recv.List[0].Type
					if star, ok := recv.(*ast.StarExpr); ok {
						recv = star.X
					}
					if id, ok := recv.(*ast.Ident); !ok || !id.IsExported() {
						continue // methods of unexported types are not public API
					}
				}
				var b bytes.Buffer
				fn := *d
				fn.Body, fn.Doc = nil, nil
				if err := printer.Fprint(&b, fset, &fn); err != nil {
					t.Fatal(err)
				}
				exported = append(exported, b.String())
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if !s.Name.IsExported() {
							continue
						}
						exported = append(exported, "type "+s.Name.Name)
						st, ok := s.Type.(*ast.StructType)
						if !ok || s.TypeParams != nil {
							t.Errorf("exported type %s must be an opaque struct", s.Name.Name)
							continue
						}
						for _, field := range st.Fields.List {
							for _, name := range field.Names {
								if name.IsExported() {
									t.Errorf("exported field %s.%s", s.Name.Name, name.Name)
								}
							}
							if len(field.Names) == 0 {
								t.Errorf("embedded field in %s", s.Name.Name)
							}
						}
					case *ast.ValueSpec:
						for _, name := range s.Names {
							if name.IsExported() {
								exported = append(exported, "value "+name.Name)
							}
						}
					}
				}
			}
		}
	}
	documented := documentedSignatures(t)
	for _, sig := range implementedAPI {
		if !slices.Contains(documented, sig) {
			t.Errorf("implemented signature is not documented: %s", sig)
		}
	}
	slices.Sort(exported)
	want := slices.Clone(implementedAPI)
	slices.Sort(want)
	if !slices.Equal(exported, want) {
		t.Fatalf("exported identifiers\n got: %q\nwant: %q", exported, want)
	}
}

// The lifecycle shell performs no network activity: the root package's
// production files import no network package, directly or through test
// support. Indirect activity through adapters is a review rule.
func TestRootPackageImportsNoNetworkPackages(t *testing.T) {
	_, files := productionFiles(t)
	for _, f := range files {
		for _, spec := range f.Imports {
			imp, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			if networkIO(imp) || under(imp, modulePath+"/internal/testkit") {
				t.Errorf("root production file imports %s", imp)
			}
		}
	}
}
