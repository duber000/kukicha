# stdlib test TODO

Status of every `*_test.kuki` file and what needs doing.
Run `kukicha check stdlib/<pkg>/<pkg>_test.kuki` to verify after each fix.

Reference pattern: `stdlib/math/math_test.kuki`, `stdlib/slice/slice_test.kuki`.
See **stdlib/CLAUDE.md § Testing Stdlib Packages** for the full convention.

---

## Already done ✓

| Package | Status |
|---------|--------|
| `datetime` | table-driven, passes |
| `math`     | table-driven, passes |
| `slice`    | table-driven, passes |
| `string`   | table-driven, passes |
| `a2a` | fixed, passes |
| `cast` | fixed, passes |
| `ctx` | fixed, passes |
| `encoding` | fixed, passes |
| `env` | fixed, passes |
| `http` | fixed, passes |
| `iterator` | fixed, passes |
| `maps` | fixed, passes |
| `mcp` | fixed, passes |
| `obs` | fixed, passes |
| `parse` | fixed, passes |
| `validate` | fixed, passes |
| `container` | fixed, passes |
| `concurrent` | fixed, passes |
| `errors` | fixed, passes |
| `net` | fixed, passes |
| `retry` | fixed, passes |
| `input` | fixed, passes |
| `llm` | fixed, passes |
| `must` | fixed, passes |
| `template` | fixed, passes |
| `cli` | fixed, passes |
| `fetch` | fixed, passes |
| `files` | fixed, passes |
| `json` | fixed, passes |
| `kube` | fixed, passes |
| `netguard` | fixed, passes |
| `pg` | fixed, passes |
| `random` | fixed, passes |
| `sandbox` | fixed, passes |
| `shell` | fixed, passes |

**Reference implementations:** datetime, math, slice, string. When in doubt, look at one of them.
**Recently fixed:** All other packages in this list have been systematically refactored per the patterns below.

---

## Broken syntax — full rewrite needed

**Status:** ✅ All fixed! (input, llm, must, template now pass)

*Previously, these files had parse errors and could not compile. All have been refactored.*

### Common issues found across these files

**`test <name>` block syntax** (original issue — now fixed)
The files use an unimplemented `test Name` block form (originally written for a future test-DSL feature). Replace every `test FooBar` block with a standard test function:
```kukicha
# WRONG — unimplemented syntax
test BackgroundFunction
    h := ctx.Background()
    test.AssertNotEmpty(h)

# RIGHT
func TestBackground(t reference testing.T)
    h := ctx.Background()
    test.AssertNotEmpty(t, h)
```
Also add a missing `petiole <pkg>_test` declaration at the top if absent.

**Multi-return in `if` condition** (original issue — now fixed in cast, validate, net)
`if val, err := f(); condition` is not valid Kukicha — split onto two lines:
```kukicha
# WRONG
if val, err := cast.SmartInt(5); err != empty or val not equals 5
    t.Fatalf("failed")

# RIGHT
val, err := cast.SmartInt(5)
if err != empty or val != 5
    t.Fatalf("SmartInt failed: {err}")
```

**`defer func() { ... }()` closures** (input, must, template)
Anonymous deferred closures for panic recovery don't parse. Use a named helper or
restructure the test to avoid recover-based panic testing entirely:
```kukicha
# WRONG — anonymous closure in defer
defer func()
    if recover() equals empty
        t.Error("Expected panic")
    return
}()

# RIGHT — use t.Run isolation or skip panic testing
# (functions that are supposed to panic are hard to unit-test; skip or document)
```

**`for _, item in list`** (original issue — now fixed in maps)
Kukicha for-in doesn't expose an index variable. Drop the `_,`:
```kukicha
# WRONG
for _, item in list

# RIGHT
for item in list
```

**Struct literals inside a list** (llm, mcp)
If the compiler rejects `list of Foo{Foo{field: val}, ...}`, use the element type
explicitly: `list of FooCase{FooCase{...}, FooCase{...}}`.

**Private field access** (original issue — now fixed in ctx, obs, a2a)
Tests must only use public API — remove assertions on unexported fields and test
the observable behaviour instead (e.g. call the function and check its return value).

---

### Per-package notes — all previously broken packages are now fixed

*(All packages in "Broken syntax" section have been successfully refactored and now pass tests)*

---

## Semantic errors — small targeted fixes

These parse correctly but fail type checking.

*(All packages in this section have been fixed and moved to "Already done ✓" above)*

---

## Go compile errors — logic/API mismatches

These pass `kukicha check` but fail `go build` or `go test`.

*(All packages in this section have been fixed and moved to "Already done ✓" above)*

---

## Test logic bugs — compile OK but assertions fail

*(All packages in this section have been fixed and moved to "Already done ✓" above)*

---

---

## Quick-start checklist for fixing a broken file

1. Add `petiole <pkg>_test` at the top if missing
2. Add `import "stdlib/test"` and `import "testing"` if missing
3. Replace every `test FooBar` block with `func TestFooBar(t reference testing.T)`
4. Split multi-return-in-`if` into two statements
5. Remove assertions on private struct fields — test observable output only
6. Add a case type and `for tc in cases` / `t.Run(tc.name, ...)` loop
7. Replace bare `t.Errorf(...)` checks with `test.AssertEqual(t, got, want)`
8. Run `kukicha check stdlib/<pkg>/<pkg>_test.kuki`
9. Run `./kukicha build stdlib/<pkg>/<pkg>_test.kuki` (generates `_test.go`)
10. Run `go test ./stdlib/<pkg>/...`
