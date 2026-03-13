package semantic

import (
	"strings"
	"testing"

	"github.com/duber000/kukicha/internal/parser"
)

// ---------------------------------------------------------------------------
// Phase 3B: # kuki:deprecated warnings
// ---------------------------------------------------------------------------

func TestDeprecatedFunctionWarning(t *testing.T) {
	input := `# kuki:deprecated "Use NewFunc instead"
func OldFunc() string
    return "old"

func main()
    result := OldFunc()
    print(result)
`
	_, warnings := analyzeInputWithFile(t, input, "test.kuki")
	found := false
	for _, w := range warnings {
		if strings.Contains(w.Error(), "deprecated") && strings.Contains(w.Error(), "OldFunc") {
			found = true
			if !strings.Contains(w.Error(), "Use NewFunc instead") {
				t.Errorf("expected deprecation message, got: %s", w)
			}
			break
		}
	}
	if !found {
		t.Errorf("expected deprecation warning for OldFunc, got warnings: %v", warnings)
	}
}

func TestDeprecatedFunctionNoWarningWhenNotCalled(t *testing.T) {
	input := `# kuki:deprecated "Use NewFunc instead"
func OldFunc() string
    return "old"

func main()
    print("hello")
`
	_, warnings := analyzeInputWithFile(t, input, "test.kuki")
	for _, w := range warnings {
		if strings.Contains(w.Error(), "deprecated") {
			t.Errorf("unexpected deprecation warning when function not called: %v", w)
		}
	}
}

func TestDeprecatedFunctionMultipleCalls(t *testing.T) {
	input := `# kuki:deprecated "Use NewFunc"
func OldFunc() string
    return "old"

func main()
    a := OldFunc()
    b := OldFunc()
    print(a)
    print(b)
`
	_, warnings := analyzeInputWithFile(t, input, "test.kuki")
	count := 0
	for _, w := range warnings {
		if strings.Contains(w.Error(), "deprecated") && strings.Contains(w.Error(), "OldFunc") {
			count++
		}
	}
	if count != 2 {
		t.Errorf("expected 2 deprecation warnings (one per call), got %d; warnings: %v", count, warnings)
	}
}

func TestNonDeprecatedFunctionNoWarning(t *testing.T) {
	input := `func GoodFunc() string
    return "good"

func main()
    result := GoodFunc()
    print(result)
`
	_, warnings := analyzeInputWithFile(t, input, "test.kuki")
	for _, w := range warnings {
		if strings.Contains(w.Error(), "deprecated") {
			t.Errorf("unexpected deprecation warning: %v", w)
		}
	}
}

func TestDeprecatedTypeWarning(t *testing.T) {
	// For now, type deprecation is tracked but not yet warned on at usage sites.
	// This test documents the current behavior.
	input := `# kuki:deprecated "Use NewUser instead"
type OldUser
    name string

func main()
    print("hello")
`
	p, err := parser.New(input, "test.kuki")
	if err != nil {
		t.Fatalf("parser init error: %v", err)
	}
	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}
	analyzer := NewWithFile(program, "test.kuki")
	_ = analyzer.Analyze()
	// Verify the type was registered as deprecated
	if msg, ok := analyzer.deprecatedTypes["OldUser"]; !ok {
		t.Error("expected OldUser to be in deprecatedTypes map")
	} else if msg != "Use NewUser instead" {
		t.Errorf("expected deprecation message 'Use NewUser instead', got %q", msg)
	}
}
