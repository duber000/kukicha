---
name: kukicha
description: Help write, debug, and understand Kukicha code - a beginner-friendly language that transpiles to Go. Use when working with .kuki files, discussing Kukicha syntax, error handling with onerr, pipe operators, or the Kukicha compiler/transpiler.
---

# Kukicha Language Skill

You are helping with **Kukicha** (茎 = "stem"), a beginner-friendly programming language that compiles to idiomatic Go code. The core philosophy is **"It's Just Go"** - Kukicha is syntactic sugar with zero runtime overhead.

## Quick Reference

### File Extension
- `.kuki` files contain Kukicha source code
- Transpiles to `.go` files

### CLI Commands
```bash
kukicha init                   # Extract stdlib, configure go.mod (run once per project)
kukicha build <file.kuki>      # Compile to Go binary
kukicha run <file.kuki>        # Compile and run
kukicha check <file.kuki>      # Type-check only
kukicha fmt [options] <path>   # Format code
kukicha version                # Show version
```

**Note:** `kukicha init` extracts the embedded stdlib to `.kukicha/stdlib/` and adds a `replace` directive to your `go.mod`. This is required when using `import "stdlib/..."` packages. The `build` and `run` commands auto-extract if needed, but running `init` explicitly is recommended for new projects.

## Syntax Essentials

### Variables (Walrus Operator)
```kukicha
count := 42          # Create new binding (short declaration)
count = 100          # Reassign existing variable
```

### Functions (Explicit Types Required)
```kukicha
func Greet(name string) string
    return "Hello {name}"

# Multiple returns
func Divide(a int, b int) int, error
    if b equals 0
        return 0, error "division by zero"
    return a / b, empty

# No return value
func PrintMessage(msg string)
    fmt.Println(msg)
```

### Methods (Receiver Syntax)
```kukicha
# Syntax: func Name on receiverName ReceiverType ReturnType
func Display on todo Todo string
    return "{todo.id}: {todo.title}"

# Pointer receiver
func SetTitle on todo reference Todo
    todo.title = "New Title"
```

### String Interpolation
```kukicha
name := "World"
greeting := "Hello {name}!"        # Transpiles to fmt.Sprintf
complex := "Result: {a + b}"       # Expressions allowed
```

### Error Handling (onerr Operator)
```kukicha
# Panic on error
content := file.read("config.json") onerr panic "missing file"

# Provide default value
port := env.get("PORT") onerr "8080"

# Return error to caller
data := fetchData() onerr return empty, error "{error}"

# Discard error (use sparingly)
result := riskyOp() onerr discard

# Explain syntax - wrap error with hint message
data := fetchData() onerr explain "failed to fetch data"  # Standalone: returns wrapped error
data := fetchData() onerr 0 explain "fetch failed"        # With handler: wraps error, then runs handler

# Block-style onerr (multi-statement error handling)
users := csvData |> parse.CsvWithHeader() onerr
    print("Failed to parse: {error}")    # {error} refers to the caught error
    return
```

### Pipe Operator
```kukicha
result := data
    |> parse()
    |> transform()
    |> process()

# With arrow lambdas — concise inline predicates
users := fetchUsers()
    |> slice.Filter((u User) => u.active)
    |> slice.Map((u User) => u.name)

# Named functions also work as callbacks
func isActive(u User) bool
    return u.active

users |> slice.Filter(isActive)

# Placeholder strategy: use _ to specify where piped value goes
# Useful when piped value isn't the first argument
todo |> json.MarshalWrite(writer, _)     # Becomes: json.MarshalWrite(writer, todo)
data |> encode(options, _, format)       # Becomes: encode(options, data, format)

# Bare identifier as pipe target (no parentheses needed)
data |> print                            # Becomes: fmt.Println(data)
```

