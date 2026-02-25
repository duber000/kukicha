package codegen

import (
	"strings"
	"testing"
	"testing/synctest"

	"github.com/duber000/kukicha/internal/parser"
	"github.com/duber000/kukicha/internal/semantic"
)

func TestSimpleFunction(t *testing.T) {
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

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "func Add(a int, b int) int") {
		t.Errorf("expected function signature, got: %s", output)
	}

	if !strings.Contains(output, "return (a + b)") {
		t.Errorf("expected return statement, got: %s", output)
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

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "type Person struct") {
		t.Errorf("expected struct declaration, got: %s", output)
	}

	if !strings.Contains(output, "Name string") {
		t.Errorf("expected Name field, got: %s", output)
	}

	if !strings.Contains(output, "Age int") {
		t.Errorf("expected Age field, got: %s", output)
	}
}

func TestTypeDeclarationFieldAliasGeneratesJSONTag(t *testing.T) {
	input := `type Repo
    Stars int as "stargazers_count"
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "Stars int `json:\"stargazers_count\"`") {
		t.Errorf("expected generated json tag from field alias, got: %s", output)
	}
}

func TestFunctionTypeAlias(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic func type alias",
			input:    "type Handler func(string)\n",
			expected: "type Handler func(string)",
		},
		{
			name:     "func type alias with return",
			input:    "type Transform func(string) string\n",
			expected: "type Transform func(string) string",
		},
		{
			name:     "func type alias with multiple returns",
			input:    "type Callback func(string, int) (bool, error)\n",
			expected: "type Callback func(string, int) (bool, error)",
		},
		{
			name:     "func type alias no params",
			input:    "type Factory func() error\n",
			expected: "type Factory func() error",
		},
		{
			name:     "func type alias with map param",
			input:    "type ToolHandler func(map of string to any) (any, error)\n",
			expected: "type ToolHandler func(map[string]any) (any, error)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.New(tt.input, "test.kuki")
			if err != nil {
				t.Fatalf("parser error: %v", err)
			}

			program, parseErrors := p.Parse()
			if len(parseErrors) > 0 {
				t.Fatalf("parse errors: %v", parseErrors)
			}

			gen := New(program)
			output, err := gen.Generate()
			if err != nil {
				t.Fatalf("codegen error: %v", err)
			}

			if !strings.Contains(output, tt.expected) {
				t.Errorf("expected output to contain %q, got:\n%s", tt.expected, output)
			}

			// Ensure it's NOT a struct
			if strings.Contains(output, "struct") {
				t.Errorf("function type alias should not generate struct, got:\n%s", output)
			}
		})
	}
}

func TestListType(t *testing.T) {
	input := `func GetItems() list of int
    return [1, 2, 3]
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "[]int") {
		t.Errorf("expected slice type, got: %s", output)
	}
}

func TestMapType(t *testing.T) {
	input := `func GetMap() map of string to int
    return empty map of string to int
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "map[string]int") {
		t.Errorf("expected map type, got: %s", output)
	}
}

func TestForLoop(t *testing.T) {
	input := `func Sum(items list of int) int
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

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "for _, item := range items") {
		t.Errorf("expected for range loop, got: %s", output)
	}
}

func TestNumericForLoop(t *testing.T) {
	input := `func Test()
    for i from 0 to 10
        x := i
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "for i := range 10") {
		t.Errorf("expected range-over-int for loop, got: %s", output)
	}
}

func TestSwitchStatement(t *testing.T) {
	input := `func Route(command string) string
    switch command
        when "fetch", "pull"
            return "ok"
        otherwise
            return "unknown"
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "switch command {") {
		t.Errorf("expected value switch, got: %s", output)
	}
	if !strings.Contains(output, "case \"fetch\", \"pull\":") {
		t.Errorf("expected case values, got: %s", output)
	}
	if !strings.Contains(output, "default:") {
		t.Errorf("expected default branch, got: %s", output)
	}
}

func TestConditionSwitchStatement(t *testing.T) {
	input := `func Label(stars int) string
    switch
        when stars >= 100
            return "hot"
        otherwise
            return "new"
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "switch {") {
		t.Errorf("expected condition switch form, got: %s", output)
	}
	if !strings.Contains(output, "case (stars >= 100):") {
		t.Errorf("expected condition branch, got: %s", output)
	}
}

