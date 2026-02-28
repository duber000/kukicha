# internal/CLAUDE.md

Compiler internals reference. Read this when working in `internal/`. For language syntax and build commands see the root `CLAUDE.md`.

## Pipeline Overview

```
source (.kuki)
  → lexer/     — runes → []Token (INDENT/DEDENT injected)
  → parser/    — []Token → *ast.Program
  → semantic/  — validates AST, infers return counts, enforces security checks
  → codegen/   — *ast.Program → Go source string
```

Semantic analysis produces `exprReturnCounts map[ast.Expression]int` which is passed to codegen via `generator.SetExprReturnCounts(...)`. This tells codegen how many values an expression returns so it can emit the right `val, err := f()` split for `onerr`.

The formatter (`formatter/`) is a separate pipeline that re-parses and pretty-prints. The LSP (`lsp/`) wraps the compiler pipeline and is independent of the above.

---

## Lexer (`internal/lexer/`)

**Key files:** `lexer.go`, `token.go`

### INDENT/DEDENT

Kukicha is indentation-sensitive. The lexer converts 4-space indentation changes into `TOKEN_INDENT` / `TOKEN_DEDENT` tokens using an `indentStack []int` (always starts at `[0]`).

- Indentation must be multiples of 4 spaces — tabs are rejected
- Each increase must be exactly +4 spaces
- Dedents can skip multiple levels (e.g., 8→0 emits two `TOKEN_DEDENT`)
- Blank lines and comment-only lines do not affect the indent stack

### Line continuation

`TOKEN_NEWLINE` is suppressed (continuation mode) when:
- Previous token was `TOKEN_PIPE` (`|>`)
- Next line starts with `|>` (checked by `isPipeAtStartOfNextLine`)
- Next line starts with `onerr` (checked by `isOnErrAtStartOfNextLine`)
- Inside `[]` or `{}` (`braceDepth > 0`)

`()` (parentheses) do NOT suppress newlines when inside a function literal body — closures need `INDENT/DEDENT` for their block structure.

### Adding a new keyword

Add the keyword string → `TokenType` mapping in `token.go`'s `keywords` map and define the `TokenType` constant there.

---

## Parser (`internal/parser/`)

**Key files** (split from the original monolithic `parser.go`):

| File | Contents |
|------|---------|
| `parser.go` | Core struct, `New`, `Parse`, token helpers (`peekToken`, `consume`, `advance`, …) |
| `parser_type.go` | `parseTypeAnnotation` and all type sub-parsers |
| `parser_decl.go` | Declaration parsers (`parseFunctionDecl`, `parseTypeDecl`, `parseVarDeclaration`, …) |
| `parser_stmt.go` | Statement parsers (`parseBlock`, `parseStatement`, `parseIfStmt`, `parseForStmt`, `parseOnErrClause`, …) |
| `parser_expr.go` | Expression parsers (`parseExpression`, `parsePipeExpr`, `parseArrowLambda`, …) |

### Design

- Recursive descent
- **Error collection** (not fail-fast): errors are appended to `p.errors`, parsing continues. This allows multiple errors per compile.
- `peekToken()` calls `skipIgnoredTokens()` first, which skips `TOKEN_COMMENT` and `TOKEN_SEMICOLON`
- Context-sensitive keywords: `list`, `map`, `channel` are only keywords when followed by `of` in a type context — this allows them as variable names elsewhere. `empty` and `error` are context-sensitive too: `isIdentifierFollower()` checks if the next token indicates identifier usage (assignment, operators, delimiters); if so, they parse as identifiers instead of `EmptyExpr`/`ErrorExpr`

### Key helpers

| Helper | Purpose |
|--------|---------|
| `peekToken()` | Look at next meaningful token (skips comments/semicolons) |
| `consume(type, msg)` | Advance and return token, or record error |
| `skipNewlines()` | Skip `TOKEN_NEWLINE` tokens |
| `parseBlock()` | Parse `INDENT … DEDENT` block into `*ast.BlockStmt` |
| `parseTypeAnnotation()` | Parse any Kukicha type (`list of T`, `map of K to V`, `reference T`, etc.) |
| `isIdentifierFollower()` | Returns true if next token indicates `empty`/`error` is being used as an identifier |

### Adding a new statement

1. Add a `TOKEN_*` keyword in `lexer/token.go`
2. In `parser_stmt.go`'s `parseStatement()` switch, add a `case TOKEN_*:` branch calling your new `parseXxxStmt()` method
3. Return the new `*ast.XxxStmt` node (defined in `ast/ast.go`)

### Adding a new expression

1. Hook into `parsePrimaryExpr()` in `parser_expr.go` for new literal/prefix forms, or the operator precedence chain for binary forms
2. Return a new `*ast.XxxExpr` node

---

## AST (`internal/ast/`)

**Key file:** `ast.go` (~960 lines)

### Interface hierarchy

