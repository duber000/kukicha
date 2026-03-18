package codegen

import (
	"fmt"
	"testing"
)

// TestProperty_CodegenProducesValidGo verifies that a broad set of valid Kukicha
// programs all produce syntactically valid Go through the full pipeline.
// This extends the integration tests with greater breadth.
func TestProperty_CodegenProducesValidGo(t *testing.T) {
	programs := []struct {
		name   string
		source string
	}{
		{
			name:   "simple_function",
			source: "func Add(a int, b int) int\n    return a + b\n",
		},
		{
			name: "multiple_returns",
			source: `func Divide(a int, b int) (int, error)
    if b equals 0
        return 0, error "division by zero"
    return a / b, empty
`,
		},
		{
			name:   "type_and_method",
			source: "type User\n    name string\n    age int\n\nfunc GetName on u User string\n    return u.name\n",
		},
		{
			name:   "pointer_receiver",
			source: "type Counter\n    value int\n\nfunc Increment on c reference Counter\n    c.value = c.value + 1\n",
		},
		{
			name:   "string_interpolation",
			source: "func Greet(name string) string\n    return \"Hello {name}!\"\n",
		},
		{
			name:   "escaped_braces",
			source: "func Json() string\n    return \"\\{key\\}: value\"\n",
		},
		{
			name:   "if_else",
			source: "func Abs(x int) int\n    if x < 0\n        return -x\n    return x\n",
		},
		{
			name:   "if_else_chain",
			source: "func Classify(x int) string\n    if x > 100\n        return \"big\"\n    else if x > 0\n        return \"small\"\n    else\n        return \"non-positive\"\n",
		},
		{
			name:   "for_range",
			source: "func Sum(items list of int) int\n    total := 0\n    for item in items\n        total = total + item\n    return total\n",
		},
		{
			name:   "for_numeric_exclusive",
			source: "func Count() int\n    total := 0\n    for i from 0 to 10\n        total = total + i\n    return total\n",
		},
		{
			name:   "for_numeric_inclusive",
			source: "func Count() int\n    total := 0\n    for i from 0 through 10\n        total = total + i\n    return total\n",
		},
		{
			name:   "bare_switch",
			source: "func Describe(x int) string\n    switch\n        when x > 100\n            return \"big\"\n        otherwise\n            return \"small\"\n",
		},
		{
			name:   "value_switch",
			source: "func Handle(cmd string) string\n    switch cmd\n        when \"help\"\n            return \"showing help\"\n        when \"quit\"\n            return \"bye\"\n        otherwise\n            return \"unknown\"\n",
		},
		{
			name:   "list_literal",
			source: "func MakeList() list of string\n    return list of string{\"a\", \"b\", \"c\"}\n",
		},
		{
			name:   "map_literal",
			source: "func MakeMap() map of string to int\n    return map of string to int{\"x\": 1, \"y\": 2}\n",
		},
		{
			name:   "variadic",
			source: "func Sum(many numbers int) int\n    total := 0\n    for n in numbers\n        total = total + n\n    return total\n",
		},
		{
			name:   "boolean_operators",
			source: "func Check(a bool, b bool) bool\n    return a and b or not a\n",
		},
		{
			name:   "empty_function",
			source: "func Nothing()\n    return\n",
		},
		{
			name:   "multiple_functions",
			source: "func Add(a int, b int) int\n    return a + b\n\nfunc Sub(a int, b int) int\n    return a - b\n",
		},
		{
			name:   "lambda_expression",
			source: "func Apply(x int, f func(int) int) int\n    return f(x)\n",
		},
		{
			name: "type_with_json_tag",
			source: "type Todo\n    id int64\n    title string as \"title\"\n    done bool\n",
		},
		{
			name:   "negative_index",
			source: "func Last(items list of string) string\n    return items[-1]\n",
		},
		{
			name: "interface_decl",
			source: `interface Stringer
    String() string
`,
		},
	}

	for _, tc := range programs {
		t.Run(tc.name, func(t *testing.T) {
			output := fullPipeline(t, tc.source, "property.kuki")
			assertValidGo(t, output)
		})
	}
}

// TestProperty_EmptyProgramProducesValidGo verifies that programs with no
// meaningful declarations still produce valid Go output.
func TestProperty_EmptyProgramProducesValidGo(t *testing.T) {
	sources := []string{
		"",
		"\n",
		"\n\n\n",
		"# just a comment\n",
	}
	for i, src := range sources {
		t.Run(fmt.Sprintf("empty_%d", i), func(t *testing.T) {
			output := fullPipeline(t, src, "empty.kuki")
			assertValidGo(t, output)
		})
	}
}
