# Transpiler TODO

Bugs discovered while building `stdlib/semver`, upgrading `stdlib/cli` with subcommands, and rewriting `examples/gh-semver-release`. Each entry includes the workaround applied and where to clean up after the fix.

---

## ~~1. Pipes inside string interpolation produce invalid Go~~ ✅ FIXED

Fixed in commit `6edc7e4`. Added `parseAndGenerateInterpolatedExpr()` in `codegen_expr.go` — re-parses pipe expressions through the parser to produce valid Go.

Cleanups applied: `stdlib/cli/cli.kuki` `padded` variable inlined into interpolation.

---

## ~~2. Default parameters not applied in stdlib-to-stdlib calls~~ ✅ FIXED

Fixed in commit `6edc7e4`. Added `DefaultValues` field to `goStdlibEntry`, populated by `cmd/genstdlibregistry`. Codegen fills missing trailing args via `fillStdlibDefaults()` in both `generateMethodCallExpr()` and `generatePipeExpr()`.

Cleanups applied: removed explicit `" "` from `cli.kuki` and `examples/gh-semver-release/main.kuki`.

---

## ~~3. `onerr continue` / `onerr break` do not parse~~ ✅ FIXED

Fixed in commit `6edc7e4`. Added `ShorthandContinue`/`ShorthandBreak` to `OnErrClause`, parser support in `parseInlineOnErrHandler`, and IR lowering.

Cleanups applied: `stdlib/semver/semver.kuki` `Highest()`, `stdlib/container/container.kuki` `buildImage()`.