func TestEmptyTypedZeroValues(t *testing.T) {
	input := `func ZeroInt() int
    return empty int

func ZeroDuration() time.Duration
    return empty time.Duration
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "return 0") {
		t.Errorf("expected zero int return, got: %s", output)
	}
	if !strings.Contains(output, "return *new(time.Duration)") {
		t.Errorf("expected zero time.Duration return, got: %s", output)
	}
}

func TestOnErrErrorMessageUsesProvidedString(t *testing.T) {
	input := `func bar() (int, error)
    return 0, error "nope"

func Foo() (int, error)
    x := 0
    x = bar() onerr error "bad: {error}"
    return x, empty
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "errors.New") || !strings.Contains(output, "bad: %v") {
		t.Errorf("expected errors.New with interpolated message, got: %s", output)
	}

	if !strings.Contains(output, `"fmt"`) {
		t.Errorf("expected auto-import of fmt for interpolated error string, got: %s", output)
	}
}

func TestErrorInterpolationAutoImportsFmt(t *testing.T) {
	// Standalone error "..." with interpolation in a return statement
	// should auto-import fmt without the user needing to add import "fmt"
	input := `func GetEnv(key string) (string, error)
    return "", error "environment variable {key} not set"
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	t.Logf("Generated output:\n%s", output)

	if !strings.Contains(output, `"fmt"`) {
		t.Errorf("expected auto-import of fmt for interpolated error in return, got: %s", output)
	}

	if !strings.Contains(output, "fmt.Sprintf") {
		t.Errorf("expected fmt.Sprintf for interpolated error message, got: %s", output)
	}

	if !strings.Contains(output, `"errors"`) {
		t.Errorf("expected auto-import of errors, got: %s", output)
	}
}

func TestErrorNoInterpolationNoFmtImport(t *testing.T) {
	// error "..." WITHOUT interpolation should NOT auto-import fmt
	input := `func Fail() error
    return error "something went wrong"
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	t.Logf("Generated output:\n%s", output)

	if strings.Contains(output, `"fmt"`) {
		t.Errorf("should NOT import fmt when error message has no interpolation, got: %s", output)
	}

	if !strings.Contains(output, `"errors"`) {
		t.Errorf("expected auto-import of errors, got: %s", output)
	}
}

func TestOnErrDiscardStatementUsesReturnCount(t *testing.T) {
	input := `func One() error
    return empty

func Two() (int, error)
    return 0, empty

func Use()
    One() onerr discard
    Two() onerr discard
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "_ = One()") {
		t.Errorf("expected single-value discard, got: %s", output)
	}
	if !strings.Contains(output, "_, _ = Two()") {
		t.Errorf("expected two-value discard, got: %s", output)
	}
}

func TestOnErrDiscardUsesSemanticReturnCounts(t *testing.T) {
	input := `import "os"

func Use()
    os.LookupEnv("X") onerr discard
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := semantic.New(program)
	semanticErrors := analyzer.Analyze()
	if len(semanticErrors) > 0 {
		t.Fatalf("semantic errors: %v", semanticErrors)
	}

	gen := New(program)
	gen.SetExprReturnCounts(analyzer.ReturnCounts())
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "_, _ = os.LookupEnv(\"X\")") {
		t.Errorf("expected two-value discard using semantic return counts, got: %s", output)
	}
}

func TestOnErrDiscardMethodCall(t *testing.T) {
	input := `type Writer
    path string

func Write on w Writer error
    return empty

func Use()
    w := Writer{path: "out.txt"}
    w.Write() onerr discard
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := semantic.New(program)
	semanticErrors := analyzer.Analyze()
	if len(semanticErrors) > 0 {
		t.Fatalf("semantic errors: %v", semanticErrors)
	}

	gen := New(program)
	gen.SetExprReturnCounts(analyzer.ReturnCounts())
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "_ = w.Write()") {
		t.Errorf("expected single-value discard for method call, got: %s", output)
	}
}

func TestNumericForLoopThrough(t *testing.T) {
	input := `func Test()
    for i from 0 through 10
        x := i
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "for i := 0; i <= 10; i++") {
		t.Errorf("expected numeric for loop with <=, got: %s", output)
	}
}

