package semantic

import (
	"strings"
	"testing"
	"testing/synctest"

	"github.com/duber000/kukicha/internal/parser"
)

func TestSimpleFunctionAnalysis(t *testing.T) {
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

	analyzer := New(program)
	errors := analyzer.Analyze()

	if len(errors) > 0 {
		t.Fatalf("semantic errors: %v", errors)
	}
}

func TestUndefinedVariable(t *testing.T) {
	input := `func Test() int
    return x
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

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

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

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

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

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

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	if len(errors) > 0 {
		t.Fatalf("unexpected semantic errors: %v", errors)
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

	analyzer := New(program)
	errors := analyzer.Analyze()

	if len(errors) > 0 {
		t.Fatalf("unexpected semantic errors: %v", errors)
	}
}

func TestMethodReceiver(t *testing.T) {
	input := `type Counter
    Value int
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	if len(errors) > 0 {
		t.Fatalf("unexpected semantic errors: %v", errors)
	}
}

func TestReturnValueCount(t *testing.T) {
	input := `func GetPair() (int, int)
    return 1
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

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

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

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

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	if len(errors) > 0 {
		t.Fatalf("unexpected semantic errors: %v", errors)
	}
}

func TestBooleanExpression(t *testing.T) {
	input := `func Test(x int, y int) bool
    return x > 5 and y < 10
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	if len(errors) > 0 {
		t.Fatalf("unexpected semantic errors: %v", errors)
	}
}

func TestInvalidBooleanOperand(t *testing.T) {
	input := `func Test(x int) bool
    return x and 5
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := New(program)
	errors := analyzer.Analyze()

	if len(errors) == 0 {
		t.Fatal("expected error for non-boolean operands to 'and'")
	}

	if !strings.Contains(errors[0].Error(), "logical operator requires boolean") {
		t.Errorf("expected boolean operator error, got: %v", errors[0])
	}
}

// TestConcurrentSemanticAnalysis tests that the semantic analyzer is thread-safe
// and multiple analyzers can run concurrently using synctest
func TestConcurrentSemanticAnalysis(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Test that semantic analyzer is thread-safe
		// Multiple analyzers should be able to run concurrently

		programs := []string{
			`func Add(a int, b int) int
    return a + b`,
			`type User
    name string
    age int`,
			`func Test() bool
    x := 5
    return x > 3`,
		}

		results := make(chan bool, len(programs))

		for _, src := range programs {
			go func(source string) {
				p, err := parser.New(source, "test.kuki")
				if err != nil {
					t.Errorf("parser error: %v", err)
					results <- false
					return
				}
				program, parseErrors := p.Parse()
				if len(parseErrors) > 0 {
					t.Errorf("parse errors: %v", parseErrors)
					results <- false
					return
				}
				analyzer := New(program)
				errors := analyzer.Analyze()
				if len(errors) > 0 {
					t.Errorf("semantic errors: %v", errors)
					results <- false
					return
				}
				results <- true
			}(src)
		}

		synctest.Wait()

		// Verify all completed successfully
		successCount := 0
		for range programs {
			select {
			case success := <-results:
				if success {
					successCount++
				}
			default:
				t.Error("Expected result not received")
			}
		}

		if successCount != len(programs) {
			t.Errorf("Expected %d successful analyses, got %d", len(programs), successCount)
		}
	})
}
