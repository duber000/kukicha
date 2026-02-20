# Kukicha Language Limitations

This document explains why some stdlib packages still contain hand-written Go helper files
(`*_helper.go`) or tool files (`*_tool.go`) rather than being expressed purely in Kukicha,
and notes areas where the stdlib overlaps with dedicated tooling.

## Current Packages with Go Helpers

| Package | File | Lines | Reason |
|---------|------|-------|--------|
| `stdlib/container` | `container_helper.go` | ~643 | Functional options, streaming I/O, tar archive handling, `select` |
| `stdlib/kube` | `kube_helper.go` | 69 | `select` statement in `watchPodsWithContext` (+ 2 callers) |
| `stdlib/netguard` | `netguard_helper.go` | 73 | Closure-as-struct-field pattern in `DialContext` |

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

## 3. Complex Streaming I/O

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
declarations inline in function bodies.

**Affects:** `container.Pull`, `container.PullAuth`, `container.buildImage`,
`container.loadDockerAuth`.

Note: `container.containerLogs` was migrated to Kukicha — it uses `stdcopy.StdCopy`
for demuxing, not anonymous structs.

---

## 4. `onerr` + Multi-Value `return` Inside Inline Callback Bodies (fixed)

The codegen now correctly handles `error "{error}"` inside multi-value `return` expressions
in `onerr` handlers, including inside inline callback/lambda bodies. The `{error}` placeholder
is properly substituted with the caught error variable (e.g., `err_1`).

```kukicha
func(args map of string to any) (any, error)
    data := sandbox.ReadString(box, path) onerr return mcp.ErrorResult(error.Error()), empty
    return data as any, empty
```

This pattern now works correctly without needing a named helper function.

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

### Blocker 1: `select` statement support (critical)

Kukicha has no `select` keyword. This blocks 4 functions that wait on channels:

- `container.Wait` / `container.WaitCtx` — select on wait channel vs error channel
- `container.eventsWithContext` — select on ctx.Done vs error vs message channels
- `kube.watchPodsWithContext` — select on ctx.Done vs watcher result channel (69 lines remaining in `kube_helper.go`)

Without `select`, any function that multiplexes channels must stay in Go.
Note: the kube `WatchPods`/`WatchPodsCtx` wrappers also remain in Go because the
Kukicha semantic checker only resolves identifiers within `.kuki` files, so they
cannot call the Go-only `watchPodsWithContext`.

### Blocker 2: Anonymous struct literals

Used for inline JSON decode targets in streaming I/O:

```go
var msg struct {
    Status string `json:"status"`
    ID     string `json:"id"`
}
json.Unmarshal(scanner.Bytes(), &msg)
```

This blocks 4 functions in `container_helper.go`: `buildImage`, `loadDockerAuth`,
`Pull`, `PullAuth`.

### Blocker 3: Variadic interface arg spreading

The Docker SDK uses `client.NewClientWithOpts(opts...)` where `opts` is a
`[]client.Opt` built incrementally. Kukicha's `many` keyword can't build and spread
a slice of function-typed values.

This blocks 3 functions: `newClient`, `Connect`/`ConnectRemote`, `Open`.

### Blocker 4: Multi-statement closure as callback argument (resolved)

The MCP SDK's `server.AddTool` requires a `func(ctx, req) (resp, error)` closure
that captures variables and contains branching logic. This was migrated to Kukicha
using multi-statement function literals as arguments. `mcp_tool.go` has been deleted.

### Migrated to .kuki (completed)

#### Kube (41 functions)

41 kube helper functions were migrated from `kube_helper.go` to `kube.kuki`,
reducing the Go helper from 621 lines to 69 lines. Only `watchPodsWithContext`
(which uses `select`) and its two callers remain in Go.

