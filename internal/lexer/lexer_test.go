package lexer

import (
	"testing"
)

func TestBasicTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:  "simple function",
			input: "func Hello()\n",
			expected: []TokenType{
				TOKEN_FUNC, TOKEN_IDENTIFIER, TOKEN_LPAREN, TOKEN_RPAREN, TOKEN_NEWLINE, TOKEN_EOF,
			},
		},
		{
			name:  "variable declaration",
			input: "count := 42\n",
			expected: []TokenType{
				TOKEN_IDENTIFIER, TOKEN_WALRUS, TOKEN_INTEGER, TOKEN_NEWLINE, TOKEN_EOF,
			},
		},
		{
			name:  "assignment",
			input: "count = 10\n",
			expected: []TokenType{
				TOKEN_IDENTIFIER, TOKEN_ASSIGN, TOKEN_INTEGER, TOKEN_NEWLINE, TOKEN_EOF,
			},
		},
		{
			name:  "keywords",
			input: "petiole import type interface\n",
			expected: []TokenType{
				TOKEN_PETIOLE, TOKEN_IMPORT, TOKEN_TYPE, TOKEN_INTERFACE, TOKEN_NEWLINE, TOKEN_EOF,
			},
		},
		{
			name:  "operators",
			input: "+ - * / %\n",
			expected: []TokenType{
				TOKEN_PLUS, TOKEN_MINUS, TOKEN_STAR, TOKEN_SLASH, TOKEN_PERCENT, TOKEN_NEWLINE, TOKEN_EOF,
			},
		},
		{
			name:  "comparison operators",
			input: "< > <= >= == !=\n",
			expected: []TokenType{
				TOKEN_LT, TOKEN_GT, TOKEN_LTE, TOKEN_GTE, TOKEN_DOUBLE_EQUALS, TOKEN_NOT_EQUALS, TOKEN_NEWLINE, TOKEN_EOF,
			},
		},
		{
			name:  "pipe operator",
			input: "data |> process()\n",
			expected: []TokenType{
				TOKEN_IDENTIFIER, TOKEN_PIPE, TOKEN_IDENTIFIER, TOKEN_LPAREN, TOKEN_RPAREN, TOKEN_NEWLINE, TOKEN_EOF,
			},
		},
		{
			name:  "boolean operators",
			input: "and or not\n",
			expected: []TokenType{
				TOKEN_AND, TOKEN_OR, TOKEN_NOT, TOKEN_NEWLINE, TOKEN_EOF,
			},
		},
		{
			name:  "go-style boolean operators",
			input: "&& ||\n",
			expected: []TokenType{
				TOKEN_AND_AND, TOKEN_OR_OR, TOKEN_NEWLINE, TOKEN_EOF,
			},
		},
		{
			name:  "channel operators",
			input: "send receive <-\n",
			expected: []TokenType{
				TOKEN_SEND, TOKEN_RECEIVE, TOKEN_ARROW_LEFT, TOKEN_NEWLINE, TOKEN_EOF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input, "test.kuki")
			tokens, err := lexer.ScanTokens()

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(tokens) != len(tt.expected) {
				t.Fatalf("Expected %d tokens, got %d", len(tt.expected), len(tokens))
			}

			for i, expected := range tt.expected {
				if tokens[i].Type != expected {
					t.Errorf("Token %d: expected %s, got %s", i, expected, tokens[i].Type)
				}
			}
		})
	}
}

func TestIndentation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name: "simple indent",
			input: `func Hello()
    print "hi"
`,
			expected: []TokenType{
				TOKEN_FUNC, TOKEN_IDENTIFIER, TOKEN_LPAREN, TOKEN_RPAREN, TOKEN_NEWLINE,
				TOKEN_INDENT, TOKEN_IDENTIFIER, TOKEN_STRING, TOKEN_NEWLINE,
				TOKEN_DEDENT, TOKEN_EOF,
			},
		},
		{
			name: "multiple indent levels",
			input: `if condition
    if nested
        doSomething()
`,
			expected: []TokenType{
				TOKEN_IF, TOKEN_IDENTIFIER, TOKEN_NEWLINE,
				TOKEN_INDENT, TOKEN_IF, TOKEN_IDENTIFIER, TOKEN_NEWLINE,
				TOKEN_INDENT, TOKEN_IDENTIFIER, TOKEN_LPAREN, TOKEN_RPAREN, TOKEN_NEWLINE,
				TOKEN_DEDENT, TOKEN_DEDENT, TOKEN_EOF,
			},
		},
		{
			name: "dedent to previous level",
			input: `if a
    if b
        do1()
    do2()
`,
			expected: []TokenType{
				TOKEN_IF, TOKEN_IDENTIFIER, TOKEN_NEWLINE,
				TOKEN_INDENT, TOKEN_IF, TOKEN_IDENTIFIER, TOKEN_NEWLINE,
				TOKEN_INDENT, TOKEN_IDENTIFIER, TOKEN_LPAREN, TOKEN_RPAREN, TOKEN_NEWLINE,
				TOKEN_DEDENT, TOKEN_IDENTIFIER, TOKEN_LPAREN, TOKEN_RPAREN, TOKEN_NEWLINE,
				TOKEN_DEDENT, TOKEN_EOF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input, "test.kuki")
			tokens, err := lexer.ScanTokens()

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(tokens) != len(tt.expected) {
				t.Fatalf("Expected %d tokens, got %d\nTokens: %v", len(tt.expected), len(tokens), tokens)
			}

			for i, expected := range tt.expected {
				if tokens[i].Type != expected {
					t.Errorf("Token %d: expected %s, got %s", i, expected, tokens[i].Type)
				}
			}
		})
	}
}

func TestStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    `"hello"`,
			expected: "hello",
		},
		{
			name:     "string with interpolation",
			input:    `"Hello {name}"`,
			expected: "Hello {name}",
		},
		{
			name:     "string with escape sequences",
			input:    `"line1\nline2"`,
			expected: "line1\nline2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input, "test.kuki")
			tokens, err := lexer.ScanTokens()

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(tokens) < 1 || tokens[0].Type != TOKEN_STRING {
				t.Fatalf("Expected STRING token")
			}

			if tokens[0].Lexeme != tt.expected {
				t.Errorf("Expected string %q, got %q", tt.expected, tokens[0].Lexeme)
			}
		})
	}
}

func TestRuneLiterals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple rune",
			input:    `'a'`,
			expected: "a",
		},
		{
			name:     "newline escape",
			input:    `'\n'`,
			expected: "\n",
		},
		{
			name:     "tab escape",
			input:    `'\t'`,
			expected: "\t",
		},
		{
			name:     "single quote escape",
			input:    `'\''`,
			expected: "'",
		},
		{
			name:     "digit rune",
			input:    `'0'`,
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input, "test.kuki")
			tokens, err := lexer.ScanTokens()

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(tokens) < 1 || tokens[0].Type != TOKEN_RUNE {
				t.Fatalf("Expected RUNE token, got %v", tokens[0].Type)
			}

			if tokens[0].Lexeme != tt.expected {
				t.Errorf("Expected rune %q, got %q", tt.expected, tokens[0].Lexeme)
			}
		})
	}
}

func TestNumbers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected TokenType
	}{
		{
			name:     "integer",
			input:    "42",
			expected: TOKEN_INTEGER,
		},
		{
			name:     "float",
			input:    "3.14",
			expected: TOKEN_FLOAT,
		},
		{
			name:     "zero",
			input:    "0",
			expected: TOKEN_INTEGER,
		},
		{
			name:     "large number",
			input:    "123456789",
			expected: TOKEN_INTEGER,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input, "test.kuki")
			tokens, err := lexer.ScanTokens()

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(tokens) < 1 {
				t.Fatalf("Expected at least one token")
			}

			if tokens[0].Type != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tokens[0].Type)
			}
		})
	}
}

func TestComments(t *testing.T) {
	input := `# This is a comment
func Hello()
    # Another comment
    print "hi"
`

	lexer := NewLexer(input, "test.kuki")
	tokens, err := lexer.ScanTokens()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Comments should be skipped
	for _, token := range tokens {
		if token.Type == TOKEN_IDENTIFIER && token.Lexeme[0] == '#' {
			t.Errorf("Comment was not skipped: %v", token)
		}
	}
}

func TestRealWorldExample(t *testing.T) {
	input := `petiole todo

import time

type Todo
    id int64
    title string
    completed bool

func CreateTodo(id, title)
    return Todo
        id: id
        title: title
        completed: false

func Display on todo Todo
    status := "pending"
    if todo.completed
        status = "done"
    return "{status}: {todo.title}"
`

	lexer := NewLexer(input, "test.kuki")
	tokens, err := lexer.ScanTokens()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Just verify we got tokens without errors
	if len(tokens) == 0 {
		t.Fatalf("Expected tokens, got none")
	}

	// Check that we have the main keywords
	foundPetiole := false
	foundImport := false
	foundType := false
	foundFunc := false

	for _, token := range tokens {
		switch token.Type {
		case TOKEN_PETIOLE:
			foundPetiole = true
		case TOKEN_IMPORT:
			foundImport = true
		case TOKEN_TYPE:
			foundType = true
		case TOKEN_FUNC:
			foundFunc = true
		}
	}

	if !foundPetiole {
		t.Error("Expected to find PITOLE token")
	}
	if !foundImport {
		t.Error("Expected to find IMPORT token")
	}
	if !foundType {
		t.Error("Expected to find TYPE token")
	}
	if !foundFunc {
		t.Error("Expected to find FUNC token")
	}
}

