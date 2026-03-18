# Nim-Inspired Improvements — Implementation Plan

Concrete improvements to Kukicha informed by the [Nim language review](nim-language-review.md), scoped to items that directly improve real-world DevOps workflows. Each phase includes before/after examples drawn from [`examples/gh-semver-release`](../examples/gh-semver-release/).

---

## Phase 1: `stdlib/shell` Builder Improvements

**Complexity**: Low | **Impact**: Immediate readability win
**Status**: Done

Added `FlagIf`, `Preview`, and `Args` to `stdlib/shell`. All tests pass.

### API

```kukicha
# Conditional flag — appended only when condition is true
shell.FlagIf(cmd Command, condition bool, many args string) Command

# Display command as a string (for logging / dry-run output)
shell.Preview(cmd Command) string

# Append additional args to an existing command
shell.Args(cmd Command, many args string) Command
```

### Before (main.kuki:230-243)

```kukicha
cmdArgs := list of string{"release", "create", next, "--repo", repo, "--title", next}

if not tagAlreadyExists
    cmdArgs = append(cmdArgs, "--target", branch)
if generateNotes
    cmdArgs = append(cmdArgs, "--generate-notes")
if draft
    cmdArgs = append(cmdArgs, "--draft")

cmdStr := cmdArgs |> string.Join(" ")
print("Command:")
print("  gh {cmdStr}")
```

### After

```kukicha
cmd := shell.New("gh", "release", "create", next, "--repo", repo, "--title", next)
    |> shell.FlagIf(not tagAlreadyExists, "--target", branch)
    |> shell.FlagIf(generateNotes, "--generate-notes")
    |> shell.FlagIf(draft, "--draft")

print("Command:")
print("  {shell.Preview(cmd)}")
```

### Files changed

- `stdlib/shell/shell.kuki` — added `FlagIf`, `Preview`, `Args`
- `stdlib/shell/shell_test.kuki` — 4 new test functions (7 cases)
- Registry regenerated (736 → 739 entries)

---

## Phase 2: `stdlib/regex` — New Package

**Complexity**: Low | **Impact**: High — pattern matching without raw Go imports
**Status**: Done

Wraps Go's `regexp` with a beginner-friendly API. All functions annotated with `# kuki:security "regex"` for ReDoS detection.

### API

```kukicha
import "stdlib/regex"

# Core matching
regex.Match(pattern string, text string) bool
regex.Find(pattern string, text string) (string, error)
regex.FindAll(pattern string, text string) list of string
regex.FindGroups(pattern string, text string) (list of string, error)
regex.FindAllGroups(pattern string, text string) (list of list of string, error)

# Replacement
regex.Replace(pattern string, replacement string, text string) string
regex.ReplaceFunc(pattern string, replacer func(string) string, text string) string

# Splitting
regex.Split(pattern string, text string) list of string

# Validation
regex.IsValid(pattern string) bool

# Compiled (for hot paths)
regex.Compile(pattern string) (Pattern, error)
regex.MatchCompiled(p Pattern, text string) bool
regex.FindCompiled(p Pattern, text string) (string, error)
```

### Example improvement (semver prefix parsing)

**Before** (manual string ops):

```kukicha
if raw |> string.HasPrefix("v")
    prefix = "v"
    raw = raw |> string.TrimPrefix("v")
```

**After**:

```kukicha
match := regex.FindGroups(`^(v?)(\d+\.\d+\.\d+.*)$`, raw) onerr return empty, error "invalid version"
prefix = match[1]
raw = match[2]
```

Also includes compiled pattern variants for hot paths: `MustCompile`, `MatchCompiled`, `FindCompiled`, `FindAllCompiled`, `FindGroupsCompiled`, `ReplaceCompiled`, `SplitCompiled`.

### Files created

- `stdlib/regex/regex.kuki` — 17 functions + `Pattern` type
- `stdlib/regex/regex_test.kuki` — 15 test functions (30 cases)
- Registry regenerated (739 → 753 entries)

---

## Phase 3: `stdlib/git` — New Package

**Complexity**: Medium | **Impact**: High — gh-semver-release becomes dramatically simpler
**Status**: Done

