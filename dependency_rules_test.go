package barometer_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// The import rules below mechanically enforce the subset of
// docs/package-architecture.md "Allowed Dependency Direction" and
// "Forbidden Dependencies and Shortcuts" that is visible in import graphs.
// Behavioral rules (no clamping, no network activity during validation, and
// similar) remain review rules; passing this check does not establish them.

const modulePath = "github.com/tv-rain/barometer-go"

type importFile struct {
	path    string // repository-relative file path, for diagnostics only
	pkg     string // importing package path
	test    bool   // _test.go file
	imports []string
}

type importRule struct {
	name string
	// tests reports whether _test.go files are also subject to the rule.
	tests     bool
	applies   func(pkg string) bool
	forbidden func(pkg, imp string) bool
}

func internalPath(name string) string { return modulePath + "/internal/" + name }

// under reports whether p is the package prefix or one of its subpackages.
func under(p, prefix string) bool { return p == prefix || strings.HasPrefix(p, prefix+"/") }

func underAny(p string, names ...string) bool {
	for _, name := range names {
		if under(p, internalPath(name)) {
			return true
		}
	}
	return false
}

func moduleImport(imp string) bool { return under(imp, modulePath) }

// standardLibrary uses the Go convention that standard-library paths have no
// dot in their first element.
func standardLibrary(imp string) bool {
	first, _, _ := strings.Cut(imp, "/")
	return !strings.Contains(first, ".")
}

// networkIO reports network packages. net/netip and net/url are pure value
// parsers and remain allowed.
func networkIO(imp string) bool {
	return under(imp, "net") && imp != "net/netip" && imp != "net/url"
}

// directIO approximates I/O and network access that pure models and
// validation must not perform.
func directIO(imp string) bool {
	return networkIO(imp) || imp == "os" || under(imp, "os/exec") || under(imp, "syscall") || under(imp, "io/fs")
}

func measurementName(p string) string {
	rest, ok := strings.CutPrefix(p, internalPath("measurement")+"/")
	if !ok {
		return ""
	}
	name, _, _ := strings.Cut(rest, "/")
	return name
}