### Arrow Lambdas (`=>`)
```kukicha
# Expression lambda — auto-return, no `return` keyword needed
repos |> slice.Filter((r Repo) => r.Stars > 100)
repos |> slice.Map((r Repo) => r.Name)

# Single untyped param — no parens needed
numbers |> slice.Filter(n => n > 0)

# Zero params
button.OnClick(() => print("clicked"))

# Block lambda — multi-statement, explicit return
repos |> slice.Filter((r Repo) =>
    name := r.Name |> string.ToLower()
    return name |> string.Contains("go")
)
```

### Control Flow
```kukicha
# If statements (use 'equals' for comparison)
if count equals 0
    return "empty"
else if count < 10
    return "small"
else
    return "large"

# For loops
for item in items
    process(item)

for index, item in items
    fmt.Println("{index}: {item}")

for i from 0 to 10           # 0 to 9 (exclusive)
    fmt.Println(i)

for i from 0 through 10      # 0 to 10 (inclusive)
    fmt.Println(i)

for count > 0                # While-style loop
    count--

 for                          # Infinite loop
     if something
         break
     continue

# Switch with when/otherwise
switch command
    when "fetch", "pull"
        fetchRepos()
    when "help"
        showHelp()
    otherwise
        print("Unknown command")

# Condition switch (bare switch, like Go's switch {})
switch
    when stars >= 1000
        print("Popular")
    when stars >= 100
        print("Growing")
    otherwise
        print("New")

# Type switch
switch event as e
    when reference a2a.TaskStatusUpdateEvent
        print(e.Status.State)
    when reference a2a.Task
        print(e.ID)
    when string
        print(e)
    otherwise
        print("Unknown event")
```

### Types
```kukicha
type Todo
    id int64
    title string
    completed bool
    tags list of string
    metadata map of string to string

# With struct tags (for JSON, database mapping, etc.)
type User
    ID int64 json:"id"
    Name string json:"name"
    Email string json:"email"
    Active bool json:"active"

# Struct tags support any Go format: json, xml, db, validate, etc.
# Multiple tags: json:"name" db:"user_name"

# JSON alias sugar (shorthand for json:"alias")
type Article
    id int64
    title string as "title"       # Generates json:"title" struct tag
    publishedAt string as "published_at"

# Function type aliases
type Handler func(string)
type Transform func(int) (string, error)
type Predicate func(reference User) bool

# Interface
interface Storage
    Save(item Todo) error
    Load(id int64) Todo, error
```

### Collections
```kukicha
# Lists (slices)
items := list of string{"a", "b", "c"}
last := items[-1]                    # Negative indexing supported

# Maps — literal syntax works
config := map of string to int{"port": 8080}

# Or use make for empty maps
config := make(map of string to int)
config["port"] = 8080

# Check membership (use slices.Contains — 'in' only works in for loops)
if slices.Contains(admins, user)
    grantAccess()
```

### Pointers & References
```kukicha
# Pointer type
userPtr reference User

# Address-of
ptr := reference of user

# Dereference
val := dereference ptr

# Common pattern with json
json.Unmarshal(data, reference of result) onerr panic "parse failed"
```

### Concurrency
```kukicha
# Goroutines — single call
go fetchData(url)

# Goroutines — block form (multi-statement)
go
    mu.Lock()
    doWork()
    mu.Unlock()

# Channels
ch := make channel of string
send ch, "message"
msg := receive from ch
close(ch)
```

### Null/Nil
```kukicha
# Use 'empty' instead of nil
if user equals empty
    return error "user not found"

# Typed empty for returns
return empty list of string
```

### Boolean Operators
```kukicha
# Use English keywords
if a and b
if a or b
if not done

# Equality
if x equals y      # Same as ==
if x != y          # Not equals still uses !=

# Complex Logic (Improved Parser)
if err != empty or failed
    handleError()
```

### Variadic Arguments (`many`)
```kukicha
# Declare variadic param with "many" before param name
func Sum(many numbers int) int
    total := 0
    for n in numbers
        total = total + n
    return total

# Call with individual args
result := Sum(1, 2, 3)

# Spread a slice with "many" at call site
args := list of int{1, 2, 3}
result := Sum(many args)
```

