package parser

import "testing"

func FuzzParser(f *testing.F) {
	// Seed corpus: valid Kukicha programs
	seeds := []string{
		// Simple function
		"func Add(a int, b int) int\n    return a + b\n",
		// Function with default params
		"func Greet(name string, greeting string = \"Hello\") string\n    return \"{greeting}, {name}!\"\n",
		// Type declaration
		"type User\n    name string\n    age int\n",
		// Method
		"type User\n    name string\n\nfunc GetName on u User string\n    return u.name\n",
		// Import
		"import \"stdlib/slice\"\n",
		"import \"stdlib/json\" as jsonpkg\n",
		// Variables
		"func Main()\n    x := 42\n    y := \"hello\"\n",
		// Pipes
		"func Main()\n    result := data |> parse() |> transform()\n",
		// Error handling
		"func Main()\n    data := fetchData() onerr panic \"failed\"\n",
		"func Main()\n    data := fetchData() onerr return\n",
		"func Main()\n    port := getPort() onerr 8080\n",
		// If/else
		"func Check(x int) string\n    if x equals 0\n        return \"zero\"\n    else if x < 0\n        return \"negative\"\n    else\n        return \"positive\"\n",
		// For loops
		"func Main()\n    for i from 0 to 10\n        print(i)\n",
		"func Main()\n    for item in items\n        process(item)\n",
		// Switch
		"func Main()\n    switch command\n        when \"help\"\n            showHelp()\n        otherwise\n            print(\"unknown\")\n",
		// Collections
		"func Main()\n    items := list of string{\"a\", \"b\"}\n    config := map of string to int{\"port\": 8080}\n",
		// Lambdas
		"func Main()\n    f := (x int) => x + 1\n",
		// Concurrency
		"func Main()\n    ch := make channel of string\n    send \"msg\" to ch\n    msg := receive from ch\n",
		// Multiple functions
		"func Add(a int, b int) int\n    return a + b\n\nfunc Sub(a int, b int) int\n    return a - b\n",
		// String interpolation
		"func Main()\n    name := \"world\"\n    print(\"hello {name}\")\n",
		// Edge cases
		"",
		"\n",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, data string) {
		p, err := New(data, "fuzz.kuki")
		if err != nil {
			return // lexer error is fine
		}
		p.Parse() // must not panic; parse errors are OK
	})
}
