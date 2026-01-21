# Kukicha Programming Language

Kukicha (茎 = stem in Japanese) is a high-level, beginner-friendly programming language that compiles to idiomatic Go code. It combines English-like syntax with Go's performance and explicit type system.

## Project Status

**Current Version:** 1.0.0 🎉- Ready for Testing

## Quick Start

### Prerequisites

- Go 1.24+ (required)
- Go 1.25+ recommended for Green Tea GC performance improvements (10-40% faster GC)

### Installation

```bash
git clone https://github.com/duber000/kukicha.git
cd kukicha
go mod tidy
go build -o kukicha ./cmd/kukicha
```

This builds the `kukicha` binary in your current directory. You can move it to a location in your `$PATH`:

```bash
# Optional: install to ~/go/bin (or any directory in your PATH)
go install ./cmd/kukicha
```

### Enabling Green Tea GC (Go 1.25+)

For optimal performance, build with the Green Tea garbage collector:

```bash
GOEXPERIMENT=greenteagc go build -o kukicha ./cmd/kukicha
```

Or run tests with Green Tea GC:

```bash
GOEXPERIMENT=greenteagc go test ./...
```

Note: Green Tea GC will be enabled by default in Go 1.26+ (expected February 2026)

### Usage

Once built, use the `kukicha` CLI to work with `.kuki` files:

```bash
# Transpile and build a Kukicha file to a Go binary
./kukicha build hello.kuki

# Transpile and run immediately
./kukicha run hello.kuki

# Type-check without compiling
./kukicha check hello.kuki

# Format Kukicha source files
./kukicha fmt hello.kuki          # Output to stdout
./kukicha fmt -w hello.kuki       # Write formatted output back to file
./kukicha fmt --check hello.kuki  # Check if file is formatted (exit 1 if not)
./kukicha fmt ./src/              # Format all .kuki files in directory

# Show version
./kukicha version
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run lexer tests with verbose output
go test ./internal/lexer/... -v
```

## Language Features

### Core Design Decisions (v1.0.0)

Kukicha v1.0.0 introduces key refinements that balance simplicity, performance, and consistency:

1. **📦 Optional Leaf Declarations** - Folder-based package model with automatic Stem (package) calculation from file path. No more header/directory sync issues!

2. **🎯 Signature-First Type Inference** - Explicit types required for function parameters and returns; inference only for local variables. Maintains Go's performance while reducing boilerplate.

3. **⚡ Literal vs Dynamic Indexing** - Negative indices with literal constants (e.g., `items[-1]`) compile to zero-overhead code. Dynamic indices require explicit `.at()` method.

4. **📏 Indentation as Canonical** - The `kuki fmt` tool converts all code to standard 4-space indentation format, preventing "dialect drift" between coding styles.

5. **🔧 Context-Sensitive Type Keywords** - `list`, `map`, and `channel` are context-sensitive. In type contexts (parameters, fields), they start composite types. No lookahead needed at lexer level.

6. **📝 Dual Method Syntax** - Readable Kukicha style (`func Display on this Todo` with explicit receiver) and Go-compatible style (`func (t Todo) Display()`) both supported.

7. **🔄 Empty Literal Lookahead** - `empty` uses 1-token lookahead to determine if it's standalone (`nil`) or typed (`empty list of Todo`).

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

# Method with explicit 'this' - readable Kukicha style
func Display on this Todo string
    return "{this.id}: {this.title}"

# Go-style also works (for copy-paste from Go tutorials)
func (t Todo) Summary() string
    return t.title
```

#### Error Handling (OnErr Operator)

```kukicha
# Auto-unwrap (T, error) tuples
content := file.read("config.json") onerr panic "missing file"

# Provide default value
port := env.get("PORT") onerr "8080"
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
│   └── kukicha/           # ✅ CLI entry point
├── internal/
│   ├── lexer/             # ✅ Lexer implementation
│   │   ├── lexer.go
│   │   ├── token.go
│   │   └── lexer_test.go
│   ├── parser/            # ✅ Parser implementation
│   ├── semantic/          # ✅ Semantic analysis
│   ├── codegen/           # ✅ Code generation
│   ├── formatter/         # ✅ Code formatter (kukicha fmt)
│   └── ast/               # ✅ AST definitions
├── docs/                  # Language documentation
├── examples/              # Example programs
├── testdata/              # Test fixtures
├── go.mod
└── README.md
```

## Transpiler Implementation

The Kukicha transpiler converts `.kuki` source files into idiomatic Go code through four phases:

1. **Lexer** - Tokenizes source with indentation support
2. **Parser** - Builds Abstract Syntax Tree (AST)
3. **Semantic Analysis** - Type checking and validation
4. **Code Generation** - Produces idiomatic Go code

### Example

```kukicha
func Greet(name string) string
    return "Hello {name}"
```

Transpiles to:

```go
func Greet(name string) string {
    return fmt.Sprintf("Hello %s", name)
}
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/lexer/... -v
go test ./internal/parser/... -v
go test ./internal/semantic/... -v
go test ./internal/codegen/... -v
```

## Documentation

- [Language Syntax Reference](docs/kukicha-syntax-v1.0.md) - Complete syntax guide
- [Quick Reference](docs/kukicha-quick-reference.md) - Developer cheat sheet
- [Compiler Architecture](docs/kukicha-compiler-architecture.md) - Implementation details
- [Grammar (EBNF)](docs/kukicha-grammar.ebnf.md) - Formal grammar definition
- [Standard Library Roadmap](docs/kukicha-stdlib-roadmap.md) - Future library features

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

## Future Enhancements

Key word changes:

1. Not really happy with the this keyword
2. onerr doesn't roll off the tongue.
3. Generics placeholder keywords are clunky

See [Standard Library Roadmap](docs/kukicha-stdlib-roadmap.md) for planned features:

1. **Standard Library** - HTTP, JSON, File I/O, Docker, K8s, LLM packages
2. **Package Manager** - Dependency management and versioning
3. **IDE Support** - VS Code extension with syntax highlighting and IntelliSense
4. **Debugger** - Source-level debugging support
5. **Formatter Enhancements** - Style options (`--style=compact`, `--style=expanded`), auto-fix

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

**Status**: Ready for testing!
**Version**: 1.0.0
**Go Compatibility**: 1.24+ (1.25+ recommended for Green Tea GC optimization)
