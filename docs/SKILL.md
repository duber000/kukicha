## Writing Kukicha

Kukicha transpiles to Go. Write `.kuki` files with Kukicha syntax — not Go.

### Syntax vs Go

| Kukicha | Go |
|---------|-----|
| `and`, `or`, `not` | `&&`, `\|\|`, `!` |
| `equals` | `==` |
| `empty` | `nil` |
| `list of string` | `[]string` |
| `map of string to int` | `map[string]int` |
| `reference User` | `*User` |
| `reference of x` | `&x` |
| `dereference ptr` | `*ptr` |
| `func Method on t T` | `func (t T) Method()` |
| `many args` | `args...` |
| `make channel of T` | `make(chan T)` |
| `send val to ch` / `receive from ch` | `ch <- val` / `<-ch` |
| `defer f()` | `defer f()` (same keyword) |
| 4-space indentation | `{ }` braces |

### Variables and Functions

```kukicha
count := 42           # inferred type
count = 100           # reassignment

func Add(a int, b int) int
    return a + b

func Divide(a int, b int) int, error
    if b equals 0
        return 0, error "division by zero"
    return a / b, empty

# Default parameter value
func Greet(name string, greeting string = "Hello") string
    return "{greeting}, {name}!"

# Named argument at call site
result := Greet("Alice", greeting: "Hi")
files.Copy(from: src, to: dst)
```

### Strings and Interpolation

```kukicha
greeting := "Hello {name}!"          # {expr} is interpolated
json := "key: \{value\}"             # \{ and \} produce literal braces
path := "{dir}\sep{file}"            # \sep → OS path separator at runtime
```

### Types

```kukicha
type Repo
    name  string as "name"            # JSON field alias
    stars int    as "stargazers_count"
    tags  list of string
    meta  map of string to string
```

### Methods

```kukicha
func Display on todo Todo string
    return "{todo.id}: {todo.title}"

func SetDone on todo reference Todo       # pointer receiver
    todo.done = true
```

### Error Handling (`onerr`)

The caught error is always `{error}` — never `{err}`. Using `{err}` is a compile-time error. To use a custom name in a block handler, write `onerr as e`.

```kukicha
data := fetch.Get(url) onerr panic "failed: {error}"        # stop with message
data := fetch.Get(url) onerr return                         # propagate (shorthand — raw error, zero values)
data := fetch.Get(url) onerr return empty, error "{error}"  # propagate (verbose, wraps error)
port := getPort()      onerr 8080                           # default value
_    := riskyOp()      onerr discard                        # ignore
v    := parse(item)    onerr continue                       # skip iteration (inside for loop)
v    := parse(item)    onerr break                          # exit loop (inside for loop)
data := fetch.Get(url) onerr explain "context hint"         # wrap and propagate

# Block form — multiple statements
users := parse() onerr
    print("failed: {error}")
    return

# Block form with named alias
users := parse() onerr as e
    print("failed: {e}")    # {e} and {error} both work
    return
```

### Pipes

```kukicha
result := data |> parse() |> transform()

# _ placeholder: pipe into a non-first argument position
todo |> json.MarshalWrite(w, _)   # → json.MarshalWrite(w, todo)

# Bare identifier as target
data |> print                     # → fmt.Println(data)

# Pipeline-level onerr — catches errors from any step
items := fetch.Get(url)
    |> fetch.CheckStatus()
    |> fetch.Json(list of Repo)
    onerr panic "{error}"

# Piped switch — pipe a value into a switch
user.Role |> switch
    when "admin"
        grantAccess()
    when "guest"
        denyAccess()
    otherwise
        checkPermissions()
```

### Control Flow

```kukicha
if count equals 0
    return "empty"
else if count < 10
    return "small"

for item in items
    process(item)

for i from 0 to 10        # 0..9 (exclusive)
for i from 0 through 10   # 0..10 (inclusive)
for i from 10 through 0   # 10..0 (descending)

switch command
    when "fetch", "pull"
        fetchRepos()
    when "help"
        showHelp()
    otherwise
        print("Unknown: {command}")

# Bare switch (condition-based)
switch
    when stars >= 1000
        print("popular")
    otherwise
        print("new")

# Type switch
switch event as e
    when string
        print(e)
    when reference TaskEvent
        print(e.Status)
    otherwise
        print("unknown")
```