Wraps the `gh` CLI to provide typed git/GitHub operations. Eliminates jq dependency, raw shell parsing, and manual string splitting. Uses `shell.FlagIf` from Phase 1 for clean conditional flag building in `CreateRelease`.

### API

```kukicha
import "stdlib/git"

# Tags
git.ListTags(repo string) (list of string, error)
git.TagExists(repo string, tag string) (bool, error)

# Branches
git.DefaultBranch(repo string) (string, error)
git.CurrentBranch() (string, error)

# Releases
git.ReleaseExists(repo string, tag string) (bool, error)
git.CreateRelease(repo string, tag string, opts ReleaseOptions) error
git.PreviewRelease(repo string, tag string, opts ReleaseOptions) string

type ReleaseOptions
    Title string
    Target string
    Draft bool
    GenerateNotes bool

# Repository info
git.RepoExists(repo string) (bool, error)
git.CurrentUser() (string, error)

# Local operations
git.Clone(url string, path string) error
git.CloneShallow(url string, path string, depth int) error
```

### Before (main.kuki:55-65 — raw shell + jq + manual line splitting)

```kukicha
func highestSemverTag(repo string) (string, error)
    raw := shell.Output("gh", "api", "repos/{repo}/tags",
        "--paginate", "--jq", ".[].name") onerr return "", error "failed to fetch tags for {repo}"

    if raw |> string.IsBlank()
        return "", empty

    tags := raw |> string.Lines()
    best, err := semver.Highest(tags)
    if err not equals empty
        return "", empty
    return best, empty
```

### After

```kukicha
func highestSemverTag(repo string) (string, error)
    tags := git.ListTags(repo) onerr return "", error "failed to fetch tags for {repo}"
    if tags |> slice.IsEmpty()
        return "", empty
    return semver.Highest(tags)
```

### Other simplifications in gh-semver-release

| Function | Before | After |
|----------|--------|-------|
| `defaultBranch` | 3 lines (shell.Output + jq + TrimSpace) | `return git.DefaultBranch(repo)` |
| `releaseExists` | 3 lines (shell.New + Execute + Success) | `return git.ReleaseExists(repo, tag)` |
| `tagExists` | 3 lines (shell.New + Execute + Success) | `return git.TagExists(repo, tag)` |
| `currentUser` | 3 lines (shell.Output + jq + TrimSpace) | `return git.CurrentUser()` |
| Release creation | 20+ lines (conditional append + shell.Output) | `git.CreateRelease(repo, tag, opts)` |

### Design notes

- Wraps `gh` (not raw git plumbing) since gh-semver-release and most DevOps scripts work with GitHub
- `PreviewRelease` enables dry-run output without executing
- `ReleaseOptions` fields are exported (capitalized) for cross-package access
- Network-dependent tests skip gracefully when `gh` is unavailable

### Files created

- `stdlib/git/git.kuki` — 11 functions + `ReleaseOptions` type
- `stdlib/git/git_test.kuki` — 9 test functions (14 cases; 5 network with skip guard)
- Registry regenerated (753 → 764 entries)

---

## Phase 4: Lambda Parameter Type Inference

**Complexity**: Medium-High | **Impact**: Highest for general Kukicha readability
**Status**: Done

Semantic analysis now infers lambda parameter types from call context, so codegen emits fully typed Go lambdas without requiring explicit annotations in `.kuki` source.

### Inference strategy (priority order)

1. **Generic type parameter instantiation**: When piped value is `list of string` and the function expects `func(T) bool`, resolve `T = string`
2. **Expected parameter type at call site**: When a function parameter is `func(cli.Args)`, the lambda param must be `cli.Args`
3. **Multi-return context**: Resolve return types for functions like `slice.Map`

### Before (gh-semver-release today)

```kukicha
# main.kuki:127 — type is redundant, repos is list of string
repos |> slice.Filter((r string) => not string.IsBlank(r))

# main.kuki:132 — same redundancy
repos = repos |> slice.Filter((r string) => r |> string.HasPrefix("{org}/"))

# main.kuki:159 — entries is list of RepoEntry, so e must be RepoEntry
entries = entries |> sort.ByKey((e RepoEntry) => e.name)

# main.kuki:323-325 — CommandAction expects func(cli.Args)
|> cli.CommandAction("list", (a cli.Args) => doList(a, initialTag))
|> cli.CommandAction("release", (a cli.Args) => doRelease(a, initialTag))
|> cli.CommandAction("pick", (a cli.Args) => doPick(a, initialTag))
```