```
Node
├── Declaration  (declNode marker)  — FunctionDecl, TypeDecl, ImportDecl, …
├── Statement    (stmtNode marker)  — IfStmt, ForRangeStmt, ReturnStmt, …
├── Expression   (exprNode marker)  — CallExpr, PipeExpr, ArrowLambda, …
└── TypeAnnotation (typeNode marker) — ListType, MapType, ReferenceType, …
```

### Convention for new nodes

Every node must implement `Node`:
```go
type XxxStmt struct {
    Token lexer.Token  // The keyword token (for position info)
    // ... fields
}

func (s *XxxStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *XxxStmt) Pos() Position {
    return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File}
}
func (s *XxxStmt) stmtNode() {} // or declNode() / exprNode() / typeNode()
```

Always store the keyword's `lexer.Token` as the first field — it carries line/column for error messages.

### OnErrClause

`OnErrClause` is **not** a standalone `Statement` or `Expression`. It is an optional field on `VarDeclStmt`, `AssignStmt`, and `ExpressionStmt`. The `Handler` field holds the parsed error handler expression (`PanicExpr`, `EmptyExpr`, `DiscardExpr`, `ReturnExpr`, or a default value expression).

---

## Semantic Analysis (`internal/semantic/`)

**Key files:**

| File | Contents |
|------|---------|
| `semantic.go` | Core `Analyzer` struct, `New`, `Analyze`, `Warnings`, `ReturnCounts`, `error`/`warn` helpers |
| `semantic_declarations.go` | Package name validation, skill validation, declaration collection/analysis (`collectDeclarations`, `analyzeFunctionDecl`, …) |
| `semantic_statements.go` | Statement analysis (`analyzeBlock`, `analyzeStatement`, `analyzeIfStmt`, `analyzeForRangeStmt`, `analyzeVarDeclStmt`, …) |
| `semantic_expressions.go` | Expression analysis (`analyzeExpression`, `analyzeIdentifier`, `analyzeBinaryExpr`, `analyzePipeExprMulti`, `analyzeListLiteral`, …) |
| `semantic_onerr.go` | `onerr` clause analysis (`analyzeOnErrClause`, `funcReturnsError`, `analyzeStringInterpolation`) |
| `semantic_types.go` | Type annotation validation and conversion (`validateTypeAnnotation`, `typeAnnotationToTypeInfo`, `typesCompatible`) |
| `semantic_helpers.go` | Pure utilities (`isValidIdentifier`, `extractPackageName`, `isExported`, `isNumericType`, `primitiveTypeFromString`) |
| `semantic_calls.go` | `analyzeCallExpr`, `analyzeMethodCallExpr` |
| `semantic_security.go` | Security checks (`checkSQLInterpolation`, `checkHTMLNonLiteral`, `checkFetchInHandler`, `checkFilesInHandler`, `checkShellRunNonLiteral`, `checkRedirectNonLiteral`, `isInHTTPHandler`) |
| `symbols.go` | Symbol table and type info |
| `stdlib_registry_gen.go` | Generated stdlib return-count registry |

### Two-pass analysis

1. **`collectDeclarations()`** — registers all top-level types, interfaces, and function signatures into the symbol table (so functions can call each other regardless of order)
2. **`analyzeDeclarations()`** — validates function bodies, infers `exprReturnCounts`, enforces security checks

### exprReturnCounts

The analyzer infers how many values an expression returns and stores it in `a.exprReturnCounts[expr]`. Codegen reads this to decide whether to emit `val, err := f()` (2-return) vs `val := f()` (1-return) for pipe + onerr chains.

When a new stdlib function is added to a `.kuki` file, run `make genstdlibregistry` to regenerate `stdlib_registry_gen.go` so the analyzer knows the function's return count.

### Security checks

Security checks run during `analyzeDeclarations()`. The analyzer detects "inside an HTTP handler" by checking whether the enclosing `FunctionDecl` has an `http.ResponseWriter` parameter. The `inOnerr bool` field tracks whether the analyzer is currently inside an `onerr` block (used to enforce `{error}` not `{err}`).

### Adding a new security check

Add a new `checkXxx` method in `semantic_security.go` following the existing pattern, then call it from `analyzeMethodCallExpr` in `semantic_calls.go`. Emit an error via `a.error(expr.Pos(), ...)`.

### stdlib_registry_gen.go

Auto-generated by `cmd/genstdlibregistry/`. Do not edit manually. Regenerate with:
```bash
make genstdlibregistry   # or: make generate (runs everything)
```

---

## Codegen (`internal/codegen/`)

**Key files** (split from the original monolithic `codegen.go`):

