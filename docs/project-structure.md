# Project Structure

This document describes the organization of the Kukicha codebase.

## Directory Layout

```
kukicha/
├── cmd/
│   └── kukicha/              # CLI entry point
│       └── main.go           # Command-line interface
│
├── internal/                 # Compiler implementation
│   ├── ast/                  # Abstract Syntax Tree definitions
│   │   └── ast.go            # All AST node types
│   │
│   ├── lexer/                # Lexical analysis (tokenization)
│   │   ├── lexer.go          # Main lexer implementation
│   │   ├── token.go          # Token type definitions
│   │   └── lexer_test.go     # Lexer tests
│   │
│   ├── parser/               # Syntax analysis (parsing)
│   │   ├── parser.go         # Recursive descent parser
│   │   └── parser_test.go    # Parser tests
│   │
│   ├── semantic/             # Semantic analysis
│   │   ├── semantic.go       # Type checking & validation
│   │   ├── symbols.go        # Symbol table implementation
│   │   └── semantic_test.go  # Semantic tests
│   │
│   ├── codegen/              # Code generation
│   │   ├── codegen.go        # Go code generator
│   │   └── codegen_test.go   # Code generation tests
│   │
│   └── formatter/            # Code formatting (kukicha fmt)
│       ├── formatter.go      # Main formatter logic
│       ├── printer.go        # Output generation
│       ├── preprocessor.go   # Pre-processing
│       ├── comments.go       # Comment handling
│       └── formatter_test.go # Formatter tests
│
├── stdlib/                   # Kukicha Standard Library
│   ├── iter/                 # Functional iterators
│   │   └── iter.kuki
│   ├── slice/                # Slice operations
│   │   └── slice.kuki
│   ├── string/               # String utilities
│   │   └── string.kuki
│   ├── fetch/                # HTTP client
│   │   └── fetch.kuki
│   ├── files/                # File operations
│   │   └── files.kuki
│   └── parse/                # Data format parsing
│       └── parse.kuki
│
├── docs/                     # Documentation
│   ├── beginner-tutorial.md
│   ├── web-app-tutorial.md
│   ├── kukicha-syntax-v1.0.md
│   ├── kukicha-quick-reference.md
│   ├── kukicha-design-philosophy.md
│   ├── kukicha-compiler-architecture.md
│   ├── kukicha-grammar.ebnf.md
│   └── ...
│
├── examples/                 # Example programs
│   ├── hello.kuki
│   ├── hello_newbie.kuki
│   └── ...
│
├── testdata/                 # Test fixtures
│
├── .claude/                  # Claude Code configuration
│   └── skills/
│       └── kukicha/          # Kukicha language skill
│
├── go.mod                    # Go module definition
├── go.sum                    # Go module checksums
├── LICENSE                   # License file
└── README.md                 # Project overview
```

## Compiler Pipeline

The Kukicha compiler processes source files through four phases:

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Source    │    │   Tokens    │    │     AST     │    │   Go Code   │
│   (.kuki)   │───▶│   Stream    │───▶│    Tree     │───▶│   (.go)     │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                        │                   │                   │
                    Lexer              Parser              CodeGen
                  (lexer/)           (parser/)           (codegen/)
                                          │
                                    Semantic Analysis
                                     (semantic/)
```

### Phase 1: Lexer (`internal/lexer/`)

Converts source code into a token stream.

**Key Features:**
- Indentation-based tokenization (INDENT/DEDENT tokens)
- String interpolation support (`{variable}` syntax)
- 70+ token types

**Files:**
- `lexer.go` - Main tokenization logic
- `token.go` - Token type definitions
- `lexer_test.go` - Test suite

### Phase 2: Parser (`internal/parser/`)

Builds an Abstract Syntax Tree from tokens.

**Key Features:**
- Recursive descent parsing
- Context-sensitive keyword handling
- Implicit petiole calculation

**Files:**
- `parser.go` - Parser implementation
- `parser_test.go` - Test suite

### Phase 3: Semantic Analysis (`internal/semantic/`)

Type checking and validation.

**Key Features:**
- Signature-first type inference
- Symbol table management
- Interface implementation verification

**Files:**
- `semantic.go` - Type checking logic
- `symbols.go` - Symbol table
- `semantic_test.go` - Test suite

### Phase 4: Code Generation (`internal/codegen/`)

Produces idiomatic Go code from the AST.

**Key Features:**
- String interpolation → `fmt.Sprintf`
- Negative indexing → `len()-based` access
- Keyword translation (`and` → `&&`, etc.)
- Special stdlib transpilation (generics for iter)

**Files:**
- `codegen.go` - Go code generation
- `codegen_test.go` - Test suite

## Standard Library (`stdlib/`)

Kukicha's standard library provides pipe-friendly utilities:

| Package | Purpose | Key Functions |
|---------|---------|---------------|
| `iter` | Functional iterators | Filter, Map, Take, Skip, Reduce |
| `slice` | Slice operations | First, Last, Reverse, Unique, Sort |
| `string` | String utilities | ToUpper, Split, Contains, Trim |
| `fetch` | HTTP client (jsonv2) | Get, Post, Json, Text |
| `files` | File operations | Read, Write, List, Exists |
| `parse` | Data parsing (jsonv2) | Json, JsonFromReader, JsonLines, JsonPretty, Csv, Yaml |
| `concurrent` | Concurrency helpers | Parallel, ParallelWithLimit, Go |
| `http` | HTTP server helpers | WithCSRF, Serve |

## The Green Tea Naming Convention

Kukicha uses a botanical metaphor inspired by Japanese green tea:

| Term | Meaning | Go Equivalent |
|------|---------|---------------|
| **Kukicha** | Green tea made from stems | The language itself |
| **Stem** | Project root | Go module |
| **Petiole** | Package (leaf stem) | Go package |
| **Leaf** | Source file | Go source file |

This connects to Go 1.25's **Green Tea GC** (garbage collector) - Kukicha code is optimized for the Green Tea GC's performance improvements.

## CLI Commands

The `kukicha` binary provides these commands:

```bash
kukicha build <file.kuki>     # Compile to Go binary
kukicha run <file.kuki>       # Compile and execute
kukicha check <file.kuki>     # Type-check only
kukicha fmt [options] <path>  # Format source files
kukicha version               # Show version
kukicha help                  # Show help
```

## Testing

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/lexer/... -v
go test ./internal/parser/... -v
go test ./internal/semantic/... -v
go test ./internal/codegen/... -v
go test ./internal/formatter/... -v

# Run with coverage
go test ./... -cover

# Run with race detection
go test ./... -race
```
