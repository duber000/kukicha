# Kukicha Programming Language

Kukicha (茎 = stem in Japanese) is a high-level, beginner-friendly programming language that compiles to idiomatic Go code. It combines English-like syntax with Go's performance and explicit type system.

## Project Status

**Current Phase:** Lexer Implementation ✅

### Completed

- ✅ Language specification (v1.0)
- ✅ Formal grammar (EBNF)
- ✅ Compiler architecture design
- ✅ **Lexer implementation**
  - Token types for all language constructs
  - Indentation-based syntax (4 spaces, tabs rejected)
  - String interpolation support
  - All operators and keywords
  - Comprehensive test suite

### In Progress

- Parser implementation (next phase)
- Semantic analysis
- Code generation

## Quick Start

### Prerequisites

- Go 1.25+ (for Green Tea GC support)

### Installation

```bash
git clone https://github.com/duber000/kukicha.git
cd kukicha
go mod tidy
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run lexer tests with verbose output
go test ./internal/lexer/... -v
```

## Language Features

### Core Design Decisions (v1.1.0)

Kukicha v1.1.0 introduces four key refinements that balance simplicity, performance, and consistency:

1. **📦 Optional Leaf Declarations** - Folder-based package model with automatic Stem (package) calculation from file path. No more header/directory sync issues!

2. **🎯 Signature-First Type Inference** - Explicit types required for function parameters and returns; inference only for local variables. Maintains Go's performance while reducing boilerplate.

3. **⚡ Literal vs Dynamic Indexing** - Negative indices with literal constants (e.g., `items[-1]`) compile to zero-overhead code. Dynamic indices require explicit `.at()` method.

4. **📏 Indentation as Canonical** - The `kuki fmt` tool converts all code to standard 4-space indentation format, preventing "dialect drift" between coding styles.

### Philosophy

Kukicha smooths Go's rough edges while preserving its power:

- ✅ **Keep**: Explicit types, static typing, performance, Go's stdlib
- ✅ **Smooth**: Symbols minimized, English-like keywords, consistent syntax
- ✅ **Star**: The walrus operator `:=` for clean variable binding
- ✅ **Simple**: Three-level module hierarchy (Twig → Stem → Leaf)

### Key Syntax Highlights

#### Variables (Walrus Operator ⭐)

```kukicha
# Create new binding
count := 42

# Reassign existing
count = 100
```

#### Functions & Methods

```kukicha
# Function with explicit types (required)
func Greet(name string) string
    return "Hello {name}"

# Method with implicit receiver
func Display on Todo () string
    return "{this.id}: {this.title}"
```

#### Error Handling (Or Operator)

```kukicha
# Auto-unwrap (T, error) tuples
content := file.read("config.json") or panic "missing file"

# Provide default value
port := env.get("PORT") or "8080"
```

#### Pipe Operator

```kukicha
# Clean data pipelines
result := data
    |> parse()
    |> transform()
    |> process()
```

#### Concurrency

```kukicha
# Goroutines
go fetchData(url)

# Channels
ch := make channel of string
send ch, "message"
msg := receive ch
```

#### Collections with Membership Testing

```kukicha
# Lists
items := list of string{"a", "b", "c"}
last := items[-1]  # Negative indexing

# Membership testing
if user in admins
    grantAccess()
```

## Project Structure

```
kukicha/
├── cmd/
│   └── kukicha/           # CLI entry point (coming soon)
├── internal/
│   ├── lexer/             # ✅ Lexer implementation
│   │   ├── lexer.go
│   │   ├── token.go
│   │   └── lexer_test.go
│   ├── parser/            # 🔄 Next: Parser
│   ├── semantic/          # ⏳ Semantic analysis
│   ├── codegen/           # ⏳ Code generation
│   └── compiler/          # ⏳ Compiler orchestration
├── docs/                  # Language documentation
├── examples/              # Example programs
├── testdata/              # Test fixtures
├── go.mod
└── README.md
```

## Lexer Implementation

The lexer (tokenizer) converts Kukicha source code into a stream of tokens.

### Features

- **Indentation-based syntax**: 4 spaces per level (tabs rejected)
- **String interpolation**: `"Hello {name}"`
- **All operators**: `|>`, `or`, `:=`, `==`, `in`, etc.
- **Keywords**: 35+ keywords including `leaf`, `func`, `type`, `interface`
- **Error reporting**: Clear error messages with line/column information

### Example

```kukicha
func Greet(name string)
    print "Hello {name}"
```

Tokenizes to:
```
FUNC, IDENTIFIER(Greet), LPAREN, IDENTIFIER(name), IDENTIFIER(string), RPAREN, NEWLINE
INDENT, IDENTIFIER(print), STRING("Hello {name}"), NEWLINE
DEDENT, EOF
```

### Running Lexer Tests

```bash
# All lexer tests
go test ./internal/lexer/... -v

# Specific test
go test ./internal/lexer/... -run TestIndentation -v
```

## Documentation

- [Language Syntax Reference](kukicha-syntax-v1.0.md) - Complete syntax guide
- [Compiler Architecture](kukicha-compiler-architecture.md) - Implementation details
- [Grammar (EBNF)](kukicha-grammar.ebnf.md) - Formal grammar definition
- [Quick Reference](kukicha-quick-reference.md) - Developer cheat sheet

## Development

### Adding New Features

1. Update the specification in `docs/`
2. Update the grammar in `kukicha-grammar.ebnf.md`
3. Implement in the appropriate phase (lexer/parser/semantic/codegen)
4. Add comprehensive tests
5. Update documentation

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run specific package
go test ./internal/lexer/...
```

## Next Steps

1. **Parser** - Build Abstract Syntax Tree from tokens
2. **Semantic Analysis** - Type checking and validation
3. **Code Generation** - Transform AST to idiomatic Go
4. **CLI Tool** - `kukicha build`, `kukicha run`, etc.
5. **Standard Library** - HTTP, JSON, File I/O, Docker, K8s, LLM packages

## Contributing

Contributions are welcome! Please:

1. Follow the existing code style
2. Add tests for new features
3. Update documentation
4. Ensure all tests pass

## License

See [LICENSE](LICENSE) file for details.

## Acknowledgments

Kukicha is designed for programming beginners while maintaining compatibility with Go's ecosystem and performance characteristics.

---

**Status**: Active development
**Version**: 0.1.0 (Lexer complete)
**Target Go Version**: 1.25+ with Green Tea GC