### Lambdas

Lambda parameter types are inferred from calling context — explicit annotations are optional.

```kukicha
# Inferred param type (preferred) — compiler resolves type from the list being piped
repos   |> slice.Filter(r => r.stars > 100)
entries |> sort.ByKey(e => e.name)

# Explicit param type (optional — only needed when inference can't determine the type)
repos |> slice.Filter((r Repo) => r.stars > 100)

# Zero params
button.OnClick(() => print("clicked"))

# Block lambda (multi-statement, explicit return)
repos |> slice.Filter(r =>
    name := r.name |> strpkg.ToLower()
    return name |> strpkg.Contains("go")
)

# sort.By — two params, both inferred
repos |> sort.By((a, b) => a.stars < b.stars)
```

### Collections

```kukicha
items  := list of string{"a", "b", "c"}
config := map of string to int{"port": 8080}
last   := items[-1]    # negative indexing
```

### Variadic Arguments (`many`)

```kukicha
func Sum(many numbers int) int
    total := 0
    for n in numbers
        total = total + n
    return total

result := Sum(1, 2, 3)
nums := list of int{1, 2, 3}
result := Sum(many nums)              # spread a slice
```

### Type Assertions

```kukicha
result, ok := value.(string)          # safe (two-value)
s := value.(string)                   # panics if wrong type
```

### Concurrency

```kukicha
ch := make channel of string
send "message" to ch
msg := receive from ch
go doWork()

# Multi-statement goroutine
go
    mu.Lock()
    doWork()
    mu.Unlock()

# Select (channel multiplexing)
select
    when receive from done
        return
    when msg := receive from ch
        print(msg)
    when send "ping" to out
        print("sent")
    otherwise
        print("nothing ready")
```

### Defer

```kukicha
defer resource.Close()                # runs when enclosing function exits
```

### Imports and Canonical Aliases

```kukicha
import "stdlib/slice"
import "stdlib/ctx"       as ctxpkg     # clashes with local 'ctx' variable
import "stdlib/errors"    as errs       # clashes with local 'err' / 'errors'
import "stdlib/json"      as jsonpkg    # clashes with 'encoding/json'
import "stdlib/string"    as strpkg     # clashes with 'string' type name
import "stdlib/container" as docker     # clashes with local 'container' variables
import "stdlib/http"      as httphelper # clashes with 'net/http'
import "stdlib/net"       as netutil    # clashes with 'net' package

import "github.com/jackc/pgx/v5" as pgx  # external package
```

Always use these aliases — clashes cause compile errors.

### Project Structure and Commands

```kukicha
petiole main                   # package declaration (Go's `package main`)
```

`petiole` is optional for single-file programs but required for multi-file packages and tests.

```bash
kukicha init [module]          # initialize project (go mod init + extract stdlib)
kukicha check file.kuki        # validate syntax without compiling
kukicha run file.kuki          # transpile, compile, and run
kukicha build file.kuki        # transpile and compile to binary
kukicha fmt -w file.kuki       # format in place
```

---

### Stdlib Packages

**stdlib/fetch** — HTTP requests

```kukicha
# Simple GET with typed JSON decode
repos := fetch.Get(url)
    |> fetch.CheckStatus()
    |> fetch.Json(list of Repo) onerr panic "{error}"

# fetch.Json sample arg tells the compiler what to decode into:
#   fetch.Json(list of Repo)            → JSON array  → []Repo
#   fetch.Json(empty Repo)              → JSON object → Repo
#   fetch.Json(map of string to string) → JSON object → map[string]string

# Builder: auth, timeout, retry
resp := fetch.New(url)
    |> fetch.BearerAuth(token)
    |> fetch.Retry(3, 500)
    |> fetch.Do() onerr panic "{error}"
text := fetch.Text(resp) onerr panic "{error}"

# SSRF-protected GET — use inside HTTP handlers or server code
resp := fetch.SafeGet(url) onerr panic "{error}"

# Cap response body size (prevent OOM)
resp := fetch.New(url) |> fetch.MaxBodySize(1 << 20) |> fetch.Do() onerr panic "{error}"

# Safe URL construction
url := fetch.URLTemplate("https://api.example.com/users/{id}",
    map of string to string{"id": userID}) onerr panic "{error}"
url  = fetch.URLWithQuery(url,
    map of string to string{"per_page": "30"}) onerr panic "{error}"
```