var importRules = []importRule{
	{
		name:      "production packages must not import testkit",
		applies:   func(pkg string) bool { return !underAny(pkg, "testkit") },
		forbidden: func(_, imp string) bool { return underAny(imp, "testkit") },
	},
	{
		name:      "internal packages must not import the root package",
		tests:     true,
		applies:   func(pkg string) bool { return under(pkg, modulePath+"/internal") },
		forbidden: func(_, imp string) bool { return imp == modulePath },
	},
	{
		name:    "domain models import only other models and the shared strict JSON leaf",
		applies: func(pkg string) bool { return underAny(pkg, "model") },
		forbidden: func(_, imp string) bool {
			if moduleImport(imp) {
				return !underAny(imp, "model", "jsondecode")
			}
			return !standardLibrary(imp) || directIO(imp)
		},
	},
	{
		name:      "observation models must not depend on result envelopes",
		applies:   func(pkg string) bool { return underAny(pkg, "model/observation") },
		forbidden: func(_, imp string) bool { return underAny(imp, "model/result", "result", "bundle") },
	},
	{
		name:      "jsondecode is a standard-library leaf",
		applies:   func(pkg string) bool { return underAny(pkg, "jsondecode") },
		forbidden: func(_, imp string) bool { return !standardLibrary(imp) || directIO(imp) },
	},
	{
		name:    "validation is pure and must not reach runtime or I/O owners",
		applies: func(pkg string) bool { return underAny(pkg, "validation") },
		forbidden: func(_, imp string) bool {
			return underAny(imp, "admission", "execution", "measurement", "protocol", "networkcontext", "persistence", "platform", "result", "bundle") || directIO(imp)
		},
	},
	{
		name:    "registry imports only task models, task decoding, strict JSON and the runner contract",
		applies: func(pkg string) bool { return underAny(pkg, "registry") },
		forbidden: func(_, imp string) bool {
			if moduleImport(imp) {
				return !underAny(imp, "model/task", "taskdecode", "jsondecode") && imp != internalPath("measurement")
			}
			return !standardLibrary(imp) || directIO(imp)
		},
	},
	{
		name:      "measurement packages must not import the registry",
		applies:   func(pkg string) bool { return underAny(pkg, "measurement") },
		forbidden: func(_, imp string) bool { return underAny(imp, "registry") },
	},
	{
		name:    "measurement packages must not import one another or bundles",
		applies: func(pkg string) bool { return measurementName(pkg) != "" },
		forbidden: func(pkg, imp string) bool {
			other := measurementName(imp)
			return other != "" && other != measurementName(pkg) || underAny(imp, "bundle")
		},
	},
	{
		name:    "protocol primitives must not schedule, persist, or build results",
		applies: func(pkg string) bool { return underAny(pkg, "protocol") },
		forbidden: func(_, imp string) bool {
			return underAny(imp, "admission", "execution", "measurement", "registry", "persistence", "result", "bundle")
		},
	},
	{
		name:    "result and bundle construction must not import execution or network implementations",
		applies: func(pkg string) bool { return underAny(pkg, "result", "bundle") },
		forbidden: func(_, imp string) bool {
			return underAny(imp, "admission", "execution", "measurement", "protocol", "networkcontext") || networkIO(imp)
		},
	},
	{
		name: "capability contracts must not import adapters or orchestration",
		// Concrete adapters in subpackages may import the contract they implement.
		applies: func(pkg string) bool { return pkg == internalPath("platform") || pkg == internalPath("persistence") },
		forbidden: func(_, imp string) bool {
			if moduleImport(imp) {
				return !underAny(imp, "model")
			}
			return !standardLibrary(imp)
		},
	},
	{
		name:    "admission imports only task models, decoding, validation, registry and capability contracts",
		applies: func(pkg string) bool { return underAny(pkg, "admission") },
		forbidden: func(_, imp string) bool {
			if moduleImport(imp) {
				return !underAny(imp, "model/task", "taskdecode", "validation", "registry", "jsondecode") && imp != internalPath("platform") && imp != internalPath("persistence")
			}
			return !standardLibrary(imp) || directIO(imp)
		},
	},
	{
		name:      "measurement packages must not import admission",
		applies:   func(pkg string) bool { return underAny(pkg, "measurement") },
		forbidden: func(_, imp string) bool { return underAny(imp, "admission") },
	},
	{
		// The registry, and through it admission, links the contract package.
		name:    "the runner contract imports only task and observation models",
		applies: func(pkg string) bool { return pkg == internalPath("measurement") },
		forbidden: func(_, imp string) bool {
			if moduleImport(imp) {
				return !underAny(imp, "model/task", "model/observation")
			}
			return !standardLibrary(imp) || directIO(imp)
		},
	},
	{
		// The direct and proxy path dialers in internal/execution are the
		// production socket owners. Protocol primitives may import network
		// packages to run a protocol over a stream those dialers opened, as
		// internal/protocol/tls does; they never dial.
		name:      "only the network path dialers open sockets",
		applies:   func(pkg string) bool { return !underAny(pkg, "execution", "protocol", "testkit") },
		forbidden: func(_, imp string) bool { return networkIO(imp) },
	},
	{
		name:    "network context imports only models and platform capability contracts",
		applies: func(pkg string) bool { return underAny(pkg, "networkcontext") },
		forbidden: func(_, imp string) bool {
			if moduleImport(imp) {
				return !underAny(imp, "model", "jsondecode") && imp != internalPath("platform")
			}
			return !standardLibrary(imp)
		},
	},
}

