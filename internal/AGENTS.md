# internal/CLAUDE.md

Compiler internals reference. Read this when working in `internal/`. For language syntax and build commands see the root `CLAUDE.md`.

## Pipeline Overview

```
source (.kuki)
  → lexer/     — runes → []Token (INDENT/DEDENT injected)
  → parser/    — []Token → *ast.Program
  → semantic/  — validates AST, infers return counts, enforces security checks
  → codegen/   — *ast.Program → IR (via Lowerer) → Go source string (via emitIR)
```

Semantic analysis produces two maps that are passed to codegen:
- `exprReturnCounts map[ast.Expression]int` — passed via `generator.SetExprReturnCounts(...)`. Tells codegen how many values an expression returns so it can emit the right `val, err := f()` split for `onerr`.
- `exprTypes map[ast.Expression]*TypeInfo` — passed via `generator.SetExprTypes(...)`. Records the inferred type of every analyzed expression. Consumed by codegen for: error-only pipe step detection (`isErrorOnlyReturn`), piped switch return type inference, `empty` keyword resolution, expression return type inference (`inferExprReturnType`), and typed zero-value generation (`zeroValueForType`). In `analyzePipeExprMulti`, types are explicitly recorded on pipe step nodes via `recordType(right, types[0])` since steps bypass `analyzeExpression`. Pipe placeholder `_` identifiers get the piped value's type recorded when inside a call with a known function signature.

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
- Error messages include actionable detail (e.g., nearest valid indent level, valid dedent targets)

### Line continuation

`TOKEN_NEWLINE` is suppressed (continuation mode) when:
- Previous token was `TOKEN_PIPE` (`|>`)
- Next line starts with `|>` (checked by `isPipeAtStartOfNextLine`)
- Next line starts with `onerr` (checked by `isOnErrAtStartOfNextLine`)
- Inside `[]` or `{}` (`braceDepth > 0`)

`()` (parentheses) do NOT suppress newlines when inside a function literal body — closures need `INDENT/DEDENT` for their block structure.

### Adding a new keyword

Add the keyword string → `TokenType` mapping in `token.go`'s `keywords` map and define the `TokenType` constant there.

### Directives (`TOKEN_DIRECTIVE`)

Comments starting with `# kuki:` are emitted as `TOKEN_DIRECTIVE` instead of `TOKEN_COMMENT`. The lexer's `scanComment` checks the prefix and selects the token type. `TOKEN_DIRECTIVE` is excluded from `lastTokenType` tracking (like `TOKEN_COMMENT`).

### String escape sequences and PUA sentinels

