---
title: Kukicha
sub_title: A beginner-friendly language that transpiles to Go
author: "github.com/tluker/kukicha"
---

<!-- end_slide -->

<!-- jump_to_middle -->

# Why does this language exist?

<!-- end_slide -->

## The Problem

AI writes code faster than humans can review it.

Modern Python has great tools — type hints, mypy, uv, pydantic.
But they're all **opt-in**. AI-generated code routinely skips them:

- Type hints omitted or incomplete — errors found at runtime, not review time
- Exception handling forgotten or too broad (`except Exception:`)
- No enforcement that the "right" tools are used
- Scripts that work locally but need packaging expertise to deploy

> Studies show **2.74×** more security vulnerabilities in AI-generated code
> than human-written code.

The problem isn't that Python *can't* be safe. It's that nothing *forces* it.

<!-- end_slide -->

## The Python Pain

```python
# What AI actually generates — valid Python, but how do you audit it?
def process(items, opts={}):
    if not items:
        raise ValueError("no items")
    for u in items:
        if u is not None and u.active and not u.banned:
            do_thing(u)
    return {"count": len(items)}
```

<!-- pause -->

Yes — a human would write `items: list[User]`, use pydantic, avoid `opts={}`.

But AI doesn't always follow best practices. And Python doesn't **require** it.

<!-- end_slide -->

## The R Pain

```r
# R analysis code — great for exploration, painful to productionize
process <- function(items, opts = list()) {
  if (length(items) == 0) stop("no items")
  results <- lapply(items, function(u) {
    if (!is.null(u) && u$active && !u$banned) do_thing(u)
  })
  list(count = length(items))
}
```

R has improved — `{box}` for modules, `{targets}` for pipelines, Posit Connect for hosting.

But there's still no path from analysis script to **standalone deployed binary**.
Error handling remains optional. Types remain advisory at best.

<!-- end_slide -->

## The Kukicha Answer

```
function process(items list of User, opts map of string to any) map of string to int, error
    if items equals empty or len(items) equals 0
        return empty, error "no items"
    for u in items
        if u.active and not u.banned
            doThing(u)
    return map of string to int{"count": len(items)}, empty
```

Same semantics. Plain English. **Every line is auditable.**

Compiles to a **single binary** — no Python environment, no R installation needed.

<!-- end_slide -->

## Why Not Just Write More Python?

<!-- pause -->

**Enforcement, not convention** — mypy and type hints are great, but optional.
Kukicha **requires** explicit types. AI can't skip them.

<!-- pause -->

**Deployment** — `uv` has made Python packaging much better. But you still
need a Python runtime on the target. Kukicha ships as one static binary.

<!-- pause -->

**Performance** — Kukicha compiles to native Go. Typical **10–100×** faster
than equivalent Python for CPU-bound work.

<!-- pause -->

**Parallelism** — Python 3.13+ has experimental free-threading (no GIL).
Kukicha has goroutines today — stable, production-ready, on every core.

<!-- end_slide -->

## Why Not Just Write More R?

<!-- pause -->

**Production gap** — R has gotten better — Plumber APIs, Posit Connect hosting,
`{targets}` pipelines. But deploying R still means installing R.
Kukicha builds one binary that runs anywhere.

<!-- pause -->

**Reproducibility** — `renv` is solid for locking dependencies.
But a single binary with zero runtime dependencies is a different level.

<!-- pause -->

**Performance** — R is fast for vectorized operations. For everything else,
Kukicha compiles to native Go.

<!-- pause -->

**But the pipes stay** — R's `|>` operator is Kukicha's `|>` operator.
Same idea. Same feel. More power.

<!-- end_slide -->

<!-- jump_to_middle -->

# What is Kukicha?

<!-- end_slide -->

## The Tea

Kukicha is a Japanese green tea from **Uji** — the birthplace of tea cultivation in Japan.

<!-- pause -->

Most green teas are made from the bud and first three leaves of each branch.
Kukicha is made from what's **left over**: the stems and stalks.

<!-- pause -->

The stems are cut to **exact, uniform lengths**, withered, and dried — transforming
something discarded into something refined.

<!-- pause -->

> "The leftover parts, treated with care, become something worth drinking."

<!-- end_slide -->

## Why We Chose This Name

Three parallels that felt too good to ignore:

<!-- pause -->

**The leftovers** — Kukicha is the layer that gets discarded once you compile.
Your `.kuki` files transpile to Go and the intermediate source is thrown away.
We are, literally, the stems.

<!-- pause -->

**Uniform cuts** — Just as kukicha stems are cut to exact lengths for a consistent
appearance, Kukicha code enforces 4-space indentation and explicit signatures.
Every file looks the same. Readable, reviewable, uniform.

<!-- pause -->

