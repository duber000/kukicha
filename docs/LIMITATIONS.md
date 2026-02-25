# Kukicha Language Limitations

This document tracks language features that were previously missing from Kukicha,
and notes remaining gaps where Kukicha syntax differs from Go.

## Resolved Limitations (all Go helpers eliminated)

All stdlib packages that previously required hand-written Go helper files
(`*_helper.go`) have been fully migrated to Kukicha. The following limitations
were resolved using existing compiler capabilities:

### 1. Functional Options / Variadic `...T` of Interface Type ✅ RESOLVED

Go libraries often accept a variadic slice of option functions:

```go
opts := []client.Opt{
    client.FromEnv,
    client.WithAPIVersionNegotiation(),
    client.WithHost(host),
}
cli, err := client.NewClientWithOpts(opts...)
```

**Resolution:** Kukicha's `list of` syntax and `many` spread already support this pattern:

```kukicha
opts := list of client.Opt{client.FromEnv, client.WithAPIVersionNegotiation()}
if host != ""
    opts = append(opts, client.WithHost(host))
cli := client.NewClientWithOpts(many opts) onerr return empty, error "{error}"
```

**Migrated:** `container.newClient`, `container.Connect`, `container.ConnectRemote`, `container.Open`.

---

### 2. Complex Streaming I/O ✅ RESOLVED

Anonymous struct declarations are not supported inline in function bodies.
The **workaround** is to define named types at package level:

```kukicha
# Named types replace anonymous structs
type buildStreamMsg
    Stream string json:"stream"
    Aux buildStreamAux json:"aux"
    Error string json:"error"
```

**Migrated:** `Pull`, `PullAuth`, `loadDockerAuth`, `LoginFromConfig`, `buildImage`, `Build`.

---

### 3. Closure-as-Struct-Field ✅ RESOLVED

Assigning closures to struct fields works in Kukicha:

```kukicha
dialer := net.Dialer{Timeout: datetime.Seconds(30)}
dialer.Control = func(network string, address string, c syscall.RawConn) error
    # closure body
    return empty
```

**Migrated:** `netguard.DialContext`, `netguard.HTTPTransport`.

---

### 4. filepath.WalkDir Closures ✅ RESOLVED

Passing closures to `filepath.WalkDir` (and `filepath.Walk`) works:

```kukicha
walkErr := filepath.WalkDir(path, func(walkPath string, d os.DirEntry, err error) error
    if err != empty
        return err
    if d.IsDir() and d.Name() == ".git"
        return filepath.SkipDir
    # ... process file
    return empty
)
```

**Migrated:** `container.buildImage`, `container.createTarFromPath`, `container.CopyFrom`, `container.CopyTo`.

---

### 5. http.HandlerFunc Middleware ✅ RESOLVED

Wrapping closures in `http.HandlerFunc` for middleware works:

```kukicha
func SecureHeaders(handler any) any
    h := handler.(http.Handler)
    return http.HandlerFunc(func(w http.ResponseWriter, r reference http.Request)
        w.Header().Set("X-Content-Type-Options", "nosniff")
        h.ServeHTTP(w, r)
    )
```

**Migrated:** `http.SecureHeaders`.

---

## Remaining Syntax Gaps (no Go helpers needed)

These are Kukicha language limitations that do **not** require Go helper files
but are worth documenting:

### Anonymous struct literals

Kukicha does not support anonymous struct declarations inline in function bodies.
Use named types at package level instead. This is a style difference, not a blocker.

### `stdlib/string` does not wrap all of Go's `strings` package

A few `strings` functions have no equivalent in `stdlib/string`:

| Go `strings` function | Replacement |
|---|---|
| `strings.CutPrefix(s, prefix)` → `(after, found)` | `string.HasPrefix(s, prefix)` + `string.TrimPrefix(s, prefix)` |
| `strings.NewReader(s)` → `io.Reader` | `bytes.NewBufferString(s)` (import `"bytes"`) |

When migrating `.kuki` files from raw `import "strings"` to `import "stdlib/string"`, check for
these two functions. Both replacements are slightly more verbose but fully equivalent.

## Current Packages with Go Helpers

| Package | File | Lines | Reason |
|---------|------|-------|--------|
| `stdlib/mcp` | `mcp_tool.go` | ~100 | Callback bridge for MCP tool/resource/prompt handlers |

The `mcp` package's `_tool.go` uses a Go callback bridge pattern that requires
runtime reflection and interface{} dispatch not yet expressible in Kukicha.