`scanString` handles escape sequences in the switch on the character after `\`. Two kinds of escapes exist:

**Compile-time (character substitution):** The escaped sequence maps directly to a Unicode Private Use Area (PUA) sentinel stored in the token value. Codegen's `escapeString` converts the sentinel back to the literal character when emitting Go string literals — no `fmt.Sprintf` needed.

| Escape | PUA sentinel | Emitted as |
|--------|-------------|-----------|
| `\{`   | `\uE000`    | literal `{` |
| `\}`   | `\uE001`    | literal `}` |

**Runtime (expression injection):** The escape expands to a Go expression evaluated at runtime. The sentinel is included in literal `TOKEN_STRING_PART`/`TOKEN_STRING_MID`/`TOKEN_STRING_TAIL` lexemes; codegen's `generateStringFromParts` detects `\uE002` and expands it to `string(filepath.Separator)`.

| Escape  | PUA sentinel | Emitted as |
|---------|-------------|-----------|
| `\sep`  | `\uE002`    | `string(filepath.Separator)` — auto-imports `path/filepath` |

`\sep` is a multi-character escape: `scanStringEscape` checks `l.peek() == 'e' && l.peekNext() == 'p'` before consuming the `ep` suffix.

`generateStringLiteral` and `exprHasNonPrintfInterpolation` (in `codegen_walk.go`) both check `strings.ContainsRune(value, '\uE002')` to correctly handle strings that contain `\sep` but no `{expr}` interpolation (i.e., where `StringLiteral.Interpolated` is `false`).

### String interpolation tokenization

For interpolated strings (containing `{expr}`), the lexer emits multiple tokens instead of a single `TOKEN_STRING`:

| Token | Purpose | Example in `"Hello {name}, age {age}!"` |
|-------|---------|------------------------------------------|
| `TOKEN_STRING_HEAD` | Leading literal before first `{` | `"Hello "` |
| `TOKEN_STRING_MID` | Literal between two interpolations | `", age "` |
| `TOKEN_STRING_TAIL` | Trailing literal after last `}` | `"!"` |

Between HEAD→MID and MID→MID/TAIL, normal expression tokens are emitted. The parser calls `parseExpression()` directly on these tokens — no sub-parser needed.

**Brace depth tracking:** `interpStack []int` on the `Lexer` tracks nesting within each interpolation level. `{` inside an interpolation increments `interpStack[top]`; `}` at depth 0 ends the interpolation and resumes string scanning via `scanStringContinuation()`. This correctly handles nested braces like `{MyStruct{field: 1}}`.

**Interpolation detection:** `isInterpStart()` checks if the character after `{` is alpha or `_`. Non-identifier starts like `{2,}` are treated as literal text, matching the old regex behavior.

Non-interpolated strings still emit a single `TOKEN_STRING`.

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
- Context-sensitive keywords: `list`, `map`, `channel` are only keywords when followed by `of` in a type context — this allows them as variable names elsewhere. `empty` and `error` are context-sensitive too: `isIdentifierFollower()` checks if the next token indicates identifier usage (assignment, operators, delimiters); if so, they parse as identifiers instead of `EmptyExpr`/`ErrorExpr`. For `empty` specifically, the parser also treats it as an identifier when followed by `|>`, `)`, or `,` — so `empty |> iterator.Values()` and `print(empty)` work when `empty` is a user-defined variable.

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

### PipedSwitchExpr

`PipedSwitchExpr` represents both regular and typed piped switches:

- Regular: `expr |> switch`
- Typed: `expr |> switch as v`

The AST stores the body as a `PipedSwitchBody`, implemented by both `*SwitchStmt` and `*TypeSwitchStmt`. `parsePipeExpr()` creates a `PipedSwitchExpr` when it sees `TOKEN_SWITCH` after `|>`, and dispatches to `parseSwitchBody()` or `parseTypeSwitchBody()` depending on whether `as binding` is present.

In codegen, value-producing piped switches are wrapped in an IIFE. Regular piped switches generate `switch left { ... }`; typed piped switches generate `switch v := left.(type) { ... }`. Return-type inference for typed piped switches special-cases `return v` so the IIFE can stay strongly typed instead of falling back to `any`.

### Directive

`Directive` represents a `# kuki:name args...` annotation. It has `Name string`, `Args []string`, and `Token lexer.Token`. `FunctionDecl`, `TypeDecl`, and `InterfaceDecl` all have a `Directives []Directive` field. The parser collects `TOKEN_DIRECTIVE` tokens in `skipIgnoredTokens` and attaches them to the next declaration via `drainDirectives()`.

Currently supported directives:
- `# kuki:deprecated "message"` — marks a function/type/interface as deprecated; semantic analysis warns at usage sites
- `# kuki:security "category"` — marks a function as security-sensitive (categories: `sql`, `html`, `fetch`, `files`, `redirect`, `shell`); drives compile-time security checks in `semantic_security.go`

### OnErrClause

`OnErrClause` is **not** a standalone `Statement` or `Expression`. It is an optional field on `VarDeclStmt`, `AssignStmt`, and `ExpressionStmt`. The `Handler` field holds the parsed error handler expression (`PanicExpr`, `EmptyExpr`, `DiscardExpr`, `ReturnExpr`, or a default value expression). Shorthand forms use boolean flags instead of `Handler`: `ShorthandReturn`, `ShorthandContinue`, `ShorthandBreak`.

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
| `semantic_calls.go` | `analyzeCallExpr`, `analyzeMethodCallExpr`, `analyzeFieldAccessExpr` |
| `semantic_security.go` | Security checks (`checkSQLInterpolation`, `checkHTMLNonLiteral`, `checkFetchInHandler`, `checkFilesInHandler`, `checkShellRunNonLiteral`, `checkRedirectNonLiteral`, `isInHTTPHandler`) |
| `symbols.go` | Symbol table and type info |
| `stdlib_registry_gen.go` | Generated Kukicha stdlib return-count registry (from `.kuki` files) |
| `go_stdlib_gen.go` | Generated Go stdlib return-count + type registry (from `go/importer`) |