### Type Assertions
```kukicha
# Two-value form (safe — check ok before using result)
result, ok := value.(string)
if ok
    print("string: {result}")

# Direct assertion (panics if wrong type)
s := value.(string)
```

### Output (print Builtin)
```kukicha
# print automatically imports fmt and transpiles to fmt.Println()
print("Hello World")
print("Value:", count, "items")       # Variadic - accepts multiple arguments
print(user.Name, user.Age)            # Works with any types
```

### Function Types (Callbacks & Higher-Order Functions)
```kukicha
# Simple callback: takes int, returns bool
func Filter(items list of int, predicate func(int) bool) list of int
    result := list of int{}
    for item in items
        if predicate(item)
            result = append(result, item)
    return result

# Arrow lambda — preferred for short inline predicates
evens := Filter(list of int{1, 2, 3, 4}, (n int) => n % 2 equals 0)

# Named function — use for complex or reusable logic
func isEven(n int) bool
    return n % 2 equals 0

evens := Filter(list of int{1, 2, 3, 4}, isEven)
```

## Transpilation Patterns

| Kukicha | Go |
|---------|-----|
| `list of int` | `[]int` |
| `map of string to int` | `map[string]int` |
| `func(int) bool` | `func(int) bool` |
| `func(string)` | `func(string)` |
| `reference User` | `*User` |
| `reference of x` | `&x` |
| `dereference ptr` | `*ptr` |
| `empty` | `nil` |
| `and`, `or`, `not` | &&, \|\|, ! |
| `equals` | `==` |
| `print(...)` | `fmt.Println(...)` |
| `"Hello {name}"` | `fmt.Sprintf("Hello %s", name)` |
| `items[-1]` | `items[len(items)-1]` |
| `json:"name"` (struct tag) | `` `json:"name"` `` (backtick-quoted) |
| Indentation blocks | `{ }` braces |
| `a \|> f(b)` | `f(a, b)` |
| `a \|> f(b, _)` | `f(b, a)` (placeholder) |
| `a \|> print` | `fmt.Println(a)` (bare identifier) |
| `x := f() onerr "default"` | `x, err := f(); if err != nil { x = "default" }` |
| `x := f() onerr discard` | `x, _ := f()` |
| `x := f() onerr explain "hint"` | `x, err := f(); if err != nil { return ..., fmt.Errorf("hint: %w", err) }` |
| `x := f() onerr 0 explain "hint"` | `x, err := f(); if err != nil { x = 0; err = fmt.Errorf("hint: %w", err) }` |
| `break` | `break` |
| `continue` | `continue` |
| `for` (bare) | `for { ... }` |
| `(r Repo) => r.Stars > 100` | `func(r Repo) bool { return r.Stars > 100 }` |
| `go` + indented block | `go func() { ... }()` |
| `switch x` / `when a, b` / `otherwise` | `switch x { case a, b: ... default: ... }` |
| `switch x as v` / `when reference T` | `switch v := x.(type) { case *T: ... }` |

`onerr` can be placed on a continuation line after a pipe chain:
```kukicha
result := fetch.Get(url)
    |> fetch.CheckStatus()
    |> fetch.Json(list of Repo)
    onerr return empty list of Repo
```

## Standard Library

Located in `stdlib/`:

| Package | Purpose |
|---------|---------|
| `a2a` | Agent-to-Agent protocol client (Discover, Ask, Send, Stream) |
| `cast` | Smart type coercion: any → scalar (SmartInt, SmartFloat64, SmartBool, SmartString) |
| `cli` | CLI argument parsing builder (New, String, Int, Bool, Parse) |
| `concurrent` | Concurrency helpers (Parallel, ParallelWithLimit) |
| `container` | Docker/Podman client (Connect, Run, Stop, Logs, Wait, Events, Exec, CopyFrom/To) |
| `ctx` | Context timeout/cancellation helpers (Background, WithTimeoutMs, WithDeadlineUnix, Cancel, Done, Err) |
| `datetime` | Named formats and durations (Format, Seconds, Minutes, Hours) |
| `encoding` | Base64 and hex encoding/decoding (Base64Encode, Base64Decode, HexEncode, HexDecode) |
| `env` | Typed environment variable access (Get, GetInt, GetBool, GetOr, Set) |
| `errors` | Error wrapping and inspection (Wrap, Is, Unwrap, New, Join) |
| `fetch` | HTTP client with builder, auth, retry, safe URL helpers (Get, Post, Json, New/Header/Timeout/Retry/Do, BearerAuth, URLTemplate, URLWithQuery) |
| `files` | File I/O operations (Read, Write, Append, Exists, Copy, Move, Delete, Watch) |
| `http` | HTTP response helpers (JSON, JSONError, JSONNotFound, ReadJSON, SafeURL) |
| `input` | User input utilities (Line, Confirm, Choose) |
| `iterator` | Functional iterators with Go 1.26+ generics (Filter, Map, Take, Skip, Reduce) |
| `json` | Pipe-friendly jsonv2 wrapper (Marshal, Unmarshal, MarshalWrite, UnmarshalRead) |
| `kube` | Kubernetes client via client-go (Connect, ListPods, ScaleDeployment, WatchPods, PodLogs) |
| `llm` | LLM chat client with retry (Ask/Send for OpenAI, MAsk/MSend for Anthropic, Retry) |
| `maps` | Map utilities (Keys, Values, Has, Merge) |
| `mcp` | Model Context Protocol server support (NewServer, Tool, Resource, Prompt) |
| `must` | Panic-on-error startup helpers (Env, EnvInt, EnvIntOr, Do, OkMsg) |
| `net` | IP address and CIDR utilities (ParseIP, ParseCIDR, Contains, IsPrivate, SplitHostPort) |
| `netguard` | Network restriction & SSRF protection (NewSSRFGuard, NewAllow, NewBlock, HTTPTransport) |
| `obs` | Structured observability helpers (New, Component, WithCorrelation, Info, Warn, Error, Start/Stop) |
| `parse` | Data format parsing (CSV, YAML, JsonLines, JsonPretty) |
| `pg` | PostgreSQL client via pgx (Connect, New/MaxConns/Retry/Open, Query, QueryRow, Exec, ScanRow, CollectRows) |
| `random` | Random number generation (Int, IntRange, Float, String, Choice) |
| `retry` | Manual retry helpers (New, Attempts, Delay, Sleep) |
| `sandbox` | os.Root filesystem sandboxing (New, Read, Write, List, Exists, Delete) |
| `shell` | Safe command execution builder (New, Dir, Env, Execute, Which, Output) |
| `slice` | Slice operations with generics (Filter, Map, GroupBy, FirstOr, LastOr, Find, Pop, GetOr) |
| `string` | String utilities (Split, Join, Trim, Contains, Replace, ToUpper, ToLower) |
| `template` | Text templating for code gen or reports (Execute, New) |
| `validate` | Input validation (Email, URL, InRange, NotEmpty, MinLen, MaxLen) |

Example with stdlib:
```kukicha
import "stdlib/fetch"
import "stdlib/slice"

# Fetch data with typed JSON decode
repos := fetch.Get("https://api.github.com/users/golang/repos")
    |> fetch.CheckStatus()
    |> fetch.Json(list of Repo) onerr panic "fetch failed"

# Filter with pipes — arrow lambda for concise predicates
activeRepos := repos |> slice.Filter((r Repo) => r.Stars > 100)
```

`fetch.Json(list of Repo)` and `fetch.Json(map of string to string)` are valid typed-empty forms.
For struct targets, use `fetch.Json(empty Repo)`.

