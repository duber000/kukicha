package lexer

import "fmt"

// TokenType represents the type of a token
type TokenType int

const (
	// Literals
	TOKEN_IDENTIFIER TokenType = iota
	TOKEN_INTEGER
	TOKEN_FLOAT
	TOKEN_STRING
	TOKEN_TRUE
	TOKEN_FALSE

	// Keywords
	TOKEN_LEAF
	TOKEN_IMPORT
	TOKEN_TYPE
	TOKEN_INTERFACE
	TOKEN_FUNC
	TOKEN_RETURN
	TOKEN_IF
	TOKEN_ELSE
	TOKEN_FOR
	TOKEN_IN
	TOKEN_FROM
	TOKEN_TO
	TOKEN_THROUGH
	TOKEN_GO
	TOKEN_DEFER
	TOKEN_MAKE
	TOKEN_LIST
	TOKEN_MAP
	TOKEN_CHANNEL
	TOKEN_SEND
	TOKEN_RECEIVE
	TOKEN_CLOSE
	TOKEN_PANIC
	TOKEN_RECOVER
	TOKEN_ERROR
	TOKEN_EMPTY
	TOKEN_REFERENCE
	TOKEN_ON
	TOKEN_THIS
	TOKEN_DISCARD
	TOKEN_AT
	TOKEN_OF
	TOKEN_AS

	// Variadic keyword
	TOKEN_MANY

	// Function type keyword
	TOKEN_FUNCTION

	// Operators
	TOKEN_WALRUS        // :=
	TOKEN_ASSIGN        // =
	TOKEN_EQUALS        // equals
	TOKEN_DOUBLE_EQUALS // ==
	TOKEN_NOT_EQUALS    // !=
	TOKEN_LT            // <
	TOKEN_GT            // >
	TOKEN_LTE           // <=
	TOKEN_GTE           // >=
	TOKEN_PLUS          // +
	TOKEN_MINUS         // -
	TOKEN_STAR          // *
	TOKEN_SLASH         // /
	TOKEN_PERCENT       // %
	TOKEN_AND           // and
	TOKEN_AND_AND       // &&
	TOKEN_OR            // or
	TOKEN_OR_OR         // ||
	TOKEN_ONERR         // onerr
	TOKEN_NOT           // not
	TOKEN_BANG          // !
	TOKEN_PIPE          // |>
	TOKEN_ARROW_LEFT    // <-

	// Delimiters
	TOKEN_LPAREN   // (
	TOKEN_RPAREN   // )
	TOKEN_LBRACKET // [
	TOKEN_RBRACKET // ]
	TOKEN_LBRACE   // {
	TOKEN_RBRACE   // }
	TOKEN_COMMA    // ,
	TOKEN_DOT      // .
	TOKEN_COLON    // :

	// Special
	TOKEN_NEWLINE
	TOKEN_INDENT
	TOKEN_DEDENT
	TOKEN_EOF
	TOKEN_COMMENT   // # comment (standalone on its own line)
	TOKEN_SEMICOLON // ; (for Go-style syntax support)
)

// Token represents a single token in the source code
type Token struct {
	Type   TokenType
	Lexeme string
	Line   int
	Column int
	File   string
}