### TypeKind: `TypeKindNil`

The `empty` keyword has its own type kind (`TypeKindNil`) in `symbols.go`. This distinguishes `empty`-as-nil-literal from `empty`-as-variable-name. When semantic analysis encounters an `EmptyExpr` or an `Identifier` named `"empty"` that isn't shadowed by a user variable, it records `TypeKindNil`. Codegen checks this to decide whether to emit `nil` or preserve the variable name `empty`. The `isReferenceType()` helper in `semantic_types.go` determines which types are nil-compatible (references, lists, maps, channels, functions, interfaces), and `typesCompatible()` uses it so `TypeKindNil` is accepted where a reference type is expected.

### Struct literal validation

The semantic analyzer validates struct literal field names and types at compile time. During `collectDeclarations()`, each struct type's field names and types are stored in `TypeInfo.Fields`. When a `StructLiteralExpr` is analyzed, the analyzer resolves the struct's symbol and checks that every field name exists on the struct and that the value type is compatible with the declared field type. Unknown fields produce `unknown field 'x' on struct 'Y'` errors; type mismatches produce `cannot use T1 as T2 in field 'x' of struct 'Y'`.

### Method and field resolution

`TypeInfo.Methods` (added alongside `Fields`) maps method names to their function `TypeInfo`. During `collectDeclarations()`, `registerMethod()` attaches each method's signature to its receiver type's symbol. At analysis time, `FieldAccessExpr` nodes resolve through `resolveFieldType()`, while `MethodCallExpr` nodes resolve through `resolveMethodType()`. Both handle pointer/reference receivers by dereferencing first.

### Analysis passes

0. **`collectDeprecations()`** — scans all declarations for `# kuki:deprecated` directives, populating `deprecatedFuncs`/`deprecatedTypes` maps
1. **`collectDeclarations()`** — registers all top-level types, interfaces, and function signatures into the symbol table (so functions can call each other regardless of order)
2. **`analyzeDeclarations()`** — validates function bodies, infers `exprReturnCounts`, enforces security checks, warns on deprecated calls

### exprReturnCounts

The analyzer infers how many values an expression returns and stores it in `a.exprReturnCounts[expr]`. Codegen reads this to decide whether to emit `val, err := f()` (2-return) vs `val := f()` (1-return) for pipe + onerr chains.

For typed piped switches, semantic analysis does not fully analyze the switch as a statement. Instead it analyzes the piped input expression plus the return expressions inside each case body, entering a fresh scope per case/otherwise branch so the `as` binding is defined there.

When a new stdlib function is added to a `.kuki` file, run `make genstdlibregistry` to regenerate `stdlib_registry_gen.go` so the analyzer knows the function's return count.

### knownExternalReturns and Go stdlib registry

`knownExternalReturns` is a unified map of qualified function name → return count, built from two auto-generated sources:

1. **`generatedStdlibRegistry`** (`stdlib_registry_gen.go`) — return counts, **per-position return types**, and **parameter names** for Kukicha stdlib functions, generated from `.kuki` files by `cmd/genstdlibregistry`. Uses the shared `goStdlibEntry` struct.
2. **`generatedGoStdlib`** (`go_stdlib_gen.go`) — return counts **and per-position type info** for Go stdlib functions, generated from Go's `go/importer` by `cmd/gengostdlib`. Includes `TypeKind` and optional `Name` for each return position (e.g., `os.WriteFile` → `[{TypeKindNamed, "error"}]`).

Both registries use the `goStdlibEntry` and `goStdlibType` types defined in `stdlib_types.go`. The Kukicha registry additionally populates `ParamNames` for named argument support and `DefaultValues` for default parameter filling in codegen.

In `analyzeMethodCallExpr`, the Go stdlib registry is checked first, then the Kukicha registry. Both provide full type info for package-level functions and methods. Named arguments on Kukicha stdlib calls are resolved using `ParamNames` from the registry.

To add a new Go stdlib function: add it to the curated list in `cmd/gengostdlib/main.go` and run `make gengostdlib`.

### Security checks

Security checks run during `analyzeDeclarations()`. The analyzer detects "inside an HTTP handler" by checking whether the enclosing `FunctionDecl` has an `http.ResponseWriter` parameter. The `inOnerr bool` field tracks whether the analyzer is currently inside an `onerr` block (used to enforce `{error}` not `{err}`).