func TestPipeContinuation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			// Trailing |> suppresses NEWLINE; continuation-line indentation
			// does not produce INDENT/DEDENT.
			name: "single continuation",
			input: "func Test() string\n    return x |>\n        ToUpper()\n",
			expected: []TokenType{
				TOKEN_FUNC, TOKEN_IDENTIFIER, TOKEN_LPAREN, TOKEN_RPAREN, TOKEN_IDENTIFIER, TOKEN_NEWLINE,
				TOKEN_INDENT, TOKEN_RETURN, TOKEN_IDENTIFIER, TOKEN_PIPE,
				TOKEN_IDENTIFIER, TOKEN_LPAREN, TOKEN_RPAREN, TOKEN_NEWLINE,
				TOKEN_DEDENT, TOKEN_EOF,
			},
		},
		{
			// Two trailing |> in a row; only one INDENT/DEDENT pair at the
			// end when the chain returns to the base indentation.
			name: "chained continuation",
			input: "func Test() string\n    return x |>\n        f() |>\n        g()\n",
			expected: []TokenType{
				TOKEN_FUNC, TOKEN_IDENTIFIER, TOKEN_LPAREN, TOKEN_RPAREN, TOKEN_IDENTIFIER, TOKEN_NEWLINE,
				TOKEN_INDENT, TOKEN_RETURN, TOKEN_IDENTIFIER, TOKEN_PIPE,
				TOKEN_IDENTIFIER, TOKEN_LPAREN, TOKEN_RPAREN, TOKEN_PIPE,
				TOKEN_IDENTIFIER, TOKEN_LPAREN, TOKEN_RPAREN, TOKEN_NEWLINE,
				TOKEN_DEDENT, TOKEN_EOF,
			},
		},
		{
			// Leading |> at column 0 (no indentation before the pipe).
			// isPipeAtStartOfNextLine must start scanning at l.current,
			// not l.current+1, or the '|' is skipped.
			name: "leading pipe at column zero",
			input: "x := y\n|> foo()\n",
			expected: []TokenType{
				TOKEN_IDENTIFIER, TOKEN_WALRUS, TOKEN_IDENTIFIER, TOKEN_PIPE,
				TOKEN_IDENTIFIER, TOKEN_LPAREN, TOKEN_RPAREN, TOKEN_NEWLINE,
				TOKEN_EOF,
			},
		},
		{
			// A comment after the trailing |> is transparent; the next
			// line is still treated as a continuation.
			name: "comment after pipe",
			input: "func Test() string\n    return x |> # comment\n        ToUpper()\n",
			expected: []TokenType{
				TOKEN_FUNC, TOKEN_IDENTIFIER, TOKEN_LPAREN, TOKEN_RPAREN, TOKEN_IDENTIFIER, TOKEN_NEWLINE,
				TOKEN_INDENT, TOKEN_RETURN, TOKEN_IDENTIFIER, TOKEN_PIPE, TOKEN_COMMENT,
				TOKEN_IDENTIFIER, TOKEN_LPAREN, TOKEN_RPAREN, TOKEN_NEWLINE,
				TOKEN_DEDENT, TOKEN_EOF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input, "test.kuki")
			tokens, err := lexer.ScanTokens()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if len(tokens) != len(tt.expected) {
				types := make([]string, len(tokens))
				for i, tok := range tokens {
					types[i] = tok.Type.String()
				}
				t.Fatalf("Expected %d tokens, got %d\nGot: %v", len(tt.expected), len(tokens), types)
			}
			for i, expected := range tt.expected {
				if tokens[i].Type != expected {
					t.Errorf("Token %d: expected %s, got %s", i, expected, tokens[i].Type)
				}
			}
		})
	}
}

func TestErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "unterminated string",
			input: `"hello`,
		},
		{
			name: "tabs for indentation",
			input: `func test()
	print "bad"
`,
		},
		{
			name: "invalid indentation",
			input: `func test()
  print "bad"
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input, "test.kuki")
			_, err := lexer.ScanTokens()

			if err == nil {
				t.Errorf("Expected error for %s, got none", tt.name)
			}
		})
	}
}

func TestKeywordRecognition(t *testing.T) {
	keywords := []struct {
		keyword string
		token   TokenType
	}{
		{"petiole", TOKEN_PETIOLE},
		{"import", TOKEN_IMPORT},
		{"type", TOKEN_TYPE},
		{"interface", TOKEN_INTERFACE},
		{"func", TOKEN_FUNC},
		{"return", TOKEN_RETURN},
		{"if", TOKEN_IF},
		{"else", TOKEN_ELSE},
		{"for", TOKEN_FOR},
		{"in", TOKEN_IN},
		{"from", TOKEN_FROM},
		{"to", TOKEN_TO},
		{"through", TOKEN_THROUGH},
		{"go", TOKEN_GO},
		{"defer", TOKEN_DEFER},
		{"make", TOKEN_MAKE},
		{"list", TOKEN_LIST},
		{"map", TOKEN_MAP},
		{"channel", TOKEN_CHANNEL},
		{"send", TOKEN_SEND},
		{"receive", TOKEN_RECEIVE},
		{"panic", TOKEN_PANIC},
		{"recover", TOKEN_RECOVER},
		{"empty", TOKEN_EMPTY},
		{"nil", TOKEN_EMPTY}, // nil is an alias for empty
		{"reference", TOKEN_REFERENCE},
		{"on", TOKEN_ON},
		{"discard", TOKEN_DISCARD},
		{"at", TOKEN_AT},
		{"of", TOKEN_OF},
		{"true", TOKEN_TRUE},
		{"false", TOKEN_FALSE},
		{"equals", TOKEN_EQUALS},
		{"and", TOKEN_AND},
		{"or", TOKEN_OR},
		{"not", TOKEN_NOT},
	}

	for _, kw := range keywords {
		t.Run(kw.keyword, func(t *testing.T) {
			lexer := NewLexer(kw.keyword, "test.kuki")
			tokens, err := lexer.ScanTokens()

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(tokens) < 1 {
				t.Fatalf("Expected at least one token")
			}

			if tokens[0].Type != kw.token {
				t.Errorf("Expected %s, got %s for keyword %s", kw.token, tokens[0].Type, kw.keyword)
			}
		})
	}
}

func TestClosureInFunctionCall(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name: "simple closure in function argument",
			input: `filtered := items |> Filter(func(x int) bool
    return x > 2
)`,
			expected: []TokenType{
				TOKEN_IDENTIFIER,  // filtered
				TOKEN_WALRUS,      // :=
				TOKEN_IDENTIFIER,  // items
				TOKEN_PIPE,        // |>
				TOKEN_IDENTIFIER,  // Filter
				TOKEN_LPAREN,      // (
				TOKEN_FUNC,        // func
				TOKEN_LPAREN,      // (
				TOKEN_IDENTIFIER,  // x
				TOKEN_IDENTIFIER,  // int
				TOKEN_RPAREN,      // )
				TOKEN_IDENTIFIER,  // bool
				TOKEN_NEWLINE,     // newline after return type
				TOKEN_INDENT,      // indentation for closure body
				TOKEN_RETURN,      // return
				TOKEN_IDENTIFIER,  // x
				TOKEN_GT,          // >
				TOKEN_INTEGER,     // 2
				TOKEN_NEWLINE,     // newline after return
				TOKEN_DEDENT,      // dedent from closure body
				TOKEN_RPAREN,      // ) closing function call
				TOKEN_EOF,
			},
		},
		{
			name: "closure with multiple lines",
			input: `result := data |> Map(func(item string) string
    trimmed := trim(item)
    return trimmed
)`,
			expected: []TokenType{
				TOKEN_IDENTIFIER,  // result
				TOKEN_WALRUS,      // :=
				TOKEN_IDENTIFIER,  // data
				TOKEN_PIPE,        // |>
				TOKEN_IDENTIFIER,  // Map
				TOKEN_LPAREN,      // (
				TOKEN_FUNC,        // func
				TOKEN_LPAREN,      // (
				TOKEN_IDENTIFIER,  // item
				TOKEN_IDENTIFIER,  // string
				TOKEN_RPAREN,      // )
				TOKEN_IDENTIFIER,  // string
				TOKEN_NEWLINE,
				TOKEN_INDENT,
				TOKEN_IDENTIFIER,  // trimmed
				TOKEN_WALRUS,      // :=
				TOKEN_IDENTIFIER,  // trim
				TOKEN_LPAREN,      // (
				TOKEN_IDENTIFIER,  // item
				TOKEN_RPAREN,      // )
				TOKEN_NEWLINE,
				TOKEN_RETURN,      // return
				TOKEN_IDENTIFIER,  // trimmed
				TOKEN_NEWLINE,
				TOKEN_DEDENT,
				TOKEN_RPAREN,      // ) closing function call
				TOKEN_EOF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lex := NewLexer(tt.input, "test.kuki")
			tokens, err := lex.ScanTokens()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(tokens) != len(tt.expected) {
				t.Fatalf("Expected %d tokens, got %d", len(tt.expected), len(tokens))
			}

			for i, expected := range tt.expected {
				if tokens[i].Type != expected {
					t.Errorf("Token %d: expected %s, got %s (lexeme: %q)", i, expected, tokens[i].Type, tokens[i].Lexeme)
				}
			}
		})
	}
}
