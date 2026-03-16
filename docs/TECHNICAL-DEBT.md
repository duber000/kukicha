# Technical Debt & Improvement Plan

Comprehensive audit of shortcuts, heuristics, and gaps in the Kukicha compiler.
Items 1–4 from Tier 1 were fixed in commit `a41558c`.
Dead code and lint violations cleaned up in commits `5fbf182` and `6d04e16`;
`golangci-lint` added to prevent future accumulation (`make lint`).

---

## Tier 1: Silent Wrong Code Generation ✅ DONE

### ~~1. `exprToString` returns `""` for unknown expressions~~ ✅ FIXED
**File:** `codegen_expr.go:190`
Now panics with expression type and source location.

### ~~2. Pipe fallback emits literal `|>`~~ ✅ FIXED
**File:** `codegen_expr.go:712`
Now panics with pipe target type and source location.

### ~~3. `onerr error "msg"` assumes 2 return values~~ ✅ FIXED
**File:** `codegen_onerr.go:87-96`
Now uses `currentReturnTypes` + `zeroValueForType` for 3+ return functions.

### ~~4. Return count fallback discards values~~ ✅ FIXED
**Files:** `codegen_onerr.go:166-169`, `lower.go:437-442`
Now emits a `// kukicha:` comment when inference fails so the problem is visible.

---

## Tier 2: Crashes and Security Gaps ✅ DONE

### ~~5. Parser nil returns cause codegen panics~~ ✅ FIXED
**Files:** `parser_expr.go:486-496`, `parser_type.go:148-151`, `parser_expr.go:479-483`

`parseIdentifier()`, `parseTypeAnnotation()`, and `parsePrimaryExpr()` now return sentinel values instead of nil on error. The error is still recorded; codegen won't run on programs with parse errors. `parseIntegerLiteral()` and `parseFloatLiteral()` also return sentinels now.

### ~~6. Security checks skip piped values~~ ✅ FIXED
**File:** `semantic_security.go:162-170`

`checkShellRunNonLiteral` now emits a warning (not error) when a value is piped into `shell.Run()`. The other security checks (SQL, HTML, redirect) already handle piped args correctly by adjusting the argument index — the piped value is the connection/writer, not the dangerous input.

### ~~7. `peekAt()` doesn't skip comments~~ ✅ FIXED
**File:** `parser.go:192-212`

`peekAt()` now counts only meaningful tokens (skipping comments, semicolons, directives) when computing the offset, matching the behavior of `peekToken()` and `peekNextToken()`.

---

## Tier 3: Heuristics That Should Be Proper Logic

### 8. Type assertion vs. conversion heuristic
**File:** `codegen_expr.go:115-121`

Decides between `x.(Type)` (assertion) and `Type(x)` (conversion) by checking if the type name contains a dot — with a special exception for `iter.Seq`:

```go
if strings.Contains(targetType, ".") && !strings.Contains(targetType, "iter.Seq") {
    return fmt.Sprintf("%s.(%s)", expr, targetType)
}
return fmt.Sprintf("%s(%s)", targetType, expr)
```

This is wrong for:
- Local interface types (no dot → uses conversion instead of assertion)
- Qualified non-interface types (has dot → uses assertion instead of conversion)

**Fix:** Use `isLikelyInterfaceType()` (which already exists) instead of string matching on dots.

### 9. `isLikelyInterfaceType` hardcoded list
**File:** `codegen_stdlib.go:404-422`

The function name says it all — "likely." It checks:
1. The literal string `"error"`
2. Interface declarations in the current file
3. A hardcoded `knownInterfaces` map (lines 52-74)

User-defined interfaces from other packages are missed entirely.

**Fix:** Extend to resolve interfaces from imported packages, or add an `isInterface` flag to the type info passed from semantic analysis.

### 10. Hardcoded security function lists
**File:** `semantic_security.go:11-49`

Security-sensitive functions are tracked in manually-maintained maps:
- `sqlFunctions`
- `htmlFunctions`
- `fetchFunctions`
- `filesFunctions`
- `redirectFunctions`

Adding a new stdlib function that is security-sensitive requires updating these maps manually.

**Fix:** Move to registry annotations (e.g., `# kuki:security sql` directives on stdlib function declarations) so the security check table is maintained alongside the code it protects.

### 11. Hardcoded generic/comparable function lists
**File:** `codegen_stdlib.go:19-50`

`genericSafe` and `comparableSafe` maps must be updated manually when new stdlib functions are added. Same for `knownInterfaces` (lines 52-74).

**Fix:** Derive these from the stdlib `.kuki` source files during `make genstdlibregistry`. Functions using `any`/`any2` placeholders can be auto-classified.