func TestIfElse(t *testing.T) {
	input := `func Max(a int, b int) int
    if a > b
        return a
    else
        return b
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "if (a > b)") {
		t.Errorf("expected if statement, got: %s", output)
	}

	if !strings.Contains(output, "} else {") {
		t.Errorf("expected else clause, got: %s", output)
	}
}

func TestBooleanOperators(t *testing.T) {
	input := `func Test(a bool, b bool) bool
    return a and b or a
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "&&") {
		t.Errorf("expected && operator, got: %s", output)
	}

	if !strings.Contains(output, "||") {
		t.Errorf("expected || operator, got: %s", output)
	}
}

func TestStringInterpolation(t *testing.T) {
	input := `func Greet(name string) string
    return "Hello {name}!"
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	t.Logf("Generated output:\n%s", output)

	if !strings.Contains(output, "fmt.Sprintf") {
		t.Errorf("expected fmt.Sprintf for string interpolation, got: %s", output)
	}

	if !strings.Contains(output, "\"fmt\"") {
		t.Errorf("expected fmt import, got: %s", output)
	}
}

func TestPrintfStyleMethodInterpolation(t *testing.T) {
	// Test that t.Errorf with interpolation generates inline format args (not fmt.Sprintf)
	input := `import "testing"

func TestExample(t reference testing.T)
    name := "world"
    t.Errorf("hello {name}")
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	t.Logf("Generated output:\n%s", output)

	// Should generate: t.Errorf("hello %v", name)
	if !strings.Contains(output, `t.Errorf("hello %v", name)`) {
		t.Errorf("expected inline format args for t.Errorf, got: %s", output)
	}

	// Should NOT use fmt.Sprintf for the error message
	if strings.Contains(output, "fmt.Sprintf") {
		t.Errorf("should not use fmt.Sprintf for printf-style methods, got: %s", output)
	}

	// Should NOT import fmt (no other interpolation uses it)
	if strings.Contains(output, `"fmt"`) {
		t.Errorf("should not import fmt when only printf-style interpolation is used, got: %s", output)
	}
}

func TestStringInterpolationInSwitchAddsFmtImport(t *testing.T) {
	input := `func main()
    v := 1
    switch v
        when 1
            panic "bad {v}"
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "fmt.Sprintf") {
		t.Errorf("expected fmt.Sprintf in switch body interpolation, got: %s", output)
	}

	if !strings.Contains(output, `"fmt"`) {
		t.Errorf("expected fmt import for switch body interpolation, got: %s", output)
	}
}

func TestStringInterpolationInOnErrBlockAddsFmtImport(t *testing.T) {
	input := `func main()
    data, err := os.ReadFile("foo.txt")
    _ = data
    _ = err
    result := string(data) onerr
        panic "read failed: {error}"
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, genErr := gen.Generate()
	if genErr != nil {
		t.Fatalf("codegen error: %v", genErr)
	}

	if !strings.Contains(output, `"fmt"`) {
		t.Errorf("expected fmt import for onerr block interpolation, got: %s", output)
	}
}

func TestReferenceType(t *testing.T) {
	input := `type Person
    Name string
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	// Just verify that types are generated correctly
	if !strings.Contains(output, "type Person struct") {
		t.Errorf("expected struct type, got: %s", output)
	}
}

func TestPackageDeclaration(t *testing.T) {
	input := `petiole mypackage

func Test()
    x := 1
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "package mypackage") {
		t.Errorf("expected output to contain 'package mypackage', got: %s", output)
	}
}