**Go green** — We originally targeted Go's experimental **green tea garbage collector**.
The tea name fit the Go ecosystem's green aesthetic — and stayed.

<!-- end_slide -->

## The Metaphor Runs Deep

The tea anatomy shows up directly in the language itself:

<!-- pause -->

**`stem.toml`** — the project root config file.
In the tea plant, the **stem** is the central stalk everything grows from.
In a Kukicha project, `stem.toml` is the anchor the compiler uses to
locate packages and calculate their names.

<!-- pause -->

**`petiole`** — the Kukicha keyword for package declarations.
In botany, the **petiole** is the thin stalk connecting a leaf blade to the stem.

```
petiole mypackage

import "stdlib/fetch"

func GetData() string
    ...
```

<!-- pause -->

Each source file is a leaf. Its petiole connects it back to the stem.

<!-- end_slide -->

## The Plant in One Diagram

```
  stem.toml              ← the stem (project root)
  │
  ├── src/auth/auth.kuki      petiole auth    ← leaf via petiole
  ├── src/api/api.kuki        petiole api     ← leaf via petiole
  └── main.kuki               (implicit main) ← the bud
```

<!-- pause -->

Just like kukicha tea: the **bud and leaves** become the product (Go binary).
The **stems and petioles** — the `.kuki` source files — are what's left over.

<!-- pause -->

The naming isn't decorative. It's a consistent metaphor for how code is structured.

<!-- end_slide -->

## Kukicha in One Sentence

Kukicha is a **beginner-friendly language that transpiles to Go.**

```
  .kuki file  →  Kukicha compiler  →  Go source  →  native binary
```

- Human-readable syntax (English operators, Python-style indentation)
- Compiles to a **single static binary** — no runtime, no Docker required
- Familiar concepts from Python and R (skills transfer directly)
- Built for the **AI-assisted code review** workflow

<!-- end_slide -->

## The Workflow

```
  1. Human describes what they want (in plain language)
       ↓
  2. AI writes Kukicha code
       ↓
  3. Human reviews — easy because it reads like English
       ↓
  4. kukicha build → single binary, ship it
```

<!-- pause -->

The human stays **in the loop** without needing to know Go internals.

<!-- end_slide -->

## Project Status

| Thing | Status |
|-------|--------|
| Version | 0.0.10 |
| Go requirement | 1.26+ |
| License | MIT |
| LSP support | ✓ (Zed, Neovim, etc.) |
| Stdlib packages | 35+ |
| Editor extension | Zed (tree-sitter + LSP) |

<!-- end_slide -->

<!-- jump_to_middle -->

# The Syntax

<!-- end_slide -->

## Operators: Already Familiar

If you know Python, you already know Kukicha's logic operators:

<!-- column_layout: [1, 1] -->

<!-- column: 0 -->

**Python / Kukicha**
```
x and y
x or y
not x
x == y        # or: x equals y
x != y        # or: x not equals y
None          # Kukicha: empty
```

<!-- column: 1 -->

**R**
```r
x && y
x || y
!x
x == y
x != y
NULL / NA / NaN
```

<!-- reset_layout -->

<!-- pause -->

Kukicha uses the same words as Python. R programmers: no more `&&` vs `||`.

One `empty` instead of R's three "nothing" values (`NULL`, `NA`, `NaN`).

<!-- end_slide -->

## Types: Readable Declarations

Python type hints exist, but they're optional and verbose. Kukicha's types
are concise, required, and self-explanatory:

<!-- column_layout: [1, 1] -->

<!-- column: 0 -->

**Python type hints**
```python
list[str]
dict[str, int]
Optional[User]
# or: User | None
```

<!-- column: 1 -->

**Kukicha**
```
list of string
map of string to int
reference User
```

<!-- reset_layout -->

<!-- pause -->

R has no static types — runtime surprises are common.

Kukicha catches type mismatches **at compile time**, not when a user hits the bug.

<!-- end_slide -->

## Blocks: Indentation (Like Python)

Python programmers: you'll feel at home immediately.

<!-- column_layout: [1, 1] -->

<!-- column: 0 -->

**Python**
```python
def process(n):
    if n > 0:
        print("positive")
    else:
        print("negative")
```

<!-- column: 1 -->

**Kukicha**
```
func process(n int)
    if n > 0
        print("positive")
    else
        print("negative")
```

<!-- reset_layout -->

<!-- pause -->

4-space indentation. No colons. No braces (sorry, R).

Explicit parameter and return types catch errors before they ever run.

<!-- end_slide -->

## Your First Kukicha Program

```
function main()
    print("Hello, World!")
```

<!-- pause -->

Build and run it:

```bash
kukicha run hello.kuki
# Hello, World!
```

<!-- pause -->

Or compile to a binary:

```bash
kukicha build hello.kuki
./hello
```

<!-- end_slide -->

