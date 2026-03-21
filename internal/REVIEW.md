# Code Review: `internal/` Go Code

**Date:** 2026-03-21
**Scope:** All production (non-test, non-generated) Go files in `internal/`
**Approach:** Expert Go review focused on correctness bugs, panics, performance, concurrency, and dead code

---

## Summary

| Severity | Count |
|----------|-------|
| High     | 12    |
| Medium   | 16    |
| Low      | 14    |

The compiler is well-structured with clean separation across pipeline stages. The most concerning issues cluster around: (1) child `Generator` instances missing semantic state, (2) the formatter emitting duplicate/wrong output for methods and variadics, (3) missing recursion in AST walkers, and (4) several nil-safety gaps in semantic analysis.

---

## HIGH Severity

### H1. `codegen_stmt.go:297` — `generateIfStmt` creates a fresh `Generator` that lacks semantic state

```go
tempGen := New(g.program)
```

When an `if` statement has an init clause, a brand-new `Generator` is created. It does **not** inherit `exprTypes`, `exprReturnCounts`, `pkgAliases`, `placeholderMap`, `currentOnErrVar`, `funcDefaults`, `mcpTarget`, `reservedNames`, or `sourceFile`. If the init expression uses string interpolation, piped expressions, package aliases, `empty`, or `print` under MCP mode, the generated Go code will be wrong. Should use `g.childGenerator(0)` or manually copy the required fields.

### H2. `codegen_expr.go:209-220` — `generatePipedSwitchExpr` manually constructs a `Generator` missing critical fields

The manual `&Generator{...}` is missing `exprReturnCounts`, `currentReturnIndex`, `stdlibModuleBase`, `reservedNames`, `funcDefaults`, `mcpTarget`, and `currentOnErrVar`/`currentOnErrAlias`. This means `onerr` inside a piped switch expression body won't generate proper error variable splitting, `print` won't get MCP treatment, and `uniqueId` calls could collide with user variables.

### H3. `codegen_decl.go:336` — Arrow lambda with unknown return type generates invalid Go

```go
return fmt.Sprintf("func(%s) { return %s }", params, bodyStr)
```

A `return` inside a void function body will be rejected by the Go compiler (`too many return values`). Should either infer the return type or omit `return`.

### H4. `codegen_onerr.go:439-471` — `stmtHasExplain` missing recursion into switch/select/else blocks

Checks `IfStmt`, `ForRangeStmt`, `ForNumericStmt`, `ForConditionStmt` but misses `SwitchStmt` case bodies, `TypeSwitchStmt` case bodies, `SelectStmt` case bodies, `ElseStmt.Alternative` bodies, and `GoStmt` blocks. If `onerr explain` appears in any of these, the `fmt` import won't be added, causing a Go compilation error.

### H5. `codegen_imports.go:378-444` — `scanExprForAutoImports` missing recursion into many expression types

Doesn't recurse into `ErrorExpr`, `PanicExpr`, `ReturnExpr`, `TypeCastExpr`, `TypeAssertionExpr`, `AddressOfExpr`, `DerefExpr`, `BlockExpr`, or `PipedSwitchExpr`. If a `\sep` string appears in any of these positions, the `path/filepath` auto-import will be missed.

### H6. `printer.go:130-163` — Formatter emits method signatures twice

When `decl.Receiver != nil`, line 136 calls `p.writeLine(...)` emitting a partial method signature. Then the `else` branch at line 154 calls `p.writeLine(...)` again with the full signature. Both lines are written to output — the method signature appears twice in formatted code.

### H7. `printer.go:182` — Variadic parameter formatting is reversed

Emits `name many type` but Kukicha syntax is `many name type`. The formatter produces `numbers many int` instead of `many numbers int`.

### H8. `formatter.go:267-309` — `PrinterWithComments` missing `BreakStmt`/`ContinueStmt`

These statement types are handled in `Printer.printStatement` but missing from `printStatementWithComments`. Functions containing `break` or `continue` will have those statements silently dropped from formatted output.

### H9. `semantic_declarations.go:146-149` — All type aliases marked `TypeKindFunction`

```go
if decl.AliasType != nil {
    typeKind = TypeKindFunction
}
```

Unconditionally tags aliases as `TypeKindFunction`, even for `type ID int64` or `type Names list of string`. Non-function aliases get incorrect type kind, causing false negatives in type checking.

### H10. `semantic_declarations.go:367-368` — Global multi-var declarations always infer type from first value

