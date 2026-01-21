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
			input: "leaf import type interface\n",
			expected: []TokenType{
				TOKEN_LEAF, TOKEN_IMPORT, TOKEN_TYPE, TOKEN_INTERFACE, TOKEN_NEWLINE, TOKEN_EOF,
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
		{
			name:     "single quoted string",
			input:    `'hello'`,
			expected: "hello",
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
	input := `leaf todo

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
	foundLeaf := false
	foundImport := false
	foundType := false
	foundFunc := false

	for _, token := range tokens {
		switch token.Type {
		case TOKEN_LEAF:
			foundLeaf = true
		case TOKEN_IMPORT:
			foundImport = true
		case TOKEN_TYPE:
			foundType = true
		case TOKEN_FUNC:
			foundFunc = true
		}
	}

	if !foundLeaf {
		t.Error("Expected to find LEAF token")
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
		{"leaf", TOKEN_LEAF},
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
