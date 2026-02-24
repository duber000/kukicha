package semantic

import (
	"strings"
	"testing"

	"github.com/duber000/kukicha/internal/parser"
)

// ---------------------------------------------------------------------------
// Proposal A: onerr return semantic validation
// ---------------------------------------------------------------------------

// TestOnErrShorthandReturnValidInErrorReturningFunc verifies that bare "onerr return"
// is accepted when the enclosing function has a compatible (T, error) signature.
func TestOnErrShorthandReturnValidInErrorReturningFunc(t *testing.T) {
	input := `func readData(path string) (string, error)
    return "data", empty

func Process(path string) (string, error)
    data := readData(path) onerr return
    return data, empty
`
	errors := analyzeInput(t, input)
	if len(errors) > 0 {
		t.Errorf("expected no semantic errors for valid onerr return, got: %v", errors)
	}
}

// TestOnErrShorthandReturnValidInErrorOnlyFunc verifies acceptance in a (error)-only function.
func TestOnErrShorthandReturnValidInErrorOnlyFunc(t *testing.T) {
	input := `func readData(path string) error
    return empty

func Process(path string) error
    readData(path) onerr return
    return empty
`
	errors := analyzeInput(t, input)
	if len(errors) > 0 {
		t.Errorf("expected no semantic errors, got: %v", errors)
	}
}

// TestOnErrShorthandReturnRejectedInVoidFunc verifies that "onerr return" is rejected
// when the enclosing function has no error return.
func TestOnErrShorthandReturnRejectedInVoidFunc(t *testing.T) {
	input := `func readData(path string) (string, error)
    return "data", empty

func Process(path string)
    data := readData(path) onerr return
`
	errors := analyzeInput(t, input)
	if len(errors) == 0 {
		t.Fatal("expected semantic error for onerr return in non-error-returning function")
	}
	if !strings.Contains(errors[0].Error(), "onerr return") {
		t.Errorf("expected onerr return error, got: %v", errors[0])
	}
}

// TestOnErrShorthandReturnRejectedInIntReturningFunc verifies rejection when the function
// returns (int) — a non-error return type.
func TestOnErrShorthandReturnRejectedInIntReturningFunc(t *testing.T) {
	input := `func readData(path string) (string, error)
    return "data", empty

func Process(path string) int
    data := readData(path) onerr return
    return 0
`
	errors := analyzeInput(t, input)
	if len(errors) == 0 {
		t.Fatal("expected semantic error for onerr return in int-returning function")
	}
	if !strings.Contains(errors[0].Error(), "onerr return") {
		t.Errorf("expected onerr return error, got: %v", errors[0])
	}
}

// ---------------------------------------------------------------------------
// Proposal B: onerr as e — {err} diagnostic improvement
// ---------------------------------------------------------------------------

// TestOnErrAliasHintIncludesAliasName verifies that the {err} diagnostic mentions
// the alias when "onerr as e" is active.
func TestOnErrAliasHintIncludesAliasName(t *testing.T) {
	input := `func readData(path string) (string, error)
    return "data", empty

func Process(path string) (string, error)
    data := readData(path) onerr as myErr
        print("error: {err}")
        return "", empty
    return data, empty
`
	errors := analyzeInput(t, input)
	if len(errors) == 0 {
		t.Fatal("expected semantic error for {err} inside onerr")
	}
	found := errors[0].Error()
	if !strings.Contains(found, "myErr") {
		t.Errorf("expected alias name 'myErr' in diagnostic, got: %s", found)
	}
}

// TestOnErrNoAliasHintMentionsOnerr verifies that without an alias the {err}
// diagnostic still suggests "onerr as e".
func TestOnErrNoAliasHintMentionsOnerr(t *testing.T) {
	input := `func readData(path string) (string, error)
    return "data", empty

func Process(path string) (string, error)
    data := readData(path) onerr
        print("error: {err}")
        return "", empty
    return data, empty
`
	errors := analyzeInput(t, input)
	if len(errors) == 0 {
		t.Fatal("expected semantic error for {err} inside onerr block")
	}
	found := errors[0].Error()
	if !strings.Contains(found, "onerr as e") {
		t.Errorf("expected 'onerr as e' suggestion in diagnostic, got: %s", found)
	}
}

