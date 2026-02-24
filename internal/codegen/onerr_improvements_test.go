package codegen_test

import (
	"strings"
	"testing"

	"github.com/duber000/kukicha/internal/ast"
	"github.com/duber000/kukicha/internal/parser"
	. "github.com/duber000/kukicha/internal/codegen"
)

// ---------------------------------------------------------------------------
// Proposal A: onerr return shorthand
// ---------------------------------------------------------------------------

// TestOnErrShorthandReturnSingleError checks that bare "onerr return" in a
// function returning only error emits "return <errVar>" (raw propagation, no wrapping).
func TestOnErrShorthandReturnSingleError(t *testing.T) {
	input := `func readData(path string) error
    return empty

func Process(path string) error
    readData(path) onerr return
    return empty
`
	program := mustParse(t, input)
	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	t.Logf("Generated output:\n%s", output)

	// Must contain a raw "return err_1" (or similar), not errors.New wrapping.
	if strings.Contains(output, "errors.New") {
		t.Errorf("onerr return shorthand must not wrap error; got: %s", output)
	}
	if !strings.Contains(output, "return err") {
		t.Errorf("expected raw return of error variable, got: %s", output)
	}
}

// TestOnErrShorthandReturnMultiValue checks that bare "onerr return" in a
// (string, error) function emits zero values for non-error positions.
func TestOnErrShorthandReturnMultiValue(t *testing.T) {
	input := `func readData(path string) (string, error)
    return "data", empty

func Process(path string) (string, error)
    data := readData(path) onerr return
    return data, empty
`
	program := mustParse(t, input)
	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	t.Logf("Generated output:\n%s", output)

	// Should return zero string "" plus the raw error variable.
	if !strings.Contains(output, `return "", err`) {
		t.Errorf("expected zero-value return with raw error, got: %s", output)
	}
	// Must not wrap the error.
	if strings.Contains(output, "errors.New") || strings.Contains(output, "fmt.Errorf") {
		t.Errorf("onerr return shorthand must not wrap error; got: %s", output)
	}
}

// TestOnErrShorthandReturnDoesNotAffectVerboseForm verifies that the existing
// verbose "onerr return empty, error "{error}"" form still works unchanged.
func TestOnErrShorthandReturnDoesNotAffectVerboseForm(t *testing.T) {
	input := `func readData(path string) (string, error)
    return "data", empty

func Process(path string) (string, error)
    data := readData(path) onerr return "", error "{error}"
    return data, empty
`
	program := mustParse(t, input)
	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	t.Logf("Generated output:\n%s", output)

	// Verbose form wraps with errors.New.
	if !strings.Contains(output, "errors.New") {
		t.Errorf("verbose onerr return should still produce errors.New; got: %s", output)
	}
}

// ---------------------------------------------------------------------------
// Proposal B: onerr as e
// ---------------------------------------------------------------------------

// TestOnErrAliasBlockResolvesAlias checks that {e} in an "onerr as e" block
// resolves to the caught error variable, not the literal identifier "e".
func TestOnErrAliasBlockResolvesAlias(t *testing.T) {
	input := `func readData(path string) (string, error)
    return "data", empty

func Process(path string) (string, error)
    data := readData(path) onerr as e
        print("fetch failed: {e}")
        return "", empty
    return data, empty
`
	program := mustParse(t, input)
	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	t.Logf("Generated output:\n%s", output)

	// {e} must resolve to the error variable (err_1), not the literal string "e".
	if strings.Contains(output, `"fetch failed: e"`) {
		t.Errorf("{e} alias was not resolved to the error variable; got: %s", output)
	}
	if !strings.Contains(output, "err_1") {
		t.Errorf("expected error variable err_1 in output; got: %s", output)
	}
}

// TestOnErrAliasBlockErrorStillValid checks that {error} remains valid alongside an alias.
func TestOnErrAliasBlockErrorStillValid(t *testing.T) {
	input := `func readData(path string) (string, error)
    return "data", empty

func Process(path string) (string, error)
    data := readData(path) onerr as e
        print("error: {error}")
        return "", empty
    return data, empty
`
	program := mustParse(t, input)
	gen := New(program)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	t.Logf("Generated output:\n%s", output)

	// {error} must still resolve to the error variable.
	if strings.Contains(output, `"error: error"`) {
		t.Errorf("{error} was not resolved inside onerr as e block; got: %s", output)
	}
	if !strings.Contains(output, "err_1") {
		t.Errorf("expected error variable err_1 in output; got: %s", output)
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func mustParse(t *testing.T, input string) *ast.Program {
	t.Helper()
	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser init error: %v", err)
	}
	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}
	return program
}
