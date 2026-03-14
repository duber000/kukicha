package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/duber000/kukicha/internal/ast"
	"github.com/duber000/kukicha/internal/parser"
)

func writeKukiFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestScanRegistry_ExportedFunctions(t *testing.T) {
	dir := t.TempDir()
	path := writeKukiFile(t, dir, "mylib/mylib.kuki", `petiole mylib

func Add(a int, b int) int
    return a + b

func Divide(a int, b int) (int, error)
    return a / b, empty
`)

	result, errs := scanRegistry([]string{path})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if result.registry["mylib.Add"].count != 1 {
		t.Errorf("expected mylib.Add count=1, got %d", result.registry["mylib.Add"].count)
	}
	if result.registry["mylib.Divide"].count != 2 {
		t.Errorf("expected mylib.Divide count=2, got %d", result.registry["mylib.Divide"].count)
	}
}

func TestScanRegistry_ReturnTypes(t *testing.T) {
	dir := t.TempDir()
	path := writeKukiFile(t, dir, "mylib/mylib.kuki", `petiole mylib

func GetName() string
    return "test"

func ReadFile(path string) (list of byte, error)
    return empty, empty
`)

	result, errs := scanRegistry([]string{path})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	entry := result.registry["mylib.GetName"]
	if len(entry.types) != 1 || entry.types[0].kind != "TypeKindString" {
		t.Errorf("expected GetName return TypeKindString, got %v", entry.types)
	}

	entry = result.registry["mylib.ReadFile"]
	if len(entry.types) != 2 {
		t.Fatalf("expected ReadFile 2 return types, got %d", len(entry.types))
	}
	if entry.types[0].kind != "TypeKindList" {
		t.Errorf("expected ReadFile[0] TypeKindList, got %s", entry.types[0].kind)
	}
	if entry.types[1].kind != "TypeKindNamed" || entry.types[1].name != "error" {
		t.Errorf("expected ReadFile[1] TypeKindNamed/error, got %s/%s", entry.types[1].kind, entry.types[1].name)
	}
}

func TestScanRegistry_ParamNames(t *testing.T) {
	dir := t.TempDir()
	path := writeKukiFile(t, dir, "mylib/mylib.kuki", `petiole mylib

func Greet(name string, greeting string) string
    return greeting
`)

	result, errs := scanRegistry([]string{path})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	entry := result.registry["mylib.Greet"]
	if len(entry.paramNames) != 2 || entry.paramNames[0] != "name" || entry.paramNames[1] != "greeting" {
		t.Errorf("expected param names [name, greeting], got %v", entry.paramNames)
	}
}

func TestScanRegistry_SkipsUnexported(t *testing.T) {
	dir := t.TempDir()
	path := writeKukiFile(t, dir, "mylib/mylib.kuki", `petiole mylib

func helper(x int) int
    return x

func Public(x int) int
    return x
`)

	result, errs := scanRegistry([]string{path})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if _, ok := result.registry["mylib.helper"]; ok {
		t.Error("unexported 'helper' should not be in registry")
	}
	if _, ok := result.registry["mylib.Public"]; !ok {
		t.Error("exported 'Public' should be in registry")
	}
}

func TestScanRegistry_SkipsMethods(t *testing.T) {
	dir := t.TempDir()
	path := writeKukiFile(t, dir, "mylib/mylib.kuki", `petiole mylib

type Counter
    value int

func Increment on c reference Counter
    c.value = c.value + 1

func NewCounter() Counter
    return Counter{}
`)

	result, errs := scanRegistry([]string{path})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if _, ok := result.registry["mylib.Increment"]; ok {
		t.Error("method 'Increment' should not be in registry")
	}
	if result.registry["mylib.NewCounter"].count != 1 {
		t.Errorf("expected mylib.NewCounter count=1, got %d", result.registry["mylib.NewCounter"].count)
	}
}

