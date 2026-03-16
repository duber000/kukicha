# Tier 5: Architecture Improvements — Implementation Plan

## Approach

Items are ordered by impact-to-effort ratio. Items 21 and 20 are quick wins; items 17/19 are coupled and should be done together; item 18 is the riskiest and deferred last.

---

## Phase 1: Quick Wins (Items 21, 20)

### Item 21: Error message rewriting is fragile

**Goal:** Determine if `rewriteGoErrors()` is still needed; remove or harden it.

**Steps:**

1. **Write a test** that compiles a `.kuki` file with an intentional Go-level error (e.g., type mismatch deferred to Go). Capture stderr and verify the error references the `.kuki` file path (not `.go`).
2. **Temporarily disable** `rewriteGoErrors()` and re-run the test. If Go's `//line` directives make the error already reference `.kuki`, the function is dead code.
3. **If still needed:** Replace the naive `strings.ReplaceAll` with a regex that only matches file paths at the start of error lines (Go format: `path:line:col: message`). Add a test for the rewriting logic.
4. **If not needed:** Delete `rewriteGoErrors()` and add a comment explaining that `//line` directives handle source mapping.

**Files:** `cmd/kukicha/main.go`
**Risk:** Low — removal is safe if tests pass; hardening is a small regex change.

### Item 20: Formatter comment handling has zero test coverage

**Goal:** Add test coverage for `ExtractComments()` and `AttachComments()`. Optionally share parse results.

**Steps:**

1. **Add unit tests** in `internal/formatter/comments_test.go`:
   - Leading comments (directly above a declaration)
   - Trailing comments (same line as a statement)
   - Multiple consecutive comments
   - Comments separated by blank lines from the next node
   - Comments inside nested blocks (if/for/switch)
   - Inline comments after expressions
   - Directive comments (`# kuki:...`) — verify they're excluded from formatting
2. **Add integration tests** in `internal/formatter/formatter_test.go`:
   - Round-trip test: format a file with comments, verify comments are preserved in correct positions
   - Edge case: comment at end of file
   - Edge case: comment between two functions
3. **Fix any bugs** discovered by the new tests (likely: comments lost between blocks, or trailing comment attached to wrong node).

**Files:** `internal/formatter/comments.go`, `internal/formatter/formatter_test.go` (new: `comments_test.go`)
**Risk:** Low — adding tests only; fixes are localized to formatter.

> **Note:** Sharing parse results between compiler and formatter (the "Option A" from the debt doc) is a larger refactor. Defer it — the formatter's independent parsing is not a correctness issue, just duplication. Tests are the priority.

---

## Phase 2: IR Expansion (Items 17, 19 — done together)

Items 17 and 19 are coupled: both require extending the IR layer. Doing them together avoids two rounds of IR changes.

### Item 17: Replace RawStmt escape hatch with proper IR nodes

**Goal:** Reduce RawStmt usage from ~8 production occurrences to 0-2 (keeping only truly one-off cases like the fallback comment).

**Steps:**

1. **Add new IR node types** in `internal/ir/ir.go`:
   - `ReturnStmt` — `return expr1, expr2, ...`
   - `ExprStmt` — standalone expression (e.g., `panic(...)`, `continue`, `break`)
   - `Comment` — `// text` (for the fallback comment case)

2. **Update emitter** in `internal/codegen/emit.go`:
   - Add `case *ir.ReturnStmt`, `case *ir.ExprStmt`, `case *ir.Comment` to `emitIRNode`

3. **Update lowerer** in `internal/codegen/lower.go`:
   - Replace `RawStmt{Code: "continue"}` → `ExprStmt{Code: "continue"}`
   - Replace `RawStmt{Code: "break"}` → `ExprStmt{Code: "break"}`
   - Replace `RawStmt{Code: "return ..."}` → `ReturnStmt{Values: [...]}`
   - Replace the handler block `RawStmt` with a proper `Block` of IR nodes (lower each statement in the handler body individually)
   - Replace the switch `RawStmt` with either a new `SwitchStmt` IR node or leave as `RawStmt` (switches are complex; a dedicated node may not be worth it yet)

4. **Add IR-level tests** in `internal/ir/ir_test.go` and `internal/codegen/emit_test.go` for the new node types.

5. **Run full test suite** to verify identical output.

