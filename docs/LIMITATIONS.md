# Kukicha Language Limitations

This document explains why some stdlib packages still contain hand-written Go helper files
(`*_helper.go`) or tool files (`*_tool.go`) rather than being expressed purely in Kukicha,
and notes areas where the stdlib overlaps with dedicated tooling.

## Current Packages with Go Helpers

| Package | File | Lines | Reason |
|---------|------|-------|--------|
| `stdlib/container` | `container_helper.go` | ~400 | Functional options, tar archive handling, `filepath.WalkDir` closures |

---

## 1. Functional Options / Variadic `...T` of Interface Type

Go libraries often accept a variadic slice of option functions:

```go
// Go — not expressible in Kukicha
opts := []client.Opt{
    client.FromEnv,
    client.WithAPIVersionNegotiation(),
    client.WithHost(host),
}
cli, err := client.NewClientWithOpts(opts...)
```

Kukicha's `many` keyword handles variadic arguments but cannot build a `[]client.Opt` slice
incrementally and pass it with spread syntax when the element type is a function type.

**Affects:** `container.Connect`, `container.ConnectRemote`, `container.Open`.

---

## 2. Multi-Statement SDK Callback Closures (fixed)

Kukicha now supports multi-statement function literals as arguments to function calls.
The `mcp.Tool` function — which passes a multi-statement `func(ctx, req) (resp, error)`
closure to `server.AddTool` — has been fully migrated from `mcp_tool.go` to `mcp.kuki`.
The `mcp_tool.go` file has been deleted.

Note: bare `error` as a return value in `onerr` handlers is still parsed as the
`error "message"` keyword, so conventional `if err != empty` is used instead of `onerr`
for `json.Unmarshal` in the migrated `Tool()` function.

---

## 3. Complex Streaming I/O (partially resolved)

Some helper functions mix `bufio.Scanner`, anonymous struct JSON targets, and multi-pass
stream processing that is idiomatic in Go but verbose in Kukicha:

```go
// Go — terse due to anonymous struct
var msg struct {
    Status string `json:"status"`
    Digest string `json:"id"`
}
json.Unmarshal(scanner.Bytes(), &msg)
```

Kukicha requires named types for struct literals and does not support anonymous struct
declarations inline in function bodies. However, the **workaround** is straightforward:
define named types at package level and use them in place of anonymous structs.

The following functions were migrated from `container_helper.go` to `container.kuki`
using named types (`pullStatusMsg`, `buildStreamMsg`, `dockerAuthEntry`, `dockerConfig`):

- `Pull` — image pull with digest extraction from JSON stream
- `PullAuth` — authenticated image pull
- `loadDockerAuth` — reads `~/.docker/config.json` credentials
- `LoginFromConfig` — creates `Auth` from Docker config file

Note: `container.containerLogs` was previously migrated to Kukicha — it uses
`stdcopy.StdCopy` for demuxing, not anonymous structs.

**Still in Go:** `container.buildImage` — uses anonymous structs *and* a
`filepath.WalkDir` closure, which requires the closure-as-argument pattern.


---

## 5. Kube Operations That Overlap with GitOps Controllers (ArgoCD, Flux)

Several `stdlib/kube` functions perform imperative mutations that, in production clusters
managed by ArgoCD or Flux, will be reverted or cause drift alerts. These functions are
useful for local dev clusters, CI test environments, and one-off scripts, but should
**not** be used against GitOps-managed namespaces.

| Function | What it does | ArgoCD conflict |
|----------|-------------|-----------------|
| `ScaleDeployment` | Imperatively sets replica count | ArgoCD will revert to the count in the Git manifest on the next sync. Use `argocd app set` or edit the manifest repo instead. |
| `DeleteDeployment` | Deletes a deployment by name | ArgoCD will recreate it on the next sync since the manifest still exists in Git. For actual removal, delete from the manifest repo. |
| `DeletePod` | Deletes a pod by name | Generally safe (the controller recreates it), but in ArgoCD "self-heal" mode this may trigger an unnecessary sync. |
| `RolloutRestart` | Patches the pod template annotation to trigger a rollout | ArgoCD will detect the annotation drift. In ArgoCD-managed apps, use `argocd app actions run <app> restart --kind Deployment` instead. |

### Functions that are safe alongside ArgoCD

These are **read-only** or **observe-only** and do not conflict with GitOps controllers:

- `Connect`, `Open`, `Namespace` — connection setup
- `ListPods`, `GetPod`, `ListDeployments`, `GetDeployment` — read-only queries
- `ListServices`, `GetService`, `ListNodes`, `GetNode`, `ListNamespaces` — read-only queries
- `PodLogs`, `PodLogsTail` — read-only log retrieval
- `WatchPods`, `WatchPodsCtx` — read-only event streaming
- `WaitDeploymentReady`, `WaitPodReady` (and Ctx variants) — read-only polling
- All accessor functions (`PodName`, `PodStatus`, `DeploymentImage`, etc.) — pure data extraction

### Recommendation

If your cluster uses ArgoCD or Flux, treat the mutating kube functions as **dev/test
only**. For production deployments, push changes to your Git manifest repo and let the
GitOps controller reconcile. The read-only and observability functions in `stdlib/kube`
remain valuable for dashboards, health checks, and CI verification scripts that run
*after* ArgoCD has synced.