**stdlib/slice** — List operations

```kukicha
active  := slice.Filter(items, x => x.active)
names   := slice.Map(items, x => x.name)
byGroup := slice.GroupBy(items, x => x.category)
first   := slice.FirstOr(items, defaultVal)
val     := slice.GetOr(items, 0, defaultVal)
ok      := slices.Contains(items, value)   # note: slices (Go stdlib), not slice
```

**stdlib/files** — File I/O

```kukicha
data := files.Read("path.txt")        onerr panic "{error}"
       files.Write("out.txt", data)   onerr panic "{error}"
       files.Append("log.txt", line)  onerr discard
ok   := files.Exists("path.txt")
       files.Copy(from: src, to: dst) onerr panic "{error}"
```

**stdlib/cli** — Command-line apps

```kukicha
func run(args cli.Args)
    name := cli.GetString(args, "name")
    port := cli.GetInt(args, "port")

func main()
    app := cli.New("myapp")
        |> cli.AddFlag("name", "Your name", "world")
        |> cli.AddFlag("port", "Port number", "8080")
        |> cli.Action(run)
    cli.RunApp(app) onerr panic "{error}"
```

**stdlib/must** and **stdlib/env** — Config

```kukicha
import "stdlib/must"
# must: panics at startup if env var is missing (use for required config)
apiKey := must.Env("API_KEY")
port   := must.EnvIntOr("PORT", 8080)

import "stdlib/env"
# env: returns error via onerr (use for optional or runtime config)
debug  := env.GetBool("DEBUG") onerr false
token  := env.Get("TOKEN")     onerr panic "TOKEN required"
```

**stdlib/json** (import as `jsonpkg`) — JSON encode/decode

```kukicha
import "stdlib/json" as jsonpkg
data   := jsonpkg.Marshal(value)              onerr panic "{error}"
result := jsonpkg.Unmarshal(data, empty Repo) onerr panic "{error}"
```

**stdlib/mcp** — MCP server

```kukicha
func getPrice(symbol string) string
    return "GOOG: $180.00"

func main()
    server := mcp.NewServer()
    server |> mcp.Tool("get_price", "Get stock price by ticker", getPrice)
    server |> mcp.Serve()
```

**stdlib/shell** — Run commands

```kukicha
# Run: for fixed string literals only (no variable interpolation)
diff := shell.Run("git diff --staged") onerr panic "{error}"

# Output: use when any argument is a variable — args passed directly to OS
out := shell.Output("git", "log", "--oneline", userBranch) onerr panic "{error}"

# Builder: add working directory, env vars, timeout
result := shell.New("npm", "test") |> shell.Dir(projectPath) |> shell.Env("CI", "true") |> shell.Execute()
if not shell.Success(result)
    print(shell.GetError(result) as string)
```

**stdlib/obs** — Structured logging

```kukicha
logger := obs.New("myapp", "prod") |> obs.Component("worker")
logger |> obs.Info("starting", map of string to any{"job": "build"})
logger |> obs.Error("failed",  map of string to any{"err": err})
```

**stdlib/validate** — Input validation

```kukicha
email |> validate.Email()          onerr return
age   |> validate.InRange(18, 120) onerr return
name  |> validate.NotEmpty()       onerr return
```

**stdlib/parse** — Data parsing

```kukicha
rows := csvData  |> parse.CsvWithHeader() onerr panic "{error}"
cfg  := yamlData |> parse.Yaml()          onerr panic "{error}"
```

**stdlib/llm** — LLM calls

```kukicha
# OpenAI-compatible
reply := llm.New("openai:gpt-4o-mini")
    |> llm.Retry(3, 2000)
    |> llm.Ask("Hello!") onerr panic "{error}"

# Anthropic
reply := llm.NewMessages("claude-opus-4-6")
    |> llm.MRetry(3, 2000)
    |> llm.MAsk("Summarize this") onerr panic "{error}"
```

**stdlib/pg** — PostgreSQL