```go
} else if len(stmt.Values) > 0 {
    varType = a.analyzeExpression(stmt.Values[0])
```

For `var a, b = 1, "hello"`, every variable gets the type of `Values[0]`. Should use `Values[i]` when available.

### H11. `semantic_calls.go:74-79` — Generic placeholder `result` resolved incorrectly

All four placeholders (`any`, `any2`, `ordered`, `result`) are resolved to the same `elementType`. But `result` represents the transform output type (e.g., in `slice.Map`), not the input element type. This causes wrong type inference for lambda parameters in `slice.Map` and `concurrent.Map`.

### H12. `parser/parser.go:184-190` — `peekNextToken` does not skip ignored tokens at `pos+1`

`peekNextToken()` calls `skipIgnoredTokens()` at the current position, then returns `p.tokens[p.pos+1]`. But the token at `pos+1` might be a comment or semicolon. This makes `peekNextToken` non-equivalent to `peekAt(1)` (which correctly skips), causing incorrect lookahead decisions throughout the parser when comments appear between tokens.

---

## MEDIUM Severity

### M1. `lexer.go:664-680` — `\x` escape silently drops characters on incomplete hex sequence

If source ends with `\xA` (one hex digit), `h1` is consumed but `h2` is never read. No error is emitted and the `\xA` silently vanishes. Should emit an error for incomplete hex escapes.

### M2. `lexer.go:842` — Column calculation can go negative

```go
Column: l.column - len([]rune(lexeme)),
```

For `TOKEN_NEWLINE` (where `l.column` is reset to 0), this yields `Column: -1`. Similarly affected: `TOKEN_DEDENT` and `TOKEN_INDENT`.

### M3. `lexer.go:514-526` — Tab over-stripping in `dedentTripleQuote`

When stripping `minIndent` spaces, a tab counts as 4 but the loop doesn't handle partial tab consumption. If `minIndent = 2` and a line starts with a tab (worth 4), 4 columns are stripped instead of 2.

### M4. `token.go:135-345` — `TOKEN_DIRECTIVE` missing from `TokenType.String()`

Falls through to `default: return "UNKNOWN"`. Any debug output involving a directive token shows "UNKNOWN" instead of "DIRECTIVE".

### M5. `token.go:420-428` — `Keywords()` returns mutable internal slice

Any caller that appends to the returned slice corrupts the cache for all subsequent callers.

### M6. `parser_expr.go:226-228` — Incorrect position revert for `reference` without `of`

`p.pos--` after `p.match(TOKEN_REFERENCE)` doesn't account for tokens skipped by `skipIgnoredTokens()`. If directives exist between the previous token and `reference`, the revert lands on the directive.

### M7. `parser_stmt.go:159-212` — `parseIfStmt` re-parse duplicates pending directives

When the first `parseExpression()` is followed by a semicolon, `p.pos = savePos` reverts position but `p.pendingDirectives` still contains directives collected during the first parse. Re-parsing collects them again.

### M8. `parser_stmt.go:562-608` — `parseInterpolatedStringLiteral` infinite loop risk

The loop breaks only on `TOKEN_STRING_TAIL` or error. If the lexer produces `TOKEN_STRING_MID` tokens without ever producing `TOKEN_STRING_TAIL`, the loop runs forever. Adding `&& !p.isAtEnd()` to the condition would fix this.

### M9. `semantic_onerr.go` — Missing `onerr continue`/`onerr break` validation outside loops

`analyzeOnErrClause` validates `ShorthandReturn` but not `ShorthandContinue` or `ShorthandBreak`. `onerr continue` outside a loop produces no semantic error.

### M10. `semantic_types.go:241-246` — `typesCompatible` treats same-named types from different packages as compatible

```go
return unqualifiedName(t1.Name) == unqualifiedName(t2.Name)
```

`http.Request` and `custom.Request` are incorrectly accepted as compatible.

### M11. `semantic_types.go:152-161` — `isReferenceType` can stack overflow on recursive type aliases

If a named type's symbol resolves back to itself through aliases, this recurses infinitely.

### M12. `semantic_expressions.go:494` — Off-by-one in list element error message

When ranging `Elements[1:]`, the error says element `i+1` but the actual position is `i+2` in the original list.

### M13. `codegen_stdlib.go:184-187` — Dead code: `inferMapsTypeParameters`/`isStdlibMaps` defined but never called