<!-- jump_to_middle -->

# Variables & Functions

<!-- end_slide -->

## Variables

Inside functions — type inferred with `:=`:

```
count := 42
name := "Alice"
active := true
```

<!-- pause -->

Python: `:=` is like `=`. R: it's like `<-`. The type is inferred automatically.

Reassign with `=`:

```
count = count + 1
```

<!-- pause -->

File-scope variables need explicit types:

```
var AppVersion string = "1.0.0"
variable MaxRetries int = 3
```

<!-- end_slide -->

## Functions

Parameters and return types are **always explicit** — no runtime surprises:

```
func Add(a int, b int) int
    return a + b

func Greet(name string) string
    return "Hello, {name}!"
```

<!-- pause -->

Compare to Python, where types are optional and only checked by linters:

```python
def add(a, b):     # What types? Caught at runtime, not review time.
    return a + b
```

<!-- pause -->

Multiple return values (like returning a tuple in Python):

```
func Divide(a int, b int) (int, error)
    if b equals 0
        return 0, error "cannot divide by zero"
    return a / b, empty
```

<!-- end_slide -->

## Default Parameters & Named Arguments

Python has both. So does Kukicha:

```
func Greet(name string, greeting string = "Hello") string
    return "{greeting}, {name}!"
```

<!-- pause -->

Call with default:

```
msg := Greet("Alice")
# → "Hello, Alice!"
```

<!-- pause -->

Call with named argument:

```
msg := Greet("Alice", greeting: "Howdy")
# → "Howdy, Alice!"
```

<!-- end_slide -->

## String Interpolation

Python f-strings use `{var}`. Kukicha uses the same syntax — no `f"..."` prefix:

```
name := "Alice"
age := 30

greeting := "Hello, {name}! You are {age} years old."
```

<!-- pause -->

Works with expressions:

```
status := "The sum of 5+3 is {5 + 3}"
```

<!-- pause -->

Works with function calls:

```
import "stdlib/string"
upper := "Welcome, {name |> string.ToUpper()}!"
```

<!-- end_slide -->

## Methods on Types

Python puts methods inside a class. Kukicha separates them — function name
first, receiver type second:

```
func Display on todo Todo string
    return "{todo.id}: {todo.title}"
```

<!-- pause -->