### Parse Package (jsonv2 powered)
```kukicha
# Standard JSON parsing - use json.Unmarshal with parse.Json()
config := Config{}
jsonStr |> parse.Json() |> json.Unmarshal(_, reference of config) onerr panic "parse failed"

# NDJSON (newline-delimited JSON) parsing
lines := logData |> parse.JsonLines()  # Returns list of JSON strings
logs := list of LogEntry{}
for line in lines
    entry := LogEntry{}
    parseErr := json.Unmarshal(list of byte(line), reference of entry)
    if parseErr equals empty
        logs = append(logs, entry)

# Pretty-print JSON - delegates to json.MarshalPretty
output := config |> parse.JsonPretty()
```

### Concurrent Package
```kukicha
import "stdlib/concurrent"

# Run multiple tasks in parallel
concurrent.Parallel(task1, task2, task3)

# Limit concurrent execution (at most 4 tasks at once)
concurrent.ParallelWithLimit(4, many tasks)

# Run a function in a goroutine
concurrent.Go(myFunc)
```

### HTTP Package
```kukicha
import "stdlib/http"

# Add Cross-Origin protection to handler
handler := http.WithCSRF(myHandler)

# Start server
http.Serve(":8080", handler)
```

### Shell Package
```kukicha
import "stdlib/shell"

# Simple command
cmd := shell.New("ls", "-la")
result := shell.Execute(cmd)
if shell.Success(result)
    print(string(shell.GetOutput(result)))

# Command with options using builder pattern
cmd := shell.New("git", "status")
    |> shell.Dir("./repo")
    |> shell.SetTimeout(30)
result := shell.Execute(cmd)

if shell.Success(result)
    print(string(shell.GetOutput(result)))
else
    print("Error: {string(shell.GetError(result))}")

# Check if command exists
if shell.Which("docker")
    print("Docker is installed")

# Run with environment variables
cmd := shell.New("npm", "install")
    |> shell.Dir("./frontend")
    |> shell.Env("NODE_ENV", "production")
result := shell.Execute(cmd)
if shell.Success(result)
    print("npm install succeeded")
```

### CLI Package
```kukicha
import "stdlib/cli"

func main()
    app := cli.New("mytool")
        |> cli.Arg("command", "Command to run (fetch|process)")
        |> cli.Arg("input", "Input file or URL")
        |> cli.AddFlag("verbose", "Enable verbose output", "false")
        |> cli.AddFlag("format", "Output format", "json")
        |> cli.Action(handleCommand)

    cli.RunApp(app) onerr panic "command failed"

func handleCommand(args cli.Args)
    command := cli.GetString(args, "command")
    input := cli.GetString(args, "input")
    verbose := cli.GetBool(args, "verbose")
    format := cli.GetString(args, "format")

    switch command
        when "fetch"
            print("Fetching {input} with format {format}")
        when "process"
            print("Processing {input}")
        otherwise
            print("Unknown command: {command}")
```

### Fetch Package (Builder Pattern)
```kukicha
import "stdlib/fetch"

# Simple GET with typed JSON decode
repos := fetch.Get("https://api.github.com/users/golang/repos")
    |> fetch.CheckStatus()
    |> fetch.Json(list of Repo) onerr panic "{error}"

# Builder pattern with auth and retry
resp := fetch.New(url)
    |> fetch.BearerAuth(token)
    |> fetch.Timeout(30000000000)
    |> fetch.Retry(3, 500)
    |> fetch.Do() onerr panic "{error}"
text := fetch.Text(resp) onerr panic "{error}"

# Safe URL construction (avoids injection)
safeURL := fetch.URLTemplate("https://api.github.com/users/{username}/repos",
    map of string to string{"username": username}) onerr panic "{error}"
safeURL = fetch.URLWithQuery(safeURL, map of string to string{"per_page": "30"}) onerr panic "{error}"
```

