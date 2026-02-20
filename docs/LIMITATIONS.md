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

## 2. Complex Streaming I/O (partially resolved)

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


### `stdlib/string` does not wrap all of Go's `strings` package

A few `strings` functions have no equivalent in `stdlib/string`:

| Go `strings` function | Replacement |
|---|---|
| `strings.CutPrefix(s, prefix)` → `(after, found)` | `string.HasPrefix(s, prefix)` + `string.TrimPrefix(s, prefix)` |
| `strings.NewReader(s)` → `io.Reader` | `bytes.NewBufferString(s)` (import `"bytes"`) |

When migrating `.kuki` files from raw `import "strings"` to `import "stdlib/string"`, check for
these two functions. Both replacements are slightly more verbose but fully equivalent.

---

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
| Variadic interface spreading | 4 functions (Docker client init) | Low-medium — extend `many` |
| Anonymous struct literals | 4 functions (streaming JSON decode + auth in container) | Medium — parser + codegen |
| Closure-as-struct-field | 1 function (netguard `DialContext`) | Low — codegen for struct field func literals |