Mutable receiver (modifies the value — like Python's `self` with mutable state):

```
func SetDone on todo reference Todo
    todo.done = true
```

<!-- pause -->

The receiver comes **after** the function name. It reads: "Display, defined on a Todo."

<!-- end_slide -->

## Types (Structs)

Like a Python `@dataclass` or a named R list — but checked at compile time:

```
type Todo
    id int64
    title string as "title"
    done bool as "done"
    tags list of string as "tags"
```

<!-- pause -->

`as "field"` = JSON key name. No decorator syntax, no extra imports.

<!-- pause -->

Function type aliases:

```
type Handler func(string)
type Transform func(int) (string, error)
```

<!-- end_slide -->

<!-- jump_to_middle -->

# Control Flow

<!-- end_slide -->

## If / Else

```
if count equals 0
    print("empty")
else if count < 10
    print("small")
else
    print("large")
```

<!-- pause -->

Python programmers: same structure, no colons required.

R programmers: no braces, `else if` instead of `else if (...)`.

<!-- end_slide -->

## For Loops

Iterate a collection (Python's `for item in items`):

```
for item in items
    process(item)
```

<!-- pause -->

Range with exclusive end (Python's `range(0, 10)`):

```
for i from 0 to 10      # 0, 1, 2, ... 9
    print(i)
```

<!-- pause -->

Range with inclusive end (R's `0:10`):

```
for i from 0 through 10 # 0, 1, 2, ... 10
    print(i)
```

<!-- end_slide -->

## Switch

```
switch command
    when "fetch", "pull"
        fetchRepos()
    when "help"
        showHelp()
    otherwise
        print("Unknown: {command}")
```

<!-- pause -->

Condition switch — like Python's `match/case` or a chain of `if/else if` in R:

```
switch
    when stars >= 1000
        print("Popular")
    when stars >= 100
        print("Growing")
    otherwise
        print("New project")
```

<!-- end_slide -->

## Negative Indexing & Collections

Python has negative indexing. So does Kukicha:

```
items := list of string{"a", "b", "c"}
last := items[-1]     # "c" — just like Python
```

<!-- pause -->

Maps (Python dicts / R named lists):

```
config := map of string to int{
    "port": 8080,
    "workers": 4,
}
```

<!-- end_slide -->

<!-- jump_to_middle -->

# Error Handling with `onerr`

<!-- end_slide -->

## The Python Way (Before)

```python
try:
    data = fetch_data()
except Exception as e:
    return None, f"failed to fetch: {e}"

try:
    port = get_port()
except Exception:
    port = 8080

try:
    result = compute()
except Exception as e:
    raise RuntimeError(f"compute failed: {e}")
```

Even well-written Python: 3 operations = 9 lines of error handling.
Each one is a decision point. Each one can be written differently.

<!-- end_slide -->

## The R Way (Before)

```r
data <- tryCatch(
  fetch_data(),
  error = function(e) stop(paste("failed to fetch:", e$message))
)

port <- tryCatch(get_port(), error = function(e) 8080)

result <- tryCatch(
  compute(),
  error = function(e) stop(paste("compute failed:", e$message))
)
```

Verbose. Deeply nested. Easily skipped entirely.

<!-- end_slide -->

## The `onerr` Way

Each operation declares its own error strategy **inline**:

```
data := fetchData() onerr return
port := getPort() onerr 8080
result := compute() onerr panic "compute failed: {error}"
_ := riskyOp() onerr discard
```

<!-- pause -->

One line. One intent. Hard to miss.

<!-- end_slide -->

## `onerr` Reference

| Form | Example | What it does |
|------|---------|--------------|
| Default value | `x := f() onerr 0` | Use fallback on error |
| Panic | `x := f() onerr panic "msg"` | Crash with message |
| Propagate (short) | `x := f() onerr return` | Pass error up unchanged |
| Propagate (verbose) | `x := f() onerr return empty, error "{error}"` | Wrap and propagate |
| Discard | `_ := f() onerr discard` | Ignore error |
| Explain | `x := f() onerr explain "context"` | Wrap with context |

<!-- end_slide -->

## Block-Style `onerr`

When you need more than one statement in the error handler:

```
users := csvData |> parse.CsvWithHeader() onerr
    print("Failed to parse CSV: {error}")
    metrics.Inc("parse.errors")
    return
```

<!-- pause -->

Custom error variable name with `as`:

```
payload := fetchData() onerr as e
    log.Error("fetch failed", "err", e)
    return
```

<!-- pause -->

**Critical rule:** The error is always named `error` inside `onerr`. Use `{error}` to
interpolate it. Use `onerr as name` to rename it.

<!-- end_slide -->

## `onerr explain` — Wrapping Errors

```
data := fetchData() onerr explain "failed to fetch data"
```

Wraps the original error with context — like Python's `raise X from e`.

<!-- pause -->

With a handler:

```
data := fetchData() onerr 0 explain "fetch failed"
```

Wraps the error, then runs the handler.

<!-- pause -->

Think of it as: "If this fails, **explain** the context, then handle it."

<!-- end_slide -->

<!-- jump_to_middle -->

# The Pipe Operator `|>`

<!-- end_slide -->

## R Programmers: You Already Know This

R's native pipe (`|>`, since R 4.1) and magrittr's `%>%` are widely loved:

```r
# R — left to right, readable
repos %>%
  filter(stars > 1000) %>%
  group_by(language) %>%
  summarise(count = n())
```

<!-- pause -->

Kukicha uses the **exact same operator** (`|>`):

```
repos
    |> slice.Filter((r Repo) => r.stars > 1000)
    |> slice.GroupBy((r Repo) => r.language)
```

Same idea. Same readability. Now compiles to a fast binary.

<!-- end_slide -->

## Python Programmers: Welcome to Pipes

Python's nested function calls read inside-out:

```python
# Python — read from inside out
result = format(transform(parse(validate(data))))
```

<!-- pause -->

Kukicha pipes go **left to right**:

```
result := data |> validate() |> parse() |> transform() |> format()
```

Reads like a recipe. Top to bottom. Step by step.

<!-- end_slide -->

## Pipe Basics

The left value becomes the **first argument**:

```
data |> parse()
# → parse(data)
```

<!-- pause -->

Chain multiple operations:

```
repos := fetch.Get(url)
    |> fetch.CheckStatus()
    |> fetch.Json(list of Repo)
    onerr panic "failed: {error}"
```

<!-- end_slide -->

## Pipe Placeholder `_`

When you need the piped value in a non-first position:

```
todo |> json.MarshalWrite(writer, _)
# → json.MarshalWrite(writer, todo)
```

<!-- pause -->

```
data |> slice.Insert(0, _)
# → slice.Insert(0, data)
```

<!-- pause -->

The `_` gives you full control of argument placement.

<!-- end_slide -->

## Bare Identifier Pipe

No parentheses needed for simple function calls:

```
data |> print
# → fmt.Println(data)
```

<!-- pause -->

Method shorthand — starts with `.`:

```
ctx |> exec.CommandContext("ls") |> .Output()
# → exec.CommandContext(ctx, "ls").Output()
```

<!-- end_slide -->

## Pipes + Lambdas

R programmers: this is purrr's `map` and `filter` with a cleaner lambda syntax.
Python programmers: like list comprehensions, but composable.

```
repos |> slice.Filter((r Repo) => r.stars > 100)
names := repos |> slice.Map((r Repo) => r.name)
```

<!-- pause -->

Single untyped param — concise like R's `~.x`:

```
numbers |> slice.Filter(n => n > 0)
```

<!-- pause -->

Block lambda for multi-step logic:

```
valid := items |> slice.Filter((item Item) =>
    processed := transform(item)
    return processed.valid
)
```

<!-- end_slide -->

## Arrow Lambdas Quick Reference

| Form | Example |
|------|---------|
| Expression, typed param | `(r Repo) => r.stars > 100` |
| Expression, untyped | `n => n > 0` |
| Zero params | `() => print("clicked")` |
| Block (multi-statement) | `(r Repo) =>` + indented body |

All auto-return the final expression (except block form).

<!-- end_slide -->

<!-- jump_to_middle -->

# Typed JSON Decoding

<!-- end_slide -->

## The Python Way (Before)

```python
import requests

response = requests.get(url)
response.raise_for_status()

# Returns a raw dict — shape is unknown until runtime
repos = response.json()

# Or with dataclasses — still a runtime KeyError if shape is wrong
repos = [Repo(**r) for r in response.json()]
```

Shape errors show up at runtime. No compile-time guarantee.

<!-- end_slide -->

## Kukicha's Typed Decode

Pass the type as a hint to `fetch.Json`:

```
repos := fetch.Get(url)
    |> fetch.CheckStatus()
    |> fetch.Json(list of Repo)
    onerr panic "fetch failed: {error}"
```

<!-- pause -->

The compiler knows the shape. Wrong field? Compile error, not runtime crash.

<!-- pause -->

Decode a single object:

```
user := fetch.Get(url)
    |> fetch.CheckStatus()
    |> fetch.Json(empty User)
    onerr panic "{error}"
```

<!-- pause -->

Decode a map:

```
config := fetch.Get(url)
    |> fetch.CheckStatus()
    |> fetch.Json(map of string to string)
    onerr panic "{error}"
```

<!-- end_slide -->

## How the Type Hint Works

The argument to `fetch.Json` is a **compile-time hint**, not a runtime value:

```
fetch.Json(list of Repo)             # Decodes []Repo
fetch.Json(empty Repo)               # Decodes Repo (zero value as hint)
fetch.Json(map of string to string)  # Decodes map[string]string
```

<!-- pause -->

The compiler infers the generic type and generates the right code.

No `json.loads()` ceremony, no `**kwargs` unpacking, no KeyError surprises.

<!-- end_slide -->

<!-- jump_to_middle -->

# The Standard Library

<!-- end_slide -->

## 35+ Packages, All in Kukicha Source

The stdlib is written in `.kuki` files — embedded in the compiler binary, extracted on `kukicha init`.

<!-- pause -->

Core categories:

| Category | Packages |
|----------|---------|
| Data | `slice`, `string`, `maps`, `json`, `parse`, `iterator` |
| HTTP | `fetch`, `http`, `validate`, `netguard` |
| Files | `files`, `encoding`, `template` |
| DB | `pg` (PostgreSQL), `kube` (Kubernetes), `container` (Docker) |
| AI/LLM | `llm`, `mcp`, `a2a` |
| Ops | `env`, `must`, `retry`, `ctx`, `concurrent`, `shell` |

<!-- end_slide -->

## `slice` Package

R programmers: this is your `dplyr` / `purrr` for lists.
Python programmers: this is `itertools` + list comprehensions, as a clean API.

```
import "stdlib/slice"

repos := fetchRepos()

popular := repos |> slice.Filter((r Repo) => r.stars > 1000)
names   := repos |> slice.Map((r Repo) => r.name)
byLang  := repos |> slice.GroupBy((r Repo) => r.language)
top     := repos |> slice.First(10)
```

<!-- pause -->

All functions are pipe-first. They just chain naturally.

<!-- end_slide -->

## `fetch` Package

```
import "stdlib/fetch"

type Repo
    name string as "name"
    stars int as "stargazers_count"

func getRepos(user string) list of Repo, error
    url := "https://api.github.com/users/{user}/repos"
    return fetch.Get(url)
        |> fetch.CheckStatus()
        |> fetch.Json(list of Repo)
        onerr return
```

<!-- pause -->

Authentication:

```
fetch.Get(url)
    |> fetch.BearerAuth(token)
    |> fetch.CheckStatus()
    |> fetch.Json(empty Response)
    onerr panic "{error}"
```

<!-- end_slide -->

## `env` and `must` Packages

Read environment variables with types:

```
import "stdlib/env"

port := env.GetOr("PORT", "8080")
debug := env.GetBool("DEBUG") onerr false
workers := env.GetInt("WORKERS") onerr 4
```

<!-- pause -->

For startup requirements (panic if missing):

```
import "stdlib/must"

dbURL := must.Env("DATABASE_URL")
port := must.EnvIntOr("PORT", 8080)
```

<!-- end_slide -->

## `retry` Package

```
import "stdlib/retry"

result := retry.New()
    |> retry.Attempts(3)
    |> retry.Delay(500)
    |> retry.Do(fetchData)
    onerr panic "all retries failed: {error}"
```

<!-- pause -->

Exponential backoff, configurable attempts, clean pipe syntax.

<!-- end_slide -->

## `concurrent` Package

Run tasks in parallel — with real parallelism, not Python's GIL:

```
import "stdlib/concurrent"

urls := list of string{
    "https://example.com",
    "https://google.com",
    "https://github.com",
}

results := urls |> concurrent.Parallel((url string) string, error =>
    return fetch.Get(url) |> fetch.Text() onerr return
) onerr panic "{error}"
```

<!-- pause -->

All URLs fetched simultaneously. No threading module. No asyncio. No GIL.

<!-- end_slide -->

## `llm` Package

Built-in LLM integration — multiple providers:

```
import "stdlib/llm"

# Quick one-shot completion
reply := llm.Complete("openai:gpt-4o-mini", "What is Kukicha?")
    onerr panic "{error}"
print("Reply: {reply}")
```

<!-- pause -->

Builder pattern:

```
answer := llm.New("anthropic:claude-3-5-sonnet")
    |> llm.System("You are an expert who explains things simply.")
    |> llm.Temperature(0.3)
    |> llm.Ask("Explain goroutines in one sentence")
    onerr panic "{error}"
```

<!-- end_slide -->

## `mcp` Package — Model Context Protocol

Expose your Kukicha tools to AI agents:

```
import "stdlib/mcp"
import "stdlib/cast"

func add(args map of string to any) (any, error)
    a := args["a"] |> cast.SmartInt() onerr 0
    b := args["b"] |> cast.SmartInt() onerr 0
    return "{a} + {b} = {a + b}" as any, empty

func main()
    server := mcp.New("Calculator", "1.0.0")
    server |> mcp.Tool("add", "Add two numbers", schema, add)
    mcp.Serve(server) onerr panic "Server error: {error}"
```

<!-- end_slide -->

<!-- jump_to_middle -->

# Deploy Anywhere

<!-- end_slide -->

## The Python Deployment Problem

```
Python script works locally.
    ↓
uv sync (much better than pip freeze!)
    ↓
Docker image — still needs Python runtime (~100 MB with uv, slim base)
    ↓
Or: PyInstaller / Nuitka (bundles runtime, ~50-80 MB, platform-specific)
    ↓
Works — but you're carrying a runtime either way
```

<!-- pause -->

Kukicha:

```
kukicha build app.kuki
./app        ← one static binary (~10 MB), no runtime, cross-compiles
```

<!-- end_slide -->

## The R Deployment Problem

```
R analysis script works in RStudio.
    ↓
Plumber API or Shiny app to expose it
    ↓
renv to lock package versions (works well now)
    ↓
Posit Connect — great if your org pays for it
    ↓
Self-hosting: still need R + compiled packages on the server
```

<!-- pause -->

Kukicha:

```
kukicha build analysis-api.kuki
./analysis-api    ← one binary, no R required, runs anywhere
```

<!-- end_slide -->

## Single Binary. Ship It.

```bash
kukicha build app.kuki
# → ./app

scp ./app user@server:/opt/app
ssh user@server /opt/app
```

<!-- pause -->

- No Python runtime to install
- No R packages to compile
- No Docker (unless you want it)
- Cross-compile for any OS/architecture

<!-- pause -->

This is what Go gives you. Kukicha just makes it readable.

<!-- end_slide -->

<!-- jump_to_middle -->

# Concurrency

<!-- end_slide -->

## Channels & Goroutines

Python's GIL and R's single-threaded model make true parallelism painful.
Kukicha has goroutines — lightweight threads that use all your cores.

<!-- pause -->

Create a channel:

```
ch := make(channel of string)
done := make(channel of bool)
```

<!-- pause -->

Send and receive:

```
send "hello" to ch
msg := receive from ch
```

<!-- pause -->

Spawn a goroutine:

```
go doWork()
```

<!-- pause -->

Multi-statement goroutine block:

```
go
    mu.Lock()
    doWork()
    mu.Unlock()
```

<!-- end_slide -->

## Select — Channel Multiplexing

```
select
    when msg := receive from ch
        print("Got: {msg}")
    when receive from done
        print("All done")
    when send "ping" to out
        print("Sent ping")
    otherwise
        print("Nothing ready — non-blocking")
```

<!-- pause -->

Reads exactly like the behavior it describes.

<!-- end_slide -->

## Variadic Arguments

Declare with `many`:

```
func Sum(many numbers int) int
    total := 0
    for n in numbers
        total = total + n
    return total
```

<!-- pause -->

Call with individual args:

```
result := Sum(1, 2, 3, 4, 5)
```

<!-- pause -->

Spread a slice:

```
args := list of int{10, 20, 30}
result := Sum(many args)
```

<!-- end_slide -->

<!-- jump_to_middle -->

# Security Built-In

<!-- end_slide -->

## Compiler-Enforced Security Checks

The Kukicha compiler rejects code with common vulnerabilities **at compile time**:

| Attack | What the compiler catches |
|--------|--------------------------|
| SQL injection | String interpolation in raw queries |
| XSS | Unescaped user input in HTML responses |
| SSRF | Unvalidated URLs passed to `fetch.Get` |
| Path traversal | User input in file paths |
| Command injection | User input in shell commands |
| Open redirect | User-controlled redirect URLs |

<!-- pause -->

Safe alternatives are provided in the stdlib (`netguard`, `validate`, `http.SafeURL`).

<!-- end_slide -->

## Security Example

This is rejected at compile time:

```
# ❌ COMPILE ERROR: SQL injection risk
query := "SELECT * FROM users WHERE id = {userId}"
db.Query(query)
```

<!-- pause -->

The safe way:

```
# ✓ Parameterized query
result := pg.QueryRow(db, "SELECT * FROM users WHERE id = $1", userId)
    onerr panic "{error}"
```

<!-- end_slide -->

<!-- jump_to_middle -->

# Real-World Example

<!-- end_slide -->

## GitHub Repo Explorer

```
import "stdlib/fetch"
import "stdlib/slice"

type Repo
    name string as "name"
    stars int as "stargazers_count"
    language string as "language"

func main()
    repos := fetch.Get("https://api.github.com/users/golang/repos")
        |> fetch.CheckStatus()
        |> fetch.Json(list of Repo)
        onerr panic "fetch failed: {error}"

    popular := repos |> slice.Filter((r Repo) => r.stars > 1000)
    byLang := popular |> slice.GroupBy((r Repo) => r.language)

    for lang, group in byLang
        print("{lang}: {len(group)} repos")
```

<!-- end_slide -->

## Concurrent URL Health Checker

```
import "stdlib/fetch"
import "stdlib/concurrent"

func checkURL(url string) (string, error)
    resp := fetch.Get(url) onerr return "", error "unreachable: {error}"
    return "{url}: {resp.status}", empty

func main()
    urls := list of string{
        "https://google.com",
        "https://github.com",
        "https://example.com",
    }

    results := urls |> concurrent.Parallel(checkURL) onerr panic "{error}"

    for result in results
        print(result)
```

<!-- end_slide -->

## PostgreSQL Integration

```
import "stdlib/pg"
import "stdlib/must"

type User
    id int64 as "id"
    email string as "email"
    active bool as "active"

func main()
    db := pg.Connect(must.Env("DATABASE_URL")) onerr panic "{error}"
    defer db.Close()

    users := pg.Query(db, User{}, "SELECT id, email, active FROM users WHERE active = $1", true)
        onerr panic "query failed: {error}"

    for user in users
        print("{user.email}")
```

<!-- end_slide -->

<!-- jump_to_middle -->

# The Compiler

<!-- end_slide -->

## Five-Phase Pipeline

```
  .kuki source
       │
       ▼
  1. Lexer        — tokens, INDENT/DEDENT handling
       │
       ▼
  2. Parser       — recursive descent → AST
       │
       ▼
  3. Semantic     — type checking, symbol table
       │
       ▼
  4. Codegen      — AST → idiomatic Go source
       │
       ▼
  5. go build     — native binary
```

<!-- end_slide -->

## Codegen: `onerr` Expansion

Kukicha:

```
data := fetchData() onerr return empty, error "fetch failed: {error}"
```

<!-- pause -->

Generated Go:

```go
data, err_1 := fetchData()
if err_1 != nil {
    return nil, fmt.Errorf("fetch failed: %w", err_1)
}
```

<!-- pause -->

Unique variable names (`err_1`, `err_2`) prevent shadowing.

<!-- end_slide -->

## Codegen: Pipe Expansion

Kukicha:

```
result := data
    |> slice.Filter((r Repo) => r.stars > 100)
    |> slice.Map((r Repo) => r.name)
```

<!-- pause -->

Generated Go:

```go
result := slice.Map(
    slice.Filter(data, func(r Repo) bool {
        return r.Stars > 100
    }),
    func(r Repo) string {
        return r.Name
    },
)
```

<!-- end_slide -->

## Generic Type Inference

Stdlib uses placeholder names for generic parameters:

```
# stdlib/slice source (you don't write this)
func Filter(items list of any, predicate func(any) bool) list of any
```

<!-- pause -->

When you write:

```
repos |> slice.Filter((r Repo) => r.stars > 100)
```

The compiler infers `T = Repo` and generates the correct generic Go code.

No generics syntax in user code. Ever.

<!-- end_slide -->

<!-- jump_to_middle -->

# Getting Started

<!-- end_slide -->

## Install

```bash
# Build from source
git clone https://github.com/tluker/kukicha
cd kukicha
make build

# Or install directly
go install github.com/tluker/kukicha/cmd/kukicha@latest
```

<!-- end_slide -->

## Project Initialization

```bash
mkdir myapp && cd myapp
kukicha init
```

<!-- pause -->

This runs `go mod init` and extracts the stdlib to `.kukicha/stdlib/`.

<!-- pause -->

```bash
myapp/
├── go.mod              # Go module (managed automatically)
├── go.sum
└── .kukicha/
    └── stdlib/         # Standard library, ready to import
```

<!-- end_slide -->

## CLI Commands

| Command | What it does |
|---------|-------------|
| `kukicha run file.kuki` | Transpile, compile, and run |
| `kukicha build file.kuki` | Compile to binary |
| `kukicha check file.kuki` | Validate syntax (no compile) |
| `kukicha fmt -w file.kuki` | Format in place |
| `kukicha init` | Initialize a new project |

<!-- pause -->

Always run `kukicha check` before committing `.kuki` changes.

<!-- end_slide -->

## Editor Support

LSP server for autocompletion and diagnostics:

```bash
make install-lsp
# Installs kukicha-lsp to your PATH
```

<!-- pause -->

**Zed** — install the official extension (tree-sitter grammar + LSP)

**Neovim** — configure `kukicha-lsp` in your LSP setup

Any editor that supports LSP works out of the box.

<!-- end_slide -->

## Writing Your First Real Program

```
import "stdlib/fetch"

type Post
    id int as "id"
    title string as "title"
    body string as "body"

function main()
    posts := fetch.Get("https://jsonplaceholder.typicode.com/posts")
        |> fetch.CheckStatus()
        |> fetch.Json(list of Post)
        onerr panic "failed: {error}"

    for post in posts
        print("{post.id}: {post.title}")
```

<!-- end_slide -->

## Key Rules to Remember

1. **4-space indentation only** — tabs are rejected
2. **Explicit function signatures** — no implicit types on params/returns
3. **`{error}` inside onerr** — not `{err}`, always `{error}`
4. **`onerr` is mandatory** — errors must be handled inline
5. **`kukicha check` first** — validate before you commit

<!-- end_slide -->

<!-- jump_to_middle -->

# Summary

<!-- end_slide -->

## What We Covered

<!-- incremental_lists: true -->

- **The problem**: AI code is hard to review safely; Python/R errors hide until runtime
- **The solution**: English-syntax language that compiles to a fast, deployable binary
- **Python fit**: familiar indentation, `and`/`or`/`not`, f-string interpolation, negative indexing
- **R fit**: the `|>` pipe you love, `slice.Map/Filter/GroupBy` like dplyr/purrr
- **Functions**: explicit signatures, default params, named args, methods
- **Error handling**: `onerr` inline — replaces try/except and tryCatch
- **Pipes**: left-to-right data flow with `|>` — R's native operator
- **Typed JSON**: `fetch.Json(list of T)` — no dict unpacking, no runtime KeyError
- **Stdlib**: 35+ packages — HTTP, DB, LLM, Kubernetes, MCP
- **Security**: compile-time injection and traversal checks
- **Concurrency**: real parallelism — no GIL, no single-threaded limitation
- **Deploy**: one binary, no runtime, ships anywhere

<!-- end_slide -->

## Kukicha's Place in the Stack

```
  You describe what you want  (natural language)
       ↓
  AI writes Kukicha           (readable, auditable)
       ↓
  You review                  (plain English syntax)
       ↓
  kukicha build               (single native binary)
       ↓
  Ship it                     (no Python env, no R installation)
```

<!-- pause -->

Kukicha is the **translation layer** between human intent and machine execution
in the age of AI-assisted development.

<!-- end_slide -->

## Resources

| Resource | Location |
|----------|---------|
| Source code | `github.com/tluker/kukicha` |
| Documentation | `docs/` |
| Tutorials | `docs/tutorials/` |
| Quick reference | `docs/kukicha-quick-reference.md` |
| Stdlib reference | `stdlib/CLAUDE.md` |
| FAQ | `docs/faq.md` |
| Grammar spec | `docs/kukicha-grammar.ebnf.md` |

<!-- end_slide -->

<!-- jump_to_middle -->

# Questions?

<!-- new_line -->

```
kukicha run your-idea.kuki
```

<!-- end_slide -->
