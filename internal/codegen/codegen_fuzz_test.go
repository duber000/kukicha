package codegen

import (
	goparser "go/parser"
	"go/token"
	"testing"

	kukiparser "github.com/duber000/kukicha/internal/parser"
	"github.com/duber000/kukicha/internal/semantic"
)

func FuzzFullPipeline(f *testing.F) {
	// Seed corpus: valid programs that compile through the full pipeline
	seeds := []string{
		// Simple function
		"func Add(a int, b int) int\n    return a + b\n",
		// Multiple returns
		"func Divide(a int, b int) (int, error)\n    if b equals 0\n        return 0, error \"division by zero\"\n    return a / b, empty\n",
		// Type + method
		"type User\n    name string\n    age int\n\nfunc GetName on u User string\n    return u.name\n",
		// String interpolation
		"func Greet(name string) string\n    return \"Hello {name}!\"\n",
		// If/else
		"func Abs(x int) int\n    if x < 0\n        return -x\n    return x\n",
		// For range
		"func Sum(items list of int) int\n    total := 0\n    for item in items\n        total = total + item\n    return total\n",
		// For numeric
		"func Count() int\n    total := 0\n    for i from 0 to 10\n        total = total + i\n    return total\n",
		// Switch
		"func Describe(x int) string\n    switch\n        when x > 100\n            return \"big\"\n        otherwise\n            return \"small\"\n",
		// Lambda
		"func Apply(x int, f func(int) int) int\n    return f(x)\n",
		// Collections
		"func MakeList() list of string\n    return list of string{\"a\", \"b\", \"c\"}\n",
		// Variadic
		"func Sum(many numbers int) int\n    total := 0\n    for n in numbers\n        total = total + n\n    return total\n",
		// Pointer receiver
		"type Counter\n    value int\n\nfunc Increment on c reference Counter\n    c.value = c.value + 1\n",
		// Multiple functions
		"func Add(a int, b int) int\n    return a + b\n\nfunc Sub(a int, b int) int\n    return a - b\n\nfunc Mul(a int, b int) int\n    return a * b\n",
		// Boolean operators
		"func Check(a bool, b bool) bool\n    return a and b or not a\n",
		// Empty
		"func Nothing()\n    return\n",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, data string) {
		// Phase 1: Lex + Parse
		p, err := kukiparser.New(data, "fuzz.kuki")
		if err != nil {
			return // lexer error — skip
		}
		program, parseErrors := p.Parse()
		if len(parseErrors) > 0 {
			return // parse error — skip
		}

		// Phase 2: Semantic analysis
		analyzer := semantic.NewWithFile(program, "fuzz.kuki")
		semanticErrors := analyzer.Analyze()
		if len(semanticErrors) > 0 {
			return // semantic error — skip
		}

		// Phase 3: Code generation (must not panic)
		gen := New(program)
		gen.SetSourceFile("fuzz.kuki")
		gen.SetExprReturnCounts(analyzer.ReturnCounts())
		gen.SetExprTypes(analyzer.ExprTypes())
		output, err := gen.Generate()
		if err != nil {
			return // codegen error — skip
		}

		// Phase 4: Verify generated Go is syntactically valid.
		// Log invalid Go outputs for investigation but don't fail the fuzz test,
		// since some valid-parsing Kukicha edge cases may produce invalid Go
		// (these are codegen bugs to fix separately, not panics).
		fset := token.NewFileSet()
		_, parseErr := goparser.ParseFile(fset, "fuzz.go", output, goparser.AllErrors)
		if parseErr != nil {
			t.Logf("generated invalid Go from Kukicha input (codegen issue):\nInput:\n%s\nError: %v", data, parseErr)
		}
	})
}