```kukicha
pool := pg.Connect(url) onerr panic "db: {error}"
defer pg.ClosePool(pool)
rows := pg.Query(pool, "SELECT name FROM users WHERE active = $1", true) onerr panic "{error}"
defer pg.Close(rows)
for pg.Next(rows)
    name := pg.ScanString(rows) onerr continue
```

**stdlib/http** (`import "stdlib/http" as httphelper`) — HTTP helpers + security

```kukicha
httphelper.JSON(w, data)                        # 200 OK with JSON body
httphelper.JSONCreated(w, data)                 # 201 Created
httphelper.JSONNotFound(w, "not found")         # 404
httphelper.JSONBadRequest(w, "bad input")       # 400
httphelper.JSONError(w, 500, "server error")    # any status

httphelper.ReadJSONLimit(r, 1<<20, reference of input) onerr return   # parse + size cap
httphelper.SafeHTML(w, userContent)             # HTML-escape before write
httphelper.SafeRedirect(w, r, url, "myapp.com") onerr return  # host-allowlist redirect
httphelper.SetSecureHeaders(w)                  # per-handler security headers
http.ListenAndServe(":8080", httphelper.SecureHeaders(mux))   # middleware form
```

**stdlib/template** — Templating

```kukicha
# text/template (no HTML escaping — for plain text only)
tmpl := template.New("t") |> template.Parse(src) onerr return
template.Execute(tmpl, data) onerr return

# html/template (auto-escapes {{ }} values — use for HTML responses)
html := template.HTMLRenderSimple(tmplStr, map of string to any{"name": username}) onerr return
```

**stdlib/string** (import as `strpkg`) — String utilities

```kukicha
import "stdlib/string" as strpkg
parts  := strpkg.Split(line, ",")
joined := strpkg.Join(parts, " | ")
lower  := strpkg.ToLower(name)
ok     := strpkg.Contains(text, "TODO")
clean  := strpkg.Replace(raw, "\t", " ")
```

**stdlib/errors** (import as `errs`) — Error wrapping

```kukicha
import "stdlib/errors" as errs
wrapped := errs.Wrap(err, "loading config")   # preserves Is/As chain
opaque  := errs.Opaque(pgErr, "db connect")   # breaks chain at boundaries
if errs.Is(err, io.EOF)
    return
# Dual-message: internal detail + user-safe message
e := errs.NewPublic("pg: refused 10.0.0.1", "database unavailable")
print(errs.Public(e))   # "database unavailable"
```

**stdlib/input** — Interactive CLI input

```kukicha
name := input.ReadLine("Enter name: ") onerr return
name := input.Prompt("Enter name: ")            # panics on stdin failure
ok   := input.Confirm("Proceed?") onerr return
idx  := input.Choose("Select:", options) onerr return
```

**stdlib/concurrent** — Parallel execution and concurrent map

```kukicha
# Run zero-argument functions concurrently, wait for all to finish
concurrent.Parallel(
    () => processChunk(chunkA),
    () => processChunk(chunkB),
)

# Same with a concurrency limit
concurrent.ParallelWithLimit(4,
    () => processChunk(chunkA),
    () => processChunk(chunkB),
)

# Transform every element in parallel, results in original order
results := concurrent.Map(urls, url => check(url))

# Same with a concurrency cap (useful for rate-limited APIs)
results := concurrent.MapWithLimit(repos, 4, r => fetchDetails(r))
```

**stdlib/datetime** — Time formatting and durations

```kukicha
formatted := datetime.Format(t, "iso8601")     # not Go's "2006-01-02"!
timeout   := datetime.Seconds(30)
```

**stdlib/encoding** — Base64 and hex

```kukicha
encoded := encoding.Base64Encode(data)
decoded := encoding.Base64Decode(encoded) onerr panic "{error}"
hex     := encoding.HexEncode(hashBytes)
```

**stdlib/retry** — Retry with backoff

```kukicha
cfg := retry.New() |> retry.Attempts(5) |> retry.Delay(200)
for attempt from 0 to cfg.MaxAttempts
    result, err := doWork()
    if err equals empty
        break
    retry.Sleep(cfg, attempt)
```

**stdlib/sandbox** — Filesystem sandboxing (use in HTTP handlers)

```kukicha
box     := sandbox.New("/var/data") onerr return
content := sandbox.Read(box, userPath) onerr return   # can't escape root
sandbox.Write(box, "out.txt", data) onerr return
```

