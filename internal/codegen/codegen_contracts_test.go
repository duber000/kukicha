package codegen

import (
	"strings"
	"testing"

	kukiparser "github.com/duber000/kukicha/internal/parser"
	"github.com/duber000/kukicha/internal/semantic"
)

// fullPipelineRelease runs the pipeline with releaseMode enabled.
func fullPipelineRelease(t *testing.T, source, filename string) string {
	t.Helper()

	p, err := kukiparser.New(source, filename)
	if err != nil {
		t.Fatalf("lexer/parser init error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := semantic.NewWithFile(program, filename)
	semanticErrors := analyzer.Analyze()
	if len(semanticErrors) > 0 {
		t.Fatalf("semantic errors: %v", semanticErrors)
	}

	gen := New(program)
	gen.SetSourceFile(filename)
	gen.SetExprReturnCounts(analyzer.ReturnCounts())
	gen.SetExprTypes(analyzer.ExprTypes())
	gen.SetReleaseMode(true)

	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	return output
}

func TestRequiresDirective(t *testing.T) {
	source := `# kuki:requires "len(items) > 0"
func First(items list of string) string
    return items[0]
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)

	if !strings.Contains(output, `if !(len(items) > 0)`) {
		t.Errorf("expected requires check in output:\n%s", output)
	}
	if !strings.Contains(output, `panic("requires violated: len(items) > 0")`) {
		t.Errorf("expected requires panic message in output:\n%s", output)
	}
}

func TestEnsuresDirective(t *testing.T) {
	source := `# kuki:ensures "result >= 0"
func Abs(x int) int
    if x < 0
        return -x
    return x
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)

	if !strings.Contains(output, "defer func()") {
		t.Errorf("expected defer for ensures check in output:\n%s", output)
	}
	if !strings.Contains(output, `result >= 0`) {
		t.Errorf("expected ensures condition in output:\n%s", output)
	}
	if !strings.Contains(output, `panic("ensures violated: result >= 0")`) {
		t.Errorf("expected ensures panic message in output:\n%s", output)
	}
	// Named return should be present
	if !strings.Contains(output, "result int") {
		t.Errorf("expected named return for ensures in output:\n%s", output)
	}
}

func TestInvariantDirective(t *testing.T) {
	source := `# kuki:invariant "self.min <= self.max"
type Range
    min int
    max int
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)

	if !strings.Contains(output, "func (r Range) Validate()") {
		t.Errorf("expected Validate method in output:\n%s", output)
	}
	if !strings.Contains(output, `r.min <= r.max`) {
		t.Errorf("expected invariant condition in output:\n%s", output)
	}
	if !strings.Contains(output, `panic("invariant violated: self.min <= self.max")`) {
		t.Errorf("expected invariant panic message in output:\n%s", output)
	}
}

func TestRequiresWithKukichaKeywords(t *testing.T) {
	source := `# kuki:requires "x > 0 and x < 100"
func Bounded(x int) int
    return x
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)

	if !strings.Contains(output, "x > 0 && x < 100") {
		t.Errorf("expected 'and' translated to '&&' in output:\n%s", output)
	}
}

func TestReleaseMode_StripsContracts(t *testing.T) {
	source := `# kuki:requires "x > 0"
# kuki:ensures "result > 0"
func Double(x int) int
    return x * 2
`
	output := fullPipelineRelease(t, source, "test.kuki")
	assertValidGo(t, output)

	if strings.Contains(output, "requires violated") {
		t.Errorf("release mode should strip requires checks:\n%s", output)
	}
	if strings.Contains(output, "ensures violated") {
		t.Errorf("release mode should strip ensures checks:\n%s", output)
	}
}

func TestReleaseMode_StripsInvariant(t *testing.T) {
	source := `# kuki:invariant "self.min <= self.max"
type Range
    min int
    max int
`
	output := fullPipelineRelease(t, source, "test.kuki")
	assertValidGo(t, output)

	if strings.Contains(output, "Validate") {
		t.Errorf("release mode should strip invariant Validate method:\n%s", output)
	}
}

func TestMultipleRequires(t *testing.T) {
	source := `# kuki:requires "a > 0"
# kuki:requires "b > 0"
func Add(a int, b int) int
    return a + b
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)

	if !strings.Contains(output, `requires violated: a > 0`) {
		t.Errorf("expected first requires check:\n%s", output)
	}
	if !strings.Contains(output, `requires violated: b > 0`) {
		t.Errorf("expected second requires check:\n%s", output)
	}
}

func TestMultipleInvariants(t *testing.T) {
	source := `# kuki:invariant "self.min >= 0"
# kuki:invariant "self.max >= self.min"
type Bounds
    min int
    max int
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)

	if !strings.Contains(output, `invariant violated: self.min >= 0`) {
		t.Errorf("expected first invariant:\n%s", output)
	}
	if !strings.Contains(output, `invariant violated: self.max >= self.min`) {
		t.Errorf("expected second invariant:\n%s", output)
	}
}

func TestKukichaToGoExpr(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"x > 0 and x < 100", "x > 0 && x < 100"},
		{"a or b", "a || b"},
		{"not valid", "!valid"},
		{"a equals b", "a == b"},
		{"x > 0 and not done", "x > 0 && !done"},
	}
	for _, tt := range tests {
		result := kukichaToGoExpr(tt.input)
		if result != tt.expected {
			t.Errorf("kukichaToGoExpr(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
