# Kukicha Language Limitations

This document explains why some stdlib packages still contain hand-written Go helper files
(`*_helper.go`) or tool files (`*_tool.go`) rather than being expressed purely in Kukicha,
and notes areas where the stdlib overlaps with dedicated tooling.

## Current Packages with Go Helpers

| Package | File | Reason |
|---------|------|--------|
| `stdlib/container` | `container_helper.go` | Functional options, streaming I/O, tar archive handling |
| `stdlib/kube` | `kube_helper.go` | client-go SDK types, watch/wait polling loops |
| `stdlib/mcp` | `mcp_tool.go` | Multi-statement SDK callback closure |

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

## 2. Multi-Statement SDK Callback Closures

Some Go SDK APIs require passing a multi-statement `func(ctx, req) (resp, error)` callback
at registration time. Kukicha has function literals and block lambdas, but the MCP SDK's
`server.AddTool` requires a closure that captures variables, unmarshals JSON, dispatches to
a handler, and wraps the response — a pattern that spans many statements with SDK-specific
types. The named function type (`ToolHandler`) is now expressible in Kukicha, but the
multi-statement closure body of the `Tool` function itself remains in Go.

**Affects:** `mcp.Tool` — the MCP SDK's `server.AddTool` requires a context-aware callback.

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
`container.containerLogs`.

---

## 4. `onerr` + Multi-Value `return` Inside Inline Callback Bodies

In some inline callback/lambda contexts (for example, handler functions passed directly to
APIs like `mcp.Tool`), parser/codegen handling is still limited when `onerr` is followed by
a multi-value `return` in the same inline function body.

Example pattern that may fail in inline callback bodies:

```kukicha
func(args map of string to any) (any, error)
    data := sandbox.ReadString(box, path) onerr return mcp.ErrorResult(error.Error()), empty
    return data as any, empty
```

Current workaround: use explicit error variables in the inline callback body, or move logic
to a named helper function and return from there.

**Affects:** Inline tool/handler callbacks that need `(value, error)` fallback returns.

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

## What It Would Actually Take to Eliminate the Go Helpers

Many kube helper functions already use only patterns that Kukicha supports (type
assertions, pointer dereferencing, struct literals, defer, loops). The actual blockers
are more specific than "anonymous struct literals."

### Blocker 1: `select` statement support (critical)

Kukicha has no `select` keyword. This blocks 4 functions that wait on channels:

- `container.Wait` / `container.WaitCtx` — select on wait channel vs error channel
- `container.eventsWithContext` — select on ctx.Done vs error vs message channels
- `kube.watchPodsWithContext` — select on ctx.Done vs watcher result channel

Without `select`, any function that multiplexes channels must stay in Go.

### Blocker 2: Anonymous struct literals

Used for inline JSON decode targets in streaming I/O:

```go
var msg struct {
    Status string `json:"status"`
    ID     string `json:"id"`
}
json.Unmarshal(scanner.Bytes(), &msg)
```

This blocks 4 functions in `container_helper.go`: `buildImage`, `containerLogs`,
`Pull`, `PullAuth`. Also blocks `loadDockerAuth`.

### Blocker 3: Variadic interface arg spreading

The Docker SDK uses `client.NewClientWithOpts(opts...)` where `opts` is a
`[]client.Opt` built incrementally. Kukicha's `many` keyword can't build and spread
a slice of function-typed values.

This blocks 3 functions: `newClient`, `Connect`/`ConnectRemote`, `Open`.

### Blocker 4: Multi-statement closure as callback argument

The MCP SDK's `server.AddTool` requires a `func(ctx, req) (resp, error)` closure
that captures variables and contains branching logic. Kukicha's block lambdas can
express simple closures but not the full dispatch pattern needed here.

This blocks 1 function: `mcp.Tool`.

### Already expressible — could migrate to .kuki today

The **majority of kube helper functions** use only patterns Kukicha already supports.
These could be moved to `kube.kuki` with appropriate imports:

| Category | Functions | Count |
|----------|-----------|-------|
| CRUD operations | ListPods, GetPod, DeletePod, ListDeployments, GetDeployment, DeleteDeployment, ScaleDeployment, ListServices, GetService, ListNodes, GetNode, ListNamespaces | 12 |
| Mutation ops | RolloutRestart | 1 |
| Wait/poll loops | WaitDeploymentReady, WaitPodReady (and Ctx variants) | 4 |
| Log retrieval | PodLogs, PodLogsTail | 2 |
| Type accessors | Pods, Deployments, Services, Nodes, Namespaces, pod, deployment, service, node, nsItem, and all Pod*/Deployment*/Service*/Node*/Namespace* | ~25 |
| Connection | Connect, Open | 2 |

These only need: struct literals with named SDK types, type assertions (`value.(Type)`),
pointer dereferencing (`dereference`), `reference of`, loops, and conditionals — all
of which Kukicha supports. The main requirement is adding `k8s.io/...` imports to
`kube.kuki`.

### Impact summary

| Blocker | Helpers it would unlock | Effort |
|---------|------------------------|--------|
| Migrate already-expressible kube funcs | ~44 functions (most of kube_helper.go) | Low — just rewrite in .kuki syntax |
| `select` statement | 4 functions (channel multiplexing) | Medium — new keyword, parser, codegen |
| Anonymous struct literals | 5 functions (streaming JSON decode) | Medium — parser + codegen |
| Variadic interface spreading | 3 functions (Docker client init) | Low-medium — extend `many` |
| Multi-statement closure callbacks | 1 function (MCP tool registration) | Medium — extend block lambdas |
