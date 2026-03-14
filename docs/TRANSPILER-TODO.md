# Transpiler TODO

Bugs discovered while building `stdlib/semver`, upgrading `stdlib/cli` with subcommands, and rewriting `examples/gh-semver-release`. Each entry includes the workaround applied and where to clean up after the fix.

---

## 1. Pipes inside string interpolation produce invalid Go

**Severity:** High — pipes and string interpolation are both core features; their combo is natural and common.

**Reproduction:**
```kukicha
print("  {cmd.name |> string.PadRight(14)}{cmd.description}")
```

**Expected:** Generates `kukistring.PadRight(cmd.name, 14)` nested inside `fmt.Sprintf`.

**Actual:** Generates Go with a literal `|>` operator:
```go
fmt.Println(fmt.Sprintf("  %v%v", cmd.name |> string.PadRight(14), cmd.description))
```
This fails to compile: `syntax error: unexpected >, expected expression`.

**Workaround:** Extract the pipe result into a local variable before interpolation:
```kukicha
padded := string.PadRight(cmd.name, 14, " ")
print("  {padded}{cmd.description}")
```

**Cleanup locations after fix:**
- `stdlib/cli/cli.kuki` — `printHelp()` and `printCommandHelp()` functions: inline the `padded` / `label` variables back into the interpolation strings

---

## 2. Default parameters not applied in stdlib-to-stdlib calls

**Severity:** Medium — only affects stdlib authors, not end users.

**Reproduction:**
```kukicha
# In stdlib/cli/cli.kuki, calling stdlib/string:
padded := string.PadRight(cmd.name, 14)   # omitting 3rd arg (default " ")
```

**Expected:** Transpiler inserts the default value, generating `kukistring.PadRight(cmd.name, 14, " ")`.

**Actual:** Generates `kukistring.PadRight(cmd.name, 14)` — the default is not inserted. Go compilation fails:
```
not enough arguments in call to kukistring.PadRight
    have (string, number)
    want (string, int, string)
```

**Root cause:** The `make generate` pipeline transpiles each `.kuki` file independently. Default parameter insertion happens in the main compiler's semantic pass, but the stdlib transpiler doesn't run that pass (or doesn't have access to cross-package default metadata).

**Workaround:** Always pass all arguments explicitly when calling between stdlib packages:
```kukicha
padded := string.PadRight(cmd.name, 14, " ")   # explicit 3rd arg
```

**Cleanup locations after fix:**
- `stdlib/cli/cli.kuki` — all 6 calls to `string.PadRight`: remove the explicit `" "` third argument
- `examples/gh-semver-release/main.kuki` — `doPick` function: 2 calls to `string.PadRight` with explicit `" "`

---

## 3. `onerr continue` does not parse

**Severity:** Medium — `onerr` supports `return`, `panic`, default values, and block forms, but not `continue`.

**Reproduction:**
```kukicha
for i from 1 to len(valid)
    v := Parse(valid[i]) onerr continue
```

**Expected:** Generates `if err != nil { continue }`.

**Actual:** Parse error: `unexpected token in expression: CONTINUE`.

**Workaround:** Manual error check:
```kukicha
for i from 1 to len(valid)
    v, err := Parse(valid[i])
    if err != empty
        continue
```

**Cleanup locations after fix:**
- `stdlib/semver/semver.kuki` — `Highest()` function: replace the 3-line manual check with `v := Parse(valid[i]) onerr continue`

---

## Notes

`onerr break` likely has the same issue as `onerr continue` — worth testing if the parser fix covers both loop-control keywords.