### Obs Package (Structured Observability)
```kukicha
import "stdlib/obs"

logger := obs.New("myapp", "prod")
    |> obs.Component("api")
    |> obs.WithCorrelation(obs.NewCorrelationID())

logger |> obs.Info("request received", map of string to any{"path": "/users"})
logger |> obs.Warn("slow query", map of string to any{"ms": 450})
logger |> obs.Error("db failed", map of string to any{"err": "{dbErr}"})

# Span-style timing
logger |> obs.Start("deploy")
# ... work ...
logger |> obs.Stop("deploy")
```

### Ctx Package (Context Helpers)
```kukicha
import "stdlib/ctx"

c := ctx.Background() |> ctx.WithTimeout(30)
defer ctx.Cancel(c)

if ctx.Done(c)
    print("canceled: {ctx.Err(c)}")

# Use ctx-enabled stdlib ops for cancellable waits
kube.WaitDeploymentReadyCtx(cluster, c, "api") onerr panic "{error}"
```

### LLM Package
```kukicha
import "stdlib/llm"

# OpenAI-compatible (gpt-4o, etc.)
reply := llm.New("openai:gpt-4o-mini")
    |> llm.Retry(3, 2000)
    |> llm.Ask("Summarize this in one sentence.") onerr panic "{error}"

# Anthropic
reply := llm.NewMessages("claude-opus-4-6")
    |> llm.MRetry(3, 2000)
    |> llm.MAsk("Hello!") onerr panic "{error}"
```

### Pg Package (PostgreSQL)
```kukicha
import "stdlib/pg"

# Connect (with optional startup retry)
pool := pg.New(url) |> pg.Retry(5, 500) |> pg.Open() onerr panic "db: {error}"
defer pg.ClosePool(pool)

# Query rows
rows := pg.Query(pool, "SELECT name FROM users WHERE active = $1", true) onerr panic "{error}"
defer pg.Close(rows)
for pg.Next(rows)
    name := pg.ScanString(rows) onerr continue
    print(name)

# Single row
user := pg.QueryRow(pool, "SELECT id, name FROM users WHERE id = $1", id) onerr panic "{error}"
```

### Kube Package (Kubernetes)
```kukicha
import "stdlib/kube"

cluster := kube.New() |> kube.Retry(5, 1000) |> kube.Open() onerr panic "k8s: {error}"
pods := kube.Namespace(cluster, "default") |> kube.ListPods() onerr panic "{error}"
for pod in kube.Pods(pods)
    print("{kube.PodName(pod)}: {kube.PodStatus(pod)}")

kube.ScaleDeployment(cluster, "default", "api", 3) onerr panic "{error}"
kube.WaitDeploymentReady(cluster, "default", "api") onerr panic "{error}"
```

### Container Package (Docker/Podman)
```kukicha
import "stdlib/container"

engine := container.Connect() onerr panic "not running: {error}"
defer container.Close(engine)

container.Pull(engine, "alpine:latest") onerr panic "{error}"
id := container.Run(engine, "alpine:latest", list of string{"echo", "hello"}) onerr panic "{error}"
logs := container.Logs(engine, id) onerr panic "{error}"
print(logs)
code := container.Wait(engine, id, 60) onerr panic "{error}"
container.Remove(engine, id) onerr discard
```

### Slice Package with Generics
```kukicha
import "stdlib/slice"

type LogEntry
    level string
    message string

# Group log entries by level (automatically generates [T any, K comparable])
func getLevel(e LogEntry) string
    return e.level

entries := logs |> slice.GroupBy(getLevel)
# Result: map[string][]LogEntry with keys "ERROR", "WARN", "INFO", etc.

# GroupBy is Go 1.26+ generic - you write simple Kukicha code, transpiler handles the generics
```

### DevOps & SRE Patterns

Kukicha excels at infrastructure automation.