// ---------------------------------------------------------------------------
// Proposal C: lint warnings for risky onerr handlers
// ---------------------------------------------------------------------------

// TestOnErrDiscardOutsideTestFileWarns verifies that onerr discard in a non-test
// source file produces a lint warning.
func TestOnErrDiscardOutsideTestFileWarns(t *testing.T) {
	input := `func readData(path string) (string, error)
    return "data", empty

func Process(path string) string
    data := readData(path) onerr discard
    return data
`
	_, warnings := analyzeInputWithFile(t, input, "app.kuki")
	if len(warnings) == 0 {
		t.Fatal("expected lint warning for onerr discard outside test file, got none")
	}
	if !strings.Contains(warnings[0].Error(), "discard") {
		t.Errorf("expected 'discard' in warning message, got: %s", warnings[0])
	}
}

// TestOnErrDiscardInTestFileNoWarn verifies that onerr discard in a _test.kuki
// file does NOT produce a lint warning.
func TestOnErrDiscardInTestFileNoWarn(t *testing.T) {
	input := `func readData(path string) (string, error)
    return "data", empty

func Process(path string) string
    data := readData(path) onerr discard
    return data
`
	_, warnings := analyzeInputWithFile(t, input, "app_test.kuki")
	if len(warnings) != 0 {
		t.Errorf("expected no lint warnings in test file, got: %v", warnings)
	}
}

// TestOnErrPanicInLibraryPackageWarns verifies that onerr panic in a non-main
// package produces a lint warning.
func TestOnErrPanicInLibraryPackageWarns(t *testing.T) {
	input := `petiole mylib

func readData(path string) (string, error)
    return "data", empty

func Process(path string) string
    data := readData(path) onerr panic "operation failed"
    return data
`
	_, warnings := analyzeInputWithFile(t, input, "mylib.kuki")
	if len(warnings) == 0 {
		t.Fatal("expected lint warning for onerr panic in library package, got none")
	}
	if !strings.Contains(warnings[0].Error(), "panic") {
		t.Errorf("expected 'panic' in warning message, got: %s", warnings[0])
	}
}

// TestOnErrPanicInMainPackageNoWarn verifies that onerr panic in the main
// package does NOT produce a lint warning.
func TestOnErrPanicInMainPackageNoWarn(t *testing.T) {
	input := `petiole main

func readData(path string) (string, error)
    return "data", empty

func main()
    data := readData("file.txt") onerr panic "operation failed"
    print(data)
`
	_, warnings := analyzeInputWithFile(t, input, "main.kuki")
	if len(warnings) != 0 {
		t.Errorf("expected no lint warnings for onerr panic in main package, got: %v", warnings)
	}
}

// TestOnErrPanicNoPetioleNoWarn verifies that onerr panic with no petiole
// declaration (implicit main) does NOT produce a lint warning.
func TestOnErrPanicNoPetioleNoWarn(t *testing.T) {
	input := `func readData(path string) (string, error)
    return "data", empty

func main()
    data := readData("file.txt") onerr panic "operation failed"
    print(data)
`
	_, warnings := analyzeInputWithFile(t, input, "main.kuki")
	if len(warnings) != 0 {
		t.Errorf("expected no lint warnings when no petiole (implicit main), got: %v", warnings)
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func analyzeInput(t *testing.T, input string) []error {
	t.Helper()
	errs, _ := analyzeInputWithFile(t, input, "test.kuki")
	return errs
}

func analyzeInputWithFile(t *testing.T, input, filename string) (errs []error, warnings []error) {
	t.Helper()
	p, err := parser.New(input, filename)
	if err != nil {
		t.Fatalf("parser init error: %v", err)
	}
	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}
	analyzer := NewWithFile(program, filename)
	errs = analyzer.Analyze()
	warnings = analyzer.Warnings()
	return
}
