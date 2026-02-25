# Kukicha Language Limitations

Known gaps between Kukicha syntax and Go patterns. These are compiler limitations, not stdlib coverage issues — all stdlib packages are pure Kukicha.

## ~~1. Variadic Slice Spreading~~ (RESOLVED)

**Status:** Fixed. The `many` keyword now supports spreading dynamically-built slices into variadic parameters. The semantic analyzer correctly validates that a `list of T` argument is compatible with a variadic `T` parameter.

```kukicha
# Static spread — always worked
args := list of int{1, 2, 3}
result := Sum(many args)

# Dynamic spread — NOW WORKS
opts := list of client.Opt{client.FromEnv}
if host != empty
    opts = append(opts, client.WithHost(host))
cli := client.NewClientWithOpts(many opts)
```

## ~~2. Assigning Closures to Struct Fields~~ (RESOLVED)

**Status:** Fixed. Function literals can be assigned to struct fields. The semantic analyzer now properly analyzes `FunctionLiteral` expressions, validating parameter types, return types, and the closure body.

```kukicha
type Handler
    name string
    process func(string) string

h := Handler{name: "test"}
h.process = func(s string) string
    return "processed: " + s
result := h.process("hello")
```