func TestScanRegistry_SkipsVoidFunctions(t *testing.T) {
	dir := t.TempDir()
	path := writeKukiFile(t, dir, "mylib/mylib.kuki", `petiole mylib

func DoSomething()
    print("hello")

func GetValue() string
    return "ok"
`)

	result, errs := scanRegistry([]string{path})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if _, ok := result.registry["mylib.DoSomething"]; ok {
		t.Error("void function 'DoSomething' should not be in registry")
	}
	if result.registry["mylib.GetValue"].count != 1 {
		t.Errorf("expected mylib.GetValue count=1, got %d", result.registry["mylib.GetValue"].count)
	}
}

func TestScanRegistry_SkipsTestFiles(t *testing.T) {
	dir := t.TempDir()
	srcPath := writeKukiFile(t, dir, "mylib/mylib.kuki", `petiole mylib

func Real() int
    return 1
`)
	testPath := writeKukiFile(t, dir, "mylib/mylib_test.kuki", `petiole mylib

func TestHelper() int
    return 42
`)

	result, errs := scanRegistry([]string{srcPath, testPath})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if _, ok := result.registry["mylib.Real"]; !ok {
		t.Error("expected 'Real' from source file in registry")
	}
	if _, ok := result.registry["mylib.TestHelper"]; ok {
		t.Error("function from _test.kuki should not be in registry")
	}
}

func TestScanRegistry_NoPetioleSkipsFile(t *testing.T) {
	dir := t.TempDir()
	path := writeKukiFile(t, dir, "orphan/orphan.kuki", `func Orphan() int
    return 42
`)

	result, errs := scanRegistry([]string{path})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if len(result.registry) != 0 {
		t.Errorf("expected empty registry for file without petiole, got %d entries", len(result.registry))
	}
}

func TestScanRegistry_NonExistentFile(t *testing.T) {
	result, errs := scanRegistry([]string{"/nonexistent/file.kuki"})
	if len(errs) == 0 {
		t.Fatal("expected error for non-existent file")
	}
	if len(result.registry) != 0 {
		t.Error("expected empty registry on read error")
	}
}

func TestScanRegistry_MultiplePackages(t *testing.T) {
	dir := t.TempDir()
	path1 := writeKukiFile(t, dir, "alpha/alpha.kuki", `petiole alpha

func First() int
    return 1
`)
	path2 := writeKukiFile(t, dir, "beta/beta.kuki", `petiole beta

func First() string
    return "one"

func Second() (string, error)
    return "two", empty
`)

	result, errs := scanRegistry([]string{path1, path2})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if result.registry["alpha.First"].count != 1 {
		t.Errorf("expected alpha.First count=1, got %d", result.registry["alpha.First"].count)
	}
	if result.registry["beta.First"].count != 1 {
		t.Errorf("expected beta.First count=1, got %d", result.registry["beta.First"].count)
	}
	if result.registry["beta.Second"].count != 2 {
		t.Errorf("expected beta.Second count=2, got %d", result.registry["beta.Second"].count)
	}
}

func TestScanRegistry_KeepsLargerReturnCount(t *testing.T) {
	dir := t.TempDir()
	path1 := writeKukiFile(t, dir, "pkg/a.kuki", `petiole pkg

func Ambiguous() int
    return 1
`)
	path2 := writeKukiFile(t, dir, "pkg/b.kuki", `petiole pkg

func Ambiguous() (int, error)
    return 1, empty
`)

	result, errs := scanRegistry([]string{path1, path2})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if result.registry["pkg.Ambiguous"].count != 2 {
		t.Errorf("expected pkg.Ambiguous count=2 (larger count wins), got %d", result.registry["pkg.Ambiguous"].count)
	}
}

func TestScanRegistry_EmptyInput(t *testing.T) {
	result, errs := scanRegistry(nil)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors for empty input: %v", errs)
	}
	if len(result.registry) != 0 {
		t.Errorf("expected empty registry, got %d entries", len(result.registry))
	}
}

