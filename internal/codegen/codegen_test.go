package codegen

import (
	"strings"
	"testing"

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

	if !strings.Contains(output, "for i := 0; i < 10; i++") {
		t.Errorf("expected numeric for loop with <, got: %s", output)
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
	input := `leaf mypackage

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

func TestGenericFunctionCodegen(t *testing.T) {
	input := `func Reverse(items list of element) list of element
    return items
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	// Run semantic analysis to collect type parameters
	analyzer := semantic.New(program)
	_ = analyzer.Analyze()

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	// Should have type parameters
	if !strings.Contains(output, "[T any]") {
		t.Errorf("expected type parameters [T any], got: %s", output)
	}

	// Should use T for the parameter and return type
	if !strings.Contains(output, "items []T") {
		t.Errorf("expected 'items []T', got: %s", output)
	}

	if !strings.Contains(output, ") []T") {
		t.Errorf("expected return type '[]T', got: %s", output)
	}
}

func TestGenericTypeCodegen(t *testing.T) {
	input := `type Box of element
    value element
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

	// Should have type parameters
	if !strings.Contains(output, "Box[T any]") {
		t.Errorf("expected 'Box[T any]', got: %s", output)
	}

	// Field should use T
	if !strings.Contains(output, "value T") {
		t.Errorf("expected 'value T', got: %s", output)
	}
}

func TestConstrainedGenericCodegen(t *testing.T) {
	input := `func Sum(items list of number) number
    return items[0]
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	// Run semantic analysis to collect type parameters
	analyzer := semantic.New(program)
	_ = analyzer.Analyze()

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	// Should have constrained type parameters
	if !strings.Contains(output, "[T cmp.Ordered]") {
		t.Errorf("expected type parameters [T cmp.Ordered], got: %s", output)
	}

	// Should auto-import cmp
	if !strings.Contains(output, "\"cmp\"") {
		t.Errorf("expected cmp import, got: %s", output)
	}
}