func TestImports(t *testing.T) {
	input := `import "fmt"
import "strings" as str

func Test()
    x := 1
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "\"fmt\"") {
		t.Errorf("expected fmt import, got: %s", output)
	}

	if !strings.Contains(output, "str \"strings\"") {
		t.Errorf("expected aliased strings import, got: %s", output)
	}
}

// Tests for new generic features

func TestVariadicCodegen(t *testing.T) {
	input := `func Print(many values)
    return values
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "values ...interface{}") {
		t.Errorf("expected variadic syntax, got: %s", output)
	}
}

func TestTypedVariadicCodegen(t *testing.T) {
	input := `func Sum(many numbers int) int
    return 0
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "numbers ...int") {
		t.Errorf("expected typed variadic syntax, got: %s", output)
	}
}

func TestAddressOfExpr(t *testing.T) {
	input := `func GetUserPtr(user User) reference User
    return reference of user
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "return &user") {
		t.Errorf("expected '&user', got: %s", output)
	}
}

func TestDerefExpr(t *testing.T) {
	input := `func GetUserValue(userPtr reference User) User
    return dereference userPtr
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "return *userPtr") {
		t.Errorf("expected '*userPtr', got: %s", output)
	}
}

func TestDerefAssignment(t *testing.T) {
	input := `func SwapValues(a reference int, b reference int)
    temp := dereference a
    dereference a = dereference b
    dereference b = temp
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "*a = *b") {
		t.Errorf("expected '*a = *b', got: %s", output)
	}
}

func TestAddressOfWithFieldAccess(t *testing.T) {
	input := `func ScanField(row Row, field reference string)
    row.Scan(reference of field)
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "&field") {
		t.Errorf("expected '&field', got: %s", output)
	}
}

func TestAddressOfCallExprUsesNew(t *testing.T) {
	// Go 1.26 new(expr): function call returns are non-addressable,
	// so `reference of someFunc()` should generate `new(someFunc())`.
	input := `func GetName() string
    return "alice"

func main()
    ptr := reference of GetName()
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "new(GetName())") {
		t.Errorf("expected 'new(GetName())', got: %s", output)
	}
}

