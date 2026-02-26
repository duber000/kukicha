package lsp

import (
	"errors"
	"testing"

	"github.com/sourcegraph/go-lsp"
)

func TestErrorToDiagnostic_StandardFormat(t *testing.T) {
	err := errors.New("test.kuki:5:10: undefined identifier 'foo'")
	diag := errorToDiagnostic(err)

	// Line 5, column 10 → 0-indexed: 4, 9
	if diag.Range.Start.Line != 4 {
		t.Errorf("expected line 4, got %d", diag.Range.Start.Line)
	}
	if diag.Range.Start.Character != 9 {
		t.Errorf("expected character 9, got %d", diag.Range.Start.Character)
	}
	if diag.Message != "undefined identifier 'foo'" {
		t.Errorf("expected message 'undefined identifier 'foo'', got: %s", diag.Message)
	}
	if diag.Severity != lsp.Error {
		t.Errorf("expected severity Error, got %d", diag.Severity)
	}
	if diag.Source != "kukicha" {
		t.Errorf("expected source 'kukicha', got %s", diag.Source)
	}
}

func TestErrorToDiagnostic_NonStandardFormat(t *testing.T) {
	err := errors.New("some generic error message")
	diag := errorToDiagnostic(err)

	// Should fallback to line 0, col 0
	if diag.Range.Start.Line != 0 {
		t.Errorf("expected line 0, got %d", diag.Range.Start.Line)
	}
	if diag.Range.Start.Character != 0 {
		t.Errorf("expected character 0, got %d", diag.Range.Start.Character)
	}
	if diag.Message != "some generic error message" {
		t.Errorf("expected full message, got: %s", diag.Message)
	}
}

func TestErrorToDiagnostic_Line1Col1(t *testing.T) {
	err := errors.New("test.kuki:1:1: syntax error")
	diag := errorToDiagnostic(err)

	// Line 1, col 1 → 0-indexed: 0, 0
	if diag.Range.Start.Line != 0 {
		t.Errorf("expected line 0, got %d", diag.Range.Start.Line)
	}
	if diag.Range.Start.Character != 0 {
		t.Errorf("expected character 0, got %d", diag.Range.Start.Character)
	}
}

func TestErrorToDiagnostic_EndRange(t *testing.T) {
	err := errors.New("test.kuki:10:5: type mismatch")
	diag := errorToDiagnostic(err)

	// End should be one character after start
	if diag.Range.End.Line != diag.Range.Start.Line {
		t.Errorf("expected end line to equal start line")
	}
	if diag.Range.End.Character != diag.Range.Start.Character+1 {
		t.Errorf("expected end character to be start+1, got %d", diag.Range.End.Character)
	}
}

func TestErrorToDiagnostic_PathWithColons(t *testing.T) {
	// Windows-style paths or paths with colons in the filename portion
	// should still parse correctly if they match the regex
	err := errors.New("C:/Users/test/file.kuki:3:7: unexpected token")
	diag := errorToDiagnostic(err)

	// The regex matches from the right, so this should work
	if diag.Message != "unexpected token" {
		t.Errorf("expected 'unexpected token', got: %s", diag.Message)
	}
}

func TestErrorToDiagnostic_LargeLineNumbers(t *testing.T) {
	err := errors.New("big.kuki:1000:50: error at end of file")
	diag := errorToDiagnostic(err)

	// 0-indexed: 999, 49
	if diag.Range.Start.Line != 999 {
		t.Errorf("expected line 999, got %d", diag.Range.Start.Line)
	}
	if diag.Range.Start.Character != 49 {
		t.Errorf("expected character 49, got %d", diag.Range.Start.Character)
	}
}

func TestErrorToDiagnostic_MessageWithColons(t *testing.T) {
	err := errors.New("test.kuki:2:3: expected: int, got: string")
	diag := errorToDiagnostic(err)

	// Should capture everything after the position as the message
	if diag.Message != "expected: int, got: string" {
		t.Errorf("expected message with colons, got: %s", diag.Message)
	}
	if diag.Range.Start.Line != 1 { // 2 - 1
		t.Errorf("expected line 1, got %d", diag.Range.Start.Line)
	}
}