**Files:** `internal/ir/ir.go`, `internal/codegen/emit.go`, `internal/codegen/lower.go`
**Risk:** Medium — changing IR nodes could affect edge cases, but integration tests catch regressions.

### Item 19: Replace temporary generators with IR for lambda codegen

**Goal:** Eliminate throwaway `Generator` instances in lambda/function literal codegen.

**Steps:**

1. **Add IR nodes** (may already exist from Item 17):
   - `FuncLiteral` — parameters, return types, body `Block`
   - Or simpler: `InlineBlock` — a block of statements to be emitted inline with custom indent

2. **Refactor `generateFunctionLiteral()`** (`codegen_decl.go:243-275`):
   - Instead of creating a temp generator, lower the function literal body into an IR `Block`
   - Emit the block inline using the existing emitter with adjusted indent

3. **Refactor `generateArrowLambda()`** (`codegen_decl.go:321-350`):
   - Same approach — lower body to IR, emit inline

4. **Alternative (lower effort):** If full IR lowering is too invasive, extract a `childGenerator()` helper method on `Generator` that copies all necessary fields. This eliminates the error-prone manual field copying (8+ fields) while keeping the same architecture:
   ```go
   func (g *Generator) childGenerator(extraIndent int) *Generator
   ```

5. **Add tests** verifying lambda output is unchanged.

**Files:** `internal/codegen/codegen_decl.go`, possibly `internal/ir/ir.go`
**Risk:** Medium — same as Item 17. The `childGenerator()` alternative is lower risk.

**Decision point:** If IR expansion from Item 17 naturally covers lambda bodies, go full IR. If it doesn't (e.g., lambdas have many statement types not yet in IR), use the `childGenerator()` helper as an incremental improvement.

---

## Phase 3: Parser-Level Fix (Item 18 — deferred, highest risk)

### Item 18: String re-parsing for interpolated pipes

**Goal:** Store parsed AST expressions in `StringLiteral` interpolation slots instead of raw strings that need re-parsing at codegen time.

**Steps:**

1. **Modify AST** (`internal/ast/ast.go`):
   - Add `ParsedSlots []Expression` field to `StringLiteral` (or change `Slots` from `[]string` to a union type)
   - Keep `Value string` as-is for the raw string representation

2. **Modify parser** (`internal/parser/parser_expr.go`):
   - In `parseStringLiteral()`, when an interpolation slot contains complex expressions (pipes, calls), parse them into AST nodes immediately
   - Store the parsed expressions in `ParsedSlots`
   - For simple identifiers, either store as `Identifier` nodes or keep as strings

3. **Modify codegen** (`internal/codegen/codegen_expr.go`):
   - In `parseStringInterpolation()`, check `ParsedSlots` first
   - If a slot has a parsed AST expression, call `exprToString()` directly — no re-parsing needed
   - Remove `parseAndGenerateInterpolatedExpr()` once all slots are pre-parsed

4. **Modify formatter** (`internal/formatter/formatter.go`):
   - Update string literal formatting to handle `ParsedSlots`

5. **Modify semantic analysis** (`internal/semantic/`):
   - Analyze expressions in `ParsedSlots` during `analyzeExpression`

6. **Add comprehensive tests:**
   - Simple interpolation: `"{name}"`
   - Pipe in interpolation: `"{data |> slice.First()}"`
   - Nested calls: `"{f(g(x))}"`
   - Method chains: `"{obj.Method().Field}"`
   - Mixed: `"prefix {a |> f()} middle {b} suffix"`

**Files:** `internal/ast/ast.go`, `internal/parser/parser_expr.go`, `internal/codegen/codegen_expr.go`, `internal/formatter/formatter.go`, `internal/semantic/`
**Risk:** High — touches AST definition, parser, codegen, formatter, and semantic analysis. Requires careful incremental approach and full regression testing at each step.

**Mitigation:** Keep `parseAndGenerateInterpolatedExpr()` as a fallback during transition. Only remove it once all interpolation forms are covered by the new pre-parsed path.

---

## Summary

| Phase | Items | Effort | Risk | Key Deliverable |
|-------|-------|--------|------|-----------------|
| 1 | 21, 20 | ~1 day | Low | Remove dead code; add formatter tests |
| 2 | 17, 19 | ~2 days | Medium | Proper IR nodes; cleaner lambda codegen |
| 3 | 18 | ~2-3 days | High | Pre-parsed interpolation slots |

Total estimated scope: ~5-6 days of focused work.