```kukicha
# 1. Resource Validation
"user@domain.com" |> validate.Email() onerr panic "invalid contact"
env.GetInt("REPLICA_COUNT") onerr 3 |> validate.InRange(1, 10) onerr panic

# 2. Resilient Retries
func deploy()
    cfg := retry.New() |> retry.Attempts(5)
    attempt := 0
    for attempt < cfg.MaxAttempts
        shell.New("kubectl", "apply", "-f", "manifest.yaml") |> shell.Execute() onerr
            retry.Sleep(cfg, attempt)
            attempt = attempt + 1
            continue
        return

# 3. Concurrent Health Checks
tasks := list of func(){}
for url in endpoints
    u := url
    tasks = append(tasks, func()
        fetch.Get(u) |> fetch.CheckStatus() onerr print "FAILED: {u}"
    )
concurrent.Parallel(tasks...)
```

## Transparent Go 1.26+ Generics

**Important:** You don't write generic syntax in Kukicha! The transpiler automatically generates proper Go generics for stdlib functions.

When you use `slice.GroupBy`, `iter.Map`, etc., the transpiler:
- Infers type parameters from your code (`T` for element type, `K` for key type)
- Applies proper constraints where needed (`K comparable` for map keys)
- Generates correct Go 1.26+ generic syntax
- All type safety benefits without the syntax burden

This is part of Kukicha's philosophy: **"It's Just Go"** - zero runtime overhead, full type safety, no learning curve for generics.

## Project Structure

- `petiole` = package (Go equivalent)
- `stem.toml` = project root config
- Package name auto-calculated from file path if not declared

```kukicha
petiole mypackage    # Optional - usually inferred from directory

import "fmt"
import "encoding/json"
import "stdlib/slice"
import "stdlib/ctx" as ctxpkg          # alias — use when name conflicts with local variable
import "stdlib/errors" as errs         # avoids clash with 'err' / 'errors' variables
import "stdlib/http" as httphelper     # avoids clash with 'net/http'
import "github.com/jackc/pgx/v5" as pgx  # external package with alias
```

**Common cases needing aliases:** `ctx`, `errors`, `json`, `container`, `string`.

## Common Patterns

### Struct Literals
```kukicha
# Use inline form (multiline struct literals don't parse)
todo := Todo{id: 1, title: "Learn Kukicha", completed: false}
```

### Defer
```kukicha
func ProcessFile(path string)
    file := os.Open(path) onerr panic "cannot open"
    defer file.Close()
    # ... process file
```

## Go 1.26+ Features

Kukicha leverages modern Go features:

- **encoding/json/v2**: 2-10x faster JSON parsing in `stdlib/json` and `stdlib/fetch`
- **testing/synctest**: Deterministic concurrency testing in compiler tests

## Architecture (for compiler work)

The compiler has 4 phases:

1. **Lexer** (`internal/lexer/`) - Tokenization with INDENT/DEDENT
2. **Parser** (`internal/parser/`) - AST building
3. **Semantic** (`internal/semantic/`) - Type checking, validation
4. **CodeGen** (`internal/codegen/`) - Go code generation

Key design decisions:
- **Signature-first inference**: Function params/returns need explicit types; local vars are inferred
- **Indentation-based**: 4-space blocks (no braces in canonical form)
- **Context-sensitive keywords**: `list`, `map`, `channel` are keywords only in type contexts

## Testing

```bash
go test ./...                        # Run all tests
go test ./internal/lexer/... -v      # Verbose lexer tests
go test ./... -cover                 # With coverage
```

## When Writing Kukicha Code

1. Always use explicit types for function parameters and returns
2. Use `onerr` for error handling instead of manual `if err != nil`
3. Prefer pipe operators for data transformation chains
4. Use English keywords (`and`, `or`, `not`, `equals`, `empty`)
5. Use 4-space indentation (tabs not allowed)
6. Use `reference of` and `dereference` instead of `&` and `*`
7. **Add `import "fmt"`** to any file using string interpolation inside `error ""` literals — the compiler generates `errors.New(fmt.Sprintf(...))` but does NOT auto-import `fmt`
8. Use `import "stdlib/pkg" as alias` when a package name conflicts with a local variable (common: `ctx`, `errors`, `json`, `container`, `string`)
9. Never edit `stdlib/*/*.go` directly — edit `.kuki` source files and run `make generate`
