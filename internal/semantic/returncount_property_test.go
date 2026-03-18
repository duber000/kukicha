package semantic

import (
	"testing"
)

// TestProperty_StdlibReturnCountMatchesTypes verifies that for every entry in the
// generated Kukicha stdlib registry, the declared Count matches len(Types).
// A mismatch indicates a bug in genstdlibregistry.
func TestProperty_StdlibReturnCountMatchesTypes(t *testing.T) {
	for name, entry := range generatedStdlibRegistry {
		if entry.Count != len(entry.Types) {
			t.Errorf("stdlib registry %q: Count=%d but len(Types)=%d",
				name, entry.Count, len(entry.Types))
		}
	}
}

// TestProperty_GoStdlibReturnCountMatchesTypes verifies the same property for
// the Go stdlib registry.
func TestProperty_GoStdlibReturnCountMatchesTypes(t *testing.T) {
	for name, entry := range generatedGoStdlib {
		if entry.Count != len(entry.Types) {
			t.Errorf("Go stdlib registry %q: Count=%d but len(Types)=%d",
				name, entry.Count, len(entry.Types))
		}
	}
}

// TestProperty_StdlibParamNamesConsistent verifies that Kukicha stdlib entries
// with parameters have non-empty parameter names. Functions with zero parameters
// (like datetime.Now()) are expected to have empty ParamNames.
func TestProperty_StdlibParamNamesConsistent(t *testing.T) {
	for name, entry := range generatedStdlibRegistry {
		for i, pname := range entry.ParamNames {
			if pname == "" {
				t.Errorf("stdlib registry %q: ParamNames[%d] is empty string", name, i)
			}
		}
	}
}

// TestProperty_SecurityFunctionsExistInRegistry verifies that every function
// registered as security-sensitive either exists in the stdlib registry or is
// a void function (no return value, so excluded from the registry).
// Void security functions (like http.Redirect) are legitimate — they perform
// side effects and are checked for argument safety, not return values.
func TestProperty_SecurityFunctionsExistInRegistry(t *testing.T) {
	// Known void security functions that are intentionally excluded from the
	// stdlib return-count registry (they have no return value).
	knownVoid := map[string]bool{
		"http.Redirect":          true,
		"http.RedirectPermanent": true,
	}
	for name, category := range generatedSecurityFunctions {
		if knownVoid[name] {
			continue
		}
		if _, ok := generatedStdlibRegistry[name]; !ok {
			t.Errorf("security function %q (category %q) not found in stdlib registry and not in knownVoid list",
				name, category)
		}
	}
}

// TestProperty_StdlibCountPositive verifies all return counts are positive.
func TestProperty_StdlibCountPositive(t *testing.T) {
	for name, entry := range generatedStdlibRegistry {
		if entry.Count <= 0 {
			t.Errorf("stdlib registry %q: Count=%d (should be > 0)", name, entry.Count)
		}
	}
	for name, entry := range generatedGoStdlib {
		if entry.Count <= 0 {
			t.Errorf("Go stdlib registry %q: Count=%d (should be > 0)", name, entry.Count)
		}
	}
}
