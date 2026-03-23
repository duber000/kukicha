# stdlib/CLAUDE.md

Kukicha standard library reference. Each package lives in `stdlib/<name>/` with:
- `<name>.kuki` — Kukicha source (types, function signatures, inline implementations)
- `<name>.go` — **Generated** by `make generate` from the `.kuki` file. Never edit directly.

Import with: `import "stdlib/slice"`

## Packages

| Package | Purpose | Key Functions |
|---------|---------|---------------|
| `stdlib/a2a` | Agent-to-Agent protocol client | Discover, Ask, Send, Stream, New/Text/Context |
| `stdlib/cast` | Smart type coercion (any → scalar) | SmartInt, SmartFloat64, SmartBool, SmartString, Atoi, ParseFloat |
| `stdlib/cli` | CLI argument parsing with subcommands | New, Description, Arg, AddFlag, Action, RunApp, Command, CommandFlag, CommandAction, GlobalFlag, CommandName, GetString, GetBool, GetInt, NewArgs, IsJSON |
| `stdlib/concurrent` | Parallel execution and concurrent map | Parallel, ParallelWithLimit, Map, MapWithLimit, Go |
| `stdlib/container` | Docker/Podman client via Docker SDK | Connect, ConnectRemote, New/Host/APIVersion/Open, ListContainers, ListImages, Pull, PullAuth, LoginFromConfig, Run, Stop, Remove, Build, Logs, LogsTail, Inspect, Wait/WaitCtx, Exec, Events/EventsCtx, CopyFrom, CopyTo |
| `stdlib/crypto` | Hashing, HMAC, and secure random (Go stdlib only) | SHA256, SHA256Bytes, HMAC, HMACBytes, RandomToken, RandomBytes, Equal |
| `stdlib/ctx` | Context timeout/cancellation helpers | Background, WithTimeout, WithTimeoutMs, WithDeadlineUnix, Cancel, Done, Err, Value |
| `stdlib/datetime` | Named formats, duration helpers, arithmetic, comparison | Format, Parse, Now, Today, AddDays, IsBefore, Unix, Sleep; Constants: ISO8601, RFC3339, Date, Time, DateTime |
| `stdlib/encoding` | Base64 and hex encoding/decoding | Base64Encode, Base64Decode, Base64URLEncode, Base64URLDecode, Base64RawEncode, Base64RawURLEncode, HexEncode, HexDecode |
| `stdlib/env` | Typed env vars with onerr | Get, GetOr, GetInt, GetIntOrDefault, GetBool, GetBoolOrDefault, GetFloat, GetList, Set, Unset, IsSet, All |
| `stdlib/errors` | Error wrapping and inspection | Wrap, Opaque, Is, Unwrap, New, Join, NewPublic, Public |
| `stdlib/fetch` | HTTP client (Builder, Auth, Sessions, Safe URL helpers, Retry) | Get, SafeGet, Post, Json, Decode, Text, Bytes, CheckStatus, URLTemplate, URLWithQuery, PathEscape, QueryEscape, New/Header/Timeout/Retry/MaxBodySize/Transport/Do, BearerAuth, BasicAuth, FormData, NewSession, DownloadTo |
| `stdlib/files` | File I/O operations | Read, ReadBytes, Write, WriteString, Append, AppendString, Exists, IsDir, IsFile, Copy, Move, Delete, DeleteAll, List, ListRecursive, MkDir, MkDirAll, TempFile, TempDir, Size, ModTime, Basename, Dirname, Extension, Join, Abs, UseWith, Watch |
| `stdlib/git` | Git/GitHub operations via gh CLI | ListTags, TagExists, DefaultBranch, CurrentBranch, ReleaseExists, CreateRelease, PreviewRelease, RepoExists, CurrentUser, Clone, CloneShallow |
| `stdlib/http` | HTTP response/request helpers + security | JSON, JSONStatus, JSONCreated, JSONError, JSONBadRequest, JSONNotFound, Text, HTML, SafeHTML, ReadJSON, ReadJSONLimit, Redirect, SafeRedirect, SafeURL, SetSecureHeaders, SecureHeaders, WithCSRF, Serve, MethodNotAllowed, IsGet/IsPost/IsPut/IsDelete/IsPatch, GetQueryParam, GetHeader; Constants: StatusOK/NotFound/etc, HeaderContentType, ContentJSON |
| `stdlib/input` | User input utilities | ReadLine, Prompt, Confirm, Choose |
| `stdlib/iterator` | Functional iteration (Go 1.23 iter.Seq) | Values, Filter, Map, FlatMap, Take, Skip, Enumerate, Chunk, Zip, Reduce, Collect, Any, All, Find |
| `stdlib/json` | encoding/json wrapper | Marshal, MarshalPretty, Unmarshal, MarshalWrite, UnmarshalRead, DecodeRead, NewEncoder, NewDecoder, Encode, Decode, WithDeterministic, WithIndent, WriteOutput |
| `stdlib/kube` | Kubernetes client via client-go | Connect, New/Kubeconfig/Context/InCluster/Retry/Open, Namespace, ListPods, ListPodsLabeled, GetPod, DeletePod, PodLogs, PodLogsTail, ListDeployments, GetDeployment, ScaleDeployment, RolloutRestart, DeleteDeployment, WaitDeploymentReady/WaitDeploymentReadyCtx, WaitPodReady/WaitPodReadyCtx, WatchPods/WatchPodsCtx, ListServices, ListNodes, ListNamespaces |
| `stdlib/llm` | Large language model client (Chat Completions, OpenResponses, Anthropic; Retry) | New/Ask/Send/SendRaw/Complete, NewResponse/RAsk/RSend/Respond, NewMessages/MAsk/MSend/AnthropicComplete, Retry/RRetry/MRetry, Stream/RStream/MStream |
| `stdlib/maps` | Map utilities | Keys, Values, Contains, Has, Merge, SortedKeys |
| `stdlib/mcp` | Model Context Protocol server | New, Serve, Tool, Prop, Schema, Required, TextResult, ErrorResult |
| `stdlib/must` | Panic-on-error startup helpers | Do, DoMsg, Ok, OkMsg, Env, EnvOr, EnvInt, EnvIntOr, EnvBool, EnvBoolOr, EnvList, EnvListOr, True, False, NotEmpty, NotNil |
| `stdlib/net` | IP address and CIDR utilities | ParseIP, ParseCIDR, Contains, SplitHostPort, JoinHostPort, LookupHost, IsLoopback, IsPrivate, IsMulticast, IsNil, IPString |
| `stdlib/netguard` | Network restriction & SSRF protection | NewSSRFGuard, NewAllow, NewBlock, Check, DialContext, HTTPTransport, HTTPClient |
| `stdlib/obs` | Structured observability helpers | New, Component, WithCorrelation, NewCorrelationID, Debug, Info, Warn, Error, Log, Start, Stop, Fail |
| `stdlib/parse` | Data format parsing | Json, JsonLines, JsonPretty, Csv, CsvWithHeader, Yaml, YamlPretty |
| `stdlib/pg` | PostgreSQL client via pgx | Connect, New/MaxConns/MinConns/MaxConnLifetime/MaxConnIdleTime/Retry/Open, Query, QueryRow, Exec, Begin, Commit, Rollback, Scan, ScanString, ScanInt, ScanInt64, ScanBool, ScanFloat64, ScanRow, CollectRows, Next, Close, ClosePool, RowsAffected |
| `stdlib/random` | Random string generation | String, Alphanumeric |
| `stdlib/regex` | Regular expression matching and replacement | Match, Find, FindAll, FindGroups, FindAllGroups, Replace, ReplaceFunc, Split, IsValid, Compile, MustCompile + compiled variants |
| `stdlib/retry` | Retry with backoff | New, Attempts, Delay, Linear, Sleep |
| `stdlib/sandbox` | os.Root filesystem sandboxing | New, Close, Read, ReadString, Write, WriteString, Append, AppendString, MkDir, MkDirAll, List, Exists, IsDir, IsFile, Stat, Delete, DeleteAll, Rename, Path, FS |
| `stdlib/semver` | Semantic versioning (parse, bump, compare) | Parse, Bump, Format, Valid, Compare, Greater, Highest |
| `stdlib/shell` | Safe command execution | Run, Output, New/Dir/SetTimeout/Env/Execute, Args/FlagIf/Preview, Success, GetOutput, GetError, ExitCode, Which, Getenv, Setenv, Unsetenv, Environ |
| `stdlib/skills` | Runtime discovery of agent SKILL.md manifests | Discover, AgentSkills, ClaudeSkills |
| `stdlib/slice` | Slice operations (all generic) | Filter, Map, GroupBy, Sort, SortBy, First, Last, Drop, DropLast, Reverse, Unique, Chunk, Contains, IndexOf, Concat, Get, GetOr, FirstOne, FirstOr, LastOne, LastOr, Find, FindOr, FindIndex, FindLast, FindLastOr, IsEmpty, IsNotEmpty, Pop, Shift |
| `stdlib/sort` | Sorting slices (strings, ints, floats, custom) | Strings, Ints, Float64s, By, ByKey, Reverse |
| `stdlib/string` | String utilities | ToUpper, ToLower, Title, Trim, TrimSpace, TrimPrefix, TrimSuffix, TrimLeft, TrimRight, Split, SplitN, Join, Fields, Contains, HasPrefix, HasSuffix, Index, LastIndex, Count, Replace, ReplaceAll, Repeat, PadRight, PadLeft, Concat, EqualFold, Len, IsEmpty, IsBlank, Lines |
| `stdlib/table` | Terminal table rendering (plain, box, markdown) | New, AddRow, Print, PrintWithStyle, ToString, ToStringWithStyle |
| `stdlib/template` | Text templating (plain + HTML-safe) | New, Render, Parse, Data, WithContent, Execute, RenderSimple, HTMLExecute, HTMLRenderSimple, Must, Funcs |
| `stdlib/test` | Test assertion helpers (use in `*_test.kuki` only) | AssertEqual, AssertNotEqual, AssertTrue, AssertFalse, AssertNoError, AssertError, AssertNotEmpty, AssertNil, AssertNotNil |
| `stdlib/validate` | Input validation | NotEmpty, MinLength, MaxLength, Length, LengthBetween, Matches, Email, URL, Alpha, Alphanumeric, Numeric, NoWhitespace, StartsWith, EndsWith, Contains, OneOf, Positive, Negative, NonNegative, NonZero, InRange, Min, Max, PositiveFloat, InRangeFloat, ParseInt, ParsePositiveInt, ParseFloat, ParseBool, NotEmptyList, ListMinLength, ListMaxLength, WithMessage, Require, NoHTML, SafeFilename, NoNullBytes |