| Category | Functions | Count |
|----------|-----------|-------|
| Connection | Connect, Open, clientset | 3 |
| Pod CRUD | ListPods, ListPodsLabeled, GetPod, DeletePod | 4 |
| Deployment CRUD | ListDeployments, GetDeployment, ScaleDeployment, DeleteDeployment | 4 |
| Mutation | RolloutRestart | 1 |
| Wait/poll | WaitDeploymentReady, WaitPodReady, WaitDeploymentReadyCtx, WaitPodReadyCtx | 4 |
| Services | ListServices, GetService | 2 |
| Nodes | ListNodes, GetNode | 2 |
| Namespaces | ListNamespaces | 1 |
| List accessors | Pods, Deployments, Services, Nodes, Namespaces | 5 |
| Pod accessors | pod, PodName, PodStatus, PodIP, PodNode, PodAge, PodReady, PodRestarts, PodLabels | 9 |
| Deployment accessors | deployment, DeploymentName, DeploymentReplicas, DeploymentReady, DeploymentImage | 5 |
| Service accessors | service, ServiceName, ServiceType, ServiceClusterIP, ServicePorts | 5 |
| Node accessors | node, NodeName, NodeReady, NodeRoles, NodeVersion | 5 |
| Namespace accessors | nsItem, NamespaceName | 2 |
| Logs | PodLogs, PodLogsTail | 2 |

#### MCP (1 function — `mcp_tool.go` deleted)

`mcp.Tool()` was migrated from `mcp_tool.go` to `mcp.kuki`. The Go helper file was
deleted entirely. This demonstrated that multi-statement function literals as arguments
work in Kukicha — the closure passed to `server.AddTool` contains JSON unmarshalling,
error handling, type switching, and JSON marshalling.

#### Container (6 functions)

6 container functions were migrated from `container_helper.go` to `container.kuki`,
reducing the Go helper from ~782 lines to ~643 lines.

| Function | Pattern |
|----------|---------|
| `containerLogs` | `stdcopy.StdCopy` for demuxing, `io.ReadAll` fallback |
| `Logs` | Thin wrapper calling `containerLogs` |
| `LogsTail` | Wrapper with `fmt.Sprintf` for tail param |
| `Run` | `ContainerCreate` + `ContainerStart` with SDK struct literals |
| `Inspect` | `ContainerInspect` → `ContainerInfo` mapping |
| `Exec` | `ContainerExecCreate/Attach/Inspect` with `stdcopy.StdCopy` |

Patterns used across all migrations: `reference of` (address-of), `dereference` (pointer deref),
`as` (type assertions and conversions), struct literals with external SDK types, bare `for` loops,
`many` (variadic params), `k8s.io/...` imports with `as` aliases, and multi-statement
function literals as arguments.

### Still in Go — remaining blockers

| Function | Blocker | File |
|----------|---------|------|
| `watchPodsWithContext` | `select` statement | `kube_helper.go` |
| `WatchPods`, `WatchPodsCtx` | Call `watchPodsWithContext` (semantic checker only sees .kuki) | `kube_helper.go` |
| `Wait`, `WaitCtx` | `select` statement | `container_helper.go` |
| `eventsWithContext` | `select` statement | `container_helper.go` |
| `Events`, `EventsCtx` | Call `eventsWithContext` | `container_helper.go` |
| `convertEvent` | Called by `eventsWithContext` | `container_helper.go` |
| `newClient` | Variadic interface spreading (`opts...`) | `container_helper.go` |
| `Connect`, `ConnectRemote` | Call `newClient` | `container_helper.go` |
| `Open` | Variadic interface spreading (`opts...`) | `container_helper.go` |
| `buildImage` | Anonymous structs + `filepath.WalkDir` closure | `container_helper.go` |
| `Build` | Calls `buildImage` | `container_helper.go` |
| `loadDockerAuth` | Anonymous structs (nested) | `container_helper.go` |
| `LoginFromConfig` | Calls `loadDockerAuth` | `container_helper.go` |
| `Pull`, `PullAuth` | Anonymous structs for JSON decode | `container_helper.go` |
| `CopyFrom`, `CopyTo` + helpers | `filepath.WalkDir` closure, tar archive handling | `container_helper.go` |
| `DialContext` | Closure-as-struct-field in `net.Dialer` | `netguard_helper.go` |

### Impact summary

| Blocker | Helpers it would unlock | Effort |
|---------|------------------------|--------|
| `select` statement | 7 functions (kube watch + container wait/events) | Medium — new keyword, parser, codegen |
| Anonymous struct literals | 4 functions (streaming JSON decode + auth in container) | Medium — parser + codegen |
| Variadic interface spreading | 4 functions (Docker client init) | Low-medium — extend `many` |
| Closure-as-struct-field | 1 function (netguard `DialContext`) | Low — codegen for struct field func literals |
| ~~Multi-statement closure callbacks~~ | ~~1 function (MCP tool registration)~~ | **Resolved** — migrated to .kuki |
