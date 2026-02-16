# Kukicha Language Limitations

This document explains why some stdlib packages still contain hand-written Go helper files
(`*_helper.go`) or tool files (`*_tool.go`) rather than being expressed purely in Kukicha.

## Current Packages with Go Helpers

| Package | File | Reason |
|---------|------|--------|
| `stdlib/a2a` | `a2a_helper.go` | Function types, range-over-func |
| `stdlib/container` | `container_helper.go` | Functional options, streaming I/O |
| `stdlib/kube` | `kube_helper.go` | Complex API client construction |
| `stdlib/mcp` | `mcp_tool.go` | SDK callback closures |

---

## 1. Function Types

Kukicha has no syntax for declaring a named function type:

```go
// Go — not expressible in Kukicha
type TextHandler func(string)
type StatusHandler func(StatusUpdate)
```

This blocks moving `stdlib/a2a` to pure Kukicha, because `Request` holds `onText` and
`onStatus` fields of these callback types, and the streaming functions dispatch to them.

**Workaround until resolved:** Keep the type declarations and functions that use them in
`a2a_helper.go`.

---

## 2. Functional Options / Variadic `...T` of Interface Type

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

**Affects:** `container.Connect`, `container.ConnectRemote`, `container.Open`,
`kube.Connect`, `kube.Open`.

---

## 3. Range-over-Func (Go 1.23+)

Go 1.23 introduced iterators via `range` over function values. Several Go SDKs expose
streaming APIs this way:

```go
// Go — Kukicha for-in does not support this
for event, err := range client.SendStreamingMessage(ctx, params) {
    ...
}
```

**Affects:** `a2a.streamRequest` — the A2A streaming response is an iterator function.

---

## 4. SDK Callback Closures

Some Go SDK APIs require passing a `func(ctx context.Context, req *T) (*R, error)` callback
at registration time. Kukicha lambdas (`(x T) => expr`) work for simple expression lambdas
but multi-statement closures that capture external state and use SDK-specific types are
awkward to express without full closure syntax.

**Affects:** `mcp.Tool` — the MCP SDK's `server.AddTool` requires a context-aware callback.

---

## 5. Complex Streaming I/O

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

## Roadmap

These limitations are tracked in `docs/PLAN-Drop-Go-Helpers.md`. The planned language
additions that would eliminate most remaining helpers:

- **Function types** — `type Handler func(string)` — eliminates A2A and MCP helpers
- **Anonymous struct literals** — enables inline JSON decode targets
- **Range-over-func** — enables iterator-based streaming APIs
