# Kukicha Language Limitations

Known gaps between Kukicha syntax and Go patterns. These are compiler limitations, not stdlib coverage issues — all stdlib packages are pure Kukicha.

## 1. Variadic Slice Spreading

**Go pattern:** Build a slice of options dynamically, then spread it into a variadic call with `opts...`.

```go
opts := []client.Opt{client.FromEnv}
if host != "" {
    opts = append(opts, client.WithHost(host))
}
cli, err := client.NewClientWithOpts(opts...)
```

**Kukicha today:** The `many` keyword handles static variadic calls (`Sum(many args)`) but cannot spread a dynamically-built slice into a variadic parameter.

```kukicha
# This works — static spread of a known slice
args := list of int{1, 2, 3}
result := Sum(many args)

# This does NOT work — conditional option building + spread
opts := list of client.Opt{client.FromEnv}
if host != empty
    opts = append(opts, client.WithHost(host))
cli := client.NewClientWithOpts(many opts)  # compiler rejects this
```

**Impact:** High. The functional options pattern (`WithTimeout()`, `WithHost()`, etc.) is ubiquitous in Go libraries. Any code that conditionally builds an options slice hits this wall. The only workaround is dropping to a `.go` helper file — which defeats the purpose for beginners who don't know Go.

**Compiler components involved:** Parser (extend `many` in call expressions to accept arbitrary expressions), AST (minor), Semantic (validate spread type matches variadic param), Codegen (emit `...` on the spread expression).

## 2. Assigning Closures to Struct Fields

**Go pattern:** Create a struct, then assign a closure to one of its function-typed fields.

```go
dialer := net.Dialer{Timeout: 30 * time.Second}
dialer.Control = func(network, address string, c syscall.RawConn) error {
    // closure body — captures variables from enclosing scope
    return nil
}
transport := &http.Transport{DialContext: dialer.DialContext}
```

**Kukicha today:** Function literals and arrow lambdas work as values (variable assignment, pipe arguments, callback parameters), but assigning a closure to a struct field is not supported.

```kukicha
# This works — closure as callback argument
items |> slice.Filter(func(n int) bool
    return n > threshold
)

# This does NOT work — closure assigned to struct field
dialer := net.Dialer{Timeout: datetime.Seconds(30)}
dialer.Control = func(network string, address string, c syscall.RawConn) error
    return empty
```

**Impact:** Medium. Hits users building custom HTTP transports, network dialers, or any struct with callback fields. Less common than variadic spreading but blocks real use cases (SSRF guards, connection interceptors, middleware wrappers).

**Compiler components involved:** Semantic (allow `FunctionType` in struct field type-checking and assignment validation), Parser (minor — verify function type annotation parses in field context). Most infrastructure already exists — `FunctionType` is defined in the AST and closures generate correctly.
