# Nim-Inspired Improvements Plan

Concrete improvements to Kukicha informed by the [Nim language review](nim-language-review.md), prioritized by impact on real-world DevOps workflows. Each item includes before/after examples drawn from [`examples/gh-semver-release`](../examples/gh-semver-release/).

---

## Language Features

### 1. Lambda Parameter Type Inference — Tier 1

**Status**: Not started
**Complexity**: Medium-high
**Impact**: Highest — reduces the most common source of verbosity in pipe chains

The compiler already knows the expected callback signature from generic function definitions. When a lambda is passed to a function whose parameter type is `func(T) U`, the compiler should infer `T` for untyped lambda parameters.

**Before** (gh-semver-release today):

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

**After** (with inference):

```kukicha
repos |> slice.Filter(r => not string.IsBlank(r))

repos = repos |> slice.Filter(r => r |> string.HasPrefix("{org}/"))

entries = entries |> sort.ByKey(e => e.name)

|> cli.CommandAction("list", a => doList(a, initialTag))
|> cli.CommandAction("release", a => doRelease(a, initialTag))
|> cli.CommandAction("pick", a => doPick(a, initialTag))
```

**Implementation touches**:
- `internal/semantic/semantic_expressions.go` — infer lambda param types from call context
- `internal/parser/parser_expr.go` — extend multi-param untyped lambda support (currently only single-param `n => expr` works; need `(a, b) => expr`)
- `internal/codegen/codegen_decl.go` — emit inferred types in generated Go lambda output

**Inference sources** (in priority order):
1. Generic type parameter instantiation from piped value type
2. Expected `func(...)` parameter type at the call site
3. Multi-return context (e.g., `slice.Map` return type)

---

### 2. `# kuki:panics` Directive — Tier 2

**Status**: Not started
**Complexity**: Low
**Impact**: Moderate — Go has no mechanism to warn callers about functions that panic; this fills that gap

Extends the existing `# kuki:deprecated` / `# kuki:security` directive system. The compiler emits a warning at each call site of a function annotated with `# kuki:panics`.

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

**Implementation touches**:
- `internal/semantic/directives.go` — register `panics` as a valid directive
- `internal/semantic/check_calls.go` — emit warning at call sites
- `stdlib/must/*.kuki` — annotate all `must` functions as first consumers
- `cmd/genstdlibregistry/` — propagate `panics` metadata through the registry

**Design decisions**:
- Warning, not error — panics are valid in startup/must contexts
- A `--strict-panics` flag could promote to error for CI pipelines
- The directive message should appear in LSP hover tooltips

---

### 3. `# kuki:todo` Directive — Tier 2

**Status**: Not started
**Complexity**: Low
**Impact**: Moderate — especially useful for AI-generated code that flags incomplete sections

Emits a compile-time warning for any function or type annotated with `# kuki:todo`. Useful in CI to catch code that was left incomplete.

```kukicha
# kuki:todo "Add retry logic"
func fetchConfig(url string) (Config, error)
    return fetch.Get(url) |> fetch.Json(Config)
```

Output:

```
warning: TODO: "Add retry logic" on fetchConfig (config.kuki:15)
```

**Implementation touches**:
- `internal/semantic/directives.go` — register `todo` as a valid directive
- `internal/semantic/check_declarations.go` — emit warning for annotated declarations

**Design decisions**:
- Warning by default, `--strict-todos` flag promotes to error (for CI)
- Works on `func` and `type` declarations (same as other directives)
- Future consideration: inline form `# kuki:todo "msg"` on any line (not just declarations)

---

## Stdlib Additions

### 4. `stdlib/regex` — New Package