### After (with inference)

```kukicha
repos |> slice.Filter(r => not string.IsBlank(r))

repos = repos |> slice.Filter(r => r |> string.HasPrefix("{org}/"))

entries = entries |> sort.ByKey(e => e.name)

|> cli.CommandAction("list", a => doList(a, initialTag))
|> cli.CommandAction("release", a => doRelease(a, initialTag))
|> cli.CommandAction("pick", a => doPick(a, initialTag))
```

### Files changed

- `internal/semantic/semantic_calls.go` — `inferLambdaParamTypes` and `inferLambdaParamTypesMethod` infer types from call context (Cases A/B/C)
- `cmd/genstdlibregistry/main.go` — populates `ParamFuncParams` on `goStdlibEntry` for func-typed parameters
- `internal/codegen/codegen_decl.go` — `generateArrowLambda` checks `exprTypes` for inferred types
- `internal/semantic/semantic_lambdas_test.go` — inference test cases
- `internal/codegen/codegen_lambdas_test.go` — codegen emission test cases

---

## Phase 5: `# kuki:panics` and `# kuki:todo` Directives

**Complexity**: Low | **Impact**: Moderate
**Status**: Done

Both extend the existing directive infrastructure (`# kuki:deprecated`, `# kuki:security`).

### `# kuki:panics`

Emits a compiler warning at each call site of an annotated function. Go has no mechanism to warn callers about functions that panic — this fills that gap.

```kukicha
# kuki:panics "when input is empty"
func MustParse(s string) Config
    if s equals ""
        panic "empty input"
    return parse(s)
```

Callers see:

```
warning: MustParse may panic: "when input is empty" (main.kuki:42)
```

### `# kuki:todo`

Emits a compile-time warning for any annotated declaration. Useful for AI-generated code that flags incomplete sections, and in CI to catch code left unfinished.

```kukicha
# kuki:todo "Add retry logic"
func fetchConfig(url string) (Config, error)
    return fetch.Get(url) |> fetch.Json(Config)
```

Output:

```
warning: TODO: "Add retry logic" on fetchConfig (config.kuki:15)
```

### Design decisions

- Warning by default; `--strict-panics` / `--strict-todos` flags could promote to error for CI
- Both work on `func` and `type` declarations (same as existing directives)
- `# kuki:panics` message should appear in LSP hover tooltips

### Files changed

- `internal/semantic/semantic.go` — `collectDeprecations` registers `panics` into `panickedFuncs` map and emits `todo` warnings on declarations
- `internal/semantic/semantic_calls.go` — `checkPanics` emits warnings at call sites (local + stdlib via `generatedStdlibPanics`)
- `cmd/genstdlibregistry/main.go` — generates `generatedStdlibPanics` map from `# kuki:panics` directives
- `internal/semantic/directive_test.go` — tests for `todo` (func + type) and `panics` warnings
- `stdlib/input/input.kuki` — first stdlib usage (`# kuki:panics "if reading from stdin fails"` on `Prompt`)

---

## Execution Order Summary

| Order | Phase | Item | Status |
|-------|-------|------|--------|
| 1 | Phase 1 | Shell builder (FlagIf, Preview, Args) | Done |
| 2 | Phase 2 | `stdlib/regex` | Done |
| 3 | Phase 3 | `stdlib/git` | Done |
| 4 | Phase 4 | Lambda type inference | Done |
| 5 | Phase 5 | `# kuki:panics` + `# kuki:todo` | Done |

### Rewritten gh-semver-release (target state after all phases)

After all five phases, `examples/gh-semver-release/main.kuki` should:

- Drop ~40% of its line count
- Eliminate all raw `shell.Output("gh", ...)` calls in favor of `stdlib/git`
- Replace conditional `append` boilerplate with `shell.FlagIf` chains
- Use untyped lambdas everywhere (`r =>` instead of `(r string) =>`)
- Remove dependency on external `jq` for JSON field extraction