func TestScanRegistry_SkipsTypeDecls(t *testing.T) {
	dir := t.TempDir()
	path := writeKukiFile(t, dir, "mylib/mylib.kuki", `petiole mylib

type Config
    Port int
    Host string

func NewConfig() Config
    return Config{}
`)

	result, errs := scanRegistry([]string{path})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if len(result.registry) != 1 {
		t.Errorf("expected 1 entry, got %d: %v", len(result.registry), result.registry)
	}
	if result.registry["mylib.NewConfig"].count != 1 {
		t.Errorf("expected mylib.NewConfig count=1, got %d", result.registry["mylib.NewConfig"].count)
	}
}

// =============================================================================
// Deprecated directive scanning tests
// =============================================================================

func TestScanRegistry_DeprecatedFunction(t *testing.T) {
	dir := t.TempDir()
	path := writeKukiFile(t, dir, "mylib/mylib.kuki", `petiole mylib

# kuki:deprecated "Use NewFunc instead"
func OldFunc() string
    return "old"

func NewFunc() string
    return "new"
`)

	result, errs := scanRegistry([]string{path})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	// Both should be in the registry
	if result.registry["mylib.OldFunc"].count != 1 {
		t.Errorf("expected mylib.OldFunc count=1, got %d", result.registry["mylib.OldFunc"].count)
	}
	if result.registry["mylib.NewFunc"].count != 1 {
		t.Errorf("expected mylib.NewFunc count=1, got %d", result.registry["mylib.NewFunc"].count)
	}

	// Only OldFunc should be deprecated
	if msg, ok := result.deprecated["mylib.OldFunc"]; !ok {
		t.Error("expected mylib.OldFunc to be in deprecated map")
	} else if msg != "Use NewFunc instead" {
		t.Errorf("expected deprecation message 'Use NewFunc instead', got %q", msg)
	}
	if _, ok := result.deprecated["mylib.NewFunc"]; ok {
		t.Error("NewFunc should not be in deprecated map")
	}
}

func TestScanRegistry_DeprecatedVoidFunction(t *testing.T) {
	dir := t.TempDir()
	path := writeKukiFile(t, dir, "mylib/mylib.kuki", `petiole mylib

# kuki:deprecated "Use DoNew instead"
func DoOld()
    print("old")
`)

	result, errs := scanRegistry([]string{path})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	// Void functions are not in the return-count registry but should still be in deprecated
	if _, ok := result.registry["mylib.DoOld"]; ok {
		t.Error("void function should not be in return-count registry")
	}
	if msg, ok := result.deprecated["mylib.DoOld"]; !ok {
		t.Error("expected mylib.DoOld to be in deprecated map")
	} else if msg != "Use DoNew instead" {
		t.Errorf("expected deprecation message, got %q", msg)
	}
}

// =============================================================================
// formatRegistry tests
// =============================================================================

func TestFormatRegistry_Empty(t *testing.T) {
	output := formatRegistry(scanResult{
		registry:   map[string]registryEntry{},
		deprecated: map[string]string{},
	})

	src := string(output)
	if !strings.Contains(src, "package semantic") {
		t.Error("expected 'package semantic' in output")
	}
	if !strings.Contains(src, "generatedStdlibRegistry") {
		t.Error("expected 'generatedStdlibRegistry' in output")
	}
	if !strings.Contains(src, "generatedStdlibDeprecated") {
		t.Error("expected 'generatedStdlibDeprecated' in output")
	}
	if !strings.Contains(src, "DO NOT EDIT") {
		t.Error("expected 'DO NOT EDIT' comment in output")
	}
}

func TestFormatRegistry_SortedEntries(t *testing.T) {
	result := scanResult{
		registry: map[string]registryEntry{
			"z.Zebra":  {count: 1, types: []typeRepr{{kind: "TypeKindInt"}}, paramNames: []string{"x"}},
			"a.Alpha":  {count: 2, types: []typeRepr{{kind: "TypeKindString"}, {kind: "TypeKindNamed", name: "error"}}, paramNames: []string{"s"}},
			"m.Middle": {count: 1, types: []typeRepr{{kind: "TypeKindBool"}}, paramNames: []string{"v"}},
		},
		deprecated: map[string]string{},
	}

	output := string(formatRegistry(result))

	alphaIdx := strings.Index(output, `"a.Alpha"`)
	middleIdx := strings.Index(output, `"m.Middle"`)
	zebraIdx := strings.Index(output, `"z.Zebra"`)

	if alphaIdx == -1 || middleIdx == -1 || zebraIdx == -1 {
		t.Fatalf("missing entries in output:\n%s", output)
	}

	if !(alphaIdx < middleIdx && middleIdx < zebraIdx) {
		t.Errorf("entries not sorted: alpha@%d, middle@%d, zebra@%d", alphaIdx, middleIdx, zebraIdx)
	}
}

