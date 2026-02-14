# Kukicha v0.0.2 Roadmap

## Language Features

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
| `interface` | In grammar, zero examples | New section in Production tutorial or standalone |
| `channel` / `send` / `receive` | In grammar, never used | New "Concurrency" tutorial or Production addendum |
| `as` (type assertions) | Barely mentioned | Alongside interface tutorial |

### Missing from tutorials (stdlib packages)

| Package | Key functions | Natural home |
|---------|--------------|-------------|
| `parse` | CSV, YAML | Standalone example or LLM tutorial extension |
| `files` | Read, Write, Append | Covered in LLM tutorial but deserves beginner coverage |
| `datetime` | Format, Parse, Now | Covered in LLM tutorial but deserves standalone examples |

### Potential new tutorials

- **Concurrency patterns** — goroutines, channels, fan-out/fan-in, select. Build a concurrent URL health checker
- **Interface patterns** — defining and implementing interfaces, the io.Reader/Writer pattern, error interface
- **File processing** — read CSV, transform with pipes, write JSON. Shows files + parse + iterator in one flow

