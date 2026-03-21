# cmd/ — CLI Entry Points

The `cmd/` directory contains three standalone binaries and two code generators.

## Binaries

### `cmd/kukicha/` — Main CLI

The primary `kukicha` binary. Subcommands:

| Command | File | Description |
|---------|------|-------------|
| `build` | `main.go` | Transpile `.kuki` to `.go`, then `go build`. Flags: `--target`, `--skip-build`, `--if-changed`, `--vulncheck` |
| `run` | `main.go` | Transpile to a temp `.go` file and `go run` it. Passes extra args to the script |
| `check` | `main.go` | Parse + semantic analysis only (no codegen). Flag: `--strict-onerr` |
| `fmt` | `fmt.go` | Format `.kuki` files (tabs→spaces, trailing whitespace, brace conversion). Flags: `-w`, `--check` |
| `pack` | `pack.go` | Package a skill declaration into a directory with `SKILL.md` + compiled binary |
| `audit` | `audit.go` | Run `govulncheck` against project dependencies. Flags: `--json`, `--warn-only` |
| `init` | `init.go` | Initialize a Kukicha project (`go mod init`, extract stdlib, update AGENTS.md) |
| `version` | `main.go` | Print version from `internal/version/version.go` |

Key internal functions in `main.go`:

- **`compile()`** — Shared pipeline: resolve path → parse → analyze → detect target → codegen → gofmt. Returns `compileResult` used by `build`, `run`, and `pack`.
- **`loadAndAnalyze()`** — Parse + semantic analysis, returns AST + return counts + expr types.
- **`rewriteGoErrors()`** — Replaces generated `.go` file paths in Go compiler stderr with original `.kuki` paths.
- **`stripFirstLine()`** — Strips version header for `--if-changed` body comparison.

Key internal functions in `stdlib.go`:

- **`ensureStdlib()`** — Extracts embedded stdlib to `.kukicha/stdlib/`, version-stamped to avoid redundant extraction.
- **`ensureGoMod()`** — Adds `require` + `replace` directives for the stdlib module to the project's `go.mod`.
- **`needsStdlib()`** — Checks if generated Go code imports any `github.com/duber000/kukicha/stdlib/` packages (skips if inside the kukicha repo itself).
- **`extractAgentDocs()`** — Upserts Kukicha skill section into `AGENTS.md` and appends `@AGENTS.md` to `CLAUDE.md`.

`stdlib.go` also contains `stdlibGoMod` and `stdlibGoSum` constants — the `go.mod`/`go.sum` for the extracted stdlib module. Update these when adding or upgrading stdlib dependencies.

### `cmd/kukicha-lsp/` — Language Server

Thin entry point that creates an `lsp.NewServer` (from `internal/lsp`) and runs it over stdin/stdout. All logic lives in `internal/lsp/`.

## Code Generators

Both are invoked via `//go:generate` directives in `cmd/kukicha/main.go` and run automatically during `make build`.

### `cmd/genstdlibregistry/`

Scans `stdlib/*/*.kuki` files and generates `internal/semantic/stdlib_registry_gen.go`. Extracts:

- Return counts and per-position return types
- Parameter names and default values
- Func-typed parameter inner types (for lambda inference)
- `# kuki:deprecated`, `# kuki:security`, `# kuki:panics` directives
- Generic placeholder classification (`T`, `K`, `TK`, `TR`, `TO`, `O`)
- Exported interface declarations

Run standalone: `make genstdlibregistry` or `go run ./cmd/genstdlibregistry`

### `cmd/gengostdlib/`

Uses `go/importer` to extract Go stdlib function signatures and generates `internal/semantic/go_stdlib_gen.go`. Covers ~100 functions across `os`, `strconv`, `fmt`, `net`, `time`, `sync`, etc. Also extracts interface types and method signatures for exported types.

The curated function list is in the `packages` variable. Add new entries there when Kukicha needs to know about additional Go stdlib functions.

Run standalone: `make gengostdlib` or `go run ./cmd/gengostdlib`

## Testing

All packages use `package main` tests (white-box). Run with:

```bash
go test ./cmd/kukicha/... -v
go test ./cmd/genstdlibregistry/... -v
```

Test files and what they cover:

| File | Tests |
|------|-------|
| `kukicha/audit_test.go` | `findProjectRoot`, `runAudit` (no-go.mod case) |
| `kukicha/fmt_test.go` | `checkFile`, `formatFileInPlace`, `formatFileToStdout` |
| `kukicha/init_test.go` | `ensureStdlib` (extract, skip, re-extract), `ensureGoMod`, `upsertSkillSection`, `appendIfMissing`, `findProjectDir` |
| `kukicha/pack_test.go` | `generateSkillMD` YAML output, `defaultValueToYAML` |
| `kukicha/stdlib_test.go` | `needsStdlib` (no import, kukicha repo, user project) |
| `kukicha/rewrite_errors_test.go` | `rewriteGoErrors` (basic, multi, empty, no-match, nil) |
| `genstdlibregistry/main_test.go` | `scanRegistry` (exported, types, params, skips, deprecated), `formatRegistry`, `typeAnnotationToRepr` |

## Release Process

Version bumps touch `internal/version/version.go`. After bumping, all stdlib `.go` files need force-regeneration to update their version headers (see the release skill at `.claude/skills/release/SKILL.md` and `docs/contributing.md` for the full checklist).