func checkImports(files []importFile) []string {
	var violations []string
	for _, f := range files {
		for _, rule := range importRules {
			if f.test && !rule.tests || !rule.applies(f.pkg) {
				continue
			}
			for _, imp := range f.imports {
				if rule.forbidden(f.pkg, imp) {
					violations = append(violations, f.path+": imports "+imp+": "+rule.name)
				}
			}
		}
	}
	sort.Strings(violations)
	return violations
}

// loadImports parses every Go file in the module regardless of build tags,
// so tagged files such as the mobile smoke facade are also checked.
func loadImports(t *testing.T, root string) []importFile {
	t.Helper()
	var files []importFile
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if path != root && (strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") || name == "testdata") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(name, ".go") {
			return nil
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		pkg := modulePath
		if dir := filepath.ToSlash(filepath.Dir(rel)); dir != "." {
			pkg += "/" + dir
		}
		f := importFile{path: rel, pkg: pkg, test: strings.HasSuffix(name, "_test.go")}
		for _, spec := range parsed.Imports {
			imp, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			f.imports = append(f.imports, imp)
		}
		files = append(files, f)
		return nil
	})
	if err != nil {
		t.Fatalf("load imports: %v", err)
	}
	return files
}

func TestDependencyRules(t *testing.T) {
	files := loadImports(t, ".")
	// Guard against a vacuous pass if the walk stops finding the package tree.
	seen := map[string]bool{}
	for _, f := range files {
		seen[f.pkg] = true
	}
	for _, want := range []string{modulePath, internalPath("jsondecode"), internalPath("model/task"), internalPath("taskdecode"), internalPath("validation"), internalPath("registry"), internalPath("admission"), internalPath("testkit/conformance"), internalPath("testkit/delivery"), internalPath("persistence"), internalPath("testkit/replay"), internalPath("measurement"), internalPath("execution")} {
		if !seen[want] {
			t.Fatalf("dependency check did not load %s", want)
		}
	}
	for _, v := range checkImports(files) {
		t.Error(v)
	}
}

