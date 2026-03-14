package semantic

import (
	"strings"
	"testing"

	"github.com/duber000/kukicha/internal/ast"
)

func TestSimpleFunctionAnalysis(t *testing.T) {
	input := `func Add(a int, b int) int
    return a + b
`

	analyzer, errors := analyzeSource(t, input)
	_ = analyzer

	if len(errors) > 0 {
		t.Fatalf("semantic errors: %v", errors)
	}
}

func TestUndefinedVariable(t *testing.T) {
	input := `func Test() int
    return x
`

	analyzer, errors := analyzeSource(t, input)
	_ = analyzer

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

	analyzer, errors := analyzeSource(t, input)
	_ = analyzer

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

	analyzer, errors := analyzeSource(t, input)
	_ = analyzer

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

	analyzer, errors := analyzeSource(t, input)
	_ = analyzer

	if len(errors) > 0 {
		t.Fatalf("unexpected semantic errors: %v", errors)
	}
}

func TestTypeDeclaration(t *testing.T) {
	input := `type Person
    Name string
    Age int
`

	analyzer, errors := analyzeSource(t, input)
	_ = analyzer

	if len(errors) > 0 {
		t.Fatalf("unexpected semantic errors: %v", errors)
	}
}

func TestMethodReceiver(t *testing.T) {
	input := `type Counter
    Value int
`

	analyzer, errors := analyzeSource(t, input)
	_ = analyzer

	if len(errors) > 0 {
		t.Fatalf("unexpected semantic errors: %v", errors)
	}
}

func TestReturnValueCount(t *testing.T) {
	input := `func GetPair() (int, int)
    return 1
`

	analyzer, errors := analyzeSource(t, input)
	_ = analyzer

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

	analyzer, errors := analyzeSource(t, input)
	_ = analyzer

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

	analyzer, errors := analyzeSource(t, input)
	_ = analyzer

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

	analyzer, errors := analyzeSource(t, input)
	_ = analyzer

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

	analyzer, errors := analyzeSource(t, input)
	_ = analyzer

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

func TestTypedPipedSwitchSemantic(t *testing.T) {
	input := `func Convert(value any) string
    result := value |> switch as v
        when string
            return v
        when int
            return "number"
        otherwise
            return "other"
    return result
`

	analyzer, errors := analyzeSource(t, input)
	_ = analyzer

	if len(errors) > 0 {
		t.Fatalf("unexpected semantic errors: %v", errors)
	}
}

func TestTypedPipedSwitchComputedReturnSemantic(t *testing.T) {
	input := `import "os/exec"

func ExitCodeOrOne(err error) int
    code := err |> switch as exitErr
        when reference exec.ExitError
            return exitErr.ExitCode()
        otherwise
            return 1
    return code
`

	analyzer, errors := analyzeSource(t, input)
	_ = analyzer

	if len(errors) > 0 {
		t.Fatalf("unexpected semantic errors: %v", errors)
	}
}

func TestBooleanExpression(t *testing.T) {
	input := `func Test(x int, y int) bool
    return x > 5 and y < 10
`

	analyzer, errors := analyzeSource(t, input)
	_ = analyzer

	if len(errors) > 0 {
		t.Fatalf("unexpected semantic errors: %v", errors)
	}
}

func TestInvalidBooleanOperand(t *testing.T) {
	input := `func Test(x int) bool
    return x and 5
`

	analyzer, errors := analyzeSource(t, input)
	_ = analyzer

	if len(errors) == 0 {
		t.Fatal("expected error for non-boolean operands to 'and'")
	}

	if !strings.Contains(errors[0].Error(), "logical operator requires boolean") {
		t.Errorf("expected boolean operator error, got: %v", errors[0])
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

	analyzer, errors := analyzeSource(t, input)
	_ = analyzer

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

	analyzer, errors := analyzeSource(t, input)
	_ = analyzer

	if len(errors) == 0 {
		t.Fatal("expected type mismatch error for incompatible pipe return")
	}

	if !strings.Contains(errors[0].Error(), "cannot return") {
		t.Errorf("expected type mismatch error, got: %v", errors[0])
	}
}

func TestPlaceholderTypingInPipedCall(t *testing.T) {
	input := `func WriteJSON(w string, data string) error
    return empty

func Process(w string, data string) error
    data |> WriteJSON(w, _)
    return empty
`
	analyzer, errors := analyzeSource(t, input)
	if len(errors) > 0 {
		t.Fatalf("unexpected errors: %v", errors)
	}

	// Walk exprTypes looking for a "_" identifier with a resolved type
	found := false
	for expr, ti := range analyzer.ExprTypes() {
		if ident, ok := expr.(*ast.Identifier); ok && ident.Value == "_" {
			if ti.Kind == TypeKindString {
				found = true
			} else if ti.Kind != TypeKindUnknown {
				t.Errorf("expected string type for _, got %v", ti)
			}
		}
	}
	if !found {
		t.Error("expected _ placeholder to be typed as string from WriteJSON's second parameter")
	}
}