## Testing Stdlib Packages

Use the **table-driven pattern** for all `*_test.kuki` files. This produces self-describing failure messages (`TestFirst/3_elements` instead of a bare `t.Errorf`) and makes adding new cases trivial.

```kukicha
petiole slice_test

import "stdlib/slice"
import "stdlib/test"
import "testing"

type FirstCase
    name    string
    n       int
    wantLen int

func TestFirst(t reference testing.T)
    items := list of string{"a", "b", "c", "d", "e"}
    cases := list of FirstCase{
        FirstCase{name: "3 elements", n: 3, wantLen: 3},
        FirstCase{name: "n > length", n: 10, wantLen: 5},
        FirstCase{name: "n=0", n: 0, wantLen: 0},
    }
    for tc in cases
        t.Run(tc.name, (t reference testing.T) =>
            result := slice.First(items, tc.n)
            test.AssertEqual(t, len(result), tc.wantLen)
        )
```

**Conventions:**
- Case types are declared at file scope, named `<FunctionName>Case`; `name string` is always the first field
- `t.Run(tc.name, (t reference testing.T) => ...)` wraps every assertion body
- Use `test.AssertEqual`, `test.AssertNoError`, `test.AssertError` in preference to bare `t.Errorf`
- A comment `# --- TestFoo ---` separates each function's table from the next
- Import `stdlib/test` only in `*_test.kuki` files, never in library code