// String returns a string representation of the token type
func (t TokenType) String() string {
	switch t {
	case TOKEN_IDENTIFIER:
		return "IDENTIFIER"
	case TOKEN_INTEGER:
		return "INTEGER"
	case TOKEN_FLOAT:
		return "FLOAT"
	case TOKEN_STRING:
		return "STRING"
	case TOKEN_TRUE:
		return "TRUE"
	case TOKEN_FALSE:
		return "FALSE"

	// Keywords
	case TOKEN_LEAF:
		return "LEAF"
	case TOKEN_IMPORT:
		return "IMPORT"
	case TOKEN_TYPE:
		return "TYPE"
	case TOKEN_INTERFACE:
		return "INTERFACE"
	case TOKEN_FUNC:
		return "FUNC"
	case TOKEN_RETURN:
		return "RETURN"
	case TOKEN_IF:
		return "IF"
	case TOKEN_ELSE:
		return "ELSE"
	case TOKEN_FOR:
		return "FOR"
	case TOKEN_IN:
		return "IN"
	case TOKEN_FROM:
		return "FROM"
	case TOKEN_TO:
		return "TO"
	case TOKEN_THROUGH:
		return "THROUGH"
	case TOKEN_GO:
		return "GO"
	case TOKEN_DEFER:
		return "DEFER"
	case TOKEN_MAKE:
		return "MAKE"
	case TOKEN_LIST:
		return "LIST"
	case TOKEN_MAP:
		return "MAP"
	case TOKEN_CHANNEL:
		return "CHANNEL"
	case TOKEN_SEND:
		return "SEND"
	case TOKEN_RECEIVE:
		return "RECEIVE"
	case TOKEN_CLOSE:
		return "CLOSE"
	case TOKEN_PANIC:
		return "PANIC"
	case TOKEN_RECOVER:
		return "RECOVER"
	case TOKEN_ERROR:
		return "ERROR"
	case TOKEN_EMPTY:
		return "EMPTY"
	case TOKEN_REFERENCE:
		return "REFERENCE"
	case TOKEN_ON:
		return "ON"
	case TOKEN_THIS:
		return "THIS"
	case TOKEN_DISCARD:
		return "DISCARD"
	case TOKEN_AT:
		return "AT"
	case TOKEN_OF:
		return "OF"
	case TOKEN_AS:
		return "AS"

	// Variadic keyword
	case TOKEN_MANY:
		return "MANY"

	// Function type keyword
	case TOKEN_FUNCTION:
		return "FUNCTION"

	// Operators
	case TOKEN_WALRUS:
		return "WALRUS"
	case TOKEN_ASSIGN:
		return "ASSIGN"
	case TOKEN_EQUALS:
		return "EQUALS"
	case TOKEN_DOUBLE_EQUALS:
		return "DOUBLE_EQUALS"
	case TOKEN_NOT_EQUALS:
		return "NOT_EQUALS"
	case TOKEN_LT:
		return "LT"
	case TOKEN_GT:
		return "GT"
	case TOKEN_LTE:
		return "LTE"
	case TOKEN_GTE:
		return "GTE"
	case TOKEN_PLUS:
		return "PLUS"
	case TOKEN_MINUS:
		return "MINUS"
	case TOKEN_STAR:
		return "STAR"
	case TOKEN_SLASH:
		return "SLASH"
	case TOKEN_PERCENT:
		return "PERCENT"
	case TOKEN_AND:
		return "AND"
	case TOKEN_AND_AND:
		return "AND_AND"
	case TOKEN_OR:
		return "OR"
	case TOKEN_OR_OR:
		return "OR_OR"
	case TOKEN_ONERR:
		return "ONERR"
	case TOKEN_NOT:
		return "NOT"
	case TOKEN_BANG:
		return "BANG"
	case TOKEN_PIPE:
		return "PIPE"
	case TOKEN_ARROW_LEFT:
		return "ARROW_LEFT"

	// Delimiters
	case TOKEN_LPAREN:
		return "LPAREN"
	case TOKEN_RPAREN:
		return "RPAREN"
	case TOKEN_LBRACKET:
		return "LBRACKET"
	case TOKEN_RBRACKET:
		return "RBRACKET"
	case TOKEN_LBRACE:
		return "LBRACE"
	case TOKEN_RBRACE:
		return "RBRACE"
	case TOKEN_COMMA:
		return "COMMA"
	case TOKEN_DOT:
		return "DOT"
	case TOKEN_COLON:
		return "COLON"

	// Special
	case TOKEN_NEWLINE:
		return "NEWLINE"
	case TOKEN_INDENT:
		return "INDENT"
	case TOKEN_DEDENT:
		return "DEDENT"
	case TOKEN_EOF:
		return "EOF"
	case TOKEN_COMMENT:
		return "COMMENT"
	case TOKEN_SEMICOLON:
		return "SEMICOLON"
	default:
		return "UNKNOWN"
	}
}

// String returns a string representation of the token
func (t Token) String() string {
	return fmt.Sprintf("Token{%s, %q, %d:%d}", t.Type, t.Lexeme, t.Line, t.Column)
}

// keywords maps keyword strings to their token types
var keywords = map[string]TokenType{
	"leaf":      TOKEN_LEAF,
	"import":    TOKEN_IMPORT,
	"type":      TOKEN_TYPE,
	"interface": TOKEN_INTERFACE,
	"func":      TOKEN_FUNC,
	"return":    TOKEN_RETURN,
	"if":        TOKEN_IF,
	"else":      TOKEN_ELSE,
	"for":       TOKEN_FOR,
	"in":        TOKEN_IN,
	"from":      TOKEN_FROM,
	"to":        TOKEN_TO,
	"through":   TOKEN_THROUGH,
	"go":        TOKEN_GO,
	"defer":     TOKEN_DEFER,
	"make":      TOKEN_MAKE,
	"list":      TOKEN_LIST,
	"map":       TOKEN_MAP,
	"channel":   TOKEN_CHANNEL,
	"send":      TOKEN_SEND,
	"receive":   TOKEN_RECEIVE,
	"close":     TOKEN_CLOSE,
	"panic":     TOKEN_PANIC,
	"recover":   TOKEN_RECOVER,
	"error":     TOKEN_ERROR,
	"empty":     TOKEN_EMPTY,
	"nil":       TOKEN_EMPTY, // nil is an alias for empty
	"reference": TOKEN_REFERENCE,
	"on":        TOKEN_ON,
	"this":      TOKEN_THIS,
	"discard":   TOKEN_DISCARD,
	"at":        TOKEN_AT,
	"of":        TOKEN_OF,
	"as":        TOKEN_AS,
	"many":      TOKEN_MANY,
	"function":  TOKEN_FUNCTION,
	"true":      TOKEN_TRUE,
	"false":     TOKEN_FALSE,
	"equals":    TOKEN_EQUALS,
	"and":       TOKEN_AND,
	"or":        TOKEN_OR,
	"onerr":     TOKEN_ONERR,
	"not":       TOKEN_NOT,
}

// LookupKeyword returns the token type for a keyword, or TOKEN_IDENTIFIER if not a keyword
func LookupKeyword(identifier string) TokenType {
	if tok, ok := keywords[identifier]; ok {
		return tok
	}
	return TOKEN_IDENTIFIER
}