func TestAddressOfVariableStillUsesAmpersand(t *testing.T) {
	// Plain variables are addressable, so `reference of x` stays as `&x`.
	input := `func Ptr(x int) reference int
    return reference of x
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "&x") {
		t.Errorf("expected '&x', got: %s", output)
	}
	if strings.Contains(output, "new(x)") {
		t.Errorf("should not use new() for addressable variable, got: %s", output)
	}
}

// REMOVED: Old generics tests - generics syntax has been removed from Kukicha
// Generic functionality is now provided by the stdlib (written in Go) with special transpilation
// See stdlib/iterator/ for examples of special transpilation

// TestConcurrentCodeGeneration tests that multiple code generators can run
// concurrently without data races or interference using synctest
func TestConcurrentCodeGeneration(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Test that multiple code generators can run concurrently
		// without data races or interference

		programs := []string{
			`func main()
    x := 1`,
			`func add(a int, b int) int
    return a + b`,
			`type User
    name string`,
		}

		results := make(chan string, len(programs))

		for _, src := range programs {
			go func(source string) {
				p, err := parser.New(source, "test.kuki")
				if err != nil {
					t.Errorf("parser error: %v", err)
					results <- ""
					return
				}
				program, parseErrors := p.Parse()
				if len(parseErrors) > 0 {
					t.Errorf("parse errors: %v", parseErrors)
					results <- ""
					return
				}
				gen := New(program)
				code, err := gen.Generate()
				if err != nil {
					t.Errorf("codegen error: %v", err)
					results <- ""
					return
				}
				results <- code
			}(src)
		}

		synctest.Wait()

		// Verify all completed
		for range programs {
			select {
			case result := <-results:
				if result == "" {
					t.Error("Expected non-empty result")
				}
			default:
				t.Error("Expected result not received")
			}
		}
	})
}

func TestGroupByGenerics(t *testing.T) {
	input := `petiole slice

func GroupBy(items list of any, keyFunc func(any) any2) map of any2 to list of any
    result := make(map of any2 to list of any)
    for item in items
        key := keyFunc(item)
        result[key] = append(result[key], item)
    return result
`

	p, err := parser.New(input, "stdlib/slice/slice.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	gen.SetSourceFile("stdlib/slice/slice.kuki")
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	// Verify generic type parameters are generated
	if !strings.Contains(output, "func GroupBy[T any, K comparable]") {
		t.Errorf("expected generic function signature with [T any, K comparable], got: %s", output)
	}

	// Verify the parameter signature
	if !strings.Contains(output, "(items []T, keyFunc func(T) K)") {
		t.Errorf("expected correct parameter types, got: %s", output)
	}

	// Verify return type
	if !strings.Contains(output, "map[K][]T") {
		t.Errorf("expected return type map[K][]T, got: %s", output)
	}
}

func TestGroupByFunction(t *testing.T) {
	input := `func GroupBy(items list of any, keyFunc func(any) any2) map of any2 to list of any
    result := make(map of any2 to list of any)
    for item in items
        key := keyFunc(item)
        result[key] = append(result[key], item)
    return result
`

	p, err := parser.New(input, "stdlib/slice/slice.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	gen.SetSourceFile("stdlib/slice/slice.kuki")
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	// Verify function creates the result map properly
	if !strings.Contains(output, "result := make(map[K][]T)") {
		t.Errorf("expected make(map[K][]T), got: %s", output)
	}

	// Verify append is called correctly
	if !strings.Contains(output, "result[key] = append(result[key], item)") {
		t.Errorf("expected append to result[key], got: %s", output)
	}
}

func TestFetchJsonGenerics(t *testing.T) {
	input := `petiole fetch

func Json(resp reference http.Response, sample any) (any, error)
    data := sample
    return data, empty
`

	p, err := parser.New(input, "stdlib/fetch/fetch.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	gen.SetSourceFile("stdlib/fetch/fetch.kuki")
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "func Json[T any](resp *http.Response, sample T) (T, error)") {
		t.Errorf("expected fetch.Json generic signature, got: %s", output)
	}
}

func TestJSONDecodeReadGenerics(t *testing.T) {
	input := `petiole json

func DecodeRead(reader io.Reader, sample any) (any, error)
    data := sample
    return data, empty
`

	p, err := parser.New(input, "stdlib/json/json.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	gen.SetSourceFile("stdlib/json/json.kuki")
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "func DecodeRead[T any](reader io.Reader, sample T) (T, error)") {
		t.Errorf("expected json.DecodeRead generic signature, got: %s", output)
	}
}

func TestStdlibImportRewriting(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		expectedImport string
		shouldContain  string
	}{
		{
			name: "stdlib/json import",
			source: `import "stdlib/json"

type Config
    Name string

func main()
    cfg := Config{}
    cfg.Name = "test"
    data, _ := json.Marshal(cfg)
`,
			expectedImport: `"github.com/duber000/kukicha/stdlib/json"`,
			shouldContain:  "json.Marshal",
		},
		{
			name: "stdlib/fetch import",
			source: `import "stdlib/fetch"

func main()
    req := fetch.New("https://example.com")
`,
			expectedImport: `"github.com/duber000/kukicha/stdlib/fetch"`,
			shouldContain:  "fetch.New",
		},
		{
			name: "stdlib/json with alias",
			source: `import "stdlib/json" as j

type Data
    Value string

func main()
    d := Data{}
    j.Marshal(d)
`,
			expectedImport: `j "github.com/duber000/kukicha/stdlib/json"`,
			shouldContain:  "j.Marshal",
		},
		{
			name: "multiple imports with stdlib",
			source: `import "fmt"
import "stdlib/json"

type User
    Name string

func main()
    u := User{}
    data, _ := json.Marshal(u)
    fmt.Println(data)
`,
			expectedImport: `"github.com/duber000/kukicha/stdlib/json"`,
			shouldContain:  "json.Marshal",
		},
		{
			name: "non-stdlib import unchanged",
			source: `import "encoding/json"

type Data
    Value string

func main()
    d := Data{}
    json.Marshal(d)
`,
			expectedImport: `"encoding/json"`,
			shouldContain:  "json.Marshal",
		},
		{
			name: "version suffix import gets alias",
			source: `import "encoding/json/v2"

type Data
    Value string

func main()
    d := Data{}
    json.Marshal(d)
`,
			expectedImport: `json "encoding/json/v2"`,
			shouldContain:  "json.Marshal",
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

			gen := New(program)
			output, err := gen.Generate()
			if err != nil {
				t.Fatalf("codegen error: %v", err)
			}

			// Verify the import was rewritten correctly
			if !strings.Contains(output, tt.expectedImport) {
				t.Errorf("expected import %s in output, got: %s", tt.expectedImport, output)
			}

			// Verify the code using the import is present
			if !strings.Contains(output, tt.shouldContain) {
				t.Errorf("expected code %s in output, got: %s", tt.shouldContain, output)
			}
		})
	}
}

func TestMultiLinePipeCodegen(t *testing.T) {
	// Multi-line pipe must produce the same nested call as single-line.
	singleLine := `func Test() string
    return "hello" |> ToUpper() |> TrimSpace()
`
	multiLine := `func Test() string
    return "hello" |>
        ToUpper() |>
        TrimSpace()
`

	generate := func(src string) string {
		t.Helper()
		p, err := parser.New(src, "test.kuki")
		if err != nil {
			t.Fatalf("lexer error: %v", err)
		}
		program, parseErrors := p.Parse()
		if len(parseErrors) > 0 {
			t.Fatalf("parse errors: %v", parseErrors)
		}
		gen := New(program)
		output, err := gen.Generate()
		if err != nil {
			t.Fatalf("codegen error: %v", err)
		}
		return output
	}

	singleOut := generate(singleLine)
	multiOut := generate(multiLine)

	// Both must contain the fully-nested call.
	want := "TrimSpace(ToUpper(\"hello\"))"
	if !strings.Contains(singleOut, want) {
		t.Errorf("single-line output missing %q:\n%s", want, singleOut)
	}
	if !strings.Contains(multiOut, want) {
		t.Errorf("multi-line output missing %q:\n%s", want, multiOut)
	}
}

func TestOnErrPipeChainPreservesIntermediateErrors(t *testing.T) {
	input := `import "stdlib/fetch"

type Repo
    name string as "name"

func Load(url string) list of Repo
    repos := fetch.Get(url)
        |> fetch.CheckStatus()
        |> fetch.Json(empty list of Repo) onerr return empty list of Repo
    return repos
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := semantic.New(program)
	analyzer.Analyze()

	gen := New(program)
	gen.SetExprReturnCounts(analyzer.ReturnCounts())
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if strings.Contains(output, "val, _ :=") {
		t.Fatalf("expected no intermediate error discards in pipe onerr lowering, got: %s", output)
	}
	if !strings.Contains(output, "pipe_1, err_2 := fetch.Get(url)") {
		t.Fatalf("expected explicit fetch.Get error capture, got: %s", output)
	}
	if !strings.Contains(output, "pipe_3, err_4 := fetch.CheckStatus(pipe_1)") {
		t.Fatalf("expected explicit fetch.CheckStatus error capture, got: %s", output)
	}
	if !strings.Contains(output, "pipe_5, err_6 := fetch.Json(pipe_3, []Repo{})") {
		t.Fatalf("expected explicit fetch.Json error capture, got: %s", output)
	}
}