### 12. Hardcoded "Enumerate" special case
**File:** `codegen_decl.go:466`

```go
if g.currentFuncName == "Enumerate" {
    return "iter.Seq2[int, T]"
}
```

**Fix:** Make this data-driven via a return-type annotation in the stdlib registry rather than a name check.

---

## Tier 4: Testing Gaps

### 13. No tests for core codegen functions

The following files have **zero test coverage**:

| File | Exported functions |
|------|-------------------|
| `codegen_imports.go` | `generateImports`, `rewriteStdlibImport`, + 6 more |
| `codegen_types.go` | `typeInfoToGoString` |
| `codegen_walk.go` | 13+ `needs*` helper predicates |
| `codegen_decl.go` | `generateTypeDecl`, `generateInterfaceDecl`, `generateGlobalVarDecl`, `generateFunctionDecl` |
| `codegen_stdlib.go` | `inferStdlibTypeParameters`, `zeroValueForType` |

### 14. All codegen tests use `strings.Contains`

Tests pass if a substring appears anywhere in the output. This means:
- False positives: unrelated code changes could accidentally satisfy assertions
- No structural verification of the generated Go AST
- Fragile when output formatting changes

**Fix (pragmatic):** At minimum, add `strings.Count` checks to verify substrings appear exactly once. Better: add snapshot tests or compile the generated Go to verify it's valid.

### 15. Sparse error case tests

~14 error-case test functions across the entire compiler vs. hundreds of happy-path tests. Key gaps:
- Complex nested pipes with `onerr` at multiple levels
- Malformed input recovery (does the parser cascade errors?)
- Circular type definitions
- Import collision scenarios
- Deeply nested indentation (10+ levels)

### 16. Zero integration tests in `internal/`

No end-to-end tests that run the full pipeline (lex → parse → semantic → codegen → `go build`). The `examples/` directory also has no tests.

**Fix:** Add integration tests that compile example `.kuki` programs to Go and verify they build successfully with `go build`.

---

## Tier 5: Architecture Improvements (Lower Priority)

### 17. RawStmt escape hatch undermines IR
**File:** `lower.go:106-137`

Many codegen paths bypass proper IR nodes and emit raw Go strings via `ir.RawStmt`. This defeats the purpose of the IR layer.

**Status:** Acceptable for now. The IR was introduced incrementally and covers the most complex paths (pipe chains, onerr). Expanding IR coverage is a gradual effort.

### 18. String re-parsing for interpolated pipes
**File:** `codegen_expr.go:519-546`

`parseAndGenerateInterpolatedExpr()` creates a fake function wrapper, re-parses it, extracts the AST, and re-generates. This is a full parser round-trip at codegen time.

**Fix:** Store pipe expressions as AST nodes in `StringLiteral` interpolation slots during parsing, rather than as raw strings that need re-parsing.

### 19. Temporary generators for lambda codegen
**Files:** `codegen_decl.go:243-275, 321-350`

Creates throwaway `Generator` instances to capture output for function literals and arrow lambdas, rather than composing IR nodes.

**Status:** Works but wasteful. Would benefit from the IR layer being extended to cover lambda bodies.

### 20. Formatter re-parses from scratch
**File:** `internal/formatter/`

The formatter doesn't reuse parse results — it re-parses the source independently. Comment handling was bolted on via `ExtractComments()`/`AttachComments()` with zero test coverage.

**Fix:** Share parse results between compiler and formatter, or at minimum add tests for the comment handling.

### 21. Error message rewriting is fragile
**File:** `cmd/kukicha/main.go:194-199`

`rewriteGoErrors()` does post-hoc string replacement of `.go` paths with `.kuki` paths. If Go's error message format changes, this breaks.

**Fix:** Use proper source maps (the `//line` directives are already emitted — Go's errors should reference `.kuki` files). Investigate whether this rewriting is still needed.

---

## Reference: Deferred-to-Go Validation

These are intentional design decisions, not bugs. Documenting for awareness:

| What | Location | Consequence |
|------|----------|-------------|
| External type validation | `semantic_types.go:44-46` | `io.Reader` accepted without verification |
| Interface satisfaction | `semantic_types.go:215-219` | Always returns `true` |
| Collection element types | `semantic_types.go:227-231` | `list of int` vs `list of string` not caught |
| Import resolution | `codegen_imports.go:28` | No check that packages exist |
| Untyped lambda params | `codegen_decl.go:288-291` | Emitted as bare identifiers |
| Named args for external funcs | `semantic_calls.go:106` | Only works for Kukicha stdlib |

These produce Go compiler errors rather than Kukicha compiler errors, which means worse error messages for users. Improving these requires building more of a type system, which is a larger effort.