The maps stdlib package won't get generic type parameter inference.

### M14. `lsp/server.go:138,154,171,185` — Nil pointer dereference on `req.Params`

All LSP handlers do `json.Unmarshal(*req.Params, ...)` without checking for nil. A malformed client sending null params causes a panic.

### M15. `lsp/document.go:57,67` — Analysis runs under write lock

`newDocument` calls `doc.analyze()` (lexer + parser + semantic) while holding `ds.mu.Lock()`. This blocks all other document operations for the duration of analysis.

### M16. `lower.go:49` — IIFE for multi-return base uses `any` return type

```go
baseExpr = fmt.Sprintf("func() any { val, %s := %s; return val }()", ...)
```

Returns `any` instead of the actual first return type, losing type information for downstream pipe steps.

---

## LOW Severity

### L1. `lexer.go:655-663` — `\s` (not `\sep`) silently drops the backslash

`\s` becomes `s` with no warning. A user writing `\something` may not realize the backslash is eaten.

### L2. `lexer.go:143-174` — Duplicated `\n`/`\r` newline handling

Nearly identical logic for line continuation, indent tracking, and token emission. Should share a helper.

### L3. `lexer.go:566-570` — `scanStringFromContent` copies the entire source array per triple-quoted string

O(n * m) where n is source length and m is number of triple-quoted strings.

### L4. `lexer.go:895-897` — `isLetter` uses `unicode.IsLetter` vs `isAlpha` (ASCII-only) inconsistency

`isLetter` is only used in `isOnErrAtStartOfNextLine`, creating subtle edge-case differences.

### L5. `parser_stmt.go:466` — Swallowed error from `consume` in `parseSelectCase`

```go
second, _ := p.consume(lexer.TOKEN_IDENTIFIER, ...)
```

If consume fails, `second.Lexeme` is empty, producing an empty binding name in the AST.

### L6. `parser_decl.go:56` — Swallowed error from `consume` in `parseSkillDecl`

Missing colon causes the next `p.advance()` to consume the wrong token.

### L7. `parser_type.go:139` — Swallowed error from `consume` for qualified type

Produces invalid type name `"pkg."` when the identifier after `.` is missing.

### L8. `semantic_helpers.go:78` — `isNumericType` and `isBitwiseType` don't guard against nil

Unlike `isReferenceType` which checks `t == nil`, these will panic on nil input.

### L9. `semantic_expressions.go:220-226` — `analyzeExpressionMulti` analyzes `IndexExpr.Left` twice

First via `analyzeExpression(e.Left)`, then again inside `analyzeIndexExpr(e)`. Errors could double-fire.

### L10. `semantic_declarations.go:19` — `checkPackageName` uses `strings.Contains(sourceFile, "stdlib/")`

A user project at `/home/user/my-stdlib/helper/main.kuki` would incorrectly bypass the package name collision check.

### L11. `codegen_stmt.go:95-108` — `generatePipedSwitchStmt` mutates AST nodes in-place

Temporarily changes `stmt.Expression`, which is unsafe if the AST is cached or accessed concurrently.

### L12. `formatter.go:143-148` — Trailing comment handling is O(n) per comment

Materializes and rewrites the entire `strings.Builder` content for each trailing comment.

### L13. `formatter/comments.go:222-233` — All three branches of leading-comment loop do the same thing

The conditional logic distinguishing "standalone" vs "adjacent" comments has no effect.

### L14. `lsp/diagnostics.go:40` — Regex compiled on every call

`regexp.MustCompile` inside `errorToDiagnostic` should be a package-level var.

---

## Patterns to Watch

1. **Child generators**: Any new code that creates `tempGen := New(g.program)` instead of `g.childGenerator(0)` will lose semantic state. Consider adding a linter check or removing the `New` constructor from codegen-internal use.

2. **AST walker completeness**: `stmtHasExplain`, `scanExprForAutoImports`, `needsStringInterpolation`, and similar walkers must be updated whenever new AST node types are added. Consider a visitor pattern or code generation to avoid manual maintenance.

3. **`peekNextToken` vs `peekAt(1)`**: These should be equivalent but aren't. Either fix `peekNextToken` to skip ignored tokens at `pos+1`, or replace all its call sites with `peekAt(1)`.

4. **Formatter coverage**: The `PrinterWithComments` switch must mirror every case in `Printer.printStatement`. Missing cases silently drop statements. Consider a shared dispatch table.