func TestOnErrPipeChainUserDefinedMultiReturn(t *testing.T) {
	input := `import "strconv"

func GetInput() string
    return "42"

func Parse(data string) (int, error)
    return strconv.Atoi(data)

func Run() int
    result := GetInput() |> Parse() onerr return 0
    return result
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := semantic.New(program)
	analyzer.Analyze()

	gen := New(program)
	gen.SetExprReturnCounts(analyzer.ReturnCounts())
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	// GetInput() returns 1 value, so no error capture
	if !strings.Contains(output, "pipe_1 := GetInput()") {
		t.Fatalf("expected simple assignment for GetInput(), got: %s", output)
	}
	// Parse() returns (int, error), so should capture error
	if !strings.Contains(output, "pipe_2, err_3 := Parse(pipe_1)") {
		t.Fatalf("expected error capture for user-defined Parse(), got: %s", output)
	}
}

func TestGoBlockSyntax(t *testing.T) {
	input := `func main()
    go
        doSomething()
        doSomethingElse()
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "go func() {") {
		t.Errorf("expected 'go func() {' in output, got: %s", output)
	}

	if !strings.Contains(output, "}()") {
		t.Errorf("expected '}()' in output, got: %s", output)
	}

	if !strings.Contains(output, "doSomething()") {
		t.Errorf("expected 'doSomething()' in output, got: %s", output)
	}
}