## Common Patterns

```kukicha
# Semver (parse, bump, compare)
import "stdlib/semver"
v := semver.Parse("v1.2.3") onerr panic "{error}"
next := v |> semver.Bump("minor") |> semver.Format()   # "v1.3.0"
if semver.Valid("v2.0.0")
    print("valid")
best := semver.Highest(tags) onerr panic "{error}"

# CLI with subcommands
import "stdlib/cli"
_ = cli.New("mytool")
    |> cli.Description("A useful tool")
    |> cli.GlobalFlag("verbose", "Enable verbose output", "false")
    |> cli.Command("list", "List all items")
    |> cli.CommandFlag("list", "format", "Output format", "table")
    |> cli.CommandAction("list", doList)
    |> cli.Command("add", "Add a new item")
    |> cli.CommandAction("add", doAdd)
    |> cli.RunApp() onerr fatal("{error}")

# Validation (returns error for onerr)
import "stdlib/validate"
email |> validate.Email() onerr return
age |> validate.InRange(18, 120) onerr return

# Startup config (panics if missing/invalid)
import "stdlib/must"
apiKey := must.Env("API_KEY")
port := must.EnvIntOr("PORT", 8080)

# Runtime config (returns error for onerr)
import "stdlib/env"
debug := env.GetBoolOrDefault("DEBUG", false)

# Interactive user input (CLI scripts)
import "stdlib/input"
# ReadLine: read a line with optional prompt (returns error)
name := input.ReadLine("Enter name: ") onerr return
# Prompt: panics on error — for simple scripts where stdin failure is fatal
name := input.Prompt("Enter name: ")
# Confirm: yes/no prompt — returns false on 'n', error only if stdin fails
ok := input.Confirm("Proceed?") onerr return
if not ok
    return
# Choose: numbered menu, returns 0-based index; error on cancel or bad input
repos := list of string{"myorg/api", "myorg/web"}
i := input.Choose("Select a repo:", repos) onerr
    print("Cancelled.")
    return
print("You chose: {repos[i]}")

# Structured logs with correlation IDs
import "stdlib/obs"
logger := obs.New("deployctl", "prod") |> obs.Component("rollout")
logger = logger |> obs.WithCorrelation(obs.NewCorrelationID())
logger |> obs.Info("starting deployment", map of string to any{"app": "billing"})

# Context timeout helpers
import "stdlib/ctx"
c := ctx.Background() |> ctx.WithTimeout(30)
defer ctx.Cancel(c)
if ctx.Done(c)
    print("request canceled: {ctx.Err(c)}")
# Use ctx-enabled operations for cancellable waits/streams
kube.WaitDeploymentReadyCtx(cluster, c, "api") onerr panic "{error}"
container.EventsCtx(engine, c) onerr panic "{error}"

# HTTP responses
import "stdlib/http" as httphelper
httphelper.JSON(w, data)
httphelper.JSONNotFound(w, "User not found")

# Time formatting
import "stdlib/datetime"
datetime.Format(t, datetime.ISO8601)   # Use named constants instead of raw strings
datetime.Format(t, datetime.Date)      # "2006-01-02"
datetime.Format(t, "iso8601")          # String names still work
timeout := datetime.Seconds(30)

# PostgreSQL
import "stdlib/pg"
pool := pg.Connect(url) onerr panic "db: {error}"
defer pg.ClosePool(pool)
rows := pg.Query(pool, "SELECT name FROM users WHERE active = $1", true) onerr panic "{error}"
defer pg.Close(rows)
for pg.Next(rows)
    name := pg.ScanString(rows) onerr continue

# Kubernetes
import "stdlib/kube"
cluster := kube.Connect() onerr panic "k8s: {error}"
pods := kube.Namespace(cluster, "default") |> kube.ListPods() onerr panic "{error}"
for pod in kube.Pods(pods)
    print("{kube.PodName(pod)}: {kube.PodStatus(pod)}")
# Collect pod events for 20 seconds
events := kube.WatchPods(kube.Namespace(cluster, "default"), 20) onerr panic "{error}"
for event in events
    print("{kube.PodEventType(event)} {kube.PodEventName(event)} ready={kube.PodEventReady(event)}")
# For apply/patch workflows, prefer GitOps tools (e.g., Argo CD) and use kube stdlib
# for operational reads, rollout actions, and watches.

# Retry on transient failures (fetch: 429/503 + network errors)
import "stdlib/fetch"
resp := fetch.New(url) |> fetch.BearerAuth(token) |> fetch.Retry(3, 500) |> fetch.Do() onerr panic "{error}"
text := fetch.Text(resp) onerr panic "{error}"

# LLM with retry on rate limits
import "stdlib/llm"
reply := llm.New("openai:gpt-4o-mini") |> llm.Retry(3, 2000) |> llm.Ask("Hello!") onerr panic "{error}"
# Anthropic with retry
reply := llm.NewMessages("claude-opus-4-6") |> llm.MRetry(3, 2000) |> llm.MAsk("Hello!") onerr panic "{error}"

# PostgreSQL with startup retry (database may not be ready yet)
import "stdlib/pg"
pool := pg.New(url) |> pg.Retry(5, 500) |> pg.Open() onerr panic "db: {error}"

# Kubernetes with startup retry
import "stdlib/kube"
cluster := kube.New() |> kube.Retry(5, 1000) |> kube.Open() onerr panic "k8s: {error}"

# Concurrent map — transform every element in parallel, ordered results
import "stdlib/concurrent"
results := concurrent.Map(urls, url => check(url))

# With concurrency cap (useful for rate-limited APIs)
results := concurrent.MapWithLimit(repos, 4, r => fetchDetails(r))

# Iterator-based pipelines (lazy evaluation via Go 1.23 iter.Seq)
import "stdlib/iterator"
names := repos
    |> iterator.Values()
    |> iterator.Filter(r => r.Stars > 100)
    |> iterator.Map(r => r.Name)
    |> iterator.Collect()

# Take first 5 results lazily
top5 := items
    |> iterator.Values()
    |> iterator.Filter(x => x.Active)
    |> iterator.Take(5)
    |> iterator.Collect()

# Piped switch — pipe a value into a switch expression (wraps in IIFE)
user.Role |> switch
    when "admin"
        grantAccess()
    when "guest"
        denyAccess()
    otherwise
        checkPermissions()

# Pipeline-level onerr — catches errors from any step in a pipe chain
processed := data
    |> parse.Json(list of User)
    |> fetch.EnrichWithDB()
    |> validate.Safe()
    onerr panic "pipeline failed: {error}"

# Bidirectional Loops
# Use 'through' to iterate in either direction (ascending or descending).
# The compiler handles the comparison logic automatically.
for i from 10 through 0
    print("Countdown: {i}")

# Manual retry loop (for custom retry conditions)
import "stdlib/retry"
cfg := retry.New() |> retry.Attempts(5) |> retry.Delay(200)
for attempt from 0 to cfg.MaxAttempts
    result, err := doSomething()
    if err == empty
        break
    retry.Sleep(cfg, attempt)

# HTTP fetch with builder
resp := fetch.New(url) |> fetch.BearerAuth(token) |> fetch.Timeout(30000000000) |> fetch.Do() onerr panic "{error}"
text := fetch.Text(resp) onerr panic "{error}"

# Typed JSON decode (readable API flow)
# fetch.Json takes a typed zero value — the compiler uses it to infer the decode target type:
#   fetch.Json(list of Repo)           → decodes JSON array into []Repo
#   fetch.Json(empty Repo)             → decodes JSON object into Repo
#   fetch.Json(map of string to string) → decodes JSON object into map[string]string
repos := fetch.Get(url) |> fetch.CheckStatus() |> fetch.Json(list of Repo) onerr panic "{error}"

# Safe URL construction (path + query encoding)
base := fetch.URLTemplate("https://api.github.com/users/{username}/repos", map of string to string{"username": username}) onerr panic "{error}"
safeURL := fetch.URLWithQuery(base, map of string to string{"per_page": "30", "sort": "stars"}) onerr panic "{error}"

# Network-restricted fetch (SSRF protection)
# Preferred: fetch.SafeGet wraps netguard automatically — use in any HTTP handler
import "stdlib/fetch"
resp := fetch.SafeGet(url) onerr panic "{error}"
# Builder pattern: add headers/retry and still get SSRF protection
import "stdlib/netguard"
guard := netguard.NewSSRFGuard()
resp := fetch.New(url) |> fetch.Transport(netguard.HTTPTransport(guard)) |> fetch.Retry(3, 500) |> fetch.Do() onerr panic "{error}"

# Container management (Docker/Podman)
import "stdlib/container"
engine := container.Connect() onerr panic "not running: {error}"
defer container.Close(engine)
images := engine |> container.ListImages() onerr panic "{error}"
for img in images
    print("{container.ImageID(img)}: {container.ImageTags(img)}")

# Pull and run a container
container.Pull(engine, "alpine:latest") onerr panic "{error}"
id := container.Run(engine, "alpine:latest", list of string{"echo", "hello"}) onerr panic "{error}"
logs := container.Logs(engine, id) onerr panic "{error}"
print(logs)
code := container.Wait(engine, id, 60) onerr panic "{error}"
print("exit code: {code}")
events := container.Events(engine, 5) onerr panic "{error}"
for event in events
    print("{container.EventTime(event)} {container.EventAction(event)} {container.EventID(event)}")
container.Remove(engine, id) onerr discard

# IP address and CIDR utilities
import "stdlib/net" as netutil
ip := netutil.ParseIP("192.168.1.100")
if netutil.IsNil(ip)
    panic("invalid IP")
network := netutil.ParseCIDR("192.168.0.0/16") onerr panic "{error}"
if netutil.Contains(network, ip)
    print("in private range")
if netutil.IsPrivate(ip)
    print("private address")
host, port, err := netutil.SplitHostPort("example.com:8080") onerr panic "{error}"

# Shell command execution
import "stdlib/shell"
# Run: for fixed string literals only — splits on whitespace, no quoting awareness
diff := shell.Run("git diff --staged") onerr return
# Output: use when any argument comes from user input or a variable — args passed
# directly to the OS, no shell involved, so metacharacters are never interpreted
out := shell.Output("git", "log", "--oneline", userBranch) onerr return
# Builder pattern: add working directory, env vars, or timeout
result := shell.New("npm", "test") |> shell.Dir(projectPath) |> shell.Env("CI", "true") |> shell.Execute()
if not shell.Success(result)
    print(shell.GetError(result) as string)
# Fluent command builder with conditional flags
cmd := shell.New("gh", "release", "create", "v1.0.0", "--repo", "org/repo")
    |> shell.FlagIf(isDraft, "--draft")
    |> shell.FlagIf(hasTarget, "--target", "main")
    |> shell.Args("--title", "v1.0.0")
print("Running: {shell.Preview(cmd)}")
result := cmd |> shell.Execute()

# Git/GitHub operations (requires gh CLI)
import "stdlib/git"
tags := git.ListTags("owner/repo") onerr panic "{error}"
branch := git.DefaultBranch("owner/repo") onerr panic "{error}"
exists, _ := git.TagExists("owner/repo", "v1.0.0")
me := git.CurrentUser() onerr panic "{error}"
# Create a release with options
opts := git.ReleaseOptions{Title: "v1.0.0", Target: "main", Draft: true, GenerateNotes: true}
git.CreateRelease("owner/repo", "v1.0.0", opts) onerr panic "{error}"
# Dry-run: preview the command without executing
print("Would run: {git.PreviewRelease("owner/repo", "v1.0.0", opts)}")

# Regular expressions
import "stdlib/regex"
if regex.Match("\\d+", text)
    print("contains a number")
groups := regex.FindGroups("^(v?)(\\d+\\.\\d+\\.\\d+)$", tag) onerr panic "{error}"
prefix := groups[1]
version := groups[2]
cleaned := regex.Replace("\\s+", " ", messy)
# Compiled patterns for hot paths
p := regex.MustCompile("\\d+")
nums := regex.FindAllCompiled(p, "a1 b2 c3")

# Error wrapping and inspection
import "stdlib/errors"
err := errors.Wrap(originalErr, "loading config")
# err.Error() == "loading config: <original message>"
if errors.Is(err, io.EOF)
    print("end of file")

# Opaque wrap — breaks errors.Is/As chain at subsystem boundaries
# Use when crossing DB/infra boundaries to prevent internal type leakage
dbErr := errors.Opaque(pgxErr, "pg connect")  # callers cannot errors.As into pgx

# Dual-message errors — separate internal detail from user-safe message
# Log e.Error() internally; return errors.Public(e) in HTTP responses
e := errors.NewPublic("pg: connection refused to 10.0.0.1:5432", "database unavailable")
print(e.Error())           # "pg: connection refused to 10.0.0.1:5432" — log this
print(errors.Public(e))    # "database unavailable" — safe to return to users
# Falls back to "an error occurred" for non-PublicError errors
print(errors.Public(plainErr))  # "an error occurred"

# Hashing, HMAC, and secure random
import "stdlib/crypto"
hash := crypto.SHA256("hello world")                 # hex-encoded SHA-256
mac := crypto.HMAC("secret-key", "message-body")    # hex-encoded HMAC-SHA256
token, err := crypto.RandomToken(32) onerr panic "{error}"  # 64-char hex token
if crypto.Equal(expected, actual)
    print("match")

# Sorting slices
import "stdlib/sort"
sorted := sort.Strings(list of string{"banana", "apple", "cherry"})
nums := sort.Ints(list of int{3, 1, 4, 1, 5})
byLen := sort.By(words, (a, b) => len(a) < len(b))
byName := sort.ByKey(repos, r => r.Name)
reversed := sort.Reverse(words, (a, b) => len(a) < len(b))

# Convenience sort via slice package (pipe-friendly)
import "stdlib/slice"
sorted := repos |> slice.Sort((a, b) => a.Stars < b.Stars)
sorted := repos |> slice.SortBy(r => r.Name)

# Deterministic map key iteration
import "stdlib/maps"
keys := maps.SortedKeys(config)    # sorted string keys for deterministic output

# Terminal tables (plain, box, markdown)
import "stdlib/table"
tbl := table.New(list of string{"Name", "Stars"})
tbl = tbl |> table.AddRow(list of string{"go", "115000"})
tbl = tbl |> table.AddRow(list of string{"rust", "97000"})
table.Print(tbl)                                     # plain output
table.PrintWithStyle(tbl, "markdown")                # markdown table
s := table.ToString(tbl)                             # as string
s2 := table.ToStringWithStyle(tbl, "box")            # box-drawing style

# Base64 and hex encoding
import "stdlib/encoding"
encoded := encoding.Base64Encode("hello" as list of byte)
decoded := encoding.Base64Decode(encoded) onerr panic "invalid base64: {error}"
hexStr := encoding.HexEncode(hashBytes)

# MCP server
import "stdlib/mcp"
server := mcp.New("my-tool", "1.0.0")
schema := mcp.Schema(list of mcp.SchemaProperty{
    mcp.Prop("query", "string", "The search query"),
}) |> mcp.Required(list of string{"query"})
mcp.Tool(server, "search", "Search for items", schema, handler)
mcp.Serve(server) onerr panic "{error}"

# A2A client
import "stdlib/a2a"
agent := a2a.Discover("https://agent.example.com") onerr panic "{error}"
reply := a2a.Ask(agent, "What's the weather?") onerr panic "{error}"
print(reply)

# Sandbox (restricted filesystem)
import "stdlib/sandbox"
box := sandbox.New("/var/data") onerr panic "{error}"
defer sandbox.Close(box)
content := sandbox.Read(box, "config.json") onerr panic "{error}"
sandbox.WriteString(box, "hello", "output.txt") onerr panic "{error}"

# Template rendering
import "stdlib/template"
result := template.RenderSimple("Hello {{.Name}}!", map of string to any{"Name": "World"}) onerr panic "{error}"
# HTML-safe rendering (auto-escapes values)
safe := template.HTMLRenderSimple(tmplStr, data) onerr panic "{error}"
```