func TestDependencyRulesRejectForbiddenImports(t *testing.T) {
	file := func(pkg string, test bool, imports ...string) importFile {
		return importFile{path: "synthetic.go", pkg: internalPath(pkg), test: test, imports: imports}
	}
	cases := []struct {
		name string
		file importFile
		rule string
	}{
		{"production imports testkit", file("result", false, internalPath("testkit")), importRules[0].name},
		{"production imports testkit subpackage", file("taskdecode", false, internalPath("testkit/conformance")), importRules[0].name},
		{"internal imports root", file("config", false, modulePath), importRules[1].name},
		{"internal test imports root", file("config", true, modulePath), importRules[1].name},
		{"model imports validation", file("model/task", false, internalPath("validation")), importRules[2].name},
		{"model imports platform", file("model/result", false, internalPath("platform")), importRules[2].name},
		{"model performs network I/O", file("model/observation", false, "net"), importRules[2].name},
		{"model imports third party", file("model/task", false, "example.com/dependency"), importRules[2].name},
		{"observation imports result model", file("model/observation", false, internalPath("model/result")), importRules[3].name},
		{"jsondecode imports module package", file("jsondecode", false, internalPath("model/task")), importRules[4].name},
		{"jsondecode reads files", file("jsondecode", false, "os"), importRules[4].name},
		{"validation imports admission", file("validation", false, internalPath("admission")), importRules[5].name},
		{"validation performs network I/O", file("validation", false, "net/http"), importRules[5].name},
		{"registry imports admission", file("registry", false, internalPath("admission")), importRules[6].name},
		{"registry imports execution", file("registry", false, internalPath("execution")), importRules[6].name},
		{"registry imports persistence", file("registry", false, internalPath("persistence")), importRules[6].name},
		{"registry imports platform", file("registry", false, internalPath("platform")), importRules[6].name},
		{"registry imports testkit", file("registry", false, internalPath("testkit")), importRules[6].name},
		{"registry imports a measurement", file("registry", false, internalPath("measurement/tcp01")), importRules[6].name},
		{"registry reads files", file("registry", false, "os"), importRules[6].name},
		{"registry imports third party", file("registry", false, "example.com/dependency"), importRules[6].name},
		{"measurement imports registry", file("measurement/tcp01", false, internalPath("registry")), importRules[7].name},
		{"measurement primitive imports registry", file("measurement", false, internalPath("registry")), importRules[7].name},
		{"measurement imports another measurement", file("measurement/dns01", false, internalPath("measurement/tcp01")), importRules[8].name},
		{"measurement builds bundles", file("measurement/tls01", false, internalPath("bundle")), importRules[8].name},
		{"protocol persists state", file("protocol/dns", false, internalPath("persistence")), importRules[9].name},
		{"protocol schedules retries", file("protocol/http", false, internalPath("execution")), importRules[9].name},
		{"result imports execution", file("result", false, internalPath("execution")), importRules[10].name},
		{"bundle imports protocol", file("bundle", false, internalPath("protocol/http")), importRules[10].name},
		{"capability imports adapter", file("platform", false, internalPath("platform/android")), importRules[11].name},
		{"persistence imports orchestration", file("persistence", false, internalPath("admission")), importRules[11].name},
		{"persistence imports testkit", file("persistence", false, internalPath("testkit/replay")), importRules[0].name},
		{"admission imports the test replay store", file("admission", false, internalPath("testkit/replay")), importRules[0].name},
		{"admission imports execution", file("admission", false, internalPath("execution")), importRules[12].name},
		{"admission imports a measurement", file("admission", false, internalPath("measurement/tcp01")), importRules[12].name},
		{"admission imports network context", file("admission", false, internalPath("networkcontext")), importRules[12].name},
		{"admission imports results", file("admission", false, internalPath("result")), importRules[12].name},
		{"admission imports configuration", file("admission", false, internalPath("config")), importRules[12].name},
		{"admission imports a platform adapter", file("admission", false, internalPath("platform/android")), importRules[12].name},
		{"admission imports testkit", file("admission", false, internalPath("testkit/delivery")), importRules[0].name},
		{"admission opens sockets", file("admission", false, "net"), importRules[12].name},
		{"admission reads files", file("admission", false, "os"), importRules[12].name},
		{"admission imports third party", file("admission", false, "example.com/dependency"), importRules[12].name},
		{"model imports admission", file("model/task", false, internalPath("admission")), importRules[2].name},
		{"measurement imports admission", file("measurement/tcp01", false, internalPath("admission")), importRules[13].name},
		{"protocol imports admission", file("protocol/http", false, internalPath("admission")), importRules[9].name},
		{"runner contract imports a runner", file("measurement", false, internalPath("measurement/tcp01")), importRules[14].name},
		{"runner contract imports execution", file("measurement", false, internalPath("execution")), importRules[14].name},
		{"runner contract opens sockets", file("measurement", false, "net"), importRules[14].name},
		{"network context opens sockets", file("networkcontext", false, "net"), importRules[15].name},
		{"runner opens sockets", file("measurement/tcp01", false, "net"), importRules[15].name},
		{"root facade opens sockets", importFile{path: "synthetic.go", pkg: modulePath, imports: []string{"net/http"}}, importRules[15].name},
		{"network context imports execution", file("networkcontext", false, internalPath("execution")), importRules[16].name},
		{"network context imports results", file("networkcontext", false, internalPath("result")), importRules[16].name},
		{"network context imports a platform adapter", file("networkcontext", false, internalPath("platform/android")), importRules[16].name},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			violations := checkImports([]importFile{tc.file})
			found := false
			for _, v := range violations {
				found = found || strings.HasSuffix(v, ": "+tc.rule)
			}
			if !found {
				t.Fatalf("expected rule %q, got %q", tc.rule, violations)
			}
		})
	}

	allowed := []importFile{
		file("testkit/conformance", false, internalPath("testkit"), internalPath("taskdecode")),
		file("taskdecode", true, internalPath("testkit")),
		file("taskdecode", false, internalPath("validation"), internalPath("jsondecode"), internalPath("model/task")),
		file("registry", false, internalPath("model/task"), internalPath("taskdecode"), "errors"),
		file("admission", false, internalPath("registry"), internalPath("model/task"), internalPath("taskdecode"), internalPath("validation"), internalPath("platform"), internalPath("persistence"), "context", "crypto/rand", "sync"),
		file("admission", true, internalPath("testkit"), internalPath("testkit/delivery")),
		file("testkit/delivery", false, internalPath("model/task"), "crypto/ed25519"),
		file("persistence", false, internalPath("model/task"), "context", "crypto/sha256"),
		file("testkit/replay", false, internalPath("persistence"), internalPath("model/task"), "encoding/json"),
		file("admission", true, internalPath("testkit/replay")),
		file("execution", false, internalPath("registry")),
		file("registry", false, internalPath("measurement")),
		file("measurement", false, internalPath("model/task"), internalPath("model/observation"), "context", "io", "time"),
		file("execution", false, internalPath("measurement"), internalPath("result"), internalPath("validation"), "crypto/rand"),
		file("model/observation", false, internalPath("jsondecode"), "strings", "net/netip", "net/url"),
		file("model/result", false, internalPath("model/observation"), internalPath("model/task")),
		file("measurement/dns01", false, internalPath("measurement"), internalPath("protocol/dns"), internalPath("model/observation")),
		file("measurement/tcp01", false, internalPath("measurement"), internalPath("taskdecode"), internalPath("jsondecode"), internalPath("model/task"), internalPath("protocol/tcp")),
		file("measurement/tls01", false, internalPath("measurement"), internalPath("taskdecode"), internalPath("model/observation"), internalPath("protocol/tcp"), internalPath("protocol/tls"), "crypto/x509"),
		file("protocol/tls", false, internalPath("model/observation"), "crypto/tls", "net"),
		{path: "registry.go", pkg: modulePath, imports: []string{internalPath("registry"), internalPath("measurement/tcp01"), internalPath("measurement/tls01")}},
		file("bundle", false, internalPath("config"), internalPath("result"), internalPath("model/result")),
		file("platform", false, "context", "time"),
		file("platform/android", false, internalPath("platform")),
		file("execution", false, "net", internalPath("networkcontext")),
		file("networkcontext", false, internalPath("platform"), internalPath("model/observation"), "encoding/json"),
		{path: "synthetic.go", pkg: modulePath, imports: []string{internalPath("execution")}},
	}
	if v := checkImports(allowed); len(v) != 0 {
		t.Fatalf("allowed imports rejected: %q", v)
	}
}

