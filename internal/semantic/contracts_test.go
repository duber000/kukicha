package semantic

import (
	"strings"
	"testing"
)

func TestContractDirective_RequiresValid(t *testing.T) {
	source := `# kuki:requires "len(items) > 0"
func First(items list of string) string
    return items[0]
`
	_, errs := analyzeSource(t, source)
	if len(errs) > 0 {
		t.Fatalf("valid requires directive should not produce errors: %v", errs)
	}
}

func TestContractDirective_RequiresNoArg(t *testing.T) {
	source := `# kuki:requires
func First(items list of string) string
    return items[0]
`
	_, errs := analyzeSource(t, source)
	if !containsError(errs, "requires a condition expression") {
		t.Fatalf("requires without args should produce error, got: %v", errs)
	}
}

func TestContractDirective_EnsuresValid(t *testing.T) {
	source := `# kuki:ensures "result >= 0"
func Abs(x int) int
    if x < 0
        return -x
    return x
`
	_, errs := analyzeSource(t, source)
	if len(errs) > 0 {
		t.Fatalf("valid ensures directive should not produce errors: %v", errs)
	}
}

func TestContractDirective_EnsuresNoReturnValues(t *testing.T) {
	source := `# kuki:ensures "result >= 0"
func Process(x int)
    return
`
	_, errs := analyzeSource(t, source)
	if !containsError(errs, "requires the function to have return values") {
		t.Fatalf("ensures on void function should produce error, got: %v", errs)
	}
}

func TestContractDirective_EnsuresNoArg(t *testing.T) {
	source := `# kuki:ensures
func Abs(x int) int
    return x
`
	_, errs := analyzeSource(t, source)
	if !containsError(errs, "requires a condition expression") {
		t.Fatalf("ensures without args should produce error, got: %v", errs)
	}
}

func TestContractDirective_InvariantValid(t *testing.T) {
	source := `# kuki:invariant "self.min <= self.max"
type Range
    min int
    max int
`
	_, errs := analyzeSource(t, source)
	if len(errs) > 0 {
		t.Fatalf("valid invariant directive should not produce errors: %v", errs)
	}
}

func TestContractDirective_InvariantNoArg(t *testing.T) {
	source := `# kuki:invariant
type Range
    min int
    max int
`
	_, errs := analyzeSource(t, source)
	if !containsError(errs, "requires a condition expression") {
		t.Fatalf("invariant without args should produce error, got: %v", errs)
	}
}

func TestContractDirective_InvariantUnknownField(t *testing.T) {
	source := `# kuki:invariant "self.minimum <= self.max"
type Range
    min int
    max int
`
	_, errs := analyzeSource(t, source)
	if !containsError(errs, "unknown field 'self.minimum'") {
		t.Fatalf("invariant referencing unknown field should produce error, got: %v", errs)
	}
}

func TestContractDirective_InvariantOnAlias(t *testing.T) {
	source := `# kuki:invariant "self.x > 0"
type Handler func(string)
`
	_, errs := analyzeSource(t, source)
	if !containsError(errs, "cannot be applied to type aliases") {
		t.Fatalf("invariant on alias should produce error, got: %v", errs)
	}
}

// --- Tier 5 tests ---

func TestAgentSecurity_UnboundedLoop(t *testing.T) {
	source := `import "stdlib/http"

func Handler(w http.ResponseWriter, r reference http.Request)
    for true
        w = w
`
	analyzer, _ := analyzeSource(t, source)
	warnings := analyzer.Warnings()
	found := false
	for _, w := range warnings {
		if strings.Contains(w.Error(), "unbounded loop") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected unbounded loop warning in HTTP handler, got warnings: %v", warnings)
	}
}

func TestAgentSecurity_UnboundedLoopWithBreak(t *testing.T) {
	// A loop with break should NOT trigger the warning
	source := `import "stdlib/http"

func Handler(w http.ResponseWriter, r reference http.Request)
    for true
        break
`
	analyzer, _ := analyzeSource(t, source)
	warnings := analyzer.Warnings()
	for _, w := range warnings {
		if strings.Contains(w.Error(), "unbounded loop") {
			t.Fatalf("loop with break should not trigger unbounded loop warning, got: %v", w)
		}
	}
}

func TestAgentSecurity_UnboundedLoopOutsideHandler(t *testing.T) {
	// Outside HTTP handler should NOT trigger
	source := `func Worker()
    for true
        x := 1
        _ = x
`
	analyzer, _ := analyzeSource(t, source)
	warnings := analyzer.Warnings()
	for _, w := range warnings {
		if strings.Contains(w.Error(), "unbounded loop") {
			t.Fatalf("loop outside handler should not trigger warning, got: %v", w)
		}
	}
}

func TestAgentSecurity_PrivilegeEscalation(t *testing.T) {
	source := `import "stdlib/http"
import "stdlib/shell"

func Handler(w http.ResponseWriter, r reference http.Request)
    shell.Run("ls")
`
	analyzer, _ := analyzeSource(t, source)
	warnings := analyzer.Warnings()
	found := false
	for _, w := range warnings {
		if strings.Contains(w.Error(), "privilege escalation") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected privilege escalation warning, got warnings: %v", warnings)
	}
}