func TestGoCallSyntaxStillWorks(t *testing.T) {
	input := `func main()
    go processItem(item)
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "go processItem(item)") {
		t.Errorf("expected 'go processItem(item)' in output, got: %s", output)
	}
}

func TestArrowLambdaTypedExpression(t *testing.T) {
	input := `func main()
    f := (r Repo) => r.Stars > 100
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "func(r Repo) bool") {
		t.Errorf("expected 'func(r Repo) bool' in output, got: %s", output)
	}

	if !strings.Contains(output, "return (r.Stars > 100)") {
		t.Errorf("expected 'return (r.Stars > 100)' in output, got: %s", output)
	}
}

func TestArrowLambdaZeroParams(t *testing.T) {
	input := `func main()
    f := () => print("hello")
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "func()") {
		t.Errorf("expected 'func()' in output, got: %s", output)
	}
}

func TestArrowLambdaBlockForm(t *testing.T) {
	input := `func main()
    f := (r Repo) =>
        name := r.Name
        return name
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "func(r Repo)") {
		t.Errorf("expected 'func(r Repo)' in output, got: %s", output)
	}

	if !strings.Contains(output, "name := r.Name") {
		t.Errorf("expected 'name := r.Name' in output, got: %s", output)
	}

	if !strings.Contains(output, "return name") {
		t.Errorf("expected 'return name' in output, got: %s", output)
	}
}

// ============================================================================
// Skill codegen tests
// ============================================================================

func TestSkillComment(t *testing.T) {
	input := `petiole weather

skill WeatherService
    description: "Provides real-time weather data."
    version: "2.1.0"

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

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "// Skill: WeatherService") {
		t.Errorf("expected '// Skill: WeatherService' in output, got: %s", output)
	}

	if !strings.Contains(output, "// Description: Provides real-time weather data.") {
		t.Errorf("expected '// Description: Provides real-time weather data.' in output, got: %s", output)
	}

	if !strings.Contains(output, "// Version: 2.1.0") {
		t.Errorf("expected '// Version: 2.1.0' in output, got: %s", output)
	}
}

func TestSkillCommentNoFields(t *testing.T) {
	input := `petiole myskill

skill MySkill

func Hello() string
    return "hello"
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "// Skill: MySkill") {
		t.Errorf("expected '// Skill: MySkill' in output, got: %s", output)
	}

	// Should NOT contain Description or Version lines since they're empty
	if strings.Contains(output, "// Description:") {
		t.Errorf("did not expect '// Description:' in output when description is empty")
	}
	if strings.Contains(output, "// Version:") {
		t.Errorf("did not expect '// Version:' in output when version is empty")
	}
}

// ============================================================================
// onerr explain codegen tests
// ============================================================================

