# Kukicha v0.0.2 Roadmap

## Language Features

### switch/case/default
- `switch`, `case`, `default` are reserved keywords in the lexer but have **no parser, AST, or compiler support**
- The if/else chain in the CLI Explorer tutorial is the only option for command dispatch today
- Implementing switch/case would clean up handler routing in both CLI and web tutorials
- Priority: **High** — one of the most common control flow patterns in Go

### interface
- `interface` keyword exists in the grammar but has zero tutorial coverage
- No working examples anywhere in docs or tutorials
- Priority: **Medium** — important Go concept, needed for idiomatic patterns (io.Reader, error interface, etc.)

### Channels and concurrency primitives
- `channel of Type`, `send`, `receive`, `close` are in the grammar
- `go func()` is shown in the production tutorial (click tracking) but channels are never used
- Priority: **Medium** — core Go differentiator, but the current tutorials focus on beginner-intermediate

---

## Documentation / Tutorial Gaps

### Missing from tutorials (language features)

| Feature | Status | Natural home |
|---------|--------|-------------|
| `switch/case` | Reserved, not implemented | CLI Explorer command dispatch (once implemented) |
| `interface` | In grammar, zero examples | New section in Production tutorial or standalone |
| `channel` / `send` / `receive` | In grammar, never used | New "Concurrency" tutorial or Production addendum |
| `recover` | In grammar, zero examples | Production tutorial error handling section |
| `as` (type assertions) | Barely mentioned | Alongside interface tutorial |
| `many` (variadic params) | Zero examples | Beginner tutorial functions section |
| `++` / `--` | In grammar, zero examples | Beginner tutorial loops section |

### Missing from tutorials (stdlib packages)

| Package | Key functions | Natural home |
|---------|--------------|-------------|
| `iterator` | Filter, Map, FlatMap, Reduce, Find, Any, All | CLI Explorer filtering section — functional alternative to slice.Filter |
| `concurrent` | Parallel, ParallelWithLimit | CLI Explorer — fetch repos for multiple users at once |
| `cli` | Args, Flags, Actions | CLI Explorer — replace manual arg parsing |
| `template` | Render, Execute, Parse | Link Shortener — serve an HTML landing page |
| `result` | Some, None, Ok, Err, Map, UnwrapOr | Production tutorial — Rust-style optionals |
| `parse` | CSV, YAML | Standalone example or LLM tutorial extension |
| `retry` | Retry with backoff | Production tutorial — resilient HTTP calls |
| `files` | Read, Write, Append, List | Covered in LLM tutorial but deserves beginner coverage |
| `datetime` | Format, Parse, Now, AddDays | Covered in LLM tutorial but deserves standalone examples |

### Potential new tutorials

- **Concurrency patterns** — goroutines, channels, fan-out/fan-in, select. Build a concurrent URL health checker
- **Interface patterns** — defining and implementing interfaces, the io.Reader/Writer pattern, error interface
- **File processing** — read CSV, transform with pipes, write JSON. Shows files + parse + iterator in one flow
