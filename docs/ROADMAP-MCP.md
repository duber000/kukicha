# MCP Stdlib & Agent Support Plan

> **Status**: ✅ Fully Implemented

This roadmap has been completed. See [Verification](#verification) for testing instructions.

## Context

Kukicha needs first-class MCP (Model Context Protocol) support to enable writing MCP servers in beginner-friendly syntax. Three problems to solve:

1. **LLMs send numbers as strings** — `args["port"] as int` fails on `"8080"` because `as` does a Go type assertion, not smart conversion
2. **`print()` corrupts MCP transport** — MCP uses stdout for JSON-RPC 2.0; `print()` compiles to `fmt.Println()` which writes to stdout
3. **No MCP server library** — users must hand-write JSON-RPC plumbing

Function metadata/schema from comments is deferred — the official Go SDK's generic `mcp.AddTool` auto-infers JSON Schema from Go struct tags, which is a better approach.

---

## Phase 1: `stdlib/cast/` — Smart Type Conversion

**No compiler changes.** Pure Go stdlib package (Kukicha doesn't support type switches, so this is hand-written `.go`).

### Create: `stdlib/cast/cast.go`

Functions (all take `any`, return `(T, error)`, pipe-friendly):

| Function | Handles |
|----------|---------|
| `SmartInt(value any) (int, error)` | int, int64, float64, string, json.Number, bool |
| `SmartFloat64(value any) (float64, error)` | float64, float32, int, int64, string, json.Number, bool |
| `SmartBool(value any) (bool, error)` | bool, int, float64, string |
| `SmartString(value any) (string, error)` | string, int, int64, float64, bool, []byte, fmt.Stringer |

Implementation: Go type switch + `strconv` conversions. Includes `json.Number` handling since JSON unmarshaling produces these types.

**Usage in Kukicha:**
```kukicha
port := cast.SmartInt(args["port"]) onerr return error
# or with pipe:
port := args["port"] |> cast.SmartInt() onerr return error
```

### Create: `stdlib/cast/cast_test.go`

Test each SmartX function with native types, string representations, json.Number, edge cases.

---

## Phase 2: `--target mcp` + File Directive

Support both CLI flag and file-level directive:
- CLI: `kukicha build --target mcp server.kuki`
- Directive: `# target: mcp` at top of `.kuki` file (before `petiole`)

### Modify: `internal/codegen/codegen.go`

1. Add field to Generator struct (~line 42):
   ```go
   mcpTarget bool
   ```

2. Add setter: `func (g *Generator) SetMCPTarget(v bool)`

3. Modify print transpilation in `generateCallExpr` (~line 2231) and pipe print (~line 2124):
   - Normal: `print(...)` → `fmt.Println(...)`
   - MCP: `print(...)` → `fmt.Fprintln(os.Stderr, ...)`

4. Modify `generateImports` (~line 148): when `mcpTarget && needsPrintBuiltin()`, add `"os"` to auto-imports

### Modify: `internal/lexer/lexer.go` or `internal/parser/parser.go`

Detect `# target: mcp` directive in the first few lines (before `petiole`). This is a comment-based directive, so it needs to be extracted during lexing or early parsing. Simplest approach: scan raw source for `# target: mcp` before tokenizing, similar to how Go handles `//go:build` directives.

### Modify: `cmd/kukicha/main.go`

- Parse `--target mcp` flag in `build` and `run` commands
- Also check for directive in source file before codegen
- Pass to generator via `gen.SetMCPTarget(true)`
- Update usage text

### Modify: `internal/ast/ast.go`

Add `Target string` field to `Program` struct to carry directive info from parser to codegen.

---

## Phase 3: `stdlib/mcp/` — MCP Server Wrapper

Uses the official **`github.com/modelcontextprotocol/go-sdk/mcp`** package. The Kukicha `stdlib/mcp` is a thin, pipe-friendly wrapper that makes the SDK beginner-accessible.

### Add dependency: `go.mod`

```
require github.com/modelcontextprotocol/go-sdk v1.x.x
```

### Create: `stdlib/mcp/mcp.kuki` + `stdlib/mcp/mcp.go`

The wrapper provides a builder-pattern API that hides Go generics, context.Context, and callback signatures behind Kukicha-friendly functions.

**Server creation & builder functions:**
```kukicha
# New creates an MCP server
func New(name string, version string) reference mcp.Server

# Serve runs the server on stdio transport (blocking)
func Serve(server reference mcp.Server) error
```

**Tool registration** (wraps the raw `server.AddTool` since Kukicha can't express Go generics):
```kukicha
# Tool registers a tool with manual schema
func Tool(server reference mcp.Server, name string, description string,
          schema reference jsonschema.Schema, handler ToolHandler)

# ToolHandler is: func(args map of string to any) (string, error)
# Simplified from the SDK's generic handler — the wrapper handles
# CallToolRequest/CallToolResult conversion
```

**Schema helpers** (builds `jsonschema.Schema` objects):
```kukicha
func Schema(properties ...SchemaProperty) reference jsonschema.Schema
func Prop(name string, typ string, description string) SchemaProperty
func Required(schema reference jsonschema.Schema, names ...string) reference jsonschema.Schema
```

**Result helpers:**
```kukicha
func TextResult(text string) reference mcp.CallToolResult
func ErrorResult(msg string) reference mcp.CallToolResult
```

**Resource & Prompt registration** (similar thin wrappers):
```kukicha
func Resource(server reference mcp.Server, uri string, name string, description string, handler ResourceHandler)
func Prompt(server reference mcp.Server, name string, description string, args list of mcp.PromptArgument, handler PromptHandler)
```

**Key design decisions:**
- The wrapper's `ToolHandler` type is `func(args map[string]any) (string, error)` — simpler than the SDK's generic `ToolHandlerFor[In, Out]`. The wrapper internally creates a `CallToolResult` with `TextContent`.
- For advanced use, users can import `github.com/modelcontextprotocol/go-sdk/mcp` directly and use the full SDK API.
- `Serve()` wraps `server.Run(context.Background(), &mcp.StdioTransport{})` — one-liner for beginners.
- The wrapper uses `server.AddTool` (the non-generic version that takes `ToolHandler`) internally, not the generic `mcp.AddTool`.

---

## Example: Complete MCP Server in Kukicha

```kukicha
# target: mcp

import "stdlib/mcp"
import "stdlib/cast"

func main()
    server := mcp.New("greeter", "1.0.0")

    mcp.Tool(server, "greet", "Greet a person",
        mcp.Schema(
            mcp.Prop("name", "string", "Name to greet"),
            mcp.Prop("count", "integer", "Times to repeat"),
        ) |> mcp.Required("name"),
        greetHandler)

    mcp.Serve(server) onerr panic

func greetHandler(args map of string to any) (string, error)
    name := cast.SmartString(args["name"]) onerr return "", err
    count := cast.SmartInt(args["count"]) onerr return "", err

    print("Handling greet for {name}")  # safe: goes to stderr via target directive

    greeting := ""
    i := 0
    for i < count
        greeting = greeting + "Hello, {name}! "
        i = i + 1

    return greeting, empty
```

Build: `kukicha build server.kuki` (target auto-detected from directive)

---

## Implementation Order

```
Phase 1 (cast)  ────┐
                     ├──→ Phase 3 (mcp stdlib)
Phase 2 (--target) ──┘
```

Phases 1 and 2 are independent and can be done in parallel. Phase 3 uses both.

---

## Files Summary

| Action | File | Phase |
|--------|------|-------|
| Create | `stdlib/cast/cast.go` | 1 |
| Create | `stdlib/cast/cast_test.go` | 1 |
| Modify | `internal/codegen/codegen.go` (Generator struct + print codegen + imports) | 2 |
| Modify | `internal/ast/ast.go` (Target field on Program) | 2 |
| Modify | `cmd/kukicha/main.go` (CLI flag + directive detection) | 2 |
| Create | `stdlib/mcp/mcp.kuki` | 3 |
| Create | `stdlib/mcp/mcp.go` (transpiled output) | 3 |
| Modify | `go.mod` (add `modelcontextprotocol/go-sdk` dependency) | 3 |
| Modify | `Makefile` (add cast/mcp to generate target) | 3 |

---

## Verification

1. **cast**: Run `go test ./stdlib/cast/` — test SmartInt/SmartFloat64/SmartBool/SmartString with native types, string representations, json.Number, and error cases
2. **--target mcp**: Build a test `.kuki` file with `print()` both with `--target mcp` flag and `# target: mcp` directive. Verify generated Go uses `fmt.Fprintln(os.Stderr, ...)`. Build without target, verify `fmt.Println(...)`.
3. **mcp stdlib**: Build the example server, test with:
   ```
   echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | ./server
   ```
   Verify JSON-RPC response on stdout, debug prints on stderr.
4. Run `make test` and `make generate` to ensure no regressions
