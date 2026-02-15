# stdlib/AGENTS.md

Kukicha standard library reference. Each package lives in `stdlib/<name>/` with:
- `<name>.kuki` — Kukicha source (types, function signatures, inline implementations)
- `<name>_helper.go` — Go implementation (for packages that wrap Go libraries)
- `<name>.go` — **Generated** by `make generate` from the `.kuki` file. Never edit directly.

Import with: `import "stdlib/slice"`

## Packages

| Package | Purpose | Key Functions |
|---------|---------|---------------|
| `stdlib/a2a` | Agent-to-Agent protocol client | Discover, Ask, Send, Stream, New/Text/Context |
| `stdlib/cast` | Type casting utilities | ToString, ToInt, ToFloat, ToBool |
| `stdlib/cli` | CLI argument parsing | New, String, Int, Bool, Parse |
| `stdlib/concurrent` | Parallel execution | Parallel, ParallelWithLimit |
| `stdlib/datetime` | Named formats, duration helpers | Format, Seconds, Minutes, Hours |
| `stdlib/env` | Typed env vars with onerr | Get, GetInt, GetBool, GetFloat, GetOr, Set |
| `stdlib/fetch` | HTTP client (Builder, Auth, Sessions) | Get, Post, New/Header/Timeout/Do, BearerAuth, BasicAuth, FormData, NewSession |
| `stdlib/files` | File I/O operations | Read, Write, Append, Exists, Copy, Move, Delete, Watch |
| `stdlib/http` | HTTP response helpers | JSON, JSONError, JSONNotFound, ReadJSON |
| `stdlib/input` | User input utilities | Line, Confirm, Choose |
| `stdlib/iterator` | Functional iteration | Map, Filter, Reduce |
| `stdlib/json` | jsonv2 wrapper | Marshal, Unmarshal, UnmarshalRead, MarshalWrite |
| `stdlib/kube` | Kubernetes client via client-go | Connect, Namespace, ListPods, GetPod, ListDeployments, ScaleDeployment, PodLogs |
| `stdlib/llm` | Large language model client | Chat, Stream, System, User |
| `stdlib/maps` | Map utilities | Keys, Values, Has, Merge |
| `stdlib/mcp` | Model Context Protocol support | NewServer, Tool, Resource, Prompt |
| `stdlib/must` | Panic-on-error startup helpers | Env, EnvInt, EnvIntOr, Do, OkMsg |
| `stdlib/netguard` | Network restriction & SSRF protection | NewSSRFGuard, NewAllow, NewBlock, Check, HTTPTransport |
| `stdlib/parse` | CSV and YAML parsing | CSV, YAML |
| `stdlib/pg` | PostgreSQL client via pgx | Connect, Query, QueryRow, Exec, Begin, Commit, Rollback, ScanRow, CollectRows |
| `stdlib/random` | Random number generation | Int, IntRange, Float, String, Choice |
| `stdlib/result` | Result and Optional types | Ok, Err, Unwrap, UnwrapOr |
| `stdlib/retry` | Retry patterns | Do, WithBackoff |
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

# HTTP fetch with builder
import "stdlib/fetch"
resp := fetch.New(url) |> fetch.BearerAuth(token) |> fetch.Timeout(30000000000) |> fetch.Do() onerr panic "{error}"
text := fetch.Text(resp) onerr panic "{error}"

# Network-restricted fetch (SSRF protection)
import "stdlib/netguard"
guard := netguard.NewSSRFGuard()
resp := fetch.New(url) |> fetch.Transport(netguard.HTTPTransport(guard)) |> fetch.Do() onerr panic "{error}"
```

## Module Structure

Each stdlib module follows one of two patterns:

### Pure Kukicha (types + logic in .kuki)
Used when the implementation is straightforward Kukicha code.
Examples: `slice`, `string`, `validate`, `env`, `must`, `fetch`

### Kukicha types + Go helper (types in .kuki, implementation in _helper.go)
Used when wrapping complex Go libraries. The `.kuki` file defines types visible to Kukicha code, and the `_helper.go` provides the implementation in Go.
Examples: `a2a`, `netguard`, `sandbox`, `pg`, `kube`

## Critical Rules

1. **Never edit `*.go` files in stdlib** — edit `.kuki` source, then `make generate`
2. **Helper files are hand-written Go** — `*_helper.go` files are NOT generated
3. **Types must be defined in `.kuki`** — so the Kukicha compiler knows about them
4. **Functions in `_helper.go` must match exported signatures** — field names must match the `.kuki` struct definitions exactly (lowercase)
