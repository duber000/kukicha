package codegen

import (
	"strings"
	"testing"
)

func TestGenerateStringFromParts_SimpleInterpolation(t *testing.T) {
	input := `func Greet(name string) string
    return "Hello {name}!"
`
	output := generateSource(t, input)

	if !strings.Contains(output, "fmt.Sprintf") {
		t.Errorf("expected fmt.Sprintf, got: %s", output)
	}
	if !strings.Contains(output, `"Hello %v!"`) {
		t.Errorf("expected format string with %%v, got: %s", output)
	}
	if !strings.Contains(output, "name") {
		t.Errorf("expected 'name' as arg, got: %s", output)
	}
}

func TestGenerateStringFromParts_PipeExpr(t *testing.T) {
	input := `func Show(name string) string
    return "{name |> toUpper()}"
`
	output := generateSource(t, input)

	// Pipe should be resolved to a function call
	if strings.Contains(output, "|>") {
		t.Errorf("pipe should not appear in generated Go: %s", output)
	}
	if !strings.Contains(output, "toUpper(name)") {
		t.Errorf("expected pipe resolved to function call, got: %s", output)
	}
}

func TestGenerateStringFromParts_TypeCast(t *testing.T) {
	input := `func Show(x int) string
    return "{x as string}"
`
	output := generateSource(t, input)

	if !strings.Contains(output, "string(x)") {
		t.Errorf("expected type cast string(x), got: %s", output)
	}
}

func TestGenerateStringFromParts_MultipleExprs(t *testing.T) {
	input := `func Fmt(a int, b int) string
    return "{a} and {b}"
`
	output := generateSource(t, input)

	if !strings.Contains(output, "fmt.Sprintf") {
		t.Errorf("expected fmt.Sprintf, got: %s", output)
	}
	if !strings.Contains(output, "%v and %v") {
		t.Errorf("expected format with two %%v, got: %s", output)
	}
}

func TestGenerateStringFromParts_OnErrErrorSubstitution(t *testing.T) {
	input := `func readFile(path string) (string, error)
    return path, empty

func Process(path string) (string, error)
    data := readFile(path) onerr
        return "", error "read failed: {error}"
    return data, empty
`
	output := generateSource(t, input)

	// {error} inside onerr should be substituted with the generated error variable
	if !strings.Contains(output, "fmt.Sprintf") || !strings.Contains(output, "read failed:") {
		t.Errorf("expected onerr error interpolation, got: %s", output)
	}
	// Should NOT contain literal "error" as an argument (it should be the err variable)
	if strings.Contains(output, `Sprintf("read failed: %v", error)`) {
		t.Errorf("expected error variable substitution, not literal 'error', got: %s", output)
	}
}

func TestGenerateStringFromParts_NoInterpolation(t *testing.T) {
	input := `func Plain() string
    return "no interpolation here"
`
	output := generateSource(t, input)

	// Should NOT use fmt.Sprintf for plain strings
	if strings.Contains(output, "fmt.Sprintf") {
		t.Errorf("plain string should not use fmt.Sprintf, got: %s", output)
	}
}