### Adding a new security check

Security-sensitive functions are annotated with `# kuki:security "category"` directives in their `.kuki` source files. The `genstdlibregistry` generator scans these and emits `generatedSecurityFunctions`. Security checks in `semantic_security.go` use `securityCategory()` to look up the category (with alias support for `httphelper.X → http.X`).

To add a new security check: (1) add `# kuki:security "category"` to the function in its `.kuki` file, (2) add a `checkXxx` method in `semantic_security.go`, (3) call it from `analyzeMethodCallExpr` in `semantic_calls.go`, (4) run `make genstdlibregistry`.

### stdlib_types.go

Defines the shared `goStdlibType` and `goStdlibEntry` structs used by both generated registries. Not auto-generated — edit directly when adding fields. Exports accessors for codegen: `GetStdlibEntry(name)`, `GetSliceGenericClass(name)`, `GetSecurityCategory(name)`, `IsKnownInterface(name)`.

### stdlib_registry_gen.go

Auto-generated by `cmd/genstdlibregistry/`. Do not edit manually. Regenerated automatically by `make build` via `go generate ./...`. Standalone: `make genstdlibregistry`.

Contains five maps:
- `generatedStdlibRegistry` — function name → `goStdlibEntry` (count, types, param names)
- `generatedStdlibDeprecated` — function name → deprecation message (from `# kuki:deprecated` directives)
- `generatedSecurityFunctions` — function name → security category (from `# kuki:security` directives)
- `generatedSliceGenericClass` — function name → generic class (`T`, `K`, or `TK`, from placeholder usage)
- `generatedStdlibInterfaces` — interface names (from `InterfaceDecl` nodes in `.kuki` files)

### go_stdlib_gen.go

Auto-generated by `cmd/gengostdlib/`. Do not edit manually. Regenerated automatically by `make build` via `go generate ./...`. Standalone: `make gengostdlib`.

Contains two maps:
- `generatedGoStdlib` — function name → `goStdlibEntry` (count, types)
- `generatedGoInterfaces` — qualified interface type names (e.g., `io.Reader`, `http.Handler`) extracted via `go/types`

To add a new Go stdlib function, add it to the curated `packages` list in `cmd/gengostdlib/main.go`. Interfaces are auto-detected from all scanned packages.

---

## IR (`internal/ir/`)

**Key file:** `ir.go`

The IR (Intermediate Representation) package defines Go-level imperative nodes used between AST lowering and code emission. The Lowerer (in `codegen/lower.go`) transforms high-level Kukicha constructs (pipes, onerr chains, piped switches) into IR node sequences, and the emitter (in `codegen/emit.go`) walks IR blocks to produce Go source text.

### IR Node types

| Node | Purpose |
|------|---------|
| `Block` | Ordered sequence of IR nodes |
| `Assign` | `names := expr` or `names = expr` |
| `VarDecl` | `var name type [= value]` |
| `IfErrCheck` | `if errVar != nil { body }` |
| `Goto` | `goto Label` |
| `Label` | `LabelName:` |
| `ScopedBlock` | Bare `{ ... }` block for variable scoping |
| `RawStmt` | Pre-rendered Go statement (escape hatch) |
| `ReturnStmt` | `return val1, val2, ...` |
| `ExprStmt` | Standalone expression (`continue`, `break`, `panic(...)`) |
| `Comment` | `// text` |

The IR is intentionally thin — it models only the constructs needed by the onerr/pipe lowering passes. Other codegen paths still emit Go text directly.

---

## Codegen (`internal/codegen/`)

**Key files** (split from the original monolithic `codegen.go`):

