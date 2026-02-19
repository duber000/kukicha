# stdlib/AGENTS.md

Kukicha standard library reference. Each package lives in `stdlib/<name>/` with:
- `<name>.kuki` — Kukicha source (types, function signatures, inline implementations)
- optional `<name>_helper.go` / `<name>_tool.go` — hand-written Go for cases not yet expressible in Kukicha
- `<name>.go` — **Generated** by `make generate` from the `.kuki` file. Never edit directly.

Import with: `import "stdlib/slice"`

## Packages

| Package | Purpose | Key Functions |
|---------|---------|---------------|
| `stdlib/a2a` | Agent-to-Agent protocol client | Discover, Ask, Send, Stream, New/Text/Context |
| `stdlib/cast` | Type casting utilities | ToString, ToInt, ToFloat, ToBool |
| `stdlib/cli` | CLI argument parsing | New, String, Int, Bool, Parse |
| `stdlib/concurrent` | Parallel execution | Parallel, ParallelWithLimit |
| `stdlib/container` | Docker/Podman client via Docker SDK | Connect, ListContainers, ListImages, Pull, Run, Stop, Remove, Build, Logs, Inspect, Wait/WaitCtx, Exec, Events/EventsCtx, CopyFrom, CopyTo |
| `stdlib/ctx` | Context timeout/cancellation helpers | Background, WithTimeoutMs, WithDeadlineUnix, Cancel, Done, Err |
| `stdlib/datetime` | Named formats, duration helpers | Format, Seconds, Minutes, Hours |
| `stdlib/encoding` | Base64 and hex encoding/decoding | Base64Encode, Base64Decode, Base64URLEncode, HexEncode, HexDecode |
| `stdlib/env` | Typed env vars with onerr | Get, GetInt, GetBool, GetFloat, GetOr, Set |
| `stdlib/errors` | Error wrapping and inspection | Wrap, Is, Unwrap, New, Join |
| `stdlib/fetch` | HTTP client (Builder, Auth, Sessions, Safe URL helpers, Retry) | Get, Post, Json, Decode, URLTemplate, URLWithQuery, PathEscape, QueryEscape, New/Header/Timeout/Retry/Do, BearerAuth, BasicAuth, FormData, NewSession |
| `stdlib/files` | File I/O operations | Read, Write, Append, Exists, Copy, Move, Delete, Watch |
| `stdlib/http` | HTTP response helpers | JSON, JSONError, JSONNotFound, ReadJSON, SafeURL |
| `stdlib/input` | User input utilities | Line, Confirm, Choose |
| `stdlib/iterator` | Functional iteration | Map, Filter, Reduce |
| `stdlib/json` | jsonv2 wrapper | Marshal, Unmarshal, UnmarshalRead, MarshalWrite, DecodeRead |
| `stdlib/kube` | Kubernetes client via client-go | Connect, New/Kubeconfig/Context/InCluster/Retry/Open, Namespace, ListPods, GetPod, ListDeployments, ScaleDeployment, RolloutRestart, WaitDeploymentReady/WaitDeploymentReadyCtx, WaitPodReady/WaitPodReadyCtx, WatchPods/WatchPodsCtx, PodLogs |
| `stdlib/llm` | Large language model client (Chat Completions, OpenResponses, Anthropic; Retry) | Ask/Send/Complete, RAsk/RSend/Respond, MAsk/MSend/AnthropicComplete, Retry/RRetry/MRetry |
| `stdlib/maps` | Map utilities | Keys, Values, Has, Merge |
| `stdlib/mcp` | Model Context Protocol support | NewServer, Tool, Resource, Prompt |
| `stdlib/must` | Panic-on-error startup helpers | Env, EnvInt, EnvIntOr, Do, OkMsg |
| `stdlib/net` | IP address and CIDR utilities | ParseIP, ParseCIDR, Contains, SplitHostPort, LookupHost, IsLoopback, IsPrivate |
| `stdlib/netguard` | Network restriction & SSRF protection | NewSSRFGuard, NewAllow, NewBlock, Check, HTTPTransport |
| `stdlib/obs` | Structured observability helpers | New, Component, WithCorrelation, NewCorrelationID, Info, Warn, Error, Start, Stop, Fail |
| `stdlib/parse` | CSV and YAML parsing | CSV, YAML |
| `stdlib/pg` | PostgreSQL client via pgx | Connect, New/MaxConns/MinConns/Retry/Open, Query, QueryRow, Exec, Begin, Commit, Rollback, ScanRow, CollectRows |
| `stdlib/random` | Random number generation | Int, IntRange, Float, String, Choice |
| `stdlib/retry` | Retry with backoff | New, Attempts, Delay, Sleep |
| `stdlib/sandbox` | os.Root filesystem sandboxing | New, Read, Write, List, Exists, Delete |
| `stdlib/shell` | Safe command execution | New/Dir/Env/Execute, Output, Which, Getenv |
| `stdlib/slice` | Slice operations | Filter, Map, GroupBy, GetOr, FirstOr, Find, Pop |
| `stdlib/string` | String utilities | Split, Join, Trim, Contains, Replace, ToUpper, ToLower |
| `stdlib/template` | Text templating | Execute, New |
| `stdlib/validate` | Input validation | Email, URL, InRange, NotEmpty, MinLen, MaxLen |

