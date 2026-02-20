package codegen_test

import (
	"strings"
	"testing"

	"github.com/duber000/kukicha/internal/parser"
	. "github.com/duber000/kukicha/internal/codegen"
)

func TestOnErrMultiValueReturnInline(t *testing.T) {
	// onerr return with error "{error}" should substitute the caught error variable
	input := `func readData(path string) (string, error)
    return "data", empty

func Process(path string) (string, error)
    data := readData(path) onerr return "", error "{error}"
    return data, empty
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

	// The error variable should be err_1, not the Go builtin "error"
	if strings.Contains(output, `fmt.Sprintf("%v", error)`) {
		t.Errorf("error variable not substituted: {error} should become err_1, got: %s", output)
	}

	// Should use the actual error variable in the errors.New(fmt.Sprintf(...))
	if !strings.Contains(output, "err_1") {
		t.Errorf("expected err_1 error variable in output, got: %s", output)
	}

	// Should produce a multi-value return with errors.New
	if !strings.Contains(output, `return "", errors.New`) {
		t.Errorf("expected multi-value return with errors.New, got: %s", output)
	}
}

func TestOnErrMultiValueReturnInCallback(t *testing.T) {
	// Same issue inside a function literal (inline callback body)
	input := `func readData(path string) (string, error)
    return "data", empty

func MakeHandler() func(string) (string, error)
    return func(path string) (string, error)
        data := readData(path) onerr return "", error "{error}"
        return data, empty
`

	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		for _, e := range parseErrors {
			t.Logf("parse error: %s", e)
		}
		t.Fatalf("got %d parse errors", len(parseErrors))
	}

	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	t.Logf("Generated output:\n%s", output)

	// The error variable should be err_1 inside the function literal too
	if strings.Contains(output, `fmt.Sprintf("%v", error)`) {
		t.Errorf("error variable not substituted in callback: {error} should become err_1, got: %s", output)
	}

	if !strings.Contains(output, "err_1") {
		t.Errorf("expected err_1 error variable in callback output, got: %s", output)
	}

	if !strings.Contains(output, `return "", errors.New`) {
		t.Errorf("expected multi-value return with errors.New in callback, got: %s", output)
	}
}

func TestOnErrReturnWithStaticErrorMessage(t *testing.T) {
	// onerr return with a static (non-interpolated) error message
	input := `func readData(path string) (string, error)
    return "data", empty

func Process(path string) (string, error)
    data := readData(path) onerr return "", error "read failed"
    return data, empty
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

	if !strings.Contains(output, `return "", errors.New("read failed")`) {
		t.Errorf("expected static error message in return, got: %s", output)
	}
}

func TestOnErrReturnEmptyAndError(t *testing.T) {
	// onerr return empty, error "{error}" — common pattern for (T, error) functions
	input := `func readData(path string) (int, error)
    return 42, empty

func Process(path string) (int, error)
    data := readData(path) onerr return 0, error "{error}"
    return data, empty
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

	// Should substitute err_1 for {error}
	if strings.Contains(output, `fmt.Sprintf("%v", error)`) {
		t.Errorf("error variable not substituted, got: %s", output)
	}

	if !strings.Contains(output, "return 0, errors.New") {
		t.Errorf("expected return 0, errors.New(...) in output, got: %s", output)
	}
}