| File | Contents |
|------|---------|
| `codegen.go` | Core `Generator` struct, public API (`New`, `SetSourceFile`, `SetExprReturnCounts`, …), `Generate`, top-level `generatePackage/Skill/Declaration`, output helpers (`write`, `writeLine`, `emitLineDirective`, `uniqueId`) |
| `codegen_decl.go` | Declaration generators (`generateTypeDecl`, `generateInterfaceDecl`, `generateFunctionDecl`, `generateFunctionLiteral`, `generateArrowLambda`, `generateTypeAnnotation`, `generateReturnTypes`, …) |
| `codegen_stmt.go` | Statement generators (`generateBlock`, `generateStatement`, `generateVarDeclStmt`, `generateAssignStmt`, `generateReturnStmt`, `coerceReturnValue`, `generateIfStmt`, `generateFor*`, `generateSwitch*`, `generateSelect*`) |
| `codegen_expr.go` | Expression generators (`exprToString`, `generatePipeExpr`, `generateCallExpr`, string interpolation, …) |
| `codegen_onerr.go` | `onerr` code generation (`generateOnErrVarDecl`, `generateOnErrHandler`); delegates pipe-chain and piped-switch onerr to Lowerer |
| `lower.go` | `Lowerer` struct — transforms pipe chains, onerr clauses, and piped switches into IR nodes. Key methods: `lowerPipeChain`, `lowerOnErrPipeChain`, `lowerPipedSwitchVarDecl`, `lowerOnErrWithExplicitErr` |
| `emit.go` | `emitIR` / `emitIRNode` — walks IR blocks and emits Go source via `g.writeLine`. Handles `Assign`, `VarDecl`, `IfErrCheck`, `Goto`, `Label`, `ScopedBlock`, `RawStmt`, `ReturnStmt`, `ExprStmt`, `Comment` |
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
| `isStdlibIter bool` | True if generating `stdlib/iterator` code (enables iter-specific generic transpilation) |
| `sourceFile string` | Source file path for detecting stdlib packages |
| `currentFuncName string` | Current function being generated (for context-aware decisions) |
| `currentReturnTypes []ast.TypeAnnotation` | Return types of current function (for type coercion in returns and `onerr` zero-value generation) |
| `currentOnErrVar string` | Error variable name in active `onerr` block (for `{error}` interpolation) |
| `currentOnErrAlias string` | User-specified alias in `onerr as e` blocks |
| `currentReturnIndex int` | Index of return value being generated (-1 if not in return); used to emit `*new(T)` vs `nil` for bare `empty` in generic stdlib functions |
| `tempCounter int` | Counter for generating unique temporary variable names via `uniqueId()` |
| `exprReturnCounts map[ast.Expression]int` | From semantic — drives `onerr` multi-value split |
| `exprTypes map[ast.Expression]*TypeInfo` | From semantic — used by `isErrorOnlyReturn()` for error-only pipe step detection |
| `mcpTarget bool` | True if targeting MCP (Model Context Protocol) — affects main function generation |
| `reservedNames map[string]bool` | User-declared identifiers — `uniqueId` skips these to avoid collisions |
| `stdlibModuleBase string` | Base module path for rewriting `"stdlib/X"` imports (default: `github.com/duber000/kukicha`) |

### onerr code generation (Lowerer + IR)

`onerr` is the most complex part of codegen. It is now split into two phases:

1. **Lowering** (`lower.go`): The `Lowerer` transforms pipe chains and onerr clauses into IR nodes (`ir.Assign`, `ir.IfErrCheck`, `ir.Goto`, `ir.Label`, etc.). This makes the logic testable independently of string emission.
2. **Emission** (`emit.go`): `emitIR` walks the IR block and produces Go source text via `g.writeLine`.

Simple onerr cases (single call, no pipe) still use direct emission in `codegen_onerr.go`. Pipe chain and piped switch onerr delegate to the Lowerer. `currentOnErrVar` holds the generated error variable name so that `{error}` in string interpolation inside the block resolves to it.

`lowerOnErrWithExplicitErr` handles multi-return cases where the user provides the error variable as the last LHS name (e.g., `a, b, err := f() onerr ...`). If the last name is `_`, it replaces it with a generated unique error variable, since Go's blank identifier cannot be read as a value in `if _ != nil`.

For statement-level multi-return calls with `onerr` (no LHS capture), `lowerOnErrStmt` uses `inferReturnCount` to add `_` discard variables for non-error return values. When `inferReturnCount` fails (e.g., external method calls), the `.kuki` source should use `_ :=` to explicitly capture the non-error return value (e.g., `_ := f() onerr ...`).

The Lowerer handles three cases per pipe step:
1. **Multi-return** (count ≥ 2): split into `val, err := call()`; check err
2. **Error-only** (count == 1 and type is `error`): `err := call()`; check err; keep current pipe variable unchanged (the step produces no data value)
3. **Single value** (count == 1, non-error): `pipe := call()`; advance pipe variable

