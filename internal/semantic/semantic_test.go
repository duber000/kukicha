package semantic

import (
	"strings"
	"testing"
	"testing/synctest"

	"github.com/duber000/kukicha/internal/parser"
)

func TestSimpleFunctionAnalysis(t *testing.T) {
	input := `func Add(a int, b int) int
    return a + b
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	if len(errors) > 0 {
		t.Fatalf("semantic errors: %v", errors)
	}
}

func TestUndefinedVariable(t *testing.T) {
	input := `func Test() int
    return x
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	if len(errors) == 0 {
		t.Fatal("expected error for undefined variable")
	}

	if !strings.Contains(errors[0].Error(), "undefined identifier 'x'") {
		t.Errorf("expected undefined identifier error, got: %v", errors[0])
	}
}

func TestTypeCompatibility(t *testing.T) {
	input := `func Test() int
    x := "hello"
    return x
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	if len(errors) == 0 {
		t.Fatal("expected error for type mismatch")
	}

	if !strings.Contains(errors[0].Error(), "cannot return") {
		t.Errorf("expected type mismatch error, got: %v", errors[0])
	}
}

func TestVariableDeclaration(t *testing.T) {
	input := `func Test() int
    x := 42
    y := x + 10
    return y
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	if len(errors) > 0 {
		t.Fatalf("unexpected semantic errors: %v", errors)
	}
}

func TestForLoopVariables(t *testing.T) {
	input := `func Test(items list of int) int
    sum := 0
    for item in items
        sum = sum + item
    return sum
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	if len(errors) > 0 {
		t.Fatalf("unexpected semantic errors: %v", errors)
	}
}

func TestTypeDeclaration(t *testing.T) {
	input := `type Person
    Name string
    Age int
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	if len(errors) > 0 {
		t.Fatalf("unexpected semantic errors: %v", errors)
	}
}

func TestMethodReceiver(t *testing.T) {
	input := `type Counter
    Value int
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	if len(errors) > 0 {
		t.Fatalf("unexpected semantic errors: %v", errors)
	}
}

func TestReturnValueCount(t *testing.T) {
	input := `func GetPair() (int, int)
    return 1
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	if len(errors) == 0 {
		t.Fatal("expected error for wrong return value count")
	}

	if !strings.Contains(errors[0].Error(), "expected 2 return values") {
		t.Errorf("expected wrong return count error, got: %v", errors[0])
	}
}

func TestUndefinedType(t *testing.T) {
	input := `func Test(p UnknownType)
    print(p)
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	if len(errors) == 0 {
		t.Fatal("expected error for undefined type")
	}

	if !strings.Contains(errors[0].Error(), "undefined type") {
		t.Errorf("expected undefined type error, got: %v", errors[0])
	}
}

func TestListOperations(t *testing.T) {
	input := `func Test() int
    items := [1, 2, 3]
    first := items[0]
    slice := items[1:3]
    return first
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	if len(errors) > 0 {
		t.Fatalf("unexpected semantic errors: %v", errors)
	}
}

func TestBreakInsideSwitchIsAllowed(t *testing.T) {
	input := `func Route(command string)
    switch command
        when "quit"
            break
        otherwise
            print("ok")
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	if len(errors) > 0 {
		t.Fatalf("unexpected semantic errors: %v", errors)
	}
}

func TestConditionSwitchRequiresBoolWhenBranches(t *testing.T) {
	input := `func Bad()
    switch
        when 42
            print("bad")
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	if len(errors) == 0 {
		t.Fatal("expected semantic error for non-bool switch condition branch")
	}

	found := false
	for _, err := range errors {
		if strings.Contains(err.Error(), "switch condition branch must be bool") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected switch condition bool error, got: %v", errors)
	}
}

func TestBooleanExpression(t *testing.T) {
	input := `func Test(x int, y int) bool
    return x > 5 and y < 10
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	if len(errors) > 0 {
		t.Fatalf("unexpected semantic errors: %v", errors)
	}
}

func TestInvalidBooleanOperand(t *testing.T) {
	input := `func Test(x int) bool
    return x and 5
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	if len(errors) == 0 {
		t.Fatal("expected error for non-boolean operands to 'and'")
	}

	if !strings.Contains(errors[0].Error(), "logical operator requires boolean") {
		t.Errorf("expected boolean operator error, got: %v", errors[0])
	}
}

