# Kukicha Language Limitations

This document explains why some stdlib packages still contain hand-written Go helper files
(`*_helper.go`) or tool files (`*_tool.go`) rather than being expressed purely in Kukicha.

## Current Packages with Go Helpers

| Package | File | Reason |
|---------|------|--------|
| `stdlib/a2a` | `a2a_helper.go` | Range-over-func, type switch |
| `stdlib/container` | `container_helper.go` | Functional options, streaming I/O |
| `stdlib/mcp` | `mcp_tool.go` | Multi-statement SDK callback closure |

---

## ~~1. Function Types~~ (Resolved)

Kukicha now supports named function type aliases:

```kukicha
type TextHandler func(string)
type StatusHandler func(StatusUpdate)
type ToolHandler func(map of string to any) (any, error)
```

This unblocked moving most of `stdlib/a2a` and the `ToolHandler` type from `stdlib/mcp` to
pure Kukicha.

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

**Affects:** `container.Connect`, `container.ConnectRemote`, `container.Open`.

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

## 4. Multi-Statement SDK Callback Closures

Some Go SDK APIs require passing a multi-statement `func(ctx, req) (resp, error)` callback
at registration time. Kukicha has function literals and block lambdas, but the MCP SDK's
`server.AddTool` requires a closure that captures variables, unmarshals JSON, dispatches to
a handler, and wraps the response — a pattern that spans many statements with SDK-specific
types. The named function type (`ToolHandler`) is now expressible in Kukicha, but the
multi-statement closure body of the `Tool` function itself remains in Go.

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

- ~~**Function types**~~ — `type Handler func(string)` — **implemented**
- **Anonymous struct literals** — enables inline JSON decode targets
- **Range-over-func** — enables iterator-based streaming APIs
