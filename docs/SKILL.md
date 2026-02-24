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

### Types

```kukicha
type Repo
    name  string as "name"            # JSON field alias
    stars int    as "stargazers_count"
    tags  list of string
    meta  map of string to string
```

### Error Handling (`onerr`)

The caught error is always `{error}` — never `{err}`. Using `{err}` is a compile-time error. To use a custom name in a block handler, write `onerr as e`.

```kukicha
data := fetch.Get(url) onerr panic "failed: {error}"        # stop with message
data := fetch.Get(url) onerr return                         # propagate (shorthand — raw error, zero values)
data := fetch.Get(url) onerr return empty, error "{error}"  # propagate (verbose, wraps error)
port := getPort()      onerr 8080                           # default value
_    := riskyOp()      onerr discard                        # ignore
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

# onerr on a pipe chain
items := fetch.Get(url)
    |> fetch.CheckStatus()
    |> fetch.Json(list of Repo)
    onerr panic "{error}"
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
```

### Lambdas

```kukicha
# Expression lambda (auto-return)
repos |> slice.Filter((r Repo) => r.stars > 100)

# Single untyped param
numbers |> slice.Filter(n => n > 0)

# Block lambda (multi-statement, explicit return)
repos |> slice.Filter((r Repo) =>
    name := r.name |> strpkg.ToLower()
    return name |> strpkg.Contains("go")
)
```

### Collections

```kukicha
items  := list of string{"a", "b", "c"}
config := map of string to int{"port": 8080}
last   := items[-1]    # negative indexing
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
active  := slice.Filter(items, (x Item) => x.active)
names   := slice.Map(items, (x Item) => x.name)
byGroup := slice.GroupBy(items, (x Item) => x.category)
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
email |> validate.Email()          onerr return error "{error}"
age   |> validate.InRange(18, 120) onerr return error "{error}"
name  |> validate.NotEmpty()       onerr return error "{error}"
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

**All available packages:** `a2a`, `cast`, `cli`, `concurrent`, `container`, `ctx`, `datetime`, `encoding`, `env`, `errors`, `fetch`, `files`, `http`, `input`, `iterator`, `json`, `kube`, `llm`, `math`, `maps`, `mcp`, `must`, `net`, `netguard`, `obs`, `parse`, `pg`, `random`, `retry`, `sandbox`, `shell`, `slice`, `string`, `template`, `validate`
