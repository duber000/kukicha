package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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

	if result.registry["mylib.Add"] != 1 {
		t.Errorf("expected mylib.Add=1, got %d", result.registry["mylib.Add"])
	}
	if result.registry["mylib.Divide"] != 2 {
		t.Errorf("expected mylib.Divide=2, got %d", result.registry["mylib.Divide"])
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
	if result.registry["mylib.NewCounter"] != 1 {
		t.Errorf("expected mylib.NewCounter=1, got %d", result.registry["mylib.NewCounter"])
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
	if result.registry["mylib.GetValue"] != 1 {
		t.Errorf("expected mylib.GetValue=1, got %d", result.registry["mylib.GetValue"])
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

	if result.registry["alpha.First"] != 1 {
		t.Errorf("expected alpha.First=1, got %d", result.registry["alpha.First"])
	}
	if result.registry["beta.First"] != 1 {
		t.Errorf("expected beta.First=1, got %d", result.registry["beta.First"])
	}
	if result.registry["beta.Second"] != 2 {
		t.Errorf("expected beta.Second=2, got %d", result.registry["beta.Second"])
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

	if result.registry["pkg.Ambiguous"] != 2 {
		t.Errorf("expected pkg.Ambiguous=2 (larger count wins), got %d", result.registry["pkg.Ambiguous"])
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
	if result.registry["mylib.NewConfig"] != 1 {
		t.Errorf("expected mylib.NewConfig=1, got %d", result.registry["mylib.NewConfig"])
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
	if result.registry["mylib.OldFunc"] != 1 {
		t.Errorf("expected mylib.OldFunc=1, got %d", result.registry["mylib.OldFunc"])
	}
	if result.registry["mylib.NewFunc"] != 1 {
		t.Errorf("expected mylib.NewFunc=1, got %d", result.registry["mylib.NewFunc"])
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
		registry:   map[string]int{},
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
		registry: map[string]int{
			"z.Zebra":  1,
			"a.Alpha":  2,
			"m.Middle": 1,
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
		registry: map[string]int{
			"slice.Filter": 1,
			"pg.Query":     2,
			"fetch.Get":    2,
		},
		deprecated: map[string]string{},
	}

	output := formatRegistry(result)
	src := string(output)
	if !strings.Contains(src, `"slice.Filter"`) || !strings.Contains(src, "1") {
		t.Errorf("expected 'slice.Filter' with value 1 in output:\n%s", src)
	}
	if !strings.Contains(src, `"pg.Query"`) || !strings.Contains(src, "2") {
		t.Errorf("expected 'pg.Query' with value 2 in output:\n%s", src)
	}
	if !strings.Contains(src, `"fetch.Get"`) {
		t.Errorf("expected 'fetch.Get' in output:\n%s", src)
	}
}

func TestFormatRegistry_ReturnCountValues(t *testing.T) {
	result := scanResult{
		registry: map[string]int{
			"pkg.Single": 1,
			"pkg.Double": 2,
			"pkg.Triple": 3,
		},
		deprecated: map[string]string{},
	}

	output := string(formatRegistry(result))

	if !strings.Contains(output, `"pkg.Single": 1`) {
		t.Error("expected Single: 1")
	}
	if !strings.Contains(output, `"pkg.Double": 2`) {
		t.Error("expected Double: 2")
	}
	if !strings.Contains(output, `"pkg.Triple": 3`) {
		t.Error("expected Triple: 3")
	}
}

func TestFormatRegistry_WithDeprecated(t *testing.T) {
	result := scanResult{
		registry: map[string]int{
			"pkg.OldFunc": 1,
			"pkg.NewFunc": 1,
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
