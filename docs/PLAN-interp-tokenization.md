# Plan: Lexer-Level String Interpolation Tokenization

**Tech debt item:** #18 (String re-parsing for interpolated pipes)
**Status:** Phase 1 complete

## Problem

String interpolation expressions are parsed by creating a sub-parser per `{expr}`:

1. **Parser** (`parser_expr.go:600-632`): `parseInterpolationExpr()` wraps each expression
   in `func __interp__() \n    print(expr)`, creates `parser.New()`, parses, and extracts the
   expression AST. Runs once per `{expr}` at parse time.

2. **Codegen fallback** (`codegen_expr.go:573-634`): When `Parts` is empty (parse failure),
   `transformInterpolatedExpr()` and `parseAndGenerateInterpolatedExpr()` do the same sub-parser
   trick at codegen time.

3. **Regex limitation**: Both paths split on `\{([a-zA-Z_][^}]*)\}` which cannot handle nested
   `}` in expressions (e.g., `{MyStruct{field: 1}}` or `{m["key"]}`).

Commit `9afadc7` moved parsing from codegen-time to parse-time (a major improvement), but the
sub-parser architecture and regex splitting remain.

## Solution: Lexer emits interpolation tokens

The lexer splits interpolated strings into multiple tokens (like JS template literals, Kotlin,
Swift). The parser then calls its normal `parseExpression()` on the token stream — no sub-parser,
no regex, and nested braces work automatically via brace-depth tracking.

### New token types (JS template literal model)

| Token | Emitted for | Example in `"Hello {name}, age {age}!"` |
|-------|-------------|------------------------------------------|
| `TOKEN_STRING` | Non-interpolated strings (unchanged) | `"plain string"` |
| `TOKEN_STRING_HEAD` | Leading literal before first `{` | `"Hello "` |
| `TOKEN_STRING_MID` | Literal between two interpolations | `", age "` |
| `TOKEN_STRING_TAIL` | Trailing literal after last `}` | `"!"` |

Token sequence for `"Hello {name}, age {age}!"`:
```
TOKEN_STRING_HEAD  "Hello "
TOKEN_IDENTIFIER   "name"
TOKEN_STRING_MID   ", age "
TOKEN_IDENTIFIER   "age"
TOKEN_STRING_TAIL  "!"
```

For `"{name}"` (no literal parts):
```
TOKEN_STRING_HEAD  ""
TOKEN_IDENTIFIER   "name"
TOKEN_STRING_TAIL  ""
```

### Brace depth tracking

The lexer adds an `interpStack []int` field. Each entry is the brace depth within that
interpolation level:

- `{` inside an interpolation: increment `interpStack[top]`, emit `TOKEN_LBRACE` normally
- `}` at `interpStack[top] == 0`: end of interpolation, pop stack, resume string scanning
- `}` at `interpStack[top] > 0`: decrement, emit `TOKEN_RBRACE` normally

This correctly handles `{MyStruct{field: 1}}` — the inner `{}` increments/decrements brace
depth, and only the outer `}` ends the interpolation.

### Interpolation detection

`{` inside a string triggers interpolation only when followed by an identifier-start character
(`[a-zA-Z_]`). This matches the current regex behavior:
- `{name}` → interpolation
- `{2,}` → literal `{2,}` (regex quantifier in a string)
- `\{name\}` → literal `{name}` (PUA sentinels `\uE000`/`\uE001`, no change)

---

## Phases

### Phase 1: Lexer + parser (this PR) ✅ DONE

**Lexer (`internal/lexer/`):**
1. Add `TOKEN_STRING_HEAD`, `TOKEN_STRING_MID`, `TOKEN_STRING_TAIL` to `token.go`
2. Add `interpStack []int` field to `Lexer` struct
3. Modify `scanString()`: when hitting `{` + alpha, emit `TOKEN_STRING_HEAD`, push interp, return
4. Add `scanStringContinuation()`: called after `}` closes interp; on `{` + alpha emit
   `TOKEN_STRING_MID`, on `"` emit `TOKEN_STRING_TAIL`
