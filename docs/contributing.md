# Contributing to Kukicha

Thank you for your interest in contributing to Kukicha! This document provides guidelines for contributing to the project.

## Getting Started

### Prerequisites

- Go 1.26 or later
- Git
- A text editor or IDE with Go support

### Setting Up Your Development Environment

```bash
# Clone the repository
git clone https://github.com/duber000/kukicha.git
cd kukicha

# Install dependencies
go mod tidy

# Build the compiler
go build -o kukicha ./cmd/kukicha

# Run tests to verify setup
go test ./...
```

## Development Workflow

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

### 2. Make Your Changes

Follow the existing code style and patterns in the codebase.

### 3. Test Your Changes

```bash
# Run all tests
go test ./...

# Run with verbose output
go test ./... -v

# Run with coverage
go test ./... -cover

# Run specific package tests
go test ./internal/lexer/... -v
```

### 4. Commit Your Changes

Write clear, concise commit messages:

```bash
git commit -m "feat: add support for ternary expressions"
git commit -m "fix: correct negative indexing for empty slices"
git commit -m "docs: update syntax reference with new examples"
```

### 5. Submit a Pull Request

Push your branch and create a pull request on GitHub.

## Adding New Features

When adding new language features, follow this process:

### Step 1: Update Documentation

1. Update the specification in `docs/kukicha-syntax-v1.0.md`
2. Update the grammar in `docs/kukicha-grammar.ebnf.md`
3. Add examples to `docs/language-features.md`

### Step 2: Implement in the Compiler

Determine which phase(s) need modification:

| Change Type | Affected Phase(s) |
|------------|-------------------|
| New keyword | Lexer, Parser |
| New syntax | Parser, possibly Lexer |
| New operator | Lexer, Parser, CodeGen |
| Type system change | Semantic, possibly Parser |
| New transpilation pattern | CodeGen |

### Step 3: Add Tests

Add comprehensive tests in the appropriate `*_test.go` file:

```go
func TestYourNewFeature(t *testing.T) {
    input := `your kukicha code here`

    // Test lexer if applicable
    // Test parser if applicable
    // Test semantic analysis if applicable
    // Test code generation if applicable
}
```

### Step 4: Update Examples

Add example code to `examples/` if the feature is significant.

## Code Style

### Go Code

- Follow standard Go conventions (`gofmt`)
- Use descriptive variable and function names
- Add comments for non-obvious logic
- Keep functions focused and reasonably sized

### Kukicha Code (Examples/Tests)

- Use 4-space indentation
- Follow the patterns in existing examples
- Use English keywords (`and`, `or`, `not`, `equals`)

## Testing Guidelines

### Unit Tests

Each compiler phase should have unit tests:

```go
func TestFeatureName(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string // or appropriate type
    }{
        {"basic case", "input", "expected"},
        {"edge case", "input2", "expected2"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Integration Tests

For end-to-end testing, ensure the full pipeline works:

1. Kukicha source → Lexer → Tokens
2. Tokens → Parser → AST
3. AST → Semantic → Validated AST
4. AST → CodeGen → Go code
5. Go code → `go build` → Binary

## Reporting Issues

When reporting issues, please include:

1. **Description**: Clear description of the problem
2. **Reproduction**: Steps to reproduce the issue
3. **Expected Behavior**: What you expected to happen
4. **Actual Behavior**: What actually happened
5. **Environment**: Go version, OS, Kukicha version
6. **Code Sample**: Minimal Kukicha code that demonstrates the issue

## Pull Request Guidelines

- Keep PRs focused on a single feature or fix
- Include tests for new functionality
- Update documentation as needed
- Ensure all tests pass
- Request review from maintainers

## Project Areas

### Core Compiler (`internal/`)

The compiler implementation. Changes here require careful testing.

### Standard Library (`stdlib/`)

Kukicha's built-in packages. New packages or functions welcome!

### Documentation (`docs/`)

Always appreciated! Improvements to tutorials, references, and examples help everyone.

### Examples (`examples/`)

Real-world examples showing Kukicha in action.

### CLI (`cmd/kukicha/`)

Command-line interface improvements.

## Questions?

If you have questions about contributing:

1. Check existing documentation
2. Look at similar features in the codebase
3. Open an issue for discussion

## License

By contributing to Kukicha, you agree that your contributions will be licensed under the same license as the project.

---

Thank you for contributing to Kukicha!
