# Draft: Sum Types in gh-semver-release

> This draft shows how sum types would improve the gh-semver-release example.
> The changes are surgical — most of the program stays the same.

---

## The Bump Type

The current code passes bump as a raw string everywhere:

```kukicha
const DefaultBump = "patch"

func nextTag(current string, bump string, initialTag string) (string, error)
    # ...
    return v |> semver.Bump(bump) |> semver.Format(), empty
```

The CLI flag accepts `"patch"`, `"minor"`, or `"major"` — but the type system doesn't enforce this. If someone refactors and passes `"PATCH"` or `"bug"`, the error shows up at runtime inside `semver.Bump`, not at the call site.

### With a sum type

```kukicha
type Bump = Patch | Minor | Major
```

One line. No fields — these are pure labels (like an enum). Now `bump` is a `Bump`, not a `string`, and invalid values are impossible:

```kukicha
func nextTag(current string, bump Bump, initialTag string) (string, error)
    if current equals ""
        return initialTag, empty
    v := semver.Parse(current) onerr return "", error "{error}"
    return v |> semver.Bump(bump) |> semver.Format(), empty
```

### Parsing the CLI flag

You still need to convert the user's string input into a `Bump`. That happens once, at the boundary:

```kukicha
func parseBump(s string) (Bump, error)
    return s |> switch
        when "patch"
            return Patch{}, empty
        when "minor"
            return Minor{}, empty
        when "major"
            return Major{}, empty
        otherwise
            return Patch{}, error "invalid bump type: '{s}' (expected patch, minor, or major)"
```

After this point, every function in the program works with `Bump` — no string matching, no validation, no "what if someone passes an empty string?" defensive checks.

### Where the types flow

Here's the cascade of changes — most are just `string` -> `Bump` in the signature:

```kukicha
# Before: bump is a string everywhere
func buildEntry(repo string, bump string, initialTag string) RepoEntry
func doList(args cli.Args, initialTag string)
    bump := cli.GetString(args, "bump")

# After: bump is parsed once, typed everywhere
func buildEntry(repo string, bump Bump, initialTag string) RepoEntry
func doList(args cli.Args, initialTag string)
    bump := cli.GetString(args, "bump") |> parseBump() onerr
        fatal("Error: {error}")
```

The only place that deals with the raw string is `parseBump`. Everything downstream gets a `Bump` — the compiler guarantees it's valid.

---

## Release Status (optional, bigger refactor)

The `doRelease` function has several possible outcomes scattered across branches:

```kukicha
# Current code — outcomes encoded as control flow
if repo equals ""
    fatal("release requires --repo OWNER/REPO")
# ...
if exists
    print("Release already exists:")
    return
# ...
if dryRun
    print("Dry run only. No release created.")
    return
# ...
git.CreateRelease(repo, next, opts) onerr
    fatal("(another process may have created the release or tag).")
```

You could model this as a sum type if you wanted to separate "decide what to do" from "do it":

```kukicha
type ReleaseOutcome
    = Created(repo string, tag string)
    | AlreadyExists(repo string, tag string)
    | DryRun(repo string, tag string, command string)
    | Failed(reason string)
```

But honestly, this is over-engineering for a CLI tool — the current control flow is clear enough. Sum types shine most when the variants are **used in multiple places** (like `Bump` flowing through 5 functions, or `HealthStatus` being switched on in display, logging, and alerting). A one-shot outcome in a single function doesn't benefit much.

---

## Revised Types Section

Here's what the top of `main.kuki` would look like:

```kukicha
# ── Types ──────────────────────────────────────────────────────

type Bump = Patch | Minor | Major

type RepoEntry
    name string
    current string
    next string

# ── Helpers ───────────────────────────────────────────────────

func fatal(msg string)
    fmt.Fprintln(os.Stderr, msg)
    os.Exit(1)

func warn(msg string)
    fmt.Fprintln(os.Stderr, "Warning: {msg}")

func parseBump(s string) (Bump, error)
    return s |> switch
        when "patch"
            return Patch{}, empty
        when "minor"
            return Minor{}, empty
        when "major"
            return Major{}, empty
        otherwise
            return Patch{}, error "invalid bump type: '{s}' (expected patch, minor, or major)"

func nextTag(current string, bump Bump, initialTag string) (string, error)
    if current equals ""
        return initialTag, empty
    v := semver.Parse(current) onerr return "", error "{error}"
    return v |> semver.Bump(bump) |> semver.Format(), empty
```

The `const DefaultBump = "patch"` stays as a string (it's the CLI flag default). Parsing happens when the flag is read:

```kukicha
func doList(args cli.Args, initialTag string)
    org := cli.GetString(args, "org")
    bump := cli.GetString(args, "bump") |> parseBump() onerr
        fatal("{error}")
    csv := cli.GetBool(args, "csv")

    # From here on, bump is a Bump — not a string
    entries := concurrent.MapWithLimit(repos, 4, r => buildEntry(r, bump, initialTag))
        |> slice.Filter(e => e.name not equals "")
        |> sort.ByKey(e => e.name)
    # ... rest unchanged
```

---

## What Changes, What Doesn't

| Part of the program | Changes? | Why |
|---------------------|----------|-----|
| `type Bump` declaration | **New** | Replaces the implicit `"patch" \| "minor" \| "major"` contract |
| `parseBump` function | **New** | Converts CLI string to typed `Bump` once |
| `nextTag` signature | `string` -> `Bump` | Carries the type through |
| `buildEntry` signature | `string` -> `Bump` | Carries the type through |
| `doList` / `doRelease` / `doPick` | Parse bump at the top | One `parseBump()` call each |
| CLI flag definition | **Unchanged** | `cli.GlobalFlag("bump", ...)` still takes a string default |
| GraphQL query, jq filter | **Unchanged** | No relation to bump type |
| `RepoEntry`, table display | **Unchanged** | These deal with repo names and tags, not bump |
| `concurrent.MapWithLimit` | **Unchanged** | Just passes `bump` through |

The refactor is small — add `type Bump`, add `parseBump`, change 4-5 function signatures from `string` to `Bump`. The payoff is that the compiler now prevents passing garbage into `semver.Bump()`.

---

## The `|` Separator — A Note on Syntax

In the single-line form:

```kukicha
type Bump = Patch | Minor | Major
```

The `|` means "or" — it separates the variants of a sum type. This is distinct from `|>` (the pipe operator):

- `|` appears only in `type` declarations, between variant names
- `|>` appears in expressions, chaining function calls

They never appear in the same context, so there's no ambiguity for the compiler or the reader. If you're scanning code and see `|` after an `=` inside a `type`, it's a sum type. If you see `|>` in an expression, it's a pipe.

Multi-line form uses the same `|` but with indentation:

```kukicha
type HealthStatus
    = Up
    | Down(reason string)
    | HttpError(code int)
```
