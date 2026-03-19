# Plan: Fix Agreed Issues from Minimax Review

## Issue 1: Add `# kuki:security "files"` to `files.Copy` and `files.Move`

**Files:** `stdlib/files/files.kuki`
**Steps:**
1. Add `# kuki:security "files"` before the `func Copy(src string, dst string) error` declaration (before line 144)
2. Add `# kuki:security "files"` before the `func Move(src string, dst string) error` declaration (before line 157)
3. Run `make generate` to regenerate `stdlib_registry_gen.go` with the new security entries
4. Verify `files.Copy` and `files.Move` appear in `generatedSecurityFunctions` in the regenerated file
5. Run `make test` to confirm nothing breaks

---

## Issue 8: Scope printf detection to known packages

**Files:** `internal/codegen/codegen_expr.go`
**Approach:** Instead of checking just the method name, also check the receiver/object to ensure it belongs to a known printf-style package or type. The current call site in `generateMethodCallExpr` has access to `expr.Object` (the receiver). We can qualify the check by combining the object's resolved package with the method name.

**Steps:**
1. Change `printfMethods` from `map[string]bool` to a set of qualified `package.Method` or `type.Method` patterns (e.g., `fmt.Printf`, `testing.T.Errorf`, `log.Printf`)
2. At the call site, resolve the object's package/type from `expr.Object` and the generator's alias/import context
3. Only apply printf-style formatting when the qualified name matches a known printf function
4. Add a test case with a user-defined `Errorf` method to confirm it's NOT treated as printf-style
5. Run `make test`

**Note:** This is the most complex fix. If resolving receiver types proves too involved in the current codegen architecture, a simpler alternative is to limit printf detection to known package-level functions (e.g., `fmt.Sprintf`, `log.Printf`) and only apply method detection for `testing.T`/`testing.B` receivers. Evaluate feasibility before implementing.

---

## Issue 10: Eliminate double `genstdlibregistry` run in `make generate`

**Files:** `Makefile`
**Problem:** `generate` depends on `genstdlibregistry` explicitly, then `build` runs `go generate ./...` which re-runs it via `//go:generate` in `main.go`.

**Steps:**
1. Remove `genstdlibregistry` from the `generate` target's dependency list, since `build` (via `go generate ./...`) already runs it
2. Change `generate: genstdlibregistry build generate-tests` → `generate: build generate-tests`
3. Run `make generate` and verify it still produces correct output
4. Run `make test`

---

## Issue 12: Add main `.kuki` → `.go` staleness check

**Files:** `Makefile`
**Steps:**
1. Add a `check-main-staleness` target modeled on `check-test-staleness`
2. Iterate over `$(KUKI_MAIN)`, check each `.kuki` has a corresponding `.go` file that is not older
3. Wire `check-main-staleness` into the `test` target (alongside `check-test-staleness`)
4. Run `make test` to verify it passes in a clean state
5. Touch a `.kuki` file without regenerating to verify it catches staleness

---

## Issue 13: Fix misleading comment on `json.Encode`

**Files:** `stdlib/json/json.kuki`
**Problem:** Comment says "indent/prefix options not yet supported" but `WithIndent`/`WithPrefix` exist. However, `Encode` doesn't actually use the stored `indent`/`prefix` fields — it just calls the standard encoder.

**Steps:**
1. Wire the `indent` and `prefix` fields into `Encode` — call `encoder.SetIndent(enc.prefix, enc.indent)` when either is non-empty
2. Remove the "not yet supported" comment
3. Run `make generate` to regenerate the `.go` file
4. Add a test case that verifies `WithIndent` actually produces indented output
5. Run `make test`

---

## Issue 14: Fix `slice.First`/`Last` to preserve element type

**Files:** `stdlib/slice/slice.kuki`
**Problem:** `First` and `Last` return `list of any` but should use the generic `any` placeholder like `Filter` and `Map` do, so the element type is preserved.

**Steps:**
1. The signatures already use `list of any` which IS the generic placeholder — but the `make(list of any, 0)` calls inside the function may be causing the issue. Verify by checking the generated Go code to see if `First`/`Last` get `[T any]` type parameters like `Filter` does
2. If the generated code is correct (already generic), this issue is a false positive in the review — add a note
3. If the generated code is NOT generic, investigate why the codegen treats these differently from `Filter` (likely because `Filter` uses `func(any)` in its signature which triggers the generic detection, while `First`/`Last` only use `list of any`)
4. Fix by ensuring the codegen recognizes `list of any` parameter + `list of any` return as requiring a type parameter
5. Run `make generate` and verify the output
6. Run `make test`

---

## Execution Order

1. **Issue 1** (security directive) — highest priority, simplest fix
2. **Issue 13** (json.Encode) — small, self-contained
3. **Issue 10** (Makefile double-run) — small, low risk
4. **Issue 12** (staleness check) — small, additive
5. **Issue 14** (slice.First/Last) — needs investigation first
6. **Issue 8** (printf detection) — most complex, may need scoping down
