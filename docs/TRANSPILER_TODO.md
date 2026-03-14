# Transpiler Known Limitations

Tracked limitations in the Kukicha compiler (`internal/`). Each entry includes the relevant source location(s) and a description of the gap. Items are grouped by compiler phase.

---

## Semantic Analysis / Type Inference

### Method return type resolution limited to hand-coded stdlib entries
**File:** `internal/semantic/semantic_calls.go:321`

For non-stdlib methods the analyzer returns `TypeKindUnknown`. Full method resolution would require tracking the declared type of the receiver. Only a small set of Go stdlib methods (`time.Time.*`, `bufio.Scanner.*`, `regexp.Regexp.*`, `exec.ExitError.*`) have hand-coded type info; everything else is unknown.

### `exprTypes` map not yet fully consumed by codegen
**File:** `internal/semantic/semantic.go:19`

The semantic analyzer populates `exprTypes` (expression → inferred type) during analysis and passes it to codegen. It is currently consumed only by `isErrorOnlyReturn()` for error-only pipe step detection. Planned uses — contextual type inference for untyped arrow lambda parameters, smarter pipe chain error handling, typed zero-value generation — are not yet implemented.

### Pipe placeholder `_` type is always Unknown
**File:** `internal/semantic/semantic_expressions.go:247`

The `_` placeholder in piped calls (e.g., `todo |> json.MarshalWrite(w, _)`) is always typed `TypeKindUnknown`. The second-argument position cannot be type-checked even when the surrounding context would supply enough information.

---

## Parser

### Field access parsed as zero-argument method call
**File:** `internal/parser/parser_expr.go:322`

`obj.Field` (no parentheses) is represented in the AST as a `MethodCallExpr` with an empty argument list. The compiler does not distinguish struct field reads from zero-arg method calls at the parse level. This works in practice because codegen emits the same `.Field` syntax for both, but it prevents the semantic analyzer from ever knowing whether a dotted access is a field read or a method call.

---

## String Interpolation

### Complex interpolated expressions not semantically analyzed
**File:** `internal/semantic/semantic_onerr.go:96`

`analyzeStringInterpolation` validates bare identifier references (e.g., `{name}`, `{count}`) by parsing and analyzing them. However, complex expressions inside braces (e.g., `{obj.Field}`, `{fn(x)}`, `{a + b}`) are skipped — they pass through without semantic analysis. Only the `{error}` vs `{err}` rule inside `onerr` is enforced for all forms.

---

## Parser / Semantic Interaction

### Inline `onerr as` syntax fails to parse
**File:** `internal/parser/parser_stmt.go`
**Test:** `TestOnerrAliasInterpolationIsValid`

`onerr as myErr return` (inline alias with immediate handler) is rejected by the parser with `'onerr as <ident>' must be followed by an indented block`. The block form (`onerr as myErr` + indented body) works correctly. The parser's `parseOnErrClause` requires an `INDENT` after `as <ident>` and has no fallback for inline handlers. This causes a persistent test failure.

---

## Summary

| Phase | Items |
|-------|-------|
| Semantic / type inference | 3 |
| Parser | 1 |
| String interpolation | 1 |
| Parser / semantic interaction | 1 |
| **Total** | **6** |