## Security Patterns

The compiler enforces several security checks. Use the safe alternatives below to avoid compile errors.

### Security Check Table

| Category | Unsafe (compiler error) | Safe alternative |
|----------|------------------------|------------------|
| **XSS** | `http.HTML(w, userInput)` | `http.SafeHTML(w, userInput)` or `template.HTMLRenderSimple(...)` |
| **SQL Injection** | `pg.Query(pool, "... '{name}'")` | `pg.Query(pool, "... $1", name)` |
| **SSRF** | `fetch.Get(url)` (in HTTP handler) | `fetch.SafeGet(url)` |
| **Open Redirect** | `http.Redirect(w, r, userURL)` | `http.SafeRedirect(w, r, url, "example.com")` |
| **Path Traversal** | `files.Read(userInput)` (in HTTP handler) | `sandbox.New("/var/data")` + `sandbox.Read(box, userInput)` |
| **Command Injection** | `shell.Run("git log {branch}")` | `shell.Output("git", "log", branch)` |

```kukicha
# --- XSS Prevention ---
import "stdlib/http" as httphelper

# UNSAFE — triggers compiler error for non-literal content
httphelper.HTML(w, userInput)  # XSS risk: http.HTML with non-literal content — use http.SafeHTML

# SAFE — HTML-escapes content before writing
httphelper.SafeHTML(w, userInput)

# --- SQL Injection Prevention ---
import "stdlib/pg"
# UNSAFE — string interpolation before parameterization
pg.Query(pool, "SELECT * FROM users WHERE name = '{name}'")  # compiler error

# SAFE — $N parameters
pg.Query(pool, "SELECT * FROM users WHERE name = $1", name)

# --- SSRF Prevention (inside HTTP handlers) ---
# UNSAFE — triggers compiler error inside any HTTP handler
fetch.Get(url)   # SSRF risk: fetch.Get inside an HTTP handler — use fetch.SafeGet

# SAFE — wraps netguard SSRF protection automatically
resp := fetch.SafeGet(url) onerr return

# --- Open Redirect Prevention ---
# UNSAFE — triggers compiler error for non-literal URL
httphelper.Redirect(w, r, userSuppliedURL)  # open redirect risk

# SAFE — validates host against explicit allowlist; relative URLs always pass
httphelper.SafeRedirect(w, r, returnURL, "example.com", "api.example.com") onerr return

# --- Path Traversal Prevention (inside HTTP handlers) ---
# UNSAFE — triggers compiler error inside any HTTP handler
files.Read(userInput)  # path traversal risk: files.Read inside an HTTP handler

# SAFE — use sandbox with a restricted root
import "stdlib/sandbox"
box := sandbox.New("/var/data") onerr return
content := sandbox.Read(box, userInput) onerr return

# --- Command Injection Prevention ---
# UNSAFE — triggers compiler error for non-literal argument
shell.Run("git log {branch}")  # command injection risk

# SAFE — pass arguments separately (no shell interpolation)
out := shell.Output("git", "log", branch) onerr return

# --- Response Body Size Limits ---
# Add a size cap to prevent OOM from oversized responses
resp := fetch.New(url) |> fetch.MaxBodySize(1 << 20) |> fetch.Do() onerr return
text := fetch.Text(resp) onerr return

# Cap request body when reading JSON (1 MB example)
httphelper.ReadJSONLimit(r, 1 << 20, reference of input) onerr return

# --- Security Headers ---
# Middleware: wraps an entire mux or handler
import "stdlib/http" as httphelper
http.ListenAndServe(":8080", httphelper.SecureHeaders(mux))

# Per-handler: set at the top of each handler
httphelper.SetSecureHeaders(w)

# --- HTML Templates (auto-escaping) ---
# UNSAFE — text/template performs NO HTML escaping
import "stdlib/template"
tmpl := template.New() |> template.WithContent(tmplStr)
template.Execute(tmpl) onerr return  # WARNING: plaintext only — no HTML escaping

# SAFE — html/template auto-escapes {{ }} values
result := template.HTMLRenderSimple(tmplStr, data) onerr return
```