**Status**: Not started
**Complexity**: Low (wraps Go's `regexp`)
**Impact**: High — currently no way to do pattern matching without importing Go's `regexp` directly

The gh-semver-release example uses manual string prefix/suffix checks where regex would be cleaner, and validates git tags via expensive API calls that could be local regex checks.

**Before** (semver.kuki:30-32):

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

**Before** (main.kuki:75-77 — API call just to check if a tag exists):

```kukicha
func tagExists(repo string, tag string) bool
    result := shell.New("gh", "api", "repos/{repo}/git/ref/tags/{tag}") |> shell.Execute()
    return shell.Success(result)
```

**After** (with regex + git stdlib):

```kukicha
tags := git.ListTags(repo) onerr return false
return tags |> slice.Contains(tag)
```

**Proposed API**:

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
# ... etc
```

**Security**: The `# kuki:security` system should flag `regex.Match(userInput, ...)` where the pattern comes from untrusted input (ReDoS risk). Safe alternative: `regex.Compile` with timeout or `regex.MatchLiteral`.

---

### 5. `stdlib/git` — New Package

**Status**: Not started
**Complexity**: Medium (wraps `gh` and `git` CLI tools)
**Impact**: High for DevOps — gh-semver-release shells out to `git`/`gh` 10+ times

The example builds raw shell commands for every git/GitHub operation. A typed stdlib would provide structured returns, proper error context, and eliminate jq dependency.

**Before** (main.kuki:55-65 — raw shell + jq + manual line splitting):

```kukicha
func highestSemverTag(repo string) (string, error)
    raw := shell.Output("gh", "api", "repos/{repo}/tags",
        "--paginate", "--jq", ".[].name") onerr return "", error "failed to fetch tags for {repo}"

    if raw |> string.IsBlank()
        return "", empty

    tags := raw |> string.Lines()
    best, err := semver.Highest(tags)
    ...
```

**After**:

```kukicha
func highestSemverTag(repo string) (string, error)
    tags := git.ListTags(repo) onerr return "", error "failed to fetch tags for {repo}"
    if tags |> slice.IsEmpty()
        return "", empty
    return semver.Highest(tags)
```

**Before** (main.kuki:67-69):

```kukicha
func defaultBranch(repo string) (string, error)
    branch := shell.Output("gh", "repo", "view", repo,
        "--json", "defaultBranchRef", "--jq", ".defaultBranchRef.name") onerr return "", error "{error}"
    return branch |> string.TrimSpace(), empty
```

**After**:

```kukicha
func defaultBranch(repo string) (string, error)
    return git.DefaultBranch(repo)
```

**Proposed API** (GitHub-aware, uses `gh` CLI under the hood):

```kukicha
import "stdlib/git"

# Tags
git.ListTags(repo string) (list of string, error)
git.TagExists(repo string, tag string) (bool, error)
git.CreateTag(repo string, tag string, target string) error

# Branches
git.DefaultBranch(repo string) (string, error)
git.CurrentBranch() (string, error)
git.ListBranches(repo string) (list of string, error)

# Releases
git.CreateRelease(repo string, tag string, opts ReleaseOptions) error

type ReleaseOptions
    title string
    target string
    draft bool
    generateNotes bool

# Commits
git.Log(repo string, since string) (list of Commit, error)
git.LatestCommit(repo string) (Commit, error)

type Commit
    hash string
    message string
    author string
    date string

# Repository info
git.ListRepos(owner string) (list of string, error)
git.RepoExists(repo string) bool

# Local operations
git.Clone(url string, path string) error
git.CloneShallow(url string, path string, depth int) error
```

**Design note**: This package wraps `gh` (not raw git plumbing) since gh-semver-release and most DevOps scripts work with GitHub. A separate `stdlib/gitlocal` could wrap local `git` commands if needed later.

---

### 6. Shell Command Builder Improvements

**Status**: Not started
**Complexity**: Low (extends existing `stdlib/shell`)
**Impact**: Moderate — eliminates verbose conditional argument assembly

The example builds command argument lists with repeated `append()` calls and conditional flags.

**Before** (main.kuki:230-237):

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

**After**:

```kukicha
cmd := shell.New("gh", "release", "create", next, "--repo", repo, "--title", next)
    |> shell.FlagIf(not tagAlreadyExists, "--target", branch)
    |> shell.FlagIf(generateNotes, "--generate-notes")
    |> shell.FlagIf(draft, "--draft")

print("Command:")
print("  {shell.Preview(cmd)}")
```

**Proposed additions to `stdlib/shell`**:

```kukicha
# Conditional flag (added only when condition is true)
shell.FlagIf(cmd Command, condition bool, many args string) Command

# Display command as a string (properly quoted)
shell.Preview(cmd Command) string

# Append args to an existing command
shell.Args(cmd Command, many args string) Command

# Stdin piping
shell.Stdin(cmd Command, input string) Command
```

These are additive — no changes to existing `shell.New`, `shell.Execute`, etc.

---

### 7. `stdlib/parse` JSON Querying Additions

**Status**: Not started
**Complexity**: Medium
**Impact**: High for DevOps — eliminates dependency on external `jq` tool

The example embeds multi-line jq filters as unvalidated strings, delegating all JSON querying to the `jq` binary via shell.

**Before** (main.kuki:112-121 — 10-line jq filter as raw string):

```kukicha
jqFilter := """
    .data.user.repositoriesContributedTo.nodes[]
    | select(.isArchived == false and .isDisabled == false)
    | select(
        .viewerPermission == "ADMIN" or
        .viewerPermission == "MAINTAIN" or
        .viewerPermission == "WRITE"
      )
    | .nameWithOwner
    """

raw := shell.Output("gh", "api", "graphql", "--paginate",
    "-f", "login={me}", "-f", "query={query}",
    "--jq", jqFilter) onerr return list of string{}, error "failed to list repos: {error}"
```

**After** (fetch JSON, query in Kukicha):

```kukicha
raw := shell.Output("gh", "api", "graphql", "--paginate",
    "-f", "login={me}", "-f", "query={query}") onerr return list of string{}, error "failed to list repos: {error}"

repos := raw
    |> parse.JsonQuery(".data.user.repositoriesContributedTo.nodes")
    |> parse.JsonFilterObjects((obj map of string to any) =>
        obj["isArchived"] == false and obj["isDisabled"] == false and
        list of string{"ADMIN", "MAINTAIN", "WRITE"} |> slice.Contains(obj["viewerPermission"].(string))
    )
    |> parse.JsonPluck("nameWithOwner")
    onerr return list of string{}, error "failed to parse repos: {error}"
```

**Before** (main.kuki:56 — simple jq extraction):

```kukicha
raw := shell.Output("gh", "api", "repos/{repo}/tags",
    "--paginate", "--jq", ".[].name") onerr ...
```

**After**:

```kukicha
raw := shell.Output("gh", "api", "repos/{repo}/tags",
    "--paginate") onerr ...
names := raw |> parse.JsonQuery(".[].name") onerr ...
```

**Proposed additions to `stdlib/parse`**:

```kukicha
# Path-based extraction (jq-like dot notation)
parse.JsonQuery(data string, path string) (string, error)

# Query returning typed results
parse.JsonQueryStrings(data string, path string) (list of string, error)
parse.JsonQueryInts(data string, path string) (list of int, error)

# Filter array of objects by predicate
parse.JsonFilterObjects(data string, predicate func(map of string to any) bool) (string, error)

# Extract a single field from each object in an array
parse.JsonPluck(data string, field string) (list of string, error)

# Check if path exists
parse.JsonHas(data string, path string) bool
```

**Supported path syntax** (subset of jq):
- `.field` — object field access
- `.field.nested` — nested access
- `.[]` — iterate array
- `.[].field` — pluck field from array elements
- `.[0]` — index access

Full jq compatibility is a non-goal. The aim is to cover the 80% case (path extraction, array filtering, field plucking) that currently forces DevOps scripts to shell out to `jq`.

---

## Summary

| # | Item | Type | Tier | Complexity | Primary Benefit |
|---|------|------|------|------------|-----------------|
| 1 | Lambda parameter type inference | Language | 1 | Medium-high | Eliminates most common verbosity in pipe chains |
| 2 | `# kuki:panics` directive | Language | 2 | Low | Warns callers about panic risk (Go has no equivalent) |
| 3 | `# kuki:todo` directive | Language | 2 | Low | Compile-time reminders for incomplete code |
| 4 | `stdlib/regex` | Stdlib | 1 | Low | Pattern matching without raw Go imports |
| 5 | `stdlib/git` | Stdlib | 1 | Medium | Typed git/GitHub operations, eliminates jq + shell parsing |
| 6 | Shell builder improvements | Stdlib | 2 | Low | Conditional flag building without append boilerplate |
| 7 | `stdlib/parse` JSON querying | Stdlib | 1 | Medium | jq-like querying in Kukicha, no external tool dependency |

### Implementation Order (suggested)

1. **`stdlib/regex`** — unblocks many patterns, low complexity, no language changes
2. **Shell builder improvements** — small additive change to existing package
3. **`# kuki:panics` directive** — extends existing infrastructure, low risk
4. **`# kuki:todo` directive** — same infrastructure, trivial
5. **Lambda parameter type inference** — highest impact but most complex; benefits from having the other items landed first so the gh-semver-release rewrite can use everything
6. **`stdlib/git`** — depends on design decisions around gh vs local git
7. **`stdlib/parse` JSON querying** — most complex stdlib item, needs path syntax design