Error-only detection uses `isErrorOnlyReturn()` which checks both `exprReturnCounts` (count == 1) and `exprTypes` (type is `error`). Known error-only Go stdlib functions (`os.WriteFile`, `os.Remove`, etc.) are registered in `knownExternalReturns` in `semantic_calls.go` with proper type info.

Piped switches participate in the same machinery. For `pipe |> switch ... onerr ...`, codegen first lowers the upstream pipe chain with error checks, then runs either a regular switch or typed type-switch over the final pipe value. Typed piped switches are supported in both statement position and value-producing declarations/assignments.

### `empty` keyword in codegen

When codegen encounters an `Identifier` with value `"empty"`, it consults `exprTypes` to decide what to emit:

- **`TypeKindNil`** (not shadowed) → emit `nil`. In generic stdlib context with a placeholder return type, emit `*new(T)` or `*new(K)` instead.
- **Not `TypeKindNil`** (shadowed by a user variable) → emit `empty` as-is, preserving the variable name.

This prevents `empty |> iterator.Values()` from silently becoming `iterator.Values(nil)` when the user declared `empty := list of int{}`.

### Arrow lambdas: no implicit `it`

Arrow lambdas do **not** support an implicit `it` parameter. This feature was removed — lambdas must declare their parameters explicitly. If you see `() => print(it)`, that's a bug; it should be `(it SomeType) => print(it)`. The removed code lived in `codegen_walk.go` (`exprHasItIdentifier`, `blockHasItIdentifier`) and `semantic_expressions.go` (`arrowLambdaHasIt`).

### Generics via placeholders

When `isStdlibIter` is true (or per-function for `stdlib/slice`), the generator detects `any`/`any2` placeholders in type annotations and:
1. Builds a `placeholderMap` mapping placeholder → Go type param name (`T`, `K`)
2. Emits `[T any, K comparable]` on the function signature
3. Substitutes placeholders throughout parameter and return types
4. Emits `*new(T)` or `*new(K)` for bare `empty` in return position when the return type uses a placeholder (otherwise emits `nil`)

All `stdlib/slice` functions are generic. The generic classification (`T`, `K`, or `TK`) is auto-derived from placeholder usage in `.kuki` function signatures and stored in `generatedSliceGenericClass` (generated by `genstdlibregistry`). `inferSliceTypeParameters` reads this via `semantic.GetSliceGenericClass()`.

Application code never sees this — it just calls functions normally.

### String literal codegen (`codegen_expr.go`)

`generateStringLiteral` routes to one of three paths:
- **Plain string** (`Interpolated == false` and no `\uE002` sentinel): emits `"escaped"` via `escapeString`
- **Sep-only string** (`Parts` empty but `\uE002` present): calls `generateSepOnlyString` — splits on `\uE002` and emits `string(filepath.Separator)` concatenation
- **Interpolated string** (`Parts` populated): calls `generateStringFromParts` — iterates pre-parsed `StringInterpolation` parts, building `fmt.Sprintf` with `%v` placeholders for expression parts and escaped literals. Handles `\uE002` expansion and onerr `{error}` substitution inline.

The lexer splits interpolated strings into `TOKEN_STRING_HEAD`, expression tokens, `TOKEN_STRING_MID`/`TOKEN_STRING_TAIL`. The parser's `parseInterpolatedStringLiteral()` calls `parseExpression()` directly on the token stream, building `StringLiteral.Parts` with no sub-parser or regex. Semantic analysis and codegen both consume the pre-parsed AST nodes directly.

`parseStringPartsOrInterpolation(lit)` is a shared helper that returns `(format, args)` from pre-parsed Parts. Used by the printf-style method path.

`escapeString` handles compile-time PUA sentinels (`\uE000` → `{`, `\uE001` → `}`) and standard Go escapes (`\n`, `\t`, `\\`, `\"`, `\x00`). It is **not** responsible for `\uE002` — that sentinel is expanded before `escapeString` sees it.

### Child generators for inline code blocks

Use `g.childGenerator(extraIndent)` when generating inline function bodies (function literals, arrow lambda blocks). The child shares the parent's semantic state by reference (program, autoImports, pkgAliases, exprTypes, etc.) and writes to a fresh `strings.Builder`. This replaces the old pattern of manually copying 8+ fields into a throwaway `Generator`.

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