// The admission path cannot reach the network: the complete transitive
// dependency closure of internal/admission, as built for the host platform,
// contains no network, TLS, process-execution or plugin package. Since
// DEV-644 the root package links the direct-path dialer of
// internal/execution, so its closure contains package net, and since DEV-647
// crypto/tls for the Network Context endpoint request; since DEV-665 the
// TLS-01 runner links the bounded handshake of internal/protocol/tls. The
// check then requires that internal/execution and internal/protocol/tls are
// the only module packages in that closure importing a network or TLS
// package, and that the closure contains no
// process-execution, plugin or testkit package. The import rule "only the
// direct path opens sockets" keeps the root package and admission from
// importing network packages themselves. Rejected input therefore produces no
// network activity through code linked into the facade unless it reaches the
// engine, which admission precedes. Activity through a future trusted
// adapter seam, and builds for other platforms, are not covered.
func TestAdmissionHasNoNetworkDependency(t *testing.T) {
	// Prefer the toolchain running this test, which go test exports as GOROOT.
	gocmd, err := exec.LookPath(filepath.Join(os.Getenv("GOROOT"), "bin", "go"))
	if os.Getenv("GOROOT") == "" || err != nil {
		if gocmd, err = exec.LookPath("go"); err != nil {
			t.Fatal("go command unavailable; the dependency closure cannot be checked")
		}
	}
	for _, pkg := range []string{".", "./internal/admission"} {
		out, err := exec.Command(gocmd, "list", "-deps", "-f", `{{.ImportPath}} {{join .Imports " "}}`, pkg).Output()
		if err != nil {
			t.Fatalf("go list %s: %v", pkg, err)
		}
		var deps []string
		importers := map[string][]string{}    // network package -> module packages importing it
		tlsImporters := map[string][]string{} // TLS package -> module packages importing it
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			fields := strings.Fields(line)
			if len(fields) == 0 {
				continue
			}
			deps = append(deps, fields[0])
			for _, imp := range fields[1:] {
				if moduleImport(fields[0]) && networkIO(imp) {
					importers[imp] = append(importers[imp], fields[0])
				}
				if moduleImport(fields[0]) && under(imp, "crypto/tls") {
					tlsImporters[imp] = append(tlsImporters[imp], fields[0])
				}
			}
		}
		// Guard against a vacuous pass on an empty or truncated listing.
		if len(deps) < 10 || !slices.Contains(deps, internalPath("admission")) || !slices.Contains(deps, internalPath("jsondecode")) {
			t.Fatalf("go list %s returned an implausible closure", pkg)
		}
		root := pkg == "."
		for _, dep := range deps {
			if (networkIO(dep) || under(dep, "crypto/tls")) && !root || under(dep, "os/exec") || dep == "plugin" || under(dep, internalPath("testkit")) {
				t.Errorf("%s transitively depends on %s", pkg, dep)
			}
		}
		for imp, by := range tlsImporters {
			for _, p := range by {
				if p != internalPath("execution") && p != internalPath("protocol/tls") {
					t.Errorf("%s: %s imports %s", pkg, p, imp)
				}
			}
		}
		for imp, by := range importers {
			for _, p := range by {
				if p != internalPath("execution") && p != internalPath("protocol/tls") && !(p == internalPath("protocol/http") && imp == "net/http") {
					t.Errorf("%s: %s imports %s", pkg, p, imp)
				}
			}
		}
		if root && !slices.Contains(importers["net"], internalPath("execution")) {
			t.Fatalf("go list %s: the direct-path dialer is missing from the closure", pkg)
		}
	}
}

