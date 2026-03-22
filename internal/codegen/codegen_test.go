package codegen

import (
	"github.com/duber000/kukicha/internal/parser"
	"strings"
	"testing"
)

func TestSimpleFunction(t *testing.T) {
	input := `func Add(a int, b int) int
    return a + b
`

	output := generateSource(t, input)

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

	output := generateSource(t, input)

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

	output := generateSource(t, input)

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
			t.Parallel()
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

	output := generateSource(t, input)

	if !strings.Contains(output, "[]int") {
		t.Errorf("expected slice type, got: %s", output)
	}
}

func TestMapType(t *testing.T) {
	input := `func GetMap() map of string to int
    return empty map of string to int
`

	output := generateSource(t, input)

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

	output := generateSource(t, input)

	if !strings.Contains(output, "for _, item := range items") {
		t.Errorf("expected for range loop, got: %s", output)
	}
}

func TestNumericForLoop(t *testing.T) {
	input := `func Test()
    for i from 0 to 10
        x := i
`

	output := generateSource(t, input)

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

	output := generateSource(t, input)

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

	output := generateSource(t, input)

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

	output := generateSource(t, input)

	if !strings.Contains(output, "return 0") {
		t.Errorf("expected zero int return, got: %s", output)
	}
	if !strings.Contains(output, "var _zero0 time.Duration") || !strings.Contains(output, "return _zero0") {
		t.Errorf("expected var _zero0 time.Duration + return _zero0 for named type zero value, got: %s", output)
	}
}

func TestIfElse(t *testing.T) {
	input := `func Max(a int, b int) int
    if a > b
        return a
    else
        return b
`

	output := generateSource(t, input)

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

	output := generateSource(t, input)

	if !strings.Contains(output, "&&") {
		t.Errorf("expected && operator, got: %s", output)
	}

	if !strings.Contains(output, "||") {
		t.Errorf("expected || operator, got: %s", output)
	}
}

func TestReferenceType(t *testing.T) {
	input := `type Person
    Name string
`

	output := generateSource(t, input)

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

	output := generateSource(t, input)

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

	output := generateSource(t, input)

	if !strings.Contains(output, "\"fmt\"") {
		t.Errorf("expected fmt import, got: %s", output)
	}

	if !strings.Contains(output, "str \"strings\"") {
		t.Errorf("expected aliased strings import, got: %s", output)
	}
}

func TestVariadicCodegen(t *testing.T) {
	input := `func Print(many values)
    return values
`

	output := generateSource(t, input)

	if !strings.Contains(output, "values ...interface{}") {
		t.Errorf("expected variadic syntax, got: %s", output)
	}
}

func TestTypedVariadicCodegen(t *testing.T) {
	input := `func Sum(many numbers int) int
    return 0
`

	output := generateSource(t, input)

	if !strings.Contains(output, "numbers ...int") {
		t.Errorf("expected typed variadic syntax, got: %s", output)
	}
}

func TestFloatLiteralPrecision(t *testing.T) {
	input := `func Main()
    x := 0.000000001
    y := 3.14159265358979
`

	output := generateSource(t, input)

	if !strings.Contains(output, "0.000000001") {
		t.Errorf("expected float 0.000000001 to be preserved, got: %s", output)
	}
	if !strings.Contains(output, "3.14159265358979") {
		t.Errorf("expected float 3.14159265358979 to be preserved, got: %s", output)
	}
}

func TestStdlibDefaultParametersApplied(t *testing.T) {
	input := `import "stdlib/string"

func Format(s string) string
    return string.PadRight(s, 10)
`
	output := generateSource(t, input)

	// Should fill in the default " " for padChar
	if !strings.Contains(output, `kukistring.PadRight(s, 10, " ")`) {
		t.Errorf("expected default padChar to be filled in, got: %s", output)
	}
}

func TestStdlibDefaultParametersExplicitOverride(t *testing.T) {
	input := `import "stdlib/string"

func Format(s string) string
    return string.PadRight(s, 10, "-")
`
	output := generateSource(t, input)

	// Explicit override should be preserved
	if !strings.Contains(output, `kukistring.PadRight(s, 10, "-")`) {
		t.Errorf("expected explicit padChar override, got: %s", output)
	}
}

func TestStdlibDefaultParametersInPipe(t *testing.T) {
	input := `import "stdlib/string"

func Format(s string) string
    return s |> string.PadRight(10)
`
	output := generateSource(t, input)

	// Should fill in the default " " for padChar even through a pipe
	if !strings.Contains(output, `kukistring.PadRight(s, 10, " ")`) {
		t.Errorf("expected default padChar to be filled in via pipe, got: %s", output)
	}
}

func TestOnErrContinue(t *testing.T) {
	input := `func Process(items list of string)
    for item in items
        v := doWork(item) onerr continue
        print(v)
`
	output := generateSource(t, input)

	if !strings.Contains(output, "continue") {
		t.Errorf("expected 'continue' in output, got: %s", output)
	}
	if !strings.Contains(output, "!= nil") {
		t.Errorf("expected error check, got: %s", output)
	}
}

func TestPipeInStringInterpolation(t *testing.T) {
	input := `func Show(name string) string
    return "{name |> toUpper()}"
`
	output := generateSource(t, input)

	// Should NOT contain "|>" in the generated Go
	if strings.Contains(output, "|>") {
		t.Errorf("pipe operator should not appear in generated Go, got: %s", output)
	}
	// Should contain the pipe result as a function call
	if !strings.Contains(output, "toUpper(name)") {
		t.Errorf("expected pipe to be converted to function call, got: %s", output)
	}
}

func TestMultiPipeInStringInterpolation(t *testing.T) {
	input := `func Show(name string) string
    return "{name |> trim() |> toUpper()}"
`
	output := generateSource(t, input)

	if strings.Contains(output, "|>") {
		t.Errorf("pipe operator should not appear in generated Go, got: %s", output)
	}
	if !strings.Contains(output, "toUpper(trim(name))") {
		t.Errorf("expected nested pipe calls, got: %s", output)
	}
}

func TestOnErrBreak(t *testing.T) {
	input := `func Process(items list of string)
    for item in items
        v := doWork(item) onerr break
        print(v)
`
	output := generateSource(t, input)

	if !strings.Contains(output, "break") {
		t.Errorf("expected 'break' in output, got: %s", output)
	}
	if !strings.Contains(output, "!= nil") {
		t.Errorf("expected error check, got: %s", output)
	}
}
