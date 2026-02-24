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

# Build the compiler
make build

# Run tests to verify setup
make test

# Install pre-commit hooks
make install-hooks
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
make test

# Run with verbose output
go test ./... -v

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

1. Update the grammar in `docs/kukicha-grammar.ebnf.md`
2. Update  `docs/kukicha-quick-reference.md`
3. Update `AGENTS.md` we are in their world now.

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

## Modifying the Standard Library

The stdlib is written in Kukicha (`.kuki` files) and transpiled to Go. The generated `.go` files are embedded into the `kukicha` binary at build time via `//go:embed stdlib/*/*.go`. **Never edit `stdlib/*/*.go` directly** — always edit the `.kuki` source and regenerate.

### Build sequence

```bash
make generate   # transpile all stdlib/*.kuki → *.go, rebuild compiler
make build      # re-embed the updated .go files into the kukicha binary
```

`make generate` already rebuilds the compiler internally (it needs a working binary to transpile), but that intermediate binary doesn't yet contain the newly generated `.go` files. The final `make build` is what bakes them in.

### When to run `make genstdlibregistry`

`make generate` calls `genstdlibregistry` automatically as its first step, so you rarely need to run it standalone. It regenerates `internal/semantic/stdlib_registry_gen.go`, which is a map of every exported stdlib function to its return-value count. The compiler's semantic analyzer uses this to correctly decompose pipe chains and `onerr` clauses.

You need it (via `make generate`) when:
- Adding a new stdlib package
- Adding, removing, or changing the return signature of an exported stdlib function

You do **not** need it when:
- Editing the body of an existing function without changing its signature

### Adding a new stdlib package

1. Create `stdlib/<pkg>/<pkg>.kuki` with a `petiole <pkg>` declaration
2. Run `make generate && make build`
3. Run `kukicha check stdlib/<pkg>/<pkg>.kuki` to validate
4. Add the package to `stdlib/AGENTS.md` so AI agents know it exists

### Documentation (`docs/`)

Always appreciated! Improvements to tutorials, references, and examples help everyone.

### Examples (`examples/`)

Real-world examples showing Kukicha in action.

### Editor Extensions (`editors/`)

Syntax highlighting, tree-sitter grammars, and LSP integration for editors.

### CLI (`cmd/kukicha/`)

Command-line interface improvements.

## Git Hooks

Run `make install-hooks` to install the pre-commit hook. This links `scripts/pre-commit` into `.git/hooks/` and runs automatically on every commit to catch common issues before they reach CI.

## Zed Extension

The Zed editor extension lives in `editors/zed/` and includes:

- **Tree-sitter grammar** (`editors/zed/grammars/kukicha/`) — parsing for syntax highlighting
- **Highlight queries** (`editors/zed/languages/kukicha/highlights.scm`) — the source of truth for highlighting rules
- **LSP integration** (`editors/zed/src/lib.rs`) — connects Zed to `kukicha-lsp`

### Testing

```bash
make zed-test
```

This runs three checks:
1. `cargo check` — verifies the Rust extension compiles
2. `check-highlights.sh` — verifies highlight queries are in sync between `languages/` and `grammars/`
3. `npm test` (in `grammars/kukicha/`) — runs tree-sitter grammar tests

### Editing highlights

Edit `editors/zed/languages/kukicha/highlights.scm` (the source of truth), then sync:

```bash
editors/zed/scripts/sync-highlights.sh
```

Never edit `editors/zed/grammars/kukicha/queries/highlights.scm` directly — it gets overwritten by the sync script.

### Adding tree-sitter tests

Add test cases to `editors/zed/grammars/kukicha/test/corpus/`. Each test file uses the tree-sitter test format: a name header, source code, a `---` separator, and the expected S-expression parse tree.

## Releasing a New Version

Follow these steps in order. Skipping step 3 is how the stdlib `.go` files end up out of date with the tagged release.

1. Bump the version constant in `internal/version/version.go`.
2. Update the version references in `README.md` (the `go install` snippet and the **Status** section at the bottom).
3. Run `make generate && make build` to regenerate all stdlib `.go` files with the new version header and rebuild the compiler with the updated files embedded.
4. Commit everything — source `.kuki` files, regenerated `.go` files, and doc/version updates — in a single commit.
5. Tag and push:

```bash
git tag v0.0.X
git push && git push --tags
```

## Questions?

If you have questions about contributing:

1. Check existing documentation
2. Look at similar features in the codebase
3. Open an issue for discussion

## License

By contributing to Kukicha, you agree that your contributions will be licensed under the same license as the project.

---

Thank you for contributing to Kukicha!