## Module Structure

Every stdlib module is **pure Kukicha**: `<name>.kuki` source + `<name>.go` generated output. No `_helper.go` or `_tool.go` files.

All packages: `a2a`, `cast`, `cli`, `concurrent`, `container`, `crypto`, `ctx`, `datetime`, `encoding`, `env`, `errors`, `fetch`, `files`,
`git`, `http`, `input`, `iterator`, `json`, `kube`, `llm`, `maps`, `math`, `mcp`, `must`, `net`, `netguard`, `obs`, `parse`, `pg`,
`random`, `regex`, `retry`, `sandbox`, `semver`, `shell`, `slice`, `sort`, `string`, `table`, `template`, `test`, `validate`

## Import Aliases

When a package's last path segment collides with a local variable name, use `as`. Always use these standard aliases:

| Package | Standard alias | Reason |
|---------|----------------|--------|
| `stdlib/ctx` | `ctxpkg` | Clashes with local `ctx` variable |
| `stdlib/errors` | `errs` | Clashes with local `err` / `errors` |
| `stdlib/json` | `jsonpkg` | Clashes with `encoding/json` |
| `stdlib/string` | `strpkg` | Clashes with `string` type name |
| `stdlib/container` | `docker` | Clashes with local `container` variables |
| `stdlib/http` | `httphelper` | Clashes with `net/http` |
| `stdlib/net` | `netutil` | Clashes with `net` stdlib package |

