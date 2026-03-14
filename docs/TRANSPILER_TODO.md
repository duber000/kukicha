# Transpiler Known Limitations

Tracked limitations in the Kukicha compiler (`internal/`). Each entry includes the relevant source location(s) and a description of the gap. Items are grouped by compiler phase.

---

## Semantic Analysis / Type Inference

### Method return type resolution limited to hand-coded Go stdlib entries
**File:** `internal/semantic/semantic_calls.go:321`

User-defined methods are resolved via `registerMethod()` and `resolveMethodType()`. Kukicha stdlib methods are resolved via the generated registry. However, Go stdlib methods beyond a small hand-coded set (`time.Time.*`, `bufio.Scanner.*`, `regexp.Regexp.*`, `exec.ExitError.*`) still return `TypeKindUnknown`. Full resolution would require extending `cmd/gengostdlib/` to generate method entries from `go/types` package method sets.

### Field access parsed as zero-argument method call
**File:** `internal/parser/parser_expr.go:322`

`obj.Field` (no parentheses) is represented in the AST as a `MethodCallExpr` with an empty argument list. The `IsCall` field distinguishes field reads from method calls at the semantic level (`resolveFieldType` vs `resolveMethodType`), and codegen emits the same `.Field` syntax for both. The limitation is purely at the AST representation level — a dedicated `FieldAccessExpr` node would be cleaner but is not functionally necessary.

---

## Summary

| Phase | Items |
|-------|-------|
| Semantic / type inference | 1 |
| Parser (cosmetic) | 1 |
| **Total** | **2** |
