# Transpiler Known Limitations

Tracked limitations in the Kukicha compiler (`internal/`). Each entry includes the relevant source location(s) and a description of the gap. Items are grouped by compiler phase.

---

## Semantic Analysis / Type Inference

### Named arguments not supported on method calls
**File:** `internal/semantic/semantic_calls.go:183`

Named arguments (e.g., `files.Copy(from: src, to: dst)`) work for locally declared functions but are rejected on method call expressions (`obj.Method(name: val)`). A hard error is emitted at the call site. Positional arguments must be used instead.

### Named arguments not supported on imported/unknown functions
**File:** `internal/semantic/semantic_calls.go:77`

When a function is resolved via the stdlib registry or Go stdlib (rather than a local `func` declaration), named arguments are also rejected. This covers most stdlib functions that lack a local Kukicha `func` wrapper.

### Method return type resolution limited to hand-coded stdlib entries
**File:** `internal/semantic/semantic_calls.go:321`

For non-stdlib methods the analyzer returns `TypeKindUnknown`. Full method resolution would require tracking the declared type of the receiver. Only a small set of Go stdlib methods (`time.Time.*`, `bufio.Scanner.*`, `regexp.Regexp.*`, `exec.ExitError.*`) have hand-coded type info; everything else is unknown.

### Method return counts not inferred in codegen
**File:** `internal/codegen/codegen_stdlib.go:497`

`inferReturnCount` skips `MethodCallExpr` nodes entirely, returning `(0, false)`. This means onerr splits on method-call results fall back to a default of 1 return value rather than the real count from the method signature.

### `exprTypes` map not yet fully consumed by codegen
**File:** `internal/semantic/semantic.go:19`

The semantic analyzer populates `exprTypes` (expression → inferred type) during analysis and passes it to codegen. It is currently consumed only by `isErrorOnlyReturn()` for error-only pipe step detection. Planned uses — contextual type inference for untyped arrow lambda parameters, smarter pipe chain error handling, typed zero-value generation — are not yet implemented.

### Pipe placeholder `_` type is always Unknown
**File:** `internal/semantic/semantic_expressions.go:247`

The `_` placeholder in piped calls (e.g., `todo |> json.MarshalWrite(w, _)`) is always typed `TypeKindUnknown`. The second-argument position cannot be type-checked even when the surrounding context would supply enough information.

### Kukicha stdlib functions lack per-position type info
**File:** `internal/semantic/semantic_calls.go` (knownExternalReturns fallback)

The generated `generatedStdlibRegistry` contains return *counts* only (no `TypeKind` per position). Per-position types are only available for Go stdlib functions in `generatedGoStdlib`. As a result, Kukicha stdlib function returns type as `TypeKindUnknown` per position even when the `.kuki` signature is fully typed.

### `TypeKindUnknown` used as a numeric-compatible type
**File:** `internal/semantic/semantic_helpers.go:79`

`isNumericType` returns `true` for `TypeKindUnknown`. This means arithmetic on a value whose type could not be inferred passes semantic analysis without error, silently losing type safety.

---

## Parser

### Field access parsed as zero-argument method call
**File:** `internal/parser/parser_expr.go:322`

`obj.Field` (no parentheses) is represented in the AST as a `MethodCallExpr` with an empty argument list. The compiler does not distinguish struct field reads from zero-arg method calls at the parse level. This works in practice because codegen emits the same `.Field` syntax for both, but it prevents the semantic analyzer from ever knowing whether a dotted access is a field read or a method call.

### Directives silently ignored on interface declarations
**File:** `internal/parser/parser_decl.go` (directive attachment)

`# kuki:deprecated` and other compiler directives are collected and attached to `FunctionDecl` and `TypeDecl` nodes. Interface declarations (`InterfaceDecl`) have no `Directives` field and are not wired into the directive drain path; directives placed before an interface are silently dropped.

### Bitwise AND not supported
**File:** `internal/lexer/lexer.go:273`

The `&` character is rejected at the lexer level with an error message directing users to `and` or `&&`. Bitwise AND (`&`), address-of for non-struct contexts, and bitwise-AND-assign (`&=`) are not available.

---

## Deprecation Tracking

### Type deprecation not warned at usage sites
**File:** `internal/semantic/semantic_declarations.go` (deprecation check), noted in `internal/semantic/directive_test.go:95`

`# kuki:deprecated` on a `type` declaration is parsed and stored in `deprecatedTypes`, but the analyzer does not check `deprecatedTypes` when a type name appears in type annotations or struct literals. Only function deprecation warnings are emitted at call sites.

---

## String Interpolation

### Interpolated expressions not semantically analyzed
**File:** `internal/semantic/semantic_onerr.go:82`

`analyzeStringInterpolation` uses a regex to find `{expr}` placeholders in string literals. It validates that the placeholder is non-empty and enforces the `{error}` vs `{err}` rule inside `onerr`, but does not parse or type-check the expression inside the braces. References to undefined variables or type-mismatched expressions inside string interpolations pass analysis undetected.

---

## Formatter

### Block-form arrow lambda body not rendered
**File:** `internal/formatter/printer.go:725`

The formatter (`kukicha fmt`) renders multi-statement block lambdas as `params => ...` instead of expanding the body. Single-expression lambdas (`(x T) => expr`) format correctly; block lambdas (`(x T) =>\n    stmt\n    return val`) lose their body in formatted output.

---

## LSP / IDE

### Hover does not resolve local variables or parameters
**File:** `internal/lsp/hover.go:85`

Hovering over a local variable or function parameter shows nothing. The hover handler only resolves top-level declarations (functions, types, interfaces). Resolving locals would require tracking the cursor position within a scope chain, which is not yet implemented.

### `findSymbolInScope` always returns nil
**File:** `internal/lsp/hover.go:237`

The function exists and is called, but its body is a stub that unconditionally returns `nil`. No local-scope symbol lookup is performed.

---

## Summary

| Phase | Items |
|-------|-------|
| Semantic / type inference | 7 |
| Parser | 3 |
| Deprecation | 1 |
| String interpolation | 1 |
| Formatter | 1 |
| LSP | 2 |
| **Total** | **15** |
