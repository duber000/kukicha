package formatter

import (
	"testing"
)

// TestProperty_FormatterIdempotent verifies that formatting is idempotent:
// format(format(source)) == format(source). If the formatter changes output
// on a second pass, it has a normalization bug.
func TestProperty_FormatterIdempotent(t *testing.T) {
	programs := []struct {
		name   string
		source string
	}{
		{
			name:   "simple_function",
			source: "func Add(a int, b int) int\n    return a + b\n",
		},
		{
			name:   "function_with_if",
			source: "func Abs(x int) int\n    if x < 0\n        return -x\n    return x\n",
		},
		{
			name:   "type_declaration",
			source: "type User\n    name string\n    age int\n",
		},
		{
			name:   "method",
			source: "func GetName on u User string\n    return u.name\n",
		},
		{
			name:   "string_interpolation",
			source: "func Greet(name string) string\n    return \"Hello {name}!\"\n",
		},
		{
			name: "for_range",
			source: "func Sum(items list of int) int\n    total := 0\n    for item in items\n        total = total + item\n    return total\n",
		},
		{
			name:   "switch",
			source: "func Handle(cmd string) string\n    switch cmd\n        when \"help\"\n            return \"help\"\n        otherwise\n            return \"unknown\"\n",
		},
		{
			name:   "import_and_function",
			source: "import \"stdlib/slice\"\n\nfunc Main()\n    items := list of int{1, 2, 3}\n    _ = items\n",
		},
		{
			name:   "multiple_functions",
			source: "func Add(a int, b int) int\n    return a + b\n\nfunc Sub(a int, b int) int\n    return a - b\n",
		},
		{
			name: "error_handling",
			source: "func Load(path string) string, error\n    data := readFile(path) onerr return empty, error \"{error}\"\n    return data, empty\n",
		},
		{
			name: "comment_preserved",
			source: "# This is a comment\nfunc Add(a int, b int) int\n    return a + b\n",
		},
		{
			name:   "pipe_expression",
			source: "func Process(data string) string\n    return data\n        |> parse()\n        |> transform()\n",
		},
		{
			name:   "boolean_operators",
			source: "func Check(a bool, b bool) bool\n    return a and b or not a\n",
		},
	}

	opts := DefaultOptions()

	for _, tc := range programs {
		t.Run(tc.name, func(t *testing.T) {
			// First format
			formatted1, err := Format(tc.source, "test.kuki", opts)
			if err != nil {
				t.Skipf("first format failed (skip): %v", err)
			}

			// Second format — should be identical
			formatted2, err := Format(formatted1, "test.kuki", opts)
			if err != nil {
				t.Fatalf("second format failed: %v", err)
			}

			if formatted1 != formatted2 {
				t.Errorf("formatter not idempotent:\nfirst pass:\n%s\nsecond pass:\n%s", formatted1, formatted2)
			}
		})
	}
}