// TestConcurrentSemanticAnalysis tests that the semantic analyzer is thread-safe
// and multiple analyzers can run concurrently using synctest
func TestConcurrentSemanticAnalysis(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Test that semantic analyzer is thread-safe
		// Multiple analyzers should be able to run concurrently

		programs := []string{
			`func Add(a int, b int) int
    return a + b`,
			`type User
    name string
    age int`,
			`func Test() bool
    x := 5
    return x > 3`,
		}

		results := make(chan bool, len(programs))

		for _, src := range programs {
			go func(source string) {
				p, err := parser.New(source, "test.kuki")
				if err != nil {
					t.Errorf("parser error: %v", err)
					results <- false
					return
				}
				program, parseErrors := p.Parse()
				if len(parseErrors) > 0 {
					t.Errorf("parse errors: %v", parseErrors)
					results <- false
					return
				}
				analyzer := New(program)
				errors := analyzer.Analyze()
				if len(errors) > 0 {
					t.Errorf("semantic errors: %v", errors)
					results <- false
					return
				}
				results <- true
			}(src)
		}

		synctest.Wait()

		// Verify all completed successfully
		successCount := 0
		for range programs {
			select {
			case success := <-results:
				if success {
					successCount++
				}
			default:
				t.Error("Expected result not received")
			}
		}

		if successCount != len(programs) {
			t.Errorf("Expected %d successful analyses, got %d", len(programs), successCount)
		}
	})
}

func TestQualifiedTypes(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid qualified type in struct field",
			source: `
import "io"

type Writer
    output io.Writer
`,
			wantErr: false,
		},
		{
			name: "valid qualified type in function parameter",
			source: `
import "io"

func Write(w io.Writer, data string)
    return
`,
			wantErr: false,
		},
		{
			name: "valid qualified type in function return",
			source: `
import "io"

func GetWriter() io.Writer
    return empty
`,
			wantErr: false,
		},
		{
			name: "multiple qualified types",
			source: `
import "io"
import "bytes"

type Wrapper
    writer io.Writer
    reader io.Reader
    buffer bytes.Buffer
`,
			wantErr: false,
		},
		{
			name: "unimported package",
			source: `
type Writer
    output io.Writer
`,
			wantErr: true,
			errMsg:  "package 'io' not imported",
		},
		{
			name: "qualified type in list",
			source: `
import "io"

type Readers
    readers list of io.Reader
`,
			wantErr: false,
		},
		{
			name: "qualified type in map",
			source: `
import "io"

type WriterMap
    writers map of string to io.Writer
`,
			wantErr: false,
		},
		{
			name: "qualified type as pointer",
			source: `
import "bytes"

type BufferPtr
    buf reference bytes.Buffer
`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.New(tt.source, "test.kuki")
			if err != nil {
				t.Fatalf("parser error: %v", err)
			}

			program, parseErrors := p.Parse()
			if len(parseErrors) > 0 {
				t.Fatalf("parse errors: %v", parseErrors)
			}

			analyzer := New(program)
			errors := analyzer.Analyze()

			if tt.wantErr {
				if len(errors) == 0 {
					t.Fatalf("expected error containing '%s', but got no errors", tt.errMsg)
				}
				found := false
				for _, err := range errors {
					if strings.Contains(err.Error(), tt.errMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error containing '%s', got: %v", tt.errMsg, errors)
				}
			} else {
				if len(errors) > 0 {
					t.Errorf("expected no errors, got: %v", errors)
				}
			}
		})
	}
}

