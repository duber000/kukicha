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

## Tier 5: Architecture Improvements (4 of 5 done)

### ~~17. RawStmt escape hatch undermines IR~~ ✅ IMPROVED
**File:** `lower.go`

Added `ir.ReturnStmt`, `ir.ExprStmt`, and `ir.Comment` IR node types. Replaced `RawStmt` usage for `continue`, `break`, shorthand return, explain wrapping, and inference comments with proper IR nodes. Remaining `RawStmt` usage: rendered handler blocks (requires full handler lowering) and rendered switch statements (complex, deferred).

### 18. String re-parsing for interpolated pipes
**File:** `codegen_expr.go:519-546`

`parseAndGenerateInterpolatedExpr()` creates a fake function wrapper, re-parses it, extracts the AST, and re-generates. This is a full parser round-trip at codegen time.

**Fix:** Store pipe expressions as AST nodes in `StringLiteral` interpolation slots during parsing, rather than as raw strings that need re-parsing.

### ~~19. Temporary generators for lambda codegen~~ ✅ FIXED
**Files:** `codegen_decl.go`, `codegen.go`

Introduced `childGenerator(extraIndent)` method that creates a child generator sharing the parent's semantic state (program, auto-imports, aliases, type info, expression types) but writing to a fresh output buffer. Replaces manual 8-field copies in `generateFunctionLiteral()` and `generateArrowLambda()`, eliminating the risk of missing fields when new state is added to `Generator`.

### ~~20. Formatter comment handling has test coverage~~ ✅ IMPROVED
**File:** `internal/formatter/`

Added `comments_test.go` with comprehensive tests: `ExtractComments` (no comments, single, multiple, trailing, directive exclusion) and `AttachComments` (empty, leading on function, multiple leading, trailing on statement, inside if/for/switch blocks, between functions, on imports, on type fields). Also added formatter round-trip integration tests verifying comments are preserved through format cycles.

The formatter still re-parses independently (not sharing parse results with the compiler). This is a larger refactor deferred for now — the tests ensure correctness.

### ~~21. Error message rewriting is documented and tested~~ ✅ IMPROVED
**File:** `cmd/kukicha/main.go`

Investigation confirmed `rewriteGoErrors()` is still needed: `//line` directives handle most source mapping, but Go compiler errors for package-level issues, import failures, and syntax errors in generated code reference the physical file path directly. The function now has an early return for empty input and thorough documentation explaining why the broad `strings.ReplaceAll` approach is safe (the goFile is a unique temp/output path). Added `rewrite_errors_test.go` with tests for basic rewriting, multiple occurrences, empty input, no-match, and nil input.

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