**stdlib/math** — Math operations

```kukicha
val    := math.Clamp(input, 0.0, 100.0)
abs    := math.Abs(-5.0)
bigger := math.Max(a, b)
```

**stdlib/ctx** (import as `ctxpkg`) — Context helpers

```kukicha
import "stdlib/ctx" as ctxpkg
c := ctxpkg.Background() |> ctxpkg.WithTimeout(30)
defer ctxpkg.Cancel(c)
```

**stdlib/cast** — Type coercion

```kukicha
n := cast.SmartInt(value) onerr 0
s := cast.SmartString(value) onerr ""
```

**stdlib/maps** — Map utilities

```kukicha
keys := maps.Keys(config)
vals := maps.Values(config)
ok   := maps.Contains(config, "port")
```

**stdlib/random** — Random generation

```kukicha
token := random.String(32)
```

**stdlib/net** (import as `netutil`) — IP/CIDR utilities

```kukicha
import "stdlib/net" as netutil
ip      := netutil.ParseIP("192.168.1.100")
network := netutil.ParseCIDR("192.168.0.0/16") onerr panic "{error}"
if netutil.Contains(network, ip) and netutil.IsPrivate(ip)
    print("private range")
```

---

### Security — Compiler-Enforced Checks

The compiler **rejects** these patterns as errors (not warnings):

| Pattern | Error | Fix |
|---------|-------|-----|
| `pg.Query(pool, "... {var}")` | SQL injection | `pg.Query(pool, "... $1", val)` |
| `httphelper.HTML(w, nonLiteral)` | XSS risk | `httphelper.SafeHTML(w, content)` |
| `fetch.Get(url)` in HTTP handler | SSRF risk | `fetch.SafeGet(url)` |
| `files.Read(path)` in HTTP handler | Path traversal | `sandbox.Read(box, path)` |
| `shell.Run("cmd {var}")` | Command injection | `shell.Output("cmd", arg)` |
| `httphelper.Redirect(w, r, nonLiteral)` | Open redirect | `httphelper.SafeRedirect(w, r, url, "host")` |

HTTP handler detection: any function with an `http.ResponseWriter` parameter triggers the handler-context checks.

---

**stdlib/iterator** — Lazy iteration (Go 1.23 iter.Seq)

```kukicha
import "stdlib/iterator"
names := repos
    |> iterator.Values()
    |> iterator.Filter((r Repo) => r.Stars > 100)
    |> iterator.Map((r Repo) => r.Name)
    |> iterator.Take(5)
    |> iterator.Collect()
```

Functions: `Values`, `Filter`, `Map`, `FlatMap`, `Take`, `Skip`, `Enumerate`, `Chunk`, `Zip`, `Reduce`, `Collect`, `Any`, `All`, `Find`.

---

### Testing

Test files use the `*_test.kuki` suffix and the table-driven pattern:

```kukicha
petiole math_test

import "stdlib/math"
import "stdlib/test"
import "testing"

type ClampCase
    name string
    val  float64
    lo   float64
    hi   float64
    want float64

func TestClamp(t reference testing.T)
    cases := list of ClampCase{
        ClampCase{name: "within range", val: 5.0, lo: 0.0, hi: 10.0, want: 5.0},
        ClampCase{name: "below min", val: -5.0, lo: 0.0, hi: 10.0, want: 0.0},
    }
    for tc in cases
        t.Run(tc.name, (t reference testing.T) =>
            got := math.Clamp(tc.val, tc.lo, tc.hi)
            test.AssertEqual(t, got, tc.want)
        )
```

Assertions: `test.AssertEqual`, `test.AssertTrue`, `test.AssertFalse`, `test.AssertNoError`, `test.AssertError`, `test.AssertNotEmpty`.

---

**All available packages:** `a2a`, `cast`, `cli`, `concurrent`, `container`, `ctx`, `datetime`, `encoding`, `env`, `errors`, `fetch`, `files`, `http`, `input`, `iterator`, `json`, `kube`, `llm`, `maps`, `math`, `mcp`, `must`, `net`, `netguard`, `obs`, `parse`, `pg`, `random`, `retry`, `sandbox`, `shell`, `slice`, `string`, `template`, `test`, `validate`
