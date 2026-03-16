package codegen

import (
	"strings"
	"testing"
	"testing/synctest"

	"github.com/duber000/kukicha/internal/ast"
	"github.com/duber000/kukicha/internal/parser"
)

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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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

func TestZeroValueForType(t *testing.T) {
	prog := mustParseProgram(t, "func main()\n    x := 1\n")
	gen := New(prog)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"string param", `func f(s string)
    x := s
`, `""`},
		{"bool param", `func f(b bool)
    x := b
`, "false"},
		{"int param", `func f(n int)
    x := n
`, "0"},
		{"list param", `func f(items list of string)
    x := items
`, "[]string{}"},
		{"map param", `func f(m map of string to int)
    x := m
`, "map[string]int{}"},
		{"reference param", `func f(p reference int)
    x := p
`, "nil"},
		{"channel param", `func f(ch channel of string)
    x := ch
`, "nil"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := mustParseProgram(t, tt.input)
			// Find the function and get first param type
			for _, decl := range p.Declarations {
				if fn, ok := decl.(*ast.FunctionDecl); ok {
					if len(fn.Parameters) > 0 {
						got := gen.zeroValueForType(fn.Parameters[0].Type)
						if got != tt.expected {
							t.Errorf("zeroValueForType() = %q, want %q", got, tt.expected)
						}
					}
				}
			}
		})
	}
}

func TestIsLikelyInterfaceType(t *testing.T) {
	// Test with a program that declares a local interface
	input := `interface Storer
    Store(data string) error

func main()
    x := 1
`

	prog := mustParseProgram(t, input)
	gen := New(prog)

	tests := []struct {
		typeName string
		expected bool
	}{
		{"error", true},
		{"Storer", true},     // local interface
		{"io.Reader", true},  // Go stdlib interface
		{"http.Handler", true},
		{"User", false},      // unknown type, not interface
		{"int", false},
	}

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			got := gen.isLikelyInterfaceType(tt.typeName)
			if got != tt.expected {
				t.Errorf("isLikelyInterfaceType(%q) = %v, want %v", tt.typeName, got, tt.expected)
			}
		})
	}
}

func TestInferExprReturnType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "integer literal",
			input: `func f() int
    return 42
`,
			expected: "int",
		},
		{
			name: "string literal",
			input: `func f() string
    return "hello"
`,
			expected: "string",
		},
		{
			name: "bool literal",
			input: `func f() bool
    return true
`,
			expected: "bool",
		},
		{
			name: "float literal",
			input: `func f() float64
    return 3.14
`,
			expected: "float64",
		},
		{
			name: "comparison operator",
			input: `func f(a int, b int) bool
    return a > b
`,
			expected: "bool",
		},
		{
			name: "equality operator",
			input: `func f(a int, b int) bool
    return a equals b
`,
			expected: "bool",
		},
		{
			name: "logical not",
			input: `func f(b bool) bool
    return not b
`,
			expected: "bool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prog := mustParseProgram(t, tt.input)
			gen := New(prog)

			// Find the return expression
			for _, decl := range prog.Declarations {
				if fn, ok := decl.(*ast.FunctionDecl); ok && fn.Body != nil {
					for _, stmt := range fn.Body.Statements {
						if ret, ok := stmt.(*ast.ReturnStmt); ok && len(ret.Values) > 0 {
							got := gen.inferExprReturnType(ret.Values[0])
							if got != tt.expected {
								t.Errorf("inferExprReturnType() = %q, want %q", got, tt.expected)
							}
						}
					}
				}
			}
		})
	}
}

func TestTypeContainsPlaceholder(t *testing.T) {
	prog := mustParseProgram(t, "func f(items list of any)\n    x := items\n")
	gen := New(prog)

	for _, decl := range prog.Declarations {
		if fn, ok := decl.(*ast.FunctionDecl); ok {
			if len(fn.Parameters) > 0 {
				got := gen.typeContainsPlaceholder(fn.Parameters[0].Type, "any")
				if !got {
					t.Error("expected list of any to contain 'any' placeholder")
				}
				got = gen.typeContainsPlaceholder(fn.Parameters[0].Type, "any2")
				if got {
					t.Error("expected list of any NOT to contain 'any2' placeholder")
				}
			}
		}
	}
}

func TestTypeContainsPlaceholderNested(t *testing.T) {
	prog := mustParseProgram(t, "func f(m map of any2 to list of any)\n    x := m\n")
	gen := New(prog)

	for _, decl := range prog.Declarations {
		if fn, ok := decl.(*ast.FunctionDecl); ok {
			if len(fn.Parameters) > 0 {
				if !gen.typeContainsPlaceholder(fn.Parameters[0].Type, "any") {
					t.Error("expected map value type to contain 'any' placeholder")
				}
				if !gen.typeContainsPlaceholder(fn.Parameters[0].Type, "any2") {
					t.Error("expected map key type to contain 'any2' placeholder")
				}
			}
		}
	}
}

func TestTypeContainsPlaceholderNil(t *testing.T) {
	prog := mustParseProgram(t, "func main()\n    x := 1\n")
	gen := New(prog)

	if gen.typeContainsPlaceholder(nil, "any") {
		t.Error("expected nil type annotation to return false")
	}
}

func TestInferReturnCount(t *testing.T) {
	input := `func single() int
    return 1

func double() (int, error)
    return 0, empty

func main()
    a := single()
    b, e := double()
`

	prog := mustParseProgram(t, input)
	gen := New(prog)

	count, ok := gen.returnCountForFunctionName("single")
	if !ok || count != 1 {
		t.Errorf("returnCountForFunctionName(single) = (%d, %v), want (1, true)", count, ok)
	}

	count, ok = gen.returnCountForFunctionName("double")
	if !ok || count != 2 {
		t.Errorf("returnCountForFunctionName(double) = (%d, %v), want (2, true)", count, ok)
	}

	count, ok = gen.returnCountForFunctionName("nonexistent")
	if ok {
		t.Errorf("returnCountForFunctionName(nonexistent) should return false, got (%d, %v)", count, ok)
	}
}

func TestExternalInterfaceTypeInFunctionSignature(t *testing.T) {
	input := `import "net/http"

func Wrap(handler http.Handler) http.Handler
    return handler
`

	output := generateSource(t, input)

	if !strings.Contains(output, "func Wrap(handler http.Handler) http.Handler") {
		t.Errorf("expected http.Handler in function signature, got: %s", output)
	}

	if !strings.Contains(output, "return handler") {
		t.Errorf("expected return handler, got: %s", output)
	}
}