| File | Contents |
|------|---------|
| `codegen.go` | Core `Generator` struct, public API (`New`, `SetSourceFile`, `SetExprReturnCounts`, …), `Generate`, top-level `generatePackage/Skill/Declaration`, output helpers (`write`, `writeLine`, `emitLineDirective`, `uniqueId`) |
| `codegen_decl.go` | Declaration generators (`generateTypeDecl`, `generateInterfaceDecl`, `generateFunctionDecl`, `generateFunctionLiteral`, `generateArrowLambda`, `generateTypeAnnotation`, `generateReturnTypes`, …) |
| `codegen_stmt.go` | Statement generators (`generateBlock`, `generateStatement`, `generateVarDeclStmt`, `generateAssignStmt`, `generateReturnStmt`, `coerceReturnValue`, `generateIfStmt`, `generateFor*`, `generateSwitch*`, `generateSelect*`) |
| `codegen_expr.go` | Expression generators (`exprToString`, `generatePipeExpr`, `generateCallExpr`, string interpolation, …) |
| `codegen_onerr.go` | `onerr` code generation (`generateOnErrVarDecl`, `generateOnErrHandler`, pipe-chain onerr, …) |
| `codegen_imports.go` | Import generation and auto-import scanning (`generateImports`, `scanStmtForAutoImports`, …) |
| `codegen_stdlib.go` | Stdlib/generics type inference (`inferStdlibTypeParameters`, `zeroValueForType`, …) |
| `codegen_walk.go` | Unified AST visitor (`walkProgram`, `walkBlock`, `walkStmt`, `walkExpr`) and `needsXxx` helpers |

### Generator state

| Field | Purpose |
|-------|---------|
| `output strings.Builder` | Accumulates generated Go source |
| `indent int` | Current indentation level (each level = 1 tab in output) |
| `autoImports map[string]bool` | Packages auto-imported by codegen (e.g., `fmt`, `errors`) |
| `pkgAliases map[string]string` | Collision aliases (e.g., `json` → `kukijson`) |
| `funcDefaults map[string]*FuncDefaults` | Default parameter info for wrapper generation |
| `placeholderMap map[string]string` | Generic placeholder substitution (`"any"→"T"`, `"any2"→"K"`) |
| `currentOnErrVar string` | Error variable name in active `onerr` block (for `{error}` interpolation) |
| `currentReturnIndex int` | Index of return value being generated (-1 if not in return); used to emit `*new(T)` vs `nil` for bare `empty` in generic stdlib functions |
| `exprReturnCounts map[ast.Expression]int` | From semantic — drives `onerr` multi-value split |

### onerr code generation

`onerr` is the most complex part of codegen. The generator wraps the call in a temporary assignment, checks the error, and runs the handler. `currentOnErrVar` holds the generated error variable name so that `{error}` in string interpolation inside the block resolves to it.

### Generics via placeholders

When `isStdlibIter` is true (or per-function for `stdlib/slice`), the generator detects `any`/`any2` placeholders in type annotations and:
1. Builds a `placeholderMap` mapping placeholder → Go type param name (`T`, `K`)
2. Emits `[T any, K comparable]` on the function signature
3. Substitutes placeholders throughout parameter and return types
4. Emits `*new(T)` or `*new(K)` for bare `empty` in return position when the return type uses a placeholder (otherwise emits `nil`)

All `stdlib/slice` functions are generic: `genericSafe` map lists `[T any]` functions, `comparableSafe` map lists `[K comparable]` functions (`Unique`, `Contains`, `IndexOf`), and `GroupBy` gets both `[T any, K comparable]`.

Application code never sees this — it just calls functions normally.

### Writing to output

Use `g.write(str)` (no indent) or `g.writeLine(str)` (with current indent + newline). Do not write to `g.output` directly.

---

## Adding a Feature: End-to-End Example

**Example: add `repeat N times` loop (`for i repeat 5`)**

1. **Lexer** (`lexer/token.go`): add `TOKEN_REPEAT`, add `"repeat"` to `keywords` map
2. **AST** (`ast/ast.go`): add `ForRepeatStmt { Token, Count Expression, Body *BlockStmt }`
3. **Parser** (`parser/parser_stmt.go`): in `parseStatement()` add `case lexer.TOKEN_REPEAT:` → `parseForRepeatStmt()`
4. **Semantic** (`semantic/semantic_statements.go`): in `analyzeStatement()` add `case *ast.ForRepeatStmt:` → validate `Count` is numeric
5. **Codegen** (`codegen/codegen.go`): in `generateStatement()` add `case *ast.ForRepeatStmt:` → emit `for _i := 0; _i < N; _i++ { ... }`
6. **Tests**: add test cases in each package's `*_test.go`

### Bidirectional Loops

The `for` loop now supports the `through` keyword for descending loops:
- `for i from 10 through 0`: generates `for i := 10; i >= 0; i--`
- `for i from 0 through 10`: generates `for i := 0; i <= 10; i++`
- The compiler generates a bidirectional condition `(start <= end && i <= end) || (start > end && i >= end)` to handle cases where bounds are variables.

---

## Test Patterns

Each package has its own `*_test.go`. The pattern is:
- **Lexer tests**: feed source string → check token types/lexemes
- **Parser tests**: feed source string → check AST structure
- **Codegen tests**: feed source string → check generated Go string (often with `strings.Contains`)
- **Semantic tests**: feed source string → check error messages

Run all tests:
```bash
make test
```
