package codegen

import (
	"github.com/duber000/kukicha/internal/parser"
	"github.com/duber000/kukicha/internal/semantic"
	"strings"
	"testing"
)

func TestOnErrErrorMessageUsesProvidedString(t *testing.T) {
	input := `func bar() (int, error)
    return 0, error "nope"

func Foo() (int, error)
    x := 0
    x = bar() onerr error "bad: {error}"
    return x, empty
`

	output := generateSource(t, input)

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

	output := generateSource(t, input)

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

	output := generateSource(t, input)

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

	output := generateSource(t, input)

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

func TestStringInterpolation(t *testing.T) {
	input := `func Greet(name string) string
    return "Hello {name}!"
`

	output := generateSource(t, input)

	t.Logf("Generated output:\n%s", output)

	if !strings.Contains(output, "fmt.Sprintf") {
		t.Errorf("expected fmt.Sprintf for string interpolation, got: %s", output)
	}

	if !strings.Contains(output, "\"fmt\"") {
		t.Errorf("expected fmt import, got: %s", output)
	}
}

func TestEscapedBracesLiteral(t *testing.T) {
	// \{ and \} should produce literal braces in output, not interpolation
	input := `func Format() string
    return "json: \{key: value\}"
`

	output := generateSource(t, input)

	t.Logf("Generated output:\n%s", output)

	// Should produce a plain string with literal braces, NOT fmt.Sprintf
	if strings.Contains(output, "fmt.Sprintf") {
		t.Errorf("escaped braces should not trigger fmt.Sprintf, got: %s", output)
	}

	if !strings.Contains(output, "{key: value}") {
		t.Errorf("expected literal braces in output, got: %s", output)
	}
}

func TestEscapedBracesMixedWithInterpolation(t *testing.T) {
	// \{ and \} should produce literal braces even when mixed with {expr} interpolation
	input := `func Format(name string) string
    return "\{name\}: {name}"
`

	output := generateSource(t, input)

	t.Logf("Generated output:\n%s", output)

	// Should have fmt.Sprintf for the interpolated part
	if !strings.Contains(output, "fmt.Sprintf") {
		t.Errorf("expected fmt.Sprintf for interpolated part, got: %s", output)
	}

	// The literal braces should appear as { and } in the format string
	if !strings.Contains(output, "{name}") {
		t.Errorf("expected literal {name} in format string, got: %s", output)
	}
}

func TestPrintfStyleMethodInterpolation(t *testing.T) {
	// Test that t.Errorf with interpolation generates inline format args (not fmt.Sprintf)
	input := `import "testing"

func TestExample(t reference testing.T)
    name := "world"
    t.Errorf("hello {name}")
`

	output := generateSource(t, input)

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

	output := generateSource(t, input)

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

func TestErrorAsVariableName(t *testing.T) {
	input := `func Main()
    val, error := divide(10, 0)
    print(error)

func divide(a int, b int) (int, error)
    if b == 0
        return 0, error "division by zero"
    return a / b, empty
`

	output := generateSource(t, input)

	// error used as variable name in multi-value assignment
	if !strings.Contains(output, "val, error := divide(10, 0)") {
		t.Errorf("expected 'val, error := divide(10, 0)', got: %s", output)
	}
	// error keyword still works as errors.New()
	if !strings.Contains(output, `errors.New("division by zero")`) {
		t.Errorf("expected errors.New for error keyword, got: %s", output)
	}
}

func TestEmptyAsVariableName(t *testing.T) {
	input := `func Main()
    empty := 42
    print(empty)
`

	output := generateSource(t, input)

	if !strings.Contains(output, "empty := 42") {
		t.Errorf("expected 'empty := 42', got: %s", output)
	}
}

func TestEmptyIdentifierInPipeExpression(t *testing.T) {
	input := `import "stdlib/iterator"

func addInts(acc int, n int) int
    return acc + n

func Run() int
    empty := list of int{}
    return empty |> iterator.Values() |> iterator.Reduce(42, addInts)
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
	gen.SetExprTypes(analyzer.ExprTypes())
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if strings.Contains(output, "iterator.Values(nil)") {
		t.Fatalf("expected pipe to preserve identifier named empty, got: %s", output)
	}
	if !strings.Contains(output, "iterator.Values(empty)") {
		t.Fatalf("expected pipe to use the empty variable, got: %s", output)
	}
}

func TestFilepathSeparatorEscape(t *testing.T) {
	// \sep should produce string(filepath.Separator) at runtime, not a literal character
	input := `func Sep() string
    return "\sep"
`

	output := generateSource(t, input)

	t.Logf("Generated output:\n%s", output)

	if !strings.Contains(output, "string(filepath.Separator)") {
		t.Errorf("expected string(filepath.Separator) in output, got: %s", output)
	}
	if !strings.Contains(output, `"path/filepath"`) {
		t.Errorf("expected path/filepath auto-import, got: %s", output)
	}
}

func TestFilepathSeparatorEscapeMixedWithInterpolation(t *testing.T) {
	// \sep mixed with {expr} interpolation should work together
	input := `func JoinParts(dir string, file string) string
    return "{dir}\sep{file}"
`

	output := generateSource(t, input)

	t.Logf("Generated output:\n%s", output)

	if !strings.Contains(output, "string(filepath.Separator)") {
		t.Errorf("expected string(filepath.Separator) in output, got: %s", output)
	}
	if !strings.Contains(output, "dir") {
		t.Errorf("expected dir variable in output, got: %s", output)
	}
	if !strings.Contains(output, "file") {
		t.Errorf("expected file variable in output, got: %s", output)
	}
}

func TestOnErrErrorExprThreeReturnValues(t *testing.T) {
	// Fix 3: onerr error "msg" (shorthand) should emit zero values for all
	// non-error return positions when function has 3+ return values.
	input := `func fetch() (string, int, error)
    return "", 0, empty

func Process() (string, int, error)
    x := fetch() onerr error "failed: {error}"
    return x, 42, empty
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

	t.Logf("Generated output:\n%s", output)

	// Should emit zero values for all non-error return positions
	if !strings.Contains(output, `return "", 0,`) {
		t.Errorf("expected zero values for string and int return types, got: %s", output)
	}
}

func TestOnErrDiscardFallbackEmitsComment(t *testing.T) {
	// Fix 4: when return count inference fails for discard, emit a comment warning.
	// We simulate this by using a method call on an unknown external type
	// where inferReturnCount cannot determine the count.
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

	// With semantic analysis, inferReturnCount succeeds — should NOT have warning comment
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

	if strings.Contains(output, "// kukicha: could not infer") {
		t.Errorf("should not emit warning comment when semantic return counts are available, got: %s", output)
	}

	// Without semantic analysis, inferReturnCount fails — should have warning comment
	gen2 := New(program)
	output2, err := gen2.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output2, "// kukicha: could not infer") {
		t.Errorf("expected warning comment when return count inference fails, got: %s", output2)
	}
}

func TestTypeCastConcreteUsesConversion(t *testing.T) {
	// Casting to a concrete type (int, string) should use conversion syntax: int(x)
	input := `func Foo(x float64) int
    return x as int
`
	output := generateSource(t, input)

	if !strings.Contains(output, "int(x)") {
		t.Errorf("expected conversion syntax int(x), got: %s", output)
	}
	if strings.Contains(output, ".(int)") {
		t.Errorf("should not use assertion syntax for concrete type, got: %s", output)
	}
}

func TestTypeCastInterfaceUsesAssertion(t *testing.T) {
	// Casting to an interface type should use assertion syntax: x.(error)
	input := `func Foo(x any) error
    return x as error
`
	output := generateSource(t, input)

	if !strings.Contains(output, ".(error)") {
		t.Errorf("expected assertion syntax .(error), got: %s", output)
	}
}

func TestTypeCastLocalInterfaceUsesAssertion(t *testing.T) {
	// Casting to a locally-declared interface should use assertion syntax
	input := `interface Stringer
    String() string

func Foo(x any) Stringer
    return x as Stringer
`
	output := generateSource(t, input)

	if !strings.Contains(output, ".(Stringer)") {
		t.Errorf("expected assertion syntax .(Stringer) for local interface, got: %s", output)
	}
}

func TestDeeplyNestedIndentation(t *testing.T) {
	// 10+ levels of nesting should compile correctly
	input := `func deep(n int) int
    if n > 0
        if n > 1
            if n > 2
                if n > 3
                    if n > 4
                        if n > 5
                            if n > 6
                                if n > 7
                                    if n > 8
                                        if n > 9
                                            return n
    return 0
`

	output := generateSource(t, input)

	if !strings.Contains(output, "return n") {
		t.Errorf("expected deeply nested return, got:\n%s", output)
	}

	// Count tab depth — innermost return should have 11 tabs
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimLeft(line, "\t")
		if strings.Contains(trimmed, "return n") {
			tabs := len(line) - len(trimmed)
			if tabs < 11 {
				t.Errorf("expected at least 11 tabs for deeply nested return, got %d", tabs)
			}
		}
	}
}