```kukicha
import "stdlib/ctx" as ctxpkg          # avoids clash with local 'ctx' variables
import "stdlib/errors" as errs         # avoids clash with local 'err' / 'errors' variables
import "stdlib/json" as jsonpkg        # avoids clash with 'encoding/json'
import "github.com/jackc/pgx/v5" as pgx
```

## Vulnerability Auditing

The stdlib is extracted as a standalone Go module (`.kukicha/stdlib/`) with its own `go.mod` (embedded in `cmd/kukicha/stdlib.go`). When users run `kukicha audit` in their project, govulncheck follows the `replace` directive into the extracted stdlib and checks its dependencies transitively.

```bash
# In the kukicha repo: audit compiler + stdlib dependencies together
kukicha audit

# In a user project: audits project deps including stdlib transitively
kukicha audit

# During build
kukicha build --vulncheck app.kuki
```

When updating stdlib dependencies (the `stdlibGoMod` constant in `cmd/kukicha/stdlib.go`), always run `kukicha audit` afterward to verify no new vulnerabilities were introduced.

## Common Pitfalls

Patterns that look correct but introduce subtle bugs. Each has burned us before.

### WaitGroups in goroutines — use `defer wg.Done()`

Always make `defer wg.Done()` the **first statement** in a goroutine body. An explicit `wg.Done()` at the end is silently skipped if the task panics, hanging `wg.Wait()` forever.