## Common Patterns

```kukicha
# Validation (returns error for onerr)
import "stdlib/validate"
email |> validate.Email() onerr return error "{error}"
age |> validate.InRange(18, 120) onerr return error "{error}"

# Startup config (panics if missing/invalid)
import "stdlib/must"
apiKey := must.Env("API_KEY")
port := must.EnvIntOr("PORT", 8080)

# Runtime config (returns error for onerr)
import "stdlib/env"
debug := env.GetBoolOrDefault("DEBUG", false)

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
datetime.Format(t, "iso8601")  # Not "2006-01-02T15:04:05Z07:00"!
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

# Manual retry loop (for custom retry conditions)
import "stdlib/retry"
cfg := retry.New() |> retry.Attempts(5) |> retry.Delay(200)
attempt := 0
for attempt < cfg.MaxAttempts
    result, err := doSomething()
    if err == empty
        break
    retry.Sleep(cfg, attempt)
    attempt = attempt + 1

# HTTP fetch with builder
resp := fetch.New(url) |> fetch.BearerAuth(token) |> fetch.Timeout(30000000000) |> fetch.Do() onerr panic "{error}"
text := fetch.Text(resp) onerr panic "{error}"

# Typed JSON decode (readable API flow)
repos := fetch.Get(url) |> fetch.CheckStatus() |> fetch.Json(list of Repo) onerr panic "{error}"

# Safe URL construction (path + query encoding)
base := fetch.URLTemplate("https://api.github.com/users/{username}/repos", map of string to string{"username": username}) onerr panic "{error}"
safeURL := fetch.URLWithQuery(base, map of string to string{"per_page": "30", "sort": "stars"}) onerr panic "{error}"

# Network-restricted fetch (SSRF protection)
import "stdlib/netguard"
guard := netguard.NewSSRFGuard()
resp := fetch.New(url) |> fetch.Transport(netguard.HTTPTransport(guard)) |> fetch.Do() onerr panic "{error}"

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

# Error wrapping and inspection
import "stdlib/errors"
err := errors.Wrap(originalErr, "loading config")
# err.Error() == "loading config: <original message>"
if errors.Is(err, io.EOF)
    print("end of file")

# Base64 and hex encoding
import "stdlib/encoding"
encoded := encoding.Base64Encode("hello" as list of byte)
decoded := encoding.Base64Decode(encoded) onerr panic "invalid base64: {error}"
hexStr := encoding.HexEncode(hashBytes)
```

## Module Structure

Each stdlib module follows one of two patterns:

### Pure Kukicha (types + logic in .kuki)
Used when the implementation is straightforward Kukicha code.
Examples: `a2a`, `slice`, `string`, `validate`, `env`, `must`, `fetch`, `net`, `errors`, `encoding`

### Kukicha types + Go helper (types in .kuki, implementation in _helper.go)
Used when wrapping complex Go libraries. The `.kuki` file defines types visible to Kukicha code, and the `_helper.go` provides the implementation in Go.
Function type aliases (`type Handler func(string)`) are supported in `.kuki` files, enabling callback types for packages like `mcp`.
Examples: `container`, `kube`

### Mixed (most logic in .kuki, thin Go helper for syscall-level ops)
Used when most logic can be pure Kukicha but some low-level Go operations are needed.
Examples: `netguard` (IP/CIDR logic in .kuki, DNS+dialer in `_helper.go`), `mcp` (core in `.kuki`, callback bridge in `_tool.go`)

## Critical Rules

1. **Never edit generated `*.go` files in stdlib** — edit `.kuki` source, then `make generate`
2. **Helper/tool files are hand-written Go** — `*_helper.go` and `*_tool.go` are NOT generated
3. **Types must be defined in `.kuki`** — so the Kukicha compiler knows about them
4. **Functions in helper/tool files must match exported signatures** — field names must match the `.kuki` struct definitions exactly (lowercase)