func TestVersionedPackageNameInference(t *testing.T) {
	tests := []struct {
		name        string
		importPath  string
		expectedPkg string
		source      string
	}{
		{
			name:        "slash-version suffix v2",
			importPath:  "encoding/json/v2",
			expectedPkg: "json",
			source: `import "encoding/json/v2"

type Config
    Name string

func main()
    cfg := Config{}
    cfg.Name = "test"
    data, _ := json.Marshal(cfg)
`,
		},
		{
			name:        "slash-version suffix v3",
			importPath:  "google.golang.org/protobuf/v3",
			expectedPkg: "protobuf",
			source: `import "google.golang.org/protobuf/v3"

func main()
    protobuf.NewMessage()
`,
		},
		{
			name:        "slash-version suffix v10",
			importPath:  "example.com/pkg/v10",
			expectedPkg: "pkg",
			source: `import "example.com/pkg/v10"

func main()
    pkg.DoSomething()
`,
		},
		{
			name:        "dot-version suffix (gopkg.in style)",
			importPath:  "gopkg.in/yaml.v3",
			expectedPkg: "yaml",
			source: `import "gopkg.in/yaml.v3"

type Data
    Value string

func main()
    d := Data{}
    yaml.Marshal(d)
`,
		},
		{
			name:        "no version suffix",
			importPath:  "encoding/json",
			expectedPkg: "json",
			source: `import "encoding/json"

type Data
    Value string

func main()
    d := Data{}
    json.Marshal(d)
`,
		},
		{
			name:        "package named vendor (not a version)",
			importPath:  "github.com/company/vendor",
			expectedPkg: "vendor",
			source: `import "github.com/company/vendor"

func main()
    vendor.DoSomething()
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.New(tt.source, "test.kuki")
			if err != nil {
				t.Fatalf("parser error: %v", err)
			}

			program, parseErrors := p.Parse()
			if len(parseErrors) > 0 {
				t.Fatalf("parse errors: %v", parseErrors)
			}

			analyzer := New(program)
			errors := analyzer.Analyze()

			// We expect no errors because the package name should be inferred correctly
			// and the functions should be resolved
			if len(errors) > 0 {
				t.Errorf("expected no errors for package %s (inferred as %s), got: %v", tt.importPath, tt.expectedPkg, errors)
			}

			// Verify the package was added to the symbol table with the correct name
			pkgSymbol := analyzer.symbolTable.Resolve(tt.expectedPkg)
			if pkgSymbol == nil {
				t.Errorf("expected package %s to be in symbol table, but it wasn't found", tt.expectedPkg)
			}
			if pkgSymbol != nil && pkgSymbol.Kind != SymbolVariable {
				t.Errorf("expected symbol %s to be a variable (imports are stored as variables), got kind: %v", tt.expectedPkg, pkgSymbol.Kind)
			}
		})
	}
}

func TestPipeMultiValueReturn(t *testing.T) {
	// Test the fix for: "Semantic limit on multi-value pipe return"
	// This should now work: return x |> f() where f() returns (T, error)
	input := `func Test() (int, error)
    return 42 |> someFunc()

func someFunc(x int) (int, error)
    return x + 1, empty
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	if len(errors) > 0 {
		t.Fatalf("expected no semantic errors for pipe multi-value return, got: %v", errors)
	}
}

func TestPipeMultiValueReturnTypeMismatch(t *testing.T) {
	// Test that type checking still works with pipe multi-value returns
	input := `func Test() (string, error)
    return 42 |> someFunc()

func someFunc(x int) (int, error)
    return x + 1, empty
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	if len(errors) == 0 {
		t.Fatal("expected type mismatch error for incompatible pipe return")
	}

	if !strings.Contains(errors[0].Error(), "cannot return") {
		t.Errorf("expected type mismatch error, got: %v", errors[0])
	}
}

// ============================================================================
// Skill declaration semantic tests
// ============================================================================

func TestSkillDeclValid(t *testing.T) {
	input := `petiole weather

skill WeatherService
    description: "Provides weather data."
    version: "1.0.0"

func GetForecast(city string) string
    return city
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	if len(errors) > 0 {
		t.Fatalf("expected no errors, got: %v", errors)
	}
}

func TestSkillDeclWithoutPetiole(t *testing.T) {
	input := `skill WeatherService
    description: "Provides weather data."
    version: "1.0.0"
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	found := false
	for _, e := range errors {
		if strings.Contains(e.Error(), "requires a petiole") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 'requires a petiole' error, got: %v", errors)
	}
}

func TestSkillDeclLowercaseName(t *testing.T) {
	input := `petiole myskill

skill weatherService
    description: "Provides weather data."
    version: "1.0.0"
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	found := false
	for _, e := range errors {
		if strings.Contains(e.Error(), "must be exported") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 'must be exported' error, got: %v", errors)
	}
}

func TestSkillDeclEmptyDescription(t *testing.T) {
	input := `petiole myskill

skill MySkill
    version: "1.0.0"
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	found := false
	for _, e := range errors {
		if strings.Contains(e.Error(), "should have a description") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 'should have a description' error, got: %v", errors)
	}
}

func TestSkillDeclBadSemver(t *testing.T) {
	input := `petiole myskill

skill MySkill
    description: "A skill."
    version: "not-a-version"
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	found := false
	for _, e := range errors {
		if strings.Contains(e.Error(), "should follow semver") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 'should follow semver' error, got: %v", errors)
	}
}

func TestOnerrBlockErrInterpolationIsError(t *testing.T) {
	input := `func readFile(path string) (string, error)
    return path, empty

func Process(path string) (string, error)
    data := readFile(path) onerr
        return "", error "{err}"
    return data, empty
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	found := false
	for _, e := range errors {
		if strings.Contains(e.Error(), "use {error} not {err} inside onerr") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 'use {error} not {err} inside onerr' error, got: %v", errors)
	}
}

func TestOnerrInlineErrInterpolationIsError(t *testing.T) {
	input := `func readFile(path string) (string, error)
    return path, empty

func Process(path string) (string, error)
    data := readFile(path) onerr return "", error "{err}"
    return data, empty
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	found := false
	for _, e := range errors {
		if strings.Contains(e.Error(), "use {error} not {err} inside onerr") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 'use {error} not {err} inside onerr' error, got: %v", errors)
	}
}

func TestOnerrErrorInterpolationIsValid(t *testing.T) {
	input := `func readFile(path string) (string, error)
    return path, empty

func Process(path string) (string, error)
    data := readFile(path) onerr return "", error "{error}"
    return data, empty
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	for _, e := range errors {
		if strings.Contains(e.Error(), "use {error} not {err} inside onerr") {
			t.Fatalf("unexpected onerr interpolation error: %v", e)
		}
	}
}