5. Modify `}` handling in `scanToken()`: check interpStack before normal handling
6. Modify `{` handling in `scanToken()`: increment interpStack braceDepth if in interp

**Parser (`internal/parser/`):**
1. Rewrite `parseStringLiteral()`: handle `TOKEN_STRING_HEAD` → `parseExpression()` →
   `TOKEN_STRING_MID`/`TOKEN_STRING_TAIL` loop
2. Build `StringLiteral.Parts` directly from parsed expressions
3. Delete `parseStringParts()`, `parseInterpolationExpr()`, `interpolationRe` regex

**Tests:**
- New lexer tests for all three interp token types, nested braces, escaped braces, empty parts
- Update parser interpolation tests to work through new token path
- All existing codegen/integration tests must pass unchanged

### Phase 2: Codegen cleanup ✅ DONE

Deleted from `codegen_expr.go`:
- `parseStringInterpolation()` — regex-based splitter
- `transformInterpolatedExpr()` — `as`/pipe string transformer
- `parseAndGenerateInterpolatedExpr()` — codegen-time re-parser
- `generateStringInterpolation()` — fallback path

Simplified `generateStringLiteral()` to only use `generateStringFromParts()`.
Simplified `errorValueExpr()` and panic handler to use `generateStringLiteral()`
(which handles `{error}` substitution via `currentOnErrVar` set by `renderHandler`).
Removed `regexp` and `parser` imports from `codegen_expr.go`.
Added `\sep` detection to `scanExprForAutoImports` for proper `path/filepath` auto-import.

### Phase 3: Semantic cleanup (follow-up PR)

Delete fallback regex path in `analyzeStringInterpolation()` (`semantic_onerr.go:116-154`).

### Phase 4: Edge case tests (follow-up PR)

Add tests for cases that were previously impossible:
- `{MyStruct{field: val}}` — struct literal in interpolation
- `{m[key]}` — map access
- `{f(func() int { return 1 })}` — closure in interpolation (if sensible)
- Verify `\sep` sentinel handling is unchanged

---

## Risk assessment

| Risk | Mitigation |
|------|-----------|
| Lexer state machine complexity | interpStack is tiny; JS/Kotlin/Swift prove the pattern |
| Line continuation interaction | Strings are single-line — no INDENT/DEDENT inside interp |
| `\sep` sentinel | Handled in escape sequence before `{` check — no conflict |
| braceDepth for line continuation | Inner `{}` in interp increment both interpStack and braceDepth |
| Formatter uses its own parser | Formatter re-parses independently; will need same token changes |

## Files affected

| File | Phase | Change |
|------|-------|--------|
| `internal/lexer/token.go` | 1 | Add 3 token types + String() cases |
| `internal/lexer/lexer.go` | 1 | interpStack, scanString split, scanStringContinuation, `{}` handling |
| `internal/lexer/lexer_test.go` | 1 | New interpolation tokenization tests |
| `internal/parser/parser_expr.go` | 1 | Rewrite parseStringLiteral, delete parseStringParts/parseInterpolationExpr |
| `internal/parser/parser.go` | 1 | Add STRING_MID/TAIL to isIdentifierFollower |
| `internal/parser/parser_interpolation_test.go` | 1 | Update tests for new token-based parsing |
| `internal/codegen/codegen_expr.go` | 2 | Delete 4 fallback functions, remove regexp/parser imports |
| `internal/codegen/codegen_onerr.go` | 2 | Simplify panic handler to use exprToString |
| `internal/codegen/codegen_stdlib.go` | 2 | Simplify errorValueExpr to use generateStringLiteral |
| `internal/codegen/codegen_walk.go` | 2 | Remove fallback in exprHasNonPrintfInterpolation |
| `internal/codegen/codegen_imports.go` | 2 | Add \sep detection to scanExprForAutoImports |
| `internal/formatter/lexer.go` | — | Mirror token changes (formatter has its own lexer) |
