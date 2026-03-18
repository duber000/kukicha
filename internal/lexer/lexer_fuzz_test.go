package lexer

import "testing"

func FuzzLexer(f *testing.F) {
	// Seed corpus: valid Kukicha snippets covering key syntax
	seeds := []string{
		// Functions
		"func Add(a int, b int) int\n    return a + b\n",
		"func Greet(name string, greeting string = \"Hello\") string\n    return \"{greeting}, {name}!\"\n",
		// Variables
		"x := 42\n",
		"count := 0\ncount = 100\n",
		// Strings and interpolation
		"name := \"hello {world}\"\n",
		"path := \"{dir}\\sep{file}\"\n",
		"json := \"key: \\{value\\}\"\n",
		// Collections
		"items := list of string{\"a\", \"b\", \"c\"}\n",
		"config := map of string to int{\"port\": 8080}\n",
		// Control flow
		"if count equals 0\n    return \"empty\"\n",
		"for i from 0 to 10\n    print(i)\n",
		"for item in items\n    process(item)\n",
		// Pipes
		"result := data |> parse() |> transform()\n",
		// Error handling
		"data := f() onerr panic \"fail\"\n",
		"data := f() onerr return\n",
		"port := getPort() onerr 8080\n",
		// Comments and directives
		"# comment\n",
		"# kuki:deprecated \"Use NewFunc\"\n",
		// Types
		"type Todo\n    id int64\n    title string\n",
		// Switch
		"switch command\n    when \"help\"\n        showHelp()\n    otherwise\n        print(\"unknown\")\n",
		// Concurrency
		"ch := make channel of string\nsend \"msg\" to ch\n",
		// Lambdas
		"f := (x int) => x + 1\n",
		// Imports
		"import \"stdlib/slice\"\n",
		// Empty/edge cases
		"",
		"\n",
		"\n\n\n",
		"    ",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, data string) {
		l := NewLexer(data, "fuzz.kuki")
		l.ScanTokens() // must not panic; errors are OK
	})
}