func TestFormatRegistry_ValidGo(t *testing.T) {
	result := scanResult{
		registry: map[string]registryEntry{
			"slice.Filter": {count: 1, types: []typeRepr{{kind: "TypeKindList"}}, paramNames: []string{"items", "predicate"}},
			"pg.Query":     {count: 2, types: []typeRepr{{kind: "TypeKindNamed", name: "Rows"}, {kind: "TypeKindNamed", name: "error"}}, paramNames: []string{"pool", "query"}},
			"fetch.Get":    {count: 2, types: []typeRepr{{kind: "TypeKindReference"}, {kind: "TypeKindNamed", name: "error"}}, paramNames: []string{"url"}},
		},
		deprecated: map[string]string{},
	}

	output := string(formatRegistry(result))
	if !strings.Contains(output, `"slice.Filter"`) {
		t.Errorf("expected 'slice.Filter' in output:\n%s", output)
	}
	if !strings.Contains(output, `"pg.Query"`) {
		t.Errorf("expected 'pg.Query' in output:\n%s", output)
	}
	if !strings.Contains(output, `"fetch.Get"`) {
		t.Errorf("expected 'fetch.Get' in output:\n%s", output)
	}
	// Verify type info is present
	if !strings.Contains(output, "TypeKindList") {
		t.Errorf("expected TypeKindList in output:\n%s", output)
	}
	if !strings.Contains(output, "TypeKindReference") {
		t.Errorf("expected TypeKindReference in output:\n%s", output)
	}
}

func TestFormatRegistry_WithDeprecated(t *testing.T) {
	result := scanResult{
		registry: map[string]registryEntry{
			"pkg.OldFunc": {count: 1, types: []typeRepr{{kind: "TypeKindString"}}, paramNames: []string{}},
			"pkg.NewFunc": {count: 1, types: []typeRepr{{kind: "TypeKindString"}}, paramNames: []string{}},
		},
		deprecated: map[string]string{
			"pkg.OldFunc": "Use NewFunc instead",
		},
	}

	output := string(formatRegistry(result))
	if !strings.Contains(output, `"pkg.OldFunc": "Use NewFunc instead"`) {
		t.Errorf("expected deprecated entry in output:\n%s", output)
	}
}

func TestTypeAnnotationToRepr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantKind string
		wantName string
	}{
		{"int", "int", "TypeKindInt", ""},
		{"string", "string", "TypeKindString", ""},
		{"bool", "bool", "TypeKindBool", ""},
		{"float64", "float64", "TypeKindFloat", ""},
		{"error", "error", "TypeKindNamed", "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse a minimal .kuki with the return type
			src := "petiole test\nfunc F() " + tt.input + "\n    return empty\n"
			p, err := parser.New(src, "test.kuki")
			if err != nil {
				t.Fatalf("parser init error: %v", err)
			}
			prog, parseErrs := p.Parse()
			if len(parseErrs) > 0 {
				t.Fatalf("parse errors: %v", parseErrs)
			}

			result, scanErrs := scanRegistry([]string{})
			_ = result
			_ = scanErrs

			// Get the function declaration
			fd := prog.Declarations[0].(*ast.FunctionDecl)
			repr := typeAnnotationToRepr(fd.Returns[0])

			if repr.kind != tt.wantKind {
				t.Errorf("kind = %q, want %q", repr.kind, tt.wantKind)
			}
			if repr.name != tt.wantName {
				t.Errorf("name = %q, want %q", repr.name, tt.wantName)
			}
		})
	}
}
