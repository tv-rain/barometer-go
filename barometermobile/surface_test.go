package barometermobile

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// mobileAPI is the reviewed binding surface of decision 0014. A change here is
// a breaking binding change and a public-API review point; the documented
// block in docs/public-api.md must match it.
var mobileAPI = []string{
	"type Records interface",
	"type LinkageKeyStore interface",
	"type Clock interface",
	"type Network interface",
	"type NetworkObserver interface",
	"type Consent interface",
	"type Device interface",
	"type TimeZone interface",
	"type SystemResolver interface",
	"type Adapters struct",
	"type RecordValue struct",
	"type StringList struct",
	"type AccessState struct",
	"type DeviceState struct",
	"type Client struct",
	"func NewRecordValue(value []byte) *RecordValue",
	"func NewStringList() *StringList",
	"func (l *StringList) Add(s string)",
	"func NewAccessState(generation int64, networkType string, vpn string, ipv4 string, ipv6 string, validated string, captivePortal string) *AccessState",
	"func NewDeviceState(batteryPercent int64, batteryKnown bool, charging bool, chargingKnown bool) *DeviceState",
	"func NewClient(adapters *Adapters, trust []byte) (*Client, error)",
	"func (c *Client) InstallEmbeddedTask(issuerID string, taskJSON []byte) (string, error)",
	"func (c *Client) InstallRemoteTask(signedPackage []byte) (string, error)",
	"func (c *Client) Initialize(configJSON string) error",
	"func (c *Client) AdmitTask(taskJSON string, sourceRef string) (string, error)",
	"func (c *Client) StartRun(taskHandle string) (string, error)",
	"func (c *Client) RunState(runHandle string) (string, error)",
	"func (c *Client) CancelRun(runHandle string) error",
	"func (c *Client) NextBundle() (string, error)",
	"func (c *Client) AcknowledgeBundle(bundleID string) error",
	"func (c *Client) SetProxy(proxyJSON string) error",
	"func (c *Client) ClearProxy() error",
	"func (c *Client) ClearMeasurementData() error",
	"func (c *Client) Close() error",
}

// structsWithFields are the only exported structs with exported fields.
var structsWithFields = []string{"Adapters"}

// rootLifecycle is the root Client lifecycle the facade must mirror.
var rootLifecycle = []string{"Initialize", "AdmitTask", "StartRun", "RunState", "CancelRun", "NextBundle", "AcknowledgeBundle", "SetProxy", "ClearProxy", "ClearMeasurementData", "Close"}

func facadeFiles(t *testing.T) (*token.FileSet, []*ast.File) {
	t.Helper()
	paths, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	var files []*ast.File
	for _, p := range paths {
		if strings.HasSuffix(p, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, p, nil, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}
		if len(f.Comments) > 0 && f.Comments[0].Pos() < f.Package && strings.HasPrefix(f.Comments[0].List[0].Text, "//go:build") {
			t.Errorf("%s: build-constrained facade file", p)
		}
		files = append(files, f)
	}
	return fset, files
}

// allowedType reports whether e is one of the types decision 0014 lets cross
// the facade: string, []byte, int64, bool, error, a facade interface, or a
// pointer to a facade struct.
func allowedType(e ast.Expr, interfaces, structs map[string]bool) bool {
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name == "string" || x.Name == "int64" || x.Name == "bool" || x.Name == "error" || interfaces[x.Name]
	case *ast.ArrayType:
		id, ok := x.Elt.(*ast.Ident)
		return x.Len == nil && ok && id.Name == "byte"
	case *ast.StarExpr:
		id, ok := x.X.(*ast.Ident)
		return ok && structs[id.Name]
	}
	return false
}

func checkFields(t *testing.T, where string, fl *ast.FieldList, interfaces, structs map[string]bool) {
	t.Helper()
	if fl == nil {
		return
	}
	for _, f := range fl.List {
		if !allowedType(f.Type, interfaces, structs) {
			t.Errorf("%s: type %s may not cross the facade", where, types(f.Type))
		}
	}
}

func types(e ast.Expr) string {
	var b bytes.Buffer
	printer.Fprint(&b, token.NewFileSet(), e)
	return b.String()
}