```kukicha
# WRONG — hangs if t() panics
go func()
    t()
    wg.Done()
()

# CORRECT — defer fires even on panic
go func()
    defer wg.Done()
    t()
()
```

### Context/cancel lifetime — defer cancel in the function that owns the resource

Never call `defer ctxpkg.Cancel(h)` inside a helper that *returns* a resource built with that context. The cancel fires when the helper returns — before the resource is ever used — making the timeout dead on arrival. Defer cancel in the same function whose lifetime covers the resource's use.

```kukicha
# WRONG — cancel fires when buildCmd returns, context is already done
func buildCmd(cmd Command) reference exec.Cmd
    h := ctxpkg.WithTimeout(ctxpkg.Background(), cmd.timeout as int64)
    defer ctxpkg.Cancel(h)   # fires here, before exec.Run
    return exec.CommandContext(ctxpkg.Value(h), cmd.name, many cmd.args)

# CORRECT — cancel deferred in Execute, fires after execCmd.Run() completes
func Execute(cmd Command) Result
    execCmd := exec.Command(cmd.name, many cmd.args)
    if cmd.timeout > 0
        h := ctxpkg.WithTimeout(ctxpkg.Background(), cmd.timeout as int64)
        defer ctxpkg.Cancel(h)   # fires after Run()
        execCmd = exec.CommandContext(ctxpkg.Value(h), cmd.name, many cmd.args)
    ...
```

