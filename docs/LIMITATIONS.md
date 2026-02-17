# Kukicha Language Limitations

This document explains why some stdlib packages still contain hand-written Go helper files
(`*_helper.go`) or tool files (`*_tool.go`) rather than being expressed purely in Kukicha.

## Current Packages with Go Helpers

| Package | File | Reason |
|---------|------|--------|
| `stdlib/container` | `container_helper.go` | Functional options, streaming I/O |
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

## Roadmap

Additions that would eliminate most remaining helpers:

- **Anonymous struct literals** — enables inline JSON decode targets
