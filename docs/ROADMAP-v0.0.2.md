# Kukicha v0.0.2 Roadmap

## Language Features

### switch/when/otherwise
- `switch` is implemented with beginner-friendly branch keywords: `when` and `otherwise`
- Transpilation is Go-compatible: `when` maps to Go `case`, `otherwise` maps to Go `default`
- This now replaces long if/else command-dispatch chains in tutorial code
- Priority: **Completed in v0.0.2**

**Implemented beginner syntax (English-sounding, Go-compatible):**
- Keep `switch` (already Go-native and widely recognized)
- Use `when` instead of `case`
- Use `otherwise` instead of `default`
- Keep Go behavior: first matching branch runs, no implicit fallthrough
- `default` is accepted as an alias for easier transition from Go

```kukicha
switch command
    when "fetch"
        fetchRepos()
    when "help", "h"
        showHelp()
    otherwise
        print "Unknown command: {command}"
```

**Condition switch (Go's `switch true` style):**
```kukicha
switch
    when stars >= 1000
        print "Popular"
    when stars >= 100
        print "Growing"
    otherwise
        print "New"
```

**Go mapping:**
- `switch expr` -> `switch expr`
- `when a, b` -> `case a, b`
- `otherwise` -> `default`
- Bare `switch` + `when condition` -> `switch { case condition: ... }`

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
| `switch/when` | **Done** — CLI Explorer, Link Shortener, Production tutorials updated | CLI Explorer command dispatch, HTTP method dispatch |
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

---

## Stdlib Gaps (discovered writing tutorials)

These are pain points hit while developing real tutorial code. Go stdlib fallbacks exist but hurt discoverability and add boilerplate that works against the "Go for Scripts" pitch.

### High priority

**Console input helper** (`stdlib/input`)
- Every interactive tutorial needs this 3-line boilerplate:
  ```kukicha
  reader := bufio.NewReader(os.Stdin)
  input := reader.ReadString('\n') onerr ""
  input = input |> string.TrimSpace()
  ```
- Proposal: `input.ReadLine("> ")` → `(string, error)` or `input.Prompt("Enter command: ")` → `string`
- Affected tutorials: CLI Explorer, any future interactive tool
- This is the first thing a beginner hits when making anything interactive

**Shell output shorthand** (`stdlib/shell`)
- Getting command output currently takes two lines + a type conversion:
  ```kukicha
  result := shell.New("git", "diff", "--staged") |> shell.Execute()
  diff := string(shell.GetOutput(result))
  ```
- Proposal: `shell.Output("git", "diff", "--staged")` → `(string, error)`
- Affected tutorials: LLM Scripting (every example would get cleaner)

### Medium priority

**Random string generation** (`stdlib/string` or `stdlib/random`)
- The production tutorial builds a random code generator from scratch (7 lines)
- Proposal: `string.Random(6)` or `string.RandomAlphanumeric(n)` → `string`
- Common in web services: tokens, session IDs, short codes, API keys

**Map utilities** (`stdlib/maps` or built into language)
- Link shortener uses `map of string to Link` but no helpers exist
- Go's `, ok` pattern (`link, exists := store.links[code]`) works but isn't beginner-discoverable
- Proposal: `maps.Keys()`, `maps.Values()`, `maps.Contains()` — matching what `slice` already provides

**String padding** (`stdlib/string`)
- CLI Explorer's `Summary` method tries to align columns but can't pad strings
- Proposal: `string.PadRight(s, width)`, `string.PadLeft(s, width)` for formatted terminal output

### Low priority

**`delete()` documentation**
- `delete(store.links, code)` is a Go builtin that works in Kukicha but isn't in the quick reference, stdlib docs, or any tutorial explanations. It's invisible to someone reading only Kukicha docs.
- Fix: Add to quick reference under "Map Operations" alongside map literal syntax

---

## Syntax Friction (discovered writing tutorials)

### Verbose inline functions
- `slice.Filter` requires a full `function(r Repo) bool` with a body block. Every filter/map call is 3+ lines:
  ```kukicha
  repos |> slice.Filter(function(r Repo) bool
      return r.Stars > 100
  )
  ```
- A short lambda form (e.g., `|r| r.Stars > 100` or `r => r.Stars > 100`) would make the pipe + functional style dramatically cleaner
- Priority: **High** — this is the main friction point when composing pipes with functional helpers

### `go func()...()` IIFE pattern
- The production tutorial's click tracking goroutine uses a Go IIFE:
  ```kukicha
  go func()
      s.mu.Lock()
      s.db.IncrementClicks(code)
      s.mu.Unlock()
  ()
  ```
- The trailing `()` to invoke the anonymous function is non-obvious for beginners
- Proposal: Allow `go` to accept a block directly, or document the pattern clearly
- Priority: **Low** — rare in beginner code, but confusing when encountered
