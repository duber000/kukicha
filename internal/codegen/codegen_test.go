package codegen

import (
	"strings"
	"testing"
	"testing/synctest"

	"github.com/duber000/kukicha/internal/parser"
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

// REMOVED: Old generics tests - generics syntax has been removed from Kukicha
// Generic functionality is now provided by the stdlib (written in Go) with special transpilation
// See stdlib/iter/ for examples of special transpilation

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