---

## 6. Compiler Code-Generation Quirks

### `error "..."` interpolation auto-imports `fmt` (fixed)

The compiler generates `errors.New(fmt.Sprintf(...))` for interpolated `error ""` strings
and now **automatically imports both `"errors"` and `"fmt"`** when needed. No manual
`import "fmt"` is required.

```go
// Kukicha source
return 0, error "environment variable {key} not set"

// Generated Go (fmt and errors auto-imported)
return 0, errors.New(fmt.Sprintf("environment variable %v not set", key))
```

### `stdlib/string` does not wrap all of Go's `strings` package

A few `strings` functions have no equivalent in `stdlib/string`:

| Go `strings` function | Replacement |
|---|---|
| `strings.CutPrefix(s, prefix)` → `(after, found)` | `string.HasPrefix(s, prefix)` + `string.TrimPrefix(s, prefix)` |
| `strings.NewReader(s)` → `io.Reader` | `bytes.NewBufferString(s)` (import `"bytes"`) |

When migrating `.kuki` files from raw `import "strings"` to `import "stdlib/string"`, check for
these two functions. Both replacements are slightly more verbose but fully equivalent.

---

## What It Would Actually Take to Eliminate the Go Helpers

Many kube helper functions already use only patterns that Kukicha supports (type
assertions, pointer dereferencing, struct literals, defer, loops). The actual blockers
are more specific than "anonymous struct literals."

### ~~Blocker 1: `select` statement support~~ (resolved)

Kukicha now supports the `select` keyword. The following functions were migrated from Go
helpers to Kukicha using `select`:

- `kube.WatchPods`, `kube.WatchPodsCtx`, `kube.watchPodsWithContext` — migrated to `kube.kuki`; `kube_helper.go` deleted
- `container.Wait`, `container.WaitCtx` — migrated to `container.kuki`
- `container.Events`, `container.EventsCtx`, `container.eventsWithContext`, `container.convertEvent` — migrated to `container.kuki`

### ~~Blocker 2: Anonymous struct literals~~ (resolved via named types)

Go uses anonymous structs for inline JSON decode targets:

```go
var msg struct {
    Status string `json:"status"`
    ID     string `json:"id"`
}
json.Unmarshal(scanner.Bytes(), &msg)
```

Kukicha doesn't support anonymous struct declarations, but the workaround is to define
named types at package level with `json:"tag"` struct tags. This approach was used to
migrate `Pull`, `PullAuth`, `loadDockerAuth`, and `LoginFromConfig` from
`container_helper.go` to `container.kuki`.

Note: `buildImage` still remains in Go because it *also* depends on a
`filepath.WalkDir` closure (Blocker 3-adjacent).

### Blocker 3: Variadic interface arg spreading

The Docker SDK uses `client.NewClientWithOpts(opts...)` where `opts` is a
`[]client.Opt` built incrementally. Kukicha's `many` keyword can't build and spread
a slice of function-typed values.

This blocks 3 functions: `newClient`, `Connect`/`ConnectRemote`, `Open`.

### ~~Blocker 4: Closure-as-struct-field~~ (resolved)

Kukicha's brace-delimited struct literals suppress newlines (preventing multi-line
function literal bodies inside `{...}`), and the indented struct literal form only
works for unqualified types. However, this was resolved by using **separate field
assignment**: create the struct, then assign the closure field separately.

The following functions were migrated from `netguard_helper.go` to `netguard.kuki`:

- `DialContext` — uses `dialer.Control = func(...)` instead of embedding in struct literal
- `HTTPTransport` — uses `t.DialContext = func(...)` instead of embedding in struct literal
- `HTTPClient` — simple struct literal with no closure (no workaround needed)

The `netguard_helper.go` file has been deleted.

### Still in Go — remaining blockers

| Function | Blocker | File |
|----------|---------|------|
| `newClient` | Variadic interface spreading (`opts...`) | `container_helper.go` |
| `Connect`, `ConnectRemote` | Call `newClient` | `container_helper.go` |
| `Open` | Variadic interface spreading (`opts...`) | `container_helper.go` |
| `buildImage` | `filepath.WalkDir` closure + anonymous structs | `container_helper.go` |
| `Build` | Calls `buildImage` | `container_helper.go` |
| `CopyFrom`, `CopyTo` + helpers | `filepath.WalkDir` closure, tar archive handling | `container_helper.go` |
### Impact summary

| Blocker | Helpers it would unlock | Effort |
|---------|------------------------|--------|
| ~~`select` statement~~ | ~~7 functions (kube watch + container wait/events)~~ | **Resolved** — migrated to .kuki |
| ~~Anonymous struct literals~~ | ~~4 functions (streaming JSON decode + auth in container)~~ | **Resolved** — named types workaround |
| Variadic interface spreading | 4 functions (Docker client init) | Low-medium — extend `many` |
| ~~Closure-as-struct-field~~ | ~~3 functions (netguard DialContext, HTTPTransport, HTTPClient)~~ | **Resolved** — migrated to .kuki using separate field assignment |
| ~~Multi-statement closure callbacks~~ | ~~1 function (MCP tool registration)~~ | **Resolved** — migrated to .kuki |
