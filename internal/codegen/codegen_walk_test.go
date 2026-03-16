package codegen

import (
	"testing"

	"github.com/duber000/kukicha/internal/ast"
)

func TestNeedsPrintBuiltin(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name: "has print call",
			input: `func main()
    print("hello")
`,
			expected: true,
		},
		{
			name: "no print call",
			input: `func main()
    x := 1
`,
			expected: false,
		},
		{
			name: "print in nested if",
			input: `func main()
    if true
        print("inside")
`,
			expected: true,
		},
		{
			name: "fmt.Println is not print builtin",
			input: `import "fmt"

func main()
    fmt.Println("hello")
`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prog := mustParseProgram(t, tt.input)
			gen := New(prog)
			got := gen.needsPrintBuiltin()
			if got != tt.expected {
				t.Errorf("needsPrintBuiltin() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNeedsErrorsPackage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name: "has error expression",
			input: `func fail() error
    return error "bad"
`,
			expected: true,
		},
		{
			name: "no error expression",
			input: `func ok() int
    return 42
`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prog := mustParseProgram(t, tt.input)
			gen := New(prog)
			got := gen.needsErrorsPackage()
			if got != tt.expected {
				t.Errorf("needsErrorsPackage() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNeedsStringInterpolation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name: "has interpolation",
			input: `func greet(name string) string
    return "Hello {name}!"
`,
			expected: true,
		},
		{
			name: "no interpolation",
			input: `func greet() string
    return "Hello world!"
`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prog := mustParseProgram(t, tt.input)
			gen := New(prog)
			got := gen.needsStringInterpolation()
			if got != tt.expected {
				t.Errorf("needsStringInterpolation() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCollectReservedNames(t *testing.T) {
	input := `func process(name string, count int)
    result := name
    for item in list of string{"a", "b"}
        x := item
`

	prog := mustParseProgram(t, input)
	gen := New(prog)
	gen.collectReservedNames()

	expectedNames := []string{"name", "count", "result", "item", "x"}
	for _, name := range expectedNames {
		if !gen.reservedNames[name] {
			t.Errorf("expected %q in reservedNames, but it was not found", name)
		}
	}
}

func TestCollectReservedNamesWithReceiver(t *testing.T) {
	input := `type User
    name string

func GetName on u User string
    return u.name
`

	prog := mustParseProgram(t, input)
	gen := New(prog)
	gen.collectReservedNames()

	if !gen.reservedNames["u"] {
		t.Error("expected receiver name 'u' in reservedNames")
	}
}

func TestCollectReservedNamesForLoop(t *testing.T) {
	input := `func count()
    for i from 0 to 10
        x := i
`

	prog := mustParseProgram(t, input)
	gen := New(prog)
	gen.collectReservedNames()

	if !gen.reservedNames["i"] {
		t.Error("expected 'i' in reservedNames from for-numeric loop")
	}
	if !gen.reservedNames["x"] {
		t.Error("expected 'x' in reservedNames from loop body")
	}
}

func TestWalkProgramShortCircuits(t *testing.T) {
	input := `func main()
    x := 1
    y := 2
    z := 3
`

	prog := mustParseProgram(t, input)
	gen := New(prog)

	callCount := 0
	gen.walkProgram(func(e ast.Expression) bool {
		callCount++
		return true // short-circuit immediately
	})

	if callCount != 1 {
		t.Errorf("expected walkProgram to short-circuit after 1 call, got %d", callCount)
	}
}
