# Cleanup: onerr refactoring and bare for codegen

## 1. Bare `for` loop codegen (DONE)

**File:** `internal/codegen/codegen.go`

**Change:** `generateForConditionStmt` now emits `for {` (idiomatic Go) instead of `for true {` when the condition is the synthetic `true` literal created by the parser for bare `for` loops.

---

## 2. onerr refactored to statement-level only (DONE)

### What changed

`onerr` is no longer an expression operator. It is a statement-level clause attached to `VarDeclStmt`, `AssignStmt`, or `ExpressionStmt`.

### Why

The previous implementation had `onerr` in the expression precedence chain as `OnErrExpr`. When onerr appeared as a nested sub-expression (e.g., inside a function call argument), the codegen fell back to an IIFE pattern:

```go
func() interface{} { if err := expr; err != nil { return handler }; return expr }()
```

This caused:
- **Type erasure:** Returns `interface{}`, breaking typed assignments.
- **Broken control flow:** `onerr return` inside the IIFE only returns from the anonymous function, not the enclosing function.

### What was removed

- `OnErrExpr` AST node (expression-level onerr)
- `parseOnErrExpr()` in the parser
- `generateOnErrExpr()` IIFE codegen function
- Pipe+onerr restructuring hacks in codegen (the codegen previously had to detect `PipeExpr{Right: OnErrExpr}` and restructure it)

### What was added

- `OnErrClause` struct in the AST (not a node — a helper struct with `Token` and `Handler`)
- `OnErr *OnErrClause` field on `VarDeclStmt`, `AssignStmt`, `ExpressionStmt`
- `parseOnErrClause()` in the parser
- `generateOnErrAssign()` in codegen (assignment onerr was previously unsupported)
- OnErr handler scanning in `checkStmtForInterpolation`, `checkStmtForPrint`, `checkStmtForErrors`

### Breaking change

`return f() onerr default` no longer parses. Write:
```kukicha
val := f() onerr default
return val
```

No existing `.kuki` files used this pattern.

### Files modified

| File | Changes |
|------|---------|
| `internal/ast/ast.go` | Added `OnErrClause`; added `OnErr` to 3 statement types; removed `OnErrExpr` |
| `internal/parser/parser.go` | Removed `parseOnErrExpr`; added `parseOnErrClause`; updated statement parsers |
| `internal/codegen/codegen.go` | Removed IIFE; removed pipe+onerr hacks; added `generateOnErrAssign`; updated scan helpers |
| `internal/semantic/semantic.go` | Removed `analyzeOnErrExpr`; added `analyzeOnErrClause` |
| `internal/formatter/printer.go` | Removed `OnErrExpr` case; added `onErrSuffix` helper |
| `internal/parser/parser_test.go` | Updated test from expression-level to statement-level |