### io.ReadCloser wrapping — never use `io.NopCloser` on a live body

`io.NopCloser` replaces `Close()` with a no-op. Wrapping a response body with it means `Close()` never reaches the underlying connection, leaking it. When you need to cap reads with `io.LimitReader`, preserve the original closer with a wrapper type that delegates both `Read` and `Close`.

```kukicha
# WRONG — NopCloser silences Close(), TCP connection is never released
resp.Body = io.NopCloser(io.LimitReader(resp.Body, maxSize))

# CORRECT — limitReadCloser delegates Read to LimitReader, Close to original body
type limitReadCloser
    r io.Reader
    c io.Closer

func Read on b reference limitReadCloser (p list of byte) (int, error)
    return b.r.Read(p)

func Close on b reference limitReadCloser () error
    return b.c.Close()

resp.Body = reference of limitReadCloser{r: io.LimitReader(resp.Body, maxSize), c: resp.Body}
```

## Critical Rules

1. **Never edit generated `*.go` files in stdlib** — edit `.kuki` source, then `make generate`
2. **Never edit `internal/semantic/stdlib_registry_gen.go` or `go_stdlib_gen.go`** — both are auto-generated; `make generate` regenerates `stdlib_registry_gen.go` automatically, and `make gengostdlib` regenerates `go_stdlib_gen.go` from Go stdlib signatures via `go/importer`
3. **Types must be defined in `.kuki`** — so the Kukicha compiler knows about them
4. **After adding an exported function to a stdlib `.kuki` file**, run `make genstdlibregistry` (or just `make generate`) so `onerr` and pipe expressions work correctly with the new function
5. **To deprecate a stdlib function**, add `# kuki:deprecated "Use NewFunc instead"` above it in the `.kuki` source, then run `make genstdlibregistry` — callers will get a compile-time warning
6. **To mark a function as security-sensitive**, add `# kuki:security "category"` above it (categories: `sql`, `html`, `fetch`, `files`, `redirect`, `shell`), then run `make genstdlibregistry` — the compiler will enforce the corresponding security check
7. **Every stdlib package must have a `*_test.kuki` file** using the table-driven pattern (see "Testing Stdlib Packages" above)
8. **`stdlib/test` is test-only** — import it only in `*_test.kuki` files, never in library `.kuki` files