func TestFacadeSurfaceIsTheReviewedBindingSurface(t *testing.T) {
	fset, files := facadeFiles(t)
	interfaces, structs := map[string]bool{}, map[string]bool{}
	for _, f := range files {
		for _, d := range f.Decls {
			if g, ok := d.(*ast.GenDecl); ok {
				for _, s := range g.Specs {
					if ts, ok := s.(*ast.TypeSpec); ok && ts.Name.IsExported() {
						switch ts.Type.(type) {
						case *ast.InterfaceType:
							interfaces[ts.Name.Name] = true
						case *ast.StructType:
							structs[ts.Name.Name] = true
						default:
							t.Errorf("exported type %s is neither an interface nor a struct", ts.Name.Name)
						}
					}
				}
			}
		}
	}
	var exported []string
	var methods []string
	for _, f := range files {
		for _, d := range f.Decls {
			switch d := d.(type) {
			case *ast.FuncDecl:
				if !d.Name.IsExported() {
					continue
				}
				if d.Recv != nil {
					recv := d.Recv.List[0].Type
					star, pointer := recv.(*ast.StarExpr)
					if pointer {
						recv = star.X
					}
					id, ok := recv.(*ast.Ident)
					if !ok || !id.IsExported() {
						continue // methods of unexported types do not cross the facade
					}
					if !pointer {
						t.Errorf("method %s.%s has a value receiver", id.Name, d.Name.Name)
					}
					if id.Name == "Client" {
						methods = append(methods, d.Name.Name)
					}
				}
				if d.Type.TypeParams != nil {
					t.Errorf("exported generic function %s", d.Name.Name)
				}
				checkFields(t, d.Name.Name, d.Type.Params, interfaces, structs)
				checkFields(t, d.Name.Name, d.Type.Results, interfaces, structs)
				var b bytes.Buffer
				fn := *d
				fn.Body, fn.Doc = nil, nil
				if err := printer.Fprint(&b, fset, &fn); err != nil {
					t.Fatal(err)
				}
				exported = append(exported, b.String())
			case *ast.GenDecl:
				for _, s := range d.Specs {
					switch s := s.(type) {
					case *ast.ValueSpec:
						for _, name := range s.Names {
							if name.IsExported() {
								t.Errorf("exported value %s", name.Name)
							}
						}
					case *ast.TypeSpec:
						if !s.Name.IsExported() {
							continue
						}
						switch ty := s.Type.(type) {
						case *ast.InterfaceType:
							exported = append(exported, "type "+s.Name.Name+" interface")
							for _, m := range ty.Methods.List {
								ft, ok := m.Type.(*ast.FuncType)
								if !ok {
									t.Errorf("interface %s embeds %s", s.Name.Name, types(m.Type))
									continue
								}
								checkFields(t, s.Name.Name, ft.Params, interfaces, structs)
								checkFields(t, s.Name.Name, ft.Results, interfaces, structs)
							}
						case *ast.StructType:
							exported = append(exported, "type "+s.Name.Name+" struct")
							for _, field := range ty.Fields.List {
								if len(field.Names) == 0 {
									t.Errorf("embedded field in %s", s.Name.Name)
								}
								for _, name := range field.Names {
									if !name.IsExported() {
										continue
									}
									if !slices.Contains(structsWithFields, s.Name.Name) {
										t.Errorf("exported field %s.%s", s.Name.Name, name.Name)
									}
									if id, ok := field.Type.(*ast.Ident); !ok || !interfaces[id.Name] {
										t.Errorf("field %s.%s is not a facade interface", s.Name.Name, name.Name)
									}
								}
							}
						}
					}
				}
			}
		}
	}
	slices.Sort(exported)
	want := slices.Clone(mobileAPI)
	slices.Sort(want)
	if !slices.Equal(exported, want) {
		t.Fatalf("exported facade surface\n got: %q\nwant: %q", exported, want)
	}
	for _, m := range rootLifecycle {
		if !slices.Contains(methods, m) {
			t.Errorf("facade lacks root lifecycle method %s", m)
		}
	}
	documented := documentedMobileSignatures(t)
	for _, sig := range mobileAPI {
		if !slices.Contains(documented, sig) {
			t.Errorf("facade signature is not documented: %s", sig)
		}
	}
	for _, sig := range documented {
		if !slices.Contains(mobileAPI, sig) {
			t.Errorf("documented facade signature is not implemented: %s", sig)
		}
	}
}

// The facade mirrors the root lifecycle signatures exactly.
func TestFacadeLifecycleMatchesTheRoot(t *testing.T) {
	src, err := os.ReadFile("../public_api_test.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range rootLifecycle {
		re := regexp.MustCompile(`"func \(c \*Client\) ` + m + `\(([^"]*)"`)
		root := re.FindStringSubmatch(string(src))
		if root == nil {
			t.Fatalf("root signature of %s not found", m)
		}
		if !slices.Contains(mobileAPI, "func (c *Client) "+m+"("+root[1]) {
			t.Errorf("facade %s differs from the root signature %s", m, root[1])
		}
	}
}

// documentedMobileSignatures reads the facade block of docs/public-api.md.
// Interface and struct declarations are reduced to their first line.
func documentedMobileSignatures(t *testing.T) []string {
	t.Helper()
	doc, err := os.ReadFile("../docs/public-api.md")
	if err != nil {
		t.Fatal(err)
	}
	_, section, ok := strings.Cut(string(doc), "<!-- mobile-api-signatures:start -->")
	if !ok {
		t.Fatal("documented facade block not found")
	}
	section, _, _ = strings.Cut(section, "<!-- mobile-api-signatures:end -->")
	m := regexp.MustCompile("(?s)```go\n(.*?)\n```").FindStringSubmatch(section)
	if m == nil {
		t.Fatal("documented facade signature block not found")
	}
	var out []string
	for _, line := range strings.Split(m[1], "\n") {
		switch {
		case strings.HasPrefix(line, "func "):
			out = append(out, line)
		case strings.HasPrefix(line, "type "):
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				kind, _, _ := strings.Cut(fields[2], "{")
				out = append(out, "type "+fields[1]+" "+kind)
			}
		}
	}
	return out
}