func TestOnErrExplainStandaloneCodegen(t *testing.T) {
	input := `func Test() (string, error)
    x := foo() onerr explain "City names must be capitalized"
    return x, nil

func foo() (string, error)
    return "hello", nil
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := semantic.New(program)
	semErrors := analyzer.Analyze()
	if len(semErrors) > 0 {
		t.Fatalf("semantic errors: %v", semErrors)
	}

	gen := New(program)
	gen.SetExprReturnCounts(analyzer.ReturnCounts())
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	// Should contain fmt.Errorf wrapping
	if !strings.Contains(output, `fmt.Errorf("City names must be capitalized: %w"`) {
		t.Errorf("expected fmt.Errorf wrapping in output, got: %s", output)
	}

	// Should import fmt
	if !strings.Contains(output, `"fmt"`) {
		t.Errorf("expected fmt import in output, got: %s", output)
	}
}

func TestOnErrExplainWithHandlerCodegen(t *testing.T) {
	input := `func Test()
    x := foo() onerr 0 explain "Expected a positive integer"

func foo() (int, error)
    return 42, nil
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := semantic.New(program)
	semErrors := analyzer.Analyze()
	if len(semErrors) > 0 {
		t.Fatalf("semantic errors: %v", semErrors)
	}

	gen := New(program)
	gen.SetExprReturnCounts(analyzer.ReturnCounts())
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	// Should contain fmt.Errorf wrapping
	if !strings.Contains(output, `fmt.Errorf("Expected a positive integer: %w"`) {
		t.Errorf("expected fmt.Errorf wrapping in output, got: %s", output)
	}

	// Should also assign the default value 0
	if !strings.Contains(output, "= 0") {
		t.Errorf("expected default value assignment '= 0' in output, got: %s", output)
	}
}

func TestTypeSwitchStatement(t *testing.T) {
	input := `func Handle(event any)
    switch event as e
        when reference a2a.TaskStatusUpdateEvent
            print(e.Status.State)
        when string
            print(e)
        otherwise
            print("unknown")
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "switch e := event.(type) {") {
		t.Errorf("expected type switch, got: %s", output)
	}
	if !strings.Contains(output, "case *a2a.TaskStatusUpdateEvent:") {
		t.Errorf("expected pointer type case, got: %s", output)
	}
	if !strings.Contains(output, "case string:") {
		t.Errorf("expected string type case, got: %s", output)
	}
	if !strings.Contains(output, "default:") {
		t.Errorf("expected default branch, got: %s", output)
	}
}

func TestTypeSwitchNoOtherwise(t *testing.T) {
	input := `func Handle(event any)
    switch event as e
        when int
            print(e)
        when string
            print(e)
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "switch e := event.(type) {") {
		t.Errorf("expected type switch, got: %s", output)
	}
	if !strings.Contains(output, "case int:") {
		t.Errorf("expected int type case, got: %s", output)
	}
	if !strings.Contains(output, "case string:") {
		t.Errorf("expected string type case, got: %s", output)
	}
	if strings.Contains(output, "default:") {
		t.Errorf("should not have default branch, got: %s", output)
	}
}

func TestSelectStatementCodegen(t *testing.T) {
	input := `func Run(ch channel of string, done channel of string, out channel of string)
    select
        when receive from done
            return
        when msg := receive from ch
            return
        when msg, ok := receive from ch
            return
        when send "ping" to out
            return
        otherwise
            return
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "select {") {
		t.Errorf("expected 'select {', got: %s", output)
	}
	if !strings.Contains(output, "case <-done:") {
		t.Errorf("expected bare receive case 'case <-done:', got: %s", output)
	}
	if !strings.Contains(output, "case msg := <-ch:") {
		t.Errorf("expected 1-var receive case 'case msg := <-ch:', got: %s", output)
	}
	if !strings.Contains(output, "case msg, ok := <-ch:") {
		t.Errorf("expected 2-var receive case 'case msg, ok := <-ch:', got: %s", output)
	}
	if !strings.Contains(output, "case out <- \"ping\":") {
		t.Errorf("expected send case 'case out <- \"ping\":', got: %s", output)
	}
	if !strings.Contains(output, "default:") {
		t.Errorf("expected default branch, got: %s", output)
	}
}

func TestExternalInterfaceTypeInFunctionSignature(t *testing.T) {
	input := `import "net/http"

func Wrap(handler http.Handler) http.Handler
    return handler
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "func Wrap(handler http.Handler) http.Handler") {
		t.Errorf("expected http.Handler in function signature, got: %s", output)
	}

	if !strings.Contains(output, "return handler") {
		t.Errorf("expected return handler, got: %s", output)
	}
}
