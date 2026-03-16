# Technical Debt & Improvement Plan

Comprehensive audit of shortcuts, heuristics, and gaps in the Kukicha compiler.
Items 1–4 from Tier 1 were fixed in commit `a41558c`.
Dead code and lint violations cleaned up in commits `5fbf182` and `6d04e16`;
`golangci-lint` added to prevent future accumulation (`make lint`).
Tier 3 items 8–12 fixed by replacing hardcoded lists with data-driven approaches
via `genstdlibregistry`, `gengostdlib`, and `# kuki:security`/`# kuki:deprecated` directives.

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

## Tier 3: Heuristics That Should Be Proper Logic ✅ DONE

### ~~8. Type assertion vs. conversion heuristic~~ ✅ FIXED
**File:** `codegen_expr.go`

Now uses `isLikelyInterfaceType()` instead of string-matching on dots. Correctly handles local interfaces, known Go interfaces, and `error`.

### ~~9. `isLikelyInterfaceType` hardcoded list~~ ✅ FIXED
**File:** `codegen_stdlib.go`

Deleted the hardcoded `knownInterfaces` map. `isLikelyInterfaceType` now checks: (1) `"error"`, (2) local interface declarations, (3) auto-generated `generatedGoInterfaces` map (52 interfaces extracted from Go stdlib via `go/types` in `gengostdlib`), and (4) auto-generated `generatedStdlibInterfaces` map (from `genstdlibregistry` scanning `InterfaceDecl` nodes in `.kuki` files).

### ~~10. Hardcoded security function lists~~ ✅ FIXED
**File:** `semantic_security.go`

Security-sensitive functions are now annotated with `# kuki:security "category"` directives in their `.kuki` source files. The `genstdlibregistry` generator scans these directives and emits a `generatedSecurityFunctions` map. Security checks use `securityCategory()` which reads from this generated map (with alias support for `httphelper.X → http.X`).

### ~~11. Hardcoded generic/comparable function lists~~ ✅ FIXED
**File:** `codegen_stdlib.go`

`genericSafe` and `comparableSafe` maps deleted. `inferSliceTypeParameters` now reads from the generated `generatedSliceGenericClass` map (via `semantic.GetSliceGenericClass()`), which is auto-derived from placeholder usage in `.kuki` function signatures.

### ~~12. Hardcoded "Enumerate" special case~~ ✅ FIXED
**File:** `codegen_decl.go`

Introduced `iter.Seq2Int` type name convention instead of checking function name. `stdlib/iterator/iterator.kuki` updated to use the new return type.

---

## Tier 4: Testing Gaps ✅ DONE

### ~~13. No tests for core codegen functions~~ ✅ FIXED

Added test coverage for all five previously-untested files:

| File | Test file | Tests added |
|------|-----------|-------------|
| `codegen_imports.go` | `codegen_imports_test.go` | `extractPkgName`, `rewriteStdlibImport`, collision aliasing, builtin aliasing, version suffix aliasing, import format, auto-imports |
| `codegen_types.go` | `codegen_types_test.go` | `typeInfoToGoString` for all `TypeKind` variants, package alias rewriting |
| `codegen_walk.go` | `codegen_walk_test.go` | `needsPrintBuiltin`, `needsErrorsPackage`, `needsStringInterpolation`, `collectReservedNames`, `walkProgram` short-circuit |
| `codegen_decl.go` | `codegen_decl_test.go` | `generateInterfaceDecl`, `generateGlobalVarDecl`, method/pointer receiver, variadic, `generateTypeAnnotation`, `generateReturnTypes`, type alias, JSON tags |
| `codegen_stdlib.go` | `codegen_stdlib_test.go` | `zeroValueForType`, `isLikelyInterfaceType`, `inferExprReturnType`, `typeContainsPlaceholder`, `returnCountForFunctionName` |

### ~~14. All codegen tests use `strings.Contains`~~ ✅ MITIGATED

Existing unit tests still use `strings.Contains`, but the risk of false positives is now mitigated by 25 integration tests (`codegen_integration_test.go`) that run the full pipeline (lex → parse → semantic → codegen) and verify the generated Go is syntactically valid using `go/parser.ParseFile`. This catches structural issues that substring checks miss.

### ~~15. Sparse error case tests~~ ✅ IMPROVED

Added tests for most of the identified gaps:
- **Deeply nested indentation (10+ levels):** `TestDeeplyNestedIndentation` — verifies correct tab depth
- **Parser error cascading:** `TestParserCascadesMultipleErrors` — verifies parser reports errors for malformed input
- **Import collision scenarios:** `TestImportCollisionAutoAlias`, `TestImportBuiltinTypeAlias` in `codegen_imports_test.go`
- **onerr continue/break in loops:** `TestOnErrContinueInLoop`, `TestOnErrBreakInLoop`
- **onerr block (multi-statement):** `TestOnErrBlockMultiStatement`

Remaining gap: circular type definitions (rare edge case, deferred).

### ~~16. Zero integration tests in `internal/`~~ ✅ FIXED

Added `codegen_integration_test.go` with 25 integration tests that run the full pipeline (lex → parse → semantic → codegen) and verify the generated Go parses as valid Go syntax. Covers: functions, types, methods, string interpolation, error handling, `onerr` (return/default/panic), loops (range, numeric, through), switch, lists/maps, interfaces, global vars, variadics, channels, default params, type aliases, nested control flow, multiple returns, negative indexing, arrow lambdas, JSON tags, defer.

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