func TestParserCascadesMultipleErrors(t *testing.T) {
	// Parser should report multiple errors, not stop at the first one
	input := `func Bad(a, b int) int
    return a +
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	_, parseErrors := p.Parse()
	if len(parseErrors) == 0 {
		t.Error("expected parse errors for malformed input, got none")
	}
}

func TestOnErrContinueInLoop(t *testing.T) {
	input := `import "strconv"

func sumValid(items list of string) int
    total := 0
    for item in items
        n := strconv.Atoi(item) onerr continue
        total = total + n
    return total
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := semantic.NewWithFile(program, "test.kuki")
	semErrors := analyzer.Analyze()
	if len(semErrors) > 0 {
		t.Fatalf("semantic errors: %v", semErrors)
	}

	gen := New(program)
	gen.SetExprReturnCounts(analyzer.ReturnCounts())
	gen.SetExprTypes(analyzer.ExprTypes())
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "continue") {
		t.Errorf("expected 'continue' in onerr handler, got:\n%s", output)
	}
}

func TestOnErrBreakInLoop(t *testing.T) {
	input := `import "strconv"

func firstValid(items list of string) int
    result := 0
    for item in items
        n := strconv.Atoi(item) onerr break
        result = n
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

	analyzer := semantic.NewWithFile(program, "test.kuki")
	semErrors := analyzer.Analyze()
	if len(semErrors) > 0 {
		t.Fatalf("semantic errors: %v", semErrors)
	}

	gen := New(program)
	gen.SetExprReturnCounts(analyzer.ReturnCounts())
	gen.SetExprTypes(analyzer.ExprTypes())
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "break") {
		t.Errorf("expected 'break' in onerr handler, got:\n%s", output)
	}
}

func TestOnErrBlockMultiStatement(t *testing.T) {
	input := `import "os"

func readOrDie(path string) list of byte
    data := os.ReadFile(path) onerr
        print("Failed to read {path}: {error}")
        panic "aborting"
    return data
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := semantic.NewWithFile(program, "test.kuki")
	semErrors := analyzer.Analyze()
	if len(semErrors) > 0 {
		t.Fatalf("semantic errors: %v", semErrors)
	}

	gen := New(program)
	gen.SetExprReturnCounts(analyzer.ReturnCounts())
	gen.SetExprTypes(analyzer.ExprTypes())
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(output, "!= nil") {
		t.Errorf("expected nil check in onerr block, got:\n%s", output)
	}
	if !strings.Contains(output, "panic") {
		t.Errorf("expected panic in onerr block, got:\n%s", output)
	}
}