// The clearing path cannot reach the network (DEV-564): persistence.Clear
// and the stores it clears live in internal/persistence, whose transitive
// dependency closure for the host platform contains no network, TLS,
// process-execution or plugin package. The root package's closure is checked
// by TestAdmissionHasNoNetworkDependency. Erasers supplied by a future native
// adapter are not covered.
func TestClearingHasNoNetworkDependency(t *testing.T) {
	gocmd, err := exec.LookPath(filepath.Join(os.Getenv("GOROOT"), "bin", "go"))
	if os.Getenv("GOROOT") == "" || err != nil {
		if gocmd, err = exec.LookPath("go"); err != nil {
			t.Fatal("go command unavailable; the dependency closure cannot be checked")
		}
	}
	for _, pkg := range []string{"./internal/persistence", "./internal/persistence/pending"} {
		out, err := exec.Command(gocmd, "list", "-deps", pkg).Output()
		if err != nil {
			t.Fatalf("go list %s: %v", pkg, err)
		}
		deps := strings.Fields(string(out))
		if len(deps) < 10 || !slices.Contains(deps, internalPath("persistence")) || !slices.Contains(deps, "context") {
			t.Fatalf("go list %s returned an implausible closure", pkg)
		}
		for _, dep := range deps {
			if networkIO(dep) || under(dep, "crypto/tls") || under(dep, "os/exec") || dep == "plugin" || under(dep, internalPath("testkit")) {
				t.Errorf("%s transitively depends on %s", pkg, dep)
			}
		}
	}
}
