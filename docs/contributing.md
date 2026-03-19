# Contributing to Kukicha

Thank you for your interest in contributing to Kukicha! This document provides guidelines for contributing to the project.

## Getting Started

### Prerequisites

- Go 1.26.1 or later
- Git
- [golangci-lint](https://golangci-lint.run/) v2 (`go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`)
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

### 3. Test and Lint Your Changes

```bash
# Run all tests
make test

# Run linter
make lint

# Run with verbose output
go test ./... -v

# Run specific package tests
go test ./internal/lexer/... -v
```

The linter (`golangci-lint`) catches unused code, unchecked errors, and common mistakes that tests alone don't detect. All lint issues must be resolved before merging.

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
3. Update the relevant `AGENTS.md` files (root, `stdlib/`, `internal/`) with any new syntax or API changes.

### Step 2: Implement in the Compiler

Determine which phase(s) need modification:

| Change Type | Affected Phase(s) |
|------------|-------------------|
| New keyword | Lexer, Parser |
| New syntax | Parser, possibly Lexer |
| New operator | Lexer, Parser, CodeGen |
| Type system change | Semantic, possibly Parser |
| New transpilation pattern | CodeGen |
| New `# kuki:` directive | Lexer (already handles), Parser (already collects), Semantic |

### Step 3: Check for Vulnerabilities

After adding or updating stdlib dependencies, run a vulnerability audit:

```bash
kukicha audit             # check all dependencies for known CVEs
kukicha audit --warn-only # same but exit 0 (useful in CI)
```

Since Kukicha transpiles to Go, the stdlib's Go dependencies (declared in `cmd/kukicha/stdlib.go`) are part of the user's dependency graph. When a user runs `kukicha audit` in their project, govulncheck follows the `replace` directive into `.kukicha/stdlib/` and checks stdlib dependencies transitively. Keep them up to date.

To check the compiler's own dependencies, run `kukicha audit` (or `go run ./cmd/kukicha audit`) in the kukicha repo root.

### Step 4: Add Tests

For **compiler internals** (`internal/`), add tests in the appropriate Go `*_test.go` file:

```go
func TestYourNewFeature(t *testing.T) {
    input := `your kukicha code here`

    // Test lexer if applicable
    // Test parser if applicable
    // Test semantic analysis if applicable
    // Test code generation if applicable
}
```

For **stdlib packages**, write tests in `*_test.kuki` using the table-driven pattern (see "Writing stdlib tests" under "Modifying the Standard Library").

### Step 5: Update Examples

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
- Ensure all tests pass (`make test`)
- Ensure lint is clean (`make lint`)
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
make generate   # transpile all stdlib/*.kuki → *.go (including *_test.kuki → *_test.go), rebuild compiler
make build      # re-embed the updated .go files into the kukicha binary
```

`make generate` already rebuilds the compiler internally (it needs a working binary to transpile), but that intermediate binary doesn't yet contain the newly generated `.go` files. The final `make build` is what bakes them in.

`make generate` also calls `generate-tests` to regenerate `*_test.go` files from `*_test.kuki` sources. You can regenerate only the test files without touching the main stdlib:

```bash
make generate-tests   # transpile stdlib/*_test.kuki → *_test.go only
```

### Staleness check

`make test` checks that every `.go` file is up to date with its `.kuki` source before running the test suite. This covers both test files (`*_test.kuki`) and main library files. If any generated file is missing or has an older timestamp than its source, the build fails immediately:

```
STALE: stdlib/files/files_test.go is older than stdlib/files/files_test.kuki (run 'make generate')
Run 'make generate' to regenerate test files.
```

To check staleness without running tests:

```bash
make check-test-staleness    # test files only
make check-main-staleness    # non-test library files only
```

### Registry files (`stdlib_registry_gen.go`, `go_stdlib_gen.go`)

Both registry files are regenerated automatically by `make build` via `go generate ./...`. You do **not** need to run `make genstdlibregistry` or `make gengostdlib` manually.

**`stdlib_registry_gen.go`** contains two maps generated by `cmd/genstdlibregistry`:
- `generatedStdlibRegistry` — every exported stdlib function → return-value count (used to decompose pipe chains and `onerr` clauses)
- `generatedStdlibDeprecated` — functions with `# kuki:deprecated` directives → deprecation message (used to warn callers at compile time)

**`go_stdlib_gen.go`** contains return counts and per-position type info for Go stdlib functions (e.g., `os.ReadFile` → 2 returns: `[]byte`, `error`). The generator (`cmd/gengostdlib/main.go`) uses `go/importer` to extract signatures from Go's compiled export data. The curated list covers ~100 functions across `os`, `strconv`, `fmt`, `net`, `net/url`, `io`, `bufio`, `strings`, `bytes`, `time`, `sync`, and more. Instance methods (e.g., `bufio.Scanner.Scan`) remain hand-coded in `semantic_calls.go` since `go/importer` extracts package-level functions only.

You only need `make genstdlibregistry` or `make gengostdlib` as standalone commands if you want to regenerate a registry file without doing a full build.

### Adding a new stdlib package

1. Create `stdlib/<pkg>/<pkg>.kuki` with a `petiole <pkg>` declaration
2. Create `stdlib/<pkg>/<pkg>_test.kuki` with table-driven tests (see below)
3. Run `make generate && make build`
4. Run `kukicha check stdlib/<pkg>/<pkg>.kuki` to validate
5. Add the package to `stdlib/AGENTS.md` and `stdlib/CLAUDE.md` so AI agents know it exists

### Writing stdlib tests

Every stdlib package needs a `*_test.kuki` file. Use the **table-driven pattern** — it makes failures self-describing (`TestClamp/below_min` instead of a bare error) and keeps the test body minimal:

```kukicha
petiole mypackage_test

import "stdlib/mypackage"
import "stdlib/test"
import "testing"

# --- TestFoo ---
type FooCase
    name  string
    input string
    want  string

func TestFoo(t reference testing.T)
    cases := list of FooCase{
        FooCase{name: "basic",      input: "hello", want: "HELLO"},
        FooCase{name: "empty",      input: "",      want: ""},
        FooCase{name: "mixed case", input: "Hello", want: "HELLO"},
    }
    for tc in cases
        t.Run(tc.name, (t reference testing.T) =>
            got := mypackage.Foo(tc.input)
            test.AssertEqual(t, got, tc.want)
        )
```

**Conventions:**
- Case types at file scope, named `<FunctionName>Case`; `name string` is the first field
- `t.Run(tc.name, (t reference testing.T) => ...)` wraps every assertion body
- Prefer `test.AssertEqual` / `test.AssertNoError` / `test.AssertError` over bare `t.Errorf`
- A comment `# --- TestFoo ---` separates each function's table
- Import `stdlib/test` only in `*_test.kuki` files, never in library code

After writing the test file, regenerate and verify:

```bash
kukicha check stdlib/<pkg>/<pkg>_test.kuki
./kukicha build stdlib/<pkg>/<pkg>_test.kuki   # generates _test.go
go test ./stdlib/<pkg>/...
```

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
3. Run `make generate && make build`, then **force-regenerate** all stdlib files to update version headers. `make generate` uses `--if-changed` and skips `.kuki` files whose source hasn't changed, so a version-only bump will leave stale headers. After the initial generate, re-transpile without `--if-changed` and rebuild:
   ```bash
   make generate && make build
   # Force-regenerate all stdlib .go files (updates version headers):
   for f in stdlib/*/*.kuki; do
       [[ "$f" == *_test.kuki ]] && continue
       [[ "$f" == stdlib/test/test.kuki ]] && continue
       ./kukicha build --skip-build "$f"
   done
   for f in stdlib/*/*_test.kuki; do ./kukicha build --skip-build "$f"; done
   make build   # re-embed the updated .go files
   ```
4. Run `make test && make lint` to confirm everything passes before tagging.
5. Commit the regenerated `.go`/`*_test.go` files and doc/version updates in a single commit. (The `.kuki` sources are inputs, not outputs — only stage them if you changed them.)
6. Tag and push (push the tag explicitly — `--follow-tags` can silently skip tags):

```bash
git tag v0.0.X
git push origin main
git push origin v0.0.X
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
